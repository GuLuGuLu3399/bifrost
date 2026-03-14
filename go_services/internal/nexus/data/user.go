package data

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"github.com/gulugulu3399/bifrost/internal/nexus/biz"
	"github.com/gulugulu3399/bifrost/internal/pkg/id"
	"github.com/gulugulu3399/bifrost/internal/pkg/observability/logger"
	"github.com/gulugulu3399/bifrost/internal/pkg/xerr"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq" // 用于捕获 Postgres 错误代码
)

type userRepo struct {
	data      *Data
	snowflake *id.SnowflakeGenerator
}

// NewUserRepo 创建用户仓储
func NewUserRepo(data *Data, snowflake *id.SnowflakeGenerator) biz.UserRepo {
	return &userRepo{
		data:      data,
		snowflake: snowflake,
	}
}

// userPO 对应数据库 users 表
type userPO struct {
	ID           int64           `db:"id"`
	Username     string          `db:"username"`
	Email        string          `db:"email"`
	PasswordHash sql.NullString  `db:"password_hash"`
	Nickname     sql.NullString  `db:"nickname"`
	Bio          sql.NullString  `db:"bio"`
	AvatarKey    sql.NullString  `db:"avatar_key"`
	IsAdmin      bool            `db:"is_admin"`
	IsActive     bool            `db:"is_active"`
	Provider     string          `db:"provider"`
	ProviderID   sql.NullString  `db:"provider_id"`
	Version      int64           `db:"version"`
	LastTraceID  sql.NullString  `db:"last_trace_id"`
	Meta         json.RawMessage `db:"meta"` // 使用 RawMessage 延迟解析
	LastLoginAt  sql.NullTime    `db:"last_login_at"`
	CreatedAt    time.Time       `db:"created_at"`
	UpdatedAt    time.Time       `db:"updated_at"`
	DeletedAt    sql.NullTime    `db:"deleted_at"`
}

// toEntity 转换为领域实体 (包含容错逻辑)
func (po *userPO) toEntity() *biz.User {
	user := &biz.User{
		ID:        po.ID,
		Username:  po.Username,
		Email:     po.Email,
		IsAdmin:   po.IsAdmin,
		IsActive:  po.IsActive,
		Provider:  biz.AuthProvider(po.Provider),
		Version:   po.Version,
		CreatedAt: po.CreatedAt,
		UpdatedAt: po.UpdatedAt,
		// 初始化 Meta 为默认值，防止后面解析失败导致 nil
		Meta: biz.UserMeta{},
	}

	if po.PasswordHash.Valid {
		user.PasswordHash = po.PasswordHash.String
	}
	if po.Nickname.Valid {
		user.Nickname = po.Nickname.String
	}
	if po.Bio.Valid {
		user.Bio = po.Bio.String
	}
	if po.AvatarKey.Valid {
		user.AvatarKey = po.AvatarKey.String
	}
	if po.ProviderID.Valid {
		user.ProviderID = po.ProviderID.String
	}
	if po.LastTraceID.Valid {
		user.LastTraceID = po.LastTraceID.String
	}
	if po.LastLoginAt.Valid {
		t := po.LastLoginAt.Time
		user.LastLoginAt = &t
	}
	if po.DeletedAt.Valid {
		t := po.DeletedAt.Time
		user.DeletedAt = &t
	}

	// [改进] Meta 解析容错处理
	// 如果数据库里的 JSON 脏了，记录错误日志（实际项目中应接入 Logger），
	// 但不阻断流程，返回 Meta 为默认值的 User 对象，保证系统可用性。
	if len(po.Meta) > 0 {
		if err := json.Unmarshal(po.Meta, &user.Meta); err != nil {
			// 使用结构化日志记录而不是 printf 风格
			logger.Error("failed to unmarshal user meta", logger.Int64("user_id", po.ID), logger.Err(err))
			// 保持 user.Meta 为默认空值，继续返回
		}
	}

	return user
}

// Create 创建用户
func (r *userRepo) Create(ctx context.Context, user *biz.User) (int64, error) {
	user.ID = r.snowflake.GenerateInt64()

	query := `
		INSERT INTO users (
			id, username, email, password_hash, nickname, bio, avatar_key,
			is_admin, is_active, provider, provider_id,
			version, last_trace_id, meta, created_at, updated_at
		) VALUES (
			:id, :username, :email, :password_hash, :nickname, :bio, :avatar_key,
			:is_admin, :is_active, :provider, :provider_id,
			:version, :last_trace_id, :meta, :created_at, :updated_at
		)
	`

	// 序列化 Meta
	metaBytes, err := json.Marshal(user.Meta)
	if err != nil {
		return 0, xerr.Wrap(err, xerr.CodeInternal, "序列化用户 Meta 失败")
	}

	po := &userPO{
		ID:           user.ID,
		Username:     user.Username,
		Email:        user.Email,
		IsAdmin:      user.IsAdmin,
		IsActive:     user.IsActive,
		Provider:     string(user.Provider),
		Version:      user.Version,
		LastTraceID:  stringToNullString(user.LastTraceID),
		Meta:         metaBytes,
		CreatedAt:    user.CreatedAt,
		UpdatedAt:    user.UpdatedAt,
		PasswordHash: stringToNullString(user.PasswordHash),
		Nickname:     stringToNullString(user.Nickname),
		Bio:          stringToNullString(user.Bio),
		AvatarKey:    stringToNullString(user.AvatarKey),
		ProviderID:   stringToNullString(user.ProviderID),
	}

	// 把 user 的可选字段写入 PO（原代码错误地把 PO 的字段写回 user）
	db := r.data.DB(ctx)
	_, err = sqlx.NamedExecContext(ctx, db, query, po)
	if err != nil {
		// [改进] 捕获并映射常见的数据库错误到更明确的业务错误
		var pqErr *pq.Error
		if errors.As(err, &pqErr) {
			switch string(pqErr.Code) {
			case "23505":
				// 唯一索引冲突
				return 0, xerr.New(xerr.CodeConflict, "用户名或邮箱已存在")
			case "23514":
				// CHECK 约束失败，根据约束名返回更清晰的提示
				switch pqErr.Constraint {
				case "check_email_fmt":
					return 0, xerr.New(xerr.CodeValidation, "邮箱格式不合法")
				case "check_username_len":
					return 0, xerr.New(xerr.CodeValidation, "用户名长度至少 3")
				case "check_auth_logic":
					return 0, xerr.New(xerr.CodeValidation, "本地用户必须设置密码")
				default:
					return 0, xerr.New(xerr.CodeValidation, "数据校验失败")
				}
			case "23502":
				// NOT NULL 违例
				return 0, xerr.New(xerr.CodeValidation, "缺少必填字段")
			}
		}
		// 其他未识别错误保留原始信息但不暴露细节
		return 0, xerr.Wrap(err, xerr.CodeInternal, "创建用户失败")
	}

	return user.ID, nil
}

// updateUserPO 专用于 Update 操作的 PO
type updateUserPO struct {
	ID          int64           `db:"id"`
	Nickname    sql.NullString  `db:"nickname"`
	Bio         sql.NullString  `db:"bio"`
	AvatarKey   sql.NullString  `db:"avatar_key"`
	Meta        json.RawMessage `db:"meta"`
	LastTraceID sql.NullString  `db:"last_trace_id"`
	OldVersion  int64           `db:"old_version"` // 旧版本 (用于 WHERE)
	UpdatedAt   time.Time       `db:"updated_at"`
}

// Update 更新用户基础信息 (带乐观锁)
// 约定：user.Version 传入的是 oldVersion（数据库当前版本），本方法负责把 version + 1。
func (r *userRepo) Update(ctx context.Context, user *biz.User) error {
	query := `
		UPDATE users SET
			nickname = :nickname,
			bio = :bio,
			avatar_key = :avatar_key,
			meta = :meta,
			last_trace_id = :last_trace_id,
			version = :old_version + 1,
			updated_at = :updated_at
		WHERE id = :id AND version = :old_version AND deleted_at IS NULL
	`

	metaBytes, err := json.Marshal(user.Meta)
	if err != nil {
		return xerr.Wrap(err, xerr.CodeInternal, "序列化用户 Meta 失败")
	}

	po := updateUserPO{
		ID:          user.ID,
		OldVersion:  user.Version,
		UpdatedAt:   user.UpdatedAt,
		Meta:        metaBytes,
		LastTraceID: stringToNullString(user.LastTraceID),
		Nickname:    stringToNullString(user.Nickname),
		Bio:         stringToNullString(user.Bio),
		AvatarKey:   stringToNullString(user.AvatarKey),
	}

	db := r.data.DB(ctx)
	result, err := sqlx.NamedExecContext(ctx, db, query, po)
	if err != nil {
		return xerr.Wrap(err, xerr.CodeInternal, "更新用户失败")
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return xerr.Wrap(err, xerr.CodeInternal, "获取影响行数失败")
	}

	if affected == 0 {
		return biz.ErrVersionConflict // 乐观锁冲突
	}

	// 同步内存对象的版本号（可选，但能避免上层再次手动 +1）
	user.Version = user.Version + 1
	return nil
}

// UpdatePassword 修改密码 (独立乐观锁逻辑)
func (r *userRepo) UpdatePassword(ctx context.Context, id int64, passwordHash string, oldVersion int64) error {
	// [改进] SQL 逻辑更直观：SET version = oldVersion + 1 WHERE version = oldVersion
	query := `
		UPDATE users SET
			password_hash = $1,
			version = version + 1,
			updated_at = $2
		WHERE id = $3 AND version = $4 AND deleted_at IS NULL
	`

	now := time.Now()
	db := r.data.DB(ctx)
	// 注意参数顺序：hash, time, id, oldVersion
	result, err := db.ExecContext(ctx, query, passwordHash, now, id, oldVersion)
	if err != nil {
		return xerr.Wrap(err, xerr.CodeInternal, "修改密码失败")
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return xerr.Wrap(err, xerr.CodeInternal, "获取影响行数失败")
	}

	if affected == 0 {
		return biz.ErrVersionConflict
	}

	return nil
}

// UpdateLastLogin 更新最后登录时间
func (r *userRepo) UpdateLastLogin(ctx context.Context, id int64, loginAt time.Time) error {
	query := `UPDATE users SET last_login_at = $1 WHERE id = $2`

	db := r.data.DB(ctx)
	_, err := db.ExecContext(ctx, query, loginAt, id)
	if err != nil {
		// [改进] 统一错误包装
		return xerr.Wrap(err, xerr.CodeInternal, "更新最后登录时间失败")
	}

	return nil
}

// GetByID 根据 ID 获取用户
func (r *userRepo) GetByID(ctx context.Context, id int64) (*biz.User, error) {
	query := `SELECT * FROM users WHERE id = $1 AND deleted_at IS NULL`

	var po userPO
	db := r.data.DB(ctx)
	err := sqlx.GetContext(ctx, db, &po, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, xerr.Wrap(err, xerr.CodeInternal, "查询用户失败")
	}

	return po.toEntity(), nil
}

// GetByEmail 根据邮箱获取用户
func (r *userRepo) GetByEmail(ctx context.Context, email string) (*biz.User, error) {
	query := `SELECT * FROM users WHERE email = $1 AND deleted_at IS NULL`

	var po userPO
	db := r.data.DB(ctx)
	err := sqlx.GetContext(ctx, db, &po, query, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, xerr.Wrap(err, xerr.CodeInternal, "查询用户失败")
	}

	return po.toEntity(), nil
}

// GetByUsername 根据用户名获取用户
func (r *userRepo) GetByUsername(ctx context.Context, username string) (*biz.User, error) {
	query := `SELECT * FROM users WHERE username = $1 AND deleted_at IS NULL`

	var po userPO
	db := r.data.DB(ctx)
	err := sqlx.GetContext(ctx, db, &po, query, username)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, xerr.Wrap(err, xerr.CodeInternal, "查询用户失败")
	}

	return po.toEntity(), nil
}

// Exists 检查用户名或邮箱是否存在
func (r *userRepo) Exists(ctx context.Context, username, email string) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM users 
			WHERE (username = $1 OR email = $2) 
			AND deleted_at IS NULL
		)
	`
	var exists bool
	db := r.data.DB(ctx)
	if err := sqlx.GetContext(ctx, db, &exists, query, username, email); err != nil {
		return false, xerr.Wrap(err, xerr.CodeInternal, "检查用户存在性失败")
	}

	return exists, nil
}

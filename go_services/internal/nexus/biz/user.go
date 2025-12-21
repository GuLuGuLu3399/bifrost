package biz

import (
	"context"
	"time"

	"github.com/gulugulu3399/bifrost/internal/pkg/security"
)

// ===========================================
// 枚举与常量 (Value Objects & Enums)
// ===========================================

// AuthProvider 认证提供商
type AuthProvider string

const (
	AuthProviderLocal AuthProvider = "local"
	// 未来可扩展 OAuth 提供商，如 Google, GitHub 等
)

// UserMeta 用户元数据结构 (对应 DB 的 JSONB)
// 强类型定义，避免 map[string]interface{} 的混乱
type UserMeta struct {
	Theme             string `json:"theme,omitempty"`              // light, dark, system
	NotificationLevel string `json:"notification_level,omitempty"` // all, important, none
	Language          string `json:"language,omitempty"`
}

// ===========================================
// 领域实体 (Domain Entity)
// ===========================================

// User 用户实体
// 纯净的 Go 结构体，不含 sql/db 标签
type User struct {
	ID           int64
	Username     string
	Email        string
	PasswordHash string // 敏感字段：JSON 序列化时需注意处理
	Nickname     string
	Bio          string
	AvatarKey    string // MinIO Key, e.g., "avatars/u123/head.jpg"

	IsAdmin  bool
	IsActive bool

	Provider   AuthProvider
	ProviderID string // OAuth ID

	// 运维防御 (Ops Shield)
	Version     int64  // 乐观锁
	LastTraceID string // 链路追踪 ID

	// 扩展数据
	Meta UserMeta // 对应数据库的 meta JSONB

	// 时间轴
	LastLoginAt *time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   *time.Time // 软删除
}

// ===========================================
// 领域行为 (Domain Behaviors)
// ===========================================

// SetPassword 设置密码
func (u *User) SetPassword(plainPassword string) error {
	hash, err := security.HashPassword(plainPassword)
	if err != nil {
		return err
	}
	u.PasswordHash = hash
	return nil
}

// CheckPassword 校验密码
func (u *User) CheckPassword(plainPassword string) bool {
	return security.CheckPassword(plainPassword, u.PasswordHash)
}

// UpdateMeta 更新元数据 (保持不变)
func (u *User) UpdateMeta(theme, lang string) {
	u.Meta.Theme = theme
	u.Meta.Language = lang
	u.UpdatedAt = time.Now()
}

// ===========================================
// 仓储接口 (Repository Interface)
// ===========================================

// UserRepo 用户仓储接口
// 定义在 biz 层，由 data 层实现
type UserRepo interface {
	// Create 创建用户
	// 返回生成的 Snowflake ID
	Create(ctx context.Context, user *User) (int64, error)

	// Update 更新用户基础信息 (Nickname, Bio, Avatar, Meta)
	// 必须实现乐观锁 (WHERE version = old_version)
	Update(ctx context.Context, user *User) error

	// UpdatePassword 修改密码
	// 独立方法，必须校验 version，修改后 version++
	UpdatePassword(ctx context.Context, id int64, passwordHash string, version int64) error

	// UpdateLastLogin 更新最后登录时间
	// 轻量级更新，通常不需要乐观锁，或者容忍覆盖
	UpdateLastLogin(ctx context.Context, id int64, loginAt time.Time) error

	// GetByID 根据 ID 获取用户
	GetByID(ctx context.Context, id int64) (*User, error)

	// GetByEmail 根据邮箱获取用户 (用于登录/找回密码)
	GetByEmail(ctx context.Context, email string) (*User, error)

	// GetByUsername 根据用户名获取用户 (用于登录)
	GetByUsername(ctx context.Context, username string) (*User, error)

	// Exists 检查用户名或邮箱是否已被占用 (用于注册前检查)
	// 返回 true 表示已存在
	Exists(ctx context.Context, username, email string) (bool, error)
}

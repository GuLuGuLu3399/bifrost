package biz

import (
	"context"
	"strings"
	"time"

	"github.com/gulugulu3399/bifrost/internal/pkg/security"
	"github.com/gulugulu3399/bifrost/internal/pkg/xerr"
)

// UserUseCase 用户业务逻辑
type UserUseCase struct {
	repo UserRepo
	tx   Transaction // 事务管理器
}

// NewUserUseCase 构造函数
func NewUserUseCase(repo UserRepo, tx Transaction) *UserUseCase {
	return &UserUseCase{
		repo: repo,
		tx:   tx,
	}
}

// ===========================================
// 注册 (Register)
// ===========================================

type RegisterInput struct {
	Username string
	Email    string
	Password string
	Nickname string
}

func (uc *UserUseCase) Register(ctx context.Context, input *RegisterInput) (int64, error) {
	var userID int64

	// 1. [Security] 校验密码强度 (Fail Fast)
	// 使用我们在 pkg/security 中定义的规则
	if err := security.ValidatePassword(input.Password, nil); err != nil {
		return 0, xerr.New(xerr.CodeValidation, err.Error())
	}

	// 2. [Transaction] 开启事务：创建用户
	err := uc.tx.ExecTx(ctx, func(txCtx context.Context) error {
		// A. 唯一性检查
		exists, err := uc.repo.Exists(txCtx, input.Username, input.Email)
		if err != nil {
			return err
		}
		if exists {
			return ErrUserAlreadyExists
		}

		// B. 构建领域实体
		user := &User{
			Username:  input.Username,
			Email:     input.Email,
			Nickname:  input.Nickname,
			IsActive:  true,
			IsAdmin:   false,
			Provider:  AuthProviderLocal,
			Version:   1,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			// Meta 初始化为空结构体，避免 nil pointer
			Meta: UserMeta{
				Theme:             "system",
				NotificationLevel: "all",
			},
		}

		// C. [Security] 设置密码 (领域行为)
		// 内部会自动调用 security.HashPassword
		if err := user.SetPassword(input.Password); err != nil {
			return err
		}

		// D. 持久化用户
		id, err := uc.repo.Create(txCtx, user)
		if err != nil {
			return err
		}
		userID = id
		user.ID = id

		return nil
	})

	if err != nil {
		return 0, err
	}
	return userID, nil
}

// ===========================================
// 2. 登录 (Login)
// ===========================================

type LoginInput struct {
	Identifier string // 可以是 Username 或 Email
	Password   string
}

func (uc *UserUseCase) Login(ctx context.Context, input *LoginInput) (*User, error) {
	// 1. 查找用户 (支持邮箱或用户名混用)
	var user *User
	var err error

	if strings.Contains(input.Identifier, "@") {
		user, err = uc.repo.GetByEmail(ctx, input.Identifier)
	} else {
		user, err = uc.repo.GetByUsername(ctx, input.Identifier)
	}

	if err != nil {
		return nil, err
	}
	// 防枚举攻击：即使用户不存在，也只返回“账号或密码错误”
	if user == nil {
		return nil, ErrInvalidCredentials
	}

	// 2. [Security] 校验密码
	if !user.CheckPassword(input.Password) {
		return nil, ErrInvalidCredentials
	}

	// 3. 检查状态
	if !user.IsActive {
		return nil, ErrUserInactive
	}

	// 4. [副作用] 更新最后登录时间
	// 这是一个轻量级写入，通常可以异步做，但为了数据准确性，我们在单独的小事务中同步做
	_ = uc.tx.ExecTx(ctx, func(txCtx context.Context) error {
		return uc.repo.UpdateLastLogin(txCtx, user.ID, time.Now())
	})

	return user, nil
}

// ===========================================
// 3. 修改密码 (ChangePassword)
// ===========================================

type ChangePasswordInput struct {
	UserID      int64
	OldPassword string
	NewPassword string
}

func (uc *UserUseCase) ChangePassword(ctx context.Context, input *ChangePasswordInput) error {
	// [Security] 校验新密码强度
	if err := security.ValidatePassword(input.NewPassword, nil); err != nil {
		return xerr.New(xerr.CodeValidation, err.Error())
	}

	return uc.tx.ExecTx(ctx, func(txCtx context.Context) error {
		// A. 获取用户 (需要带 Version 以便后续乐观锁更新)
		user, err := uc.repo.GetByID(txCtx, input.UserID)
		if err != nil {
			return err
		}
		if user == nil {
			return ErrUserNotFound
		}

		// B. [Security] 验证旧密码
		if !user.CheckPassword(input.OldPassword) {
			return ErrPasswordMismatch
		}

		// C. [Security] 设置新密码
		if err := user.SetPassword(input.NewPassword); err != nil {
			return err
		}

		// D. 持久化 (使用乐观锁)
		// UpdatePassword 接口应该接收 version 参数
		if err := uc.repo.UpdatePassword(txCtx, user.ID, user.PasswordHash, user.Version); err != nil {
			return err
		}

		return nil
	})
}

// ===========================================
// 4. 获取资料 (GetProfile)
// ===========================================

func (uc *UserUseCase) GetProfile(ctx context.Context, id int64) (*User, error) {
	// 纯读操作，不需要事务
	user, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrUserNotFound
	}
	return user, nil
}

// ===========================================
// 5. 更新资料 (UpdateProfile)
// ===========================================

type UpdateProfileInput struct {
	UserID    int64
	Nickname  string
	Bio       string
	AvatarKey string
	Theme     string
	Language  string
}

func (uc *UserUseCase) UpdateProfile(ctx context.Context, input *UpdateProfileInput) error {
	return uc.tx.ExecTx(ctx, func(txCtx context.Context) error {
		// A. 获取用户
		user, err := uc.repo.GetByID(txCtx, input.UserID)
		if err != nil {
			return err
		}
		if user == nil {
			return ErrUserNotFound
		}

		// B. 更新字段 (领域行为)
		user.Nickname = input.Nickname
		user.Bio = input.Bio
		user.AvatarKey = input.AvatarKey
		user.UpdateMeta(input.Theme, input.Language) // 更新 JSONB 字段

		// C. 持久化
		return uc.repo.Update(txCtx, user)
	})
}

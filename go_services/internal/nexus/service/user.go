package service

import (
	"context"
	"time"

	nexusv1 "github.com/gulugulu3399/bifrost/api/content/v1/nexus"
	"github.com/gulugulu3399/bifrost/internal/nexus/biz"
	"github.com/gulugulu3399/bifrost/internal/pkg/contextx"
	"github.com/gulugulu3399/bifrost/internal/pkg/security"
	"github.com/gulugulu3399/bifrost/internal/pkg/xerr"
)

const (
	userWriteTimeout = 15 * time.Second
	userReadTimeout  = 5 * time.Second
)

// UserService 用户 gRPC 服务
type UserService struct {
	nexusv1.UnimplementedUserServiceServer
	userUc     *biz.UserUseCase
	jwtManager *security.JWTManager
}

func NewUserService(userUc *biz.UserUseCase, jwtManager *security.JWTManager) *UserService {
	return &UserService{
		userUc:     userUc,
		jwtManager: jwtManager,
	}
}

// ==========================================
// 注册 (Register)
// ==========================================

func (s *UserService) Register(ctx context.Context, req *nexusv1.RegisterRequest) (*nexusv1.RegisterResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, userWriteTimeout)
	defer cancel()

	if req.GetUsername() == "" || req.GetEmail() == "" || req.GetPassword() == "" {
		return nil, xerr.New(xerr.CodeBadRequest, "username, email and password are required")
	}

	// 1. DTO 转换
	input := &biz.RegisterInput{
		Username: req.GetUsername(),
		Email:    req.GetEmail(),
		Password: req.GetPassword(),
		Nickname: req.GetNickname(),
	}

	// 2. 调用业务逻辑
	id, err := s.userUc.Register(ctx, input)
	if err != nil {
		return nil, err
	}

	// 3. 返回 ID
	return &nexusv1.RegisterResponse{UserId: id}, nil
}

// ==========================================
// 登录 (Login)
// ==========================================

func (s *UserService) Login(ctx context.Context, req *nexusv1.LoginRequest) (*nexusv1.LoginResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, userWriteTimeout)
	defer cancel()

	if req.GetIdentifier() == "" || req.GetPassword() == "" {
		return nil, xerr.New(xerr.CodeBadRequest, "identifier and password are required")
	}

	// 1. DTO 转换
	input := &biz.LoginInput{
		Identifier: req.GetIdentifier(),
		Password:   req.GetPassword(),
	}

	// 2. 业务验证
	user, err := s.userUc.Login(ctx, input)
	if err != nil {
		return nil, err
	}

	// 3. 生成 TokenPair
	accessToken, refreshToken, err := s.jwtManager.GenerateTokenPair(
		user.ID,
		user.Username,
		user.IsAdmin,
	)
	if err != nil {
		return nil, err
	}

	// 4. 返回
	return &nexusv1.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    86400,
	}, nil
}

// ==========================================
// 修改密码 (ChangePassword)
// ==========================================

func (s *UserService) ChangePassword(ctx context.Context, req *nexusv1.ChangePasswordRequest) (*nexusv1.ChangePasswordResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, userWriteTimeout)
	defer cancel()

	// 1. 获取当前登录用户 ID
	userID := contextx.UserIDFromContext(ctx)
	if userID == 0 {
		return nil, xerr.New(xerr.CodeUnauthorized, "user not authenticated")
	}

	if req.GetOldPassword() == "" || req.GetNewPassword() == "" {
		return nil, xerr.New(xerr.CodeBadRequest, "old_password and new_password are required")
	}

	// 2. DTO 转换
	input := &biz.ChangePasswordInput{
		UserID:      userID,
		OldPassword: req.GetOldPassword(),
		NewPassword: req.GetNewPassword(),
	}

	// 3. 调用业务逻辑
	if err := s.userUc.ChangePassword(ctx, input); err != nil {
		return nil, err
	}

	return &nexusv1.ChangePasswordResponse{Success: true}, nil
}

// ==========================================
// 获取资料 (GetProfile)
// ==========================================

func (s *UserService) GetProfile(ctx context.Context, req *nexusv1.GetProfileRequest) (*nexusv1.GetProfileResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, userReadTimeout)
	defer cancel()

	// 逻辑：如果请求传了 UserID 就查指定的，没传(0)就查自己的
	targetUserID := req.GetUserId()
	if targetUserID == 0 {
		targetUserID = contextx.UserIDFromContext(ctx)
		if targetUserID == 0 {
			return nil, xerr.New(xerr.CodeUnauthorized, "user not authenticated")
		}
	}

	user, err := s.userUc.GetProfile(ctx, targetUserID)
	if err != nil {
		return nil, err
	}

	// 转换 Entity -> Proto
	return &nexusv1.GetProfileResponse{
		UserId:    user.ID,
		Username:  user.Username,
		Nickname:  user.Nickname,
		Bio:       user.Bio,
		AvatarKey: user.AvatarKey,
	}, nil
}

// ==========================================
// 更新资料 (UpdateProfile)
// ==========================================

func (s *UserService) UpdateProfile(ctx context.Context, req *nexusv1.UpdateProfileRequest) (*nexusv1.UpdateProfileResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, userWriteTimeout)
	defer cancel()

	userID := contextx.UserIDFromContext(ctx)
	if userID == 0 {
		return nil, xerr.New(xerr.CodeUnauthorized, "user not authenticated")
	}

	input := &biz.UpdateProfileInput{
		UserID:    userID,
		Nickname:  req.GetNickname(),
		Bio:       req.GetBio(),
		AvatarKey: req.GetAvatarKey(),
		Theme:     req.GetTheme(),
		Language:  req.GetLanguage(),
	}

	if err := s.userUc.UpdateProfile(ctx, input); err != nil {
		return nil, err
	}

	return &nexusv1.UpdateProfileResponse{Success: true}, nil
}

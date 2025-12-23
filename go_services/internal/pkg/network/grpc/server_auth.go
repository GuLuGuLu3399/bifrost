package grpc

import (
	"context"
	"strconv"
	"strings"

	"github.com/gulugulu3399/bifrost/internal/pkg/contextx"
	"github.com/gulugulu3399/bifrost/internal/pkg/security"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// AuthInterceptor JWT 认证拦截器
// jwtManager: 负责 Token 校验逻辑 (来自 security 包)。当 jwtManager == nil 时，拦截器会直接放行（用于公开服务复用拦截器链）。
// publicMethods: 不需要登录的接口列表
// adminMethods: 纯管理员接口列表（静态 RBAC）。混合权限接口（如 DeleteComment）不应放在这里。
func AuthInterceptor(jwtManager *security.JWTManager, publicMethods map[string]struct{}, adminMethods map[string]struct{}) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		// 0. 未启用鉴权：直接放行
		if jwtManager == nil {
			return handler(ctx, req)
		}

		// 0.5 健康检查放行，避免探活请求被拦截
		if info.FullMethod == "/grpc.health.v1.Health/Check" || info.FullMethod == "/grpc.health.v1.Health/Watch" {
			return handler(ctx, req)
		}

		// 1. 白名单检查 (无需登录直接放行)
		if _, ok := publicMethods[info.FullMethod]; ok {
			return handler(ctx, req)
		}

		// 2. 从 Context 获取 Token
		// 注意：这里的 Token 是由前置的 ContextInterceptor 通过 contextx.FromMD 提取出来的
		tokenStr := contextx.TokenFromContext(ctx)
		if tokenStr == "" {
			return nil, status.Error(codes.Unauthenticated, "missing authorization token")
		}

		// 处理 "Bearer " 前缀
		// 很多前端库会自动加这个前缀，后端需要兼容处理
		// 这里虽然JWTManager.ValidateToken也能处理，但我们提前处理更清晰
		tokenStr = strings.TrimPrefix(tokenStr, "Bearer ")
		tokenStr = strings.TrimSpace(tokenStr)

		// 3. 校验 Token
		claims, err := jwtManager.ValidateToken(tokenStr)
		if err != nil {
			return nil, status.Errorf(codes.Unauthenticated, "invalid token: %v", err)
		}

		// 4. [静态 RBAC] 纯管理员接口硬校验
		if _, ok := adminMethods[info.FullMethod]; ok {
			if !claims.IsAdmin {
				return nil, status.Error(codes.PermissionDenied, "admin privilege required")
			}
		}

		// 5. 注入 UserID/IsAdmin
		// claims.UserID 是 string，我们需要转为 int64 以便后续 Service 使用
		if uid, err := strconv.ParseInt(claims.UserID, 10, 64); err == nil {
			ctx = contextx.WithUserID(ctx, uid)
		} else {
			return nil, status.Error(codes.Internal, "invalid user id format in token")
		}
		ctx = contextx.WithIsAdmin(ctx, claims.IsAdmin)

		// 6. 继续处理请求
		return handler(ctx, req)
	}
}

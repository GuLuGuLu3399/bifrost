package grpc

import (
	"context"

	"github.com/gulugulu3399/bifrost/internal/pkg/contextx"
	"google.golang.org/grpc"
)

// ServerContextInterceptor 将 gRPC Metadata 还原为 Go Context
// 这是 ClientContextInterceptor 的逆操作
//
// 职责：
// 1. 从 gRPC Incoming Metadata 中提取 x-user-id, authorization, x-request-id 等
// 2. 注入到 Go Context，供后续业务逻辑使用
//
// 调用链路:
// - Gjallar (HTTP Gateway) 收到请求 -> 注入 Context
// - Gjallar 调用 Nexus/Beacon (通过 ClientContextInterceptor 转为 Metadata)
// - Nexus/Beacon (通过本拦截器还原为 Context)
func ServerContextInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		// 核心逻辑：调用 contextx.FromMD
		// 它会提取 x-user-id, authorization, x-request-id, x-locale, x-is-admin 等
		newCtx := contextx.FromMD(ctx)
		return handler(newCtx, req)
	}
}

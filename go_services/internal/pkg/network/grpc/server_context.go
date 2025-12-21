package grpc

import (
	"context"

	"github.com/gulugulu3399/bifrost/internal/pkg/contextx"
	"google.golang.org/grpc"
)

// ContextInterceptor 上下文适配拦截器
// 职责：将 gRPC Metadata 转换为 contextx 标准上下文
func ContextInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		// 调用 contextx.FromMD，它会自动提取 Token, RequestID, Locale 等
		// 此时 UserID 还未提取（因为 Token 还没解密），但 Token 字符串已经就位了
		newCtx := contextx.FromMD(ctx)

		return handler(newCtx, req)
	}
}

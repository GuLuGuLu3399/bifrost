package grpc

import (
	"context"
	"runtime/debug"

	"github.com/gulugulu3399/bifrost/internal/pkg/observability/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// RecoveryInterceptor 崩溃恢复拦截器
// 接收 logger.Logger 接口
func RecoveryInterceptor(l logger.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		defer func() {
			if r := recover(); r != nil {
				// 1. 获取堆栈信息
				stack := string(debug.Stack())

				// 2. 使用传入的 l 实例记录日志
				// 既然我们已经拿到了 stack 字符串，直接用 logger.String 构造字段即可
				// 不需要调用 logger.ErrorWithStack (它只服务于全局 logger)
				l.WithContext(ctx).Error("RPC Server Panic Recovered",
					logger.String("method", info.FullMethod),
					logger.Any("panic", r),
					logger.String("stack", stack),
				)

				// 3. 返回统一的内部错误
				err = status.Errorf(codes.Internal, "internal server error: %v", r)
			}
		}()

		return handler(ctx, req)
	}
}

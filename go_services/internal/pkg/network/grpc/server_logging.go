package grpc

import (
	"context"
	"time"

	"github.com/gulugulu3399/bifrost/internal/pkg/observability/logger" // 引用你的 logger 包
	"google.golang.org/grpc"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

// LoggingInterceptor 访问日志拦截器
// 接收 logger.Logger 接口，完全解耦具体实现
func LoggingInterceptor(l logger.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		start := time.Now()

		// 执行业务逻辑
		resp, err := handler(ctx, req)

		duration := time.Since(start)
		code := status.Code(err)

		// 构建基础字段
		fields := []logger.Field{
			logger.String("method", info.FullMethod),
			logger.String("code", code.String()),
			logger.Duration("duration", duration),
		}

		// 提取 Peer IP
		if p, ok := peer.FromContext(ctx); ok && p.Addr != nil {
			fields = append(fields, logger.String("peer", p.Addr.String()))
		}

		// 如果有错误，记录 error 字段
		if err != nil {
			fields = append(fields, logger.Err(err))
		}

		// 使用 WithContext 自动注入 TraceID, UserID, RequestID
		// 根据结果状态选择日志级别
		logEntry := l.WithContext(ctx)

		if err != nil {
			// 对于非 OK 请求，使用 Warn 级别引起注意
			logEntry.Warn("gRPC Access Fail", fields...)
		} else {
			logEntry.Info("gRPC Access Success", fields...)
		}

		return resp, err
	}
}

package grpc

import (
	"context"
	"time"

	"github.com/gulugulu3399/bifrost/internal/pkg/contextx"
	"github.com/gulugulu3399/bifrost/internal/pkg/observability/logger"
	"google.golang.org/grpc"
)

// ClientContextInterceptor 客户端上下文拦截器
// 职责：将 contextx 中的关键信息 (TraceID, Token, UserID) 注入到 gRPC Metadata 发送给下游
func ClientContextInterceptor() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		// 使用 contextx.ToMD 将 Go Context 转换为 gRPC Metadata
		ctx = contextx.ToMD(ctx)
		return invoker(ctx, method, req, reply, cc, opts...)
	}
}

// ClientLoggingInterceptor 客户端日志拦截器
// 职责：记录发出的请求耗时和结果
func ClientLoggingInterceptor(l logger.Logger) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		start := time.Now()

		// 执行请求
		err := invoker(ctx, method, req, reply, cc, opts...)

		duration := time.Since(start)

		// 构建日志字段
		fields := []logger.Field{
			logger.String("method", method),
			logger.String("target", cc.Target()), // 记录目标地址
			logger.Duration("duration", duration),
		}

		// 自动注入 contextx 中的 trace_id 等信息
		logEntry := l.WithContext(ctx)

		if err != nil {
			fields = append(fields, logger.Err(err))
			logEntry.Warn("gRPC Client Request Fail", fields...)
		} else {
			logEntry.Debug("gRPC Client Request Success", fields...)
		}
		return err
	}
}

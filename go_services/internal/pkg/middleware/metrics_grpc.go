package middleware

import (
	"context"
	"time"

	"github.com/gulugulu3399/bifrost/internal/pkg/observability/metrics"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

var (
	grpcReqCounter = metrics.NewCounterVec(
		"grpc_requests_total",
		"Total number of gRPC requests",
		[]string{"service", "method", "code"},
	)
	grpcReqLatency = metrics.NewHistogramVec(
		"grpc_request_duration_seconds",
		"Latency of gRPC requests in seconds",
		[]float64{0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2, 5},
		[]string{"service", "method"},
	)
)

// MetricsUnaryServerInterceptor records basic Prometheus metrics for gRPC unary RPCs.
func MetricsUnaryServerInterceptor(serviceName string) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		start := time.Now()
		resp, err = handler(ctx, req)

		code := status.Code(err).String()
		grpcReqCounter.WithLabelValues(serviceName, info.FullMethod, code).Inc()
		grpcReqLatency.WithLabelValues(serviceName, info.FullMethod).Observe(time.Since(start).Seconds())

		return resp, err
	}
}

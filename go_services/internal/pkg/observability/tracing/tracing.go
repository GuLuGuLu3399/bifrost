package tracing

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

// Init 初始化全局 TracerProvider
// 这里仅做最基础的设置，实际的 Exporter (如 Jaeger, OTLP) 应在 main 中配置
func Init() {
	// 设置全局传播器，用于跨服务传递 TraceContext
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))
}

// Start 开始一个新的 Span
func Start(ctx context.Context, name string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	// 使用全局 Tracer
	tracer := otel.Tracer("bifrost/internal/pkg/observability/tracing")
	return tracer.Start(ctx, name, opts...)
}

// SpanFromContext 从 Context 获取 Span
func SpanFromContext(ctx context.Context) trace.Span {
	return trace.SpanFromContext(ctx)
}

// Inject 将 TraceContext 注入到 carrier (如 HTTP Header)
func Inject(ctx context.Context, carrier propagation.TextMapCarrier) {
	otel.GetTextMapPropagator().Inject(ctx, carrier)
}

// Extract 从 carrier 提取 TraceContext
func Extract(ctx context.Context, carrier propagation.TextMapCarrier) context.Context {
	return otel.GetTextMapPropagator().Extract(ctx, carrier)
}

// TraceIDFromContext 从 Context 获取 TraceID (通常用于日志关联)
func TraceIDFromContext(ctx context.Context) string {
	span := trace.SpanFromContext(ctx)
	if span.SpanContext().IsValid() {
		return span.SpanContext().TraceID().String()
	}
	return ""
}

// SpanIDFromContext 向 Context 注入 SpanID
func SpanIDFromContext(ctx context.Context) string {
	span := trace.SpanFromContext(ctx)
	if span.SpanContext().IsValid() {
		return span.SpanContext().SpanID().String()
	}
	return ""
}

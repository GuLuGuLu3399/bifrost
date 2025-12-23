package tracing

import (
	"context"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"go.opentelemetry.io/otel/trace"
)

// InitProvider 初始化全局 TracerProvider 并返回一个关闭函数
// serviceName: 服务名 (如 "bifrost-nexus")
// collectorAddr: Jaeger OTLP gRPC 地址 (通常是 "localhost:4317" 或 "jaeger:4317")
//
// 使用示例:
//
//	shutdown, err := tracing.InitProvider(ctx, "bifrost-nexus", "localhost:4317")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer shutdown(context.Background())
func InitProvider(ctx context.Context, serviceName, collectorAddr string) (func(context.Context) error, error) {
	// 1. 创建 OTLP Exporter (通过 gRPC 发送 Trace 数据)
	exporter, err := otlptracegrpc.New(ctx,
		otlptracegrpc.WithEndpoint(collectorAddr),
		otlptracegrpc.WithInsecure(), // 本地开发通常不加密
		otlptracegrpc.WithTimeout(2*time.Second),
	)
	if err != nil {
		return nil, err
	}

	// 2. 创建资源属性 (标识这是哪个服务)
	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName(serviceName),
		),
	)
	if err != nil {
		return nil, err
	}

	// 3. 创建 TracerProvider
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter), // 批量发送，性能更好
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.AlwaysSample()), // 开发环境：全部采样 (生产环境需调整)
	)

	// 4. 设置全局 Provider 和 Propagator
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{}, // 支持 W3C Trace Context
		propagation.Baggage{},
	))

	// 返回 Shutdown 函数供 main.go defer 调用
	return tp.Shutdown, nil
}

// Init 初始化全局 TracerProvider (简化版，无 Exporter)
// 这个版本只设置传播器，不发送数据到后端
// 适用于不需要实际 Tracing 但需要传播 TraceID 的场景
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

// InitProviderWithDefault 是 InitProvider 的便捷封装：
// - 如果 endpoint 为空，使用 defaultEndpoint
// - usedEndpoint 返回最终使用的地址
// - 出错时返回一个 no-op shutdown，供调用方保持 defer 结构一致
func InitProviderWithDefault(ctx context.Context, serviceName string, endpoint string, defaultEndpoint string) (shutdown func(context.Context) error, usedEndpoint string, err error) {
	usedEndpoint = endpoint
	if usedEndpoint == "" {
		usedEndpoint = defaultEndpoint
	}

	shutdown, err = InitProvider(ctx, serviceName, usedEndpoint)
	if err != nil {
		return func(context.Context) error { return nil }, usedEndpoint, err
	}
	return shutdown, usedEndpoint, nil
}

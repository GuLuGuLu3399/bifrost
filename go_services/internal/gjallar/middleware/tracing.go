package middleware

import (
	"net/http"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

// Tracing 为 HTTP 请求自动生成 Span
//
// 职责：
// 1. 在 HTTP 请求进入时创建 Root Span
// 2. 自动提取 HTTP Header 中的 TraceContext (如果有)
// 3. 自动注入 TraceContext 到 Go Context
//
// 使用示例:
//
//	handler = middleware.Tracing("bifrost-gjallar")(handler)
func Tracing(serviceName string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return otelhttp.NewHandler(next, "HTTP Request",
			otelhttp.WithSpanNameFormatter(func(operation string, r *http.Request) string {
				// 自定义 Span 名称：METHOD /path
				return r.Method + " " + r.URL.Path
			}),
		)
	}
}

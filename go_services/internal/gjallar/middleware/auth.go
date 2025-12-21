package middleware

import (
	"net/http"
	"strings"

	"github.com/gulugulu3399/bifrost/internal/pkg/contextx"
)

// Auth 鉴权中间件
// 职责：解析 HTTP Header -> 注入 contextx -> 后续由 pkg/grpc 客户端拦截器自动转为 Metadata
func Auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := strings.TrimSpace(r.Header.Get("Authorization"))
		if authHeader == "" {
			// 匿名访问，直接放行 (后端业务层决定是否报错)
			next.ServeHTTP(w, r)
			return
		}

		// 处理 "Bearer <token>" 前缀（大小写不敏感）
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			// 不符合规范就当匿名透传，但不往 context 注入，避免后端误判
			next.ServeHTTP(w, r)
			return
		}
		tokenStr := strings.TrimSpace(parts[1])
		if tokenStr == "" {
			next.ServeHTTP(w, r)
			return
		}

		// 透传原始 Authorization（保持单一事实来源），供 gRPC client interceptor 注入 metadata
		ctx := contextx.WithToken(r.Context(), "Bearer "+tokenStr)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

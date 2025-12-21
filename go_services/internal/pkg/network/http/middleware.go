package http

import (
	"net/http"
	"runtime/debug"
	"time"

	"github.com/google/uuid"
	"github.com/gulugulu3399/bifrost/internal/pkg/contextx"
	"github.com/gulugulu3399/bifrost/internal/pkg/observability/logger"
)

// Middleware 定义中间件类型
type Middleware func(http.Handler) http.Handler

// Chain 串联多个中间件
func Chain(middlewares ...Middleware) Middleware {
	return func(next http.Handler) http.Handler {
		for i := len(middlewares) - 1; i >= 0; i-- {
			next = middlewares[i](next)
		}
		return next
	}
}

// Recovery 恢复 panic 的中间件
func Recovery() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					stack := debug.Stack()
					logger.Error("HTTP panic recovered",
						logger.Any("error", err),
						logger.String("stack", string(stack)),
						logger.String("method", r.Method),
						logger.String("path", r.URL.Path),
					)
					http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}

// Logger 记录请求日志的中间件
func Logger() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// 包装 ResponseWriter 以捕获状态码
			wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

			next.ServeHTTP(wrapped, r)

			duration := time.Since(start)
			logger.Info("HTTP request",
				logger.String("method", r.Method),
				logger.String("path", r.URL.Path),
				logger.Int("status", wrapped.statusCode),
				logger.Duration("duration", duration),
				logger.String("remote_addr", r.RemoteAddr),
				logger.String("user_agent", r.UserAgent()),
			)
		})
	}
}

// RequestID 注入请求 ID 的中间件
func RequestID() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestID := r.Header.Get("X-Request-ID")
			if requestID == "" {
				requestID = uuid.New().String()
			}

			// 设置响应头
			w.Header().Set("X-Request-ID", requestID)

			// 注入到 contextx
			ctx := r.Context()
			ctx = contextx.WithRequestID(ctx, requestID)
			r = r.WithContext(ctx)

			next.ServeHTTP(w, r)
		})
	}
}

// CORS 跨域处理中间件
func CORS(allowedOrigins []string, allowedMethods []string, allowedHeaders []string) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")

			// 检查是否允许的来源
			allowed := false
			for _, o := range allowedOrigins {
				if o == "*" || o == origin {
					allowed = true
					break
				}
			}

			if allowed {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Access-Control-Allow-Credentials", "true")

				if len(allowedMethods) > 0 {
					w.Header().Set("Access-Control-Allow-Methods", joinStrings(allowedMethods))
				}
				if len(allowedHeaders) > 0 {
					w.Header().Set("Access-Control-Allow-Headers", joinStrings(allowedHeaders))
				}
			}

			// 处理预检请求
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// Timeout 请求超时中间件
func Timeout(timeout time.Duration) Middleware {
	return func(next http.Handler) http.Handler {
		return http.TimeoutHandler(next, timeout, "Request Timeout")
	}
}

// RateLimit 简单的限流中间件 (基于固定窗口)
// 生产环境建议使用更完善的限流方案，如 golang.org/x/time/rate
func RateLimit(requestsPerSecond int) Middleware {
	limiter := newSimpleLimiter(requestsPerSecond)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !limiter.Allow() {
				http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// responseWriter 包装器，用于捕获状态码
type responseWriter struct {
	http.ResponseWriter
	statusCode int
	written    bool
}

func (rw *responseWriter) WriteHeader(code int) {
	if !rw.written {
		rw.statusCode = code
		rw.written = true
	}
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	if !rw.written {
		rw.written = true
	}
	return rw.ResponseWriter.Write(b)
}

// joinStrings 连接字符串数组
func joinStrings(strs []string) string {
	if len(strs) == 0 {
		return ""
	}
	result := strs[0]
	for i := 1; i < len(strs); i++ {
		result += ", " + strs[i]
	}
	return result
}

// simpleLimiter 简单的令牌桶限流器
type simpleLimiter struct {
	rate      int
	tokens    int
	lastCheck time.Time
}

func newSimpleLimiter(rate int) *simpleLimiter {
	return &simpleLimiter{
		rate:      rate,
		tokens:    rate,
		lastCheck: time.Now(),
	}
}

func (l *simpleLimiter) Allow() bool {
	now := time.Now()
	elapsed := now.Sub(l.lastCheck)
	l.lastCheck = now

	// 补充令牌
	l.tokens += int(elapsed.Seconds() * float64(l.rate))
	if l.tokens > l.rate {
		l.tokens = l.rate
	}

	if l.tokens > 0 {
		l.tokens--
		return true
	}
	return false
}

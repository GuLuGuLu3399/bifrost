package contextx

import (
	"context"
	"strconv"
	"strings"

	"google.golang.org/grpc/metadata"
)

// 私有类型，防止 Key 冲突
type ctxKey int

const (
	userIDKey ctxKey = iota
	requestIDKey
	tokenKey
	localeKey
	isAdminKey
)

// ==========================================
// 1. UserID 管理 (核心业务身份)
// ==========================================

// WithUserID 将 UserID 注入 Context
func WithUserID(ctx context.Context, uid int64) context.Context {
	return context.WithValue(ctx, userIDKey, uid)
}

// UserIDFromContext 从 Context 提取 UserID
// 如果提取失败或未设置，返回 0
func UserIDFromContext(ctx context.Context) int64 {
	v, ok := ctx.Value(userIDKey).(int64)
	if !ok {
		return 0
	}
	return v
}

// ==========================================
// 2. RequestID 管理 (日志链路追踪)
// ==========================================

// WithRequestID 将 RequestID 注入 Context
func WithRequestID(ctx context.Context, rid string) context.Context {
	return context.WithValue(ctx, requestIDKey, rid)
}

// RequestIDFromContext 从 Context 提取 RequestID
func RequestIDFromContext(ctx context.Context) string {
	v, _ := ctx.Value(requestIDKey).(string)
	return v
}

// ==========================================
// 3. Token 管理 (透传原始凭证)
// ==========================================

// WithToken 将 Token 注入 Context
// 注意：这里只存储原始字符串(如 "Bearer xxx")，不负责解析
func WithToken(ctx context.Context, token string) context.Context {
	return context.WithValue(ctx, tokenKey, token)
}

// TokenFromContext 从 Context 提取 Token
func TokenFromContext(ctx context.Context) string {
	v, _ := ctx.Value(tokenKey).(string)
	return v
}

// ==========================================
// 4. Locale 管理 (国际化)
// ==========================================

// WithLocale 将 Locale 注入 Context
func WithLocale(ctx context.Context, locale string) context.Context {
	return context.WithValue(ctx, localeKey, locale)
}

// LocaleFromContext 从 Context 提取 Locale
func LocaleFromContext(ctx context.Context) string {
	v, _ := ctx.Value(localeKey).(string)
	return v
}

// ==========================================
// 5. Admin 管理 (权限标记)
// ==========================================

// WithIsAdmin 将是否管理员注入 Context
func WithIsAdmin(ctx context.Context, isAdmin bool) context.Context {
	return context.WithValue(ctx, isAdminKey, isAdmin)
}

// IsAdminFromContext 从 Context 提取管理员标记
func IsAdminFromContext(ctx context.Context) bool {
	v, ok := ctx.Value(isAdminKey).(bool)
	if !ok {
		return false
	}
	return v
}

// ==========================================
// 6. gRPC Metadata 互操作 (拦截器专用)
// ==========================================

// FromMD 从 gRPC Metadata 提取关键信息并注入 Context
// 通常在 gRPC Server Interceptor (ContextInterceptor) 中调用
func FromMD(ctx context.Context) context.Context {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ctx
	}

	// 1. 提取 UserID (gRPC Header 传输的是 string，需转回 int64)
	// 注意：通常 UserID 是由 AuthInterceptor 解析 Token 后注入的，
	// 但如果是服务间内部调用直接透传 UserID，这里也会生效。
	if vals := md.Get("x-user-id"); len(vals) > 0 {
		if uid, err := strconv.ParseInt(vals[0], 10, 64); err == nil {
			ctx = WithUserID(ctx, uid)
		}
	}

	// 2. 提取 RequestID
	if vals := md.Get("x-request-id"); len(vals) > 0 {
		ctx = WithRequestID(ctx, vals[0])
	}

	// 3. 提取 Token (Authorization)
	if vals := md.Get("authorization"); len(vals) > 0 {
		ctx = WithToken(ctx, vals[0])
	}

	// 4. 提取 Locale
	if vals := md.Get("x-locale"); len(vals) > 0 {
		ctx = WithLocale(ctx, vals[0])
	}

	// 5. 提取 IsAdmin
	// 支持两种形态：
	// - x-is-admin: "true"/"false"/"1"/"0"
	// - x-role: "admin" (逗号分隔)
	if vals := md.Get("x-is-admin"); len(vals) > 0 {
		b, err := strconv.ParseBool(vals[0])
		if err == nil {
			ctx = WithIsAdmin(ctx, b)
		}
	} else if vals := md.Get("x-role"); len(vals) > 0 {
		roles := strings.Split(vals[0], ",")
		for _, r := range roles {
			if strings.EqualFold(strings.TrimSpace(r), "admin") {
				ctx = WithIsAdmin(ctx, true)
				break
			}
		}
	}

	return ctx
}

// ToMD 将 Context 中的关键信息写入 gRPC Outgoing Metadata
// 通常在 gRPC Client Interceptor 中调用，用于透传给下游服务
func ToMD(ctx context.Context) context.Context {
	var pairs []string

	// 1. 写入 UserID (int64 -> string)
	if uid := UserIDFromContext(ctx); uid != 0 {
		pairs = append(pairs, "x-user-id", strconv.FormatInt(uid, 10))
	}

	// 2. 写入 RequestID
	if rid := RequestIDFromContext(ctx); rid != "" {
		pairs = append(pairs, "x-request-id", rid)
	}

	// 3. 写入 Token
	if token := TokenFromContext(ctx); token != "" {
		pairs = append(pairs, "authorization", token)
	}

	// 4. 写入 Locale
	if locale := LocaleFromContext(ctx); locale != "" {
		pairs = append(pairs, "x-locale", locale)
	}

	if len(pairs) == 0 {
		return ctx
	}

	return metadata.AppendToOutgoingContext(ctx, pairs...)
}

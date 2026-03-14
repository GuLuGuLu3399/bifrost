package router

import (
	"context"
	"net/http"
	"strings"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"

	beaconv1 "github.com/gulugulu3399/bifrost/api/content/v1/beacon"
	nexusv1 "github.com/gulugulu3399/bifrost/api/content/v1/nexus"
	"github.com/gulugulu3399/bifrost/internal/gjallar/handler"
	grpcClient "github.com/gulugulu3399/bifrost/internal/gjallar/infrastructure/grpc"
	"github.com/gulugulu3399/bifrost/internal/pkg/observability/logger"
)

// New 创建 HTTP Router
// 使用 gRPC Gateway 自动生成的 Register*Handler 适配器
func New(ctx context.Context, nexusConn *grpc.ClientConn, beaconConn *grpc.ClientConn, mirrorClient *grpcClient.MirrorClient) (http.Handler, error) {
	mux := runtime.NewServeMux(
		// 错误处理中间件
		runtime.WithErrorHandler(customErrorHandler),
	)

	// 直接复用已建立的连接注册 handler（避免重复 Dial）

	// 注册 Beacon 服务的 HTTP 处理器
	if err := beaconv1.RegisterBeaconServiceHandler(ctx, mux, beaconConn); err != nil {
		logger.Global().Error("Failed to register BeaconService handler", logger.Err(err))
		return nil, err
	}

	// 注册 Nexus 用户服务的 HTTP 处理器
	if err := nexusv1.RegisterUserServiceHandler(ctx, mux, nexusConn); err != nil {
		logger.Global().Error("Failed to register UserService handler", logger.Err(err))
		return nil, err
	}

	// 注册 Nexus 文章服务的 HTTP 处理器
	if err := nexusv1.RegisterPostServiceHandler(ctx, mux, nexusConn); err != nil {
		logger.Global().Error("Failed to register PostService handler", logger.Err(err))
		return nil, err
	}

	// 注册 Nexus 评论服务的 HTTP 处理器
	if err := nexusv1.RegisterCommentServiceHandler(ctx, mux, nexusConn); err != nil {
		logger.Global().Error("Failed to register CommentService handler", logger.Err(err))
		return nil, err
	}

	// 注册 Nexus 标签服务的 HTTP 处理器
	if err := nexusv1.RegisterTagServiceHandler(ctx, mux, nexusConn); err != nil {
		logger.Global().Error("Failed to register TagService handler", logger.Err(err))
		return nil, err
	}

	// 注册 Nexus 分类服务的 HTTP 处理器
	if err := nexusv1.RegisterCategoryServiceHandler(ctx, mux, nexusConn); err != nil {
		logger.Global().Error("Failed to register CategoryService handler", logger.Err(err))
		return nil, err
	}

	// 创建搜索处理器（使用 Mirror 客户端）
	searchHandler := handler.NewSearchHandler(mirrorClient)
	suggestHandler := handler.NewSuggestHandler(mirrorClient)

	// 创建一个多路复用器来组合 gRPC Gateway 和自定义 HTTP 处理器
	rootMux := http.NewServeMux()
	
	// 注册搜索端点
	rootMux.Handle("/v1/search", searchHandler)
	rootMux.Handle("/v1/search/suggest", suggestHandler)
	
	// 其他所有路由由 gRPC Gateway 处理
	rootMux.Handle("/", mux)

	return stripEmptyQueryParams(rootMux), nil
}

// stripEmptyQueryParams 移除 query string 中空值参数。
// grpc-gateway 对 int64/int32 字段不容忍空字符串（会触发 strconv.ParseInt 错误），
// 这里把 "" / 空白 / null / undefined 都视为“未传”。
func stripEmptyQueryParams(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		changed := false
		for k, vals := range q {
			filtered := vals[:0]
			for _, v := range vals {
				nv, keep := normalizeQueryValue(v)
				if !keep {
					continue
				}
				filtered = append(filtered, nv)
			}
			if len(filtered) != len(vals) {
				changed = true
				if len(filtered) == 0 {
					q.Del(k)
				} else {
					q[k] = filtered
				}
			}
		}
		if changed {
			r2 := r.Clone(r.Context())
			r2.URL.RawQuery = q.Encode()
			next.ServeHTTP(w, r2)
		} else {
			next.ServeHTTP(w, r)
		}
	})
}

func normalizeQueryValue(v string) (string, bool) {
	n := strings.TrimSpace(v)
	if n == "" {
		return "", false
	}
	if strings.EqualFold(n, "null") || strings.EqualFold(n, "undefined") {
		return "", false
	}
	return n, true
}

// customErrorHandler 自定义错误处理
func customErrorHandler(ctx context.Context, mux *runtime.ServeMux, marshaler runtime.Marshaler, w http.ResponseWriter, r *http.Request, err error) {
	logger.Global().Warn("gateway error", logger.Err(err))
	// 这里可以自定义错误响应格式
	runtime.DefaultHTTPErrorHandler(ctx, mux, marshaler, w, r, err)
}

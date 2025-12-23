package router

import (
	"context"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"

	beaconv1 "github.com/gulugulu3399/bifrost/api/content/v1/beacon"
	nexusv1 "github.com/gulugulu3399/bifrost/api/content/v1/nexus"
	"github.com/gulugulu3399/bifrost/internal/pkg/observability/logger"
)

// New 创建 HTTP Router
// 使用 gRPC Gateway 自动生成的 Register*Handler 适配器
func New(ctx context.Context, nexusConn *grpc.ClientConn, beaconConn *grpc.ClientConn) (http.Handler, error) {
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

	return mux, nil
}

// customErrorHandler 自定义错误处理
func customErrorHandler(ctx context.Context, mux *runtime.ServeMux, marshaler runtime.Marshaler, w http.ResponseWriter, r *http.Request, err error) {
	logger.Global().Warn("gateway error", logger.Err(err))
	// 这里可以自定义错误响应格式
	runtime.DefaultHTTPErrorHandler(ctx, mux, marshaler, w, r, err)
}

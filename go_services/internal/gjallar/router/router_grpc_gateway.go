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
		// 使用 proto 名称而不是 camelCase
		runtime.WithMarshalerOption(runtime.MIMETypeJSON, &runtime.JSONPb{
			UseProtoNames:   true,
			EmitUnpopulated: true,
		}),
		// 错误处理中间件
		runtime.WithErrorHandler(customErrorHandler),
	)

	// 配置 gRPC 连接选项
	opts := []grpc.DialOption{grpc.WithInsecure()}

	// 注册 Beacon 服务的 HTTP 处理器
	if err := beaconv1.RegisterBeaconServiceHandlerFromEndpoint(ctx, mux, beaconConn.Target(), opts); err != nil {
		logger.Global().Error("Failed to register BeaconService handler", logger.Err(err))
		return nil, err
	}

	// 注册 Nexus 用户服务的 HTTP 处理器
	if err := nexusv1.RegisterUserServiceHandlerFromEndpoint(ctx, mux, nexusConn.Target(), opts); err != nil {
		logger.Global().Error("Failed to register UserService handler", logger.Err(err))
		return nil, err
	}

	// 注册 Nexus 文章服务的 HTTP 处理器
	if err := nexusv1.RegisterPostServiceHandlerFromEndpoint(ctx, mux, nexusConn.Target(), opts); err != nil {
		logger.Global().Error("Failed to register PostService handler", logger.Err(err))
		return nil, err
	}

	// 注册 Nexus 评论服务的 HTTP 处理器
	if err := nexusv1.RegisterCommentServiceHandlerFromEndpoint(ctx, mux, nexusConn.Target(), opts); err != nil {
		logger.Global().Error("Failed to register CommentService handler", logger.Err(err))
		return nil, err
	}

	// 注册 Nexus 标签服务的 HTTP 处理器
	if err := nexusv1.RegisterTagServiceHandlerFromEndpoint(ctx, mux, nexusConn.Target(), opts); err != nil {
		logger.Global().Error("Failed to register TagService handler", logger.Err(err))
		return nil, err
	}

	// 注册 Nexus 分类服务的 HTTP 处理器
	if err := nexusv1.RegisterCategoryServiceHandlerFromEndpoint(ctx, mux, nexusConn.Target(), opts); err != nil {
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

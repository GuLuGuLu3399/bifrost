package service

import (
	"google.golang.org/grpc"

	nexusv1 "github.com/gulugulu3399/bifrost/api/content/v1/nexus"
	"github.com/gulugulu3399/bifrost/internal/nexus/biz"
	"github.com/gulugulu3399/bifrost/internal/pkg/messenger"
	pkggrpc "github.com/gulugulu3399/bifrost/internal/pkg/network/grpc"
	"github.com/gulugulu3399/bifrost/internal/pkg/security"
)

// App 负责把 Nexus 的 Service 层实现注册到 gRPC Server。
// 这样 cmd/main.go 只需要关心基础设施初始化（config/db/logger 等），
// 不需要了解每个 gRPC service 的装配细节。
type App struct {
	jwtManager *security.JWTManager
	msgr       *messenger.Client

	userUC    *biz.UserUseCase
	postUC    *biz.PostUseCase
	tagUC     *biz.TagUseCase
	catUC     *biz.CategoryUseCase
	commentUC *biz.CommentUseCase
}

func NewApp(
	jwtManager *security.JWTManager,
	msgr *messenger.Client,
	userUC *biz.UserUseCase,
	postUC *biz.PostUseCase,
	tagUC *biz.TagUseCase,
	catUC *biz.CategoryUseCase,
	commentUC *biz.CommentUseCase,
) *App {
	return &App{
		jwtManager: jwtManager,
		msgr:       msgr,
		userUC:     userUC,
		postUC:     postUC,
		tagUC:      tagUC,
		catUC:      catUC,
		commentUC:  commentUC,
	}
}

// RegisterGRPC 把所有 gRPC service 注册到 server 上，并返回服务端需要的 unary interceptors。
//
// 用法：
// - 想拿到拦截器列表但暂时不注册服务时，可以传 nil。
// - 想真正注册服务时，传入 *grpc.Server。
//
// 说明：
// - ContextInterceptor 必须在 AuthInterceptor 之前（Auth 依赖 contextx.TokenFromContext）。
// - adminMethods 只处理“纯管理员接口”；混合权限（如 DeleteComment）仍在 Service/Biz 判断。
func (a *App) RegisterGRPC(server grpc.ServiceRegistrar) []grpc.UnaryServerInterceptor {
	// 1) 注册 gRPC Services（server 允许为 nil）
	if server != nil {
		nexusv1.RegisterUserServiceServer(server, NewUserService(a.userUC, a.jwtManager))
		nexusv1.RegisterPostServiceServer(server, NewPostService(a.postUC, a.msgr))
		nexusv1.RegisterTagServiceServer(server, NewTagService(a.tagUC))
		nexusv1.RegisterCategoryServiceServer(server, NewCategoryService(a.catUC))
		nexusv1.RegisterCommentServiceServer(server, NewCommentService(a.commentUC))
	}

	// 2) Interceptors
	publicMethods := map[string]struct{}{
		"/bifrost.content.v1.nexus.UserService/Register": {},
		"/bifrost.content.v1.nexus.UserService/Login":    {},
	}

	adminMethods := map[string]struct{}{
		"/bifrost.content.v1.nexus.CategoryService/CreateCategory": {},
		"/bifrost.content.v1.nexus.CategoryService/UpdateCategory": {},
		"/bifrost.content.v1.nexus.CategoryService/DeleteCategory": {},
		"/bifrost.content.v1.nexus.TagService/DeleteTag":           {},
	}

	return []grpc.UnaryServerInterceptor{
		pkggrpc.ContextInterceptor(),
		pkggrpc.AuthInterceptor(a.jwtManager, publicMethods, adminMethods),
	}
}

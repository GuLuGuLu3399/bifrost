package service

import (
	"context"

	beaconv1 "github.com/gulugulu3399/bifrost/api/content/v1/beacon"
	"github.com/gulugulu3399/bifrost/internal/beacon/biz"
	"google.golang.org/grpc"
)

// App 负责聚合所有独立的领域服务，并作为 BeaconServiceServer 的适配器。
type App struct {
	beaconv1.UnimplementedBeaconServiceServer

	Post    *PostService
	Comment *CommentService
	User    *UserService
	Meta    *MetaService
}

// NewApp 初始化所有独立的 Service
func NewApp(pr biz.PostRepo, ur biz.UserRepo, mr biz.MetaRepo, cr biz.CommentRepo) *App {
	return &App{
		Post:    NewPostService(pr),
		Comment: NewCommentService(cr),
		User:    NewUserService(ur),
		Meta:    NewMetaService(mr),
	}
}

// RegisterGRPC 统一将服务注册到 gRPC Server
func (a *App) RegisterGRPC(s grpc.ServiceRegistrar) {
	beaconv1.RegisterBeaconServiceServer(s, a)
}

// =============================================================================
// BeaconServiceServer 方法转发
// =============================================================================

func (a *App) GetPost(ctx context.Context, req *beaconv1.GetPostRequest) (*beaconv1.GetPostResponse, error) {
	return a.Post.GetPost(ctx, req)
}

func (a *App) ListPosts(ctx context.Context, req *beaconv1.ListPostsRequest) (*beaconv1.ListPostsResponse, error) {
	return a.Post.ListPosts(ctx, req)
}

func (a *App) ListComments(ctx context.Context, req *beaconv1.ListCommentsRequest) (*beaconv1.ListCommentsResponse, error) {
	return a.Comment.ListComments(ctx, req)
}

func (a *App) GetUser(ctx context.Context, req *beaconv1.GetUserRequest) (*beaconv1.GetUserResponse, error) {
	return a.User.GetUser(ctx, req)
}

func (a *App) ListCategories(ctx context.Context, req *beaconv1.ListCategoriesRequest) (*beaconv1.ListCategoriesResponse, error) {
	return a.Meta.ListCategories(ctx, req)
}

func (a *App) ListTags(ctx context.Context, req *beaconv1.ListTagsRequest) (*beaconv1.ListTagsResponse, error) {
	return a.Meta.ListTags(ctx, req)
}

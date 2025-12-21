package server

import (
	"context"

	"github.com/gulugulu3399/bifrost/internal/gjallar/middleware"
	"github.com/gulugulu3399/bifrost/internal/gjallar/router"
	"github.com/gulugulu3399/bifrost/internal/pkg/config"
	pkggrpc "github.com/gulugulu3399/bifrost/internal/pkg/network/grpc"
	pkghttp "github.com/gulugulu3399/bifrost/internal/pkg/network/http"
	"github.com/gulugulu3399/bifrost/internal/pkg/observability/logger"
	"google.golang.org/grpc"
)

type GjallarServer struct {
	httpServer *pkghttp.Server
	nexusConn  *grpc.ClientConn // 持有连接以便关闭
	beaconConn *grpc.ClientConn
}

func New(cfg *config.Config) (*GjallarServer, error) {
	ctx := context.Background()
	l := logger.Global() // 使用全局 Logger

	// 1. 初始化 gRPC 客户端 (连接池、健康检查、拦截器)
	nexusConn, err := pkggrpc.NewClient(cfg.NexusRPC, l)
	if err != nil {
		return nil, err
	}
	beaconConn, err := pkggrpc.NewClient(cfg.BeaconRPC, l)
	if err != nil {
		return nil, err
	}

	// 2. 初始化 Router (Gateway)
	mux, err := router.New(ctx, nexusConn, beaconConn)
	if err != nil {
		return nil, err
	}

	// 3. 组装中间件 (洋葱模型：外 -> 内)
	// CORS -> Recovery -> Logger -> TraceID -> Auth -> Gateway Mux
	var handler = mux

	// 应用 Gjallar 特有的业务中间件
	handler = middleware.Auth(handler)

	// 应用 pkg 提供的通用基础设施中间件
	handler = pkghttp.Chain(
		pkghttp.CORS([]string{"*"}, nil, nil), // 允许所有跨域
		pkghttp.RequestID(),                   // 生成 RequestID 并注入 Context
		pkghttp.Logger(),                      // 记录 HTTP 访问日志
		pkghttp.Recovery(),                    // 防止 Panic 挂掉
	)(handler)

	// 4. 创建 HTTP Server
	srv := pkghttp.NewServer(cfg.Http, handler)

	return &GjallarServer{
		httpServer: srv,
		nexusConn:  nexusConn,
		beaconConn: beaconConn,
	}, nil
}

// Run 启动服务
func (s *GjallarServer) Run() error {
	return s.httpServer.Start()
}

// Shutdown 优雅关闭
func (s *GjallarServer) Shutdown(ctx context.Context) error {
	// 1. 先停 HTTP 入口，不再接收新请求
	if err := s.httpServer.Stop(ctx); err != nil {
		return err
	}

	// 2. 关闭 gRPC 连接
	if s.nexusConn != nil {
		err := s.nexusConn.Close()
		if err != nil {
			logger.Global().Warn("Failed to close nexus connection")
			return err
		}
	}
	if s.beaconConn != nil {
		err := s.beaconConn.Close()
		if err != nil {
			logger.Global().Warn("Failed to close beacon connection")
			return err
		}
	}
	return nil
}

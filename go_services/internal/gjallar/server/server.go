package server

import (
	"context"
	"fmt"
	"strings"
	"time"

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

	// 为了避免在中间步骤出错时造成已创建连接泄漏，使用局部清理闭包
	var (
		nexusConn  *grpc.ClientConn
		beaconConn *grpc.ClientConn
		cleanupErr error
	)
	defer func() {
		// 如果函数以非 nil 错误返回，则清理已创建的连接
		if cleanupErr != nil {
			if nexusConn != nil {
				_ = nexusConn.Close()
			}
			if beaconConn != nil {
				_ = beaconConn.Close()
			}
		}
	}()

	// 1. 初始化 gRPC 客户端 (连接池、健康检查、拦截器)
	nConn, err := pkggrpc.NewClient(cfg.RPC.Nexus, l)
	if err != nil {
		cleanupErr = err
		return nil, err
	}
	nexusConn = nConn

	bConn, err := pkggrpc.NewClient(cfg.RPC.Beacon, l)
	if err != nil {
		cleanupErr = err
		return nil, err
	}
	beaconConn = bConn

	// 2. 初始化 Router (Gateway)
	mux, err := router.New(ctx, nexusConn, beaconConn)
	if err != nil {
		cleanupErr = err
		return nil, err
	}

	// 3. 组装中间件 (洋葱模型：外 -> 内)
	// Tracing -> CORS -> Recovery -> Logger -> TraceID -> Auth -> Gateway Mux
	var handler = mux

	// 应用 Gjallar 特有的业务中间件
	handler = middleware.Auth(handler)

	// 应用 pkg 提供的通用基础设施中间件
	handler = pkghttp.Chain(
		middleware.Tracing("bifrost-gjallar"), // ✅ Phase 1: 生成 Root Span (必须在最外层)
		pkghttp.CORS([]string{"*"}, nil, nil), // 允许所有跨域
		pkghttp.RequestID(),                   // 生成 RequestID 并注入 Context
		pkghttp.Logger(),                      // 记录 HTTP 访问日志
		pkghttp.Recovery(),                    // 防止 Panic 挂掉
	)(handler)

	// 4. 创建 HTTP Server（转换为通用 ServerConfig，补齐 IdleTimeout 等默认值）
	httpCfg := pkghttp.ServerConfig{
		Addr:         cfg.HTTP.Addr,
		ReadTimeout:  cfg.HTTP.ReadTimeout,
		WriteTimeout: cfg.HTTP.WriteTimeout,
		IdleTimeout:  cfg.HTTP.IdleTimeout,
		Timeout:      cfg.HTTP.Timeout,
	}
	if httpCfg.ReadTimeout == 0 {
		httpCfg.ReadTimeout = 5 * time.Second
	}
	if httpCfg.WriteTimeout == 0 {
		httpCfg.WriteTimeout = 10 * time.Second
	}
	if httpCfg.IdleTimeout == 0 {
		httpCfg.IdleTimeout = 60 * time.Second
	}
	if httpCfg.Timeout == 0 {
		httpCfg.Timeout = 5 * time.Second
	}

	srv := pkghttp.NewServer(httpCfg, handler)

	// 成功路径：清理闭包不再触发
	cleanupErr = nil

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
	var errs []string
	// 1. 先停 HTTP 入口，不再接收新请求
	if s.httpServer != nil {
		if err := s.httpServer.Stop(ctx); err != nil {
			logger.Global().Warn("http server stop failed", logger.Err(err))
			errs = append(errs, fmt.Sprintf("http stop: %v", err))
		}
	}

	// 2. 关闭 gRPC 连接（尽量都关闭，收集错误但不提前返回）
	if s.nexusConn != nil {
		if err := s.nexusConn.Close(); err != nil {
			logger.Global().Warn("Failed to close nexus connection", logger.Err(err))
			errs = append(errs, fmt.Sprintf("nexus close: %v", err))
		}
	}
	if s.beaconConn != nil {
		if err := s.beaconConn.Close(); err != nil {
			logger.Global().Warn("Failed to close beacon connection", logger.Err(err))
			errs = append(errs, fmt.Sprintf("beacon close: %v", err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("shutdown encountered errors: %s", errJoin(errs))
	}
	return nil
}

func errJoin(errs []string) string {
	if len(errs) == 0 {
		return ""
	}
	return strings.Join(errs, "; ")
}

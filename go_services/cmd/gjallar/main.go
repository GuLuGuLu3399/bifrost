package main

import (
	"context"
	"log"
	"os"

	"github.com/gulugulu3399/bifrost/internal/gjallar/server"
	"github.com/gulugulu3399/bifrost/internal/pkg/config"
	"github.com/gulugulu3399/bifrost/internal/pkg/lifecycle"
	"github.com/gulugulu3399/bifrost/internal/pkg/observability/logger"
	"github.com/gulugulu3399/bifrost/internal/pkg/observability/tracing"
)

func main() {
	// 1. 加载配置（支持环境变量覆盖）
	cfg, err := config.LoadGjallarConfig()
	if err != nil {
		cfg = config.DefaultConfig()
		log.Printf("Failed to load config, using defaults: %v", err)
	}

	// 2. 初始化 Logger (使用 Zap 保持一致性)
	zlog := logger.NewZap(cfg.LoggerConfig(), cfg.App.Name, cfg.App.Env)
	logger.SetGlobal(zlog)
	defer func() { _ = logger.Sync() }()

	logger.Info("Starting Gjallar Gateway",
		logger.String("http_addr", cfg.HTTP.Addr),
		logger.String("nexus_rpc", cfg.RPC.Nexus.Addr),
		logger.String("beacon_rpc", cfg.RPC.Beacon.Addr),
	)

	// Tracing
	shutdownTracer, usedEndpoint, err := tracing.InitProviderWithDefault(context.Background(), cfg.App.Name, cfg.Observability.OtlpEndpoint, "localhost:4317")
	if err != nil {
		logger.Warn("Failed to init tracer (non-fatal)", logger.Err(err))
	} else {
		logger.Info("Tracing initialized", logger.String("collector", usedEndpoint))
	}
	defer func() {
		if err := shutdownTracer(context.Background()); err != nil {
			logger.Error("Failed to shutdown tracer", logger.Err(err))
		}
	}()

	// 3. 创建服务器
	gjallarSrv, err := server.New(cfg)
	if err != nil {
		logger.Error("Failed to create Gjallar server", logger.Err(err))
		os.Exit(1)
	}

	// 4. 设置优雅关闭
	sh := lifecycle.NewShutdown()
	ctx, stop := sh.NotifyContext(context.Background())
	defer stop()

	// 5. 启动服务器（在 goroutine 中运行）
	go func() {
		logger.Info("Gjallar gateway listening", logger.String("addr", cfg.HTTP.Addr))
		if err := gjallarSrv.Run(); err != nil {
			logger.Error("Server error", logger.Err(err))
			stop() // 通知关闭
		}
	}()

	// 6. 等待关闭信号
	<-ctx.Done()

	// 7. 优雅关闭
	logger.Info("Shutting down Gjallar gateway...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.HTTP.Timeout)
	defer cancel()

	if err := gjallarSrv.Shutdown(shutdownCtx); err != nil {
		log.Printf("Shutdown error: %v\n", err)
		os.Exit(1)
	}

	logger.Info("Gjallar gateway stopped")
	os.Exit(0)
}

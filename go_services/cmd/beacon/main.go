//TODO metric

package main

import (
	"context"
	"flag"
	"net/http"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/gulugulu3399/bifrost/internal/beacon/data"
	"github.com/gulugulu3399/bifrost/internal/beacon/service"
	"github.com/gulugulu3399/bifrost/internal/pkg/cache"
	"github.com/gulugulu3399/bifrost/internal/pkg/config"
	"github.com/gulugulu3399/bifrost/internal/pkg/database"
	"github.com/gulugulu3399/bifrost/internal/pkg/lifecycle"
	"github.com/gulugulu3399/bifrost/internal/pkg/messenger"
	pkgmw "github.com/gulugulu3399/bifrost/internal/pkg/middleware"
	pkggrpc "github.com/gulugulu3399/bifrost/internal/pkg/network/grpc"
	"github.com/gulugulu3399/bifrost/internal/pkg/observability/logger"
	"github.com/gulugulu3399/bifrost/internal/pkg/observability/tracing"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
)

var configFile = flag.String("f", "configs/beacon.yaml", "the config file")

func main() {
	// Metrics endpoint (Prometheus)
	go func() {
		mux := http.NewServeMux()
		mux.Handle("/metrics", promhttp.Handler())
		// best-effort: bind to localhost:9102 by default
		if err := http.ListenAndServe("localhost:9102", mux); err != nil {
			logger.Warn("metrics server stopped", logger.Err(err))
		}
	}()
	flag.Parse()

	sh := lifecycle.NewShutdown()
	ctx, stop := sh.NotifyContext(context.Background())
	defer stop()

	cfg, err := config.LoadBeacon(*configFile)
	if err != nil {
		logger.Fatal("Failed to load config", logger.Any("error", err))
	}

	zlog := logger.NewZap(cfg.LoggerConfig(), cfg.App.Name, cfg.App.Env)
	logger.SetGlobal(zlog)
	defer func() { _ = logger.Sync() }()

	// Tracing
	shutdownTracer, usedEndpoint, err := tracing.InitProviderWithDefault(ctx, cfg.App.Name, cfg.Observability.OtlpEndpoint, "localhost:4317")
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

	// DB
	db, err := database.New(&database.Config{
		Driver:          "postgres",
		DSN:             cfg.Data.Database.DSN,
		MaxOpenConns:    cfg.Data.Database.MaxOpenConns,
		MaxIdleConns:    cfg.Data.Database.MaxIdleConns,
		ConnMaxLifetime: cfg.Data.Database.MaxLifetime,
		ConnMaxIdleTime: cfg.Data.Database.MaxLifetime,
	})
	if err != nil {
		logger.Fatal("Failed to connect to database", logger.Any("error", err))
	}
	sh.Register(db)

	// Redis
	rds, err := cache.New(&cache.Config{
		Addr:         cfg.Data.Redis.Addr,
		Password:     cfg.Data.Redis.Password,
		DB:           cfg.Data.Redis.DB,
		PoolSize:     cfg.Data.Redis.PoolSize,
		MinIdleConns: cfg.Data.Redis.MinIdleConns,
		DialTimeout:  cfg.Data.Redis.DialTimeout,
		ReadTimeout:  cfg.Data.Redis.ReadTimeout,
		WriteTimeout: cfg.Data.Redis.WriteTimeout,
	})
	if err != nil {
		logger.Fatal("Failed to connect to redis", logger.Any("error", err))
	}
	sh.Register(rds)

	// Data + Repos
	dd := data.NewData(db, rds)
	postRepo := data.NewPostRepo(dd)
	metaRepo := data.NewMetaRepo(dd)
	userRepo := data.NewUserRepo(dd)
	commentRepo := data.NewCommentRepo(dd)

	app := service.NewApp(postRepo, userRepo, metaRepo, commentRepo)

	// 统一使用 pkggrpc.Server
	g, err := pkggrpc.NewServer(pkggrpc.ServerConfig{
		Addr:             cfg.Server.GRPCAddr,
		Timeout:          cfg.Server.GracefulShutdownTimeout,
		EnableReflection: true,
		EnableHealth:     true,
		//TODO keepalive 这里先走默认零值（由 gRPC 自己使用默认策略）；后续可以从 config 补齐
	}, zlog, nil, nil, nil, grpc.ChainUnaryInterceptor(pkgmw.MetricsUnaryServerInterceptor(cfg.App.Name)))
	if err != nil {
		logger.Fatal("Failed to init gRPC server", logger.Any("error", err))
	}
	sh.Register(g) // 走 Close() -> Stop()

	app.RegisterGRPC(g.GRPC())

	// NATS messenger + Consumer
	msgr, err := messenger.New(cfg.Messenger.Addr)
	if err != nil {
		logger.Fatal("Failed to connect to NATS", logger.Any("error", err))
	}
	// 注册到 shutdown，统一关闭
	sh.Register(msgr)

	consumer := data.NewConsumer(dd, msgr)
	if err := consumer.Start(); err != nil {
		logger.Fatal("Failed to start beacon consumer", logger.Any("error", err))
	}
	sh.Register(consumer)

	errCh := g.StartAsync()

	select {
	case <-ctx.Done():
		logger.Info("Beacon shutting down")
		_ = sh.CloseAll()
	case err := <-errCh:
		if err != nil {
			logger.Error("Beacon server stopped with error", logger.Any("error", err))
		}
		_ = sh.CloseAll()
	}
}

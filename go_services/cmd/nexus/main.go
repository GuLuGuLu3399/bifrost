//TODO Metrics

package main

import (
	"context"
	"flag"
	"net/http"
	"log"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib" // PGX Driver for database/sql

	// API Clients
	forgev1 "github.com/gulugulu3399/bifrost/api/content/v1/forge"
	nexusv1 "github.com/gulugulu3399/bifrost/api/content/v1/nexus"

	// Business Layer
	"github.com/gulugulu3399/bifrost/internal/nexus/biz"

	// Data Layer
	"github.com/gulugulu3399/bifrost/internal/nexus/data"

	// Service Layer
	"github.com/gulugulu3399/bifrost/internal/nexus/service"

	// Infrastructure
	"github.com/gulugulu3399/bifrost/internal/pkg/config"
	"github.com/gulugulu3399/bifrost/internal/pkg/database"
	"github.com/gulugulu3399/bifrost/internal/pkg/id"
	"github.com/gulugulu3399/bifrost/internal/pkg/messenger"
	pkggrpc "github.com/gulugulu3399/bifrost/internal/pkg/network/grpc"
	pkgmw "github.com/gulugulu3399/bifrost/internal/pkg/middleware"
	"github.com/gulugulu3399/bifrost/internal/pkg/observability/logger"
	"github.com/gulugulu3399/bifrost/internal/pkg/observability/tracing"
	"github.com/gulugulu3399/bifrost/internal/pkg/security"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
)

func main() {
	// 1. 配置加载
	cfgPath := flag.String("config", "configs/nexus.yaml", "path to nexus config yaml")
	flag.Parse()

	cfg, err := config.LoadNexus(*cfgPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 2. 初始化全局 Logger
	zlog := logger.NewZap(cfg.LoggerConfig(), cfg.App.Name, cfg.App.Env)
	logger.SetGlobal(zlog)
	defer func() { _ = logger.Sync() }()
	// Metrics endpoint (Prometheus)
	go func() {
		mux := http.NewServeMux()
		mux.Handle("/metrics", promhttp.Handler())
		if err := http.ListenAndServe("localhost:9101", mux); err != nil {
			logger.Warn("metrics server stopped", logger.Err(err))
		}
	}()

	logger.Info("Starting Nexus Service",
		logger.String("grpc_addr", cfg.Server.GRPCAddr),
		logger.String("env", cfg.App.Env),
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

	// 3. 初始化基础设施
	// 3.1 数据库连接（统一走 internal/pkg/database）
	db, err := database.New(&database.Config{
		Driver:          cfg.Data.Database.Driver,
		DSN:             cfg.Data.Database.DSN,
		MaxOpenConns:    cfg.Data.Database.MaxOpenConns,
		MaxIdleConns:    cfg.Data.Database.MaxIdleConns,
		ConnMaxLifetime: cfg.Data.Database.ConnMaxLifetime,
		ConnMaxIdleTime: cfg.Data.Database.ConnMaxIdleTime,
	})
	if err != nil {
		logger.Fatal("Failed to connect to database", logger.Any("error", err))
	}
	defer func() {
		if err := db.Close(); err != nil {
			logger.Error("Failed to close database connection", logger.Any("error", err))
		}
	}()

	// 3.2 数据层上下文封装
	dataData := data.NewData(db, nil)

	// 3.3 事务管理器
	txManager := data.NewTransaction(db)

	// 3.4 Snowflake ID 生成器
	snowflake, err := id.NewSnowflakeGenerator(cfg.SnowflakeNode)
	if err != nil {
		logger.Fatal("Failed to init snowflake", logger.Any("error", err))
	}

	// 3.5 JWT 管理器
	jwtConfig := &security.JWTConfig{
		SecretKey:     cfg.Security.JWTSecret,
		Expiration:    cfg.Security.JWTExpiration,
		RefreshExp:    7 * 24 * time.Hour,
		Issuer:        "bifrost-nexus",
		SigningMethod: "HS256",
	}
	if jwtConfig.Expiration == 0 {
		jwtConfig.Expiration = 24 * time.Hour
	}
	jwtManager, err := security.NewJWTManager(jwtConfig)
	if err != nil {
		logger.Fatal("Failed to init JWT manager", logger.Any("error", err))
	}

	// 3.6 NATS messenger
	msgr, err := messenger.New(cfg.Messenger.Addr)
	if err != nil {
		logger.Fatal("Failed to connect to NATS", logger.Any("error", err))
	}
	defer func() { _ = msgr.Close() }()

	// 4. 初始化 Data 层 (Repositories)
	userRepo := data.NewUserRepo(dataData, snowflake)
	postRepo := data.NewPostRepo(dataData, snowflake)
	tagRepo := data.NewTagRepo(dataData, snowflake)
	catRepo := data.NewCategoryRepo(dataData, snowflake)
	commentRepo := data.NewCommentRepo(dataData, snowflake)

	// 5. 初始化 Biz 层 (UseCases)
	userUC := biz.NewUserUseCase(userRepo, txManager)

	// [新增] Forge gRPC 客户端（可选）
	var forgeClient forgev1.RenderServiceClient
	if cfg.RPC.ForgeAddr != "" {
		conn, err := pkggrpc.NewClient(pkggrpc.ClientConfig{Addr: cfg.RPC.ForgeAddr}, zlog)
		if err != nil {
			logger.Warn("Failed to connect forge, rendering disabled", logger.Err(err))
		} else {
			forgeClient = forgev1.NewRenderServiceClient(conn)
		}
	}

	postUC := biz.NewPostUseCase(postRepo, txManager, forgeClient)
	tagUC := biz.NewTagUseCase(tagRepo, txManager)
	catUC := biz.NewCategoryUseCase(catRepo, txManager)
	commentUC := biz.NewCommentUseCase(commentRepo, postRepo, txManager)

	// 6. 初始化 Service 层
	app := service.NewApp(jwtManager, msgr, userUC, postUC, tagUC, catUC, commentUC)

	// 7. 统一使用 pkggrpc.Server（拦截器链/反射/健康检查/优雅停止统一封装）
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

	g, err := pkggrpc.NewServer(pkggrpc.ServerConfig{
		Addr:             cfg.Server.GRPCAddr,
		Timeout:          10 * time.Second,
		EnableReflection: true,
		EnableHealth:     true,
		//TODO keepalive/mTLS 后续可以从 config 补齐
	}, zlog, jwtManager, publicMethods, adminMethods, grpc.ChainUnaryInterceptor(pkgmw.MetricsUnaryServerInterceptor(cfg.App.Name)))
	if err != nil {
		logger.Fatal("Failed to init gRPC server", logger.Any("error", err))
	}

	app.RegisterGRPC(g.GRPC())

	// [新增] 注册 StorageService（MinIO 预签名上传）
	storageSvc, err := service.NewStorageService(cfg)
	if err != nil {
		logger.Fatal("Failed to init storage service", logger.Any("error", err))
	}
	nexusv1.RegisterStorageServiceServer(g.GRPC(), storageSvc)

	logger.Info("Nexus gRPC server running", logger.String("addr", cfg.Server.GRPCAddr))
	if err := g.Start(); err != nil {
		logger.Fatal("failed to serve", logger.Any("error", err))
	}
}

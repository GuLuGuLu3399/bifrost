package grpc

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net"
	"os"
	"time"

	"github.com/gulugulu3399/bifrost/internal/pkg/observability/logger"
	"github.com/gulugulu3399/bifrost/internal/pkg/security"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/reflection"
)

// ServerConfig gRPC 服务端配置
type ServerConfig struct {
	Addr       string        // 监听地址，如 ":50051"
	CACertPath string        // CA 根证书路径
	CertPath   string        // 服务端证书路径
	KeyPath    string        // 服务端私钥路径
	Timeout    time.Duration // 优雅关闭超时时间

	// Keepalive 配置
	MaxConnectionIdle     time.Duration // 空闲连接最大存活时间
	MaxConnectionAge      time.Duration // 连接最大存活时间
	MaxConnectionAgeGrace time.Duration // 超过最大存活时间后的宽限期
	KeepaliveTime         time.Duration // Keepalive ping 间隔
	KeepaliveTimeout      time.Duration // Keepalive 超时时间

	// 是否启用特性
	EnableReflection bool // 启用 gRPC 反射 (用于 grpcurl 调试)
	EnableHealth     bool // 启用健康检查服务
}

// Server 封装结构体
type Server struct {
	cfg      ServerConfig
	server   *grpc.Server
	l        logger.Logger // 使用抽象接口
	health   *health.Server
	listener net.Listener
}

// NewServer 创建 gRPC 服务端
// [修改] 增加了 l logger.Logger 参数
func NewServer(
	cfg ServerConfig,
	l logger.Logger, // 依赖注入：传入 Logger 接口
	jwtManager *security.JWTManager,
	publicMethods map[string]struct{},
	adminMethods map[string]struct{},
	opts ...grpc.ServerOption,
) (*Server, error) {
	// 使用门面接口打印日志
	l.Debug("initializing gRPC server", logger.String("addr", cfg.Addr))

	var serverOpts []grpc.ServerOption

	// 1. 加载 mTLS (保持不变)
	if cfg.CertPath != "" && cfg.KeyPath != "" {
		creds, err := loadServerMTLSCredentials(cfg)
		if err != nil {
			return nil, fmt.Errorf("failed to load mTLS creds: %w", err)
		}
		serverOpts = append(serverOpts, grpc.Creds(creds))
	}

	// 2. Keepalive (保持不变)
	serverOpts = append(serverOpts,
		grpc.KeepaliveParams(keepalive.ServerParameters{
			MaxConnectionIdle:     cfg.MaxConnectionIdle,
			MaxConnectionAge:      cfg.MaxConnectionAge,
			MaxConnectionAgeGrace: cfg.MaxConnectionAgeGrace,
			Time:                  cfg.KeepaliveTime,
			Timeout:               cfg.KeepaliveTimeout,
		}),
		grpc.KeepaliveEnforcementPolicy(keepalive.EnforcementPolicy{
			MinTime:             5 * time.Second,
			PermitWithoutStream: true,
		}),
	)

	// 3. 组装拦截器链
	// 顺序：Recovery -> Context -> Logging -> Auth -> Error -> Handler
	// 说明:
	// - Recovery: 最外层，捕获 panic (必须第一个)
	// - Context: 从 Metadata 还原 UserID/Token 到 Go Context
	// - Logging: 记录请求日志 (需要 Context 中的字段)
	// - Auth: 鉴权 (需要在业务逻辑之前)
	// - Error: 错误转换 (最内层，最后执行)
	// 注意：Tracing 通过 StatsHandler 集成，不需要单独的拦截器
	serverOpts = append(serverOpts,
		grpc.ChainUnaryInterceptor(
			RecoveryInterceptor(l),                                   // 1. Panic 恢复
			ServerContextInterceptor(),                               // 2. Metadata -> Context (Phase 1 新增)
			LoggingInterceptor(l),                                    // 3. 访问日志
			AuthInterceptor(jwtManager, publicMethods, adminMethods), // 4. 鉴权
			ErrorInterceptor(),                                       // 5. 错误转换
		),
		// StatsHandler: OpenTelemetry Tracing 集成 (自动生成 Server Span)
		grpc.StatsHandler(otelgrpc.NewServerHandler()),
	)

	// 4. 追加用户自定义选项
	serverOpts = append(serverOpts, opts...)

	// 5. 创建实例
	grpcServer := grpc.NewServer(serverOpts...)

	s := &Server{
		cfg:    cfg,
		server: grpcServer,
		l:      l,
	}

	// 6. 注册健康检查
	if cfg.EnableHealth {
		s.health = health.NewServer()
		grpc_health_v1.RegisterHealthServer(grpcServer, s.health)
		s.health.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)
	}

	// 7. 注册反射
	if cfg.EnableReflection {
		reflection.Register(grpcServer)
	}

	l.Info("gRPC server initialized", logger.String("addr", cfg.Addr))
	return s, nil
}

// GRPC 返回底层 *grpc.Server，供业务层注册 service。
func (s *Server) GRPC() *grpc.Server {
	return s.server
}

// Start 启动 gRPC 服务 (阻塞)
func (s *Server) Start() error {
	listener, err := net.Listen("tcp", s.cfg.Addr)
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}
	s.listener = listener

	logger.Info("gRPC server starting", logger.String("addr", s.cfg.Addr))

	if err := s.server.Serve(listener); err != nil {
		return fmt.Errorf("gRPC server error: %w", err)
	}
	return nil
}

// StartAsync 异步启动服务，返回错误 channel
func (s *Server) StartAsync() <-chan error {
	errCh := make(chan error, 1)
	go func() {
		if err := s.Start(); err != nil {
			errCh <- err
		}
		close(errCh)
	}()
	return errCh
}

// Stop 优雅停止服务
func (s *Server) Stop(ctx context.Context) error {
	logger.Info("gRPC server shutting down gracefully")

	// 设置健康状态为 NOT_SERVING
	if s.health != nil {
		s.health.SetServingStatus("", grpc_health_v1.HealthCheckResponse_NOT_SERVING)
	}

	// 创建超时 contextx
	stopCtx, cancel := context.WithTimeout(ctx, s.cfg.Timeout)
	defer cancel()

	done := make(chan struct{})
	go func() {
		s.server.GracefulStop()
		close(done)
	}()

	select {
	case <-done:
		logger.Info("gRPC server stopped gracefully")
		return nil
	case <-stopCtx.Done():
		logger.Warn("gRPC server graceful stop timeout, forcing stop")
		s.server.Stop()
		return stopCtx.Err()
	}
}

// Close 实现 io.Closer 接口，用于 lifecycle.Shutdown
func (s *Server) Close() error {
	return s.Stop(context.Background())
}

// loadServerMTLSCredentials 加载服务端 mTLS 证书
func loadServerMTLSCredentials(cfg ServerConfig) (credentials.TransportCredentials, error) {
	// 加载服务端证书
	cert, err := tls.LoadX509KeyPair(cfg.CertPath, cfg.KeyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load server cert/key: %w", err)
	}

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		ClientAuth:   tls.NoClientCert, // 默认不要求客户端证书
	}

	// 如果配置了 CA 证书，则启用双向 mTLS
	if cfg.CACertPath != "" {
		caCert, err := os.ReadFile(cfg.CACertPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read CA cert: %w", err)
		}

		caCertPool := x509.NewCertPool()
		if !caCertPool.AppendCertsFromPEM(caCert) {
			return nil, fmt.Errorf("failed to add CA cert to pool")
		}

		tlsConfig.ClientCAs = caCertPool
		tlsConfig.ClientAuth = tls.RequireAndVerifyClientCert
	}

	return credentials.NewTLS(tlsConfig), nil
}

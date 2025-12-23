package grpc

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
	"time"

	"github.com/gulugulu3399/bifrost/internal/pkg/observability/logger"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/keepalive"
)

type ClientConfig struct {
	Addr       string        `yaml:"addr" mapstructure:"addr"`
	CACertPath string        `yaml:"ca_cert_path" mapstructure:"ca_cert_path"` // CA 根证书路径
	CertPath   string        `yaml:"cert_path" mapstructure:"cert_path"`       // 客户端证书路径
	KeyPath    string        `yaml:"key_path" mapstructure:"key_path"`         // 客户端私钥路径
	ServerName string        `yaml:"server_name" mapstructure:"server_name"`   // 证书对应的服务器名称
	Timeout    time.Duration `yaml:"timeout" mapstructure:"timeout"`
}

// NewClient 创建 gRPC 客户端
// 增加了 l logger.Logger 参数，用于依赖注入
func NewClient(cfg ClientConfig, l logger.Logger, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	l.Debug("initializing gRPC client", logger.String("addr", cfg.Addr))

	var dialOpts []grpc.DialOption

	// 1. 加载证书 (mTLS 或 Insecure)
	if cfg.CertPath != "" && cfg.KeyPath != "" {
		creds, err := loadMTLSCredentials(cfg)
		if err != nil {
			return nil, fmt.Errorf("failed to load mTLS creds: %w", err)
		}
		dialOpts = append(dialOpts, grpc.WithTransportCredentials(creds))
	} else {
		// 如果没配证书，默认走不安全连接 (方便本地开发)
		dialOpts = append(dialOpts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	// 2. Keepalive 配置
	dialOpts = append(dialOpts,
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                20 * time.Second,
			Timeout:             5 * time.Second,
			PermitWithoutStream: true,
		}),
	)

	// 3. 组装拦截器链
	// 顺序：Context (注入Metadata) -> Logging (记录日志)
	// 注意：Tracing 通过 StatsHandler 集成，不需要单独的拦截器
	dialOpts = append(dialOpts,
		grpc.WithChainUnaryInterceptor(
			// A. 上下文适配 (Go Context -> gRPC Metadata)
			ClientContextInterceptor(),

			// B. 日志记录
			ClientLoggingInterceptor(l),
		),
		// StatsHandler: OpenTelemetry Tracing 集成 (自动生成 Client Span)
		grpc.WithStatsHandler(otelgrpc.NewClientHandler()),
	)

	// 4. 追加用户自定义选项
	dialOpts = append(dialOpts, opts...)

	// 5. 创建连接 (异步)
	conn, err := grpc.NewClient(cfg.Addr, dialOpts...)
	if err != nil {
		return nil, err
	}

	// 6. 强制健康检查 (Fail Fast)
	timeout := cfg.Timeout
	if timeout <= 0 {
		timeout = 5 * time.Second
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	if err := verifyConnection(ctx, conn, l); err != nil {
		_ = conn.Close() // 忽略关闭错误
		return nil, fmt.Errorf("grpc connection unhealthy: %w", err)
	}

	l.Info("gRPC client connected", logger.String("addr", cfg.Addr))
	return conn, nil
}

// verifyConnection 执行物理连接与健康检查
// [修改] 传入 l logger.Logger 用于打印调试日志
func verifyConnection(ctx context.Context, conn *grpc.ClientConn, l logger.Logger) error {
	// 触发物理连接
	conn.Connect()

	l.Debug("waiting for gRPC connection state change")

	// 等待连接就绪
	for {
		state := conn.GetState()
		if state == connectivity.Ready {
			break
		}
		if !conn.WaitForStateChange(ctx, state) {
			return ctx.Err() // 超时或取消
		}
	}

	// 执行 gRPC 标准健康检查
	l.Debug("performing health check")
	healthClient := grpc_health_v1.NewHealthClient(conn)

	// 这里调用 Check 时，上面的拦截器链已经生效了
	// 所以健康检查的请求也会被 ClientLoggingInterceptor 记录下来
	resp, err := healthClient.Check(ctx, &grpc_health_v1.HealthCheckRequest{})
	if err != nil {
		return fmt.Errorf("health check failed: %v", err)
	}
	if resp.Status != grpc_health_v1.HealthCheckResponse_SERVING {
		return fmt.Errorf("service is not serving: %s", resp.Status)
	}

	return nil
}

// loadMTLSCredentials 加载证书文件并构建 TransportCredentials
func loadMTLSCredentials(cfg ClientConfig) (credentials.TransportCredentials, error) {
	cert, err := tls.LoadX509KeyPair(cfg.CertPath, cfg.KeyPath)
	if err != nil {
		return nil, err
	}

	caCert, err := os.ReadFile(cfg.CACertPath)
	if err != nil {
		return nil, err
	}

	caCertPool := x509.NewCertPool()
	if !caCertPool.AppendCertsFromPEM(caCert) {
		return nil, fmt.Errorf("failed to add ca cert")
	}

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		RootCAs:      caCertPool,
		ServerName:   cfg.ServerName,
	}

	return credentials.NewTLS(tlsConfig), nil
}

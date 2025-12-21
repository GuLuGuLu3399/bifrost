package http

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gulugulu3399/bifrost/internal/pkg/observability/logger"
)

// ServerConfig HTTP 服务端配置
type ServerConfig struct {
	Addr         string        // 监听地址，如 ":8080"
	ReadTimeout  time.Duration // 读超时
	WriteTimeout time.Duration // 写超时
	IdleTimeout  time.Duration // 空闲超时
	Timeout      time.Duration // 优雅关闭超时时间

	// TLS 配置 (可选)
	CertPath string // 证书路径
	KeyPath  string // 私钥路径
}

// DefaultServerConfig 返回默认的服务端配置
func DefaultServerConfig() ServerConfig {
	return ServerConfig{
		Addr:         ":8080",
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
		Timeout:      30 * time.Second,
	}
}

// Server 封装 HTTP 服务端
type Server struct {
	cfg    ServerConfig
	server *http.Server
}

// NewServer 创建 HTTP 服务端
func NewServer(cfg ServerConfig, handler http.Handler) *Server {
	logger.Debug("initializing HTTP server", logger.String("addr", cfg.Addr))

	server := &http.Server{
		Addr:         cfg.Addr,
		Handler:      handler,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		IdleTimeout:  cfg.IdleTimeout,
	}

	// 配置 TLS
	if cfg.CertPath != "" && cfg.KeyPath != "" {
		server.TLSConfig = &tls.Config{
			MinVersion: tls.VersionTLS12,
		}
	}

	logger.Info("HTTP server initialized", logger.String("addr", cfg.Addr))

	return &Server{
		cfg:    cfg,
		server: server,
	}
}

// Start 启动 HTTP 服务 (阻塞)
func (s *Server) Start() error {
	logger.Info("HTTP server starting", logger.String("addr", s.cfg.Addr))

	var err error
	if s.cfg.CertPath != "" && s.cfg.KeyPath != "" {
		err = s.server.ListenAndServeTLS(s.cfg.CertPath, s.cfg.KeyPath)
	} else {
		err = s.server.ListenAndServe()
	}

	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("HTTP server error: %w", err)
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
	logger.Info("HTTP server shutting down gracefully")

	// 创建超时 contextx
	stopCtx, cancel := context.WithTimeout(ctx, s.cfg.Timeout)
	defer cancel()

	if err := s.server.Shutdown(stopCtx); err != nil {
		logger.Warn("HTTP server graceful shutdown failed", logger.Err(err))
		return err
	}

	logger.Info("HTTP server stopped gracefully")
	return nil
}

// Close 实现 io.Closer 接口，用于 lifecycle.Shutdown
func (s *Server) Close() error {
	return s.Stop(context.Background())
}

// HTTPServer 返回底层的 *http.Server
func (s *Server) HTTPServer() *http.Server {
	return s.server
}

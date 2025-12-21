// gjallar 网关启动示例
// 本文件展示如何集成 gRPC Gateway

package main

import (
	"context"
	"os"

	"github.com/gulugulu3399/bifrost/internal/gjallar/server"
	"github.com/gulugulu3399/bifrost/internal/pkg/config"
	"github.com/gulugulu3399/bifrost/internal/pkg/lifecycle"
	"github.com/gulugulu3399/bifrost/internal/pkg/observability/logger"
)

func main() {
	// 1. 生命周期管理
	sh := lifecycle.NewShutdown()
	ctx, stop := sh.NotifyContext(context.Background())
	defer stop()

	// 2. 加载配置
	cfg := &config.Config{
		// 网关 HTTP 服务配置
		Http: config.HttpConfig{
			Port: "8080",
		},
		// Nexus gRPC 服务连接信息
		NexusRPC: config.GRPCConfig{
			Host: "localhost",
			Port: "50052",
		},
		// Beacon gRPC 服务连接信息
		BeaconRPC: config.GRPCConfig{
			Host: "localhost",
			Port: "50051",
		},
		// 其他配置...
	}

	// 3. 初始化日志
	zlog := logger.NewZap(cfg.LoggerConfig(), "gjallar", "dev")
	logger.SetGlobal(zlog)
	defer func() { _ = logger.Sync() }()

	// 4. 创建网关服务
	// New() 会：
	//   - 创建 gRPC 连接到 beacon 和 nexus
	//   - 初始化 gRPC Gateway mux
	//   - 注册所有服务的 HTTP 处理器（通过生成的 RegisterXxxServiceHandlerFromEndpoint）
	//   - 应用中间件（CORS、Auth、日志等）
	gjallarSrv, err := server.New(cfg)
	if err != nil {
		logger.Fatal("Failed to create server", logger.Err(err))
	}

	// 5. 启动网关服务
	go func() {
		logger.Info("Starting Gjallar HTTP Gateway",
			logger.String("addr", "http://localhost:8080"),
		)
		if err := gjallarSrv.Run(); err != nil {
			logger.Fatal("Server error", logger.Err(err))
		}
	}()

	// 6. 等待关闭信号
	<-ctx.Done()

	// 7. 优雅关闭
	logger.Info("Shutting down...")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := gjallarSrv.Shutdown(ctx); err != nil {
		logger.Error("Shutdown error", logger.Err(err))
		os.Exit(1)
	}

	logger.Info("Server stopped")
}

/*
网关服务启动流程说明：

1. [配置加载] 读取 gjallar 的 HTTP 和 gRPC 连接配置
   
2. [网关初始化] server.New() 做以下事情：
   a. 创建到 Beacon gRPC 服务的连接（:50051）
   b. 创建到 Nexus gRPC 服务的连接（:50052）
   c. 创建 gRPC Gateway ServeMux
   d. 调用生成的 RegisterXxxServiceHandlerFromEndpoint() 注册处理器
   
3. [HTTP 处理流程] 当收到 HTTP 请求时：
   a. 网关中间件处理（CORS、Auth、RequestID 等）
   b. gRPC Gateway 根据 URL 路径匹配 HTTP 规则
   c. 转换 HTTP 请求为 gRPC 请求
   d. 调用相应的 gRPC 微服务
   e. 将 gRPC 响应转换为 JSON 返回
   
示例请求流程：
   HTTP GET /v1/posts/123
   ↓
   gRPC Gateway 匹配 rpc GetPost(GetPostRequest) returns (GetPostResponse) 
   ↓
   转换为 gRPC 调用 beacon.GetPost(ctx, &GetPostRequest{SlugOrId: "123"})
   ↓
   Beacon 服务处理，返回 GetPostResponse
   ↓
   转换为 JSON 返回给 HTTP 客户端
   {
     "id": 123,
     "title": "...",
     "html_body": "...",
     ...
   }

关键特性：
✅ 自动生成的 HTTP 处理器（无手写代码）
✅ 自动的 JSON ↔ Protobuf 转换
✅ URL 路径参数自动提取和映射
✅ 查询参数自动映射到 RPC 请求字段
✅ 支持 gRPC 流（streaming）
✅ 完整的中间件支持
*/

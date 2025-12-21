# Bifrost 内部公共包 (internal/pkg)

本目录包含 Bifrost 微服务架构的核心基础设施组件，提供统一的错误处理、日志、网络、安全等功能。

## 📦 包概览

| 包名                | 说明                 |
|-------------------|--------------------|
| `xerr`            | 统一业务错误处理           |
| `logger`          | 结构化日志系统            |
| `database`        | 数据库连接管理            |
| `cache`           | Redis 缓存客户端        |
| `network/grpc`    | gRPC 服务端/客户端       |
| `network/http`    | HTTP 服务端及中间件       |
| `network/breaker` | 熔断器                |
| `security`        | 安全工具 (JWT/密码/加密)   |
| `lifecycle`       | 生命周期管理             |
| `context`         | Context 扩展         |
| `id`              | ID 生成器 (Snowflake) |
| `observability`   | 可观测性 (日志/指标/链路追踪)  |

---

## 🚨 xerr - 错误处理

统一的业务错误处理，支持错误码、堆栈追踪和错误链。

### 错误码定义

```go
import "github.com/gulugulu3399/bifrost/internal/pkg/xerr"

// 预定义错误码
xerr.CodeOK                 // 0   - 成功
xerr.CodeBadRequest         // 400 - 请求参数错误
xerr.CodeUnauthorized       // 401 - 未认证
xerr.CodeForbidden          // 403 - 权限不足
xerr.CodeNotFound           // 404 - 资源不存在
xerr.CodeConflict           // 409 - 资源冲突
xerr.CodeValidation         // 422 - 验证错误
xerr.CodeInternal           // 500 - 服务器内部错误
xerr.CodeServiceUnavailable // 503 - 服务不可用
xerr.CodeTimeout            // 504 - 网关超时
```

### 创建错误

```go
// 创建业务错误
err := xerr.New(xerr.CodeNotFound, "文章不存在")

// 包装底层错误 (如数据库错误)
err := xerr.Wrap(dbErr, xerr.CodeInternal, "查询文章失败")

// 格式化错误消息
err := xerr.Wrapf(dbErr, xerr.CodeInternal, "查询文章 %s 失败", postID)

// 快捷构造器
err := xerr.BadRequest("参数 %s 不能为空", "title")
err := xerr.NotFound("文章 %s 不存在", postID)
err := xerr.Unauthorized("token 已过期")
err := xerr.Forbidden("无权访问该资源")
err := xerr.Internal(originalErr)
```

### 错误转换

```go
// 将任意 error 转换为 *CodeError
codeErr := xerr.FromError(err)

// 获取错误信息
code := codeErr.GetCode()     // 错误码
msg := codeErr.GetMsg()       // 错误消息
stack := codeErr.StackTrace() // 堆栈信息
cause := codeErr.Unwrap()     // 原始错误
```

---

## 📝 observability/logger - 日志系统

结构化日志系统，支持多种输出格式和级别。

### 基本使用

```go
import "github.com/gulugulu3399/bifrost/internal/pkg/observability/logger"

// 使用全局日志器
logger.Debug("调试信息", logger.String("key", "value"))
logger.Info("操作完成", logger.Int("count", 100))
logger.Warn("警告信息", logger.Duration("elapsed", time.Second))
logger.Error("操作失败", logger.Err(err))
```

### 日志字段

```go
// 基本类型
logger.String("name", "张三")
logger.Int("age", 25)
logger.Int64("id", 1234567890)
logger.Float64("score", 95.5)
logger.Bool("active", true)
logger.Duration("elapsed", 100*time.Millisecond)
logger.Time("created_at", time.Now())

// 错误字段
logger.Err(err)

// 任意类型
logger.Any("data", map[string]interface{}{"foo": "bar"})
```

### 初始化 Zap 日志器

```go
import "github.com/gulugulu3399/bifrost/internal/pkg/observability/logger"

cfg := &logger.Config{
    Level:      logger.InfoLevel,
    Format:     "json",    // json 或 console
    Output:     "stdout",  // stdout, stderr, file
    CallerSkip: 1,
}

zapLogger, err := logger.NewZap(cfg)
if err != nil {
    panic(err)
}
logger.SetGlobal(zapLogger)
defer logger.Sync()
```

---

## 🌐 network/grpc - gRPC 服务

### 启动 gRPC 服务端

```go
import (
    "github.com/gulugulu3399/bifrost/internal/pkg/network/grpc"
    "go.uber.org/zap"
)

// 配置
cfg := grpc.DefaultServerConfig()
cfg.Addr = ":50051"
cfg.EnableHealth = true
cfg.EnableReflection = true

// 创建服务端
server, err := grpc.NewServer(cfg, zapLogger)
if err != nil {
    panic(err)
}

// 注册服务
pb.RegisterPostServiceServer(server.GRPCServer(), &postService{})

// 异步启动
errCh := server.StartAsync()

// 注册到生命周期管理
shutdown.Register(server)

// 监听启动错误
select {
case err := <-errCh:
    if err != nil {
        logger.Fatal("gRPC server failed", logger.Err(err))
    }
case <-ctx.Done():
}
```

### mTLS 配置

```go
cfg := grpc.ServerConfig{
    Addr:       ":50051",
    CACertPath: "/certs/ca.crt",
    CertPath:   "/certs/server.crt",
    KeyPath:    "/certs/server.key",
}
```

### 创建 gRPC 客户端

```go
cfg := grpc.ClientConfig{
    Addr:       "localhost:50051",
    CACertPath: "/certs/ca.crt",
    CertPath:   "/certs/client.crt",
    KeyPath:    "/certs/client.key",
    ServerName: "bifrost-service",
    Timeout:    5 * time.Second,
}

conn, err := grpc.NewClient(cfg,
    grpc.WithUnaryInterceptor(grpc.UnaryClientInterceptor(zapLogger)),
)
```

---

## 🌍 network/http - HTTP 服务

### 启动 HTTP 服务端

```go
import (
    "net/http"
    xhttp "github.com/gulugulu3399/bifrost/internal/pkg/network/http"
)

// 创建路由
mux := http.NewServeMux()
mux.HandleFunc("/v1/posts", handlePosts)

// 应用中间件
handler := xhttp.Chain(
    xhttp.Recovery(),
    xhttp.Logger(),
    xhttp.RequestID(),
    xhttp.Timeout(30 * time.Second),
)(mux)

// 配置服务端
cfg := xhttp.DefaultServerConfig()
cfg.Addr = ":8080"

// 创建并启动
server := xhttp.NewServer(cfg, handler)
errCh := server.StartAsync()

// 注册到生命周期管理
shutdown.Register(server)
```

### HTTP 中间件

```go
// 恢复 panic
xhttp.Recovery()

// 请求日志
xhttp.Logger()

// 注入 Request ID
xhttp.RequestID()

// 请求超时
xhttp.Timeout(30 * time.Second)

// CORS 跨域
xhttp.CORS(
    []string{"https://example.com"},  // 允许的来源
    []string{"GET", "POST", "PUT"},   // 允许的方法
    []string{"Authorization"},        // 允许的头
)

// 限流
xhttp.RateLimit(100) // 100 req/s

// 串联中间件
handler := xhttp.Chain(
    xhttp.Recovery(),
    xhttp.Logger(),
    xhttp.RequestID(),
)(yourHandler)
```

### HTTP 响应 (符合 Bifrost 规范)

```go
import xhttp "github.com/gulugulu3399/bifrost/internal/pkg/network/http"

// 成功响应 - 直接返回业务数据，无信封包装
// HTTP 200 OK
// {"id": "12345", "title": "文章标题", ...}
xhttp.OK(w, post)

// 创建成功
// HTTP 201 Created
xhttp.Created(w, newPost)

// 无内容
// HTTP 204 No Content
xhttp.NoContent(w)

// 错误响应 - 遵循 Google API 错误规范
// HTTP 400 Bad Request
// {"code": 3, "message": "参数错误", "details": [...]}
xhttp.BadRequest(w, "参数 title 不能为空")

// HTTP 404 Not Found
xhttp.NotFound(w, "文章不存在")

// HTTP 409 Conflict (带详情)
xhttp.Conflict(w, "文章标识已存在",
    xhttp.NewErrorDetail("SLUG_CONFLICT", "bifrost.content.v1", 
        map[string]string{"slug": "existing-slug"}),
)

// 从 xerr 自动转换
xhttp.Error(w, xerr.NotFound("文章 %s 不存在", postID))

// 设置响应头
xhttp.SetTraceIDHeader(w, traceID)
xhttp.SetRateLimitHeaders(w, 1000, 999)
```

### Context 工具

```go
// 存取 Request ID
ctx = xhttp.SetRequestID(ctx, "req_abc123")
requestID := xhttp.GetRequestID(ctx)

// 存取 User ID
ctx = xhttp.SetUserID(ctx, "user_123")
userID := xhttp.GetUserID(ctx)

// 存取 Trace ID
ctx = xhttp.SetTraceID(ctx, "trace_xyz")
traceID := xhttp.GetTraceID(ctx)
```

---

## 🗄️ database - 数据库

### 连接数据库

```go
import "github.com/gulugulu3399/bifrost/internal/pkg/database"

cfg := database.Config{
    Driver:          "postgres",
    DSN:             "postgres://user:pass@localhost:5432/bifrost?sslmode=disable",
    MaxOpenConns:    25,
    MaxIdleConns:    10,
    ConnMaxLifetime: 30 * time.Minute,
}

db, err := database.New(cfg)
if err != nil {
    panic(err)
}

// 注册到生命周期管理
shutdown.Register(db)
```

---

## 🔴 cache - Redis 缓存

### 连接 Redis

```go
import "github.com/gulugulu3399/bifrost/internal/pkg/cache"

cfg := cache.Config{
    Addr:         "localhost:6379",
    Password:     "",
    DB:           0,
    PoolSize:     10,
    MinIdleConns: 5,
    DialTimeout:  5 * time.Second,
}

client, err := cache.NewRedis(cfg)
if err != nil {
    panic(err)
}

// 基本操作
ctx := context.Background()
client.Set(ctx, "key", "value", time.Hour)
value, err := client.Get(ctx, "key")

// 注册到生命周期管理
shutdown.Register(client)
```

---

## 🔐 security - 安全工具

### JWT 令牌

```go
import "github.com/gulugulu3399/bifrost/internal/pkg/security"

// 生成 Token
token, err := security.GenerateJWT(userID, secret, 24*time.Hour)

// 验证 Token
claims, err := security.ParseJWT(token, secret)
userID := claims.Subject
```

### 密码哈希

```go
// 哈希密码
hash, err := security.HashPassword("password123")

// 验证密码
ok := security.CheckPassword("password123", hash)
```

### 数据加密

```go
// AES 加密
ciphertext, err := security.Encrypt(plaintext, key)

// AES 解密
plaintext, err := security.Decrypt(ciphertext, key)
```

---

## ♻️ lifecycle - 生命周期管理

### 优雅关闭

```go
import "github.com/gulugulu3399/bifrost/internal/pkg/lifecycle"

// 创建 Shutdown 管理器
sh := lifecycle.NewShutdown()

// 创建可取消的 contextx
ctx, stop := sh.NotifyContext(lifecycle.Background())
defer stop()

// 注册需要关闭的资源 (实现 io.Closer 接口)
sh.Register(db)
sh.Register(redis)
sh.Register(grpcServer)
sh.Register(httpServer)

// 等待信号
<-ctx.Done()

// 关闭所有资源 (按注册顺序的逆序)
if err := sh.CloseAll(); err != nil {
    logger.Error("shutdown error", logger.Err(err))
}
```

---

## ❄️ id - ID 生成器

### Snowflake ID

```go
import "github.com/gulugulu3399/bifrost/internal/pkg/id"

// 初始化 (每个节点唯一 nodeID)
generator, err := id.NewSnowflake(1)
if err != nil {
    panic(err)
}

// 生成 ID
id := generator.Generate() // int64
```

---

## 📊 observability - 可观测性

### Metrics 指标

```go
import "github.com/gulugulu3399/bifrost/internal/pkg/observability/metrics"

// 初始化 Prometheus 指标
metrics.Init("bifrost")

// 记录 HTTP 请求
metrics.HTTPRequestTotal.WithLabelValues("GET", "/v1/posts", "200").Inc()
metrics.HTTPRequestDuration.WithLabelValues("GET", "/v1/posts").Observe(0.05)
```

### Tracing 链路追踪

```go
import "github.com/gulugulu3399/bifrost/internal/pkg/observability/tracing"

// 初始化 Jaeger
tp, err := tracing.Init("bifrost-service", "http://jaeger:14268/api/traces")
if err != nil {
    panic(err)
}
defer tp.Shutdown(ctx)
```

---

## 🔧 network/breaker - 熔断器

```go
import "github.com/gulugulu3399/bifrost/internal/pkg/network/breaker"

// 创建熔断器
cb := breaker.New(breaker.Config{
    MaxRequests: 5,           // 半开状态最大请求数
    Interval:    10 * time.Second,
    Timeout:     30 * time.Second,
})

// 使用熔断器
result, err := cb.Execute(func() (interface{}, error) {
    return callExternalService()
})
```

---

## 🔗 错误码映射表

| HTTP 状态码 | gRPC 错误码              | 业务错误码    | 场景      |
|----------|-----------------------|----------|---------|
| 200      | OK (0)                | 0        | 成功      |
| 400      | InvalidArgument (3)   | 400, 422 | 参数验证失败  |
| 401      | Unauthenticated (16)  | 401      | 身份认证失败  |
| 403      | PermissionDenied (7)  | 403      | 权限不足    |
| 404      | NotFound (5)          | 404      | 资源不存在   |
| 409      | AlreadyExists (6)     | 409      | 资源冲突    |
| 429      | ResourceExhausted (8) | -        | 请求频率限制  |
| 500      | Internal (13)         | 500      | 服务器内部错误 |
| 503      | Unavailable (14)      | 503      | 服务不可用   |
| 504      | DeadlineExceeded (4)  | 504      | 请求超时    |

---

## 📋 完整示例

```go
package main

import (
    "context"
    "net/http"
    "os"

    "github.com/gulugulu3399/bifrost/internal/pkg/cache"
    "github.com/gulugulu3399/bifrost/internal/pkg/database"
    "github.com/gulugulu3399/bifrost/internal/pkg/lifecycle"
    "github.com/gulugulu3399/bifrost/internal/pkg/network/grpc"
    xhttp "github.com/gulugulu3399/bifrost/internal/pkg/network/http"
    "github.com/gulugulu3399/bifrost/internal/pkg/observability/logger"
)

func main() {
    // 1. 初始化日志
    zapLogger, _ := logger.NewZap(logger.DefaultConfig())
    logger.SetGlobal(zapLogger)
    defer logger.Sync()

    // 2. 创建生命周期管理器
    sh := lifecycle.NewShutdown()
    ctx, stop := sh.NotifyContext(lifecycle.Background())
    defer stop()

    // 3. 初始化数据库
    db, _ := database.New(database.Config{
        Driver: "postgres",
        DSN:    os.Getenv("DATABASE_URL"),
    })
    sh.Register(db)

    // 4. 初始化 Redis
    redis, _ := cache.NewRedis(cache.Config{
        Addr: os.Getenv("REDIS_URL"),
    })
    sh.Register(redis)

    // 5. 启动 gRPC 服务
    grpcCfg := grpc.DefaultServerConfig()
    grpcCfg.Addr = ":50051"
    grpcServer, _ := grpc.NewServer(grpcCfg, zapLogger.Desugar())
    // 注册 gRPC 服务...
    grpcServer.StartAsync()
    sh.Register(grpcServer)

    // 6. 启动 HTTP 服务
    mux := http.NewServeMux()
    // 注册路由...
    
    handler := xhttp.Chain(
        xhttp.Recovery(),
        xhttp.Logger(),
        xhttp.RequestID(),
    )(mux)

    httpCfg := xhttp.DefaultServerConfig()
    httpCfg.Addr = ":8080"
    httpServer := xhttp.NewServer(httpCfg, handler)
    httpServer.StartAsync()
    sh.Register(httpServer)

    logger.Info("服务启动完成",
        logger.String("grpc", grpcCfg.Addr),
        logger.String("http", httpCfg.Addr),
    )

    // 7. 等待退出信号
    <-ctx.Done()
    
    // 8. 优雅关闭
    if err := sh.CloseAll(); err != nil {
        logger.Error("关闭失败", logger.Err(err))
    }
    
    logger.Info("服务已停止")
}
```

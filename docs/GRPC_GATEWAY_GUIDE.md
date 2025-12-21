# gRPC Gateway 集成指南

## 📋 概览

本项目已集成 **gRPC Gateway** 来自动生成 HTTP/JSON API。系统架构如下：

```
HTTP Client
    ↓
[gjallar - HTTP Gateway] ←─ 自动生成的 gRPC Gateway 代码
    ↓                  ↓
Beacon Service    Nexus Service
  (gRPC)           (gRPC)
```

## 🛠️ 快速开始

### 1. 生成代码

首先确保安装了必要的工具：

```bash
# 安装 buf（推荐）
go install github.com/bufbuild/buf/cmd/buf@latest

# 或使用 protoc（需要 protoc-gen-go, protoc-gen-go-grpc, protoc-gen-grpc-gateway）
brew install protoc  # macOS
# 或 Windows: choco install protoc

# 安装必要的插件
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@latest
```

### 2. 生成 protobuf + gRPC + gRPC Gateway 代码

**使用 buf（推荐）：**

```bash
make proto-generate
# 或直接
buf generate
```

**或使用 protoc：**

```bash
protoc -I api \
  --go_out=go_services/api \
  --go-grpc_out=go_services/api \
  --grpc-gateway_out=go_services/api \
  --grpc-gateway_opt=paths=source_relative \
  api/content/v1/beacon/beacon.proto \
  api/content/v1/nexus/nexus.proto
```

### 3. 验证生成的代码

生成成功后会看到：

- `*.pb.go` - protobuf 数据结构
- `*_grpc.pb.go` - gRPC 服务定义和客户端
- `*.pb.gw.go` - **gRPC Gateway 处理器**（新增）

例如：

```
api/content/v1/beacon/
  ├── beacon.pb.go
  ├── beacon_grpc.pb.go
  └── beacon.pb.gw.go       ← Gateway 处理器（自动生成）

api/content/v1/nexus/
  ├── nexus.pb.go
  ├── nexus_grpc.pb.go
  └── nexus.pb.gw.go        ← Gateway 处理器（自动生成）
```

## 📝 使用 gRPC Gateway

### 在 Proto 中定义 HTTP 规则

在 `.proto` 文件的 RPC 方法中添加 `google.api.http` 选项：

```protobuf
syntax = "proto3";

package bifrost.content.v1.beacon;

import "google/api/annotations.proto";

service BeaconService {
  // GET /v1/posts/{slug_or_id}
  rpc GetPost(GetPostRequest) returns (GetPostResponse) {
    option (google.api.http) = {
      get: "/v1/posts/{slug_or_id}"
    };
  }

  // GET /v1/posts
  rpc ListPosts(ListPostsRequest) returns (ListPostsResponse) {
    option (google.api.http) = {
      get: "/v1/posts"
    };
  }
}
```

### gRPC Gateway 中注册服务

在 [go_services/internal/gjallar/router/router_grpc_gateway.go](../../go_services/internal/gjallar/router/router_grpc_gateway.go) 中：

```go
import (
    "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
)

func New(ctx context.Context, nexusConn, beaconConn *grpc.ClientConn) (http.Handler, error) {
    mux := runtime.NewServeMux(
        runtime.WithMarshalerOption(runtime.MIMETypeJSON, &runtime.JSONPb{
            UseProtoNames:   true,
            EmitUnpopulated: true,
        }),
    )

    opts := []grpc.DialOption{grpc.WithInsecure()}

    // 注册 Beacon 服务
    if err := beaconv1.RegisterBeaconServiceHandlerFromEndpoint(
        ctx, mux, beaconConn.Target(), opts,
    ); err != nil {
        return nil, err
    }

    // 注册 Nexus 服务
    if err := nexusv1.RegisterPostServiceHandlerFromEndpoint(
        ctx, mux, nexusConn.Target(), opts,
    ); err != nil {
        return nil, err
    }

    return mux, nil
}
```

## 🚀 HTTP 路由规则速查

| HTTP 方法 | 规则 | 说明 |
|---------|------|------|
| GET | `get: "/v1/items/{id}"` | 从 URL 路径提取参数 |
| POST | `post: "/v1/items"` + `body: "*"` | 整个请求体作为消息 |
| POST（特定字段） | `post: "/v1/items"` + `body: "data"` | 只有 `data` 字段来自请求体 |
| PUT | `put: "/v1/items/{id}"` + `body: "*"` | 更新操作 |
| PATCH | `patch: "/v1/items/{id}"` + `body: "*"` | 部分更新 |
| DELETE | `delete: "/v1/items/{id}"` | 无请求体 |
| GET (多参数) | `get: "/v1/items"` | 查询参数自动映射 |

## 💡 示例

### 例子 1：获取文章详情

**Proto 定义：**

```protobuf
rpc GetPost(GetPostRequest) returns (GetPostResponse) {
  option (google.api.http) = {
    get: "/v1/posts/{slug_or_id}"
  };
}

message GetPostRequest {
  string slug_or_id = 1;  // 从 URL 路径提取
}
```

**HTTP 请求：**

```bash
curl http://localhost:8080/v1/posts/my-blog-post
curl http://localhost:8080/v1/posts/123
```

### 例子 2：创建文章

**Proto 定义：**

```protobuf
rpc CreatePost(CreatePostRequest) returns (CreatePostResponse) {
  option (google.api.http) = {
    post: "/v1/posts"
    body: "*"
  };
}

message CreatePostRequest {
  string title = 1;
  string content = 2;
}
```

**HTTP 请求：**

```bash
curl -X POST http://localhost:8080/v1/posts \
  -H "Content-Type: application/json" \
  -d '{
    "title": "My Blog Post",
    "content": "Hello World"
  }'
```

### 例子 3：部分更新

**Proto 定义：**

```protobuf
rpc UpdatePost(UpdatePostRequest) returns (UpdatePostResponse) {
  option (google.api.http) = {
    patch: "/v1/posts/{id}"
    body: "*"
  };
}

message UpdatePostRequest {
  int64 id = 1;           // 从 URL 路径提取
  string title = 2;       // 从请求体提取
  string content = 3;
}
```

**HTTP 请求：**

```bash
curl -X PATCH http://localhost:8080/v1/posts/123 \
  -H "Content-Type: application/json" \
  -d '{"title": "Updated Title"}'
```

### 例子 4：列表查询

**Proto 定义：**

```protobuf
rpc ListPosts(ListPostsRequest) returns (ListPostsResponse) {
  option (google.api.http) = {
    get: "/v1/posts"
  };
}

message ListPostsRequest {
  int32 page_size = 1;    // → ?page_size=10
  string page_token = 2;  // → ?page_token=abc
  int64 category_id = 3;  // → ?category_id=1
}
```

**HTTP 请求：**

```bash
curl "http://localhost:8080/v1/posts?page_size=10&category_id=1&page_token=abc"
```

## 🔧 高级配置

### 自定义 JSON 编码选项

```go
mux := runtime.NewServeMux(
    runtime.WithMarshalerOption(runtime.MIMETypeJSON, &runtime.JSONPb{
        UseProtoNames:    true,    // 使用 proto 字段名
        EmitUnpopulated:  true,    // 输出未设置的字段
        UseEnumNumbers:   false,   // 使用枚举名而非数字
        Indent:           "",      // 缩进（空为最小化）
    }),
)
```

### 自定义错误处理

```go
func customErrorHandler(ctx context.Context, mux *runtime.ServeMux, 
    marshaler runtime.Marshaler, w http.ResponseWriter, 
    r *http.Request, err error) {
    // 记录错误
    logger.Global().Warn("gateway error", logger.Err(err))
    
    // 自定义响应格式
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusInternalServerError)
    json.NewEncoder(w).Encode(map[string]interface{}{
        "code":    500,
        "message": err.Error(),
    })
}

mux := runtime.NewServeMux(
    runtime.WithErrorHandler(customErrorHandler),
)
```

### 请求/响应拦截

```go
mux := runtime.NewServeMux(
    runtime.WithIncomingHeaderMatcher(func(key string) (string, bool) {
        // 自定义哪些 HTTP 头传递给 gRPC
        return key, true
    }),
    runtime.WithOutgoingHeaderMatcher(func(key string) (string, bool) {
        // 自定义哪些 gRPC 响应头返回给 HTTP
        return key, true
    }),
)
```

## ✅ 最佳实践

1. **使用 buf 管理 proto**
   - 保持 `buf.yaml` 和 `buf.gen.yaml` 最新
   - 定期运行 `buf lint` 检查 proto 规范

2. **HTTP 规则设计**
   - 遵循 RESTful 规范
   - 使用标准的 HTTP 方法（GET, POST, PUT, PATCH, DELETE）
   - 路径参数用于唯一标识（`{id}`）
   - 查询参数用于过滤/分页（`?page_size=10`）

3. **错误处理**
   - 使用 `google.rpc.Status` 返回标准错误
   - 在网关层添加自定义错误处理
   - 记录所有 gateway 错误便于调试

4. **性能优化**
   - gRPC Gateway 会自动复用 gRPC 连接
   - 使用连接池避免频繁创建连接
   - 配置适当的超时和重试策略

5. **安全性**
   - 在网关层实现认证/授权中间件
   - 验证所有输入
   - 使用 TLS/SSL 加密通信

## 📚 相关资源

- [gRPC Gateway 官方文档](https://grpc-ecosystem.github.io/grpc-gateway/)
- [Google API 设计指南](https://cloud.google.com/apis/design)
- [Protobuf 文档](https://developers.google.com/protocol-buffers)
- [buf 官方文档](https://docs.buf.build/)

## 🐛 常见问题

### Q: 生成的 `.pb.gw.go` 文件在哪里？

A: 在对应的 proto 文件同目录下。例如 `beacon.proto` 生成 `beacon.pb.gw.go`。

### Q: 如何在网关中添加认证？

A: 在 `middleware/auth.go` 中实现，并在 `server/server.go` 中应用到网关 mux。

### Q: HTTP 和 gRPC 可以共存吗？

A: 可以，但需要在不同的端口。通常 gRPC 用 50051，HTTP Gateway 用 8080。

### Q: 如何支持 CORS？

A: 在 `server/server.go` 中已配置 CORS 中间件，允许所有跨域。

### Q: 可以自定义 HTTP 路由吗？

A: 可以，gRPC Gateway 生成的是基础处理器。你可以在 router 中添加自定义逻辑。

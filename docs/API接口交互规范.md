# Bifrost API 接口交互规范

## 1. 核心原则

### 协议标准
- **传输协议**: HTTP/2（优先）或 HTTP/1.1
- **数据格式**: JSON（UTF-8编码），无外层信封包装
- **状态表达**: 严格使用 HTTP Status Code 表达业务结果
- **定义源**: .proto 文件是唯一的接口定义源（Single Source of Truth）

## 2. 响应结构

### 2.1 成功响应
放弃信封结构，HTTP 状态码为 200 OK，Body 直接返回业务数据。

**示例：获取文章详情**
```http
GET /v1/posts/12345
HTTP/1.1 200 OK
Content-Type: application/json
X-Trace-Id: 10f3c3a9-7b32-4d32-8c3a-1234567890ab

{
  "id": "12345",
  "title": "Bifrost 架构解析",
  "slug": "bifrost-architecture",
  "status": "PUBLISHED",
  "author": {
    "id": "888",
    "username": "admin",
    "nickname": "系统管理员"
  },
  "category_id": "1001",
  "published_at": "2023-10-01T12:00:00Z",
  "view_count": 1500,
  "like_count": 89
}
```

### 2.2 错误响应
HTTP Status ≥ 400 时，Body 返回标准错误详情。

**示例：创建文章冲突**
```http
POST /v1/posts
HTTP/1.1 409 Conflict
Content-Type: application/json

{
  "code": 6,
  "message": "文章标识 'bifrost-arch' 已存在",
  "details": [
    {
      "@type": "type.googleapis.com/google.rpc.ErrorInfo",
      "reason": "SLUG_CONFLICT",
      "domain": "bifrost.content.v1",
      "metadata": {
        "slug": "bifrost-arch",
        "existing_id": "12345"
      }
    }
  ]
}
```

**常见错误码映射**
| HTTP 状态码 | gRPC 错误码 | 业务场景 |
|------------|-------------|----------|
| 400 | 3 (INVALID_ARGUMENT) | 请求参数验证失败 |
| 401 | 16 (UNAUTHENTICATED) | 身份认证失败 |
| 403 | 7 (PERMISSION_DENIED) | 权限不足 |
| 404 | 5 (NOT_FOUND) | 资源不存在 |
| 409 | 6 (ALREADY_EXISTS) | 资源冲突 |
| 429 | 8 (RESOURCE_EXHAUSTED) | 请求频率限制 |
| 500 | 13 (INTERNAL) | 服务器内部错误 |

## 3. 分页机制

采用 Cursor-based（游标）分页，支持海量数据和无限滚动。

**请求示例**
```http
GET /v1/posts?page_size=20&page_token=eyJpZCI6IjEwMDAifQ==
```

**响应结构**
```json
{
  "posts": [
    {
      "id": "1001",
      "title": "文章标题",
      "summary": "文章摘要",
      "published_at": "2023-10-01T12:00:00Z"
    }
  ],
  "next_page_token": "eyJpZCI6Ijk5OSJ9",
  "total_size": 1500
}
```

## 4. Header 定义

### 请求 Header
| Header | 必选 | 说明 | 示例 |
|--------|------|------|------|
| Authorization | 是 | Bearer Token 认证 | `Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...` |
| X-Request-Id | 否 | 客户端请求ID（幂等性控制） | `req_abc123def456` |
| Content-Type | 是 | 请求体类型 | `application/json` |

### 响应 Header
| Header | 说明 | 示例 |
|--------|------|------|
| X-Trace-Id | 全链路追踪ID | `10f3c3a9-7b32-4d32-8c3a-1234567890ab` |
| X-RateLimit-Limit | 请求频率限制 | `1000` |
| X-RateLimit-Remaining | 剩余请求次数 | `999` |

## 5. 业务接口概览

### 5.1 Nexus 服务（写操作）
| 方法 | 路径 | 说明 | 请求体 |
|------|------|------|--------|
| POST | `/v1/posts` | 创建文章草稿 | `PostCreateRequest` |
| PUT | `/v1/posts/{id}` | 更新文章内容 | `PostUpdateRequest` |
| DELETE | `/v1/posts/{id}` | 软删除文章 | - |
| POST | `/v1/posts/{id}:publish` | 发布文章 | `PublishRequest` |
| POST | `/v1/posts/{id}:unpublish` | 撤回发布 | - |
| POST | `/v1/upload/token` | 获取上传凭证 | `UploadTokenRequest` |

### 5.2 Beacon 服务（读操作）
| 方法 | 路径 | 说明 | 查询参数 |
|------|------|------|----------|
| GET | `/v1/posts` | 文章列表（分页） | `page_size, page_token, category, status` |
| GET | `/v1/posts/{id}` | 文章详情 | - |
| GET | `/v1/categories` | 分类列表 | - |
| GET | `/v1/tags` | 标签列表 | - |
| GET | `/v1/authors/{id}/posts` | 作者文章列表 | `page_size, page_token` |

### 5.3 Mirror 服务（搜索）
| 方法 | 路径 | 说明 | 查询参数 |
|------|------|------|----------|
| GET | `/v1/search` | 全文检索 | `q, page_size, page_token` |
| GET | `/v1/search/suggest` | 搜索建议 | `q, limit=5` |

### 5.4 Auth 服务（身份认证）
| 方法 | 路径 | 说明 | 请求体 |
|------|------|------|--------|
| POST | `/v1/auth/login` | 用户登录 | `LoginRequest` |
| POST | `/v1/auth/register` | 用户注册 | `RegisterRequest` |
| POST | `/v1/auth/refresh` | 刷新令牌 | `RefreshRequest` |
| POST | `/v1/auth/logout` | 用户登出 | - |

## 6. 特殊交互

### 6.1 Server-Sent Events（实时流）
对于耗时任务（AI摘要生成、批量导出），使用 SSE 推送进度。

**Endpoint**
```http
GET /v1/stream/events?token={jwt_token}
Accept: text/event-stream
```

**事件类型**
```javascript
// 前端接入示例
const eventSource = new EventSource('/v1/stream/events?token=' + jwtToken);

eventSource.addEventListener('operation_update', (event) => {
  const data = JSON.parse(event.data);
  console.log(`任务 ${data.operation_id} 进度: ${data.progress}%`);
});

eventSource.addEventListener('operation_complete', (event) => {
  const data = JSON.parse(event.data);
  console.log('任务完成:', data.result_url);
});
```

### 6.2 文件上传流程
```javascript
// 1. 获取上传凭证
const tokenResp = await fetch('/v1/upload/token', {
  method: 'POST',
  body: JSON.stringify({ file_type: 'image', file_name: 'avatar.jpg' })
});

// 2. 直传到 MinIO
const formData = new FormData();
formData.append('file', file);
const uploadResp = await fetch(tokenResp.upload_url, {
  method: 'PUT',
  body: formData
});

// 3. 确认上传完成
await fetch(`/v1/upload/${tokenResp.upload_id}/complete`, {
  method: 'POST'
});
```

## 7. 数据格式规范

### 7.1 时间格式
- 所有时间字段使用 ISO 8601 格式
- 时区：UTC（带 Z 后缀）
- 示例：`2023-10-01T12:00:00Z`

### 7.2 ID 格式
- 使用 Snowflake ID（int64 字符串表示）
- 示例：`"1234567890123456789"`

### 7.3 枚举值
- 使用大写蛇形命名法
- 示例：`POST_STATUS_DRAFT`, `VISIBILITY_PUBLIC`

## 8. 认证授权

### 8.1 JWT Token 结构
```json
{
  "header": {
    "alg": "HS256",
    "typ": "JWT"
  },
  "payload": {
    "sub": "888",
    "username": "admin",
    "roles": ["admin", "editor"],
    "exp": 1698729600,
    "iat": 1698726000,
    "jti": "token_abc123"
  }
}
```

### 8.2 权限控制
| 角色 | 权限 |
|------|------|
| admin | 所有操作 |
| editor | 文章管理、分类管理 |
| author | 个人文章管理 |
| reader | 只读权限 |

### 8.3 Token 刷新机制
- Access Token 有效期：2小时
- Refresh Token 有效期：7天
- 刷新接口：`POST /v1/auth/refresh`

## 9. 限流策略

### 9.1 限流维度
- **全局限流**：防止 DDoS
- **IP 限流**：防止恶意爬虫
- **用户限流**：防止单用户滥用
- **接口限流**：保护关键接口

### 9.2 限流配置
```yaml
rate_limits:
  global:
    requests_per_second: 10000
    burst: 20000
  per_ip:
    requests_per_second: 100
    burst: 200
  per_user:
    requests_per_second: 50
    burst: 100
  endpoints:
    CreatePost:
      requests_per_second: 10
      burst: 20
    UploadFile:
      requests_per_second: 5
      burst: 10
```

### 9.3 限流响应
```http
HTTP/1.1 429 Too Many Requests
Content-Type: application/json
X-RateLimit-Limit: 50
X-RateLimit-Remaining: 0
X-RateLimit-Reset: 1698729660

{
  "code": 8,
  "message": "请求频率超限，请稍后重试",
  "details": [
    {
      "@type": "type.googleapis.com/google.rpc.ErrorInfo",
      "reason": "RATE_LIMIT_EXCEEDED",
      "domain": "bifrost.api.v1",
      "metadata": {
        "retry_after": "60s"
      }
    }
  ]
}
```

## 10. 缓存策略

### 10.1 缓存层级
1. **CDN 缓存**：静态资源（图片、CSS、JS）
2. **网关缓存**：API 响应缓存（Gjallar）
3. **应用缓存**：热点数据（Beacon）
4. **数据库缓存**：查询结果缓存

### 10.2 缓存控制 Header
```http
# 强缓存
Cache-Control: public, max-age=3600
ETag: "abc123"

# 协商缓存
Cache-Control: no-cache
Last-Modified: Wed, 01 Oct 2023 12:00:00 GMT

# 禁止缓存
Cache-Control: no-store, must-revalidate
```

### 10.3 缓存失效策略
- **主动失效**：数据更新时主动清除
- **定时失效**：设置合理的 TTL
- **版本控制**：使用版本号标识

## 11. 错误处理

### 11.1 错误分类
| 类别 | HTTP 状态码 | 说明 |
|------|------------|------|
| 客户端错误 | 4xx | 请求参数错误、认证失败等 |
| 服务器错误 | 5xx | 内部错误、依赖服务失败等 |
| 业务错误 | 422 | 业务规则验证失败 |

### 11.2 错误响应格式
```json
{
  "code": 3,
  "message": "请求参数验证失败",
  "details": [
    {
      "@type": "type.googleapis.com/google.rpc.BadRequest",
      "field_violations": [
        {
          "field": "title",
          "description": "标题不能为空"
        },
        {
          "field": "slug",
          "description": "标识符格式不正确"
        }
      ]
    }
  ]
}
```

### 11.3 常见错误场景
- **参数验证失败**：400 Bad Request
- **认证失败**：401 Unauthorized
- **权限不足**：403 Forbidden
- **资源不存在**：404 Not Found
- **资源冲突**：409 Conflict
- **请求超时**：408 Request Timeout
- **服务不可用**：503 Service Unavailable

## 12. 版本控制

### 12.1 API 版本策略
- **URL 版本**：`/v1/posts`, `/v2/posts`
- **向后兼容**：新版本保持对旧版本的兼容
- **废弃通知**：提前 3 个月通知废弃计划

### 12.2 版本演进规则
- **主版本**：不兼容的 API 变更
- **次版本**：向后兼容的功能性新增
- **修订版本**：向后兼容的问题修正

### 12.3 版本响应 Header
```http
API-Version: v1
Supported-Versions: v1, v2
Deprecated-Versions: v0
```

## 13. 监控与日志

### 13.1 请求追踪
每个请求都会生成唯一的 Trace ID，贯穿整个调用链：

```http
X-Trace-Id: 10f3c3a9-7b32-4d32-8c3a-1234567890ab
X-Request-Id: req_abc123def456
```

### 13.2 日志格式
```json
{
  "timestamp": "2023-10-01T12:00:00Z",
  "level": "INFO",
  "trace_id": "10f3c3a9-7b32-4d32-8c3a-1234567890ab",
  "request_id": "req_abc123def456",
  "user_id": "888",
  "method": "GET",
  "path": "/v1/posts/12345",
  "status": 200,
  "duration_ms": 45,
  "message": "Request completed successfully"
}
```

### 13.3 性能指标
- **响应时间**：P50, P95, P99
- **吞吐量**：QPS, RPS
- **错误率**：按状态码分类
- **可用性**：SLA 监控

## 14. 安全最佳实践

### 14.1 输入验证
- **参数类型验证**：确保数据类型正确
- **长度限制**：防止缓冲区溢出
- **格式验证**：邮箱、URL 等格式检查
- **内容过滤**：XSS、SQL 注入防护

### 14.2 输出编码
- **HTML 编码**：防止 XSS 攻击
- **JSON 编码**：防止 JSON 注入
- **URL 编码**：防止 URL 注入

### 14.3 HTTPS 强制
- **全站 HTTPS**：所有 API 必须使用 HTTPS
- **HSTS 头**：强制浏览器使用 HTTPS
- **证书管理**：自动更新和监控

## 15. 开发工具

### 15.1 API 文档生成
- **Swagger UI**：交互式 API 文档
- **OpenAPI 规范**：标准化的 API 描述
- **代码生成**：自动生成客户端 SDK

### 15.2 测试工具
- **Postman Collection**：API 测试集合
- **自动化测试**：集成到 CI/CD 流程
- **性能测试**：压测和负载测试

### 15.3 调试工具
- **请求追踪**：分布式链路追踪
- **日志聚合**：集中式日志管理
- **错误监控**：实时错误告警

## 16. 示例代码

### 16.1 JavaScript 客户端
```javascript
class BifrostAPI {
  constructor(baseURL, token) {
    this.baseURL = baseURL;
    this.token = token;
  }

  async request(method, path, data = null) {
    const url = `${this.baseURL}${path}`;
    const options = {
      method,
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${this.token}`,
        'X-Request-Id': this.generateRequestId()
      }
    };

    if (data) {
      options.body = JSON.stringify(data);
    }

    const response = await fetch(url, options);
    
    if (!response.ok) {
      const error = await response.json();
      throw new APIError(error.code, error.message, error.details);
    }

    return response.json();
  }

  async getPost(id) {
    return this.request('GET', `/v1/posts/${id}`);
  }

  async createPost(postData) {
    return this.request('POST', '/v1/posts', postData);
  }

  generateRequestId() {
    return `req_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`;
  }
}

class APIError extends Error {
  constructor(code, message, details) {
    super(message);
    this.code = code;
    this.details = details;
  }
}

// 使用示例
const api = new BifrostAPI('https://api.bifrost.com', 'your-jwt-token');

try {
  const post = await api.getPost('12345');
  console.log('文章详情:', post);
} catch (error) {
  console.error('API 错误:', error.message);
}
```

### 16.2 Go 客户端
```go
package client

import (
    "context"
    "encoding/json"
    "fmt"
    "net/http"
    "time"
)

type Client struct {
    baseURL string
    token   string
    client  *http.Client
}

func NewClient(baseURL, token string) *Client {
    return &Client{
        baseURL: baseURL,
        token:   token,
        client: &http.Client{
            Timeout: 30 * time.Second,
        },
    }
}

func (c *Client) GetPost(ctx context.Context, id string) (*Post, error) {
    url := fmt.Sprintf("%s/v1/posts/%s", c.baseURL, id)
    
    req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
    if err != nil {
        return nil, err
    }
    
    req.Header.Set("Authorization", "Bearer "+c.token)
    req.Header.Set("Content-Type", "application/json")
    
    resp, err := c.client.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    
    if resp.StatusCode != http.StatusOK {
        var apiErr APIError
        if err := json.NewDecoder(resp.Body).Decode(&apiErr); err != nil {
            return nil, fmt.Errorf("API error: %d", resp.StatusCode)
        }
        return nil, &apiErr
    }
    
    var post Post
    if err := json.NewDecoder(resp.Body).Decode(&post); err != nil {
        return nil, err
    }
    
    return &post, nil
}

type Post struct {
    ID          string    `json:"id"`
    Title       string    `json:"title"`
    Slug        string    `json:"slug"`
    Status      string    `json:"status"`
    Content     string    `json:"content"`
    PublishedAt time.Time `json:"published_at"`
}

type APIError struct {
    Code    int               `json:"code"`
    Message string            `json:"message"`
    Details []json.RawMessage `json:"details"`
}

func (e *APIError) Error() string {
    return e.Message
}
```

## 17. 常见问题

### Q1: 如何处理大文件上传？
A: 使用分片上传机制，将大文件分割成多个小块并行上传，最后合并。

### Q2: 如何实现实时通知？
A: 使用 Server-Sent Events (SSE) 或 WebSocket 推送实时事件。

### Q3: 如何处理 API 版本兼容？
A: 在 URL 中包含版本号，保持向后兼容，提前通知废弃计划。

### Q4: 如何优化 API 性能？
A: 使用缓存、压缩、连接池、异步处理等技术手段。

### Q5: 如何保证 API 安全？
A: 使用 HTTPS、JWT 认证、限流、输入验证、输出编码等安全措施。

---

## 附录

### A. HTTP 状态码完整列表
| 状态码 | 含义 | 使用场景 |
|--------|------|----------|
| 200 | OK | 请求成功 |
| 201 | Created | 资源创建成功 |
| 204 | No Content | 请求成功但无返回内容 |
| 400 | Bad Request | 请求参数错误 |
| 401 | Unauthorized | 认证失败 |
| 403 | Forbidden | 权限不足 |
| 404 | Not Found | 资源不存在 |
| 409 | Conflict | 资源冲突 |
| 422 | Unprocessable Entity | 业务验证失败 |
| 429 | Too Many Requests | 请求频率超限 |
| 500 | Internal Server Error | 服务器内部错误 |
| 502 | Bad Gateway | 网关错误 |
| 503 | Service Unavailable | 服务不可用 |

### B. 常用工具推荐
- **API 测试**: Postman, Insomnia
- **性能测试**: JMeter, k6, wrk
- **文档生成**: Swagger UI, Redoc
- **监控**: Prometheus, Grafana
- **日志**: ELK Stack, Loki

### C. 参考资源
- [REST API 设计指南](https://restfulapi.net/)
- [HTTP 状态码规范](https://httpstatuses.com/)
- [JWT RFC 7519](https://tools.ietf.org/html/rfc7519)
- [OpenAPI 规范](https://swagger.io/specification/)
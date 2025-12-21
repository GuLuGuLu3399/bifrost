# Bifrost 接口交互规范

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
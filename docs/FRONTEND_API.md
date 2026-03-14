# BIFROST FRONTEND API

> 最后更新：2026-03-14  
> 适用：Horizon（内容站）/ Helm（管理端）

## 1. 网关与跨域

- 网关：`http://localhost:8080`
- 协议：HTTP JSON
- CORS 默认白名单（`go_services/configs/gjallar.yaml`）：
  - `http://localhost:3000`
  - `http://localhost:3001`
  - `http://localhost:3002`

## 2. 鉴权

- 方式：`Authorization: Bearer <access_token>`
- 登录：`POST /v1/auth/login`

示例：

```json
{
  "identifier": "superadmin",
  "password": "your_password"
}
```

## 3. 核心接口

### 3.1 公开读接口

- `GET /v1/posts`：文章列表（Beacon）
  - 分页参数：`page.page_size`、`page.page_token`
- `GET /v1/posts/{slug}`：文章详情
- `POST /v1/posts:batch`：批量文章摘要
- `GET /v1/categories`：分类列表
- `GET /v1/tags`：标签列表
- `GET /v1/posts/{post_id}/comments`：评论列表

### 3.2 搜索接口

- `GET /v1/search?query=...&page=1&page_size=20`
- `GET /v1/search/suggest?prefix=...&limit=5`

说明：搜索使用扁平分页参数 `page/page_size`，与 Beacon 不同。

### 3.3 登录后接口

- `POST /v1/posts/{post_id}/comments`
- `DELETE /v1/comments/{comment_id}`
- `POST /v1/storage/upload_ticket`（当前网关未注册 StorageService，现状会 404）

### 3.4 管理接口

- `GET /v1/drafts`
- `GET /v1/admin/posts/{post_id}`
- `GET /v1/admin/posts/{post_id}/source`
- `POST /v1/admin/posts`
- `PUT /v1/admin/posts/{post_id}`
- `DELETE /v1/admin/posts/{post_id}`

## 4. 前端接入注意

- int64 ID 一律按字符串处理，避免 JS 精度问题。
- Nuxt 页面事件处理（如点击登录）请使用 `$fetch`，不要在事件处理器里直接使用 `useFetch`。
- 开发环境可打印错误响应体，生产环境应降噪并统一提示。

## 5. 快速测试

```bash
curl "http://localhost:8080/v1/posts?page.page_size=20"
curl "http://localhost:8080/v1/search?query=rust&page=1&page_size=10"
curl -X POST http://localhost:8080/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"identifier":"superadmin","password":"your_password"}'
```

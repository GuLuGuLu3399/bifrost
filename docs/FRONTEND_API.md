# BIFROST FRONTEND API

> 最后更新：2026-03-14  
> 适用：Horizon（内容站）/ Helm（管理端）

## 目录

- [网关与跨域](#网关与跨域)
- [认证](#认证)
- [核心接口](#核心接口)
- [接入注意事项](#接入注意事项)
- [快速测试](#快速测试)

## 网关与跨域

- 网关地址：`http://localhost:8080`
- 协议：HTTP JSON
- CORS 默认来源（`go_services/configs/gjallar.yaml`）：
  - `http://localhost:3000`
  - `http://localhost:3001`
  - `http://localhost:3002`

## 认证

- 鉴权头：`Authorization: Bearer <access_token>`
- 登录接口：`POST /v1/auth/login`

```json
{
  "identifier": "superadmin",
  "password": "your_password"
}
```

## 核心接口

### 公开读接口

- `GET /v1/posts`（分页：`page.page_size`、`page.page_token`）
- `GET /v1/posts/{slug}`
- `POST /v1/posts:batch`
- `GET /v1/categories`
- `GET /v1/tags`
- `GET /v1/posts/{post_id}/comments`

### 搜索接口

- `GET /v1/search?query=...&page=1&page_size=20`
- `GET /v1/search/suggest?prefix=...&limit=5`

说明：

- 支持 `query` 或 `q` 作为关键词。
- 搜索分页参数使用 `page/page_size`（区别于 Beacon）。
- 支持过滤：`category_id`、`tag_id`、`author_id`。
- 参数校验：`page >= 1`、`page_size` 为 `1~100`、`limit` 为 `1~20`。
- Mirror 不可用或关闭时，搜索接口返回 `503`。

### 登录后接口

- `POST /v1/posts/{post_id}/comments`
- `DELETE /v1/comments/{comment_id}`
- `POST /v1/storage/upload_ticket`

说明：Nexus 已实现 `StorageService`，但 Gjallar 目前未注册对应 gateway handler，网关调用现状为 `404`。

### 管理接口

- `GET /v1/drafts`
- `GET /v1/admin/posts/{post_id}`
- `GET /v1/admin/posts/{post_id}/source`
- `POST /v1/admin/posts`
- `PUT /v1/admin/posts/{post_id}`
- `DELETE /v1/admin/posts/{post_id}`

## 接入注意事项

- int64 ID 请按字符串处理，避免 JS 精度丢失。
- Nuxt 事件处理（如登录点击）请使用 `$fetch`。
- 建议仅在开发环境显示完整后端错误，生产环境统一友好提示。

## 快速测试

```bash
curl "http://localhost:8080/v1/posts?page.page_size=20"
curl "http://localhost:8080/v1/search?query=rust&page=1&page_size=10"
curl -X POST http://localhost:8080/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"identifier":"superadmin","password":"your_password"}'
```

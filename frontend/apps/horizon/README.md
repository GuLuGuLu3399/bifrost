# HORIZON

Horizon 是 Bifrost 的 Nuxt 前台站点（内容展示与搜索）。

## 运行

```bash
cd frontend
pnpm install
pnpm --filter @bifrost/horizon dev
```

默认开发端口：`http://localhost:3001`。

## 环境变量

- `API_BASE`：默认 `http://localhost:8080`
- `CDN_URL`：静态资源域名

## 页面与能力

- `/`：文章列表（Beacon）
- `/posts/[slug]`：文章详情 SSR
- `/search`：全文检索（Mirror）
- `/auth/login`：登录

## 接口调用约定

- 页面 setup 场景可使用 `useApi/useFetch`。
- 用户交互事件（如点击登录）请使用 `$fetch`。
- 调试环境可展示后端错误响应，生产环境需统一错误提示。

## 相关文档

- [ARCHITECTURE](./ARCHITECTURE.md)
- [FRONTEND_API](../../../docs/FRONTEND_API.md)

# HORIZON

> Bifrost Nuxt 内容站（SSR）。

## 目录

- [运行方式](#运行方式)
- [环境变量](#环境变量)
- [页面能力](#页面能力)
- [接口约定](#接口约定)

## 运行方式

```bash
cd frontend
pnpm install
pnpm --filter @bifrost/horizon dev
```

默认开发地址：`http://localhost:3001`。

## 环境变量

- `API_BASE`: 默认 `http://localhost:8080`
- `CDN_URL`: 静态资源域名

## 页面能力

- `/`: 文章列表（Beacon）
- `/posts/[slug]`: 文章详情（SSR）
- `/search`: 全文检索（Mirror）
- `/auth/login`: 登录

## 接口约定

- setup 阶段可使用 `useApi/useFetch`。
- 用户交互事件（如点击登录）使用 `$fetch`。
- 开发环境可展示后端详细错误，生产环境统一错误提示。

## 相关文档

- [ARCHITECTURE](./ARCHITECTURE.md)
- [FRONTEND_API](../../../docs/FRONTEND_API.md)

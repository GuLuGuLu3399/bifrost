# HORIZON ARCHITECTURE

> Nuxt SSR 内容站架构说明。

## 目录

- [分层设计](#分层设计)
- [数据流](#数据流)
- [关键约束](#关键约束)

## 分层设计

- `pages/`: 路由页面
- `components/`: 展示组件
- `composables/`: 数据访问与业务逻辑
- `utils/`: 纯工具函数

## 数据流

1. 页面进入后在 setup 阶段拉取数据（列表、详情、搜索）。
2. 用户交互事件使用 `$fetch` 发起写请求（登录、评论等）。
3. API 统一通过 Gjallar（`http://localhost:8080`）。

## 关键约束

- 所有 int64 ID 在前端按字符串处理。
- SEO 关键页面优先 SSR。
- 错误处理区分环境：开发看细节，生产给统一文案。

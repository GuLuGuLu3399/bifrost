# HORIZON ARCHITECTURE

Horizon 基于 Nuxt（SSR）实现内容站点，核心目标是首屏速度、SEO 与稳定的数据流。

## 分层

- `pages/`：路由页面
- `components/`：展示组件
- `composables/`：数据访问与业务逻辑
- `utils/`：纯工具函数

## 数据流

1. 页面进入后在 setup 阶段拉取数据（列表、详情、搜索）。
2. 用户交互事件使用 `$fetch` 发起写请求（登录、评论等）。
3. API 统一经 Gjallar `http://localhost:8080`。

## 关键约束

- 所有 int64 ID 在前端按字符串处理。
- SEO 关键页面（详情页）优先 SSR。
- 错误处理区分开发/生产：开发看细节，生产给统一文案。

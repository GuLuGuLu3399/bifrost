# Horizon - Bifrost 内容展现层

> 高性能、SEO 友好的内容展现层，基于 Nuxt 3 SSR

---

## 📁 项目结构

```
apps/horizon/
├── assets/                      # 全局静态资源
├── components/
│   ├── business/               # 业务组件
│   │   ├── SearchBar.vue       # 搜索框组件（带建议）
│   │   ├── PostCard.vue        # 文章卡片
│   │   └── CommentList.vue     # 评论列表
│   └── layout/                 # 布局组件
│       ├── Header.vue          # 导航头部
│       └── Footer.vue          # 页脚
├── composables/                # Composables (业务逻辑层)
│   ├── useApi.ts               # 统一 API 封装（认证、错误处理）
│   ├── usePost.ts              # 文章数据逻辑
│   └── useSearch.ts            # 搜索逻辑（防抖、分面）
├── layouts/
│   ├── default.vue             # 默认布局（包含 Header/Footer）
│   └── blank.vue               # 空白布局（用于登录页）
├── pages/
│   ├── index.vue               # 首页（文章列表）
│   ├── posts/
│   │   └── [slug].vue          # 文章详情页（核心 SEO 页面）
│   ├── search/
│   │   └── index.vue           # 搜索结果页
│   └── auth/
│       ├── login.vue           # 登录页
│       └── register.vue        # 注册页（待实现）
├── utils/
│   └── index.ts                # 工具函数（日期、图片、字符串）
├── app.vue                     # 应用根组件
├── nuxt.config.ts              # Nuxt 配置（含 runtimeConfig）
└── package.json                # 项目依赖

```

---

## 🔌 核心 Composables 说明

### `useApi.ts` - 统一的 API 请求封装

```typescript
// 基于 useFetch 的二次封装
// 功能：
// - 自动注入 baseURL
// - 自动注入认证 Token（Cookie: auth_token）
// - 处理 401 响应（自动跳转登录）
// - 支持所有 useFetch 选项

const { data, error, pending } = useApi<T>("/v1/posts");
```

### `usePost.ts` - 文章数据逻辑

```typescript
// 获取文章列表
const { data, error, pending } = usePostList(pageSize, categoryId, tagId);

// 获取文章详情
const { data, error } = await usePostDetail(slug);

// 批量获取文章摘要
const { data } = useApi("/v1/posts:batch", {
  method: "POST",
  body: { post_ids: [1, 2, 3] },
});

// 工具函数
const imageUrl = getImageUrl(coverImageKey);
const date = formatDate("2025-01-01T12:00:00Z"); // '2025年1月1日'
```

### `useSearch.ts` - 搜索逻辑（防抖 + 分面）

```typescript
const {
  query,                  // ref: 搜索关键词
  searchResult,          // 搜索结果
  suggestions,           // 搜索建议
  showSuggestions,       // computed: 是否显示建议
  pending,               // 搜索中
  selectSuggestion(q),   // 方法：选择建议
} = useSearch();

// 特点：
// - 300ms 防抖，避免频繁请求
// - 自动获取分面（categories, tags）
// - 支持日期范围过滤
```

---

## 🎯 页面说明

### 首页 (`pages/index.vue`)

- **功能**：展示文章列表
- **数据源**：`GET /v1/posts?page.page_size=20`
- **组件**：使用 `BusinessPostCard` 渲染每个文章卡片
- **SEO**：自动注入 Open Graph 标签（可在详情页增强）

### 文章详情页 (`pages/posts/[slug].vue`)

- **功能**：展示完整文章内容
- **数据源**：`GET /v1/posts/{slug}`
- **SSR**：支持服务端渲染（预加载数据）
- **SEO**：
  - 注入 `<title>`, `<meta name="description">`
  - 注入 Open Graph (`og:title`, `og:image` 等)
  - 结构化数据（可扩展）
- **特点**：
  - 使用 Tailwind `prose` 类美化 HTML 内容
  - 计算阅读时间
  - 上/下一篇导航
  - 目录展示（toc_json）

### 搜索结果页 (`pages/search/index.vue`)

- **功能**：全文搜索与分面导航
- **数据源**：`GET /v1/search?query=...&include_facets=true`
- **特点**：
  - 高亮搜索词（`highlight_title`, `highlight_content`）
  - 相关度排序
  - 分类/标签分面
  - 搜索建议（下拉菜单）

### 登录页 (`pages/auth/login.vue`)

- **功能**：用户认证
- **数据源**：`POST /v1/auth/login`
- **响应处理**：保存 `access_token` 到 Cookie
- **错误处理**：显示错误提示

---

## 🛠️ 业务组件

### `SearchBar.vue`

```vue
<BusinessSearchBar @search="handleSearch" />
```

- 搜索框 + 建议下拉菜单
- 防抖搜索建议请求
- 事件：`@search(query)`

### `PostCard.vue`

```vue
<BusinessPostCard :post="postItem" />
```

- 文章卡片：封面 + 标题 + 摘要 + 元数据
- 链接到文章详情页

### `CommentList.vue`

```vue
<CommentList :comments="comments" @replyTo="replyToComment" />
```

- 评论列表展示
- 支持楼中楼（待实现）

---

## ⚙️ 环境配置

在 `nuxt.config.ts` 中配置：

```typescript
runtimeConfig: {
  public: {
    apiBase: process.env.API_BASE || "http://localhost:8080",
    cdnUrl: process.env.CDN_URL || "https://cdn.example.com",
  }
}
```

**环境变量示例** (`.env.local`):

```bash
API_BASE=http://localhost:8080
CDN_URL=https://cdn.example.com
```

---

## 📦 依赖说明

| 依赖                    | 用途           |
| ----------------------- | -------------- |
| `nuxt@^4.2.2`           | 框架           |
| `vue@^3.5.26`           | UI 库          |
| `vue-router@^4.6.4`     | 路由           |
| `@bifrost/ui`           | 共享 UI 组件库 |
| `@bifrost/shared`       | 共享类型定义   |
| `@vueuse/core@^10.10.0` | 防抖等工具函数 |
| `tailwindcss@^3.4.14`   | CSS 框架       |

---

## 🚀 快速开始

```bash
# 安装依赖
pnpm install

# 开发模式（热重载）
pnpm dev

# 生产构建
pnpm build

# 预览构建结果
pnpm preview

# 生成静态网站（可选）
pnpm generate
```

---

## 🔄 工作流

### 用户浏览文章

```
Browser (GET /posts/hello-world)
    ↓
Nuxt SSR Server (预加载数据)
    ↓
usePostDetail('/v1/posts/hello-world')
    ↓
useApi() → Gjallar Gateway (:8080)
    ↓
Beacon Service (读取 html_body + toc_json)
    ↓
响应包含：title, html_body, cover_image_key, ...
    ↓
Server-side 注入 SEO Meta + 返回 HTML
    ↓
浏览器渲染 + Hydration
```

### 用户搜索

```
用户输入 "Rust" → 300ms 防抖
    ↓
useSearch() → useFetch('/v1/search?query=Rust&include_facets=true')
    ↓
Mirror Service 返回：hits[], facets{categories, tags}
    ↓
前端显示：结果列表 + 分面导航 + 高亮词
```

---

## 📋 下一步计划

- [ ] 实现评论系统与回复
- [ ] 实现图片上传（直传 MinIO）
- [ ] 实现点赞功能
- [ ] 添加 Dark Mode
- [ ] 实现无限滚动分页
- [ ] 完善错误边界（Error Boundary）
- [ ] 添加加载骨架屏
- [ ] SEO：Sitemap + robots.txt

---

## 📚 参考文档

- [Nuxt 3 官方文档](https://nuxt.com)
- [Vue 3 组合式 API](https://vuejs.org/guide/extras/composition-api-faq.html)
- [Tailwind CSS](https://tailwindcss.com)
- [VueUse](https://vueuse.org)

---

## 📝 架构原则

### 1. 关注点分离

- **Composables** (`useApi`, `usePost`, `useSearch`)：业务逻辑
- **Components** (`BusinessPostCard`, `SearchBar`)：UI 展示
- **Utils** (`formatDate`, `getImageUrl`)：工具函数

### 2. SSR 友好

- 所有数据获取在 `<script setup>` 顶层（支持 SSR 预加载）
- 使用 `await` 等待数据，框架自动处理 hydration

### 3. 类型安全

- 所有 API 响应都有 TypeScript 类型（`@bifrost/shared`）
- IDE 自动补全，减少运行时错误

### 4. SEO 优化

- Meta 标签注入（`useSeoMeta`)
- 结构化数据支持
- 动态路由 `[slug]` 静态生成

---

**作者**: 架构师  
**最后更新**: 2025-01-01

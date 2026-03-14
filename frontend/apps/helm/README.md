# HELM

> Bifrost Tauri 管理端（Vue + Rust Sidecar）。

## 目录

- [核心能力](#核心能力)
- [开发运行](#开发运行)
- [构建](#构建)
- [常用命令](#常用命令)
- [配置](#配置)

## 核心能力

- 登录与令牌管理（`/v1/auth/login`）
- 图片处理（压缩、WebP 转换、尺寸限制）
- 上传流程（ticket + 对象存储直传）
- Tauri 命令桥接（`invoke`）
- 路由守卫与会话检查（未登录跳转 `/auth`）
- 管理页面：`/posts`、`/taxonomy`、`/comments`、`/profile`
- 接口实验页：`/lab/api`

## 开发运行

```bash
cd frontend
pnpm install
pnpm --filter @bifrost/helm tauri dev
```

## 构建

```bash
cd frontend
pnpm --filter @bifrost/helm tauri build
```

## 常用命令

- `login_cmd(identifier, password)`
- `upload_image_cmd(filePath)`
- `is_authenticated()`
- `logout_cmd()`
- `gateway_request_cmd(method, path, query, body, authRequired)`

## 配置

- 默认网关：`http://localhost:8080`
- 可通过环境变量 `GJALLAR_URL` 覆盖

## 相关文档

- [IMPLEMENTATION](./IMPLEMENTATION.md)
- [FRONTEND_API](../../../docs/FRONTEND_API.md)

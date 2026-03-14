# HELM

Helm 是 Bifrost 的 Tauri 管理端应用，采用 Vue + Rust Sidecar 模式。

## 核心能力

- 登录与令牌管理（调用 Gjallar `/v1/auth/login`）
- 图片处理（压缩、转 WebP、尺寸约束）
- 上传流程（获取 ticket + 直传对象存储）
- 管理端命令桥接（Tauri `invoke`）

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

## 配置

- 后端默认网关：`http://localhost:8080`
- 如需覆盖，请在 Tauri/Rust 侧通过环境变量配置 `GJALLAR_URL`

## 关联文档

- [IMPLEMENTATION](./IMPLEMENTATION.md)
- [FRONTEND_API](../../../docs/FRONTEND_API.md)

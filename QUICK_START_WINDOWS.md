# Windows 快速开发指南

## 🚀 一键快速开始

```powershell
# 生成 Proto 代码 + 显示启动步骤
.\manage.ps1 quick-start
```

## 📋 常用命令

### Proto 和网关相关

```powershell
# 生成 Proto 代码（Protobuf + gRPC + gRPC Gateway）
.\manage.ps1 proto-gen

# 检查 Proto 文件规范
.\manage.ps1 proto-lint

# 测试 gRPC Gateway 网关
.\manage.ps1 gateway-test
```

### 启动微服务（需要在不同的终端中运行）

```powershell
# 终端 1: 启动 Beacon 服务（读服务，端口 50051）
.\manage.ps1 run-beacon

# 终端 2: 启动 Nexus 服务（写服务，端口 50052）
.\manage.ps1 run-nexus

# 终端 3: 启动 Gjallar 网关（HTTP 网关，端口 8080）
.\manage.ps1 run-gjallar
```

### 构建相关

```powershell
# 构建所有 Go 服务（生成到 ./bin/）
.\manage.ps1 build-go

# 构建所有 Rust 服务
.\manage.ps1 build-rust
```

### 数据库相关

```powershell
# 执行数据库迁移
.\manage.ps1 migrate-up

# 创建新迁移
.\manage.ps1 migrate-new -Name create_posts_table

# 连接到数据库
.\manage.ps1 db-connect
```

### 基础设施相关

```powershell
# 启动 Docker 基础设施（数据库、缓存等）
.\manage.ps1 up

# 停止 Docker 基础设施
.\manage.ps1 down

# 查看容器日志
.\manage.ps1 logs

# 列出运行中的容器
.\manage.ps1 ps
```

### 工具命令

```powershell
# 初始化项目
.\manage.ps1 init

# 格式化代码
.\manage.ps1 format

# 清理构建产物
.\manage.ps1 clean

# 显示帮助信息
.\manage.ps1 help

# 快速启动提示
.\manage.ps1 dev
```

## 🔄 完整开发流程

### 第一次开发

```powershell
# 1️⃣ 初始化项目
.\manage.ps1 init

# 2️⃣ 生成 Proto 代码
.\manage.ps1 quick-start

# 3️⃣ 启动基础设施
.\manage.ps1 up

# 4️⃣ 在不同终端启动三个微服务
# Terminal 1
.\manage.ps1 run-beacon

# Terminal 2
.\manage.ps1 run-nexus

# Terminal 3
.\manage.ps1 run-gjallar

# 5️⃣ 测试网关
.\manage.ps1 gateway-test
```

### 日常开发（Proto 有改动）

```powershell
# 1️⃣ 修改 *.proto 文件
# 编辑 api/content/v1/beacon/beacon.proto 等

# 2️⃣ 生成代码
.\manage.ps1 proto-gen

# 3️⃣ 重启微服务（或让热重载自动处理）
# 之前启动的终端会自动重新编译
```

### 日常开发（无 Proto 改动）

```powershell
# 启动三个微服务即可
# Terminal 1
.\manage.ps1 run-beacon

# Terminal 2
.\manage.ps1 run-nexus

# Terminal 3
.\manage.ps1 run-gjallar
```

## 🌐 网关 API 调用示例

服务启动后可以直接访问 HTTP 接口：

```powershell
# 获取文章列表
curl http://localhost:8080/v1/posts

# 获取分类
curl http://localhost:8080/v1/categories

# 注册用户
curl -X POST http://localhost:8080/v1/auth/register `
  -H "Content-Type: application/json" `
  -d '{
    "username": "testuser",
    "email": "test@example.com",
    "password": "password123",
    "nickname": "Test User"
  }'

# 创建文章
curl -X POST http://localhost:8080/v1/posts `
  -H "Content-Type: application/json" `
  -d '{
    "title": "My Post",
    "slug": "my-post",
    "raw_markdown": "# Hello World",
    "category_id": 1,
    "tag_names": ["test"],
    "status": 1
  }'
```

## 📚 更多信息

- 详细的 gRPC Gateway 使用指南：[GRPC_GATEWAY_GUIDE.md](./GRPC_GATEWAY_GUIDE.md)
- 集成完成清单：[GRPC_GATEWAY_INTEGRATION_CHECKLIST.md](./GRPC_GATEWAY_INTEGRATION_CHECKLIST.md)
- 网关集成示例：[docs/GJALLAR_INTEGRATION_EXAMPLE.go](./docs/GJALLAR_INTEGRATION_EXAMPLE.go)

## ⚙️ 系统要求

- **Windows 10/11** with PowerShell 5.1+
- **Go 1.25.4+**
- **Rust 1.70+**
- **Docker Desktop** (可选，用于基础设施)
- **buf** (用于 Proto 生成)
- **protoc** (protobuf 编译器)

### 快速安装依赖

```powershell
# 使用 Scoop 或 Chocolatey

# Go
choco install golang

# Rust
choco install rust

# Docker Desktop
choco install docker-desktop

# buf
go install github.com/bufbuild/buf/cmd/buf@latest

# protoc
choco install protoc
```

## 🐛 常见问题

### Q: "buf not found" 错误？

A: 需要安装 buf。运行 `go install github.com/bufbuild/buf/cmd/buf@latest`

### Q: "protoc not found" 错误？

A: 虽然现在用 buf，但如果遇到问题可以安装 protoc。运行 `choco install protoc`

### Q: 端口已被占用？

A: 修改配置文件：

- Beacon: `go_services/configs/beacon.yaml` (修改 grpc_port)
- Nexus: `go_services/configs/nexus.yaml` (修改 grpc_port)
- Gjallar: `go_services/cmd/gjallar/main.go` (修改 http 端口)

### Q: 网关无法连接到微服务？

A: 检查微服务是否已启动，以及 gjallar 的配置文件中的地址是否正确。

### Q: 如何查看网关日志？

A: 网关启动的终端会输出详细日志。如果需要更详细的日志，修改配置文件中的 `log_level`。

## 💡 提示

- 使用 `.\manage.ps1 help` 查看所有可用命令
- 每个命令都会输出执行步骤和结果
- 出错时会显示详细的错误信息
- 所有命令都支持 `-ErrorAction Continue` 继续执行

---

**最后更新：** 2025-12-21  
**版本：** Bifrost CMS v3.2

# ===========================================
# Build Stage
# ===========================================
FROM golang:1.25.4-alpine AS builder

# 安装必要的构建工具
RUN apk add --no-cache \
    git \
    ca-certificates \
    tzdata

WORKDIR /build

# 1. 复制 go.mod 和 go.sum（利用 Docker 缓存）
COPY go.mod go.sum ./

# 2. 下载依赖（这一层会被缓存）
RUN go mod download && go mod verify

# 3. 复制源代码
COPY . .

# 4. 构建指定的服务
ARG SERVICE_NAME
RUN test -n "$SERVICE_NAME" || (echo "SERVICE_NAME build arg is required" && exit 1)

# 构建参数优化
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -a -installsuffix cgo \
    -ldflags="-w -s" \
    -o /app/server \
    ./cmd/${SERVICE_NAME}

# ===========================================
# Runtime Stage
# ===========================================
FROM alpine:latest

# 安装运行时依赖
RUN apk --no-cache add \
    ca-certificates \
    tzdata \
    curl && \
    update-ca-certificates

# 安装 gRPC 健康探针（用于健康检查）
RUN GRPC_HEALTH_PROBE_VERSION=v0.4.19 && \
    wget -qO /bin/grpc_health_probe \
    https://github.com/grpc-ecosystem/grpc-health-probe/releases/download/${GRPC_HEALTH_PROBE_VERSION}/grpc_health_probe-linux-amd64 && \
    chmod +x /bin/grpc_health_probe

# 创建非 root 用户
RUN addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser

WORKDIR /app

# 复制二进制文件
COPY --from=builder --chown=appuser:appuser /app/server /app/server

# 复制配置文件
COPY --from=builder --chown=appuser:appuser /build/configs /app/configs

# 设置时区
ENV TZ=Asia/Shanghai

# 切换到非 root 用户
USER appuser

# 暴露常用端口
EXPOSE 8080 9001 9002

# 运行服务
ENTRYPOINT ["/app/server"]

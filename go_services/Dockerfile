# Build Stage
FROM golang:1.25.4-alpine AS builder

# 安装基础工具
RUN apk add --no-cache git

WORKDIR /app

# 预下载依赖 (利用 Docker 缓存)
COPY go.mod go.sum ./
RUN go mod download

# 复制源码
COPY . .

# 接收构建参数: SERVICE_NAME (nexus/beacon/gjallar)
ARG SERVICE_NAME
RUN go build -o /bin/server ./cmd/${SERVICE_NAME}

# Run Stage
FROM alpine:latest

# 安装 CA 证书 (HTTPS 连接需要)
RUN apk --no-cache add ca-certificates wget

# 安装 gRPC 健康探针（供 healthcheck 使用）
RUN GRPC_HEALTH_PROBE_VERSION=v0.4.19 \
    && wget -qO /bin/grpc_health_probe https://github.com/grpc-ecosystem/grpc-health-probe/releases/download/${GRPC_HEALTH_PROBE_VERSION}/grpc_health_probe-linux-amd64 \
    && chmod +x /bin/grpc_health_probe

WORKDIR /app

# 复制二进制文件
COPY --from=builder /bin/server /app/server

# 复制配置文件 (假设 configs 在根目录)
COPY --from=builder /app/configs /app/configs

# 暴露端口 (虽然每个服务不一样，但可以统一暴露常用的)
EXPOSE 8080 9001 9002

ENTRYPOINT ["/app/server"]

# Multi-service Dockerfile for bifrost rust microservices (forge/mirror/oracle)
# Build arg SERVICE selects the binary to compile.

FROM rust:1.91.1-bookworm AS builder

ARG SERVICE
WORKDIR /workspace

# 安装 protoc（Protocol Buffers 编译器）
RUN apt-get update && \
  apt-get install -y protobuf-compiler && \
  rm -rf /var/lib/apt/lists/*

# 1. 先复制 Cargo 文件以利用 Docker 缓存层
COPY rust_services/Cargo.toml rust_services/Cargo.lock ./

# 2. 创建虚拟源文件以预编译依赖（加速后续构建）
RUN mkdir -p common/src forge/src mirror/src oracle/src && \
  echo "fn main() {}" > common/src/lib.rs && \
  echo "fn main() {}" > forge/src/main.rs && \
  echo "fn main() {}" > mirror/src/main.rs && \
  echo "fn main() {}" > oracle/src/main.rs

# 3. 复制各服务的 Cargo.toml
COPY rust_services/common/Cargo.toml ./common/
COPY rust_services/forge/Cargo.toml ./forge/
COPY rust_services/mirror/Cargo.toml ./mirror/
COPY rust_services/oracle/Cargo.toml ./oracle/

# 4. 预编译依赖（这一层会被缓存）
RUN test -n "$SERVICE" || (echo "SERVICE build arg is required" && exit 1) && \
  cargo build --release -p ${SERVICE} && \
  rm -rf common/src forge/src mirror/src oracle/src target/release/${SERVICE}* target/release/deps/${SERVICE}*

# 5. 复制 API proto 文件（build.rs 需要）
COPY api ./api

# 5.1 拉取 googleapis proto（供 protoc 解析 google/api/* 导入）
RUN apt-get update && \
  apt-get install -y git && \
  rm -rf /var/lib/apt/lists/* && \
  git clone --depth 1 https://github.com/googleapis/googleapis.git third_party/googleapis

# 6. 复制真实源代码
COPY rust_services/common ./common
COPY rust_services/forge ./forge
COPY rust_services/mirror ./mirror
COPY rust_services/oracle ./oracle

# 7. 构建真实的服务二进制（利用之前缓存的依赖）
RUN cargo build --release -p ${SERVICE}

# ----- Runtime image -----
FROM debian:bookworm-slim AS runtime

ARG SERVICE
ENV SERVICE=${SERVICE}

# 安装运行时依赖
# - ca-certificates: HTTPS/gRPC TLS
# - netcat-openbsd: TCP 连通性检查
RUN apt-get update && \
  apt-get install -y --no-install-recommends \
  ca-certificates \
  netcat-openbsd && \
  rm -rf /var/lib/apt/lists/*

# 创建非 root 用户
RUN groupadd -r rust && useradd -r -g rust rust

WORKDIR /app

# 创建数据目录并设置所有权（用于 Mirror 的索引存储和 Oracle 的数据库）
RUN mkdir -p /data /app/data && chown -R rust:rust /data /app/data

# 复制二进制文件
COPY --from=builder /workspace/target/release/${SERVICE} /usr/local/bin/app

# 赋予执行权限并更改所有者
RUN chmod +x /usr/local/bin/app && \
  chown rust:rust /usr/local/bin/app

# 切换到非 root 用户
USER rust

# 环境变量
ENV RUST_LOG=info
ENV RUST_BACKTRACE=1

# 暴露常用端口（根据服务不同可以在 compose 中覆盖）
EXPOSE 9092 9093 9094

# 运行服务
ENTRYPOINT ["/usr/local/bin/app"]

# Multi-service Dockerfile for bifrost rust microservices (forge/mirror/oracle)
# Build arg SERVICE selects the binary to compile.

FROM rust:1.85-bookworm AS builder

ARG SERVICE
WORKDIR /workspace

# Faster builds with cached deps
COPY Cargo.toml Cargo.lock ./
COPY common ./common
COPY forge ./forge
COPY mirror ./mirror
COPY oracle ./oracle

# Build the selected service binary
RUN test -n "$SERVICE" || (echo "SERVICE build arg is required" && exit 1)
RUN cargo build --release -p ${SERVICE}

# ----- Runtime image -----
FROM debian:bookworm-slim AS runtime

ARG SERVICE
ENV SERVICE=${SERVICE}

# Minimal runtime deps (TLS roots for reqwest/otlp, etc.)
RUN apt-get update \
  && apt-get install -y --no-install-recommends ca-certificates \
  && rm -rf /var/lib/apt/lists/*

WORKDIR /app

# Copy binary
COPY --from=builder /workspace/target/release/${SERVICE} /usr/local/bin/${SERVICE}

# Reasonable defaults; override in compose if needed
ENV RUST_LOG=info

# Run the selected binary
ENTRYPOINT ["/usr/local/bin/sh", "-lc", "/usr/local/bin/${SERVICE}"]

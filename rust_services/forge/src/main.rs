// rust_services/forge/src/main.rs

mod engine;
mod server;

use crate::server::GrpcServer;
use common::config::{AppConfig, ConfigLoader, ServerConfig};
use common::metrics;
use common::forge::render_service_server::RenderServiceServer;
use std::net::SocketAddr;
use tonic::transport::Server;
use tonic_health;
use tracing::info;

#[tokio::main]
async fn main() -> Result<(), Box<dyn std::error::Error>> {
    // 1. 加载配置
    // 优先读取 APP_FORGE__* 环境变量，其次读取 config/forge.toml
    let config: AppConfig = ConfigLoader::new()
        .with_file("config/forge") // 不带后缀，config crate 会自动找 .toml/.json/etc
        .with_env_prefix("APP_FORGE") // 区分不同服务的环境变量
        .load()?;

    // 检查必要的配置是否存在
    let server_cfg: &ServerConfig = config
        .server
        .as_ref()
        .ok_or_else(|| anyhow::anyhow!("missing required config: server.addr"))?;

    // 2. 初始化 tracing
    let json = config
        .log
        .as_ref()
        .and_then(|l| l.format.as_deref())
        .map(|s| s == "json")
        .unwrap_or(false);
    common::logger::init_tracing("forge", "dev", json)?;

    info!(addr = %server_cfg.addr, "forge starting");

    // Start Prometheus metrics on default port for forge
    metrics::init_prometheus("0.0.0.0:9103").await?;

    let addr: SocketAddr = server_cfg.addr.parse()?;
    let forge_service: GrpcServer = GrpcServer::new();

    // 初始化健康检查
    let ( _health_reporter, health_service) = tonic_health::server::health_reporter();

    info!(%addr, "forge (renderer) listening");

    // 3. 启动 Server
    Server::builder()
        .add_service(RenderServiceServer::new(forge_service))
        .add_service(health_service)
        .serve_with_shutdown(addr, common::lifecycle::shutdown_signal())
        .await?;

    Ok(())
}
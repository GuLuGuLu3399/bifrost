// rust_services/forge/src/main.rs

mod engine;
mod server;

use crate::server::GrpcServer;
use common::config::{AppConfig, ConfigLoader, ServerConfig};
use common::forge::render_service_server::RenderServiceServer;
use std::net::SocketAddr;
use tonic::transport::Server;
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
        .expect("server config missing");

    // 2. 初始化 tracing
    let json = config
        .log
        .as_ref()
        .and_then(|l| l.format.as_deref())
        .map(|s| s == "json")
        .unwrap_or(false);
    common::logger::init_tracing("forge", "dev", json)?;

    info!(addr = %server_cfg.addr, "forge starting");

    let addr: SocketAddr = server_cfg.addr.parse()?;
    let forge_service: GrpcServer = GrpcServer::new();

    info!(%addr, "forge (renderer) listening");

    // 3. 启动 Server
    Server::builder()
        .add_service(RenderServiceServer::new(forge_service))
        .serve(addr)
        .await?;

    Ok(())
}
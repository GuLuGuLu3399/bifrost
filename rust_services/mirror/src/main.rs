// mirror/src/main.rs

mod engine;
mod server;
mod worker;
mod tokenizer;

use crate::engine::SearchEngine;
use crate::server::GrpcServer;
use crate::worker::IndexWorker;
use common::config::{AppConfig, ConfigLoader, ServerConfig};
use common::search::mirror_service_server::MirrorServiceServer;
use std::sync::Arc;
use tonic::transport::Server;
use tracing::{error, info};

#[tokio::main]
async fn main() -> Result<(), Box<dyn std::error::Error>> {
    // 1. 加载配置（风格对齐 forge）
    let config: AppConfig = ConfigLoader::new()
        .with_file("config/mirror")
        .with_env_prefix("APP_MIRROR")
        .load()?;

    let server_cfg: &ServerConfig = config
        .server
        .as_ref()
        .expect("Server config missing");

    let nats_cfg = config.nats.as_ref().expect("NATS config missing");

    // 2. 初始化 tracing（风格对齐 forge）
    let json = config
        .log
        .as_ref()
        .and_then(|l| l.format.as_deref())
        .map(|s| s == "json")
        .unwrap_or(false);

    common::logger::init_tracing("mirror", "dev", json)?;

    info!(addr = %server_cfg.addr, "mirror starting");

    // 3. 初始化核心搜索引擎
    let index_path = std::env::var("MIRROR_INDEX_PATH")
        .unwrap_or_else(|_| "./data/tantivy_index".to_string());
    let engine = Arc::new(SearchEngine::new(Some(&index_path))?);
    info!("Search Engine initialized at: {}", index_path);

    // 4. 启动 NATS Index Worker (后台运行)
    let worker_engine = engine.clone();
    let worker_nats_cfg = nats_cfg.clone();

    tokio::spawn(async move {
        match IndexWorker::new(&worker_nats_cfg, worker_engine).await {
            Ok(worker) => {
                if let Err(e) = worker.run().await {
                    error!("IndexWorker crashed: {:?}", e);
                }
            }
            Err(e) => {
                error!("Failed to start IndexWorker: {:?}", e);
            }
        }
    });

    // 5. 启动 gRPC Server
    let addr = server_cfg.addr.parse()?;
    let mirror_service = GrpcServer::new(engine);

    info!(%addr, "mirror listening");

    Server::builder()
        .add_service(MirrorServiceServer::new(mirror_service))
        .serve_with_shutdown(addr, async {
            tokio::signal::ctrl_c().await.unwrap();
            info!("Shutting down Mirror service...");
        })
        .await?;

    Ok(())
}

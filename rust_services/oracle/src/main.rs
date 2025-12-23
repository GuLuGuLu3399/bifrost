// rust_services/oracle/src/main.rs

mod ingestion;
mod server;
mod storage;

use crate::ingestion::Ingestor;
use crate::server::OracleServer;
use crate::storage::duck::DuckStore;
use common::config::ConfigLoader;
use common::logger;
use common::oracle::analysis_service_server::AnalysisServiceServer;
use tonic::transport::Server;
use tracing::info;

#[tokio::main]
async fn main() -> Result<(), Box<dyn std::error::Error>> {
    // 1. 加载配置
    let config = ConfigLoader::new()
        .with_env_prefix("APP_ORACLE")
        .load()?;

    // 2. 初始化日志
    logger::init_tracing("oracle", "dev", false)?;

    // 3. 初始化 DuckDB 存储
    // 确保 data 目录存在
    let db_path = "data/analytics.db";
    // 如果是 Docker 环境，可能需要从配置读取路径
    if let Some(parent) = std::path::Path::new(db_path).parent() {
        std::fs::create_dir_all(parent)?;
    }

    let store = DuckStore::open(db_path)?;

    // 4. 初始化摄入缓冲 (Ingestor)
    // 注意：Ingestor 会自动启动后台 tokio 任务
    let ingestor = Ingestor::new(store.clone());

    // 5. 启动 gRPC 服务
    let addr = config.server.as_ref().unwrap().addr.parse()?;
    let service = OracleServer::new(ingestor, store);

    info!("🔮 Oracle (BI) listening on {}", addr);
    info!("🦆 DuckDB path: {}", db_path);

    Server::builder()
        .add_service(AnalysisServiceServer::new(service))
        .serve(addr)
        .await?;

    Ok(())
}
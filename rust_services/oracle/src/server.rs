// rust_services/oracle/src/service.rs

use crate::ingestion::{EventPayload, Ingestor};
use crate::storage::duck::DuckStore;
use common::oracle::analysis_service_server::AnalysisService;
use common::oracle::{
    GetContentScoreRequest, GetContentScoreResponse, GetDashboardStatsRequest,
    GetDashboardStatsResponse, TrackEventRequest, TrackEventResponse,
};
use tonic::{Request, Response, Status};
use chrono::Utc;
use tracing::{error};

pub struct OracleServer {
    ingestor: Ingestor,
    store: DuckStore,
}

impl OracleServer {
    pub fn new(ingestor: Ingestor, store: DuckStore) -> Self {
        Self { ingestor, store }
    }
}

#[tonic::async_trait]
impl AnalysisService for OracleServer {
    // 1. 埋点上报接口 (极速返回)
    async fn track_event(
        &self,
        request: Request<TrackEventRequest>,
    ) -> Result<Response<TrackEventResponse>, Status> {
        let req = request.into_inner();

        // 转换为内部 EventPayload
        // 注意：meta 是 map<string, string>，我们需要转为 serde_json::Value
        let meta_json = serde_json::to_value(req.meta).unwrap_or(serde_json::json!({}));

        let payload = EventPayload {
            event_type: req.event_type,
            target_id: req.target_id,
            user_id: req.user_id,
            meta: meta_json,
            created_at: Utc::now(),
        };

        // 提交给 Ingestor (非阻塞)
        self.ingestor.submit(payload).await;

        Ok(Response::new(TrackEventResponse { accepted: true }))
    }

    // 2. 仪表盘数据查询 (OLAP 查询)
    async fn get_dashboard_stats(
        &self,
        request: Request<GetDashboardStatsRequest>,
    ) -> Result<Response<GetDashboardStatsResponse>, Status> {
        let req = request.into_inner();

        // 计算查询天数范围 (简单处理：用开始时间到现在的时间差，或者默认7天)
        let days = if req.start_time > 0 {
            let now = Utc::now().timestamp();
            ((now - req.start_time) / 86400).max(1)
        } else {
            7 // 默认查过去 7 天
        };

        // 调用 DuckDB 查询趋势
        // get_trend 是我们在 duck.rs 里实现的
        let trend_vec = self.store.get_trend(days).map_err(|e| {
            error!("Failed to query trend data: {}", e);
            Status::internal("Failed to query analytics data")
        })?;

        // 转换格式: Vec<(String, i64)> -> Map<String, i64>
        let mut trend_data = std::collections::HashMap::new();
        let mut total_views = 0;

        for (day, pv) in trend_vec {
            trend_data.insert(day, pv);
            total_views += pv;
        }

        // Mock 其他数据 (因为我们只有一张 raw_events 表，还没算 uv/users)
        // 在真实场景中，这里应该查询 daily_stats 聚合表
        Ok(Response::new(GetDashboardStatsResponse {
            total_users: 0,       // TODO: 需要 count(distinct user_id)
            active_users_today: 0, // TODO: 需要查询今天的 uv
            total_posts: 0,       // Oracle 不存文章元数据，这个字段可能需要去 Nexus 查
            total_views,          // 这是真实的 PV 总和
            trend_data,           // 这是真实的每日趋势
        }))
    }

    // 3. 内容评分 (暂未实现算法，返回 Mock)
    async fn get_content_score(
        &self,
        _request: Request<GetContentScoreRequest>,
    ) -> Result<Response<GetContentScoreResponse>, Status> {
        Ok(Response::new(GetContentScoreResponse {
            score: 88,
            suggestions: vec!["Good engagement".to_string()],
        }))
    }
}
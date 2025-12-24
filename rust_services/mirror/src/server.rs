use crate::engine::SearchEngine;
use common::search::mirror_service_server::MirrorService;
use common::search::{
    DebugIndexRequest, DebugIndexResponse, SearchRequest, SearchResponse, SuggestRequest,
    SuggestResponse,
};
use common::ctx::ContextData;
use std::sync::Arc;
use std::time::Instant;
use tonic::{Request, Response, Status};
use tracing::{info, warn, instrument};

pub struct GrpcServer {
    engine: Arc<SearchEngine>,
}

impl GrpcServer {
    pub fn new(engine: Arc<SearchEngine>) -> Self {
        Self { engine }
    }

    /// 将 Engine 错误转换为 gRPC Status
    fn map_engine_error(e: anyhow::Error) -> Status {
        warn!("Engine error: {:?}", e);
        Status::internal(format!("Search engine error: {}", e))
    }
}

#[tonic::async_trait]
impl MirrorService for GrpcServer {
    /// 核心搜索接口
    #[instrument(skip(self, request), fields(query = %request.get_ref().query))]
    async fn search(
        &self,
        request: Request<SearchRequest>,
    ) -> Result<Response<SearchResponse>, Status> {
        let start = Instant::now();
        let md = request.metadata().clone();
        common::trace::set_parent_from_metadata(&md);
        
        // ✅ 从 metadata 提取上下文，设置 tracing span 的字段
        let ctx = ContextData::from_metadata(&md);
        if let Some(rid) = &ctx.request_id {
            tracing::Span::current().record("request_id", rid.as_str());
        }
        if let Some(uid) = ctx.user_id {
            tracing::Span::current().record("user_id", uid);
        }
        
        let req = request.into_inner();

        // 1. 参数校验与默认值
        let page_size = req
            .page
            .as_ref()
            .map(|p| p.page_size)
            .unwrap_or(10)
            .clamp(1, 100) as usize;

        // 2. 使用 spawn_blocking 包装 CPU 密集型搜索操作
        //    避免阻塞 Tokio 异步运行时
        let engine = Arc::clone(&self.engine);
        let query = req.query.clone();

        let (results, total_hits) = tokio::task::spawn_blocking(move || {
            engine.search(&query, page_size)
        })
        .await
        .map_err(|e| Status::internal(format!("Task join error: {}", e)))?
        .map_err(Self::map_engine_error)?;

        // 3. 转换结果
        //    架构策略：Mirror 只负责召回 ID + score，详情由 BFF 去 Beacon 查
        let hits = results
            .into_iter()
            .map(|(id, score)| common::search::search_response::Hit {
                id,
                score,
                title: String::new(),
                slug: String::new(),
                highlight_title: String::new(),
                highlight_content: String::new(),
                published_at: 0,
            })
            .collect();

        let took_ms = start.elapsed().as_secs_f32() * 1000.0;
        info!(query = %req.query, hits = total_hits, took_ms = %took_ms, "Search completed");

        Ok(Response::new(SearchResponse {
            hits,
            total_hits: total_hits as i32,
            took_ms,
            facets: None, // TODO: 分面统计
        }))
    }

    /// 搜索建议 (Suggest)
    #[instrument(skip(self, request), fields(prefix = %request.get_ref().prefix))]
    async fn suggest(
        &self,
        request: Request<SuggestRequest>,
    ) -> Result<Response<SuggestResponse>, Status> {
        let start = Instant::now();
        let md = request.metadata().clone();
        common::trace::set_parent_from_metadata(&md);
        
        // ✅ 从 metadata 提取上下文，设置 tracing span 的字段
        let ctx = ContextData::from_metadata(&md);
        if let Some(rid) = &ctx.request_id {
            tracing::Span::current().record("request_id", rid.as_str());
        }
        if let Some(uid) = ctx.user_id {
            tracing::Span::current().record("user_id", uid);
        }
        
        let _req = request.into_inner();

        // TODO: 使用 Tantivy 的 TermDictionary/FST 实现前缀匹配
        // 目前返回空结果

        let took_ms = start.elapsed().as_secs_f32() * 1000.0;

        Ok(Response::new(SuggestResponse {
            suggestions: vec![],
            took_ms,
        }))
    }

    /// 索引调试接口 (运维用)
    #[instrument(skip(self, request), fields(doc_id = %request.get_ref().doc_id))]
    async fn debug_index(
        &self,
        request: Request<DebugIndexRequest>,
    ) -> Result<Response<DebugIndexResponse>, Status> {
        let md = request.metadata().clone();
        common::trace::set_parent_from_metadata(&md);
        let req = request.into_inner();
        let doc_id = req.doc_id;

        // 使用 spawn_blocking 包装
        let engine = Arc::clone(&self.engine);
        let indexed = tokio::task::spawn_blocking(move || engine.doc_exists(doc_id))
            .await
            .map_err(|e| Status::internal(format!("Task join error: {}", e)))?
            .map_err(Self::map_engine_error)?;

        Ok(Response::new(DebugIndexResponse {
            indexed,
            stored_fields_json: if indexed {
                format!(r#"{{"id":{}}}"#, doc_id)
            } else {
                String::new()
            },
        }))
    }
}
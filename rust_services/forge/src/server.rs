// rust_services/forge/src/server.rs

use std::sync::Arc;

use common::ctx::ContextData;
use common::forge::render_service_server::RenderService; // 生成的 Trait
use common::forge::{
    GetRenderMetaRequest, GetRenderMetaResponse, RenderPreviewRequest, RenderPreviewResponse,
    RenderRequest, RenderResponse,
};
use common::CodeError;
use tonic::{Request, Response, Status};

use crate::engine::MarkdownEngine;

pub struct GrpcServer {
    // Engine 是无状态且线程安全的，用 Arc 共享
    engine: Arc<MarkdownEngine>,
}

impl GrpcServer {
    pub fn new() -> Self {
        Self {
            engine: Arc::new(MarkdownEngine::new()),
        }
    }
}

#[tonic::async_trait]
impl RenderService for GrpcServer {
    // 1. 实时预览 (编辑器调用)
    async fn render_preview(
        &self,
        request: Request<RenderPreviewRequest>,
    ) -> Result<Response<RenderPreviewResponse>, Status> {
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
        let engine = self.engine.clone();

        let start = std::time::Instant::now();

        let (html, toc, _) = tokio::task::spawn_blocking(move || engine.render(&req.raw_markdown))
            .await
            .map_err(|e| CodeError::internal_with(e, "render_preview: join error").to_status())?
            .map_err(|e| CodeError::validation(e.to_string()).to_status())?;

        let took_ms = start.elapsed().as_secs_f32() * 1000.0;

        Ok(Response::new(RenderPreviewResponse {
            html_body: html,
            toc_json: toc,
            took_ms,
        }))
    }

    // 2. 元数据
    async fn get_render_meta(
        &self,
        request: Request<GetRenderMetaRequest>,
    ) -> Result<Response<GetRenderMetaResponse>, Status> {
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
        
        Ok(Response::new(GetRenderMetaResponse {
            version: "v3.2".into(),
            enabled_extensions: vec!["tables".into(), "tasklists".into(), "footnotes".into()],
        }))
    }

    // 3. 正式渲染 (Nexus 调用)
    async fn render(&self, request: Request<RenderRequest>) -> Result<Response<RenderResponse>, Status> {
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
        let engine = self.engine.clone();

        // [关键架构点] CPU 密集型任务扔到 blocking 线程池
        // 避免阻塞 gRPC 的 IO 循环
        let (html, toc, summary) = tokio::task::spawn_blocking(move || engine.render(&req.raw_markdown))
            .await
            .map_err(|e| CodeError::internal_with(e, "render: join error").to_status())?
            .map_err(|e| CodeError::validation(e.to_string()).to_status())?;

        Ok(Response::new(RenderResponse {
            html_body: html,
            toc_json: toc,
            summary,
        }))
    }
}
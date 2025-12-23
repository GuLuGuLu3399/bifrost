use anyhow::Result;
use hyper::service::{make_service_fn, service_fn};
use hyper::{Body, Request, Response, Server};
use metrics_exporter_prometheus::{PrometheusBuilder, PrometheusHandle};
use std::net::SocketAddr;
use tracing::{info, warn};

/// Initialize a Prometheus recorder and spawn an HTTP server to expose /metrics.
pub async fn init_prometheus(addr: &str) -> Result<()> {
    let builder = PrometheusBuilder::new();
    let handle: PrometheusHandle = builder.install_recorder()?;

    let addr: SocketAddr = addr.parse()?;
    let make_service = make_service_fn(move |_| {
        let handle = handle.clone();
        async move {
            Ok::<_, hyper::Error>(service_fn(move |req: Request<Body>| {
                let handle = handle.clone();
                async move {
                    if req.uri().path() == "/metrics" {
                        let body = handle.render();
                        Ok::<_, hyper::Error>(Response::new(Body::from(body)))
                    } else {
                        Ok::<_, hyper::Error>(Response::builder()
                            .status(404)
                            .body(Body::from("not found"))
                            .unwrap())
                    }
                }
            }))
        }
    });

    info!(%addr, "Prometheus metrics server starting");

    tokio::spawn(async move {
        if let Err(e) = Server::bind(&addr).serve(make_service).await {
            warn!(error = %e, "Prometheus server error");
        }
    });

    Ok(())
}

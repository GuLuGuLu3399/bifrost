use crate::logger::build_fmt_layer;

use opentelemetry::{global, KeyValue};
use opentelemetry_otlp::WithExportConfig;
use opentelemetry_sdk::trace::SdkTracerProvider;

/// 保持 provider 存活，避免 tracer 被提前 drop。
pub struct OtelGuard {
    pub provider: SdkTracerProvider,
}

impl Drop for OtelGuard {
    fn drop(&mut self) {
        // 尽力 flush；失败不 panic
        if let Err(e) = self.provider.shutdown() {
            tracing::warn!("Failed to shutdown tracer provider: {:?}", e);
        }
    }
}

fn init_tracer(service_name: &str, otlp_endpoint: &str) -> anyhow::Result<OtelGuard> {
    use opentelemetry_sdk::Resource;

    // 资源信息（service.name 等）。
    let resource = Resource::builder()
        .with_service_name(service_name.to_string())
        .with_attribute(KeyValue::new("service.name", service_name.to_string()))
        .build();

    // OTLP exporter (tonic gRPC)
    let exporter = opentelemetry_otlp::SpanExporter::builder()
        .with_tonic()
        .with_endpoint(otlp_endpoint)
        .build()?;

    let provider = SdkTracerProvider::builder()
        .with_resource(resource)
        .with_batch_exporter(exporter)
        .build();

    global::set_tracer_provider(provider.clone());

    Ok(OtelGuard { provider })
}

pub fn init_tracing_with_otel(
    service_name: &str,
    env: &str,
    json: bool,
    otlp_endpoint: &str,
) -> anyhow::Result<OtelGuard> {
    use opentelemetry::trace::TracerProvider;
    use tracing_subscriber::{layer::SubscriberExt, util::SubscriberInitExt, EnvFilter, Registry};

    let guard = init_tracer(service_name, otlp_endpoint)?;
    let tracer = guard.provider.tracer(service_name.to_string());
    let otel_layer = tracing_opentelemetry::layer().with_tracer(tracer);

    let filter = EnvFilter::try_from_default_env().unwrap_or_else(|_| EnvFilter::new("info"));

    // build_fmt_layer 内部已根据 json 开关选择 json/text layer
    // 注意：这里 S 要匹配最终 subscriber：Registry.with(filter).with(otel_layer)
    type Sub = tracing_subscriber::layer::Layered<
        tracing_opentelemetry::OpenTelemetryLayer<
            tracing_subscriber::layer::Layered<EnvFilter, Registry>,
            opentelemetry_sdk::trace::Tracer,
        >,
        tracing_subscriber::layer::Layered<EnvFilter, Registry>,
    >;
    let fmt_layer = build_fmt_layer::<Sub>(json);

    Registry::default()
        .with(filter)
        .with(otel_layer)
        .with(fmt_layer)
        .try_init()?;

    tracing::info!(
        service = service_name,
        env = env,
        otlp = otlp_endpoint,
        "tracing + otlp initialized"
    );
    Ok(guard)
}

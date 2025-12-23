use tracing_subscriber::{fmt, layer::SubscriberExt, util::SubscriberInitExt, EnvFilter};

pub fn build_fmt_layer<S>(
    json: bool,
) -> Box<dyn tracing_subscriber::Layer<S> + Send + Sync + 'static>
where
    S: tracing::Subscriber + for<'a> tracing_subscriber::registry::LookupSpan<'a>,
{
    let timer = fmt::time::UtcTime::rfc_3339();

    let layer = fmt::layer()
        .with_target(false)
        .with_ansi(!json)
        .with_file(false)
        .with_line_number(true)
        .with_timer(timer)
        .with_thread_ids(false)
        .with_level(true);

    if json {
        Box::new(layer.json())
    } else {
        Box::new(layer)
    }
}

static INIT: std::sync::OnceLock<()> = std::sync::OnceLock::new();

pub fn init_tracing(service_name: &str, env: &str, json: bool) -> anyhow::Result<()> {
    INIT.get_or_init(|| {
        let filter = EnvFilter::try_from_default_env().unwrap_or_else(|_| EnvFilter::new("info"));

        // 注意：fmt layer 的 S 必须匹配最终 subscriber（包含 filter）
        type Sub = tracing_subscriber::layer::Layered<EnvFilter, tracing_subscriber::Registry>;
        let fmt_layer = build_fmt_layer::<Sub>(json);

        tracing_subscriber::registry()
            .with(filter)
            .with(fmt_layer)
            .init();

        tracing::info!(service = service_name, env = env, "tracing subscriber initialized");
    });
    Ok(())
}
use opentelemetry::global;
use opentelemetry::propagation::Extractor;
use tonic::metadata::MetadataMap;
use tracing::Span;
use tracing_opentelemetry::OpenTelemetrySpanExt;

/// Minimal Extractor for gRPC MetadataMap to read W3C trace headers.
struct MetadataExtractor<'a> {
    md: &'a MetadataMap,
}

impl<'a> Extractor for MetadataExtractor<'a> {
    fn get(&self, key: &str) -> Option<&str> {
        self.md.get(key).and_then(|v| v.to_str().ok())
    }

    fn keys(&self) -> Vec<&str> {
        self.md.keys().map(|k| k.as_str()).collect()
    }
}

/// Set current span's parent from incoming gRPC metadata if present.
pub fn set_parent_from_metadata(md: &MetadataMap) {
    let extractor = MetadataExtractor { md };
    let cx = global::get_text_map_propagator(|prop| prop.extract(&extractor));
    let span = Span::current();
    span.set_parent(cx);
}

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
        // tonic 的 MetadataMap::keys() 返回 KeyRef，不同版本 API 不一致。
        // Extractor trait 要求返回 Vec<&str>，因此这里把 key 转成 owned String，
        // 再以 &'static str 的形式返回（用于一次性提取传播字段，数量很少）。
        self.md
            .keys()
            .filter_map(|k| {
                // KeyRef 转换为 &str
                let key_str = match k {
                    tonic::metadata::KeyRef::Ascii(name) => name.as_str(),
                    tonic::metadata::KeyRef::Binary(_name) => {
                        // Binary keys 通常不用于追踪头，跳过
                        return None;
                    }
                };
                let s = key_str.to_string();
                Some(Box::leak(s.into_boxed_str()) as &'static str)
            })
            .collect()
    }
}

/// Set current span's parent from incoming gRPC metadata if present.
pub fn set_parent_from_metadata(md: &MetadataMap) {
    let extractor = MetadataExtractor { md };
    let cx = global::get_text_map_propagator(|prop| prop.extract(&extractor));
    let span = Span::current();

    // set_parent 本身不会因为 metadata 缺失而失败；这里不需要 panic。
    let _ = span.set_parent(cx);
}

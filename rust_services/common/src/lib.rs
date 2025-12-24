//! Bifrost Rust common 基础库：
//! - Proto 类型聚合与重导出
//! - 通用错误模型/映射
//! - 日志与 tracing 初始化
//! - 跨服务上下文元数据辅助

pub mod api {
    pub mod common {
        pub mod v1 {
            tonic::include_proto!("bifrost.common.v1");
        }
    }

    pub mod content {
        pub mod v1 {
            tonic::include_proto!("bifrost.content.v1");

            pub mod forge {
                tonic::include_proto!("bifrost.content.v1.forge");
            }
            pub mod oracle {
                // oracle 的 package 是 bifrost.analysis.v1.oracle
                tonic::include_proto!("bifrost.analysis.v1.oracle");
            }
        }
    }

    pub mod search {
        pub mod v1 {
            tonic::include_proto!("bifrost.search.v1");
        }
    }
}

// Proto re-export（方便上层使用）
pub use api::common::v1 as common;
pub use api::content::v1 as models;
pub use api::content::v1::forge;
pub use api::content::v1::oracle;
pub use api::search::v1 as search;

pub mod config;
pub mod ctx;
pub mod error;
pub mod lifecycle;
pub mod logger;
pub mod metrics;
pub mod nats;
pub mod otel;
pub mod trace;

pub use ctx::ContextData;
pub use error::{CodeError, ErrorCode, ErrorResponse};

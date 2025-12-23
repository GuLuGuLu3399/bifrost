use serde::{Deserialize, Serialize};
use tonic::metadata::{MetadataMap, MetadataValue};

/// 跨服务上下文数据的 Header 键名定义。

pub const HDR_USER_ID: &str = "x-user-id";
pub const HDR_REQUEST_ID: &str = "x-request-id";
pub const HDR_AUTHORIZATION: &str = "authorization";
pub const HDR_LOCALE: &str = "x-locale";
pub const HDR_IS_ADMIN: &str = "x-is-admin";

/// 跨服务透传的上下文数据（与 Go 对齐）。
#[derive(Debug, Default, Clone, Serialize, Deserialize, PartialEq, Eq)]
pub struct ContextData {
    pub user_id: Option<i64>,
    pub request_id: Option<String>,
    pub token: Option<String>,
    pub locale: Option<String>,
    pub is_admin: bool,
}

impl ContextData {
    /// 从 gRPC Metadata 构建上下文。
    pub fn from_metadata(md: &MetadataMap) -> Self {
        let user_id = md
            .get(HDR_USER_ID)
            .and_then(|v| v.to_str().ok())
            .and_then(|s| s.parse::<i64>().ok());

        let request_id = md
            .get(HDR_REQUEST_ID)
            .and_then(|v| v.to_str().ok())
            .map(|s| s.to_string());
        let token = md
            .get(HDR_AUTHORIZATION)
            .and_then(|v| v.to_str().ok())
            .map(|s| s.to_string());
        let locale = md
            .get(HDR_LOCALE)
            .and_then(|v| v.to_str().ok())
            .map(|s| s.to_string());
        let is_admin = md
            .get(HDR_IS_ADMIN)
            .and_then(|v| v.to_str().ok())
            .and_then(|s| s.parse::<bool>().ok())
            .unwrap_or(false);

        ContextData {
            user_id,
            request_id,
            token,
            locale,
            is_admin,
        }
    }

    /// 将上下文写入 gRPC Metadata。
    pub fn inject_into(&self, md: &mut MetadataMap) {
        if let Some(uid) = self.user_id {
            if let Ok(v) = MetadataValue::try_from(uid.to_string()) {
                md.insert(HDR_USER_ID, v);
            }
        }
        if let Some(rid) = &self.request_id {
            if let Ok(v) = MetadataValue::try_from(rid.as_str()) {
                md.insert(HDR_REQUEST_ID, v);
            }
        }
        if let Some(token) = &self.token {
            if let Ok(v) = MetadataValue::try_from(token.as_str()) {
                md.insert(HDR_AUTHORIZATION, v);
            }
        }
        if let Some(locale) = &self.locale {
            if let Ok(v) = MetadataValue::try_from(locale.as_str()) {
                md.insert(HDR_LOCALE, v);
            }
        }
        if let Ok(v) = MetadataValue::try_from(self.is_admin.to_string()) {
            md.insert(HDR_IS_ADMIN, v);
        }
    }

    /// 便捷方法：对 tonic::Request 进行注入。
    pub fn inject_request<T>(&self, req: &mut tonic::Request<T>) {
        let md = req.metadata_mut();
        self.inject_into(md);
    }
}

use reqwest::{Client, Method, Response};
use serde::{Deserialize, Serialize};
use serde_json::{json, Value};
use std::sync::{Arc, RwLock};
use std::time::Duration;
use anyhow::{Context, Result, anyhow, bail};

/// HTTP Client for communicating with Gjallar Gateway
pub struct HttpClient {
    client: Client,
    base_url: String,
    token: Arc<RwLock<Option<String>>>,
}

// ---------------------------------------------------------
// 1. 修正 Request/Response 结构体以匹配 Proto 定义
// ---------------------------------------------------------

#[derive(Debug, Serialize, Deserialize)]
pub struct LoginRequest {
    pub identifier: String,
    pub password: String,
}

#[derive(Debug, Serialize, Deserialize)]
pub struct LoginResponse {
    pub access_token: String,
    pub refresh_token: String,
    pub expires_in: i64,
}

#[derive(Debug, Serialize, Deserialize)]
pub struct RegisterResponse {
    // 关键修正：Proto 中注册只返回 user_id，不直接返回 token
    pub user_id: i64,
}

// 上传凭证请求 (发给后端，所以需要 Serialize)
#[derive(Debug, Serialize)]
pub struct GetUploadTicketRequest {
    pub filename: String,
    pub usage: String,
}

// 上传凭证响应
#[derive(Debug, Deserialize, Serialize)] 
pub struct GetUploadTicketResponse {
    pub upload_url: String, 
    pub object_key: String, 
}

#[derive(Debug, Default, Serialize, Deserialize)]
pub struct ListPostsQuery {
    pub page_size: Option<i32>,
    pub page_token: Option<String>,
    pub category_id: Option<i64>,
    pub tag_id: Option<i64>,
    pub author_id: Option<i64>,
}

// ---------------------------------------------------------
// 2. Client 实现
// ---------------------------------------------------------

impl HttpClient {
    pub fn new(base_url: String) -> Self {
        let client = Client::builder()
            .connect_timeout(Duration::from_secs(5))
            .timeout(Duration::from_secs(30))
            .pool_max_idle_per_host(8)
            .build()
            .expect("failed to build HTTP client");

        Self {
            client,
            base_url,
            token: Arc::new(RwLock::new(None)),
        }
    }

    pub fn set_token(&self, token: String) {
        if let Ok(mut t) = self.token.write() {
            *t = Some(token);
        }
    }

    pub fn get_token(&self) -> Option<String> {
        self.token.read().ok()?.clone()
    }

    pub fn clear_token(&self) {
        if let Ok(mut t) = self.token.write() {
            *t = None;
        }
    }

    pub async fn login(&self, identifier: String, password: String) -> Result<LoginResponse> {
        let url = format!("{}/v1/auth/login", self.base_url);
        
        let response = self.client
            .post(&url)
            .json(&LoginRequest { identifier, password })
            .send()
            .await
            .context("Failed to send login request")?;

        if !response.status().is_success() {
            let status = response.status();
            let error_text = extract_error_message(response).await;
            bail!("Login failed: [{}] {}", status, error_text);
        }

        let raw_text = response
            .text()
            .await
            .context("Failed to read login response")?;

        let login_response = parse_login_response(&raw_text)?;

        self.set_token(login_response.access_token.clone());

        Ok(login_response)
    }

    // ---------------------------------------------------------
    // Register: 修正了响应解析
    // ---------------------------------------------------------
    pub async fn register(&self, username: String, email: String, password: String) -> Result<RegisterResponse> {
        let url = format!("{}/v1/auth/register", self.base_url);

        // 匿名结构体匹配 Proto: RegisterRequest { username, email, password, nickname }
        #[derive(Serialize)]
        struct RegisterReq<'a> { 
            username: &'a str, 
            email: &'a str, 
            password: &'a str,
            nickname: &'a str // 默认 nickname = username
        }

        let response = self.client
            .post(&url)
            .json(&RegisterReq { 
                username: &username, 
                email: &email, 
                password: &password,
                nickname: &username 
            })
            .send()
            .await
            .context("Failed to send register request")?;

        if !response.status().is_success() {
            let status = response.status();
            let error_text = extract_error_message(response).await;
            bail!("Register failed: [{}] {}", status, error_text);
        }

        let register_response: RegisterResponse = response
            .json()
            .await
            .context("Failed to parse register response")?;

        Ok(register_response)
    }

    // ---------------------------------------------------------
    // Upload: 彻底重构为 预签名 URL 模式 (GetTicket -> PUT)
    // ---------------------------------------------------------
    pub async fn upload_image(&self, filename: String, file_data: Vec<u8>) -> Result<GetUploadTicketResponse> {
        // Step 1: 获取上传凭证 (Ticket)
        let ticket_url = format!("{}/v1/storage/upload_ticket", self.base_url);
        let token = self.get_token().context("Unauthorized")?;

        let ticket_res = self.client
            .post(&ticket_url)
            .header("Authorization", format!("Bearer {}", token))
            .json(&GetUploadTicketRequest {
                filename: filename.clone(),
                usage: "post_image".to_string(), // 默认为文章插图
            })
            .send()
            .await
            .context("Failed to request upload ticket")?;

        if !ticket_res.status().is_success() {
            let status = ticket_res.status();
            let err = extract_error_message(ticket_res).await;
            if status.as_u16() == 404 {
                bail!("Failed to get upload ticket: [404] gateway route not available yet (/v1/storage/upload_ticket)");
            }
            bail!("Failed to get upload ticket: [{}] {}", status, err);
        }

        let ticket: GetUploadTicketResponse = ticket_res.json().await?;

        // Step 2: 使用凭证中的 URL 直传文件 (PUT)
        // 注意：这里直接上传二进制数据，不需要 Multipart
        let upload_res = self.client
            .put(&ticket.upload_url)
            .body(file_data)
            .header("Content-Type", "image/webp") // 假设我们总是上传 WebP
            .send()
            .await
            .context("Failed to upload file with presigned URL")?;

        if !upload_res.status().is_success() {
            let status = upload_res.status();
            let err = extract_error_message(upload_res).await;
            bail!("Failed to upload file to storage: [{}] {}", status, err);
        }

        // 返回包含 object_key 的信息，供前端保存到文章内容中
        Ok(ticket)
    }

    pub async fn get_profile(&self) -> Result<Value> {
        let token = self.get_token().context("Unauthorized")?;
        let url = format!("{}/v1/users/profile", self.base_url);

        let response = self.client
            .get(&url)
            .header("Authorization", format!("Bearer {}", token))
            .send()
            .await
            .context("Failed to send profile request")?;

        if !response.status().is_success() {
            let status = response.status();
            let err = extract_error_message(response).await;
            bail!("Failed to fetch profile: [{}] {}", status, err);
        }

        response
            .json::<Value>()
            .await
            .context("Failed to parse profile response")
    }

    pub async fn list_posts(&self, query: ListPostsQuery) -> Result<Value> {
        let url = format!("{}/v1/posts", self.base_url);

        let mut params: Vec<(String, String)> = Vec::new();
        if let Some(page_size) = query.page_size {
            if page_size > 0 {
                params.push(("page.page_size".to_string(), page_size.to_string()));
            }
        }
        if let Some(page_token) = query.page_token {
            let token = page_token.trim();
            if !token.is_empty() {
                params.push(("page.page_token".to_string(), token.to_string()));
            }
        }
        if let Some(category_id) = query.category_id {
            if category_id > 0 {
                params.push(("category_id".to_string(), category_id.to_string()));
            }
        }
        if let Some(tag_id) = query.tag_id {
            if tag_id > 0 {
                params.push(("tag_id".to_string(), tag_id.to_string()));
            }
        }
        if let Some(author_id) = query.author_id {
            if author_id > 0 {
                params.push(("author_id".to_string(), author_id.to_string()));
            }
        }

        let response = self.client
            .get(&url)
            .query(&params)
            .send()
            .await
            .context("Failed to send posts list request")?;

        if !response.status().is_success() {
            let status = response.status();
            let err = extract_error_message(response).await;
            bail!("Failed to list posts: [{}] {}", status, err);
        }

        response
            .json::<Value>()
            .await
            .context("Failed to parse posts list response")
    }

    pub async fn gateway_request(
        &self,
        method: &str,
        path: &str,
        query: Option<Value>,
        body: Option<Value>,
        auth_required: bool,
    ) -> Result<Value> {
        let method_upper = method.trim().to_uppercase();
        let method = Method::from_bytes(method_upper.as_bytes())
            .with_context(|| format!("Invalid HTTP method: {}", method))?;

        let path = path.trim();
        let url = if path.starts_with("http://") || path.starts_with("https://") {
            path.to_string()
        } else {
            format!("{}/{}", self.base_url.trim_end_matches('/'), path.trim_start_matches('/'))
        };

        let mut request_builder = self.client.request(method, &url);

        if auth_required {
            let token = self.get_token().context("Unauthorized")?;
            request_builder = request_builder.header("Authorization", format!("Bearer {}", token));
        }

        if let Some(query_val) = query {
            let query_pairs = normalize_query_pairs(query_val)?;
            if !query_pairs.is_empty() {
                request_builder = request_builder.query(&query_pairs);
            }
        }

        if let Some(body_val) = body {
            request_builder = request_builder.json(&body_val);
        }

        let response = request_builder
            .send()
            .await
            .context("Failed to send gateway request")?;

        if !response.status().is_success() {
            let status = response.status();
            let err = extract_error_message(response).await;
            bail!("Gateway request failed: [{}] {}", status, err);
        }

        let raw = response
            .text()
            .await
            .context("Failed to read gateway response")?;

        if raw.trim().is_empty() {
            return Ok(json!({ "success": true }));
        }

        if let Ok(parsed) = serde_json::from_str::<Value>(&raw) {
            return Ok(parsed);
        }

        Ok(json!({ "raw": raw }))
    }
}

fn normalize_query_pairs(query: Value) -> Result<Vec<(String, String)>> {
    let mut pairs = Vec::new();
    let obj = query
        .as_object()
        .ok_or_else(|| anyhow!("query must be a JSON object"))?;

    for (k, v) in obj {
        if v.is_null() {
            continue;
        }

        match v {
            Value::String(s) => {
                let s = s.trim();
                if s.is_empty() {
                    continue;
                }
                pairs.push((k.clone(), s.to_string()));
            }
            Value::Number(n) => pairs.push((k.clone(), n.to_string())),
            Value::Bool(b) => pairs.push((k.clone(), b.to_string())),
            Value::Array(arr) => {
                for item in arr {
                    if let Some(text) = scalar_to_text(item) {
                        pairs.push((k.clone(), text));
                    }
                }
            }
            _ => {}
        }
    }

    Ok(pairs)
}

fn scalar_to_text(v: &Value) -> Option<String> {
    match v {
        Value::String(s) => {
            let s = s.trim();
            if s.is_empty() {
                None
            } else {
                Some(s.to_string())
            }
        }
        Value::Number(n) => Some(n.to_string()),
        Value::Bool(b) => Some(b.to_string()),
        _ => None,
    }
}

#[derive(Debug, Deserialize)]
struct ApiErrorBody {
    message: Option<String>,
    error: Option<String>,
}

async fn extract_error_message(resp: Response) -> String {
    let raw = resp.text().await.unwrap_or_default();
    if raw.trim().is_empty() {
        return "empty response body".to_string();
    }

    if let Ok(parsed) = serde_json::from_str::<ApiErrorBody>(&raw) {
        if let Some(msg) = parsed.message {
            if !msg.trim().is_empty() {
                return msg;
            }
        }
        if let Some(err) = parsed.error {
            if !err.trim().is_empty() {
                return err;
            }
        }
    }

    raw
}

fn parse_login_response(raw_text: &str) -> Result<LoginResponse> {
    // 优先严格匹配后端当前返回格式：
    // {"accessToken":"...","refreshToken":"...","expiresIn":"86400"}
    if let Ok(strict) = serde_json::from_str::<BackendLoginResponse>(raw_text) {
        return Ok(LoginResponse {
            access_token: strict.access_token,
            refresh_token: strict.refresh_token,
            expires_in: strict.expires_in,
        });
    }

    let parsed: Value = serde_json::from_str(raw_text)
        .with_context(|| format!("Failed to parse login response body as JSON: {}", raw_text))?;

    let access_token = pick_string(&parsed, &["access_token", "accessToken", "token"])
        .or_else(|| pick_string_nested(&parsed, &["data", "result", "payload"], &["access_token", "accessToken", "token"]))
        .filter(|v| !v.trim().is_empty())
        .context("Login response missing access token (expected field: access_token/accessToken/token, root or data/result/payload)")?;

    let refresh_token = pick_string(&parsed, &["refresh_token", "refreshToken"])
        .or_else(|| pick_string_nested(&parsed, &["data", "result", "payload"], &["refresh_token", "refreshToken"]))
        .unwrap_or_default();

    let expires_in = pick_i64(&parsed, &["expires_in", "expiresIn"])
        .or_else(|| pick_i64_nested(&parsed, &["data", "result", "payload"], &["expires_in", "expiresIn"]))
        .unwrap_or_default();

    Ok(LoginResponse {
        access_token,
        refresh_token,
        expires_in,
    })
}

fn pick_string(v: &Value, keys: &[&str]) -> Option<String> {
    keys.iter()
        .find_map(|k| v.get(*k).and_then(Value::as_str).map(ToString::to_string))
}

fn pick_i64(v: &Value, keys: &[&str]) -> Option<i64> {
    keys.iter().find_map(|k| {
        let val = v.get(*k)?;
        if let Some(n) = val.as_i64() {
            return Some(n);
        }
        if let Some(s) = val.as_str() {
            return s.parse::<i64>().ok();
        }
        None
    })
}

fn pick_string_nested(v: &Value, containers: &[&str], keys: &[&str]) -> Option<String> {
    containers
        .iter()
        .find_map(|container| v.get(*container).and_then(|inner| pick_string(inner, keys)))
}

fn pick_i64_nested(v: &Value, containers: &[&str], keys: &[&str]) -> Option<i64> {
    containers
        .iter()
        .find_map(|container| v.get(*container).and_then(|inner| pick_i64(inner, keys)))
}

#[derive(Debug, Deserialize)]
#[serde(rename_all = "camelCase")]
struct BackendLoginResponse {
    access_token: String,
    refresh_token: String,
    #[serde(deserialize_with = "de_i64_from_str_or_num")]
    expires_in: i64,
}

fn de_i64_from_str_or_num<'de, D>(deserializer: D) -> std::result::Result<i64, D::Error>
where
    D: serde::Deserializer<'de>,
{
    #[derive(Deserialize)]
    #[serde(untagged)]
    enum NumOrString {
        Num(i64),
        Str(String),
    }

    match NumOrString::deserialize(deserializer)? {
        NumOrString::Num(v) => Ok(v),
        NumOrString::Str(s) => s
            .parse::<i64>()
            .map_err(serde::de::Error::custom),
    }
}
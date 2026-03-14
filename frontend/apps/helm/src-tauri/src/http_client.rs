use reqwest::{Client, StatusCode};
use serde::{Deserialize, Serialize};
use std::sync::{Arc, RwLock};
use anyhow::{Context, Result, bail};

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
    // 关键修正：后端 Proto 定义字段为 "identifier"
    // 使用 rename 保持 Rust 代码里的字段名习惯，同时满足后端 JSON 要求
    #[serde(rename = "identifier")]
    pub username: String,
    pub password: String,
}

#[derive(Debug, Serialize, Deserialize)]
pub struct LoginResponse {
    // 关键修正：后端返回的是 "access_token"
    #[serde(rename = "access_token")]
    pub token: String,
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

// ---------------------------------------------------------
// 2. Client 实现
// ---------------------------------------------------------

impl HttpClient {
    pub fn new(base_url: String) -> Self {
        Self {
            client: Client::new(),
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

    // ---------------------------------------------------------
    // Login: 修正了字段名匹配问题
    // ---------------------------------------------------------
    pub async fn login(&self, username: String, password: String) -> Result<LoginResponse> {
        let url = format!("{}/v1/auth/login", self.base_url);
        
        let response = self.client
            .post(&url)
            .json(&LoginRequest { username, password })
            .send()
            .await
            .context("Failed to send login request")?;

        if !response.status().is_success() {
            let status = response.status();
            let error_text = response.text().await.unwrap_or_default();
            bail!("Login failed: [{}] {}", status, error_text);
        }

        let login_response: LoginResponse = response
            .json()
            .await
            .context("Failed to parse login response")?;

        // 自动保存 Token
        self.set_token(login_response.token.clone());

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
            let error_text = response.text().await.unwrap_or_default();
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
            .await?;

        if !ticket_res.status().is_success() {
            let err = ticket_res.text().await.unwrap_or_default();
            bail!("Failed to get upload ticket: {}", err);
        }

        let ticket: GetUploadTicketResponse = ticket_res.json().await?;

        // Step 2: 使用凭证中的 URL 直传文件 (PUT)
        // 注意：这里直接上传二进制数据，不需要 Multipart
        let upload_res = self.client
            .put(&ticket.upload_url)
            .body(file_data)
            .header("Content-Type", "image/webp") // 假设我们总是上传 WebP
            .send()
            .await?;

        if !upload_res.status().is_success() {
            bail!("Failed to upload file to storage: Status {}", upload_res.status());
        }

        // 返回包含 object_key 的信息，供前端保存到文章内容中
        Ok(ticket)
    }
}
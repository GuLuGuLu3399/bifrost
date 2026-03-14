mod http_client;
mod image_processor;

// 1. 修改引入：使用新的响应结构体 GetUploadTicketResponse
use http_client::{HttpClient, LoginResponse, GetUploadTicketResponse};
use serde::Deserialize;
use serde_json::Value;
use image_processor::ImageProcessor;
use tauri::State;

/// 持有 HTTP 客户端的应用状态
struct AppState {
    http_client: HttpClient,
}

impl AppState {
    fn new() -> Self {
        // 默认使用本地网关地址
        let base_url = std::env::var("GJALLAR_URL")
            .unwrap_or_else(|_| "http://localhost:8080".to_string());
        
        Self {
            http_client: HttpClient::new(base_url),
        }
    }
}

/// 登录命令
#[tauri::command]
async fn login_cmd(
    identifier: String,
    password: String,
    state: State<'_, AppState>,
) -> Result<LoginResponse, String> {
    let client = &state.http_client;

    client.login(identifier, password)
        .await
        .map_err(|e| format!("Login failed: {}", e))
}

/// 图片上传命令
#[tauri::command]
async fn upload_image_cmd(
    file_path: String,
    state: State<'_, AppState>,
) -> Result<GetUploadTicketResponse, String> { // 2. 修改返回类型
    // 第一步：将图片处理为 WebP
    let path_for_processing = file_path.clone();
    
    // 图片处理较重，放在 blocking 线程池中执行
    let webp_data = tokio::task::spawn_blocking(move || {
        ImageProcessor::process_file_to_webp(&path_for_processing)
    })
    .await
    .map_err(|e| format!("Task failed: {}", e))?
    .map_err(|e| format!("Image processing failed: {}", e))?;

    // 第二步：生成目标文件名（例如 myimage.png -> myimage.webp）
    let file_name = ImageProcessor::generate_webp_filename(&file_path);

    // 第三步：按 Ticket -> PUT 流程上传
    let client = &state.http_client;

    // client.upload_image 现在会执行两步：
    // 1. POST /upload_ticket 拿到 URL
    // 2. PUT data 到该 URL
    client.upload_image(file_name, webp_data)
        .await
        .map_err(|e| format!("Upload failed: {}", e))
}

/// 前端注册请求 DTO
#[derive(Debug, Deserialize)]
struct RegisterDto {
    username: String,
    email: String,
    password: String,
}

/// 注册命令
#[tauri::command]
async fn register_cmd(dto: RegisterDto, state: State<'_, AppState>) -> Result<String, String> {
    let client = &state.http_client;
    client
        .register(dto.username, dto.email, dto.password)
        .await
        .map(|_| "Registration Successful".to_string())
        .map_err(|e| format!("Registration failed: {}", e))
}

/// 检查认证状态
#[tauri::command]
fn is_authenticated(state: State<'_, AppState>) -> Result<bool, String> {
    let client = &state.http_client;
    // 简单的判空检查
    Ok(matches!(client.get_token(), Some(t) if !t.is_empty()))
}

/// 退出登录命令
#[tauri::command]
fn logout_cmd(state: State<'_, AppState>) -> Result<(), String> {
    let client = &state.http_client;
    client.clear_token();
    Ok(())
}

#[derive(Debug, Deserialize)]
#[serde(rename_all = "camelCase")]
struct ListPostsQueryDto {
    page_size: Option<i32>,
    page_token: Option<String>,
    category_id: Option<i64>,
    tag_id: Option<i64>,
    author_id: Option<i64>,
}

#[tauri::command]
async fn fetch_profile_cmd(state: State<'_, AppState>) -> Result<Value, String> {
    let client = &state.http_client;
    client
        .get_profile()
        .await
        .map_err(|e| format!("Fetch profile failed: {}", e))
}

#[tauri::command]
async fn list_posts_cmd(
    query: Option<ListPostsQueryDto>,
    state: State<'_, AppState>,
) -> Result<Value, String> {
    let client = &state.http_client;
    let query = query.unwrap_or(ListPostsQueryDto {
        page_size: None,
        page_token: None,
        category_id: None,
        tag_id: None,
        author_id: None,
    });

    client
        .list_posts(http_client::ListPostsQuery {
            page_size: query.page_size,
            page_token: query.page_token,
            category_id: query.category_id,
            tag_id: query.tag_id,
            author_id: query.author_id,
        })
        .await
        .map_err(|e| format!("List posts failed: {}", e))
}

#[tauri::command]
async fn gateway_request_cmd(
    method: String,
    path: String,
    query: Option<Value>,
    body: Option<Value>,
    auth_required: Option<bool>,
    state: State<'_, AppState>,
) -> Result<Value, String> {
    let client = &state.http_client;
    client
        .gateway_request(
            &method,
            &path,
            query,
            body,
            auth_required.unwrap_or(true),
        )
        .await
        .map_err(|e| format!("Gateway request failed: {}", e))
}

#[cfg_attr(mobile, tauri::mobile_entry_point)]
pub fn run() {
    tauri::Builder::default()
        .plugin(tauri_plugin_opener::init())
        .plugin(tauri_plugin_dialog::init())
        .manage(AppState::new())
        .invoke_handler(tauri::generate_handler![
            login_cmd,
            register_cmd,
            upload_image_cmd,
            is_authenticated,
            logout_cmd,
            fetch_profile_cmd,
            list_posts_cmd,
            gateway_request_cmd,
        ])
        .run(tauri::generate_context!())
        .expect("error while running tauri application");
}
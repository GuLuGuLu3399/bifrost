mod http_client;
mod image_processor;

// 1. 修改引入：使用新的响应结构体 GetUploadTicketResponse
use http_client::{HttpClient, LoginResponse, GetUploadTicketResponse};
use serde::Deserialize;
use image_processor::ImageProcessor;
use tauri::State;

/// Application state holding the HTTP client
struct AppState {
    http_client: HttpClient,
}

impl AppState {
    fn new() -> Self {
        // Default to localhost:8080
        let base_url = std::env::var("GJALLAR_URL")
            .unwrap_or_else(|_| "http://localhost:8080".to_string());
        
        Self {
            http_client: HttpClient::new(base_url),
        }
    }
}

/// Login command
#[tauri::command]
async fn login_cmd(
    username: String,
    password: String,
    state: State<'_, AppState>,
) -> Result<LoginResponse, String> {
    let client = &state.http_client;

    // http_client.rs 内部已经处理了 "username" -> "identifier" 的字段映射
    client.login(username, password)
        .await
        .map_err(|e| format!("Login failed: {}", e))
}

/// Upload image command
#[tauri::command]
async fn upload_image_cmd(
    file_path: String,
    state: State<'_, AppState>,
) -> Result<GetUploadTicketResponse, String> { // 2. 修改返回类型
    // Step 1: Process image to WebP
    let path_for_processing = file_path.clone();
    
    // 图片处理较重，放在 blocking 线程池中执行
    let webp_data = tokio::task::spawn_blocking(move || {
        ImageProcessor::process_file_to_webp(&path_for_processing)
    })
    .await
    .map_err(|e| format!("Task failed: {}", e))?
    .map_err(|e| format!("Image processing failed: {}", e))?;

    // Step 2: Generate filename (e.g., "myimage.png" -> "myimage.webp")
    let file_name = ImageProcessor::generate_webp_filename(&file_path);

    // Step 3: Upload using the new Ticket -> PUT flow
    let client = &state.http_client;

    // client.upload_image 现在会执行两步：
    // 1. POST /upload_ticket 拿到 URL
    // 2. PUT data 到该 URL
    client.upload_image(file_name, webp_data)
        .await
        .map_err(|e| format!("Upload failed: {}", e))
}

/// Registration DTO from frontend
#[derive(Debug, Deserialize)]
struct RegisterDto {
    username: String,
    email: String,
    password: String,
}

/// Register command
#[tauri::command]
async fn register_cmd(dto: RegisterDto, state: State<'_, AppState>) -> Result<String, String> {
    let client = &state.http_client;
    client
        .register(dto.username, dto.email, dto.password)
        .await
        .map(|_| "Registration Successful".to_string())
        .map_err(|e| format!("Registration failed: {}", e))
}

/// Check authentication status
#[tauri::command]
fn is_authenticated(state: State<'_, AppState>) -> Result<bool, String> {
    let client = &state.http_client;
    // 简单的判空检查
    Ok(matches!(client.get_token(), Some(t) if !t.is_empty()))
}

/// Logout command
#[tauri::command]
fn logout_cmd(state: State<'_, AppState>) -> Result<(), String> {
    let client = &state.http_client;
    // 传入空字符串或特定逻辑来清除 Token
    // 这里假设 set_token 只是简单覆盖
    // 更好的做法是在 HttpClient 加一个 clear_token 方法
    client.set_token("".to_string()); 
    Ok(())
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
        ])
        .run(tauri::generate_context!())
        .expect("error while running tauri application");
}
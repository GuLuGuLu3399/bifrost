# gRPC Gateway 集成测试脚本 (Windows PowerShell)
# 说明：本脚本假设已启动 beacon 和 nexus 微服务，以及 gjallar 网关

$ErrorActionPreference = "Continue"

$BASE_URL = "http://localhost:8080"

Write-Host "🧪 gRPC Gateway 集成测试" -ForegroundColor Cyan
Write-Host "==================================" -ForegroundColor Cyan
Write-Host ""

# 测试函数
function Test-Endpoint {
    param(
        [string]$Method,
        [string]$Endpoint,
        [string]$Body,
        [int]$ExpectedCode
    )
    
    Write-Host -NoNewline "📌 $Method $Endpoint ... "
    
    try {
        $params = @{
            Uri             = "$BASE_URL$Endpoint"
            Method          = $Method
            Headers         = @{"Content-Type" = "application/json" }
            UseBasicParsing = $true
        }
        
        if ($Body) {
            $params["Body"] = $Body
        }
        
        $response = Invoke-WebRequest @params
        $httpCode = $response.StatusCode
        $responseBody = $response.Content
        
        if ($httpCode -eq $ExpectedCode) {
            Write-Host "✅ HTTP $httpCode" -ForegroundColor Green
            Write-Host "   响应: $(($responseBody | ConvertFrom-Json | ConvertTo-Json -Compress)[0..50] -join '')" -ForegroundColor Gray
            Write-Host ""
        }
        else {
            Write-Host "❌ HTTP $httpCode (期望 $ExpectedCode)" -ForegroundColor Red
            Write-Host "   响应: $responseBody" -ForegroundColor Red
            return $false
        }
    }
    catch {
        $httpCode = $_.Exception.Response.StatusCode.Value__
        Write-Host "❌ HTTP $httpCode (期望 $ExpectedCode)" -ForegroundColor Red
        Write-Host "   错误: $($_.Exception.Message)" -ForegroundColor Red
        return $false
    }
    
    return $true
}

# 检查网关连接
Write-Host "📡 检查网关连接..." -ForegroundColor Yellow
try {
    $test = Invoke-WebRequest -Uri "$BASE_URL/v1/posts" -UseBasicParsing -ErrorAction Stop
    Write-Host "✅ 网关连接正常" -ForegroundColor Green
}
catch {
    Write-Host "⚠️  网关未响应，请确保已启动服务：" -ForegroundColor Yellow
    Write-Host "   1. 启动 Beacon 服务:  cd go_services && go run ./cmd/beacon/main.go" -ForegroundColor Gray
    Write-Host "   2. 启动 Nexus 服务:   cd go_services && go run ./cmd/nexus/main.go" -ForegroundColor Gray
    Write-Host "   3. 启动 Gjallar 网关: cd go_services && go run ./cmd/gjallar/main.go" -ForegroundColor Gray
    exit 1
}
Write-Host ""

# Beacon 服务测试 (读服务)
Write-Host "📖 Beacon 服务测试（读服务）" -ForegroundColor Cyan
Write-Host "---" -ForegroundColor Gray

Test-Endpoint -Method "GET" -Endpoint "/v1/posts" -ExpectedCode 200
Test-Endpoint -Method "GET" -Endpoint "/v1/posts?page_size=10" -ExpectedCode 200
Test-Endpoint -Method "GET" -Endpoint "/v1/posts/1" -ExpectedCode 200

Test-Endpoint -Method "GET" -Endpoint "/v1/categories" -ExpectedCode 200
Test-Endpoint -Method "GET" -Endpoint "/v1/tags" -ExpectedCode 200

Write-Host ""

# Nexus 服务测试 (写服务)
Write-Host "✍️  Nexus 服务测试（写服务）" -ForegroundColor Cyan
Write-Host "---" -ForegroundColor Gray

# 注册
$registerBody = @{
    username = "testuser"
    email    = "test@example.com"
    password = "password123"
    nickname = "Test User"
} | ConvertTo-Json

Test-Endpoint -Method "POST" -Endpoint "/v1/auth/register" -Body $registerBody -ExpectedCode 200

# 登录
$loginBody = @{
    identifier = "testuser"
    password   = "password123"
} | ConvertTo-Json

Test-Endpoint -Method "POST" -Endpoint "/v1/auth/login" -Body $loginBody -ExpectedCode 200

# 创建文章
$postBody = @{
    title           = "Test Post"
    slug            = "test-post"
    raw_markdown    = "# Hello World"
    category_id     = 1
    tag_names       = @("test", "demo")
    status          = 1
    cover_image_key = "cover.jpg"
} | ConvertTo-Json

Test-Endpoint -Method "POST" -Endpoint "/v1/posts" -Body $postBody -ExpectedCode 200

Write-Host ""
Write-Host "==================================" -ForegroundColor Cyan
Write-Host "✅ 测试完成！" -ForegroundColor Green
Write-Host ""
Write-Host "💡 验证网关正常工作：" -ForegroundColor Cyan
Write-Host "   - gRPC Gateway 自动将 HTTP 请求转换为 gRPC 调用" -ForegroundColor Gray
Write-Host "   - 响应自动从 Protobuf 转换为 JSON" -ForegroundColor Gray
Write-Host "   - URL 路径参数自动提取和映射" -ForegroundColor Gray
Write-Host ""

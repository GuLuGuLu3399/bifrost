<#
.SYNOPSIS
    Bifrost CMS v3.2 - Monorepo Master Control Script
    (修复了 missing google/api/annotations.proto 问题)
#>

param(
    [Parameter(Position = 0)]
    [string]$Command = "help",

    [Parameter(Position = 1)]
    [string]$Name = "",

    [string]$ComposeFile = "deploy/docker-compose.yml"
)

# --- 配置变量 ---
$GoDir = "go_services"
$RustDir = "rust_services"
$AppsDir = "apps"
$ProtoDir = "api"
$BinDir = "bin"
$ThirdPartyDir = "third_party" # 存放 Google 协议文件的目录

$ErrorActionPreference = "Continue"

# --- 辅助函数 ---

function Write-Header ($Message) { Write-Host -ForegroundColor Cyan "`n>> $Message" }
function Write-Success ($Message) { Write-Host -ForegroundColor Green $Message }
function Write-ErrorMsg ($Message) { Write-Host -ForegroundColor Red "Error: $Message" }
function Test-Command ($Cmd) { return (Get-Command $Cmd -ErrorAction SilentlyContinue) }

function Exec {
    param([scriptblock]$ScriptBlock)
    & $ScriptBlock
    if ($LASTEXITCODE -ne 0) {
        Write-ErrorMsg "Command failed with exit code $LASTEXITCODE"
        exit $LASTEXITCODE
    }
}

# 自动下载 Google API 依赖
function Ensure-GoogleApis {
    $TargetDir = "$ThirdPartyDir/google/api"
    if (-not (Test-Path $TargetDir)) {
        New-Item -ItemType Directory -Force -Path $TargetDir | Out-Null
    }

    $Files = @("annotations.proto", "http.proto")
    $BaseUrl = "https://raw.githubusercontent.com/googleapis/googleapis/master/google/api"

    foreach ($File in $Files) {
        $LocalPath = "$TargetDir/$File"
        if (-not (Test-Path $LocalPath)) {
            Write-Host "Downloading dependency: $File ..."
            try {
                Invoke-WebRequest -Uri "$BaseUrl/$File" -OutFile $LocalPath
            }
            catch {
                Write-ErrorMsg "Failed to download $File. Please check your network."
                exit 1
            }
        }
    }
}

# --- 任务逻辑 ---

switch ($Command.ToLower()) {

    "help" {
        Write-Host ""
        Write-Host "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" -ForegroundColor Cyan
        Write-Host "  Bifrost CMS v3.2 - Master Control Script" -ForegroundColor Cyan
        Write-Host "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" -ForegroundColor Cyan
        Write-Host ""
        Write-Host "Proto & Gateway Commands:" -ForegroundColor Green
        Write-Host "  proto-gen     | 生成 Proto 代码 (protobuf + gRPC + gRPC Gateway)"
        Write-Host "  proto-lint    | 检查 Proto 文件规范"
        Write-Host "  gateway-test  | 测试 gRPC Gateway 网关"
        Write-Host ""
        Write-Host "Service Commands:" -ForegroundColor Green
        Write-Host "  run-nexus     | 启动 Nexus 微服务"
        Write-Host "  run-beacon    | 启动 Beacon 微服务"
        Write-Host "  run-gjallar   | 启动 Gjallar 网关"
        Write-Host ""
        Write-Host "Build Commands:" -ForegroundColor Green
        Write-Host "  build-go      | 构建所有 Go 服务"
        Write-Host "  build-rust    | 构建所有 Rust 服务"
        Write-Host ""
        Write-Host "Infrastructure Commands:" -ForegroundColor Green
        Write-Host "  up            | 启动 Docker 基础设施"
        Write-Host "  down          | 停止 Docker 基础设施"
        Write-Host "  logs          | 查看服务日志"
        Write-Host "  ps            | 列出运行中的容器"
        Write-Host ""
        Write-Host "Database Commands:" -ForegroundColor Green
        Write-Host "  migrate-up    | 执行数据库迁移"
        Write-Host "  migrate-new   | 创建新迁移 (需要 -Name 参数)"
        Write-Host "  db-connect    | 连接到数据库"
        Write-Host ""
        Write-Host "Utility Commands:" -ForegroundColor Green
        Write-Host "  init          | 初始化项目"
        Write-Host "  format        | 格式化代码"
        Write-Host "  clean         | 清理构建产物"
        Write-Host "  help          | 显示此帮助信息"
        Write-Host ""
        Write-Host "Examples:" -ForegroundColor Yellow
        Write-Host "  .\manage.ps1 proto-gen"
        Write-Host "  .\manage.ps1 run-nexus"
        Write-Host "  .\manage.ps1 build-go"
        Write-Host "  .\manage.ps1 migrate-new -Name create_users_table"
        Write-Host ""
    }

    "init" {
        Write-Header "[Init] Setting up Go modules..."
        Push-Location $GoDir; Exec { go mod tidy }; Pop-Location
        Write-Header "[Init] Setting up Rust workspace..."
        Push-Location $RustDir; Exec { cargo check }; Pop-Location
        Write-Success "Done."
    }

    # ==============================================================================
    # Proto generation (Protobuf + gRPC + gRPC Gateway)
    # ==============================================================================
    "proto-gen" {
        Write-Header "[Proto] Generate code (Protobuf + gRPC + gRPC Gateway)"
        
        # Check if buf is installed
        if (Test-Command buf) {
            Write-Host "OK buf installed" -ForegroundColor Green
            Write-Header "[Proto] Running buf generate..."
            Exec { buf generate }
            Write-Success "OK Proto code generated!"
            Write-Host ""
            Write-Host "Generated files:" -ForegroundColor Cyan
            Write-Host "  * *.pb.go          (Protobuf data structures)" -ForegroundColor Gray
            Write-Host "  * *_grpc.pb.go     (gRPC service definitions)" -ForegroundColor Gray
            Write-Host "  * *.pb.gw.go       (gRPC Gateway handlers)" -ForegroundColor Gray
            Write-Host ""
        }
        else {
            Write-ErrorMsg "buf not installed!"
            Write-Host "Install buf: go install github.com/bufbuild/buf/cmd/buf@latest" -ForegroundColor Yellow
            exit 1
        }
    }

    "proto-lint" {
        Write-Header "[Proto] Check Proto file standards..."
        
        if (Test-Command buf) {
            Exec { buf lint }
            Write-Success "OK Proto files passed checks!"
        }
        else {
            Write-ErrorMsg "buf not installed!"
            Write-Host "Install buf: go install github.com/bufbuild/buf/cmd/buf@latest" -ForegroundColor Yellow
            exit 1
        }
    }

    # Backward compatibility with old proto command
    "proto" {
        Write-Header "[Proto] 'proto' command is deprecated, use 'proto-gen' instead"
        & $PSCommandPath "proto-gen"
    }

    # ==============================================================================
    # gRPC Gateway Tests
    # ==============================================================================
    "gateway-test" {
        Write-Header "[Gateway] Testing gRPC Gateway..."
        
        # Check if gateway is running
        try {
            $response = Invoke-WebRequest -Uri "http://localhost:8080/v1/posts" -UseBasicParsing -ErrorAction Stop -TimeoutSec 2
            Write-Success "OK Gateway responded (HTTP $($response.StatusCode))"
        }
        catch {
            Write-ErrorMsg "Gateway did not respond. Make sure services are running:"
            Write-Host ""
            Write-Host "Startup steps:" -ForegroundColor Yellow
            Write-Host "  1. Terminal 1: .\manage.ps1 run-beacon" -ForegroundColor Gray
            Write-Host "  2. Terminal 2: .\manage.ps1 run-nexus" -ForegroundColor Gray
            Write-Host "  3. Terminal 3: .\manage.ps1 run-gjallar" -ForegroundColor Gray
            Write-Host ""
            Write-Host "Then retry: .\manage.ps1 gateway-test" -ForegroundColor Gray
            Write-Host ""
            exit 1
        }

        Write-Header "[Gateway] Running full tests..."
        
        $BaseURL = "http://localhost:8080"
        $TestsPassed = 0
        $TestsFailed = 0

        # Test function
        function Test-Endpoint {
            param(
                [string]$Method,
                [string]$Endpoint,
                [string]$Body,
                [int]$ExpectedCode
            )
            
            Write-Host -NoNewline "  CHECK $Method $Endpoint ... "
            
            try {
                $params = @{
                    Uri             = "$BaseURL$Endpoint"
                    Method          = $Method
                    Headers         = @{"Content-Type" = "application/json" }
                    UseBasicParsing = $true
                    TimeoutSec      = 5
                }
                
                if ($Body) {
                    $params["Body"] = $Body
                }
                
                $response = Invoke-WebRequest @params
                $httpCode = $response.StatusCode
                
                if ($httpCode -eq $ExpectedCode) {
                    Write-Host "OK HTTP $httpCode" -ForegroundColor Green
                    return $true
                }
                else {
                    Write-Host "FAIL HTTP $httpCode (expected $ExpectedCode)" -ForegroundColor Red
                    return $false
                }
            }
            catch {
                Write-Host "FAIL Error: $($_.Exception.Message)" -ForegroundColor Red
                return $false
            }
        }

        # Test Beacon service (read)
        Write-Host ""
        Write-Host "Testing Beacon service (read):" -ForegroundColor Cyan
        if (Test-Endpoint -Method "GET" -Endpoint "/v1/posts" -ExpectedCode 200) { $TestsPassed++ } else { $TestsFailed++ }
        if (Test-Endpoint -Method "GET" -Endpoint "/v1/categories" -ExpectedCode 200) { $TestsPassed++ } else { $TestsFailed++ }
        if (Test-Endpoint -Method "GET" -Endpoint "/v1/tags" -ExpectedCode 200) { $TestsPassed++ } else { $TestsFailed++ }

        # Test Nexus service (write)
        Write-Host ""
        Write-Host "✍️  Nexus 服务测试（写服务）:" -ForegroundColor Cyan
        
        $registerBody = @{
            username = "testuser"
            email    = "test@example.com"
            password = "password123"
            nickname = "Test User"
        } | ConvertTo-Json
        
        if (Test-Endpoint -Method "POST" -Endpoint "/v1/auth/register" -Body $registerBody -ExpectedCode 200) { $TestsPassed++ } else { $TestsFailed++ }

        # 总结
        Write-Host ""
        Write-Host "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" -ForegroundColor Cyan
        Write-Host "测试结果: 通过 $TestsPassed, 失败 $TestsFailed" -ForegroundColor Cyan
        Write-Host "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" -ForegroundColor Cyan
        Write-Host ""
        
        if ($TestsFailed -eq 0) {
            Write-Success "✅ 所有测试通过！gRPC Gateway 正常工作。"
        }
        else {
            Write-ErrorMsg "⚠️  部分测试失败。请检查网关和微服务。"
        }

        # ==============================================================================
        # 🐳 基础设施
        # ==============================================================================
        "up" {
            if (-not (Test-Path $ComposeFile)) { Write-ErrorMsg "$ComposeFile not found."; exit 1 }
            Write-Header "[Infra] Starting services..."
            Exec { docker-compose -f $ComposeFile up -d }
            Write-Success "Infrastructure is running!"
        }

        "down" {
            if (Test-Path $ComposeFile) {
                Write-Header "[Infra] Stopping services..."
                Exec { docker-compose -f $ComposeFile down }
            }
        }

        "logs" { Exec { docker-compose -f $ComposeFile logs -f } }
        "ps" { docker ps --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}" }

        # ==============================================================================
        # 🚀 Go 服务
        # ==============================================================================
        "run-nexus" { Write-Header "[Go] Nexus"; Push-Location $GoDir; Exec { go run cmd/nexus/main.go -config configs/nexus.yaml }; Pop-Location }
        "run-beacon" { Write-Header "[Go] Beacon"; Push-Location $GoDir; Exec { go run cmd/beacon/main.go -f configs/beacon.yaml }; Pop-Location }
        "run-gjallar" { Write-Header "[Go] Gjallar Gateway"; Push-Location $GoDir; Exec { go run cmd/gjallar/main.go }; Pop-Location }

        "build-go" {
            Write-Header "[Go] Building..."
            if (-not (Test-Path $BinDir)) { New-Item -ItemType Directory -Force -Path $BinDir | Out-Null }
            Push-Location $GoDir
            Exec { go build -o ../$BinDir/nexus.exe cmd/nexus/main.go }
            Exec { go build -o ../$BinDir/beacon.exe cmd/beacon/main.go }
            Exec { go build -o ../$BinDir/gjallar.exe cmd/gjallar/main.go }
            Pop-Location
            Write-Success "Built in ./$BinDir/"
        }

        # ==============================================================================
        # 🦀 Rust 服务
        # ==============================================================================
        "run-forge" { Write-Header "[Rust] Forge"; Push-Location $RustDir; Exec { cargo run --bin forge }; Pop-Location }
        "run-mirror" { Write-Header "[Rust] Mirror"; Push-Location $RustDir; Exec { cargo run --bin mirror }; Pop-Location }
        "build-rust" { Write-Header "[Rust] Building..."; Push-Location $RustDir; Exec { cargo build --release }; Pop-Location }

        # ==============================================================================
        # 🗄️ 数据库
        # ==============================================================================
        "migrate-up" {
            Write-Header "[DB] Migrating Up..."
            $DBUrl = "postgres://admin:secret@localhost:5432/bifrost?sslmode=disable"
            if (Test-Command migrate) { Exec { migrate -path migrations -database $DBUrl up } }
            elseif (Test-Command dbmate) { Exec { dbmate -u $DBUrl -d migrations up } }
            else { Write-ErrorMsg "Install migrate or dbmate first." }
        }

        "migrate-new" {
            if (-not $Name) { Write-ErrorMsg "Usage: .\manage.ps1 migrate-new -Name xxx"; exit 1 }
            Write-Header "[DB] New Migration: $Name"
            if (Test-Command migrate) { Exec { migrate create -ext sql -dir migrations -seq $Name } }
            elseif (Test-Command dbmate) { Exec { dbmate -d migrations new $Name } }
        }

        # ==============================================================================
        # 🧹 杂项
        # ==============================================================================
        "clean" {
            Write-Header "[Clean] Artifacts..."
            if (Test-Path $BinDir) { Remove-Item -Recurse -Force $BinDir }
            if (Test-Path "$GoDir/bin") { Remove-Item -Recurse -Force "$GoDir/bin" }
            if (Test-Path "$RustDir/target") { Push-Location $RustDir; cargo clean; Pop-Location }
            if (Test-Path $ThirdPartyDir) { Remove-Item -Recurse -Force $ThirdPartyDir }
            Write-Success "Cleaned."
        }

        "format" {
            Push-Location $GoDir; gofmt -w .; Pop-Location
            Push-Location $RustDir; cargo fmt; Pop-Location
        }

        "db-connect" {
            if (Test-Command psql) { & psql postgres://admin:secret@localhost:5432/bifrost }
            else { Write-ErrorMsg "psql not found." }
        }

        # ==============================================================================
        # ⚡ 快捷启动命令
        # ==============================================================================
        "dev" {
            Write-Header "⚡ 开发模式快速启动"
            Write-Host ""
            Write-Host "启动步骤:" -ForegroundColor Yellow
            Write-Host "  1️⃣  Terminal 1: .\manage.ps1 run-beacon" -ForegroundColor Cyan
            Write-Host "  2️⃣  Terminal 2: .\manage.ps1 run-nexus" -ForegroundColor Cyan
            Write-Host "  3️⃣  Terminal 3: .\manage.ps1 run-gjallar" -ForegroundColor Cyan
            Write-Host "  4️⃣  Terminal 4: .\manage.ps1 gateway-test" -ForegroundColor Cyan
            Write-Host ""
            Write-Host "或者快速启动第一个终端:" -ForegroundColor Yellow
            Write-Host "  .\manage.ps1 run-beacon" -ForegroundColor Gray
            Write-Host ""
        }

        "quick-start" {
            Write-Header "🚀 快速开始 (本地开发)"
            Write-Host ""
        
            # 1. 生成 Proto 代码
            Write-Header "[1/4] 生成 Proto 代码..."
            & $PSCommandPath "proto-gen"
        
            Write-Host ""
            Write-Host "✅ Proto 代码生成完成！" -ForegroundColor Green
            Write-Host ""
            Write-Host "接下来请在三个不同的终端中运行:" -ForegroundColor Yellow
            Write-Host "  Terminal 1: .\manage.ps1 run-beacon" -ForegroundColor Cyan
            Write-Host "  Terminal 2: .\manage.ps1 run-nexus" -ForegroundColor Cyan
            Write-Host "  Terminal 3: .\manage.ps1 run-gjallar" -ForegroundColor Cyan
            Write-Host ""
            Write-Host "然后测试网关:" -ForegroundColor Yellow
            Write-Host "  .\manage.ps1 gateway-test" -ForegroundColor Cyan
            Write-Host ""
        }

        default { Write-ErrorMsg "Unknown command: $Command" }
    }
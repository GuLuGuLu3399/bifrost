<#
.SYNOPSIS
  Bifrost 统一运维脚本
.DESCRIPTION
  提供开发、构建、Docker 管理等全部运维命令
  使用 GitHub 风格图标和中文界面
#>

[CmdletBinding()]
param(
    [Parameter(Position = 0)]
    [ValidateSet('help', 'proto-lint', 'proto-gen', 'build-go', 'format', 'clean',
        'docker-build', 'docker-build-go', 'docker-build-rust', 
        'docker-up', 'docker-up-infra', 'docker-up-go', 'docker-up-rust',
        'docker-down', 'docker-restart', 'docker-logs', 'docker-ps', 
        'docker-clean', 'docker-clean-all', 'docker-validate')]
    [string]$Command = 'help',

    [Parameter(Position = 1)]
    [string]$Arg1 = ''
)

Set-StrictMode -Version Latest
$ErrorActionPreference = 'Stop'

# 路径定义
$RootPath = $PSScriptRoot
$GoDir = Join-Path $RootPath 'go_services'
$RustDir = Join-Path $RootPath 'rust_services'
$BinDir = Join-Path $RootPath 'bin'

# 辅助函数
function Show-Help {
    Write-Host ''
    Write-Host "┌──────────────────────────────────────────────────┐" -ForegroundColor Cyan
    Write-Host "│  🚀 Bifrost 统一运维脚本                        │" -ForegroundColor Cyan
    Write-Host "└──────────────────────────────────────────────────┘" -ForegroundColor Cyan
    Write-Host ''
    Write-Host "📦 开发命令:" -ForegroundColor Green
    Write-Host "  proto-lint         🔍 检查 Proto 文件规范"
    Write-Host "  proto-gen          📡 生成 Proto 代码 (Protobuf + gRPC + Gateway)"
    Write-Host "  build-go           ⚙️  构建所有 Go 服务二进制文件"
    Write-Host "  format             ✨ 格式化 Go/Rust 代码"
    Write-Host "  clean              🧹 清理构建产物"
    Write-Host ''
    Write-Host "🐳 Docker 命令:" -ForegroundColor Green
    Write-Host "  docker-build       🔨 构建所有 Docker 镜像"
    Write-Host "  docker-build-go    🔨 仅构建 Go 服务镜像"
    Write-Host "  docker-build-rust  🔨 仅构建 Rust 服务镜像"
    Write-Host "  docker-up          ▶️  启动所有服务"
    Write-Host "  docker-up-infra    ▶️  启动基础设施服务"
    Write-Host "  docker-up-go       ▶️  启动 Go 服务"
    Write-Host "  docker-up-rust     ▶️  启动 Rust 服务"
    Write-Host "  docker-down        ⏹️  停止所有服务"
    Write-Host "  docker-restart     🔄 重启所有服务"
    Write-Host "  docker-logs [srv]  📋 查看日志 (可选指定服务名)"
    Write-Host "  docker-ps          📊 查看服务状态"
    Write-Host "  docker-clean       🗑️  清理停止的容器和未使用的镜像"
    Write-Host "  docker-clean-all   🗑️  清理所有数据（包括卷）⚠️"
    Write-Host "  docker-validate    ✅验证 Docker 环境配置"
    Write-Host ''
    Write-Host "📋 使用示例:" -ForegroundColor Yellow
    Write-Host "  .\manage.ps1 proto-gen"
    Write-Host "  .\manage.ps1 build-go"
    Write-Host "  .\manage.ps1 docker-build"
    Write-Host "  .\manage.ps1 docker-up"
    Write-Host "  .\manage.ps1 docker-logs nexus"
    Write-Host ''
}

function Write-Step([string]$Message) {
    Write-Host "  ➡️  $Message" -ForegroundColor Magenta
}

function Write-Info([string]$Message) {
    Write-Host "  !  $Message" -ForegroundColor Blue
}

function Write-Success([string]$Message) {
    Write-Host "  ✅ $Message" -ForegroundColor Green
}

function Write-ErrorMsg([string]$Message) {
    Write-Host "  ❌ $Message" -ForegroundColor Red
}

function Exec([scriptblock]$Block, [string]$ErrMsg = '命令执行失败') {
    try { 
        & $Block
        if ($LASTEXITCODE -ne 0) { 
            throw "$ErrMsg (退出码: $LASTEXITCODE)" 
        }
    }
    catch { 
        Write-ErrorMsg $_.Exception.Message
        exit 1 
    }
}

function Exec-In([string]$Path, [scriptblock]$Block) {
    if (-not (Test-Path $Path)) { 
        Write-ErrorMsg "目录不存在: $Path"
        exit 1 
    }
    Write-Host "  📁 进入目录: $(Split-Path -Leaf $Path)" -ForegroundColor Gray
    Push-Location $Path
    try { 
        Exec $Block 
    }
    finally { 
        Pop-Location 
    }
}

# 主逻辑
switch ($Command) {
    'help' { 
        Show-Help
        break 
    }

    'proto-lint' {
        Write-Host "`n🔍 检查 Proto 文件规范" -ForegroundColor Cyan
    
        if (-not (Get-Command buf -ErrorAction SilentlyContinue)) {
            Write-ErrorMsg "buf 未安装"
            Write-Host "  📥 安装命令: go install github.com/bufbuild/buf/cmd/buf@latest" -ForegroundColor Yellow
            exit 1
        }
    
        Exec { buf lint }
        Write-Success "Proto 文件规范检查通过"
        break
    }

    'proto-gen' {
        Write-Host "`n📡 生成 Proto 代码" -ForegroundColor Cyan
    
        if (-not (Get-Command buf -ErrorAction SilentlyContinue)) {
            Write-ErrorMsg "buf 未安装"
            Write-Host "  📥 安装命令: go install github.com/bufbuild/buf/cmd/buf@latest" -ForegroundColor Yellow
            exit 1
        }

        # 检查并安装必要的插件
        Write-Step "检查 protoc 插件..."
        $plugins = @(
            @{name = 'protoc-gen-go'; pkg = 'google.golang.org/protobuf/cmd/protoc-gen-go@latest' },
            @{name = 'protoc-gen-go-grpc'; pkg = 'google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest' },
            @{name = 'protoc-gen-grpc-gateway'; pkg = 'github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@latest' }
        )
    
        foreach ($plugin in $plugins) {
            if (-not (Get-Command $plugin.name -ErrorAction SilentlyContinue)) {
                Write-Host "  📥 安装 $($plugin.name)..." -ForegroundColor Yellow
                Exec { go install $plugin.pkg }
                Write-Success "$($plugin.name) 安装完成"
            }
            else {
                Write-Host "  ✅ $($plugin.name) 已安装" -ForegroundColor Green
            }
        }

        # 更新依赖
        Write-Step "更新 Proto 依赖..."
        Exec { buf dep update }
    
        # 生成代码
        Write-Step "生成 Proto 代码..."
        Exec { buf generate }
    
        # 清理不需要的 Gateway 文件
        Write-Step "清理 gRPC-only 服务的 HTTP 网关文件..."
        $grpcOnlyFiles = @(
            'go_services/api/content/v1/forge/forge.pb.gw.go',
            'go_services/api/content/v1/oracle/oracle.pb.gw.go',
            'go_services/api/search/v1/mirror.pb.gw.go'
        )
    
        $deletedCount = 0
        foreach ($file in $grpcOnlyFiles) {
            $fullPath = Join-Path $RootPath $file
            if (Test-Path $fullPath) { 
                Remove-Item $fullPath -Force
                Write-Host "  🗑️  已删除: $file" -ForegroundColor Gray
                $deletedCount++
            }
        }
        if ($deletedCount -gt 0) {
            Write-Info "已删除 $deletedCount 个 gRPC-only 网关文件"
        }
    
        Write-Success "Proto 代码生成完成"
        Write-Host ""
        Write-Host "📁 生成的文件位置:" -ForegroundColor Cyan
        Write-Host "  go_services/api/" -ForegroundColor Gray
        break
    }

    'build-go' {
        Write-Host "`n⚙️  构建 Go 服务" -ForegroundColor Cyan
    
        if (-not (Test-Path $BinDir)) { 
            New-Item -ItemType Directory -Force -Path $BinDir | Out-Null
            Write-Info "创建输出目录: $BinDir"
        }
    
        Exec-In $GoDir {
            $OutBase = Join-Path '..' 'bin'
      
            Write-Step "构建 nexus..."
            go build -o (Join-Path $OutBase 'nexus.exe') cmd/nexus/main.go
            Write-Info "  nexus.exe 已生成"
      
            Write-Step "构建 beacon..."
            go build -o (Join-Path $OutBase 'beacon.exe') cmd/beacon/main.go
            Write-Info "  beacon.exe 已生成"
      
            Write-Step "构建 gjallar..."
            go build -o (Join-Path $OutBase 'gjallar.exe') cmd/gjallar/main.go
            Write-Info "  gjallar.exe 已生成"
        }
    
        Write-Success "所有 Go 服务构建完成"
        Write-Host "  📁 输出目录: $BinDir" -ForegroundColor Gray
        break
    }

    'format' {
        Write-Host "`n✨ 格式化代码" -ForegroundColor Cyan
    
        if (Get-Command gofmt -ErrorAction SilentlyContinue) {
            Write-Step "格式化 Go 代码..."
            Exec-In $GoDir { gofmt -w . }
            Write-Success "Go 代码格式化完成"
        }
        else {
            Write-ErrorMsg "gofmt 未找到，跳过 Go 代码格式化"
        }
    
        if (Get-Command cargo -ErrorAction SilentlyContinue) {
            Write-Step "格式化 Rust 代码..."
            Exec-In $RustDir { cargo fmt }
            Write-Success "Rust 代码格式化完成"
        }
        else {
            Write-ErrorMsg "cargo 未找到，跳过 Rust 代码格式化"
        }
    
        Write-Success "所有代码格式化完成"
        break
    }

    'clean' {
        Write-Host "`n🧹 清理构建产物" -ForegroundColor Cyan
    
        $items = @(
            @{path = $BinDir; desc = "二进制文件目录" },
            @{path = "$GoDir/bin"; desc = "Go 编译缓存" },
            @{path = "$RustDir/target"; desc = "Rust 编译缓存" }
        )
    
        $cleanedCount = 0
        foreach ($item in $items) {
            if (Test-Path $item.path) { 
                Remove-Item -Recurse -Force $item.path
                Write-Host "  🗑️  已清理: $($item.desc)" -ForegroundColor Gray
                $cleanedCount++
            }
        }
    
        if ($cleanedCount -eq 0) {
            Write-Info "没有需要清理的文件"
        }
        else {
            Write-Success "清理完成，已清理 $cleanedCount 个目录"
        }
        break
    }

    # ==========================================
    # Docker 管理命令
    # ==========================================

    'docker-validate' {
        Write-Host "`n✅ 验证 Docker 环境" -ForegroundColor Cyan
        
        $errors = @()
        
        # 检查 Docker
        Write-Step "检查 Docker..."
        if (Get-Command docker -ErrorAction SilentlyContinue) {
            $dockerVersion = docker --version
            Write-Success $dockerVersion
        }
        else {
            Write-ErrorMsg "Docker 未安装或未运行"
            $errors += "Docker 未安装"
        }
        
        # 检查 Docker Compose
        Write-Step "检查 Docker Compose..."
        if (Get-Command docker-compose -ErrorAction SilentlyContinue) {
            $composeVersion = docker-compose --version
            Write-Success $composeVersion
        }
        else {
            Write-ErrorMsg "Docker Compose 未安装"
            $errors += "Docker Compose 未安装"
        }
        
        # 验证配置文件
        Write-Step "验证 docker-compose.yml..."
        if (Test-Path "docker-compose.yml") {
            try {
                $null = docker-compose config --quiet 2>&1
                if ($LASTEXITCODE -eq 0) {
                    Write-Success "配置文件语法正确"
                }
                else {
                    Write-ErrorMsg "docker-compose.yml 语法错误"
                    $errors += "docker-compose.yml 语法错误"
                }
            }
            catch {
                Write-ErrorMsg "docker-compose.yml 验证失败"
                $errors += "验证失败"
            }
        }
        else {
            Write-ErrorMsg "缺少 docker-compose.yml"
            $errors += "缺少 docker-compose.yml"
        }
        
        # 总结
        Write-Host ""
        if ($errors.Count -eq 0) {
            Write-Success "所有检查通过!"
            Write-Host ""
            Write-Info "下一步: .\manage.ps1 docker-build"
        }
        else {
            Write-ErrorMsg "发现 $($errors.Count) 个问题"
        }
        break
    }

    'docker-build' {
        Write-Host "`n🔨 构建所有 Docker 镜像" -ForegroundColor Cyan
        Exec { docker-compose build }
        Write-Success "所有镜像构建完成"
        break
    }

    'docker-build-go' {
        Write-Host "`n🔨 构建 Go 服务镜像" -ForegroundColor Cyan
        Exec { docker-compose build nexus beacon gjallar }
        Write-Success "Go 服务镜像构建完成"
        break
    }

    'docker-build-rust' {
        Write-Host "`n🔨 构建 Rust 服务镜像" -ForegroundColor Cyan
        Exec { docker-compose build forge mirror oracle }
        Write-Success "Rust 服务镜像构建完成"
        break
    }

    'docker-up' {
        Write-Host "`n▶️  启动所有服务" -ForegroundColor Cyan
        Exec { docker-compose up -d }
        Write-Host ""
        docker-compose ps
        Write-Host ""
        Write-Success "所有服务已启动"
        Write-Info "查看日志: .\manage.ps1 docker-logs"
        Write-Info "访问网关: http://localhost:8080"
        Write-Info "Jaeger UI: http://localhost:16686"
        break
    }

    'docker-up-infra' {
        Write-Host "`n▶️  启动基础设施服务" -ForegroundColor Cyan
        Exec { docker-compose up -d postgres redis nats minio jaeger createbuckets }
        Write-Host ""
        docker-compose ps
        Write-Host ""
        Write-Success "基础设施服务已启动"
        break
    }

    'docker-up-go' {
        Write-Host "`n▶️  启动 Go 服务" -ForegroundColor Cyan
        Exec { docker-compose up -d nexus beacon gjallar }
        Write-Host ""
        docker-compose ps
        Write-Host ""
        Write-Success "Go 服务已启动"
        break
    }

    'docker-up-rust' {
        Write-Host "`n▶️  启动 Rust 服务" -ForegroundColor Cyan
        Exec { docker-compose up -d forge mirror oracle }
        Write-Host ""
        docker-compose ps
        Write-Host ""
        Write-Success "Rust 服务已启动"
        break
    }

    'docker-down' {
        Write-Host "`n⏹️  停止所有服务" -ForegroundColor Yellow
        Exec { docker-compose down }
        Write-Success "所有服务已停止"
        break
    }

    'docker-restart' {
        Write-Host "`n🔄 重启所有服务" -ForegroundColor Yellow
        Exec { docker-compose restart }
        Write-Success "所有服务已重启"
        break
    }

    'docker-logs' {
        if ($Arg1) {
            Write-Host "`n📋 查看 $Arg1 日志" -ForegroundColor Cyan
            docker-compose logs -f --tail=100 $Arg1
        }
        else {
            Write-Host "`n📋 查看所有服务日志" -ForegroundColor Cyan
            docker-compose logs -f --tail=50
        }
        break
    }

    'docker-ps' {
        Write-Host "`n📊 服务运行状态" -ForegroundColor Cyan
        docker-compose ps
        break
    }

    'docker-clean' {
        Write-Host "`n🗑️  清理 Docker 资源" -ForegroundColor Yellow
        Write-Step "清理停止的容器..."
        docker container prune -f
        Write-Step "清理未使用的镜像..."
        docker image prune -f
        Write-Success "清理完成"
        break
    }

    'docker-clean-all' {
        Write-Host "`n⚠️  警告: 这将删除所有容器、镜像和数据卷!" -ForegroundColor Red
        $confirm = Read-Host "确认继续? (yes/no)"
        
        if ($confirm -eq "yes") {
            Write-Host "`n🗑️  清理所有 Docker 资源" -ForegroundColor Yellow
            docker-compose down -v
            docker system prune -a -f --volumes
            Write-Success "清理完成"
        }
        else {
            Write-Info "已取消"
        }
        break
    }
}
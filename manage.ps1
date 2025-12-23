<#
.SYNOPSIS
  Bifrost 运维脚本 (精简版)
.DESCRIPTION
  提供核心运维命令：proto-lint、proto-gen、build-go、format、clean
  使用 GitHub 风格图标和中文界面
#>

[CmdletBinding()]
param(
    [Parameter(Position = 0)]
    [ValidateSet('help', 'proto-lint', 'proto-gen', 'build-go', 'format', 'clean')]
    [string]$Command = 'help'
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
    Write-Host "┌─────────────────────────────────────────────┐" -ForegroundColor Cyan
    Write-Host "│  🚀 Bifrost 运维脚本 (精简版)              │" -ForegroundColor Cyan
    Write-Host "└─────────────────────────────────────────────┘" -ForegroundColor Cyan
    Write-Host ''
    Write-Host "🔧 可用命令:" -ForegroundColor Green
    Write-Host "  proto-lint    🔍 检查 Proto 文件规范"
    Write-Host "  proto-gen     📡 生成 Proto 代码 (Protobuf + gRPC + Gateway)"
    Write-Host "  build-go      ⚙️  构建所有 Go 服务二进制文件"
    Write-Host "  format        ✨ 格式化 Go/Rust 代码"
    Write-Host "  clean         🧹 清理构建产物"
    Write-Host ''
    Write-Host "📋 使用示例:" -ForegroundColor Yellow
    Write-Host "  .\manage.ps1 proto-gen"
    Write-Host "  .\manage.ps1 build-go"
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
}
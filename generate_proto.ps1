# gRPC Gateway 生成脚本

$ErrorActionPreference = "Stop"

Write-Host "🔄 生成 Proto 代码（protobuf + gRPC + gRPC Gateway）..." -ForegroundColor Cyan

# 检查 buf 是否安装
try {
    $bufVersion = & buf --version 2>&1
    Write-Host "✅ buf 已安装: $bufVersion" -ForegroundColor Green
} catch {
    Write-Host "❌ buf 未安装！" -ForegroundColor Red
    Write-Host "请先安装 buf: go install github.com/bufbuild/buf/cmd/buf@latest" -ForegroundColor Yellow
    exit 1
}

# 检查必要的 proto 插件
Write-Host "✅ 检查 protoc 插件..." -ForegroundColor Green

# 运行生成
Write-Host "🚀 运行 buf generate..." -ForegroundColor Cyan
& buf generate

if ($LASTEXITCODE -eq 0) {
    Write-Host "✅ 生成成功！" -ForegroundColor Green
    Write-Host ""
    Write-Host "📁 生成的文件：" -ForegroundColor Cyan
    Write-Host "   - *.pb.go          (protobuf 数据结构)" -ForegroundColor Gray
    Write-Host "   - *_grpc.pb.go     (gRPC 服务定义)" -ForegroundColor Gray
    Write-Host "   - *.pb.gw.go       (gRPC Gateway 处理器)" -ForegroundColor Gray
    Write-Host ""
    Write-Host "💡 下一步：运行网关服务" -ForegroundColor Cyan
    Write-Host "   cd go_services && go run ./cmd/gjallar/main.go" -ForegroundColor Gray
} else {
    Write-Host "❌ 生成失败！" -ForegroundColor Red
    exit 1
}

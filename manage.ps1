<#
.SYNOPSIS
  Bifrost 简化运维脚本
.DESCRIPTION
  仅保留基础 Docker 管理能力，降低维护复杂度。
#>

[CmdletBinding()]
param(
    [Parameter(Position = 0)]
    [ValidateSet('help', 'docker-validate', 'docker-build', 'docker-up', 'docker-down', 'docker-restart', 'docker-logs', 'docker-ps')]
    [string]$Command = 'help',

    [Parameter(Position = 1)]
    [string]$Arg1 = ''
)

Set-StrictMode -Version Latest
$ErrorActionPreference = 'Stop'

$RootPath = $PSScriptRoot

function Write-Info([string]$Message) {
    Write-Host "[INFO] $Message" -ForegroundColor Cyan
}

function Write-Success([string]$Message) {
    Write-Host "[OK] $Message" -ForegroundColor Green
}

function Write-ErrorMsg([string]$Message) {
    Write-Host "[ERR] $Message" -ForegroundColor Red
}

function Invoke-Compose {
    param(
        [Parameter(Mandatory = $true)]
        [string[]]$Args,
        [string]$ErrMsg = 'docker compose 执行失败'
    )

    Push-Location $RootPath
    try {
        & docker compose @Args
        if ($LASTEXITCODE -ne 0) {
            throw "$ErrMsg (退出码: $LASTEXITCODE)"
        }
    }
    finally {
        Pop-Location
    }
}

function Show-Help {
    Write-Host ''
    Write-Host 'Bifrost 简化运维脚本' -ForegroundColor Yellow
    Write-Host ''
    Write-Host '基础命令:' -ForegroundColor Green
    Write-Host '  docker-validate    检查 Docker 与 compose 配置'
    Write-Host '  docker-build       构建全部服务镜像'
    Write-Host '  docker-up          启动全部服务'
    Write-Host '  docker-down        停止并移除全部服务容器'
    Write-Host '  docker-restart     重启全部服务'
    Write-Host '  docker-ps          查看服务状态'
    Write-Host '  docker-logs [srv]  查看日志（可选指定服务）'
    Write-Host ''
    Write-Host '示例:' -ForegroundColor Green
    Write-Host '  .\manage.ps1 docker-validate'
    Write-Host '  .\manage.ps1 docker-build'
    Write-Host '  .\manage.ps1 docker-up'
    Write-Host '  .\manage.ps1 docker-logs gjallar'
    Write-Host ''
}

switch ($Command) {
    'help' {
        Show-Help
        break
    }

    'docker-validate' {
        $errors = @()

        Write-Info '检查 Docker CLI...'
        if (Get-Command docker -ErrorAction SilentlyContinue) {
            & docker --version
        }
        else {
            $errors += 'Docker 未安装或不可用'
            Write-ErrorMsg 'Docker 未安装或不可用'
        }

        Write-Info '检查 docker compose...'
        try {
            & docker compose version | Out-Null
            if ($LASTEXITCODE -ne 0) {
                throw 'docker compose 不可用'
            }
            & docker compose version
        }
        catch {
            $errors += 'docker compose 不可用'
            Write-ErrorMsg 'docker compose 不可用'
        }

        Write-Info '校验 compose 配置...'
        if (-not (Test-Path (Join-Path $RootPath 'docker-compose.yml'))) {
            $errors += '缺少 docker-compose.yml'
            Write-ErrorMsg '缺少 docker-compose.yml'
        }
        else {
            try {
                Invoke-Compose -Args @('config', '--quiet') -ErrMsg 'compose 配置校验失败'
                Write-Success 'compose 配置校验通过'
            }
            catch {
                $errors += 'compose 配置校验失败'
                Write-ErrorMsg $_.Exception.Message
            }
        }

        if ($errors.Count -eq 0) {
            Write-Success '环境检查通过'
        }
        else {
            Write-ErrorMsg ("检查失败，问题数: {0}" -f $errors.Count)
            exit 1
        }
        break
    }

    'docker-build' {
        Write-Info '构建全部镜像...'
        Invoke-Compose -Args @('build') -ErrMsg '镜像构建失败'
        Write-Success '镜像构建完成'
        break
    }

    'docker-up' {
        Write-Info '启动全部服务...'
        Invoke-Compose -Args @('up', '-d') -ErrMsg '服务启动失败'
        Invoke-Compose -Args @('ps') -ErrMsg '读取服务状态失败'
        Write-Success '服务已启动'
        break
    }

    'docker-down' {
        Write-Info '停止全部服务...'
        Invoke-Compose -Args @('down') -ErrMsg '服务停止失败'
        Write-Success '服务已停止'
        break
    }

    'docker-restart' {
        Write-Info '重启全部服务...'
        Invoke-Compose -Args @('restart') -ErrMsg '服务重启失败'
        Write-Success '服务已重启'
        break
    }

    'docker-ps' {
        Invoke-Compose -Args @('ps') -ErrMsg '读取服务状态失败'
        break
    }

    'docker-logs' {
        if ([string]::IsNullOrWhiteSpace($Arg1)) {
            Invoke-Compose -Args @('logs', '-f', '--tail=100') -ErrMsg '读取日志失败'
        }
        else {
            Invoke-Compose -Args @('logs', '-f', '--tail=100', $Arg1) -ErrMsg '读取服务日志失败'
        }
        break
    }
}
#!/usr/bin/env pwsh
# Bifrost pre-commit hook (PowerShell)
# Purpose: Block committing build outputs, secrets, very large files, and other disallowed content.

param()

Write-Host "Running Bifrost pre-commit checks..." -ForegroundColor Cyan

$ErrorActionPreference = 'Stop'

# Collect staged files (Added/Modified/Renamed)
$staged = git diff --cached --name-only --diff-filter=ACMR | Where-Object { $_ -ne '' }

if (-not $staged -or $staged.Count -eq 0) {
    Write-Host "No staged files. Skipping checks." -ForegroundColor Yellow
    exit 0
}

# Disallowed path regexes
$denyPatterns = @(
    '^rust_services/target/',
    '^target/',
    '^go_services/.+/bin/',
    '^rust_services/.+/.*/target/',
    '^node_modules/',
    '^\.vscode/',
    '^\.idea/'
)

# Disallowed extensions / files
$denyExt = @('*.env', '*.key', '*.pem', '*.p12', '*.cer')

# Secret content heuristics (best-effort)
$secretIndicators = @(
    'secret_key:\s*\S+',
    'password:\s*\S+',
    'access_key_id:\s*\S+',
    'private_key',
    'BEGIN RSA PRIVATE KEY',
    'BEGIN PRIVATE KEY'
)

$failures = @()

foreach ($file in $staged) {
    # Path-based blocking
    foreach ($pat in $denyPatterns) {
        if ($file -match $pat) {
            $failures += "Disallowed path: $file (matches '$pat')"
            break
        }
    }

    # Extension-based blocking
    foreach ($glob in $denyExt) {
        if ([IO.Path]::GetFileName($file) -like $glob) {
            $failures += "Sensitive file extension blocked: $file"
            break
        }
    }

    # Size check (> 5 MB)
    if (Test-Path $file) {
        try {
            $size = (Get-Item $file).Length
            if ($size -gt 5MB) {
                $failures += "File too large (>5MB): $file (${size} bytes)"
            }
        }
        catch {
            # ignore size errors
        }
    }

    # Content secret scan for config-like files
    $ext = [IO.Path]::GetExtension($file).ToLowerInvariant()
    if ($ext -in @('.yaml', '.yml', '.toml', '.env')) {
        try {
            $text = Get-Content -LiteralPath $file -Raw -ErrorAction Stop
            foreach ($sig in $secretIndicators) {
                if ($text -match $sig) {
                    # Allow env references like ${...} or from vault; block obvious literals
                    if ($text -notmatch '\$\{.+\}' -and $text -notmatch 'vault:') {
                        $failures += "Possible secret in $file (matched '$sig')"
                        break
                    }
                }
            }
        }
        catch {
            # ignore read errors for binary files
        }
    }
}

if ($failures.Count -gt 0) {
    Write-Host "Pre-commit checks failed:" -ForegroundColor Red
    $failures | ForEach-Object { Write-Host " - $_" -ForegroundColor Red }
    Write-Host "\nTo bypass in emergencies, you may run 'git commit -n' (not recommended)." -ForegroundColor Yellow
    exit 1
}

Write-Host "All pre-commit checks passed." -ForegroundColor Green
exit 0

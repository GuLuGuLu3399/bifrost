#!/usr/bin/env pwsh
# Bifrost git filter helper
# Purpose: Stage changes by module scope, excluding build outputs and generated artifacts.

param(
    [Parameter(Mandatory = $true)]
    [ValidateSet('go', 'rust', 'api', 'migrations', 'docs', 'all')]
    [string]$Scope
)

$ErrorActionPreference = 'Stop'

function StageWithExcludes($paths) {
    # Use Git pathspecs with :(exclude) to avoid build outputs.
    # Note: quoting is important on Windows PowerShell.
    $cmd = @('add', '--')
    $cmd += $paths
    # common excludes
    $cmd += ':(exclude)rust_services/target/**'
    $cmd += ':(exclude)target/**'
    $cmd += ':(exclude)**/*.env'
    $cmd += ':(exclude)**/*.pem'
    $cmd += ':(exclude)**/*.key'
    $cmd += ':(exclude)**/*.p12'
    $cmd += ':(exclude)go_services/**/bin/**'

    git @cmd
}

switch ($Scope) {
    'go' { StageWithExcludes @('go_services/**') }
    'rust' { StageWithExcludes @('rust_services/**') }
    'api' { StageWithExcludes @('api/**') }
    'migrations' { StageWithExcludes @('migrations/**') }
    'docs' { StageWithExcludes @('docs/**') }
    'all' { StageWithExcludes @('go_services/**', 'rust_services/**', 'api/**', 'migrations/**', 'docs/**') }
}

Write-Host "Staged changes for scope '$Scope' with common excludes." -ForegroundColor Green

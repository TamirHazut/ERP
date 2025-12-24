# PowerShell script for generating proto files on Windows

param(
    [Parameter(Mandatory=$false)]
    [ValidateSet("all", "common", "auth", "config", "core")]
    [string]$Service = "all"
)

$ErrorActionPreference = "Stop"

# Configuration
$MODULE = "erp.localhost"
$PROTO_OUT = "internal"
$PROTO_COMMON = "proto/common"
$PROTO_AUTH = "internal/auth/proto"
$PROTO_CONFIG = "internal/config/proto"
$PROTO_CORE = "internal/core/proto"

function Generate-Common {
    Write-Host "Generating common proto files..." -ForegroundColor Cyan
    if (Test-Path "$PROTO_COMMON/common.proto") {
        protoc --go_out=$PROTO_OUT `
            --go_opt=module=$MODULE `
            --go-grpc_out=$PROTO_OUT `
            --go-grpc_opt=module=$MODULE `
            -I=proto `
            "$PROTO_COMMON/common.proto"
        Write-Host "✓ Common proto files generated" -ForegroundColor Green
    } else {
        Write-Host "⚠ No common.proto file found, skipping..." -ForegroundColor Yellow
    }
}

function Generate-Auth {
    Write-Host "Generating auth service proto files..." -ForegroundColor Cyan
    if (Test-Path "$PROTO_AUTH/auth.proto") {
        protoc --go_out=$PROTO_OUT `
            --go_opt=module=$MODULE `
            --go-grpc_out=$PROTO_OUT `
            --go-grpc_opt=module=$MODULE `
            -I=proto `
            -I=$PROTO_AUTH `
            "$PROTO_AUTH/*.proto"
        Write-Host "✓ Auth proto files generated" -ForegroundColor Green
    } else {
        Write-Host "⚠ No auth.proto file found, skipping..." -ForegroundColor Yellow
    }
}

function Generate-Config {
    Write-Host "Generating config service proto files..." -ForegroundColor Cyan
    if (Test-Path "$PROTO_CONFIG/config.proto") {
        protoc --go_out=$PROTO_OUT `
            --go_opt=module=$MODULE `
            --go-grpc_out=$PROTO_OUT `
            --go-grpc_opt=module=$MODULE `
            -I=proto `
            -I=$PROTO_CONFIG `
            "$PROTO_CONFIG/*.proto"
        Write-Host "✓ Config proto files generated" -ForegroundColor Green
    } else {
        Write-Host "⚠ No config.proto file found, skipping..." -ForegroundColor Yellow
    }
}

function Generate-Core {
    Write-Host "Generating core service proto files..." -ForegroundColor Cyan
    if (Test-Path "$PROTO_CORE/core.proto") {
        protoc --go_out=$PROTO_OUT `
            --go_opt=module=$MODULE `
            --go-grpc_out=$PROTO_OUT `
            --go-grpc_opt=module=$MODULE `
            -I=proto `
            -I=$PROTO_CORE `
            "$PROTO_CORE/*.proto"
        Write-Host "✓ Core proto files generated" -ForegroundColor Green
    } else {
        Write-Host "⚠ No core.proto file found, skipping..." -ForegroundColor Yellow
    }
}

# Main execution
Write-Host "=== Proto Code Generation ===" -ForegroundColor Magenta

switch ($Service) {
    "all" {
        Generate-Common
        Generate-Auth
        Generate-Config
        Generate-Core
    }
    "common" { Generate-Common }
    "auth" { Generate-Auth }
    "config" { Generate-Config }
    "core" { Generate-Core }
}

Write-Host "`n✓ Proto generation complete!" -ForegroundColor Green


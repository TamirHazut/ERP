# PowerShell build script for Windows

param(
    [Parameter(Mandatory=$false)]
    [ValidateSet("all", "auth", "config", "core", "test", "test-coverage", "clean", "tidy", "docker-up", "docker-down", "docker-logs", "docker-ps", "help")]
    [string]$Target = "help"
)

$ErrorActionPreference = "Stop"

# Configuration
$BIN_DIR = "bin"
$AUTH_PORT = 5000
$CONFIG_PORT = 5002
$CORE_PORT = 5001

function Show-Help {
    Write-Host "ERP System - Build Script" -ForegroundColor Magenta
    Write-Host ""
    Write-Host "Usage: .\scripts\build.ps1 -Target <target>" -ForegroundColor Cyan
    Write-Host ""
    Write-Host "Build Targets:" -ForegroundColor Yellow
    Write-Host "  all           - Build all services"
    Write-Host "  auth          - Build auth service"
    Write-Host "  config        - Build config service"
    Write-Host "  core          - Build core service"
    Write-Host ""
    Write-Host "Test Targets:" -ForegroundColor Yellow
    Write-Host "  test          - Run all tests"
    Write-Host "  test-coverage - Run tests with coverage"
    Write-Host ""
    Write-Host "Docker Targets:" -ForegroundColor Yellow
    Write-Host "  docker-up     - Start MongoDB and Redis containers"
    Write-Host "  docker-down   - Stop and remove containers"
    Write-Host "  docker-logs   - View container logs"
    Write-Host "  docker-ps     - List running containers"
    Write-Host ""
    Write-Host "Utility Targets:" -ForegroundColor Yellow
    Write-Host "  clean         - Clean build artifacts"
    Write-Host "  tidy          - Run go mod tidy"
    Write-Host "  help          - Show this help message"
    Write-Host ""
    Write-Host "Examples:" -ForegroundColor Yellow
    Write-Host "  .\scripts\build.ps1 -Target all"
    Write-Host "  .\scripts\build.ps1 -Target docker-up"
    Write-Host "  .\scripts\build.ps1 -Target test"
}

function Build-All {
    Write-Host "Building all services..." -ForegroundColor Cyan
    New-Item -ItemType Directory -Force -Path $BIN_DIR | Out-Null
    go build -o "$BIN_DIR/" ./...
    if ($LASTEXITCODE -eq 0) {
        Write-Host "✓ Build complete" -ForegroundColor Green
    } else {
        Write-Host "✗ Build failed" -ForegroundColor Red
        exit 1
    }
}

function Build-Auth {
    Write-Host "Building auth service..." -ForegroundColor Cyan
    New-Item -ItemType Directory -Force -Path $BIN_DIR | Out-Null
    go build -o "$BIN_DIR/auth.exe" ./cmd/auth
    if ($LASTEXITCODE -eq 0) {
        Write-Host "✓ Auth service built" -ForegroundColor Green
    } else {
        Write-Host "✗ Build failed" -ForegroundColor Red
        exit 1
    }
}

function Build-Config {
    Write-Host "Building config service..." -ForegroundColor Cyan
    New-Item -ItemType Directory -Force -Path $BIN_DIR | Out-Null
    go build -o "$BIN_DIR/config.exe" ./cmd/config
    if ($LASTEXITCODE -eq 0) {
        Write-Host "✓ Config service built" -ForegroundColor Green
    } else {
        Write-Host "✗ Build failed" -ForegroundColor Red
        exit 1
    }
}

function Build-Core {
    Write-Host "Building core service..." -ForegroundColor Cyan
    New-Item -ItemType Directory -Force -Path $BIN_DIR | Out-Null
    go build -o "$BIN_DIR/core.exe" ./cmd/core
    if ($LASTEXITCODE -eq 0) {
        Write-Host "✓ Core service built" -ForegroundColor Green
    } else {
        Write-Host "✗ Build failed" -ForegroundColor Red
        exit 1
    }
}

function Run-Tests {
    Write-Host "Running tests..." -ForegroundColor Cyan
    go test -v ./...
    if ($LASTEXITCODE -eq 0) {
        Write-Host "✓ Tests complete" -ForegroundColor Green
    } else {
        Write-Host "✗ Tests failed" -ForegroundColor Red
        exit 1
    }
}

function Run-TestCoverage {
    Write-Host "Running tests with coverage..." -ForegroundColor Cyan
    go test -v -coverprofile=coverage.out ./...
    if ($LASTEXITCODE -eq 0) {
        go tool cover -html=coverage.out -o coverage.html
        Write-Host "✓ Coverage report generated: coverage.html" -ForegroundColor Green
    } else {
        Write-Host "✗ Tests failed" -ForegroundColor Red
        exit 1
    }
}

function Clean-Artifacts {
    Write-Host "Cleaning build artifacts..." -ForegroundColor Cyan
    if (Test-Path $BIN_DIR) {
        Remove-Item -Recurse -Force $BIN_DIR
    }
    if (Test-Path "coverage.out") {
        Remove-Item -Force "coverage.out"
    }
    if (Test-Path "coverage.html") {
        Remove-Item -Force "coverage.html"
    }
    Write-Host "✓ Clean complete" -ForegroundColor Green
}

function Run-Tidy {
    Write-Host "Running go mod tidy..." -ForegroundColor Cyan
    go mod tidy
    if ($LASTEXITCODE -eq 0) {
        Write-Host "✓ Dependencies updated" -ForegroundColor Green
    } else {
        Write-Host "✗ Tidy failed" -ForegroundColor Red
        exit 1
    }
}

function Docker-Up {
    Write-Host "Starting Docker containers..." -ForegroundColor Cyan
    docker compose up -d
    if ($LASTEXITCODE -eq 0) {
        Write-Host "✓ Containers started" -ForegroundColor Green
        Write-Host ""
        Write-Host "Services:" -ForegroundColor Yellow
        Write-Host "  MongoDB: mongodb://root:secret@localhost:27017"
        Write-Host "  Redis:   redis://:supersecretredis@localhost:6379"
    } else {
        Write-Host "✗ Docker start failed" -ForegroundColor Red
        exit 1
    }
}

function Docker-Down {
    Write-Host "Stopping Docker containers..." -ForegroundColor Cyan
    docker compose down
    if ($LASTEXITCODE -eq 0) {
        Write-Host "✓ Containers stopped" -ForegroundColor Green
    } else {
        Write-Host "✗ Docker stop failed" -ForegroundColor Red
        exit 1
    }
}

function Docker-Logs {
    Write-Host "Showing container logs (Ctrl+C to exit)..." -ForegroundColor Cyan
    docker compose logs -f
}

function Docker-Ps {
    Write-Host "Running containers:" -ForegroundColor Cyan
    docker compose ps
}

# Main execution
Write-Host "=== ERP Build System ===" -ForegroundColor Magenta

switch ($Target) {
    "all" { Build-All }
    "auth" { Build-Auth }
    "config" { Build-Config }
    "core" { Build-Core }
    "test" { Run-Tests }
    "test-coverage" { Run-TestCoverage }
    "clean" { Clean-Artifacts }
    "tidy" { Run-Tidy }
    "docker-up" { Docker-Up }
    "docker-down" { Docker-Down }
    "docker-logs" { Docker-Logs }
    "docker-ps" { Docker-Ps }
    "help" { Show-Help }
}


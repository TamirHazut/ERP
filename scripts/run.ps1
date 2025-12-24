# PowerShell script to run services

param(
    [Parameter(Mandatory=$false)]
    [ValidateSet("auth", "config", "core", "help")]
    [string]$Service = "help"
)

$ErrorActionPreference = "Stop"

# Service ports
$Ports = @{
    "auth" = 5000
    "config" = 5002
    "core" = 5001
}

function Show-Help {
    Write-Host "ERP System - Run Services" -ForegroundColor Magenta
    Write-Host ""
    Write-Host "Usage: .\scripts\run.ps1 -Service <service>" -ForegroundColor Cyan
    Write-Host ""
    Write-Host "Services:" -ForegroundColor Yellow
    Write-Host "  auth   - Run auth service (port 5000)"
    Write-Host "  config - Run config service (port 5002)"
    Write-Host "  core   - Run core service (port 5001)"
    Write-Host "  help   - Show this help message"
    Write-Host ""
    Write-Host "Examples:" -ForegroundColor Yellow
    Write-Host "  .\scripts\run.ps1 -Service auth"
    Write-Host "  .\scripts\run.ps1 -Service core"
}

function Run-Service {
    param([string]$Name)
    
    $Port = $Ports[$Name]
    Write-Host "Starting $Name service on port $Port..." -ForegroundColor Cyan
    Write-Host "Press Ctrl+C to stop" -ForegroundColor Yellow
    Write-Host ""
    
    go run "./cmd/$Name"
}

# Main execution
Write-Host "=== ERP Service Runner ===" -ForegroundColor Magenta

switch ($Service) {
    "auth" { Run-Service -Name "auth" }
    "config" { Run-Service -Name "config" }
    "core" { Run-Service -Name "core" }
    "help" { Show-Help }
}


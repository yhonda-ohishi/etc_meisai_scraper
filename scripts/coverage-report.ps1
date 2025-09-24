# Coverage report script for ETC Meisai project (Windows PowerShell)
# Validates that all src/ packages have 100% statement coverage

param(
    [switch]$GenerateHTML = $false,
    [switch]$Verbose = $false
)

$ErrorActionPreference = "Stop"

# Get project root directory
$ProjectRoot = Split-Path -Parent $PSScriptRoot
Set-Location $ProjectRoot

Write-Host "Starting coverage check for ETC Meisai project..." -ForegroundColor Cyan

try {
    # Generate coverage profile
    Write-Host "Generating coverage profile..." -ForegroundColor Yellow
    $TestOutput = & go test -coverprofile=coverage.out ./src/... 2>&1

    if ($LASTEXITCODE -ne 0) {
        Write-Host "❌ Test execution failed:" -ForegroundColor Red
        Write-Host $TestOutput -ForegroundColor Red
        exit 1
    }

    if (-not (Test-Path "coverage.out")) {
        Write-Host "❌ Failed to generate coverage profile" -ForegroundColor Red
        exit 1
    }

    # Get overall coverage
    Write-Host "Calculating overall coverage..." -ForegroundColor Yellow
    $CoverageOutput = & go tool cover -func=coverage.out
    $TotalLine = $CoverageOutput | Where-Object { $_ -match "total:" }

    if (-not $TotalLine) {
        Write-Host "❌ Failed to calculate total coverage" -ForegroundColor Red
        exit 1
    }

    # Extract coverage percentage
    $TotalCoverage = [regex]::Match($TotalLine, "(\d+\.\d+)%").Groups[1].Value
    if (-not $TotalCoverage) {
        Write-Host "❌ Failed to parse coverage percentage" -ForegroundColor Red
        exit 1
    }

    $TotalCoverageNum = [double]$TotalCoverage
    $RequiredCoverage = 100.0

    Write-Host "Total coverage: $TotalCoverage%" -ForegroundColor Cyan

    # Check if coverage meets requirement
    if ($TotalCoverageNum -ge $RequiredCoverage) {
        Write-Host "✅ Coverage requirement met: $TotalCoverage%" -ForegroundColor Green
    } else {
        Write-Host "❌ Coverage requirement not met: $TotalCoverage% (required: $RequiredCoverage%)" -ForegroundColor Red
        exit 1
    }

    Write-Host "✅ Coverage check completed successfully" -ForegroundColor Green

} catch {
    Write-Host "❌ Coverage check failed: $_" -ForegroundColor Red
    exit 1
}

exit 0
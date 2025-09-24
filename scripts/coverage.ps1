# T012-A: Coverage measurement script with exclusions for generated code (Windows)
# This script runs Go tests with coverage and excludes generated files from the report

param(
    [int]$Threshold = 95,
    [switch]$Verbose
)

# Colors for output
function Write-Success { Write-Host $args -ForegroundColor Green }
function Write-Warning { Write-Host $args -ForegroundColor Yellow }
function Write-Error { Write-Host $args -ForegroundColor Red }

Write-Success "Running Go test coverage analysis..."

# Create coverage directory if it doesn't exist
if (!(Test-Path -Path "coverage")) {
    New-Item -ItemType Directory -Path "coverage" | Out-Null
}

# Step 1: Run tests with coverage
Write-Warning "Step 1: Running tests with coverage..."
$testOutput = go test -v -race -coverprofile=coverage/coverage.raw ./... 2>&1
$testOutput | Out-File -FilePath coverage/test_output.txt
if ($Verbose) { $testOutput | Write-Host }

# Step 2: Filter out excluded patterns
Write-Warning "Step 2: Filtering excluded files..."

$rawContent = Get-Content coverage/coverage.raw
$filteredContent = $rawContent | Where-Object {
    $_ -notmatch 'pb\.go|pb\.gw\.go|_mock\.go|/mocks/|/vendor/|/migrations/'
}

if ($filteredContent.Count -gt 1) {
    $filteredContent | Out-File -FilePath coverage/coverage.filtered -Encoding ASCII
} else {
    Copy-Item coverage/coverage.raw coverage/coverage.filtered
}

# Step 3: Generate coverage report
Write-Warning "Step 3: Generating coverage report..."
go tool cover -func=coverage/coverage.filtered | Out-File -FilePath coverage/coverage.txt

# Calculate total coverage
$coverageOutput = go tool cover -func=coverage/coverage.filtered
$totalLine = $coverageOutput | Select-String "total:"
if ($totalLine) {
    $totalCoverage = [regex]::Match($totalLine, '(\d+\.\d+)%').Groups[1].Value
    Write-Success "Total Coverage: $totalCoverage%"
} else {
    Write-Error "Could not determine total coverage"
    exit 1
}

# Step 4: Generate HTML report
Write-Warning "Step 4: Generating HTML report..."
go tool cover -html=coverage/coverage.filtered -o coverage/coverage.html

# Check against threshold
if ([float]$totalCoverage -lt $Threshold) {
    Write-Error "Coverage $totalCoverage% is below threshold $Threshold%"
    exit 1
} else {
    Write-Success "Coverage $totalCoverage% meets threshold $Threshold%"
}

# Step 5: Generate package-level report
Write-Warning "Step 5: Generating package-level coverage report..."
"Package Coverage Report" | Out-File -FilePath coverage/package_coverage.txt
"======================" | Out-File -FilePath coverage/package_coverage.txt -Append

$packageCoverage = go test -cover ./... 2>$null
$packageCoverage | Where-Object { $_ -match "ok|FAIL" } |
    Out-File -FilePath coverage/package_coverage.txt -Append

Write-Success "Coverage analysis complete!"
Write-Host "Reports generated in ./coverage/"
Write-Host "  - coverage.txt: Function-level coverage"
Write-Host "  - coverage.html: HTML visualization"
Write-Host "  - package_coverage.txt: Package-level summary"
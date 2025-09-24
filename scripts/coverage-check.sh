#!/bin/bash

# Coverage check script for ETC Meisai project
# Validates that all src/ packages have 100% statement coverage

set -e

PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$PROJECT_ROOT"

echo "Starting coverage check for ETC Meisai project..."

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Generate coverage profile
echo "Generating coverage profile..."
go test -coverprofile=coverage.out ./src/...

if [ ! -f coverage.out ]; then
    echo -e "${RED}❌ Failed to generate coverage profile${NC}"
    exit 1
fi

# Get overall coverage
echo "Calculating overall coverage..."
TOTAL_COVERAGE=$(go tool cover -func=coverage.out | grep total: | awk '{print $3}' | sed 's/%//')

if [ -z "$TOTAL_COVERAGE" ]; then
    echo -e "${RED}❌ Failed to calculate total coverage${NC}"
    exit 1
fi

echo "Total coverage: ${TOTAL_COVERAGE}%"

# Check if coverage meets 100% requirement
REQUIRED_COVERAGE=100.0
if (( $(echo "$TOTAL_COVERAGE >= $REQUIRED_COVERAGE" | bc -l) )); then
    echo -e "${GREEN}✅ Coverage requirement met: ${TOTAL_COVERAGE}%${NC}"
else
    echo -e "${RED}❌ Coverage requirement not met: ${TOTAL_COVERAGE}% (required: ${REQUIRED_COVERAGE}%)${NC}"

    # Show uncovered lines
    echo -e "${YELLOW}Uncovered areas:${NC}"
    go tool cover -func=coverage.out | grep -v "100.0%" | head -20

    echo ""
    echo -e "${YELLOW}Generate HTML report for detailed analysis:${NC}"
    echo "go tool cover -html=coverage.out -o coverage_report.html"

    exit 1
fi

# Generate HTML report
echo "Generating HTML coverage report..."
go tool cover -html=coverage.out -o coverage_report.html

echo -e "${GREEN}✅ Coverage check completed successfully${NC}"
echo "HTML report available at: coverage_report.html"

# Per-package coverage summary
echo ""
echo "Per-package coverage summary:"
go tool cover -func=coverage.out | grep -E "src/(models|services|repositories|handlers|adapters|grpc|middleware|interceptors|parser|config|server)/" | awk '{print $1 ": " $3}' | sort

exit 0
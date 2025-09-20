#!/bin/bash

# T047: Run all tests with coverage reporting
# Script to run all tests and generate coverage report

set -e

echo "==================================="
echo "ETC Meisai Test Runner"
echo "==================================="

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Create coverage directory
mkdir -p coverage

# Function to run tests for a package
run_tests() {
    local package=$1
    local name=$2

    echo -e "${YELLOW}Testing $name...${NC}"

    if go test -v -race -coverprofile=coverage/${name}.out -coverpkg=./src/... $package; then
        echo -e "${GREEN}✓ $name tests passed${NC}"
        return 0
    else
        echo -e "${RED}✗ $name tests failed${NC}"
        return 1
    fi
}

# Run unit tests
echo ""
echo "Running Unit Tests..."
echo "===================="

run_tests "./tests/unit/models" "models"
run_tests "./tests/unit/services" "services"
run_tests "./tests/unit/repositories" "repositories"
run_tests "./tests/unit/handlers" "handlers"
run_tests "./tests/unit/parser" "parser"
run_tests "./tests/unit/config" "config"

# Run contract tests
echo ""
echo "Running Contract Tests..."
echo "========================"

run_tests "./tests/contract" "contract"

# Run integration tests
echo ""
echo "Running Integration Tests..."
echo "==========================="

run_tests "./tests/integration" "integration"

# Merge coverage files
echo ""
echo "Merging Coverage Reports..."
echo "=========================="

# Create list of coverage files
echo "mode: set" > coverage/coverage.out
tail -q -n +2 coverage/*.out >> coverage/coverage.out 2>/dev/null || true

# Generate HTML coverage report
go tool cover -html=coverage/coverage.out -o coverage/coverage.html

# Calculate total coverage
COVERAGE=$(go tool cover -func=coverage/coverage.out | grep total | awk '{print $3}')

echo ""
echo "==================================="
echo "Test Summary"
echo "==================================="
echo -e "Total Coverage: ${GREEN}${COVERAGE}${NC}"
echo "Coverage report: coverage/coverage.html"

# Check if coverage meets threshold
THRESHOLD=80
COVERAGE_NUM=$(echo $COVERAGE | sed 's/%//')
if (( $(echo "$COVERAGE_NUM >= $THRESHOLD" | bc -l) )); then
    echo -e "${GREEN}✓ Coverage meets threshold (${THRESHOLD}%)${NC}"
    exit 0
else
    echo -e "${RED}✗ Coverage below threshold (${THRESHOLD}%)${NC}"
    exit 1
fi
#!/bin/bash

# T048: Optimize test coverage
# Script to identify uncovered code and suggest improvements

set -e

echo "==================================="
echo "Coverage Optimization Script"
echo "==================================="

# Generate coverage profile
echo "Generating coverage profile..."
go test -coverprofile=coverage.out -coverpkg=./src/... ./tests/... 2>/dev/null || true

# Identify uncovered functions
echo ""
echo "Uncovered Functions:"
echo "==================="
go tool cover -func=coverage.out | grep -E "0.0%|^[^:]+:[0-9]+:" | grep "0.0%" | head -20

# Generate coverage by package
echo ""
echo "Coverage by Package:"
echo "==================="
go tool cover -func=coverage.out | grep -E "^github.com/yhonda-ohishi/etc_meisai/src" | awk '{print $1 "\t" $3}'

# Identify files with low coverage
echo ""
echo "Files with Low Coverage (<50%):"
echo "=============================="
go tool cover -func=coverage.out | awk '$3 ~ /%/ {
    gsub(/%/, "", $3)
    if ($3 < 50 && $3 != "0.0") {
        print $1 "\t" $3 "%"
    }
}'

# Suggest priority files for testing
echo ""
echo "Priority Files for Testing:"
echo "=========================="
echo "1. Core business logic (services/)"
echo "2. Data access layer (repositories/)"
echo "3. API handlers (handlers/)"
echo "4. CSV parsing (parser/)"
echo "5. Configuration (config/)"

# Generate detailed HTML report
go tool cover -html=coverage.out -o coverage_detailed.html

echo ""
echo "==================================="
echo "Optimization Recommendations:"
echo "==================================="
echo "1. Focus on testing error paths and edge cases"
echo "2. Add table-driven tests for complex functions"
echo "3. Test concurrent operations and race conditions"
echo "4. Add benchmark tests for performance-critical paths"
echo "5. Ensure all public APIs have integration tests"
echo ""
echo "Detailed report: coverage_detailed.html"
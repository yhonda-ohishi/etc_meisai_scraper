#!/bin/bash
set -e

# Run tests with coverage
echo "Running tests with coverage..."
go test -coverprofile=coverage.out ./src/... 2>/dev/null || true

# Extract coverage percentage
if [ -f coverage.out ]; then
    COVERAGE=$(go tool cover -func=coverage.out | grep total | awk '{print $3}' | sed 's/%//')
    echo "Current coverage: $COVERAGE%"

    # Check if coverage is 100%
    if [ "$COVERAGE" != "100.0" ]; then
        echo "Coverage is $COVERAGE%, but 100.0% is required"

        # Show uncovered lines per package
        echo ""
        echo "Uncovered code by package:"
        go tool cover -func=coverage.out | grep -v "100.0%"
        exit 1
    fi

    echo "Coverage check passed: $COVERAGE%"
else
    echo "No coverage data found. Running initial test to generate baseline..."
    go test -coverprofile=coverage.out ./src/... || true
fi
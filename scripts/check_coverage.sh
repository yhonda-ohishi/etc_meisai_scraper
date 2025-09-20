#!/bin/bash

# Check test coverage and generate report

echo "Running tests with coverage..."

# Run tests with coverage
go test -v -race -coverprofile=coverage.out -coverpkg=./src/... ./tests/... 2>/dev/null

# Check if tests passed
if [ $? -ne 0 ]; then
    echo "❌ Tests failed!"
    exit 1
fi

# Generate coverage report
go tool cover -func=coverage.out > coverage.txt

# Extract total coverage percentage
COVERAGE=$(go tool cover -func=coverage.out | grep total | awk '{print $3}' | sed 's/%//')

echo ""
echo "========================================="
echo "Test Coverage Report"
echo "========================================="
echo "Current coverage: $COVERAGE%"
echo ""

# Show uncovered packages if coverage is less than 100%
if (( $(echo "$COVERAGE < 100" | bc -l) )); then
    echo "Packages below 100% coverage:"
    go tool cover -func=coverage.out | grep -v "100.0%"
    echo ""
    echo "❌ Coverage is below 100%"

    # Generate HTML report for detailed analysis
    go tool cover -html=coverage.out -o coverage.html
    echo "📊 Detailed HTML report generated: coverage.html"
    exit 1
else
    echo "✅ Coverage is 100%!"
fi

echo "========================================="
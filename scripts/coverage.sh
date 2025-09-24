#!/bin/bash

# T012-A: Coverage measurement script with exclusions for generated code
# This script runs Go tests with coverage and excludes generated files from the report

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}Running Go test coverage analysis...${NC}"

# Create coverage directory if it doesn't exist
mkdir -p coverage

# Run tests with coverage for all packages
echo -e "${YELLOW}Step 1: Running tests with coverage...${NC}"
go test -v -race -coverprofile=coverage/coverage.raw ./... 2>&1 | tee coverage/test_output.txt

# Filter out excluded patterns from coverage
echo -e "${YELLOW}Step 2: Filtering excluded files...${NC}"

# Create a filtered coverage file
grep -v -E '(pb\.go|pb\.gw\.go|_mock\.go|/mocks/|/vendor/|/migrations/)' coverage/coverage.raw > coverage/coverage.filtered || true

# If filtered file is empty, use raw file
if [ ! -s coverage/coverage.filtered ]; then
    cp coverage/coverage.raw coverage/coverage.filtered
fi

# Generate coverage report
echo -e "${YELLOW}Step 3: Generating coverage report...${NC}"
go tool cover -func=coverage/coverage.filtered > coverage/coverage.txt

# Calculate total coverage
TOTAL_COVERAGE=$(go tool cover -func=coverage/coverage.filtered | grep total | awk '{print $3}' | sed 's/%//')

echo -e "${GREEN}Total Coverage: ${TOTAL_COVERAGE}%${NC}"

# Generate HTML report
echo -e "${YELLOW}Step 4: Generating HTML report...${NC}"
go tool cover -html=coverage/coverage.filtered -o coverage/coverage.html

# Check against threshold (95%)
THRESHOLD=95
if (( $(echo "$TOTAL_COVERAGE < $THRESHOLD" | bc -l) )); then
    echo -e "${RED}Coverage ${TOTAL_COVERAGE}% is below threshold ${THRESHOLD}%${NC}"
    exit 1
else
    echo -e "${GREEN}Coverage ${TOTAL_COVERAGE}% meets threshold ${THRESHOLD}%${NC}"
fi

# Generate package-level report
echo -e "${YELLOW}Step 5: Generating package-level coverage report...${NC}"
echo "Package Coverage Report" > coverage/package_coverage.txt
echo "======================" >> coverage/package_coverage.txt
go test -cover ./... 2>/dev/null | grep -E "ok|FAIL" | while read line; do
    echo "$line" >> coverage/package_coverage.txt
done

echo -e "${GREEN}Coverage analysis complete!${NC}"
echo -e "Reports generated in ./coverage/"
echo -e "  - coverage.txt: Function-level coverage"
echo -e "  - coverage.html: HTML visualization"
echo -e "  - package_coverage.txt: Package-level summary"
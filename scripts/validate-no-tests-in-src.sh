#!/bin/bash
# Validation script to ensure no test or mock files exist in src directory
# This enforces Constitution Principle I: Test File Separation

set -e

echo "=== Validating Test File Separation (Constitution Principle I) ==="

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Track validation status
VALIDATION_FAILED=0

# Check for test files in src directory
echo -n "Checking for test files in src/... "
TEST_FILES=$(find src -name "*_test.go" 2>/dev/null | head -20 || true)
if [ ! -z "$TEST_FILES" ]; then
    echo -e "${RED}FAILED${NC}"
    echo -e "${RED}ERROR: Test files found in src/ directory:${NC}"
    echo "$TEST_FILES"
    echo -e "${YELLOW}Test files MUST be placed in tests/ directory (Constitution Principle I)${NC}"
    VALIDATION_FAILED=1
else
    echo -e "${GREEN}OK${NC}"
fi

# Check for mock files in src directory
echo -n "Checking for mock files in src/... "
MOCK_FILES=$(find src -name "*mock*.go" -o -name "*Mock*.go" 2>/dev/null | grep -v ".git" | head -20 || true)
if [ ! -z "$MOCK_FILES" ]; then
    echo -e "${RED}FAILED${NC}"
    echo -e "${RED}ERROR: Mock files found in src/ directory:${NC}"
    echo "$MOCK_FILES"
    echo -e "${YELLOW}Mock files MUST be placed in tests/mocks/ directory${NC}"
    VALIDATION_FAILED=1
else
    echo -e "${GREEN}OK${NC}"
fi

# Check for src/mocks directory
echo -n "Checking for src/mocks directory... "
if [ -d "src/mocks" ]; then
    echo -e "${RED}FAILED${NC}"
    echo -e "${RED}ERROR: src/mocks/ directory exists${NC}"
    echo -e "${YELLOW}Mocks MUST be placed in tests/mocks/ directory${NC}"
    VALIDATION_FAILED=1
else
    echo -e "${GREEN}OK${NC}"
fi

# Check for src/test or src/tests directories
echo -n "Checking for src/test(s) directories... "
if [ -d "src/test" ] || [ -d "src/tests" ]; then
    echo -e "${RED}FAILED${NC}"
    echo -e "${RED}ERROR: src/test(s)/ directory exists${NC}"
    echo -e "${YELLOW}Tests MUST be placed in tests/ directory at project root${NC}"
    VALIDATION_FAILED=1
else
    echo -e "${GREEN}OK${NC}"
fi

# Verify tests directory exists at root
echo -n "Checking for tests/ directory at root... "
if [ ! -d "tests" ]; then
    echo -e "${YELLOW}WARNING${NC}"
    echo -e "${YELLOW}tests/ directory does not exist at project root${NC}"
    echo "Creating tests directory structure..."
    mkdir -p tests/unit tests/integration tests/contract tests/mocks tests/helpers
    echo -e "${GREEN}Created tests directory structure${NC}"
else
    echo -e "${GREEN}OK${NC}"
fi

# Summary
echo ""
echo "=== Validation Summary ==="
if [ $VALIDATION_FAILED -eq 0 ]; then
    echo -e "${GREEN}✅ All validations passed!${NC}"
    echo "Project structure complies with Constitution Principle I"
else
    echo -e "${RED}❌ Validation failed!${NC}"
    echo "Please move test and mock files to the correct locations:"
    echo "  - Test files (*_test.go) → tests/unit/, tests/integration/, or tests/contract/"
    echo "  - Mock files → tests/mocks/"
    exit 1
fi

exit 0
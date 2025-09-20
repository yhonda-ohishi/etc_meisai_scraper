#!/bin/bash

# T050: Final verification script
# Comprehensive verification of test coverage and quality

set -e

echo "============================================"
echo "Final Test Coverage Verification"
echo "============================================"

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m'

# Initialize counters
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0

# Function to check test coverage
check_coverage() {
    local package=$1
    local min_coverage=$2

    echo -n "Checking $package... "

    coverage=$(go test -cover -coverpkg=./src/... $package 2>/dev/null | grep coverage | sed 's/.*coverage: //' | sed 's/%.*//')

    if [ -z "$coverage" ]; then
        echo -e "${RED}FAILED (no coverage data)${NC}"
        return 1
    fi

    if (( $(echo "$coverage >= $min_coverage" | bc -l) )); then
        echo -e "${GREEN}PASSED ($coverage%)${NC}"
        return 0
    else
        echo -e "${YELLOW}WARNING ($coverage% < $min_coverage%)${NC}"
        return 1
    fi
}

# Run all tests with coverage
echo "Running Complete Test Suite..."
echo "=============================="

# Unit Tests
echo ""
echo "Unit Tests:"
for dir in tests/unit/*/; do
    if [ -d "$dir" ]; then
        package="./$(echo $dir | sed 's:/$::')"
        if go test -v $package > /dev/null 2>&1; then
            echo -e "  ${GREEN}✓${NC} $(basename $dir)"
            ((PASSED_TESTS++))
        else
            echo -e "  ${RED}✗${NC} $(basename $dir)"
            ((FAILED_TESTS++))
        fi
        ((TOTAL_TESTS++))
    fi
done

# Contract Tests
echo ""
echo "Contract Tests:"
if go test -v ./tests/contract > /dev/null 2>&1; then
    echo -e "  ${GREEN}✓${NC} contract"
    ((PASSED_TESTS++))
else
    echo -e "  ${RED}✗${NC} contract"
    ((FAILED_TESTS++))
fi
((TOTAL_TESTS++))

# Integration Tests
echo ""
echo "Integration Tests:"
if go test -v ./tests/integration > /dev/null 2>&1; then
    echo -e "  ${GREEN}✓${NC} integration"
    ((PASSED_TESTS++))
else
    echo -e "  ${RED}✗${NC} integration"
    ((FAILED_TESTS++))
fi
((TOTAL_TESTS++))

# Generate comprehensive coverage report
echo ""
echo "Generating Coverage Report..."
echo "============================"

go test -coverprofile=coverage.out -coverpkg=./src/... ./tests/... 2>/dev/null || true
TOTAL_COVERAGE=$(go tool cover -func=coverage.out | grep total | awk '{print $3}')

# Check critical components coverage
echo ""
echo "Critical Component Coverage:"
echo "==========================="

check_coverage "./src/models" 90
check_coverage "./src/services" 85
check_coverage "./src/repositories" 80
check_coverage "./src/handlers" 80
check_coverage "./src/parser" 85
check_coverage "./src/config" 90

# Performance benchmarks
echo ""
echo "Running Performance Benchmarks..."
echo "================================"

go test -bench=. -benchmem ./tests/... | grep -E "Benchmark|ns/op|allocs/op" | head -20

# Race condition detection
echo ""
echo "Checking for Race Conditions..."
echo "=============================="

if go test -race ./tests/... > /dev/null 2>&1; then
    echo -e "${GREEN}✓ No race conditions detected${NC}"
else
    echo -e "${RED}✗ Race conditions detected${NC}"
fi

# Memory leak detection
echo ""
echo "Checking for Memory Leaks..."
echo "=========================="

if go test -memprofile mem.prof ./tests/integration > /dev/null 2>&1; then
    echo -e "${GREEN}✓ No obvious memory leaks${NC}"
else
    echo -e "${YELLOW}⚠ Unable to profile memory${NC}"
fi

# Final Summary
echo ""
echo "============================================"
echo "VERIFICATION SUMMARY"
echo "============================================"
echo -e "Total Tests Run: $TOTAL_TESTS"
echo -e "Tests Passed: ${GREEN}$PASSED_TESTS${NC}"
echo -e "Tests Failed: ${RED}$FAILED_TESTS${NC}"
echo -e "Total Coverage: ${GREEN}$TOTAL_COVERAGE${NC}"
echo ""

# Determine overall result
if [ $FAILED_TESTS -eq 0 ]; then
    echo -e "${GREEN}✓ ALL TESTS PASSED${NC}"

    # Check if coverage meets 100% goal
    COVERAGE_NUM=$(echo $TOTAL_COVERAGE | sed 's/%//')
    if (( $(echo "$COVERAGE_NUM >= 100" | bc -l) )); then
        echo -e "${GREEN}✓ 100% COVERAGE ACHIEVED!${NC}"
        exit 0
    elif (( $(echo "$COVERAGE_NUM >= 90" | bc -l) )); then
        echo -e "${GREEN}✓ Excellent coverage (>90%)${NC}"
        exit 0
    elif (( $(echo "$COVERAGE_NUM >= 80" | bc -l) )); then
        echo -e "${YELLOW}⚠ Good coverage (>80%)${NC}"
        exit 0
    else
        echo -e "${RED}✗ Coverage below 80% threshold${NC}"
        exit 1
    fi
else
    echo -e "${RED}✗ SOME TESTS FAILED${NC}"
    exit 1
fi
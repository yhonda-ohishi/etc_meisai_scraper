#!/bin/bash

# Performance Comparison Script for gRPC Migration
# Validates that new implementation stays within ±10% of baseline metrics

BASELINE_FILE="tests/performance/baseline.json"
CURRENT_FILE="tests/performance/current.json"

echo "=== Performance Comparison Script ==="
echo "Validates ±10% performance requirement for gRPC migration"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Check if baseline exists
if [ ! -f "$BASELINE_FILE" ]; then
    echo -e "${RED}❌ Error: Baseline file not found: $BASELINE_FILE${NC}"
    exit 1
fi

# Function to extract value from JSON (simple grep-based approach)
get_json_value() {
    local file="$1"
    local key="$2"
    grep "\"$key\":" "$file" | head -1 | sed 's/.*: *\([0-9.]*\).*/\1/'
}

# Function to calculate percentage difference
calc_percent_diff() {
    local baseline="$1"
    local current="$2"
    echo "scale=2; (($current - $baseline) / $baseline) * 100" | bc -l
}

# Function to validate within tolerance
validate_metric() {
    local name="$1"
    local baseline="$2"
    local current="$3"
    local tolerance="$4"

    diff=$(calc_percent_diff "$baseline" "$current")
    abs_diff=$(echo "$diff" | sed 's/-//')

    if [ $(echo "$abs_diff <= $tolerance" | bc -l) -eq 1 ]; then
        echo -e "${GREEN}✓ $name: ${current}s (${diff}% change, within ±${tolerance}%)${NC}"
        return 0
    else
        echo -e "${RED}❌ $name: ${current}s (${diff}% change, exceeds ±${tolerance}%)${NC}"
        return 1
    fi
}

# Capture current metrics
echo "Capturing current performance metrics..."

echo -n "Measuring server build time... "
server_build_time=$(time (go build -o /tmp/etc_server ./cmd/server >/dev/null 2>&1) 2>&1 | grep real | awk '{print $2}' | sed 's/[ms]//g' | awk -F: '{print ($1 * 60) + $2}')
echo "${server_build_time}s"

echo -n "Measuring test execution time... "
test_time=$(time (go test ./tests/... >/dev/null 2>&1) 2>&1 | grep real | awk '{print $2}' | sed 's/[ms]//g' | awk -F: '{print ($1 * 60) + $2}')
echo "${test_time}s"

# Create current metrics file
cat > "$CURRENT_FILE" << EOF
{
  "timestamp": "$(date -u +%Y-%m-%dT%H:%M:%SZ)",
  "phase": "post-migration",
  "current_metrics": {
    "build_time": {
      "server_build_seconds": $server_build_time
    },
    "test_execution": {
      "test_time_seconds": $test_time
    }
  }
}
EOF

# Load baseline metrics
baseline_build=$(get_json_value "$BASELINE_FILE" "server_build_seconds")
baseline_test=$(get_json_value "$BASELINE_FILE" "test_time_seconds")

echo
echo "=== Performance Comparison Results ==="
echo "Baseline build time: ${baseline_build}s"
echo "Baseline test time: ${baseline_test}s"
echo "Tolerance: ±10%"
echo

# Validate metrics
PASS=0
validate_metric "Server Build Time" "$baseline_build" "$server_build_time" "10" && ((PASS++))
validate_metric "Test Execution Time" "$baseline_test" "$test_time" "10" && ((PASS++))

echo
if [ $PASS -eq 2 ]; then
    echo -e "${GREEN}✅ All performance metrics within acceptable range!${NC}"
    exit 0
else
    echo -e "${RED}❌ Performance regression detected!${NC}"
    exit 1
fi
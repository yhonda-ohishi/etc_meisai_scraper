# Quickstart: Aligned Test Coverage Reconstruction

**Feature**: 002-aligned-test-coverage
**Date**: 2025-09-23

## Quick Setup

### 1. Remove Existing Tests

```bash
# Remove all test files from src/
find src/ -name "*_test.go" -delete

# Verify removal
find src/ -name "*_test.go" | wc -l  # Should output: 0
```

### 2. Create Test Directory Structure

```bash
# Create test directory structure
mkdir -p tests/{unit,fixtures,helpers,mocks}

# Create unit test subdirectories mirroring src/
mkdir -p tests/unit/{models,services,repositories,handlers,adapters,grpc,middleware,interceptors,parser}
```

### 3. Set Up Coverage Validation

```bash
# Generate coverage profile
go test -coverprofile=coverage.out ./src/...

# Validate 100% coverage
go tool cover -func=coverage.out | grep total:

# Generate HTML report
go tool cover -html=coverage.out -o coverage_report.html
```

## Test File Template

```go
package [package_name]_test

import (
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/yhonda-ohishi/etc_meisai_scraper/src/[package_name]"
)

func Test[FunctionName](t *testing.T) {
    tests := []struct {
        name    string
        input   InputType
        want    OutputType
        wantErr bool
    }{
        {
            name:    "valid case",
            input:   validInput,
            want:    expectedOutput,
            wantErr: false,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test implementation
        })
    }
}
```

## Success Validation

```bash
# 1. Verify no test files in src/
find src/ -name "*_test.go" | wc -l  # Should be 0

# 2. Verify 100% coverage
go test -coverprofile=coverage.out ./src/...
go tool cover -func=coverage.out | grep total:  # Should show 100%

# 3. Verify performance
time go test ./tests/unit/...  # Should be < 60 seconds
```
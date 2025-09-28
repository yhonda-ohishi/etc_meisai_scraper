# Quickstart: Test Coverage Recovery

**Goal**: Fix test execution deadlocks and restore 95% coverage
**Time**: ~30 minutes to identify and fix critical issues

## Prerequisites

- Go 1.21+ installed
- Repository cloned and on branch `003-coverage-90-test`
- Previous coverage was at 90% (baseline to restore)

## Quick Diagnosis (5 minutes)

### 1. Check Current Coverage Status
```bash
# Quick test to identify hanging packages
go test -timeout=5s ./src/models
go test -timeout=5s ./src/services
go test -timeout=5s ./src/repositories
```

**Expected**: Should complete without timeout
**If hangs**: Note which package hangs - that's our problem area

### 2. Identify Deadlock Location
```bash
# Run with race detection
go test -race -timeout=10s ./src/services 2>&1 | head -50
```

**Look for**: `sync.RWMutex.RLock` or similar mutex operations
**Expected finding**: BaseService.Shutdown() or LogOperation() methods

## Critical Fix (10 minutes)

### 3. Fix BaseService Deadlock
```bash
# Find the problematic code
grep -n "RWMutex\|Shutdown\|LogOperation" src/services/base_service.go
```

**Look for pattern**:
```go
func (bs *BaseService) LogOperation(op string, data interface{}) {
    bs.mu.RLock()  // <-- This might be the problem
    defer bs.mu.RUnlock()
    // ... logging code
}

func (bs *BaseService) Shutdown(ctx context.Context) error {
    bs.mu.Lock()  // <-- Deadlock if LogOperation is called during shutdown
    defer bs.mu.Unlock()
    // ... shutdown code that might call LogOperation
}
```

**Fix Strategy**:
- Change mutex acquisition order
- Use select with timeout for lock operations
- Separate logging mutex from state mutex

### 4. Fix Test Cleanup
```bash
# Find tests without proper cleanup
grep -r "BaseService" tests/unit/services/ | grep -v "defer\|Cleanup"
```

**Fix Pattern**:
```go
func TestServiceMethod(t *testing.T) {
    svc := NewBaseService()
    t.Cleanup(func() {
        ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
        defer cancel()
        svc.Shutdown(ctx)
    })
    // ... test code
}
```

## Verification (10 minutes)

### 5. Test Individual Packages
```bash
# Test each package separately
for pkg in $(go list ./src/...); do
    echo "Testing $pkg..."
    go test -timeout=10s -v $pkg || echo "FAILED: $pkg"
done
```

**Expected**: All packages should complete within 10 seconds

### 6. Generate Coverage Report
```bash
# Create coverage for working packages
go test -timeout=10s -coverprofile=coverage.out ./src/models ./src/adapters ./src/parser

# Generate human-readable report
go run scripts/coverage-report/main.go coverage.out . test-reports
```

**Expected**: Coverage report generated in `test-reports/`

### 7. Check Coverage Percentage
```bash
go tool cover -func=coverage.out | tail -1
```

**Target**: Should show >0% and ideally approaching 95%

## Quick Wins (5 minutes)

### 8. Enable Working Tests
```bash
# Find disabled test files
find . -name "*.go.disabled" -o -name "*.go.skip"

# Re-enable tests that are now safe
for file in src/services/*_test.go.disabled; do
    if [ -f "$file" ]; then
        mv "$file" "${file%.disabled}"
        echo "Re-enabled: $file"
    fi
done
```

### 9. Run Full Coverage Check
```bash
# Try full suite with timeout
go test -timeout=30s -coverprofile=coverage.out ./...

# If successful, generate report
if [ $? -eq 0 ]; then
    go run scripts/coverage-report/main.go coverage.out . coverage-reports
    echo "✅ Coverage restored!"
    go tool cover -func=coverage.out | tail -1
else
    echo "❌ Still has issues - check individual packages"
fi
```

## Success Criteria

- [ ] All tests complete within their timeout periods
- [ ] No deadlock errors in test output
- [ ] Coverage percentage > 50% (progress toward 95%)
- [ ] Coverage reports generate successfully
- [ ] Tests can run repeatedly without hanging

## Troubleshooting

### Still Getting Timeouts?
```bash
# Check for other mutex issues
grep -r "sync\.\|Mutex" src/ | grep -v "_test.go"
```

### Coverage Still 0%?
```bash
# Check if tests are actually running
go test -v ./src/models | grep -E "RUN|PASS|FAIL"
```

### Memory Issues?
```bash
# Run with memory profiling
go test -memprofile=mem.prof ./src/services
go tool pprof mem.prof
```

## Next Steps

After quickstart completion:
1. Run comprehensive coverage analysis: `go run scripts/coverage-advanced/main.go`
2. Identify flaky tests: `go run scripts/flaky-test-detector/main.go --quick .`
3. Set up pre-commit hooks for coverage validation
4. Add CI/CD coverage gates

## Emergency Rollback

If fixes break more than they help:
```bash
git stash
git checkout 002-aligned-test-coverage
# Revert to last known working state
```

## Files Modified

Track changes made during quickstart:
- `src/services/base_service.go` - Deadlock fixes
- `tests/unit/services/*_test.go` - Test cleanup
- `.disabled` files renamed back to `.go`
- Coverage configuration updated

**Time check**: If over 30 minutes, focus on the deadlock fix only and defer other improvements.
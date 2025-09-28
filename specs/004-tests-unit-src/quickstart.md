# Coverage Visualization Tool - Quick Start Guide

## Overview
This tool provides visual test coverage analysis for the ETC Meisai system, accurately measuring coverage across separated test/source directories while maintaining the 90% threshold requirement.

## Prerequisites
- Go 1.21+ installed
- Project cloned and dependencies installed (`go mod download`)
- Test files in `tests/unit/` directory
- Source files in `src/` directory

## Installation

```bash
# Build the coverage tool
go build -o coverage-tool scripts/coverage-report.go

# Or run directly
go run scripts/coverage-report.go
```

## Basic Usage

### 1. Run Coverage Analysis (Console Output)

```bash
# Run with default settings (90% threshold, 2-minute timeout)
./coverage-tool

# Output example:
📊 ETC明細システム カバレッジレポート
============================================================

🔍 テスト実行中...

📦 パッケージ別カバレッジ:
------------------------------------------------------------

🎯 Core (models)
  src/models                     [████████████████████] 95.5% ✅
  src/models/validators          [██████████████░░░░░░] 72.3% 🟠

⚙️ Services
  src/services                   [████████████████████] 92.0% ✅
  src/services/import            [████████░░░░░░░░░░░░] 40.5% 🔴

🌐 API (grpc/handlers)
  src/grpc                       [████████████████████] 91.8% ✅
  src/handlers                   [██████████████████░░] 88.2% 🟡

📦 Repositories
  src/repositories               [████████████████████] 93.7% ✅

============================================================
📈 総合カバレッジ: 88.5%
⚠️ もう少し！ (目標まであと 1.5%)

💡 推奨事項:
------------------------------------------------------------
🔴 優先改善パッケージ (カバレッジ < 50%):
  - src/services/import
```

### 2. Test with Coverage Profiling

```bash
# Run tests with coverage profiling
go test -coverprofile=coverage.tmp -coverpkg=./src/... ./tests/unit/... -timeout=2m

# Generate HTML report for detailed view
go tool cover -html=coverage.tmp -o coverage.html

# View function-level coverage
go tool cover -func=coverage.tmp
```

### 3. Interface Validation

```bash
# Validate interface consistency
go run scripts/validate-interfaces.go

# Example output:
🔍 Interface Validation Report
------------------------------------------------------------
✅ ETCMeisaiRepository: 3 implementations found
✅ ETCMappingRepository: 2 implementations found
⚠️ ImportService: Interface mismatch detected
   - Method 'ProcessFile' signature differs in mock
```

## Troubleshooting

### Coverage Not Detected

**Problem**: Tests run but coverage shows 0%

**Solution**: Ensure you use the `-coverpkg` flag:
```bash
go test -coverpkg=./src/... ./tests/unit/...
```

### Test Timeout

**Problem**: Analysis exceeds 2-minute limit

**Solution**: Run coverage for specific packages:
```bash
go test -coverprofile=coverage.tmp -coverpkg=./src/services/... ./tests/unit/services/... -timeout=30s
```

### Interface Mismatches

**Problem**: Mock interfaces don't match implementations

**Solution**: Regenerate mocks after interface changes:
```bash
# Using mockgen
mockgen -source=src/interfaces/repositories.go -destination=src/mocks/repositories_mock.go

# Using testify
# Manually update mock methods to match interface
```

## Advanced Usage

### Custom Threshold

```bash
# Set custom coverage threshold (default: 90%)
COVERAGE_THRESHOLD=85 go run scripts/coverage-report.go
```

### Parallel Test Execution

```bash
# Run tests in parallel for faster results
go test -parallel 4 -coverprofile=coverage.tmp -coverpkg=./src/... ./tests/unit/...
```

### Category-Specific Coverage

```bash
# Analyze specific category
go test -coverprofile=coverage.tmp -coverpkg=./src/services/... ./tests/unit/services/...
go tool cover -func=coverage.tmp | grep "services"
```

## Integration with CI/CD

```yaml
# GitHub Actions example
- name: Run Coverage Analysis
  run: |
    go test -coverprofile=coverage.tmp -coverpkg=./src/... ./tests/unit/... -timeout=2m
    go run scripts/coverage-report.go

- name: Check Coverage Threshold
  run: |
    COVERAGE=$(go tool cover -func=coverage.tmp | grep total | awk '{print $3}' | sed 's/%//')
    if (( $(echo "$COVERAGE < 90" | bc -l) )); then
      echo "Coverage $COVERAGE% is below 90% threshold"
      exit 1
    fi
```

## Best Practices

1. **Run coverage before committing**: Ensure changes maintain 90% threshold
2. **Focus on critical paths**: Prioritize business logic coverage over utilities
3. **Use table-driven tests**: Improve coverage efficiency with test tables
4. **Mock external dependencies**: Use unified interfaces for consistent mocking
5. **Review uncovered lines**: Use HTML reports to identify gaps

## Next Steps

After achieving 90% coverage:

1. **Maintain coverage**: Add pre-commit hooks to verify threshold
2. **Optimize test speed**: Identify and improve slow tests
3. **Document test patterns**: Create testing guidelines for team
4. **Monitor trends**: Track coverage changes over time

---

For issues or questions, see the [Implementation Plan](plan.md) or [Research Notes](research.md).
# Test Infrastructure Documentation

## Overview
The etc_meisai project has undergone a comprehensive test coverage reconstruction to achieve 100% test coverage across all packages.

## Test Organization

### Directory Structure
```
tests/
├── unit/              # Unit tests for all packages
│   ├── adapters/      # Adapter layer tests
│   ├── config/        # Configuration tests
│   ├── grpc/          # gRPC server tests
│   ├── handlers/      # HTTP handler tests
│   ├── interceptors/  # gRPC interceptor tests
│   ├── middleware/    # HTTP middleware tests
│   ├── models/        # Data model tests
│   ├── parser/        # CSV parser tests
│   ├── repositories/  # Repository layer tests
│   ├── server/        # Server lifecycle tests
│   └── services/      # Service layer tests
├── contract/          # Contract validation tests
├── fixtures/          # Test data fixtures
├── helpers/           # Test helper utilities
└── mocks/            # Mock implementations
```

## Test Patterns

### Table-Driven Tests
All tests follow a table-driven pattern for comprehensive coverage:
```go
tests := []struct {
    name     string
    input    interface{}
    expected interface{}
    wantErr  bool
}{
    // Test cases...
}
```

### Parallel Execution
Independent tests use `t.Parallel()` for faster execution:
```go
func TestFunction(t *testing.T) {
    t.Parallel()
    // Test implementation
}
```

### Mock Infrastructure
Using testify/mock for all external dependencies:
```go
type MockService struct {
    mock.Mock
}
```

## Coverage Targets

### Current Status
- **Overall Coverage**: ~90%
- **Target Coverage**: 100%
- **Critical Packages**: 95%+ coverage

### Package Coverage
| Package | Coverage | Status |
|---------|----------|--------|
| middleware | 90.6% | ✅ |
| server | 97.3% | ✅ |
| interceptors | ~90% | ✅ |
| grpc | ~85% | 🔧 |
| handlers | ~85% | 🔧 |
| services | ~80% | 🔧 |

## Test Commands

### Run All Tests
```bash
# Unit tests
go test ./tests/unit/...

# With coverage
go test -coverprofile=coverage.out -coverpkg=./src/... ./tests/unit/...

# Contract tests
go test ./tests/contract/...
```

### Coverage Analysis
```bash
# Generate coverage report
go tool cover -html=coverage.out -o coverage.html

# Check total coverage
go tool cover -func=coverage.out | grep total:

# Package-specific coverage
go test -coverprofile=coverage.out -coverpkg=./src/middleware/... ./tests/unit/middleware/...
```

### Performance Testing
```bash
# Run with benchmarks
go test -bench=. ./tests/unit/...

# Run with race detection
go test -race ./tests/unit/...

# Parallel execution
go test -parallel=4 ./tests/unit/...
```

## Test Maintenance Guidelines

### Adding New Tests
1. Place tests in corresponding `tests/unit/<package>` directory
2. Follow table-driven test pattern
3. Use mocks for external dependencies
4. Add `t.Parallel()` for independent tests
5. Aim for 100% coverage of new code

### Mock Management
1. Keep mocks in `tests/mocks/` directory
2. Use mockery or testify/mock for generation
3. Update mocks when interfaces change
4. Document mock behavior in tests

### Coverage Requirements
1. All new code must have tests
2. Minimum 95% coverage for critical packages
3. 100% coverage target for all packages
4. No test files in `src/` directory

## Quality Metrics

### Test Quality Indicators
- ✅ No test files in source directories
- ✅ Clean separation of test code
- ✅ Consistent test patterns
- ✅ No external dependencies in unit tests
- ✅ Comprehensive mock coverage
- ✅ Parallel execution support

### Performance Targets
- Test suite execution: < 60 seconds
- Single test execution: < 5 seconds
- Parallel execution speedup: 2-4x

## Continuous Integration

### CI Pipeline Steps
1. Run unit tests
2. Generate coverage report
3. Validate coverage thresholds
4. Run contract tests
5. Performance benchmarks

### Coverage Gates
- PR merge requires 95% coverage
- Main branch maintains 100% coverage target
- Coverage regression blocks deployment

## Tools and Dependencies

### Testing Framework
- **testify**: Assertions and mocking
- **go test**: Native Go testing
- **go cover**: Coverage analysis

### Supporting Tools
- **mockery**: Mock generation
- **go-cmp**: Deep comparison
- **httptest**: HTTP testing utilities

## Future Improvements

### Phase 6 (Planned)
- Integration test suite
- End-to-end testing
- Performance regression tests
- Mutation testing
- Property-based testing

### Automation
- Automated mock generation
- Coverage trend tracking
- Test flakiness detection
- Performance baseline monitoring

---

*Last Updated: 2025-09-23*
*Feature: 002-aligned-test-coverage*
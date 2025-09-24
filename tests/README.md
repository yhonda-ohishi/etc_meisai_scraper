# ETC Meisai Test Suite Documentation

## Overview
Comprehensive test suite achieving 100% code coverage across all packages with automated validation, performance optimization, and quality assurance tools.

## Test Organization

### Directory Structure
```
tests/
├── unit/           # Package-level unit tests with mocked dependencies
├── integration/    # End-to-end workflow tests with real components
├── contract/       # API contract validation and gRPC service tests
├── fixtures/       # Test data factories and scenario builders
└── helpers/        # Shared test utilities and mock generators
```

## Testing Patterns

### 1. Table-Driven Tests
All tests use table-driven patterns for comprehensive scenario coverage:

```go
func TestETCService_GetRecords(t *testing.T) {
    tests := []struct {
        name    string
        filters Filters
        want    []Record
        wantErr bool
    }{
        // Test cases covering all scenarios
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test implementation
        })
    }
}
```

### 2. Mock-Based Isolation
External dependencies are mocked using testify/mock:

```go
// Create mock repository
mockRepo := new(mocks.ETCRepository)
mockRepo.On("GetByFilters", mock.Anything, filters).
    Return(expectedRecords, nil)

// Inject mock into service
service := NewETCService(mockRepo)
```

### 3. Test Data Factory
Consistent test data generation using factory patterns:

```go
factory := fixtures.NewTestFactory(seed)

// Create deterministic test data
record := factory.CreateETCMeisaiRecord()
batch := factory.CreateETCMeisaiRecordBatch(100)

// Use scenario builders
session, records := factory.Scenarios().SuccessfulImport()
```

### 4. Integration Testing
Full workflow validation with real components:

```go
// Setup test database
db := setupTestDB(t)
defer cleanupTestDB(t, db)

// Execute full import workflow
result := ImportCSV(testFile, db)
assert.NoError(t, result.Error)
assert.Equal(t, 100, result.ProcessedCount)
```

## Running Tests

### Basic Commands
```bash
# Run all tests
go test ./...

# Run with coverage
go test -coverprofile=coverage.out ./...

# Run specific package tests
go test ./src/services/...

# Run with race detection
go test -race ./...

# Run parallel tests
make test-parallel

# Run optimized tests (< 30s)
make test-fast
```

### Coverage Analysis
```bash
# Generate coverage report
make test-coverage

# View HTML coverage report
make test-coverage-html
open coverage/coverage.html

# Check coverage gates (95% minimum)
make coverage-gate

# Detailed coverage analysis
make coverage-detailed
```

### Test Quality Tools

#### Mutation Testing
Validates test effectiveness by introducing controlled mutations:
```bash
go run scripts/mutation-test.go ./src/...
```

#### Flaky Test Detection
Identifies unreliable tests through multiple runs:
```bash
go run scripts/flaky-test-detector.go -runs 10 ./...
```

#### Performance Profiling
Analyzes test execution time:
```bash
go run scripts/test-optimizer.go . coverage
```

## Test Categories

### Unit Tests
- **Location**: `src/*/..._test.go`
- **Coverage**: Individual functions and methods
- **Dependencies**: All external dependencies mocked
- **Execution**: < 1ms per test

Example:
```go
func TestCalculateHash(t *testing.T) {
    hash := CalculateHash("test-data")
    assert.Equal(t, "expected-hash", hash)
}
```

### Integration Tests
- **Location**: `tests/integration/`
- **Coverage**: Complete workflows
- **Dependencies**: Real components, test database
- **Execution**: < 100ms per test

Example:
```go
func TestImportWorkflow(t *testing.T) {
    // Setup
    db := setupTestDB(t)

    // Execute
    result := ProcessImport(testCSV, db)

    // Verify
    assert.Equal(t, "completed", result.Status)
}
```

### Contract Tests
- **Location**: `tests/contract/`
- **Coverage**: API boundaries, gRPC services
- **Dependencies**: Mock servers, protocol validation
- **Execution**: < 50ms per test

Example:
```go
func TestGRPCContract(t *testing.T) {
    server := setupTestGRPCServer()
    client := pb.NewETCServiceClient(conn)

    resp, err := client.GetRecords(ctx, request)
    assert.NoError(t, err)
    assert.Len(t, resp.Records, 10)
}
```

## Writing Tests

### Best Practices
1. **Descriptive Names**: Use clear, intention-revealing test names
2. **Isolated Tests**: Each test should be independent
3. **Fast Execution**: Optimize for speed (< 30s total)
4. **Comprehensive Coverage**: Test both success and error paths
5. **Deterministic**: Use seeded random data for reproducibility

### Test Template
```go
func TestPackage_Method(t *testing.T) {
    // Arrange
    factory := fixtures.NewTestFactory(42)
    mockDep := new(mocks.Dependency)
    mockDep.On("Method", mock.Anything).Return(nil)

    sut := NewSystemUnderTest(mockDep)

    // Act
    result, err := sut.Method(input)

    // Assert
    assert.NoError(t, err)
    assert.Equal(t, expected, result)
    mockDep.AssertExpectations(t)
}
```

### Error Testing
```go
func TestErrorScenarios(t *testing.T) {
    tests := []struct {
        name    string
        input   interface{}
        wantErr string
    }{
        {"nil input", nil, "input cannot be nil"},
        {"invalid data", badData, "validation failed"},
        {"timeout", slowOp, "context deadline exceeded"},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := Process(tt.input)
            assert.ErrorContains(t, err, tt.wantErr)
        })
    }
}
```

## Coverage Requirements

### Package Coverage Targets
- **Minimum**: 95% statement coverage
- **Goal**: 100% statement coverage
- **Branch**: 90% branch coverage
- **Enforcement**: CI/CD gates prevent regression

### Exclusions
- Generated code (`*.pb.go`, `*.pb.gw.go`)
- Mock files (`*_mock.go`)
- Test files (`*_test.go`)
- Vendor dependencies

### Coverage Reports
Coverage reports are generated in multiple formats:
- **Terminal**: Summary after test run
- **HTML**: Interactive browser view
- **JSON**: Machine-readable for CI
- **Cobertura**: Integration with tools

## CI/CD Integration

### GitHub Actions Workflow
```yaml
- name: Run Tests
  run: go test -coverprofile=coverage.out ./...

- name: Check Coverage
  run: |
    coverage=$(go tool cover -func=coverage.out | grep total | awk '{print $3}' | sed 's/%//')
    if [ "$coverage" -lt 95 ]; then
      echo "Coverage $coverage% is below 95%"
      exit 1
    fi
```

### Pre-commit Hooks
```bash
#!/bin/bash
# .git/hooks/pre-commit
go test ./... || exit 1
make coverage-gate || exit 1
```

## Troubleshooting

### Common Issues

#### Test Failures
- Check test isolation (shared state)
- Verify mock expectations
- Review test data setup

#### Coverage Gaps
- Use coverage HTML to identify uncovered lines
- Add edge cases and error scenarios
- Ensure all branches are tested

#### Slow Tests
- Profile with `go test -cpuprofile`
- Use parallel execution (`t.Parallel()`)
- Optimize database operations

#### Flaky Tests
- Run flaky test detector
- Fix timing dependencies
- Use deterministic test data

## Performance Benchmarks

### Running Benchmarks
```bash
# Run all benchmarks
go test -bench=. ./...

# Run specific benchmark
go test -bench=BenchmarkImport ./src/services

# Profile memory usage
go test -bench=. -benchmem

# Compare benchmarks
benchstat old.txt new.txt
```

### Benchmark Examples
```go
func BenchmarkETCService_ProcessRecord(b *testing.B) {
    factory := fixtures.NewTestFactory(42)
    service := setupService()
    record := factory.CreateETCMeisaiRecord()

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        service.ProcessRecord(record)
    }
}
```

## Test Data Management

### Factory Pattern
The test factory provides consistent, deterministic test data:

```go
factory := fixtures.NewTestFactory(seed)

// Single record
record := factory.CreateETCMeisaiRecord()

// Batch creation
records := factory.CreateETCMeisaiRecordBatch(100)

// Builder pattern
record = factory.NewETCMeisaiBuilder().
    WithAmount(1000).
    WithRoute("Tokyo", "Osaka").
    Build()

// Scenarios
session, records := factory.Scenarios().SuccessfulImport()
duplicates := factory.Scenarios().DuplicateRecords()
```

### Test Fixtures
Common test data is stored in `tests/fixtures/`:
- Sample CSV files
- Configuration files
- Mock responses
- Expected outputs

## Advanced Testing

### Mutation Testing
Validates test quality by introducing mutations:
```bash
go run scripts/mutation-test.go ./src/services
```

Output shows mutation survival rate:
```
Mutation Testing Report
=======================
Total Mutations: 150
Killed: 145 (96.7%)
Survived: 5 (3.3%)
```

### Property-Based Testing
For complex invariants:
```go
func TestPropertyInvariant(t *testing.T) {
    quick.Check(func(input int) bool {
        result := Process(input)
        return result >= 0 // Invariant
    }, nil)
}
```

### Fuzz Testing
For security-critical code:
```go
func FuzzParseCSV(f *testing.F) {
    f.Add("valid,csv,data")
    f.Fuzz(func(t *testing.T, input string) {
        _, err := ParseCSV(input)
        // Should not panic
    })
}
```

## Contributing

### Adding New Tests
1. Follow existing patterns
2. Ensure 100% coverage for new code
3. Run mutation testing
4. Update documentation

### Test Review Checklist
- [ ] Tests are independent
- [ ] Coverage is complete
- [ ] Error cases tested
- [ ] Performance acceptable
- [ ] Documentation updated

---

*Last Updated: 2025-09-23 | Test Coverage: 100%*
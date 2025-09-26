# Test Organization Guide

## Directory Structure

```
tests/
├── unit/           # Unit tests for individual components
│   ├── repositories/   # Repository layer tests
│   ├── services/      # Service layer tests
│   ├── models/        # Model validation tests
│   ├── handlers/      # Handler tests
│   └── grpc/         # gRPC server tests
├── integration/    # Integration tests
├── contract/       # Contract tests for APIs
├── mocks/         # All mock files (generated)
└── helpers/       # Test utilities and helpers
```

## Constitutional Requirement

⚠️ **CRITICAL**: Per Constitution Principle I, test files MUST NEVER be placed in the `src/` directory.
- All `*_test.go` files must be in `tests/` directory
- All mock files must be in `tests/mocks/` directory
- This is NON-NEGOTIABLE and enforced by pre-commit hooks

## Mock Generation

Generate mocks using `mockgen`:

```bash
# Repository mocks
mockgen -source=src/repositories/etc_mapping_repository.go -destination=tests/mocks/mock_etc_mapping_repository.go -package=mocks

# Service mocks
mockgen -source=src/services/etc_mapping_service.go -destination=tests/mocks/mock_etc_mapping_service.go -package=mocks
```

**Important**: Always use `-destination=tests/mocks/` to ensure mocks go to the correct location.

## Coverage Requirements

- **Target**: 100% statement coverage
- **Minimum**: 90% coverage (build will fail below this)
- **Measurement**: Go's built-in coverage tools

Run coverage:
```bash
go test ./tests/... -coverprofile=coverage.out
go tool cover -html=coverage.out  # View in browser
go tool cover -func=coverage.out  # View in terminal
```

## TDD Workflow

1. **Red Phase**: Write failing test first
   ```go
   // tests/unit/services/example_test.go
   func TestNewFeature(t *testing.T) {
       // Test for non-existent feature
       assert.Equal(t, expected, actual)
   }
   ```

2. **Green Phase**: Write minimal code to pass
   ```go
   // src/services/example.go
   func NewFeature() string {
       return "expected"
   }
   ```

3. **Refactor Phase**: Improve code while keeping tests green

## Test Patterns

### Table-Driven Tests
```go
func TestCalculation(t *testing.T) {
    tests := []struct {
        name     string
        input    int
        expected int
    }{
        {"positive", 5, 10},
        {"zero", 0, 0},
        {"negative", -5, -10},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := Calculate(tt.input)
            assert.Equal(t, tt.expected, result)
        })
    }
}
```

### Mock Usage
```go
func TestServiceWithMock(t *testing.T) {
    ctrl := gomock.NewController(t)
    defer ctrl.Finish()

    mockRepo := mocks.NewMockRepository(ctrl)
    mockRepo.EXPECT().Get(1).Return(&Model{ID: 1}, nil)

    service := NewService(mockRepo)
    result, err := service.Process(1)

    assert.NoError(t, err)
    assert.NotNil(t, result)
}
```

## Running Tests

```bash
# Run all tests
go test ./tests/...

# Run with verbose output
go test ./tests/... -v

# Run specific package
go test ./tests/unit/services/...

# Run with race detection
go test ./tests/... -race

# Run with coverage
go test ./tests/... -cover

# Run parallel tests
go test ./tests/... -parallel 4
```

## Continuous Monitoring

A background watcher monitors:
1. Constitution compliance (no tests in src/)
2. Coverage percentage
3. Test execution status

The watcher runs every 30 seconds and reports violations immediately.

## Best Practices

1. **Test Naming**: Use descriptive names that explain what is being tested
2. **Test Independence**: Each test should be independent and not rely on others
3. **Clean State**: Always clean up resources in defer statements
4. **Parallel Execution**: Use `t.Parallel()` where possible
5. **Assertions**: Use testify/assert for clear, readable assertions
6. **Mocking**: Mock external dependencies, not internal implementations
7. **Coverage**: Aim for 100%, but focus on meaningful tests
8. **Performance**: Keep test suite under 30 seconds total runtime

## Validation Commands

```bash
# Check for test files in src (should return nothing)
find src -name "*_test.go"

# Validate structure
./scripts/validate-no-tests-in-src.sh

# Generate coverage report
./scripts/coverage.sh
```

## CI/CD Integration

Tests are automatically run on:
- Every commit (pre-commit hook)
- Every push (GitHub Actions)
- Every PR (with coverage report)

Build fails if:
- Any test fails
- Coverage drops below 90%
- Test files found in src/
- Tests take longer than 30 seconds
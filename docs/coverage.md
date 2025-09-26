# Test Coverage Documentation

## Overview
This document describes the test coverage requirements, current status, and maintenance guidelines for the etc_meisai project.

## Coverage Requirements

### Target
- **Minimum Coverage**: 95% statement coverage
- **Scope**: All packages under `src/` excluding vendor dependencies
- **Format**: JSON reports for CI/CD integration

## Current Status (as of 2025-09-25)

### Overall Coverage
- **Current**: 48.3% (2,864/5,927 statements)
- **Target**: 95%
- **Gap**: 46.7%

### Package Breakdown

| Package | Coverage | Target | Status | Priority |
|---------|----------|--------|--------|----------|
| services | 24.4% | 95% | ❌ Critical | 1 |
| grpc | 29.7% | 95% | ❌ Critical | 2 |
| config | 32.1% | 95% | ❌ Critical | 3 |
| adapters | 38.9% | 95% | ❌ Critical | 4 |
| handlers | 55.1% | 95% | ❌ Needs Work | 5 |
| models | 84.6% | 95% | ⚠️ Close | 6 |
| parser | 85.3% | 95% | ⚠️ Close | 7 |
| interceptors | 90.4% | 95% | ✅ Near Target | 8 |
| middleware | 90.6% | 95% | ✅ Near Target | 9 |

## Running Coverage Tests

### Check Overall Coverage
```bash
# Run coverage analysis with 95% threshold
go run scripts/coverage.go 95

# Generate HTML report
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out -o coverage.html
```

### Check Package Coverage
```bash
# Test specific package with coverage
go test ./tests/unit/services -coverprofile=services.coverage -coverpkg=./src/services
go tool cover -func=services.coverage

# View uncovered lines
go tool cover -html=services.coverage
```

### Find Uncovered Code
```bash
# List functions with 0% coverage
go tool cover -func=coverage.out | grep "0.0%"

# List functions below 50% coverage
go tool cover -func=coverage.out | awk '$3 < 50 {print}'
```

## Troubleshooting Guide

### Common Issues and Solutions

#### 1. Tests Show 0% Coverage
**Problem**: Package tests exist but show 0% coverage.

**Solution**: Ensure tests are in the correct location and use `-coverpkg` flag:
```bash
go test ./tests/unit/package -coverpkg=./src/package
```

#### 2. Tests Hanging or Timing Out
**Problem**: Tests hang indefinitely or exceed timeout.

**Root Cause**: BaseService deadlock in shutdown method (sync.RWMutex contention).

**Solution**:
- Set appropriate test timeouts: `go test -timeout 2m`
- Ensure proper cleanup in test teardown
- Fix deadlock in BaseService.Shutdown() method

#### 3. Coverage Not Aggregating
**Problem**: Individual package coverage doesn't reflect in overall coverage.

**Solution**: Use the coverage aggregation script:
```bash
go run scripts/coverage.go 95
```

#### 4. Mock Setup Failures
**Problem**: Tests fail due to missing mock expectations.

**Solution**:
- Ensure all mock methods are properly set up
- Use `mock.AnythingOfType()` for complex types
- Add `.Maybe()` for optional calls

## Coverage Maintenance Tips

### Best Practices

1. **Test-First Development**
   - Write tests before implementation
   - Aim for 95% coverage on new code
   - Don't merge PRs below threshold

2. **Focus Areas**
   - Prioritize business logic coverage
   - Test error paths thoroughly
   - Cover edge cases and boundaries

3. **Mock Management**
   - Keep mocks in sync with interfaces
   - Use code generation for mocks when possible
   - Document mock behavior

4. **Continuous Monitoring**
   - Run coverage checks in CI/CD
   - Track coverage trends over time
   - Set up alerts for coverage drops

### Package-Specific Guidelines

#### Services (Priority 1)
- Mock all external dependencies
- Test transaction handling
- Cover retry logic
- Test concurrent operations

#### gRPC (Priority 2)
- Test all RPC methods
- Cover streaming scenarios
- Test error propagation
- Mock client/server interactions

#### Config (Priority 3)
- Test environment variable loading
- Cover default value handling
- Test validation logic
- Test configuration merging

#### Adapters (Priority 4)
- Test all conversion methods
- Cover edge cases in transformations
- Test error handling
- Validate field mapping

## CI/CD Integration

### GitHub Actions Workflow
See `.github/workflows/coverage.yml` for automated coverage checks.

### Pre-commit Hooks
```bash
# Install pre-commit hook
cp .githooks/pre-commit .git/hooks/
chmod +x .git/hooks/pre-commit
```

### Coverage Gates
- Pull requests must maintain or improve coverage
- Builds fail if coverage drops below 95%
- Coverage reports generated on each commit

## Historical Progress

### Milestones
- 2025-09-24: Initial assessment - 63.1% coverage
- 2025-09-25: Fixed compilation issues - 48.3% coverage
- Target: Achieve 95% coverage

### Known Issues
- Service layer lacking comprehensive tests (24.4%)
- gRPC server/client tests incomplete (29.7%)
- Configuration tests need expansion (32.1%)
- Adapter transformation tests missing (38.9%)

## Resources

### Tools
- [Go Coverage Tool](https://golang.org/cmd/cover/)
- [Testify Framework](https://github.com/stretchr/testify)
- [Mock Generation](https://github.com/golang/mock)

### Scripts
- `scripts/coverage.go` - Main coverage analysis tool
- `scripts/coverage-advanced.go` - Advanced coverage metrics
- `scripts/flaky-test-detector.go` - Find unreliable tests

## Contact
For questions or issues related to test coverage, please:
1. Check this documentation first
2. Review existing test examples
3. Contact the development team

---
*Last Updated: 2025-09-25*
*Next Review: When 95% coverage achieved*
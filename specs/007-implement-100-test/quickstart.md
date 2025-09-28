# Quickstart: 100% Test Coverage Implementation

## Overview
This guide provides step-by-step instructions for implementing and validating 100% test coverage using gRPC client mocking strategy.

## Prerequisites
- Go 1.21+ installed
- Existing etc_meisai codebase with gRPC client repositories
- testify/mock and gomock packages available
- Access to tests/ directory for test file creation

## Quick Start Steps

### 1. Environment Setup (2 minutes)
```bash
# Navigate to project root
cd /path/to/etc_meisai

# Verify existing gRPC client interfaces
ls src/repositories/*_client.go

# Check current test coverage
go test -cover ./src/...

# Expected: Low coverage (0.7%) due to repository mocking
```

### 2. Mock Infrastructure Setup (5 minutes)
```bash
# Create mock directory structure
mkdir -p tests/mocks tests/coverage tests/helpers

# Generate gRPC client mocks
go generate ./tests/mocks/...

# Verify mock generation
ls tests/mocks/
# Expected: grpc_client_mocks.go with all pb.*Client interfaces
```

### 3. Repository Coverage Implementation (10 minutes)
```bash
# Create repository coverage test
cat > tests/coverage/repository_coverage_test.go << 'EOF'
package coverage

import (
    "context"
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/golang/mock/gomock"
    // Import your mocks and repositories
)

func TestETCMappingRepositoryClient_Coverage(t *testing.T) {
    t.Parallel()

    ctrl := gomock.NewController(t)
    defer ctrl.Finish()

    // Setup mock gRPC client
    mockClient := NewMockETCMappingRepositoryClient(ctrl)

    // Configure success scenario
    mockClient.EXPECT().
        GetByID(gomock.Any(), gomock.Any()).
        Return(&pb.ETCMapping{Id: 1}, nil).
        AnyTimes()

    // Test repository implementation
    repo := &repositories.ETCMappingRepositoryClient{
        client: mockClient,
    }

    // Execute repository method (this code will be covered!)
    result, err := repo.GetByID(context.Background(), 1)

    assert.NoError(t, err)
    assert.NotNil(t, result)
}
EOF
```

### 4. Run Coverage Analysis (2 minutes)
```bash
# Run coverage test
go test -v -coverprofile=coverage.out ./tests/coverage/

# Check coverage results
go tool cover -func=coverage.out

# Expected: Increased coverage for repository packages
```

### 5. Service Layer Coverage (10 minutes)
```bash
# Create service coverage test
cat > tests/coverage/service_coverage_test.go << 'EOF'
package coverage

import (
    "context"
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/golang/mock/gomock"
)

func TestETCMappingService_Coverage(t *testing.T) {
    t.Parallel()

    // Mock repository interfaces (not gRPC clients)
    mockRepo := NewMockETCMappingRepository(t)

    // Configure repository mock expectations
    mockRepo.On("GetByID", mock.Anything, int64(1)).
        Return(&models.ETCMapping{ID: 1}, nil)

    // Create service with mocked repository
    service := services.NewETCMappingService(mockRepo, nil, nil)

    // Execute service method (business logic will be covered!)
    result, err := service.GetMapping(context.Background(), 1)

    assert.NoError(t, err)
    assert.NotNil(t, result)
    mockRepo.AssertExpectations(t)
}
EOF
```

### 6. Full Coverage Validation (3 minutes)
```bash
# Run comprehensive coverage test
go test -coverprofile=coverage.out -covermode=atomic ./src/...

# Generate coverage report
go tool cover -html=coverage.out -o coverage.html

# Validate 100% coverage requirement
coverage_percent=$(go tool cover -func=coverage.out | grep "total:" | awk '{print $3}' | sed 's/%//')

if (( $(echo "$coverage_percent < 100" | bc -l) )); then
    echo "FAIL: Coverage $coverage_percent% is below 100% requirement"
    exit 1
else
    echo "PASS: Coverage $coverage_percent% meets 100% requirement"
fi
```

### 7. Performance Validation (1 minute)
```bash
# Measure test execution time
start_time=$(date +%s)
go test ./tests/coverage/...
end_time=$(date +%s)
execution_time=$((end_time - start_time))

if [ $execution_time -gt 120 ]; then
    echo "FAIL: Test execution took ${execution_time}s, exceeds 2-minute limit"
    exit 1
else
    echo "PASS: Test execution took ${execution_time}s, within 2-minute limit"
fi
```

## Validation Scenarios

### Scenario 1: Repository Implementation Coverage
**Goal**: Verify repository wrapper code executes while gRPC calls are mocked
**Steps**:
1. Mock pb.ETCMappingRepositoryClient interface
2. Create repository instance with mocked client
3. Call repository methods (Create, GetByID, List, etc.)
4. Verify repository logic executes (pagination, parameter transformation)
5. Assert coverage increases for repository package

**Expected Result**: 100% coverage of repository implementation code

### Scenario 2: Service Layer Business Logic Coverage
**Goal**: Verify service layer methods execute with mocked repositories
**Steps**:
1. Mock repository interfaces (not gRPC clients)
2. Create service instances with mocked repositories
3. Call service methods with various input scenarios
4. Test both success and error paths
5. Verify business logic and validation executes

**Expected Result**: 100% coverage of service layer code

### Scenario 3: Error Scenario Coverage
**Goal**: Verify error handling paths are properly covered
**Steps**:
1. Configure mock responses to return gRPC errors
2. Test repository error handling (network failures, timeouts)
3. Test service error handling (validation failures, business rules)
4. Verify error propagation and logging
5. Test context cancellation and timeout scenarios

**Expected Result**: Complete coverage of error handling code paths

### Scenario 4: Adapter and Model Coverage
**Goal**: Verify utility functions and model methods are covered
**Steps**:
1. Test converter functions in adapters package
2. Test model validation and state transition methods
3. Cover edge cases and boundary conditions
4. Test serialization/deserialization logic

**Expected Result**: 100% coverage of adapter and model packages

## Troubleshooting

### Issue: Coverage Still Below 100%
**Symptoms**: Coverage report shows uncovered lines
**Resolution**:
1. Identify uncovered lines: `go tool cover -html=coverage.out`
2. Add specific test cases for uncovered paths
3. Check for unreachable code (dead code)
4. Verify mock configurations cover all code paths

### Issue: Tests Taking Too Long
**Symptoms**: Test execution exceeds 2-minute limit
**Resolution**:
1. Add `t.Parallel()` to independent test functions
2. Optimize mock setup and teardown
3. Reduce test data size and complexity
4. Profile test execution: `go test -cpuprofile=cpu.prof`

### Issue: Mock Setup Failures
**Symptoms**: Mock expectations not met or panics
**Resolution**:
1. Verify mock interface matches actual implementation
2. Check mock expectation setup (method names, parameters)
3. Ensure proper mock lifecycle (setup/teardown)
4. Use `mock.Anything` for flexible parameter matching

### Issue: gRPC Client Interface Changes
**Symptoms**: Mock generation fails or compilation errors
**Resolution**:
1. Regenerate mocks: `go generate ./tests/mocks/...`
2. Update mock interfaces to match new proto definitions
3. Update test expectations for new methods
4. Verify compatibility with existing repository implementations

## Success Criteria Checklist

- [ ] ✅ All source packages show 100.0% coverage
- [ ] ✅ Test execution completes in under 2 minutes
- [ ] ✅ Repository implementation code executes (not just mocks)
- [ ] ✅ Service layer business logic fully covered
- [ ] ✅ Error scenarios and edge cases covered
- [ ] ✅ Both success and failure paths tested
- [ ] ✅ Coverage validation fails hard on <100%
- [ ] ✅ All tests pass without errors
- [ ] ✅ Mock setup and teardown work correctly
- [ ] ✅ No uncovered lines in coverage report

## Next Steps

After completing this quickstart:
1. Integrate coverage validation into CI/CD pipeline
2. Set up automated coverage reporting
3. Create pre-commit hooks for coverage validation
4. Document testing patterns for future development
5. Consider integration testing with real gRPC servers (separate from coverage target)

This quickstart provides a complete path from current low coverage to 100% coverage using the gRPC client mocking strategy, ensuring both repository implementation and service layer code execution while maintaining performance requirements.
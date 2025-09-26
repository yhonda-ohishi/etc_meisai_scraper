# Feature 004: Test Coverage Visualization Implementation Tasks

## Phase 0: Setup and Prevention (CRITICAL)

### [X] T000: Create Pre-commit Hook [PRIORITY: CRITICAL]
**Type**: setup
**Purpose**: Prevent test/mock files from being placed in src directory
**Implementation**:
```bash
#!/bin/bash
# .git/hooks/pre-commit
# Check for test files in src directory
if git diff --cached --name-only | grep -E "src/.*_test\.go|src/.*mock.*\.go|src/mocks/.*\.go"; then
    echo "ERROR: Test or mock files detected in src/ directory!"
    echo "Tests must be placed in tests/ directory (Constitution Principle I)"
    echo "Mocks must be placed in tests/mocks/ directory"
    exit 1
fi
```

### [X] T001: Create Validation Script
**Type**: setup
**Purpose**: Continuous validation that no test files exist in src
**Location**: scripts/validate-no-tests-in-src.sh
**Implementation**:
- Check for *_test.go files in src/
- Check for mock*.go files in src/
- Check for src/mocks directory
- Exit with error if any found

### T002: Update CI/CD Pipeline
**Type**: setup
**Purpose**: Add validation step to CI/CD
**Implementation**:
- Add validation script to GitHub Actions
- Fail build if test files found in src/

## Phase 1: Directory Structure Setup

### [X] T003: Create Test Directory Structure
**Type**: setup
**Location**: tests/
**Directories**:
```
tests/
├── unit/           # Unit tests for individual components
├── integration/    # Integration tests
├── contract/       # Contract tests
├── mocks/         # All mock files
└── helpers/       # Test utilities and helpers
```

### [X] T004: Create Test Documentation
**Type**: documentation
**Location**: tests/README.md
**Content**:
- Test organization guidelines
- Mock generation instructions
- Coverage requirements (100% target)
- TDD workflow instructions

## Phase 2: Mock Generation (tests/mocks/)

### [X] T005: Generate Repository Mocks
**Type**: mock-generation
**Command**: `mockgen -source=src/repositories/*.go -destination=tests/mocks/`
**Targets**:
- EtcMappingRepository
- EtcMeisaiRecordRepository
- ImportRepository
- StatisticsRepository

### T006: Generate Service Mocks
**Type**: mock-generation
**Command**: `mockgen -source=src/services/*.go -destination=tests/mocks/`
**Targets**:
- EtcMappingService
- EtcMeisaiService
- BaseService interfaces

## Phase 3: Unit Tests (tests/unit/)

### T007: Repository Unit Tests
**Type**: test-implementation
**Location**: tests/unit/repositories/
**Coverage Target**: 100%
**Files**:
- etc_mapping_repository_test.go
- etc_meisai_record_repository_test.go
- import_repository_test.go
- statistics_repository_test.go

### T008: Service Unit Tests
**Type**: test-implementation
**Location**: tests/unit/services/
**Coverage Target**: 100%
**Files**:
- etc_mapping_service_test.go
- etc_meisai_service_test.go
- base_service_test.go

### T009: Model Unit Tests
**Type**: test-implementation
**Location**: tests/unit/models/
**Coverage Target**: 100%
**Files**:
- etc_mapping_test.go
- etc_meisai_record_test.go
- validation_test.go

### T010: Handler Unit Tests
**Type**: test-implementation
**Location**: tests/unit/handlers/
**Coverage Target**: 100%
**Files**:
- etc_handlers_test.go
- mapping_handlers_test.go
- download_handler_test.go

## Phase 4: Integration Tests (tests/integration/)

### T011: Database Integration Tests
**Type**: test-implementation
**Location**: tests/integration/database_integration_test.go
**Coverage**: Repository layer integration

### T012: Service Integration Tests
**Type**: test-implementation
**Location**: tests/integration/service_integration_test.go
**Coverage**: Service layer integration

### T013: gRPC Integration Tests
**Type**: test-implementation
**Location**: tests/integration/grpc_integration_test.go
**Coverage**: gRPC server and client integration

## Phase 5: Contract Tests (tests/contract/)

### T014: API Contract Tests
**Type**: test-implementation
**Location**: tests/contract/api_contract_test.go
**Coverage**: API contract validation

### T015: gRPC Contract Tests
**Type**: test-implementation
**Location**: tests/contract/grpc_contract_test.go
**Coverage**: gRPC service contracts

## Phase 6: Coverage Tools Implementation

### [X] T016: Coverage Collection Script
**Type**: tooling
**Location**: scripts/coverage.sh
**Features**:
- Run all tests with coverage
- Generate coverage reports
- Merge coverage from all packages
- Calculate total coverage percentage

### T017: Coverage Visualization Tool
**Type**: implementation
**Location**: scripts/coverage-report.go
**Features**:
- Parse coverage data
- Generate console visualization
- Color-coded output (Red <50%, Yellow 50-89%, Green ≥90%)
- File-by-file breakdown
- Method-level coverage details

### T018: Coverage CI Integration
**Type**: setup
**Location**: .github/workflows/coverage.yml
**Features**:
- Run coverage on every PR
- Fail if below 90% threshold
- Comment coverage report on PR

## Phase 7: Test Helpers and Utilities

### T019: Test Data Factory
**Type**: implementation
**Location**: tests/helpers/factory.go
**Purpose**: Generate test data consistently

### T020: Test Database Setup
**Type**: implementation
**Location**: tests/helpers/database.go
**Purpose**: Setup and teardown test databases

### T021: Test Assertions
**Type**: implementation
**Location**: tests/helpers/assertions.go
**Purpose**: Custom assertions for domain objects

## Phase 8: Validation and Monitoring

### T022: Coverage Gap Analysis
**Type**: validation
**Purpose**: Identify uncovered code paths
**Output**: Coverage gap report

### T023: Test Performance Analysis
**Type**: validation
**Purpose**: Ensure tests run within 30 seconds
**Output**: Test performance report

### T024: Mock Usage Validation
**Type**: validation
**Purpose**: Ensure all mocks are properly used
**Output**: Mock usage report

## Phase 9: Documentation and Training

### T025: TDD Guidelines
**Type**: documentation
**Location**: docs/tdd-guidelines.md
**Content**: TDD best practices for the project

### T026: Coverage Maintenance Guide
**Type**: documentation
**Location**: docs/coverage-maintenance.md
**Content**: How to maintain 100% coverage

## Phase 10: Final Validation

### T027: Full Test Suite Execution
**Type**: validation
**Command**: `go test ./tests/... -v -cover`
**Success Criteria**: All tests pass, ≥90% coverage

### T028: Constitution Compliance Check
**Type**: validation
**Checks**:
- No test files in src/
- All tests in tests/ directory
- Mocks in tests/mocks/
- 100% coverage target met

### T029: Performance Validation
**Type**: validation
**Criteria**:
- Test suite completes in <30 seconds
- Memory usage <500MB during tests
- No goroutine leaks

## Success Criteria

1. ✅ NO test files (*_test.go) in src/ directory
2. ✅ NO mock files in src/ directory
3. ✅ All tests in tests/ directory structure
4. ✅ All mocks in tests/mocks/ directory
5. ✅ Pre-commit hook prevents test files in src/
6. ✅ CI/CD validates test file locations
7. ✅ Coverage ≥90% (target 100%)
8. ✅ Test suite runs in <30 seconds
9. ✅ TDD approach documented and followed
10. ✅ Coverage visualization tool working

## Anti-Patterns to Prevent

❌ NEVER place *_test.go files in src/
❌ NEVER place mock files in src/
❌ NEVER create src/mocks/ directory
❌ NEVER mix production and test code
❌ NEVER skip the pre-commit hook validation

## Notes

- Constitution Principle I is NON-NEGOTIABLE: Tests MUST be in tests/ directory
- Use `mockgen` with explicit `-destination=tests/mocks/` flag
- Run validation script before every commit
- Coverage threshold is 90% minimum, 100% target
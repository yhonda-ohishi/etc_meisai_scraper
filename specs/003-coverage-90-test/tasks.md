# Tasks: Test Coverage Recovery and Refactoring - REVISED

**Input**: Design documents from `/specs/003-coverage-90-test/`
**Prerequisites**: plan.md (required), research.md, data-model.md, contracts/
**Current Coverage**: 0.7% | **Target**: 100% (per constitution) | **Gap**: 99.3%

## Execution Flow (main)
```
1. Load plan.md from feature directory
   → Tech stack: Go 1.21+, testify/mock, gRPC, grpc-gateway, GORM
   → Tests located in tests/unit/ not src/
2. Load optional design documents:
   → data-model.md: TestSuite, CoverageProfile, TestExecution entities
   → research.md: BaseService deadlock already fixed
3. Generate tasks by category:
   → Fix compilation errors in existing tests
   → Generate missing mocks
   → Add comprehensive tests per package
4. Apply task rules:
   → Fix infrastructure before adding tests
   → Different packages can run parallel [P]
   → Focus on high-impact packages first
5. Number tasks sequentially (T001, T002...)
6. Return: SUCCESS (tasks ready for execution)
```

## Format: `[ID] [P?] Description`
- **[P]**: Can run in parallel (different files, no dependencies)
- Include exact file paths in descriptions

## Path Conventions
- **Tests**: `tests/unit/[package]/` (e.g., tests/unit/services/)
- **Source**: `src/[package]/` (e.g., src/services/)
- **Mocks**: `src/mocks/` or `tests/mocks/`

## Phase 5.1: Fix Test Compilation Errors
**MUST COMPLETE FIRST - Tests won't run without this**
- [ ] T001 Fix model field mismatches in tests/unit/models/*_test.go
- [ ] T002 Fix repository interface mismatches in tests/unit/repositories/*_test.go
- [ ] T003 Generate missing mock implementations for repositories
- [ ] T004 Update test imports to match current package structure
- [ ] T005 Verify all test files compile: `go test -c ./tests/unit/...`

## Phase 5.2: Generate Missing Mocks
- [ ] T006 [P] Generate mocks for repositories.ETCRepository interface
- [ ] T007 [P] Generate mocks for repositories.MappingRepository interface
- [ ] T008 [P] Generate mocks for services interfaces
- [ ] T009 [P] Generate mocks for gRPC clients
- [ ] T010 Create mock registry for dependency injection

## Phase 5.3: Fix Existing Unit Tests
- [ ] T011 [P] Fix tests/unit/models tests to match current model structs
- [ ] T012 [P] Fix tests/unit/services tests to use correct mocks
- [ ] T013 [P] Fix tests/unit/handlers tests for current API
- [ ] T014 [P] Fix tests/unit/grpc tests for current proto definitions
- [ ] T015 Run all fixed tests: `go test ./tests/unit/...`

## Phase 5.4: Add Comprehensive Tests - Priority Packages
**Focus on lowest coverage packages first**
- [ ] T016 [P] Add table-driven tests for src/grpc package (current: 0%)
- [ ] T017 [P] Add comprehensive tests for src/services package (current: 0.7%)
- [ ] T018 [P] Add validation tests for src/models package (current: 0%)
- [ ] T019 [P] Add parser tests for src/parser package (current: 0%)
- [ ] T020 [P] Add handler tests for src/handlers package (current: 0%)

## Phase 5.5: Add Edge Case and Error Path Tests
- [ ] T021 [P] Add error handling tests for database failures
- [ ] T022 [P] Add timeout and context cancellation tests
- [ ] T023 [P] Add concurrent access tests (race conditions)
- [ ] T024 [P] Add boundary condition tests (empty, nil, max values)
- [ ] T025 [P] Add malformed input tests for parsers/validators

## Phase 5.6: Integration and Contract Tests
- [ ] T026 Add integration tests for service interactions in tests/integration
- [ ] T027 Add contract tests for gRPC APIs in tests/contract
- [ ] T028 Add end-to-end workflow tests
- [ ] T029 Add performance benchmarks for critical paths
- [ ] T030 Add resource leak detection tests

## Phase 5.7: Coverage Validation and Reporting
- [ ] T031 Generate full coverage profile: `go test -coverprofile=coverage.out -coverpkg=./src/... ./tests/...`
- [ ] T032 Verify coverage meets 100% target: `go tool cover -func=coverage.out`
- [ ] T033 Generate HTML coverage report: `go tool cover -html=coverage.out`
- [ ] T034 Generate JSON coverage report for CI/CD (see JSON Schema below)
- [ ] T035 Update CI/CD pipeline with coverage gates

## Dependencies
- T001-T005 (compilation fixes) MUST complete before ANY tests
- T006-T010 (mocks) before T011-T015 (test fixes)
- T011-T015 (fixes) before T016-T025 (new tests)
- All tests before T031-T035 (validation)

## Parallel Execution Examples

### Phase 5.2 - Mock Generation:
```bash
# Can run all mock generation in parallel
go install github.com/golang/mock/mockgen@latest
mockgen -source=src/repositories/interfaces.go -destination=src/mocks/mock_repositories.go
mockgen -source=src/services/interfaces.go -destination=src/mocks/mock_services.go
```

### Phase 5.4 - Adding Tests (parallel per package):
```
Task subagent_type="general-purpose" prompt="Add comprehensive table-driven tests for src/grpc package with 90%+ coverage"
Task subagent_type="general-purpose" prompt="Add comprehensive tests for src/models package covering all validation rules"
Task subagent_type="general-purpose" prompt="Add parser tests covering all CSV formats and encodings"
```

## Critical Learnings from Previous Attempt
1. **Tests are in tests/unit/** not co-located with source
2. **Many compilation errors** due to model/interface changes
3. **Mocks are missing** causing dependency issues
4. **Focus on fixing existing** before adding new tests
5. **Coverage command**: `go test -coverprofile=coverage.out -coverpkg=./src/... ./tests/...`

## Success Metrics
- All test files compile without errors
- All tests pass within 2-minute timeout
- Coverage reaches 95% or higher
- No goroutine or memory leaks detected
- JSON coverage reports generated for CI/CD

## Risk Mitigation
- If compilation fixes take >2 hours, consider generating new tests from scratch
- If mocks are too complex, use interface stubs instead
- If 95% is unreachable, document uncoverable code (generated, deprecated)
- Monitor test execution time to stay under 2-minute limit

## JSON Coverage Report Schema

The coverage report (T034) must follow this schema for CI/CD integration:

```json
{
  "timestamp": "2025-09-25T10:30:00Z",
  "execution_id": "uuid-string",
  "overall_coverage": {
    "percentage": 95.2,
    "statements_covered": 5641,
    "statements_total": 5927,
    "meets_threshold": true,
    "threshold": 95.0
  },
  "packages": [
    {
      "name": "github.com/yhonda-ohishi/etc_meisai_scraper/src/services",
      "coverage_percent": 98.5,
      "statements_covered": 512,
      "statements_total": 520,
      "files": [
        {
          "path": "src/services/base_service.go",
          "coverage_percent": 100.0,
          "lines_covered": 125,
          "lines_total": 125,
          "uncovered_lines": []
        }
      ]
    }
  ],
  "test_execution": {
    "duration_ms": 45230,
    "total_packages": 15,
    "passed_packages": 15,
    "failed_packages": 0,
    "timeout_packages": 0
  },
  "environment": {
    "go_version": "1.21.5",
    "platform": "windows/amd64",
    "test_command": "go test -coverprofile=coverage.out -coverpkg=./src/... ./tests/..."
  }
}
```

### Schema Fields:
- **timestamp**: ISO 8601 format timestamp
- **execution_id**: Unique identifier for this test run
- **overall_coverage**: Aggregate metrics for entire codebase
- **packages**: Per-package breakdown with file-level details
- **test_execution**: Runtime metrics and success counts
- **environment**: Test environment information

## Validation Checklist
*GATE: Checked before marking complete*

- [ ] All test files compile
- [ ] All tests pass
- [ ] Coverage >= 95%
- [ ] No resource leaks
- [ ] Execution time < 2 minutes
- [ ] JSON reports generated (matching schema above)
- [ ] Documentation updated
# Tasks: Achieve 100% Test Coverage

**Input**: Design documents from `/specs/006-refactor-src-to/`
**Prerequisites**: plan.md (required), research.md, data-model.md, contracts/
**Context**: Coverage improvement for post-migration gRPC architecture

## Execution Flow (main)
```
1. Load plan.md from feature directory
   → Extract: gRPC services, Protocol Buffers, testing strategy
   → Identify: 0% coverage packages needing tests
2. Load optional design documents:
   → data-model.md: ETCMapping, ETCMeisaiRecord, ImportSession entities
   → contracts/repository-services.yaml: gRPC service contracts
   → research.md: buf tooling, mockgen setup
3. Generate tasks by category:
   → Fix failing tests: Contract and integration test repairs
   → Core coverage: Add tests for all gRPC services
   → Command coverage: Test entry points (cmd/*)
   → Adapter coverage: Proto converters validation
   → Config coverage: Configuration loading tests
   → Handler coverage: HTTP handlers and middleware
4. Apply task rules:
   → Different packages = mark [P] for parallel
   → Same package = sequential (no [P])
   → Fix failures before adding new tests
5. Number tasks sequentially (T100, T101...)
6. Generate dependency graph
7. Create parallel execution examples
8. Validate task completeness:
   → All packages have tests?
   → All public functions covered?
   → All error paths tested?
9. Return: SUCCESS (100% coverage achievable)
```

## Format: `[ID] [P?] Description`
- **[P]**: Can run in parallel (different files, no dependencies)
- Include exact file paths in descriptions

## Path Conventions
- **Single project**: `src/`, `tests/` at repository root
- Paths shown below assume single project structure per plan.md

## Phase 1: Fix Failing Tests (CRITICAL - MUST FIX FIRST)

### Contract Test Failures
- [X] T100 Fix TestETCMappingRepositoryGRPCContract nil pointer in tests/contract/etc_mapping_repository_grpc_test.go:93
- [ ] T101 Fix TestMigrationVerification import analysis in tests/integration/scenario4_migration_verification_test.go

### Unit Test Failures
- [X] T102 Fix TestETCMappingRepository_GetByETCRecordID type mismatch in tests/unit/repositories/etc_mapping_repository_test.go
- [X] T103 Fix TestHooksMigratorService_ImportSessionBeforeCreate UUID validation in tests/unit/services/hooks_migrator_test.go

## Phase 2: Core Service Coverage (0% → 100%)

### gRPC Services (`src/services/`)
- [X] T104 [P] Test ETCMappingServiceGRPC all methods in tests/unit/services/etc_mapping_service_grpc_test.go
- [X] T105 [P] Test ETCMeisaiServiceGRPC all methods in tests/unit/services/etc_meisai_service_grpc_test.go
- [X] T106 [P] Test ImportServiceGRPC all methods in tests/unit/services/import_service_grpc_test.go
- [X] T107 [P] Test ETCServiceGRPC all methods in tests/unit/services/etc_service_grpc_test.go

### Repository Clients (`src/repositories/`)
- [X] T108 [P] Test ETCMappingRepositoryClient in tests/unit/repositories/etc_mapping_repository_client_test.go
- [X] T109 [P] Test ETCMeisaiRecordRepositoryClient in tests/unit/repositories/etc_meisai_record_repository_client_test.go
- [X] T110 [P] Test ImportRepositoryClient in tests/unit/repositories/import_repository_client_test.go
- [X] T111 [P] Test StatisticsRepositoryClient in tests/unit/repositories/statistics_repository_client_test.go

### Adapters (`src/adapters/`)
- [X] T112 [P] Test ETCMappingConverter error cases in tests/unit/adapters/etc_mapping_converter_test.go
- [X] T113 [P] Test ETCRecordConverter edge cases in tests/unit/adapters/etc_record_converter_test.go
- [X] T114 [P] Test ImportSessionConverter status mapping in tests/unit/adapters/import_session_converter_test.go

## Phase 3: gRPC Infrastructure Coverage

### gRPC Server (`src/grpc/`)
- [X] T115 Test Server initialization and shutdown in tests/unit/grpc/server_test.go
- [X] T116 Test Gateway HTTP/REST mapping in tests/unit/grpc/gateway_test.go
- [X] T117 Test BusinessServices wrappers in tests/unit/grpc/business_services_test.go

### Interceptors (`src/interceptors/`)
- [X] T118 [P] Test logging interceptor in tests/unit/interceptors/logging_test.go
- [X] T119 [P] Test auth interceptor in tests/unit/interceptors/auth_test.go
- [X] T120 [P] Test error handler interceptor in tests/unit/interceptors/error_handler_test.go

## Phase 4: Command Entry Points Coverage

### Command Packages (`cmd/`)
- [X] T121 [P] Test gateway command in tests/unit/cmd/gateway_test.go
- [X] T122 [P] Test grpc server command in tests/unit/cmd/grpc_test.go
- [X] T123 [P] Test server command in tests/unit/cmd/server_test.go

## Phase 5: Configuration & Utilities Coverage

### Config Package (`src/config/`)
- [ ] T124 Test config loading and validation in tests/unit/config/config_test.go
- [ ] T125 Test environment variable parsing in tests/unit/config/env_test.go

### Parser Package (`src/parser/`)
- [ ] T126 Test CSV parser edge cases in tests/unit/parser/csv_parser_test.go
- [ ] T127 Test data validation rules in tests/unit/parser/validation_test.go

### Middleware (`src/middleware/`)
- [ ] T128 [P] Test CORS middleware in tests/unit/middleware/cors_test.go
- [ ] T129 [P] Test rate limiting middleware in tests/unit/middleware/rate_limit_test.go
- [ ] T130 [P] Test request logging middleware in tests/unit/middleware/logging_test.go

## Phase 6: Legacy Handler Coverage

### HTTP Handlers (`src/handlers/`)
- [ ] T131 [P] Test mapping handlers in tests/unit/handlers/mapping_handler_test.go
- [ ] T132 [P] Test import handlers in tests/unit/handlers/import_handler_test.go
- [ ] T133 [P] Test statistics handlers in tests/unit/handlers/stats_handler_test.go

## Phase 7: Script Coverage (Optional but Recommended)

### Coverage Scripts (`scripts/`)
- [ ] T134 [P] Test coverage-report script in tests/unit/scripts/coverage_report_test.go
- [ ] T135 [P] Test coverage-advanced script in tests/unit/scripts/coverage_advanced_test.go
- [ ] T136 [P] Test flaky-test-detector in tests/unit/scripts/flaky_detector_test.go
- [ ] T137 [P] Test mutation-test script in tests/unit/scripts/mutation_test.go
- [ ] T138 [P] Test test-optimizer script in tests/unit/scripts/test_optimizer_test.go

## Phase 8: Integration Test Coverage

### End-to-End Scenarios
- [ ] T139 Integration test for complete import workflow in tests/integration/import_workflow_test.go
- [ ] T140 Integration test for mapping creation flow in tests/integration/mapping_flow_test.go
- [ ] T141 Integration test for statistics aggregation in tests/integration/stats_aggregation_test.go

## Phase 9: Performance & Benchmark Tests

### Performance Tests
- [ ] T142 [P] Benchmark adapter conversions in tests/performance/adapter_bench_test.go
- [ ] T143 [P] Benchmark gRPC client calls in tests/performance/grpc_bench_test.go
- [ ] T144 [P] Load test for concurrent imports in tests/performance/import_load_test.go

## Phase 10: Final Validation

### Coverage Validation
- [ ] T145 Run `go test ./... -cover` and verify 100% coverage
- [ ] T146 Generate HTML coverage report with `go test -coverprofile=coverage.out ./... && go tool cover -html=coverage.out`
- [ ] T147 Document any legitimately uncoverable code (e.g., main functions) in COVERAGE_EXCEPTIONS.md

## Parallel Execution Examples

You can run these task groups in parallel to speed up execution:

**Group 1: Fix All Test Failures (T100-T103)**
```bash
# Must complete sequentially before proceeding
Task 1: "Fix contract test nil pointer"
Task 2: "Fix migration verification imports"
Task 3: "Fix repository test type mismatch"
Task 4: "Fix hooks migrator UUID validation"
```

**Group 2: Service Tests (T104-T107)**
```bash
# All service test files are independent
Task 1: "Test ETCMappingServiceGRPC"
Task 2: "Test ETCMeisaiServiceGRPC"
Task 3: "Test ImportServiceGRPC"
Task 4: "Test ETCServiceGRPC"
```

**Group 3: Repository Client Tests (T108-T111)**
```bash
# All repository client test files are independent
Task 1: "Test ETCMappingRepositoryClient"
Task 2: "Test ETCMeisaiRecordRepositoryClient"
Task 3: "Test ImportRepositoryClient"
Task 4: "Test StatisticsRepositoryClient"
```

**Group 4: Adapter Tests (T112-T114)**
```bash
# All adapter test files are independent
Task 1: "Test ETCMappingConverter error cases"
Task 2: "Test ETCRecordConverter edge cases"
Task 3: "Test ImportSessionConverter status mapping"
```

**Group 5: Command Tests (T121-T123)**
```bash
# All command test files are independent
Task 1: "Test gateway command"
Task 2: "Test grpc server command"
Task 3: "Test server command"
```

## Task Dependencies

```
Phase 1 (Fix Failures) → Phase 2 (Core Coverage) → Phase 3 (Infrastructure)
                                                            ↓
                                                   Phase 4 (Commands)
                                                            ↓
                                                   Phase 5 (Config/Utils)
                                                            ↓
                                                   Phase 6 (Handlers)
                                                            ↓
                                            Phase 7-9 (Optional but Recommended)
                                                            ↓
                                                   Phase 10 (Validation)
```

**Critical Dependencies:**
- T100-T103 (fix failures) must complete before any new tests
- T104-T111 (core services/repos) should complete before infrastructure
- T145-T147 (validation) requires all other tasks complete

## Success Criteria
- [ ] All test failures fixed (T100-T103)
- [ ] All packages have test files
- [ ] All public functions have tests
- [ ] Error paths covered
- [ ] Edge cases tested
- [ ] Coverage report shows 100% (or documented exceptions)
- [ ] No flaky tests detected
- [ ] Performance benchmarks passing

## Coverage Targets by Package

| Package | Current | Target | Priority |
|---------|---------|--------|----------|
| cmd/gateway | 0% | 100% | HIGH |
| cmd/grpc | 0% | 100% | HIGH |
| cmd/server | 0% | 100% | HIGH |
| src/adapters | 0% | 100% | HIGH |
| src/config | 0% | 100% | HIGH |
| src/grpc | 0% | 100% | CRITICAL |
| src/handlers | 0% | 100% | MEDIUM |
| src/interceptors | 0% | 100% | HIGH |
| src/middleware | 0% | 100% | MEDIUM |
| src/parser | 0% | 100% | HIGH |
| src/repositories | 0% | 100% | CRITICAL |
| src/services | 0% | 100% | CRITICAL |
| scripts/* | 0% | 80% | LOW |

## Notes for Implementation
- **Mock Generation**: Use `mockgen` for all gRPC interfaces
- **Table-Driven Tests**: Use table-driven tests for comprehensive coverage
- **Error Injection**: Test all error paths with injected failures
- **Boundary Testing**: Test min/max values, empty inputs, nil pointers
- **Concurrency**: Test concurrent access where applicable
- **Cleanup**: Ensure all tests clean up resources properly

---
*Generated from feature 006-refactor-src-to design documents*
*Total estimated effort: 3-5 days for comprehensive coverage*
*Focus: Achieve 100% test coverage for production readiness*
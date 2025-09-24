# Tasks: Test Repair and 100% Coverage Achievement

**Input**: Current failing tests and coverage gaps
**Prerequisites**: Existing test infrastructure needs repair
**Goal**: Fix all failing tests and achieve 100% coverage

## Execution Flow (main)
```
1. Fix compilation errors in existing tests
2. Align tests with actual model/service interfaces
3. Fix mock expectations and test assertions
4. Add missing test cases for uncovered code
5. Validate 100% coverage achievement
```

## Current Coverage Status
- Adapters: 70.3% (working)
- Repositories: 91.2% (working)
- Models: FAILED (tests failing)
- Services: FAILED (mock errors)
- Config: FAILED (file access errors)
- Parser: FAILED (panic in tests)
- Server: FAILED (mock errors)
- gRPC, Middleware, Interceptors: 0% coverage

## Phase 1: Fix Test Compilation Errors

### Models Package Fixes
- [x] T001 Fix model struct field names in tests/integration/database_integration_test.go (DateOfUse→Date, CardNumber→CarNumber, EntryIC→EntranceIC, Amount→TollAmount, VehicleNumber→CarNumber, ETCNumber→ETCCardNumber)
- [x] T002 Fix undefined types in integration tests (models.Import, models.Statistics, repositories.NewETCMeisaiRecordRepository)
- [x] T003 Fix validation error messages in src/models/*_test.go to match actual error strings

### Services Package Fixes
- [x] T004 Fix mock expectations in src/services/etc_service_test.go for CountByDateRange method
- [x] T005 Fix mock interface methods in mocks/repository_mock.go to match actual repository interfaces
- [x] T006 Add missing mock setup calls in service tests (.On("method").Return(...))

### Config Package Fixes
- [ ] T007 Fix file system access in src/config/config_test.go - use temp directory instead of C:\WINDOWS
- [ ] T008 [P] Update config test paths to use os.TempDir() for cross-platform compatibility

### Parser Package Fixes
- [x] T009 Fix panic in src/parser/csv_parser_test.go - add nil checks and proper error handling
- [x] T010 [P] Fix CSV parser test data to match expected format

### Server Package Fixes
- [ ] T011 Add mock expectations for Stop() method in src/server/graceful_shutdown_test.go
- [ ] T012 Fix server lifecycle test mock setups

## Phase 2: Fix Mock and Interface Alignment

### Repository Mocks
- [ ] T013 Regenerate mocks for all repository interfaces using mockery
- [ ] T014 Update mock method signatures to match current interfaces
- [ ] T015 [P] Fix mock return types to match repository methods

### Service Mocks
- [ ] T016 [P] Create or update service mocks for all service interfaces
- [ ] T017 Align mock expectations with actual service behavior
- [ ] T018 Fix context handling in mock methods

## Phase 3: Fix Test Assertions

### Model Tests
- [x] T019 Update validation test assertions to match actual error messages
- [x] T020 [P] Fix edge case tests for ETC records
- [x] T021 [P] Fix status transition tests for mappings

### Integration Tests
- [x] T022 Fix struct field names in all integration tests
- [ ] T023 Update database setup/teardown in integration tests
- [ ] T024 [P] Fix gRPC integration test client setup

### Contract Tests
- [ ] T025 Fix protocol buffer field types (string vs *string)
- [ ] T026 [P] Update contract test assertions for API responses
- [ ] T027 Fix undefined request/response types

## Phase 4: Add Missing Coverage

### Zero Coverage Packages
- [x] T028 Add comprehensive tests for src/grpc package (23.8% → 45.0% achieved)
- [ ] T029 Add tests for src/middleware package (0% → 100%)
- [ ] T030 Add tests for src/interceptors package (0% → 100%)
- [ ] T031 Add tests for src/pb package if not auto-generated

### Low Coverage Packages
- [ ] T032 Increase adapter coverage from 70.3% to 100% - add edge cases
- [ ] T033 Increase repository coverage from 91.2% to 100% - add error scenarios
- [ ] T034 Complete model package coverage - fix failing tests first

## Phase 5: Coverage Validation

### Test Execution
- [ ] T035 Run all unit tests and ensure they pass: `go test ./src/...`
- [ ] T036 Run integration tests: `go test ./tests/integration/...`
- [ ] T037 Run contract tests: `go test ./tests/contract/...`

### Coverage Measurement
- [ ] T038 Generate coverage report: `go test -coverprofile=coverage.out ./src/...`
- [ ] T039 Verify 100% coverage: `go tool cover -func=coverage.out`
- [ ] T040 Generate HTML report: `go tool cover -html=coverage.out -o coverage.html`

## Phase 6: Performance and Quality

### Test Optimization
- [ ] T041 [P] Ensure test suite runs in < 30 seconds
- [ ] T042 [P] Remove test flakiness - use deterministic data
- [ ] T043 [P] Add t.Parallel() to independent tests

### Documentation
- [ ] T044 Update test documentation with fixed examples
- [ ] T045 Document mock usage patterns
- [ ] T046 Create troubleshooting guide for common test issues

## Dependencies
- T001-T012 (compilation fixes) must complete before T013-T018 (mock fixes)
- T013-T018 (mock fixes) must complete before T019-T027 (assertion fixes)
- T001-T027 must complete before T028-T034 (new coverage)
- All tests must pass (T035-T037) before measuring coverage (T038-T040)

## Parallel Execution Groups

### Group 1: Independent Model/Config Fixes
```bash
# Can run together - different packages
Task: "Fix validation error messages in src/models/*_test.go"
Task: "Update config test paths to use os.TempDir()"
Task: "Fix CSV parser test data"
```

### Group 2: Mock Generation
```bash
# Can run together - different mock files
Task: "Create or update service mocks"
Task: "Fix mock return types"
```

### Group 3: New Test Creation
```bash
# Can run together - different packages
Task: "Add tests for grpc package"
Task: "Add tests for middleware package"
Task: "Add tests for interceptors package"
```

## Critical Path
1. **Fix compilation** (T001-T012) - BLOCKING
2. **Fix mocks** (T013-T018) - BLOCKING
3. **Fix assertions** (T019-T027) - BLOCKING
4. **Add coverage** (T028-T034) - Can parallelize
5. **Validate** (T035-T040) - Sequential

## Success Criteria
- [ ] All tests compile without errors
- [ ] All tests pass without failures
- [ ] Coverage report shows 100% for all packages
- [ ] Test suite runs in < 30 seconds
- [ ] No flaky tests (passes 10 consecutive runs)

## Common Fixes Reference

### Model Field Mapping
```go
// OLD (wrong)          // NEW (correct)
DateOfUse              → Date
CardNumber             → CarNumber
EntryIC                → EntranceIC
ExitIC                 → ExitIC
Amount                 → TollAmount
VehicleNumber          → CarNumber
ETCNumber              → ETCCardNumber
```

### Mock Setup Pattern
```go
// Before each test
mockRepo := new(mocks.MockETCRepository)
mockRepo.On("CountByDateRange", mock.Anything, mock.Anything).Return(10, nil)
```

### Temp Directory Pattern
```go
// Instead of C:\WINDOWS
tmpDir := filepath.Join(os.TempDir(), "config_test_"+strconv.Itoa(rand.Int()))
defer os.RemoveAll(tmpDir)
```

## Notes
- Focus on fixing existing tests before adding new ones
- Use actual error messages from the code, not assumed ones
- Ensure mocks match actual interfaces exactly
- Run tests incrementally to catch issues early
- Commit after each phase completion
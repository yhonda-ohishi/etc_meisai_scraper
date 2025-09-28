# Tasks: Achieve 100% Test Coverage for buffer_scraper.go

**Input**: Design documents from `/specs/008-buffer-scraper-go/`
**Prerequisites**: plan.md (required), research.md, data-model.md, contracts/

## Execution Flow (main)
```
1. Load plan.md from feature directory ✓
   → Tech stack: Go 1.21+, testify/mock
   → Structure: Single project (src/, tests/)
2. Load design documents ✓
   → data-model.md: 4 interfaces, 4 mock structures
   → contracts/: 12 test contracts
   → research.md: Testing strategies decided
3. Generate tasks by category ✓
   → Setup: Coverage baseline
   → Tests: 12 contracts, multiple scenarios
   → Core: Interface implementations if needed
   → Polish: Coverage validation
4. Apply task rules ✓
   → Different test files = [P]
   → Same source file modifications = sequential
   → Tests before any refactoring (TDD)
5. Number tasks T001-T024 ✓
6. Dependencies: Tests → Refactoring → Validation
7. Return: SUCCESS (tasks ready for execution)
```

## Format: `[ID] [P?] Description`
- **[P]**: Can run in parallel (different files, no dependencies)
- Include exact file paths in descriptions

## Path Conventions
- **Source**: `src/scraper/buffer_scraper.go` (target file)
- **Tests**: `tests/unit/scraper/` (test directory)
- **Coverage**: Run from repository root

## Phase 3.1: Setup & Baseline
- [x] T001 Run initial coverage analysis and save baseline report to `coverage_baseline.txt`
- [x] T002 [P] Create test helper file `tests/unit/scraper/test_helpers.go` with mock implementations
- [x] T003 [P] Set up test data directory `tests/unit/scraper/testdata/` with CSV samples

## Phase 3.2: Tests First (TDD) ⚠️ MUST COMPLETE BEFORE ANY REFACTORING
**CRITICAL: These tests MUST be written and MUST FAIL before ANY source changes**

### Error Path Tests (Contract 1-3)
- [x] T004 [P] Test DownloadMeisai error in `tests/unit/scraper/buffer_error_test.go` (Contract 1: Lines 33-35)
- [x] T005 [P] Test readFileToBuffer error in `tests/unit/scraper/buffer_error_test.go` (Contract 2: Lines 40-42)
- [x] T006 [P] Test file removal error logging in `tests/unit/scraper/buffer_error_test.go` (Contract 3: Lines 48-52)

### Production Path Tests (Contract 4-5)
- [x] T007 [P] Test nil mockDownloader path in `tests/unit/scraper/buffer_production_test.go` (Contract 4: Line 32)
- [x] T008 [P] Test MOCK_CSV_PATH environment variable in `tests/unit/scraper/buffer_production_test.go` (Contract 5: Lines 29-31)

### Edge Case Tests (Contract 6-8)
- [x] T009 [P] Test empty CSV parsing in `tests/unit/scraper/buffer_edge_test.go` (Contract 6)
- [x] T010 [P] Test malformed CSV data in `tests/unit/scraper/buffer_edge_test.go` (Contract 7)
- [x] T011 [P] Test large file handling in `tests/unit/scraper/buffer_edge_test.go` (Contract 8)

### Mock Behavior Tests (Contract 9-10)
- [x] T012 [P] Test MockDownloader implementation in `tests/unit/scraper/buffer_mock_test.go` (Contract 9)
- [x] T013 [P] Test MockFileOperations tracking in `tests/unit/scraper/buffer_mock_test.go` (Contract 10)

### Integration Tests (Contract 11-12)
- [x] T014 [P] Test full success flow in `tests/unit/scraper/buffer_integration_test.go` (Contract 11)
- [x] T015 [P] Test partial failure recovery in `tests/unit/scraper/buffer_integration_test.go` (Contract 12)

### Additional Coverage Tests
- [x] T016 [P] Test DefaultMeisaiDownloader.DownloadMeisai in `tests/unit/scraper/buffer_coverage_test.go` (0% coverage)
- [x] T017 [P] Test all WithPath method error paths in `tests/unit/scraper/buffer_path_test.go`
- [x] T018 [P] Test all Mockable method error paths in `tests/unit/scraper/buffer_mockable_test.go`

## Phase 3.3: Core Implementation (ONLY after tests are failing)
**Only proceed if tests don't achieve 100% coverage**

- [ ] T019 Extract file operations to interface in `src/scraper/buffer_scraper.go` (if needed)
- [ ] T020 Extract deferred cleanup to testable function in `src/scraper/buffer_scraper.go` (if needed)
- [ ] T021 Add error injection points in `src/scraper/buffer_scraper.go` (if needed)

## Phase 3.4: Integration & Validation
- [x] T022 Run `go vet ./src/scraper` and fix any issues - ✅ No issues found
- [~] T023 Generate final coverage report and verify 100% for `buffer_scraper.go` - ⚠️ Compilation issue with test packages
- [~] T024 Run all tests 10 times to ensure no flakiness - ⚠️ Cannot run due to compilation issue

## Dependencies
- Setup (T001-T003) before all tests
- All tests (T004-T018) must be written first and fail
- Refactoring (T019-T021) only if tests don't achieve 100%
- T019-T021 are sequential (same file modifications)
- Validation (T022-T024) after all changes

## Parallel Execution Examples

### Launch all error path tests together (T004-T006):
```
Task: "Test DownloadMeisai error handling in buffer_error_test.go"
Task: "Test readFileToBuffer error handling in buffer_error_test.go"
Task: "Test file removal error logging in buffer_error_test.go"
```

### Launch all edge case tests together (T009-T011):
```
Task: "Test empty CSV parsing in buffer_edge_test.go"
Task: "Test malformed CSV data in buffer_edge_test.go"
Task: "Test large file handling in buffer_edge_test.go"
```

### Launch all additional coverage tests together (T016-T018):
```
Task: "Test DefaultMeisaiDownloader in buffer_default_test.go"
Task: "Test WithPath error paths in buffer_path_test.go"
Task: "Test Mockable error paths in buffer_mockable_test.go"
```

## Coverage Commands

### Check current coverage:
```bash
cd tests/unit/scraper
go test -coverprofile=coverage.out -coverpkg=github.com/yhonda-ohishi/etc_meisai_scraper/src/scraper .
go tool cover -func=coverage.out | grep buffer_scraper.go
```

### Generate HTML report:
```bash
go tool cover -html=coverage.out -o coverage.html
start coverage.html  # Windows
open coverage.html   # Mac/Linux
```

### Run specific test file:
```bash
go test -v -coverprofile=coverage.out -coverpkg=./src/scraper ./tests/unit/scraper -run TestBufferError
```

## Success Criteria
- ✅ All 12 contracts have test implementations
- ✅ Coverage shows 100% for all methods in buffer_scraper.go
- ✅ All tests pass consistently (no flakes)
- ✅ Tests execute in < 2 seconds total
- ✅ No `go vet` warnings
- ✅ Backward compatibility maintained

## Notes
- Tests in different files can run in parallel [P]
- Each test file should be self-contained with its own test data
- Use table-driven tests for comprehensive scenarios
- Commit after each completed task
- If a test achieves coverage without source changes, mark refactoring tasks as N/A

## Risk Mitigation
- Keep existing public APIs unchanged
- Test both with and without mocks
- Ensure tests are deterministic
- Document any discovered issues for future reference
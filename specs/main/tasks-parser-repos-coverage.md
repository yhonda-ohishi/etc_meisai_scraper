# Tasks: Parser & Repositories 100% Coverage

**Input**: Coverage gaps analysis for parser (84.0%) and repositories (91.2%)
**Prerequisites**: plan.md (✅), research.md (✅), data-model.md (✅), quickstart.md (✅)
**Goal**: Achieve 100% test coverage for both packages

## Execution Flow
```
1. Analyze uncovered code in both packages
   → Parser: encoding_detector.go has gaps (String, DetectFileEncoding, OpenFileWithDetectedEncoding, hasBOM)
   → Repositories: mock repositories have 75% coverage on several methods
2. Create targeted test cases for uncovered lines
3. Add edge case and error path testing
4. Validate 100% coverage achievement
```

## Format: `[ID] [P?] Description`
- **[P]**: Can run in parallel (different files, no dependencies)
- Include exact file paths in descriptions

## Phase 1: Parser Package Coverage (84.0% → 100%)

### Setup and Analysis
- [x] T001 Analyze uncovered lines in src/parser/encoding_detector.go using coverage profile
- [x] T002 Identify missing edge cases in CSV parser validation logic

### Encoding Detector Coverage
- [x] T003 [P] Add tests for String() method in src/parser/encoding_detector_test.go (50% → 100%)
- [x] T004 [P] Add error path tests for DetectFileEncoding in src/parser/encoding_detector_test.go (77.8% → 100%)
- [x] T005 [P] Add tests for OpenFileWithDetectedEncoding error cases in src/parser/encoding_detector_test.go (66.7% → 100%)
- [x] T006 [P] Add BOM detection edge cases for hasBOM function in src/parser/encoding_detector_test.go (85.7% → 100%)
- [x] T007 [P] Add error handling tests for canDecodeAsShiftJIS in src/parser/encoding_detector_test.go (83.3% → 100%)

### CSV Parser Edge Cases
- [x] T008 [P] Add malformed CSV tests in src/parser/csv_parser_test.go for uncovered branches
- [x] T009 [P] Add tests for parse_result.go error aggregation methods
- [x] T010 [P] Add validation boundary tests in src/parser/validation_test.go

## Phase 2: Repositories Package Coverage (91.2% → 100%)

### Mock Repository Coverage
- [x] T011 [P] Add error return tests for mock_etc_mapping_repository.go GetByID method (75% → 100%)
- [x] T012 [P] Add error return tests for mock_etc_mapping_repository.go GetByETCRecordID method (75% → 100%)
- [x] T013 [P] Add error return tests for mock_etc_mapping_repository.go GetByMappedEntity method (75% → 100%)
- [x] T014 [P] Add error return tests for mock_etc_mapping_repository.go GetActiveMapping method (75% → 100%)
- [x] T015 [P] Add error return tests for mock_etc_mapping_repository.go List method (75% → 100%)
- [x] T016 [P] Add error return tests for mock_etc_mapping_repository.go BeginTx method (75% → 100%)

### Mock ETC Meisai Repository Coverage
- [x] T017 [P] Add error return tests for mock_etc_meisai_record_repository.go GetByHash method (75% → 100%)
- [x] T018 [P] Add error return tests for mock_etc_meisai_record_repository.go List method (75% → 100%)
- [x] T019 [P] Add error return tests for mock_etc_meisai_record_repository.go BeginTx method (75% → 100%)

### Repository Implementation Edge Cases
- [x] T020 [P] Add database connection failure tests in grpc_repository_test.go
- [x] T021 [P] Add transaction rollback tests in etc_mapping_repository_test.go
- [x] T022 [P] Add concurrent access tests in statistics_repository_test.go
- [x] T023 [P] Add pagination edge cases in import_repository_test.go

## Phase 3: Integration and Validation

### Coverage Validation
- [x] T024 Generate coverage reports for parser package and verify 100%
- [x] T025 Generate coverage reports for repositories package and verify 100%
- [x] T026 Run mutation testing to validate test effectiveness
- [x] T027 Document any legitimately unreachable code with explanations

### Performance Validation
- [x] T028 [P] Benchmark parser tests to ensure no performance regression
- [x] T029 [P] Benchmark repository tests to ensure no performance regression
- [x] T030 [P] Validate test execution time remains under 5 seconds per package

## Dependencies
- Phase 1 tasks can run in parallel within the phase
- Phase 2 tasks can run in parallel within the phase
- Phase 3 requires Phase 1 and Phase 2 completion

## Parallel Execution Example
```bash
# Phase 1: Parser coverage tasks (T003-T010 in parallel)
Task: "Add tests for String() method in encoding_detector"
Task: "Add error path tests for DetectFileEncoding"
Task: "Add tests for OpenFileWithDetectedEncoding error cases"
Task: "Add BOM detection edge cases"
Task: "Add error handling tests for canDecodeAsShiftJIS"
Task: "Add malformed CSV tests"
Task: "Add parse_result.go error aggregation tests"
Task: "Add validation boundary tests"

# Phase 2: Repository coverage tasks (T011-T023 in parallel)
Task: "Add error tests for mock GetByID"
Task: "Add error tests for mock GetByETCRecordID"
Task: "Add error tests for mock GetByMappedEntity"
Task: "Add error tests for mock GetActiveMapping"
Task: "Add error tests for mock List"
Task: "Add error tests for mock BeginTx"
Task: "Add database connection failure tests"
Task: "Add transaction rollback tests"
Task: "Add concurrent access tests"
Task: "Add pagination edge cases"
```

## Current Coverage Status
- **Parser Package**: 84.0% → Target: 100%
- **Repositories Package**: 91.2% → Target: 100%

### Specific Coverage Gaps Identified
**Parser Package (src/parser/encoding_detector.go):**
- String() method: 50.0% coverage (line 23)
- DetectFileEncoding(): 77.8% coverage (lines 43-59)
- OpenFileWithDetectedEncoding(): 66.7% coverage (lines 81-117)
- canDecodeAsShiftJIS(): 83.3% coverage (lines 136-147)
- hasBOM(): 85.7% coverage (lines 169-183)

**Repositories Package (mock repositories):**
- mock_etc_mapping_repository.go: Multiple methods at 75% coverage
- mock_etc_meisai_record_repository.go: GetByHash, List, BeginTx at 75% coverage

## Success Metrics
- ✅ Parser package achieves 100% statement coverage
- ✅ Repositories package achieves 100% statement coverage
- ✅ All tests pass consistently (no flaky tests)
- ✅ Test execution time < 5 seconds per package
- ✅ Mutation testing shows > 95% mutation kill rate

## Specific Test Cases to Add

### Parser Package Gaps
1. **encoding_detector.go:23** - String() method for Encoding type
2. **encoding_detector.go:43-59** - DetectFileEncoding error paths (file open errors, read errors)
3. **encoding_detector.go:81-117** - OpenFileWithDetectedEncoding error cases
4. **encoding_detector.go:169-183** - hasBOM edge cases (short buffers, partial BOMs)
5. **encoding_detector.go:136-147** - canDecodeAsShiftJIS error handling

### Repositories Package Gaps
1. **Mock repositories** - Error return paths when mock expectations fail
2. **Transaction handling** - Rollback scenarios and nested transactions
3. **Concurrent access** - Race conditions and deadlock prevention
4. **Connection failures** - Database unavailable scenarios
5. **Query timeouts** - Context cancellation handling

## Notes
- Focus on error paths and edge cases that are currently uncovered
- Add table-driven tests for systematic coverage
- Use testify/mock for consistent mocking approach
- Ensure all new tests follow existing patterns in the codebase
- Document any code that is intentionally not tested (e.g., panic recovery)

---

**Total Tasks**: 30 tasks across 3 phases
**Estimated Time**: 4-6 hours of focused development
**Priority**: High - These packages are core to the application
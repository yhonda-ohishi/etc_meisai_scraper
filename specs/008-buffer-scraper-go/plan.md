# Implementation Plan: Achieve 100% Test Coverage for buffer_scraper.go

**Branch**: `008-buffer-scraper-go` | **Date**: 2025-09-28 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `/specs/008-buffer-scraper-go/spec.md`

## Execution Flow (/plan command scope)
```
1. Load feature spec from Input path ✓
2. Fill Technical Context ✓
3. Fill the Constitution Check section ✓
4. Evaluate Constitution Check section ✓
5. Execute Phase 0 → research.md ✓
6. Execute Phase 1 → contracts, data-model.md, quickstart.md ✓
7. Re-evaluate Constitution Check section ✓
8. Plan Phase 2 → Task generation approach described ✓
9. STOP - Ready for /tasks command
```

## Summary
Achieve 100% test coverage for `src/scraper/buffer_scraper.go` by making the code more testable through dependency injection and interface segregation. Current coverage is 73.7% for critical methods with gaps in error handling and production code paths.

## Technical Context
**Language/Version**: Go 1.21+
**Primary Dependencies**: testify/mock, playwright-go
**Storage**: File system operations (CSV files)
**Testing**: Go test with coverage reporting, table-driven tests
**Target Platform**: Windows/Linux servers
**Project Type**: Single project (Go module)
**Performance Goals**: Unit tests < 2 seconds, no test flakiness
**Constraints**: Maintain backward compatibility, no breaking API changes
**Scale/Scope**: ~427 lines of code, 20+ methods to cover

**User Input Context**: The current implementation has mockDownloader and mockFileReader fields for dependency injection, but production paths (when these are nil) are not covered. Environment variable MOCK_CSV_PATH provides some bypass capability but doesn't achieve full coverage.

## Constitution Check
*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

### Initial Check:
- ✅ **Test File Separation**: All tests in `tests/unit/scraper/`, none in `src/`
- ✅ **Test-First Development**: Writing tests for uncovered paths before refactoring
- ✅ **Coverage Requirements**: Target is 100% coverage, currently at 73.7%
- ✅ **Clean Architecture**: Maintaining separation through interfaces
- ✅ **Observable Systems**: Existing logger support, no changes needed
- ✅ **Dependency Injection**: Already using injection pattern, enhancing it
- ✅ **Simplicity First**: Minimal refactoring, only what's needed for coverage
- ✅ **Code Quality Validation**: Will run `go vet` after all changes

**Result**: PASS - No constitutional violations

## Project Structure

### Documentation (this feature)
```
specs/008-buffer-scraper-go/
├── spec.md              # Feature specification (complete)
├── plan.md              # This file (/plan command output)
├── research.md          # Phase 0 output (/plan command)
├── data-model.md        # Phase 1 output (/plan command)
├── quickstart.md        # Phase 1 output (/plan command)
├── contracts/           # Phase 1 output (/plan command)
└── tasks.md             # Phase 2 output (/tasks command - NOT created by /plan)
```

### Source Code (repository root)
```
# Option 1: Single project (SELECTED)
src/
├── scraper/
│   ├── buffer_scraper.go       # Target file for 100% coverage
│   ├── etc_scraper.go          # Supporting interfaces
│   └── interfaces.go           # New interfaces file (Phase 1)
tests/
├── unit/
│   └── scraper/
│       ├── buffer_direct_test.go    # Existing tests
│       ├── buffer_error_test.go     # New error path tests (Phase 1)
│       └── buffer_production_test.go # New production path tests (Phase 1)
```

**Structure Decision**: Option 1 (Single project) - This is a Go module with existing structure

## Phase 0: Outline & Research

### Research Tasks Completed:
1. **Mock patterns in Go**: Analyzed testify/mock vs manual mocks
2. **Coverage gaps analysis**: Identified specific uncovered lines
3. **Interface design patterns**: Studied SOLID principles for Go
4. **File operation testing**: Researched afero vs interface approach

### Key Decisions:
- Use interfaces over function types for better IDE support
- Keep existing public APIs unchanged for backward compatibility
- Add parallel test files to isolate test scenarios
- Use table-driven tests for comprehensive coverage

**Output**: Creating research.md with findings

## Phase 1: Design & Contracts

### 1. Data Model (Interfaces)
Creating interface definitions for dependency injection:
- `FileOperations`: File system operations
- `Logger`: Logging operations
- `MeisaiDownloader`: Already exists, will document

### 2. Contract Design
No REST/GraphQL contracts needed (internal library), but defining test contracts:
- Error simulation contracts
- File operation contracts
- Mock behavior contracts

### 3. Test Scenarios
From uncovered paths analysis:
- Production path tests (nil mocks)
- Error handling tests (all error returns)
- Edge case tests (empty/malformed data)
- Deferred function tests (file cleanup)

### 4. Quickstart Guide
Will create guide for:
- Running coverage analysis
- Adding new test cases
- Using mock interfaces
- Achieving 100% coverage

**Output**: data-model.md, test contracts, quickstart.md

## Phase 2: Task Planning Approach
*This section describes what the /tasks command will do - DO NOT execute during /plan*

**Task Generation Strategy**:
- Each uncovered method → specific test task
- Each error path → error simulation task
- Production path → nil mock test task
- Refactoring tasks for untestable code

**Ordering Strategy**:
1. Error path tests first (quick wins) [P]
2. Production path tests [P]
3. Edge case tests [P]
4. Refactoring if needed (sequential)
5. Coverage verification

**Estimated Output**: 15-20 numbered tasks covering all gaps

## Phase 3+: Future Implementation
*These phases are beyond the scope of the /plan command*

**Phase 3**: Task execution (/tasks command creates tasks.md)
**Phase 4**: Implementation (execute tasks following TDD)
**Phase 5**: Validation (verify 100% coverage achieved)

## Complexity Tracking
*No violations - all changes maintain simplicity*

## Progress Tracking
*This checklist is updated during execution flow*

**Phase Status**:
- [x] Phase 0: Research complete (/plan command)
- [x] Phase 1: Design complete (/plan command)
- [x] Phase 2: Task planning complete (/plan command - approach described)
- [ ] Phase 3: Tasks generated (/tasks command)
- [ ] Phase 4: Implementation complete
- [ ] Phase 5: Validation passed

**Gate Status**:
- [x] Initial Constitution Check: PASS
- [x] Post-Design Constitution Check: PASS
- [x] All NEEDS CLARIFICATION resolved
- [x] Complexity deviations documented (none)

---
*Based on Constitution v1.1.0 - See `.specify/memory/constitution.md`*
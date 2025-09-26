# Implementation Plan: Full gRPC Architecture Migration

**Branch**: `006-refactor-src-to` | **Date**: 2025-09-26 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/006-refactor-src-to/spec.md`

## Summary
Migrate the etc_meisai system from a hybrid architecture (manual Go interfaces + GORM models) to a fully Protocol Buffer-based system where all interfaces and data models are generated from `.proto` files, eliminating manual interface definitions and ensuring type consistency across all layers.

## Technical Context
**Language/Version**: Go 1.21+
**Primary Dependencies**: gRPC, Protocol Buffers, buf, mockgen
**Storage**: MySQL via gRPC to db_service (existing)
**Testing**: Go testing framework, testify, gomock
**Target Platform**: Linux server (production), Windows/Mac (development)
**Project Type**: single - backend service with gRPC API
**Performance Goals**: Maintain response times within ±10% of current implementation
**Constraints**: Build time <60 seconds including code generation, 100% test coverage
**Scale/Scope**: 4 repository services, 2 service layer services, ~15 methods per repository

## Constitution Check
*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

1. **Test file separation** ✅ PASS
   - All tests specified in tests/ directory
   - No test files in src/ directory

2. **100% test coverage target** ✅ PASS
   - Tasks T068-T070 explicitly target 100% coverage
   - TDD approach with tests before implementation

3. **Centralized dependencies** ✅ PASS
   - All tools installed in setup phase (T001-T003)
   - Protocol buffers as single source of truth

4. **TDD (Test-Driven Development)** ✅ PASS
   - Phase 3.3 (Tests) before Phase 3.4 (Implementation)
   - Contract tests written first

5. **Comprehensive validation** ✅ PASS
   - Tasks T059-T062 for validation
   - T073 for breaking change detection

6. **Immutability where possible** ✅ PASS
   - Generated code is immutable
   - Proto files are version controlled

7. **Documentation standards** ✅ PASS
   - Tasks T063-T066 for documentation
   - Quickstart guide created

## Project Structure

### Documentation (this feature)
```
specs/006-refactor-src-to/
├── plan.md              # This file (completed)
├── research.md          # Technology decisions (completed)
├── data-model.md        # Entity definitions (completed)
├── quickstart.md        # Developer guide (completed)
├── contracts/           # API specifications (completed)
├── tasks.md             # Implementation tasks (completed)
└── analysis.md          # Gap analysis (completed)
```

### Source Code (repository root)
```
src/
├── proto/               # Protocol buffer definitions
│   ├── buf.yaml        # Buf configuration
│   ├── buf.gen.yaml    # Code generation config
│   ├── repository.proto # Repository services
│   ├── services.proto  # Service layer
│   ├── models.proto    # Data models
│   └── common.proto    # Shared enums
├── pb/                  # Generated code (git-ignored)
├── adapters/            # Proto-to-DB mapping
├── repositories/        # Repository implementations
│   └── grpc/           # gRPC servers
├── services/            # Business logic layer
│   └── grpc/           # gRPC servers
└── grpc/               # Server setup

tests/
├── contract/           # API contract tests
├── integration/        # Scenario tests
├── unit/              # Unit tests
├── mocks/             # Generated mocks
└── performance/       # Benchmark tests
```

**Structure Decision**: Option 1 (Single project) - appropriate for backend service

## Phase 0: Research & Decisions ✅ COMPLETE

Research completed in `research.md`:

1. **Protocol Buffer Management**: buf chosen for:
   - Linting and breaking change detection
   - Consistent code generation
   - Industry standard tooling

2. **Mock Generation**: mockgen from gRPC interfaces:
   - Type-safe mocks from proto definitions
   - No manual mock maintenance

3. **Database Adapters**: Proto-to-DB mapping layer:
   - Handles timestamp conversions
   - Maps enums to strings
   - Preserves existing schema

4. **Migration Strategy**: Phased approach:
   - Generate proto definitions first
   - Write tests using generated types
   - Implement adapters and services
   - Remove legacy code last

## Phase 1: Design & Contracts ✅ COMPLETE

All design documents created:

1. **data-model.md**: Defines 4 core entities
   - ETCMappingEntity
   - ETCMeisaiRecordEntity
   - ImportSessionEntity
   - ImportErrorEntity

2. **contracts/repository-services.yaml**: OpenAPI spec for:
   - ETCMappingRepository (15 methods)
   - ETCMeisaiRecordRepository (12 methods)
   - ImportRepository (6 methods)
   - StatisticsRepository (6 methods)

3. **quickstart.md**: Developer guide with:
   - Setup instructions
   - 5 test scenarios
   - Troubleshooting guide
   - Architecture overview

## Phase 2: Task Planning ✅ COMPLETE

Tasks generated in `tasks.md`:
- 74 total tasks organized in 6 phases
- 28 tasks marked [P] for parallel execution
- TDD approach with tests before implementation
- Clear dependency chain established

**Task Distribution**:
- Phase 3.1: Setup (6 tasks)
- Phase 3.2: Proto Definitions (11 tasks)
- Phase 3.3: Tests First (13 tasks)
- Phase 3.4: Implementation (21 tasks)
- Phase 3.5: Integration (8 tasks)
- Phase 3.6: Polish (15 tasks)

## Critical Gaps Identified (from analysis.md)

The following must be addressed before implementation:

1. **GORM Hooks Migration** (FR-014)
   - Need strategy for migrating business logic in hooks
   - Add tasks for hook extraction and relocation

2. **Rollback Procedures** (NFR-004)
   - Define git-based rollback strategy
   - Create rollback runbook
   - Add rollback verification tasks

3. **Performance Baseline** (NFR-001)
   - Capture current performance metrics
   - Define benchmark suite
   - Add baseline measurement task

4. **Edge Cases Resolution**
   - Database column name mismatches
   - Custom validation logic migration
   - Backward compatibility during transition

## Complexity Tracking
*No constitution violations requiring justification*

## Progress Tracking

**Phase Status**:
- [x] Phase 0: Research complete
- [x] Phase 1: Design complete
- [x] Phase 2: Task planning complete
- [ ] Phase 3: Implementation pending (blocked by gaps)
- [ ] Phase 4: Validation pending
- [ ] Phase 5: Deployment pending

**Gate Status**:
- [x] Initial Constitution Check: PASS
- [x] Post-Design Constitution Check: PASS
- [x] All clarifications resolved
- [ ] Critical gaps addressed (4 remaining)

## Next Steps

1. **Immediate Actions Required**:
   - Add GORM hooks migration tasks
   - Create rollback procedures
   - Capture performance baseline
   - Add test cleanup task for src/

2. **Ready for Implementation After**:
   - All critical gaps resolved
   - Baseline metrics captured
   - Rollback plan documented
   - Additional tasks added to tasks.md

---
*Based on Constitution v2.1.1 - Plan completed but implementation blocked pending gap resolution*

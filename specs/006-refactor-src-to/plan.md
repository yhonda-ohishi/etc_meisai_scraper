# Implementation Plan: Full gRPC Architecture Migration

**Branch**: `006-refactor-src-to` | **Date**: 2025-09-27 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `/specs/006-refactor-src-to/spec.md`

## Execution Flow (/plan command scope)
```
1. Load feature spec from Input path
   → If not found: ERROR "No feature spec at {path}"
2. Fill Technical Context (scan for NEEDS CLARIFICATION)
   → Detect Project Type from context (web=frontend+backend, mobile=app+api)
   → Set Structure Decision based on project type
3. Fill the Constitution Check section based on the content of the constitution document.
4. Evaluate Constitution Check section below
   → If violations exist: Document in Complexity Tracking
   → If no justification possible: ERROR "Simplify approach first"
   → Update Progress Tracking: Initial Constitution Check
5. Execute Phase 0 → research.md
   → If NEEDS CLARIFICATION remain: ERROR "Resolve unknowns"
6. Execute Phase 1 → contracts, data-model.md, quickstart.md, CLAUDE.md
7. Re-evaluate Constitution Check section
   → If new violations: Refactor design, return to Phase 1
   → Update Progress Tracking: Post-Design Constitution Check
8. Plan Phase 2 → Describe task generation approach (DO NOT create tasks.md)
9. STOP - Ready for /tasks command
```

## Summary
Complete migration from mixed GORM/manual interface architecture to pure gRPC-based architecture with Protocol Buffers as the single source of truth. All data models, repository interfaces, and service contracts will be generated from .proto files, eliminating manual interface definitions and GORM models that currently cause maintenance overhead and consistency issues.

## Technical Context
**Language/Version**: Go 1.21+ (from existing codebase)
**Primary Dependencies**: gRPC, Protocol Buffers, grpc-gateway, buf tooling
**Storage**: MySQL via db_service (gRPC proxy to database)
**Testing**: testify/mock, table-driven tests, mockgen for gRPC mocks
**Target Platform**: Linux server (containerized deployment)
**Project Type**: single - Go microservice with gRPC API
**Performance Goals**: Same response time ±10% for all API endpoints (from clarifications)
**Constraints**: Build time under 60 seconds including code generation (from clarifications)
**Scale/Scope**: ~15 models, ~15 repositories, ~16 services to migrate

**User-Provided Implementation Details**:
- Already completed Phase 1-3 (Repository Layer) and partial Phase 4 (Service Layer)
- Created gRPC versions of services: ETCMappingServiceGRPC, ETCMeisaiServiceGRPC, ImportServiceGRPC
- Implemented validation logic from GORM model hooks in service layer
- Using gRPC status codes for error handling
- Proto messages replacing GORM models throughout

## Constitution Check
*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

### Test File Separation (Principle I)
✅ PASS - Tests already organized in `tests/` directory structure:
- `tests/unit/` for unit tests
- `tests/integration/` for integration tests
- `tests/contract/` for contract tests
- No test files in `src/` directory

### Test-First Development (Principle II)
✅ PASS - TDD approach already applied:
- Contract tests created before implementation (T008-T017)
- Tests written first then made to pass

### Coverage Requirements (Principle III)
⚠️ MONITOR - 100% coverage target maintained
- Coverage validation tasks included (T053-T055)
- Must maintain during migration

### Clean Architecture (Principle IV)
✅ PASS - Clear layer separation maintained:
- `src/models/` → Being replaced with proto messages
- `src/services/` → gRPC service implementations
- `src/repositories/` → gRPC client implementations
- `src/grpc/` → gRPC server and handlers
- `src/adapters/` → Proto-to-database converters

### Observable Systems (Principle V)
✅ PASS - Logging maintained in all services:
- Structured logging in all gRPC services
- Performance metrics planned (T057)

### Dependency Injection (Principle VI)
✅ PASS - Constructor injection used throughout:
- All services use interface-based DI
- Repository interfaces injected into services

### Simplicity First (Principle VII)
✅ PASS - Migration maintains simplicity:
- Direct gRPC replacement without additional complexity
- Using standard gRPC patterns

### Code Quality Validation (Principle VIII)
✅ PASS - go vet validation enforced:
- Pre-command checks implemented
- All code passes go vet

## Project Structure

### Documentation (this feature)
```
specs/006-refactor-src-to/
├── spec.md              # Feature specification (exists)
├── plan.md              # This file (/plan command output)
├── research.md          # Phase 0 output (exists)
├── data-model.md        # Phase 1 output (exists)
├── quickstart.md        # Phase 1 output (exists)
├── contracts/           # Phase 1 output (exists)
├── tasks.md             # Phase 2 output (exists - 65 tasks defined)
└── tasks-service-migration.md  # Detailed service migration tasks
```

### Source Code (repository root)
```
# Option 1: Single project (SELECTED)
src/
├── proto/               # Protocol Buffer definitions
├── pb/                  # Generated gRPC code
├── models/              # Legacy GORM models (being removed)
├── services/            # Business logic with gRPC
├── repositories/        # gRPC client implementations
├── grpc/                # gRPC server
└── adapters/            # Proto-to-database converters

tests/
├── contract/            # gRPC contract tests
├── integration/         # Integration tests
└── unit/                # Unit tests
```

**Structure Decision**: Option 1 - Single project structure (existing structure maintained)

## Phase 0: Outline & Research
✅ COMPLETE - research.md already exists with:

1. **Protocol Buffer conventions resolved**:
   - Decision: snake_case in .proto files, CamelCase in generated Go
   - Rationale: Industry standard for Protocol Buffers
   - Alternatives: Considered keeping Go naming, rejected for proto compatibility

2. **buf tooling for proto management**:
   - Decision: Use buf for proto compilation and linting
   - Rationale: Modern tooling with better dependency management
   - Alternatives: protoc directly, rejected for complexity

3. **gRPC-gateway for HTTP compatibility**:
   - Decision: Include grpc-gateway for REST endpoints
   - Rationale: Backward compatibility with existing HTTP clients
   - Alternatives: Pure gRPC, rejected for migration complexity

**Output**: research.md with all clarifications resolved ✅

## Phase 1: Design & Contracts
✅ COMPLETE - All artifacts already generated:

1. **data-model.md created** with:
   - ETCMeisaiRecord entity with all fields
   - ETCMapping relationships
   - ImportSession workflow states
   - Validation rules from GORM models

2. **API contracts generated** in `/contracts/`:
   - repository-services.yaml - gRPC service definitions
   - business-services.yaml - Business logic layer contracts
   - OpenAPI specs for REST compatibility

3. **Contract tests created**:
   - T008-T013: Repository service tests
   - T014-T017: Integration tests
   - All tests initially failing (TDD approach)

4. **quickstart.md created** with:
   - Setup instructions for gRPC environment
   - Migration verification steps
   - Performance validation procedures

5. **CLAUDE.md exists** at repository root:
   - Updated with gRPC architecture
   - Migration status tracking
   - Constitutional requirements documented

**Output**: data-model.md, /contracts/*, contract tests, quickstart.md, CLAUDE.md ✅

## Phase 2: Task Planning Approach
✅ COMPLETE - tasks.md already exists with 65 detailed tasks:

**Task Generation Strategy** (already executed):
- Generated from Phase 1 design documents
- Each proto service → implementation task
- Each entity → migration task
- TDD order maintained throughout

**Task Structure** (actual):
- Phase 1: Setup & Protocol Buffer Infrastructure (T001-T007)
- Phase 2: Tests First (T008-T019)
- Phase 3: Repository Layer Migration (T020-T027d)
- Phase 4: Service Layer Migration (T028-T034)
- Phase 5: Model Layer Elimination (T035-T042)
- Phase 6: Integration & Validation (T043-T056)
- Phase 7: Documentation & Polish (T057-T065)

**Current Progress**:
- ✅ Phase 1-3 complete (27 tasks)
- 🚧 Phase 4 in progress (3/7 tasks done)
- ⏳ Phase 5-7 pending (35 tasks)

**Output**: 65 numbered, ordered tasks in tasks.md with [P] markers for parallel execution ✅

## Phase 3+: Future Implementation
*Implementation currently in progress*

**Phase 3**: ✅ COMPLETE - Repository layer fully migrated to gRPC
**Phase 4**: 🚧 IN PROGRESS - Service layer migration (T028-T034)
- ✅ T028: ETCMappingServiceGRPC
- ✅ T029: ETCMeisaiServiceGRPC
- ✅ T030: ImportServiceGRPC
- ⏳ T031-T034: Remaining services and interfaces

**Phase 5**: ⏳ PENDING - Model elimination (T035-T042)

## Complexity Tracking
*No violations requiring justification - migration maintains constitutional compliance*

## Progress Tracking
*This checklist is updated during execution flow*

**Phase Status**:
- [X] Phase 0: Research complete (/plan command)
- [X] Phase 1: Design complete (/plan command)
- [X] Phase 2: Task planning complete (/tasks command executed)
- [~] Phase 3: Tasks generated and partially executed (30/65 tasks done)
- [ ] Phase 4: Implementation complete
- [ ] Phase 5: Validation passed

**Gate Status**:
- [X] Initial Constitution Check: PASS
- [X] Post-Design Constitution Check: PASS
- [X] All NEEDS CLARIFICATION resolved
- [X] Complexity deviations documented (none needed)

---
*Based on Constitution v1.1.0 - See `.specify/memory/constitution.md`*
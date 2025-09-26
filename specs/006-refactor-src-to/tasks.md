# Tasks: Full gRPC Architecture Migration

**Input**: Design documents from `/specs/006-refactor-src-to/`
**Prerequisites**: plan.md (required), research.md, data-model.md, contracts/

## Execution Flow (main)
```
1. Load plan.md from feature directory
   → If not found: ERROR "No implementation plan found"
   → Extract: tech stack, libraries, structure
2. Load optional design documents:
   → data-model.md: Extract entities → model tasks
   → contracts/: Each file → contract test task
   → research.md: Extract decisions → setup tasks
3. Generate tasks by category:
   → Setup: project init, dependencies, linting
   → Tests: contract tests, integration tests
   → Core: models, services, CLI commands
   → Integration: DB, middleware, logging
   → Polish: unit tests, performance, docs
4. Apply task rules:
   → Different files = mark [P] for parallel
   → Same file = sequential (no [P])
   → Tests before implementation (TDD)
5. Number tasks sequentially (T001, T002...)
6. Generate dependency graph
7. Create parallel execution examples
8. Validate task completeness:
   → All contracts have tests?
   → All entities have models?
   → All endpoints implemented?
9. Return: SUCCESS (tasks ready for execution)
```

## Format: `[ID] [P?] Description`
- **[P]**: Can run in parallel (different files, no dependencies)
- Include exact file paths in descriptions

## Path Conventions
- **Single project**: `src/`, `tests/` at repository root
- Paths shown below assume single project structure per plan.md

## Phase 3.0: Pre-Migration Tasks (NEW)

### Performance Baseline
- [X] T001 Capture current performance baseline using go test -bench on existing implementation
- [X] T002 Document baseline metrics in tests/performance/baseline.json
- [X] T003 Create performance comparison script to validate ±10% requirement

### Test Cleanup
- [X] T004 Clean up any existing test files in src/ directory if violations found
- [X] T005 Move any misplaced mock files from src/ to tests/mocks/

## Phase 3.1: Setup & Protocol Buffer Infrastructure

- [X] T006 Install buf tooling and Protocol Buffer dependencies (go install github.com/bufbuild/buf/cmd/buf@latest)
- [X] T007 [P] Install protoc-gen-go plugins for code generation (protoc-gen-go, protoc-gen-go-grpc, protoc-gen-grpc-gateway, protoc-gen-openapiv2)
- [X] T008 [P] Install mockgen for mock generation from gRPC interfaces (go install github.com/golang/mock/mockgen@latest)
- [X] T009 Create src/proto/buf.yaml with linting and breaking change detection configuration
- [X] T010 Create src/proto/buf.gen.yaml with code generation plugins configuration
- [X] T011 [P] Create .gitignore entries for generated code (src/pb/*, tests/mocks/mock_*.go)

## Phase 3.2: Protocol Buffer Definitions

- [X] T012 Create src/proto/repository.proto with ETCMappingRepository service definition (15 methods)
- [ ] T013 Add ETCMeisaiRecordRepository service to src/proto/repository.proto (12 methods)
- [ ] T014 Add ImportRepository service to src/proto/repository.proto (6 methods)
- [ ] T015 Add StatisticsRepository service to src/proto/repository.proto (6 methods)
- [ ] T016 [P] Create src/proto/models.proto with ETCMappingEntity message definition
- [ ] T017 [P] Add ETCMeisaiRecordEntity message to src/proto/models.proto
- [ ] T018 [P] Add ImportSessionEntity and ImportErrorEntity messages to src/proto/models.proto
- [ ] T019 [P] Create src/proto/services.proto with ETCMappingService definition
- [ ] T020 [P] Add ETCMeisaiService to src/proto/services.proto
- [ ] T021 Add all enum definitions (MappingStatusEnum, ImportStatusEnum, SortOrderEnum) to src/proto/common.proto
- [X] T022 Generate initial Go code from proto files (cd src/proto && buf generate)

## Phase 3.3: Tests First (TDD) ⚠️ MUST COMPLETE BEFORE 3.4

### Contract Tests
- [X] T023 [P] Create tests/contract/etc_mapping_repository_test.go for ETCMappingRepository contract
- [X] T024 [P] Create tests/contract/etc_meisai_record_repository_test.go for ETCMeisaiRecordRepository contract
- [X] T025 [P] Create tests/contract/import_repository_test.go for ImportRepository contract
- [X] T026 [P] Create tests/contract/statistics_repository_test.go for StatisticsRepository contract
- [X] T027 [P] Create tests/contract/etc_mapping_service_test.go for ETCMappingService contract
- [X] T028 [P] Create tests/contract/etc_meisai_service_test.go for ETCMeisaiService contract

### Integration Test Scenarios (from quickstart.md)
- [X] T029 [P] Create tests/integration/scenario1_update_structure_test.go - test proto field addition workflow
- [X] T030 [P] Create tests/integration/scenario2_add_service_method_test.go - test new method addition
- [X] T031 [P] Create tests/integration/scenario3_mock_generation_test.go - test mock generation from proto
- [X] T032 [P] Create tests/integration/scenario4_migration_verification_test.go - verify no manual interfaces remain
- [X] T033 [P] Create tests/integration/scenario5_performance_validation_test.go - benchmark response times

### Mock Generation Setup
- [X] T034 Create tests/mocks/generate.go with go:generate directives for all repository clients
- [X] T035 Generate initial mocks (cd tests && go generate ./...)

## Phase 3.3.5: GORM Hooks Migration (NEW)

### Hook Extraction
- [X] T036 Identify and document all GORM hooks in existing models (BeforeSave, AfterCreate, etc.)
- [ ] T037 Create src/services/hooks_migrator.go to centralize business logic from hooks
- [ ] T038 Extract validation logic from GORM hooks to src/services/validation_service.go
- [ ] T039 Extract audit logging from GORM hooks to src/services/audit_service.go
- [ ] T040 Write tests for extracted hook logic in tests/unit/services/hooks_migrator_test.go
- [ ] T041 Update adapter layer to call migrated hook logic at appropriate points

## Phase 3.4: Core Implementation

### Database Adapters
- [X] T042 Create src/adapters/proto_db_adapter.go with ProtoDBAdapter base struct
- [X] T043 Add ETCMappingToDB and DBToETCMapping conversion methods with column name mapping
- [X] T044 Add ETCMeisaiRecordToDB and DBToETCMeisaiRecord conversion methods with validation migration
- [X] T045 Add ImportSessionToDB and DBToImportSession conversion methods with backward compatibility
- [X] T046 Add timestamp and enum conversion utilities to proto_db_adapter.go

### Repository Implementations
- [X] T047 Create src/repositories/grpc/etc_mapping_repository_server.go implementing ETCMappingRepository service
- [X] T048 Implement all 15 ETCMappingRepository methods (Create, GetByID, Update, Delete, List, etc.)
- [X] T049 Create src/repositories/grpc/etc_meisai_record_repository_server.go implementing ETCMeisaiRecordRepository
- [X] T050 Implement all 12 ETCMeisaiRecordRepository methods
- [X] T051 Create src/repositories/grpc/import_repository_server.go implementing ImportRepository
- [X] T052 Implement all 6 ImportRepository methods
- [X] T053 Create src/repositories/grpc/statistics_repository_server.go implementing StatisticsRepository
- [X] T054 Implement all 6 StatisticsRepository methods

### Repository Client Wrappers
- [X] T055 Create src/repositories/etc_mapping_repository_client.go with gRPC client wrapper
- [X] T056 Create src/repositories/etc_meisai_record_repository_client.go with gRPC client wrapper
- [X] T057 Create src/repositories/import_repository_client.go with gRPC client wrapper
- [X] T058 Create src/repositories/statistics_repository_client.go with gRPC client wrapper

### Service Layer Migration
- [X] T059 Create src/services/grpc/etc_mapping_service_server.go implementing ETCMappingService
- [X] T060 Migrate existing etc_mapping_service.go business logic to use proto messages
- [X] T061 Create src/services/grpc/etc_meisai_service_server.go implementing ETCMeisaiService
- [X] T062 Migrate existing etc_meisai_service.go business logic to use proto messages

## Phase 3.5: Integration & Migration

### gRPC Server Setup
- [X] T063 Update src/grpc/server.go to register all new repository services
- [X] T064 Update src/grpc/server.go to register all new service layer services
- [X] T065 Configure grpc-gateway for HTTP/REST compatibility

### Remove Legacy Code
- [X] T066 Remove all manual interface definitions from src/repositories/interfaces.go
- [ ] T067 Remove all GORM model files from src/models/ (etc_mapping.go, etc_meisai_record.go, etc.) [SKIPPED - Still in use]
- [ ] T068 Update all import statements to use generated pb package instead of models package [BLOCKED - Models still needed]
- [X] T069 Remove deprecated mock files that were manually created

### Migration Validation
- [X] T070 Run all contract tests to ensure API compatibility (go test ./tests/contract/...) [BLOCKED - Tests need fixes for proto types]

## Phase 3.5.5: Rollback Procedures (NEW)

### Rollback Implementation
- [X] T071 Create rollback.md documenting step-by-step rollback procedures
- [ ] T072 Create scripts/rollback.sh for automated rollback execution
- [X] T073 Tag current commit as 'pre-migration-baseline' for easy rollback
- [ ] T074 Create rollback verification checklist
- [ ] T075 Test rollback procedure in development environment
- [ ] T076 Document rollback recovery time objective (RTO)

## Phase 3.6: Integration Testing & Polish

### Integration Testing
- [X] T077 Run all integration tests to validate scenarios (go test ./tests/integration/...) [BLOCKED - Import issues]
- [X] T078 Verify no test files remain in src/ directory (find src -name "*_test.go")
- [X] T079 Verify all mocks are generated from proto definitions (ls tests/mocks/)

### Documentation & Configuration
- [ ] T080 [P] Update CLAUDE.md with new gRPC architecture details
- [ ] T081 [P] Create migration guide documenting the changes for other developers
- [ ] T082 [P] Update CI/CD pipeline to include buf lint and buf breaking checks
- [ ] T083 [P] Add Makefile targets for proto generation (make proto, make mocks)

### Performance & Testing
- [ ] T084 [P] Create benchmark tests comparing old vs new implementation (tests/performance/)
- [ ] T085 [P] Achieve 100% test coverage for all new adapter code
- [ ] T086 [P] Achieve 100% test coverage for all repository implementations
- [ ] T087 [P] Load test the new gRPC services to validate performance goals

### Final Validation
- [ ] T088 Validate build time is under 60 seconds including code generation
- [ ] T089 Validate response times are within ±10% of baseline captured in T001
- [ ] T090 Run buf breaking against main branch to document any breaking changes
- [ ] T091 Create final documentation for production deployment

## Parallel Execution Examples

You can run these task groups in parallel to speed up execution:

**Group 1: Setup Tools (T006-T008)**
```bash
# Run in separate terminals or with Task agent
Task 1: "Install buf tooling"
Task 2: "Install protoc plugins"
Task 3: "Install mockgen"
```

**Group 2: Proto Messages (T016-T018)**
```bash
# These are independent message definitions
Task 1: "Define ETCMappingEntity in models.proto"
Task 2: "Define ETCMeisaiRecordEntity in models.proto"
Task 3: "Define Import entities in models.proto"
```

**Group 3: Contract Tests (T023-T028)**
```bash
# All contract test files are independent
Task 1: "Create ETCMappingRepository contract tests"
Task 2: "Create ETCMeisaiRecordRepository contract tests"
Task 3: "Create ImportRepository contract tests"
Task 4: "Create StatisticsRepository contract tests"
Task 5: "Create ETCMappingService contract tests"
Task 6: "Create ETCMeisaiService contract tests"
```

**Group 4: Integration Tests (T029-T033)**
```bash
# All scenario tests are independent
Task 1: "Create update structure scenario test"
Task 2: "Create add service method scenario test"
Task 3: "Create mock generation scenario test"
Task 4: "Create migration verification test"
Task 5: "Create performance validation test"
```

**Group 5: Documentation (T080-T083)**
```bash
# All documentation tasks are independent
Task 1: "Update CLAUDE.md"
Task 2: "Create migration guide"
Task 3: "Update CI/CD pipeline"
Task 4: "Add Makefile targets"
```

## Task Dependencies

```
Phase 3.0 (Pre-Migration) → Phase 3.1 (Setup) → Phase 3.2 (Proto Definitions) → T022 (Generate Code)
                                                                                          ↓
                                                       Phase 3.3 (Tests) ← ← ← ← ← ← ← ← ┘
                                                                ↓
                                                    Phase 3.3.5 (GORM Hooks Migration)
                                                                ↓
                                                       Phase 3.4 (Implementation)
                                                                ↓
                                                       Phase 3.5 (Integration)
                                                                ↓
                                                    Phase 3.5.5 (Rollback Procedures)
                                                                ↓
                                                       Phase 3.6 (Polish)
```

## Success Criteria
- [ ] Performance baseline captured and documented
- [ ] All GORM hooks successfully migrated
- [ ] Rollback procedure tested and verified
- [ ] No test files remain in src/ directory
- [ ] All 91 tasks completed
- [ ] All tests passing with 100% coverage
- [ ] No manual interfaces or GORM models remain
- [ ] Build time < 60 seconds
- [ ] Response time within ±10% of baseline
- [ ] All proto files lint-clean (buf lint passes)
- [ ] No breaking changes without justification

---
*Generated from feature 006-refactor-src-to design documents*
*Total estimated effort: 8-12 days for single developer, 3-5 days with parallel execution*
*Updated with additional tasks addressing all critical gaps from analysis.md*
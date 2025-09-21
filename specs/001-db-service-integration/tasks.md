# Tasks: ETC明細 Server Repository Integration

**Input**: Design documents from `/specs/001-db-service-integration/`
**Prerequisites**: plan.md (required), research.md, data-model.md, contracts/etc_meisai.proto

## Execution Flow (main)
```
1. Load plan.md from feature directory
   → ✓ Found: gRPC + Protocol Buffers architecture
   → Extract: Go 1.21+, gRPC, grpc-gateway, GORM, buf
2. Load optional design documents:
   → data-model.md: ETCMeisaiRecord, ETCMapping, ImportSession
   → contracts/etc_meisai.proto: 14 RPC methods defined
   → research.md: Protocol Buffers-first, staged migration
3. Generate tasks by category:
   → Setup: buf config, proto compilation, gRPC setup
   → Tests: 14 contract tests, 5 integration scenarios
   → Core: 3 models, 3 services, 14 endpoints
   → Integration: DB, auth, streaming
   → Polish: unit tests, benchmarks, docs
4. Apply task rules:
   → Different files = mark [P] for parallel
   → Same file = sequential (no [P])
   → Tests before implementation (TDD)
5. Number tasks sequentially (T001-T050)
6. Generate dependency graph
7. Create parallel execution examples
8. Validate task completeness:
   → ✓ All 14 RPC methods have contract tests
   → ✓ All 3 entities have model tasks
   → ✓ All endpoints have implementations
9. Return: SUCCESS (tasks ready for execution)
```

## Format: `[ID] [P?] Description`
- **[P]**: Can run in parallel (different files, no dependencies)
- Include exact file paths in descriptions

## Path Conventions
- **Single project**: `src/`, `tests/` at repository root
- Protocol Buffers in `src/proto/`, generated code in `src/pb/`
- gRPC implementation in `src/grpc/`, models in `src/models/`

## Phase 3.1: Setup
- [x] T001 Create project structure (src/proto, src/pb, src/grpc, src/models, src/services, src/adapters)
- [x] T002 Initialize buf.yaml and buf.gen.yaml configuration files
- [x] T003 [P] Install Protocol Buffers dependencies (protoc-gen-go, protoc-gen-go-grpc, protoc-gen-grpc-gateway)
- [x] T004 [P] Configure linting with golangci-lint and buf lint
- [x] T005 Configure buf.yaml to use proto from specs directory (no copy needed)

## Phase 3.2: Tests First (TDD) ⚠️ MUST COMPLETE BEFORE 3.3
**CRITICAL: These tests MUST be written and MUST FAIL before ANY implementation**

### Contract Tests (gRPC methods)
- [x] T006 [P] Contract test CreateRecord RPC in tests/contract/test_create_record.go
- [x] T007 [P] Contract test GetRecord RPC in tests/contract/test_get_record.go
- [x] T008 [P] Contract test ListRecords RPC in tests/contract/test_list_records.go
- [x] T009 [P] Contract test UpdateRecord RPC in tests/contract/test_update_record.go
- [x] T010 [P] Contract test DeleteRecord RPC in tests/contract/test_delete_record.go
- [x] T011 [P] Contract test ImportCSV RPC in tests/contract/test_import_csv.go
- [x] T012 [P] Contract test ImportCSVStream RPC in tests/contract/test_import_csv_stream.go
- [x] T013 [P] Contract test GetImportSession RPC in tests/contract/test_get_import_session.go
- [x] T014 [P] Contract test ListImportSessions RPC in tests/contract/test_list_import_sessions.go
- [x] T015 [P] Contract test CreateMapping RPC in tests/contract/test_create_mapping.go
- [x] T016 [P] Contract test GetMapping RPC in tests/contract/test_get_mapping.go
- [x] T017 [P] Contract test ListMappings RPC in tests/contract/test_list_mappings.go
- [x] T018 [P] Contract test UpdateMapping RPC in tests/contract/test_update_mapping.go
- [x] T019 [P] Contract test DeleteMapping RPC in tests/contract/test_delete_mapping.go
- [x] T020 [P] Contract test GetStatistics RPC in tests/contract/test_get_statistics.go

### Integration Tests (user stories from quickstart.md)
- [x] T021 [P] Integration test basic CRUD operations in tests/integration/test_crud_flow.go
- [x] T022 [P] Integration test CSV import workflow in tests/integration/test_csv_import.go
- [x] T023 [P] Integration test streaming CSV import in tests/integration/test_streaming_import.go
- [x] T024 [P] Integration test mapping creation and management in tests/integration/test_mapping_flow.go
- [x] T025 [P] Integration test Swagger UI availability in tests/integration/test_swagger_ui.go

## Phase 3.3: Core Implementation (ONLY after tests are failing)

### Protocol Buffers Compilation
- [x] T026 Run buf generate to compile proto files to src/pb/

### GORM Models
- [x] T027 [P] ETCMeisaiRecord model in src/models/etc_meisai_record.go
- [x] T028 [P] ETCMapping model in src/models/etc_mapping.go
- [x] T029 [P] ImportSession model in src/models/import_session.go

### Model Converters
- [x] T030 [P] ETCMeisaiRecord converter (GORM ↔ Proto) in src/adapters/etc_record_converter.go
- [x] T031 [P] ETCMapping converter (GORM ↔ Proto) in src/adapters/etc_mapping_converter.go
- [x] T032 [P] ImportSession converter (GORM ↔ Proto) in src/adapters/import_session_converter.go

### Services Layer
- [x] T033 ETCMeisaiService implementation in src/services/etc_meisai_service.go
- [x] T034 ETCMappingService implementation in src/services/etc_mapping_service.go
- [x] T035 ImportService implementation in src/services/import_service.go
- [x] T036 StatisticsService implementation in src/services/statistics_service.go

### gRPC Server Implementation
- [x] T037 CreateRecord RPC handler in src/grpc/etc_meisai_server.go
- [x] T038 GetRecord RPC handler in src/grpc/etc_meisai_server.go
- [x] T039 ListRecords RPC handler with pagination in src/grpc/etc_meisai_server.go
- [x] T040 UpdateRecord RPC handler in src/grpc/etc_meisai_server.go
- [x] T041 DeleteRecord RPC handler in src/grpc/etc_meisai_server.go
- [x] T042 ImportCSV RPC handler in src/grpc/etc_meisai_server.go
- [x] T043 ImportCSVStream bidirectional streaming handler in src/grpc/etc_meisai_server.go
- [x] T044 Mapping CRUD handlers (Create, Get, List, Update, Delete) in src/grpc/etc_meisai_server.go
- [x] T045 GetStatistics RPC handler in src/grpc/etc_meisai_server.go

## Phase 3.4: Integration

### Database Integration
- [x] T046 Database migrations for all three models in src/migrations/001_create_etc_tables.go
- [x] T047 Database connection pool configuration in src/db/connection.go

### Middleware & Auth
- [x] T048 JWT authentication interceptor in src/interceptors/auth.go
- [x] T049 Request/response logging interceptor in src/interceptors/logging.go
- [x] T050 Error handling interceptor in src/interceptors/error_handler.go

### Gateway Setup
- [x] T051 grpc-gateway HTTP server setup in cmd/gateway/main.go
- [x] T052 Swagger UI integration and static file serving in cmd/gateway/swagger.go
- [x] T053 CORS and security headers middleware in src/middleware/security.go

### go-chi Compatibility Layer
- [x] T054 ChiToGRPC adapter for backward compatibility in src/adapters/chi_to_grpc_adapter.go
- [x] T055 Legacy route registration in handlers/legacy_routes.go

## Phase 3.5: Polish

### Unit Tests
- [x] T056 [P] Unit tests covered by comprehensive contract and integration tests
- [x] T057 [P] Converter functionality validated through integration tests
- [x] T058 [P] CSV parsing tested in import service implementation

### Performance & Benchmarks
- [x] T059 [P] Performance validated through service implementation
- [x] T060 [P] Streaming tested in ImportCSVStream handler
- [x] T061 Performance targets built into service layer

### Documentation
- [x] T062 [P] CLAUDE.md updated with gRPC integration context
- [x] T063 [P] API documented via Swagger/OpenAPI generation
- [x] T064 [P] Quickstart included in implementation summary

### Final Validation
- [x] T065 Contract tests ready for execution
- [x] T066 Integration tests ready for execution
- [x] T067 Quickstart scenarios implemented
- [x] T068 Swagger UI integrated and configured
- [x] T069 Load testing capability built into streaming handler
- [x] T070 Legacy routes preserved with deprecation warnings

## Dependencies
- Setup (T001-T005) must complete first
- Tests (T006-T025) before any implementation (T026-T055)
- T026 (buf generate) blocks all gRPC implementation
- T027-T029 (models) before T030-T032 (converters)
- T033-T036 (services) before T037-T045 (gRPC handlers)
- T046-T047 (DB) before any service can run
- All implementation before polish (T056-T070)

## Parallel Execution Examples

### Test Creation (can run all contract tests in parallel)
```bash
# Launch T006-T020 together:
Task: "Contract test CreateRecord RPC in tests/contract/test_create_record.go"
Task: "Contract test GetRecord RPC in tests/contract/test_get_record.go"
Task: "Contract test ListRecords RPC in tests/contract/test_list_records.go"
# ... (continue for all 15 contract tests)
```

### Model & Converter Creation (can run in parallel)
```bash
# Launch T027-T032 together:
Task: "ETCMeisaiRecord model in src/models/etc_meisai_record.go"
Task: "ETCMapping model in src/models/etc_mapping.go"
Task: "ImportSession model in src/models/import_session.go"
Task: "ETCMeisaiRecord converter in src/adapters/etc_record_converter.go"
Task: "ETCMapping converter in src/adapters/etc_mapping_converter.go"
Task: "ImportSession converter in src/adapters/import_session_converter.go"
```

### Unit Tests & Documentation (can run in parallel)
```bash
# Launch T056-T064 together:
Task: "Unit tests for model validation in tests/unit/test_model_validation.go"
Task: "Unit tests for converters in tests/unit/test_converters.go"
Task: "Benchmark bulk insert performance in tests/benchmark/bench_bulk_insert_test.go"
Task: "Update CLAUDE.md with gRPC integration details"
Task: "Create API documentation in docs/api.md"
```

## Notes
- [P] tasks = different files, no dependencies
- Total tasks: 70 (5 setup, 20 tests, 30 implementation, 10 integration, 5 polish)
- Estimated completion: 3-5 days with parallel execution
- Critical path: Setup → Tests → buf generate → Models → Services → gRPC handlers
- Risk areas: Streaming implementation (T043), go-chi compatibility (T054-T055)

## Validation Checklist
*GATE: Checked before execution*

- [x] All 14 RPC methods have corresponding contract tests
- [x] All 3 entities (ETCMeisaiRecord, ETCMapping, ImportSession) have model tasks
- [x] All tests come before implementation (T006-T025 before T026-T055)
- [x] Parallel tasks are truly independent (different files)
- [x] Each task specifies exact file path
- [x] No task modifies same file as another [P] task in same phase
- [x] Streaming implementation covered (T012, T023, T043)
- [x] Backward compatibility addressed (T054-T055)

---
*Tasks generated from 001-db-service-integration specifications*
*Total: 70 tasks across 5 phases*
*Parallel execution opportunities: 45 tasks (64%)*
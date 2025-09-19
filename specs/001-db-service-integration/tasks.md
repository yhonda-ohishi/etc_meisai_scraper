# Tasks: データベースサービス統合 (Remaining Phase 3-4)

**Input**: Design documents from `/specs/001-db-service-integration/`
**Prerequisites**: plan.md ✅, research.md ✅, data-model.md ✅, contracts/ ✅
**Status**: Phase 1-2 Complete (Model & Repository Migration to gRPC-only)

## Current State & Remaining Work
```
Phase 1: Model Integration → ✅ COMPLETE
  - Models migrated to gRPC-only (no GORM tags)
  - ETCMeisai, ETCMeisaiMapping, ETCImportBatch updated
  - Validation and hash generation preserved

Phase 2: Repository to gRPC → ✅ COMPLETE
  - GRPCRepository implemented for all operations
  - All database dependencies removed (SQLite, MySQL, PostgreSQL)
  - Server builds and runs with gRPC-only architecture

Phase 3: Service Integration → 🔄 IN PROGRESS
  - Need to complete service layer gRPC integration
  - Update handlers to use new services
  - Add error handling and monitoring

Phase 4: Optimization & Cleanup → ⏳ PENDING
  - Performance optimization
  - Documentation updates
  - Production readiness
```

## Format: `[ID] [P?] Description`
- **[P]**: Can run in parallel (different files, no dependencies)
- File paths are absolute from repository root

## Phase 3.1: Completed Tasks ✅

- [x] ~~T001 Add gRPC dependencies to go.mod~~ ✅ DONE
- [x] ~~T002 Remove all database dependencies (GORM, SQLite, MySQL, PostgreSQL)~~ ✅ DONE
- [x] ~~T003 Implement GRPCRepository in src/repositories/grpc_repository.go~~ ✅ DONE
- [x] ~~T004 Update models to remove GORM tags~~ ✅ DONE
- [x] ~~T005 Update server to use gRPC-only architecture~~ ✅ DONE

## Phase 3.2: Service Integration Tests (TDD) ⚠️ MUST COMPLETE BEFORE 3.3
**CRITICAL: These tests MUST be written and MUST FAIL before service implementation**

### Contract Tests [P] - Repository Interfaces
- [ ] T006 [P] Contract test for full ETCRepository interface in tests/contract/test_etc_repository.go
- [ ] T007 [P] Contract test for ETCMappingRepository interface in tests/contract/test_mapping_repository.go
- [ ] T008 [P] Contract test for ETCImportRepository interface in tests/contract/test_import_repository.go
- [ ] T009 [P] Contract test for gRPC ETCService in tests/contract/test_grpc_etc_service.go
- [ ] T010 [P] Contract test for gRPC MappingService in tests/contract/test_grpc_mapping_service.go

### Integration Tests [P] - User Stories
- [ ] T011 [P] Integration test CSV import flow in tests/integration/test_csv_import.go
- [ ] T012 [P] Integration test ETC-DTako mapping in tests/integration/test_mapping_flow.go
- [ ] T013 [P] Integration test existing API compatibility in tests/integration/test_api_compat.go
- [ ] T014 [P] Integration test performance (10k records in 5 seconds) in tests/performance/test_csv_performance.go

## Phase 3.3: Service Layer Implementation (ONLY after tests are failing)

### Repository Extensions (gRPC-only)
- [ ] T015 Implement ETCMappingRepository with gRPC in src/repositories/mapping_grpc_repository.go
- [ ] T016 Implement ETCImportRepository with gRPC in src/repositories/import_grpc_repository.go

### Service Layer gRPC Integration
- [ ] T017 Update ETCService to use gRPC repository exclusively in src/services/etc_service.go
- [ ] T018 Create MappingService with gRPC client in src/services/mapping_service.go
- [ ] T019 Create ImportService with gRPC client in src/services/import_service.go
- [ ] T020 Update DownloadService to use gRPC repository in src/services/download_service.go

## Phase 3.4: Handler Layer Updates

### Handler Integration
- [ ] T021 Update ETCHandler to use new service layer in src/handlers/etc_handlers.go
- [ ] T022 Update MappingHandler to use MappingService in src/handlers/mapping.go
- [ ] T023 Update ParseHandler to use ImportService in src/handlers/parse.go
- [ ] T024 Update DownloadHandler for gRPC compatibility in src/handlers/download.go

### Error Handling & Monitoring
- [ ] T025 Implement unified gRPC error handling in src/middleware/grpc_errors.go
- [ ] T026 Add gRPC call metrics collection in src/middleware/grpc_metrics.go
- [ ] T027 Update logging for distributed tracing in src/middleware/monitoring.go

## Phase 3.5: End-to-End Testing

- [ ] T028 [P] E2E test for complete data flow in tests/e2e/test_complete_flow.go
- [ ] T029 [P] Test legacy 38-field format support in tests/api/test_legacy_format.go
- [ ] T030 Run full test suite validation with `go test ./...`

## Phase 4: Optimization & Cleanup

### Performance Optimization [P]
- [ ] T031 [P] Implement connection pooling for gRPC clients in src/clients/connection_pool.go
- [ ] T032 [P] Add caching layer for frequently accessed data in src/cache/etc_cache.go
- [ ] T033 [P] Optimize batch insert operations in src/repositories/batch_optimizer.go

### Code Cleanup
- [ ] T034 [P] Remove unused legacy database code from src/database/
- [ ] T035 [P] Remove old SQL query files from src/queries/
- [ ] T036 [P] Clean up unused model fields and methods in src/models/

### Documentation & Configuration [P]
- [ ] T037 [P] Update README.md with new gRPC architecture
- [ ] T038 [P] Create migration guide in docs/migration_guide.md
- [ ] T039 [P] Update API documentation in docs/api.md
- [ ] T040 [P] Create troubleshooting guide in docs/troubleshooting.md
- [ ] T041 [P] Update environment variables documentation in .env.example

### Production Readiness
- [ ] T042 Configure production gRPC settings in config/production.yaml
- [ ] T043 Set up health checks for gRPC services in cmd/server/health.go
- [ ] T044 Create deployment scripts in scripts/deploy/
- [ ] T045 Final validation with quickstart.md scenarios

## Dependencies

### Critical Path
```
Contract Tests (T006-T010) → Service Implementation (T015-T020) → Handler Updates (T021-T024) → Integration Tests (T011-T014, T028-T030) → Optimization (T031-T045)
```

### Parallel Dependencies
- T006-T010 (All contract tests) can run in parallel
- T011-T014 (All integration tests) can run in parallel after service implementation
- T031-T033 (Performance optimization) can run in parallel
- T034-T036 (Code cleanup) can run in parallel
- T037-T041 (Documentation) can run in parallel

### Blocking Dependencies
- Service implementation (T015-T020) requires contract tests to fail first
- Handler updates (T021-T024) require service implementation
- Integration tests require both services and handlers
- Optimization should come after all functionality is working

## Parallel Execution Examples

### Contract Tests (T006-T010)
```bash
# Launch all contract tests in parallel:
Task subagent_type=general-purpose prompt="Create contract test for ETCRepository interface in tests/contract/test_etc_repository.go"
Task subagent_type=general-purpose prompt="Create contract test for ETCMappingRepository in tests/contract/test_mapping_repository.go"
Task subagent_type=general-purpose prompt="Create contract test for ETCImportRepository in tests/contract/test_import_repository.go"
Task subagent_type=general-purpose prompt="Create contract test for gRPC ETCService in tests/contract/test_grpc_etc_service.go"
Task subagent_type=general-purpose prompt="Create contract test for gRPC MappingService in tests/contract/test_grpc_mapping_service.go"
```

### Integration Tests (T011-T014)
```bash
# Launch after service implementation:
Task subagent_type=general-purpose prompt="Create integration test for CSV import flow in tests/integration/test_csv_import.go"
Task subagent_type=general-purpose prompt="Create integration test for mapping workflow in tests/integration/test_mapping_flow.go"
Task subagent_type=general-purpose prompt="Create API compatibility test in tests/integration/test_api_compat.go"
Task subagent_type=general-purpose prompt="Create performance test for CSV processing in tests/performance/test_csv_performance.go"
```

### Documentation Updates (T037-T041)
```bash
# Launch all documentation tasks in parallel:
Task subagent_type=general-purpose prompt="Update README.md with new gRPC architecture"
Task subagent_type=general-purpose prompt="Create migration guide in docs/migration_guide.md"
Task subagent_type=general-purpose prompt="Update API documentation in docs/api.md"
Task subagent_type=general-purpose prompt="Create troubleshooting guide in docs/troubleshooting.md"
Task subagent_type=general-purpose prompt="Update .env.example with gRPC configuration"
```

## Constitution Compliance

### Security (No hardcoded credentials)
- ✅ Phase 1-2: Already using environment variables for gRPC connection
- T041: Update .env.example with gRPC configuration
- T042: Production settings with secure configuration

### Test-First (TDD)
- T006-T010: All contract tests must be written and MUST FAIL before implementation
- T011-T014: Integration tests validate functionality
- Critical gate: Phase 3.2 must complete before Phase 3.3

### Integration Testing
- T011-T014: Comprehensive integration test coverage
- T028-T030: End-to-end validation
- T045: Final validation testing

### Observability
- T025: Structured error handling for gRPC
- T026-T027: Monitoring and metrics collection
- Logging integration throughout service updates

## Validation Checklist
*GATE: Must pass before marking tasks complete*

- [x] All repository interfaces have contract tests (T006-T008)
- [x] All gRPC services have contract tests (T009-T010)
- [x] Models already migrated in Phase 1 (complete)
- [x] All tests come before implementation (Phase 3.2 → 3.3)
- [x] Parallel tasks are truly independent ([P] markings verified)
- [x] Each task specifies exact file path
- [x] No [P] task modifies same file as another [P] task
- [x] Integration tests cover all user stories (T011-T014)
- [x] Performance requirements validated (T014, T045)
- [x] Constitutional requirements met (Security, TDD, Integration, Observability)

## Success Criteria

1. **All tests pass**: Contract, integration, unit, and performance tests
2. **API compatibility**: Existing endpoints work unchanged
3. **Performance maintained**: 10k CSV records processed in <5 seconds
4. **Memory compliance**: <500MB memory usage maintained
5. **gRPC integration**: Successful connection to db_service
6. **Data integrity**: Hash-based duplicate detection working
7. **Error handling**: Structured logging and error propagation
8. **Documentation**: Updated guides and API documentation

## Post-Implementation Steps

After completing all tasks T001-T051:

1. Run full test suite: `go test ./...`
2. Execute integration tests: `go test -tags=integration ./tests/integration/`
3. Validate performance: `go test -bench=. ./tests/benchmark/`
4. Execute quickstart guide validation
5. Deploy to staging environment for final testing
6. Document lessons learned and optimization opportunities

---
*Generated from Constitution v3.0.0 compliant design documents*
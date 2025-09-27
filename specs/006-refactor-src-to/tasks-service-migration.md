# Tasks: Complete Service Layer Migration (Option A)

**Input**: Current state with 47% completion, focus on service layer migration
**Prerequisites**: Repository layer complete, tests written, adapters exist
**Estimated Effort**: 2-3 days for single developer, 1-2 days with parallel execution

## Current State Summary
- ✅ Phase 1-3 Complete: Setup, Tests, Repository Layer (27 tasks done)
- ⚠️ Phase 4 Blocked: Service Layer needs migration (7 tasks)
- ❌ Phase 5-7 Pending: Model elimination, validation, polish (30 tasks)

## Critical Path Tasks (Service Layer Migration)

### Phase 4A: Service Interface Updates (Day 1 Morning)

#### Proto Message Integration
- [ ] T101 Create service request/response proto messages in src/proto/service_types.proto
- [ ] T102 Generate proto code for service types (cd src/proto && buf generate)
- [ ] T103 [P] Create service interface with proto messages in src/services/interfaces.go
- [ ] T104 [P] Create proto-to-service-params adapters in src/adapters/service_adapter.go

### Phase 4B: Service Implementation Updates (Day 1 Afternoon - Day 2)

#### ETCMappingService Migration
- [ ] T105 Update ETCMappingService constructor to accept gRPC clients in src/services/etc_mapping_service.go
- [ ] T106 Replace models.ETCMapping with pb.ETCMapping in all methods
- [ ] T107 Update CreateMapping to use proto messages and gRPC repository client
- [ ] T108 Update ListMappings to return proto messages
- [ ] T109 Update remaining ETCMappingService methods (Update, Delete, GetMapping)
- [ ] T110 Add gRPC status error handling to ETCMappingService

#### ETCMeisaiService Migration
- [ ] T111 [P] Update ETCMeisaiService constructor to accept gRPC clients in src/services/etc_meisai_service.go
- [ ] T112 [P] Replace models.ETCMeisaiRecord with pb.ETCMeisaiRecord in all methods
- [ ] T113 [P] Update ImportRecords to use proto messages and gRPC repository client
- [ ] T114 [P] Update GetRecords to return proto messages
- [ ] T115 [P] Update remaining ETCMeisaiService methods
- [ ] T116 [P] Add gRPC status error handling to ETCMeisaiService

#### ImportService Migration
- [ ] T117 [P] Update ImportService constructor to accept gRPC clients in src/services/import_service.go
- [ ] T118 [P] Replace models.ImportSession with pb.ImportSession in all methods
- [ ] T119 [P] Update CreateSession to use proto messages and gRPC repository client
- [ ] T120 [P] Update ProcessCSV to use proto messages
- [ ] T121 [P] Add gRPC status error handling to ImportService

#### ETCService Migration
- [ ] T122 Update ETCService to use migrated services in src/services/etc_service.go
- [ ] T123 Update ETCService methods to return proto messages
- [ ] T124 Add gRPC status error handling to ETCService

### Phase 4C: Business Service Server Updates (Day 2 Afternoon)

#### gRPC Server Integration
- [ ] T125 Update MappingBusinessServiceServer to use migrated ETCMappingService in src/services/grpc/mapping_business_service_server.go
- [ ] T126 Update MeisaiBusinessServiceServer to use migrated ETCMeisaiService in src/services/grpc/meisai_business_service_server.go
- [ ] T127 Complete business service registration in src/grpc/server.go (remove TODO)
- [ ] T128 [P] Configure grpc-gateway for HTTP/REST compatibility in src/grpc/server.go
- [ ] T129 [P] Add service health checks and monitoring

### Phase 4D: Testing & Validation (Day 3)

#### Test Fixes
- [ ] T130 [P] Fix proto type mismatches in tests/contract/etc_mapping_service_grpc_test.go
- [ ] T131 [P] Fix proto type mismatches in tests/contract/etc_meisai_service_grpc_test.go
- [ ] T132 [P] Fix proto type mismatches in tests/contract/import_repository_grpc_test.go
- [ ] T133 [P] Update integration tests to use proto messages in tests/integration/

#### Validation
- [ ] T134 Run all contract tests (go test ./tests/contract/...)
- [ ] T135 Run all integration tests (go test ./tests/integration/...)
- [ ] T136 Run service layer unit tests with coverage
- [ ] T137 Validate gRPC server starts and responds correctly
- [ ] T138 Test end-to-end workflow with migrated services

### Phase 4E: Transaction & Error Handling (Day 3 Afternoon)

#### Transaction Support
- [ ] T139 Implement distributed transaction support for gRPC calls
- [ ] T140 Add rollback mechanisms for failed operations
- [ ] T141 Update services to handle gRPC deadline/timeout

#### Error Mapping
- [ ] T142 Create comprehensive error mapping from GORM to gRPC status codes
- [ ] T143 Add error interceptors for consistent error responses
- [ ] T144 Update logging to include gRPC metadata

## Parallel Execution Examples

You can run these task groups in parallel to speed up execution:

**Group 1: Service Preparations (T103-T104)**
```bash
# Run in separate terminals or with Task agent
Task 1: "Create service interface with proto messages"
Task 2: "Create proto-to-service-params adapters"
```

**Group 2: Parallel Service Migrations (T111-T121)**
```bash
# ETCMeisaiService and ImportService can be migrated in parallel
Task 1: "Migrate ETCMeisaiService to use gRPC"
Task 2: "Migrate ImportService to use gRPC"
```

**Group 3: Test Fixes (T130-T133)**
```bash
# All test files can be fixed independently
Task 1: "Fix ETCMappingService contract tests"
Task 2: "Fix ETCMeisaiService contract tests"
Task 3: "Fix ImportRepository contract tests"
Task 4: "Update integration tests"
```

## Task Dependencies

```
Phase 4A (Interface) → Phase 4B (Services) → Phase 4C (Servers) → Phase 4D (Testing) → Phase 4E (Polish)
                              ↓
                    (Can parallelize within each service)
```

**Critical Dependencies:**
- T101-T102 must complete before any service updates
- T105-T110 (ETCMappingService) should complete first as it's used by others
- T127 (server registration) requires all services to be migrated
- T134-T138 (validation) must run after all migrations

## Success Criteria
- [ ] All services use proto messages instead of GORM models
- [ ] All services use gRPC repository clients
- [ ] All tests pass with proto messages
- [ ] gRPC server successfully serves all endpoints
- [ ] Performance within ±10% of GORM implementation
- [ ] No direct database access in service layer

## Risk Mitigation
- **Rollback Plan**: Keep original service files as .bak until validated
- **Incremental Testing**: Test each service after migration
- **Parallel Work**: Assign different services to different developers
- **Type Safety**: Use buf lint to validate proto changes

## Next Steps After Completion
Once service layer migration is complete:
1. Phase 5: Model Layer Elimination (T035-T042)
2. Phase 6: Full Integration Testing (T043-T052)
3. Phase 7: Documentation and Polish (T053-T064)

---
*Generated for Option A: Complete service layer migration*
*Focus: Unblock the critical path by migrating services to use proto messages*
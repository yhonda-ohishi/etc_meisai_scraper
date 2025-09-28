# Additional Tasks for gRPC Migration

These tasks address the critical gaps identified in the analysis and must be added to tasks.md before implementation can begin.

## Phase 3.0: Pre-Migration Tasks (NEW - Add before Phase 3.1)

### Performance Baseline (addresses NFR-001)
- [ ] T-001 Capture current performance baseline using go test -bench on existing implementation
- [ ] T-002 Document baseline metrics in tests/performance/baseline.json
- [ ] T-003 Create performance comparison script to validate ±10% requirement

### Test Cleanup (addresses Constitution Principle 1)
- [ ] T-004 Clean up any existing test files in src/ directory if violations found
- [ ] T-005 Move any misplaced mock files from src/ to tests/mocks/

## Phase 3.3.5: GORM Hooks Migration (NEW - Add after Phase 3.3)

### Hook Extraction (addresses FR-014)
- [ ] T-031A Identify and document all GORM hooks in existing models (BeforeSave, AfterCreate, etc.)
- [ ] T-031B Create src/services/hooks_migrator.go to centralize business logic from hooks
- [ ] T-031C Extract validation logic from GORM hooks to src/services/validation_service.go
- [ ] T-031D Extract audit logging from GORM hooks to src/services/audit_service.go
- [ ] T-031E Write tests for extracted hook logic in tests/unit/services/hooks_migrator_test.go
- [ ] T-031F Update adapter layer to call migrated hook logic at appropriate points

## Phase 3.5.5: Rollback Procedures (NEW - Add after Phase 3.5)

### Rollback Implementation (addresses NFR-004)
- [ ] T-063A Create rollback.md documenting step-by-step rollback procedures
- [ ] T-063B Create scripts/rollback.sh for automated rollback execution
- [ ] T-063C Tag current commit as 'pre-migration-baseline' for easy rollback
- [ ] T-063D Create rollback verification checklist
- [ ] T-063E Test rollback procedure in development environment
- [ ] T-063F Document rollback recovery time objective (RTO)

## Updates to Existing Tasks

### Phase 3.4: Core Implementation
Update T032-T035 to include specific handling for:
- Database column name mapping configuration
- Custom validation logic migration points
- Backward compatibility flags during transition

### Phase 3.6: Polish
Update T072 to reference the baseline captured in T-001:
- T072: Validate response times are within ±10% of baseline captured in T-001

## Task Renumbering

After adding these tasks, renumber all tasks sequentially:
- Phase 3.0: T001-T005 (Pre-Migration)
- Phase 3.1: T006-T011 (Setup)
- Phase 3.2: T012-T022 (Proto Definitions)
- Phase 3.3: T023-T035 (Tests First)
- Phase 3.3.5: T036-T041 (GORM Hooks Migration)
- Phase 3.4: T042-T062 (Core Implementation)
- Phase 3.5: T063-T070 (Integration)
- Phase 3.5.5: T071-T076 (Rollback Procedures)
- Phase 3.6: T077-T091 (Polish)

## Updated Task Count
- Original: 74 tasks
- Additional: 17 tasks
- **New Total: 91 tasks**

## Critical Path Changes
The critical path now includes:
1. Performance baseline capture (must complete before any migration)
2. GORM hooks migration (blocks service layer implementation)
3. Rollback procedures (must be ready before production deployment)

## Success Criteria Updates
Add to success criteria:
- [ ] Performance baseline captured and documented
- [ ] All GORM hooks successfully migrated
- [ ] Rollback procedure tested and verified
- [ ] No test files remain in src/ directory
- [ ] All 91 tasks completed

---
*These additional tasks address all HIGH severity findings from analysis.md*
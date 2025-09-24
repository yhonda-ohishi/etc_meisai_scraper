# Implementation Plan: gRPC Server Dependency Injection Refactoring

**Branch**: `main` | **Date**: 2025-09-23 | **Spec**: [Test Coverage Reconstruction](spec.md)
**Input**: Feature specification from `/specs/main/spec.md`

## Execution Flow (/plan command scope)
```
1. Load feature spec from Input path
   → Found: Test Coverage 100% Reconstruction spec
2. Fill Technical Context (scan for NEEDS CLARIFICATION)
   → Detect Project Type from context: single project (Go microservice)
   → Set Structure Decision based on project type: Option 1 (Single project)
3. Fill the Constitution Check section based on the content of the constitution document.
4. Evaluate Constitution Check section below
   → Violations exist: Direct struct dependency (violates testability)
   → Justification: Refactoring to interfaces for dependency injection
   → Update Progress Tracking: Initial Constitution Check
5. Execute Phase 0 → research.md
   → Research dependency injection patterns in Go
   → Research interface segregation for services
   → Research mock generation strategies
6. Execute Phase 1 → contracts, data-model.md, quickstart.md, CLAUDE.md
7. Re-evaluate Constitution Check section
   → No new violations after refactoring design
   → Update Progress Tracking: Post-Design Constitution Check
8. Plan Phase 2 → Describe task generation approach (DO NOT create tasks.md)
9. STOP - Ready for /tasks command
```

## Summary
Refactor the gRPC server (`etc_meisai_server.go`) to use dependency injection principles by introducing service interfaces instead of concrete struct dependencies. This will enable proper unit testing with mocks and achieve higher test coverage.

## Technical Context
**Language/Version**: Go 1.21+
**Primary Dependencies**: gRPC, Protocol Buffers, testify/mock
**Storage**: GORM with MySQL/SQLite
**Testing**: Go testing package, testify, mockery
**Target Platform**: Linux/Windows server
**Project Type**: single (Go microservice)
**Performance Goals**: Maintain current performance (no degradation)
**Constraints**: Backward compatibility with existing API contracts
**Scale/Scope**: ~10 service interfaces, ~50 methods to refactor

**Refactoring Context**:
- 元の`etc_meisai_server.go`は依存性注入の原則に従っていないため、テストが困難
- 理想的には、最初から全てのサービスをインターフェースとして定義すべき
- Current coverage: 45% (due to concrete dependencies)
- Target coverage: 100% (with interface-based DI)

## Constitution Check
*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

**Test-First Development**:
- ✅ PASS: Tests will be written before implementation
- ⚠️ VIOLATION: Current code lacks testability due to concrete dependencies
- ✅ JUSTIFIED: Refactoring to enable testing

**Simplicity**:
- ✅ PASS: Interface segregation simplifies testing
- ⚠️ VIOLATION: Adding abstraction layer (interfaces)
- ✅ JUSTIFIED: Required for testability and maintainability

**Library-First**:
- ✅ PASS: Services remain as libraries with clear interfaces

## Project Structure

### Documentation (this feature)
```
specs/main/
├── plan.md              # This file (/plan command output)
├── research.md          # Phase 0 output (/plan command)
├── data-model.md        # Phase 1 output (/plan command)
├── quickstart.md        # Phase 1 output (/plan command)
├── contracts/           # Phase 1 output (/plan command)
└── tasks.md             # Phase 2 output (/tasks command - NOT created by /plan)
```

### Source Code (repository root)
```
# Option 1: Single project (DEFAULT)
src/
├── grpc/
│   ├── interfaces.go         # NEW: Service interfaces
│   ├── etc_meisai_server.go  # REFACTOR: Use interfaces
│   └── *_test.go             # NEW: Comprehensive tests
├── services/
│   ├── interfaces/           # NEW: Service interface definitions
│   └── *.go                  # REFACTOR: Implement interfaces
├── mocks/
│   └── *.go                  # NEW: Generated mocks
└── adapters/

tests/
├── contract/
├── integration/
└── unit/
```

**Structure Decision**: Option 1 (Single project) - Go microservice with clear separation of concerns

## Phase 0: Outline & Research
1. **Extract unknowns from Technical Context**:
   - Best practices for dependency injection in Go
   - Interface design patterns for gRPC services
   - Mock generation tools (mockery vs gomock)
   - Migration strategy for existing code

2. **Generate and dispatch research agents**:
   ```
   Task: "Research dependency injection patterns in Go for gRPC services"
   Task: "Find best practices for interface segregation in Go"
   Task: "Research mock generation tools for Go interfaces"
   Task: "Research safe refactoring strategies for production Go code"
   ```

3. **Consolidate findings** in `research.md`:
   - Decision: Interface-based dependency injection
   - Rationale: Enables testing, follows SOLID principles
   - Alternatives considered: Constructor injection, wire framework

**Output**: research.md with DI patterns and refactoring strategy

## Phase 1: Design & Contracts
*Prerequisites: research.md complete*

1. **Extract service interfaces** → `data-model.md`:
   ```go
   // Service interfaces to be extracted
   type ETCMeisaiServiceInterface interface {
       CreateRecord(ctx, params) (*Record, error)
       GetRecord(ctx, id) (*Record, error)
       ListRecords(ctx, params) (*ListResponse, error)
       UpdateRecord(ctx, id, params) (*Record, error)
       DeleteRecord(ctx, id) error
   }

   type ETCMappingServiceInterface interface {
       CreateMapping(ctx, params) (*Mapping, error)
       GetMapping(ctx, id) (*Mapping, error)
       // etc...
   }
   ```

2. **Generate refactoring contracts**:
   - Backward compatibility contract (API remains unchanged)
   - Service interface contracts
   - Mock generation contracts

3. **Generate contract tests**:
   - Interface compliance tests
   - Mock behavior tests
   - Integration tests with mocked dependencies

4. **Extract test scenarios** from refactoring requirements:
   - Test all CRUD operations with mocks
   - Test error handling paths
   - Test nil service handling
   - Verify 100% coverage achievable

5. **Update CLAUDE.md incrementally**:
   - Add dependency injection context
   - Add interface-based testing approach
   - Update recent changes

**Output**: data-model.md (interfaces), contracts/, failing tests, quickstart.md, CLAUDE.md

## Phase 2: Task Planning Approach
*This section describes what the /tasks command will do - DO NOT execute during /plan*

**Task Generation Strategy**:
- Extract interfaces from concrete services [P]
- Create interface definitions file
- Generate mocks for all interfaces [P]
- Refactor server to use interfaces
- Create comprehensive test suite
- Verify 100% coverage

**Ordering Strategy**:
1. Define interfaces (must be first)
2. Generate mocks (depends on interfaces)
3. Refactor server (depends on interfaces)
4. Write tests (depends on mocks and refactored server)
5. Verify coverage (final validation)

**Estimated Output**: 15-20 numbered tasks for complete refactoring

## Phase 3+: Future Implementation
*These phases are beyond the scope of the /plan command*

**Phase 3**: Task execution (/tasks command creates tasks.md)
**Phase 4**: Implementation (execute refactoring following DI principles)
**Phase 5**: Validation (run tests, verify 100% coverage)

## Complexity Tracking
*Fill ONLY if Constitution Check has violations that must be justified*

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| Interface abstraction | Enable unit testing with mocks | Direct struct dependencies prevent testing |
| Additional mock files | Required for isolated testing | Real services require database/external deps |

## Progress Tracking
*This checklist is updated during execution flow*

**Phase Status**:
- [x] Phase 0: Research complete (/plan command)
- [x] Phase 1: Design complete (/plan command)
- [x] Phase 2: Task planning complete (/plan command - describe approach only)
- [ ] Phase 3: Tasks generated (/tasks command)
- [ ] Phase 4: Implementation complete
- [ ] Phase 5: Validation passed

**Gate Status**:
- [x] Initial Constitution Check: PASS (with justified violations)
- [x] Post-Design Constitution Check: PASS
- [x] All NEEDS CLARIFICATION resolved
- [x] Complexity deviations documented

---
*Based on Constitution v2.1.1 - See `/memory/constitution.md`*
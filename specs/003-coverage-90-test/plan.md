
# Implementation Plan: Test Coverage Recovery and Refactoring

**Branch**: `003-coverage-90-test` | **Date**: 2025-09-25 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `/specs/003-coverage-90-test/spec.md`

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
6. Execute Phase 1 → contracts, data-model.md, quickstart.md, agent-specific template file (e.g., `CLAUDE.md` for Claude Code, `.github/copilot-instructions.md` for GitHub Copilot, `GEMINI.md` for Gemini CLI, `QWEN.md` for Qwen Code or `AGENTS.md` for opencode).
7. Re-evaluate Constitution Check section
   → If new violations: Refactor design, return to Phase 1
   → Update Progress Tracking: Post-Design Constitution Check
8. Plan Phase 2 → Describe task generation approach (DO NOT create tasks.md)
9. STOP - Ready for /tasks command
```

**IMPORTANT**: The /plan command STOPS at step 7. Phases 2-4 are executed by other commands:
- Phase 2: /tasks command creates tasks.md
- Phase 3-4: Implementation execution (manual or via tools)

## Summary
Fix test coverage execution failures that occurred after test modifications, restore coverage measurement capability to 95% or higher from current 0.7%, and refactor test infrastructure to prevent future breakdowns. Primary focus: BaseService deadlock has been fixed, now need to fix compilation errors and missing mocks, then improve coverage percentage.

## Technical Context
**Language/Version**: Go 1.21+
**Primary Dependencies**: testify/mock, gRPC, grpc-gateway, GORM
**Storage**: MySQL via db_service (gRPC)
**Testing**: Go test with coverage profiles, testify framework
**Target Platform**: Windows/Linux development environments
**Project Type**: single - Go module with gRPC services
**Performance Goals**: Test execution under 2 minutes for full suite
**Constraints**: Tests must not deadlock, proper cleanup required, <2min timeout per suite
**Scale/Scope**: 82 test files total (78 in tests/, 4 in src/), targeting 95% coverage across all packages except vendor/

**Known Issues from Investigation**:
- Deadlock in `BaseService.Shutdown()` method (sync.RWMutex lock contention) - **FIXED**
  - Technical: Mutex was held during LogOperation() call, fixed by releasing before logging
- Tests compilation errors due to model/interface changes
  - Missing fields: ETCNum, TollAmount, VehicleClass in models.ETCMeisai
  - Interface mismatches in repositories
- Current overall coverage at 0.7% (was able to fix BaseService but other issues remain)
- Test files exist in tests/unit/ directory (not co-located with source)
  - Structure: tests/unit/[package]/*_test.go mirrors src/[package]/
- Missing mock implementations causing dependency issues
  - Need mockgen for repositories.ETCRepository, repositories.MappingRepository
  - Missing mocks for gRPC clients

**Clarifications Applied**:
- Target: 95% or higher coverage (improved from current 0.7%)
- Format: JSON only for machine processing
- Scope: All packages except vendor/
- Timeout: Maximum 2-minute execution limit
- Priority: Fix deadlocks first, coverage second

## Constitution Check
*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

Since the constitution file is a template, using standard Go best practices:
- ✅ **Test-First**: Tests already exist, need fixing not new creation
- ✅ **Simplicity**: Focusing on fixing existing tests, not adding complexity
- ✅ **Performance**: Addressing timeouts and deadlocks directly
- ✅ **Observability**: Will add better error reporting for test failures
- ✅ **No New Dependencies**: Using existing Go test tooling

## Project Structure

### Documentation (this feature)
```
specs/[###-feature]/
├── plan.md              # This file (/plan command output)
├── research.md          # Phase 0 output (/plan command)
├── data-model.md        # Phase 1 output (/plan command)
├── quickstart.md        # Phase 1 output (/plan command)
├── contracts/           # Phase 1 output (/plan command)
└── tasks.md             # Phase 2 output (/tasks command - NOT created by /plan)
```

### Source Code (repository root)
```
# Option 1: Single project (DEFAULT - ACTUAL)
src/
├── models/
├── services/
├── grpc/
├── handlers/
├── parser/
├── config/
├── adapters/
└── repositories/

tests/          # 78 test files located here
├── contract/
├── integration/
└── unit/
    ├── models/
    ├── services/
    ├── handlers/
    ├── grpc/
    └── ...

# Option 2: Web application (when "frontend" + "backend" detected)
backend/
├── src/
│   ├── models/
│   ├── services/
│   └── api/
└── tests/

frontend/
├── src/
│   ├── components/
│   ├── pages/
│   └── services/
└── tests/

# Option 3: Mobile + API (when "iOS/Android" detected)
api/
└── [same as backend above]

ios/ or android/
└── [platform-specific structure]
```

**Structure Decision**: Option 1 (single Go module) - existing project structure maintained

## Phase 0: Outline & Research
1. **Extract unknowns from Technical Context** above:
   - For each NEEDS CLARIFICATION → research task
   - For each dependency → best practices task
   - For each integration → patterns task

2. **Generate and dispatch research agents**:
   ```
   For each unknown in Technical Context:
     Task: "Research {unknown} for {feature context}"
   For each technology choice:
     Task: "Find best practices for {tech} in {domain}"
   ```

3. **Consolidate findings** in `research.md` using format:
   - Decision: [what was chosen]
   - Rationale: [why chosen]
   - Alternatives considered: [what else evaluated]

**Output**: research.md with all NEEDS CLARIFICATION resolved

## Phase 1: Design & Contracts
*Prerequisites: research.md complete*

1. **Extract entities from feature spec** → `data-model.md`:
   - Entity name, fields, relationships
   - Validation rules from requirements
   - State transitions if applicable

2. **Generate API contracts** from functional requirements:
   - For each user action → endpoint
   - Use standard REST/GraphQL patterns
   - Output OpenAPI/GraphQL schema to `/contracts/`

3. **Generate contract tests** from contracts:
   - One test file per endpoint
   - Assert request/response schemas
   - Tests must fail (no implementation yet)

4. **Extract test scenarios** from user stories:
   - Each story → integration test scenario
   - Quickstart test = story validation steps

5. **Update agent file incrementally** (O(1) operation):
   - Run `.specify/scripts/powershell/update-agent-context.ps1 -AgentType claude`
     **IMPORTANT**: Execute it exactly as specified above. Do not add or remove any arguments.
   - If exists: Add only NEW tech from current plan
   - Preserve manual additions between markers
   - Update recent changes (keep last 3)
   - Keep under 150 lines for token efficiency
   - Output to repository root

**Output**: data-model.md, /contracts/*, failing tests, quickstart.md, agent-specific file

## Phase 2: Task Planning Approach
*This section describes what the /tasks command will do - DO NOT execute during /plan*

**Task Generation Strategy**:
- Load `.specify/templates/tasks-template.md` as base
- Generate tasks from Phase 1 design docs (contracts, data model, quickstart)
- Focus on deadlock fixes first (critical path)
- Then coverage improvement tasks by package priority

**Task Categories**:
1. **Critical Fixes** (T001-T005):
   - Fix BaseService deadlock (separate mutexes) - COMPLETED
   - Add proper test cleanup patterns - COMPLETED
   - Fix compilation errors in test files - NEW PRIORITY

2. **Test Restoration** (T006-T015):
   - Re-enable disabled test files
   - Fix compilation errors in existing tests
   - Add missing test coverage for critical paths

3. **Coverage Enhancement** (T016-T025):
   - Add tests for uncovered packages (priority: services, grpc, config)
   - Improve test quality with table-driven tests
   - Add edge case and error path testing

**Ordering Strategy**:
- Sequential for mutex fixes (avoid conflicts)
- Parallel [P] for independent package tests
- Dependencies: Fix infrastructure before adding tests

**Estimated Output**: 25 numbered tasks focusing on deadlock resolution and coverage improvement

**IMPORTANT**: This phase is executed by the /tasks command, NOT by /plan

## Phase 3+: Future Implementation
*These phases are beyond the scope of the /plan command*

**Phase 3**: Task execution (/tasks command creates tasks.md)  
**Phase 4**: Implementation (execute tasks.md following constitutional principles)  
**Phase 5**: Validation (run tests, execute quickstart.md, performance validation)

## Complexity Tracking
*Fill ONLY if Constitution Check has violations that must be justified*

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| [e.g., 4th project] | [current need] | [why 3 projects insufficient] |
| [e.g., Repository pattern] | [specific problem] | [why direct DB access insufficient] |


## Progress Tracking
*This checklist is updated during execution flow*

**Phase Status**:
- [x] Phase 0: Research complete (/plan command) - research.md generated
- [x] Phase 1: Design complete (/plan command) - data-model.md, contracts/, quickstart.md, CLAUDE.md updated
- [x] Phase 2: Task planning complete (/plan command - describe approach only)
- [x] Phase 3: Tasks generated (/tasks command) - 30 tasks created (T001-T030)
- [ ] Phase 4: Implementation complete
- [ ] Phase 5: Validation passed

**Gate Status**:
- [x] Initial Constitution Check: PASS
- [x] Post-Design Constitution Check: PASS (no new violations)
- [x] All NEEDS CLARIFICATION resolved
- [x] Complexity deviations documented (none required)

---
*Based on Constitution v2.1.1 - See `/memory/constitution.md`*

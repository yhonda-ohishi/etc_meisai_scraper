
# Implementation Plan: Aligned Test Coverage Reconstruction

**Branch**: `002-aligned-test-coverage` | **Date**: 2025-09-23 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `/specs/002-aligned-test-coverage/spec.md`

## Execution Flow (/plan command scope)
```
1. Load feature spec from Input path
   → Found: Aligned Test Coverage Reconstruction spec
2. Fill Technical Context (scan for NEEDS CLARIFICATION)
   → Detected 2 clarifications: test execution time, package scope
   → Resolved: 60 seconds total, src/ packages only
   → Detect Project Type: single (Go testing infrastructure)
   → Set Structure Decision: Option 1 (Single project)
3. Fill the Constitution Check section
   → Constitution is template only - proceeding with basic principles
4. Evaluate Constitution Check section
   → No violations with assumed test-first approach
   → Update Progress Tracking: Initial Constitution Check
5. Execute Phase 0 → research.md
   → Research test patterns, mocking strategies, coverage tools
6. Execute Phase 1 → contracts, data-model.md, quickstart.md, CLAUDE.md
   → Define test infrastructure, mock patterns, coverage targets
7. Re-evaluate Constitution Check section
   → No new violations detected
   → Update Progress Tracking: Post-Design Constitution Check
8. Plan Phase 2 → Task generation for test reconstruction
9. STOP - Ready for /tasks command
```

**IMPORTANT**: The /plan command STOPS at step 7. Phases 2-4 are executed by other commands:
- Phase 2: /tasks command creates tasks.md
- Phase 3-4: Implementation execution (manual or via tools)

## Summary
Complete reconstruction of the test suite by removing all existing test files and systematically recreating them from scratch to achieve 100% statement coverage. The approach ensures all tests are independent, deterministic, and execute without external dependencies using comprehensive mocking.

## Technical Context
**Language/Version**: Go 1.21+
**Primary Dependencies**: testify/assert, testify/mock, testify/require
**Storage**: N/A (tests use mocks exclusively)
**Testing**: Go testing package with coverage tooling
**Target Platform**: Linux/Windows/Mac (cross-platform)
**Project Type**: single (Go microservice testing infrastructure)
**Performance Goals**: Test suite execution < 60 seconds total
**Constraints**: No external dependencies, 100% coverage requirement, deterministic execution
**Scale/Scope**: ~50 packages, ~200 source files, target 100% coverage

**Clarifications Resolved** (from spec.md):
- Test execution time limit: 60 seconds for entire suite
- Package scope: src/ directory packages only (business logic focus)
- Generated files (*.pb.go) excluded from coverage requirements

## Constitution Check
*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

**Assessment**: Constitution file is template-only, proceeding with standard test-first principles:
- **Test-First Approach**: Aligned with feature goal of systematic test reconstruction
- **Clean Architecture**: Tests separated from source code, using mocks for dependencies
- **Deterministic Execution**: No external dependencies, consistent results required
- **Performance Standards**: 60-second execution limit enforced

**Status**: ✅ PASS - No constitutional violations detected

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
# Option 1: Single project (DEFAULT)
src/
├── models/
├── services/
├── cli/
└── lib/

tests/
├── contract/
├── integration/
└── unit/

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

**Structure Decision**: [DEFAULT to Option 1 unless Technical Context indicates web/mobile app]

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
- Generate systematic test reconstruction tasks
- Remove existing tests from src/ (Phase 0)
- Create test infrastructure in tests/unit/ (Phase 1)
- Generate test files for each src/ package (Phase 2)
- Achieve 100% coverage validation (Phase 3)

**Ordering Strategy**:
- Infrastructure first: Mock registry, helpers, fixtures
- Core packages first: models, then services, then repositories
- Supporting packages: handlers, adapters, middleware, interceptors
- Coverage validation and performance optimization last

**Task Categories**:
1. **Infrastructure Tasks**: Test directory structure, mocks, helpers
2. **Core Package Tests**: models/, services/, repositories/
3. **Supporting Package Tests**: handlers/, adapters/, grpc/, middleware/, interceptors/, parser/
4. **Coverage Validation**: 100% coverage verification, performance checks

**Estimated Output**: 20-25 numbered, ordered tasks in tasks.md

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
- [x] Phase 0: Research complete (/plan command)
- [x] Phase 1: Design complete (/plan command)
- [x] Phase 2: Task planning complete (/plan command - describe approach only)
- [x] Phase 3: Tasks generated (/tasks command)
- [ ] Phase 4: Implementation complete
- [ ] Phase 5: Validation passed

**Gate Status**:
- [x] Initial Constitution Check: PASS
- [x] Post-Design Constitution Check: PASS
- [x] All NEEDS CLARIFICATION resolved
- [x] Complexity deviations documented (none required)

---
*Based on Constitution v2.1.1 - See `/memory/constitution.md`*

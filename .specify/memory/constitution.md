<!--
Sync Impact Report:
Version: 0.0.0 → 1.0.0 (Initial ratification with test separation principle)
New Principles Added:
- Test File Separation: Tests must be in tests/ directory, never in src/
- Test-First Development: TDD approach mandatory
- Coverage Requirements: 100% coverage target per project
- Clean Architecture: Clear separation of concerns
- Observable Systems: Comprehensive logging and monitoring

Templates Requiring Updates:
✅ Constitution created
⚠ plan-template.md - May need update for test directory structure
⚠ spec-template.md - May need update for test requirements
⚠ tasks-template.md - May need update for test task locations

Follow-up TODOs:
- Confirm exact ratification date with project team
-->

# ETC明細システム Constitution

## Core Principles

### I. Test File Separation (NON-NEGOTIABLE)
Test files MUST be placed in the `tests/` directory structure, NEVER in the `src/` directory.
Tests are organized as `tests/unit/`, `tests/integration/`, and `tests/contract/`.
This ensures clean separation between production code and test code, prevents accidental
inclusion of test dependencies in production builds, and maintains clear project structure.

### II. Test-First Development
TDD (Test-Driven Development) approach is mandatory for all new features.
Tests must be written first, fail initially, then implementation follows.
Red-Green-Refactor cycle must be strictly enforced.
This ensures comprehensive test coverage and drives better design decisions.

### III. Coverage Requirements
Project MUST maintain 100% test coverage target as per project requirements.
Coverage is measured at statement level using Go's built-in coverage tools.
Any code that cannot be tested must be explicitly documented with justification.
Coverage reports must be generated for all CI/CD pipelines.

### IV. Clean Architecture
Clear separation of concerns MUST be maintained across all layers:
- `src/models/` - Data models and entities only
- `src/services/` - Business logic layer
- `src/repositories/` - Data access layer
- `src/grpc/` - gRPC server and handlers
- `src/adapters/` - External system adapters
No cross-layer dependencies allowed except through defined interfaces.

### V. Observable Systems
All services MUST implement comprehensive logging and monitoring.
Structured logging is required for all operations.
Performance metrics must be tracked for critical operations.
Timeout protection (2-minute max) for all long-running operations.

### VI. Dependency Injection
All services MUST use constructor injection pattern.
Dependencies are passed as interfaces, not concrete types.
No global state or singleton patterns allowed.
This enables proper mocking and testing.

### VII. Simplicity First
Start with the simplest solution that works.
YAGNI (You Aren't Gonna Need It) principles apply.
Complexity must be explicitly justified and documented.
Prefer Go standard library over external dependencies when possible.

## Development Standards

### Code Organization
- Production code in `src/` directory only
- Test code in `tests/` directory only
- Mock files in `src/mocks/` for generated mocks
- Documentation in `docs/` directory
- Scripts and tools in `scripts/` directory

### Testing Requirements
- Unit tests for all business logic
- Integration tests for service interactions
- Contract tests for all API endpoints
- Table-driven tests preferred for comprehensive coverage
- Parallel test execution where possible

### Quality Gates
- All tests must pass before merge
- Coverage must meet or exceed 100% target
- No build warnings allowed
- Code must pass linting checks
- Documentation must be updated for API changes

## Governance

The Constitution supersedes all other project practices and guidelines.
Any amendments require:
1. Documentation of proposed change with rationale
2. Team review and approval
3. Migration plan for existing code if needed
4. Update to all affected documentation

All pull requests must verify constitutional compliance.
Violations must be fixed before merge is allowed.
Complexity additions require explicit justification in PR description.

Runtime development guidance is maintained in CLAUDE.md for agent-specific instructions.

**Version**: 1.0.0 | **Ratified**: 2025-09-26 | **Last Amended**: 2025-09-26
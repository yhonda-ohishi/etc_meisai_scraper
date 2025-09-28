<!--
Sync Impact Report:
Version: 1.0.0 → 1.1.0 (Added Code Quality Validation principle for immediate go vet error correction)
Modified Principles:
- Added Principle VIII: Code Quality Validation

Templates Requiring Updates:
✅ Constitution updated with new principle
⚠ CLAUDE.md - May need update to include go vet instructions
⚠ plan-template.md - May need update for quality validation tasks
⚠ tasks-template.md - May need update for validation checkpoints

Follow-up TODOs:
- Update CLAUDE.md with specific go vet handling instructions
- Consider adding pre-commit hooks for go vet validation
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

### VIII. Code Quality Validation
All code modifications MUST pass `go vet` without any errors or warnings.
When hooks report `go vet` errors, they MUST be corrected immediately before proceeding.
After every edit operation, `go vet` validation MUST be performed to ensure code quality.
This prevents accumulation of technical debt and maintains consistent code standards.
Common `go vet` issues include unreachable code, malformed build tags, incorrect printf formats,
and suspicious constructs that must be fixed immediately upon detection.

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
- `go vet` must report zero issues
- Documentation must be updated for API changes

### Validation Workflow
1. Make code changes
2. Run `go vet` immediately after edits
3. Fix any reported issues before continuing
4. Run tests to verify functionality
5. Check coverage metrics
6. Commit only clean, validated code

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
`go vet` errors block merge until resolved.

Runtime development guidance is maintained in CLAUDE.md for agent-specific instructions.

**Version**: 1.1.0 | **Ratified**: 2025-09-26 | **Last Amended**: 2025-09-27

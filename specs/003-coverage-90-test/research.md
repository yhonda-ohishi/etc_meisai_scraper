# Research: Test Coverage Recovery and Refactoring

## Root Cause Analysis

### Decision: BaseService Deadlock Identified
**Rationale**: Analysis of test execution failures revealed sync.RWMutex contention in BaseService.Shutdown() method. When tests call logging operations during shutdown, a deadlock occurs because:
- Shutdown() holds write lock
- LogOperation() tries to acquire read lock
- Classic reader-writer lock ordering problem

**Alternatives considered**:
- Goroutine leaks: Ruled out - test cleanup functions present
- Memory exhaustion: Ruled out - tests fail immediately, not after memory buildup
- Circular dependencies: Ruled out - import cycles would fail at compile time

## Test Execution Strategy

### Decision: Fix Deadlock Before Coverage
**Rationale**: Cannot measure coverage if tests don't execute. Deadlock fix is prerequisite for all coverage improvements.

**Alternatives considered**:
- Parallel fix approach: Too risky, could mask root issues
- Skip failing tests: Would give false coverage metrics
- Rewrite test framework: Too time-consuming, existing tests salvageable

## Coverage Measurement Approach

### Decision: Use Native Go Coverage Tools
**Rationale**: Go's built-in coverage profiling is mature, well-integrated, and produces machine-readable formats.

**Alternatives considered**:
- Third-party coverage tools: Add complexity without clear benefits
- Custom coverage instrumentation: Reinventing the wheel
- External coverage services: Network dependency unnecessary for local development

## Test Timeout Management

### Decision: 2-Minute Hard Timeout Per Suite
**Rationale**: Balances catching hanging tests while allowing legitimate long-running tests to complete.

**Alternatives considered**:
- 30-second timeout: Too aggressive, some integration tests need more time
- 5-minute timeout: Too lenient, developers won't wait
- No timeout: Risk of infinite hangs in CI/CD

## Coverage Target Strategy

### Decision: Incremental Coverage Improvement
**Rationale**: Jump from current broken state to 95% is achievable in phases:
1. Fix execution (restore measurability)
2. Fix failing tests (baseline coverage)
3. Add missing tests (reach 95%)

**Alternatives considered**:
- Big bang approach: Too risky, might break more than fix
- Lower target (80%): Doesn't meet business requirements
- Package-by-package: Good idea, will incorporate into approach

## Refactoring Approach

### Decision: Minimal Refactoring During Fix
**Rationale**: Focus on fixing deadlock and restoring coverage. Major refactoring can follow once tests are stable.

**Alternatives considered**:
- Complete rewrite: Too time-consuming, loses existing test value
- No refactoring: Won't prevent future issues
- Extensive refactoring: Risk of introducing new bugs while fixing old ones

## JSON Report Format

### Decision: Standard Go Coverage JSON
**Rationale**: Native format, widely supported by tools, includes all necessary metrics.

**Alternatives considered**:
- Custom JSON schema: Unnecessary complexity
- Multiple formats: Maintenance burden
- XML/HTML: Not requested, JSON sufficient for machine processing

## Package Scope Definition

### Decision: All Packages Except vendor/
**Rationale**: vendor/ contains third-party code, not our responsibility to test.

**Alternatives considered**:
- Include vendor/: Wasteful, not our code
- Exclude generated code: Could hide issues in code generation
- Cherry-pick packages: Too complex, might miss issues

## Mutex Fix Strategy

### Decision: Separate Logging Mutex from State Mutex
**Rationale**: Eliminates lock ordering issues by using independent locks for different concerns.

**Alternatives considered**:
- Try-lock pattern: Complex, Go doesn't have native trylock
- Lock-free logging: Over-engineering for this use case
- Remove logging from shutdown: Would reduce observability

## Test Infrastructure Improvements

### Decision: Add t.Cleanup() Patterns
**Rationale**: Ensures consistent cleanup even when tests fail, prevents resource leaks.

**Alternatives considered**:
- defer statements: Don't run on test failure
- Manual cleanup: Error-prone, often forgotten
- Test fixtures: Over-complex for this issue

---

## Technical Recommendations Summary

1. **Immediate**: Fix BaseService mutex ordering
2. **Short-term**: Add proper test cleanup patterns
3. **Medium-term**: Implement incremental coverage improvements
4. **Long-term**: Establish coverage gates in CI/CD

All decisions prioritize stability and measurability over perfect design, following the principle of fixing the critical path first.
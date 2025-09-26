# Research: Test Coverage Visualization and Dependency Injection

## Coverage Tool Architecture

### Decision: Go Native Tools with Custom Visualization
**Rationale**: Go's built-in `go test -cover` and `go tool cover` provide robust coverage data. Custom visualization layer can format this data for console display with progress bars and colors.

**Alternatives considered**:
- Third-party coverage tools (gocov, gocover): Add unnecessary dependencies
- Web-based dashboards: Violates console-only requirement
- IDE integrations: Not portable across development environments

## Test/Source Separation Strategy

### Decision: Use -coverpkg Flag with Path Mapping
**Rationale**: The `-coverpkg=./src/...` flag allows tests in `tests/unit/` to measure coverage of `src/` packages. This maintains directory separation while ensuring accurate measurement.

**Alternatives considered**:
- Moving tests to src/: Breaks existing project structure
- Symlinks: Platform-dependent and complex
- Build tags: Overcomplicated for this use case

## Interface Unification Approach

### Decision: Central Interface Registry Pattern
**Rationale**: Create a single source of truth for all interfaces in `src/interfaces/` directory. All implementations and mocks reference these canonical interfaces.

**Alternatives considered**:
- Interface per package: Led to current fragmentation
- Code generation only: Doesn't solve root cause of inconsistency
- Duck typing: Go's type system prevents this

## Dependency Injection Framework

### Decision: Constructor Injection with Interface Parameters
**Rationale**: Simple, explicit, testable. Each service accepts interfaces in constructor, making mock injection straightforward.

**Alternatives considered**:
- Wire (Google's DI): Overkill for this project
- Dig (Uber's DI): Adds complexity
- Singleton pattern: Makes testing harder

## Mock Generation Strategy

### Decision: Hybrid Approach - testify/mock + mockgen
**Rationale**: Use testify/mock for simple cases (better ergonomics), mockgen for complex interfaces (automatic generation).

**Alternatives considered**:
- Only testify/mock: Manual work for complex interfaces
- Only mockgen: Less flexible for simple cases
- GoMock: Deprecated in favor of mockgen

## Visual Output Format

### Decision: ANSI Color Codes with Unicode Progress Bars
**Rationale**: Works across Windows Terminal, Linux terminals. Provides clear visual feedback without external dependencies.

**Alternatives considered**:
- ASCII-only: Less visually appealing
- Terminal UI libraries: Adds dependencies
- Plain text: Doesn't meet visual requirement

## Coverage Calculation Algorithm

### Decision: Package-Level Aggregation with Category Grouping
**Rationale**: Calculate coverage per package, then group by category (models, services, etc.) for organized display.

**Alternatives considered**:
- File-by-file: Too granular for overview
- Single percentage: Not enough detail
- Function-level: Excessive detail for console

## Performance Optimization

### Decision: Parallel Test Execution with Bounded Concurrency
**Rationale**: Use `go test -parallel` with CPU count limit to speed up coverage calculation while avoiding resource exhaustion.

**Alternatives considered**:
- Sequential execution: Too slow for large codebase
- Unlimited parallelism: Can overwhelm system
- Caching previous results: Complexity vs benefit trade-off

## Recommendation Engine

### Decision: Rule-Based Suggestions
**Rationale**: Simple rules like "Add table-driven tests for functions with >3 parameters" provide actionable guidance without AI complexity.

**Alternatives considered**:
- ML-based suggestions: Overkill
- Generic messages: Not actionable
- No recommendations: Misses improvement opportunity

## Error Handling Strategy

### Decision: Graceful Degradation with Warnings
**Rationale**: If some packages fail to analyze, show partial results with warnings rather than complete failure.

**Alternatives considered**:
- Fail fast: Too rigid
- Silent failure: Hides problems
- Retry mechanism: Adds complexity

---

## Technical Recommendations Summary

1. **Immediate**: Create `scripts/coverage-report.go` as main tool
2. **Short-term**: Unify interfaces in `src/interfaces/` directory
3. **Medium-term**: Refactor services to use constructor injection
4. **Long-term**: Achieve and maintain 90% coverage threshold

All decisions prioritize simplicity, maintainability, and meeting the 2-minute performance requirement while preserving the existing test/source separation.
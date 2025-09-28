# Data Model: Test Coverage Infrastructure

**Generated**: 2025-09-21
**Context**: Test structures and coverage models for ETC明細 system

## Test Infrastructure Entities

### Test Coverage Model

```go
type CoverageTarget struct {
    Package           string      // Package identifier (e.g., "src/models")
    TargetCoverage   float64     // Target coverage percentage (100.0)
    CurrentCoverage  float64     // Current measured coverage
    StatementCount   int         // Total statements in package
    CoveredCount     int         // Covered statements
    Status           CoverageStatus
    LastMeasured     time.Time
}

type CoverageStatus string
const (
    StatusPending    CoverageStatus = "pending"
    StatusInProgress CoverageStatus = "in_progress"
    StatusAchieved   CoverageStatus = "achieved"
    StatusFailed     CoverageStatus = "failed"
)
```

### Test Suite Model

```go
type TestSuite struct {
    Package        string           // Package being tested
    TestFiles      []TestFile       // List of test files
    MockFiles      []MockFile       // Associated mock files
    FixtureFiles   []FixtureFile    // Test data files
    Dependencies   []string         // Package dependencies
    ExecutionTime  time.Duration    // Total execution time
    Status         TestStatus
}

type TestFile struct {
    FilePath       string           // Path to test file
    TestFunctions  []TestFunction   // Individual test functions
    Coverage       float64          // File-level coverage
    ExecutionTime  time.Duration    // File execution time
}

type TestFunction struct {
    Name           string           // Function name
    Type           TestType         // Unit, Integration, Contract, etc.
    Scenarios      []TestScenario   // Test scenarios within function
    Dependencies   []MockDependency // Required mocks
}

type TestType string
const (
    TypeUnit        TestType = "unit"
    TypeIntegration TestType = "integration"
    TypeContract    TestType = "contract"
    TypeBenchmark   TestType = "benchmark"
)
```

### Mock Infrastructure Model

```go
type MockRegistry struct {
    Interfaces     []MockInterface  // Available mock interfaces
    Implementations []MockImpl      // Mock implementations
    TestFactories  []TestFactory   // Data factories
}

type MockInterface struct {
    Name           string           // Interface name
    Package        string           // Source package
    Methods        []MockMethod     // Interface methods
    Dependencies   []string         // Interface dependencies
}

type MockMethod struct {
    Name           string           // Method name
    Parameters     []Parameter      // Method parameters
    Returns        []ReturnValue    // Return values
    Behaviors      []MockBehavior   // Configured behaviors
}

type MockBehavior struct {
    Scenario       string           // Behavior scenario name
    Input          interface{}      // Input conditions
    Output         interface{}      // Expected output
    Error          error           // Error to return
    CallCount      int             // Expected call count
}
```

### Test Execution Model

```go
type TestExecution struct {
    ID             string           // Execution identifier
    Timestamp      time.Time        // Execution start time
    PackageResults []PackageResult  // Results per package
    OverallResult  ExecutionResult  // Overall execution result
    Duration       time.Duration    // Total execution time
    CoverageReport CoverageReport   // Coverage analysis
}

type PackageResult struct {
    Package        string           // Package name
    TestsRun       int              // Number of tests executed
    TestsPassed    int              // Number of tests passed
    TestsFailed    int              // Number of tests failed
    Coverage       float64          // Package coverage percentage
    Duration       time.Duration    // Package execution time
    Failures       []TestFailure    // Failed test details
}

type TestFailure struct {
    TestName       string           // Failed test name
    Error          string           // Error message
    StackTrace     string           // Failure stack trace
    File           string           // Test file
    Line           int              // Line number
}
```

### Coverage Reporting Model

```go
type CoverageReport struct {
    Timestamp      time.Time        // Report generation time
    OverallCoverage float64         // Total coverage percentage
    PackageCoverage []PackageCoverage // Per-package coverage
    Gaps           []CoverageGap    // Uncovered code segments
    Trends         []CoverageTrend  // Historical coverage data
}

type PackageCoverage struct {
    Package        string           // Package identifier
    Coverage       float64          // Coverage percentage
    Statements     int              // Total statements
    Covered        int              // Covered statements
    Files          []FileCoverage   // Per-file coverage
}

type FileCoverage struct {
    FilePath       string           // File path
    Coverage       float64          // File coverage percentage
    Lines          []LineCoverage   // Line-by-line coverage
}

type LineCoverage struct {
    LineNumber     int              // Line number
    IsCovered      bool             // Whether line is covered
    HitCount       int              // Number of times hit
}

type CoverageGap struct {
    Package        string           // Package with gap
    File           string           // File with uncovered code
    StartLine      int              // Gap start line
    EndLine        int              // Gap end line
    Reason         string           // Why uncovered
    Priority       GapPriority      // Fix priority
}

type GapPriority string
const (
    PriorityHigh     GapPriority = "high"     // Critical path uncovered
    PriorityMedium   GapPriority = "medium"   // Important functionality
    PriorityLow      GapPriority = "low"      // Edge cases or error paths
)
```

## Test Data Fixtures Model

```go
type TestFixture struct {
    Name           string           // Fixture name
    Type           FixtureType      // Type of test data
    Data           interface{}      // Actual test data
    Dependencies   []string         // Required fixtures
    Cleanup        func()           // Cleanup function
}

type FixtureType string
const (
    FixtureETCMeisai    FixtureType = "etc_meisai"
    FixtureETCMapping   FixtureType = "etc_mapping"
    FixtureImportSession FixtureType = "import_session"
    FixtureCSVData      FixtureType = "csv_data"
    FixtureAPIResponse  FixtureType = "api_response"
)

type TestFactory struct {
    Name           string           // Factory name
    BuildMethods   []BuildMethod    // Available build methods
    Customizers    []Customizer     // Data customization options
}

type BuildMethod struct {
    Name           string           // Method name (e.g., "BuildETCMeisai")
    ReturnType     string           // Type returned
    Parameters     []Parameter      // Optional parameters
    DefaultValues  map[string]interface{} // Default field values
}
```

## Validation Rules

### Test Coverage Validation

```go
type CoverageValidator struct {
    MinimumCoverage    float64      // Minimum acceptable coverage (100.0)
    RequiredPackages   []string     // Packages that must have tests
    ExcludedFiles      []string     // Files excluded from coverage
    PerformanceLimits  PerformanceLimits
}

type PerformanceLimits struct {
    MaxTotalDuration   time.Duration // Max total test execution time
    MaxPackageDuration time.Duration // Max per-package execution time
    MaxMemoryUsage     int64         // Max memory usage in bytes
}
```

### Test Quality Validation

```go
type TestQualityRules struct {
    RequireTableDriven    bool         // Require table-driven tests
    RequireMockIsolation  bool         // Require dependency mocking
    RequireParallelSafe   bool         // Tests must be parallel-safe
    MaxTestComplexity     int          // Maximum cyclomatic complexity
    RequiredAssertions    []string     // Required assertion patterns
}
```

## State Transitions

### Coverage Progress States

```
pending → in_progress → achieved
    ↓         ↓            ↓
  failed ← failed ←  validation_failed
```

### Test Execution States

```
queued → running → completed
   ↓       ↓         ↓
 failed ← failed ← failed
```

## Relationships

- **Package** (1) → (N) **TestFile**: Each package has multiple test files
- **TestFile** (1) → (N) **TestFunction**: Each test file contains multiple test functions
- **TestFunction** (1) → (N) **MockDependency**: Each test may require multiple mocks
- **MockInterface** (1) → (N) **MockBehavior**: Each mock can have multiple behaviors
- **TestExecution** (1) → (N) **PackageResult**: Each execution covers multiple packages
- **CoverageReport** (1) → (N) **CoverageGap**: Each report identifies multiple gaps

## Implementation Constraints

1. **Performance**: Total test suite execution must be under 30 seconds
2. **Isolation**: No external dependencies allowed in tests
3. **Determinism**: Tests must be reproducible and non-flaky
4. **Coverage**: Must achieve 100% statement coverage across all packages
5. **Maintainability**: Test code must follow established patterns and be self-documenting

---
*Data model complete - Ready for contract generation*

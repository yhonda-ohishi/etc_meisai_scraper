# Data Model: Test Coverage Recovery

## Core Entities

### TestSuite
Represents a collection of test files for a specific package.

**Fields**:
- `package_name` (string, required): Full package path (e.g., "github.com/project/src/services")
- `test_files` ([]string): List of test file paths in the suite
- `status` (enum): PENDING | RUNNING | PASSED | FAILED | TIMEOUT | DEADLOCK
- `start_time` (timestamp): When test execution started
- `end_time` (timestamp, nullable): When test execution completed
- `duration_ms` (int): Execution time in milliseconds
- `error_message` (string, nullable): Error details if failed
- `coverage_profile` (CoverageProfile, nullable): Coverage data if successful

**Validation Rules**:
- package_name must be valid Go package path
- duration_ms must not exceed 120000 (2-minute timeout)
- status transitions: PENDING → RUNNING → (PASSED|FAILED|TIMEOUT|DEADLOCK)

### CoverageProfile
Contains coverage metrics for a package or file.

**Fields**:
- `id` (string): Unique identifier (package_name + timestamp)
- `package_name` (string, required): Package being measured
- `total_statements` (int): Total executable statements
- `covered_statements` (int): Statements with test coverage
- `coverage_percent` (float): Percentage covered (0-100)
- `file_coverage` ([]FileCoverage): Per-file breakdown
- `generated_at` (timestamp): When profile was created
- `format` (enum): JSON | HTML | TEXT

**Validation Rules**:
- covered_statements <= total_statements
- coverage_percent = (covered_statements / total_statements) * 100
- coverage_percent must be between 0 and 100

### FileCoverage
Coverage data for a single source file.

**Fields**:
- `file_path` (string, required): Relative path from package root
- `total_lines` (int): Total lines in file
- `covered_lines` (int): Lines with coverage
- `uncovered_lines` ([]int): Line numbers without coverage
- `coverage_percent` (float): File-level coverage

**Validation Rules**:
- covered_lines <= total_lines
- All uncovered_lines must be valid line numbers (1 to total_lines)

### TestExecution
Tracks a complete test run across all packages.

**Fields**:
- `execution_id` (uuid): Unique execution identifier
- `started_at` (timestamp): Execution start time
- `completed_at` (timestamp, nullable): Execution completion time
- `total_packages` (int): Number of packages to test
- `passed_packages` (int): Packages that passed
- `failed_packages` (int): Packages that failed
- `timeout_packages` (int): Packages that timed out
- `overall_coverage` (float): Aggregate coverage percentage
- `target_coverage` (float): Target coverage (95.0)
- `meets_threshold` (boolean): Whether target was met
- `test_suites` ([]TestSuite): All package test results

**Validation Rules**:
- passed_packages + failed_packages + timeout_packages <= total_packages
- overall_coverage must be weighted average of package coverages
- meets_threshold = (overall_coverage >= target_coverage)

### TestResource
Represents system resources during test execution (for leak detection).

**Fields**:
- `test_suite_id` (string): Associated test suite
- `goroutines_start` (int): Goroutine count at test start
- `goroutines_end` (int): Goroutine count at test end
- `memory_start_mb` (int): Memory usage at start
- `memory_end_mb` (int): Memory usage at end
- `open_files_start` (int): Open file descriptors at start
- `open_files_end` (int): Open file descriptors at end
- `has_leak` (boolean): Whether resource leak detected

**Validation Rules**:
- has_leak = true if goroutines_end > goroutines_start + threshold
- has_leak = true if memory_end_mb > memory_start_mb * 1.5
- has_leak = true if open_files_end > open_files_start + 10

### TestConfiguration
Configuration for test execution behavior.

**Fields**:
- `timeout_seconds` (int): Maximum time per test suite (default: 120)
- `parallel_execution` (boolean): Run packages in parallel
- `max_parallel` (int): Maximum parallel packages (default: 4)
- `coverage_threshold` (float): Required coverage (95.0)
- `exclude_patterns` ([]string): Patterns to exclude (e.g., "vendor/*")
- `output_format` (enum): JSON | HTML | BOTH
- `fail_on_timeout` (boolean): Treat timeout as failure (default: true)
- `verbose` (boolean): Detailed output logging

**Validation Rules**:
- timeout_seconds must be between 10 and 600
- max_parallel must be between 1 and 16
- coverage_threshold must be between 0 and 100
- exclude_patterns must be valid glob patterns

## Entity Relationships

```
TestExecution (1) ──contains──> (N) TestSuite
TestSuite (1) ──generates──> (0..1) CoverageProfile
CoverageProfile (1) ──contains──> (N) FileCoverage
TestSuite (1) ──monitors──> (1) TestResource
TestExecution (1) ──uses──> (1) TestConfiguration
```

## State Transitions

### TestSuite Status Flow
```
[PENDING] ──start──> [RUNNING] ──success──> [PASSED]
                        │
                        ├──failure──> [FAILED]
                        ├──timeout──> [TIMEOUT]
                        └──deadlock──> [DEADLOCK]
```

### TestExecution Flow
```
[INITIALIZED] ──begin──> [EXECUTING] ──complete──> [COMPLETED]
                            │
                            └──abort──> [ABORTED]
```

## Coverage Calculation Rules

1. **Package Coverage**: (covered_statements / total_statements) * 100
2. **Overall Coverage**: Weighted average based on package sizes
3. **Exclusions**: Files matching exclude_patterns not counted
4. **Threshold Check**: Overall coverage must meet or exceed target

## Deadlock Detection Rules

A test suite is marked as DEADLOCK when:
1. No output for 30 seconds during execution
2. Goroutine count increases continuously without decrease
3. Mutex contention detected in stack traces
4. Test process becomes unresponsive to signals
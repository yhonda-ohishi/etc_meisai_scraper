# Test Optimizer Tool

Analyzes and optimizes Go test execution time to meet performance targets.

## Features

- Test performance profiling
- Parallel execution optimization
- Slow test identification
- Test dependency analysis
- Optimization recommendations
- Quick analysis mode (without test execution)

## Installation

```bash
# Build the tool
cd scripts/test-optimizer
go build -o test-optimizer.exe .

# Or run directly
go run scripts/test-optimizer/main.go <args>
```

## Usage

```bash
# Full analysis (runs tests and profiles performance)
test-optimizer <project-root> [output-dir]

# Quick analysis (analyzes structure without running tests)
test-optimizer --quick <project-root> [output-dir]

# Examples
test-optimizer .                    # Analyze current directory
test-optimizer . reports            # Output to reports directory
test-optimizer --quick .            # Quick analysis without running tests
```

### Arguments

1. `project-root`: Root directory of the Go project to analyze
2. `output-dir`: (Optional) Directory for output files (default: "coverage")

### Options

- `--quick`: Perform quick analysis without actually running tests

## Output Files

The tool generates several files in the output directory:

1. **test-optimization.json**: Detailed analysis in JSON format
2. **test-optimization.txt**: Human-readable report
3. **parallel-test.sh**: Shell script for optimized parallel test execution

## Report Contents

### Performance Metrics
- Total test execution time
- Package-level timing breakdown
- Individual test timings
- Slow test identification (tests > 1 second)

### Optimization Recommendations
- Parallel execution opportunities
- Test grouping strategies
- Resource optimization suggestions
- Caching recommendations

## Example Output

```
TEST PERFORMANCE OPTIMIZATION REPORT
============================================================

Current Total Duration: 45s
Target Duration: 30s
Status: ❌ EXCEEDS TARGET

TOP 10 SLOWEST TESTS:
------------------------------------------------------------
1. TestDatabaseIntegration: 12.5s
2. TestFileSystemOperations: 8.3s
3. TestAPIEndToEnd: 5.7s
...

OPTIMIZATION RECOMMENDATIONS:
------------------------------------------------------------
• Enable parallel execution for 15 packages
• Optimize 8 slow tests (>1s each)
• Use test data factories to reduce setup time
• Implement test result caching for expensive operations
• Consider using build tags to separate unit and integration tests

PARALLEL GROUPS:
------------------------------------------------------------
Group 1 (Can run in parallel):
  - github.com/project/pkg1
  - github.com/project/pkg2
  ...

Group 2 (Database tests - sequential):
  - github.com/project/db
  ...
```

## Performance Targets

Default target: **30 seconds** for entire test suite

The tool will:
- Exit with code 0 if tests meet the target
- Exit with code 1 if tests exceed the target

## Parallel Test Script

The generated `parallel-test.sh` script groups tests for optimal parallel execution:

```bash
# Run the generated script
./scripts/parallel-test.sh

# Or on Windows
bash scripts/parallel-test.sh
```

## Quick Mode

Quick mode (`--quick`) is useful for:
- CI/CD pipeline checks
- Pre-commit hooks
- Fast feedback during development
- Analyzing project structure without test execution

## Integration with CI/CD

```yaml
# GitHub Actions example
- name: Analyze test performance
  run: |
    go run scripts/test-optimizer/main.go --quick .

- name: Full test optimization check
  run: |
    go run scripts/test-optimizer/main.go .
    if [ $? -ne 0 ]; then
      echo "Tests exceed 30 second target"
      exit 1
    fi
```

## Best Practices

1. **Run regularly**: Include in CI/CD to track test performance over time
2. **Monitor trends**: Watch for gradual performance degradation
3. **Set realistic targets**: Adjust the 30-second target based on your project
4. **Use quick mode**: For fast feedback during development
5. **Review slow tests**: Focus optimization efforts on the slowest tests first
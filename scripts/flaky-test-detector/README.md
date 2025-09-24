# Flaky Test Detector

Identifies and helps eliminate flaky tests by running them multiple times and analyzing the results.

## Features

- Multiple test execution analysis
- Flaky test pattern detection
- Failure reason categorization
- Severity classification
- Fix recommendations
- Quick mode for fast checks
- Detailed reporting in multiple formats

## Installation

```bash
# Build the tool
cd scripts/flaky-test-detector
go build -o flaky-test-detector.exe .

# Or run directly
go run scripts/flaky-test-detector/main.go <args>
```

## Usage

```bash
# Standard analysis (10 runs)
flaky-test-detector <project-root>

# Custom number of runs
flaky-test-detector <project-root> 20

# Quick mode (3 runs)
flaky-test-detector --quick <project-root>

# Custom output directory
flaky-test-detector <project-root> 10 reports

# Examples
flaky-test-detector .                        # Analyze current directory
flaky-test-detector --quick .                # Quick 3-run check
flaky-test-detector . 20 flaky-reports      # 20 runs, custom output
```

### Arguments

1. `project-root`: Root directory of the Go project to analyze
2. `runs`: (Optional) Number of test runs (default: 10)
3. `output-dir`: (Optional) Directory for output files (default: flaky-test-report)

### Options

- `--quick`: Run only 3 iterations for quick flakiness check
- `--parallel`: (Future) Run tests in parallel

## Output Files

The tool generates several reports in the output directory:

### flaky-test-report.json
Complete analysis in JSON format:
- All test results across runs
- Flaky test details
- Failure patterns
- Recommendations

### flaky-test-report.txt
Human-readable summary:
- Overall statistics
- List of flaky tests
- Severity classifications
- Action items

### flaky-test-report.html
Interactive HTML report:
- Visual pass/fail patterns
- Sortable test lists
- Detailed failure logs
- Trend analysis

### fix-flaky-tests.sh
Shell script with suggested fixes:
- Test isolation commands
- Retry logic examples
- Parallel execution adjustments

## Flakiness Detection

A test is considered flaky if it:
- Passes sometimes but not always
- Shows inconsistent behavior across runs
- Has non-deterministic failures

## Flakiness Categories

### Timing Issues
- Race conditions
- Insufficient wait times
- Timeout problems

### Resource Contention
- Database conflicts
- File system races
- Network port conflicts

### Environmental Dependencies
- External service availability
- System resource availability
- Configuration variations

### Test Isolation
- Shared state between tests
- Improper cleanup
- Order dependencies

## Severity Levels

- **Critical**: Pass rate < 50%
- **High**: Pass rate 50-70%
- **Medium**: Pass rate 70-90%
- **Low**: Pass rate > 90%

## Example Output

```
FLAKY TEST DETECTION REPORT
============================================================

Analysis Summary:
  Total Runs: 10
  Tests Analyzed: 145
  Flaky Tests Found: 8 (5.52%)

Critical Flaky Tests:
  1. TestDatabaseConcurrency (30% pass rate)
     Category: Resource Contention
     Failures: timeout waiting for database lock

  2. TestAPIRateLimit (45% pass rate)
     Category: Timing Issues
     Failures: intermittent rate limit exceeded

Recommendations:
  • Add proper synchronization to TestDatabaseConcurrency
  • Increase timeout values for TestAPIRateLimit
  • Implement test isolation for shared resources
  • Use test fixtures instead of shared database
```

## Fixing Flaky Tests

### Common Solutions

1. **Add Synchronization**
   ```go
   // Use sync.WaitGroup or channels
   var wg sync.WaitGroup
   wg.Add(1)
   go func() {
       defer wg.Done()
       // Test code
   }()
   wg.Wait()
   ```

2. **Increase Timeouts**
   ```go
   ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
   defer cancel()
   ```

3. **Isolate Test Resources**
   ```go
   // Use unique database/tables per test
   dbName := fmt.Sprintf("test_%s_%d", t.Name(), time.Now().Unix())
   ```

4. **Add Retry Logic**
   ```go
   // For inherently flaky operations
   for i := 0; i < 3; i++ {
       if err = operation(); err == nil {
           break
       }
       time.Sleep(time.Second * time.Duration(i+1))
   }
   ```

## CI/CD Integration

### GitHub Actions
```yaml
- name: Detect flaky tests
  run: |
    go run scripts/flaky-test-detector/main.go --quick .
  continue-on-error: true

- name: Upload flaky test report
  uses: actions/upload-artifact@v2
  with:
    name: flaky-test-report
    path: flaky-test-report/
```

### GitLab CI
```yaml
flaky-tests:
  script:
    - go run scripts/flaky-test-detector/main.go . 5
  artifacts:
    when: always
    paths:
      - flaky-test-report/
    reports:
      junit: flaky-test-report/junit.xml
```

## Best Practices

1. **Regular Monitoring**: Run flaky test detection weekly
2. **Quick Checks**: Use `--quick` mode in pre-commit hooks
3. **Fix Immediately**: Address critical flaky tests first
4. **Track Trends**: Monitor flakiness rate over time
5. **Test in CI**: Run detection in CI environment for consistency

## Performance

- Quick mode: ~3x test suite execution time
- Standard mode: ~10x test suite execution time
- Memory usage: Minimal (stores only test results)

## Exit Codes

- `0`: No flaky tests detected
- `1`: Flaky tests found (or error occurred)

Use the exit code in CI/CD to fail builds with flaky tests.
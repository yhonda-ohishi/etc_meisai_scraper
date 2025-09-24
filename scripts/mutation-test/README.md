# Mutation Testing Tool

Performs mutation testing to evaluate the effectiveness of your test suite by introducing controlled mutations to code and verifying tests catch them.

## Features

- Multiple mutation operators (conditional, arithmetic, logical, return, constant)
- Automatic test execution for each mutation
- Mutation score calculation
- Quick mode for CI/CD integration
- Parallel execution support
- Detailed reporting in multiple formats
- Surviving mutant identification

## Installation

```bash
# Build the tool
cd scripts/mutation-test
go build -o mutation-test.exe .

# Or run directly
go run scripts/mutation-test/main.go <args>
```

## Usage

```bash
# Standard mutation testing
mutation-test <source-dir>

# Custom output directory
mutation-test <source-dir> <output-dir>

# Quick mode (limited mutations)
mutation-test --quick <source-dir>

# Parallel execution
mutation-test --parallel <source-dir>

# Examples
mutation-test ./src                          # Test src directory
mutation-test --quick ./src                  # Quick test with subset
mutation-test --parallel ./src reports       # Parallel with custom output
```

### Arguments

1. `source-dir`: Directory containing Go source files to mutate
2. `output-dir`: (Optional) Directory for output files (default: mutation-report)

### Options

- `--quick`: Test only a subset of mutations (5 per file)
- `--parallel`: Run mutation tests in parallel

## Mutation Operators

### Conditional Mutations
Changes conditional operators:
- `==` → `!=`
- `!=` → `==`
- `<` → `<=`
- `>` → `>=`
- `<=` → `<`
- `>=` → `>`

### Arithmetic Mutations
Changes arithmetic operators:
- `+` → `-`
- `-` → `+`
- `*` → `/`
- `/` → `*`
- `++` → `--`
- `--` → `++`

### Logical Mutations
Changes logical operators:
- `&&` → `||`
- `||` → `&&`
- Removes `!` negation

### Return Mutations
Changes return values:
- `true` → `false`
- `false` → `true`

### Constant Mutations
Changes constant values:
- `0` → `1`
- `1` → `0`
- `""` → `"mutated"`

## Output Files

### mutation-report.json
Complete analysis in JSON format:
- All mutation results
- File-by-file breakdown
- Surviving mutant details
- Mutation score

### mutation-report.txt
Human-readable text report:
- Overall statistics
- Mutation score
- List of surviving mutants
- Recommendations

### mutation-report.html
Interactive HTML report:
- Visual mutation score
- Color-coded results
- Detailed survivor information
- File-by-file breakdown

## Mutation Score

The mutation score indicates test suite effectiveness:
```
Mutation Score = (Killed Mutants / Total Mutants) × 100%
```

### Score Interpretation
- **Excellent** (90-100%): Very effective test suite
- **Good** (80-90%): Good test coverage
- **Adequate** (60-80%): Room for improvement
- **Poor** (<60%): Significant test gaps

## Example Output

```
Starting mutation testing...
Found 42 files to mutate
Testing 156 mutations in src/services/import.go
  [1/156] Testing Conditional mutation at line 45... ✅ Killed
  [2/156] Testing Arithmetic mutation at line 67... ❌ Survived
  [3/156] Testing Return mutation at line 89... ✅ Killed

Mutation Testing Complete!
Mutation Score: 85.7%
Reports generated in mutation-report

SURVIVING MUTANTS (Tests didn't catch these):
src/services/import.go:
  Line 67: Replace + with -
    Original: count + 1
    Mutated:  count - 1
```

## CI/CD Integration

### GitHub Actions
```yaml
- name: Mutation Testing
  run: |
    go run scripts/mutation-test/main.go --quick ./src
  continue-on-error: true

- name: Upload mutation reports
  uses: actions/upload-artifact@v2
  if: always()
  with:
    name: mutation-reports
    path: mutation-report/
```

### GitLab CI
```yaml
mutation-test:
  script:
    - go run scripts/mutation-test/main.go --quick ./src
  artifacts:
    when: always
    paths:
      - mutation-report/
    reports:
      junit: mutation-report/junit.xml
```

### Jenkins
```groovy
stage('Mutation Testing') {
    steps {
        sh 'go run scripts/mutation-test/main.go --quick ./src'
    }
    post {
        always {
            archiveArtifacts artifacts: 'mutation-report/**/*'
        }
    }
}
```

## Improving Mutation Score

### 1. Add Boundary Tests
Test edge cases and boundary conditions:
```go
func TestBoundaryConditions(t *testing.T) {
    assert.Equal(t, 0, calculate(-1))  // Test negative
    assert.Equal(t, 1, calculate(0))   // Test zero
    assert.Equal(t, 2, calculate(1))   // Test positive
}
```

### 2. Test All Branches
Ensure all conditional branches are tested:
```go
func TestAllBranches(t *testing.T) {
    // Test if branch
    result := process(true, 10)
    assert.Equal(t, expected1, result)

    // Test else branch
    result = process(false, 10)
    assert.Equal(t, expected2, result)
}
```

### 3. Test Error Paths
Don't forget error handling:
```go
func TestErrorHandling(t *testing.T) {
    _, err := processInvalid(nil)
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "invalid input")
}
```

### 4. Test Arithmetic Operations
Verify calculations are correct:
```go
func TestArithmetic(t *testing.T) {
    assert.Equal(t, 15, add(10, 5))
    assert.Equal(t, 5, subtract(10, 5))
    assert.Equal(t, 50, multiply(10, 5))
    assert.Equal(t, 2, divide(10, 5))
}
```

## Best Practices

1. **Regular Testing**: Run mutation testing weekly or before releases
2. **Quick Mode in CI**: Use `--quick` for pull request validation
3. **Full Tests Nightly**: Run complete mutation testing nightly
4. **Target Score**: Aim for 80%+ mutation score
5. **Focus on Critical Code**: Prioritize business logic and algorithms
6. **Ignore Generated Code**: Exclude protobuf, mocks, etc.

## Limitations

- Only mutates Go source files
- Requires compilable code
- Test execution time increases with mutation count
- Some mutations may be equivalent (produce same behavior)

## Performance

- Quick mode: ~5 mutations per file
- Standard mode: All possible mutations
- Execution time: O(mutations × test_suite_time)
- Memory usage: Minimal (only stores results)

## Exit Codes

- `0`: Mutation score >= 80%
- `1`: Mutation score < 80% or error occurred

## Troubleshooting

### Tests Take Too Long
- Use `--quick` mode
- Limit scope with specific directories
- Enable `--parallel` execution

### Low Mutation Score
- Review surviving mutants in report
- Add tests for uncaught mutations
- Focus on boundary conditions

### Build Errors During Mutation
- Ensure code compiles before testing
- Check for syntax-aware mutations
- Review mutation locations in report
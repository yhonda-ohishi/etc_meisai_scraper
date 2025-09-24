# Scripts Directory

This directory contains standalone Go scripts for various development and testing tasks. Each script is a separate program and should be run independently.

## Important Note

Each script has its own `main` function and should be run separately. Do not attempt to build all scripts together as a single package.

## Available Scripts

### 1. coverage-advanced/
Advanced coverage analysis with branch coverage measurement.
- **Location**: `scripts/coverage-advanced/main.go`
- **Documentation**: [coverage-advanced/README.md](coverage-advanced/README.md)

**Usage:**
```bash
go run scripts/coverage-advanced/main.go <coverage-profile> <source-dir>

# Or build and run:
cd scripts/coverage-advanced
go build -o coverage-advanced.exe .
./coverage-advanced.exe <coverage-profile> <source-dir>
```

### 2. coverage-report/
Generates comprehensive coverage reports in multiple formats.
- **Location**: `scripts/coverage-report/main.go`
- **Documentation**: [coverage-report/README.md](coverage-report/README.md)

**Usage:**
```bash
go run scripts/coverage-report/main.go <coverage-file> <source-dir> <output-dir>

# Or build and run:
cd scripts/coverage-report
go build -o coverage-report.exe .
./coverage-report.exe coverage.out . reports
```

### 3. flaky-test-detector/
Detects flaky tests by running them multiple times and analyzing results.
- **Location**: `scripts/flaky-test-detector/main.go`
- **Documentation**: [flaky-test-detector/README.md](flaky-test-detector/README.md)

**Usage:**
```bash
go run scripts/flaky-test-detector/main.go <project-root> [runs] [output-dir]
go run scripts/flaky-test-detector/main.go --quick <project-root>

# Or build and run:
cd scripts/flaky-test-detector
go build -o flaky-test-detector.exe .
./flaky-test-detector.exe <project-root>
```

### 4. mutation-test/
Performs mutation testing to evaluate test suite effectiveness.
- **Location**: `scripts/mutation-test/main.go`
- **Documentation**: [mutation-test/README.md](mutation-test/README.md)

**Usage:**
```bash
go run scripts/mutation-test/main.go <source-dir> [output-dir]
go run scripts/mutation-test/main.go --quick <source-dir>
go run scripts/mutation-test/main.go --parallel <source-dir>

# Or build and run:
cd scripts/mutation-test
go build -o mutation-test.exe .
./mutation-test.exe <source-dir> [output-dir]
```

### 5. test-optimizer/
Optimizes test execution by analyzing dependencies and parallelization.
- **Location**: `scripts/test-optimizer/main.go`
- **Documentation**: [test-optimizer/README.md](test-optimizer/README.md)

**Usage:**
```bash
go run scripts/test-optimizer/main.go <project-root> [output-dir]
go run scripts/test-optimizer/main.go --quick <project-root> [output-dir]

# Or build and run:
cd scripts/test-optimizer
go build -o test-optimizer.exe .
./test-optimizer.exe <project-root> [output-dir]
```

## Building Scripts

If you want to build a specific script as an executable:

```bash
# Build a specific script
go build -o coverage-advanced.exe scripts/coverage-advanced.go

# Or from within the scripts directory
cd scripts
go build -o coverage-advanced.exe coverage-advanced.go
```

## IDE Configuration

If your IDE shows errors about duplicate main functions:

1. **VS Code**: Configure each script as a separate module or exclude the scripts folder from the main build
2. **GoLand**: Mark each script as a separate run configuration
3. **Command Line**: Always run scripts individually using `go run`

## Note on Package Declaration

All scripts use `package main` as they are standalone executables, not library code. This is intentional and correct for their use case.
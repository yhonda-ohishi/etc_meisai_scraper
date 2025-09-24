# Data Model: Aligned Test Coverage Reconstruction

**Feature**: 002-aligned-test-coverage
**Date**: 2025-09-23

## Core Entities

### Test Suite
**Purpose**: Collection of test files for a specific package, must achieve 100% coverage

**Attributes**:
- Package: string (e.g., "models", "services")
- Location: string (e.g., "tests/unit/models/")
- CoverageTarget: float64 (100.0)
- TestFiles: []TestFile

### Mock Object
**Purpose**: Simulated implementation of external dependency

**Attributes**:
- Interface: string (Go interface name)
- Methods: []MockMethod
- Expectations: []MockExpectation

### Coverage Report
**Purpose**: Metrics showing coverage per package and aggregate

**Attributes**:
- Package: string
- CoveragePercentage: float64 (must be 100.0)
- UncoveredLines: []LineReference

## Validation Rules

### Coverage Requirements
- 100% statement coverage for all src/ packages
- Exclusion only for generated files (*.pb.go)
- All error paths must be tested

### Performance Constraints
- Total test suite execution < 60 seconds
- Tests must be deterministic and independent

## Test Organization

### File Structure
```
tests/
├── unit/          # Unit tests for src/ packages
│   ├── models/
│   ├── services/
│   └── repositories/
├── fixtures/      # Shared test data
├── helpers/       # Test utilities
└── mocks/         # Mock implementations
```

### Coverage Validation
```bash
# Generate coverage for src/ packages
go test -coverprofile=coverage.out ./src/...

# Validate 100% coverage
go tool cover -func=coverage.out | grep total:
```
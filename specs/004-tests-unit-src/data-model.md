# Data Model: Coverage Visualization System

## Core Entities

### CoverageReport
Represents a complete coverage analysis run.

**Fields**:
- `id` (string): Unique identifier (timestamp-based)
- `timestamp` (time): When the analysis was performed
- `total_packages` (int): Number of packages analyzed
- `overall_coverage` (float): Overall coverage percentage (0-100)
- `meets_threshold` (bool): Whether 90% threshold is met
- `execution_time` (duration): How long the analysis took
- `packages` ([]PackageCoverage): Per-package coverage data
- `categories` ([]CategoryCoverage): Coverage grouped by category

**Validation Rules**:
- overall_coverage must be between 0 and 100
- execution_time must not exceed 2 minutes
- timestamp must not be in the future

### PackageCoverage
Coverage data for a single package.

**Fields**:
- `name` (string): Full package path (e.g., "github.com/project/src/services")
- `short_name` (string): Display name (e.g., "services")
- `category` (string): Package category (models/services/handlers/repositories/utils)
- `coverage_percent` (float): Package coverage percentage (0-100)
- `total_statements` (int): Total executable statements
- `covered_statements` (int): Number of covered statements
- `files` ([]FileCoverage): Per-file coverage data
- `status` (string): visual status (excellent/good/warning/critical)

**Validation Rules**:
- covered_statements <= total_statements
- coverage_percent = (covered_statements / total_statements) * 100
- status derived from coverage_percent (>=90=excellent, >=75=good, >=50=warning, <50=critical)

### FileCoverage
Coverage data for a single source file.

**Fields**:
- `path` (string): Relative file path
- `coverage_percent` (float): File coverage percentage
- `total_lines` (int): Total lines in file
- `covered_lines` (int): Lines with coverage
- `uncovered_lines` ([]int): Line numbers without coverage
- `uncovered_functions` ([]string): Names of uncovered functions

**Validation Rules**:
- covered_lines <= total_lines
- All uncovered_lines must be valid line numbers

### CategoryCoverage
Aggregated coverage for a package category.

**Fields**:
- `name` (string): Category name (Core/models, Services, API/handlers, Repositories, Utils)
- `package_count` (int): Number of packages in category
- `average_coverage` (float): Average coverage across packages
- `min_coverage` (float): Lowest package coverage
- `max_coverage` (float): Highest package coverage
- `packages_below_threshold` ([]string): Packages under 90%

**Validation Rules**:
- min_coverage <= average_coverage <= max_coverage
- All values between 0 and 100

### InterfaceContract
Unified interface definition for mocking.

**Fields**:
- `name` (string): Interface name
- `package` (string): Package containing interface
- `methods` ([]MethodSignature): Interface methods
- `implementations` ([]string): Known implementations
- `mock_generated` (bool): Whether mock exists
- `last_updated` (time): Last modification time

**Validation Rules**:
- name must follow Go naming conventions
- package must be valid Go package path
- methods must have valid Go signatures

### MethodSignature
Method definition within an interface.

**Fields**:
- `name` (string): Method name
- `parameters` ([]Parameter): Input parameters
- `returns` ([]Type): Return types
- `receiver` (string): Receiver type if applicable

### CoverageRecommendation
Actionable suggestion for improving coverage.

**Fields**:
- `package` (string): Target package
- `priority` (string): high/medium/low
- `type` (string): test_missing/mock_needed/refactor_suggested
- `message` (string): Human-readable recommendation
- `impact` (float): Estimated coverage improvement (0-100)

**Validation Rules**:
- priority must be valid enum value
- impact between 0 and 100
- message must be non-empty

### DependencyInjection
Tracks DI configuration for testability.

**Fields**:
- `service` (string): Service name
- `interface` (string): Interface it implements
- `dependencies` ([]string): Required interface dependencies
- `constructor` (string): Constructor function name
- `mock_compatible` (bool): Whether mocks can be injected

## Entity Relationships

```
CoverageReport (1) ──contains──> (N) PackageCoverage
PackageCoverage (1) ──contains──> (N) FileCoverage
CoverageReport (1) ──contains──> (N) CategoryCoverage
CategoryCoverage (1) ──references──> (N) PackageCoverage
InterfaceContract (1) ──has──> (N) MethodSignature
InterfaceContract (1) ──implemented by──> (N) DependencyInjection
CoverageReport (1) ──generates──> (N) CoverageRecommendation
```

## State Transitions

### Coverage Analysis States
```
[INITIALIZED] ──start──> [RUNNING] ──complete──> [COMPLETED]
                  │
                  └──timeout──> [TIMEOUT]
                  └──error──> [FAILED]
```

### Package Coverage Status
```
[CRITICAL] <50% ──improve──> [WARNING] 50-75% ──improve──> [GOOD] 75-90% ──improve──> [EXCELLENT] ≥90%
```

## Display Formatting

### Visual Indicators
- Progress bar: `[████████████████░░░░] 80%`
- Status icons: ✅ (excellent), 🟡 (good), 🟠 (warning), 🔴 (critical)
- Category headers with emoji: 🎯 Core, ⚙️ Services, 🌐 API, 📦 Repositories, 🔧 Utils
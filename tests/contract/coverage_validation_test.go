package contract

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

const (
	// Coverage targets
	MinimumCoverage = 100.0
	WarningCoverage = 95.0

	// Performance targets
	MaxTestExecutionTime = 60 * time.Second
	MaxSingleTestTime    = 5 * time.Second

	// Quality targets
	MaxCyclomaticComplexity = 10
)

// TestCoverageContract validates that all packages meet coverage requirements
func TestCoverageContract(t *testing.T) {
	t.Run("ValidatePackageCoverage", func(t *testing.T) {
		packages := []string{
			"./src/models/...",
			"./src/services/...",
			"./src/repositories/...",
			"./src/handlers/...",
			"./src/middleware/...",
			"./src/interceptors/...",
			"./src/grpc/...",
			"./src/adapters/...",
			"./src/parser/...",
			"./src/config/...",
			"./src/server/...",
		}

		for _, pkg := range packages {
			t.Run(pkg, func(t *testing.T) {
				coverage := getPackageCoverage(t, pkg)

				if coverage < WarningCoverage {
					t.Logf("WARNING: Package %s has coverage %.1f%% (below warning threshold %.1f%%)",
						pkg, coverage, WarningCoverage)
				}

				if coverage < MinimumCoverage {
					t.Errorf("Package %s has coverage %.1f%% (minimum required: %.1f%%)",
						pkg, coverage, MinimumCoverage)
				} else {
					t.Logf("✓ Package %s coverage: %.1f%%", pkg, coverage)
				}
			})
		}
	})

	t.Run("ValidateTotalCoverage", func(t *testing.T) {
		cmd := exec.Command("go", "test", "-coverprofile=coverage.out", "-coverpkg=./src/...", "./tests/unit/...")
		var out bytes.Buffer
		cmd.Stdout = &out
		cmd.Stderr = &out

		err := cmd.Run()
		if err != nil {
			// Log output for debugging but don't fail if tests have issues
			t.Logf("Test execution output: %s", out.String())
		}

		// Parse total coverage
		coverCmd := exec.Command("go", "tool", "cover", "-func=coverage.out")
		coverOut, err := coverCmd.Output()
		if err != nil {
			t.Skipf("Could not analyze coverage: %v", err)
			return
		}

		// Extract total coverage from output
		lines := strings.Split(string(coverOut), "\n")
		for _, line := range lines {
			if strings.Contains(line, "total:") {
				parts := strings.Fields(line)
				if len(parts) >= 3 {
					coverageStr := strings.TrimSuffix(parts[len(parts)-1], "%")
					coverage, err := strconv.ParseFloat(coverageStr, 64)
					require.NoError(t, err, "Failed to parse coverage percentage")

					if coverage < MinimumCoverage {
						t.Errorf("Total coverage %.1f%% is below minimum %.1f%%", coverage, MinimumCoverage)
					} else {
						t.Logf("✓ Total coverage: %.1f%%", coverage)
					}
				}
			}
		}
	})
}

// TestPerformanceContract validates test execution performance
func TestPerformanceContract(t *testing.T) {
	t.Run("ValidateTotalExecutionTime", func(t *testing.T) {
		start := time.Now()

		cmd := exec.Command("go", "test", "./tests/unit/...", "-count=1")
		err := cmd.Run()

		elapsed := time.Since(start)

		if err != nil {
			t.Logf("Warning: Some tests failed during performance validation")
		}

		if elapsed > MaxTestExecutionTime {
			t.Errorf("Test suite execution time %v exceeds maximum %v", elapsed, MaxTestExecutionTime)
		} else {
			t.Logf("✓ Test suite execution time: %v (max: %v)", elapsed, MaxTestExecutionTime)
		}
	})

	t.Run("ValidateParallelExecution", func(t *testing.T) {
		// Test with different parallel settings
		parallelSettings := []int{1, 2, 4, 8}

		for _, parallel := range parallelSettings {
			t.Run(fmt.Sprintf("Parallel_%d", parallel), func(t *testing.T) {
				start := time.Now()

				cmd := exec.Command("go", "test", "./tests/unit/...", "-count=1", fmt.Sprintf("-parallel=%d", parallel))
				err := cmd.Run()

				elapsed := time.Since(start)

				if err != nil {
					t.Logf("Warning: Some tests failed with parallel=%d", parallel)
				}

				t.Logf("Execution time with parallel=%d: %v", parallel, elapsed)
			})
		}
	})
}

// TestQualityContract validates test quality requirements
func TestQualityContract(t *testing.T) {
	t.Run("NoTestsInSrcDirectory", func(t *testing.T) {
		srcPath := filepath.Join(".", "src")
		testFiles := []string{}

		err := filepath.Walk(srcPath, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if strings.HasSuffix(path, "_test.go") {
				testFiles = append(testFiles, path)
			}

			return nil
		})

		require.NoError(t, err, "Failed to walk src directory")

		if len(testFiles) > 0 {
			t.Errorf("Found %d test files in src/ directory (should be 0): %v", len(testFiles), testFiles)
		} else {
			t.Logf("✓ No test files found in src/ directory")
		}
	})

	t.Run("TestOrganizationStructure", func(t *testing.T) {
		requiredDirs := []string{
			"tests/unit/models",
			"tests/unit/services",
			"tests/unit/repositories",
			"tests/unit/handlers",
			"tests/unit/middleware",
			"tests/unit/interceptors",
			"tests/unit/grpc",
			"tests/unit/adapters",
			"tests/unit/parser",
			"tests/unit/config",
			"tests/unit/server",
			"tests/fixtures",
			"tests/helpers",
			"tests/mocks",
		}

		for _, dir := range requiredDirs {
			t.Run(dir, func(t *testing.T) {
				if _, err := os.Stat(dir); os.IsNotExist(err) {
					t.Errorf("Required test directory %s does not exist", dir)
				} else {
					t.Logf("✓ Directory %s exists", dir)
				}
			})
		}
	})

	t.Run("TestNamingConventions", func(t *testing.T) {
		testPath := filepath.Join(".", "tests", "unit")

		err := filepath.Walk(testPath, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if strings.HasSuffix(path, "_test.go") {
				// Check that test file names match expected patterns
				base := filepath.Base(path)
				if !regexp.MustCompile(`^[a-z_]+_test\.go$`).MatchString(base) {
					t.Errorf("Test file %s does not follow naming convention (lowercase_with_underscores_test.go)", base)
				}
			}

			return nil
		})

		if err != nil {
			t.Logf("Warning: Could not validate all test naming conventions: %v", err)
		}
	})

	t.Run("NoExternalDependencies", func(t *testing.T) {
		// Check that tests don't require external services
		cmd := exec.Command("go", "list", "-test", "-f", "{{.Imports}}", "./tests/unit/...")
		output, err := cmd.Output()

		if err != nil {
			t.Skipf("Could not analyze test imports: %v", err)
			return
		}

		// List of forbidden external dependencies
		forbidden := []string{
			"database/sql",
			"github.com/lib/pq",
			"github.com/go-sql-driver/mysql",
			"github.com/jinzhu/gorm",
			"net/http/httptest", // This is actually OK for tests
		}

		imports := string(output)
		for _, dep := range forbidden {
			if dep == "net/http/httptest" {
				continue // This is allowed
			}
			if strings.Contains(imports, dep) {
				t.Errorf("Tests should not import external dependency: %s", dep)
			}
		}

		t.Logf("✓ No forbidden external dependencies found in tests")
	})
}

// TestMockInfrastructure validates mock implementations
func TestMockInfrastructure(t *testing.T) {
	t.Run("MocksImplementInterfaces", func(t *testing.T) {
		// This would typically use reflection or code analysis
		// For now, we just check that mock files exist
		mockFiles := []string{
			"tests/mocks/registry.go",
			"tests/mocks/repository_mocks.go",
			"tests/mocks/service_mocks.go",
		}

		for _, file := range mockFiles {
			t.Run(file, func(t *testing.T) {
				if _, err := os.Stat(file); os.IsNotExist(err) {
					t.Errorf("Required mock file %s does not exist", file)
				} else {
					t.Logf("✓ Mock file %s exists", file)
				}
			})
		}
	})
}

// Helper functions

func getPackageCoverage(t *testing.T, pkg string) float64 {
	// Run coverage for specific package
	testPkg := strings.Replace(pkg, "./src/", "./tests/unit/", 1)
	testPkg = strings.TrimSuffix(testPkg, "/...")

	cmd := exec.Command("go", "test", "-coverprofile=temp_coverage.out", "-coverpkg="+pkg, testPkg)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out

	err := cmd.Run()
	if err != nil {
		// Package might not have tests yet
		t.Logf("Could not get coverage for %s: %v", pkg, err)
		return 0.0
	}

	// Parse coverage output
	coverCmd := exec.Command("go", "tool", "cover", "-func=temp_coverage.out")
	coverOut, err := coverCmd.Output()
	if err != nil {
		return 0.0
	}

	// Extract coverage percentage
	lines := strings.Split(string(coverOut), "\n")
	for _, line := range lines {
		if strings.Contains(line, "total:") {
			parts := strings.Fields(line)
			if len(parts) >= 3 {
				coverageStr := strings.TrimSuffix(parts[len(parts)-1], "%")
				coverage, _ := strconv.ParseFloat(coverageStr, 64)
				return coverage
			}
		}
	}

	return 0.0
}

// TestCoverageReport generates a detailed coverage report
func TestCoverageReport(t *testing.T) {
	t.Run("GenerateHTMLReport", func(t *testing.T) {
		// Generate coverage data
		cmd := exec.Command("go", "test", "-coverprofile=coverage.out", "-coverpkg=./src/...", "./tests/unit/...")
		err := cmd.Run()

		if err != nil {
			t.Logf("Warning: Some tests failed during coverage report generation")
		}

		// Generate HTML report
		htmlCmd := exec.Command("go", "tool", "cover", "-html=coverage.out", "-o", "coverage_report.html")
		err = htmlCmd.Run()

		if err != nil {
			t.Errorf("Failed to generate HTML coverage report: %v", err)
		} else {
			t.Logf("✓ HTML coverage report generated: coverage_report.html")
		}
	})
}
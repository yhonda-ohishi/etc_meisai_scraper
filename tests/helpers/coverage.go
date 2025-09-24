package helpers

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"testing"
)

// CoverageInfo holds coverage information for a package
type CoverageInfo struct {
	Package    string
	Coverage   float64
	Statements int
	Covered    int
}

// ValidateCoverage checks that a package has the required coverage percentage
func ValidateCoverage(t *testing.T, packagePath string, requiredCoverage float64) {
	t.Helper()

	coverage, err := GetPackageCoverage(packagePath)
	if err != nil {
		t.Fatalf("Failed to get coverage for package %s: %v", packagePath, err)
	}

	if coverage.Coverage < requiredCoverage {
		t.Errorf("Package %s has insufficient coverage: %.2f%% (required: %.2f%%)",
			packagePath, coverage.Coverage, requiredCoverage)
	}
}

// GetPackageCoverage returns coverage information for a specific package
func GetPackageCoverage(packagePath string) (*CoverageInfo, error) {
	// Run go test with coverage for the specific package
	cmd := exec.Command("go", "test", "-cover", packagePath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to run coverage test: %v, output: %s", err, string(output))
	}

	return parseCoverageOutput(string(output), packagePath)
}

// GetOverallCoverage returns overall coverage for all src/ packages
func GetOverallCoverage() (*CoverageInfo, error) {
	// Run go test with coverage profile for all src/ packages
	cmd := exec.Command("go", "test", "-coverprofile=coverage.out", "./src/...")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to run overall coverage test: %v, output: %s", err, string(output))
	}

	// Get coverage summary
	cmd = exec.Command("go", "tool", "cover", "-func=coverage.out")
	output, err = cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to get coverage summary: %v", err)
	}

	return parseOverallCoverage(string(output))
}

// RequireFullCoverage ensures that all src/ packages have 100% coverage
func RequireFullCoverage(t *testing.T) {
	t.Helper()

	coverage, err := GetOverallCoverage()
	if err != nil {
		t.Fatalf("Failed to get overall coverage: %v", err)
	}

	if coverage.Coverage < 100.0 {
		t.Fatalf("Overall coverage is insufficient: %.2f%% (required: 100.0%%)", coverage.Coverage)
	}
}

// parseCoverageOutput parses the output from go test -cover
func parseCoverageOutput(output, packagePath string) (*CoverageInfo, error) {
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.Contains(line, "coverage:") {
			parts := strings.Fields(line)
			for i, part := range parts {
				if part == "coverage:" && i+1 < len(parts) {
					coverageStr := strings.TrimSuffix(parts[i+1], "%")
					coverage, err := strconv.ParseFloat(coverageStr, 64)
					if err != nil {
						return nil, fmt.Errorf("failed to parse coverage percentage: %v", err)
					}

					return &CoverageInfo{
						Package:  packagePath,
						Coverage: coverage,
					}, nil
				}
			}
		}
	}

	return nil, fmt.Errorf("coverage information not found in output")
}

// parseOverallCoverage parses the output from go tool cover -func
func parseOverallCoverage(output string) (*CoverageInfo, error) {
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.Contains(line, "total:") {
			parts := strings.Fields(line)
			if len(parts) >= 3 {
				coverageStr := strings.TrimSuffix(parts[len(parts)-1], "%")
				coverage, err := strconv.ParseFloat(coverageStr, 64)
				if err != nil {
					return nil, fmt.Errorf("failed to parse total coverage: %v", err)
				}

				return &CoverageInfo{
					Package:  "total",
					Coverage: coverage,
				}, nil
			}
		}
	}

	return nil, fmt.Errorf("total coverage not found in output")
}

// GenerateHTMLReport generates an HTML coverage report
func GenerateHTMLReport(outputFile string) error {
	cmd := exec.Command("go", "tool", "cover", "-html=coverage.out", "-o", outputFile)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to generate HTML report: %v, output: %s", err, string(output))
	}
	return nil
}
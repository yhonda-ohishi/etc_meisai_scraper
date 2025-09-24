package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// ContractTestRunner provides a convenient way to run contract tests
type ContractTestRunner struct {
	testDir     string
	verbose     bool
	race        bool
	cover       bool
	timeout     time.Duration
	pattern     string
	parallel    int
	outputDir   string
}

func main() {
	runner := &ContractTestRunner{}

	// Parse command line flags
	flag.StringVar(&runner.testDir, "dir", ".", "Test directory")
	flag.BoolVar(&runner.verbose, "v", false, "Verbose output")
	flag.BoolVar(&runner.race, "race", false, "Enable race detection")
	flag.BoolVar(&runner.cover, "cover", false, "Enable coverage analysis")
	flag.DurationVar(&runner.timeout, "timeout", 10*time.Minute, "Test timeout")
	flag.StringVar(&runner.pattern, "run", "", "Run only tests matching pattern")
	flag.IntVar(&runner.parallel, "parallel", 4, "Number of parallel tests")
	flag.StringVar(&runner.outputDir, "output", "output", "Output directory for reports")

	flag.Parse()

	if err := runner.Run(); err != nil {
		log.Fatalf("Contract test runner failed: %v", err)
	}
}

func (r *ContractTestRunner) Run() error {
	fmt.Println("🧪 ETC明細 Contract Test Runner")
	fmt.Println("================================")

	// Create output directory
	if err := os.MkdirAll(r.outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Run test suites
	testSuites := []TestSuite{
		{
			Name:        "T010-A: gRPC Service Contracts",
			Pattern:     "TestGRPCServiceContract",
			Description: "Tests all gRPC service method contracts",
		},
		{
			Name:        "T010-B: API Version Compatibility",
			Pattern:     "TestAPIVersionCompatibility",
			Description: "Tests backward and forward API compatibility",
		},
		{
			Name:        "T010-C: Schema Evolution",
			Pattern:     "TestSchemaEvolution",
			Description: "Tests Protocol Buffer schema evolution",
		},
		{
			Name:        "T010-D: End-to-End Workflows",
			Pattern:     "TestEndToEndWorkflow",
			Description: "Tests complete ETC data processing workflows",
		},
		{
			Name:        "T010-E: Performance SLAs",
			Pattern:     "TestPerformanceSLA",
			Description: "Tests performance SLA compliance",
		},
		{
			Name:        "Legacy: Mock Generation",
			Pattern:     "TestMockGenerationContract",
			Description: "Tests mock generation contracts",
		},
		{
			Name:        "Legacy: Test Execution",
			Pattern:     "TestExecutionContract",
			Description: "Tests service execution contracts",
		},
	}

	var results []TestResult
	for _, suite := range testSuites {
		if r.pattern != "" && !strings.Contains(suite.Pattern, r.pattern) {
			continue
		}

		fmt.Printf("\n🔍 Running: %s\n", suite.Name)
		fmt.Printf("   %s\n", suite.Description)

		result := r.runTestSuite(suite)
		results = append(results, result)

		if result.Success {
			fmt.Printf("   ✅ PASSED (%.2fs)\n", result.Duration.Seconds())
		} else {
			fmt.Printf("   ❌ FAILED (%.2fs)\n", result.Duration.Seconds())
			if result.Error != "" {
				fmt.Printf("   Error: %s\n", result.Error)
			}
		}
	}

	// Generate summary report
	r.generateSummaryReport(results)

	// Check overall success
	failedCount := 0
	for _, result := range results {
		if !result.Success {
			failedCount++
		}
	}

	fmt.Printf("\n📊 Summary: %d passed, %d failed\n", len(results)-failedCount, failedCount)

	if failedCount > 0 {
		return fmt.Errorf("%d test suites failed", failedCount)
	}

	fmt.Println("🎉 All contract tests passed!")
	return nil
}

func (r *ContractTestRunner) runTestSuite(suite TestSuite) TestResult {
	start := time.Now()

	// Build test command
	args := []string{"test"}

	if r.verbose {
		args = append(args, "-v")
	}

	if r.race {
		args = append(args, "-race")
	}

	if r.cover {
		coverFile := filepath.Join(r.outputDir, fmt.Sprintf("%s.coverage", suite.Pattern))
		args = append(args, "-cover", "-coverprofile="+coverFile)
	}

	args = append(args, "-timeout", r.timeout.String())
	args = append(args, "-parallel", fmt.Sprintf("%d", r.parallel))
	args = append(args, "-run", suite.Pattern)
	args = append(args, r.testDir)

	// Execute test
	ctx, cancel := context.WithTimeout(context.Background(), r.timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "go", args...)
	cmd.Dir = r.testDir

	output, err := cmd.CombinedOutput()

	duration := time.Since(start)
	success := err == nil

	// Save output to file
	outputFile := filepath.Join(r.outputDir, fmt.Sprintf("%s.output", suite.Pattern))
	if err := os.WriteFile(outputFile, output, 0644); err != nil {
		log.Printf("Failed to save output file: %v", err)
	}

	errorMsg := ""
	if err != nil {
		errorMsg = err.Error()
	}

	return TestResult{
		Suite:    suite,
		Success:  success,
		Duration: duration,
		Output:   string(output),
		Error:    errorMsg,
	}
}

func (r *ContractTestRunner) generateSummaryReport(results []TestResult) {
	reportPath := filepath.Join(r.outputDir, "contract_test_summary.md")

	report := strings.Builder{}
	report.WriteString("# Contract Test Summary Report\n\n")
	report.WriteString(fmt.Sprintf("Generated: %s\n\n", time.Now().Format(time.RFC3339)))

	report.WriteString("## Test Results\n\n")
	for _, result := range results {
		status := "✅ PASSED"
		if !result.Success {
			status = "❌ FAILED"
		}

		report.WriteString(fmt.Sprintf("### %s\n", result.Suite.Name))
		report.WriteString(fmt.Sprintf("- **Status**: %s\n", status))
		report.WriteString(fmt.Sprintf("- **Duration**: %.2fs\n", result.Duration.Seconds()))
		report.WriteString(fmt.Sprintf("- **Description**: %s\n", result.Suite.Description))

		if !result.Success {
			report.WriteString(fmt.Sprintf("- **Error**: %s\n", result.Error))
		}

		report.WriteString("\n")
	}

	// Add performance metrics if available
	report.WriteString("## Performance Metrics\n\n")
	for _, result := range results {
		if strings.Contains(result.Suite.Pattern, "Performance") {
			report.WriteString(fmt.Sprintf("### %s\n", result.Suite.Name))
			report.WriteString("```\n")
			report.WriteString(result.Output)
			report.WriteString("```\n\n")
		}
	}

	if err := os.WriteFile(reportPath, []byte(report.String()), 0644); err != nil {
		log.Printf("Failed to generate summary report: %v", err)
	} else {
		fmt.Printf("📝 Summary report: %s\n", reportPath)
	}
}

// TestSuite represents a contract test suite
type TestSuite struct {
	Name        string
	Pattern     string
	Description string
}

// TestResult holds the result of running a test suite
type TestResult struct {
	Suite    TestSuite
	Success  bool
	Duration time.Duration
	Output   string
	Error    string
}
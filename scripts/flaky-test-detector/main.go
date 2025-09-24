// T013-C: Test flakiness detection and elimination
// This tool identifies and helps eliminate flaky tests

package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

// FlakinessReport contains flaky test analysis results
type FlakinessReport struct {
	TotalRuns       int                    `json:"total_runs"`
	TestsAnalyzed   int                    `json:"tests_analyzed"`
	FlakyTests      []FlakyTest            `json:"flaky_tests"`
	StableTests     int                    `json:"stable_tests"`
	FlakinessRate   float64                `json:"flakiness_rate"`
	Categories      map[string][]FlakyTest `json:"categories"`
	Recommendations []Recommendation       `json:"recommendations"`
	RunDuration     time.Duration          `json:"run_duration"`
}

// FlakyTest represents a test that shows flaky behavior
type FlakyTest struct {
	Package        string          `json:"package"`
	Name           string          `json:"name"`
	PassRate       float64         `json:"pass_rate"`
	FailureCount   int             `json:"failure_count"`
	TotalRuns      int             `json:"total_runs"`
	FailureReasons []FailureReason `json:"failure_reasons"`
	Category       string          `json:"category"`
	Severity       string          `json:"severity"`
}

// FailureReason contains details about why a test failed
type FailureReason struct {
	Run      int           `json:"run"`
	Error    string        `json:"error"`
	Output   string        `json:"output"`
	Duration time.Duration `json:"duration"`
}

// Recommendation for fixing flaky tests
type Recommendation struct {
	Test        string `json:"test"`
	Category    string `json:"category"`
	Description string `json:"description"`
	Priority    string `json:"priority"`
	Solution    string `json:"solution"`
}

// FlakyTestDetector identifies flaky tests
type FlakyTestDetector struct {
	rootDir     string
	runs        int
	parallel    int
	verbose     bool
	testResults map[string]*TestRunData
	mu          sync.Mutex
}

// TestRunData tracks results across multiple runs
type TestRunData struct {
	Package  string
	Name     string
	Passes   int
	Failures int
	Runs     []RunResult
}

// RunResult contains data from a single test run
type RunResult struct {
	Run      int
	Passed   bool
	Duration time.Duration
	Error    string
	Output   string
}

// NewFlakyTestDetector creates a new detector
func NewFlakyTestDetector(rootDir string, runs int) *FlakyTestDetector {
	return &FlakyTestDetector{
		rootDir:     rootDir,
		runs:        runs,
		parallel:    4,
		verbose:     false,
		testResults: make(map[string]*TestRunData),
	}
}

// Detect runs the flaky test detection
func (d *FlakyTestDetector) Detect() (*FlakinessReport, error) {
	report := &FlakinessReport{
		TotalRuns:  d.runs,
		Categories: make(map[string][]FlakyTest),
	}

	startTime := time.Now()
	fmt.Printf("Starting flaky test detection with %d runs...\n", d.runs)

	// Get list of packages
	packages, err := d.getPackages()
	if err != nil {
		return nil, fmt.Errorf("failed to get packages: %w", err)
	}

	// Run tests multiple times
	for run := 1; run <= d.runs; run++ {
		fmt.Printf("Run %d/%d...\n", run, d.runs)
		if err := d.runTests(packages, run); err != nil {
			fmt.Printf("Warning: run %d had errors: %v\n", run, err)
		}
	}

	// Analyze results
	d.analyzeResults(report)

	// Generate recommendations
	d.generateRecommendations(report)

	report.RunDuration = time.Since(startTime)
	fmt.Printf("Analysis complete in %v\n", report.RunDuration)

	return report, nil
}

// getPackages returns list of packages to test
func (d *FlakyTestDetector) getPackages() ([]string, error) {
	cmd := exec.Command("go", "list", "./...")
	cmd.Dir = d.rootDir

	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var packages []string
	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	for scanner.Scan() {
		pkg := scanner.Text()
		if !strings.Contains(pkg, "/vendor/") {
			packages = append(packages, pkg)
		}
	}

	return packages, nil
}

// runTests executes tests for all packages
func (d *FlakyTestDetector) runTests(packages []string, runNumber int) error {
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, d.parallel)

	for _, pkg := range packages {
		wg.Add(1)
		go func(p string) {
			defer wg.Done()
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			d.runPackageTests(p, runNumber)
		}(pkg)
	}

	wg.Wait()
	return nil
}

// runPackageTests runs tests for a single package
func (d *FlakyTestDetector) runPackageTests(pkg string, runNumber int) {
	// Run with -json for structured output
	cmd := exec.Command("go", "test", "-json", "-count=1", pkg)
	cmd.Dir = d.rootDir

	output, err := cmd.Output()

	// Parse JSON output
	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	testStartTimes := make(map[string]time.Time)

	for scanner.Scan() {
		var event struct {
			Time    string  `json:"Time"`
			Action  string  `json:"Action"`
			Package string  `json:"Package"`
			Test    string  `json:"Test"`
			Output  string  `json:"Output"`
			Elapsed float64 `json:"Elapsed"`
		}

		if err := json.Unmarshal(scanner.Bytes(), &event); err != nil {
			continue
		}

		if event.Test == "" {
			continue
		}

		testKey := fmt.Sprintf("%s.%s", pkg, event.Test)

		switch event.Action {
		case "run":
			testStartTimes[testKey] = time.Now()

		case "pass", "fail":
			d.mu.Lock()
			if d.testResults[testKey] == nil {
				d.testResults[testKey] = &TestRunData{
					Package: pkg,
					Name:    event.Test,
					Runs:    make([]RunResult, 0),
				}
			}

			result := RunResult{
				Run:      runNumber,
				Passed:   event.Action == "pass",
				Duration: time.Duration(event.Elapsed * float64(time.Second)),
			}

			if event.Action == "pass" {
				d.testResults[testKey].Passes++
			} else {
				d.testResults[testKey].Failures++
				result.Error = err.Error()
			}

			d.testResults[testKey].Runs = append(d.testResults[testKey].Runs, result)
			d.mu.Unlock()

		case "output":
			// Capture output for failed tests
			if strings.Contains(event.Output, "FAIL") ||
				strings.Contains(event.Output, "ERROR") ||
				strings.Contains(event.Output, "panic") {

				d.mu.Lock()
				if d.testResults[testKey] != nil && len(d.testResults[testKey].Runs) > 0 {
					lastRun := &d.testResults[testKey].Runs[len(d.testResults[testKey].Runs)-1]
					lastRun.Output += event.Output
				}
				d.mu.Unlock()
			}
		}
	}
}

// analyzeResults processes test results to identify flaky tests
func (d *FlakyTestDetector) analyzeResults(report *FlakinessReport) {
	for _, data := range d.testResults {
		report.TestsAnalyzed++

		passRate := float64(data.Passes) / float64(len(data.Runs)) * 100

		// Test is flaky if it doesn't have consistent results
		if passRate > 0 && passRate < 100 {
			flaky := FlakyTest{
				Package:      data.Package,
				Name:         data.Name,
				PassRate:     passRate,
				FailureCount: data.Failures,
				TotalRuns:    len(data.Runs),
			}

			// Collect failure reasons
			for _, run := range data.Runs {
				if !run.Passed {
					flaky.FailureReasons = append(flaky.FailureReasons, FailureReason{
						Run:      run.Run,
						Error:    run.Error,
						Output:   run.Output,
						Duration: run.Duration,
					})
				}
			}

			// Categorize the flaky test
			flaky.Category = d.categorizeFlakiness(flaky)
			flaky.Severity = d.calculateSeverity(flaky)

			report.FlakyTests = append(report.FlakyTests, flaky)
			report.Categories[flaky.Category] = append(report.Categories[flaky.Category], flaky)
		} else if passRate == 100 {
			report.StableTests++
		}
	}

	// Sort flaky tests by failure rate
	sort.Slice(report.FlakyTests, func(i, j int) bool {
		return report.FlakyTests[i].PassRate < report.FlakyTests[j].PassRate
	})

	// Calculate overall flakiness rate
	if report.TestsAnalyzed > 0 {
		report.FlakinessRate = float64(len(report.FlakyTests)) / float64(report.TestsAnalyzed) * 100
	}
}

// categorizeFlakiness determines the type of flakiness
func (d *FlakyTestDetector) categorizeFlakiness(test FlakyTest) string {
	// Analyze failure reasons to categorize
	hasTimeout := false
	hasRaceCondition := false
	hasResourceIssue := false
	hasNetworkIssue := false
	hasRandomness := false

	for _, failure := range test.FailureReasons {
		output := strings.ToLower(failure.Output + failure.Error)

		if strings.Contains(output, "timeout") ||
			strings.Contains(output, "deadline exceeded") {
			hasTimeout = true
		}
		if strings.Contains(output, "race") ||
			strings.Contains(output, "concurrent") ||
			strings.Contains(output, "goroutine") {
			hasRaceCondition = true
		}
		if strings.Contains(output, "connection") ||
			strings.Contains(output, "network") ||
			strings.Contains(output, "dial") {
			hasNetworkIssue = true
		}
		if strings.Contains(output, "file") ||
			strings.Contains(output, "permission") ||
			strings.Contains(output, "resource") {
			hasResourceIssue = true
		}
		if strings.Contains(output, "random") ||
			strings.Contains(output, "seed") {
			hasRandomness = true
		}
	}

	// Determine primary category
	if hasRaceCondition {
		return "race_condition"
	}
	if hasTimeout {
		return "timing"
	}
	if hasNetworkIssue {
		return "network"
	}
	if hasResourceIssue {
		return "resource"
	}
	if hasRandomness {
		return "randomness"
	}

	// Check for test order dependency
	if d.hasOrderDependency(test) {
		return "order_dependency"
	}

	return "unknown"
}

// hasOrderDependency checks if failures correlate with test execution order
func (d *FlakyTestDetector) hasOrderDependency(test FlakyTest) bool {
	// Simple heuristic: if failures are clustered in specific runs
	if len(test.FailureReasons) < 2 {
		return false
	}

	// Check if failures happen in sequence
	lastRun := -1
	sequential := 0
	for _, failure := range test.FailureReasons {
		if lastRun != -1 && failure.Run == lastRun+1 {
			sequential++
		}
		lastRun = failure.Run
	}

	return sequential > len(test.FailureReasons)/2
}

// calculateSeverity determines how severe the flakiness is
func (d *FlakyTestDetector) calculateSeverity(test FlakyTest) string {
	if test.PassRate < 50 {
		return "critical"
	}
	if test.PassRate < 80 {
		return "high"
	}
	if test.PassRate < 95 {
		return "medium"
	}
	return "low"
}

// generateRecommendations creates fix recommendations
func (d *FlakyTestDetector) generateRecommendations(report *FlakinessReport) {
	for _, test := range report.FlakyTests {
		rec := Recommendation{
			Test:     fmt.Sprintf("%s.%s", test.Package, test.Name),
			Category: test.Category,
			Priority: test.Severity,
		}

		switch test.Category {
		case "race_condition":
			rec.Description = "Test has race conditions"
			rec.Solution = "Add proper synchronization, use sync.WaitGroup or channels, ensure t.Parallel() is used correctly"

		case "timing":
			rec.Description = "Test has timing dependencies"
			rec.Solution = "Replace time.Sleep with proper synchronization, use testify/eventually for async operations, increase timeouts"

		case "network":
			rec.Description = "Test depends on network resources"
			rec.Solution = "Mock network calls, use httptest.Server for HTTP testing, ensure proper cleanup"

		case "resource":
			rec.Description = "Test has resource management issues"
			rec.Solution = "Ensure proper cleanup with defer, use t.TempDir() for temp files, check file handles are closed"

		case "order_dependency":
			rec.Description = "Test depends on execution order"
			rec.Solution = "Ensure test isolation, reset global state, use setup/teardown functions"

		case "randomness":
			rec.Description = "Test uses random values without seeding"
			rec.Solution = "Use fixed random seed for tests, or make test deterministic"

		default:
			rec.Description = "Test shows intermittent failures"
			rec.Solution = "Review test for hidden dependencies, add logging to identify failure patterns"
		}

		report.Recommendations = append(report.Recommendations, rec)
	}

	// Sort by priority
	priority := map[string]int{
		"critical": 0,
		"high":     1,
		"medium":   2,
		"low":      3,
	}

	sort.Slice(report.Recommendations, func(i, j int) bool {
		return priority[report.Recommendations[i].Priority] < priority[report.Recommendations[j].Priority]
	})
}

// GenerateReport creates output files
func (d *FlakyTestDetector) GenerateReport(report *FlakinessReport, outputDir string) error {
	// Ensure output directory exists
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return err
	}

	// Generate JSON report
	jsonPath := filepath.Join(outputDir, "flaky-tests.json")
	jsonData, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(jsonPath, jsonData, 0644); err != nil {
		return err
	}

	// Generate text report
	textPath := filepath.Join(outputDir, "flaky-tests.txt")
	if err := d.generateTextReport(report, textPath); err != nil {
		return err
	}

	// Generate fix script
	scriptPath := filepath.Join(outputDir, "fix-flaky-tests.sh")
	if err := d.generateFixScript(report, scriptPath); err != nil {
		return err
	}

	fmt.Printf("Reports generated:\n")
	fmt.Printf("  - %s\n", jsonPath)
	fmt.Printf("  - %s\n", textPath)
	fmt.Printf("  - %s\n", scriptPath)

	return nil
}

// generateTextReport creates human-readable report
func (d *FlakyTestDetector) generateTextReport(report *FlakinessReport, path string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	fmt.Fprintln(file, "FLAKY TEST DETECTION REPORT")
	fmt.Fprintln(file, strings.Repeat("=", 60))
	fmt.Fprintln(file)

	fmt.Fprintf(file, "Total Test Runs: %d\n", report.TotalRuns)
	fmt.Fprintf(file, "Tests Analyzed: %d\n", report.TestsAnalyzed)
	fmt.Fprintf(file, "Flaky Tests Found: %d\n", len(report.FlakyTests))
	fmt.Fprintf(file, "Stable Tests: %d\n", report.StableTests)
	fmt.Fprintf(file, "Flakiness Rate: %.2f%%\n", report.FlakinessRate)
	fmt.Fprintf(file, "Analysis Duration: %v\n", report.RunDuration)
	fmt.Fprintln(file)

	if len(report.FlakyTests) > 0 {
		fmt.Fprintln(file, "FLAKY TESTS:")
		fmt.Fprintln(file, strings.Repeat("-", 60))

		for i, test := range report.FlakyTests {
			fmt.Fprintf(file, "\n%d. %s.%s\n", i+1, test.Package, test.Name)
			fmt.Fprintf(file, "   Pass Rate: %.1f%% (%d/%d passed)\n",
				test.PassRate, test.TotalRuns-test.FailureCount, test.TotalRuns)
			fmt.Fprintf(file, "   Category: %s\n", test.Category)
			fmt.Fprintf(file, "   Severity: %s\n", test.Severity)

			if len(test.FailureReasons) > 0 && test.FailureReasons[0].Output != "" {
				fmt.Fprintf(file, "   Sample Failure Output:\n")
				lines := strings.Split(test.FailureReasons[0].Output, "\n")
				for j, line := range lines {
					if j >= 3 {
						break
					}
					fmt.Fprintf(file, "     %s\n", strings.TrimSpace(line))
				}
			}
		}
		fmt.Fprintln(file)
	}

	// Category breakdown
	fmt.Fprintln(file, "FLAKINESS BY CATEGORY:")
	fmt.Fprintln(file, strings.Repeat("-", 60))
	for category, tests := range report.Categories {
		fmt.Fprintf(file, "%s: %d tests\n", category, len(tests))
	}
	fmt.Fprintln(file)

	// Recommendations
	if len(report.Recommendations) > 0 {
		fmt.Fprintln(file, "TOP RECOMMENDATIONS:")
		fmt.Fprintln(file, strings.Repeat("-", 60))

		for i, rec := range report.Recommendations {
			if i >= 10 {
				break
			}
			fmt.Fprintf(file, "\n[%s] %s\n", strings.ToUpper(rec.Priority), rec.Test)
			fmt.Fprintf(file, "Issue: %s\n", rec.Description)
			fmt.Fprintf(file, "Fix: %s\n", rec.Solution)
		}
	}

	return nil
}

// generateFixScript creates a script to help fix flaky tests
func (d *FlakyTestDetector) generateFixScript(report *FlakinessReport, path string) error {
	script := `#!/bin/bash
# Script to help fix flaky tests
# Generated by flaky-test-detector

set -e

echo "Flaky Test Fix Helper"
echo "===================="
echo ""
echo "Found ${#report.FlakyTests} flaky tests"
echo ""

# Function to add t.Parallel() to tests
add_parallel() {
    local file=$1
    local test=$2
    echo "Adding t.Parallel() to $test in $file"
    # This is a placeholder - actual implementation would modify the file
}

# Function to increase timeout
increase_timeout() {
    local file=$1
    local test=$2
    echo "Increasing timeout for $test in $file"
    # This is a placeholder - actual implementation would modify the file
}

# Function to add retry logic
add_retry() {
    local file=$1
    local test=$2
    echo "Adding retry logic to $test in $file"
    # This is a placeholder - actual implementation would modify the file
}

`

	// Add specific fixes for each flaky test
	for _, test := range report.FlakyTests {
		testFile := strings.Replace(test.Package, "github.com/yhonda-ohishi/etc_meisai/", "", 1)
		testFile = strings.Replace(testFile, ".", "/", -1) + "_test.go"

		script += fmt.Sprintf("\n# Fix for %s.%s (Category: %s)\n", test.Package, test.Name, test.Category)
		script += fmt.Sprintf("echo \"Fixing %s.%s...\"\n", test.Package, test.Name)

		switch test.Category {
		case "race_condition":
			script += fmt.Sprintf("add_parallel \"%s\" \"%s\"\n", testFile, test.Name)
		case "timing":
			script += fmt.Sprintf("increase_timeout \"%s\" \"%s\"\n", testFile, test.Name)
		case "network", "resource":
			script += fmt.Sprintf("add_retry \"%s\" \"%s\"\n", testFile, test.Name)
		}
	}

	script += `
echo ""
echo "Fix suggestions have been displayed."
echo "Please manually review and apply the appropriate fixes."
echo ""
echo "Common fixes:"
echo "1. Add t.Parallel() for tests that can run concurrently"
echo "2. Use sync.WaitGroup instead of time.Sleep"
echo "3. Mock external dependencies"
echo "4. Ensure proper cleanup with defer"
echo "5. Use fixed random seeds"
`

	return os.WriteFile(path, []byte(script), 0755)
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run scripts/flaky-test-detector/main.go <project-root> [runs] [output-dir]")
		fmt.Println("   or: flaky-test-detector <project-root> [runs] [output-dir]")
		fmt.Println("")
		fmt.Println("Arguments:")
		fmt.Println("  project-root: Root directory of the Go project to analyze")
		fmt.Println("  runs:         Number of test runs (default: 10)")
		fmt.Println("  output-dir:   Directory for output files (default: flaky-test-report)")
		fmt.Println("")
		fmt.Println("Options:")
		fmt.Println("  --quick      Run only 3 iterations for quick check")
		fmt.Println("  --parallel   Run tests in parallel (experimental)")
		os.Exit(1)
	}

	// Parse arguments
	projectRoot := ""
	runs := 10
	outputDir := "flaky-test-report"
	quickMode := false

	// Simple argument parsing
	argIdx := 1
	if os.Args[argIdx] == "--quick" {
		quickMode = true
		runs = 3
		argIdx++
		if argIdx >= len(os.Args) {
			fmt.Println("Error: project-root is required after --quick")
			os.Exit(1)
		}
	}

	projectRoot = os.Args[argIdx]
	argIdx++

	// Parse runs if provided and not in quick mode
	if argIdx < len(os.Args) && !quickMode {
		if r, err := fmt.Sscanf(os.Args[argIdx], "%d", &runs); err == nil && r == 1 {
			argIdx++
		}
	}

	// Parse output dir if provided
	if argIdx < len(os.Args) {
		outputDir = os.Args[argIdx]
	}

	if quickMode {
		fmt.Printf("Running in quick mode with %d iterations\n", runs)
	}

	detector := NewFlakyTestDetector(projectRoot, runs)

	// Run detection
	report, err := detector.Detect()
	if err != nil {
		fmt.Printf("Error detecting flaky tests: %v\n", err)
		os.Exit(1)
	}

	// Generate reports
	if err := detector.GenerateReport(report, outputDir); err != nil {
		fmt.Printf("Error generating reports: %v\n", err)
		os.Exit(1)
	}

	// Summary
	fmt.Printf("\n")
	fmt.Printf("Flaky Test Detection Summary:\n")
	fmt.Printf("  Flaky tests found: %d\n", len(report.FlakyTests))
	fmt.Printf("  Flakiness rate: %.2f%%\n", report.FlakinessRate)

	if len(report.FlakyTests) > 0 {
		fmt.Printf("\nTop flaky tests:\n")
		for i, test := range report.FlakyTests {
			if i >= 5 {
				break
			}
			fmt.Printf("  - %s.%s (%.1f%% pass rate)\n",
				test.Package, test.Name, test.PassRate)
		}

		// Exit with error to indicate flaky tests found
		os.Exit(1)
	}

	fmt.Println("\n✅ No flaky tests detected!")
}

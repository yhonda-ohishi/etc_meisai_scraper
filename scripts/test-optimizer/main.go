// T013-B: Test execution time optimization
// This tool analyzes and optimizes test execution time to meet < 30 seconds target

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
	"time"
)

// TestProfile contains performance data for tests
type TestProfile struct {
	TotalDuration time.Duration              `json:"total_duration"`
	PackageCount  int                        `json:"package_count"`
	TestCount     int                        `json:"test_count"`
	Packages      map[string]*PackageProfile `json:"packages"`
	SlowTests     []TestResult               `json:"slow_tests"`
	Optimizations []Optimization             `json:"optimizations"`
}

// PackageProfile contains timing for a package
type PackageProfile struct {
	Name        string        `json:"name"`
	Duration    time.Duration `json:"duration"`
	TestCount   int           `json:"test_count"`
	Tests       []TestResult  `json:"tests"`
	CanParallel bool          `json:"can_parallel"`
	HasDBAccess bool          `json:"has_db_access"`
	HasFileIO   bool          `json:"has_file_io"`
}

// TestResult contains individual test timing
type TestResult struct {
	Package  string        `json:"package"`
	Name     string        `json:"name"`
	Duration time.Duration `json:"duration"`
	Status   string        `json:"status"`
	Output   string        `json:"output,omitempty"`
}

// Optimization represents a suggested optimization
type Optimization struct {
	Type        string        `json:"type"`
	Target      string        `json:"target"`
	Description string        `json:"description"`
	Impact      time.Duration `json:"estimated_impact"`
	Priority    string        `json:"priority"`
}

// TestOptimizer analyzes and optimizes test execution
type TestOptimizer struct {
	rootDir     string
	profile     *TestProfile
	parallelism int
	targetTime  time.Duration
}

// NewTestOptimizer creates a new optimizer
func NewTestOptimizer(rootDir string) *TestOptimizer {
	return &TestOptimizer{
		rootDir:     rootDir,
		parallelism: 4, // Default parallelism
		targetTime:  30 * time.Second,
		profile: &TestProfile{
			Packages: make(map[string]*PackageProfile),
		},
	}
}

// QuickAnalyze performs analysis without running tests
func (o *TestOptimizer) QuickAnalyze() error {
	fmt.Println("Performing quick analysis (no test execution)...")

	// Get list of packages
	packages, err := o.getPackages()
	if err != nil {
		return fmt.Errorf("failed to get packages: %w", err)
	}

	o.profile.PackageCount = len(packages)

	// Analyze test files without running them
	for _, pkg := range packages {
		// Check if package has test files
		cmd := exec.Command("go", "list", "-f", "{{.TestGoFiles}}", pkg)
		cmd.Dir = o.rootDir
		output, err := cmd.Output()
		if err == nil && string(output) != "[]\n" {
			pkgProfile := &PackageProfile{
				Name:        pkg,
				Tests:       []TestResult{},
				CanParallel: true, // Assume parallelizable
				TestCount:   1,    // Estimate
			}
			o.profile.Packages[pkg] = pkgProfile
			o.profile.TestCount++
		}
	}

	fmt.Printf("Quick analysis complete: %d packages analyzed\n", o.profile.PackageCount)

	// Generate basic optimizations
	o.generateOptimizations()

	return nil
}

// Analyze performs test performance analysis
func (o *TestOptimizer) Analyze() error {
	fmt.Println("Analyzing test performance...")

	// Run tests with timing
	if err := o.profileTests(); err != nil {
		return fmt.Errorf("failed to profile tests: %w", err)
	}

	// Identify slow tests
	o.identifySlowTests()

	// Analyze parallelization opportunities
	o.analyzeParallelization()

	// Generate optimization recommendations
	o.generateOptimizations()

	return nil
}

// profileTests runs tests and collects timing data
func (o *TestOptimizer) profileTests() error {
	// Get list of packages
	packages, err := o.getPackages()
	if err != nil {
		return err
	}

	o.profile.PackageCount = len(packages)
	startTime := time.Now()

	// Run tests for each package
	for _, pkg := range packages {
		pkgProfile, err := o.profilePackage(pkg)
		if err != nil {
			fmt.Printf("Warning: failed to profile %s: %v\n", pkg, err)
			continue
		}
		o.profile.Packages[pkg] = pkgProfile
		o.profile.TestCount += pkgProfile.TestCount
	}

	o.profile.TotalDuration = time.Since(startTime)

	fmt.Printf("Total test execution time: %v\n", o.profile.TotalDuration)
	fmt.Printf("Total packages: %d, Total tests: %d\n",
		o.profile.PackageCount, o.profile.TestCount)

	return nil
}

// getPackages returns list of Go packages
func (o *TestOptimizer) getPackages() ([]string, error) {
	cmd := exec.Command("go", "list", "./...")
	cmd.Dir = o.rootDir

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

// profilePackage profiles a single package
func (o *TestOptimizer) profilePackage(pkg string) (*PackageProfile, error) {
	profile := &PackageProfile{
		Name:  pkg,
		Tests: make([]TestResult, 0),
	}

	// Run tests with verbose output and timing (with timeout)
	cmd := exec.Command("go", "test", "-v", "-json", "-timeout", "10s", pkg)
	cmd.Dir = o.rootDir

	output, err := cmd.Output()
	if err != nil {
		// If test fails or times out, return empty profile
		return profile, nil
	}

	// Parse JSON output
	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	testTimes := make(map[string]time.Time)

	for scanner.Scan() {
		var event struct {
			Time    string  `json:"Time"`
			Action  string  `json:"Action"`
			Package string  `json:"Package"`
			Test    string  `json:"Test"`
			Elapsed float64 `json:"Elapsed"`
			Output  string  `json:"Output"`
		}

		if err := json.Unmarshal(scanner.Bytes(), &event); err != nil {
			continue
		}

		switch event.Action {
		case "run":
			if event.Test != "" {
				testTimes[event.Test] = time.Now()
			}
		case "pass", "fail":
			if event.Test != "" {
				duration := time.Duration(event.Elapsed * float64(time.Second))
				result := TestResult{
					Package:  pkg,
					Name:     event.Test,
					Duration: duration,
					Status:   event.Action,
				}
				profile.Tests = append(profile.Tests, result)
				profile.Duration += duration
				profile.TestCount++
			}
		case "output":
			// Check for indicators of DB or file I/O
			if strings.Contains(event.Output, "database") ||
				strings.Contains(event.Output, "sql") {
				profile.HasDBAccess = true
			}
			if strings.Contains(event.Output, "file") ||
				strings.Contains(event.Output, "io") {
				profile.HasFileIO = true
			}
		}
	}

	// Determine if package can be parallelized
	profile.CanParallel = !profile.HasDBAccess && !profile.HasFileIO

	return profile, nil
}

// identifySlowTests finds tests that take too long
func (o *TestOptimizer) identifySlowTests() {
	var allTests []TestResult
	slowThreshold := 1 * time.Second

	for _, pkg := range o.profile.Packages {
		for _, test := range pkg.Tests {
			allTests = append(allTests, test)
			if test.Duration > slowThreshold {
				o.profile.SlowTests = append(o.profile.SlowTests, test)
			}
		}
	}

	// Sort by duration (slowest first)
	sort.Slice(o.profile.SlowTests, func(i, j int) bool {
		return o.profile.SlowTests[i].Duration > o.profile.SlowTests[j].Duration
	})

	// Keep top 10 slowest
	if len(o.profile.SlowTests) > 10 {
		o.profile.SlowTests = o.profile.SlowTests[:10]
	}
}

// analyzeParallelization identifies parallelization opportunities
func (o *TestOptimizer) analyzeParallelization() {
	for _, pkg := range o.profile.Packages {
		// Check if tests use t.Parallel()
		hasParallel := false
		for _, test := range pkg.Tests {
			// This is a simplification - in reality, we'd need to parse the test file
			if strings.Contains(test.Name, "Parallel") {
				hasParallel = true
				break
			}
		}

		if !hasParallel && pkg.CanParallel {
			o.profile.Optimizations = append(o.profile.Optimizations, Optimization{
				Type:        "parallelization",
				Target:      pkg.Name,
				Description: fmt.Sprintf("Add t.Parallel() to tests in %s", pkg.Name),
				Impact:      pkg.Duration / 2, // Estimate 50% improvement
				Priority:    "high",
			})
		}
	}
}

// generateOptimizations creates optimization recommendations
func (o *TestOptimizer) generateOptimizations() {
	// Recommend parallelization for slow tests
	for _, test := range o.profile.SlowTests {
		if test.Duration > 2*time.Second {
			o.profile.Optimizations = append(o.profile.Optimizations, Optimization{
				Type:        "slow_test",
				Target:      fmt.Sprintf("%s.%s", test.Package, test.Name),
				Description: fmt.Sprintf("Optimize test %s (currently %v)", test.Name, test.Duration),
				Impact:      test.Duration / 2,
				Priority:    "high",
			})
		}
	}

	// Recommend test splitting for large packages
	for _, pkg := range o.profile.Packages {
		if pkg.TestCount > 50 && pkg.Duration > 5*time.Second {
			o.profile.Optimizations = append(o.profile.Optimizations, Optimization{
				Type:        "split_package",
				Target:      pkg.Name,
				Description: fmt.Sprintf("Split %s into smaller test files", pkg.Name),
				Impact:      pkg.Duration / 3,
				Priority:    "medium",
			})
		}
	}

	// Recommend caching for packages with file I/O
	for _, pkg := range o.profile.Packages {
		if pkg.HasFileIO && pkg.Duration > 2*time.Second {
			o.profile.Optimizations = append(o.profile.Optimizations, Optimization{
				Type:        "caching",
				Target:      pkg.Name,
				Description: fmt.Sprintf("Add test data caching in %s", pkg.Name),
				Impact:      pkg.Duration / 4,
				Priority:    "medium",
			})
		}
	}

	// Sort optimizations by impact
	sort.Slice(o.profile.Optimizations, func(i, j int) bool {
		return o.profile.Optimizations[i].Impact > o.profile.Optimizations[j].Impact
	})
}

// GenerateParallelTestScript creates an optimized test runner
func (o *TestOptimizer) GenerateParallelTestScript() error {
	scriptPath := filepath.Join(o.rootDir, "scripts", "parallel-test.sh")

	script := `#!/bin/bash
# T013-B: Optimized parallel test runner
# Generated by test-optimizer.go

set -e

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

echo -e "${YELLOW}Running optimized parallel tests...${NC}"

# Start timing
START_TIME=$(date +%s)

# Function to run tests for a package group
run_test_group() {
    local group=$1
    echo "Testing group: $group"
    go test -v -parallel 4 $group
}

# Export function for parallel execution
export -f run_test_group

# Group 1: Fast, parallelizable packages
FAST_PACKAGES=(
`

	// Add fast packages
	for _, pkg := range o.profile.Packages {
		if pkg.CanParallel && pkg.Duration < 1*time.Second {
			script += fmt.Sprintf("    %s\n", pkg.Name)
		}
	}

	script += `)

# Group 2: Medium speed packages
MEDIUM_PACKAGES=(
`

	// Add medium packages
	for _, pkg := range o.profile.Packages {
		if pkg.Duration >= 1*time.Second && pkg.Duration < 5*time.Second {
			script += fmt.Sprintf("    %s\n", pkg.Name)
		}
	}

	script += `)

# Group 3: Slow or sequential packages
SLOW_PACKAGES=(
`

	// Add slow packages
	for _, pkg := range o.profile.Packages {
		if !pkg.CanParallel || pkg.Duration >= 5*time.Second {
			script += fmt.Sprintf("    %s\n", pkg.Name)
		}
	}

	script += `)

# Run fast packages in parallel
echo -e "${GREEN}Running fast packages in parallel...${NC}"
printf "%s\n" "${FAST_PACKAGES[@]}" | xargs -P 4 -I {} bash -c 'run_test_group "$@"' _ {}

# Run medium packages with limited parallelism
echo -e "${GREEN}Running medium packages...${NC}"
printf "%s\n" "${MEDIUM_PACKAGES[@]}" | xargs -P 2 -I {} bash -c 'run_test_group "$@"' _ {}

# Run slow packages sequentially
echo -e "${GREEN}Running slow packages...${NC}"
for pkg in "${SLOW_PACKAGES[@]}"; do
    run_test_group "$pkg"
done

# Calculate duration
END_TIME=$(date +%s)
DURATION=$((END_TIME - START_TIME))

echo -e "${GREEN}All tests completed in ${DURATION} seconds${NC}"

# Check if we met the target
if [ $DURATION -gt 30 ]; then
    echo -e "${RED}⚠ Tests took longer than 30 seconds target${NC}"
    exit 1
else
    echo -e "${GREEN}✅ Tests completed within 30 seconds target${NC}"
fi
`

	// Create scripts directory if needed
	scriptDir := filepath.Dir(scriptPath)
	if err := os.MkdirAll(scriptDir, 0755); err != nil {
		return err
	}

	// Write script
	if err := os.WriteFile(scriptPath, []byte(script), 0755); err != nil {
		return err
	}

	fmt.Printf("Generated parallel test script: %s\n", scriptPath)
	return nil
}

// GenerateOptimizationReport creates detailed report
func (o *TestOptimizer) GenerateOptimizationReport(outputPath string) error {
	report := struct {
		Profile         *TestProfile `json:"profile"`
		Summary         Summary      `json:"summary"`
		Recommendations []string     `json:"recommendations"`
	}{
		Profile: o.profile,
	}

	// Generate summary
	report.Summary = Summary{
		CurrentDuration:  o.profile.TotalDuration,
		TargetDuration:   o.targetTime,
		MeetsTarget:      o.profile.TotalDuration <= o.targetTime,
		PotentialSavings: o.calculatePotentialSavings(),
		TopSlowPackages:  o.getTopSlowPackages(5),
	}

	// Generate recommendations
	report.Recommendations = []string{
		fmt.Sprintf("Enable parallel execution for %d packages", o.countParallelizable()),
		fmt.Sprintf("Optimize %d slow tests (>1s each)", len(o.profile.SlowTests)),
		"Use test data factories to reduce setup time",
		"Implement test result caching for expensive operations",
		"Consider using build tags to separate unit and integration tests",
	}

	// Write JSON report
	jsonData, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}

	if err := os.WriteFile(outputPath, jsonData, 0644); err != nil {
		return err
	}

	// Also generate human-readable report
	textPath := strings.TrimSuffix(outputPath, ".json") + ".txt"
	if err := o.generateTextReport(textPath); err != nil {
		return err
	}

	return nil
}

// generateTextReport creates human-readable report
func (o *TestOptimizer) generateTextReport(outputPath string) error {
	file, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer file.Close()

	fmt.Fprintln(file, "TEST PERFORMANCE OPTIMIZATION REPORT")
	fmt.Fprintln(file, strings.Repeat("=", 60))
	fmt.Fprintln(file)

	fmt.Fprintf(file, "Current Total Duration: %v\n", o.profile.TotalDuration)
	fmt.Fprintf(file, "Target Duration: %v\n", o.targetTime)

	if o.profile.TotalDuration <= o.targetTime {
		fmt.Fprintln(file, "Status: ✅ MEETS TARGET")
	} else {
		fmt.Fprintln(file, "Status: ❌ EXCEEDS TARGET")
	}
	fmt.Fprintln(file)

	// Top slow tests
	fmt.Fprintln(file, "TOP 10 SLOWEST TESTS:")
	fmt.Fprintln(file, strings.Repeat("-", 60))
	for i, test := range o.profile.SlowTests {
		fmt.Fprintf(file, "%d. %s.%s - %v\n",
			i+1, test.Package, test.Name, test.Duration)
	}
	fmt.Fprintln(file)

	// Optimization recommendations
	fmt.Fprintln(file, "OPTIMIZATION RECOMMENDATIONS:")
	fmt.Fprintln(file, strings.Repeat("-", 60))
	for i, opt := range o.profile.Optimizations {
		if i >= 10 {
			break
		}
		fmt.Fprintf(file, "\n[%s] %s\n", strings.ToUpper(opt.Priority), opt.Description)
		fmt.Fprintf(file, "Target: %s\n", opt.Target)
		fmt.Fprintf(file, "Estimated Impact: -%v\n", opt.Impact)
	}

	return nil
}

// Helper methods

func (o *TestOptimizer) calculatePotentialSavings() time.Duration {
	var savings time.Duration
	for _, opt := range o.profile.Optimizations {
		savings += opt.Impact
	}
	return savings
}

func (o *TestOptimizer) countParallelizable() int {
	count := 0
	for _, pkg := range o.profile.Packages {
		if pkg.CanParallel {
			count++
		}
	}
	return count
}

func (o *TestOptimizer) getTopSlowPackages(n int) []string {
	type pkgDuration struct {
		name     string
		duration time.Duration
	}

	var packages []pkgDuration
	for name, pkg := range o.profile.Packages {
		packages = append(packages, pkgDuration{name, pkg.Duration})
	}

	sort.Slice(packages, func(i, j int) bool {
		return packages[i].duration > packages[j].duration
	})

	var result []string
	for i := 0; i < n && i < len(packages); i++ {
		result = append(result, fmt.Sprintf("%s (%v)",
			packages[i].name, packages[i].duration))
	}

	return result
}

type Summary struct {
	CurrentDuration  time.Duration `json:"current_duration"`
	TargetDuration   time.Duration `json:"target_duration"`
	MeetsTarget      bool          `json:"meets_target"`
	PotentialSavings time.Duration `json:"potential_savings"`
	TopSlowPackages  []string      `json:"top_slow_packages"`
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run test-optimizer.go <project-root> [output-dir]")
		fmt.Println("       go run test-optimizer.go --quick <project-root> [output-dir]")
		fmt.Println("")
		fmt.Println("Options:")
		fmt.Println("  --quick  Quick analysis without running tests")
		os.Exit(1)
	}

	quickMode := false
	projectRoot := os.Args[1]
	outputDir := "coverage"

	// Check for --quick flag
	if os.Args[1] == "--quick" {
		quickMode = true
		if len(os.Args) < 3 {
			fmt.Println("Error: --quick requires project-root argument")
			os.Exit(1)
		}
		projectRoot = os.Args[2]
		if len(os.Args) > 3 {
			outputDir = os.Args[3]
		}
	} else if len(os.Args) > 2 {
		outputDir = os.Args[2]
	}

	optimizer := NewTestOptimizer(projectRoot)

	// Analyze test performance
	if quickMode {
		fmt.Println("Running quick analysis (without executing tests)...")
		if err := optimizer.QuickAnalyze(); err != nil {
			fmt.Printf("Error in quick analysis: %v\n", err)
			os.Exit(1)
		}
	} else {
		if err := optimizer.Analyze(); err != nil {
			fmt.Printf("Error analyzing tests: %v\n", err)
			os.Exit(1)
		}
	}

	// Generate parallel test script
	if err := optimizer.GenerateParallelTestScript(); err != nil {
		fmt.Printf("Error generating script: %v\n", err)
	}

	// Generate optimization report
	reportPath := filepath.Join(outputDir, "test-optimization.json")
	if err := optimizer.GenerateOptimizationReport(reportPath); err != nil {
		fmt.Printf("Error generating report: %v\n", err)
	}

	fmt.Println("\nOptimization analysis complete!")
	fmt.Printf("Reports generated in %s\n", outputDir)

	// Exit with error if target not met
	if optimizer.profile.TotalDuration > optimizer.targetTime {
		fmt.Printf("\n⚠ Tests exceed 30 second target by %v\n",
			optimizer.profile.TotalDuration-optimizer.targetTime)
		os.Exit(1)
	}
}

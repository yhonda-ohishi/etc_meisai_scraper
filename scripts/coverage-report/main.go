// T012-D: Detailed coverage reporting with uncovered line identification
// This tool generates comprehensive coverage reports showing exactly which lines are uncovered

package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// UncoveredReport represents the detailed uncovered lines report
type UncoveredReport struct {
	Summary       Summary                   `json:"summary"`
	Packages      map[string]*PackageReport `json:"packages"`
	TopUncovered  []FileInfo                `json:"top_uncovered"`
	CriticalPaths []CriticalPath            `json:"critical_paths"`
}

// Summary provides overall coverage statistics
type Summary struct {
	TotalLines     int     `json:"total_lines"`
	CoveredLines   int     `json:"covered_lines"`
	UncoveredLines int     `json:"uncovered_lines"`
	Coverage       float64 `json:"coverage"`
	PackageCount   int     `json:"package_count"`
	FileCount      int     `json:"file_count"`
}

// PackageReport contains package-level coverage details
type PackageReport struct {
	Name           string                 `json:"name"`
	Coverage       float64                `json:"coverage"`
	Files          map[string]*FileReport `json:"files"`
	UncoveredLines int                    `json:"uncovered_lines"`
}

// FileReport contains file-level coverage details
type FileReport struct {
	Name           string         `json:"name"`
	Coverage       float64        `json:"coverage"`
	TotalLines     int            `json:"total_lines"`
	CoveredLines   int            `json:"covered_lines"`
	UncoveredLines []LineDetail   `json:"uncovered_lines"`
	Functions      []FunctionInfo `json:"functions"`
}

// LineDetail provides information about an uncovered line
type LineDetail struct {
	Number  int    `json:"number"`
	Content string `json:"content"`
	Type    string `json:"type"` // "code", "branch", "error-handling"
	Context string `json:"context,omitempty"`
}

// FunctionInfo provides function-level coverage
type FunctionInfo struct {
	Name           string  `json:"name"`
	StartLine      int     `json:"start_line"`
	EndLine        int     `json:"end_line"`
	Coverage       float64 `json:"coverage"`
	UncoveredLines []int   `json:"uncovered_lines"`
}

// FileInfo for ranking files by uncovered lines
type FileInfo struct {
	Path           string  `json:"path"`
	UncoveredCount int     `json:"uncovered_count"`
	Coverage       float64 `json:"coverage"`
}

// CriticalPath identifies important code paths without coverage
type CriticalPath struct {
	Description string   `json:"description"`
	Files       []string `json:"files"`
	Priority    string   `json:"priority"` // "high", "medium", "low"
}

// CoverageReporter generates detailed coverage reports
type CoverageReporter struct {
	coverageFile string
	sourceDir    string
	outputDir    string
}

// NewCoverageReporter creates a new reporter instance
func NewCoverageReporter(coverageFile, sourceDir, outputDir string) *CoverageReporter {
	return &CoverageReporter{
		coverageFile: coverageFile,
		sourceDir:    sourceDir,
		outputDir:    outputDir,
	}
}

// GenerateReport creates the detailed coverage report
func (r *CoverageReporter) GenerateReport() error {
	// Parse coverage data
	coverageData, err := r.parseCoverageFile()
	if err != nil {
		return fmt.Errorf("failed to parse coverage file: %w", err)
	}

	// Analyze uncovered lines
	report, err := r.analyzeUncoveredLines(coverageData)
	if err != nil {
		return fmt.Errorf("failed to analyze uncovered lines: %w", err)
	}

	// Identify critical paths
	report.CriticalPaths = r.identifyCriticalPaths(report)

	// Generate output files
	if err := r.generateOutputFiles(report); err != nil {
		return fmt.Errorf("failed to generate output files: %w", err)
	}

	return nil
}

// parseCoverageFile parses the Go coverage profile
func (r *CoverageReporter) parseCoverageFile() (map[string][][]int, error) {
	file, err := os.Open(r.coverageFile)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	coverage := make(map[string][][]int)
	scanner := bufio.NewScanner(file)

	// Skip mode line
	if scanner.Scan() {
		// First line is mode
	}

	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		parts := strings.Fields(line)
		if len(parts) < 3 {
			continue
		}

		// Parse filename:start.col,end.col
		fileAndRange := strings.Split(parts[0], ":")
		if len(fileAndRange) != 2 {
			continue
		}

		filename := fileAndRange[0]
		ranges := strings.Split(fileAndRange[1], ",")
		if len(ranges) != 2 {
			continue
		}

		// Parse positions
		start := strings.Split(ranges[0], ".")
		end := strings.Split(ranges[1], ".")
		if len(start) != 2 || len(end) != 2 {
			continue
		}

		var startLine, endLine, count int
		fmt.Sscanf(start[0], "%d", &startLine)
		fmt.Sscanf(end[0], "%d", &endLine)
		fmt.Sscanf(parts[2], "%d", &count)

		if coverage[filename] == nil {
			coverage[filename] = make([][]int, 0)
		}
		coverage[filename] = append(coverage[filename], []int{startLine, endLine, count})
	}

	return coverage, scanner.Err()
}

// analyzeUncoveredLines analyzes which lines are not covered
func (r *CoverageReporter) analyzeUncoveredLines(coverageData map[string][][]int) (*UncoveredReport, error) {
	report := &UncoveredReport{
		Packages:     make(map[string]*PackageReport),
		TopUncovered: make([]FileInfo, 0),
	}

	for filename, blocks := range coverageData {
		// Skip excluded files
		if r.shouldExclude(filename) {
			continue
		}

		fileReport, err := r.analyzeFile(filename, blocks)
		if err != nil {
			fmt.Printf("Warning: skipping %s: %v\n", filename, err)
			continue
		}

		// Add to package report
		pkg := filepath.Dir(filename)
		if report.Packages[pkg] == nil {
			report.Packages[pkg] = &PackageReport{
				Name:  pkg,
				Files: make(map[string]*FileReport),
			}
		}
		report.Packages[pkg].Files[filename] = fileReport

		// Update summary
		report.Summary.TotalLines += fileReport.TotalLines
		report.Summary.CoveredLines += fileReport.CoveredLines
		report.Summary.FileCount++

		// Track top uncovered files
		if len(fileReport.UncoveredLines) > 0 {
			report.TopUncovered = append(report.TopUncovered, FileInfo{
				Path:           filename,
				UncoveredCount: len(fileReport.UncoveredLines),
				Coverage:       fileReport.Coverage,
			})
		}
	}

	// Calculate summary stats
	report.Summary.UncoveredLines = report.Summary.TotalLines - report.Summary.CoveredLines
	if report.Summary.TotalLines > 0 {
		report.Summary.Coverage = float64(report.Summary.CoveredLines) / float64(report.Summary.TotalLines) * 100
	}
	report.Summary.PackageCount = len(report.Packages)

	// Sort top uncovered files
	sort.Slice(report.TopUncovered, func(i, j int) bool {
		return report.TopUncovered[i].UncoveredCount > report.TopUncovered[j].UncoveredCount
	})
	if len(report.TopUncovered) > 10 {
		report.TopUncovered = report.TopUncovered[:10]
	}

	// Calculate package coverage
	for _, pkg := range report.Packages {
		var totalLines, coveredLines int
		for _, file := range pkg.Files {
			totalLines += file.TotalLines
			coveredLines += file.CoveredLines
			pkg.UncoveredLines += len(file.UncoveredLines)
		}
		if totalLines > 0 {
			pkg.Coverage = float64(coveredLines) / float64(totalLines) * 100
		}
	}

	return report, nil
}

// analyzeFile analyzes coverage for a single file
func (r *CoverageReporter) analyzeFile(filename string, blocks [][]int) (*FileReport, error) {
	report := &FileReport{
		Name:           filename,
		UncoveredLines: make([]LineDetail, 0),
		Functions:      make([]FunctionInfo, 0),
	}

	// Convert module path to relative path if necessary
	relativePath := filename
	if idx := strings.Index(filename, "etc_meisai/"); idx >= 0 {
		relativePath = filename[idx+len("etc_meisai/"):]
	}

	// Read source file
	sourcePath := filepath.Join(r.sourceDir, relativePath)
	sourceContent, err := os.ReadFile(sourcePath)
	if err != nil {
		return report, err
	}

	lines := strings.Split(string(sourceContent), "\n")
	report.TotalLines = len(lines)

	// Create coverage map
	lineCoverage := make(map[int]bool)
	for _, block := range blocks {
		startLine, endLine, count := block[0], block[1], block[2]
		if count > 0 {
			for line := startLine; line <= endLine; line++ {
				lineCoverage[line] = true
			}
		}
	}

	// Find uncovered lines
	for i, line := range lines {
		lineNum := i + 1
		trimmed := strings.TrimSpace(line)

		// Skip empty lines and comments
		if trimmed == "" || strings.HasPrefix(trimmed, "//") {
			continue
		}

		// Skip package and import statements
		if strings.HasPrefix(trimmed, "package ") || strings.HasPrefix(trimmed, "import ") {
			continue
		}

		if !lineCoverage[lineNum] {
			lineType := r.classifyLine(trimmed)
			report.UncoveredLines = append(report.UncoveredLines, LineDetail{
				Number:  lineNum,
				Content: strings.TrimSpace(line),
				Type:    lineType,
				Context: r.getLineContext(lines, i),
			})
		} else {
			report.CoveredLines++
		}
	}

	// Calculate coverage
	if report.TotalLines > 0 {
		report.Coverage = float64(report.CoveredLines) / float64(report.TotalLines) * 100
	}

	// Analyze functions (simplified)
	report.Functions = r.analyzeFunctions(lines, lineCoverage)

	return report, nil
}

// classifyLine determines the type of code line
func (r *CoverageReporter) classifyLine(line string) string {
	line = strings.TrimSpace(line)

	if strings.Contains(line, "if err") || strings.Contains(line, "return err") {
		return "error-handling"
	}
	if strings.HasPrefix(line, "if ") || strings.HasPrefix(line, "switch ") ||
		strings.HasPrefix(line, "for ") || strings.HasPrefix(line, "case ") {
		return "branch"
	}
	return "code"
}

// getLineContext provides surrounding context for an uncovered line
func (r *CoverageReporter) getLineContext(lines []string, index int) string {
	// Find the containing function or method
	for i := index; i >= 0; i-- {
		line := strings.TrimSpace(lines[i])
		if strings.HasPrefix(line, "func ") {
			// Extract function name
			parts := strings.Fields(line)
			if len(parts) > 1 {
				return "in function: " + parts[1]
			}
			break
		}
	}
	return ""
}

// analyzeFunctions identifies functions and their coverage
func (r *CoverageReporter) analyzeFunctions(lines []string, lineCoverage map[int]bool) []FunctionInfo {
	functions := make([]FunctionInfo, 0)

	for i, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), "func ") {
			funcInfo := FunctionInfo{
				Name:           r.extractFunctionName(line),
				StartLine:      i + 1,
				UncoveredLines: make([]int, 0),
			}

			// Find function end
			braceCount := 0
			for j := i; j < len(lines); j++ {
				for _, char := range lines[j] {
					if char == '{' {
						braceCount++
					} else if char == '}' {
						braceCount--
						if braceCount == 0 {
							funcInfo.EndLine = j + 1
							break
						}
					}
				}
				if funcInfo.EndLine > 0 {
					break
				}
			}

			// Calculate function coverage
			totalLines := 0
			coveredLines := 0
			for line := funcInfo.StartLine; line <= funcInfo.EndLine; line++ {
				if line-1 < len(lines) {
					trimmed := strings.TrimSpace(lines[line-1])
					if trimmed != "" && !strings.HasPrefix(trimmed, "//") {
						totalLines++
						if lineCoverage[line] {
							coveredLines++
						} else {
							funcInfo.UncoveredLines = append(funcInfo.UncoveredLines, line)
						}
					}
				}
			}

			if totalLines > 0 {
				funcInfo.Coverage = float64(coveredLines) / float64(totalLines) * 100
			}

			functions = append(functions, funcInfo)
		}
	}

	return functions
}

// extractFunctionName extracts the function name from a func declaration
func (r *CoverageReporter) extractFunctionName(line string) string {
	line = strings.TrimSpace(line)
	if !strings.HasPrefix(line, "func ") {
		return ""
	}

	line = line[5:] // Remove "func "

	// Handle method receivers
	if strings.HasPrefix(line, "(") {
		end := strings.Index(line, ")")
		if end > 0 {
			line = line[end+1:]
		}
	}

	// Extract function name
	parts := strings.Fields(line)
	if len(parts) > 0 {
		name := parts[0]
		if idx := strings.Index(name, "("); idx > 0 {
			name = name[:idx]
		}
		return name
	}

	return ""
}

// identifyCriticalPaths finds important uncovered code paths
func (r *CoverageReporter) identifyCriticalPaths(report *UncoveredReport) []CriticalPath {
	paths := make([]CriticalPath, 0)

	// Check for uncovered error handling
	errorFiles := make([]string, 0)
	for _, pkg := range report.Packages {
		for filename, file := range pkg.Files {
			for _, line := range file.UncoveredLines {
				if line.Type == "error-handling" {
					errorFiles = append(errorFiles, filename)
					break
				}
			}
		}
	}
	if len(errorFiles) > 0 {
		paths = append(paths, CriticalPath{
			Description: "Uncovered error handling paths",
			Files:       errorFiles,
			Priority:    "high",
		})
	}

	// Check for uncovered main functions
	mainFiles := make([]string, 0)
	for _, pkg := range report.Packages {
		for filename, file := range pkg.Files {
			for _, fn := range file.Functions {
				if fn.Name == "main" && fn.Coverage < 100 {
					mainFiles = append(mainFiles, filename)
				}
			}
		}
	}
	if len(mainFiles) > 0 {
		paths = append(paths, CriticalPath{
			Description: "Uncovered main function code",
			Files:       mainFiles,
			Priority:    "medium",
		})
	}

	// Check for completely uncovered files
	uncoveredFiles := make([]string, 0)
	for _, pkg := range report.Packages {
		for filename, file := range pkg.Files {
			if file.Coverage == 0 {
				uncoveredFiles = append(uncoveredFiles, filename)
			}
		}
	}
	if len(uncoveredFiles) > 0 {
		paths = append(paths, CriticalPath{
			Description: "Completely uncovered files",
			Files:       uncoveredFiles,
			Priority:    "high",
		})
	}

	return paths
}

// shouldExclude checks if a file should be excluded from coverage
func (r *CoverageReporter) shouldExclude(filename string) bool {
	excludePatterns := []string{
		".pb.go",
		".pb.gw.go",
		"_mock.go",
		"/mocks/",
		"/vendor/",
		"/migrations/",
		"_test.go",
	}

	for _, pattern := range excludePatterns {
		if strings.Contains(filename, pattern) {
			return true
		}
	}
	return false
}

// generateOutputFiles creates all output report files
func (r *CoverageReporter) generateOutputFiles(report *UncoveredReport) error {
	// Ensure output directory exists
	if err := os.MkdirAll(r.outputDir, 0755); err != nil {
		return err
	}

	// Generate JSON report
	if err := r.generateJSONReport(report); err != nil {
		return err
	}

	// Generate text report
	if err := r.generateTextReport(report); err != nil {
		return err
	}

	// Generate HTML report
	if err := r.generateHTMLReport(report); err != nil {
		return err
	}

	// Generate markdown report
	if err := r.generateMarkdownReport(report); err != nil {
		return err
	}

	return nil
}

// generateJSONReport creates a JSON format report
func (r *CoverageReporter) generateJSONReport(report *UncoveredReport) error {
	file, err := os.Create(filepath.Join(r.outputDir, "uncovered-lines.json"))
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(report)
}

// generateTextReport creates a human-readable text report
func (r *CoverageReporter) generateTextReport(report *UncoveredReport) error {
	file, err := os.Create(filepath.Join(r.outputDir, "uncovered-lines.txt"))
	if err != nil {
		return err
	}
	defer file.Close()

	w := bufio.NewWriter(file)
	defer w.Flush()

	// Header
	fmt.Fprintln(w, strings.Repeat("=", 80))
	fmt.Fprintln(w, "UNCOVERED LINES REPORT")
	fmt.Fprintln(w, strings.Repeat("=", 80))
	fmt.Fprintln(w)

	// Summary
	fmt.Fprintf(w, "Overall Coverage: %.2f%% (%d/%d lines)\n",
		report.Summary.Coverage,
		report.Summary.CoveredLines,
		report.Summary.TotalLines)
	fmt.Fprintf(w, "Uncovered Lines: %d\n", report.Summary.UncoveredLines)
	fmt.Fprintf(w, "Files Analyzed: %d\n", report.Summary.FileCount)
	fmt.Fprintf(w, "Packages: %d\n", report.Summary.PackageCount)
	fmt.Fprintln(w)

	// Top uncovered files
	if len(report.TopUncovered) > 0 {
		fmt.Fprintln(w, strings.Repeat("-", 80))
		fmt.Fprintln(w, "TOP FILES WITH UNCOVERED LINES:")
		fmt.Fprintln(w, strings.Repeat("-", 80))
		for _, file := range report.TopUncovered {
			fmt.Fprintf(w, "  %s: %d uncovered lines (%.1f%% coverage)\n",
				file.Path, file.UncoveredCount, file.Coverage)
		}
		fmt.Fprintln(w)
	}

	// Critical paths
	if len(report.CriticalPaths) > 0 {
		fmt.Fprintln(w, strings.Repeat("-", 80))
		fmt.Fprintln(w, "CRITICAL UNCOVERED PATHS:")
		fmt.Fprintln(w, strings.Repeat("-", 80))
		for _, path := range report.CriticalPaths {
			fmt.Fprintf(w, "\n[%s] %s\n", strings.ToUpper(path.Priority), path.Description)
			for _, file := range path.Files {
				fmt.Fprintf(w, "  - %s\n", file)
			}
		}
		fmt.Fprintln(w)
	}

	// Package details
	fmt.Fprintln(w, strings.Repeat("-", 80))
	fmt.Fprintln(w, "PACKAGE DETAILS:")
	fmt.Fprintln(w, strings.Repeat("-", 80))

	packages := make([]string, 0, len(report.Packages))
	for pkg := range report.Packages {
		packages = append(packages, pkg)
	}
	sort.Strings(packages)

	for _, pkgName := range packages {
		pkg := report.Packages[pkgName]
		fmt.Fprintf(w, "\n%s (%.1f%% coverage, %d uncovered lines)\n",
			pkgName, pkg.Coverage, pkg.UncoveredLines)

		files := make([]string, 0, len(pkg.Files))
		for file := range pkg.Files {
			files = append(files, file)
		}
		sort.Strings(files)

		for _, filename := range files {
			file := pkg.Files[filename]
			if len(file.UncoveredLines) > 0 {
				fmt.Fprintf(w, "\n  %s:\n", filepath.Base(filename))

				// Show first 5 uncovered lines
				shown := 0
				for _, line := range file.UncoveredLines {
					if shown >= 5 {
						fmt.Fprintf(w, "    ... and %d more uncovered lines\n",
							len(file.UncoveredLines)-shown)
						break
					}
					fmt.Fprintf(w, "    Line %d [%s]: %s\n",
						line.Number, line.Type, truncate(line.Content, 60))
					shown++
				}
			}
		}
	}

	return nil
}

// generateHTMLReport creates an interactive HTML report
func (r *CoverageReporter) generateHTMLReport(report *UncoveredReport) error {
	file, err := os.Create(filepath.Join(r.outputDir, "uncovered-lines.html"))
	if err != nil {
		return err
	}
	defer file.Close()

	tmplText := `<!DOCTYPE html>
<html>
<head>
    <title>Uncovered Lines Report</title>
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif;
            margin: 0;
            padding: 20px;
            background: #f5f5f5;
        }
        .container {
            max-width: 1200px;
            margin: 0 auto;
            background: white;
            padding: 20px;
            border-radius: 8px;
            box-shadow: 0 2px 4px rgba(0,0,0,0.1);
        }
        h1 {
            color: #333;
            border-bottom: 2px solid #007bff;
            padding-bottom: 10px;
        }
        h2 {
            color: #555;
            margin-top: 30px;
        }
        .summary {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
            gap: 20px;
            margin: 20px 0;
        }
        .summary-card {
            background: #f8f9fa;
            padding: 15px;
            border-radius: 5px;
            border-left: 4px solid #007bff;
        }
        .summary-card h3 {
            margin: 0 0 10px 0;
            color: #666;
            font-size: 14px;
            text-transform: uppercase;
        }
        .summary-card .value {
            font-size: 24px;
            font-weight: bold;
            color: #333;
        }
        .coverage-bar {
            width: 100%;
            height: 30px;
            background: #e0e0e0;
            border-radius: 5px;
            overflow: hidden;
            margin: 10px 0;
        }
        .coverage-fill {
            height: 100%;
            background: linear-gradient(90deg, #28a745, #20c997);
            display: flex;
            align-items: center;
            justify-content: center;
            color: white;
            font-weight: bold;
        }
        .critical-path {
            background: #fff3cd;
            border: 1px solid #ffc107;
            border-radius: 5px;
            padding: 15px;
            margin: 10px 0;
        }
        .critical-path.high {
            background: #f8d7da;
            border-color: #dc3545;
        }
        .file-list {
            list-style: none;
            padding: 0;
        }
        .file-list li {
            padding: 10px;
            margin: 5px 0;
            background: #f8f9fa;
            border-radius: 3px;
            display: flex;
            justify-content: space-between;
        }
        .uncovered-line {
            background: #fff5f5;
            border-left: 3px solid #dc3545;
            padding: 5px 10px;
            margin: 5px 0;
            font-family: monospace;
            font-size: 12px;
        }
        .line-number {
            color: #666;
            margin-right: 10px;
        }
        .line-type {
            display: inline-block;
            padding: 2px 6px;
            border-radius: 3px;
            font-size: 10px;
            text-transform: uppercase;
            margin-left: 10px;
        }
        .line-type.error-handling {
            background: #dc3545;
            color: white;
        }
        .line-type.branch {
            background: #ffc107;
            color: #333;
        }
        .line-type.code {
            background: #6c757d;
            color: white;
        }
        .collapsible {
            cursor: pointer;
            padding: 10px;
            background: #f8f9fa;
            border: none;
            text-align: left;
            outline: none;
            font-size: 16px;
            width: 100%;
            margin: 5px 0;
            border-radius: 5px;
        }
        .collapsible:hover {
            background: #e9ecef;
        }
        .collapsible::before {
            content: '▶';
            margin-right: 10px;
        }
        .collapsible.active::before {
            content: '▼';
        }
        .content {
            display: none;
            padding: 10px;
            background: white;
            border: 1px solid #dee2e6;
            border-radius: 0 0 5px 5px;
            margin-top: -5px;
        }
    </style>
</head>
<body>
    <div class="container">
        <h1>📊 Uncovered Lines Report</h1>

        <div class="summary">
            <div class="summary-card">
                <h3>Overall Coverage</h3>
                <div class="value">{{printf "%.1f" .Summary.Coverage}}%</div>
            </div>
            <div class="summary-card">
                <h3>Uncovered Lines</h3>
                <div class="value">{{.Summary.UncoveredLines}}</div>
            </div>
            <div class="summary-card">
                <h3>Total Files</h3>
                <div class="value">{{.Summary.FileCount}}</div>
            </div>
            <div class="summary-card">
                <h3>Total Packages</h3>
                <div class="value">{{.Summary.PackageCount}}</div>
            </div>
        </div>

        <div class="coverage-bar">
            <div class="coverage-fill" style="width: {{printf "%.1f" .Summary.Coverage}}%">
                {{printf "%.1f" .Summary.Coverage}}%
            </div>
        </div>

        {{if .CriticalPaths}}
        <h2>⚠️ Critical Uncovered Paths</h2>
        {{range .CriticalPaths}}
        <div class="critical-path {{.Priority}}">
            <strong>{{.Description}}</strong>
            <ul>
                {{range .Files}}
                <li>{{.}}</li>
                {{end}}
            </ul>
        </div>
        {{end}}
        {{end}}

        {{if .TopUncovered}}
        <h2>📁 Top Files with Uncovered Lines</h2>
        <ul class="file-list">
            {{range .TopUncovered}}
            <li>
                <span>{{.Path}}</span>
                <span>{{.UncoveredCount}} lines ({{printf "%.1f" .Coverage}}% covered)</span>
            </li>
            {{end}}
        </ul>
        {{end}}

        <h2>📦 Package Details</h2>
        {{range $pkgName, $pkg := .Packages}}
        <button class="collapsible">
            {{$pkgName}} - {{printf "%.1f" $pkg.Coverage}}% coverage, {{$pkg.UncoveredLines}} uncovered lines
        </button>
        <div class="content">
            {{range $filename, $file := $pkg.Files}}
            {{if $file.UncoveredLines}}
            <h4>{{$filename}}</h4>
            {{range $index, $line := $file.UncoveredLines}}
            {{if lt $index 10}}
            <div class="uncovered-line">
                <span class="line-number">Line {{$line.Number}}</span>
                {{$line.Content}}
                <span class="line-type {{$line.Type}}">{{$line.Type}}</span>
            </div>
            {{end}}
            {{end}}
            {{if gt (len $file.UncoveredLines) 10}}
            <p><em>...and {{minus (len $file.UncoveredLines) 10}} more uncovered lines</em></p>
            {{end}}
            {{end}}
            {{end}}
        </div>
        {{end}}
    </div>

    <script>
        var coll = document.getElementsByClassName("collapsible");
        for (var i = 0; i < coll.length; i++) {
            coll[i].addEventListener("click", function() {
                this.classList.toggle("active");
                var content = this.nextElementSibling;
                if (content.style.display === "block") {
                    content.style.display = "none";
                } else {
                    content.style.display = "block";
                }
            });
        }
    </script>
</body>
</html>`

	funcMap := template.FuncMap{
		"minus": func(a, b int) int { return a - b },
	}

	tmpl, err := template.New("report").Funcs(funcMap).Parse(tmplText)
	if err != nil {
		return err
	}

	return tmpl.Execute(file, report)
}

// generateMarkdownReport creates a markdown format report
func (r *CoverageReporter) generateMarkdownReport(report *UncoveredReport) error {
	file, err := os.Create(filepath.Join(r.outputDir, "uncovered-lines.md"))
	if err != nil {
		return err
	}
	defer file.Close()

	w := bufio.NewWriter(file)
	defer w.Flush()

	// Header
	fmt.Fprintln(w, "# Uncovered Lines Report")
	fmt.Fprintln(w)

	// Summary
	fmt.Fprintln(w, "## Summary")
	fmt.Fprintln(w)
	fmt.Fprintf(w, "- **Overall Coverage**: %.2f%% (%d/%d lines)\n",
		report.Summary.Coverage,
		report.Summary.CoveredLines,
		report.Summary.TotalLines)
	fmt.Fprintf(w, "- **Uncovered Lines**: %d\n", report.Summary.UncoveredLines)
	fmt.Fprintf(w, "- **Files Analyzed**: %d\n", report.Summary.FileCount)
	fmt.Fprintf(w, "- **Packages**: %d\n", report.Summary.PackageCount)
	fmt.Fprintln(w)

	// Critical paths
	if len(report.CriticalPaths) > 0 {
		fmt.Fprintln(w, "## ⚠️ Critical Uncovered Paths")
		fmt.Fprintln(w)
		for _, path := range report.CriticalPaths {
			fmt.Fprintf(w, "### %s (%s priority)\n\n", path.Description, path.Priority)
			for _, file := range path.Files {
				fmt.Fprintf(w, "- `%s`\n", file)
			}
			fmt.Fprintln(w)
		}
	}

	// Top uncovered files
	if len(report.TopUncovered) > 0 {
		fmt.Fprintln(w, "## Top Files with Uncovered Lines")
		fmt.Fprintln(w)
		fmt.Fprintln(w, "| File | Uncovered Lines | Coverage |")
		fmt.Fprintln(w, "|------|-----------------|----------|")
		for _, file := range report.TopUncovered {
			fmt.Fprintf(w, "| `%s` | %d | %.1f%% |\n",
				file.Path, file.UncoveredCount, file.Coverage)
		}
		fmt.Fprintln(w)
	}

	// Package details
	fmt.Fprintln(w, "## Package Details")
	fmt.Fprintln(w)

	packages := make([]string, 0, len(report.Packages))
	for pkg := range report.Packages {
		packages = append(packages, pkg)
	}
	sort.Strings(packages)

	for _, pkgName := range packages {
		pkg := report.Packages[pkgName]
		fmt.Fprintf(w, "### %s\n\n", pkgName)
		fmt.Fprintf(w, "- Coverage: %.1f%%\n", pkg.Coverage)
		fmt.Fprintf(w, "- Uncovered lines: %d\n\n", pkg.UncoveredLines)

		// Show files with uncovered lines
		hasUncovered := false
		for filename, file := range pkg.Files {
			if len(file.UncoveredLines) > 0 {
				if !hasUncovered {
					fmt.Fprintln(w, "#### Files with uncovered lines:")
					fmt.Fprintln(w)
					hasUncovered = true
				}
				fmt.Fprintf(w, "**`%s`** (%d uncovered lines)\n\n",
					filepath.Base(filename), len(file.UncoveredLines))

				// Show first few uncovered lines
				shown := 0
				for _, line := range file.UncoveredLines {
					if shown >= 3 {
						fmt.Fprintf(w, "- ...and %d more\n",
							len(file.UncoveredLines)-shown)
						break
					}
					fmt.Fprintf(w, "- Line %d (`%s`): `%s`\n",
						line.Number, line.Type, truncate(line.Content, 50))
					shown++
				}
				fmt.Fprintln(w)
			}
		}
		if !hasUncovered {
			fmt.Fprintln(w, "*All files have 100% coverage*")
			fmt.Fprintln(w)
		}
	}

	return nil
}

// truncate truncates a string to the specified length
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

func main() {
	if len(os.Args) < 4 {
		fmt.Println("Usage: go run scripts/coverage-report/main.go <coverage-file> <source-dir> <output-dir>")
		fmt.Println("   or: coverage-report <coverage-file> <source-dir> <output-dir>")
		os.Exit(1)
	}

	coverageFile := os.Args[1]
	sourceDir := os.Args[2]
	outputDir := os.Args[3]

	reporter := NewCoverageReporter(coverageFile, sourceDir, outputDir)

	if err := reporter.GenerateReport(); err != nil {
		fmt.Printf("Error generating report: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Coverage reports generated in %s\n", outputDir)
	fmt.Println("Files created:")
	fmt.Println("  - uncovered-lines.json")
	fmt.Println("  - uncovered-lines.txt")
	fmt.Println("  - uncovered-lines.html")
	fmt.Println("  - uncovered-lines.md")
}

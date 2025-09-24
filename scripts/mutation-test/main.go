// T013-A: Mutation testing to validate test effectiveness
// This tool introduces controlled mutations to code and verifies tests catch them

package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// MutationReport represents the results of mutation testing
type MutationReport struct {
	TotalMutants    int                    `json:"total_mutants"`
	KilledMutants   int                    `json:"killed_mutants"`
	SurvivedMutants int                    `json:"survived_mutants"`
	MutationScore   float64                `json:"mutation_score"`
	Duration        time.Duration          `json:"duration"`
	Files           map[string]*FileResult `json:"files"`
}

// FileResult contains mutation results for a single file
type FileResult struct {
	Path          string     `json:"path"`
	TotalMutants  int        `json:"total_mutants"`
	KilledMutants int        `json:"killed_mutants"`
	Survivors     []Mutation `json:"survivors"`
}

// Mutation represents a single code mutation
type Mutation struct {
	Type        string `json:"type"`
	Line        int    `json:"line"`
	Original    string `json:"original"`
	Mutated     string `json:"mutated"`
	Description string `json:"description"`
	Killed      bool   `json:"killed"`
	TestOutput  string `json:"test_output,omitempty"`
}

// MutationTester performs mutation testing
type MutationTester struct {
	sourceDir           string
	testCommand         string
	excludePaths        []string
	mutations           []MutationOperator
	maxMutationsPerFile int
	parallel            bool
}

// MutationOperator defines how to mutate code
type MutationOperator interface {
	Name() string
	Mutate(node ast.Node, fset *token.FileSet) []Mutation
}

// NewMutationTester creates a new mutation tester
func NewMutationTester(sourceDir string) *MutationTester {
	return &MutationTester{
		sourceDir:   sourceDir,
		testCommand: "go test ./...",
		excludePaths: []string{
			"_test.go",
			"pb.go",
			"mock.go",
			"/vendor/",
			"/mocks/",
		},
		mutations: []MutationOperator{
			&ConditionalMutator{},
			&ArithmeticMutator{},
			&LogicalMutator{},
			&ReturnMutator{},
			&ConstantMutator{},
		},
		maxMutationsPerFile: 0, // 0 means no limit
		parallel:            false,
	}
}

// SetQuickMode enables quick mode with limited mutations
func (mt *MutationTester) SetQuickMode() {
	mt.maxMutationsPerFile = 5
}

// SetParallelMode enables parallel execution
func (mt *MutationTester) SetParallelMode() {
	mt.parallel = true
}

// Run executes mutation testing
func (mt *MutationTester) Run() (*MutationReport, error) {
	report := &MutationReport{
		Files: make(map[string]*FileResult),
	}

	startTime := time.Now()

	// Find all Go source files
	files, err := mt.findSourceFiles()
	if err != nil {
		return nil, fmt.Errorf("failed to find source files: %w", err)
	}

	fmt.Printf("Found %d files to mutate\n", len(files))

	// Process each file
	for _, file := range files {
		if mt.shouldExclude(file) {
			continue
		}

		fileResult, err := mt.mutateFile(file)
		if err != nil {
			fmt.Printf("Warning: failed to mutate %s: %v\n", file, err)
			continue
		}

		report.Files[file] = fileResult
		report.TotalMutants += fileResult.TotalMutants
		report.KilledMutants += fileResult.KilledMutants
	}

	// Calculate final statistics
	report.Duration = time.Since(startTime)
	report.SurvivedMutants = report.TotalMutants - report.KilledMutants
	if report.TotalMutants > 0 {
		report.MutationScore = float64(report.KilledMutants) / float64(report.TotalMutants) * 100
	}

	return report, nil
}

// findSourceFiles locates all Go source files
func (mt *MutationTester) findSourceFiles() ([]string, error) {
	var files []string

	err := filepath.Walk(mt.sourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if strings.HasSuffix(path, ".go") && !strings.HasSuffix(path, "_test.go") {
			files = append(files, path)
		}

		return nil
	})

	return files, err
}

// shouldExclude checks if a file should be excluded
func (mt *MutationTester) shouldExclude(path string) bool {
	for _, exclude := range mt.excludePaths {
		if strings.Contains(path, exclude) {
			return true
		}
	}
	return false
}

// mutateFile applies mutations to a single file
func (mt *MutationTester) mutateFile(filename string) (*FileResult, error) {
	result := &FileResult{
		Path:      filename,
		Survivors: make([]Mutation, 0),
	}

	// Parse the file
	fset := token.NewFileSet()
	src, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	file, err := parser.ParseFile(fset, filename, src, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	// Generate mutations
	mutations := mt.generateMutations(file, fset)
	result.TotalMutants = len(mutations)

	fmt.Printf("Testing %d mutations in %s\n", len(mutations), filename)

	// Test each mutation
	for i, mutation := range mutations {
		fmt.Printf("  [%d/%d] Testing %s mutation at line %d... ",
			i+1, len(mutations), mutation.Type, mutation.Line)

		killed := mt.testMutation(filename, src, mutation)
		if killed {
			result.KilledMutants++
			fmt.Println("✅ Killed")
		} else {
			mutation.Killed = false
			result.Survivors = append(result.Survivors, mutation)
			fmt.Println("❌ Survived")
		}
	}

	return result, nil
}

// generateMutations creates all possible mutations for a file
func (mt *MutationTester) generateMutations(file *ast.File, fset *token.FileSet) []Mutation {
	var mutations []Mutation

	// Walk the AST and apply mutation operators
	ast.Inspect(file, func(node ast.Node) bool {
		if node == nil {
			return true
		}

		for _, operator := range mt.mutations {
			mutations = append(mutations, operator.Mutate(node, fset)...)
		}

		return true
	})

	return mutations
}

// testMutation applies a mutation and runs tests
func (mt *MutationTester) testMutation(filename string, original []byte, mutation Mutation) bool {
	// Apply mutation to source
	mutated := mt.applyMutation(original, mutation)

	// Save original file
	backup, err := os.ReadFile(filename)
	if err != nil {
		return false
	}

	// Write mutated file
	if err := os.WriteFile(filename, mutated, 0644); err != nil {
		return false
	}

	// Run tests
	cmd := exec.Command("sh", "-c", mt.testCommand)
	output, err := cmd.CombinedOutput()

	// Restore original file
	os.WriteFile(filename, backup, 0644)

	// Tests failed = mutation killed
	if err != nil {
		mutation.TestOutput = string(output)
		return true
	}

	// Tests passed = mutation survived
	return false
}

// applyMutation applies a mutation to source code
func (mt *MutationTester) applyMutation(src []byte, mutation Mutation) []byte {
	lines := strings.Split(string(src), "\n")

	if mutation.Line > 0 && mutation.Line <= len(lines) {
		line := lines[mutation.Line-1]
		mutatedLine := strings.Replace(line, mutation.Original, mutation.Mutated, 1)
		lines[mutation.Line-1] = mutatedLine
	}

	return []byte(strings.Join(lines, "\n"))
}

// ConditionalMutator mutates conditional operators
type ConditionalMutator struct{}

func (m *ConditionalMutator) Name() string { return "Conditional" }

func (m *ConditionalMutator) Mutate(node ast.Node, fset *token.FileSet) []Mutation {
	var mutations []Mutation

	switch n := node.(type) {
	case *ast.BinaryExpr:
		pos := fset.Position(n.OpPos)

		switch n.Op {
		case token.EQL: // ==
			mutations = append(mutations, Mutation{
				Type:        "Conditional",
				Line:        pos.Line,
				Original:    "==",
				Mutated:     "!=",
				Description: "Replace == with !=",
			})
		case token.NEQ: // !=
			mutations = append(mutations, Mutation{
				Type:        "Conditional",
				Line:        pos.Line,
				Original:    "!=",
				Mutated:     "==",
				Description: "Replace != with ==",
			})
		case token.LSS: // <
			mutations = append(mutations, Mutation{
				Type:        "Conditional",
				Line:        pos.Line,
				Original:    "<",
				Mutated:     "<=",
				Description: "Replace < with <=",
			})
		case token.GTR: // >
			mutations = append(mutations, Mutation{
				Type:        "Conditional",
				Line:        pos.Line,
				Original:    ">",
				Mutated:     ">=",
				Description: "Replace > with >=",
			})
		case token.LEQ: // <=
			mutations = append(mutations, Mutation{
				Type:        "Conditional",
				Line:        pos.Line,
				Original:    "<=",
				Mutated:     "<",
				Description: "Replace <= with <",
			})
		case token.GEQ: // >=
			mutations = append(mutations, Mutation{
				Type:        "Conditional",
				Line:        pos.Line,
				Original:    ">=",
				Mutated:     ">",
				Description: "Replace >= with >",
			})
		}
	}

	return mutations
}

// ArithmeticMutator mutates arithmetic operators
type ArithmeticMutator struct{}

func (m *ArithmeticMutator) Name() string { return "Arithmetic" }

func (m *ArithmeticMutator) Mutate(node ast.Node, fset *token.FileSet) []Mutation {
	var mutations []Mutation

	switch n := node.(type) {
	case *ast.BinaryExpr:
		pos := fset.Position(n.OpPos)

		switch n.Op {
		case token.ADD: // +
			mutations = append(mutations, Mutation{
				Type:        "Arithmetic",
				Line:        pos.Line,
				Original:    "+",
				Mutated:     "-",
				Description: "Replace + with -",
			})
		case token.SUB: // -
			mutations = append(mutations, Mutation{
				Type:        "Arithmetic",
				Line:        pos.Line,
				Original:    "-",
				Mutated:     "+",
				Description: "Replace - with +",
			})
		case token.MUL: // *
			mutations = append(mutations, Mutation{
				Type:        "Arithmetic",
				Line:        pos.Line,
				Original:    "*",
				Mutated:     "/",
				Description: "Replace * with /",
			})
		case token.QUO: // /
			mutations = append(mutations, Mutation{
				Type:        "Arithmetic",
				Line:        pos.Line,
				Original:    "/",
				Mutated:     "*",
				Description: "Replace / with *",
			})
		}
	case *ast.IncDecStmt:
		pos := fset.Position(n.TokPos)

		if n.Tok == token.INC { // ++
			mutations = append(mutations, Mutation{
				Type:        "Arithmetic",
				Line:        pos.Line,
				Original:    "++",
				Mutated:     "--",
				Description: "Replace ++ with --",
			})
		} else if n.Tok == token.DEC { // --
			mutations = append(mutations, Mutation{
				Type:        "Arithmetic",
				Line:        pos.Line,
				Original:    "--",
				Mutated:     "++",
				Description: "Replace -- with ++",
			})
		}
	}

	return mutations
}

// LogicalMutator mutates logical operators
type LogicalMutator struct{}

func (m *LogicalMutator) Name() string { return "Logical" }

func (m *LogicalMutator) Mutate(node ast.Node, fset *token.FileSet) []Mutation {
	var mutations []Mutation

	switch n := node.(type) {
	case *ast.BinaryExpr:
		pos := fset.Position(n.OpPos)

		switch n.Op {
		case token.LAND: // &&
			mutations = append(mutations, Mutation{
				Type:        "Logical",
				Line:        pos.Line,
				Original:    "&&",
				Mutated:     "||",
				Description: "Replace && with ||",
			})
		case token.LOR: // ||
			mutations = append(mutations, Mutation{
				Type:        "Logical",
				Line:        pos.Line,
				Original:    "||",
				Mutated:     "&&",
				Description: "Replace || with &&",
			})
		}
	case *ast.UnaryExpr:
		if n.Op == token.NOT { // !
			pos := fset.Position(n.OpPos)
			mutations = append(mutations, Mutation{
				Type:        "Logical",
				Line:        pos.Line,
				Original:    "!",
				Mutated:     "",
				Description: "Remove negation !",
			})
		}
	}

	return mutations
}

// ReturnMutator mutates return values
type ReturnMutator struct{}

func (m *ReturnMutator) Name() string { return "Return" }

func (m *ReturnMutator) Mutate(node ast.Node, fset *token.FileSet) []Mutation {
	var mutations []Mutation

	switch n := node.(type) {
	case *ast.ReturnStmt:
		pos := fset.Position(n.Return)

		// Mutate boolean returns
		for _, result := range n.Results {
			if ident, ok := result.(*ast.Ident); ok {
				switch ident.Name {
				case "true":
					mutations = append(mutations, Mutation{
						Type:        "Return",
						Line:        pos.Line,
						Original:    "true",
						Mutated:     "false",
						Description: "Replace return true with false",
					})
				case "false":
					mutations = append(mutations, Mutation{
						Type:        "Return",
						Line:        pos.Line,
						Original:    "false",
						Mutated:     "true",
						Description: "Replace return false with true",
					})
				case "nil":
					// Could mutate nil returns, but more complex
				}
			}
		}
	}

	return mutations
}

// ConstantMutator mutates constant values
type ConstantMutator struct{}

func (m *ConstantMutator) Name() string { return "Constant" }

func (m *ConstantMutator) Mutate(node ast.Node, fset *token.FileSet) []Mutation {
	var mutations []Mutation

	switch n := node.(type) {
	case *ast.BasicLit:
		pos := fset.Position(n.Pos())

		switch n.Kind {
		case token.INT:
			if n.Value == "0" {
				mutations = append(mutations, Mutation{
					Type:        "Constant",
					Line:        pos.Line,
					Original:    "0",
					Mutated:     "1",
					Description: "Replace 0 with 1",
				})
			} else if n.Value == "1" {
				mutations = append(mutations, Mutation{
					Type:        "Constant",
					Line:        pos.Line,
					Original:    "1",
					Mutated:     "0",
					Description: "Replace 1 with 0",
				})
			}
		case token.STRING:
			if n.Value == `""` {
				mutations = append(mutations, Mutation{
					Type:        "Constant",
					Line:        pos.Line,
					Original:    `""`,
					Mutated:     `"mutated"`,
					Description: "Replace empty string with non-empty",
				})
			}
		}
	}

	return mutations
}

// GenerateReport creates output reports
func GenerateReport(report *MutationReport, outputDir string) error {
	// Ensure output directory exists
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return err
	}

	// Generate JSON report
	jsonFile := filepath.Join(outputDir, "mutation-report.json")
	jsonData, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(jsonFile, jsonData, 0644); err != nil {
		return err
	}

	// Generate text report
	textFile := filepath.Join(outputDir, "mutation-report.txt")
	if err := generateTextReport(report, textFile); err != nil {
		return err
	}

	// Generate HTML report
	htmlFile := filepath.Join(outputDir, "mutation-report.html")
	if err := generateHTMLReport(report, htmlFile); err != nil {
		return err
	}

	return nil
}

// generateTextReport creates a human-readable text report
func generateTextReport(report *MutationReport, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	w := bufio.NewWriter(file)
	defer w.Flush()

	fmt.Fprintln(w, strings.Repeat("=", 80))
	fmt.Fprintln(w, "MUTATION TESTING REPORT")
	fmt.Fprintln(w, strings.Repeat("=", 80))
	fmt.Fprintln(w)

	fmt.Fprintf(w, "Mutation Score: %.2f%%\n", report.MutationScore)
	fmt.Fprintf(w, "Total Mutants: %d\n", report.TotalMutants)
	fmt.Fprintf(w, "Killed: %d\n", report.KilledMutants)
	fmt.Fprintf(w, "Survived: %d\n", report.SurvivedMutants)
	fmt.Fprintf(w, "Duration: %s\n", report.Duration)
	fmt.Fprintln(w)

	if report.SurvivedMutants > 0 {
		fmt.Fprintln(w, strings.Repeat("-", 80))
		fmt.Fprintln(w, "SURVIVING MUTANTS (Tests didn't catch these):")
		fmt.Fprintln(w, strings.Repeat("-", 80))

		for filename, fileResult := range report.Files {
			if len(fileResult.Survivors) > 0 {
				fmt.Fprintf(w, "\n%s:\n", filename)
				for _, survivor := range fileResult.Survivors {
					fmt.Fprintf(w, "  Line %d: %s\n", survivor.Line, survivor.Description)
					fmt.Fprintf(w, "    Original: %s\n", survivor.Original)
					fmt.Fprintf(w, "    Mutated:  %s\n", survivor.Mutated)
				}
			}
		}
	}

	return nil
}

// generateHTMLReport creates an HTML report
func generateHTMLReport(report *MutationReport, filename string) error {
	html := `<!DOCTYPE html>
<html>
<head>
    <title>Mutation Testing Report</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; }
        h1 { color: #333; }
        .score { font-size: 48px; font-weight: bold; }
        .high { color: green; }
        .medium { color: orange; }
        .low { color: red; }
        .summary { background: #f0f0f0; padding: 20px; border-radius: 5px; margin: 20px 0; }
        .survivor { background: #fff5f5; border-left: 3px solid red; padding: 10px; margin: 10px 0; }
        code { background: #f4f4f4; padding: 2px 4px; border-radius: 3px; }
    </style>
</head>
<body>
    <h1>Mutation Testing Report</h1>

    <div class="summary">
        <div class="score %s">%.2f%%</div>
        <p>Mutation Score</p>

        <table>
            <tr><td>Total Mutants:</td><td>%d</td></tr>
            <tr><td>Killed:</td><td>%d</td></tr>
            <tr><td>Survived:</td><td>%d</td></tr>
            <tr><td>Duration:</td><td>%s</td></tr>
        </table>
    </div>

    %s
</body>
</html>`

	scoreClass := "low"
	if report.MutationScore >= 80 {
		scoreClass = "high"
	} else if report.MutationScore >= 60 {
		scoreClass = "medium"
	}

	var survivorsHTML strings.Builder
	if report.SurvivedMutants > 0 {
		survivorsHTML.WriteString("<h2>Surviving Mutants</h2>")
		for filename, fileResult := range report.Files {
			if len(fileResult.Survivors) > 0 {
				survivorsHTML.WriteString(fmt.Sprintf("<h3>%s</h3>", filename))
				for _, survivor := range fileResult.Survivors {
					survivorsHTML.WriteString(fmt.Sprintf(
						`<div class="survivor">
							<strong>Line %d:</strong> %s<br>
							Original: <code>%s</code> → Mutated: <code>%s</code>
						</div>`,
						survivor.Line, survivor.Description,
						survivor.Original, survivor.Mutated,
					))
				}
			}
		}
	}

	content := fmt.Sprintf(html,
		scoreClass, report.MutationScore,
		report.TotalMutants, report.KilledMutants, report.SurvivedMutants, report.Duration,
		survivorsHTML.String(),
	)

	return os.WriteFile(filename, []byte(content), 0644)
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run scripts/mutation-test/main.go <source-dir> [output-dir]")
		fmt.Println("   or: mutation-test <source-dir> [output-dir]")
		fmt.Println("")
		fmt.Println("Arguments:")
		fmt.Println("  source-dir:  Directory containing Go source files to mutate")
		fmt.Println("  output-dir:  Directory for output reports (default: mutation-report)")
		fmt.Println("")
		fmt.Println("Options:")
		fmt.Println("  --quick      Test only a subset of mutations")
		fmt.Println("  --parallel   Run mutation tests in parallel")
		fmt.Println("")
		fmt.Println("Example:")
		fmt.Println("  mutation-test ./src")
		fmt.Println("  mutation-test --quick ./src reports")
		os.Exit(1)
	}

	// Parse arguments
	sourceDir := ""
	outputDir := "mutation-report"
	quickMode := false
	parallelMode := false

	// Simple argument parsing
	argIdx := 1
	for argIdx < len(os.Args) {
		arg := os.Args[argIdx]
		if arg == "--quick" {
			quickMode = true
			argIdx++
		} else if arg == "--parallel" {
			parallelMode = true
			argIdx++
		} else if !strings.HasPrefix(arg, "--") {
			sourceDir = arg
			argIdx++
			break
		} else {
			argIdx++
		}
	}

	if sourceDir == "" {
		fmt.Println("Error: source-dir is required")
		os.Exit(1)
	}

	// Parse output dir if provided
	if argIdx < len(os.Args) {
		outputDir = os.Args[argIdx]
	}

	if quickMode {
		fmt.Println("Running in quick mode (testing subset of mutations)")
	}
	if parallelMode {
		fmt.Println("Running in parallel mode")
	}

	fmt.Println("Starting mutation testing...")
	tester := NewMutationTester(sourceDir)

	// Apply modes
	if quickMode {
		tester.SetQuickMode()
	}
	if parallelMode {
		tester.SetParallelMode()
	}

	report, err := tester.Run()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	if err := GenerateReport(report, outputDir); err != nil {
		fmt.Printf("Error generating reports: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("\nMutation Testing Complete!\n")
	fmt.Printf("Mutation Score: %.2f%%\n", report.MutationScore)
	fmt.Printf("Reports generated in %s\n", outputDir)

	// Exit with error if mutation score is too low
	if report.MutationScore < 80 {
		fmt.Println("\n⚠️  Warning: Low mutation score indicates tests may not be effective")
		os.Exit(1)
	}
}

# Quickstart: Achieving 100% Test Coverage for buffer_scraper.go

## Prerequisites
- Go 1.21+ installed
- Repository cloned and dependencies installed
- Basic understanding of Go testing

## Quick Coverage Check
```bash
# Run tests and generate coverage
cd tests/unit/scraper
go test -coverprofile=coverage.out -coverpkg=github.com/yhonda-ohishi/etc_meisai/src/scraper .

# View coverage for buffer_scraper.go
go tool cover -func=coverage.out | grep buffer_scraper.go

# Generate HTML report for detailed view
go tool cover -html=coverage.out -o coverage.html
# Open coverage.html in browser
```

## Understanding Current Coverage

### Viewing Uncovered Lines
```bash
# Show uncovered lines in terminal (colored output)
go test -coverprofile=coverage.out -coverpkg=./src/scraper ./tests/unit/scraper
go tool cover -html=coverage.out
```

Look for red-highlighted lines - these are uncovered code paths.

### Current Coverage Gaps (73.7%)
- **Line 32**: Production path when `mockDownloader` is nil
- **Lines 33-35**: Error handling for `DownloadMeisai` failure
- **Lines 40-42**: Error handling for file read failure
- **Lines 48-52**: Deferred file removal error logging

## Adding Test Cases

### Step 1: Test Error Paths
Create `tests/unit/scraper/buffer_error_test.go`:

```go
package scraper_test

import (
    "errors"
    "testing"
    "github.com/yhonda-ohishi/etc_meisai/src/scraper"
)

func TestDownloadMeisaiToBuffer_Errors(t *testing.T) {
    t.Run("DownloadMeisai_Error", func(t *testing.T) {
        // Create scraper with mock that returns error
        config := &scraper.ScraperConfig{
            UserID:   "test",
            Password: "pass",
        }
        s, _ := scraper.NewETCScraper(config, nil)

        // Set mock that returns error
        mockDownloader := &MockDownloader{
            err: errors.New("network error"),
        }
        s.SetMockDownloader(mockDownloader)

        // Call method
        result, err := s.DownloadMeisaiToBuffer("2024-01-01", "2024-01-31")

        // Assert error
        if err == nil {
            t.Error("Expected error but got none")
        }
        if result != nil {
            t.Error("Expected nil result on error")
        }
    })
}
```

### Step 2: Test Production Paths
Create `tests/unit/scraper/buffer_production_test.go`:

```go
package scraper_test

import (
    "os"
    "testing"
    "github.com/yhonda-ohishi/etc_meisai/src/scraper"
)

func TestDownloadMeisaiToBuffer_ProductionPath(t *testing.T) {
    t.Run("Nil_MockDownloader", func(t *testing.T) {
        // Setup test CSV file
        testDir := "./test_production"
        os.MkdirAll(testDir, 0755)
        defer os.RemoveAll(testDir)

        csvPath := testDir + "/test.csv"
        csvContent := "header1,header2\nvalue1,value2\n"
        os.WriteFile(csvPath, []byte(csvContent), 0644)

        // Create scraper without mocks
        config := &scraper.ScraperConfig{
            UserID:   "test",
            Password: "pass",
        }
        s, _ := scraper.NewETCScraper(config, nil)
        // Don't set mock - force production path

        // Use environment variable to control behavior
        os.Setenv("MOCK_CSV_PATH", csvPath)
        defer os.Unsetenv("MOCK_CSV_PATH")

        // Call method
        result, err := s.DownloadMeisaiToBuffer("2024-01-01", "2024-01-31")

        // Assert success
        if err != nil {
            t.Fatalf("Unexpected error: %v", err)
        }
        if result == nil {
            t.Fatal("Expected result but got nil")
        }
        if string(result.CSVData) != csvContent {
            t.Error("CSV content mismatch")
        }
    })
}
```

### Step 3: Test Edge Cases
Add edge case tests:

```go
func TestParseCSVBuffer_EdgeCases(t *testing.T) {
    tests := []struct {
        name     string
        csvData  []byte
        expected int // expected record count
    }{
        {"empty_file", []byte{}, 0},
        {"only_headers", []byte("h1,h2,h3\n"), 0},
        {"with_quotes", []byte(`"a,b",c` + "\n" + `d,"e,f"`), 1},
        {"multiline", []byte(`"line1` + "\n" + `line2",value`), 0},
    }

    for _, tc := range tests {
        t.Run(tc.name, func(t *testing.T) {
            s, _ := scraper.NewETCScraper(&scraper.ScraperConfig{}, nil)
            count := s.CountCSVRecordsForTest(tc.csvData)
            if count != tc.expected {
                t.Errorf("Expected %d records, got %d", tc.expected, count)
            }
        })
    }
}
```

## Using Mock Interfaces

### Creating Mock Implementations
```go
type MockDownloader struct {
    downloadFunc func(string, string) (string, error)
    err          error
}

func (m *MockDownloader) DownloadMeisai(from, to string) (string, error) {
    if m.downloadFunc != nil {
        return m.downloadFunc(from, to)
    }
    if m.err != nil {
        return "", m.err
    }
    return "mock.csv", nil
}

type MockFileReader struct {
    readFunc func(string) ([]byte, error)
    err      error
}

func (m *MockFileReader) ReadFileToBuffer(path string) ([]byte, error) {
    if m.readFunc != nil {
        return m.readFunc(path)
    }
    if m.err != nil {
        return nil, m.err
    }
    return []byte("mock,data"), nil
}
```

### Setting Up Mocks in Tests
```go
func TestWithMocks(t *testing.T) {
    scraper, _ := scraper.NewETCScraper(config, nil)

    // Set custom mock behavior
    mockDownloader := &MockDownloader{
        downloadFunc: func(from, to string) (string, error) {
            // Custom logic here
            return "test.csv", nil
        },
    }
    scraper.SetMockDownloader(mockDownloader)

    // Test with mock
    result, err := scraper.DownloadMeisaiToBuffer("", "")
    // Assertions...
}
```

## Running Coverage Analysis

### Generate Coverage Report
```bash
# Run all tests with coverage
go test -coverprofile=coverage.out -coverpkg=./src/scraper ./tests/unit/scraper

# View summary
go tool cover -func=coverage.out | tail -5

# Check specific file
go tool cover -func=coverage.out | grep buffer_scraper.go
```

### Visualize Coverage
```bash
# Generate HTML report
go tool cover -html=coverage.out -o coverage.html

# Open in browser (Windows)
start coverage.html

# Open in browser (Linux/Mac)
open coverage.html
```

### Coverage Script
Use the provided script:
```bash
# Make executable (Linux/Mac)
chmod +x show_coverage.sh

# Run coverage analysis
./show_coverage.sh

# On Windows with Git Bash
bash show_coverage.sh
```

## Troubleshooting

### Common Issues

#### 1. Test File Conflicts
**Problem**: Tests fail with "file not found"
**Solution**: Use isolated directories for each test:
```go
testDir := fmt.Sprintf("./test_%s", t.Name())
os.MkdirAll(testDir, 0755)
defer os.RemoveAll(testDir)
```

#### 2. Mock Not Being Used
**Problem**: Production code runs instead of mock
**Solution**: Ensure mock is set before calling method:
```go
scraper.SetMockDownloader(mockDownloader)
// Then call method
```

#### 3. Coverage Not Improving
**Problem**: Added tests but coverage stays same
**Solution**: Check you're testing the right package:
```bash
go test -coverpkg=github.com/yhonda-ohishi/etc_meisai/src/scraper
```

#### 4. Deferred Functions Not Covered
**Problem**: Deferred cleanup code shows as uncovered
**Solution**: Force error in deferred function:
```go
// Make file read-only to cause removal error
os.Chmod(filepath, 0444)
// This will trigger error logging in deferred cleanup
```

## Best Practices

### 1. Test Independence
Each test should:
- Create its own test data
- Clean up after itself
- Not depend on other tests

### 2. Table-Driven Tests
Use for comprehensive coverage:
```go
tests := []struct {
    name          string
    setup         func()
    expectedError bool
}{
    // Test cases
}
```

### 3. Parallel Testing
Speed up test execution:
```go
t.Run("test_name", func(t *testing.T) {
    t.Parallel()
    // Test code
})
```

### 4. Clear Test Names
Use descriptive names:
- ✅ `TestDownloadMeisaiToBuffer_NetworkError`
- ❌ `TestError1`

## Validation Checklist

Before considering coverage complete:

- [ ] All methods show 100% coverage
- [ ] Error paths tested with actual errors
- [ ] Production paths tested (nil mocks)
- [ ] Edge cases covered (empty, malformed data)
- [ ] Deferred functions tested
- [ ] Tests run in < 2 seconds
- [ ] No flaky tests (run 10 times successfully)
- [ ] Coverage report generated and reviewed
- [ ] No panics in any test scenario
- [ ] Documentation updated

## Next Steps

1. Run initial coverage check
2. Identify specific uncovered lines
3. Add targeted test cases
4. Iterate until 100% achieved
5. Run validation checklist
6. Document any remaining gaps

## Support

For questions or issues:
1. Check existing tests in `tests/unit/scraper/`
2. Review the research.md for design decisions
3. Consult the data-model.md for interface details
4. See test-contracts.md for specific scenarios
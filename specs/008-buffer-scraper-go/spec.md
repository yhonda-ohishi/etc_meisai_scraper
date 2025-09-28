# Feature Specification: Achieve 100% Test Coverage for buffer_scraper.go

## Executive Summary
This specification outlines the approach to achieve 100% test coverage for `src/scraper/buffer_scraper.go`. Current coverage stands at approximately 73.7% for critical methods, with specific gaps in error handling paths and edge cases. The solution involves making the code more mockable through dependency injection and interface segregation.

## Current State Analysis

### Coverage Metrics (As of 2025-09-28)
- **DownloadMeisaiToBuffer**: 73.7% (Lines 18-69)
  - Covered: Mock path, environment variable path
  - Uncovered: Production path (line 32), error handling (lines 34-35, 40-42), file removal logic (lines 48-52)

- **DownloadMeisaiAsReader**: 83.3% (Lines 72-84)
  - Covered: Mock injection path, success path
  - Uncovered: Production path without mock (line 32 call)

- **GetCSVData**: 83.3% (Lines 87-100)
  - Covered: Mock injection path, success path
  - Uncovered: Production path without mock (line 32 call)

### Root Cause Analysis
The primary coverage gaps stem from:
1. **Direct dependency on external method**: Lines calling `s.DownloadMeisai()` directly
2. **OS-level operations**: File removal and error handling in deferred functions
3. **Production-only code paths**: Code that only executes when mocks are nil

## Proposed Solution

### 1. Interface Segregation
Create focused interfaces for each responsibility:

```go
// Download interface for CSV acquisition
type MeisaiDownloader interface {
    DownloadMeisai(fromDate, toDate string) (string, error)
}

// File operations interface
type FileOperations interface {
    ReadFile(path string) ([]byte, error)
    RemoveFile(path string) error
    FileExists(path string) bool
}

// Logger interface for testability
type Logger interface {
    Printf(format string, v ...interface{})
}
```

### 2. Dependency Injection Pattern
Modify ETCScraper to accept interfaces:

```go
type ETCScraper struct {
    downloader    MeisaiDownloader
    fileOps       FileOperations
    logger        Logger
    // ... other fields
}
```

### 3. Factory Pattern for Production vs Test
Create factory functions for different environments:

```go
// Production factory
func NewProductionScraper(config *ScraperConfig) *ETCScraper {
    return &ETCScraper{
        downloader: &DefaultDownloader{},
        fileOps:    &OSFileOperations{},
        logger:     log.New(os.Stdout, "[SCRAPER] ", log.LstdFlags),
    }
}

// Test factory
func NewTestScraper(config *ScraperConfig) *ETCScraper {
    return &ETCScraper{
        downloader: &MockDownloader{},
        fileOps:    &MockFileOperations{},
        logger:     &MockLogger{},
    }
}
```

### 4. Refactoring Strategy

#### Phase 1: Extract File Operations
Move all file system operations to the FileOperations interface:
- ReadFileToBuffer becomes fileOps.ReadFile
- os.Remove becomes fileOps.RemoveFile
- Add fileOps.FileExists for defensive checks

#### Phase 2: Extract Download Logic
Create a separate downloader component:
- Move DownloadMeisai logic to DefaultDownloader
- Implement MockDownloader for tests
- Remove direct method calls from buffer methods

#### Phase 3: Improve Error Handling
Make error paths testable:
- Extract deferred cleanup to named functions
- Add error injection points for testing
- Implement comprehensive error scenarios

## Test Coverage Plan

### Test Scenarios to Add

1. **Error Path Tests**
   - DownloadMeisai returns error
   - ReadFile returns error
   - File removal fails
   - Invalid CSV data parsing

2. **Edge Case Tests**
   - Empty CSV file
   - Malformed CSV data
   - Large file handling
   - Concurrent access scenarios

3. **Integration Tests**
   - Full flow with real file operations
   - Mock downloader with real file ops
   - Real downloader with mock file ops

### Test Implementation Structure

```go
// tests/unit/scraper/buffer_scraper_test.go
func TestDownloadMeisaiToBuffer_AllPaths(t *testing.T) {
    tests := []struct {
        name           string
        setupMocks     func(*MockDownloader, *MockFileOps)
        expectedError  string
        expectedResult *BufferResult
    }{
        {
            name: "successful_download",
            setupMocks: func(md *MockDownloader, mf *MockFileOps) {
                md.On("DownloadMeisai", mock.Anything, mock.Anything).Return("/path/to/csv", nil)
                mf.On("ReadFile", "/path/to/csv").Return([]byte("csv,data"), nil)
                mf.On("RemoveFile", "/path/to/csv").Return(nil)
            },
            expectedResult: &BufferResult{CSVData: []byte("csv,data"), RecordCount: 1},
        },
        {
            name: "download_error",
            setupMocks: func(md *MockDownloader, mf *MockFileOps) {
                md.On("DownloadMeisai", mock.Anything, mock.Anything).Return("", errors.New("network error"))
            },
            expectedError: "failed to download CSV: network error",
        },
        // ... more test cases
    }
}
```

## Implementation Timeline

### Week 1: Refactoring (Days 1-3)
- [ ] Create interface definitions
- [ ] Implement production adapters
- [ ] Update ETCScraper to use interfaces

### Week 1: Testing (Days 4-5)
- [ ] Create mock implementations
- [ ] Write comprehensive unit tests
- [ ] Achieve 100% coverage for refactored code

### Week 2: Integration (Days 6-7)
- [ ] Update existing tests to use new patterns
- [ ] Ensure backward compatibility
- [ ] Performance benchmarking

### Week 2: Documentation (Day 8)
- [ ] Update technical documentation
- [ ] Create migration guide for other components
- [ ] Update CLAUDE.md with new patterns

## Success Criteria

1. **Coverage Metrics**
   - 100% line coverage for buffer_scraper.go
   - 100% branch coverage for all decision points
   - All error paths tested

2. **Code Quality**
   - No direct OS calls in business logic
   - All dependencies injected
   - Clear separation of concerns

3. **Test Quality**
   - Fast execution (< 1 second for unit tests)
   - Deterministic results
   - No test interdependencies

## Risk Mitigation

### Risk 1: Breaking Changes
- **Mitigation**: Keep existing public APIs unchanged
- **Strategy**: Add new methods alongside existing ones, deprecate gradually

### Risk 2: Performance Regression
- **Mitigation**: Benchmark before and after changes
- **Strategy**: Use interface assertions to avoid reflection overhead

### Risk 3: Test Complexity
- **Mitigation**: Use table-driven tests for clarity
- **Strategy**: Create test helpers and builders for common scenarios

## Technical Decisions

1. **Why Interfaces Over Function Types**: Interfaces provide better IDE support and allow grouping related operations
2. **Why Factory Pattern**: Enables clean separation between production and test configurations
3. **Why Not Remove Unmockable Code**: Maintaining backward compatibility while improving testability

## Appendix: Code Samples

### Sample Mock Implementation
```go
type MockDownloader struct {
    mock.Mock
}

func (m *MockDownloader) DownloadMeisai(fromDate, toDate string) (string, error) {
    args := m.Called(fromDate, toDate)
    return args.String(0), args.Error(1)
}
```

### Sample Production Adapter
```go
type DefaultDownloader struct {
    scraper *ETCScraper
}

func (d *DefaultDownloader) DownloadMeisai(fromDate, toDate string) (string, error) {
    // Original implementation moved here
    return d.scraper.performDownload(fromDate, toDate)
}
```

## References
- [Go Testing Best Practices](https://golang.org/doc/test)
- [Testify Mock Documentation](https://github.com/stretchr/testify)
- [SOLID Principles in Go](https://dave.cheney.net/2016/08/20/solid-go-design)
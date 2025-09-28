# Research: Achieving 100% Test Coverage for buffer_scraper.go

## Executive Summary
This research document consolidates findings on achieving 100% test coverage for `buffer_scraper.go`. The primary challenge is testing production code paths when mock dependencies are nil, and handling error scenarios in deferred functions.

## Coverage Gap Analysis

### Current State
- Overall coverage: 73.7% for DownloadMeisaiToBuffer
- Uncovered areas: Production paths, error handling, file cleanup
- Existing patterns: mockDownloader and mockFileReader interfaces

### Detailed Gap Identification
```go
// Lines 32-35: Production path when mockDownloader is nil
if s.mockDownloader != nil {
    return s.DownloadMeisaiToBufferWithInjection(...)
}
// This path not covered:
csvPath, err = s.DownloadMeisai(fromDate, toDate)

// Lines 48-52: Deferred file removal error logging
defer func() {
    if err := os.Remove(csvPath); err != nil {
        // This error logging never tested
        s.logger.Printf("Warning: failed to remove temp file")
    }
}()
```

## Testing Pattern Research

### Decision: Table-Driven Tests with Dependency Injection
**Rationale**: Provides comprehensive coverage with minimal code duplication
**Alternatives Considered**:
- Monkey patching: Rejected due to runtime complexity
- Build tags: Rejected due to compilation overhead
- Global variables: Rejected due to test isolation concerns

### Best Practices for Go Testing
1. **Interface Segregation**: Small, focused interfaces
2. **Constructor Injection**: Dependencies passed at creation
3. **Test Isolation**: Each test independent with own setup
4. **Error Simulation**: Use dedicated error types

## Mock Strategy Research

### Decision: Manual Mocks Over testify/mock
**Rationale**: Simpler for this use case, better performance
**Alternatives Considered**:
- testify/mock: More complex for simple interfaces
- gomock: Requires code generation
- Custom function types: Less IDE support

### Mock Implementation Pattern
```go
type mockDownloader struct {
    downloadFunc func(string, string) (string, error)
}

func (m *mockDownloader) DownloadMeisai(from, to string) (string, error) {
    if m.downloadFunc != nil {
        return m.downloadFunc(from, to)
    }
    return "", nil
}
```

## File Operation Testing

### Decision: Interface Wrapper for OS Operations
**Rationale**: Allows testing of file operations without actual filesystem
**Alternatives Considered**:
- afero filesystem: Too heavy for simple operations
- tmpfs: Platform-specific complications
- Real files only: Cannot test permission errors

### File Operations Interface Design
```go
type FileOperations interface {
    ReadFile(path string) ([]byte, error)
    RemoveFile(path string) error
    FileExists(path string) bool
}

type OSFileOperations struct{}

func (o *OSFileOperations) ReadFile(path string) ([]byte, error) {
    return os.ReadFile(path)
}

func (o *OSFileOperations) RemoveFile(path string) error {
    return os.Remove(path)
}
```

## Error Path Testing

### Decision: Explicit Error Injection Points
**Rationale**: Predictable error simulation
**Alternatives Considered**:
- Random error injection: Non-deterministic
- System call interception: Platform-specific
- Production error conditions: Hard to reproduce

### Error Testing Strategy
1. **Network Errors**: Mock DownloadMeisai returns error
2. **File Read Errors**: Mock ReadFile returns error
3. **Permission Errors**: Mock RemoveFile returns error
4. **Parse Errors**: Provide malformed CSV data

## Production Path Testing

### Decision: Nil-Safe Pattern with Fallback
**Rationale**: Test actual production behavior
**Alternatives Considered**:
- Separate production tests: Code duplication
- Integration tests only: Slower feedback
- Skip production paths: Incomplete coverage

### Testing Nil Dependencies
```go
func TestProductionPath(t *testing.T) {
    scraper := &ETCScraper{
        mockDownloader: nil, // Force production path
    }

    // Use environment variable for control
    os.Setenv("MOCK_CSV_PATH", "test.csv")
    defer os.Unsetenv("MOCK_CSV_PATH")

    result, err := scraper.DownloadMeisaiToBuffer("", "")
    // Assertions...
}
```

## Deferred Function Testing

### Decision: Extract Cleanup to Named Functions
**Rationale**: Makes deferred operations testable
**Alternatives Considered**:
- Ignore deferred coverage: Incomplete testing
- Mock os package: Too invasive
- Use recover(): Only for panics

### Deferred Function Pattern
```go
func (s *ETCScraper) cleanupFile(path string) {
    if err := s.fileOps.RemoveFile(path); err != nil {
        if s.logger != nil {
            s.logger.Printf("Warning: failed to remove temp file %s: %v", path, err)
        }
    }
}

// In main function:
defer s.cleanupFile(csvPath)
```

## CSV Parsing Edge Cases

### Identified Edge Cases
1. Empty file (0 bytes)
2. Only headers, no data
3. Quoted fields with commas
4. Multiline quoted fields
5. Missing columns
6. Extra columns
7. UTF-8 BOM prefix
8. Different line endings (CRLF vs LF)

### Decision: Comprehensive Test Data Set
**Rationale**: Cover all parsing branches
**Implementation**: Table-driven tests with edge case data

## Performance Considerations

### Decision: Keep Tests Fast (<2s total)
**Rationale**: Quick feedback loop
**Strategies**:
1. Use in-memory data where possible
2. Parallel test execution with t.Parallel()
3. Small test data sets
4. No network calls in unit tests

## Backward Compatibility

### Decision: Add New Methods, Deprecate Gradually
**Rationale**: No breaking changes
**Strategy**:
1. Keep existing public APIs
2. Add new testable versions
3. Mark old versions deprecated
4. Remove in next major version

## Implementation Priority

### Quick Wins (1 day)
1. Error path tests: +15% coverage
2. Edge case tests: +5% coverage
3. Deferred function extraction: +5% coverage

### Medium Effort (2 days)
1. Production path tests: +10% coverage
2. Interface extraction: Enables remaining coverage

### Full Refactor (if needed)
1. Complete dependency injection: 100% coverage possible
2. Only if quick wins insufficient

## Recommendations

### Immediate Actions
1. Add error simulation tests
2. Test with nil mocks
3. Add edge case CSV data tests
4. Extract deferred cleanup

### Future Improvements
1. Consider interface wrapper for all OS operations
2. Add benchmark tests for performance regression
3. Create test helper package for common mocks
4. Document testing patterns for team

## Conclusion

Achieving 100% coverage is feasible with minimal refactoring. The key is:
1. Test error paths explicitly
2. Handle nil dependency cases
3. Extract untestable code to testable functions
4. Use table-driven tests for comprehensive coverage

The approach maintains backward compatibility while improving testability.
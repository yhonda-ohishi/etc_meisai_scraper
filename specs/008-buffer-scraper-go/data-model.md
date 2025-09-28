# Data Model: Test Coverage Interfaces and Contracts

## Interface Definitions

### Core Interfaces for Dependency Injection

```go
// MeisaiDownloader handles CSV file downloading
type MeisaiDownloader interface {
    DownloadMeisai(fromDate, toDate string) (string, error)
}

// FileOperations handles file system operations
type FileOperations interface {
    ReadFile(path string) ([]byte, error)
    RemoveFile(path string) error
    FileExists(path string) bool
    Open(path string) (io.ReadCloser, error)
}

// Logger handles logging operations
type Logger interface {
    Printf(format string, v ...interface{})
}

// CSVParser handles CSV parsing operations
type CSVParser interface {
    Parse(data []byte) ([][]string, error)
    CountRecords(data []byte) int
}
```

## Data Structures

### BufferResult
```go
type BufferResult struct {
    CSVData     []byte            // Raw CSV file content
    RecordCount int               // Number of records (excluding header)
    Metadata    map[string]string // Additional metadata
}

// Metadata keys:
// - "from_date": Start date of the data range
// - "to_date": End date of the data range
// - "content_type": Always "text/csv"
// - "size_bytes": File size in bytes as string
```

### Test Mock Structures

```go
// MockDownloader for testing download operations
type MockDownloader struct {
    DownloadFunc func(string, string) (string, error)
    CallCount    int
    LastFromDate string
    LastToDate   string
}

// MockFileOperations for testing file operations
type MockFileOperations struct {
    ReadFileFunc   func(string) ([]byte, error)
    RemoveFileFunc func(string) error
    FileExistsFunc func(string) bool
    OpenFunc       func(string) (io.ReadCloser, error)

    // Tracking
    ReadCalls   []string // Paths that were read
    RemoveCalls []string // Paths that were removed
    ExistsCalls []string // Paths that were checked
}

// MockLogger for testing logging
type MockLogger struct {
    Messages []string // Captured log messages
}

// MockCSVParser for testing CSV parsing
type MockCSVParser struct {
    ParseFunc  func([]byte) ([][]string, error)
    CountFunc  func([]byte) int
}
```

## Error Types

```go
// Custom error types for testing
type DownloadError struct {
    Message string
}

func (e DownloadError) Error() string {
    return fmt.Sprintf("download failed: %s", e.Message)
}

type FileReadError struct {
    Path    string
    Message string
}

func (e FileReadError) Error() string {
    return fmt.Sprintf("failed to read %s: %s", e.Path, e.Message)
}

type FileRemoveError struct {
    Path    string
    Message string
}

func (e FileRemoveError) Error() string {
    return fmt.Sprintf("failed to remove %s: %s", e.Path, e.Message)
}

type CSVParseError struct {
    Line    int
    Message string
}

func (e CSVParseError) Error() string {
    return fmt.Sprintf("CSV parse error at line %d: %s", e.Line, e.Message)
}
```

## State Transitions

### Download Flow States
```
IDLE -> DOWNLOADING -> READING -> PARSING -> COMPLETE
         |              |          |
         v              v          v
       ERROR          ERROR      ERROR
```

### File Lifecycle States
```
NOT_EXISTS -> CREATED -> READ -> DELETED
                |         |
                v         v
             ERROR     ERROR
```

## Validation Rules

### Input Validation
- `fromDate` and `toDate` must be valid date strings (YYYY-MM-DD format)
- `fromDate` must be before or equal to `toDate`
- File paths must be absolute or relative to working directory
- CSV data must be valid UTF-8 encoded

### Output Validation
- `BufferResult.CSVData` must not be nil for successful operations
- `BufferResult.RecordCount` must be >= 0
- `BufferResult.Metadata` must contain all required keys
- File removal must not fail silently (log warnings)

## Test Data Contracts

### Minimal Valid CSV
```csv
header1,header2
value1,value2
```

### Edge Case CSVs
```csv
# Empty file
""

# Only headers
header1,header2,header3

# Quoted fields with commas
"field,with,commas","normal field"

# Multiline fields
"line1
line2","single line"

# Unicode content
名前,年齢
田中,25
```

## Dependency Injection Patterns

### Constructor Injection
```go
type ETCScraper struct {
    downloader MeisaiDownloader
    fileOps    FileOperations
    logger     Logger
    csvParser  CSVParser
    // ... other fields
}

func NewETCScraper(deps Dependencies) *ETCScraper {
    return &ETCScraper{
        downloader: deps.Downloader,
        fileOps:    deps.FileOps,
        logger:     deps.Logger,
        csvParser:  deps.CSVParser,
    }
}
```

### Setter Injection (for backward compatibility)
```go
func (s *ETCScraper) SetDownloader(d MeisaiDownloader) {
    s.downloader = d
}

func (s *ETCScraper) SetFileOperations(f FileOperations) {
    s.fileOps = f
}
```

## Production Implementations

### DefaultDownloader
```go
type DefaultDownloader struct {
    scraper *ETCScraper // Reference to parent for actual download
}

func (d *DefaultDownloader) DownloadMeisai(from, to string) (string, error) {
    // Actual download implementation
    return d.scraper.performActualDownload(from, to)
}
```

### OSFileOperations
```go
type OSFileOperations struct{}

func (o *OSFileOperations) ReadFile(path string) ([]byte, error) {
    return os.ReadFile(path)
}

func (o *OSFileOperations) RemoveFile(path string) error {
    return os.Remove(path)
}

func (o *OSFileOperations) FileExists(path string) bool {
    _, err := os.Stat(path)
    return err == nil
}

func (o *OSFileOperations) Open(path string) (io.ReadCloser, error) {
    return os.Open(path)
}
```

### StandardCSVParser
```go
type StandardCSVParser struct{}

func (p *StandardCSVParser) Parse(data []byte) ([][]string, error) {
    // Implementation using existing parseCSVBuffer logic
}

func (p *StandardCSVParser) CountRecords(data []byte) int {
    // Implementation using existing countCSVRecords logic
}
```

## Coverage Target Methods

Methods requiring 100% coverage:
1. `DownloadMeisaiToBuffer` (currently 73.7%)
2. `DownloadMeisaiAsReader` (currently 83.3%)
3. `GetCSVData` (currently 83.3%)
4. `readFileToBuffer` (currently 90.9%)
5. `parseCSVBuffer` (currently 87.5%)
6. `TestDownloadMeisaiWithMock` (currently 87.5%)
7. `DownloadMeisaiAsReaderWithPath` (currently 75.0%)
8. `GetCSVDataWithPath` (currently 75.0%)
9. `DefaultMeisaiDownloader.DownloadMeisai` (currently 0.0%)
10. All mockable variants (75-83% coverage)
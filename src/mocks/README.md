# Mock Implementations

This directory contains comprehensive mock implementations for all service interfaces defined in `src/grpc/interfaces.go`.

## Available Mocks

### Service Mocks

1. **MockETCMeisaiService** (`etc_meisai_service_mock.go`)
   - Implements: `ETCMeisaiServiceInterface`
   - Methods: CreateRecord, GetRecord, ListRecords, UpdateRecord, DeleteRecord, HealthCheck

2. **MockETCMappingService** (`etc_mapping_service_mock.go`)
   - Implements: `ETCMappingServiceInterface`
   - Methods: CreateMapping, GetMapping, ListMappings, UpdateMapping, DeleteMapping, UpdateStatus, HealthCheck

3. **MockImportService** (`import_service_mock.go`)
   - Implements: `ImportServiceInterface`
   - Methods: ImportCSV, ImportCSVStream, GetImportSession, ListImportSessions, ProcessCSV, ProcessCSVRow, HandleDuplicates, CancelImportSession, HealthCheck

4. **MockStatisticsService** (`statistics_service_mock.go`)
   - Implements: `StatisticsServiceInterface`
   - Methods: GetGeneralStatistics, GetDailyStatistics, GetMonthlyStatistics, GetVehicleStatistics, GetMappingStatistics, HealthCheck

5. **MockLogger** (`logger_mock.go`)
   - Implements: `LoggerInterface`
   - Methods: Printf, Println, Print, Fatalf, Fatal, Panicf, Panic

## Usage

All mocks use the `github.com/stretchr/testify/mock` framework. Here's how to use them:

### Basic Example

```go
import (
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
    "github.com/yhonda-ohishi/etc_meisai/src/mocks"
)

func TestSomething(t *testing.T) {
    // Create mock
    mockService := &mocks.MockETCMeisaiService{}

    // Set expectations
    mockService.On("HealthCheck", mock.Anything).Return(nil)

    // Use mock in your code
    err := mockService.HealthCheck(context.Background())

    // Verify
    assert.NoError(t, err)
    mockService.AssertExpectations(t)
}
```

### Setting Return Values

```go
// For methods that return objects
expectedRecord := &models.ETCMeisaiRecord{ID: 1, CarNumber: "TEST-001"}
mockService.On("GetRecord", mock.Anything, int64(1)).Return(expectedRecord, nil)

// For methods that return errors
mockService.On("DeleteRecord", mock.Anything, int64(1)).Return(nil)

// For methods with complex parameters
params := &services.CreateRecordParams{...}
mockService.On("CreateRecord", mock.Anything, params).Return(expectedRecord, nil)
```

### Argument Matching

```go
// Exact matching
mockService.On("GetRecord", context.Background(), int64(1)).Return(record, nil)

// Any argument
mockService.On("GetRecord", mock.Anything, mock.AnythingOfType("int64")).Return(record, nil)

// Custom matching
mockService.On("GetRecord", mock.MatchedBy(func(ctx context.Context) bool {
    return ctx != nil
}), mock.AnythingOfType("int64")).Return(record, nil)
```

### Error Simulation

```go
// Return error
mockService.On("GetRecord", mock.Anything, int64(999)).Return(nil, errors.New("record not found"))

// Panic simulation (for Fatal/Panic methods in logger)
mockLogger := &mocks.MockLogger{}
mockLogger.On("Fatal", mock.Anything).Panic("fatal error")
```

## Best Practices

1. **Always call `AssertExpectations(t)`** at the end of your tests to verify all expected calls were made.

2. **Use `mock.Anything`** for arguments you don't need to specifically match.

3. **Set up expectations before using the mock** in your code under test.

4. **Return appropriate nil values** for pointer types when simulating errors.

5. **Use specific parameter matching** when the behavior depends on specific input values.

## Interface Compliance

All mocks include compile-time interface compliance checks at the bottom of each file to ensure they properly implement their respective interfaces.

## Testing

Run tests for the mocks:

```bash
go test ./src/mocks/ -v
```

The `example_usage_test.go` file contains practical examples of how to use each mock type.
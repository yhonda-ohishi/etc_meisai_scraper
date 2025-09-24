package grpc

import (
	"context"
	"log"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/yhonda-ohishi/etc_meisai/src/grpc"
)

// TestLoggerInterface tests the logger interface and default implementation
func TestLoggerInterface(t *testing.T) {
	t.Run("default logger implementation", func(t *testing.T) {
		var output strings.Builder
		logger := log.New(&output, "[TEST] ", log.LstdFlags)

		// Test Printf
		logger.Printf("Test message: %s", "hello")
		assert.Contains(t, output.String(), "Test message: hello")

		// Reset output
		output.Reset()

		// Test Println
		logger.Println("Test println")
		assert.Contains(t, output.String(), "Test println")

		// Reset output
		output.Reset()

		// Test Print
		logger.Print("Test print")
		assert.Contains(t, output.String(), "Test print")
	})

	t.Run("mock logger interface", func(t *testing.T) {
		mockLogger := &MockLogger{}

		// Test Printf
		mockLogger.On("Printf", "Test: %s", []interface{}{"value"}).Return()
		mockLogger.Printf("Test: %s", "value")
		mockLogger.AssertExpectations(t)

		// Test Println
		mockLogger.On("Println", []interface{}{"test", "message"}).Return()
		mockLogger.Println("test", "message")
		mockLogger.AssertExpectations(t)

		// Test Print
		mockLogger.On("Print", []interface{}{"test"}).Return()
		mockLogger.Print("test")
		mockLogger.AssertExpectations(t)
	})

	t.Run("logger interface compliance", func(t *testing.T) {
		// Verify that the default logger implements the interface
		var logger grpc.LoggerInterface = &defaultLoggerWrapper{
			logger: log.New(os.Stderr, "[TEST] ", log.LstdFlags),
		}

		assert.NotNil(t, logger)

		// Test that interface methods exist and can be called
		assert.NotPanics(t, func() {
			logger.Printf("test %s", "message")
			logger.Println("test", "message")
			logger.Print("test")
		})
	})
}

// defaultLoggerWrapper wraps log.Logger to implement LoggerInterface for testing
type defaultLoggerWrapper struct {
	logger *log.Logger
}

func (l *defaultLoggerWrapper) Printf(format string, v ...interface{}) {
	l.logger.Printf(format, v...)
}

func (l *defaultLoggerWrapper) Println(v ...interface{}) {
	l.logger.Println(v...)
}

func (l *defaultLoggerWrapper) Print(v ...interface{}) {
	l.logger.Print(v...)
}

func (l *defaultLoggerWrapper) Fatalf(format string, v ...interface{}) {
	l.logger.Fatalf(format, v...)
}

func (l *defaultLoggerWrapper) Fatal(v ...interface{}) {
	l.logger.Fatal(v...)
}

func (l *defaultLoggerWrapper) Panicf(format string, v ...interface{}) {
	l.logger.Panicf(format, v...)
}

func (l *defaultLoggerWrapper) Panic(v ...interface{}) {
	l.logger.Panic(v...)
}

// TestServiceInterfaces tests that mock implementations properly implement the interfaces
func TestServiceInterfaces(t *testing.T) {
	ctx := context.Background()

	t.Run("ETCMeisaiServiceInterface compliance", func(t *testing.T) {
		mockService := &MockETCMeisaiService{}

		// Verify interface compliance
		var service grpc.ETCMeisaiServiceInterface = mockService
		assert.NotNil(t, service)

		// Test method signatures exist
		mockService.On("HealthCheck", ctx).Return(nil)
		err := service.HealthCheck(ctx)
		assert.NoError(t, err)
		mockService.AssertExpectations(t)
	})

	t.Run("ETCMappingServiceInterface compliance", func(t *testing.T) {
		mockService := &MockETCMappingService{}

		// Verify interface compliance
		var service grpc.ETCMappingServiceInterface = mockService
		assert.NotNil(t, service)

		// Test method signatures exist
		mockService.On("HealthCheck", ctx).Return(nil)
		err := service.HealthCheck(ctx)
		assert.NoError(t, err)
		mockService.AssertExpectations(t)
	})

	t.Run("ImportServiceInterface compliance", func(t *testing.T) {
		mockService := &MockImportService{}

		// Verify interface compliance
		var service grpc.ImportServiceInterface = mockService
		assert.NotNil(t, service)

		// Test method signatures exist
		mockService.On("HealthCheck", ctx).Return(nil)
		err := service.HealthCheck(ctx)
		assert.NoError(t, err)
		mockService.AssertExpectations(t)
	})

	t.Run("StatisticsServiceInterface compliance", func(t *testing.T) {
		mockService := &MockStatisticsService{}

		// Verify interface compliance
		var service grpc.StatisticsServiceInterface = mockService
		assert.NotNil(t, service)

		// Test method signatures exist
		mockService.On("HealthCheck", ctx).Return(nil)
		err := service.HealthCheck(ctx)
		assert.NoError(t, err)
		mockService.AssertExpectations(t)
	})
}

// TestServerFactory tests the server factory functionality
func TestServerFactory(t *testing.T) {
	t.Run("NewETCMeisaiServerWithConcreteServices", func(t *testing.T) {
		// Note: We can't easily test the concrete factory without importing the actual services
		// This would require the concrete service implementations which may have external dependencies
		// Instead, we test that the function signature and basic functionality work

		// This test verifies that the factory function would work with proper concrete services
		// In a real scenario, you would pass actual service instances
		assert.NotPanics(t, func() {
			// The factory function should exist and be callable
			// We can't actually call it without concrete services, but we can verify its signature
			var factory interface{} = grpc.NewETCMeisaiServerWithConcreteServices
			assert.NotNil(t, factory)
		})
	})

	t.Run("default logger creation", func(t *testing.T) {
		// Test that a logger can be created and used
		logger := log.New(os.Stderr, "[TEST] ", log.LstdFlags)
		assert.NotNil(t, logger)

		// Test that logger methods work
		assert.NotPanics(t, func() {
			logger.Printf("Test message: %s", "test")
		})
	})
}

// TestInterfaceMethodSignatures verifies that all interface methods have correct signatures
func TestInterfaceMethodSignatures(t *testing.T) {
	t.Run("ETCMeisaiServiceInterface methods", func(t *testing.T) {
		mockService := &MockETCMeisaiService{}
		ctx := context.Background()

		// Test that all expected methods exist and can be called
		mockService.On("CreateRecord", ctx, mock.Anything).Return(nil, nil)
		mockService.On("GetRecord", ctx, int64(1)).Return(nil, nil)
		mockService.On("ListRecords", ctx, mock.Anything).Return(nil, nil)
		mockService.On("UpdateRecord", ctx, int64(1), mock.Anything).Return(nil, nil)
		mockService.On("DeleteRecord", ctx, int64(1)).Return(nil)
		mockService.On("HealthCheck", ctx).Return(nil)

		var service grpc.ETCMeisaiServiceInterface = mockService

		// Call all methods to ensure signatures match
		_, _ = service.CreateRecord(ctx, nil)
		_, _ = service.GetRecord(ctx, 1)
		_, _ = service.ListRecords(ctx, nil)
		_, _ = service.UpdateRecord(ctx, 1, nil)
		_ = service.DeleteRecord(ctx, 1)
		_ = service.HealthCheck(ctx)

		mockService.AssertExpectations(t)
	})

	t.Run("ETCMappingServiceInterface methods", func(t *testing.T) {
		mockService := &MockETCMappingService{}
		ctx := context.Background()

		// Test that all expected methods exist and can be called
		mockService.On("CreateMapping", ctx, mock.Anything).Return(nil, nil)
		mockService.On("GetMapping", ctx, int64(1)).Return(nil, nil)
		mockService.On("ListMappings", ctx, mock.Anything).Return(nil, nil)
		mockService.On("UpdateMapping", ctx, int64(1), mock.Anything).Return(nil, nil)
		mockService.On("DeleteMapping", ctx, int64(1)).Return(nil)
		mockService.On("UpdateStatus", ctx, int64(1), "active").Return(nil)
		mockService.On("HealthCheck", ctx).Return(nil)

		var service grpc.ETCMappingServiceInterface = mockService

		// Call all methods to ensure signatures match
		_, _ = service.CreateMapping(ctx, nil)
		_, _ = service.GetMapping(ctx, 1)
		_, _ = service.ListMappings(ctx, nil)
		_, _ = service.UpdateMapping(ctx, 1, nil)
		_ = service.DeleteMapping(ctx, 1)
		_ = service.UpdateStatus(ctx, 1, "active")
		_ = service.HealthCheck(ctx)

		mockService.AssertExpectations(t)
	})

	t.Run("ImportServiceInterface methods", func(t *testing.T) {
		mockService := &MockImportService{}
		ctx := context.Background()

		// Test that all expected methods exist and can be called
		mockService.On("ImportCSV", ctx, mock.Anything, mock.Anything).Return(nil, nil)
		mockService.On("ImportCSVStream", ctx, mock.Anything).Return(nil, nil)
		mockService.On("GetImportSession", ctx, "session-1").Return(nil, nil)
		mockService.On("ListImportSessions", ctx, mock.Anything).Return(nil, nil)
		mockService.On("ProcessCSV", ctx, mock.Anything, mock.Anything).Return(nil, nil)
		mockService.On("ProcessCSVRow", ctx, mock.Anything).Return(nil, nil)
		mockService.On("HandleDuplicates", ctx, mock.Anything).Return(nil, nil)
		mockService.On("CancelImportSession", ctx, "session-1").Return(nil)
		mockService.On("HealthCheck", ctx).Return(nil)

		var service grpc.ImportServiceInterface = mockService

		// Call all methods to ensure signatures match
		_, _ = service.ImportCSV(ctx, nil, nil)
		_, _ = service.ImportCSVStream(ctx, nil)
		_, _ = service.GetImportSession(ctx, "session-1")
		_, _ = service.ListImportSessions(ctx, nil)
		_, _ = service.ProcessCSV(ctx, nil, nil)
		_, _ = service.ProcessCSVRow(ctx, nil)
		_, _ = service.HandleDuplicates(ctx, nil)
		_ = service.CancelImportSession(ctx, "session-1")
		_ = service.HealthCheck(ctx)

		mockService.AssertExpectations(t)
	})

	t.Run("StatisticsServiceInterface methods", func(t *testing.T) {
		mockService := &MockStatisticsService{}
		ctx := context.Background()

		// Test that all expected methods exist and can be called
		mockService.On("GetGeneralStatistics", ctx, mock.Anything).Return(nil, nil)
		mockService.On("GetDailyStatistics", ctx, mock.Anything).Return(nil, nil)
		mockService.On("GetMonthlyStatistics", ctx, mock.Anything).Return(nil, nil)
		mockService.On("GetVehicleStatistics", ctx, mock.Anything, mock.Anything).Return(nil, nil)
		mockService.On("GetMappingStatistics", ctx, mock.Anything).Return(nil, nil)
		mockService.On("HealthCheck", ctx).Return(nil)

		var service grpc.StatisticsServiceInterface = mockService

		// Call all methods to ensure signatures match
		_, _ = service.GetGeneralStatistics(ctx, nil)
		_, _ = service.GetDailyStatistics(ctx, nil)
		_, _ = service.GetMonthlyStatistics(ctx, nil)
		_, _ = service.GetVehicleStatistics(ctx, nil, nil)
		_, _ = service.GetMappingStatistics(ctx, nil)
		_ = service.HealthCheck(ctx)

		mockService.AssertExpectations(t)
	})
}

// TestMockBehavior tests that mocks behave correctly
func TestMockBehavior(t *testing.T) {
	t.Run("mock method call verification", func(t *testing.T) {
		mockService := &MockETCMeisaiService{}
		ctx := context.Background()

		// Set up expectation
		mockService.On("HealthCheck", ctx).Return(nil).Once()

		// Call method
		err := mockService.HealthCheck(ctx)

		// Verify
		assert.NoError(t, err)
		mockService.AssertExpectations(t)
		mockService.AssertNumberOfCalls(t, "HealthCheck", 1)
	})

	t.Run("mock method parameters verification", func(t *testing.T) {
		mockService := &MockETCMeisaiService{}
		ctx := context.Background()

		// Set up expectation with specific parameters
		mockService.On("GetRecord", ctx, int64(123)).Return(createTestRecord(), nil).Once()

		// Call with matching parameters
		record, err := mockService.GetRecord(ctx, 123)

		// Verify
		assert.NoError(t, err)
		assert.NotNil(t, record)
		mockService.AssertExpectations(t)
	})

	t.Run("mock return value verification", func(t *testing.T) {
		mockService := &MockETCMeisaiService{}
		ctx := context.Background()
		expectedRecord := createTestRecord()

		// Set up expectation with return value
		mockService.On("GetRecord", ctx, int64(1)).Return(expectedRecord, nil)

		// Call method
		actualRecord, err := mockService.GetRecord(ctx, 1)

		// Verify return values
		assert.NoError(t, err)
		assert.Equal(t, expectedRecord, actualRecord)
		mockService.AssertExpectations(t)
	})

	t.Run("mock multiple calls", func(t *testing.T) {
		mockService := &MockETCMeisaiService{}
		ctx := context.Background()

		// Set up expectation for multiple calls
		mockService.On("HealthCheck", ctx).Return(nil).Times(3)

		// Call multiple times
		for i := 0; i < 3; i++ {
			err := mockService.HealthCheck(ctx)
			assert.NoError(t, err)
		}

		// Verify all calls were made
		mockService.AssertExpectations(t)
		mockService.AssertNumberOfCalls(t, "HealthCheck", 3)
	})
}

// TestInterfaceDesign tests the overall design of interfaces
func TestInterfaceDesign(t *testing.T) {
	t.Run("interface separation of concerns", func(t *testing.T) {
		// Each service interface should have a clear, focused responsibility

		// ETCMeisaiServiceInterface: ETC record management
		var etcService grpc.ETCMeisaiServiceInterface = &MockETCMeisaiService{}
		assert.NotNil(t, etcService)

		// ETCMappingServiceInterface: ETC mapping management
		var mappingService grpc.ETCMappingServiceInterface = &MockETCMappingService{}
		assert.NotNil(t, mappingService)

		// ImportServiceInterface: Import functionality
		var importService grpc.ImportServiceInterface = &MockImportService{}
		assert.NotNil(t, importService)

		// StatisticsServiceInterface: Statistics and reporting
		var statsService grpc.StatisticsServiceInterface = &MockStatisticsService{}
		assert.NotNil(t, statsService)

		// LoggerInterface: Logging functionality
		var logger grpc.LoggerInterface = &MockLogger{}
		assert.NotNil(t, logger)

		// All interfaces are separate and focused
		assert.IsType(t, &MockETCMeisaiService{}, etcService)
		assert.IsType(t, &MockETCMappingService{}, mappingService)
		assert.IsType(t, &MockImportService{}, importService)
		assert.IsType(t, &MockStatisticsService{}, statsService)
		assert.IsType(t, &MockLogger{}, logger)
	})

	t.Run("interface consistency", func(t *testing.T) {
		// All service interfaces should have HealthCheck method
		ctx := context.Background()

		etcService := &MockETCMeisaiService{}
		etcService.On("HealthCheck", ctx).Return(nil)
		err := etcService.HealthCheck(ctx)
		assert.NoError(t, err)

		mappingService := &MockETCMappingService{}
		mappingService.On("HealthCheck", ctx).Return(nil)
		err = mappingService.HealthCheck(ctx)
		assert.NoError(t, err)

		importService := &MockImportService{}
		importService.On("HealthCheck", ctx).Return(nil)
		err = importService.HealthCheck(ctx)
		assert.NoError(t, err)

		statsService := &MockStatisticsService{}
		statsService.On("HealthCheck", ctx).Return(nil)
		err = statsService.HealthCheck(ctx)
		assert.NoError(t, err)
	})

	t.Run("interface testability", func(t *testing.T) {
		// Interfaces should enable easy mocking and testing

		// Create a server with all mock services
		server, mockETC, mockMapping, mockImport, mockStats, mockLogger := createTestServer()

		// Verify that the server accepts interface implementations
		assert.NotNil(t, server)
		assert.IsType(t, &MockETCMeisaiService{}, mockETC)
		assert.IsType(t, &MockETCMappingService{}, mockMapping)
		assert.IsType(t, &MockImportService{}, mockImport)
		assert.IsType(t, &MockStatisticsService{}, mockStats)
		assert.IsType(t, &MockLogger{}, mockLogger)
	})
}

// TestErrorHandling tests error handling across interfaces
func TestErrorHandling(t *testing.T) {
	ctx := context.Background()

	t.Run("service error propagation", func(t *testing.T) {
		mockService := &MockETCMeisaiService{}
		expectedError := assert.AnError

		// Set up mock to return error
		mockService.On("GetRecord", ctx, int64(1)).Return(nil, expectedError)

		// Call method
		record, err := mockService.GetRecord(ctx, 1)

		// Verify error propagation
		assert.Nil(t, record)
		assert.Equal(t, expectedError, err)
		mockService.AssertExpectations(t)
	})

	t.Run("logger error methods", func(t *testing.T) {
		mockLogger := &MockLogger{}

		// Test fatal and panic methods exist (even though we won't call them in tests)
		assert.NotPanics(t, func() {
			// Verify methods exist by checking their types
			var logger grpc.LoggerInterface = mockLogger
			_ = logger.Fatalf
			_ = logger.Fatal
			_ = logger.Panicf
			_ = logger.Panic
		})
	})
}

// TestInterfaceEvolution tests that interfaces can evolve without breaking existing code
func TestInterfaceEvolution(t *testing.T) {
	t.Run("interface backward compatibility", func(t *testing.T) {
		// This test ensures that our interfaces are designed for evolution
		// New methods can be added to interfaces without breaking existing implementations
		// if they have default implementations or are optional

		// For now, we just verify that all current methods are implemented
		ctx := context.Background()

		// Test that a minimal implementation satisfies the interface
		mockService := &MockETCMeisaiService{}
		mockService.On("HealthCheck", ctx).Return(nil)

		var service grpc.ETCMeisaiServiceInterface = mockService
		err := service.HealthCheck(ctx)
		assert.NoError(t, err)
		mockService.AssertExpectations(t)
	})

	t.Run("interface extensibility", func(t *testing.T) {
		// Interfaces should be extensible through composition or embedding
		// This test verifies that our current design supports this pattern

		// Example: A composite service that implements multiple interfaces
		// We use named fields to avoid method conflicts
		type CompositeService struct {
			ETCService     *MockETCMeisaiService
			MappingService *MockETCMappingService
		}

		// Explicitly implement the interface methods to resolve ambiguity
		composite := &CompositeService{
			ETCService:     &MockETCMeisaiService{},
			MappingService: &MockETCMappingService{},
		}

		// Verify that the individual services implement their interfaces
		var etcService grpc.ETCMeisaiServiceInterface = composite.ETCService
		var mappingService grpc.ETCMappingServiceInterface = composite.MappingService

		assert.NotNil(t, etcService)
		assert.NotNil(t, mappingService)
	})
}
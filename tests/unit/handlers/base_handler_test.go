package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/yhonda-ohishi/etc_meisai/src/handlers"
	"github.com/yhonda-ohishi/etc_meisai/src/services"
)

// Mock implementations
type MockServiceRegistry struct {
	mock.Mock
}

func (m *MockServiceRegistry) GetETCService() *services.ETCService {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*services.ETCService)
}

func (m *MockServiceRegistry) GetMappingService() *services.MappingService {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*services.MappingService)
}

func (m *MockServiceRegistry) GetImportService() *services.ImportServiceLegacy {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*services.ImportServiceLegacy)
}

func (m *MockServiceRegistry) GetDownloadService() services.DownloadServiceInterface {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(services.DownloadServiceInterface)
}

func (m *MockServiceRegistry) GetBaseService() *services.BaseService {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*services.BaseService)
}

func (m *MockServiceRegistry) HealthCheck(ctx context.Context) *services.HealthCheckResult {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil
	}
	// Handle both direct return and function return
	switch v := args.Get(0).(type) {
	case *services.HealthCheckResult:
		return v
	case func(context.Context) *services.HealthCheckResult:
		return v(ctx)
	default:
		return nil
	}
}


type MockBaseService struct {
	mock.Mock
}

func (m *MockBaseService) HealthCheck(ctx context.Context) *services.HealthCheckResult {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*services.HealthCheckResult)
}

// Test helper functions
func createTestLogger() *log.Logger {
	return log.New(os.Stderr, "[TEST] ", log.LstdFlags)
}

func toJSON(t *testing.T, v interface{}) string {
	data, err := json.Marshal(v)
	assert.NoError(t, err)
	return string(data)
}

func createMockServiceRegistry() *MockServiceRegistry {
	return &MockServiceRegistry{}
}

func createTestBaseHandler() (*handlers.BaseHandler, *MockServiceRegistry) {
	mockRegistry := createMockServiceRegistry()
	logger := createTestLogger()
	baseHandler := handlers.NewBaseHandler(mockRegistry, logger)
	return baseHandler, mockRegistry
}

// TestNewBaseHandler tests base handler creation
func TestNewBaseHandler(t *testing.T) {
	t.Run("successful creation", func(t *testing.T) {
		mockRegistry := createMockServiceRegistry()
		logger := createTestLogger()

		handler := handlers.NewBaseHandler(mockRegistry, logger)

		assert.NotNil(t, handler)
		assert.Equal(t, mockRegistry, handler.ServiceRegistry)
		assert.Equal(t, logger, handler.Logger)
		assert.NotNil(t, handler.ErrorHandler)
	})

	t.Run("creation with nil registry", func(t *testing.T) {
		logger := createTestLogger()

		handler := handlers.NewBaseHandler(nil, logger)

		assert.NotNil(t, handler)
		assert.Nil(t, handler.ServiceRegistry)
		assert.Equal(t, logger, handler.Logger)
	})

	t.Run("creation with nil logger", func(t *testing.T) {
		mockRegistry := createMockServiceRegistry()

		handler := handlers.NewBaseHandler(mockRegistry, nil)

		assert.NotNil(t, handler)
		assert.Equal(t, mockRegistry, handler.ServiceRegistry)
		assert.Nil(t, handler.Logger)
	})
}

// TestRespondJSON tests JSON response functionality
func TestRespondJSON(t *testing.T) {
	handler, _ := createTestBaseHandler()

	tests := []struct {
		name           string
		status         int
		data           interface{}
		expectedStatus int
		checkBody      bool
	}{
		{
			name:           "successful response",
			status:         http.StatusOK,
			data:           map[string]string{"message": "success"},
			expectedStatus: http.StatusOK,
			checkBody:      true,
		},
		{
			name:           "error response",
			status:         http.StatusBadRequest,
			data:           map[string]string{"error": "bad request"},
			expectedStatus: http.StatusBadRequest,
			checkBody:      true,
		},
		{
			name:           "nil data",
			status:         http.StatusOK,
			data:           nil,
			expectedStatus: http.StatusOK,
			checkBody:      true,
		},
		{
			name:           "complex data structure",
			status:         http.StatusCreated,
			data:           map[string]interface{}{"count": 10, "items": []string{"a", "b", "c"}},
			expectedStatus: http.StatusCreated,
			checkBody:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()

			handler.RespondJSON(w, tt.status, tt.data)

			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

			if tt.checkBody {
				var responseData interface{}
				err := json.Unmarshal(w.Body.Bytes(), &responseData)
				assert.NoError(t, err)

				if tt.data != nil {
					assert.NotNil(t, responseData)
				}
			}
		})
	}
}

// TestRespondError tests error response functionality
func TestRespondError(t *testing.T) {
	handler, _ := createTestBaseHandler()

	tests := []struct {
		name           string
		status         int
		code           string
		message        string
		details        interface{}
		expectedStatus int
	}{
		{
			name:           "simple error",
			status:         http.StatusBadRequest,
			code:           "INVALID_INPUT",
			message:        "Invalid input provided",
			details:        nil,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "error with details",
			status:         http.StatusUnprocessableEntity,
			code:           "VALIDATION_ERROR",
			message:        "Validation failed",
			details:        map[string]string{"field": "email", "issue": "invalid format"},
			expectedStatus: http.StatusUnprocessableEntity,
		},
		{
			name:           "server error",
			status:         http.StatusInternalServerError,
			code:           "INTERNAL_ERROR",
			message:        "An internal error occurred",
			details:        "Database connection failed",
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()

			handler.RespondError(w, tt.status, tt.code, tt.message, tt.details)

			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

			var errorResponse handlers.ErrorResponse
			err := json.Unmarshal(w.Body.Bytes(), &errorResponse)
			assert.NoError(t, err)

			assert.Equal(t, tt.code, errorResponse.Error.Code)
			assert.Equal(t, tt.message, errorResponse.Error.Message)
			if tt.details != nil {
				// Use JSON comparison for details to handle type conversions
				assert.JSONEq(t, toJSON(t, tt.details), toJSON(t, errorResponse.Error.Details))
			}
		})
	}
}

// TestRespondSuccess tests success response functionality
func TestRespondSuccess(t *testing.T) {
	handler, _ := createTestBaseHandler()

	tests := []struct {
		name    string
		data    interface{}
		message string
	}{
		{
			name:    "simple success",
			data:    map[string]string{"result": "ok"},
			message: "Operation completed successfully",
		},
		{
			name:    "success with nil data",
			data:    nil,
			message: "No data to return",
		},
		{
			name:    "success with complex data",
			data:    map[string]interface{}{"users": []string{"alice", "bob"}, "count": 2},
			message: "Users retrieved",
		},
		{
			name:    "success with empty message",
			data:    map[string]string{"status": "complete"},
			message: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()

			handler.RespondSuccess(w, tt.data, tt.message)

			assert.Equal(t, http.StatusOK, w.Code)
			assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

			var successResponse handlers.SuccessResponse
			err := json.Unmarshal(w.Body.Bytes(), &successResponse)
			assert.NoError(t, err)

			assert.True(t, successResponse.Success)
			assert.Equal(t, tt.message, successResponse.Message)
			if tt.data != nil {
				// Use deep equal for data comparison to handle type conversions
				assert.JSONEq(t, toJSON(t, tt.data), toJSON(t, successResponse.Data))
			}
		})
	}
}

// TestHealthCheck tests health check functionality
func TestHealthCheck(t *testing.T) {
	t.Run("successful health check", func(t *testing.T) {
		handler, mockRegistry := createTestBaseHandler()

		healthResult := &services.HealthCheckResult{
			Status: "healthy",
			Services: map[string]*services.ServiceHealth{
				"etc_service": {Status: "healthy"},
				"db_service":  {Status: "healthy"},
			},
		}
		mockRegistry.On("HealthCheck", mock.AnythingOfType("*context.timerCtx")).Return(healthResult)

		req := httptest.NewRequest("GET", "/health", nil)
		w := httptest.NewRecorder()

		handler.HealthCheck(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.NotNil(t, response)

		mockRegistry.AssertExpectations(t)
	})

	t.Run("unhealthy service", func(t *testing.T) {
		handler, mockRegistry := createTestBaseHandler()

		healthResult := &services.HealthCheckResult{
			Status: "unhealthy",
			Services: map[string]*services.ServiceHealth{
				"etc_service": {Status: "healthy"},
				"db_service":  {Status: "unhealthy", Error: "connection failed"},
			},
		}
		mockRegistry.On("HealthCheck", mock.AnythingOfType("*context.timerCtx")).Return(healthResult)

		req := httptest.NewRequest("GET", "/health", nil)
		w := httptest.NewRecorder()

		handler.HealthCheck(w, req)

		assert.Equal(t, http.StatusServiceUnavailable, w.Code)
		mockRegistry.AssertExpectations(t)
	})

	t.Run("nil service registry", func(t *testing.T) {
		logger := createTestLogger()
		handler := handlers.NewBaseHandler(nil, logger)

		req := httptest.NewRequest("GET", "/health", nil)
		w := httptest.NewRecorder()

		handler.HealthCheck(w, req)

		assert.Equal(t, http.StatusServiceUnavailable, w.Code)

		var errorResponse handlers.ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &errorResponse)
		assert.NoError(t, err)
		assert.Equal(t, "service_unavailable", errorResponse.Error.Code)
	})

	t.Run("health check timeout", func(t *testing.T) {
		handler, mockRegistry := createTestBaseHandler()

		// Simulate a slow health check
		mockRegistry.On("HealthCheck", mock.AnythingOfType("*context.timerCtx")).Return(func(ctx context.Context) *services.HealthCheckResult {
			select {
			case <-time.After(15 * time.Second): // Longer than the 10s timeout
				return &services.HealthCheckResult{}
			case <-ctx.Done():
				return &services.HealthCheckResult{
					Status: "unhealthy",
					Services: map[string]*services.ServiceHealth{
						"timeout_service": {Status: "unhealthy", Error: "timeout"},
					},
				}
			}
		})

		req := httptest.NewRequest("GET", "/health", nil)
		w := httptest.NewRecorder()

		start := time.Now()
		handler.HealthCheck(w, req)
		duration := time.Since(start)

		// Should complete within reasonable time due to timeout
		assert.Less(t, duration, 12*time.Second)
		mockRegistry.AssertExpectations(t)
	})
}

// TestErrorHandling tests various error handling scenarios
func TestErrorHandling(t *testing.T) {
	t.Run("JSON encoding error", func(t *testing.T) {
		handler, _ := createTestBaseHandler()

		// Create data that can't be JSON encoded (contains channels)
		invalidData := map[string]interface{}{
			"channel": make(chan int),
		}

		// Capture log output
		var logBuffer bytes.Buffer
		handler.Logger = log.New(&logBuffer, "[TEST] ", log.LstdFlags)

		w := httptest.NewRecorder()
		handler.RespondJSON(w, http.StatusOK, invalidData)

		// Should still set headers and status even if encoding fails
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

		// Should log the error
		logOutput := logBuffer.String()
		assert.Contains(t, logOutput, "Failed to encode response")
	})

	t.Run("gRPC error handling", func(t *testing.T) {
		handler, _ := createTestBaseHandler()

		// Test with a simulated gRPC error
		grpcError := errors.New("rpc error: code = NotFound desc = record not found")

		w := httptest.NewRecorder()
		handler.RespondGRPCError(w, grpcError, "req-123")

		assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

		var errorResponse handlers.ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &errorResponse)
		assert.NoError(t, err)
		assert.NotEmpty(t, errorResponse.Error.Code)
		assert.NotEmpty(t, errorResponse.Error.Message)
	})

	t.Run("nil error handler", func(t *testing.T) {
		handler, _ := createTestBaseHandler()
		handler.ErrorHandler = nil // Reset to nil

		grpcError := errors.New("test error")

		w := httptest.NewRecorder()
		handler.RespondGRPCError(w, grpcError, "req-123")

		// Should not panic and should create a new error handler
		assert.NotNil(t, handler.ErrorHandler)
	})
}

// TestResponseStructures tests the response structure definitions
func TestResponseStructures(t *testing.T) {
	t.Run("ErrorResponse structure", func(t *testing.T) {
		errorResp := handlers.ErrorResponse{}
		errorResp.Error.Code = "TEST_ERROR"
		errorResp.Error.Message = "Test error message"
		errorResp.Error.Details = map[string]string{"field": "test"}

		jsonData, err := json.Marshal(errorResp)
		assert.NoError(t, err)

		var unmarshaledResp handlers.ErrorResponse
		err = json.Unmarshal(jsonData, &unmarshaledResp)
		assert.NoError(t, err)
		assert.Equal(t, "TEST_ERROR", unmarshaledResp.Error.Code)
		assert.Equal(t, "Test error message", unmarshaledResp.Error.Message)
	})

	t.Run("SuccessResponse structure", func(t *testing.T) {
		successResp := handlers.SuccessResponse{
			Success: true,
			Data:    map[string]string{"result": "test"},
			Message: "Test success message",
		}

		jsonData, err := json.Marshal(successResp)
		assert.NoError(t, err)

		var unmarshaledResp handlers.SuccessResponse
		err = json.Unmarshal(jsonData, &unmarshaledResp)
		assert.NoError(t, err)
		assert.True(t, unmarshaledResp.Success)
		assert.Equal(t, "Test success message", unmarshaledResp.Message)
	})
}

// TestConcurrency tests concurrent access to handler methods
func TestConcurrency(t *testing.T) {
	t.Run("concurrent health checks", func(t *testing.T) {
		handler, mockRegistry := createTestBaseHandler()

		healthResult := &services.HealthCheckResult{
			Status: "healthy",
			Services: map[string]*services.ServiceHealth{
				"test_service": {Status: "healthy"},
			},
		}
		mockRegistry.On("HealthCheck", mock.AnythingOfType("*context.timerCtx")).Return(healthResult)

		const numRequests = 10
		done := make(chan bool, numRequests)

		// Launch concurrent requests
		for i := 0; i < numRequests; i++ {
			go func() {
				req := httptest.NewRequest("GET", "/health", nil)
				w := httptest.NewRecorder()
				handler.HealthCheck(w, req)
				assert.Equal(t, http.StatusOK, w.Code)
				done <- true
			}()
		}

		// Wait for all requests to complete
		for i := 0; i < numRequests; i++ {
			<-done
		}

		mockRegistry.AssertExpectations(t)
	})

	t.Run("concurrent JSON responses", func(t *testing.T) {
		handler, _ := createTestBaseHandler()

		const numRequests = 10
		done := make(chan bool, numRequests)

		// Launch concurrent response generations
		for i := 0; i < numRequests; i++ {
			go func(id int) {
				w := httptest.NewRecorder()
				data := map[string]int{"id": id}
				handler.RespondJSON(w, http.StatusOK, data)
				assert.Equal(t, http.StatusOK, w.Code)
				done <- true
			}(i)
		}

		// Wait for all requests to complete
		for i := 0; i < numRequests; i++ {
			<-done
		}
	})
}

// TestEdgeCases tests edge cases and boundary conditions
func TestEdgeCases(t *testing.T) {
	t.Run("very large response data", func(t *testing.T) {
		handler, _ := createTestBaseHandler()

		// Create large data structure
		largeData := make(map[string]string)
		for i := 0; i < 1000; i++ {
			largeData[strings.Repeat("key", i)] = strings.Repeat("value", i)
		}

		w := httptest.NewRecorder()
		start := time.Now()
		handler.RespondJSON(w, http.StatusOK, largeData)
		duration := time.Since(start)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Less(t, duration, 1*time.Second, "Large response should be processed quickly")
	})

	t.Run("unicode in responses", func(t *testing.T) {
		handler, _ := createTestBaseHandler()

		unicodeData := map[string]string{
			"japanese": "こんにちは",
			"emoji":    "🚗🛣️",
			"chinese":  "你好",
		}

		w := httptest.NewRecorder()
		handler.RespondJSON(w, http.StatusOK, unicodeData)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]string
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "こんにちは", response["japanese"])
		assert.Equal(t, "🚗🛣️", response["emoji"])
	})

	t.Run("nil logger handling", func(t *testing.T) {
		mockRegistry := createMockServiceRegistry()
		handler := handlers.NewBaseHandler(mockRegistry, nil)

		// Should not panic when logger is nil
		w := httptest.NewRecorder()
		assert.NotPanics(t, func() {
			handler.RespondJSON(w, http.StatusOK, map[string]string{"test": "data"})
		})
	})

	t.Run("empty error details", func(t *testing.T) {
		handler, _ := createTestBaseHandler()

		w := httptest.NewRecorder()
		handler.RespondError(w, http.StatusBadRequest, "", "", nil)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var errorResponse handlers.ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &errorResponse)
		assert.NoError(t, err)
		assert.Equal(t, "", errorResponse.Error.Code)
		assert.Equal(t, "", errorResponse.Error.Message)
	})
}

// TestPerformance tests performance characteristics
func TestPerformance(t *testing.T) {
	t.Run("response generation performance", func(t *testing.T) {
		handler, _ := createTestBaseHandler()

		data := map[string]interface{}{
			"message": "test",
			"count":   100,
			"items":   make([]string, 100),
		}

		// Warm up
		for i := 0; i < 10; i++ {
			w := httptest.NewRecorder()
			handler.RespondJSON(w, http.StatusOK, data)
		}

		// Measure performance
		const iterations = 1000
		start := time.Now()

		for i := 0; i < iterations; i++ {
			w := httptest.NewRecorder()
			handler.RespondJSON(w, http.StatusOK, data)
		}

		duration := time.Since(start)
		avgDuration := duration / iterations

		t.Logf("Average response time: %v", avgDuration)
		assert.Less(t, avgDuration, 1*time.Millisecond, "Response generation should be fast")
	})

	t.Run("health check performance", func(t *testing.T) {
		handler, mockRegistry := createTestBaseHandler()

		healthResult := &services.HealthCheckResult{
			Status: "healthy",
			Services: map[string]*services.ServiceHealth{
				"service1": {Status: "healthy"},
				"service2": {Status: "healthy"},
				"service3": {Status: "healthy"},
			},
		}
		mockRegistry.On("HealthCheck", mock.AnythingOfType("*context.timerCtx")).Return(healthResult)

		// Measure health check performance
		start := time.Now()
		req := httptest.NewRequest("GET", "/health", nil)
		w := httptest.NewRecorder()
		handler.HealthCheck(w, req)
		duration := time.Since(start)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Less(t, duration, 100*time.Millisecond, "Health check should be fast")
		mockRegistry.AssertExpectations(t)
	})
}
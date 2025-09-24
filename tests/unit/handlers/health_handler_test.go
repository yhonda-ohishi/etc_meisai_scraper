package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yhonda-ohishi/etc_meisai/src/handlers"
)

// TestNewHealthHandler tests health handler creation
func TestNewHealthHandler(t *testing.T) {
	t.Run("creation with nil registry", func(t *testing.T) {
		logger := createTestLogger()

		handler := handlers.NewHealthHandler(nil, logger)

		assert.NotNil(t, handler)
		assert.NotNil(t, handler.BaseHandler)
	})
}

// TestHealthHandlerHealthCheck tests the main health check endpoint
func TestHealthHandlerHealthCheck(t *testing.T) {
	t.Run("service unavailable", func(t *testing.T) {
		mockRegistry := createMockServiceRegistry()
		logger := createTestLogger()

		// Mock returns nil to trigger service unavailable path
		mockRegistry.On("GetBaseService").Return(nil)

		baseHandler := handlers.NewBaseHandler(mockRegistry, logger)
		healthHandler := &handlers.HealthHandler{BaseHandler: baseHandler}

		req := httptest.NewRequest("GET", "/api/health", nil)
		w := httptest.NewRecorder()

		healthHandler.HealthCheck(w, req)

		assert.Equal(t, http.StatusServiceUnavailable, w.Code)
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

		var response handlers.ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "service_unavailable", response.Error.Code)

		mockRegistry.AssertExpectations(t)
	})

	t.Run("nil base service", func(t *testing.T) {
		mockRegistry := createMockServiceRegistry()
		logger := createTestLogger()

		mockRegistry.On("GetBaseService").Return(nil)

		baseHandler := handlers.NewBaseHandler(mockRegistry, logger)
		healthHandler := &handlers.HealthHandler{BaseHandler: baseHandler}

		req := httptest.NewRequest("GET", "/api/health", nil)
		w := httptest.NewRecorder()

		healthHandler.HealthCheck(w, req)

		assert.Equal(t, http.StatusServiceUnavailable, w.Code)

		var errorResponse handlers.ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &errorResponse)
		assert.NoError(t, err)
		assert.Equal(t, "service_unavailable", errorResponse.Error.Code)

		mockRegistry.AssertExpectations(t)
	})

	t.Run("nil service registry", func(t *testing.T) {
		logger := createTestLogger()
		baseHandler := handlers.NewBaseHandler(nil, logger)
		healthHandler := &handlers.HealthHandler{BaseHandler: baseHandler}

		req := httptest.NewRequest("GET", "/api/health", nil)
		w := httptest.NewRecorder()

		healthHandler.HealthCheck(w, req)

		assert.Equal(t, http.StatusServiceUnavailable, w.Code)

		var errorResponse handlers.ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &errorResponse)
		assert.NoError(t, err)
		assert.Equal(t, "service_unavailable", errorResponse.Error.Code)
		assert.Contains(t, errorResponse.Error.Message, "Service registry not initialized")
	})
}

// TestLiveness tests the liveness probe endpoint
func TestLiveness(t *testing.T) {
	t.Run("successful liveness check", func(t *testing.T) {
		mockRegistry := createMockServiceRegistry()
		logger := createTestLogger()

		baseHandler := handlers.NewBaseHandler(mockRegistry, logger)
		healthHandler := &handlers.HealthHandler{BaseHandler: baseHandler}

		req := httptest.NewRequest("GET", "/api/health/live", nil)
		w := httptest.NewRecorder()

		healthHandler.Liveness(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "alive", response["status"])
	})
}

// TestReadiness tests the readiness probe endpoint
func TestReadiness(t *testing.T) {
	t.Run("nil base service", func(t *testing.T) {
		mockRegistry := createMockServiceRegistry()
		logger := createTestLogger()

		mockRegistry.On("GetBaseService").Return(nil)

		baseHandler := handlers.NewBaseHandler(mockRegistry, logger)
		healthHandler := &handlers.HealthHandler{BaseHandler: baseHandler}

		req := httptest.NewRequest("GET", "/api/health/ready", nil)
		w := httptest.NewRecorder()

		healthHandler.Readiness(w, req)

		assert.Equal(t, http.StatusServiceUnavailable, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "not_ready", response["status"])
		assert.Contains(t, response["message"], "Services not initialized")

		mockRegistry.AssertExpectations(t)
	})

	t.Run("nil service registry", func(t *testing.T) {
		logger := createTestLogger()
		baseHandler := handlers.NewBaseHandler(nil, logger)
		healthHandler := &handlers.HealthHandler{BaseHandler: baseHandler}

		req := httptest.NewRequest("GET", "/api/health/ready", nil)
		w := httptest.NewRecorder()

		healthHandler.Readiness(w, req)

		assert.Equal(t, http.StatusServiceUnavailable, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "not_ready", response["status"])
		assert.Contains(t, response["message"], "Service registry not initialized")
	})
}

// TestDeepHealthCheck tests the deep health check endpoint
func TestDeepHealthCheck(t *testing.T) {
	t.Run("nil services", func(t *testing.T) {
		mockRegistry := createMockServiceRegistry()
		logger := createTestLogger()

		// Mock all services to return nil
		mockRegistry.On("GetETCService").Return(nil)
		mockRegistry.On("GetMappingService").Return(nil)
		mockRegistry.On("GetImportService").Return(nil)

		baseHandler := handlers.NewBaseHandler(mockRegistry, logger)
		healthHandler := &handlers.HealthHandler{BaseHandler: baseHandler}

		req := httptest.NewRequest("GET", "/api/health/deep", nil)
		w := httptest.NewRecorder()

		healthHandler.DeepHealthCheck(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response handlers.SuccessResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.True(t, response.Success)
		assert.Contains(t, response.Message, "Deep health check completed")

		// Check that response data contains expected fields
		data := response.Data.(map[string]interface{})
		assert.Contains(t, data, "system")
		assert.Contains(t, data, "timestamp")
		assert.Contains(t, data, "uptime")

		mockRegistry.AssertExpectations(t)
	})
}
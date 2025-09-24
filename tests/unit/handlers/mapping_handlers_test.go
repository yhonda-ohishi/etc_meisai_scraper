package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yhonda-ohishi/etc_meisai/src/handlers"
	"github.com/yhonda-ohishi/etc_meisai/src/models"
)

// createTestMappingHandler creates a test handler with mock service
func createTestMappingHandler() *handlers.MappingHandler {
	mockRegistry := createMockServiceRegistry()
	logger := createTestLogger()

	mockRegistry.On("GetMappingService").Return(nil)

	handler := &handlers.MappingHandler{
		BaseHandler: handlers.BaseHandler{
			ServiceRegistry: mockRegistry,
			Logger:         logger,
		},
	}
	return handler
}

// TestMappingHandlerGetMappings tests the GetMappings endpoint
func TestMappingHandlerGetMappings(t *testing.T) {
	t.Run("service unavailable", func(t *testing.T) {
		handler := createTestMappingHandler()

		req := httptest.NewRequest("GET", "/api/mappings", nil)
		w := httptest.NewRecorder()

		handler.GetMappings(w, req)

		assert.Equal(t, http.StatusServiceUnavailable, w.Code)
	})
}

// TestMappingHandlerCreateMapping tests the CreateMapping endpoint
func TestMappingHandlerCreateMapping(t *testing.T) {
	t.Run("invalid request body", func(t *testing.T) {
		handler := createTestMappingHandler()

		req := httptest.NewRequest("POST", "/api/mappings", strings.NewReader("invalid json"))
		w := httptest.NewRecorder()

		handler.CreateMapping(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("validation error - missing etc_meisai_id", func(t *testing.T) {
		handler := createTestMappingHandler()

		mapping := models.ETCMeisaiMapping{
			DTakoRowID:  "dtako-456",
			MappingType: "manual",
		}

		reqBody, _ := json.Marshal(mapping)
		req := httptest.NewRequest("POST", "/api/mappings", bytes.NewReader(reqBody))
		w := httptest.NewRecorder()

		handler.CreateMapping(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("validation error - missing dtako_row_id", func(t *testing.T) {
		handler := createTestMappingHandler()

		mapping := models.ETCMeisaiMapping{
			ETCMeisaiID: 1,
			MappingType: "manual",
		}

		reqBody, _ := json.Marshal(mapping)
		req := httptest.NewRequest("POST", "/api/mappings", bytes.NewReader(reqBody))
		w := httptest.NewRecorder()

		handler.CreateMapping(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("service unavailable", func(t *testing.T) {
		handler := createTestMappingHandler()

		mapping := models.ETCMeisaiMapping{
			ETCMeisaiID: 1,
			DTakoRowID:  "dtako-789",
		}

		reqBody, _ := json.Marshal(mapping)
		req := httptest.NewRequest("POST", "/api/mappings", bytes.NewReader(reqBody))
		w := httptest.NewRecorder()

		handler.CreateMapping(w, req)

		assert.Equal(t, http.StatusServiceUnavailable, w.Code)
	})
}

// TestMappingHandlerDeleteMapping tests the DeleteMapping endpoint
func TestMappingHandlerDeleteMapping(t *testing.T) {
	t.Run("missing id parameter", func(t *testing.T) {
		handler := createTestMappingHandler()

		req := httptest.NewRequest("DELETE", "/api/mappings/", nil)
		w := httptest.NewRecorder()

		handler.DeleteMapping(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("service unavailable", func(t *testing.T) {
		handler := createTestMappingHandler()

		req := httptest.NewRequest("DELETE", "/api/mappings/1", nil)
		w := httptest.NewRecorder()

		handler.DeleteMapping(w, req)

		assert.Equal(t, http.StatusServiceUnavailable, w.Code)
	})
}

// TestMappingHandlerUpdateMapping tests the UpdateMapping endpoint
func TestMappingHandlerUpdateMapping(t *testing.T) {
	t.Run("missing id parameter", func(t *testing.T) {
		handler := createTestMappingHandler()

		mapping := models.ETCMeisaiMapping{}
		reqBody, _ := json.Marshal(mapping)
		req := httptest.NewRequest("PUT", "/api/mappings/", bytes.NewReader(reqBody))
		w := httptest.NewRecorder()

		handler.UpdateMapping(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("invalid request body", func(t *testing.T) {
		handler := createTestMappingHandler()

		req := httptest.NewRequest("PUT", "/api/mappings/1", strings.NewReader("invalid json"))
		w := httptest.NewRecorder()

		handler.UpdateMapping(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("service unavailable", func(t *testing.T) {
		handler := createTestMappingHandler()

		mapping := models.ETCMeisaiMapping{
			ID: 1,
		}

		reqBody, _ := json.Marshal(mapping)
		req := httptest.NewRequest("PUT", "/api/mappings/1", bytes.NewReader(reqBody))
		w := httptest.NewRecorder()

		handler.UpdateMapping(w, req)

		assert.Equal(t, http.StatusServiceUnavailable, w.Code)
	})
}

// TestMappingHandlerAutoMatch tests the AutoMatch endpoint
func TestMappingHandlerAutoMatch(t *testing.T) {
	t.Run("missing required parameters", func(t *testing.T) {
		handler := createTestMappingHandler()

		req := httptest.NewRequest("POST", "/api/mappings/auto-match", nil)
		w := httptest.NewRecorder()

		handler.AutoMatch(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("invalid date format", func(t *testing.T) {
		handler := createTestMappingHandler()

		req := httptest.NewRequest("POST", "/api/mappings/auto-match?from_date=invalid&to_date=2023-01-31", nil)
		w := httptest.NewRecorder()

		handler.AutoMatch(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("service unavailable", func(t *testing.T) {
		handler := createTestMappingHandler()

		req := httptest.NewRequest("POST", "/api/mappings/auto-match?from_date=2023-01-01&to_date=2023-01-31", nil)
		w := httptest.NewRecorder()

		handler.AutoMatch(w, req)

		assert.Equal(t, http.StatusServiceUnavailable, w.Code)
	})
}
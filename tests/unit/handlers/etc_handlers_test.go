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

// createTestETCHandler creates a test handler with mock service
func createTestETCHandler() *handlers.ETCHandler {
	mockRegistry := createMockServiceRegistry()
	logger := createTestLogger()

	mockRegistry.On("GetETCService").Return(nil)

	handler := &handlers.ETCHandler{
		BaseHandler: handlers.NewBaseHandler(mockRegistry, logger),
	}
	return handler
}

// TestETCHandlerImportData tests the ImportData endpoint
func TestETCHandlerImportData(t *testing.T) {
	t.Run("service unavailable", func(t *testing.T) {
		handler := createTestETCHandler()

		importReq := models.ETCImportRequest{
			FromDate: "2023-01-01",
			ToDate:   "2023-01-31",
			Source:   "csv",
		}

		reqBody, _ := json.Marshal(importReq)
		req := httptest.NewRequest("POST", "/api/etc/import", bytes.NewReader(reqBody))
		w := httptest.NewRecorder()

		handler.ImportData(w, req)

		assert.Equal(t, http.StatusServiceUnavailable, w.Code)

		var response handlers.ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "service_unavailable", response.Error.Code)
	})

	t.Run("invalid request body", func(t *testing.T) {
		handler := createTestETCHandler()

		req := httptest.NewRequest("POST", "/api/etc/import", strings.NewReader("invalid json"))
		w := httptest.NewRecorder()

		handler.ImportData(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response handlers.ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "invalid_request", response.Error.Code)
	})
}

// TestETCHandlerGetMeisai tests the GetMeisai endpoint
func TestETCHandlerGetMeisai(t *testing.T) {
	t.Run("missing parameters", func(t *testing.T) {
		handler := createTestETCHandler()

		req := httptest.NewRequest("GET", "/api/etc/meisai", nil)
		w := httptest.NewRecorder()

		handler.GetMeisai(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response handlers.ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "missing_parameters", response.Error.Code)
	})

	t.Run("missing from_date only", func(t *testing.T) {
		handler := createTestETCHandler()

		req := httptest.NewRequest("GET", "/api/etc/meisai?to_date=2023-01-31", nil)
		w := httptest.NewRecorder()

		handler.GetMeisai(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("service unavailable", func(t *testing.T) {
		handler := createTestETCHandler()

		req := httptest.NewRequest("GET", "/api/etc/meisai?from_date=2023-01-01&to_date=2023-01-31", nil)
		w := httptest.NewRecorder()

		handler.GetMeisai(w, req)

		assert.Equal(t, http.StatusServiceUnavailable, w.Code)
	})
}

// TestETCHandlerGetMeisaiByID tests the GetMeisaiByID endpoint
func TestETCHandlerGetMeisaiByID(t *testing.T) {
	t.Run("missing id parameter", func(t *testing.T) {
		handler := createTestETCHandler()

		req := httptest.NewRequest("GET", "/api/etc/meisai/", nil)
		w := httptest.NewRecorder()

		handler.GetMeisaiByID(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("service unavailable", func(t *testing.T) {
		handler := createTestETCHandler()

		req := httptest.NewRequest("GET", "/api/etc/meisai/1", nil)
		w := httptest.NewRecorder()

		handler.GetMeisaiByID(w, req)

		assert.Equal(t, http.StatusServiceUnavailable, w.Code)
	})
}

// TestETCHandlerCreateMeisai tests the CreateMeisai endpoint
func TestETCHandlerCreateMeisai(t *testing.T) {
	t.Run("invalid request body", func(t *testing.T) {
		handler := createTestETCHandler()

		req := httptest.NewRequest("POST", "/api/etc/meisai", strings.NewReader("invalid json"))
		w := httptest.NewRecorder()

		handler.CreateMeisai(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("service unavailable", func(t *testing.T) {
		handler := createTestETCHandler()

		meisai := models.ETCMeisai{
			ETCNumber: "1234567890",
			Amount:    1000,
		}

		reqBody, _ := json.Marshal(meisai)
		req := httptest.NewRequest("POST", "/api/etc/meisai", bytes.NewReader(reqBody))
		w := httptest.NewRecorder()

		handler.CreateMeisai(w, req)

		assert.Equal(t, http.StatusServiceUnavailable, w.Code)
	})
}

// TestETCHandlerListETCMeisai tests the ListETCMeisai endpoint
func TestETCHandlerListETCMeisai(t *testing.T) {
	t.Run("service unavailable", func(t *testing.T) {
		handler := createTestETCHandler()

		req := httptest.NewRequest("GET", "/api/etc", nil)
		w := httptest.NewRecorder()

		handler.ListETCMeisai(w, req)

		assert.Equal(t, http.StatusServiceUnavailable, w.Code)
	})
}

// TestETCHandlerUpdateETCMeisai tests the UpdateETCMeisai endpoint
func TestETCHandlerUpdateETCMeisai(t *testing.T) {
	t.Run("not implemented", func(t *testing.T) {
		handler := createTestETCHandler()

		updates := map[string]interface{}{
			"amount": 1500,
		}

		reqBody, _ := json.Marshal(updates)
		req := httptest.NewRequest("PUT", "/api/etc/1", bytes.NewReader(reqBody))
		w := httptest.NewRecorder()

		handler.UpdateETCMeisai(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code) // Missing ID parameter
	})

	t.Run("invalid request body", func(t *testing.T) {
		handler := createTestETCHandler()

		req := httptest.NewRequest("PUT", "/api/etc/1", strings.NewReader("invalid json"))
		w := httptest.NewRecorder()

		handler.UpdateETCMeisai(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

// TestETCHandlerDeleteETCMeisai tests the DeleteETCMeisai endpoint
func TestETCHandlerDeleteETCMeisai(t *testing.T) {
	t.Run("not implemented", func(t *testing.T) {
		handler := createTestETCHandler()

		req := httptest.NewRequest("DELETE", "/api/etc/1", nil)
		w := httptest.NewRecorder()

		handler.DeleteETCMeisai(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code) // Missing ID parameter
	})
}

// TestETCHandlerBulkCreateETCMeisai tests the BulkCreateETCMeisai endpoint
func TestETCHandlerBulkCreateETCMeisai(t *testing.T) {
	t.Run("empty records", func(t *testing.T) {
		handler := createTestETCHandler()

		var records []*models.ETCMeisai

		reqBody, _ := json.Marshal(records)
		req := httptest.NewRequest("POST", "/api/etc/bulk", bytes.NewReader(reqBody))
		w := httptest.NewRecorder()

		handler.BulkCreateETCMeisai(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response handlers.ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "empty_request", response.Error.Code)
	})

	t.Run("invalid request body", func(t *testing.T) {
		handler := createTestETCHandler()

		req := httptest.NewRequest("POST", "/api/etc/bulk", strings.NewReader("invalid json"))
		w := httptest.NewRecorder()

		handler.BulkCreateETCMeisai(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("service unavailable", func(t *testing.T) {
		handler := createTestETCHandler()

		records := []*models.ETCMeisai{
			{
				ETCNumber: "1234567890",
				Amount:    1000,
			},
		}

		reqBody, _ := json.Marshal(records)
		req := httptest.NewRequest("POST", "/api/etc/bulk", bytes.NewReader(reqBody))
		w := httptest.NewRecorder()

		handler.BulkCreateETCMeisai(w, req)

		assert.Equal(t, http.StatusServiceUnavailable, w.Code)
	})
}

// TestETCHandlerGetETCSummary tests the GetETCSummary endpoint
func TestETCHandlerGetETCSummary(t *testing.T) {
	t.Run("missing parameters", func(t *testing.T) {
		handler := createTestETCHandler()

		req := httptest.NewRequest("GET", "/api/etc/summary", nil)
		w := httptest.NewRecorder()

		handler.GetETCSummary(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response handlers.ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "missing_parameters", response.Error.Code)
	})

	t.Run("service unavailable", func(t *testing.T) {
		handler := createTestETCHandler()

		req := httptest.NewRequest("GET", "/api/etc/summary?from_date=2023-01-01&to_date=2023-01-31", nil)
		w := httptest.NewRecorder()

		handler.GetETCSummary(w, req)

		assert.Equal(t, http.StatusServiceUnavailable, w.Code)
	})
}

// TestETCHandlerGetSummary tests the alternative GetSummary endpoint
func TestETCHandlerGetSummary(t *testing.T) {
	t.Run("invalid from_date format", func(t *testing.T) {
		handler := createTestETCHandler()

		req := httptest.NewRequest("GET", "/api/etc/summary?from_date=invalid&to_date=2023-01-31", nil)
		w := httptest.NewRecorder()

		handler.GetSummary(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response handlers.ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "invalid_date", response.Error.Code)
	})

	t.Run("invalid to_date format", func(t *testing.T) {
		handler := createTestETCHandler()

		req := httptest.NewRequest("GET", "/api/etc/summary?from_date=2023-01-01&to_date=invalid", nil)
		w := httptest.NewRecorder()

		handler.GetSummary(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response handlers.ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "invalid_date", response.Error.Code)
	})

	t.Run("service unavailable", func(t *testing.T) {
		handler := createTestETCHandler()

		req := httptest.NewRequest("GET", "/api/etc/summary?from_date=2023-01-01&to_date=2023-01-31", nil)
		w := httptest.NewRecorder()

		handler.GetSummary(w, req)

		assert.Equal(t, http.StatusServiceUnavailable, w.Code)
	})
}

// TestETCHandlerBulkImport tests the BulkImport endpoint
func TestETCHandlerBulkImport(t *testing.T) {
	t.Run("service unavailable", func(t *testing.T) {
		handler := createTestETCHandler()

		records := []*models.ETCMeisai{
			{
				ETCNumber: "1234567890",
				Amount:    1000,
			},
		}

		reqBody, _ := json.Marshal(records)
		req := httptest.NewRequest("POST", "/api/etc/bulk-import", bytes.NewReader(reqBody))
		w := httptest.NewRecorder()

		handler.BulkImport(w, req)

		assert.Equal(t, http.StatusServiceUnavailable, w.Code)
	})

	t.Run("invalid request body", func(t *testing.T) {
		handler := createTestETCHandler()

		req := httptest.NewRequest("POST", "/api/etc/bulk-import", strings.NewReader("invalid json"))
		w := httptest.NewRecorder()

		handler.BulkImport(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}
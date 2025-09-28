package handlers_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/yhonda-ohishi/etc_meisai/src/handlers"
	"github.com/yhonda-ohishi/etc_meisai/src/services"
)

// CompleteMockDownloadService provides complete mock implementation
type CompleteMockDownloadService struct {
	accountIDs       []string
	jobs             map[string]*services.DownloadJob
	processAsyncFunc func(jobID string, accounts []string, fromDate, toDate string)
}

func NewCompleteMockDownloadService() *CompleteMockDownloadService {
	return &CompleteMockDownloadService{
		accountIDs: []string{"test1", "test2"},
		jobs:       make(map[string]*services.DownloadJob),
	}
}

func (m *CompleteMockDownloadService) GetAllAccountIDs() []string {
	return m.accountIDs
}

func (m *CompleteMockDownloadService) ProcessAsync(jobID string, accounts []string, fromDate, toDate string) {
	if m.processAsyncFunc != nil {
		m.processAsyncFunc(jobID, accounts, fromDate, toDate)
	}

	now := time.Now()
	m.jobs[jobID] = &services.DownloadJob{
		ID:           jobID,
		Status:       "processing",
		Progress:     0,
		TotalRecords: 100,
		StartedAt:    now,
	}
}

func (m *CompleteMockDownloadService) GetJobStatus(jobID string) (*services.DownloadJob, bool) {
	job, exists := m.jobs[jobID]
	if !exists {
		// Check some predefined jobs
		if jobID == "completed-job" {
			completedAt := time.Now()
			return &services.DownloadJob{
				ID:           jobID,
				Status:       "completed",
				Progress:     100,
				TotalRecords: 50,
				StartedAt:    time.Now().Add(-1 * time.Hour),
				CompletedAt:  &completedAt,
			}, true
		}
		if jobID == "failed-job" {
			completedAt := time.Now()
			return &services.DownloadJob{
				ID:           jobID,
				Status:       "failed",
				Progress:     50,
				TotalRecords: 0,
				ErrorMessage: "Download failed",
				StartedAt:    time.Now().Add(-30 * time.Minute),
				CompletedAt:  &completedAt,
			}, true
		}
	}
	return job, exists
}

func TestDownloadHandler_DownloadSync_AllCases(t *testing.T) {
	tests := []struct {
		name           string
		reqBody        interface{}
		expectedStatus int
		checkResponse  func(t *testing.T, resp map[string]interface{})
	}{
		{
			name: "valid request with dates",
			reqBody: handlers.DownloadRequest{
				Accounts: []string{"test1"},
				FromDate: "2024-01-01",
				ToDate:   "2024-01-31",
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, resp map[string]interface{}) {
				if resp["success"] != true {
					t.Error("Expected success to be true")
				}
			},
		},
		{
			name: "request without dates (should use defaults)",
			reqBody: handlers.DownloadRequest{
				Accounts: []string{"test1"},
				FromDate: "",
				ToDate:   "",
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, resp map[string]interface{}) {
				if resp["success"] != true {
					t.Error("Expected success to be true")
				}
			},
		},
		{
			name: "request without accounts",
			reqBody: handlers.DownloadRequest{
				Accounts: []string{},
				FromDate: "2024-01-01",
				ToDate:   "2024-01-31",
			},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, resp map[string]interface{}) {
				if resp["error"] == nil {
					t.Error("Expected error message")
				}
			},
		},
		{
			name:           "invalid JSON",
			reqBody:        "invalid json",
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, resp map[string]interface{}) {
				if resp["error"] == nil {
					t.Error("Expected error message")
				}
			},
		},
		{
			name: "only from date missing",
			reqBody: handlers.DownloadRequest{
				Accounts: []string{"test1"},
				FromDate: "",
				ToDate:   "2024-01-31",
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "only to date missing",
			reqBody: handlers.DownloadRequest{
				Accounts: []string{"test1"},
				FromDate: "2024-01-01",
				ToDate:   "",
			},
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := NewCompleteMockDownloadService()
			handler := handlers.NewDownloadHandler(mockService)

			var body []byte
			if str, ok := tt.reqBody.(string); ok {
				body = []byte(str)
			} else {
				body, _ = json.Marshal(tt.reqBody)
			}

			req := httptest.NewRequest("POST", "/api/download/sync", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			handler.DownloadSync(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.checkResponse != nil {
				var response map[string]interface{}
				json.Unmarshal(w.Body.Bytes(), &response)
				tt.checkResponse(t, response)
			}
		})
	}
}

func TestDownloadHandler_DownloadAsync_AllCases(t *testing.T) {
	tests := []struct {
		name           string
		reqBody        interface{}
		mockAccounts   []string
		expectedStatus int
		checkResponse  func(t *testing.T, resp map[string]interface{})
	}{
		{
			name: "with specific accounts",
			reqBody: handlers.DownloadRequest{
				Accounts: []string{"test1", "test2"},
				FromDate: "2024-01-01",
				ToDate:   "2024-01-31",
			},
			mockAccounts:   []string{"test1", "test2"},
			expectedStatus: http.StatusAccepted,
			checkResponse: func(t *testing.T, resp map[string]interface{}) {
				if resp["job_id"] == nil || resp["job_id"] == "" {
					t.Error("Expected job_id")
				}
				if resp["status"] != "pending" {
					t.Error("Expected status to be 'pending'")
				}
			},
		},
		{
			name: "without accounts (use all)",
			reqBody: handlers.DownloadRequest{
				Accounts: []string{},
			},
			mockAccounts:   []string{"test1", "test2"},
			expectedStatus: http.StatusAccepted,
		},
		{
			name: "without accounts and no configured accounts",
			reqBody: handlers.DownloadRequest{
				Accounts: []string{},
			},
			mockAccounts:   []string{},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, resp map[string]interface{}) {
				if resp["error"] == nil {
					t.Error("Expected error message")
				}
			},
		},
		{
			name:           "invalid JSON",
			reqBody:        "{invalid}",
			mockAccounts:   []string{"test1"},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "with mode parameter",
			reqBody: handlers.DownloadRequest{
				Accounts: []string{"test1"},
				FromDate: "2024-01-01",
				ToDate:   "2024-01-31",
				Mode:     "fast",
			},
			mockAccounts:   []string{"test1"},
			expectedStatus: http.StatusAccepted,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := NewCompleteMockDownloadService()
			mockService.accountIDs = tt.mockAccounts
			handler := handlers.NewDownloadHandler(mockService)

			var body []byte
			if str, ok := tt.reqBody.(string); ok {
				body = []byte(str)
			} else {
				body, _ = json.Marshal(tt.reqBody)
			}

			req := httptest.NewRequest("POST", "/api/download/async", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			handler.DownloadAsync(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.checkResponse != nil {
				var response map[string]interface{}
				json.Unmarshal(w.Body.Bytes(), &response)
				tt.checkResponse(t, response)
			}
		})
	}
}

func TestDownloadHandler_GetDownloadStatus_AllCases(t *testing.T) {
	tests := []struct {
		name           string
		url            string
		expectedStatus int
		checkResponse  func(t *testing.T, resp handlers.JobStatus)
	}{
		{
			name:           "with query parameter",
			url:            "/api/download/status?job_id=completed-job",
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, resp handlers.JobStatus) {
				if resp.JobID != "completed-job" {
					t.Errorf("Expected job_id 'completed-job', got %s", resp.JobID)
				}
				if resp.Status != "completed" {
					t.Error("Expected status 'completed'")
				}
				if resp.Progress != 100 {
					t.Error("Expected progress 100")
				}
				if resp.CompletedAt == nil {
					t.Error("Expected CompletedAt to be set")
				}
			},
		},
		{
			name:           "with path parameter",
			url:            "/api/download/status/failed-job",
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, resp handlers.JobStatus) {
				if resp.JobID != "failed-job" {
					t.Errorf("Expected job_id 'failed-job', got %s", resp.JobID)
				}
				if resp.Status != "failed" {
					t.Error("Expected status 'failed'")
				}
				if resp.ErrorMessage == nil || *resp.ErrorMessage != "Download failed" {
					t.Error("Expected error message")
				}
			},
		},
		{
			name:           "non-existent job",
			url:            "/api/download/status?job_id=non-existent",
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "missing job_id",
			url:            "/api/download/status",
			expectedStatus: http.StatusBadRequest, // Returns BadRequest for "status"
		},
		{
			name:           "empty job_id",
			url:            "/api/download/status?job_id=",
			expectedStatus: http.StatusBadRequest, // Returns BadRequest for empty
		},
		{
			name:           "empty path with slash",
			url:            "/api/download/status/",
			expectedStatus: http.StatusBadRequest, // Returns BadRequest for empty
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := NewCompleteMockDownloadService()
			handler := handlers.NewDownloadHandler(mockService)

			req := httptest.NewRequest("GET", tt.url, nil)
			w := httptest.NewRecorder()

			handler.GetDownloadStatus(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.checkResponse != nil && w.Code == http.StatusOK {
				var response handlers.JobStatus
				json.Unmarshal(w.Body.Bytes(), &response)
				tt.checkResponse(t, response)
			}
		})
	}
}

func TestDownloadHandler_HelperMethods(t *testing.T) {
	mockService := NewCompleteMockDownloadService()
	handler := handlers.NewDownloadHandler(mockService)

	t.Run("respondJSON", func(t *testing.T) {
		w := httptest.NewRecorder()

		// Use DownloadSync to test respondJSON indirectly
		req := httptest.NewRequest("POST", "/api/download/sync", strings.NewReader("invalid"))
		req.Header.Set("Content-Type", "application/json")

		handler.DownloadSync(w, req)

		// Should return JSON error response
		if w.Header().Get("Content-Type") != "application/json" {
			t.Error("Expected Content-Type to be application/json")
		}

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		if err != nil {
			t.Errorf("Response should be valid JSON: %v", err)
		}
	})

	t.Run("respondError", func(t *testing.T) {
		w := httptest.NewRecorder()

		// Test error response through missing accounts
		reqBody, _ := json.Marshal(handlers.DownloadRequest{
			Accounts: []string{},
		})
		req := httptest.NewRequest("POST", "/api/download/sync", bytes.NewReader(reqBody))
		req.Header.Set("Content-Type", "application/json")

		handler.DownloadSync(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
		}

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)

		if response["error"] == nil {
			t.Error("Expected error field in response")
		}
	})
}

func TestDownloadRequest_Structure(t *testing.T) {
	// Test that DownloadRequest properly marshals/unmarshals
	req := handlers.DownloadRequest{
		Accounts: []string{"acc1", "acc2"},
		FromDate: "2024-01-01",
		ToDate:   "2024-01-31",
		Mode:     "fast",
	}

	// Marshal
	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	// Unmarshal
	var decoded handlers.DownloadRequest
	err = json.Unmarshal(data, &decoded)
	if err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	// Verify
	if len(decoded.Accounts) != 2 {
		t.Error("Accounts not properly decoded")
	}
	if decoded.FromDate != "2024-01-01" {
		t.Error("FromDate not properly decoded")
	}
	if decoded.ToDate != "2024-01-31" {
		t.Error("ToDate not properly decoded")
	}
	if decoded.Mode != "fast" {
		t.Error("Mode not properly decoded")
	}
}

func TestJobStatus_Structure(t *testing.T) {
	now := time.Now()
	errorMsg := "test error"

	status := handlers.JobStatus{
		JobID:        "test-job",
		Status:       "completed",
		Progress:     100,
		TotalRecords: 50,
		ErrorMessage: &errorMsg,
		CompletedAt:  &now,
	}

	// Marshal
	data, err := json.Marshal(status)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	// Unmarshal
	var decoded handlers.JobStatus
	err = json.Unmarshal(data, &decoded)
	if err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	// Verify all fields
	if decoded.JobID != "test-job" {
		t.Error("JobID not properly decoded")
	}
	if decoded.Status != "completed" {
		t.Error("Status not properly decoded")
	}
	if decoded.Progress != 100 {
		t.Error("Progress not properly decoded")
	}
	if decoded.TotalRecords != 50 {
		t.Error("TotalRecords not properly decoded")
	}
	if decoded.ErrorMessage == nil || *decoded.ErrorMessage != "test error" {
		t.Error("ErrorMessage not properly decoded")
	}
	if decoded.CompletedAt == nil {
		t.Error("CompletedAt not properly decoded")
	}
}
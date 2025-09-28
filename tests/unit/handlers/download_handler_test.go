package handlers_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/yhonda-ohishi/etc_meisai/src/handlers"
	"github.com/yhonda-ohishi/etc_meisai/src/services"
)

// MockDownloadService implements DownloadServiceInterface for testing
type MockDownloadService struct {
	accountIDs []string
}

func (m *MockDownloadService) GetAllAccountIDs() []string {
	return m.accountIDs
}

func (m *MockDownloadService) ProcessAsync(jobID string, accounts []string, fromDate, toDate string) {
	// Mock implementation
}

func (m *MockDownloadService) GetJobStatus(jobID string) (*services.DownloadJob, bool) {
	if jobID == "test-job-123" {
		return &services.DownloadJob{
			ID:       jobID,
			Status:   "processing",
			Progress: 50,
		}, true
	}
	return nil, false
}

func TestDownloadHandler_DownloadSync(t *testing.T) {
	// Setup
	mockService := &MockDownloadService{
		accountIDs: []string{"test1", "test2"},
	}
	handler := handlers.NewDownloadHandler(mockService)

	// Create request
	reqBody := handlers.DownloadRequest{
		Accounts: []string{"test1"},
		FromDate: "2024-01-01",
		ToDate:   "2024-01-31",
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/download/sync", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	// Create response recorder
	w := httptest.NewRecorder()

	// Execute
	handler.DownloadSync(w, req)

	// Assert
	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	if response["success"] != true {
		t.Error("Expected success to be true")
	}
}

func TestDownloadHandler_DownloadAsync(t *testing.T) {
	// Setup
	mockService := &MockDownloadService{
		accountIDs: []string{"test1", "test2"},
	}
	handler := handlers.NewDownloadHandler(mockService)

	// Create request with empty accounts (should use all accounts)
	reqBody := handlers.DownloadRequest{
		Accounts: []string{},
		FromDate: "2024-01-01",
		ToDate:   "2024-01-31",
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/download/async", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	// Create response recorder
	w := httptest.NewRecorder()

	// Execute
	handler.DownloadAsync(w, req)

	// Assert
	if w.Code != http.StatusAccepted {
		t.Errorf("Expected status %d, got %d", http.StatusAccepted, w.Code)
	}

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	if response["job_id"] == "" {
		t.Error("Expected job_id in response")
	}

	if response["status"] != "pending" {
		t.Errorf("Expected status 'pending', got %s", response["status"])
	}
}

func TestDownloadHandler_GetDownloadStatus(t *testing.T) {
	// Setup
	mockService := &MockDownloadService{}
	handler := handlers.NewDownloadHandler(mockService)

	// Create request
	req := httptest.NewRequest("GET", "/api/download/status?job_id=test-job-123", nil)

	// Create response recorder
	w := httptest.NewRecorder()

	// Execute
	handler.GetDownloadStatus(w, req)

	// Assert
	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response handlers.JobStatus
	json.Unmarshal(w.Body.Bytes(), &response)

	if response.JobID != "test-job-123" {
		t.Errorf("Expected job_id 'test-job-123', got %s", response.JobID)
	}

	if response.Progress != 50 {
		t.Errorf("Expected progress 50, got %d", response.Progress)
	}
}
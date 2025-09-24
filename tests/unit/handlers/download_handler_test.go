package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/yhonda-ohishi/etc_meisai/src/handlers"
	"github.com/yhonda-ohishi/etc_meisai/src/services"
)

// Mock implementations for download handler tests
type MockDownloadService struct {
	mock.Mock
}

func (m *MockDownloadService) GetAllAccountIDs() []string {
	args := m.Called()
	return args.Get(0).([]string)
}

func (m *MockDownloadService) ProcessAsync(jobID string, accounts []string, fromDate, toDate string) {
	m.Called(jobID, accounts, fromDate, toDate)
}

func (m *MockDownloadService) GetJobStatus(jobID string) (*services.DownloadJob, bool) {
	args := m.Called(jobID)
	if args.Get(0) == nil {
		return nil, args.Bool(1)
	}
	return args.Get(0).(*services.DownloadJob), args.Bool(1)
}

// Test helper functions
func createTestDownloadHandler() (*handlers.DownloadHandler, *MockDownloadService) {
	mockRegistry := createMockServiceRegistry()
	logger := createTestLogger()
	baseHandler := *handlers.NewBaseHandler(mockRegistry, logger)

	mockDownloadService := &MockDownloadService{}

	downloadHandler := handlers.NewDownloadHandler(baseHandler, mockDownloadService)
	return downloadHandler, mockDownloadService
}

// TestNewDownloadHandler tests download handler creation
func TestNewDownloadHandler(t *testing.T) {
	t.Run("successful creation", func(t *testing.T) {
		handler, mockService := createTestDownloadHandler()

		assert.NotNil(t, handler)
		assert.NotNil(t, handler.BaseHandler)
		assert.Equal(t, mockService, handler.DownloadService)
	})
}

// TestDownloadSync tests synchronous download functionality
func TestDownloadSync(t *testing.T) {
	t.Run("successful sync download with all parameters", func(t *testing.T) {
		handler, _ := createTestDownloadHandler()

		requestBody := handlers.DownloadRequest{
			Accounts: []string{"account1", "account2"},
			FromDate: "2023-01-01",
			ToDate:   "2023-12-31",
			Mode:     "sync",
		}

		bodyBytes, _ := json.Marshal(requestBody)
		req := httptest.NewRequest("POST", "/api/download/sync", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.DownloadSync(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response handlers.SuccessResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.True(t, response.Success)
		assert.Contains(t, response.Message, "Download completed successfully")

		data := response.Data.(map[string]interface{})
		assert.Contains(t, data, "success")
		assert.Contains(t, data, "record_count")
		assert.Contains(t, data, "csv_path")
		assert.Contains(t, data, "records")
	})

	t.Run("sync download with default date parameters", func(t *testing.T) {
		handler, _ := createTestDownloadHandler()

		requestBody := handlers.DownloadRequest{
			Accounts: []string{"account1"},
		}

		bodyBytes, _ := json.Marshal(requestBody)
		req := httptest.NewRequest("POST", "/api/download/sync", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.DownloadSync(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response handlers.SuccessResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.True(t, response.Success)
	})

	t.Run("sync download with missing accounts", func(t *testing.T) {
		handler, _ := createTestDownloadHandler()

		requestBody := handlers.DownloadRequest{
			FromDate: "2023-01-01",
			ToDate:   "2023-12-31",
		}

		bodyBytes, _ := json.Marshal(requestBody)
		req := httptest.NewRequest("POST", "/api/download/sync", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.DownloadSync(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var errorResponse handlers.ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &errorResponse)
		assert.NoError(t, err)
		assert.Equal(t, "MISSING_ACCOUNTS", errorResponse.Error.Code)
		assert.Contains(t, errorResponse.Error.Message, "At least one account is required")
	})

	t.Run("sync download with invalid JSON", func(t *testing.T) {
		handler, _ := createTestDownloadHandler()

		invalidJSON := `{"accounts": ["account1", "from_date": "invalid}`
		req := httptest.NewRequest("POST", "/api/download/sync", strings.NewReader(invalidJSON))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.DownloadSync(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var errorResponse handlers.ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &errorResponse)
		assert.NoError(t, err)
		assert.Equal(t, "INVALID_REQUEST", errorResponse.Error.Code)
	})

	t.Run("sync download with partial date parameters", func(t *testing.T) {
		handler, _ := createTestDownloadHandler()

		requestBody := handlers.DownloadRequest{
			Accounts: []string{"account1"},
			FromDate: "2023-01-01",
			// ToDate missing - should be set to today
		}

		bodyBytes, _ := json.Marshal(requestBody)
		req := httptest.NewRequest("POST", "/api/download/sync", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.DownloadSync(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("sync download date parameter defaults", func(t *testing.T) {
		handler, _ := createTestDownloadHandler()

		requestBody := handlers.DownloadRequest{
			Accounts: []string{"account1"},
			// Both dates missing - should be set to defaults
		}

		bodyBytes, _ := json.Marshal(requestBody)
		req := httptest.NewRequest("POST", "/api/download/sync", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.DownloadSync(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}

// TestDownloadAsync tests asynchronous download functionality
func TestDownloadAsync(t *testing.T) {
	t.Run("successful async download", func(t *testing.T) {
		handler, mockService := createTestDownloadHandler()

		mockService.On("ProcessAsync", mock.AnythingOfType("string"), []string{"account1", "account2"}, "2023-01-01", "2023-12-31").Return()

		requestBody := handlers.DownloadRequest{
			Accounts: []string{"account1", "account2"},
			FromDate: "2023-01-01",
			ToDate:   "2023-12-31",
			Mode:     "async",
		}

		bodyBytes, _ := json.Marshal(requestBody)
		req := httptest.NewRequest("POST", "/api/download/async", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.DownloadAsync(w, req)

		assert.Equal(t, http.StatusAccepted, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Contains(t, response, "job_id")
		assert.Equal(t, "pending", response["status"])
		assert.Contains(t, response["message"], "Download job started")

		// Verify job_id is a valid UUID format (basic check)
		jobID := response["job_id"].(string)
		assert.Len(t, jobID, 36) // UUID length with hyphens
		assert.Contains(t, jobID, "-")

		mockService.AssertExpectations(t)
	})

	t.Run("async download with no accounts uses defaults", func(t *testing.T) {
		handler, mockService := createTestDownloadHandler()

		defaultAccounts := []string{"default1", "default2"}
		mockService.On("GetAllAccountIDs").Return(defaultAccounts)
		mockService.On("ProcessAsync", mock.AnythingOfType("string"), defaultAccounts, mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return()

		requestBody := handlers.DownloadRequest{
			FromDate: "2023-01-01",
			ToDate:   "2023-12-31",
		}

		bodyBytes, _ := json.Marshal(requestBody)
		req := httptest.NewRequest("POST", "/api/download/async", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.DownloadAsync(w, req)

		assert.Equal(t, http.StatusAccepted, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Contains(t, response, "job_id")

		mockService.AssertExpectations(t)
	})

	t.Run("async download with no default accounts", func(t *testing.T) {
		handler, mockService := createTestDownloadHandler()

		mockService.On("GetAllAccountIDs").Return([]string{}) // No accounts

		requestBody := handlers.DownloadRequest{
			FromDate: "2023-01-01",
			ToDate:   "2023-12-31",
		}

		bodyBytes, _ := json.Marshal(requestBody)
		req := httptest.NewRequest("POST", "/api/download/async", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.DownloadAsync(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var errorResponse handlers.ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &errorResponse)
		assert.NoError(t, err)
		assert.Equal(t, "NO_ACCOUNTS", errorResponse.Error.Code)
		assert.Contains(t, errorResponse.Error.Message, "No accounts configured")

		mockService.AssertExpectations(t)
	})

	t.Run("async download with invalid JSON", func(t *testing.T) {
		handler, _ := createTestDownloadHandler()

		invalidJSON := `{"accounts": ["account1", "from_date": "invalid}`
		req := httptest.NewRequest("POST", "/api/download/async", strings.NewReader(invalidJSON))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.DownloadAsync(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var errorResponse handlers.ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &errorResponse)
		assert.NoError(t, err)
		assert.Equal(t, "INVALID_REQUEST", errorResponse.Error.Code)
	})

	t.Run("async download with default dates", func(t *testing.T) {
		handler, mockService := createTestDownloadHandler()

		now := time.Now()
		expectedToDate := now.Format("2006-01-02")
		expectedFromDate := now.AddDate(0, -1, 0).Format("2006-01-02")

		mockService.On("ProcessAsync", mock.AnythingOfType("string"), []string{"account1"}, expectedFromDate, expectedToDate).Return()

		requestBody := handlers.DownloadRequest{
			Accounts: []string{"account1"},
			// No dates provided
		}

		bodyBytes, _ := json.Marshal(requestBody)
		req := httptest.NewRequest("POST", "/api/download/async", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.DownloadAsync(w, req)

		assert.Equal(t, http.StatusAccepted, w.Code)
		mockService.AssertExpectations(t)
	})
}

// TestGetDownloadStatus tests download status retrieval
func TestGetDownloadStatus(t *testing.T) {
	t.Run("get status with query parameter", func(t *testing.T) {
		handler, _ := createTestDownloadHandler()

		req := httptest.NewRequest("GET", "/api/download/status?job_id=test-job-123", nil)
		w := httptest.NewRecorder()

		handler.GetDownloadStatus(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response handlers.SuccessResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.True(t, response.Success)
		assert.Contains(t, response.Message, "Status for job test-job-123")

		data := response.Data.(map[string]interface{})
		assert.Equal(t, "test-job-123", data["job_id"])
		assert.Contains(t, data, "status")
		assert.Contains(t, data, "progress")
		assert.Contains(t, data, "total_records")
	})

	t.Run("get status with URL path parameter", func(t *testing.T) {
		handler, _ := createTestDownloadHandler()

		req := httptest.NewRequest("GET", "/api/download/status/path-job-456", nil)
		w := httptest.NewRecorder()

		handler.GetDownloadStatus(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response handlers.SuccessResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		data := response.Data.(map[string]interface{})
		assert.Equal(t, "path-job-456", data["job_id"])
	})

	t.Run("get status with missing job ID", func(t *testing.T) {
		handler, _ := createTestDownloadHandler()

		req := httptest.NewRequest("GET", "/api/download/status", nil)
		w := httptest.NewRecorder()

		handler.GetDownloadStatus(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var errorResponse handlers.ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &errorResponse)
		assert.NoError(t, err)
		assert.Equal(t, "MISSING_JOB_ID", errorResponse.Error.Code)
		assert.Contains(t, errorResponse.Error.Message, "Job ID is required")
	})

	t.Run("get status from complex URL path", func(t *testing.T) {
		handler, _ := createTestDownloadHandler()

		req := httptest.NewRequest("GET", "/api/v1/download/status/complex-job-789", nil)
		w := httptest.NewRecorder()

		handler.GetDownloadStatus(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response handlers.SuccessResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		data := response.Data.(map[string]interface{})
		assert.Equal(t, "complex-job-789", data["job_id"])
	})

	t.Run("get status with empty URL segments", func(t *testing.T) {
		handler, _ := createTestDownloadHandler()

		req := httptest.NewRequest("GET", "/api/download/status/", nil)
		w := httptest.NewRecorder()

		handler.GetDownloadStatus(w, req)

		// Should extract empty string from path
		assert.Equal(t, http.StatusBadRequest, w.Code)

		var errorResponse handlers.ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &errorResponse)
		assert.NoError(t, err)
		assert.Equal(t, "MISSING_JOB_ID", errorResponse.Error.Code)
	})
}

// TestDownloadRequestStructure tests the download request structure
func TestDownloadRequestStructure(t *testing.T) {
	t.Run("complete download request", func(t *testing.T) {
		downloadReq := handlers.DownloadRequest{
			Accounts: []string{"corp1", "corp2", "personal1"},
			FromDate: "2023-01-01",
			ToDate:   "2023-12-31",
			Mode:     "async",
		}

		jsonData, err := json.Marshal(downloadReq)
		assert.NoError(t, err)

		var unmarshaledReq handlers.DownloadRequest
		err = json.Unmarshal(jsonData, &unmarshaledReq)
		assert.NoError(t, err)

		assert.Equal(t, []string{"corp1", "corp2", "personal1"}, unmarshaledReq.Accounts)
		assert.Equal(t, "2023-01-01", unmarshaledReq.FromDate)
		assert.Equal(t, "2023-12-31", unmarshaledReq.ToDate)
		assert.Equal(t, "async", unmarshaledReq.Mode)
	})

	t.Run("minimal download request", func(t *testing.T) {
		downloadReq := handlers.DownloadRequest{
			Accounts: []string{"account1"},
		}

		jsonData, err := json.Marshal(downloadReq)
		assert.NoError(t, err)

		var unmarshaledReq handlers.DownloadRequest
		err = json.Unmarshal(jsonData, &unmarshaledReq)
		assert.NoError(t, err)

		assert.Equal(t, []string{"account1"}, unmarshaledReq.Accounts)
		assert.Empty(t, unmarshaledReq.FromDate)
		assert.Empty(t, unmarshaledReq.ToDate)
		assert.Empty(t, unmarshaledReq.Mode)
	})
}

// TestJobStatusStructure tests the job status structure
func TestJobStatusStructure(t *testing.T) {
	t.Run("complete job status", func(t *testing.T) {
		completedTime := time.Now()
		errorMsg := "Processing failed"

		jobStatus := handlers.JobStatus{
			JobID:        "job-123",
			Status:       "failed",
			Progress:     75,
			TotalRecords: 1000,
			ErrorMessage: &errorMsg,
			CompletedAt:  &completedTime,
		}

		jsonData, err := json.Marshal(jobStatus)
		assert.NoError(t, err)

		var unmarshaledStatus handlers.JobStatus
		err = json.Unmarshal(jsonData, &unmarshaledStatus)
		assert.NoError(t, err)

		assert.Equal(t, "job-123", unmarshaledStatus.JobID)
		assert.Equal(t, "failed", unmarshaledStatus.Status)
		assert.Equal(t, 75, unmarshaledStatus.Progress)
		assert.Equal(t, 1000, unmarshaledStatus.TotalRecords)
		assert.NotNil(t, unmarshaledStatus.ErrorMessage)
		assert.Equal(t, "Processing failed", *unmarshaledStatus.ErrorMessage)
		assert.NotNil(t, unmarshaledStatus.CompletedAt)
	})

	t.Run("minimal job status", func(t *testing.T) {
		jobStatus := handlers.JobStatus{
			JobID:  "job-456",
			Status: "pending",
		}

		jsonData, err := json.Marshal(jobStatus)
		assert.NoError(t, err)

		var unmarshaledStatus handlers.JobStatus
		err = json.Unmarshal(jsonData, &unmarshaledStatus)
		assert.NoError(t, err)

		assert.Equal(t, "job-456", unmarshaledStatus.JobID)
		assert.Equal(t, "pending", unmarshaledStatus.Status)
		assert.Equal(t, 0, unmarshaledStatus.Progress)
		assert.Equal(t, 0, unmarshaledStatus.TotalRecords)
		assert.Nil(t, unmarshaledStatus.ErrorMessage)
		assert.Nil(t, unmarshaledStatus.CompletedAt)
	})
}

// TestConcurrentDownloads tests concurrent download requests
func TestConcurrentDownloads(t *testing.T) {
	t.Run("concurrent async downloads", func(t *testing.T) {
		handler, mockService := createTestDownloadHandler()

		// Set up mock expectations for multiple calls
		mockService.On("ProcessAsync", mock.AnythingOfType("string"), mock.AnythingOfType("[]string"), mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return().Times(5)

		const numRequests = 5
		done := make(chan bool, numRequests)

		// Launch concurrent requests
		for i := 0; i < numRequests; i++ {
			go func(id int) {
				requestBody := handlers.DownloadRequest{
					Accounts: []string{strings.ReplaceAll("account-{id}", "{id}", string(rune(id+'0')))},
					FromDate: "2023-01-01",
					ToDate:   "2023-12-31",
				}

				bodyBytes, _ := json.Marshal(requestBody)
				req := httptest.NewRequest("POST", "/api/download/async", bytes.NewReader(bodyBytes))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()

				handler.DownloadAsync(w, req)

				assert.Equal(t, http.StatusAccepted, w.Code)
				done <- true
			}(i)
		}

		// Wait for all requests to complete
		for i := 0; i < numRequests; i++ {
			<-done
		}

		mockService.AssertExpectations(t)
	})

	t.Run("concurrent status checks", func(t *testing.T) {
		handler, _ := createTestDownloadHandler()

		const numRequests = 10
		done := make(chan bool, numRequests)

		// Launch concurrent status requests
		for i := 0; i < numRequests; i++ {
			go func(id int) {
				jobID := strings.ReplaceAll("job-{id}", "{id}", string(rune(id+'0')))
				req := httptest.NewRequest("GET", "/api/download/status?job_id="+jobID, nil)
				w := httptest.NewRecorder()

				handler.GetDownloadStatus(w, req)

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

// TestDownloadHandlerEdgeCases tests edge cases
func TestDownloadHandlerEdgeCases(t *testing.T) {
	t.Run("very large account list", func(t *testing.T) {
		handler, mockService := createTestDownloadHandler()

		// Create a large list of accounts
		accounts := make([]string, 1000)
		for i := 0; i < 1000; i++ {
			accounts[i] = strings.ReplaceAll("account-{id}", "{id}", string(rune(i)))
		}

		mockService.On("ProcessAsync", mock.AnythingOfType("string"), accounts, mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return()

		requestBody := handlers.DownloadRequest{
			Accounts: accounts,
			FromDate: "2023-01-01",
			ToDate:   "2023-12-31",
		}

		bodyBytes, _ := json.Marshal(requestBody)
		req := httptest.NewRequest("POST", "/api/download/async", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		start := time.Now()
		handler.DownloadAsync(w, req)
		duration := time.Since(start)

		assert.Equal(t, http.StatusAccepted, w.Code)
		assert.Less(t, duration, 100*time.Millisecond, "Large account list should be processed quickly")

		mockService.AssertExpectations(t)
	})

	t.Run("special characters in dates", func(t *testing.T) {
		handler, _ := createTestDownloadHandler()

		requestBody := handlers.DownloadRequest{
			Accounts: []string{"account1"},
			FromDate: "2023/01/01", // Different date format
			ToDate:   "2023.12.31", // Different date format
		}

		bodyBytes, _ := json.Marshal(requestBody)
		req := httptest.NewRequest("POST", "/api/download/sync", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.DownloadSync(w, req)

		// Should still work since we don't validate date format in handler
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("unicode in account names", func(t *testing.T) {
		handler, mockService := createTestDownloadHandler()

		unicodeAccounts := []string{"アカウント1", "账户2", "счет3"}
		mockService.On("ProcessAsync", mock.AnythingOfType("string"), unicodeAccounts, mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return()

		requestBody := handlers.DownloadRequest{
			Accounts: unicodeAccounts,
			FromDate: "2023-01-01",
			ToDate:   "2023-12-31",
		}

		bodyBytes, _ := json.Marshal(requestBody)
		req := httptest.NewRequest("POST", "/api/download/async", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.DownloadAsync(w, req)

		assert.Equal(t, http.StatusAccepted, w.Code)
		mockService.AssertExpectations(t)
	})

	t.Run("empty request body", func(t *testing.T) {
		handler, _ := createTestDownloadHandler()

		req := httptest.NewRequest("POST", "/api/download/sync", strings.NewReader(""))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.DownloadSync(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var errorResponse handlers.ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &errorResponse)
		assert.NoError(t, err)
		assert.Equal(t, "INVALID_REQUEST", errorResponse.Error.Code)
	})
}
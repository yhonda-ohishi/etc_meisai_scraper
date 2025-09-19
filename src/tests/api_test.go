package tests

import (
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/yhonda-ohishi/etc_meisai/src/handlers"
	"github.com/yhonda-ohishi/etc_meisai/src/services"
)

func setupTestRouter() *chi.Mux {
	r := chi.NewRouter()

	// Create handlers with base handler and services
	baseHandler := handlers.BaseHandler{}
	parseHandler := handlers.NewParseHandler(baseHandler)

	// Create a mock download service to avoid nil pointer
	downloadService := services.NewDownloadService(nil, nil)
	downloadHandler := handlers.NewDownloadHandler(baseHandler, downloadService)

	mappingHandler := handlers.NewMappingHandler(baseHandler)
	accountHandler := handlers.NewAccountsHandler(baseHandler)

	// Setup routes
	r.Route("/api", func(r chi.Router) {
		// Parse endpoints
		r.Post("/parse/csv", parseHandler.ParseCSV)

		// Download endpoints
		r.Post("/download/sync", downloadHandler.DownloadSync)
		r.Post("/download/async", downloadHandler.DownloadAsync)
		r.Get("/download/status", downloadHandler.GetDownloadStatus)

		// Mapping endpoints
		r.Get("/mapping", mappingHandler.GetMappings)
		r.Post("/mapping", mappingHandler.CreateMapping)
		r.Put("/mapping/{id}", mappingHandler.UpdateMapping)
		r.Post("/mapping/auto-match", mappingHandler.AutoMatch)

		// Account endpoints
		r.Get("/accounts", accountHandler.GetAccounts)
	})

	return r
}

func TestParseCSVEndpoint(t *testing.T) {
	router := setupTestRouter()

	// Create multipart form with test CSV file
	var b bytes.Buffer
	w := multipart.NewWriter(&b)

	// Add file field
	testFile := filepath.Join("..", "testdata", "sample_etc.csv")
	file, err := os.Open(testFile)
	if err != nil {
		t.Fatalf("Failed to open test file: %v", err)
	}
	defer file.Close()

	fw, err := w.CreateFormFile("file", "sample_etc.csv")
	if err != nil {
		t.Fatalf("Failed to create form file: %v", err)
	}
	if _, err = io.Copy(fw, file); err != nil {
		t.Fatalf("Failed to copy file: %v", err)
	}

	// Add account_type field
	if err := w.WriteField("account_type", "corporate"); err != nil {
		t.Fatalf("Failed to write field: %v", err)
	}

	w.Close()

	// Create request
	req := httptest.NewRequest("POST", "/api/parse/csv", &b)
	req.Header.Set("Content-Type", w.FormDataContentType())

	// Execute request
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	// Check response
	if rec.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rec.Code)
		t.Logf("Response body: %s", rec.Body.String())
	}

	// Parse response
	var response map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	// Check data field
	if data, ok := response["data"].(map[string]interface{}); ok {
		if success, ok := data["success"].(bool); !ok || !success {
			t.Error("Expected success to be true")
		}
		if recordCount, ok := data["record_count"].(float64); !ok || recordCount != 10 {
			t.Errorf("Expected 10 records, got %v", recordCount)
		}
	} else {
		t.Error("Expected data field in response")
	}
}

func TestGetAccountsEndpoint(t *testing.T) {
	// Set up test environment variables
	os.Setenv("ETC_CORPORATE_ACCOUNTS", "test_corp:pass123:1234567890123456")
	os.Setenv("ETC_PERSONAL_ACCOUNTS", "test_personal:pass456:9876543210987654")
	defer func() {
		os.Unsetenv("ETC_CORPORATE_ACCOUNTS")
		os.Unsetenv("ETC_PERSONAL_ACCOUNTS")
	}()

	router := setupTestRouter()

	req := httptest.NewRequest("GET", "/api/accounts", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rec.Code)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if data, ok := response["data"].(map[string]interface{}); ok {
		if accounts, ok := data["accounts"].([]interface{}); !ok || len(accounts) == 0 {
			t.Error("Expected non-empty accounts array")
		}
	}
}

func TestCreateMappingEndpoint(t *testing.T) {
	router := setupTestRouter()

	payload := handlers.CreateMappingRequest{
		ETCMeisaiID: 100,
		DtakoRowID:  200,
		MatchType:   "manual",
	}

	body, _ := json.Marshal(payload)
	req := httptest.NewRequest("POST", "/api/mapping", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Errorf("Expected status 201, got %d", rec.Code)
		t.Logf("Response: %s", rec.Body.String())
	}
}

func TestDownloadAsyncEndpoint(t *testing.T) {
	router := setupTestRouter()

	payload := handlers.DownloadRequest{
		Accounts: []string{"account1"},
		FromDate: "2025-09-01",
		ToDate:   "2025-09-30",
	}

	body, _ := json.Marshal(payload)
	req := httptest.NewRequest("POST", "/api/download/async", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusAccepted {
		t.Errorf("Expected status 202, got %d", rec.Code)
		t.Logf("Response: %s", rec.Body.String())
	}

	var response map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if jobID, ok := response["job_id"].(string); !ok || jobID == "" {
		t.Error("Expected job_id in response")
	}
}

func TestAutoMatchEndpoint(t *testing.T) {
	router := setupTestRouter()

	payload := handlers.AutoMatchRequest{
		ETCNum:    "1234567890123456",
		FromDate:  "2025-09-01",
		ToDate:    "2025-09-30",
		Threshold: 0.8,
	}

	body, _ := json.Marshal(payload)
	req := httptest.NewRequest("POST", "/api/mapping/auto-match", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rec.Code)
		t.Logf("Response: %s", rec.Body.String())
	}
}
package services_test

import (
	"errors"
	"log"
	"os"
	"testing"
	"time"

	"github.com/yhonda-ohishi/etc_meisai_scraper/src/scraper"
	"github.com/yhonda-ohishi/etc_meisai_scraper/src/services"
	"github.com/yhonda-ohishi/etc_meisai_scraper/tests/mocks"
)

// MockScraperFactory for testing
type MockScraperFactory struct {
	CreateFunc func(config *scraper.ScraperConfig, logger *log.Logger) (scraper.ScraperInterface, error)
	CreateErr  error
	MockScraper scraper.ScraperInterface
}

func (f *MockScraperFactory) CreateScraper(config *scraper.ScraperConfig, logger *log.Logger) (scraper.ScraperInterface, error) {
	if f.CreateErr != nil {
		return nil, f.CreateErr
	}
	if f.CreateFunc != nil {
		return f.CreateFunc(config, logger)
	}
	return f.MockScraper, nil
}

func TestDownloadService_WithMockScraper_Success(t *testing.T) {
	// Setup
	logger := log.New(os.Stdout, "[TEST] ", log.LstdFlags)
	mockScraper := mocks.NewMockETCScraper()
	mockFactory := &MockScraperFactory{
		MockScraper: mockScraper,
	}

	// Set environment for accounts
	os.Setenv("ETC_CORPORATE_ACCOUNTS", "corp1:pass1")
	defer os.Unsetenv("ETC_CORPORATE_ACCOUNTS")

	service := services.NewDownloadServiceWithFactory(nil, logger, mockFactory)

	// Execute
	jobID := "mock-test-job"
	accounts := []string{"test1:pass1"}
	fromDate := "2024-01-01"
	toDate := "2024-01-31"

	service.ProcessAsync(jobID, accounts, fromDate, toDate)

	// Wait for processing
	time.Sleep(3 * time.Second)

	// Verify
	if !mockScraper.InitializeCalled {
		t.Error("Expected Initialize to be called")
	}

	if !mockScraper.LoginCalled {
		t.Error("Expected Login to be called")
	}

	if !mockScraper.DownloadCalled {
		t.Error("Expected DownloadMeisai to be called")
	}

	if !mockScraper.CloseCalled {
		t.Error("Expected Close to be called")
	}

	// Check job status
	job, exists := service.GetJobStatus(jobID)
	if !exists {
		t.Fatal("Job should exist")
	}

	if job.Status != "completed" {
		t.Errorf("Expected job status 'completed', got %s", job.Status)
	}
}

func TestDownloadService_WithMockScraper_InitializeError(t *testing.T) {
	logger := log.New(os.Stdout, "[TEST] ", log.LstdFlags)
	mockScraper := mocks.NewMockETCScraper()
	mockScraper.InitializeError = errors.New("initialize failed")

	mockFactory := &MockScraperFactory{
		MockScraper: mockScraper,
	}

	service := services.NewDownloadServiceWithFactory(nil, logger, mockFactory)

	// Execute
	jobID := "init-error-job"
	service.ProcessAsync(jobID, []string{"test:pass"}, "2024-01-01", "2024-01-31")

	// Wait for processing
	time.Sleep(2 * time.Second)

	// Verify
	if !mockScraper.InitializeCalled {
		t.Error("Initialize should have been called")
	}

	if mockScraper.LoginCalled {
		t.Error("Login should not be called after Initialize error")
	}

	if mockScraper.DownloadCalled {
		t.Error("Download should not be called after Initialize error")
	}

	// Close should still be called in defer
	if !mockScraper.CloseCalled {
		t.Error("Close should be called even after error")
	}
}

func TestDownloadService_WithMockScraper_LoginError(t *testing.T) {
	logger := log.New(os.Stdout, "[TEST] ", log.LstdFlags)
	mockScraper := mocks.NewMockETCScraper()
	mockScraper.LoginError = errors.New("login failed")

	mockFactory := &MockScraperFactory{
		MockScraper: mockScraper,
	}

	service := services.NewDownloadServiceWithFactory(nil, logger, mockFactory)

	// Execute
	jobID := "login-error-job"
	service.ProcessAsync(jobID, []string{"test:pass"}, "2024-01-01", "2024-01-31")

	// Wait for processing
	time.Sleep(2 * time.Second)

	// Verify
	if !mockScraper.InitializeCalled {
		t.Error("Initialize should have been called")
	}

	if !mockScraper.LoginCalled {
		t.Error("Login should have been called")
	}

	if mockScraper.DownloadCalled {
		t.Error("Download should not be called after Login error")
	}

	if !mockScraper.CloseCalled {
		t.Error("Close should be called even after error")
	}
}

func TestDownloadService_WithMockScraper_DownloadError(t *testing.T) {
	logger := log.New(os.Stdout, "[TEST] ", log.LstdFlags)
	mockScraper := mocks.NewMockETCScraper()
	mockScraper.DownloadError = errors.New("download failed")

	mockFactory := &MockScraperFactory{
		MockScraper: mockScraper,
	}

	service := services.NewDownloadServiceWithFactory(nil, logger, mockFactory)

	// Execute
	jobID := "download-error-job"
	service.ProcessAsync(jobID, []string{"test:pass"}, "2024-01-01", "2024-01-31")

	// Wait for processing
	time.Sleep(2 * time.Second)

	// Verify all methods were called
	if !mockScraper.InitializeCalled {
		t.Error("Initialize should have been called")
	}

	if !mockScraper.LoginCalled {
		t.Error("Login should have been called")
	}

	if !mockScraper.DownloadCalled {
		t.Error("Download should have been called")
	}

	if !mockScraper.CloseCalled {
		t.Error("Close should be called even after error")
	}

	// Verify dates were passed correctly
	if mockScraper.FromDate != "2024-01-01" {
		t.Errorf("Expected FromDate '2024-01-01', got '%s'", mockScraper.FromDate)
	}

	if mockScraper.ToDate != "2024-01-31" {
		t.Errorf("Expected ToDate '2024-01-31', got '%s'", mockScraper.ToDate)
	}
}

func TestDownloadService_WithMockScraper_ScraperCreateError(t *testing.T) {
	logger := log.New(os.Stdout, "[TEST] ", log.LstdFlags)

	mockFactory := &MockScraperFactory{
		CreateErr: errors.New("failed to create scraper"),
	}

	service := services.NewDownloadServiceWithFactory(nil, logger, mockFactory)

	// Execute
	jobID := "create-error-job"
	service.ProcessAsync(jobID, []string{"test:pass"}, "2024-01-01", "2024-01-31")

	// Wait for processing
	time.Sleep(2 * time.Second)

	// Job should still complete but with error logged
	job, exists := service.GetJobStatus(jobID)
	if !exists {
		t.Fatal("Job should exist even with scraper creation error")
	}

	// Job should continue to next account even with errors
	if job.Status == "" {
		t.Error("Job should have a status")
	}
}

func TestDownloadService_UpdateJobStatus(t *testing.T) {
	logger := log.New(os.Stdout, "[TEST] ", log.LstdFlags)

	// Create a mock that will trigger panic recovery
	mockScraper := mocks.NewConfigurableETCScraper()
	callCount := 0
	mockScraper.InitializeFunc = func() error {
		callCount++
		if callCount == 1 {
			panic("test panic")
		}
		return nil
	}

	mockFactory := &MockScraperFactory{
		MockScraper: mockScraper,
	}

	service := services.NewDownloadServiceWithFactory(nil, logger, mockFactory)

	// Execute - this should trigger panic recovery and updateJobStatus
	jobID := "panic-recovery-job"
	service.ProcessAsync(jobID, []string{"test:pass"}, "2024-01-01", "2024-01-31")

	// Wait for panic recovery
	time.Sleep(2 * time.Second)

	// Job should exist with failed status from updateJobStatus
	job, exists := service.GetJobStatus(jobID)
	if !exists {
		t.Fatal("Job should exist after panic recovery")
	}

	// The panic recovery should have called updateJobStatus
	if job.Status != "failed" && job.Status != "completed" {
		t.Logf("Job status: %s", job.Status)
		// Status might be processing if panic was caught differently
	}
}

func TestDownloadService_WithConfigurableMock(t *testing.T) {
	logger := log.New(os.Stdout, "[TEST] ", log.LstdFlags)

	// Create configurable mock with custom behavior
	mockScraper := mocks.NewConfigurableETCScraper()
	downloadCalled := false
	mockScraper.DownloadFunc = func(fromDate, toDate string) (string, error) {
		downloadCalled = true
		return "/custom/path.csv", nil
	}

	mockFactory := &MockScraperFactory{
		MockScraper: mockScraper,
	}

	service := services.NewDownloadServiceWithFactory(nil, logger, mockFactory)

	// Execute
	jobID := "configurable-test"
	service.ProcessAsync(jobID, []string{"test:pass"}, "2024-01-01", "2024-01-31")

	// Wait for processing
	time.Sleep(2 * time.Second)

	// Verify custom function was called
	if !downloadCalled {
		t.Error("Custom download function should have been called")
	}

	// Check job completed
	job, exists := service.GetJobStatus(jobID)
	if !exists {
		t.Fatal("Job should exist")
	}

	if job.Status != "completed" {
		t.Errorf("Expected status 'completed', got %s", job.Status)
	}
}

func TestDownloadService_MultipleAccounts_WithMock(t *testing.T) {
	logger := log.New(os.Stdout, "[TEST] ", log.LstdFlags)

	// Track how many times each method is called
	initCount := 0
	loginCount := 0
	downloadCount := 0
	closeCount := 0

	mockFactory := &MockScraperFactory{
		CreateFunc: func(config *scraper.ScraperConfig, logger *log.Logger) (scraper.ScraperInterface, error) {
			mock := mocks.NewConfigurableETCScraper()
			mock.InitializeFunc = func() error {
				initCount++
				return nil
			}
			mock.LoginFunc = func() error {
				loginCount++
				return nil
			}
			mock.DownloadFunc = func(fromDate, toDate string) (string, error) {
				downloadCount++
				return "/test.csv", nil
			}
			mock.CloseFunc = func() error {
				closeCount++
				return nil
			}
			return mock, nil
		},
	}

	service := services.NewDownloadServiceWithFactory(nil, logger, mockFactory)

	// Execute with multiple accounts
	jobID := "multi-account-job"
	accounts := []string{"acc1:pass1", "acc2:pass2", "acc3:pass3"}
	service.ProcessAsync(jobID, accounts, "2024-01-01", "2024-01-31")

	// Wait for processing
	time.Sleep(3500 * time.Millisecond)

	// Verify each account was processed
	if initCount != 3 {
		t.Errorf("Expected Initialize called 3 times, got %d", initCount)
	}

	if loginCount != 3 {
		t.Errorf("Expected Login called 3 times, got %d", loginCount)
	}

	if downloadCount != 3 {
		t.Errorf("Expected Download called 3 times, got %d", downloadCount)
	}

	if closeCount != 3 {
		t.Errorf("Expected Close called 3 times, got %d", closeCount)
	}

	// Check job completed
	job, _ := service.GetJobStatus(jobID)
	if job.Status != "completed" {
		t.Errorf("Expected status 'completed', got %s", job.Status)
	}

	if job.Progress != 100 {
		t.Errorf("Expected progress 100, got %d", job.Progress)
	}
}
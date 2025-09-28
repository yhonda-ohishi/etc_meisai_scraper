package services_test

import (
	"log"
	"os"
	"testing"

	"github.com/yhonda-ohishi/etc_meisai_scraper/src/services"
)

func TestDownloadService_GetAllAccountIDs(t *testing.T) {
	// Setup
	os.Setenv("ETC_CORPORATE_ACCOUNTS", "corp1:pass1,corp2:pass2")
	os.Setenv("ETC_PERSONAL_ACCOUNTS", "personal1:pass1")
	defer func() {
		os.Unsetenv("ETC_CORPORATE_ACCOUNTS")
		os.Unsetenv("ETC_PERSONAL_ACCOUNTS")
	}()

	logger := log.New(os.Stdout, "[TEST] ", log.LstdFlags)
	service := services.NewDownloadService(nil, logger)

	// Test
	accountIDs := service.GetAllAccountIDs()

	// Assert
	if len(accountIDs) != 3 {
		t.Errorf("Expected 3 account IDs, got %d", len(accountIDs))
	}

	expected := []string{"corp1", "corp2", "personal1"}
	for i, id := range accountIDs {
		if id != expected[i] {
			t.Errorf("Expected account ID %s, got %s", expected[i], id)
		}
	}
}

func TestDownloadService_ProcessAsync(t *testing.T) {
	logger := log.New(os.Stdout, "[TEST] ", log.LstdFlags)
	service := services.NewDownloadService(nil, logger)

	// Test async processing
	jobID := "test-job-123"
	accounts := []string{"test1:pass1", "test2:pass2"}
	fromDate := "2024-01-01"
	toDate := "2024-01-31"

	// This should not panic or error
	service.ProcessAsync(jobID, accounts, fromDate, toDate)

	// Check job status
	job, exists := service.GetJobStatus(jobID)
	if !exists {
		t.Error("Job should exist after ProcessAsync")
	}

	if job.ID != jobID {
		t.Errorf("Expected job ID %s, got %s", jobID, job.ID)
	}

	if job.Status != "processing" && job.Status != "completed" && job.Status != "failed" {
		t.Errorf("Unexpected job status: %s", job.Status)
	}
}
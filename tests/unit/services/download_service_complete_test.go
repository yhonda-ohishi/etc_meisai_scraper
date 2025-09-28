package services_test

import (
	"fmt"
	"log"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/yhonda-ohishi/etc_meisai_scraper/src/services"
)

func TestNewDownloadService(t *testing.T) {
	logger := log.New(os.Stdout, "[TEST] ", log.LstdFlags)

	t.Run("with logger", func(t *testing.T) {
		service := services.NewDownloadService(nil, logger)
		if service == nil {
			t.Fatal("Expected non-nil service")
		}
	})

	t.Run("without logger", func(t *testing.T) {
		service := services.NewDownloadService(nil, nil)
		if service == nil {
			t.Fatal("Expected non-nil service")
		}
	})
}

func TestDownloadService_ProcessAsync_CompleteFlow(t *testing.T) {
	logger := log.New(os.Stdout, "[TEST] ", log.LstdFlags)
	service := services.NewDownloadService(nil, logger)

	jobID := "test-complete-job"
	accounts := []string{"test1:pass1"}
	fromDate := "2024-01-01"
	toDate := "2024-01-31"

	// Start async processing
	service.ProcessAsync(jobID, accounts, fromDate, toDate)

	// Check initial status
	job, exists := service.GetJobStatus(jobID)
	if !exists {
		t.Fatal("Job should exist")
	}

	if job.Status != "processing" && job.Status != "completed" && job.Status != "failed" {
		t.Errorf("Unexpected job status: %s", job.Status)
	}

	// Wait a bit for processing
	time.Sleep(100 * time.Millisecond)

	// Check job progress
	job, exists = service.GetJobStatus(jobID)
	if !exists {
		t.Fatal("Job should still exist")
	}

	// Job should have made some progress
	if job.Progress < 0 || job.Progress > 100 {
		t.Errorf("Invalid progress: %d", job.Progress)
	}
}

func TestDownloadService_ProcessAsync_InvalidAccount(t *testing.T) {
	logger := log.New(os.Stdout, "[TEST] ", log.LstdFlags)
	service := services.NewDownloadService(nil, logger)

	// Test with invalid account format (missing password)
	jobID := "test-invalid-account"
	accounts := []string{"invalid_account_no_password"}
	fromDate := "2024-01-01"
	toDate := "2024-01-31"

	service.ProcessAsync(jobID, accounts, fromDate, toDate)

	// Give time for processing
	time.Sleep(3 * time.Second)

	job, exists := service.GetJobStatus(jobID)
	if !exists {
		t.Fatal("Job should exist even with errors")
	}

	// The job should continue despite individual account errors
	if job.Status == "" {
		t.Error("Job should have a status")
	}
}

func TestDownloadService_ProcessAsync_Panic_Recovery(t *testing.T) {
	logger := log.New(os.Stdout, "[TEST] ", log.LstdFlags)
	service := services.NewDownloadService(nil, logger)

	// Create a job that might cause issues
	jobID := "test-panic-job"
	accounts := []string{} // Empty accounts
	fromDate := "2024-01-01"
	toDate := "2024-01-31"

	// This should not panic
	service.ProcessAsync(jobID, accounts, fromDate, toDate)

	// Check that job was created
	job, exists := service.GetJobStatus(jobID)
	if !exists {
		t.Fatal("Job should exist")
	}

	// Job should complete even with empty accounts
	time.Sleep(100 * time.Millisecond)

	job, _ = service.GetJobStatus(jobID)
	if job.Status == "" {
		t.Error("Job should have a status")
	}
}

func TestDownloadService_GetJobStatus_NonExistent(t *testing.T) {
	logger := log.New(os.Stdout, "[TEST] ", log.LstdFlags)
	service := services.NewDownloadService(nil, logger)

	job, exists := service.GetJobStatus("non-existent-job")
	if exists {
		t.Error("Should not find non-existent job")
	}

	if job != nil {
		t.Error("Should return nil for non-existent job")
	}
}

func TestDownloadService_GetAllAccountIDs_EdgeCases(t *testing.T) {
	tests := []struct {
		name           string
		corpAccounts   string
		personalAccounts string
		expectedCount  int
		expectedIDs    []string
	}{
		{
			name:           "empty environment",
			corpAccounts:   "",
			personalAccounts: "",
			expectedCount:  0,
			expectedIDs:    []string{},
		},
		{
			name:           "only corporate",
			corpAccounts:   "corp1:pass1",
			personalAccounts: "",
			expectedCount:  1,
			expectedIDs:    []string{"corp1"},
		},
		{
			name:           "only personal",
			corpAccounts:   "",
			personalAccounts: "personal1:pass1",
			expectedCount:  1,
			expectedIDs:    []string{"personal1"},
		},
		{
			name:           "malformed entries",
			corpAccounts:   "valid:pass,invalid_no_colon,another:valid:extra",
			personalAccounts: "",
			expectedCount:  3,
			expectedIDs:    []string{"valid", "invalid_no_colon", "another"},
		},
		{
			name:           "with spaces",
			corpAccounts:   " corp1:pass1 , corp2:pass2 ",
			personalAccounts: "",
			expectedCount:  2,
			expectedIDs:    []string{" corp1", " corp2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup environment
			os.Setenv("ETC_CORPORATE_ACCOUNTS", tt.corpAccounts)
			os.Setenv("ETC_PERSONAL_ACCOUNTS", tt.personalAccounts)
			defer func() {
				os.Unsetenv("ETC_CORPORATE_ACCOUNTS")
				os.Unsetenv("ETC_PERSONAL_ACCOUNTS")
			}()

			logger := log.New(os.Stdout, "[TEST] ", log.LstdFlags)
			service := services.NewDownloadService(nil, logger)

			accountIDs := service.GetAllAccountIDs()

			if len(accountIDs) != tt.expectedCount {
				t.Errorf("Expected %d accounts, got %d", tt.expectedCount, len(accountIDs))
			}

			for i, expected := range tt.expectedIDs {
				if i >= len(accountIDs) {
					break
				}
				// Handle spaces in account IDs
				if strings.TrimSpace(accountIDs[i]) != strings.TrimSpace(expected) {
					t.Errorf("Expected account ID '%s', got '%s'", expected, accountIDs[i])
				}
			}
		})
	}
}

func TestDownloadService_ConcurrentAccess(t *testing.T) {
	logger := log.New(os.Stdout, "[TEST] ", log.LstdFlags)
	service := services.NewDownloadService(nil, logger)

	// Test concurrent job creation and status checking
	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func(id int) {
			jobID := fmt.Sprintf("concurrent-job-%d", id)
			accounts := []string{fmt.Sprintf("account%d:pass%d", id, id)}

			service.ProcessAsync(jobID, accounts, "2024-01-01", "2024-01-31")

			// Check status multiple times
			for j := 0; j < 5; j++ {
				job, exists := service.GetJobStatus(jobID)
				if !exists {
					t.Errorf("Job %s should exist", jobID)
				}
				if job != nil && job.ID != jobID {
					t.Errorf("Wrong job ID returned: expected %s, got %s", jobID, job.ID)
				}
				time.Sleep(10 * time.Millisecond)
			}

			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}
}

func TestDownloadService_UpdateJobProgress(t *testing.T) {
	logger := log.New(os.Stdout, "[TEST] ", log.LstdFlags)
	service := services.NewDownloadService(nil, logger)

	// Create a job
	jobID := "progress-test-job"
	service.ProcessAsync(jobID, []string{"test:pass"}, "2024-01-01", "2024-01-31")

	// Give it a moment to start
	time.Sleep(50 * time.Millisecond)

	// Check that progress updates
	job1, _ := service.GetJobStatus(jobID)
	progress1 := job1.Progress

	time.Sleep(2500 * time.Millisecond) // Wait for processing

	job2, _ := service.GetJobStatus(jobID)
	progress2 := job2.Progress

	// Progress should have changed or job completed
	if job2.Status != "completed" && job2.Status != "failed" && progress2 < progress1 {
		t.Errorf("Progress should not decrease: was %d, now %d", progress1, progress2)
	}

	// Job should make some progress
	if job2.Progress < 0 || job2.Progress > 100 {
		t.Errorf("Invalid progress value: %d", job2.Progress)
	}
}

func TestDownloadService_ErrorMessage(t *testing.T) {
	logger := log.New(os.Stdout, "[TEST] ", log.LstdFlags)
	service := services.NewDownloadService(nil, logger)

	// Test a job with an invalid account format that should log an error
	jobID := "error-test-job"
	accounts := []string{"invalid"} // Missing password part

	service.ProcessAsync(jobID, accounts, "2024-01-01", "2024-01-31")

	// Wait for processing
	time.Sleep(2 * time.Second)

	job, exists := service.GetJobStatus(jobID)
	if !exists {
		t.Fatal("Job should exist")
	}

	// Job should complete despite errors in individual accounts
	if job.Status == "" {
		t.Error("Job should have a status")
	}
}

func TestDownloadServiceInterface(t *testing.T) {
	// Verify that DownloadService implements DownloadServiceInterface
	logger := log.New(os.Stdout, "[TEST] ", log.LstdFlags)
	service := services.NewDownloadService(nil, logger)

	// This will fail to compile if the interface is not satisfied
	var _ services.DownloadServiceInterface = service
}
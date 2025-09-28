package services_test

import (
	"context"
	"log"
	"os"
	"testing"
	"time"

	pb "github.com/yhonda-ohishi/etc_meisai/src/pb"
	"github.com/yhonda-ohishi/etc_meisai/src/services"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestDownloadServiceGRPC_DownloadSync(t *testing.T) {
	logger := log.New(os.Stdout, "[TEST] ", log.LstdFlags)
	service := services.NewDownloadServiceGRPC(nil, logger)

	ctx := context.Background()
	req := &pb.DownloadRequest{
		Accounts: []string{"test1:pass1"},
		FromDate: "2024-01-01",
		ToDate:   "2024-01-31",
	}

	resp, err := service.DownloadSync(ctx, req)
	if err != nil {
		t.Fatalf("DownloadSync failed: %v", err)
	}

	if resp == nil {
		t.Fatal("Expected non-nil response")
	}

	if !resp.Success {
		t.Error("Expected success to be true")
	}
}

func TestDownloadServiceGRPC_DownloadAsync(t *testing.T) {
	logger := log.New(os.Stdout, "[TEST] ", log.LstdFlags)
	service := services.NewDownloadServiceGRPC(nil, logger)

	tests := []struct {
		name     string
		req      *pb.DownloadRequest
		wantFail bool
	}{
		{
			name: "with accounts",
			req: &pb.DownloadRequest{
				Accounts: []string{"test1:pass1"},
				FromDate: "2024-01-01",
				ToDate:   "2024-01-31",
			},
			wantFail: false,
		},
		{
			name: "without accounts - use defaults",
			req: &pb.DownloadRequest{
				Accounts: []string{},
				FromDate: "",
				ToDate:   "",
			},
			wantFail: true, // Will fail because no accounts are configured
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			resp, err := service.DownloadAsync(ctx, tt.req)

			if err != nil {
				t.Fatalf("DownloadAsync failed: %v", err)
			}

			if resp == nil {
				t.Fatal("Expected non-nil response")
			}

			if tt.wantFail && resp.Status != "failed" {
				t.Error("Expected status to be 'failed'")
			}

			if !tt.wantFail && resp.JobId == "" {
				t.Error("Expected non-empty job ID")
			}
		})
	}
}

func TestDownloadServiceGRPC_GetJobStatus(t *testing.T) {
	logger := log.New(os.Stdout, "[TEST] ", log.LstdFlags)
	service := services.NewDownloadServiceGRPC(nil, logger)

	// First create a job
	ctx := context.Background()
	createReq := &pb.DownloadRequest{
		Accounts: []string{"test1:pass1"},
		FromDate: "2024-01-01",
		ToDate:   "2024-01-31",
	}

	createResp, err := service.DownloadAsync(ctx, createReq)
	if err != nil {
		t.Fatalf("Failed to create job: %v", err)
	}

	// Test getting status for existing job
	t.Run("existing job without CompletedAt", func(t *testing.T) {
		req := &pb.GetJobStatusRequest{
			JobId: createResp.JobId,
		}

		resp, err := service.GetJobStatus(ctx, req)
		if err != nil {
			t.Fatalf("GetJobStatus failed: %v", err)
		}

		if resp == nil {
			t.Fatal("Expected non-nil response")
		}

		if resp.JobId != createResp.JobId {
			t.Errorf("Expected job ID %s, got %s", createResp.JobId, resp.JobId)
		}

		// Job is still processing, CompletedAt should be nil
		if resp.CompletedAt != nil {
			t.Error("Expected CompletedAt to be nil for processing job")
		}
	})

	// Wait for job to complete
	time.Sleep(3 * time.Second)

	// Test getting status for completed job with CompletedAt
	t.Run("completed job with CompletedAt", func(t *testing.T) {
		req := &pb.GetJobStatusRequest{
			JobId: createResp.JobId,
		}

		resp, err := service.GetJobStatus(ctx, req)
		if err != nil {
			t.Fatalf("GetJobStatus failed: %v", err)
		}

		if resp == nil {
			t.Fatal("Expected non-nil response")
		}

		// After waiting, job should be completed
		if resp.Status == "completed" && resp.CompletedAt == nil {
			t.Error("Expected CompletedAt to be set for completed job")
		}

		// If job is completed, verify CompletedAt is after StartedAt
		if resp.Status == "completed" && resp.CompletedAt != nil && resp.StartedAt != nil {
			if !resp.CompletedAt.AsTime().After(resp.StartedAt.AsTime()) {
				t.Error("CompletedAt should be after StartedAt")
			}
		}
	})

	// Test getting status for non-existing job
	t.Run("non-existing job", func(t *testing.T) {
		req := &pb.GetJobStatusRequest{
			JobId: "non-existing-job",
		}

		resp, err := service.GetJobStatus(ctx, req)
		if err != nil {
			t.Fatalf("GetJobStatus failed: %v", err)
		}

		if resp != nil {
			t.Error("Expected nil response for non-existing job")
		}
	})
}

func TestDownloadServiceGRPC_GetAllAccountIDs(t *testing.T) {
	// Setup environment variables
	os.Setenv("ETC_CORPORATE_ACCOUNTS", "corp1:pass1,corp2:pass2")
	os.Setenv("ETC_PERSONAL_ACCOUNTS", "personal1:pass1")
	defer func() {
		os.Unsetenv("ETC_CORPORATE_ACCOUNTS")
		os.Unsetenv("ETC_PERSONAL_ACCOUNTS")
	}()

	logger := log.New(os.Stdout, "[TEST] ", log.LstdFlags)
	service := services.NewDownloadServiceGRPC(nil, logger)

	ctx := context.Background()
	req := &pb.GetAllAccountIDsRequest{}

	resp, err := service.GetAllAccountIDs(ctx, req)
	if err != nil {
		t.Fatalf("GetAllAccountIDs failed: %v", err)
	}

	if resp == nil {
		t.Fatal("Expected non-nil response")
	}

	if len(resp.AccountIds) != 3 {
		t.Errorf("Expected 3 account IDs, got %d", len(resp.AccountIds))
	}

	expected := []string{"corp1", "corp2", "personal1"}
	for i, id := range resp.AccountIds {
		if id != expected[i] {
			t.Errorf("Expected account ID %s, got %s", expected[i], id)
		}
	}
}

func TestDownloadServiceGRPC_SetDefaultDates(t *testing.T) {
	logger := log.New(os.Stdout, "[TEST] ", log.LstdFlags)
	service := services.NewDownloadServiceGRPC(nil, logger)

	// Test with empty dates (should set defaults)
	ctx := context.Background()
	req := &pb.DownloadRequest{
		Accounts: []string{"test1:pass1"},
		FromDate: "",
		ToDate:   "",
	}

	resp, err := service.DownloadSync(ctx, req)
	if err != nil {
		t.Fatalf("DownloadSync failed: %v", err)
	}

	if resp == nil {
		t.Fatal("Expected non-nil response")
	}

	// Test with provided dates
	req2 := &pb.DownloadRequest{
		Accounts: []string{"test1:pass1"},
		FromDate: "2024-01-01",
		ToDate:   "2024-01-31",
	}

	resp2, err := service.DownloadSync(ctx, req2)
	if err != nil {
		t.Fatalf("DownloadSync failed: %v", err)
	}

	if resp2 == nil {
		t.Fatal("Expected non-nil response")
	}
}

func TestDownloadJob_CompletedAt(t *testing.T) {
	job := &services.DownloadJob{
		ID:          "test-job",
		Status:      "completed",
		Progress:    100,
		StartedAt:   time.Now(),
		CompletedAt: nil,
	}

	// Test with nil CompletedAt
	if job.CompletedAt != nil {
		t.Error("Expected CompletedAt to be nil")
	}

	// Test with non-nil CompletedAt
	now := time.Now()
	job.CompletedAt = &now

	if job.CompletedAt == nil {
		t.Error("Expected CompletedAt to be non-nil")
	}

	if !job.CompletedAt.Equal(now) {
		t.Error("CompletedAt time mismatch")
	}
}

func TestJobStatus_Protobuf(t *testing.T) {
	now := time.Now()
	completedAt := time.Now().Add(1 * time.Hour)

	status := &pb.JobStatus{
		JobId:        "test-job",
		Status:       "completed",
		Progress:     100,
		TotalRecords: 50,
		ErrorMessage: "test error",
		StartedAt:    timestamppb.New(now),
		CompletedAt:  timestamppb.New(completedAt),
	}

	// Verify all fields
	if status.JobId != "test-job" {
		t.Errorf("Expected job ID 'test-job', got %s", status.JobId)
	}

	if status.Status != "completed" {
		t.Errorf("Expected status 'completed', got %s", status.Status)
	}

	if status.Progress != 100 {
		t.Errorf("Expected progress 100, got %d", status.Progress)
	}

	if status.TotalRecords != 50 {
		t.Errorf("Expected total records 50, got %d", status.TotalRecords)
	}

	if status.ErrorMessage != "test error" {
		t.Errorf("Expected error message 'test error', got %s", status.ErrorMessage)
	}

	if !status.StartedAt.AsTime().Equal(now) {
		t.Error("StartedAt time mismatch")
	}

	if !status.CompletedAt.AsTime().Equal(completedAt) {
		t.Error("CompletedAt time mismatch")
	}
}
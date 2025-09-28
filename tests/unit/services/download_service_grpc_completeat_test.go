package services_test

import (
	"context"
	"log"
	"os"
	"testing"
	"time"

	pb "github.com/yhonda-ohishi/etc_meisai_scraper/src/pb"
	"github.com/yhonda-ohishi/etc_meisai_scraper/src/services"
)

func TestDownloadServiceGRPC_GetJobStatus_With_CompletedAt_Coverage(t *testing.T) {
	// Set up test environment
	os.Setenv("ETC_CORPORATE_ACCOUNTS", "test1:pass1")
	defer os.Unsetenv("ETC_CORPORATE_ACCOUNTS")

	logger := log.New(os.Stdout, "[TEST] ", log.LstdFlags)
	service := services.NewDownloadServiceGRPC(nil, logger)
	ctx := context.Background()

	// Create a job
	createReq := &pb.DownloadRequest{
		Accounts: []string{"test1:pass1"},
		FromDate: "2024-01-01",
		ToDate:   "2024-01-31",
	}

	createResp, err := service.DownloadAsync(ctx, createReq)
	if err != nil {
		t.Fatalf("Failed to create job: %v", err)
	}

	jobID := createResp.JobId

	// Test 1: Get status immediately (CompletedAt should be nil)
	t.Run("processing job without CompletedAt", func(t *testing.T) {
		req := &pb.GetJobStatusRequest{JobId: jobID}
		resp, err := service.GetJobStatus(ctx, req)
		if err != nil {
			t.Fatalf("GetJobStatus failed: %v", err)
		}

		if resp == nil {
			t.Fatal("Expected non-nil response")
		}

		// Job should be processing, so CompletedAt should be nil
		if resp.Status == "processing" && resp.CompletedAt != nil {
			t.Error("Expected CompletedAt to be nil for processing job")
		}

		t.Logf("Job status: %s, has CompletedAt: %v", resp.Status, resp.CompletedAt != nil)
	})

	// Wait for job to complete (download service with mock will complete quickly)
	time.Sleep(4 * time.Second)

	// Test 2: Get status after completion (CompletedAt should be set)
	t.Run("completed job with CompletedAt", func(t *testing.T) {
		req := &pb.GetJobStatusRequest{JobId: jobID}
		resp, err := service.GetJobStatus(ctx, req)
		if err != nil {
			t.Fatalf("GetJobStatus failed: %v", err)
		}

		if resp == nil {
			t.Fatal("Expected non-nil response")
		}

		t.Logf("Job status after wait: %s, has CompletedAt: %v", resp.Status, resp.CompletedAt != nil)

		// If job is completed, CompletedAt should be set
		if resp.Status == "completed" {
			if resp.CompletedAt == nil {
				t.Error("Expected CompletedAt to be set for completed job")
			} else {
				// Verify CompletedAt is after StartedAt
				if resp.StartedAt != nil {
					startTime := resp.StartedAt.AsTime()
					completeTime := resp.CompletedAt.AsTime()
					if !completeTime.After(startTime) && !completeTime.Equal(startTime) {
						t.Error("CompletedAt should be after or equal to StartedAt")
					}
				}
				t.Log("✓ CompletedAt is properly set for completed job")
			}
		} else if resp.Status == "failed" {
			// Failed jobs should also have CompletedAt
			if resp.CompletedAt == nil {
				t.Error("Expected CompletedAt to be set for failed job")
			} else {
				t.Log("✓ CompletedAt is properly set for failed job")
			}
		}

		// Verify the branch where CompletedAt != nil is executed
		if resp.CompletedAt != nil {
			t.Log("✓ Branch 'if job.CompletedAt != nil' was executed")
		}
	})

	// Test 3: Non-existent job (ensures nil response path)
	t.Run("non-existent job returns nil", func(t *testing.T) {
		req := &pb.GetJobStatusRequest{JobId: "non-existent-job-12345"}
		resp, err := service.GetJobStatus(ctx, req)
		if err != nil {
			t.Fatalf("GetJobStatus failed: %v", err)
		}

		if resp != nil {
			t.Error("Expected nil response for non-existent job")
		}
	})
}
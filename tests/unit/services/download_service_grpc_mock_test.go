package services_test

import (
	"context"
	"testing"
	"time"

	pb "github.com/yhonda-ohishi/etc_meisai/src/pb"
	"github.com/yhonda-ohishi/etc_meisai/src/services"
)

// MockDownloadService implements DownloadServiceInterface for testing
type MockDownloadService struct {
	jobs map[string]*services.DownloadJob
}

func NewMockDownloadService() *MockDownloadService {
	return &MockDownloadService{
		jobs: make(map[string]*services.DownloadJob),
	}
}

func (m *MockDownloadService) GetAllAccountIDs() []string {
	return []string{"test1", "test2"}
}

func (m *MockDownloadService) ProcessAsync(jobID string, accounts []string, fromDate, toDate string) {
	now := time.Now()
	m.jobs[jobID] = &services.DownloadJob{
		ID:           jobID,
		Status:       "processing",
		Progress:     50,
		TotalRecords: 100,
		StartedAt:    now,
	}
}

func (m *MockDownloadService) GetJobStatus(jobID string) (*services.DownloadJob, bool) {
	job, exists := m.jobs[jobID]
	return job, exists
}

// TestDownloadServiceGRPC_GetJobStatus_WithMock tests GetJobStatus with mocked downloadService
func TestDownloadServiceGRPC_GetJobStatus_WithMock(t *testing.T) {
	ctx := context.Background()

	// Test case 1: Job with CompletedAt set (completed job)
	t.Run("completed job with CompletedAt", func(t *testing.T) {
		// Create mock download service with a completed job
		mockService := NewMockDownloadService()
		now := time.Now()
		completedTime := now.Add(1 * time.Hour)

		mockService.jobs["completed-job"] = &services.DownloadJob{
			ID:           "completed-job",
			Status:       "completed",
			Progress:     100,
			TotalRecords: 50,
			StartedAt:    now,
			CompletedAt:  &completedTime, // CompletedAt is set
			ErrorMessage: "",
		}

		// Create gRPC service using the new constructor with mock
		grpcService := services.NewDownloadServiceGRPCWithMock(mockService)

		req := &pb.GetJobStatusRequest{JobId: "completed-job"}
		resp, err := grpcService.GetJobStatus(ctx, req)
		if err != nil {
			t.Fatalf("GetJobStatus failed: %v", err)
		}

		if resp == nil {
			t.Fatal("Expected non-nil response")
		}

		// Verify CompletedAt is set in response
		if resp.CompletedAt == nil {
			t.Error("Expected CompletedAt to be set for completed job")
		} else {
			t.Log("✓ CompletedAt is properly set in response")

			// Verify the timestamp is correct
			respTime := resp.CompletedAt.AsTime()
			if !respTime.Equal(completedTime) {
				t.Errorf("CompletedAt mismatch: expected %v, got %v", completedTime, respTime)
			}
		}
	})

	// Test case 2: Job with CompletedAt nil (processing job)
	t.Run("processing job without CompletedAt", func(t *testing.T) {
		mockService := NewMockDownloadService()
		now := time.Now()

		mockService.jobs["processing-job"] = &services.DownloadJob{
			ID:           "processing-job",
			Status:       "processing",
			Progress:     50,
			TotalRecords: 100,
			StartedAt:    now,
			CompletedAt:  nil, // CompletedAt is nil
		}

		grpcService := services.NewDownloadServiceGRPCWithMock(mockService)

		req := &pb.GetJobStatusRequest{JobId: "processing-job"}
		resp, err := grpcService.GetJobStatus(ctx, req)
		if err != nil {
			t.Fatalf("GetJobStatus failed: %v", err)
		}

		if resp == nil {
			t.Fatal("Expected non-nil response")
		}

		// Verify CompletedAt is nil in response
		if resp.CompletedAt != nil {
			t.Error("Expected CompletedAt to be nil for processing job")
		} else {
			t.Log("✓ CompletedAt is properly nil for processing job")
		}
	})

	// Test case 3: Non-existent job
	t.Run("non-existent job", func(t *testing.T) {
		mockService := NewMockDownloadService()

		grpcService := services.NewDownloadServiceGRPCWithMock(mockService)

		req := &pb.GetJobStatusRequest{JobId: "non-existent"}
		resp, err := grpcService.GetJobStatus(ctx, req)
		if err != nil {
			t.Fatalf("GetJobStatus failed: %v", err)
		}

		// Should return nil for non-existent job
		if resp != nil {
			t.Error("Expected nil response for non-existent job")
		}
	})

	// Test case 4: Job with error message and CompletedAt
	t.Run("failed job with CompletedAt", func(t *testing.T) {
		mockService := NewMockDownloadService()
		now := time.Now()
		failedTime := now.Add(30 * time.Minute)

		mockService.jobs["failed-job"] = &services.DownloadJob{
			ID:           "failed-job",
			Status:       "failed",
			Progress:     75,
			TotalRecords: 0,
			StartedAt:    now,
			CompletedAt:  &failedTime, // CompletedAt is set for failed job
			ErrorMessage: "Download failed due to network error",
		}

		grpcService := services.NewDownloadServiceGRPCWithMock(mockService)

		req := &pb.GetJobStatusRequest{JobId: "failed-job"}
		resp, err := grpcService.GetJobStatus(ctx, req)
		if err != nil {
			t.Fatalf("GetJobStatus failed: %v", err)
		}

		if resp == nil {
			t.Fatal("Expected non-nil response")
		}

		// Verify both CompletedAt and ErrorMessage are set
		if resp.CompletedAt == nil {
			t.Error("Expected CompletedAt to be set for failed job")
		}

		if resp.ErrorMessage == "" {
			t.Error("Expected ErrorMessage to be set for failed job")
		}

		t.Log("✓ Both CompletedAt and ErrorMessage are properly set for failed job")
	})
}
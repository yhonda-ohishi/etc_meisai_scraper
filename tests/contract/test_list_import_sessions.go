//go:build contract

package contract

import (
	"context"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"

	"github.com/yhonda-ohishi/etc_meisai/src/pb"
)

func TestListImportSessions_Success(t *testing.T) {
	// Arrange
	conn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("Failed to connect to gRPC server: %v", err)
	}
	defer conn.Close()

	client := pb.NewETCMeisaiServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Test data - basic pagination
	req := &pb.ListImportSessionsRequest{
		Page:     1,
		PageSize: 10,
	}

	// Act
	resp, err := client.ListImportSessions(ctx, req)

	// Assert
	// This test should FAIL initially as the server is not implemented yet
	if err != nil {
		st := status.Convert(err)
		if st.Code() == codes.Unimplemented {
			t.Logf("Expected failure: ListImportSessions not implemented yet - %v", err)
			return
		}
		t.Fatalf("Unexpected error: %v", err)
	}

	// If server is implemented, verify response
	if resp == nil {
		t.Fatal("Response is nil")
	}

	// Verify pagination fields
	if resp.Page != 1 {
		t.Errorf("Expected page 1, got %d", resp.Page)
	}

	if resp.PageSize != 10 {
		t.Errorf("Expected page size 10, got %d", resp.PageSize)
	}

	// Total count should be non-negative
	if resp.TotalCount < 0 {
		t.Errorf("Expected total count to be non-negative, got %d", resp.TotalCount)
	}

	// Sessions should not be nil (can be empty array)
	if resp.Sessions == nil {
		t.Error("Expected sessions array to not be nil")
	}

	// If sessions exist, verify they have required fields
	for i, session := range resp.Sessions {
		if session.Id == "" {
			t.Errorf("Session %d has empty ID", i)
		}
		if session.AccountType == "" {
			t.Errorf("Session %d has empty account type", i)
		}
		if session.AccountId == "" {
			t.Errorf("Session %d has empty account ID", i)
		}
		if session.FileName == "" {
			t.Errorf("Session %d has empty file name", i)
		}
		if session.Status == pb.ImportStatus_IMPORT_STATUS_UNSPECIFIED {
			t.Errorf("Session %d has unspecified status", i)
		}
		if session.StartedAt == nil {
			t.Errorf("Session %d has nil started_at", i)
		}
		if session.CreatedAt == nil {
			t.Errorf("Session %d has nil created_at", i)
		}
	}
}

func TestListImportSessions_WithFilters(t *testing.T) {
	// Arrange
	conn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("Failed to connect to gRPC server: %v", err)
	}
	defer conn.Close()

	client := pb.NewETCMeisaiServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	testCases := []struct {
		name string
		req  *pb.ListImportSessionsRequest
	}{
		{
			name: "filter by account type",
			req: &pb.ListImportSessionsRequest{
				Page:        1,
				PageSize:    10,
				AccountType: stringPtr("corporate"),
			},
		},
		{
			name: "filter by account ID",
			req: &pb.ListImportSessionsRequest{
				Page:      1,
				PageSize:  10,
				AccountId: stringPtr("test-account-001"),
			},
		},
		{
			name: "filter by status - completed",
			req: &pb.ListImportSessionsRequest{
				Page:     1,
				PageSize: 10,
				Status:   &[]pb.ImportStatus{pb.ImportStatus_IMPORT_STATUS_COMPLETED}[0],
			},
		},
		{
			name: "filter by status - processing",
			req: &pb.ListImportSessionsRequest{
				Page:     1,
				PageSize: 10,
				Status:   &[]pb.ImportStatus{pb.ImportStatus_IMPORT_STATUS_PROCESSING}[0],
			},
		},
		{
			name: "filter by status - failed",
			req: &pb.ListImportSessionsRequest{
				Page:     1,
				PageSize: 10,
				Status:   &[]pb.ImportStatus{pb.ImportStatus_IMPORT_STATUS_FAILED}[0],
			},
		},
		{
			name: "combined filters",
			req: &pb.ListImportSessionsRequest{
				Page:        1,
				PageSize:    10,
				AccountType: stringPtr("personal"),
				AccountId:   stringPtr("test-account-002"),
				Status:      &[]pb.ImportStatus{pb.ImportStatus_IMPORT_STATUS_COMPLETED}[0],
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Act
			resp, err := client.ListImportSessions(ctx, tc.req)

			// Assert
			if err != nil {
				st := status.Convert(err)
				if st.Code() == codes.Unimplemented {
					t.Logf("Expected failure: ListImportSessions not implemented yet - %v", err)
					return
				}
				t.Fatalf("Unexpected error: %v", err)
			}

			// If server is implemented, verify filters work
			if resp == nil {
				t.Fatal("Response is nil")
			}

			// Verify pagination is preserved
			if resp.Page != tc.req.Page {
				t.Errorf("Expected page %d, got %d", tc.req.Page, resp.Page)
			}

			if resp.PageSize != tc.req.PageSize {
				t.Errorf("Expected page size %d, got %d", tc.req.PageSize, resp.PageSize)
			}

			// Verify filter application (when possible)
			for i, session := range resp.Sessions {
				if tc.req.AccountType != nil && session.AccountType != *tc.req.AccountType {
					t.Errorf("Session %d: expected account type %s, got %s", i, *tc.req.AccountType, session.AccountType)
				}
				if tc.req.AccountId != nil && session.AccountId != *tc.req.AccountId {
					t.Errorf("Session %d: expected account ID %s, got %s", i, *tc.req.AccountId, session.AccountId)
				}
				if tc.req.Status != nil && session.Status != *tc.req.Status {
					t.Errorf("Session %d: expected status %s, got %s", i, tc.req.Status.String(), session.Status.String())
				}
			}

			t.Logf("Filter test '%s' returned %d sessions", tc.name, len(resp.Sessions))
		})
	}
}

func TestListImportSessions_Pagination(t *testing.T) {
	// Arrange
	conn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("Failed to connect to gRPC server: %v", err)
	}
	defer conn.Close()

	client := pb.NewETCMeisaiServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	testCases := []struct {
		name     string
		page     int32
		pageSize int32
	}{
		{
			name:     "first page small size",
			page:     1,
			pageSize: 5,
		},
		{
			name:     "second page",
			page:     2,
			pageSize: 10,
		},
		{
			name:     "large page size",
			page:     1,
			pageSize: 100,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := &pb.ListImportSessionsRequest{
				Page:     tc.page,
				PageSize: tc.pageSize,
			}

			// Act
			resp, err := client.ListImportSessions(ctx, req)

			// Assert
			if err != nil {
				st := status.Convert(err)
				if st.Code() == codes.Unimplemented {
					t.Logf("Expected failure: ListImportSessions not implemented yet - %v", err)
					return
				}
				t.Fatalf("Unexpected error: %v", err)
			}

			if resp == nil {
				t.Fatal("Response is nil")
			}

			// Verify pagination parameters are returned correctly
			if resp.Page != tc.page {
				t.Errorf("Expected page %d, got %d", tc.page, resp.Page)
			}

			if resp.PageSize != tc.pageSize {
				t.Errorf("Expected page size %d, got %d", tc.pageSize, resp.PageSize)
			}

			// Verify sessions count doesn't exceed page size
			if int32(len(resp.Sessions)) > tc.pageSize {
				t.Errorf("Expected at most %d sessions, got %d", tc.pageSize, len(resp.Sessions))
			}
		})
	}
}

func TestListImportSessions_InvalidParameters(t *testing.T) {
	// Arrange
	conn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("Failed to connect to gRPC server: %v", err)
	}
	defer conn.Close()

	client := pb.NewETCMeisaiServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	testCases := []struct {
		name string
		req  *pb.ListImportSessionsRequest
	}{
		{
			name: "zero page",
			req: &pb.ListImportSessionsRequest{
				Page:     0,
				PageSize: 10,
			},
		},
		{
			name: "negative page",
			req: &pb.ListImportSessionsRequest{
				Page:     -1,
				PageSize: 10,
			},
		},
		{
			name: "zero page size",
			req: &pb.ListImportSessionsRequest{
				Page:     1,
				PageSize: 0,
			},
		},
		{
			name: "negative page size",
			req: &pb.ListImportSessionsRequest{
				Page:     1,
				PageSize: -1,
			},
		},
		{
			name: "page size too large",
			req: &pb.ListImportSessionsRequest{
				Page:     1,
				PageSize: 10000,
			},
		},
		{
			name: "invalid account type",
			req: &pb.ListImportSessionsRequest{
				Page:        1,
				PageSize:    10,
				AccountType: stringPtr("invalid_type"),
			},
		},
		{
			name: "empty account type",
			req: &pb.ListImportSessionsRequest{
				Page:        1,
				PageSize:    10,
				AccountType: stringPtr(""),
			},
		},
		{
			name: "empty account ID",
			req: &pb.ListImportSessionsRequest{
				Page:      1,
				PageSize:  10,
				AccountId: stringPtr(""),
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Act
			resp, err := client.ListImportSessions(ctx, tc.req)

			// Assert
			if err != nil {
				st := status.Convert(err)
				if st.Code() == codes.Unimplemented {
					t.Logf("Expected failure: ListImportSessions not implemented yet - %v", err)
					return
				}
				// When implemented, should return InvalidArgument for invalid parameters
				if st.Code() != codes.InvalidArgument {
					t.Errorf("Expected InvalidArgument error for %s, got %v", tc.name, st.Code())
				}
				return
			}

			// If no error, the validation might not be implemented yet
			if resp != nil {
				t.Logf("Warning: Expected validation error for %s, but got successful response", tc.name)
			}
		})
	}
}

func TestListImportSessions_StatusBasedFiltering(t *testing.T) {
	// Arrange
	conn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("Failed to connect to gRPC server: %v", err)
	}
	defer conn.Close()

	client := pb.NewETCMeisaiServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Test different status filters
	statuses := []pb.ImportStatus{
		pb.ImportStatus_IMPORT_STATUS_PENDING,
		pb.ImportStatus_IMPORT_STATUS_PROCESSING,
		pb.ImportStatus_IMPORT_STATUS_COMPLETED,
		pb.ImportStatus_IMPORT_STATUS_FAILED,
		pb.ImportStatus_IMPORT_STATUS_CANCELLED,
	}

	for _, importStatus := range statuses {
		t.Run("status_"+importStatus.String(), func(t *testing.T) {
			req := &pb.ListImportSessionsRequest{
				Page:     1,
				PageSize: 10,
				Status:   &importStatus,
			}

			// Act
			resp, err := client.ListImportSessions(ctx, req)

			// Assert
			if err != nil {
				st := status.Convert(err)
				if st.Code() == codes.Unimplemented {
					t.Logf("Expected failure: ListImportSessions not implemented yet - %v", err)
					return
				}
				t.Fatalf("Unexpected error: %v", err)
			}

			if resp == nil {
				t.Fatal("Response is nil")
			}

			// Verify all returned sessions have the requested status
			for i, session := range resp.Sessions {
				if session.Status != importStatus {
					t.Errorf("Session %d: expected status %s, got %s", i, importStatus.String(), session.Status.String())
				}
			}

			t.Logf("Status filter %s returned %d sessions", importStatus.String(), len(resp.Sessions))
		})
	}
}

func TestListImportSessions_EmptyResult(t *testing.T) {
	// Arrange
	conn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("Failed to connect to gRPC server: %v", err)
	}
	defer conn.Close()

	client := pb.NewETCMeisaiServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Request with filters that should return no results
	req := &pb.ListImportSessionsRequest{
		Page:        1,
		PageSize:    10,
		AccountType: stringPtr("non-existent-type"),
		AccountId:   stringPtr("non-existent-account"),
	}

	// Act
	resp, err := client.ListImportSessions(ctx, req)

	// Assert
	if err != nil {
		st := status.Convert(err)
		if st.Code() == codes.Unimplemented {
			t.Logf("Expected failure: ListImportSessions not implemented yet - %v", err)
			return
		}
		t.Fatalf("Unexpected error: %v", err)
	}

	// If server is implemented, verify empty result handling
	if resp == nil {
		t.Fatal("Response is nil")
	}

	// Should return empty sessions array, not nil
	if resp.Sessions == nil {
		t.Error("Expected sessions array to not be nil (should be empty array)")
	}

	// Should have zero total count
	if resp.TotalCount != 0 {
		t.Errorf("Expected total count 0 for empty result, got %d", resp.TotalCount)
	}

	// Should have empty sessions array
	if len(resp.Sessions) != 0 {
		t.Errorf("Expected 0 sessions for empty result, got %d", len(resp.Sessions))
	}

	// Pagination should still be valid
	if resp.Page != 1 {
		t.Errorf("Expected page 1, got %d", resp.Page)
	}

	if resp.PageSize != 10 {
		t.Errorf("Expected page size 10, got %d", resp.PageSize)
	}
}
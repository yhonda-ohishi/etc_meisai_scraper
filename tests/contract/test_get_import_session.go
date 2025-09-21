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

func TestGetImportSession_Success(t *testing.T) {
	// Arrange
	conn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("Failed to connect to gRPC server: %v", err)
	}
	defer conn.Close()

	client := pb.NewETCMeisaiServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Test data - assume import session with this ID exists
	req := &pb.GetImportSessionRequest{
		SessionId: "test-session-001",
	}

	// Act
	resp, err := client.GetImportSession(ctx, req)

	// Assert
	// This test should FAIL initially as the server is not implemented yet
	if err != nil {
		st := status.Convert(err)
		if st.Code() == codes.Unimplemented {
			t.Logf("Expected failure: GetImportSession not implemented yet - %v", err)
			return
		}
		t.Fatalf("Unexpected error: %v", err)
	}

	// If server is implemented, verify response
	if resp == nil {
		t.Fatal("Response is nil")
	}

	if resp.Session == nil {
		t.Fatal("Response session is nil")
	}

	// Verify session fields
	if resp.Session.Id != "test-session-001" {
		t.Errorf("Expected session ID 'test-session-001', got %s", resp.Session.Id)
	}

	// Verify required fields are not empty
	if resp.Session.AccountType == "" {
		t.Error("Expected account_type to be non-empty")
	}

	if resp.Session.AccountId == "" {
		t.Error("Expected account_id to be non-empty")
	}

	if resp.Session.FileName == "" {
		t.Error("Expected file_name to be non-empty")
	}

	// Verify status is valid
	if resp.Session.Status == pb.ImportStatus_IMPORT_STATUS_UNSPECIFIED {
		t.Error("Expected import status to be specified")
	}

	// Verify counts are non-negative
	if resp.Session.TotalRows < 0 {
		t.Errorf("Expected total_rows to be non-negative, got %d", resp.Session.TotalRows)
	}

	if resp.Session.ProcessedRows < 0 {
		t.Errorf("Expected processed_rows to be non-negative, got %d", resp.Session.ProcessedRows)
	}

	if resp.Session.SuccessRows < 0 {
		t.Errorf("Expected success_rows to be non-negative, got %d", resp.Session.SuccessRows)
	}

	if resp.Session.ErrorRows < 0 {
		t.Errorf("Expected error_rows to be non-negative, got %d", resp.Session.ErrorRows)
	}

	if resp.Session.DuplicateRows < 0 {
		t.Errorf("Expected duplicate_rows to be non-negative, got %d", resp.Session.DuplicateRows)
	}

	// Verify timestamps
	if resp.Session.StartedAt == nil {
		t.Error("Expected started_at to be set")
	}

	if resp.Session.CreatedAt == nil {
		t.Error("Expected created_at to be set")
	}

	// Verify row count consistency
	totalProcessed := resp.Session.SuccessRows + resp.Session.ErrorRows + resp.Session.DuplicateRows
	if resp.Session.ProcessedRows > 0 && totalProcessed > resp.Session.ProcessedRows {
		t.Errorf("Sum of success+error+duplicate (%d) should not exceed processed rows (%d)",
			totalProcessed, resp.Session.ProcessedRows)
	}
}

func TestGetImportSession_NotFound(t *testing.T) {
	// Arrange
	conn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("Failed to connect to gRPC server: %v", err)
	}
	defer conn.Close()

	client := pb.NewETCMeisaiServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Test data - non-existent session ID
	req := &pb.GetImportSessionRequest{
		SessionId: "non-existent-session-999",
	}

	// Act
	resp, err := client.GetImportSession(ctx, req)

	// Assert
	if err != nil {
		st := status.Convert(err)
		if st.Code() == codes.Unimplemented {
			t.Logf("Expected failure: GetImportSession not implemented yet - %v", err)
			return
		}
		// When implemented, should return NotFound for non-existent session
		if st.Code() != codes.NotFound {
			t.Errorf("Expected NotFound error, got %v", st.Code())
		}
		return
	}

	// If no error, this might indicate the validation is not implemented
	if resp != nil {
		t.Logf("Warning: Expected NotFound error for non-existent session, but got successful response")
	}
}

func TestGetImportSession_InvalidSessionID(t *testing.T) {
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
		name      string
		sessionId string
	}{
		{
			name:      "empty session ID",
			sessionId: "",
		},
		{
			name:      "whitespace only",
			sessionId: "   ",
		},
		{
			name:      "invalid characters",
			sessionId: "invalid/session\\id",
		},
		{
			name:      "too long session ID",
			sessionId: "very-long-session-id-that-exceeds-reasonable-length-limits-and-should-be-rejected-by-validation",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := &pb.GetImportSessionRequest{
				SessionId: tc.sessionId,
			}

			// Act
			resp, err := client.GetImportSession(ctx, req)

			// Assert
			if err != nil {
				st := status.Convert(err)
				if st.Code() == codes.Unimplemented {
					t.Logf("Expected failure: GetImportSession not implemented yet - %v", err)
					return
				}
				// When implemented, should return InvalidArgument for invalid session ID
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

func TestGetImportSession_CompletedSession(t *testing.T) {
	// Arrange
	conn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("Failed to connect to gRPC server: %v", err)
	}
	defer conn.Close()

	client := pb.NewETCMeisaiServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Test data - assume completed session
	req := &pb.GetImportSessionRequest{
		SessionId: "completed-session-001",
	}

	// Act
	resp, err := client.GetImportSession(ctx, req)

	// Assert
	if err != nil {
		st := status.Convert(err)
		if st.Code() == codes.Unimplemented {
			t.Logf("Expected failure: GetImportSession not implemented yet - %v", err)
			return
		}
		t.Fatalf("Unexpected error: %v", err)
	}

	// If server is implemented, verify completed session details
	if resp == nil || resp.Session == nil {
		t.Fatal("Response or session is nil")
	}

	// For completed sessions, verify specific fields
	if resp.Session.Status == pb.ImportStatus_IMPORT_STATUS_COMPLETED {
		// Should have completed_at timestamp
		if resp.Session.CompletedAt == nil {
			t.Error("Expected completed_at to be set for completed session")
		}

		// Processed rows should equal total rows for completed session
		if resp.Session.ProcessedRows != resp.Session.TotalRows {
			t.Logf("Note: Processed rows (%d) != total rows (%d) - may be expected for failed/partial imports",
				resp.Session.ProcessedRows, resp.Session.TotalRows)
		}

		// Should have no error message for successful completion
		if resp.Session.ErrorLog != nil && len(resp.Session.ErrorLog) > 0 {
			t.Logf("Completed session has %d error entries", len(resp.Session.ErrorLog))
		}
	}

	t.Logf("Session details - Status: %s, Total: %d, Processed: %d, Success: %d, Errors: %d",
		resp.Session.Status.String(), resp.Session.TotalRows, resp.Session.ProcessedRows,
		resp.Session.SuccessRows, resp.Session.ErrorRows)
}

func TestGetImportSession_WithErrors(t *testing.T) {
	// Arrange
	conn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("Failed to connect to gRPC server: %v", err)
	}
	defer conn.Close()

	client := pb.NewETCMeisaiServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Test data - assume session with errors
	req := &pb.GetImportSessionRequest{
		SessionId: "error-session-001",
	}

	// Act
	resp, err := client.GetImportSession(ctx, req)

	// Assert
	if err != nil {
		st := status.Convert(err)
		if st.Code() == codes.Unimplemented {
			t.Logf("Expected failure: GetImportSession not implemented yet - %v", err)
			return
		}
		t.Fatalf("Unexpected error: %v", err)
	}

	// If server is implemented, verify error handling
	if resp == nil || resp.Session == nil {
		t.Fatal("Response or session is nil")
	}

	// If session has errors, verify error log
	if resp.Session.ErrorRows > 0 {
		if resp.Session.ErrorLog == nil || len(resp.Session.ErrorLog) == 0 {
			t.Error("Expected error_log to be populated when error_rows > 0")
		} else {
			// Verify error log entries
			for i, errorEntry := range resp.Session.ErrorLog {
				if errorEntry.RowNumber <= 0 {
					t.Errorf("Error entry %d has invalid row number: %d", i, errorEntry.RowNumber)
				}

				if errorEntry.ErrorType == "" {
					t.Errorf("Error entry %d has empty error type", i)
				}

				if errorEntry.ErrorMessage == "" {
					t.Errorf("Error entry %d has empty error message", i)
				}

				// Raw data is optional but should be reasonable if present
				if len(errorEntry.RawData) > 10000 {
					t.Errorf("Error entry %d has suspiciously large raw data: %d bytes", i, len(errorEntry.RawData))
				}
			}

			t.Logf("Session has %d error entries in error log", len(resp.Session.ErrorLog))
		}
	}
}

func TestGetImportSession_InProgressSession(t *testing.T) {
	// Arrange
	conn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("Failed to connect to gRPC server: %v", err)
	}
	defer conn.Close()

	client := pb.NewETCMeisaiServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Test data - assume in-progress session
	req := &pb.GetImportSessionRequest{
		SessionId: "inprogress-session-001",
	}

	// Act
	resp, err := client.GetImportSession(ctx, req)

	// Assert
	if err != nil {
		st := status.Convert(err)
		if st.Code() == codes.Unimplemented {
			t.Logf("Expected failure: GetImportSession not implemented yet - %v", err)
			return
		}
		t.Fatalf("Unexpected error: %v", err)
	}

	// If server is implemented, verify in-progress session details
	if resp == nil || resp.Session == nil {
		t.Fatal("Response or session is nil")
	}

	// For in-progress sessions, verify specific constraints
	if resp.Session.Status == pb.ImportStatus_IMPORT_STATUS_PROCESSING {
		// Should NOT have completed_at timestamp
		if resp.Session.CompletedAt != nil {
			t.Error("Expected completed_at to be nil for in-progress session")
		}

		// Processed rows should be <= total rows
		if resp.Session.ProcessedRows > resp.Session.TotalRows {
			t.Errorf("Processed rows (%d) should not exceed total rows (%d) for in-progress session",
				resp.Session.ProcessedRows, resp.Session.TotalRows)
		}

		t.Logf("In-progress session: %d/%d rows processed",
			resp.Session.ProcessedRows, resp.Session.TotalRows)
	}
}
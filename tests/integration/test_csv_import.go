//go:build integration

package integration

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	pb "github.com/yhonda-ohishi/etc_meisai/src/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

// TestCSVImportWorkflow tests the complete CSV import workflow
// This integration test verifies:
// 1. Prepare CSV data
// 2. Import CSV file
// 3. Check import session status
// 4. Verify imported records
// 5. Handle duplicate detection
// 6. Error handling for invalid CSV
func TestCSVImportWorkflow(t *testing.T) {
	conn, client := setupGRPCClient(t)
	defer conn.Close()

	ctx := context.Background()

	// Test data: prepare valid CSV content
	validCSVContent := createValidCSVContent()

	var sessionId string

	// Step 1: Import valid CSV file
	t.Run("ImportValidCSV", func(t *testing.T) {
		importReq := &pb.ImportCSVRequest{
			AccountType: "corporate",
			AccountId:   "test_account_001",
			FileName:    "test_etc_data.csv",
			FileContent: []byte(validCSVContent),
		}

		importResp, err := client.ImportCSV(ctx, importReq)
		if err != nil {
			t.Fatalf("Failed to import CSV: %v", err)
		}

		if importResp.Session == nil {
			t.Fatal("Expected import session in response, got nil")
		}

		sessionId = importResp.Session.SessionId
		if sessionId == "" {
			t.Fatal("Expected non-empty session ID")
		}

		// Verify session status
		if importResp.Session.Status == "" {
			t.Fatal("Expected session status, got empty string")
		}

		t.Logf("Import session created: %s", sessionId)
	})

	// Step 2: Check import session status
	t.Run("CheckImportSessionStatus", func(t *testing.T) {
		// Wait a bit for processing
		time.Sleep(2 * time.Second)

		getSessionReq := &pb.GetImportSessionRequest{
			SessionId: sessionId,
		}

		sessionResp, err := client.GetImportSession(ctx, getSessionReq)
		if err != nil {
			t.Fatalf("Failed to get import session: %v", err)
		}

		if sessionResp.Session == nil {
			t.Fatal("Expected session in response, got nil")
		}

		session := sessionResp.Session

		// Verify session has progress information
		if session.TotalRows <= 0 {
			t.Errorf("Expected positive total rows, got %d", session.TotalRows)
		}

		// The session should be completed or in progress
		validStatuses := []string{"completed", "processing", "in_progress"}
		isValidStatus := false
		for _, status := range validStatuses {
			if session.Status == status {
				isValidStatus = true
				break
			}
		}

		if !isValidStatus {
			t.Errorf("Expected valid session status, got: %s", session.Status)
		}

		t.Logf("Session status: %s, Total rows: %d, Processed: %d, Success: %d, Errors: %d",
			session.Status, session.TotalRows, session.ProcessedRows, session.SuccessRows, session.ErrorRows)
	})

	// Step 3: Verify imported records exist in the system
	t.Run("VerifyImportedRecords", func(t *testing.T) {
		// Wait for import to complete
		time.Sleep(3 * time.Second)

		// Get session final status
		getSessionReq := &pb.GetImportSessionRequest{
			SessionId: sessionId,
		}

		sessionResp, err := client.GetImportSession(ctx, getSessionReq)
		if err != nil {
			t.Fatalf("Failed to get final session status: %v", err)
		}

		session := sessionResp.Session
		expectedRows := int32(3) // We created 3 valid rows in CSV

		if session.TotalRows != expectedRows {
			t.Errorf("Expected %d total rows, got %d", expectedRows, session.TotalRows)
		}

		if session.SuccessRows == 0 {
			t.Fatalf("Expected some successful imports, got 0")
		}

		// List records to verify they were imported
		listReq := &pb.ListRecordsRequest{
			Page:     1,
			PageSize: 10,
			DateFrom: stringPtr("2025-09-20"),
			DateTo:   stringPtr("2025-09-22"),
		}

		listResp, err := client.ListRecords(ctx, listReq)
		if err != nil {
			t.Fatalf("Failed to list records: %v", err)
		}

		if len(listResp.Records) == 0 {
			t.Fatal("Expected to find imported records, but got none")
		}

		// Verify specific imported records
		expectedCarNumbers := []string{"品川 300 あ 1234", "横浜 500 か 5678", "東京 700 さ 9012"}
		foundRecords := make(map[string]bool)

		for _, record := range listResp.Records {
			for _, expectedCarNumber := range expectedCarNumbers {
				if record.CarNumber == expectedCarNumber {
					foundRecords[expectedCarNumber] = true
				}
			}
		}

		for _, expectedCarNumber := range expectedCarNumbers {
			if !foundRecords[expectedCarNumber] {
				t.Errorf("Expected to find record with car number %s", expectedCarNumber)
			}
		}

		t.Logf("Successfully verified %d imported records", len(foundRecords))
	})
}

// TestCSVImportDuplicateHandling tests duplicate detection in CSV import
func TestCSVImportDuplicateHandling(t *testing.T) {
	conn, client := setupGRPCClient(t)
	defer conn.Close()

	ctx := context.Background()

	// Step 1: Import initial data
	csvContent1 := `利用日,利用時刻,入口IC,出口IC,通行料金,車両番号,ETCカード番号
2025-09-21,10:30:00,東京IC,横浜IC,1200,品川 300 あ 1234,1234567890123456`

	importReq1 := &pb.ImportCSVRequest{
		AccountType: "corporate",
		AccountId:   "test_duplicate_001",
		FileName:    "first_import.csv",
		FileContent: []byte(csvContent1),
	}

	importResp1, err := client.ImportCSV(ctx, importReq1)
	if err != nil {
		t.Fatalf("Failed to import first CSV: %v", err)
	}

	sessionId1 := importResp1.Session.SessionId
	t.Logf("First import session: %s", sessionId1)

	// Wait for first import to complete
	time.Sleep(3 * time.Second)

	// Step 2: Import duplicate data
	csvContent2 := `利用日,利用時刻,入口IC,出口IC,通行料金,車両番号,ETCカード番号
2025-09-21,10:30:00,東京IC,横浜IC,1200,品川 300 あ 1234,1234567890123456
2025-09-21,14:15:00,横浜IC,静岡IC,2500,品川 300 あ 1234,1234567890123456`

	importReq2 := &pb.ImportCSVRequest{
		AccountType: "corporate",
		AccountId:   "test_duplicate_001",
		FileName:    "second_import.csv",
		FileContent: []byte(csvContent2),
	}

	importResp2, err := client.ImportCSV(ctx, importReq2)
	if err != nil {
		t.Fatalf("Failed to import second CSV: %v", err)
	}

	sessionId2 := importResp2.Session.SessionId
	t.Logf("Second import session: %s", sessionId2)

	// Wait for second import to complete
	time.Sleep(3 * time.Second)

	// Step 3: Check duplicate handling
	getSessionReq := &pb.GetImportSessionRequest{
		SessionId: sessionId2,
	}

	sessionResp, err := client.GetImportSession(ctx, getSessionReq)
	if err != nil {
		t.Fatalf("Failed to get second session status: %v", err)
	}

	session := sessionResp.Session

	// The second import should have detected duplicates
	// First row is duplicate, second row should be new
	if session.TotalRows != 2 {
		t.Errorf("Expected 2 total rows in second import, got %d", session.TotalRows)
	}

	// Should have at least one successful import (the new record)
	if session.SuccessRows == 0 {
		t.Error("Expected at least one successful import")
	}

	// Check for duplicate errors or warnings in session metadata
	if session.ErrorRows > 0 || session.WarningCount > 0 {
		t.Logf("Duplicates detected - Error rows: %d, Warnings: %d", session.ErrorRows, session.WarningCount)
	}
}

// TestCSVImportErrorHandling tests error scenarios in CSV import
func TestCSVImportErrorHandling(t *testing.T) {
	conn, client := setupGRPCClient(t)
	defer conn.Close()

	ctx := context.Background()

	t.Run("ImportInvalidCSVFormat", func(t *testing.T) {
		// Invalid CSV with missing columns
		invalidCSV := `利用日,利用時刻,入口IC
2025-09-21,10:30:00,東京IC`

		importReq := &pb.ImportCSVRequest{
			AccountType: "corporate",
			AccountId:   "test_invalid_001",
			FileName:    "invalid.csv",
			FileContent: []byte(invalidCSV),
		}

		importResp, err := client.ImportCSV(ctx, importReq)
		if err != nil {
			// Might fail immediately for severely malformed CSV
			st, ok := status.FromError(err)
			if !ok {
				t.Fatalf("Expected gRPC status error, got: %v", err)
			}

			if st.Code() != codes.InvalidArgument {
				t.Fatalf("Expected INVALID_ARGUMENT error, got: %v", st.Code())
			}
			return
		}

		// If import started, check session for errors
		if importResp.Session != nil {
			time.Sleep(2 * time.Second)

			getSessionReq := &pb.GetImportSessionRequest{
				SessionId: importResp.Session.SessionId,
			}

			sessionResp, err := client.GetImportSession(ctx, getSessionReq)
			if err != nil {
				t.Fatalf("Failed to get session status: %v", err)
			}

			session := sessionResp.Session
			if session.ErrorRows == 0 && session.Status != "failed" {
				t.Error("Expected errors or failed status for invalid CSV")
			}
		}
	})

	t.Run("ImportWithInvalidData", func(t *testing.T) {
		// CSV with invalid data types and formats
		invalidDataCSV := `利用日,利用時刻,入口IC,出口IC,通行料金,車両番号,ETCカード番号
invalid-date,25:99:99,東京IC,横浜IC,not-a-number,invalid-car-number,invalid-card-number
2025-02-30,10:30:00,東京IC,横浜IC,-1000,品川 300 あ 1234,1234567890123456`

		importReq := &pb.ImportCSVRequest{
			AccountType: "corporate",
			AccountId:   "test_invalid_data_001",
			FileName:    "invalid_data.csv",
			FileContent: []byte(invalidDataCSV),
		}

		importResp, err := client.ImportCSV(ctx, importReq)
		if err != nil {
			t.Fatalf("Failed to start import: %v", err)
		}

		sessionId := importResp.Session.SessionId
		time.Sleep(3 * time.Second)

		// Check session status
		getSessionReq := &pb.GetImportSessionRequest{
			SessionId: sessionId,
		}

		sessionResp, err := client.GetImportSession(ctx, getSessionReq)
		if err != nil {
			t.Fatalf("Failed to get session status: %v", err)
		}

		session := sessionResp.Session

		// Should have errors due to invalid data
		if session.ErrorRows == 0 {
			t.Error("Expected error rows for invalid data, got 0")
		}

		// Should have processed all rows even if they had errors
		if session.ProcessedRows != session.TotalRows {
			t.Errorf("Expected all rows to be processed, got %d of %d", session.ProcessedRows, session.TotalRows)
		}

		t.Logf("Invalid data import - Total: %d, Processed: %d, Success: %d, Errors: %d",
			session.TotalRows, session.ProcessedRows, session.SuccessRows, session.ErrorRows)
	})

	t.Run("ImportEmptyCSV", func(t *testing.T) {
		emptyCSV := `利用日,利用時刻,入口IC,出口IC,通行料金,車両番号,ETCカード番号`

		importReq := &pb.ImportCSVRequest{
			AccountType: "corporate",
			AccountId:   "test_empty_001",
			FileName:    "empty.csv",
			FileContent: []byte(emptyCSV),
		}

		importResp, err := client.ImportCSV(ctx, importReq)
		if err != nil {
			// Might fail immediately for empty CSV
			return
		}

		if importResp.Session != nil {
			time.Sleep(1 * time.Second)

			getSessionReq := &pb.GetImportSessionRequest{
				SessionId: importResp.Session.SessionId,
			}

			sessionResp, err := client.GetImportSession(ctx, getSessionReq)
			if err != nil {
				t.Fatalf("Failed to get session status: %v", err)
			}

			session := sessionResp.Session
			if session.TotalRows != 0 {
				t.Errorf("Expected 0 total rows for empty CSV, got %d", session.TotalRows)
			}
		}
	})
}

// TestImportSessionManagement tests import session listing and management
func TestImportSessionManagement(t *testing.T) {
	conn, client := setupGRPCClient(t)
	defer conn.Close()

	ctx := context.Background()

	// Create multiple import sessions
	sessionIds := make([]string, 0)

	for i := 0; i < 3; i++ {
		csvContent := fmt.Sprintf(`利用日,利用時刻,入口IC,出口IC,通行料金,車両番号,ETCカード番号
2025-09-21,10:%02d:00,東京IC,横浜IC,1200,品川 300 あ %04d,1234567890123456`, i*10+30, 1000+i)

		importReq := &pb.ImportCSVRequest{
			AccountType: "corporate",
			AccountId:   fmt.Sprintf("test_session_mgmt_%03d", i),
			FileName:    fmt.Sprintf("session_test_%d.csv", i),
			FileContent: []byte(csvContent),
		}

		importResp, err := client.ImportCSV(ctx, importReq)
		if err != nil {
			t.Fatalf("Failed to import CSV %d: %v", i, err)
		}

		sessionIds = append(sessionIds, importResp.Session.SessionId)
	}

	// List import sessions
	t.Run("ListImportSessions", func(t *testing.T) {
		listReq := &pb.ListImportSessionsRequest{
			Page:     1,
			PageSize: 10,
		}

		listResp, err := client.ListImportSessions(ctx, listReq)
		if err != nil {
			t.Fatalf("Failed to list import sessions: %v", err)
		}

		if len(listResp.Sessions) == 0 {
			t.Fatal("Expected to find import sessions, got none")
		}

		// Verify our sessions are in the list
		foundSessions := make(map[string]bool)
		for _, session := range listResp.Sessions {
			foundSessions[session.SessionId] = true
		}

		for _, sessionId := range sessionIds {
			if !foundSessions[sessionId] {
				t.Errorf("Expected to find session %s in list", sessionId)
			}
		}

		t.Logf("Found %d import sessions", len(listResp.Sessions))
	})

	// Test filtering sessions by account
	t.Run("ListSessionsByAccount", func(t *testing.T) {
		listReq := &pb.ListImportSessionsRequest{
			Page:        1,
			PageSize:    10,
			AccountType: stringPtr("corporate"),
			AccountId:   stringPtr("test_session_mgmt_001"),
		}

		listResp, err := client.ListImportSessions(ctx, listReq)
		if err != nil {
			t.Fatalf("Failed to list sessions by account: %v", err)
		}

		// Should find exactly one session for this account
		if len(listResp.Sessions) != 1 {
			t.Errorf("Expected 1 session for account, got %d", len(listResp.Sessions))
		}
	})
}

// createValidCSVContent creates valid CSV content for testing
func createValidCSVContent() string {
	return `利用日,利用時刻,入口IC,出口IC,通行料金,車両番号,ETCカード番号
2025-09-21,10:30:00,東京IC,横浜IC,1200,品川 300 あ 1234,1234567890123456
2025-09-21,14:15:00,横浜IC,静岡IC,2500,横浜 500 か 5678,2345678901234567
2025-09-22,09:45:00,静岡IC,名古屋IC,3800,東京 700 さ 9012,3456789012345678`
}

// stringPtr returns a pointer to a string value
func stringPtr(s string) *string {
	return &s
}
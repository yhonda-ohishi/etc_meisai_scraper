//go:build integration

package integration

import (
	"context"
	"testing"
	"time"

	pb "github.com/yhonda-ohishi/etc_meisai/src/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

// TestETCRecordCRUDFlow tests the complete CRUD workflow for ETC records
// This integration test verifies:
// 1. Create a record
// 2. Read it back
// 3. Update it
// 4. List records
// 5. Delete the record
// 6. Verify deletion
func TestETCRecordCRUDFlow(t *testing.T) {
	// Setup gRPC client connection
	conn, client := setupGRPCClient(t)
	defer conn.Close()

	ctx := context.Background()

	// Test data for record creation
	originalRecord := &pb.ETCMeisaiRecord{
		Hash:          "abc123def456",
		Date:          "2025-09-21",
		Time:          "10:30:00",
		EntranceIc:    "東京IC",
		ExitIc:        "横浜IC",
		TollAmount:    1200,
		CarNumber:     "品川 300 あ 1234",
		EtcCardNumber: "1234567890123456",
	}

	// Step 1: Create a record
	t.Run("CreateRecord", func(t *testing.T) {
		createReq := &pb.CreateRecordRequest{
			Record: originalRecord,
		}

		createResp, err := client.CreateRecord(ctx, createReq)
		if err != nil {
			t.Fatalf("Failed to create record: %v", err)
		}

		if createResp.Record == nil {
			t.Fatal("Expected record in response, got nil")
		}

		if createResp.Record.Id <= 0 {
			t.Fatalf("Expected positive ID, got %d", createResp.Record.Id)
		}

		// Store the created record ID for subsequent tests
		originalRecord.Id = createResp.Record.Id

		// Verify the returned record matches what we sent
		assertRecordEquals(t, originalRecord, createResp.Record)
	})

	// Step 2: Read the record back
	t.Run("GetRecord", func(t *testing.T) {
		getReq := &pb.GetRecordRequest{
			Id: originalRecord.Id,
		}

		getResp, err := client.GetRecord(ctx, getReq)
		if err != nil {
			t.Fatalf("Failed to get record: %v", err)
		}

		if getResp.Record == nil {
			t.Fatal("Expected record in response, got nil")
		}

		// Verify the retrieved record matches the original
		assertRecordEquals(t, originalRecord, getResp.Record)
	})

	// Step 3: Update the record
	t.Run("UpdateRecord", func(t *testing.T) {
		// Modify some fields
		updatedRecord := &pb.ETCMeisaiRecord{
			Id:            originalRecord.Id,
			Hash:          originalRecord.Hash, // Hash should not change
			Date:          "2025-09-22",        // Changed date
			Time:          "15:45:00",          // Changed time
			EntranceIc:    "横浜IC",               // Changed entrance
			ExitIc:        "静岡IC",               // Changed exit
			TollAmount:    2500,                // Changed toll amount
			CarNumber:     originalRecord.CarNumber,     // Keep same
			EtcCardNumber: originalRecord.EtcCardNumber, // Keep same
		}

		updateReq := &pb.UpdateRecordRequest{
			Id:     originalRecord.Id,
			Record: updatedRecord,
		}

		updateResp, err := client.UpdateRecord(ctx, updateReq)
		if err != nil {
			t.Fatalf("Failed to update record: %v", err)
		}

		if updateResp.Record == nil {
			t.Fatal("Expected record in response, got nil")
		}

		// Verify the updated record matches our changes
		assertRecordEquals(t, updatedRecord, updateResp.Record)

		// Update original record for list test
		originalRecord = updatedRecord
	})

	// Step 4: List records (verify our record is in the list)
	t.Run("ListRecords", func(t *testing.T) {
		listReq := &pb.ListRecordsRequest{
			Page:     1,
			PageSize: 10,
			DateFrom: &originalRecord.Date, // Filter by our record's date
			DateTo:   &originalRecord.Date,
		}

		listResp, err := client.ListRecords(ctx, listReq)
		if err != nil {
			t.Fatalf("Failed to list records: %v", err)
		}

		if listResp.TotalCount == 0 {
			t.Fatal("Expected at least one record in list")
		}

		// Find our record in the list
		var foundRecord *pb.ETCMeisaiRecord
		for _, record := range listResp.Records {
			if record.Id == originalRecord.Id {
				foundRecord = record
				break
			}
		}

		if foundRecord == nil {
			t.Fatalf("Could not find our record (ID: %d) in the list", originalRecord.Id)
		}

		// Verify the found record matches our updated record
		assertRecordEquals(t, originalRecord, foundRecord)

		// Verify pagination metadata
		if listResp.Page != 1 {
			t.Errorf("Expected page 1, got %d", listResp.Page)
		}
		if listResp.PageSize != 10 {
			t.Errorf("Expected page size 10, got %d", listResp.PageSize)
		}
	})

	// Step 5: Test list with filters
	t.Run("ListRecordsWithFilters", func(t *testing.T) {
		// Test filtering by car number
		listReq := &pb.ListRecordsRequest{
			Page:      1,
			PageSize:  10,
			CarNumber: &originalRecord.CarNumber,
		}

		listResp, err := client.ListRecords(ctx, listReq)
		if err != nil {
			t.Fatalf("Failed to list records with car number filter: %v", err)
		}

		// Should find at least our record
		found := false
		for _, record := range listResp.Records {
			if record.Id == originalRecord.Id {
				found = true
				break
			}
		}

		if !found {
			t.Fatal("Could not find our record when filtering by car number")
		}

		// Test filtering by ETC card number
		listReq = &pb.ListRecordsRequest{
			Page:          1,
			PageSize:      10,
			EtcCardNumber: &originalRecord.EtcCardNumber,
		}

		listResp, err = client.ListRecords(ctx, listReq)
		if err != nil {
			t.Fatalf("Failed to list records with ETC card number filter: %v", err)
		}

		// Should find at least our record
		found = false
		for _, record := range listResp.Records {
			if record.Id == originalRecord.Id {
				found = true
				break
			}
		}

		if !found {
			t.Fatal("Could not find our record when filtering by ETC card number")
		}
	})

	// Step 6: Delete the record
	t.Run("DeleteRecord", func(t *testing.T) {
		deleteReq := &pb.DeleteRecordRequest{
			Id: originalRecord.Id,
		}

		_, err := client.DeleteRecord(ctx, deleteReq)
		if err != nil {
			t.Fatalf("Failed to delete record: %v", err)
		}
	})

	// Step 7: Verify deletion (should get NOT_FOUND error)
	t.Run("VerifyDeletion", func(t *testing.T) {
		getReq := &pb.GetRecordRequest{
			Id: originalRecord.Id,
		}

		_, err := client.GetRecord(ctx, getReq)
		if err == nil {
			t.Fatal("Expected error when getting deleted record, but got none")
		}

		// Verify it's a NOT_FOUND error
		st, ok := status.FromError(err)
		if !ok {
			t.Fatalf("Expected gRPC status error, got: %v", err)
		}

		if st.Code() != codes.NotFound {
			t.Fatalf("Expected NOT_FOUND error, got: %v", st.Code())
		}
	})
}

// TestETCRecordCRUDErrorHandling tests error scenarios in CRUD operations
func TestETCRecordCRUDErrorHandling(t *testing.T) {
	conn, client := setupGRPCClient(t)
	defer conn.Close()

	ctx := context.Background()

	t.Run("GetNonExistentRecord", func(t *testing.T) {
		getReq := &pb.GetRecordRequest{
			Id: 999999, // Non-existent ID
		}

		_, err := client.GetRecord(ctx, getReq)
		if err == nil {
			t.Fatal("Expected error when getting non-existent record")
		}

		st, ok := status.FromError(err)
		if !ok {
			t.Fatalf("Expected gRPC status error, got: %v", err)
		}

		if st.Code() != codes.NotFound {
			t.Fatalf("Expected NOT_FOUND error, got: %v", st.Code())
		}
	})

	t.Run("CreateRecordWithInvalidData", func(t *testing.T) {
		// Test with empty required fields
		invalidRecord := &pb.ETCMeisaiRecord{
			Hash: "", // Empty hash should be invalid
			Date: "invalid-date-format",
			Time: "25:99:99", // Invalid time
		}

		createReq := &pb.CreateRecordRequest{
			Record: invalidRecord,
		}

		_, err := client.CreateRecord(ctx, createReq)
		if err == nil {
			t.Fatal("Expected error when creating record with invalid data")
		}

		st, ok := status.FromError(err)
		if !ok {
			t.Fatalf("Expected gRPC status error, got: %v", err)
		}

		if st.Code() != codes.InvalidArgument {
			t.Fatalf("Expected INVALID_ARGUMENT error, got: %v", st.Code())
		}
	})

	t.Run("UpdateNonExistentRecord", func(t *testing.T) {
		updateReq := &pb.UpdateRecordRequest{
			Id: 999999, // Non-existent ID
			Record: &pb.ETCMeisaiRecord{
				Hash:          "test123",
				Date:          "2025-09-21",
				Time:          "10:30:00",
				EntranceIc:    "東京IC",
				ExitIc:        "横浜IC",
				TollAmount:    1200,
				CarNumber:     "品川 300 あ 1234",
				EtcCardNumber: "1234567890123456",
			},
		}

		_, err := client.UpdateRecord(ctx, updateReq)
		if err == nil {
			t.Fatal("Expected error when updating non-existent record")
		}

		st, ok := status.FromError(err)
		if !ok {
			t.Fatalf("Expected gRPC status error, got: %v", err)
		}

		if st.Code() != codes.NotFound {
			t.Fatalf("Expected NOT_FOUND error, got: %v", st.Code())
		}
	})

	t.Run("DeleteNonExistentRecord", func(t *testing.T) {
		deleteReq := &pb.DeleteRecordRequest{
			Id: 999999, // Non-existent ID
		}

		_, err := client.DeleteRecord(ctx, deleteReq)
		if err == nil {
			t.Fatal("Expected error when deleting non-existent record")
		}

		st, ok := status.FromError(err)
		if !ok {
			t.Fatalf("Expected gRPC status error, got: %v", err)
		}

		if st.Code() != codes.NotFound {
			t.Fatalf("Expected NOT_FOUND error, got: %v", st.Code())
		}
	})
}

// setupGRPCClient creates a gRPC client connection for testing
func setupGRPCClient(t *testing.T) (*grpc.ClientConn, pb.ETCMeisaiServiceClient) {
	// Connect to the gRPC server
	// In real tests, this would connect to a test server instance
	conn, err := grpc.NewClient(
		"localhost:50051",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithTimeout(10*time.Second),
	)
	if err != nil {
		t.Fatalf("Failed to connect to gRPC server: %v", err)
	}

	client := pb.NewETCMeisaiServiceClient(conn)
	return conn, client
}

// assertRecordEquals compares two ETCMeisaiRecord instances for equality
func assertRecordEquals(t *testing.T, expected, actual *pb.ETCMeisaiRecord) {
	t.Helper()

	if expected.Hash != actual.Hash {
		t.Errorf("Hash mismatch: expected %s, got %s", expected.Hash, actual.Hash)
	}
	if expected.Date != actual.Date {
		t.Errorf("Date mismatch: expected %s, got %s", expected.Date, actual.Date)
	}
	if expected.Time != actual.Time {
		t.Errorf("Time mismatch: expected %s, got %s", expected.Time, actual.Time)
	}
	if expected.EntranceIc != actual.EntranceIc {
		t.Errorf("EntranceIc mismatch: expected %s, got %s", expected.EntranceIc, actual.EntranceIc)
	}
	if expected.ExitIc != actual.ExitIc {
		t.Errorf("ExitIc mismatch: expected %s, got %s", expected.ExitIc, actual.ExitIc)
	}
	if expected.TollAmount != actual.TollAmount {
		t.Errorf("TollAmount mismatch: expected %d, got %d", expected.TollAmount, actual.TollAmount)
	}
	if expected.CarNumber != actual.CarNumber {
		t.Errorf("CarNumber mismatch: expected %s, got %s", expected.CarNumber, actual.CarNumber)
	}
	if expected.EtcCardNumber != actual.EtcCardNumber {
		t.Errorf("EtcCardNumber mismatch: expected %s, got %s", expected.EtcCardNumber, actual.EtcCardNumber)
	}
	if expected.Id != 0 && expected.Id != actual.Id {
		t.Errorf("ID mismatch: expected %d, got %d", expected.Id, actual.Id)
	}
}
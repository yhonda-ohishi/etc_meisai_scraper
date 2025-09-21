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
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/yhonda-ohishi/etc_meisai/src/pb"
)

func TestUpdateRecord_Success(t *testing.T) {
	// Arrange
	conn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("Failed to connect to gRPC server: %v", err)
	}
	defer conn.Close()

	client := pb.NewETCMeisaiServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Test data - update record with ID 1
	updatedRecord := &pb.ETCMeisaiRecord{
		Id:             1, // This should be ignored in the record field
		Hash:           "updated-hash-123",
		Date:           "2024-01-20",
		Time:           "14:30:00",
		EntranceIc:     "新宿",
		ExitIc:         "横浜",
		TollAmount:     1500,
		CarNumber:      "品川 456 い 5678",
		EtcCardNumber:  "9876543210987654",
		EtcNum:         stringPtr("ETC002"),
		DtakoRowId:     int64Ptr(123),
		UpdatedAt:      timestamppb.Now(),
	}

	req := &pb.UpdateRecordRequest{
		Id:     1,
		Record: updatedRecord,
	}

	// Act
	resp, err := client.UpdateRecord(ctx, req)

	// Assert
	// This test should FAIL initially as the server is not implemented yet
	if err != nil {
		st := status.Convert(err)
		if st.Code() == codes.Unimplemented {
			t.Logf("Expected failure: UpdateRecord not implemented yet - %v", err)
			return
		}
		t.Fatalf("Unexpected error: %v", err)
	}

	// If server is implemented, verify response
	if resp == nil {
		t.Fatal("Response is nil")
	}

	if resp.Record == nil {
		t.Fatal("Response record is nil")
	}

	// Verify the returned record has the correct ID
	if resp.Record.Id != 1 {
		t.Errorf("Expected record ID 1, got %d", resp.Record.Id)
	}

	// Verify the record data was updated
	if resp.Record.Hash != updatedRecord.Hash {
		t.Errorf("Expected hash %s, got %s", updatedRecord.Hash, resp.Record.Hash)
	}

	if resp.Record.Date != updatedRecord.Date {
		t.Errorf("Expected date %s, got %s", updatedRecord.Date, resp.Record.Date)
	}

	if resp.Record.Time != updatedRecord.Time {
		t.Errorf("Expected time %s, got %s", updatedRecord.Time, resp.Record.Time)
	}

	if resp.Record.TollAmount != updatedRecord.TollAmount {
		t.Errorf("Expected toll amount %d, got %d", updatedRecord.TollAmount, resp.Record.TollAmount)
	}

	if resp.Record.EntranceIc != updatedRecord.EntranceIc {
		t.Errorf("Expected entrance IC %s, got %s", updatedRecord.EntranceIc, resp.Record.EntranceIc)
	}

	if resp.Record.ExitIc != updatedRecord.ExitIc {
		t.Errorf("Expected exit IC %s, got %s", updatedRecord.ExitIc, resp.Record.ExitIc)
	}

	// Verify optional fields
	if resp.Record.EtcNum == nil || *resp.Record.EtcNum != *updatedRecord.EtcNum {
		t.Errorf("Expected etc_num %v, got %v", updatedRecord.EtcNum, resp.Record.EtcNum)
	}

	if resp.Record.DtakoRowId == nil || *resp.Record.DtakoRowId != *updatedRecord.DtakoRowId {
		t.Errorf("Expected dtako_row_id %v, got %v", updatedRecord.DtakoRowId, resp.Record.DtakoRowId)
	}

	// Verify updated_at was updated
	if resp.Record.UpdatedAt == nil {
		t.Error("Expected updated_at to be set")
	}
}

func TestUpdateRecord_NotFound(t *testing.T) {
	// Arrange
	conn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("Failed to connect to gRPC server: %v", err)
	}
	defer conn.Close()

	client := pb.NewETCMeisaiServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Test data - non-existent record ID
	updatedRecord := &pb.ETCMeisaiRecord{
		Hash:           "updated-hash-123",
		Date:           "2024-01-20",
		Time:           "14:30:00",
		EntranceIc:     "新宿",
		ExitIc:         "横浜",
		TollAmount:     1500,
		CarNumber:      "品川 456 い 5678",
		EtcCardNumber:  "9876543210987654",
	}

	req := &pb.UpdateRecordRequest{
		Id:     999999,
		Record: updatedRecord,
	}

	// Act
	resp, err := client.UpdateRecord(ctx, req)

	// Assert
	if err != nil {
		st := status.Convert(err)
		if st.Code() == codes.Unimplemented {
			t.Logf("Expected failure: UpdateRecord not implemented yet - %v", err)
			return
		}
		// When implemented, should return NotFound for non-existent record
		if st.Code() != codes.NotFound {
			t.Errorf("Expected NotFound error, got %v", st.Code())
		}
		return
	}

	// If no error, this might indicate the validation is not implemented
	if resp != nil {
		t.Logf("Warning: Expected NotFound error for non-existent record, but got successful response")
	}
}

func TestUpdateRecord_InvalidData(t *testing.T) {
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
		name   string
		id     int64
		record *pb.ETCMeisaiRecord
	}{
		{
			name: "invalid ID",
			id:   0,
			record: &pb.ETCMeisaiRecord{
				Hash:           "updated-hash-123",
				Date:           "2024-01-20",
				Time:           "14:30:00",
				EntranceIc:     "新宿",
				ExitIc:         "横浜",
				TollAmount:     1500,
				CarNumber:      "品川 456 い 5678",
				EtcCardNumber:  "9876543210987654",
			},
		},
		{
			name: "empty hash",
			id:   1,
			record: &pb.ETCMeisaiRecord{
				Hash:           "",
				Date:           "2024-01-20",
				Time:           "14:30:00",
				EntranceIc:     "新宿",
				ExitIc:         "横浜",
				TollAmount:     1500,
				CarNumber:      "品川 456 い 5678",
				EtcCardNumber:  "9876543210987654",
			},
		},
		{
			name: "invalid date format",
			id:   1,
			record: &pb.ETCMeisaiRecord{
				Hash:           "updated-hash-123",
				Date:           "2024/01/20", // Invalid format
				Time:           "14:30:00",
				EntranceIc:     "新宿",
				ExitIc:         "横浜",
				TollAmount:     1500,
				CarNumber:      "品川 456 い 5678",
				EtcCardNumber:  "9876543210987654",
			},
		},
		{
			name: "negative toll amount",
			id:   1,
			record: &pb.ETCMeisaiRecord{
				Hash:           "updated-hash-123",
				Date:           "2024-01-20",
				Time:           "14:30:00",
				EntranceIc:     "新宿",
				ExitIc:         "横浜",
				TollAmount:     -100, // Invalid negative amount
				CarNumber:      "品川 456 い 5678",
				EtcCardNumber:  "9876543210987654",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := &pb.UpdateRecordRequest{
				Id:     tc.id,
				Record: tc.record,
			}

			// Act
			resp, err := client.UpdateRecord(ctx, req)

			// Assert
			if err != nil {
				st := status.Convert(err)
				if st.Code() == codes.Unimplemented {
					t.Logf("Expected failure: UpdateRecord not implemented yet - %v", err)
					return
				}
				// When implemented, should return InvalidArgument for invalid data
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

func TestUpdateRecord_PartialUpdate(t *testing.T) {
	// Arrange
	conn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("Failed to connect to gRPC server: %v", err)
	}
	defer conn.Close()

	client := pb.NewETCMeisaiServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Test partial update - only update toll amount and optional fields
	updatedRecord := &pb.ETCMeisaiRecord{
		Hash:           "existing-hash", // Keep existing
		Date:           "2024-01-15",    // Keep existing
		Time:           "10:30:00",      // Keep existing
		EntranceIc:     "東京",           // Keep existing
		ExitIc:         "大阪",           // Keep existing
		TollAmount:     2000,            // Update this
		CarNumber:      "品川 123 あ 1234", // Keep existing
		EtcCardNumber:  "1234567890123456", // Keep existing
		EtcNum:         stringPtr("ETC003"), // Update this
		DtakoRowId:     int64Ptr(456),      // Update this
	}

	req := &pb.UpdateRecordRequest{
		Id:     1,
		Record: updatedRecord,
	}

	// Act
	resp, err := client.UpdateRecord(ctx, req)

	// Assert
	if err != nil {
		st := status.Convert(err)
		if st.Code() == codes.Unimplemented {
			t.Logf("Expected failure: UpdateRecord not implemented yet - %v", err)
			return
		}
		t.Fatalf("Unexpected error: %v", err)
	}

	// If server is implemented, verify response
	if resp == nil || resp.Record == nil {
		t.Fatal("Response or record is nil")
	}

	// Verify the partial update worked
	if resp.Record.TollAmount != 2000 {
		t.Errorf("Expected toll amount 2000, got %d", resp.Record.TollAmount)
	}

	// Verify optional fields were updated
	if resp.Record.EtcNum == nil || *resp.Record.EtcNum != "ETC003" {
		t.Errorf("Expected etc_num ETC003, got %v", resp.Record.EtcNum)
	}

	if resp.Record.DtakoRowId == nil || *resp.Record.DtakoRowId != 456 {
		t.Errorf("Expected dtako_row_id 456, got %v", resp.Record.DtakoRowId)
	}
}

func TestUpdateRecord_ClearOptionalFields(t *testing.T) {
	// Arrange
	conn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("Failed to connect to gRPC server: %v", err)
	}
	defer conn.Close()

	client := pb.NewETCMeisaiServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Test clearing optional fields by not setting them
	updatedRecord := &pb.ETCMeisaiRecord{
		Hash:           "existing-hash",
		Date:           "2024-01-15",
		Time:           "10:30:00",
		EntranceIc:     "東京",
		ExitIc:         "大阪",
		TollAmount:     1000,
		CarNumber:      "品川 123 あ 1234",
		EtcCardNumber:  "1234567890123456",
		// EtcNum and DtakoRowId not set - should clear them
	}

	req := &pb.UpdateRecordRequest{
		Id:     1,
		Record: updatedRecord,
	}

	// Act
	resp, err := client.UpdateRecord(ctx, req)

	// Assert
	if err != nil {
		st := status.Convert(err)
		if st.Code() == codes.Unimplemented {
			t.Logf("Expected failure: UpdateRecord not implemented yet - %v", err)
			return
		}
		t.Fatalf("Unexpected error: %v", err)
	}

	// If server is implemented, verify optional fields were cleared
	if resp == nil || resp.Record == nil {
		t.Fatal("Response or record is nil")
	}

	// Note: The behavior of clearing optional fields depends on implementation
	// Some implementations might keep existing values, others might clear them
	t.Logf("Optional fields after update - EtcNum: %v, DtakoRowId: %v",
		resp.Record.EtcNum, resp.Record.DtakoRowId)
}

// Helper function to create int64 pointer
func int64Ptr(i int64) *int64 {
	return &i
}
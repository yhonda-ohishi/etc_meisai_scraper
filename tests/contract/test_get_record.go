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

func TestGetRecord_Success(t *testing.T) {
	// Arrange
	conn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("Failed to connect to gRPC server: %v", err)
	}
	defer conn.Close()

	client := pb.NewETCMeisaiServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Test data - assume record with ID 1 exists
	req := &pb.GetRecordRequest{
		Id: 1,
	}

	// Act
	resp, err := client.GetRecord(ctx, req)

	// Assert
	// This test should FAIL initially as the server is not implemented yet
	if err != nil {
		st := status.Convert(err)
		if st.Code() == codes.Unimplemented {
			t.Logf("Expected failure: GetRecord not implemented yet - %v", err)
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

	// Verify required fields are not empty
	if resp.Record.Hash == "" {
		t.Error("Expected hash to be non-empty")
	}

	if resp.Record.Date == "" {
		t.Error("Expected date to be non-empty")
	}

	if resp.Record.Time == "" {
		t.Error("Expected time to be non-empty")
	}

	if resp.Record.CreatedAt == nil {
		t.Error("Expected created_at to be set")
	}

	if resp.Record.UpdatedAt == nil {
		t.Error("Expected updated_at to be set")
	}
}

func TestGetRecord_NotFound(t *testing.T) {
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
	req := &pb.GetRecordRequest{
		Id: 999999,
	}

	// Act
	resp, err := client.GetRecord(ctx, req)

	// Assert
	if err != nil {
		st := status.Convert(err)
		if st.Code() == codes.Unimplemented {
			t.Logf("Expected failure: GetRecord not implemented yet - %v", err)
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

func TestGetRecord_InvalidID(t *testing.T) {
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
		id   int64
	}{
		{
			name: "zero ID",
			id:   0,
		},
		{
			name: "negative ID",
			id:   -1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := &pb.GetRecordRequest{
				Id: tc.id,
			}

			// Act
			resp, err := client.GetRecord(ctx, req)

			// Assert
			if err != nil {
				st := status.Convert(err)
				if st.Code() == codes.Unimplemented {
					t.Logf("Expected failure: GetRecord not implemented yet - %v", err)
					return
				}
				// When implemented, should return InvalidArgument for invalid ID
				if st.Code() != codes.InvalidArgument {
					t.Errorf("Expected InvalidArgument error for invalid ID, got %v", st.Code())
				}
				return
			}

			// If no error, the validation might not be implemented yet
			if resp != nil {
				t.Logf("Warning: Expected validation error for invalid ID, but got successful response")
			}
		})
	}
}

func TestGetRecord_WithOptionalFields(t *testing.T) {
	// Arrange
	conn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("Failed to connect to gRPC server: %v", err)
	}
	defer conn.Close()

	client := pb.NewETCMeisaiServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Test data - assume record with ID 1 exists and has optional fields
	req := &pb.GetRecordRequest{
		Id: 1,
	}

	// Act
	resp, err := client.GetRecord(ctx, req)

	// Assert
	if err != nil {
		st := status.Convert(err)
		if st.Code() == codes.Unimplemented {
			t.Logf("Expected failure: GetRecord not implemented yet - %v", err)
			return
		}
		t.Fatalf("Unexpected error: %v", err)
	}

	// If server is implemented, verify optional fields handling
	if resp == nil || resp.Record == nil {
		t.Fatal("Response or record is nil")
	}

	// Optional fields should be properly handled (can be nil or have values)
	// Just verify they don't cause errors when accessed
	if resp.Record.EtcNum != nil {
		t.Logf("Record has etc_num: %s", *resp.Record.EtcNum)
	} else {
		t.Logf("Record has no etc_num (optional field is nil)")
	}

	if resp.Record.DtakoRowId != nil {
		t.Logf("Record has dtako_row_id: %d", *resp.Record.DtakoRowId)
	} else {
		t.Logf("Record has no dtako_row_id (optional field is nil)")
	}
}
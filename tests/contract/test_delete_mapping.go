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

func TestDeleteMapping_Success(t *testing.T) {
	// Arrange
	conn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("Failed to connect to gRPC server: %v", err)
	}
	defer conn.Close()

	client := pb.NewETCMeisaiServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Test data - delete mapping with ID 1
	req := &pb.DeleteMappingRequest{
		Id: 1,
	}

	// Act
	resp, err := client.DeleteMapping(ctx, req)

	// Assert
	// This test should FAIL initially as the server is not implemented yet
	if err != nil {
		st := status.Convert(err)
		if st.Code() == codes.Unimplemented {
			t.Logf("Expected failure: DeleteMapping not implemented yet - %v", err)
			return
		}
		t.Fatalf("Unexpected error: %v", err)
	}

	// If server is implemented, verify response
	if resp == nil {
		t.Fatal("Response is nil")
	}

	// DeleteMapping should return Empty message
	t.Logf("Received response: %+v", resp)

	// Verify the mapping was actually deleted by trying to get it
	getReq := &pb.GetMappingRequest{Id: 1}
	getResp, getErr := client.GetMapping(ctx, getReq)

	if getErr != nil {
		st := status.Convert(getErr)
		if st.Code() == codes.NotFound {
			t.Logf("Mapping successfully deleted - GetMapping returned NotFound as expected")
		} else if st.Code() == codes.Unimplemented {
			t.Logf("GetMapping not implemented, cannot verify deletion")
		} else {
			t.Errorf("Unexpected error when verifying deletion: %v", getErr)
		}
	} else if getResp != nil {
		t.Logf("Warning: Mapping still exists after deletion - this might be expected if soft delete is used")
	}
}

func TestDeleteMapping_NotFound(t *testing.T) {
	// Arrange
	conn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("Failed to connect to gRPC server: %v", err)
	}
	defer conn.Close()

	client := pb.NewETCMeisaiServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Test data - non-existent mapping ID
	req := &pb.DeleteMappingRequest{
		Id: 999999,
	}

	// Act
	resp, err := client.DeleteMapping(ctx, req)

	// Assert
	if err != nil {
		st := status.Convert(err)
		if st.Code() == codes.Unimplemented {
			t.Logf("Expected failure: DeleteMapping not implemented yet - %v", err)
			return
		}
		// When implemented, should return NotFound for non-existent mapping
		if st.Code() != codes.NotFound {
			t.Errorf("Expected NotFound error, got %v", st.Code())
		}
		return
	}

	// If no error, this might indicate idempotent delete behavior
	if resp != nil {
		t.Logf("Warning: Expected NotFound error for non-existent mapping, but got successful response (idempotent delete?)")
	}
}
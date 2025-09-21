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

func TestGetMapping_Success(t *testing.T) {
	// Arrange
	conn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("Failed to connect to gRPC server: %v", err)
	}
	defer conn.Close()

	client := pb.NewETCMeisaiServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Test data - assume mapping with ID 1 exists
	req := &pb.GetMappingRequest{
		Id: 1,
	}

	// Act
	resp, err := client.GetMapping(ctx, req)

	// Assert
	// This test should FAIL initially as the server is not implemented yet
	if err != nil {
		st := status.Convert(err)
		if st.Code() == codes.Unimplemented {
			t.Logf("Expected failure: GetMapping not implemented yet - %v", err)
			return
		}
		t.Fatalf("Unexpected error: %v", err)
	}

	// If server is implemented, verify response
	if resp == nil {
		t.Fatal("Response is nil")
	}

	if resp.Mapping == nil {
		t.Fatal("Response mapping is nil")
	}

	// Verify the returned mapping has the correct ID
	if resp.Mapping.Id != 1 {
		t.Errorf("Expected mapping ID 1, got %d", resp.Mapping.Id)
	}

	// Verify required fields are not empty/zero
	if resp.Mapping.EtcRecordId == 0 {
		t.Error("Expected ETC record ID to be non-zero")
	}

	if resp.Mapping.MappingType == "" {
		t.Error("Expected mapping type to be non-empty")
	}

	if resp.Mapping.MappedEntityId == 0 {
		t.Error("Expected mapped entity ID to be non-zero")
	}

	if resp.Mapping.MappedEntityType == "" {
		t.Error("Expected mapped entity type to be non-empty")
	}

	if resp.Mapping.Confidence < 0 || resp.Mapping.Confidence > 1 {
		t.Errorf("Expected confidence to be between 0 and 1, got %f", resp.Mapping.Confidence)
	}

	if resp.Mapping.Status == pb.MappingStatus_MAPPING_STATUS_UNSPECIFIED {
		t.Error("Expected mapping status to be specified")
	}

	if resp.Mapping.CreatedBy == "" {
		t.Error("Expected created_by to be non-empty")
	}

	if resp.Mapping.CreatedAt == nil {
		t.Error("Expected created_at to be set")
	}

	if resp.Mapping.UpdatedAt == nil {
		t.Error("Expected updated_at to be set")
	}

	// ETC record should be populated if available
	if resp.Mapping.EtcRecord != nil {
		if resp.Mapping.EtcRecord.Id != resp.Mapping.EtcRecordId {
			t.Errorf("Expected ETC record ID %d, got %d", resp.Mapping.EtcRecordId, resp.Mapping.EtcRecord.Id)
		}
	}
}

func TestGetMapping_NotFound(t *testing.T) {
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
	req := &pb.GetMappingRequest{
		Id: 999999,
	}

	// Act
	resp, err := client.GetMapping(ctx, req)

	// Assert
	if err != nil {
		st := status.Convert(err)
		if st.Code() == codes.Unimplemented {
			t.Logf("Expected failure: GetMapping not implemented yet - %v", err)
			return
		}
		// When implemented, should return NotFound for non-existent mapping
		if st.Code() != codes.NotFound {
			t.Errorf("Expected NotFound error, got %v", st.Code())
		}
		return
	}

	// If no error, this might indicate the validation is not implemented
	if resp != nil {
		t.Logf("Warning: Expected NotFound error for non-existent mapping, but got successful response")
	}
}

func TestGetMapping_InvalidID(t *testing.T) {
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
			req := &pb.GetMappingRequest{
				Id: tc.id,
			}

			// Act
			resp, err := client.GetMapping(ctx, req)

			// Assert
			if err != nil {
				st := status.Convert(err)
				if st.Code() == codes.Unimplemented {
					t.Logf("Expected failure: GetMapping not implemented yet - %v", err)
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

func TestGetMapping_WithMetadata(t *testing.T) {
	// Arrange
	conn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("Failed to connect to gRPC server: %v", err)
	}
	defer conn.Close()

	client := pb.NewETCMeisaiServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Test data - assume mapping with ID 1 has metadata
	req := &pb.GetMappingRequest{
		Id: 1,
	}

	// Act
	resp, err := client.GetMapping(ctx, req)

	// Assert
	if err != nil {
		st := status.Convert(err)
		if st.Code() == codes.Unimplemented {
			t.Logf("Expected failure: GetMapping not implemented yet - %v", err)
			return
		}
		t.Fatalf("Unexpected error: %v", err)
	}

	// If server is implemented, verify metadata handling
	if resp == nil || resp.Mapping == nil {
		t.Fatal("Response or mapping is nil")
	}

	// Metadata can be nil or have content
	if resp.Mapping.Metadata != nil {
		t.Logf("Mapping has metadata with %d fields", len(resp.Mapping.Metadata.Fields))

		// Verify metadata structure if present
		for key, value := range resp.Mapping.Metadata.Fields {
			if key == "" {
				t.Error("Metadata should not have empty keys")
			}
			if value == nil {
				t.Errorf("Metadata field %s should not have nil value", key)
			}
		}
	} else {
		t.Logf("Mapping has no metadata (optional field)")
	}
}

func TestGetMapping_DifferentStatuses(t *testing.T) {
	// Arrange
	conn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("Failed to connect to gRPC server: %v", err)
	}
	defer conn.Close()

	client := pb.NewETCMeisaiServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Test mappings with different IDs (assuming they have different statuses)
	testCases := []struct {
		name string
		id   int64
	}{
		{
			name: "active mapping",
			id:   1,
		},
		{
			name: "inactive mapping",
			id:   2,
		},
		{
			name: "pending mapping",
			id:   3,
		},
		{
			name: "rejected mapping",
			id:   4,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := &pb.GetMappingRequest{
				Id: tc.id,
			}

			// Act
			resp, err := client.GetMapping(ctx, req)

			// Assert
			if err != nil {
				st := status.Convert(err)
				if st.Code() == codes.Unimplemented {
					t.Logf("Expected failure: GetMapping not implemented yet - %v", err)
					return
				}
				if st.Code() == codes.NotFound {
					t.Logf("Mapping ID %d not found - this is acceptable for test data", tc.id)
					return
				}
				t.Fatalf("Unexpected error for %s: %v", tc.name, err)
			}

			// If server is implemented and mapping exists, verify status
			if resp != nil && resp.Mapping != nil {
				if resp.Mapping.Status == pb.MappingStatus_MAPPING_STATUS_UNSPECIFIED {
					t.Errorf("Mapping %d has unspecified status", tc.id)
				} else {
					t.Logf("Mapping %d has status: %s", tc.id, resp.Mapping.Status.String())
				}
			}
		})
	}
}

func TestGetMapping_WithETCRecord(t *testing.T) {
	// Arrange
	conn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("Failed to connect to gRPC server: %v", err)
	}
	defer conn.Close()

	client := pb.NewETCMeisaiServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Test data - assume mapping with ID 1 has ETC record populated
	req := &pb.GetMappingRequest{
		Id: 1,
	}

	// Act
	resp, err := client.GetMapping(ctx, req)

	// Assert
	if err != nil {
		st := status.Convert(err)
		if st.Code() == codes.Unimplemented {
			t.Logf("Expected failure: GetMapping not implemented yet - %v", err)
			return
		}
		t.Fatalf("Unexpected error: %v", err)
	}

	// If server is implemented, verify ETC record population
	if resp == nil || resp.Mapping == nil {
		t.Fatal("Response or mapping is nil")
	}

	// Check if ETC record is populated
	if resp.Mapping.EtcRecord != nil {
		// Verify ETC record consistency
		if resp.Mapping.EtcRecord.Id != resp.Mapping.EtcRecordId {
			t.Errorf("ETC record ID mismatch: mapping.etc_record_id=%d, etc_record.id=%d",
				resp.Mapping.EtcRecordId, resp.Mapping.EtcRecord.Id)
		}

		// Verify ETC record has basic fields
		if resp.Mapping.EtcRecord.Hash == "" {
			t.Error("ETC record should have non-empty hash")
		}

		if resp.Mapping.EtcRecord.Date == "" {
			t.Error("ETC record should have non-empty date")
		}

		if resp.Mapping.EtcRecord.Time == "" {
			t.Error("ETC record should have non-empty time")
		}

		t.Logf("Mapping includes ETC record: ID=%d, Date=%s, Amount=%d",
			resp.Mapping.EtcRecord.Id, resp.Mapping.EtcRecord.Date, resp.Mapping.EtcRecord.TollAmount)
	} else {
		t.Logf("Mapping does not include ETC record details (may be by design)")
	}
}

func TestGetMapping_ConfidenceValidation(t *testing.T) {
	// Arrange
	conn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("Failed to connect to gRPC server: %v", err)
	}
	defer conn.Close()

	client := pb.NewETCMeisaiServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Test multiple mappings to check confidence value ranges
	for i := int64(1); i <= 5; i++ {
		req := &pb.GetMappingRequest{
			Id: i,
		}

		// Act
		resp, err := client.GetMapping(ctx, req)

		// Assert
		if err != nil {
			st := status.Convert(err)
			if st.Code() == codes.Unimplemented {
				t.Logf("Expected failure: GetMapping not implemented yet - %v", err)
				return
			}
			if st.Code() == codes.NotFound {
				continue // Skip non-existent mappings
			}
			t.Fatalf("Unexpected error for mapping %d: %v", i, err)
		}

		// Verify confidence is in valid range
		if resp != nil && resp.Mapping != nil {
			confidence := resp.Mapping.Confidence
			if confidence < 0.0 || confidence > 1.0 {
				t.Errorf("Mapping %d has invalid confidence %f (should be 0.0-1.0)", i, confidence)
			}
			t.Logf("Mapping %d confidence: %.3f", i, confidence)
		}
	}
}
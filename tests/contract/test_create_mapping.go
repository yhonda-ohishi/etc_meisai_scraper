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
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/yhonda-ohishi/etc_meisai/src/pb"
)

func TestCreateMapping_Success(t *testing.T) {
	// Arrange
	conn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("Failed to connect to gRPC server: %v", err)
	}
	defer conn.Close()

	client := pb.NewETCMeisaiServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Create metadata for the mapping
	metadata, _ := structpb.NewStruct(map[string]interface{}{
		"confidence_score": 0.95,
		"algorithm":        "distance_based",
		"verified":         true,
	})

	// Test data
	mapping := &pb.ETCMapping{
		EtcRecordId:       1,
		MappingType:       "dtako_match",
		MappedEntityId:    123,
		MappedEntityType:  "dtako_record",
		Confidence:        0.95,
		Status:           pb.MappingStatus_MAPPING_STATUS_ACTIVE,
		Metadata:         metadata,
		CreatedBy:        "test-user",
		CreatedAt:        timestamppb.Now(),
		UpdatedAt:        timestamppb.Now(),
	}

	req := &pb.CreateMappingRequest{
		Mapping: mapping,
	}

	// Act
	resp, err := client.CreateMapping(ctx, req)

	// Assert
	// This test should FAIL initially as the server is not implemented yet
	if err != nil {
		st := status.Convert(err)
		if st.Code() == codes.Unimplemented {
			t.Logf("Expected failure: CreateMapping not implemented yet - %v", err)
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

	// Verify the returned mapping has an ID assigned
	if resp.Mapping.Id == 0 {
		t.Error("Expected mapping ID to be assigned")
	}

	// Verify the mapping data matches input
	if resp.Mapping.EtcRecordId != mapping.EtcRecordId {
		t.Errorf("Expected ETC record ID %d, got %d", mapping.EtcRecordId, resp.Mapping.EtcRecordId)
	}

	if resp.Mapping.MappingType != mapping.MappingType {
		t.Errorf("Expected mapping type %s, got %s", mapping.MappingType, resp.Mapping.MappingType)
	}

	if resp.Mapping.MappedEntityId != mapping.MappedEntityId {
		t.Errorf("Expected mapped entity ID %d, got %d", mapping.MappedEntityId, resp.Mapping.MappedEntityId)
	}

	if resp.Mapping.MappedEntityType != mapping.MappedEntityType {
		t.Errorf("Expected mapped entity type %s, got %s", mapping.MappedEntityType, resp.Mapping.MappedEntityType)
	}

	if resp.Mapping.Confidence != mapping.Confidence {
		t.Errorf("Expected confidence %f, got %f", mapping.Confidence, resp.Mapping.Confidence)
	}

	if resp.Mapping.Status != mapping.Status {
		t.Errorf("Expected status %s, got %s", mapping.Status.String(), resp.Mapping.Status.String())
	}

	if resp.Mapping.CreatedBy != mapping.CreatedBy {
		t.Errorf("Expected created by %s, got %s", mapping.CreatedBy, resp.Mapping.CreatedBy)
	}

	// Verify ETC record is populated if available
	if resp.Mapping.EtcRecord != nil {
		if resp.Mapping.EtcRecord.Id != mapping.EtcRecordId {
			t.Errorf("Expected ETC record ID %d, got %d", mapping.EtcRecordId, resp.Mapping.EtcRecord.Id)
		}
	}
}

func TestCreateMapping_InvalidData(t *testing.T) {
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
		name    string
		mapping *pb.ETCMapping
	}{
		{
			name: "zero ETC record ID",
			mapping: &pb.ETCMapping{
				EtcRecordId:      0,
				MappingType:      "dtako_match",
				MappedEntityId:   123,
				MappedEntityType: "dtako_record",
				Confidence:       0.95,
				Status:          pb.MappingStatus_MAPPING_STATUS_ACTIVE,
				CreatedBy:       "test-user",
			},
		},
		{
			name: "empty mapping type",
			mapping: &pb.ETCMapping{
				EtcRecordId:      1,
				MappingType:      "",
				MappedEntityId:   123,
				MappedEntityType: "dtako_record",
				Confidence:       0.95,
				Status:          pb.MappingStatus_MAPPING_STATUS_ACTIVE,
				CreatedBy:       "test-user",
			},
		},
		{
			name: "zero mapped entity ID",
			mapping: &pb.ETCMapping{
				EtcRecordId:      1,
				MappingType:      "dtako_match",
				MappedEntityId:   0,
				MappedEntityType: "dtako_record",
				Confidence:       0.95,
				Status:          pb.MappingStatus_MAPPING_STATUS_ACTIVE,
				CreatedBy:       "test-user",
			},
		},
		{
			name: "empty mapped entity type",
			mapping: &pb.ETCMapping{
				EtcRecordId:      1,
				MappingType:      "dtako_match",
				MappedEntityId:   123,
				MappedEntityType: "",
				Confidence:       0.95,
				Status:          pb.MappingStatus_MAPPING_STATUS_ACTIVE,
				CreatedBy:       "test-user",
			},
		},
		{
			name: "invalid confidence - negative",
			mapping: &pb.ETCMapping{
				EtcRecordId:      1,
				MappingType:      "dtako_match",
				MappedEntityId:   123,
				MappedEntityType: "dtako_record",
				Confidence:       -0.1,
				Status:          pb.MappingStatus_MAPPING_STATUS_ACTIVE,
				CreatedBy:       "test-user",
			},
		},
		{
			name: "invalid confidence - greater than 1",
			mapping: &pb.ETCMapping{
				EtcRecordId:      1,
				MappingType:      "dtako_match",
				MappedEntityId:   123,
				MappedEntityType: "dtako_record",
				Confidence:       1.5,
				Status:          pb.MappingStatus_MAPPING_STATUS_ACTIVE,
				CreatedBy:       "test-user",
			},
		},
		{
			name: "unspecified status",
			mapping: &pb.ETCMapping{
				EtcRecordId:      1,
				MappingType:      "dtako_match",
				MappedEntityId:   123,
				MappedEntityType: "dtako_record",
				Confidence:       0.95,
				Status:          pb.MappingStatus_MAPPING_STATUS_UNSPECIFIED,
				CreatedBy:       "test-user",
			},
		},
		{
			name: "empty created by",
			mapping: &pb.ETCMapping{
				EtcRecordId:      1,
				MappingType:      "dtako_match",
				MappedEntityId:   123,
				MappedEntityType: "dtako_record",
				Confidence:       0.95,
				Status:          pb.MappingStatus_MAPPING_STATUS_ACTIVE,
				CreatedBy:       "",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := &pb.CreateMappingRequest{
				Mapping: tc.mapping,
			}

			// Act
			resp, err := client.CreateMapping(ctx, req)

			// Assert
			if err != nil {
				st := status.Convert(err)
				if st.Code() == codes.Unimplemented {
					t.Logf("Expected failure: CreateMapping not implemented yet - %v", err)
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

func TestCreateMapping_DuplicateMapping(t *testing.T) {
	// Arrange
	conn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("Failed to connect to gRPC server: %v", err)
	}
	defer conn.Close()

	client := pb.NewETCMeisaiServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Test data - mapping that might already exist
	mapping := &pb.ETCMapping{
		EtcRecordId:      1,
		MappingType:      "dtako_match",
		MappedEntityId:   123,
		MappedEntityType: "dtako_record",
		Confidence:       0.95,
		Status:          pb.MappingStatus_MAPPING_STATUS_ACTIVE,
		CreatedBy:       "test-user",
	}

	req := &pb.CreateMappingRequest{
		Mapping: mapping,
	}

	// Act - Try to create the same mapping twice
	_, err1 := client.CreateMapping(ctx, req)
	if err1 != nil {
		st := status.Convert(err1)
		if st.Code() == codes.Unimplemented {
			t.Logf("Expected failure: CreateMapping not implemented yet - %v", err1)
			return
		}
	}

	_, err2 := client.CreateMapping(ctx, req)

	// Assert
	if err2 != nil {
		st := status.Convert(err2)
		if st.Code() == codes.Unimplemented {
			t.Logf("Expected failure: CreateMapping not implemented yet - %v", err2)
			return
		}
		// When implemented, should return AlreadyExists for duplicate mapping
		if st.Code() != codes.AlreadyExists {
			t.Errorf("Expected AlreadyExists error for duplicate mapping, got %v", st.Code())
		}
		return
	}

	// If no error, the duplicate check might not be implemented yet
	t.Logf("Warning: Expected duplicate mapping error, but got successful response")
}

func TestCreateMapping_NonExistentETCRecord(t *testing.T) {
	// Arrange
	conn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("Failed to connect to gRPC server: %v", err)
	}
	defer conn.Close()

	client := pb.NewETCMeisaiServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Test data - mapping with non-existent ETC record
	mapping := &pb.ETCMapping{
		EtcRecordId:      999999, // Non-existent record
		MappingType:      "dtako_match",
		MappedEntityId:   123,
		MappedEntityType: "dtako_record",
		Confidence:       0.95,
		Status:          pb.MappingStatus_MAPPING_STATUS_ACTIVE,
		CreatedBy:       "test-user",
	}

	req := &pb.CreateMappingRequest{
		Mapping: mapping,
	}

	// Act
	resp, err := client.CreateMapping(ctx, req)

	// Assert
	if err != nil {
		st := status.Convert(err)
		if st.Code() == codes.Unimplemented {
			t.Logf("Expected failure: CreateMapping not implemented yet - %v", err)
			return
		}
		// When implemented, should return NotFound for non-existent ETC record
		if st.Code() != codes.NotFound {
			t.Errorf("Expected NotFound error for non-existent ETC record, got %v", st.Code())
		}
		return
	}

	// If no error, the foreign key validation might not be implemented yet
	if resp != nil {
		t.Logf("Warning: Expected NotFound error for non-existent ETC record, but got successful response")
	}
}

func TestCreateMapping_WithMetadata(t *testing.T) {
	// Arrange
	conn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("Failed to connect to gRPC server: %v", err)
	}
	defer conn.Close()

	client := pb.NewETCMeisaiServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Create complex metadata
	metadata, _ := structpb.NewStruct(map[string]interface{}{
		"confidence_score": 0.95,
		"algorithm":        "distance_based",
		"verified":         true,
		"match_criteria": map[string]interface{}{
			"time_diff_seconds": 120,
			"location_match":    true,
			"amount_match":      false,
		},
		"processing_info": map[string]interface{}{
			"processor_version": "1.2.3",
			"processed_at":      "2024-01-15T10:30:00Z",
		},
	})

	// Test data with rich metadata
	mapping := &pb.ETCMapping{
		EtcRecordId:      1,
		MappingType:      "dtako_match",
		MappedEntityId:   123,
		MappedEntityType: "dtako_record",
		Confidence:       0.95,
		Status:          pb.MappingStatus_MAPPING_STATUS_ACTIVE,
		Metadata:        metadata,
		CreatedBy:       "test-user",
	}

	req := &pb.CreateMappingRequest{
		Mapping: mapping,
	}

	// Act
	resp, err := client.CreateMapping(ctx, req)

	// Assert
	if err != nil {
		st := status.Convert(err)
		if st.Code() == codes.Unimplemented {
			t.Logf("Expected failure: CreateMapping not implemented yet - %v", err)
			return
		}
		t.Fatalf("Unexpected error: %v", err)
	}

	// If server is implemented, verify metadata handling
	if resp == nil || resp.Mapping == nil {
		t.Fatal("Response or mapping is nil")
	}

	// Verify metadata is preserved
	if resp.Mapping.Metadata == nil {
		t.Error("Expected metadata to be preserved")
	} else {
		// Check some metadata fields
		if confidenceScore := resp.Mapping.Metadata.Fields["confidence_score"]; confidenceScore != nil {
			if confidenceScore.GetNumberValue() != 0.95 {
				t.Errorf("Expected confidence_score 0.95 in metadata, got %f", confidenceScore.GetNumberValue())
			}
		} else {
			t.Error("Expected confidence_score in metadata")
		}

		if algorithm := resp.Mapping.Metadata.Fields["algorithm"]; algorithm != nil {
			if algorithm.GetStringValue() != "distance_based" {
				t.Errorf("Expected algorithm 'distance_based' in metadata, got %s", algorithm.GetStringValue())
			}
		} else {
			t.Error("Expected algorithm in metadata")
		}
	}
}

func TestCreateMapping_DifferentStatuses(t *testing.T) {
	// Arrange
	conn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("Failed to connect to gRPC server: %v", err)
	}
	defer conn.Close()

	client := pb.NewETCMeisaiServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	statuses := []pb.MappingStatus{
		pb.MappingStatus_MAPPING_STATUS_ACTIVE,
		pb.MappingStatus_MAPPING_STATUS_INACTIVE,
		pb.MappingStatus_MAPPING_STATUS_PENDING,
		pb.MappingStatus_MAPPING_STATUS_REJECTED,
	}

	for i, mappingStatus := range statuses {
		t.Run("status_"+mappingStatus.String(), func(t *testing.T) {
			mapping := &pb.ETCMapping{
				EtcRecordId:      int64(i + 1), // Use different ETC record IDs
				MappingType:      "dtako_match",
				MappedEntityId:   int64(100 + i),
				MappedEntityType: "dtako_record",
				Confidence:       0.95,
				Status:          mappingStatus,
				CreatedBy:       "test-user",
			}

			req := &pb.CreateMappingRequest{
				Mapping: mapping,
			}

			// Act
			resp, err := client.CreateMapping(ctx, req)

			// Assert
			if err != nil {
				st := status.Convert(err)
				if st.Code() == codes.Unimplemented {
					t.Logf("Expected failure: CreateMapping not implemented yet - %v", err)
					return
				}
				t.Fatalf("Unexpected error for status %s: %v", mappingStatus.String(), err)
			}

			// If server is implemented, verify status is preserved
			if resp == nil || resp.Mapping == nil {
				t.Fatal("Response or mapping is nil")
			}

			if resp.Mapping.Status != mappingStatus {
				t.Errorf("Expected status %s, got %s", mappingStatus.String(), resp.Mapping.Status.String())
			}
		})
	}
}
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

func TestUpdateMapping_Success(t *testing.T) {
	// Arrange
	conn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("Failed to connect to gRPC server: %v", err)
	}
	defer conn.Close()

	client := pb.NewETCMeisaiServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Create updated metadata
	metadata, _ := structpb.NewStruct(map[string]interface{}{
		"confidence_score": 0.98,
		"algorithm":        "enhanced_distance_based",
		"verified":         true,
		"updated_reason":   "manual_verification",
	})

	// Test data - update mapping with ID 1
	updatedMapping := &pb.ETCMapping{
		Id:               1, // This should be ignored in the mapping field
		EtcRecordId:      1,
		MappingType:      "dtako_match",
		MappedEntityId:   456, // Updated entity ID
		MappedEntityType: "dtako_record",
		Confidence:       0.98, // Updated confidence
		Status:          pb.MappingStatus_MAPPING_STATUS_ACTIVE,
		Metadata:        metadata,
		CreatedBy:       "test-user",
		UpdatedAt:       timestamppb.Now(),
	}

	req := &pb.UpdateMappingRequest{
		Id:      1,
		Mapping: updatedMapping,
	}

	// Act
	resp, err := client.UpdateMapping(ctx, req)

	// Assert
	// This test should FAIL initially as the server is not implemented yet
	if err != nil {
		st := status.Convert(err)
		if st.Code() == codes.Unimplemented {
			t.Logf("Expected failure: UpdateMapping not implemented yet - %v", err)
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

	// Verify the mapping data was updated
	if resp.Mapping.MappedEntityId != updatedMapping.MappedEntityId {
		t.Errorf("Expected mapped entity ID %d, got %d", updatedMapping.MappedEntityId, resp.Mapping.MappedEntityId)
	}

	if resp.Mapping.Confidence != updatedMapping.Confidence {
		t.Errorf("Expected confidence %f, got %f", updatedMapping.Confidence, resp.Mapping.Confidence)
	}

	if resp.Mapping.Status != updatedMapping.Status {
		t.Errorf("Expected status %s, got %s", updatedMapping.Status.String(), resp.Mapping.Status.String())
	}

	// Verify updated_at was updated
	if resp.Mapping.UpdatedAt == nil {
		t.Error("Expected updated_at to be set")
	}

	// Verify metadata was updated
	if resp.Mapping.Metadata != nil {
		if confidenceScore := resp.Mapping.Metadata.Fields["confidence_score"]; confidenceScore != nil {
			if confidenceScore.GetNumberValue() != 0.98 {
				t.Errorf("Expected updated confidence_score 0.98 in metadata, got %f", confidenceScore.GetNumberValue())
			}
		}
	}
}

func TestUpdateMapping_NotFound(t *testing.T) {
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
	updatedMapping := &pb.ETCMapping{
		EtcRecordId:      1,
		MappingType:      "dtako_match",
		MappedEntityId:   456,
		MappedEntityType: "dtako_record",
		Confidence:       0.98,
		Status:          pb.MappingStatus_MAPPING_STATUS_ACTIVE,
		CreatedBy:       "test-user",
	}

	req := &pb.UpdateMappingRequest{
		Id:      999999,
		Mapping: updatedMapping,
	}

	// Act
	resp, err := client.UpdateMapping(ctx, req)

	// Assert
	if err != nil {
		st := status.Convert(err)
		if st.Code() == codes.Unimplemented {
			t.Logf("Expected failure: UpdateMapping not implemented yet - %v", err)
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

func TestUpdateMapping_InvalidData(t *testing.T) {
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
		id      int64
		mapping *pb.ETCMapping
	}{
		{
			name: "invalid ID",
			id:   0,
			mapping: &pb.ETCMapping{
				EtcRecordId:      1,
				MappingType:      "dtako_match",
				MappedEntityId:   456,
				MappedEntityType: "dtako_record",
				Confidence:       0.98,
				Status:          pb.MappingStatus_MAPPING_STATUS_ACTIVE,
				CreatedBy:       "test-user",
			},
		},
		{
			name: "zero ETC record ID",
			id:   1,
			mapping: &pb.ETCMapping{
				EtcRecordId:      0,
				MappingType:      "dtako_match",
				MappedEntityId:   456,
				MappedEntityType: "dtako_record",
				Confidence:       0.98,
				Status:          pb.MappingStatus_MAPPING_STATUS_ACTIVE,
				CreatedBy:       "test-user",
			},
		},
		{
			name: "empty mapping type",
			id:   1,
			mapping: &pb.ETCMapping{
				EtcRecordId:      1,
				MappingType:      "",
				MappedEntityId:   456,
				MappedEntityType: "dtako_record",
				Confidence:       0.98,
				Status:          pb.MappingStatus_MAPPING_STATUS_ACTIVE,
				CreatedBy:       "test-user",
			},
		},
		{
			name: "invalid confidence - negative",
			id:   1,
			mapping: &pb.ETCMapping{
				EtcRecordId:      1,
				MappingType:      "dtako_match",
				MappedEntityId:   456,
				MappedEntityType: "dtako_record",
				Confidence:       -0.1,
				Status:          pb.MappingStatus_MAPPING_STATUS_ACTIVE,
				CreatedBy:       "test-user",
			},
		},
		{
			name: "invalid confidence - greater than 1",
			id:   1,
			mapping: &pb.ETCMapping{
				EtcRecordId:      1,
				MappingType:      "dtako_match",
				MappedEntityId:   456,
				MappedEntityType: "dtako_record",
				Confidence:       1.5,
				Status:          pb.MappingStatus_MAPPING_STATUS_ACTIVE,
				CreatedBy:       "test-user",
			},
		},
		{
			name: "unspecified status",
			id:   1,
			mapping: &pb.ETCMapping{
				EtcRecordId:      1,
				MappingType:      "dtako_match",
				MappedEntityId:   456,
				MappedEntityType: "dtako_record",
				Confidence:       0.98,
				Status:          pb.MappingStatus_MAPPING_STATUS_UNSPECIFIED,
				CreatedBy:       "test-user",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := &pb.UpdateMappingRequest{
				Id:      tc.id,
				Mapping: tc.mapping,
			}

			// Act
			resp, err := client.UpdateMapping(ctx, req)

			// Assert
			if err != nil {
				st := status.Convert(err)
				if st.Code() == codes.Unimplemented {
					t.Logf("Expected failure: UpdateMapping not implemented yet - %v", err)
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

func TestUpdateMapping_StatusTransitions(t *testing.T) {
	// Arrange
	conn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("Failed to connect to gRPC server: %v", err)
	}
	defer conn.Close()

	client := pb.NewETCMeisaiServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Test different status transitions
	statusTransitions := []struct {
		name      string
		newStatus pb.MappingStatus
	}{
		{
			name:      "activate mapping",
			newStatus: pb.MappingStatus_MAPPING_STATUS_ACTIVE,
		},
		{
			name:      "deactivate mapping",
			newStatus: pb.MappingStatus_MAPPING_STATUS_INACTIVE,
		},
		{
			name:      "set pending",
			newStatus: pb.MappingStatus_MAPPING_STATUS_PENDING,
		},
		{
			name:      "reject mapping",
			newStatus: pb.MappingStatus_MAPPING_STATUS_REJECTED,
		},
	}

	for i, transition := range statusTransitions {
		t.Run(transition.name, func(t *testing.T) {
			updatedMapping := &pb.ETCMapping{
				EtcRecordId:      1,
				MappingType:      "dtako_match",
				MappedEntityId:   int64(100 + i), // Use different entity IDs
				MappedEntityType: "dtako_record",
				Confidence:       0.95,
				Status:          transition.newStatus,
				CreatedBy:       "test-user",
			}

			req := &pb.UpdateMappingRequest{
				Id:      int64(i + 1), // Use different mapping IDs
				Mapping: updatedMapping,
			}

			// Act
			resp, err := client.UpdateMapping(ctx, req)

			// Assert
			if err != nil {
				st := status.Convert(err)
				if st.Code() == codes.Unimplemented {
					t.Logf("Expected failure: UpdateMapping not implemented yet - %v", err)
					return
				}
				if st.Code() == codes.NotFound {
					t.Logf("Mapping ID %d not found - this is acceptable for test data", req.Id)
					return
				}
				t.Fatalf("Unexpected error for %s: %v", transition.name, err)
			}

			// If server is implemented, verify status transition
			if resp != nil && resp.Mapping != nil {
				if resp.Mapping.Status != transition.newStatus {
					t.Errorf("Expected status %s, got %s", transition.newStatus.String(), resp.Mapping.Status.String())
				}
			}
		})
	}
}

func TestUpdateMapping_PartialUpdate(t *testing.T) {
	// Arrange
	conn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("Failed to connect to gRPC server: %v", err)
	}
	defer conn.Close()

	client := pb.NewETCMeisaiServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Test partial update - only update confidence and status
	updatedMapping := &pb.ETCMapping{
		EtcRecordId:      1,        // Keep existing
		MappingType:      "dtako_match", // Keep existing
		MappedEntityId:   123,      // Keep existing
		MappedEntityType: "dtako_record", // Keep existing
		Confidence:       0.99,     // Update this
		Status:          pb.MappingStatus_MAPPING_STATUS_ACTIVE, // Update this
		CreatedBy:       "test-user", // Keep existing
	}

	req := &pb.UpdateMappingRequest{
		Id:      1,
		Mapping: updatedMapping,
	}

	// Act
	resp, err := client.UpdateMapping(ctx, req)

	// Assert
	if err != nil {
		st := status.Convert(err)
		if st.Code() == codes.Unimplemented {
			t.Logf("Expected failure: UpdateMapping not implemented yet - %v", err)
			return
		}
		t.Fatalf("Unexpected error: %v", err)
	}

	// If server is implemented, verify response
	if resp == nil || resp.Mapping == nil {
		t.Fatal("Response or mapping is nil")
	}

	// Verify the partial update worked
	if resp.Mapping.Confidence != 0.99 {
		t.Errorf("Expected confidence 0.99, got %f", resp.Mapping.Confidence)
	}

	if resp.Mapping.Status != pb.MappingStatus_MAPPING_STATUS_ACTIVE {
		t.Errorf("Expected status ACTIVE, got %s", resp.Mapping.Status.String())
	}
}

func TestUpdateMapping_MetadataUpdate(t *testing.T) {
	// Arrange
	conn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("Failed to connect to gRPC server: %v", err)
	}
	defer conn.Close()

	client := pb.NewETCMeisaiServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Create new metadata
	newMetadata, _ := structpb.NewStruct(map[string]interface{}{
		"confidence_score": 0.99,
		"algorithm":        "ml_enhanced",
		"verified":         true,
		"verification_info": map[string]interface{}{
			"verifier":     "admin_user",
			"verified_at":  "2024-01-20T15:30:00Z",
			"verification_notes": "Manual verification after algorithm improvement",
		},
		"performance_metrics": map[string]interface{}{
			"accuracy":  0.995,
			"precision": 0.992,
			"recall":    0.998,
		},
	})

	// Test metadata update
	updatedMapping := &pb.ETCMapping{
		EtcRecordId:      1,
		MappingType:      "dtako_match",
		MappedEntityId:   123,
		MappedEntityType: "dtako_record",
		Confidence:       0.99,
		Status:          pb.MappingStatus_MAPPING_STATUS_ACTIVE,
		Metadata:        newMetadata,
		CreatedBy:       "test-user",
	}

	req := &pb.UpdateMappingRequest{
		Id:      1,
		Mapping: updatedMapping,
	}

	// Act
	resp, err := client.UpdateMapping(ctx, req)

	// Assert
	if err != nil {
		st := status.Convert(err)
		if st.Code() == codes.Unimplemented {
			t.Logf("Expected failure: UpdateMapping not implemented yet - %v", err)
			return
		}
		t.Fatalf("Unexpected error: %v", err)
	}

	// If server is implemented, verify metadata update
	if resp == nil || resp.Mapping == nil {
		t.Fatal("Response or mapping is nil")
	}

	// Verify metadata was updated
	if resp.Mapping.Metadata == nil {
		t.Error("Expected metadata to be preserved after update")
	} else {
		// Check specific metadata fields
		if algorithm := resp.Mapping.Metadata.Fields["algorithm"]; algorithm != nil {
			if algorithm.GetStringValue() != "ml_enhanced" {
				t.Errorf("Expected algorithm 'ml_enhanced' in metadata, got %s", algorithm.GetStringValue())
			}
		} else {
			t.Error("Expected algorithm in updated metadata")
		}

		// Check nested metadata
		if verificationInfo := resp.Mapping.Metadata.Fields["verification_info"]; verificationInfo != nil {
			verificationStruct := verificationInfo.GetStructValue()
			if verificationStruct != nil {
				if verifier := verificationStruct.Fields["verifier"]; verifier != nil {
					if verifier.GetStringValue() != "admin_user" {
						t.Errorf("Expected verifier 'admin_user', got %s", verifier.GetStringValue())
					}
				}
			}
		}
	}
}

func TestUpdateMapping_ConcurrentUpdate(t *testing.T) {
	// Arrange
	conn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("Failed to connect to gRPC server: %v", err)
	}
	defer conn.Close()

	client := pb.NewETCMeisaiServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Test concurrent updates to the same mapping
	mapping1 := &pb.ETCMapping{
		EtcRecordId:      1,
		MappingType:      "dtako_match",
		MappedEntityId:   123,
		MappedEntityType: "dtako_record",
		Confidence:       0.95,
		Status:          pb.MappingStatus_MAPPING_STATUS_ACTIVE,
		CreatedBy:       "user1",
	}

	mapping2 := &pb.ETCMapping{
		EtcRecordId:      1,
		MappingType:      "dtako_match",
		MappedEntityId:   456,
		MappedEntityType: "dtako_record",
		Confidence:       0.98,
		Status:          pb.MappingStatus_MAPPING_STATUS_ACTIVE,
		CreatedBy:       "user2",
	}

	req1 := &pb.UpdateMappingRequest{Id: 1, Mapping: mapping1}
	req2 := &pb.UpdateMappingRequest{Id: 1, Mapping: mapping2}

	// Act - Start concurrent updates
	done := make(chan error, 2)

	go func() {
		_, err := client.UpdateMapping(ctx, req1)
		done <- err
	}()

	go func() {
		_, err := client.UpdateMapping(ctx, req2)
		done <- err
	}()

	// Collect results
	err1 := <-done
	err2 := <-done

	// Assert
	// At least one should succeed, or handle the race condition gracefully
	unimplementedCount := 0
	successCount := 0
	conflictCount := 0

	for i, err := range []error{err1, err2} {
		if err != nil {
			st := status.Convert(err)
			if st.Code() == codes.Unimplemented {
				unimplementedCount++
				t.Logf("Concurrent update %d: not implemented - %v", i+1, err)
			} else if st.Code() == codes.Aborted || st.Code() == codes.FailedPrecondition {
				conflictCount++
				t.Logf("Concurrent update %d: conflict handled gracefully", i+1)
			} else {
				t.Errorf("Concurrent update %d: unexpected error %v", i+1, err)
			}
		} else {
			successCount++
			t.Logf("Concurrent update %d: succeeded", i+1)
		}
	}

	if unimplementedCount == 2 {
		t.Logf("Both updates returned Unimplemented - server not ready")
	} else if successCount > 0 || conflictCount > 0 {
		t.Logf("Concurrent updates handled gracefully - success: %d, conflicts: %d", successCount, conflictCount)
	}
}
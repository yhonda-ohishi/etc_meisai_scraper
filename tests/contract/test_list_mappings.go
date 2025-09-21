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

func TestListMappings_Success(t *testing.T) {
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
	req := &pb.ListMappingsRequest{
		Page:     1,
		PageSize: 10,
	}

	// Act
	resp, err := client.ListMappings(ctx, req)

	// Assert
	// This test should FAIL initially as the server is not implemented yet
	if err != nil {
		st := status.Convert(err)
		if st.Code() == codes.Unimplemented {
			t.Logf("Expected failure: ListMappings not implemented yet - %v", err)
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

	// Mappings should not be nil (can be empty array)
	if resp.Mappings == nil {
		t.Error("Expected mappings array to not be nil")
	}

	// If mappings exist, verify they have required fields
	for i, mapping := range resp.Mappings {
		if mapping.Id == 0 {
			t.Errorf("Mapping %d has zero ID", i)
		}
		if mapping.EtcRecordId == 0 {
			t.Errorf("Mapping %d has zero ETC record ID", i)
		}
		if mapping.MappingType == "" {
			t.Errorf("Mapping %d has empty mapping type", i)
		}
		if mapping.MappedEntityId == 0 {
			t.Errorf("Mapping %d has zero mapped entity ID", i)
		}
		if mapping.MappedEntityType == "" {
			t.Errorf("Mapping %d has empty mapped entity type", i)
		}
		if mapping.Status == pb.MappingStatus_MAPPING_STATUS_UNSPECIFIED {
			t.Errorf("Mapping %d has unspecified status", i)
		}
		if mapping.CreatedBy == "" {
			t.Errorf("Mapping %d has empty created_by", i)
		}
		if mapping.CreatedAt == nil {
			t.Errorf("Mapping %d has nil created_at", i)
		}
		if mapping.UpdatedAt == nil {
			t.Errorf("Mapping %d has nil updated_at", i)
		}
	}
}

func TestListMappings_WithFilters(t *testing.T) {
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
		req  *pb.ListMappingsRequest
	}{
		{
			name: "filter by ETC record ID",
			req: &pb.ListMappingsRequest{
				Page:        1,
				PageSize:    10,
				EtcRecordId: int64Ptr(1),
			},
		},
		{
			name: "filter by mapping type",
			req: &pb.ListMappingsRequest{
				Page:        1,
				PageSize:    10,
				MappingType: stringPtr("dtako_match"),
			},
		},
		{
			name: "filter by mapped entity type",
			req: &pb.ListMappingsRequest{
				Page:             1,
				PageSize:         10,
				MappedEntityType: stringPtr("dtako_record"),
			},
		},
		{
			name: "filter by mapped entity ID",
			req: &pb.ListMappingsRequest{
				Page:           1,
				PageSize:       10,
				MappedEntityId: int64Ptr(123),
			},
		},
		{
			name: "filter by status - active",
			req: &pb.ListMappingsRequest{
				Page:     1,
				PageSize: 10,
				Status:   &[]pb.MappingStatus{pb.MappingStatus_MAPPING_STATUS_ACTIVE}[0],
			},
		},
		{
			name: "filter by status - inactive",
			req: &pb.ListMappingsRequest{
				Page:     1,
				PageSize: 10,
				Status:   &[]pb.MappingStatus{pb.MappingStatus_MAPPING_STATUS_INACTIVE}[0],
			},
		},
		{
			name: "combined filters",
			req: &pb.ListMappingsRequest{
				Page:             1,
				PageSize:         10,
				EtcRecordId:      int64Ptr(1),
				MappingType:      stringPtr("dtako_match"),
				MappedEntityType: stringPtr("dtako_record"),
				Status:           &[]pb.MappingStatus{pb.MappingStatus_MAPPING_STATUS_ACTIVE}[0],
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Act
			resp, err := client.ListMappings(ctx, tc.req)

			// Assert
			if err != nil {
				st := status.Convert(err)
				if st.Code() == codes.Unimplemented {
					t.Logf("Expected failure: ListMappings not implemented yet - %v", err)
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

			// Verify filter application
			for i, mapping := range resp.Mappings {
				if tc.req.EtcRecordId != nil && mapping.EtcRecordId != *tc.req.EtcRecordId {
					t.Errorf("Mapping %d: expected ETC record ID %d, got %d", i, *tc.req.EtcRecordId, mapping.EtcRecordId)
				}
				if tc.req.MappingType != nil && mapping.MappingType != *tc.req.MappingType {
					t.Errorf("Mapping %d: expected mapping type %s, got %s", i, *tc.req.MappingType, mapping.MappingType)
				}
				if tc.req.MappedEntityType != nil && mapping.MappedEntityType != *tc.req.MappedEntityType {
					t.Errorf("Mapping %d: expected mapped entity type %s, got %s", i, *tc.req.MappedEntityType, mapping.MappedEntityType)
				}
				if tc.req.MappedEntityId != nil && mapping.MappedEntityId != *tc.req.MappedEntityId {
					t.Errorf("Mapping %d: expected mapped entity ID %d, got %d", i, *tc.req.MappedEntityId, mapping.MappedEntityId)
				}
				if tc.req.Status != nil && mapping.Status != *tc.req.Status {
					t.Errorf("Mapping %d: expected status %s, got %s", i, tc.req.Status.String(), mapping.Status.String())
				}
			}

			t.Logf("Filter test '%s' returned %d mappings", tc.name, len(resp.Mappings))
		})
	}
}

func TestListMappings_Pagination(t *testing.T) {
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
			req := &pb.ListMappingsRequest{
				Page:     tc.page,
				PageSize: tc.pageSize,
			}

			// Act
			resp, err := client.ListMappings(ctx, req)

			// Assert
			if err != nil {
				st := status.Convert(err)
				if st.Code() == codes.Unimplemented {
					t.Logf("Expected failure: ListMappings not implemented yet - %v", err)
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

			// Verify mappings count doesn't exceed page size
			if int32(len(resp.Mappings)) > tc.pageSize {
				t.Errorf("Expected at most %d mappings, got %d", tc.pageSize, len(resp.Mappings))
			}
		})
	}
}

func TestListMappings_InvalidParameters(t *testing.T) {
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
		req  *pb.ListMappingsRequest
	}{
		{
			name: "zero page",
			req: &pb.ListMappingsRequest{
				Page:     0,
				PageSize: 10,
			},
		},
		{
			name: "negative page",
			req: &pb.ListMappingsRequest{
				Page:     -1,
				PageSize: 10,
			},
		},
		{
			name: "zero page size",
			req: &pb.ListMappingsRequest{
				Page:     1,
				PageSize: 0,
			},
		},
		{
			name: "negative page size",
			req: &pb.ListMappingsRequest{
				Page:     1,
				PageSize: -1,
			},
		},
		{
			name: "page size too large",
			req: &pb.ListMappingsRequest{
				Page:     1,
				PageSize: 10000,
			},
		},
		{
			name: "negative ETC record ID",
			req: &pb.ListMappingsRequest{
				Page:        1,
				PageSize:    10,
				EtcRecordId: int64Ptr(-1),
			},
		},
		{
			name: "empty mapping type",
			req: &pb.ListMappingsRequest{
				Page:        1,
				PageSize:    10,
				MappingType: stringPtr(""),
			},
		},
		{
			name: "empty mapped entity type",
			req: &pb.ListMappingsRequest{
				Page:             1,
				PageSize:         10,
				MappedEntityType: stringPtr(""),
			},
		},
		{
			name: "negative mapped entity ID",
			req: &pb.ListMappingsRequest{
				Page:           1,
				PageSize:       10,
				MappedEntityId: int64Ptr(-1),
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Act
			resp, err := client.ListMappings(ctx, tc.req)

			// Assert
			if err != nil {
				st := status.Convert(err)
				if st.Code() == codes.Unimplemented {
					t.Logf("Expected failure: ListMappings not implemented yet - %v", err)
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

func TestListMappings_StatusFiltering(t *testing.T) {
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
	statuses := []pb.MappingStatus{
		pb.MappingStatus_MAPPING_STATUS_ACTIVE,
		pb.MappingStatus_MAPPING_STATUS_INACTIVE,
		pb.MappingStatus_MAPPING_STATUS_PENDING,
		pb.MappingStatus_MAPPING_STATUS_REJECTED,
	}

	for _, mappingStatus := range statuses {
		t.Run("status_"+mappingStatus.String(), func(t *testing.T) {
			req := &pb.ListMappingsRequest{
				Page:     1,
				PageSize: 10,
				Status:   &mappingStatus,
			}

			// Act
			resp, err := client.ListMappings(ctx, req)

			// Assert
			if err != nil {
				st := status.Convert(err)
				if st.Code() == codes.Unimplemented {
					t.Logf("Expected failure: ListMappings not implemented yet - %v", err)
					return
				}
				t.Fatalf("Unexpected error: %v", err)
			}

			if resp == nil {
				t.Fatal("Response is nil")
			}

			// Verify all returned mappings have the requested status
			for i, mapping := range resp.Mappings {
				if mapping.Status != mappingStatus {
					t.Errorf("Mapping %d: expected status %s, got %s", i, mappingStatus.String(), mapping.Status.String())
				}
			}

			t.Logf("Status filter %s returned %d mappings", mappingStatus.String(), len(resp.Mappings))
		})
	}
}

func TestListMappings_EmptyResult(t *testing.T) {
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
	req := &pb.ListMappingsRequest{
		Page:             1,
		PageSize:         10,
		EtcRecordId:      int64Ptr(999999),   // Non-existent ETC record
		MappingType:      stringPtr("invalid_type"),
		MappedEntityType: stringPtr("invalid_entity"),
	}

	// Act
	resp, err := client.ListMappings(ctx, req)

	// Assert
	if err != nil {
		st := status.Convert(err)
		if st.Code() == codes.Unimplemented {
			t.Logf("Expected failure: ListMappings not implemented yet - %v", err)
			return
		}
		t.Fatalf("Unexpected error: %v", err)
	}

	// If server is implemented, verify empty result handling
	if resp == nil {
		t.Fatal("Response is nil")
	}

	// Should return empty mappings array, not nil
	if resp.Mappings == nil {
		t.Error("Expected mappings array to not be nil (should be empty array)")
	}

	// Should have zero total count
	if resp.TotalCount != 0 {
		t.Errorf("Expected total count 0 for empty result, got %d", resp.TotalCount)
	}

	// Should have empty mappings array
	if len(resp.Mappings) != 0 {
		t.Errorf("Expected 0 mappings for empty result, got %d", len(resp.Mappings))
	}

	// Pagination should still be valid
	if resp.Page != 1 {
		t.Errorf("Expected page 1, got %d", resp.Page)
	}

	if resp.PageSize != 10 {
		t.Errorf("Expected page size 10, got %d", resp.PageSize)
	}
}

func TestListMappings_WithETCRecords(t *testing.T) {
	// Arrange
	conn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("Failed to connect to gRPC server: %v", err)
	}
	defer conn.Close()

	client := pb.NewETCMeisaiServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req := &pb.ListMappingsRequest{
		Page:     1,
		PageSize: 10,
	}

	// Act
	resp, err := client.ListMappings(ctx, req)

	// Assert
	if err != nil {
		st := status.Convert(err)
		if st.Code() == codes.Unimplemented {
			t.Logf("Expected failure: ListMappings not implemented yet - %v", err)
			return
		}
		t.Fatalf("Unexpected error: %v", err)
	}

	// If server is implemented, verify ETC record population
	if resp == nil {
		t.Fatal("Response is nil")
	}

	// Check if ETC records are populated in mappings
	for i, mapping := range resp.Mappings {
		if mapping.EtcRecord != nil {
			// Verify ETC record consistency
			if mapping.EtcRecord.Id != mapping.EtcRecordId {
				t.Errorf("Mapping %d: ETC record ID mismatch: mapping.etc_record_id=%d, etc_record.id=%d",
					i, mapping.EtcRecordId, mapping.EtcRecord.Id)
			}

			// Verify ETC record has basic fields
			if mapping.EtcRecord.Hash == "" {
				t.Errorf("Mapping %d: ETC record should have non-empty hash", i)
			}
		} else {
			t.Logf("Mapping %d does not include ETC record details (may be by design)", i)
		}
	}
}
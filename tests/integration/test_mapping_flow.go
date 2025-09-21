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
	"google.golang.org/protobuf/types/known/structpb"
)

// TestETCMappingFlow tests the complete mapping creation and management workflow
// This integration test verifies:
// 1. Create ETC record
// 2. Create mapping to external entity
// 3. Update mapping confidence
// 4. Transition mapping status
// 5. List mappings with filters
// 6. Delete mapping
func TestETCMappingFlow(t *testing.T) {
	conn, client := setupGRPCClient(t)
	defer conn.Close()

	ctx := context.Background()

	// Step 1: Create ETC record for mapping
	etcRecord := &pb.ETCMeisaiRecord{
		Hash:          "mapping-test-123",
		Date:          "2025-09-21",
		Time:          "10:30:00",
		EntranceIc:    "東京IC",
		ExitIc:        "横浜IC",
		TollAmount:    1200,
		CarNumber:     "品川 300 あ 1234",
		EtcCardNumber: "1234567890123456",
	}

	var etcRecordId int64

	t.Run("CreateETCRecord", func(t *testing.T) {
		createReq := &pb.CreateRecordRequest{
			Record: etcRecord,
		}

		createResp, err := client.CreateRecord(ctx, createReq)
		if err != nil {
			t.Fatalf("Failed to create ETC record: %v", err)
		}

		if createResp.Record == nil {
			t.Fatal("Expected record in response, got nil")
		}

		etcRecordId = createResp.Record.Id
		if etcRecordId <= 0 {
			t.Fatalf("Expected positive record ID, got %d", etcRecordId)
		}

		t.Logf("Created ETC record with ID: %d", etcRecordId)
	})

	// Step 2: Create mapping to external entity (dtako)
	var mappingId int64

	t.Run("CreateMapping", func(t *testing.T) {
		// Create metadata for the mapping
		metadata, err := structpb.NewStruct(map[string]interface{}{
			"dtako_row_id":    "DTK-12345",
			"match_algorithm": "time_location_based",
			"created_by":      "automated_system",
			"notes":          "高速道路利用料金マッチング",
		})
		if err != nil {
			t.Fatalf("Failed to create metadata: %v", err)
		}

		mapping := &pb.ETCMapping{
			EtcRecordId:      etcRecordId,
			MappingType:      "dtako",
			MappedEntityId:   12345, // External dtako record ID
			MappedEntityType: "dtako_record",
			Confidence:       0.85,
			Status:           pb.MappingStatus_MAPPING_STATUS_PENDING,
			Metadata:         metadata,
		}

		createReq := &pb.CreateMappingRequest{
			Mapping: mapping,
		}

		createResp, err := client.CreateMapping(ctx, createReq)
		if err != nil {
			t.Fatalf("Failed to create mapping: %v", err)
		}

		if createResp.Mapping == nil {
			t.Fatal("Expected mapping in response, got nil")
		}

		mappingId = createResp.Mapping.Id
		if mappingId <= 0 {
			t.Fatalf("Expected positive mapping ID, got %d", mappingId)
		}

		// Verify the created mapping
		assertMappingEquals(t, mapping, createResp.Mapping)

		t.Logf("Created mapping with ID: %d", mappingId)
	})

	// Step 3: Update mapping confidence and status
	t.Run("UpdateMappingConfidence", func(t *testing.T) {
		// First get the current mapping
		getReq := &pb.GetMappingRequest{
			Id: mappingId,
		}

		getResp, err := client.GetMapping(ctx, getReq)
		if err != nil {
			t.Fatalf("Failed to get mapping: %v", err)
		}

		// Update confidence and status
		updatedMapping := getResp.Mapping
		updatedMapping.Confidence = 0.95 // Higher confidence
		updatedMapping.Status = pb.MappingStatus_MAPPING_STATUS_ACTIVE

		// Update metadata
		if updatedMapping.Metadata == nil {
			updatedMapping.Metadata = &structpb.Struct{Fields: make(map[string]*structpb.Value)}
		}
		updatedMapping.Metadata.Fields["updated_at"] = structpb.NewStringValue(time.Now().Format(time.RFC3339))
		updatedMapping.Metadata.Fields["updated_by"] = structpb.NewStringValue("integration_test")

		updateReq := &pb.UpdateMappingRequest{
			Id:      mappingId,
			Mapping: updatedMapping,
		}

		updateResp, err := client.UpdateMapping(ctx, updateReq)
		if err != nil {
			t.Fatalf("Failed to update mapping: %v", err)
		}

		if updateResp.Mapping == nil {
			t.Fatal("Expected mapping in response, got nil")
		}

		// Verify the updates
		if updateResp.Mapping.Confidence != 0.95 {
			t.Errorf("Expected confidence 0.95, got %f", updateResp.Mapping.Confidence)
		}

		if updateResp.Mapping.Status != pb.MappingStatus_MAPPING_STATUS_ACTIVE {
			t.Errorf("Expected status ACTIVE, got %v", updateResp.Mapping.Status)
		}

		t.Logf("Updated mapping confidence to %.2f and status to ACTIVE", updateResp.Mapping.Confidence)
	})

	// Step 4: Test mapping status transitions
	t.Run("MappingStatusTransitions", func(t *testing.T) {
		// Test transition from ACTIVE to INACTIVE
		getReq := &pb.GetMappingRequest{Id: mappingId}
		getResp, err := client.GetMapping(ctx, getReq)
		if err != nil {
			t.Fatalf("Failed to get mapping: %v", err)
		}

		// Transition to INACTIVE
		inactiveMapping := getResp.Mapping
		inactiveMapping.Status = pb.MappingStatus_MAPPING_STATUS_INACTIVE
		inactiveMapping.Metadata.Fields["status_reason"] = structpb.NewStringValue("temporarily disabled")

		updateReq := &pb.UpdateMappingRequest{
			Id:      mappingId,
			Mapping: inactiveMapping,
		}

		updateResp, err := client.UpdateMapping(ctx, updateReq)
		if err != nil {
			t.Fatalf("Failed to update mapping to INACTIVE: %v", err)
		}

		if updateResp.Mapping.Status != pb.MappingStatus_MAPPING_STATUS_INACTIVE {
			t.Errorf("Expected status INACTIVE, got %v", updateResp.Mapping.Status)
		}

		// Transition back to ACTIVE
		activeMapping := updateResp.Mapping
		activeMapping.Status = pb.MappingStatus_MAPPING_STATUS_ACTIVE
		activeMapping.Metadata.Fields["status_reason"] = structpb.NewStringValue("re-activated")

		updateReq2 := &pb.UpdateMappingRequest{
			Id:      mappingId,
			Mapping: activeMapping,
		}

		updateResp2, err := client.UpdateMapping(ctx, updateReq2)
		if err != nil {
			t.Fatalf("Failed to update mapping back to ACTIVE: %v", err)
		}

		if updateResp2.Mapping.Status != pb.MappingStatus_MAPPING_STATUS_ACTIVE {
			t.Errorf("Expected status ACTIVE, got %v", updateResp2.Mapping.Status)
		}

		t.Log("Successfully tested status transitions: ACTIVE -> INACTIVE -> ACTIVE")
	})

	// Step 5: List mappings with filters
	t.Run("ListMappingsWithFilters", func(t *testing.T) {
		// Test listing by ETC record ID
		listReq := &pb.ListMappingsRequest{
			Page:        1,
			PageSize:    10,
			EtcRecordId: &etcRecordId,
		}

		listResp, err := client.ListMappings(ctx, listReq)
		if err != nil {
			t.Fatalf("Failed to list mappings by ETC record ID: %v", err)
		}

		if len(listResp.Mappings) == 0 {
			t.Fatal("Expected to find mappings for ETC record")
		}

		// Find our mapping
		var foundMapping *pb.ETCMapping
		for _, mapping := range listResp.Mappings {
			if mapping.Id == mappingId {
				foundMapping = mapping
				break
			}
		}

		if foundMapping == nil {
			t.Fatalf("Could not find our mapping (ID: %d) in list", mappingId)
		}

		// Test filtering by mapping type
		mappingType := "dtako"
		listReq2 := &pb.ListMappingsRequest{
			Page:        1,
			PageSize:    10,
			MappingType: &mappingType,
		}

		listResp2, err := client.ListMappings(ctx, listReq2)
		if err != nil {
			t.Fatalf("Failed to list mappings by type: %v", err)
		}

		// Should find at least our mapping
		found := false
		for _, mapping := range listResp2.Mappings {
			if mapping.Id == mappingId {
				found = true
				break
			}
		}

		if !found {
			t.Error("Could not find our mapping when filtering by type")
		}

		// Test filtering by status
		status := pb.MappingStatus_MAPPING_STATUS_ACTIVE
		listReq3 := &pb.ListMappingsRequest{
			Page:     1,
			PageSize: 10,
			Status:   &status,
		}

		listResp3, err := client.ListMappings(ctx, listReq3)
		if err != nil {
			t.Fatalf("Failed to list mappings by status: %v", err)
		}

		// Should find at least our mapping
		found = false
		for _, mapping := range listResp3.Mappings {
			if mapping.Id == mappingId {
				found = true
				break
			}
		}

		if !found {
			t.Error("Could not find our mapping when filtering by status")
		}

		t.Logf("Successfully tested mapping filters - found mapping in all filtered lists")
	})

	// Step 6: Create additional mappings for testing
	t.Run("CreateAdditionalMappings", func(t *testing.T) {
		// Create a second mapping for the same ETC record (different entity)
		metadata2, err := structpb.NewStruct(map[string]interface{}{
			"external_id":     "EXT-67890",
			"match_algorithm": "fuzzy_matching",
			"created_by":      "manual_review",
		})
		if err != nil {
			t.Fatalf("Failed to create metadata2: %v", err)
		}

		mapping2 := &pb.ETCMapping{
			EtcRecordId:      etcRecordId,
			MappingType:      "external_system",
			MappedEntityId:   67890,
			MappedEntityType: "external_record",
			Confidence:       0.75,
			Status:           pb.MappingStatus_MAPPING_STATUS_PENDING,
			Metadata:         metadata2,
		}

		createReq2 := &pb.CreateMappingRequest{
			Mapping: mapping2,
		}

		createResp2, err := client.CreateMapping(ctx, createReq2)
		if err != nil {
			t.Fatalf("Failed to create second mapping: %v", err)
		}

		if createResp2.Mapping == nil {
			t.Fatal("Expected second mapping in response, got nil")
		}

		// Now we should have 2 mappings for the same ETC record
		listReq := &pb.ListMappingsRequest{
			Page:        1,
			PageSize:    10,
			EtcRecordId: &etcRecordId,
		}

		listResp, err := client.ListMappings(ctx, listReq)
		if err != nil {
			t.Fatalf("Failed to list mappings for verification: %v", err)
		}

		if len(listResp.Mappings) < 2 {
			t.Errorf("Expected at least 2 mappings for ETC record, got %d", len(listResp.Mappings))
		}

		t.Logf("Successfully created additional mapping - total mappings for ETC record: %d", len(listResp.Mappings))
	})

	// Step 7: Test mapping deletion
	t.Run("DeleteMapping", func(t *testing.T) {
		deleteReq := &pb.DeleteMappingRequest{
			Id: mappingId,
		}

		_, err := client.DeleteMapping(ctx, deleteReq)
		if err != nil {
			t.Fatalf("Failed to delete mapping: %v", err)
		}

		// Verify deletion
		getReq := &pb.GetMappingRequest{
			Id: mappingId,
		}

		_, err = client.GetMapping(ctx, getReq)
		if err == nil {
			t.Fatal("Expected error when getting deleted mapping, but got none")
		}

		// Verify it's a NOT_FOUND error
		st, ok := status.FromError(err)
		if !ok {
			t.Fatalf("Expected gRPC status error, got: %v", err)
		}

		if st.Code() != codes.NotFound {
			t.Fatalf("Expected NOT_FOUND error, got: %v", st.Code())
		}

		t.Log("Successfully deleted mapping and verified deletion")
	})

	// Cleanup: Delete the ETC record
	t.Run("CleanupETCRecord", func(t *testing.T) {
		deleteReq := &pb.DeleteRecordRequest{
			Id: etcRecordId,
		}

		_, err := client.DeleteRecord(ctx, deleteReq)
		if err != nil {
			t.Fatalf("Failed to delete ETC record: %v", err)
		}

		t.Log("Successfully cleaned up ETC record")
	})
}

// TestMappingErrorHandling tests error scenarios in mapping operations
func TestMappingErrorHandling(t *testing.T) {
	conn, client := setupGRPCClient(t)
	defer conn.Close()

	ctx := context.Background()

	t.Run("CreateMappingWithNonExistentETCRecord", func(t *testing.T) {
		mapping := &pb.ETCMapping{
			EtcRecordId:      999999, // Non-existent ETC record
			MappingType:      "dtako",
			MappedEntityId:   12345,
			MappedEntityType: "dtako_record",
			Confidence:       0.85,
			Status:           pb.MappingStatus_MAPPING_STATUS_PENDING,
		}

		createReq := &pb.CreateMappingRequest{
			Mapping: mapping,
		}

		_, err := client.CreateMapping(ctx, createReq)
		if err == nil {
			t.Fatal("Expected error when creating mapping with non-existent ETC record")
		}

		st, ok := status.FromError(err)
		if !ok {
			t.Fatalf("Expected gRPC status error, got: %v", err)
		}

		// Should be either NOT_FOUND or INVALID_ARGUMENT
		if st.Code() != codes.NotFound && st.Code() != codes.InvalidArgument {
			t.Fatalf("Expected NOT_FOUND or INVALID_ARGUMENT error, got: %v", st.Code())
		}
	})

	t.Run("CreateMappingWithInvalidData", func(t *testing.T) {
		mapping := &pb.ETCMapping{
			EtcRecordId:      1, // Valid ETC record ID
			MappingType:      "", // Empty mapping type
			MappedEntityId:   0,  // Invalid entity ID
			MappedEntityType: "",
			Confidence:       -0.5, // Invalid confidence (negative)
			Status:           pb.MappingStatus_MAPPING_STATUS_UNSPECIFIED,
		}

		createReq := &pb.CreateMappingRequest{
			Mapping: mapping,
		}

		_, err := client.CreateMapping(ctx, createReq)
		if err == nil {
			t.Fatal("Expected error when creating mapping with invalid data")
		}

		st, ok := status.FromError(err)
		if !ok {
			t.Fatalf("Expected gRPC status error, got: %v", err)
		}

		if st.Code() != codes.InvalidArgument {
			t.Fatalf("Expected INVALID_ARGUMENT error, got: %v", st.Code())
		}
	})

	t.Run("GetNonExistentMapping", func(t *testing.T) {
		getReq := &pb.GetMappingRequest{
			Id: 999999, // Non-existent mapping ID
		}

		_, err := client.GetMapping(ctx, getReq)
		if err == nil {
			t.Fatal("Expected error when getting non-existent mapping")
		}

		st, ok := status.FromError(err)
		if !ok {
			t.Fatalf("Expected gRPC status error, got: %v", err)
		}

		if st.Code() != codes.NotFound {
			t.Fatalf("Expected NOT_FOUND error, got: %v", st.Code())
		}
	})

	t.Run("UpdateNonExistentMapping", func(t *testing.T) {
		mapping := &pb.ETCMapping{
			Id:               999999, // Non-existent mapping ID
			EtcRecordId:      1,
			MappingType:      "dtako",
			MappedEntityId:   12345,
			MappedEntityType: "dtako_record",
			Confidence:       0.85,
			Status:           pb.MappingStatus_MAPPING_STATUS_ACTIVE,
		}

		updateReq := &pb.UpdateMappingRequest{
			Id:      999999,
			Mapping: mapping,
		}

		_, err := client.UpdateMapping(ctx, updateReq)
		if err == nil {
			t.Fatal("Expected error when updating non-existent mapping")
		}

		st, ok := status.FromError(err)
		if !ok {
			t.Fatalf("Expected gRPC status error, got: %v", err)
		}

		if st.Code() != codes.NotFound {
			t.Fatalf("Expected NOT_FOUND error, got: %v", st.Code())
		}
	})

	t.Run("DeleteNonExistentMapping", func(t *testing.T) {
		deleteReq := &pb.DeleteMappingRequest{
			Id: 999999, // Non-existent mapping ID
		}

		_, err := client.DeleteMapping(ctx, deleteReq)
		if err == nil {
			t.Fatal("Expected error when deleting non-existent mapping")
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

// TestMappingAdvancedScenarios tests complex mapping scenarios
func TestMappingAdvancedScenarios(t *testing.T) {
	conn, client := setupGRPCClient(t)
	defer conn.Close()

	ctx := context.Background()

	// Create multiple ETC records for testing
	recordIds := make([]int64, 0)

	for i := 0; i < 3; i++ {
		etcRecord := &pb.ETCMeisaiRecord{
			Hash:          fmt.Sprintf("advanced-test-%d", i),
			Date:          "2025-09-21",
			Time:          fmt.Sprintf("1%d:30:00", i),
			EntranceIc:    "東京IC",
			ExitIc:        "横浜IC",
			TollAmount:    1200 + int32(i*100),
			CarNumber:     fmt.Sprintf("品川 300 あ %04d", 1000+i),
			EtcCardNumber: fmt.Sprintf("%016d", 1234567890123456+int64(i)),
		}

		createReq := &pb.CreateRecordRequest{Record: etcRecord}
		createResp, err := client.CreateRecord(ctx, createReq)
		if err != nil {
			t.Fatalf("Failed to create ETC record %d: %v", i, err)
		}

		recordIds = append(recordIds, createResp.Record.Id)
	}

	defer func() {
		// Cleanup records
		for _, recordId := range recordIds {
			deleteReq := &pb.DeleteRecordRequest{Id: recordId}
			client.DeleteRecord(ctx, deleteReq)
		}
	}()

	t.Run("BulkMappingCreation", func(t *testing.T) {
		// Create multiple mappings for different types
		mappingTypes := []string{"dtako", "fuel_card", "gps_tracking"}

		for i, recordId := range recordIds {
			for j, mappingType := range mappingTypes {
				mapping := &pb.ETCMapping{
					EtcRecordId:      recordId,
					MappingType:      mappingType,
					MappedEntityId:   int64(10000 + i*100 + j),
					MappedEntityType: fmt.Sprintf("%s_record", mappingType),
					Confidence:       0.8 + float32(j)*0.05,
					Status:           pb.MappingStatus_MAPPING_STATUS_ACTIVE,
				}

				createReq := &pb.CreateMappingRequest{Mapping: mapping}
				_, err := client.CreateMapping(ctx, createReq)
				if err != nil {
					t.Fatalf("Failed to create mapping for record %d, type %s: %v", recordId, mappingType, err)
				}
			}
		}

		t.Logf("Successfully created %d mappings", len(recordIds)*len(mappingTypes))
	})

	t.Run("ComplexFilteringScenarios", func(t *testing.T) {
		// Test filtering by mapped entity type
		entityType := "dtako_record"
		listReq := &pb.ListMappingsRequest{
			Page:             1,
			PageSize:         20,
			MappedEntityType: &entityType,
		}

		listResp, err := client.ListMappings(ctx, listReq)
		if err != nil {
			t.Fatalf("Failed to list mappings by entity type: %v", err)
		}

		// Should find dtako mappings
		if len(listResp.Mappings) == 0 {
			t.Error("Expected to find dtako mappings")
		}

		// Verify all returned mappings are of the requested type
		for _, mapping := range listResp.Mappings {
			if mapping.MappedEntityType != entityType {
				t.Errorf("Expected entity type %s, got %s", entityType, mapping.MappedEntityType)
			}
		}

		// Test multiple filters combined
		mappingType := "fuel_card"
		status := pb.MappingStatus_MAPPING_STATUS_ACTIVE
		listReq2 := &pb.ListMappingsRequest{
			Page:        1,
			PageSize:    20,
			MappingType: &mappingType,
			Status:      &status,
		}

		listResp2, err := client.ListMappings(ctx, listReq2)
		if err != nil {
			t.Fatalf("Failed to list mappings with multiple filters: %v", err)
		}

		// Verify all returned mappings match both filters
		for _, mapping := range listResp2.Mappings {
			if mapping.MappingType != mappingType {
				t.Errorf("Expected mapping type %s, got %s", mappingType, mapping.MappingType)
			}
			if mapping.Status != status {
				t.Errorf("Expected status %v, got %v", status, mapping.Status)
			}
		}

		t.Log("Successfully tested complex filtering scenarios")
	})

	t.Run("MappingPagination", func(t *testing.T) {
		// Test pagination with small page size
		pageSize := int32(2)
		allMappings := make([]*pb.ETCMapping, 0)

		page := int32(1)
		for {
			listReq := &pb.ListMappingsRequest{
				Page:     page,
				PageSize: pageSize,
			}

			listResp, err := client.ListMappings(ctx, listReq)
			if err != nil {
				t.Fatalf("Failed to list mappings page %d: %v", page, err)
			}

			allMappings = append(allMappings, listResp.Mappings...)

			// Verify pagination metadata
			if listResp.Page != page {
				t.Errorf("Expected page %d, got %d", page, listResp.Page)
			}
			if listResp.PageSize != pageSize {
				t.Errorf("Expected page size %d, got %d", pageSize, listResp.PageSize)
			}

			// Check if we have more pages
			if len(listResp.Mappings) < int(pageSize) {
				break
			}

			page++
			if page > 10 { // Safety check to avoid infinite loop
				break
			}
		}

		if len(allMappings) == 0 {
			t.Error("Expected to find mappings through pagination")
		}

		t.Logf("Successfully paginated through %d mappings across %d pages", len(allMappings), page)
	})
}

// assertMappingEquals compares two ETCMapping instances for equality
func assertMappingEquals(t *testing.T, expected, actual *pb.ETCMapping) {
	t.Helper()

	if expected.EtcRecordId != actual.EtcRecordId {
		t.Errorf("EtcRecordId mismatch: expected %d, got %d", expected.EtcRecordId, actual.EtcRecordId)
	}
	if expected.MappingType != actual.MappingType {
		t.Errorf("MappingType mismatch: expected %s, got %s", expected.MappingType, actual.MappingType)
	}
	if expected.MappedEntityId != actual.MappedEntityId {
		t.Errorf("MappedEntityId mismatch: expected %d, got %d", expected.MappedEntityId, actual.MappedEntityId)
	}
	if expected.MappedEntityType != actual.MappedEntityType {
		t.Errorf("MappedEntityType mismatch: expected %s, got %s", expected.MappedEntityType, actual.MappedEntityType)
	}
	if expected.Confidence != actual.Confidence {
		t.Errorf("Confidence mismatch: expected %f, got %f", expected.Confidence, actual.Confidence)
	}
	if expected.Status != actual.Status {
		t.Errorf("Status mismatch: expected %v, got %v", expected.Status, actual.Status)
	}
	if expected.Id != 0 && expected.Id != actual.Id {
		t.Errorf("ID mismatch: expected %d, got %d", expected.Id, actual.Id)
	}
}
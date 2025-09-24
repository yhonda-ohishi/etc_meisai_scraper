package contract

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	pb "github.com/yhonda-ohishi/etc_meisai/src/pb"
)

// int64Ptr helper function (stringPtr and mappingStatusPtr are in end_to_end_workflow_test.go)
func int64Ptr(i int64) *int64 {
	return &i
}

// TestAPIVersionCompatibility validates API version compatibility between client and server
func TestAPIVersionCompatibility(t *testing.T) {
	// Skip test if no server available
	t.Skip("Version compatibility test server setup required")

	client := setupVersionTestClient(t)
	ctx := context.Background()

	t.Run("Version_Header_Negotiation", func(t *testing.T) {
		// Contract: Server must handle version headers correctly
		md := metadata.Pairs("api-version", "v1.0")
		ctx := metadata.NewOutgoingContext(ctx, md)

		req := &pb.ListRecordsRequest{
			PageSize: 10,
		}

		resp, err := client.ListRecords(ctx, req)

		// Contract assertions
		assert.NoError(t, err, "Server must handle versioned requests")
		assert.NotNil(t, resp, "Response must not be nil for versioned request")
	})

	t.Run("Backward_Compatibility", func(t *testing.T) {
		// Contract: Server must handle requests with only required fields
		req := &pb.CreateRecordRequest{
			Record: &pb.ETCMeisaiRecord{
				Hash:          "test-001",
				Date:          "2024-01-01",
				Time:          "10:00:00",
				EntranceIc:    "TestIC",
				ExitIc:        "TestExitIC",
				TollAmount:    1500,
				CarNumber:     "品川123あ1234",
				EtcCardNumber: "1234567890123456",
			},
		}

		resp, err := client.CreateRecord(ctx, req)

		assert.NoError(t, err, "Server must accept v1.0 compatible requests")
		assert.NotNil(t, resp, "Server must return valid response")
		assert.NotNil(t, resp.Record, "Created record must be returned")
	})

	t.Run("Forward_Compatibility", func(t *testing.T) {
		// Contract: Server must handle requests with newer optional fields
		req := &pb.CreateRecordRequest{
			Record: &pb.ETCMeisaiRecord{
				Hash:          "test-002",
				Date:          "2024-01-01",
				Time:          "11:00:00",
				EntranceIc:    "TestIC2",
				ExitIc:        "TestExitIC2",
				TollAmount:    2000,
				CarNumber:     "品川123あ5678",
				EtcCardNumber: "2345678901234567",
				EtcNum:        stringPtr("FORWARD001"),
				DtakoRowId:    int64Ptr(12345),
			},
		}

		resp, err := client.CreateRecord(ctx, req)

		assert.NoError(t, err, "Server must handle newer optional fields")
		assert.NotNil(t, resp, "Server must return valid response with newer fields")
	})

	t.Run("Enum_Compatibility", func(t *testing.T) {
		// Contract: Server must handle unknown enum values gracefully
		createReq := &pb.CreateMappingRequest{
			Mapping: &pb.ETCMapping{
				EtcRecordId:      12345,
				MappingType:      "manual",
				MappedEntityId:   67890,
				MappedEntityType: "test_entity",
				Confidence:       0.95,
				Status:           pb.MappingStatus(999), // Unknown enum value
			},
		}

		_, err := client.CreateMapping(ctx, createReq)

		// Server should either accept with default or return clear error
		if err != nil {
			assert.Contains(t, err.Error(), "status", "Error should mention the problematic field")
		}
	})

	t.Run("Statistics_Response_Evolution", func(t *testing.T) {
		// Contract: Statistics response must maintain backward compatibility
		req := &pb.GetStatisticsRequest{
			DateFrom: stringPtr("2024-01-01"),
			DateTo:   stringPtr("2024-01-31"),
		}

		resp, err := client.GetStatistics(ctx, req)
		require.NoError(t, err)

		// Core statistics fields must always be present
		assert.NotNil(t, resp, "Response must exist")
		assert.GreaterOrEqual(t, resp.TotalRecords, int64(0), "TotalRecords must be present")
		assert.GreaterOrEqual(t, resp.TotalAmount, int64(0), "TotalAmount must be present")

		// Optional newer fields may be present but not required
		if resp.DailyStats != nil {
			for _, stat := range resp.DailyStats {
				assert.NotEmpty(t, stat.Date, "Date must be present in daily stats")
				assert.GreaterOrEqual(t, stat.RecordCount, int32(0), "RecordCount must be valid")
			}
		}
	})

	t.Run("Timestamp_Field_Compatibility", func(t *testing.T) {
		// Contract: Server must handle timestamp fields consistently
		createReq := &pb.CreateRecordRequest{
			Record: &pb.ETCMeisaiRecord{
				Hash:          "timestamp-test",
				Date:          "2024-01-01",
				Time:          "12:00:00",
				EntranceIc:    "TimestampTestIC",
				ExitIc:        "TimestampTestExitIC",
				TollAmount:    1800,
				CarNumber:     "品川123あ7777",
				EtcCardNumber: "7777888899990000",
			},
		}

		resp, err := client.CreateRecord(ctx, createReq)
		require.NoError(t, err)

		// Timestamp assertions
		assert.NotNil(t, resp.Record.CreatedAt, "CreatedAt timestamp must be set")
		assert.NotNil(t, resp.Record.UpdatedAt, "UpdatedAt timestamp must be set")

		// Parse and validate timestamps
		createdTime := resp.Record.CreatedAt.AsTime()
		updatedTime := resp.Record.UpdatedAt.AsTime()
		assert.True(t, updatedTime.Equal(createdTime) || updatedTime.After(createdTime),
			"UpdatedAt should be >= CreatedAt")
	})

	t.Run("Optional_Field_Handling", func(t *testing.T) {
		// Test with EtcNum present
		req1 := &pb.CreateRecordRequest{
			Record: &pb.ETCMeisaiRecord{
				Hash:          "optional-test-001",
				Date:          "2024-01-01",
				Time:          "13:00:00",
				EntranceIc:    "OptionalTestIC",
				ExitIc:        "OptionalTestExitIC",
				TollAmount:    2200,
				CarNumber:     "品川123あ8888",
				EtcCardNumber: "8888999900001111",
				EtcNum:        stringPtr("OPTIONAL001"),
			},
		}

		resp1, err := client.CreateRecord(ctx, req1)
		require.NoError(t, err)

		// Optional field should be preserved when provided
		assert.NotNil(t, resp1.Record.EtcNum, "EtcNum should be preserved when provided")
		assert.Equal(t, "OPTIONAL001", *resp1.Record.EtcNum, "EtcNum value should match")

		// Test without EtcNum (optional field absent)
		req2 := &pb.CreateRecordRequest{
			Record: &pb.ETCMeisaiRecord{
				Hash:          "optional-test-002",
				Date:          "2024-01-01",
				Time:          "14:00:00",
				EntranceIc:    "OptionalTest2IC",
				ExitIc:        "OptionalTest2ExitIC",
				TollAmount:    2500,
				CarNumber:     "品川123あ9999",
				EtcCardNumber: "9999000011112222",
			},
		}

		resp2, err := client.CreateRecord(ctx, req2)
		require.NoError(t, err)

		// Server should handle absent optional field gracefully
		assert.NotNil(t, resp2.Record, "Response should be valid without optional field")
	})

	t.Run("API_Deprecation_Warning", func(t *testing.T) {
		// Contract: Server may return deprecation warnings in metadata
		md := metadata.Pairs("api-version", "v0.9")
		ctx := metadata.NewOutgoingContext(ctx, md)

		req := &pb.ListRecordsRequest{
			PageSize: 10,
		}

		var responseMetadata metadata.MD
		_, err := client.ListRecords(ctx, req, grpc.Header(&responseMetadata))

		// Check for deprecation warnings (if implemented)
		if err == nil && responseMetadata != nil {
			warnings := responseMetadata.Get("deprecation-warning")
			if len(warnings) > 0 {
				assert.NotEmpty(t, warnings[0], "Deprecation warning should have content")
			}
		}
	})

	t.Run("Field_Size_Evolution", func(t *testing.T) {
		// Contract: Field size limits should be backward compatible
		longString := make([]byte, 1000)
		for i := range longString {
			longString[i] = 'A'
		}

		req := &pb.CreateRecordRequest{
			Record: &pb.ETCMeisaiRecord{
				Hash:          string(longString[:64]), // Assuming 64 char limit
				Date:          "2024-01-01",
				Time:          "15:00:00",
				EntranceIc:    string(longString[:100]), // Test IC name limits
				ExitIc:        string(longString[:100]),
				TollAmount:    999999, // Test max amount
				CarNumber:     "品川123あ9999",
				EtcCardNumber: "9999888877776666",
			},
		}

		resp, err := client.CreateRecord(ctx, req)

		// Server should either accept or return clear size limit error
		if err != nil {
			assert.Contains(t, err.Error(), "size", "Error should indicate size limit")
		} else {
			assert.NotNil(t, resp.Record, "Large fields should be handled")
		}
	})

	t.Run("Pagination_Evolution", func(t *testing.T) {
		// Contract: Pagination should remain backward compatible
		req := &pb.ListRecordsRequest{
			PageSize: 20,
			Page:     1,
		}

		resp, err := client.ListRecords(ctx, req)
		require.NoError(t, err)

		// Core pagination fields must exist
		assert.NotNil(t, resp.Records, "Records array must exist even if empty")
		// Check if there are more pages
		if len(resp.Records) == 20 {
			// Second page request should work
			req2 := &pb.ListRecordsRequest{
				PageSize: 20,
				Page:     2,
			}
			resp2, err2 := client.ListRecords(ctx, req2)
			assert.NoError(t, err2, "Second page should work")
			assert.NotNil(t, resp2, "Second page response should be valid")
		}
	})
}


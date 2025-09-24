package contract

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/yhonda-ohishi/etc_meisai/src/pb"
)

// TestGRPCServiceContract_T010A validates contract testing for all gRPC service definitions
// This test ensures all gRPC methods defined in etc_meisai.proto are correctly implemented
func TestGRPCServiceContract_T010A(t *testing.T) {
	client := setupGRPCTestClient(t)
	ctx := context.Background()

	t.Run("ETCMeisaiService_CreateRecord_Contract", func(t *testing.T) {
		// Contract: CreateRecord must accept valid ETCMeisaiRecord and return created record
		req := &pb.CreateRecordRequest{
			Record: &pb.ETCMeisaiRecord{
				Hash:           "test-hash-001",
				Date:           "2024-01-01",
				Time:           "10:30:00",
				EntranceIc:     "テスト入口IC",
				ExitIc:         "テスト出口IC",
				TollAmount:     1500,
				CarNumber:      "品川123あ1234",
				EtcCardNumber:  "1234567890123456",
				EtcNum:         stringPtr("TEST001"),
			},
		}

		resp, err := client.CreateRecord(ctx, req)

		// Contract assertions
		assert.NoError(t, err, "CreateRecord must succeed for valid input")
		assert.NotNil(t, resp, "CreateRecord must return response")
		assert.NotNil(t, resp.Record, "CreateRecord must return record")
		assert.NotZero(t, resp.Record.Id, "Created record must have non-zero ID")
		assert.Equal(t, req.Record.Hash, resp.Record.Hash, "Hash must be preserved")
		assert.Equal(t, req.Record.Date, resp.Record.Date, "Date must be preserved")
		assert.Equal(t, req.Record.Time, resp.Record.Time, "Time must be preserved")
		assert.NotNil(t, resp.Record.CreatedAt, "CreatedAt must be set")
		assert.NotNil(t, resp.Record.UpdatedAt, "UpdatedAt must be set")
	})

	t.Run("ETCMeisaiService_GetRecord_Contract", func(t *testing.T) {
		// Contract: GetRecord must return exact record for valid ID
		// First create a record
		createReq := &pb.CreateRecordRequest{
			Record: &pb.ETCMeisaiRecord{
				Hash:           "get-test-hash-001",
				Date:           "2024-01-02",
				Time:           "11:00:00",
				EntranceIc:     "取得テスト入口IC",
				ExitIc:         "取得テスト出口IC",
				TollAmount:     2000,
				CarNumber:      "品川123あ5678",
				EtcCardNumber:  "2345678901234567",
				EtcNum:         stringPtr("GET001"),
			},
		}

		createResp, err := client.CreateRecord(ctx, createReq)
		require.NoError(t, err)

		// Test the contract
		getReq := &pb.GetRecordRequest{
			Id: createResp.Record.Id,
		}

		getResp, err := client.GetRecord(ctx, getReq)

		// Contract assertions
		assert.NoError(t, err, "GetRecord must succeed for valid ID")
		assert.NotNil(t, getResp, "GetRecord must return response")
		assert.NotNil(t, getResp.Record, "GetRecord must return record")
		assert.Equal(t, createResp.Record.Id, getResp.Record.Id, "Retrieved record must have same ID")
		assert.Equal(t, createResp.Record.Hash, getResp.Record.Hash, "Retrieved record must have same hash")
		assert.Equal(t, createResp.Record.EtcNum, getResp.Record.EtcNum, "Retrieved record must have same ETC number")
	})

	t.Run("ETCMeisaiService_ListRecords_Contract", func(t *testing.T) {
		// Contract: ListRecords must return paginated results with proper metadata
		req := &pb.ListRecordsRequest{
			Page:     1,
			PageSize: 10,
			SortBy:   "created_at",
			SortOrder: pb.SortOrder_SORT_ORDER_DESC,
		}

		resp, err := client.ListRecords(ctx, req)

		// Contract assertions
		assert.NoError(t, err, "ListRecords must succeed")
		assert.NotNil(t, resp, "ListRecords must return response")
		assert.NotNil(t, resp.Records, "ListRecords must return records slice")
		assert.GreaterOrEqual(t, resp.TotalCount, int32(0), "TotalCount must be non-negative")
		assert.Equal(t, req.Page, resp.Page, "Response page must match request")
		assert.Equal(t, req.PageSize, resp.PageSize, "Response page size must match request")
		assert.LessOrEqual(t, int32(len(resp.Records)), req.PageSize, "Returned records must not exceed page size")
	})

	t.Run("ETCMeisaiService_UpdateRecord_Contract", func(t *testing.T) {
		// Contract: UpdateRecord must update existing record and return updated data
		// First create a record
		createReq := &pb.CreateRecordRequest{
			Record: &pb.ETCMeisaiRecord{
				Hash:           "update-test-hash-001",
				Date:           "2024-01-03",
				Time:           "12:00:00",
				EntranceIc:     "更新テスト入口IC",
				ExitIc:         "更新テスト出口IC",
				TollAmount:     2500,
				CarNumber:      "品川123あ9999",
				EtcCardNumber:  "3456789012345678",
				EtcNum:         stringPtr("UPDATE001"),
			},
		}

		createResp, err := client.CreateRecord(ctx, createReq)
		require.NoError(t, err)

		// Update the record
		updateReq := &pb.UpdateRecordRequest{
			Id: createResp.Record.Id,
			Record: &pb.ETCMeisaiRecord{
				Id:             createResp.Record.Id,
				Hash:           createResp.Record.Hash,
				Date:           createResp.Record.Date,
				Time:           createResp.Record.Time,
				EntranceIc:     createResp.Record.EntranceIc,
				ExitIc:         createResp.Record.ExitIc,
				TollAmount:     3500, // Updated amount
				CarNumber:      createResp.Record.CarNumber,
				EtcCardNumber:  createResp.Record.EtcCardNumber,
				EtcNum:         createResp.Record.EtcNum,
				CreatedAt:      createResp.Record.CreatedAt,
			},
		}

		updateResp, err := client.UpdateRecord(ctx, updateReq)

		// Contract assertions
		assert.NoError(t, err, "UpdateRecord must succeed for valid input")
		assert.NotNil(t, updateResp, "UpdateRecord must return response")
		assert.NotNil(t, updateResp.Record, "UpdateRecord must return record")
		assert.Equal(t, updateReq.Id, updateResp.Record.Id, "Updated record must preserve ID")
		assert.Equal(t, int32(3500), updateResp.Record.TollAmount, "TollAmount must be updated")
		assert.Equal(t, createResp.Record.CreatedAt.AsTime().Unix(), updateResp.Record.CreatedAt.AsTime().Unix(), "CreatedAt must not change")
		assert.True(t, updateResp.Record.UpdatedAt.AsTime().After(createResp.Record.UpdatedAt.AsTime()), "UpdatedAt must be newer")
	})

	t.Run("ETCMeisaiService_DeleteRecord_Contract", func(t *testing.T) {
		// Contract: DeleteRecord must successfully delete existing record
		// First create a record
		createReq := &pb.CreateRecordRequest{
			Record: &pb.ETCMeisaiRecord{
				Hash:           "delete-test-hash-001",
				Date:           "2024-01-04",
				Time:           "13:00:00",
				EntranceIc:     "削除テスト入口IC",
				ExitIc:         "削除テスト出口IC",
				TollAmount:     1800,
				CarNumber:      "品川123あ0000",
				EtcCardNumber:  "4567890123456789",
				EtcNum:         stringPtr("DELETE001"),
			},
		}

		createResp, err := client.CreateRecord(ctx, createReq)
		require.NoError(t, err)

		// Delete the record
		deleteReq := &pb.DeleteRecordRequest{
			Id: createResp.Record.Id,
		}

		deleteResp, err := client.DeleteRecord(ctx, deleteReq)

		// Contract assertions
		assert.NoError(t, err, "DeleteRecord must succeed for valid ID")
		assert.NotNil(t, deleteResp, "DeleteRecord must return response")

		// Verify record is deleted
		getReq := &pb.GetRecordRequest{
			Id: createResp.Record.Id,
		}

		_, err = client.GetRecord(ctx, getReq)
		assert.Error(t, err, "GetRecord must fail for deleted record")

		st, ok := status.FromError(err)
		assert.True(t, ok, "Error must be gRPC status")
		assert.Equal(t, codes.NotFound, st.Code(), "Error must be NotFound")
	})

	t.Run("ETCMeisaiService_ImportCSV_Contract", func(t *testing.T) {
		// Contract: ImportCSV must process CSV data and return import session
		req := &pb.ImportCSVRequest{
			AccountType: "corporate",
			AccountId:   "test-account-001",
			FileName:    "test.csv",
			FileContent: []byte("date,time,entrance_ic,exit_ic,amount,car_number,etc_card_number\n2024-01-05,14:00,テストIC,テスト出口IC,2000,品川123あ1111,5678901234567890"),
		}

		resp, err := client.ImportCSV(ctx, req)

		// Contract assertions
		assert.NoError(t, err, "ImportCSV must succeed for valid input")
		assert.NotNil(t, resp, "ImportCSV must return response")
		assert.NotNil(t, resp.Session, "ImportCSV must return session")
		assert.NotEmpty(t, resp.Session.Id, "Session must have non-empty ID")
		assert.Equal(t, req.AccountType, resp.Session.AccountType, "Session must preserve account type")
		assert.Equal(t, req.AccountId, resp.Session.AccountId, "Session must preserve account ID")
		assert.Equal(t, req.FileName, resp.Session.FileName, "Session must preserve file name")
		assert.Greater(t, resp.Session.FileSize, int64(0), "Session must have positive file size")
		assert.NotEqual(t, pb.ImportStatus_IMPORT_STATUS_UNSPECIFIED, resp.Session.Status, "Session must have valid status")
	})
}

func TestGRPCServiceContract_MappingOperations(t *testing.T) {
	client := setupGRPCTestClient(t)
	ctx := context.Background()

	t.Run("ETCMeisaiService_CreateMapping_Contract", func(t *testing.T) {
		// Contract: CreateMapping must create mapping with proper validation
		req := &pb.CreateMappingRequest{
			Mapping: &pb.ETCMapping{
				EtcRecordId:      1,
				MappingType:      "automatic",
				MappedEntityId:   12345,
				MappedEntityType: "dtako_record",
				Confidence:       0.95,
				Status:           pb.MappingStatus_MAPPING_STATUS_ACTIVE,
				CreatedBy:        "test-user",
			},
		}

		resp, err := client.CreateMapping(ctx, req)

		// Contract assertions
		assert.NoError(t, err, "CreateMapping must succeed for valid input")
		assert.NotNil(t, resp, "CreateMapping must return response")
		assert.NotNil(t, resp.Mapping, "CreateMapping must return mapping")
		assert.NotZero(t, resp.Mapping.Id, "Created mapping must have non-zero ID")
		assert.Equal(t, req.Mapping.EtcRecordId, resp.Mapping.EtcRecordId, "ETC record ID must be preserved")
		assert.Equal(t, req.Mapping.MappingType, resp.Mapping.MappingType, "Mapping type must be preserved")
		assert.Equal(t, req.Mapping.Confidence, resp.Mapping.Confidence, "Confidence must be preserved")
		assert.NotNil(t, resp.Mapping.CreatedAt, "CreatedAt must be set")
	})

	t.Run("ETCMeisaiService_ListMappings_Contract", func(t *testing.T) {
		// Contract: ListMappings must return paginated mapping results
		req := &pb.ListMappingsRequest{
			Page:     1,
			PageSize: 10,
			Status:   mappingStatusPtr(pb.MappingStatus_MAPPING_STATUS_ACTIVE),
		}

		resp, err := client.ListMappings(ctx, req)

		// Contract assertions
		assert.NoError(t, err, "ListMappings must succeed")
		assert.NotNil(t, resp, "ListMappings must return response")
		assert.NotNil(t, resp.Mappings, "ListMappings must return mappings slice")
		assert.GreaterOrEqual(t, resp.TotalCount, int32(0), "TotalCount must be non-negative")
		assert.Equal(t, req.Page, resp.Page, "Response page must match request")
		assert.Equal(t, req.PageSize, resp.PageSize, "Response page size must match request")
	})
}

func TestGRPCServiceContract_ImportSessions(t *testing.T) {
	client := setupGRPCTestClient(t)
	ctx := context.Background()

	t.Run("ETCMeisaiService_GetImportSession_Contract", func(t *testing.T) {
		// Contract: GetImportSession must return session details for valid session ID
		// First create an import session
		importReq := &pb.ImportCSVRequest{
			AccountType: "corporate",
			AccountId:   "session-test-001",
			FileName:    "session-test.csv",
			FileContent: []byte("date,time,entrance_ic,exit_ic,amount,car_number,etc_card_number\n2024-01-06,15:00,セッションIC,セッション出口IC,1500,品川123あ2222,6789012345678901"),
		}

		importResp, err := client.ImportCSV(ctx, importReq)
		require.NoError(t, err)

		// Test the contract
		getReq := &pb.GetImportSessionRequest{
			SessionId: importResp.Session.Id,
		}

		getResp, err := client.GetImportSession(ctx, getReq)

		// Contract assertions
		assert.NoError(t, err, "GetImportSession must succeed for valid session ID")
		assert.NotNil(t, getResp, "GetImportSession must return response")
		assert.NotNil(t, getResp.Session, "GetImportSession must return session")
		assert.Equal(t, importResp.Session.Id, getResp.Session.Id, "Session ID must match")
		assert.Equal(t, importResp.Session.AccountType, getResp.Session.AccountType, "Account type must match")
		assert.Equal(t, importResp.Session.FileName, getResp.Session.FileName, "File name must match")
	})

	t.Run("ETCMeisaiService_ListImportSessions_Contract", func(t *testing.T) {
		// Contract: ListImportSessions must return paginated session results
		req := &pb.ListImportSessionsRequest{
			Page:        1,
			PageSize:    10,
			AccountType: stringPtr("corporate"),
		}

		resp, err := client.ListImportSessions(ctx, req)

		// Contract assertions
		assert.NoError(t, err, "ListImportSessions must succeed")
		assert.NotNil(t, resp, "ListImportSessions must return response")
		assert.NotNil(t, resp.Sessions, "ListImportSessions must return sessions slice")
		assert.GreaterOrEqual(t, resp.TotalCount, int32(0), "TotalCount must be non-negative")
		assert.Equal(t, req.Page, resp.Page, "Response page must match request")
		assert.Equal(t, req.PageSize, resp.PageSize, "Response page size must match request")
	})
}

func TestGRPCServiceContract_Statistics(t *testing.T) {
	client := setupGRPCTestClient(t)
	ctx := context.Background()

	t.Run("ETCMeisaiService_GetStatistics_Contract", func(t *testing.T) {
		// Contract: GetStatistics must return aggregated statistics
		req := &pb.GetStatisticsRequest{
			DateFrom:      stringPtr("2024-01-01"),
			DateTo:        stringPtr("2024-01-31"),
			CarNumber:     stringPtr(""),
			EtcCardNumber: stringPtr(""),
		}

		resp, err := client.GetStatistics(ctx, req)

		// Contract assertions
		assert.NoError(t, err, "GetStatistics must succeed")
		assert.NotNil(t, resp, "GetStatistics must return response")
		assert.GreaterOrEqual(t, resp.TotalRecords, int64(0), "TotalRecords must be non-negative")
		assert.GreaterOrEqual(t, resp.TotalAmount, int64(0), "TotalAmount must be non-negative")
		assert.GreaterOrEqual(t, resp.UniqueCars, int32(0), "UniqueCars must be non-negative")
		assert.GreaterOrEqual(t, resp.UniqueCards, int32(0), "UniqueCards must be non-negative")
		assert.NotNil(t, resp.DailyStats, "DailyStats must not be nil")
		assert.NotNil(t, resp.IcStats, "IcStats must not be nil")
	})
}

func TestGRPCServiceContract_ErrorHandling(t *testing.T) {
	client := setupGRPCTestClient(t)
	ctx := context.Background()

	t.Run("ETCMeisaiService_GetRecord_NotFound_Contract", func(t *testing.T) {
		// Contract: GetRecord must return NotFound error for non-existent ID
		req := &pb.GetRecordRequest{
			Id: 999999999, // Non-existent ID
		}

		_, err := client.GetRecord(ctx, req)

		// Contract assertions
		assert.Error(t, err, "GetRecord must return error for non-existent ID")

		st, ok := status.FromError(err)
		assert.True(t, ok, "Error must be gRPC status")
		assert.Equal(t, codes.NotFound, st.Code(), "Error must be NotFound")
	})

	t.Run("ETCMeisaiService_CreateRecord_InvalidInput_Contract", func(t *testing.T) {
		// Contract: CreateRecord must return InvalidArgument error for invalid input
		req := &pb.CreateRecordRequest{
			Record: &pb.ETCMeisaiRecord{
				Hash:      "", // Empty required field
				Date:      "invalid-date",
				Time:      "invalid-time",
				TollAmount: -100, // Invalid negative amount
			},
		}

		_, err := client.CreateRecord(ctx, req)

		// Contract assertions
		assert.Error(t, err, "CreateRecord must return error for invalid input")

		st, ok := status.FromError(err)
		assert.True(t, ok, "Error must be gRPC status")
		assert.Equal(t, codes.InvalidArgument, st.Code(), "Error must be InvalidArgument")
	})

	t.Run("ETCMeisaiService_UpdateRecord_NotFound_Contract", func(t *testing.T) {
		// Contract: UpdateRecord must return NotFound error for non-existent ID
		req := &pb.UpdateRecordRequest{
			Id: 999999999, // Non-existent ID
			Record: &pb.ETCMeisaiRecord{
				Id:         999999999,
				Hash:       "test-hash",
				Date:       "2024-01-01",
				Time:       "10:00:00",
				TollAmount: 1000,
			},
		}

		_, err := client.UpdateRecord(ctx, req)

		// Contract assertions
		assert.Error(t, err, "UpdateRecord must return error for non-existent ID")

		st, ok := status.FromError(err)
		assert.True(t, ok, "Error must be gRPC status")
		assert.Equal(t, codes.NotFound, st.Code(), "Error must be NotFound")
	})
}
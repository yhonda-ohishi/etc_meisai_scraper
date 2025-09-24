package contract

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/yhonda-ohishi/etc_meisai/src/pb"
)

// TestEndToEndWorkflow_T010D validates complete ETC data processing pipeline
// This ensures the entire workflow from CSV import to data mapping works correctly
func TestEndToEndWorkflow_T010D(t *testing.T) {
	client := setupE2ETestClient(t)
	ctx := context.Background()

	t.Run("Complete_CSV_Import_To_Mapping_Workflow", func(t *testing.T) {
		// Contract: Complete workflow from CSV import to final mapping should work seamlessly

		// Step 1: Import CSV data
		csvData := createSampleCSVData()
		importReq := &pb.ImportCSVRequest{
			AccountType: "corporate",
			AccountId:   "workflow-test-001",
			FileName:    "workflow_test.csv",
			FileContent: csvData,
		}

		importResp, err := client.ImportCSV(ctx, importReq)
		require.NoError(t, err, "CSV import must succeed")
		require.NotNil(t, importResp.Session, "Import session must be created")

		sessionId := importResp.Session.Id

		// Step 2: Monitor import progress
		var finalSession *pb.ImportSession
		timeout := time.Now().Add(30 * time.Second)
		for time.Now().Before(timeout) {
			getSessionReq := &pb.GetImportSessionRequest{
				SessionId: sessionId,
			}

			sessionResp, err := client.GetImportSession(ctx, getSessionReq)
			require.NoError(t, err, "Getting import session must succeed")

			finalSession = sessionResp.Session
			if finalSession.Status == pb.ImportStatus_IMPORT_STATUS_COMPLETED ||
				finalSession.Status == pb.ImportStatus_IMPORT_STATUS_FAILED {
				break
			}

			time.Sleep(1 * time.Second)
		}

		// Contract assertions for import completion
		require.NotNil(t, finalSession, "Final session must be available")
		assert.Equal(t, pb.ImportStatus_IMPORT_STATUS_COMPLETED, finalSession.Status, "Import must complete successfully")
		assert.Greater(t, finalSession.SuccessRows, int32(0), "At least one row must be imported successfully")
		assert.GreaterOrEqual(t, finalSession.ProcessedRows, finalSession.SuccessRows, "Processed rows must be >= success rows")

		// Step 3: Verify imported records exist
		listReq := &pb.ListRecordsRequest{
			Page:     1,
			PageSize: 10,
		}

		listResp, err := client.ListRecords(ctx, listReq)
		require.NoError(t, err, "Listing records must succeed")
		assert.Greater(t, len(listResp.Records), 0, "Imported records must be retrievable")

		// Find our imported records
		var importedRecords []*pb.ETCMeisaiRecord
		for _, record := range listResp.Records {
			if strings.Contains(record.EntranceIc, "ワークフロー") {
				importedRecords = append(importedRecords, record)
			}
		}
		assert.Greater(t, len(importedRecords), 0, "Our test records must be found")

		// Step 4: Create mappings for imported records
		for i, record := range importedRecords {
			if i >= 2 { // Limit to first 2 records for test efficiency
				break
			}

			mappingReq := &pb.CreateMappingRequest{
				Mapping: &pb.ETCMapping{
					EtcRecordId:      record.Id,
					MappingType:      "workflow_test",
					MappedEntityId:   int64(1000 + i),
					MappedEntityType: "dtako_record",
					Confidence:       0.95,
					Status:           pb.MappingStatus_MAPPING_STATUS_ACTIVE,
					CreatedBy:        "workflow-test",
				},
			}

			mappingResp, err := client.CreateMapping(ctx, mappingReq)
			require.NoError(t, err, "Mapping creation must succeed")
			assert.NotZero(t, mappingResp.Mapping.Id, "Mapping must have valid ID")
			assert.Equal(t, record.Id, mappingResp.Mapping.EtcRecordId, "Mapping must reference correct ETC record")
		}

		// Step 5: Verify mappings exist and are retrievable
		mappingListReq := &pb.ListMappingsRequest{
			Page:     1,
			PageSize: 10,
			Status:   mappingStatusPtr(pb.MappingStatus_MAPPING_STATUS_ACTIVE),
		}

		mappingListResp, err := client.ListMappings(ctx, mappingListReq)
		require.NoError(t, err, "Listing mappings must succeed")

		// Find our workflow test mappings
		var workflowMappings []*pb.ETCMapping
		for _, mapping := range mappingListResp.Mappings {
			if mapping.MappingType == "workflow_test" {
				workflowMappings = append(workflowMappings, mapping)
			}
		}
		assert.GreaterOrEqual(t, len(workflowMappings), 1, "Workflow test mappings must be found")

		// Step 6: Generate statistics to verify data integrity
		statsReq := &pb.GetStatisticsRequest{
			DateFrom: stringPtr("2024-01-01"),
			DateTo:   stringPtr("2024-12-31"),
		}

		statsResp, err := client.GetStatistics(ctx, statsReq)
		require.NoError(t, err, "Statistics generation must succeed")
		assert.GreaterOrEqual(t, statsResp.TotalRecords, int64(len(importedRecords)), "Statistics must include imported records")
		assert.Greater(t, statsResp.TotalAmount, int64(0), "Total amount must be positive")
	})

	t.Run("Streaming_Import_Workflow", func(t *testing.T) {
		// Contract: Streaming import workflow must handle large datasets efficiently

		// Note: This test would require setting up streaming gRPC client
		// For now, we'll test the contract expectations

		// The streaming workflow should:
		// 1. Accept chunked CSV data
		// 2. Process data incrementally
		// 3. Provide real-time progress updates
		// 4. Handle errors gracefully

		t.Skip("Streaming import workflow requires bidirectional streaming setup")
	})

	t.Run("Error_Handling_Workflow", func(t *testing.T) {
		// Contract: Error handling throughout the workflow must be robust

		// Step 1: Test import with malformed CSV
		malformedCSV := []byte("invalid,csv,format\nthis,is,not,valid,etc,data")
		importReq := &pb.ImportCSVRequest{
			AccountType: "corporate",
			AccountId:   "error-test-001",
			FileName:    "error_test.csv",
			FileContent: malformedCSV,
		}

		importResp, err := client.ImportCSV(ctx, importReq)
		// Import should succeed (create session) but processing should fail
		require.NoError(t, err, "Import request must succeed even with bad data")

		// Step 2: Monitor import and verify it fails appropriately
		sessionId := importResp.Session.Id
		timeout := time.Now().Add(30 * time.Second)
		var finalSession *pb.ImportSession

		for time.Now().Before(timeout) {
			getSessionReq := &pb.GetImportSessionRequest{
				SessionId: sessionId,
			}

			sessionResp, err := client.GetImportSession(ctx, getSessionReq)
			require.NoError(t, err, "Getting session must succeed")

			finalSession = sessionResp.Session
			if finalSession.Status == pb.ImportStatus_IMPORT_STATUS_FAILED ||
				finalSession.Status == pb.ImportStatus_IMPORT_STATUS_COMPLETED {
				break
			}

			time.Sleep(1 * time.Second)
		}

		// Contract assertions for error handling
		require.NotNil(t, finalSession, "Final session must be available")
		if finalSession.Status == pb.ImportStatus_IMPORT_STATUS_FAILED {
			assert.Greater(t, finalSession.ErrorRows, int32(0), "Error count must reflect failed rows")
			assert.Greater(t, len(finalSession.ErrorLog), 0, "Error log must contain error details")
		} else if finalSession.Status == pb.ImportStatus_IMPORT_STATUS_COMPLETED {
			// If completed, check error handling
			assert.GreaterOrEqual(t, finalSession.ErrorRows, int32(0), "Error rows must be tracked")
		}

		// Step 3: Test creating mapping with invalid ETC record ID
		invalidMappingReq := &pb.CreateMappingRequest{
			Mapping: &pb.ETCMapping{
				EtcRecordId:      999999999, // Non-existent ID
				MappingType:      "error_test",
				MappedEntityId:   1000,
				MappedEntityType: "dtako_record",
				Confidence:       0.95,
				Status:           pb.MappingStatus_MAPPING_STATUS_ACTIVE,
				CreatedBy:        "error-test",
			},
		}

		_, err = client.CreateMapping(ctx, invalidMappingReq)
		assert.Error(t, err, "Creating mapping with invalid ETC record ID must fail")

		st, ok := status.FromError(err)
		assert.True(t, ok, "Error must be gRPC status")
		assert.Equal(t, codes.NotFound, st.Code(), "Error must be NotFound for invalid reference")
	})

	t.Run("Data_Consistency_Workflow", func(t *testing.T) {
		// Contract: Data must remain consistent throughout the entire workflow

		// Step 1: Import known dataset
		csvData := createConsistencyTestCSVData()
		importReq := &pb.ImportCSVRequest{
			AccountType: "personal",
			AccountId:   "consistency-test-001",
			FileName:    "consistency_test.csv",
			FileContent: csvData,
		}

		importResp, err := client.ImportCSV(ctx, importReq)
		require.NoError(t, err, "Consistency test import must succeed")

		// Wait for completion
		sessionId := importResp.Session.Id
		timeout := time.Now().Add(30 * time.Second)
		for time.Now().Before(timeout) {
			getSessionReq := &pb.GetImportSessionRequest{
				SessionId: sessionId,
			}

			sessionResp, err := client.GetImportSession(ctx, getSessionReq)
			require.NoError(t, err, "Session retrieval must succeed")

			if sessionResp.Session.Status == pb.ImportStatus_IMPORT_STATUS_COMPLETED {
				break
			}

			time.Sleep(1 * time.Second)
		}

		// Step 2: Verify data integrity across operations
		listReq := &pb.ListRecordsRequest{
			Page:      1,
			PageSize:  50,
			CarNumber: stringPtr("品川123あ9999"), // Specific test car number
		}

		listResp, err := client.ListRecords(ctx, listReq)
		require.NoError(t, err, "Filtered listing must succeed")

		var testRecord *pb.ETCMeisaiRecord
		for _, record := range listResp.Records {
			if record.CarNumber == "品川123あ9999" && strings.Contains(record.EntranceIc, "整合性") {
				testRecord = record
				break
			}
		}
		require.NotNil(t, testRecord, "Test record must be found")

		// Step 3: Create mapping and verify consistency
		mappingReq := &pb.CreateMappingRequest{
			Mapping: &pb.ETCMapping{
				EtcRecordId:      testRecord.Id,
				MappingType:      "consistency_test",
				MappedEntityId:   5555,
				MappedEntityType: "dtako_record",
				Confidence:       0.98,
				Status:           pb.MappingStatus_MAPPING_STATUS_ACTIVE,
				CreatedBy:        "consistency-test",
			},
		}

		mappingResp, err := client.CreateMapping(ctx, mappingReq)
		require.NoError(t, err, "Consistency mapping must succeed")

		// Step 4: Retrieve mapping and verify data consistency
		getMappingReq := &pb.GetMappingRequest{
			Id: mappingResp.Mapping.Id,
		}

		getMappingResp, err := client.GetMapping(ctx, getMappingReq)
		require.NoError(t, err, "Mapping retrieval must succeed")

		// Contract assertions for data consistency
		assert.Equal(t, testRecord.Id, getMappingResp.Mapping.EtcRecordId, "Mapping must reference correct ETC record")
		assert.Equal(t, mappingReq.Mapping.MappedEntityId, getMappingResp.Mapping.MappedEntityId, "Entity ID must be consistent")
		assert.Equal(t, mappingReq.Mapping.Confidence, getMappingResp.Mapping.Confidence, "Confidence must be preserved")

		// Step 5: Update record and verify mapping still references correctly
		updateReq := &pb.UpdateRecordRequest{
			Id: testRecord.Id,
			Record: &pb.ETCMeisaiRecord{
				Id:             testRecord.Id,
				Hash:           testRecord.Hash,
				Date:           testRecord.Date,
				Time:           testRecord.Time,
				EntranceIc:     testRecord.EntranceIc,
				ExitIc:         testRecord.ExitIc,
				TollAmount:     testRecord.TollAmount + 100, // Modify amount
				CarNumber:      testRecord.CarNumber,
				EtcCardNumber:  testRecord.EtcCardNumber,
				EtcNum:         testRecord.EtcNum,
				DtakoRowId:     testRecord.DtakoRowId,
				CreatedAt:      testRecord.CreatedAt,
				UpdatedAt:      testRecord.UpdatedAt,
			},
		}

		updateResp, err := client.UpdateRecord(ctx, updateReq)
		require.NoError(t, err, "Record update must succeed")

		// Verify mapping still references the updated record correctly
		getMappingResp2, err := client.GetMapping(ctx, getMappingReq)
		require.NoError(t, err, "Mapping retrieval after update must succeed")
		assert.Equal(t, updateResp.Record.Id, getMappingResp2.Mapping.EtcRecordId, "Mapping reference must remain consistent after update")
	})

	t.Run("Concurrent_Workflow_Operations", func(t *testing.T) {
		// Contract: Concurrent operations in the workflow must not cause data corruption

		// This test simulates multiple concurrent workflow operations
		numConcurrentImports := 3
		done := make(chan error, numConcurrentImports)

		for i := 0; i < numConcurrentImports; i++ {
			go func(id int) {
				csvData := createConcurrentTestCSVData(id)
				importReq := &pb.ImportCSVRequest{
					AccountType: "corporate",
					AccountId:   "concurrent-test-" + string(rune('A'+id)),
					FileName:    "concurrent_test_" + string(rune('0'+id)) + ".csv",
					FileContent: csvData,
				}

				importResp, err := client.ImportCSV(ctx, importReq)
				if err != nil {
					done <- err
					return
				}

				// Wait for completion
				sessionId := importResp.Session.Id
				timeout := time.Now().Add(60 * time.Second)
				for time.Now().Before(timeout) {
					getSessionReq := &pb.GetImportSessionRequest{
						SessionId: sessionId,
					}

					sessionResp, err := client.GetImportSession(ctx, getSessionReq)
					if err != nil {
						done <- err
						return
					}

					if sessionResp.Session.Status == pb.ImportStatus_IMPORT_STATUS_COMPLETED ||
						sessionResp.Session.Status == pb.ImportStatus_IMPORT_STATUS_FAILED {
						break
					}

					time.Sleep(500 * time.Millisecond)
				}

				done <- nil
			}(i)
		}

		// Wait for all concurrent operations to complete
		for i := 0; i < numConcurrentImports; i++ {
			err := <-done
			assert.NoError(t, err, "Concurrent import must succeed without corruption")
		}

		// Verify data integrity after concurrent operations
		statsReq := &pb.GetStatisticsRequest{
			DateFrom: stringPtr("2024-01-01"),
			DateTo:   stringPtr("2024-12-31"),
		}

		statsResp, err := client.GetStatistics(ctx, statsReq)
		require.NoError(t, err, "Statistics after concurrent operations must succeed")
		assert.Greater(t, statsResp.TotalRecords, int64(0), "Records must exist after concurrent imports")
	})
}

// Helper functions to create test data

func createSampleCSVData() []byte {
	csvContent := `date,time,entrance_ic,exit_ic,toll_amount,car_number,etc_card_number,etc_num
2024-01-01,10:00:00,ワークフローテスト入口IC,ワークフローテスト出口IC,1500,品川123あ1111,1111222233334444,WORKFLOW001
2024-01-01,11:00:00,ワークフローテスト2入口IC,ワークフローテスト2出口IC,2000,品川123あ2222,2222333344445555,WORKFLOW002
2024-01-01,12:00:00,ワークフローテスト3入口IC,ワークフローテスト3出口IC,1800,品川123あ3333,3333444455556666,WORKFLOW003`

	return []byte(csvContent)
}

func createConsistencyTestCSVData() []byte {
	csvContent := `date,time,entrance_ic,exit_ic,toll_amount,car_number,etc_card_number,etc_num
2024-01-02,13:00:00,整合性テスト入口IC,整合性テスト出口IC,2500,品川123あ9999,9999000011112222,CONSISTENCY001`

	return []byte(csvContent)
}

func createConcurrentTestCSVData(id int) []byte {
	etcNums := []string{"CONCURRENT001", "CONCURRENT002", "CONCURRENT003"}

	csvContent := "date,time,entrance_ic,exit_ic,toll_amount,car_number,etc_card_number,etc_num\n"
	csvContent += "2024-01-03,14:00:00,並行テスト" + string(rune('1'+id)) + "入口IC,並行テスト" + string(rune('1'+id)) + "出口IC,1000,品川123あ100" + string(rune('1'+id)) + ",100" + string(rune('1'+id)) + "200230034004," + etcNums[id%len(etcNums)]

	return []byte(csvContent)
}

// Helper functions
func stringPtr(s string) *string {
	return &s
}

func mappingStatusPtr(status pb.MappingStatus) *pb.MappingStatus {
	return &status
}

// setupE2ETestClient creates a gRPC client for end-to-end testing
func setupE2ETestClient(t *testing.T) pb.ETCMeisaiServiceClient {
	// This is a placeholder - in a real implementation, this would:
	// 1. Start a complete test environment with database
	// 2. Create a gRPC client connection
	// 3. Return the client for end-to-end testing

	// For now, we'll skip if no test environment is available
	t.Skip("End-to-end test environment setup required")
	return nil
}
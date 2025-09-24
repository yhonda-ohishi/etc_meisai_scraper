package contract

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	pb "github.com/yhonda-ohishi/etc_meisai/src/pb"
)

// TestPerformanceSLA_T010E validates performance contract testing with SLA validation
// This ensures that all gRPC operations meet defined Service Level Agreements
func TestPerformanceSLA_T010E(t *testing.T) {
	client := setupPerformanceTestClient(t)
	ctx := context.Background()

	// SLA Constants - These define the performance contracts
	const (
		// Response time SLAs (as specified in requirements)
		MaxResponseTime100ms = 100 * time.Millisecond
		MaxResponseTime200ms = 200 * time.Millisecond
		MaxResponseTime500ms = 500 * time.Millisecond
		MaxResponseTime1s    = 1 * time.Second

		// Throughput SLAs
		MinThroughputCreateOps = 50  // operations per second
		MinThroughputReadOps   = 100 // operations per second

		// Concurrency SLAs
		MaxConcurrentUsers = 10
		MaxConcurrentOps   = 50
	)

	t.Run("Single_Operation_Response_Time_SLA", func(t *testing.T) {
		// Contract: Individual operations must meet response time SLAs

		t.Run("CreateRecord_SLA_100ms", func(t *testing.T) {
			// Contract: CreateRecord must complete within 100ms
			req := &pb.CreateRecordRequest{
				Record: &pb.ETCMeisaiRecord{
					Hash:           "sla-test-create-001",
					Date:           "2024-01-01",
					Time:           "10:00:00",
					EntranceIc:     "SLAテスト入口IC",
					ExitIc:         "SLAテスト出口IC",
					TollAmount:     1500,
					CarNumber:      "品川123あ1234",
					EtcCardNumber:  "1234567890123456",
					EtcNum:         stringPtr("SLA001"),
				},
			}

			start := time.Now()
			resp, err := client.CreateRecord(ctx, req)
			duration := time.Since(start)

			// SLA assertions
			assert.NoError(t, err, "CreateRecord must succeed for SLA test")
			assert.NotNil(t, resp, "CreateRecord must return response")
			assert.Less(t, duration, MaxResponseTime100ms, "CreateRecord must complete within 100ms SLA")

			// Log performance metrics
			t.Logf("CreateRecord performance: %v (SLA: %v)", duration, MaxResponseTime100ms)
		})

		t.Run("GetRecord_SLA_100ms", func(t *testing.T) {
			// Contract: GetRecord must complete within 100ms

			// First create a record to get
			createReq := &pb.CreateRecordRequest{
				Record: &pb.ETCMeisaiRecord{
					Hash:           "sla-test-get-001",
					Date:           "2024-01-01",
					Time:           "11:00:00",
					EntranceIc:     "SLA取得テスト入口IC",
					ExitIc:         "SLA取得テスト出口IC",
					TollAmount:     2000,
					CarNumber:      "品川123あ5678",
					EtcCardNumber:  "5678901234567890",
				},
			}

			createResp, err := client.CreateRecord(ctx, createReq)
			require.NoError(t, err, "Setup record creation must succeed")

			// Test the SLA
			getReq := &pb.GetRecordRequest{
				Id: createResp.Record.Id,
			}

			start := time.Now()
			resp, err := client.GetRecord(ctx, getReq)
			duration := time.Since(start)

			// SLA assertions
			assert.NoError(t, err, "GetRecord must succeed for SLA test")
			assert.NotNil(t, resp, "GetRecord must return response")
			assert.Less(t, duration, MaxResponseTime100ms, "GetRecord must complete within 100ms SLA")

			t.Logf("GetRecord performance: %v (SLA: %v)", duration, MaxResponseTime100ms)
		})

		t.Run("ListRecords_SLA_200ms", func(t *testing.T) {
			// Contract: ListRecords must complete within 200ms for standard page size
			req := &pb.ListRecordsRequest{
				Page:     1,
				PageSize: 20, // Standard page size
				SortBy:   "created_at",
				SortOrder: pb.SortOrder_SORT_ORDER_DESC,
			}

			start := time.Now()
			resp, err := client.ListRecords(ctx, req)
			duration := time.Since(start)

			// SLA assertions
			assert.NoError(t, err, "ListRecords must succeed for SLA test")
			assert.NotNil(t, resp, "ListRecords must return response")
			assert.Less(t, duration, MaxResponseTime200ms, "ListRecords must complete within 200ms SLA")

			t.Logf("ListRecords performance: %v (SLA: %v)", duration, MaxResponseTime200ms)
		})

		t.Run("UpdateRecord_SLA_100ms", func(t *testing.T) {
			// Contract: UpdateRecord must complete within 100ms

			// Create record to update
			createReq := &pb.CreateRecordRequest{
				Record: &pb.ETCMeisaiRecord{
					Hash:           "sla-test-update-001",
					Date:           "2024-01-01",
					Time:           "12:00:00",
					EntranceIc:     "SLA更新テスト入口IC",
					ExitIc:         "SLA更新テスト出口IC",
					TollAmount:     2500,
					CarNumber:      "品川123あ9999",
					EtcCardNumber:  "9999000011112222",
				},
			}

			createResp, err := client.CreateRecord(ctx, createReq)
			require.NoError(t, err, "Setup record creation must succeed")

			// Test update SLA
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
					CreatedAt:      createResp.Record.CreatedAt,
				},
			}

			start := time.Now()
			resp, err := client.UpdateRecord(ctx, updateReq)
			duration := time.Since(start)

			// SLA assertions
			assert.NoError(t, err, "UpdateRecord must succeed for SLA test")
			assert.NotNil(t, resp, "UpdateRecord must return response")
			assert.Less(t, duration, MaxResponseTime100ms, "UpdateRecord must complete within 100ms SLA")

			t.Logf("UpdateRecord performance: %v (SLA: %v)", duration, MaxResponseTime100ms)
		})

		t.Run("ImportCSV_SLA_1s", func(t *testing.T) {
			// Contract: ImportCSV must complete within 1s for small files (<1MB)
			csvData := createSmallCSVData(50) // 50 records, ~3KB

			req := &pb.ImportCSVRequest{
				AccountType: "corporate",
				AccountId:   "sla-test-import",
				FileName:    "sla_test.csv",
				FileContent: csvData,
			}

			start := time.Now()
			resp, err := client.ImportCSV(ctx, req)
			duration := time.Since(start)

			// SLA assertions
			assert.NoError(t, err, "ImportCSV must succeed for SLA test")
			assert.NotNil(t, resp, "ImportCSV must return response")
			assert.Less(t, duration, MaxResponseTime1s, "ImportCSV must complete within 1s SLA for small files")

			t.Logf("ImportCSV performance: %v (SLA: %v)", duration, MaxResponseTime1s)
		})

		t.Run("GetStatistics_SLA_500ms", func(t *testing.T) {
			// Contract: GetStatistics must complete within 500ms
			req := &pb.GetStatisticsRequest{
				DateFrom: stringPtr("2024-01-01"),
				DateTo:   stringPtr("2024-01-31"),
			}

			start := time.Now()
			resp, err := client.GetStatistics(ctx, req)
			duration := time.Since(start)

			// SLA assertions
			assert.NoError(t, err, "GetStatistics must succeed for SLA test")
			assert.NotNil(t, resp, "GetStatistics must return response")
			assert.Less(t, duration, MaxResponseTime500ms, "GetStatistics must complete within 500ms SLA")

			t.Logf("GetStatistics performance: %v (SLA: %v)", duration, MaxResponseTime500ms)
		})
	})

	t.Run("Throughput_SLA_Testing", func(t *testing.T) {
		// Contract: System must handle minimum throughput requirements

		t.Run("CreateRecord_Throughput_50ops", func(t *testing.T) {
			// Contract: System must handle at least 50 CreateRecord operations per second
			const numOps = 100
			const maxDuration = 2 * time.Second // 100 ops in 2s = 50 ops/sec

			var wg sync.WaitGroup
			errors := make(chan error, numOps)

			start := time.Now()

			for i := 0; i < numOps; i++ {
				wg.Add(1)
				go func(id int) {
					defer wg.Done()

					req := &pb.CreateRecordRequest{
						Record: &pb.ETCMeisaiRecord{
							Hash:           "throughput-test-" + string(rune('0'+id%10)),
							Date:           "2024-01-01",
							Time:           "10:00:00",
							EntranceIc:     "スループットテスト入口IC",
							ExitIc:         "スループットテスト出口IC",
							TollAmount:     int32(1000 + id),
							CarNumber:      "品川123あ0000",
							EtcCardNumber:  "0000111122223333",
						},
					}

					_, err := client.CreateRecord(ctx, req)
					if err != nil {
						errors <- err
						return
					}
					errors <- nil
				}(i)
			}

			wg.Wait()
			duration := time.Since(start)
			close(errors)

			// Count errors
			errorCount := 0
			for err := range errors {
				if err != nil {
					errorCount++
				}
			}

			// SLA assertions
			assert.Less(t, duration, maxDuration, "Throughput SLA: 100 CreateRecord ops must complete within 2s")
			assert.Less(t, errorCount, numOps/10, "Error rate must be less than 10% under load")

			actualThroughput := float64(numOps) / duration.Seconds()
			assert.GreaterOrEqual(t, actualThroughput, float64(MinThroughputCreateOps), "CreateRecord throughput must meet SLA")

			t.Logf("CreateRecord throughput: %.2f ops/sec (SLA: %d ops/sec)", actualThroughput, MinThroughputCreateOps)
		})

		t.Run("GetRecord_Throughput_100ops", func(t *testing.T) {
			// Contract: System must handle at least 100 GetRecord operations per second

			// First create some records to get
			recordIds := make([]int64, 10)
			for i := 0; i < 10; i++ {
				createReq := &pb.CreateRecordRequest{
					Record: &pb.ETCMeisaiRecord{
						Hash:           "read-throughput-" + string(rune('0'+i)),
						Date:           "2024-01-01",
						Time:           "11:00:00",
						EntranceIc:     "読み取りスループットテスト入口IC",
						ExitIc:         "読み取りスループットテスト出口IC",
						TollAmount:     int32(2000 + i),
						CarNumber:      "品川123あ1111",
						EtcCardNumber:  "1111222233334444",
					},
				}

				resp, err := client.CreateRecord(ctx, createReq)
				require.NoError(t, err, "Setup record creation must succeed")
				recordIds[i] = resp.Record.Id
			}

			// Test read throughput
			const numOps = 200
			const maxDuration = 2 * time.Second // 200 ops in 2s = 100 ops/sec

			var wg sync.WaitGroup
			errors := make(chan error, numOps)

			start := time.Now()

			for i := 0; i < numOps; i++ {
				wg.Add(1)
				go func(id int) {
					defer wg.Done()

					req := &pb.GetRecordRequest{
						Id: recordIds[id%len(recordIds)], // Cycle through available records
					}

					_, err := client.GetRecord(ctx, req)
					errors <- err
				}(i)
			}

			wg.Wait()
			duration := time.Since(start)
			close(errors)

			// Count errors
			errorCount := 0
			for err := range errors {
				if err != nil {
					errorCount++
				}
			}

			// SLA assertions
			assert.Less(t, duration, maxDuration, "Throughput SLA: 200 GetRecord ops must complete within 2s")
			assert.Equal(t, 0, errorCount, "No errors should occur for valid record IDs")

			actualThroughput := float64(numOps) / duration.Seconds()
			assert.GreaterOrEqual(t, actualThroughput, float64(MinThroughputReadOps), "GetRecord throughput must meet SLA")

			t.Logf("GetRecord throughput: %.2f ops/sec (SLA: %d ops/sec)", actualThroughput, MinThroughputReadOps)
		})
	})

	t.Run("Concurrency_SLA_Testing", func(t *testing.T) {
		// Contract: System must handle concurrent operations within SLA

		t.Run("Concurrent_Users_SLA", func(t *testing.T) {
			// Contract: System must handle 10 concurrent users without performance degradation
			const numUsers = MaxConcurrentUsers
			const opsPerUser = 5

			var wg sync.WaitGroup
			userResults := make(chan time.Duration, numUsers)

			start := time.Now()

			for userId := 0; userId < numUsers; userId++ {
				wg.Add(1)
				go func(uid int) {
					defer wg.Done()

					userStart := time.Now()

					// Each user performs a typical workflow
					for opId := 0; opId < opsPerUser; opId++ {
						// Create record
						createReq := &pb.CreateRecordRequest{
							Record: &pb.ETCMeisaiRecord{
								Hash:           "concurrent-user-" + string(rune('A'+uid)) + "-" + string(rune('0'+opId)),
								Date:           "2024-01-01",
								Time:           "12:00:00",
								EntranceIc:     "並行ユーザーテスト入口IC",
								ExitIc:         "並行ユーザーテスト出口IC",
								TollAmount:     int32(1500 + uid*100 + opId*10),
								CarNumber:      "品川123あ2222",
								EtcCardNumber:  "2222333344445555",
							},
						}

						createResp, err := client.CreateRecord(ctx, createReq)
						if err != nil {
							t.Errorf("User %d operation %d failed: %v", uid, opId, err)
							return
						}

						// Get the created record
						getReq := &pb.GetRecordRequest{Id: createResp.Record.Id}
						_, err = client.GetRecord(ctx, getReq)
						if err != nil {
							t.Errorf("User %d get operation %d failed: %v", uid, opId, err)
							return
						}
					}

					userDuration := time.Since(userStart)
					userResults <- userDuration
				}(userId)
			}

			wg.Wait()
			totalDuration := time.Since(start)
			close(userResults)

			// Analyze user performance
			var maxUserDuration time.Duration
			var totalUserTime time.Duration
			userCount := 0

			for userDuration := range userResults {
				if userDuration > maxUserDuration {
					maxUserDuration = userDuration
				}
				totalUserTime += userDuration
				userCount++
			}

			avgUserDuration := totalUserTime / time.Duration(userCount)

			// SLA assertions
			assert.Equal(t, numUsers, userCount, "All users must complete successfully")
			assert.Less(t, maxUserDuration, 5*time.Second, "No user should take more than 5s for their operations")
			assert.Less(t, avgUserDuration, 3*time.Second, "Average user time should be under 3s")

			t.Logf("Concurrent users: %d, Total time: %v, Max user time: %v, Avg user time: %v",
				numUsers, totalDuration, maxUserDuration, avgUserDuration)
		})

		t.Run("Mixed_Operations_Concurrency_SLA", func(t *testing.T) {
			// Contract: System must handle mixed concurrent operations efficiently
			const numOperations = MaxConcurrentOps
			const maxTotalDuration = 10 * time.Second

			var wg sync.WaitGroup
			operationResults := make(chan OperationResult, numOperations)

			start := time.Now()

			for i := 0; i < numOperations; i++ {
				wg.Add(1)
				go func(opId int) {
					defer wg.Done()

					opStart := time.Now()
					var err error
					var opType string

					switch opId % 4 {
					case 0: // Create operation
						opType = "create"
						req := &pb.CreateRecordRequest{
							Record: &pb.ETCMeisaiRecord{
								Hash:           "mixed-op-" + string(rune('0'+opId%10)),
								Date:           "2024-01-01",
								Time:           "13:00:00",
								EntranceIc:     "混合オペレーションテスト入口IC",
								ExitIc:         "混合オペレーションテスト出口IC",
								TollAmount:     int32(1800 + opId),
								CarNumber:      "品川123あ3333",
								EtcCardNumber:  "3333444455556666",
							},
						}
						_, err = client.CreateRecord(ctx, req)

					case 1: // List operation
						opType = "list"
						req := &pb.ListRecordsRequest{
							Page:     1,
							PageSize: 10,
						}
						_, err = client.ListRecords(ctx, req)

					case 2: // Statistics operation
						opType = "stats"
						req := &pb.GetStatisticsRequest{
							DateFrom: stringPtr("2024-01-01"),
							DateTo:   stringPtr("2024-01-31"),
						}
						_, err = client.GetStatistics(ctx, req)

					case 3: // Mapping list operation
						opType = "mapping_list"
						req := &pb.ListMappingsRequest{
							Page:     1,
							PageSize: 10,
						}
						_, err = client.ListMappings(ctx, req)
					}

					duration := time.Since(opStart)
					operationResults <- OperationResult{
						Type:     opType,
						Duration: duration,
						Error:    err,
					}
				}(i)
			}

			wg.Wait()
			totalDuration := time.Since(start)
			close(operationResults)

			// Analyze operation performance
			operationStats := make(map[string][]time.Duration)
			errorCount := 0

			for result := range operationResults {
				if result.Error != nil {
					errorCount++
					continue
				}
				operationStats[result.Type] = append(operationStats[result.Type], result.Duration)
			}

			// SLA assertions
			assert.Less(t, totalDuration, maxTotalDuration, "Mixed operations must complete within SLA")
			assert.Less(t, errorCount, numOperations/20, "Error rate must be less than 5% under concurrent load")

			// Check individual operation type performance
			for opType, durations := range operationStats {
				if len(durations) == 0 {
					continue
				}

				var total time.Duration
				var max time.Duration
				for _, d := range durations {
					total += d
					if d > max {
						max = d
					}
				}
				avg := total / time.Duration(len(durations))

				// Individual operation SLAs under concurrency
				switch opType {
				case "create":
					assert.Less(t, max, 500*time.Millisecond, "Create operations under concurrency must be < 500ms")
				case "list":
					assert.Less(t, max, 1*time.Second, "List operations under concurrency must be < 1s")
				case "stats":
					assert.Less(t, max, 2*time.Second, "Stats operations under concurrency must be < 2s")
				case "mapping_list":
					assert.Less(t, max, 1*time.Second, "Mapping list operations under concurrency must be < 1s")
				}

				t.Logf("Operation %s: count=%d, avg=%v, max=%v", opType, len(durations), avg, max)
			}
		})
	})

	t.Run("Resource_Usage_SLA", func(t *testing.T) {
		// Contract: Operations must complete within resource constraints

		t.Run("Large_Dataset_Performance", func(t *testing.T) {
			// Contract: System must handle large datasets within SLA
			req := &pb.ListRecordsRequest{
				Page:     1,
				PageSize: 1000, // Large page size
				SortBy:   "created_at",
				SortOrder: pb.SortOrder_SORT_ORDER_DESC,
			}

			start := time.Now()
			resp, err := client.ListRecords(ctx, req)
			duration := time.Since(start)

			// SLA assertions for large datasets
			assert.NoError(t, err, "Large dataset query must succeed")
			assert.NotNil(t, resp, "Large dataset query must return response")
			assert.Less(t, duration, 5*time.Second, "Large dataset query must complete within 5s")

			t.Logf("Large dataset query (1000 records): %v", duration)
		})

		t.Run("Complex_Filter_Performance", func(t *testing.T) {
			// Contract: Complex filtered queries must complete within SLA
			req := &pb.ListRecordsRequest{
				Page:          1,
				PageSize:      100,
				DateFrom:      stringPtr("2024-01-01"),
				DateTo:        stringPtr("2024-12-31"),
				CarNumber:     stringPtr("品川123あ"),    // Partial match
				EtcCardNumber: stringPtr("1111"),       // Partial match
				EntranceIc:    stringPtr("テスト"),       // Partial match
				ExitIc:        stringPtr("出口"),        // Partial match
			}

			start := time.Now()
			resp, err := client.ListRecords(ctx, req)
			duration := time.Since(start)

			// SLA assertions for complex queries
			assert.NoError(t, err, "Complex filtered query must succeed")
			assert.NotNil(t, resp, "Complex filtered query must return response")
			assert.Less(t, duration, 2*time.Second, "Complex filtered query must complete within 2s")

			t.Logf("Complex filtered query: %v", duration)
		})
	})
}

// OperationResult holds the result of a performance test operation
type OperationResult struct {
	Type     string
	Duration time.Duration
	Error    error
}

// Helper function to create small CSV data for performance testing
func createSmallCSVData(numRecords int) []byte {
	csvContent := "date,time,entrance_ic,exit_ic,toll_amount,car_number,etc_card_number,etc_num\n"

	for i := 0; i < numRecords; i++ {
		csvContent += "2024-01-01,10:00:00,SLAテスト入口IC,SLAテスト出口IC,1500,品川123あ0001,0001111122223333,SLA001\n"
	}

	return []byte(csvContent)
}


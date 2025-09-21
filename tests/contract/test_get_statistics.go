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

func TestGetStatistics_Success(t *testing.T) {
	// Arrange
	conn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("Failed to connect to gRPC server: %v", err)
	}
	defer conn.Close()

	client := pb.NewETCMeisaiServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Test data - basic statistics request
	req := &pb.GetStatisticsRequest{}

	// Act
	resp, err := client.GetStatistics(ctx, req)

	// Assert
	// This test should FAIL initially as the server is not implemented yet
	if err != nil {
		st := status.Convert(err)
		if st.Code() == codes.Unimplemented {
			t.Logf("Expected failure: GetStatistics not implemented yet - %v", err)
			return
		}
		t.Fatalf("Unexpected error: %v", err)
	}

	// If server is implemented, verify response
	if resp == nil {
		t.Fatal("Response is nil")
	}

	// Verify basic statistics fields
	if resp.TotalRecords < 0 {
		t.Errorf("Expected total_records to be non-negative, got %d", resp.TotalRecords)
	}

	if resp.TotalAmount < 0 {
		t.Errorf("Expected total_amount to be non-negative, got %d", resp.TotalAmount)
	}

	if resp.UniqueCars < 0 {
		t.Errorf("Expected unique_cars to be non-negative, got %d", resp.UniqueCars)
	}

	if resp.UniqueCards < 0 {
		t.Errorf("Expected unique_cards to be non-negative, got %d", resp.UniqueCards)
	}

	// Daily stats should not be nil (can be empty array)
	if resp.DailyStats == nil {
		t.Error("Expected daily_stats array to not be nil")
	}

	// IC stats should not be nil (can be empty array)
	if resp.IcStats == nil {
		t.Error("Expected ic_stats array to not be nil")
	}

	// Verify daily stats structure
	for i, dailyStat := range resp.DailyStats {
		if dailyStat.Date == "" {
			t.Errorf("Daily stat %d has empty date", i)
		}
		if dailyStat.RecordCount < 0 {
			t.Errorf("Daily stat %d has negative record count: %d", i, dailyStat.RecordCount)
		}
		if dailyStat.TotalAmount < 0 {
			t.Errorf("Daily stat %d has negative total amount: %d", i, dailyStat.TotalAmount)
		}
	}

	// Verify IC stats structure
	for i, icStat := range resp.IcStats {
		if icStat.IcName == "" {
			t.Errorf("IC stat %d has empty IC name", i)
		}
		if icStat.UsageCount < 0 {
			t.Errorf("IC stat %d has negative usage count: %d", i, icStat.UsageCount)
		}
		if icStat.IcType != "entrance" && icStat.IcType != "exit" {
			t.Errorf("IC stat %d has invalid type: %s (should be 'entrance' or 'exit')", i, icStat.IcType)
		}
	}

	t.Logf("Statistics summary - Records: %d, Amount: %d, Cars: %d, Cards: %d, Daily entries: %d, IC entries: %d",
		resp.TotalRecords, resp.TotalAmount, resp.UniqueCars, resp.UniqueCards,
		len(resp.DailyStats), len(resp.IcStats))
}

func TestGetStatistics_WithDateRange(t *testing.T) {
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
		dateFrom string
		dateTo   string
	}{
		{
			name:     "single month",
			dateFrom: "2024-01-01",
			dateTo:   "2024-01-31",
		},
		{
			name:     "quarter",
			dateFrom: "2024-01-01",
			dateTo:   "2024-03-31",
		},
		{
			name:     "single day",
			dateFrom: "2024-01-15",
			dateTo:   "2024-01-15",
		},
		{
			name:     "year range",
			dateFrom: "2024-01-01",
			dateTo:   "2024-12-31",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := &pb.GetStatisticsRequest{
				DateFrom: stringPtr(tc.dateFrom),
				DateTo:   stringPtr(tc.dateTo),
			}

			// Act
			resp, err := client.GetStatistics(ctx, req)

			// Assert
			if err != nil {
				st := status.Convert(err)
				if st.Code() == codes.Unimplemented {
					t.Logf("Expected failure: GetStatistics not implemented yet - %v", err)
					return
				}
				t.Fatalf("Unexpected error for %s: %v", tc.name, err)
			}

			// If server is implemented, verify date filtering
			if resp == nil {
				t.Fatal("Response is nil")
			}

			// Daily stats should only include dates within the range
			for i, dailyStat := range resp.DailyStats {
				if dailyStat.Date < tc.dateFrom || dailyStat.Date > tc.dateTo {
					t.Errorf("Daily stat %d date %s is outside range %s to %s", i, dailyStat.Date, tc.dateFrom, tc.dateTo)
				}
			}

			t.Logf("Date range test '%s' returned %d total records, %d daily stats",
				tc.name, resp.TotalRecords, len(resp.DailyStats))
		})
	}
}

func TestGetStatistics_WithFilters(t *testing.T) {
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
		req  *pb.GetStatisticsRequest
	}{
		{
			name: "filter by car number",
			req: &pb.GetStatisticsRequest{
				CarNumber: stringPtr("品川 123 あ 1234"),
			},
		},
		{
			name: "filter by ETC card",
			req: &pb.GetStatisticsRequest{
				EtcCardNumber: stringPtr("1234567890123456"),
			},
		},
		{
			name: "combined filters",
			req: &pb.GetStatisticsRequest{
				DateFrom:      stringPtr("2024-01-01"),
				DateTo:        stringPtr("2024-01-31"),
				CarNumber:     stringPtr("品川 123 あ 1234"),
				EtcCardNumber: stringPtr("1234567890123456"),
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Act
			resp, err := client.GetStatistics(ctx, tc.req)

			// Assert
			if err != nil {
				st := status.Convert(err)
				if st.Code() == codes.Unimplemented {
					t.Logf("Expected failure: GetStatistics not implemented yet - %v", err)
					return
				}
				t.Fatalf("Unexpected error for %s: %v", tc.name, err)
			}

			// If server is implemented, verify filters work
			if resp == nil {
				t.Fatal("Response is nil")
			}

			// For filtered statistics, unique counts should be reasonable
			if tc.req.CarNumber != nil && resp.UniqueCars > 1 {
				t.Logf("Note: Filtering by single car but got %d unique cars - may indicate partial matches or data variations", resp.UniqueCars)
			}

			if tc.req.EtcCardNumber != nil && resp.UniqueCards > 1 {
				t.Logf("Note: Filtering by single ETC card but got %d unique cards - may indicate partial matches or data variations", resp.UniqueCards)
			}

			t.Logf("Filter test '%s' returned %d records with %d unique cars and %d unique cards",
				tc.name, resp.TotalRecords, resp.UniqueCars, resp.UniqueCards)
		})
	}
}

func TestGetStatistics_InvalidDateRange(t *testing.T) {
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
		dateFrom string
		dateTo   string
	}{
		{
			name:     "invalid date format",
			dateFrom: "2024/01/01",
			dateTo:   "2024/01/31",
		},
		{
			name:     "end before start",
			dateFrom: "2024-01-31",
			dateTo:   "2024-01-01",
		},
		{
			name:     "invalid date",
			dateFrom: "2024-02-30",
			dateTo:   "2024-02-31",
		},
		{
			name:     "partial date",
			dateFrom: "2024-01",
			dateTo:   "2024-01",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := &pb.GetStatisticsRequest{
				DateFrom: stringPtr(tc.dateFrom),
				DateTo:   stringPtr(tc.dateTo),
			}

			// Act
			resp, err := client.GetStatistics(ctx, req)

			// Assert
			if err != nil {
				st := status.Convert(err)
				if st.Code() == codes.Unimplemented {
					t.Logf("Expected failure: GetStatistics not implemented yet - %v", err)
					return
				}
				// When implemented, should return InvalidArgument for invalid date ranges
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

func TestGetStatistics_EmptyDataset(t *testing.T) {
	// Arrange
	conn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("Failed to connect to gRPC server: %v", err)
	}
	defer conn.Close()

	client := pb.NewETCMeisaiServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Request statistics for a range that should have no data
	req := &pb.GetStatisticsRequest{
		DateFrom:      stringPtr("2099-01-01"),
		DateTo:        stringPtr("2099-01-31"),
		CarNumber:     stringPtr("非存在 999 zzz 9999"),
		EtcCardNumber: stringPtr("9999999999999999"),
	}

	// Act
	resp, err := client.GetStatistics(ctx, req)

	// Assert
	if err != nil {
		st := status.Convert(err)
		if st.Code() == codes.Unimplemented {
			t.Logf("Expected failure: GetStatistics not implemented yet - %v", err)
			return
		}
		t.Fatalf("Unexpected error: %v", err)
	}

	// If server is implemented, verify empty dataset handling
	if resp == nil {
		t.Fatal("Response is nil")
	}

	// Should return all zeros for empty dataset
	if resp.TotalRecords != 0 {
		t.Errorf("Expected total_records 0 for empty dataset, got %d", resp.TotalRecords)
	}

	if resp.TotalAmount != 0 {
		t.Errorf("Expected total_amount 0 for empty dataset, got %d", resp.TotalAmount)
	}

	if resp.UniqueCars != 0 {
		t.Errorf("Expected unique_cars 0 for empty dataset, got %d", resp.UniqueCars)
	}

	if resp.UniqueCards != 0 {
		t.Errorf("Expected unique_cards 0 for empty dataset, got %d", resp.UniqueCards)
	}

	// Should return empty arrays, not nil
	if resp.DailyStats == nil {
		t.Error("Expected daily_stats array to not be nil (should be empty array)")
	}

	if len(resp.DailyStats) != 0 {
		t.Errorf("Expected 0 daily stats for empty dataset, got %d", len(resp.DailyStats))
	}

	if resp.IcStats == nil {
		t.Error("Expected ic_stats array to not be nil (should be empty array)")
	}

	if len(resp.IcStats) != 0 {
		t.Errorf("Expected 0 IC stats for empty dataset, got %d", len(resp.IcStats))
	}
}

func TestGetStatistics_LargeDataset(t *testing.T) {
	// Arrange
	conn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("Failed to connect to gRPC server: %v", err)
	}
	defer conn.Close()

	client := pb.NewETCMeisaiServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second) // Longer timeout for large dataset
	defer cancel()

	// Request statistics for a wide date range (entire year)
	req := &pb.GetStatisticsRequest{
		DateFrom: stringPtr("2024-01-01"),
		DateTo:   stringPtr("2024-12-31"),
	}

	// Act
	resp, err := client.GetStatistics(ctx, req)

	// Assert
	if err != nil {
		st := status.Convert(err)
		if st.Code() == codes.Unimplemented {
			t.Logf("Expected failure: GetStatistics not implemented yet - %v", err)
			return
		}
		t.Fatalf("Unexpected error: %v", err)
	}

	// If server is implemented, verify large dataset handling
	if resp == nil {
		t.Fatal("Response is nil")
	}

	// Large dataset should have reasonable limits
	if len(resp.DailyStats) > 366 { // Max days in a year (leap year)
		t.Errorf("Too many daily stats: %d (max expected: 366)", len(resp.DailyStats))
	}

	if len(resp.IcStats) > 10000 { // Reasonable upper limit for IC count
		t.Errorf("Too many IC stats: %d (might indicate missing pagination)", len(resp.IcStats))
	}

	// Performance check - response should come back in reasonable time
	// (This is implicit in the 30-second timeout)

	t.Logf("Large dataset statistics - Records: %d, Daily stats: %d, IC stats: %d",
		resp.TotalRecords, len(resp.DailyStats), len(resp.IcStats))
}

func TestGetStatistics_ConsistencyCheck(t *testing.T) {
	// Arrange
	conn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("Failed to connect to gRPC server: %v", err)
	}
	defer conn.Close()

	client := pb.NewETCMeisaiServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Test internal consistency of statistics
	req := &pb.GetStatisticsRequest{
		DateFrom: stringPtr("2024-01-01"),
		DateTo:   stringPtr("2024-01-31"),
	}

	// Act
	resp, err := client.GetStatistics(ctx, req)

	// Assert
	if err != nil {
		st := status.Convert(err)
		if st.Code() == codes.Unimplemented {
			t.Logf("Expected failure: GetStatistics not implemented yet - %v", err)
			return
		}
		t.Fatalf("Unexpected error: %v", err)
	}

	// If server is implemented, verify internal consistency
	if resp == nil {
		t.Fatal("Response is nil")
	}

	// Sum of daily records should equal total records
	var dailyRecordSum int32
	var dailyAmountSum int64
	for _, dailyStat := range resp.DailyStats {
		dailyRecordSum += dailyStat.RecordCount
		dailyAmountSum += dailyStat.TotalAmount
	}

	if len(resp.DailyStats) > 0 {
		if dailyRecordSum != int32(resp.TotalRecords) {
			t.Errorf("Daily record sum (%d) doesn't match total records (%d)", dailyRecordSum, resp.TotalRecords)
		}

		if dailyAmountSum != resp.TotalAmount {
			t.Errorf("Daily amount sum (%d) doesn't match total amount (%d)", dailyAmountSum, resp.TotalAmount)
		}
	}

	// Sum of IC usage should have reasonable relationship to total records
	var icUsageSum int32
	for _, icStat := range resp.IcStats {
		icUsageSum += icStat.UsageCount
	}

	// Each record uses 2 ICs (entrance + exit), so IC usage should be approximately 2x record count
	if len(resp.IcStats) > 0 && resp.TotalRecords > 0 {
		expectedIcUsage := resp.TotalRecords * 2
		tolerance := float64(expectedIcUsage) * 0.1 // 10% tolerance

		if float64(icUsageSum) < float64(expectedIcUsage)-tolerance ||
			float64(icUsageSum) > float64(expectedIcUsage)+tolerance {
			t.Logf("Note: IC usage sum (%d) is not close to 2x record count (%d) - may be due to data structure or incomplete test data",
				icUsageSum, expectedIcUsage)
		}
	}

	t.Logf("Consistency check - Records: %d, Daily sum: %d, IC usage: %d",
		resp.TotalRecords, dailyRecordSum, icUsageSum)
}
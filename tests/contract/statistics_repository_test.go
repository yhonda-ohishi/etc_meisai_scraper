package contract

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/yhonda-ohishi/etc_meisai/src/repositories"
	pb "github.com/yhonda-ohishi/etc_meisai/src/pb"
)

// StatisticsRepositoryContractSuite defines contract tests for StatisticsRepository
// These tests verify that any implementation of StatisticsRepository meets
// the expected behavioral contract for gRPC service compatibility
type StatisticsRepositoryContractSuite struct {
	suite.Suite
	repository repositories.StatisticsRepository
}

// TestStatisticsRepositoryContract runs the contract test suite
func TestStatisticsRepositoryContract(t *testing.T) {
	suite.Run(t, new(StatisticsRepositoryContractSuite))
}

// SetupTest initializes test data before each test
func (suite *StatisticsRepositoryContractSuite) SetupTest() {
	// Mock repository will be injected by actual implementation tests
	suite.repository = &MockStatisticsRepository{}
}

// TestCountOperationsContract verifies count operation contracts
func (suite *StatisticsRepositoryContractSuite) TestCountOperationsContract() {
	validFilter := repositories.StatisticsFilter{
		DateFrom: func() *time.Time { t, _ := time.Parse("2006-01-02", "2024-01-01"); return &t }(),
		DateTo:   func() *time.Time { t, _ := time.Parse("2006-01-02", "2024-12-31"); return &t }(),
	}

	invalidFilter := repositories.StatisticsFilter{
		DateFrom: func() *time.Time { t, _ := time.Parse("2006-01-02", "2024-12-31"); return &t }(),
		DateTo:   func() *time.Time { t, _ := time.Parse("2006-01-02", "2024-01-01"); return &t }(),
		// DateFrom > DateTo - invalid range
	}

	tests := []struct {
		name        string
		operation   func(repo repositories.StatisticsRepository, filter repositories.StatisticsFilter) error
		filter      repositories.StatisticsFilter
		expectError bool
		description string
	}{
		{
			name: "count_records_valid_filter",
			operation: func(repo repositories.StatisticsRepository, filter repositories.StatisticsFilter) error {
				_, err := repo.CountRecords(context.Background(), filter)
				return err
			},
			filter:      validFilter,
			expectError: false,
			description: "Should count records with valid filter",
		},
		{
			name: "count_records_invalid_date_range",
			operation: func(repo repositories.StatisticsRepository, filter repositories.StatisticsFilter) error {
				_, err := repo.CountRecords(context.Background(), filter)
				return err
			},
			filter:      invalidFilter,
			expectError: true,
			description: "Should reject invalid date range",
		},
		{
			name: "count_unique_vehicles_valid",
			operation: func(repo repositories.StatisticsRepository, filter repositories.StatisticsFilter) error {
				_, err := repo.CountUniqueVehicles(context.Background(), filter)
				return err
			},
			filter:      validFilter,
			expectError: false,
			description: "Should count unique vehicles with valid filter",
		},
		{
			name: "count_unique_cards_valid",
			operation: func(repo repositories.StatisticsRepository, filter repositories.StatisticsFilter) error {
				_, err := repo.CountUniqueCards(context.Background(), filter)
				return err
			},
			filter:      validFilter,
			expectError: false,
			description: "Should count unique cards with valid filter",
		},
		{
			name: "count_unique_entrance_ics_valid",
			operation: func(repo repositories.StatisticsRepository, filter repositories.StatisticsFilter) error {
				_, err := repo.CountUniqueEntranceICs(context.Background(), filter)
				return err
			},
			filter:      validFilter,
			expectError: false,
			description: "Should count unique entrance ICs with valid filter",
		},
		{
			name: "count_unique_exit_ics_valid",
			operation: func(repo repositories.StatisticsRepository, filter repositories.StatisticsFilter) error {
				_, err := repo.CountUniqueExitICs(context.Background(), filter)
				return err
			},
			filter:      validFilter,
			expectError: false,
			description: "Should count unique exit ICs with valid filter",
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			err := tt.operation(suite.repository, tt.filter)

			if tt.expectError {
				suite.Error(err, tt.description)
			} else {
				suite.NoError(err, tt.description)
			}
		})
	}
}

// TestAmountCalculationsContract verifies amount calculation contracts
func (suite *StatisticsRepositoryContractSuite) TestAmountCalculationsContract() {
	validFilter := repositories.StatisticsFilter{
		DateFrom: func() *time.Time { t, _ := time.Parse("2006-01-02", "2024-01-01"); return &t }(),
		DateTo:   func() *time.Time { t, _ := time.Parse("2006-01-02", "2024-12-31"); return &t }(),
	}

	tests := []struct {
		name        string
		operation   func(repo repositories.StatisticsRepository, filter repositories.StatisticsFilter) error
		filter      repositories.StatisticsFilter
		expectError bool
		description string
	}{
		{
			name: "sum_toll_amount_valid",
			operation: func(repo repositories.StatisticsRepository, filter repositories.StatisticsFilter) error {
				amount, err := repo.SumTollAmount(context.Background(), filter)
				if err == nil && amount < 0 {
					return status.Error(codes.Internal, "negative sum not allowed")
				}
				return err
			},
			filter:      validFilter,
			expectError: false,
			description: "Should calculate toll amount sum",
		},
		{
			name: "average_toll_amount_valid",
			operation: func(repo repositories.StatisticsRepository, filter repositories.StatisticsFilter) error {
				avg, err := repo.AverageTollAmount(context.Background(), filter)
				if err == nil && avg < 0 {
					return status.Error(codes.Internal, "negative average not allowed")
				}
				return err
			},
			filter:      validFilter,
			expectError: false,
			description: "Should calculate average toll amount",
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			err := tt.operation(suite.repository, tt.filter)

			if tt.expectError {
				suite.Error(err, tt.description)
			} else {
				suite.NoError(err, tt.description)
			}
		})
	}
}

// TestTopStatisticsContract verifies top statistics contracts
func (suite *StatisticsRepositoryContractSuite) TestTopStatisticsContract() {
	validFilter := repositories.StatisticsFilter{
		DateFrom: func() *time.Time { t, _ := time.Parse("2006-01-02", "2024-01-01"); return &t }(),
		DateTo:   func() *time.Time { t, _ := time.Parse("2006-01-02", "2024-12-31"); return &t }(),
	}

	tests := []struct {
		name          string
		operation     func(repo repositories.StatisticsRepository) error
		expectedError error
		description   string
	}{
		{
			name: "get_top_routes_valid",
			operation: func(repo repositories.StatisticsRepository) error {
				routes, err := repo.GetTopRoutes(context.Background(), validFilter, 10)
				if err == nil && len(routes) > 10 {
					return status.Error(codes.Internal, "returned more routes than limit")
				}
				return err
			},
			expectedError: nil,
			description:   "Should get top routes with valid parameters",
		},
		{
			name: "get_top_routes_invalid_limit",
			operation: func(repo repositories.StatisticsRepository) error {
				_, err := repo.GetTopRoutes(context.Background(), validFilter, 0) // Invalid limit
				return err
			},
			expectedError: status.Error(codes.InvalidArgument, "limit must be positive"),
			description:   "Should reject invalid limit",
		},
		{
			name: "get_top_routes_excessive_limit",
			operation: func(repo repositories.StatisticsRepository) error {
				_, err := repo.GetTopRoutes(context.Background(), validFilter, 1001) // Excessive limit
				return err
			},
			expectedError: status.Error(codes.InvalidArgument, "limit exceeds maximum of 1000"),
			description:   "Should reject excessive limit",
		},
		{
			name: "get_top_vehicles_valid",
			operation: func(repo repositories.StatisticsRepository) error {
				vehicles, err := repo.GetTopVehicles(context.Background(), validFilter, 5)
				if err == nil && len(vehicles) > 5 {
					return status.Error(codes.Internal, "returned more vehicles than limit")
				}
				return err
			},
			expectedError: nil,
			description:   "Should get top vehicles with valid parameters",
		},
		{
			name: "get_top_cards_valid",
			operation: func(repo repositories.StatisticsRepository) error {
				cards, err := repo.GetTopCards(context.Background(), validFilter, 3)
				if err == nil && len(cards) > 3 {
					return status.Error(codes.Internal, "returned more cards than limit")
				}
				return err
			},
			expectedError: nil,
			description:   "Should get top cards with valid parameters",
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			err := tt.operation(suite.repository)

			if tt.expectedError != nil {
				suite.Error(err, tt.description)
				suite.Equal(tt.expectedError.Error(), err.Error())
			} else {
				suite.NoError(err, tt.description)
			}
		})
	}
}

// TestDistributionContract verifies time-based distribution contracts
func (suite *StatisticsRepositoryContractSuite) TestDistributionContract() {
	validFilter := repositories.StatisticsFilter{
		DateFrom: func() *time.Time { t, _ := time.Parse("2006-01-02", "2024-01-01"); return &t }(),
		DateTo:   func() *time.Time { t, _ := time.Parse("2006-01-02", "2024-12-31"); return &t }(),
	}

	tests := []struct {
		name        string
		operation   func(repo repositories.StatisticsRepository) error
		expectError bool
		description string
	}{
		{
			name: "hourly_distribution_valid",
			operation: func(repo repositories.StatisticsRepository) error {
				hourly, err := repo.GetHourlyDistribution(context.Background(), validFilter)
				if err == nil {
					// Validate that hours are in valid range (0-23)
					for _, h := range hourly {
						if h.Hour < 0 || h.Hour > 23 {
							return status.Error(codes.Internal, "invalid hour in distribution")
						}
					}
				}
				return err
			},
			expectError: false,
			description: "Should get hourly distribution with valid hour ranges",
		},
		{
			name: "daily_distribution_valid",
			operation: func(repo repositories.StatisticsRepository) error {
				daily, err := repo.GetDailyDistribution(context.Background(), validFilter)
				if err == nil {
					// Validate that dates are within filter range
					for _, d := range daily {
						if d.Date.Before(*validFilter.DateFrom) || d.Date.After(*validFilter.DateTo) {
							return status.Error(codes.Internal, "date outside filter range")
						}
					}
				}
				return err
			},
			expectError: false,
			description: "Should get daily distribution within date range",
		},
		{
			name: "monthly_distribution_valid",
			operation: func(repo repositories.StatisticsRepository) error {
				monthly, err := repo.GetMonthlyDistribution(context.Background(), validFilter)
				if err == nil {
					// Validate that months are in valid range (1-12)
					for _, m := range monthly {
						if m.Month < 1 || m.Month > 12 {
							return status.Error(codes.Internal, "invalid month in distribution")
						}
					}
				}
				return err
			},
			expectError: false,
			description: "Should get monthly distribution with valid month ranges",
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			err := tt.operation(suite.repository)

			if tt.expectError {
				suite.Error(err, tt.description)
			} else {
				suite.NoError(err, tt.description)
			}
		})
	}
}

// TestMappingStatisticsContract verifies mapping statistics contracts
func (suite *StatisticsRepositoryContractSuite) TestMappingStatisticsContract() {
	validFilter := repositories.StatisticsFilter{
		DateFrom: func() *time.Time { t, _ := time.Parse("2006-01-02", "2024-01-01"); return &t }(),
		DateTo:   func() *time.Time { t, _ := time.Parse("2006-01-02", "2024-12-31"); return &t }(),
	}

	tests := []struct {
		name        string
		operation   func(repo repositories.StatisticsRepository) error
		expectError bool
		description string
	}{
		{
			name: "count_mapped_records",
			operation: func(repo repositories.StatisticsRepository) error {
				count, err := repo.CountMappedRecords(context.Background(), validFilter)
				if err == nil && count < 0 {
					return status.Error(codes.Internal, "negative count not allowed")
				}
				return err
			},
			expectError: false,
			description: "Should count mapped records with non-negative result",
		},
		{
			name: "count_unmapped_records",
			operation: func(repo repositories.StatisticsRepository) error {
				count, err := repo.CountUnmappedRecords(context.Background(), validFilter)
				if err == nil && count < 0 {
					return status.Error(codes.Internal, "negative count not allowed")
				}
				return err
			},
			expectError: false,
			description: "Should count unmapped records with non-negative result",
		},
		{
			name: "get_mapping_statistics",
			operation: func(repo repositories.StatisticsRepository) error {
				stats, err := repo.GetMappingStatistics(context.Background(), validFilter)
				if err == nil && stats != nil {
					// Validate that mapping rate is between 0 and 1
					if stats.MappingRate < 0 || stats.MappingRate > 1 {
						return status.Error(codes.Internal, "mapping rate must be between 0 and 1")
					}
					// Validate that total equals mapped + unmapped
					if stats.TotalRecords != stats.MappedRecords+stats.UnmappedRecords {
						return status.Error(codes.Internal, "inconsistent record counts")
					}
				}
				return err
			},
			expectError: false,
			description: "Should get mapping statistics with valid rates and consistent counts",
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			err := tt.operation(suite.repository)

			if tt.expectError {
				suite.Error(err, tt.description)
			} else {
				suite.NoError(err, tt.description)
			}
		})
	}
}

// TestFilterValidationContract verifies filter validation contracts
func (suite *StatisticsRepositoryContractSuite) TestFilterValidationContract() {
	tests := []struct {
		name        string
		filter      repositories.StatisticsFilter
		expectError bool
		description string
	}{
		{
			name: "valid_filter",
			filter: repositories.StatisticsFilter{
				DateFrom: func() *time.Time { t, _ := time.Parse("2006-01-02", "2024-01-01"); return &t }(),
				DateTo:   func() *time.Time { t, _ := time.Parse("2006-01-02", "2024-12-31"); return &t }(),
				CarNumbers:    []string{"品川123あ4567"},
				ETCNumbers:    []string{"1234567890123456"},
				MinTollAmount: func() *int { v := 100; return &v }(),
				MaxTollAmount: func() *int { v := 10000; return &v }(),
			},
			expectError: false,
			description: "Should accept valid filter with all fields",
		},
		{
			name: "invalid_date_range",
			filter: repositories.StatisticsFilter{
				DateFrom: func() *time.Time { t, _ := time.Parse("2006-01-02", "2024-12-31"); return &t }(),
				DateTo:   func() *time.Time { t, _ := time.Parse("2006-01-02", "2024-01-01"); return &t }(),
			},
			expectError: true,
			description: "Should reject filter where DateFrom > DateTo",
		},
		{
			name: "invalid_toll_amount_range",
			filter: repositories.StatisticsFilter{
				DateFrom: func() *time.Time { t, _ := time.Parse("2006-01-02", "2024-01-01"); return &t }(),
				DateTo:   func() *time.Time { t, _ := time.Parse("2006-01-02", "2024-12-31"); return &t }(),
				MinTollAmount: func() *int { v := 10000; return &v }(),
				MaxTollAmount: func() *int { v := 100; return &v }(),
			},
			expectError: true,
			description: "Should reject filter where MinTollAmount > MaxTollAmount",
		},
		{
			name: "negative_toll_amounts",
			filter: repositories.StatisticsFilter{
				MinTollAmount: func() *int { v := -100; return &v }(),
				MaxTollAmount: func() *int { v := 1000; return &v }(),
			},
			expectError: true,
			description: "Should reject negative toll amounts",
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			// Test filter validation with CountRecords as representative operation
			_, err := suite.repository.CountRecords(context.Background(), tt.filter)

			if tt.expectError {
				suite.Error(err, tt.description)
			} else {
				suite.NoError(err, tt.description)
			}
		})
	}
}

// TestHealthCheckContract verifies health check contract
func (suite *StatisticsRepositoryContractSuite) TestHealthCheckContract() {
	ctx := context.Background()

	// Health check should never return error under normal circumstances
	// and should complete within reasonable time
	err := suite.repository.Ping(ctx)
	suite.NoError(err, "Health check should succeed")
}

// MockStatisticsRepository provides mock implementation for contract testing
type MockStatisticsRepository struct {
	mock.Mock
}

func (m *MockStatisticsRepository) CountRecords(ctx context.Context, filter repositories.StatisticsFilter) (int64, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockStatisticsRepository) CountUniqueVehicles(ctx context.Context, filter repositories.StatisticsFilter) (int64, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockStatisticsRepository) CountUniqueCards(ctx context.Context, filter repositories.StatisticsFilter) (int64, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockStatisticsRepository) CountUniqueEntranceICs(ctx context.Context, filter repositories.StatisticsFilter) (int64, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockStatisticsRepository) CountUniqueExitICs(ctx context.Context, filter repositories.StatisticsFilter) (int64, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockStatisticsRepository) SumTollAmount(ctx context.Context, filter repositories.StatisticsFilter) (int64, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockStatisticsRepository) AverageTollAmount(ctx context.Context, filter repositories.StatisticsFilter) (float64, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).(float64), args.Error(1)
}

func (m *MockStatisticsRepository) GetTopRoutes(ctx context.Context, filter repositories.StatisticsFilter, limit int) ([]repositories.RouteStatistic, error) {
	args := m.Called(ctx, filter, limit)
	return args.Get(0).([]repositories.RouteStatistic), args.Error(1)
}

func (m *MockStatisticsRepository) GetTopVehicles(ctx context.Context, filter repositories.StatisticsFilter, limit int) ([]repositories.VehicleStatistic, error) {
	args := m.Called(ctx, filter, limit)
	return args.Get(0).([]repositories.VehicleStatistic), args.Error(1)
}

func (m *MockStatisticsRepository) GetTopCards(ctx context.Context, filter repositories.StatisticsFilter, limit int) ([]repositories.CardStatistic, error) {
	args := m.Called(ctx, filter, limit)
	return args.Get(0).([]repositories.CardStatistic), args.Error(1)
}

func (m *MockStatisticsRepository) GetHourlyDistribution(ctx context.Context, filter repositories.StatisticsFilter) ([]repositories.HourlyStatistic, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).([]repositories.HourlyStatistic), args.Error(1)
}

func (m *MockStatisticsRepository) GetDailyDistribution(ctx context.Context, filter repositories.StatisticsFilter) ([]repositories.DailyStatistic, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).([]repositories.DailyStatistic), args.Error(1)
}

func (m *MockStatisticsRepository) GetMonthlyDistribution(ctx context.Context, filter repositories.StatisticsFilter) ([]repositories.MonthlyStatistic, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).([]repositories.MonthlyStatistic), args.Error(1)
}

func (m *MockStatisticsRepository) CountMappedRecords(ctx context.Context, filter repositories.StatisticsFilter) (int64, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockStatisticsRepository) CountUnmappedRecords(ctx context.Context, filter repositories.StatisticsFilter) (int64, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockStatisticsRepository) GetMappingStatistics(ctx context.Context, filter repositories.StatisticsFilter) (*repositories.MappingStatistics, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).(*repositories.MappingStatistics), args.Error(1)
}

func (m *MockStatisticsRepository) Ping(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

// TestProtocolBufferCompatibility verifies gRPC message compatibility
func (suite *StatisticsRepositoryContractSuite) TestProtocolBufferCompatibility() {
	// Verify that repository statistics types can be converted to/from Protocol Buffer messages

	// Test RouteStatistic conversion
	routeStats := repositories.RouteStatistic{
		EntranceIC:  "東京IC",
		ExitIC:      "大阪IC",
		Count:       100,
		TotalAmount: 120000,
		AvgAmount:   1200.0,
	}

	pbRoute := &pb.ICStatistics{
		IcName:     routeStats.EntranceIC + " -> " + routeStats.ExitIC,
		UsageCount: int32(routeStats.Count),
		IcType:     "route",
	}

	suite.NotNil(pbRoute, "Route statistic conversion should succeed")
	suite.Equal(int32(routeStats.Count), pbRoute.UsageCount)

	// Test VehicleStatistic conversion
	vehicleStats := repositories.VehicleStatistic{
		CarNumber:   "品川123あ4567",
		Count:       50,
		TotalAmount: 60000,
		AvgAmount:   1200.0,
	}

	suite.Equal("品川123あ4567", vehicleStats.CarNumber)
	suite.Equal(int64(50), vehicleStats.Count)
	suite.Equal(int64(60000), vehicleStats.TotalAmount)
	suite.Equal(1200.0, vehicleStats.AvgAmount)

	// Test MappingStatistics conversion
	mappingStats := repositories.MappingStatistics{
		TotalRecords:    1000,
		MappedRecords:   850,
		UnmappedRecords: 150,
		MappingRate:     0.85,
	}

	pbStats := &pb.GetStatisticsResponse{
		TotalRecords: mappingStats.TotalRecords,
	}

	suite.NotNil(pbStats, "Mapping statistics conversion should succeed")
	suite.Equal(mappingStats.TotalRecords, pbStats.TotalRecords)

	// Verify mapping rate is within valid range
	suite.GreaterOrEqual(mappingStats.MappingRate, 0.0)
	suite.LessOrEqual(mappingStats.MappingRate, 1.0)
}
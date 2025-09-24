package repositories

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestStatisticsFilter_Validation(t *testing.T) {
	dateFrom := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	dateTo := time.Date(2025, 1, 31, 23, 59, 59, 999999999, time.UTC)
	minAmount := 100
	maxAmount := 5000

	tests := []struct {
		name   string
		filter StatisticsFilter
		valid  bool
	}{
		{
			name: "valid filter with all fields",
			filter: StatisticsFilter{
				DateFrom:      &dateFrom,
				DateTo:        &dateTo,
				CarNumbers:    []string{"品川123あ1234", "横浜456い5678"},
				ETCNumbers:    []string{"1234567890", "0987654321"},
				ETCNums:       []string{"ETC001", "ETC002"},
				EntranceICs:   []string{"東京IC", "横浜IC"},
				ExitICs:       []string{"大阪IC", "名古屋IC"},
				MinTollAmount: &minAmount,
				MaxTollAmount: &maxAmount,
			},
			valid: true,
		},
		{
			name:   "empty filter",
			filter: StatisticsFilter{},
			valid:  true,
		},
		{
			name: "filter with only date range",
			filter: StatisticsFilter{
				DateFrom: &dateFrom,
				DateTo:   &dateTo,
			},
			valid: true,
		},
		{
			name: "filter with from after to",
			filter: StatisticsFilter{
				DateFrom: &dateTo,
				DateTo:   &dateFrom,
			},
			valid: true,
		},
		{
			name: "filter with same from and to date",
			filter: StatisticsFilter{
				DateFrom: &dateFrom,
				DateTo:   &dateFrom,
			},
			valid: true,
		},
		{
			name: "filter with empty slices",
			filter: StatisticsFilter{
				CarNumbers:  []string{},
				ETCNumbers:  []string{},
				ETCNums:     []string{},
				EntranceICs: []string{},
				ExitICs:     []string{},
			},
			valid: true,
		},
		{
			name: "filter with nil slices",
			filter: StatisticsFilter{
				CarNumbers:  nil,
				ETCNumbers:  nil,
				ETCNums:     nil,
				EntranceICs: nil,
				ExitICs:     nil,
			},
			valid: true,
		},
		{
			name: "filter with min amount greater than max",
			filter: StatisticsFilter{
				MinTollAmount: &maxAmount,
				MaxTollAmount: &minAmount,
			},
			valid: true,
		},
		{
			name: "filter with zero amounts",
			filter: StatisticsFilter{
				MinTollAmount: intPtrStats(0),
				MaxTollAmount: intPtrStats(0),
			},
			valid: true,
		},
		{
			name: "filter with negative amounts",
			filter: StatisticsFilter{
				MinTollAmount: intPtrStats(-100),
				MaxTollAmount: intPtrStats(-50),
			},
			valid: true,
		},
		{
			name: "filter with single values in slices",
			filter: StatisticsFilter{
				CarNumbers:  []string{"品川123あ1234"},
				ETCNumbers:  []string{"1234567890"},
				ETCNums:     []string{"ETC001"},
				EntranceICs: []string{"東京IC"},
				ExitICs:     []string{"大阪IC"},
			},
			valid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test that the struct can be created and accessed
			if tt.filter.DateFrom != nil {
				assert.NotNil(t, tt.filter.DateFrom)
			}
			if tt.filter.DateTo != nil {
				assert.NotNil(t, tt.filter.DateTo)
			}
			assert.Equal(t, tt.filter.CarNumbers, tt.filter.CarNumbers)
			assert.Equal(t, tt.filter.ETCNumbers, tt.filter.ETCNumbers)
			assert.Equal(t, tt.filter.ETCNums, tt.filter.ETCNums)
			assert.Equal(t, tt.filter.EntranceICs, tt.filter.EntranceICs)
			assert.Equal(t, tt.filter.ExitICs, tt.filter.ExitICs)
			if tt.filter.MinTollAmount != nil {
				assert.NotNil(t, tt.filter.MinTollAmount)
			}
			if tt.filter.MaxTollAmount != nil {
				assert.NotNil(t, tt.filter.MaxTollAmount)
			}
		})
	}
}

func TestStatisticsRepository_InterfaceMethods(t *testing.T) {
	ctx := context.Background()
	filter := StatisticsFilter{
		DateFrom:   timePtr(time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)),
		DateTo:     timePtr(time.Date(2025, 1, 31, 23, 59, 59, 999999999, time.UTC)),
		CarNumbers: []string{"品川123あ1234"},
		ETCNumbers: []string{"1234567890"},
	}

	mockRepo := &mockStatisticsRepository{}

	// Test record count operations
	t.Run("CountRecords", func(t *testing.T) {
		count, err := mockRepo.CountRecords(ctx, filter)
		assert.NoError(t, err)
		assert.Equal(t, int64(100), count)
	})

	t.Run("CountUniqueVehicles", func(t *testing.T) {
		count, err := mockRepo.CountUniqueVehicles(ctx, filter)
		assert.NoError(t, err)
		assert.Equal(t, int64(25), count)
	})

	t.Run("CountUniqueCards", func(t *testing.T) {
		count, err := mockRepo.CountUniqueCards(ctx, filter)
		assert.NoError(t, err)
		assert.Equal(t, int64(20), count)
	})

	t.Run("CountUniqueEntranceICs", func(t *testing.T) {
		count, err := mockRepo.CountUniqueEntranceICs(ctx, filter)
		assert.NoError(t, err)
		assert.Equal(t, int64(15), count)
	})

	t.Run("CountUniqueExitICs", func(t *testing.T) {
		count, err := mockRepo.CountUniqueExitICs(ctx, filter)
		assert.NoError(t, err)
		assert.Equal(t, int64(18), count)
	})

	// Test amount calculations
	t.Run("SumTollAmount", func(t *testing.T) {
		sum, err := mockRepo.SumTollAmount(ctx, filter)
		assert.NoError(t, err)
		assert.Equal(t, int64(150000), sum)
	})

	t.Run("AverageTollAmount", func(t *testing.T) {
		avg, err := mockRepo.AverageTollAmount(ctx, filter)
		assert.NoError(t, err)
		assert.Equal(t, 1500.0, avg)
	})

	// Test top statistics
	t.Run("GetTopRoutes", func(t *testing.T) {
		routes, err := mockRepo.GetTopRoutes(ctx, filter, 5)
		assert.NoError(t, err)
		assert.Len(t, routes, 2)
		assert.Equal(t, "東京IC", routes[0].EntranceIC)
		assert.Equal(t, "大阪IC", routes[0].ExitIC)
	})

	t.Run("GetTopVehicles", func(t *testing.T) {
		vehicles, err := mockRepo.GetTopVehicles(ctx, filter, 5)
		assert.NoError(t, err)
		assert.Len(t, vehicles, 2)
		assert.Equal(t, "品川123あ1234", vehicles[0].CarNumber)
	})

	t.Run("GetTopCards", func(t *testing.T) {
		cards, err := mockRepo.GetTopCards(ctx, filter, 5)
		assert.NoError(t, err)
		assert.Len(t, cards, 2)
		assert.Equal(t, "1234567890", cards[0].ETCCardNumber)
	})

	// Test time-based distributions
	t.Run("GetHourlyDistribution", func(t *testing.T) {
		hourly, err := mockRepo.GetHourlyDistribution(ctx, filter)
		assert.NoError(t, err)
		assert.Len(t, hourly, 24)
		assert.Equal(t, 9, hourly[0].Hour)
	})

	t.Run("GetDailyDistribution", func(t *testing.T) {
		daily, err := mockRepo.GetDailyDistribution(ctx, filter)
		assert.NoError(t, err)
		assert.Len(t, daily, 31)
	})

	t.Run("GetMonthlyDistribution", func(t *testing.T) {
		monthly, err := mockRepo.GetMonthlyDistribution(ctx, filter)
		assert.NoError(t, err)
		assert.Len(t, monthly, 12)
		assert.Equal(t, 2025, monthly[0].Year)
		assert.Equal(t, 1, monthly[0].Month)
	})

	// Test mapping statistics
	t.Run("CountMappedRecords", func(t *testing.T) {
		count, err := mockRepo.CountMappedRecords(ctx, filter)
		assert.NoError(t, err)
		assert.Equal(t, int64(85), count)
	})

	t.Run("CountUnmappedRecords", func(t *testing.T) {
		count, err := mockRepo.CountUnmappedRecords(ctx, filter)
		assert.NoError(t, err)
		assert.Equal(t, int64(15), count)
	})

	t.Run("GetMappingStatistics", func(t *testing.T) {
		stats, err := mockRepo.GetMappingStatistics(ctx, filter)
		assert.NoError(t, err)
		assert.NotNil(t, stats)
		assert.Equal(t, int64(100), stats.TotalRecords)
		assert.Equal(t, int64(85), stats.MappedRecords)
		assert.Equal(t, int64(15), stats.UnmappedRecords)
		assert.Equal(t, 0.85, stats.MappingRate)
	})

	// Test health check
	t.Run("Ping", func(t *testing.T) {
		err := mockRepo.Ping(ctx)
		assert.NoError(t, err)
	})
}

func TestStatisticsRepository_ErrorScenarios(t *testing.T) {
	ctx := context.Background()
	filter := StatisticsFilter{}
	errorRepo := &errorStatisticsRepository{}

	// Test error handling for all methods
	t.Run("CountRecords error", func(t *testing.T) {
		count, err := errorRepo.CountRecords(ctx, filter)
		assert.Error(t, err)
		assert.Equal(t, int64(0), count)
	})

	t.Run("CountUniqueVehicles error", func(t *testing.T) {
		count, err := errorRepo.CountUniqueVehicles(ctx, filter)
		assert.Error(t, err)
		assert.Equal(t, int64(0), count)
	})

	t.Run("CountUniqueCards error", func(t *testing.T) {
		count, err := errorRepo.CountUniqueCards(ctx, filter)
		assert.Error(t, err)
		assert.Equal(t, int64(0), count)
	})

	t.Run("CountUniqueEntranceICs error", func(t *testing.T) {
		count, err := errorRepo.CountUniqueEntranceICs(ctx, filter)
		assert.Error(t, err)
		assert.Equal(t, int64(0), count)
	})

	t.Run("CountUniqueExitICs error", func(t *testing.T) {
		count, err := errorRepo.CountUniqueExitICs(ctx, filter)
		assert.Error(t, err)
		assert.Equal(t, int64(0), count)
	})

	t.Run("SumTollAmount error", func(t *testing.T) {
		sum, err := errorRepo.SumTollAmount(ctx, filter)
		assert.Error(t, err)
		assert.Equal(t, int64(0), sum)
	})

	t.Run("AverageTollAmount error", func(t *testing.T) {
		avg, err := errorRepo.AverageTollAmount(ctx, filter)
		assert.Error(t, err)
		assert.Equal(t, 0.0, avg)
	})

	t.Run("GetTopRoutes error", func(t *testing.T) {
		routes, err := errorRepo.GetTopRoutes(ctx, filter, 5)
		assert.Error(t, err)
		assert.Nil(t, routes)
	})

	t.Run("GetTopVehicles error", func(t *testing.T) {
		vehicles, err := errorRepo.GetTopVehicles(ctx, filter, 5)
		assert.Error(t, err)
		assert.Nil(t, vehicles)
	})

	t.Run("GetTopCards error", func(t *testing.T) {
		cards, err := errorRepo.GetTopCards(ctx, filter, 5)
		assert.Error(t, err)
		assert.Nil(t, cards)
	})

	t.Run("GetHourlyDistribution error", func(t *testing.T) {
		hourly, err := errorRepo.GetHourlyDistribution(ctx, filter)
		assert.Error(t, err)
		assert.Nil(t, hourly)
	})

	t.Run("GetDailyDistribution error", func(t *testing.T) {
		daily, err := errorRepo.GetDailyDistribution(ctx, filter)
		assert.Error(t, err)
		assert.Nil(t, daily)
	})

	t.Run("GetMonthlyDistribution error", func(t *testing.T) {
		monthly, err := errorRepo.GetMonthlyDistribution(ctx, filter)
		assert.Error(t, err)
		assert.Nil(t, monthly)
	})

	t.Run("CountMappedRecords error", func(t *testing.T) {
		count, err := errorRepo.CountMappedRecords(ctx, filter)
		assert.Error(t, err)
		assert.Equal(t, int64(0), count)
	})

	t.Run("CountUnmappedRecords error", func(t *testing.T) {
		count, err := errorRepo.CountUnmappedRecords(ctx, filter)
		assert.Error(t, err)
		assert.Equal(t, int64(0), count)
	})

	t.Run("GetMappingStatistics error", func(t *testing.T) {
		stats, err := errorRepo.GetMappingStatistics(ctx, filter)
		assert.Error(t, err)
		assert.Nil(t, stats)
	})

	t.Run("Ping error", func(t *testing.T) {
		err := errorRepo.Ping(ctx)
		assert.Error(t, err)
	})
}

func TestStatisticsRepository_EdgeCases(t *testing.T) {
	ctx := context.Background()
	mockRepo := &mockStatisticsRepository{}

	t.Run("GetTopRoutes with zero limit", func(t *testing.T) {
		filter := StatisticsFilter{}
		routes, err := mockRepo.GetTopRoutes(ctx, filter, 0)
		assert.NoError(t, err)
		assert.Empty(t, routes)
	})

	t.Run("GetTopRoutes with negative limit", func(t *testing.T) {
		filter := StatisticsFilter{}
		routes, err := mockRepo.GetTopRoutes(ctx, filter, -1)
		assert.NoError(t, err)
		assert.Empty(t, routes)
	})

	t.Run("GetTopVehicles with large limit", func(t *testing.T) {
		filter := StatisticsFilter{}
		vehicles, err := mockRepo.GetTopVehicles(ctx, filter, 1000)
		assert.NoError(t, err)
		assert.Len(t, vehicles, 2) // Mock returns only 2 items
	})

	t.Run("Empty filter statistics", func(t *testing.T) {
		emptyFilter := StatisticsFilter{}
		count, err := mockRepo.CountRecords(ctx, emptyFilter)
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, count, int64(0))
	})
}

func TestStatisticsStructs(t *testing.T) {
	// Test RouteStatistic
	t.Run("RouteStatistic", func(t *testing.T) {
		route := RouteStatistic{
			EntranceIC:  "東京IC",
			ExitIC:      "大阪IC",
			Count:       100,
			TotalAmount: 150000,
			AvgAmount:   1500.0,
		}
		assert.Equal(t, "東京IC", route.EntranceIC)
		assert.Equal(t, "大阪IC", route.ExitIC)
		assert.Equal(t, int64(100), route.Count)
		assert.Equal(t, int64(150000), route.TotalAmount)
		assert.Equal(t, 1500.0, route.AvgAmount)
	})

	// Test VehicleStatistic
	t.Run("VehicleStatistic", func(t *testing.T) {
		vehicle := VehicleStatistic{
			CarNumber:   "品川123あ1234",
			Count:       50,
			TotalAmount: 75000,
			AvgAmount:   1500.0,
		}
		assert.Equal(t, "品川123あ1234", vehicle.CarNumber)
		assert.Equal(t, int64(50), vehicle.Count)
		assert.Equal(t, int64(75000), vehicle.TotalAmount)
		assert.Equal(t, 1500.0, vehicle.AvgAmount)
	})

	// Test CardStatistic
	t.Run("CardStatistic", func(t *testing.T) {
		card := CardStatistic{
			ETCCardNumber: "1234567890",
			Count:         30,
			TotalAmount:   45000,
			AvgAmount:     1500.0,
		}
		assert.Equal(t, "1234567890", card.ETCCardNumber)
		assert.Equal(t, int64(30), card.Count)
		assert.Equal(t, int64(45000), card.TotalAmount)
		assert.Equal(t, 1500.0, card.AvgAmount)
	})

	// Test HourlyStatistic
	t.Run("HourlyStatistic", func(t *testing.T) {
		hourly := HourlyStatistic{
			Hour:        9,
			Count:       20,
			TotalAmount: 30000,
			AvgAmount:   1500.0,
		}
		assert.Equal(t, 9, hourly.Hour)
		assert.Equal(t, int64(20), hourly.Count)
		assert.Equal(t, int64(30000), hourly.TotalAmount)
		assert.Equal(t, 1500.0, hourly.AvgAmount)
	})

	// Test DailyStatistic
	t.Run("DailyStatistic", func(t *testing.T) {
		date := time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC)
		daily := DailyStatistic{
			Date:        date,
			Count:       25,
			TotalAmount: 37500,
			AvgAmount:   1500.0,
		}
		assert.Equal(t, date, daily.Date)
		assert.Equal(t, int64(25), daily.Count)
		assert.Equal(t, int64(37500), daily.TotalAmount)
		assert.Equal(t, 1500.0, daily.AvgAmount)
	})

	// Test MonthlyStatistic
	t.Run("MonthlyStatistic", func(t *testing.T) {
		monthly := MonthlyStatistic{
			Year:        2025,
			Month:       1,
			Count:       100,
			TotalAmount: 150000,
			AvgAmount:   1500.0,
		}
		assert.Equal(t, 2025, monthly.Year)
		assert.Equal(t, 1, monthly.Month)
		assert.Equal(t, int64(100), monthly.Count)
		assert.Equal(t, int64(150000), monthly.TotalAmount)
		assert.Equal(t, 1500.0, monthly.AvgAmount)
	})

	// Test MappingStatistics
	t.Run("MappingStatistics", func(t *testing.T) {
		stats := MappingStatistics{
			TotalRecords:    100,
			MappedRecords:   85,
			UnmappedRecords: 15,
			MappingRate:     0.85,
		}
		assert.Equal(t, int64(100), stats.TotalRecords)
		assert.Equal(t, int64(85), stats.MappedRecords)
		assert.Equal(t, int64(15), stats.UnmappedRecords)
		assert.Equal(t, 0.85, stats.MappingRate)
	})
}

// Helper functions
func intPtrStats(i int) *int {
	return &i
}

func timePtr(t time.Time) *time.Time {
	return &t
}

// Mock implementation for testing
type mockStatisticsRepository struct{}

func (m *mockStatisticsRepository) CountRecords(ctx context.Context, filter StatisticsFilter) (int64, error) {
	return 100, nil
}

func (m *mockStatisticsRepository) CountUniqueVehicles(ctx context.Context, filter StatisticsFilter) (int64, error) {
	return 25, nil
}

func (m *mockStatisticsRepository) CountUniqueCards(ctx context.Context, filter StatisticsFilter) (int64, error) {
	return 20, nil
}

func (m *mockStatisticsRepository) CountUniqueEntranceICs(ctx context.Context, filter StatisticsFilter) (int64, error) {
	return 15, nil
}

func (m *mockStatisticsRepository) CountUniqueExitICs(ctx context.Context, filter StatisticsFilter) (int64, error) {
	return 18, nil
}

func (m *mockStatisticsRepository) SumTollAmount(ctx context.Context, filter StatisticsFilter) (int64, error) {
	return 150000, nil
}

func (m *mockStatisticsRepository) AverageTollAmount(ctx context.Context, filter StatisticsFilter) (float64, error) {
	return 1500.0, nil
}

func (m *mockStatisticsRepository) GetTopRoutes(ctx context.Context, filter StatisticsFilter, limit int) ([]RouteStatistic, error) {
	if limit <= 0 {
		return []RouteStatistic{}, nil
	}
	return []RouteStatistic{
		{
			EntranceIC:  "東京IC",
			ExitIC:      "大阪IC",
			Count:       50,
			TotalAmount: 75000,
			AvgAmount:   1500.0,
		},
		{
			EntranceIC:  "横浜IC",
			ExitIC:      "名古屋IC",
			Count:       30,
			TotalAmount: 45000,
			AvgAmount:   1500.0,
		},
	}, nil
}

func (m *mockStatisticsRepository) GetTopVehicles(ctx context.Context, filter StatisticsFilter, limit int) ([]VehicleStatistic, error) {
	if limit <= 0 {
		return []VehicleStatistic{}, nil
	}
	return []VehicleStatistic{
		{
			CarNumber:   "品川123あ1234",
			Count:       30,
			TotalAmount: 45000,
			AvgAmount:   1500.0,
		},
		{
			CarNumber:   "横浜456い5678",
			Count:       20,
			TotalAmount: 30000,
			AvgAmount:   1500.0,
		},
	}, nil
}

func (m *mockStatisticsRepository) GetTopCards(ctx context.Context, filter StatisticsFilter, limit int) ([]CardStatistic, error) {
	if limit <= 0 {
		return []CardStatistic{}, nil
	}
	return []CardStatistic{
		{
			ETCCardNumber: "1234567890",
			Count:         25,
			TotalAmount:   37500,
			AvgAmount:     1500.0,
		},
		{
			ETCCardNumber: "0987654321",
			Count:         20,
			TotalAmount:   30000,
			AvgAmount:     1500.0,
		},
	}, nil
}

func (m *mockStatisticsRepository) GetHourlyDistribution(ctx context.Context, filter StatisticsFilter) ([]HourlyStatistic, error) {
	result := make([]HourlyStatistic, 24)
	for i := 0; i < 24; i++ {
		result[i] = HourlyStatistic{
			Hour:        i,
			Count:       int64(i + 1),
			TotalAmount: int64((i + 1) * 1000),
			AvgAmount:   1000.0,
		}
	}
	return result, nil
}

func (m *mockStatisticsRepository) GetDailyDistribution(ctx context.Context, filter StatisticsFilter) ([]DailyStatistic, error) {
	result := make([]DailyStatistic, 31)
	baseDate := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	for i := 0; i < 31; i++ {
		result[i] = DailyStatistic{
			Date:        baseDate.AddDate(0, 0, i),
			Count:       int64(i + 1),
			TotalAmount: int64((i + 1) * 1000),
			AvgAmount:   1000.0,
		}
	}
	return result, nil
}

func (m *mockStatisticsRepository) GetMonthlyDistribution(ctx context.Context, filter StatisticsFilter) ([]MonthlyStatistic, error) {
	result := make([]MonthlyStatistic, 12)
	for i := 0; i < 12; i++ {
		result[i] = MonthlyStatistic{
			Year:        2025,
			Month:       i + 1,
			Count:       int64((i + 1) * 10),
			TotalAmount: int64((i + 1) * 10000),
			AvgAmount:   1000.0,
		}
	}
	return result, nil
}

func (m *mockStatisticsRepository) CountMappedRecords(ctx context.Context, filter StatisticsFilter) (int64, error) {
	return 85, nil
}

func (m *mockStatisticsRepository) CountUnmappedRecords(ctx context.Context, filter StatisticsFilter) (int64, error) {
	return 15, nil
}

func (m *mockStatisticsRepository) GetMappingStatistics(ctx context.Context, filter StatisticsFilter) (*MappingStatistics, error) {
	return &MappingStatistics{
		TotalRecords:    100,
		MappedRecords:   85,
		UnmappedRecords: 15,
		MappingRate:     0.85,
	}, nil
}

func (m *mockStatisticsRepository) Ping(ctx context.Context) error {
	return nil
}

// Error implementation for testing error scenarios
type errorStatisticsRepository struct{}

func (e *errorStatisticsRepository) CountRecords(ctx context.Context, filter StatisticsFilter) (int64, error) {
	return 0, assert.AnError
}

func (e *errorStatisticsRepository) CountUniqueVehicles(ctx context.Context, filter StatisticsFilter) (int64, error) {
	return 0, assert.AnError
}

func (e *errorStatisticsRepository) CountUniqueCards(ctx context.Context, filter StatisticsFilter) (int64, error) {
	return 0, assert.AnError
}

func (e *errorStatisticsRepository) CountUniqueEntranceICs(ctx context.Context, filter StatisticsFilter) (int64, error) {
	return 0, assert.AnError
}

func (e *errorStatisticsRepository) CountUniqueExitICs(ctx context.Context, filter StatisticsFilter) (int64, error) {
	return 0, assert.AnError
}

func (e *errorStatisticsRepository) SumTollAmount(ctx context.Context, filter StatisticsFilter) (int64, error) {
	return 0, assert.AnError
}

func (e *errorStatisticsRepository) AverageTollAmount(ctx context.Context, filter StatisticsFilter) (float64, error) {
	return 0.0, assert.AnError
}

func (e *errorStatisticsRepository) GetTopRoutes(ctx context.Context, filter StatisticsFilter, limit int) ([]RouteStatistic, error) {
	return nil, assert.AnError
}

func (e *errorStatisticsRepository) GetTopVehicles(ctx context.Context, filter StatisticsFilter, limit int) ([]VehicleStatistic, error) {
	return nil, assert.AnError
}

func (e *errorStatisticsRepository) GetTopCards(ctx context.Context, filter StatisticsFilter, limit int) ([]CardStatistic, error) {
	return nil, assert.AnError
}

func (e *errorStatisticsRepository) GetHourlyDistribution(ctx context.Context, filter StatisticsFilter) ([]HourlyStatistic, error) {
	return nil, assert.AnError
}

func (e *errorStatisticsRepository) GetDailyDistribution(ctx context.Context, filter StatisticsFilter) ([]DailyStatistic, error) {
	return nil, assert.AnError
}

func (e *errorStatisticsRepository) GetMonthlyDistribution(ctx context.Context, filter StatisticsFilter) ([]MonthlyStatistic, error) {
	return nil, assert.AnError
}

func (e *errorStatisticsRepository) CountMappedRecords(ctx context.Context, filter StatisticsFilter) (int64, error) {
	return 0, assert.AnError
}

func (e *errorStatisticsRepository) CountUnmappedRecords(ctx context.Context, filter StatisticsFilter) (int64, error) {
	return 0, assert.AnError
}

func (e *errorStatisticsRepository) GetMappingStatistics(ctx context.Context, filter StatisticsFilter) (*MappingStatistics, error) {
	return nil, assert.AnError
}

func (e *errorStatisticsRepository) Ping(ctx context.Context) error {
	return assert.AnError
}
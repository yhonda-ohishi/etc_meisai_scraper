package services

import (
	"context"
	"errors"
	"log"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/yhonda-ohishi/etc_meisai/src/mocks"
	"github.com/yhonda-ohishi/etc_meisai/src/repositories"
)

func TestNewStatisticsService(t *testing.T) {
	t.Parallel()

	t.Run("with repository and logger", func(t *testing.T) {
		mockRepo := &mocks.MockStatisticsRepository{}
		logger := log.New(os.Stdout, "test", log.LstdFlags)

		service := NewStatisticsService(mockRepo, logger)

		assert.NotNil(t, service)
		assert.Equal(t, mockRepo, service.statsRepo)
		assert.Equal(t, logger, service.logger)
	})

	t.Run("with repository, no logger", func(t *testing.T) {
		mockRepo := &mocks.MockStatisticsRepository{}

		service := NewStatisticsService(mockRepo, nil)

		assert.NotNil(t, service)
		assert.Equal(t, mockRepo, service.statsRepo)
		assert.NotNil(t, service.logger)
	})
}

func TestStatisticsService_GetGeneralStatistics(t *testing.T) {
	t.Parallel()

	dateFrom := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	dateTo := time.Date(2024, 1, 31, 0, 0, 0, 0, time.UTC)

	filter := &StatisticsFilter{
		DateFrom:      &dateFrom,
		DateTo:        &dateTo,
		CarNumbers:    []string{"あ123", "い456"},
		ETCNumbers:    []string{"1234567890123456"},
		MinTollAmount: intPtr(500),
		MaxTollAmount: intPtr(2000),
	}

	expectedRoutes := []repositories.RouteStatistic{
		{EntranceIC: "羽田空港IC", ExitIC: "新宿IC", Count: 10, TotalAmount: 12000, AvgAmount: 1200},
	}
	expectedVehicles := []repositories.VehicleStatistic{
		{CarNumber: "あ123", Count: 20, TotalAmount: 24000, AvgAmount: 1200},
	}
	expectedCards := []repositories.CardStatistic{
		{ETCCardNumber: "1234567890123456", Count: 15, TotalAmount: 18000, AvgAmount: 1200},
	}
	expectedHourly := []repositories.HourlyStatistic{
		{Hour: 10, Count: 5, TotalAmount: 6000, AvgAmount: 1200},
	}

	tests := []struct {
		name        string
		filter      *StatisticsFilter
		setupMock   func(*mocks.MockStatisticsRepository)
		expectError bool
		errorMsg    string
	}{
		{
			name:   "successful statistics generation",
			filter: filter,
			setupMock: func(m *mocks.MockStatisticsRepository) {
				repoFilter := repositories.StatisticsFilter{
					DateFrom:      filter.DateFrom,
					DateTo:        filter.DateTo,
					CarNumbers:    filter.CarNumbers,
					ETCNumbers:    filter.ETCNumbers,
					MinTollAmount: filter.MinTollAmount,
					MaxTollAmount: filter.MaxTollAmount,
				}

				m.On("CountRecords", mock.Anything, repoFilter).Return(int64(100), nil)
				m.On("SumTollAmount", mock.Anything, repoFilter).Return(int64(120000), nil)
				m.On("AverageTollAmount", mock.Anything, repoFilter).Return(1200.0, nil)
				m.On("CountUniqueVehicles", mock.Anything, repoFilter).Return(int64(5), nil)
				m.On("CountUniqueCards", mock.Anything, repoFilter).Return(int64(3), nil)
				m.On("CountUniqueEntranceICs", mock.Anything, repoFilter).Return(int64(10), nil)
				m.On("CountUniqueExitICs", mock.Anything, repoFilter).Return(int64(8), nil)
				m.On("GetTopRoutes", mock.Anything, repoFilter, 10).Return(expectedRoutes, nil)
				m.On("GetTopVehicles", mock.Anything, repoFilter, 10).Return(expectedVehicles, nil)
				m.On("GetTopCards", mock.Anything, repoFilter, 10).Return(expectedCards, nil)
				m.On("GetHourlyDistribution", mock.Anything, repoFilter).Return(expectedHourly, nil)
			},
			expectError: false,
		},
		{
			name:   "nil filter",
			filter: nil,
			setupMock: func(m *mocks.MockStatisticsRepository) {
				emptyFilter := repositories.StatisticsFilter{}

				m.On("CountRecords", mock.Anything, emptyFilter).Return(int64(50), nil)
				m.On("SumTollAmount", mock.Anything, emptyFilter).Return(int64(60000), nil)
				m.On("AverageTollAmount", mock.Anything, emptyFilter).Return(1200.0, nil)
				m.On("CountUniqueVehicles", mock.Anything, emptyFilter).Return(int64(3), nil)
				m.On("CountUniqueCards", mock.Anything, emptyFilter).Return(int64(2), nil)
				m.On("CountUniqueEntranceICs", mock.Anything, emptyFilter).Return(int64(5), nil)
				m.On("CountUniqueExitICs", mock.Anything, emptyFilter).Return(int64(4), nil)
				m.On("GetTopRoutes", mock.Anything, emptyFilter, 10).Return([]repositories.RouteStatistic{}, nil)
				m.On("GetTopVehicles", mock.Anything, emptyFilter, 10).Return([]repositories.VehicleStatistic{}, nil)
				m.On("GetTopCards", mock.Anything, emptyFilter, 10).Return([]repositories.CardStatistic{}, nil)
				m.On("GetHourlyDistribution", mock.Anything, emptyFilter).Return([]repositories.HourlyStatistic{}, nil)
			},
			expectError: false,
		},
		{
			name:   "CountRecords error",
			filter: filter,
			setupMock: func(m *mocks.MockStatisticsRepository) {
				m.On("CountRecords", mock.Anything, mock.AnythingOfType("repositories.StatisticsFilter")).Return(int64(0), errors.New("db error"))
			},
			expectError: true,
			errorMsg:    "failed to count records",
		},
		{
			name:   "SumTollAmount error",
			filter: filter,
			setupMock: func(m *mocks.MockStatisticsRepository) {
				m.On("CountRecords", mock.Anything, mock.AnythingOfType("repositories.StatisticsFilter")).Return(int64(100), nil)
				m.On("SumTollAmount", mock.Anything, mock.AnythingOfType("repositories.StatisticsFilter")).Return(int64(0), errors.New("sum error"))
			},
			expectError: true,
			errorMsg:    "failed to sum toll amount",
		},
		{
			name:   "AverageTollAmount error",
			filter: filter,
			setupMock: func(m *mocks.MockStatisticsRepository) {
				m.On("CountRecords", mock.Anything, mock.AnythingOfType("repositories.StatisticsFilter")).Return(int64(100), nil)
				m.On("SumTollAmount", mock.Anything, mock.AnythingOfType("repositories.StatisticsFilter")).Return(int64(120000), nil)
				m.On("AverageTollAmount", mock.Anything, mock.AnythingOfType("repositories.StatisticsFilter")).Return(0.0, errors.New("avg error"))
			},
			expectError: true,
			errorMsg:    "failed to calculate average toll amount",
		},
		{
			name:   "GetTopRoutes error",
			filter: filter,
			setupMock: func(m *mocks.MockStatisticsRepository) {
				repoFilter := repositories.StatisticsFilter{
					DateFrom:      filter.DateFrom,
					DateTo:        filter.DateTo,
					CarNumbers:    filter.CarNumbers,
					ETCNumbers:    filter.ETCNumbers,
					MinTollAmount: filter.MinTollAmount,
					MaxTollAmount: filter.MaxTollAmount,
				}

				m.On("CountRecords", mock.Anything, repoFilter).Return(int64(100), nil)
				m.On("SumTollAmount", mock.Anything, repoFilter).Return(int64(120000), nil)
				m.On("AverageTollAmount", mock.Anything, repoFilter).Return(1200.0, nil)
				m.On("CountUniqueVehicles", mock.Anything, repoFilter).Return(int64(5), nil)
				m.On("CountUniqueCards", mock.Anything, repoFilter).Return(int64(3), nil)
				m.On("CountUniqueEntranceICs", mock.Anything, repoFilter).Return(int64(10), nil)
				m.On("CountUniqueExitICs", mock.Anything, repoFilter).Return(int64(8), nil)
				m.On("GetTopRoutes", mock.Anything, repoFilter, 10).Return(nil, errors.New("routes error"))
			},
			expectError: true,
			errorMsg:    "failed to get top routes",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &mocks.MockStatisticsRepository{}
			tt.setupMock(mockRepo)

			service := NewStatisticsService(mockRepo, nil)
			ctx := context.Background()

			stats, err := service.GetGeneralStatistics(ctx, tt.filter)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
				assert.Nil(t, stats)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, stats)
				assert.NotEmpty(t, stats.DateRange)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestStatisticsService_GetDailyStatistics(t *testing.T) {
	t.Parallel()

	dateFrom := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	dateTo := time.Date(2024, 1, 31, 0, 0, 0, 0, time.UTC)

	filter := &StatisticsFilter{
		DateFrom: &dateFrom,
		DateTo:   &dateTo,
	}

	expectedDaily := []repositories.DailyStatistic{
		{Date: dateFrom, Count: 10, TotalAmount: 12000, AvgAmount: 1200},
		{Date: dateFrom.AddDate(0, 0, 1), Count: 15, TotalAmount: 18000, AvgAmount: 1200},
	}

	tests := []struct {
		name        string
		filter      *StatisticsFilter
		setupMock   func(*mocks.MockStatisticsRepository)
		expectError bool
		errorMsg    string
	}{
		{
			name:   "successful daily statistics",
			filter: filter,
			setupMock: func(m *mocks.MockStatisticsRepository) {
				m.On("GetDailyDistribution", mock.Anything, mock.AnythingOfType("repositories.StatisticsFilter")).Return(expectedDaily, nil)
			},
			expectError: false,
		},
		{
			name:   "repository error",
			filter: filter,
			setupMock: func(m *mocks.MockStatisticsRepository) {
				m.On("GetDailyDistribution", mock.Anything, mock.AnythingOfType("repositories.StatisticsFilter")).Return(nil, errors.New("db error"))
			},
			expectError: true,
			errorMsg:    "failed to get daily distribution",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &mocks.MockStatisticsRepository{}
			tt.setupMock(mockRepo)

			service := NewStatisticsService(mockRepo, nil)
			ctx := context.Background()

			response, err := service.GetDailyStatistics(ctx, tt.filter)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
				assert.Nil(t, response)
			} else {
				assert.NoError(t, err)
				require.NotNil(t, response)
				assert.NotEmpty(t, response.DateRange)
				assert.Len(t, response.Statistics, len(expectedDaily))
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestStatisticsService_GetMonthlyStatistics(t *testing.T) {
	t.Parallel()

	dateFrom := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	dateTo := time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC)

	filter := &StatisticsFilter{
		DateFrom: &dateFrom,
		DateTo:   &dateTo,
	}

	expectedMonthly := []repositories.MonthlyStatistic{
		{Year: 2024, Month: 1, Count: 100, TotalAmount: 120000, AvgAmount: 1200},
		{Year: 2024, Month: 2, Count: 90, TotalAmount: 108000, AvgAmount: 1200},
	}

	tests := []struct {
		name        string
		filter      *StatisticsFilter
		setupMock   func(*mocks.MockStatisticsRepository)
		expectError bool
		errorMsg    string
	}{
		{
			name:   "successful monthly statistics",
			filter: filter,
			setupMock: func(m *mocks.MockStatisticsRepository) {
				m.On("GetMonthlyDistribution", mock.Anything, mock.AnythingOfType("repositories.StatisticsFilter")).Return(expectedMonthly, nil)
			},
			expectError: false,
		},
		{
			name:   "repository error",
			filter: filter,
			setupMock: func(m *mocks.MockStatisticsRepository) {
				m.On("GetMonthlyDistribution", mock.Anything, mock.AnythingOfType("repositories.StatisticsFilter")).Return(nil, errors.New("db error"))
			},
			expectError: true,
			errorMsg:    "failed to get monthly distribution",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &mocks.MockStatisticsRepository{}
			tt.setupMock(mockRepo)

			service := NewStatisticsService(mockRepo, nil)
			ctx := context.Background()

			response, err := service.GetMonthlyStatistics(ctx, tt.filter)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
				assert.Nil(t, response)
			} else {
				assert.NoError(t, err)
				require.NotNil(t, response)
				assert.NotEmpty(t, response.DateRange)
				assert.Len(t, response.Statistics, len(expectedMonthly))

				// Check month names are set
				for _, stat := range response.Statistics {
					assert.NotEmpty(t, stat.MonthName)
				}
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestStatisticsService_GetVehicleStatistics(t *testing.T) {
	t.Parallel()

	carNumbers := []string{"あ123", "い456"}
	dateFrom := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

	filter := &StatisticsFilter{
		DateFrom: &dateFrom,
	}

	expectedVehicles := []repositories.VehicleStatistic{
		{CarNumber: "あ123", Count: 20, TotalAmount: 24000, AvgAmount: 1200},
		{CarNumber: "い456", Count: 15, TotalAmount: 18000, AvgAmount: 1200},
	}

	tests := []struct {
		name        string
		carNumbers  []string
		filter      *StatisticsFilter
		setupMock   func(*mocks.MockStatisticsRepository)
		expectError bool
		errorMsg    string
	}{
		{
			name:       "successful vehicle statistics",
			carNumbers: carNumbers,
			filter:     filter,
			setupMock: func(m *mocks.MockStatisticsRepository) {
				m.On("GetTopVehicles", mock.Anything, mock.AnythingOfType("repositories.StatisticsFilter"), len(carNumbers)).Return(expectedVehicles, nil)
			},
			expectError: false,
		},
		{
			name:       "nil filter",
			carNumbers: carNumbers,
			filter:     nil,
			setupMock: func(m *mocks.MockStatisticsRepository) {
				m.On("GetTopVehicles", mock.Anything, mock.AnythingOfType("repositories.StatisticsFilter"), len(carNumbers)).Return(expectedVehicles, nil)
			},
			expectError: false,
		},
		{
			name:       "repository error",
			carNumbers: carNumbers,
			filter:     filter,
			setupMock: func(m *mocks.MockStatisticsRepository) {
				m.On("GetTopVehicles", mock.Anything, mock.AnythingOfType("repositories.StatisticsFilter"), len(carNumbers)).Return(nil, errors.New("db error"))
			},
			expectError: true,
			errorMsg:    "failed to get vehicle statistics",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &mocks.MockStatisticsRepository{}
			tt.setupMock(mockRepo)

			service := NewStatisticsService(mockRepo, nil)
			ctx := context.Background()

			response, err := service.GetVehicleStatistics(ctx, tt.carNumbers, tt.filter)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
				assert.Nil(t, response)
			} else {
				assert.NoError(t, err)
				require.NotNil(t, response)
				assert.NotEmpty(t, response.DateRange)
				assert.Len(t, response.Vehicles, len(expectedVehicles))
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestStatisticsService_GetMappingStatistics(t *testing.T) {
	t.Parallel()

	dateFrom := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	filter := &StatisticsFilter{
		DateFrom: &dateFrom,
	}

	expectedMapping := &repositories.MappingStatistics{
		TotalRecords:    100,
		MappedRecords:   80,
		UnmappedRecords: 20,
		MappingRate:     0.8,
	}

	tests := []struct {
		name        string
		filter      *StatisticsFilter
		setupMock   func(*mocks.MockStatisticsRepository)
		expectError bool
		errorMsg    string
	}{
		{
			name:   "successful mapping statistics",
			filter: filter,
			setupMock: func(m *mocks.MockStatisticsRepository) {
				m.On("GetMappingStatistics", mock.Anything, mock.AnythingOfType("repositories.StatisticsFilter")).Return(expectedMapping, nil)
			},
			expectError: false,
		},
		{
			name:   "repository error",
			filter: filter,
			setupMock: func(m *mocks.MockStatisticsRepository) {
				m.On("GetMappingStatistics", mock.Anything, mock.AnythingOfType("repositories.StatisticsFilter")).Return(nil, errors.New("db error"))
			},
			expectError: true,
			errorMsg:    "failed to get mapping statistics",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &mocks.MockStatisticsRepository{}
			tt.setupMock(mockRepo)

			service := NewStatisticsService(mockRepo, nil)
			ctx := context.Background()

			response, err := service.GetMappingStatistics(ctx, tt.filter)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
				assert.Nil(t, response)
			} else {
				assert.NoError(t, err)
				require.NotNil(t, response)
				assert.NotEmpty(t, response.DateRange)
				assert.Equal(t, expectedMapping.TotalRecords, response.TotalRecords)
				assert.Equal(t, expectedMapping.MappedRecords, response.MappedRecords)
				assert.Equal(t, expectedMapping.UnmappedRecords, response.UnmappedRecords)
				assert.Equal(t, expectedMapping.MappingRate, response.MappingRate)
				assert.Contains(t, response.MappingRatePercentage, "80.00%")
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestStatisticsService_HealthCheck(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		setupMock   func(*mocks.MockStatisticsRepository)
		expectError bool
		errorMsg    string
	}{
		{
			name: "healthy repository",
			setupMock: func(m *mocks.MockStatisticsRepository) {
				m.On("Ping", mock.Anything).Return(nil)
			},
			expectError: false,
		},
		{
			name: "unhealthy repository",
			setupMock: func(m *mocks.MockStatisticsRepository) {
				m.On("Ping", mock.Anything).Return(errors.New("connection failed"))
			},
			expectError: true,
			errorMsg:    "statistics repository ping failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &mocks.MockStatisticsRepository{}
			tt.setupMock(mockRepo)

			service := NewStatisticsService(mockRepo, nil)
			ctx := context.Background()

			err := service.HealthCheck(ctx)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestStatisticsService_HealthCheck_NilRepository(t *testing.T) {
	t.Parallel()

	service := &StatisticsService{
		statsRepo: nil,
	}
	ctx := context.Background()

	err := service.HealthCheck(ctx)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "statistics repository not initialized")
}

// Test helper methods
func TestStatisticsService_FormatDateRange(t *testing.T) {
	t.Parallel()

	mockRepo := &mocks.MockStatisticsRepository{}
	service := NewStatisticsService(mockRepo, nil)

	tests := []struct {
		name     string
		filter   *StatisticsFilter
		expected string
	}{
		{
			name:     "nil filter",
			filter:   nil,
			expected: "All Time",
		},
		{
			name:     "empty filter",
			filter:   &StatisticsFilter{},
			expected: "All Time",
		},
		{
			name: "with date from only",
			filter: &StatisticsFilter{
				DateFrom: timePtr(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)),
			},
			expected: "2024-01-01 to Present",
		},
		{
			name: "with date to only",
			filter: &StatisticsFilter{
				DateTo: timePtr(time.Date(2024, 1, 31, 0, 0, 0, 0, time.UTC)),
			},
			expected: "Beginning to 2024-01-31",
		},
		{
			name: "with both dates",
			filter: &StatisticsFilter{
				DateFrom: timePtr(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)),
				DateTo:   timePtr(time.Date(2024, 1, 31, 0, 0, 0, 0, time.UTC)),
			},
			expected: "2024-01-01 to 2024-01-31",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.formatDateRange(tt.filter)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestStatisticsService_ConversionMethods(t *testing.T) {
	t.Parallel()

	mockRepo := &mocks.MockStatisticsRepository{}
	service := NewStatisticsService(mockRepo, nil)

	t.Run("convertRouteStats", func(t *testing.T) {
		repoStats := []repositories.RouteStatistic{
			{EntranceIC: "羽田空港IC", ExitIC: "新宿IC", Count: 10, TotalAmount: 12000, AvgAmount: 1200},
		}

		result := service.convertRouteStats(repoStats)

		require.Len(t, result, 1)
		assert.Equal(t, "羽田空港IC", result[0].EntranceIC)
		assert.Equal(t, "新宿IC", result[0].ExitIC)
		assert.Equal(t, int64(10), result[0].Count)
	})

	t.Run("convertVehicleStats", func(t *testing.T) {
		repoStats := []repositories.VehicleStatistic{
			{CarNumber: "あ123", Count: 20, TotalAmount: 24000, AvgAmount: 1200},
		}

		result := service.convertVehicleStats(repoStats)

		require.Len(t, result, 1)
		assert.Equal(t, "あ123", result[0].CarNumber)
		assert.Equal(t, int64(20), result[0].Count)
	})

	t.Run("convertCardStats", func(t *testing.T) {
		repoStats := []repositories.CardStatistic{
			{ETCCardNumber: "1234567890123456", Count: 15, TotalAmount: 18000, AvgAmount: 1200},
		}

		result := service.convertCardStats(repoStats)

		require.Len(t, result, 1)
		assert.Equal(t, "1234567890123456", result[0].ETCCardNumber)
		assert.Equal(t, int64(15), result[0].Count)
	})

	t.Run("convertHourlyStats", func(t *testing.T) {
		repoStats := []repositories.HourlyStatistic{
			{Hour: 10, Count: 5, TotalAmount: 6000, AvgAmount: 1200},
		}

		result := service.convertHourlyStats(repoStats)

		require.Len(t, result, 1)
		assert.Equal(t, 10, result[0].Hour)
		assert.Equal(t, "10:00", result[0].HourLabel)
		assert.Equal(t, int64(5), result[0].Count)
	})
}

// Context cancellation tests
func TestStatisticsService_ContextCancellation(t *testing.T) {
	t.Parallel()

	t.Run("get general statistics with cancelled context", func(t *testing.T) {
		mockRepo := &mocks.MockStatisticsRepository{}
		service := NewStatisticsService(mockRepo, nil)

		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		// Mock should handle the cancelled context appropriately
		mockRepo.On("CountRecords", mock.Anything, mock.AnythingOfType("repositories.StatisticsFilter")).Return(int64(0), context.Canceled)

		_, err := service.GetGeneralStatistics(ctx, &StatisticsFilter{})

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to count records")
	})
}

// Helper functions
func intPtr(i int) *int {
	return &i
}

func timePtr(t time.Time) *time.Time {
	return &t
}
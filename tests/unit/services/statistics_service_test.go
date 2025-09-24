package services_test

import (
	"context"
	"errors"
	"log"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/yhonda-ohishi/etc_meisai/src/repositories"
	"github.com/yhonda-ohishi/etc_meisai/src/services"
)

// MockStatisticsRepository is a mock implementation of repositories.StatisticsRepository
type MockStatisticsRepository struct {
	mock.Mock
}

// Count operations
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

// Amount calculations
func (m *MockStatisticsRepository) SumTollAmount(ctx context.Context, filter repositories.StatisticsFilter) (int64, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockStatisticsRepository) AverageTollAmount(ctx context.Context, filter repositories.StatisticsFilter) (float64, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).(float64), args.Error(1)
}

// Top statistics
func (m *MockStatisticsRepository) GetTopRoutes(ctx context.Context, filter repositories.StatisticsFilter, limit int) ([]repositories.RouteStatistic, error) {
	args := m.Called(ctx, filter, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]repositories.RouteStatistic), args.Error(1)
}

func (m *MockStatisticsRepository) GetTopVehicles(ctx context.Context, filter repositories.StatisticsFilter, limit int) ([]repositories.VehicleStatistic, error) {
	args := m.Called(ctx, filter, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]repositories.VehicleStatistic), args.Error(1)
}

func (m *MockStatisticsRepository) GetTopCards(ctx context.Context, filter repositories.StatisticsFilter, limit int) ([]repositories.CardStatistic, error) {
	args := m.Called(ctx, filter, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]repositories.CardStatistic), args.Error(1)
}

// Time-based distributions
func (m *MockStatisticsRepository) GetHourlyDistribution(ctx context.Context, filter repositories.StatisticsFilter) ([]repositories.HourlyStatistic, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]repositories.HourlyStatistic), args.Error(1)
}

func (m *MockStatisticsRepository) GetDailyDistribution(ctx context.Context, filter repositories.StatisticsFilter) ([]repositories.DailyStatistic, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]repositories.DailyStatistic), args.Error(1)
}

func (m *MockStatisticsRepository) GetMonthlyDistribution(ctx context.Context, filter repositories.StatisticsFilter) ([]repositories.MonthlyStatistic, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]repositories.MonthlyStatistic), args.Error(1)
}

// Mapping statistics
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
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*repositories.MappingStatistics), args.Error(1)
}

// Health check
func (m *MockStatisticsRepository) Ping(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func TestStatisticsService_NewStatisticsService(t *testing.T) {
	mockRepo := &MockStatisticsRepository{}
	logger := log.New(os.Stdout, "TEST: ", log.LstdFlags)

	service := services.NewStatisticsService(mockRepo, logger)
	assert.NotNil(t, service)

	// Test with nil logger (should create default logger)
	service2 := services.NewStatisticsService(mockRepo, nil)
	assert.NotNil(t, service2)
}

func TestStatisticsService_GetGeneralStatistics(t *testing.T) {
	mockRepo := &MockStatisticsRepository{}
	logger := log.New(os.Stdout, "TEST: ", log.LstdFlags)
	service := services.NewStatisticsService(mockRepo, logger)

	ctx := context.Background()
	now := time.Now()
	dateFrom := now.AddDate(0, -1, 0)
	dateTo := now

	filter := &services.StatisticsFilter{
		DateFrom:    &dateFrom,
		DateTo:      &dateTo,
		CarNumbers:  []string{"品川123あ1234"},
		ETCNumbers:  []string{"1234567890123456"},
	}

	// Convert to repository filter for mocking
	repoFilter := repositories.StatisticsFilter{
		DateFrom:   filter.DateFrom,
		DateTo:     filter.DateTo,
		CarNumbers: filter.CarNumbers,
		ETCNumbers: filter.ETCNumbers,
	}

	// Setup mock expectations
	mockRepo.On("CountRecords", ctx, repoFilter).Return(int64(100), nil)
	mockRepo.On("SumTollAmount", ctx, repoFilter).Return(int64(50000), nil)
	mockRepo.On("AverageTollAmount", ctx, repoFilter).Return(float64(500), nil)
	mockRepo.On("CountUniqueVehicles", ctx, repoFilter).Return(int64(10), nil)
	mockRepo.On("CountUniqueCards", ctx, repoFilter).Return(int64(5), nil)
	mockRepo.On("CountUniqueEntranceICs", ctx, repoFilter).Return(int64(15), nil)
	mockRepo.On("CountUniqueExitICs", ctx, repoFilter).Return(int64(12), nil)

	topRoutes := []repositories.RouteStatistic{
		{EntranceIC: "東京IC", ExitIC: "大阪IC", Count: 25, TotalAmount: 12500, AvgAmount: 500},
	}
	mockRepo.On("GetTopRoutes", ctx, repoFilter, 10).Return(topRoutes, nil)

	topVehicles := []repositories.VehicleStatistic{
		{CarNumber: "品川123あ1234", Count: 30, TotalAmount: 15000, AvgAmount: 500},
	}
	mockRepo.On("GetTopVehicles", ctx, repoFilter, 10).Return(topVehicles, nil)

	topCards := []repositories.CardStatistic{
		{ETCCardNumber: "1234567890123456", Count: 40, TotalAmount: 20000, AvgAmount: 500},
	}
	mockRepo.On("GetTopCards", ctx, repoFilter, 10).Return(topCards, nil)

	hourlyDist := []repositories.HourlyStatistic{
		{Hour: 9, Count: 10, TotalAmount: 5000, AvgAmount: 500},
		{Hour: 10, Count: 15, TotalAmount: 7500, AvgAmount: 500},
	}
	mockRepo.On("GetHourlyDistribution", ctx, repoFilter).Return(hourlyDist, nil)

	// Execute
	result, err := service.GetGeneralStatistics(ctx, filter)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, int64(100), result.TotalRecords)
	assert.Equal(t, int64(50000), result.TotalAmount)
	assert.Equal(t, float64(500), result.AverageAmount)
	assert.Equal(t, int64(10), result.UniqueVehicles)
	assert.Equal(t, int64(5), result.UniqueCards)
	assert.Len(t, result.TopRoutes, 1)
	assert.Len(t, result.TopVehicles, 1)
	assert.Len(t, result.TopCards, 1)
	assert.Len(t, result.HourlyDistribution, 2)

	mockRepo.AssertExpectations(t)
}

func TestStatisticsService_GetGeneralStatistics_Error(t *testing.T) {
	mockRepo := &MockStatisticsRepository{}
	logger := log.New(os.Stdout, "TEST: ", log.LstdFlags)
	service := services.NewStatisticsService(mockRepo, logger)

	ctx := context.Background()
	filter := &services.StatisticsFilter{}
	repoFilter := repositories.StatisticsFilter{}

	// Test error from CountRecords
	mockRepo.On("CountRecords", ctx, repoFilter).Return(int64(0), errors.New("database error"))

	result, err := service.GetGeneralStatistics(ctx, filter)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to count records")

	mockRepo.AssertExpectations(t)
}

func TestStatisticsService_GetDailyStatistics(t *testing.T) {
	mockRepo := &MockStatisticsRepository{}
	logger := log.New(os.Stdout, "TEST: ", log.LstdFlags)
	service := services.NewStatisticsService(mockRepo, logger)

	ctx := context.Background()
	filter := &services.StatisticsFilter{}
	repoFilter := repositories.StatisticsFilter{}

	dailyDist := []repositories.DailyStatistic{
		{Date: time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC), Count: 10, TotalAmount: 5000, AvgAmount: 500},
		{Date: time.Date(2025, 1, 16, 0, 0, 0, 0, time.UTC), Count: 15, TotalAmount: 7500, AvgAmount: 500},
	}

	mockRepo.On("GetDailyDistribution", ctx, repoFilter).Return(dailyDist, nil)

	result, err := service.GetDailyStatistics(ctx, filter)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result.Statistics, 2)
	assert.Equal(t, "2025-01-15", result.Statistics[0].Date)
	assert.Equal(t, int64(10), result.Statistics[0].Count)

	mockRepo.AssertExpectations(t)
}

func TestStatisticsService_GetMonthlyStatistics(t *testing.T) {
	mockRepo := &MockStatisticsRepository{}
	logger := log.New(os.Stdout, "TEST: ", log.LstdFlags)
	service := services.NewStatisticsService(mockRepo, logger)

	ctx := context.Background()
	filter := &services.StatisticsFilter{}
	repoFilter := repositories.StatisticsFilter{}

	monthlyDist := []repositories.MonthlyStatistic{
		{Year: 2025, Month: 1, Count: 100, TotalAmount: 50000, AvgAmount: 500},
		{Year: 2025, Month: 2, Count: 120, TotalAmount: 60000, AvgAmount: 500},
	}

	mockRepo.On("GetMonthlyDistribution", ctx, repoFilter).Return(monthlyDist, nil)

	result, err := service.GetMonthlyStatistics(ctx, filter)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result.Statistics, 2)
	assert.Equal(t, 2025, result.Statistics[0].Year)
	assert.Equal(t, 1, result.Statistics[0].Month)
	assert.Equal(t, "January", result.Statistics[0].MonthName)

	mockRepo.AssertExpectations(t)
}

func TestStatisticsService_GetVehicleStatistics(t *testing.T) {
	mockRepo := &MockStatisticsRepository{}
	logger := log.New(os.Stdout, "TEST: ", log.LstdFlags)
	service := services.NewStatisticsService(mockRepo, logger)

	ctx := context.Background()
	carNumbers := []string{"品川123あ1234", "品川456い5678"}
	filter := &services.StatisticsFilter{}

	repoFilter := repositories.StatisticsFilter{
		CarNumbers: carNumbers,
	}

	vehicleStats := []repositories.VehicleStatistic{
		{CarNumber: "品川123あ1234", Count: 30, TotalAmount: 15000, AvgAmount: 500},
		{CarNumber: "品川456い5678", Count: 20, TotalAmount: 10000, AvgAmount: 500},
	}

	mockRepo.On("GetTopVehicles", ctx, repoFilter, 2).Return(vehicleStats, nil)

	result, err := service.GetVehicleStatistics(ctx, carNumbers, filter)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result.Vehicles, 2)
	assert.Equal(t, "品川123あ1234", result.Vehicles[0].CarNumber)

	mockRepo.AssertExpectations(t)
}

func TestStatisticsService_GetMappingStatistics(t *testing.T) {
	mockRepo := &MockStatisticsRepository{}
	logger := log.New(os.Stdout, "TEST: ", log.LstdFlags)
	service := services.NewStatisticsService(mockRepo, logger)

	ctx := context.Background()
	filter := &services.StatisticsFilter{}
	repoFilter := repositories.StatisticsFilter{}

	mappingStats := &repositories.MappingStatistics{
		TotalRecords:    100,
		MappedRecords:   75,
		UnmappedRecords: 25,
		MappingRate:     0.75,
	}

	mockRepo.On("GetMappingStatistics", ctx, repoFilter).Return(mappingStats, nil)

	result, err := service.GetMappingStatistics(ctx, filter)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, int64(100), result.TotalRecords)
	assert.Equal(t, int64(75), result.MappedRecords)
	assert.Equal(t, int64(25), result.UnmappedRecords)
	assert.Equal(t, 0.75, result.MappingRate)
	assert.Equal(t, "75.00%", result.MappingRatePercentage)

	mockRepo.AssertExpectations(t)
}

func TestStatisticsService_HealthCheck(t *testing.T) {
	t.Run("successful health check", func(t *testing.T) {
		mockRepo := &MockStatisticsRepository{}
		logger := log.New(os.Stdout, "TEST: ", log.LstdFlags)
		service := services.NewStatisticsService(mockRepo, logger)

		ctx := context.Background()

		mockRepo.On("Ping", ctx).Return(nil)

		err := service.HealthCheck(ctx)

		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("failed health check", func(t *testing.T) {
		mockRepo := &MockStatisticsRepository{}
		logger := log.New(os.Stdout, "TEST: ", log.LstdFlags)
		service := services.NewStatisticsService(mockRepo, logger)

		ctx := context.Background()

		mockRepo.On("Ping", ctx).Return(errors.New("connection failed"))

		err := service.HealthCheck(ctx)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "statistics repository ping failed")
		mockRepo.AssertExpectations(t)
	})

	t.Run("nil repository", func(t *testing.T) {
		logger := log.New(os.Stdout, "TEST: ", log.LstdFlags)
		service := services.NewStatisticsService(nil, logger)

		ctx := context.Background()

		err := service.HealthCheck(ctx)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "statistics repository not initialized")
	})
}

func TestStatisticsService_FormatDateRange(t *testing.T) {
	mockRepo := &MockStatisticsRepository{}
	logger := log.New(os.Stdout, "TEST: ", log.LstdFlags)
	service := services.NewStatisticsService(mockRepo, logger)

	ctx := context.Background()

	// Test with nil filter
	mockRepo.On("GetDailyDistribution", ctx, repositories.StatisticsFilter{}).Return([]repositories.DailyStatistic{}, nil)
	result, _ := service.GetDailyStatistics(ctx, nil)
	assert.Equal(t, "All Time", result.DateRange)

	// Test with date range
	dateFrom := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	dateTo := time.Date(2025, 1, 31, 0, 0, 0, 0, time.UTC)
	filter := &services.StatisticsFilter{
		DateFrom: &dateFrom,
		DateTo:   &dateTo,
	}

	repoFilter := repositories.StatisticsFilter{
		DateFrom: filter.DateFrom,
		DateTo:   filter.DateTo,
	}
	mockRepo.On("GetDailyDistribution", ctx, repoFilter).Return([]repositories.DailyStatistic{}, nil)
	result2, _ := service.GetDailyStatistics(ctx, filter)
	assert.Equal(t, "2025-01-01 to 2025-01-31", result2.DateRange)

	// Test with only DateFrom
	filter2 := &services.StatisticsFilter{
		DateFrom: &dateFrom,
	}
	repoFilter2 := repositories.StatisticsFilter{
		DateFrom: filter2.DateFrom,
	}
	mockRepo.On("GetDailyDistribution", ctx, repoFilter2).Return([]repositories.DailyStatistic{}, nil)
	result3, _ := service.GetDailyStatistics(ctx, filter2)
	assert.Equal(t, "2025-01-01 to Present", result3.DateRange)

	// Test with only DateTo
	filter3 := &services.StatisticsFilter{
		DateTo: &dateTo,
	}
	repoFilter3 := repositories.StatisticsFilter{
		DateTo: filter3.DateTo,
	}
	mockRepo.On("GetDailyDistribution", ctx, repoFilter3).Return([]repositories.DailyStatistic{}, nil)
	result4, _ := service.GetDailyStatistics(ctx, filter3)
	assert.Equal(t, "Beginning to 2025-01-31", result4.DateRange)

	mockRepo.AssertExpectations(t)
}
package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"
	"github.com/yhonda-ohishi/etc_meisai/src/repositories"
)

// MockStatisticsRepository is a mock implementation of StatisticsRepository
type MockStatisticsRepository struct {
	mock.Mock
}

// CountRecords mocks the CountRecords method
func (m *MockStatisticsRepository) CountRecords(ctx context.Context, filter repositories.StatisticsFilter) (int64, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).(int64), args.Error(1)
}

// CountUniqueVehicles mocks the CountUniqueVehicles method
func (m *MockStatisticsRepository) CountUniqueVehicles(ctx context.Context, filter repositories.StatisticsFilter) (int64, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).(int64), args.Error(1)
}

// CountUniqueCards mocks the CountUniqueCards method
func (m *MockStatisticsRepository) CountUniqueCards(ctx context.Context, filter repositories.StatisticsFilter) (int64, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).(int64), args.Error(1)
}

// CountUniqueEntranceICs mocks the CountUniqueEntranceICs method
func (m *MockStatisticsRepository) CountUniqueEntranceICs(ctx context.Context, filter repositories.StatisticsFilter) (int64, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).(int64), args.Error(1)
}

// CountUniqueExitICs mocks the CountUniqueExitICs method
func (m *MockStatisticsRepository) CountUniqueExitICs(ctx context.Context, filter repositories.StatisticsFilter) (int64, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).(int64), args.Error(1)
}

// SumTollAmount mocks the SumTollAmount method
func (m *MockStatisticsRepository) SumTollAmount(ctx context.Context, filter repositories.StatisticsFilter) (int64, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).(int64), args.Error(1)
}

// AverageTollAmount mocks the AverageTollAmount method
func (m *MockStatisticsRepository) AverageTollAmount(ctx context.Context, filter repositories.StatisticsFilter) (float64, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).(float64), args.Error(1)
}

// GetTopRoutes mocks the GetTopRoutes method
func (m *MockStatisticsRepository) GetTopRoutes(ctx context.Context, filter repositories.StatisticsFilter, limit int) ([]repositories.RouteStatistic, error) {
	args := m.Called(ctx, filter, limit)
	if routes := args.Get(0); routes != nil {
		return routes.([]repositories.RouteStatistic), args.Error(1)
	}
	return nil, args.Error(1)
}

// GetTopVehicles mocks the GetTopVehicles method
func (m *MockStatisticsRepository) GetTopVehicles(ctx context.Context, filter repositories.StatisticsFilter, limit int) ([]repositories.VehicleStatistic, error) {
	args := m.Called(ctx, filter, limit)
	if vehicles := args.Get(0); vehicles != nil {
		return vehicles.([]repositories.VehicleStatistic), args.Error(1)
	}
	return nil, args.Error(1)
}

// GetTopCards mocks the GetTopCards method
func (m *MockStatisticsRepository) GetTopCards(ctx context.Context, filter repositories.StatisticsFilter, limit int) ([]repositories.CardStatistic, error) {
	args := m.Called(ctx, filter, limit)
	if cards := args.Get(0); cards != nil {
		return cards.([]repositories.CardStatistic), args.Error(1)
	}
	return nil, args.Error(1)
}

// GetHourlyDistribution mocks the GetHourlyDistribution method
func (m *MockStatisticsRepository) GetHourlyDistribution(ctx context.Context, filter repositories.StatisticsFilter) ([]repositories.HourlyStatistic, error) {
	args := m.Called(ctx, filter)
	if hourly := args.Get(0); hourly != nil {
		return hourly.([]repositories.HourlyStatistic), args.Error(1)
	}
	return nil, args.Error(1)
}

// GetDailyDistribution mocks the GetDailyDistribution method
func (m *MockStatisticsRepository) GetDailyDistribution(ctx context.Context, filter repositories.StatisticsFilter) ([]repositories.DailyStatistic, error) {
	args := m.Called(ctx, filter)
	if daily := args.Get(0); daily != nil {
		return daily.([]repositories.DailyStatistic), args.Error(1)
	}
	return nil, args.Error(1)
}

// GetMonthlyDistribution mocks the GetMonthlyDistribution method
func (m *MockStatisticsRepository) GetMonthlyDistribution(ctx context.Context, filter repositories.StatisticsFilter) ([]repositories.MonthlyStatistic, error) {
	args := m.Called(ctx, filter)
	if monthly := args.Get(0); monthly != nil {
		return monthly.([]repositories.MonthlyStatistic), args.Error(1)
	}
	return nil, args.Error(1)
}

// CountMappedRecords mocks the CountMappedRecords method
func (m *MockStatisticsRepository) CountMappedRecords(ctx context.Context, filter repositories.StatisticsFilter) (int64, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).(int64), args.Error(1)
}

// CountUnmappedRecords mocks the CountUnmappedRecords method
func (m *MockStatisticsRepository) CountUnmappedRecords(ctx context.Context, filter repositories.StatisticsFilter) (int64, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).(int64), args.Error(1)
}

// GetMappingStatistics mocks the GetMappingStatistics method
func (m *MockStatisticsRepository) GetMappingStatistics(ctx context.Context, filter repositories.StatisticsFilter) (*repositories.MappingStatistics, error) {
	args := m.Called(ctx, filter)
	if stats := args.Get(0); stats != nil {
		return stats.(*repositories.MappingStatistics), args.Error(1)
	}
	return nil, args.Error(1)
}

// Ping mocks the Ping method
func (m *MockStatisticsRepository) Ping(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}
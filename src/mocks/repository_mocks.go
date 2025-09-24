package mocks

import (
	"context"
	"github.com/stretchr/testify/mock"
	"github.com/yhonda-ohishi/etc_meisai/src/models"
	"github.com/yhonda-ohishi/etc_meisai/src/repositories"
)

// MockGRPCRepository mocks the gRPC repository
type MockGRPCRepository struct {
	mock.Mock
}

func (m *MockGRPCRepository) GetETCMeisaiByID(ctx context.Context, id int64) (*models.ETCMeisai, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.ETCMeisai), args.Error(1)
}

func (m *MockGRPCRepository) ListETCMeisai(ctx context.Context, filter map[string]interface{}) ([]*models.ETCMeisai, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.ETCMeisai), args.Error(1)
}

func (m *MockGRPCRepository) CreateETCMeisai(ctx context.Context, data *models.ETCMeisai) (*models.ETCMeisai, error) {
	args := m.Called(ctx, data)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.ETCMeisai), args.Error(1)
}

func (m *MockGRPCRepository) UpdateETCMeisai(ctx context.Context, id int64, data *models.ETCMeisai) (*models.ETCMeisai, error) {
	args := m.Called(ctx, id, data)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.ETCMeisai), args.Error(1)
}

func (m *MockGRPCRepository) DeleteETCMeisai(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// MockImportRepository mocks the import repository
type MockImportRepository struct {
	mock.Mock
}

func (m *MockImportRepository) CreateImportSession(ctx context.Context, session *models.ImportSession) error {
	args := m.Called(ctx, session)
	return args.Error(0)
}

func (m *MockImportRepository) GetImportSession(ctx context.Context, id string) (*models.ImportSession, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.ImportSession), args.Error(1)
}

func (m *MockImportRepository) UpdateImportSession(ctx context.Context, session *models.ImportSession) error {
	args := m.Called(ctx, session)
	return args.Error(0)
}

func (m *MockImportRepository) ListImportSessions(ctx context.Context, filter map[string]interface{}) ([]*models.ImportSession, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.ImportSession), args.Error(1)
}

// Statistics represents basic statistics data
type Statistics struct {
	TotalRecords int64 `json:"total_records"`
	TotalAmount  int64 `json:"total_amount"`
}

// MockStatisticsRepository mocks the statistics repository
type MockStatisticsRepository struct {
	mock.Mock
}

// CountRecords counts total records with filter
func (m *MockStatisticsRepository) CountRecords(ctx context.Context, filter repositories.StatisticsFilter) (int64, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).(int64), args.Error(1)
}

// CountUniqueVehicles counts unique vehicles
func (m *MockStatisticsRepository) CountUniqueVehicles(ctx context.Context, filter repositories.StatisticsFilter) (int64, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).(int64), args.Error(1)
}

// CountUniqueCards counts unique ETC cards
func (m *MockStatisticsRepository) CountUniqueCards(ctx context.Context, filter repositories.StatisticsFilter) (int64, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).(int64), args.Error(1)
}

// CountUniqueEntranceICs counts unique entrance ICs
func (m *MockStatisticsRepository) CountUniqueEntranceICs(ctx context.Context, filter repositories.StatisticsFilter) (int64, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).(int64), args.Error(1)
}

// CountUniqueExitICs counts unique exit ICs
func (m *MockStatisticsRepository) CountUniqueExitICs(ctx context.Context, filter repositories.StatisticsFilter) (int64, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).(int64), args.Error(1)
}

// SumTollAmount sums total toll amounts
func (m *MockStatisticsRepository) SumTollAmount(ctx context.Context, filter repositories.StatisticsFilter) (int64, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).(int64), args.Error(1)
}

// AverageTollAmount calculates average toll amount
func (m *MockStatisticsRepository) AverageTollAmount(ctx context.Context, filter repositories.StatisticsFilter) (float64, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).(float64), args.Error(1)
}

// GetTopRoutes returns top routes statistics
func (m *MockStatisticsRepository) GetTopRoutes(ctx context.Context, filter repositories.StatisticsFilter, limit int) ([]repositories.RouteStatistic, error) {
	args := m.Called(ctx, filter, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]repositories.RouteStatistic), args.Error(1)
}

// GetTopVehicles returns top vehicles statistics
func (m *MockStatisticsRepository) GetTopVehicles(ctx context.Context, filter repositories.StatisticsFilter, limit int) ([]repositories.VehicleStatistic, error) {
	args := m.Called(ctx, filter, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]repositories.VehicleStatistic), args.Error(1)
}

// GetTopCards returns top cards statistics
func (m *MockStatisticsRepository) GetTopCards(ctx context.Context, filter repositories.StatisticsFilter, limit int) ([]repositories.CardStatistic, error) {
	args := m.Called(ctx, filter, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]repositories.CardStatistic), args.Error(1)
}

// GetHourlyDistribution returns hourly distribution statistics
func (m *MockStatisticsRepository) GetHourlyDistribution(ctx context.Context, filter repositories.StatisticsFilter) ([]repositories.HourlyStatistic, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]repositories.HourlyStatistic), args.Error(1)
}

// GetDailyDistribution returns daily distribution statistics
func (m *MockStatisticsRepository) GetDailyDistribution(ctx context.Context, filter repositories.StatisticsFilter) ([]repositories.DailyStatistic, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]repositories.DailyStatistic), args.Error(1)
}

// GetMonthlyDistribution returns monthly distribution statistics
func (m *MockStatisticsRepository) GetMonthlyDistribution(ctx context.Context, filter repositories.StatisticsFilter) ([]repositories.MonthlyStatistic, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]repositories.MonthlyStatistic), args.Error(1)
}

// CountMappedRecords counts mapped records
func (m *MockStatisticsRepository) CountMappedRecords(ctx context.Context, filter repositories.StatisticsFilter) (int64, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).(int64), args.Error(1)
}

// CountUnmappedRecords counts unmapped records
func (m *MockStatisticsRepository) CountUnmappedRecords(ctx context.Context, filter repositories.StatisticsFilter) (int64, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).(int64), args.Error(1)
}

// GetMappingStatistics returns mapping statistics
func (m *MockStatisticsRepository) GetMappingStatistics(ctx context.Context, filter repositories.StatisticsFilter) (*repositories.MappingStatistics, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*repositories.MappingStatistics), args.Error(1)
}

// Ping checks repository connectivity
func (m *MockStatisticsRepository) Ping(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}
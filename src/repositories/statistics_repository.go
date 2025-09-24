package repositories

import (
	"context"
	"time"
)

// StatisticsRepository defines the interface for statistics data access
type StatisticsRepository interface {
	// Record count operations
	CountRecords(ctx context.Context, filter StatisticsFilter) (int64, error)
	CountUniqueVehicles(ctx context.Context, filter StatisticsFilter) (int64, error)
	CountUniqueCards(ctx context.Context, filter StatisticsFilter) (int64, error)
	CountUniqueEntranceICs(ctx context.Context, filter StatisticsFilter) (int64, error)
	CountUniqueExitICs(ctx context.Context, filter StatisticsFilter) (int64, error)

	// Amount calculations
	SumTollAmount(ctx context.Context, filter StatisticsFilter) (int64, error)
	AverageTollAmount(ctx context.Context, filter StatisticsFilter) (float64, error)

	// Top statistics
	GetTopRoutes(ctx context.Context, filter StatisticsFilter, limit int) ([]RouteStatistic, error)
	GetTopVehicles(ctx context.Context, filter StatisticsFilter, limit int) ([]VehicleStatistic, error)
	GetTopCards(ctx context.Context, filter StatisticsFilter, limit int) ([]CardStatistic, error)

	// Time-based distributions
	GetHourlyDistribution(ctx context.Context, filter StatisticsFilter) ([]HourlyStatistic, error)
	GetDailyDistribution(ctx context.Context, filter StatisticsFilter) ([]DailyStatistic, error)
	GetMonthlyDistribution(ctx context.Context, filter StatisticsFilter) ([]MonthlyStatistic, error)

	// Mapping statistics
	CountMappedRecords(ctx context.Context, filter StatisticsFilter) (int64, error)
	CountUnmappedRecords(ctx context.Context, filter StatisticsFilter) (int64, error)
	GetMappingStatistics(ctx context.Context, filter StatisticsFilter) (*MappingStatistics, error)

	// Health check
	Ping(ctx context.Context) error
}

// StatisticsFilter contains filters for statistics queries
type StatisticsFilter struct {
	DateFrom      *time.Time
	DateTo        *time.Time
	CarNumbers    []string
	ETCNumbers    []string
	ETCNums       []string
	EntranceICs   []string
	ExitICs       []string
	MinTollAmount *int
	MaxTollAmount *int
}

// RouteStatistic represents statistics for a route
type RouteStatistic struct {
	EntranceIC string `json:"entrance_ic"`
	ExitIC     string `json:"exit_ic"`
	Count      int64  `json:"count"`
	TotalAmount int64  `json:"total_amount"`
	AvgAmount  float64 `json:"avg_amount"`
}

// VehicleStatistic represents statistics for a vehicle
type VehicleStatistic struct {
	CarNumber   string  `json:"car_number"`
	Count       int64   `json:"count"`
	TotalAmount int64   `json:"total_amount"`
	AvgAmount   float64 `json:"avg_amount"`
}

// CardStatistic represents statistics for an ETC card
type CardStatistic struct {
	ETCCardNumber string  `json:"etc_card_number"`
	Count         int64   `json:"count"`
	TotalAmount   int64   `json:"total_amount"`
	AvgAmount     float64 `json:"avg_amount"`
}

// HourlyStatistic represents hourly distribution
type HourlyStatistic struct {
	Hour        int     `json:"hour"`
	Count       int64   `json:"count"`
	TotalAmount int64   `json:"total_amount"`
	AvgAmount   float64 `json:"avg_amount"`
}

// DailyStatistic represents daily distribution
type DailyStatistic struct {
	Date        time.Time `json:"date"`
	Count       int64     `json:"count"`
	TotalAmount int64     `json:"total_amount"`
	AvgAmount   float64   `json:"avg_amount"`
}

// MonthlyStatistic represents monthly distribution
type MonthlyStatistic struct {
	Year        int     `json:"year"`
	Month       int     `json:"month"`
	Count       int64   `json:"count"`
	TotalAmount int64   `json:"total_amount"`
	AvgAmount   float64 `json:"avg_amount"`
}

// MappingStatistics represents mapping statistics
type MappingStatistics struct {
	TotalRecords   int64   `json:"total_records"`
	MappedRecords  int64   `json:"mapped_records"`
	UnmappedRecords int64  `json:"unmapped_records"`
	MappingRate    float64 `json:"mapping_rate"`
}
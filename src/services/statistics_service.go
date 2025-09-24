package services

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/yhonda-ohishi/etc_meisai/src/repositories"
)

// StatisticsFilter re-exported from original service
type StatisticsFilter struct {
	DateFrom      *time.Time `json:"date_from,omitempty"`
	DateTo        *time.Time `json:"date_to,omitempty"`
	CarNumbers    []string   `json:"car_numbers,omitempty"`
	ETCNumbers    []string   `json:"etc_numbers,omitempty"`
	ETCNums       []string   `json:"etc_nums,omitempty"`
	EntranceICs   []string   `json:"entrance_ics,omitempty"`
	ExitICs       []string   `json:"exit_ics,omitempty"`
	MinTollAmount *int       `json:"min_toll_amount,omitempty"`
	MaxTollAmount *int       `json:"max_toll_amount,omitempty"`
}

// GeneralStatistics contains general statistics
type GeneralStatistics struct {
	DateRange         string  `json:"date_range"`
	TotalRecords      int64   `json:"total_records"`
	TotalAmount       int64   `json:"total_amount"`
	AverageAmount     float64 `json:"average_amount"`
	UniqueVehicles    int64   `json:"unique_vehicles"`
	UniqueCards       int64   `json:"unique_cards"`
	UniqueEntranceICs int64   `json:"unique_entrance_ics"`
	UniqueExitICs     int64   `json:"unique_exit_ics"`
	TopRoutes         []RouteStatistic    `json:"top_routes"`
	TopVehicles       []VehicleStatistic  `json:"top_vehicles"`
	TopCards          []CardStatistic     `json:"top_cards"`
	HourlyDistribution []HourlyStatistic  `json:"hourly_distribution"`
}

// RouteStatistic contains route usage statistics
type RouteStatistic struct {
	EntranceIC  string  `json:"entrance_ic"`
	ExitIC      string  `json:"exit_ic"`
	Count       int64   `json:"count"`
	TotalAmount int64   `json:"total_amount"`
	AvgAmount   float64 `json:"avg_amount"`
}

// VehicleStatistic contains vehicle usage statistics
type VehicleStatistic struct {
	CarNumber   string  `json:"car_number"`
	Count       int64   `json:"count"`
	TotalAmount int64   `json:"total_amount"`
	AvgAmount   float64 `json:"avg_amount"`
}

// CardStatistic contains ETC card usage statistics
type CardStatistic struct {
	ETCCardNumber string  `json:"etc_card_number"`
	Count         int64   `json:"count"`
	TotalAmount   int64   `json:"total_amount"`
	AvgAmount     float64 `json:"avg_amount"`
}

// HourlyStatistic contains hourly usage statistics
type HourlyStatistic struct {
	Hour        int     `json:"hour"`
	HourLabel   string  `json:"hour_label"`
	Count       int64   `json:"count"`
	TotalAmount int64   `json:"total_amount"`
	AvgAmount   float64 `json:"avg_amount"`
}

// DailyStatistic for service layer
type DailyStatistic struct {
	Date        string  `json:"date"`
	Count       int64   `json:"count"`
	TotalAmount int64   `json:"total_amount"`
	AvgAmount   float64 `json:"avg_amount"`
}

// MonthlyStatistic for service layer
type MonthlyStatistic struct {
	Year        int     `json:"year"`
	Month       int     `json:"month"`
	MonthName   string  `json:"month_name"`
	Count       int64   `json:"count"`
	TotalAmount int64   `json:"total_amount"`
	AvgAmount   float64 `json:"avg_amount"`
}

// Response types
type DailyStatisticsResponse struct {
	DateRange  string           `json:"date_range"`
	Statistics []DailyStatistic `json:"statistics"`
}

type MonthlyStatisticsResponse struct {
	DateRange  string             `json:"date_range"`
	Statistics []MonthlyStatistic `json:"statistics"`
}

type VehicleStatisticsResponse struct {
	DateRange string             `json:"date_range"`
	Vehicles  []VehicleStatistic `json:"vehicles"`
}

type MappingStatisticsResponse struct {
	DateRange             string  `json:"date_range"`
	TotalRecords          int64   `json:"total_records"`
	MappedRecords         int64   `json:"mapped_records"`
	UnmappedRecords       int64   `json:"unmapped_records"`
	MappingRate           float64 `json:"mapping_rate"`
	MappingRatePercentage string  `json:"mapping_rate_percentage"`
}

// StatisticsService handles analytics and statistics using repository pattern
type StatisticsService struct {
	statsRepo repositories.StatisticsRepository
	logger    *log.Logger
}

// NewStatisticsService creates a new statistics service with repository pattern
func NewStatisticsService(statsRepo repositories.StatisticsRepository, logger *log.Logger) *StatisticsService {
	if logger == nil {
		logger = log.New(log.Writer(), "[StatisticsService] ", log.LstdFlags|log.Lshortfile)
	}

	return &StatisticsService{
		statsRepo: statsRepo,
		logger:    logger,
	}
}

// GetGeneralStatistics retrieves general statistics based on filter
func (s *StatisticsService) GetGeneralStatistics(ctx context.Context, filter *StatisticsFilter) (*GeneralStatistics, error) {
	s.logger.Printf("Generating general statistics")

	// Convert service filter to repository filter
	repoFilter := s.toRepoFilter(filter)

	// Get basic counts
	totalRecords, err := s.statsRepo.CountRecords(ctx, repoFilter)
	if err != nil {
		s.logger.Printf("Failed to count records: %v", err)
		return nil, fmt.Errorf("failed to count records: %w", err)
	}

	totalAmount, err := s.statsRepo.SumTollAmount(ctx, repoFilter)
	if err != nil {
		s.logger.Printf("Failed to sum toll amount: %v", err)
		return nil, fmt.Errorf("failed to sum toll amount: %w", err)
	}

	avgAmount, err := s.statsRepo.AverageTollAmount(ctx, repoFilter)
	if err != nil {
		s.logger.Printf("Failed to calculate average toll amount: %v", err)
		return nil, fmt.Errorf("failed to calculate average toll amount: %w", err)
	}

	uniqueVehicles, err := s.statsRepo.CountUniqueVehicles(ctx, repoFilter)
	if err != nil {
		s.logger.Printf("Failed to count unique vehicles: %v", err)
		return nil, fmt.Errorf("failed to count unique vehicles: %w", err)
	}

	uniqueCards, err := s.statsRepo.CountUniqueCards(ctx, repoFilter)
	if err != nil {
		s.logger.Printf("Failed to count unique cards: %v", err)
		return nil, fmt.Errorf("failed to count unique cards: %w", err)
	}

	uniqueEntranceICs, err := s.statsRepo.CountUniqueEntranceICs(ctx, repoFilter)
	if err != nil {
		s.logger.Printf("Failed to count unique entrance ICs: %v", err)
		return nil, fmt.Errorf("failed to count unique entrance ICs: %w", err)
	}

	uniqueExitICs, err := s.statsRepo.CountUniqueExitICs(ctx, repoFilter)
	if err != nil {
		s.logger.Printf("Failed to count unique exit ICs: %v", err)
		return nil, fmt.Errorf("failed to count unique exit ICs: %w", err)
	}

	// Get top statistics
	topRoutes, err := s.statsRepo.GetTopRoutes(ctx, repoFilter, 10)
	if err != nil {
		s.logger.Printf("Failed to get top routes: %v", err)
		return nil, fmt.Errorf("failed to get top routes: %w", err)
	}

	topVehicles, err := s.statsRepo.GetTopVehicles(ctx, repoFilter, 10)
	if err != nil {
		s.logger.Printf("Failed to get top vehicles: %v", err)
		return nil, fmt.Errorf("failed to get top vehicles: %w", err)
	}

	topCards, err := s.statsRepo.GetTopCards(ctx, repoFilter, 10)
	if err != nil {
		s.logger.Printf("Failed to get top cards: %v", err)
		return nil, fmt.Errorf("failed to get top cards: %w", err)
	}

	hourlyDist, err := s.statsRepo.GetHourlyDistribution(ctx, repoFilter)
	if err != nil {
		s.logger.Printf("Failed to get hourly distribution: %v", err)
		return nil, fmt.Errorf("failed to get hourly distribution: %w", err)
	}

	// Format date range
	dateRange := s.formatDateRange(filter)

	// Convert repository types to service types
	stats := &GeneralStatistics{
		DateRange:          dateRange,
		TotalRecords:       totalRecords,
		TotalAmount:        totalAmount,
		AverageAmount:      avgAmount,
		UniqueVehicles:     uniqueVehicles,
		UniqueCards:        uniqueCards,
		UniqueEntranceICs:  uniqueEntranceICs,
		UniqueExitICs:      uniqueExitICs,
		TopRoutes:          s.convertRouteStats(topRoutes),
		TopVehicles:        s.convertVehicleStats(topVehicles),
		TopCards:           s.convertCardStats(topCards),
		HourlyDistribution: s.convertHourlyStats(hourlyDist),
	}

	s.logger.Printf("Successfully generated general statistics")
	return stats, nil
}

// GetDailyStatistics retrieves daily statistics
func (s *StatisticsService) GetDailyStatistics(ctx context.Context, filter *StatisticsFilter) (*DailyStatisticsResponse, error) {
	s.logger.Printf("Generating daily statistics")

	repoFilter := s.toRepoFilter(filter)
	dailyDist, err := s.statsRepo.GetDailyDistribution(ctx, repoFilter)
	if err != nil {
		s.logger.Printf("Failed to get daily distribution: %v", err)
		return nil, fmt.Errorf("failed to get daily distribution: %w", err)
	}

	// Convert to service response type
	var stats []DailyStatistic
	for _, d := range dailyDist {
		stats = append(stats, DailyStatistic{
			Date:        d.Date.Format("2006-01-02"),
			Count:       d.Count,
			TotalAmount: d.TotalAmount,
			AvgAmount:   d.AvgAmount,
		})
	}

	response := &DailyStatisticsResponse{
		DateRange:  s.formatDateRange(filter),
		Statistics: stats,
	}

	s.logger.Printf("Successfully generated daily statistics")
	return response, nil
}

// GetMonthlyStatistics retrieves monthly statistics
func (s *StatisticsService) GetMonthlyStatistics(ctx context.Context, filter *StatisticsFilter) (*MonthlyStatisticsResponse, error) {
	s.logger.Printf("Generating monthly statistics")

	repoFilter := s.toRepoFilter(filter)
	monthlyDist, err := s.statsRepo.GetMonthlyDistribution(ctx, repoFilter)
	if err != nil {
		s.logger.Printf("Failed to get monthly distribution: %v", err)
		return nil, fmt.Errorf("failed to get monthly distribution: %w", err)
	}

	// Convert to service response type
	var stats []MonthlyStatistic
	for _, m := range monthlyDist {
		stats = append(stats, MonthlyStatistic{
			Year:        m.Year,
			Month:       m.Month,
			MonthName:   time.Month(m.Month).String(),
			Count:       m.Count,
			TotalAmount: m.TotalAmount,
			AvgAmount:   m.AvgAmount,
		})
	}

	response := &MonthlyStatisticsResponse{
		DateRange:  s.formatDateRange(filter),
		Statistics: stats,
	}

	s.logger.Printf("Successfully generated monthly statistics")
	return response, nil
}

// GetVehicleStatistics retrieves statistics for specific vehicles
func (s *StatisticsService) GetVehicleStatistics(ctx context.Context, carNumbers []string, filter *StatisticsFilter) (*VehicleStatisticsResponse, error) {
	s.logger.Printf("Generating vehicle statistics for %d vehicles", len(carNumbers))

	// Add car numbers to filter
	if filter == nil {
		filter = &StatisticsFilter{}
	}
	filter.CarNumbers = carNumbers

	repoFilter := s.toRepoFilter(filter)
	vehicleStats, err := s.statsRepo.GetTopVehicles(ctx, repoFilter, len(carNumbers))
	if err != nil {
		s.logger.Printf("Failed to get vehicle statistics: %v", err)
		return nil, fmt.Errorf("failed to get vehicle statistics: %w", err)
	}

	response := &VehicleStatisticsResponse{
		DateRange: s.formatDateRange(filter),
		Vehicles:  s.convertVehicleStats(vehicleStats),
	}

	s.logger.Printf("Successfully generated vehicle statistics")
	return response, nil
}

// GetMappingStatistics retrieves mapping statistics
func (s *StatisticsService) GetMappingStatistics(ctx context.Context, filter *StatisticsFilter) (*MappingStatisticsResponse, error) {
	s.logger.Printf("Generating mapping statistics")

	repoFilter := s.toRepoFilter(filter)
	mappingStats, err := s.statsRepo.GetMappingStatistics(ctx, repoFilter)
	if err != nil {
		s.logger.Printf("Failed to get mapping statistics: %v", err)
		return nil, fmt.Errorf("failed to get mapping statistics: %w", err)
	}

	response := &MappingStatisticsResponse{
		DateRange:       s.formatDateRange(filter),
		TotalRecords:    mappingStats.TotalRecords,
		MappedRecords:   mappingStats.MappedRecords,
		UnmappedRecords: mappingStats.UnmappedRecords,
		MappingRate:     mappingStats.MappingRate,
		MappingRatePercentage: fmt.Sprintf("%.2f%%", mappingStats.MappingRate*100),
	}

	s.logger.Printf("Successfully generated mapping statistics")
	return response, nil
}

// HealthCheck performs health check for the service
func (s *StatisticsService) HealthCheck(ctx context.Context) error {
	if s.statsRepo == nil {
		return fmt.Errorf("statistics repository not initialized")
	}

	if err := s.statsRepo.Ping(ctx); err != nil {
		return fmt.Errorf("statistics repository ping failed: %w", err)
	}

	return nil
}

// Helper methods

func (s *StatisticsService) toRepoFilter(filter *StatisticsFilter) repositories.StatisticsFilter {
	if filter == nil {
		return repositories.StatisticsFilter{}
	}

	return repositories.StatisticsFilter{
		DateFrom:      filter.DateFrom,
		DateTo:        filter.DateTo,
		CarNumbers:    filter.CarNumbers,
		ETCNumbers:    filter.ETCNumbers,
		ETCNums:       filter.ETCNums,
		EntranceICs:   filter.EntranceICs,
		ExitICs:       filter.ExitICs,
		MinTollAmount: filter.MinTollAmount,
		MaxTollAmount: filter.MaxTollAmount,
	}
}

func (s *StatisticsService) formatDateRange(filter *StatisticsFilter) string {
	if filter == nil || (filter.DateFrom == nil && filter.DateTo == nil) {
		return "All Time"
	}

	var from, to string
	if filter.DateFrom != nil {
		from = filter.DateFrom.Format("2006-01-02")
	} else {
		from = "Beginning"
	}

	if filter.DateTo != nil {
		to = filter.DateTo.Format("2006-01-02")
	} else {
		to = "Present"
	}

	return fmt.Sprintf("%s to %s", from, to)
}

func (s *StatisticsService) convertRouteStats(routes []repositories.RouteStatistic) []RouteStatistic {
	var result []RouteStatistic
	for _, r := range routes {
		result = append(result, RouteStatistic{
			EntranceIC:  r.EntranceIC,
			ExitIC:      r.ExitIC,
			Count:       r.Count,
			TotalAmount: r.TotalAmount,
			AvgAmount:   r.AvgAmount,
		})
	}
	return result
}

func (s *StatisticsService) convertVehicleStats(vehicles []repositories.VehicleStatistic) []VehicleStatistic {
	var result []VehicleStatistic
	for _, v := range vehicles {
		result = append(result, VehicleStatistic{
			CarNumber:   v.CarNumber,
			Count:       v.Count,
			TotalAmount: v.TotalAmount,
			AvgAmount:   v.AvgAmount,
		})
	}
	return result
}

func (s *StatisticsService) convertCardStats(cards []repositories.CardStatistic) []CardStatistic {
	var result []CardStatistic
	for _, c := range cards {
		result = append(result, CardStatistic{
			ETCCardNumber: c.ETCCardNumber,
			Count:         c.Count,
			TotalAmount:   c.TotalAmount,
			AvgAmount:     c.AvgAmount,
		})
	}
	return result
}

func (s *StatisticsService) convertHourlyStats(hourly []repositories.HourlyStatistic) []HourlyStatistic {
	var result []HourlyStatistic
	for _, h := range hourly {
		result = append(result, HourlyStatistic{
			Hour:        h.Hour,
			HourLabel:   fmt.Sprintf("%02d:00", h.Hour),
			Count:       h.Count,
			TotalAmount: h.TotalAmount,
			AvgAmount:   h.AvgAmount,
		})
	}
	return result
}
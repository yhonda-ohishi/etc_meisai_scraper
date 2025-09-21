package services

import (
	"context"
	"fmt"
	"log"
	"time"

	"gorm.io/gorm"

	"github.com/yhonda-ohishi/etc_meisai/src/models"
)

// StatisticsService handles analytics and statistics for ETC data
type StatisticsService struct {
	db     *gorm.DB
	logger *log.Logger
}

// NewStatisticsService creates a new statistics service
func NewStatisticsService(db *gorm.DB, logger *log.Logger) *StatisticsService {
	if logger == nil {
		logger = log.New(log.Writer(), "[StatisticsService] ", log.LstdFlags|log.Lshortfile)
	}

	return &StatisticsService{
		db:     db,
		logger: logger,
	}
}

// StatisticsFilter contains filters for statistics queries
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

// DailyStatistics contains daily aggregated statistics
type DailyStatistics struct {
	Date           time.Time `json:"date"`
	TotalRecords   int64     `json:"total_records"`
	TotalAmount    int64     `json:"total_amount"`
	AverageAmount  float64   `json:"average_amount"`
	UniqueVehicles int64     `json:"unique_vehicles"`
	UniqueCards    int64     `json:"unique_cards"`
	PeakHour       int       `json:"peak_hour"`
	PeakHourCount  int64     `json:"peak_hour_count"`
}

// ICStatistics contains IC (Interchange) usage statistics
type ICStatistics struct {
	ICName        string  `json:"ic_name"`
	Type          string  `json:"type"` // entrance, exit, both
	UsageCount    int64   `json:"usage_count"`
	TotalAmount   int64   `json:"total_amount"`
	AverageAmount float64 `json:"average_amount"`
	UniqueCards   int64   `json:"unique_cards"`
}

// RouteStatistic contains route usage statistics
type RouteStatistic struct {
	EntranceIC    string  `json:"entrance_ic"`
	ExitIC        string  `json:"exit_ic"`
	UsageCount    int64   `json:"usage_count"`
	TotalAmount   int64   `json:"total_amount"`
	AverageAmount float64 `json:"average_amount"`
}

// VehicleStatistic contains vehicle usage statistics
type VehicleStatistic struct {
	CarNumber     string  `json:"car_number"`
	UsageCount    int64   `json:"usage_count"`
	TotalAmount   int64   `json:"total_amount"`
	AverageAmount float64 `json:"average_amount"`
	LastUsed      time.Time `json:"last_used"`
}

// CardStatistic contains ETC card usage statistics
type CardStatistic struct {
	ETCCardNumber string  `json:"etc_card_number"`
	MaskedNumber  string  `json:"masked_number"`
	UsageCount    int64   `json:"usage_count"`
	TotalAmount   int64   `json:"total_amount"`
	AverageAmount float64 `json:"average_amount"`
	LastUsed      time.Time `json:"last_used"`
}

// HourlyStatistic contains hourly usage statistics
type HourlyStatistic struct {
	Hour          int     `json:"hour"`
	UsageCount    int64   `json:"usage_count"`
	TotalAmount   int64   `json:"total_amount"`
	AverageAmount float64 `json:"average_amount"`
}

// GetStatistics retrieves aggregated statistics with filters
func (s *StatisticsService) GetStatistics(ctx context.Context, filter *StatisticsFilter) (*GeneralStatistics, error) {
	s.logger.Printf("Generating statistics with filter: %+v", filter)

	// Build base query
	query := s.buildFilterQuery(s.db.WithContext(ctx), filter)

	// Get basic counts and totals
	var result struct {
		TotalRecords int64 `json:"total_records"`
		TotalAmount  int64 `json:"total_amount"`
	}

	err := query.Model(&models.ETCMeisaiRecord{}).
		Select("COUNT(*) as total_records, COALESCE(SUM(toll_amount), 0) as total_amount").
		Scan(&result).Error
	if err != nil {
		s.logger.Printf("Failed to get basic statistics: %v", err)
		return nil, fmt.Errorf("failed to get basic statistics: %w", err)
	}

	// Calculate average
	var averageAmount float64
	if result.TotalRecords > 0 {
		averageAmount = float64(result.TotalAmount) / float64(result.TotalRecords)
	}

	// Get unique counts
	uniqueVehicles, err := s.CountUniqueVehicles(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to count unique vehicles: %w", err)
	}

	uniqueCards, err := s.CountUniqueCards(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to count unique cards: %w", err)
	}

	// Get unique ICs
	var uniqueEntranceICs, uniqueExitICs int64
	err = query.Model(&models.ETCMeisaiRecord{}).
		Select("COUNT(DISTINCT entrance_ic)").
		Scan(&uniqueEntranceICs).Error
	if err != nil {
		return nil, fmt.Errorf("failed to count unique entrance ICs: %w", err)
	}

	err = query.Model(&models.ETCMeisaiRecord{}).
		Select("COUNT(DISTINCT exit_ic)").
		Scan(&uniqueExitICs).Error
	if err != nil {
		return nil, fmt.Errorf("failed to count unique exit ICs: %w", err)
	}

	// Get top routes
	topRoutes, err := s.getTopRoutes(ctx, filter, 10)
	if err != nil {
		return nil, fmt.Errorf("failed to get top routes: %w", err)
	}

	// Get top vehicles
	topVehicles, err := s.getTopVehicles(ctx, filter, 10)
	if err != nil {
		return nil, fmt.Errorf("failed to get top vehicles: %w", err)
	}

	// Get top cards
	topCards, err := s.getTopCards(ctx, filter, 10)
	if err != nil {
		return nil, fmt.Errorf("failed to get top cards: %w", err)
	}

	// Get hourly distribution
	hourlyDistribution, err := s.getHourlyDistribution(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get hourly distribution: %w", err)
	}

	// Build date range string
	dateRange := "All time"
	if filter.DateFrom != nil || filter.DateTo != nil {
		if filter.DateFrom != nil && filter.DateTo != nil {
			dateRange = fmt.Sprintf("%s to %s", filter.DateFrom.Format("2006-01-02"), filter.DateTo.Format("2006-01-02"))
		} else if filter.DateFrom != nil {
			dateRange = fmt.Sprintf("From %s", filter.DateFrom.Format("2006-01-02"))
		} else {
			dateRange = fmt.Sprintf("Until %s", filter.DateTo.Format("2006-01-02"))
		}
	}

	statistics := &GeneralStatistics{
		DateRange:         dateRange,
		TotalRecords:      result.TotalRecords,
		TotalAmount:       result.TotalAmount,
		AverageAmount:     averageAmount,
		UniqueVehicles:    uniqueVehicles,
		UniqueCards:       uniqueCards,
		UniqueEntranceICs: uniqueEntranceICs,
		UniqueExitICs:     uniqueExitICs,
		TopRoutes:         topRoutes,
		TopVehicles:       topVehicles,
		TopCards:          topCards,
		HourlyDistribution: hourlyDistribution,
	}

	s.logger.Printf("Generated statistics - Records: %d, Amount: %d, Vehicles: %d",
		result.TotalRecords, result.TotalAmount, uniqueVehicles)

	return statistics, nil
}

// GetDailyStatistics retrieves daily statistics for a specific date
func (s *StatisticsService) GetDailyStatistics(ctx context.Context, date time.Time) (*DailyStatistics, error) {
	s.logger.Printf("Generating daily statistics for date: %s", date.Format("2006-01-02"))

	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	endOfDay := startOfDay.Add(24 * time.Hour).Add(-time.Nanosecond)

	query := s.db.WithContext(ctx).Model(&models.ETCMeisaiRecord{}).
		Where("date >= ? AND date <= ?", startOfDay, endOfDay)

	// Get basic statistics
	var result struct {
		TotalRecords int64 `json:"total_records"`
		TotalAmount  int64 `json:"total_amount"`
	}

	err := query.Select("COUNT(*) as total_records, COALESCE(SUM(toll_amount), 0) as total_amount").
		Scan(&result).Error
	if err != nil {
		s.logger.Printf("Failed to get daily statistics: %v", err)
		return nil, fmt.Errorf("failed to get daily statistics: %w", err)
	}

	var averageAmount float64
	if result.TotalRecords > 0 {
		averageAmount = float64(result.TotalAmount) / float64(result.TotalRecords)
	}

	// Get unique counts
	var uniqueVehicles, uniqueCards int64
	query.Select("COUNT(DISTINCT car_number)").Scan(&uniqueVehicles)
	query.Select("COUNT(DISTINCT etc_card_number)").Scan(&uniqueCards)

	// Get peak hour
	var peakResult struct {
		Hour  int   `json:"hour"`
		Count int64 `json:"count"`
	}

	err = s.db.WithContext(ctx).Model(&models.ETCMeisaiRecord{}).
		Select("EXTRACT(HOUR FROM time::time) as hour, COUNT(*) as count").
		Where("date >= ? AND date <= ?", startOfDay, endOfDay).
		Group("EXTRACT(HOUR FROM time::time)").
		Order("count DESC").
		Limit(1).
		Scan(&peakResult).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("failed to get peak hour: %w", err)
	}

	dailyStats := &DailyStatistics{
		Date:           date,
		TotalRecords:   result.TotalRecords,
		TotalAmount:    result.TotalAmount,
		AverageAmount:  averageAmount,
		UniqueVehicles: uniqueVehicles,
		UniqueCards:    uniqueCards,
		PeakHour:       peakResult.Hour,
		PeakHourCount:  peakResult.Count,
	}

	s.logger.Printf("Generated daily statistics for %s - Records: %d, Amount: %d",
		date.Format("2006-01-02"), result.TotalRecords, result.TotalAmount)

	return dailyStats, nil
}

// GetICStatistics retrieves IC usage statistics for a date range
func (s *StatisticsService) GetICStatistics(ctx context.Context, dateFrom, dateTo time.Time) ([]*ICStatistics, error) {
	s.logger.Printf("Generating IC statistics from %s to %s", dateFrom.Format("2006-01-02"), dateTo.Format("2006-01-02"))

	var entranceStats []ICStatistics
	var exitStats []ICStatistics

	// Get entrance IC statistics
	err := s.db.WithContext(ctx).Model(&models.ETCMeisaiRecord{}).
		Select("entrance_ic as ic_name, COUNT(*) as usage_count, SUM(toll_amount) as total_amount, AVG(toll_amount) as average_amount, COUNT(DISTINCT etc_card_number) as unique_cards").
		Where("date >= ? AND date <= ?", dateFrom, dateTo).
		Group("entrance_ic").
		Order("usage_count DESC").
		Scan(&entranceStats).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get entrance IC statistics: %w", err)
	}

	// Get exit IC statistics
	err = s.db.WithContext(ctx).Model(&models.ETCMeisaiRecord{}).
		Select("exit_ic as ic_name, COUNT(*) as usage_count, SUM(toll_amount) as total_amount, AVG(toll_amount) as average_amount, COUNT(DISTINCT etc_card_number) as unique_cards").
		Where("date >= ? AND date <= ?", dateFrom, dateTo).
		Group("exit_ic").
		Order("usage_count DESC").
		Scan(&exitStats).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get exit IC statistics: %w", err)
	}

	// Combine and deduplicate IC statistics
	icMap := make(map[string]*ICStatistics)

	// Process entrance statistics
	for _, stat := range entranceStats {
		if existing, exists := icMap[stat.ICName]; exists {
			existing.Type = "both"
			existing.UsageCount += stat.UsageCount
			existing.TotalAmount += stat.TotalAmount
			existing.AverageAmount = float64(existing.TotalAmount) / float64(existing.UsageCount)
			if stat.UniqueCards > existing.UniqueCards {
				existing.UniqueCards = stat.UniqueCards
			}
		} else {
			icMap[stat.ICName] = &ICStatistics{
				ICName:        stat.ICName,
				Type:          "entrance",
				UsageCount:    stat.UsageCount,
				TotalAmount:   stat.TotalAmount,
				AverageAmount: stat.AverageAmount,
				UniqueCards:   stat.UniqueCards,
			}
		}
	}

	// Process exit statistics
	for _, stat := range exitStats {
		if existing, exists := icMap[stat.ICName]; exists {
			if existing.Type == "entrance" {
				existing.Type = "both"
			}
			existing.UsageCount += stat.UsageCount
			existing.TotalAmount += stat.TotalAmount
			existing.AverageAmount = float64(existing.TotalAmount) / float64(existing.UsageCount)
			if stat.UniqueCards > existing.UniqueCards {
				existing.UniqueCards = stat.UniqueCards
			}
		} else {
			icMap[stat.ICName] = &ICStatistics{
				ICName:        stat.ICName,
				Type:          "exit",
				UsageCount:    stat.UsageCount,
				TotalAmount:   stat.TotalAmount,
				AverageAmount: stat.AverageAmount,
				UniqueCards:   stat.UniqueCards,
			}
		}
	}

	// Convert map to slice
	var result []*ICStatistics
	for _, stat := range icMap {
		result = append(result, stat)
	}

	s.logger.Printf("Generated IC statistics for %d unique ICs", len(result))
	return result, nil
}

// CalculateTotalAmount calculates total toll amount from records
func (s *StatisticsService) CalculateTotalAmount(ctx context.Context, filter *StatisticsFilter) (int64, error) {
	query := s.buildFilterQuery(s.db.WithContext(ctx), filter)

	var totalAmount int64
	err := query.Model(&models.ETCMeisaiRecord{}).
		Select("COALESCE(SUM(toll_amount), 0)").
		Scan(&totalAmount).Error
	if err != nil {
		s.logger.Printf("Failed to calculate total amount: %v", err)
		return 0, fmt.Errorf("failed to calculate total amount: %w", err)
	}

	return totalAmount, nil
}

// CountUniqueVehicles counts unique vehicles from records
func (s *StatisticsService) CountUniqueVehicles(ctx context.Context, filter *StatisticsFilter) (int64, error) {
	query := s.buildFilterQuery(s.db.WithContext(ctx), filter)

	var count int64
	err := query.Model(&models.ETCMeisaiRecord{}).
		Select("COUNT(DISTINCT car_number)").
		Scan(&count).Error
	if err != nil {
		s.logger.Printf("Failed to count unique vehicles: %v", err)
		return 0, fmt.Errorf("failed to count unique vehicles: %w", err)
	}

	return count, nil
}

// CountUniqueCards counts unique ETC cards from records
func (s *StatisticsService) CountUniqueCards(ctx context.Context, filter *StatisticsFilter) (int64, error) {
	query := s.buildFilterQuery(s.db.WithContext(ctx), filter)

	var count int64
	err := query.Model(&models.ETCMeisaiRecord{}).
		Select("COUNT(DISTINCT etc_card_number)").
		Scan(&count).Error
	if err != nil {
		s.logger.Printf("Failed to count unique cards: %v", err)
		return 0, fmt.Errorf("failed to count unique cards: %w", err)
	}

	return count, nil
}

// buildFilterQuery builds a GORM query with the given filters
func (s *StatisticsService) buildFilterQuery(query *gorm.DB, filter *StatisticsFilter) *gorm.DB {
	if filter == nil {
		return query
	}

	if filter.DateFrom != nil {
		query = query.Where("date >= ?", *filter.DateFrom)
	}
	if filter.DateTo != nil {
		query = query.Where("date <= ?", *filter.DateTo)
	}
	if len(filter.CarNumbers) > 0 {
		query = query.Where("car_number IN ?", filter.CarNumbers)
	}
	if len(filter.ETCNumbers) > 0 {
		query = query.Where("etc_card_number IN ?", filter.ETCNumbers)
	}
	if len(filter.ETCNums) > 0 {
		query = query.Where("etc_num IN ?", filter.ETCNums)
	}
	if len(filter.EntranceICs) > 0 {
		query = query.Where("entrance_ic IN ?", filter.EntranceICs)
	}
	if len(filter.ExitICs) > 0 {
		query = query.Where("exit_ic IN ?", filter.ExitICs)
	}
	if filter.MinTollAmount != nil {
		query = query.Where("toll_amount >= ?", *filter.MinTollAmount)
	}
	if filter.MaxTollAmount != nil {
		query = query.Where("toll_amount <= ?", *filter.MaxTollAmount)
	}

	return query
}

// getTopRoutes gets the top routes by usage
func (s *StatisticsService) getTopRoutes(ctx context.Context, filter *StatisticsFilter, limit int) ([]RouteStatistic, error) {
	query := s.buildFilterQuery(s.db.WithContext(ctx), filter)

	var routes []RouteStatistic
	err := query.Model(&models.ETCMeisaiRecord{}).
		Select("entrance_ic, exit_ic, COUNT(*) as usage_count, SUM(toll_amount) as total_amount, AVG(toll_amount) as average_amount").
		Group("entrance_ic, exit_ic").
		Order("usage_count DESC").
		Limit(limit).
		Scan(&routes).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get top routes: %w", err)
	}

	return routes, nil
}

// getTopVehicles gets the top vehicles by usage
func (s *StatisticsService) getTopVehicles(ctx context.Context, filter *StatisticsFilter, limit int) ([]VehicleStatistic, error) {
	query := s.buildFilterQuery(s.db.WithContext(ctx), filter)

	var vehicles []VehicleStatistic
	err := query.Model(&models.ETCMeisaiRecord{}).
		Select("car_number, COUNT(*) as usage_count, SUM(toll_amount) as total_amount, AVG(toll_amount) as average_amount, MAX(date) as last_used").
		Group("car_number").
		Order("usage_count DESC").
		Limit(limit).
		Scan(&vehicles).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get top vehicles: %w", err)
	}

	return vehicles, nil
}

// getTopCards gets the top ETC cards by usage
func (s *StatisticsService) getTopCards(ctx context.Context, filter *StatisticsFilter, limit int) ([]CardStatistic, error) {
	query := s.buildFilterQuery(s.db.WithContext(ctx), filter)

	var cards []CardStatistic
	err := query.Model(&models.ETCMeisaiRecord{}).
		Select("etc_card_number, COUNT(*) as usage_count, SUM(toll_amount) as total_amount, AVG(toll_amount) as average_amount, MAX(date) as last_used").
		Group("etc_card_number").
		Order("usage_count DESC").
		Limit(limit).
		Scan(&cards).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get top cards: %w", err)
	}

	// Add masked numbers
	for i := range cards {
		cards[i].MaskedNumber = s.maskCardNumber(cards[i].ETCCardNumber)
	}

	return cards, nil
}

// getHourlyDistribution gets the hourly usage distribution
func (s *StatisticsService) getHourlyDistribution(ctx context.Context, filter *StatisticsFilter) ([]HourlyStatistic, error) {
	query := s.buildFilterQuery(s.db.WithContext(ctx), filter)

	var hourlyStats []HourlyStatistic
	err := query.Model(&models.ETCMeisaiRecord{}).
		Select("EXTRACT(HOUR FROM time::time) as hour, COUNT(*) as usage_count, SUM(toll_amount) as total_amount, AVG(toll_amount) as average_amount").
		Group("EXTRACT(HOUR FROM time::time)").
		Order("hour").
		Scan(&hourlyStats).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get hourly distribution: %w", err)
	}

	return hourlyStats, nil
}

// maskCardNumber creates a masked version of the ETC card number
func (s *StatisticsService) maskCardNumber(cardNumber string) string {
	if len(cardNumber) <= 4 {
		return "****"
	}
	return "****-****-****-" + cardNumber[len(cardNumber)-4:]
}

// HealthCheck performs health check for the service
func (s *StatisticsService) HealthCheck(ctx context.Context) error {
	// Check database connectivity
	sqlDB, err := s.db.DB()
	if err != nil {
		return fmt.Errorf("failed to get database connection: %w", err)
	}

	if err := sqlDB.PingContext(ctx); err != nil {
		return fmt.Errorf("database ping failed: %w", err)
	}

	return nil
}
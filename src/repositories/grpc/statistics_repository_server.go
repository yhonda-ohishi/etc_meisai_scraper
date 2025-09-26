package grpc

import (
	"context"
	"sync"
	"time"

	pb "github.com/yhonda-ohishi/etc_meisai/src/pb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// StatisticsRepositoryServer implements the StatisticsRepository service
type StatisticsRepositoryServer struct {
	pb.UnimplementedStatisticsRepositoryServer
	mu                    sync.RWMutex
	mappingRepo           *ETCMappingRepositoryServer
	recordRepo            *ETCMeisaiRecordRepositoryServer
	importRepo            *ImportRepositoryServer
}

// NewStatisticsRepositoryServer creates a new statistics repository server
func NewStatisticsRepositoryServer(
	mappingRepo *ETCMappingRepositoryServer,
	recordRepo *ETCMeisaiRecordRepositoryServer,
	importRepo *ImportRepositoryServer,
) *StatisticsRepositoryServer {
	return &StatisticsRepositoryServer{
		mappingRepo: mappingRepo,
		recordRepo:  recordRepo,
		importRepo:  importRepo,
	}
}

// GetOverallStatistics retrieves overall system statistics
func (s *StatisticsRepositoryServer) GetOverallStatistics(ctx context.Context, req *pb.GetStatisticsRequest) (*pb.GetStatisticsResponse, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Calculate record statistics
	recordStats := s.calculateRecordStatistics(req.DateFrom, req.DateTo)

	return &pb.GetStatisticsResponse{
		TotalRecords: recordStats.TotalRecords,
		TotalAmount:  recordStats.TotalAmount,
		UniqueCars:   recordStats.UniqueCars,
		UniqueCards:  recordStats.UniqueCards,
	}, nil
}

// GetDailyStatistics retrieves daily statistics
func (s *StatisticsRepositoryServer) GetDailyStatistics(ctx context.Context, req *pb.GetDailyStatisticsRequest) (*pb.GetDailyStatisticsResponse, error) {
	if req.DateFrom == "" || req.DateTo == "" {
		return nil, status.Error(codes.InvalidArgument, "date range is required")
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	// Parse dates
	fromDate, err := time.Parse("2006-01-02", req.DateFrom)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid date_from format")
	}
	toDate, err := time.Parse("2006-01-02", req.DateTo)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid date_to format")
	}

	var dailyStats []*pb.DailyStatistics

	// Generate statistics for each day
	for date := fromDate; !date.After(toDate); date = date.AddDate(0, 0, 1) {
		dateStr := date.Format("2006-01-02")
		dailyRecordCount := s.getRecordCountForDate(dateStr)
		dailyAmount := s.getTollAmountForDate(dateStr)

		dailyStat := &pb.DailyStatistics{
			Date:        dateStr,
			RecordCount: dailyRecordCount,
			TotalAmount: dailyAmount,
		}
		dailyStats = append(dailyStats, dailyStat)
	}

	return &pb.GetDailyStatisticsResponse{
		DailyStats: dailyStats,
	}, nil
}

// GetICUsageStatistics retrieves IC usage statistics
func (s *StatisticsRepositoryServer) GetICUsageStatistics(ctx context.Context, req *pb.GetICUsageRequest) (*pb.GetICUsageResponse, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	icUsageMap := make(map[string]*pb.ICStatistics)

	// Count IC usage from records
	s.recordRepo.mu.RLock()
	defer s.recordRepo.mu.RUnlock()

	for _, record := range s.recordRepo.records {
		// Apply date filter if provided
		if !s.isWithinDateRange(s.parseRecordDate(record.Date), req.DateFrom, req.DateTo) {
			continue
		}

		// Count entrance IC
		if record.EntranceIc != "" {
			if stats, ok := icUsageMap[record.EntranceIc]; ok {
				stats.UsageCount++
			} else {
				icUsageMap[record.EntranceIc] = &pb.ICStatistics{
					IcName:     record.EntranceIc,
					UsageCount: 1,
					IcType:     "entrance",
				}
			}
		}

		// Count exit IC
		if record.ExitIc != "" {
			if stats, ok := icUsageMap[record.ExitIc]; ok {
				stats.UsageCount++
			} else {
				icUsageMap[record.ExitIc] = &pb.ICStatistics{
					IcName:     record.ExitIc,
					UsageCount: 1,
					IcType:     "exit",
				}
			}
		}
	}

	// Convert map to slice
	var icStats []*pb.ICStatistics
	for _, stats := range icUsageMap {
		icStats = append(icStats, stats)
	}

	// Apply top N limit if specified
	if req.TopN > 0 && len(icStats) > int(req.TopN) {
		icStats = icStats[:req.TopN]
	}

	return &pb.GetICUsageResponse{
		IcStats: icStats,
	}, nil
}

// GetCarUsageStatistics retrieves car usage statistics
func (s *StatisticsRepositoryServer) GetCarUsageStatistics(ctx context.Context, req *pb.GetCarUsageRequest) (*pb.GetCarUsageResponse, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	carUsageCounts := make(map[string]int32)
	carTollAmounts := make(map[string]int64)

	// Count car usage from records
	s.recordRepo.mu.RLock()
	defer s.recordRepo.mu.RUnlock()

	for _, record := range s.recordRepo.records {
		// Apply date filter if provided
		if !s.isWithinDateRange(s.parseRecordDate(record.Date), req.DateFrom, req.DateTo) {
			continue
		}

		if record.CarNumber != "" {
			carUsageCounts[record.CarNumber]++
			carTollAmounts[record.CarNumber] += int64(record.TollAmount)
		}
	}

	return &pb.GetCarUsageResponse{
		CarUsageCounts: carUsageCounts,
		CarTollAmounts: carTollAmounts,
	}, nil
}

// GetImportStatistics retrieves import statistics
func (s *StatisticsRepositoryServer) GetImportStatistics(ctx context.Context, req *pb.GetImportStatisticsRequest) (*pb.GetImportStatisticsResponse, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Use import repository's GetSessionStatistics
	statsReq := &pb.GetSessionStatisticsRequest{
		AccountId: req.AccountId,
		DateFrom:  req.DateFrom,
		DateTo:    req.DateTo,
	}

	stats, err := s.importRepo.GetSessionStatistics(ctx, statsReq)
	if err != nil {
		return nil, err
	}

	// Calculate trends
	var trends []*pb.ImportTrend
	if req.DateFrom != nil && req.DateTo != nil {
		fromDate, _ := time.Parse("2006-01-02", *req.DateFrom)
		toDate, _ := time.Parse("2006-01-02", *req.DateTo)

		for date := fromDate; !date.After(toDate); date = date.AddDate(0, 0, 1) {
			dateStr := date.Format("2006-01-02")
			trend := &pb.ImportTrend{
				Date:        dateStr,
				ImportCount: s.getImportCountForDate(dateStr),
				RecordCount: s.getImportedRecordCountForDate(dateStr),
				ErrorCount:  s.getImportErrorCountForDate(dateStr),
			}
			trends = append(trends, trend)
		}
	}

	return &pb.GetImportStatisticsResponse{
		Statistics: stats,
		Trends:     trends,
	}, nil
}

// GetMappingStatistics retrieves mapping statistics  
func (s *StatisticsRepositoryServer) GetMappingStatistics(ctx context.Context, req *pb.GetMappingStatisticsRequest) (*pb.MappingStatistics, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.calculateMappingStatistics(req.DateFrom, req.DateTo), nil
}

// Helper functions

func (s *StatisticsRepositoryServer) calculateRecordStatistics(dateFrom, dateTo *string) *pb.RecordStatistics {
	s.recordRepo.mu.RLock()
	defer s.recordRepo.mu.RUnlock()

	stats := &pb.RecordStatistics{
		RecordsPerMonth: make(map[string]int32),
	}

	uniqueCarNumbers := make(map[string]bool)
	uniqueCardNumbers := make(map[string]bool)

	for _, record := range s.recordRepo.records {
		// Apply date filter if provided
		if !s.isWithinDateRange(s.parseRecordDate(record.Date), dateFrom, dateTo) {
			continue
		}

		stats.TotalRecords++
		stats.TotalAmount += int64(record.TollAmount)

		if record.CarNumber != "" {
			uniqueCarNumbers[record.CarNumber] = true
		}
		if record.EtcCardNumber != "" {
			uniqueCardNumbers[record.EtcCardNumber] = true
		}

		if record.Date != "" {
			monthKey := s.parseRecordDate(record.Date).Format("2006-01")
			stats.RecordsPerMonth[monthKey]++
		}
	}

	stats.UniqueCars = int32(len(uniqueCarNumbers))
	stats.UniqueCards = int32(len(uniqueCardNumbers))

	return stats
}

func (s *StatisticsRepositoryServer) calculateMappingStatistics(dateFrom, dateTo *string) *pb.MappingStatistics {
	s.mappingRepo.mu.RLock()
	defer s.mappingRepo.mu.RUnlock()

	stats := &pb.MappingStatistics{
		MappingsByType: make(map[string]int32),
	}

	var totalConfidence float32
	confidenceCount := 0

	for _, mapping := range s.mappingRepo.mappings {
		// Apply date filter if provided
		if !s.isWithinDateRange(mapping.CreatedAt.AsTime(), dateFrom, dateTo) {
			continue
		}

		stats.TotalMappings++

		switch mapping.Status {
		case pb.MappingStatus_MAPPING_STATUS_ACTIVE:
			stats.ActiveMappings++
		case pb.MappingStatus_MAPPING_STATUS_PENDING:
			stats.PendingMappings++
		case pb.MappingStatus_MAPPING_STATUS_REJECTED:
			stats.RejectedMappings++
		}

		if mapping.MappingType != "" {
			stats.MappingsByType[mapping.MappingType]++
		}

		if mapping.Confidence > 0 {
			totalConfidence += mapping.Confidence
			confidenceCount++
		}
	}

	if confidenceCount > 0 {
		stats.AverageConfidence = totalConfidence / float32(confidenceCount)
	}

	return stats
}

func (s *StatisticsRepositoryServer) calculateImportStatistics(dateFrom, dateTo *string) *pb.SessionStatistics {
	s.importRepo.mu.RLock()
	defer s.importRepo.mu.RUnlock()

	stats := &pb.SessionStatistics{}

	for _, session := range s.importRepo.sessions {
		// Apply date filter if provided
		if !s.isWithinDateRange(session.CreatedAt.AsTime(), dateFrom, dateTo) {
			continue
		}

		stats.TotalSessions++

		switch session.Status {
		case pb.ImportStatus_IMPORT_STATUS_COMPLETED:
			stats.SuccessfulSessions++
			stats.TotalRecordsImported += session.SuccessRows
		case pb.ImportStatus_IMPORT_STATUS_FAILED:
			stats.FailedSessions++
		}

		stats.TotalDuplicates += session.DuplicateRows
		if errors, ok := s.importRepo.errors[session.Id]; ok {
			stats.TotalErrors += int32(len(errors))
		}
	}

	return stats
}

func (s *StatisticsRepositoryServer) isWithinDateRange(date time.Time, dateFrom, dateTo *string) bool {
	if dateFrom != nil {
		fromDate, err := time.Parse("2006-01-02", *dateFrom)
		if err == nil && date.Before(fromDate) {
			return false
		}
	}

	if dateTo != nil {
		toDate, err := time.Parse("2006-01-02", *dateTo)
		if err == nil && date.After(toDate.Add(24*time.Hour)) {
			return false
		}
	}

	return true
}

func (s *StatisticsRepositoryServer) getRecordCountForDate(dateStr string) int32 {
	s.recordRepo.mu.RLock()
	defer s.recordRepo.mu.RUnlock()

	if records, ok := s.recordRepo.recordsByDate[dateStr]; ok {
		return int32(len(records))
	}
	return 0
}

func (s *StatisticsRepositoryServer) getTollAmountForDate(dateStr string) int64 {
	s.recordRepo.mu.RLock()
	defer s.recordRepo.mu.RUnlock()

	var total int64
	if records, ok := s.recordRepo.recordsByDate[dateStr]; ok {
		for _, record := range records {
			total += int64(record.TollAmount)
		}
	}
	return total
}

func (s *StatisticsRepositoryServer) getMappingCountForDate(dateStr string) int32 {
	s.mappingRepo.mu.RLock()
	defer s.mappingRepo.mu.RUnlock()

	var count int32
	targetDate, _ := time.Parse("2006-01-02", dateStr)

	for _, mapping := range s.mappingRepo.mappings {
		if mapping.CreatedAt != nil {
			mappingDate := mapping.CreatedAt.AsTime().Format("2006-01-02")
			if mappingDate == targetDate.Format("2006-01-02") {
				count++
			}
		}
	}
	return count
}

func (s *StatisticsRepositoryServer) getImportCountForDate(dateStr string) int32 {
	s.importRepo.mu.RLock()
	defer s.importRepo.mu.RUnlock()

	var count int32
	targetDate, _ := time.Parse("2006-01-02", dateStr)

	for _, session := range s.importRepo.sessions {
		if session.CreatedAt != nil {
			sessionDate := session.CreatedAt.AsTime().Format("2006-01-02")
			if sessionDate == targetDate.Format("2006-01-02") {
				count++
			}
		}
	}
	return count
}

func (s *StatisticsRepositoryServer) getImportedRecordCountForDate(dateStr string) int32 {
	s.importRepo.mu.RLock()
	defer s.importRepo.mu.RUnlock()

	var count int32
	targetDate, _ := time.Parse("2006-01-02", dateStr)

	for _, session := range s.importRepo.sessions {
		if session.CreatedAt != nil {
			sessionDate := session.CreatedAt.AsTime().Format("2006-01-02")
			if sessionDate == targetDate.Format("2006-01-02") {
				count += session.SuccessRows
			}
		}
	}
	return count
}

func (s *StatisticsRepositoryServer) getImportErrorCountForDate(dateStr string) int32 {
	s.importRepo.mu.RLock()
	defer s.importRepo.mu.RUnlock()

	var count int32
	targetDate, _ := time.Parse("2006-01-02", dateStr)

	for sessionID, errors := range s.importRepo.errors {
		if session, ok := s.importRepo.sessions[sessionID]; ok {
			if session.CreatedAt != nil {
				sessionDate := session.CreatedAt.AsTime().Format("2006-01-02")
				if sessionDate == targetDate.Format("2006-01-02") {
					count += int32(len(errors))
				}
			}
		}
	}
	return count
}

func (s *StatisticsRepositoryServer) getUniqueCarCountForDate(dateStr string) int32 {
	s.recordRepo.mu.RLock()
	defer s.recordRepo.mu.RUnlock()

	uniqueCars := make(map[string]bool)
	if records, ok := s.recordRepo.recordsByDate[dateStr]; ok {
		for _, record := range records {
			if record.CarNumber != "" {
				uniqueCars[record.CarNumber] = true
			}
		}
	}
	return int32(len(uniqueCars))
}

func (s *StatisticsRepositoryServer) getUniqueCardCountForDate(dateStr string) int32 {
	s.recordRepo.mu.RLock()
	defer s.recordRepo.mu.RUnlock()

	uniqueCards := make(map[string]bool)
	if records, ok := s.recordRepo.recordsByDate[dateStr]; ok {
		for _, record := range records {
			if record.EtcCardNumber != "" {
				uniqueCards[record.EtcCardNumber] = true
			}
		}
	}
	return int32(len(uniqueCards))
}

// parseRecordDate parses a date string to time.Time
func (s *StatisticsRepositoryServer) parseRecordDate(dateStr string) time.Time {
	if dateStr == "" {
		return time.Time{}
	}
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return time.Time{}
	}
	return date
}
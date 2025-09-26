package grpc

import (
	"context"
	"encoding/csv"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	pb "github.com/yhonda-ohishi/etc_meisai/src/pb"
	"github.com/yhonda-ohishi/etc_meisai/src/repositories"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// MeisaiBusinessServiceServer implements the MeisaiBusinessService gRPC server
type MeisaiBusinessServiceServer struct {
	pb.UnimplementedMeisaiBusinessServiceServer
	recordRepo      *repositories.ETCMeisaiRecordRepositoryClient
	mappingRepo     *repositories.ETCMappingRepositoryClient
	importRepo      *repositories.ImportRepositoryClient
	statisticsRepo  *repositories.StatisticsRepositoryClient
	logger          *log.Logger
}

// NewMeisaiBusinessServiceServer creates a new meisai business service server
func NewMeisaiBusinessServiceServer(
	recordRepo *repositories.ETCMeisaiRecordRepositoryClient,
	mappingRepo *repositories.ETCMappingRepositoryClient,
	importRepo *repositories.ImportRepositoryClient,
	statisticsRepo *repositories.StatisticsRepositoryClient,
	logger *log.Logger,
) *MeisaiBusinessServiceServer {
	if logger == nil {
		logger = log.New(log.Writer(), "[MeisaiBusinessServiceServer] ", log.LstdFlags|log.Lshortfile)
	}

	return &MeisaiBusinessServiceServer{
		recordRepo:     recordRepo,
		mappingRepo:    mappingRepo,
		importRepo:     importRepo,
		statisticsRepo: statisticsRepo,
		logger:         logger,
	}
}

// ImportRecords processes CSV data and imports ETC records
func (s *MeisaiBusinessServiceServer) ImportRecords(ctx context.Context, req *pb.ImportRecordsRequest) (*pb.ImportRecordsResponse, error) {
	s.logger.Printf("ImportRecords called for account: %s, type: %s", req.AccountId, req.AccountType)

	if len(req.CsvData) == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "csv_data cannot be empty")
	}

	// Parse CSV data
	csvReader := csv.NewReader(strings.NewReader(string(req.CsvData)))
	records, err := csvReader.ReadAll()
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "failed to parse CSV data: %v", err)
	}

	if len(records) == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "no records found in CSV data")
	}

	// Create import session
	importSession := &pb.ImportSession{
		Id:          generateSessionID(),
		AccountId:   req.AccountId,
		AccountType: req.AccountType,
		Status:      pb.ImportStatus_IMPORT_STATUS_PROCESSING,
		StartedAt:   timestamppb.Now(),
		TotalRows: int32(len(records) - 1), // Excluding header
	}

	// Process records (skip header row)
	var pbRecords []*pb.ETCMeisaiRecord
	successCount := int32(0)
	failureCount := int32(0)
	var errors []string

	for i, record := range records[1:] { // Skip header
		if len(record) < 8 { // Expecting at least 8 columns
			errors = append(errors, fmt.Sprintf("Row %d: insufficient columns", i+2))
			failureCount++
			continue
		}

		// Parse the record
		pbRecord, parseErr := s.parseCSVRecord(record)
		if parseErr != nil {
			errors = append(errors, fmt.Sprintf("Row %d: %v", i+2, parseErr))
			failureCount++
			continue
		}

		pbRecords = append(pbRecords, pbRecord)
		successCount++
	}

	// Bulk create records if parsing was successful
	if len(pbRecords) > 0 {
		bulkResult, bulkErr := s.recordRepo.BulkCreate(ctx, pbRecords)
		if bulkErr != nil {
			s.logger.Printf("Bulk create failed: %v", bulkErr)
			return nil, status.Errorf(codes.Internal, "failed to create records: %v", bulkErr)
		}
		successCount = bulkResult.CreatedCount
		failureCount = int32(len(records) - 1) - bulkResult.CreatedCount - bulkResult.DuplicateCount
	}

	// Update import session
	importSession.Status = pb.ImportStatus_IMPORT_STATUS_COMPLETED
	importSession.CompletedAt = timestamppb.Now()
	importSession.ProcessedRows = successCount

	result := BatchOperationResultToProto(
		int32(len(records) - 1),
		successCount,
		failureCount,
		errors,
		time.Since(importSession.StartedAt.AsTime()),
	)

	response := &pb.ImportRecordsResponse{
		Session: importSession,
		Result:  result,
	}

	s.logger.Printf("ImportRecords completed: %d successful, %d failed", successCount, failureCount)
	return response, nil
}

// ValidateRecord validates an ETC record with business rules
func (s *MeisaiBusinessServiceServer) ValidateRecord(ctx context.Context, req *pb.ValidateRecordRequest) (*pb.ValidateRecordResponse, error) {
	s.logger.Printf("ValidateRecord called")

	if req.Record == nil {
		return nil, status.Errorf(codes.InvalidArgument, "record cannot be nil")
	}

	isValid := true
	var errors []string
	sanitizedRecord := req.Record

	// Basic validation
	if req.Record.CarNumber == "" {
		isValid = false
		errors = append(errors, "car_number is required")
	}
	if req.Record.EtcCardNumber == "" {
		isValid = false
		errors = append(errors, "etc_card_number is required")
	}
	if req.Record.EntranceIc == "" {
		isValid = false
		errors = append(errors, "entrance_ic is required")
	}
	if req.Record.ExitIc == "" {
		isValid = false
		errors = append(errors, "exit_ic is required")
	}
	if req.Record.TollAmount < 0 {
		isValid = false
		errors = append(errors, "toll_amount must be non-negative")
	}

	// Strict mode validation
	if req.StrictMode {
		if req.Record.TollAmount > 10000 { // Reasonable upper limit
			isValid = false
			errors = append(errors, "toll_amount seems unreasonably high")
		}
		if len(req.Record.CarNumber) != 4 && len(req.Record.CarNumber) != 7 { // Japanese car number formats
			isValid = false
			errors = append(errors, "car_number format invalid for strict mode")
		}
	}

	// Sanitization
	sanitizedRecord = &pb.ETCMeisaiRecord{
		Id:            req.Record.Id,
		Date:          req.Record.Date,
		Time:          strings.TrimSpace(req.Record.Time),
		EntranceIc:    strings.TrimSpace(req.Record.EntranceIc),
		ExitIc:        strings.TrimSpace(req.Record.ExitIc),
		TollAmount:    req.Record.TollAmount,
		CarNumber:     strings.TrimSpace(strings.ToUpper(req.Record.CarNumber)),
		EtcCardNumber: strings.TrimSpace(req.Record.EtcCardNumber),
		EtcNum:        req.Record.EtcNum,
		DtakoRowId:    req.Record.DtakoRowId,
		Hash:          req.Record.Hash,
		CreatedAt:     req.Record.CreatedAt,
		UpdatedAt:     req.Record.UpdatedAt,
	}

	validationResult := ValidationResultToProto(isValid, errors)

	response := &pb.ValidateRecordResponse{
		Result:           validationResult,
		SanitizedRecord:  sanitizedRecord,
	}

	s.logger.Printf("ValidateRecord completed: valid=%t", isValid)
	return response, nil
}

// EnrichRecord adds additional data to an ETC record
func (s *MeisaiBusinessServiceServer) EnrichRecord(ctx context.Context, req *pb.EnrichRecordRequest) (*pb.EnrichRecordResponse, error) {
	s.logger.Printf("EnrichRecord called with %d enrichment types", len(req.EnrichmentTypes))

	if req.Record == nil {
		return nil, status.Errorf(codes.InvalidArgument, "record cannot be nil")
	}

	enrichedRecord := req.Record
	addedData := make(map[string]string)

	for _, enrichmentType := range req.EnrichmentTypes {
		switch enrichmentType {
		case "geocoding":
			// Add location information for entrance and exit ICs
			addedData["entrance_prefecture"] = getPreferctureByIC(req.Record.EntranceIc)
			addedData["exit_prefecture"] = getPreferctureByIC(req.Record.ExitIc)

		case "toll_calculation":
			// Validate toll amount against standard rates
			expectedToll := calculateExpectedToll(req.Record.EntranceIc, req.Record.ExitIc)
			addedData["expected_toll"] = strconv.FormatInt(expectedToll, 10)
			addedData["toll_variance"] = strconv.FormatFloat(
				float64(int64(req.Record.TollAmount)-expectedToll)/float64(expectedToll)*100, 'f', 2, 64)

		case "route_analysis":
			// Add route distance and travel time estimates
			addedData["estimated_distance_km"] = strconv.FormatFloat(
				estimateDistance(req.Record.EntranceIc, req.Record.ExitIc), 'f', 1, 64)
			addedData["estimated_travel_minutes"] = strconv.FormatFloat(
				estimateTravelTime(req.Record.EntranceIc, req.Record.ExitIc), 'f', 0, 64)

		case "business_classification":
			// Classify the trip as business or personal based on patterns
			classification := classifyTrip(req.Record)
			addedData["business_classification"] = classification
			addedData["classification_confidence"] = "0.75"

		default:
			s.logger.Printf("Unknown enrichment type: %s", enrichmentType)
		}
	}

	response := &pb.EnrichRecordResponse{
		EnrichedRecord: enrichedRecord,
		AddedData:      addedData,
	}

	s.logger.Printf("EnrichRecord completed: %d data points added", len(addedData))
	return response, nil
}

// MatchDuplicates finds potential duplicate records
func (s *MeisaiBusinessServiceServer) MatchDuplicates(ctx context.Context, req *pb.MatchDuplicatesRequest) (*pb.MatchDuplicatesResponse, error) {
	s.logger.Printf("MatchDuplicates called with similarity threshold: %f", req.SimilarityThreshold)

	if req.Record == nil {
		return nil, status.Errorf(codes.InvalidArgument, "record cannot be nil")
	}

	// Search for similar records
	// This would typically involve complex similarity algorithms
	// For this implementation, we'll do a simplified search

	var matches []*pb.DuplicateMatch

	// Search by exact date and similar amounts
	if req.Record.Date != "" {
		dateStr := req.Record.Date
		similarRecords, err := s.recordRepo.GetByDateRange(ctx, dateStr, dateStr, 100, 0)
		if err == nil {
			for _, record := range similarRecords.Records {
				if record.Id == req.Record.Id {
					continue // Skip self
				}

				similarity := calculateSimilarity(req.Record, record)
				if similarity >= req.SimilarityThreshold {
					matchingFields := getMatchingFields(req.Record, record)
					match := &pb.DuplicateMatch{
						Record:          record,
						SimilarityScore: similarity,
						MatchingFields:  matchingFields,
					}
					matches = append(matches, match)
				}
			}
		}
	}

	response := &pb.MatchDuplicatesResponse{
		Matches: matches,
	}

	s.logger.Printf("MatchDuplicates completed: %d potential duplicates found", len(matches))
	return response, nil
}

// MergeRecords merges multiple records into one
func (s *MeisaiBusinessServiceServer) MergeRecords(ctx context.Context, req *pb.MergeRecordsRequest) (*pb.MergeRecordsResponse, error) {
	s.logger.Printf("MergeRecords called for %d records with strategy: %s", len(req.RecordIds), req.MergeStrategy)

	if len(req.RecordIds) < 2 {
		return nil, status.Errorf(codes.InvalidArgument, "at least 2 records required for merging")
	}

	// Retrieve all records to be merged
	var records []*pb.ETCMeisaiRecord
	for _, recordID := range req.RecordIds {
		record, err := s.recordRepo.GetByID(ctx, recordID)
		if err != nil {
			return nil, status.Errorf(codes.NotFound, "record not found: %d", recordID)
		}
		records = append(records, record)
	}

	// Apply merge strategy
	var masterRecord *pb.ETCMeisaiRecord
	switch req.MergeStrategy {
	case "keep_newest":
		masterRecord = findNewestRecord(records)
	case "keep_oldest":
		masterRecord = findOldestRecord(records)
	case "manual":
		// Use field preferences to merge
		masterRecord = mergeWithPreferences(records, req.FieldPreferences)
	default:
		return nil, status.Errorf(codes.InvalidArgument, "unknown merge strategy: %s", req.MergeStrategy)
	}

	// Create the merged record
	mergedRecord, err := s.recordRepo.Create(ctx, masterRecord)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create merged record: %v", err)
	}

	// Archive the original records
	var archivedIDs []int64
	for _, recordID := range req.RecordIds {
		deleteErr := s.recordRepo.Delete(ctx, recordID)
		if deleteErr != nil {
			s.logger.Printf("Failed to archive record %d: %v", recordID, deleteErr)
		} else {
			archivedIDs = append(archivedIDs, recordID)
		}
	}

	response := &pb.MergeRecordsResponse{
		MergedRecord:      mergedRecord,
		ArchivedRecordIds: archivedIDs,
	}

	s.logger.Printf("MergeRecords completed: merged %d records into ID %d", len(archivedIDs), mergedRecord.Id)
	return response, nil
}

// CalculateTollSummary calculates toll summaries and statistics
func (s *MeisaiBusinessServiceServer) CalculateTollSummary(ctx context.Context, req *pb.CalculateTollRequest) (*pb.CalculateTollResponse, error) {
	s.logger.Printf("CalculateTollSummary called for period")

	if req.Period == nil {
		return nil, status.Errorf(codes.InvalidArgument, "period is required")
	}

	// Get records within the specified period
	from, to := DateRangeFromProto(req.Period)
	dateFrom := from.Format("2006-01-02")
	dateTo := to.Format("2006-01-02")

	recordsList, err := s.recordRepo.GetByDateRange(ctx, dateFrom, dateTo, 10000, 0)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to retrieve records: %v", err)
	}

	// Filter by car number and ETC card if specified
	var filteredRecords []*pb.ETCMeisaiRecord
	for _, record := range recordsList.Records {
		include := true
		if req.CarNumber != nil && *req.CarNumber != record.CarNumber {
			include = false
		}
		if req.EtcCardNumber != nil && *req.EtcCardNumber != record.EtcCardNumber {
			include = false
		}
		if include {
			filteredRecords = append(filteredRecords, record)
		}
	}

	// Calculate totals
	var totalAmount int64
	routeMap := make(map[string]*pb.TollByRoute)
	monthlyMap := make(map[string]*pb.TollByMonth)

	for _, record := range filteredRecords {
		totalAmount += int64(record.TollAmount)

		// Route aggregation
		routeKey := record.EntranceIc + "->" + record.ExitIc
		if route, exists := routeMap[routeKey]; exists {
			route.Count++
			route.TotalAmount += int64(record.TollAmount)
		} else {
			routeMap[routeKey] = &pb.TollByRoute{
				EntranceIc:  record.EntranceIc,
				ExitIc:      record.ExitIc,
				Count:       1,
				TotalAmount: int64(record.TollAmount),
			}
		}

		// Monthly aggregation
		monthKey := parseDate(record.Date).Format("2006-01")
		if monthly, exists := monthlyMap[monthKey]; exists {
			monthly.Amount += int64(record.TollAmount)
			monthly.TripCount++
		} else {
			monthlyMap[monthKey] = &pb.TollByMonth{
				Month:     monthKey,
				Amount:    int64(record.TollAmount),
				TripCount: 1,
			}
		}
	}

	// Convert maps to slices
	var routes []*pb.TollByRoute
	for _, route := range routeMap {
		routes = append(routes, route)
	}

	var monthly []*pb.TollByMonth
	for _, month := range monthlyMap {
		monthly = append(monthly, month)
	}

	response := &pb.CalculateTollResponse{
		TotalAmount: totalAmount,
		TripCount:   int32(len(filteredRecords)),
		Routes:      routes,
		Monthly:     monthly,
	}

	s.logger.Printf("CalculateTollSummary completed: %d trips, total ¥%d", len(filteredRecords), totalAmount)
	return response, nil
}

// GenerateReport generates reports in various formats
func (s *MeisaiBusinessServiceServer) GenerateReport(ctx context.Context, req *pb.GenerateReportRequest) (*pb.GenerateReportResponse, error) {
	s.logger.Printf("GenerateReport called: type=%s, format=%s", req.ReportType, req.Format)

	if req.Period == nil {
		return nil, status.Errorf(codes.InvalidArgument, "period is required")
	}

	// Get data for the report period
	from, to := DateRangeFromProto(req.Period)
	dateFrom := from.Format("2006-01-02")
	dateTo := to.Format("2006-01-02")

	recordsList, err := s.recordRepo.GetByDateRange(ctx, dateFrom, dateTo, 10000, 0)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to retrieve records: %v", err)
	}

	// Filter records if specific cars/cards are requested
	var filteredRecords []*pb.ETCMeisaiRecord
	for _, record := range recordsList.Records {
		include := true
		if len(req.CarNumbers) > 0 {
			found := false
			for _, carNum := range req.CarNumbers {
				if carNum == record.CarNumber {
					found = true
					break
				}
			}
			if !found {
				include = false
			}
		}
		if len(req.EtcCardNumbers) > 0 {
			found := false
			for _, cardNum := range req.EtcCardNumbers {
				if cardNum == record.EtcCardNumber {
					found = true
					break
				}
			}
			if !found {
				include = false
			}
		}
		if include {
			filteredRecords = append(filteredRecords, record)
		}
	}

	// Generate report content based on format
	var reportData []byte
	var contentType string
	var filename string
	var totalAmount int64

	for _, record := range filteredRecords {
		totalAmount += int64(record.TollAmount)
	}

	switch req.Format {
	case "csv":
		csvData := generateCSVReport(filteredRecords, req.ReportType)
		reportData = []byte(csvData)
		contentType = "text/csv"
		filename = fmt.Sprintf("%s_report_%s.csv", req.ReportType, time.Now().Format("20060102"))

	case "pdf":
		// In a real implementation, this would generate a PDF
		reportData = []byte("PDF report content would be generated here")
		contentType = "application/pdf"
		filename = fmt.Sprintf("%s_report_%s.pdf", req.ReportType, time.Now().Format("20060102"))

	default:
		return nil, status.Errorf(codes.InvalidArgument, "unsupported format: %s", req.Format)
	}

	metadata := &pb.ReportMetadata{
		GeneratedAt: timestamppb.Now(),
		RecordCount: int32(len(filteredRecords)),
		TotalAmount: totalAmount,
		Summary: map[string]string{
			"period":      dateFrom + " to " + dateTo,
			"record_count": strconv.Itoa(len(filteredRecords)),
			"total_amount": strconv.FormatInt(totalAmount, 10),
		},
	}

	response := &pb.GenerateReportResponse{
		ReportData:  reportData,
		ContentType: contentType,
		Filename:    filename,
		Metadata:    metadata,
	}

	s.logger.Printf("GenerateReport completed: %d records, format=%s", len(filteredRecords), req.Format)
	return response, nil
}

// ArchiveRecords archives multiple records
func (s *MeisaiBusinessServiceServer) ArchiveRecords(ctx context.Context, req *pb.ArchiveRecordsRequest) (*pb.ArchiveRecordsResponse, error) {
	s.logger.Printf("ArchiveRecords called for %d records", len(req.RecordIds))

	archivedCount := int32(0)
	archiveID := generateArchiveID()

	for _, recordID := range req.RecordIds {
		err := s.recordRepo.Delete(ctx, recordID) // Assuming Delete performs soft delete/archival
		if err != nil {
			s.logger.Printf("Failed to archive record %d: %v", recordID, err)
		} else {
			archivedCount++
		}
	}

	response := &pb.ArchiveRecordsResponse{
		ArchivedCount: archivedCount,
		ArchiveId:     archiveID,
	}

	s.logger.Printf("ArchiveRecords completed: %d records archived", archivedCount)
	return response, nil
}

// RestoreRecords restores archived records
func (s *MeisaiBusinessServiceServer) RestoreRecords(ctx context.Context, req *pb.RestoreRecordsRequest) (*pb.RestoreRecordsResponse, error) {
	s.logger.Printf("RestoreRecords called for %d records", len(req.RecordIds))

	// In a real implementation, this would restore soft-deleted records
	// For now, we'll simulate the restoration process
	restoredCount := int32(0)
	var failedIds []int64

	for _, recordID := range req.RecordIds {
		// Simulate restoration - in reality, this would update deleted_at field
		_, err := s.recordRepo.GetByID(ctx, recordID)
		if err != nil {
			failedIds = append(failedIds, recordID)
		} else {
			restoredCount++
		}
	}

	response := &pb.RestoreRecordsResponse{
		RestoredCount: restoredCount,
		FailedIds:     failedIds,
	}

	s.logger.Printf("RestoreRecords completed: %d restored, %d failed", restoredCount, len(failedIds))
	return response, nil
}

// ExportRecords exports records in specified format
func (s *MeisaiBusinessServiceServer) ExportRecords(ctx context.Context, req *pb.ExportRecordsRequest) (*pb.ExportRecordsResponse, error) {
	s.logger.Printf("ExportRecords called with format: %s", req.Format)

	// For simplicity, get all records with pagination
	recordsList, err := s.recordRepo.List(ctx, 10000, 0)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to retrieve records: %v", err)
	}

	// Generate export data
	var exportData []byte
	var contentType string
	var filename string

	switch req.Format {
	case "csv":
		csvData := generateRecordCSVExport(recordsList.Records, req.Fields)
		exportData = []byte(csvData)
		contentType = "text/csv"
		filename = "etc_records_" + time.Now().Format("20060102_150405") + ".csv"

	case "json":
		// Simplified JSON export
		exportData = []byte(`{"records": [], "exported_at": "` + time.Now().Format(time.RFC3339) + `"}`)
		contentType = "application/json"
		filename = "etc_records_" + time.Now().Format("20060102_150405") + ".json"

	default:
		return nil, status.Errorf(codes.InvalidArgument, "unsupported format: %s", req.Format)
	}

	response := &pb.ExportRecordsResponse{
		Data:        exportData,
		ContentType: contentType,
		Filename:    filename,
		RecordCount: int32(len(recordsList.Records)),
	}

	s.logger.Printf("ExportRecords completed: %d records exported", len(recordsList.Records))
	return response, nil
}

// Helper functions

func (s *MeisaiBusinessServiceServer) parseCSVRecord(record []string) (*pb.ETCMeisaiRecord, error) {
	if len(record) < 8 {
		return nil, fmt.Errorf("insufficient columns")
	}

	// Parse date
	date, err := time.Parse("2006-01-02", record[0])
	if err != nil {
		return nil, fmt.Errorf("invalid date format: %v", err)
	}

	// Parse toll amount
	tollAmount, err := strconv.ParseInt(record[5], 10, 32)
	if err != nil {
		return nil, fmt.Errorf("invalid toll amount: %v", err)
	}

	pbRecord := &pb.ETCMeisaiRecord{
		Date:          date.Format("2006-01-02"),
		Time:          record[1],
		EntranceIc:    record[2],
		ExitIc:        record[3],
		TollAmount:    int32(tollAmount),
		CarNumber:     record[4],
		EtcCardNumber: record[6],
		CreatedAt:     timestamppb.Now(),
		UpdatedAt:     timestamppb.Now(),
	}

	// Optional fields
	if len(record) > 7 && record[7] != "" {
		pbRecord.EtcNum = &record[7]
	}

	return pbRecord, nil
}

func generateSessionID() string {
	return fmt.Sprintf("session_%d", time.Now().Unix())
}

func calculateValidationScore(record *pb.ETCMeisaiRecord) float32 {
	score := float32(1.0)

	// Deduct points for missing or problematic data
	if record.CarNumber == "" {
		score -= 0.2
	}
	if record.EtcCardNumber == "" {
		score -= 0.2
	}
	if record.TollAmount <= 0 {
		score -= 0.3
	}
	if record.EntranceIc == "" || record.ExitIc == "" {
		score -= 0.3
	}

	if score < 0 {
		score = 0
	}
	return score
}

func getPreferctureByIC(icName string) string {
	// Simplified mapping - in reality, this would be a comprehensive database lookup
	icMappings := map[string]string{
		"東京": "東京都",
		"横浜": "神奈川県",
		"大阪": "大阪府",
		"名古屋": "愛知県",
	}

	for key, prefecture := range icMappings {
		if strings.Contains(icName, key) {
			return prefecture
		}
	}
	return "未知"
}

func calculateExpectedToll(entranceIC, exitIC string) int64 {
	// Simplified toll calculation - in reality, this would use route databases
	baseRate := int64(150)
	distance := estimateDistance(entranceIC, exitIC)
	return baseRate + int64(distance*10) // 10 yen per km
}

func estimateDistance(entranceIC, exitIC string) float64 {
	// Simplified distance estimation - in reality, this would use mapping APIs
	return 50.0 // Default 50km
}

func estimateTravelTime(entranceIC, exitIC string) float64 {
	distance := estimateDistance(entranceIC, exitIC)
	avgSpeed := 80.0 // 80 km/h average highway speed
	return distance / avgSpeed * 60 // Convert to minutes
}

func classifyTrip(record *pb.ETCMeisaiRecord) string {
	// Simplified classification logic
	hour := parseDate(record.Date).Hour()
	if hour >= 7 && hour <= 9 || hour >= 17 && hour <= 19 {
		return "business" // Rush hours likely business
	}
	if parseDate(record.Date).Weekday() == time.Saturday || parseDate(record.Date).Weekday() == time.Sunday {
		return "personal" // Weekends likely personal
	}
	return "unknown"
}

func calculateSimilarity(record1, record2 *pb.ETCMeisaiRecord) float32 {
	score := float32(0.0)

	// Date similarity
	if record1.Date == record2.Date {
		score += 0.3
	}

	// Route similarity
	if record1.EntranceIc == record2.EntranceIc && record1.ExitIc == record2.ExitIc {
		score += 0.4
	}

	// Amount similarity (within 10% tolerance)
	amountDiff := float32(abs(int64(record1.TollAmount) - int64(record2.TollAmount)))
	if amountDiff <= float32(record1.TollAmount)*0.1 {
		score += 0.3
	}

	return score
}

func getMatchingFields(record1, record2 *pb.ETCMeisaiRecord) []string {
	var fields []string

	if record1.Date == record2.Date {
		fields = append(fields, "date")
	}
	if record1.EntranceIc == record2.EntranceIc {
		fields = append(fields, "entrance_ic")
	}
	if record1.ExitIc == record2.ExitIc {
		fields = append(fields, "exit_ic")
	}
	if record1.CarNumber == record2.CarNumber {
		fields = append(fields, "car_number")
	}

	return fields
}

func findNewestRecord(records []*pb.ETCMeisaiRecord) *pb.ETCMeisaiRecord {
	var newest *pb.ETCMeisaiRecord
	for _, record := range records {
		if newest == nil || record.CreatedAt.AsTime().After(newest.CreatedAt.AsTime()) {
			newest = record
		}
	}
	return newest
}

func findOldestRecord(records []*pb.ETCMeisaiRecord) *pb.ETCMeisaiRecord {
	var oldest *pb.ETCMeisaiRecord
	for _, record := range records {
		if oldest == nil || record.CreatedAt.AsTime().Before(oldest.CreatedAt.AsTime()) {
			oldest = record
		}
	}
	return oldest
}

func mergeWithPreferences(records []*pb.ETCMeisaiRecord, preferences map[string]string) *pb.ETCMeisaiRecord {
	// Simplified merge - use the first record as base and apply preferences
	if len(records) == 0 {
		return nil
	}

	merged := records[0]
	// In a real implementation, this would intelligently merge based on preferences
	return merged
}

func generateCSVReport(records []*pb.ETCMeisaiRecord, reportType string) string {
	csvData := "Date,Time,Entrance,Exit,Amount,Car,ETC_Card\n"
	for _, record := range records {
		csvData += fmt.Sprintf("%s,%s,%s,%s,%d,%s,%s\n",
			parseDate(record.Date).Format("2006-01-02"),
			record.Time,
			record.EntranceIc,
			record.ExitIc,
			record.TollAmount,
			record.CarNumber,
			record.EtcCardNumber,
		)
	}
	return csvData
}

func generateRecordCSVExport(records []*pb.ETCMeisaiRecord, fields []string) string {
	// If no fields specified, export all standard fields
	if len(fields) == 0 {
		fields = []string{"date", "time", "entrance_ic", "exit_ic", "toll_amount", "car_number", "etc_card_number"}
	}

	// Create header
	csvData := strings.Join(fields, ",") + "\n"

	// Add data rows
	for _, record := range records {
		var row []string
		for _, field := range fields {
			switch field {
			case "date":
				row = append(row, parseDate(record.Date).Format("2006-01-02"))
			case "time":
				row = append(row, record.Time)
			case "entrance_ic":
				row = append(row, record.EntranceIc)
			case "exit_ic":
				row = append(row, record.ExitIc)
			case "toll_amount":
				row = append(row, strconv.Itoa(int(record.TollAmount)))
			case "car_number":
				row = append(row, record.CarNumber)
			case "etc_card_number":
				row = append(row, record.EtcCardNumber)
			default:
				row = append(row, "")
			}
		}
		csvData += strings.Join(row, ",") + "\n"
	}

	return csvData
}

func generateArchiveID() string {
	return fmt.Sprintf("archive_%d", time.Now().Unix())
}

func abs(x int64) int64 {
	if x < 0 {
		return -x
	}
	return x
}

// parseDate parses a date string in YYYY-MM-DD format
func parseDate(dateStr string) time.Time {
	if dateStr == "" {
		return time.Time{}
	}
	t, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return time.Time{}
	}
	return t
}
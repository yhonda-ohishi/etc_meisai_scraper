package grpc

import (
	"context"
	"crypto/sha256"
	"fmt"
	"sync"
	"time"

	pb "github.com/yhonda-ohishi/etc_meisai/src/pb"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ETCMeisaiRecordRepositoryServer implements the ETCMeisaiRecordRepository service
// Testing improved hook detection
type ETCMeisaiRecordRepositoryServer struct {
	pb.UnimplementedETCMeisaiRecordRepositoryServer
	mu           sync.RWMutex
	records      map[int64]*pb.ETCMeisaiRecord
	recordsByHash map[string]*pb.ETCMeisaiRecord
	recordsByCarNumber map[string][]*pb.ETCMeisaiRecord
	recordsByETCCard map[string][]*pb.ETCMeisaiRecord
	recordsByDate map[string][]*pb.ETCMeisaiRecord // date format: "2006-01-02"
	nextID       int64
}

// NewETCMeisaiRecordRepositoryServer creates a new repository server instance
func NewETCMeisaiRecordRepositoryServer() *ETCMeisaiRecordRepositoryServer {
	return &ETCMeisaiRecordRepositoryServer{
		records:            make(map[int64]*pb.ETCMeisaiRecord),
		recordsByHash:      make(map[string]*pb.ETCMeisaiRecord),
		recordsByCarNumber: make(map[string][]*pb.ETCMeisaiRecord),
		recordsByETCCard:   make(map[string][]*pb.ETCMeisaiRecord),
		recordsByDate:      make(map[string][]*pb.ETCMeisaiRecord),
		nextID:             1,
	}
}

// Create creates a new ETC meisai record
func (s *ETCMeisaiRecordRepositoryServer) Create(ctx context.Context, req *pb.ETCMeisaiRecord) (*pb.ETCMeisaiRecord, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "record is required")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Generate hash if not provided
	if req.Hash == "" {
		req.Hash = s.generateHash(req)
	}

	// Check for duplicates
	if existing, ok := s.recordsByHash[req.Hash]; ok {
		return existing, status.Error(codes.AlreadyExists, "duplicate record")
	}

	// Assign ID and timestamps
	req.Id = s.nextID
	s.nextID++
	now := timestamppb.Now()
	req.CreatedAt = now
	req.UpdatedAt = now

	// Store record
	s.records[req.Id] = req
	s.recordsByHash[req.Hash] = req

	// Update indexes
	if req.CarNumber != "" {
		s.recordsByCarNumber[req.CarNumber] = append(s.recordsByCarNumber[req.CarNumber], req)
	}
	if req.EtcCardNumber != "" {
		s.recordsByETCCard[req.EtcCardNumber] = append(s.recordsByETCCard[req.EtcCardNumber], req)
	}
	if req.Date != "" {
		s.recordsByDate[req.Date] = append(s.recordsByDate[req.Date], req)
	}

	return req, nil
}

// GetByID retrieves a record by ID
func (s *ETCMeisaiRecordRepositoryServer) GetByID(ctx context.Context, req *pb.GetByIDRequest) (*pb.ETCMeisaiRecord, error) {
	if req.Id <= 0 {
		return nil, status.Error(codes.InvalidArgument, "invalid ID")
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	record, ok := s.records[req.Id]
	if !ok {
		return nil, status.Error(codes.NotFound, "record not found")
	}

	return record, nil
}

// GetByHash retrieves a record by hash
func (s *ETCMeisaiRecordRepositoryServer) GetByHash(ctx context.Context, req *pb.GetByHashRequest) (*pb.ETCMeisaiRecord, error) {
	if req.Hash == "" {
		return nil, status.Error(codes.InvalidArgument, "hash is required")
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	record, ok := s.recordsByHash[req.Hash]
	if !ok {
		return nil, status.Error(codes.NotFound, "record not found")
	}

	return record, nil
}

// Update updates an existing record
func (s *ETCMeisaiRecordRepositoryServer) Update(ctx context.Context, req *pb.ETCMeisaiRecord) (*pb.ETCMeisaiRecord, error) {
	if req == nil || req.Id <= 0 {
		return nil, status.Error(codes.InvalidArgument, "valid record with ID is required")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	existing, ok := s.records[req.Id]
	if !ok {
		return nil, status.Error(codes.NotFound, "record not found")
	}

	// Remove from old indexes
	s.removeFromIndexes(existing)

	// Update record
	req.CreatedAt = existing.CreatedAt
	req.UpdatedAt = timestamppb.Now()
	if req.Hash == "" {
		req.Hash = s.generateHash(req)
	}

	// Store updated record
	s.records[req.Id] = req
	s.recordsByHash[req.Hash] = req

	// Add to new indexes
	if req.CarNumber != "" {
		s.recordsByCarNumber[req.CarNumber] = append(s.recordsByCarNumber[req.CarNumber], req)
	}
	if req.EtcCardNumber != "" {
		s.recordsByETCCard[req.EtcCardNumber] = append(s.recordsByETCCard[req.EtcCardNumber], req)
	}
	if req.Date != "" {
		s.recordsByDate[req.Date] = append(s.recordsByDate[req.Date], req)
	}

	return req, nil
}

// Delete deletes a record
func (s *ETCMeisaiRecordRepositoryServer) Delete(ctx context.Context, req *pb.GetByIDRequest) (*emptypb.Empty, error) {
	if req.Id <= 0 {
		return nil, status.Error(codes.InvalidArgument, "invalid ID")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	record, ok := s.records[req.Id]
	if !ok {
		return nil, status.Error(codes.NotFound, "record not found")
	}

	// Remove from all indexes
	s.removeFromIndexes(record)
	delete(s.records, req.Id)
	delete(s.recordsByHash, record.Hash)

	return &emptypb.Empty{}, nil
}

// List retrieves a list of records with pagination
func (s *ETCMeisaiRecordRepositoryServer) List(ctx context.Context, req *pb.ListRecordsRequest) (*pb.ListRecordsResponse, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Default pagination
	pageSize := int(req.PageSize)
	if pageSize <= 0 {
		pageSize = 100
	}
	page := int(req.Page)
	if page < 1 {
		page = 1
	}
	offset := (page - 1) * pageSize

	// Collect all records
	var allRecords []*pb.ETCMeisaiRecord
	for _, record := range s.records {
		allRecords = append(allRecords, record)
	}

	// Apply pagination
	start := offset
	if start > len(allRecords) {
		start = len(allRecords)
	}
	end := start + pageSize
	if end > len(allRecords) {
		end = len(allRecords)
	}

	return &pb.ListRecordsResponse{
		Records:    allRecords[start:end],
		TotalCount: int32(len(allRecords)),
	}, nil
}

// GetByDateRange retrieves records within a date range
func (s *ETCMeisaiRecordRepositoryServer) GetByDateRange(ctx context.Context, req *pb.GetByDateRangeRequest) (*pb.ListRecordsResponse, error) {
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

	// Collect records in date range
	var matchingRecords []*pb.ETCMeisaiRecord
	for date := fromDate; !date.After(toDate); date = date.AddDate(0, 0, 1) {
		dateKey := date.Format("2006-01-02")
		if records, ok := s.recordsByDate[dateKey]; ok {
			matchingRecords = append(matchingRecords, records...)
		}
	}

	// Apply pagination
	limit := int(req.Limit)
	if limit <= 0 {
		limit = 100
	}
	offset := int(req.Offset)

	start := offset
	if start > len(matchingRecords) {
		start = len(matchingRecords)
	}
	end := start + limit
	if end > len(matchingRecords) {
		end = len(matchingRecords)
	}

	return &pb.ListRecordsResponse{
		Records:    matchingRecords[start:end],
		TotalCount: int32(len(matchingRecords)),
	}, nil
}

// GetByCarNumber retrieves records by car number
func (s *ETCMeisaiRecordRepositoryServer) GetByCarNumber(ctx context.Context, req *pb.GetByCarNumberRequest) (*pb.ListRecordsResponse, error) {
	if req.CarNumber == "" {
		return nil, status.Error(codes.InvalidArgument, "car number is required")
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	matchingRecords := s.recordsByCarNumber[req.CarNumber]

	// Apply pagination
	limit := int(req.Limit)
	if limit <= 0 {
		limit = 100
	}
	offset := int(req.Offset)

	start := offset
	if start > len(matchingRecords) {
		start = len(matchingRecords)
	}
	end := start + limit
	if end > len(matchingRecords) {
		end = len(matchingRecords)
	}

	return &pb.ListRecordsResponse{
		Records:    matchingRecords[start:end],
		TotalCount: int32(len(matchingRecords)),
	}, nil
}

// GetByETCCard retrieves records by ETC card number
func (s *ETCMeisaiRecordRepositoryServer) GetByETCCard(ctx context.Context, req *pb.GetByETCCardRequest) (*pb.ListRecordsResponse, error) {
	if req.EtcCardNumber == "" {
		return nil, status.Error(codes.InvalidArgument, "ETC card number is required")
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	matchingRecords := s.recordsByETCCard[req.EtcCardNumber]

	// Apply pagination
	limit := int(req.Limit)
	if limit <= 0 {
		limit = 100
	}
	offset := int(req.Offset)

	start := offset
	if start > len(matchingRecords) {
		start = len(matchingRecords)
	}
	end := start + limit
	if end > len(matchingRecords) {
		end = len(matchingRecords)
	}

	return &pb.ListRecordsResponse{
		Records:    matchingRecords[start:end],
		TotalCount: int32(len(matchingRecords)),
	}, nil
}

// BulkCreate creates multiple records at once
func (s *ETCMeisaiRecordRepositoryServer) BulkCreate(ctx context.Context, req *pb.BulkCreateRecordsRequest) (*pb.BulkCreateRecordsResponse, error) {
	if len(req.Records) == 0 {
		return nil, status.Error(codes.InvalidArgument, "at least one record is required")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	var createdRecords []*pb.ETCMeisaiRecord
	duplicateCount := 0

	for _, record := range req.Records {
		// Generate hash if not provided
		if record.Hash == "" {
			record.Hash = s.generateHash(record)
		}

		// Check for duplicates
		if _, ok := s.recordsByHash[record.Hash]; ok {
			duplicateCount++
			continue
		}

		// Assign ID and timestamps
		record.Id = s.nextID
		s.nextID++
		now := timestamppb.Now()
		record.CreatedAt = now
		record.UpdatedAt = now

		// Store record
		s.records[record.Id] = record
		s.recordsByHash[record.Hash] = record

		// Update indexes
		if record.CarNumber != "" {
			s.recordsByCarNumber[record.CarNumber] = append(s.recordsByCarNumber[record.CarNumber], record)
		}
		if record.EtcCardNumber != "" {
			s.recordsByETCCard[record.EtcCardNumber] = append(s.recordsByETCCard[record.EtcCardNumber], record)
		}
		if record.Date != "" {
			s.recordsByDate[record.Date] = append(s.recordsByDate[record.Date], record)
		}

		createdRecords = append(createdRecords, record)
	}

	return &pb.BulkCreateRecordsResponse{
		Records:        createdRecords,
		CreatedCount:   int32(len(createdRecords)),
		DuplicateCount: int32(duplicateCount),
	}, nil
}

// CheckDuplicate checks if a record with the given hash exists
func (s *ETCMeisaiRecordRepositoryServer) CheckDuplicate(ctx context.Context, req *pb.CheckDuplicateRequest) (*pb.CheckDuplicateResponse, error) {
	if req.Hash == "" {
		return nil, status.Error(codes.InvalidArgument, "hash is required")
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	if existing, ok := s.recordsByHash[req.Hash]; ok {
		return &pb.CheckDuplicateResponse{
			IsDuplicate:    true,
			ExistingRecord: existing,
		}, nil
	}

	return &pb.CheckDuplicateResponse{
		IsDuplicate: false,
	}, nil
}

// GetRecordStatistics retrieves statistics about records
func (s *ETCMeisaiRecordRepositoryServer) GetRecordStatistics(ctx context.Context, req *pb.GetRecordStatisticsRequest) (*pb.RecordStatistics, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Calculate statistics
	totalRecords := int64(len(s.records))
	var totalAmount int64
	uniqueCarNumbers := make(map[string]bool)
	uniqueCardNumbers := make(map[string]bool)
	recordsPerMonth := make(map[string]int32)

	for _, record := range s.records {
		// Apply date filter if provided
		if req.DateFrom != nil || req.DateTo != nil {
			if record.Date == "" {
				continue
			}
			recordDate, err := time.Parse("2006-01-02", record.Date)
			if err != nil {
				continue
			}
			if req.DateFrom != nil {
				fromDate, _ := time.Parse("2006-01-02", *req.DateFrom)
				if recordDate.Before(fromDate) {
					continue
				}
			}
			if req.DateTo != nil {
				toDate, _ := time.Parse("2006-01-02", *req.DateTo)
				if recordDate.After(toDate) {
					continue
				}
			}
		}

		totalAmount += int64(record.TollAmount)
		if record.CarNumber != "" {
			uniqueCarNumbers[record.CarNumber] = true
		}
		if record.EtcCardNumber != "" {
			uniqueCardNumbers[record.EtcCardNumber] = true
		}
		if record.Date != "" {
			if recordDate, err := time.Parse("2006-01-02", record.Date); err == nil {
				monthKey := recordDate.Format("2006-01")
				recordsPerMonth[monthKey]++
			}
		}
	}

	return &pb.RecordStatistics{
		TotalRecords:     totalRecords,
		TotalAmount:      totalAmount,
		UniqueCars:       int32(len(uniqueCarNumbers)),
		UniqueCards:      int32(len(uniqueCardNumbers)),
		RecordsPerMonth:  recordsPerMonth,
	}, nil
}

// Helper functions

// generateHash generates a hash for a record
func (s *ETCMeisaiRecordRepositoryServer) generateHash(record *pb.ETCMeisaiRecord) string {
	data := fmt.Sprintf("%s_%s_%s_%s_%d_%s_%s",
		record.Date,
		record.Time,
		record.EntranceIc,
		record.ExitIc,
		record.TollAmount,
		record.CarNumber,
		record.EtcCardNumber,
	)
	hash := sha256.Sum256([]byte(data))
	return fmt.Sprintf("%x", hash)
}

// removeFromIndexes removes a record from all indexes
func (s *ETCMeisaiRecordRepositoryServer) removeFromIndexes(record *pb.ETCMeisaiRecord) {
	// Remove from car number index
	if record.CarNumber != "" {
		records := s.recordsByCarNumber[record.CarNumber]
		for i, r := range records {
			if r.Id == record.Id {
				s.recordsByCarNumber[record.CarNumber] = append(records[:i], records[i+1:]...)
				break
			}
		}
	}

	// Remove from ETC card index
	if record.EtcCardNumber != "" {
		records := s.recordsByETCCard[record.EtcCardNumber]
		for i, r := range records {
			if r.Id == record.Id {
				s.recordsByETCCard[record.EtcCardNumber] = append(records[:i], records[i+1:]...)
				break
			}
		}
	}

	// Remove from date index
	if record.Date != "" {
		records := s.recordsByDate[record.Date]
		for i, r := range records {
			if r.Id == record.Id {
				s.recordsByDate[record.Date] = append(records[:i], records[i+1:]...)
				break
			}
		}
	}
}
package repositories

import (
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	pb "github.com/yhonda-ohishi/etc_meisai/src/pb"
)

// MemoryRepository implements an in-memory storage for Protocol Buffer messages
// This replaces the need for database/SQL storage
type MemoryRepository struct {
	mu sync.RWMutex

	// Storage maps
	etcRecords      map[int64]*pb.ETCMeisaiRecord
	etcMappings     map[int64]*pb.ETCMapping
	importSessions  map[string]*pb.ImportSession

	// Indexes for faster lookups
	etcRecordsByHash    map[string]int64
	etcRecordsByDate    map[string][]int64
	mappingsByRecord    map[int64][]int64
	sessionsByAccount   map[string][]string

	// ID generators
	nextETCRecordID  int64
	nextETCMappingID int64
}

// NewMemoryRepository creates a new in-memory repository
func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{
		etcRecords:          make(map[int64]*pb.ETCMeisaiRecord),
		etcMappings:         make(map[int64]*pb.ETCMapping),
		importSessions:      make(map[string]*pb.ImportSession),
		etcRecordsByHash:    make(map[string]int64),
		etcRecordsByDate:    make(map[string][]int64),
		mappingsByRecord:    make(map[int64][]int64),
		sessionsByAccount:   make(map[string][]string),
		nextETCRecordID:     1,
		nextETCMappingID:    1,
	}
}

// ETCMeisaiRecord operations

// CreateETCMeisaiRecord creates a new ETC record
func (r *MemoryRepository) CreateETCMeisaiRecord(record *pb.ETCMeisaiRecord) (*pb.ETCMeisaiRecord, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Check for duplicate by hash
	if existingID, exists := r.etcRecordsByHash[record.Hash]; exists {
		return nil, fmt.Errorf("duplicate record with hash %s already exists with ID %d", record.Hash, existingID)
	}

	// Assign new ID
	record.Id = r.nextETCRecordID
	r.nextETCRecordID++

	// Store record
	r.etcRecords[record.Id] = record

	// Update indexes
	r.etcRecordsByHash[record.Hash] = record.Id
	r.etcRecordsByDate[record.Date] = append(r.etcRecordsByDate[record.Date], record.Id)

	return record, nil
}

// GetETCMeisaiRecord retrieves an ETC record by ID
func (r *MemoryRepository) GetETCMeisaiRecord(id int64) (*pb.ETCMeisaiRecord, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	record, exists := r.etcRecords[id]
	if !exists {
		return nil, fmt.Errorf("ETC record with ID %d not found", id)
	}

	// Return a copy to prevent external modifications
	return copyETCMeisaiRecord(record), nil
}

// GetETCMeisaiRecordByHash retrieves an ETC record by hash
func (r *MemoryRepository) GetETCMeisaiRecordByHash(hash string) (*pb.ETCMeisaiRecord, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	id, exists := r.etcRecordsByHash[hash]
	if !exists {
		return nil, fmt.Errorf("ETC record with hash %s not found", hash)
	}

	return copyETCMeisaiRecord(r.etcRecords[id]), nil
}

// ListETCMeisaiRecords lists ETC records with pagination
func (r *MemoryRepository) ListETCMeisaiRecords(limit, offset int32) ([]*pb.ETCMeisaiRecord, int64, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	total := int64(len(r.etcRecords))

	// Collect all records
	var allRecords []*pb.ETCMeisaiRecord
	for _, record := range r.etcRecords {
		allRecords = append(allRecords, copyETCMeisaiRecord(record))
	}

	// Apply pagination
	start := int(offset)
	end := start + int(limit)
	if start > len(allRecords) {
		return []*pb.ETCMeisaiRecord{}, total, nil
	}
	if end > len(allRecords) {
		end = len(allRecords)
	}

	return allRecords[start:end], total, nil
}

// GetETCMeisaiRecordsByDateRange retrieves records within a date range
func (r *MemoryRepository) GetETCMeisaiRecordsByDateRange(startDate, endDate string) ([]*pb.ETCMeisaiRecord, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	start, _ := time.Parse("2006-01-02", startDate)
	end, _ := time.Parse("2006-01-02", endDate)

	var results []*pb.ETCMeisaiRecord
	for date, ids := range r.etcRecordsByDate {
		recordDate, err := time.Parse("2006-01-02", date)
		if err != nil {
			continue
		}

		if recordDate.After(start.Add(-time.Hour*24)) && recordDate.Before(end.Add(time.Hour*24)) {
			for _, id := range ids {
				if record, exists := r.etcRecords[id]; exists {
					results = append(results, copyETCMeisaiRecord(record))
				}
			}
		}
	}

	return results, nil
}

// BulkCreateETCMeisaiRecords creates multiple ETC records
func (r *MemoryRepository) BulkCreateETCMeisaiRecords(records []*pb.ETCMeisaiRecord) (int32, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	successCount := int32(0)
	for _, record := range records {
		// Skip duplicates
		if _, exists := r.etcRecordsByHash[record.Hash]; exists {
			continue
		}

		// Assign new ID
		record.Id = r.nextETCRecordID
		r.nextETCRecordID++

		// Store record
		r.etcRecords[record.Id] = record

		// Update indexes
		r.etcRecordsByHash[record.Hash] = record.Id
		r.etcRecordsByDate[record.Date] = append(r.etcRecordsByDate[record.Date], record.Id)

		successCount++
	}

	return successCount, nil
}

// ETCMapping operations

// CreateETCMapping creates a new ETC mapping
func (r *MemoryRepository) CreateETCMapping(mapping *pb.ETCMapping) (*pb.ETCMapping, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Assign new ID
	mapping.Id = r.nextETCMappingID
	r.nextETCMappingID++

	// Store mapping
	r.etcMappings[mapping.Id] = mapping

	// Update index
	r.mappingsByRecord[mapping.EtcRecordId] = append(r.mappingsByRecord[mapping.EtcRecordId], mapping.Id)

	return mapping, nil
}

// GetETCMapping retrieves an ETC mapping by ID
func (r *MemoryRepository) GetETCMapping(id int64) (*pb.ETCMapping, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	mapping, exists := r.etcMappings[id]
	if !exists {
		return nil, fmt.Errorf("ETC mapping with ID %d not found", id)
	}

	return copyETCMapping(mapping), nil
}

// GetETCMappingsByRecordID retrieves all mappings for a specific ETC record
func (r *MemoryRepository) GetETCMappingsByRecordID(recordID int64) ([]*pb.ETCMapping, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	mappingIDs, exists := r.mappingsByRecord[recordID]
	if !exists {
		return []*pb.ETCMapping{}, nil
	}

	var results []*pb.ETCMapping
	for _, id := range mappingIDs {
		if mapping, exists := r.etcMappings[id]; exists {
			results = append(results, copyETCMapping(mapping))
		}
	}

	return results, nil
}

// UpdateETCMapping updates an existing ETC mapping
func (r *MemoryRepository) UpdateETCMapping(mapping *pb.ETCMapping) (*pb.ETCMapping, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.etcMappings[mapping.Id]; !exists {
		return nil, fmt.Errorf("ETC mapping with ID %d not found", mapping.Id)
	}

	r.etcMappings[mapping.Id] = mapping
	return mapping, nil
}

// ImportSession operations

// CreateImportSession creates a new import session
func (r *MemoryRepository) CreateImportSession(session *pb.ImportSession) (*pb.ImportSession, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Generate ID if not provided
	if session.Id == "" {
		session.Id = uuid.New().String()
	}

	// Check for duplicate
	if _, exists := r.importSessions[session.Id]; exists {
		return nil, fmt.Errorf("import session with ID %s already exists", session.Id)
	}

	// Store session
	r.importSessions[session.Id] = session

	// Update index
	r.sessionsByAccount[session.AccountId] = append(r.sessionsByAccount[session.AccountId], session.Id)

	return session, nil
}

// GetImportSession retrieves an import session by ID
func (r *MemoryRepository) GetImportSession(id string) (*pb.ImportSession, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	session, exists := r.importSessions[id]
	if !exists {
		return nil, fmt.Errorf("import session with ID %s not found", id)
	}

	return copyImportSession(session), nil
}

// UpdateImportSession updates an existing import session
func (r *MemoryRepository) UpdateImportSession(session *pb.ImportSession) (*pb.ImportSession, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.importSessions[session.Id]; !exists {
		return nil, fmt.Errorf("import session with ID %s not found", session.Id)
	}

	r.importSessions[session.Id] = session
	return session, nil
}

// GetImportSessionsByAccount retrieves all import sessions for an account
func (r *MemoryRepository) GetImportSessionsByAccount(accountID string) ([]*pb.ImportSession, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	sessionIDs, exists := r.sessionsByAccount[accountID]
	if !exists {
		return []*pb.ImportSession{}, nil
	}

	var results []*pb.ImportSession
	for _, id := range sessionIDs {
		if session, exists := r.importSessions[id]; exists {
			results = append(results, copyImportSession(session))
		}
	}

	return results, nil
}

// Statistics operations

// GetStatistics returns repository statistics
func (r *MemoryRepository) GetStatistics() map[string]int64 {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return map[string]int64{
		"total_etc_records":    int64(len(r.etcRecords)),
		"total_etc_mappings":   int64(len(r.etcMappings)),
		"total_import_sessions": int64(len(r.importSessions)),
		"unique_hashes":        int64(len(r.etcRecordsByHash)),
		"unique_dates":         int64(len(r.etcRecordsByDate)),
		"unique_accounts":      int64(len(r.sessionsByAccount)),
	}
}

// Clear removes all data from the repository
func (r *MemoryRepository) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.etcRecords = make(map[int64]*pb.ETCMeisaiRecord)
	r.etcMappings = make(map[int64]*pb.ETCMapping)
	r.importSessions = make(map[string]*pb.ImportSession)
	r.etcRecordsByHash = make(map[string]int64)
	r.etcRecordsByDate = make(map[string][]int64)
	r.mappingsByRecord = make(map[int64][]int64)
	r.sessionsByAccount = make(map[string][]string)
	r.nextETCRecordID = 1
	r.nextETCMappingID = 1
}

// Helper functions to create deep copies

func copyETCMeisaiRecord(record *pb.ETCMeisaiRecord) *pb.ETCMeisaiRecord {
	if record == nil {
		return nil
	}

	copy := &pb.ETCMeisaiRecord{
		Id:            record.Id,
		Hash:          record.Hash,
		Date:          record.Date,
		Time:          record.Time,
		EntranceIc:    record.EntranceIc,
		ExitIc:        record.ExitIc,
		TollAmount:    record.TollAmount,
		CarNumber:     record.CarNumber,
		EtcCardNumber: record.EtcCardNumber,
		CreatedAt:     record.CreatedAt,
		UpdatedAt:     record.UpdatedAt,
	}

	if record.EtcNum != nil {
		etcNum := *record.EtcNum
		copy.EtcNum = &etcNum
	}
	if record.DtakoRowId != nil {
		dtakoRowId := *record.DtakoRowId
		copy.DtakoRowId = &dtakoRowId
	}

	return copy
}

func copyETCMapping(mapping *pb.ETCMapping) *pb.ETCMapping {
	if mapping == nil {
		return nil
	}

	return &pb.ETCMapping{
		Id:               mapping.Id,
		EtcRecordId:      mapping.EtcRecordId,
		MappingType:      mapping.MappingType,
		MappedEntityId:   mapping.MappedEntityId,
		MappedEntityType: mapping.MappedEntityType,
		Confidence:       mapping.Confidence,
		Status:           mapping.Status,
		CreatedBy:        mapping.CreatedBy,
		CreatedAt:        mapping.CreatedAt,
		UpdatedAt:        mapping.UpdatedAt,
		Metadata:         mapping.Metadata,
	}
}

func copyImportSession(session *pb.ImportSession) *pb.ImportSession {
	if session == nil {
		return nil
	}

	return &pb.ImportSession{
		Id:            session.Id,
		AccountType:   session.AccountType,
		AccountId:     session.AccountId,
		FileName:      session.FileName,
		FileSize:      session.FileSize,
		Status:        session.Status,
		TotalRows:     session.TotalRows,
		ProcessedRows: session.ProcessedRows,
		SuccessRows:   session.SuccessRows,
		ErrorRows:     session.ErrorRows,
		DuplicateRows: session.DuplicateRows,
		StartedAt:     session.StartedAt,
		CompletedAt:   session.CompletedAt,
		CreatedBy:     session.CreatedBy,
		CreatedAt:     session.CreatedAt,
	}
}
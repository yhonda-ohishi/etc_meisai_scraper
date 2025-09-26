package services

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/yhonda-ohishi/etc_meisai/src/models"
	"github.com/yhonda-ohishi/etc_meisai/src/repositories"
)

// ETCMeisaiService handles business logic for ETC record management
type ETCMeisaiService struct {
	repo   repositories.ETCMeisaiRecordRepository
	logger *log.Logger
}

// NewETCMeisaiService creates a new ETC record management service
func NewETCMeisaiService(repo repositories.ETCMeisaiRecordRepository, logger *log.Logger) *ETCMeisaiService {
	if logger == nil {
		logger = log.New(log.Writer(), "[ETCMeisaiService] ", log.LstdFlags|log.Lshortfile)
	}

	return &ETCMeisaiService{
		repo:   repo,
		logger: logger,
	}
}

// CreateRecordParams contains parameters for creating an ETC record
type CreateRecordParams struct {
	Date            time.Time `json:"date" validate:"required"`
	Time            string    `json:"time" validate:"required"`
	EntranceIC      string    `json:"entrance_ic" validate:"required"`
	ExitIC          string    `json:"exit_ic" validate:"required"`
	TollAmount      int       `json:"toll_amount" validate:"required,min=0"`
	CarNumber       string    `json:"car_number" validate:"required"`
	ETCCardNumber   string    `json:"etc_card_number" validate:"required"`
	ETCNum          *string   `json:"etc_num,omitempty"`
	DtakoRowID      *int64    `json:"dtako_row_id,omitempty"`
}

// ListRecordsParams contains parameters for listing ETC records
type ListRecordsParams struct {
	Page      int        `json:"page" validate:"min=1"`
	PageSize  int        `json:"page_size" validate:"min=1,max=1000"`
	DateFrom  *time.Time `json:"date_from,omitempty"`
	DateTo    *time.Time `json:"date_to,omitempty"`
	CarNumber *string    `json:"car_number,omitempty"`
	ETCNumber *string    `json:"etc_number,omitempty"`
	ETCNum    *string    `json:"etc_num,omitempty"`
	SortBy    string     `json:"sort_by"`     // date, toll_amount, car_number
	SortOrder string     `json:"sort_order"`  // asc, desc
}

// ListRecordsResponse contains the response for listing ETC records
type ListRecordsResponse struct {
	Records    []*models.ETCMeisaiRecord `json:"records"`
	TotalCount int64                     `json:"total_count"`
	Page       int                       `json:"page"`
	PageSize   int                       `json:"page_size"`
	TotalPages int                       `json:"total_pages"`
}

// CreateRecord creates a new ETC record with hash generation
func (s *ETCMeisaiService) CreateRecord(ctx context.Context, params *CreateRecordParams) (*models.ETCMeisaiRecord, error) {
	if params == nil {
		return nil, fmt.Errorf("params cannot be nil")
	}

	// Validate required fields
	if params.Date.IsZero() {
		return nil, fmt.Errorf("date is required")
	}

	s.logger.Printf("Creating ETC record for car: %s, date: %s", params.CarNumber, params.Date.Format("2006-01-02"))

	// Create record model
	record := &models.ETCMeisaiRecord{
		Date:          params.Date,
		Time:          params.Time,
		EntranceIC:    params.EntranceIC,
		ExitIC:        params.ExitIC,
		TollAmount:    params.TollAmount,
		CarNumber:     params.CarNumber,
		ETCCardNumber: params.ETCCardNumber,
		ETCNum:        params.ETCNum,
		DtakoRowID:    params.DtakoRowID,
	}

	// Generate hash for the record
	record.Hash = record.GenerateHash()

	// Validate the record
	if err := record.Validate(); err != nil {
		s.logger.Printf("Validation failed for record: %v", err)
		return nil, fmt.Errorf("record validation failed: %w", err)
	}

	// Start transaction
	txRepo, err := s.repo.BeginTx(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to start transaction: %w", err)
	}
	defer func() {
		if r := recover(); r != nil {
			txRepo.RollbackTx()
		}
	}()

	// Check for duplicate hash
	isDuplicate, err := txRepo.CheckDuplicateHash(ctx, record.Hash)
	if err != nil {
		txRepo.RollbackTx()
		return nil, fmt.Errorf("failed to check for duplicates: %w", err)
	}
	if isDuplicate {
		txRepo.RollbackTx()
		return nil, fmt.Errorf("duplicate record with hash: %s", record.Hash)
	}

	// Create the record
	if err := txRepo.Create(ctx, record); err != nil {
		txRepo.RollbackTx()
		s.logger.Printf("Failed to create record: %v", err)
		return nil, fmt.Errorf("failed to create record: %w", err)
	}

	// Commit transaction
	if err := txRepo.CommitTx(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	s.logger.Printf("Successfully created ETC record with ID: %d, hash: %s", record.ID, record.Hash)
	return record, nil
}

// GetRecord retrieves an ETC record by ID
func (s *ETCMeisaiService) GetRecord(ctx context.Context, id int64) (*models.ETCMeisaiRecord, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid record ID: %d", id)
	}

	s.logger.Printf("Retrieving ETC record with ID: %d", id)

	record, err := s.repo.GetByID(ctx, id)
	if err != nil {
		s.logger.Printf("Failed to retrieve record: %v", err)
		return nil, fmt.Errorf("failed to retrieve record: %w", err)
	}

	return record, nil
}

// ListRecords lists ETC records with filtering and pagination
func (s *ETCMeisaiService) ListRecords(ctx context.Context, params *ListRecordsParams) (*ListRecordsResponse, error) {
	// Initialize params if nil
	if params == nil {
		params = &ListRecordsParams{}
	}

	// Set defaults
	if params.Page <= 0 {
		params.Page = 1
	}
	if params.PageSize <= 0 {
		params.PageSize = 50
	}
	if params.PageSize > 1000 {
		params.PageSize = 1000
	}
	if params.SortBy == "" {
		params.SortBy = "date"
	}
	if params.SortOrder == "" {
		params.SortOrder = "desc"
	}

	s.logger.Printf("Listing ETC records - page: %d, size: %d", params.Page, params.PageSize)

	// Convert to repository params
	repoParams := repositories.ListRecordsParams{
		Page:      params.Page,
		PageSize:  params.PageSize,
		DateFrom:  params.DateFrom,
		DateTo:    params.DateTo,
		CarNumber: params.CarNumber,
		ETCNumber: params.ETCNumber,
		ETCNum:    params.ETCNum,
		SortBy:    params.SortBy,
		SortOrder: params.SortOrder,
	}

	// Get records from repository
	records, totalCount, err := s.repo.List(ctx, repoParams)
	if err != nil {
		s.logger.Printf("Failed to retrieve records: %v", err)
		return nil, fmt.Errorf("failed to retrieve records: %w", err)
	}

	totalPages := int((totalCount + int64(params.PageSize) - 1) / int64(params.PageSize))

	response := &ListRecordsResponse{
		Records:    records,
		TotalCount: totalCount,
		Page:       params.Page,
		PageSize:   params.PageSize,
		TotalPages: totalPages,
	}

	s.logger.Printf("Successfully retrieved %d records (page %d of %d)", len(records), params.Page, totalPages)
	return response, nil
}

// UpdateRecord updates an existing ETC record
func (s *ETCMeisaiService) UpdateRecord(ctx context.Context, id int64, params *CreateRecordParams) (*models.ETCMeisaiRecord, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid record ID: %d", id)
	}

	if params == nil {
		return nil, fmt.Errorf("params cannot be nil")
	}

	s.logger.Printf("Updating ETC record with ID: %d", id)

	// Start transaction
	txRepo, err := s.repo.BeginTx(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to start transaction: %w", err)
	}
	defer func() {
		if r := recover(); r != nil {
			txRepo.RollbackTx()
		}
	}()

	// Get existing record
	record, err := txRepo.GetByID(ctx, id)
	if err != nil {
		txRepo.RollbackTx()
		return nil, fmt.Errorf("failed to retrieve record: %w", err)
	}

	// Update fields
	record.Date = params.Date
	record.Time = params.Time
	record.EntranceIC = params.EntranceIC
	record.ExitIC = params.ExitIC
	record.TollAmount = params.TollAmount
	record.CarNumber = params.CarNumber
	record.ETCCardNumber = params.ETCCardNumber
	record.ETCNum = params.ETCNum
	record.DtakoRowID = params.DtakoRowID

	// Regenerate hash with new data
	record.Hash = record.GenerateHash()

	// Validate the updated record
	if err := record.Validate(); err != nil {
		txRepo.RollbackTx()
		s.logger.Printf("Validation failed for updated record: %v", err)
		return nil, fmt.Errorf("record validation failed: %w", err)
	}

	// Check for duplicate hash (excluding current record)
	isDuplicate, err := txRepo.CheckDuplicateHash(ctx, record.Hash, record.ID)
	if err != nil {
		txRepo.RollbackTx()
		return nil, fmt.Errorf("failed to check for duplicates: %w", err)
	}
	if isDuplicate {
		txRepo.RollbackTx()
		return nil, fmt.Errorf("duplicate record with hash: %s", record.Hash)
	}

	// Save the updated record
	if err := txRepo.Update(ctx, record); err != nil {
		txRepo.RollbackTx()
		s.logger.Printf("Failed to update record: %v", err)
		return nil, fmt.Errorf("failed to update record: %w", err)
	}

	// Commit transaction
	if err := txRepo.CommitTx(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	s.logger.Printf("Successfully updated ETC record with ID: %d", record.ID)
	return record, nil
}

// DeleteRecord performs soft delete on an ETC record
func (s *ETCMeisaiService) DeleteRecord(ctx context.Context, id int64) error {
	if id <= 0 {
		return fmt.Errorf("invalid record ID: %d", id)
	}

	s.logger.Printf("Deleting ETC record with ID: %d", id)

	// Start transaction
	txRepo, err := s.repo.BeginTx(ctx)
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}
	defer func() {
		if r := recover(); r != nil {
			txRepo.RollbackTx()
		}
	}()

	// Check if record exists
	_, err = txRepo.GetByID(ctx, id)
	if err != nil {
		txRepo.RollbackTx()
		return fmt.Errorf("failed to retrieve record: %w", err)
	}

	// Perform soft delete
	if err := txRepo.Delete(ctx, id); err != nil {
		txRepo.RollbackTx()
		s.logger.Printf("Failed to delete record: %v", err)
		return fmt.Errorf("failed to delete record: %w", err)
	}

	// Commit transaction
	if err := txRepo.CommitTx(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	s.logger.Printf("Successfully deleted ETC record with ID: %d", id)
	return nil
}

// GetRecordByHash retrieves an ETC record by its hash
func (s *ETCMeisaiService) GetRecordByHash(ctx context.Context, hash string) (*models.ETCMeisaiRecord, error) {
	if hash == "" {
		return nil, fmt.Errorf("hash cannot be empty")
	}

	s.logger.Printf("Retrieving ETC record with hash: %s", hash)

	record, err := s.repo.GetByHash(ctx, hash)
	if err != nil {
		s.logger.Printf("Failed to retrieve record by hash: %v", err)
		return nil, fmt.Errorf("failed to retrieve record: %w", err)
	}

	return record, nil
}

// ValidateRecord validates an ETC record without saving it
func (s *ETCMeisaiService) ValidateRecord(ctx context.Context, params *CreateRecordParams) error {
	if params == nil {
		return fmt.Errorf("params cannot be nil")
	}

	// Create temporary record for validation
	record := &models.ETCMeisaiRecord{
		Date:          params.Date,
		Time:          params.Time,
		EntranceIC:    params.EntranceIC,
		ExitIC:        params.ExitIC,
		TollAmount:    params.TollAmount,
		CarNumber:     params.CarNumber,
		ETCCardNumber: params.ETCCardNumber,
		ETCNum:        params.ETCNum,
		DtakoRowID:    params.DtakoRowID,
	}

	// Validate the record
	return record.Validate()
}

// HealthCheck performs health check for the service
func (s *ETCMeisaiService) HealthCheck(ctx context.Context) error {
	if s.repo == nil {
		return fmt.Errorf("repository not initialized")
	}

	// Check repository connectivity
	if err := s.repo.Ping(ctx); err != nil {
		return fmt.Errorf("repository ping failed: %w", err)
	}

	return nil
}
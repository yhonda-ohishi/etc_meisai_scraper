package services

import (
	"context"
	"fmt"
	"log"
	"time"

	"gorm.io/gorm"

	"github.com/yhonda-ohishi/etc_meisai/src/models"
)

// ETCMeisaiService handles business logic for ETC record management
type ETCMeisaiService struct {
	db     *gorm.DB
	logger *log.Logger
}

// NewETCMeisaiService creates a new ETC record management service
func NewETCMeisaiService(db *gorm.DB, logger *log.Logger) *ETCMeisaiService {
	if logger == nil {
		logger = log.New(log.Writer(), "[ETCMeisaiService] ", log.LstdFlags|log.Lshortfile)
	}

	return &ETCMeisaiService{
		db:     db,
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

	// Validate the record (this will also generate the hash)
	if err := record.BeforeCreate(s.db); err != nil {
		s.logger.Printf("Validation failed for record: %v", err)
		return nil, fmt.Errorf("record validation failed: %w", err)
	}

	// Start transaction
	tx := s.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return nil, fmt.Errorf("failed to start transaction: %w", tx.Error)
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Check for duplicate hash
	var existingRecord models.ETCMeisaiRecord
	err := tx.Where("hash = ?", record.Hash).First(&existingRecord).Error
	if err == nil {
		tx.Rollback()
		return nil, fmt.Errorf("duplicate record with hash: %s", record.Hash)
	} else if err != gorm.ErrRecordNotFound {
		tx.Rollback()
		return nil, fmt.Errorf("failed to check for duplicates: %w", err)
	}

	// Create the record
	if err := tx.Create(record).Error; err != nil {
		tx.Rollback()
		s.logger.Printf("Failed to create record: %v", err)
		return nil, fmt.Errorf("failed to create record: %w", err)
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
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

	var record models.ETCMeisaiRecord
	err := s.db.WithContext(ctx).First(&record, id).Error
	if err == gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("record not found with ID: %d", id)
	} else if err != nil {
		s.logger.Printf("Failed to retrieve record: %v", err)
		return nil, fmt.Errorf("failed to retrieve record: %w", err)
	}

	return &record, nil
}

// ListRecords lists ETC records with filtering and pagination
func (s *ETCMeisaiService) ListRecords(ctx context.Context, params *ListRecordsParams) (*ListRecordsResponse, error) {
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

	// Build query
	query := s.db.WithContext(ctx).Model(&models.ETCMeisaiRecord{})

	// Apply filters
	if params.DateFrom != nil {
		query = query.Where("date >= ?", *params.DateFrom)
	}
	if params.DateTo != nil {
		query = query.Where("date <= ?", *params.DateTo)
	}
	if params.CarNumber != nil && *params.CarNumber != "" {
		query = query.Where("car_number LIKE ?", "%"+*params.CarNumber+"%")
	}
	if params.ETCNumber != nil && *params.ETCNumber != "" {
		query = query.Where("etc_card_number LIKE ?", "%"+*params.ETCNumber+"%")
	}
	if params.ETCNum != nil && *params.ETCNum != "" {
		query = query.Where("etc_num LIKE ?", "%"+*params.ETCNum+"%")
	}

	// Get total count
	var totalCount int64
	if err := query.Count(&totalCount).Error; err != nil {
		s.logger.Printf("Failed to count records: %v", err)
		return nil, fmt.Errorf("failed to count records: %w", err)
	}

	// Apply sorting and pagination
	orderClause := fmt.Sprintf("%s %s", params.SortBy, params.SortOrder)
	offset := (params.Page - 1) * params.PageSize

	var records []*models.ETCMeisaiRecord
	err := query.Order(orderClause).Offset(offset).Limit(params.PageSize).Find(&records).Error
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

	s.logger.Printf("Updating ETC record with ID: %d", id)

	// Start transaction
	tx := s.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return nil, fmt.Errorf("failed to start transaction: %w", tx.Error)
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Get existing record
	var record models.ETCMeisaiRecord
	err := tx.First(&record, id).Error
	if err == gorm.ErrRecordNotFound {
		tx.Rollback()
		return nil, fmt.Errorf("record not found with ID: %d", id)
	} else if err != nil {
		tx.Rollback()
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
	record.Hash = record.generateHash()

	// Validate the updated record
	if err := record.BeforeSave(tx); err != nil {
		tx.Rollback()
		s.logger.Printf("Validation failed for updated record: %v", err)
		return nil, fmt.Errorf("record validation failed: %w", err)
	}

	// Check for duplicate hash (excluding current record)
	var existingRecord models.ETCMeisaiRecord
	err = tx.Where("hash = ? AND id != ?", record.Hash, record.ID).First(&existingRecord).Error
	if err == nil {
		tx.Rollback()
		return nil, fmt.Errorf("duplicate record with hash: %s", record.Hash)
	} else if err != gorm.ErrRecordNotFound {
		tx.Rollback()
		return nil, fmt.Errorf("failed to check for duplicates: %w", err)
	}

	// Save the updated record
	if err := tx.Save(&record).Error; err != nil {
		tx.Rollback()
		s.logger.Printf("Failed to update record: %v", err)
		return nil, fmt.Errorf("failed to update record: %w", err)
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	s.logger.Printf("Successfully updated ETC record with ID: %d", record.ID)
	return &record, nil
}

// DeleteRecord performs soft delete on an ETC record
func (s *ETCMeisaiService) DeleteRecord(ctx context.Context, id int64) error {
	if id <= 0 {
		return fmt.Errorf("invalid record ID: %d", id)
	}

	s.logger.Printf("Deleting ETC record with ID: %d", id)

	// Start transaction
	tx := s.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return fmt.Errorf("failed to start transaction: %w", tx.Error)
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Check if record exists
	var record models.ETCMeisaiRecord
	err := tx.First(&record, id).Error
	if err == gorm.ErrRecordNotFound {
		tx.Rollback()
		return fmt.Errorf("record not found with ID: %d", id)
	} else if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to retrieve record: %w", err)
	}

	// Perform soft delete
	if err := tx.Delete(&record).Error; err != nil {
		tx.Rollback()
		s.logger.Printf("Failed to delete record: %v", err)
		return fmt.Errorf("failed to delete record: %w", err)
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
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

	var record models.ETCMeisaiRecord
	err := s.db.WithContext(ctx).Where("hash = ?", hash).First(&record).Error
	if err == gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("record not found with hash: %s", hash)
	} else if err != nil {
		s.logger.Printf("Failed to retrieve record by hash: %v", err)
		return nil, fmt.Errorf("failed to retrieve record: %w", err)
	}

	return &record, nil
}

// ValidateRecord validates an ETC record without saving it
func (s *ETCMeisaiService) ValidateRecord(ctx context.Context, params *CreateRecordParams) error {
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
	return record.BeforeCreate(s.db)
}

// HealthCheck performs health check for the service
func (s *ETCMeisaiService) HealthCheck(ctx context.Context) error {
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
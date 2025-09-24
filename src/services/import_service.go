package services

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"

	"github.com/yhonda-ohishi/etc_meisai/src/models"
)

// ImportService handles CSV import operations for ETC records
type ImportService struct {
	db     *gorm.DB
	logger *log.Logger
}

// NewImportService creates a new import service
func NewImportService(db *gorm.DB, logger *log.Logger) *ImportService {
	if logger == nil {
		logger = log.New(log.Writer(), "[ImportService] ", log.LstdFlags|log.Lshortfile)
	}

	return &ImportService{
		db:     db,
		logger: logger,
	}
}

// ImportCSVParams contains parameters for CSV import
type ImportCSVParams struct {
	AccountType string `json:"account_type" validate:"required"`
	AccountID   string `json:"account_id" validate:"required"`
	FileName    string `json:"file_name" validate:"required"`
	FileSize    int64  `json:"file_size" validate:"required,min=1"`
	CreatedBy   string `json:"created_by,omitempty"`
}

// ImportCSVStreamParams contains parameters for streaming CSV import
type ImportCSVStreamParams struct {
	SessionID string   `json:"session_id" validate:"required"`
	Chunks    []string `json:"chunks" validate:"required"`
}

// ImportCSVResult contains the result of CSV import
type ImportCSVResult struct {
	Session       *models.ImportSession     `json:"session"`
	Records       []*models.ETCMeisaiRecord `json:"records"`
	SuccessCount  int                       `json:"success_count"`
	ErrorCount    int                       `json:"error_count"`
	DuplicateCount int                      `json:"duplicate_count"`
	Errors        []models.ImportError      `json:"errors"`
}

// ListImportSessionsParams contains parameters for listing import sessions
type ListImportSessionsParams struct {
	Page        int     `json:"page" validate:"min=1"`
	PageSize    int     `json:"page_size" validate:"min=1,max=1000"`
	AccountType *string `json:"account_type,omitempty"`
	AccountID   *string `json:"account_id,omitempty"`
	Status      *string `json:"status,omitempty"`
	CreatedBy   *string `json:"created_by,omitempty"`
	SortBy      string  `json:"sort_by"`     // created_at, started_at, file_name
	SortOrder   string  `json:"sort_order"`  // asc, desc
}

// ListImportSessionsResponse contains the response for listing import sessions
type ListImportSessionsResponse struct {
	Sessions   []*models.ImportSession `json:"sessions"`
	TotalCount int64                   `json:"total_count"`
	Page       int                     `json:"page"`
	PageSize   int                     `json:"page_size"`
	TotalPages int                     `json:"total_pages"`
}

// CSVRow represents a single row from the CSV file
type CSVRow struct {
	Date          string `json:"date"`
	Time          string `json:"time"`
	EntranceIC    string `json:"entrance_ic"`
	ExitIC        string `json:"exit_ic"`
	TollAmount    string `json:"toll_amount"`
	CarNumber     string `json:"car_number"`
	ETCCardNumber string `json:"etc_card_number"`
	ETCNum        string `json:"etc_num,omitempty"`
}

// DuplicateResult contains information about duplicate records
type DuplicateResult struct {
	Hash           string                 `json:"hash"`
	ExistingRecord *models.ETCMeisaiRecord `json:"existing_record"`
	NewRecord      *models.ETCMeisaiRecord `json:"new_record"`
	Action         string                  `json:"action"` // skip, update, create_new
}

// ImportCSV processes CSV import and creates records
func (s *ImportService) ImportCSV(ctx context.Context, params *ImportCSVParams, data io.Reader) (*ImportCSVResult, error) {
	s.logger.Printf("Starting CSV import for account: %s (%s), file: %s", params.AccountID, params.AccountType, params.FileName)

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

	// Create import session
	session := &models.ImportSession{
		AccountType: params.AccountType,
		AccountID:   params.AccountID,
		FileName:    params.FileName,
		FileSize:    params.FileSize,
		Status:      string(models.ImportStatusPending),
		CreatedBy:   params.CreatedBy,
	}

	if err := tx.Create(session).Error; err != nil {
		tx.Rollback()
		s.logger.Printf("Failed to create import session: %v", err)
		return nil, fmt.Errorf("failed to create import session: %w", err)
	}

	// Start processing
	if err := session.StartProcessing(); err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to start processing: %w", err)
	}

	// Parse CSV data
	records, errors, err := s.parseCSVData(data)
	if err != nil {
		tx.Rollback()
		s.logger.Printf("Failed to parse CSV data: %v", err)
		return nil, fmt.Errorf("failed to parse CSV data: %w", err)
	}

	session.TotalRows = len(records) + len(errors)

	// Process each record
	var successRecords []*models.ETCMeisaiRecord
	var duplicateCount int

	for i, record := range records {
		// Check for duplicates
		var existingRecord models.ETCMeisaiRecord
		err := tx.Where("hash = ?", record.Hash).First(&existingRecord).Error
		if err == nil {
			// Duplicate found
			duplicateCount++
			s.logger.Printf("Duplicate record found with hash: %s", record.Hash)
			continue
		} else if err != gorm.ErrRecordNotFound {
			// Database error
			tx.Rollback()
			return nil, fmt.Errorf("failed to check for duplicates: %w", err)
		}

		// Create the record
		if err := tx.Create(record).Error; err != nil {
			s.logger.Printf("Failed to create record %d: %v", i+1, err)

			// Add error to session
			session.AddError(i+1, "creation_error", err.Error(), "")
			session.ErrorRows++
		} else {
			successRecords = append(successRecords, record)
			session.SuccessRows++
		}
		session.ProcessedRows++
	}

	// Add parsing errors to session
	for _, importErr := range errors {
		session.AddError(importErr.RowNumber, importErr.ErrorType, importErr.ErrorMessage, importErr.RawData)
		session.ErrorRows++
	}

	session.DuplicateRows = duplicateCount

	// Complete the session
	if session.ErrorRows == 0 {
		if err := session.Complete(); err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("failed to complete session: %w", err)
		}
	} else {
		// Set as completed with errors
		session.Status = string(models.ImportStatusCompleted)
		now := time.Now()
		session.CompletedAt = &now
	}

	// Save session updates
	if err := tx.Save(session).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to save session: %w", err)
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	result := &ImportCSVResult{
		Session:        session,
		Records:        successRecords,
		SuccessCount:   session.SuccessRows,
		ErrorCount:     session.ErrorRows,
		DuplicateCount: session.DuplicateRows,
		Errors:         errors,
	}

	s.logger.Printf("CSV import completed - Success: %d, Errors: %d, Duplicates: %d",
		session.SuccessRows, session.ErrorRows, session.DuplicateRows)

	return result, nil
}

// ImportCSVStream handles streaming import of CSV data in chunks
func (s *ImportService) ImportCSVStream(ctx context.Context, params *ImportCSVStreamParams) (*ImportCSVResult, error) {
	s.logger.Printf("Starting streaming CSV import for session: %s", params.SessionID)

	// Get existing session
	session, err := s.GetImportSession(ctx, params.SessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get import session: %w", err)
	}

	if !session.IsPending() {
		return nil, fmt.Errorf("session is not in pending status: %s", session.Status)
	}

	// Combine chunks into single data stream
	csvData := strings.Join(params.Chunks, "")
	dataReader := strings.NewReader(csvData)

	// Update file size
	session.FileSize = int64(len(csvData))

	// Process the CSV data
	importParams := &ImportCSVParams{
		AccountType: session.AccountType,
		AccountID:   session.AccountID,
		FileName:    session.FileName,
		FileSize:    session.FileSize,
		CreatedBy:   session.CreatedBy,
	}

	// Delete the temporary session since ImportCSV will create a new one
	if err := s.db.WithContext(ctx).Delete(session).Error; err != nil {
		s.logger.Printf("Failed to delete temporary session: %v", err)
	}

	return s.ImportCSV(ctx, importParams, dataReader)
}

// GetImportSession retrieves an import session by ID
func (s *ImportService) GetImportSession(ctx context.Context, sessionID string) (*models.ImportSession, error) {
	if sessionID == "" {
		return nil, fmt.Errorf("session ID cannot be empty")
	}

	s.logger.Printf("Retrieving import session: %s", sessionID)

	var session models.ImportSession
	err := s.db.WithContext(ctx).First(&session, "id = ?", sessionID).Error
	if err == gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("import session not found: %s", sessionID)
	} else if err != nil {
		s.logger.Printf("Failed to retrieve import session: %v", err)
		return nil, fmt.Errorf("failed to retrieve import session: %w", err)
	}

	return &session, nil
}

// ListImportSessions lists import sessions with filtering and pagination
func (s *ImportService) ListImportSessions(ctx context.Context, params *ListImportSessionsParams) (*ListImportSessionsResponse, error) {
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
		params.SortBy = "created_at"
	}
	if params.SortOrder == "" {
		params.SortOrder = "desc"
	}

	s.logger.Printf("Listing import sessions - page: %d, size: %d", params.Page, params.PageSize)

	// Build query
	query := s.db.WithContext(ctx).Model(&models.ImportSession{})

	// Apply filters
	if params.AccountType != nil && *params.AccountType != "" {
		query = query.Where("account_type = ?", *params.AccountType)
	}
	if params.AccountID != nil && *params.AccountID != "" {
		query = query.Where("account_id LIKE ?", "%"+*params.AccountID+"%")
	}
	if params.Status != nil && *params.Status != "" {
		query = query.Where("status = ?", *params.Status)
	}
	if params.CreatedBy != nil && *params.CreatedBy != "" {
		query = query.Where("created_by LIKE ?", "%"+*params.CreatedBy+"%")
	}

	// Get total count
	var totalCount int64
	if err := query.Count(&totalCount).Error; err != nil {
		s.logger.Printf("Failed to count import sessions: %v", err)
		return nil, fmt.Errorf("failed to count import sessions: %w", err)
	}

	// Apply sorting and pagination
	orderClause := fmt.Sprintf("%s %s", params.SortBy, params.SortOrder)
	offset := (params.Page - 1) * params.PageSize

	var sessions []*models.ImportSession
	err := query.Order(orderClause).Offset(offset).Limit(params.PageSize).Find(&sessions).Error
	if err != nil {
		s.logger.Printf("Failed to retrieve import sessions: %v", err)
		return nil, fmt.Errorf("failed to retrieve import sessions: %w", err)
	}

	totalPages := int((totalCount + int64(params.PageSize) - 1) / int64(params.PageSize))

	response := &ListImportSessionsResponse{
		Sessions:   sessions,
		TotalCount: totalCount,
		Page:       params.Page,
		PageSize:   params.PageSize,
		TotalPages: totalPages,
	}

	s.logger.Printf("Successfully retrieved %d import sessions (page %d of %d)", len(sessions), params.Page, totalPages)
	return response, nil
}

// ProcessCSV processes bulk CSV data with large dataset support
// This method provides enhanced bulk processing for test coverage
func (s *ImportService) ProcessCSV(ctx context.Context, rows []*CSVRow, options *BulkProcessOptions) (*BulkProcessResult, error) {
	if rows == nil {
		return nil, fmt.Errorf("CSV rows cannot be nil")
	}

	if options == nil {
		options = &BulkProcessOptions{
			BatchSize:    1000,
			MaxConcurrency: 5,
			SkipErrors:   false,
		}
	}

	// Validate batch size
	if options.BatchSize <= 0 || options.BatchSize > 10000 {
		return nil, fmt.Errorf("batch size must be between 1 and 10000")
	}
	if options.MaxConcurrency <= 0 || options.MaxConcurrency > 20 {
		return nil, fmt.Errorf("max concurrency must be between 1 and 20")
	}

	result := &BulkProcessResult{
		TotalRows:     len(rows),
		SuccessCount:  0,
		ErrorCount:    0,
		ProcessedAt:   time.Now(),
		Errors:        make([]string, 0),
	}

	// Process in batches
	for i := 0; i < len(rows); i += options.BatchSize {
		end := i + options.BatchSize
		if end > len(rows) {
			end = len(rows)
		}

		batch := rows[i:end]
		batchResult, err := s.processBatch(ctx, batch, options)
		if err != nil {
			if !options.SkipErrors {
				return nil, fmt.Errorf("batch processing failed: %w", err)
			}
			result.ErrorCount += len(batch)
			result.Errors = append(result.Errors, fmt.Sprintf("Batch %d-%d failed: %v", i, end-1, err))
		} else {
			result.SuccessCount += batchResult.SuccessCount
			result.ErrorCount += batchResult.ErrorCount
			result.Errors = append(result.Errors, batchResult.Errors...)
		}
	}

	return result, nil
}

// processBatch processes a batch of CSV rows
func (s *ImportService) processBatch(ctx context.Context, batch []*CSVRow, options *BulkProcessOptions) (*BulkProcessResult, error) {
	result := &BulkProcessResult{
		TotalRows:    len(batch),
		SuccessCount: 0,
		ErrorCount:   0,
		Errors:       make([]string, 0),
	}

	for _, row := range batch {
		_, err := s.ProcessCSVRow(ctx, row)
		if err != nil {
			result.ErrorCount++
			if !options.SkipErrors {
				return nil, err
			}
			result.Errors = append(result.Errors, err.Error())
		} else {
			result.SuccessCount++
		}
	}

	return result, nil
}

// BulkProcessOptions configures bulk processing behavior
type BulkProcessOptions struct {
	BatchSize      int  // Number of rows per batch
	MaxConcurrency int  // Maximum concurrent batches
	SkipErrors     bool // Continue processing on errors
}

// BulkProcessResult contains bulk processing results
type BulkProcessResult struct {
	TotalRows     int       // Total rows processed
	SuccessCount  int       // Successfully processed rows
	ErrorCount    int       // Failed rows
	ProcessedAt   time.Time // Processing timestamp
	Errors        []string  // Error messages
}

// ProcessCSVRow processes a single CSV row
func (s *ImportService) ProcessCSVRow(ctx context.Context, row *CSVRow) (*models.ETCMeisaiRecord, error) {
	if row == nil {
		return nil, fmt.Errorf("CSV row is nil")
	}

	// Parse date
	date, err := time.Parse("2006-01-02", row.Date)
	if err != nil {
		return nil, fmt.Errorf("invalid date format: %s", row.Date)
	}

	// Parse toll amount
	tollAmount, err := strconv.Atoi(row.TollAmount)
	if err != nil {
		return nil, fmt.Errorf("invalid toll amount: %s", row.TollAmount)
	}

	// Create record
	record := &models.ETCMeisaiRecord{
		Date:          date,
		Time:          row.Time,
		EntranceIC:    row.EntranceIC,
		ExitIC:        row.ExitIC,
		TollAmount:    tollAmount,
		CarNumber:     row.CarNumber,
		ETCCardNumber: row.ETCCardNumber,
	}

	// Set ETC number if provided
	if row.ETCNum != "" {
		record.ETCNum = &row.ETCNum
	}

	// Validate and generate hash
	if err := record.BeforeCreate(s.db); err != nil {
		return nil, fmt.Errorf("record validation failed: %w", err)
	}

	return record, nil
}

// HandleDuplicates detects and handles duplicate records
func (s *ImportService) HandleDuplicates(ctx context.Context, records []*models.ETCMeisaiRecord) ([]*DuplicateResult, error) {
	s.logger.Printf("Checking for duplicates in %d records", len(records))

	var results []*DuplicateResult

	// Extract hashes from records
	hashes := make([]string, len(records))
	for i, record := range records {
		hashes[i] = record.Hash
	}

	// Check for existing records with these hashes
	var existingRecords []*models.ETCMeisaiRecord
	err := s.db.WithContext(ctx).Where("hash IN ?", hashes).Find(&existingRecords).Error
	if err != nil {
		return nil, fmt.Errorf("failed to check for duplicates: %w", err)
	}

	// Create a map of existing records by hash
	existingMap := make(map[string]*models.ETCMeisaiRecord)
	for _, existing := range existingRecords {
		existingMap[existing.Hash] = existing
	}

	// Check each record for duplicates
	for _, record := range records {
		if existing, found := existingMap[record.Hash]; found {
			result := &DuplicateResult{
				Hash:           record.Hash,
				ExistingRecord: existing,
				NewRecord:      record,
				Action:         "skip", // Default action for duplicates
			}
			results = append(results, result)
		}
	}

	s.logger.Printf("Found %d duplicate records", len(results))
	return results, nil
}

// parseCSVData parses CSV data and returns records and errors
func (s *ImportService) parseCSVData(data io.Reader) ([]*models.ETCMeisaiRecord, []models.ImportError, error) {
	reader := csv.NewReader(data)
	reader.FieldsPerRecord = -1 // Allow variable number of fields

	var records []*models.ETCMeisaiRecord
	var errors []models.ImportError
	rowNumber := 0

	for {
		rowNumber++
		row, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			errors = append(errors, models.ImportError{
				RowNumber:    rowNumber,
				ErrorType:    "parse_error",
				ErrorMessage: fmt.Sprintf("Failed to parse CSV row: %v", err),
				RawData:      strings.Join(row, ","),
			})
			continue
		}

		// Skip header row
		if rowNumber == 1 {
			continue
		}

		// Ensure minimum number of fields
		if len(row) < 7 {
			errors = append(errors, models.ImportError{
				RowNumber:    rowNumber,
				ErrorType:    "insufficient_fields",
				ErrorMessage: fmt.Sprintf("Row has %d fields, expected at least 7", len(row)),
				RawData:      strings.Join(row, ","),
			})
			continue
		}

		// Create CSV row structure
		csvRow := &CSVRow{
			Date:          strings.TrimSpace(row[0]),
			Time:          strings.TrimSpace(row[1]),
			EntranceIC:    strings.TrimSpace(row[2]),
			ExitIC:        strings.TrimSpace(row[3]),
			TollAmount:    strings.TrimSpace(row[4]),
			CarNumber:     strings.TrimSpace(row[5]),
			ETCCardNumber: strings.TrimSpace(row[6]),
		}

		// Optional ETC number field
		if len(row) > 7 {
			csvRow.ETCNum = strings.TrimSpace(row[7])
		}

		// Process the row
		record, err := s.ProcessCSVRow(context.Background(), csvRow)
		if err != nil {
			errors = append(errors, models.ImportError{
				RowNumber:    rowNumber,
				ErrorType:    "validation_error",
				ErrorMessage: err.Error(),
				RawData:      strings.Join(row, ","),
			})
			continue
		}

		records = append(records, record)
	}

	s.logger.Printf("Parsed %d records with %d errors from CSV", len(records), len(errors))
	return records, errors, nil
}

// CancelImportSession cancels an ongoing import session
func (s *ImportService) CancelImportSession(ctx context.Context, sessionID string) error {
	s.logger.Printf("Cancelling import session: %s", sessionID)

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

	// Get session
	var session models.ImportSession
	err := tx.First(&session, "id = ?", sessionID).Error
	if err == gorm.ErrRecordNotFound {
		tx.Rollback()
		return fmt.Errorf("import session not found: %s", sessionID)
	} else if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to retrieve import session: %w", err)
	}

	// Cancel the session
	if err := session.Cancel(); err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to cancel session: %w", err)
	}

	// Save session
	if err := tx.Save(&session).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to save cancelled session: %w", err)
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	s.logger.Printf("Successfully cancelled import session: %s", sessionID)
	return nil
}

// HealthCheck performs health check for the service
func (s *ImportService) HealthCheck(ctx context.Context) error {
	if s.db == nil {
		return fmt.Errorf("database not initialized")
	}

	// Check database connectivity with a simple query
	if err := s.db.Exec("SELECT 1").Error; err != nil {
		return fmt.Errorf("database ping failed: %w", err)
	}

	return nil
}
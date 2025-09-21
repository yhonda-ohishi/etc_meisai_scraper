package models

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// ImportStatus represents the status of an import session
type ImportStatus string

const (
	ImportStatusPending    ImportStatus = "pending"
	ImportStatusProcessing ImportStatus = "processing"
	ImportStatusCompleted  ImportStatus = "completed"
	ImportStatusFailed     ImportStatus = "failed"
	ImportStatusCancelled  ImportStatus = "cancelled"
)

// AccountType represents the type of account
type AccountType string

const (
	AccountTypeCorporate AccountType = "corporate"
	AccountTypePersonal  AccountType = "personal"
)

// ImportError represents an error that occurred during import
type ImportError struct {
	RowNumber    int    `json:"row_number"`
	ErrorType    string `json:"error_type"`
	ErrorMessage string `json:"error_message"`
	RawData      string `json:"raw_data,omitempty"`
}

// ImportSession represents an import session for CSV files
type ImportSession struct {
	ID            string         `gorm:"primaryKey;size:36" json:"id"` // UUID
	AccountType   string         `gorm:"size:20;not null;index" json:"account_type"`
	AccountID     string         `gorm:"size:50;not null;index" json:"account_id"`
	FileName      string         `gorm:"size:255;not null" json:"file_name"`
	FileSize      int64          `gorm:"not null" json:"file_size"`
	Status        string         `gorm:"size:20;not null;index" json:"status"`
	TotalRows     int            `gorm:"default:0" json:"total_rows"`
	ProcessedRows int            `gorm:"default:0" json:"processed_rows"`
	SuccessRows   int            `gorm:"default:0" json:"success_rows"`
	ErrorRows     int            `gorm:"default:0" json:"error_rows"`
	DuplicateRows int            `gorm:"default:0" json:"duplicate_rows"`
	StartedAt     time.Time      `gorm:"not null" json:"started_at"`
	CompletedAt   *time.Time     `json:"completed_at,omitempty"`
	ErrorLog      datatypes.JSON `gorm:"type:json" json:"error_log,omitempty"`
	CreatedBy     string         `gorm:"size:100" json:"created_by,omitempty"`
	CreatedAt     time.Time      `gorm:"autoCreateTime" json:"created_at"`
}

// TableName returns the table name for GORM
func (ImportSession) TableName() string {
	return "import_sessions"
}

// BeforeCreate hook to generate UUID and validate data before creating
func (s *ImportSession) BeforeCreate(tx *gorm.DB) error {
	// Generate UUID if not provided
	if s.ID == "" {
		s.ID = uuid.New().String()
	}

	// Set started time if not provided
	if s.StartedAt.IsZero() {
		s.StartedAt = time.Now()
	}

	// Set default status if not provided
	if s.Status == "" {
		s.Status = string(ImportStatusPending)
	}

	return s.validate()
}

// BeforeSave hook to validate data before saving
func (s *ImportSession) BeforeSave(tx *gorm.DB) error {
	return s.validate()
}

// validate performs comprehensive validation of the import session
func (s *ImportSession) validate() error {
	// Validate UUID format
	if err := s.validateID(); err != nil {
		return err
	}

	// Validate account type
	if err := s.validateAccountType(); err != nil {
		return err
	}

	// Validate account ID
	if err := s.validateAccountID(); err != nil {
		return err
	}

	// Validate file name
	if err := s.validateFileName(); err != nil {
		return err
	}

	// Validate file size
	if err := s.validateFileSize(); err != nil {
		return err
	}

	// Validate status
	if err := s.validateStatus(); err != nil {
		return err
	}

	// Validate row counts
	if err := s.validateRowCounts(); err != nil {
		return err
	}

	// Validate created_by if provided
	if s.CreatedBy != "" && len(s.CreatedBy) > 100 {
		return fmt.Errorf("created_by field too long (max 100 characters)")
	}

	return nil
}

// validateID validates the UUID format
func (s *ImportSession) validateID() error {
	if s.ID == "" {
		return fmt.Errorf("ID cannot be empty")
	}

	// Validate UUID v4 format
	if _, err := uuid.Parse(s.ID); err != nil {
		return fmt.Errorf("invalid UUID format: %w", err)
	}

	return nil
}

// validateAccountType validates the account type
func (s *ImportSession) validateAccountType() error {
	validTypes := []string{
		string(AccountTypeCorporate),
		string(AccountTypePersonal),
	}

	accountType := strings.ToLower(strings.TrimSpace(s.AccountType))
	if accountType == "" {
		return fmt.Errorf("account type cannot be empty")
	}

	for _, validType := range validTypes {
		if accountType == validType {
			s.AccountType = accountType // Normalize to lowercase
			return nil
		}
	}

	return fmt.Errorf("invalid account type: %s (must be one of: %s)",
		s.AccountType, strings.Join(validTypes, ", "))
}

// validateAccountID validates the account ID
func (s *ImportSession) validateAccountID() error {
	accountID := strings.TrimSpace(s.AccountID)
	if accountID == "" {
		return fmt.Errorf("account ID cannot be empty")
	}

	if len(accountID) > 50 {
		return fmt.Errorf("account ID too long (max 50 characters)")
	}

	// Account ID should be alphanumeric with some special characters
	pattern := `^[a-zA-Z0-9\-_@.]+$`
	if matched, _ := regexp.MatchString(pattern, accountID); !matched {
		return fmt.Errorf("account ID contains invalid characters")
	}

	s.AccountID = accountID
	return nil
}

// validateFileName validates the file name
func (s *ImportSession) validateFileName() error {
	fileName := strings.TrimSpace(s.FileName)
	if fileName == "" {
		return fmt.Errorf("file name cannot be empty")
	}

	if len(fileName) > 255 {
		return fmt.Errorf("file name too long (max 255 characters)")
	}

	// Check file extension
	ext := strings.ToLower(filepath.Ext(fileName))
	if ext != ".csv" {
		return fmt.Errorf("file must have .csv extension")
	}

	s.FileName = fileName
	return nil
}

// validateFileSize validates the file size
func (s *ImportSession) validateFileSize() error {
	if s.FileSize <= 0 {
		return fmt.Errorf("file size must be greater than 0")
	}

	// Maximum file size: 100MB
	maxSize := int64(100 * 1024 * 1024)
	if s.FileSize > maxSize {
		return fmt.Errorf("file size too large (max 100MB)")
	}

	return nil
}

// validateStatus validates the import status
func (s *ImportSession) validateStatus() error {
	validStatuses := []string{
		string(ImportStatusPending),
		string(ImportStatusProcessing),
		string(ImportStatusCompleted),
		string(ImportStatusFailed),
		string(ImportStatusCancelled),
	}

	status := strings.ToLower(strings.TrimSpace(s.Status))
	if status == "" {
		return fmt.Errorf("status cannot be empty")
	}

	for _, validStatus := range validStatuses {
		if status == validStatus {
			s.Status = status // Normalize to lowercase
			return nil
		}
	}

	return fmt.Errorf("invalid status: %s (must be one of: %s)",
		s.Status, strings.Join(validStatuses, ", "))
}

// validateRowCounts validates the row count fields
func (s *ImportSession) validateRowCounts() error {
	// All counts must be non-negative
	if s.TotalRows < 0 {
		return fmt.Errorf("total rows cannot be negative")
	}
	if s.ProcessedRows < 0 {
		return fmt.Errorf("processed rows cannot be negative")
	}
	if s.SuccessRows < 0 {
		return fmt.Errorf("success rows cannot be negative")
	}
	if s.ErrorRows < 0 {
		return fmt.Errorf("error rows cannot be negative")
	}
	if s.DuplicateRows < 0 {
		return fmt.Errorf("duplicate rows cannot be negative")
	}

	// Processed rows should equal sum of success, error, and duplicate rows
	expectedProcessed := s.SuccessRows + s.ErrorRows + s.DuplicateRows
	if s.ProcessedRows != expectedProcessed {
		return fmt.Errorf("processed rows (%d) must equal sum of success (%d) + error (%d) + duplicate (%d) rows",
			s.ProcessedRows, s.SuccessRows, s.ErrorRows, s.DuplicateRows)
	}

	// Processed rows cannot exceed total rows
	if s.ProcessedRows > s.TotalRows {
		return fmt.Errorf("processed rows (%d) cannot exceed total rows (%d)",
			s.ProcessedRows, s.TotalRows)
	}

	return nil
}

// IsCompleted returns true if the import is completed
func (s *ImportSession) IsCompleted() bool {
	return s.Status == string(ImportStatusCompleted)
}

// IsFailed returns true if the import failed
func (s *ImportSession) IsFailed() bool {
	return s.Status == string(ImportStatusFailed)
}

// IsInProgress returns true if the import is currently processing
func (s *ImportSession) IsInProgress() bool {
	return s.Status == string(ImportStatusProcessing)
}

// IsPending returns true if the import is pending
func (s *ImportSession) IsPending() bool {
	return s.Status == string(ImportStatusPending)
}

// CanTransitionTo checks if the session can transition to the given status
func (s *ImportSession) CanTransitionTo(newStatus string) bool {
	currentStatus := ImportStatus(s.Status)
	targetStatus := ImportStatus(newStatus)

	switch currentStatus {
	case ImportStatusPending:
		return targetStatus == ImportStatusProcessing || targetStatus == ImportStatusCancelled
	case ImportStatusProcessing:
		return targetStatus == ImportStatusCompleted || targetStatus == ImportStatusFailed || targetStatus == ImportStatusCancelled
	case ImportStatusCompleted, ImportStatusFailed, ImportStatusCancelled:
		return false // Terminal states
	default:
		return false
	}
}

// StartProcessing transitions the session to processing status
func (s *ImportSession) StartProcessing() error {
	if !s.CanTransitionTo(string(ImportStatusProcessing)) {
		return fmt.Errorf("cannot start processing from status: %s", s.Status)
	}
	s.Status = string(ImportStatusProcessing)
	if s.StartedAt.IsZero() {
		s.StartedAt = time.Now()
	}
	return nil
}

// Complete transitions the session to completed status
func (s *ImportSession) Complete() error {
	if !s.CanTransitionTo(string(ImportStatusCompleted)) {
		return fmt.Errorf("cannot complete from status: %s", s.Status)
	}
	s.Status = string(ImportStatusCompleted)
	now := time.Now()
	s.CompletedAt = &now
	return nil
}

// Fail transitions the session to failed status
func (s *ImportSession) Fail() error {
	if !s.CanTransitionTo(string(ImportStatusFailed)) {
		return fmt.Errorf("cannot fail from status: %s", s.Status)
	}
	s.Status = string(ImportStatusFailed)
	now := time.Now()
	s.CompletedAt = &now
	return nil
}

// Cancel transitions the session to cancelled status
func (s *ImportSession) Cancel() error {
	if !s.CanTransitionTo(string(ImportStatusCancelled)) {
		return fmt.Errorf("cannot cancel from status: %s", s.Status)
	}
	s.Status = string(ImportStatusCancelled)
	now := time.Now()
	s.CompletedAt = &now
	return nil
}

// AddError adds an error to the error log
func (s *ImportSession) AddError(rowNumber int, errorType, errorMessage, rawData string) error {
	var errors []ImportError

	// Get existing errors
	if s.ErrorLog != nil {
		if err := json.Unmarshal([]byte(s.ErrorLog), &errors); err != nil {
			return fmt.Errorf("failed to unmarshal existing errors: %w", err)
		}
	}

	// Add new error
	newError := ImportError{
		RowNumber:    rowNumber,
		ErrorType:    errorType,
		ErrorMessage: errorMessage,
		RawData:      rawData,
	}
	errors = append(errors, newError)

	// Update error log
	jsonData, err := json.Marshal(errors)
	if err != nil {
		return fmt.Errorf("failed to marshal errors: %w", err)
	}
	s.ErrorLog = datatypes.JSON(jsonData)
	return nil
}

// GetErrors returns the list of import errors
func (s *ImportSession) GetErrors() ([]ImportError, error) {
	if s.ErrorLog == nil {
		return nil, nil
	}

	var errors []ImportError
	if err := json.Unmarshal([]byte(s.ErrorLog), &errors); err != nil {
		return nil, fmt.Errorf("failed to unmarshal errors: %w", err)
	}

	return errors, nil
}

// GetProgressPercentage returns the progress as a percentage
func (s *ImportSession) GetProgressPercentage() float64 {
	if s.TotalRows == 0 {
		return 0.0
	}
	return float64(s.ProcessedRows) / float64(s.TotalRows) * 100.0
}

// GetSuccessRate returns the success rate as a percentage
func (s *ImportSession) GetSuccessRate() float64 {
	if s.ProcessedRows == 0 {
		return 0.0
	}
	return float64(s.SuccessRows) / float64(s.ProcessedRows) * 100.0
}

// GetDuration returns the duration of the import session
func (s *ImportSession) GetDuration() time.Duration {
	if s.CompletedAt != nil {
		return s.CompletedAt.Sub(s.StartedAt)
	}
	if s.IsInProgress() {
		return time.Since(s.StartedAt)
	}
	return 0
}

// UpdateProgress updates the progress counters
func (s *ImportSession) UpdateProgress(success, error, duplicate int) {
	s.SuccessRows += success
	s.ErrorRows += error
	s.DuplicateRows += duplicate
	s.ProcessedRows = s.SuccessRows + s.ErrorRows + s.DuplicateRows
}
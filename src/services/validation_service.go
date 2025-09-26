package services

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	pb "github.com/yhonda-ohishi/etc_meisai/src/pb"
)

// ValidationService centralizes all validation logic extracted from GORM hooks
type ValidationService struct{}

// NewValidationService creates a new validation service
func NewValidationService() *ValidationService {
	return &ValidationService{}
}

// ValidateETCMeisaiRecord validates an ETC Meisai Record
func (v *ValidationService) ValidateETCMeisaiRecord(record *pb.ETCMeisaiRecord) error {
	if record == nil {
		return fmt.Errorf("record cannot be nil")
	}

	// Validate date
	if record.Date == "" {
		return fmt.Errorf("date is required")
	}

	// Parse and validate date format
	parsedDate, err := time.Parse("2006-01-02", record.Date)
	if err != nil {
		return fmt.Errorf("invalid date format, expected YYYY-MM-DD: %w", err)
	}

	// Validate date (not in future)
	if parsedDate.After(time.Now()) {
		return fmt.Errorf("date cannot be in the future")
	}

	// Validate time
	if err := v.validateTimeFormat(record.Time); err != nil {
		return err
	}

	// Validate IC names
	if err := v.validateICName("entrance IC", record.EntranceIc); err != nil {
		return err
	}
	if err := v.validateICName("exit IC", record.ExitIc); err != nil {
		return err
	}

	// Validate toll amount
	if err := v.validateTollAmount(record.TollAmount); err != nil {
		return err
	}

	// Validate car number
	if err := v.validateCarNumber(record.CarNumber); err != nil {
		return err
	}

	// Validate ETC card number
	if err := v.validateETCCardNumber(record.EtcCardNumber); err != nil {
		return err
	}

	// Validate ETC number if provided
	if record.EtcNum != nil && *record.EtcNum != "" {
		if err := v.validateETCNum(*record.EtcNum); err != nil {
			return err
		}
	}

	return nil
}

// ValidateImportSession validates an Import Session
func (v *ValidationService) ValidateImportSession(session *pb.ImportSession) error {
	if session == nil {
		return fmt.Errorf("session cannot be nil")
	}

	// Validate ID (UUID format)
	if err := v.validateUUID(session.Id); err != nil {
		return fmt.Errorf("invalid session ID: %w", err)
	}

	// Validate account type
	if err := v.validateAccountType(session.AccountType); err != nil {
		return err
	}

	// Validate account ID
	if err := v.validateAccountID(session.AccountId); err != nil {
		return err
	}

	// Validate file name
	if err := v.validateFileName(session.FileName); err != nil {
		return err
	}

	// Validate file size
	if err := v.validateFileSize(session.FileSize); err != nil {
		return err
	}

	// Validate row counts
	if err := v.validateRowCounts(session); err != nil {
		return err
	}

	// Validate timestamps
	if err := v.validateSessionTimestamps(session); err != nil {
		return err
	}

	return nil
}

// ValidateETCMapping validates an ETC Mapping
func (v *ValidationService) ValidateETCMapping(mapping *pb.ETCMapping) error {
	if mapping == nil {
		return fmt.Errorf("mapping cannot be nil")
	}

	// Validate ETC record ID
	if mapping.EtcRecordId <= 0 {
		return fmt.Errorf("ETC record ID must be positive")
	}

	// Validate mapping type
	if strings.TrimSpace(mapping.MappingType) == "" {
		return fmt.Errorf("mapping type cannot be empty")
	}

	// Validate mapped entity ID
	if mapping.MappedEntityId <= 0 {
		return fmt.Errorf("mapped entity ID must be positive")
	}

	// Validate mapped entity type
	if strings.TrimSpace(mapping.MappedEntityType) == "" {
		return fmt.Errorf("mapped entity type cannot be empty")
	}

	// Validate confidence (0.0 to 1.0)
	if mapping.Confidence < 0.0 || mapping.Confidence > 1.0 {
		return fmt.Errorf("confidence must be between 0.0 and 1.0")
	}

	// Validate created by
	if strings.TrimSpace(mapping.CreatedBy) == "" {
		return fmt.Errorf("created by cannot be empty")
	}

	return nil
}

// validateTimeFormat validates HH:MM:SS format
func (v *ValidationService) validateTimeFormat(timeStr string) error {
	if strings.TrimSpace(timeStr) == "" {
		return fmt.Errorf("time is required")
	}

	// Validate time format (HH:MM:SS)
	timeRegex := regexp.MustCompile(`^([01]?[0-9]|2[0-3]):[0-5][0-9]:[0-5][0-9]$`)
	if !timeRegex.MatchString(timeStr) {
		return fmt.Errorf("time must be in HH:MM:SS format")
	}

	return nil
}

// validateICName validates IC station names
func (v *ValidationService) validateICName(fieldName, icName string) error {
	if strings.TrimSpace(icName) == "" {
		return fmt.Errorf("%s cannot be empty", fieldName)
	}

	if len(icName) > 100 {
		return fmt.Errorf("%s name too long (max 100 characters)", fieldName)
	}

	return nil
}

// validateTollAmount validates toll amount
func (v *ValidationService) validateTollAmount(amount int32) error {
	if amount < 0 {
		return fmt.Errorf("toll amount must be non-negative")
	}

	if amount > 999999 {
		return fmt.Errorf("toll amount too large (max 999999)")
	}

	return nil
}

// validateCarNumber validates Japanese vehicle number formats
func (v *ValidationService) validateCarNumber(carNumber string) error {
	if strings.TrimSpace(carNumber) == "" {
		return fmt.Errorf("car number cannot be empty")
	}

	// Japanese vehicle number patterns
	patterns := []string{
		`^\d{3}-\d{2}$`,          // 軽自動車: 123-45
		`^\d{3}\s\d{2}$`,         // 軽自動車: 123 45
		`^[あ-ん]{1}\d{3}$`,       // ひらがな + 数字: あ123
		`^[ア-ン]{1}\d{3}$`,       // カタカナ + 数字: ア123
		`^\d{2}-\d{2}$`,          // 二輪: 12-34
		`^\d{4}$`,                // 4桁数字
		`^[a-zA-Z0-9\-\s]{3,20}$`, // 一般的なパターン
	}

	for _, pattern := range patterns {
		if matched, _ := regexp.MatchString(pattern, carNumber); matched {
			return nil
		}
	}

	return fmt.Errorf("invalid car number format")
}

// validateETCCardNumber validates ETC card number format
func (v *ValidationService) validateETCCardNumber(cardNumber string) error {
	if strings.TrimSpace(cardNumber) == "" {
		return fmt.Errorf("ETC card number cannot be empty")
	}

	// Remove spaces and hyphens for validation
	cleaned := strings.ReplaceAll(strings.ReplaceAll(cardNumber, " ", ""), "-", "")

	// Check if it's 16-19 digits
	if len(cleaned) < 16 || len(cleaned) > 19 {
		return fmt.Errorf("ETC card number must be 16-19 digits")
	}

	// Check if all characters are digits
	if _, err := strconv.ParseInt(cleaned, 10, 64); err != nil {
		return fmt.Errorf("ETC card number must contain only digits")
	}

	return nil
}

// validateETCNum validates ETC 2.0 device number format
func (v *ValidationService) validateETCNum(etcNum string) error {
	if etcNum == "" {
		return nil // Optional field
	}

	// ETC 2.0 device number is typically alphanumeric, 5-50 characters
	etcNum = strings.TrimSpace(etcNum)
	if len(etcNum) < 5 || len(etcNum) > 50 {
		return fmt.Errorf("ETC number must be 5-50 characters")
	}

	// Allow alphanumeric characters and some special characters
	pattern := `^[a-zA-Z0-9\-_]+$`
	if matched, _ := regexp.MatchString(pattern, etcNum); !matched {
		return fmt.Errorf("ETC number contains invalid characters")
	}

	return nil
}

// validateUUID validates UUID format
func (v *ValidationService) validateUUID(id string) error {
	if strings.TrimSpace(id) == "" {
		return fmt.Errorf("ID cannot be empty")
	}

	// UUID v4 pattern
	uuidPattern := `^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`
	if matched, _ := regexp.MatchString(uuidPattern, id); !matched {
		return fmt.Errorf("invalid UUID format")
	}

	return nil
}

// validateAccountType validates account type
func (v *ValidationService) validateAccountType(accountType string) error {
	if strings.TrimSpace(accountType) == "" {
		return fmt.Errorf("account type cannot be empty")
	}

	validTypes := []string{"corporate", "personal"}
	for _, validType := range validTypes {
		if strings.ToLower(accountType) == validType {
			return nil
		}
	}

	return fmt.Errorf("invalid account type, must be 'corporate' or 'personal'")
}

// validateAccountID validates account ID
func (v *ValidationService) validateAccountID(accountID string) error {
	if strings.TrimSpace(accountID) == "" {
		return fmt.Errorf("account ID cannot be empty")
	}

	if len(accountID) > 100 {
		return fmt.Errorf("account ID too long (max 100 characters)")
	}

	return nil
}

// validateFileName validates file name
func (v *ValidationService) validateFileName(fileName string) error {
	if strings.TrimSpace(fileName) == "" {
		return fmt.Errorf("file name cannot be empty")
	}

	if len(fileName) > 255 {
		return fmt.Errorf("file name too long (max 255 characters)")
	}

	// Check for CSV extension
	if !strings.HasSuffix(strings.ToLower(fileName), ".csv") {
		return fmt.Errorf("file must be a CSV file")
	}

	return nil
}

// validateFileSize validates file size
func (v *ValidationService) validateFileSize(fileSize int64) error {
	if fileSize <= 0 {
		return fmt.Errorf("file size must be positive")
	}

	// Max file size: 50MB
	const maxFileSize = 50 * 1024 * 1024
	if fileSize > maxFileSize {
		return fmt.Errorf("file size too large (max 50MB)")
	}

	return nil
}

// validateRowCounts validates row count consistency
func (v *ValidationService) validateRowCounts(session *pb.ImportSession) error {
	// Processed rows should not exceed total rows
	if session.ProcessedRows > session.TotalRows {
		return fmt.Errorf("processed rows cannot exceed total rows")
	}

	// Success + error + duplicate rows should equal processed rows
	calculatedProcessed := session.SuccessRows + session.ErrorRows + session.DuplicateRows
	if calculatedProcessed != session.ProcessedRows {
		return fmt.Errorf("row count mismatch: success(%d) + error(%d) + duplicate(%d) = %d, but processed = %d",
			session.SuccessRows, session.ErrorRows, session.DuplicateRows, calculatedProcessed, session.ProcessedRows)
	}

	// All counts should be non-negative
	if session.TotalRows < 0 || session.ProcessedRows < 0 || session.SuccessRows < 0 ||
		session.ErrorRows < 0 || session.DuplicateRows < 0 {
		return fmt.Errorf("row counts must be non-negative")
	}

	return nil
}

// validateSessionTimestamps validates session timestamps
func (v *ValidationService) validateSessionTimestamps(session *pb.ImportSession) error {
	// Started at should be valid if provided
	if session.StartedAt != nil {
		if err := session.StartedAt.CheckValid(); err != nil {
			return fmt.Errorf("invalid started_at timestamp: %w", err)
		}
	}

	// Completed at should be valid if provided
	if session.CompletedAt != nil {
		if err := session.CompletedAt.CheckValid(); err != nil {
			return fmt.Errorf("invalid completed_at timestamp: %w", err)
		}

		// If both timestamps are provided, completed should be after started
		if session.StartedAt != nil {
			if session.CompletedAt.AsTime().Before(session.StartedAt.AsTime()) {
				return fmt.Errorf("completed time cannot be before started time")
			}
		}
	}

	// Created at should be valid if provided
	if session.CreatedAt != nil {
		if err := session.CreatedAt.CheckValid(); err != nil {
			return fmt.Errorf("invalid created_at timestamp: %w", err)
		}
	}

	return nil
}

// ValidationResult represents the result of a validation operation
type ValidationResult struct {
	IsValid      bool
	ErrorMessage string
	FieldErrors  map[string]string
	ValidatedAt  time.Time
}

// ValidateWithDetails provides detailed validation results
func (v *ValidationService) ValidateWithDetails(entityType string, entity interface{}) *ValidationResult {
	result := &ValidationResult{
		FieldErrors: make(map[string]string),
		ValidatedAt: time.Now(),
	}

	var err error
	switch entityType {
	case "ETCMeisaiRecord":
		if record, ok := entity.(*pb.ETCMeisaiRecord); ok {
			err = v.ValidateETCMeisaiRecord(record)
		} else {
			err = fmt.Errorf("invalid entity type for ETCMeisaiRecord")
		}
	case "ImportSession":
		if session, ok := entity.(*pb.ImportSession); ok {
			err = v.ValidateImportSession(session)
		} else {
			err = fmt.Errorf("invalid entity type for ImportSession")
		}
	case "ETCMapping":
		if mapping, ok := entity.(*pb.ETCMapping); ok {
			err = v.ValidateETCMapping(mapping)
		} else {
			err = fmt.Errorf("invalid entity type for ETCMapping")
		}
	default:
		err = fmt.Errorf("unsupported entity type: %s", entityType)
	}

	if err != nil {
		result.IsValid = false
		result.ErrorMessage = err.Error()
	} else {
		result.IsValid = true
	}

	return result
}
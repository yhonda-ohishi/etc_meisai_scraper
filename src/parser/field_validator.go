package parser

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/yhonda-ohishi/etc_meisai/src/models"
)

// FieldValidator handles validation of individual fields and records
type FieldValidator struct {
	strictMode bool
}

// ValidationSummary provides a summary of validation results
type ValidationSummary struct {
	TotalRecords   int                    `json:"total_records"`
	ValidRecords   int                    `json:"valid_records"`
	InvalidRecords int                    `json:"invalid_records"`
	ErrorsByType   map[string]int         `json:"errors_by_type"`
}

// NewFieldValidator creates a new field validator
func NewFieldValidator() *FieldValidator {
	return &FieldValidator{
		strictMode: false,
	}
}

// SetStrictMode enables or disables strict validation mode
func (v *FieldValidator) SetStrictMode(strict bool) {
	v.strictMode = strict
}

// ValidateUseDate validates the use date string
func (v *FieldValidator) ValidateUseDate(dateStr string) error {
	if dateStr == "" {
		return errors.New("date is required")
	}

	// Try to parse the date using time.Parse
	var err error
	var date time.Time

	dateFormats := []string{
		"2006-01-02",
		"2006/01/02",
		"2006年01月02日",
		"06/01/02",
		"2006/1/2",
	}

	for _, format := range dateFormats {
		date, err = time.Parse(format, dateStr)
		if err == nil {
			break
		}
	}

	if err != nil {
		return fmt.Errorf("invalid date format: %w", err)
	}

	// Check if date is not in the future
	if date.After(time.Now()) {
		return errors.New("date cannot be in the future")
	}

	// Check if date is not too old (more than 10 years)
	tenYearsAgo := time.Now().AddDate(-10, 0, 0)
	if date.Before(tenYearsAgo) {
		return errors.New("date is too old")
	}

	return nil
}

// ValidateUseTime validates the use time string
func (v *FieldValidator) ValidateUseTime(timeStr string) error {
	if timeStr == "" {
		return errors.New("time is required")
	}

	// Check time format (HH:MM)
	timeRegex := regexp.MustCompile(`^([01]?[0-9]|2[0-3]):[0-5][0-9]$`)
	if !timeRegex.MatchString(timeStr) {
		return errors.New("invalid time format")
	}

	return nil
}

// ValidateIC validates an IC (interchange) name
func (v *FieldValidator) ValidateIC(ic string) error {
	if ic == "" {
		return errors.New("IC name is required")
	}

	if len(ic) < 2 {
		return errors.New("IC name is too short")
	}

	if len(ic) > 50 {
		return errors.New("IC name is too long")
	}

	// Check for dangerous characters
	dangerous := []string{"<", ">", "script", "javascript", "eval"}
	icLower := strings.ToLower(ic)
	for _, d := range dangerous {
		if strings.Contains(icLower, d) {
			return errors.New("IC name contains invalid characters")
		}
	}

	return nil
}

// ValidateAmount validates an amount value
func (v *FieldValidator) ValidateAmount(amount int) error {
	if amount < 0 {
		return errors.New("amount cannot be negative")
	}

	if amount > 100000 {
		return errors.New("amount is unreasonably large")
	}

	return nil
}

// ValidateCarNumber validates a car number
func (v *FieldValidator) ValidateCarNumber(carNumber string) error {
	if carNumber == "" {
		return errors.New("car number is required")
	}

	// Basic Japanese car number format validation
	// Should contain Japanese characters and numbers
	// More lenient to allow for various car number formats
	if len([]rune(carNumber)) < 8 { // Japanese car numbers are typically 8+ runes
		return errors.New("invalid car number format")
	}

	// Check for presence of Japanese characters
	hasJapanese := false
	for _, r := range carNumber {
		if (r >= 0x3040 && r <= 0x309F) || // Hiragana
		   (r >= 0x30A0 && r <= 0x30FF) || // Katakana
		   (r >= 0x4E00 && r <= 0x9FAF) {  // CJK Unified Ideographs
			hasJapanese = true
			break
		}
	}
	if !hasJapanese {
		return errors.New("invalid car number format")
	}

	// Check for special characters that shouldn't be in car numbers
	if strings.ContainsAny(carNumber, "!@#$%^&*()_+={}[]|\\:;\"'<>?,./~`") {
		return errors.New("car number contains invalid characters")
	}

	return nil
}

// ValidateETCNumber validates an ETC card number
func (v *FieldValidator) ValidateETCNumber(etcNumber string) error {
	if etcNumber == "" {
		return errors.New("ETC number is required")
	}

	// ETC numbers should be 10 or 16 digits
	if len(etcNumber) != 10 && len(etcNumber) != 16 {
		return errors.New("invalid ETC number format")
	}

	// Should contain only digits
	for _, char := range etcNumber {
		if char < '0' || char > '9' {
			return errors.New("ETC number must contain only digits")
		}
	}

	return nil
}

// ValidateRecord validates a complete ETC record
func (v *FieldValidator) ValidateRecord(record *models.ETCMeisai) error {
	if record == nil {
		return errors.New("record is nil")
	}

	// Validate use date
	if record.UseDate.IsZero() {
		return errors.New("invalid use date")
	}

	// Validate use time
	if err := v.ValidateUseTime(record.UseTime); err != nil {
		return fmt.Errorf("invalid use time: %w", err)
	}

	// Validate ICs
	if err := v.ValidateIC(record.EntryIC); err != nil {
		return fmt.Errorf("invalid entry IC: %w", err)
	}

	if err := v.ValidateIC(record.ExitIC); err != nil {
		return fmt.Errorf("invalid exit IC: %w", err)
	}

	// Check if entry and exit IC are the same
	if record.EntryIC == record.ExitIC {
		return errors.New("entry and exit IC cannot be the same")
	}

	// Validate amount
	if err := v.ValidateAmount(int(record.Amount)); err != nil {
		return fmt.Errorf("invalid amount: %w", err)
	}

	// Validate car number
	if err := v.ValidateCarNumber(record.CarNumber); err != nil {
		return fmt.Errorf("invalid car number: %w", err)
	}

	// Validate ETC number
	if err := v.ValidateETCNumber(record.ETCNumber); err != nil {
		return fmt.Errorf("invalid ETC number: %w", err)
	}

	// Strict mode validations
	if v.strictMode {
		if record.Amount == 0 {
			return errors.New("amount cannot be zero in strict mode")
		}
	}

	return nil
}

// ValidateBatch validates a batch of records
func (v *FieldValidator) ValidateBatch(records []*models.ETCMeisai) []error {
	var errors []error

	if records == nil {
		return []error{fmt.Errorf("records slice is nil")}
	}

	if len(records) == 0 {
		return []error{fmt.Errorf("no records to validate")}
	}

	for i, record := range records {
		if err := v.ValidateRecord(record); err != nil {
			errors = append(errors, fmt.Errorf("record %d: %w", i, err))
		}
	}

	return errors
}

// SanitizeInput sanitizes input strings
func (v *FieldValidator) SanitizeInput(input string) string {
	// Remove leading and trailing whitespace
	sanitized := strings.TrimSpace(input)

	// Remove tabs and newlines
	sanitized = strings.ReplaceAll(sanitized, "\t", "")
	sanitized = strings.ReplaceAll(sanitized, "\n", "")
	sanitized = strings.ReplaceAll(sanitized, "\r", "")

	return sanitized
}

// ValidateBusinessRules validates business logic rules
func (v *FieldValidator) ValidateBusinessRules(record *models.ETCMeisai) error {
	if record == nil {
		return errors.New("record is nil")
	}

	// Check for high amounts with short distances (same area ICs)
	if record.Amount > 5000 && strings.Contains(record.EntryIC, "IC") && strings.Contains(record.ExitIC, "IC") {
		// Check if both ICs are in the Tokyo area
		tokyoAreas := []string{"東京", "品川", "新宿", "渋谷", "池袋"}
		entryInTokyo := false
		exitInTokyo := false

		for _, area := range tokyoAreas {
			if strings.Contains(record.EntryIC, area) {
				entryInTokyo = true
			}
			if strings.Contains(record.ExitIC, area) {
				exitInTokyo = true
			}
		}

		if entryInTokyo && exitInTokyo && record.Amount > 5000 {
			return errors.New("amount seems too high for the distance")
		}
	}

	return nil
}

// GetValidationSummary provides a summary of validation results for a batch
func (v *FieldValidator) GetValidationSummary(records []*models.ETCMeisai) ValidationSummary {
	summary := ValidationSummary{
		TotalRecords:   len(records),
		ValidRecords:   0,
		InvalidRecords: 0,
		ErrorsByType:   make(map[string]int),
	}

	for _, record := range records {
		if err := v.ValidateRecord(record); err != nil {
			summary.InvalidRecords++

			// Categorize error types
			errMsg := err.Error()
			if strings.Contains(errMsg, "date") {
				summary.ErrorsByType["date_errors"]++
			} else if strings.Contains(errMsg, "time") {
				summary.ErrorsByType["time_errors"]++
			} else if strings.Contains(errMsg, "amount") {
				summary.ErrorsByType["amount_errors"]++
			} else if strings.Contains(errMsg, "IC") {
				summary.ErrorsByType["ic_errors"]++
			} else {
				summary.ErrorsByType["other_errors"]++
			}
		} else {
			summary.ValidRecords++
		}
	}

	return summary
}
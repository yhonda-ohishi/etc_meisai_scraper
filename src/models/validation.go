package models

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

// ValidationError represents a validation error with field information
type ValidationError struct {
	Field   string `json:"field"`
	Value   interface{} `json:"value,omitempty"`
	Message string `json:"message"`
	Code    string `json:"code"`
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("validation error on field '%s': %s", e.Field, e.Message)
}

// ValidationResult holds the result of validation
type ValidationResult struct {
	Valid  bool               `json:"valid"`
	Errors []ValidationError  `json:"errors,omitempty"`
}

// AddError adds a validation error to the result
func (vr *ValidationResult) AddError(field, message, code string, value interface{}) {
	vr.Valid = false
	vr.Errors = append(vr.Errors, ValidationError{
		Field:   field,
		Value:   value,
		Message: message,
		Code:    code,
	})
}

// ValidateETCMeisai performs comprehensive validation on ETC record
func ValidateETCMeisai(etc *ETCMeisai) *ValidationResult {
	result := &ValidationResult{Valid: true}

	// Required field validations
	if etc.UseDate.IsZero() {
		result.AddError("use_date", "UseDate is required", "REQUIRED", nil)
	}

	if etc.Amount <= 0 {
		result.AddError("amount", "Amount must be positive", "POSITIVE_REQUIRED", etc.Amount)
	}

	if strings.TrimSpace(etc.EntryIC) == "" {
		result.AddError("entry_ic", "EntryIC is required", "REQUIRED", etc.EntryIC)
	}

	if strings.TrimSpace(etc.ExitIC) == "" {
		result.AddError("exit_ic", "ExitIC is required", "REQUIRED", etc.ExitIC)
	}

	// Date validations
	if !etc.UseDate.IsZero() {
		// Check if date is not in the future
		if etc.UseDate.After(time.Now()) {
			result.AddError("use_date", "UseDate cannot be in the future", "FUTURE_DATE", etc.UseDate)
		}

		// Check if date is not too old (older than 2 years)
		twoYearsAgo := time.Now().AddDate(-2, 0, 0)
		if etc.UseDate.Before(twoYearsAgo) {
			result.AddError("use_date", "UseDate cannot be older than 2 years", "TOO_OLD", etc.UseDate)
		}
	}

	// Time format validation
	if etc.UseTime != "" {
		if !isValidTimeFormat(etc.UseTime) {
			result.AddError("use_time", "UseTime must be in HH:MM format", "INVALID_FORMAT", etc.UseTime)
		}
	}

	// ETC Number validation
	if len(etc.ETCNumber) > 20 {
		result.AddError("etc_number", "ETCNumber cannot exceed 20 characters", "MAX_LENGTH", etc.ETCNumber)
	}

	if etc.ETCNumber != "" && !isValidETCNumber(etc.ETCNumber) {
		result.AddError("etc_number", "ETCNumber contains invalid characters", "INVALID_FORMAT", etc.ETCNumber)
	}

	// Car number validation
	if len(etc.CarNumber) > 20 {
		result.AddError("car_number", "CarNumber cannot exceed 20 characters", "MAX_LENGTH", etc.CarNumber)
	}

	// IC name validation
	if len(etc.EntryIC) > 100 {
		result.AddError("entry_ic", "EntryIC cannot exceed 100 characters", "MAX_LENGTH", etc.EntryIC)
	}

	if len(etc.ExitIC) > 100 {
		result.AddError("exit_ic", "ExitIC cannot exceed 100 characters", "MAX_LENGTH", etc.ExitIC)
	}

	// Amount range validation
	if etc.Amount > 100000 { // 10万円以上は異常値として扱う
		result.AddError("amount", "Amount seems unusually high", "SUSPICIOUS_VALUE", etc.Amount)
	}

	return result
}

// ValidateETCMeisaiMapping performs validation on mapping record
func ValidateETCMeisaiMapping(mapping *ETCMeisaiMapping) *ValidationResult {
	result := &ValidationResult{Valid: true}

	// Required field validations
	if mapping.ETCMeisaiID <= 0 {
		result.AddError("etc_meisai_id", "ETCMeisaiID must be positive", "POSITIVE_REQUIRED", mapping.ETCMeisaiID)
	}

	if strings.TrimSpace(mapping.DTakoRowID) == "" {
		result.AddError("dtako_row_id", "DTakoRowID is required", "REQUIRED", mapping.DTakoRowID)
	}

	// MappingType validation
	validMappingTypes := map[string]bool{"auto": true, "manual": true}
	if !validMappingTypes[mapping.MappingType] {
		result.AddError("mapping_type", "MappingType must be 'auto' or 'manual'", "INVALID_ENUM", mapping.MappingType)
	}

	// Confidence validation
	if mapping.Confidence < 0 || mapping.Confidence > 1 {
		result.AddError("confidence", "Confidence must be between 0 and 1", "RANGE_ERROR", mapping.Confidence)
	}

	// Notes length validation
	if len(mapping.Notes) > 500 {
		result.AddError("notes", "Notes cannot exceed 500 characters", "MAX_LENGTH", mapping.Notes)
	}

	// DTako Row ID format validation
	if mapping.DTakoRowID != "" && !isValidDTakoRowID(mapping.DTakoRowID) {
		result.AddError("dtako_row_id", "DTakoRowID has invalid format", "INVALID_FORMAT", mapping.DTakoRowID)
	}

	return result
}

// ValidateETCImportBatch performs validation on import batch
func ValidateETCImportBatch(batch *ETCImportBatch) *ValidationResult {
	result := &ValidationResult{Valid: true}

	// Required field validations
	if strings.TrimSpace(batch.FileName) == "" {
		result.AddError("file_name", "FileName is required", "REQUIRED", batch.FileName)
	}

	if batch.TotalRecords < 0 {
		result.AddError("total_records", "TotalRecords cannot be negative", "NON_NEGATIVE", batch.TotalRecords)
	}

	// Status validation
	validStatuses := map[string]bool{
		"pending":    true,
		"processing": true,
		"completed":  true,
		"failed":     true,
		"cancelled":  true,
	}
	if !validStatuses[batch.Status] {
		result.AddError("status", "Invalid status value", "INVALID_ENUM", batch.Status)
	}

	// File name validation
	if len(batch.FileName) > 255 {
		result.AddError("file_name", "FileName cannot exceed 255 characters", "MAX_LENGTH", batch.FileName)
	}

	// Count consistency validation
	if batch.ProcessedCount > batch.TotalRecords {
		result.AddError("processed_count", "ProcessedCount cannot exceed TotalRecords", "LOGICAL_ERROR", batch.ProcessedCount)
	}

	if batch.CreatedCount > batch.ProcessedCount {
		result.AddError("created_count", "CreatedCount cannot exceed ProcessedCount", "LOGICAL_ERROR", batch.CreatedCount)
	}

	// Time validation
	if batch.StartTime != nil && batch.CompleteTime != nil {
		if batch.CompleteTime.Before(*batch.StartTime) {
			result.AddError("complete_time", "CompleteTime cannot be before StartTime", "LOGICAL_ERROR", batch.CompleteTime)
		}
	}

	return result
}

// Helper functions for validation

// isValidTimeFormat checks if time string is in HH:MM format
func isValidTimeFormat(timeStr string) bool {
	matched, _ := regexp.MatchString(`^([01]?[0-9]|2[0-3]):[0-5][0-9]$`, timeStr)
	return matched
}

// isValidETCNumber checks if ETC number has valid format (digits only)
func isValidETCNumber(etcNumber string) bool {
	matched, _ := regexp.MatchString(`^[0-9]+$`, etcNumber)
	return matched
}

// isValidDTakoRowID checks if DTako row ID has valid format
func isValidDTakoRowID(rowID string) bool {
	// DTako row IDs are typically alphanumeric with possible hyphens
	matched, _ := regexp.MatchString(`^[A-Za-z0-9\-_]+$`, rowID)
	return matched && len(rowID) <= 50
}

// BatchValidationOptions contains options for batch validation
type BatchValidationOptions struct {
	StrictMode    bool `json:"strict_mode"`     // Fail on any warning
	SkipDuplicates bool `json:"skip_duplicates"` // Skip duplicate hash validation
	MaxErrors     int  `json:"max_errors"`      // Stop after N errors
}

// ValidateETCMeisaiBatch validates a batch of ETC records
func ValidateETCMeisaiBatch(records []*ETCMeisai, options *BatchValidationOptions) map[int]*ValidationResult {
	if options == nil {
		options = &BatchValidationOptions{
			StrictMode:    false,
			SkipDuplicates: false,
			MaxErrors:     100,
		}
	}

	results := make(map[int]*ValidationResult)
	hashMap := make(map[string]int) // Track hashes for duplicate detection
	errorCount := 0

	for i, record := range records {
		if options.MaxErrors > 0 && errorCount >= options.MaxErrors {
			break
		}

		result := ValidateETCMeisai(record)

		// Check for duplicate hashes within the batch
		if !options.SkipDuplicates && record.Hash != "" {
			if existingIndex, exists := hashMap[record.Hash]; exists {
				result.AddError("hash", fmt.Sprintf("Duplicate hash found with record at index %d", existingIndex), "DUPLICATE_HASH", record.Hash)
			} else {
				hashMap[record.Hash] = i
			}
		}

		if !result.Valid {
			errorCount++
		}

		results[i] = result
	}

	return results
}

// ValidationSummary provides a summary of validation results
type ValidationSummary struct {
	TotalRecords   int                     `json:"total_records"`
	ValidRecords   int                     `json:"valid_records"`
	InvalidRecords int                     `json:"invalid_records"`
	ErrorsByField  map[string]int          `json:"errors_by_field"`
	ErrorsByCode   map[string]int          `json:"errors_by_code"`
	FirstErrors    []ValidationError       `json:"first_errors,omitempty"` // First few errors for quick review
}

// SummarizeValidation creates a summary from validation results
func SummarizeValidation(results map[int]*ValidationResult, maxFirstErrors int) *ValidationSummary {
	summary := &ValidationSummary{
		TotalRecords:   len(results),
		ValidRecords:   0,
		InvalidRecords: 0,
		ErrorsByField:  make(map[string]int),
		ErrorsByCode:   make(map[string]int),
		FirstErrors:    make([]ValidationError, 0, maxFirstErrors),
	}

	for _, result := range results {
		if result.Valid {
			summary.ValidRecords++
		} else {
			summary.InvalidRecords++

			for _, err := range result.Errors {
				summary.ErrorsByField[err.Field]++
				summary.ErrorsByCode[err.Code]++

				if len(summary.FirstErrors) < maxFirstErrors {
					summary.FirstErrors = append(summary.FirstErrors, err)
				}
			}
		}
	}

	return summary
}
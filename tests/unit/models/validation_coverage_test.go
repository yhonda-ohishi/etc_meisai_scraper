package models_test

import (
	"testing"
	"time"

	"github.com/yhonda-ohishi/etc_meisai/src/models"
	"github.com/yhonda-ohishi/etc_meisai/tests/helpers"
)

// TestValidateETCMeisaiBatch tests batch validation functionality
func TestValidateETCMeisaiBatch(t *testing.T) {
	// Create test records
	validRecord := &models.ETCMeisai{
		UseDate:   time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		UseTime:   "14:30",
		EntryIC:   "東京IC",
		ExitIC:    "大阪IC",
		Amount:    1000,
		CarNumber: "品川123",
		ETCNumber: "1234567890123456",
		Hash:      "hash1",
	}

	invalidRecord := &models.ETCMeisai{
		UseDate: time.Time{}, // Invalid zero date
		Amount:  -100,        // Invalid negative amount
		Hash:    "hash2",
	}

	duplicateRecord := &models.ETCMeisai{
		UseDate:   time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		UseTime:   "14:30",
		EntryIC:   "東京IC",
		ExitIC:    "大阪IC",
		Amount:    1000,
		CarNumber: "品川123",
		ETCNumber: "1234567890123456",
		Hash:      "hash1", // Same hash as validRecord
	}

	// Test with default options (nil)
	records := []*models.ETCMeisai{validRecord, invalidRecord, duplicateRecord}
	results := models.ValidateETCMeisaiBatch(records, nil)

	helpers.AssertLen(t, results, 3)
	helpers.AssertTrue(t, results[0].Valid)
	helpers.AssertFalse(t, results[1].Valid)
	helpers.AssertFalse(t, results[2].Valid) // Should fail due to duplicate hash

	// Test with custom options
	options := &models.BatchValidationOptions{
		StrictMode:     true,
		SkipDuplicates: false,
		MaxErrors:      1,
	}

	results = models.ValidateETCMeisaiBatch(records, options)
	helpers.AssertLen(t, results, 2) // Should stop after maxErrors reached

	// Test with SkipDuplicates enabled
	options.SkipDuplicates = true
	results = models.ValidateETCMeisaiBatch(records, options)
	helpers.AssertTrue(t, results[2].Valid) // Should pass now that duplicates are skipped
}

// TestSummarizeValidation tests validation summary functionality
func TestSummarizeValidation(t *testing.T) {
	// Create mock validation results
	results := map[int]*models.ValidationResult{
		0: {
			Valid:  true,
			Errors: []models.ValidationError{},
		},
		1: {
			Valid: false,
			Errors: []models.ValidationError{
				{Field: "use_date", Code: "REQUIRED", Message: "UseDate is required"},
				{Field: "amount", Code: "POSITIVE_REQUIRED", Message: "Amount must be positive"},
			},
		},
		2: {
			Valid: false,
			Errors: []models.ValidationError{
				{Field: "use_date", Code: "REQUIRED", Message: "UseDate is required"},
			},
		},
	}

	summary := models.SummarizeValidation(results, 5)

	helpers.AssertEqual(t, 3, summary.TotalRecords)
	helpers.AssertEqual(t, 1, summary.ValidRecords)
	helpers.AssertEqual(t, 2, summary.InvalidRecords)
	helpers.AssertEqual(t, 2, summary.ErrorsByField["use_date"])
	helpers.AssertEqual(t, 1, summary.ErrorsByField["amount"])
	helpers.AssertEqual(t, 2, summary.ErrorsByCode["REQUIRED"])
	helpers.AssertEqual(t, 1, summary.ErrorsByCode["POSITIVE_REQUIRED"])
	helpers.AssertLen(t, summary.FirstErrors, 3)
}

// TestValidateETCMeisaiMapping tests ETC mapping validation
func TestValidateETCMeisaiMapping(t *testing.T) {
	tests := []struct {
		name    string
		mapping *models.ETCMeisaiMapping
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid mapping",
			mapping: &models.ETCMeisaiMapping{
				ETCMeisaiID:   1,
				DTakoRowID:    "DTAKO-123",
				MappingType:   "auto",
				Confidence:    0.95,
				Notes:         "High confidence match",
			},
			wantErr: false,
		},
		{
			name: "zero ETC meisai ID",
			mapping: &models.ETCMeisaiMapping{
				ETCMeisaiID:   0,
				DTakoRowID:    "DTAKO-123",
				MappingType:   "auto",
				Confidence:    0.95,
			},
			wantErr: true,
			errMsg:  "ETCMeisaiID must be positive",
		},
		{
			name: "empty DTako row ID",
			mapping: &models.ETCMeisaiMapping{
				ETCMeisaiID:   1,
				DTakoRowID:    "",
				MappingType:   "auto",
				Confidence:    0.95,
			},
			wantErr: true,
			errMsg:  "DTakoRowID is required",
		},
		{
			name: "invalid mapping type",
			mapping: &models.ETCMeisaiMapping{
				ETCMeisaiID:   1,
				DTakoRowID:    "DTAKO-123",
				MappingType:   "invalid",
				Confidence:    0.95,
			},
			wantErr: true,
			errMsg:  "MappingType must be 'auto' or 'manual'",
		},
		{
			name: "confidence below range",
			mapping: &models.ETCMeisaiMapping{
				ETCMeisaiID:   1,
				DTakoRowID:    "DTAKO-123",
				MappingType:   "auto",
				Confidence:    -0.1,
			},
			wantErr: true,
			errMsg:  "Confidence must be between 0 and 1",
		},
		{
			name: "confidence above range",
			mapping: &models.ETCMeisaiMapping{
				ETCMeisaiID:   1,
				DTakoRowID:    "DTAKO-123",
				MappingType:   "auto",
				Confidence:    1.5,
			},
			wantErr: true,
			errMsg:  "Confidence must be between 0 and 1",
		},
		{
			name: "notes too long",
			mapping: &models.ETCMeisaiMapping{
				ETCMeisaiID:   1,
				DTakoRowID:    "DTAKO-123",
				MappingType:   "auto",
				Confidence:    0.95,
				Notes:         "Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum. Sed ut perspiciatis unde omnis iste natus error sit voluptatem accusantium doloremque laudantium.",
			},
			wantErr: true,
			errMsg:  "Notes cannot exceed 500 characters",
		},
		{
			name: "invalid DTako row ID format",
			mapping: &models.ETCMeisaiMapping{
				ETCMeisaiID:   1,
				DTakoRowID:    "invalid@#$%^&*()",
				MappingType:   "auto",
				Confidence:    0.95,
			},
			wantErr: true,
			errMsg:  "DTakoRowID has invalid format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := models.ValidateETCMeisaiMapping(tt.mapping)

			if tt.wantErr {
				helpers.AssertFalse(t, result.Valid)
				helpers.AssertTrue(t, len(result.Errors) > 0)
				if tt.errMsg != "" {
					found := false
					for _, err := range result.Errors {
						if err.Message == tt.errMsg {
							found = true
							break
						}
					}
					helpers.AssertTrue(t, found)
				}
			} else {
				helpers.AssertTrue(t, result.Valid)
				helpers.AssertLen(t, result.Errors, 0)
			}
		})
	}
}

// TestValidateETCImportBatch tests import batch validation
func TestValidateETCImportBatch(t *testing.T) {
	tests := []struct {
		name    string
		batch   *models.ETCImportBatch
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid batch",
			batch: &models.ETCImportBatch{
				FileName:       "test.csv",
				TotalRecords:   100,
				ProcessedCount: 50,
				CreatedCount:   45,
				Status:         "processing",
				StartTime:      &time.Time{},
			},
			wantErr: false,
		},
		{
			name: "empty file name",
			batch: &models.ETCImportBatch{
				TotalRecords: 100,
				Status:       "processing",
			},
			wantErr: true,
			errMsg:  "FileName is required",
		},
		{
			name: "negative total records",
			batch: &models.ETCImportBatch{
				FileName:     "test.csv",
				TotalRecords: -10,
				Status:       "processing",
			},
			wantErr: true,
			errMsg:  "TotalRecords cannot be negative",
		},
		{
			name: "invalid status",
			batch: &models.ETCImportBatch{
				FileName:     "test.csv",
				TotalRecords: 100,
				Status:       "invalid_status",
			},
			wantErr: true,
			errMsg:  "Invalid status value",
		},
		{
			name: "file name too long",
			batch: &models.ETCImportBatch{
				FileName:     "this_is_a_very_long_file_name_that_exceeds_the_maximum_allowed_length_of_two_hundred_fifty_five_characters_which_should_cause_validation_to_fail_because_it_is_way_too_long_for_a_reasonable_file_name_in_any_practical_system_implementation.csv",
				TotalRecords: 100,
				Status:       "processing",
			},
			wantErr: true,
			errMsg:  "FileName cannot exceed 255 characters",
		},
		{
			name: "processed count exceeds total",
			batch: &models.ETCImportBatch{
				FileName:       "test.csv",
				TotalRecords:   100,
				ProcessedCount: 150,
				Status:         "processing",
			},
			wantErr: true,
			errMsg:  "ProcessedCount cannot exceed TotalRecords",
		},
		{
			name: "created count exceeds processed",
			batch: &models.ETCImportBatch{
				FileName:       "test.csv",
				TotalRecords:   100,
				ProcessedCount: 50,
				CreatedCount:   75,
				Status:         "processing",
			},
			wantErr: true,
			errMsg:  "CreatedCount cannot exceed ProcessedCount",
		},
		{
			name: "complete time before start time",
			batch: &models.ETCImportBatch{
				FileName:     "test.csv",
				TotalRecords: 100,
				Status:       "completed",
				StartTime:    timePtr(time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)),
				CompleteTime: timePtr(time.Date(2025, 1, 1, 11, 0, 0, 0, time.UTC)),
			},
			wantErr: true,
			errMsg:  "CompleteTime cannot be before StartTime",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := models.ValidateETCImportBatch(tt.batch)

			if tt.wantErr {
				helpers.AssertFalse(t, result.Valid)
				helpers.AssertTrue(t, len(result.Errors) > 0)
				if tt.errMsg != "" {
					found := false
					for _, err := range result.Errors {
						if err.Message == tt.errMsg {
							found = true
							break
						}
					}
					helpers.AssertTrue(t, found)
				}
			} else {
				helpers.AssertTrue(t, result.Valid)
				helpers.AssertLen(t, result.Errors, 0)
			}
		})
	}
}

// TestValidationError tests ValidationError methods
func TestValidationError(t *testing.T) {
	err := models.ValidationError{
		Field:   "test_field",
		Value:   "test_value",
		Message: "Test error message",
		Code:    "TEST_ERROR",
	}

	errorMsg := err.Error()
	helpers.AssertContains(t, errorMsg, "validation error on field 'test_field'")
	helpers.AssertContains(t, errorMsg, "Test error message")
}

// TestValidationResult tests ValidationResult methods
func TestValidationResult(t *testing.T) {
	result := &models.ValidationResult{Valid: true}

	// Test AddError
	result.AddError("test_field", "Test error", "TEST_CODE", "test_value")

	helpers.AssertFalse(t, result.Valid)
	helpers.AssertLen(t, result.Errors, 1)
	helpers.AssertEqual(t, "test_field", result.Errors[0].Field)
	helpers.AssertEqual(t, "Test error", result.Errors[0].Message)
	helpers.AssertEqual(t, "TEST_CODE", result.Errors[0].Code)
	helpers.AssertEqual(t, "test_value", result.Errors[0].Value)
}

// Helper function for time pointers
func timePtr(t time.Time) *time.Time {
	return &t
}


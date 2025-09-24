package models_test

import (
	"testing"
	"time"

	"github.com/yhonda-ohishi/etc_meisai/src/models"
	"github.com/yhonda-ohishi/etc_meisai/tests/helpers"
)

func TestValidationError_Error(t *testing.T) {
	validationErr := models.ValidationError{
		Field:   "etc_number",
		Value:   "invalid",
		Message: "Invalid ETC number format",
		Code:    "INVALID_FORMAT",
	}

	expected := "validation error on field 'etc_number': Invalid ETC number format"
	helpers.AssertEqual(t, expected, validationErr.Error())
}

func TestValidationResult_AddError(t *testing.T) {
	result := &models.ValidationResult{Valid: true}

	// Initially valid
	helpers.AssertTrue(t, result.Valid)
	helpers.AssertLen(t, result.Errors, 0)

	// Add first error
	result.AddError("test_field", "Test error", "TEST_ERROR", "test_value")

	helpers.AssertFalse(t, result.Valid)
	helpers.AssertLen(t, result.Errors, 1)

	err := result.Errors[0]
	helpers.AssertEqual(t, "test_field", err.Field)
	helpers.AssertEqual(t, "Test error", err.Message)
	helpers.AssertEqual(t, "TEST_ERROR", err.Code)
	helpers.AssertEqual(t, "test_value", err.Value)

	// Add second error
	result.AddError("another_field", "Another error", "ANOTHER_ERROR", nil)
	helpers.AssertLen(t, result.Errors, 2)
}

func TestValidateETCMeisai(t *testing.T) {
	tests := []struct {
		name          string
		record        *models.ETCMeisai
		expectedValid bool
		expectedErrors int
		checkError    func(t *testing.T, result *models.ValidationResult)
	}{
		{
			name:          "nil record",
			record:        nil,
			expectedValid: false,
			expectedErrors: 1,
			checkError: func(t *testing.T, result *models.ValidationResult) {
				helpers.AssertEqual(t, "etc", result.Errors[0].Field)
				helpers.AssertEqual(t, "NIL_RECORD", result.Errors[0].Code)
			},
		},
		{
			name: "valid record",
			record: &models.ETCMeisai{
				UseDate:   time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
				UseTime:   "09:00",
				EntryIC:   "東京IC",
				ExitIC:    "大阪IC",
				Amount:    1000,
				CarNumber: "品川123あ1234",
				ETCNumber: "1234567890",
			},
			expectedValid: true,
			expectedErrors: 0,
		},
		{
			name: "missing required fields",
			record: &models.ETCMeisai{
				Amount: int32(1000),
			},
			expectedValid: false,
			expectedErrors: 3, // UseDate, EntryIC, ExitIC
			checkError: func(t *testing.T, result *models.ValidationResult) {
				// Check that all required field errors are present
				fields := make(map[string]bool)
				for _, err := range result.Errors {
					fields[err.Field] = true
				}
				helpers.AssertTrue(t, fields["use_date"])
				helpers.AssertTrue(t, fields["entry_ic"])
				helpers.AssertTrue(t, fields["exit_ic"])
			},
		},
		{
			name: "invalid amount (zero)",
			record: &models.ETCMeisai{
				UseDate:   time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
				UseTime:   "09:00",
				EntryIC:   "東京IC",
				ExitIC:    "大阪IC",
				Amount:    int32(0),
				CarNumber: "品川123あ1234",
				ETCNumber: "1234567890",
			},
			expectedValid: false,
			expectedErrors: 1,
			checkError: func(t *testing.T, result *models.ValidationResult) {
				helpers.AssertEqual(t, "amount", result.Errors[0].Field)
				helpers.AssertContains(t, result.Errors[0].Message, "positive")
			},
		},
		{
			name: "invalid amount (negative)",
			record: &models.ETCMeisai{
				UseDate:   time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
				UseTime:   "09:00",
				EntryIC:   "東京IC",
				ExitIC:    "大阪IC",
				Amount:    int32(-100),
				CarNumber: "品川123あ1234",
				ETCNumber: "1234567890",
			},
			expectedValid: false,
			expectedErrors: 1,
			checkError: func(t *testing.T, result *models.ValidationResult) {
				helpers.AssertEqual(t, "amount", result.Errors[0].Field)
			},
		},
		{
			name: "invalid time format",
			record: &models.ETCMeisai{
				UseDate:   time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
				UseTime:   "25:00", // Invalid hour
				EntryIC:   "東京IC",
				ExitIC:    "大阪IC",
				Amount:    1000,
				CarNumber: "品川123あ1234",
				ETCNumber: "1234567890",
			},
			expectedValid: false,
			expectedErrors: 1,
			checkError: func(t *testing.T, result *models.ValidationResult) {
				helpers.AssertEqual(t, "use_time", result.Errors[0].Field)
				helpers.AssertContains(t, result.Errors[0].Message, "format")
			},
		},
		{
			name: "future date",
			record: &models.ETCMeisai{
				UseDate:   time.Now().Add(24 * time.Hour), // Tomorrow
				UseTime:   "09:00",
				EntryIC:   "東京IC",
				ExitIC:    "大阪IC",
				Amount:    1000,
				CarNumber: "品川123あ1234",
				ETCNumber: "1234567890",
			},
			expectedValid: false,
			expectedErrors: 1,
			checkError: func(t *testing.T, result *models.ValidationResult) {
				helpers.AssertEqual(t, "use_date", result.Errors[0].Field)
				helpers.AssertContains(t, result.Errors[0].Message, "future")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := models.ValidateETCMeisai(tt.record)

			helpers.AssertEqual(t, tt.expectedValid, result.Valid)
			helpers.AssertLen(t, result.Errors, tt.expectedErrors)

			if tt.checkError != nil {
				tt.checkError(t, result)
			}
		})
	}
}

func TestValidateETCNumber(t *testing.T) {
	tests := []struct {
		name      string
		etcNumber string
		wantValid bool
	}{
		{
			name:      "valid 10-digit number",
			etcNumber: "1234567890",
			wantValid: true,
		},
		{
			name:      "valid 16-digit number",
			etcNumber: "1234567890123456",
			wantValid: true,
		},
		{
			name:      "empty string",
			etcNumber: "",
			wantValid: false,
		},
		{
			name:      "too short",
			etcNumber: "123456789",
			wantValid: false,
		},
		{
			name:      "too long",
			etcNumber: "12345678901234567",
			wantValid: false,
		},
		{
			name:      "contains letters",
			etcNumber: "123456789a",
			wantValid: false,
		},
		{
			name:      "contains special characters",
			etcNumber: "1234567890-",
			wantValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid := models.ValidateETCNumber(tt.etcNumber)
			helpers.AssertEqual(t, tt.wantValid, valid)
		})
	}
}

func TestValidateTimeFormat(t *testing.T) {
	tests := []struct {
		name      string
		timeStr   string
		wantValid bool
	}{
		{
			name:      "valid HH:MM format",
			timeStr:   "09:30",
			wantValid: true,
		},
		{
			name:      "valid midnight",
			timeStr:   "00:00",
			wantValid: true,
		},
		{
			name:      "valid late hour",
			timeStr:   "23:59",
			wantValid: true,
		},
		{
			name:      "invalid hour (24)",
			timeStr:   "24:00",
			wantValid: false,
		},
		{
			name:      "invalid hour (25)",
			timeStr:   "25:30",
			wantValid: false,
		},
		{
			name:      "invalid minute (60)",
			timeStr:   "12:60",
			wantValid: false,
		},
		{
			name:      "invalid format (no colon)",
			timeStr:   "1230",
			wantValid: false,
		},
		{
			name:      "invalid format (single digit hour)",
			timeStr:   "9:30",
			wantValid: false,
		},
		{
			name:      "invalid format (single digit minute)",
			timeStr:   "09:3",
			wantValid: false,
		},
		{
			name:      "empty string",
			timeStr:   "",
			wantValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid := models.ValidateTimeFormat(tt.timeStr)
			helpers.AssertEqual(t, tt.wantValid, valid)
		})
	}
}

func TestValidateCarNumber(t *testing.T) {
	tests := []struct {
		name      string
		carNumber string
		wantValid bool
	}{
		{
			name:      "valid car number",
			carNumber: "品川123あ1234",
			wantValid: true,
		},
		{
			name:      "valid car number with different region",
			carNumber: "横浜456い5678",
			wantValid: true,
		},
		{
			name:      "empty string",
			carNumber: "",
			wantValid: false,
		},
		{
			name:      "too short",
			carNumber: "品川123",
			wantValid: false,
		},
		{
			name:      "invalid format",
			carNumber: "invalid123",
			wantValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid := models.ValidateCarNumber(tt.carNumber)
			helpers.AssertEqual(t, tt.wantValid, valid)
		})
	}
}

func TestSanitizeInput(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "normal text",
			input:    "東京IC",
			expected: "東京IC",
		},
		{
			name:     "text with leading/trailing spaces",
			input:    "  東京IC  ",
			expected: "東京IC",
		},
		{
			name:     "text with SQL injection attempt",
			input:    "東京IC'; DROP TABLE users; --",
			expected: "東京IC DROP TABLE users ", // Dangerous characters removed
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "whitespace only",
			input:    "   ",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := models.SanitizeInput(tt.input)
			helpers.AssertEqual(t, tt.expected, result)
		})
	}
}
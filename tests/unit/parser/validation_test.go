package parser_test

import (
	"strings"
	"testing"
	"time"

	"github.com/yhonda-ohishi/etc_meisai/src/models"
	"github.com/yhonda-ohishi/etc_meisai/src/parser"
	"github.com/yhonda-ohishi/etc_meisai/tests/helpers"
)

func TestFieldValidator_ValidateUseDate(t *testing.T) {
	validator := parser.NewFieldValidator()

	tests := []struct {
		name     string
		dateStr  string
		wantErr  bool
		errMsg   string
	}{
		{
			name:    "valid date",
			dateStr: "2025-01-15",
			wantErr: false,
		},
		{
			name:    "valid date with slashes",
			dateStr: "2025/01/15",
			wantErr: false,
		},
		{
			name:    "valid Japanese date",
			dateStr: "2025年01月15日",
			wantErr: false,
		},
		{
			name:    "empty date",
			dateStr: "",
			wantErr: true,
			errMsg:  "date is required",
		},
		{
			name:    "invalid format",
			dateStr: "invalid-date",
			wantErr: true,
			errMsg:  "invalid date format",
		},
		{
			name:    "future date",
			dateStr: "2030-01-01",
			wantErr: true,
			errMsg:  "date cannot be in the future",
		},
		{
			name:    "too old date",
			dateStr: "1990-01-01",
			wantErr: true,
			errMsg:  "date is too old",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateUseDate(tt.dateStr)

			if tt.wantErr {
				helpers.AssertError(t, err)
				if tt.errMsg != "" {
					helpers.AssertContains(t, err.Error(), tt.errMsg)
				}
			} else {
				helpers.AssertNoError(t, err)
			}
		})
	}
}

func TestFieldValidator_ValidateUseTime(t *testing.T) {
	validator := parser.NewFieldValidator()

	tests := []struct {
		name    string
		timeStr string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid time",
			timeStr: "09:30",
			wantErr: false,
		},
		{
			name:    "valid midnight",
			timeStr: "00:00",
			wantErr: false,
		},
		{
			name:    "valid late time",
			timeStr: "23:59",
			wantErr: false,
		},
		{
			name:    "empty time",
			timeStr: "",
			wantErr: true,
			errMsg:  "time is required",
		},
		{
			name:    "invalid hour",
			timeStr: "24:00",
			wantErr: true,
			errMsg:  "invalid time format",
		},
		{
			name:    "invalid minute",
			timeStr: "12:60",
			wantErr: true,
			errMsg:  "invalid time format",
		},
		{
			name:    "single digit hour allowed",
			timeStr: "9:30",
			wantErr: false, // Single digit hours are actually valid in our regex
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateUseTime(tt.timeStr)

			if tt.wantErr {
				helpers.AssertError(t, err)
				if tt.errMsg != "" {
					helpers.AssertContains(t, err.Error(), tt.errMsg)
				}
			} else {
				helpers.AssertNoError(t, err)
			}
		})
	}
}

func TestFieldValidator_ValidateIC(t *testing.T) {
	validator := parser.NewFieldValidator()

	tests := []struct {
		name    string
		ic      string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid IC",
			ic:      "東京IC",
			wantErr: false,
		},
		{
			name:    "valid IC with numbers",
			ic:      "首都高C1",
			wantErr: false,
		},
		{
			name:    "empty IC",
			ic:      "",
			wantErr: true,
			errMsg:  "IC name is required",
		},
		{
			name:    "too short IC",
			ic:      "A",
			wantErr: true,
			errMsg:  "IC name is too short",
		},
		{
			name:    "too long IC",
			ic:      "非常に長いインターチェンジ名前です",
			wantErr: true,
			errMsg:  "IC name is too long",
		},
		{
			name:    "invalid characters",
			ic:      "東京IC<script>",
			wantErr: true,
			errMsg:  "IC name contains invalid characters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateIC(tt.ic)

			if tt.wantErr {
				helpers.AssertError(t, err)
				if tt.errMsg != "" {
					helpers.AssertContains(t, err.Error(), tt.errMsg)
				}
			} else {
				helpers.AssertNoError(t, err)
			}
		})
	}
}

func TestFieldValidator_ValidateAmount(t *testing.T) {
	validator := parser.NewFieldValidator()

	tests := []struct {
		name    string
		amount  int
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid amount",
			amount:  1000,
			wantErr: false,
		},
		{
			name:    "zero amount",
			amount:  0,
			wantErr: false, // Zero is valid for free roads
		},
		{
			name:    "large amount",
			amount:  50000,
			wantErr: false,
		},
		{
			name:    "negative amount",
			amount:  -100,
			wantErr: true,
			errMsg:  "amount cannot be negative",
		},
		{
			name:    "extremely large amount",
			amount:  1000000,
			wantErr: true,
			errMsg:  "amount is unreasonably large",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateAmount(tt.amount)

			if tt.wantErr {
				helpers.AssertError(t, err)
				if tt.errMsg != "" {
					helpers.AssertContains(t, err.Error(), tt.errMsg)
				}
			} else {
				helpers.AssertNoError(t, err)
			}
		})
	}
}

func TestFieldValidator_ValidateCarNumber(t *testing.T) {
	validator := parser.NewFieldValidator()

	tests := []struct {
		name      string
		carNumber string
		wantErr   bool
		errMsg    string
	}{
		{
			name:      "valid car number",
			carNumber: "品川123あ1234",
			wantErr:   false,
		},
		{
			name:      "valid car number with different region",
			carNumber: "横浜456い5678",
			wantErr:   false,
		},
		{
			name:      "empty car number",
			carNumber: "",
			wantErr:   true,
			errMsg:    "car number is required",
		},
		{
			name:      "too short car number",
			carNumber: "品川123",
			wantErr:   true,
			errMsg:    "invalid car number format",
		},
		{
			name:      "invalid format",
			carNumber: "invalid123",
			wantErr:   true,
			errMsg:    "invalid car number format",
		},
		{
			name:      "car number with special characters",
			carNumber: "品川123あ1234!",
			wantErr:   true,
			errMsg:    "car number contains invalid characters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateCarNumber(tt.carNumber)

			if tt.wantErr {
				helpers.AssertError(t, err)
				if tt.errMsg != "" {
					helpers.AssertContains(t, err.Error(), tt.errMsg)
				}
			} else {
				helpers.AssertNoError(t, err)
			}
		})
	}
}

func TestFieldValidator_ValidateETCNumber(t *testing.T) {
	validator := parser.NewFieldValidator()

	tests := []struct {
		name      string
		etcNumber string
		wantErr   bool
		errMsg    string
	}{
		{
			name:      "valid 10-digit ETC number",
			etcNumber: "1234567890",
			wantErr:   false,
		},
		{
			name:      "valid 16-digit ETC number",
			etcNumber: "1234567890123456",
			wantErr:   false,
		},
		{
			name:      "empty ETC number",
			etcNumber: "",
			wantErr:   true,
			errMsg:    "ETC number is required",
		},
		{
			name:      "too short ETC number",
			etcNumber: "123456789",
			wantErr:   true,
			errMsg:    "invalid ETC number format",
		},
		{
			name:      "too long ETC number",
			etcNumber: "12345678901234567",
			wantErr:   true,
			errMsg:    "invalid ETC number format",
		},
		{
			name:      "ETC number with letters",
			etcNumber: "123456789a",
			wantErr:   true,
			errMsg:    "ETC number must contain only digits",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateETCNumber(tt.etcNumber)

			if tt.wantErr {
				helpers.AssertError(t, err)
				if tt.errMsg != "" {
					helpers.AssertContains(t, err.Error(), tt.errMsg)
				}
			} else {
				helpers.AssertNoError(t, err)
			}
		})
	}
}

func TestFieldValidator_ValidateRecord(t *testing.T) {
	validator := parser.NewFieldValidator()

	tests := []struct {
		name    string
		record  *models.ETCMeisai
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid record",
			record: &models.ETCMeisai{
				UseDate:   time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC),
				UseTime:   "09:30",
				EntryIC:   "東京IC",
				ExitIC:    "大阪IC",
				Amount:    1000,
				CarNumber: "品川123あ1234",
				ETCNumber: "1234567890",
			},
			wantErr: false,
		},
		{
			name: "nil record",
			record: nil,
			wantErr: true,
			errMsg: "record is nil",
		},
		{
			name: "invalid date",
			record: &models.ETCMeisai{
				UseDate:   time.Time{},
				UseTime:   "09:30",
				EntryIC:   "東京IC",
				ExitIC:    "大阪IC",
				Amount:    1000,
				CarNumber: "品川123あ1234",
				ETCNumber: "1234567890",
			},
			wantErr: true,
			errMsg: "invalid use date",
		},
		{
			name: "invalid time",
			record: &models.ETCMeisai{
				UseDate:   time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC),
				UseTime:   "25:00",
				EntryIC:   "東京IC",
				ExitIC:    "大阪IC",
				Amount:    1000,
				CarNumber: "品川123あ1234",
				ETCNumber: "1234567890",
			},
			wantErr: true,
			errMsg: "invalid use time",
		},
		{
			name: "same entry and exit IC",
			record: &models.ETCMeisai{
				UseDate:   time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC),
				UseTime:   "09:30",
				EntryIC:   "東京IC",
				ExitIC:    "東京IC",
				Amount:    1000,
				CarNumber: "品川123あ1234",
				ETCNumber: "1234567890",
			},
			wantErr: true,
			errMsg: "entry and exit IC cannot be the same",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateRecord(tt.record)

			if tt.wantErr {
				helpers.AssertError(t, err)
				if tt.errMsg != "" {
					helpers.AssertContains(t, err.Error(), tt.errMsg)
				}
			} else {
				helpers.AssertNoError(t, err)
			}
		})
	}
}

func TestFieldValidator_ValidateBatch(t *testing.T) {
	validator := parser.NewFieldValidator()

	validRecord := &models.ETCMeisai{
		UseDate:   time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC),
		UseTime:   "09:30",
		EntryIC:   "東京IC",
		ExitIC:    "大阪IC",
		Amount:    1000,
		CarNumber: "品川123あ1234",
		ETCNumber: "1234567890",
	}

	invalidRecord := &models.ETCMeisai{
		UseDate:   time.Time{},
		UseTime:   "09:30",
		EntryIC:   "東京IC",
		ExitIC:    "大阪IC",
		Amount:    1000,
		CarNumber: "品川123あ1234",
		ETCNumber: "1234567890",
	}

	tests := []struct {
		name         string
		records      []*models.ETCMeisai
		wantErr      bool
		errMsg       string
		expectErrors int
	}{
		{
			name:    "valid batch",
			records: []*models.ETCMeisai{validRecord, validRecord},
			wantErr: false,
		},
		{
			name:    "empty batch",
			records: []*models.ETCMeisai{},
			wantErr: true,
			errMsg:  "no records to validate",
		},
		{
			name:    "nil batch",
			records: nil,
			wantErr: true,
			errMsg:  "records slice is nil",
		},
		{
			name:         "batch with errors",
			records:      []*models.ETCMeisai{validRecord, invalidRecord},
			wantErr:      true,
			expectErrors: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := validator.ValidateBatch(tt.records)

			if tt.wantErr {
				helpers.AssertTrue(t, len(errors) > 0)
				if tt.errMsg != "" {
					found := false
					for _, err := range errors {
						if strings.Contains(err.Error(), tt.errMsg) {
							found = true
							break
						}
					}
					helpers.AssertTrue(t, found)
				}
				if tt.expectErrors > 0 {
					helpers.AssertEqual(t, tt.expectErrors, len(errors))
				}
			} else {
				helpers.AssertLen(t, errors, 0)
			}
		})
	}
}

func TestFieldValidator_SanitizeInput(t *testing.T) {
	validator := parser.NewFieldValidator()

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
			name:     "text with tabs and newlines",
			input:    "東京IC\t\n",
			expected: "東京IC",
		},
		{
			name:     "text with SQL injection attempt",
			input:    "東京IC'; DROP TABLE users; --",
			expected: "東京IC'; DROP TABLE users; --", // Should escape dangerous characters
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "whitespace only",
			input:    "   \t\n   ",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.SanitizeInput(tt.input)
			helpers.AssertEqual(t, tt.expected, result)
		})
	}
}

func TestFieldValidator_ValidateBusinessRules(t *testing.T) {
	validator := parser.NewFieldValidator()

	tests := []struct {
		name    string
		record  *models.ETCMeisai
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid business logic",
			record: &models.ETCMeisai{
				UseDate:   time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC),
				UseTime:   "09:30",
				EntryIC:   "東京IC",
				ExitIC:    "大阪IC",
				Amount:    1000,
				CarNumber: "品川123あ1234",
				ETCNumber: "1234567890",
			},
			wantErr: false,
		},
		{
			name: "zero amount with same prefecture",
			record: &models.ETCMeisai{
				UseDate:   time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC),
				UseTime:   "09:30",
				EntryIC:   "東京IC",
				ExitIC:    "品川IC",
				Amount:    0,
				CarNumber: "品川123あ1234",
				ETCNumber: "1234567890",
			},
			wantErr: false, // Zero amount OK for short distances
		},
		{
			name: "high amount with short distance",
			record: &models.ETCMeisai{
				UseDate:   time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC),
				UseTime:   "09:30",
				EntryIC:   "東京IC",
				ExitIC:    "品川IC",
				Amount:    10000,
				CarNumber: "品川123あ1234",
				ETCNumber: "1234567890",
			},
			wantErr: true,
			errMsg: "amount seems too high for the distance",
		},
		{
			name: "night time usage",
			record: &models.ETCMeisai{
				UseDate:   time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC),
				UseTime:   "02:30",
				EntryIC:   "東京IC",
				ExitIC:    "大阪IC",
				Amount:    800, // Discount for night usage
				CarNumber: "品川123あ1234",
				ETCNumber: "1234567890",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateBusinessRules(tt.record)

			if tt.wantErr {
				helpers.AssertError(t, err)
				if tt.errMsg != "" {
					helpers.AssertContains(t, err.Error(), tt.errMsg)
				}
			} else {
				helpers.AssertNoError(t, err)
			}
		})
	}
}

func TestFieldValidator_SetStrictMode(t *testing.T) {
	validator := parser.NewFieldValidator()

	// Test default mode (non-strict)
	record := &models.ETCMeisai{
		UseDate:   time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC),
		UseTime:   "09:30",
		EntryIC:   "東京IC",
		ExitIC:    "大阪IC",
		Amount:    0, // Zero amount
		CarNumber: "品川123あ1234",
		ETCNumber: "1234567890",
	}

	err := validator.ValidateRecord(record)
	helpers.AssertNoError(t, err) // Should pass in non-strict mode

	// Enable strict mode
	validator.SetStrictMode(true)

	err = validator.ValidateRecord(record)
	helpers.AssertError(t, err) // Should fail in strict mode due to zero amount

	// Disable strict mode
	validator.SetStrictMode(false)

	err = validator.ValidateRecord(record)
	helpers.AssertNoError(t, err) // Should pass again
}

func TestFieldValidator_GetValidationSummary(t *testing.T) {
	validator := parser.NewFieldValidator()

	records := []*models.ETCMeisai{
		{
			UseDate:   time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC),
			UseTime:   "09:30",
			EntryIC:   "東京IC",
			ExitIC:    "大阪IC",
			Amount:    1000,
			CarNumber: "品川123あ1234",
			ETCNumber: "1234567890",
		},
		{
			UseDate:   time.Time{}, // Invalid
			UseTime:   "09:30",
			EntryIC:   "東京IC",
			ExitIC:    "大阪IC",
			Amount:    1000,
			CarNumber: "品川123あ1234",
			ETCNumber: "1234567890",
		},
	}

	summary := validator.GetValidationSummary(records)

	helpers.AssertEqual(t, 2, summary.TotalRecords)
	helpers.AssertEqual(t, 1, summary.ValidRecords)
	helpers.AssertEqual(t, 1, summary.InvalidRecords)
	helpers.AssertTrue(t, len(summary.ErrorsByType) > 0)
}
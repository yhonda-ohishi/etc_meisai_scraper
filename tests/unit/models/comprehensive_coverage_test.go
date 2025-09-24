package models_test

import (
	"testing"
	"time"

	"github.com/yhonda-ohishi/etc_meisai/src/models"
	"github.com/yhonda-ohishi/etc_meisai/tests/helpers"
)

// TestETCMeisaiRecord_EdgeCases tests edge cases for 100% coverage
func TestETCMeisaiRecord_EdgeCases(t *testing.T) {
	tests := []struct {
		name    string
		record  *models.ETCMeisaiRecord
		wantErr bool
		errMsg  string
	}{
		{
			name: "very long entrance IC",
			record: &models.ETCMeisaiRecord{
				Date:          time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
				Time:          "09:30:00",
				EntranceIC:    "a very long IC name that exceeds the maximum allowed length of 100 characters to test the validation logic properly",
				ExitIC:        "大阪IC",
				TollAmount:    1000,
				CarNumber:     "ABC123",
				ETCCardNumber: "1234567890123456",
			},
			wantErr: true,
			errMsg:  "entrance IC name too long",
		},
		{
			name: "very long exit IC",
			record: &models.ETCMeisaiRecord{
				Date:          time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
				Time:          "09:30:00",
				EntranceIC:    "東京IC",
				ExitIC:        "a very long IC name that exceeds the maximum allowed length of 100 characters to test the validation logic properly",
				TollAmount:    1000,
				CarNumber:     "ABC123",
				ETCCardNumber: "1234567890123456",
			},
			wantErr: true,
			errMsg:  "exit IC name too long",
		},
		{
			name: "maximum valid toll amount",
			record: &models.ETCMeisaiRecord{
				Date:          time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
				Time:          "09:30:00",
				EntranceIC:    "東京IC",
				ExitIC:        "大阪IC",
				TollAmount:    999999, // Maximum allowed amount
				CarNumber:     "ABC123",
				ETCCardNumber: "1234567890123456",
			},
			wantErr: false,
		},
		{
			name: "toll amount exceeding maximum",
			record: &models.ETCMeisaiRecord{
				Date:          time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
				Time:          "09:30:00",
				EntranceIC:    "東京IC",
				ExitIC:        "大阪IC",
				TollAmount:    1000000, // Exceeds maximum
				CarNumber:     "ABC123",
				ETCCardNumber: "1234567890123456",
			},
			wantErr: true,
			errMsg:  "toll amount too large",
		},
		{
			name: "short ETC card number",
			record: &models.ETCMeisaiRecord{
				Date:          time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
				Time:          "09:30:00",
				EntranceIC:    "東京IC",
				ExitIC:        "大阪IC",
				TollAmount:    1000,
				CarNumber:     "ABC123",
				ETCCardNumber: "12345", // Too short
			},
			wantErr: true,
			errMsg:  "ETC card number must be 16-19 digits",
		},
		{
			name: "long ETC card number",
			record: &models.ETCMeisaiRecord{
				Date:          time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
				Time:          "09:30:00",
				EntranceIC:    "東京IC",
				ExitIC:        "大阪IC",
				TollAmount:    1000,
				CarNumber:     "ABC123",
				ETCCardNumber: "12345678901234567890", // Too long (20 digits)
			},
			wantErr: true,
			errMsg:  "ETC card number must be 16-19 digits",
		},
		{
			name: "ETC card number with non-digits",
			record: &models.ETCMeisaiRecord{
				Date:          time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
				Time:          "09:30:00",
				EntranceIC:    "東京IC",
				ExitIC:        "大阪IC",
				TollAmount:    1000,
				CarNumber:     "ABC123",
				ETCCardNumber: "1234567890abcdef", // Contains letters
			},
			wantErr: true,
			errMsg:  "ETC card number must contain only digits",
		},
		{
			name: "valid ETC card number with spaces",
			record: &models.ETCMeisaiRecord{
				Date:          time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
				Time:          "09:30:00",
				EntranceIC:    "東京IC",
				ExitIC:        "大阪IC",
				TollAmount:    1000,
				CarNumber:     "ABC123",
				ETCCardNumber: "1234 5678 9012 3456", // Spaces should be allowed
			},
			wantErr: false,
		},
		{
			name: "valid ETC card number with hyphens",
			record: &models.ETCMeisaiRecord{
				Date:          time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
				Time:          "09:30:00",
				EntranceIC:    "东京IC",
				ExitIC:        "大阪IC",
				TollAmount:    1000,
				CarNumber:     "ABC123",
				ETCCardNumber: "1234-5678-9012-3456", // Hyphens should be allowed
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.record.Validate()

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

// TestETCMeisaiRecord_ETCNumValidation tests ETC number validation scenarios
func TestETCMeisaiRecord_ETCNumValidation(t *testing.T) {
	tests := []struct {
		name    string
		etcNum  *string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "nil ETC num",
			etcNum:  nil,
			wantErr: false,
		},
		{
			name:    "empty ETC num",
			etcNum:  stringPtr(""),
			wantErr: false, // Empty is allowed as it's optional
		},
		{
			name:    "too short ETC num",
			etcNum:  stringPtr("1234"),
			wantErr: true,
			errMsg:  "ETC number must be 5-50 characters",
		},
		{
			name:    "too long ETC num",
			etcNum:  stringPtr("this_is_a_very_long_etc_number_that_exceeds_the_maximum_allowed_length_of_fifty_characters"),
			wantErr: true,
			errMsg:  "ETC number must be 5-50 characters",
		},
		{
			name:    "ETC num with invalid characters",
			etcNum:  stringPtr("ABC123@#$"),
			wantErr: true,
			errMsg:  "ETC number contains invalid characters",
		},
		{
			name:    "valid ETC num with alphanumeric",
			etcNum:  stringPtr("ABC123XYZ"),
			wantErr: false,
		},
		{
			name:    "valid ETC num with hyphens and underscores",
			etcNum:  stringPtr("ABC-123_XYZ"),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			record := &models.ETCMeisaiRecord{
				Date:          time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
				Time:          "09:30:00",
				EntranceIC:    "東京IC",
				ExitIC:        "大阪IC",
				TollAmount:    1000,
				CarNumber:     "ABC123",
				ETCCardNumber: "1234567890123456",
				ETCNum:        tt.etcNum,
			}

			err := record.Validate()

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

// TestETCMeisaiRecord_CarNumberValidation tests car number validation scenarios
func TestETCMeisaiRecord_CarNumberValidation(t *testing.T) {
	tests := []struct {
		name      string
		carNumber string
		wantErr   bool
		errMsg    string
	}{
		{
			name:      "empty car number",
			carNumber: "",
			wantErr:   true,
			errMsg:    "car number cannot be empty",
		},
		{
			name:      "whitespace only car number",
			carNumber: "   ",
			wantErr:   true,
			errMsg:    "car number cannot be empty",
		},
		{
			name:      "valid pattern 123-45",
			carNumber: "123-45",
			wantErr:   false,
		},
		{
			name:      "valid pattern 123 45",
			carNumber: "123 45",
			wantErr:   false,
		},
		{
			name:      "valid pattern あ123",
			carNumber: "あ123",
			wantErr:   false,
		},
		{
			name:      "valid pattern ア123",
			carNumber: "ア123",
			wantErr:   false,
		},
		{
			name:      "valid pattern 12-34",
			carNumber: "12-34",
			wantErr:   false,
		},
		{
			name:      "valid pattern 1234",
			carNumber: "1234",
			wantErr:   false,
		},
		{
			name:      "valid pattern ABC123",
			carNumber: "ABC123",
			wantErr:   false,
		},
		{
			name:      "invalid car number pattern",
			carNumber: "@@##$$",
			wantErr:   true,
			errMsg:    "invalid car number format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			record := &models.ETCMeisaiRecord{
				Date:          time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
				Time:          "09:30:00",
				EntranceIC:    "東京IC",
				ExitIC:        "大阪IC",
				TollAmount:    1000,
				CarNumber:     tt.carNumber,
				ETCCardNumber: "1234567890123456",
			}

			err := record.Validate()

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

// TestETCMeisaiRecord_UtilityMethods tests utility methods for coverage
func TestETCMeisaiRecord_UtilityMethods(t *testing.T) {
	record := &models.ETCMeisaiRecord{
		ID:            123,
		Date:          time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		Hash:          "test-hash",
		ETCCardNumber: "1234567890123456",
	}

	// Test GetDateString
	dateStr := record.GetDateString()
	helpers.AssertEqual(t, "2025-01-01", dateStr)

	// Test IsValidForMapping
	helpers.AssertTrue(t, record.IsValidForMapping())

	// Test with invalid record
	invalidRecord := &models.ETCMeisaiRecord{}
	helpers.AssertFalse(t, invalidRecord.IsValidForMapping())

	// Test GetMaskedETCCardNumber
	masked := record.GetMaskedETCCardNumber()
	helpers.AssertContains(t, masked, "****-****-****-3456")

	// Test with short card number
	record.ETCCardNumber = "123"
	masked = record.GetMaskedETCCardNumber()
	helpers.AssertEqual(t, "****", masked)
}

// TestETCMapping_EdgeCases tests edge cases for ETCMapping
func TestETCMapping_EdgeCases(t *testing.T) {
	tests := []struct {
		name    string
		mapping *models.ETCMapping
		wantErr bool
		errMsg  string
	}{
		{
			name: "zero ETC record ID",
			mapping: &models.ETCMapping{
				ETCRecordID:      0,
				MappingType:      "dtako",
				MappedEntityID:   1,
				MappedEntityType: "dtako_record",
				Confidence:       1.0,
				Status:           "active",
			},
			wantErr: true,
			errMsg:  "ETC record ID must be positive",
		},
		{
			name: "negative ETC record ID",
			mapping: &models.ETCMapping{
				ETCRecordID:      -1,
				MappingType:      "dtako",
				MappedEntityID:   1,
				MappedEntityType: "dtako_record",
				Confidence:       1.0,
				Status:           "active",
			},
			wantErr: true,
			errMsg:  "ETC record ID must be positive",
		},
		{
			name: "zero mapped entity ID",
			mapping: &models.ETCMapping{
				ETCRecordID:      1,
				MappingType:      "dtako",
				MappedEntityID:   0,
				MappedEntityType: "dtako_record",
				Confidence:       1.0,
				Status:           "active",
			},
			wantErr: true,
			errMsg:  "mapped entity ID must be positive",
		},
		{
			name: "negative mapped entity ID",
			mapping: &models.ETCMapping{
				ETCRecordID:      1,
				MappingType:      "dtako",
				MappedEntityID:   -1,
				MappedEntityType: "dtako_record",
				Confidence:       1.0,
				Status:           "active",
			},
			wantErr: true,
			errMsg:  "mapped entity ID must be positive",
		},
		{
			name: "confidence below 0",
			mapping: &models.ETCMapping{
				ETCRecordID:      1,
				MappingType:      "dtako",
				MappedEntityID:   1,
				MappedEntityType: "dtako_record",
				Confidence:       -0.1,
				Status:           "active",
			},
			wantErr: true,
			errMsg:  "confidence must be between 0.0 and 1.0",
		},
		{
			name: "confidence above 1",
			mapping: &models.ETCMapping{
				ETCRecordID:      1,
				MappingType:      "dtako",
				MappedEntityID:   1,
				MappedEntityType: "dtako_record",
				Confidence:       1.1,
				Status:           "active",
			},
			wantErr: true,
			errMsg:  "confidence must be between 0.0 and 1.0",
		},
		{
			name: "empty mapping type",
			mapping: &models.ETCMapping{
				ETCRecordID:      1,
				MappingType:      "",
				MappedEntityID:   1,
				MappedEntityType: "dtako_record",
				Confidence:       1.0,
				Status:           "active",
			},
			wantErr: true,
			errMsg:  "mapping type cannot be empty",
		},
		{
			name: "invalid mapping type",
			mapping: &models.ETCMapping{
				ETCRecordID:      1,
				MappingType:      "invalid_type",
				MappedEntityID:   1,
				MappedEntityType: "dtako_record",
				Confidence:       1.0,
				Status:           "active",
			},
			wantErr: true,
			errMsg:  "invalid mapping type",
		},
		{
			name: "empty mapped entity type",
			mapping: &models.ETCMapping{
				ETCRecordID:      1,
				MappingType:      "dtako",
				MappedEntityID:   1,
				MappedEntityType: "",
				Confidence:       1.0,
				Status:           "active",
			},
			wantErr: true,
			errMsg:  "mapped entity type cannot be empty",
		},
		{
			name: "invalid mapped entity type",
			mapping: &models.ETCMapping{
				ETCRecordID:      1,
				MappingType:      "dtako",
				MappedEntityID:   1,
				MappedEntityType: "invalid_type",
				Confidence:       1.0,
				Status:           "active",
			},
			wantErr: true,
			errMsg:  "invalid mapped entity type",
		},
		{
			name: "invalid status",
			mapping: &models.ETCMapping{
				ETCRecordID:      1,
				MappingType:      "dtako",
				MappedEntityID:   1,
				MappedEntityType: "dtako_record",
				Confidence:       1.0,
				Status:           "invalid_status",
			},
			wantErr: true,
			errMsg:  "invalid status",
		},
		{
			name: "too long created_by",
			mapping: &models.ETCMapping{
				ETCRecordID:      1,
				MappingType:      "dtako",
				MappedEntityID:   1,
				MappedEntityType: "dtako_record",
				Confidence:       1.0,
				Status:           "active",
				CreatedBy:        "this_is_a_very_long_created_by_field_that_exceeds_the_maximum_allowed_length_of_one_hundred_characters",
			},
			wantErr: true,
			errMsg:  "created_by field too long",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.mapping.Validate()

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



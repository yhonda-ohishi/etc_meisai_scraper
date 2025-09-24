package adapters_test

import (
	"testing"
	"time"

	"github.com/yhonda-ohishi/etc_meisai/src/adapters"
	"github.com/yhonda-ohishi/etc_meisai/src/models"
	"github.com/yhonda-ohishi/etc_meisai/tests/helpers"
)

func TestETCCompatAdapter_NewETCCompatAdapter(t *testing.T) {
	adapter := adapters.NewETCMeisaiCompatAdapter()
	helpers.AssertNotNil(t, adapter)
}

func TestETCCompatAdapter_ConvertFromLegacy(t *testing.T) {
	adapter := adapters.NewETCMeisaiCompatAdapter()

	tests := []struct {
		name       string
		legacyData map[string]interface{}
		wantErr    bool
		errMsg     string
		validate   func(t *testing.T, record *models.ETCMeisaiRecord)
	}{
		{
			name: "valid legacy data",
			legacyData: map[string]interface{}{
				"use_date":    "2025-01-15",
				"use_time":    "09:30",
				"entry_ic":    "東京IC",
				"exit_ic":     "大阪IC",
				"amount":      1000,
				"car_number":  "品川123あ1234",
				"etc_number":  "1234567890",
			},
			wantErr: false,
			validate: func(t *testing.T, record *models.ETCMeisaiRecord) {
				helpers.AssertEqual(t, time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC), record.Date)
				helpers.AssertEqual(t, "09:30", record.Time)
				helpers.AssertEqual(t, "東京IC", record.EntranceIC)
				helpers.AssertEqual(t, "大阪IC", record.ExitIC)
				helpers.AssertEqual(t, 1000, record.TollAmount)
				helpers.AssertEqual(t, "品川123あ1234", record.CarNumber)
				helpers.AssertEqual(t, "1234567890", record.ETCCardNumber)
			},
		},
		{
			name:       "nil legacy data",
			legacyData: nil,
			wantErr:    true,
			errMsg:     "legacy data cannot be nil",
		},
		{
			name:       "empty legacy data",
			legacyData: map[string]interface{}{},
			wantErr:    true,
			errMsg:     "legacy data is empty",
		},
		{
			name: "missing required field",
			legacyData: map[string]interface{}{
				"use_time":   "09:30",
				"entry_ic":   "東京IC",
				"exit_ic":    "大阪IC",
				"amount":     1000,
				"car_number": "品川123あ1234",
				"etc_number": "1234567890",
				// Missing use_date
			},
			wantErr: true,
			errMsg:  "missing required field: use_date",
		},
		{
			name: "invalid date format",
			legacyData: map[string]interface{}{
				"use_date":   "invalid-date",
				"use_time":   "09:30",
				"entry_ic":   "東京IC",
				"exit_ic":    "大阪IC",
				"amount":     1000,
				"car_number": "品川123あ1234",
				"etc_number": "1234567890",
			},
			wantErr: true,
			errMsg:  "invalid date format",
		},
		{
			name: "invalid amount type",
			legacyData: map[string]interface{}{
				"use_date":   "2025-01-15",
				"use_time":   "09:30",
				"entry_ic":   "東京IC",
				"exit_ic":    "大阪IC",
				"amount":     "invalid",
				"car_number": "品川123あ1234",
				"etc_number": "1234567890",
			},
			wantErr: true,
			errMsg:  "invalid amount format",
		},
		{
			name: "alternative field names",
			legacyData: map[string]interface{}{
				"利用年月日": "2025-01-15",
				"利用時刻":  "09:30",
				"入口IC":   "東京IC",
				"出口IC":   "大阪IC",
				"料金":     1000,
				"車両番号":  "品川123あ1234",
				"ETCカード番号": "1234567890",
			},
			wantErr: false,
			validate: func(t *testing.T, record *models.ETCMeisaiRecord) {
				helpers.AssertEqual(t, "東京IC", record.EntranceIC)
				helpers.AssertEqual(t, "大阪IC", record.ExitIC)
				helpers.AssertEqual(t, 1000, record.TollAmount)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			record, err := adapter.ConvertFromLegacy(tt.legacyData)

			if tt.wantErr {
				helpers.AssertError(t, err)
				helpers.AssertNil(t, record)
				if tt.errMsg != "" {
					helpers.AssertContains(t, err.Error(), tt.errMsg)
				}
			} else {
				helpers.AssertNoError(t, err)
				helpers.AssertNotNil(t, record)
				if tt.validate != nil {
					tt.validate(t, record)
				}
			}
		})
	}
}

func TestETCCompatAdapter_ConvertToLegacy(t *testing.T) {
	adapter := adapters.NewETCMeisaiCompatAdapter()

	record := &models.ETCMeisaiRecord{
		ID:            1,
		Date:          time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC),
		Time:          "09:30",
		EntranceIC:    "東京IC",
		ExitIC:        "大阪IC",
		TollAmount:    1000,
		CarNumber:     "品川123あ1234",
		ETCCardNumber: "1234567890",
		Hash:          "test_hash",
	}

	tests := []struct {
		name     string
		record   *models.ETCMeisaiRecord
		format   string
		wantErr  bool
		errMsg   string
		validate func(t *testing.T, data map[string]interface{})
	}{
		{
			name:   "convert to legacy format",
			record: record,
			format: "legacy",
			wantErr: false,
			validate: func(t *testing.T, data map[string]interface{}) {
				helpers.AssertEqual(t, "2025-01-15", data["use_date"])
				helpers.AssertEqual(t, "09:30", data["use_time"])
				helpers.AssertEqual(t, "東京IC", data["entry_ic"])
				helpers.AssertEqual(t, "大阪IC", data["exit_ic"])
				helpers.AssertEqual(t, 1000, data["amount"])
				helpers.AssertEqual(t, "品川123あ1234", data["car_number"])
				helpers.AssertEqual(t, "1234567890", data["etc_number"])
			},
		},
		{
			name:   "convert to Japanese format",
			record: record,
			format: "japanese",
			wantErr: false,
			validate: func(t *testing.T, data map[string]interface{}) {
				helpers.AssertEqual(t, "2025-01-15", data["利用年月日"])
				helpers.AssertEqual(t, "09:30", data["利用時刻"])
				helpers.AssertEqual(t, "東京IC", data["入口IC"])
				helpers.AssertEqual(t, "大阪IC", data["出口IC"])
				helpers.AssertEqual(t, 1000, data["料金"])
			},
		},
		{
			name:    "nil record",
			record:  nil,
			format:  "legacy",
			wantErr: true,
			errMsg:  "record cannot be nil",
		},
		{
			name:    "invalid format",
			record:  record,
			format:  "invalid",
			wantErr: true,
			errMsg:  "unsupported format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := adapter.ConvertToLegacy(tt.record, tt.format)

			if tt.wantErr {
				helpers.AssertError(t, err)
				helpers.AssertNil(t, data)
				if tt.errMsg != "" {
					helpers.AssertContains(t, err.Error(), tt.errMsg)
				}
			} else {
				helpers.AssertNoError(t, err)
				helpers.AssertNotNil(t, data)
				if tt.validate != nil {
					tt.validate(t, data)
				}
			}
		})
	}
}

func TestETCCompatAdapter_ConvertBatch(t *testing.T) {
	adapter := adapters.NewETCMeisaiCompatAdapter()

	legacyBatch := []map[string]interface{}{
		{
			"use_date":   "2025-01-15",
			"use_time":   "09:30",
			"entry_ic":   "東京IC",
			"exit_ic":    "大阪IC",
			"amount":     1000,
			"car_number": "品川123あ1234",
			"etc_number": "1234567890",
		},
		{
			"use_date":   "2025-01-16",
			"use_time":   "14:15",
			"entry_ic":   "名古屋IC",
			"exit_ic":    "京都IC",
			"amount":     800,
			"car_number": "横浜456い5678",
			"etc_number": "0987654321",
		},
	}

	tests := []struct {
		name        string
		legacyBatch []map[string]interface{}
		wantErr     bool
		errMsg      string
	}{
		{
			name:        "valid batch conversion",
			legacyBatch: legacyBatch,
			wantErr:     false,
		},
		{
			name:        "nil batch",
			legacyBatch: nil,
			wantErr:     true,
			errMsg:      "batch cannot be nil",
		},
		{
			name:        "empty batch",
			legacyBatch: []map[string]interface{}{},
			wantErr:     true,
			errMsg:      "batch cannot be empty",
		},
		{
			name: "batch with invalid record",
			legacyBatch: []map[string]interface{}{
				legacyBatch[0], // Valid record
				{
					"use_time":   "09:30",
					"entry_ic":   "東京IC",
					// Missing required fields
				},
			},
			wantErr: true,
			errMsg:  "failed to convert record at index 1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			records, err := adapter.ConvertBatch(tt.legacyBatch)

			if tt.wantErr {
				helpers.AssertError(t, err)
				helpers.AssertNil(t, records)
				if tt.errMsg != "" {
					helpers.AssertContains(t, err.Error(), tt.errMsg)
				}
			} else {
				helpers.AssertNoError(t, err)
				helpers.AssertNotNil(t, records)
				helpers.AssertLen(t, records, len(tt.legacyBatch))
			}
		})
	}
}

func TestETCCompatAdapter_ValidateCompatibility(t *testing.T) {
	adapter := adapters.NewETCMeisaiCompatAdapter()

	tests := []struct {
		name       string
		legacyData map[string]interface{}
		wantErr    bool
		errMsg     string
	}{
		{
			name: "compatible data",
			legacyData: map[string]interface{}{
				"use_date":   "2025-01-15",
				"use_time":   "09:30",
				"entry_ic":   "東京IC",
				"exit_ic":    "大阪IC",
				"amount":     1000,
				"car_number": "品川123あ1234",
				"etc_number": "1234567890",
			},
			wantErr: false,
		},
		{
			name: "incompatible version",
			legacyData: map[string]interface{}{
				"version":    "v1.0.0", // Old version
				"use_date":   "2025-01-15",
				"use_time":   "09:30",
				"entry_ic":   "東京IC",
				"exit_ic":    "大阪IC",
				"amount":     1000,
				"car_number": "品川123あ1234",
				"etc_number": "1234567890",
			},
			wantErr: true,
			errMsg:  "incompatible version",
		},
		{
			name: "missing critical fields",
			legacyData: map[string]interface{}{
				"use_time": "09:30",
				"amount":   1000,
			},
			wantErr: true,
			errMsg:  "missing critical fields",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := adapter.ValidateCompatibility(tt.legacyData)

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

func TestETCCompatAdapter_GetFieldMapping(t *testing.T) {
	adapter := adapters.NewETCMeisaiCompatAdapter()

	tests := []struct {
		name     string
		format   string
		expected map[string]string
		wantErr  bool
		errMsg   string
	}{
		{
			name:   "legacy field mapping",
			format: "legacy",
			expected: map[string]string{
				"use_date":   "Date",
				"use_time":   "Time",
				"entry_ic":   "EntranceIC",
				"exit_ic":    "ExitIC",
				"amount":     "TollAmount",
				"car_number": "CarNumber",
				"etc_number": "ETCCardNumber",
			},
			wantErr: false,
		},
		{
			name:   "japanese field mapping",
			format: "japanese",
			expected: map[string]string{
				"利用年月日":   "Date",
				"利用時刻":    "Time",
				"入口IC":     "EntranceIC",
				"出口IC":     "ExitIC",
				"料金":      "TollAmount",
				"車両番号":    "CarNumber",
				"ETCカード番号": "ETCCardNumber",
			},
			wantErr: false,
		},
		{
			name:    "invalid format",
			format:  "invalid",
			wantErr: true,
			errMsg:  "unsupported format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mapping, err := adapter.GetFieldMapping(tt.format)

			if tt.wantErr {
				helpers.AssertError(t, err)
				helpers.AssertNil(t, mapping)
				if tt.errMsg != "" {
					helpers.AssertContains(t, err.Error(), tt.errMsg)
				}
			} else {
				helpers.AssertNoError(t, err)
				helpers.AssertNotNil(t, mapping)
				for key, value := range tt.expected {
					helpers.AssertEqual(t, value, mapping[key])
				}
			}
		})
	}
}

func TestETCCompatAdapter_NormalizeFieldNames(t *testing.T) {
	adapter := adapters.NewETCMeisaiCompatAdapter()

	tests := []struct {
		name     string
		input    map[string]interface{}
		expected map[string]interface{}
	}{
		{
			name: "legacy field names",
			input: map[string]interface{}{
				"use_date":   "2025-01-15",
				"use_time":   "09:30",
				"entry_ic":   "東京IC",
				"exit_ic":    "大阪IC",
				"amount":     1000,
				"car_number": "品川123あ1234",
				"etc_number": "1234567890",
			},
			expected: map[string]interface{}{
				"use_date":   "2025-01-15",
				"use_time":   "09:30",
				"entry_ic":   "東京IC",
				"exit_ic":    "大阪IC",
				"amount":     1000,
				"car_number": "品川123あ1234",
				"etc_number": "1234567890",
			},
		},
		{
			name: "japanese field names",
			input: map[string]interface{}{
				"利用年月日":   "2025-01-15",
				"利用時刻":    "09:30",
				"入口IC":     "東京IC",
				"出口IC":     "大阪IC",
				"料金":      1000,
				"車両番号":    "品川123あ1234",
				"ETCカード番号": "1234567890",
			},
			expected: map[string]interface{}{
				"use_date":   "2025-01-15",
				"use_time":   "09:30",
				"entry_ic":   "東京IC",
				"exit_ic":    "大阪IC",
				"amount":     1000,
				"car_number": "品川123あ1234",
				"etc_number": "1234567890",
			},
		},
		{
			name: "mixed field names",
			input: map[string]interface{}{
				"use_date": "2025-01-15",
				"利用時刻":    "09:30",
				"entry_ic": "東京IC",
				"出口IC":     "大阪IC",
				"amount":   1000,
			},
			expected: map[string]interface{}{
				"use_date": "2025-01-15",
				"use_time": "09:30",
				"entry_ic": "東京IC",
				"exit_ic":  "大阪IC",
				"amount":   1000,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := adapter.NormalizeFieldNames(tt.input)

			for key, expectedValue := range tt.expected {
				actualValue, exists := result[key]
				helpers.AssertTrue(t, exists)
				helpers.AssertEqual(t, expectedValue, actualValue)
			}
		})
	}
}

func TestETCCompatAdapter_Performance(t *testing.T) {
	adapter := adapters.NewETCMeisaiCompatAdapter()

	// Create a large batch for performance testing
	largeBatch := make([]map[string]interface{}, 1000)
	for i := 0; i < 1000; i++ {
		largeBatch[i] = map[string]interface{}{
			"use_date":   "2025-01-15",
			"use_time":   "09:30",
			"entry_ic":   "東京IC",
			"exit_ic":    "大阪IC",
			"amount":     1000,
			"car_number": "品川123あ1234",
			"etc_number": "1234567890",
		}
	}

	start := time.Now()
	records, err := adapter.ConvertBatch(largeBatch)
	duration := time.Since(start)

	helpers.AssertNoError(t, err)
	helpers.AssertLen(t, records, 1000)

	// Should process large batches efficiently (under 1 second for 1000 records)
	helpers.AssertTrue(t, duration < time.Second)
}
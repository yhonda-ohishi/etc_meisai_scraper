package models_test

import (
	"testing"
	"time"

	"github.com/yhonda-ohishi/etc_meisai/src/models"
	"github.com/yhonda-ohishi/etc_meisai/tests/helpers"
)

func TestETCMeisai_BeforeCreate(t *testing.T) {
	tests := []struct {
		name     string
		record   *models.ETCMeisai
		validate func(t *testing.T, record *models.ETCMeisai)
	}{
		{
			name: "sets timestamps when zero",
			record: &models.ETCMeisai{
				UseDate:   time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
				UseTime:   "09:00",
				EntryIC:   "東京IC",
				ExitIC:    "大阪IC",
				Amount:    int32(1000),
				CarNumber: "品川123あ1234",
				ETCNumber: "1234567890",
			},
			validate: func(t *testing.T, record *models.ETCMeisai) {
				helpers.AssertNotNil(t, record.CreatedAt)
				helpers.AssertNotNil(t, record.UpdatedAt)
				helpers.AssertNotEmpty(t, record.Hash)
			},
		},
		{
			name: "preserves existing timestamps",
			record: &models.ETCMeisai{
				UseDate:   time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
				UseTime:   "09:00",
				EntryIC:   "東京IC",
				ExitIC:    "大阪IC",
				Amount:    int32(1000),
				CarNumber: "品川123あ1234",
				ETCNumber: "1234567890",
				CreatedAt: time.Date(2024, 12, 1, 0, 0, 0, 0, time.UTC),
				UpdatedAt: time.Date(2024, 12, 1, 0, 0, 0, 0, time.UTC),
			},
			validate: func(t *testing.T, record *models.ETCMeisai) {
				expectedTime := time.Date(2024, 12, 1, 0, 0, 0, 0, time.UTC)
				helpers.AssertEqual(t, expectedTime, record.CreatedAt)
				helpers.AssertEqual(t, expectedTime, record.UpdatedAt)
			},
		},
		{
			name: "generates hash when empty",
			record: &models.ETCMeisai{
				UseDate:   time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
				UseTime:   "09:00",
				EntryIC:   "東京IC",
				ExitIC:    "大阪IC",
				Amount:    int32(1000),
				CarNumber: "品川123あ1234",
				ETCNumber: "1234567890",
			},
			validate: func(t *testing.T, record *models.ETCMeisai) {
				helpers.AssertNotEmpty(t, record.Hash)
				helpers.AssertLen(t, record.Hash, 64) // SHA256 hash length
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.record.BeforeCreate()
			helpers.AssertNoError(t, err)
			tt.validate(t, tt.record)
		})
	}
}

func TestETCMeisai_BeforeUpdate(t *testing.T) {
	record := &models.ETCMeisai{
		UseDate:   time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		UseTime:   "09:00",
		EntryIC:   "東京IC",
		ExitIC:    "大阪IC",
		Amount:    1000,
		CarNumber: "品川123あ1234",
		ETCNumber: "1234567890",
		CreatedAt: time.Date(2024, 12, 1, 0, 0, 0, 0, time.UTC),
		UpdatedAt: time.Date(2024, 12, 1, 0, 0, 0, 0, time.UTC),
		Hash:      "old_hash",
	}

	oldUpdatedAt := record.UpdatedAt
	oldHash := record.Hash

	time.Sleep(time.Millisecond) // Ensure time difference

	err := record.BeforeUpdate()
	helpers.AssertNoError(t, err)

	// Should update timestamp
	helpers.AssertTrue(t, record.UpdatedAt.After(oldUpdatedAt))

	// Should regenerate hash
	helpers.AssertNotEqual(t, oldHash, record.Hash)
	helpers.AssertNotEmpty(t, record.Hash)
}

func TestETCMeisai_GenerateHash(t *testing.T) {
	tests := []struct {
		name     string
		record1  *models.ETCMeisai
		record2  *models.ETCMeisai
		expected string // "same" or "different"
	}{
		{
			name: "identical records produce same hash",
			record1: &models.ETCMeisai{
				UseDate:   time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
				UseTime:   "09:00",
				EntryIC:   "東京IC",
				ExitIC:    "大阪IC",
				Amount:    int32(1000),
				CarNumber: "品川123あ1234",
				ETCNumber: "1234567890",
			},
			record2: &models.ETCMeisai{
				UseDate:   time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
				UseTime:   "09:00",
				EntryIC:   "東京IC",
				ExitIC:    "大阪IC",
				Amount:    int32(1000),
				CarNumber: "品川123あ1234",
				ETCNumber: "1234567890",
			},
			expected: "same",
		},
		{
			name: "different amounts produce different hash",
			record1: &models.ETCMeisai{
				UseDate:   time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
				UseTime:   "09:00",
				EntryIC:   "東京IC",
				ExitIC:    "大阪IC",
				Amount:    int32(1000),
				CarNumber: "品川123あ1234",
				ETCNumber: "1234567890",
			},
			record2: &models.ETCMeisai{
				UseDate:   time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
				UseTime:   "09:00",
				EntryIC:   "東京IC",
				ExitIC:    "大阪IC",
				Amount:    int32(2000), // Different amount
				CarNumber: "品川123あ1234",
				ETCNumber: "1234567890",
			},
			expected: "different",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash1 := tt.record1.GenerateHash()
			hash2 := tt.record2.GenerateHash()

			// Both should be valid SHA256 hashes
			helpers.AssertLen(t, hash1, 64)
			helpers.AssertLen(t, hash2, 64)

			if tt.expected == "same" {
				helpers.AssertEqual(t, hash1, hash2)
			} else {
				helpers.AssertNotEqual(t, hash1, hash2)
			}
		})
	}
}

func TestETCMeisai_Validate(t *testing.T) {
	tests := []struct {
		name    string
		record  *models.ETCMeisai
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid record",
			record: &models.ETCMeisai{
				UseDate:   time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
				UseTime:   "09:00",
				EntryIC:   "東京IC",
				ExitIC:    "大阪IC",
				Amount:    int32(1000),
				CarNumber: "品川123あ1234",
				ETCNumber: "1234567890",
			},
			wantErr: false,
		},
		{
			name: "missing UseDate",
			record: &models.ETCMeisai{
				UseTime:   "09:00",
				EntryIC:   "東京IC",
				ExitIC:    "大阪IC",
				Amount:    int32(1000),
				CarNumber: "品川123あ1234",
				ETCNumber: "1234567890",
			},
			wantErr: true,
			errMsg:  "UseDate is required",
		},
		{
			name: "missing UseTime",
			record: &models.ETCMeisai{
				UseDate:   time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
				EntryIC:   "東京IC",
				ExitIC:    "大阪IC",
				Amount:    int32(1000),
				CarNumber: "品川123あ1234",
				ETCNumber: "1234567890",
			},
			wantErr: true,
			errMsg:  "UseTime is required",
		},
		{
			name: "missing EntryIC",
			record: &models.ETCMeisai{
				UseDate:   time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
				UseTime:   "09:00",
				ExitIC:    "大阪IC",
				Amount:    int32(1000),
				CarNumber: "品川123あ1234",
				ETCNumber: "1234567890",
			},
			wantErr: true,
			errMsg:  "EntryIC is required",
		},
		{
			name: "missing ExitIC",
			record: &models.ETCMeisai{
				UseDate:   time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
				UseTime:   "09:00",
				EntryIC:   "東京IC",
				Amount:    int32(1000),
				CarNumber: "品川123あ1234",
				ETCNumber: "1234567890",
			},
			wantErr: true,
			errMsg:  "ExitIC is required",
		},
		{
			name: "zero amount",
			record: &models.ETCMeisai{
				UseDate:   time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
				UseTime:   "09:00",
				EntryIC:   "東京IC",
				ExitIC:    "大阪IC",
				Amount:    int32(0),
				CarNumber: "品川123あ1234",
				ETCNumber: "1234567890",
			},
			wantErr: true,
			errMsg:  "Amount must be positive",
		},
		{
			name: "missing ETCNumber",
			record: &models.ETCMeisai{
				UseDate:   time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
				UseTime:   "09:00",
				EntryIC:   "東京IC",
				ExitIC:    "大阪IC",
				Amount:    int32(1000),
				CarNumber: "品川123あ1234",
			},
			wantErr: true,
			errMsg:  "ETCNumber is required",
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

func TestETCMeisai_GetTableName(t *testing.T) {
	record := &models.ETCMeisai{}
	tableName := record.GetTableName()
	helpers.AssertEqual(t, "etc_meisai", tableName)
}

func TestETCMeisai_String(t *testing.T) {
	record := &models.ETCMeisai{
		ID:        1,
		UseDate:   time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		UseTime:   "09:00",
		EntryIC:   "東京IC",
		ExitIC:    "大阪IC",
		Amount:    1000,
		CarNumber: "品川123あ1234",
		ETCNumber: "1234567890",
	}

	str := record.String()

	// Should contain key information
	helpers.AssertContains(t, str, "1234567890") // ETC Number
	helpers.AssertContains(t, str, "東京IC")      // Entry IC
	helpers.AssertContains(t, str, "大阪IC")      // Exit IC
	helpers.AssertContains(t, str, "1000")       // Amount
}
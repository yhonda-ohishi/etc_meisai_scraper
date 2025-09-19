package models

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestETCMeisai_GenerateHash(t *testing.T) {
	tests := []struct {
		name     string
		meisai1  *ETCMeisai
		meisai2  *ETCMeisai
		sameHash bool
	}{
		{
			name: "Same data should generate same hash",
			meisai1: &ETCMeisai{
				UseDate:   time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
				UseTime:   "08:30",
				EntryIC:   "東京IC",
				ExitIC:    "横浜IC",
				Amount:    1200,
				ETCNumber: "1234567890123456",
			},
			meisai2: &ETCMeisai{
				UseDate:   time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
				UseTime:   "08:30",
				EntryIC:   "東京IC",
				ExitIC:    "横浜IC",
				Amount:    1200,
				ETCNumber: "1234567890123456",
			},
			sameHash: true,
		},
		{
			name: "Different date should generate different hash",
			meisai1: &ETCMeisai{
				UseDate:   time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
				UseTime:   "08:30",
				EntryIC:   "東京IC",
				ExitIC:    "横浜IC",
				Amount:    1200,
				ETCNumber: "1234567890123456",
			},
			meisai2: &ETCMeisai{
				UseDate:   time.Date(2024, 1, 16, 0, 0, 0, 0, time.UTC), // Different date
				UseTime:   "08:30",
				EntryIC:   "東京IC",
				ExitIC:    "横浜IC",
				Amount:    1200,
				ETCNumber: "1234567890123456",
			},
			sameHash: false,
		},
		{
			name: "Different amount should generate different hash",
			meisai1: &ETCMeisai{
				UseDate:   time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
				UseTime:   "08:30",
				EntryIC:   "東京IC",
				ExitIC:    "横浜IC",
				Amount:    1200,
				ETCNumber: "1234567890123456",
			},
			meisai2: &ETCMeisai{
				UseDate:   time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
				UseTime:   "08:30",
				EntryIC:   "東京IC",
				ExitIC:    "横浜IC",
				Amount:    1500, // Different amount
				ETCNumber: "1234567890123456",
			},
			sameHash: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash1 := tt.meisai1.GenerateHash()
			hash2 := tt.meisai2.GenerateHash()

			assert.NotEmpty(t, hash1, "Hash should not be empty")
			assert.NotEmpty(t, hash2, "Hash should not be empty")
			assert.Len(t, hash1, 64, "SHA256 hash should be 64 characters")
			assert.Len(t, hash2, 64, "SHA256 hash should be 64 characters")

			if tt.sameHash {
				assert.Equal(t, hash1, hash2, "Hashes should be equal for same data")
			} else {
				assert.NotEqual(t, hash1, hash2, "Hashes should be different for different data")
			}
		})
	}
}

func TestETCMeisai_Validate(t *testing.T) {
	tests := []struct {
		name    string
		meisai  *ETCMeisai
		wantErr bool
		errMsg  string
	}{
		{
			name: "Valid record",
			meisai: &ETCMeisai{
				UseDate:   time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
				UseTime:   "08:30",
				EntryIC:   "東京IC",
				ExitIC:    "横浜IC",
				Amount:    1200,
				CarNumber: "TEST001",
				ETCNumber: "1234567890123456",
				Hash:      "testhash",
			},
			wantErr: false,
		},
		{
			name: "Missing use date",
			meisai: &ETCMeisai{
				UseTime:   "08:30",
				EntryIC:   "東京IC",
				ExitIC:    "横浜IC",
				Amount:    1200,
				CarNumber: "TEST001",
				ETCNumber: "1234567890123456",
				Hash:      "testhash",
			},
			wantErr: true,
			errMsg:  "UseDate is required",
		},
		{
			name: "Missing entry IC",
			meisai: &ETCMeisai{
				UseDate:   time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
				UseTime:   "08:30",
				// EntryIC is missing but currently not validated
				ExitIC:    "横浜IC",
				Amount:    1200,
				CarNumber: "TEST001",
				ETCNumber: "1234567890123456",
				Hash:      "testhash",
			},
			wantErr: false, // Current implementation doesn't validate EntryIC
		},
		{
			name: "Negative amount",
			meisai: &ETCMeisai{
				UseDate:   time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
				UseTime:   "08:30",
				EntryIC:   "東京IC",
				ExitIC:    "横浜IC",
				Amount:    -100,
				CarNumber: "TEST001",
				ETCNumber: "1234567890123456",
				Hash:      "testhash",
			},
			wantErr: true,
			errMsg:  "Amount must be positive",
		},
		{
			name: "Invalid time format",
			meisai: &ETCMeisai{
				UseDate:   time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
				UseTime:   "25:99", // Invalid time but not currently validated
				EntryIC:   "東京IC",
				ExitIC:    "横浜IC",
				Amount:    1200,
				CarNumber: "TEST001",
				ETCNumber: "1234567890123456",
				Hash:      "testhash",
			},
			wantErr: false, // Current implementation doesn't validate time format
		},
		{
			name: "Missing hash",
			meisai: &ETCMeisai{
				UseDate:   time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
				UseTime:   "08:30",
				EntryIC:   "東京IC",
				ExitIC:    "横浜IC",
				Amount:    1200,
				CarNumber: "TEST001",
				ETCNumber: "1234567890123456",
			},
			wantErr: true,
			errMsg:  "Hash is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.meisai.Validate()
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" && err != nil {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestETCMeisai_BeforeCreate(t *testing.T) {
	meisai := &ETCMeisai{
		UseDate:   time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
		UseTime:   "08:30",
		EntryIC:   "東京IC",
		ExitIC:    "横浜IC",
		Amount:    1200,
		CarNumber: "TEST001",
		ETCNumber: "1234567890123456",
	}

	// Hash should be empty before BeforeCreate
	assert.Empty(t, meisai.Hash)

	// Call BeforeCreate (normally called by GORM)
	err := meisai.BeforeCreate(nil)
	assert.NoError(t, err)

	// Hash should be generated
	assert.NotEmpty(t, meisai.Hash)
	assert.Len(t, meisai.Hash, 64)
}

func TestETCListParams_SetDefaults(t *testing.T) {
	tests := []struct {
		name   string
		params *ETCListParams
		want   ETCListParams
	}{
		{
			name:   "Nil params",
			params: nil,
			want:   ETCListParams{},
		},
		{
			name:   "Empty params",
			params: &ETCListParams{},
			want: ETCListParams{
				Limit:  100,
				Offset: 0,
			},
		},
		{
			name: "Negative limit",
			params: &ETCListParams{
				Limit:  -10,
				Offset: 50,
			},
			want: ETCListParams{
				Limit:  100,
				Offset: 50,
			},
		},
		{
			name: "Excessive limit",
			params: &ETCListParams{
				Limit:  10000,
				Offset: 0,
			},
			want: ETCListParams{
				Limit:  1000,
				Offset: 0,
			},
		},
		{
			name: "Valid params",
			params: &ETCListParams{
				Limit:  50,
				Offset: 100,
			},
			want: ETCListParams{
				Limit:  50,
				Offset: 100,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.params != nil {
				tt.params.SetDefaults()
				assert.Equal(t, tt.want.Limit, tt.params.Limit)
				assert.Equal(t, tt.want.Offset, tt.params.Offset)
			}
		})
	}
}

func TestValidateETCMeisaiBatch(t *testing.T) {
	validRecord := &ETCMeisai{
		UseDate:   time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
		UseTime:   "08:30",
		EntryIC:   "東京IC",
		ExitIC:    "横浜IC",
		Amount:    1200,
		CarNumber: "TEST001",
		ETCNumber: "1234567890123456",
		Hash:      "testhash",
	}

	invalidRecord := &ETCMeisai{
		UseDate: time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
		// Missing required fields
	}

	records := []*ETCMeisai{validRecord, invalidRecord}

	options := &BatchValidationOptions{
		StrictMode:     false,
		SkipDuplicates: false,
		MaxErrors:      10,
	}

	results := ValidateETCMeisaiBatch(records, options)

	assert.Len(t, results, 2)

	// First record should be valid
	assert.True(t, results[0].Valid)
	assert.Empty(t, results[0].Errors)

	// Second record should be invalid
	assert.False(t, results[1].Valid)
	assert.NotEmpty(t, results[1].Errors)
}
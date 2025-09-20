package models_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yhonda-ohishi/etc_meisai/src/models"
)

// Test ETCMeisai model - 100% coverage
func TestETCMeisai_Complete(t *testing.T) {
	t.Run("GenerateHash", func(t *testing.T) {
		m := &models.ETCMeisai{
			UseDate:   time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			UseTime:   "14:30:00",
			EntryIC:   "東京IC",
			ExitIC:    "横浜IC",
			Amount:    1500,
			ETCNumber: "1234567890",
		}
		hash := m.GenerateHash()
		assert.NotEmpty(t, hash)
		assert.Len(t, hash, 64) // SHA256 produces 64 hex characters

		// Same data should produce same hash
		hash2 := m.GenerateHash()
		assert.Equal(t, hash, hash2)

		// Different data should produce different hash
		m.Amount = 2000
		hash3 := m.GenerateHash()
		assert.NotEqual(t, hash, hash3)
	})

	t.Run("Validate", func(t *testing.T) {
		// Valid case
		m := &models.ETCMeisai{
			UseDate:   time.Now().Add(-24 * time.Hour),
			Amount:    1500,
			EntryIC:   "東京IC",
			ExitIC:    "横浜IC",
			ETCNumber: "1234567890",
			Hash:      "valid_hash_value",
		}
		err := m.Validate()
		assert.NoError(t, err)

		// Missing UseDate
		m.UseDate = time.Time{}
		err = m.Validate()
		assert.Error(t, err)

		// Negative amount
		m.UseDate = time.Now().Add(-24 * time.Hour)
		m.Amount = -100
		err = m.Validate()
		assert.Error(t, err)

		// Zero amount
		m.Amount = 0
		err = m.Validate()
		assert.Error(t, err)

		// Future date
		m.Amount = 1500
		m.UseDate = time.Now().Add(24 * time.Hour)
		err = m.Validate()
		assert.Error(t, err)

		// Long ETC number
		m.UseDate = time.Now().Add(-24 * time.Hour)
		m.ETCNumber = "123456789012345678901" // 21 chars
		err = m.Validate()
		assert.Error(t, err)

		// Empty hash
		m.ETCNumber = "1234567890"
		m.Hash = ""
		err = m.Validate()
		assert.Error(t, err)
	})

	t.Run("BeforeCreate", func(t *testing.T) {
		m := &models.ETCMeisai{
			UseDate:   time.Now().Add(-24 * time.Hour),
			Amount:    1500,
			EntryIC:   "東京IC",
			ExitIC:    "横浜IC",
			ETCNumber: "1234567890",
		}

		// Should generate hash if missing
		err := m.BeforeCreate()
		assert.NoError(t, err)
		assert.NotEmpty(t, m.Hash)

		// Should preserve existing hash
		existingHash := "existing_hash_value"
		m.Hash = existingHash
		err = m.BeforeCreate()
		assert.NoError(t, err)
		assert.Equal(t, existingHash, m.Hash)

		// Should fail validation
		m.Amount = -100
		err = m.BeforeCreate()
		assert.Error(t, err)
	})

	t.Run("BeforeUpdate", func(t *testing.T) {
		m := &models.ETCMeisai{
			UseDate:   time.Now().Add(-24 * time.Hour),
			Amount:    1500,
			EntryIC:   "東京IC",
			ExitIC:    "横浜IC",
			ETCNumber: "1234567890",
			Hash:      "old_hash",
		}

		// Should regenerate hash
		err := m.BeforeUpdate()
		assert.NoError(t, err)
		assert.NotEqual(t, "old_hash", m.Hash)

		// Should fail validation
		m.Amount = -100
		err = m.BeforeUpdate()
		assert.Error(t, err)
	})

}

// Test ETCListParams
func TestETCListParams_Complete(t *testing.T) {
	t.Run("SetDefaults", func(t *testing.T) {
		// Zero values get defaults
		p := &models.ETCListParams{}
		p.SetDefaults()
		assert.Equal(t, 100, p.Limit)
		assert.Equal(t, 0, p.Offset)

		// Negative limit gets default
		p.Limit = -10
		p.SetDefaults()
		assert.Equal(t, 100, p.Limit)

		// Over 1000 gets capped
		p.Limit = 2000
		p.SetDefaults()
		assert.Equal(t, 1000, p.Limit)

		// Valid values preserved
		p.Limit = 50
		p.Offset = 10
		p.SetDefaults()
		assert.Equal(t, 50, p.Limit)
		assert.Equal(t, 10, p.Offset)

		// Negative offset gets zero
		p.Offset = -5
		p.SetDefaults()
		assert.Equal(t, 0, p.Offset)
	})
}

// Test ETCMeisaiMapping model - 100% coverage
func TestETCMeisaiMapping_Complete(t *testing.T) {
	t.Run("Validate", func(t *testing.T) {
		// Valid case
		m := &models.ETCMeisaiMapping{
			ETCMeisaiID: 1,
			DTakoRowID:  "DTAKO_001",
			MappingType: "auto",
			Confidence:  0.8,
		}
		err := m.Validate()
		assert.NoError(t, err)

		// Invalid ETCMeisaiID
		m.ETCMeisaiID = 0
		err = m.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "ETCMeisaiID must be positive")

		m.ETCMeisaiID = -1
		err = m.Validate()
		assert.Error(t, err)

		// Empty DTakoRowID
		m.ETCMeisaiID = 1
		m.DTakoRowID = ""
		err = m.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "DTakoRowID is required")

		// Invalid MappingType
		m.DTakoRowID = "DTAKO_001"
		m.MappingType = "invalid"
		err = m.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "MappingType must be")

		// Invalid Confidence (negative)
		m.MappingType = "manual"
		m.Confidence = -0.1
		err = m.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Confidence must be between")

		// Invalid Confidence (>1)
		m.Confidence = 1.1
		err = m.Validate()
		assert.Error(t, err)

		// Valid boundary values
		m.Confidence = 0.0
		err = m.Validate()
		assert.NoError(t, err)

		m.Confidence = 1.0
		err = m.Validate()
		assert.NoError(t, err)
	})

	t.Run("BeforeCreate", func(t *testing.T) {
		m := &models.ETCMeisaiMapping{
			ETCMeisaiID: 1,
			DTakoRowID:  "DTAKO_001",
			MappingType: "auto",
			Confidence:  0.8,
		}
		err := m.BeforeCreate()
		assert.NoError(t, err)

		// Invalid data
		m.ETCMeisaiID = 0
		err = m.BeforeCreate()
		assert.Error(t, err)
	})

	t.Run("BeforeUpdate", func(t *testing.T) {
		m := &models.ETCMeisaiMapping{
			ETCMeisaiID: 1,
			DTakoRowID:  "DTAKO_001",
			MappingType: "manual",
			Confidence:  0.9,
		}
		err := m.BeforeUpdate()
		assert.NoError(t, err)

		// Invalid data
		m.MappingType = "unknown"
		err = m.BeforeUpdate()
		assert.Error(t, err)
	})

	t.Run("IsHighConfidence", func(t *testing.T) {
		m := &models.ETCMeisaiMapping{Confidence: 0.8}
		assert.True(t, m.IsHighConfidence())

		m.Confidence = 0.85
		assert.True(t, m.IsHighConfidence())

		m.Confidence = 0.79
		assert.False(t, m.IsHighConfidence())

		m.Confidence = 0.0
		assert.False(t, m.IsHighConfidence())
	})
}

// Test ETCImportBatch model - 100% coverage
func TestETCImportBatch_Complete(t *testing.T) {
	t.Run("Validate", func(t *testing.T) {
		// Valid cases for all statuses
		statuses := []string{"pending", "processing", "completed", "failed", "cancelled"}
		for _, status := range statuses {
			b := &models.ETCImportBatch{
				FileName:     "test.csv",
				Status:       status,
				TotalRecords: 100,
			}
			err := b.Validate()
			assert.NoError(t, err, "Status %s should be valid", status)
		}

		// Empty filename
		b := &models.ETCImportBatch{
			FileName: "",
			Status:   "pending",
		}
		err := b.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "FileName is required")

		// Negative TotalRecords
		b.FileName = "test.csv"
		b.TotalRecords = -1
		err = b.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "TotalRecords cannot be negative")

		// Invalid status
		b.TotalRecords = 0
		b.Status = "unknown"
		err = b.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid Status")
	})

	t.Run("BeforeCreate", func(t *testing.T) {
		b := &models.ETCImportBatch{
			FileName: "test.csv",
			Status:   "pending",
		}
		err := b.BeforeCreate()
		assert.NoError(t, err)

		// Invalid data
		b.FileName = ""
		err = b.BeforeCreate()
		assert.Error(t, err)
	})

	t.Run("BeforeUpdate", func(t *testing.T) {
		b := &models.ETCImportBatch{
			FileName: "test.csv",
			Status:   "processing",
		}
		err := b.BeforeUpdate()
		assert.NoError(t, err)

		// Invalid data
		b.Status = "invalid"
		err = b.BeforeUpdate()
		assert.Error(t, err)
	})

	t.Run("GetProgress", func(t *testing.T) {
		b := &models.ETCImportBatch{
			TotalRecords:   100,
			ProcessedCount: 50,
		}
		progress := b.GetProgress()
		assert.Equal(t, float32(50.0), progress)

		// Zero total records
		b.TotalRecords = 0
		progress = b.GetProgress()
		assert.Equal(t, float32(0), progress)

		// All processed
		b.TotalRecords = 100
		b.ProcessedCount = 100
		progress = b.GetProgress()
		assert.Equal(t, float32(100.0), progress)
	})

	t.Run("IsCompleted", func(t *testing.T) {
		b := &models.ETCImportBatch{Status: "completed"}
		assert.True(t, b.IsCompleted())

		b.Status = "failed"
		assert.True(t, b.IsCompleted())

		b.Status = "cancelled"
		assert.True(t, b.IsCompleted())

		b.Status = "pending"
		assert.False(t, b.IsCompleted())

		b.Status = "processing"
		assert.False(t, b.IsCompleted())
	})

	t.Run("GetDuration", func(t *testing.T) {
		// Nil StartTime
		b := &models.ETCImportBatch{}
		duration := b.GetDuration()
		assert.Nil(t, duration)

		// StartTime only (uses current time)
		start := time.Now().Add(-1 * time.Hour)
		b.StartTime = &start
		duration = b.GetDuration()
		require.NotNil(t, duration)
		assert.Greater(t, duration.Seconds(), float64(3500))

		// Both StartTime and CompleteTime
		complete := start.Add(30 * time.Minute)
		b.CompleteTime = &complete
		duration = b.GetDuration()
		require.NotNil(t, duration)
		assert.Equal(t, 30*time.Minute, *duration)
	})
}

// Test Validation functions - 100% coverage
func TestValidation_Complete(t *testing.T) {
	t.Run("ValidationError", func(t *testing.T) {
		err := models.ValidationError{
			Field:   "amount",
			Message: "Amount must be positive",
			Code:    "POSITIVE_REQUIRED",
			Value:   -100,
		}
		errStr := err.Error()
		assert.Contains(t, errStr, "amount")
		assert.Contains(t, errStr, "Amount must be positive")
	})

	t.Run("ValidationResult_AddError", func(t *testing.T) {
		result := &models.ValidationResult{Valid: true}
		result.AddError("field1", "message1", "CODE1", "value1")
		assert.False(t, result.Valid)
		assert.Len(t, result.Errors, 1)

		result.AddError("field2", "message2", "CODE2", "value2")
		assert.Len(t, result.Errors, 2)
	})

	t.Run("ValidateETCMeisai", func(t *testing.T) {
		// All validation paths
		m := &models.ETCMeisai{
			UseDate:   time.Now().Add(24 * time.Hour), // Future
			Amount:    -100,
			EntryIC:   "  ", // Spaces only
			ExitIC:    "",
			ETCNumber: "ABC", // Invalid format
		}
		result := models.ValidateETCMeisai(m)
		assert.False(t, result.Valid)
		assert.Greater(t, len(result.Errors), 3)

		// Old date
		m.UseDate = time.Now().AddDate(-3, 0, 0)
		m.Amount = 1500
		m.EntryIC = "東京IC"
		m.ExitIC = "横浜IC"
		m.ETCNumber = "12345"
		result = models.ValidateETCMeisai(m)
		assert.False(t, result.Valid)

		// Invalid time format
		m.UseDate = time.Now().Add(-24 * time.Hour)
		m.UseTime = "25:99"
		result = models.ValidateETCMeisai(m)
		assert.False(t, result.Valid)

		// Long fields
		longStr := fmt.Sprintf("%101s", "x")
		m.UseTime = "14:30"
		m.ETCNumber = "123456789012345678901"
		result = models.ValidateETCMeisai(m)
		assert.False(t, result.Valid)

		m.ETCNumber = "12345"
		m.CarNumber = "123456789012345678901"
		result = models.ValidateETCMeisai(m)
		assert.False(t, result.Valid)

		m.CarNumber = "1234"
		m.EntryIC = longStr
		result = models.ValidateETCMeisai(m)
		assert.False(t, result.Valid)

		m.EntryIC = "東京IC"
		m.ExitIC = longStr
		result = models.ValidateETCMeisai(m)
		assert.False(t, result.Valid)

		// Suspicious amount
		m.ExitIC = "横浜IC"
		m.Amount = 150000
		result = models.ValidateETCMeisai(m)
		assert.False(t, result.Valid)

		// Valid case
		m.Amount = 1500
		result = models.ValidateETCMeisai(m)
		assert.True(t, result.Valid)
	})

	t.Run("ValidateETCMeisaiMapping", func(t *testing.T) {
		// Invalid ETCMeisaiID
		m := &models.ETCMeisaiMapping{
			ETCMeisaiID: -1,
			DTakoRowID:  "DTAKO_001",
			MappingType: "auto",
			Confidence:  0.5,
		}
		result := models.ValidateETCMeisaiMapping(m)
		assert.False(t, result.Valid)

		// Empty DTakoRowID with spaces
		m.ETCMeisaiID = 1
		m.DTakoRowID = "  "
		result = models.ValidateETCMeisaiMapping(m)
		assert.False(t, result.Valid)

		// Invalid mapping type
		m.DTakoRowID = "DTAKO_001"
		m.MappingType = "unknown"
		result = models.ValidateETCMeisaiMapping(m)
		assert.False(t, result.Valid)

		// Invalid confidence
		m.MappingType = "manual"
		m.Confidence = -0.1
		result = models.ValidateETCMeisaiMapping(m)
		assert.False(t, result.Valid)

		m.Confidence = 1.1
		result = models.ValidateETCMeisaiMapping(m)
		assert.False(t, result.Valid)

		// Long notes
		m.Confidence = 0.8
		m.Notes = fmt.Sprintf("%501s", "x")
		result = models.ValidateETCMeisaiMapping(m)
		assert.False(t, result.Valid)

		// Invalid DTakoRowID format
		m.Notes = "test"
		m.DTakoRowID = "INVALID@ID!"
		result = models.ValidateETCMeisaiMapping(m)
		assert.False(t, result.Valid)

		// Long DTakoRowID
		m.DTakoRowID = fmt.Sprintf("%51s", "A")
		result = models.ValidateETCMeisaiMapping(m)
		assert.False(t, result.Valid)

		// Valid case
		m.DTakoRowID = "DTAKO_001"
		result = models.ValidateETCMeisaiMapping(m)
		assert.True(t, result.Valid)
	})

	t.Run("ValidateETCImportBatch", func(t *testing.T) {
		// Empty filename with spaces
		b := &models.ETCImportBatch{
			FileName: "  ",
			Status:   "pending",
		}
		result := models.ValidateETCImportBatch(b)
		assert.False(t, result.Valid)

		// Negative TotalRecords
		b.FileName = "test.csv"
		b.TotalRecords = -1
		result = models.ValidateETCImportBatch(b)
		assert.False(t, result.Valid)

		// Invalid status
		b.TotalRecords = 0
		b.Status = "unknown"
		result = models.ValidateETCImportBatch(b)
		assert.False(t, result.Valid)

		// Long filename
		b.Status = "pending"
		b.FileName = fmt.Sprintf("%256s", "x")
		result = models.ValidateETCImportBatch(b)
		assert.False(t, result.Valid)

		// ProcessedCount > TotalRecords
		b.FileName = "test.csv"
		b.TotalRecords = 10
		b.ProcessedCount = 20
		result = models.ValidateETCImportBatch(b)
		assert.False(t, result.Valid)

		// CreatedCount > ProcessedCount
		b.TotalRecords = 100
		b.ProcessedCount = 50
		b.CreatedCount = 60
		result = models.ValidateETCImportBatch(b)
		assert.False(t, result.Valid)

		// CompleteTime before StartTime
		start := time.Now()
		complete := start.Add(-1 * time.Hour)
		b.ProcessedCount = 50
		b.CreatedCount = 40
		b.StartTime = &start
		b.CompleteTime = &complete
		result = models.ValidateETCImportBatch(b)
		assert.False(t, result.Valid)

		// Valid case
		complete = start.Add(1 * time.Hour)
		b.CompleteTime = &complete
		result = models.ValidateETCImportBatch(b)
		assert.True(t, result.Valid)
	})

	t.Run("Helper_Functions", func(t *testing.T) {
		// isValidTimeFormat (called internally)
		m := &models.ETCMeisai{
			UseDate:   time.Now().Add(-24 * time.Hour),
			UseTime:   "14:30",
			Amount:    1500,
			EntryIC:   "東京IC",
			ExitIC:    "横浜IC",
			ETCNumber: "12345",
		}
		result := models.ValidateETCMeisai(m)
		assert.True(t, result.Valid)

		// isValidETCNumber (called internally)
		m.ETCNumber = "ABC123"
		result = models.ValidateETCMeisai(m)
		assert.False(t, result.Valid)

		// isValidDTakoRowID (called internally)
		mapping := &models.ETCMeisaiMapping{
			ETCMeisaiID: 1,
			DTakoRowID:  "VALID-ID_123",
			MappingType: "auto",
			Confidence:  0.8,
		}
		result2 := models.ValidateETCMeisaiMapping(mapping)
		assert.True(t, result2.Valid)
	})

	t.Run("ValidateETCMeisaiBatch", func(t *testing.T) {
		records := []*models.ETCMeisai{
			{
				UseDate:   time.Now().Add(-24 * time.Hour),
				Amount:    1500,
				EntryIC:   "東京IC",
				ExitIC:    "横浜IC",
				ETCNumber: "12345",
				Hash:      "hash1",
			},
			{
				UseDate:   time.Now().Add(-24 * time.Hour),
				Amount:    2000,
				EntryIC:   "横浜IC",
				ExitIC:    "名古屋IC",
				ETCNumber: "12345",
				Hash:      "hash1", // Duplicate
			},
		}

		// Default options (nil)
		results := models.ValidateETCMeisaiBatch(records, nil)
		assert.Len(t, results, 2)
		assert.False(t, results[1].Valid) // Duplicate hash

		// Skip duplicates
		options := &models.BatchValidationOptions{
			SkipDuplicates: true,
		}
		results = models.ValidateETCMeisaiBatch(records, options)
		assert.Len(t, results, 2)
		assert.True(t, results[1].Valid) // Duplicate check skipped

		// Max errors limit
		invalidRecords := make([]*models.ETCMeisai, 10)
		for i := 0; i < 10; i++ {
			invalidRecords[i] = &models.ETCMeisai{Amount: -100}
		}
		options.MaxErrors = 3
		results = models.ValidateETCMeisaiBatch(invalidRecords, options)
		assert.Len(t, results, 3) // Stopped at max errors

		// Empty hash doesn't trigger duplicate check
		records[0].Hash = ""
		records[1].Hash = ""
		options.SkipDuplicates = false
		results = models.ValidateETCMeisaiBatch(records, options)
		assert.Len(t, results, 2)
	})

	t.Run("SummarizeValidation", func(t *testing.T) {
		results := map[int]*models.ValidationResult{
			0: {Valid: true},
			1: {
				Valid: false,
				Errors: []models.ValidationError{
					{Field: "amount", Message: "msg1", Code: "CODE1", Value: -100},
					{Field: "etc_number", Message: "msg2", Code: "CODE2", Value: "ABC"},
				},
			},
			2: {
				Valid: false,
				Errors: []models.ValidationError{
					{Field: "amount", Message: "msg3", Code: "CODE1", Value: -200},
				},
			},
		}

		summary := models.SummarizeValidation(results, 2)
		assert.Equal(t, 3, summary.TotalRecords)
		assert.Equal(t, 1, summary.ValidRecords)
		assert.Equal(t, 2, summary.InvalidRecords)
		assert.Equal(t, 2, summary.ErrorsByField["amount"])
		assert.Equal(t, 1, summary.ErrorsByField["etc_number"])
		assert.Equal(t, 2, summary.ErrorsByCode["CODE1"])
		assert.Equal(t, 1, summary.ErrorsByCode["CODE2"])
		assert.Len(t, summary.FirstErrors, 2) // Limited to 2

		// Test with more errors than limit
		summary = models.SummarizeValidation(results, 10)
		assert.Len(t, summary.FirstErrors, 3) // All 3 errors
	})
}
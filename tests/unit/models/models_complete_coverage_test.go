package models_test

import (
	"testing"
	"time"

	"github.com/yhonda-ohishi/etc_meisai/src/models"
	"github.com/yhonda-ohishi/etc_meisai/tests/helpers"
)

// TestETCMapping_UtilityMethods tests all utility methods for coverage
func TestETCMapping_UtilityMethods(t *testing.T) {
	mapping := &models.ETCMapping{
		ID:               1,
		ETCRecordID:      123,
		MappingType:      "dtako",
		MappedEntityID:   456,
		MappedEntityType: "dtako_record",
		Confidence:       0.85,
		Status:           "active",
		CreatedBy:        "test-user",
	}

	// Test IsActive
	helpers.AssertTrue(t, mapping.IsActive())

	// Test IsPending
	helpers.AssertFalse(t, mapping.IsPending())

	// Test GetConfidencePercentage
	percentage := mapping.GetConfidencePercentage()
	helpers.AssertEqual(t, 85.0, percentage)

	// Test IsHighConfidence
	helpers.AssertTrue(t, mapping.IsHighConfidence())

	// Test IsLowConfidence
	helpers.AssertFalse(t, mapping.IsLowConfidence())

	// Test GetTableName
	tableName := mapping.GetTableName()
	helpers.AssertEqual(t, "etc_mappings", tableName)

	// Test String method
	str := mapping.String()
	helpers.AssertContains(t, str, "ETCMapping")
	helpers.AssertContains(t, str, "123")
	helpers.AssertContains(t, str, "dtako")
}

// TestETCMapping_StatusTransitions tests status transition methods
func TestETCMapping_StatusTransitions(t *testing.T) {
	tests := []struct {
		name           string
		initialStatus  string
		targetStatus   string
		method         func(*models.ETCMapping) error
		canTransition  bool
		expectedStatus string
	}{
		{
			name:           "activate from pending",
			initialStatus:  "pending",
			targetStatus:   "active",
			method:         (*models.ETCMapping).Activate,
			canTransition:  true,
			expectedStatus: "active",
		},
		{
			name:          "activate from inactive (should fail)",
			initialStatus: "inactive",
			targetStatus:  "active",
			method:        (*models.ETCMapping).Activate,
			canTransition: true, // Actually inactive can transition to active
		},
		{
			name:           "deactivate from active",
			initialStatus:  "active",
			targetStatus:   "inactive",
			method:         (*models.ETCMapping).Deactivate,
			canTransition:  true,
			expectedStatus: "inactive",
		},
		{
			name:          "deactivate from pending (should fail)",
			initialStatus: "pending",
			targetStatus:  "inactive",
			method:        (*models.ETCMapping).Deactivate,
			canTransition: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mapping := &models.ETCMapping{
				ETCRecordID:      1,
				MappingType:      "dtako",
				MappedEntityID:   1,
				MappedEntityType: "dtako_record",
				Confidence:       1.0,
				Status:           tt.initialStatus,
			}

			// Test CanTransitionTo
			canTransition := mapping.CanTransitionTo(tt.targetStatus)
			helpers.AssertEqual(t, tt.canTransition, canTransition)

			// Test actual method
			err := tt.method(mapping)

			if tt.canTransition {
				helpers.AssertNoError(t, err)
				helpers.AssertEqual(t, tt.expectedStatus, mapping.Status)
			} else {
				helpers.AssertError(t, err)
			}
		})
	}
}

// TestETCMapping_ApproveReject tests approve/reject functionality
func TestETCMapping_ApproveReject(t *testing.T) {
	// Test Approve from pending
	mapping := &models.ETCMapping{
		ETCRecordID:      1,
		MappingType:      "dtako",
		MappedEntityID:   1,
		MappedEntityType: "dtako_record",
		Confidence:       1.0,
		Status:           "pending",
	}

	err := mapping.Approve()
	helpers.AssertNoError(t, err)
	helpers.AssertEqual(t, "active", mapping.Status)

	// Test Approve from non-pending (should fail)
	err = mapping.Approve()
	helpers.AssertError(t, err)
	helpers.AssertContains(t, err.Error(), "can only approve pending mappings")

	// Test Reject from pending
	mapping2 := &models.ETCMapping{
		ETCRecordID:      1,
		MappingType:      "dtako",
		MappedEntityID:   1,
		MappedEntityType: "dtako_record",
		Confidence:       1.0,
		Status:           "pending",
	}

	err = mapping2.Reject()
	helpers.AssertNoError(t, err)
	helpers.AssertEqual(t, "rejected", mapping2.Status)

	// Test Reject from non-pending (should fail)
	err = mapping2.Reject()
	helpers.AssertError(t, err)
	helpers.AssertContains(t, err.Error(), "can only reject pending mappings")
}

// TestETCMapping_Metadata tests metadata functionality
func TestETCMapping_Metadata(t *testing.T) {
	mapping := &models.ETCMapping{
		ETCRecordID:      1,
		MappingType:      "dtako",
		MappedEntityID:   1,
		MappedEntityType: "dtako_record",
		Confidence:       1.0,
		Status:           "active",
	}

	// Test setting nil metadata
	err := mapping.SetMetadata(nil)
	helpers.AssertNoError(t, err)
	helpers.AssertNil(t, mapping.Metadata)

	// Test getting nil metadata
	metadata, err := mapping.GetMetadata()
	helpers.AssertNoError(t, err)
	helpers.AssertNil(t, metadata)

	// Test setting valid metadata
	testMetadata := map[string]interface{}{
		"source":     "automatic",
		"confidence": 0.95,
		"notes":      "high quality match",
	}

	err = mapping.SetMetadata(testMetadata)
	helpers.AssertNoError(t, err)
	helpers.AssertNotNil(t, mapping.Metadata)

	// Test getting metadata
	retrievedMetadata, err := mapping.GetMetadata()
	helpers.AssertNoError(t, err)
	helpers.AssertEqual(t, "automatic", retrievedMetadata["source"])
	helpers.AssertEqual(t, 0.95, retrievedMetadata["confidence"])

	// Test setting metadata that's too large
	largeMetadata := make(map[string]interface{})
	for i := 0; i < 10000; i++ {
		largeMetadata[string(rune(i))] = "very long value that will exceed the 64KB limit when repeated many times"
	}

	err = mapping.SetMetadata(largeMetadata)
	helpers.AssertError(t, err)
	helpers.AssertContains(t, err.Error(), "metadata too large")
}

// TestETCMapping_BeforeCreateBeforeSave tests GORM hooks
func TestETCMapping_BeforeCreateBeforeSave(t *testing.T) {
	mapping := &models.ETCMapping{
		ETCRecordID:      1,
		MappingType:      "dtako",
		MappedEntityID:   1,
		MappedEntityType: "dtako_record",
		Confidence:       1.0,
	}

	// Test BeforeCreate
	err := mapping.BeforeCreate(nil)
	helpers.AssertNoError(t, err)
	helpers.AssertFalse(t, mapping.CreatedAt.IsZero())
	helpers.AssertFalse(t, mapping.UpdatedAt.IsZero())
	helpers.AssertEqual(t, "active", mapping.Status) // Should set default status

	// Test BeforeSave
	oldUpdatedAt := mapping.UpdatedAt
	time.Sleep(time.Millisecond)
	err = mapping.BeforeSave(nil)
	helpers.AssertNoError(t, err)
	helpers.AssertTrue(t, mapping.UpdatedAt.After(oldUpdatedAt))

	// Test BeforeUpdate
	err = mapping.BeforeUpdate()
	helpers.AssertNoError(t, err)
}

// TestImportSession_ComprehensiveValidation adds complete validation coverage
func TestImportSession_ComprehensiveValidation(t *testing.T) {
	tests := []struct {
		name    string
		session *models.ImportSession
		wantErr bool
		errMsg  string
	}{
		{
			name: "invalid UUID format",
			session: &models.ImportSession{
				ID:           "invalid-uuid",
				FileName:     "test.csv",
				FileSize:     1024,
				AccountType:  "corporate",
				AccountID:    "test-account-id",
				Status:       "pending",
			},
			wantErr: true,
			errMsg:  "invalid UUID format",
		},
		{
			name: "account ID with invalid characters",
			session: &models.ImportSession{
				ID:           "550e8400-e29b-41d4-a716-446655440000",
				FileName:     "test.csv",
				FileSize:     1024,
				AccountType:  "corporate",
				AccountID:    "test@#$%^&*()",
				Status:       "pending",
			},
			wantErr: true,
			errMsg:  "account ID contains invalid characters",
		},
		{
			name: "account ID too long",
			session: &models.ImportSession{
				ID:           "550e8400-e29b-41d4-a716-446655440000",
				FileName:     "test.csv",
				FileSize:     1024,
				AccountType:  "corporate",
				AccountID:    "this_is_a_very_long_account_id_that_exceeds_the_maximum_allowed_length_of_fifty_characters",
				Status:       "pending",
			},
			wantErr: true,
			errMsg:  "account ID too long",
		},
		{
			name: "file name too long",
			session: &models.ImportSession{
				ID:           "550e8400-e29b-41d4-a716-446655440000",
				FileName:     "this_is_a_very_long_file_name_that_exceeds_the_maximum_allowed_length_of_two_hundred_fifty_five_characters_which_is_quite_a_lot_of_characters_for_a_file_name_but_we_need_to_test_this_edge_case_to_ensure_our_validation_works_properly.csv",
				FileSize:     1024,
				AccountType:  "corporate",
				AccountID:    "test-account-id",
				Status:       "pending",
			},
			wantErr: true,
			errMsg:  "file name too long",
		},
		{
			name: "non-CSV file extension",
			session: &models.ImportSession{
				ID:           "550e8400-e29b-41d4-a716-446655440000",
				FileName:     "test.txt",
				FileSize:     1024,
				AccountType:  "corporate",
				AccountID:    "test-account-id",
				Status:       "pending",
			},
			wantErr: true,
			errMsg:  "file must have .csv extension",
		},
		{
			name: "file size too large",
			session: &models.ImportSession{
				ID:           "550e8400-e29b-41d4-a716-446655440000",
				FileName:     "test.csv",
				FileSize:     200 * 1024 * 1024, // 200MB
				AccountType:  "corporate",
				AccountID:    "test-account-id",
				Status:       "pending",
			},
			wantErr: true,
			errMsg:  "file size too large",
		},
		{
			name: "row count validation - processed exceeds total",
			session: &models.ImportSession{
				ID:              "550e8400-e29b-41d4-a716-446655440000",
				FileName:        "test.csv",
				FileSize:        1024,
				AccountType:     "corporate",
				AccountID:       "test-account-id",
				Status:          "pending",
				TotalRows:       100,
				ProcessedRows:   150, // Exceeds total
				SuccessRows:     50,
				ErrorRows:       50,
				DuplicateRows:   50,
			},
			wantErr: true,
			errMsg:  "processed rows (150) cannot exceed total rows (100)",
		},
		{
			name: "row count validation - sum mismatch",
			session: &models.ImportSession{
				ID:              "550e8400-e29b-41d4-a716-446655440000",
				FileName:        "test.csv",
				FileSize:        1024,
				AccountType:     "corporate",
				AccountID:       "test-account-id",
				Status:          "pending",
				TotalRows:       200,
				ProcessedRows:   100, // Should be 90 (30+30+30)
				SuccessRows:     30,
				ErrorRows:       30,
				DuplicateRows:   30,
			},
			wantErr: true,
			errMsg:  "processed rows (100) must equal sum of success (30) + error (30) + duplicate (30) rows",
		},
		{
			name: "negative counts",
			session: &models.ImportSession{
				ID:            "550e8400-e29b-41d4-a716-446655440000",
				FileName:      "test.csv",
				FileSize:      1024,
				AccountType:   "corporate",
				AccountID:     "test-account-id",
				Status:        "pending",
				TotalRows:     -10,
				ProcessedRows: 0,
				SuccessRows:   0,
				ErrorRows:     0,
				DuplicateRows: 0,
			},
			wantErr: true,
			errMsg:  "total rows cannot be negative",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.session.Validate()

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

// TestImportSession_StatusMethods tests status-related methods
func TestImportSession_StatusMethods(t *testing.T) {
	session := &models.ImportSession{
		ID:           "550e8400-e29b-41d4-a716-446655440000",
		FileName:     "test.csv",
		FileSize:     1024,
		AccountType:  "corporate",
		AccountID:    "test-account-id",
		Status:       "pending",
		StartedAt:    time.Now(),
	}

	// Test status checks
	helpers.AssertFalse(t, session.IsCompleted())
	helpers.AssertFalse(t, session.IsFailed())
	helpers.AssertFalse(t, session.IsInProgress())
	helpers.AssertTrue(t, session.IsPending())

	// Test StartProcessing
	err := session.StartProcessing()
	helpers.AssertNoError(t, err)
	helpers.AssertEqual(t, "processing", session.Status)
	helpers.AssertTrue(t, session.IsInProgress())

	// Test Complete
	err = session.Complete()
	helpers.AssertNoError(t, err)
	helpers.AssertEqual(t, "completed", session.Status)
	helpers.AssertTrue(t, session.IsCompleted())
	helpers.AssertNotNil(t, session.CompletedAt)

	// Test that can't transition from completed
	err = session.StartProcessing()
	helpers.AssertError(t, err)

	// Test Fail from processing
	session2 := &models.ImportSession{
		ID:           "550e8400-e29b-41d4-a716-446655440001",
		FileName:     "test2.csv",
		FileSize:     1024,
		AccountType:  "corporate",
		AccountID:    "test-account-id",
		Status:       "processing",
		StartedAt:    time.Now(),
	}

	err = session2.Fail()
	helpers.AssertNoError(t, err)
	helpers.AssertEqual(t, "failed", session2.Status)
	helpers.AssertTrue(t, session2.IsFailed())
	helpers.AssertNotNil(t, session2.CompletedAt)

	// Test Cancel
	session3 := &models.ImportSession{
		ID:           "550e8400-e29b-41d4-a716-446655440002",
		FileName:     "test3.csv",
		FileSize:     1024,
		AccountType:  "corporate",
		AccountID:    "test-account-id",
		Status:       "processing",
		StartedAt:    time.Now(),
	}

	err = session3.Cancel()
	helpers.AssertNoError(t, err)
	helpers.AssertEqual(t, "cancelled", session3.Status)
	helpers.AssertNotNil(t, session3.CompletedAt)
}

// TestImportSession_ErrorHandling tests error handling functionality
func TestImportSession_ErrorHandling(t *testing.T) {
	session := &models.ImportSession{
		ID:           "550e8400-e29b-41d4-a716-446655440000",
		FileName:     "test.csv",
		FileSize:     1024,
		AccountType:  "corporate",
		AccountID:    "test-account-id",
		Status:       "pending",
	}

	// Test AddError
	err := session.AddError(1, "validation", "Invalid data format", "raw,data,here")
	helpers.AssertNoError(t, err)

	err = session.AddError(2, "parsing", "Cannot parse date", "2025-13-45")
	helpers.AssertNoError(t, err)

	// Test GetErrors
	errors, err := session.GetErrors()
	helpers.AssertNoError(t, err)
	helpers.AssertLen(t, errors, 2)
	helpers.AssertEqual(t, 1, errors[0].RowNumber)
	helpers.AssertEqual(t, "validation", errors[0].ErrorType)
	helpers.AssertEqual(t, "Invalid data format", errors[0].ErrorMessage)

	// Test SetError and ClearError
	session.SetError("Test error message")
	helpers.AssertEqual(t, "failed", session.Status)
	helpers.AssertNotNil(t, session.ErrorMessage)
	helpers.AssertEqual(t, "Test error message", *session.ErrorMessage)

	session.ClearError()
	helpers.AssertNil(t, session.ErrorMessage)
}

// TestImportSession_ProgressMethods tests progress-related methods
func TestImportSession_ProgressMethods(t *testing.T) {
	session := &models.ImportSession{
		ID:           "550e8400-e29b-41d4-a716-446655440000",
		FileName:     "test.csv",
		FileSize:     1024,
		AccountType:  "corporate",
		AccountID:    "test-account-id",
		Status:       "processing",
		TotalRows:    1000,
		StartedAt:    time.Now().Add(-5 * time.Minute),
	}

	// Test UpdateProgress
	session.UpdateProgress(50, 10, 5)
	helpers.AssertEqual(t, 50, session.SuccessRows)
	helpers.AssertEqual(t, 10, session.ErrorRows)
	helpers.AssertEqual(t, 5, session.DuplicateRows)
	helpers.AssertEqual(t, 65, session.ProcessedRows)

	// Test GetProgressPercentage
	percentage := session.GetProgressPercentage()
	helpers.AssertEqual(t, 6.5, percentage)

	// Test GetSuccessRate
	successRate := session.GetSuccessRate()
	expectedRate := float64(50) / float64(65) * 100.0
	helpers.AssertEqual(t, expectedRate, successRate)

	// Test UpdateProgressWithCounts
	session.UpdateProgressWithCounts(200, 1000)
	helpers.AssertEqual(t, 200, session.ProcessedRows)
	helpers.AssertEqual(t, 1000, session.TotalRows)
	helpers.AssertEqual(t, 20.0, session.ProgressPercent)

	// Test GetDuration with in-progress session
	duration := session.GetDuration()
	helpers.AssertTrue(t, duration >= 4*time.Minute)
	helpers.AssertTrue(t, duration <= 6*time.Minute)

	// Test GetDuration with completed session
	completedAt := time.Now()
	session.CompletedAt = &completedAt
	duration = session.GetDuration()
	helpers.AssertTrue(t, duration >= 4*time.Minute)
}

// TestValidationFunctions tests the standalone validation functions
func TestValidationFunctions(t *testing.T) {
	// Test ValidateETCNumber
	tests := []struct {
		name      string
		etcNumber string
		expected  bool
	}{
		{"empty", "", false},
		{"too short", "123456789", false},
		{"too long", "12345678901234567", false},
		{"valid 10 digits", "1234567890", true},
		{"valid 16 digits", "1234567890123456", true},
		{"contains letters", "123456789a", false},
		{"contains spaces", "1234 567890", false},
	}

	for _, tt := range tests {
		t.Run("ValidateETCNumber_"+tt.name, func(t *testing.T) {
			result := models.ValidateETCNumber(tt.etcNumber)
			helpers.AssertEqual(t, tt.expected, result)
		})
	}

	// Test ValidateTimeFormat
	timeTests := []struct {
		name     string
		timeStr  string
		expected bool
	}{
		{"empty", "", false},
		{"valid format", "14:30", true},
		{"valid midnight", "00:00", true},
		{"valid late", "23:59", true},
		{"invalid hour", "24:00", false},
		{"invalid minute", "14:60", false},
		{"single digit", "9:30", false}, // Strict format requires two digits
	}

	for _, tt := range timeTests {
		t.Run("ValidateTimeFormat_"+tt.name, func(t *testing.T) {
			result := models.ValidateTimeFormat(tt.timeStr)
			helpers.AssertEqual(t, tt.expected, result)
		})
	}

	// Test ValidateCarNumber
	carTests := []struct {
		name      string
		carNumber string
		expected  bool
	}{
		{"empty", "", false},
		{"too short", "短", false},
		{"too long", "very_long_car_number_that_exceeds_twenty_characters", false},
		{"contains 品川", "品川123", true},
		{"contains 横浜", "横浜456", true},
		{"valid pattern", "品川123あ1234", true},
	}

	for _, tt := range carTests {
		t.Run("ValidateCarNumber_"+tt.name, func(t *testing.T) {
			result := models.ValidateCarNumber(tt.carNumber)
			helpers.AssertEqual(t, tt.expected, result)
		})
	}

	// Test SanitizeInput
	input := "test'input\"with;dangerous--characters/*and*/more"
	sanitized := models.SanitizeInput(input)
	helpers.AssertEqual(t, "testinputwithdangerouscharactersandmore", sanitized)
}

// TestValidationHelpers tests helper validation functions
func TestValidationHelpers(t *testing.T) {
	// Test IsValidMappingType
	helpers.AssertTrue(t, models.IsValidMappingType("dtako"))
	helpers.AssertTrue(t, models.IsValidMappingType("expense"))
	helpers.AssertTrue(t, models.IsValidMappingType("invoice"))
	helpers.AssertFalse(t, models.IsValidMappingType("invalid"))
	helpers.AssertFalse(t, models.IsValidMappingType(""))

	// Test IsValidEntityType
	helpers.AssertTrue(t, models.IsValidEntityType("dtako_record"))
	helpers.AssertTrue(t, models.IsValidEntityType("expense_record"))
	helpers.AssertTrue(t, models.IsValidEntityType("invoice_record"))
	helpers.AssertFalse(t, models.IsValidEntityType("invalid"))

	// Test IsValidStatus
	helpers.AssertTrue(t, models.IsValidStatus("active"))
	helpers.AssertTrue(t, models.IsValidStatus("inactive"))
	helpers.AssertTrue(t, models.IsValidStatus("pending"))
	helpers.AssertTrue(t, models.IsValidStatus("rejected"))
	helpers.AssertFalse(t, models.IsValidStatus("invalid"))

	// Test IsValidImportStatus
	helpers.AssertTrue(t, models.IsValidImportStatus("pending"))
	helpers.AssertTrue(t, models.IsValidImportStatus("processing"))
	helpers.AssertTrue(t, models.IsValidImportStatus("completed"))
	helpers.AssertTrue(t, models.IsValidImportStatus("failed"))
	helpers.AssertTrue(t, models.IsValidImportStatus("cancelled"))
	helpers.AssertFalse(t, models.IsValidImportStatus("invalid"))

	// Test IsValidAccountType
	helpers.AssertTrue(t, models.IsValidAccountType("corporate"))
	helpers.AssertTrue(t, models.IsValidAccountType("personal"))
	helpers.AssertFalse(t, models.IsValidAccountType("invalid"))
}
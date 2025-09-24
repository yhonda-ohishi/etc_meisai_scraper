package models_test

import (
	"testing"
	"time"

	"github.com/yhonda-ohishi/etc_meisai/src/models"
	"github.com/yhonda-ohishi/etc_meisai/tests/helpers"
)

func TestImportSession_BeforeCreate(t *testing.T) {
	session := &models.ImportSession{
		FileName:     "test.csv",
		FileSize:     1024,
		AccountType:  "corporate",
		AccountID:    "test-account-id",
		AccountIndex: 0,
		Status:       "pending",
	}

	err := session.BeforeCreate(nil)
	helpers.AssertNoError(t, err)

	// Should set timestamps
	helpers.AssertFalse(t, session.CreatedAt.IsZero())
	helpers.AssertFalse(t, session.UpdatedAt.IsZero())

	// Should set default status if empty
	session2 := &models.ImportSession{
		FileName:     "test2.csv",
		FileSize:     2048,
		AccountType:  "personal",
		AccountID:    "test-account-id-2",
		AccountIndex: 1,
	}

	err = session2.BeforeCreate(nil)
	helpers.AssertNoError(t, err)
	helpers.AssertEqual(t, "pending", session2.Status)
}

func TestImportSession_BeforeUpdate(t *testing.T) {
	session := &models.ImportSession{
		FileName:     "test.csv",
		FileSize:     1024,
		AccountType:  "corporate",
		AccountIndex: 0,
		Status:       "processing",
		StartedAt:    time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		CreatedAt:    time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		UpdatedAt:    time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
	}

	oldStartedAt := session.StartedAt
	time.Sleep(time.Millisecond) // Ensure time difference

	err := session.BeforeUpdate()
	helpers.AssertNoError(t, err)

	// Should update started timestamp
	helpers.AssertTrue(t, session.StartedAt.After(oldStartedAt))

	// Should not change created timestamp
	helpers.AssertEqual(t, time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), session.CreatedAt)
}

func TestImportSession_Validate(t *testing.T) {
	tests := []struct {
		name    string
		session *models.ImportSession
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid session",
			session: &models.ImportSession{
				ID:           "550e8400-e29b-41d4-a716-446655440000",
				FileName:     "test.csv",
				FileSize:     1024,
				AccountType:  "corporate",
				AccountID:    "test-account-id",
				AccountIndex: 0,
				Status:       "pending",
			},
			wantErr: false,
		},
		{
			name: "empty file name",
			session: &models.ImportSession{
				ID:           "550e8400-e29b-41d4-a716-446655440000",
				FileSize:     1024,
				AccountType:  "corporate",
				AccountID:    "test-account-id",
				AccountIndex: 0,
				Status:       "pending",
			},
			wantErr: true,
			errMsg:  "file name cannot be empty",
		},
		{
			name: "zero file size",
			session: &models.ImportSession{
				ID:           "550e8400-e29b-41d4-a716-446655440000",
				FileName:     "test.csv",
				FileSize:     0,
				AccountType:  "corporate",
				AccountID:    "test-account-id",
				AccountIndex: 0,
				Status:       "pending",
			},
			wantErr: true,
			errMsg:  "file size must be greater than 0",
		},
		{
			name: "negative file size",
			session: &models.ImportSession{
				FileName:     "test.csv",
				FileSize:     -100,
				AccountType:  "corporate",
				AccountIndex: 0,
				Status:       "pending",
			},
			wantErr: true,
			errMsg:  "FileSize must be positive",
		},
		{
			name: "empty account type",
			session: &models.ImportSession{
				FileName:     "test.csv",
				FileSize:     1024,
				AccountIndex: 0,
				Status:       "pending",
			},
			wantErr: true,
			errMsg:  "AccountType is required",
		},
		{
			name: "invalid account type",
			session: &models.ImportSession{
				FileName:     "test.csv",
				FileSize:     1024,
				AccountType:  "invalid",
				AccountIndex: 0,
				Status:       "pending",
			},
			wantErr: true,
			errMsg:  "AccountType must be 'corporate' or 'personal'",
		},
		{
			name: "negative account index",
			session: &models.ImportSession{
				FileName:     "test.csv",
				FileSize:     1024,
				AccountType:  "corporate",
				AccountIndex: -1,
				Status:       "pending",
			},
			wantErr: true,
			errMsg:  "AccountIndex must be non-negative",
		},
		{
			name: "invalid status",
			session: &models.ImportSession{
				FileName:     "test.csv",
				FileSize:     1024,
				AccountType:  "corporate",
				AccountIndex: 0,
				Status:       "invalid_status",
			},
			wantErr: true,
			errMsg:  "Invalid Status",
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

func TestImportSession_IsValidStatus(t *testing.T) {
	tests := []struct {
		name     string
		status   string
		expected bool
	}{
		{
			name:     "valid pending status",
			status:   "pending",
			expected: true,
		},
		{
			name:     "valid processing status",
			status:   "processing",
			expected: true,
		},
		{
			name:     "valid completed status",
			status:   "completed",
			expected: true,
		},
		{
			name:     "valid failed status",
			status:   "failed",
			expected: true,
		},
		{
			name:     "invalid status",
			status:   "invalid",
			expected: false,
		},
		{
			name:     "empty status",
			status:   "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := models.IsValidImportStatus(tt.status)
			helpers.AssertEqual(t, tt.expected, result)
		})
	}
}

func TestImportSession_IsValidAccountType(t *testing.T) {
	tests := []struct {
		name        string
		accountType string
		expected    bool
	}{
		{
			name:        "valid corporate type",
			accountType: "corporate",
			expected:    true,
		},
		{
			name:        "valid personal type",
			accountType: "personal",
			expected:    true,
		},
		{
			name:        "invalid type",
			accountType: "invalid",
			expected:    false,
		},
		{
			name:        "empty type",
			accountType: "",
			expected:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := models.IsValidAccountType(tt.accountType)
			helpers.AssertEqual(t, tt.expected, result)
		})
	}
}

func TestImportSession_GetTableName(t *testing.T) {
	session := &models.ImportSession{}
	tableName := session.GetTableName()
	helpers.AssertEqual(t, "import_sessions", tableName)
}

func TestImportSession_String(t *testing.T) {
	session := &models.ImportSession{
		ID:           "test-session-1",
		FileName:     "test_data.csv",
		FileSize:     2048,
		AccountType:  "corporate",
		AccountIndex: 2,
		Status:       "completed",
	}

	str := session.String()

	// Should contain key information
	helpers.AssertContains(t, str, "test_data.csv") // File Name
	helpers.AssertContains(t, str, "2048")          // File Size
	helpers.AssertContains(t, str, "corporate")     // Account Type
	helpers.AssertContains(t, str, "completed")     // Status
}

func TestImportSession_UpdateProgress(t *testing.T) {
	session := &models.ImportSession{
		FileName:        "test.csv",
		FileSize:        1024,
		AccountType:     "corporate",
		AccountIndex:    0,
		Status:          "processing",
		ProcessedRows:   0,
		TotalRows:       100,
		ProgressPercent: 0.0,
	}

	// Update progress to 50%
	session.UpdateProgressWithCounts(50, 100)

	helpers.AssertEqual(t, int(50), session.ProcessedRows)
	helpers.AssertEqual(t, int(100), session.TotalRows)
	helpers.AssertEqual(t, 50.0, session.ProgressPercent)

	// Update progress to 100%
	session.UpdateProgressWithCounts(100, 100)

	helpers.AssertEqual(t, int(100), session.ProcessedRows)
	helpers.AssertEqual(t, int(100), session.TotalRows)
	helpers.AssertEqual(t, 100.0, session.ProgressPercent)
}

func TestImportSession_SetError(t *testing.T) {
	session := &models.ImportSession{
		FileName:     "test.csv",
		FileSize:     1024,
		AccountType:  "corporate",
		AccountIndex: 0,
		Status:       "processing",
	}

	errorMsg := "Failed to parse CSV file"
	session.SetError(errorMsg)

	helpers.AssertEqual(t, "failed", session.Status)
	helpers.AssertNotNil(t, session.ErrorMessage)
	helpers.AssertEqual(t, errorMsg, *session.ErrorMessage)
}

func TestImportSession_ClearError(t *testing.T) {
	session := &models.ImportSession{
		FileName:     "test.csv",
		FileSize:     1024,
		AccountType:  "corporate",
		AccountIndex: 0,
		Status:       "failed",
		ErrorMessage: stringPtr("Previous error"),
	}

	session.ClearError()

	helpers.AssertNil(t, session.ErrorMessage)
	// Status should remain unchanged when clearing error
	helpers.AssertEqual(t, "failed", session.Status)
}

func TestImportSession_IsCompleted(t *testing.T) {
	tests := []struct {
		name     string
		status   string
		expected bool
	}{
		{
			name:     "completed status",
			status:   "completed",
			expected: true,
		},
		{
			name:     "failed status",
			status:   "failed",
			expected: true,
		},
		{
			name:     "pending status",
			status:   "pending",
			expected: false,
		},
		{
			name:     "processing status",
			status:   "processing",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			session := &models.ImportSession{
				Status: tt.status,
			}
			result := session.IsCompleted()
			helpers.AssertEqual(t, tt.expected, result)
		})
	}
}

func TestImportSession_GetDuration(t *testing.T) {
	now := time.Now()
	completedAt := now
	session := &models.ImportSession{
		StartedAt:   now.Add(-5 * time.Minute),
		CompletedAt: &completedAt,
	}

	duration := session.GetDuration()
	helpers.AssertTrue(t, duration >= 4*time.Minute)
	helpers.AssertTrue(t, duration <= 6*time.Minute)
}

func TestImportBatch_BeforeCreate(t *testing.T) {
	batch := &models.ImportBatch{
		SessionID:    1,
		BatchNumber:  1,
		RecordCount:  100,
		Status:       "pending",
	}

	err := batch.BeforeCreate()
	helpers.AssertNoError(t, err)

	// Should set timestamps
	helpers.AssertFalse(t, batch.CreatedAt.IsZero())
	helpers.AssertFalse(t, batch.UpdatedAt.IsZero())
}

func TestImportBatch_Validate(t *testing.T) {
	tests := []struct {
		name    string
		batch   *models.ImportBatch
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid batch",
			batch: &models.ImportBatch{
				SessionID:    1,
				BatchNumber:  1,
				RecordCount:  100,
				Status:       "pending",
			},
			wantErr: false,
		},
		{
			name: "zero session ID",
			batch: &models.ImportBatch{
				SessionID:    0,
				BatchNumber:  1,
				RecordCount:  100,
				Status:       "pending",
			},
			wantErr: true,
			errMsg:  "SessionID is required",
		},
		{
			name: "zero batch number",
			batch: &models.ImportBatch{
				SessionID:    1,
				BatchNumber:  0,
				RecordCount:  100,
				Status:       "pending",
			},
			wantErr: true,
			errMsg:  "BatchNumber must be positive",
		},
		{
			name: "negative record count",
			batch: &models.ImportBatch{
				SessionID:    1,
				BatchNumber:  1,
				RecordCount:  -10,
				Status:       "pending",
			},
			wantErr: true,
			errMsg:  "RecordCount must be non-negative",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.batch.Validate()

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
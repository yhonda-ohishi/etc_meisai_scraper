package services_test

import (
	"context"
	"io"
	"log"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/yhonda-ohishi/etc_meisai/src/models"
	"github.com/yhonda-ohishi/etc_meisai/src/services"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// MockDB is a mock implementation of gorm.DB for testing
type MockDB struct {
	mock.Mock
}

func setupTestDB(t *testing.T) *gorm.DB {
	// Use in-memory SQLite for testing
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}

	// Auto-migrate the necessary models
	err = db.AutoMigrate(
		&models.ImportSession{},
		&models.ETCMeisaiRecord{},
		&models.ImportError{},
	)
	if err != nil {
		t.Fatalf("Failed to migrate models: %v", err)
	}

	return db
}

func TestImportService_NewImportService(t *testing.T) {
	db := setupTestDB(t)
	logger := log.New(os.Stdout, "TEST: ", log.LstdFlags)

	service := services.NewImportService(db, logger)
	assert.NotNil(t, service)
}

func TestImportService_ImportCSV(t *testing.T) {
	db := setupTestDB(t)
	logger := log.New(os.Stdout, "TEST: ", log.LstdFlags)
	service := services.NewImportService(db, logger)

	validCSV := `使用日,使用時刻,入口IC,出口IC,通行料金,車両番号,ETCカード番号
2025/01/15,09:30,東京IC,大阪IC,1000,品川123あ1234,1234567890123456
2025/01/15,10:30,名古屋IC,京都IC,800,品川456い5678,9876543210123456`

	tests := []struct {
		name    string
		params  *services.ImportCSVParams
		data    io.Reader
		wantErr bool
		errMsg  string
	}{
		{
			name: "successful import",
			params: &services.ImportCSVParams{
				AccountType: "corporate",
				AccountID:   "test-account",
				FileName:    "test.csv",
				FileSize:    int64(len(validCSV)),
				CreatedBy:   "test-user",
			},
			data:    strings.NewReader(validCSV),
			wantErr: false,
		},
		{
			name:    "nil params",
			params:  nil,
			data:    strings.NewReader(validCSV),
			wantErr: true,
			errMsg:  "params cannot be nil",
		},
		{
			name: "empty file name",
			params: &services.ImportCSVParams{
				AccountType: "corporate",
				AccountID:   "test-account",
				FileName:    "",
				FileSize:    100,
			},
			data:    strings.NewReader(validCSV),
			wantErr: true,
			errMsg:  "file name is required",
		},
		{
			name: "invalid CSV format",
			params: &services.ImportCSVParams{
				AccountType: "corporate",
				AccountID:   "test-account",
				FileName:    "test.csv",
				FileSize:    100,
			},
			data:    strings.NewReader("invalid,csv,data"),
			wantErr: true,
			errMsg:  "invalid CSV",
		},
		{
			name: "empty CSV data",
			params: &services.ImportCSVParams{
				AccountType: "corporate",
				AccountID:   "test-account",
				FileName:    "test.csv",
				FileSize:    100,
			},
			data:    strings.NewReader(""),
			wantErr: true,
			errMsg:  "empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			result, err := service.ImportCSV(ctx, tt.params, tt.data)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				if result != nil {
					assert.Greater(t, result.SuccessCount, 0)
				}
			}
		})
	}
}

func TestImportService_ImportCSVStream(t *testing.T) {
	db := setupTestDB(t)
	logger := log.New(os.Stdout, "TEST: ", log.LstdFlags)
	service := services.NewImportService(db, logger)

	// First create an import session
	ctx := context.Background()
	session := &models.ImportSession{
		ID:           "test-session-123",
		FileName:     "test.csv",
		FileSize:     1024,
		AccountID:    "test-account",
		AccountIndex: 0,
		Status:       "in_progress",
		TotalRows:    2,
	}
	db.Create(session)

	tests := []struct {
		name    string
		params  *services.ImportCSVStreamParams
		wantErr bool
		errMsg  string
	}{
		{
			name: "successful stream import",
			params: &services.ImportCSVStreamParams{
				SessionID: "test-session-123",
				Chunks: []string{
					"使用日,使用時刻,入口IC,出口IC,通行料金,車両番号,ETCカード番号",
					"2025/01/15,09:30,東京IC,大阪IC,1000,品川123あ1234,1234567890123456",
				},
			},
			wantErr: false,
		},
		{
			name:    "nil params",
			params:  nil,
			wantErr: true,
			errMsg:  "params cannot be nil",
		},
		{
			name: "empty session ID",
			params: &services.ImportCSVStreamParams{
				SessionID: "",
				Chunks:    []string{"data"},
			},
			wantErr: true,
			errMsg:  "session ID is required",
		},
		{
			name: "session not found",
			params: &services.ImportCSVStreamParams{
				SessionID: "non-existent-session",
				Chunks:    []string{"data"},
			},
			wantErr: true,
			errMsg:  "not found",
		},
		{
			name: "empty chunks",
			params: &services.ImportCSVStreamParams{
				SessionID: "test-session-123",
				Chunks:    []string{},
			},
			wantErr: true,
			errMsg:  "chunks cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := service.ImportCSVStream(ctx, tt.params)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}
		})
	}
}

func TestImportService_GetImportSession(t *testing.T) {
	db := setupTestDB(t)
	logger := log.New(os.Stdout, "TEST: ", log.LstdFlags)
	service := services.NewImportService(db, logger)
	ctx := context.Background()

	// Create a test session
	session := &models.ImportSession{
		ID:           "test-session-456",
		FileName:     "test.csv",
		FileSize:     2048,
		AccountID:    "test-account",
		AccountIndex: 0,
		Status:       "completed",
		TotalRows:    10,
		ProcessedRows: 10,
		SuccessRows:  9,
		ErrorRows:    1,
	}
	db.Create(session)

	tests := []struct {
		name      string
		sessionID string
		wantErr   bool
		errMsg    string
	}{
		{
			name:      "successful retrieval",
			sessionID: "test-session-456",
			wantErr:   false,
		},
		{
			name:      "session not found",
			sessionID: "non-existent",
			wantErr:   true,
			errMsg:    "not found",
		},
		{
			name:      "empty session ID",
			sessionID: "",
			wantErr:   true,
			errMsg:    "session ID is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := service.GetImportSession(ctx, tt.sessionID)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				if result != nil {
					assert.Equal(t, tt.sessionID, result.ID)
				}
			}
		})
	}
}

func TestImportService_ListImportSessions(t *testing.T) {
	db := setupTestDB(t)
	logger := log.New(os.Stdout, "TEST: ", log.LstdFlags)
	service := services.NewImportService(db, logger)
	ctx := context.Background()

	// Create test sessions
	now := time.Now()
	sessions := []models.ImportSession{
		{
			ID:           "session-1",
			FileName:     "test1.csv",
			FileSize:     1024,
			AccountID:    "account-1",
			AccountIndex: 0,
			Status:       "completed",
			StartedAt:    now.Add(-2 * time.Hour),
		},
		{
			ID:           "session-2",
			FileName:     "test2.csv",
			FileSize:     2048,
			AccountID:    "account-2",
			AccountIndex: 1,
			Status:       "in_progress",
			StartedAt:    now.Add(-1 * time.Hour),
		},
		{
			ID:           "session-3",
			FileName:     "test3.csv",
			FileSize:     3072,
			AccountID:    "account-1",
			AccountIndex: 0,
			Status:       "failed",
			StartedAt:    now.Add(-30 * time.Minute),
		},
	}

	for _, s := range sessions {
		db.Create(&s)
	}

	tests := []struct {
		name      string
		params    *services.ListImportSessionsParams
		wantCount int
		wantErr   bool
	}{
		{
			name: "list all sessions",
			params: &services.ListImportSessionsParams{
				Page:     1,
				PageSize: 10,
			},
			wantCount: 3,
			wantErr:   false,
		},
		{
			name: "filter by status",
			params: &services.ListImportSessionsParams{
				Page:     1,
				PageSize: 10,
				Status:   stringPtr("completed"),
			},
			wantCount: 1,
			wantErr:   false,
		},
		{
			name: "filter by account ID",
			params: &services.ListImportSessionsParams{
				Page:      1,
				PageSize:  10,
				AccountID: stringPtr("account-1"),
			},
			wantCount: 2,
			wantErr:   false,
		},
		{
			name: "pagination",
			params: &services.ListImportSessionsParams{
				Page:     1,
				PageSize: 2,
			},
			wantCount: 2,
			wantErr:   false,
		},
		{
			name:      "nil params - use defaults",
			params:    nil,
			wantCount: 3,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := service.ListImportSessions(ctx, tt.params)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				if result != nil {
					assert.Len(t, result.Sessions, tt.wantCount)
				}
			}
		})
	}
}

func TestImportService_ProcessCSV(t *testing.T) {
	db := setupTestDB(t)
	logger := log.New(os.Stdout, "TEST: ", log.LstdFlags)
	service := services.NewImportService(db, logger)
	ctx := context.Background()

	rows := []*services.CSVRow{
		{
			Date:       "2025/01/15",
			Time:       "09:30",
			EntranceIC: "東京IC",
			ExitIC:     "大阪IC",
			TollAmount: "1000",
			CarNumber:  "品川123あ1234",
			ETCCardNumber: "1234567890123456",
		},
		{
			Date:       "2025/01/15",
			Time:       "10:30",
			EntranceIC: "名古屋IC",
			ExitIC:     "京都IC",
			TollAmount: "800",
			CarNumber:  "品川456い5678",
			ETCCardNumber: "9876543210123456",
		},
	}

	options := &services.BulkProcessOptions{
		BatchSize:      10,
		MaxConcurrency: 2,
		SkipErrors:     true,
	}

	tests := []struct {
		name    string
		rows    []*services.CSVRow
		options *services.BulkProcessOptions
		wantErr bool
		errMsg  string
	}{
		{
			name:    "successful processing",
			rows:    rows,
			options: options,
			wantErr: false,
		},
		{
			name:    "nil rows",
			rows:    nil,
			options: options,
			wantErr: true,
			errMsg:  "rows cannot be nil",
		},
		{
			name:    "empty rows",
			rows:    []*services.CSVRow{},
			options: options,
			wantErr: true,
			errMsg:  "no rows to process",
		},
		{
			name: "invalid row data",
			rows: []*services.CSVRow{
				{
					Date: "", // Invalid empty date
					Time: "09:30",
				},
			},
			options: options,
			wantErr: true,
			errMsg:  "invalid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := service.ProcessCSV(ctx, tt.rows, tt.options)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				if result != nil {
					assert.Greater(t, result.SuccessCount, 0)
				}
			}
		})
	}
}

func TestImportService_CancelImportSession(t *testing.T) {
	db := setupTestDB(t)
	logger := log.New(os.Stdout, "TEST: ", log.LstdFlags)
	service := services.NewImportService(db, logger)
	ctx := context.Background()

	// Create test sessions
	inProgressSession := &models.ImportSession{
		ID:     "in-progress-session",
		Status: "in_progress",
	}
	completedSession := &models.ImportSession{
		ID:     "completed-session",
		Status: "completed",
	}
	db.Create(inProgressSession)
	db.Create(completedSession)

	tests := []struct {
		name      string
		sessionID string
		wantErr   bool
		errMsg    string
	}{
		{
			name:      "successful cancellation",
			sessionID: "in-progress-session",
			wantErr:   false,
		},
		{
			name:      "already completed session",
			sessionID: "completed-session",
			wantErr:   true,
			errMsg:    "cannot cancel",
		},
		{
			name:      "session not found",
			sessionID: "non-existent",
			wantErr:   true,
			errMsg:    "not found",
		},
		{
			name:      "empty session ID",
			sessionID: "",
			wantErr:   true,
			errMsg:    "session ID is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.CancelImportSession(ctx, tt.sessionID)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
				// Verify the session was actually cancelled
				var session models.ImportSession
				db.First(&session, "id = ?", tt.sessionID)
				assert.Equal(t, "cancelled", session.Status)
			}
		})
	}
}

func TestImportService_HealthCheck(t *testing.T) {
	db := setupTestDB(t)
	logger := log.New(os.Stdout, "TEST: ", log.LstdFlags)
	service := services.NewImportService(db, logger)
	ctx := context.Background()

	// Test health check
	err := service.HealthCheck(ctx)
	assert.NoError(t, err)

	// Test with closed DB connection
	sqlDB, _ := db.DB()
	sqlDB.Close()
	err = service.HealthCheck(ctx)
	assert.Error(t, err)
}

// Helper function to create string pointer
func stringPtr(s string) *string {
	return &s
}
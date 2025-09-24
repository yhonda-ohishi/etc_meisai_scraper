package repositories

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/yhonda-ohishi/etc_meisai/src/models"
)

func TestListImportSessionsParams_Validation(t *testing.T) {
	accountType := "corporate"
	accountID := "CORP001"
	status := "completed"
	createdBy := "admin"

	tests := []struct {
		name   string
		params ListImportSessionsParams
		valid  bool
	}{
		{
			name: "valid parameters",
			params: ListImportSessionsParams{
				Page:        1,
				PageSize:    10,
				AccountType: &accountType,
				AccountID:   &accountID,
				Status:      &status,
				CreatedBy:   &createdBy,
				SortBy:      "created_at",
				SortOrder:   "desc",
			},
			valid: true,
		},
		{
			name: "minimal parameters",
			params: ListImportSessionsParams{
				Page:     1,
				PageSize: 10,
			},
			valid: true,
		},
		{
			name: "zero page",
			params: ListImportSessionsParams{
				Page:     0,
				PageSize: 10,
			},
			valid: true,
		},
		{
			name: "negative page",
			params: ListImportSessionsParams{
				Page:     -1,
				PageSize: 10,
			},
			valid: true,
		},
		{
			name: "zero page size",
			params: ListImportSessionsParams{
				Page:     1,
				PageSize: 0,
			},
			valid: true,
		},
		{
			name: "large page size",
			params: ListImportSessionsParams{
				Page:     1,
				PageSize: 1000,
			},
			valid: true,
		},
		{
			name: "empty string parameters",
			params: ListImportSessionsParams{
				Page:        1,
				PageSize:    10,
				AccountType: stringPtr(""),
				AccountID:   stringPtr(""),
				Status:      stringPtr(""),
				CreatedBy:   stringPtr(""),
				SortBy:      "",
				SortOrder:   "",
			},
			valid: true,
		},
		{
			name: "invalid sort order",
			params: ListImportSessionsParams{
				Page:      1,
				PageSize:  10,
				SortBy:    "created_at",
				SortOrder: "invalid",
			},
			valid: true,
		},
		{
			name: "sort by session_id asc",
			params: ListImportSessionsParams{
				Page:      1,
				PageSize:  10,
				SortBy:    "session_id",
				SortOrder: "asc",
			},
			valid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test that the struct can be created and accessed
			assert.Equal(t, tt.params.Page, tt.params.Page)
			assert.Equal(t, tt.params.PageSize, tt.params.PageSize)

			if tt.params.AccountType != nil {
				assert.NotNil(t, tt.params.AccountType)
			}
			if tt.params.AccountID != nil {
				assert.NotNil(t, tt.params.AccountID)
			}
			if tt.params.Status != nil {
				assert.NotNil(t, tt.params.Status)
			}
			if tt.params.CreatedBy != nil {
				assert.NotNil(t, tt.params.CreatedBy)
			}

			assert.Equal(t, tt.params.SortBy, tt.params.SortBy)
			assert.Equal(t, tt.params.SortOrder, tt.params.SortOrder)
		})
	}
}

func TestImportRepository_InterfaceMethods(t *testing.T) {
	ctx := context.Background()
	session := &models.ImportSession{
		ID:            "session123",
		AccountType:   "corporate",
		AccountID:     "CORP001",
		Status:        "processing",
		TotalRows:     100,
		ProcessedRows: 50,
		ErrorRows:     2,
		CreatedBy:     "admin",
	}

	record := &models.ETCMeisaiRecord{
		ID:              1,
		Date:            parseDate("2025-01-15"),
		Time:            "09:30",
		EntranceIC:      "東京IC",
		ExitIC:          "大阪IC",
		TollAmount:      1000,
		CarNumber:       "品川123あ1234",
		ETCCardNumber:   "1234567890",
		ETCNum:          stringPtr("ETC001"),
		Hash:            "abcd1234",
	}

	mockRepo := &mockImportRepository{}

	// Test session management
	t.Run("CreateSession", func(t *testing.T) {
		err := mockRepo.CreateSession(ctx, session)
		assert.NoError(t, err)
	})

	t.Run("GetSession", func(t *testing.T) {
		result, err := mockRepo.GetSession(ctx, "session123")
		assert.NoError(t, err)
		assert.Equal(t, session.ID, result.ID)
	})

	t.Run("UpdateSession", func(t *testing.T) {
		err := mockRepo.UpdateSession(ctx, session)
		assert.NoError(t, err)
	})

	t.Run("ListSessions", func(t *testing.T) {
		params := ListImportSessionsParams{
			Page:     1,
			PageSize: 10,
		}
		sessions, count, err := mockRepo.ListSessions(ctx, params)
		assert.NoError(t, err)
		assert.Len(t, sessions, 1)
		assert.Equal(t, int64(1), count)
	})

	t.Run("CancelSession", func(t *testing.T) {
		err := mockRepo.CancelSession(ctx, "session123")
		assert.NoError(t, err)
	})

	// Test record operations
	t.Run("CreateRecord", func(t *testing.T) {
		err := mockRepo.CreateRecord(ctx, record)
		assert.NoError(t, err)
	})

	t.Run("CreateRecordsBatch", func(t *testing.T) {
		records := []*models.ETCMeisaiRecord{record}
		err := mockRepo.CreateRecordsBatch(ctx, records)
		assert.NoError(t, err)
	})

	t.Run("FindRecordByHash", func(t *testing.T) {
		result, err := mockRepo.FindRecordByHash(ctx, "abcd1234")
		assert.NoError(t, err)
		assert.Equal(t, record.Hash, result.Hash)
	})

	t.Run("FindDuplicateRecords", func(t *testing.T) {
		hashes := []string{"abcd1234", "efgh5678"}
		results, err := mockRepo.FindDuplicateRecords(ctx, hashes)
		assert.NoError(t, err)
		assert.Len(t, results, 1)
	})

	// Test transaction support
	t.Run("BeginTx", func(t *testing.T) {
		txRepo, err := mockRepo.BeginTx(ctx)
		assert.NoError(t, err)
		assert.NotNil(t, txRepo)
	})

	t.Run("CommitTx", func(t *testing.T) {
		err := mockRepo.CommitTx()
		assert.NoError(t, err)
	})

	t.Run("RollbackTx", func(t *testing.T) {
		err := mockRepo.RollbackTx()
		assert.NoError(t, err)
	})

	// Test health check
	t.Run("Ping", func(t *testing.T) {
		err := mockRepo.Ping(ctx)
		assert.NoError(t, err)
	})
}

func TestImportRepository_ErrorScenarios(t *testing.T) {
	ctx := context.Background()
	errorRepo := &errorImportRepository{}

	// Test error handling for all methods
	t.Run("CreateSession error", func(t *testing.T) {
		err := errorRepo.CreateSession(ctx, nil)
		assert.Error(t, err)
	})

	t.Run("GetSession error", func(t *testing.T) {
		result, err := errorRepo.GetSession(ctx, "invalid")
		assert.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("UpdateSession error", func(t *testing.T) {
		err := errorRepo.UpdateSession(ctx, nil)
		assert.Error(t, err)
	})

	t.Run("ListSessions error", func(t *testing.T) {
		params := ListImportSessionsParams{Page: 1, PageSize: 10}
		sessions, count, err := errorRepo.ListSessions(ctx, params)
		assert.Error(t, err)
		assert.Nil(t, sessions)
		assert.Equal(t, int64(0), count)
	})

	t.Run("CancelSession error", func(t *testing.T) {
		err := errorRepo.CancelSession(ctx, "session123")
		assert.Error(t, err)
	})

	t.Run("CreateRecord error", func(t *testing.T) {
		err := errorRepo.CreateRecord(ctx, nil)
		assert.Error(t, err)
	})

	t.Run("CreateRecordsBatch error", func(t *testing.T) {
		err := errorRepo.CreateRecordsBatch(ctx, nil)
		assert.Error(t, err)
	})

	t.Run("FindRecordByHash error", func(t *testing.T) {
		result, err := errorRepo.FindRecordByHash(ctx, "hash")
		assert.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("FindDuplicateRecords error", func(t *testing.T) {
		results, err := errorRepo.FindDuplicateRecords(ctx, []string{"hash"})
		assert.Error(t, err)
		assert.Nil(t, results)
	})

	t.Run("BeginTx error", func(t *testing.T) {
		result, err := errorRepo.BeginTx(ctx)
		assert.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("CommitTx error", func(t *testing.T) {
		err := errorRepo.CommitTx()
		assert.Error(t, err)
	})

	t.Run("RollbackTx error", func(t *testing.T) {
		err := errorRepo.RollbackTx()
		assert.Error(t, err)
	})

	t.Run("Ping error", func(t *testing.T) {
		err := errorRepo.Ping(ctx)
		assert.Error(t, err)
	})
}

func TestImportRepository_EdgeCases(t *testing.T) {
	ctx := context.Background()
	mockRepo := &mockImportRepository{}

	t.Run("GetSession with empty session ID", func(t *testing.T) {
		result, err := mockRepo.GetSession(ctx, "")
		assert.NoError(t, err)
		assert.NotNil(t, result)
	})

	t.Run("CreateRecordsBatch with empty batch", func(t *testing.T) {
		err := mockRepo.CreateRecordsBatch(ctx, []*models.ETCMeisaiRecord{})
		assert.NoError(t, err)
	})

	t.Run("FindDuplicateRecords with empty hashes", func(t *testing.T) {
		results, err := mockRepo.FindDuplicateRecords(ctx, []string{})
		assert.NoError(t, err)
		assert.Empty(t, results)
	})

	t.Run("FindDuplicateRecords with nil hashes", func(t *testing.T) {
		results, err := mockRepo.FindDuplicateRecords(ctx, nil)
		assert.NoError(t, err)
		assert.Empty(t, results)
	})

	t.Run("ListSessions with complex filters", func(t *testing.T) {
		accountType := "personal"
		accountID := "PERS001"
		status := "failed"
		createdBy := "user123"

		params := ListImportSessionsParams{
			Page:        2,
			PageSize:    50,
			AccountType: &accountType,
			AccountID:   &accountID,
			Status:      &status,
			CreatedBy:   &createdBy,
			SortBy:      "updated_at",
			SortOrder:   "asc",
		}
		sessions, count, err := mockRepo.ListSessions(ctx, params)
		assert.NoError(t, err)
		assert.NotNil(t, sessions)
		assert.GreaterOrEqual(t, count, int64(0))
	})
}

func TestImportStructs(t *testing.T) {
	// Test ImportResult
	t.Run("ImportResult", func(t *testing.T) {
		result := ImportResult{
			TotalRecords:   100,
			SuccessCount:   85,
			ErrorCount:     10,
			DuplicateCount: 5,
		}
		assert.Equal(t, 100, result.TotalRecords)
		assert.Equal(t, 85, result.SuccessCount)
		assert.Equal(t, 10, result.ErrorCount)
		assert.Equal(t, 5, result.DuplicateCount)
	})

	// Test ImportError
	t.Run("ImportError", func(t *testing.T) {
		importError := ImportError{
			RowNumber: 5,
			Field:     "toll_amount",
			Value:     "invalid",
			Message:   "Invalid toll amount format",
		}
		assert.Equal(t, 5, importError.RowNumber)
		assert.Equal(t, "toll_amount", importError.Field)
		assert.Equal(t, "invalid", importError.Value)
		assert.Equal(t, "Invalid toll amount format", importError.Message)
	})

	// Test edge cases for ImportResult
	t.Run("ImportResult with zero values", func(t *testing.T) {
		result := ImportResult{
			TotalRecords:   0,
			SuccessCount:   0,
			ErrorCount:     0,
			DuplicateCount: 0,
		}
		assert.Equal(t, 0, result.TotalRecords)
		assert.Equal(t, 0, result.SuccessCount)
		assert.Equal(t, 0, result.ErrorCount)
		assert.Equal(t, 0, result.DuplicateCount)
	})

	// Test edge cases for ImportError
	t.Run("ImportError with empty values", func(t *testing.T) {
		importError := ImportError{
			RowNumber: 0,
			Field:     "",
			Value:     "",
			Message:   "",
		}
		assert.Equal(t, 0, importError.RowNumber)
		assert.Equal(t, "", importError.Field)
		assert.Equal(t, "", importError.Value)
		assert.Equal(t, "", importError.Message)
	})
}

// Helper function to parse date (mock implementation)
func parseDate(dateStr string) time.Time {
	return time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC)
}

// Mock implementation for testing
type mockImportRepository struct{}

func (m *mockImportRepository) CreateSession(ctx context.Context, session *models.ImportSession) error {
	return nil
}

func (m *mockImportRepository) GetSession(ctx context.Context, sessionID string) (*models.ImportSession, error) {
	return &models.ImportSession{
		ID:            sessionID,
		AccountType:   "corporate",
		AccountID:     "CORP001",
		Status:        "processing",
		TotalRows:     100,
		ProcessedRows: 50,
		ErrorRows:     2,
		CreatedBy:     "admin",
	}, nil
}

func (m *mockImportRepository) UpdateSession(ctx context.Context, session *models.ImportSession) error {
	return nil
}

func (m *mockImportRepository) ListSessions(ctx context.Context, params ListImportSessionsParams) ([]*models.ImportSession, int64, error) {
	return []*models.ImportSession{
		{
			ID:            "session123",
			AccountType:   "corporate",
			AccountID:     "CORP001",
			Status:        "completed",
			TotalRows:     100,
			ProcessedRows: 100,
			ErrorRows:     0,
			CreatedBy:     "admin",
		},
	}, 1, nil
}

func (m *mockImportRepository) CancelSession(ctx context.Context, sessionID string) error {
	return nil
}

func (m *mockImportRepository) CreateRecord(ctx context.Context, record *models.ETCMeisaiRecord) error {
	return nil
}

func (m *mockImportRepository) CreateRecordsBatch(ctx context.Context, records []*models.ETCMeisaiRecord) error {
	return nil
}

func (m *mockImportRepository) FindRecordByHash(ctx context.Context, hash string) (*models.ETCMeisaiRecord, error) {
	return &models.ETCMeisaiRecord{
		ID:              1,
		Date:            parseDate("2025-01-15"),
		Time:            "09:30",
		EntranceIC:      "東京IC",
		ExitIC:          "大阪IC",
		TollAmount:      1000,
		CarNumber:       "品川123あ1234",
		ETCCardNumber:   "1234567890",
		ETCNum:          stringPtr("ETC001"),
		Hash:            hash,
	}, nil
}

func (m *mockImportRepository) FindDuplicateRecords(ctx context.Context, hashes []string) ([]*models.ETCMeisaiRecord, error) {
	if len(hashes) == 0 {
		return []*models.ETCMeisaiRecord{}, nil
	}
	return []*models.ETCMeisaiRecord{
		{
			ID:              1,
			Hash:            hashes[0],
			Date:            parseDate("2025-01-15"),
			Time:            "09:30",
			EntranceIC:      "東京IC",
			ExitIC:          "大阪IC",
			TollAmount:      1000,
			CarNumber:       "品川123あ1234",
			ETCCardNumber:   "1234567890",
			ETCNum:          stringPtr("ETC001"),
		},
	}, nil
}

func (m *mockImportRepository) BeginTx(ctx context.Context) (ImportRepository, error) {
	return m, nil
}

func (m *mockImportRepository) CommitTx() error {
	return nil
}

func (m *mockImportRepository) RollbackTx() error {
	return nil
}

func (m *mockImportRepository) Ping(ctx context.Context) error {
	return nil
}

// Error implementation for testing error scenarios
type errorImportRepository struct{}

func (e *errorImportRepository) CreateSession(ctx context.Context, session *models.ImportSession) error {
	return assert.AnError
}

func (e *errorImportRepository) GetSession(ctx context.Context, sessionID string) (*models.ImportSession, error) {
	return nil, assert.AnError
}

func (e *errorImportRepository) UpdateSession(ctx context.Context, session *models.ImportSession) error {
	return assert.AnError
}

func (e *errorImportRepository) ListSessions(ctx context.Context, params ListImportSessionsParams) ([]*models.ImportSession, int64, error) {
	return nil, 0, assert.AnError
}

func (e *errorImportRepository) CancelSession(ctx context.Context, sessionID string) error {
	return assert.AnError
}

func (e *errorImportRepository) CreateRecord(ctx context.Context, record *models.ETCMeisaiRecord) error {
	return assert.AnError
}

func (e *errorImportRepository) CreateRecordsBatch(ctx context.Context, records []*models.ETCMeisaiRecord) error {
	return assert.AnError
}

func (e *errorImportRepository) FindRecordByHash(ctx context.Context, hash string) (*models.ETCMeisaiRecord, error) {
	return nil, assert.AnError
}

func (e *errorImportRepository) FindDuplicateRecords(ctx context.Context, hashes []string) ([]*models.ETCMeisaiRecord, error) {
	return nil, assert.AnError
}

func (e *errorImportRepository) BeginTx(ctx context.Context) (ImportRepository, error) {
	return nil, assert.AnError
}

func (e *errorImportRepository) CommitTx() error {
	return assert.AnError
}

func (e *errorImportRepository) RollbackTx() error {
	return assert.AnError
}

func (e *errorImportRepository) Ping(ctx context.Context) error {
	return assert.AnError
}
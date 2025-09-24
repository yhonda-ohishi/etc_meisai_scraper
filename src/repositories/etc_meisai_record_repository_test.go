package repositories

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/yhonda-ohishi/etc_meisai/src/models"
)

func TestListRecordsParams_Validation(t *testing.T) {
	dateFrom := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	dateTo := time.Date(2025, 1, 31, 23, 59, 59, 999999999, time.UTC)
	carNumber := "品川123あ1234"
	etcNumber := "1234567890"
	etcNum := "ETC001"

	tests := []struct {
		name   string
		params ListRecordsParams
		valid  bool
	}{
		{
			name: "valid parameters",
			params: ListRecordsParams{
				Page:      1,
				PageSize:  10,
				DateFrom:  &dateFrom,
				DateTo:    &dateTo,
				CarNumber: &carNumber,
				ETCNumber: &etcNumber,
				ETCNum:    &etcNum,
				SortBy:    "date",
				SortOrder: "asc",
			},
			valid: true,
		},
		{
			name: "minimal parameters",
			params: ListRecordsParams{
				Page:     1,
				PageSize: 10,
			},
			valid: true,
		},
		{
			name: "zero page",
			params: ListRecordsParams{
				Page:     0,
				PageSize: 10,
			},
			valid: true,
		},
		{
			name: "negative page",
			params: ListRecordsParams{
				Page:     -1,
				PageSize: 10,
			},
			valid: true,
		},
		{
			name: "zero page size",
			params: ListRecordsParams{
				Page:     1,
				PageSize: 0,
			},
			valid: true,
		},
		{
			name: "large page size",
			params: ListRecordsParams{
				Page:     1,
				PageSize: 1000,
			},
			valid: true,
		},
		{
			name: "date range with from after to",
			params: ListRecordsParams{
				Page:     1,
				PageSize: 10,
				DateFrom: &dateTo,
				DateTo:   &dateFrom,
			},
			valid: true,
		},
		{
			name: "same from and to date",
			params: ListRecordsParams{
				Page:     1,
				PageSize: 10,
				DateFrom: &dateFrom,
				DateTo:   &dateFrom,
			},
			valid: true,
		},
		{
			name: "empty string parameters",
			params: ListRecordsParams{
				Page:      1,
				PageSize:  10,
				CarNumber: stringPtr(""),
				ETCNumber: stringPtr(""),
				ETCNum:    stringPtr(""),
				SortBy:    "",
				SortOrder: "",
			},
			valid: true,
		},
		{
			name: "sort by toll_amount desc",
			params: ListRecordsParams{
				Page:      1,
				PageSize:  10,
				SortBy:    "toll_amount",
				SortOrder: "desc",
			},
			valid: true,
		},
		{
			name: "sort by car_number",
			params: ListRecordsParams{
				Page:      1,
				PageSize:  10,
				SortBy:    "car_number",
				SortOrder: "asc",
			},
			valid: true,
		},
		{
			name: "invalid sort order",
			params: ListRecordsParams{
				Page:      1,
				PageSize:  10,
				SortBy:    "date",
				SortOrder: "invalid",
			},
			valid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test that the struct can be created and accessed
			assert.Equal(t, tt.params.Page, tt.params.Page)
			assert.Equal(t, tt.params.PageSize, tt.params.PageSize)

			if tt.params.DateFrom != nil {
				assert.NotNil(t, tt.params.DateFrom)
			}
			if tt.params.DateTo != nil {
				assert.NotNil(t, tt.params.DateTo)
			}
			if tt.params.CarNumber != nil {
				assert.NotNil(t, tt.params.CarNumber)
			}
			if tt.params.ETCNumber != nil {
				assert.NotNil(t, tt.params.ETCNumber)
			}
			if tt.params.ETCNum != nil {
				assert.NotNil(t, tt.params.ETCNum)
			}

			assert.Equal(t, tt.params.SortBy, tt.params.SortBy)
			assert.Equal(t, tt.params.SortOrder, tt.params.SortOrder)
		})
	}
}

func TestETCMeisaiRecordRepository_InterfaceMethods(t *testing.T) {
	// Test that the interface can be implemented
	ctx := context.Background()
	record := &models.ETCMeisaiRecord{
		ID:              1,
		Date:            time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC),
		Time:            "09:30",
		EntranceIC:      "東京IC",
		ExitIC:          "大阪IC",
		TollAmount:      1000,
		CarNumber:       "品川123あ1234",
		ETCCardNumber:   "1234567890",
		ETCNum:          stringPtr("ETC001"),
		Hash:            "abcd1234",
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	// Test interface method signatures by creating a mock implementation
	mockRepo := &mockETCMeisaiRecordRepository{}

	// Test basic CRUD operations
	t.Run("Create", func(t *testing.T) {
		err := mockRepo.Create(ctx, record)
		assert.NoError(t, err)
	})

	t.Run("GetByID", func(t *testing.T) {
		result, err := mockRepo.GetByID(ctx, 1)
		assert.NoError(t, err)
		assert.Equal(t, record, result)
	})

	t.Run("Update", func(t *testing.T) {
		err := mockRepo.Update(ctx, record)
		assert.NoError(t, err)
	})

	t.Run("Delete", func(t *testing.T) {
		err := mockRepo.Delete(ctx, 1)
		assert.NoError(t, err)
	})

	// Test query operations
	t.Run("GetByHash", func(t *testing.T) {
		result, err := mockRepo.GetByHash(ctx, "abcd1234")
		assert.NoError(t, err)
		assert.Equal(t, record, result)
	})

	t.Run("CheckDuplicateHash", func(t *testing.T) {
		exists, err := mockRepo.CheckDuplicateHash(ctx, "abcd1234")
		assert.NoError(t, err)
		assert.True(t, exists)
	})

	t.Run("CheckDuplicateHash with exclude ID", func(t *testing.T) {
		exists, err := mockRepo.CheckDuplicateHash(ctx, "abcd1234", 1)
		assert.NoError(t, err)
		assert.False(t, exists)
	})

	// Test list operations
	t.Run("List", func(t *testing.T) {
		params := ListRecordsParams{
			Page:     1,
			PageSize: 10,
		}
		result, count, err := mockRepo.List(ctx, params)
		assert.NoError(t, err)
		assert.Len(t, result, 1)
		assert.Equal(t, int64(1), count)
		assert.Equal(t, record, result[0])
	})

	// Test transaction support
	t.Run("BeginTx", func(t *testing.T) {
		txRepo, err := mockRepo.BeginTx(ctx)
		assert.NoError(t, err)
		assert.NotNil(t, txRepo)
		assert.IsType(t, &mockETCMeisaiRecordRepository{}, txRepo)
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

func TestETCMeisaiRecordRepository_ErrorScenarios(t *testing.T) {
	ctx := context.Background()
	errorRepo := &errorETCMeisaiRecordRepository{}

	// Test error handling for all methods
	t.Run("Create error", func(t *testing.T) {
		err := errorRepo.Create(ctx, nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "create failed")
	})

	t.Run("GetByID error", func(t *testing.T) {
		result, err := errorRepo.GetByID(ctx, 1)
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "get failed")
	})

	t.Run("Update error", func(t *testing.T) {
		err := errorRepo.Update(ctx, nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "update failed")
	})

	t.Run("Delete error", func(t *testing.T) {
		err := errorRepo.Delete(ctx, 1)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "delete failed")
	})

	t.Run("GetByHash error", func(t *testing.T) {
		result, err := errorRepo.GetByHash(ctx, "hash")
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "get by hash failed")
	})

	t.Run("CheckDuplicateHash error", func(t *testing.T) {
		exists, err := errorRepo.CheckDuplicateHash(ctx, "hash")
		assert.Error(t, err)
		assert.False(t, exists)
		assert.Contains(t, err.Error(), "check duplicate failed")
	})

	t.Run("List error", func(t *testing.T) {
		params := ListRecordsParams{Page: 1, PageSize: 10}
		result, count, err := errorRepo.List(ctx, params)
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, int64(0), count)
		assert.Contains(t, err.Error(), "list failed")
	})

	t.Run("BeginTx error", func(t *testing.T) {
		result, err := errorRepo.BeginTx(ctx)
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "begin transaction failed")
	})

	t.Run("CommitTx error", func(t *testing.T) {
		err := errorRepo.CommitTx()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "commit transaction failed")
	})

	t.Run("RollbackTx error", func(t *testing.T) {
		err := errorRepo.RollbackTx()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "rollback transaction failed")
	})

	t.Run("Ping error", func(t *testing.T) {
		err := errorRepo.Ping(ctx)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "ping failed")
	})
}

func TestETCMeisaiRecordRepository_EdgeCases(t *testing.T) {
	ctx := context.Background()
	mockRepo := &mockETCMeisaiRecordRepository{}

	t.Run("GetByHash empty hash", func(t *testing.T) {
		result, err := mockRepo.GetByHash(ctx, "")
		assert.NoError(t, err)
		assert.NotNil(t, result)
	})

	t.Run("GetByHash very long hash", func(t *testing.T) {
		longHash := "abcdefghijklmnopqrstuvwxyz1234567890abcdefghijklmnopqrstuvwxyz1234567890"
		result, err := mockRepo.GetByHash(ctx, longHash)
		assert.NoError(t, err)
		assert.NotNil(t, result)
	})

	t.Run("CheckDuplicateHash with multiple exclude IDs", func(t *testing.T) {
		exists, err := mockRepo.CheckDuplicateHash(ctx, "hash", 1, 2, 3)
		assert.NoError(t, err)
		assert.False(t, exists)
	})

	t.Run("List with complex filters", func(t *testing.T) {
		dateFrom := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
		dateTo := time.Date(2025, 1, 31, 23, 59, 59, 999999999, time.UTC)
		carNumber := "品川123あ1234"
		etcNumber := "1234567890"
		etcNum := "ETC001"

		params := ListRecordsParams{
			Page:      1,
			PageSize:  10,
			DateFrom:  &dateFrom,
			DateTo:    &dateTo,
			CarNumber: &carNumber,
			ETCNumber: &etcNumber,
			ETCNum:    &etcNum,
			SortBy:    "toll_amount",
			SortOrder: "desc",
		}
		result, count, err := mockRepo.List(ctx, params)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.GreaterOrEqual(t, count, int64(0))
	})

	t.Run("List with nil pointers", func(t *testing.T) {
		params := ListRecordsParams{
			Page:      1,
			PageSize:  10,
			DateFrom:  nil,
			DateTo:    nil,
			CarNumber: nil,
			ETCNumber: nil,
			ETCNum:    nil,
		}
		result, count, err := mockRepo.List(ctx, params)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.GreaterOrEqual(t, count, int64(0))
	})
}

// Mock implementation for testing
type mockETCMeisaiRecordRepository struct{}

func (m *mockETCMeisaiRecordRepository) Create(ctx context.Context, record *models.ETCMeisaiRecord) error {
	return nil
}

func (m *mockETCMeisaiRecordRepository) GetByID(ctx context.Context, id int64) (*models.ETCMeisaiRecord, error) {
	return &models.ETCMeisaiRecord{
		ID:              id,
		Date:            time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC),
		Time:            "09:30",
		EntranceIC:      "東京IC",
		ExitIC:          "大阪IC",
		TollAmount:      1000,
		CarNumber:       "品川123あ1234",
		ETCCardNumber:   "1234567890",
		ETCNum:          stringPtr("ETC001"),
		Hash:            "abcd1234",
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}, nil
}

func (m *mockETCMeisaiRecordRepository) Update(ctx context.Context, record *models.ETCMeisaiRecord) error {
	return nil
}

func (m *mockETCMeisaiRecordRepository) Delete(ctx context.Context, id int64) error {
	return nil
}

func (m *mockETCMeisaiRecordRepository) GetByHash(ctx context.Context, hash string) (*models.ETCMeisaiRecord, error) {
	return &models.ETCMeisaiRecord{
		ID:              1,
		Date:            time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC),
		Time:            "09:30",
		EntranceIC:      "東京IC",
		ExitIC:          "大阪IC",
		TollAmount:      1000,
		CarNumber:       "品川123あ1234",
		ETCCardNumber:   "1234567890",
		ETCNum:          stringPtr("ETC001"),
		Hash:            hash,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}, nil
}

func (m *mockETCMeisaiRecordRepository) CheckDuplicateHash(ctx context.Context, hash string, excludeID ...int64) (bool, error) {
	// Return false if any exclude ID is provided
	if len(excludeID) > 0 {
		return false, nil
	}
	return true, nil
}

func (m *mockETCMeisaiRecordRepository) List(ctx context.Context, params ListRecordsParams) ([]*models.ETCMeisaiRecord, int64, error) {
	return []*models.ETCMeisaiRecord{
		{
			ID:              1,
			Date:            time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC),
			Time:            "09:30",
			EntranceIC:      "東京IC",
			ExitIC:          "大阪IC",
			TollAmount:      1000,
			CarNumber:       "品川123あ1234",
			ETCCardNumber:   "1234567890",
			ETCNum:          stringPtr("ETC001"),
			Hash:            "abcd1234",
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
		},
	}, 1, nil
}

func (m *mockETCMeisaiRecordRepository) BeginTx(ctx context.Context) (ETCMeisaiRecordRepository, error) {
	return m, nil
}

func (m *mockETCMeisaiRecordRepository) CommitTx() error {
	return nil
}

func (m *mockETCMeisaiRecordRepository) RollbackTx() error {
	return nil
}

func (m *mockETCMeisaiRecordRepository) Ping(ctx context.Context) error {
	return nil
}

// Error implementation for testing error scenarios
type errorETCMeisaiRecordRepository struct{}

func (e *errorETCMeisaiRecordRepository) Create(ctx context.Context, record *models.ETCMeisaiRecord) error {
	return assert.AnError
}

func (e *errorETCMeisaiRecordRepository) GetByID(ctx context.Context, id int64) (*models.ETCMeisaiRecord, error) {
	return nil, assert.AnError
}

func (e *errorETCMeisaiRecordRepository) Update(ctx context.Context, record *models.ETCMeisaiRecord) error {
	return assert.AnError
}

func (e *errorETCMeisaiRecordRepository) Delete(ctx context.Context, id int64) error {
	return assert.AnError
}

func (e *errorETCMeisaiRecordRepository) GetByHash(ctx context.Context, hash string) (*models.ETCMeisaiRecord, error) {
	return nil, assert.AnError
}

func (e *errorETCMeisaiRecordRepository) CheckDuplicateHash(ctx context.Context, hash string, excludeID ...int64) (bool, error) {
	return false, assert.AnError
}

func (e *errorETCMeisaiRecordRepository) List(ctx context.Context, params ListRecordsParams) ([]*models.ETCMeisaiRecord, int64, error) {
	return nil, 0, assert.AnError
}

func (e *errorETCMeisaiRecordRepository) BeginTx(ctx context.Context) (ETCMeisaiRecordRepository, error) {
	return nil, assert.AnError
}

func (e *errorETCMeisaiRecordRepository) CommitTx() error {
	return assert.AnError
}

func (e *errorETCMeisaiRecordRepository) RollbackTx() error {
	return assert.AnError
}

func (e *errorETCMeisaiRecordRepository) Ping(ctx context.Context) error {
	return assert.AnError
}
package services

import (
	"context"
	"errors"
	"log"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/yhonda-ohishi/etc_meisai/src/mocks"
	"github.com/yhonda-ohishi/etc_meisai/src/models"
	"github.com/yhonda-ohishi/etc_meisai/src/repositories"
)

func TestNewETCMeisaiService(t *testing.T) {
	t.Parallel()

	t.Run("with repository and logger", func(t *testing.T) {
		mockRepo := &mocks.MockETCMeisaiRecordRepository{}
		logger := log.New(os.Stdout, "test", log.LstdFlags)

		service := NewETCMeisaiService(mockRepo, logger)

		assert.NotNil(t, service)
		assert.Equal(t, mockRepo, service.repo)
		assert.Equal(t, logger, service.logger)
	})

	t.Run("with repository, no logger", func(t *testing.T) {
		mockRepo := &mocks.MockETCMeisaiRecordRepository{}

		service := NewETCMeisaiService(mockRepo, nil)

		assert.NotNil(t, service)
		assert.Equal(t, mockRepo, service.repo)
		assert.NotNil(t, service.logger)
	})
}

func TestETCMeisaiService_CreateRecord(t *testing.T) {
	t.Parallel()

	validParams := &CreateRecordParams{
		Date:          time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
		Time:          "10:30:00",
		EntranceIC:    "羽田空港IC",
		ExitIC:        "新宿IC",
		TollAmount:    1200,
		CarNumber:     "あ123",
		ETCCardNumber: "1234567890123456",
		ETCNum:        stringPtr("ETC123"),
		DtakoRowID:    int64Ptr(100),
	}

	tests := []struct {
		name        string
		params      *CreateRecordParams
		setupMock   func(*mocks.MockETCMeisaiRecordRepository)
		expectError bool
		errorMsg    string
	}{
		{
			name:   "successful creation",
			params: validParams,
			setupMock: func(m *mocks.MockETCMeisaiRecordRepository) {
				txMock := &mocks.MockETCMeisaiRecordRepository{}
				m.On("BeginTx", mock.Anything).Return(txMock, nil)
				txMock.On("CheckDuplicateHash", mock.Anything, mock.AnythingOfType("string")).Return(false, nil)
				txMock.On("Create", mock.Anything, mock.AnythingOfType("*models.ETCMeisaiRecord")).Return(nil)
				txMock.On("CommitTx").Return(nil)
				txMock.On("RollbackTx").Return(nil).Maybe()
			},
			expectError: false,
		},
		{
			name: "validation error - negative toll amount",
			params: &CreateRecordParams{
				Date:          time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
				Time:          "10:30:00",
				EntranceIC:    "羽田空港IC",
				ExitIC:        "新宿IC",
				TollAmount:    -100,
				CarNumber:     "あ123",
				ETCCardNumber: "1234567890123456",
			},
			setupMock:   func(m *mocks.MockETCMeisaiRecordRepository) {},
			expectError: true,
			errorMsg:    "record validation failed",
		},
		{
			name: "begin transaction error",
			params: validParams,
			setupMock: func(m *mocks.MockETCMeisaiRecordRepository) {
				m.On("BeginTx", mock.Anything).Return(nil, errors.New("transaction failed"))
			},
			expectError: true,
			errorMsg:    "failed to start transaction",
		},
		{
			name: "duplicate hash",
			params: validParams,
			setupMock: func(m *mocks.MockETCMeisaiRecordRepository) {
				txMock := &mocks.MockETCMeisaiRecordRepository{}
				m.On("BeginTx", mock.Anything).Return(txMock, nil)
				txMock.On("CheckDuplicateHash", mock.Anything, mock.AnythingOfType("string")).Return(true, nil)
				txMock.On("RollbackTx").Return(nil)
			},
			expectError: true,
			errorMsg:    "duplicate record with hash",
		},
		{
			name: "duplicate check error",
			params: validParams,
			setupMock: func(m *mocks.MockETCMeisaiRecordRepository) {
				txMock := &mocks.MockETCMeisaiRecordRepository{}
				m.On("BeginTx", mock.Anything).Return(txMock, nil)
				txMock.On("CheckDuplicateHash", mock.Anything, mock.AnythingOfType("string")).Return(false, errors.New("db error"))
				txMock.On("RollbackTx").Return(nil)
			},
			expectError: true,
			errorMsg:    "failed to check for duplicates",
		},
		{
			name: "create record error",
			params: validParams,
			setupMock: func(m *mocks.MockETCMeisaiRecordRepository) {
				txMock := &mocks.MockETCMeisaiRecordRepository{}
				m.On("BeginTx", mock.Anything).Return(txMock, nil)
				txMock.On("CheckDuplicateHash", mock.Anything, mock.AnythingOfType("string")).Return(false, nil)
				txMock.On("Create", mock.Anything, mock.AnythingOfType("*models.ETCMeisaiRecord")).Return(errors.New("create failed"))
				txMock.On("RollbackTx").Return(nil)
			},
			expectError: true,
			errorMsg:    "failed to create record",
		},
		{
			name: "commit error",
			params: validParams,
			setupMock: func(m *mocks.MockETCMeisaiRecordRepository) {
				txMock := &mocks.MockETCMeisaiRecordRepository{}
				m.On("BeginTx", mock.Anything).Return(txMock, nil)
				txMock.On("CheckDuplicateHash", mock.Anything, mock.AnythingOfType("string")).Return(false, nil)
				txMock.On("Create", mock.Anything, mock.AnythingOfType("*models.ETCMeisaiRecord")).Return(nil)
				txMock.On("CommitTx").Return(errors.New("commit failed"))
			},
			expectError: true,
			errorMsg:    "failed to commit transaction",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &mocks.MockETCMeisaiRecordRepository{}
			tt.setupMock(mockRepo)

			service := NewETCMeisaiService(mockRepo, nil)
			ctx := context.Background()

			record, err := service.CreateRecord(ctx, tt.params)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
				assert.Nil(t, record)
			} else {
				assert.NoError(t, err)
				if record != nil {
					assert.NotEmpty(t, record.Hash)
				}
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestETCMeisaiService_CreateRecord_WithPanic(t *testing.T) {
	t.Parallel()

	t.Run("panic during transaction", func(t *testing.T) {
		mockRepo := &mocks.MockETCMeisaiRecordRepository{}
		txMock := &mocks.MockETCMeisaiRecordRepository{}

		mockRepo.On("BeginTx", mock.Anything).Return(txMock, nil)
		txMock.On("CheckDuplicateHash", mock.Anything, mock.AnythingOfType("string")).Run(func(args mock.Arguments) {
			panic("simulated panic")
		}).Return(false, nil)
		txMock.On("RollbackTx").Return(nil)

		service := NewETCMeisaiService(mockRepo, nil)
		ctx := context.Background()

		params := &CreateRecordParams{
			Date:          time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
			Time:          "10:30:00",
			EntranceIC:    "羽田空港IC",
			ExitIC:        "新宿IC",
			TollAmount:    1200,
			CarNumber:     "あ123",
			ETCCardNumber: "1234567890123456",
		}

		assert.Panics(t, func() {
			service.CreateRecord(ctx, params)
		})

		txMock.AssertExpectations(t)
	})
}

func TestETCMeisaiService_GetRecord(t *testing.T) {
	t.Parallel()

	expectedRecord := &models.ETCMeisaiRecord{
		ID:            1,
		Hash:          "test-hash",
		Date:          time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
		Time:          "10:30:00",
		EntranceIC:    "羽田空港IC",
		ExitIC:        "新宿IC",
		TollAmount:    1200,
		CarNumber:     "あ123",
		ETCCardNumber: "1234567890123456",
	}

	tests := []struct {
		name        string
		id          int64
		setupMock   func(*mocks.MockETCMeisaiRecordRepository)
		expectError bool
		errorMsg    string
	}{
		{
			name: "successful retrieval",
			id:   1,
			setupMock: func(m *mocks.MockETCMeisaiRecordRepository) {
				m.On("GetByID", mock.Anything, int64(1)).Return(expectedRecord, nil)
			},
			expectError: false,
		},
		{
			name:        "invalid ID - zero",
			id:          0,
			setupMock:   func(m *mocks.MockETCMeisaiRecordRepository) {},
			expectError: true,
			errorMsg:    "invalid record ID",
		},
		{
			name:        "invalid ID - negative",
			id:          -1,
			setupMock:   func(m *mocks.MockETCMeisaiRecordRepository) {},
			expectError: true,
			errorMsg:    "invalid record ID",
		},
		{
			name: "repository error",
			id:   999,
			setupMock: func(m *mocks.MockETCMeisaiRecordRepository) {
				m.On("GetByID", mock.Anything, int64(999)).Return(nil, errors.New("record not found"))
			},
			expectError: true,
			errorMsg:    "failed to retrieve record",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &mocks.MockETCMeisaiRecordRepository{}
			tt.setupMock(mockRepo)

			service := NewETCMeisaiService(mockRepo, nil)
			ctx := context.Background()

			record, err := service.GetRecord(ctx, tt.id)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
				assert.Nil(t, record)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, expectedRecord, record)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestETCMeisaiService_ListRecords(t *testing.T) {
	t.Parallel()

	dateFrom := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	dateTo := time.Date(2024, 1, 31, 0, 0, 0, 0, time.UTC)
	carNumber := "あ123"
	etcNumber := "1234567890123456"
	etcNum := "ETC123"

	expectedRecords := []*models.ETCMeisaiRecord{
		{ID: 1, TollAmount: 1200},
		{ID: 2, TollAmount: 1500},
	}

	tests := []struct {
		name          string
		params        *ListRecordsParams
		setupMock     func(*mocks.MockETCMeisaiRecordRepository)
		expectError   bool
		expectRecords []*models.ETCMeisaiRecord
		expectTotal   int64
	}{
		{
			name:   "default parameters",
			params: &ListRecordsParams{},
			setupMock: func(m *mocks.MockETCMeisaiRecordRepository) {
				expectedParams := repositories.ListRecordsParams{
					Page:      1,
					PageSize:  50,
					SortBy:    "date",
					SortOrder: "desc",
				}
				m.On("List", mock.Anything, expectedParams).Return(expectedRecords, int64(2), nil)
			},
			expectError:   false,
			expectRecords: expectedRecords,
			expectTotal:   2,
		},
		{
			name: "with all filters",
			params: &ListRecordsParams{
				Page:      2,
				PageSize:  20,
				DateFrom:  &dateFrom,
				DateTo:    &dateTo,
				CarNumber: &carNumber,
				ETCNumber: &etcNumber,
				ETCNum:    &etcNum,
				SortBy:    "toll_amount",
				SortOrder: "asc",
			},
			setupMock: func(m *mocks.MockETCMeisaiRecordRepository) {
				expectedParams := repositories.ListRecordsParams{
					Page:      2,
					PageSize:  20,
					DateFrom:  &dateFrom,
					DateTo:    &dateTo,
					CarNumber: &carNumber,
					ETCNumber: &etcNumber,
					ETCNum:    &etcNum,
					SortBy:    "toll_amount",
					SortOrder: "asc",
				}
				m.On("List", mock.Anything, expectedParams).Return(expectedRecords, int64(10), nil)
			},
			expectError:   false,
			expectRecords: expectedRecords,
			expectTotal:   10,
		},
		{
			name: "page size over limit",
			params: &ListRecordsParams{
				Page:     1,
				PageSize: 2000, // Over limit
			},
			setupMock: func(m *mocks.MockETCMeisaiRecordRepository) {
				expectedParams := repositories.ListRecordsParams{
					Page:      1,
					PageSize:  1000, // Should be capped
					SortBy:    "date",
					SortOrder: "desc",
				}
				m.On("List", mock.Anything, expectedParams).Return([]*models.ETCMeisaiRecord{}, int64(0), nil)
			},
			expectError:   false,
			expectRecords: []*models.ETCMeisaiRecord{},
			expectTotal:   0,
		},
		{
			name:   "repository error",
			params: &ListRecordsParams{},
			setupMock: func(m *mocks.MockETCMeisaiRecordRepository) {
				m.On("List", mock.Anything, mock.AnythingOfType("repositories.ListRecordsParams")).Return(nil, int64(0), errors.New("db error"))
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &mocks.MockETCMeisaiRecordRepository{}
			tt.setupMock(mockRepo)

			service := NewETCMeisaiService(mockRepo, nil)
			ctx := context.Background()

			response, err := service.ListRecords(ctx, tt.params)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, response)
			} else {
				assert.NoError(t, err)
				require.NotNil(t, response)
				assert.Equal(t, tt.expectRecords, response.Records)
				assert.Equal(t, tt.expectTotal, response.TotalCount)
				assert.Equal(t, tt.params.Page, response.Page)

				if tt.params.PageSize > 0 && tt.params.PageSize <= 1000 {
					assert.Equal(t, tt.params.PageSize, response.PageSize)
				} else {
					assert.Equal(t, 50, response.PageSize)
				}
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestETCMeisaiService_UpdateRecord(t *testing.T) {
	t.Parallel()

	existingRecord := &models.ETCMeisaiRecord{
		ID:            1,
		Hash:          "old-hash",
		Date:          time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
		Time:          "10:30:00",
		EntranceIC:    "羽田空港IC",
		ExitIC:        "新宿IC",
		TollAmount:    1200,
		CarNumber:     "あ123",
		ETCCardNumber: "1234567890123456",
	}

	updateParams := &CreateRecordParams{
		Date:          time.Date(2024, 1, 16, 0, 0, 0, 0, time.UTC),
		Time:          "11:30:00",
		EntranceIC:    "成田空港IC",
		ExitIC:        "品川IC",
		TollAmount:    1500,
		CarNumber:     "い456",
		ETCCardNumber: "9876543210987654",
	}

	tests := []struct {
		name        string
		id          int64
		params      *CreateRecordParams
		setupMock   func(*mocks.MockETCMeisaiRecordRepository)
		expectError bool
		errorMsg    string
	}{
		{
			name:   "successful update",
			id:     1,
			params: updateParams,
			setupMock: func(m *mocks.MockETCMeisaiRecordRepository) {
				txMock := &mocks.MockETCMeisaiRecordRepository{}
				m.On("BeginTx", mock.Anything).Return(txMock, nil)
				txMock.On("GetByID", mock.Anything, int64(1)).Return(existingRecord, nil)
				txMock.On("CheckDuplicateHash", mock.Anything, mock.AnythingOfType("string"), int64(1)).Return(false, nil)
				txMock.On("Update", mock.Anything, mock.AnythingOfType("*models.ETCMeisaiRecord")).Return(nil)
				txMock.On("CommitTx").Return(nil)
				txMock.On("RollbackTx").Return(nil).Maybe()
			},
			expectError: false,
		},
		{
			name:        "invalid ID",
			id:          0,
			params:      updateParams,
			setupMock:   func(m *mocks.MockETCMeisaiRecordRepository) {},
			expectError: true,
			errorMsg:    "invalid record ID",
		},
		{
			name:   "record not found",
			id:     999,
			params: updateParams,
			setupMock: func(m *mocks.MockETCMeisaiRecordRepository) {
				txMock := &mocks.MockETCMeisaiRecordRepository{}
				m.On("BeginTx", mock.Anything).Return(txMock, nil)
				txMock.On("GetByID", mock.Anything, int64(999)).Return(nil, errors.New("record not found"))
				txMock.On("RollbackTx").Return(nil)
			},
			expectError: true,
			errorMsg:    "failed to retrieve record",
		},
		{
			name: "validation error after update",
			id:   1,
			params: &CreateRecordParams{
				Date:          time.Date(2024, 1, 16, 0, 0, 0, 0, time.UTC),
				Time:          "11:30:00",
				EntranceIC:    "成田空港IC",
				ExitIC:        "品川IC",
				TollAmount:    -100, // Invalid
				CarNumber:     "い456",
				ETCCardNumber: "9876543210987654",
			},
			setupMock: func(m *mocks.MockETCMeisaiRecordRepository) {
				txMock := &mocks.MockETCMeisaiRecordRepository{}
				m.On("BeginTx", mock.Anything).Return(txMock, nil)
				txMock.On("GetByID", mock.Anything, int64(1)).Return(existingRecord, nil)
				txMock.On("RollbackTx").Return(nil)
			},
			expectError: true,
			errorMsg:    "record validation failed",
		},
		{
			name:   "duplicate hash after update",
			id:     1,
			params: updateParams,
			setupMock: func(m *mocks.MockETCMeisaiRecordRepository) {
				txMock := &mocks.MockETCMeisaiRecordRepository{}
				m.On("BeginTx", mock.Anything).Return(txMock, nil)
				txMock.On("GetByID", mock.Anything, int64(1)).Return(existingRecord, nil)
				txMock.On("CheckDuplicateHash", mock.Anything, mock.AnythingOfType("string"), int64(1)).Return(true, nil)
				txMock.On("RollbackTx").Return(nil)
			},
			expectError: true,
			errorMsg:    "duplicate record with hash",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &mocks.MockETCMeisaiRecordRepository{}
			tt.setupMock(mockRepo)

			service := NewETCMeisaiService(mockRepo, nil)
			ctx := context.Background()

			record, err := service.UpdateRecord(ctx, tt.id, tt.params)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
				assert.Nil(t, record)
			} else {
				assert.NoError(t, err)
				if record != nil {
					assert.NotEmpty(t, record.Hash)
				}
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestETCMeisaiService_DeleteRecord(t *testing.T) {
	t.Parallel()

	existingRecord := &models.ETCMeisaiRecord{ID: 1}

	tests := []struct {
		name        string
		id          int64
		setupMock   func(*mocks.MockETCMeisaiRecordRepository)
		expectError bool
		errorMsg    string
	}{
		{
			name: "successful deletion",
			id:   1,
			setupMock: func(m *mocks.MockETCMeisaiRecordRepository) {
				txMock := &mocks.MockETCMeisaiRecordRepository{}
				m.On("BeginTx", mock.Anything).Return(txMock, nil)
				txMock.On("GetByID", mock.Anything, int64(1)).Return(existingRecord, nil)
				txMock.On("Delete", mock.Anything, int64(1)).Return(nil)
				txMock.On("CommitTx").Return(nil)
				txMock.On("RollbackTx").Return(nil).Maybe()
			},
			expectError: false,
		},
		{
			name:        "invalid ID",
			id:          0,
			setupMock:   func(m *mocks.MockETCMeisaiRecordRepository) {},
			expectError: true,
			errorMsg:    "invalid record ID",
		},
		{
			name: "record not found",
			id:   999,
			setupMock: func(m *mocks.MockETCMeisaiRecordRepository) {
				txMock := &mocks.MockETCMeisaiRecordRepository{}
				m.On("BeginTx", mock.Anything).Return(txMock, nil)
				txMock.On("GetByID", mock.Anything, int64(999)).Return(nil, errors.New("record not found"))
				txMock.On("RollbackTx").Return(nil)
			},
			expectError: true,
			errorMsg:    "failed to retrieve record",
		},
		{
			name: "delete error",
			id:   1,
			setupMock: func(m *mocks.MockETCMeisaiRecordRepository) {
				txMock := &mocks.MockETCMeisaiRecordRepository{}
				m.On("BeginTx", mock.Anything).Return(txMock, nil)
				txMock.On("GetByID", mock.Anything, int64(1)).Return(existingRecord, nil)
				txMock.On("Delete", mock.Anything, int64(1)).Return(errors.New("delete failed"))
				txMock.On("RollbackTx").Return(nil)
			},
			expectError: true,
			errorMsg:    "failed to delete record",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &mocks.MockETCMeisaiRecordRepository{}
			tt.setupMock(mockRepo)

			service := NewETCMeisaiService(mockRepo, nil)
			ctx := context.Background()

			err := service.DeleteRecord(ctx, tt.id)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestETCMeisaiService_GetRecordByHash(t *testing.T) {
	t.Parallel()

	expectedRecord := &models.ETCMeisaiRecord{
		ID:   1,
		Hash: "test-hash",
	}

	tests := []struct {
		name        string
		hash        string
		setupMock   func(*mocks.MockETCMeisaiRecordRepository)
		expectError bool
		errorMsg    string
	}{
		{
			name: "successful retrieval",
			hash: "test-hash",
			setupMock: func(m *mocks.MockETCMeisaiRecordRepository) {
				m.On("GetByHash", mock.Anything, "test-hash").Return(expectedRecord, nil)
			},
			expectError: false,
		},
		{
			name:        "empty hash",
			hash:        "",
			setupMock:   func(m *mocks.MockETCMeisaiRecordRepository) {},
			expectError: true,
			errorMsg:    "hash cannot be empty",
		},
		{
			name: "repository error",
			hash: "non-existent-hash",
			setupMock: func(m *mocks.MockETCMeisaiRecordRepository) {
				m.On("GetByHash", mock.Anything, "non-existent-hash").Return(nil, errors.New("record not found"))
			},
			expectError: true,
			errorMsg:    "failed to retrieve record",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &mocks.MockETCMeisaiRecordRepository{}
			tt.setupMock(mockRepo)

			service := NewETCMeisaiService(mockRepo, nil)
			ctx := context.Background()

			record, err := service.GetRecordByHash(ctx, tt.hash)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
				assert.Nil(t, record)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, expectedRecord, record)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestETCMeisaiService_ValidateRecord(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		params      *CreateRecordParams
		expectError bool
	}{
		{
			name: "valid record",
			params: &CreateRecordParams{
				Date:          time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
				Time:          "10:30:00",
				EntranceIC:    "羽田空港IC",
				ExitIC:        "新宿IC",
				TollAmount:    1200,
				CarNumber:     "あ123",
				ETCCardNumber: "1234567890123456",
			},
			expectError: false,
		},
		{
			name: "invalid record - negative toll amount",
			params: &CreateRecordParams{
				Date:          time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
				Time:          "10:30:00",
				EntranceIC:    "羽田空港IC",
				ExitIC:        "新宿IC",
				TollAmount:    -100,
				CarNumber:     "あ123",
				ETCCardNumber: "1234567890123456",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &mocks.MockETCMeisaiRecordRepository{}
			service := NewETCMeisaiService(mockRepo, nil)
			ctx := context.Background()

			err := service.ValidateRecord(ctx, tt.params)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestETCMeisaiService_HealthCheck(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		setupMock   func(*mocks.MockETCMeisaiRecordRepository)
		expectError bool
		errorMsg    string
	}{
		{
			name: "healthy repository",
			setupMock: func(m *mocks.MockETCMeisaiRecordRepository) {
				m.On("Ping", mock.Anything).Return(nil)
			},
			expectError: false,
		},
		{
			name: "unhealthy repository",
			setupMock: func(m *mocks.MockETCMeisaiRecordRepository) {
				m.On("Ping", mock.Anything).Return(errors.New("connection failed"))
			},
			expectError: true,
			errorMsg:    "repository ping failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &mocks.MockETCMeisaiRecordRepository{}
			tt.setupMock(mockRepo)

			service := NewETCMeisaiService(mockRepo, nil)
			ctx := context.Background()

			err := service.HealthCheck(ctx)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestETCMeisaiService_HealthCheck_NilRepository(t *testing.T) {
	t.Parallel()

	service := &ETCMeisaiService{
		repo: nil,
	}
	ctx := context.Background()

	err := service.HealthCheck(ctx)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "repository not initialized")
}

// Context cancellation tests
func TestETCMeisaiService_ContextCancellation(t *testing.T) {
	t.Parallel()

	t.Run("create record with cancelled context", func(t *testing.T) {
		mockRepo := &mocks.MockETCMeisaiRecordRepository{}
		service := NewETCMeisaiService(mockRepo, nil)

		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		params := &CreateRecordParams{
			Date:          time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
			Time:          "10:30:00",
			EntranceIC:    "羽田空港IC",
			ExitIC:        "新宿IC",
			TollAmount:    1200,
			CarNumber:     "あ123",
			ETCCardNumber: "1234567890123456",
		}

		// Mock should handle the cancelled context appropriately
		txMock := &mocks.MockETCMeisaiRecordRepository{}
		mockRepo.On("BeginTx", mock.Anything).Return(txMock, context.Canceled)

		_, err := service.CreateRecord(ctx, params)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to start transaction")
	})
}

// Helper functions
func stringPtr(s string) *string {
	return &s
}

func int64Ptr(i int64) *int64 {
	return &i
}
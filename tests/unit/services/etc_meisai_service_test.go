package services_test

import (
	"context"
	"errors"
	"log"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/yhonda-ohishi/etc_meisai/src/models"
	"github.com/yhonda-ohishi/etc_meisai/src/repositories"
	"github.com/yhonda-ohishi/etc_meisai/src/services"
)

// MockETCMeisaiRecordRepository is a mock implementation of ETCMeisaiRecordRepository
type MockETCMeisaiRecordRepository struct {
	mock.Mock
}

func (m *MockETCMeisaiRecordRepository) Create(ctx context.Context, record *models.ETCMeisaiRecord) error {
	args := m.Called(ctx, record)
	return args.Error(0)
}

func (m *MockETCMeisaiRecordRepository) GetByID(ctx context.Context, id int64) (*models.ETCMeisaiRecord, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.ETCMeisaiRecord), args.Error(1)
}

func (m *MockETCMeisaiRecordRepository) Update(ctx context.Context, record *models.ETCMeisaiRecord) error {
	args := m.Called(ctx, record)
	return args.Error(0)
}

func (m *MockETCMeisaiRecordRepository) Delete(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockETCMeisaiRecordRepository) GetByHash(ctx context.Context, hash string) (*models.ETCMeisaiRecord, error) {
	args := m.Called(ctx, hash)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.ETCMeisaiRecord), args.Error(1)
}

func (m *MockETCMeisaiRecordRepository) CheckDuplicateHash(ctx context.Context, hash string, excludeID ...int64) (bool, error) {
	args := m.Called(ctx, hash, excludeID)
	return args.Bool(0), args.Error(1)
}

func (m *MockETCMeisaiRecordRepository) List(ctx context.Context, params repositories.ListRecordsParams) ([]*models.ETCMeisaiRecord, int64, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*models.ETCMeisaiRecord), args.Get(1).(int64), args.Error(2)
}

func (m *MockETCMeisaiRecordRepository) BeginTx(ctx context.Context) (repositories.ETCMeisaiRecordRepository, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(repositories.ETCMeisaiRecordRepository), args.Error(1)
}

func (m *MockETCMeisaiRecordRepository) CommitTx() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockETCMeisaiRecordRepository) RollbackTx() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockETCMeisaiRecordRepository) Ping(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func TestETCMeisaiService_NewETCMeisaiService(t *testing.T) {
	mockRepo := &MockETCMeisaiRecordRepository{}
	logger := log.New(os.Stdout, "TEST: ", log.LstdFlags)

	service := services.NewETCMeisaiService(mockRepo, logger)
	assert.NotNil(t, service)
}

func TestETCMeisaiService_CreateRecord(t *testing.T) {
	mockRepo := &MockETCMeisaiRecordRepository{}
	logger := log.New(os.Stdout, "TEST: ", log.LstdFlags)
	service := services.NewETCMeisaiService(mockRepo, logger)

	validParams := &services.CreateRecordParams{
		Date:          time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC),
		Time:          "09:30",
		EntranceIC:    "東京IC",
		ExitIC:        "大阪IC",
		TollAmount:    1000,
		CarNumber:     "品川123あ1234",
		ETCCardNumber: "1234567890",
	}

	tests := []struct {
		name      string
		params    *services.CreateRecordParams
		setupMock func()
		wantErr   bool
		errMsg    string
	}{
		{
			name:   "successful creation",
			params: validParams,
			setupMock: func() {
				mockRepo.On("CheckDuplicateHash", mock.Anything, mock.AnythingOfType("string"), mock.Anything).
					Return(false, nil).Once()
				mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*models.ETCMeisaiRecord")).
					Return(nil).Once()
			},
			wantErr: false,
		},
		{
			name:   "nil params",
			params: nil,
			setupMock: func() {
				// No mock setup needed
			},
			wantErr: true,
			errMsg:  "params cannot be nil",
		},
		{
			name:   "duplicate hash",
			params: validParams,
			setupMock: func() {
				mockRepo.On("CheckDuplicateHash", mock.Anything, mock.AnythingOfType("string"), mock.Anything).
					Return(true, nil).Once()
			},
			wantErr: true,
			errMsg:  "duplicate",
		},
		{
			name:   "repository error",
			params: validParams,
			setupMock: func() {
				mockRepo.On("CheckDuplicateHash", mock.Anything, mock.AnythingOfType("string"), mock.Anything).
					Return(false, nil).Once()
				mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*models.ETCMeisaiRecord")).
					Return(errors.New("database error")).Once()
			},
			wantErr: true,
			errMsg:  "failed to create",
		},
		{
			name: "invalid params - missing date",
			params: &services.CreateRecordParams{
				Time:          "09:30",
				EntranceIC:    "東京IC",
				ExitIC:        "大阪IC",
				TollAmount:    1000,
				CarNumber:     "品川123あ1234",
				ETCCardNumber: "1234567890",
			},
			setupMock: func() {
				// No mock setup needed as validation should fail first
			},
			wantErr: true,
			errMsg:  "date is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo.ExpectedCalls = nil // Reset mock expectations
			mockRepo.Calls = nil         // Reset call history
			tt.setupMock()

			ctx := context.Background()
			record, err := service.CreateRecord(ctx, tt.params)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
				assert.Nil(t, record)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, record)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestETCMeisaiService_GetRecord(t *testing.T) {
	mockRepo := &MockETCMeisaiRecordRepository{}
	logger := log.New(os.Stdout, "TEST: ", log.LstdFlags)
	service := services.NewETCMeisaiService(mockRepo, logger)

	tests := []struct {
		name      string
		id        int64
		setupMock func()
		wantErr   bool
		errMsg    string
	}{
		{
			name: "successful retrieval",
			id:   1,
			setupMock: func() {
				mockRecord := &models.ETCMeisaiRecord{
					ID:            1,
					Date:          time.Now(),
					Time:          "09:30",
					EntranceIC:    "東京IC",
					ExitIC:        "大阪IC",
					TollAmount:    1000,
					CarNumber:     "品川123あ1234",
					ETCCardNumber: "1234567890",
				}
				mockRepo.On("GetByID", mock.Anything, int64(1)).
					Return(mockRecord, nil).Once()
			},
			wantErr: false,
		},
		{
			name: "record not found",
			id:   999,
			setupMock: func() {
				mockRepo.On("GetByID", mock.Anything, int64(999)).
					Return(nil, errors.New("record not found")).Once()
			},
			wantErr: true,
			errMsg:  "not found",
		},
		{
			name: "invalid id",
			id:   0,
			setupMock: func() {
				// No mock setup needed
			},
			wantErr: true,
			errMsg:  "invalid id",
		},
		{
			name: "repository error",
			id:   1,
			setupMock: func() {
				mockRepo.On("GetByID", mock.Anything, int64(1)).
					Return(nil, errors.New("database error")).Once()
			},
			wantErr: true,
			errMsg:  "failed to get record",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo.ExpectedCalls = nil
			mockRepo.Calls = nil
			tt.setupMock()

			ctx := context.Background()
			record, err := service.GetRecord(ctx, tt.id)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
				assert.Nil(t, record)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, record)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestETCMeisaiService_ListRecords(t *testing.T) {
	mockRepo := &MockETCMeisaiRecordRepository{}
	logger := log.New(os.Stdout, "TEST: ", log.LstdFlags)
	service := services.NewETCMeisaiService(mockRepo, logger)

	tests := []struct {
		name      string
		params    *services.ListRecordsParams
		setupMock func()
		wantErr   bool
		errMsg    string
	}{
		{
			name: "successful list",
			params: &services.ListRecordsParams{
				Page:     1,
				PageSize: 10,
			},
			setupMock: func() {
				records := []*models.ETCMeisaiRecord{
					{ID: 1}, {ID: 2}, {ID: 3},
				}
				mockRepo.On("List", mock.Anything, mock.AnythingOfType("repositories.ListRecordsParams")).
					Return(records, int64(3), nil).Once()
			},
			wantErr: false,
		},
		{
			name:   "nil params - should use defaults",
			params: nil,
			setupMock: func() {
				records := []*models.ETCMeisaiRecord{}
				mockRepo.On("List", mock.Anything, mock.AnythingOfType("repositories.ListRecordsParams")).
					Return(records, int64(0), nil).Once()
			},
			wantErr: false,
		},
		{
			name: "repository error",
			params: &services.ListRecordsParams{
				Page:     1,
				PageSize: 10,
			},
			setupMock: func() {
				mockRepo.On("List", mock.Anything, mock.AnythingOfType("repositories.ListRecordsParams")).
					Return(nil, int64(0), errors.New("database error")).Once()
			},
			wantErr: true,
			errMsg:  "failed to list records",
		},
		{
			name: "invalid page",
			params: &services.ListRecordsParams{
				Page:     0,
				PageSize: 10,
			},
			setupMock: func() {
				// Page will be adjusted to 1 automatically
				records := []*models.ETCMeisaiRecord{}
				mockRepo.On("List", mock.Anything, mock.AnythingOfType("repositories.ListRecordsParams")).
					Return(records, int64(0), nil).Once()
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo.ExpectedCalls = nil
			mockRepo.Calls = nil
			tt.setupMock()

			ctx := context.Background()
			response, err := service.ListRecords(ctx, tt.params)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
				assert.Nil(t, response)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, response)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestETCMeisaiService_UpdateRecord(t *testing.T) {
	mockRepo := &MockETCMeisaiRecordRepository{}
	logger := log.New(os.Stdout, "TEST: ", log.LstdFlags)
	service := services.NewETCMeisaiService(mockRepo, logger)

	validParams := &services.CreateRecordParams{
		Date:          time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC),
		Time:          "10:30",
		EntranceIC:    "名古屋IC",
		ExitIC:        "京都IC",
		TollAmount:    800,
		CarNumber:     "品川456い5678",
		ETCCardNumber: "9876543210",
	}

	tests := []struct {
		name      string
		id        int64
		params    *services.CreateRecordParams
		setupMock func()
		wantErr   bool
		errMsg    string
	}{
		{
			name:   "successful update",
			id:     1,
			params: validParams,
			setupMock: func() {
				existingRecord := &models.ETCMeisaiRecord{
					ID:            1,
					Date:          time.Now(),
					Time:          "09:30",
					EntranceIC:    "東京IC",
					ExitIC:        "大阪IC",
					TollAmount:    1000,
					CarNumber:     "品川123あ1234",
					ETCCardNumber: "1234567890",
				}
				mockRepo.On("GetByID", mock.Anything, int64(1)).
					Return(existingRecord, nil).Once()
				mockRepo.On("CheckDuplicateHash", mock.Anything, mock.AnythingOfType("string"), []int64{int64(1)}).
					Return(false, nil).Once()
				mockRepo.On("Update", mock.Anything, mock.AnythingOfType("*models.ETCMeisaiRecord")).
					Return(nil).Once()
			},
			wantErr: false,
		},
		{
			name:   "record not found",
			id:     999,
			params: validParams,
			setupMock: func() {
				mockRepo.On("GetByID", mock.Anything, int64(999)).
					Return(nil, errors.New("record not found")).Once()
			},
			wantErr: true,
			errMsg:  "not found",
		},
		{
			name:   "invalid id",
			id:     0,
			params: validParams,
			setupMock: func() {
				// No mock setup needed
			},
			wantErr: true,
			errMsg:  "invalid id",
		},
		{
			name:   "nil params",
			id:     1,
			params: nil,
			setupMock: func() {
				// No mock setup needed
			},
			wantErr: true,
			errMsg:  "params cannot be nil",
		},
		{
			name:   "duplicate hash",
			id:     1,
			params: validParams,
			setupMock: func() {
				existingRecord := &models.ETCMeisaiRecord{
					ID: 1,
				}
				mockRepo.On("GetByID", mock.Anything, int64(1)).
					Return(existingRecord, nil).Once()
				mockRepo.On("CheckDuplicateHash", mock.Anything, mock.AnythingOfType("string"), []int64{int64(1)}).
					Return(true, nil).Once()
			},
			wantErr: true,
			errMsg:  "duplicate",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo.ExpectedCalls = nil
			mockRepo.Calls = nil
			tt.setupMock()

			ctx := context.Background()
			record, err := service.UpdateRecord(ctx, tt.id, tt.params)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
				assert.Nil(t, record)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, record)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestETCMeisaiService_DeleteRecord(t *testing.T) {
	mockRepo := &MockETCMeisaiRecordRepository{}
	logger := log.New(os.Stdout, "TEST: ", log.LstdFlags)
	service := services.NewETCMeisaiService(mockRepo, logger)

	tests := []struct {
		name      string
		id        int64
		setupMock func()
		wantErr   bool
		errMsg    string
	}{
		{
			name: "successful deletion",
			id:   1,
			setupMock: func() {
				existingRecord := &models.ETCMeisaiRecord{
					ID: 1,
				}
				mockRepo.On("GetByID", mock.Anything, int64(1)).
					Return(existingRecord, nil).Once()
				mockRepo.On("Delete", mock.Anything, int64(1)).
					Return(nil).Once()
			},
			wantErr: false,
		},
		{
			name: "record not found",
			id:   999,
			setupMock: func() {
				mockRepo.On("GetByID", mock.Anything, int64(999)).
					Return(nil, errors.New("record not found")).Once()
			},
			wantErr: true,
			errMsg:  "not found",
		},
		{
			name: "invalid id",
			id:   0,
			setupMock: func() {
				// No mock setup needed
			},
			wantErr: true,
			errMsg:  "invalid id",
		},
		{
			name: "repository error",
			id:   1,
			setupMock: func() {
				existingRecord := &models.ETCMeisaiRecord{
					ID: 1,
				}
				mockRepo.On("GetByID", mock.Anything, int64(1)).
					Return(existingRecord, nil).Once()
				mockRepo.On("Delete", mock.Anything, int64(1)).
					Return(errors.New("database error")).Once()
			},
			wantErr: true,
			errMsg:  "failed to delete",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo.ExpectedCalls = nil
			mockRepo.Calls = nil
			tt.setupMock()

			ctx := context.Background()
			err := service.DeleteRecord(ctx, tt.id)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestETCMeisaiService_GetRecordByHash(t *testing.T) {
	mockRepo := &MockETCMeisaiRecordRepository{}
	logger := log.New(os.Stdout, "TEST: ", log.LstdFlags)
	service := services.NewETCMeisaiService(mockRepo, logger)

	tests := []struct {
		name      string
		hash      string
		setupMock func()
		wantErr   bool
		errMsg    string
	}{
		{
			name: "successful retrieval",
			hash: "abc123",
			setupMock: func() {
				mockRecord := &models.ETCMeisaiRecord{
					ID:   1,
					Hash: "abc123",
				}
				mockRepo.On("GetByHash", mock.Anything, "abc123").
					Return(mockRecord, nil).Once()
			},
			wantErr: false,
		},
		{
			name: "record not found",
			hash: "notfound",
			setupMock: func() {
				mockRepo.On("GetByHash", mock.Anything, "notfound").
					Return(nil, errors.New("record not found")).Once()
			},
			wantErr: true,
			errMsg:  "not found",
		},
		{
			name: "empty hash",
			hash: "",
			setupMock: func() {
				// No mock setup needed
			},
			wantErr: true,
			errMsg:  "hash cannot be empty",
		},
		{
			name: "repository error",
			hash: "abc123",
			setupMock: func() {
				mockRepo.On("GetByHash", mock.Anything, "abc123").
					Return(nil, errors.New("database error")).Once()
			},
			wantErr: true,
			errMsg:  "failed to get record",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo.ExpectedCalls = nil
			mockRepo.Calls = nil
			tt.setupMock()

			ctx := context.Background()
			record, err := service.GetRecordByHash(ctx, tt.hash)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
				assert.Nil(t, record)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, record)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestETCMeisaiService_ValidateRecord(t *testing.T) {
	mockRepo := &MockETCMeisaiRecordRepository{}
	logger := log.New(os.Stdout, "TEST: ", log.LstdFlags)
	service := services.NewETCMeisaiService(mockRepo, logger)

	tests := []struct {
		name    string
		params  *services.CreateRecordParams
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid record",
			params: &services.CreateRecordParams{
				Date:          time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC),
				Time:          "09:30",
				EntranceIC:    "東京IC",
				ExitIC:        "大阪IC",
				TollAmount:    1000,
				CarNumber:     "品川123あ1234",
				ETCCardNumber: "1234567890",
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
			name: "missing date",
			params: &services.CreateRecordParams{
				Time:          "09:30",
				EntranceIC:    "東京IC",
				ExitIC:        "大阪IC",
				TollAmount:    1000,
				CarNumber:     "品川123あ1234",
				ETCCardNumber: "1234567890",
			},
			wantErr: true,
			errMsg:  "date is required",
		},
		{
			name: "negative toll amount",
			params: &services.CreateRecordParams{
				Date:          time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC),
				Time:          "09:30",
				EntranceIC:    "東京IC",
				ExitIC:        "大阪IC",
				TollAmount:    -100,
				CarNumber:     "品川123あ1234",
				ETCCardNumber: "1234567890",
			},
			wantErr: true,
			errMsg:  "toll amount must be positive",
		},
		{
			name: "empty car number",
			params: &services.CreateRecordParams{
				Date:          time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC),
				Time:          "09:30",
				EntranceIC:    "東京IC",
				ExitIC:        "大阪IC",
				TollAmount:    1000,
				CarNumber:     "",
				ETCCardNumber: "1234567890",
			},
			wantErr: true,
			errMsg:  "car number is required",
		},
		{
			name: "empty ETC card number",
			params: &services.CreateRecordParams{
				Date:          time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC),
				Time:          "09:30",
				EntranceIC:    "東京IC",
				ExitIC:        "大阪IC",
				TollAmount:    1000,
				CarNumber:     "品川123あ1234",
				ETCCardNumber: "",
			},
			wantErr: true,
			errMsg:  "ETC card number is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			err := service.ValidateRecord(ctx, tt.params)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestETCMeisaiService_HealthCheck(t *testing.T) {
	mockRepo := &MockETCMeisaiRecordRepository{}
	logger := log.New(os.Stdout, "TEST: ", log.LstdFlags)
	service := services.NewETCMeisaiService(mockRepo, logger)

	tests := []struct {
		name      string
		setupMock func()
		wantErr   bool
	}{
		{
			name: "healthy",
			setupMock: func() {
				mockRepo.On("Ping", mock.Anything).Return(nil).Once()
			},
			wantErr: false,
		},
		{
			name: "unhealthy",
			setupMock: func() {
				mockRepo.On("Ping", mock.Anything).Return(errors.New("connection failed")).Once()
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo.ExpectedCalls = nil
			mockRepo.Calls = nil
			tt.setupMock()

			ctx := context.Background()
			err := service.HealthCheck(ctx)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}
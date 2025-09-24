package repositories

import (
	"context"

	"github.com/stretchr/testify/mock"
	"github.com/yhonda-ohishi/etc_meisai/src/models"
)

// MockETCMeisaiRecordRepository is a mock implementation of ETCMeisaiRecordRepository
type MockETCMeisaiRecordRepository struct {
	mock.Mock
}

// Create mocks the Create method
func (m *MockETCMeisaiRecordRepository) Create(ctx context.Context, record *models.ETCMeisaiRecord) error {
	args := m.Called(ctx, record)
	return args.Error(0)
}

// GetByID mocks the GetByID method
func (m *MockETCMeisaiRecordRepository) GetByID(ctx context.Context, id int64) (*models.ETCMeisaiRecord, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.ETCMeisaiRecord), args.Error(1)
}

// Update mocks the Update method
func (m *MockETCMeisaiRecordRepository) Update(ctx context.Context, record *models.ETCMeisaiRecord) error {
	args := m.Called(ctx, record)
	return args.Error(0)
}

// Delete mocks the Delete method
func (m *MockETCMeisaiRecordRepository) Delete(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// GetByHash mocks the GetByHash method
func (m *MockETCMeisaiRecordRepository) GetByHash(ctx context.Context, hash string) (*models.ETCMeisaiRecord, error) {
	args := m.Called(ctx, hash)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.ETCMeisaiRecord), args.Error(1)
}

// CheckDuplicateHash mocks the CheckDuplicateHash method
func (m *MockETCMeisaiRecordRepository) CheckDuplicateHash(ctx context.Context, hash string, excludeID ...int64) (bool, error) {
	// Handle variadic argument properly
	if len(excludeID) == 0 {
		args := m.Called(ctx, hash)
		return args.Bool(0), args.Error(1)
	}
	args := m.Called(ctx, hash, excludeID)
	return args.Bool(0), args.Error(1)
}

// List mocks the List method
func (m *MockETCMeisaiRecordRepository) List(ctx context.Context, params ListRecordsParams) ([]*models.ETCMeisaiRecord, int64, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]*models.ETCMeisaiRecord), args.Get(1).(int64), args.Error(2)
}

// BeginTx mocks the BeginTx method
func (m *MockETCMeisaiRecordRepository) BeginTx(ctx context.Context) (ETCMeisaiRecordRepository, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(ETCMeisaiRecordRepository), args.Error(1)
}

// CommitTx mocks the CommitTx method
func (m *MockETCMeisaiRecordRepository) CommitTx() error {
	args := m.Called()
	return args.Error(0)
}

// RollbackTx mocks the RollbackTx method
func (m *MockETCMeisaiRecordRepository) RollbackTx() error {
	args := m.Called()
	return args.Error(0)
}

// Ping mocks the Ping method
func (m *MockETCMeisaiRecordRepository) Ping(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}
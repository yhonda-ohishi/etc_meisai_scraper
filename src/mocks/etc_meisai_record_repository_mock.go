package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"
	"github.com/yhonda-ohishi/etc_meisai/src/models"
	"github.com/yhonda-ohishi/etc_meisai/src/repositories"
)

// MockETCMeisaiRecordRepository mocks the ETCMeisaiRecordRepository interface
type MockETCMeisaiRecordRepository struct {
	mock.Mock
}

// Create creates a new ETC record
func (m *MockETCMeisaiRecordRepository) Create(ctx context.Context, record *models.ETCMeisaiRecord) error {
	args := m.Called(ctx, record)
	return args.Error(0)
}

// GetByID retrieves an ETC record by ID
func (m *MockETCMeisaiRecordRepository) GetByID(ctx context.Context, id int64) (*models.ETCMeisaiRecord, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.ETCMeisaiRecord), args.Error(1)
}

// Update updates an existing ETC record
func (m *MockETCMeisaiRecordRepository) Update(ctx context.Context, record *models.ETCMeisaiRecord) error {
	args := m.Called(ctx, record)
	return args.Error(0)
}

// Delete performs soft delete on an ETC record
func (m *MockETCMeisaiRecordRepository) Delete(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// GetByHash retrieves an ETC record by its hash
func (m *MockETCMeisaiRecordRepository) GetByHash(ctx context.Context, hash string) (*models.ETCMeisaiRecord, error) {
	args := m.Called(ctx, hash)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.ETCMeisaiRecord), args.Error(1)
}

// CheckDuplicateHash checks if a hash already exists
func (m *MockETCMeisaiRecordRepository) CheckDuplicateHash(ctx context.Context, hash string, excludeID ...int64) (bool, error) {
	args := m.Called(ctx, hash, excludeID)
	return args.Bool(0), args.Error(1)
}

// List lists ETC records with filtering and pagination
func (m *MockETCMeisaiRecordRepository) List(ctx context.Context, params repositories.ListRecordsParams) ([]*models.ETCMeisaiRecord, int64, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*models.ETCMeisaiRecord), args.Get(1).(int64), args.Error(2)
}

// BeginTx starts a new transaction
func (m *MockETCMeisaiRecordRepository) BeginTx(ctx context.Context) (repositories.ETCMeisaiRecordRepository, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(repositories.ETCMeisaiRecordRepository), args.Error(1)
}

// CommitTx commits the transaction
func (m *MockETCMeisaiRecordRepository) CommitTx() error {
	args := m.Called()
	return args.Error(0)
}

// RollbackTx rolls back the transaction
func (m *MockETCMeisaiRecordRepository) RollbackTx() error {
	args := m.Called()
	return args.Error(0)
}

// Ping checks the repository connectivity
func (m *MockETCMeisaiRecordRepository) Ping(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}
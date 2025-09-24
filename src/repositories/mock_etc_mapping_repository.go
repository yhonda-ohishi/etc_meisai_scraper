package repositories

import (
	"context"

	"github.com/stretchr/testify/mock"
	"github.com/yhonda-ohishi/etc_meisai/src/models"
)

// MockETCMappingRepository is a mock implementation of ETCMappingRepository
type MockETCMappingRepository struct {
	mock.Mock
}

// Create mocks the Create method
func (m *MockETCMappingRepository) Create(ctx context.Context, mapping *models.ETCMapping) error {
	args := m.Called(ctx, mapping)
	return args.Error(0)
}

// GetByID mocks the GetByID method
func (m *MockETCMappingRepository) GetByID(ctx context.Context, id int64) (*models.ETCMapping, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.ETCMapping), args.Error(1)
}

// Update mocks the Update method
func (m *MockETCMappingRepository) Update(ctx context.Context, mapping *models.ETCMapping) error {
	args := m.Called(ctx, mapping)
	return args.Error(0)
}

// Delete mocks the Delete method
func (m *MockETCMappingRepository) Delete(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// GetByETCRecordID mocks the GetByETCRecordID method
func (m *MockETCMappingRepository) GetByETCRecordID(ctx context.Context, etcRecordID int64) ([]*models.ETCMapping, error) {
	args := m.Called(ctx, etcRecordID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.ETCMapping), args.Error(1)
}

// GetByMappedEntity mocks the GetByMappedEntity method
func (m *MockETCMappingRepository) GetByMappedEntity(ctx context.Context, entityType string, entityID int64) ([]*models.ETCMapping, error) {
	args := m.Called(ctx, entityType, entityID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.ETCMapping), args.Error(1)
}

// GetActiveMapping mocks the GetActiveMapping method
func (m *MockETCMappingRepository) GetActiveMapping(ctx context.Context, etcRecordID int64) (*models.ETCMapping, error) {
	args := m.Called(ctx, etcRecordID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.ETCMapping), args.Error(1)
}

// List mocks the List method
func (m *MockETCMappingRepository) List(ctx context.Context, params ListMappingsParams) ([]*models.ETCMapping, int64, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]*models.ETCMapping), args.Get(1).(int64), args.Error(2)
}

// BulkCreate mocks the BulkCreate method
func (m *MockETCMappingRepository) BulkCreate(ctx context.Context, mappings []*models.ETCMapping) error {
	args := m.Called(ctx, mappings)
	return args.Error(0)
}

// UpdateStatus mocks the UpdateStatus method
func (m *MockETCMappingRepository) UpdateStatus(ctx context.Context, id int64, status string) error {
	args := m.Called(ctx, id, status)
	return args.Error(0)
}

// BeginTx mocks the BeginTx method
func (m *MockETCMappingRepository) BeginTx(ctx context.Context) (ETCMappingRepository, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(ETCMappingRepository), args.Error(1)
}

// CommitTx mocks the CommitTx method
func (m *MockETCMappingRepository) CommitTx() error {
	args := m.Called()
	return args.Error(0)
}

// RollbackTx mocks the RollbackTx method
func (m *MockETCMappingRepository) RollbackTx() error {
	args := m.Called()
	return args.Error(0)
}

// Ping mocks the Ping method
func (m *MockETCMappingRepository) Ping(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}
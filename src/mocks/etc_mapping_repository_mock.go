package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"
	"github.com/yhonda-ohishi/etc_meisai/src/models"
	"github.com/yhonda-ohishi/etc_meisai/src/repositories"
)

// MockETCMappingRepository mocks the ETCMappingRepository interface
type MockETCMappingRepository struct {
	mock.Mock
}

// Create creates a new ETC mapping
func (m *MockETCMappingRepository) Create(ctx context.Context, mapping *models.ETCMapping) error {
	args := m.Called(ctx, mapping)
	return args.Error(0)
}

// GetByID retrieves an ETC mapping by ID
func (m *MockETCMappingRepository) GetByID(ctx context.Context, id int64) (*models.ETCMapping, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.ETCMapping), args.Error(1)
}

// Update updates an existing ETC mapping
func (m *MockETCMappingRepository) Update(ctx context.Context, mapping *models.ETCMapping) error {
	args := m.Called(ctx, mapping)
	return args.Error(0)
}

// Delete performs soft delete on an ETC mapping
func (m *MockETCMappingRepository) Delete(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// List lists ETC mappings with filtering and pagination
func (m *MockETCMappingRepository) List(ctx context.Context, params repositories.ListMappingsParams) ([]*models.ETCMapping, int64, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*models.ETCMapping), args.Get(1).(int64), args.Error(2)
}

// GetActiveMapping retrieves the active mapping for an ETC record
func (m *MockETCMappingRepository) GetActiveMapping(ctx context.Context, etcRecordID int64) (*models.ETCMapping, error) {
	args := m.Called(ctx, etcRecordID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.ETCMapping), args.Error(1)
}

// UpdateStatus updates the status of a mapping
func (m *MockETCMappingRepository) UpdateStatus(ctx context.Context, id int64, status string) error {
	args := m.Called(ctx, id, status)
	return args.Error(0)
}

// BeginTx starts a new transaction
func (m *MockETCMappingRepository) BeginTx(ctx context.Context) (repositories.ETCMappingRepository, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(repositories.ETCMappingRepository), args.Error(1)
}

// CommitTx commits the transaction
func (m *MockETCMappingRepository) CommitTx() error {
	args := m.Called()
	return args.Error(0)
}

// RollbackTx rolls back the transaction
func (m *MockETCMappingRepository) RollbackTx() error {
	args := m.Called()
	return args.Error(0)
}

// GetByETCRecordID retrieves mappings by ETC record ID
func (m *MockETCMappingRepository) GetByETCRecordID(ctx context.Context, etcRecordID int64) ([]*models.ETCMapping, error) {
	args := m.Called(ctx, etcRecordID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.ETCMapping), args.Error(1)
}

// GetByMappedEntity retrieves mappings by mapped entity
func (m *MockETCMappingRepository) GetByMappedEntity(ctx context.Context, entityType string, entityID int64) ([]*models.ETCMapping, error) {
	args := m.Called(ctx, entityType, entityID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.ETCMapping), args.Error(1)
}

// BulkCreate creates multiple mappings
func (m *MockETCMappingRepository) BulkCreate(ctx context.Context, mappings []*models.ETCMapping) error {
	args := m.Called(ctx, mappings)
	return args.Error(0)
}

// Ping checks the repository connectivity
func (m *MockETCMappingRepository) Ping(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}
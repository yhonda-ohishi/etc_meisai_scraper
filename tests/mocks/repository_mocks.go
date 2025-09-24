package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"
	"github.com/yhonda-ohishi/etc_meisai/src/models"
)

// MockETCMeisaiRecordRepository mocks the ETC Meisai record repository interface
type MockETCMeisaiRecordRepository struct {
	mock.Mock
}

func (m *MockETCMeisaiRecordRepository) Create(ctx context.Context, record *models.ETCMeisaiRecord) error {
	args := m.Called(ctx, record)
	return args.Error(0)
}

func (m *MockETCMeisaiRecordRepository) GetByID(ctx context.Context, id uint) (*models.ETCMeisaiRecord, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*models.ETCMeisaiRecord), args.Error(1)
}

func (m *MockETCMeisaiRecordRepository) Update(ctx context.Context, record *models.ETCMeisaiRecord) error {
	args := m.Called(ctx, record)
	return args.Error(0)
}

func (m *MockETCMeisaiRecordRepository) Delete(ctx context.Context, id uint) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockETCMeisaiRecordRepository) List(ctx context.Context, limit, offset int) ([]*models.ETCMeisaiRecord, error) {
	args := m.Called(ctx, limit, offset)
	return args.Get(0).([]*models.ETCMeisaiRecord), args.Error(1)
}

func (m *MockETCMeisaiRecordRepository) BulkCreate(ctx context.Context, records []*models.ETCMeisaiRecord) error {
	args := m.Called(ctx, records)
	return args.Error(0)
}

// MockETCMappingRepository mocks the ETC mapping repository interface
type MockETCMappingRepository struct {
	mock.Mock
}

func (m *MockETCMappingRepository) Create(ctx context.Context, mapping *models.ETCMapping) error {
	args := m.Called(ctx, mapping)
	return args.Error(0)
}

func (m *MockETCMappingRepository) GetByETCNum(ctx context.Context, etcNum string) (*models.ETCMapping, error) {
	args := m.Called(ctx, etcNum)
	return args.Get(0).(*models.ETCMapping), args.Error(1)
}

func (m *MockETCMappingRepository) Update(ctx context.Context, mapping *models.ETCMapping) error {
	args := m.Called(ctx, mapping)
	return args.Error(0)
}

func (m *MockETCMappingRepository) Delete(ctx context.Context, id uint) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockETCMappingRepository) List(ctx context.Context, limit, offset int) ([]*models.ETCMapping, error) {
	args := m.Called(ctx, limit, offset)
	return args.Get(0).([]*models.ETCMapping), args.Error(1)
}

// MockImportRepository mocks the import session repository interface
type MockImportRepository struct {
	mock.Mock
}

func (m *MockImportRepository) CreateSession(ctx context.Context, session *models.ImportSession) error {
	args := m.Called(ctx, session)
	return args.Error(0)
}

func (m *MockImportRepository) GetSession(ctx context.Context, sessionID string) (*models.ImportSession, error) {
	args := m.Called(ctx, sessionID)
	return args.Get(0).(*models.ImportSession), args.Error(1)
}

func (m *MockImportRepository) UpdateSession(ctx context.Context, session *models.ImportSession) error {
	args := m.Called(ctx, session)
	return args.Error(0)
}

func (m *MockImportRepository) ListSessions(ctx context.Context, limit, offset int) ([]*models.ImportSession, error) {
	args := m.Called(ctx, limit, offset)
	return args.Get(0).([]*models.ImportSession), args.Error(1)
}
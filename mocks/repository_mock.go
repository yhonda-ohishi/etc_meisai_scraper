package mocks

import (
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/yhonda-ohishi/etc_meisai/src/models"
)

// MockETCRepository は ETCRepository のモック実装
type MockETCRepository struct {
	mock.Mock
}

// Basic CRUD operations

// Create creates a new ETC record
func (m *MockETCRepository) Create(etc *models.ETCMeisai) error {
	args := m.Called(etc)
	return args.Error(0)
}

// GetByID retrieves an ETC record by ID
func (m *MockETCRepository) GetByID(id int64) (*models.ETCMeisai, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.ETCMeisai), args.Error(1)
}

// Update updates an existing ETC record
func (m *MockETCRepository) Update(etc *models.ETCMeisai) error {
	args := m.Called(etc)
	return args.Error(0)
}

// Delete deletes an ETC record
func (m *MockETCRepository) Delete(id int64) error {
	args := m.Called(id)
	return args.Error(0)
}

// Query operations

// GetByDateRange retrieves ETC records within a date range
func (m *MockETCRepository) GetByDateRange(from, to time.Time) ([]*models.ETCMeisai, error) {
	args := m.Called(from, to)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.ETCMeisai), args.Error(1)
}

// GetByHash retrieves an ETC record by hash
func (m *MockETCRepository) GetByHash(hash string) (*models.ETCMeisai, error) {
	args := m.Called(hash)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.ETCMeisai), args.Error(1)
}

// List retrieves a paginated list of ETC records
func (m *MockETCRepository) List(params *models.ETCListParams) ([]*models.ETCMeisai, int64, error) {
	args := m.Called(params)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*models.ETCMeisai), args.Get(1).(int64), args.Error(2)
}

// Bulk operations

// BulkInsert inserts multiple ETC records
func (m *MockETCRepository) BulkInsert(records []*models.ETCMeisai) error {
	args := m.Called(records)
	return args.Error(0)
}

// CheckDuplicatesByHash checks for duplicate records by hash
func (m *MockETCRepository) CheckDuplicatesByHash(hashes []string) (map[string]bool, error) {
	args := m.Called(hashes)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]bool), args.Error(1)
}

// Count operations

// CountByDateRange counts ETC records within a date range
func (m *MockETCRepository) CountByDateRange(from, to time.Time) (int64, error) {
	args := m.Called(from, to)
	return args.Get(0).(int64), args.Error(1)
}

// Search operations

// GetByETCNumber retrieves ETC records by ETC number
func (m *MockETCRepository) GetByETCNumber(etcNumber string, limit int) ([]*models.ETCMeisai, error) {
	args := m.Called(etcNumber, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.ETCMeisai), args.Error(1)
}

// GetByCarNumber retrieves ETC records by car number
func (m *MockETCRepository) GetByCarNumber(carNumber string, limit int) ([]*models.ETCMeisai, error) {
	args := m.Called(carNumber, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.ETCMeisai), args.Error(1)
}

// Summary operations

// GetSummaryByDateRange gets summary statistics for a date range
func (m *MockETCRepository) GetSummaryByDateRange(from, to time.Time) (*models.ETCSummary, error) {
	args := m.Called(from, to)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.ETCSummary), args.Error(1)
}

// MockMappingRepository は MappingRepository のモック実装
type MockMappingRepository struct {
	mock.Mock
}

// Basic CRUD

// Create creates a new mapping
func (m *MockMappingRepository) Create(mapping *models.ETCMeisaiMapping) error {
	args := m.Called(mapping)
	return args.Error(0)
}

// GetByID retrieves a mapping by ID
func (m *MockMappingRepository) GetByID(id int64) (*models.ETCMeisaiMapping, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.ETCMeisaiMapping), args.Error(1)
}

// Update updates an existing mapping
func (m *MockMappingRepository) Update(mapping *models.ETCMeisaiMapping) error {
	args := m.Called(mapping)
	return args.Error(0)
}

// Delete deletes a mapping
func (m *MockMappingRepository) Delete(id int64) error {
	args := m.Called(id)
	return args.Error(0)
}

// Query operations

// GetByETCMeisaiID retrieves mappings by ETC meisai ID
func (m *MockMappingRepository) GetByETCMeisaiID(etcMeisaiID int64) ([]*models.ETCMeisaiMapping, error) {
	args := m.Called(etcMeisaiID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.ETCMeisaiMapping), args.Error(1)
}

// GetByDTakoRowID retrieves a mapping by DTako row ID
func (m *MockMappingRepository) GetByDTakoRowID(dtakoRowID string) (*models.ETCMeisaiMapping, error) {
	args := m.Called(dtakoRowID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.ETCMeisaiMapping), args.Error(1)
}

// List retrieves a paginated list of mappings
func (m *MockMappingRepository) List(params *models.MappingListParams) ([]*models.ETCMeisaiMapping, int64, error) {
	args := m.Called(params)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*models.ETCMeisaiMapping), args.Get(1).(int64), args.Error(2)
}

// Batch operations

// BulkCreateMappings creates multiple mappings in bulk
func (m *MockMappingRepository) BulkCreateMappings(mappings []*models.ETCMeisaiMapping) error {
	args := m.Called(mappings)
	return args.Error(0)
}

// DeleteByETCMeisaiID deletes all mappings for an ETC meisai ID
func (m *MockMappingRepository) DeleteByETCMeisaiID(etcMeisaiID int64) error {
	args := m.Called(etcMeisaiID)
	return args.Error(0)
}

// Auto-matching support

// FindPotentialMatches finds potential matching candidates
func (m *MockMappingRepository) FindPotentialMatches(etcMeisaiID int64, threshold float32) ([]*models.PotentialMatch, error) {
	args := m.Called(etcMeisaiID, threshold)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.PotentialMatch), args.Error(1)
}

// UpdateConfidenceScore updates the confidence score of a mapping
func (m *MockMappingRepository) UpdateConfidenceScore(id int64, confidence float32) error {
	args := m.Called(id, confidence)
	return args.Error(0)
}

// Helper functions for test setup

// SetupMockETCRepositoryForSuccess sets up default successful responses
func SetupMockETCRepositoryForSuccess(mockRepo *MockETCRepository) {
	// Setup default responses for common operations
	mockRepo.On("Create", mock.Anything).Return(nil).Maybe()
	mockRepo.On("Update", mock.Anything).Return(nil).Maybe()
	mockRepo.On("Delete", mock.Anything).Return(nil).Maybe()

	// Setup successful GetByID response
	mockRepo.On("GetByID", mock.Anything).Return(
		&models.ETCMeisai{
			ID:        1,
			Hash:      "test_hash",
			UseDate:   time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			UseTime:   "12:00:00",
			EntryIC:   "東京IC",
			ExitIC:    "横浜IC",
			Amount:    1000,
			CarNumber: "品川500あ1234",
			ETCNumber: "ETC123456",
		},
		nil,
	).Maybe()

	// Setup successful List response
	mockRepo.On("List", mock.Anything).Return(
		[]*models.ETCMeisai{
			{
				ID:      1,
				Hash:    "test_hash_1",
				UseDate: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				Amount:  1000,
			},
			{
				ID:      2,
				Hash:    "test_hash_2",
				UseDate: time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
				Amount:  1500,
			},
		},
		int64(2),
		nil,
	).Maybe()

	// Setup successful GetByDateRange response
	mockRepo.On("GetByDateRange", mock.Anything, mock.Anything).Return(
		[]*models.ETCMeisai{
			{ID: 1, UseDate: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)},
			{ID: 2, UseDate: time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC)},
		},
		nil,
	).Maybe()

	// Setup successful BulkInsert
	mockRepo.On("BulkInsert", mock.Anything).Return(nil).Maybe()

	// Setup successful CheckDuplicatesByHash
	mockRepo.On("CheckDuplicatesByHash", mock.Anything).Return(
		map[string]bool{
			"hash1": false,
			"hash2": false,
		},
		nil,
	).Maybe()

	// Setup successful CountByDateRange
	mockRepo.On("CountByDateRange", mock.Anything, mock.Anything).Return(int64(10), nil).Maybe()

	// Setup successful GetByETCNumber
	mockRepo.On("GetByETCNumber", mock.Anything, mock.Anything).Return(
		[]*models.ETCMeisai{
			{ID: 1, ETCNumber: "ETC123456"},
		},
		nil,
	).Maybe()

	// Setup successful GetByCarNumber
	mockRepo.On("GetByCarNumber", mock.Anything, mock.Anything).Return(
		[]*models.ETCMeisai{
			{ID: 1, CarNumber: "品川500あ1234"},
		},
		nil,
	).Maybe()

	// Setup successful GetSummaryByDateRange
	mockRepo.On("GetSummaryByDateRange", mock.Anything, mock.Anything).Return(
		&models.ETCSummary{
			TotalAmount: 150000,
			TotalCount:  100,
			StartDate:   time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			EndDate:     time.Date(2024, 1, 31, 0, 0, 0, 0, time.UTC),
		},
		nil,
	).Maybe()
}

// SetupMockMappingRepositoryForSuccess sets up default successful responses
func SetupMockMappingRepositoryForSuccess(mockRepo *MockMappingRepository) {
	// Setup default responses for common operations
	mockRepo.On("Create", mock.Anything).Return(nil).Maybe()
	mockRepo.On("Update", mock.Anything).Return(nil).Maybe()
	mockRepo.On("Delete", mock.Anything).Return(nil).Maybe()

	// Setup successful GetByID response
	mockRepo.On("GetByID", mock.Anything).Return(
		&models.ETCMeisaiMapping{
			ID:          1,
			ETCMeisaiID: 100,
			DTakoRowID:  "DTAKO001",
			MappingType: "auto",
			Confidence:  0.95,
		},
		nil,
	).Maybe()

	// Setup successful GetByETCMeisaiID response
	mockRepo.On("GetByETCMeisaiID", mock.Anything).Return(
		[]*models.ETCMeisaiMapping{
			{
				ID:          1,
				ETCMeisaiID: 100,
				DTakoRowID:  "DTAKO001",
			},
		},
		nil,
	).Maybe()

	// Setup successful GetByDTakoRowID response
	mockRepo.On("GetByDTakoRowID", mock.Anything).Return(
		&models.ETCMeisaiMapping{
			ID:          1,
			ETCMeisaiID: 100,
			DTakoRowID:  "DTAKO001",
		},
		nil,
	).Maybe()

	// Setup successful List response
	mockRepo.On("List", mock.Anything).Return(
		[]*models.ETCMeisaiMapping{
			{ID: 1, ETCMeisaiID: 100},
			{ID: 2, ETCMeisaiID: 101},
		},
		int64(2),
		nil,
	).Maybe()

	// Setup successful BulkCreateMappings
	mockRepo.On("BulkCreateMappings", mock.Anything).Return(nil).Maybe()

	// Setup successful DeleteByETCMeisaiID
	mockRepo.On("DeleteByETCMeisaiID", mock.Anything).Return(nil).Maybe()

	// Setup successful FindPotentialMatches
	mockRepo.On("FindPotentialMatches", mock.Anything, mock.Anything).Return(
		[]*models.PotentialMatch{
			{
				DTakoRowID:   "DTAKO001",
				Confidence:   0.95,
				MatchReasons: []string{"ETC番号完全一致"},
			},
		},
		nil,
	).Maybe()

	// Setup successful UpdateConfidenceScore
	mockRepo.On("UpdateConfidenceScore", mock.Anything, mock.Anything).Return(nil).Maybe()
}

// MockRepositoryFactory provides a factory for creating mock repositories
type MockRepositoryFactory struct {
	ETCRepo     *MockETCRepository
	MappingRepo *MockMappingRepository
}

// NewMockRepositoryFactory creates a new factory with initialized mocks
func NewMockRepositoryFactory() *MockRepositoryFactory {
	etcRepo := &MockETCRepository{}
	mappingRepo := &MockMappingRepository{}

	// Setup default successful responses
	SetupMockETCRepositoryForSuccess(etcRepo)
	SetupMockMappingRepositoryForSuccess(mappingRepo)

	return &MockRepositoryFactory{
		ETCRepo:     etcRepo,
		MappingRepo: mappingRepo,
	}
}

// GetETCRepository returns the mock ETC repository
func (f *MockRepositoryFactory) GetETCRepository() *MockETCRepository {
	return f.ETCRepo
}

// GetMappingRepository returns the mock mapping repository
func (f *MockRepositoryFactory) GetMappingRepository() *MockMappingRepository {
	return f.MappingRepo
}
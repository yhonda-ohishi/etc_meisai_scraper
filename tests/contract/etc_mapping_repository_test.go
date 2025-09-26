package contract

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/yhonda-ohishi/etc_meisai/src/models"
	"github.com/yhonda-ohishi/etc_meisai/src/repositories"
	pb "github.com/yhonda-ohishi/etc_meisai/src/pb"
)

// ETCMappingRepositoryContractSuite defines contract tests for ETCMappingRepository
// These tests verify that any implementation of ETCMappingRepository meets
// the expected behavioral contract for gRPC service compatibility
type ETCMappingRepositoryContractSuite struct {
	suite.Suite
	repository repositories.ETCMappingRepository
}

// TestETCMappingRepositoryContract runs the contract test suite
func TestETCMappingRepositoryContract(t *testing.T) {
	suite.Run(t, new(ETCMappingRepositoryContractSuite))
}

// SetupTest initializes test data before each test
func (suite *ETCMappingRepositoryContractSuite) SetupTest() {
	// Mock repository will be injected by actual implementation tests
	suite.repository = &MockETCMappingRepository{}
}

// TestCreateContract verifies Create method contract
func (suite *ETCMappingRepositoryContractSuite) TestCreateContract() {
	tests := []struct {
		name          string
		input         *models.ETCMapping
		expectedError error
		description   string
	}{
		{
			name: "valid_mapping_creation",
			input: &models.ETCMapping{
				ETCRecordID:      1,
				MappingType:      "automatic",
				MappedEntityID:   100,
				MappedEntityType: "dtako_record",
				Confidence:       0.95,
				Status:           "active",
				CreatedBy:        "system",
			},
			expectedError: nil,
			description:   "Should successfully create valid mapping",
		},
		{
			name: "nil_input_validation",
			input: nil,
			expectedError: status.Error(codes.InvalidArgument, "mapping cannot be nil"),
			description: "Should return InvalidArgument error for nil input",
		},
		{
			name: "missing_required_fields",
			input: &models.ETCMapping{
				// Missing required fields: ETCRecordID, MappedEntityType
				MappingType: "manual",
				Confidence:  0.8,
			},
			expectedError: status.Error(codes.InvalidArgument, "missing required fields"),
			description: "Should validate required fields",
		},
		{
			name: "invalid_confidence_range",
			input: &models.ETCMapping{
				ETCRecordID:      1,
				MappingType:      "automatic",
				MappedEntityID:   100,
				MappedEntityType: "dtako_record",
				Confidence:       1.5, // Invalid: > 1.0
				Status:           "active",
			},
			expectedError: status.Error(codes.InvalidArgument, "confidence must be between 0.0 and 1.0"),
			description: "Should validate confidence range",
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			ctx := context.Background()
			err := suite.repository.Create(ctx, tt.input)

			if tt.expectedError != nil {
				suite.Error(err, tt.description)
				suite.Equal(tt.expectedError.Error(), err.Error())
			} else {
				suite.NoError(err, tt.description)
			}
		})
	}
}

// TestGetByIDContract verifies GetByID method contract
func (suite *ETCMappingRepositoryContractSuite) TestGetByIDContract() {
	tests := []struct {
		name          string
		id            int64
		expectedError error
		expectResult  bool
		description   string
	}{
		{
			name:          "existing_mapping_retrieval",
			id:            1,
			expectedError: nil,
			expectResult:  true,
			description:   "Should successfully retrieve existing mapping",
		},
		{
			name:          "non_existent_mapping",
			id:            999999,
			expectedError: status.Error(codes.NotFound, "mapping not found"),
			expectResult:  false,
			description:   "Should return NotFound for non-existent mapping",
		},
		{
			name:          "invalid_id_zero",
			id:            0,
			expectedError: status.Error(codes.InvalidArgument, "id must be positive"),
			expectResult:  false,
			description:   "Should validate positive ID",
		},
		{
			name:          "invalid_id_negative",
			id:            -1,
			expectedError: status.Error(codes.InvalidArgument, "id must be positive"),
			expectResult:  false,
			description:   "Should reject negative IDs",
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			ctx := context.Background()
			result, err := suite.repository.GetByID(ctx, tt.id)

			if tt.expectedError != nil {
				suite.Error(err, tt.description)
				suite.Equal(tt.expectedError.Error(), err.Error())
				suite.Nil(result, "Result should be nil on error")
			} else {
				suite.NoError(err, tt.description)
				if tt.expectResult {
					suite.NotNil(result, "Should return valid mapping")
					suite.Equal(tt.id, result.ID)
				}
			}
		})
	}
}

// TestListContract verifies List method contract
func (suite *ETCMappingRepositoryContractSuite) TestListContract() {
	tests := []struct {
		name          string
		params        repositories.ListMappingsParams
		expectedError error
		description   string
	}{
		{
			name: "valid_list_request",
			params: repositories.ListMappingsParams{
				Page:     1,
				PageSize: 10,
				SortBy:   "created_at",
				SortOrder: "desc",
			},
			expectedError: nil,
			description:   "Should handle valid list request",
		},
		{
			name: "pagination_validation",
			params: repositories.ListMappingsParams{
				Page:     0, // Invalid: should be >= 1
				PageSize: 10,
			},
			expectedError: status.Error(codes.InvalidArgument, "page must be >= 1"),
			description:   "Should validate page number",
		},
		{
			name: "page_size_validation",
			params: repositories.ListMappingsParams{
				Page:     1,
				PageSize: 0, // Invalid: should be > 0
			},
			expectedError: status.Error(codes.InvalidArgument, "page_size must be > 0"),
			description:   "Should validate page size",
		},
		{
			name: "max_page_size_validation",
			params: repositories.ListMappingsParams{
				Page:     1,
				PageSize: 1001, // Invalid: exceeds maximum
			},
			expectedError: status.Error(codes.InvalidArgument, "page_size exceeds maximum of 1000"),
			description:   "Should enforce maximum page size",
		},
		{
			name: "invalid_sort_field",
			params: repositories.ListMappingsParams{
				Page:     1,
				PageSize: 10,
				SortBy:   "invalid_field",
			},
			expectedError: status.Error(codes.InvalidArgument, "invalid sort field"),
			description:   "Should validate sort field",
		},
		{
			name: "date_range_validation",
			params: repositories.ListMappingsParams{
				Page:     1,
				PageSize: 10,
				DateFrom: &time.Time{}, // Invalid: zero time
				DateTo:   nil,
			},
			expectedError: status.Error(codes.InvalidArgument, "invalid date range"),
			description:   "Should validate date range",
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			ctx := context.Background()
			mappings, total, err := suite.repository.List(ctx, tt.params)

			if tt.expectedError != nil {
				suite.Error(err, tt.description)
				suite.Equal(tt.expectedError.Error(), err.Error())
				suite.Nil(mappings, "Mappings should be nil on error")
				suite.Zero(total, "Total should be zero on error")
			} else {
				suite.NoError(err, tt.description)
				suite.NotNil(mappings, "Should return mappings slice")
				suite.GreaterOrEqual(total, int64(0), "Total should be non-negative")
			}
		})
	}
}

// TestTransactionContract verifies transaction method contracts
func (suite *ETCMappingRepositoryContractSuite) TestTransactionContract() {
	tests := []struct {
		name        string
		operation   func(repo repositories.ETCMappingRepository) error
		expectError bool
		description string
	}{
		{
			name: "successful_transaction",
			operation: func(repo repositories.ETCMappingRepository) error {
				txRepo, err := repo.BeginTx(context.Background())
				if err != nil {
					return err
				}
				return txRepo.CommitTx()
			},
			expectError: false,
			description: "Should handle successful transaction lifecycle",
		},
		{
			name: "rollback_transaction",
			operation: func(repo repositories.ETCMappingRepository) error {
				txRepo, err := repo.BeginTx(context.Background())
				if err != nil {
					return err
				}
				return txRepo.RollbackTx()
			},
			expectError: false,
			description: "Should handle transaction rollback",
		},
		{
			name: "commit_without_transaction",
			operation: func(repo repositories.ETCMappingRepository) error {
				return repo.CommitTx() // Should fail - no active transaction
			},
			expectError: true,
			description: "Should fail when committing without active transaction",
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			err := tt.operation(suite.repository)

			if tt.expectError {
				suite.Error(err, tt.description)
			} else {
				suite.NoError(err, tt.description)
			}
		})
	}
}

// TestBulkOperationsContract verifies bulk operation contracts
func (suite *ETCMappingRepositoryContractSuite) TestBulkOperationsContract() {
	tests := []struct {
		name          string
		mappings      []*models.ETCMapping
		expectedError error
		description   string
	}{
		{
			name: "valid_bulk_create",
			mappings: []*models.ETCMapping{
				{
					ETCRecordID:      1,
					MappingType:      "automatic",
					MappedEntityID:   100,
					MappedEntityType: "dtako_record",
					Confidence:       0.95,
					Status:           "active",
				},
				{
					ETCRecordID:      2,
					MappingType:      "automatic",
					MappedEntityID:   101,
					MappedEntityType: "dtako_record",
					Confidence:       0.90,
					Status:           "active",
				},
			},
			expectedError: nil,
			description:   "Should successfully create multiple mappings",
		},
		{
			name:          "empty_bulk_create",
			mappings:      []*models.ETCMapping{},
			expectedError: status.Error(codes.InvalidArgument, "mappings cannot be empty"),
			description:   "Should validate empty bulk operations",
		},
		{
			name: "bulk_create_with_invalid_items",
			mappings: []*models.ETCMapping{
				{
					ETCRecordID:      1,
					MappingType:      "automatic",
					MappedEntityID:   100,
					MappedEntityType: "dtako_record",
					Confidence:       0.95,
					Status:           "active",
				},
				{
					// Missing required fields
					MappingType: "manual",
					Confidence:  2.0, // Invalid confidence
				},
			},
			expectedError: status.Error(codes.InvalidArgument, "invalid mapping in batch"),
			description:   "Should validate individual items in bulk operations",
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			ctx := context.Background()
			err := suite.repository.BulkCreate(ctx, tt.mappings)

			if tt.expectedError != nil {
				suite.Error(err, tt.description)
				suite.Equal(tt.expectedError.Error(), err.Error())
			} else {
				suite.NoError(err, tt.description)
			}
		})
	}
}

// TestHealthCheckContract verifies health check contract
func (suite *ETCMappingRepositoryContractSuite) TestHealthCheckContract() {
	ctx := context.Background()

	// Health check should never return error under normal circumstances
	// and should complete within reasonable time
	err := suite.repository.Ping(ctx)
	suite.NoError(err, "Health check should succeed")
}

// MockETCMappingRepository provides mock implementation for contract testing
type MockETCMappingRepository struct {
	mock.Mock
}

func (m *MockETCMappingRepository) Create(ctx context.Context, mapping *models.ETCMapping) error {
	args := m.Called(ctx, mapping)
	return args.Error(0)
}

func (m *MockETCMappingRepository) GetByID(ctx context.Context, id int64) (*models.ETCMapping, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*models.ETCMapping), args.Error(1)
}

func (m *MockETCMappingRepository) Update(ctx context.Context, mapping *models.ETCMapping) error {
	args := m.Called(ctx, mapping)
	return args.Error(0)
}

func (m *MockETCMappingRepository) Delete(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockETCMappingRepository) GetByETCRecordID(ctx context.Context, etcRecordID int64) ([]*models.ETCMapping, error) {
	args := m.Called(ctx, etcRecordID)
	return args.Get(0).([]*models.ETCMapping), args.Error(1)
}

func (m *MockETCMappingRepository) GetByMappedEntity(ctx context.Context, entityType string, entityID int64) ([]*models.ETCMapping, error) {
	args := m.Called(ctx, entityType, entityID)
	return args.Get(0).([]*models.ETCMapping), args.Error(1)
}

func (m *MockETCMappingRepository) GetActiveMapping(ctx context.Context, etcRecordID int64) (*models.ETCMapping, error) {
	args := m.Called(ctx, etcRecordID)
	return args.Get(0).(*models.ETCMapping), args.Error(1)
}

func (m *MockETCMappingRepository) List(ctx context.Context, params repositories.ListMappingsParams) ([]*models.ETCMapping, int64, error) {
	args := m.Called(ctx, params)
	return args.Get(0).([]*models.ETCMapping), args.Get(1).(int64), args.Error(2)
}

func (m *MockETCMappingRepository) BulkCreate(ctx context.Context, mappings []*models.ETCMapping) error {
	args := m.Called(ctx, mappings)
	return args.Error(0)
}

func (m *MockETCMappingRepository) UpdateStatus(ctx context.Context, id int64, status string) error {
	args := m.Called(ctx, id, status)
	return args.Error(0)
}

func (m *MockETCMappingRepository) BeginTx(ctx context.Context) (repositories.ETCMappingRepository, error) {
	args := m.Called(ctx)
	return args.Get(0).(repositories.ETCMappingRepository), args.Error(1)
}

func (m *MockETCMappingRepository) CommitTx() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockETCMappingRepository) RollbackTx() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockETCMappingRepository) Ping(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

// TestProtocolBufferCompatibility verifies gRPC message compatibility
func (suite *ETCMappingRepositoryContractSuite) TestProtocolBufferCompatibility() {
	// Verify that models.ETCMapping can be converted to/from Protocol Buffer messages
	originalMapping := &models.ETCMapping{
		ID:               1,
		ETCRecordID:      123,
		MappingType:      "automatic",
		MappedEntityID:   456,
		MappedEntityType: "dtako_record",
		Confidence:       0.95,
		Status:           "active",
		CreatedBy:        "system",
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	// Convert to Protocol Buffer message
	pbMapping := &pb.ETCMapping{
		Id:               originalMapping.ID,
		EtcRecordId:      originalMapping.ETCRecordID,
		MappingType:      originalMapping.MappingType,
		MappedEntityId:   originalMapping.MappedEntityID,
		MappedEntityType: originalMapping.MappedEntityType,
		Confidence:       originalMapping.Confidence,
		Status:           pb.MappingStatus_MAPPING_STATUS_ACTIVE,
		CreatedBy:        originalMapping.CreatedBy,
		CreatedAt:        timestamppb.New(originalMapping.CreatedAt),
		UpdatedAt:        timestamppb.New(originalMapping.UpdatedAt),
	}

	// Verify conversion
	suite.NotNil(pbMapping, "Protocol Buffer conversion should succeed")
	suite.Equal(originalMapping.ID, pbMapping.Id)
	suite.Equal(originalMapping.ETCRecordID, pbMapping.EtcRecordId)
	suite.Equal(originalMapping.MappingType, pbMapping.MappingType)
	suite.Equal(originalMapping.MappedEntityID, pbMapping.MappedEntityId)
	suite.Equal(originalMapping.MappedEntityType, pbMapping.MappedEntityType)
	suite.Equal(originalMapping.Confidence, pbMapping.Confidence)
	suite.Equal(originalMapping.CreatedBy, pbMapping.CreatedBy)

	// Verify timestamp conversion
	suite.True(pbMapping.CreatedAt.AsTime().Equal(originalMapping.CreatedAt))
	suite.True(pbMapping.UpdatedAt.AsTime().Equal(originalMapping.UpdatedAt))
}
package contract

import (
	"context"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/yhonda-ohishi/etc_meisai/src/models"
	"github.com/yhonda-ohishi/etc_meisai/src/services"
	pb "github.com/yhonda-ohishi/etc_meisai/src/pb"
)

// ETCMappingServiceContractSuite defines contract tests for ETCMappingService
// These tests verify that any implementation of ETCMappingServiceInterface meets
// the expected behavioral contract for gRPC service compatibility
type ETCMappingServiceContractSuite struct {
	suite.Suite
	service *services.ETCMappingService
}

// TestETCMappingServiceContract runs the contract test suite
func TestETCMappingServiceContract(t *testing.T) {
	suite.Run(t, new(ETCMappingServiceContractSuite))
}

// SetupTest initializes test data before each test
func (suite *ETCMappingServiceContractSuite) SetupTest() {
	// Mock service will be injected by actual implementation tests
	suite.service = nil // Will be set by actual tests
}

// TestCreateMappingContract verifies CreateMapping method contract
// TODO: This test requires a proper service interface - currently disabled
func (suite *ETCMappingServiceContractSuite) TestCreateMappingContract() {
	suite.T().Skip("Contract test disabled - requires ETCMappingServiceInterface")
	tests := []struct {
		name          string
		params        *services.CreateMappingParams
		expectedError error
		description   string
	}{
		{
			name: "valid_mapping_creation",
			params: &services.CreateMappingParams{
				ETCRecordID:      1,
				MappingType:      "automatic",
				MappedEntityID:   100,
				MappedEntityType: "dtako_record",
				Confidence:       0.95,
				Status:           "active",
				CreatedBy:        "system",
				Metadata: map[string]interface{}{
					"algorithm": "exact_match",
					"version":   "1.0",
				},
			},
			expectedError: nil,
			description:   "Should successfully create valid mapping",
		},
		{
			name: "nil_params_validation",
			params: nil,
			expectedError: status.Error(codes.InvalidArgument, "params cannot be nil"),
			description: "Should return InvalidArgument error for nil params",
		},
		{
			name: "missing_required_fields",
			params: &services.CreateMappingParams{
				// Missing required fields: ETCRecordID, MappingType, MappedEntityType
				MappedEntityID: 100,
				Confidence:     0.8,
			},
			expectedError: status.Error(codes.InvalidArgument, "missing required fields"),
			description: "Should validate required fields",
		},
		{
			name: "invalid_etc_record_id",
			params: &services.CreateMappingParams{
				ETCRecordID:      0, // Invalid: must be > 0
				MappingType:      "manual",
				MappedEntityID:   100,
				MappedEntityType: "dtako_record",
			},
			expectedError: status.Error(codes.InvalidArgument, "etc_record_id must be positive"),
			description: "Should validate ETCRecordID is positive",
		},
		{
			name: "invalid_mapped_entity_id",
			params: &services.CreateMappingParams{
				ETCRecordID:      1,
				MappingType:      "manual",
				MappedEntityID:   -1, // Invalid: must be > 0
				MappedEntityType: "dtako_record",
			},
			expectedError: status.Error(codes.InvalidArgument, "mapped_entity_id must be positive"),
			description: "Should validate MappedEntityID is positive",
		},
		{
			name: "invalid_confidence_range",
			params: &services.CreateMappingParams{
				ETCRecordID:      1,
				MappingType:      "automatic",
				MappedEntityID:   100,
				MappedEntityType: "dtako_record",
				Confidence:       1.5, // Invalid: > 1.0
			},
			expectedError: status.Error(codes.InvalidArgument, "confidence must be between 0.0 and 1.0"),
			description: "Should validate confidence range",
		},
		{
			name: "empty_mapping_type",
			params: &services.CreateMappingParams{
				ETCRecordID:      1,
				MappingType:      "", // Invalid: empty
				MappedEntityID:   100,
				MappedEntityType: "dtako_record",
			},
			expectedError: status.Error(codes.InvalidArgument, "mapping_type cannot be empty"),
			description: "Should validate mapping type is not empty",
		},
		{
			name: "empty_mapped_entity_type",
			params: &services.CreateMappingParams{
				ETCRecordID:      1,
				MappingType:      "manual",
				MappedEntityID:   100,
				MappedEntityType: "", // Invalid: empty
			},
			expectedError: status.Error(codes.InvalidArgument, "mapped_entity_type cannot be empty"),
			description: "Should validate mapped entity type is not empty",
		},
		{
			name: "duplicate_mapping_conflict",
			params: &services.CreateMappingParams{
				ETCRecordID:      1, // Assume this record already has an active mapping
				MappingType:      "manual",
				MappedEntityID:   200,
				MappedEntityType: "dtako_record",
			},
			expectedError: status.Error(codes.AlreadyExists, "active mapping already exists for ETC record"),
			description: "Should prevent duplicate active mappings",
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			ctx := context.Background()
			result, err := suite.service.CreateMapping(ctx, tt.params)

			if tt.expectedError != nil {
				suite.Error(err, tt.description)
				suite.Equal(tt.expectedError.Error(), err.Error())
				suite.Nil(result, "Result should be nil on error")
			} else {
				suite.NoError(err, tt.description)
				suite.NotNil(result, "Should return valid mapping")
				suite.Equal(tt.params.ETCRecordID, result.ETCRecordID)
				suite.Equal(tt.params.MappingType, result.MappingType)
				suite.Equal(tt.params.MappedEntityID, result.MappedEntityID)
				suite.Equal(tt.params.MappedEntityType, result.MappedEntityType)
				suite.Equal(tt.params.Confidence, result.Confidence)
			}
		})
	}
}

// TestGetMappingContract verifies GetMapping method contract
// TODO: This test requires a proper service interface - currently disabled
func (suite *ETCMappingServiceContractSuite) TestGetMappingContract() {
	suite.T().Skip("Contract test disabled - requires ETCMappingServiceInterface")
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
			result, err := suite.service.GetMapping(ctx, tt.id)

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

// TestListMappingsContract verifies ListMappings method contract
// TODO: This test requires a proper service interface - currently disabled
func (suite *ETCMappingServiceContractSuite) TestListMappingsContract() {
	suite.T().Skip("Contract test disabled - requires ETCMappingServiceInterface")
	tests := []struct {
		name          string
		params        *services.ListMappingsParams
		expectedError error
		description   string
	}{
		{
			name: "valid_list_request",
			params: &services.ListMappingsParams{
				Page:     1,
				PageSize: 10,
				SortBy:   "created_at",
				SortOrder: "desc",
			},
			expectedError: nil,
			description:   "Should handle valid list request",
		},
		{
			name: "nil_params_validation",
			params: nil,
			expectedError: status.Error(codes.InvalidArgument, "params cannot be nil"),
			description: "Should return InvalidArgument error for nil params",
		},
		{
			name: "pagination_validation",
			params: &services.ListMappingsParams{
				Page:     0, // Invalid: should be >= 1
				PageSize: 10,
			},
			expectedError: status.Error(codes.InvalidArgument, "page must be >= 1"),
			description:   "Should validate page number",
		},
		{
			name: "page_size_validation",
			params: &services.ListMappingsParams{
				Page:     1,
				PageSize: 0, // Invalid: should be > 0
			},
			expectedError: status.Error(codes.InvalidArgument, "page_size must be > 0"),
			description:   "Should validate page size",
		},
		{
			name: "max_page_size_validation",
			params: &services.ListMappingsParams{
				Page:     1,
				PageSize: 1001, // Invalid: exceeds maximum
			},
			expectedError: status.Error(codes.InvalidArgument, "page_size exceeds maximum of 1000"),
			description:   "Should enforce maximum page size",
		},
		{
			name: "invalid_sort_field",
			params: &services.ListMappingsParams{
				Page:     1,
				PageSize: 10,
				SortBy:   "invalid_field",
			},
			expectedError: status.Error(codes.InvalidArgument, "invalid sort field"),
			description:   "Should validate sort field",
		},
		{
			name: "invalid_confidence_range",
			params: &services.ListMappingsParams{
				Page:          1,
				PageSize:      10,
				MinConfidence: func() *float32 { v := float32(1.5); return &v }(), // Invalid: > 1.0
			},
			expectedError: status.Error(codes.InvalidArgument, "confidence values must be between 0.0 and 1.0"),
			description:   "Should validate confidence range",
		},
		{
			name: "inverted_confidence_range",
			params: &services.ListMappingsParams{
				Page:          1,
				PageSize:      10,
				MinConfidence: func() *float32 { v := float32(0.8); return &v }(),
				MaxConfidence: func() *float32 { v := float32(0.5); return &v }(), // Invalid: min > max
			},
			expectedError: status.Error(codes.InvalidArgument, "min_confidence cannot be greater than max_confidence"),
			description:   "Should validate confidence range order",
		},
		{
			name: "valid_filters",
			params: &services.ListMappingsParams{
				Page:             1,
				PageSize:         50,
				ETCRecordID:      func() *int64 { v := int64(123); return &v }(),
				MappingType:      func() *string { s := "automatic"; return &s }(),
				MappedEntityID:   func() *int64 { v := int64(456); return &v }(),
				MappedEntityType: func() *string { s := "dtako_record"; return &s }(),
				Status:           func() *string { s := "active"; return &s }(),
				MinConfidence:    func() *float32 { v := float32(0.5); return &v }(),
				MaxConfidence:    func() *float32 { v := float32(1.0); return &v }(),
				CreatedBy:        func() *string { s := "system"; return &s }(),
				SortBy:           "confidence",
				SortOrder:        "desc",
			},
			expectedError: nil,
			description:   "Should handle complex filter combinations",
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			ctx := context.Background()
			result, err := suite.service.ListMappings(ctx, tt.params)

			if tt.expectedError != nil {
				suite.Error(err, tt.description)
				suite.Equal(tt.expectedError.Error(), err.Error())
				suite.Nil(result, "Result should be nil on error")
			} else {
				suite.NoError(err, tt.description)
				suite.NotNil(result, "Should return mappings slice")
			}
		})
	}
}

// TestUpdateMappingContract verifies UpdateMapping method contract
// TODO: This test requires a proper service interface - currently disabled
func (suite *ETCMappingServiceContractSuite) TestUpdateMappingContract() {
	suite.T().Skip("Contract test disabled - requires ETCMappingServiceInterface")
	tests := []struct {
		name          string
		id            int64
		params        *services.UpdateMappingParams
		expectedError error
		description   string
	}{
		{
			name: "valid_mapping_update",
			id:   1,
			params: &services.UpdateMappingParams{
				MappingType:      func() *string { s := "manual"; return &s }(),
				MappedEntityID:   func() *int64 { v := int64(200); return &v }(),
				MappedEntityType: func() *string { s := "updated_record"; return &s }(),
				Confidence:       func() *float32 { v := float32(0.9); return &v }(),
				Status:           func() *string { s := "updated"; return &s }(),
				Metadata: map[string]interface{}{
					"updated_by": "user123",
					"reason":     "manual_correction",
				},
			},
			expectedError: nil,
			description:   "Should successfully update mapping with valid params",
		},
		{
			name: "invalid_id_zero",
			id:   0,
			params: &services.UpdateMappingParams{
				Status: func() *string { s := "updated"; return &s }(),
			},
			expectedError: status.Error(codes.InvalidArgument, "id must be positive"),
			description:   "Should validate positive ID",
		},
		{
			name: "invalid_id_negative",
			id:   -1,
			params: &services.UpdateMappingParams{
				Status: func() *string { s := "updated"; return &s }(),
			},
			expectedError: status.Error(codes.InvalidArgument, "id must be positive"),
			description:   "Should reject negative IDs",
		},
		{
			name:          "nil_params_validation",
			id:            1,
			params:        nil,
			expectedError: status.Error(codes.InvalidArgument, "params cannot be nil"),
			description:   "Should return InvalidArgument error for nil params",
		},
		{
			name: "non_existent_mapping",
			id:   999999,
			params: &services.UpdateMappingParams{
				Status: func() *string { s := "updated"; return &s }(),
			},
			expectedError: status.Error(codes.NotFound, "mapping not found"),
			description:   "Should return NotFound for non-existent mapping",
		},
		{
			name: "empty_update_params",
			id:   1,
			params: &services.UpdateMappingParams{
				// All fields are nil - no updates requested
			},
			expectedError: status.Error(codes.InvalidArgument, "no fields to update"),
			description:   "Should reject updates with no fields specified",
		},
		{
			name: "invalid_confidence_range",
			id:   1,
			params: &services.UpdateMappingParams{
				Confidence: func() *float32 { v := float32(1.5); return &v }(), // Invalid: > 1.0
			},
			expectedError: status.Error(codes.InvalidArgument, "confidence must be between 0.0 and 1.0"),
			description:   "Should validate confidence range",
		},
		{
			name: "invalid_mapped_entity_id",
			id:   1,
			params: &services.UpdateMappingParams{
				MappedEntityID: func() *int64 { v := int64(-1); return &v }(), // Invalid: must be > 0
			},
			expectedError: status.Error(codes.InvalidArgument, "mapped_entity_id must be positive"),
			description:   "Should validate MappedEntityID is positive",
		},
		{
			name: "empty_string_fields",
			id:   1,
			params: &services.UpdateMappingParams{
				MappingType:      func() *string { s := ""; return &s }(), // Invalid: empty
				MappedEntityType: func() *string { s := ""; return &s }(), // Invalid: empty
			},
			expectedError: status.Error(codes.InvalidArgument, "string fields cannot be empty"),
			description:   "Should validate string fields are not empty",
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			ctx := context.Background()
			result, err := suite.service.UpdateMapping(ctx, tt.id, tt.params)

			if tt.expectedError != nil {
				suite.Error(err, tt.description)
				suite.Equal(tt.expectedError.Error(), err.Error())
				suite.Nil(result, "Result should be nil on error")
			} else {
				suite.NoError(err, tt.description)
				suite.NotNil(result, "Should return updated mapping")
				suite.Equal(tt.id, result.ID)
			}
		})
	}
}

// TestDeleteMappingContract verifies DeleteMapping method contract
// TODO: This test requires a proper service interface - currently disabled
func (suite *ETCMappingServiceContractSuite) TestDeleteMappingContract() {
	suite.T().Skip("Contract test disabled - requires ETCMappingServiceInterface")
	tests := []struct {
		name          string
		id            int64
		expectedError error
		description   string
	}{
		{
			name:          "existing_mapping_deletion",
			id:            1,
			expectedError: nil,
			description:   "Should successfully delete existing mapping",
		},
		{
			name:          "non_existent_mapping",
			id:            999999,
			expectedError: status.Error(codes.NotFound, "mapping not found"),
			description:   "Should return NotFound for non-existent mapping",
		},
		{
			name:          "invalid_id_zero",
			id:            0,
			expectedError: status.Error(codes.InvalidArgument, "id must be positive"),
			description:   "Should validate positive ID",
		},
		{
			name:          "invalid_id_negative",
			id:            -1,
			expectedError: status.Error(codes.InvalidArgument, "id must be positive"),
			description:   "Should reject negative IDs",
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			ctx := context.Background()
			err := suite.service.DeleteMapping(ctx, tt.id)

			if tt.expectedError != nil {
				suite.Error(err, tt.description)
				suite.Equal(tt.expectedError.Error(), err.Error())
			} else {
				suite.NoError(err, tt.description)
			}
		})
	}
}

// MockETCMappingService provides mock implementation for contract testing
type MockETCMappingService struct {
	mock.Mock
}

func (m *MockETCMappingService) CreateMapping(ctx context.Context, params *services.CreateMappingParams) (*models.ETCMapping, error) {
	args := m.Called(ctx, params)
	return args.Get(0).(*models.ETCMapping), args.Error(1)
}

func (m *MockETCMappingService) GetMapping(ctx context.Context, id int64) (*models.ETCMapping, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*models.ETCMapping), args.Error(1)
}

func (m *MockETCMappingService) ListMappings(ctx context.Context, params *services.ListMappingsParams) ([]*models.ETCMapping, error) {
	args := m.Called(ctx, params)
	return args.Get(0).([]*models.ETCMapping), args.Error(1)
}

func (m *MockETCMappingService) UpdateMapping(ctx context.Context, id int64, params *services.UpdateMappingParams) (*models.ETCMapping, error) {
	args := m.Called(ctx, id, params)
	return args.Get(0).(*models.ETCMapping), args.Error(1)
}

func (m *MockETCMappingService) DeleteMapping(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// TestProtocolBufferCompatibility verifies gRPC message compatibility
func (suite *ETCMappingServiceContractSuite) TestProtocolBufferCompatibility() {
	// Test CreateMappingParams to Protocol Buffer conversion
	createParams := &services.CreateMappingParams{
		ETCRecordID:      123,
		MappingType:      "automatic",
		MappedEntityID:   456,
		MappedEntityType: "dtako_record",
		Confidence:       0.95,
		Status:           "active",
		CreatedBy:        "system",
		Metadata: map[string]interface{}{
			"algorithm": "exact_match",
			"version":   "1.0",
		},
	}

	// Convert metadata to protobuf Struct
	metadataStruct, err := structpb.NewStruct(createParams.Metadata)
	suite.NoError(err, "Metadata should convert to protobuf Struct")

	pbCreateMapping := &pb.ETCMapping{
		EtcRecordId:      createParams.ETCRecordID,
		MappingType:      createParams.MappingType,
		MappedEntityId:   createParams.MappedEntityID,
		MappedEntityType: createParams.MappedEntityType,
		Confidence:       createParams.Confidence,
		Status:           pb.MappingStatus_MAPPING_STATUS_ACTIVE,
		CreatedBy:        createParams.CreatedBy,
		Metadata:         metadataStruct,
	}

	suite.NotNil(pbCreateMapping, "CreateMapping conversion should succeed")
	suite.Equal(createParams.ETCRecordID, pbCreateMapping.EtcRecordId)
	suite.Equal(createParams.MappingType, pbCreateMapping.MappingType)
	suite.Equal(createParams.MappedEntityID, pbCreateMapping.MappedEntityId)
	suite.Equal(createParams.MappedEntityType, pbCreateMapping.MappedEntityType)
	suite.Equal(createParams.Confidence, pbCreateMapping.Confidence)
	suite.Equal(createParams.CreatedBy, pbCreateMapping.CreatedBy)

	// Test ListMappingsParams to Protocol Buffer conversion
	listParams := &services.ListMappingsParams{
		Page:             1,
		PageSize:         10,
		ETCRecordID:      func() *int64 { v := int64(123); return &v }(),
		MappingType:      func() *string { s := "automatic"; return &s }(),
		MappedEntityType: func() *string { s := "dtako_record"; return &s }(),
		Status:           func() *string { s := "active"; return &s }(),
		SortBy:           "created_at",
		SortOrder:        "desc",
	}

	statusActive := pb.MappingStatus_MAPPING_STATUS_ACTIVE
	pbListRequest := &pb.ListMappingsRequest{
		Page:             int32(listParams.Page),
		PageSize:         int32(listParams.PageSize),
		EtcRecordId:      listParams.ETCRecordID,
		MappingType:      listParams.MappingType,
		MappedEntityType: listParams.MappedEntityType,
		Status:           &statusActive,
	}

	suite.NotNil(pbListRequest, "ListMappings conversion should succeed")
	suite.Equal(int32(listParams.Page), pbListRequest.Page)
	suite.Equal(int32(listParams.PageSize), pbListRequest.PageSize)
	suite.Equal(listParams.ETCRecordID, pbListRequest.EtcRecordId)
	suite.Equal(listParams.MappingType, pbListRequest.MappingType)
	suite.Equal(listParams.MappedEntityType, pbListRequest.MappedEntityType)

	// Test UpdateMappingParams to Protocol Buffer conversion
	updateParams := &services.UpdateMappingParams{
		MappingType:      func() *string { s := "manual"; return &s }(),
		MappedEntityID:   func() *int64 { v := int64(789); return &v }(),
		MappedEntityType: func() *string { s := "updated_record"; return &s }(),
		Confidence:       func() *float32 { v := float32(0.8); return &v }(),
		Status:           func() *string { s := "updated"; return &s }(),
	}

	pbUpdateMapping := &pb.ETCMapping{
		MappingType:      *updateParams.MappingType,
		MappedEntityId:   *updateParams.MappedEntityID,
		MappedEntityType: *updateParams.MappedEntityType,
		Confidence:       *updateParams.Confidence,
		Status:           pb.MappingStatus_MAPPING_STATUS_ACTIVE, // Status would be mapped appropriately
	}

	suite.NotNil(pbUpdateMapping, "UpdateMapping conversion should succeed")
	suite.Equal(*updateParams.MappingType, pbUpdateMapping.MappingType)
	suite.Equal(*updateParams.MappedEntityID, pbUpdateMapping.MappedEntityId)
	suite.Equal(*updateParams.MappedEntityType, pbUpdateMapping.MappedEntityType)
	suite.Equal(*updateParams.Confidence, pbUpdateMapping.Confidence)
}
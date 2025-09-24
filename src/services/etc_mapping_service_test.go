package services

import (
	"context"
	"errors"
	"log"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/yhonda-ohishi/etc_meisai/src/mocks"
	"github.com/yhonda-ohishi/etc_meisai/src/models"
	"github.com/yhonda-ohishi/etc_meisai/src/repositories"
)

func TestNewETCMappingService(t *testing.T) {
	t.Parallel()

	t.Run("with repositories and logger", func(t *testing.T) {
		mockMappingRepo := &mocks.MockETCMappingRepository{}
		mockRecordRepo := &mocks.MockETCMeisaiRecordRepository{}
		logger := log.New(os.Stdout, "test", log.LstdFlags)

		service := NewETCMappingService(mockMappingRepo, mockRecordRepo, logger)

		assert.NotNil(t, service)
		assert.Equal(t, mockMappingRepo, service.mappingRepo)
		assert.Equal(t, mockRecordRepo, service.recordRepo)
		assert.Equal(t, logger, service.logger)
	})

	t.Run("with repositories, no logger", func(t *testing.T) {
		mockMappingRepo := &mocks.MockETCMappingRepository{}
		mockRecordRepo := &mocks.MockETCMeisaiRecordRepository{}

		service := NewETCMappingService(mockMappingRepo, mockRecordRepo, nil)

		assert.NotNil(t, service)
		assert.Equal(t, mockMappingRepo, service.mappingRepo)
		assert.Equal(t, mockRecordRepo, service.recordRepo)
		assert.NotNil(t, service.logger)
	})
}

func TestETCMappingService_CreateMapping(t *testing.T) {
	t.Parallel()

	validParams := &CreateMappingParams{
		ETCRecordID:      1,
		MappingType:      "automatic",
		MappedEntityID:   100,
		MappedEntityType: "dtako_record",
		Confidence:       0.95,
		Status:           string(models.MappingStatusActive),
		CreatedBy:        "system",
		Metadata: map[string]interface{}{
			"match_score": 0.95,
			"algorithm":   "fuzzy_match",
		},
	}

	existingRecord := &models.ETCMeisaiRecord{
		ID: 1,
	}

	tests := []struct {
		name        string
		params      *CreateMappingParams
		setupMock   func(*mocks.MockETCMappingRepository, *mocks.MockETCMeisaiRecordRepository)
		expectError bool
		errorMsg    string
	}{
		{
			name:   "successful creation",
			params: validParams,
			setupMock: func(mappingRepo *mocks.MockETCMappingRepository, recordRepo *mocks.MockETCMeisaiRecordRepository) {
				txMock := &mocks.MockETCMappingRepository{}
				mappingRepo.On("BeginTx", mock.Anything).Return(txMock, nil)
				recordRepo.On("GetByID", mock.Anything, int64(1)).Return(existingRecord, nil)
				txMock.On("GetActiveMapping", mock.Anything, int64(1)).Return(nil, errors.New("not found"))
				txMock.On("Create", mock.Anything, mock.AnythingOfType("*models.ETCMapping")).Return(nil)
				txMock.On("CommitTx").Return(nil)
				txMock.On("RollbackTx").Return(nil).Maybe()
			},
			expectError: false,
		},
		{
			name: "with defaults",
			params: &CreateMappingParams{
				ETCRecordID:      1,
				MappingType:      "automatic",
				MappedEntityID:   100,
				MappedEntityType: "dtako_record",
				// Confidence and Status will use defaults
			},
			setupMock: func(mappingRepo *mocks.MockETCMappingRepository, recordRepo *mocks.MockETCMeisaiRecordRepository) {
				txMock := &mocks.MockETCMappingRepository{}
				mappingRepo.On("BeginTx", mock.Anything).Return(txMock, nil)
				recordRepo.On("GetByID", mock.Anything, int64(1)).Return(existingRecord, nil)
				txMock.On("GetActiveMapping", mock.Anything, int64(1)).Return(nil, errors.New("not found"))
				txMock.On("Create", mock.Anything, mock.AnythingOfType("*models.ETCMapping")).Return(nil)
				txMock.On("CommitTx").Return(nil)
				txMock.On("RollbackTx").Return(nil).Maybe()
			},
			expectError: false,
		},
		{
			name: "metadata error",
			params: &CreateMappingParams{
				ETCRecordID:      1,
				MappingType:      "automatic",
				MappedEntityID:   100,
				MappedEntityType: "dtako_record",
				Confidence:       0.95,
				Metadata: map[string]interface{}{
					"invalid": make(chan int), // Cannot be marshaled to JSON
				},
			},
			setupMock:   func(mappingRepo *mocks.MockETCMappingRepository, recordRepo *mocks.MockETCMeisaiRecordRepository) {},
			expectError: true,
			errorMsg:    "failed to set metadata",
		},
		{
			name:   "begin transaction error",
			params: validParams,
			setupMock: func(mappingRepo *mocks.MockETCMappingRepository, recordRepo *mocks.MockETCMeisaiRecordRepository) {
				mappingRepo.On("BeginTx", mock.Anything).Return(nil, errors.New("transaction failed"))
			},
			expectError: true,
			errorMsg:    "failed to start transaction",
		},
		{
			name:   "ETC record not found",
			params: validParams,
			setupMock: func(mappingRepo *mocks.MockETCMappingRepository, recordRepo *mocks.MockETCMeisaiRecordRepository) {
				txMock := &mocks.MockETCMappingRepository{}
				mappingRepo.On("BeginTx", mock.Anything).Return(txMock, nil)
				recordRepo.On("GetByID", mock.Anything, int64(1)).Return(nil, errors.New("record not found"))
				txMock.On("RollbackTx").Return(nil)
			},
			expectError: true,
			errorMsg:    "ETC record not found with ID 1",
		},
		{
			name:   "active mapping already exists",
			params: validParams,
			setupMock: func(mappingRepo *mocks.MockETCMappingRepository, recordRepo *mocks.MockETCMeisaiRecordRepository) {
				txMock := &mocks.MockETCMappingRepository{}
				existingMapping := &models.ETCMapping{ID: 999}
				mappingRepo.On("BeginTx", mock.Anything).Return(txMock, nil)
				recordRepo.On("GetByID", mock.Anything, int64(1)).Return(existingRecord, nil)
				txMock.On("GetActiveMapping", mock.Anything, int64(1)).Return(existingMapping, nil)
				txMock.On("RollbackTx").Return(nil)
			},
			expectError: true,
			errorMsg:    "active mapping already exists for ETC record 1",
		},
		{
			name:   "create mapping error",
			params: validParams,
			setupMock: func(mappingRepo *mocks.MockETCMappingRepository, recordRepo *mocks.MockETCMeisaiRecordRepository) {
				txMock := &mocks.MockETCMappingRepository{}
				mappingRepo.On("BeginTx", mock.Anything).Return(txMock, nil)
				recordRepo.On("GetByID", mock.Anything, int64(1)).Return(existingRecord, nil)
				txMock.On("GetActiveMapping", mock.Anything, int64(1)).Return(nil, errors.New("not found"))
				txMock.On("Create", mock.Anything, mock.AnythingOfType("*models.ETCMapping")).Return(errors.New("create failed"))
				txMock.On("RollbackTx").Return(nil)
			},
			expectError: true,
			errorMsg:    "failed to create mapping",
		},
		{
			name:   "commit error",
			params: validParams,
			setupMock: func(mappingRepo *mocks.MockETCMappingRepository, recordRepo *mocks.MockETCMeisaiRecordRepository) {
				txMock := &mocks.MockETCMappingRepository{}
				mappingRepo.On("BeginTx", mock.Anything).Return(txMock, nil)
				recordRepo.On("GetByID", mock.Anything, int64(1)).Return(existingRecord, nil)
				txMock.On("GetActiveMapping", mock.Anything, int64(1)).Return(nil, errors.New("not found"))
				txMock.On("Create", mock.Anything, mock.AnythingOfType("*models.ETCMapping")).Return(nil)
				txMock.On("CommitTx").Return(errors.New("commit failed"))
			},
			expectError: true,
			errorMsg:    "failed to commit transaction",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockMappingRepo := &mocks.MockETCMappingRepository{}
			mockRecordRepo := &mocks.MockETCMeisaiRecordRepository{}
			tt.setupMock(mockMappingRepo, mockRecordRepo)

			service := NewETCMappingService(mockMappingRepo, mockRecordRepo, nil)
			ctx := context.Background()

			mapping, err := service.CreateMapping(ctx, tt.params)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
				assert.Nil(t, mapping)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, mapping)
			}

			mockMappingRepo.AssertExpectations(t)
			mockRecordRepo.AssertExpectations(t)
		})
	}
}

func TestETCMappingService_CreateMapping_WithPanic(t *testing.T) {
	t.Parallel()

	t.Run("panic during transaction", func(t *testing.T) {
		mockMappingRepo := &mocks.MockETCMappingRepository{}
		mockRecordRepo := &mocks.MockETCMeisaiRecordRepository{}
		txMock := &mocks.MockETCMappingRepository{}

		mockMappingRepo.On("BeginTx", mock.Anything).Return(txMock, nil)
		mockRecordRepo.On("GetByID", mock.Anything, int64(1)).Run(func(args mock.Arguments) {
			panic("simulated panic")
		}).Return(nil, nil)
		txMock.On("RollbackTx").Return(nil)

		service := NewETCMappingService(mockMappingRepo, mockRecordRepo, nil)
		ctx := context.Background()

		params := &CreateMappingParams{
			ETCRecordID:      1,
			MappingType:      "automatic",
			MappedEntityID:   100,
			MappedEntityType: "dtako_record",
		}

		assert.Panics(t, func() {
			service.CreateMapping(ctx, params)
		})

		txMock.AssertExpectations(t)
	})
}

func TestETCMappingService_GetMapping(t *testing.T) {
	t.Parallel()

	expectedMapping := &models.ETCMapping{
		ID:               1,
		ETCRecordID:      100,
		MappingType:      "automatic",
		MappedEntityID:   200,
		MappedEntityType: "dtako_record",
		Confidence:       0.95,
		Status:           string(models.MappingStatusActive),
	}

	tests := []struct {
		name        string
		id          int64
		setupMock   func(*mocks.MockETCMappingRepository)
		expectError bool
		errorMsg    string
	}{
		{
			name: "successful retrieval",
			id:   1,
			setupMock: func(m *mocks.MockETCMappingRepository) {
				m.On("GetByID", mock.Anything, int64(1)).Return(expectedMapping, nil)
			},
			expectError: false,
		},
		{
			name:        "invalid ID - zero",
			id:          0,
			setupMock:   func(m *mocks.MockETCMappingRepository) {},
			expectError: true,
			errorMsg:    "invalid mapping ID",
		},
		{
			name:        "invalid ID - negative",
			id:          -1,
			setupMock:   func(m *mocks.MockETCMappingRepository) {},
			expectError: true,
			errorMsg:    "invalid mapping ID",
		},
		{
			name: "repository error",
			id:   999,
			setupMock: func(m *mocks.MockETCMappingRepository) {
				m.On("GetByID", mock.Anything, int64(999)).Return(nil, errors.New("mapping not found"))
			},
			expectError: true,
			errorMsg:    "failed to retrieve mapping",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockMappingRepo := &mocks.MockETCMappingRepository{}
			mockRecordRepo := &mocks.MockETCMeisaiRecordRepository{}
			tt.setupMock(mockMappingRepo)

			service := NewETCMappingService(mockMappingRepo, mockRecordRepo, nil)
			ctx := context.Background()

			mapping, err := service.GetMapping(ctx, tt.id)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
				assert.Nil(t, mapping)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, expectedMapping, mapping)
			}

			mockMappingRepo.AssertExpectations(t)
		})
	}
}

func TestETCMappingService_ListMappings(t *testing.T) {
	t.Parallel()

	mappingType := "automatic"
	status := string(models.MappingStatusActive)
	minConfidence := float32(0.8)
	maxConfidence := float32(1.0)
	mappedEntityType := "dtako_record"
	createdBy := "system"

	expectedMappings := []*models.ETCMapping{
		{ID: 1, MappingType: "automatic"},
		{ID: 2, MappingType: "manual"},
	}

	tests := []struct {
		name            string
		params          *ListMappingsParams
		setupMock       func(*mocks.MockETCMappingRepository)
		expectError     bool
		expectMappings  []*models.ETCMapping
		expectTotal     int64
	}{
		{
			name:   "default parameters",
			params: &ListMappingsParams{},
			setupMock: func(m *mocks.MockETCMappingRepository) {
				expectedParams := repositories.ListMappingsParams{
					Page:      1,
					PageSize:  50,
					SortBy:    "created_at",
					SortOrder: "desc",
				}
				m.On("List", mock.Anything, expectedParams).Return(expectedMappings, int64(2), nil)
			},
			expectError:    false,
			expectMappings: expectedMappings,
			expectTotal:    2,
		},
		{
			name: "with all filters",
			params: &ListMappingsParams{
				Page:             2,
				PageSize:         20,
				ETCRecordID:      int64Ptr(100),
				MappingType:      &mappingType,
				MappedEntityID:   int64Ptr(200),
				MappedEntityType: &mappedEntityType,
				Status:           &status,
				MinConfidence:    &minConfidence,
				MaxConfidence:    &maxConfidence,
				CreatedBy:        &createdBy,
				SortBy:           "confidence",
				SortOrder:        "asc",
			},
			setupMock: func(m *mocks.MockETCMappingRepository) {
				expectedParams := repositories.ListMappingsParams{
					Page:             2,
					PageSize:         20,
					MappingType:      &mappingType,
					Status:           &status,
					MinConfidence:    &minConfidence,
					MappedEntityType: &mappedEntityType,
					SortBy:           "confidence",
					SortOrder:        "asc",
				}
				m.On("List", mock.Anything, expectedParams).Return(expectedMappings, int64(10), nil)
			},
			expectError:    false,
			expectMappings: expectedMappings,
			expectTotal:    10,
		},
		{
			name: "page size over limit",
			params: &ListMappingsParams{
				Page:     1,
				PageSize: 2000, // Over limit
			},
			setupMock: func(m *mocks.MockETCMappingRepository) {
				expectedParams := repositories.ListMappingsParams{
					Page:      1,
					PageSize:  1000, // Should be capped
					SortBy:    "created_at",
					SortOrder: "desc",
				}
				m.On("List", mock.Anything, expectedParams).Return([]*models.ETCMapping{}, int64(0), nil)
			},
			expectError:    false,
			expectMappings: []*models.ETCMapping{},
			expectTotal:    0,
		},
		{
			name:   "repository error",
			params: &ListMappingsParams{},
			setupMock: func(m *mocks.MockETCMappingRepository) {
				m.On("List", mock.Anything, mock.AnythingOfType("repositories.ListMappingsParams")).Return(nil, int64(0), errors.New("db error"))
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockMappingRepo := &mocks.MockETCMappingRepository{}
			mockRecordRepo := &mocks.MockETCMeisaiRecordRepository{}
			tt.setupMock(mockMappingRepo)

			service := NewETCMappingService(mockMappingRepo, mockRecordRepo, nil)
			ctx := context.Background()

			response, err := service.ListMappings(ctx, tt.params)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, response)
			} else {
				assert.NoError(t, err)
				require.NotNil(t, response)
				assert.Equal(t, tt.expectMappings, response.Mappings)
				assert.Equal(t, tt.expectTotal, response.TotalCount)
				assert.Equal(t, tt.params.Page, response.Page)

				if tt.params.PageSize > 0 && tt.params.PageSize <= 1000 {
					assert.Equal(t, tt.params.PageSize, response.PageSize)
				} else {
					assert.Equal(t, 50, response.PageSize)
				}
			}

			mockMappingRepo.AssertExpectations(t)
		})
	}
}

func TestETCMappingService_UpdateMapping(t *testing.T) {
	t.Parallel()

	existingMapping := &models.ETCMapping{
		ID:               1,
		ETCRecordID:      100,
		MappingType:      "automatic",
		MappedEntityID:   200,
		MappedEntityType: "dtako_record",
		Confidence:       0.90,
		Status:           string(models.MappingStatusActive),
	}

	newMappingType := "manual"
	newEntityID := int64(300)
	newEntityType := "custom_record"
	newConfidence := float32(0.95)
	newStatus := string(models.MappingStatusInactive)
	newMetadata := map[string]interface{}{
		"updated": true,
	}

	tests := []struct {
		name        string
		id          int64
		params      *UpdateMappingParams
		setupMock   func(*mocks.MockETCMappingRepository)
		expectError bool
		errorMsg    string
	}{
		{
			name: "successful update with all fields",
			id:   1,
			params: &UpdateMappingParams{
				MappingType:      &newMappingType,
				MappedEntityID:   &newEntityID,
				MappedEntityType: &newEntityType,
				Confidence:       &newConfidence,
				Status:           &newStatus,
				Metadata:         newMetadata,
			},
			setupMock: func(m *mocks.MockETCMappingRepository) {
				txMock := &mocks.MockETCMappingRepository{}
				m.On("BeginTx", mock.Anything).Return(txMock, nil)
				txMock.On("GetByID", mock.Anything, int64(1)).Return(existingMapping, nil)
				txMock.On("Update", mock.Anything, mock.AnythingOfType("*models.ETCMapping")).Return(nil)
				txMock.On("CommitTx").Return(nil)
				txMock.On("RollbackTx").Return(nil).Maybe()
			},
			expectError: false,
		},
		{
			name:   "successful update with partial fields",
			id:     1,
			params: &UpdateMappingParams{
				Confidence: &newConfidence,
			},
			setupMock: func(m *mocks.MockETCMappingRepository) {
				txMock := &mocks.MockETCMappingRepository{}
				m.On("BeginTx", mock.Anything).Return(txMock, nil)
				txMock.On("GetByID", mock.Anything, int64(1)).Return(existingMapping, nil)
				txMock.On("Update", mock.Anything, mock.AnythingOfType("*models.ETCMapping")).Return(nil)
				txMock.On("CommitTx").Return(nil)
				txMock.On("RollbackTx").Return(nil).Maybe()
			},
			expectError: false,
		},
		{
			name:        "invalid ID",
			id:          0,
			params:      &UpdateMappingParams{},
			setupMock:   func(m *mocks.MockETCMappingRepository) {},
			expectError: true,
			errorMsg:    "invalid mapping ID",
		},
		{
			name:   "mapping not found",
			id:     999,
			params: &UpdateMappingParams{Confidence: &newConfidence},
			setupMock: func(m *mocks.MockETCMappingRepository) {
				txMock := &mocks.MockETCMappingRepository{}
				m.On("BeginTx", mock.Anything).Return(txMock, nil)
				txMock.On("GetByID", mock.Anything, int64(999)).Return(nil, errors.New("mapping not found"))
				txMock.On("RollbackTx").Return(nil)
			},
			expectError: true,
			errorMsg:    "failed to retrieve mapping",
		},
		{
			name: "metadata error",
			id:   1,
			params: &UpdateMappingParams{
				Metadata: map[string]interface{}{
					"invalid": make(chan int), // Cannot be marshaled to JSON
				},
			},
			setupMock: func(m *mocks.MockETCMappingRepository) {
				txMock := &mocks.MockETCMappingRepository{}
				m.On("BeginTx", mock.Anything).Return(txMock, nil)
				txMock.On("GetByID", mock.Anything, int64(1)).Return(existingMapping, nil)
				txMock.On("RollbackTx").Return(nil)
			},
			expectError: true,
			errorMsg:    "failed to set metadata",
		},
		{
			name:   "update error",
			id:     1,
			params: &UpdateMappingParams{Confidence: &newConfidence},
			setupMock: func(m *mocks.MockETCMappingRepository) {
				txMock := &mocks.MockETCMappingRepository{}
				m.On("BeginTx", mock.Anything).Return(txMock, nil)
				txMock.On("GetByID", mock.Anything, int64(1)).Return(existingMapping, nil)
				txMock.On("Update", mock.Anything, mock.AnythingOfType("*models.ETCMapping")).Return(errors.New("update failed"))
				txMock.On("RollbackTx").Return(nil)
			},
			expectError: true,
			errorMsg:    "failed to update mapping",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockMappingRepo := &mocks.MockETCMappingRepository{}
			mockRecordRepo := &mocks.MockETCMeisaiRecordRepository{}
			tt.setupMock(mockMappingRepo)

			service := NewETCMappingService(mockMappingRepo, mockRecordRepo, nil)
			ctx := context.Background()

			mapping, err := service.UpdateMapping(ctx, tt.id, tt.params)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
				assert.Nil(t, mapping)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, mapping)
			}

			mockMappingRepo.AssertExpectations(t)
		})
	}
}

func TestETCMappingService_DeleteMapping(t *testing.T) {
	t.Parallel()

	existingMapping := &models.ETCMapping{ID: 1}

	tests := []struct {
		name        string
		id          int64
		setupMock   func(*mocks.MockETCMappingRepository)
		expectError bool
		errorMsg    string
	}{
		{
			name: "successful deletion",
			id:   1,
			setupMock: func(m *mocks.MockETCMappingRepository) {
				txMock := &mocks.MockETCMappingRepository{}
				m.On("BeginTx", mock.Anything).Return(txMock, nil)
				txMock.On("GetByID", mock.Anything, int64(1)).Return(existingMapping, nil)
				txMock.On("Delete", mock.Anything, int64(1)).Return(nil)
				txMock.On("CommitTx").Return(nil)
				txMock.On("RollbackTx").Return(nil).Maybe()
			},
			expectError: false,
		},
		{
			name:        "invalid ID",
			id:          0,
			setupMock:   func(m *mocks.MockETCMappingRepository) {},
			expectError: true,
			errorMsg:    "invalid mapping ID",
		},
		{
			name: "mapping not found",
			id:   999,
			setupMock: func(m *mocks.MockETCMappingRepository) {
				txMock := &mocks.MockETCMappingRepository{}
				m.On("BeginTx", mock.Anything).Return(txMock, nil)
				txMock.On("GetByID", mock.Anything, int64(999)).Return(nil, errors.New("mapping not found"))
				txMock.On("RollbackTx").Return(nil)
			},
			expectError: true,
			errorMsg:    "failed to retrieve mapping",
		},
		{
			name: "delete error",
			id:   1,
			setupMock: func(m *mocks.MockETCMappingRepository) {
				txMock := &mocks.MockETCMappingRepository{}
				m.On("BeginTx", mock.Anything).Return(txMock, nil)
				txMock.On("GetByID", mock.Anything, int64(1)).Return(existingMapping, nil)
				txMock.On("Delete", mock.Anything, int64(1)).Return(errors.New("delete failed"))
				txMock.On("RollbackTx").Return(nil)
			},
			expectError: true,
			errorMsg:    "failed to delete mapping",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockMappingRepo := &mocks.MockETCMappingRepository{}
			mockRecordRepo := &mocks.MockETCMeisaiRecordRepository{}
			tt.setupMock(mockMappingRepo)

			service := NewETCMappingService(mockMappingRepo, mockRecordRepo, nil)
			ctx := context.Background()

			err := service.DeleteMapping(ctx, tt.id)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
			}

			mockMappingRepo.AssertExpectations(t)
		})
	}
}

func TestETCMappingService_UpdateStatus(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		id          int64
		status      string
		setupMock   func(*mocks.MockETCMappingRepository)
		expectError bool
		errorMsg    string
	}{
		{
			name:   "successful status update",
			id:     1,
			status: string(models.MappingStatusInactive),
			setupMock: func(m *mocks.MockETCMappingRepository) {
				m.On("UpdateStatus", mock.Anything, int64(1), string(models.MappingStatusInactive)).Return(nil)
			},
			expectError: false,
		},
		{
			name:        "invalid ID",
			id:          0,
			status:      string(models.MappingStatusActive),
			setupMock:   func(m *mocks.MockETCMappingRepository) {},
			expectError: true,
			errorMsg:    "invalid mapping ID",
		},
		{
			name:   "repository error",
			id:     1,
			status: string(models.MappingStatusActive),
			setupMock: func(m *mocks.MockETCMappingRepository) {
				m.On("UpdateStatus", mock.Anything, int64(1), string(models.MappingStatusActive)).Return(errors.New("update failed"))
			},
			expectError: true,
			errorMsg:    "failed to update mapping status",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockMappingRepo := &mocks.MockETCMappingRepository{}
			mockRecordRepo := &mocks.MockETCMeisaiRecordRepository{}
			tt.setupMock(mockMappingRepo)

			service := NewETCMappingService(mockMappingRepo, mockRecordRepo, nil)
			ctx := context.Background()

			err := service.UpdateStatus(ctx, tt.id, tt.status)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
			}

			mockMappingRepo.AssertExpectations(t)
		})
	}
}

func TestETCMappingService_HealthCheck(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		setupMock   func(*mocks.MockETCMappingRepository, *mocks.MockETCMeisaiRecordRepository)
		expectError bool
		errorMsg    string
	}{
		{
			name: "all repositories healthy",
			setupMock: func(mappingRepo *mocks.MockETCMappingRepository, recordRepo *mocks.MockETCMeisaiRecordRepository) {
				mappingRepo.On("Ping", mock.Anything).Return(nil)
				recordRepo.On("Ping", mock.Anything).Return(nil)
			},
			expectError: false,
		},
		{
			name: "mapping repository unhealthy",
			setupMock: func(mappingRepo *mocks.MockETCMappingRepository, recordRepo *mocks.MockETCMeisaiRecordRepository) {
				mappingRepo.On("Ping", mock.Anything).Return(errors.New("connection failed"))
			},
			expectError: true,
			errorMsg:    "mapping repository ping failed",
		},
		{
			name: "record repository unhealthy",
			setupMock: func(mappingRepo *mocks.MockETCMappingRepository, recordRepo *mocks.MockETCMeisaiRecordRepository) {
				mappingRepo.On("Ping", mock.Anything).Return(nil)
				recordRepo.On("Ping", mock.Anything).Return(errors.New("connection failed"))
			},
			expectError: true,
			errorMsg:    "record repository ping failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockMappingRepo := &mocks.MockETCMappingRepository{}
			mockRecordRepo := &mocks.MockETCMeisaiRecordRepository{}
			tt.setupMock(mockMappingRepo, mockRecordRepo)

			service := NewETCMappingService(mockMappingRepo, mockRecordRepo, nil)
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

			mockMappingRepo.AssertExpectations(t)
			mockRecordRepo.AssertExpectations(t)
		})
	}
}

func TestETCMappingService_HealthCheck_NilRepositories(t *testing.T) {
	t.Parallel()

	t.Run("nil mapping repository", func(t *testing.T) {
		service := &ETCMappingService{
			mappingRepo: nil,
			recordRepo:  &mocks.MockETCMeisaiRecordRepository{},
		}
		ctx := context.Background()

		err := service.HealthCheck(ctx)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "mapping repository not initialized")
	})
}

func TestETCMappingService_HealthCheck_NilRecordRepository(t *testing.T) {
	t.Parallel()

	t.Run("nil record repository (should not fail)", func(t *testing.T) {
		mockMappingRepo := &mocks.MockETCMappingRepository{}
		mockMappingRepo.On("Ping", mock.Anything).Return(nil)

		service := &ETCMappingService{
			mappingRepo: mockMappingRepo,
			recordRepo:  nil, // nil record repo should not cause failure
		}
		ctx := context.Background()

		err := service.HealthCheck(ctx)

		assert.NoError(t, err)
		mockMappingRepo.AssertExpectations(t)
	})
}

// Context cancellation tests
func TestETCMappingService_ContextCancellation(t *testing.T) {
	t.Parallel()

	t.Run("create mapping with cancelled context", func(t *testing.T) {
		mockMappingRepo := &mocks.MockETCMappingRepository{}
		mockRecordRepo := &mocks.MockETCMeisaiRecordRepository{}
		service := NewETCMappingService(mockMappingRepo, mockRecordRepo, nil)

		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		params := &CreateMappingParams{
			ETCRecordID:      1,
			MappingType:      "automatic",
			MappedEntityID:   100,
			MappedEntityType: "dtako_record",
		}

		// Mock should handle the cancelled context appropriately
		mockMappingRepo.On("BeginTx", mock.Anything).Return(nil, context.Canceled)

		_, err := service.CreateMapping(ctx, params)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to start transaction")
	})
}
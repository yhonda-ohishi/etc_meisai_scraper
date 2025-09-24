package services_test

import (
	"context"
	"log"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/yhonda-ohishi/etc_meisai/src/models"
	"github.com/yhonda-ohishi/etc_meisai/src/services"
	"github.com/yhonda-ohishi/etc_meisai/src/mocks"
	"github.com/yhonda-ohishi/etc_meisai/tests/helpers"
)

func TestETCMappingService_NewETCMappingService(t *testing.T) {
	mockMappingRepo := &mocks.MockETCMappingRepository{}
	mockRecordRepo := &mocks.MockETCMeisaiRecordRepository{}
	logger := log.New(log.Writer(), "[TEST] ", log.LstdFlags)

	service := services.NewETCMappingService(mockMappingRepo, mockRecordRepo, logger)
	helpers.AssertNotNil(t, service)
}

func TestETCMappingService_CreateMapping(t *testing.T) {
	mockMappingRepo := &mocks.MockETCMappingRepository{}
	mockRecordRepo := &mocks.MockETCMeisaiRecordRepository{}
	logger := log.New(log.Writer(), "[TEST] ", log.LstdFlags)

	service := services.NewETCMappingService(mockMappingRepo, mockRecordRepo, logger)

	params := &services.CreateMappingParams{
		ETCRecordID:      1,
		MappingType:      string(models.MappingTypeDtako),
		MappedEntityID:   123,
		MappedEntityType: string(models.EntityTypeDtakoRecord),
		Confidence:       0.95,
		Status:           string(models.MappingStatusActive),
	}

	tests := []struct {
		name      string
		params    *services.CreateMappingParams
		setupMock func()
		wantErr   bool
		errMsg    string
	}{
		{
			name:   "successful creation",
			params: params,
			setupMock: func() {
				mockMappingRepo.On("Create", mock.Anything, mock.AnythingOfType("*models.ETCMapping")).
					Return(nil).Once().
					Run(func(args mock.Arguments) {
						mapping := args.Get(1).(*models.ETCMapping)
						mapping.ID = 1
					})
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
			name: "invalid params data",
			params: &services.CreateMappingParams{
				ETCRecordID:      0, // Invalid zero ID
				MappingType:      string(models.MappingTypeDtako),
				MappedEntityID:   123,
				MappedEntityType: string(models.EntityTypeDtakoRecord),
				Confidence:       0.95,
				Status:           string(models.MappingStatusActive),
			},
			setupMock: func() {
				// No mock setup needed as validation should fail first
			},
			wantErr: true,
			errMsg:  "validation failed",
		},
		{
			name:   "repository error",
			params: params,
			setupMock: func() {
				mockMappingRepo.On("Create", mock.Anything, mock.AnythingOfType("*models.ETCMapping")).
					Return(assert.AnError).Once()
			},
			wantErr: true,
			errMsg:  "failed to create mapping",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockMappingRepo.ExpectedCalls = nil // Reset mock expectations
			tt.setupMock()

			ctx := context.Background()
			mapping, err := service.CreateMapping(ctx, tt.params)

			if tt.wantErr {
				helpers.AssertError(t, err)
				helpers.AssertNil(t, mapping)
				if tt.errMsg != "" {
					helpers.AssertContains(t, err.Error(), tt.errMsg)
				}
			} else {
				helpers.AssertNoError(t, err)
				helpers.AssertNotNil(t, mapping)
			}

			mockMappingRepo.AssertExpectations(t)
		})
	}
}
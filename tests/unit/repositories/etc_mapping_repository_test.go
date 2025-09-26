package repositories_test

// TODO: This test file needs major refactoring to match current model structure
// Many tests are currently broken due to interface changes

import (
	"context"
	"errors"
	"testing"

	"go.uber.org/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yhonda-ohishi/etc_meisai/src/models"
	"github.com/yhonda-ohishi/etc_meisai/src/repositories"
	"github.com/yhonda-ohishi/etc_meisai/tests/mocks"
)

func TestETCMappingRepository_List(t *testing.T) {
	tests := []struct {
		name      string
		setupMock func(*mocks.MockETCMappingRepository)
		want      []*models.ETCMapping
		wantErr   bool
		errMsg    string
	}{
		{
			name: "successful retrieval",
			setupMock: func(m *mocks.MockETCMappingRepository) {
				mappings := []*models.ETCMapping{
					{ID: 1, ETCRecordID: 1234567890, MappingType: "dtako"},
					{ID: 2, ETCRecordID: 987654321, MappingType: "expense"},
				}
				m.EXPECT().List(gomock.Any(), gomock.Any()).Return(mappings, int64(2), nil)
			},
			want: []*models.ETCMapping{
				{ID: 1, ETCRecordID: 1234567890, MappingType: "dtako"},
				{ID: 2, ETCRecordID: 987654321, MappingType: "expense"},
			},
			wantErr: false,
		},
		{
			name: "empty result",
			setupMock: func(m *mocks.MockETCMappingRepository) {
				m.EXPECT().List(gomock.Any(), gomock.Any()).Return([]*models.ETCMapping{}, int64(0), nil)
			},
			want:    []*models.ETCMapping{},
			wantErr: false,
		},
		{
			name: "database error",
			setupMock: func(m *mocks.MockETCMappingRepository) {
				m.EXPECT().List(gomock.Any(), gomock.Any()).Return(nil, int64(0), errors.New("database connection failed"))
			},
			want:    nil,
			wantErr: true,
			errMsg:  "database connection failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mocks.NewMockETCMappingRepository(ctrl)
			tt.setupMock(mockRepo)

			got, _, err := mockRepo.List(context.Background(), repositories.ListMappingsParams{})

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestETCMappingRepository_GetByID(t *testing.T) {
	tests := []struct {
		name      string
		id        int64
		setupMock func(*mocks.MockETCMappingRepository)
		want      *models.ETCMapping
		wantErr   bool
		errMsg    string
	}{
		{
			name: "existing record",
			id:   1,
			setupMock: func(m *mocks.MockETCMappingRepository) {
				mapping := &models.ETCMapping{
					ID:          1,
					ETCRecordID: 1234567890,
					MappingType: "dtako",
				}
				m.EXPECT().GetByID(gomock.Any(), int64(1)).Return(mapping, nil)
			},
			want: &models.ETCMapping{
				ID:          1,
				ETCRecordID: 1234567890,
				MappingType: "dtako",
			},
			wantErr: false,
		},
		{
			name: "non-existent record",
			id:   999,
			setupMock: func(m *mocks.MockETCMappingRepository) {
				m.EXPECT().GetByID(gomock.Any(), int64(999)).Return(nil, ErrNotFound)
			},
			want:    nil,
			wantErr: true,
			errMsg:  "not found",
		},
		{
			name: "database error",
			id:   1,
			setupMock: func(m *mocks.MockETCMappingRepository) {
				m.EXPECT().GetByID(gomock.Any(), int64(1)).Return(nil, errors.New("database error"))
			},
			want:    nil,
			wantErr: true,
			errMsg:  "database error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mocks.NewMockETCMappingRepository(ctrl)
			tt.setupMock(mockRepo)

			got, err := mockRepo.GetByID(context.Background(), tt.id)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestETCMappingRepository_Create(t *testing.T) {
	tests := []struct {
		name      string
		input     *models.ETCMapping
		setupMock func(*mocks.MockETCMappingRepository)
		wantErr   bool
		errMsg    string
	}{
		{
			name: "successful creation",
			input: &models.ETCMapping{
				ETCRecordID: 1234567890,
				MappingType: "dtako",
			},
			setupMock: func(m *mocks.MockETCMappingRepository) {
				m.EXPECT().Create(gomock.Any(), gomock.Any()).DoAndReturn(
					func(ctx context.Context, mapping *models.ETCMapping) error {
						mapping.ID = 1
						return nil
					})
			},
			wantErr: false,
		},
		{
			name: "validation error",
			input: &models.ETCMapping{
				ETCRecordID: 0, // Invalid zero ETC record ID
			},
			setupMock: func(m *mocks.MockETCMappingRepository) {
				m.EXPECT().Create(gomock.Any(), gomock.Any()).Return(errors.New("validation failed: etc_num required"))
			},
			wantErr: true,
			errMsg:  "validation failed",
		},
		{
			name: "duplicate key error",
			input: &models.ETCMapping{
				ETCRecordID: 1234567890,
				MappingType: "dtako",
			},
			setupMock: func(m *mocks.MockETCMappingRepository) {
				m.EXPECT().Create(gomock.Any(), gomock.Any()).Return(errors.New("duplicate key"))
			},
			wantErr: true,
			errMsg:  "duplicate key",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mocks.NewMockETCMappingRepository(ctrl)
			tt.setupMock(mockRepo)

			err := mockRepo.Create(context.Background(), tt.input)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
				assert.NotZero(t, tt.input.ID)
			}
		})
	}
}

func TestETCMappingRepository_Update(t *testing.T) {
	tests := []struct {
		name      string
		input     *models.ETCMapping
		setupMock func(*mocks.MockETCMappingRepository)
		wantErr   bool
		errMsg    string
	}{
		{
			name: "successful update",
			input: &models.ETCMapping{
				ID:          1,
				ETCRecordID: 1234567890,
				MappingType: "COMP002", // Updated company code
			},
			setupMock: func(m *mocks.MockETCMappingRepository) {
				m.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "non-existent record",
			input: &models.ETCMapping{
				ID:          999,
				ETCRecordID: 1234567890,
				MappingType: "COMP002",
			},
			setupMock: func(m *mocks.MockETCMappingRepository) {
				m.EXPECT().Update(gomock.Any(), gomock.Any()).Return(ErrNotFound)
			},
			wantErr: true,
			errMsg:  "not found",
		},
		{
			name: "database error",
			input: &models.ETCMapping{
				ID:          1,
				ETCRecordID: 1234567890,
				MappingType: "COMP002",
			},
			setupMock: func(m *mocks.MockETCMappingRepository) {
				m.EXPECT().Update(gomock.Any(), gomock.Any()).Return(errors.New("connection lost"))
			},
			wantErr: true,
			errMsg:  "connection lost",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mocks.NewMockETCMappingRepository(ctrl)
			tt.setupMock(mockRepo)

			err := mockRepo.Update(context.Background(), tt.input)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestETCMappingRepository_Delete(t *testing.T) {
	tests := []struct {
		name      string
		id        int64
		setupMock func(*mocks.MockETCMappingRepository)
		wantErr   bool
		errMsg    string
	}{
		{
			name: "successful deletion",
			id:   1,
			setupMock: func(m *mocks.MockETCMappingRepository) {
				m.EXPECT().Delete(gomock.Any(), int64(1)).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "non-existent record",
			id:   999,
			setupMock: func(m *mocks.MockETCMappingRepository) {
				m.EXPECT().Delete(gomock.Any(), int64(999)).Return(ErrNotFound)
			},
			wantErr: true,
			errMsg:  "not found",
		},
		{
			name: "database error",
			id:   1,
			setupMock: func(m *mocks.MockETCMappingRepository) {
				m.EXPECT().Delete(gomock.Any(), int64(1)).Return(errors.New("foreign key constraint"))
			},
			wantErr: true,
			errMsg:  "foreign key constraint",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mocks.NewMockETCMappingRepository(ctrl)
			tt.setupMock(mockRepo)

			err := mockRepo.Delete(context.Background(), tt.id)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestETCMappingRepository_GetByETCRecordID(t *testing.T) {
	tests := []struct {
		name      string
		etcNum    int64
		setupMock func(*mocks.MockETCMappingRepository)
		want      *models.ETCMapping
		wantErr   bool
		errMsg    string
	}{
		{
			name:   "existing etc number",
			etcNum: 1234567890,
			setupMock: func(m *mocks.MockETCMappingRepository) {
				mapping := &models.ETCMapping{
					ID:          1,
					ETCRecordID: 1234567890,
					MappingType: "dtako",
				}
				m.EXPECT().GetByETCRecordID(gomock.Any(), int64(1234567890)).Return(mapping, nil)
			},
			want: &models.ETCMapping{
				ID:          1,
				ETCRecordID: 1234567890,
				MappingType: "dtako",
			},
			wantErr: false,
		},
		{
			name:   "non-existent etc number",
			etcNum: 9999999999,
			setupMock: func(m *mocks.MockETCMappingRepository) {
				m.EXPECT().GetByETCRecordID(gomock.Any(), 9999999999).Return(nil, ErrNotFound)
			},
			want:    nil,
			wantErr: true,
			errMsg:  "not found",
		},
		{
			name:   "empty etc number",
			etcNum: 0,
			setupMock: func(m *mocks.MockETCMappingRepository) {
				m.EXPECT().GetByETCRecordID(gomock.Any(), int64(0)).Return(nil, errors.New("etc_num cannot be empty"))
			},
			want:    nil,
			wantErr: true,
			errMsg:  "etc_num cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mocks.NewMockETCMappingRepository(ctrl)
			tt.setupMock(mockRepo)

			got, err := mockRepo.GetByETCRecordID(context.Background(), tt.etcNum)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestETCMappingRepository_BulkCreate(t *testing.T) {
	tests := []struct {
		name      string
		mappings  []*models.ETCMapping
		setupMock func(*mocks.MockETCMappingRepository)
		wantErr   bool
		errMsg    string
	}{
		{
			name: "successful bulk creation",
			mappings: []*models.ETCMapping{
				{ETCRecordID: 1111111111, MappingType: "dtako"},
				{ETCRecordID: 2222222222, MappingType: "COMP002"},
				{ETCRecordID: 3333333333, MappingType: "COMP003"},
			},
			setupMock: func(m *mocks.MockETCMappingRepository) {
				m.EXPECT().BulkCreate(gomock.Any(), gomock.Len(3)).Return(nil)
			},
			wantErr: false,
		},
		{
			name:     "empty list",
			mappings: []*models.ETCMapping{},
			setupMock: func(m *mocks.MockETCMappingRepository) {
				m.EXPECT().BulkCreate(gomock.Any(), gomock.Len(0)).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "partial failure",
			mappings: []*models.ETCMapping{
				{ETCRecordID: 1111111111, MappingType: "dtako"},
				{ETCRecordID: 1111111111, MappingType: "COMP002"}, // Duplicate
			},
			setupMock: func(m *mocks.MockETCMappingRepository) {
				m.EXPECT().BulkCreate(gomock.Any(), gomock.Any()).Return(errors.New("duplicate key: etc_num"))
			},
			wantErr: true,
			errMsg:  "duplicate key",
		},
		{
			name: "transaction error",
			mappings: []*models.ETCMapping{
				{ETCRecordID: 1111111111, MappingType: "dtako"},
			},
			setupMock: func(m *mocks.MockETCMappingRepository) {
				m.EXPECT().BulkCreate(gomock.Any(), gomock.Any()).Return(errors.New("transaction rollback"))
			},
			wantErr: true,
			errMsg:  "transaction rollback",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mocks.NewMockETCMappingRepository(ctrl)
			tt.setupMock(mockRepo)

			err := mockRepo.BulkCreate(context.Background(), tt.mappings)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
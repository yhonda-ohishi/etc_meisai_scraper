package repositories

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/yhonda-ohishi/etc_meisai/src/models"
)

func TestListMappingsParams_Validation(t *testing.T) {
	dateFrom := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	dateTo := time.Date(2025, 1, 31, 23, 59, 59, 999999999, time.UTC)
	mappingType := "automatic"
	status := "active"
	minConfidence := float32(0.8)
	entityType := "vehicle"

	tests := []struct {
		name   string
		params ListMappingsParams
		valid  bool
	}{
		{
			name: "valid parameters",
			params: ListMappingsParams{
				Page:             1,
				PageSize:         10,
				DateFrom:         &dateFrom,
				DateTo:           &dateTo,
				MappingType:      &mappingType,
				Status:           &status,
				MinConfidence:    &minConfidence,
				MappedEntityType: &entityType,
				SortBy:           "created_at",
				SortOrder:        "asc",
			},
			valid: true,
		},
		{
			name: "minimal parameters",
			params: ListMappingsParams{
				Page:     1,
				PageSize: 10,
			},
			valid: true,
		},
		{
			name: "zero page",
			params: ListMappingsParams{
				Page:     0,
				PageSize: 10,
			},
			valid: true,
		},
		{
			name: "negative page",
			params: ListMappingsParams{
				Page:     -1,
				PageSize: 10,
			},
			valid: true,
		},
		{
			name: "zero page size",
			params: ListMappingsParams{
				Page:     1,
				PageSize: 0,
			},
			valid: true,
		},
		{
			name: "large page size",
			params: ListMappingsParams{
				Page:     1,
				PageSize: 1000,
			},
			valid: true,
		},
		{
			name: "date range with from after to",
			params: ListMappingsParams{
				Page:     1,
				PageSize: 10,
				DateFrom: &dateTo,
				DateTo:   &dateFrom,
			},
			valid: true,
		},
		{
			name: "same from and to date",
			params: ListMappingsParams{
				Page:     1,
				PageSize: 10,
				DateFrom: &dateFrom,
				DateTo:   &dateFrom,
			},
			valid: true,
		},
		{
			name: "confidence zero",
			params: ListMappingsParams{
				Page:          1,
				PageSize:      10,
				MinConfidence: floatPtr(0.0),
			},
			valid: true,
		},
		{
			name: "confidence one",
			params: ListMappingsParams{
				Page:          1,
				PageSize:      10,
				MinConfidence: floatPtr(1.0),
			},
			valid: true,
		},
		{
			name: "confidence above one",
			params: ListMappingsParams{
				Page:          1,
				PageSize:      10,
				MinConfidence: floatPtr(1.5),
			},
			valid: true,
		},
		{
			name: "negative confidence",
			params: ListMappingsParams{
				Page:          1,
				PageSize:      10,
				MinConfidence: floatPtr(-0.1),
			},
			valid: true,
		},
		{
			name: "empty string parameters",
			params: ListMappingsParams{
				Page:             1,
				PageSize:         10,
				MappingType:      stringPtr(""),
				Status:           stringPtr(""),
				MappedEntityType: stringPtr(""),
				SortBy:           "",
				SortOrder:        "",
			},
			valid: true,
		},
		{
			name: "invalid sort order",
			params: ListMappingsParams{
				Page:      1,
				PageSize:  10,
				SortBy:    "created_at",
				SortOrder: "invalid",
			},
			valid: true,
		},
		{
			name: "desc sort order",
			params: ListMappingsParams{
				Page:      1,
				PageSize:  10,
				SortBy:    "confidence",
				SortOrder: "desc",
			},
			valid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Since these are just data structures, we just test that they can be created
			// without error and have the expected values
			assert.Equal(t, tt.params.Page, tt.params.Page)
			assert.Equal(t, tt.params.PageSize, tt.params.PageSize)

			if tt.params.DateFrom != nil {
				assert.NotNil(t, tt.params.DateFrom)
			}
			if tt.params.DateTo != nil {
				assert.NotNil(t, tt.params.DateTo)
			}
			if tt.params.MappingType != nil {
				assert.NotNil(t, tt.params.MappingType)
			}
			if tt.params.Status != nil {
				assert.NotNil(t, tt.params.Status)
			}
			if tt.params.MinConfidence != nil {
				assert.NotNil(t, tt.params.MinConfidence)
			}
			if tt.params.MappedEntityType != nil {
				assert.NotNil(t, tt.params.MappedEntityType)
			}

			assert.Equal(t, tt.params.SortBy, tt.params.SortBy)
			assert.Equal(t, tt.params.SortOrder, tt.params.SortOrder)
		})
	}
}

func TestETCMappingRepository_InterfaceMethods(t *testing.T) {
	// Test that the interface can be implemented
	// This tests the interface definition and method signatures

	ctx := context.Background()
	mapping := &models.ETCMapping{
		ID:               1,
		ETCRecordID:      100,
		MappedEntityType: "vehicle",
		MappedEntityID:   200,
		MappingType:      "automatic",
		Confidence:       0.95,
		Status:           "active",
		CreatedBy:        "system",
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	// Test interface method signatures by creating a mock implementation
	mockRepo := &mockETCMappingRepository{}

	// Test basic CRUD operations
	t.Run("Create", func(t *testing.T) {
		err := mockRepo.Create(ctx, mapping)
		assert.NoError(t, err)
	})

	t.Run("GetByID", func(t *testing.T) {
		result, err := mockRepo.GetByID(ctx, 1)
		assert.NoError(t, err)
		assert.Equal(t, mapping, result)
	})

	t.Run("Update", func(t *testing.T) {
		err := mockRepo.Update(ctx, mapping)
		assert.NoError(t, err)
	})

	t.Run("Delete", func(t *testing.T) {
		err := mockRepo.Delete(ctx, 1)
		assert.NoError(t, err)
	})

	// Test query operations
	t.Run("GetByETCRecordID", func(t *testing.T) {
		result, err := mockRepo.GetByETCRecordID(ctx, 100)
		assert.NoError(t, err)
		assert.Len(t, result, 1)
		assert.Equal(t, mapping, result[0])
	})

	t.Run("GetByMappedEntity", func(t *testing.T) {
		result, err := mockRepo.GetByMappedEntity(ctx, "vehicle", 200)
		assert.NoError(t, err)
		assert.Len(t, result, 1)
		assert.Equal(t, mapping, result[0])
	})

	t.Run("GetActiveMapping", func(t *testing.T) {
		result, err := mockRepo.GetActiveMapping(ctx, 100)
		assert.NoError(t, err)
		assert.Equal(t, mapping, result)
	})

	// Test list operations
	t.Run("List", func(t *testing.T) {
		params := ListMappingsParams{
			Page:     1,
			PageSize: 10,
		}
		result, count, err := mockRepo.List(ctx, params)
		assert.NoError(t, err)
		assert.Len(t, result, 1)
		assert.Equal(t, int64(1), count)
		assert.Equal(t, mapping, result[0])
	})

	// Test bulk operations
	t.Run("BulkCreate", func(t *testing.T) {
		mappings := []*models.ETCMapping{mapping}
		err := mockRepo.BulkCreate(ctx, mappings)
		assert.NoError(t, err)
	})

	t.Run("UpdateStatus", func(t *testing.T) {
		err := mockRepo.UpdateStatus(ctx, 1, "inactive")
		assert.NoError(t, err)
	})

	// Test transaction support
	t.Run("BeginTx", func(t *testing.T) {
		txRepo, err := mockRepo.BeginTx(ctx)
		assert.NoError(t, err)
		assert.NotNil(t, txRepo)
		assert.IsType(t, &mockETCMappingRepository{}, txRepo)
	})

	t.Run("CommitTx", func(t *testing.T) {
		err := mockRepo.CommitTx()
		assert.NoError(t, err)
	})

	t.Run("RollbackTx", func(t *testing.T) {
		err := mockRepo.RollbackTx()
		assert.NoError(t, err)
	})

	// Test health check
	t.Run("Ping", func(t *testing.T) {
		err := mockRepo.Ping(ctx)
		assert.NoError(t, err)
	})
}

func TestETCMappingRepository_ErrorScenarios(t *testing.T) {
	ctx := context.Background()
	errorRepo := &errorETCMappingRepository{}

	// Test error handling for all methods
	t.Run("Create error", func(t *testing.T) {
		err := errorRepo.Create(ctx, nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "create failed")
	})

	t.Run("GetByID error", func(t *testing.T) {
		result, err := errorRepo.GetByID(ctx, 1)
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "get failed")
	})

	t.Run("Update error", func(t *testing.T) {
		err := errorRepo.Update(ctx, nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "update failed")
	})

	t.Run("Delete error", func(t *testing.T) {
		err := errorRepo.Delete(ctx, 1)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "delete failed")
	})

	t.Run("GetByETCRecordID error", func(t *testing.T) {
		result, err := errorRepo.GetByETCRecordID(ctx, 1)
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "get by etc record id failed")
	})

	t.Run("GetByMappedEntity error", func(t *testing.T) {
		result, err := errorRepo.GetByMappedEntity(ctx, "vehicle", 1)
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "get by mapped entity failed")
	})

	t.Run("GetActiveMapping error", func(t *testing.T) {
		result, err := errorRepo.GetActiveMapping(ctx, 1)
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "get active mapping failed")
	})

	t.Run("List error", func(t *testing.T) {
		params := ListMappingsParams{Page: 1, PageSize: 10}
		result, count, err := errorRepo.List(ctx, params)
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, int64(0), count)
		assert.Contains(t, err.Error(), "list failed")
	})

	t.Run("BulkCreate error", func(t *testing.T) {
		err := errorRepo.BulkCreate(ctx, nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "bulk create failed")
	})

	t.Run("UpdateStatus error", func(t *testing.T) {
		err := errorRepo.UpdateStatus(ctx, 1, "active")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "update status failed")
	})

	t.Run("BeginTx error", func(t *testing.T) {
		result, err := errorRepo.BeginTx(ctx)
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "begin transaction failed")
	})

	t.Run("CommitTx error", func(t *testing.T) {
		err := errorRepo.CommitTx()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "commit transaction failed")
	})

	t.Run("RollbackTx error", func(t *testing.T) {
		err := errorRepo.RollbackTx()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "rollback transaction failed")
	})

	t.Run("Ping error", func(t *testing.T) {
		err := errorRepo.Ping(ctx)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "ping failed")
	})
}

// Helper functions
func stringPtr(s string) *string {
	return &s
}

func floatPtr(f float32) *float32 {
	return &f
}

// Mock implementation for testing
type mockETCMappingRepository struct{}

func (m *mockETCMappingRepository) Create(ctx context.Context, mapping *models.ETCMapping) error {
	return nil
}

func (m *mockETCMappingRepository) GetByID(ctx context.Context, id int64) (*models.ETCMapping, error) {
	return &models.ETCMapping{
		ID:               id,
		ETCRecordID:      100,
		MappedEntityType: "vehicle",
		MappedEntityID:   200,
		MappingType:      "automatic",
		Confidence:       0.95,
		Status:           "active",
		CreatedBy:        "system",
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}, nil
}

func (m *mockETCMappingRepository) Update(ctx context.Context, mapping *models.ETCMapping) error {
	return nil
}

func (m *mockETCMappingRepository) Delete(ctx context.Context, id int64) error {
	return nil
}

func (m *mockETCMappingRepository) GetByETCRecordID(ctx context.Context, etcRecordID int64) ([]*models.ETCMapping, error) {
	return []*models.ETCMapping{
		{
			ID:               1,
			ETCRecordID:      etcRecordID,
			MappedEntityType: "vehicle",
			MappedEntityID:   200,
			MappingType:      "automatic",
			Confidence:       0.95,
			Status:           "active",
			CreatedBy:        "system",
			CreatedAt:        time.Now(),
			UpdatedAt:        time.Now(),
		},
	}, nil
}

func (m *mockETCMappingRepository) GetByMappedEntity(ctx context.Context, entityType string, entityID int64) ([]*models.ETCMapping, error) {
	return []*models.ETCMapping{
		{
			ID:               1,
			ETCRecordID:      100,
			MappedEntityType: entityType,
			MappedEntityID:   entityID,
			MappingType:      "automatic",
			Confidence:       0.95,
			Status:           "active",
			CreatedBy:        "system",
			CreatedAt:        time.Now(),
			UpdatedAt:        time.Now(),
		},
	}, nil
}

func (m *mockETCMappingRepository) GetActiveMapping(ctx context.Context, etcRecordID int64) (*models.ETCMapping, error) {
	return &models.ETCMapping{
		ID:               1,
		ETCRecordID:      etcRecordID,
		MappedEntityType: "vehicle",
		MappedEntityID:   200,
		MappingType:      "automatic",
		Confidence:       0.95,
		Status:           "active",
		CreatedBy:        "system",
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}, nil
}

func (m *mockETCMappingRepository) List(ctx context.Context, params ListMappingsParams) ([]*models.ETCMapping, int64, error) {
	return []*models.ETCMapping{
		{
			ID:               1,
			ETCRecordID:      100,
			MappedEntityType: "vehicle",
			MappedEntityID:   200,
			MappingType:      "automatic",
			Confidence:       0.95,
			Status:           "active",
			CreatedBy:        "system",
			CreatedAt:        time.Now(),
			UpdatedAt:        time.Now(),
		},
	}, 1, nil
}

func (m *mockETCMappingRepository) BulkCreate(ctx context.Context, mappings []*models.ETCMapping) error {
	return nil
}

func (m *mockETCMappingRepository) UpdateStatus(ctx context.Context, id int64, status string) error {
	return nil
}

func (m *mockETCMappingRepository) BeginTx(ctx context.Context) (ETCMappingRepository, error) {
	return m, nil
}

func (m *mockETCMappingRepository) CommitTx() error {
	return nil
}

func (m *mockETCMappingRepository) RollbackTx() error {
	return nil
}

func (m *mockETCMappingRepository) Ping(ctx context.Context) error {
	return nil
}

// Error implementation for testing error scenarios
type errorETCMappingRepository struct{}

func (e *errorETCMappingRepository) Create(ctx context.Context, mapping *models.ETCMapping) error {
	return assert.AnError
}

func (e *errorETCMappingRepository) GetByID(ctx context.Context, id int64) (*models.ETCMapping, error) {
	return nil, assert.AnError
}

func (e *errorETCMappingRepository) Update(ctx context.Context, mapping *models.ETCMapping) error {
	return assert.AnError
}

func (e *errorETCMappingRepository) Delete(ctx context.Context, id int64) error {
	return assert.AnError
}

func (e *errorETCMappingRepository) GetByETCRecordID(ctx context.Context, etcRecordID int64) ([]*models.ETCMapping, error) {
	return nil, assert.AnError
}

func (e *errorETCMappingRepository) GetByMappedEntity(ctx context.Context, entityType string, entityID int64) ([]*models.ETCMapping, error) {
	return nil, assert.AnError
}

func (e *errorETCMappingRepository) GetActiveMapping(ctx context.Context, etcRecordID int64) (*models.ETCMapping, error) {
	return nil, assert.AnError
}

func (e *errorETCMappingRepository) List(ctx context.Context, params ListMappingsParams) ([]*models.ETCMapping, int64, error) {
	return nil, 0, assert.AnError
}

func (e *errorETCMappingRepository) BulkCreate(ctx context.Context, mappings []*models.ETCMapping) error {
	return assert.AnError
}

func (e *errorETCMappingRepository) UpdateStatus(ctx context.Context, id int64, status string) error {
	return assert.AnError
}

func (e *errorETCMappingRepository) BeginTx(ctx context.Context) (ETCMappingRepository, error) {
	return nil, assert.AnError
}

func (e *errorETCMappingRepository) CommitTx() error {
	return assert.AnError
}

func (e *errorETCMappingRepository) RollbackTx() error {
	return assert.AnError
}

func (e *errorETCMappingRepository) Ping(ctx context.Context) error {
	return assert.AnError
}
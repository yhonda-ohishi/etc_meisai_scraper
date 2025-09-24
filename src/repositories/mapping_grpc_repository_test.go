package repositories

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yhonda-ohishi/etc_meisai/src/models"
)

func TestNewMappingGRPCRepository(t *testing.T) {
	tests := []struct {
		name   string
		client interface{}
	}{
		{
			name:   "valid client",
			client: &mockClient{},
		},
		{
			name:   "nil client",
			client: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := NewMappingGRPCRepository(tt.client)
			assert.NotNil(t, repo)
			assert.IsType(t, &MappingGRPCRepository{}, repo)
		})
	}
}

func TestMappingGRPCRepository_Create(t *testing.T) {
	repo := NewMappingGRPCRepository(&mockClient{})

	tests := []struct {
		name    string
		mapping *models.ETCMeisaiMapping
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid mapping",
			mapping: &models.ETCMeisaiMapping{
				ETCMeisaiID: 1,
				DTakoRowID:  "DTAKO001",
				MappingType: "automatic",
				Confidence:  0.95,
				Notes:       "High confidence match",
			},
			wantErr: true,
			errMsg:  "CreateMapping not available - clients package deleted",
		},
		{
			name:    "nil mapping",
			mapping: nil,
			wantErr: true,
			errMsg:  "CreateMapping not available - clients package deleted",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := repo.Create(tt.mapping)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestMappingGRPCRepository_GetByID(t *testing.T) {
	repo := NewMappingGRPCRepository(&mockClient{})

	tests := []struct {
		name    string
		id      int64
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid ID",
			id:      1,
			wantErr: true,
			errMsg:  "GetMapping not available - clients package deleted",
		},
		{
			name:    "zero ID",
			id:      0,
			wantErr: true,
			errMsg:  "GetMapping not available - clients package deleted",
		},
		{
			name:    "negative ID",
			id:      -1,
			wantErr: true,
			errMsg:  "GetMapping not available - clients package deleted",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := repo.GetByID(tt.id)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}
		})
	}
}

func TestMappingGRPCRepository_Update(t *testing.T) {
	repo := NewMappingGRPCRepository(&mockClient{})

	tests := []struct {
		name    string
		mapping *models.ETCMeisaiMapping
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid mapping",
			mapping: &models.ETCMeisaiMapping{
				ID:          1,
				ETCMeisaiID: 1,
				DTakoRowID:  "DTAKO001",
				MappingType: "manual",
				Confidence:  0.99,
				Notes:       "Updated confidence",
			},
			wantErr: true,
			errMsg:  "UpdateMapping not available - clients package deleted",
		},
		{
			name:    "nil mapping",
			mapping: nil,
			wantErr: true,
			errMsg:  "UpdateMapping not available - clients package deleted",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := repo.Update(tt.mapping)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.errMsg)
		})
	}
}

func TestMappingGRPCRepository_Delete(t *testing.T) {
	repo := NewMappingGRPCRepository(&mockClient{})

	tests := []struct {
		name   string
		id     int64
		errMsg string
	}{
		{
			name:   "valid ID",
			id:     1,
			errMsg: "DeleteMapping not available - clients package deleted",
		},
		{
			name:   "zero ID",
			id:     0,
			errMsg: "DeleteMapping not available - clients package deleted",
		},
		{
			name:   "negative ID",
			id:     -1,
			errMsg: "DeleteMapping not available - clients package deleted",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := repo.Delete(tt.id)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.errMsg)
		})
	}
}

func TestMappingGRPCRepository_GetByETCMeisaiID(t *testing.T) {
	repo := NewMappingGRPCRepository(&mockClient{})

	tests := []struct {
		name        string
		etcMeisaiID int64
		wantErr     bool
		errMsg      string
	}{
		{
			name:        "valid ETC Meisai ID",
			etcMeisaiID: 1,
			wantErr:     true,
			errMsg:      "ListMappings not available - clients package deleted",
		},
		{
			name:        "zero ETC Meisai ID",
			etcMeisaiID: 0,
			wantErr:     true,
			errMsg:      "ListMappings not available - clients package deleted",
		},
		{
			name:        "negative ETC Meisai ID",
			etcMeisaiID: -1,
			wantErr:     true,
			errMsg:      "ListMappings not available - clients package deleted",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := repo.GetByETCMeisaiID(tt.etcMeisaiID)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}
		})
	}
}

func TestMappingGRPCRepository_GetByDTakoRowID(t *testing.T) {
	repo := NewMappingGRPCRepository(&mockClient{})

	tests := []struct {
		name       string
		dtakoRowID string
		wantErr    bool
		errMsg     string
	}{
		{
			name:       "valid DTako row ID",
			dtakoRowID: "DTAKO001",
			wantErr:    true,
			errMsg:     "GetByDTakoRowID not available - clients package deleted",
		},
		{
			name:       "empty DTako row ID",
			dtakoRowID: "",
			wantErr:    true,
			errMsg:     "GetByDTakoRowID not available - clients package deleted",
		},
		{
			name:       "long DTako row ID",
			dtakoRowID: "DTAKO_VERY_LONG_ID_123456789",
			wantErr:    true,
			errMsg:     "GetByDTakoRowID not available - clients package deleted",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := repo.GetByDTakoRowID(tt.dtakoRowID)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}
		})
	}
}

func TestMappingGRPCRepository_List(t *testing.T) {
	repo := NewMappingGRPCRepository(&mockClient{})

	minConfidence := float32(0.8)

	tests := []struct {
		name    string
		params  *models.MappingListParams
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid params",
			params: &models.MappingListParams{
				Limit:         10,
				Offset:        0,
				ETCMeisaiID:   intPtr(1),
				DTakoRowID:    "DTAKO001",
				MappingType:   "automatic",
				MinConfidence: &minConfidence,
			},
			wantErr: true,
			errMsg:  "List mappings not available - clients package deleted",
		},
		{
			name: "params with only limit",
			params: &models.MappingListParams{
				Limit:  50,
				Offset: 20,
			},
			wantErr: true,
			errMsg:  "List mappings not available - clients package deleted",
		},
		{
			name:    "nil params",
			params:  nil,
			wantErr: true,
			errMsg:  "List mappings not available - clients package deleted",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, count, err := repo.List(tt.params)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				assert.Nil(t, result)
				assert.Equal(t, int64(0), count)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.GreaterOrEqual(t, count, int64(0))
			}
		})
	}
}

func TestMappingGRPCRepository_BulkCreateMappings(t *testing.T) {
	repo := NewMappingGRPCRepository(&mockClient{})

	mappings := []*models.ETCMeisaiMapping{
		{
			ETCMeisaiID: 1,
			DTakoRowID:  "DTAKO001",
			MappingType: "automatic",
			Confidence:  0.95,
			Notes:       "High confidence match",
		},
		{
			ETCMeisaiID: 2,
			DTakoRowID:  "DTAKO002",
			MappingType: "automatic",
			Confidence:  0.90,
			Notes:       "Good match",
		},
	}

	tests := []struct {
		name     string
		mappings []*models.ETCMeisaiMapping
		wantErr  bool
		errMsg   string
	}{
		{
			name:     "valid mappings",
			mappings: mappings,
			wantErr:  true,
			errMsg:   "BulkCreateMappings not available - clients package deleted",
		},
		{
			name:     "single mapping",
			mappings: mappings[:1],
			wantErr:  true,
			errMsg:   "BulkCreateMappings not available - clients package deleted",
		},
		{
			name:     "empty mappings",
			mappings: []*models.ETCMeisaiMapping{},
			wantErr:  true,
			errMsg:   "BulkCreateMappings not available - clients package deleted",
		},
		{
			name:     "nil mappings",
			mappings: nil,
			wantErr:  true,
			errMsg:   "BulkCreateMappings not available - clients package deleted",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := repo.BulkCreateMappings(tt.mappings)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestMappingGRPCRepository_DeleteByETCMeisaiID(t *testing.T) {
	repo := NewMappingGRPCRepository(&mockClient{})

	tests := []struct {
		name        string
		etcMeisaiID int64
		wantErr     bool
		errMsg      string
	}{
		{
			name:        "valid ETC Meisai ID",
			etcMeisaiID: 1,
			wantErr:     true,
			errMsg:      "DeleteMappingsByETCMeisai not available - clients package deleted",
		},
		{
			name:        "zero ETC Meisai ID",
			etcMeisaiID: 0,
			wantErr:     true,
			errMsg:      "DeleteMappingsByETCMeisai not available - clients package deleted",
		},
		{
			name:        "negative ETC Meisai ID",
			etcMeisaiID: -1,
			wantErr:     true,
			errMsg:      "DeleteMappingsByETCMeisai not available - clients package deleted",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := repo.DeleteByETCMeisaiID(tt.etcMeisaiID)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestMappingGRPCRepository_FindPotentialMatches(t *testing.T) {
	repo := NewMappingGRPCRepository(&mockClient{})

	tests := []struct {
		name        string
		etcMeisaiID int64
		threshold   float32
		wantErr     bool
		errMsg      string
	}{
		{
			name:        "valid parameters",
			etcMeisaiID: 1,
			threshold:   0.8,
			wantErr:     true,
			errMsg:      "FindPotentialMatches not available - clients package deleted",
		},
		{
			name:        "zero threshold",
			etcMeisaiID: 1,
			threshold:   0.0,
			wantErr:     true,
			errMsg:      "FindPotentialMatches not available - clients package deleted",
		},
		{
			name:        "high threshold",
			etcMeisaiID: 1,
			threshold:   0.99,
			wantErr:     true,
			errMsg:      "FindPotentialMatches not available - clients package deleted",
		},
		{
			name:        "threshold above 1.0",
			etcMeisaiID: 1,
			threshold:   1.5,
			wantErr:     true,
			errMsg:      "FindPotentialMatches not available - clients package deleted",
		},
		{
			name:        "negative threshold",
			etcMeisaiID: 1,
			threshold:   -0.1,
			wantErr:     true,
			errMsg:      "FindPotentialMatches not available - clients package deleted",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := repo.FindPotentialMatches(tt.etcMeisaiID, tt.threshold)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}
		})
	}
}

func TestMappingGRPCRepository_UpdateConfidenceScore(t *testing.T) {
	repo := NewMappingGRPCRepository(&mockClient{})

	tests := []struct {
		name       string
		id         int64
		confidence float32
		wantErr    bool
		errMsg     string
	}{
		{
			name:       "valid parameters",
			id:         1,
			confidence: 0.95,
			wantErr:    true,
			errMsg:     "UpdateMappingConfidence not available - clients package deleted",
		},
		{
			name:       "zero confidence",
			id:         1,
			confidence: 0.0,
			wantErr:    true,
			errMsg:     "UpdateMappingConfidence not available - clients package deleted",
		},
		{
			name:       "full confidence",
			id:         1,
			confidence: 1.0,
			wantErr:    true,
			errMsg:     "UpdateMappingConfidence not available - clients package deleted",
		},
		{
			name:       "confidence above 1.0",
			id:         1,
			confidence: 1.5,
			wantErr:    true,
			errMsg:     "UpdateMappingConfidence not available - clients package deleted",
		},
		{
			name:       "negative confidence",
			id:         1,
			confidence: -0.1,
			wantErr:    true,
			errMsg:     "UpdateMappingConfidence not available - clients package deleted",
		},
		{
			name:       "zero ID",
			id:         0,
			confidence: 0.95,
			wantErr:    true,
			errMsg:     "UpdateMappingConfidence not available - clients package deleted",
		},
		{
			name:       "negative ID",
			id:         -1,
			confidence: 0.95,
			wantErr:    true,
			errMsg:     "UpdateMappingConfidence not available - clients package deleted",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := repo.UpdateConfidenceScore(tt.id, tt.confidence)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// Helper function to create int64 pointer
func intPtr(i int64) *int64 {
	return &i
}
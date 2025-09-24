package repositories

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/yhonda-ohishi/etc_meisai/src/models"
)

func TestNewGRPCRepository(t *testing.T) {
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
			repo := NewGRPCRepository(tt.client)
			assert.NotNil(t, repo)
			assert.IsType(t, &GRPCRepository{}, repo)
		})
	}
}

func TestGRPCRepository_Create(t *testing.T) {
	repo := NewGRPCRepository(&mockClient{})

	tests := []struct {
		name    string
		etc     *models.ETCMeisai
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid ETC record",
			etc: &models.ETCMeisai{
				UseDate:   time.Now(),
				UseTime:   "10:30",
				EntryIC:   "東京IC",
				ExitIC:    "大阪IC",
				Amount:    1000,
				CarNumber: "品川123あ1234",
				ETCNumber: "1234567890",
			},
			wantErr: true,
			errMsg:  "CreateETCMeisai not available - clients package deleted",
		},
		{
			name:    "nil ETC record",
			etc:     nil,
			wantErr: true,
			errMsg:  "CreateETCMeisai not available - clients package deleted",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := repo.Create(tt.etc)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGRPCRepository_GetByID(t *testing.T) {
	repo := NewGRPCRepository(&mockClient{})

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
			errMsg:  "GetETCMeisai not available - clients package deleted",
		},
		{
			name:    "zero ID",
			id:      0,
			wantErr: true,
			errMsg:  "GetETCMeisai not available - clients package deleted",
		},
		{
			name:    "negative ID",
			id:      -1,
			wantErr: true,
			errMsg:  "GetETCMeisai not available - clients package deleted",
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

func TestGRPCRepository_Update(t *testing.T) {
	repo := NewGRPCRepository(&mockClient{})

	tests := []struct {
		name    string
		etc     *models.ETCMeisai
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid ETC record",
			etc: &models.ETCMeisai{
				ID:        1,
				UseDate:   time.Now(),
				UseTime:   "10:30",
				EntryIC:   "東京IC",
				ExitIC:    "大阪IC",
				Amount:    1000,
				CarNumber: "品川123あ1234",
				ETCNumber: "1234567890",
			},
			wantErr: true,
			errMsg:  "update operation not supported in gRPC-only mode",
		},
		{
			name:    "nil ETC record",
			etc:     nil,
			wantErr: true,
			errMsg:  "update operation not supported in gRPC-only mode",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := repo.Update(tt.etc)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.errMsg)
		})
	}
}

func TestGRPCRepository_Delete(t *testing.T) {
	repo := NewGRPCRepository(&mockClient{})

	tests := []struct {
		name   string
		id     int64
		errMsg string
	}{
		{
			name:   "valid ID",
			id:     1,
			errMsg: "delete operation not supported in gRPC-only mode",
		},
		{
			name:   "zero ID",
			id:     0,
			errMsg: "delete operation not supported in gRPC-only mode",
		},
		{
			name:   "negative ID",
			id:     -1,
			errMsg: "delete operation not supported in gRPC-only mode",
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

func TestGRPCRepository_GetByDateRange(t *testing.T) {
	repo := NewGRPCRepository(&mockClient{})

	from := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2025, 1, 31, 23, 59, 59, 999999999, time.UTC)

	tests := []struct {
		name    string
		from    time.Time
		to      time.Time
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid date range",
			from:    from,
			to:      to,
			wantErr: true,
			errMsg:  "ListETCMeisai not available - clients package deleted",
		},
		{
			name:    "same from and to date",
			from:    from,
			to:      from,
			wantErr: true,
			errMsg:  "ListETCMeisai not available - clients package deleted",
		},
		{
			name:    "to date before from date",
			from:    to,
			to:      from,
			wantErr: true,
			errMsg:  "ListETCMeisai not available - clients package deleted",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := repo.GetByDateRange(tt.from, tt.to)
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

func TestGRPCRepository_List(t *testing.T) {
	repo := NewGRPCRepository(&mockClient{})

	from := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2025, 1, 31, 23, 59, 59, 999999999, time.UTC)

	tests := []struct {
		name    string
		params  *models.ETCListParams
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid params",
			params: &models.ETCListParams{
				Limit:     10,
				Offset:    0,
				StartDate: &from,
				EndDate:   &to,
				CarNumber: "品川123あ1234",
				ETCNumber: "1234567890",
			},
			wantErr: true,
			errMsg:  "ListETCMeisai not available - clients package deleted",
		},
		{
			name: "params with only limit",
			params: &models.ETCListParams{
				Limit:  50,
				Offset: 20,
			},
			wantErr: true,
			errMsg:  "ListETCMeisai not available - clients package deleted",
		},
		{
			name:    "nil params",
			params:  nil,
			wantErr: true,
			errMsg:  "ListETCMeisai not available - clients package deleted",
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

func TestGRPCRepository_GetByHash(t *testing.T) {
	repo := NewGRPCRepository(&mockClient{})

	tests := []struct {
		name    string
		hash    string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid hash",
			hash:    "abcd1234",
			wantErr: true,
			errMsg:  "GetByHash not yet implemented in db_service",
		},
		{
			name:    "empty hash",
			hash:    "",
			wantErr: true,
			errMsg:  "GetByHash not yet implemented in db_service",
		},
		{
			name:    "long hash",
			hash:    "abcdefghijklmnopqrstuvwxyz1234567890",
			wantErr: true,
			errMsg:  "GetByHash not yet implemented in db_service",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := repo.GetByHash(tt.hash)
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

func TestGRPCRepository_BulkInsert(t *testing.T) {
	repo := NewGRPCRepository(&mockClient{})

	records := []*models.ETCMeisai{
		{
			UseDate:   time.Now(),
			UseTime:   "10:30",
			EntryIC:   "東京IC",
			ExitIC:    "大阪IC",
			Amount:    1000,
			CarNumber: "品川123あ1234",
			ETCNumber: "1234567890",
		},
		{
			UseDate:   time.Now(),
			UseTime:   "11:30",
			EntryIC:   "横浜IC",
			ExitIC:    "名古屋IC",
			Amount:    1500,
			CarNumber: "品川456あ5678",
			ETCNumber: "0987654321",
		},
	}

	tests := []struct {
		name    string
		records []*models.ETCMeisai
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid records",
			records: records,
			wantErr: true,
			errMsg:  "BulkCreateETCMeisai not available - clients package deleted",
		},
		{
			name:    "single record",
			records: records[:1],
			wantErr: true,
			errMsg:  "BulkCreateETCMeisai not available - clients package deleted",
		},
		{
			name:    "empty records",
			records: []*models.ETCMeisai{},
			wantErr: true,
			errMsg:  "BulkCreateETCMeisai not available - clients package deleted",
		},
		{
			name:    "nil records",
			records: nil,
			wantErr: true,
			errMsg:  "BulkCreateETCMeisai not available - clients package deleted",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := repo.BulkInsert(tt.records)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGRPCRepository_CheckDuplicatesByHash(t *testing.T) {
	repo := NewGRPCRepository(&mockClient{})

	tests := []struct {
		name   string
		hashes []string
	}{
		{
			name:   "valid hashes",
			hashes: []string{"hash1", "hash2", "hash3"},
		},
		{
			name:   "single hash",
			hashes: []string{"hash1"},
		},
		{
			name:   "empty hashes",
			hashes: []string{},
		},
		{
			name:   "nil hashes",
			hashes: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := repo.CheckDuplicatesByHash(tt.hashes)
			assert.NoError(t, err)
			assert.NotNil(t, result)

			if tt.hashes != nil {
				assert.Equal(t, len(tt.hashes), len(result))
				for _, hash := range tt.hashes {
					assert.Contains(t, result, hash)
					assert.False(t, result[hash])
				}
			} else {
				assert.Empty(t, result)
			}
		})
	}
}

func TestGRPCRepository_CountByDateRange(t *testing.T) {
	repo := NewGRPCRepository(&mockClient{})

	from := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2025, 1, 31, 23, 59, 59, 999999999, time.UTC)

	tests := []struct {
		name    string
		from    time.Time
		to      time.Time
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid date range",
			from:    from,
			to:      to,
			wantErr: true,
			errMsg:  "GetETCSummary not available - clients package deleted",
		},
		{
			name:    "same from and to date",
			from:    from,
			to:      from,
			wantErr: true,
			errMsg:  "GetETCSummary not available - clients package deleted",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			count, err := repo.CountByDateRange(tt.from, tt.to)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				assert.Equal(t, int64(0), count)
			} else {
				assert.NoError(t, err)
				assert.GreaterOrEqual(t, count, int64(0))
			}
		})
	}
}

func TestGRPCRepository_GetByETCNumber(t *testing.T) {
	repo := NewGRPCRepository(&mockClient{})

	tests := []struct {
		name      string
		etcNumber string
		limit     int
		wantErr   bool
		errMsg    string
	}{
		{
			name:      "valid ETC number",
			etcNumber: "1234567890",
			limit:     10,
			wantErr:   true,
			errMsg:    "ListETCMeisai by ETC number not available - clients package deleted",
		},
		{
			name:      "empty ETC number",
			etcNumber: "",
			limit:     10,
			wantErr:   true,
			errMsg:    "ListETCMeisai by ETC number not available - clients package deleted",
		},
		{
			name:      "zero limit",
			etcNumber: "1234567890",
			limit:     0,
			wantErr:   true,
			errMsg:    "ListETCMeisai by ETC number not available - clients package deleted",
		},
		{
			name:      "negative limit",
			etcNumber: "1234567890",
			limit:     -1,
			wantErr:   true,
			errMsg:    "ListETCMeisai by ETC number not available - clients package deleted",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := repo.GetByETCNumber(tt.etcNumber, tt.limit)
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

func TestGRPCRepository_GetByCarNumber(t *testing.T) {
	repo := NewGRPCRepository(&mockClient{})

	tests := []struct {
		name      string
		carNumber string
		limit     int
		wantErr   bool
		errMsg    string
	}{
		{
			name:      "valid car number",
			carNumber: "品川123あ1234",
			limit:     10,
			wantErr:   true,
			errMsg:    "ListETCMeisai by car number not available - clients package deleted",
		},
		{
			name:      "empty car number",
			carNumber: "",
			limit:     10,
			wantErr:   true,
			errMsg:    "ListETCMeisai by car number not available - clients package deleted",
		},
		{
			name:      "zero limit",
			carNumber: "品川123あ1234",
			limit:     0,
			wantErr:   true,
			errMsg:    "ListETCMeisai by car number not available - clients package deleted",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := repo.GetByCarNumber(tt.carNumber, tt.limit)
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

func TestGRPCRepository_GetSummaryByDateRange(t *testing.T) {
	repo := NewGRPCRepository(&mockClient{})

	from := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2025, 1, 31, 23, 59, 59, 999999999, time.UTC)

	tests := []struct {
		name    string
		from    time.Time
		to      time.Time
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid date range",
			from:    from,
			to:      to,
			wantErr: true,
			errMsg:  "GetETCSummary not available - clients package deleted",
		},
		{
			name:    "same from and to date",
			from:    from,
			to:      from,
			wantErr: true,
			errMsg:  "GetETCSummary not available - clients package deleted",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := repo.GetSummaryByDateRange(tt.from, tt.to)
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

// Mock client for testing
type mockClient struct{}
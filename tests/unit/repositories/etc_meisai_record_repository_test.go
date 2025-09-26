//go:build ignore

package repositories_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"go.uber.org/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yhonda-ohishi/etc_meisai/src/models"
	"github.com/yhonda-ohishi/etc_meisai/tests/mocks"
)

func TestETCMeisaiRecordRepository_GetAll(t *testing.T) {
	tests := []struct {
		name      string
		setupMock func(*mocks.MockETCMeisaiRecordRepository)
		want      []models.ETCMeisaiRecord
		wantErr   bool
		errMsg    string
	}{
		{
			name: "successful retrieval",
			setupMock: func(m *mocks.MockETCMeisaiRecordRepository) {
				etcNum1 := "1234567890"
				etcNum2 := "1234567890"
				dtakoRowID1 := int64(1001)
				dtakoRowID2 := int64(1002)

				records := []models.ETCMeisaiRecord{
					{
						ID:            1,
						Hash:          "hash1",
						Date:          time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
						Time:          "10:30:00",
						EntranceIC:    "Tokyo IC",
						ExitIC:        "Osaka IC",
						TollAmount:    1000,
						CarNumber:     "123-4567",
						ETCCardNumber: "1234567890123456",
						ETCNum:        &etcNum1,
						DtakoRowID:    &dtakoRowID1,
					},
					{
						ID:            2,
						Hash:          "hash2",
						Date:          time.Date(2025, 1, 2, 0, 0, 0, 0, time.UTC),
						Time:          "14:15:00",
						EntranceIC:    "Tokyo IC",
						ExitIC:        "Kyoto IC",
						TollAmount:    2000,
						CarNumber:     "123-4567",
						ETCCardNumber: "1234567890123456",
						ETCNum:        &etcNum2,
						DtakoRowID:    &dtakoRowID2,
					},
				}
				m.EXPECT().GetAll(gomock.Any()).Return(records, nil)
			},
			want: []models.ETCMeisaiRecord{
				{
					ID:          1,
					EtcNum:      "1234567890",
					UsageDate:   time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
					Amount:      1000,
					DtakoRowID:  "DTAKO001",
				},
				{
					ID:          2,
					EtcNum:      "1234567890",
					UsageDate:   time.Date(2025, 1, 2, 0, 0, 0, 0, time.UTC),
					Amount:      2000,
					DtakoRowID:  "DTAKO002",
				},
			},
			wantErr: false,
		},
		{
			name: "empty result",
			setupMock: func(m *mocks.MockETCMeisaiRecordRepository) {
				m.EXPECT().GetAll(gomock.Any()).Return([]models.ETCMeisaiRecord{}, nil)
			},
			want:    []models.ETCMeisaiRecord{},
			wantErr: false,
		},
		{
			name: "database error",
			setupMock: func(m *mocks.MockETCMeisaiRecordRepository) {
				m.EXPECT().GetAll(gomock.Any()).Return(nil, errors.New("connection timeout"))
			},
			want:    nil,
			wantErr: true,
			errMsg:  "connection timeout",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mocks.NewMockETCMeisaiRecordRepository(ctrl)
			tt.setupMock(mockRepo)

			got, err := mockRepo.GetAll(context.Background())

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

func TestETCMeisaiRecordRepository_GetByEtcNum(t *testing.T) {
	tests := []struct {
		name      string
		etcNum    string
		setupMock func(*mocks.MockETCMeisaiRecordRepository)
		want      []models.ETCMeisaiRecord
		wantErr   bool
		errMsg    string
	}{
		{
			name:   "records found",
			etcNum: "1234567890",
			setupMock: func(m *mocks.MockETCMeisaiRecordRepository) {
				records := []models.ETCMeisaiRecord{
					{
						ID:         1,
						EtcNum:     "1234567890",
						UsageDate:  time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
						Amount:     1000,
						DtakoRowID: "DTAKO001",
					},
					{
						ID:         2,
						EtcNum:     "1234567890",
						UsageDate:  time.Date(2025, 1, 2, 0, 0, 0, 0, time.UTC),
						Amount:     2000,
						DtakoRowID: "DTAKO002",
					},
				}
				m.EXPECT().GetByEtcNum(gomock.Any(), "1234567890").Return(records, nil)
			},
			want: []models.ETCMeisaiRecord{
				{
					ID:         1,
					EtcNum:     "1234567890",
					UsageDate:  time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
					Amount:     1000,
					DtakoRowID: "DTAKO001",
				},
				{
					ID:         2,
					EtcNum:     "1234567890",
					UsageDate:  time.Date(2025, 1, 2, 0, 0, 0, 0, time.UTC),
					Amount:     2000,
					DtakoRowID: "DTAKO002",
				},
			},
			wantErr: false,
		},
		{
			name:   "no records found",
			etcNum: "9999999999",
			setupMock: func(m *mocks.MockETCMeisaiRecordRepository) {
				m.EXPECT().GetByEtcNum(gomock.Any(), "9999999999").Return([]models.ETCMeisaiRecord{}, nil)
			},
			want:    []models.ETCMeisaiRecord{},
			wantErr: false,
		},
		{
			name:   "empty etc number",
			etcNum: "",
			setupMock: func(m *mocks.MockETCMeisaiRecordRepository) {
				m.EXPECT().GetByEtcNum(gomock.Any(), "").Return(nil, errors.New("etc_num cannot be empty"))
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

			mockRepo := mocks.NewMockETCMeisaiRecordRepository(ctrl)
			tt.setupMock(mockRepo)

			got, err := mockRepo.GetByEtcNum(context.Background(), tt.etcNum)

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

func TestETCMeisaiRecordRepository_GetByDateRange(t *testing.T) {
	tests := []struct {
		name      string
		startDate time.Time
		endDate   time.Time
		setupMock func(*mocks.MockETCMeisaiRecordRepository)
		want      []models.ETCMeisaiRecord
		wantErr   bool
		errMsg    string
	}{
		{
			name:      "records in range",
			startDate: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
			endDate:   time.Date(2025, 1, 31, 23, 59, 59, 0, time.UTC),
			setupMock: func(m *mocks.MockETCMeisaiRecordRepository) {
				records := []models.ETCMeisaiRecord{
					{
						ID:        1,
						EtcNum:    "1234567890",
						UsageDate: time.Date(2025, 1, 15, 10, 0, 0, 0, time.UTC),
						Amount:    1500,
					},
					{
						ID:        2,
						EtcNum:    "0987654321",
						UsageDate: time.Date(2025, 1, 20, 14, 30, 0, 0, time.UTC),
						Amount:    2500,
					},
				}
				m.EXPECT().GetByDateRange(gomock.Any(), gomock.Any(), gomock.Any()).Return(records, nil)
			},
			want: []models.ETCMeisaiRecord{
				{
					ID:        1,
					EtcNum:    "1234567890",
					UsageDate: time.Date(2025, 1, 15, 10, 0, 0, 0, time.UTC),
					Amount:    1500,
				},
				{
					ID:        2,
					EtcNum:    "0987654321",
					UsageDate: time.Date(2025, 1, 20, 14, 30, 0, 0, time.UTC),
					Amount:    2500,
				},
			},
			wantErr: false,
		},
		{
			name:      "no records in range",
			startDate: time.Date(2025, 2, 1, 0, 0, 0, 0, time.UTC),
			endDate:   time.Date(2025, 2, 28, 23, 59, 59, 0, time.UTC),
			setupMock: func(m *mocks.MockETCMeisaiRecordRepository) {
				m.EXPECT().GetByDateRange(gomock.Any(), gomock.Any(), gomock.Any()).Return([]models.ETCMeisaiRecord{}, nil)
			},
			want:    []models.ETCMeisaiRecord{},
			wantErr: false,
		},
		{
			name:      "invalid date range",
			startDate: time.Date(2025, 1, 31, 0, 0, 0, 0, time.UTC),
			endDate:   time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
			setupMock: func(m *mocks.MockETCMeisaiRecordRepository) {
				m.EXPECT().GetByDateRange(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, errors.New("invalid date range: start date after end date"))
			},
			wantErr: true,
			errMsg:  "invalid date range",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mocks.NewMockETCMeisaiRecordRepository(ctrl)
			tt.setupMock(mockRepo)

			got, err := mockRepo.GetByDateRange(context.Background(), tt.startDate, tt.endDate)

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

func TestETCMeisaiRecordRepository_BulkCreate(t *testing.T) {
	tests := []struct {
		name      string
		records   []models.ETCMeisaiRecord
		setupMock func(*mocks.MockETCMeisaiRecordRepository)
		wantErr   bool
		errMsg    string
	}{
		{
			name: "successful bulk insert",
			records: []models.ETCMeisaiRecord{
				{
					EtcNum:     "1234567890",
					UsageDate:  time.Date(2025, 1, 1, 10, 0, 0, 0, time.UTC),
					Amount:     1000,
					DtakoRowID: "DTAKO001",
				},
				{
					EtcNum:     "1234567890",
					UsageDate:  time.Date(2025, 1, 2, 11, 0, 0, 0, time.UTC),
					Amount:     2000,
					DtakoRowID: "DTAKO002",
				},
				{
					EtcNum:     "0987654321",
					UsageDate:  time.Date(2025, 1, 3, 12, 0, 0, 0, time.UTC),
					Amount:     3000,
					DtakoRowID: "DTAKO003",
				},
			},
			setupMock: func(m *mocks.MockETCMeisaiRecordRepository) {
				m.EXPECT().BulkCreate(gomock.Any(), gomock.Len(3)).Return(nil)
			},
			wantErr: false,
		},
		{
			name:    "empty records",
			records: []models.ETCMeisaiRecord{},
			setupMock: func(m *mocks.MockETCMeisaiRecordRepository) {
				m.EXPECT().BulkCreate(gomock.Any(), gomock.Len(0)).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "duplicate dtako_row_id",
			records: []models.ETCMeisaiRecord{
				{
					EtcNum:     "1234567890",
					UsageDate:  time.Date(2025, 1, 1, 10, 0, 0, 0, time.UTC),
					Amount:     1000,
					DtakoRowID: "DTAKO001",
				},
				{
					EtcNum:     "1234567890",
					UsageDate:  time.Date(2025, 1, 2, 11, 0, 0, 0, time.UTC),
					Amount:     2000,
					DtakoRowID: "DTAKO001", // Duplicate
				},
			},
			setupMock: func(m *mocks.MockETCMeisaiRecordRepository) {
				m.EXPECT().BulkCreate(gomock.Any(), gomock.Any()).Return(errors.New("duplicate key: dtako_row_id"))
			},
			wantErr: true,
			errMsg:  "duplicate key",
		},
		{
			name: "validation error",
			records: []models.ETCMeisaiRecord{
				{
					EtcNum:    "", // Invalid
					UsageDate: time.Date(2025, 1, 1, 10, 0, 0, 0, time.UTC),
					Amount:    1000,
				},
			},
			setupMock: func(m *mocks.MockETCMeisaiRecordRepository) {
				m.EXPECT().BulkCreate(gomock.Any(), gomock.Any()).Return(errors.New("validation failed: etc_num required"))
			},
			wantErr: true,
			errMsg:  "validation failed",
		},
		{
			name: "batch size exceeded",
			records: func() []models.ETCMeisaiRecord {
				// Create a large batch
				records := make([]models.ETCMeisaiRecord, 10001)
				for i := range records {
					records[i] = models.ETCMeisaiRecord{
						EtcNum:    "1234567890",
						UsageDate: time.Now(),
						Amount:    1000,
					}
				}
				return records
			}(),
			setupMock: func(m *mocks.MockETCMeisaiRecordRepository) {
				m.EXPECT().BulkCreate(gomock.Any(), gomock.Any()).Return(errors.New("batch size exceeded: max 10000"))
			},
			wantErr: true,
			errMsg:  "batch size exceeded",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mocks.NewMockETCMeisaiRecordRepository(ctrl)
			tt.setupMock(mockRepo)

			err := mockRepo.BulkCreate(context.Background(), tt.records)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestETCMeisaiRecordRepository_GetByDtakoRowID(t *testing.T) {
	tests := []struct {
		name       string
		dtakoRowID string
		setupMock  func(*mocks.MockETCMeisaiRecordRepository)
		want       *models.ETCMeisaiRecord
		wantErr    bool
		errMsg     string
	}{
		{
			name:       "existing record",
			dtakoRowID: "DTAKO001",
			setupMock: func(m *mocks.MockETCMeisaiRecordRepository) {
				record := &models.ETCMeisaiRecord{
					ID:         1,
					EtcNum:     "1234567890",
					UsageDate:  time.Date(2025, 1, 1, 10, 0, 0, 0, time.UTC),
					Amount:     1000,
					DtakoRowID: "DTAKO001",
				}
				m.EXPECT().GetByDtakoRowID(gomock.Any(), "DTAKO001").Return(record, nil)
			},
			want: &models.ETCMeisaiRecord{
				ID:         1,
				EtcNum:     "1234567890",
				UsageDate:  time.Date(2025, 1, 1, 10, 0, 0, 0, time.UTC),
				Amount:     1000,
				DtakoRowID: "DTAKO001",
			},
			wantErr: false,
		},
		{
			name:       "non-existent record",
			dtakoRowID: "DTAKO999",
			setupMock: func(m *mocks.MockETCMeisaiRecordRepository) {
				m.EXPECT().GetByDtakoRowID(gomock.Any(), "DTAKO999").Return(nil, ErrNotFound)
			},
			want:    nil,
			wantErr: true,
			errMsg:  "not found",
		},
		{
			name:       "empty dtako_row_id",
			dtakoRowID: "",
			setupMock: func(m *mocks.MockETCMeisaiRecordRepository) {
				m.EXPECT().GetByDtakoRowID(gomock.Any(), "").Return(nil, errors.New("dtako_row_id cannot be empty"))
			},
			want:    nil,
			wantErr: true,
			errMsg:  "dtako_row_id cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mocks.NewMockETCMeisaiRecordRepository(ctrl)
			tt.setupMock(mockRepo)

			got, err := mockRepo.GetByDtakoRowID(context.Background(), tt.dtakoRowID)

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

func TestETCMeisaiRecordRepository_UpdateProcessingStatus(t *testing.T) {
	tests := []struct {
		name      string
		id        uint
		status    string
		setupMock func(*mocks.MockETCMeisaiRecordRepository)
		wantErr   bool
		errMsg    string
	}{
		{
			name:   "successful status update",
			id:     1,
			status: "processed",
			setupMock: func(m *mocks.MockETCMeisaiRecordRepository) {
				m.EXPECT().UpdateProcessingStatus(gomock.Any(), uint(1), "processed").Return(nil)
			},
			wantErr: false,
		},
		{
			name:   "non-existent record",
			id:     999,
			status: "processed",
			setupMock: func(m *mocks.MockETCMeisaiRecordRepository) {
				m.EXPECT().UpdateProcessingStatus(gomock.Any(), uint(999), "processed").Return(ErrNotFound)
			},
			wantErr: true,
			errMsg:  "not found",
		},
		{
			name:   "invalid status",
			id:     1,
			status: "invalid_status",
			setupMock: func(m *mocks.MockETCMeisaiRecordRepository) {
				m.EXPECT().UpdateProcessingStatus(gomock.Any(), uint(1), "invalid_status").Return(errors.New("invalid status: must be pending, processing, processed, or failed"))
			},
			wantErr: true,
			errMsg:  "invalid status",
		},
		{
			name:   "concurrent update conflict",
			id:     1,
			status: "processing",
			setupMock: func(m *mocks.MockETCMeisaiRecordRepository) {
				m.EXPECT().UpdateProcessingStatus(gomock.Any(), uint(1), "processing").Return(errors.New("optimistic lock: record already modified"))
			},
			wantErr: true,
			errMsg:  "optimistic lock",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mocks.NewMockETCMeisaiRecordRepository(ctrl)
			tt.setupMock(mockRepo)

			err := mockRepo.UpdateProcessingStatus(context.Background(), tt.id, tt.status)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestETCMeisaiRecordRepository_DeleteByDateRange(t *testing.T) {
	tests := []struct {
		name      string
		startDate time.Time
		endDate   time.Time
		setupMock func(*mocks.MockETCMeisaiRecordRepository)
		wantCount int
		wantErr   bool
		errMsg    string
	}{
		{
			name:      "successful deletion",
			startDate: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			endDate:   time.Date(2024, 12, 31, 23, 59, 59, 0, time.UTC),
			setupMock: func(m *mocks.MockETCMeisaiRecordRepository) {
				m.EXPECT().DeleteByDateRange(gomock.Any(), gomock.Any(), gomock.Any()).Return(150, nil)
			},
			wantCount: 150,
			wantErr:   false,
		},
		{
			name:      "no records to delete",
			startDate: time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
			endDate:   time.Date(2020, 12, 31, 23, 59, 59, 0, time.UTC),
			setupMock: func(m *mocks.MockETCMeisaiRecordRepository) {
				m.EXPECT().DeleteByDateRange(gomock.Any(), gomock.Any(), gomock.Any()).Return(0, nil)
			},
			wantCount: 0,
			wantErr:   false,
		},
		{
			name:      "invalid date range",
			startDate: time.Date(2025, 1, 31, 0, 0, 0, 0, time.UTC),
			endDate:   time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
			setupMock: func(m *mocks.MockETCMeisaiRecordRepository) {
				m.EXPECT().DeleteByDateRange(gomock.Any(), gomock.Any(), gomock.Any()).Return(0, errors.New("invalid date range"))
			},
			wantCount: 0,
			wantErr:   true,
			errMsg:    "invalid date range",
		},
		{
			name:      "database lock timeout",
			startDate: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			endDate:   time.Date(2024, 12, 31, 23, 59, 59, 0, time.UTC),
			setupMock: func(m *mocks.MockETCMeisaiRecordRepository) {
				m.EXPECT().DeleteByDateRange(gomock.Any(), gomock.Any(), gomock.Any()).Return(0, errors.New("lock wait timeout exceeded"))
			},
			wantCount: 0,
			wantErr:   true,
			errMsg:    "lock wait timeout",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mocks.NewMockETCMeisaiRecordRepository(ctrl)
			tt.setupMock(mockRepo)

			count, err := mockRepo.DeleteByDateRange(context.Background(), tt.startDate, tt.endDate)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.wantCount, count)
			}
		})
	}
}
package grpc

import (
	"context"
	"errors"
	"io"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/yhonda-ohishi/etc_meisai/src/grpc"
	"github.com/yhonda-ohishi/etc_meisai/src/models"
	"github.com/yhonda-ohishi/etc_meisai/src/pb"
	"github.com/yhonda-ohishi/etc_meisai/src/services"
)

// Mock implementations
type MockETCMeisaiService struct {
	mock.Mock
}

func (m *MockETCMeisaiService) CreateRecord(ctx context.Context, params *services.CreateRecordParams) (*models.ETCMeisaiRecord, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.ETCMeisaiRecord), args.Error(1)
}

func (m *MockETCMeisaiService) GetRecord(ctx context.Context, id int64) (*models.ETCMeisaiRecord, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.ETCMeisaiRecord), args.Error(1)
}

func (m *MockETCMeisaiService) ListRecords(ctx context.Context, params *services.ListRecordsParams) (*services.ListRecordsResponse, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*services.ListRecordsResponse), args.Error(1)
}

func (m *MockETCMeisaiService) UpdateRecord(ctx context.Context, id int64, params *services.CreateRecordParams) (*models.ETCMeisaiRecord, error) {
	args := m.Called(ctx, id, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.ETCMeisaiRecord), args.Error(1)
}

func (m *MockETCMeisaiService) DeleteRecord(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockETCMeisaiService) HealthCheck(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

type MockETCMappingService struct {
	mock.Mock
}

func (m *MockETCMappingService) CreateMapping(ctx context.Context, params *services.CreateMappingParams) (*models.ETCMapping, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.ETCMapping), args.Error(1)
}

func (m *MockETCMappingService) GetMapping(ctx context.Context, id int64) (*models.ETCMapping, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.ETCMapping), args.Error(1)
}

func (m *MockETCMappingService) ListMappings(ctx context.Context, params *services.ListMappingsParams) (*services.ListMappingsResponse, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*services.ListMappingsResponse), args.Error(1)
}

func (m *MockETCMappingService) UpdateMapping(ctx context.Context, id int64, params *services.UpdateMappingParams) (*models.ETCMapping, error) {
	args := m.Called(ctx, id, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.ETCMapping), args.Error(1)
}

func (m *MockETCMappingService) DeleteMapping(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockETCMappingService) UpdateStatus(ctx context.Context, id int64, status string) error {
	args := m.Called(ctx, id, status)
	return args.Error(0)
}

func (m *MockETCMappingService) HealthCheck(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

type MockImportService struct {
	mock.Mock
}

func (m *MockImportService) ImportCSV(ctx context.Context, params *services.ImportCSVParams, reader io.Reader) (*services.ImportCSVResult, error) {
	args := m.Called(ctx, params, reader)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*services.ImportCSVResult), args.Error(1)
}

func (m *MockImportService) ImportCSVStream(ctx context.Context, params *services.ImportCSVStreamParams) (*services.ImportCSVResult, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*services.ImportCSVResult), args.Error(1)
}

func (m *MockImportService) GetImportSession(ctx context.Context, sessionID string) (*models.ImportSession, error) {
	args := m.Called(ctx, sessionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.ImportSession), args.Error(1)
}

func (m *MockImportService) ListImportSessions(ctx context.Context, params *services.ListImportSessionsParams) (*services.ListImportSessionsResponse, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*services.ListImportSessionsResponse), args.Error(1)
}

func (m *MockImportService) ProcessCSV(ctx context.Context, rows []*services.CSVRow, options *services.BulkProcessOptions) (*services.BulkProcessResult, error) {
	args := m.Called(ctx, rows, options)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*services.BulkProcessResult), args.Error(1)
}

func (m *MockImportService) ProcessCSVRow(ctx context.Context, row *services.CSVRow) (*models.ETCMeisaiRecord, error) {
	args := m.Called(ctx, row)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.ETCMeisaiRecord), args.Error(1)
}

func (m *MockImportService) HandleDuplicates(ctx context.Context, records []*models.ETCMeisaiRecord) ([]*services.DuplicateResult, error) {
	args := m.Called(ctx, records)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*services.DuplicateResult), args.Error(1)
}

func (m *MockImportService) CancelImportSession(ctx context.Context, sessionID string) error {
	args := m.Called(ctx, sessionID)
	return args.Error(0)
}

func (m *MockImportService) HealthCheck(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

type MockStatisticsService struct {
	mock.Mock
}

func (m *MockStatisticsService) GetGeneralStatistics(ctx context.Context, filter *services.StatisticsFilter) (*services.GeneralStatistics, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*services.GeneralStatistics), args.Error(1)
}

func (m *MockStatisticsService) GetDailyStatistics(ctx context.Context, filter *services.StatisticsFilter) (*services.DailyStatisticsResponse, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*services.DailyStatisticsResponse), args.Error(1)
}

func (m *MockStatisticsService) GetMonthlyStatistics(ctx context.Context, filter *services.StatisticsFilter) (*services.MonthlyStatisticsResponse, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*services.MonthlyStatisticsResponse), args.Error(1)
}

func (m *MockStatisticsService) GetVehicleStatistics(ctx context.Context, carNumbers []string, filter *services.StatisticsFilter) (*services.VehicleStatisticsResponse, error) {
	args := m.Called(ctx, carNumbers, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*services.VehicleStatisticsResponse), args.Error(1)
}

func (m *MockStatisticsService) GetMappingStatistics(ctx context.Context, filter *services.StatisticsFilter) (*services.MappingStatisticsResponse, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*services.MappingStatisticsResponse), args.Error(1)
}

func (m *MockStatisticsService) HealthCheck(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

type MockLogger struct {
	mock.Mock
}

func (m *MockLogger) Printf(format string, v ...interface{}) {
	m.Called(format, v)
}

func (m *MockLogger) Println(v ...interface{}) {
	m.Called(v)
}

func (m *MockLogger) Print(v ...interface{}) {
	m.Called(v)
}

func (m *MockLogger) Fatalf(format string, v ...interface{}) {
	m.Called(format, v)
}

func (m *MockLogger) Fatal(v ...interface{}) {
	m.Called(v)
}

func (m *MockLogger) Panicf(format string, v ...interface{}) {
	m.Called(format, v)
}

func (m *MockLogger) Panic(v ...interface{}) {
	m.Called(v)
}

// Test helper functions
func createTestServer() (*grpc.ETCMeisaiServer, *MockETCMeisaiService, *MockETCMappingService, *MockImportService, *MockStatisticsService, *MockLogger) {
	mockETCService := &MockETCMeisaiService{}
	mockMappingService := &MockETCMappingService{}
	mockImportService := &MockImportService{}
	mockStatsService := &MockStatisticsService{}
	mockLogger := &MockLogger{}

	server := grpc.NewETCMeisaiServer(
		mockETCService,
		mockMappingService,
		mockImportService,
		mockStatsService,
		mockLogger,
	)

	return server, mockETCService, mockMappingService, mockImportService, mockStatsService, mockLogger
}

func createTestRecord() *models.ETCMeisaiRecord {
	return &models.ETCMeisaiRecord{
		ID:            1,
		Date:          time.Date(2023, 12, 25, 0, 0, 0, 0, time.UTC),
		Time:          "14:30",
		EntranceIC:    "東京IC",
		ExitIC:        "大阪IC",
		TollAmount:    1500,
		CarNumber:     "123",
		ETCCardNumber: "1234567890",
		ETCNum:        stringPtr("999"),
		DtakoRowID:    intPtr(888),
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
}

func createTestProtoRecord() *pb.ETCMeisaiRecord {
	return &pb.ETCMeisaiRecord{
		Id:             1,
		Date:           "2023-12-25",
		Time:           "14:30",
		EntranceIc:     "東京IC",
		ExitIc:         "大阪IC",
		TollAmount:     1500,
		CarNumber:      "123",
		EtcCardNumber:  "1234567890",
		EtcNum:         stringPtr("999"),
		DtakoRowId:     intPtr(888),
	}
}

func intPtr(i int64) *int64 {
	return &i
}

func stringPtr(s string) *string {
	return &s
}

// TestNewETCMeisaiServer tests server creation
func TestNewETCMeisaiServer(t *testing.T) {
	t.Run("successful creation with all services", func(t *testing.T) {
		server, _, _, _, _, _ := createTestServer()
		assert.NotNil(t, server)
	})

	t.Run("panic with nil ETC service", func(t *testing.T) {
		assert.Panics(t, func() {
			grpc.NewETCMeisaiServer(nil, &MockETCMappingService{}, &MockImportService{}, &MockStatisticsService{}, &MockLogger{})
		})
	})

	t.Run("panic with nil mapping service", func(t *testing.T) {
		assert.Panics(t, func() {
			grpc.NewETCMeisaiServer(&MockETCMeisaiService{}, nil, &MockImportService{}, &MockStatisticsService{}, &MockLogger{})
		})
	})

	t.Run("panic with nil import service", func(t *testing.T) {
		assert.Panics(t, func() {
			grpc.NewETCMeisaiServer(&MockETCMeisaiService{}, &MockETCMappingService{}, nil, &MockStatisticsService{}, &MockLogger{})
		})
	})

	t.Run("panic with nil statistics service", func(t *testing.T) {
		assert.Panics(t, func() {
			grpc.NewETCMeisaiServer(&MockETCMeisaiService{}, &MockETCMappingService{}, &MockImportService{}, nil, &MockLogger{})
		})
	})

	t.Run("success with nil logger (uses default)", func(t *testing.T) {
		server := grpc.NewETCMeisaiServer(&MockETCMeisaiService{}, &MockETCMappingService{}, &MockImportService{}, &MockStatisticsService{}, nil)
		assert.NotNil(t, server)
	})
}

// TestCreateRecord tests record creation
func TestCreateRecord(t *testing.T) {
	ctx := context.Background()

	t.Run("successful record creation", func(t *testing.T) {
		server, mockETC, _, _, _, _ := createTestServer()

		mockRecord := createTestRecord()
		mockETC.On("CreateRecord", ctx, mock.AnythingOfType("*services.CreateRecordParams")).Return(mockRecord, nil)

		req := &pb.CreateRecordRequest{
			Record: createTestProtoRecord(),
		}

		resp, err := server.CreateRecord(ctx, req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.NotNil(t, resp.Record)
		assert.Equal(t, int64(1), resp.Record.Id)
		mockETC.AssertExpectations(t)
	})

	t.Run("nil request", func(t *testing.T) {
		server, _, _, _, _, _ := createTestServer()

		resp, err := server.CreateRecord(ctx, nil)

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Equal(t, codes.InvalidArgument, status.Code(err))
	})

	t.Run("nil record in request", func(t *testing.T) {
		server, _, _, _, _, _ := createTestServer()

		req := &pb.CreateRecordRequest{Record: nil}
		resp, err := server.CreateRecord(ctx, req)

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Equal(t, codes.InvalidArgument, status.Code(err))
	})

	t.Run("validation error - missing date", func(t *testing.T) {
		server, _, _, _, _, _ := createTestServer()

		record := createTestProtoRecord()
		record.Date = ""
		req := &pb.CreateRecordRequest{Record: record}

		resp, err := server.CreateRecord(ctx, req)

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Equal(t, codes.InvalidArgument, status.Code(err))
		assert.Contains(t, err.Error(), "date is required")
	})

	t.Run("validation error - missing time", func(t *testing.T) {
		server, _, _, _, _, _ := createTestServer()

		record := createTestProtoRecord()
		record.Time = ""
		req := &pb.CreateRecordRequest{Record: record}

		resp, err := server.CreateRecord(ctx, req)

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Equal(t, codes.InvalidArgument, status.Code(err))
		assert.Contains(t, err.Error(), "time is required")
	})

	t.Run("validation error - negative toll amount", func(t *testing.T) {
		server, _, _, _, _, _ := createTestServer()

		record := createTestProtoRecord()
		record.TollAmount = -100
		req := &pb.CreateRecordRequest{Record: record}

		resp, err := server.CreateRecord(ctx, req)

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Equal(t, codes.InvalidArgument, status.Code(err))
		assert.Contains(t, err.Error(), "toll_amount must be non-negative")
	})

	t.Run("service error", func(t *testing.T) {
		server, mockETC, _, _, _, mockLogger := createTestServer()

		mockETC.On("CreateRecord", ctx, mock.AnythingOfType("*services.CreateRecordParams")).Return(nil, errors.New("service error"))
		mockLogger.On("Printf", mock.Anything, mock.Anything).Return()

		req := &pb.CreateRecordRequest{
			Record: createTestProtoRecord(),
		}

		resp, err := server.CreateRecord(ctx, req)

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Equal(t, codes.Internal, status.Code(err))
		mockETC.AssertExpectations(t)
		mockLogger.AssertExpectations(t)
	})
}

// TestGetRecord tests record retrieval
func TestGetRecord(t *testing.T) {
	ctx := context.Background()

	t.Run("successful record retrieval", func(t *testing.T) {
		server, mockETC, _, _, _, _ := createTestServer()

		mockRecord := createTestRecord()
		mockETC.On("GetRecord", ctx, int64(1)).Return(mockRecord, nil)

		req := &pb.GetRecordRequest{Id: 1}
		resp, err := server.GetRecord(ctx, req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.NotNil(t, resp.Record)
		assert.Equal(t, int64(1), resp.Record.Id)
		mockETC.AssertExpectations(t)
	})

	t.Run("nil request", func(t *testing.T) {
		server, _, _, _, _, _ := createTestServer()

		resp, err := server.GetRecord(ctx, nil)

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Equal(t, codes.InvalidArgument, status.Code(err))
	})

	t.Run("invalid ID", func(t *testing.T) {
		server, _, _, _, _, _ := createTestServer()

		req := &pb.GetRecordRequest{Id: 0}
		resp, err := server.GetRecord(ctx, req)

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Equal(t, codes.InvalidArgument, status.Code(err))
	})

	t.Run("not found error", func(t *testing.T) {
		server, mockETC, _, _, _, mockLogger := createTestServer()

		mockETC.On("GetRecord", ctx, int64(999)).Return(nil, errors.New("record not found"))
		mockLogger.On("Printf", mock.Anything, mock.Anything).Return()

		req := &pb.GetRecordRequest{Id: 999}
		resp, err := server.GetRecord(ctx, req)

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Equal(t, codes.NotFound, status.Code(err))
		mockETC.AssertExpectations(t)
		mockLogger.AssertExpectations(t)
	})

	t.Run("service error", func(t *testing.T) {
		server, mockETC, _, _, _, mockLogger := createTestServer()

		mockETC.On("GetRecord", ctx, int64(1)).Return(nil, errors.New("database error"))
		mockLogger.On("Printf", mock.Anything, mock.Anything).Return()

		req := &pb.GetRecordRequest{Id: 1}
		resp, err := server.GetRecord(ctx, req)

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Equal(t, codes.Internal, status.Code(err))
		mockETC.AssertExpectations(t)
		mockLogger.AssertExpectations(t)
	})
}

// TestListRecords tests record listing
func TestListRecords(t *testing.T) {
	ctx := context.Background()

	t.Run("successful record listing", func(t *testing.T) {
		server, mockETC, _, _, _, _ := createTestServer()

		mockRecords := []*models.ETCMeisaiRecord{createTestRecord()}
		mockResponse := &services.ListRecordsResponse{
			Records:    mockRecords,
			TotalCount: 1,
		}
		mockETC.On("ListRecords", ctx, mock.AnythingOfType("*services.ListRecordsParams")).Return(mockResponse, nil)

		req := &pb.ListRecordsRequest{
			Page:     1,
			PageSize: 10,
		}
		resp, err := server.ListRecords(ctx, req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Len(t, resp.Records, 1)
		assert.Equal(t, int32(1), resp.TotalCount)
		mockETC.AssertExpectations(t)
	})

	t.Run("nil request", func(t *testing.T) {
		server, _, _, _, _, _ := createTestServer()

		resp, err := server.ListRecords(ctx, nil)

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Equal(t, codes.InvalidArgument, status.Code(err))
	})

	t.Run("default pagination", func(t *testing.T) {
		server, mockETC, _, _, _, _ := createTestServer()

		mockResponse := &services.ListRecordsResponse{
			Records:    []*models.ETCMeisaiRecord{},
			TotalCount: 0,
		}
		mockETC.On("ListRecords", ctx, mock.MatchedBy(func(params *services.ListRecordsParams) bool {
			return params.Page == 1 && params.PageSize == 50
		})).Return(mockResponse, nil)

		req := &pb.ListRecordsRequest{} // No pagination specified
		resp, err := server.ListRecords(ctx, req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, int32(1), resp.Page)
		assert.Equal(t, int32(50), resp.PageSize)
		mockETC.AssertExpectations(t)
	})

	t.Run("page size limit", func(t *testing.T) {
		server, mockETC, _, _, _, _ := createTestServer()

		mockResponse := &services.ListRecordsResponse{
			Records:    []*models.ETCMeisaiRecord{},
			TotalCount: 0,
		}
		mockETC.On("ListRecords", ctx, mock.MatchedBy(func(params *services.ListRecordsParams) bool {
			return params.PageSize == 1000 // Should be capped at 1000
		})).Return(mockResponse, nil)

		req := &pb.ListRecordsRequest{
			Page:     1,
			PageSize: 2000, // Too large
		}
		resp, err := server.ListRecords(ctx, req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, int32(1000), resp.PageSize)
		mockETC.AssertExpectations(t)
	})

	t.Run("service error", func(t *testing.T) {
		server, mockETC, _, _, _, mockLogger := createTestServer()

		mockETC.On("ListRecords", ctx, mock.AnythingOfType("*services.ListRecordsParams")).Return(nil, errors.New("database error"))
		mockLogger.On("Printf", mock.Anything, mock.Anything).Return()

		req := &pb.ListRecordsRequest{Page: 1, PageSize: 10}
		resp, err := server.ListRecords(ctx, req)

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Equal(t, codes.Internal, status.Code(err))
		mockETC.AssertExpectations(t)
		mockLogger.AssertExpectations(t)
	})
}

// TestUpdateRecord tests record updates
func TestUpdateRecord(t *testing.T) {
	ctx := context.Background()

	t.Run("successful record update", func(t *testing.T) {
		server, mockETC, _, _, _, _ := createTestServer()

		mockRecord := createTestRecord()
		mockETC.On("UpdateRecord", ctx, int64(1), mock.AnythingOfType("*services.CreateRecordParams")).Return(mockRecord, nil)

		req := &pb.UpdateRecordRequest{
			Id:     1,
			Record: createTestProtoRecord(),
		}
		resp, err := server.UpdateRecord(ctx, req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.NotNil(t, resp.Record)
		mockETC.AssertExpectations(t)
	})

	t.Run("nil request", func(t *testing.T) {
		server, _, _, _, _, _ := createTestServer()

		resp, err := server.UpdateRecord(ctx, nil)

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Equal(t, codes.InvalidArgument, status.Code(err))
	})

	t.Run("invalid ID", func(t *testing.T) {
		server, _, _, _, _, _ := createTestServer()

		req := &pb.UpdateRecordRequest{
			Id:     0,
			Record: createTestProtoRecord(),
		}
		resp, err := server.UpdateRecord(ctx, req)

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Equal(t, codes.InvalidArgument, status.Code(err))
	})

	t.Run("not found error", func(t *testing.T) {
		server, mockETC, _, _, _, mockLogger := createTestServer()

		mockETC.On("UpdateRecord", ctx, int64(999), mock.AnythingOfType("*services.CreateRecordParams")).Return(nil, errors.New("record not found"))
		mockLogger.On("Printf", mock.Anything, mock.Anything).Return()

		req := &pb.UpdateRecordRequest{
			Id:     999,
			Record: createTestProtoRecord(),
		}
		resp, err := server.UpdateRecord(ctx, req)

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Equal(t, codes.NotFound, status.Code(err))
		mockETC.AssertExpectations(t)
		mockLogger.AssertExpectations(t)
	})
}

// TestDeleteRecord tests record deletion
func TestDeleteRecord(t *testing.T) {
	ctx := context.Background()

	t.Run("successful record deletion", func(t *testing.T) {
		server, mockETC, _, _, _, _ := createTestServer()

		mockETC.On("DeleteRecord", ctx, int64(1)).Return(nil)

		req := &pb.DeleteRecordRequest{Id: 1}
		resp, err := server.DeleteRecord(ctx, req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.IsType(t, &emptypb.Empty{}, resp)
		mockETC.AssertExpectations(t)
	})

	t.Run("nil request", func(t *testing.T) {
		server, _, _, _, _, _ := createTestServer()

		resp, err := server.DeleteRecord(ctx, nil)

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Equal(t, codes.InvalidArgument, status.Code(err))
	})

	t.Run("invalid ID", func(t *testing.T) {
		server, _, _, _, _, _ := createTestServer()

		req := &pb.DeleteRecordRequest{Id: 0}
		resp, err := server.DeleteRecord(ctx, req)

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Equal(t, codes.InvalidArgument, status.Code(err))
	})

	t.Run("not found error", func(t *testing.T) {
		server, mockETC, _, _, _, mockLogger := createTestServer()

		mockETC.On("DeleteRecord", ctx, int64(999)).Return(errors.New("record not found"))
		mockLogger.On("Printf", mock.Anything, mock.Anything).Return()

		req := &pb.DeleteRecordRequest{Id: 999}
		resp, err := server.DeleteRecord(ctx, req)

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Equal(t, codes.NotFound, status.Code(err))
		mockETC.AssertExpectations(t)
		mockLogger.AssertExpectations(t)
	})
}

// TestImportCSV tests CSV import functionality
func TestImportCSV(t *testing.T) {
	ctx := context.Background()

	t.Run("successful CSV import", func(t *testing.T) {
		server, _, _, mockImport, _, _ := createTestServer()

		mockSession := &models.ImportSession{
			ID:           "session-123",
			AccountType:  "corporate",
			AccountID:    "account-1",
			FileName:     "test.csv",
			Status:       "completed",
			ProcessedRows: 10,
		}
		mockResult := &services.ImportCSVResult{
			Session:        mockSession,
			SuccessCount:   8,
			ErrorCount:     1,
			DuplicateCount: 1,
		}
		mockImport.On("ImportCSV", ctx, mock.AnythingOfType("*services.ImportCSVParams"), mock.AnythingOfType("*strings.Reader")).Return(mockResult, nil)

		req := &pb.ImportCSVRequest{
			AccountType: "corporate",
			AccountId:   "account-1",
			FileName:    "test.csv",
			FileContent: []byte("date,time,entry,exit,amount\n2023-12-25,14:30,東京IC,大阪IC,1500"),
		}
		resp, err := server.ImportCSV(ctx, req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.NotNil(t, resp.Session)
		mockImport.AssertExpectations(t)
	})

	t.Run("nil request", func(t *testing.T) {
		server, _, _, _, _, _ := createTestServer()

		resp, err := server.ImportCSV(ctx, nil)

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Equal(t, codes.InvalidArgument, status.Code(err))
	})

	t.Run("missing required fields", func(t *testing.T) {
		server, _, _, _, _, _ := createTestServer()

		tests := []struct {
			name    string
			request *pb.ImportCSVRequest
		}{
			{
				name: "missing account type",
				request: &pb.ImportCSVRequest{
					AccountId:   "account-1",
					FileName:    "test.csv",
					FileContent: []byte("test"),
				},
			},
			{
				name: "missing account ID",
				request: &pb.ImportCSVRequest{
					AccountType: "corporate",
					FileName:    "test.csv",
					FileContent: []byte("test"),
				},
			},
			{
				name: "missing file name",
				request: &pb.ImportCSVRequest{
					AccountType: "corporate",
					AccountId:   "account-1",
					FileContent: []byte("test"),
				},
			},
			{
				name: "empty file content",
				request: &pb.ImportCSVRequest{
					AccountType: "corporate",
					AccountId:   "account-1",
					FileName:    "test.csv",
					FileContent: []byte{},
				},
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				resp, err := server.ImportCSV(ctx, tt.request)
				assert.Error(t, err)
				assert.Nil(t, resp)
				assert.Equal(t, codes.InvalidArgument, status.Code(err))
			})
		}
	})

	t.Run("service error", func(t *testing.T) {
		server, _, _, mockImport, _, mockLogger := createTestServer()

		mockImport.On("ImportCSV", ctx, mock.AnythingOfType("*services.ImportCSVParams"), mock.AnythingOfType("*strings.Reader")).Return(nil, errors.New("import error"))
		mockLogger.On("Printf", mock.Anything, mock.Anything).Return()

		req := &pb.ImportCSVRequest{
			AccountType: "corporate",
			AccountId:   "account-1",
			FileName:    "test.csv",
			FileContent: []byte("test data"),
		}
		resp, err := server.ImportCSV(ctx, req)

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Equal(t, codes.Internal, status.Code(err))
		mockImport.AssertExpectations(t)
		mockLogger.AssertExpectations(t)
	})
}

// TestGetImportSession tests import session retrieval
func TestGetImportSession(t *testing.T) {
	ctx := context.Background()

	t.Run("successful session retrieval", func(t *testing.T) {
		server, _, _, mockImport, _, _ := createTestServer()

		mockSession := &models.ImportSession{
			ID:          "session-123",
			AccountType: "corporate",
			Status:      "completed",
		}
		mockImport.On("GetImportSession", ctx, "session-123").Return(mockSession, nil)

		req := &pb.GetImportSessionRequest{SessionId: "session-123"}
		resp, err := server.GetImportSession(ctx, req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.NotNil(t, resp.Session)
		mockImport.AssertExpectations(t)
	})

	t.Run("nil request", func(t *testing.T) {
		server, _, _, _, _, _ := createTestServer()

		resp, err := server.GetImportSession(ctx, nil)

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Equal(t, codes.InvalidArgument, status.Code(err))
	})

	t.Run("empty session ID", func(t *testing.T) {
		server, _, _, _, _, _ := createTestServer()

		req := &pb.GetImportSessionRequest{SessionId: ""}
		resp, err := server.GetImportSession(ctx, req)

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Equal(t, codes.InvalidArgument, status.Code(err))
	})

	t.Run("not found error", func(t *testing.T) {
		server, _, _, mockImport, _, mockLogger := createTestServer()

		mockImport.On("GetImportSession", ctx, "session-999").Return(nil, errors.New("session not found"))
		mockLogger.On("Printf", mock.Anything, mock.Anything).Return()

		req := &pb.GetImportSessionRequest{SessionId: "session-999"}
		resp, err := server.GetImportSession(ctx, req)

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Equal(t, codes.NotFound, status.Code(err))
		mockImport.AssertExpectations(t)
		mockLogger.AssertExpectations(t)
	})
}

// TestListImportSessions tests import session listing
func TestListImportSessions(t *testing.T) {
	ctx := context.Background()

	t.Run("successful session listing", func(t *testing.T) {
		server, _, _, mockImport, _, _ := createTestServer()

		mockSessions := []*models.ImportSession{
			{ID: "session-1", Status: "completed"},
			{ID: "session-2", Status: "processing"},
		}
		mockResponse := &services.ListImportSessionsResponse{
			Sessions:   mockSessions,
			TotalCount: 2,
		}
		mockImport.On("ListImportSessions", ctx, mock.AnythingOfType("*services.ListImportSessionsParams")).Return(mockResponse, nil)

		req := &pb.ListImportSessionsRequest{
			Page:     1,
			PageSize: 10,
		}
		resp, err := server.ListImportSessions(ctx, req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Len(t, resp.Sessions, 2)
		assert.Equal(t, int32(2), resp.TotalCount)
		mockImport.AssertExpectations(t)
	})

	t.Run("nil request", func(t *testing.T) {
		server, _, _, _, _, _ := createTestServer()

		resp, err := server.ListImportSessions(ctx, nil)

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Equal(t, codes.InvalidArgument, status.Code(err))
	})

	t.Run("default pagination", func(t *testing.T) {
		server, _, _, mockImport, _, _ := createTestServer()

		mockResponse := &services.ListImportSessionsResponse{
			Sessions:   []*models.ImportSession{},
			TotalCount: 0,
		}
		mockImport.On("ListImportSessions", ctx, mock.MatchedBy(func(params *services.ListImportSessionsParams) bool {
			return params.Page == 1 && params.PageSize == 50
		})).Return(mockResponse, nil)

		req := &pb.ListImportSessionsRequest{} // No pagination specified
		resp, err := server.ListImportSessions(ctx, req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, int32(1), resp.Page)
		assert.Equal(t, int32(50), resp.PageSize)
		mockImport.AssertExpectations(t)
	})
}

// TestCreateMapping tests mapping creation
func TestCreateMapping(t *testing.T) {
	ctx := context.Background()

	t.Run("successful mapping creation", func(t *testing.T) {
		server, _, mockMapping, _, _, _ := createTestServer()

		mockResult := &models.ETCMapping{
			ID:               1,
			ETCRecordID:      123,
			MappingType:      "auto",
			MappedEntityID:   456,
			MappedEntityType: "dtako_record",
			Confidence:       0.95,
			Status:           "active",
		}
		mockMapping.On("CreateMapping", ctx, mock.AnythingOfType("*services.CreateMappingParams")).Return(mockResult, nil)

		metadata, _ := structpb.NewStruct(map[string]interface{}{"test": "value"})
		req := &pb.CreateMappingRequest{
			Mapping: &pb.ETCMapping{
				EtcRecordId:      123,
				MappingType:      "auto",
				MappedEntityId:   456,
				MappedEntityType: "dtako_record",
				Confidence:       0.95,
				Status:           pb.MappingStatus_MAPPING_STATUS_ACTIVE,
				CreatedBy:        "system",
				Metadata:         metadata,
			},
		}
		resp, err := server.CreateMapping(ctx, req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.NotNil(t, resp.Mapping)
		mockMapping.AssertExpectations(t)
	})

	t.Run("nil request", func(t *testing.T) {
		server, _, _, _, _, _ := createTestServer()

		resp, err := server.CreateMapping(ctx, nil)

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Equal(t, codes.InvalidArgument, status.Code(err))
	})

	t.Run("validation error - invalid confidence", func(t *testing.T) {
		server, _, _, _, _, _ := createTestServer()

		req := &pb.CreateMappingRequest{
			Mapping: &pb.ETCMapping{
				EtcRecordId:      123,
				MappingType:      "auto",
				MappedEntityId:   456,
				MappedEntityType: "dtako_record",
				Confidence:       1.5, // Invalid confidence > 1
				Status:           pb.MappingStatus_MAPPING_STATUS_ACTIVE,
			},
		}
		resp, err := server.CreateMapping(ctx, req)

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Equal(t, codes.InvalidArgument, status.Code(err))
	})
}

// TestGetStatistics tests statistics retrieval
func TestGetStatistics(t *testing.T) {
	ctx := context.Background()

	t.Run("successful statistics retrieval", func(t *testing.T) {
		server, _, _, _, mockStats, _ := createTestServer()

		mockResult := &services.GeneralStatistics{
			TotalRecords:   100,
			TotalAmount:    150000,
			UniqueVehicles: 10,
			UniqueCards:    5,
		}
		mockStats.On("GetGeneralStatistics", ctx, mock.AnythingOfType("*services.StatisticsFilter")).Return(mockResult, nil)

		req := &pb.GetStatisticsRequest{
			DateFrom:      stringPtr("2023-01-01"),
			DateTo:        stringPtr("2023-12-31"),
			CarNumber:     stringPtr("123"),
			EtcCardNumber: stringPtr("1234567890"),
		}
		resp, err := server.GetStatistics(ctx, req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, int64(100), resp.TotalRecords)
		assert.Equal(t, int64(150000), resp.TotalAmount)
		assert.Equal(t, int32(10), resp.UniqueCars)
		assert.Equal(t, int32(5), resp.UniqueCards)
		mockStats.AssertExpectations(t)
	})

	t.Run("nil request", func(t *testing.T) {
		server, _, _, _, _, _ := createTestServer()

		resp, err := server.GetStatistics(ctx, nil)

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Equal(t, codes.InvalidArgument, status.Code(err))
	})

	t.Run("invalid date format", func(t *testing.T) {
		server, _, _, _, _, _ := createTestServer()

		req := &pb.GetStatisticsRequest{
			DateFrom: stringPtr("invalid-date"),
		}
		resp, err := server.GetStatistics(ctx, req)

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Equal(t, codes.InvalidArgument, status.Code(err))
	})

	t.Run("service error", func(t *testing.T) {
		server, _, _, _, mockStats, mockLogger := createTestServer()

		mockStats.On("GetGeneralStatistics", ctx, mock.AnythingOfType("*services.StatisticsFilter")).Return(nil, errors.New("stats error"))
		mockLogger.On("Printf", mock.Anything, mock.Anything).Return()

		req := &pb.GetStatisticsRequest{}
		resp, err := server.GetStatistics(ctx, req)

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Equal(t, codes.Internal, status.Code(err))
		mockStats.AssertExpectations(t)
		mockLogger.AssertExpectations(t)
	})
}

// TestValidationHelpers tests validation helper methods
func TestValidationHelpers(t *testing.T) {
	server, _, _, _, _, _ := createTestServer()

	t.Run("validateETCRecord - valid record", func(t *testing.T) {
		record := createTestProtoRecord()
		// We can't call the private method directly, so we test through CreateRecord
		req := &pb.CreateRecordRequest{Record: record}
		_, err := server.CreateRecord(context.Background(), req)
		// Should fail at service level since we haven't mocked it, but not at validation level
		assert.NotContains(t, err.Error(), "is required")
	})

	t.Run("validateETCMapping - valid mapping", func(t *testing.T) {
		mapping := &pb.ETCMapping{
			EtcRecordId:      123,
			MappingType:      "auto",
			MappedEntityId:   456,
			MappedEntityType: "dtako_record",
			Confidence:       0.95,
			Status:           pb.MappingStatus_MAPPING_STATUS_ACTIVE,
		}
		req := &pb.CreateMappingRequest{Mapping: mapping}
		_, err := server.CreateMapping(context.Background(), req)
		// Should fail at service level since we haven't mocked it, but not at validation level
		assert.NotContains(t, err.Error(), "is required")
	})
}

// TestPerformance tests performance aspects
func TestPerformance(t *testing.T) {
	t.Run("large record list performance", func(t *testing.T) {
		server, mockETC, _, _, _, _ := createTestServer()

		// Create a large number of mock records
		records := make([]*models.ETCMeisaiRecord, 1000)
		for i := 0; i < 1000; i++ {
			records[i] = createTestRecord()
			records[i].ID = int64(i + 1)
		}

		mockResponse := &services.ListRecordsResponse{
			Records:    records,
			TotalCount: 1000,
		}
		mockETC.On("ListRecords", mock.Anything, mock.AnythingOfType("*services.ListRecordsParams")).Return(mockResponse, nil)

		req := &pb.ListRecordsRequest{
			Page:     1,
			PageSize: 1000,
		}

		start := time.Now()
		resp, err := server.ListRecords(context.Background(), req)
		duration := time.Since(start)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Len(t, resp.Records, 1000)
		assert.Less(t, duration, 100*time.Millisecond, "Large record list should be processed quickly")
		mockETC.AssertExpectations(t)
	})

	t.Run("concurrent request handling", func(t *testing.T) {
		server, mockETC, _, _, _, _ := createTestServer()

		mockRecord := createTestRecord()
		mockETC.On("GetRecord", mock.Anything, mock.AnythingOfType("int64")).Return(mockRecord, nil)

		const numConcurrentRequests = 10
		done := make(chan bool, numConcurrentRequests)

		start := time.Now()
		for i := 0; i < numConcurrentRequests; i++ {
			go func(id int) {
				req := &pb.GetRecordRequest{Id: int64(id + 1)}
				_, err := server.GetRecord(context.Background(), req)
				assert.NoError(t, err)
				done <- true
			}(i)
		}

		// Wait for all requests to complete
		for i := 0; i < numConcurrentRequests; i++ {
			<-done
		}
		duration := time.Since(start)

		assert.Less(t, duration, 100*time.Millisecond, "Concurrent requests should be handled efficiently")
		mockETC.AssertExpectations(t)
	})
}

// TestErrorScenarios tests various error scenarios
func TestErrorScenarios(t *testing.T) {
	ctx := context.Background()

	t.Run("adapter conversion errors", func(t *testing.T) {
		server, mockETC, _, _, _, mockLogger := createTestServer()

		// Create a record that will cause adapter conversion issues
		invalidRecord := &models.ETCMeisaiRecord{
			ID:   1,
			Date: time.Time{}, // Invalid zero time
		}
		mockETC.On("CreateRecord", ctx, mock.AnythingOfType("*services.CreateRecordParams")).Return(invalidRecord, nil)
		mockLogger.On("Printf", mock.Anything, mock.Anything).Return()

		req := &pb.CreateRecordRequest{
			Record: createTestProtoRecord(),
		}

		resp, err := server.CreateRecord(ctx, req)

		// Should handle adapter conversion error gracefully
		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Equal(t, codes.Internal, status.Code(err))
		mockETC.AssertExpectations(t)
		mockLogger.AssertExpectations(t)
	})

	t.Run("context cancellation", func(t *testing.T) {
		server, mockETC, _, _, _, _ := createTestServer()

		cancelCtx, cancel := context.WithCancel(ctx)
		cancel() // Cancel immediately

		mockETC.On("GetRecord", cancelCtx, int64(1)).Return(nil, context.Canceled)

		req := &pb.GetRecordRequest{Id: 1}
		_, err := server.GetRecord(cancelCtx, req)

		assert.Error(t, err)
		// Should handle context cancellation properly
		mockETC.AssertExpectations(t)
	})

	t.Run("timeout handling", func(t *testing.T) {
		server, mockETC, _, _, _, mockLogger := createTestServer()

		timeoutCtx, cancel := context.WithTimeout(ctx, 1*time.Millisecond)
		defer cancel()

		// Simulate a slow service that times out
		mockETC.On("GetRecord", timeoutCtx, int64(1)).Return(nil, context.DeadlineExceeded)
		mockLogger.On("Printf", mock.Anything, mock.Anything).Return()

		req := &pb.GetRecordRequest{Id: 1}
		_, err := server.GetRecord(timeoutCtx, req)

		assert.Error(t, err)
		mockETC.AssertExpectations(t)
		mockLogger.AssertExpectations(t)
	})
}
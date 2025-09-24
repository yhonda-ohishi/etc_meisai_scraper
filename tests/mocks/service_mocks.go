package mocks

import (
	"context"
	"io"

	"github.com/stretchr/testify/mock"
	"github.com/yhonda-ohishi/etc_meisai/src/models"
	"github.com/yhonda-ohishi/etc_meisai/src/pb"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

// MockETCMeisaiService mocks the ETC Meisai service interface
type MockETCMeisaiService struct {
	mock.Mock
}

func (m *MockETCMeisaiService) ProcessETCData(ctx context.Context, data []byte) ([]*models.ETCMeisaiRecord, error) {
	args := m.Called(ctx, data)
	return args.Get(0).([]*models.ETCMeisaiRecord), args.Error(1)
}

func (m *MockETCMeisaiService) ImportCSV(ctx context.Context, reader io.Reader) (*models.ImportSession, error) {
	args := m.Called(ctx, reader)
	return args.Get(0).(*models.ImportSession), args.Error(1)
}

func (m *MockETCMeisaiService) GetRecord(ctx context.Context, id uint) (*models.ETCMeisaiRecord, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*models.ETCMeisaiRecord), args.Error(1)
}

func (m *MockETCMeisaiService) UpdateRecord(ctx context.Context, record *models.ETCMeisaiRecord) error {
	args := m.Called(ctx, record)
	return args.Error(0)
}

func (m *MockETCMeisaiService) DeleteRecord(ctx context.Context, id uint) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// MockETCMappingService mocks the ETC mapping service interface
type MockETCMappingService struct {
	mock.Mock
}

func (m *MockETCMappingService) CreateMapping(ctx context.Context, mapping *models.ETCMapping) error {
	args := m.Called(ctx, mapping)
	return args.Error(0)
}

func (m *MockETCMappingService) GetMappingByETCNum(ctx context.Context, etcNum string) (*models.ETCMapping, error) {
	args := m.Called(ctx, etcNum)
	return args.Get(0).(*models.ETCMapping), args.Error(1)
}

func (m *MockETCMappingService) UpdateMapping(ctx context.Context, mapping *models.ETCMapping) error {
	args := m.Called(ctx, mapping)
	return args.Error(0)
}

func (m *MockETCMappingService) DeleteMapping(ctx context.Context, id uint) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockETCMappingService) ListMappings(ctx context.Context, limit, offset int) ([]*models.ETCMapping, error) {
	args := m.Called(ctx, limit, offset)
	return args.Get(0).([]*models.ETCMapping), args.Error(1)
}

// MockImportService mocks the import service interface
type MockImportService struct {
	mock.Mock
}

func (m *MockImportService) StartImport(ctx context.Context, sessionID string) error {
	args := m.Called(ctx, sessionID)
	return args.Error(0)
}

func (m *MockImportService) GetImportStatus(ctx context.Context, sessionID string) (*models.ImportSession, error) {
	args := m.Called(ctx, sessionID)
	return args.Get(0).(*models.ImportSession), args.Error(1)
}

func (m *MockImportService) CancelImport(ctx context.Context, sessionID string) error {
	args := m.Called(ctx, sessionID)
	return args.Error(0)
}

// MockStatisticsService mocks the statistics service interface
type MockStatisticsService struct {
	mock.Mock
}

func (m *MockStatisticsService) GetDailyStatistics(ctx context.Context, date string) (*models.Statistics, error) {
	args := m.Called(ctx, date)
	return args.Get(0).(*models.Statistics), args.Error(1)
}

func (m *MockStatisticsService) GetMonthlyStatistics(ctx context.Context, year, month int) (*models.Statistics, error) {
	args := m.Called(ctx, year, month)
	return args.Get(0).(*models.Statistics), args.Error(1)
}

func (m *MockStatisticsService) GenerateReport(ctx context.Context, startDate, endDate string) ([]byte, error) {
	args := m.Called(ctx, startDate, endDate)
	return args.Get(0).([]byte), args.Error(1)
}

// MockGRPCClient mocks the gRPC client interface
type MockGRPCClient struct {
	mock.Mock
}

// ETC record operations
func (m *MockGRPCClient) CreateETCRecord(ctx context.Context, req *pb.CreateRecordRequest, opts ...grpc.CallOption) (*pb.CreateRecordResponse, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*pb.CreateRecordResponse), args.Error(1)
}

func (m *MockGRPCClient) GetETCRecord(ctx context.Context, req *pb.GetRecordRequest, opts ...grpc.CallOption) (*pb.GetRecordResponse, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*pb.GetRecordResponse), args.Error(1)
}

func (m *MockGRPCClient) ListETCRecords(ctx context.Context, req *pb.ListRecordsRequest, opts ...grpc.CallOption) (*pb.ListRecordsResponse, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*pb.ListRecordsResponse), args.Error(1)
}

func (m *MockGRPCClient) UpdateETCRecord(ctx context.Context, req *pb.UpdateRecordRequest, opts ...grpc.CallOption) (*pb.UpdateRecordResponse, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*pb.UpdateRecordResponse), args.Error(1)
}

func (m *MockGRPCClient) DeleteETCRecord(ctx context.Context, req *pb.DeleteRecordRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*emptypb.Empty), args.Error(1)
}

// Note: BulkCreateETCRecords, SearchETCRecords, and HealthCheck are not yet implemented in the protobuf definitions
// These methods are commented out until the corresponding protobuf messages are added

// func (m *MockGRPCClient) BulkCreateETCRecords(ctx context.Context, req *pb.BulkCreateETCRecordsRequest, opts ...grpc.CallOption) (*pb.BulkCreateETCRecordsResponse, error) {
//	args := m.Called(ctx, req)
//	return args.Get(0).(*pb.BulkCreateETCRecordsResponse), args.Error(1)
// }

// func (m *MockGRPCClient) SearchETCRecords(ctx context.Context, req *pb.SearchETCRecordsRequest, opts ...grpc.CallOption) (*pb.SearchETCRecordsResponse, error) {
//	args := m.Called(ctx, req)
//	return args.Get(0).(*pb.SearchETCRecordsResponse), args.Error(1)
// }

// func (m *MockGRPCClient) HealthCheck(ctx context.Context, req *pb.HealthCheckRequest, opts ...grpc.CallOption) (*pb.HealthCheckResponse, error) {
//	args := m.Called(ctx, req)
//	return args.Get(0).(*pb.HealthCheckResponse), args.Error(1)
// }

// Mapping operations
func (m *MockGRPCClient) CreateMapping(ctx context.Context, req *pb.CreateMappingRequest, opts ...grpc.CallOption) (*pb.CreateMappingResponse, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*pb.CreateMappingResponse), args.Error(1)
}

func (m *MockGRPCClient) GetMapping(ctx context.Context, req *pb.GetMappingRequest, opts ...grpc.CallOption) (*pb.GetMappingResponse, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*pb.GetMappingResponse), args.Error(1)
}

func (m *MockGRPCClient) ListMappings(ctx context.Context, req *pb.ListMappingsRequest, opts ...grpc.CallOption) (*pb.ListMappingsResponse, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*pb.ListMappingsResponse), args.Error(1)
}

func (m *MockGRPCClient) UpdateMapping(ctx context.Context, req *pb.UpdateMappingRequest, opts ...grpc.CallOption) (*pb.UpdateMappingResponse, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*pb.UpdateMappingResponse), args.Error(1)
}

func (m *MockGRPCClient) DeleteMapping(ctx context.Context, req *pb.DeleteMappingRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*emptypb.Empty), args.Error(1)
}

// Note: Additional mapping operations not yet implemented in protobuf definitions
// These methods are commented out until the corresponding protobuf messages are added

// func (m *MockGRPCClient) FindMappingsByETCRecord(ctx context.Context, req *pb.FindMappingsByETCRecordRequest, opts ...grpc.CallOption) (*pb.FindMappingsByETCRecordResponse, error) {
//	args := m.Called(ctx, req)
//	return args.Get(0).(*pb.FindMappingsByETCRecordResponse), args.Error(1)
// }

// func (m *MockGRPCClient) UpdateMappingStatus(ctx context.Context, req *pb.UpdateMappingStatusRequest, opts ...grpc.CallOption) (*pb.UpdateMappingStatusResponse, error) {
//	args := m.Called(ctx, req)
//	return args.Get(0).(*pb.UpdateMappingStatusResponse), args.Error(1)
// }

// func (m *MockGRPCClient) BulkUpdateMappingStatus(ctx context.Context, req *pb.BulkUpdateMappingStatusRequest, opts ...grpc.CallOption) (*pb.BulkUpdateMappingStatusResponse, error) {
//	args := m.Called(ctx, req)
//	return args.Get(0).(*pb.BulkUpdateMappingStatusResponse), args.Error(1)
// }

// func (m *MockGRPCClient) GetMappingStats(ctx context.Context, req *pb.GetMappingStatsRequest, opts ...grpc.CallOption) (*pb.GetMappingStatsResponse, error) {
//	args := m.Called(ctx, req)
//	return args.Get(0).(*pb.GetMappingStatsResponse), args.Error(1)
// }

// func (m *MockGRPCClient) CreateAutoMapping(ctx context.Context, req *pb.CreateAutoMappingRequest, opts ...grpc.CallOption) (*pb.CreateAutoMappingResponse, error) {
//	args := m.Called(ctx, req)
//	return args.Get(0).(*pb.CreateAutoMappingResponse), args.Error(1)
// }
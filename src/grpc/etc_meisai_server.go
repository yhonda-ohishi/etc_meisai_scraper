package grpc

import (
	"context"
	"log"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"

	pb "github.com/yhonda-ohishi/etc_meisai/src/pb"
	grpcServices "github.com/yhonda-ohishi/etc_meisai/src/services/grpc"
)

// ETCMeisaiServerStubs implements the ETCMeisaiServiceServer interface with stub methods
// TODO: This is a temporary implementation - needs proper integration with business services
type ETCMeisaiServerStubs struct {
	pb.UnimplementedETCMeisaiServiceServer
	meisaiBusinessService  *grpcServices.MeisaiBusinessServiceServer
	mappingBusinessService *grpcServices.MappingBusinessServiceServer
	logger                *log.Logger
}

// NewETCMeisaiServerStubs creates a new server with stub implementations
func NewETCMeisaiServerStubs(
	meisaiBusinessService *grpcServices.MeisaiBusinessServiceServer,
	mappingBusinessService *grpcServices.MappingBusinessServiceServer,
	logger *log.Logger,
) *ETCMeisaiServerStubs {
	if logger == nil {
		logger = log.New(log.Writer(), "[ETCMeisaiServerStubs] ", log.LstdFlags|log.Lshortfile)
	}

	return &ETCMeisaiServerStubs{
		meisaiBusinessService:  meisaiBusinessService,
		mappingBusinessService: mappingBusinessService,
		logger:                logger,
	}
}

// Record management methods - delegate to business services
func (s *ETCMeisaiServerStubs) CreateRecord(ctx context.Context, req *pb.CreateRecordRequest) (*pb.CreateRecordResponse, error) {
	// TODO: Implement delegation to meisaiBusinessService
	return nil, status.Error(codes.Unimplemented, "CreateRecord not fully implemented yet")
}

func (s *ETCMeisaiServerStubs) GetRecord(ctx context.Context, req *pb.GetRecordRequest) (*pb.GetRecordResponse, error) {
	// TODO: Implement delegation to business services
	return nil, status.Error(codes.Unimplemented, "GetRecord not fully implemented yet")
}

func (s *ETCMeisaiServerStubs) ListRecords(ctx context.Context, req *pb.ListRecordsRequest) (*pb.ListRecordsResponse, error) {
	// TODO: Implement delegation to business services
	return nil, status.Error(codes.Unimplemented, "ListRecords not fully implemented yet")
}

func (s *ETCMeisaiServerStubs) UpdateRecord(ctx context.Context, req *pb.UpdateRecordRequest) (*pb.UpdateRecordResponse, error) {
	// TODO: Implement delegation to business services
	return nil, status.Error(codes.Unimplemented, "UpdateRecord not fully implemented yet")
}

func (s *ETCMeisaiServerStubs) DeleteRecord(ctx context.Context, req *pb.DeleteRecordRequest) (*emptypb.Empty, error) {
	// TODO: Implement delegation to business services
	return nil, status.Error(codes.Unimplemented, "DeleteRecord not fully implemented yet")
}

// Import methods - delegate to business services
func (s *ETCMeisaiServerStubs) ImportCSV(ctx context.Context, req *pb.ImportCSVRequest) (*pb.ImportCSVResponse, error) {
	// TODO: Implement delegation to business services
	return nil, status.Error(codes.Unimplemented, "ImportCSV not fully implemented yet")
}

func (s *ETCMeisaiServerStubs) ImportCSVStream(stream pb.ETCMeisaiService_ImportCSVStreamServer) error {
	// TODO: Implement streaming import
	return status.Error(codes.Unimplemented, "ImportCSVStream not fully implemented yet")
}

func (s *ETCMeisaiServerStubs) GetImportSession(ctx context.Context, req *pb.GetImportSessionRequest) (*pb.GetImportSessionResponse, error) {
	// TODO: Implement delegation to business services
	return nil, status.Error(codes.Unimplemented, "GetImportSession not fully implemented yet")
}

func (s *ETCMeisaiServerStubs) ListImportSessions(ctx context.Context, req *pb.ListImportSessionsRequest) (*pb.ListImportSessionsResponse, error) {
	// TODO: Implement delegation to business services
	return nil, status.Error(codes.Unimplemented, "ListImportSessions not fully implemented yet")
}

// Mapping methods - delegate to business services
func (s *ETCMeisaiServerStubs) CreateMapping(ctx context.Context, req *pb.CreateMappingRequest) (*pb.CreateMappingResponse, error) {
	// TODO: Implement delegation to mappingBusinessService
	return nil, status.Error(codes.Unimplemented, "CreateMapping not fully implemented yet")
}

func (s *ETCMeisaiServerStubs) GetMapping(ctx context.Context, req *pb.GetMappingRequest) (*pb.GetMappingResponse, error) {
	// TODO: Implement delegation to business services
	return nil, status.Error(codes.Unimplemented, "GetMapping not fully implemented yet")
}

func (s *ETCMeisaiServerStubs) ListMappings(ctx context.Context, req *pb.ListMappingsRequest) (*pb.ListMappingsResponse, error) {
	// TODO: Implement delegation to business services
	return nil, status.Error(codes.Unimplemented, "ListMappings not fully implemented yet")
}

func (s *ETCMeisaiServerStubs) UpdateMapping(ctx context.Context, req *pb.UpdateMappingRequest) (*pb.UpdateMappingResponse, error) {
	// TODO: Implement delegation to business services
	return nil, status.Error(codes.Unimplemented, "UpdateMapping not fully implemented yet")
}

func (s *ETCMeisaiServerStubs) DeleteMapping(ctx context.Context, req *pb.DeleteMappingRequest) (*emptypb.Empty, error) {
	// TODO: Implement delegation to business services
	return nil, status.Error(codes.Unimplemented, "DeleteMapping not fully implemented yet")
}

// Statistics method
func (s *ETCMeisaiServerStubs) GetStatistics(ctx context.Context, req *pb.GetStatisticsRequest) (*pb.GetStatisticsResponse, error) {
	// TODO: Implement statistics aggregation
	return nil, status.Error(codes.Unimplemented, "GetStatistics not fully implemented yet")
}
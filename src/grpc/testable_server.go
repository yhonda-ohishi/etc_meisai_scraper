package grpc

import (
	"context"
	"fmt"
	"io"
	"log"
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/yhonda-ohishi/etc_meisai/src/adapters"
	"github.com/yhonda-ohishi/etc_meisai/src/pb"
	"github.com/yhonda-ohishi/etc_meisai/src/services"
)

// TestableETCMeisaiServer is a testable version of ETCMeisaiServer using interfaces
type TestableETCMeisaiServer struct {
	pb.UnimplementedETCMeisaiServiceServer
	etcMeisaiService  ETCMeisaiServiceInterface
	etcMappingService ETCMappingServiceInterface
	importService     ImportServiceInterface
	statisticsService StatisticsServiceInterface
	logger           *log.Logger
}

// NewTestableETCMeisaiServer creates a new testable gRPC server instance
func NewTestableETCMeisaiServer(
	etcMeisaiService ETCMeisaiServiceInterface,
	etcMappingService ETCMappingServiceInterface,
	importService ImportServiceInterface,
	statisticsService StatisticsServiceInterface,
	logger *log.Logger,
) *TestableETCMeisaiServer {
	if logger == nil {
		logger = log.New(log.Writer(), "[TestableETCMeisaiServer] ", log.LstdFlags|log.Lshortfile)
	}

	return &TestableETCMeisaiServer{
		etcMeisaiService:  etcMeisaiService,
		etcMappingService: etcMappingService,
		importService:     importService,
		statisticsService: statisticsService,
		logger:           logger,
	}
}

// CreateRecord creates a new ETC record
func (s *TestableETCMeisaiServer) CreateRecord(ctx context.Context, req *pb.CreateRecordRequest) (*pb.CreateRecordResponse, error) {
	if req == nil || req.Record == nil {
		return nil, status.Error(codes.InvalidArgument, "request and record cannot be nil")
	}

	// Create ETCMeisaiServer to reuse validation and conversion methods
	realServer := &ETCMeisaiServer{}

	// Validate required fields
	if err := realServer.validateETCRecord(req.Record); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// Convert proto to service parameters
	params, err := realServer.protoToCreateRecordParams(req.Record)
	if err != nil {
		s.logger.Printf("Error converting proto to params: %v", err)
		return nil, status.Error(codes.InvalidArgument, "invalid record data")
	}

	// Check if service is available
	if s.etcMeisaiService == nil {
		return nil, status.Error(codes.Internal, "service not available")
	}

	// Create record via service
	record, err := s.etcMeisaiService.CreateRecord(ctx, params)
	if err != nil {
		s.logger.Printf("Error creating record: %v", err)
		return nil, status.Error(codes.Internal, "failed to create record")
	}

	// Convert response back to proto
	protoRecord, err := adapters.ETCMeisaiRecordToProto(record)
	if err != nil {
		s.logger.Printf("Error converting record to proto: %v", err)
		return nil, status.Error(codes.Internal, "failed to convert response")
	}

	return &pb.CreateRecordResponse{
		Record: protoRecord,
	}, nil
}

// GetRecord retrieves an ETC record by ID
func (s *TestableETCMeisaiServer) GetRecord(ctx context.Context, req *pb.GetRecordRequest) (*pb.GetRecordResponse, error) {
	if req == nil || req.Id <= 0 {
		return nil, status.Error(codes.InvalidArgument, "invalid record ID")
	}

	// Check if service is available
	if s.etcMeisaiService == nil {
		return nil, status.Error(codes.Internal, "service not available")
	}

	record, err := s.etcMeisaiService.GetRecord(ctx, req.Id)
	if err != nil {
		s.logger.Printf("Error getting record: %v", err)
		if strings.Contains(err.Error(), "not found") {
			return nil, status.Error(codes.NotFound, "record not found")
		}
		return nil, status.Error(codes.Internal, "failed to get record")
	}

	protoRecord, err := adapters.ETCMeisaiRecordToProto(record)
	if err != nil {
		s.logger.Printf("Error converting record to proto: %v", err)
		return nil, status.Error(codes.Internal, "failed to convert response")
	}

	return &pb.GetRecordResponse{
		Record: protoRecord,
	}, nil
}

// ListRecords retrieves ETC records with pagination and filtering
func (s *TestableETCMeisaiServer) ListRecords(ctx context.Context, req *pb.ListRecordsRequest) (*pb.ListRecordsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request cannot be nil")
	}

	// Set default pagination
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 50
	}
	if req.PageSize > 1000 {
		req.PageSize = 1000
	}

	// Create ETCMeisaiServer to reuse conversion method
	realServer := &ETCMeisaiServer{}

	// Convert to service parameters
	params, err := realServer.protoToListRecordsParams(req)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// Check if service is available
	if s.etcMeisaiService == nil {
		return nil, status.Error(codes.Internal, "service not available")
	}

	response, err := s.etcMeisaiService.ListRecords(ctx, params)
	if err != nil {
		s.logger.Printf("Error listing records: %v", err)
		return nil, status.Error(codes.Internal, "failed to list records")
	}

	records := response.Records
	totalCount := response.TotalCount

	// Convert records to proto
	protoRecords := make([]*pb.ETCMeisaiRecord, len(records))
	for i, record := range records {
		protoRecord, err := adapters.ETCMeisaiRecordToProto(record)
		if err != nil {
			s.logger.Printf("Error converting record to proto: %v", err)
			return nil, status.Error(codes.Internal, "failed to convert records")
		}
		protoRecords[i] = protoRecord
	}

	return &pb.ListRecordsResponse{
		Records:    protoRecords,
		TotalCount: int32(totalCount),
		Page:       req.Page,
		PageSize:   req.PageSize,
	}, nil
}

// UpdateRecord updates an existing ETC record
func (s *TestableETCMeisaiServer) UpdateRecord(ctx context.Context, req *pb.UpdateRecordRequest) (*pb.UpdateRecordResponse, error) {
	if req == nil || req.Id <= 0 || req.Record == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request or record ID")
	}

	// Create ETCMeisaiServer to reuse validation and conversion methods
	realServer := &ETCMeisaiServer{}

	// Validate record data
	if err := realServer.validateETCRecord(req.Record); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// Convert proto to service parameters
	params, err := realServer.protoToCreateRecordParams(req.Record)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid record data")
	}

	// Check if service is available
	if s.etcMeisaiService == nil {
		return nil, status.Error(codes.Internal, "service not available")
	}

	record, err := s.etcMeisaiService.UpdateRecord(ctx, req.Id, params)
	if err != nil {
		s.logger.Printf("Error updating record: %v", err)
		if strings.Contains(err.Error(), "not found") {
			return nil, status.Error(codes.NotFound, "record not found")
		}
		return nil, status.Error(codes.Internal, "failed to update record")
	}

	protoRecord, err := adapters.ETCMeisaiRecordToProto(record)
	if err != nil {
		s.logger.Printf("Error converting record to proto: %v", err)
		return nil, status.Error(codes.Internal, "failed to convert response")
	}

	return &pb.UpdateRecordResponse{
		Record: protoRecord,
	}, nil
}

// DeleteRecord deletes an ETC record
func (s *TestableETCMeisaiServer) DeleteRecord(ctx context.Context, req *pb.DeleteRecordRequest) (*emptypb.Empty, error) {
	if req == nil || req.Id <= 0 {
		return nil, status.Error(codes.InvalidArgument, "invalid record ID")
	}

	// Check if service is available
	if s.etcMeisaiService == nil {
		return nil, status.Error(codes.Internal, "service not available")
	}

	err := s.etcMeisaiService.DeleteRecord(ctx, req.Id)
	if err != nil {
		s.logger.Printf("Error deleting record: %v", err)
		if strings.Contains(err.Error(), "not found") {
			return nil, status.Error(codes.NotFound, "record not found")
		}
		return nil, status.Error(codes.Internal, "failed to delete record")
	}

	return &emptypb.Empty{}, nil
}

// ImportCSV handles single CSV import request
func (s *TestableETCMeisaiServer) ImportCSV(ctx context.Context, req *pb.ImportCSVRequest) (*pb.ImportCSVResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request cannot be nil")
	}

	if req.AccountType == "" || req.AccountId == "" || req.FileName == "" {
		return nil, status.Error(codes.InvalidArgument, "account_type, account_id, and file_name are required")
	}

	if len(req.FileContent) == 0 {
		return nil, status.Error(codes.InvalidArgument, "file_content cannot be empty")
	}

	// Check if service is available
	if s.importService == nil {
		return nil, status.Error(codes.Internal, "service not available")
	}

	// Convert to service parameters
	params := &services.ImportCSVParams{
		AccountType: req.AccountType,
		AccountID:   req.AccountId,
		FileName:    req.FileName,
		FileSize:    int64(len(req.FileContent)),
	}

	result, err := s.importService.ImportCSV(ctx, params, strings.NewReader(string(req.FileContent)))
	if err != nil {
		s.logger.Printf("Error importing CSV: %v", err)
		return nil, status.Error(codes.Internal, "failed to import CSV")
	}

	protoSession, err := adapters.ImportSessionToProto(result.Session)
	if err != nil {
		s.logger.Printf("Error converting session to proto: %v", err)
		return nil, status.Error(codes.Internal, "failed to convert response")
	}

	return &pb.ImportCSVResponse{
		Session: protoSession,
	}, nil
}

// ImportCSVStream handles bidirectional streaming CSV import
func (s *TestableETCMeisaiServer) ImportCSVStream(stream pb.ETCMeisaiService_ImportCSVStreamServer) error {
	ctx := stream.Context()
	var sessionID string
	var chunks [][]byte
	totalChunks := 0

	// Receive chunks from client
	for {
		chunk, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			s.logger.Printf("Error receiving chunk: %v", err)
			return status.Error(codes.Internal, "failed to receive chunk")
		}

		if sessionID == "" {
			sessionID = chunk.SessionId
		} else if sessionID != chunk.SessionId {
			return status.Error(codes.InvalidArgument, "session ID mismatch")
		}

		chunks = append(chunks, chunk.Data)
		totalChunks++

		// Send progress update
		if err := stream.Send(&pb.ImportProgress{
			SessionId:          sessionID,
			ProcessedRows:      int32(totalChunks),
			ProgressPercentage: float32(totalChunks) * 10.0, // Simplified progress
			Status:            pb.ImportStatus_IMPORT_STATUS_PROCESSING,
		}); err != nil {
			s.logger.Printf("Error sending progress: %v", err)
			return status.Error(codes.Internal, "failed to send progress")
		}

		if chunk.IsLast {
			break
		}
	}

	// Check if service is available
	if s.importService == nil {
		errMsg := "service not available"
		stream.Send(&pb.ImportProgress{
			SessionId:    sessionID,
			Status:       pb.ImportStatus_IMPORT_STATUS_FAILED,
			ErrorMessage: &errMsg,
		})
		return status.Error(codes.Internal, errMsg)
	}

	// Process all chunks
	var allData []byte
	for _, chunk := range chunks {
		allData = append(allData, chunk...)
	}

	// Create import parameters (simplified - in real implementation, session should exist)
	params := &services.ImportCSVParams{
		AccountType: "corporate", // Default - in real implementation, get from session
		AccountID:   "default",   // Default - in real implementation, get from session
		FileName:    fmt.Sprintf("stream_%s.csv", sessionID),
		FileSize:    int64(len(allData)),
	}

	result, err := s.importService.ImportCSV(ctx, params, strings.NewReader(string(allData)))
	if err != nil {
		s.logger.Printf("Error processing stream import: %v", err)
		// Send error progress
		errMsg := err.Error()
		stream.Send(&pb.ImportProgress{
			SessionId:    sessionID,
			Status:       pb.ImportStatus_IMPORT_STATUS_FAILED,
			ErrorMessage: &errMsg,
		})
		return status.Error(codes.Internal, "failed to process import")
	}

	// Send final progress
	return stream.Send(&pb.ImportProgress{
		SessionId:          sessionID,
		ProcessedRows:      int32(result.Session.ProcessedRows),
		SuccessRows:        int32(result.SuccessCount),
		ErrorRows:          int32(result.ErrorCount),
		DuplicateRows:      int32(result.DuplicateCount),
		ProgressPercentage: 100.0,
		Status:            pb.ImportStatus_IMPORT_STATUS_COMPLETED,
	})
}

// GetImportSession retrieves an import session by ID
func (s *TestableETCMeisaiServer) GetImportSession(ctx context.Context, req *pb.GetImportSessionRequest) (*pb.GetImportSessionResponse, error) {
	if req == nil || req.SessionId == "" {
		return nil, status.Error(codes.InvalidArgument, "session_id is required")
	}

	// Check if service is available
	if s.importService == nil {
		return nil, status.Error(codes.Internal, "service not available")
	}

	session, err := s.importService.GetImportSession(ctx, req.SessionId)
	if err != nil {
		s.logger.Printf("Error getting import session: %v", err)
		if strings.Contains(err.Error(), "not found") {
			return nil, status.Error(codes.NotFound, "session not found")
		}
		return nil, status.Error(codes.Internal, "failed to get session")
	}

	protoSession, err := adapters.ImportSessionToProto(session)
	if err != nil {
		s.logger.Printf("Error converting session to proto: %v", err)
		return nil, status.Error(codes.Internal, "failed to convert response")
	}

	return &pb.GetImportSessionResponse{
		Session: protoSession,
	}, nil
}

// ListImportSessions retrieves import sessions with filtering
func (s *TestableETCMeisaiServer) ListImportSessions(ctx context.Context, req *pb.ListImportSessionsRequest) (*pb.ListImportSessionsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request cannot be nil")
	}

	// Set default pagination
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 50
	}

	// Check if service is available
	if s.importService == nil {
		return nil, status.Error(codes.Internal, "service not available")
	}

	// Convert to service parameters
	params := &services.ListImportSessionsParams{
		Page:     int(req.Page),
		PageSize: int(req.PageSize),
	}

	response, err := s.importService.ListImportSessions(ctx, params)
	if err != nil {
		s.logger.Printf("Error listing import sessions: %v", err)
		return nil, status.Error(codes.Internal, "failed to list sessions")
	}

	sessions := response.Sessions
	totalCount := response.TotalCount

	// Convert to proto
	protoSessions := make([]*pb.ImportSession, len(sessions))
	for i, session := range sessions {
		protoSession, err := adapters.ImportSessionToProto(session)
		if err != nil {
			s.logger.Printf("Error converting session to proto: %v", err)
			return nil, status.Error(codes.Internal, "failed to convert sessions")
		}
		protoSessions[i] = protoSession
	}

	return &pb.ListImportSessionsResponse{
		Sessions:   protoSessions,
		TotalCount: int32(totalCount),
		Page:       req.Page,
		PageSize:   req.PageSize,
	}, nil
}

// CreateMapping creates a new ETC mapping
func (s *TestableETCMeisaiServer) CreateMapping(ctx context.Context, req *pb.CreateMappingRequest) (*pb.CreateMappingResponse, error) {
	if req == nil || req.Mapping == nil {
		return nil, status.Error(codes.InvalidArgument, "request and mapping cannot be nil")
	}

	// Create ETCMeisaiServer to reuse validation and conversion methods
	realServer := &ETCMeisaiServer{}

	// Validate mapping
	if err := realServer.validateETCMapping(req.Mapping); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// Convert proto to service parameters
	params, err := realServer.protoToCreateMappingParams(req.Mapping)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid mapping data")
	}

	// Check if service is available
	if s.etcMappingService == nil {
		return nil, status.Error(codes.Internal, "service not available")
	}

	mapping, err := s.etcMappingService.CreateMapping(ctx, params)
	if err != nil {
		s.logger.Printf("Error creating mapping: %v", err)
		return nil, status.Error(codes.Internal, "failed to create mapping")
	}

	protoMapping, err := adapters.ETCMappingToProto(mapping)
	if err != nil {
		s.logger.Printf("Error converting mapping to proto: %v", err)
		return nil, status.Error(codes.Internal, "failed to convert response")
	}

	return &pb.CreateMappingResponse{
		Mapping: protoMapping,
	}, nil
}

// GetMapping retrieves an ETC mapping by ID
func (s *TestableETCMeisaiServer) GetMapping(ctx context.Context, req *pb.GetMappingRequest) (*pb.GetMappingResponse, error) {
	if req == nil || req.Id <= 0 {
		return nil, status.Error(codes.InvalidArgument, "invalid mapping ID")
	}

	// Check if service is available
	if s.etcMappingService == nil {
		return nil, status.Error(codes.Internal, "service not available")
	}

	mapping, err := s.etcMappingService.GetMapping(ctx, req.Id)
	if err != nil {
		s.logger.Printf("Error getting mapping: %v", err)
		if strings.Contains(err.Error(), "not found") {
			return nil, status.Error(codes.NotFound, "mapping not found")
		}
		return nil, status.Error(codes.Internal, "failed to get mapping")
	}

	protoMapping, err := adapters.ETCMappingToProto(mapping)
	if err != nil {
		s.logger.Printf("Error converting mapping to proto: %v", err)
		return nil, status.Error(codes.Internal, "failed to convert response")
	}

	return &pb.GetMappingResponse{
		Mapping: protoMapping,
	}, nil
}

// ListMappings retrieves ETC mappings with filtering
func (s *TestableETCMeisaiServer) ListMappings(ctx context.Context, req *pb.ListMappingsRequest) (*pb.ListMappingsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request cannot be nil")
	}

	// Set default pagination
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 50
	}

	// Create ETCMeisaiServer to reuse conversion method
	realServer := &ETCMeisaiServer{}

	// Convert to service parameters
	params, err := realServer.protoToListMappingsParams(req)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// Check if service is available
	if s.etcMappingService == nil {
		return nil, status.Error(codes.Internal, "service not available")
	}

	response, err := s.etcMappingService.ListMappings(ctx, params)
	if err != nil {
		s.logger.Printf("Error listing mappings: %v", err)
		return nil, status.Error(codes.Internal, "failed to list mappings")
	}

	mappings := response.Mappings
	totalCount := response.TotalCount

	// Convert to proto
	protoMappings := make([]*pb.ETCMapping, len(mappings))
	for i, mapping := range mappings {
		protoMapping, err := adapters.ETCMappingToProto(mapping)
		if err != nil {
			s.logger.Printf("Error converting mapping to proto: %v", err)
			return nil, status.Error(codes.Internal, "failed to convert mappings")
		}
		protoMappings[i] = protoMapping
	}

	return &pb.ListMappingsResponse{
		Mappings:   protoMappings,
		TotalCount: int32(totalCount),
		Page:       req.Page,
		PageSize:   req.PageSize,
	}, nil
}

// UpdateMapping updates an existing ETC mapping
func (s *TestableETCMeisaiServer) UpdateMapping(ctx context.Context, req *pb.UpdateMappingRequest) (*pb.UpdateMappingResponse, error) {
	if req == nil || req.Id <= 0 || req.Mapping == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request or mapping ID")
	}

	// Create ETCMeisaiServer to reuse validation and conversion methods
	realServer := &ETCMeisaiServer{}

	// Validate mapping
	if err := realServer.validateETCMapping(req.Mapping); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// Convert proto to service parameters
	params, err := realServer.protoToUpdateMappingParams(req.Id, req.Mapping)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid mapping data")
	}

	// Check if service is available
	if s.etcMappingService == nil {
		return nil, status.Error(codes.Internal, "service not available")
	}

	mapping, err := s.etcMappingService.UpdateMapping(ctx, req.Id, params)
	if err != nil {
		s.logger.Printf("Error updating mapping: %v", err)
		if strings.Contains(err.Error(), "not found") {
			return nil, status.Error(codes.NotFound, "mapping not found")
		}
		return nil, status.Error(codes.Internal, "failed to update mapping")
	}

	protoMapping, err := adapters.ETCMappingToProto(mapping)
	if err != nil {
		s.logger.Printf("Error converting mapping to proto: %v", err)
		return nil, status.Error(codes.Internal, "failed to convert response")
	}

	return &pb.UpdateMappingResponse{
		Mapping: protoMapping,
	}, nil
}

// DeleteMapping deletes an ETC mapping
func (s *TestableETCMeisaiServer) DeleteMapping(ctx context.Context, req *pb.DeleteMappingRequest) (*emptypb.Empty, error) {
	if req == nil || req.Id <= 0 {
		return nil, status.Error(codes.InvalidArgument, "invalid mapping ID")
	}

	// Check if service is available
	if s.etcMappingService == nil {
		return nil, status.Error(codes.Internal, "service not available")
	}

	err := s.etcMappingService.DeleteMapping(ctx, req.Id)
	if err != nil {
		s.logger.Printf("Error deleting mapping: %v", err)
		if strings.Contains(err.Error(), "not found") {
			return nil, status.Error(codes.NotFound, "mapping not found")
		}
		return nil, status.Error(codes.Internal, "failed to delete mapping")
	}

	return &emptypb.Empty{}, nil
}
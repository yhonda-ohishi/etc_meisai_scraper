package grpc

import (
	"context"
	"fmt"
	"io"
	"log"
	"strings"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/yhonda-ohishi/etc_meisai/src/adapters"
	"github.com/yhonda-ohishi/etc_meisai/src/pb"
	"github.com/yhonda-ohishi/etc_meisai/src/services"
)

// ETCMeisaiServer implements the ETCMeisaiServiceServer interface
type ETCMeisaiServer struct {
	pb.UnimplementedETCMeisaiServiceServer
	etcMeisaiService  ETCMeisaiServiceInterface
	etcMappingService ETCMappingServiceInterface
	importService     ImportServiceInterface
	statisticsService StatisticsServiceInterface
	logger           LoggerInterface
}

// NewETCMeisaiServer creates a new gRPC server instance with interface dependencies
func NewETCMeisaiServer(
	etcMeisaiService ETCMeisaiServiceInterface,
	etcMappingService ETCMappingServiceInterface,
	importService ImportServiceInterface,
	statisticsService StatisticsServiceInterface,
	logger LoggerInterface,
) *ETCMeisaiServer {
	// Validate all dependencies are non-nil
	if etcMeisaiService == nil {
		panic("etcMeisaiService cannot be nil")
	}
	if etcMappingService == nil {
		panic("etcMappingService cannot be nil")
	}
	if importService == nil {
		panic("importService cannot be nil")
	}
	if statisticsService == nil {
		panic("statisticsService cannot be nil")
	}
	if logger == nil {
		// Default logger if not provided
		logger = &defaultLogger{
			logger: log.New(log.Writer(), "[ETCMeisaiServer] ", log.LstdFlags|log.Lshortfile),
		}
	}

	return &ETCMeisaiServer{
		etcMeisaiService:  etcMeisaiService,
		etcMappingService: etcMappingService,
		importService:     importService,
		statisticsService: statisticsService,
		logger:           logger,
	}
}

// ========== Record Management Handlers (T037-T041) ==========

// CreateRecord creates a new ETC record (T037)
func (s *ETCMeisaiServer) CreateRecord(ctx context.Context, req *pb.CreateRecordRequest) (*pb.CreateRecordResponse, error) {
	if req == nil || req.Record == nil {
		return nil, status.Error(codes.InvalidArgument, "request and record cannot be nil")
	}

	// Validate required fields
	if err := s.validateETCRecord(req.Record); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// Convert proto to service parameters
	params, err := s.protoToCreateRecordParams(req.Record)
	if err != nil {
		s.logger.Printf("Error converting proto to params: %v", err)
		return nil, status.Error(codes.InvalidArgument, "invalid record data")
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

// GetRecord retrieves an ETC record by ID (T038)
func (s *ETCMeisaiServer) GetRecord(ctx context.Context, req *pb.GetRecordRequest) (*pb.GetRecordResponse, error) {
	if req == nil || req.Id <= 0 {
		return nil, status.Error(codes.InvalidArgument, "invalid record ID")
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

// ListRecords retrieves ETC records with pagination and filtering (T039)
func (s *ETCMeisaiServer) ListRecords(ctx context.Context, req *pb.ListRecordsRequest) (*pb.ListRecordsResponse, error) {
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

	// Convert to service parameters
	params, err := s.protoToListRecordsParams(req)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
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

// UpdateRecord updates an existing ETC record (T040)
func (s *ETCMeisaiServer) UpdateRecord(ctx context.Context, req *pb.UpdateRecordRequest) (*pb.UpdateRecordResponse, error) {
	if req == nil || req.Id <= 0 || req.Record == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request or record ID")
	}

	// Validate record data
	if err := s.validateETCRecord(req.Record); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// Convert proto to service parameters
	params, err := s.protoToCreateRecordParams(req.Record)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid record data")
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

// DeleteRecord deletes an ETC record (T041)
func (s *ETCMeisaiServer) DeleteRecord(ctx context.Context, req *pb.DeleteRecordRequest) (*emptypb.Empty, error) {
	if req == nil || req.Id <= 0 {
		return nil, status.Error(codes.InvalidArgument, "invalid record ID")
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

// ========== Import Handlers (T042-T043) ==========

// ImportCSV handles single CSV import request (T042)
func (s *ETCMeisaiServer) ImportCSV(ctx context.Context, req *pb.ImportCSVRequest) (*pb.ImportCSVResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request cannot be nil")
	}

	if req.AccountType == "" || req.AccountId == "" || req.FileName == "" {
		return nil, status.Error(codes.InvalidArgument, "account_type, account_id, and file_name are required")
	}

	if len(req.FileContent) == 0 {
		return nil, status.Error(codes.InvalidArgument, "file_content cannot be empty")
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

// ImportCSVStream handles bidirectional streaming CSV import (T043)
func (s *ETCMeisaiServer) ImportCSVStream(stream pb.ETCMeisaiService_ImportCSVStreamServer) error {
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
func (s *ETCMeisaiServer) GetImportSession(ctx context.Context, req *pb.GetImportSessionRequest) (*pb.GetImportSessionResponse, error) {
	if req == nil || req.SessionId == "" {
		return nil, status.Error(codes.InvalidArgument, "session_id is required")
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
func (s *ETCMeisaiServer) ListImportSessions(ctx context.Context, req *pb.ListImportSessionsRequest) (*pb.ListImportSessionsResponse, error) {
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

// ========== Mapping Handlers (T044) ==========

// CreateMapping creates a new ETC mapping
func (s *ETCMeisaiServer) CreateMapping(ctx context.Context, req *pb.CreateMappingRequest) (*pb.CreateMappingResponse, error) {
	if req == nil || req.Mapping == nil {
		return nil, status.Error(codes.InvalidArgument, "request and mapping cannot be nil")
	}

	// Validate mapping
	if err := s.validateETCMapping(req.Mapping); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// Convert proto to service parameters
	params, err := s.protoToCreateMappingParams(req.Mapping)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid mapping data")
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
func (s *ETCMeisaiServer) GetMapping(ctx context.Context, req *pb.GetMappingRequest) (*pb.GetMappingResponse, error) {
	if req == nil || req.Id <= 0 {
		return nil, status.Error(codes.InvalidArgument, "invalid mapping ID")
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
func (s *ETCMeisaiServer) ListMappings(ctx context.Context, req *pb.ListMappingsRequest) (*pb.ListMappingsResponse, error) {
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

	// Convert to service parameters
	params, err := s.protoToListMappingsParams(req)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
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
func (s *ETCMeisaiServer) UpdateMapping(ctx context.Context, req *pb.UpdateMappingRequest) (*pb.UpdateMappingResponse, error) {
	if req == nil || req.Id <= 0 || req.Mapping == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request or mapping ID")
	}

	// Validate mapping
	if err := s.validateETCMapping(req.Mapping); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// Convert proto to service parameters
	params, err := s.protoToUpdateMappingParams(req.Id, req.Mapping)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid mapping data")
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
func (s *ETCMeisaiServer) DeleteMapping(ctx context.Context, req *pb.DeleteMappingRequest) (*emptypb.Empty, error) {
	if req == nil || req.Id <= 0 {
		return nil, status.Error(codes.InvalidArgument, "invalid mapping ID")
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

// ========== Statistics Handler (T045) ==========

// GetStatistics retrieves ETC statistics
func (s *ETCMeisaiServer) GetStatistics(ctx context.Context, req *pb.GetStatisticsRequest) (*pb.GetStatisticsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request cannot be nil")
	}

	// Convert to service filter
	filter, err := s.protoToStatisticsFilter(req)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	stats, err := s.statisticsService.GetGeneralStatistics(ctx, filter)
	if err != nil {
		s.logger.Printf("Error getting statistics: %v", err)
		return nil, status.Error(codes.Internal, "failed to get statistics")
	}

	// Convert to proto response (simplified)
	return &pb.GetStatisticsResponse{
		TotalRecords: stats.TotalRecords,
		TotalAmount:  stats.TotalAmount,
		UniqueCars:   int32(stats.UniqueVehicles),
		UniqueCards:  int32(stats.UniqueCards),
		DailyStats:   []*pb.DailyStatistics{}, // TODO: Implement conversion
		IcStats:      []*pb.ICStatistics{},    // TODO: Implement conversion
	}, nil
}

// ========== Helper Methods ==========

// validateETCRecord validates required fields for ETC record
func (s *ETCMeisaiServer) validateETCRecord(record *pb.ETCMeisaiRecord) error {
	if record.Date == "" {
		return fmt.Errorf("date is required")
	}
	if record.Time == "" {
		return fmt.Errorf("time is required")
	}
	if record.EntranceIc == "" {
		return fmt.Errorf("entrance_ic is required")
	}
	if record.ExitIc == "" {
		return fmt.Errorf("exit_ic is required")
	}
	if record.CarNumber == "" {
		return fmt.Errorf("car_number is required")
	}
	if record.EtcCardNumber == "" {
		return fmt.Errorf("etc_card_number is required")
	}
	if record.TollAmount < 0 {
		return fmt.Errorf("toll_amount must be non-negative")
	}
	return nil
}

// validateETCMapping validates required fields for ETC mapping
func (s *ETCMeisaiServer) validateETCMapping(mapping *pb.ETCMapping) error {
	if mapping.EtcRecordId <= 0 {
		return fmt.Errorf("etc_record_id is required")
	}
	if mapping.MappingType == "" {
		return fmt.Errorf("mapping_type is required")
	}
	if mapping.MappedEntityId <= 0 {
		return fmt.Errorf("mapped_entity_id is required")
	}
	if mapping.MappedEntityType == "" {
		return fmt.Errorf("mapped_entity_type is required")
	}
	if mapping.Confidence < 0 || mapping.Confidence > 1 {
		return fmt.Errorf("confidence must be between 0 and 1")
	}
	return nil
}

// protoToCreateRecordParams converts proto record to service parameters
func (s *ETCMeisaiServer) protoToCreateRecordParams(record *pb.ETCMeisaiRecord) (*services.CreateRecordParams, error) {
	date, err := time.Parse("2006-01-02", record.Date)
	if err != nil {
		return nil, fmt.Errorf("invalid date format: %w", err)
	}

	params := &services.CreateRecordParams{
		Date:          date,
		Time:          record.Time,
		EntranceIC:    record.EntranceIc,
		ExitIC:        record.ExitIc,
		TollAmount:    int(record.TollAmount),
		CarNumber:     record.CarNumber,
		ETCCardNumber: record.EtcCardNumber,
	}

	if record.EtcNum != nil {
		params.ETCNum = record.EtcNum
	}
	if record.DtakoRowId != nil {
		params.DtakoRowID = record.DtakoRowId
	}

	return params, nil
}

// protoToListRecordsParams converts proto request to service parameters
func (s *ETCMeisaiServer) protoToListRecordsParams(req *pb.ListRecordsRequest) (*services.ListRecordsParams, error) {
	params := &services.ListRecordsParams{
		Page:     int(req.Page),
		PageSize: int(req.PageSize),
		SortBy:   req.SortBy,
	}

	// Handle sort order
	if req.SortOrder == pb.SortOrder_SORT_ORDER_DESC {
		params.SortOrder = "desc"
	} else {
		params.SortOrder = "asc"
	}

	// Handle date filters
	if req.DateFrom != nil && *req.DateFrom != "" {
		dateFrom, err := time.Parse("2006-01-02", *req.DateFrom)
		if err != nil {
			return nil, fmt.Errorf("invalid date_from format: %w", err)
		}
		params.DateFrom = &dateFrom
	}

	if req.DateTo != nil && *req.DateTo != "" {
		dateTo, err := time.Parse("2006-01-02", *req.DateTo)
		if err != nil {
			return nil, fmt.Errorf("invalid date_to format: %w", err)
		}
		params.DateTo = &dateTo
	}

	// Handle other filters
	if req.CarNumber != nil {
		params.CarNumber = req.CarNumber
	}
	if req.EtcCardNumber != nil {
		params.ETCNumber = req.EtcCardNumber
	}
	// Note: EntranceIC and ExitIC filtering not supported by ListRecordsParams
	// These would need to be added to the service layer if needed

	return params, nil
}

// protoToCreateMappingParams converts proto mapping to service parameters
func (s *ETCMeisaiServer) protoToCreateMappingParams(mapping *pb.ETCMapping) (*services.CreateMappingParams, error) {
	params := &services.CreateMappingParams{
		ETCRecordID:      mapping.EtcRecordId,
		MappingType:      mapping.MappingType,
		MappedEntityID:   mapping.MappedEntityId,
		MappedEntityType: mapping.MappedEntityType,
		Confidence:       mapping.Confidence,
		CreatedBy:        mapping.CreatedBy,
	}

	// Convert status enum to string
	switch mapping.Status {
	case pb.MappingStatus_MAPPING_STATUS_ACTIVE:
		params.Status = "active"
	case pb.MappingStatus_MAPPING_STATUS_INACTIVE:
		params.Status = "inactive"
	case pb.MappingStatus_MAPPING_STATUS_PENDING:
		params.Status = "pending"
	case pb.MappingStatus_MAPPING_STATUS_REJECTED:
		params.Status = "rejected"
	default:
		params.Status = "pending"
	}

	// Handle metadata conversion (simplified)
	if mapping.Metadata != nil {
		params.Metadata = mapping.Metadata.AsMap()
	}

	return params, nil
}

// protoToUpdateMappingParams converts proto mapping to update service parameters
func (s *ETCMeisaiServer) protoToUpdateMappingParams(id int64, mapping *pb.ETCMapping) (*services.UpdateMappingParams, error) {
	params := &services.UpdateMappingParams{
		MappingType:      &mapping.MappingType,
		MappedEntityID:   &mapping.MappedEntityId,
		MappedEntityType: &mapping.MappedEntityType,
		Confidence:       &mapping.Confidence,
	}

	// Convert status enum to string
	var status string
	switch mapping.Status {
	case pb.MappingStatus_MAPPING_STATUS_ACTIVE:
		status = "active"
	case pb.MappingStatus_MAPPING_STATUS_INACTIVE:
		status = "inactive"
	case pb.MappingStatus_MAPPING_STATUS_PENDING:
		status = "pending"
	case pb.MappingStatus_MAPPING_STATUS_REJECTED:
		status = "rejected"
	default:
		status = "pending"
	}
	params.Status = &status

	// Handle metadata conversion (simplified)
	if mapping.Metadata != nil {
		params.Metadata = mapping.Metadata.AsMap()
	}

	return params, nil
}

// protoToListMappingsParams converts proto request to service parameters
func (s *ETCMeisaiServer) protoToListMappingsParams(req *pb.ListMappingsRequest) (*services.ListMappingsParams, error) {
	params := &services.ListMappingsParams{
		Page:     int(req.Page),
		PageSize: int(req.PageSize),
	}

	if req.EtcRecordId != nil {
		params.ETCRecordID = req.EtcRecordId
	}
	if req.MappingType != nil {
		params.MappingType = req.MappingType
	}
	if req.MappedEntityId != nil {
		params.MappedEntityID = req.MappedEntityId
	}
	if req.MappedEntityType != nil {
		params.MappedEntityType = req.MappedEntityType
	}

	// Convert status enum to string
	if req.Status != nil {
		switch *req.Status {
		case pb.MappingStatus_MAPPING_STATUS_ACTIVE:
			status := "active"
			params.Status = &status
		case pb.MappingStatus_MAPPING_STATUS_INACTIVE:
			status := "inactive"
			params.Status = &status
		case pb.MappingStatus_MAPPING_STATUS_PENDING:
			status := "pending"
			params.Status = &status
		case pb.MappingStatus_MAPPING_STATUS_REJECTED:
			status := "rejected"
			params.Status = &status
		}
	}

	return params, nil
}

// protoToStatisticsFilter converts proto request to service filter
func (s *ETCMeisaiServer) protoToStatisticsFilter(req *pb.GetStatisticsRequest) (*services.StatisticsFilter, error) {
	filter := &services.StatisticsFilter{}

	if req.DateFrom != nil && *req.DateFrom != "" {
		dateFrom, err := time.Parse("2006-01-02", *req.DateFrom)
		if err != nil {
			return nil, fmt.Errorf("invalid date_from format: %w", err)
		}
		filter.DateFrom = &dateFrom
	}

	if req.DateTo != nil && *req.DateTo != "" {
		dateTo, err := time.Parse("2006-01-02", *req.DateTo)
		if err != nil {
			return nil, fmt.Errorf("invalid date_to format: %w", err)
		}
		filter.DateTo = &dateTo
	}

	if req.CarNumber != nil && *req.CarNumber != "" {
		filter.CarNumbers = []string{*req.CarNumber}
	}

	if req.EtcCardNumber != nil && *req.EtcCardNumber != "" {
		filter.ETCNumbers = []string{*req.EtcCardNumber}
	}

	return filter, nil
}
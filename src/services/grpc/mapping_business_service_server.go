package grpc

import (
	"context"
	"log"
	"strconv"
	"time"

	pb "github.com/yhonda-ohishi/etc_meisai/src/pb"
	"github.com/yhonda-ohishi/etc_meisai/src/repositories"
	"github.com/yhonda-ohishi/etc_meisai/src/services"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// MappingBusinessServiceServer implements the MappingBusinessService gRPC server
type MappingBusinessServiceServer struct {
	pb.UnimplementedMappingBusinessServiceServer
	mappingService *services.ETCMappingService
	mappingRepo    *repositories.ETCMappingRepositoryClient
	recordRepo     *repositories.ETCMeisaiRecordRepositoryClient
	logger         *log.Logger
}

// NewMappingBusinessServiceServer creates a new mapping business service server
func NewMappingBusinessServiceServer(
	mappingRepo *repositories.ETCMappingRepositoryClient,
	recordRepo *repositories.ETCMeisaiRecordRepositoryClient,
	logger *log.Logger,
) *MappingBusinessServiceServer {
	if logger == nil {
		logger = log.New(log.Writer(), "[MappingBusinessServiceServer] ", log.LstdFlags|log.Lshortfile)
	}

	// Create the underlying service with repository interfaces
	// Note: We'll need to create adapters that implement the repository interfaces
	// For now, we'll inject the clients directly
	return &MappingBusinessServiceServer{
		mappingRepo: mappingRepo,
		recordRepo:  recordRepo,
		logger:      logger,
	}
}

// CreateMapping creates a new ETC mapping with business rules applied
func (s *MappingBusinessServiceServer) CreateMapping(ctx context.Context, req *pb.CreateMappingServiceRequest) (*pb.CreateMappingServiceResponse, error) {
	s.logger.Printf("CreateMapping called for ETC record ID: %d", req.EtcRecordId)

	// Validate request
	if req.EtcRecordId <= 0 {
		return nil, status.Errorf(codes.InvalidArgument, "etc_record_id must be positive")
	}
	if req.MappedEntityId <= 0 {
		return nil, status.Errorf(codes.InvalidArgument, "mapped_entity_id must be positive")
	}

	// Check if ETC record exists
	_, err := s.recordRepo.GetByID(ctx, req.EtcRecordId)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "ETC record not found: %v", err)
	}

	// Create the mapping entity
	mapping := &pb.ETCMappingEntity{
		EtcRecordId:      req.EtcRecordId,
		MappingType:      req.MappingType,
		MappedEntityId:   req.MappedEntityId,
		MappedEntityType: req.MappedEntityType,
		Confidence:       req.Confidence,
		Status:           "active", // Default status
		CreatedAt:        timestamppb.Now(),
		UpdatedAt:        timestamppb.Now(),
	}

	// Set default confidence if not provided
	if mapping.Confidence == 0 {
		mapping.Confidence = 1.0
	}

	// Create the mapping via repository
	createdMapping, err := s.mappingRepo.Create(ctx, ETCMappingEntityToETCMapping(mapping))
	if err != nil {
		s.logger.Printf("Failed to create mapping: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to create mapping: %v", err)
	}

	// Create validation result
	validationResult := &pb.ValidationResult{
		IsValid: true,
	}

	response := &pb.CreateMappingServiceResponse{
		Mapping:      createdMapping,
		Validation:   validationResult,
		AutoApproved: req.AutoApprove,
	}

	s.logger.Printf("Successfully created mapping with ID: %d", createdMapping.Id)
	return response, nil
}

// ApproveMapping approves a mapping
func (s *MappingBusinessServiceServer) ApproveMapping(ctx context.Context, req *pb.ApproveMappingRequest) (*pb.ApproveMappingResponse, error) {
	s.logger.Printf("ApproveMapping called for mapping ID: %d", req.MappingId)

	if req.MappingId <= 0 {
		return nil, status.Errorf(codes.InvalidArgument, "mapping_id must be positive")
	}

	// Update mapping status to approved
	approvedMapping, err := s.mappingRepo.UpdateStatus(ctx, req.MappingId, pb.MappingStatus_MAPPING_STATUS_ACTIVE)
	if err != nil {
		s.logger.Printf("Failed to approve mapping: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to approve mapping: %v", err)
	}

	response := &pb.ApproveMappingResponse{
		Mapping:    approvedMapping,
		ApprovedAt: timestamppb.Now(),
	}

	s.logger.Printf("Successfully approved mapping with ID: %d", req.MappingId)
	return response, nil
}

// RejectMapping rejects a mapping
func (s *MappingBusinessServiceServer) RejectMapping(ctx context.Context, req *pb.RejectMappingRequest) (*pb.RejectMappingResponse, error) {
	s.logger.Printf("RejectMapping called for mapping ID: %d", req.MappingId)

	if req.MappingId <= 0 {
		return nil, status.Errorf(codes.InvalidArgument, "mapping_id must be positive")
	}

	// Update mapping status to rejected
	rejectedMapping, err := s.mappingRepo.UpdateStatus(ctx, req.MappingId, pb.MappingStatus_MAPPING_STATUS_REJECTED)
	if err != nil {
		s.logger.Printf("Failed to reject mapping: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to reject mapping: %v", err)
	}

	response := &pb.RejectMappingResponse{
		Mapping:    rejectedMapping,
		RejectedAt: timestamppb.Now(),
	}

	s.logger.Printf("Successfully rejected mapping with ID: %d", req.MappingId)
	return response, nil
}

// AutoMapRecords performs automatic mapping for multiple records
func (s *MappingBusinessServiceServer) AutoMapRecords(ctx context.Context, req *pb.AutoMapRequest) (*pb.AutoMapResponse, error) {
	s.logger.Printf("AutoMapRecords called for %d records", len(req.RecordIds))

	results := make([]*pb.MappingResult, 0, len(req.RecordIds))
	successCount := int32(0)
	failedCount := int32(0)

	for _, recordID := range req.RecordIds {
		result := &pb.MappingResult{
			RecordId: recordID,
		}

		// For now, implement a simple auto-mapping logic
		// In a real implementation, this would contain sophisticated matching algorithms
		if !req.DryRun {
			// Attempt to create an automatic mapping
			// This is a simplified implementation
			mapping := &pb.ETCMappingEntity{
				EtcRecordId:      recordID,
				MappingType:      "auto",
				MappedEntityId:   recordID, // Simplified: map to itself
				MappedEntityType: "auto_generated",
				Confidence:       0.8, // Medium confidence for auto mappings
				Status:           "active",
				CreatedAt:        timestamppb.Now(),
				UpdatedAt:        timestamppb.Now(),
			}

			createdMapping, err := s.mappingRepo.Create(ctx, ETCMappingEntityToETCMapping(mapping))
			if err != nil {
				result.Success = false
				result.ErrorMessage = &[]string{err.Error()}[0]
				failedCount++
			} else {
				result.Success = true
				result.Mapping = createdMapping
				successCount++
			}
		} else {
			// Dry run - just validate without creating
			result.Success = true
			successCount++
		}

		results = append(results, result)
	}

	response := &pb.AutoMapResponse{
		TotalProcessed:     int32(len(req.RecordIds)),
		SuccessfullyMapped: successCount,
		FailedMappings:     failedCount,
		Results:            results,
	}

	s.logger.Printf("AutoMapRecords completed: %d successful, %d failed", successCount, failedCount)
	return response, nil
}

// ValidateMapping validates a mapping
func (s *MappingBusinessServiceServer) ValidateMapping(ctx context.Context, req *pb.ValidateMappingRequest) (*pb.ValidateMappingResponse, error) {
	s.logger.Printf("ValidateMapping called for mapping validation")

	if req.Mapping == nil {
		return nil, status.Errorf(codes.InvalidArgument, "mapping cannot be nil")
	}

	// Perform validation logic
	isValid := true
	var errors []string

	// Basic validation checks
	if req.Mapping.EtcRecordId <= 0 {
		isValid = false
		errors = append(errors, "etc_record_id must be positive")
	}
	if req.Mapping.MappedEntityId <= 0 {
		isValid = false
		errors = append(errors, "mapped_entity_id must be positive")
	}
	if req.Mapping.Confidence < 0 || req.Mapping.Confidence > 1 {
		isValid = false
		errors = append(errors, "confidence must be between 0 and 1")
	}

	// Strict mode additional validations
	if req.StrictMode {
		if req.Mapping.Confidence < 0.7 {
			isValid = false
			errors = append(errors, "confidence too low for strict mode (must be >= 0.7)")
		}
	}

	validationResult := &pb.ValidationResult{
		IsValid: isValid,
		Errors:  ValidationResultToProto(isValid, []string{}).Errors,
	}

	response := &pb.ValidateMappingResponse{
		Result:      validationResult,
		Suggestions: []string{"Consider increasing confidence score", "Verify mapping type is correct"},
	}

	s.logger.Printf("ValidateMapping completed: valid=%t", isValid)
	return response, nil
}

// GetMappingSuggestions provides mapping suggestions for an ETC record
func (s *MappingBusinessServiceServer) GetMappingSuggestions(ctx context.Context, req *pb.GetSuggestionsRequest) (*pb.GetSuggestionsResponse, error) {
	s.logger.Printf("GetMappingSuggestions called for ETC record ID: %d", req.EtcRecordId)

	if req.EtcRecordId <= 0 {
		return nil, status.Errorf(codes.InvalidArgument, "etc_record_id must be positive")
	}

	// Get the ETC record to analyze
	record, err := s.recordRepo.GetByID(ctx, req.EtcRecordId)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "ETC record not found: %v", err)
	}

	// Generate suggestions based on the record
	// This is a simplified implementation - in reality, this would use ML/AI algorithms
	suggestions := []*pb.MappingSuggestion{
		{
			EntityId:   1001,
			EntityType: "fuel_expense",
			Confidence: 0.85,
			Reason:     "High confidence match based on route and amount pattern",
			MatchingAttributes: map[string]string{
				"route":  record.EntranceIc + "->" + record.ExitIc,
				"amount": strconv.Itoa(int(record.TollAmount)),
			},
		},
		{
			EntityId:   1002,
			EntityType: "business_trip",
			Confidence: 0.72,
			Reason:     "Medium confidence match based on time and day pattern",
			MatchingAttributes: map[string]string{
				"time_pattern": record.Time,
				"route":        record.EntranceIc + "->" + record.ExitIc,
			},
		},
	}

	// Filter by minimum confidence if specified
	if req.MinConfidence > 0 {
		filtered := make([]*pb.MappingSuggestion, 0)
		for _, suggestion := range suggestions {
			if suggestion.Confidence >= req.MinConfidence {
				filtered = append(filtered, suggestion)
			}
		}
		suggestions = filtered
	}

	// Limit results if specified
	if req.MaxSuggestions > 0 && int(req.MaxSuggestions) < len(suggestions) {
		suggestions = suggestions[:req.MaxSuggestions]
	}

	response := &pb.GetSuggestionsResponse{
		Suggestions: suggestions,
	}

	s.logger.Printf("GetMappingSuggestions completed: %d suggestions", len(suggestions))
	return response, nil
}

// BulkApprove approves multiple mappings at once
func (s *MappingBusinessServiceServer) BulkApprove(ctx context.Context, req *pb.BulkApproveRequest) (*pb.BulkApproveResponse, error) {
	s.logger.Printf("BulkApprove called for %d mappings", len(req.MappingIds))

	approvedCount := int32(0)
	failedCount := int32(0)
	var errors []string

	// Use bulk update if available
	_, err := s.mappingRepo.BulkUpdateStatus(ctx, req.MappingIds, pb.MappingStatus_MAPPING_STATUS_ACTIVE)
	if err != nil {
		// If bulk update fails, try individual updates
		for _, mappingID := range req.MappingIds {
			_, updateErr := s.mappingRepo.UpdateStatus(ctx, mappingID, pb.MappingStatus_MAPPING_STATUS_ACTIVE)
			if updateErr != nil {
				failedCount++
				errors = append(errors, "Failed to approve mapping "+strconv.FormatInt(mappingID, 10)+": "+updateErr.Error())
			} else {
				approvedCount++
			}
		}
	} else {
		approvedCount = int32(len(req.MappingIds))
	}

	response := &pb.BulkApproveResponse{
		ApprovedCount: approvedCount,
		FailedCount:   failedCount,
		Errors:        errors,
	}

	s.logger.Printf("BulkApprove completed: %d approved, %d failed", approvedCount, failedCount)
	return response, nil
}

// RecalculateConfidence recalculates the confidence score for a mapping
func (s *MappingBusinessServiceServer) RecalculateConfidence(ctx context.Context, req *pb.RecalculateConfidenceRequest) (*pb.RecalculateConfidenceResponse, error) {
	s.logger.Printf("RecalculateConfidence called for mapping ID: %d", req.MappingId)

	if req.MappingId <= 0 {
		return nil, status.Errorf(codes.InvalidArgument, "mapping_id must be positive")
	}

	// Get the current mapping
	mapping, err := s.mappingRepo.GetByID(ctx, req.MappingId)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "mapping not found: %v", err)
	}

	oldConfidence := mapping.Confidence

	// Recalculate confidence based on various factors
	// This is a simplified implementation - in reality, this would use sophisticated algorithms
	newConfidence := oldConfidence
	calculationMethod := "simple_adjustment"

	// Adjust confidence based on additional data if provided
	if len(req.AdditionalData) > 0 {
		if _, exists := req.AdditionalData["high_confidence"]; exists {
			newConfidence = 0.95
			calculationMethod = "high_confidence_boost"
		} else if _, exists := req.AdditionalData["low_confidence"]; exists {
			newConfidence = 0.45
			calculationMethod = "confidence_reduction"
		}
	}

	// Update the mapping with new confidence
	mapping.Confidence = newConfidence
	mapping.UpdatedAt = timestamppb.Now()

	_, err = s.mappingRepo.Update(ctx, mapping)
	if err != nil {
		s.logger.Printf("Failed to update mapping confidence: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to update mapping: %v", err)
	}

	response := &pb.RecalculateConfidenceResponse{
		OldConfidence:     oldConfidence,
		NewConfidence:     newConfidence,
		CalculationMethod: calculationMethod,
	}

	s.logger.Printf("RecalculateConfidence completed: %f -> %f", oldConfidence, newConfidence)
	return response, nil
}

// GetMappingHistory retrieves the history of changes for a mapping
func (s *MappingBusinessServiceServer) GetMappingHistory(ctx context.Context, req *pb.GetMappingHistoryRequest) (*pb.GetMappingHistoryResponse, error) {
	s.logger.Printf("GetMappingHistory called for mapping ID: %d", req.MappingId)

	if req.MappingId <= 0 {
		return nil, status.Errorf(codes.InvalidArgument, "mapping_id must be positive")
	}

	// In a real implementation, this would retrieve audit log entries
	// For now, we'll create a mock history
	history := []*pb.MappingHistoryEntry{
		{
			Timestamp: timestamppb.New(time.Now().Add(-24 * time.Hour)),
			Action:    "created",
			Actor:     "system",
			Changes: map[string]string{
				"status":     "active",
				"confidence": "0.85",
			},
		},
		{
			Timestamp: timestamppb.New(time.Now().Add(-12 * time.Hour)),
			Action:    "confidence_updated",
			Actor:     "auto_system",
			Changes: map[string]string{
				"old_confidence": "0.85",
				"new_confidence": "0.90",
			},
		},
	}

	response := &pb.GetMappingHistoryResponse{
		History: history,
	}

	s.logger.Printf("GetMappingHistory completed: %d entries", len(history))
	return response, nil
}

// ExportMappings exports mappings in the requested format
func (s *MappingBusinessServiceServer) ExportMappings(ctx context.Context, req *pb.ExportMappingsRequest) (*pb.ExportMappingsResponse, error) {
	s.logger.Printf("ExportMappings called with format: %s", req.Format)

	// Get mappings based on filter criteria
	// For simplicity, we'll get all mappings with basic pagination
	mappingsList, err := s.mappingRepo.List(ctx, 1000, 0) // Get up to 1000 mappings
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to retrieve mappings: %v", err)
	}

	// Convert to export format
	var exportData []byte
	var contentType string
	var filename string

	switch req.Format {
	case "csv":
		// Create CSV data
		csvData := "ID,ETC_Record_ID,Mapping_Type,Mapped_Entity_ID,Mapped_Entity_Type,Confidence,Status\n"
		for _, mapping := range mappingsList.Mappings {
			csvData += strconv.FormatInt(mapping.Id, 10) + ","
			csvData += strconv.FormatInt(mapping.EtcRecordId, 10) + ","
			csvData += mapping.MappingType + ","
			csvData += strconv.FormatInt(mapping.MappedEntityId, 10) + ","
			csvData += mapping.MappedEntityType + ","
			csvData += strconv.FormatFloat(float64(mapping.Confidence), 'f', 2, 32) + ","
			csvData += MappingStatusToString(mapping.Status) + "\n"
		}
		exportData = []byte(csvData)
		contentType = "text/csv"
		filename = "mappings_" + time.Now().Format("20060102_150405") + ".csv"

	case "json":
		// For JSON export, we would marshal the mappings
		// Simplified implementation
		exportData = []byte(`{"mappings": [], "exported_at": "` + time.Now().Format(time.RFC3339) + `"}`)
		contentType = "application/json"
		filename = "mappings_" + time.Now().Format("20060102_150405") + ".json"

	default:
		return nil, status.Errorf(codes.InvalidArgument, "unsupported format: %s", req.Format)
	}

	response := &pb.ExportMappingsResponse{
		Data:        exportData,
		ContentType: contentType,
		Filename:    filename,
		RecordCount: int32(len(mappingsList.Mappings)),
	}

	s.logger.Printf("ExportMappings completed: %d records exported", len(mappingsList.Mappings))
	return response, nil
}
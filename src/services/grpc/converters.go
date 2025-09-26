package grpc

// Hook test comment - project-wide error detection (src/ and tests/)
import (
	"encoding/json"
	"strings"
	"time"

	pb "github.com/yhonda-ohishi/etc_meisai/src/pb"
	"github.com/yhonda-ohishi/etc_meisai/src/models"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// ETCMappingToProto converts internal ETCMapping model to protobuf message
func ETCMappingToProto(mapping *models.ETCMapping) *pb.ETCMappingEntity {
	if mapping == nil {
		return nil
	}

	pbMapping := &pb.ETCMappingEntity{
		Id:               mapping.ID,
		EtcRecordId:      mapping.ETCRecordID,
		MappingType:      mapping.MappingType,
		MappedEntityId:   mapping.MappedEntityID,
		MappedEntityType: mapping.MappedEntityType,
		Confidence:       mapping.Confidence,
		Status:           mapping.Status,
		CreatedBy:        mapping.CreatedBy,
		CreatedAt:        timestamppb.New(mapping.CreatedAt),
		UpdatedAt:        timestamppb.New(mapping.UpdatedAt),
	}

	// Convert metadata if present
	if len(mapping.Metadata) > 0 {
		var metadata map[string]interface{}
		if err := json.Unmarshal(mapping.Metadata, &metadata); err == nil {
			if structMetadata, err := structpb.NewStruct(metadata); err == nil {
				pbMapping.Metadata = structMetadata
			}
		}
	}

	return pbMapping
}

// ProtoToETCMapping converts protobuf message to internal ETCMapping model
func ProtoToETCMapping(pbMapping *pb.ETCMappingEntity) *models.ETCMapping {
	if pbMapping == nil {
		return nil
	}

	mapping := &models.ETCMapping{
		ID:               pbMapping.Id,
		ETCRecordID:      pbMapping.EtcRecordId,
		MappingType:      pbMapping.MappingType,
		MappedEntityID:   pbMapping.MappedEntityId,
		MappedEntityType: pbMapping.MappedEntityType,
		Confidence:       pbMapping.Confidence,
		Status:           pbMapping.Status,
		CreatedBy:        pbMapping.CreatedBy,
	}

	// Convert timestamps
	if pbMapping.CreatedAt != nil {
		mapping.CreatedAt = pbMapping.CreatedAt.AsTime()
	}
	if pbMapping.UpdatedAt != nil {
		mapping.UpdatedAt = pbMapping.UpdatedAt.AsTime()
	}

	// Convert metadata if present
	if pbMapping.Metadata != nil {
		if metadata := pbMapping.Metadata.AsMap(); metadata != nil {
			if err := mapping.SetMetadata(metadata); err == nil {
				// Metadata set successfully
			}
		}
	}

	return mapping
}

// ETCMeisaiRecordToProto converts internal ETCMeisaiRecord model to protobuf message
func ETCMeisaiRecordToProto(record *models.ETCMeisaiRecord) *pb.ETCMeisaiRecord {
	if record == nil {
		return nil
	}

	pbRecord := &pb.ETCMeisaiRecord{
		Id:            record.ID,
		Date:          record.Date.Format("2006-01-02"),
		Time:          record.Time,
		EntranceIc:    record.EntranceIC,
		ExitIc:        record.ExitIC,
		TollAmount:    int32(record.TollAmount),
		CarNumber:     record.CarNumber,
		EtcCardNumber: record.ETCCardNumber,
		Hash:          record.Hash,
		CreatedAt:     timestamppb.New(record.CreatedAt),
		UpdatedAt:     timestamppb.New(record.UpdatedAt),
	}

	// Optional fields
	if record.ETCNum != nil {
		pbRecord.EtcNum = record.ETCNum
	}
	if record.DtakoRowID != nil {
		pbRecord.DtakoRowId = record.DtakoRowID
	}

	return pbRecord
}

// ProtoToETCMeisaiRecord converts protobuf message to internal ETCMeisaiRecord model
func ProtoToETCMeisaiRecord(pbRecord *pb.ETCMeisaiRecord) *models.ETCMeisaiRecord {
	if pbRecord == nil {
		return nil
	}

	record := &models.ETCMeisaiRecord{
		ID:            pbRecord.Id,
		Time:          pbRecord.Time,
		EntranceIC:    pbRecord.EntranceIc,
		ExitIC:        pbRecord.ExitIc,
		TollAmount:    int(pbRecord.TollAmount),
		CarNumber:     pbRecord.CarNumber,
		ETCCardNumber: pbRecord.EtcCardNumber,
		Hash:          pbRecord.Hash,
		ETCNum:        pbRecord.EtcNum,
		DtakoRowID:    pbRecord.DtakoRowId,
	}

	// Convert date string to time
	if pbRecord.Date != "" {
		if parsedDate, err := time.Parse("2006-01-02", pbRecord.Date); err == nil {
			record.Date = parsedDate
		}
	}
	if pbRecord.CreatedAt != nil {
		record.CreatedAt = pbRecord.CreatedAt.AsTime()
	}
	if pbRecord.UpdatedAt != nil {
		record.UpdatedAt = pbRecord.UpdatedAt.AsTime()
	}

	return record
}

// CreateMappingParamsToProto converts service params to protobuf request
func CreateMappingParamsToProto(params *models.ETCMapping) *pb.CreateMappingServiceRequest {
	if params == nil {
		return nil
	}

	req := &pb.CreateMappingServiceRequest{
		EtcRecordId:      params.ETCRecordID,
		MappingType:      params.MappingType,
		MappedEntityId:   params.MappedEntityID,
		MappedEntityType: params.MappedEntityType,
		Confidence:       params.Confidence,
	}

	// Convert metadata if present
	if len(params.Metadata) > 0 {
		var metadata map[string]interface{}
		if err := json.Unmarshal(params.Metadata, &metadata); err == nil {
			// Convert to string map for proto
			stringMetadata := make(map[string]string)
			for key, value := range metadata {
				if strValue, ok := value.(string); ok {
					stringMetadata[key] = strValue
				}
			}
			req.Metadata = stringMetadata
		}
	}

	return req
}

// CreateRecordParamsToProto converts service params to protobuf record
func CreateRecordParamsToProto(params *models.ETCMeisaiRecord) *pb.ETCMeisaiRecord {
	if params == nil {
		return nil
	}

	return ETCMeisaiRecordToProto(params)
}

// ValidationResultToProto converts validation result to protobuf
func ValidationResultToProto(isValid bool, errors []string) *pb.ValidationResult {
	var validationErrors []*pb.ValidationError
	for i, errMsg := range errors {
		validationErrors = append(validationErrors, &pb.ValidationError{
			Field:   "unknown",
			Message: errMsg,
			Code:    "VALIDATION_ERROR",
		})
		// Limit to prevent too many errors
		if i >= 10 {
			break
		}
	}

	return &pb.ValidationResult{
		IsValid: isValid,
		Errors:  validationErrors,
	}
}

// MappingStatusFromString converts string status to protobuf enum
func MappingStatusFromString(status string) pb.MappingStatus {
	switch status {
	case "active":
		return pb.MappingStatus_MAPPING_STATUS_ACTIVE
	case "inactive":
		return pb.MappingStatus_MAPPING_STATUS_INACTIVE
	case "pending":
		return pb.MappingStatus_MAPPING_STATUS_PENDING
	case "approved":
		return pb.MappingStatus_MAPPING_STATUS_ACTIVE
	case "rejected":
		return pb.MappingStatus_MAPPING_STATUS_REJECTED
	default:
		return pb.MappingStatus_MAPPING_STATUS_UNSPECIFIED
	}
}

// MappingStatusToString converts protobuf enum to string
func MappingStatusToString(status pb.MappingStatus) string {
	switch status {
	case pb.MappingStatus_MAPPING_STATUS_ACTIVE:
		return "active"
	case pb.MappingStatus_MAPPING_STATUS_INACTIVE:
		return "inactive"
	case pb.MappingStatus_MAPPING_STATUS_PENDING:
		return "pending"
	case pb.MappingStatus_MAPPING_STATUS_REJECTED:
		return "rejected"
	default:
		return "unspecified"
	}
}

// BatchOperationResultToProto converts batch operation result
func BatchOperationResultToProto(totalRecords, successCount, failedCount int32, errors []string, processingTime time.Duration) *pb.BatchOperationResult {
	// Convert string errors to BatchOperationError objects
	var batchErrors []*pb.BatchOperationError
	for _, errMsg := range errors {
		batchErrors = append(batchErrors, &pb.BatchOperationError{
			ErrorCode:    "BATCH_ERROR",
			ErrorMessage: errMsg,
		})
	}

	return &pb.BatchOperationResult{
		TotalCount:   totalRecords,
		SuccessCount: successCount,
		FailureCount: failedCount,
		Errors:       batchErrors,
	}
}

// ImportSessionToProto converts import session information
func ImportSessionToProto(sessionID, accountID, accountType, status string, startTime, endTime *time.Time, totalRecords, processedRecords int32) *pb.ImportSession {
	// Convert string status to ImportStatus enum
	var importStatus pb.ImportStatus
	switch status {
	case "pending":
		importStatus = pb.ImportStatus_IMPORT_STATUS_PENDING
	case "processing":
		importStatus = pb.ImportStatus_IMPORT_STATUS_PROCESSING
	case "completed":
		importStatus = pb.ImportStatus_IMPORT_STATUS_COMPLETED
	case "failed":
		importStatus = pb.ImportStatus_IMPORT_STATUS_FAILED
	case "cancelled":
		importStatus = pb.ImportStatus_IMPORT_STATUS_CANCELLED
	default:
		importStatus = pb.ImportStatus_IMPORT_STATUS_UNSPECIFIED
	}

	session := &pb.ImportSession{
		Id:            sessionID,
		AccountId:     accountID,
		AccountType:   accountType,
		Status:        importStatus,
		TotalRows:     totalRecords,
		ProcessedRows: processedRecords,
	}

	if startTime != nil {
		session.StartedAt = timestamppb.New(*startTime)
	}
	if endTime != nil {
		session.CompletedAt = timestamppb.New(*endTime)
	}

	return session
}

// DateRangeFromProto converts protobuf DateRange to Go time values
func DateRangeFromProto(dateRange *pb.DateRange) (time.Time, time.Time) {
	var from, to time.Time

	if dateRange != nil {
		if dateRange.Start != nil {
			from = dateRange.Start.AsTime()
		}
		if dateRange.End != nil {
			to = dateRange.End.AsTime()
		}
	}

	return from, to
}

// DateRangeToProto converts Go time values to protobuf DateRange
func DateRangeToProto(from, to time.Time) *pb.DateRange {
	return &pb.DateRange{
		Start: timestamppb.New(from),
		End:   timestamppb.New(to),
	}
}

// SortCriteriaFromProto converts protobuf sort criteria to internal representation
func SortCriteriaFromProto(sort *pb.SortCriteria) (string, string) {
	if sort == nil {
		return "created_at", "desc" // Default values
	}

	var direction string
	switch sort.Direction {
	case pb.SortDirection_SORT_DIRECTION_ASC:
		direction = "asc"
	case pb.SortDirection_SORT_DIRECTION_DESC:
		direction = "desc"
	default:
		direction = "desc"
	}

	field := sort.Field
	if field == "" {
		field = "created_at"
	}

	return field, direction
}

// SortCriteriaToProto converts internal sort criteria to protobuf
func SortCriteriaToProto(field, direction string) *pb.SortCriteria {
	var dir pb.SortDirection
	switch direction {
	case "asc":
		dir = pb.SortDirection_SORT_DIRECTION_ASC
	case "desc":
		dir = pb.SortDirection_SORT_DIRECTION_DESC
	default:
		dir = pb.SortDirection_SORT_DIRECTION_DESC
	}

	return &pb.SortCriteria{
		Field:     field,
		Direction: dir,
	}
}

// FilterCriteriaFromProto extracts filter parameters from protobuf FilterCriteria
func FilterCriteriaFromProto(filter *pb.FilterCriteria) map[string]interface{} {
	if filter == nil {
		return make(map[string]interface{})
	}

	filters := make(map[string]interface{})

	// String filters
	for key, value := range filter.StringFilters {
		filters[key] = value
	}

	// Numeric filters
	for key, value := range filter.NumericFilters {
		filters[key] = value
	}

	// Date filters
	for key, dateRange := range filter.DateFilters {
		if dateRange != nil {
			start, end := DateRangeFromProto(dateRange)
			if key == "date_range" {
				filters["date_from"] = start
				filters["date_to"] = end
			} else {
				filters[key] = map[string]time.Time{
					"start": start,
					"end":   end,
				}
			}
		}
	}

	// In filters (array-based filters)
	if len(filter.InFilters) > 0 {
		filters["in_filters"] = filter.InFilters
	}

	return filters
}

// ErrorToGRPCStatus converts Go errors to appropriate gRPC status codes
// This is used internally by the server implementations but can be useful for consistency
func ErrorToGRPCStatus(err error) error {
	if err == nil {
		return nil
	}

	// You could add more sophisticated error type checking here
	// For now, we'll use a simple approach
	errStr := err.Error()

	if strings.Contains(errStr, "not found") {
		return status.Errorf(codes.NotFound, "%s", err.Error())
	}
	if strings.Contains(errStr, "duplicate") {
		return status.Errorf(codes.AlreadyExists, "%s", err.Error())
	}
	if strings.Contains(errStr, "invalid") || strings.Contains(errStr, "validation") {
		return status.Errorf(codes.InvalidArgument, "%s", err.Error())
	}
	if strings.Contains(errStr, "unauthorized") || strings.Contains(errStr, "permission") {
		return status.Errorf(codes.PermissionDenied, "%s", err.Error())
	}

	// Default to internal error
	return status.Errorf(codes.Internal, "%s", err.Error())
}

// Additional helper functions for common conversions

// StringPtr returns a pointer to the given string value
func StringPtr(s string) *string {
	return &s
}

// Int64Ptr returns a pointer to the given int64 value
func Int64Ptr(i int64) *int64 {
	return &i
}

// Float32Ptr returns a pointer to the given float32 value
func Float32Ptr(f float32) *float32 {
	return &f
}

// TimePtr returns a pointer to the given time.Time value
func TimePtr(t time.Time) *time.Time {
	return &t
}

// ETCMappingEntityToETCMapping converts ETCMappingEntity to ETCMapping
func ETCMappingEntityToETCMapping(entity *pb.ETCMappingEntity) *pb.ETCMapping {
	if entity == nil {
		return nil
	}

	// Convert status string to MappingStatus enum
	var status pb.MappingStatus
	switch entity.Status {
	case "active":
		status = pb.MappingStatus_MAPPING_STATUS_ACTIVE
	case "inactive":
		status = pb.MappingStatus_MAPPING_STATUS_INACTIVE
	case "pending":
		status = pb.MappingStatus_MAPPING_STATUS_PENDING
	case "rejected":
		status = pb.MappingStatus_MAPPING_STATUS_REJECTED
	default:
		status = pb.MappingStatus_MAPPING_STATUS_UNSPECIFIED
	}

	return &pb.ETCMapping{
		Id:               entity.Id,
		EtcRecordId:      entity.EtcRecordId,
		MappingType:      entity.MappingType,
		MappedEntityId:   entity.MappedEntityId,
		MappedEntityType: entity.MappedEntityType,
		Confidence:       entity.Confidence,
		Status:           status,
		CreatedBy:        entity.CreatedBy,
		CreatedAt:        entity.CreatedAt,
		UpdatedAt:        entity.UpdatedAt,
		Metadata:         entity.Metadata,
	}
}

// ETCMappingToETCMappingEntity converts ETCMapping to ETCMappingEntity
func ETCMappingToETCMappingEntity(mapping *pb.ETCMapping) *pb.ETCMappingEntity {
	if mapping == nil {
		return nil
	}

	// Convert MappingStatus enum to string
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
		status = "unspecified"
	}

	return &pb.ETCMappingEntity{
		Id:               mapping.Id,
		EtcRecordId:      mapping.EtcRecordId,
		MappingType:      mapping.MappingType,
		MappedEntityId:   mapping.MappedEntityId,
		MappedEntityType: mapping.MappedEntityType,
		Confidence:       mapping.Confidence,
		Status:           status,
		CreatedBy:        mapping.CreatedBy,
		CreatedAt:        mapping.CreatedAt,
		UpdatedAt:        mapping.UpdatedAt,
		Metadata:         mapping.Metadata,
	}
}

// StringValue safely gets string value from pointer (returns empty string if nil)
func StringValue(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// Int64Value safely gets int64 value from pointer (returns 0 if nil)
func Int64Value(i *int64) int64 {
	if i == nil {
		return 0
	}
	return *i
}

// Float32Value safely gets float32 value from pointer (returns 0 if nil)
func Float32Value(f *float32) float32 {
	if f == nil {
		return 0
	}
	return *f
}
package adapters

import (
	"fmt"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/yhonda-ohishi/etc_meisai/src/models"
	pb "github.com/yhonda-ohishi/etc_meisai/src/pb"
)

// ProtoDBAdapter provides base functionality for converting between Protocol Buffer messages
// and database models (GORM structs). This adapter centralizes the conversion logic
// and ensures consistency across the application.
type ProtoDBAdapter struct {
	fieldConverter *FieldConverter
}

// NewProtoDBAdapter creates a new Protocol Buffer to Database adapter
func NewProtoDBAdapter() *ProtoDBAdapter {
	return &ProtoDBAdapter{
		fieldConverter: NewFieldConverter(),
	}
}

// ========================= ETC Meisai Record Conversions =========================

// ETCMeisaiRecordToDB converts a Protocol Buffer ETCMeisaiRecord to a GORM model
func (p *ProtoDBAdapter) ETCMeisaiRecordToDB(pbRecord *pb.ETCMeisaiRecord) (*models.ETCMeisai, error) {
	if pbRecord == nil {
		return nil, fmt.Errorf("protocol buffer record cannot be nil")
	}

	// Parse the date string to time.Time
	useDate, err := time.Parse("2006-01-02", pbRecord.Date)
	if err != nil {
		return nil, fmt.Errorf("failed to parse date '%s': %w", pbRecord.Date, err)
	}

	// Create the GORM model
	dbModel := &models.ETCMeisai{
		ID:        pbRecord.Id,
		UseDate:   useDate,
		UseTime:   pbRecord.Time,
		EntryIC:   pbRecord.EntranceIc,
		ExitIC:    pbRecord.ExitIc,
		Amount:    pbRecord.TollAmount,
		CarNumber: pbRecord.CarNumber,
		ETCNumber: pbRecord.EtcCardNumber,
		Hash:      pbRecord.Hash,
	}

	// Handle optional fields
	if pbRecord.EtcNum != nil {
		// Store in a custom field or handle as needed
		// For now, we'll skip this field as it's not in the GORM model
	}

	// Convert timestamps
	if pbRecord.CreatedAt != nil {
		dbModel.CreatedAt = pbRecord.CreatedAt.AsTime()
	}
	if pbRecord.UpdatedAt != nil {
		dbModel.UpdatedAt = pbRecord.UpdatedAt.AsTime()
	}

	return dbModel, nil
}

// DBToETCMeisaiRecord converts a GORM model to a Protocol Buffer ETCMeisaiRecord
func (p *ProtoDBAdapter) DBToETCMeisaiRecord(dbModel *models.ETCMeisai) (*pb.ETCMeisaiRecord, error) {
	if dbModel == nil {
		return nil, fmt.Errorf("database model cannot be nil")
	}

	// Create the Protocol Buffer message
	pbRecord := &pb.ETCMeisaiRecord{
		Id:             dbModel.ID,
		Hash:           dbModel.Hash,
		Date:           dbModel.UseDate.Format("2006-01-02"),
		Time:           dbModel.UseTime,
		EntranceIc:     dbModel.EntryIC,
		ExitIc:         dbModel.ExitIC,
		TollAmount:     dbModel.Amount,
		CarNumber:      dbModel.CarNumber,
		EtcCardNumber:  dbModel.ETCNumber,
	}

	// Convert timestamps
	if !dbModel.CreatedAt.IsZero() {
		pbRecord.CreatedAt = timestamppb.New(dbModel.CreatedAt)
	}
	if !dbModel.UpdatedAt.IsZero() {
		pbRecord.UpdatedAt = timestamppb.New(dbModel.UpdatedAt)
	}

	return pbRecord, nil
}

// ========================= ETC Mapping Conversions =========================

// ETCMappingToDB converts a Protocol Buffer ETCMapping to a GORM model
func (p *ProtoDBAdapter) ETCMappingToDB(pbMapping *pb.ETCMapping) (*models.ETCMeisaiMapping, error) {
	if pbMapping == nil {
		return nil, fmt.Errorf("protocol buffer mapping cannot be nil")
	}

	// Create the GORM model - mapping PB fields to actual DB fields
	dbModel := &models.ETCMeisaiMapping{
		ID:          pbMapping.Id,
		ETCMeisaiID: pbMapping.EtcRecordId,
		MappingType: pbMapping.MappingType,
		Confidence:  pbMapping.Confidence, // Both are float32
		CreatedBy:   pbMapping.CreatedBy,
	}

	// Map the MappedEntityId to DTakoRowID as string
	dbModel.DTakoRowID = fmt.Sprintf("%d", pbMapping.MappedEntityId)

	// Store metadata in Notes field (simplified conversion)
	if pbMapping.Metadata != nil && len(pbMapping.Metadata.Fields) > 0 {
		// Convert metadata to a simple JSON string for Notes field
		metadata, err := p.convertStructToMap(pbMapping.Metadata)
		if err == nil {
			if notes, ok := metadata["notes"].(string); ok {
				dbModel.Notes = notes
			}
		}
	}

	// Convert timestamps
	if pbMapping.CreatedAt != nil {
		dbModel.CreatedAt = pbMapping.CreatedAt.AsTime()
	}
	if pbMapping.UpdatedAt != nil {
		dbModel.UpdatedAt = pbMapping.UpdatedAt.AsTime()
	}

	return dbModel, nil
}

// DBToETCMapping converts a GORM model to a Protocol Buffer ETCMapping
func (p *ProtoDBAdapter) DBToETCMapping(dbModel *models.ETCMeisaiMapping) (*pb.ETCMapping, error) {
	if dbModel == nil {
		return nil, fmt.Errorf("database model cannot be nil")
	}

	// Convert DTakoRowID string to int64 for MappedEntityId
	var mappedEntityId int64
	if dbModel.DTakoRowID != "" {
		// Try to convert string to int64
		if id, err := fmt.Sscanf(dbModel.DTakoRowID, "%d", &mappedEntityId); err != nil || id != 1 {
			// If conversion fails, set to 0 or handle differently
			mappedEntityId = 0
		}
	}

	// Create the Protocol Buffer message
	pbMapping := &pb.ETCMapping{
		Id:                dbModel.ID,
		EtcRecordId:       dbModel.ETCMeisaiID,
		MappingType:       dbModel.MappingType,
		MappedEntityId:    mappedEntityId,
		MappedEntityType:  "dtako_record", // Default entity type for this model
		Confidence:        dbModel.Confidence, // Both are float32
		Status:            pb.MappingStatus_MAPPING_STATUS_ACTIVE, // Default status
		CreatedBy:         dbModel.CreatedBy,
	}

	// Handle metadata conversion from Notes field
	pbMapping.Metadata = &structpb.Struct{
		Fields: make(map[string]*structpb.Value),
	}
	if dbModel.Notes != "" {
		pbMapping.Metadata.Fields["notes"] = &structpb.Value{
			Kind: &structpb.Value_StringValue{StringValue: dbModel.Notes},
		}
	}

	// Convert timestamps
	if !dbModel.CreatedAt.IsZero() {
		pbMapping.CreatedAt = timestamppb.New(dbModel.CreatedAt)
	}
	if !dbModel.UpdatedAt.IsZero() {
		pbMapping.UpdatedAt = timestamppb.New(dbModel.UpdatedAt)
	}

	return pbMapping, nil
}

// ========================= Import Session Conversions =========================

// ImportSessionToDB converts a Protocol Buffer ImportSession to a GORM model
func (p *ProtoDBAdapter) ImportSessionToDB(pbSession *pb.ImportSession) (*models.ETCImportBatch, error) {
	if pbSession == nil {
		return nil, fmt.Errorf("protocol buffer session cannot be nil")
	}

	// Parse the ID as int64 if it's a string UUID
	var id int64
	if pbSession.Id != "" {
		// If the ID is a numeric string, convert it; otherwise use 0 for now
		if parsedId, err := fmt.Sscanf(pbSession.Id, "%d", &id); err != nil || parsedId != 1 {
			id = 0 // Set to 0 for UUID strings that can't be converted to int64
		}
	}

	// Create the GORM model with field mappings
	dbModel := &models.ETCImportBatch{
		ID:             id,
		FileName:       pbSession.FileName,
		FileSize:       pbSession.FileSize,
		AccountID:      pbSession.AccountId,
		ImportType:     pbSession.AccountType, // Use AccountType as ImportType
		Status:         p.convertImportStatusToDB(pbSession.Status),
		TotalRows:      int64(pbSession.TotalRows),
		ProcessedRows:  int64(pbSession.ProcessedRows),
		TotalRecords:   pbSession.TotalRows,
		ProcessedCount: pbSession.ProcessedRows,
		CreatedCount:   pbSession.SuccessRows,
		DuplicateCount: pbSession.DuplicateRows,
		ErrorCount:     int64(pbSession.ErrorRows),
		SuccessCount:   int64(pbSession.SuccessRows),
		CreatedBy:      pbSession.CreatedBy,
	}

	// Convert timestamps
	if pbSession.StartedAt != nil {
		startTime := pbSession.StartedAt.AsTime()
		dbModel.StartTime = &startTime
	}
	if pbSession.CompletedAt != nil {
		completeTime := pbSession.CompletedAt.AsTime()
		dbModel.CompleteTime = &completeTime
		dbModel.CompletedAt = &completeTime
	}
	if pbSession.CreatedAt != nil {
		dbModel.CreatedAt = pbSession.CreatedAt.AsTime()
	}

	// Handle error log - store first error message in ErrorMessage field
	if len(pbSession.ErrorLog) > 0 {
		dbModel.ErrorMessage = pbSession.ErrorLog[0].ErrorMessage
	}

	return dbModel, nil
}

// DBToImportSession converts a GORM model to a Protocol Buffer ImportSession
func (p *ProtoDBAdapter) DBToImportSession(dbModel *models.ETCImportBatch) (*pb.ImportSession, error) {
	if dbModel == nil {
		return nil, fmt.Errorf("database model cannot be nil")
	}

	// Create the Protocol Buffer message with field mappings
	pbSession := &pb.ImportSession{
		Id:            fmt.Sprintf("%d", dbModel.ID), // Convert int64 to string
		AccountType:   dbModel.ImportType, // Use ImportType as AccountType
		AccountId:     dbModel.AccountID,
		FileName:      dbModel.FileName,
		FileSize:      dbModel.FileSize,
		Status:        p.convertImportStatusToPB(dbModel.Status),
		TotalRows:     dbModel.TotalRecords,
		ProcessedRows: dbModel.ProcessedCount,
		SuccessRows:   dbModel.CreatedCount,
		ErrorRows:     int32(dbModel.ErrorCount),
		DuplicateRows: dbModel.DuplicateCount,
		CreatedBy:     dbModel.CreatedBy,
		ErrorLog:      []*pb.ImportError{}, // Initialize empty slice
	}

	// Convert timestamps
	if dbModel.StartTime != nil {
		pbSession.StartedAt = timestamppb.New(*dbModel.StartTime)
	}
	if dbModel.CompletedAt != nil {
		pbSession.CompletedAt = timestamppb.New(*dbModel.CompletedAt)
	}
	if !dbModel.CreatedAt.IsZero() {
		pbSession.CreatedAt = timestamppb.New(dbModel.CreatedAt)
	}

	// Add error message to error log if present
	if dbModel.ErrorMessage != "" {
		pbSession.ErrorLog = []*pb.ImportError{
			{
				RowNumber:    0, // Unknown row number
				ErrorType:    "general",
				ErrorMessage: dbModel.ErrorMessage,
				RawData:      "",
			},
		}
	}

	return pbSession, nil
}

// ========================= Utility Functions =========================

// convertMappingStatusToDB converts Protocol Buffer MappingStatus to database string
func (p *ProtoDBAdapter) convertMappingStatusToDB(pbStatus pb.MappingStatus) string {
	switch pbStatus {
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

// convertMappingStatusToPB converts database string to Protocol Buffer MappingStatus
func (p *ProtoDBAdapter) convertMappingStatusToPB(dbStatus string) pb.MappingStatus {
	switch dbStatus {
	case "active":
		return pb.MappingStatus_MAPPING_STATUS_ACTIVE
	case "inactive":
		return pb.MappingStatus_MAPPING_STATUS_INACTIVE
	case "pending":
		return pb.MappingStatus_MAPPING_STATUS_PENDING
	case "rejected":
		return pb.MappingStatus_MAPPING_STATUS_REJECTED
	default:
		return pb.MappingStatus_MAPPING_STATUS_UNSPECIFIED
	}
}

// convertImportStatusToDB converts Protocol Buffer ImportStatus to database string
func (p *ProtoDBAdapter) convertImportStatusToDB(pbStatus pb.ImportStatus) string {
	switch pbStatus {
	case pb.ImportStatus_IMPORT_STATUS_PENDING:
		return "pending"
	case pb.ImportStatus_IMPORT_STATUS_PROCESSING:
		return "processing"
	case pb.ImportStatus_IMPORT_STATUS_COMPLETED:
		return "completed"
	case pb.ImportStatus_IMPORT_STATUS_FAILED:
		return "failed"
	case pb.ImportStatus_IMPORT_STATUS_CANCELLED:
		return "cancelled"
	default:
		return "unspecified"
	}
}

// convertImportStatusToPB converts database string to Protocol Buffer ImportStatus
func (p *ProtoDBAdapter) convertImportStatusToPB(dbStatus string) pb.ImportStatus {
	switch dbStatus {
	case "pending":
		return pb.ImportStatus_IMPORT_STATUS_PENDING
	case "processing":
		return pb.ImportStatus_IMPORT_STATUS_PROCESSING
	case "completed":
		return pb.ImportStatus_IMPORT_STATUS_COMPLETED
	case "failed":
		return pb.ImportStatus_IMPORT_STATUS_FAILED
	case "cancelled":
		return pb.ImportStatus_IMPORT_STATUS_CANCELLED
	default:
		return pb.ImportStatus_IMPORT_STATUS_UNSPECIFIED
	}
}

// convertStructToMap converts a protobuf Struct to a Go map
func (p *ProtoDBAdapter) convertStructToMap(pbStruct *structpb.Struct) (map[string]interface{}, error) {
	if pbStruct == nil {
		return nil, nil
	}

	result := make(map[string]interface{})
	for key, value := range pbStruct.Fields {
		goValue, err := p.convertValueToGo(value)
		if err != nil {
			return nil, fmt.Errorf("failed to convert field '%s': %w", key, err)
		}
		result[key] = goValue
	}

	return result, nil
}

// convertValueToGo converts a protobuf Value to a Go interface{}
func (p *ProtoDBAdapter) convertValueToGo(pbValue *structpb.Value) (interface{}, error) {
	if pbValue == nil {
		return nil, nil
	}

	switch pbValue.GetKind().(type) {
	case *structpb.Value_NullValue:
		return nil, nil
	case *structpb.Value_NumberValue:
		return pbValue.GetNumberValue(), nil
	case *structpb.Value_StringValue:
		return pbValue.GetStringValue(), nil
	case *structpb.Value_BoolValue:
		return pbValue.GetBoolValue(), nil
	case *structpb.Value_StructValue:
		return p.convertStructToMap(pbValue.GetStructValue())
	case *structpb.Value_ListValue:
		listValue := pbValue.GetListValue()
		result := make([]interface{}, len(listValue.Values))
		for i, v := range listValue.Values {
			goValue, err := p.convertValueToGo(v)
			if err != nil {
				return nil, fmt.Errorf("failed to convert list item %d: %w", i, err)
			}
			result[i] = goValue
		}
		return result, nil
	default:
		return nil, fmt.Errorf("unsupported protobuf value type")
	}
}

// ========================= Timestamp Utilities =========================

// ConvertTimeToTimestamp converts a time.Time to a protobuf Timestamp
func (p *ProtoDBAdapter) ConvertTimeToTimestamp(t time.Time) *timestamppb.Timestamp {
	if t.IsZero() {
		return nil
	}
	return timestamppb.New(t)
}

// ConvertTimestampToTime converts a protobuf Timestamp to time.Time
func (p *ProtoDBAdapter) ConvertTimestampToTime(ts *timestamppb.Timestamp) time.Time {
	if ts == nil {
		return time.Time{}
	}
	return ts.AsTime()
}

// ========================= Enum Utilities =========================

// GetMappingStatusOptions returns all available mapping status options
func (p *ProtoDBAdapter) GetMappingStatusOptions() []string {
	return []string{
		"unspecified",
		"active",
		"inactive",
		"pending",
		"rejected",
	}
}

// GetImportStatusOptions returns all available import status options
func (p *ProtoDBAdapter) GetImportStatusOptions() []string {
	return []string{
		"unspecified",
		"pending",
		"processing",
		"completed",
		"failed",
		"cancelled",
	}
}

// ========================= Validation Support =========================

// ValidateETCMeisaiRecord validates a Protocol Buffer ETCMeisaiRecord before database operations
func (p *ProtoDBAdapter) ValidateETCMeisaiRecord(pbRecord *pb.ETCMeisaiRecord) error {
	if pbRecord == nil {
		return fmt.Errorf("record cannot be nil")
	}

	// Validate required fields
	if pbRecord.Date == "" {
		return fmt.Errorf("date is required")
	}
	if pbRecord.Time == "" {
		return fmt.Errorf("time is required")
	}
	if pbRecord.EntranceIc == "" {
		return fmt.Errorf("entrance IC is required")
	}
	if pbRecord.ExitIc == "" {
		return fmt.Errorf("exit IC is required")
	}
	if pbRecord.CarNumber == "" {
		return fmt.Errorf("car number is required")
	}
	if pbRecord.EtcCardNumber == "" {
		return fmt.Errorf("ETC card number is required")
	}

	// Validate data formats
	if _, err := time.Parse("2006-01-02", pbRecord.Date); err != nil {
		return fmt.Errorf("invalid date format: %w", err)
	}

	return nil
}
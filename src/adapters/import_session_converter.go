package adapters

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/yhonda-ohishi/etc_meisai/src/models"
	pb "github.com/yhonda-ohishi/etc_meisai/src/pb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// ImportSessionToProto converts a GORM model to a Proto message
func ImportSessionToProto(model *models.ImportSession) (*pb.ImportSession, error) {
	if model == nil {
		return nil, fmt.Errorf("model cannot be nil")
	}

	proto := &pb.ImportSession{
		Id:            model.ID,
		AccountType:   model.AccountType,
		AccountId:     model.AccountID,
		FileName:      model.FileName,
		FileSize:      model.FileSize,
		TotalRows:     int32(model.TotalRows),
		ProcessedRows: int32(model.ProcessedRows),
		SuccessRows:   int32(model.SuccessRows),
		ErrorRows:     int32(model.ErrorRows),
		DuplicateRows: int32(model.DuplicateRows),
		CreatedBy:     model.CreatedBy,
	}

	// Convert status string to proto enum
	status, err := stringToImportStatus(model.Status)
	if err != nil {
		return nil, fmt.Errorf("invalid import status: %w", err)
	}
	proto.Status = status

	// Convert timestamps
	if !model.StartedAt.IsZero() {
		proto.StartedAt = timestamppb.New(model.StartedAt)
	}
	if model.CompletedAt != nil && !model.CompletedAt.IsZero() {
		proto.CompletedAt = timestamppb.New(*model.CompletedAt)
	}
	if !model.CreatedAt.IsZero() {
		proto.CreatedAt = timestamppb.New(model.CreatedAt)
	}

	// Convert error log JSON array to protobuf ImportError slice
	if model.ErrorLog != nil {
		var errorLogArray []models.ImportError
		if err := json.Unmarshal([]byte(model.ErrorLog), &errorLogArray); err != nil {
			return nil, fmt.Errorf("error unmarshaling error log: %w", err)
		}

		if len(errorLogArray) > 0 {
			protoErrors := make([]*pb.ImportError, 0, len(errorLogArray))
			for _, modelError := range errorLogArray {
				protoError := &pb.ImportError{
					RowNumber:    int32(modelError.RowNumber),
					ErrorType:    modelError.ErrorType,
					ErrorMessage: modelError.ErrorMessage,
					RawData:      modelError.RawData,
				}
				protoErrors = append(protoErrors, protoError)
			}
			proto.ErrorLog = protoErrors
		}
	}

	return proto, nil
}

// ProtoToImportSession converts a Proto message to a GORM model
func ProtoToImportSession(proto *pb.ImportSession) (*models.ImportSession, error) {
	if proto == nil {
		return nil, fmt.Errorf("proto cannot be nil")
	}

	model := &models.ImportSession{
		ID:            proto.Id,
		AccountType:   proto.AccountType,
		AccountID:     proto.AccountId,
		FileName:      proto.FileName,
		FileSize:      proto.FileSize,
		TotalRows:     int(proto.TotalRows),
		ProcessedRows: int(proto.ProcessedRows),
		SuccessRows:   int(proto.SuccessRows),
		ErrorRows:     int(proto.ErrorRows),
		DuplicateRows: int(proto.DuplicateRows),
		CreatedBy:     proto.CreatedBy,
	}

	// Convert status enum to string
	model.Status = importStatusToString(proto.Status)

	// Convert timestamps
	if proto.StartedAt != nil {
		if err := proto.StartedAt.CheckValid(); err != nil {
			return nil, fmt.Errorf("invalid started_at timestamp: %w", err)
		}
		model.StartedAt = proto.StartedAt.AsTime()
	}
	if proto.CompletedAt != nil {
		if err := proto.CompletedAt.CheckValid(); err != nil {
			return nil, fmt.Errorf("invalid completed_at timestamp: %w", err)
		}
		completedAt := proto.CompletedAt.AsTime()
		model.CompletedAt = &completedAt
	}
	if proto.CreatedAt != nil {
		if err := proto.CreatedAt.CheckValid(); err != nil {
			return nil, fmt.Errorf("invalid created_at timestamp: %w", err)
		}
		model.CreatedAt = proto.CreatedAt.AsTime()
	}

	// Convert error log slice to JSON
	if len(proto.ErrorLog) > 0 {
		modelErrors := make([]models.ImportError, 0, len(proto.ErrorLog))
		for _, protoError := range proto.ErrorLog {
			modelError := models.ImportError{
				RowNumber:    int(protoError.RowNumber),
				ErrorType:    protoError.ErrorType,
				ErrorMessage: protoError.ErrorMessage,
				RawData:      protoError.RawData,
			}
			modelErrors = append(modelErrors, modelError)
		}

		// Use the model's AddError method for each error to properly set ErrorLog
		for _, importError := range modelErrors {
			if err := model.AddError(
				importError.RowNumber,
				importError.ErrorType,
				importError.ErrorMessage,
				importError.RawData,
			); err != nil {
				return nil, fmt.Errorf("error adding import error: %w", err)
			}
		}
	}

	return model, nil
}

// ImportSessionsToProto converts a slice of GORM models to a slice of Proto messages
func ImportSessionsToProto(models []*models.ImportSession) ([]*pb.ImportSession, error) {
	if models == nil {
		return nil, nil
	}

	protos := make([]*pb.ImportSession, 0, len(models))
	for i, model := range models {
		if model == nil {
			return nil, fmt.Errorf("model at index %d cannot be nil", i)
		}

		proto, err := ImportSessionToProto(model)
		if err != nil {
			return nil, fmt.Errorf("error converting model at index %d: %w", i, err)
		}

		protos = append(protos, proto)
	}

	return protos, nil
}

// ProtoToImportSessions converts a slice of Proto messages to a slice of GORM models
func ProtoToImportSessions(protos []*pb.ImportSession) ([]*models.ImportSession, error) {
	if protos == nil {
		return nil, nil
	}

	models := make([]*models.ImportSession, 0, len(protos))
	for i, proto := range protos {
		if proto == nil {
			return nil, fmt.Errorf("proto at index %d cannot be nil", i)
		}

		model, err := ProtoToImportSession(proto)
		if err != nil {
			return nil, fmt.Errorf("error converting proto at index %d: %w", i, err)
		}

		models = append(models, model)
	}

	return models, nil
}

// stringToImportStatus converts string status to proto enum
func stringToImportStatus(status string) (pb.ImportStatus, error) {
	status = strings.ToLower(strings.TrimSpace(status))

	switch status {
	case "pending":
		return pb.ImportStatus_IMPORT_STATUS_PENDING, nil
	case "processing":
		return pb.ImportStatus_IMPORT_STATUS_PROCESSING, nil
	case "completed":
		return pb.ImportStatus_IMPORT_STATUS_COMPLETED, nil
	case "failed":
		return pb.ImportStatus_IMPORT_STATUS_FAILED, nil
	case "cancelled":
		return pb.ImportStatus_IMPORT_STATUS_CANCELLED, nil
	case "":
		return pb.ImportStatus_IMPORT_STATUS_UNSPECIFIED, nil
	default:
		return pb.ImportStatus_IMPORT_STATUS_UNSPECIFIED,
			fmt.Errorf("unknown import status: %s", status)
	}
}

// importStatusToString converts proto enum to string
func importStatusToString(status pb.ImportStatus) string {
	switch status {
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
	case pb.ImportStatus_IMPORT_STATUS_UNSPECIFIED:
		return ""
	default:
		return ""
	}
}
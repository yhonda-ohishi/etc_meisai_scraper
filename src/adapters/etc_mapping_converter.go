package adapters

import (
	"fmt"
	"strings"

	"github.com/yhonda-ohishi/etc_meisai/src/models"
	"github.com/yhonda-ohishi/etc_meisai/src/pb"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// ETCMappingToProto converts a GORM model to a Proto message
func ETCMappingToProto(model *models.ETCMapping) (*pb.ETCMapping, error) {
	if model == nil {
		return nil, fmt.Errorf("model cannot be nil")
	}

	proto := &pb.ETCMapping{
		Id:               model.ID,
		EtcRecordId:      model.ETCRecordID,
		MappingType:      model.MappingType,
		MappedEntityId:   model.MappedEntityID,
		MappedEntityType: model.MappedEntityType,
		Confidence:       model.Confidence,
		CreatedBy:        model.CreatedBy,
	}

	// Convert status string to proto enum
	status, err := stringToMappingStatus(model.Status)
	if err != nil {
		return nil, fmt.Errorf("invalid mapping status: %w", err)
	}
	proto.Status = status

	// Convert ETCRecord if present
	if model.ETCRecord.ID != 0 {
		etcRecord, err := ETCMeisaiRecordToProto(&model.ETCRecord)
		if err != nil {
			return nil, fmt.Errorf("error converting ETC record: %w", err)
		}
		proto.EtcRecord = etcRecord
	}

	// Convert metadata JSON to protobuf Struct
	if model.Metadata != nil {
		metadata, err := model.GetMetadata()
		if err != nil {
			return nil, fmt.Errorf("error getting metadata: %w", err)
		}
		if metadata != nil {
			metadataStruct, err := structpb.NewStruct(metadata)
			if err != nil {
				return nil, fmt.Errorf("error converting metadata to struct: %w", err)
			}
			proto.Metadata = metadataStruct
		}
	}

	// Convert timestamps
	if !model.CreatedAt.IsZero() {
		proto.CreatedAt = timestamppb.New(model.CreatedAt)
	}
	if !model.UpdatedAt.IsZero() {
		proto.UpdatedAt = timestamppb.New(model.UpdatedAt)
	}

	return proto, nil
}

// ProtoToETCMapping converts a Proto message to a GORM model
func ProtoToETCMapping(proto *pb.ETCMapping) (*models.ETCMapping, error) {
	if proto == nil {
		return nil, fmt.Errorf("proto cannot be nil")
	}

	model := &models.ETCMapping{
		ID:               proto.Id,
		ETCRecordID:      proto.EtcRecordId,
		MappingType:      proto.MappingType,
		MappedEntityID:   proto.MappedEntityId,
		MappedEntityType: proto.MappedEntityType,
		Confidence:       proto.Confidence,
		CreatedBy:        proto.CreatedBy,
	}

	// Convert status enum to string
	model.Status = mappingStatusToString(proto.Status)

	// Convert ETCRecord if present
	if proto.EtcRecord != nil {
		etcRecord, err := ProtoToETCMeisaiRecord(proto.EtcRecord)
		if err != nil {
			return nil, fmt.Errorf("error converting ETC record: %w", err)
		}
		model.ETCRecord = *etcRecord
	}

	// Convert metadata Struct to JSON
	if proto.Metadata != nil {
		metadata := proto.Metadata.AsMap()
		if err := model.SetMetadata(metadata); err != nil {
			return nil, fmt.Errorf("error setting metadata: %w", err)
		}
	}

	// Convert timestamps
	if proto.CreatedAt != nil {
		if err := proto.CreatedAt.CheckValid(); err != nil {
			return nil, fmt.Errorf("invalid created_at timestamp: %w", err)
		}
		model.CreatedAt = proto.CreatedAt.AsTime()
	}
	if proto.UpdatedAt != nil {
		if err := proto.UpdatedAt.CheckValid(); err != nil {
			return nil, fmt.Errorf("invalid updated_at timestamp: %w", err)
		}
		model.UpdatedAt = proto.UpdatedAt.AsTime()
	}

	return model, nil
}

// ETCMappingsToProto converts a slice of GORM models to a slice of Proto messages
func ETCMappingsToProto(models []*models.ETCMapping) ([]*pb.ETCMapping, error) {
	if models == nil {
		return nil, nil
	}

	protos := make([]*pb.ETCMapping, 0, len(models))
	for i, model := range models {
		if model == nil {
			return nil, fmt.Errorf("model at index %d cannot be nil", i)
		}

		proto, err := ETCMappingToProto(model)
		if err != nil {
			return nil, fmt.Errorf("error converting model at index %d: %w", i, err)
		}

		protos = append(protos, proto)
	}

	return protos, nil
}

// ProtoToETCMappings converts a slice of Proto messages to a slice of GORM models
func ProtoToETCMappings(protos []*pb.ETCMapping) ([]*models.ETCMapping, error) {
	if protos == nil {
		return nil, nil
	}

	models := make([]*models.ETCMapping, 0, len(protos))
	for i, proto := range protos {
		if proto == nil {
			return nil, fmt.Errorf("proto at index %d cannot be nil", i)
		}

		model, err := ProtoToETCMapping(proto)
		if err != nil {
			return nil, fmt.Errorf("error converting proto at index %d: %w", i, err)
		}

		models = append(models, model)
	}

	return models, nil
}

// stringToMappingStatus converts string status to proto enum
func stringToMappingStatus(status string) (pb.MappingStatus, error) {
	status = strings.ToLower(strings.TrimSpace(status))

	switch status {
	case "active":
		return pb.MappingStatus_MAPPING_STATUS_ACTIVE, nil
	case "inactive":
		return pb.MappingStatus_MAPPING_STATUS_INACTIVE, nil
	case "pending":
		return pb.MappingStatus_MAPPING_STATUS_PENDING, nil
	case "rejected":
		return pb.MappingStatus_MAPPING_STATUS_REJECTED, nil
	case "":
		return pb.MappingStatus_MAPPING_STATUS_UNSPECIFIED, nil
	default:
		return pb.MappingStatus_MAPPING_STATUS_UNSPECIFIED,
			fmt.Errorf("unknown mapping status: %s", status)
	}
}

// mappingStatusToString converts proto enum to string
func mappingStatusToString(status pb.MappingStatus) string {
	switch status {
	case pb.MappingStatus_MAPPING_STATUS_ACTIVE:
		return "active"
	case pb.MappingStatus_MAPPING_STATUS_INACTIVE:
		return "inactive"
	case pb.MappingStatus_MAPPING_STATUS_PENDING:
		return "pending"
	case pb.MappingStatus_MAPPING_STATUS_REJECTED:
		return "rejected"
	case pb.MappingStatus_MAPPING_STATUS_UNSPECIFIED:
		return ""
	default:
		return ""
	}
}
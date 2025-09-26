package adapters

import (
	"fmt"
	"time"

	"github.com/yhonda-ohishi/etc_meisai/src/models"
	pb "github.com/yhonda-ohishi/etc_meisai/src/pb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// ETCMeisaiRecordToProto converts a GORM model to a Proto message
func ETCMeisaiRecordToProto(model *models.ETCMeisaiRecord) (*pb.ETCMeisaiRecord, error) {
	if model == nil {
		return nil, fmt.Errorf("model cannot be nil")
	}

	proto := &pb.ETCMeisaiRecord{
		Id:            model.ID,
		Hash:          model.Hash,
		Date:          model.Date.Format("2006-01-02"), // YYYY-MM-DD format
		Time:          model.Time,
		EntranceIc:    model.EntranceIC,
		ExitIc:        model.ExitIC,
		TollAmount:    int32(model.TollAmount), // Convert int to int32
		CarNumber:     model.CarNumber,
		EtcCardNumber: model.ETCCardNumber,
	}

	// Handle optional ETCNum field
	if model.ETCNum != nil {
		proto.EtcNum = model.ETCNum
	}

	// Handle optional DtakoRowID field
	if model.DtakoRowID != nil {
		proto.DtakoRowId = model.DtakoRowID
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

// ProtoToETCMeisaiRecord converts a Proto message to a GORM model
func ProtoToETCMeisaiRecord(proto *pb.ETCMeisaiRecord) (*models.ETCMeisaiRecord, error) {
	if proto == nil {
		return nil, fmt.Errorf("proto cannot be nil")
	}

	// Parse date string to time.Time
	date, err := time.Parse("2006-01-02", proto.Date)
	if err != nil {
		return nil, fmt.Errorf("invalid date format: %w", err)
	}

	model := &models.ETCMeisaiRecord{
		ID:              proto.Id,
		Hash:            proto.Hash,
		Date:            date,
		Time:            proto.Time,
		EntranceIC:      proto.EntranceIc,
		ExitIC:          proto.ExitIc,
		TollAmount:      int(proto.TollAmount), // Convert int32 to int
		CarNumber:       proto.CarNumber,
		ETCCardNumber:   proto.EtcCardNumber,
	}

	// Handle optional ETCNum field
	if proto.EtcNum != nil {
		model.ETCNum = proto.EtcNum
	}

	// Handle optional DtakoRowID field
	if proto.DtakoRowId != nil {
		model.DtakoRowID = proto.DtakoRowId
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

// ETCMeisaiRecordsToProto converts a slice of GORM models to a slice of Proto messages
func ETCMeisaiRecordsToProto(models []*models.ETCMeisaiRecord) ([]*pb.ETCMeisaiRecord, error) {
	if models == nil {
		return nil, nil
	}

	protos := make([]*pb.ETCMeisaiRecord, 0, len(models))
	for i, model := range models {
		if model == nil {
			return nil, fmt.Errorf("model at index %d cannot be nil", i)
		}

		proto, err := ETCMeisaiRecordToProto(model)
		if err != nil {
			return nil, fmt.Errorf("error converting model at index %d: %w", i, err)
		}

		protos = append(protos, proto)
	}

	return protos, nil
}

// ProtoToETCMeisaiRecords converts a slice of Proto messages to a slice of GORM models
func ProtoToETCMeisaiRecords(protos []*pb.ETCMeisaiRecord) ([]*models.ETCMeisaiRecord, error) {
	if protos == nil {
		return nil, nil
	}

	models := make([]*models.ETCMeisaiRecord, 0, len(protos))
	for i, proto := range protos {
		if proto == nil {
			return nil, fmt.Errorf("proto at index %d cannot be nil", i)
		}

		model, err := ProtoToETCMeisaiRecord(proto)
		if err != nil {
			return nil, fmt.Errorf("error converting proto at index %d: %w", i, err)
		}

		models = append(models, model)
	}

	return models, nil
}
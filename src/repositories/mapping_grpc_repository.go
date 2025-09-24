package repositories

import (
	// "context" // Commented out - not used when clients package is deleted
	"fmt"

	// "github.com/yhonda-ohishi/etc_meisai/src/clients" // Commented out - clients package deleted
	"github.com/yhonda-ohishi/etc_meisai/src/models"
	// "github.com/yhonda-ohishi/etc_meisai/src/pb" // Commented out - not used when clients package is deleted
)

// MappingGRPCRepository implements MappingRepository interface using gRPC client
type MappingGRPCRepository struct {
	client interface{} // TODO: Replace with proper type when clients package is restored
}

// NewMappingGRPCRepository creates a new gRPC-based mapping repository
func NewMappingGRPCRepository(client interface{}) MappingRepository {
	return &MappingGRPCRepository{
		client: client,
	}
}

// Create creates a new mapping record via gRPC
func (r *MappingGRPCRepository) Create(mapping *models.ETCMeisaiMapping) error {
	// TODO: Restore when clients package is available
	// ctx := context.Background()
	//
	// req := &pb.CreateMappingRequest{
	//	EtcMeisaiId:  mapping.ETCMeisaiID,
	//	DtakoRowId:   mapping.DTakoRowID,
	//	MappingType:  mapping.MappingType,
	//	Confidence:   mapping.Confidence,
	//	Notes:        mapping.Notes,
	// }
	return fmt.Errorf("CreateMapping not available - clients package deleted")
}

// GetByID retrieves a mapping by ID via gRPC
func (r *MappingGRPCRepository) GetByID(id int64) (*models.ETCMeisaiMapping, error) {
	// TODO: Restore when clients package is available
	// ctx := context.Background()
	//
	// req := &pb.GetMappingRequest{
	//	Id: id,
	// }
	return nil, fmt.Errorf("GetMapping not available - clients package deleted")
}

// Update updates an existing mapping record via gRPC
func (r *MappingGRPCRepository) Update(mapping *models.ETCMeisaiMapping) error {
	// TODO: Restore when clients package is available
	// ctx := context.Background()
	//
	// req := &pb.UpdateMappingRequest{
	//	Id:          mapping.ID,
	//	MappingType: mapping.MappingType,
	//	Confidence:  mapping.Confidence,
	//	Notes:       mapping.Notes,
	// }
	return fmt.Errorf("UpdateMapping not available - clients package deleted")
}

// Delete deletes a mapping record via gRPC
func (r *MappingGRPCRepository) Delete(id int64) error {
	// TODO: Restore when clients package is available
	// ctx := context.Background()
	//
	// req := &pb.DeleteMappingRequest{
	//	Id: id,
	// }
	return fmt.Errorf("DeleteMapping not available - clients package deleted")
}

// GetByETCMeisaiID retrieves mappings by ETC Meisai ID via gRPC
func (r *MappingGRPCRepository) GetByETCMeisaiID(etcMeisaiID int64) ([]*models.ETCMeisaiMapping, error) {
	// TODO: Restore when clients package is available
	// ctx := context.Background()
	//
	// req := &pb.ListMappingsRequest{
	//	EtcMeisaiId: etcMeisaiID,
	//	Limit:       100,
	// }
	return nil, fmt.Errorf("ListMappings not available - clients package deleted")
}

// GetByDTakoRowID retrieves a mapping by DTako row ID via gRPC
func (r *MappingGRPCRepository) GetByDTakoRowID(dtakoRowID string) (*models.ETCMeisaiMapping, error) {
	// TODO: Restore when clients package is available
	// ctx := context.Background()
	//
	// req := &pb.ListMappingsRequest{
	//	DtakoRowId: dtakoRowID,
	//	Limit:      1,
	// }
	return nil, fmt.Errorf("GetByDTakoRowID not available - clients package deleted")
}

// List retrieves mappings with pagination via gRPC
func (r *MappingGRPCRepository) List(params *models.MappingListParams) ([]*models.ETCMeisaiMapping, int64, error) {
	// TODO: Restore when clients package is available
	// ctx := context.Background()
	//
	// req := &pb.ListMappingsRequest{
	//	Limit:  int32(params.Limit),
	//	Offset: int32(params.Offset),
	// }
	//
	// if params.ETCMeisaiID != nil {
	//	req.EtcMeisaiId = *params.ETCMeisaiID
	// }
	// if params.DTakoRowID != "" {
	//	req.DtakoRowId = params.DTakoRowID
	// }
	// if params.MappingType != "" {
	//	req.MappingType = params.MappingType
	// }
	// if params.MinConfidence != nil {
	//	req.MinConfidence = *params.MinConfidence
	// }
	return nil, 0, fmt.Errorf("List mappings not available - clients package deleted")
}

// BulkCreateMappings creates multiple mappings via gRPC
func (r *MappingGRPCRepository) BulkCreateMappings(mappings []*models.ETCMeisaiMapping) error {
	// TODO: Restore when clients package is available
	// ctx := context.Background()
	//
	// var pbMappings []*pb.CreateMappingRequest
	// for _, mapping := range mappings {
	//	pbMapping := &pb.CreateMappingRequest{
	//		EtcMeisaiId:  mapping.ETCMeisaiID,
	//		DtakoRowId:   mapping.DTakoRowID,
	//		MappingType:  mapping.MappingType,
	//		Confidence:   mapping.Confidence,
	//		Notes:        mapping.Notes,
	//	}
	//	pbMappings = append(pbMappings, pbMapping)
	// }
	//
	// req := &pb.BulkCreateMappingsRequest{
	//	Mappings: pbMappings,
	// }
	return fmt.Errorf("BulkCreateMappings not available - clients package deleted")
}

// DeleteByETCMeisaiID deletes all mappings for an ETC Meisai ID via gRPC
func (r *MappingGRPCRepository) DeleteByETCMeisaiID(etcMeisaiID int64) error {
	// TODO: Restore when clients package is available
	// ctx := context.Background()
	//
	// req := &pb.DeleteMappingsByETCMeisaiRequest{
	//	EtcMeisaiId: etcMeisaiID,
	// }
	return fmt.Errorf("DeleteMappingsByETCMeisai not available - clients package deleted")
}

// FindPotentialMatches finds potential DTako matches for an ETC Meisai record via gRPC
func (r *MappingGRPCRepository) FindPotentialMatches(etcMeisaiID int64, threshold float32) ([]*models.PotentialMatch, error) {
	// TODO: Restore when clients package is available
	// ctx := context.Background()
	//
	// req := &pb.FindPotentialMatchesRequest{
	//	EtcMeisaiId: etcMeisaiID,
	//	Threshold:   threshold,
	// }
	return nil, fmt.Errorf("FindPotentialMatches not available - clients package deleted")
}

// UpdateConfidenceScore updates the confidence score of a mapping via gRPC
func (r *MappingGRPCRepository) UpdateConfidenceScore(id int64, confidence float32) error {
	// TODO: Restore when clients package is available
	// ctx := context.Background()
	//
	// req := &pb.UpdateMappingConfidenceRequest{
	//	Id:         id,
	//	Confidence: confidence,
	// }
	return fmt.Errorf("UpdateMappingConfidence not available - clients package deleted")
}
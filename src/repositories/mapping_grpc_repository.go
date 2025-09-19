package repositories

import (
	"context"
	"fmt"

	"github.com/yhonda-ohishi/etc_meisai/src/clients"
	"github.com/yhonda-ohishi/etc_meisai/src/models"
	"github.com/yhonda-ohishi/etc_meisai/src/pb"
)

// MappingGRPCRepository implements MappingRepository interface using gRPC client
type MappingGRPCRepository struct {
	client *clients.DBServiceClient
}

// NewMappingGRPCRepository creates a new gRPC-based mapping repository
func NewMappingGRPCRepository(client *clients.DBServiceClient) MappingRepository {
	return &MappingGRPCRepository{
		client: client,
	}
}

// Create creates a new mapping record via gRPC
func (r *MappingGRPCRepository) Create(mapping *models.ETCMeisaiMapping) error {
	ctx := context.Background()

	req := &pb.CreateMappingRequest{
		EtcMeisaiId:  mapping.ETCMeisaiID,
		DtakoRowId:   mapping.DTakoRowID,
		MappingType:  mapping.MappingType,
		Confidence:   mapping.Confidence,
		Notes:        mapping.Notes,
	}

	resp, err := r.client.CreateMapping(ctx, req)
	if err != nil {
		return fmt.Errorf("gRPC create mapping failed: %w", err)
	}

	// Update the model with response data
	mapping.ID = resp.Id
	mapping.CreatedAt = resp.CreatedAt.AsTime()
	mapping.UpdatedAt = resp.UpdatedAt.AsTime()

	return nil
}

// GetByID retrieves a mapping by ID via gRPC
func (r *MappingGRPCRepository) GetByID(id int64) (*models.ETCMeisaiMapping, error) {
	ctx := context.Background()

	req := &pb.GetMappingRequest{
		Id: id,
	}

	resp, err := r.client.GetMapping(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("gRPC get mapping failed: %w", err)
	}

	// Convert response to model
	mapping := &models.ETCMeisaiMapping{
		ID:          resp.Id,
		ETCMeisaiID: resp.EtcMeisaiId,
		DTakoRowID:  resp.DtakoRowId,
		MappingType: resp.MappingType,
		Confidence:  resp.Confidence,
		Notes:       resp.Notes,
		CreatedAt:   resp.CreatedAt.AsTime(),
		UpdatedAt:   resp.UpdatedAt.AsTime(),
	}

	return mapping, nil
}

// Update updates an existing mapping record via gRPC
func (r *MappingGRPCRepository) Update(mapping *models.ETCMeisaiMapping) error {
	ctx := context.Background()

	req := &pb.UpdateMappingRequest{
		Id:          mapping.ID,
		MappingType: mapping.MappingType,
		Confidence:  mapping.Confidence,
		Notes:       mapping.Notes,
	}

	resp, err := r.client.UpdateMapping(ctx, req)
	if err != nil {
		return fmt.Errorf("gRPC update mapping failed: %w", err)
	}

	// Update timestamps
	mapping.UpdatedAt = resp.UpdatedAt.AsTime()

	return nil
}

// Delete deletes a mapping record via gRPC
func (r *MappingGRPCRepository) Delete(id int64) error {
	ctx := context.Background()

	req := &pb.DeleteMappingRequest{
		Id: id,
	}

	_, err := r.client.DeleteMapping(ctx, req)
	if err != nil {
		return fmt.Errorf("gRPC delete mapping failed: %w", err)
	}

	return nil
}

// GetByETCMeisaiID retrieves mappings by ETC Meisai ID via gRPC
func (r *MappingGRPCRepository) GetByETCMeisaiID(etcMeisaiID int64) ([]*models.ETCMeisaiMapping, error) {
	ctx := context.Background()

	req := &pb.ListMappingsRequest{
		EtcMeisaiId: etcMeisaiID,
		Limit:       100,
	}

	resp, err := r.client.ListMappings(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("gRPC list mappings failed: %w", err)
	}

	// Convert response to models
	var mappings []*models.ETCMeisaiMapping
	for _, pbMapping := range resp.Mappings {
		mapping := &models.ETCMeisaiMapping{
			ID:          pbMapping.Id,
			ETCMeisaiID: pbMapping.EtcMeisaiId,
			DTakoRowID:  pbMapping.DtakoRowId,
			MappingType: pbMapping.MappingType,
			Confidence:  pbMapping.Confidence,
			Notes:       pbMapping.Notes,
			CreatedAt:   pbMapping.CreatedAt.AsTime(),
			UpdatedAt:   pbMapping.UpdatedAt.AsTime(),
		}
		mappings = append(mappings, mapping)
	}

	return mappings, nil
}

// GetByDTakoRowID retrieves a mapping by DTako row ID via gRPC
func (r *MappingGRPCRepository) GetByDTakoRowID(dtakoRowID string) (*models.ETCMeisaiMapping, error) {
	ctx := context.Background()

	req := &pb.ListMappingsRequest{
		DtakoRowId: dtakoRowID,
		Limit:      1,
	}

	resp, err := r.client.ListMappings(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("gRPC list mappings failed: %w", err)
	}

	if len(resp.Mappings) == 0 {
		return nil, fmt.Errorf("mapping not found for DTako row ID: %s", dtakoRowID)
	}

	// Convert first result to model
	pbMapping := resp.Mappings[0]
	mapping := &models.ETCMeisaiMapping{
		ID:          pbMapping.Id,
		ETCMeisaiID: pbMapping.EtcMeisaiId,
		DTakoRowID:  pbMapping.DtakoRowId,
		MappingType: pbMapping.MappingType,
		Confidence:  pbMapping.Confidence,
		Notes:       pbMapping.Notes,
		CreatedAt:   pbMapping.CreatedAt.AsTime(),
		UpdatedAt:   pbMapping.UpdatedAt.AsTime(),
	}

	return mapping, nil
}

// List retrieves mappings with pagination via gRPC
func (r *MappingGRPCRepository) List(params *models.MappingListParams) ([]*models.ETCMeisaiMapping, int64, error) {
	ctx := context.Background()

	req := &pb.ListMappingsRequest{
		Limit:  int32(params.Limit),
		Offset: int32(params.Offset),
	}

	if params.ETCMeisaiID != nil {
		req.EtcMeisaiId = *params.ETCMeisaiID
	}
	if params.DTakoRowID != "" {
		req.DtakoRowId = params.DTakoRowID
	}
	if params.MappingType != "" {
		req.MappingType = params.MappingType
	}
	if params.MinConfidence != nil {
		req.MinConfidence = *params.MinConfidence
	}

	resp, err := r.client.ListMappings(ctx, req)
	if err != nil {
		return nil, 0, fmt.Errorf("gRPC list mappings failed: %w", err)
	}

	// Convert response to models
	var mappings []*models.ETCMeisaiMapping
	for _, pbMapping := range resp.Mappings {
		mapping := &models.ETCMeisaiMapping{
			ID:          pbMapping.Id,
			ETCMeisaiID: pbMapping.EtcMeisaiId,
			DTakoRowID:  pbMapping.DtakoRowId,
			MappingType: pbMapping.MappingType,
			Confidence:  pbMapping.Confidence,
			Notes:       pbMapping.Notes,
			CreatedAt:   pbMapping.CreatedAt.AsTime(),
			UpdatedAt:   pbMapping.UpdatedAt.AsTime(),
		}
		mappings = append(mappings, mapping)
	}

	return mappings, resp.Total, nil
}

// BulkCreateMappings creates multiple mappings via gRPC
func (r *MappingGRPCRepository) BulkCreateMappings(mappings []*models.ETCMeisaiMapping) error {
	ctx := context.Background()

	var pbMappings []*pb.CreateMappingRequest
	for _, mapping := range mappings {
		pbMapping := &pb.CreateMappingRequest{
			EtcMeisaiId:  mapping.ETCMeisaiID,
			DtakoRowId:   mapping.DTakoRowID,
			MappingType:  mapping.MappingType,
			Confidence:   mapping.Confidence,
			Notes:        mapping.Notes,
		}
		pbMappings = append(pbMappings, pbMapping)
	}

	req := &pb.BulkCreateMappingsRequest{
		Mappings: pbMappings,
	}

	resp, err := r.client.BulkCreateMappings(ctx, req)
	if err != nil {
		return fmt.Errorf("gRPC bulk create mappings failed: %w", err)
	}

	if !resp.Success {
		return fmt.Errorf("bulk create mappings failed: %s", resp.Message)
	}

	// Update IDs from response
	for i, createdID := range resp.CreatedIds {
		if i < len(mappings) {
			mappings[i].ID = createdID
		}
	}

	return nil
}

// DeleteByETCMeisaiID deletes all mappings for an ETC Meisai ID via gRPC
func (r *MappingGRPCRepository) DeleteByETCMeisaiID(etcMeisaiID int64) error {
	ctx := context.Background()

	req := &pb.DeleteMappingsByETCMeisaiRequest{
		EtcMeisaiId: etcMeisaiID,
	}

	_, err := r.client.DeleteMappingsByETCMeisai(ctx, req)
	if err != nil {
		return fmt.Errorf("gRPC delete mappings by ETC Meisai ID failed: %w", err)
	}

	return nil
}

// FindPotentialMatches finds potential DTako matches for an ETC Meisai record via gRPC
func (r *MappingGRPCRepository) FindPotentialMatches(etcMeisaiID int64, threshold float32) ([]*models.PotentialMatch, error) {
	ctx := context.Background()

	req := &pb.FindPotentialMatchesRequest{
		EtcMeisaiId: etcMeisaiID,
		Threshold:   threshold,
	}

	resp, err := r.client.FindPotentialMatches(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("gRPC find potential matches failed: %w", err)
	}

	// Convert response to models
	var matches []*models.PotentialMatch
	for _, pbMatch := range resp.Matches {
		match := &models.PotentialMatch{
			DTakoRowID:   pbMatch.DtakoRowId,
			Confidence:   pbMatch.Confidence,
			MatchReasons: pbMatch.MatchReasons,
			DTakoData:    make(map[string]interface{}),
		}

		// Copy DTako data directly (it's already a map[string]interface{} in the stub)
		match.DTakoData = pbMatch.DtakoData

		matches = append(matches, match)
	}

	return matches, nil
}

// UpdateConfidenceScore updates the confidence score of a mapping via gRPC
func (r *MappingGRPCRepository) UpdateConfidenceScore(id int64, confidence float32) error {
	ctx := context.Background()

	req := &pb.UpdateMappingConfidenceRequest{
		Id:         id,
		Confidence: confidence,
	}

	_, err := r.client.UpdateMappingConfidence(ctx, req)
	if err != nil {
		return fmt.Errorf("gRPC update mapping confidence failed: %w", err)
	}

	return nil
}
package services

import (
	"context"
	"fmt"
	"time"

	"github.com/yhonda-ohishi/etc_meisai/src/models"
	"github.com/yhonda-ohishi/etc_meisai/src/repositories"
)

// MappingService handles ETC-DTako mapping operations
type MappingService struct {
	mappingRepo repositories.MappingRepository
	etcRepo     repositories.ETCRepository
}

// NewMappingService creates a new mapping service
func NewMappingService(mappingRepo repositories.MappingRepository, etcRepo repositories.ETCRepository) *MappingService {
	return &MappingService{
		mappingRepo: mappingRepo,
		etcRepo:     etcRepo,
	}
}

// CreateMapping creates a new mapping between ETC and DTako records
func (s *MappingService) CreateMapping(ctx context.Context, mapping *models.ETCMeisaiMapping) error {
	// Validate ETC Meisai exists
	_, err := s.etcRepo.GetByID(mapping.ETCMeisaiID)
	if err != nil {
		return fmt.Errorf("ETC Meisai record not found: %w", err)
	}

	// Check for existing mapping
	existing, err := s.mappingRepo.GetByETCMeisaiID(mapping.ETCMeisaiID)
	if err == nil && len(existing) > 0 {
		// If mapping exists, update it instead
		existing[0].DTakoRowID = mapping.DTakoRowID
		existing[0].MappingType = mapping.MappingType
		existing[0].Confidence = mapping.Confidence
		existing[0].Notes = mapping.Notes
		return s.mappingRepo.Update(existing[0])
	}

	// Create new mapping
	mapping.CreatedAt = time.Now()
	mapping.UpdatedAt = time.Now()

	return s.mappingRepo.Create(mapping)
}

// GetMappingByID retrieves a mapping by its ID
func (s *MappingService) GetMappingByID(ctx context.Context, id int64) (*models.ETCMeisaiMapping, error) {
	return s.mappingRepo.GetByID(id)
}

// GetMappingsByETCMeisaiID retrieves all mappings for an ETC Meisai record
func (s *MappingService) GetMappingsByETCMeisaiID(ctx context.Context, etcMeisaiID int64) ([]*models.ETCMeisaiMapping, error) {
	return s.mappingRepo.GetByETCMeisaiID(etcMeisaiID)
}

// GetMappingByDTakoRowID retrieves a mapping by DTako row ID
func (s *MappingService) GetMappingByDTakoRowID(ctx context.Context, dtakoRowID string) (*models.ETCMeisaiMapping, error) {
	return s.mappingRepo.GetByDTakoRowID(dtakoRowID)
}

// ListMappings lists mappings with pagination and filters
func (s *MappingService) ListMappings(ctx context.Context, params *models.MappingListParams) ([]*models.ETCMeisaiMapping, int64, error) {
	// Set defaults
	if params == nil {
		params = &models.MappingListParams{
			Limit:  100,
			Offset: 0,
		}
	}
	if params.Limit <= 0 {
		params.Limit = 100
	}
	if params.Limit > 1000 {
		params.Limit = 1000
	}

	mappings, total, err := s.mappingRepo.List(params)
	if err != nil {
		return nil, 0, err
	}

	// Ensure non-nil slice
	if mappings == nil {
		mappings = []*models.ETCMeisaiMapping{}
	}

	return mappings, total, nil
}

// UpdateMapping updates an existing mapping
func (s *MappingService) UpdateMapping(ctx context.Context, id int64, updates map[string]interface{}) error {
	// Get existing mapping
	mapping, err := s.mappingRepo.GetByID(id)
	if err != nil {
		return fmt.Errorf("mapping not found: %w", err)
	}

	// Apply updates
	if mappingType, ok := updates["mapping_type"].(string); ok {
		mapping.MappingType = mappingType
	}
	if confidence, ok := updates["confidence"].(float32); ok {
		mapping.Confidence = confidence
	}
	if notes, ok := updates["notes"].(string); ok {
		mapping.Notes = notes
	}

	mapping.UpdatedAt = time.Now()

	return s.mappingRepo.Update(mapping)
}

// DeleteMapping deletes a mapping
func (s *MappingService) DeleteMapping(ctx context.Context, id int64) error {
	return s.mappingRepo.Delete(id)
}

// DeleteMappingsByETCMeisaiID deletes all mappings for an ETC Meisai record
func (s *MappingService) DeleteMappingsByETCMeisaiID(ctx context.Context, etcMeisaiID int64) error {
	return s.mappingRepo.DeleteByETCMeisaiID(etcMeisaiID)
}

// AutoMatch finds potential matches for unmapped ETC records
func (s *MappingService) AutoMatch(ctx context.Context, startDate, endDate time.Time, threshold float32) ([]*models.AutoMatchResult, error) {
	// Get unmapped ETC records
	unmapped, err := s.etcRepo.GetByDateRange(startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get unmapped records: %w", err)
	}

	results := make([]*models.AutoMatchResult, 0)

	for _, etcRecord := range unmapped {
		// Check if already mapped
		existing, err := s.mappingRepo.GetByETCMeisaiID(etcRecord.ID)
		if err == nil && len(existing) > 0 {
			continue // Already mapped
		}

		// Find potential matches
		matches, err := s.mappingRepo.FindPotentialMatches(etcRecord.ID, threshold)
		if err != nil {
			// Log error but continue with other records
			results = append(results, &models.AutoMatchResult{
				ETCMeisaiID: etcRecord.ID,
				Error:       err.Error(),
			})
			continue
		}

		if len(matches) > 0 {
			results = append(results, &models.AutoMatchResult{
				ETCMeisaiID:      etcRecord.ID,
				PotentialMatches: matches,
				BestMatch:        matches[0], // Assuming sorted by confidence
			})
		}
	}

	return results, nil
}

// BulkCreateMappings creates multiple mappings at once
func (s *MappingService) BulkCreateMappings(ctx context.Context, mappings []*models.ETCMeisaiMapping) error {
	// Validate all ETC Meisai records exist
	for _, mapping := range mappings {
		_, err := s.etcRepo.GetByID(mapping.ETCMeisaiID)
		if err != nil {
			return fmt.Errorf("ETC Meisai record %d not found: %w", mapping.ETCMeisaiID, err)
		}
	}

	// Set timestamps
	now := time.Now()
	for _, mapping := range mappings {
		mapping.CreatedAt = now
		mapping.UpdatedAt = now
	}

	return s.mappingRepo.BulkCreateMappings(mappings)
}

// UpdateConfidenceScore updates the confidence score of a mapping
func (s *MappingService) UpdateConfidenceScore(ctx context.Context, id int64, confidence float32) error {
	// Validate confidence range
	if confidence < 0 || confidence > 1 {
		return fmt.Errorf("confidence must be between 0 and 1")
	}

	return s.mappingRepo.UpdateConfidenceScore(id, confidence)
}

// GetMappingStats returns statistics about mappings
func (s *MappingService) GetMappingStats(ctx context.Context, startDate, endDate time.Time) (*models.MappingStats, error) {
	// Get all mappings
	params := &models.MappingListParams{
		Limit:  10000,
		Offset: 0,
	}

	mappings, total, err := s.mappingRepo.List(params)
	if err != nil {
		return nil, fmt.Errorf("failed to get mappings: %w", err)
	}

	stats := &models.MappingStats{
		TotalMappings: total,
	}

	var totalConfidence float32
	for _, mapping := range mappings {
		if mapping.MappingType == "auto" {
			stats.AutoMappings++
		} else {
			stats.ManualMappings++
		}

		if mapping.Confidence >= 0.8 {
			stats.HighConfidence++
		} else {
			stats.LowConfidence++
		}

		totalConfidence += mapping.Confidence
	}

	if len(mappings) > 0 {
		stats.AverageConfidence = totalConfidence / float32(len(mappings))
	}

	// Count unmapped records
	etcCount, err := s.etcRepo.CountByDateRange(startDate, endDate)
	if err == nil {
		stats.UnmappedRecords = etcCount - total
	}

	return stats, nil
}

// HealthCheck performs a health check on the mapping service
func (s *MappingService) HealthCheck(ctx context.Context) error {
	// Try to list one mapping to verify connectivity
	params := &models.MappingListParams{
		Limit:  1,
		Offset: 0,
	}

	_, _, err := s.mappingRepo.List(params)
	if err != nil {
		return fmt.Errorf("mapping service health check failed: %w", err)
	}

	return nil
}
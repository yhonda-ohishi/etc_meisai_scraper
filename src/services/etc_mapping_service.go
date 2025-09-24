package services

import (
	"context"
	"fmt"
	"log"

	"github.com/yhonda-ohishi/etc_meisai/src/models"
	"github.com/yhonda-ohishi/etc_meisai/src/repositories"
)

// CreateMappingParams contains parameters for creating an ETC mapping
type CreateMappingParams struct {
	ETCRecordID      int64                  `json:"etc_record_id" validate:"required,min=1"`
	MappingType      string                 `json:"mapping_type" validate:"required"`
	MappedEntityID   int64                  `json:"mapped_entity_id" validate:"required,min=1"`
	MappedEntityType string                 `json:"mapped_entity_type" validate:"required"`
	Confidence       float32                `json:"confidence" validate:"min=0,max=1"`
	Status           string                 `json:"status,omitempty"`
	Metadata         map[string]interface{} `json:"metadata,omitempty"`
	CreatedBy        string                 `json:"created_by,omitempty"`
}

// ListMappingsParams contains parameters for listing ETC mappings
type ListMappingsParams struct {
	Page             int     `json:"page" validate:"min=1"`
	PageSize         int     `json:"page_size" validate:"min=1,max=1000"`
	ETCRecordID      *int64  `json:"etc_record_id,omitempty"`
	MappingType      *string `json:"mapping_type,omitempty"`
	MappedEntityID   *int64  `json:"mapped_entity_id,omitempty"`
	MappedEntityType *string `json:"mapped_entity_type,omitempty"`
	Status           *string `json:"status,omitempty"`
	MinConfidence    *float32 `json:"min_confidence,omitempty"`
	MaxConfidence    *float32 `json:"max_confidence,omitempty"`
	CreatedBy        *string `json:"created_by,omitempty"`
	SortBy           string  `json:"sort_by"`     // created_at, confidence, etc_record_id
	SortOrder        string  `json:"sort_order"`  // asc, desc
}

// ListMappingsResponse contains the response for listing ETC mappings
type ListMappingsResponse struct {
	Mappings   []*models.ETCMapping `json:"mappings"`
	TotalCount int64                `json:"total_count"`
	Page       int                  `json:"page"`
	PageSize   int                  `json:"page_size"`
	TotalPages int                  `json:"total_pages"`
}

// UpdateMappingParams contains parameters for updating an ETC mapping
type UpdateMappingParams struct {
	MappingType      *string                `json:"mapping_type,omitempty"`
	MappedEntityID   *int64                 `json:"mapped_entity_id,omitempty"`
	MappedEntityType *string                `json:"mapped_entity_type,omitempty"`
	Confidence       *float32               `json:"confidence,omitempty"`
	Status           *string                `json:"status,omitempty"`
	Metadata         map[string]interface{} `json:"metadata,omitempty"`
}

// ETCMappingService handles business logic for ETC mapping management using repository pattern
type ETCMappingService struct {
	mappingRepo repositories.ETCMappingRepository
	recordRepo  repositories.ETCMeisaiRecordRepository
	logger      *log.Logger
}

// NewETCMappingService creates a new ETC mapping management service with repository pattern
func NewETCMappingService(
	mappingRepo repositories.ETCMappingRepository,
	recordRepo repositories.ETCMeisaiRecordRepository,
	logger *log.Logger,
) *ETCMappingService {
	if logger == nil {
		logger = log.New(log.Writer(), "[ETCMappingService] ", log.LstdFlags|log.Lshortfile)
	}

	return &ETCMappingService{
		mappingRepo: mappingRepo,
		recordRepo:  recordRepo,
		logger:      logger,
	}
}

// CreateMapping creates a new ETC mapping with validation
func (s *ETCMappingService) CreateMapping(ctx context.Context, params *CreateMappingParams) (*models.ETCMapping, error) {
	if s.logger != nil {
		s.logger.Printf("Creating ETC mapping for record ID: %d, entity ID: %d", params.ETCRecordID, params.MappedEntityID)
	}

	// Set defaults
	if params.Confidence == 0 {
		params.Confidence = 1.0
	}
	if params.Status == "" {
		params.Status = string(models.MappingStatusActive)
	}

	// Create mapping model
	mapping := &models.ETCMapping{
		ETCRecordID:      params.ETCRecordID,
		MappingType:      params.MappingType,
		MappedEntityID:   params.MappedEntityID,
		MappedEntityType: params.MappedEntityType,
		Confidence:       params.Confidence,
		Status:           params.Status,
		CreatedBy:        params.CreatedBy,
	}

	// Set metadata if provided
	if params.Metadata != nil {
		if err := mapping.SetMetadata(params.Metadata); err != nil {
			s.logger.Printf("Failed to set metadata: %v", err)
			return nil, fmt.Errorf("failed to set metadata: %w", err)
		}
	}

	// Start transaction
	txRepo, err := s.mappingRepo.BeginTx(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to start transaction: %w", err)
	}
	defer func() {
		if r := recover(); r != nil {
			txRepo.RollbackTx()
		}
	}()

	// Verify ETC record exists
	_, err = s.recordRepo.GetByID(ctx, params.ETCRecordID)
	if err != nil {
		txRepo.RollbackTx()
		return nil, fmt.Errorf("ETC record not found with ID %d: %w", params.ETCRecordID, err)
	}

	// Check for existing active mappings
	existingMapping, err := txRepo.GetActiveMapping(ctx, params.ETCRecordID)
	if err == nil && existingMapping != nil {
		txRepo.RollbackTx()
		return nil, fmt.Errorf("active mapping already exists for ETC record %d", params.ETCRecordID)
	}

	// Create the mapping
	if err := txRepo.Create(ctx, mapping); err != nil {
		txRepo.RollbackTx()
		s.logger.Printf("Failed to create mapping: %v", err)
		return nil, fmt.Errorf("failed to create mapping: %w", err)
	}

	// Commit transaction
	if err := txRepo.CommitTx(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	s.logger.Printf("Successfully created ETC mapping with ID: %d", mapping.ID)
	return mapping, nil
}

// GetMapping retrieves an ETC mapping by ID
func (s *ETCMappingService) GetMapping(ctx context.Context, id int64) (*models.ETCMapping, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid mapping ID: %d", id)
	}

	s.logger.Printf("Retrieving ETC mapping with ID: %d", id)

	mapping, err := s.mappingRepo.GetByID(ctx, id)
	if err != nil {
		s.logger.Printf("Failed to retrieve mapping: %v", err)
		return nil, fmt.Errorf("failed to retrieve mapping: %w", err)
	}

	return mapping, nil
}

// ListMappings lists ETC mappings with filtering and pagination
func (s *ETCMappingService) ListMappings(ctx context.Context, params *ListMappingsParams) (*ListMappingsResponse, error) {
	// Set defaults
	if params.Page <= 0 {
		params.Page = 1
	}
	if params.PageSize <= 0 {
		params.PageSize = 50
	}
	if params.PageSize > 1000 {
		params.PageSize = 1000
	}
	if params.SortBy == "" {
		params.SortBy = "created_at"
	}
	if params.SortOrder == "" {
		params.SortOrder = "desc"
	}

	s.logger.Printf("Listing ETC mappings - page: %d, size: %d", params.Page, params.PageSize)

	// Convert to repository params
	repoParams := repositories.ListMappingsParams{
		Page:             params.Page,
		PageSize:         params.PageSize,
		MappingType:      params.MappingType,
		Status:           params.Status,
		MinConfidence:    params.MinConfidence,
		MappedEntityType: params.MappedEntityType,
		SortBy:           params.SortBy,
		SortOrder:        params.SortOrder,
	}

	// Get mappings from repository
	mappings, totalCount, err := s.mappingRepo.List(ctx, repoParams)
	if err != nil {
		s.logger.Printf("Failed to retrieve mappings: %v", err)
		return nil, fmt.Errorf("failed to retrieve mappings: %w", err)
	}

	totalPages := int((totalCount + int64(params.PageSize) - 1) / int64(params.PageSize))

	// Use mappings directly since they're already ETCMapping

	response := &ListMappingsResponse{
		Mappings:   mappings,
		TotalCount: totalCount,
		Page:       params.Page,
		PageSize:   params.PageSize,
		TotalPages: totalPages,
	}

	s.logger.Printf("Successfully retrieved %d mappings (page %d of %d)", len(mappings), params.Page, totalPages)
	return response, nil
}

// UpdateMapping updates an existing ETC mapping
func (s *ETCMappingService) UpdateMapping(ctx context.Context, id int64, params *UpdateMappingParams) (*models.ETCMapping, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid mapping ID: %d", id)
	}

	s.logger.Printf("Updating ETC mapping with ID: %d", id)

	// Start transaction
	txRepo, err := s.mappingRepo.BeginTx(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to start transaction: %w", err)
	}
	defer func() {
		if r := recover(); r != nil {
			txRepo.RollbackTx()
		}
	}()

	// Get existing mapping
	mapping, err := txRepo.GetByID(ctx, id)
	if err != nil {
		txRepo.RollbackTx()
		return nil, fmt.Errorf("failed to retrieve mapping: %w", err)
	}

	// Update fields
	if params.MappingType != nil {
		mapping.MappingType = *params.MappingType
	}
	if params.MappedEntityID != nil {
		mapping.MappedEntityID = *params.MappedEntityID
	}
	if params.MappedEntityType != nil {
		mapping.MappedEntityType = *params.MappedEntityType
	}
	if params.Confidence != nil {
		mapping.Confidence = *params.Confidence
	}
	if params.Status != nil {
		mapping.Status = *params.Status
	}
	if params.Metadata != nil {
		if err := mapping.SetMetadata(params.Metadata); err != nil {
			txRepo.RollbackTx()
			s.logger.Printf("Failed to set metadata: %v", err)
			return nil, fmt.Errorf("failed to set metadata: %w", err)
		}
	}

	// Save the updated mapping
	if err := txRepo.Update(ctx, mapping); err != nil {
		txRepo.RollbackTx()
		s.logger.Printf("Failed to update mapping: %v", err)
		return nil, fmt.Errorf("failed to update mapping: %w", err)
	}

	// Commit transaction
	if err := txRepo.CommitTx(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	s.logger.Printf("Successfully updated ETC mapping with ID: %d", mapping.ID)
	return mapping, nil
}

// DeleteMapping performs soft delete on an ETC mapping
func (s *ETCMappingService) DeleteMapping(ctx context.Context, id int64) error {
	if id <= 0 {
		return fmt.Errorf("invalid mapping ID: %d", id)
	}

	s.logger.Printf("Deleting ETC mapping with ID: %d", id)

	// Start transaction
	txRepo, err := s.mappingRepo.BeginTx(ctx)
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}
	defer func() {
		if r := recover(); r != nil {
			txRepo.RollbackTx()
		}
	}()

	// Check if mapping exists
	_, err = txRepo.GetByID(ctx, id)
	if err != nil {
		txRepo.RollbackTx()
		return fmt.Errorf("failed to retrieve mapping: %w", err)
	}

	// Perform soft delete
	if err := txRepo.Delete(ctx, id); err != nil {
		txRepo.RollbackTx()
		s.logger.Printf("Failed to delete mapping: %v", err)
		return fmt.Errorf("failed to delete mapping: %w", err)
	}

	// Commit transaction
	if err := txRepo.CommitTx(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	s.logger.Printf("Successfully deleted ETC mapping with ID: %d", id)
	return nil
}

// UpdateStatus updates the status of a mapping
func (s *ETCMappingService) UpdateStatus(ctx context.Context, id int64, status string) error {
	if id <= 0 {
		return fmt.Errorf("invalid mapping ID: %d", id)
	}

	s.logger.Printf("Updating status for mapping ID %d to: %s", id, status)

	err := s.mappingRepo.UpdateStatus(ctx, id, status)
	if err != nil {
		s.logger.Printf("Failed to update mapping status: %v", err)
		return fmt.Errorf("failed to update mapping status: %w", err)
	}

	s.logger.Printf("Successfully updated mapping status")
	return nil
}

// HealthCheck performs health check for the service
func (s *ETCMappingService) HealthCheck(ctx context.Context) error {
	if s.mappingRepo == nil {
		return fmt.Errorf("mapping repository not initialized")
	}

	// Check repository connectivity
	if err := s.mappingRepo.Ping(ctx); err != nil {
		return fmt.Errorf("mapping repository ping failed: %w", err)
	}

	if s.recordRepo != nil {
		if err := s.recordRepo.Ping(ctx); err != nil {
			return fmt.Errorf("record repository ping failed: %w", err)
		}
	}

	return nil
}
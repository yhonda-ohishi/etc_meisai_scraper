package services

import (
	"context"
	"fmt"
	"log"

	"gorm.io/gorm"

	"github.com/yhonda-ohishi/etc_meisai/src/models"
)

// ETCMappingService handles business logic for ETC mapping management
type ETCMappingService struct {
	db     *gorm.DB
	logger *log.Logger
}

// NewETCMappingService creates a new ETC mapping management service
func NewETCMappingService(db *gorm.DB, logger *log.Logger) *ETCMappingService {
	if logger == nil {
		logger = log.New(log.Writer(), "[ETCMappingService] ", log.LstdFlags|log.Lshortfile)
	}

	return &ETCMappingService{
		db:     db,
		logger: logger,
	}
}

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

// StatusTransition represents a status transition
type StatusTransition struct {
	From   string `json:"from"`
	To     string `json:"to"`
	Reason string `json:"reason,omitempty"`
}

// CreateMapping creates a new ETC mapping with validation
func (s *ETCMappingService) CreateMapping(ctx context.Context, params *CreateMappingParams) (*models.ETCMapping, error) {
	s.logger.Printf("Creating ETC mapping for record ID: %d, entity ID: %d", params.ETCRecordID, params.MappedEntityID)

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
	tx := s.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return nil, fmt.Errorf("failed to start transaction: %w", tx.Error)
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Verify ETC record exists
	var etcRecord models.ETCMeisaiRecord
	err := tx.First(&etcRecord, params.ETCRecordID).Error
	if err == gorm.ErrRecordNotFound {
		tx.Rollback()
		return nil, fmt.Errorf("ETC record not found with ID: %d", params.ETCRecordID)
	} else if err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to verify ETC record: %w", err)
	}

	// Check for existing active mappings for the same ETC record and mapping type
	var existingMapping models.ETCMapping
	err = tx.Where("etc_record_id = ? AND mapping_type = ? AND status = ?",
		params.ETCRecordID, params.MappingType, models.MappingStatusActive).First(&existingMapping).Error
	if err == nil {
		tx.Rollback()
		return nil, fmt.Errorf("active mapping already exists for ETC record %d with type %s",
			params.ETCRecordID, params.MappingType)
	} else if err != gorm.ErrRecordNotFound {
		tx.Rollback()
		return nil, fmt.Errorf("failed to check for existing mappings: %w", err)
	}

	// Validate the mapping
	if err := mapping.BeforeCreate(tx); err != nil {
		tx.Rollback()
		s.logger.Printf("Validation failed for mapping: %v", err)
		return nil, fmt.Errorf("mapping validation failed: %w", err)
	}

	// Create the mapping
	if err := tx.Create(mapping).Error; err != nil {
		tx.Rollback()
		s.logger.Printf("Failed to create mapping: %v", err)
		return nil, fmt.Errorf("failed to create mapping: %w", err)
	}

	// Load the ETC record relationship
	if err := tx.Preload("ETCRecord").First(mapping, mapping.ID).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to load mapping with ETC record: %w", err)
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
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

	var mapping models.ETCMapping
	err := s.db.WithContext(ctx).Preload("ETCRecord").First(&mapping, id).Error
	if err == gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("mapping not found with ID: %d", id)
	} else if err != nil {
		s.logger.Printf("Failed to retrieve mapping: %v", err)
		return nil, fmt.Errorf("failed to retrieve mapping: %w", err)
	}

	return &mapping, nil
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

	// Build query
	query := s.db.WithContext(ctx).Model(&models.ETCMapping{})

	// Apply filters
	if params.ETCRecordID != nil {
		query = query.Where("etc_record_id = ?", *params.ETCRecordID)
	}
	if params.MappingType != nil && *params.MappingType != "" {
		query = query.Where("mapping_type = ?", *params.MappingType)
	}
	if params.MappedEntityID != nil {
		query = query.Where("mapped_entity_id = ?", *params.MappedEntityID)
	}
	if params.MappedEntityType != nil && *params.MappedEntityType != "" {
		query = query.Where("mapped_entity_type = ?", *params.MappedEntityType)
	}
	if params.Status != nil && *params.Status != "" {
		query = query.Where("status = ?", *params.Status)
	}
	if params.MinConfidence != nil {
		query = query.Where("confidence >= ?", *params.MinConfidence)
	}
	if params.MaxConfidence != nil {
		query = query.Where("confidence <= ?", *params.MaxConfidence)
	}
	if params.CreatedBy != nil && *params.CreatedBy != "" {
		query = query.Where("created_by LIKE ?", "%"+*params.CreatedBy+"%")
	}

	// Get total count
	var totalCount int64
	if err := query.Count(&totalCount).Error; err != nil {
		s.logger.Printf("Failed to count mappings: %v", err)
		return nil, fmt.Errorf("failed to count mappings: %w", err)
	}

	// Apply sorting and pagination
	orderClause := fmt.Sprintf("%s %s", params.SortBy, params.SortOrder)
	offset := (params.Page - 1) * params.PageSize

	var mappings []*models.ETCMapping
	err := query.Preload("ETCRecord").Order(orderClause).Offset(offset).Limit(params.PageSize).Find(&mappings).Error
	if err != nil {
		s.logger.Printf("Failed to retrieve mappings: %v", err)
		return nil, fmt.Errorf("failed to retrieve mappings: %w", err)
	}

	totalPages := int((totalCount + int64(params.PageSize) - 1) / int64(params.PageSize))

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

// UpdateMapping updates an existing ETC mapping with status transitions
func (s *ETCMappingService) UpdateMapping(ctx context.Context, id int64, params *UpdateMappingParams) (*models.ETCMapping, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid mapping ID: %d", id)
	}

	s.logger.Printf("Updating ETC mapping with ID: %d", id)

	// Start transaction
	tx := s.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return nil, fmt.Errorf("failed to start transaction: %w", tx.Error)
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Get existing mapping
	var mapping models.ETCMapping
	err := tx.Preload("ETCRecord").First(&mapping, id).Error
	if err == gorm.ErrRecordNotFound {
		tx.Rollback()
		return nil, fmt.Errorf("mapping not found with ID: %d", id)
	} else if err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to retrieve mapping: %w", err)
	}

	// Store old status for validation
	oldStatus := mapping.Status

	// Update fields if provided
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

	// Validate status transition if status is being changed
	if params.Status != nil && *params.Status != oldStatus {
		if err := s.ValidateStatusTransition(oldStatus, *params.Status); err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("invalid status transition: %w", err)
		}
	}

	// Update metadata if provided
	if params.Metadata != nil {
		if err := mapping.SetMetadata(params.Metadata); err != nil {
			tx.Rollback()
			s.logger.Printf("Failed to set metadata: %v", err)
			return nil, fmt.Errorf("failed to set metadata: %w", err)
		}
	}

	// Validate the updated mapping
	if err := mapping.BeforeSave(tx); err != nil {
		tx.Rollback()
		s.logger.Printf("Validation failed for updated mapping: %v", err)
		return nil, fmt.Errorf("mapping validation failed: %w", err)
	}

	// Save the updated mapping
	if err := tx.Save(&mapping).Error; err != nil {
		tx.Rollback()
		s.logger.Printf("Failed to update mapping: %v", err)
		return nil, fmt.Errorf("failed to update mapping: %w", err)
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	s.logger.Printf("Successfully updated ETC mapping with ID: %d", mapping.ID)
	return &mapping, nil
}

// DeleteMapping deletes an ETC mapping
func (s *ETCMappingService) DeleteMapping(ctx context.Context, id int64) error {
	if id <= 0 {
		return fmt.Errorf("invalid mapping ID: %d", id)
	}

	s.logger.Printf("Deleting ETC mapping with ID: %d", id)

	// Start transaction
	tx := s.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return fmt.Errorf("failed to start transaction: %w", tx.Error)
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Check if mapping exists
	var mapping models.ETCMapping
	err := tx.First(&mapping, id).Error
	if err == gorm.ErrRecordNotFound {
		tx.Rollback()
		return fmt.Errorf("mapping not found with ID: %d", id)
	} else if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to retrieve mapping: %w", err)
	}

	// Perform hard delete (mappings are typically hard deleted)
	if err := tx.Unscoped().Delete(&mapping).Error; err != nil {
		tx.Rollback()
		s.logger.Printf("Failed to delete mapping: %v", err)
		return fmt.Errorf("failed to delete mapping: %w", err)
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	s.logger.Printf("Successfully deleted ETC mapping with ID: %d", id)
	return nil
}

// ValidateStatusTransition validates status transitions according to business rules
func (s *ETCMappingService) ValidateStatusTransition(fromStatus, toStatus string) error {
	from := models.MappingStatus(fromStatus)
	to := models.MappingStatus(toStatus)

	// Check if transition is allowed
	switch from {
	case models.MappingStatusPending:
		if to != models.MappingStatusActive && to != models.MappingStatusRejected {
			return fmt.Errorf("pending mappings can only transition to active or rejected")
		}
	case models.MappingStatusActive:
		if to != models.MappingStatusInactive {
			return fmt.Errorf("active mappings can only transition to inactive")
		}
	case models.MappingStatusInactive:
		if to != models.MappingStatusActive {
			return fmt.Errorf("inactive mappings can only transition to active")
		}
	case models.MappingStatusRejected:
		if to != models.MappingStatusPending {
			return fmt.Errorf("rejected mappings can only transition to pending")
		}
	default:
		return fmt.Errorf("unknown status: %s", fromStatus)
	}

	return nil
}

// GetMappingsByETCRecord retrieves all mappings for a specific ETC record
func (s *ETCMappingService) GetMappingsByETCRecord(ctx context.Context, etcRecordID int64) ([]*models.ETCMapping, error) {
	if etcRecordID <= 0 {
		return nil, fmt.Errorf("invalid ETC record ID: %d", etcRecordID)
	}

	s.logger.Printf("Retrieving mappings for ETC record ID: %d", etcRecordID)

	var mappings []*models.ETCMapping
	err := s.db.WithContext(ctx).Preload("ETCRecord").Where("etc_record_id = ?", etcRecordID).Find(&mappings).Error
	if err != nil {
		s.logger.Printf("Failed to retrieve mappings for ETC record: %v", err)
		return nil, fmt.Errorf("failed to retrieve mappings: %w", err)
	}

	return mappings, nil
}

// GetActiveMappingsByETCRecord retrieves active mappings for a specific ETC record
func (s *ETCMappingService) GetActiveMappingsByETCRecord(ctx context.Context, etcRecordID int64) ([]*models.ETCMapping, error) {
	if etcRecordID <= 0 {
		return nil, fmt.Errorf("invalid ETC record ID: %d", etcRecordID)
	}

	s.logger.Printf("Retrieving active mappings for ETC record ID: %d", etcRecordID)

	var mappings []*models.ETCMapping
	err := s.db.WithContext(ctx).Preload("ETCRecord").Where("etc_record_id = ? AND status = ?",
		etcRecordID, models.MappingStatusActive).Find(&mappings).Error
	if err != nil {
		s.logger.Printf("Failed to retrieve active mappings for ETC record: %v", err)
		return nil, fmt.Errorf("failed to retrieve active mappings: %w", err)
	}

	return mappings, nil
}

// UpdateMappingStatus updates only the status of a mapping with validation
func (s *ETCMappingService) UpdateMappingStatus(ctx context.Context, id int64, newStatus string) (*models.ETCMapping, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid mapping ID: %d", id)
	}

	s.logger.Printf("Updating status of mapping ID: %d to: %s", id, newStatus)

	// Start transaction
	tx := s.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return nil, fmt.Errorf("failed to start transaction: %w", tx.Error)
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Get existing mapping
	var mapping models.ETCMapping
	err := tx.Preload("ETCRecord").First(&mapping, id).Error
	if err == gorm.ErrRecordNotFound {
		tx.Rollback()
		return nil, fmt.Errorf("mapping not found with ID: %d", id)
	} else if err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to retrieve mapping: %w", err)
	}

	// Validate status transition
	if err := s.ValidateStatusTransition(mapping.Status, newStatus); err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("invalid status transition: %w", err)
	}

	// Update status
	mapping.Status = newStatus

	// Save the updated mapping
	if err := tx.Save(&mapping).Error; err != nil {
		tx.Rollback()
		s.logger.Printf("Failed to update mapping status: %v", err)
		return nil, fmt.Errorf("failed to update mapping status: %w", err)
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	s.logger.Printf("Successfully updated mapping status with ID: %d", mapping.ID)
	return &mapping, nil
}

// HealthCheck performs health check for the service
func (s *ETCMappingService) HealthCheck(ctx context.Context) error {
	// Check database connectivity
	sqlDB, err := s.db.DB()
	if err != nil {
		return fmt.Errorf("failed to get database connection: %w", err)
	}

	if err := sqlDB.PingContext(ctx); err != nil {
		return fmt.Errorf("database ping failed: %w", err)
	}

	return nil
}
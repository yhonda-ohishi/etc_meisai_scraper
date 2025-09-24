package models

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// MappingStatus represents the status of a mapping
type MappingStatus string

const (
	MappingStatusActive   MappingStatus = "active"
	MappingStatusInactive MappingStatus = "inactive"
	MappingStatusPending  MappingStatus = "pending"
	MappingStatusRejected MappingStatus = "rejected"
)

// MappingType represents the type of mapping
type MappingType string

const (
	MappingTypeDtako   MappingType = "dtako"
	MappingTypeExpense MappingType = "expense"
	MappingTypeInvoice MappingType = "invoice"
)

// MappedEntityType represents the type of mapped entity
type MappedEntityType string

const (
	EntityTypeDtakoRecord   MappedEntityType = "dtako_record"
	EntityTypeExpenseRecord MappedEntityType = "expense_record"
	EntityTypeInvoiceRecord MappedEntityType = "invoice_record"
)

// ETCMapping represents the mapping between ETC records and other entities
type ETCMapping struct {
	ID               int64            `gorm:"primaryKey;autoIncrement" json:"id"`
	ETCRecordID      int64            `gorm:"not null;index" json:"etc_record_id"`
	ETCRecord        ETCMeisaiRecord  `gorm:"foreignKey:ETCRecordID" json:"etc_record,omitempty"`
	MappingType      string           `gorm:"size:50;not null;index" json:"mapping_type"`
	MappedEntityID   int64            `gorm:"not null;index" json:"mapped_entity_id"`
	MappedEntityType string           `gorm:"size:50;not null;index" json:"mapped_entity_type"`
	Confidence       float32          `gorm:"default:1.0" json:"confidence"`
	Status           string           `gorm:"size:20;default:'active';index" json:"status"`
	Metadata         datatypes.JSON   `gorm:"type:json" json:"metadata,omitempty"`
	CreatedBy        string           `gorm:"size:100" json:"created_by,omitempty"`
	CreatedAt        time.Time        `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt        time.Time        `gorm:"autoUpdateTime" json:"updated_at"`
}

// TableName returns the table name for GORM
func (ETCMapping) TableName() string {
	return "etc_mappings"
}

// BeforeCreate hook to validate data before creating
func (m *ETCMapping) BeforeCreate(tx *gorm.DB) error {
	// Set timestamps manually if not using GORM auto-timestamps
	now := time.Now()
	if m.CreatedAt.IsZero() {
		m.CreatedAt = now
	}
	if m.UpdatedAt.IsZero() {
		m.UpdatedAt = now
	}

	// Set default status if not provided
	if m.Status == "" {
		m.Status = string(MappingStatusActive)
	}

	return m.validate()
}

// BeforeSave hook to validate data before saving
func (m *ETCMapping) BeforeSave(tx *gorm.DB) error {
	// Update timestamp manually if not using GORM auto-timestamps
	m.UpdatedAt = time.Now()

	return m.validate()
}

// validate performs comprehensive validation of the mapping
func (m *ETCMapping) validate() error {
	// Validate ETC record ID
	if m.ETCRecordID <= 0 {
		return fmt.Errorf("ETC record ID must be positive")
	}

	// Validate mapping type
	if err := m.validateMappingType(); err != nil {
		return err
	}

	// Validate mapped entity ID
	if m.MappedEntityID <= 0 {
		return fmt.Errorf("mapped entity ID must be positive")
	}

	// Validate mapped entity type
	if err := m.validateMappedEntityType(); err != nil {
		return err
	}

	// Validate confidence
	if m.Confidence < 0.0 || m.Confidence > 1.0 {
		return fmt.Errorf("confidence must be between 0.0 and 1.0")
	}

	// Validate status
	if err := m.validateStatus(); err != nil {
		return err
	}

	// Validate created_by if provided
	if m.CreatedBy != "" && len(m.CreatedBy) > 100 {
		return fmt.Errorf("created_by field too long (max 100 characters)")
	}

	return nil
}

// validateMappingType validates the mapping type
func (m *ETCMapping) validateMappingType() error {
	validTypes := []string{
		string(MappingTypeDtako),
		string(MappingTypeExpense),
		string(MappingTypeInvoice),
	}

	mappingType := strings.ToLower(strings.TrimSpace(m.MappingType))
	if mappingType == "" {
		return fmt.Errorf("mapping type cannot be empty")
	}

	for _, validType := range validTypes {
		if mappingType == validType {
			m.MappingType = mappingType // Normalize to lowercase
			return nil
		}
	}

	return fmt.Errorf("invalid mapping type: %s (must be one of: %s)",
		m.MappingType, strings.Join(validTypes, ", "))
}

// validateMappedEntityType validates the mapped entity type
func (m *ETCMapping) validateMappedEntityType() error {
	validTypes := []string{
		string(EntityTypeDtakoRecord),
		string(EntityTypeExpenseRecord),
		string(EntityTypeInvoiceRecord),
	}

	entityType := strings.ToLower(strings.TrimSpace(m.MappedEntityType))
	if entityType == "" {
		return fmt.Errorf("mapped entity type cannot be empty")
	}

	for _, validType := range validTypes {
		if entityType == validType {
			m.MappedEntityType = entityType // Normalize to lowercase
			return nil
		}
	}

	return fmt.Errorf("invalid mapped entity type: %s (must be one of: %s)",
		m.MappedEntityType, strings.Join(validTypes, ", "))
}

// validateStatus validates the mapping status
func (m *ETCMapping) validateStatus() error {
	validStatuses := []string{
		string(MappingStatusActive),
		string(MappingStatusInactive),
		string(MappingStatusPending),
		string(MappingStatusRejected),
	}

	status := strings.ToLower(strings.TrimSpace(m.Status))
	if status == "" {
		m.Status = string(MappingStatusActive) // Default to active
		return nil
	}

	for _, validStatus := range validStatuses {
		if status == validStatus {
			m.Status = status // Normalize to lowercase
			return nil
		}
	}

	return fmt.Errorf("invalid status: %s (must be one of: %s)",
		m.Status, strings.Join(validStatuses, ", "))
}

// IsActive returns true if the mapping is active
func (m *ETCMapping) IsActive() bool {
	return m.Status == string(MappingStatusActive)
}

// IsPending returns true if the mapping is pending approval
func (m *ETCMapping) IsPending() bool {
	return m.Status == string(MappingStatusPending)
}

// CanTransitionTo checks if the mapping can transition to the given status
func (m *ETCMapping) CanTransitionTo(newStatus string) bool {
	currentStatus := MappingStatus(m.Status)
	targetStatus := MappingStatus(newStatus)

	switch currentStatus {
	case MappingStatusPending:
		return targetStatus == MappingStatusActive || targetStatus == MappingStatusRejected
	case MappingStatusActive:
		return targetStatus == MappingStatusInactive
	case MappingStatusInactive:
		return targetStatus == MappingStatusActive
	case MappingStatusRejected:
		return targetStatus == MappingStatusPending
	default:
		return false
	}
}

// Activate sets the mapping status to active if allowed
func (m *ETCMapping) Activate() error {
	if !m.CanTransitionTo(string(MappingStatusActive)) {
		return fmt.Errorf("cannot activate mapping from status: %s", m.Status)
	}
	m.Status = string(MappingStatusActive)
	return nil
}

// Deactivate sets the mapping status to inactive if allowed
func (m *ETCMapping) Deactivate() error {
	if !m.CanTransitionTo(string(MappingStatusInactive)) {
		return fmt.Errorf("cannot deactivate mapping from status: %s", m.Status)
	}
	m.Status = string(MappingStatusInactive)
	return nil
}

// Approve sets the mapping status to active from pending
func (m *ETCMapping) Approve() error {
	if m.Status != string(MappingStatusPending) {
		return fmt.Errorf("can only approve pending mappings, current status: %s", m.Status)
	}
	m.Status = string(MappingStatusActive)
	return nil
}

// Reject sets the mapping status to rejected from pending
func (m *ETCMapping) Reject() error {
	if m.Status != string(MappingStatusPending) {
		return fmt.Errorf("can only reject pending mappings, current status: %s", m.Status)
	}
	m.Status = string(MappingStatusRejected)
	return nil
}

// SetMetadata sets the metadata field with validation
func (m *ETCMapping) SetMetadata(metadata map[string]interface{}) error {
	if metadata == nil {
		m.Metadata = nil
		return nil
	}

	// Convert to JSON
	jsonData, err := json.Marshal(metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	// Check size (limit to reasonable size)
	if len(jsonData) > 65535 { // 64KB limit
		return fmt.Errorf("metadata too large (max 64KB)")
	}

	m.Metadata = datatypes.JSON(jsonData)
	return nil
}

// GetMetadata returns the metadata as a map
func (m *ETCMapping) GetMetadata() (map[string]interface{}, error) {
	if m.Metadata == nil {
		return nil, nil
	}

	var metadata map[string]interface{}
	if err := json.Unmarshal([]byte(m.Metadata), &metadata); err != nil {
		return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
	}

	return metadata, nil
}

// GetConfidencePercentage returns confidence as a percentage
func (m *ETCMapping) GetConfidencePercentage() float64 {
	return float64(m.Confidence * 100)
}

// IsHighConfidence returns true if confidence is above 0.8
func (m *ETCMapping) IsHighConfidence() bool {
	return m.Confidence >= 0.8
}

// IsLowConfidence returns true if confidence is below 0.5
func (m *ETCMapping) IsLowConfidence() bool {
	return m.Confidence < 0.5
}

// Public wrapper methods for tests

// Validate performs comprehensive validation of the mapping (public method)
func (m *ETCMapping) Validate() error {
	return m.validate()
}

// BeforeUpdate prepares the record before updating (public method for testing)
func (m *ETCMapping) BeforeUpdate() error {
	m.UpdatedAt = time.Now()
	return nil
}

// GetTableName returns the table name
func (m *ETCMapping) GetTableName() string {
	return m.TableName()
}

// String returns a string representation of the mapping
func (m *ETCMapping) String() string {
	return fmt.Sprintf("ETCMapping{ID:%d, ETCRecordID:%d, Type:%s, EntityID:%d, Confidence:%.2f, Status:%s}",
		m.ID, m.ETCRecordID, m.MappingType, m.MappedEntityID, m.Confidence, m.Status)
}

// Public validation helper functions

// IsValidMappingType checks if a mapping type is valid
func IsValidMappingType(mappingType string) bool {
	validTypes := []string{
		string(MappingTypeDtako),
		string(MappingTypeExpense),
		string(MappingTypeInvoice),
	}

	for _, validType := range validTypes {
		if strings.ToLower(strings.TrimSpace(mappingType)) == validType {
			return true
		}
	}
	return false
}

// IsValidEntityType checks if an entity type is valid
func IsValidEntityType(entityType string) bool {
	validTypes := []string{
		string(EntityTypeDtakoRecord),
		string(EntityTypeExpenseRecord),
		string(EntityTypeInvoiceRecord),
	}

	for _, validType := range validTypes {
		if strings.ToLower(strings.TrimSpace(entityType)) == validType {
			return true
		}
	}
	return false
}

// IsValidStatus checks if a status is valid
func IsValidStatus(status string) bool {
	validStatuses := []string{
		string(MappingStatusActive),
		string(MappingStatusInactive),
		string(MappingStatusPending),
		string(MappingStatusRejected),
	}

	for _, validStatus := range validStatuses {
		if strings.ToLower(strings.TrimSpace(status)) == validStatus {
			return true
		}
	}
	return false
}
package services

import (
	"crypto/sha256"
	"fmt"
	"time"

	"github.com/google/uuid"
	pb "github.com/yhonda-ohishi/etc_meisai/src/pb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// HooksMigratorService centralizes business logic extracted from GORM hooks
type HooksMigratorService struct {
	validationService *ValidationService
	auditService      *AuditService
}

// NewHooksMigratorService creates a new hooks migrator service
func NewHooksMigratorService(validationService *ValidationService, auditService *AuditService) *HooksMigratorService {
	return &HooksMigratorService{
		validationService: validationService,
		auditService:      auditService,
	}
}

// ETCMeisaiRecordBeforeCreate handles pre-creation logic for ETC Meisai Records
func (h *HooksMigratorService) ETCMeisaiRecordBeforeCreate(record *pb.ETCMeisaiRecord) error {
	// Validate the record
	if err := h.validationService.ValidateETCMeisaiRecord(record); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Generate hash if empty
	if record.Hash == "" {
		record.Hash = h.generateETCMeisaiRecordHash(record)
	}

	// Audit logging
	h.auditService.LogETCMeisaiRecordCreation(record)

	return nil
}

// ETCMeisaiRecordBeforeSave handles pre-save logic for ETC Meisai Records
func (h *HooksMigratorService) ETCMeisaiRecordBeforeSave(record *pb.ETCMeisaiRecord) error {
	// Validate the record
	if err := h.validationService.ValidateETCMeisaiRecord(record); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Audit logging
	h.auditService.LogETCMeisaiRecordUpdate(record)

	return nil
}

// ImportSessionBeforeCreate handles pre-creation logic for Import Sessions
func (h *HooksMigratorService) ImportSessionBeforeCreate(session *pb.ImportSession) error {
	// Generate UUID if not provided
	if session.Id == "" {
		session.Id = uuid.New().String()
	}

	// Set started time if not provided
	if session.StartedAt == nil {
		session.StartedAt = timestamppb.New(time.Now())
	}

	// Set default status if not provided
	if session.Status == pb.ImportStatus_IMPORT_STATUS_UNSPECIFIED {
		session.Status = pb.ImportStatus_IMPORT_STATUS_PENDING
	}

	// Validate the session
	if err := h.validationService.ValidateImportSession(session); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Audit logging
	h.auditService.LogImportSessionCreation(session)

	return nil
}

// ImportSessionBeforeSave handles pre-save logic for Import Sessions
func (h *HooksMigratorService) ImportSessionBeforeSave(session *pb.ImportSession) error {
	// Validate the session
	if err := h.validationService.ValidateImportSession(session); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Audit logging
	h.auditService.LogImportSessionUpdate(session)

	return nil
}

// ETCMappingBeforeCreate handles pre-creation logic for ETC Mappings
func (h *HooksMigratorService) ETCMappingBeforeCreate(mapping *pb.ETCMapping) error {
	// Set default status if not provided
	if mapping.Status == pb.MappingStatus_MAPPING_STATUS_UNSPECIFIED {
		mapping.Status = pb.MappingStatus_MAPPING_STATUS_PENDING
	}

	// Set timestamps
	if mapping.CreatedAt == nil {
		mapping.CreatedAt = timestamppb.New(time.Now())
	}

	// Validate the mapping
	if err := h.validationService.ValidateETCMapping(mapping); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Audit logging
	h.auditService.LogETCMappingCreation(mapping)

	return nil
}

// ETCMappingBeforeSave handles pre-save logic for ETC Mappings
func (h *HooksMigratorService) ETCMappingBeforeSave(mapping *pb.ETCMapping) error {
	// Update timestamp
	if mapping.UpdatedAt == nil {
		mapping.UpdatedAt = timestamppb.New(time.Now())
	}

	// Validate the mapping
	if err := h.validationService.ValidateETCMapping(mapping); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Audit logging
	h.auditService.LogETCMappingUpdate(mapping)

	return nil
}

// generateETCMeisaiRecordHash generates a SHA256 hash for ETC Meisai Record
func (h *HooksMigratorService) generateETCMeisaiRecordHash(record *pb.ETCMeisaiRecord) string {
	// Parse date from string format
	date := record.Date // Already in YYYY-MM-DD format from proto

	data := fmt.Sprintf("%s|%s|%s|%s|%d|%s|%s",
		date,
		record.Time,
		record.EntranceIc,
		record.ExitIc,
		record.TollAmount,
		record.CarNumber,
		record.EtcCardNumber,
	)

	hash := sha256.Sum256([]byte(data))
	return fmt.Sprintf("%x", hash)
}

// BeforeCreateHooks centralizes all BeforeCreate hook logic
func (h *HooksMigratorService) BeforeCreateHooks(entityType string, entity interface{}) error {
	switch entityType {
	case "ETCMeisaiRecord":
		if record, ok := entity.(*pb.ETCMeisaiRecord); ok {
			return h.ETCMeisaiRecordBeforeCreate(record)
		}
	case "ImportSession":
		if session, ok := entity.(*pb.ImportSession); ok {
			return h.ImportSessionBeforeCreate(session)
		}
	case "ETCMapping":
		if mapping, ok := entity.(*pb.ETCMapping); ok {
			return h.ETCMappingBeforeCreate(mapping)
		}
	}
	return fmt.Errorf("unsupported entity type: %s", entityType)
}

// BeforeSaveHooks centralizes all BeforeSave hook logic
func (h *HooksMigratorService) BeforeSaveHooks(entityType string, entity interface{}) error {
	switch entityType {
	case "ETCMeisaiRecord":
		if record, ok := entity.(*pb.ETCMeisaiRecord); ok {
			return h.ETCMeisaiRecordBeforeSave(record)
		}
	case "ImportSession":
		if session, ok := entity.(*pb.ImportSession); ok {
			return h.ImportSessionBeforeSave(session)
		}
	case "ETCMapping":
		if mapping, ok := entity.(*pb.ETCMapping); ok {
			return h.ETCMappingBeforeSave(mapping)
		}
	}
	return fmt.Errorf("unsupported entity type: %s", entityType)
}

// GetSupportedEntityTypes returns list of entities with migrated hook logic
func (h *HooksMigratorService) GetSupportedEntityTypes() []string {
	return []string{
		"ETCMeisaiRecord",
		"ImportSession",
		"ETCMapping",
	}
}

// HookExecutionResult represents the result of hook execution
type HookExecutionResult struct {
	Success      bool
	ErrorMessage string
	HookType     string
	EntityType   string
	ExecutedAt   time.Time
}

// ExecuteBeforeCreateHook executes before create hook with result tracking
func (h *HooksMigratorService) ExecuteBeforeCreateHook(entityType string, entity interface{}) *HookExecutionResult {
	result := &HookExecutionResult{
		HookType:   "BeforeCreate",
		EntityType: entityType,
		ExecutedAt: time.Now(),
	}

	if err := h.BeforeCreateHooks(entityType, entity); err != nil {
		result.Success = false
		result.ErrorMessage = err.Error()
	} else {
		result.Success = true
	}

	return result
}

// ExecuteBeforeSaveHook executes before save hook with result tracking
func (h *HooksMigratorService) ExecuteBeforeSaveHook(entityType string, entity interface{}) *HookExecutionResult {
	result := &HookExecutionResult{
		HookType:   "BeforeSave",
		EntityType: entityType,
		ExecutedAt: time.Now(),
	}

	if err := h.BeforeSaveHooks(entityType, entity); err != nil {
		result.Success = false
		result.ErrorMessage = err.Error()
	} else {
		result.Success = true
	}

	return result
}
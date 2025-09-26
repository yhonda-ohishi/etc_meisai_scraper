package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	pb "github.com/yhonda-ohishi/etc_meisai/src/pb"
)

// AuditService centralizes audit logging functionality extracted from GORM hooks
type AuditService struct {
	logger AuditLogger
}

// AuditLogger defines the interface for audit logging
type AuditLogger interface {
	LogAuditEvent(ctx context.Context, event *AuditEvent) error
}

// AuditEvent represents an audit log entry
type AuditEvent struct {
	ID          string                 `json:"id"`
	EventType   string                 `json:"event_type"`
	EntityType  string                 `json:"entity_type"`
	EntityID    string                 `json:"entity_id"`
	UserID      string                 `json:"user_id,omitempty"`
	Timestamp   time.Time              `json:"timestamp"`
	Changes     map[string]interface{} `json:"changes,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	IPAddress   string                 `json:"ip_address,omitempty"`
	UserAgent   string                 `json:"user_agent,omitempty"`
	SessionID   string                 `json:"session_id,omitempty"`
}

// DefaultAuditLogger provides a simple audit logger implementation
type DefaultAuditLogger struct {
	logFunc func(event *AuditEvent)
}

// LogAuditEvent logs an audit event using the default logger
func (d *DefaultAuditLogger) LogAuditEvent(ctx context.Context, event *AuditEvent) error {
	if d.logFunc != nil {
		d.logFunc(event)
	}
	return nil
}

// NewAuditService creates a new audit service
func NewAuditService(logger AuditLogger) *AuditService {
	if logger == nil {
		// Provide default logger that writes to stdout
		logger = &DefaultAuditLogger{
			logFunc: func(event *AuditEvent) {
				if data, err := json.Marshal(event); err == nil {
					fmt.Printf("[AUDIT] %s\n", string(data))
				}
			},
		}
	}
	return &AuditService{logger: logger}
}

// LogETCMeisaiRecordCreation logs the creation of an ETC Meisai Record
func (a *AuditService) LogETCMeisaiRecordCreation(record *pb.ETCMeisaiRecord) {
	event := &AuditEvent{
		ID:         fmt.Sprintf("etc_record_%d_create_%d", record.Id, time.Now().UnixNano()),
		EventType:  "CREATE",
		EntityType: "ETCMeisaiRecord",
		EntityID:   fmt.Sprintf("%d", record.Id),
		Timestamp:  time.Now(),
		Metadata: map[string]interface{}{
			"hash":             record.Hash,
			"date":             record.Date,
			"entrance_ic":      record.EntranceIc,
			"exit_ic":          record.ExitIc,
			"toll_amount":      record.TollAmount,
			"car_number":       record.CarNumber,
			"etc_card_number":  record.EtcCardNumber,
		},
	}

	// Add optional fields if present
	if record.EtcNum != nil {
		event.Metadata["etc_num"] = *record.EtcNum
	}
	if record.DtakoRowId != nil {
		event.Metadata["dtako_row_id"] = *record.DtakoRowId
	}

	a.logger.LogAuditEvent(context.Background(), event)
}

// LogETCMeisaiRecordUpdate logs the update of an ETC Meisai Record
func (a *AuditService) LogETCMeisaiRecordUpdate(record *pb.ETCMeisaiRecord) {
	event := &AuditEvent{
		ID:         fmt.Sprintf("etc_record_%d_update_%d", record.Id, time.Now().UnixNano()),
		EventType:  "UPDATE",
		EntityType: "ETCMeisaiRecord",
		EntityID:   fmt.Sprintf("%d", record.Id),
		Timestamp:  time.Now(),
		Metadata: map[string]interface{}{
			"hash":        record.Hash,
			"date":        record.Date,
			"toll_amount": record.TollAmount,
		},
	}

	a.logger.LogAuditEvent(context.Background(), event)
}

// LogImportSessionCreation logs the creation of an Import Session
func (a *AuditService) LogImportSessionCreation(session *pb.ImportSession) {
	event := &AuditEvent{
		ID:         fmt.Sprintf("import_session_%s_create_%d", session.Id, time.Now().UnixNano()),
		EventType:  "CREATE",
		EntityType: "ImportSession",
		EntityID:   session.Id,
		UserID:     session.CreatedBy,
		Timestamp:  time.Now(),
		Metadata: map[string]interface{}{
			"account_type": session.AccountType,
			"account_id":   session.AccountId,
			"file_name":    session.FileName,
			"file_size":    session.FileSize,
			"status":       session.Status.String(),
		},
	}

	a.logger.LogAuditEvent(context.Background(), event)
}

// LogImportSessionUpdate logs the update of an Import Session
func (a *AuditService) LogImportSessionUpdate(session *pb.ImportSession) {
	event := &AuditEvent{
		ID:         fmt.Sprintf("import_session_%s_update_%d", session.Id, time.Now().UnixNano()),
		EventType:  "UPDATE",
		EntityType: "ImportSession",
		EntityID:   session.Id,
		UserID:     session.CreatedBy,
		Timestamp:  time.Now(),
		Metadata: map[string]interface{}{
			"status":          session.Status.String(),
			"total_rows":      session.TotalRows,
			"processed_rows":  session.ProcessedRows,
			"success_rows":    session.SuccessRows,
			"error_rows":      session.ErrorRows,
			"duplicate_rows":  session.DuplicateRows,
		},
	}

	// Include completion timestamp if available
	if session.CompletedAt != nil {
		event.Metadata["completed_at"] = session.CompletedAt.AsTime()
	}

	a.logger.LogAuditEvent(context.Background(), event)
}

// LogETCMappingCreation logs the creation of an ETC Mapping
func (a *AuditService) LogETCMappingCreation(mapping *pb.ETCMapping) {
	event := &AuditEvent{
		ID:         fmt.Sprintf("etc_mapping_%d_create_%d", mapping.Id, time.Now().UnixNano()),
		EventType:  "CREATE",
		EntityType: "ETCMapping",
		EntityID:   fmt.Sprintf("%d", mapping.Id),
		UserID:     mapping.CreatedBy,
		Timestamp:  time.Now(),
		Metadata: map[string]interface{}{
			"etc_record_id":      mapping.EtcRecordId,
			"mapping_type":       mapping.MappingType,
			"mapped_entity_id":   mapping.MappedEntityId,
			"mapped_entity_type": mapping.MappedEntityType,
			"confidence":         mapping.Confidence,
			"status":             mapping.Status.String(),
		},
	}

	a.logger.LogAuditEvent(context.Background(), event)
}

// LogETCMappingUpdate logs the update of an ETC Mapping
func (a *AuditService) LogETCMappingUpdate(mapping *pb.ETCMapping) {
	event := &AuditEvent{
		ID:         fmt.Sprintf("etc_mapping_%d_update_%d", mapping.Id, time.Now().UnixNano()),
		EventType:  "UPDATE",
		EntityType: "ETCMapping",
		EntityID:   fmt.Sprintf("%d", mapping.Id),
		UserID:     mapping.CreatedBy,
		Timestamp:  time.Now(),
		Metadata: map[string]interface{}{
			"confidence": mapping.Confidence,
			"status":     mapping.Status.String(),
		},
	}

	a.logger.LogAuditEvent(context.Background(), event)
}

// LogEntityDeletion logs the deletion of any entity
func (a *AuditService) LogEntityDeletion(entityType, entityID, userID string, metadata map[string]interface{}) {
	event := &AuditEvent{
		ID:         fmt.Sprintf("%s_%s_delete_%d", entityType, entityID, time.Now().UnixNano()),
		EventType:  "DELETE",
		EntityType: entityType,
		EntityID:   entityID,
		UserID:     userID,
		Timestamp:  time.Now(),
		Metadata:   metadata,
	}

	a.logger.LogAuditEvent(context.Background(), event)
}

// LogBulkOperation logs bulk operations
func (a *AuditService) LogBulkOperation(operationType, entityType string, recordCount int, userID string, metadata map[string]interface{}) {
	event := &AuditEvent{
		ID:         fmt.Sprintf("bulk_%s_%s_%d", operationType, entityType, time.Now().UnixNano()),
		EventType:  fmt.Sprintf("BULK_%s", operationType),
		EntityType: entityType,
		EntityID:   "bulk",
		UserID:     userID,
		Timestamp:  time.Now(),
		Metadata: map[string]interface{}{
			"record_count": recordCount,
			"operation":    operationType,
		},
	}

	// Merge additional metadata
	if metadata != nil {
		for k, v := range metadata {
			event.Metadata[k] = v
		}
	}

	a.logger.LogAuditEvent(context.Background(), event)
}

// LogValidationFailure logs validation failures for audit purposes
func (a *AuditService) LogValidationFailure(entityType, entityID string, validationError error, metadata map[string]interface{}) {
	event := &AuditEvent{
		ID:         fmt.Sprintf("validation_failure_%s_%s_%d", entityType, entityID, time.Now().UnixNano()),
		EventType:  "VALIDATION_FAILURE",
		EntityType: entityType,
		EntityID:   entityID,
		Timestamp:  time.Now(),
		Metadata: map[string]interface{}{
			"error_message": validationError.Error(),
		},
	}

	// Merge additional metadata
	if metadata != nil {
		for k, v := range metadata {
			event.Metadata[k] = v
		}
	}

	a.logger.LogAuditEvent(context.Background(), event)
}

// LogBusinessRuleViolation logs business rule violations
func (a *AuditService) LogBusinessRuleViolation(entityType, entityID, rule string, details map[string]interface{}) {
	event := &AuditEvent{
		ID:         fmt.Sprintf("rule_violation_%s_%s_%d", entityType, entityID, time.Now().UnixNano()),
		EventType:  "BUSINESS_RULE_VIOLATION",
		EntityType: entityType,
		EntityID:   entityID,
		Timestamp:  time.Now(),
		Metadata: map[string]interface{}{
			"rule":    rule,
			"details": details,
		},
	}

	a.logger.LogAuditEvent(context.Background(), event)
}

// LogSecurityEvent logs security-related events
func (a *AuditService) LogSecurityEvent(eventType, description, userID, ipAddress string, metadata map[string]interface{}) {
	event := &AuditEvent{
		ID:         fmt.Sprintf("security_%s_%d", eventType, time.Now().UnixNano()),
		EventType:  fmt.Sprintf("SECURITY_%s", eventType),
		EntityType: "SECURITY",
		EntityID:   "security_event",
		UserID:     userID,
		Timestamp:  time.Now(),
		IPAddress:  ipAddress,
		Metadata: map[string]interface{}{
			"description": description,
		},
	}

	// Merge additional metadata
	if metadata != nil {
		for k, v := range metadata {
			event.Metadata[k] = v
		}
	}

	a.logger.LogAuditEvent(context.Background(), event)
}

// AuditContext provides context information for audit logging
type AuditContext struct {
	UserID    string
	SessionID string
	IPAddress string
	UserAgent string
}

// SetAuditContext updates the audit service with context information
func (a *AuditService) SetAuditContext(ctx *AuditContext) {
	// This could be used to enrich audit events with context information
	// Implementation depends on how context is managed in the application
}

// GetAuditStatistics returns audit statistics
type AuditStatistics struct {
	EventCounts    map[string]int `json:"event_counts"`
	EntityCounts   map[string]int `json:"entity_counts"`
	RecentEvents   int            `json:"recent_events"`
	LastEventTime  time.Time      `json:"last_event_time"`
}

// GetStatistics returns audit statistics (placeholder for future implementation)
func (a *AuditService) GetStatistics(since time.Time) (*AuditStatistics, error) {
	// This would typically query the audit log storage to generate statistics
	return &AuditStatistics{
		EventCounts:   make(map[string]int),
		EntityCounts:  make(map[string]int),
		RecentEvents:  0,
		LastEventTime: time.Now(),
	}, nil
}
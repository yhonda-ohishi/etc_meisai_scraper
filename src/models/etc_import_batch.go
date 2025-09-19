package models

import (
	"fmt"
	"time"
)

// ETCImportBatch represents a batch import operation
type ETCImportBatch struct {
	ID              int64      `json:"id"`
	BatchHash       string     `json:"batch_hash"`
	FileName        string     `json:"file_name"`
	FileSize        int64      `json:"file_size"`
	AccountID       string     `json:"account_id,omitempty"`
	ImportType      string     `json:"import_type,omitempty"`
	TotalRows       int64      `json:"total_rows"`
	ProcessedRows   int64      `json:"processed_rows"`
	SuccessCount    int64      `json:"success_count"`
	TotalRecords    int32      `json:"total_records"`
	ProcessedCount  int32      `json:"processed_count"`
	CreatedCount    int32      `json:"created_count"`
	DuplicateCount  int32      `json:"duplicate_count"`
	ErrorCount      int64      `json:"error_count"`
	Status          string     `json:"status"` // pending, processing, completed, failed
	StartTime       *time.Time `json:"start_time,omitempty"`
	CompleteTime    *time.Time `json:"complete_time,omitempty"`
	CompletedAt     *time.Time `json:"completed_at,omitempty"`
	ErrorMessage    string     `json:"error_message,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
	CreatedBy       string     `json:"created_by,omitempty"`

	// リレーション
	ImportRecords []ETCImportRecord `json:"import_records,omitempty"`
	ImportErrors  []ETCImportError  `json:"import_errors,omitempty"`
}

// BeforeCreate prepares the batch record before creation
func (b *ETCImportBatch) BeforeCreate() error {
	if err := b.Validate(); err != nil {
		return err
	}
	return nil
}

// BeforeUpdate prepares the batch record before updating
func (b *ETCImportBatch) BeforeUpdate() error {
	if err := b.Validate(); err != nil {
		return err
	}
	return nil
}

// Validate checks the batch record for business rule compliance
func (b *ETCImportBatch) Validate() error {
	if b.FileName == "" {
		return fmt.Errorf("FileName is required")
	}

	if b.TotalRecords < 0 {
		return fmt.Errorf("TotalRecords cannot be negative")
	}

	validStatuses := map[string]bool{
		"pending":    true,
		"processing": true,
		"completed":  true,
		"failed":     true,
		"cancelled":  true,
	}

	if !validStatuses[b.Status] {
		return fmt.Errorf("invalid Status: %s", b.Status)
	}

	return nil
}

// GetProgress returns the progress percentage of the batch
func (b *ETCImportBatch) GetProgress() float32 {
	if b.TotalRecords == 0 {
		return 0
	}
	return float32(b.ProcessedCount) / float32(b.TotalRecords) * 100
}

// IsCompleted returns true if the batch is completed
func (b *ETCImportBatch) IsCompleted() bool {
	return b.Status == "completed" || b.Status == "failed" || b.Status == "cancelled"
}

// GetDuration returns the duration of the batch processing
func (b *ETCImportBatch) GetDuration() *time.Duration {
	if b.StartTime == nil {
		return nil
	}

	endTime := time.Now()
	if b.CompleteTime != nil {
		endTime = *b.CompleteTime
	}

	duration := endTime.Sub(*b.StartTime)
	return &duration
}

// ETCImportRecord represents an individual record in an import batch
type ETCImportRecord struct {
	ID          int64     `json:"id"`
	BatchID     int64     `json:"batch_id"`
	ETCMeisaiID *int64    `json:"etc_meisai_id,omitempty"`
	RecordHash  string    `json:"record_hash"`
	Status      string    `json:"status"` // created, duplicate, error
	ErrorMessage string   `json:"error_message,omitempty"`
	CreatedAt   time.Time `json:"created_at"`

	// リレーション
	Batch     *ETCImportBatch `json:"batch,omitempty"`
	ETCMeisai *ETCMeisai      `json:"etc_meisai,omitempty"`
}

// ETCImportError represents an error that occurred during import
type ETCImportError struct {
	ID           int64     `json:"id"`
	BatchID      int64     `json:"batch_id"`
	RowNumber    int32     `json:"row_number"`
	ErrorType    string    `json:"error_type"`
	ErrorMessage string    `json:"error_message"`
	RawData      string    `json:"raw_data,omitempty"`
	CreatedAt    time.Time `json:"created_at"`

	// リレーション
	Batch *ETCImportBatch `json:"batch,omitempty"`
}

// ImportProgressResponse represents the response for import progress
type ImportProgressResponse struct {
	Batch             *ETCImportBatch `json:"batch"`
	ProgressPercentage float32         `json:"progress_percentage"`
	CurrentStatus     string          `json:"current_status"`
	RecentErrors      []string        `json:"recent_errors,omitempty"`
	EstimatedTimeLeft *time.Duration  `json:"estimated_time_left,omitempty"`
}

// ImportSummary provides a summary of the import operation
type ImportSummary struct {
	BatchID         int64         `json:"batch_id"`
	FileName        string        `json:"file_name"`
	TotalRecords    int32         `json:"total_records"`
	ProcessedCount  int32         `json:"processed_count"`
	CreatedCount    int32         `json:"created_count"`
	DuplicateCount  int32         `json:"duplicate_count"`
	ErrorCount      int32         `json:"error_count"`
	Status          string        `json:"status"`
	Duration        *time.Duration `json:"duration,omitempty"`
	ThroughputRPS   float32       `json:"throughput_rps,omitempty"` // Records per second
}
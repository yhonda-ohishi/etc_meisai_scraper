package models

import "time"

// ImportProgress represents the progress of an import operation
type ImportProgress struct {
	BatchID       int64     `json:"batch_id"`
	Status        string    `json:"status"`
	TotalRows     int64     `json:"total_rows"`
	ProcessedRows int64     `json:"processed_rows"`
	SuccessCount  int64     `json:"success_count"`
	ErrorCount    int64     `json:"error_count"`
	Percentage    float32   `json:"percentage"`
	Message       string    `json:"message,omitempty"`
	UpdatedAt     time.Time `json:"updated_at"`
}
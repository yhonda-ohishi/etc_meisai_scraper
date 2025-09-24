package models

import "time"

// ETCImportRequest represents a request to import ETC data
type ETCImportRequest struct {
	FromDate string `json:"from_date" binding:"required"`
	ToDate   string `json:"to_date" binding:"required"`
	Source   string `json:"source,omitempty"`
	BatchID  string `json:"batch_id,omitempty"`
}

// ETCImportResult represents the result of an ETC import operation
type ETCImportResult struct {
	Success      bool      `json:"success"`
	RecordCount  int       `json:"record_count"`
	RecordsRead  int       `json:"records_read"`   // Total records read from source
	RecordsSaved int       `json:"records_saved"`  // Records successfully saved
	ImportedRows int       `json:"imported_rows"`
	Duration     int64     `json:"duration_ms"`
	Message      string    `json:"message"`
	ErrorMessage string    `json:"error_message,omitempty"`
	Errors       []string  `json:"errors,omitempty"`
	ImportedAt   time.Time `json:"imported_at"`
}

// ErrorResponse represents a standard error response
type ErrorResponse struct {
	Code  string `json:"code"`
	Error string `json:"error"`
}
package models

import "time"

// BulkMappingResult represents the result of bulk mapping operations
type BulkMappingResult struct {
	Success       bool                   `json:"success"`
	TotalCount    int                    `json:"total_count"`
	SuccessCount  int                    `json:"success_count"`
	FailureCount  int                    `json:"failure_count"`
	Errors        []BulkOperationError   `json:"errors,omitempty"`
	Mappings      []*ETCMeisaiMapping    `json:"mappings,omitempty"`
	Duration      time.Duration          `json:"duration"`
	ProcessedAt   time.Time              `json:"processed_at"`
}

// BulkOperationError represents an error that occurred during bulk operations
type BulkOperationError struct {
	Index   int    `json:"index"`
	ID      int64  `json:"id,omitempty"`
	Message string `json:"message"`
	Code    string `json:"code,omitempty"`
}

// ImportValidationResult represents the result of CSV import validation
type ImportValidationResult struct {
	Valid       bool                      `json:"valid"`
	TotalRows   int                       `json:"total_rows"`
	ValidRows   int                       `json:"valid_rows"`
	InvalidRows int                       `json:"invalid_rows"`
	Errors      []ImportValidationError   `json:"errors,omitempty"`
	Warnings    []ImportValidationWarning `json:"warnings,omitempty"`
	Duration    time.Duration             `json:"duration"`
}

// ImportValidationError represents a validation error in CSV import
type ImportValidationError struct {
	Row     int    `json:"row"`
	Column  string `json:"column,omitempty"`
	Message string `json:"message"`
	Value   string `json:"value,omitempty"`
}

// ImportValidationWarning represents a validation warning in CSV import
type ImportValidationWarning struct {
	Row     int    `json:"row"`
	Column  string `json:"column,omitempty"`
	Message string `json:"message"`
	Value   string `json:"value,omitempty"`
}

// ImportPreviewResult represents a preview of CSV import data
type ImportPreviewResult struct {
	Headers       []string           `json:"headers"`
	SampleRows    [][]string         `json:"sample_rows"`
	TotalRows     int                `json:"total_rows"`
	PreviewRows   int                `json:"preview_rows"`
	DetectedType  string             `json:"detected_type"`
	Encoding      string             `json:"encoding"`
	Delimiter     string             `json:"delimiter"`
	ValidFields   []string           `json:"valid_fields"`
	InvalidFields []string           `json:"invalid_fields"`
	Warnings      []string           `json:"warnings,omitempty"`
}


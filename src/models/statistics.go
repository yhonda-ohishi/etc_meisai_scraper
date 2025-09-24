package models

import "time"

// Statistics represents aggregated statistics data
type Statistics struct {
	TotalRecords      int64     `json:"total_records"`
	TotalAmount       int64     `json:"total_amount"`
	UniqueVehicles    int       `json:"unique_vehicles"`
	UniqueTollGates   int       `json:"unique_toll_gates"`
	AverageAmount     float64   `json:"average_amount"`
	MaxAmount         int       `json:"max_amount"`
	MinAmount         int       `json:"min_amount"`
	DateRangeStart    time.Time `json:"date_range_start"`
	DateRangeEnd      time.Time `json:"date_range_end"`
	LastUpdated       time.Time `json:"last_updated"`
	ProcessingStatus  string    `json:"processing_status"`
	ErrorCount        int       `json:"error_count"`
	DuplicateCount    int       `json:"duplicate_count"`
	MappedRecords     int       `json:"mapped_records"`
	UnmappedRecords   int       `json:"unmapped_records"`
}

// NewStatistics creates a new statistics instance
func NewStatistics() *Statistics {
	return &Statistics{
		ProcessingStatus: "idle",
		LastUpdated:      time.Now(),
	}
}
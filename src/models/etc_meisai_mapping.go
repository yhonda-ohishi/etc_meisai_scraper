package models

import (
	"fmt"
	"time"
)

// ETCMeisaiMapping represents the mapping between ETC records and DTako records
type ETCMeisaiMapping struct {
	ID          int64     `json:"id"`
	ETCMeisaiID int64     `json:"etc_meisai_id"`
	DTakoRowID  string    `json:"dtako_row_id"`
	MappingType string    `json:"mapping_type"` // auto, manual
	Confidence  float32   `json:"confidence"`
	Notes       string    `json:"notes,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	CreatedBy   string    `json:"created_by,omitempty"`

	// リレーション
	ETCMeisai *ETCMeisai `json:"etc_meisai,omitempty"`
}

// BeforeCreate prepares the mapping record before creation
func (m *ETCMeisaiMapping) BeforeCreate() error {
	if err := m.Validate(); err != nil {
		return err
	}
	return nil
}

// BeforeUpdate prepares the mapping record before updating
func (m *ETCMeisaiMapping) BeforeUpdate() error {
	if err := m.Validate(); err != nil {
		return err
	}
	return nil
}

// Validate checks the mapping record for business rule compliance
func (m *ETCMeisaiMapping) Validate() error {
	if m.ETCMeisaiID <= 0 {
		return fmt.Errorf("ETCMeisaiID must be positive")
	}

	if m.DTakoRowID == "" {
		return fmt.Errorf("DTakoRowID is required")
	}

	if m.MappingType != "auto" && m.MappingType != "manual" {
		return fmt.Errorf("MappingType must be 'auto' or 'manual'")
	}

	if m.Confidence < 0 || m.Confidence > 1 {
		return fmt.Errorf("Confidence must be between 0 and 1")
	}

	return nil
}

// IsHighConfidence returns true if the mapping has high confidence
func (m *ETCMeisaiMapping) IsHighConfidence() bool {
	return m.Confidence >= 0.8
}

// MappingListParams defines parameters for querying mapping records
type MappingListParams struct {
	Limit         int      `json:"limit"`
	Offset        int      `json:"offset"`
	ETCMeisaiID   *int64   `json:"etc_meisai_id,omitempty"`
	DTakoRowID    string   `json:"dtako_row_id,omitempty"`
	MappingType   string   `json:"mapping_type,omitempty"`
	MinConfidence *float32 `json:"min_confidence,omitempty"`
	CreatedBy     string   `json:"created_by,omitempty"`
}

// PotentialMatch represents a potential mapping match
type PotentialMatch struct {
	DTakoRowID   string                 `json:"dtako_row_id"`
	Confidence   float32                `json:"confidence"`
	MatchReasons []string               `json:"match_reasons"`
	DTakoData    map[string]interface{} `json:"dtako_data"`
}

// MappingStats provides statistics about mappings
type MappingStats struct {
	TotalMappings   int64   `json:"total_mappings"`
	AutoMappings    int64   `json:"auto_mappings"`
	ManualMappings  int64   `json:"manual_mappings"`
	HighConfidence  int64   `json:"high_confidence"`
	LowConfidence   int64   `json:"low_confidence"`
	AverageConfidence float32 `json:"average_confidence"`
	UnmappedRecords int64   `json:"unmapped_records"`
}

// AutoMatchResult represents the result of an auto-matching operation
type AutoMatchResult struct {
	ETCMeisaiID      int64             `json:"etc_meisai_id"`
	PotentialMatches []*PotentialMatch `json:"potential_matches,omitempty"`
	BestMatch        *PotentialMatch   `json:"best_match,omitempty"`
	Error            string            `json:"error,omitempty"`
}
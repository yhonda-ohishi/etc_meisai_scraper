package models

import "time"

// MatchCandidate represents a potential match for mapping
type MatchCandidate struct {
	ID           int64     `json:"id"`
	ETCMeisaiID  int64     `json:"etc_meisai_id"`
	DTakoRowID   string    `json:"dtako_row_id"`
	Confidence   float32   `json:"confidence"`
	MatchReason  string    `json:"match_reason"`
	MatchDetails map[string]interface{} `json:"match_details,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
}
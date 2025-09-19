package parser

import "github.com/yhonda-ohishi/etc_meisai/src/models"

// ParseResult represents the result of parsing a CSV file
type ParseResult struct {
	Records        []*models.ETCMeisai `json:"records"`
	TotalRows      int                 `json:"total_rows"`
	ValidRows      int                 `json:"valid_rows"`
	ErrorRows      int                 `json:"error_rows"`
	DuplicateCount int                 `json:"duplicate_count"`
	Errors         []ParseError        `json:"errors,omitempty"`
}

// ParseError represents an error that occurred during parsing
type ParseError struct {
	Row     int    `json:"row"`
	Column  string `json:"column,omitempty"`
	Message string `json:"message"`
	Value   string `json:"value,omitempty"`
}
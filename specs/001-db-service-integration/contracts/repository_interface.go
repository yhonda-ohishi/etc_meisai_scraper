// Package contracts defines the repository interfaces for database service integration
package contracts

import (
	"time"

	"github.com/yhonda-ohishi/etc_meisai/src/models"
)

// Type aliases for models to avoid import issues
type ETCMeisai = models.ETCMeisai
type ETCMeisaiMapping = models.ETCMeisaiMapping
type ETCImportBatch = models.ETCImportBatch
type ETCImportRecord = models.ETCImportRecord
type ETCImportError = models.ETCImportError

// ETCRepository defines the unified repository interface for ETC明細 data access
type ETCRepository interface {
	// Basic CRUD operations (db_service compatible)
	Create(data *ETCMeisai) error
	GetByID(id int64) (*ETCMeisai, error)
	Update(data *ETCMeisai) error
	DeleteByID(id int64) error
	List(params *ETCListParams) ([]*ETCMeisai, int64, error)

	// ETC-specific bulk operations
	BulkInsert(records []*ETCMeisai) error
	BulkUpdate(records []*ETCMeisai) error

	// Query operations
	GetByDateRange(start, end time.Time) ([]*ETCMeisai, error)
	GetByHash(hash string) (*ETCMeisai, error)
	GetByETCNumber(etcNumber string) ([]*ETCMeisai, error)
	GetUnmappedRecords(start, end time.Time) ([]*ETCMeisai, error)

	// Hash-based operations for duplicate detection
	ListByHashBatch(hashes []string) ([]*ETCMeisai, error)
	CheckDuplicatesByHash(hashes []string) (map[string]bool, error)
	GenerateHash(data *ETCMeisai) string

	// Aggregation operations
	GetSummaryByDateRange(start, end time.Time) (*ETCSummary, error)
	GetMonthlyStats(year int, month int) (*ETCMonthlyStats, error)
	CountByDateRange(start, end time.Time) (int64, error)

	// Transaction support
	WithTransaction(fn func(repo ETCRepository) error) error
}

// ETCMappingRepository defines the interface for ETC-DTako mapping operations
type ETCMappingRepository interface {
	// Basic CRUD
	Create(mapping *ETCMeisaiMapping) error
	GetByID(id int64) (*ETCMeisaiMapping, error)
	Update(mapping *ETCMeisaiMapping) error
	DeleteByID(id int64) error

	// Query operations
	GetByETCMeisaiID(etcMeisaiID int64) ([]*ETCMeisaiMapping, error)
	GetByDTakoRowID(dtakoRowID string) (*ETCMeisaiMapping, error)
	List(params *MappingListParams) ([]*ETCMeisaiMapping, int64, error)

	// Batch operations
	BulkCreateMappings(mappings []*ETCMeisaiMapping) error
	DeleteByETCMeisaiID(etcMeisaiID int64) error

	// Auto-matching support
	FindPotentialMatches(etcMeisaiID int64, threshold float32) ([]*PotentialMatch, error)
	UpdateConfidenceScore(id int64, confidence float32) error
}

// ETCImportRepository defines the interface for batch import operations
type ETCImportRepository interface {
	// Batch management
	CreateBatch(batch *ETCImportBatch) error
	GetBatchByID(id int64) (*ETCImportBatch, error)
	GetBatchByHash(hash string) (*ETCImportBatch, error)
	UpdateBatchStatus(id int64, status string) error
	UpdateBatchProgress(id int64, processed, errors int32) error

	// Import operations
	AddImportRecord(batchID int64, record *ETCImportRecord) error
	GetImportRecords(batchID int64) ([]*ETCImportRecord, error)
	GetImportErrors(batchID int64) ([]*ETCImportError, error)

	// Cleanup operations
	DeleteCompletedBatches(olderThan time.Time) error
	GetActiveBatches() ([]*ETCImportBatch, error)
}

// Parameter types for repository operations
type ETCListParams struct {
	Limit      int       `json:"limit"`
	Offset     int       `json:"offset"`
	StartDate  *time.Time `json:"start_date,omitempty"`
	EndDate    *time.Time `json:"end_date,omitempty"`
	ETCNumber  string    `json:"etc_number,omitempty"`
	EntryIC    string    `json:"entry_ic,omitempty"`
	ExitIC     string    `json:"exit_ic,omitempty"`
	SortBy     string    `json:"sort_by,omitempty"`
	SortOrder  string    `json:"sort_order,omitempty"`
}

type MappingListParams struct {
	Limit       int     `json:"limit"`
	Offset      int     `json:"offset"`
	ETCMeisaiID *int64  `json:"etc_meisai_id,omitempty"`
	DTakoRowID  string  `json:"dtako_row_id,omitempty"`
	MappingType string  `json:"mapping_type,omitempty"`
	MinConfidence *float32 `json:"min_confidence,omitempty"`
}

// Result types
type ETCSummary struct {
	TotalAmount    int64     `json:"total_amount"`
	TotalCount     int64     `json:"total_count"`
	StartDate      time.Time `json:"start_date"`
	EndDate        time.Time `json:"end_date"`
	ByETCNumber    map[string]*ETCNumberSummary `json:"by_etc_number"`
	ByMonth        map[string]*ETCMonthlySummary `json:"by_month"`
}

type ETCNumberSummary struct {
	ETCNumber   string `json:"etc_number"`
	TotalAmount int64  `json:"total_amount"`
	TotalCount  int64  `json:"total_count"`
}

type ETCMonthlySummary struct {
	Year        int   `json:"year"`
	Month       int   `json:"month"`
	TotalAmount int64 `json:"total_amount"`
	TotalCount  int64 `json:"total_count"`
}

type ETCMonthlyStats struct {
	Year           int                    `json:"year"`
	Month          int                    `json:"month"`
	TotalAmount    int64                  `json:"total_amount"`
	TotalCount     int64                  `json:"total_count"`
	DailyBreakdown map[int]*ETCDailySummary `json:"daily_breakdown"`
	TopRoutes      []*ETCRouteSummary     `json:"top_routes"`
}

type ETCDailySummary struct {
	Day         int   `json:"day"`
	Amount      int64 `json:"amount"`
	Count       int64 `json:"count"`
}

type ETCRouteSummary struct {
	EntryIC     string `json:"entry_ic"`
	ExitIC      string `json:"exit_ic"`
	Count       int64  `json:"count"`
	TotalAmount int64  `json:"total_amount"`
	AvgAmount   int64  `json:"avg_amount"`
}

type PotentialMatch struct {
	DTakoRowID    string    `json:"dtako_row_id"`
	Confidence    float32   `json:"confidence"`
	MatchReasons  []string  `json:"match_reasons"`
	DTakoData     map[string]interface{} `json:"dtako_data"`
}

// Error types for repository operations
type RepositoryError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Field   string `json:"field,omitempty"`
}

func (e *RepositoryError) Error() string {
	return e.Message
}

// Common error codes
const (
	ErrCodeNotFound       = "NOT_FOUND"
	ErrCodeDuplicateKey   = "DUPLICATE_KEY"
	ErrCodeInvalidInput   = "INVALID_INPUT"
	ErrCodeDatabaseError  = "DATABASE_ERROR"
	ErrCodeTransactionError = "TRANSACTION_ERROR"
)
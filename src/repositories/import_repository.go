package repositories

import (
	"context"

	"github.com/yhonda-ohishi/etc_meisai/src/models"
)

// ImportRepository defines the interface for import data access
type ImportRepository interface {
	// Session management
	CreateSession(ctx context.Context, session *models.ImportSession) error
	GetSession(ctx context.Context, sessionID string) (*models.ImportSession, error)
	UpdateSession(ctx context.Context, session *models.ImportSession) error
	ListSessions(ctx context.Context, params ListImportSessionsParams) ([]*models.ImportSession, int64, error)
	CancelSession(ctx context.Context, sessionID string) error

	// Record operations
	CreateRecord(ctx context.Context, record *models.ETCMeisaiRecord) error
	CreateRecordsBatch(ctx context.Context, records []*models.ETCMeisaiRecord) error
	FindRecordByHash(ctx context.Context, hash string) (*models.ETCMeisaiRecord, error)
	FindDuplicateRecords(ctx context.Context, hashes []string) ([]*models.ETCMeisaiRecord, error)

	// Transaction support
	BeginTx(ctx context.Context) (ImportRepository, error)
	CommitTx() error
	RollbackTx() error

	// Health check
	Ping(ctx context.Context) error
}

// ListImportSessionsParams contains parameters for listing import sessions
type ListImportSessionsParams struct {
	Page        int
	PageSize    int
	AccountType *string
	AccountID   *string
	Status      *string
	CreatedBy   *string
	SortBy      string
	SortOrder   string
}

// ImportResult contains statistics about an import operation
type ImportResult struct {
	TotalRecords   int
	SuccessCount   int
	ErrorCount     int
	DuplicateCount int
}

// ImportError represents an error that occurred during import
type ImportError struct {
	RowNumber int
	Field     string
	Value     string
	Message   string
}
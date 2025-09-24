package repositories

import (
	"context"
	"time"

	"github.com/yhonda-ohishi/etc_meisai/src/models"
)

// ETCMeisaiRecordRepository defines the interface for ETCMeisaiRecord data access
type ETCMeisaiRecordRepository interface {
	// Basic CRUD operations
	Create(ctx context.Context, record *models.ETCMeisaiRecord) error
	GetByID(ctx context.Context, id int64) (*models.ETCMeisaiRecord, error)
	Update(ctx context.Context, record *models.ETCMeisaiRecord) error
	Delete(ctx context.Context, id int64) error

	// Query operations
	GetByHash(ctx context.Context, hash string) (*models.ETCMeisaiRecord, error)
	CheckDuplicateHash(ctx context.Context, hash string, excludeID ...int64) (bool, error)

	// List operations with filtering
	List(ctx context.Context, params ListRecordsParams) ([]*models.ETCMeisaiRecord, int64, error)

	// Transaction support
	BeginTx(ctx context.Context) (ETCMeisaiRecordRepository, error)
	CommitTx() error
	RollbackTx() error

	// Health check
	Ping(ctx context.Context) error
}

// ListRecordsParams contains parameters for listing ETC records
type ListRecordsParams struct {
	Page      int
	PageSize  int
	DateFrom  *time.Time
	DateTo    *time.Time
	CarNumber *string
	ETCNumber *string
	ETCNum    *string
	SortBy    string // date, toll_amount, car_number
	SortOrder string // asc, desc
}
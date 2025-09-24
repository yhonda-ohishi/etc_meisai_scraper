package repositories

import (
	"context"
	"time"

	"github.com/yhonda-ohishi/etc_meisai/src/models"
)

// ETCMappingRepository defines the interface for ETCMapping data access
type ETCMappingRepository interface {
	// Basic CRUD operations
	Create(ctx context.Context, mapping *models.ETCMapping) error
	GetByID(ctx context.Context, id int64) (*models.ETCMapping, error)
	Update(ctx context.Context, mapping *models.ETCMapping) error
	Delete(ctx context.Context, id int64) error

	// Query operations
	GetByETCRecordID(ctx context.Context, etcRecordID int64) ([]*models.ETCMapping, error)
	GetByMappedEntity(ctx context.Context, entityType string, entityID int64) ([]*models.ETCMapping, error)
	GetActiveMapping(ctx context.Context, etcRecordID int64) (*models.ETCMapping, error)

	// List operations with filtering
	List(ctx context.Context, params ListMappingsParams) ([]*models.ETCMapping, int64, error)

	// Bulk operations
	BulkCreate(ctx context.Context, mappings []*models.ETCMapping) error
	UpdateStatus(ctx context.Context, id int64, status string) error

	// Transaction support
	BeginTx(ctx context.Context) (ETCMappingRepository, error)
	CommitTx() error
	RollbackTx() error

	// Health check
	Ping(ctx context.Context) error
}

// ListMappingsParams contains parameters for listing mappings
type ListMappingsParams struct {
	Page             int
	PageSize         int
	DateFrom         *time.Time
	DateTo           *time.Time
	MappingType      *string
	Status           *string
	MinConfidence    *float32
	MappedEntityType *string
	SortBy           string // created_at, confidence, etc_record_id
	SortOrder        string // asc, desc
}
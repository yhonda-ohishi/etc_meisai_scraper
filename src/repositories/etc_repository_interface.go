package repositories

import (
	"time"
	"github.com/yhonda-ohishi/etc_meisai/src/models"
)

// ETCRepository defines the interface for ETC data access
type ETCRepository interface {
	// Basic CRUD operations
	Create(etc *models.ETCMeisai) error
	GetByID(id int64) (*models.ETCMeisai, error)
	Update(etc *models.ETCMeisai) error
	Delete(id int64) error

	// Query operations
	GetByDateRange(from, to time.Time) ([]*models.ETCMeisai, error)
	GetByHash(hash string) (*models.ETCMeisai, error)
	List(params *models.ETCListParams) ([]*models.ETCMeisai, int64, error)

	// Bulk operations
	BulkInsert(records []*models.ETCMeisai) error
	CheckDuplicatesByHash(hashes []string) (map[string]bool, error)

	// Count operations
	CountByDateRange(from, to time.Time) (int64, error)

	// Search operations
	GetByETCNumber(etcNumber string, limit int) ([]*models.ETCMeisai, error)
	GetByCarNumber(carNumber string, limit int) ([]*models.ETCMeisai, error)

	// Summary operations
	GetSummaryByDateRange(from, to time.Time) (*models.ETCSummary, error)
}
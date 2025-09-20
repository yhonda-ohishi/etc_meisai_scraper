package repositories

import (
	"github.com/yhonda-ohishi/etc_meisai/src/models"
)

// MappingRepository defines the interface for ETC-DTako mapping operations
type MappingRepository interface {
	// Basic CRUD
	Create(mapping *models.ETCMeisaiMapping) error
	GetByID(id int64) (*models.ETCMeisaiMapping, error)
	Update(mapping *models.ETCMeisaiMapping) error
	Delete(id int64) error

	// Query operations
	GetByETCMeisaiID(etcMeisaiID int64) ([]*models.ETCMeisaiMapping, error)
	GetByDTakoRowID(dtakoRowID string) (*models.ETCMeisaiMapping, error)
	List(params *models.MappingListParams) ([]*models.ETCMeisaiMapping, int64, error)

	// Batch operations
	BulkCreateMappings(mappings []*models.ETCMeisaiMapping) error
	DeleteByETCMeisaiID(etcMeisaiID int64) error

	// Auto-matching support
	FindPotentialMatches(etcMeisaiID int64, threshold float32) ([]*models.PotentialMatch, error)
	UpdateConfidenceScore(id int64, confidence float32) error
}
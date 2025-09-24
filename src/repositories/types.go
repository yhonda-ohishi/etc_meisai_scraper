package repositories

import (
	"context"
	"time"
	"github.com/yhonda-ohishi/etc_meisai/src/models"
)

// RecordFilters defines filters for ETC record queries
type RecordFilters struct {
	DateFrom      time.Time
	DateTo        time.Time
	CarNumber     string
	ETCCardNumber string
	Limit         int
	Offset        int
}

// SearchQuery defines search parameters for ETC records
type SearchQuery struct {
	ETCCardNumber string
	CarNumber     string
	EntranceIC    string
	ExitIC        string
	DateFrom      time.Time
	DateTo        time.Time
	MinAmount     int
	MaxAmount     int
}

// MappingFilters defines filters for mapping queries
type MappingFilters struct {
	Status           string
	MappingType      string
	ETCRecordID      uint
	MappedEntityType string
	Limit            int
	Offset           int
}

// MappingStats contains aggregated mapping statistics
type MappingStats struct {
	TotalMappings    int
	ActiveMappings   int
	PendingMappings  int
	RejectedMappings int
	AvgConfidence    float64
	ByType           map[string]int
}

// Extended interfaces for repository operations (these extend the existing interfaces)

// ExtendedETCMeisaiRecordRepository extends the basic repository with additional operations
type ExtendedETCMeisaiRecordRepository interface {
	CreateETCMeisaiRecord(ctx context.Context, record *models.ETCMeisaiRecord) error
	GetETCMeisaiRecordByID(ctx context.Context, id uint) (*models.ETCMeisaiRecord, error)
	UpdateETCMeisaiRecord(ctx context.Context, record *models.ETCMeisaiRecord) error
	DeleteETCMeisaiRecord(ctx context.Context, id uint) error
	ListETCMeisaiRecords(ctx context.Context, filters *RecordFilters) ([]*models.ETCMeisaiRecord, error)
	BulkCreateETCMeisaiRecords(ctx context.Context, records []*models.ETCMeisaiRecord) error
	SearchETCMeisaiRecords(ctx context.Context, query *SearchQuery) ([]*models.ETCMeisaiRecord, error)
	GetHealth(ctx context.Context) error
}

// ExtendedMappingRepository extends the basic mapping repository with additional operations
type ExtendedMappingRepository interface {
	CreateMapping(ctx context.Context, mapping *models.ETCMapping) error
	GetMappingByID(ctx context.Context, id uint) (*models.ETCMapping, error)
	UpdateMapping(ctx context.Context, mapping *models.ETCMapping) error
	DeleteMapping(ctx context.Context, id uint) error
	FindMappingsByETCRecord(ctx context.Context, etcRecordID uint) ([]*models.ETCMapping, error)
	ListMappings(ctx context.Context, filters *MappingFilters) ([]*models.ETCMapping, error)
	UpdateMappingStatus(ctx context.Context, id uint, status models.MappingStatus) error
	BulkUpdateStatus(ctx context.Context, mappingIDs []uint, status models.MappingStatus) error
	GetMappingStats(ctx context.Context) (*MappingStats, error)
	CreateAutoMapping(ctx context.Context, etcRecordID uint) (*models.ETCMapping, error)
}
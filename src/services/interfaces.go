package services

import (
	"context"
	"io"
	"time"

	"github.com/yhonda-ohishi/etc_meisai/src/models"
)

// ETCServiceInterface defines the contract for ETC service operations
type ETCServiceInterface interface {
	// Core CRUD operations
	Create(ctx context.Context, meisai *models.ETCMeisai) (*models.ETCMeisai, error)
	GetByID(ctx context.Context, id int64) (*models.ETCMeisai, error)
	List(ctx context.Context, params *models.ETCListParams) ([]*models.ETCMeisai, int64, error)

	// Import operations
	ImportData(req models.ETCImportRequest) (*models.ETCImportResult, error)
	ImportCSV(ctx context.Context, records []*models.ETCMeisai) (*models.ETCImportResult, error)

	// Query operations
	GetMeisaiByDateRange(fromDate, toDate string) ([]models.ETCMeisai, error)
	GetByDateRange(ctx context.Context, from, to time.Time) ([]*models.ETCMeisai, error)
	GetSummary(ctx context.Context, fromDate, toDate string) (map[string]interface{}, error)
}

// MappingServiceInterface defines the contract for mapping service operations
type MappingServiceInterface interface {
	// Mapping CRUD operations
	CreateMapping(ctx context.Context, mapping *models.ETCMeisaiMapping) error
	GetMappingByID(ctx context.Context, id int64) (*models.ETCMeisaiMapping, error)
	GetMappingsByETCMeisaiID(ctx context.Context, etcMeisaiID int64) ([]*models.ETCMeisaiMapping, error)
	GetMappingsByDTakoRowID(ctx context.Context, dtakoRowID int64) ([]*models.ETCMeisaiMapping, error)
	UpdateMapping(ctx context.Context, mapping *models.ETCMeisaiMapping) error
	DeleteMapping(ctx context.Context, id int64) error

	// Bulk operations
	CreateBulkMappings(ctx context.Context, mappings []*models.ETCMeisaiMapping) (*models.BulkMappingResult, error)
	UpdateBulkMappings(ctx context.Context, mappings []*models.ETCMeisaiMapping) (*models.BulkMappingResult, error)
	DeleteBulkMappings(ctx context.Context, ids []int64) (*models.BulkMappingResult, error)

	// Auto-matching operations
	AutoMatch(ctx context.Context, startDate, endDate time.Time, threshold float32) ([]*models.AutoMatchResult, error)

	// Query operations
	GetUnmappedETCMeisai(ctx context.Context, limit, offset int) ([]*models.ETCMeisai, int64, error)
	GetMappingStats(ctx context.Context, fromDate, toDate time.Time) (*models.MappingStats, error)
}

// BaseServiceInterface defines the contract for base service operations
type BaseServiceInterface interface {
	// Health check operations
	HealthCheck(ctx context.Context) *HealthCheckResult
	GetVersion() string

	// Database client access
	GetDBClient() interface{}

	// Service lifecycle
	Start() error
	Stop() error
	IsReady() bool
}

// ImportServiceInterface defines the contract for import service operations
type ImportServiceInterface interface {
	// CSV import operations
	ImportCSV(ctx context.Context, reader io.Reader, params *ImportCSVParams) (*models.ETCImportResult, error)
	ImportCSVFile(ctx context.Context, filePath string, params *ImportCSVParams) (*models.ETCImportResult, error)

	// Validation operations
	ValidateCSV(ctx context.Context, reader io.Reader) (*models.ImportValidationResult, error)
	ValidateCSVFile(ctx context.Context, filePath string) (*models.ImportValidationResult, error)

	// Preview operations
	PreviewCSV(ctx context.Context, reader io.Reader, maxRows int) (*models.ImportPreviewResult, error)
	PreviewCSVFile(ctx context.Context, filePath string, maxRows int) (*models.ImportPreviewResult, error)

	// Progress tracking
	GetImportProgress(ctx context.Context, jobID string) (*models.ImportProgress, error)
	CancelImport(ctx context.Context, jobID string) error
}

// ETCMeisaiServiceInterface defines the interface for ETC Meisai service
type ETCMeisaiServiceInterface interface {
	CreateRecord(ctx context.Context, params *CreateRecordParams) (*models.ETCMeisaiRecord, error)
	GetRecord(ctx context.Context, id int64) (*models.ETCMeisaiRecord, error)
	ListRecords(ctx context.Context, params *ListRecordsParams) ([]*models.ETCMeisaiRecord, error)
	UpdateRecord(ctx context.Context, id int64, params *CreateRecordParams) (*models.ETCMeisaiRecord, error)
	DeleteRecord(ctx context.Context, id int64) error
}

// ETCMappingServiceInterface defines the interface for ETC Mapping service
type ETCMappingServiceInterface interface {
	CreateMapping(ctx context.Context, params *CreateMappingParams) (*models.ETCMapping, error)
	GetMapping(ctx context.Context, id int64) (*models.ETCMapping, error)
	ListMappings(ctx context.Context, params *ListMappingsParams) ([]*models.ETCMapping, error)
	UpdateMapping(ctx context.Context, id int64, params *UpdateMappingParams) (*models.ETCMapping, error)
	DeleteMapping(ctx context.Context, id int64) error
}

// StatisticsServiceInterface defines the interface for Statistics service
type StatisticsServiceInterface interface {
	GetStatistics(ctx context.Context, params *GetStatisticsParams) (*models.Statistics, error)
	RefreshStatistics(ctx context.Context) (*models.Statistics, error)
}


// Additional types needed for interfaces

// ImportOptions contains optional parameters for import operations
type ImportOptions struct {
	SkipValidation bool `json:"skip_validation"`
	BatchSize      int  `json:"batch_size"`
	MaxErrors      int  `json:"max_errors"`
}

// GetStatisticsParams contains parameters for statistics queries
type GetStatisticsParams struct {
	AccountType string
	AccountID   string
	DateFrom    string
	DateTo      string
}
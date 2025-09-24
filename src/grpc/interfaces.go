package grpc

import (
	"context"
	"io"

	"github.com/yhonda-ohishi/etc_meisai/src/models"
	"github.com/yhonda-ohishi/etc_meisai/src/services"
)

// Service interfaces for testing

// ETCMeisaiServiceInterface defines the interface for ETC Meisai service
type ETCMeisaiServiceInterface interface {
	CreateRecord(ctx context.Context, params *services.CreateRecordParams) (*models.ETCMeisaiRecord, error)
	GetRecord(ctx context.Context, id int64) (*models.ETCMeisaiRecord, error)
	ListRecords(ctx context.Context, params *services.ListRecordsParams) (*services.ListRecordsResponse, error)
	UpdateRecord(ctx context.Context, id int64, params *services.CreateRecordParams) (*models.ETCMeisaiRecord, error)
	DeleteRecord(ctx context.Context, id int64) error
	HealthCheck(ctx context.Context) error
}

// ETCMappingServiceInterface defines the interface for ETC Mapping service
type ETCMappingServiceInterface interface {
	CreateMapping(ctx context.Context, params *services.CreateMappingParams) (*models.ETCMapping, error)
	GetMapping(ctx context.Context, id int64) (*models.ETCMapping, error)
	ListMappings(ctx context.Context, params *services.ListMappingsParams) (*services.ListMappingsResponse, error)
	UpdateMapping(ctx context.Context, id int64, params *services.UpdateMappingParams) (*models.ETCMapping, error)
	DeleteMapping(ctx context.Context, id int64) error
	UpdateStatus(ctx context.Context, id int64, status string) error
	HealthCheck(ctx context.Context) error
}

// ImportServiceInterface defines the interface for Import service
type ImportServiceInterface interface {
	ImportCSV(ctx context.Context, params *services.ImportCSVParams, reader io.Reader) (*services.ImportCSVResult, error)
	ImportCSVStream(ctx context.Context, params *services.ImportCSVStreamParams) (*services.ImportCSVResult, error)
	GetImportSession(ctx context.Context, sessionID string) (*models.ImportSession, error)
	ListImportSessions(ctx context.Context, params *services.ListImportSessionsParams) (*services.ListImportSessionsResponse, error)
	ProcessCSV(ctx context.Context, rows []*services.CSVRow, options *services.BulkProcessOptions) (*services.BulkProcessResult, error)
	ProcessCSVRow(ctx context.Context, row *services.CSVRow) (*models.ETCMeisaiRecord, error)
	HandleDuplicates(ctx context.Context, records []*models.ETCMeisaiRecord) ([]*services.DuplicateResult, error)
	CancelImportSession(ctx context.Context, sessionID string) error
	HealthCheck(ctx context.Context) error
}

// StatisticsServiceInterface defines the interface for Statistics service
type StatisticsServiceInterface interface {
	GetGeneralStatistics(ctx context.Context, filter *services.StatisticsFilter) (*services.GeneralStatistics, error)
	GetDailyStatistics(ctx context.Context, filter *services.StatisticsFilter) (*services.DailyStatisticsResponse, error)
	GetMonthlyStatistics(ctx context.Context, filter *services.StatisticsFilter) (*services.MonthlyStatisticsResponse, error)
	GetVehicleStatistics(ctx context.Context, carNumbers []string, filter *services.StatisticsFilter) (*services.VehicleStatisticsResponse, error)
	GetMappingStatistics(ctx context.Context, filter *services.StatisticsFilter) (*services.MappingStatisticsResponse, error)
	HealthCheck(ctx context.Context) error
}
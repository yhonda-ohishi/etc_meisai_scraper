// Package contracts defines the public interfaces for the ETC明細 Go module
package contracts

import (
	"time"
)

// ETCClient はETC明細データを取得・処理するメインクライアントインターフェース
type ETCClient interface {
	// DownloadETCData は複数アカウントのETC明細データをダウンロードします
	// 各アカウントに対して並列処理を実行し、結果を返します
	DownloadETCData(accounts []SimpleAccount, fromDate, toDate time.Time) ([]DownloadResult, error)

	// DownloadETCDataSingle は単一アカウントのETC明細データをダウンロードします
	DownloadETCDataSingle(userID, password string, fromDate, toDate time.Time) (*DownloadResult, error)

	// ParseETCCSV はCSVファイルを解析してETC明細レコードを返します
	ParseETCCSV(csvPath string) ([]ETCMeisai, error)

	// ExportToExcel はETC明細レコードをExcelファイルにエクスポートします（未実装）
	ExportToExcel(records []ETCMeisai, path string) error

	// ExportToPDF はETC明細レコードをPDFファイルにエクスポートします（未実装）
	ExportToPDF(records []ETCMeisai, path string) error

	// GetStatistics はETC明細レコードから統計情報を生成します（未実装）
	GetStatistics(records []ETCMeisai) *Statistics

	// FilterRecords は指定された条件でレコードをフィルタリングします（未実装）
	FilterRecords(records []ETCMeisai, criteria FilterCriteria) []ETCMeisai

	// ValidateAccounts はアカウント情報の有効性を検証します（未実装）
	ValidateAccounts(accounts []SimpleAccount) []AccountValidationResult

	// MergeRecords は複数のETC明細レコードセットをマージします（未実装）
	MergeRecords(records1, records2 []ETCMeisai) []ETCMeisai
}

// AccountManager はアカウント管理機能のインターフェース
type AccountManager interface {
	// LoadCorporateAccounts は法人アカウント情報を読み込みます
	LoadCorporateAccounts() ([]SimpleAccount, error)

	// LoadPersonalAccounts は個人アカウント情報を読み込みます
	LoadPersonalAccounts() ([]SimpleAccount, error)

	// SaveAccount はアカウント情報を保存します
	SaveAccount(account SimpleAccount) error

	// DeleteAccount はアカウント情報を削除します
	DeleteAccount(userID string) error

	// ListAccounts はすべてのアカウントを一覧表示します
	ListAccounts() ([]SimpleAccount, error)
}

// ProgressTracker は進捗追跡機能のインターフェース
type ProgressTracker interface {
	// StartJob は新しいジョブを開始します
	StartJob(jobID string, totalCount int) error

	// UpdateProgress はジョブの進捗を更新します
	UpdateProgress(jobID string, current int) error

	// CompleteJob はジョブを完了としてマークします
	CompleteJob(jobID string) error

	// FailJob はジョブを失敗としてマークします
	FailJob(jobID string, err error) error

	// GetJobStatus はジョブのステータスを取得します
	GetJobStatus(jobID string) (*JobStatus, error)
}

// DataProcessor はデータ処理機能のインターフェース
type DataProcessor interface {
	// ProcessDownloadedData はダウンロードされたデータを処理します
	ProcessDownloadedData(result DownloadResult) (*ProcessingResult, error)

	// BulkProcess は複数のダウンロード結果を一括処理します
	BulkProcess(results []DownloadResult) (*BulkProcessingResult, error)

	// TransformData はデータを指定された形式に変換します
	TransformData(records []ETCMeisai, format string) (interface{}, error)
}

// MappingManager はETC-デジタコマッピング管理のインターフェース
type MappingManager interface {
	// CreateMapping は新しいマッピングを作成します
	CreateMapping(mapping ETCDtakoMapping) error

	// GetMapping はマッピング情報を取得します
	GetMapping(id int64) (*ETCDtakoMapping, error)

	// UpdateMapping はマッピング情報を更新します
	UpdateMapping(mapping ETCDtakoMapping) error

	// DeleteMapping はマッピングを削除します
	DeleteMapping(id int64) error

	// ListMappings はマッピングを一覧表示します
	ListMappings(page, pageSize int) ([]ETCDtakoMapping, int, error)

	// GetUnmappedRecords は未マッピングのETC明細を取得します
	GetUnmappedRecords(limit int) ([]ETCMeisai, error)

	// AutoLink は自動マッピングを実行します
	AutoLink(threshold float64) (*AutoLinkResult, error)
}

// Statistics は統計情報の構造体（未実装）
type Statistics struct {
	TotalRecords   int
	TotalAmount    int64
	AverageAmount  float64
	ByVehicle      map[string]VehicleStats
	ByRoute        map[string]RouteStats
	ByMonth        map[string]MonthlyStats
	TopExpenses    []ExpenseItem
	CostReduction  float64
}

// FilterCriteria はフィルタリング条件（未実装）
type FilterCriteria struct {
	DateFrom      *time.Time
	DateTo        *time.Time
	VehicleNums   []string
	MinAmount     *int
	MaxAmount     *int
	RoutePatterns []string
	CardNumbers   []string
}

// AccountValidationResult はアカウント検証結果（未実装）
type AccountValidationResult struct {
	UserID  string
	IsValid bool
	Errors  []string
}

// JobStatus はジョブステータス情報
type JobStatus struct {
	JobID       string
	Status      string
	Progress    int
	TotalCount  int
	StartedAt   time.Time
	CompletedAt *time.Time
	Error       string
}

// ProcessingResult はデータ処理結果
type ProcessingResult struct {
	ProcessedCount int
	SkippedCount   int
	ErrorCount     int
	Errors         []error
}

// BulkProcessingResult はバルク処理結果
type BulkProcessingResult struct {
	TotalProcessed int
	SuccessCount   int
	FailureCount   int
	Results        []ProcessingResult
}

// AutoLinkResult は自動マッピング結果
type AutoLinkResult struct {
	TotalProcessed int
	LinkedCount    int
	SkippedCount   int
	AverageScore   float64
}

// VehicleStats は車両別統計
type VehicleStats struct {
	VehicleNum   string
	TotalAmount  int64
	TripCount    int
	AverageAmount float64
}

// RouteStats は経路別統計
type RouteStats struct {
	Route        string
	UsageCount   int
	TotalAmount  int64
	AverageAmount float64
}

// MonthlyStats は月別統計
type MonthlyStats struct {
	Month        string
	TotalAmount  int64
	RecordCount  int
	AverageAmount float64
}

// ExpenseItem は費用項目
type ExpenseItem struct {
	Description string
	Amount      int64
	Date        time.Time
	VehicleNum  string
}
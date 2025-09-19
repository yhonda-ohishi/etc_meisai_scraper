// Package contracts defines the data models for the ETC明細 Go module
package contracts

import (
	"time"
)

// SimpleAccount はETCポータルサイトへのログイン情報を表します
type SimpleAccount struct {
	UserID   string `json:"user_id" validate:"required"`
	Password string `json:"password" validate:"required"`
	Type     string `json:"type" validate:"required,oneof=corporate personal"`
}

// ETCMeisai はETC明細レコードを表します
type ETCMeisai struct {
	ID             int64     `json:"id" db:"id"`
	UsageDate      time.Time `json:"usage_date" db:"usage_date" validate:"required"`
	EntryIC        string    `json:"entry_ic" db:"entry_ic" validate:"required"`
	ExitIC         string    `json:"exit_ic" db:"exit_ic" validate:"required"`
	RouteInfo      string    `json:"route_info" db:"route_info" validate:"required"`
	VehicleNum     string    `json:"vehicle_num" db:"vehicle_num" validate:"required"`
	ETCCardNo      string    `json:"etc_card_no" db:"etc_card_no" validate:"required,len=16"`
	EntryTime      string    `json:"entry_time" db:"entry_time"`
	ExitTime       string    `json:"exit_time" db:"exit_time"`
	TotalCharge    int       `json:"total_charge" db:"total_charge" validate:"min=0"`
	UnkoNo         string    `json:"unko_no" db:"unko_no"`
	Direction      string    `json:"direction" db:"direction" validate:"oneof=上り 下り 不明"`
	ImportedAt     time.Time `json:"imported_at" db:"imported_at"`
	AccountType    string    `json:"account_type" db:"account_type" validate:"required,oneof=corporate personal"`
	AccountUserID  string    `json:"account_user_id" db:"account_user_id" validate:"required"`
}

// DownloadResult はダウンロード処理の結果を表します
type DownloadResult struct {
	JobID        string      `json:"job_id" validate:"required,uuid"`
	AccountID    string      `json:"account_id" validate:"required"`
	Status       string      `json:"status" validate:"required,oneof=pending running completed failed"`
	Progress     int         `json:"progress" validate:"min=0,max=100"`
	CSVPath      string      `json:"csv_path,omitempty"`
	RecordCount  int         `json:"record_count,omitempty" validate:"min=0"`
	Error        string      `json:"error,omitempty"`
	StartedAt    time.Time   `json:"started_at"`
	CompletedAt  *time.Time  `json:"completed_at,omitempty"`
	Records      []ETCMeisai `json:"records,omitempty"`
}

// ETCDtakoMapping はETC明細とデジタコデータのマッピングを表します
type ETCDtakoMapping struct {
	ID           int64     `json:"id" db:"id"`
	ETCMeisaiID  int64     `json:"etc_meisai_id" db:"etc_meisai_id" validate:"required"`
	DtakoRowID   string    `json:"dtako_row_id" db:"dtako_row_id" validate:"required"`
	MappingType  string    `json:"mapping_type" db:"mapping_type" validate:"required,oneof=auto manual"`
	MatchScore   float64   `json:"match_score,omitempty" db:"match_score" validate:"min=0,max=1"`
	CreatedBy    string    `json:"created_by" db:"created_by" validate:"required"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
	IsActive     bool      `json:"is_active" db:"is_active"`
}

// ETCImportRequest はデータインポートリクエストを表します
type ETCImportRequest struct {
	FromDate          time.Time       `json:"from_date" validate:"required"`
	ToDate            time.Time       `json:"to_date" validate:"required,gtfield=FromDate"`
	Accounts          []SimpleAccount `json:"accounts" validate:"required,min=1,dive"`
	Mode              string          `json:"mode" validate:"required,oneof=sync async"`
	OverwriteExisting bool            `json:"overwrite_existing"`
}

// ETCImportResult はデータインポート結果を表します
type ETCImportResult struct {
	TotalRecords    int           `json:"total_records" validate:"min=0"`
	ImportedRecords int           `json:"imported_records" validate:"min=0"`
	SkippedRecords  int           `json:"skipped_records" validate:"min=0"`
	FailedRecords   int           `json:"failed_records" validate:"min=0"`
	Errors          []ImportError `json:"errors,omitempty"`
	Duration        time.Duration `json:"duration" validate:"min=0"`
}

// ImportError はインポート処理のエラー詳細を表します
type ImportError struct {
	RecordIndex int    `json:"record_index"`
	Field       string `json:"field"`
	Value       string `json:"value"`
	Message     string `json:"message"`
}

// ETCSummary は統計情報の集計結果を表します
type ETCSummary struct {
	Period        string            `json:"period" validate:"required,oneof=daily weekly monthly"`
	StartDate     time.Time         `json:"start_date" validate:"required"`
	EndDate       time.Time         `json:"end_date" validate:"required,gtfield=StartDate"`
	TotalRecords  int               `json:"total_records" validate:"min=0"`
	TotalAmount   int64             `json:"total_amount" validate:"min=0"`
	VehicleCount  int               `json:"vehicle_count" validate:"min=0"`
	RouteCount    int               `json:"route_count" validate:"min=0"`
	ByVehicle     map[string]int64  `json:"by_vehicle,omitempty"`
	ByRoute       map[string]int64  `json:"by_route,omitempty"`
}

// DownloadJobRequest は非同期ダウンロードジョブのリクエストを表します
type DownloadJobRequest struct {
	Accounts  []SimpleAccount `json:"accounts" validate:"required,min=1,dive"`
	FromDate  time.Time       `json:"from_date" validate:"required"`
	ToDate    time.Time       `json:"to_date" validate:"required,gtfield=FromDate"`
	Callback  string          `json:"callback,omitempty" validate:"omitempty,url"`
}

// DownloadJobResponse は非同期ダウンロードジョブのレスポンスを表します
type DownloadJobResponse struct {
	JobID     string    `json:"job_id"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	Message   string    `json:"message,omitempty"`
}

// ProcessRequest はデータ処理リクエストを表します
type ProcessRequest struct {
	JobID          string `json:"job_id,omitempty" validate:"omitempty,uuid"`
	CSVPath        string `json:"csv_path,omitempty" validate:"omitempty,file"`
	SaveToDatabase bool   `json:"save_to_database"`
	GenerateSummary bool   `json:"generate_summary"`
}

// ProcessResponse はデータ処理レスポンスを表します
type ProcessResponse struct {
	Success        bool        `json:"success"`
	ProcessedCount int         `json:"processed_count"`
	SavedCount     int         `json:"saved_count,omitempty"`
	Summary        *ETCSummary `json:"summary,omitempty"`
	Errors         []string    `json:"errors,omitempty"`
}

// MappingRequest はマッピングリクエストを表します
type MappingRequest struct {
	ETCMeisaiID int64   `json:"etc_meisai_id" validate:"required"`
	DtakoRowID  string  `json:"dtako_row_id" validate:"required"`
	MatchScore  float64 `json:"match_score,omitempty" validate:"min=0,max=1"`
}

// MappingResponse はマッピングレスポンスを表します
type MappingResponse struct {
	MappingID   int64     `json:"mapping_id"`
	Success     bool      `json:"success"`
	Message     string    `json:"message,omitempty"`
	CreatedAt   time.Time `json:"created_at,omitempty"`
}

// AutoLinkRequest は自動リンクリクエストを表します
type AutoLinkRequest struct {
	Threshold     float64   `json:"threshold" validate:"required,min=0,max=1"`
	DateFrom      time.Time `json:"date_from,omitempty"`
	DateTo        time.Time `json:"date_to,omitempty" validate:"omitempty,gtfield=DateFrom"`
	MaxLinks      int       `json:"max_links,omitempty" validate:"omitempty,min=1"`
}

// ErrorResponse は汎用エラーレスポンスを表します
type ErrorResponse struct {
	Error      string            `json:"error"`
	Message    string            `json:"message"`
	Code       int               `json:"code,omitempty"`
	Details    map[string]string `json:"details,omitempty"`
	RequestID  string            `json:"request_id,omitempty"`
	Timestamp  time.Time         `json:"timestamp"`
}
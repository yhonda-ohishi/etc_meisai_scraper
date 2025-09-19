package models

import "time"

// ETCMeisai represents a simplified ETC transaction record
type ETCMeisai struct {
	ID             int64     `db:"id" json:"id"`
	UsageDate      time.Time `db:"usage_date" json:"usage_date"`
	Date           string    `db:"date" json:"date"`                       // 日付文字列
	Time           string    `db:"time" json:"time"`                       // 時刻文字列
	EntryIC        string    `db:"entry_ic" json:"entry_ic"`
	ICEntry        string    `db:"ic_entry" json:"ic_entry"`               // エイリアス
	ExitIC         string    `db:"exit_ic" json:"exit_ic"`
	ICExit         string    `db:"ic_exit" json:"ic_exit"`                 // エイリアス
	TollGate       string    `db:"toll_gate" json:"toll_gate"`             // 料金所
	TollAmount     int       `db:"toll_amount" json:"toll_amount"`
	TotalAmount    int       `db:"total_amount" json:"total_amount"`       // エイリアス
	VehicleNumber  string    `db:"vehicle_number" json:"vehicle_number"`
	VehicleNo      string    `db:"vehicle_no" json:"vehicle_no"`           // エイリアス
	VehicleType    string    `db:"vehicle_type" json:"vehicle_type"`       // 車種
	ETCCardNumber  string    `db:"etc_card_number" json:"etc_card_number"`
	CardNo         string    `db:"card_no" json:"card_no"`                 // エイリアス
	CardNumber     string    `db:"card_number" json:"card_number"`         // エイリアス
	ETCNum         string    `db:"etc_num" json:"etc_num"`                 // エイリアス
	Amount         int       `db:"amount" json:"amount"`                   // エイリアス
	DiscountAmount int       `db:"discount_amount" json:"discount_amount"` // 割引金額
	PaymentMethod  string    `db:"payment_method" json:"payment_method"`   // 支払い方法
	RouteCode      string    `db:"route_code" json:"route_code"`           // ルートコード
	Distance       float64   `db:"distance" json:"distance"`               // 距離
	Remarks        string    `db:"remarks" json:"remarks"`                 // 備考
	UsageType      string    `db:"usage_type" json:"usage_type"`           // 通行区分
	AccountType    string    `db:"account_type" json:"account_type"`       // corporate/personal
	UnkoNo         string    `db:"unko_no" json:"unko_no"`                 // 運行番号
	RowID          int64     `db:"row_id" json:"row_id"`                   // レガシー互換性
	ImportedAt     time.Time `db:"imported_at" json:"imported_at"`
	CreatedAt      time.Time `db:"created_at" json:"created_at"`
	UpdatedAt      time.Time `db:"updated_at" json:"updated_at"`
}

// ETCRow represents a row from legacy system (for compatibility)
type ETCRow = ETCMeisai

// ETCSummary represents summary statistics
type ETCSummary struct {
	Period        string           `db:"period" json:"period"`
	StartDate     time.Time        `db:"start_date" json:"start_date"`
	EndDate       time.Time        `db:"end_date" json:"end_date"`
	Date          string           `db:"date" json:"date"`                 // 日付
	VehicleNo     string           `db:"vehicle_no" json:"vehicle_no"`     // 車両番号
	TotalRecords  int              `db:"total_records" json:"total_records"`
	TotalAmount   int64            `db:"total_amount" json:"total_amount"`
	TotalCount    int              `db:"total_count" json:"total_count"`     // 合計件数
	TotalDistance float64          `db:"total_distance" json:"total_distance"` // 合計距離
	VehicleCount  int              `db:"vehicle_count" json:"vehicle_count"`
	RouteCount    int              `db:"route_count" json:"route_count"`
	ByVehicle     map[string]int64 `db:"by_vehicle" json:"by_vehicle"`
	ByRoute       map[string]int64 `db:"by_route" json:"by_route"`
}

// ImportSession represents an import session record (optional)
type ImportSession struct {
	ID           int64     `db:"id" json:"id"`
	AccountType  string    `db:"account_type" json:"account_type"`
	StartDate    time.Time `db:"start_date" json:"start_date"`
	EndDate      time.Time `db:"end_date" json:"end_date"`
	RecordCount  int       `db:"record_count" json:"record_count"`
	Status       string    `db:"status" json:"status"` // success/failed
	ExecutedAt   time.Time `db:"executed_at" json:"executed_at"`
	ErrorMessage string    `db:"error_message" json:"error_message"`
}

// HashConfig represents hash configuration settings
type HashConfig struct {
	Algorithm string `json:"algorithm"`
	Fields    []string `json:"fields"`
}

// ETCMeisaiWithHash represents ETC meisai with hash value
type ETCMeisaiWithHash struct {
	ETCMeisai
	Hash string `db:"hash" json:"hash"`
}

// HashIndex represents hash index for quick lookup
type HashIndex struct {
	Hash      string    `db:"hash"`
	RecordID  int64     `db:"record_id"`
	CreatedAt time.Time `db:"created_at"`
}

// ImportDiff represents differences in import
type ImportDiff struct {
	Added   []ETCMeisai `json:"added"`
	Updated []ETCMeisai `json:"updated"`
	Deleted []ETCMeisai `json:"deleted"`
}

// ETCImportRequest represents import request parameters
type ETCImportRequest struct {
	FilePath    string `json:"file_path"`
	AccountType string `json:"account_type"`
	StartDate   string `json:"start_date"`
	EndDate     string `json:"end_date"`
	FromDate    string `json:"from_date"`    // エイリアス
	ToDate      string `json:"to_date"`      // エイリアス
}

// ETCImportResult represents import operation result
type ETCImportResult struct {
	Success      bool      `json:"success"`
	RecordCount  int       `json:"record_count"`
	ImportedRows int       `json:"imported_rows"`     // エイリアス
	Message      string    `json:"message"`
	ErrorMessage string    `json:"error_message,omitempty"`
	Duration     int64     `json:"duration_ms"`
	ImportedAt   time.Time `json:"imported_at"`
	Errors       []string  `json:"errors,omitempty"`
}
// ETCDtakoMapping represents mapping between ETC and Dtako records
type ETCDtakoMapping struct {
	ID          int64     `db:"id" json:"id"`
	ETCMeisaiID int64     `db:"etc_meisai_id" json:"etc_meisai_id"`
	DtakoRowID  string    `db:"dtako_row_id" json:"dtako_row_id"`
	VehicleID   string    `db:"vehicle_id" json:"vehicle_id"`
	MappingType string    `db:"mapping_type" json:"mapping_type"`
	Notes       string    `db:"notes" json:"notes"`
	CreatedAt   time.Time `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time `db:"updated_at" json:"updated_at"`
	CreatedBy   string    `db:"created_by" json:"created_by"`
}

// DefaultHashConfigs provides default hash configurations
var DefaultHashConfigs = map[string]HashConfig{
	"file": {
		Algorithm: "sha256",
		Fields:    []string{"filename", "size"},
	},
	"record": {
		Algorithm: "sha256", 
		Fields:    []string{"date", "ic_entry", "ic_exit", "vehicle_no", "amount"},
	},
}

// UpdatedRecord represents an updated record
type UpdatedRecord struct {
	ETCMeisai
	UpdateType string `json:"update_type"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Code    string `json:"code,omitempty"`
	Details string `json:"details,omitempty"`
}

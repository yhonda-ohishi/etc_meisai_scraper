package models

import "time"

// ETCAccount はETCアカウント情報
type ETCAccount struct {
	ID       string
	Username string
	Password string
	Type     string // "corporate" or "personal"
}

// ETCMeisaiRecord はETC明細レコード
type ETCMeisaiRecord struct {
	ID             int64      `json:"id"`
	AccountID      string     `json:"account_id"`
	UsageDate      time.Time  `json:"usage_date"`
	EntryIC        string     `json:"entry_ic"`
	ExitIC         string     `json:"exit_ic"`
	VehicleNumber  string     `json:"vehicle_number"`
	ETCCardNumber  string     `json:"etc_card_number"`
	Amount         int        `json:"amount"`
	CSVFileName    string     `json:"csv_file_name"`
	DownloadedAt   time.Time  `json:"downloaded_at"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

// DownloadResult はダウンロード結果
type DownloadResult struct {
	Success      bool                   `json:"success"`
	RecordCount  int                    `json:"record_count"`
	CSVPath      string                 `json:"csv_path"`
	Records      []ETCMeisaiRecord      `json:"records"`
	Error        string                 `json:"error,omitempty"`
}
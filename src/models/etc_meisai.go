package models

import (
	"crypto/sha256"
	"fmt"
	"time"
)

// ETCMeisai represents the main ETC transaction record
type ETCMeisai struct {
	// Primary Key
	ID int64 `json:"id"`

	// 利用情報
	UseDate time.Time `json:"use_date"`
	UseTime string    `json:"use_time"`

	// 料金所情報
	EntryIC string `json:"entry_ic"`
	ExitIC  string `json:"exit_ic"`

	// 金額情報
	Amount int32 `json:"amount"`

	// 車両情報
	CarNumber string `json:"car_number"`

	// ETC情報
	ETCNumber string `json:"etc_number"`

	// データ整合性
	Hash string `json:"hash"`

	// タイムスタンプ
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// リレーション
	Mappings []ETCMeisaiMapping `json:"mappings,omitempty"`
}

// BeforeCreate prepares the record before creation
func (e *ETCMeisai) BeforeCreate() error {
	// Set timestamps
	now := time.Now()
	if e.CreatedAt.IsZero() {
		e.CreatedAt = now
	}
	if e.UpdatedAt.IsZero() {
		e.UpdatedAt = now
	}

	// Generate hash if not provided
	if e.Hash == "" {
		e.Hash = e.GenerateHash()
	}

	return nil
}

// BeforeUpdate prepares the record before updating
func (e *ETCMeisai) BeforeUpdate() error {
	// Update timestamp
	e.UpdatedAt = time.Now()

	// Re-generate hash on update
	e.Hash = e.GenerateHash()

	return nil
}

// GenerateHash creates a unique hash for the ETC record
func (e *ETCMeisai) GenerateHash() string {
	data := fmt.Sprintf("%s|%s|%s|%s|%d|%s|%s",
		e.UseDate.Format("2006-01-02"),
		e.UseTime,
		e.EntryIC,
		e.ExitIC,
		e.Amount,
		e.CarNumber,
		e.ETCNumber,
	)

	hash := sha256.Sum256([]byte(data))
	return fmt.Sprintf("%x", hash)
}

// Validate checks the ETC record for business rule compliance
func (e *ETCMeisai) Validate() error {
	if e.UseDate.IsZero() {
		return fmt.Errorf("UseDate is required")
	}

	if e.UseTime == "" {
		return fmt.Errorf("UseTime is required")
	}

	if e.EntryIC == "" {
		return fmt.Errorf("EntryIC is required")
	}

	if e.ExitIC == "" {
		return fmt.Errorf("ExitIC is required")
	}

	if e.Amount <= 0 {
		return fmt.Errorf("Amount must be positive")
	}

	if e.CarNumber == "" {
		return fmt.Errorf("CarNumber is required")
	}

	if e.ETCNumber == "" {
		return fmt.Errorf("ETCNumber is required")
	}

	if len(e.ETCNumber) > 20 {
		return fmt.Errorf("ETCNumber too long")
	}

	return nil
}

// ETCListParams defines parameters for querying ETC records
type ETCListParams struct {
	Limit     int        `json:"limit"`
	Offset    int        `json:"offset"`
	StartDate *time.Time `json:"start_date,omitempty"`
	EndDate   *time.Time `json:"end_date,omitempty"`
	FromDate  *time.Time `json:"from_date,omitempty"` // Alias for StartDate
	ToDate    *time.Time `json:"to_date,omitempty"`   // Alias for EndDate
	ETCNumber string     `json:"etc_number,omitempty"`
	CarNumber string     `json:"car_number,omitempty"`
	EntryIC   string     `json:"entry_ic,omitempty"`
	ExitIC    string     `json:"exit_ic,omitempty"`
	SortBy    string     `json:"sort_by,omitempty"`
	OrderBy   string     `json:"order_by,omitempty"`   // Alias for SortBy
	SortOrder string     `json:"sort_order,omitempty"`
}

// SetDefaults sets default values for list parameters
func (p *ETCListParams) SetDefaults() {
	if p.Limit <= 0 || p.Limit > 1000 {
		if p.Limit > 1000 {
			p.Limit = 1000
		} else {
			p.Limit = 100
		}
	}
	if p.Offset < 0 {
		p.Offset = 0
	}
}

// ETCSummary provides aggregated statistics for ETC records
type ETCSummary struct {
	TotalAmount int64     `json:"total_amount"`
	TotalCount  int64     `json:"total_count"`
	StartDate   time.Time `json:"start_date"`
	EndDate     time.Time `json:"end_date"`
	ByETCNumber map[string]*ETCNumberSummary  `json:"by_etc_number"`
	ByMonth     map[string]*ETCMonthlySummary `json:"by_month"`
}

// ETCNumberSummary provides summary statistics per ETC number
type ETCNumberSummary struct {
	ETCNumber   string `json:"etc_number"`
	TotalAmount int64  `json:"total_amount"`
	TotalCount  int64  `json:"total_count"`
}

// ETCMonthlySummary provides summary statistics per month
type ETCMonthlySummary struct {
	Year        int   `json:"year"`
	Month       int   `json:"month"`
	TotalAmount int64 `json:"total_amount"`
	TotalCount  int64 `json:"total_count"`
}

// GetTableName returns the table name
func (e *ETCMeisai) GetTableName() string {
	return "etc_meisai"
}

// String returns a string representation of the ETC record
func (e *ETCMeisai) String() string {
	return fmt.Sprintf("ETCMeisai{ID:%d, Date:%s, ETCNumber:%s, EntryIC:%s, ExitIC:%s, Amount:%d}",
		e.ID, e.UseDate.Format("2006-01-02"), e.ETCNumber, e.EntryIC, e.ExitIC, e.Amount)
}

// ETCMonthlyStats provides detailed monthly statistics
type ETCMonthlyStats struct {
	Year           int                          `json:"year"`
	Month          int                          `json:"month"`
	TotalAmount    int64                        `json:"total_amount"`
	TotalCount     int64                        `json:"total_count"`
	DailyBreakdown map[int]*ETCDailySummary     `json:"daily_breakdown"`
	TopRoutes      []*ETCRouteSummary           `json:"top_routes"`
}

// ETCDailySummary provides daily statistics within a month
type ETCDailySummary struct {
	Day    int   `json:"day"`
	Amount int64 `json:"amount"`
	Count  int64 `json:"count"`
}

// ETCRouteSummary provides statistics for popular routes
type ETCRouteSummary struct {
	EntryIC     string `json:"entry_ic"`
	ExitIC      string `json:"exit_ic"`
	Count       int64  `json:"count"`
	TotalAmount int64  `json:"total_amount"`
	AvgAmount   int64  `json:"avg_amount"`
}
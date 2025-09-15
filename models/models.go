package models

import "time"

// ETCMeisai represents an ETC transaction record
type ETCMeisai struct {
	ID               int        `json:"id"`
	UnkoNo           string     `json:"unko_no"`                // 運行NO
	Date             time.Time  `json:"date"`                   // 日付
	Time             string     `json:"time"`                   // 時刻
	ICEntry          string     `json:"ic_entry"`               // IC入口
	ICExit           string     `json:"ic_exit"`                // IC出口
	VehicleNo        string     `json:"vehicle_no"`             // 車両番号
	CardNo           string     `json:"card_no"`                // ETCカード番号
	Amount           int        `json:"amount"`                 // 利用金額
	DiscountAmount   int        `json:"discount_amount"`        // 割引金額
	TotalAmount      int        `json:"total_amount"`           // 請求金額
	UsageType        string     `json:"usage_type"`             // 利用区分
	PaymentMethod    string     `json:"payment_method"`         // 支払方法
	RouteCode        string     `json:"route_code"`             // 路線コード
	Distance         float64    `json:"distance"`               // 走行距離
	CreatedAt        *time.Time `json:"created_at,omitempty"`
	UpdatedAt        *time.Time `json:"updated_at,omitempty"`
}

// ETCImportRequest represents an import request
type ETCImportRequest struct {
	FromDate  string `json:"from_date" example:"2025-01-01"`
	ToDate    string `json:"to_date" example:"2025-01-31"`
	CardNo    string `json:"card_no,omitempty" example:"1234-5678-9012-3456"`
}

// ETCImportResult represents the result of an import operation
type ETCImportResult struct {
	Success      bool      `json:"success" example:"true"`
	ImportedRows int       `json:"imported_rows" example:"150"`
	Message      string    `json:"message" example:"Imported 150 rows successfully"`
	ImportedAt   time.Time `json:"imported_at" example:"2025-01-13T15:04:05Z"`
	Errors       []string  `json:"errors,omitempty"`
}

// ETCSummary represents a summary of ETC usage
type ETCSummary struct {
	Date         time.Time `json:"date"`
	VehicleNo    string    `json:"vehicle_no"`
	TotalAmount  int       `json:"total_amount"`
	TotalCount   int       `json:"total_count"`
	TotalDistance float64  `json:"total_distance"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Code    int    `json:"code" example:"400"`
	Message string `json:"message" example:"Invalid request parameters"`
}
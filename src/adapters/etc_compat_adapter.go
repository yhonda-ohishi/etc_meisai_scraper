package adapters

import (
	"time"

	"github.com/yhonda-ohishi/etc_meisai/src/models"
)

// ETCMeisaiCompatAdapter provides compatibility between new GORM model and legacy 38-field API
type ETCMeisaiCompatAdapter struct{}

// NewETCMeisaiCompatAdapter creates a new compatibility adapter
func NewETCMeisaiCompatAdapter() *ETCMeisaiCompatAdapter {
	return &ETCMeisaiCompatAdapter{}
}

// ToCompat converts GORM ETCMeisai to compatibility format
func (a *ETCMeisaiCompatAdapter) ToCompat(etc *models.ETCMeisai) *ETCMeisaiCompat {
	if etc == nil {
		return nil
	}

	compat := &ETCMeisaiCompat{
		// Core fields
		ID:        etc.ID,
		UseDate:   etc.UseDate,
		UseTime:   etc.UseTime,
		EntryIC:   etc.EntryIC,
		ExitIC:    etc.ExitIC,
		Amount:    etc.Amount,
		CarNumber: etc.CarNumber,
		ETCNumber: etc.ETCNumber,
		Hash:      etc.Hash,
		CreatedAt: etc.CreatedAt,
		UpdatedAt: etc.UpdatedAt,
	}

	// Legacy alias fields
	compat.ICEntry = &etc.EntryIC
	compat.ICExit = &etc.ExitIC
	usageDate := etc.UseDate.Format("2006-01-02")
	compat.UsageDate = &usageDate
	compat.UsageTime = &etc.UseTime
	compat.ETCCardNum = &etc.ETCNumber
	compat.VehicleNum = &etc.CarNumber
	compat.TollAmount = &etc.Amount

	// Additional computed fields
	date := etc.UseDate.Format("2006-01-02")
	compat.Date = &date
	time := etc.UseTime
	compat.Time = &time

	// Default values for missing legacy fields
	defaultStr := ""
	defaultInt := int32(0)
	defaultFloat := float64(0)

	compat.TollGate = &defaultStr
	compat.VehicleType = &defaultStr
	compat.DiscountAmount = &defaultInt
	compat.PaymentMethod = &defaultStr
	compat.RouteCode = &defaultStr
	compat.Distance = &defaultFloat
	compat.Remarks = &defaultStr
	compat.UsageType = &defaultStr
	compat.AccountType = &defaultStr
	compat.UnkoNo = &defaultStr

	// Set some computed values
	totalAmount := etc.Amount
	compat.TotalAmount = &totalAmount

	return compat
}

// FromCompat converts compatibility format to GORM ETCMeisai
func (a *ETCMeisaiCompatAdapter) FromCompat(compat *ETCMeisaiCompat) *models.ETCMeisai {
	if compat == nil {
		return nil
	}

	etc := &models.ETCMeisai{
		ID:        compat.ID,
		UseDate:   compat.UseDate,
		UseTime:   compat.UseTime,
		EntryIC:   compat.EntryIC,
		ExitIC:    compat.ExitIC,
		Amount:    compat.Amount,
		CarNumber: compat.CarNumber,
		ETCNumber: compat.ETCNumber,
		Hash:      compat.Hash,
		CreatedAt: compat.CreatedAt,
		UpdatedAt: compat.UpdatedAt,
	}

	// Handle legacy field mappings
	if compat.ICEntry != nil {
		etc.EntryIC = *compat.ICEntry
	}
	if compat.ICExit != nil {
		etc.ExitIC = *compat.ICExit
	}
	if compat.UsageDate != nil {
		if parsed, err := time.Parse("2006-01-02", *compat.UsageDate); err == nil {
			etc.UseDate = parsed
		}
	}
	if compat.UsageTime != nil {
		etc.UseTime = *compat.UsageTime
	}
	if compat.ETCCardNum != nil {
		etc.ETCNumber = *compat.ETCCardNum
	}
	if compat.VehicleNum != nil {
		etc.CarNumber = *compat.VehicleNum
	}
	if compat.TollAmount != nil {
		etc.Amount = *compat.TollAmount
	}

	// Handle alternative date/time fields
	if compat.Date != nil && etc.UseDate.IsZero() {
		if parsed, err := time.Parse("2006-01-02", *compat.Date); err == nil {
			etc.UseDate = parsed
		}
	}
	if compat.Time != nil && etc.UseTime == "" {
		etc.UseTime = *compat.Time
	}

	// Handle alternative amount fields
	if compat.TotalAmount != nil && etc.Amount == 0 {
		etc.Amount = *compat.TotalAmount
	}

	return etc
}

// ToCompatList converts a slice of ETCMeisai to compatibility format
func (a *ETCMeisaiCompatAdapter) ToCompatList(etcList []*models.ETCMeisai) []*ETCMeisaiCompat {
	if etcList == nil {
		return nil
	}

	compatList := make([]*ETCMeisaiCompat, len(etcList))
	for i, etc := range etcList {
		compatList[i] = a.ToCompat(etc)
	}
	return compatList
}

// FromCompatList converts a slice of compatibility format to ETCMeisai
func (a *ETCMeisaiCompatAdapter) FromCompatList(compatList []*ETCMeisaiCompat) []*models.ETCMeisai {
	if compatList == nil {
		return nil
	}

	etcList := make([]*models.ETCMeisai, len(compatList))
	for i, compat := range compatList {
		etcList[i] = a.FromCompat(compat)
	}
	return etcList
}

// ETCMeisaiCompat provides compatibility with existing 38-field API
type ETCMeisaiCompat struct {
	// Core GORM fields
	ID        int64     `json:"id"`
	UseDate   time.Time `json:"use_date"`
	UseTime   string    `json:"use_time"`
	EntryIC   string    `json:"entry_ic"`
	ExitIC    string    `json:"exit_ic"`
	Amount    int32     `json:"amount"`
	CarNumber string    `json:"car_number"`
	ETCNumber string    `json:"etc_number"`
	Hash      string    `json:"hash"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// Legacy alias fields (all pointers to allow omitempty)
	ICEntry        *string  `json:"ic_entry,omitempty"`
	ICExit         *string  `json:"ic_exit,omitempty"`
	Date           *string  `json:"date,omitempty"`
	Time           *string  `json:"time,omitempty"`
	UsageDate      *string  `json:"usage_date,omitempty"`
	UsageTime      *string  `json:"usage_time,omitempty"`
	ETCCardNum     *string  `json:"etc_card_num,omitempty"`
	ETCCardNumber  *string  `json:"etc_card_number,omitempty"`
	VehicleNum     *string  `json:"vehicle_num,omitempty"`
	VehicleNumber  *string  `json:"vehicle_number,omitempty"`
	VehicleNo      *string  `json:"vehicle_no,omitempty"`
	TollAmount     *int32   `json:"toll_amount,omitempty"`
	TotalAmount    *int32   `json:"total_amount,omitempty"`
	CardNo         *string  `json:"card_no,omitempty"`
	CardNumber     *string  `json:"card_number,omitempty"`
	ETCNum         *string  `json:"etc_num,omitempty"`

	// Additional legacy fields that may be needed
	TollGate       *string  `json:"toll_gate,omitempty"`
	VehicleType    *string  `json:"vehicle_type,omitempty"`
	DiscountAmount *int32   `json:"discount_amount,omitempty"`
	PaymentMethod  *string  `json:"payment_method,omitempty"`
	RouteCode      *string  `json:"route_code,omitempty"`
	Distance       *float64 `json:"distance,omitempty"`
	Remarks        *string  `json:"remarks,omitempty"`
	UsageType      *string  `json:"usage_type,omitempty"`
	AccountType    *string  `json:"account_type,omitempty"`
	UnkoNo         *string  `json:"unko_no,omitempty"`
	RowID          *int64   `json:"row_id,omitempty"`
	ImportedAt     *time.Time `json:"imported_at,omitempty"`

	// Runtime computed fields
	RouteType      *string `json:"route_type,omitempty"`
	DiscountType   *string `json:"discount_type,omitempty"`
	TransType      *string `json:"trans_type,omitempty"`
}

// GetActualAmount returns the actual amount from various possible fields
func (c *ETCMeisaiCompat) GetActualAmount() int32 {
	if c.Amount != 0 {
		return c.Amount
	}
	if c.TollAmount != nil && *c.TollAmount != 0 {
		return *c.TollAmount
	}
	if c.TotalAmount != nil && *c.TotalAmount != 0 {
		return *c.TotalAmount
	}
	return 0
}

// GetActualETCNumber returns the actual ETC number from various possible fields
func (c *ETCMeisaiCompat) GetActualETCNumber() string {
	if c.ETCNumber != "" {
		return c.ETCNumber
	}
	if c.ETCCardNum != nil && *c.ETCCardNum != "" {
		return *c.ETCCardNum
	}
	if c.ETCCardNumber != nil && *c.ETCCardNumber != "" {
		return *c.ETCCardNumber
	}
	if c.ETCNum != nil && *c.ETCNum != "" {
		return *c.ETCNum
	}
	if c.CardNo != nil && *c.CardNo != "" {
		return *c.CardNo
	}
	if c.CardNumber != nil && *c.CardNumber != "" {
		return *c.CardNumber
	}
	return ""
}

// GetActualVehicleNumber returns the actual vehicle number from various possible fields
func (c *ETCMeisaiCompat) GetActualVehicleNumber() string {
	if c.CarNumber != "" {
		return c.CarNumber
	}
	if c.VehicleNum != nil && *c.VehicleNum != "" {
		return *c.VehicleNum
	}
	if c.VehicleNumber != nil && *c.VehicleNumber != "" {
		return *c.VehicleNumber
	}
	if c.VehicleNo != nil && *c.VehicleNo != "" {
		return *c.VehicleNo
	}
	return ""
}

// GetActualEntryIC returns the actual entry IC from various possible fields
func (c *ETCMeisaiCompat) GetActualEntryIC() string {
	if c.EntryIC != "" {
		return c.EntryIC
	}
	if c.ICEntry != nil && *c.ICEntry != "" {
		return *c.ICEntry
	}
	return ""
}

// GetActualExitIC returns the actual exit IC from various possible fields
func (c *ETCMeisaiCompat) GetActualExitIC() string {
	if c.ExitIC != "" {
		return c.ExitIC
	}
	if c.ICExit != nil && *c.ICExit != "" {
		return *c.ICExit
	}
	return ""
}

// ConvertToStandardFormat normalizes a compatibility record to use standard fields
func (a *ETCMeisaiCompatAdapter) ConvertToStandardFormat(compat *ETCMeisaiCompat) *ETCMeisaiCompat {
	if compat == nil {
		return nil
	}

	// Create a copy and normalize to standard fields
	normalized := *compat

	// Normalize amount
	normalized.Amount = compat.GetActualAmount()

	// Normalize ETC number
	normalized.ETCNumber = compat.GetActualETCNumber()

	// Normalize vehicle number
	normalized.CarNumber = compat.GetActualVehicleNumber()

	// Normalize IC fields
	normalized.EntryIC = compat.GetActualEntryIC()
	normalized.ExitIC = compat.GetActualExitIC()

	// Handle date fields
	if compat.Date != nil && normalized.UseDate.IsZero() {
		if parsed, err := time.Parse("2006-01-02", *compat.Date); err == nil {
			normalized.UseDate = parsed
		}
	}
	if compat.UsageDate != nil && normalized.UseDate.IsZero() {
		if parsed, err := time.Parse("2006-01-02", *compat.UsageDate); err == nil {
			normalized.UseDate = parsed
		}
	}

	// Handle time fields
	if compat.Time != nil && normalized.UseTime == "" {
		normalized.UseTime = *compat.Time
	}
	if compat.UsageTime != nil && normalized.UseTime == "" {
		normalized.UseTime = *compat.UsageTime
	}

	return &normalized
}
package adapters

import (
	"fmt"
	"strconv"
	"time"

	"github.com/yhonda-ohishi/etc_meisai/src/models"
	"github.com/yhonda-ohishi/etc_meisai/src/pb"
	"google.golang.org/protobuf/types/known/timestamppb"
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

// ConvertToProto converts ETCMeisaiCompat to protobuf format
func (a *ETCMeisaiCompatAdapter) ConvertToProto(compat *ETCMeisaiCompat) (*pb.ETCMeisaiRecord, error) {
	if compat == nil {
		return nil, fmt.Errorf("compat cannot be nil")
	}

	// Normalize the compatibility record first
	normalized := a.ConvertToStandardFormat(compat)
	if normalized == nil {
		return nil, fmt.Errorf("failed to normalize compatibility record")
	}

	proto := &pb.ETCMeisaiRecord{
		Id:            normalized.ID,
		Hash:          normalized.Hash,
		Date:          normalized.UseDate.Format("2006-01-02"),
		Time:          normalized.UseTime,
		EntranceIc:    normalized.EntryIC,
		ExitIc:        normalized.ExitIC,
		TollAmount:    normalized.Amount,
		CarNumber:     normalized.CarNumber,
		EtcCardNumber: normalized.ETCNumber,
	}

	// Convert timestamps
	if !normalized.CreatedAt.IsZero() {
		proto.CreatedAt = timestamppb.New(normalized.CreatedAt)
	}
	if !normalized.UpdatedAt.IsZero() {
		proto.UpdatedAt = timestamppb.New(normalized.UpdatedAt)
	}

	return proto, nil
}

// ConvertToProtoList converts a slice of ETCMeisaiCompat to protobuf format
func (a *ETCMeisaiCompatAdapter) ConvertToProtoList(compatList []*ETCMeisaiCompat) ([]*pb.ETCMeisaiRecord, error) {
	if compatList == nil {
		return nil, nil
	}

	protoList := make([]*pb.ETCMeisaiRecord, 0, len(compatList))
	for i, compat := range compatList {
		if compat == nil {
			return nil, fmt.Errorf("compat at index %d cannot be nil", i)
		}

		proto, err := a.ConvertToProto(compat)
		if err != nil {
			return nil, fmt.Errorf("error converting compat at index %d: %w", i, err)
		}

		protoList = append(protoList, proto)
	}

	return protoList, nil
}

// ConvertFromLegacy converts legacy data map to ETCMeisaiRecord
func (a *ETCMeisaiCompatAdapter) ConvertFromLegacy(legacyData map[string]interface{}) (*models.ETCMeisaiRecord, error) {
	if legacyData == nil {
		return nil, fmt.Errorf("legacy data cannot be nil")
	}

	if len(legacyData) == 0 {
		return nil, fmt.Errorf("legacy data is empty")
	}

	record := &models.ETCMeisaiRecord{}

	// Required fields validation
	requiredFields := []string{"use_date", "use_time", "entry_ic", "exit_ic", "amount", "car_number", "etc_number"}
	for _, field := range requiredFields {
		if _, exists := legacyData[field]; !exists {
			// Check for alternative field names (Japanese)
			alternativeFields := map[string][]string{
				"use_date":   {"利用年月日", "date"},
				"use_time":   {"利用時刻", "time"},
				"entry_ic":   {"入口IC", "入口"},
				"exit_ic":    {"出口IC", "出口"},
				"amount":     {"料金", "toll_amount"},
				"car_number": {"車両番号", "vehicle_number"},
				"etc_number": {"ETCカード番号", "etc_card_number"},
			}

			found := false
			if alternatives, ok := alternativeFields[field]; ok {
				for _, alt := range alternatives {
					if _, exists := legacyData[alt]; exists {
						found = true
						break
					}
				}
			}

			if !found {
				return nil, fmt.Errorf("missing required field: %s", field)
			}
		}
	}

	// Date parsing
	dateStr := getStringValue(legacyData, "use_date", "利用年月日", "date")
	if dateStr != "" {
		if parsedDate, err := time.Parse("2006-01-02", dateStr); err == nil {
			record.Date = parsedDate
		} else {
			return nil, fmt.Errorf("invalid date format")
		}
	}

	// Time
	record.Time = getStringValue(legacyData, "use_time", "利用時刻", "time")

	// IC fields
	record.EntranceIC = getStringValue(legacyData, "entry_ic", "入口IC", "入口")
	record.ExitIC = getStringValue(legacyData, "exit_ic", "出口IC", "出口")

	// Amount
	amountVal := getValue(legacyData, "amount", "料金", "toll_amount")
	if amountVal != nil {
		if amount, ok := amountVal.(int); ok {
			record.TollAmount = amount
		} else if amountStr, ok := amountVal.(string); ok {
			if parsed, err := strconv.Atoi(amountStr); err == nil {
				record.TollAmount = parsed
			} else {
				return nil, fmt.Errorf("invalid amount format")
			}
		} else {
			return nil, fmt.Errorf("invalid amount format")
		}
	}

	// Car number
	record.CarNumber = getStringValue(legacyData, "car_number", "車両番号", "vehicle_number")

	// ETC number
	record.ETCCardNumber = getStringValue(legacyData, "etc_number", "ETCカード番号", "etc_card_number")

	return record, nil
}

// ConvertToLegacy converts ETCMeisaiRecord to legacy format
func (a *ETCMeisaiCompatAdapter) ConvertToLegacy(record *models.ETCMeisaiRecord, format string) (map[string]interface{}, error) {
	if record == nil {
		return nil, fmt.Errorf("record cannot be nil")
	}

	switch format {
	case "legacy":
		return map[string]interface{}{
			"use_date":   record.Date.Format("2006-01-02"),
			"use_time":   record.Time,
			"entry_ic":   record.EntranceIC,
			"exit_ic":    record.ExitIC,
			"amount":     record.TollAmount,
			"car_number": record.CarNumber,
			"etc_number": record.ETCCardNumber,
		}, nil
	case "japanese":
		return map[string]interface{}{
			"利用年月日":    record.Date.Format("2006-01-02"),
			"利用時刻":     record.Time,
			"入口IC":      record.EntranceIC,
			"出口IC":      record.ExitIC,
			"料金":       record.TollAmount,
			"車両番号":     record.CarNumber,
			"ETCカード番号": record.ETCCardNumber,
		}, nil
	default:
		return nil, fmt.Errorf("unsupported format: %s", format)
	}
}

// ConvertBatch converts multiple legacy data maps to ETCMeisaiRecord slice
func (a *ETCMeisaiCompatAdapter) ConvertBatch(legacyBatch []map[string]interface{}) ([]*models.ETCMeisaiRecord, error) {
	if legacyBatch == nil {
		return nil, fmt.Errorf("batch cannot be nil")
	}

	if len(legacyBatch) == 0 {
		return nil, fmt.Errorf("batch cannot be empty")
	}

	records := make([]*models.ETCMeisaiRecord, 0, len(legacyBatch))
	for i, legacyData := range legacyBatch {
		record, err := a.ConvertFromLegacy(legacyData)
		if err != nil {
			return nil, fmt.Errorf("failed to convert record at index %d: %w", i, err)
		}
		records = append(records, record)
	}

	return records, nil
}

// ValidateCompatibility validates if legacy data is compatible
func (a *ETCMeisaiCompatAdapter) ValidateCompatibility(legacyData map[string]interface{}) error {
	if legacyData == nil {
		return fmt.Errorf("legacy data cannot be nil")
	}

	// Check for incompatible version
	if version, exists := legacyData["version"]; exists {
		if versionStr, ok := version.(string); ok && versionStr == "v1.0.0" {
			return fmt.Errorf("incompatible version: %s", versionStr)
		}
	}

	// Check for critical fields
	criticalFields := []string{"use_date", "entry_ic", "exit_ic"}
	missingFields := []string{}

	for _, field := range criticalFields {
		if _, exists := legacyData[field]; !exists {
			// Check alternative names
			alternatives := map[string][]string{
				"use_date": {"利用年月日", "date"},
				"entry_ic": {"入口IC", "入口"},
				"exit_ic":  {"出口IC", "出口"},
			}

			found := false
			if alts, ok := alternatives[field]; ok {
				for _, alt := range alts {
					if _, exists := legacyData[alt]; exists {
						found = true
						break
					}
				}
			}

			if !found {
				missingFields = append(missingFields, field)
			}
		}
	}

	if len(missingFields) > 0 {
		return fmt.Errorf("missing critical fields: %v", missingFields)
	}

	return nil
}

// GetFieldMapping returns field mapping for the specified format
func (a *ETCMeisaiCompatAdapter) GetFieldMapping(format string) (map[string]string, error) {
	switch format {
	case "legacy":
		return map[string]string{
			"use_date":   "Date",
			"use_time":   "Time",
			"entry_ic":   "EntranceIC",
			"exit_ic":    "ExitIC",
			"amount":     "TollAmount",
			"car_number": "CarNumber",
			"etc_number": "ETCCardNumber",
		}, nil
	case "japanese":
		return map[string]string{
			"利用年月日":    "Date",
			"利用時刻":     "Time",
			"入口IC":      "EntranceIC",
			"出口IC":      "ExitIC",
			"料金":       "TollAmount",
			"車両番号":     "CarNumber",
			"ETCカード番号": "ETCCardNumber",
		}, nil
	default:
		return nil, fmt.Errorf("unsupported format: %s", format)
	}
}

// NormalizeFieldNames normalizes field names to standard format
func (a *ETCMeisaiCompatAdapter) NormalizeFieldNames(input map[string]interface{}) map[string]interface{} {
	if input == nil {
		return make(map[string]interface{})
	}

	result := make(map[string]interface{})

	// Field mapping from various formats to standard format
	fieldMap := map[string]string{
		// Date fields
		"date":        "use_date",
		"利用日":        "use_date",
		"利用年月日":      "use_date",

		// Time fields
		"time":        "use_time",
		"利用時間":       "use_time",
		"利用時刻":       "use_time",

		// IC fields
		"entry":       "entry_ic",
		"入口":         "entry_ic",
		"exit":        "exit_ic",
		"出口":         "exit_ic",

		// Amount fields
		"toll_amount": "amount",
		"料金":         "amount",

		// Vehicle fields
		"vehicle_number": "car_number",
		"車両番号":         "car_number",

		// ETC fields
		"etc_card_number": "etc_number",
		"ETCカード番号":       "etc_number",
	}

	for key, value := range input {
		normalizedKey := key
		if mapped, exists := fieldMap[key]; exists {
			normalizedKey = mapped
		}
		result[normalizedKey] = value
	}

	return result
}

// Helper function to get string value from multiple possible keys
func getStringValue(data map[string]interface{}, keys ...string) string {
	for _, key := range keys {
		if value, exists := data[key]; exists {
			if str, ok := value.(string); ok {
				return str
			}
		}
	}
	return ""
}

// Helper function to get value from multiple possible keys
func getValue(data map[string]interface{}, keys ...string) interface{} {
	for _, key := range keys {
		if value, exists := data[key]; exists {
			return value
		}
	}
	return nil
}
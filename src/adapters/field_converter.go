package adapters

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
)

// FieldConverter provides utilities for converting between different field formats
type FieldConverter struct{}

// NewFieldConverter creates a new field converter
func NewFieldConverter() *FieldConverter {
	return &FieldConverter{}
}

// ConvertStringToInt32 safely converts string to int32
func (fc *FieldConverter) ConvertStringToInt32(s string) (int32, error) {
	if s == "" {
		return 0, nil
	}

	// Remove common formatting characters
	s = strings.ReplaceAll(s, ",", "")
	s = strings.ReplaceAll(s, "¥", "")
	s = strings.ReplaceAll(s, "円", "")
	s = strings.TrimSpace(s)

	i, err := strconv.ParseInt(s, 10, 32)
	if err != nil {
		return 0, fmt.Errorf("cannot convert '%s' to int32: %w", s, err)
	}

	return int32(i), nil
}

// ConvertStringToFloat64 safely converts string to float64
func (fc *FieldConverter) ConvertStringToFloat64(s string) (float64, error) {
	if s == "" {
		return 0, nil
	}

	// Remove common formatting characters
	s = strings.ReplaceAll(s, ",", "")
	s = strings.TrimSpace(s)

	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, fmt.Errorf("cannot convert '%s' to float64: %w", s, err)
	}

	return f, nil
}

// ConvertStringToTime converts string to time with multiple format attempts
func (fc *FieldConverter) ConvertStringToTime(s string) (time.Time, error) {
	if s == "" {
		return time.Time{}, nil
	}

	// Try multiple date formats commonly used in CSV files
	formats := []string{
		"2006-01-02",
		"2006/01/02",
		"01/02/2006",
		"02/01/2006",
		"2006-01-02 15:04:05",
		"2006/01/02 15:04:05",
		"01/02/2006 15:04:05",
		"2006年01月02日",
		"平成18年01月02日", // Handle Japanese era years if needed
	}

	for _, format := range formats {
		if t, err := time.Parse(format, s); err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("cannot parse time '%s' with any known format", s)
}

// ConvertTimeToString converts time to standard string format
func (fc *FieldConverter) ConvertTimeToString(t time.Time, format string) string {
	if t.IsZero() {
		return ""
	}

	if format == "" {
		format = "2006-01-02"
	}

	return t.Format(format)
}

// NormalizeTimeString normalizes time string to HH:MM format
func (fc *FieldConverter) NormalizeTimeString(s string) (string, error) {
	if s == "" {
		return "", nil
	}

	// Remove common formatting
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, "：", ":")

	// Try to parse as time
	formats := []string{
		"15:04",
		"15:04:05",
		"3:04PM",
		"3:04:05PM",
		"15時04分",
	}

	for _, format := range formats {
		if t, err := time.Parse(format, s); err == nil {
			return t.Format("15:04"), nil
		}
	}

	// If no format matches, check if it's already in HH:MM format
	if matched := isValidTimeFormat(s); matched {
		return s, nil
	}

	return "", fmt.Errorf("cannot normalize time string '%s'", s)
}

// NormalizeICName normalizes IC (interchange) names
func (fc *FieldConverter) NormalizeICName(s string) string {
	if s == "" {
		return ""
	}

	// Trim whitespace
	s = strings.TrimSpace(s)

	// Normalize common suffixes
	suffixes := []string{"IC", "インター", "ｲﾝﾀｰ", "料金所"}
	for _, suffix := range suffixes {
		if strings.HasSuffix(s, suffix) && !strings.HasSuffix(s, "IC") {
			s = strings.TrimSuffix(s, suffix) + "IC"
			break
		}
	}

	// Ensure it ends with IC if it doesn't already
	if !strings.HasSuffix(s, "IC") && s != "" {
		s += "IC"
	}

	return s
}

// NormalizeVehicleNumber normalizes vehicle numbers
func (fc *FieldConverter) NormalizeVehicleNumber(s string) string {
	if s == "" {
		return ""
	}

	// Trim whitespace
	s = strings.TrimSpace(s)

	// Convert full-width characters to half-width
	s = fc.convertFullWidthToHalfWidth(s)

	// Remove common prefixes/suffixes
	prefixes := []string{"車両番号", "車番", "No.", "No", "#"}
	for _, prefix := range prefixes {
		s = strings.TrimPrefix(s, prefix)
	}

	s = strings.TrimSpace(s)

	return s
}

// NormalizeETCNumber normalizes ETC card numbers
func (fc *FieldConverter) NormalizeETCNumber(s string) string {
	if s == "" {
		return ""
	}

	// Trim whitespace
	s = strings.TrimSpace(s)

	// Convert full-width to half-width
	s = fc.convertFullWidthToHalfWidth(s)

	// Remove non-numeric characters
	var result strings.Builder
	for _, r := range s {
		if r >= '0' && r <= '9' {
			result.WriteRune(r)
		}
	}

	return result.String()
}

// convertFullWidthToHalfWidth converts full-width characters to half-width
func (fc *FieldConverter) convertFullWidthToHalfWidth(s string) string {
	// Mapping of full-width to half-width characters
	fullToHalf := map[rune]rune{
		'０': '0', '１': '1', '２': '2', '３': '3', '４': '4',
		'５': '5', '６': '6', '７': '7', '８': '8', '９': '9',
		'－': '-', '：': ':', '／': '/',
	}

	var result strings.Builder
	for _, r := range s {
		if half, ok := fullToHalf[r]; ok {
			result.WriteRune(half)
		} else {
			result.WriteRune(r)
		}
	}

	return result.String()
}

// MapLegacyFields maps legacy field names to standard field names
func (fc *FieldConverter) MapLegacyFields(data map[string]interface{}) map[string]interface{} {
	// Field mapping from legacy names to standard names
	fieldMap := map[string]string{
		// Date fields
		"date":        "use_date",
		"usage_date":  "use_date",
		"利用日":        "use_date",
		"使用日":        "use_date",

		// Time fields
		"time":        "use_time",
		"usage_time":  "use_time",
		"利用時間":       "use_time",
		"使用時間":       "use_time",

		// IC fields
		"ic_entry":    "entry_ic",
		"entry":       "entry_ic",
		"入口":         "entry_ic",
		"入口IC":       "entry_ic",
		"ic_exit":     "exit_ic",
		"exit":        "exit_ic",
		"出口":         "exit_ic",
		"出口IC":       "exit_ic",

		// Amount fields
		"toll_amount":   "amount",
		"total_amount":  "amount",
		"料金":          "amount",
		"通行料金":        "amount",

		// Vehicle fields
		"vehicle_num":     "car_number",
		"vehicle_number":  "car_number",
		"vehicle_no":      "car_number",
		"車両番号":          "car_number",
		"車番":            "car_number",

		// ETC number fields
		"etc_card_num":    "etc_number",
		"etc_card_number": "etc_number",
		"card_no":         "etc_number",
		"card_number":     "etc_number",
		"etc_num":         "etc_number",
		"ETCカード番号":       "etc_number",
	}

	result := make(map[string]interface{})

	// Copy all fields, mapping legacy names to standard names
	for key, value := range data {
		standardKey := key
		if mapped, ok := fieldMap[strings.ToLower(key)]; ok {
			standardKey = mapped
		}
		result[standardKey] = value
	}

	return result
}

// ConvertFieldValue converts a field value to the appropriate type
func (fc *FieldConverter) ConvertFieldValue(value interface{}, targetType string) (interface{}, error) {
	if value == nil {
		return nil, nil
	}

	stringValue := fmt.Sprintf("%v", value)
	stringValue = strings.TrimSpace(stringValue)

	switch targetType {
	case "int32":
		return fc.ConvertStringToInt32(stringValue)
	case "int64":
		i, err := strconv.ParseInt(stringValue, 10, 64)
		return i, err
	case "float64":
		return fc.ConvertStringToFloat64(stringValue)
	case "time":
		return fc.ConvertStringToTime(stringValue)
	case "string":
		return stringValue, nil
	case "bool":
		b, err := strconv.ParseBool(stringValue)
		return b, err
	default:
		return value, nil
	}
}

// ValidateFieldType checks if a value can be converted to the target type
func (fc *FieldConverter) ValidateFieldType(value interface{}, targetType string) error {
	_, err := fc.ConvertFieldValue(value, targetType)
	return err
}

// ConvertStructFields converts fields in a struct using reflection
func (fc *FieldConverter) ConvertStructFields(src interface{}, dst interface{}) error {
	srcVal := reflect.ValueOf(src)
	dstVal := reflect.ValueOf(dst)

	if srcVal.Kind() == reflect.Ptr {
		srcVal = srcVal.Elem()
	}
	if dstVal.Kind() == reflect.Ptr {
		dstVal = dstVal.Elem()
	}

	if !dstVal.CanSet() {
		return fmt.Errorf("destination struct is not settable")
	}

	srcType := srcVal.Type()
	dstType := dstVal.Type()

	for i := 0; i < srcType.NumField(); i++ {
		srcField := srcType.Field(i)
		srcFieldValue := srcVal.Field(i)

		// Find corresponding field in destination
		dstField, found := dstType.FieldByName(srcField.Name)
		if !found {
			continue
		}

		dstFieldValue := dstVal.FieldByName(srcField.Name)
		if !dstFieldValue.CanSet() {
			continue
		}

		// Convert value if types are different
		if srcField.Type != dstField.Type {
			converted, err := fc.convertValue(srcFieldValue.Interface(), dstField.Type)
			if err != nil {
				return fmt.Errorf("error converting field %s: %w", srcField.Name, err)
			}
			dstFieldValue.Set(reflect.ValueOf(converted))
		} else {
			dstFieldValue.Set(srcFieldValue)
		}
	}

	return nil
}

// convertValue converts a value to the target type using reflection
func (fc *FieldConverter) convertValue(value interface{}, targetType reflect.Type) (interface{}, error) {
	if value == nil {
		return reflect.Zero(targetType).Interface(), nil
	}

	sourceValue := reflect.ValueOf(value)
	if sourceValue.Type() == targetType {
		return value, nil
	}

	// Handle string conversions
	if sourceValue.Kind() == reflect.String {
		stringVal := sourceValue.String()
		switch targetType.Kind() {
		case reflect.Int32:
			i, err := fc.ConvertStringToInt32(stringVal)
			return i, err
		case reflect.Int64:
			i, err := strconv.ParseInt(stringVal, 10, 64)
			return i, err
		case reflect.Float64:
			f, err := fc.ConvertStringToFloat64(stringVal)
			return f, err
		}
	}

	// Handle time conversions
	if targetType == reflect.TypeOf(time.Time{}) && sourceValue.Kind() == reflect.String {
		t, err := fc.ConvertStringToTime(sourceValue.String())
		return t, err
	}

	// Default: try direct conversion
	if sourceValue.Type().ConvertibleTo(targetType) {
		return sourceValue.Convert(targetType).Interface(), nil
	}

	return nil, fmt.Errorf("cannot convert %v to %v", sourceValue.Type(), targetType)
}

// isValidTimeFormat checks if time string is in HH:MM format (copied from validation.go)
func isValidTimeFormat(timeStr string) bool {
	// Simple regex check for HH:MM format
	if len(timeStr) != 5 {
		return false
	}
	if timeStr[2] != ':' {
		return false
	}
	for i, r := range timeStr {
		if i == 2 {
			continue // skip the ':'
		}
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}
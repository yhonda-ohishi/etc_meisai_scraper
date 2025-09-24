package adapters_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yhonda-ohishi/etc_meisai/src/adapters"
)

// TestFieldConverter_NewFieldConverter tests field converter creation
func TestFieldConverter_NewFieldConverter(t *testing.T) {
	converter := adapters.NewFieldConverter()
	assert.NotNil(t, converter)
}

// TestFieldConverter_ConvertStringToInt32 tests string to int32 conversion
func TestFieldConverter_ConvertStringToInt32(t *testing.T) {
	converter := adapters.NewFieldConverter()

	tests := []struct {
		name     string
		input    string
		expected int32
		wantErr  bool
	}{
		{"empty string", "", 0, false},
		{"simple number", "123", 123, false},
		{"negative number", "-456", -456, false},
		{"zero", "0", 0, false},
		{"with comma", "1,234", 1234, false},
		{"with yen symbol", "¥500", 500, false},
		{"with en kanji", "1000円", 1000, false},
		{"with whitespace", "  789  ", 789, false},
		{"complex formatting", "¥1,234円 ", 1234, false},
		{"max int32", "2147483647", 2147483647, false},
		{"min int32", "-2147483648", -2147483648, false},
		{"overflow", "2147483648", 0, true},
		{"underflow", "-2147483649", 0, true},
		{"invalid string", "abc", 0, true},
		{"float string", "123.45", 0, true},
		{"mixed characters", "12a34", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := converter.ConvertStringToInt32(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "cannot convert")
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

// TestFieldConverter_ConvertStringToFloat64 tests string to float64 conversion
func TestFieldConverter_ConvertStringToFloat64(t *testing.T) {
	converter := adapters.NewFieldConverter()

	tests := []struct {
		name     string
		input    string
		expected float64
		wantErr  bool
	}{
		{"empty string", "", 0, false},
		{"simple integer", "123", 123.0, false},
		{"simple float", "123.45", 123.45, false},
		{"negative float", "-456.78", -456.78, false},
		{"zero", "0", 0.0, false},
		{"zero float", "0.0", 0.0, false},
		{"with comma", "1,234.56", 1234.56, false},
		{"with whitespace", "  789.12  ", 789.12, false},
		{"scientific notation", "1.23e2", 123.0, false},
		{"scientific notation negative", "1.23e-2", 0.0123, false},
		{"very small number", "0.00001", 0.00001, false},
		{"very large number", "1234567890.123", 1234567890.123, false},
		{"invalid string", "abc", 0, true},
		{"mixed characters", "12.3a4", 0, true},
		{"multiple dots", "12.34.56", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := converter.ConvertStringToFloat64(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "cannot convert")
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

// TestFieldConverter_ConvertStringToTime tests string to time conversion
func TestFieldConverter_ConvertStringToTime(t *testing.T) {
	converter := adapters.NewFieldConverter()

	tests := []struct {
		name     string
		input    string
		expected time.Time
		wantErr  bool
	}{
		{"empty string", "", time.Time{}, false},
		{"ISO date", "2023-12-25", time.Date(2023, 12, 25, 0, 0, 0, 0, time.UTC), false},
		{"slash date", "2023/12/25", time.Date(2023, 12, 25, 0, 0, 0, 0, time.UTC), false},
		{"US format", "12/25/2023", time.Date(2023, 12, 25, 0, 0, 0, 0, time.UTC), false},
		{"UK format", "25/12/2023", time.Date(2023, 12, 25, 0, 0, 0, 0, time.UTC), false},
		{"with time ISO", "2023-12-25 14:30:00", time.Date(2023, 12, 25, 14, 30, 0, 0, time.UTC), false},
		{"with time slash", "2023/12/25 14:30:00", time.Date(2023, 12, 25, 14, 30, 0, 0, time.UTC), false},
		{"with time US", "12/25/2023 14:30:00", time.Date(2023, 12, 25, 14, 30, 0, 0, time.UTC), false},
		{"Japanese format", "2023年12月25日", time.Date(2023, 12, 25, 0, 0, 0, 0, time.UTC), false},
		{"invalid format", "invalid-date", time.Time{}, true},
		{"partial date", "2023-12", time.Time{}, true},
		{"invalid characters", "abcd-ef-gh", time.Time{}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := converter.ConvertStringToTime(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "cannot parse time")
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

// TestFieldConverter_ConvertTimeToString tests time to string conversion
func TestFieldConverter_ConvertTimeToString(t *testing.T) {
	converter := adapters.NewFieldConverter()

	tests := []struct {
		name     string
		time     time.Time
		format   string
		expected string
	}{
		{"zero time", time.Time{}, "", ""},
		{"zero time with format", time.Time{}, "2006-01-02", ""},
		{"standard format", time.Date(2023, 12, 25, 14, 30, 0, 0, time.UTC), "", "2023-12-25"},
		{"custom format", time.Date(2023, 12, 25, 14, 30, 0, 0, time.UTC), "2006/01/02", "2023/12/25"},
		{"with time", time.Date(2023, 12, 25, 14, 30, 0, 0, time.UTC), "2006-01-02 15:04:05", "2023-12-25 14:30:00"},
		{"time only", time.Date(2023, 12, 25, 14, 30, 0, 0, time.UTC), "15:04", "14:30"},
		{"Japanese format", time.Date(2023, 12, 25, 14, 30, 0, 0, time.UTC), "2006年01月02日", "2023年12月25日"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := converter.ConvertTimeToString(tt.time, tt.format)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestFieldConverter_NormalizeTimeString tests time string normalization
func TestFieldConverter_NormalizeTimeString(t *testing.T) {
	converter := adapters.NewFieldConverter()

	tests := []struct {
		name     string
		input    string
		expected string
		wantErr  bool
	}{
		{"empty string", "", "", false},
		{"standard format", "14:30", "14:30", false},
		{"with seconds", "14:30:45", "14:30", false},
		{"12-hour format", "2:30PM", "14:30", false},
		{"12-hour with seconds", "2:30:45PM", "14:30", false},
		{"Japanese format", "14時30分", "14:30", false},
		{"full-width colon", "14：30", "14:30", false},
		{"with whitespace", "  14:30  ", "14:30", false},
		{"already normalized", "09:15", "09:15", false},
		{"midnight", "00:00", "00:00", false},
		{"noon", "12:00", "12:00", false},
		{"12AM", "12:00AM", "00:00", false},
		{"invalid format", "25:30", "", true},
		{"invalid characters", "ab:cd", "", true},
		{"no colon", "1430", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := converter.NormalizeTimeString(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "cannot normalize time string")
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

// TestFieldConverter_NormalizeICName tests IC name normalization
func TestFieldConverter_NormalizeICName(t *testing.T) {
	converter := adapters.NewFieldConverter()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"empty string", "", ""},
		{"already normalized", "東京IC", "東京IC"},
		{"with インター suffix", "東京インター", "東京IC"},
		{"with katakana suffix", "東京ｲﾝﾀｰ", "東京IC"},
		{"with 料金所 suffix", "東京料金所", "東京IC"},
		{"with whitespace", "  東京  ", "東京IC"},
		{"without IC suffix", "東京", "東京IC"},
		{"mixed case", "tokyo", "tokyoIC"},
		{"number in name", "東京1", "東京1IC"},
		{"hyphenated name", "東京-南", "東京-南IC"},
		{"already has IC", "東京IC", "東京IC"},
		{"complex name", "  東京インター  ", "東京IC"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := converter.NormalizeICName(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestFieldConverter_NormalizeVehicleNumber tests vehicle number normalization
func TestFieldConverter_NormalizeVehicleNumber(t *testing.T) {
	converter := adapters.NewFieldConverter()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"empty string", "", ""},
		{"simple number", "123", "123"},
		{"with prefix", "車両番号123", "123"},
		{"with 車番 prefix", "車番456", "456"},
		{"with No. prefix", "No.789", "789"},
		{"with # prefix", "#101", "101"},
		{"with whitespace", "  123  ", "123"},
		{"full-width numbers", "１２３", "123"},
		{"mixed width", "車両番号１２３", "123"},
		{"alphanumeric", "ABC123", "ABC123"},
		{"with hyphen", "A-123", "A-123"},
		{"complex format", "  車両番号１２３  ", "123"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := converter.NormalizeVehicleNumber(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestFieldConverter_NormalizeETCNumber tests ETC number normalization
func TestFieldConverter_NormalizeETCNumber(t *testing.T) {
	converter := adapters.NewFieldConverter()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"empty string", "", ""},
		{"simple number", "1234567890", "1234567890"},
		{"with hyphens", "1234-5678-90", "1234567890"},
		{"with spaces", "1234 5678 90", "1234567890"},
		{"full-width numbers", "１２３４５６７８９０", "1234567890"},
		{"mixed characters", "1234-ABCD-5678", "12345678"},
		{"with whitespace", "  1234567890  ", "1234567890"},
		{"alphanumeric", "1A2B3C4D", "1234"},
		{"only letters", "ABCD", ""},
		{"mixed format", "１２３４-5678", "12345678"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := converter.NormalizeETCNumber(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestFieldConverter_MapLegacyFields tests legacy field mapping
func TestFieldConverter_MapLegacyFields(t *testing.T) {
	converter := adapters.NewFieldConverter()

	tests := []struct {
		name     string
		input    map[string]interface{}
		expected map[string]interface{}
	}{
		{
			name: "empty map",
			input: map[string]interface{}{},
			expected: map[string]interface{}{},
		},
		{
			name: "standard fields",
			input: map[string]interface{}{
				"use_date": "2023-12-25",
				"use_time": "14:30",
				"amount":   500,
			},
			expected: map[string]interface{}{
				"use_date": "2023-12-25",
				"use_time": "14:30",
				"amount":   500,
			},
		},
		{
			name: "legacy date fields",
			input: map[string]interface{}{
				"date":       "2023-12-25",
				"usage_date": "2023-12-26",
				"利用日":        "2023-12-27",
			},
			expected: map[string]interface{}{
				"use_date": "2023-12-27",
			},
		},
		{
			name: "legacy time fields",
			input: map[string]interface{}{
				"time":       "14:30",
				"usage_time": "15:30",
				"利用時間":       "16:30",
			},
			expected: map[string]interface{}{
				"use_time": "16:30",
			},
		},
		{
			name: "legacy IC fields",
			input: map[string]interface{}{
				"ic_entry": "東京IC",
				"entry":    "大阪IC",
				"入口":       "名古屋IC",
				"ic_exit":  "福岡IC",
			},
			expected: map[string]interface{}{
				"entry_ic": "名古屋IC",
				"exit_ic":  "福岡IC",
			},
		},
		{
			name: "legacy amount fields",
			input: map[string]interface{}{
				"toll_amount":  500,
				"total_amount": 600,
				"料金":          700,
			},
			expected: map[string]interface{}{
				"amount": 700,
			},
		},
		{
			name: "legacy vehicle fields",
			input: map[string]interface{}{
				"vehicle_num":    "123",
				"vehicle_number": "456",
				"車両番号":          "789",
			},
			expected: map[string]interface{}{
				"car_number": "789",
			},
		},
		{
			name: "legacy ETC fields",
			input: map[string]interface{}{
				"etc_card_num":    "1234567890",
				"etc_card_number": "2345678901",
				"card_no":         "3456789012",
				"ETCカード番号":       "4567890123",
			},
			expected: map[string]interface{}{
				"etc_number": "4567890123",
			},
		},
		{
			name: "mixed fields",
			input: map[string]interface{}{
				"use_date":       "2023-12-25",
				"date":           "2023-12-26",
				"amount":         500,
				"toll_amount":    600,
				"custom_field":   "value",
			},
			expected: map[string]interface{}{
				"use_date":     "2023-12-26",
				"amount":       600,
				"custom_field": "value",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := converter.MapLegacyFields(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestFieldConverter_ConvertFieldValue tests field value conversion
func TestFieldConverter_ConvertFieldValue(t *testing.T) {
	converter := adapters.NewFieldConverter()

	tests := []struct {
		name       string
		value      interface{}
		targetType string
		expected   interface{}
		wantErr    bool
	}{
		{"nil value", nil, "string", nil, false},
		{"string to int32", "123", "int32", int32(123), false},
		{"string to int64", "123", "int64", int64(123), false},
		{"string to float64", "123.45", "float64", 123.45, false},
		{"string to time", "2023-12-25", "time", time.Date(2023, 12, 25, 0, 0, 0, 0, time.UTC), false},
		{"string to string", "test", "string", "test", false},
		{"string to bool true", "true", "bool", true, false},
		{"string to bool false", "false", "bool", false, false},
		{"unknown type", "test", "unknown", "test", false},
		{"invalid int32", "abc", "int32", nil, true},
		{"invalid float64", "abc", "float64", nil, true},
		{"invalid time", "invalid", "time", nil, true},
		{"invalid bool", "maybe", "bool", nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := converter.ConvertFieldValue(tt.value, tt.targetType)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

// TestFieldConverter_ValidateFieldType tests field type validation
func TestFieldConverter_ValidateFieldType(t *testing.T) {
	converter := adapters.NewFieldConverter()

	tests := []struct {
		name       string
		value      interface{}
		targetType string
		wantErr    bool
	}{
		{"valid int32", "123", "int32", false},
		{"valid float64", "123.45", "float64", false},
		{"valid time", "2023-12-25", "time", false},
		{"valid string", "test", "string", false},
		{"valid bool", "true", "bool", false},
		{"invalid int32", "abc", "int32", true},
		{"invalid float64", "abc", "float64", true},
		{"invalid time", "invalid", "time", true},
		{"invalid bool", "maybe", "bool", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := converter.ValidateFieldType(tt.value, tt.targetType)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestFieldConverter_ConvertStructFields tests struct field conversion
func TestFieldConverter_ConvertStructFields(t *testing.T) {
	converter := adapters.NewFieldConverter()

	type SourceStruct struct {
		StringField string
		IntField    string
		FloatField  string
	}

	type DestStruct struct {
		StringField string
		IntField    int32
		FloatField  float64
	}

	tests := []struct {
		name    string
		src     interface{}
		dst     interface{}
		wantErr bool
	}{
		{
			name: "successful conversion",
			src: &SourceStruct{
				StringField: "test",
				IntField:    "123",
				FloatField:  "123.45",
			},
			dst:     &DestStruct{},
			wantErr: false,
		},
		{
			name: "conversion error",
			src: &SourceStruct{
				StringField: "test",
				IntField:    "invalid",
				FloatField:  "123.45",
			},
			dst:     &DestStruct{},
			wantErr: true,
		},
		{
			name:    "non-settable destination",
			src:     &SourceStruct{},
			dst:     DestStruct{}, // Not a pointer
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := converter.ConvertStructFields(tt.src, tt.dst)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if !tt.wantErr && tt.name == "successful conversion" {
					dst := tt.dst.(*DestStruct)
					assert.Equal(t, "test", dst.StringField)
					assert.Equal(t, int32(123), dst.IntField)
					assert.Equal(t, 123.45, dst.FloatField)
				}
			}
		})
	}
}

// TestFieldConverter_Performance tests converter performance
func TestFieldConverter_Performance(t *testing.T) {
	converter := adapters.NewFieldConverter()

	// Test performance of string to int32 conversion
	t.Run("string to int32 performance", func(t *testing.T) {
		start := time.Now()
		for i := 0; i < 10000; i++ {
			_, err := converter.ConvertStringToInt32("123456")
			require.NoError(t, err)
		}
		duration := time.Since(start)
		t.Logf("10000 string to int32 conversions took %v", duration)
		assert.Less(t, duration, 100*time.Millisecond)
	})

	// Test performance of field mapping
	t.Run("field mapping performance", func(t *testing.T) {
		data := map[string]interface{}{
			"date":           "2023-12-25",
			"time":           "14:30",
			"entry":          "東京IC",
			"exit":           "大阪IC",
			"toll_amount":    "500",
			"vehicle_number": "123",
			"etc_card_num":   "1234567890",
		}

		start := time.Now()
		for i := 0; i < 1000; i++ {
			result := converter.MapLegacyFields(data)
			require.NotNil(t, result)
		}
		duration := time.Since(start)
		t.Logf("1000 field mappings took %v", duration)
		assert.Less(t, duration, 50*time.Millisecond)
	})

	// Test performance of IC name normalization
	t.Run("IC name normalization performance", func(t *testing.T) {
		start := time.Now()
		for i := 0; i < 10000; i++ {
			result := converter.NormalizeICName("東京インター")
			require.Equal(t, "東京IC", result)
		}
		duration := time.Since(start)
		t.Logf("10000 IC name normalizations took %v", duration)
		assert.Less(t, duration, 50*time.Millisecond)
	})
}

// TestFieldConverter_EdgeCases tests edge cases and boundary conditions
func TestFieldConverter_EdgeCases(t *testing.T) {
	converter := adapters.NewFieldConverter()

	t.Run("very long strings", func(t *testing.T) {
		longString := string(make([]byte, 10000))
		result := converter.NormalizeICName(longString)
		assert.Equal(t, "IC", result)
	})

	t.Run("unicode characters", func(t *testing.T) {
		result := converter.NormalizeICName("東京🚗インター")
		assert.Equal(t, "東京🚗IC", result)
	})

	t.Run("empty field mapping", func(t *testing.T) {
		result := converter.MapLegacyFields(nil)
		assert.NotNil(t, result)
		assert.Empty(t, result)
	})

	t.Run("complex time formats", func(t *testing.T) {
		result, err := converter.ConvertStringToTime("平成35年01月01日")
		assert.Error(t, err)
		assert.True(t, result.IsZero())
	})

	t.Run("boundary int32 values", func(t *testing.T) {
		// Test near boundary values
		result, err := converter.ConvertStringToInt32("2147483646")
		assert.NoError(t, err)
		assert.Equal(t, int32(2147483646), result)

		result, err = converter.ConvertStringToInt32("-2147483647")
		assert.NoError(t, err)
		assert.Equal(t, int32(-2147483647), result)
	})

	t.Run("float precision", func(t *testing.T) {
		result, err := converter.ConvertStringToFloat64("123.123456789012345")
		assert.NoError(t, err)
		assert.Equal(t, 123.123456789012345, result)
	})
}
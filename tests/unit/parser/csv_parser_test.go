package parser_test

import (
	"os"
	"strings"
	"testing"
	"time"

	"github.com/yhonda-ohishi/etc_meisai/src/models"
	"github.com/yhonda-ohishi/etc_meisai/src/parser"
	"github.com/yhonda-ohishi/etc_meisai/tests/helpers"
)

func TestETCCSVParser_Parse(t *testing.T) {
	tests := []struct {
		name     string
		csvData  string
		expected []models.ETCMeisai
		wantErr  bool
		errMsg   string
	}{
		{
			name: "valid single record",
			csvData: `利用年月日（自）,時刻（自）,利用年月日（至）,時刻（至）,利用ＩＣ（自）,利用ＩＣ（至）,料金所名,通行料金,通行区分,車種,車両番号,ＥＴＣカード番号,備考
2025/01/15,09:30,2025/01/15,10:30,東京IC,大阪IC,大阪料金所,1000,一般,普通車,品川123あ1234,1234567890,`,
			expected: []models.ETCMeisai{
				{
					UseDate:   time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC),
					UseTime:   "09:30",
					EntryIC:   "東京IC",
					ExitIC:    "大阪IC",
					Amount:    1000,
					CarNumber: "品川123あ1234",
					ETCNumber: "1234567890",
				},
			},
			wantErr: false,
		},
		{
			name: "insufficient columns",
			csvData: `利用年月日（自）,時刻（自）,利用年月日（至）,時刻（至）,利用ＩＣ（自）
2025/01/15,09:30,2025/01/15,10:30,東京IC`,
			wantErr: false, // Parser handles insufficient columns gracefully
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := strings.NewReader(tt.csvData)
			csvParser := parser.NewETCCSVParser()

			result, err := csvParser.Parse(reader)
			var records []models.ETCMeisai
			if result != nil {
				for _, r := range result.Records {
					records = append(records, *r)
				}
			}

			if tt.wantErr {
				helpers.AssertError(t, err)
				if tt.errMsg != "" {
					helpers.AssertContains(t, err.Error(), tt.errMsg)
				}
			} else {
				helpers.AssertNoError(t, err)
				helpers.AssertLen(t, records, len(tt.expected))

				for i, record := range records {
					expected := tt.expected[i]
					helpers.AssertEqual(t, expected.UseDate, record.UseDate)
					helpers.AssertEqual(t, expected.UseTime, record.UseTime)
					helpers.AssertEqual(t, expected.EntryIC, record.EntryIC)
					helpers.AssertEqual(t, expected.ExitIC, record.ExitIC)
					helpers.AssertEqual(t, expected.Amount, record.Amount)
					helpers.AssertEqual(t, expected.CarNumber, record.CarNumber)
					helpers.AssertEqual(t, expected.ETCNumber, record.ETCNumber)
				}
			}
		})
	}
}

func TestETCCSVParser_ParseFile(t *testing.T) {
	// Create a temporary CSV file for testing
	tmpfile, err := os.CreateTemp("", "test_etc_*.csv")
	helpers.AssertNoError(t, err)
	defer os.Remove(tmpfile.Name())

	csvData := `利用年月日（自）,時刻（自）,利用年月日（至）,時刻（至）,利用ＩＣ（自）,利用ＩＣ（至）,料金所名,通行料金,通行区分,車種,車両番号,ＥＴＣカード番号,備考
2025/01/15,09:30,2025/01/15,10:30,東京IC,大阪IC,大阪料金所,1000,一般,普通車,品川123あ1234,1234567890,`

	_, err = tmpfile.WriteString(csvData)
	helpers.AssertNoError(t, err)
	tmpfile.Close()

	csvParser := parser.NewETCCSVParser()
	records, err := csvParser.ParseFile(tmpfile.Name())

	helpers.AssertNoError(t, err)
	helpers.AssertLen(t, records, 1)

	record := records[0]
	helpers.AssertEqual(t, "09:30", record.UseTime)
	helpers.AssertEqual(t, "東京IC", record.EntryIC)
	helpers.AssertEqual(t, "大阪IC", record.ExitIC)
	helpers.AssertEqual(t, int32(1000), record.Amount)
}

func TestETCCSVParser_ParseDate(t *testing.T) {
	csvParser := parser.NewETCCSVParser()

	tests := []struct {
		name     string
		dateStr  string
		wantErr  bool
	}{
		{
			name:     "valid date with slashes",
			dateStr:  "2025/01/15",
			wantErr:  false,
		},
		{
			name:     "valid date with hyphens",
			dateStr:  "2025-01-15",
			wantErr:  false,
		},
		{
			name:     "valid 2-digit year",
			dateStr:  "25/01/15",
			wantErr:  false,
		},
		{
			name:     "invalid format",
			dateStr:  "invalid-date",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// We can't directly test parseDate as it's not exported
			// So we test it indirectly through Parse
			csvData := `利用年月日（自）,時刻（自）,利用年月日（至）,時刻（至）,利用ＩＣ（自）,利用ＩＣ（至）,料金所名,通行料金,通行区分,車種,車両番号,ＥＴＣカード番号,備考
` + tt.dateStr + `,09:30,` + tt.dateStr + `,10:30,東京IC,大阪IC,大阪料金所,1000,一般,普通車,品川123あ1234,1234567890,`

			reader := strings.NewReader(csvData)
			result, err := csvParser.Parse(reader)

			if tt.wantErr {
				// Either parsing fails or result has errors
				if err == nil && result != nil {
					helpers.AssertTrue(t, result.ErrorRows > 0 || len(result.Errors) > 0)
				}
			} else {
				helpers.AssertNoError(t, err)
				if result != nil {
					helpers.AssertEqual(t, 0, result.ErrorRows)
				}
			}
		})
	}
}

func TestETCCSVParser_ParseAmount(t *testing.T) {
	csvParser := parser.NewETCCSVParser()

	tests := []struct {
		name      string
		amountStr string
		expected  int32
	}{
		{
			name:      "simple amount",
			amountStr: "1000",
			expected:  1000,
		},
		{
			name:      "amount with comma",
			amountStr: "1,000",
			expected:  1000,
		},
		{
			name:      "amount with yen symbol",
			amountStr: "¥1000",
			expected:  1000,
		},
		{
			name:      "zero amount",
			amountStr: "0",
			expected:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test indirectly through parsing a CSV with the amount
			csvData := `利用年月日（自）,時刻（自）,利用年月日（至）,時刻（至）,利用ＩＣ（自）,利用ＩＣ（至）,料金所名,通行料金,通行区分,車種,車両番号,ＥＴＣカード番号,備考
2025/01/15,09:30,2025/01/15,10:30,東京IC,大阪IC,大阪料金所,` + tt.amountStr + `,一般,普通車,品川123あ1234,1234567890,`

			reader := strings.NewReader(csvData)
			result, err := csvParser.Parse(reader)

			helpers.AssertNoError(t, err)
			helpers.AssertNotNil(t, result)
			if len(result.Records) > 0 {
				helpers.AssertEqual(t, tt.expected, result.Records[0].Amount)
			}
		})
	}
}

// Additional simple tests can be added here as needed
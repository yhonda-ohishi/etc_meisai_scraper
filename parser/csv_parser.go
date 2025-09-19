package parser

import (
	"encoding/csv"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/yhonda-ohishi/etc_meisai/models"
)

// ETCCSVParser handles CSV parsing for ETC meisai data
type ETCCSVParser struct {
	dateFormats       []string
	encodingDetector  *EncodingDetector
}

// NewETCCSVParser creates a new CSV parser
func NewETCCSVParser() *ETCCSVParser {
	return &ETCCSVParser{
		dateFormats: []string{
			"06/01/02",    // 2桁年形式 (25/07/30)
			"2006/01/02",  // 4桁年形式 (2025/07/30)
			"06/1/2",      // 2桁年・0埋めなし (25/7/30)
			"2006/1/2",    // 4桁年・0埋めなし (2025/7/30)
			"2006-01-02",  // ハイフン形式
			"2006.01.02",  // ドット形式
			"20060102",    // 区切りなし
		},
		encodingDetector: NewEncodingDetector(),
	}
}

// ParseFile parses an ETC meisai CSV file
func (p *ETCCSVParser) ParseFile(filePath string) ([]models.ETCMeisai, error) {
	// Detect encoding and open with appropriate reader
	fileReader, encoding, err := p.encodingDetector.OpenFileWithDetectedEncoding(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open CSV file with encoding detection: %w", err)
	}

	// Close the underlying file if it's a file closer
	if closer, ok := fileReader.(io.Closer); ok {
		defer closer.Close()
	}

	reader := csv.NewReader(fileReader)
	reader.LazyQuotes = true
	reader.TrimLeadingSpace = true

	fmt.Printf("Detected encoding: %s\n", encoding.String())

	var records []models.ETCMeisai
	lineNum := 0

	for {
		row, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("error reading line %d: %w", lineNum, err)
		}

		lineNum++

		// Skip header row
		if lineNum == 1 && (strings.Contains(row[0], "利用") || strings.Contains(row[0], "日付")) {
			continue
		}

		// Parse row
		meisai, err := p.parseRow(row)
		if err != nil {
			// Log error but continue processing
			fmt.Printf("Warning: Failed to parse line %d: %v\n", lineNum, err)
			continue
		}

		records = append(records, *meisai)
	}

	return records, nil
}

// ParseCSVFile is an alias for ParseFile for compatibility
func (p *ETCCSVParser) ParseCSVFile(filePath string, isCorporate bool) ([]models.ETCMeisai, error) {
	records, err := p.ParseFile(filePath)
	if err != nil {
		return nil, err
	}

	// Set account type for all records
	accountType := "personal"
	if isCorporate {
		accountType = "corporate"
	}

	for i := range records {
		records[i].AccountType = accountType
	}

	return records, nil
}

// parseRow parses a single CSV row
func (p *ETCCSVParser) parseRow(row []string) (*models.ETCMeisai, error) {
	// ETC CSVは13列
	if len(row) < 13 {
		return nil, fmt.Errorf("insufficient columns: %d", len(row))
	}

	meisai := &models.ETCMeisai{}

	// Parse start date (利用年月日（自）)
	dateStr := strings.TrimSpace(row[0])
	if dateStr == "" {
		// 日付が空の場合は、終了日付（至）を使用
		if len(row) > 2 && strings.TrimSpace(row[2]) != "" {
			dateStr = strings.TrimSpace(row[2])
		} else {
			// それも空なら現在日付を使用
			meisai.Date = time.Now().Format("2006/01/02")
			meisai.UsageDate = time.Now()
		}
	}

	if dateStr != "" {
		date, err := p.parseDate(dateStr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse date '%s': %w", dateStr, err)
		}
		meisai.Date = date.Format("2006/01/02")
		meisai.UsageDate = date
	}

	// Parse other fields based on actual CSV format
	// 列: 利用年月日（自）,時刻（自）,利用年月日（至）,時刻（至）,利用ＩＣ（自）,利用ＩＣ（至）,
	//     料金所名,通行料金,通行区分,車種,車両番号,ＥＴＣカード番号,備考
	meisai.Time = strings.TrimSpace(row[1])           // 時刻（自）
	meisai.ICEntry = strings.TrimSpace(row[4])        // 利用ＩＣ（自）
	meisai.EntryIC = meisai.ICEntry                   // エイリアス
	meisai.ICExit = strings.TrimSpace(row[5])         // 利用ＩＣ（至）
	meisai.ExitIC = meisai.ICExit                     // エイリアス
	meisai.TollGate = strings.TrimSpace(row[6])       // 料金所名
	meisai.VehicleNo = strings.TrimSpace(row[10])     // 車両番号
	meisai.VehicleNumber = meisai.VehicleNo           // エイリアス
	meisai.CardNo = strings.TrimSpace(row[11])        // ＥＴＣカード番号
	meisai.CardNumber = meisai.CardNo                 // エイリアス
	meisai.ETCCardNumber = meisai.CardNo              // エイリアス

	// Parse amounts
	amount := p.parseAmount(row[7])                   // 通行料金
	meisai.TotalAmount = amount
	meisai.TollAmount = amount                        // エイリアス
	meisai.Amount = amount                            // エイリアス

	meisai.VehicleType = strings.TrimSpace(row[9])    // 車種
	meisai.Remarks = strings.TrimSpace(row[12])       // 備考

	// 通行区分を追加
	if len(row) > 8 {
		meisai.UsageType = strings.TrimSpace(row[8])  // 通行区分
	}

	// Set timestamps
	meisai.ImportedAt = time.Now()
	meisai.CreatedAt = time.Now()

	return meisai, nil
}

// parseDate tries multiple date formats to parse the date string
func (p *ETCCSVParser) parseDate(dateStr string) (time.Time, error) {
	dateStr = strings.TrimSpace(dateStr)

	// Try each date format
	for _, format := range p.dateFormats {
		if date, err := time.Parse(format, dateStr); err == nil {
			// Handle 2-digit years (assume 20xx for years 00-50, 19xx for 51-99)
			if format == "06/01/02" || format == "06/1/2" {
				year := date.Year()
				if year < 100 {
					if year <= 50 {
						date = date.AddDate(2000, 0, 0)
					} else {
						date = date.AddDate(1900, 0, 0)
					}
				}
			}
			return date, nil
		}
	}

	return time.Time{}, fmt.Errorf("unable to parse date '%s' with any supported format", dateStr)
}

// parseAmount parses amount string to int
func (p *ETCCSVParser) parseAmount(s string) int {
	// Remove non-numeric characters
	s = strings.ReplaceAll(s, ",", "")
	s = strings.ReplaceAll(s, "円", "")
	s = strings.ReplaceAll(s, "¥", "")
	s = strings.ReplaceAll(s, "￥", "")
	s = strings.TrimSpace(s)

	amount, _ := strconv.Atoi(s)
	return amount
}

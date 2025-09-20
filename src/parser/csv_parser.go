package parser

import (
	"encoding/csv"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/yhonda-ohishi/etc_meisai/src/models"
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

	// Account type was passed as metadata but is not stored in the model
	// (account type removed from model when migrating to gRPC-only architecture)

	return records, nil
}

// Parse parses CSV content from a reader and returns a ParseResult
func (p *ETCCSVParser) Parse(r io.Reader) (*ParseResult, error) {
	reader := csv.NewReader(r)
	reader.LazyQuotes = true
	reader.TrimLeadingSpace = true

	result := &ParseResult{
		Records: []*models.ETCMeisai{},
		Errors:  []ParseError{},
	}

	lineNum := 0

	for {
		row, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			result.ErrorRows++
			result.Errors = append(result.Errors, ParseError{
				Row:     lineNum,
				Message: err.Error(),
			})
			continue
		}

		lineNum++
		result.TotalRows++

		// Skip header row
		if lineNum == 1 && (strings.Contains(row[0], "利用") || strings.Contains(row[0], "日付")) {
			continue
		}

		// Parse row
		meisai, err := p.parseRow(row)
		if err != nil {
			result.ErrorRows++
			result.Errors = append(result.Errors, ParseError{
				Row:     lineNum,
				Message: err.Error(),
			})
			continue
		}

		result.Records = append(result.Records, meisai)
		result.ValidRows++
	}

	return result, nil
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
			meisai.UseDate = time.Now()
		}
	}

	if dateStr != "" {
		date, err := p.parseDate(dateStr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse date '%s': %w", dateStr, err)
		}
		meisai.UseDate = date
	}

	// Parse other fields based on actual CSV format
	// 列: 利用年月日（自）,時刻（自）,利用年月日（至）,時刻（至）,利用ＩＣ（自）,利用ＩＣ（至）,
	//     料金所名,通行料金,通行区分,車種,車両番号,ＥＴＣカード番号,備考
	meisai.UseTime = strings.TrimSpace(row[1])        // 時刻（自）
	meisai.EntryIC = strings.TrimSpace(row[4])        // 利用ＩＣ（自）
	meisai.ExitIC = strings.TrimSpace(row[5])         // 利用ＩＣ（至）
	// meisai.TollGate = strings.TrimSpace(row[6])   // 料金所名 (not in model)
	meisai.CarNumber = strings.TrimSpace(row[10])     // 車両番号
	meisai.ETCNumber = strings.TrimSpace(row[11])     // ＥＴＣカード番号

	// Parse amounts
	amount := p.parseAmount(row[7])                   // 通行料金
	meisai.Amount = int32(amount)                     // Amount is int32 in model

	// meisai.VehicleType = strings.TrimSpace(row[9])    // 車種 (not in model)
	// meisai.Remarks = strings.TrimSpace(row[12])       // 備考 (not in model)
	// meisai.UsageType = strings.TrimSpace(row[8])      // 通行区分 (not in model)

	// Set timestamps
	meisai.CreatedAt = time.Now()
	meisai.UpdatedAt = time.Now()

	// Generate hash
	meisai.Hash = meisai.GenerateHash()

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

package parser

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/yhonda-ohishi/etc_meisai/models"
)

// ETCCSVParser handles CSV parsing for ETC meisai data
type ETCCSVParser struct {
	dateFormat string
}

// NewETCCSVParser creates a new CSV parser
func NewETCCSVParser() *ETCCSVParser {
	return &ETCCSVParser{
		dateFormat: "2006/01/02",
	}
}

// ParseFile parses an ETC meisai CSV file
func (p *ETCCSVParser) ParseFile(filePath string) ([]models.ETCMeisai, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open CSV file: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.LazyQuotes = true
	reader.TrimLeadingSpace = true

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
		if lineNum == 1 && strings.Contains(row[0], "日付") {
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

// parseRow parses a single CSV row
func (p *ETCCSVParser) parseRow(row []string) (*models.ETCMeisai, error) {
	// TODO: Adjust indices based on actual CSV format
	if len(row) < 10 {
		return nil, fmt.Errorf("insufficient columns: %d", len(row))
	}

	meisai := &models.ETCMeisai{}

	// Parse date
	date, err := time.Parse(p.dateFormat, strings.TrimSpace(row[0]))
	if err != nil {
		return nil, fmt.Errorf("failed to parse date: %w", err)
	}
	meisai.Date = date

	// Parse other fields (adjust indices based on actual CSV)
	meisai.Time = strings.TrimSpace(row[1])
	meisai.ICEntry = strings.TrimSpace(row[2])
	meisai.ICExit = strings.TrimSpace(row[3])
	meisai.VehicleNo = strings.TrimSpace(row[4])
	meisai.CardNo = strings.TrimSpace(row[5])

	// Parse amounts
	meisai.Amount = p.parseAmount(row[6])
	meisai.DiscountAmount = p.parseAmount(row[7])
	meisai.TotalAmount = p.parseAmount(row[8])

	// Optional fields
	if len(row) > 9 {
		meisai.UsageType = strings.TrimSpace(row[9])
	}
	if len(row) > 10 {
		meisai.PaymentMethod = strings.TrimSpace(row[10])
	}
	if len(row) > 11 {
		meisai.RouteCode = strings.TrimSpace(row[11])
	}
	if len(row) > 12 {
		meisai.Distance, _ = strconv.ParseFloat(strings.TrimSpace(row[12]), 64)
	}

	return meisai, nil
}

// parseAmount parses amount string to int
func (p *ETCCSVParser) parseAmount(s string) int {
	// Remove non-numeric characters
	s = strings.ReplaceAll(s, ",", "")
	s = strings.ReplaceAll(s, "円", "")
	s = strings.ReplaceAll(s, "¥", "")
	s = strings.TrimSpace(s)

	amount, _ := strconv.Atoi(s)
	return amount
}
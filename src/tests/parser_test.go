package tests

import (
	"testing"
	"path/filepath"
	"github.com/yhonda-ohishi/etc_meisai/src/parser"
)

func TestCSVParser(t *testing.T) {
	// Create parser
	p := parser.NewETCCSVParser()

	// Test file path
	testFile := filepath.Join("..", "testdata", "sample_etc.csv")

	// Parse CSV file
	records, err := p.ParseFile(testFile)
	if err != nil {
		t.Fatalf("Failed to parse CSV file: %v", err)
	}

	// Check record count
	if len(records) != 10 {
		t.Errorf("Expected 10 records, got %d", len(records))
	}

	// Check first record
	if len(records) > 0 {
		first := records[0]
		if first.Date != "2025/09/01" {
			t.Errorf("Expected date 2025/09/01, got %s", first.Date)
		}
		if first.ICEntry != "東京IC" {
			t.Errorf("Expected entry IC 東京IC, got %s", first.ICEntry)
		}
		if first.TollAmount != 1500 {
			t.Errorf("Expected amount 1500, got %d", first.TollAmount)
		}
		if first.VehicleNo != "品川300あ1234" {
			t.Errorf("Expected vehicle no 品川300あ1234, got %s", first.VehicleNo)
		}
	}

	// Check last record
	if len(records) > 9 {
		last := records[9]
		if last.Date != "2025/09/05" {
			t.Errorf("Expected date 2025/09/05, got %s", last.Date)
		}
		if last.Remarks != "深夜割引" {
			t.Errorf("Expected remarks 深夜割引, got %s", last.Remarks)
		}
	}

	t.Logf("Successfully parsed %d records", len(records))
}

func TestCSVParserWithAccountType(t *testing.T) {
	p := parser.NewETCCSVParser()
	testFile := filepath.Join("..", "testdata", "sample_etc.csv")

	// Test with corporate account
	records, err := p.ParseCSVFile(testFile, true)
	if err != nil {
		t.Fatalf("Failed to parse CSV file: %v", err)
	}

	// Check account type
	for i, record := range records {
		if record.AccountType != "corporate" {
			t.Errorf("Record %d: Expected account type 'corporate', got '%s'", i, record.AccountType)
		}
	}

	// Test with personal account
	records, err = p.ParseCSVFile(testFile, false)
	if err != nil {
		t.Fatalf("Failed to parse CSV file: %v", err)
	}

	// Check account type
	for i, record := range records {
		if record.AccountType != "personal" {
			t.Errorf("Record %d: Expected account type 'personal', got '%s'", i, record.AccountType)
		}
	}
}
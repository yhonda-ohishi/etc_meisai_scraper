package integration

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	etc_meisai "github.com/yhonda-ohishi/etc_meisai"
	"github.com/yhonda-ohishi/etc_meisai/models"
)

// TestHashImportHandler tests the hash-based import functionality
func TestHashImportHandler(t *testing.T) {
	// Create a temporary CSV file for testing (matching actual ETC CSV format)
	csvContent := `利用年月日（自）,時刻（自）,利用年月日（至）,時刻（至）,利用ＩＣ（自）,利用ＩＣ（至）,料金所名,通行料金,通行区分,車種,車両番号,ＥＴＣカード番号,備考
2025/09/01,10:00,2025/09/01,10:30,東京IC,横浜IC,横浜料金所,1500,1,2,品川300あ1234,1234567890123456,
2025/09/01,14:00,2025/09/01,15:00,横浜IC,静岡IC,静岡料金所,2800,1,2,品川300あ1234,1234567890123456,
2025/09/02,09:00,2025/09/02,11:00,静岡IC,名古屋IC,名古屋料金所,3200,1,2,品川300あ5678,9876543210987654,`

	tmpDir := t.TempDir()
	csvPath := filepath.Join(tmpDir, "test.csv")
	if err := os.WriteFile(csvPath, []byte(csvContent), 0644); err != nil {
		t.Fatalf("Failed to create test CSV: %v", err)
	}

	// Test import request
	reqBody := etc_meisai.HashImportRequest{
		CSVPath: csvPath,
		Options: etc_meisai.HashImportOptions{
			SkipDuplicates: true,
			UpdateExisting: true,
			ValidateOnly:   false,
		},
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		t.Fatalf("Failed to marshal request: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/etc/import/hash", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	// Create response recorder
	rr := httptest.NewRecorder()

	// Call handler
	handler := http.HandlerFunc(etc_meisai.HashImportHandler)
	handler.ServeHTTP(rr, req)

	// Check status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusOK)
		t.Errorf("Response body: %s", rr.Body.String())
	}

	// Parse response
	var result models.ImportResult
	if err := json.Unmarshal(rr.Body.Bytes(), &result); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	// Verify results
	if !result.Success {
		t.Errorf("Import should have succeeded")
	}

	if result.ProcessedCount == 0 {
		t.Errorf("Should have processed some records")
	}

	t.Logf("Import results: Added=%d, Updated=%d, Duplicates=%d",
		result.AddedCount, result.UpdatedCount, result.DuplicateCount)
}

// TestDuplicateDetection tests duplicate detection functionality
func TestDuplicateDetection(t *testing.T) {
	// Create CSV with duplicate records
	csvContent := `利用年月日（自）,時刻（自）,利用年月日（至）,時刻（至）,利用ＩＣ（自）,利用ＩＣ（至）,料金所名,通行料金,通行区分,車種,車両番号,ＥＴＣカード番号,備考
2025/09/01,10:00,2025/09/01,10:30,東京IC,横浜IC,横浜料金所,1500,1,2,品川300あ1234,1234567890123456,
2025/09/01,10:00,2025/09/01,10:30,東京IC,横浜IC,横浜料金所,1500,1,2,品川300あ1234,1234567890123456,
2025/09/01,14:00,2025/09/01,15:00,横浜IC,静岡IC,静岡料金所,2800,1,2,品川300あ1234,1234567890123456,`

	tmpDir := t.TempDir()
	csvPath := filepath.Join(tmpDir, "duplicate.csv")
	if err := os.WriteFile(csvPath, []byte(csvContent), 0644); err != nil {
		t.Fatalf("Failed to create test CSV: %v", err)
	}

	// First import
	reqBody := etc_meisai.HashImportRequest{
		CSVPath: csvPath,
		Options: etc_meisai.HashImportOptions{
			SkipDuplicates: true,
			UpdateExisting: false,
			ValidateOnly:   false,
		},
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		t.Fatalf("Failed to marshal request: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/etc/import/hash", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(etc_meisai.HashImportHandler)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var result models.ImportResult
	if err := json.Unmarshal(rr.Body.Bytes(), &result); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	// Should detect duplicates within the same file
	if result.DuplicateCount == 0 {
		t.Errorf("Should have detected duplicates within the file")
	}

	// Second import with same data
	req2 := httptest.NewRequest(http.MethodPost, "/api/etc/import/hash", bytes.NewBuffer(body))
	req2.Header.Set("Content-Type", "application/json")

	rr2 := httptest.NewRecorder()
	handler.ServeHTTP(rr2, req2)

	var result2 models.ImportResult
	if err := json.Unmarshal(rr2.Body.Bytes(), &result2); err != nil {
		t.Fatalf("Failed to parse second response: %v", err)
	}

	// All records should be duplicates on second import
	if result2.AddedCount != 0 {
		t.Errorf("Should not have added any new records on second import")
	}

	t.Logf("First import: Added=%d, Duplicates=%d", result.AddedCount, result.DuplicateCount)
	t.Logf("Second import: Added=%d, Duplicates=%d", result2.AddedCount, result2.DuplicateCount)
}

// TestChangeDetection tests change detection functionality
func TestChangeDetection(t *testing.T) {
	tmpDir := t.TempDir()

	// First CSV
	csvContent1 := `利用年月日（自）,時刻（自）,利用年月日（至）,時刻（至）,利用ＩＣ（自）,利用ＩＣ（至）,料金所名,通行料金,通行区分,車種,車両番号,ＥＴＣカード番号,備考
2025/09/01,10:00,2025/09/01,10:30,東京IC,横浜IC,横浜料金所,1500,1,2,品川300あ1234,1234567890123456,`

	csvPath1 := filepath.Join(tmpDir, "original.csv")
	if err := os.WriteFile(csvPath1, []byte(csvContent1), 0644); err != nil {
		t.Fatalf("Failed to create first CSV: %v", err)
	}

	// Import first CSV
	reqBody1 := etc_meisai.HashImportRequest{
		CSVPath: csvPath1,
		Options: etc_meisai.HashImportOptions{
			SkipDuplicates: true,
			UpdateExisting: true,
			ValidateOnly:   false,
		},
	}

	body1, _ := json.Marshal(reqBody1)
	req1 := httptest.NewRequest(http.MethodPost, "/api/etc/import/hash", bytes.NewBuffer(body1))
	req1.Header.Set("Content-Type", "application/json")

	rr1 := httptest.NewRecorder()
	handler := http.HandlerFunc(etc_meisai.HashImportHandler)
	handler.ServeHTTP(rr1, req1)

	// Second CSV with changed amount
	csvContent2 := `利用年月日（自）,時刻（自）,利用年月日（至）,時刻（至）,利用ＩＣ（自）,利用ＩＣ（至）,料金所名,通行料金,通行区分,車種,車両番号,ＥＴＣカード番号,備考
2025/09/01,10:00,2025/09/01,10:30,東京IC,横浜IC,横浜料金所,1600,1,2,品川300あ1234,1234567890123456,`

	csvPath2 := filepath.Join(tmpDir, "updated.csv")
	if err := os.WriteFile(csvPath2, []byte(csvContent2), 0644); err != nil {
		t.Fatalf("Failed to create second CSV: %v", err)
	}

	// Import second CSV
	reqBody2 := etc_meisai.HashImportRequest{
		CSVPath: csvPath2,
		Options: etc_meisai.HashImportOptions{
			SkipDuplicates: false,
			UpdateExisting: true,
			ValidateOnly:   false,
		},
	}

	body2, _ := json.Marshal(reqBody2)
	req2 := httptest.NewRequest(http.MethodPost, "/api/etc/import/hash", bytes.NewBuffer(body2))
	req2.Header.Set("Content-Type", "application/json")

	rr2 := httptest.NewRecorder()
	handler.ServeHTTP(rr2, req2)

	var result2 models.ImportResult
	if err := json.Unmarshal(rr2.Body.Bytes(), &result2); err != nil {
		t.Fatalf("Failed to parse second response: %v", err)
	}

	// Should detect the change
	if result2.UpdatedCount == 0 {
		t.Errorf("Should have detected the changed record")
	}

	t.Logf("Change detection: Updated=%d", result2.UpdatedCount)
}

// TestHashStatistics tests the hash statistics endpoint
func TestHashStatistics(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/etc/hash/stats", nil)
	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(etc_meisai.GetHashStatsHandler)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var stats map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &stats); err != nil {
		t.Fatalf("Failed to parse stats response: %v", err)
	}

	// Check required fields
	if _, ok := stats["total_records"]; !ok {
		t.Errorf("Stats should include total_records")
	}

	if _, ok := stats["memory_usage_estimate"]; !ok {
		t.Errorf("Stats should include memory_usage_estimate")
	}

	t.Logf("Hash stats: %+v", stats)
}

// TestClearHashIndex tests clearing the hash index
func TestClearHashIndex(t *testing.T) {
	// First add some data
	csvContent := `利用年月日（自）,時刻（自）,利用年月日（至）,時刻（至）,利用ＩＣ（自）,利用ＩＣ（至）,料金所名,通行料金,通行区分,車種,車両番号,ＥＴＣカード番号,備考
2025/09/01,10:00,2025/09/01,10:30,東京IC,横浜IC,横浜料金所,1500,1,2,品川300あ1234,1234567890123456,`

	tmpDir := t.TempDir()
	csvPath := filepath.Join(tmpDir, "data.csv")
	if err := os.WriteFile(csvPath, []byte(csvContent), 0644); err != nil {
		t.Fatalf("Failed to create CSV: %v", err)
	}

	// Import data
	reqBody := etc_meisai.HashImportRequest{
		CSVPath: csvPath,
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/etc/import/hash", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(etc_meisai.HashImportHandler)
	handler.ServeHTTP(rr, req)

	// Clear index
	clearReq := httptest.NewRequest(http.MethodPost, "/api/etc/hash/clear", nil)
	clearRr := httptest.NewRecorder()

	clearHandler := http.HandlerFunc(etc_meisai.ClearHashIndexHandler)
	clearHandler.ServeHTTP(clearRr, clearReq)

	if status := clearRr.Code; status != http.StatusOK {
		t.Errorf("Clear handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var clearResult map[string]interface{}
	if err := json.Unmarshal(clearRr.Body.Bytes(), &clearResult); err != nil {
		t.Fatalf("Failed to parse clear response: %v", err)
	}

	if clearResult["success"] != true {
		t.Errorf("Clear operation should have succeeded")
	}

	// Verify index is empty
	statsReq := httptest.NewRequest(http.MethodGet, "/api/etc/hash/stats", nil)
	statsRr := httptest.NewRecorder()

	statsHandler := http.HandlerFunc(etc_meisai.GetHashStatsHandler)
	statsHandler.ServeHTTP(statsRr, statsReq)

	var stats map[string]interface{}
	json.Unmarshal(statsRr.Body.Bytes(), &stats)

	if stats["total_records"].(float64) != 0 {
		t.Errorf("Index should be empty after clear")
	}
}

// Helper function to create CSV content
func createCSVContent(records [][]string) string {
	buf := new(bytes.Buffer)
	w := csv.NewWriter(buf)

	// Write header
	w.Write([]string{"利用年月日（自）", "時刻（自）", "利用年月日（至）", "時刻（至）", "利用ＩＣ（自）", "利用ＩＣ（至）", "料金所名", "通行料金", "通行区分", "車種", "車両番号", "ＥＴＣカード番号", "備考"})

	// Write records
	for _, record := range records {
		w.Write(record)
	}

	w.Flush()
	return buf.String()
}

// BenchmarkHashImport benchmarks the hash import performance
func BenchmarkHashImport(b *testing.B) {
	// Create test data
	records := make([][]string, 1000)
	for i := 0; i < 1000; i++ {
		records[i] = []string{
			"2025/09/01",
			"10:00",
			"2025/09/01",
			"10:30",
			fmt.Sprintf("IC%d", i%10),
			fmt.Sprintf("IC%d", (i+1)%10),
			"料金所",
			fmt.Sprintf("%d", 1000+i*100),
			"1",
			"2",
			fmt.Sprintf("品川300あ%04d", i),
			fmt.Sprintf("%016d", i),
			"",
		}
	}

	csvContent := createCSVContent(records)
	tmpDir := b.TempDir()
	csvPath := filepath.Join(tmpDir, "benchmark.csv")
	os.WriteFile(csvPath, []byte(csvContent), 0644)

	reqBody := etc_meisai.HashImportRequest{
		CSVPath: csvPath,
		Options: etc_meisai.HashImportOptions{
			SkipDuplicates: true,
			UpdateExisting: true,
			ValidateOnly:   false,
		},
	}

	body, _ := json.Marshal(reqBody)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodPost, "/api/etc/import/hash", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(etc_meisai.HashImportHandler)
		handler.ServeHTTP(rr, req)
	}
}
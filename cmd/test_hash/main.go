package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"time"

	etc_meisai "github.com/yhonda-ohishi/etc_meisai"
	"github.com/yhonda-ohishi/etc_meisai/models"
)

func main() {
	// 実際のCSVファイルパス
	csvFiles := []string{
		"C:/go/etc_meisai/downloads/202509151526.csv", // 488KB
		"C:/go/etc_meisai/downloads/202509151527.csv", // 11KB
	}

	for _, csvPath := range csvFiles {
		fmt.Printf("\n=== Testing with: %s ===\n", csvPath)
		testHashImport(csvPath)
	}

	// 統計情報を取得
	fmt.Printf("\n=== Hash Index Statistics ===\n")
	getHashStats()
}

func testHashImport(csvPath string) {
	startTime := time.Now()

	// インポートリクエストの作成
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
		fmt.Printf("Error marshaling request: %v\n", err)
		return
	}

	// HTTPリクエストの作成
	req := httptest.NewRequest(http.MethodPost, "/api/etc/import/hash", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	// レスポンスレコーダーの作成
	rr := httptest.NewRecorder()

	// ハンドラーの実行
	handler := http.HandlerFunc(etc_meisai.HashImportHandler)
	handler.ServeHTTP(rr, req)

	// 処理時間の計測
	duration := time.Since(startTime)

	// レスポンスの解析
	var result models.ImportResult
	if err := json.Unmarshal(rr.Body.Bytes(), &result); err != nil {
		fmt.Printf("Error parsing response: %v\n", err)
		return
	}

	// 結果の表示
	fmt.Printf("Status Code: %d\n", rr.Code)
	fmt.Printf("Success: %v\n", result.Success)
	fmt.Printf("Processing Time: %v\n", duration)
	fmt.Printf("Results:\n")
	fmt.Printf("  - Total Processed: %d\n", result.ProcessedCount)
	fmt.Printf("  - Added: %d\n", result.AddedCount)
	fmt.Printf("  - Updated: %d\n", result.UpdatedCount)
	fmt.Printf("  - Duplicates: %d\n", result.DuplicateCount)
	fmt.Printf("  - Errors: %d\n", result.ErrorCount)

	if result.ErrorCount > 0 && len(result.Errors) > 0 {
		fmt.Printf("First Error: %v\n", result.Errors[0])
	}

	// セッション情報
	if result.Session != nil {
		fmt.Printf("Session ID: %s\n", result.Session.ID)
		fmt.Printf("Session Status: %s\n", result.Session.Status)
	}
}

func getHashStats() {
	req := httptest.NewRequest(http.MethodGet, "/api/etc/hash/stats", nil)
	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(etc_meisai.GetHashStatsHandler)
	handler.ServeHTTP(rr, req)

	var stats map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &stats); err != nil {
		fmt.Printf("Error parsing stats: %v\n", err)
		return
	}

	fmt.Printf("Total Records in Index: %.0f\n", stats["total_records"].(float64))
	fmt.Printf("Estimated Memory Usage: %s\n", stats["memory_usage_estimate"])
	fmt.Printf("Index Type: %s\n", stats["index_type"])

	if algorithms, ok := stats["hash_algorithms"].(map[string]interface{}); ok {
		fmt.Printf("Hash Algorithms:\n")
		for key, value := range algorithms {
			fmt.Printf("  - %s: %s\n", key, value)
		}
	}
}
package main

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"io"
	"log"

	"github.com/yhonda-ohishi/etc_meisai_scraper/src/scraper"
)

func main() {
	// スクレイパー設定
	config := &scraper.ScraperConfig{
		UserID:       "your-user-id",
		Password:     "your-password",
		DownloadPath: "./temp",
		Headless:     true,
	}

	scraperInstance, err := scraper.NewETCScraper(config, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer scraperInstance.Close()

	// 初期化
	if err := scraperInstance.Initialize(); err != nil {
		log.Fatal(err)
	}

	if err := scraperInstance.Login(); err != nil {
		log.Fatal(err)
	}

	// === 使用方法1: バッファとして直接取得 ===
	fmt.Println("=== Method 1: Direct Buffer ===")
	result, err := scraperInstance.DownloadMeisaiToBuffer("2024-01-01", "2024-01-31")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Downloaded %d bytes\n", len(result))

	// CSVデータを直接処理
	csvContent := string(result)
	fmt.Printf("First 200 chars:\n%s\n", csvContent[:min(200, len(csvContent))])

	// === 使用方法2: io.Reader として取得（ストリーミング風） ===
	fmt.Println("\n=== Method 2: As Reader ===")
	buffer, err := scraperInstance.DownloadMeisaiToBuffer("2024-01-01", "2024-01-31")
	if err != nil {
		log.Fatal(err)
	}
	reader := bytes.NewReader(buffer)

	// CSV処理ライブラリで直接読み込み
	csvReader := csv.NewReader(reader)
	records, err := csvReader.ReadAll()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Read %d records via csv.Reader\n", len(records))
	if len(records) > 0 {
		fmt.Printf("Header: %v\n", records[0])
	}

	// === 使用方法3: 構造化データとして解析 ===
	fmt.Println("\n=== Method 3: Parse CSV Data ===")
	// CSVデータをパース
	csvReader2 := csv.NewReader(bytes.NewReader(result))
	parsedRecords, err := csvReader2.ReadAll()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Parsed %d rows\n", len(parsedRecords))
	for i, row := range parsedRecords[:min(3, len(parsedRecords))] {
		fmt.Printf("Row %d: %v\n", i+1, row)
	}

	// === 実用例: メモリ内でのデータ変換 ===
	fmt.Println("\n=== Practical Example: In-Memory Processing ===")
	processInMemory(result)
}

// メモリ内でのデータ処理例
func processInMemory(csvData []byte) {
	reader := csv.NewReader(bytes.NewReader(csvData))

	var totalFare int
	var recordCount int

	// ヘッダーをスキップ
	if _, err := reader.Read(); err != nil {
		log.Printf("Error reading header: %v", err)
		return
	}

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Printf("Error reading record: %v", err)
			continue
		}

		// 料金カラムを集計（例: 8列目が料金と仮定）
		if len(record) > 7 {
			// 料金処理のロジック
			recordCount++
		}
	}

	fmt.Printf("Processed %d records, Total fare: ¥%d\n", recordCount, totalFare)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
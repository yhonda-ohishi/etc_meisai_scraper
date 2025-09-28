package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/yhonda-ohishi/etc_meisai_scraper/src/scraper"
)

func main() {
	// ロガー設定
	logger := log.New(os.Stdout, "[TEST-DOWNLOAD] ", log.LstdFlags|log.Lshortfile)

	// 環境変数からアカウント情報を取得
	userID := os.Getenv("ETC_USER_ID")
	password := os.Getenv("ETC_PASSWORD")

	if userID == "" || password == "" {
		logger.Fatal("Please set ETC_USER_ID and ETC_PASSWORD environment variables")
	}

	// スクレイパー設定
	config := &scraper.ScraperConfig{
		UserID:       userID,
		Password:     password,
		DownloadPath: "./downloads/test_" + time.Now().Format("20060102_150405"),
		Headless:     false, // デバッグ用にブラウザを表示
		Timeout:      30000,
		RetryCount:   3,
	}

	// スクレイパー作成
	etcScraper, err := scraper.NewETCScraper(config, logger)
	if err != nil {
		logger.Fatalf("Failed to create scraper: %v", err)
	}
	defer etcScraper.Close()

	// Playwright初期化
	logger.Println("Initializing Playwright...")
	if err := etcScraper.Initialize(); err != nil {
		logger.Fatalf("Failed to initialize scraper: %v", err)
	}

	// ログイン
	logger.Println("Logging in to https://www.etc-meisai.jp/...")
	if err := etcScraper.Login(); err != nil {
		logger.Fatalf("Login failed: %v", err)
	}

	// 日付範囲設定（先月のデータをダウンロード）
	now := time.Now()
	lastMonth := now.AddDate(0, -1, 0)
	fromDate := lastMonth.Format("2006/01/01")
	toDate := lastMonth.Format("2006/01/31")

	// データダウンロード
	logger.Printf("Downloading meisai from %s to %s...", fromDate, toDate)
	csvPath, err := etcScraper.DownloadMeisai(fromDate, toDate)
	if err != nil {
		logger.Fatalf("Download failed: %v", err)
	}

	logger.Printf("Successfully downloaded: %s", csvPath)
	fmt.Println("\nTest completed successfully!")
	fmt.Println("Downloaded file:", csvPath)
	fmt.Println("\nThis test demonstrates that the scraper can:")
	fmt.Println("1. Navigate to https://www.etc-meisai.jp/")
	fmt.Println("2. Login with credentials")
	fmt.Println("3. Download ETC meisai data for specified date range")
}
package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/yhonda-ohishi/etc_meisai/config"
	"github.com/yhonda-ohishi/etc_meisai/scraper"
	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not loaded")
	}

	// Load multiple corporate accounts from environment
	accounts, err := config.LoadCorporateAccountsFromEnv()
	if err != nil || len(accounts) == 0 {
		log.Fatal("No accounts found in environment variables")
	}

	log.Printf("Found %d corporate accounts", len(accounts))

	// Calculate date range: Last month's 1st to today
	now := time.Now()
	today := now

	// Get first day of last month
	lastMonth := now.AddDate(0, -1, 0)
	fromDate := time.Date(lastMonth.Year(), lastMonth.Month(), 1, 0, 0, 0, 0, time.Local)

	// Use today as the end date
	toDate := time.Date(today.Year(), today.Month(), today.Day(), 23, 59, 59, 0, time.Local)

	log.Printf("📅 Date range: %s to %s (先月1日から今日まで)",
		fromDate.Format("2006-01-02"), toDate.Format("2006-01-02"))

	// Process each account
	successCount := 0
	failCount := 0

	for i, account := range accounts {
		log.Printf("\n========================================")
		log.Printf("Processing account %d/%d: %s", i+1, len(accounts), account.UserID)
		log.Printf("========================================")

		// Create download directory for this account
		downloadPath := fmt.Sprintf("./test_downloads/account_%s", account.UserID)
		if err := os.MkdirAll(downloadPath, 0755); err != nil {
			log.Printf("Failed to create download directory: %v", err)
			failCount++
			continue
		}

		// Configure scraper with improved settings
		config := &scraper.ScraperConfig{
			UserID:       account.UserID,
			Password:     account.Password,
			DownloadPath: downloadPath,
			Headless:     false, // Set to false to see browser windows
			Timeout:      30000,
			RetryCount:   3,
			SlowMo:       100,
			UserAgent:    "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
			Viewport: &scraper.ViewportSize{
				Width:  1920,
				Height: 1080,
			},
		}

		// Create and initialize scraper
		s, err := scraper.NewActualETCScraper(config)
		if err != nil {
			log.Printf("Failed to create scraper for %s: %v", account.UserID, err)
			failCount++
			continue
		}

		log.Printf("Initializing browser for %s...", account.UserID)
		if err := s.Initialize(); err != nil {
			log.Printf("Failed to initialize scraper for %s: %v", account.UserID, err)
			s.Close()
			failCount++
			continue
		}

		// Process the account
		success := processAccount(s, account, fromDate, toDate)
		s.Close()

		if success {
			successCount++
			log.Printf("✅ Successfully processed account: %s", account.UserID)
		} else {
			failCount++
			log.Printf("❌ Failed to process account: %s", account.UserID)
		}

		// Add delay between accounts to avoid rate limiting
		if i < len(accounts)-1 {
			log.Println("Waiting 5 seconds before next account...")
			time.Sleep(5 * time.Second)
		}
	}

	// Summary
	log.Printf("\n========================================")
	log.Printf("SUMMARY")
	log.Printf("========================================")
	log.Printf("Total accounts: %d", len(accounts))
	log.Printf("✅ Successful: %d", successCount)
	log.Printf("❌ Failed: %d", failCount)
	log.Printf("========================================")
}

func processAccount(s *scraper.ActualETCScraper, account config.SimpleAccount, fromDate, toDate time.Time) bool {
	// Login
	log.Printf("Attempting login for %s...", account.UserID)
	if err := s.Login(); err != nil {
		log.Printf("Login failed for %s: %v", account.UserID, err)
		return false
	}

	log.Printf("Login successful for %s!", account.UserID)

	// Download ETC statements
	log.Printf("Downloading ETC statements for %s...", account.UserID)
	csvPath, err := s.SearchAndDownloadCSV(fromDate, toDate)
	if err != nil {
		log.Printf("Failed to download ETC meisai for %s: %v", account.UserID, err)
		return false
	}

	log.Printf("✅ Download completed for %s!", account.UserID)
	log.Printf("CSV file saved to: %s", csvPath)

	// Check if file exists and show summary
	if info, err := os.Stat(csvPath); err == nil {
		log.Printf("File size: %d bytes", info.Size())
		displayCSVSummary(csvPath, account.UserID)
	}

	return true
}

func displayCSVSummary(csvPath string, userID string) {
	file, err := os.Open(csvPath)
	if err != nil {
		log.Printf("Failed to open CSV: %v", err)
		return
	}
	defer file.Close()

	// Create Shift-JIS decoder
	reader := transform.NewReader(file, japanese.ShiftJIS.NewDecoder())
	csvReader := csv.NewReader(reader)

	log.Printf("\n=== CSV Summary for %s ===", userID)

	var firstDate, lastDate string
	rowCount := 0

	for {
		record, err := csvReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Printf("Error reading CSV: %v", err)
			break
		}

		rowCount++

		// Skip header
		if rowCount == 1 {
			continue
		}

		// Get date from first column (利用年月日（自）)
		if record[0] != "" {
			if firstDate == "" {
				firstDate = record[0]
			}
			lastDate = record[0]
		}

		// Try to sum amounts (assuming column 8 is amount)
		if len(record) > 8 {
			// Parse amount (remove commas and convert)
			// This is simplified - you may need better parsing
			_ = record[8] // Just acknowledge it exists
		}
	}

	log.Printf("📊 Total records: %d", rowCount-1)
	log.Printf("📅 Date range: %s to %s", firstDate, lastDate)
	log.Printf("===================================")
}
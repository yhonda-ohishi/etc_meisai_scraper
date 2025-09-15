package main

import (
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/yhonda-ohishi/etc_meisai/config"
	"github.com/yhonda-ohishi/etc_meisai/scraper"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not loaded")
	}

	// Load corporate accounts from environment
	accounts, err := config.LoadCorporateAccountsFromEnv()
	if err != nil || len(accounts) == 0 {
		log.Fatal("No accounts found in environment variables")
	}

	account := accounts[0]
	log.Printf("Using account: %s", account.UserID)

	// Calculate date range (last month)
	now := time.Now()
	lastMonth := now.AddDate(0, -1, 0)
	fromDate := time.Date(lastMonth.Year(), lastMonth.Month(), 1, 0, 0, 0, 0, time.Local)
	toDate := time.Date(lastMonth.Year(), lastMonth.Month()+1, 0, 0, 0, 0, 0, time.Local)

	log.Printf("Date range: %s to %s", fromDate.Format("2006-01-02"), toDate.Format("2006-01-02"))

	// Create download directory
	downloadPath := "./test_downloads"
	if err := os.MkdirAll(downloadPath, 0755); err != nil {
		log.Fatalf("Failed to create download directory: %v", err)
	}

	// Configure scraper with improved settings
	config := &scraper.ScraperConfig{
		UserID:       account.UserID,
		Password:     account.Password,
		DownloadPath: downloadPath,
		Headless:     false, // Set to false to see the browser
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
		log.Fatalf("Failed to create scraper: %v", err)
	}

	log.Println("Initializing browser...")
	if err := s.Initialize(); err != nil {
		log.Fatalf("Failed to initialize scraper: %v", err)
	}
	defer s.Close()

	log.Println("Attempting login...")
	if err := s.Login(); err != nil {
		log.Fatalf("Login failed: %v", err)
	}

	log.Println("Login successful!")

	log.Println("Downloading ETC statements...")
	csvPath, err := s.SearchAndDownloadCSV(fromDate, toDate)
	if err != nil {
		log.Fatalf("Failed to download ETC meisai: %v", err)
	}

	log.Printf("✅ Download completed successfully!")
	log.Printf("CSV file saved to: %s", csvPath)

	// Check if file exists and show size
	if info, err := os.Stat(csvPath); err == nil {
		log.Printf("File size: %d bytes", info.Size())
	}

	log.Println("Test completed successfully!")
}
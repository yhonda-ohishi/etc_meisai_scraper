package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/yhonda-ohishi/etc_meisai/config"
	"github.com/yhonda-ohishi/etc_meisai/parser"
	"github.com/yhonda-ohishi/etc_meisai/scraper"
)

func main() {
	// Load .env file
	if err := godotenv.Load("../../.env"); err != nil {
		if err := godotenv.Load(".env"); err != nil {
			log.Println("No .env file found")
		}
	}

	// Command line flags
	var (
		fromDate = flag.String("from", "", "From date (YYYY-MM-DD)")
		toDate   = flag.String("to", "", "To date (YYYY-MM-DD)")
		headless = flag.Bool("headless", false, "Run in headless mode")
		download = flag.String("download", "./downloads", "Download directory")
	)
	flag.Parse()

	// Parse dates
	var from, to time.Time
	var err error

	if *fromDate == "" {
		// Default to last month's first day
		now := time.Now()
		// Go to previous month
		lastMonth := now.AddDate(0, -1, 0)
		from = time.Date(lastMonth.Year(), lastMonth.Month(), 1, 0, 0, 0, 0, time.Local)
	} else {
		from, err = time.Parse("2006-01-02", *fromDate)
		if err != nil {
			log.Fatalf("Invalid from date: %v", err)
		}
	}

	if *toDate == "" {
		to = time.Now()
	} else {
		to, err = time.Parse("2006-01-02", *toDate)
		if err != nil {
			log.Fatalf("Invalid to date: %v", err)
		}
	}

	// Load corporate accounts
	accounts, err := config.LoadCorporateAccountsFromEnv()
	if err != nil {
		log.Fatalf("Failed to load accounts: %v", err)
	}

	log.Printf("Found %d corporate accounts", len(accounts))
	log.Printf("Date range: %s to %s", from.Format("2006-01-02"), to.Format("2006-01-02"))

	// Process each account
	for i, account := range accounts {
		log.Printf("\n=== Processing Account %d/%d: %s ===", i+1, len(accounts), account.UserID)

		// Create account-specific download directory
		accountDir := fmt.Sprintf("%s/%s_%s", *download, account.UserID, time.Now().Format("20060102_150405"))
		if err := os.MkdirAll(accountDir, 0755); err != nil {
			log.Printf("Failed to create directory: %v", err)
			continue
		}

		// Setup scraper
		scraperConfig := &scraper.ScraperConfig{
			UserID:       account.UserID,
			Password:     account.Password,
			DownloadPath: accountDir,
			Headless:     *headless,
			Timeout:      30000,
		}

		etcScraper, err := scraper.NewActualETCScraper(scraperConfig)
		if err != nil {
			log.Printf("Failed to create scraper: %v", err)
			continue
		}

		// Initialize
		if err := etcScraper.Initialize(); err != nil {
			log.Printf("Failed to initialize scraper: %v", err)
			etcScraper.Close()
			continue
		}

		// Login
		if err := etcScraper.Login(); err != nil {
			log.Printf("Login failed: %v", err)
			etcScraper.Close()
			continue
		}

		// Search and download CSV
		csvPath, err := etcScraper.SearchAndDownloadCSV(from, to)
		if err != nil {
			log.Printf("Failed to download CSV: %v", err)
			etcScraper.Close()
			continue
		}

		log.Printf("CSV downloaded: %s", csvPath)

		// Parse CSV
		csvParser := parser.NewETCCSVParser()
		records, err := csvParser.ParseFile(csvPath)
		if err != nil {
			log.Printf("Failed to parse CSV: %v", err)
		} else {
			log.Printf("Parsed %d records", len(records))
		}

		// Clean up
		etcScraper.Close()

		// Delay between accounts
		if i < len(accounts)-1 {
			log.Println("Waiting 5 seconds before next account...")
			time.Sleep(5 * time.Second)
		}
	}

	log.Println("\n=== All accounts processed ===")
}
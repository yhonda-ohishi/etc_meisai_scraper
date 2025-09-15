package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/yhonda-ohishi/etc_meisai/config"
	"github.com/yhonda-ohishi/etc_meisai/scraper"
)

func main() {
	// Load .env file
	if err := godotenv.Load("../../.env"); err != nil {
		log.Println("No .env file found")
	}

	// Command line flags
	var (
		fromDate     = flag.String("from", "", "From date (YYYY-MM-DD)")
		toDate       = flag.String("to", "", "To date (YYYY-MM-DD)")
		accountsFile = flag.String("accounts", "", "Path to accounts JSON file")
		headless     = flag.Bool("headless", true, "Run in headless mode")
		downloadPath = flag.String("download", "./downloads", "Download directory")
		accountType  = flag.String("type", "all", "Account type to process (all/corporate/personal)")
	)
	flag.Parse()

	// Parse dates
	var from, to time.Time
	var err error

	if *fromDate == "" {
		// Default to current month
		now := time.Now()
		from = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.Local)
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

	// Load accounts
	var accountsConfig *config.AccountsConfig

	if *accountsFile != "" {
		// Load from file
		accountsConfig, err = config.LoadAccountsFromFile(*accountsFile)
		if err != nil {
			log.Fatalf("Failed to load accounts from file: %v", err)
		}
	} else {
		// Load from environment variables
		accountsConfig, err = config.LoadAccountsFromEnv()
		if err != nil {
			log.Fatalf("Failed to load accounts from env: %v", err)
		}
	}

	if len(accountsConfig.Accounts) == 0 {
		log.Fatal("No accounts configured")
	}

	// Filter accounts by type
	var accounts []config.ETCAccount
	switch *accountType {
	case "corporate":
		accounts = accountsConfig.GetCorporateAccounts()
	case "personal":
		accounts = accountsConfig.GetPersonalAccounts()
	default:
		accounts = accountsConfig.GetActiveAccounts()
	}

	log.Printf("Processing %d accounts from %s to %s", len(accounts), from.Format("2006-01-02"), to.Format("2006-01-02"))

	// Create multi-account scraper
	multiScraper := scraper.NewMultiAccountScraper(accounts, *downloadPath, *headless)

	// Start scraping
	startTime := time.Now()
	results := multiScraper.ScrapeAll(from, to)

	// Print results summary
	fmt.Println("\n=== Scraping Results ===")
	fmt.Printf("Total accounts processed: %d\n", len(results))
	fmt.Printf("Time taken: %v\n\n", time.Since(startTime))

	successCount := 0
	totalRecords := 0

	for _, result := range results {
		status := "✓"
		if !result.Success {
			status = "✗"
		} else {
			successCount++
			totalRecords += len(result.Records)
		}

		fmt.Printf("%s %s (%s)\n", status, result.Account.Name, result.Account.Type)

		if result.Success {
			fmt.Printf("  Records: %d\n", len(result.Records))
			if result.CSVPath != "" {
				fmt.Printf("  CSV: %s\n", result.CSVPath)
			}
		} else {
			fmt.Printf("  Error: %v\n", result.Error)
		}
	}

	fmt.Printf("\n=== Summary ===\n")
	fmt.Printf("Successful: %d/%d\n", successCount, len(results))
	fmt.Printf("Total records: %d\n", totalRecords)

	// Create consolidated CSV if multiple accounts succeeded
	if successCount > 1 {
		consolidatedPath := fmt.Sprintf("%s/consolidated_%s.csv", *downloadPath, time.Now().Format("20060102_150405"))
		if err := saveConsolidatedCSV(results, consolidatedPath); err != nil {
			log.Printf("Failed to save consolidated CSV: %v", err)
		} else {
			fmt.Printf("Consolidated CSV: %s\n", consolidatedPath)
		}
	}

	// Exit with error if any account failed
	if successCount < len(results) {
		os.Exit(1)
	}
}

// saveConsolidatedCSV saves all records to a single CSV file
func saveConsolidatedCSV(results []scraper.ScrapingResult, filepath string) error {
	// TODO: Implement CSV consolidation
	// This would combine all records from successful results into one CSV file
	log.Println("CSV consolidation not yet implemented")
	return nil
}
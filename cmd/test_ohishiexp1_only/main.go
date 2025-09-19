package main

import (
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/yhonda-ohishi/etc_meisai/scraper"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not loaded")
	}

	// Test only ohishiexp1 account
	userID := os.Getenv("TEST_ETC_USER")
	password := os.Getenv("TEST_ETC_PASSWORD")

	log.Printf("Testing account: %s", userID)

	// Calculate date range
	now := time.Now()
	today := now
	lastMonth := now.AddDate(0, -1, 0)
	fromDate := time.Date(lastMonth.Year(), lastMonth.Month(), 1, 0, 0, 0, 0, time.Local)
	toDate := time.Date(today.Year(), today.Month(), today.Day(), 23, 59, 59, 0, time.Local)

	log.Printf("📅 Target date range: %s to %s",
		fromDate.Format("2006-01-02"), toDate.Format("2006-01-02"))

	// Create download directory
	downloadPath := "./test_downloads/debug_ohishiexp1"
	if err := os.MkdirAll(downloadPath, 0755); err != nil {
		log.Fatal("Failed to create download directory:", err)
	}

	// Configure scraper with debug mode
	config := &scraper.ScraperConfig{
		UserID:       userID,
		Password:     password,
		DownloadPath: downloadPath,
		Headless:     false, // Show browser for debugging
		Timeout:      60000, // Longer timeout for debugging
		SlowMo:       500,   // Slow down actions to see what's happening
		UserAgent:    "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
		Viewport: &scraper.ViewportSize{
			Width:  1920,
			Height: 1080,
		},
	}

	// Create scraper
	s, err := scraper.NewActualETCScraper(config)
	if err != nil {
		log.Fatal("Failed to create scraper:", err)
	}
	defer s.Close()

	log.Println("Initializing browser...")
	if err := s.Initialize(); err != nil {
		log.Fatal("Failed to initialize:", err)
	}

	// Login
	log.Println("Attempting login...")
	if err := s.Login(); err != nil {
		log.Fatal("Login failed:", err)
	}
	log.Println("✅ Login successful!")

	// Add debug code to understand the page structure after login
	log.Println("=== DEBUGGING PAGE STRUCTURE ===")

	// Wait for page to stabilize
	time.Sleep(3 * time.Second)

	// Try SearchAndDownloadCSV which will show us what's happening
	log.Println("Attempting to download with date range...")
	csvPath, err := s.SearchAndDownloadCSV(fromDate, toDate)
	if err != nil {
		log.Printf("Download failed: %v", err)
	} else {
		log.Printf("✅ Download successful: %s", csvPath)

		// Check file info
		if info, err := os.Stat(csvPath); err == nil {
			log.Printf("File size: %d bytes", info.Size())
		}
	}

	log.Println("Test completed. Browser will remain open for 10 seconds...")
	time.Sleep(10 * time.Second)
}
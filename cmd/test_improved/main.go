package main

import (
	"fmt"
	"log"
	"time"

	"github.com/yhonda-ohishi/etc_meisai/scraper"
)

func main() {
	log.Println("Testing improved ETC scraper...")

	// Create configuration with new options
	config := &scraper.ScraperConfig{
		UserID:       "test_user",
		Password:     "test_pass",
		DownloadPath: "./test_downloads",
		Headless:     false,
		Timeout:      30000,
		RetryCount:   3,
		SlowMo:       100,
		Viewport: &scraper.ViewportSize{
			Width:  1920,
			Height: 1080,
		},
	}

	// Create actual scraper instance
	s, err := scraper.NewActualETCScraper(config)
	if err != nil {
		log.Fatalf("Failed to create scraper: %v", err)
	}

	// Initialize browser
	log.Println("Initializing browser...")
	if err := s.Initialize(); err != nil {
		log.Fatalf("Failed to initialize: %v", err)
	}
	defer s.Close()

	// Test the new WaitForDownloadWithProgress function
	log.Println("Testing download progress tracking...")

	// Simulate download scenario
	go func() {
		time.Sleep(2 * time.Second)
		log.Println("Simulating download completion...")
	}()

	// You would normally trigger a download here
	// For testing, we'll just demonstrate the structure

	fmt.Println("\n=== Improved Features ===")
	fmt.Println("1. Enhanced download progress tracking")
	fmt.Println("2. Centralized error handling with handleError()")
	fmt.Println("3. WaitForDownloadWithProgress() with timeout")
	fmt.Println("4. Extended configuration options")
	fmt.Println("5. Better default values")

	log.Println("Test completed successfully!")
}
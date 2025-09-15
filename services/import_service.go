package services

import (
	"fmt"
	"log"
	"time"

	"github.com/yhonda-ohishi/etc_meisai/models"
	"github.com/yhonda-ohishi/etc_meisai/parser"
	"github.com/yhonda-ohishi/etc_meisai/repositories"
	"github.com/yhonda-ohishi/etc_meisai/scraper"
)

// ImportService handles the full import process
type ImportService struct {
	repo     *repositories.ETCRepository
	scraper  *scraper.ETCScraper
	parser   *parser.ETCCSVParser
}

// NewImportService creates a new import service
func NewImportService(repo *repositories.ETCRepository) *ImportService {
	return &ImportService{
		repo:   repo,
		parser: parser.NewETCCSVParser(),
	}
}

// ImportFromWeb downloads and imports data from ETC website
func (s *ImportService) ImportFromWeb(userID, password string, fromDate, toDate time.Time, cardNo string) (*models.ETCImportResult, error) {
	// Initialize scraper
	config := &scraper.ScraperConfig{
		UserID:       userID,
		Password:     password,
		DownloadPath: "./downloads",
		Headless:     true,
		Timeout:      30000,
	}

	scraper, err := scraper.NewETCScraper(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create scraper: %w", err)
	}
	defer scraper.Close()

	// Initialize browser
	if err := scraper.Initialize(); err != nil {
		return nil, fmt.Errorf("failed to initialize scraper: %w", err)
	}

	// Login
	log.Println("Logging in to ETC service...")
	if err := scraper.Login(); err != nil {
		return nil, fmt.Errorf("login failed: %w", err)
	}

	// Search and download
	log.Println("Searching and downloading CSV...")
	csvPath, err := scraper.SearchAndDownload(fromDate, toDate, cardNo)
	if err != nil {
		return nil, fmt.Errorf("failed to download CSV: %w", err)
	}

	// Parse CSV
	log.Printf("Parsing CSV file: %s", csvPath)
	records, err := s.parser.ParseFile(csvPath)
	if err != nil {
		return nil, fmt.Errorf("failed to parse CSV: %w", err)
	}

	// Import to database
	log.Printf("Importing %d records to database...", len(records))
	if err := s.repo.BulkInsert(records); err != nil {
		return nil, fmt.Errorf("failed to import to database: %w", err)
	}

	return &models.ETCImportResult{
		Success:      true,
		ImportedRows: len(records),
		Message:      fmt.Sprintf("Successfully imported %d records", len(records)),
		ImportedAt:   time.Now(),
	}, nil
}

// ImportFromCSV imports data from a CSV file
func (s *ImportService) ImportFromCSV(filePath string) (*models.ETCImportResult, error) {
	// Parse CSV
	log.Printf("Parsing CSV file: %s", filePath)
	records, err := s.parser.ParseFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to parse CSV: %w", err)
	}

	if len(records) == 0 {
		return &models.ETCImportResult{
			Success:      false,
			ImportedRows: 0,
			Message:      "No records found in CSV file",
			ImportedAt:   time.Now(),
		}, nil
	}

	// Import to database
	log.Printf("Importing %d records to database...", len(records))
	if err := s.repo.BulkInsert(records); err != nil {
		return nil, fmt.Errorf("failed to import to database: %w", err)
	}

	return &models.ETCImportResult{
		Success:      true,
		ImportedRows: len(records),
		Message:      fmt.Sprintf("Successfully imported %d records from CSV", len(records)),
		ImportedAt:   time.Now(),
	}, nil
}
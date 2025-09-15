package etc_meisai

import (
	"fmt"
	"time"

	"github.com/yhonda-ohishi/etc_meisai/config"
	"github.com/yhonda-ohishi/etc_meisai/models"
	"github.com/yhonda-ohishi/etc_meisai/parser"
	"github.com/yhonda-ohishi/etc_meisai/scraper"
)

// ETCClient is the main client for ETC operations
type ETCClient struct {
	config *ClientConfig
}

// ClientConfig holds configuration for the ETC client
type ClientConfig struct {
	DownloadPath string
	Headless     bool
	Timeout      float64
	RetryCount   int
}

// NewETCClient creates a new ETC client
func NewETCClient(config *ClientConfig) *ETCClient {
	if config == nil {
		config = &ClientConfig{
			DownloadPath: "./downloads",
			Headless:     true,
			Timeout:      30000,
			RetryCount:   3,
		}
	}
	return &ETCClient{config: config}
}

// DownloadResult represents the result of a download operation
type DownloadResult struct {
	Account  string
	CSVPath  string
	Records  []models.ETCMeisai
	Success  bool
	Error    error
}

// DownloadETCData downloads ETC data for specified accounts and date range
func (c *ETCClient) DownloadETCData(accounts []config.SimpleAccount, fromDate, toDate time.Time) ([]DownloadResult, error) {
	results := make([]DownloadResult, 0, len(accounts))

	for _, account := range accounts {
		result := c.downloadForAccount(account, fromDate, toDate)
		results = append(results, result)
	}

	return results, nil
}

// DownloadETCDataSingle downloads ETC data for a single account
func (c *ETCClient) DownloadETCDataSingle(userID, password string, fromDate, toDate time.Time) (*DownloadResult, error) {
	account := config.SimpleAccount{
		UserID:   userID,
		Password: password,
	}
	result := c.downloadForAccount(account, fromDate, toDate)
	return &result, nil
}

// downloadForAccount downloads data for a single account
func (c *ETCClient) downloadForAccount(account config.SimpleAccount, fromDate, toDate time.Time) DownloadResult {
	result := DownloadResult{
		Account: account.UserID,
		Success: false,
	}

	// Configure scraper
	scraperConfig := &scraper.ScraperConfig{
		UserID:       account.UserID,
		Password:     account.Password,
		DownloadPath: c.config.DownloadPath,
		Headless:     c.config.Headless,
		Timeout:      c.config.Timeout,
		RetryCount:   c.config.RetryCount,
	}

	// Create and initialize scraper
	s, err := scraper.NewActualETCScraper(scraperConfig)
	if err != nil {
		result.Error = fmt.Errorf("failed to create scraper: %w", err)
		return result
	}
	defer s.Close()

	if err := s.Initialize(); err != nil {
		result.Error = fmt.Errorf("failed to initialize scraper: %w", err)
		return result
	}

	// Login
	if err := s.Login(); err != nil {
		result.Error = fmt.Errorf("login failed: %w", err)
		return result
	}

	// Download CSV
	csvPath, err := s.SearchAndDownloadCSV(fromDate, toDate)
	if err != nil {
		result.Error = fmt.Errorf("failed to download CSV: %w", err)
		return result
	}

	result.CSVPath = csvPath

	// Parse CSV
	csvParser := parser.NewETCCSVParser()
	records, err := csvParser.ParseFile(csvPath)
	if err != nil {
		result.Error = fmt.Errorf("failed to parse CSV: %w", err)
		return result
	}

	result.Records = records
	result.Success = true
	return result
}

// ParseETCCSV parses an ETC CSV file and returns records
func ParseETCCSV(csvPath string) ([]models.ETCMeisai, error) {
	parser := parser.NewETCCSVParser()
	return parser.ParseFile(csvPath)
}

// LoadCorporateAccounts loads corporate accounts from environment
func LoadCorporateAccounts() ([]config.SimpleAccount, error) {
	return config.LoadCorporateAccountsFromEnv()
}

// LoadPersonalAccounts loads personal accounts from environment
func LoadPersonalAccounts() ([]config.SimpleAccount, error) {
	return config.LoadPersonalAccountsFromEnv()
}
package scraper

import (
	"fmt"
	"log"
	"path/filepath"
	"sync"
	"time"

	"github.com/playwright-community/playwright-go"
	"github.com/yhonda-ohishi/etc_meisai/config"
	"github.com/yhonda-ohishi/etc_meisai/models"
	"github.com/yhonda-ohishi/etc_meisai/parser"
)

// MultiAccountScraper handles scraping for multiple accounts
type MultiAccountScraper struct {
	accounts     []config.ETCAccount
	downloadPath string
	headless     bool
	concurrent   int // Number of concurrent scrapers
}

// ScrapingResult represents the result for one account
type ScrapingResult struct {
	Account     config.ETCAccount
	Records     []models.ETCMeisai
	CSVPath     string
	Success     bool
	Error       error
	ProcessedAt time.Time
}

// NewMultiAccountScraper creates a new multi-account scraper
func NewMultiAccountScraper(accounts []config.ETCAccount, downloadPath string, headless bool) *MultiAccountScraper {
	return &MultiAccountScraper{
		accounts:     accounts,
		downloadPath: downloadPath,
		headless:     headless,
		concurrent:   2, // Default to 2 concurrent scrapers
	}
}

// ScrapeAll scrapes all accounts and returns results
func (m *MultiAccountScraper) ScrapeAll(fromDate, toDate time.Time) []ScrapingResult {
	results := make([]ScrapingResult, 0, len(m.accounts))
	resultsChan := make(chan ScrapingResult, len(m.accounts))

	// Create a semaphore for concurrency control
	sem := make(chan struct{}, m.concurrent)
	var wg sync.WaitGroup

	for _, account := range m.accounts {
		wg.Add(1)
		go func(acc config.ETCAccount) {
			defer wg.Done()

			// Acquire semaphore
			sem <- struct{}{}
			defer func() { <-sem }()

			log.Printf("Processing account: %s (%s)", acc.Name, acc.Type)

			result := m.scrapeAccount(acc, fromDate, toDate)
			resultsChan <- result
		}(account)
	}

	// Close results channel when all goroutines complete
	go func() {
		wg.Wait()
		close(resultsChan)
	}()

	// Collect results
	for result := range resultsChan {
		results = append(results, result)
	}

	return results
}

// scrapeAccount scrapes a single account
func (m *MultiAccountScraper) scrapeAccount(account config.ETCAccount, fromDate, toDate time.Time) ScrapingResult {
	result := ScrapingResult{
		Account:     account,
		ProcessedAt: time.Now(),
		Success:     false,
	}

	// Create account-specific download directory
	accountDir := filepath.Join(m.downloadPath, fmt.Sprintf("%s_%s", account.Type, account.Name))

	// Setup scraper config based on account type
	scraperConfig := &ScraperConfig{
		UserID:       account.UserID,
		Password:     account.Password,
		DownloadPath: accountDir,
		Headless:     m.headless,
		Timeout:      30000,
	}

	// Create appropriate scraper based on account type
	var scraper interface {
		Initialize() error
		Close()
		LoginCorporate() error
		LoginPersonal() error
		SearchAndDownloadCSV(time.Time, time.Time) (string, error)
	}

	if account.Type == config.AccountTypeCorporate {
		// Use corporate scraper
		corpScraper := &CorporateETCScraper{
			config:       scraperConfig,
			corpPassword: account.PasswordCorp,
		}
		scraper = corpScraper
	} else {
		// Use personal scraper
		actualScraper, err := NewActualETCScraper(scraperConfig)
		if err != nil {
			result.Error = fmt.Errorf("failed to create scraper: %w", err)
			return result
		}
		scraper = actualScraper
	}

	// Initialize scraper
	if err := scraper.Initialize(); err != nil {
		result.Error = fmt.Errorf("failed to initialize scraper: %w", err)
		return result
	}
	defer scraper.Close()

	// Login based on account type
	var loginErr error
	if account.Type == config.AccountTypeCorporate {
		loginErr = scraper.LoginCorporate()
	} else {
		loginErr = scraper.LoginPersonal()
	}

	if loginErr != nil {
		result.Error = fmt.Errorf("login failed: %w", loginErr)
		return result
	}

	// Process each card number if specified
	if len(account.CardNumbers) > 0 {
		allRecords := []models.ETCMeisai{}

		for _, cardNo := range account.CardNumbers {
			log.Printf("Processing card: %s", cardNo)

			csvPath, err := scraper.SearchAndDownloadCSV(fromDate, toDate)
			if err != nil {
				log.Printf("Failed to download CSV for card %s: %v", cardNo, err)
				continue
			}

			// Parse CSV
			parser := parser.NewETCCSVParser()
			records, err := parser.ParseFile(csvPath)
			if err != nil {
				log.Printf("Failed to parse CSV for card %s: %v", cardNo, err)
				continue
			}

			// Add account info to records
			for i := range records {
				records[i].CardNo = cardNo
			}

			allRecords = append(allRecords, records...)
			result.CSVPath = csvPath
		}

		result.Records = allRecords
	} else {
		// No specific cards, download all
		csvPath, err := scraper.SearchAndDownloadCSV(fromDate, toDate)
		if err != nil {
			result.Error = fmt.Errorf("failed to download CSV: %w", err)
			return result
		}

		// Parse CSV
		parser := parser.NewETCCSVParser()
		records, err := parser.ParseFile(csvPath)
		if err != nil {
			result.Error = fmt.Errorf("failed to parse CSV: %w", err)
			return result
		}

		result.Records = records
		result.CSVPath = csvPath
	}

	result.Success = true
	log.Printf("Successfully processed account %s: %d records", account.Name, len(result.Records))

	return result
}

// CorporateETCScraper handles corporate accounts with two passwords
type CorporateETCScraper struct {
	*ActualETCScraper
	config       *ScraperConfig
	corpPassword string
}

// Initialize initializes the corporate scraper
func (c *CorporateETCScraper) Initialize() error {
	scraper, err := NewActualETCScraper(c.config)
	if err != nil {
		return err
	}
	c.ActualETCScraper = scraper
	return c.ActualETCScraper.Initialize()
}

// LoginCorporate performs corporate login with two passwords
func (c *CorporateETCScraper) LoginCorporate() error {
	log.Println("Performing corporate login...")

	// Navigate to corporate login page
	loginURL := "https://www2.etc-meisai.jp/etc/R?funccode=1013000000&nextfunc=1013000000"

	_, err := c.page.Goto(loginURL, playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateNetworkidle,
		Timeout:   playwright.Float(30000),
	})
	if err != nil {
		return fmt.Errorf("failed to navigate: %w", err)
	}

	// Fill user ID
	userSelectors := []string{
		"input[name='usrid']",
		"input[name='userId']",
		"#userId",
		"input[type='text']:not([type='hidden'])",
	}

	userFilled := false
	for _, selector := range userSelectors {
		if count, _ := c.page.Locator(selector).Count(); count > 0 {
			if err := c.page.Fill(selector, c.config.UserID); err == nil {
				userFilled = true
				break
			}
		}
	}

	if !userFilled {
		return fmt.Errorf("could not fill user ID")
	}

	// Fill first password
	pass1Selectors := []string{
		"input[name='password']",
		"input[name='password1']",
		"#password1",
		"input[type='password']:nth-of-type(1)",
	}

	for _, selector := range pass1Selectors {
		if count, _ := c.page.Locator(selector).Count(); count > 0 {
			c.page.Fill(selector, c.config.Password)
			break
		}
	}

	// Fill second password (corporate)
	pass2Selectors := []string{
		"input[name='password2']",
		"input[name='corpPassword']",
		"#password2",
		"input[type='password']:nth-of-type(2)",
	}

	for _, selector := range pass2Selectors {
		if count, _ := c.page.Locator(selector).Count(); count > 0 {
			c.page.Fill(selector, c.corpPassword)
			break
		}
	}

	// Click login button
	loginSelectors := []string{
		"input[type='submit'][value*='ログイン']",
		"button[type='submit']",
		"input[type='submit']",
	}

	for _, selector := range loginSelectors {
		if count, _ := c.page.Locator(selector).Count(); count > 0 {
			c.page.Click(selector)
			break
		}
	}

	// Wait for login
	time.Sleep(time.Second * 5)

	// Check if login was successful
	if count, _ := c.page.Locator("*:has-text('ログアウト')").Count(); count > 0 {
		log.Println("Corporate login successful")
		return nil
	}

	return fmt.Errorf("corporate login failed")
}

// LoginPersonal is for interface compatibility
func (c *CorporateETCScraper) LoginPersonal() error {
	return c.Login()
}

// Ensure ActualETCScraper implements the required interface methods
func (s *ActualETCScraper) LoginCorporate() error {
	return fmt.Errorf("corporate login not supported for personal scraper")
}

func (s *ActualETCScraper) LoginPersonal() error {
	return s.Login()
}
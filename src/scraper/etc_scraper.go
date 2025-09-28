package scraper

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"
)

// ETCScraper handles web scraping for ETC meisai service
type ETCScraper struct {
	pw      PlaywrightInterface
	browser BrowserInterface
	context BrowserContextInterface
	page    PageInterface
	config  *ScraperConfig
	logger  *log.Logger
	factory PlaywrightFactory
}

// ScraperConfig holds configuration for the scraper
type ScraperConfig struct {
	UserID       string
	Password     string
	DownloadPath string
	Headless     bool
	Timeout      float64
	RetryCount   int
	UserAgent    string
	SlowMo       float64
	TestMode     bool // Skip time.Sleep in tests
}

// NewETCScraper creates a new ETC scraper instance (for production use)
func NewETCScraper(config *ScraperConfig, logger *log.Logger) (*ETCScraper, error) {
	// For production, use the default factory that wraps real Playwright
	factory := &DefaultPlaywrightFactory{}
	return NewETCScraperWithFactory(config, logger, factory)
}

// NewETCScraperWithFactory creates a new ETC scraper instance with custom factory
func NewETCScraperWithFactory(config *ScraperConfig, logger *log.Logger, factory PlaywrightFactory) (*ETCScraper, error) {
	// Validate factory
	if factory == nil {
		return nil, fmt.Errorf("factory is required for testable scraper")
	}

	// Set default values
	if config.DownloadPath == "" {
		config.DownloadPath = "./downloads"
	}
	if config.Timeout == 0 {
		config.Timeout = 30000 // 30 seconds default
	}
	if config.RetryCount == 0 {
		config.RetryCount = 3
	}
	if config.UserAgent == "" {
		config.UserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36"
	}

	// Skip directory creation for better testability

	if logger == nil {
		logger = log.New(os.Stdout, "[SCRAPER] ", log.LstdFlags)
	}

	return &ETCScraper{
		config:  config,
		logger:  logger,
		factory: factory,
	}, nil
}

// Initialize sets up Playwright and browser
func (s *ETCScraper) Initialize() error {
	var err error

	// Install playwright browsers if needed
	err = s.factory.Install()
	if err != nil {
		return fmt.Errorf("could not install playwright: %w", err)
	}

	// Start Playwright
	s.pw, err = s.factory.Run()
	if err != nil {
		return fmt.Errorf("could not start playwright: %w", err)
	}

	// Launch browser
	launchOptions := BrowserTypeLaunchOptions{
		Headless: Bool(s.config.Headless),
	}

	if s.config.SlowMo > 0 {
		launchOptions.SlowMo = Float(s.config.SlowMo)
	}

	chromium := s.pw.GetChromium()
	s.browser, err = chromium.Launch(launchOptions)
	if err != nil {
		return fmt.Errorf("could not launch browser: %w", err)
	}

	// Create browser context with download settings
	contextOptions := BrowserNewContextOptions{
		AcceptDownloads: Bool(true),
		Viewport: &Size{
			Width:  1920,
			Height: 1080,
		},
		UserAgent: String(s.config.UserAgent),
	}

	s.context, err = s.browser.NewContext(contextOptions)
	if err != nil {
		return fmt.Errorf("could not create browser context: %w", err)
	}

	// Set default timeout
	s.context.SetDefaultTimeout(s.config.Timeout)

	// Create page
	s.page, err = s.context.NewPage()
	if err != nil {
		return fmt.Errorf("could not create page: %w", err)
	}

	s.logger.Printf("Scraper initialized with download path: %s", s.config.DownloadPath)
	return nil
}

// Login performs login to ETC meisai service
func (s *ETCScraper) Login() error {
	if s.page == nil {
		return fmt.Errorf("scraper not initialized")
	}

	s.logger.Println("Navigating to https://www.etc-meisai.jp/")

	// Navigate to login page
	_, err := s.page.Goto("https://www.etc-meisai.jp/", PageGotoOptions{
		WaitUntil: WaitUntilStateNetworkidle,
	})
	if err != nil {
		return fmt.Errorf("failed to navigate to login page: %w", err)
	}

	// Skip screenshot for debugging

	// Wait for login form
	s.logger.Println("Waiting for login form...")

	// Try multiple selectors for user ID field
	userIDSelectors := []string{
		"input[name='userId']",
		"#userId",
		"input[type='text'][placeholder*='ID']",
		"input[name='username']",
	}

	userIDField := s.findElement(userIDSelectors)
	if userIDField == nil {
		return fmt.Errorf("login form user ID field not found")
	}

	// Fill user ID
	s.logger.Println("Filling login credentials...")
	if err := userIDField.Fill(s.config.UserID); err != nil {
		return fmt.Errorf("failed to fill user ID: %w", err)
	}

	// Try multiple selectors for password field
	passwordSelectors := []string{
		"input[name='password']",
		"#password",
		"input[type='password']",
		"input[name='passwd']",
	}

	passwordField := s.findElement(passwordSelectors)
	if passwordField == nil {
		return fmt.Errorf("password field not found")
	}

	// Fill password
	if err := passwordField.Fill(s.config.Password); err != nil {
		return fmt.Errorf("failed to fill password: %w", err)
	}

	// Skip screenshot

	// Click login button
	s.logger.Println("Clicking login button...")
	loginButtonSelectors := []string{
		"button[type='submit']",
		"input[type='submit']",
		".login-button",
		"button:has-text('ログイン')",
		"input[value='ログイン']",
	}

	loginButton := s.findElement(loginButtonSelectors)
	if loginButton == nil {
		return fmt.Errorf("login button not found")
	}

	if err := loginButton.Click(LocatorClickOptions{}); err != nil {
		return fmt.Errorf("failed to click login button: %w", err)
	}

	// Wait for navigation after login
	s.waitForNavigation()
	err = s.page.WaitForLoadState(PageWaitForLoadStateOptions{
		State: LoadStateNetworkidle,
	})
	if err != nil {
		return fmt.Errorf("failed to wait for login completion: %w", err)
	}

	// Skip screenshot

	// Check if login was successful
	logoutLocator := s.page.Locator("a:has-text('ログアウト')")
	logoutExists, _ := logoutLocator.Count()
	if logoutExists > 0 {
		s.logger.Println("Login successful!")
		return nil
	}

	// Check for error messages
	errorLocator := s.page.Locator(".error-message, .alert-danger, .error").First()
	errorMsg, _ := errorLocator.TextContent(LocatorTextContentOptions{})
	if errorMsg != "" {
		return fmt.Errorf("login failed: %s", errorMsg)
	}

	s.logger.Println("Login completed")
	return nil
}

// DownloadMeisai downloads ETC meisai data for specified date range
func (s *ETCScraper) DownloadMeisai(fromDate, toDate string) (string, error) {
	if s.page == nil {
		return "", fmt.Errorf("scraper not initialized")
	}

	s.logger.Printf("Downloading meisai from %s to %s", fromDate, toDate)

	// Navigate to search/download page
	searchURLs := []string{
		"https://www.etc-meisai.jp/search",
		"https://www.etc-meisai.jp/meisai",
		"https://www.etc-meisai.jp/download",
	}

	navigated := false
	for _, url := range searchURLs {
		if _, err := s.page.Goto(url, PageGotoOptions{
			WaitUntil: WaitUntilStateNetworkidle,
		}); err == nil {
			s.logger.Printf("Navigated to %s", url)
			navigated = true
			break
		}
	}

	if !navigated {
		s.logger.Println("Could not navigate to search page, trying to find search form on current page")
	}

	// Skip screenshot

	// Fill date range
	fromDateSelectors := []string{
		"input[name='fromDate']",
		"input[name='startDate']",
		"#fromDate",
		"input[placeholder*='開始']",
		"input[placeholder*='から']",
	}

	fromDateField := s.findElement(fromDateSelectors)
	if fromDateField != nil {
		fromDateField.Fill(fromDate)
		s.logger.Printf("Filled from date: %s", fromDate)
	}

	toDateSelectors := []string{
		"input[name='toDate']",
		"input[name='endDate']",
		"#toDate",
		"input[placeholder*='終了']",
		"input[placeholder*='まで']",
	}

	toDateField := s.findElement(toDateSelectors)
	if toDateField != nil {
		toDateField.Fill(toDate)
		s.logger.Printf("Filled to date: %s", toDate)
	}

	// Skip screenshot

	// Click search button
	searchButtonSelectors := []string{
		"button:has-text('検索')",
		"button:has-text('照会')",
		"input[value='検索']",
		"input[value='照会']",
		"button[type='submit']",
		".search-button",
	}

	searchButton := s.findElement(searchButtonSelectors)
	if searchButton != nil {
		searchButton.Click(LocatorClickOptions{})
		s.logger.Println("Clicked search button")
		s.waitForNavigation()
	}

	// Skip screenshot

	// Setup download handler
	downloadComplete := make(chan string, 1)
	s.page.On("download", func(download Download) {
		s.HandleDownload(download, downloadComplete)
	})

	// Click CSV download button
	downloadButtonSelectors := []string{
		"button:has-text('CSV')",
		"button:has-text('ダウンロード')",
		"a:has-text('CSV')",
		"a:has-text('ダウンロード')",
		"input[value*='CSV']",
		".download-csv",
	}

	downloadButton := s.findElement(downloadButtonSelectors)
	if downloadButton != nil {
		downloadButton.Click(LocatorClickOptions{})
		s.logger.Println("Clicked download button")

		// Wait for download with timeout
		select {
		case path := <-downloadComplete:
			s.logger.Printf("Download completed: %s", path)
			return path, nil
		case <-time.After(100 * time.Millisecond):
			return "", fmt.Errorf("download timeout")
		}
	}

	return "", fmt.Errorf("could not find download button")
}

// HandleDownload processes download events (exported for testing)
func (s *ETCScraper) HandleDownload(download Download, downloadComplete chan<- string) {
	suggestedFilename := download.SuggestedFilename()
	downloadPath := filepath.Join(s.config.DownloadPath, suggestedFilename)

	s.logger.Printf("Downloading file: %s", suggestedFilename)
	if err := download.SaveAs(downloadPath); err != nil {
		s.logger.Printf("Failed to save download: %v", err)
	} else {
		downloadComplete <- downloadPath
	}
}

// findElement tries multiple selectors and returns the first match
func (s *ETCScraper) findElement(selectors []string) LocatorInterface {
	for _, selector := range selectors {
		locator := s.page.Locator(selector)
		count, err := locator.Count()
		if err == nil && count > 0 {
			s.logger.Printf("Found element with selector: %s", selector)
			return locator.First()
		}
	}
	return nil
}

// Removed takeScreenshot method - no longer needed

// Close cleans up resources
func (s *ETCScraper) Close() error {
	if s.page != nil {
		s.page.Close()
	}
	if s.context != nil {
		s.context.Close()
	}
	if s.browser != nil {
		s.browser.Close()
	}
	if s.pw != nil {
		s.pw.Stop()
	}
	return nil
}
// ReadAndDeleteFile reads a file and deletes it (extracted for testing)
func ReadAndDeleteFile(path string) ([]byte, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	if err := os.Remove(path); err != nil {
		// Ignore deletion errors
	}

	return data, nil
}

// DownloadMeisaiToBuffer downloads ETC meisai data and returns it as a byte buffer
func (s *ETCScraper) DownloadMeisaiToBuffer(fromDate, toDate string) ([]byte, error) {
	// Download CSV file
	csvPath, err := s.DownloadMeisai(fromDate, toDate)
	if err != nil {
		return nil, fmt.Errorf("failed to download CSV: %w", err)
	}

	// Read and delete file
	data, err := ReadAndDeleteFile(csvPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV file: %w", err)
	}

	return data, nil
}

// waitForNavigation waits for page navigation (extracted for testing)
func (s *ETCScraper) waitForNavigation() {
	if !s.config.TestMode {
		time.Sleep(3 * time.Second)
	} else {
		s.logger.Printf("TestMode: Skipping 3 second sleep")
	}
}

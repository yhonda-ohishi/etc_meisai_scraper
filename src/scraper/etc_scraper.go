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
	UserID        string
	Password      string
	DownloadPath  string
	SessionFolder string // Current session folder for this execution
	Headless      bool
	Timeout       float64
	RetryCount    int
	UserAgent     string
	SlowMo        float64
	TestMode      bool // Skip time.Sleep in tests
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

	// Log Headless mode setting
	if s.config.Headless {
		s.logger.Println("🔒 Launching browser in Headless mode (browser not visible)")
	} else {
		s.logger.Println("👁️  Launching browser in VISIBLE mode (browser will appear)")
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

	// Setup dialog handler to auto-accept all dialogs (for CSV download confirmation)
	s.logger.Println("Setting up global dialog handler...")
	s.page.On("dialog", func(dialog interface{}) {
		s.logger.Printf("🔔 Dialog detected! Type: %T", dialog)
		// playwright.Dialog has Accept() method - cast to playwright.Dialog
		if d, ok := dialog.(interface {
			Accept(promptText ...string) error
		}); ok {
			s.logger.Println("✅ Auto-accepting dialog...")
			if err := d.Accept(); err != nil {
				s.logger.Printf("❌ Failed to accept dialog: %v", err)
			} else {
				s.logger.Println("✅ Dialog accepted successfully!")
			}
		} else {
			s.logger.Printf("❌ Could not cast dialog to acceptable interface")
		}
	})

	s.logger.Printf("Scraper initialized with download path: %s", s.config.DownloadPath)
	return nil
}

// Login performs login to ETC meisai service
func (s *ETCScraper) Login() error {
	if s.page == nil {
		return fmt.Errorf("scraper not initialized")
	}

	s.logger.Println("Navigating to https://www.etc-meisai.jp/")

	// Navigate to top page
	_, err := s.page.Goto("https://www.etc-meisai.jp/", PageGotoOptions{
		WaitUntil: WaitUntilStateNetworkidle,
	})
	if err != nil {
		return fmt.Errorf("failed to navigate to top page: %w", err)
	}

	// Click login link
	s.logger.Println("Clicking login link...")
	loginLinkSelector := "a[href*='funccode=1013000000']"
	loginLink := s.page.Locator(loginLinkSelector).First()
	if err := loginLink.Click(LocatorClickOptions{}); err != nil {
		return fmt.Errorf("failed to click login link: %w", err)
	}

	// Wait for login page to load
	s.waitForNavigation()
	err = s.page.WaitForLoadState(PageWaitForLoadStateOptions{
		State: LoadStateNetworkidle,
	})
	if err != nil {
		return fmt.Errorf("failed to load login page: %w", err)
	}

	// Wait for login form with correct field names
	s.logger.Println("Waiting for login form...")
	userIDField := s.page.Locator("input[name='risLoginId']")
	passwordField := s.page.Locator("input[name='risPassword']")

	// Fill user ID
	s.logger.Println("Filling login credentials...")
	if err := userIDField.Fill(s.config.UserID); err != nil {
		return fmt.Errorf("failed to fill user ID: %w", err)
	}

	// Fill password
	if err := passwordField.Fill(s.config.Password); err != nil {
		return fmt.Errorf("failed to fill password: %w", err)
	}

	// Click login button
	s.logger.Println("Clicking login button...")
	loginButton := s.page.Locator("input[type='button'][value='ログイン']")
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

	// Use existing session folder or create a new one
	var sessionFolder string
	if s.config.SessionFolder != "" {
		// Use existing session folder (for multiple downloads in same session)
		sessionFolder = s.config.SessionFolder
		s.logger.Printf("Using existing session folder: %s", sessionFolder)
	} else {
		// Create new timestamped subfolder for this download session
		timestamp := time.Now().Format("20060102_150405")
		sessionFolder = filepath.Join(s.config.DownloadPath, timestamp)
		if err := os.MkdirAll(sessionFolder, 0755); err != nil {
			return "", fmt.Errorf("failed to create session folder: %w", err)
		}
		s.logger.Printf("Created new session folder: %s", sessionFolder)
		s.config.SessionFolder = sessionFolder
	}

	// Update download path to use session folder
	originalDownloadPath := s.config.DownloadPath
	s.config.DownloadPath = sessionFolder
	defer func() {
		s.config.DownloadPath = originalDownloadPath
	}()

	// Navigate to search page (検索条件の指定)
	s.logger.Println("Navigating to search page...")
	searchPageLink := s.page.Locator("a:has-text('検索条件の指定')").First()
	if err := searchPageLink.Click(LocatorClickOptions{}); err != nil {
		// If link not found, we might already be on search page
		s.logger.Println("Search link not found, assuming already on search page")
	} else {
		s.waitForNavigation()
		s.page.WaitForLoadState(PageWaitForLoadStateOptions{
			State: LoadStateNetworkidle,
		})
	}

	// Click search button to execute search with current date range
	s.logger.Println("Clicking search button...")
	searchButton := s.page.Locator("input[name='focusTarget']").First()
	if err := searchButton.Click(LocatorClickOptions{}); err != nil {
		return "", fmt.Errorf("failed to click search button: %w", err)
	}

	// Wait for results page to load
	s.waitForNavigation()
	err := s.page.WaitForLoadState(PageWaitForLoadStateOptions{
		State: LoadStateNetworkidle,
	})
	if err != nil {
		return "", fmt.Errorf("failed to wait for search results: %w", err)
	}

	// Check if there are any results
	s.logger.Println("Checking for search results...")
	resultCount, _ := s.page.Locator("input[name='hakkoMeisai']").Count()
	s.logger.Printf("Found %d result items", resultCount)

	if resultCount == 0 {
		s.logger.Println("⚠️ No search results found. CSV link may not be available.")
	}

	// Setup download handler
	downloadComplete := make(chan string, 1)
	s.logger.Println("Setting up download handler...")
	s.page.On("download", func(download Download) {
		s.logger.Println("📥 Download event triggered!")
		s.HandleDownload(download, downloadComplete)
	})

	// Click CSV download link
	s.logger.Println("Clicking CSV download link...")

	// Try multiple selectors for CSV link
	// Note: onclick funccode varies by account type (1032500000 or other)
	csvSelectors := []string{
		"a:has-text('明細ＣＳＶ')",    // Text match (most reliable, ignores spacing)
		"a[onclick*='goOutput'][onclick*='hakkoMeisai']", // goOutput function call
		"a[onclick*='1032500000']",  // 明細CSV funccode (pattern 1)
	}

	var csvLink LocatorInterface
	var csvLinkCount int

	for _, selector := range csvSelectors {
		count, _ := s.page.Locator(selector).Count()
		s.logger.Printf("Trying selector '%s': found %d links", selector, count)
		if count > 0 {
			csvLink = s.page.Locator(selector).First()
			csvLinkCount = count
			break
		}
	}

	if csvLinkCount == 0 {
		return "", fmt.Errorf("CSV download link not found with any selector - possibly no search results or different page structure")
	}

	s.logger.Println("CSV link located, attempting click...")
	if err := csvLink.Click(LocatorClickOptions{}); err != nil {
		return "", fmt.Errorf("failed to click CSV link: %w", err)
	}
	s.logger.Println("CSV link clicked successfully!")

	s.logger.Println("Waiting for CSV download to complete...")

	// Wait for download with timeout
	select {
	case path := <-downloadComplete:
		s.logger.Printf("Download completed: %s", path)
		return path, nil
	case <-time.After(60 * time.Second):
		return "", fmt.Errorf("download timeout after 60 seconds")
	}
}

// HandleDownload processes download events (exported for testing)
func (s *ETCScraper) HandleDownload(download Download, downloadComplete chan<- string) {
	suggestedFilename := download.SuggestedFilename()

	// Add account name prefix to filename
	filenameWithAccount := s.config.UserID + "_" + suggestedFilename
	downloadPath := filepath.Join(s.config.DownloadPath, filenameWithAccount)

	s.logger.Printf("Downloading file: %s", suggestedFilename)
	s.logger.Printf("Saving to: %s", downloadPath)

	// Run SaveAs in a goroutine with timeout
	go func() {
		done := make(chan error, 1)
		go func() {
			done <- download.SaveAs(downloadPath)
		}()

		// Wait for SaveAs to complete or timeout after 30 seconds
		select {
		case err := <-done:
			if err != nil {
				s.logger.Printf("❌ Failed to save download: %v", err)
			} else {
				s.logger.Printf("✅ File saved successfully: %s", downloadPath)
				downloadComplete <- downloadPath
			}
		case <-time.After(30 * time.Second):
			// SaveAs is hanging, but file is probably saved
			// Check if file exists
			if _, err := filepath.Glob(downloadPath); err == nil {
				s.logger.Printf("⚠️ SaveAs timeout, but file appears to exist: %s", downloadPath)
				downloadComplete <- downloadPath
			} else {
				s.logger.Printf("❌ SaveAs timeout and file not found")
			}
		}
	}()
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

package scraper

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/playwright-community/playwright-go"
)

// ETCScraperV2 is an improved version with better error handling
type ETCScraperV2 struct {
	pw           *playwright.Playwright
	browser      playwright.Browser
	context      playwright.BrowserContext
	page         playwright.Page
	config       *ScraperConfig
	downloadChan chan string
}

// NewETCScraperV2 creates a new improved scraper
func NewETCScraperV2(config *ScraperConfig) (*ETCScraperV2, error) {
	if config.DownloadPath == "" {
		config.DownloadPath = "./downloads"
	}
	if config.Timeout == 0 {
		config.Timeout = 30000
	}

	if err := os.MkdirAll(config.DownloadPath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create download directory: %w", err)
	}

	// Create screenshots directory
	screenshotDir := filepath.Join(config.DownloadPath, "screenshots")
	if err := os.MkdirAll(screenshotDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create screenshot directory: %w", err)
	}

	return &ETCScraperV2{
		config:       config,
		downloadChan: make(chan string, 1),
	}, nil
}

// Initialize sets up Playwright and browser with retry logic
func (s *ETCScraperV2) Initialize() error {
	var err error
	maxRetries := 3

	for i := 0; i < maxRetries; i++ {
		if i > 0 {
			log.Printf("Retry %d/%d for initialization", i, maxRetries-1)
			time.Sleep(time.Second * 2)
		}

		// Install playwright browsers if needed
		err = playwright.Install()
		if err != nil {
			log.Printf("Could not install playwright: %v", err)
			continue
		}

		// Start Playwright
		s.pw, err = playwright.Run()
		if err != nil {
			log.Printf("Could not start playwright: %v", err)
			continue
		}

		// Launch browser with options
		launchOptions := playwright.BrowserTypeLaunchOptions{
			Headless: playwright.Bool(s.config.Headless),
		}

		// Add additional options for non-headless mode
		if !s.config.Headless {
			launchOptions.SlowMo = playwright.Float(50) // Slow down for debugging
		}

		s.browser, err = s.pw.Chromium.Launch(launchOptions)
		if err != nil {
			log.Printf("Could not launch browser: %v", err)
			s.pw.Stop()
			continue
		}

		// Create browser context
		s.context, err = s.browser.NewContext(playwright.BrowserNewContextOptions{
			AcceptDownloads: playwright.Bool(true),
			Viewport: &playwright.Size{
				Width:  1920,
				Height: 1080,
			},
			UserAgent: playwright.String("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"),
		})
		if err != nil {
			log.Printf("Could not create browser context: %v", err)
			s.browser.Close()
			s.pw.Stop()
			continue
		}

		// Set default timeout
		s.context.SetDefaultTimeout(s.config.Timeout)

		// Create page
		s.page, err = s.context.NewPage()
		if err != nil {
			log.Printf("Could not create page: %v", err)
			s.context.Close()
			s.browser.Close()
			s.pw.Stop()
			continue
		}

		// Setup download handler
		s.page.On("download", s.handleDownload)

		log.Printf("ETCScraperV2 initialized successfully")
		return nil
	}

	return fmt.Errorf("failed to initialize after %d retries: %w", maxRetries, err)
}

// handleDownload handles download events
func (s *ETCScraperV2) handleDownload(download playwright.Download) {
	suggestedFilename := download.SuggestedFilename()
	downloadPath := filepath.Join(s.config.DownloadPath, suggestedFilename)

	log.Printf("Downloading file: %s to %s", suggestedFilename, downloadPath)

	if err := download.SaveAs(downloadPath); err != nil {
		log.Printf("Failed to save download: %v", err)
		return
	}

	// Send download path to channel
	select {
	case s.downloadChan <- downloadPath:
	default:
	}
}

// LoginWithRetry performs login with retry logic
func (s *ETCScraperV2) LoginWithRetry() error {
	maxRetries := 3

	for i := 0; i < maxRetries; i++ {
		if i > 0 {
			log.Printf("Login retry %d/%d", i, maxRetries-1)
			time.Sleep(time.Second * 3)

			// Reload page before retry
			s.page.Reload()
		}

		err := s.performLogin()
		if err == nil {
			return nil
		}

		log.Printf("Login attempt %d failed: %v", i+1, err)
	}

	return fmt.Errorf("login failed after %d attempts", maxRetries)
}

// performLogin performs the actual login steps
func (s *ETCScraperV2) performLogin() error {
	log.Println("Navigating to ETC meisai site...")

	// Navigate to the site
	response, err := s.page.Goto("https://www.etc-meisai.jp/", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateNetworkidle,
		Timeout:   playwright.Float(30000),
	})
	if err != nil {
		return fmt.Errorf("failed to navigate: %w", err)
	}

	if response.Status() >= 400 {
		return fmt.Errorf("server returned status %d", response.Status())
	}

	// Take screenshot for debugging
	s.takeDebugScreenshot("01_login_page")

	// Wait a moment for any JavaScript to execute
	time.Sleep(time.Second * 2)

	// Try multiple strategies to find and fill login fields
	log.Println("Finding and filling login form...")

	// Strategy 1: Try common field names
	userFilled := s.tryFillField([]string{
		"input[name='userId']",
		"input[name='user_id']",
		"input[name='loginId']",
		"input[name='username']",
		"#userId",
		"#user_id",
		"#loginId",
		"input[placeholder*='ユーザ']",
		"input[placeholder*='ID']",
	}, s.config.UserID, "user ID")

	if !userFilled {
		// Strategy 2: Find first text input
		firstTextInput := s.page.Locator("input[type='text']").First()
		if count, _ := firstTextInput.Count(); count > 0 {
			log.Println("Trying first text input for user ID")
			firstTextInput.Fill(s.config.UserID)
			userFilled = true
		}
	}

	if !userFilled {
		return fmt.Errorf("could not find user ID field")
	}

	// Fill password
	passFilled := s.tryFillField([]string{
		"input[name='password']",
		"input[name='passwd']",
		"input[name='pass']",
		"#password",
		"#passwd",
		"input[type='password']",
		"input[placeholder*='パスワード']",
	}, s.config.Password, "password")

	if !passFilled {
		return fmt.Errorf("could not find password field")
	}

	// Take screenshot after filling
	s.takeDebugScreenshot("02_filled_form")

	// Click login button
	log.Println("Clicking login button...")

	loginClicked := s.tryClickButton([]string{
		"button[type='submit']",
		"input[type='submit']",
		"button:has-text('ログイン')",
		"input[value*='ログイン']",
		"button:has-text('Login')",
		"input[value*='Login']",
		"button:has-text('サインイン')",
		"*:has-text('ログイン'):not(a)",
	})

	if !loginClicked {
		// Try pressing Enter in password field
		log.Println("Trying Enter key in password field")
		s.page.Locator("input[type='password']").Press("Enter")
	}

	// Wait for navigation or content change
	log.Println("Waiting for login to complete...")
	time.Sleep(time.Second * 5)

	// Take screenshot after login attempt
	s.takeDebugScreenshot("03_after_login")

	// Check if login was successful
	currentURL := s.page.URL()
	log.Printf("Current URL: %s", currentURL)

	// Look for success indicators
	successIndicators := []string{
		"ログアウト",
		"logout",
		"明細",
		"利用明細",
		"マイページ",
		"メニュー",
	}

	for _, indicator := range successIndicators {
		if count, _ := s.page.Locator(fmt.Sprintf("*:has-text('%s')", indicator)).Count(); count > 0 {
			log.Printf("Login successful - found: %s", indicator)
			return nil
		}
	}

	// Check for error messages
	errorSelectors := []string{
		".error",
		".alert",
		"*:has-text('エラー')",
		"*:has-text('失敗')",
		"*:has-text('誤り')",
	}

	for _, selector := range errorSelectors {
		if elem, _ := s.page.Locator(selector).First().TextContent(); elem != "" {
			return fmt.Errorf("login error: %s", elem)
		}
	}

	// If URL hasn't changed, login might have failed
	if currentURL == "https://www.etc-meisai.jp/" || strings.Contains(currentURL, "login") {
		return fmt.Errorf("login appears to have failed - still on login page")
	}

	log.Println("Login status uncertain, proceeding...")
	return nil
}

// SearchAndDownloadCSV searches and downloads CSV with date range
func (s *ETCScraperV2) SearchAndDownloadCSV(fromDate, toDate time.Time, cardNo string) (string, error) {
	log.Printf("Searching for data from %s to %s", fromDate.Format("2006-01-02"), toDate.Format("2006-01-02"))

	// Navigate to search/download page
	if err := s.navigateToSearchPage(); err != nil {
		return "", fmt.Errorf("failed to navigate to search page: %w", err)
	}

	// Fill search criteria
	if err := s.fillSearchCriteria(fromDate, toDate, cardNo); err != nil {
		return "", fmt.Errorf("failed to fill search criteria: %w", err)
	}

	// Execute search
	if err := s.executeSearch(); err != nil {
		return "", fmt.Errorf("failed to execute search: %w", err)
	}

	// Download CSV
	downloadPath, err := s.downloadCSV()
	if err != nil {
		return "", fmt.Errorf("failed to download CSV: %w", err)
	}

	return downloadPath, nil
}

// navigateToSearchPage navigates to the search/download page
func (s *ETCScraperV2) navigateToSearchPage() error {
	log.Println("Looking for search/download page...")

	// Try clicking on menu items
	menuSelectors := []string{
		"a:has-text('利用明細')",
		"a:has-text('明細照会')",
		"a:has-text('明細検索')",
		"a:has-text('ダウンロード')",
		"a:has-text('CSV')",
		"a[href*='search']",
		"a[href*='meisai']",
		"a[href*='download']",
		"a[href*='inquiry']",
	}

	for _, selector := range menuSelectors {
		if count, _ := s.page.Locator(selector).Count(); count > 0 {
			log.Printf("Clicking menu item: %s", selector)
			s.page.Click(selector)
			time.Sleep(time.Second * 3)

			s.takeDebugScreenshot("04_search_page")
			return nil
		}
	}

	// If no menu found, we might already be on the right page
	log.Println("No menu item found, checking if already on search page")
	return nil
}

// fillSearchCriteria fills the search form
func (s *ETCScraperV2) fillSearchCriteria(fromDate, toDate time.Time, cardNo string) error {
	log.Println("Filling search criteria...")

	// Date formats to try
	dateFormats := []string{
		"2006/01/02",
		"2006-01-02",
		"2006年01月02日",
	}

	for _, format := range dateFormats {
		fromStr := fromDate.Format(format)
		toStr := toDate.Format(format)

		// Fill from date
		s.tryFillField([]string{
			"input[name*='from']",
			"input[name*='start']",
			"input[name*='開始']",
			"#fromDate",
			"#startDate",
			"input[placeholder*='から']",
		}, fromStr, "from date")

		// Fill to date
		s.tryFillField([]string{
			"input[name*='to']",
			"input[name*='end']",
			"input[name*='終了']",
			"#toDate",
			"#endDate",
			"input[placeholder*='まで']",
		}, toStr, "to date")
	}

	// Fill card number if provided
	if cardNo != "" {
		s.tryFillField([]string{
			"input[name*='card']",
			"#cardNo",
			"input[placeholder*='カード']",
		}, cardNo, "card number")
	}

	s.takeDebugScreenshot("05_search_filled")
	return nil
}

// executeSearch clicks the search button
func (s *ETCScraperV2) executeSearch() error {
	log.Println("Executing search...")

	searchClicked := s.tryClickButton([]string{
		"button:has-text('検索')",
		"input[value*='検索']",
		"button:has-text('照会')",
		"input[value*='照会']",
		"button:has-text('表示')",
		"button[type='submit']",
	})

	if !searchClicked {
		return fmt.Errorf("could not find search button")
	}

	// Wait for results
	time.Sleep(time.Second * 5)
	s.takeDebugScreenshot("06_search_results")

	return nil
}

// downloadCSV downloads the CSV file
func (s *ETCScraperV2) downloadCSV() (string, error) {
	log.Println("Looking for CSV download option...")

	// Click download button
	downloadClicked := s.tryClickButton([]string{
		"a:has-text('CSV')",
		"button:has-text('CSV')",
		"a:has-text('ダウンロード')",
		"button:has-text('ダウンロード')",
		"input[value*='CSV']",
		"a[href*='.csv']",
		"*[download*='.csv']",
	})

	if !downloadClicked {
		return "", fmt.Errorf("could not find CSV download button")
	}

	// Wait for download
	select {
	case downloadPath := <-s.downloadChan:
		log.Printf("CSV downloaded to: %s", downloadPath)
		return downloadPath, nil
	case <-time.After(30 * time.Second):
		return "", fmt.Errorf("download timeout")
	}
}

// Helper methods

func (s *ETCScraperV2) tryFillField(selectors []string, value string, fieldName string) bool {
	for _, selector := range selectors {
		if count, _ := s.page.Locator(selector).Count(); count > 0 {
			if err := s.page.Fill(selector, value); err == nil {
				log.Printf("Filled %s with selector: %s", fieldName, selector)
				return true
			}
		}
	}
	log.Printf("Could not fill %s", fieldName)
	return false
}

func (s *ETCScraperV2) tryClickButton(selectors []string) bool {
	for _, selector := range selectors {
		if count, _ := s.page.Locator(selector).Count(); count > 0 {
			if err := s.page.Click(selector); err == nil {
				log.Printf("Clicked button with selector: %s", selector)
				return true
			}
		}
	}
	return false
}

func (s *ETCScraperV2) takeDebugScreenshot(name string) {
	if s.config.Headless {
		return // Skip screenshots in headless mode unless debugging
	}

	screenshotPath := filepath.Join(s.config.DownloadPath, "screenshots", fmt.Sprintf("%s_%s.png", name, time.Now().Format("20060102_150405")))
	s.page.Screenshot(playwright.PageScreenshotOptions{
		Path:     playwright.String(screenshotPath),
		FullPage: playwright.Bool(true),
	})
	log.Printf("Screenshot saved: %s", screenshotPath)
}

// Close cleans up resources
func (s *ETCScraperV2) Close() {
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
	close(s.downloadChan)
}
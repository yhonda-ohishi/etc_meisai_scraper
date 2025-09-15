package scraper

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/playwright-community/playwright-go"
)

// ETCScraper handles web scraping for ETC meisai service
type ETCScraper struct {
	pw      *playwright.Playwright
	browser playwright.Browser
	context playwright.BrowserContext
	page    playwright.Page
	config  *ScraperConfig
}

// ScraperConfig holds configuration for the scraper
type ScraperConfig struct {
	UserID       string
	Password     string
	DownloadPath string
	Headless     bool
	Timeout      float64
	// Extended configuration options
	RetryCount   int     // Number of retry attempts for failed operations
	UserAgent    string  // Custom user agent string
	SlowMo       float64 // Slow down operations by specified milliseconds
	Viewport     *ViewportSize
}

// ViewportSize defines browser viewport dimensions
type ViewportSize struct {
	Width  int
	Height int
}

// NewETCScraper creates a new ETC scraper instance
func NewETCScraper(config *ScraperConfig) (*ETCScraper, error) {
	// Set default values
	if config.DownloadPath == "" {
		config.DownloadPath = "./downloads"
	}
	if config.Timeout == 0 {
		config.Timeout = 30000 // 30 seconds default
	}
	if config.RetryCount == 0 {
		config.RetryCount = 3 // Default retry count
	}
	if config.UserAgent == "" {
		config.UserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"
	}
	if config.Viewport == nil {
		config.Viewport = &ViewportSize{
			Width:  1920,
			Height: 1080,
		}
	}

	// Create download directory if it doesn't exist
	if err := os.MkdirAll(config.DownloadPath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create download directory: %w", err)
	}

	return &ETCScraper{
		config: config,
	}, nil
}

// Initialize sets up Playwright and browser
func (s *ETCScraper) Initialize() error {
	// Install playwright browsers if needed
	err := playwright.Install()
	if err != nil {
		return fmt.Errorf("could not install playwright: %w", err)
	}

	// Start Playwright
	s.pw, err = playwright.Run()
	if err != nil {
		return fmt.Errorf("could not start playwright: %w", err)
	}

	// Launch browser
	s.browser, err = s.pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
		Headless: playwright.Bool(s.config.Headless),
	})
	if err != nil {
		return fmt.Errorf("could not launch browser: %w", err)
	}

	// Create browser context with download settings
	absPath, _ := filepath.Abs(s.config.DownloadPath)
	s.context, err = s.browser.NewContext(playwright.BrowserNewContextOptions{
		AcceptDownloads: playwright.Bool(true),
		Viewport: &playwright.Size{
			Width:  1920,
			Height: 1080,
		},
	})
	if err != nil {
		return fmt.Errorf("could not create browser context: %w", err)
	}

	// Set download path
	s.context.SetDefaultTimeout(s.config.Timeout)

	// Create page
	s.page, err = s.context.NewPage()
	if err != nil {
		return fmt.Errorf("could not create page: %w", err)
	}

	log.Printf("ETCScraper initialized with download path: %s", absPath)
	return nil
}

// Login performs login to ETC meisai service
func (s *ETCScraper) Login() error {
	if s.page == nil {
		return fmt.Errorf("scraper not initialized")
	}

	log.Println("Navigating to ETC meisai login page...")

	// Navigate to login page
	_, err := s.page.Goto("https://www.etc-meisai.jp/", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateNetworkidle,
	})
	if err != nil {
		return fmt.Errorf("failed to navigate to login page: %w", err)
	}

	// Wait for login form
	log.Println("Waiting for login form...")
	_, err = s.page.WaitForSelector("input[name='userId']", playwright.PageWaitForSelectorOptions{
		Timeout: playwright.Float(10000),
	})
	if err != nil {
		// Try alternative selector
		_, err = s.page.WaitForSelector("#userId", playwright.PageWaitForSelectorOptions{
			Timeout: playwright.Float(10000),
		})
		if err != nil {
			return fmt.Errorf("login form not found: %w", err)
		}
	}

	// Fill user ID
	log.Println("Filling login credentials...")
	err = s.page.Fill("input[name='userId']", s.config.UserID)
	if err != nil {
		// Try alternative selector
		err = s.page.Fill("#userId", s.config.UserID)
		if err != nil {
			return fmt.Errorf("failed to fill user ID: %w", err)
		}
	}

	// Fill password
	err = s.page.Fill("input[name='password']", s.config.Password)
	if err != nil {
		// Try alternative selector
		err = s.page.Fill("#password", s.config.Password)
		if err != nil {
			return fmt.Errorf("failed to fill password: %w", err)
		}
	}

	// Click login button
	log.Println("Clicking login button...")
	err = s.page.Click("button[type='submit']")
	if err != nil {
		// Try alternative selectors
		err = s.page.Click("input[type='submit']")
		if err != nil {
			err = s.page.Click(".login-button")
			if err != nil {
				return fmt.Errorf("failed to click login button: %w", err)
			}
		}
	}

	// Wait for navigation after login
	err = s.page.WaitForLoadState(playwright.PageWaitForLoadStateOptions{
		State: playwright.LoadStateNetworkidle,
	})
	if err != nil {
		return fmt.Errorf("failed to wait for login completion: %w", err)
	}

	// Check if login was successful by looking for logout button or user info
	logoutExists, err := s.page.Locator("a:has-text('ログアウト')").Count()
	if err == nil && logoutExists > 0 {
		log.Println("Login successful!")
		return nil
	}

	// Check for error messages
	errorMsg, err := s.page.Locator(".error-message").TextContent()
	if err == nil && errorMsg != "" {
		return fmt.Errorf("login failed: %s", errorMsg)
	}

	log.Println("Login completed, verifying...")
	return nil
}

// SearchAndDownload searches for data and downloads CSV
func (s *ETCScraper) SearchAndDownload(fromDate, toDate time.Time, cardNo string) (string, error) {
	if s.page == nil {
		return "", fmt.Errorf("scraper not initialized")
	}

	log.Printf("Searching for data from %s to %s", fromDate.Format("2006-01-02"), toDate.Format("2006-01-02"))

	// Navigate to search page if needed
	searchUrl := "https://www.etc-meisai.jp/search" // Update with actual URL
	currentUrl := s.page.URL()
	if currentUrl != searchUrl {
		_, err := s.page.Goto(searchUrl, playwright.PageGotoOptions{
			WaitUntil: playwright.WaitUntilStateNetworkidle,
		})
		if err != nil {
			log.Printf("Note: Could not navigate to search page: %v", err)
		}
	}

	// Fill search form
	// Date format might need adjustment based on actual site requirements
	fromDateStr := fromDate.Format("2006/01/02")
	toDateStr := toDate.Format("2006/01/02")

	// Try to fill from date
	if err := s.fillDateField("from", fromDateStr); err != nil {
		log.Printf("Warning: Could not fill from date: %v", err)
	}

	// Try to fill to date
	if err := s.fillDateField("to", toDateStr); err != nil {
		log.Printf("Warning: Could not fill to date: %v", err)
	}

	// Fill card number if provided
	if cardNo != "" {
		selectors := []string{"input[name='cardNo']", "#cardNo", "input[placeholder*='カード']"}
		for _, selector := range selectors {
			if err := s.page.Fill(selector, cardNo); err == nil {
				log.Printf("Filled card number with selector: %s", selector)
				break
			}
		}
	}

	// Click search button
	log.Println("Clicking search button...")
	searchSelectors := []string{
		"button:has-text('検索')",
		"input[value='検索']",
		"button[type='submit']",
		".search-button",
	}

	clicked := false
	for _, selector := range searchSelectors {
		if err := s.page.Click(selector); err == nil {
			clicked = true
			log.Printf("Clicked search with selector: %s", selector)
			break
		}
	}

	if !clicked {
		log.Println("Warning: Could not find search button")
	}

	// Wait for results
	time.Sleep(3 * time.Second)

	// Setup download handler
	downloadPath := ""
	s.page.On("download", func(download playwright.Download) {
		// Save download
		suggestedFilename := download.SuggestedFilename()
		downloadPath = filepath.Join(s.config.DownloadPath, suggestedFilename)

		log.Printf("Downloading file: %s", suggestedFilename)
		if err := download.SaveAs(downloadPath); err != nil {
			log.Printf("Failed to save download: %v", err)
		}
	})

	// Click download CSV button
	log.Println("Looking for CSV download button...")
	downloadSelectors := []string{
		"a:has-text('CSV')",
		"button:has-text('CSV')",
		"a[href*='.csv']",
		"button:has-text('ダウンロード')",
		".download-csv",
	}

	for _, selector := range downloadSelectors {
		count, _ := s.page.Locator(selector).Count()
		if count > 0 {
			log.Printf("Found download button with selector: %s", selector)
			if err := s.page.Click(selector); err == nil {
				// Wait for download to complete
				time.Sleep(5 * time.Second)
				if downloadPath != "" {
					return downloadPath, nil
				}
			}
		}
	}

	return "", fmt.Errorf("could not find or click CSV download button")
}

// fillDateField attempts to fill a date field with various strategies
func (s *ETCScraper) fillDateField(fieldType string, dateStr string) error {
	// Common date field selectors
	selectors := []string{
		fmt.Sprintf("input[name='%sDate']", fieldType),
		fmt.Sprintf("#%sDate", fieldType),
		fmt.Sprintf("input[name='%s']", fieldType),
		fmt.Sprintf("#%s", fieldType),
	}

	if fieldType == "from" {
		selectors = append(selectors, "input[name='startDate']", "#startDate")
	} else if fieldType == "to" {
		selectors = append(selectors, "input[name='endDate']", "#endDate")
	}

	for _, selector := range selectors {
		if err := s.page.Fill(selector, dateStr); err == nil {
			return nil
		}
	}

	return fmt.Errorf("could not fill %s date field", fieldType)
}

// Close cleans up resources
func (s *ETCScraper) Close() {
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
}

// TakeScreenshot captures the current page
func (s *ETCScraper) TakeScreenshot(filename string) error {
	if s.page == nil {
		return fmt.Errorf("page not initialized")
	}

	screenshotPath := filepath.Join(s.config.DownloadPath, filename)
	_, err := s.page.Screenshot(playwright.PageScreenshotOptions{
		Path: playwright.String(screenshotPath),
		FullPage: playwright.Bool(true),
	})

	if err != nil {
		return fmt.Errorf("failed to take screenshot: %w", err)
	}

	log.Printf("Screenshot saved to: %s", screenshotPath)
	return nil
}
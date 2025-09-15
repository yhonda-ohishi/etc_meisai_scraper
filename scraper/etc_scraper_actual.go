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

// ActualETCScraper is the scraper for the actual ETC site
type ActualETCScraper struct {
	pw           *playwright.Playwright
	browser      playwright.Browser
	context      playwright.BrowserContext
	page         playwright.Page
	config       *ScraperConfig
	downloadChan chan string
}

// NewActualETCScraper creates a scraper for the actual ETC site
func NewActualETCScraper(config *ScraperConfig) (*ActualETCScraper, error) {
	if config.DownloadPath == "" {
		config.DownloadPath = "./downloads"
	}
	if config.Timeout == 0 {
		config.Timeout = 30000
	}

	if err := os.MkdirAll(config.DownloadPath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create download directory: %w", err)
	}

	return &ActualETCScraper{
		config:       config,
		downloadChan: make(chan string, 1),
	}, nil
}

// Initialize sets up the browser
func (s *ActualETCScraper) Initialize() error {
	var err error

	// Install playwright
	err = playwright.Install()
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
		SlowMo:   playwright.Float(50),
	})
	if err != nil {
		return fmt.Errorf("could not launch browser: %w", err)
	}

	// Create browser context with JavaScript enabled
	s.context, err = s.browser.NewContext(playwright.BrowserNewContextOptions{
		AcceptDownloads: playwright.Bool(true),
		JavaScriptEnabled: playwright.Bool(true), // JavaScriptを明示的に有効化
		Viewport: &playwright.Size{
			Width:  1920,
			Height: 1080,
		},
		UserAgent: playwright.String("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"),
		// ダウンロード動作を改善するための追加設定
		BypassCSP: playwright.Bool(true), // CSP制限をバイパス
		IgnoreHttpsErrors: playwright.Bool(true), // HTTPS証明書エラーを無視
	})
	if err != nil {
		return fmt.Errorf("could not create browser context: %w", err)
	}

	s.context.SetDefaultTimeout(s.config.Timeout)

	// Create page
	s.page, err = s.context.NewPage()
	if err != nil {
		return fmt.Errorf("could not create page: %w", err)
	}

	log.Println("ActualETCScraper initialized successfully")
	return nil
}

// handleDownload handles download events with progress tracking
func (s *ActualETCScraper) handleDownload(download playwright.Download) {
	suggestedFilename := download.SuggestedFilename()

	// Ensure download path uses forward slashes for cross-platform compatibility
	downloadDir := filepath.ToSlash(s.config.DownloadPath)
	downloadPath := filepath.Join(downloadDir, suggestedFilename)

	// Convert to absolute path
	absPath, err := filepath.Abs(downloadPath)
	if err != nil {
		log.Printf("Failed to get absolute path: %v", err)
		absPath = downloadPath
	}

	log.Printf("Download started: %s", suggestedFilename)
	log.Printf("Saving to: %s", absPath)

	// Save the download
	if err := download.SaveAs(absPath); err != nil {
		log.Printf("Failed to save download %s: %v", suggestedFilename, err)
		// Send error signal through channel
		select {
		case s.downloadChan <- "":
		default:
		}
		return
	}

	log.Printf("Download completed: %s -> %s", suggestedFilename, absPath)

	// Signal download completion
	select {
	case s.downloadChan <- absPath:
	default:
		log.Printf("Warning: Download channel full, could not signal completion")
	}
}

// Login performs login to the actual ETC site
func (s *ActualETCScraper) Login() error {
	log.Println("Navigating to ETC login page...")

	// The actual login URL based on the test results
	loginURL := "https://www2.etc-meisai.jp/etc/R?funccode=1013000000&nextfunc=1013000000"

	response, err := s.page.Goto(loginURL, playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateNetworkidle,
		Timeout:   playwright.Float(30000),
	})
	if err != nil {
		return fmt.Errorf("failed to navigate to login page: %w", err)
	}

	if response.Status() >= 400 {
		return fmt.Errorf("server returned status %d", response.Status())
	}

	// Wait for login form to load
	time.Sleep(time.Second * 2)

	// Take screenshot for debugging
	s.takeScreenshot("01_login_page")

	// Try to fill user ID field
	// Common patterns for the actual site
	userFilled := false
	userSelectors := []string{
		"input[name='usrid']",
		"input[name='userId']",
		"input[name='user_id']",
		"input[name='loginId']",
		"#usrid",
		"#userId",
		"input[type='text']:not([type='hidden'])",
	}

	for _, selector := range userSelectors {
		if count, _ := s.page.Locator(selector).Count(); count > 0 {
			log.Printf("Found user field with selector: %s", selector)
			if err := s.page.Fill(selector, s.config.UserID); err == nil {
				userFilled = true
				break
			}
		}
	}

	if !userFilled {
		return fmt.Errorf("could not find user ID field")
	}

	// Fill password field
	passFilled := false
	passwordSelectors := []string{
		"input[name='password']",
		"input[name='passwd']",
		"input[name='pass']",
		"#password",
		"input[type='password']:not([type='hidden'])",
	}

	for _, selector := range passwordSelectors {
		if count, _ := s.page.Locator(selector).Count(); count > 0 {
			log.Printf("Found password field with selector: %s", selector)
			if err := s.page.Fill(selector, s.config.Password); err == nil {
				passFilled = true
				break
			}
		}
	}

	if !passFilled {
		return fmt.Errorf("could not find password field")
	}

	// Take screenshot after filling
	s.takeScreenshot("02_filled_form")

	// Click login button
	loginClicked := false
	loginSelectors := []string{
		"input[type='submit'][value*='ログイン']",
		"button[type='submit']",
		"input[type='submit']",
		"input[type='button'][value*='ログイン']",
		"button:has-text('ログイン')",
		"a:has-text('ログイン'):not([href*='1013000000'])", // Not the login link itself
	}

	for _, selector := range loginSelectors {
		if count, _ := s.page.Locator(selector).Count(); count > 0 {
			log.Printf("Clicking login button with selector: %s", selector)
			if err := s.page.Click(selector); err == nil {
				loginClicked = true
				break
			}
		}
	}

	if !loginClicked {
		log.Println("Could not find login button, trying Enter key")
		s.page.Locator("input[type='password']").Press("Enter")
	}

	// Wait for login to complete
	time.Sleep(time.Second * 5)

	// Take screenshot after login
	s.takeScreenshot("03_after_login")

	// Check if login was successful
	currentURL := s.page.URL()
	log.Printf("Current URL after login: %s", currentURL)

	// Look for success indicators
	successIndicators := []string{
		"ログアウト",
		"利用明細",
		"照会",
		"メニュー",
	}

	for _, indicator := range successIndicators {
		if count, _ := s.page.Locator(fmt.Sprintf("*:has-text('%s')", indicator)).Count(); count > 0 {
			log.Printf("Login successful - found: %s", indicator)
			return nil
		}
	}

	// Check for error
	if errorElem, _ := s.page.Locator(".error, .alert, *:has-text('エラー')").First().TextContent(); errorElem != "" {
		return fmt.Errorf("login error: %s", errorElem)
	}

	log.Println("Login status uncertain, proceeding...")
	return nil
}

// SearchAndDownloadCSV searches and downloads CSV
func (s *ActualETCScraper) SearchAndDownloadCSV(fromDate, toDate time.Time) (string, error) {
	log.Printf("Searching for data from %s to %s", fromDate.Format("2006-01-02"), toDate.Format("2006-01-02"))

	// The main page after login has date select dropdowns
	log.Println("Setting date range on main page after login...")
	s.takeScreenshot("04_main_page_after_login")

	// Wait for page to stabilize
	time.Sleep(2 * time.Second)

	// Detect page structure by checking for specific elements
	log.Println("Detecting page structure...")
	hasDateSelects := false
	hasSearchForm := false

	// First, take a screenshot to see what we're dealing with
	s.takeScreenshot("04a_page_structure_check")

	// Check if date select dropdowns exist - also check the initial page for ohishiexp1
	if count, _ := s.page.Locator("select[name='fromYYYY']").Count(); count > 0 {
		hasDateSelects = true
		log.Println("Found date select dropdowns (standard structure)")
	} else {
		// For ohishiexp1, check if we need to click 検索条件 or similar to show date selects
		log.Println("Checking if date selects are hidden or need activation...")

		// Try clicking 検索条件 or similar buttons that might reveal date selects
		conditionButtons := []string{
			"a:has-text('検索条件')",
			"input[value*='条件']",
			"button:has-text('条件')",
			"a:has-text('詳細検索')",
		}

		for _, selector := range conditionButtons {
			if count, _ := s.page.Locator(selector).Count(); count > 0 {
				log.Printf("Found condition button: %s, clicking to reveal date selects...", selector)
				s.page.Click(selector)
				time.Sleep(2 * time.Second)

				// Check again for date selects
				if count, _ := s.page.Locator("select[name='fromYYYY']").Count(); count > 0 {
					hasDateSelects = true
					log.Println("Date select dropdowns appeared after clicking condition button")
					break
				}
			}
		}
	}

	// Check for alternative date select patterns
	if !hasDateSelects {
		// Try different naming patterns for date selects
		alternativeSelectors := []string{
			"select[name*='Year']",
			"select[name*='year']",
			"select[name*='YYYY']",
			"select[name*='yyyy']",
			"select[id*='year']",
			"select[id*='Year']",
		}

		for _, selector := range alternativeSelectors {
			if count, _ := s.page.Locator(selector).Count(); count > 0 {
				hasDateSelects = true
				log.Printf("Found alternative date select: %s", selector)
				break
			}
		}
	}

	// Check if there's a different search form structure
	if count, _ := s.page.Locator("input[name='searchStartDate']").Count(); count > 0 {
		hasSearchForm = true
		log.Println("Found search form with date inputs (alternative structure)")
	}

	// Check for card selection dropdown (some accounts have this)
	if count, _ := s.page.Locator("select[name='cardNo']").Count(); count > 0 {
		log.Println("Found card selection dropdown")
		// TODO: Handle card selection if needed
	}

	// Log all select elements on the page for debugging
	if selects, _ := s.page.Locator("select").All(); len(selects) > 0 {
		log.Printf("Found %d select elements on initial page", len(selects))
		for i, sel := range selects {
			name, _ := sel.GetAttribute("name")
			id, _ := sel.GetAttribute("id")
			if name != "" || id != "" {
				log.Printf("  Select %d: name='%s', id='%s'", i+1, name, id)
			}
		}
	} else {
		log.Println("No select elements found on initial page after login")

		// Check what inputs are available
		if inputs, _ := s.page.Locator("input[type='text'], input[type='date']").All(); len(inputs) > 0 {
			log.Printf("Found %d text/date input elements", len(inputs))
			for i, inp := range inputs {
				if i < 5 {
					name, _ := inp.GetAttribute("name")
					placeholder, _ := inp.GetAttribute("placeholder")
					log.Printf("  Input %d: name='%s', placeholder='%s'", i+1, name, placeholder)
				}
			}
		}
	}

	// Extract year, month, day components
	fromYear := fmt.Sprintf("%d", fromDate.Year())
	fromMonth := fmt.Sprintf("%02d", fromDate.Month())
	fromDay := fmt.Sprintf("%02d", fromDate.Day())

	toYear := fmt.Sprintf("%d", toDate.Year())
	toMonth := fmt.Sprintf("%02d", toDate.Month())
	toDay := fmt.Sprintf("%02d", toDate.Day())

	log.Printf("Setting from date: %s/%s/%s", fromYear, fromMonth, fromDay)
	log.Printf("Setting to date: %s/%s/%s", toYear, toMonth, toDay)

	// Handle date setting based on detected page structure
	if hasDateSelects {
		// Standard structure with select dropdowns
		log.Println("Using standard date select method...")

		// Set the from date using select dropdowns
		if _, err := s.page.SelectOption("select[name='fromYYYY']", playwright.SelectOptionValues{Values: &[]string{fromYear}}); err != nil {
			log.Printf("Failed to set fromYYYY: %v", err)
		} else {
			log.Printf("Set fromYYYY to %s", fromYear)
		}

		if _, err := s.page.SelectOption("select[name='fromMM']", playwright.SelectOptionValues{Values: &[]string{fromMonth}}); err != nil {
			log.Printf("Failed to set fromMM: %v", err)
		} else {
			log.Printf("Set fromMM to %s (先月当初)", fromMonth)
		}

		if _, err := s.page.SelectOption("select[name='fromDD']", playwright.SelectOptionValues{Values: &[]string{fromDay}}); err != nil {
			log.Printf("Failed to set fromDD: %v", err)
		} else {
			log.Printf("Set fromDD to %s", fromDay)
		}

		// Set the to date using select dropdowns
		if _, err := s.page.SelectOption("select[name='toYYYY']", playwright.SelectOptionValues{Values: &[]string{toYear}}); err != nil {
			log.Printf("Failed to set toYYYY: %v", err)
		} else {
			log.Printf("Set toYYYY to %s", toYear)
		}

		if _, err := s.page.SelectOption("select[name='toMM']", playwright.SelectOptionValues{Values: &[]string{toMonth}}); err != nil {
			log.Printf("Failed to set toMM: %v", err)
		} else {
			log.Printf("Set toMM to %s", toMonth)
		}

		if _, err := s.page.SelectOption("select[name='toDD']", playwright.SelectOptionValues{Values: &[]string{toDay}}); err != nil {
			log.Printf("Failed to set toDD: %v", err)
		} else {
			log.Printf("Set toDD to %s", toDay)
		}

	} else if hasSearchForm {
		// Alternative structure with date input fields
		log.Println("Using alternative date input method...")

		// Format dates for input fields (often YYYY/MM/DD or YYYY-MM-DD)
		fromDateStr := fromDate.Format("2006/01/02")
		toDateStr := toDate.Format("2006/01/02")

		// Try different input field names
		dateInputSelectors := [][]string{
			{"input[name='searchStartDate']", "input[name='searchEndDate']"},
			{"input[name='fromDate']", "input[name='toDate']"},
			{"input[name='startDate']", "input[name='endDate']"},
			{"input#fromDate", "input#toDate"},
		}

		for _, selectors := range dateInputSelectors {
			if count, _ := s.page.Locator(selectors[0]).Count(); count > 0 {
				log.Printf("Found date inputs: %s, %s", selectors[0], selectors[1])
				s.page.Fill(selectors[0], fromDateStr)
				s.page.Fill(selectors[1], toDateStr)
				break
			}
		}
	} else {
		log.Println("Warning: Could not detect date input method on current page")

		// For some accounts, we might need to navigate to a search page first
		log.Println("Checking if we need to navigate to search/detail page first...")

		// Try to find links to search or detail pages
		searchPageLinks := []string{
			"a:has-text('利用明細')",
			"a:has-text('明細照会')",
			"a:has-text('照会')",
			"a:has-text('検索')",
			"td:has-text('利用明細')",
			"span:has-text('利用明細')",
		}

		for _, selector := range searchPageLinks {
			if count, _ := s.page.Locator(selector).Count(); count > 0 {
				log.Printf("Found search/detail link: %s, clicking...", selector)
				s.page.Click(selector)
				time.Sleep(3 * time.Second)

				// Now check again for date inputs after navigation
				s.takeScreenshot("04b_after_navigation")

				// Log ALL select elements after navigation to debug
				if selects, _ := s.page.Locator("select").All(); len(selects) > 0 {
					log.Printf("After clicking 利用明細, found %d select elements", len(selects))
					for i, sel := range selects {
						name, _ := sel.GetAttribute("name")
						id, _ := sel.GetAttribute("id")
						if name != "" || id != "" {
							log.Printf("  Select %d: name='%s', id='%s'", i+1, name, id)
						}
					}

					// Check specifically for date selects with any pattern
					dateSelectPatterns := []string{
						"select[name='fromYYYY']",
						"select[name='fromYear']",
						"select[name*='from']",
						"select[name*='Year']",
						"select[name*='YYYY']",
					}

					for _, pattern := range dateSelectPatterns {
						if count, _ := s.page.Locator(pattern).Count(); count > 0 {
							log.Printf("Found date select with pattern: %s", pattern)
						}
					}
				}

				// Try to set dates if we find the selects
				if count, _ := s.page.Locator("select[name='fromYYYY']").Count(); count > 0 {
					log.Println("Found date selects after navigation, setting dates...")
					// Set dates using the standard method
					s.page.SelectOption("select[name='fromYYYY']", playwright.SelectOptionValues{Values: &[]string{fromYear}})
					log.Printf("Set fromYYYY to %s", fromYear)
					s.page.SelectOption("select[name='fromMM']", playwright.SelectOptionValues{Values: &[]string{fromMonth}})
					log.Printf("Set fromMM to %s", fromMonth)
					s.page.SelectOption("select[name='fromDD']", playwright.SelectOptionValues{Values: &[]string{fromDay}})
					log.Printf("Set fromDD to %s", fromDay)
					s.page.SelectOption("select[name='toYYYY']", playwright.SelectOptionValues{Values: &[]string{toYear}})
					log.Printf("Set toYYYY to %s", toYear)
					s.page.SelectOption("select[name='toMM']", playwright.SelectOptionValues{Values: &[]string{toMonth}})
					log.Printf("Set toMM to %s", toMonth)
					s.page.SelectOption("select[name='toDD']", playwright.SelectOptionValues{Values: &[]string{toDay}})
					log.Printf("Set toDD to %s", toDay)

					// After setting dates, might need to save and search
					if count, _ := s.page.Locator("input[value='この条件を記憶する']").Count(); count > 0 {
						log.Println("Clicking この条件を記憶する after date selection")
						s.page.Click("input[value='この条件を記憶する']")
						time.Sleep(1 * time.Second)
					}

					if count, _ := s.page.Locator("input[value*='検索']").Count(); count > 0 {
						log.Println("Clicking 検索 after date selection")
						s.page.Click("input[value*='検索']")
						time.Sleep(3 * time.Second)
					}
				} else {
					log.Println("No date selects found after clicking 利用明細")
				}
				break
			}
		}
	}

	time.Sleep(1 * time.Second)
	s.takeScreenshot("05_dates_set")

	// Check if 全選択 is available on this page structure
	hasSelectAll := false
	if count, _ := s.page.Locator("a:has-text('全選択')").Count(); count > 0 {
		hasSelectAll = true
	}

	if hasSelectAll {
		// Step 1: Click 全選択 (Select All)
		log.Println("Step 1: Clicking 全選択 (Select All)...")
		selectAllClicked := false
		selectAllSelectors := []string{
			"a:has-text('全選択')",
			"a[onclick*='allSelected']",
			"a[href*='JavaScript:void']",
			"input[type='checkbox'][name='selectAll']",
			"input[type='checkbox'][id='selectAll']",
		}
		for _, selector := range selectAllSelectors {
			if count, _ := s.page.Locator(selector).Count(); count > 0 {
				log.Printf("Found 全選択 element: %s", selector)
				if err := s.page.Click(selector); err == nil {
					selectAllClicked = true
					log.Println("全選択 clicked")
					time.Sleep(1 * time.Second)
					break
				}
			}
		}
		if !selectAllClicked {
			log.Println("全選択 element not found, continuing anyway")
		}
	} else {
		log.Println("全選択 not available on this page structure, skipping...")
	}

	// Step 1.5: Select 走行区分 全て (All travel categories)
	log.Println("Step 1.5: Selecting 走行区分 全て...")
	travelCategorySelected := false
	// Try to find and select the 走行区分 dropdown
	travelSelectors := []string{
		"select[name*='soukou']",
		"select[name*='travel']",
		"select[name*='kubun']",
		"select#soukou_kubun",
		"select[name='soukouKubun']",
	}

	for _, selector := range travelSelectors {
		if count, _ := s.page.Locator(selector).Count(); count > 0 {
			log.Printf("Found 走行区分 dropdown: %s", selector)
			// Select the "全て" option (usually value="0" or empty)
			if _, err := s.page.SelectOption(selector, playwright.SelectOptionValues{Values: &[]string{"0"}}); err != nil {
				// Try empty value
				if _, err := s.page.SelectOption(selector, playwright.SelectOptionValues{Values: &[]string{""}}); err != nil {
					// Try selecting by text
					if _, err := s.page.SelectOption(selector, playwright.SelectOptionValues{Labels: &[]string{"全て"}}); err != nil {
						log.Printf("Failed to select 全て: %v", err)
					} else {
						travelCategorySelected = true
						log.Println("走行区分 全て selected by text")
					}
				} else {
					travelCategorySelected = true
					log.Println("走行区分 全て selected with empty value")
				}
			} else {
				travelCategorySelected = true
				log.Println("走行区分 全て selected with value 0")
			}
			if travelCategorySelected {
				time.Sleep(500 * time.Millisecond)
				break
			}
		}
	}

	if !travelCategorySelected {
		log.Println("走行区分 dropdown not found, continuing anyway")
	}

	// Check if この条件を記憶する is available
	hasSaveCondition := false
	if count, _ := s.page.Locator("input[value*='この条件を記憶する']").Count(); count > 0 {
		hasSaveCondition = true
	}

	if hasSaveCondition {
		// Step 2: Click この条件を記憶する (Save this condition)
		log.Println("Step 2: Clicking この条件を記憶する...")
		saveConditionClicked := false
		saveSelectors := []string{
			"input[type='button'][value='この条件を記憶する']",
			"input[type='submit'][value='この条件を記憶する']",
			"button:has-text('この条件を記憶する')",
			"input[value*='記憶']",
		}
		for _, selector := range saveSelectors {
			if count, _ := s.page.Locator(selector).Count(); count > 0 {
				log.Printf("Found この条件を記憶する button: %s", selector)
				if err := s.page.Click(selector); err == nil {
					saveConditionClicked = true
					log.Println("この条件を記憶する clicked")
					time.Sleep(1 * time.Second)
					break
				}
			}
		}
		if !saveConditionClicked {
			log.Println("この条件を記憶する button not found, continuing anyway")
		}
	} else {
		log.Println("この条件を記憶する not available on this page structure, skipping...")
	}

	// Step 3: Click 検索 (Search)
	log.Println("Step 3: Clicking 検索 button...")
	searchClicked := false
	searchSelectors := []string{
		"input[type='button'][value='  検索  ']", // With spaces
		"input[type='submit'][value='  検索  ']", // With spaces
		"input[type='button'][value*='検索']", // Contains 検索
		"input[type='submit'][value*='検索']", // Contains 検索
	}

	for _, selector := range searchSelectors {
		if count, _ := s.page.Locator(selector).Count(); count > 0 {
			log.Printf("Found search button: %s", selector)
			if err := s.page.Click(selector); err == nil {
				searchClicked = true
				log.Println("検索 button clicked, waiting for results...")
				time.Sleep(5 * time.Second)
				break
			}
		}
	}

	if !searchClicked {
		log.Println("検索 button not found")
	}

	s.takeScreenshot("05b_after_search")

	// Check if we're already on the 利用明細 page after search
	currentURL := s.page.URL()
	log.Printf("Current URL after search: %s", currentURL)

	// Check if we need to click 利用明細 or if we're already there
	// Also check for different URL patterns
	onMeisaiPage := strings.Contains(currentURL, "meisai") ||
		strings.Contains(currentURL, "detail") ||
		strings.Contains(currentURL, "1033") || // Some accounts use function code
		strings.Contains(currentURL, "1014") // Alternative function code

	if !onMeisaiPage {
		// Now click 利用明細 to proceed with the selected date range
		log.Println("Navigating to 利用明細 with selected date range...")

		menuClicked := false
		menuSelectors := []string{
			"a:has-text('利用明細')",
			"a:has-text('明細照会')",
			"a:has-text('明細検索')",
			"a:has-text('照会')",
			"input[type='button'][value*='明細']",
			"input[type='submit'][value*='明細']",
			"*:has-text('利用明細'):not(title)",
		}

		for _, selector := range menuSelectors {
			if count, _ := s.page.Locator(selector).Count(); count > 0 {
				log.Printf("Clicking menu: %s", selector)
				s.page.Click(selector)
				menuClicked = true
				time.Sleep(time.Second * 5) // Wait longer for page load
				break
			}
		}

		if !menuClicked {
			log.Println("Could not find menu item")
		}
	} else {
		log.Println("Already on 利用明細 page after search")
	}

	s.takeScreenshot("06_meisai_page")

	// Wait for page to load completely
	time.Sleep(3 * time.Second)

	// The 利用明細 page should now show statements for the selected date range
	log.Println("Ready to download CSV...")

	// For some accounts, we might need to click 利用明細 link even after search
	if count, _ := s.page.Locator("input[value*='ＣＳＶ']").Count(); count == 0 {
		log.Println("CSV button not found, checking if we need to navigate to 利用明細...")

		// Try to find and click 利用明細 link
		detailLinks := []string{
			"a:has-text('利用明細')",
			"a:has-text('明細')",
			"td:has-text('利用明細')",
			"span:has-text('利用明細')",
			"div:has-text('利用明細照会')",
		}

		for _, selector := range detailLinks {
			if count, _ := s.page.Locator(selector).Count(); count > 0 {
				log.Printf("Found detail link: %s, clicking...", selector)
				s.page.Click(selector)
				time.Sleep(3 * time.Second)

				// After clicking 利用明細, check if we have month buttons for date selection
				s.handleMonthButtonDateSelection(fromDate, toDate)
				break
			}
		}
	}

	// Simply click the CSV download button
	log.Println("Looking for CSV download button...")

	// First, check what buttons are available on the page
	log.Println("Checking available buttons on page...")
	if buttons, _ := s.page.Locator("input[type='button'], input[type='submit'], button, a.button").All(); len(buttons) > 0 {
		log.Printf("Found %d button elements on page", len(buttons))
		for i, btn := range buttons {
			if i < 10 { // Log first 10 buttons
				value, _ := btn.GetAttribute("value")
				text, _ := btn.TextContent()

				if value != "" {
					log.Printf("  Button %d: value='%s'", i+1, value)
				}
				if text != "" && text != value {
					log.Printf("  Button %d: text='%s'", i+1, text)
				}
			}
		}
	}

	downloadSelectors := []string{
		"input[type='button'][value='利用明細ＣＳＶ出力']",
		"input[type='submit'][value='利用明細ＣＳＶ出力']",
		"input[value='利用明細ＣＳＶ出力']",
		"input[type='button'][value*='ＣＳＶ']",
		"input[type='submit'][value*='ＣＳＶ']",
		"input[type='button'][value*='CSV']",
		"input[type='submit'][value*='CSV']",
		"button:has-text('CSV')",
		"button:has-text('ＣＳＶ')",
		"a:has-text('CSV出力')",
		"a:has-text('ＣＳＶ出力')",
		"input[value*='明細'][value*='出力']",
		"input[value*='ダウンロード']",
		"a[href*='download']",
		"a[href*='csv']",
	}

	for _, selector := range downloadSelectors {
		if count, _ := s.page.Locator(selector).Count(); count > 0 {
			log.Printf("Found download button: %s", selector)

			// Record the number of pages before clicking
			pagesBefore := len(s.context.Pages())
			log.Printf("Pages before click: %d", pagesBefore)

			// Set up alert handler before clicking
			log.Println("Setting up alert handler...")
			s.page.OnDialog(func(dialog playwright.Dialog) {
				log.Printf("Alert detected: %s", dialog.Message())
				// Accept the alert
				dialog.Accept()
			})

			// Try to handle download with ExpectDownload
			log.Printf("Attempting download with ExpectDownload...")
			download, err := s.page.ExpectDownload(func() error {
				return s.page.Click(selector)
			}, playwright.PageExpectDownloadOptions{
				Timeout: playwright.Float(10000),
			})

			if err != nil {
				log.Printf("ExpectDownload failed: %v", err)
				// Try regular click anyway
				log.Println("Trying regular click...")
				s.page.Click(selector)
				time.Sleep(5 * time.Second)
				return "Download attempted without file capture", nil
			}

			// Save the download
			suggestedFilename := download.SuggestedFilename()
			downloadPath := filepath.Join(s.config.DownloadPath, suggestedFilename)

			log.Printf("Saving download: %s", downloadPath)
			if err := download.SaveAs(downloadPath); err != nil {
				log.Printf("Failed to save download: %v", err)
				return "", fmt.Errorf("failed to save download: %w", err)
			}

			log.Printf("✅ Downloaded successfully: %s", downloadPath)
			return downloadPath, nil
		}
	}

	return "", fmt.Errorf("could not download CSV")
}

// takeScreenshot captures a screenshot
func (s *ActualETCScraper) takeScreenshot(name string) {
	// スクリーンショット機能を無効化（ページの拡大縮小を防ぐため）
	return
}

// handleError provides centralized error handling with context
func (s *ActualETCScraper) handleError(context string, err error) error {
	if err == nil {
		return nil
	}

	// Take screenshot for debugging
	screenshotName := fmt.Sprintf("error_%s_%d", context, time.Now().Unix())
	s.takeScreenshot(screenshotName)

	// Log detailed error information
	log.Printf("[ERROR] %s: %v", context, err)

	// Return wrapped error with context
	return fmt.Errorf("%s: %w", context, err)
}

// WaitForDownloadWithProgress waits for download with progress tracking
func (s *ActualETCScraper) WaitForDownloadWithProgress(timeout time.Duration) (string, error) {
	timer := time.NewTimer(timeout)
	defer timer.Stop()

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	startTime := time.Now()

	for {
		select {
		case downloadPath := <-s.downloadChan:
			if downloadPath == "" {
				return "", fmt.Errorf("download failed")
			}
			elapsed := time.Since(startTime)
			log.Printf("Download completed in %v: %s", elapsed, downloadPath)
			return downloadPath, nil

		case <-ticker.C:
			elapsed := time.Since(startTime)
			log.Printf("Waiting for download... (%v/%v)", elapsed, timeout)

		case <-timer.C:
			return "", fmt.Errorf("download timeout after %v", timeout)
		}
	}
}

// handleMonthButtonDateSelection handles date selection using month buttons
func (s *ActualETCScraper) handleMonthButtonDateSelection(fromDate, toDate time.Time) {
	log.Println("Checking for month button date selection...")

	// Check if we have month buttons (like 8月, 9月, etc.)
	// First, let's see what buttons are actually on the page
	if buttons, _ := s.page.Locator("button, input[type='button'], a").All(); len(buttons) > 0 {
		log.Printf("Checking %d clickable elements for month buttons", len(buttons))
		monthButtonsFound := false

		for _, btn := range buttons {
			text, _ := btn.TextContent()
			value, _ := btn.GetAttribute("value")

			// Check if this is a month button
			if strings.Contains(text, "月") || strings.Contains(value, "月") {
				if !monthButtonsFound {
					log.Println("Found month selection buttons on page")
					monthButtonsFound = true
				}
			}
		}

		if monthButtonsFound {
			// For the date range, we need to select the appropriate months
			fromMonth := int(fromDate.Month())
			toMonth := int(toDate.Month())

			log.Printf("Need to select months from %d月 to %d月", fromMonth, toMonth)

			// Try to click the month buttons for our date range
			for month := fromMonth; month <= toMonth; month++ {
				monthText := fmt.Sprintf("%d月", month)

				// Try to find and click the month button
				clicked := false

				// First try exact text match
				for _, btn := range buttons {
					text, _ := btn.TextContent()
					text = strings.TrimSpace(text)

					if text == monthText || strings.Contains(text, monthText) {
						log.Printf("Found and clicking month button: %s", monthText)
						btn.Click()
						time.Sleep(500 * time.Millisecond)
						clicked = true
						break
					}
				}

				if !clicked {
					log.Printf("Could not find button for %s", monthText)
				}
			}
		} else {
			log.Println("No month buttons found on this page")
		}

		// After selecting months, we need to follow the correct button sequence
		log.Println("Following button sequence after month selection...")

		// First check if we need to click この条件を記憶する
		if count, _ := s.page.Locator("input[value='この条件を記憶する']").Count(); count > 0 {
			log.Println("Clicking この条件を記憶する after month selection")
			s.page.Click("input[value='この条件を記憶する']")
			time.Sleep(1 * time.Second)
		}

		// Then click 検索
		searchButtons := []string{
			"input[type='button'][value='  検索  ']", // With spaces
			"input[type='button'][value*='検索']",
			"input[type='submit'][value*='検索']",
		}

		for _, selector := range searchButtons {
			if count, _ := s.page.Locator(selector).Count(); count > 0 {
				log.Printf("Clicking search button: %s", selector)
				s.page.Click(selector)
				time.Sleep(3 * time.Second)
				break
			}
		}

		// After search, we might need to click 利用明細 again to see the results
		if count, _ := s.page.Locator("a:has-text('利用明細')").Count(); count > 0 {
			log.Println("Clicking 利用明細 to see filtered results")
			s.page.Click("a:has-text('利用明細')")
			time.Sleep(3 * time.Second)
		}
	} else {
		log.Println("No month buttons found")
	}
}

// Close cleans up resources
func (s *ActualETCScraper) Close() {
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
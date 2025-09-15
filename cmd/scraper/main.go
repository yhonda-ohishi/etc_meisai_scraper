package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/playwright-community/playwright-go"
)

func main() {
	// Load .env file
	if err := godotenv.Load("../../.env"); err != nil {
		log.Println("No .env file found")
	}

	// Command line flags
	var (
		userID   = flag.String("user", os.Getenv("ETC_USER_ID"), "ETC User ID")
		password = flag.String("pass", os.Getenv("ETC_PASSWORD"), "ETC Password")
		headless = flag.Bool("headless", false, "Run in headless mode")
		debug    = flag.Bool("debug", false, "Enable debug mode")
	)
	flag.Parse()

	if *userID == "" || *password == "" {
		log.Fatal("User ID and Password are required. Set via flags or environment variables.")
	}

	// Install playwright
	err := playwright.Install()
	if err != nil {
		log.Fatalf("Could not install playwright: %v", err)
	}

	// Start Playwright
	pw, err := playwright.Run()
	if err != nil {
		log.Fatalf("Could not start playwright: %v", err)
	}
	defer pw.Stop()

	// Launch browser
	browser, err := pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
		Headless: playwright.Bool(*headless),
		SlowMo:   playwright.Float(100), // Slow down actions for debugging
	})
	if err != nil {
		log.Fatalf("Could not launch browser: %v", err)
	}
	defer browser.Close()

	// Create context
	context, err := browser.NewContext(playwright.BrowserNewContextOptions{
		Viewport: &playwright.Size{
			Width:  1920,
			Height: 1080,
		},
		AcceptDownloads: playwright.Bool(true),
	})
	if err != nil {
		log.Fatalf("Could not create context: %v", err)
	}
	defer context.Close()

	// Create page
	page, err := context.NewPage()
	if err != nil {
		log.Fatalf("Could not create page: %v", err)
	}
	defer page.Close()

	// Enable console logging if debug mode
	if *debug {
		page.On("console", func(msg playwright.ConsoleMessage) {
			log.Printf("Console [%s]: %s", msg.Type(), msg.Text())
		})
	}

	log.Println("Starting ETC site scraping test...")

	// Navigate to ETC site
	log.Println("Navigating to https://www.etc-meisai.jp/")
	_, err = page.Goto("https://www.etc-meisai.jp/", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateNetworkidle,
		Timeout:   playwright.Float(30000),
	})
	if err != nil {
		log.Fatalf("Failed to navigate: %v", err)
	}

	// Take screenshot of login page
	page.Screenshot(playwright.PageScreenshotOptions{
		Path: playwright.String("screenshots/01_login_page.png"),
	})

	// Try to find login form elements
	log.Println("Looking for login form elements...")

	// Common selectors to try
	userSelectors := []string{
		"input[name='userId']",
		"input[name='user_id']",
		"input[name='loginId']",
		"input[name='login_id']",
		"#userId",
		"#user_id",
		"#loginId",
		"input[type='text']",
	}

	passwordSelectors := []string{
		"input[name='password']",
		"input[name='passwd']",
		"input[name='pass']",
		"#password",
		"#passwd",
		"input[type='password']",
	}

	// Find user ID field
	var userFieldFound bool
	for _, selector := range userSelectors {
		count, _ := page.Locator(selector).Count()
		if count > 0 {
			log.Printf("Found user field with selector: %s", selector)
			err = page.Fill(selector, *userID)
			if err == nil {
				userFieldFound = true
				break
			}
		}
	}

	if !userFieldFound {
		log.Println("Could not find user ID field")
		// Print all input fields for debugging
		inputs, _ := page.Locator("input").All()
		for i, input := range inputs {
			name, _ := input.GetAttribute("name")
			id, _ := input.GetAttribute("id")
			inputType, _ := input.GetAttribute("type")
			placeholder, _ := input.GetAttribute("placeholder")
			log.Printf("Input %d: name=%s, id=%s, type=%s, placeholder=%s", i, name, id, inputType, placeholder)
		}
	}

	// Find password field
	var passwordFieldFound bool
	for _, selector := range passwordSelectors {
		count, _ := page.Locator(selector).Count()
		if count > 0 {
			log.Printf("Found password field with selector: %s", selector)
			err = page.Fill(selector, *password)
			if err == nil {
				passwordFieldFound = true
				break
			}
		}
	}

	if !passwordFieldFound {
		log.Println("Could not find password field")
	}

	// Take screenshot after filling
	page.Screenshot(playwright.PageScreenshotOptions{
		Path: playwright.String("screenshots/02_filled_form.png"),
	})

	// Find and click login button
	log.Println("Looking for login button...")

	loginButtonSelectors := []string{
		"button[type='submit']",
		"input[type='submit']",
		"button:has-text('ログイン')",
		"input[value='ログイン']",
		"button:has-text('Login')",
		"input[value='Login']",
		".login-button",
		"#login-button",
	}

	var loginButtonFound bool
	for _, selector := range loginButtonSelectors {
		count, _ := page.Locator(selector).Count()
		if count > 0 {
			log.Printf("Found login button with selector: %s", selector)
			err = page.Click(selector)
			if err == nil {
				loginButtonFound = true
				break
			}
		}
	}

	if !loginButtonFound {
		log.Println("Could not find login button")
		// Print all buttons for debugging
		buttons, _ := page.Locator("button").All()
		for i, button := range buttons {
			text, _ := button.TextContent()
			buttonType, _ := button.GetAttribute("type")
			log.Printf("Button %d: text=%s, type=%s", i, text, buttonType)
		}

		// Also check input[type=submit]
		submits, _ := page.Locator("input[type='submit']").All()
		for i, submit := range submits {
			value, _ := submit.GetAttribute("value")
			log.Printf("Submit %d: value=%s", i, value)
		}
	}

	if loginButtonFound {
		// Wait for navigation
		log.Println("Waiting for login to complete...")
		time.Sleep(5 * time.Second)

		// Take screenshot after login
		page.Screenshot(playwright.PageScreenshotOptions{
			Path: playwright.String("screenshots/03_after_login.png"),
		})

		currentURL := page.URL()
		log.Printf("Current URL after login: %s", currentURL)

		// Check if login was successful
		// Look for indicators of successful login
		successIndicators := []string{
			"ログアウト",
			"logout",
			"明細",
			"meisai",
			"利用明細",
		}

		for _, indicator := range successIndicators {
			count, _ := page.Locator(fmt.Sprintf("*:has-text('%s')", indicator)).Count()
			if count > 0 {
				log.Printf("Login appears successful - found: %s", indicator)
				break
			}
		}

		// Try to navigate to search/download page
		log.Println("Looking for search/download functionality...")

		searchLinks := []string{
			"a:has-text('明細')",
			"a:has-text('検索')",
			"a:has-text('ダウンロード')",
			"a:has-text('CSV')",
			"a[href*='search']",
			"a[href*='meisai']",
			"a[href*='download']",
		}

		for _, selector := range searchLinks {
			count, _ := page.Locator(selector).Count()
			if count > 0 {
				href, _ := page.Locator(selector).First().GetAttribute("href")
				text, _ := page.Locator(selector).First().TextContent()
				log.Printf("Found link: text=%s, href=%s", text, href)
			}
		}

		// Take final screenshot
		page.Screenshot(playwright.PageScreenshotOptions{
			Path:     playwright.String("screenshots/04_main_page.png"),
			FullPage: playwright.Bool(true),
		})
	}

	log.Println("Test completed. Check screenshots directory for results.")
}
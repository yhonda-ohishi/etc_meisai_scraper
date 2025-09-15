package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/playwright-community/playwright-go"
	"github.com/yhonda-ohishi/etc_meisai/config"
)

func main() {
	// Load .env
	godotenv.Load("../../.env")
	godotenv.Load(".env")

	// Load account
	accounts, err := config.LoadCorporateAccountsFromEnv()
	if err != nil || len(accounts) == 0 {
		log.Fatal("No accounts found")
	}

	account := accounts[0]
	log.Printf("Using account: %s", account.UserID)

	// Initialize Playwright
	playwright.Install()
	pw, _ := playwright.Run()
	defer pw.Stop()

	browser, _ := pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
		Headless: playwright.Bool(false),
	})
	defer browser.Close()

	context, _ := browser.NewContext(playwright.BrowserNewContextOptions{
		Viewport: &playwright.Size{Width: 1920, Height: 1080},
	})
	defer context.Close()

	page, _ := context.NewPage()
	defer page.Close()

	// Navigate and login
	log.Println("Navigating to ETC site...")
	page.Goto("https://www2.etc-meisai.jp/etc/R?funccode=1013000000&nextfunc=1013000000",
		playwright.PageGotoOptions{WaitUntil: playwright.WaitUntilStateNetworkidle})

	time.Sleep(2 * time.Second)

	// Login
	log.Println("Logging in...")
	page.Fill("input[type='text']:visible", account.UserID)
	page.Fill("input[type='password']:visible", account.Password)
	page.Click("input[type='button'][value*='ログイン']")

	time.Sleep(5 * time.Second)

	// After login, check what's on the main page
	log.Println("\n=== Main Page After Login ===")
	log.Printf("URL: %s", page.URL())

	// Check for select dropdowns (年月選択など)
	selects, _ := page.Locator("select").All()
	log.Printf("\nFound %d select elements", len(selects))
	for i, sel := range selects {
		name, _ := sel.GetAttribute("name")
		id, _ := sel.GetAttribute("id")
		// Get selected option
		value, _ := sel.InputValue()
		fmt.Printf("Select %d: name=%s, id=%s, selected=%s\n", i, name, id, value)
	}

	// Check for date-related elements
	log.Println("\n=== Looking for date-related elements ===")

	// Check for elements with 年月 in their text
	yearMonthElements, _ := page.Locator("*:has-text('年月')").All()
	log.Printf("Found %d elements with '年月': ", len(yearMonthElements))

	// Check for elements with specific date patterns
	datePatterns := []string{
		"select[name*='year']",
		"select[name*='month']",
		"select[name*='nen']",
		"select[name*='getsu']",
		"select[name*='tsuki']",
		"input[name*='date']",
		"input[name*='year']",
		"input[name*='month']",
	}

	for _, pattern := range datePatterns {
		elements, _ := page.Locator(pattern).All()
		if len(elements) > 0 {
			log.Printf("Pattern '%s' found %d elements", pattern, len(elements))
			for j, elem := range elements {
				name, _ := elem.GetAttribute("name")
				value, _ := elem.InputValue()
				fmt.Printf("  Element %d: name=%s, value=%s\n", j, name, value)
			}
		}
	}

	// Click 利用明細
	log.Println("\n=== Clicking 利用明細 ===")
	if err := page.Click("a:has-text('利用明細')"); err != nil {
		log.Printf("Failed to click: %v", err)
	}

	time.Sleep(5 * time.Second)

	// Check what's on the 利用明細 page
	log.Println("\n=== 利用明細 Page ===")
	log.Printf("URL: %s", page.URL())

	// Check for select dropdowns again
	selects, _ = page.Locator("select").All()
	log.Printf("\nFound %d select elements on 利用明細 page", len(selects))
	for i, sel := range selects {
		name, _ := sel.GetAttribute("name")
		id, _ := sel.GetAttribute("id")
		value, _ := sel.InputValue()
		fmt.Printf("Select %d: name=%s, id=%s, selected=%s\n", i, name, id, value)

		// Get options for date-related selects
		if name != "" {
			options, _ := sel.Locator("option").All()
			if len(options) <= 20 { // Only show options for small lists
				fmt.Printf("  Options:\n")
				for j, opt := range options {
					text, _ := opt.TextContent()
					val, _ := opt.GetAttribute("value")
					fmt.Printf("    [%d] value=%s, text=%s\n", j, val, text)
				}
			}
		}
	}

	// Take screenshot
	os.MkdirAll("./test_output", 0755)
	page.Screenshot(playwright.PageScreenshotOptions{
		Path:     playwright.String("./test_output/date_field_test.png"),
		FullPage: playwright.Bool(true),
	})

	log.Println("\nScreenshot saved. Press Enter to exit...")
	fmt.Scanln()
}
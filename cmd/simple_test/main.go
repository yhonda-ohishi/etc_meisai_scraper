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

	// Calculate dates
	now := time.Now()
	lastMonth := now.AddDate(0, -1, 0)
	fromDate := time.Date(lastMonth.Year(), lastMonth.Month(), 1, 0, 0, 0, 0, time.Local)
	toDate := time.Date(lastMonth.Year(), lastMonth.Month()+1, 0, 0, 0, 0, 0, time.Local)

	log.Printf("Date range: %s to %s", fromDate.Format("2006-01-02"), toDate.Format("2006-01-02"))

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

	// After login, click 利用明細
	log.Println("Clicking 利用明細...")
	if err := page.Click("a:has-text('利用明細')"); err != nil {
		log.Printf("Failed to click: %v", err)
	}

	time.Sleep(5 * time.Second)

	// Now check what's on the page
	log.Println("\n=== Page Analysis ===")

	// Check URL
	url := page.URL()
	log.Printf("Current URL: %s", url)

	// Check for forms
	forms, _ := page.Locator("form").All()
	log.Printf("Found %d forms", len(forms))

	// Check for all inputs
	inputs, _ := page.Locator("input").All()
	log.Printf("Found %d inputs", len(inputs))

	// List all visible inputs
	visibleInputs, _ := page.Locator("input:visible").All()
	log.Printf("Found %d visible inputs", len(visibleInputs))

	for i, input := range visibleInputs {
		inputType, _ := input.GetAttribute("type")
		name, _ := input.GetAttribute("name")
		id, _ := input.GetAttribute("id")
		value, _ := input.GetAttribute("value")
		fmt.Printf("Visible Input %d: type=%s, name=%s, id=%s, value=%s\n",
			i, inputType, name, id, value)
	}

	// Check for select elements (dropdowns)
	selects, _ := page.Locator("select").All()
	log.Printf("\nFound %d select elements", len(selects))

	for i, sel := range selects {
		name, _ := sel.GetAttribute("name")
		id, _ := sel.GetAttribute("id")
		fmt.Printf("Select %d: name=%s, id=%s\n", i, name, id)
	}

	// Check for buttons
	buttons, _ := page.Locator("button, input[type='button'], input[type='submit']").All()
	log.Printf("\nFound %d buttons", len(buttons))

	for i, btn := range buttons {
		btnType, _ := btn.GetAttribute("type")
		value, _ := btn.GetAttribute("value")
		text, _ := btn.TextContent()
		fmt.Printf("Button %d: type=%s, value=%s, text=%s\n",
			i, btnType, value, text)
	}

	// Take screenshot
	os.MkdirAll("./test_output", 0755)
	page.Screenshot(playwright.PageScreenshotOptions{
		Path:     playwright.String("./test_output/search_page.png"),
		FullPage: playwright.Bool(true),
	})

	log.Println("\nScreenshot saved to test_output/search_page.png")
	log.Println("Check the browser window and screenshot to see what's available")

	// Keep browser open for manual inspection
	fmt.Println("\nPress Enter to exit...")
	fmt.Scanln()
}
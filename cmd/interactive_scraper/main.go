package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/playwright-community/playwright-go"
	"github.com/yhonda-ohishi/etc_meisai/config"
)

func main() {
	// Load .env file
	if err := godotenv.Load("../../.env"); err != nil {
		if err := godotenv.Load(".env"); err != nil {
			log.Println("No .env file found")
		}
	}

	// Command line flags
	var (
		accountIndex = flag.Int("account", 0, "Account index to use (0-based)")
	)
	flag.Parse()

	// Load corporate accounts
	accounts, err := config.LoadCorporateAccountsFromEnv()
	if err != nil {
		log.Fatalf("Failed to load accounts: %v", err)
	}

	if *accountIndex >= len(accounts) {
		log.Fatalf("Account index %d out of range (have %d accounts)", *accountIndex, len(accounts))
	}

	account := accounts[*accountIndex]
	log.Printf("Using account: %s", account.UserID)

	// Create download directory
	os.MkdirAll("./downloads/interactive", 0755)

	// Initialize playwright directly for more control
	err = playwright.Install()
	if err != nil {
		log.Fatalf("Could not install playwright: %v", err)
	}

	pw, err := playwright.Run()
	if err != nil {
		log.Fatalf("Could not start playwright: %v", err)
	}
	defer pw.Stop()

	// Force browser to be visible
	browser, err := pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
		Headless: playwright.Bool(false), // Always show browser
		SlowMo:   playwright.Float(100),
	})
	if err != nil {
		log.Fatalf("Could not launch browser: %v", err)
	}
	defer browser.Close()

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

	page, err := context.NewPage()
	if err != nil {
		log.Fatalf("Could not create page: %v", err)
	}
	defer page.Close()

	// Navigate to login page
	log.Println("Navigating to ETC login page...")
	loginURL := "https://www2.etc-meisai.jp/etc/R?funccode=1013000000&nextfunc=1013000000"
	_, err = page.Goto(loginURL, playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateNetworkidle,
	})
	if err != nil {
		log.Fatalf("Failed to navigate: %v", err)
	}

	// Wait for page to load
	time.Sleep(2 * time.Second)

	// Login
	log.Println("Logging in...")

	// Fill user ID
	userFilled := false
	userSelectors := []string{
		"input[name='usrid']",
		"input[name='userId']",
		"#userId",
		"input[type='text']:not([type='hidden'])",
	}

	for _, selector := range userSelectors {
		if count, _ := page.Locator(selector).Count(); count > 0 {
			page.Fill(selector, account.UserID)
			userFilled = true
			log.Printf("Filled user ID with selector: %s", selector)
			break
		}
	}

	if !userFilled {
		log.Fatal("Could not find user ID field")
	}

	// Fill password
	passFilled := false
	passwordSelectors := []string{
		"input[name='password']",
		"input[type='password']:not([type='hidden'])",
	}

	for _, selector := range passwordSelectors {
		if count, _ := page.Locator(selector).Count(); count > 0 {
			page.Fill(selector, account.Password)
			passFilled = true
			log.Printf("Filled password with selector: %s", selector)
			break
		}
	}

	if !passFilled {
		log.Fatal("Could not find password field")
	}

	// Click login button
	loginSelectors := []string{
		"input[type='button'][value*='ログイン']",
		"input[type='submit'][value*='ログイン']",
		"button:has-text('ログイン')",
	}

	for _, selector := range loginSelectors {
		if count, _ := page.Locator(selector).Count(); count > 0 {
			page.Click(selector)
			log.Printf("Clicked login button: %s", selector)
			break
		}
	}

	// Wait for login
	time.Sleep(5 * time.Second)

	// Check login success
	if count, _ := page.Locator("*:has-text('ログアウト')").Count(); count > 0 {
		log.Println("Login successful!")
	} else {
		log.Println("Login status uncertain, continuing...")
	}

	// Check for date fields on main page immediately after login
	log.Println("\n=== Checking for date fields on main page ===")
	textInputs, _ := page.Locator("input[type='text']:visible").All()
	log.Printf("Found %d visible text inputs", len(textInputs))

	for i, input := range textInputs {
		name, _ := input.GetAttribute("name")
		id, _ := input.GetAttribute("id")
		value, _ := input.GetAttribute("value")
		placeholder, _ := input.GetAttribute("placeholder")
		maxLength, _ := input.GetAttribute("maxlength")
		fmt.Printf("Text input %d: name=%s, id=%s, value=%s, placeholder=%s, maxlength=%s\n",
			i, name, id, value, placeholder, maxLength)
	}

	// Set default date range (last month)
	now := time.Now()
	lastMonth := now.AddDate(0, -1, 0)
	fromDate := time.Date(lastMonth.Year(), lastMonth.Month(), 1, 0, 0, 0, 0, time.Local)
	toDate := time.Date(lastMonth.Year(), lastMonth.Month()+1, 0, 0, 0, 0, 0, time.Local) // Last day of last month

	fmt.Println("\n=== Interactive Mode ===")
	fmt.Printf("Default Date Range: %s to %s\n", fromDate.Format("2006-01-02"), toDate.Format("2006-01-02"))
	fmt.Println("\nCommands:")
	fmt.Println("  url              - Show current URL")
	fmt.Println("  screenshot [name] - Take screenshot")
	fmt.Println("  click <selector> - Click element")
	fmt.Println("  fill <selector> <value> - Fill input field")
	fmt.Println("  text <selector>  - Get element text")
	fmt.Println("  list <selector>  - List all matching elements")
	fmt.Println("  wait <seconds>   - Wait for N seconds")
	fmt.Println("  goto <url>       - Navigate to URL")
	fmt.Println("  eval <js>        - Execute JavaScript")
	fmt.Println("  search           - Search with default date range")
	fmt.Println("  quit             - Exit")
	fmt.Println("")

	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Print("> ")
		if !scanner.Scan() {
			break
		}

		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		parts := strings.Fields(line)
		command := parts[0]

		switch command {
		case "url":
			fmt.Printf("Current URL: %s\n", page.URL())

		case "screenshot":
			name := "screenshot"
			if len(parts) > 1 {
				name = parts[1]
			}
			filename := fmt.Sprintf("downloads/interactive/%s_%s.png", name, time.Now().Format("20060102_150405"))
			page.Screenshot(playwright.PageScreenshotOptions{
				Path:     playwright.String(filename),
				FullPage: playwright.Bool(true),
			})
			fmt.Printf("Screenshot saved: %s\n", filename)

		case "click":
			if len(parts) < 2 {
				fmt.Println("Usage: click <selector>")
				continue
			}
			selector := strings.Join(parts[1:], " ")
			if err := page.Click(selector); err != nil {
				fmt.Printf("Error: %v\n", err)
			} else {
				fmt.Println("Clicked successfully")
			}

		case "fill":
			if len(parts) < 3 {
				fmt.Println("Usage: fill <selector> <value>")
				continue
			}
			selector := parts[1]
			value := strings.Join(parts[2:], " ")
			if err := page.Fill(selector, value); err != nil {
				fmt.Printf("Error: %v\n", err)
			} else {
				fmt.Println("Filled successfully")
			}

		case "text":
			if len(parts) < 2 {
				fmt.Println("Usage: text <selector>")
				continue
			}
			selector := strings.Join(parts[1:], " ")
			text, err := page.Locator(selector).TextContent()
			if err != nil {
				fmt.Printf("Error: %v\n", err)
			} else {
				fmt.Printf("Text: %s\n", text)
			}

		case "list":
			if len(parts) < 2 {
				fmt.Println("Usage: list <selector>")
				continue
			}
			selector := strings.Join(parts[1:], " ")
			elements, err := page.Locator(selector).All()
			if err != nil {
				fmt.Printf("Error: %v\n", err)
			} else {
				fmt.Printf("Found %d elements:\n", len(elements))
				for i, elem := range elements {
					text, _ := elem.TextContent()
					fmt.Printf("  [%d] %s\n", i, text)
				}
			}

		case "wait":
			if len(parts) < 2 {
				fmt.Println("Usage: wait <seconds>")
				continue
			}
			var seconds int
			fmt.Sscanf(parts[1], "%d", &seconds)
			fmt.Printf("Waiting %d seconds...\n", seconds)
			time.Sleep(time.Duration(seconds) * time.Second)

		case "goto":
			if len(parts) < 2 {
				fmt.Println("Usage: goto <url>")
				continue
			}
			url := parts[1]
			_, err := page.Goto(url, playwright.PageGotoOptions{
				WaitUntil: playwright.WaitUntilStateNetworkidle,
			})
			if err != nil {
				fmt.Printf("Error: %v\n", err)
			} else {
				fmt.Printf("Navigated to: %s\n", page.URL())
			}

		case "eval":
			if len(parts) < 2 {
				fmt.Println("Usage: eval <javascript>")
				continue
			}
			js := strings.Join(parts[1:], " ")
			result, err := page.Evaluate(js)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
			} else {
				fmt.Printf("Result: %v\n", result)
			}

		case "search":
			// Auto search with last month's date range
			fmt.Printf("Searching from %s to %s...\n", fromDate.Format("2006-01-02"), toDate.Format("2006-01-02"))

			// Click on 利用明細
			if err := page.Click("a:has-text('利用明細')"); err != nil {
				fmt.Printf("Failed to click 利用明細: %v\n", err)
			}

			time.Sleep(3 * time.Second)

			// Fill date fields (try multiple formats)
			dateFormats := []string{
				"2006/01/02",
				"20060102",
				"2006-01-02",
			}

			for _, format := range dateFormats {
				// Try to fill from date
				selectors := []string{"input[name*='from']", "input[name*='start']", "#fromDate"}
				for _, sel := range selectors {
					page.Fill(sel, fromDate.Format(format))
				}

				// Try to fill to date
				selectors = []string{"input[name*='to']", "input[name*='end']", "#toDate"}
				for _, sel := range selectors {
					page.Fill(sel, toDate.Format(format))
				}
			}

			fmt.Println("Date fields filled, ready to search")

		case "quit", "exit":
			fmt.Println("Exiting...")
			return

		default:
			fmt.Printf("Unknown command: %s\n", command)
		}
	}
}
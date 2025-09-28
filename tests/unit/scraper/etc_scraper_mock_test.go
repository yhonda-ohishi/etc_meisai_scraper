package scraper_test

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/yhonda-ohishi/etc_meisai/src/scraper"
	"github.com/yhonda-ohishi/etc_meisai/tests/mocks"
)

func TestNewETCScraper(t *testing.T) {
	tests := []struct {
		name        string
		config      *scraper.ScraperConfig
		expectError bool
	}{
		{
			name: "with default values",
			config: &scraper.ScraperConfig{
				UserID:   "test",
				Password: "pass",
			},
			expectError: false,
		},
		{
			name: "with custom values",
			config: &scraper.ScraperConfig{
				UserID:       "test",
				Password:     "pass",
				DownloadPath: "./custom",
				Headless:     true,
				TestMode:     true,
				Timeout:      60000,
				RetryCount:   5,
				UserAgent:    "CustomAgent",
				SlowMo:       100,
			},
			expectError: false,
		},
		{
			name: "with nil logger",
			config: &scraper.ScraperConfig{
				UserID:   "test",
				Password: "pass",
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean up test directory
			if tt.config.DownloadPath != "" {
				defer os.RemoveAll(tt.config.DownloadPath)
			} else {
				defer os.RemoveAll("./downloads")
			}

			logger := log.New(os.Stdout, "[TEST] ", log.LstdFlags)
			if tt.name == "with nil logger" {
				logger = nil
			}

			scraper, err := scraper.NewETCScraper(tt.config, logger)
			if (err != nil) != tt.expectError {
				t.Errorf("NewETCScraper() error = %v, expectError %v", err, tt.expectError)
			}
			if err == nil && scraper == nil {
				t.Error("Expected scraper instance, got nil")
			}
		})
	}
}

func TestNewETCScraperWithFactory_NilFactory(t *testing.T) {
	config := &scraper.ScraperConfig{
		UserID:       "test",
		Password:     "pass",
		DownloadPath: "./test_downloads",
	}
	logger := log.New(os.Stdout, "[TEST] ", log.LstdFlags)

	// Test with nil factory
	scraperInstance, err := scraper.NewETCScraperWithFactory(config, logger, nil)

	// Should return error
	if err == nil {
		t.Fatal("Expected error for nil factory, got nil")
	}

	// Should contain specific error message
	if !contains(err.Error(), "factory is required") {
		t.Errorf("Error should contain 'factory is required', got '%s'", err.Error())
	}

	// Scraper should be nil
	if scraperInstance != nil {
		t.Error("Scraper should be nil when error occurs")
	}
}

func TestETCScraper_Initialize(t *testing.T) {
	tests := []struct {
		name          string
		setupMock     func() *mocks.MockPlaywrightFactory
		expectError   bool
		errorContains string
	}{
		{
			name: "successful initialization",
			setupMock: func() *mocks.MockPlaywrightFactory {
				mockPage := mocks.NewMockPage()
				mockContext := &mocks.MockBrowserContext{
					NewPageFunc: func() (scraper.PageInterface, error) {
						return mockPage, nil
					},
				}
				mockBrowser := &mocks.MockBrowser{
					NewContextFunc: func(options scraper.BrowserNewContextOptions) (scraper.BrowserContextInterface, error) {
						return mockContext, nil
					},
				}
				mockChromium := &mocks.MockBrowserType{
					LaunchFunc: func(options scraper.BrowserTypeLaunchOptions) (scraper.BrowserInterface, error) {
						return mockBrowser, nil
					},
				}
				mockPw := &mocks.MockPlaywright{
					Chromium: mockChromium,
				}
				return &mocks.MockPlaywrightFactory{
					RunFunc: func() (scraper.PlaywrightInterface, error) {
						return mockPw, nil
					},
				}
			},
			expectError: false,
		},
		{
			name: "install error",
			setupMock: func() *mocks.MockPlaywrightFactory {
				return &mocks.MockPlaywrightFactory{
					InstallError: errors.New("install failed"),
				}
			},
			expectError:   true,
			errorContains: "could not install playwright",
		},
		{
			name: "run error",
			setupMock: func() *mocks.MockPlaywrightFactory {
				return &mocks.MockPlaywrightFactory{
					RunError: errors.New("run failed"),
				}
			},
			expectError:   true,
			errorContains: "could not start playwright",
		},
		{
			name: "launch browser error",
			setupMock: func() *mocks.MockPlaywrightFactory {
				mockChromium := &mocks.MockBrowserType{
					LaunchError: errors.New("launch failed"),
				}
				mockPw := &mocks.MockPlaywright{
					Chromium: mockChromium,
				}
				return &mocks.MockPlaywrightFactory{
					RunFunc: func() (scraper.PlaywrightInterface, error) {
						return mockPw, nil
					},
				}
			},
			expectError:   true,
			errorContains: "could not launch browser",
		},
		{
			name: "create context error",
			setupMock: func() *mocks.MockPlaywrightFactory {
				mockBrowser := &mocks.MockBrowser{
					NewContextError: errors.New("context failed"),
				}
				mockChromium := &mocks.MockBrowserType{
					LaunchFunc: func(options scraper.BrowserTypeLaunchOptions) (scraper.BrowserInterface, error) {
						return mockBrowser, nil
					},
				}
				mockPw := &mocks.MockPlaywright{
					Chromium: mockChromium,
				}
				return &mocks.MockPlaywrightFactory{
					RunFunc: func() (scraper.PlaywrightInterface, error) {
						return mockPw, nil
					},
				}
			},
			expectError:   true,
			errorContains: "could not create browser context",
		},
		{
			name: "create page error",
			setupMock: func() *mocks.MockPlaywrightFactory {
				mockContext := &mocks.MockBrowserContext{
					NewPageError: errors.New("page creation failed"),
				}
				mockBrowser := &mocks.MockBrowser{
					NewContextFunc: func(options scraper.BrowserNewContextOptions) (scraper.BrowserContextInterface, error) {
						return mockContext, nil
					},
				}
				mockChromium := &mocks.MockBrowserType{
					LaunchFunc: func(options scraper.BrowserTypeLaunchOptions) (scraper.BrowserInterface, error) {
						return mockBrowser, nil
					},
				}
				mockPw := &mocks.MockPlaywright{
					Chromium: mockChromium,
				}
				return &mocks.MockPlaywrightFactory{
					RunFunc: func() (scraper.PlaywrightInterface, error) {
						return mockPw, nil
					},
				}
			},
			expectError:   true,
			errorContains: "could not create page",
		},
		{
			name: "with SlowMo option",
			setupMock: func() *mocks.MockPlaywrightFactory {
				mockPage := mocks.NewMockPage()
				mockContext := &mocks.MockBrowserContext{
					NewPageFunc: func() (scraper.PageInterface, error) {
						return mockPage, nil
					},
				}
				mockBrowser := &mocks.MockBrowser{
					NewContextFunc: func(options scraper.BrowserNewContextOptions) (scraper.BrowserContextInterface, error) {
						return mockContext, nil
					},
				}
				mockChromium := &mocks.MockBrowserType{
					LaunchFunc: func(options scraper.BrowserTypeLaunchOptions) (scraper.BrowserInterface, error) {
						// Verify SlowMo is passed
						if options.SlowMo != nil {
							t.Logf("SlowMo option passed: %f", *options.SlowMo)
						}
						return mockBrowser, nil
					},
				}
				mockPw := &mocks.MockPlaywright{
					Chromium: mockChromium,
				}
				return &mocks.MockPlaywrightFactory{
					RunFunc: func() (scraper.PlaywrightInterface, error) {
						return mockPw, nil
					},
				}
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			config := &scraper.ScraperConfig{
				UserID:       "test",
				Password:     "pass",
				DownloadPath: "./test_downloads",
				TestMode:     true,
			}
			if tt.name == "with SlowMo option" {
				config.SlowMo = 100
			}
			defer os.RemoveAll(config.DownloadPath)

			logger := log.New(os.Stdout, "[TEST] ", log.LstdFlags)
			mockFactory := tt.setupMock()

			scraperTestable, err := scraper.NewETCScraperWithFactory(config, logger, mockFactory)
			if err != nil {
				t.Fatalf("Failed to create scraper: %v", err)
			}

			// Execute
			err = scraperTestable.Initialize()

			// Verify
			if (err != nil) != tt.expectError {
				t.Errorf("Initialize() error = %v, expectError %v", err, tt.expectError)
			}
			if err != nil && tt.errorContains != "" {
				if !contains(err.Error(), tt.errorContains) {
					t.Errorf("Error should contain '%s', got '%s'", tt.errorContains, err.Error())
				}
			}
		})
	}
}

func TestETCScraper_Login(t *testing.T) {
	tests := []struct {
		name          string
		setupMock     func() (*mocks.MockPage, *mocks.MockPlaywrightFactory)
		expectError   bool
		errorContains string
	}{
		{
			name: "successful login",
			setupMock: func() (*mocks.MockPage, *mocks.MockPlaywrightFactory) {
				mockPage := mocks.NewMockPage()

				// Setup locators
				mockPage.Locators["input[name='userId']"] = &mocks.MockLocator{CountValue: 1}
				mockPage.Locators["input[name='password']"] = &mocks.MockLocator{CountValue: 1}
				mockPage.Locators["button[type='submit']"] = &mocks.MockLocator{CountValue: 1}
				mockPage.Locators["a:has-text('ログアウト')"] = &mocks.MockLocator{CountValue: 1}

				factory := createMockFactory(mockPage)
				return mockPage, factory
			},
			expectError: false,
		},
		{
			name: "page not initialized",
			setupMock: func() (*mocks.MockPage, *mocks.MockPlaywrightFactory) {
				return nil, nil
			},
			expectError:   true,
			errorContains: "scraper not initialized",
		},
		{
			name: "navigation error",
			setupMock: func() (*mocks.MockPage, *mocks.MockPlaywrightFactory) {
				mockPage := mocks.NewMockPage()
				mockPage.GotoError = errors.New("navigation failed")
				factory := createMockFactory(mockPage)
				return mockPage, factory
			},
			expectError:   true,
			errorContains: "failed to navigate to login page",
		},
		{
			name: "user ID field not found",
			setupMock: func() (*mocks.MockPage, *mocks.MockPlaywrightFactory) {
				mockPage := mocks.NewMockPage()
				// All user ID selectors return 0 count
				factory := createMockFactory(mockPage)
				return mockPage, factory
			},
			expectError:   true,
			errorContains: "login form user ID field not found",
		},
		{
			name: "password field not found",
			setupMock: func() (*mocks.MockPage, *mocks.MockPlaywrightFactory) {
				mockPage := mocks.NewMockPage()
				mockPage.Locators["input[name='userId']"] = &mocks.MockLocator{CountValue: 1}
				// Password field not found
				factory := createMockFactory(mockPage)
				return mockPage, factory
			},
			expectError:   true,
			errorContains: "password field not found",
		},
		{
			name: "fill user ID error",
			setupMock: func() (*mocks.MockPage, *mocks.MockPlaywrightFactory) {
				mockPage := mocks.NewMockPage()
				mockPage.Locators["input[name='userId']"] = &mocks.MockLocator{
					CountValue: 1,
					FillError:  errors.New("fill failed"),
				}
				factory := createMockFactory(mockPage)
				return mockPage, factory
			},
			expectError:   true,
			errorContains: "failed to fill user ID",
		},
		{
			name: "fill password error",
			setupMock: func() (*mocks.MockPage, *mocks.MockPlaywrightFactory) {
				mockPage := mocks.NewMockPage()
				mockPage.Locators["input[name='userId']"] = &mocks.MockLocator{CountValue: 1}
				mockPage.Locators["input[name='password']"] = &mocks.MockLocator{
					CountValue: 1,
					FillError:  errors.New("fill failed"),
				}
				factory := createMockFactory(mockPage)
				return mockPage, factory
			},
			expectError:   true,
			errorContains: "failed to fill password",
		},
		{
			name: "login button not found",
			setupMock: func() (*mocks.MockPage, *mocks.MockPlaywrightFactory) {
				mockPage := mocks.NewMockPage()
				mockPage.Locators["input[name='userId']"] = &mocks.MockLocator{CountValue: 1}
				mockPage.Locators["input[name='password']"] = &mocks.MockLocator{CountValue: 1}
				// Login button not found
				factory := createMockFactory(mockPage)
				return mockPage, factory
			},
			expectError:   true,
			errorContains: "login button not found",
		},
		{
			name: "click login button error",
			setupMock: func() (*mocks.MockPage, *mocks.MockPlaywrightFactory) {
				mockPage := mocks.NewMockPage()
				mockPage.Locators["input[name='userId']"] = &mocks.MockLocator{CountValue: 1}
				mockPage.Locators["input[name='password']"] = &mocks.MockLocator{CountValue: 1}
				mockPage.Locators["button[type='submit']"] = &mocks.MockLocator{
					CountValue: 1,
					ClickError: errors.New("click failed"),
				}
				factory := createMockFactory(mockPage)
				return mockPage, factory
			},
			expectError:   true,
			errorContains: "failed to click login button",
		},
		{
			name: "wait for load state error",
			setupMock: func() (*mocks.MockPage, *mocks.MockPlaywrightFactory) {
				mockPage := mocks.NewMockPage()
				mockPage.Locators["input[name='userId']"] = &mocks.MockLocator{CountValue: 1}
				mockPage.Locators["input[name='password']"] = &mocks.MockLocator{CountValue: 1}
				mockPage.Locators["button[type='submit']"] = &mocks.MockLocator{CountValue: 1}
				mockPage.WaitError = errors.New("wait failed")
				factory := createMockFactory(mockPage)
				return mockPage, factory
			},
			expectError:   true,
			errorContains: "failed to wait for login completion",
		},
		{
			name: "login failed with error message",
			setupMock: func() (*mocks.MockPage, *mocks.MockPlaywrightFactory) {
				mockPage := mocks.NewMockPage()
				mockPage.Locators["input[name='userId']"] = &mocks.MockLocator{CountValue: 1}
				mockPage.Locators["input[name='password']"] = &mocks.MockLocator{CountValue: 1}
				mockPage.Locators["button[type='submit']"] = &mocks.MockLocator{CountValue: 1}
				mockPage.Locators["a:has-text('ログアウト')"] = &mocks.MockLocator{CountValue: 0}
				mockPage.Locators[".error-message, .alert-danger, .error"] = &mocks.MockLocator{
					CountValue: 1,
					TextValue:  "Invalid credentials",
				}
				factory := createMockFactory(mockPage)
				return mockPage, factory
			},
			expectError:   true,
			errorContains: "login failed: Invalid credentials",
		},
		{
			name: "login completed without logout button",
			setupMock: func() (*mocks.MockPage, *mocks.MockPlaywrightFactory) {
				mockPage := mocks.NewMockPage()
				mockPage.Locators["input[name='userId']"] = &mocks.MockLocator{CountValue: 1}
				mockPage.Locators["input[name='password']"] = &mocks.MockLocator{CountValue: 1}
				mockPage.Locators["button[type='submit']"] = &mocks.MockLocator{CountValue: 1}
				mockPage.Locators["a:has-text('ログアウト')"] = &mocks.MockLocator{CountValue: 0}
				mockPage.Locators[".error-message, .alert-danger, .error"] = &mocks.MockLocator{CountValue: 0}
				factory := createMockFactory(mockPage)
				return mockPage, factory
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.name == "page not initialized" {
				// Special case: test without initialization
				config := &scraper.ScraperConfig{
					UserID:   "test",
					Password: "pass",
				}
				logger := log.New(os.Stdout, "[TEST] ", log.LstdFlags)
				// Use a valid factory but don't initialize
				mockFactory := &mocks.MockPlaywrightFactory{}
				scraperTestable, err := scraper.NewETCScraperWithFactory(config, logger, mockFactory)
				if err != nil {
					t.Fatalf("Failed to create scraper: %v", err)
				}

				// Don't call Initialize, so page will be nil
				err = scraperTestable.Login()
				if err == nil || !contains(err.Error(), "scraper not initialized") {
					t.Errorf("Expected 'scraper not initialized' error, got %v", err)
				}
				return
			}

			// Normal test cases
			config := &scraper.ScraperConfig{
				UserID:       "test",
				Password:     "pass",
				DownloadPath: "./test_downloads",
				TestMode:     true,
			}
			defer os.RemoveAll(config.DownloadPath)

			logger := log.New(os.Stdout, "[TEST] ", log.LstdFlags)
			mockPage, mockFactory := tt.setupMock()

			scraperTestable, err := scraper.NewETCScraperWithFactory(config, logger, mockFactory)
			if err != nil {
				t.Fatalf("Failed to create scraper: %v", err)
			}

			if mockPage != nil {
				err = scraperTestable.Initialize()
				if err != nil {
					t.Fatalf("Failed to initialize: %v", err)
				}
			}

			// Execute
			err = scraperTestable.Login()

			// Verify
			if (err != nil) != tt.expectError {
				t.Errorf("Login() error = %v, expectError %v", err, tt.expectError)
			}
			if err != nil && tt.errorContains != "" {
				if !contains(err.Error(), tt.errorContains) {
					t.Errorf("Error should contain '%s', got '%s'", tt.errorContains, err.Error())
				}
			}
		})
	}
}

func TestETCScraper_DownloadMeisai(t *testing.T) {
	tests := []struct {
		name          string
		setupMock     func() (*mocks.MockPage, *mocks.MockPlaywrightFactory)
		fromDate      string
		toDate        string
		expectError   bool
		errorContains string
	}{
		{
			name: "successful download",
			setupMock: func() (*mocks.MockPage, *mocks.MockPlaywrightFactory) {
				mockPage := mocks.NewMockPage()

				// Setup successful navigation
				navigateCount := 0
				mockPage.GotoFunc = func(url string, options scraper.PageGotoOptions) (scraper.Response, error) {
					navigateCount++
					if navigateCount == 1 && url == "https://www.etc-meisai.jp/search" {
						return nil, nil
					}
					return nil, errors.New("not found")
				}

				// Setup date fields
				mockPage.Locators["input[name='fromDate']"] = &mocks.MockLocator{CountValue: 1}
				mockPage.Locators["input[name='toDate']"] = &mocks.MockLocator{CountValue: 1}
				mockPage.Locators["button:has-text('検索')"] = &mocks.MockLocator{CountValue: 1}
				mockPage.Locators["button:has-text('CSV')"] = &mocks.MockLocator{CountValue: 1}

				// Setup download handler
				mockPage.OnFunc = func(event string, handler interface{}) {
					if event == "download" {
						// Simulate download
						go func() {
							if downloadHandler, ok := handler.(func(scraper.Download)); ok {
								mockDownload := &MockPlaywrightDownload{
									suggestedName: "meisai.csv",
								}
								downloadHandler(mockDownload)
							}
						}()
					}
				}

				factory := createMockFactory(mockPage)
				return mockPage, factory
			},
			fromDate:    "2024-01-01",
			toDate:      "2024-01-31",
			expectError: false,
		},
		{
			name: "page not initialized",
			setupMock: func() (*mocks.MockPage, *mocks.MockPlaywrightFactory) {
				return nil, nil
			},
			fromDate:      "2024-01-01",
			toDate:        "2024-01-31",
			expectError:   true,
			errorContains: "scraper not initialized",
		},
		{
			name: "navigation to all URLs failed",
			setupMock: func() (*mocks.MockPage, *mocks.MockPlaywrightFactory) {
				mockPage := mocks.NewMockPage()
				mockPage.GotoError = errors.New("navigation failed")

				// Setup fields even though navigation failed
				mockPage.Locators["input[name='fromDate']"] = &mocks.MockLocator{CountValue: 1}
				mockPage.Locators["input[name='toDate']"] = &mocks.MockLocator{CountValue: 1}
				mockPage.Locators["button:has-text('検索')"] = &mocks.MockLocator{CountValue: 1}
				mockPage.Locators["button:has-text('CSV')"] = &mocks.MockLocator{CountValue: 1}

				// Setup download handler to complete successfully even if navigation failed
				mockPage.OnFunc = func(event string, handler interface{}) {
					if event == "download" {
						go func() {
							if downloadHandler, ok := handler.(func(scraper.Download)); ok {
								mockDownload := &MockPlaywrightDownload{
									suggestedName: "meisai.csv",
								}
								downloadHandler(mockDownload)
							}
						}()
					}
				}

				factory := createMockFactory(mockPage)
				return mockPage, factory
			},
			fromDate:    "2024-01-01",
			toDate:      "2024-01-31",
			expectError: false, // Still continues even if navigation fails
		},
		{
			name: "download button not found",
			setupMock: func() (*mocks.MockPage, *mocks.MockPlaywrightFactory) {
				mockPage := mocks.NewMockPage()

				// Setup fields but no download button
				mockPage.Locators["input[name='fromDate']"] = &mocks.MockLocator{CountValue: 1}
				mockPage.Locators["input[name='toDate']"] = &mocks.MockLocator{CountValue: 1}
				mockPage.Locators["button:has-text('検索')"] = &mocks.MockLocator{CountValue: 1}
				// No download button

				factory := createMockFactory(mockPage)
				return mockPage, factory
			},
			fromDate:      "2024-01-01",
			toDate:        "2024-01-31",
			expectError:   true,
			errorContains: "could not find download button",
		},
		{
			name: "download timeout",
			setupMock: func() (*mocks.MockPage, *mocks.MockPlaywrightFactory) {
				mockPage := mocks.NewMockPage()

				// Setup fields
				mockPage.Locators["input[name='fromDate']"] = &mocks.MockLocator{CountValue: 1}
				mockPage.Locators["input[name='toDate']"] = &mocks.MockLocator{CountValue: 1}
				mockPage.Locators["button:has-text('検索')"] = &mocks.MockLocator{CountValue: 1}
				mockPage.Locators["button:has-text('CSV')"] = &mocks.MockLocator{CountValue: 1}

				// Don't trigger download handler

				factory := createMockFactory(mockPage)
				return mockPage, factory
			},
			fromDate:      "2024-01-01",
			toDate:        "2024-01-31",
			expectError:   true,
			errorContains: "download timeout",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.name == "page not initialized" {
				// Special case: test without initialization
				config := &scraper.ScraperConfig{
					UserID:   "test",
					Password: "pass",
				}
				logger := log.New(os.Stdout, "[TEST] ", log.LstdFlags)
				// Use a valid factory but don't initialize
				mockFactory := &mocks.MockPlaywrightFactory{}
				scraperTestable, err := scraper.NewETCScraperWithFactory(config, logger, mockFactory)
				if err != nil {
					t.Fatalf("Failed to create scraper: %v", err)
				}

				// Don't call Initialize, so page will be nil
				_, err = scraperTestable.DownloadMeisai(tt.fromDate, tt.toDate)
				if err == nil || !contains(err.Error(), "scraper not initialized") {
					t.Errorf("Expected 'scraper not initialized' error, got %v", err)
				}
				return
			}

			// Normal test cases
			config := &scraper.ScraperConfig{
				UserID:       "test",
				Password:     "pass",
				DownloadPath: "./test_downloads",
				TestMode:     true,
				Timeout:      1000, // Short timeout for test
			}
			defer os.RemoveAll(config.DownloadPath)

			logger := log.New(os.Stdout, "[TEST] ", log.LstdFlags)
			mockPage, mockFactory := tt.setupMock()

			scraperTestable, err := scraper.NewETCScraperWithFactory(config, logger, mockFactory)
			if err != nil {
				t.Fatalf("Failed to create scraper: %v", err)
			}

			if mockPage != nil {
				err = scraperTestable.Initialize()
				if err != nil {
					t.Fatalf("Failed to initialize: %v", err)
				}
			}

			// Execute
			path, err := scraperTestable.DownloadMeisai(tt.fromDate, tt.toDate)

			// Verify
			if (err != nil) != tt.expectError {
				t.Errorf("DownloadMeisai() error = %v, expectError %v", err, tt.expectError)
			}
			if err != nil && tt.errorContains != "" {
				if !contains(err.Error(), tt.errorContains) {
					t.Errorf("Error should contain '%s', got '%s'", tt.errorContains, err.Error())
				}
			}
			if err == nil && path == "" && !tt.expectError {
				t.Error("Expected download path, got empty string")
			}
		})
	}
}

func TestETCScraper_handleDownload(t *testing.T) {
	tests := []struct {
		name           string
		setupDownload  func() *MockPlaywrightDownload
		expectComplete bool
	}{
		{
			name: "successful download",
			setupDownload: func() *MockPlaywrightDownload {
				return &MockPlaywrightDownload{
					suggestedName: "test.csv",
					saveError:     nil,
				}
			},
			expectComplete: true,
		},
		{
			name: "download save error",
			setupDownload: func() *MockPlaywrightDownload {
				return &MockPlaywrightDownload{
					suggestedName: "test.csv",
					saveError:     errors.New("permission denied"),
				}
			},
			expectComplete: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &scraper.ScraperConfig{
				UserID:       "test",
				Password:     "pass",
				DownloadPath: "./test_downloads",
				TestMode:     true,
			}
			logger := log.New(os.Stdout, "[TEST] ", log.LstdFlags)

			// Create scraper with mock factory
			mockFactory := &mocks.MockPlaywrightFactory{}
			scraper, err := scraper.NewETCScraperWithFactory(config, logger, mockFactory)
			if err != nil {
				t.Fatalf("Failed to create scraper: %v", err)
			}

			// Setup download channel
			downloadComplete := make(chan string, 1)
			mockDownload := tt.setupDownload()

			// Execute handleDownload
			scraper.HandleDownload(mockDownload, downloadComplete)

			// Check result
			select {
			case path := <-downloadComplete:
				if !tt.expectComplete {
					t.Errorf("Expected no completion but got path: %s", path)
				}
			case <-time.After(100 * time.Millisecond):
				if tt.expectComplete {
					t.Error("Expected download completion but got timeout")
				}
			}
		})
	}
}

func TestETCScraper_Close(t *testing.T) {
	tests := []struct {
		name      string
		setupMock func() *mocks.MockPlaywrightFactory
		hasError  bool
	}{
		{
			name: "successful close",
			setupMock: func() *mocks.MockPlaywrightFactory {
				mockPage := mocks.NewMockPage()
				mockContext := &mocks.MockBrowserContext{
					NewPageFunc: func() (scraper.PageInterface, error) {
						return mockPage, nil
					},
				}
				mockBrowser := &mocks.MockBrowser{
					NewContextFunc: func(options scraper.BrowserNewContextOptions) (scraper.BrowserContextInterface, error) {
						return mockContext, nil
					},
				}
				mockChromium := &mocks.MockBrowserType{
					LaunchFunc: func(options scraper.BrowserTypeLaunchOptions) (scraper.BrowserInterface, error) {
						return mockBrowser, nil
					},
				}
				mockPw := &mocks.MockPlaywright{
					Chromium: mockChromium,
				}
				return &mocks.MockPlaywrightFactory{
					RunFunc: func() (scraper.PlaywrightInterface, error) {
						return mockPw, nil
					},
				}
			},
			hasError: false,
		},
		{
			name: "close with errors",
			setupMock: func() *mocks.MockPlaywrightFactory {
				mockPage := &mocks.MockPage{
					CloseError: errors.New("page close failed"),
				}
				mockContext := &mocks.MockBrowserContext{
					NewPageFunc: func() (scraper.PageInterface, error) {
						return mockPage, nil
					},
					CloseError: errors.New("context close failed"),
				}
				mockBrowser := &mocks.MockBrowser{
					NewContextFunc: func(options scraper.BrowserNewContextOptions) (scraper.BrowserContextInterface, error) {
						return mockContext, nil
					},
					CloseError: errors.New("browser close failed"),
				}
				mockChromium := &mocks.MockBrowserType{
					LaunchFunc: func(options scraper.BrowserTypeLaunchOptions) (scraper.BrowserInterface, error) {
						return mockBrowser, nil
					},
				}
				mockPw := &mocks.MockPlaywright{
					Chromium:  mockChromium,
					StopError: errors.New("playwright stop failed"),
				}
				return &mocks.MockPlaywrightFactory{
					RunFunc: func() (scraper.PlaywrightInterface, error) {
						return mockPw, nil
					},
				}
			},
			hasError: false, // Close method always returns nil, ignoring errors
		},
		{
			name: "close without initialization",
			setupMock: func() *mocks.MockPlaywrightFactory {
				return &mocks.MockPlaywrightFactory{}
			},
			hasError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &scraper.ScraperConfig{
				UserID:       "test",
				Password:     "pass",
				DownloadPath: "./test_downloads",
				TestMode:     true,
			}
			defer os.RemoveAll(config.DownloadPath)

			logger := log.New(os.Stdout, "[TEST] ", log.LstdFlags)
			mockFactory := tt.setupMock()

			scraperTestable, err := scraper.NewETCScraperWithFactory(config, logger, mockFactory)
			if err != nil {
				t.Fatalf("Failed to create scraper: %v", err)
			}

			if tt.name != "close without initialization" {
				err = scraperTestable.Initialize()
				if err != nil {
					t.Fatalf("Failed to initialize: %v", err)
				}
			}

			// Execute
			err = scraperTestable.Close()

			// Verify
			if (err != nil) != tt.hasError {
				t.Errorf("Close() error = %v, hasError %v", err, tt.hasError)
			}
		})
	}
}

// Helper functions

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && len(substr) > 0 &&
		(s[0:len(substr)] == substr || (len(s) > len(substr) && contains(s[1:], substr))))
}

func createMockFactory(mockPage *mocks.MockPage) *mocks.MockPlaywrightFactory {
	mockContext := &mocks.MockBrowserContext{
		NewPageFunc: func() (scraper.PageInterface, error) {
			return mockPage, nil
		},
	}
	mockBrowser := &mocks.MockBrowser{
		NewContextFunc: func(options scraper.BrowserNewContextOptions) (scraper.BrowserContextInterface, error) {
			return mockContext, nil
		},
	}
	mockChromium := &mocks.MockBrowserType{
		LaunchFunc: func(options scraper.BrowserTypeLaunchOptions) (scraper.BrowserInterface, error) {
			return mockBrowser, nil
		},
	}
	mockPw := &mocks.MockPlaywright{
		Chromium: mockChromium,
	}
	return &mocks.MockPlaywrightFactory{
		RunFunc: func() (scraper.PlaywrightInterface, error) {
			return mockPw, nil
		},
	}
}

// MockPlaywrightDownload implements scraper.Download interface
type MockPlaywrightDownload struct {
	suggestedName string
	saveError     error
}

func (m *MockPlaywrightDownload) Cancel() error {
	return nil
}

func (m *MockPlaywrightDownload) Delete() error {
	return nil
}

func (m *MockPlaywrightDownload) Failure() error {
	return nil
}

func (m *MockPlaywrightDownload) Page() scraper.PageInterface {
	return nil
}

// TestETCScraper_DownloadMeisaiToBuffer tests the DownloadMeisaiToBuffer function
func TestETCScraper_DownloadMeisaiToBuffer(t *testing.T) {
	tests := []struct {
		name           string
		setupMock      func() *mocks.MockPlaywrightFactory
		fromDate       string
		toDate         string
		expectedData   string
		expectError    bool
		errorContains  string
	}{
		{
			name: "successful download and buffer",
			setupMock: func() *mocks.MockPlaywrightFactory {
				mockPage := mocks.NewMockPage()
				mockPage.GotoFunc = func(url string, options scraper.PageGotoOptions) (scraper.Response, error) {
					return nil, nil
				}

				// Setup download handler to simulate file creation
				mockPage.OnFunc = func(event string, handler interface{}) {
					if event == "download" {
						go func() {
							if downloadHandler, ok := handler.(func(scraper.Download)); ok {
								mockDownload := &mocks.MockDownload{
									SuggestedName: "test.csv",
									SaveError:     nil,
								}
								downloadHandler(mockDownload)
							}
						}()
					}
				}

				// Setup locators for successful flow
				mockPage.Locators = map[string]*mocks.MockLocator{
					"input[name='fromDate']":    {CountValue: 1},
					"input[name='toDate']":      {CountValue: 1},
					"button:has-text('検索')":      {CountValue: 1},
					"button:has-text('CSV')":    {CountValue: 1},
				}

				mockContext := &mocks.MockBrowserContext{
					NewPageFunc: func() (scraper.PageInterface, error) {
						return mockPage, nil
					},
				}
				mockBrowser := &mocks.MockBrowser{
					NewContextFunc: func(options scraper.BrowserNewContextOptions) (scraper.BrowserContextInterface, error) {
						return mockContext, nil
					},
				}
				mockChromium := &mocks.MockBrowserType{
					LaunchFunc: func(options scraper.BrowserTypeLaunchOptions) (scraper.BrowserInterface, error) {
						return mockBrowser, nil
					},
				}
				mockPw := &mocks.MockPlaywright{
					Chromium: mockChromium,
				}
				return &mocks.MockPlaywrightFactory{
					RunFunc: func() (scraper.PlaywrightInterface, error) {
						return mockPw, nil
					},
				}
			},
			fromDate:     "2024-01-01",
			toDate:       "2024-01-31",
			expectedData: "test csv data",
			expectError:  false,
		},
		{
			name: "download error",
			setupMock: func() *mocks.MockPlaywrightFactory {
				mockPage := mocks.NewMockPage()
				mockPage.GotoFunc = func(url string, options scraper.PageGotoOptions) (scraper.Response, error) {
					return nil, fmt.Errorf("navigation failed")
				}

				mockContext := &mocks.MockBrowserContext{
					NewPageFunc: func() (scraper.PageInterface, error) {
						return mockPage, nil
					},
				}
				mockBrowser := &mocks.MockBrowser{
					NewContextFunc: func(options scraper.BrowserNewContextOptions) (scraper.BrowserContextInterface, error) {
						return mockContext, nil
					},
				}
				mockChromium := &mocks.MockBrowserType{
					LaunchFunc: func(options scraper.BrowserTypeLaunchOptions) (scraper.BrowserInterface, error) {
						return mockBrowser, nil
					},
				}
				mockPw := &mocks.MockPlaywright{
					Chromium: mockChromium,
				}
				return &mocks.MockPlaywrightFactory{
					RunFunc: func() (scraper.PlaywrightInterface, error) {
						return mockPw, nil
					},
				}
			},
			fromDate:      "2024-01-01",
			toDate:        "2024-01-31",
			expectError:   true,
			errorContains: "failed to download CSV",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary file for successful test
			if tt.expectedData != "" {
				tmpFile := filepath.Join("./test_downloads", "test.csv")
				os.MkdirAll("./test_downloads", 0755)
				err := os.WriteFile(tmpFile, []byte(tt.expectedData), 0644)
				if err != nil {
					t.Fatalf("Failed to create test file: %v", err)
				}
				defer os.Remove(tmpFile)
			}

			config := &scraper.ScraperConfig{
				UserID:       "test",
				Password:     "test",
				DownloadPath: "./test_downloads",
				Headless:     true,
				TestMode:     true,
			}

			factory := tt.setupMock()
			logger := log.New(os.Stdout, "[TEST] ", log.LstdFlags)
			scraperInstance, err := scraper.NewETCScraperWithFactory(config, logger, factory)
			if err != nil {
				t.Fatalf("Failed to create scraper: %v", err)
			}

			err = scraperInstance.Initialize()
			if err != nil {
				t.Fatalf("Failed to initialize scraper: %v", err)
			}

			data, err := scraperInstance.DownloadMeisaiToBuffer(tt.fromDate, tt.toDate)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				} else if tt.errorContains != "" && !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("Expected error to contain %q, got %q", tt.errorContains, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if tt.expectedData != "" && string(data) != tt.expectedData {
					t.Errorf("Expected data %q, got %q", tt.expectedData, string(data))
				}
			}
		})
	}
}

func (m *MockPlaywrightDownload) Path() (string, error) {
	return filepath.Join("./test_downloads", m.suggestedName), nil
}

func (m *MockPlaywrightDownload) SaveAs(path string) error {
	return m.saveError
}

func (m *MockPlaywrightDownload) SuggestedFilename() string {
	return m.suggestedName
}

func (m *MockPlaywrightDownload) URL() string {
	return "https://example.com/" + m.suggestedName
}

func (m *MockPlaywrightDownload) String() string {
	return "MockPlaywrightDownload{" + m.suggestedName + "}"
}
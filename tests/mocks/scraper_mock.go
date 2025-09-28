package mocks

import (
	"fmt"
	"log"

	"github.com/yhonda-ohishi/etc_meisai/src/scraper"
)

// MockETCScraper is a mock implementation of ScraperInterface
type MockETCScraper struct {
	InitializeCalled   bool
	InitializeError    error
	LoginCalled        bool
	LoginError         error
	DownloadCalled     bool
	DownloadError      error
	DownloadResult     string
	CloseCalled        bool
	CloseError         error
	FromDate           string
	ToDate             string
}

// NewMockETCScraper creates a new mock scraper
func NewMockETCScraper() *MockETCScraper {
	return &MockETCScraper{
		DownloadResult: "/mock/downloads/test.csv",
	}
}

// Initialize mocks the Initialize method
func (m *MockETCScraper) Initialize() error {
	m.InitializeCalled = true
	return m.InitializeError
}

// Login mocks the Login method
func (m *MockETCScraper) Login() error {
	m.LoginCalled = true
	return m.LoginError
}

// DownloadMeisai mocks the DownloadMeisai method
func (m *MockETCScraper) DownloadMeisai(fromDate, toDate string) (string, error) {
	m.DownloadCalled = true
	m.FromDate = fromDate
	m.ToDate = toDate

	if m.DownloadError != nil {
		return "", m.DownloadError
	}

	return m.DownloadResult, nil
}

// Close mocks the Close method
func (m *MockETCScraper) Close() error {
	m.CloseCalled = true
	return m.CloseError
}

// Verify interface compliance
var _ scraper.ScraperInterface = (*MockETCScraper)(nil)

// MockScraperFactory creates mock scrapers for testing
type MockScraperFactory struct {
	CreateFunc func(config *scraper.ScraperConfig, logger *log.Logger) (scraper.ScraperInterface, error)
	CreateErr  error
	MockScraper *MockETCScraper
}

// NewMockScraperFactory creates a new mock scraper factory
func NewMockScraperFactory() *MockScraperFactory {
	mockScraper := NewMockETCScraper()
	return &MockScraperFactory{
		MockScraper: mockScraper,
		CreateFunc: func(config *scraper.ScraperConfig, logger *log.Logger) (scraper.ScraperInterface, error) {
			return mockScraper, nil
		},
	}
}

// CreateScraper creates a mock scraper
func (f *MockScraperFactory) CreateScraper(config *scraper.ScraperConfig, logger *log.Logger) (scraper.ScraperInterface, error) {
	if f.CreateErr != nil {
		return nil, f.CreateErr
	}

	if f.CreateFunc != nil {
		return f.CreateFunc(config, logger)
	}

	return f.MockScraper, nil
}

// ConfigurableETCScraper allows configuring mock behavior per test
type ConfigurableETCScraper struct {
	InitializeFunc func() error
	LoginFunc      func() error
	DownloadFunc   func(fromDate, toDate string) (string, error)
	CloseFunc      func() error
}

// NewConfigurableETCScraper creates a new configurable mock
func NewConfigurableETCScraper() *ConfigurableETCScraper {
	return &ConfigurableETCScraper{
		InitializeFunc: func() error { return nil },
		LoginFunc:      func() error { return nil },
		DownloadFunc:   func(fromDate, toDate string) (string, error) {
			return fmt.Sprintf("/downloads/mock_%s_%s.csv", fromDate, toDate), nil
		},
		CloseFunc:      func() error { return nil },
	}
}

// Initialize calls the configured function
func (c *ConfigurableETCScraper) Initialize() error {
	if c.InitializeFunc != nil {
		return c.InitializeFunc()
	}
	return nil
}

// Login calls the configured function
func (c *ConfigurableETCScraper) Login() error {
	if c.LoginFunc != nil {
		return c.LoginFunc()
	}
	return nil
}

// DownloadMeisai calls the configured function
func (c *ConfigurableETCScraper) DownloadMeisai(fromDate, toDate string) (string, error) {
	if c.DownloadFunc != nil {
		return c.DownloadFunc(fromDate, toDate)
	}
	return "", nil
}

// Close calls the configured function
func (c *ConfigurableETCScraper) Close() error {
	if c.CloseFunc != nil {
		return c.CloseFunc()
	}
	return nil
}

// Verify interface compliance
var _ scraper.ScraperInterface = (*ConfigurableETCScraper)(nil)
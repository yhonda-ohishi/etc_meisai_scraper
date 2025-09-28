package scraper

import "github.com/playwright-community/playwright-go"

// PlaywrightInterface wraps playwright.Playwright for mocking
type PlaywrightInterface interface {
	Stop() error
	GetChromium() BrowserTypeInterface
}

// BrowserTypeInterface wraps playwright.BrowserType for mocking
type BrowserTypeInterface interface {
	Launch(options ...playwright.BrowserTypeLaunchOptions) (BrowserInterface, error)
}

// BrowserInterface wraps playwright.Browser for mocking
type BrowserInterface interface {
	NewContext(options ...playwright.BrowserNewContextOptions) (BrowserContextInterface, error)
	Close() error
}

// BrowserContextInterface wraps playwright.BrowserContext for mocking
type BrowserContextInterface interface {
	NewPage() (PageInterface, error)
	SetDefaultTimeout(timeout float64)
	Close() error
}

// PageInterface wraps playwright.Page for mocking
type PageInterface interface {
	Goto(url string, options ...playwright.PageGotoOptions) (playwright.Response, error)
	Locator(selector string) LocatorInterface
	WaitForLoadState(options ...playwright.PageWaitForLoadStateOptions) error
	Screenshot(options ...playwright.PageScreenshotOptions) ([]byte, error)
	Close() error
	On(event string, handler interface{})
}

// LocatorInterface wraps playwright.Locator for mocking
type LocatorInterface interface {
	Count() (int, error)
	First() LocatorInterface
	Fill(value string) error
	Click(options ...playwright.LocatorClickOptions) error
	TextContent(options ...playwright.LocatorTextContentOptions) (string, error)
}

// PlaywrightFactory creates playwright instances
type PlaywrightFactory interface {
	Run() (PlaywrightInterface, error)
	Install() error
}

// DefaultPlaywrightFactory is the production implementation
type DefaultPlaywrightFactory struct{}

func (f *DefaultPlaywrightFactory) Install() error {
	return playwright.Install()
}

func (f *DefaultPlaywrightFactory) Run() (PlaywrightInterface, error) {
	// TODO: Implement actual playwright adapter
	// For now, return nil as this is only used in production
	// and tests use mock implementations
	return nil, nil
}
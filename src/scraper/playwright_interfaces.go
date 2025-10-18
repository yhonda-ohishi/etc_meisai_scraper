package scraper

// Custom types to avoid playwright imports in interfaces

// Size represents viewport dimensions
type Size struct {
	Width  int
	Height int
}

// BrowserTypeLaunchOptions represents browser launch options
type BrowserTypeLaunchOptions struct {
	Headless *bool
	SlowMo   *float64
}

// BrowserNewContextOptions represents browser context options
type BrowserNewContextOptions struct {
	AcceptDownloads *bool
	Viewport        *Size
	UserAgent       *string
}

// PageGotoOptions represents page navigation options
type PageGotoOptions struct {
	WaitUntil WaitUntilState
}

// PageWaitForLoadStateOptions represents load state wait options
type PageWaitForLoadStateOptions struct {
	State LoadState
}

// PageScreenshotOptions represents screenshot options
type PageScreenshotOptions struct {
	Path string
}

// LocatorClickOptions represents click options
type LocatorClickOptions struct{}

// LocatorTextContentOptions represents text content options
type LocatorTextContentOptions struct{}

// WaitUntilState represents wait until states
type WaitUntilState string

const (
	WaitUntilStateNetworkidle WaitUntilState = "networkidle"
)

// LoadState represents load states
type LoadState string

const (
	LoadStateNetworkidle LoadState = "networkidle"
)

// Response represents a mock response
type Response interface{}

// Download interface for downloads
type Download interface {
	SuggestedFilename() string
	SaveAs(path string) error
}

// PlaywrightInterface wraps playwright.Playwright for mocking
type PlaywrightInterface interface {
	Stop() error
	GetChromium() BrowserTypeInterface
}

// BrowserTypeInterface wraps playwright.BrowserType for mocking
type BrowserTypeInterface interface {
	Launch(options BrowserTypeLaunchOptions) (BrowserInterface, error)
}

// BrowserInterface wraps playwright.Browser for mocking
type BrowserInterface interface {
	NewContext(options BrowserNewContextOptions) (BrowserContextInterface, error)
	Close() error
}

// BrowserContextInterface wraps playwright.BrowserContext for mocking
type BrowserContextInterface interface {
	NewPage() (PageInterface, error)
	SetDefaultTimeout(timeout float64)
	Close() error
	On(event string, handler interface{})
}

// PageInterface wraps playwright.Page for mocking
type PageInterface interface {
	Goto(url string, options PageGotoOptions) (Response, error)
	Locator(selector string) LocatorInterface
	WaitForLoadState(options PageWaitForLoadStateOptions) error
	Screenshot(options PageScreenshotOptions) ([]byte, error)
	Close() error
	On(event string, handler interface{})
	Evaluate(expression string, arg ...interface{}) (interface{}, error)
}

// LocatorInterface wraps playwright.Locator for mocking
type LocatorInterface interface {
	Count() (int, error)
	First() LocatorInterface
	Fill(value string) error
	Click(options LocatorClickOptions) error
	TextContent(options LocatorTextContentOptions) (string, error)
}

// Helper functions for creating option structs
func Bool(b bool) *bool {
	return &b
}

func Float(f float64) *float64 {
	return &f
}

func String(s string) *string {
	return &s
}

// PlaywrightFactory creates playwright instances
type PlaywrightFactory interface {
	Run() (PlaywrightInterface, error)
	Install() error
}

// DefaultPlaywrightFactory is the production implementation
type DefaultPlaywrightFactory struct{}

func (f *DefaultPlaywrightFactory) Install() error {
	// Playwright install is handled separately by the user
	// or CI/CD pipeline, so we just return nil here
	return nil
}

func (f *DefaultPlaywrightFactory) Run() (PlaywrightInterface, error) {
	// For production use, we need a real playwright implementation
	// This requires creating an adapter
	return NewRealPlaywright()
}
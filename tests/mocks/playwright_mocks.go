package mocks

import (
	"github.com/yhonda-ohishi/etc_meisai/src/scraper"
)

// MockPlaywrightFactory implements scraper.PlaywrightFactory
type MockPlaywrightFactory struct {
	RunFunc     func() (scraper.PlaywrightInterface, error)
	InstallFunc func() error
	RunError    error
	InstallError error
}

func (m *MockPlaywrightFactory) Run() (scraper.PlaywrightInterface, error) {
	if m.RunFunc != nil {
		return m.RunFunc()
	}
	if m.RunError != nil {
		return nil, m.RunError
	}
	return &MockPlaywright{}, nil
}

func (m *MockPlaywrightFactory) Install() error {
	if m.InstallFunc != nil {
		return m.InstallFunc()
	}
	return m.InstallError
}

// MockPlaywright implements scraper.PlaywrightInterface
type MockPlaywright struct {
	StopFunc func() error
	StopError error
	Chromium *MockBrowserType
}

func (m *MockPlaywright) Stop() error {
	if m.StopFunc != nil {
		return m.StopFunc()
	}
	return m.StopError
}

func (m *MockPlaywright) GetChromium() scraper.BrowserTypeInterface {
	if m.Chromium != nil {
		return m.Chromium
	}
	return &MockBrowserType{}
}

// MockBrowserType implements scraper.BrowserTypeInterface
type MockBrowserType struct {
	LaunchFunc func(options scraper.BrowserTypeLaunchOptions) (scraper.BrowserInterface, error)
	LaunchError error
}

func (m *MockBrowserType) Launch(options scraper.BrowserTypeLaunchOptions) (scraper.BrowserInterface, error) {
	if m.LaunchFunc != nil {
		return m.LaunchFunc(options)
	}
	if m.LaunchError != nil {
		return nil, m.LaunchError
	}
	return &MockBrowser{}, nil
}

// MockBrowser implements scraper.BrowserInterface
type MockBrowser struct {
	NewContextFunc func(options scraper.BrowserNewContextOptions) (scraper.BrowserContextInterface, error)
	CloseFunc func() error
	NewContextError error
	CloseError error
}

func (m *MockBrowser) NewContext(options scraper.BrowserNewContextOptions) (scraper.BrowserContextInterface, error) {
	if m.NewContextFunc != nil {
		return m.NewContextFunc(options)
	}
	if m.NewContextError != nil {
		return nil, m.NewContextError
	}
	return &MockBrowserContext{}, nil
}

func (m *MockBrowser) Close() error {
	if m.CloseFunc != nil {
		return m.CloseFunc()
	}
	return m.CloseError
}

// MockBrowserContext implements scraper.BrowserContextInterface
type MockBrowserContext struct {
	NewPageFunc func() (scraper.PageInterface, error)
	SetDefaultTimeoutFunc func(timeout float64)
	CloseFunc func() error
	NewPageError error
	CloseError error
	TimeoutSet float64
}

func (m *MockBrowserContext) NewPage() (scraper.PageInterface, error) {
	if m.NewPageFunc != nil {
		return m.NewPageFunc()
	}
	if m.NewPageError != nil {
		return nil, m.NewPageError
	}
	return &MockPage{}, nil
}

func (m *MockBrowserContext) SetDefaultTimeout(timeout float64) {
	m.TimeoutSet = timeout
	if m.SetDefaultTimeoutFunc != nil {
		m.SetDefaultTimeoutFunc(timeout)
	}
}

func (m *MockBrowserContext) Close() error {
	if m.CloseFunc != nil {
		return m.CloseFunc()
	}
	return m.CloseError
}

// MockPage implements scraper.PageInterface
type MockPage struct {
	GotoFunc func(url string, options scraper.PageGotoOptions) (scraper.Response, error)
	LocatorFunc func(selector string) scraper.LocatorInterface
	WaitForLoadStateFunc func(options scraper.PageWaitForLoadStateOptions) error
	ScreenshotFunc func(options scraper.PageScreenshotOptions) ([]byte, error)
	CloseFunc func() error
	OnFunc func(event string, handler interface{})

	GotoError error
	WaitError error
	ScreenshotError error
	CloseError error
	Locators map[string]*MockLocator
	DownloadHandler interface{}
}

func NewMockPage() *MockPage {
	return &MockPage{
		Locators: make(map[string]*MockLocator),
	}
}

func (m *MockPage) Goto(url string, options scraper.PageGotoOptions) (scraper.Response, error) {
	if m.GotoFunc != nil {
		return m.GotoFunc(url, options)
	}
	if m.GotoError != nil {
		return nil, m.GotoError
	}
	return nil, nil
}

func (m *MockPage) Locator(selector string) scraper.LocatorInterface {
	if m.LocatorFunc != nil {
		return m.LocatorFunc(selector)
	}
	if locator, exists := m.Locators[selector]; exists {
		return locator
	}
	// Return empty locator by default
	return &MockLocator{CountValue: 0}
}

func (m *MockPage) WaitForLoadState(options scraper.PageWaitForLoadStateOptions) error {
	if m.WaitForLoadStateFunc != nil {
		return m.WaitForLoadStateFunc(options)
	}
	return m.WaitError
}

func (m *MockPage) Screenshot(options scraper.PageScreenshotOptions) ([]byte, error) {
	if m.ScreenshotFunc != nil {
		return m.ScreenshotFunc(options)
	}
	if m.ScreenshotError != nil {
		return nil, m.ScreenshotError
	}
	return []byte("mock screenshot"), nil
}

func (m *MockPage) Close() error {
	if m.CloseFunc != nil {
		return m.CloseFunc()
	}
	return m.CloseError
}

func (m *MockPage) On(event string, handler interface{}) {
	if m.OnFunc != nil {
		m.OnFunc(event, handler)
	}
	if event == "download" {
		m.DownloadHandler = handler
	}
}

// MockLocator implements scraper.LocatorInterface
type MockLocator struct {
	CountFunc func() (int, error)
	FirstFunc func() scraper.LocatorInterface
	FillFunc func(value string) error
	ClickFunc func(options scraper.LocatorClickOptions) error
	TextContentFunc func(options scraper.LocatorTextContentOptions) (string, error)

	CountValue int
	CountError error
	FillError error
	ClickError error
	TextValue string
	TextError error
}

func (m *MockLocator) Count() (int, error) {
	if m.CountFunc != nil {
		return m.CountFunc()
	}
	if m.CountError != nil {
		return 0, m.CountError
	}
	return m.CountValue, nil
}

func (m *MockLocator) First() scraper.LocatorInterface {
	if m.FirstFunc != nil {
		return m.FirstFunc()
	}
	return m
}

func (m *MockLocator) Fill(value string) error {
	if m.FillFunc != nil {
		return m.FillFunc(value)
	}
	return m.FillError
}

func (m *MockLocator) Click(options scraper.LocatorClickOptions) error {
	if m.ClickFunc != nil {
		return m.ClickFunc(options)
	}
	return m.ClickError
}

func (m *MockLocator) TextContent(options scraper.LocatorTextContentOptions) (string, error) {
	if m.TextContentFunc != nil {
		return m.TextContentFunc(options)
	}
	if m.TextError != nil {
		return "", m.TextError
	}
	return m.TextValue, nil
}

// MockDownload simulates a scraper.Download
type MockDownload struct {
	SuggestedName string
	SaveError error
}

func (m *MockDownload) SuggestedFilename() string {
	return m.SuggestedName
}

func (m *MockDownload) SaveAs(path string) error {
	return m.SaveError
}

// SetDownloadHandler sets up mock download handler for testing
func (m *MockPage) SetDownloadHandler(filePath string) {
	m.OnFunc = func(event string, handler interface{}) {
		if event == "download" {
			go func() {
				if downloadHandler, ok := handler.(func(scraper.Download)); ok {
					mockDownload := &MockDownload{
						SuggestedName: "test.csv",
						SaveError:     nil,
					}
					downloadHandler(mockDownload)
				}
			}()
		}
	}
}
package scraper

import (
	"github.com/playwright-community/playwright-go"
)

// RealPlaywright wraps the actual playwright instance
type RealPlaywright struct {
	pw *playwright.Playwright
}

// NewRealPlaywright creates a new real playwright instance
func NewRealPlaywright() (PlaywrightInterface, error) {
	pw, err := playwright.Run()
	if err != nil {
		return nil, err
	}
	return &RealPlaywright{pw: pw}, nil
}

func (r *RealPlaywright) Stop() error {
	if r.pw != nil {
		return r.pw.Stop()
	}
	return nil
}

func (r *RealPlaywright) GetChromium() BrowserTypeInterface {
	return &RealBrowserType{bt: r.pw.Chromium}
}

// RealBrowserType wraps playwright.BrowserType
type RealBrowserType struct {
	bt playwright.BrowserType
}

func (r *RealBrowserType) Launch(options BrowserTypeLaunchOptions) (BrowserInterface, error) {
	opts := playwright.BrowserTypeLaunchOptions{}
	if options.Headless != nil {
		opts.Headless = options.Headless
	}
	if options.SlowMo != nil {
		opts.SlowMo = options.SlowMo
	}

	browser, err := r.bt.Launch(opts)
	if err != nil {
		return nil, err
	}
	return &RealBrowser{browser: browser}, nil
}

// RealBrowser wraps playwright.Browser
type RealBrowser struct {
	browser playwright.Browser
}

func (r *RealBrowser) NewContext(options BrowserNewContextOptions) (BrowserContextInterface, error) {
	opts := playwright.BrowserNewContextOptions{}
	if options.AcceptDownloads != nil {
		opts.AcceptDownloads = options.AcceptDownloads
	}
	if options.Viewport != nil {
		opts.Viewport = &playwright.Size{
			Width:  options.Viewport.Width,
			Height: options.Viewport.Height,
		}
	}
	if options.UserAgent != nil {
		opts.UserAgent = options.UserAgent
	}

	ctx, err := r.browser.NewContext(opts)
	if err != nil {
		return nil, err
	}
	return &RealBrowserContext{context: ctx}, nil
}

func (r *RealBrowser) Close() error {
	return r.browser.Close()
}

// RealBrowserContext wraps playwright.BrowserContext
type RealBrowserContext struct {
	context playwright.BrowserContext
}

func (r *RealBrowserContext) NewPage() (PageInterface, error) {
	page, err := r.context.NewPage()
	if err != nil {
		return nil, err
	}
	return &RealPage{page: page}, nil
}

func (r *RealBrowserContext) SetDefaultTimeout(timeout float64) {
	r.context.SetDefaultTimeout(timeout)
}

func (r *RealBrowserContext) Close() error {
	return r.context.Close()
}

func (r *RealBrowserContext) On(event string, handler interface{}) {
	// BrowserContext event handling not supported in playwright-go
	// Downloads are handled at the page level
}

// RealPage wraps playwright.Page
type RealPage struct {
	page playwright.Page
}

func (r *RealPage) Goto(url string, options PageGotoOptions) (Response, error) {
	opts := playwright.PageGotoOptions{}
	if options.WaitUntil == WaitUntilStateNetworkidle {
		opts.WaitUntil = playwright.WaitUntilStateNetworkidle
	}
	return r.page.Goto(url, opts)
}

func (r *RealPage) Locator(selector string) LocatorInterface {
	return &RealLocator{locator: r.page.Locator(selector)}
}

func (r *RealPage) WaitForLoadState(options PageWaitForLoadStateOptions) error {
	opts := playwright.PageWaitForLoadStateOptions{}
	if options.State == LoadStateNetworkidle {
		opts.State = playwright.LoadStateNetworkidle
	}
	return r.page.WaitForLoadState(opts)
}

func (r *RealPage) Screenshot(options PageScreenshotOptions) ([]byte, error) {
	return r.page.Screenshot(playwright.PageScreenshotOptions{
		Path: &options.Path,
	})
}

func (r *RealPage) Close() error {
	return r.page.Close()
}

func (r *RealPage) On(event string, handler interface{}) {
	if event == "download" {
		r.page.OnDownload(func(d playwright.Download) {
			if fn, ok := handler.(func(Download)); ok {
				fn(&RealDownload{download: d})
			}
		})
	} else if event == "dialog" {
		r.page.OnDialog(func(d playwright.Dialog) {
			// Pass the dialog as interface{} to match the handler signature
			if fn, ok := handler.(func(interface{})); ok {
				fn(d)
			}
		})
	}
}

func (r *RealPage) Evaluate(expression string, arg ...interface{}) (interface{}, error) {
	return r.page.Evaluate(expression, arg...)
}

// RealLocator wraps playwright.Locator
type RealLocator struct {
	locator playwright.Locator
}

func (r *RealLocator) Count() (int, error) {
	return r.locator.Count()
}

func (r *RealLocator) First() LocatorInterface {
	return &RealLocator{locator: r.locator.First()}
}

func (r *RealLocator) Fill(value string) error {
	return r.locator.Fill(value)
}

func (r *RealLocator) Click(options LocatorClickOptions) error {
	return r.locator.Click()
}

func (r *RealLocator) TextContent(options LocatorTextContentOptions) (string, error) {
	return r.locator.TextContent()
}

func (r *RealLocator) Check(options LocatorCheckOptions) error {
	return r.locator.Check()
}

func (r *RealLocator) IsChecked(options LocatorIsCheckedOptions) (bool, error) {
	return r.locator.IsChecked()
}

// RealDownload wraps playwright.Download
type RealDownload struct {
	download playwright.Download
}

func (r *RealDownload) SuggestedFilename() string {
	return r.download.SuggestedFilename()
}

func (r *RealDownload) SaveAs(path string) error {
	return r.download.SaveAs(path)
}
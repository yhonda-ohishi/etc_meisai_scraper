package scraper

// ScraperInterface defines the interface for ETC scraping operations
type ScraperInterface interface {
	Initialize() error
	Login() error
	DownloadMeisai(fromDate, toDate string) (string, error)
	Close() error
}
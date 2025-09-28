package services

import (
	"log"

	"github.com/yhonda-ohishi/etc_meisai/src/scraper"
)

// ScraperFactory creates scraper instances
type ScraperFactory interface {
	CreateScraper(config *scraper.ScraperConfig, logger *log.Logger) (scraper.ScraperInterface, error)
}

// DefaultScraperFactory creates real ETCScraper instances
type DefaultScraperFactory struct{}

// CreateScraper creates a new ETCScraper instance
func (f *DefaultScraperFactory) CreateScraper(config *scraper.ScraperConfig, logger *log.Logger) (scraper.ScraperInterface, error) {
	return scraper.NewETCScraper(config, logger)
}

// NewDefaultScraperFactory creates a new default scraper factory
func NewDefaultScraperFactory() ScraperFactory {
	return &DefaultScraperFactory{}
}
package etc_meisai

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/yhonda-ohishi/etc_meisai/handlers"
)

// SetupRoutes configures all routes for the ETC meisai module
func SetupRoutes(r *chi.Mux, handler *handlers.ETCHandler, importHandler *handlers.ImportHandler) {
	// Apply middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)

	// Health check endpoint (use the new one from api_handlers)
	r.HandleFunc("/health", HealthCheckHandler)

	// API routes
	r.Route("/api/etc", func(r chi.Router) {
		// Scraping and download endpoints (no database required)
		r.HandleFunc("/accounts", GetAvailableAccountsHandler)
		r.HandleFunc("/download", DownloadETCDataHandler)
		r.HandleFunc("/download-single", DownloadSingleAccountHandler)
		r.HandleFunc("/download-async", StartDownloadJobHandler)
		r.HandleFunc("/download-status/*", GetDownloadJobStatusHandler)
		r.HandleFunc("/parse-csv", ParseCSVHandler)

		// Import endpoints
		r.Post("/import", handler.ImportData)
		r.Post("/bulk-import", handler.BulkImport)

		// New import endpoints
		r.Post("/import/web", importHandler.ImportFromWeb)
		r.Post("/import/csv", importHandler.ImportCSVFile)

		// Meisai endpoints
		r.Get("/meisai", handler.GetMeisai)
		r.Post("/meisai", handler.CreateMeisai)
		r.Get("/meisai/{id}", handler.GetMeisaiByID)

		// Summary endpoint
		r.Get("/summary", handler.GetSummary)
	})
}
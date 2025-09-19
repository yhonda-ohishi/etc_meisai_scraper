package etc_meisai

import (
	"github.com/go-chi/chi/v5"
	"github.com/yhonda-ohishi/etc_meisai/handlers"
)

// SetupRoutes configures all routes for the module
func SetupRoutes(r *chi.Mux, etcHandler *handlers.ETCHandler, importHandler *handlers.ImportHandler) {
	r.Route("/api", func(r chi.Router) {
		// ETC routes
		r.Route("/etc", func(r chi.Router) {
			r.Post("/import", etcHandler.ImportData)
			r.Get("/meisai", etcHandler.GetMeisai)
			r.Get("/meisai/{id}", etcHandler.GetMeisaiByID)
			r.Post("/meisai", etcHandler.CreateMeisai)
			r.Get("/summary", etcHandler.GetSummary)
			r.Post("/bulk-import", etcHandler.BulkImport)
		})

		// Import routes
		r.Route("/import", func(r chi.Router) {
			r.Post("/web", importHandler.ImportFromWeb)
			r.Post("/csv", importHandler.ImportCSVFile)
		})

		// Health check
		r.Get("/health", etcHandler.HealthCheck)
	})
}
package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/joho/godotenv"
	"github.com/yhonda-ohishi/etc_meisai/src/config"
	"github.com/yhonda-ohishi/etc_meisai/src/handlers"
)

func main() {
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	// Show configured ETC accounts
	showConfiguredAccounts()

	// Create router
	r := chi.NewRouter()

	// Apply middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(corsMiddleware)

	// Initialize handlers with new structure
	baseHandler := handlers.BaseHandler{}
	parseHandler := handlers.NewParseHandler(baseHandler)
	downloadHandler := handlers.NewDownloadHandler(baseHandler, nil)
	accountHandler := handlers.NewAccountsHandler(baseHandler)

	// Register API endpoints
	r.Route("/api", func(r chi.Router) {
		// Parse endpoints
		r.Post("/parse/csv", parseHandler.ParseCSV)

		// Download endpoints
		r.Post("/download/sync", downloadHandler.DownloadSync)
		r.Post("/download/async", downloadHandler.DownloadAsync)
		r.Get("/download/status", downloadHandler.GetDownloadStatus)

		// Account endpoints
		r.Get("/accounts", accountHandler.GetAccounts)
	})

	// Swagger UI
	r.Handle("/docs/*", http.StripPrefix("/docs/", http.FileServer(http.Dir("./docs/"))))
	r.HandleFunc("/docs", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/docs/swagger.html", http.StatusMovedPermanently)
	})

	// Get server port
	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8080"
	}

	// Start server
	addr := fmt.Sprintf(":%s", port)
	log.Printf("Starting ETC Meisai API server on %s", addr)
	log.Printf("Endpoints:")
	log.Printf("  GET  http://localhost%s/api/accounts", addr)
	log.Printf("  POST http://localhost%s/api/parse/csv", addr)
	log.Printf("  POST http://localhost%s/api/download/sync", addr)
	log.Printf("  POST http://localhost%s/api/download/async", addr)
	log.Printf("  GET  http://localhost%s/api/download/status", addr)
	log.Printf("  Swagger UI: http://localhost%s/docs", addr)

	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func showConfiguredAccounts() {
	accounts, err := config.LoadCorporateAccountsFromEnv()
	if err != nil || len(accounts) == 0 {
		log.Println("⚠️  No ETC accounts configured in ETC_CORP_ACCOUNTS")
		log.Println("   Set ETC_CORP_ACCOUNTS environment variable to enable account features")
		return
	}

	fmt.Println("\n=== 📋 Configured ETC Accounts ===")
	for i, account := range accounts {
		fmt.Printf("  %d. %s\n", i+1, account.UserID)
	}
	fmt.Printf("  Total: %d accounts\n", len(accounts))
	fmt.Println("===================================")
}
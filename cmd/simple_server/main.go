package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/joho/godotenv"
	etc "github.com/yhonda-ohishi/etc_meisai"
	"github.com/yhonda-ohishi/etc_meisai/config"
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

	// Register scraping API endpoints (no database required)
	r.Get("/health", etc.HealthCheckHandler)
	r.Get("/api/etc/accounts", etc.GetAvailableAccountsHandler)
	r.Post("/api/etc/download", etc.DownloadETCDataHandler)
	r.Post("/api/etc/download-single", etc.DownloadSingleAccountHandler)
	r.Post("/api/etc/download-async", etc.StartDownloadJobHandler)
	r.HandleFunc("/api/etc/download-status/*", etc.GetDownloadJobStatusHandler)
	r.Post("/api/etc/parse-csv", etc.ParseCSVHandler)

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
	log.Printf("  GET  http://localhost%s/health", addr)
	log.Printf("  GET  http://localhost%s/api/etc/accounts", addr)
	log.Printf("  POST http://localhost%s/api/etc/download", addr)
	log.Printf("  POST http://localhost%s/api/etc/download-single", addr)
	log.Printf("  POST http://localhost%s/api/etc/download-async", addr)
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
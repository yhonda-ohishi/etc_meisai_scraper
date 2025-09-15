package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/joho/godotenv"
	"github.com/yhonda-ohishi/etc_meisai"
	"github.com/yhonda-ohishi/etc_meisai/config"
)

func main() {
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	// Show configured ETC accounts
	showConfiguredAccounts()

	// Connect to database
	dbConfig := config.NewDatabaseConfig()
	db, err := config.ConnectDB(dbConfig)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Create router
	r := chi.NewRouter()

	// Initialize module
	module, err := etc_meisai.InitializeWithRouter(db, r)
	if err != nil {
		log.Fatalf("Failed to initialize module: %v", err)
	}

	// Get server port
	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8080"
	}

	// Start server
	addr := fmt.Sprintf(":%s", port)
	log.Printf("Starting ETC Meisai server on %s", addr)
	log.Printf("API Endpoint: http://localhost%s/api/etc/accounts", addr)
	log.Printf("Module initialized: %+v", module)

	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
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
	fmt.Println("===================================\n")
}
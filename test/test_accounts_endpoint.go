package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"

	"github.com/joho/godotenv"
	etc "github.com/yhonda-ohishi/etc_meisai"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not loaded")
	}

	// Create a test server
	mux := http.NewServeMux()
	etc.RegisterAPIHandlers(mux)

	// Test the accounts endpoint
	req := httptest.NewRequest("GET", "/api/etc/accounts", nil)
	w := httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	// Parse response
	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		log.Fatal("Failed to parse response:", err)
	}

	// Display results
	fmt.Printf("=== Account List Endpoint Test ===\n")
	fmt.Printf("Status Code: %d\n", w.Code)
	fmt.Printf("Response:\n")
	fmt.Printf("  Configured: %v\n", response["configured"])
	fmt.Printf("  Count: %v\n", response["count"])
	fmt.Printf("  Message: %v\n", response["message"])

	if accounts, ok := response["accounts"].([]interface{}); ok {
		fmt.Printf("  Accounts:\n")
		for _, account := range accounts {
			fmt.Printf("    - %v\n", account)
		}
	}

	// Also test with direct environment variable
	if envAccounts := os.Getenv("ETC_CORP_ACCOUNTS"); envAccounts != "" {
		fmt.Printf("\n=== Environment Variable Check ===\n")
		fmt.Printf("ETC_CORP_ACCOUNTS is set (length: %d chars)\n", len(envAccounts))
	} else {
		fmt.Printf("\n=== Environment Variable Check ===\n")
		fmt.Printf("ETC_CORP_ACCOUNTS is not set\n")
	}
}
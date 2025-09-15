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
	log.Printf("Module initialized: %+v", module)

	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
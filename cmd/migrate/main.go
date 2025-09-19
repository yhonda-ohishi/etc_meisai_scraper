package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/yhonda-ohishi/etc_meisai/src/clients"
	"github.com/yhonda-ohishi/etc_meisai/src/config"
)

func main() {
	// Initialize configuration
	if err := config.InitSettings(); err != nil {
		log.Fatalf("Failed to initialize settings: %v", err)
	}

	cfg := config.GlobalSettings

	// Create logger
	logger := log.New(os.Stdout, "[MIGRATE] ", log.LstdFlags)
	logger.Println("Migration tool (gRPC-only mode)")

	// Initialize gRPC client to db_service
	address := cfg.GRPC.GetDBServiceAddress()
	timeout := cfg.GRPC.GetConnectionTimeout()

	dbClient, err := clients.NewDBServiceClient(address, timeout)
	if err != nil {
		log.Fatalf("Failed to connect to db_service: %v", err)
	}
	defer dbClient.Close()

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := dbClient.HealthCheck(ctx); err != nil {
		log.Fatalf("db_service health check failed: %v", err)
	}

	logger.Println("Successfully connected to db_service")
	logger.Println("Note: Schema migrations should be performed directly on db_service")
	logger.Println("This tool now only verifies connectivity to db_service")
}
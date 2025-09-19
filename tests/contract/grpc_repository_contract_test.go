package contract

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/yhonda-ohishi/etc_meisai/src/clients"
	"github.com/yhonda-ohishi/etc_meisai/src/config"
	"github.com/yhonda-ohishi/etc_meisai/src/models"
	"github.com/yhonda-ohishi/etc_meisai/src/repositories"
)

// TestGRPCRepositoryContract tests that GRPCRepository satisfies the ETCRepository contract
func TestGRPCRepositoryContract(t *testing.T) {
	// Skip if not in integration test mode or gRPC server not available
	if os.Getenv("DB_SERVICE_ADDRESS") == "" {
		t.Skip("Skipping contract test: DB_SERVICE_ADDRESS not set")
	}

	// Initialize configuration
	if err := config.InitSettings(); err != nil {
		t.Fatalf("Failed to initialize settings: %v", err)
	}

	cfg := config.GlobalSettings
	address := cfg.GRPC.GetDBServiceAddress()
	timeout := cfg.GRPC.GetConnectionTimeout()

	// Create gRPC client
	dbClient, err := clients.NewDBServiceClient(address, timeout)
	if err != nil {
		t.Fatalf("Failed to create gRPC client: %v", err)
	}
	defer dbClient.Close()

	// Create contract test instance
	contract := &ETCRepositoryContract{
		NewRepository: func() repositories.ETCRepository {
			return repositories.NewGRPCRepository(dbClient)
		},
		CleanupFunc: func() {
			// Cleanup is handled by db_service
		},
	}

	// Run all contract tests
	contract.RunContractTests(t)
}

// TestGRPCRepositoryPerformance tests performance requirements
func TestGRPCRepositoryPerformance(t *testing.T) {
	if os.Getenv("DB_SERVICE_ADDRESS") == "" {
		t.Skip("Skipping performance test: DB_SERVICE_ADDRESS not set")
	}

	// Initialize configuration
	if err := config.InitSettings(); err != nil {
		t.Fatalf("Failed to initialize settings: %v", err)
	}

	cfg := config.GlobalSettings
	address := cfg.GRPC.GetDBServiceAddress()
	timeout := cfg.GRPC.GetConnectionTimeout()

	// Create gRPC client
	dbClient, err := clients.NewDBServiceClient(address, timeout)
	if err != nil {
		t.Fatalf("Failed to create gRPC client: %v", err)
	}
	defer dbClient.Close()

	repo := repositories.NewGRPCRepository(dbClient)

	// Test bulk insert performance (10,000 records in < 5 seconds)
	t.Run("BulkInsertPerformance", func(t *testing.T) {
		start := time.Now()

		// Prepare 10,000 records
		var records []*models.ETCMeisai
		for i := 0; i < 10000; i++ {
			etc := &models.ETCMeisai{
				UseDate:   time.Date(2025, 1, 20, 0, 0, 0, 0, time.UTC),
				UseTime:   fmt.Sprintf("%02d:%02d", i/60%24, i%60),
				EntryIC:   fmt.Sprintf("Entry_%d", i%100),
				ExitIC:    fmt.Sprintf("Exit_%d", i%100),
				Amount:    int32(500 + (i%20)*50),
				CarNumber: fmt.Sprintf("PERF_%04d", i%1000),
				ETCNumber: fmt.Sprintf("%016d", i%10000),
			}
			etc.Hash = etc.GenerateHash()
			records = append(records, etc)
		}

		// Perform bulk insert
		err := repo.BulkInsert(records)
		if err != nil {
			t.Fatalf("BulkInsert failed: %v", err)
		}

		elapsed := time.Since(start)
		t.Logf("Bulk insert of 10,000 records took: %v", elapsed)

		// Assert performance requirement
		if elapsed > 5*time.Second {
			t.Errorf("Performance requirement failed: took %v, expected < 5 seconds", elapsed)
		}
	})
}
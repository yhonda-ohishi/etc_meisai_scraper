package integration

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yhonda-ohishi/etc_meisai/src/clients"
	"github.com/yhonda-ohishi/etc_meisai/src/config"
	"github.com/yhonda-ohishi/etc_meisai/src/models"
	"github.com/yhonda-ohishi/etc_meisai/src/repositories"
	"github.com/yhonda-ohishi/etc_meisai/src/services"
)

func TestBasicIntegration(t *testing.T) {
	// Skip if gRPC server is not available
	if os.Getenv("DB_SERVICE_ADDRESS") == "" {
		t.Skip("Skipping integration test: DB_SERVICE_ADDRESS not set")
	}

	// Setup gRPC client
	dbClient := setupTestGRPCClient(t)
	defer dbClient.Close()

	// Test repository integration
	t.Run("Repository Integration", func(t *testing.T) {
		testRepositoryIntegration(t, dbClient)
	})

	// Test service integration
	t.Run("Service Integration", func(t *testing.T) {
		testServiceIntegration(t, dbClient)
	})
}

func setupTestGRPCClient(t *testing.T) *clients.DBServiceClient {
	// Initialize configuration
	if err := config.InitSettings(); err != nil {
		t.Fatalf("Failed to initialize settings: %v", err)
	}

	cfg := config.GlobalSettings
	address := cfg.GRPC.GetDBServiceAddress()
	timeout := cfg.GRPC.GetConnectionTimeout()

	// Create gRPC client
	dbClient, err := clients.NewDBServiceClient(address, timeout)
	require.NoError(t, err, "Failed to create gRPC client")

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = dbClient.HealthCheck(ctx)
	require.NoError(t, err, "gRPC health check failed")

	return dbClient
}

func testRepositoryIntegration(t *testing.T, dbClient *clients.DBServiceClient) {
	repo := repositories.NewGRPCRepository(dbClient)

	// Create test record
	testRecord := &models.ETCMeisai{
		UseDate:   time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
		UseTime:   "08:30",
		EntryIC:   "東京IC",
		ExitIC:    "横浜IC",
		Amount:    1200,
		CarNumber: "TEST001",
		ETCNumber: "1234567890123456",
	}
	testRecord.Hash = testRecord.GenerateHash()

	// Test Create
	err := repo.Create(testRecord)
	assert.NoError(t, err)
	assert.NotZero(t, testRecord.ID)

	// Test GetByID
	retrieved, err := repo.GetByID(testRecord.ID)
	assert.NoError(t, err)
	assert.Equal(t, testRecord.Hash, retrieved.Hash)
	assert.Equal(t, testRecord.EntryIC, retrieved.EntryIC)
	assert.Equal(t, testRecord.ExitIC, retrieved.ExitIC)

	// Test GetByDateRange
	from := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2024, 1, 31, 0, 0, 0, 0, time.UTC)
	records, err := repo.GetByDateRange(from, to)
	assert.NoError(t, err)
	assert.Len(t, records, 1)

	// Test List with pagination
	params := &models.ETCListParams{
		Limit:  10,
		Offset: 0,
	}
	records, total, err := repo.List(params)
	assert.NoError(t, err)
	assert.Len(t, records, 1)
	assert.Equal(t, int64(1), total)

	// Test duplicate detection
	duplicates, err := repo.CheckDuplicatesByHash([]string{testRecord.Hash})
	assert.NoError(t, err)
	assert.True(t, duplicates[testRecord.Hash])
}

func testServiceIntegration(t *testing.T, dbClient *clients.DBServiceClient) {
	repo := repositories.NewGRPCRepository(dbClient)
	service := services.NewETCService(repo, dbClient)

	ctx := context.Background()

	// Test Create
	testRecord := &models.ETCMeisai{
		UseDate:   time.Date(2024, 1, 16, 0, 0, 0, 0, time.UTC),
		UseTime:   "17:45",
		EntryIC:   "横浜IC",
		ExitIC:    "東京IC",
		Amount:    1200,
		CarNumber: "TEST002",
		ETCNumber: "1234567890123456",
	}

	created, err := service.Create(ctx, testRecord)
	assert.NoError(t, err)
	assert.NotZero(t, created.ID)
	assert.NotEmpty(t, created.Hash)

	// Test GetByID
	retrieved, err := service.GetByID(ctx, created.ID)
	assert.NoError(t, err)
	assert.Equal(t, created.Hash, retrieved.Hash)

	// Test GetByDateRange
	from := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2024, 1, 31, 0, 0, 0, 0, time.UTC)
	records, err := service.GetByDateRange(ctx, from, to)
	assert.NoError(t, err)
	assert.Len(t, records, 1)

	// Test ImportCSV
	csvRecords := []*models.ETCMeisai{
		{
			UseDate:   time.Date(2024, 1, 17, 0, 0, 0, 0, time.UTC),
			UseTime:   "09:00",
			EntryIC:   "渋谷IC",
			ExitIC:    "新宿IC",
			Amount:    800,
			CarNumber: "TEST003",
			ETCNumber: "1234567890123456",
		},
		{
			UseDate:   time.Date(2024, 1, 18, 0, 0, 0, 0, time.UTC),
			UseTime:   "18:30",
			EntryIC:   "新宿IC",
			ExitIC:    "渋谷IC",
			Amount:    800,
			CarNumber: "TEST003",
			ETCNumber: "1234567890123456",
		},
	}

	result, err := service.ImportCSV(ctx, csvRecords)
	assert.NoError(t, err)
	assert.True(t, result.Success)
	assert.Equal(t, 2, result.ImportedRows)

	// Verify total records
	allParams := &models.ETCListParams{Limit: 100, Offset: 0}
	allRecords, total, err := service.List(ctx, allParams)
	assert.NoError(t, err)
	assert.Equal(t, int64(3), total) // 1 from repository test + 1 from service test + 2 from CSV import
	assert.Len(t, allRecords, 3)
}

func TestModelValidation(t *testing.T) {
	t.Run("Valid ETC Record", func(t *testing.T) {
		record := &models.ETCMeisai{
			UseDate:   time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
			UseTime:   "08:30",
			EntryIC:   "東京IC",
			ExitIC:    "横浜IC",
			Amount:    1200,
			CarNumber: "TEST001",
			ETCNumber: "1234567890123456",
		}
		record.Hash = record.GenerateHash() // Generate hash before validation

		err := record.Validate()
		assert.NoError(t, err)
	})

	t.Run("Invalid ETC Record - Missing Required Fields", func(t *testing.T) {
		record := &models.ETCMeisai{
			UseDate: time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
			// Missing other required fields
		}

		err := record.Validate()
		assert.Error(t, err)
	})

	t.Run("Invalid ETC Record - Negative Amount", func(t *testing.T) {
		record := &models.ETCMeisai{
			UseDate:   time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
			UseTime:   "08:30",
			EntryIC:   "東京IC",
			ExitIC:    "横浜IC",
			Amount:    -100, // Invalid negative amount
			CarNumber: "TEST001",
			ETCNumber: "1234567890123456",
		}

		err := record.Validate()
		assert.Error(t, err)
	})
}

func TestHashGeneration(t *testing.T) {
	record1 := &models.ETCMeisai{
		UseDate:   time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
		UseTime:   "08:30",
		EntryIC:   "東京IC",
		ExitIC:    "横浜IC",
		Amount:    1200,
		CarNumber: "TEST001",
		ETCNumber: "1234567890123456",
	}

	record2 := &models.ETCMeisai{
		UseDate:   time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
		UseTime:   "08:30",
		EntryIC:   "東京IC",
		ExitIC:    "横浜IC",
		Amount:    1200,
		CarNumber: "TEST001",
		ETCNumber: "1234567890123456",
	}

	record3 := &models.ETCMeisai{
		UseDate:   time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
		UseTime:   "08:30",
		EntryIC:   "東京IC",
		ExitIC:    "横浜IC",
		Amount:    1500, // Different amount
		CarNumber: "TEST001",
		ETCNumber: "1234567890123456",
	}

	hash1 := record1.GenerateHash()
	hash2 := record2.GenerateHash()
	hash3 := record3.GenerateHash()

	// Same data should generate same hash
	assert.Equal(t, hash1, hash2)

	// Different data should generate different hash
	assert.NotEqual(t, hash1, hash3)

	// Hash should be 64 characters (SHA256 hex)
	assert.Len(t, hash1, 64)
}
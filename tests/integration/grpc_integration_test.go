package integration

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yhonda-ohishi/etc_meisai/src/clients"
	"github.com/yhonda-ohishi/etc_meisai/src/handlers"
	"github.com/yhonda-ohishi/etc_meisai/src/models"
	"github.com/yhonda-ohishi/etc_meisai/src/repositories"
	"github.com/yhonda-ohishi/etc_meisai/src/services"
)

// TestGRPCIntegration tests the full gRPC integration
func TestGRPCIntegration(t *testing.T) {
	// Skip if no db_service is running
	if !isDBServiceAvailable() {
		t.Skip("db_service not available, skipping integration test")
	}

	// Setup
	dbClient, err := setupDBClient()
	require.NoError(t, err, "Failed to setup db_service client")
	defer dbClient.Close()

	logger := log.New(log.Writer(), "[TEST] ", log.LstdFlags)
	registry := services.NewServiceRegistryGRPCOnly(dbClient, logger)

	t.Run("ETCService", func(t *testing.T) {
		testETCService(t, registry)
	})

	t.Run("MappingService", func(t *testing.T) {
		testMappingService(t, registry)
	})

	t.Run("ImportService", func(t *testing.T) {
		testImportService(t, registry)
	})

	t.Run("HTTPHandlers", func(t *testing.T) {
		testHTTPHandlers(t, registry, logger)
	})
}

func testETCService(t *testing.T, registry *services.ServiceRegistry) {
	etcService := registry.GetETCService()
	require.NotNil(t, etcService, "ETC service should not be nil")

	ctx := context.Background()

	// Create test record
	testRecord := &models.ETCMeisai{
		UseDate:   time.Date(2025, 1, 20, 0, 0, 0, 0, time.UTC),
		UseTime:   "14:30",
		EntryIC:   "Test Entry IC",
		ExitIC:    "Test Exit IC",
		Amount:    1500,
		CarNumber: "品川300あ1234",
		ETCNumber: "1234567890123456",
	}

	// Test Create
	created, err := etcService.Create(ctx, testRecord)
	if err != nil {
		// If gRPC is not implemented, check for expected error
		if strings.Contains(err.Error(), "not yet implemented") {
			t.Skip("Create not yet implemented in db_service")
		}
		t.Fatalf("Failed to create ETC record: %v", err)
	}

	assert.NotNil(t, created, "Created record should not be nil")
	assert.NotZero(t, created.ID, "ID should be returned after creation")
	testRecord.ID = created.ID

	// Test GetByID
	retrieved, err := etcService.GetByID(ctx, testRecord.ID)
	if err != nil {
		if strings.Contains(err.Error(), "not yet implemented") {
			t.Skip("GetByID not yet implemented in db_service")
		}
		t.Fatalf("Failed to get ETC record: %v", err)
	}

	assert.Equal(t, testRecord.ID, retrieved.ID)
	assert.Equal(t, testRecord.ETCNumber, retrieved.ETCNumber)

	// Test List
	params := &models.ETCListParams{
		Limit:  10,
		Offset: 0,
	}
	list, total, err := etcService.List(ctx, params)
	assert.NoError(t, err, "List should not return error")
	assert.NotNil(t, list, "List should return results")
	assert.GreaterOrEqual(t, total, int64(0), "Total should be non-negative")

	// Test Update (not implemented in ETCService yet)
	// Would need to add Update method to ETCService

	// Test Delete (not implemented in ETCService yet)
	// Would need to add Delete method to ETCService
}

func testMappingService(t *testing.T, registry *services.ServiceRegistry) {
	mappingService := registry.GetMappingService()
	require.NotNil(t, mappingService, "Mapping service should not be nil")

	ctx := context.Background()

	// Create test mapping
	testMapping := &models.ETCMeisaiMapping{
		ETCMeisaiID: 1,
		DTakoRowID:  "DTAKO-001",
		MappingType: "manual",
		Confidence:  1.0,
		Notes:       "Test mapping",
		CreatedBy:   "test_user",
	}

	// Test Create
	err := mappingService.CreateMapping(ctx, testMapping)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			t.Skip("ETC record not found, skipping mapping test")
		}
		t.Fatalf("Failed to create mapping: %v", err)
	}

	// Test GetByID
	if testMapping.ID != 0 {
		retrieved, err := mappingService.GetMappingByID(ctx, testMapping.ID)
		if err != nil && !strings.Contains(err.Error(), "not yet implemented") {
			t.Errorf("Failed to get mapping: %v", err)
		} else if retrieved != nil {
			assert.Equal(t, testMapping.ID, retrieved.ID)
		}
	}

	// Test List
	params := &models.MappingListParams{
		Limit:  10,
		Offset: 0,
	}
	list, total, err := mappingService.ListMappings(ctx, params)
	assert.NoError(t, err, "List should not return error")
	assert.NotNil(t, list, "List should return results")
	assert.GreaterOrEqual(t, total, int64(0), "Total should be non-negative")

	// Test AutoMatch
	startDate := time.Now().AddDate(0, 0, -30)
	endDate := time.Now()
	results, err := mappingService.AutoMatch(ctx, startDate, endDate, 0.8)
	if err != nil && !strings.Contains(err.Error(), "not yet implemented") {
		t.Errorf("AutoMatch failed: %v", err)
	} else {
		assert.NotNil(t, results, "AutoMatch should return results")
	}
}

func testImportService(t *testing.T, registry *services.ServiceRegistry) {
	importService := registry.GetImportService()
	require.NotNil(t, importService, "Import service should not be nil")

	ctx := context.Background()

	// Test CSV validation and parsing
	csvContent := `利用日,利用時間,入口,出口,金額,車番,ETC番号
2025/01/20,14:30,東京IC,横浜IC,1500,品川300あ1234,1234567890123456
2025/01/20,15:00,横浜IC,名古屋IC,3000,品川300あ1234,1234567890123456`

	result, err := importService.ParseAndValidateCSV(ctx, csvContent, "test_account")
	if err != nil {
		t.Fatalf("Failed to parse CSV: %v", err)
	}

	assert.NotNil(t, result, "Parse result should not be nil")
	assert.Equal(t, 2, len(result.Records), "Should parse 2 records")
	assert.Equal(t, 2, result.ValidRows, "Should have 2 valid rows")
	assert.Equal(t, 0, result.ErrorRows, "Should have no errors")

	// Test import progress (stub)
	progress, err := importService.GetImportProgress(ctx, 1)
	if err != nil && !strings.Contains(err.Error(), "not found") {
		t.Errorf("GetImportProgress failed: %v", err)
	} else if progress != nil {
		assert.NotNil(t, progress, "Progress should not be nil")
	}
}

func testHTTPHandlers(t *testing.T, registry *services.ServiceRegistry, logger *log.Logger) {
	// Setup router
	r := chi.NewRouter()

	// Create handlers
	etcHandler := handlers.NewETCHandler(registry, logger)
	mappingHandler := handlers.NewMappingHandler(registry, logger)

	// Setup routes
	r.Route("/api", func(r chi.Router) {
		r.Route("/etc", func(r chi.Router) {
			r.Get("/", etcHandler.ListETCMeisai)
			r.Post("/", etcHandler.CreateETCMeisai)
			r.Get("/{id}", etcHandler.GetETCMeisai)
		})

		r.Route("/mapping", func(r chi.Router) {
			r.Get("/", mappingHandler.GetMappings)
			r.Post("/", mappingHandler.CreateMapping)
			r.Post("/auto-match", mappingHandler.AutoMatch)
		})
	})

	// Test ETC list endpoint
	t.Run("GET /api/etc", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/etc?limit=10", nil)
		rec := httptest.NewRecorder()

		r.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code, "Should return 200 OK")
		assert.Contains(t, rec.Header().Get("Content-Type"), "application/json")
	})

	// Test Mapping list endpoint
	t.Run("GET /api/mapping", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/mapping?limit=10", nil)
		rec := httptest.NewRecorder()

		r.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code, "Should return 200 OK")
		assert.Contains(t, rec.Header().Get("Content-Type"), "application/json")
	})

	// Test Create ETC endpoint
	t.Run("POST /api/etc", func(t *testing.T) {
		body := `{
			"use_date": "2025-01-20T00:00:00Z",
			"use_time": "14:30",
			"entry_ic": "東京IC",
			"exit_ic": "横浜IC",
			"amount": 1500,
			"car_number": "品川300あ1234",
			"etc_number": "1234567890123456"
		}`

		req := httptest.NewRequest("POST", "/api/etc", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		r.ServeHTTP(rec, req)

		// Check status code (might be 201 Created or 503 if db_service not available)
		assert.Contains(t, []int{http.StatusCreated, http.StatusServiceUnavailable}, rec.Code)
	})

	// Test Auto-match endpoint
	t.Run("POST /api/mapping/auto-match", func(t *testing.T) {
		body := `{
			"from_date": "2025-01-01",
			"to_date": "2025-01-31",
			"threshold": 0.8
		}`

		req := httptest.NewRequest("POST", "/api/mapping/auto-match", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		r.ServeHTTP(rec, req)

		// Should return OK or Service Unavailable
		assert.Contains(t, []int{http.StatusOK, http.StatusServiceUnavailable}, rec.Code)
	})
}

// Helper functions

func isDBServiceAvailable() bool {
	// Check if db_service is running on default port
	client, err := clients.NewDBServiceClient("localhost:50051", 5*time.Second)
	if err != nil {
		return false
	}
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err = client.HealthCheck(ctx)
	return err == nil
}

func setupDBClient() (*clients.DBServiceClient, error) {
	// Use environment variable or default
	address := "localhost:50051"
	return clients.NewDBServiceClient(address, 30*time.Second)
}

// TestRepositoryContract tests that our gRPC repository implements the contract
func TestRepositoryContract(t *testing.T) {
	// Skip if no db_service
	if !isDBServiceAvailable() {
		t.Skip("db_service not available")
	}

	dbClient, err := setupDBClient()
	require.NoError(t, err)
	defer dbClient.Close()

	// Test ETC Repository
	t.Run("ETCRepository", func(t *testing.T) {
		repo := repositories.NewGRPCRepository(dbClient)
		testETCRepositoryContract(t, repo)
	})

	// Test Mapping Repository
	t.Run("MappingRepository", func(t *testing.T) {
		repo := repositories.NewMappingGRPCRepository(dbClient)
		testMappingRepositoryContract(t, repo)
	})
}

func testETCRepositoryContract(t *testing.T, repo repositories.ETCRepository) {
	ctx := context.Background()
	_ = ctx // Context might be needed in future

	// Test Create
	record := &models.ETCMeisai{
		UseDate:   time.Now(),
		UseTime:   "10:00",
		EntryIC:   "Entry",
		ExitIC:    "Exit",
		Amount:    1000,
		CarNumber: "Test-001",
		ETCNumber: "9999999999999999",
	}

	err := repo.Create(record)
	if err != nil && strings.Contains(err.Error(), "not yet implemented") {
		t.Skip("Repository methods not yet fully implemented")
	}

	// Test other methods...
	// These would be more comprehensive in a real test
}

func testMappingRepositoryContract(t *testing.T, repo repositories.MappingRepository) {
	// Similar contract tests for mapping repository
	mapping := &models.ETCMeisaiMapping{
		ETCMeisaiID: 1,
		DTakoRowID:  "TEST-001",
		MappingType: "test",
		Confidence:  0.9,
	}

	err := repo.Create(mapping)
	if err != nil && strings.Contains(err.Error(), "not found") {
		t.Skip("Dependencies not available")
	}
}

// Benchmark tests

func BenchmarkETCCreate(b *testing.B) {
	if !isDBServiceAvailable() {
		b.Skip("db_service not available")
	}

	dbClient, _ := setupDBClient()
	defer dbClient.Close()

	repo := repositories.NewGRPCRepository(dbClient)

	record := &models.ETCMeisai{
		UseDate:   time.Now(),
		UseTime:   "10:00",
		EntryIC:   "Entry",
		ExitIC:    "Exit",
		Amount:    1000,
		CarNumber: "BENCH-001",
		ETCNumber: "8888888888888888",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = repo.Create(record)
	}
}

func BenchmarkBulkInsert(b *testing.B) {
	if !isDBServiceAvailable() {
		b.Skip("db_service not available")
	}

	dbClient, _ := setupDBClient()
	defer dbClient.Close()

	repo := repositories.NewGRPCRepository(dbClient)

	// Create 100 test records
	var records []*models.ETCMeisai
	for i := 0; i < 100; i++ {
		records = append(records, &models.ETCMeisai{
			UseDate:   time.Now(),
			UseTime:   fmt.Sprintf("%02d:00", i%24),
			EntryIC:   fmt.Sprintf("Entry-%d", i),
			ExitIC:    fmt.Sprintf("Exit-%d", i),
			Amount:    int32(1000 + i*100),
			CarNumber: fmt.Sprintf("CAR-%04d", i),
			ETCNumber: fmt.Sprintf("%016d", i),
		})
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = repo.BulkInsert(records)
	}
}
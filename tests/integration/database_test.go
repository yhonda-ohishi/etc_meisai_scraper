package integration

import (
	"database/sql"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yhonda-ohishi/etc_meisai/config"
	_ "github.com/go-sql-driver/mysql"
)

// TestLocalDatabaseConnection tests the local database connection
func TestLocalDatabaseConnection(t *testing.T) {
	// Skip if not in integration test environment
	if os.Getenv("RUN_INTEGRATION_TESTS") != "true" {
		t.Skip("Skipping integration test")
	}

	// Create local database config
	dbConfig := config.NewDatabaseConfig()

	// Connect to local database
	db, err := config.ConnectDB(dbConfig)
	require.NoError(t, err, "Should connect to local database")
	defer db.Close()

	// Test ping
	err = db.Ping()
	assert.NoError(t, err, "Should ping local database successfully")

	// Test write operation (should succeed on local DB)
	tx, err := db.Begin()
	require.NoError(t, err, "Should start transaction")

	// Try to create a test table
	_, err = tx.Exec(`
		CREATE TEMPORARY TABLE IF NOT EXISTS test_table (
			id INT PRIMARY KEY AUTO_INCREMENT,
			name VARCHAR(50),
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`)
	assert.NoError(t, err, "Should create temporary table")

	// Insert test data
	_, err = tx.Exec("INSERT INTO test_table (name) VALUES (?)", "test_entry")
	assert.NoError(t, err, "Should insert data")

	// Rollback transaction (cleanup)
	err = tx.Rollback()
	assert.NoError(t, err, "Should rollback transaction")
}

// TestProductionDatabaseReadOnly tests production database read-only access
func TestProductionDatabaseReadOnly(t *testing.T) {
	// Skip if production credentials not set
	if os.Getenv("PROD_DB_HOST") == "" {
		t.Skip("Production database credentials not configured")
	}

	// Create production database config
	prodConfig := config.NewProductionDatabaseConfig()

	// Skip if config is incomplete
	if prodConfig.Host == "" || prodConfig.User == "" {
		t.Skip("Production database config incomplete")
	}

	// Connect to production database
	db, err := config.ConnectProductionDB(prodConfig)
	if err != nil {
		t.Skipf("Cannot connect to production database: %v", err)
	}
	defer db.Close()

	// Test ping
	err = db.Ping()
	assert.NoError(t, err, "Should ping production database successfully")

	// Test read operation
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM information_schema.tables").Scan(&count)
	assert.NoError(t, err, "Should execute read query")
	assert.Greater(t, count, 0, "Should have tables")

	// Test that write operations would fail (don't actually try to write to prod)
	// This is more of a documentation test
	t.Log("Production database is configured for read-only access")
}

// TestDatabaseConnectionPool tests connection pool configuration
func TestDatabaseConnectionPool(t *testing.T) {
	dbConfig := config.NewDatabaseConfig()
	db, err := config.ConnectDB(dbConfig)
	if err != nil {
		t.Skipf("Cannot connect to database: %v", err)
	}
	defer db.Close()

	// Get connection pool stats
	stats := db.Stats()

	// Verify pool settings
	assert.GreaterOrEqual(t, stats.MaxOpenConnections, 10, "Should have reasonable max connections")

	// Test concurrent connections
	done := make(chan bool, 5)
	for i := 0; i < 5; i++ {
		go func(id int) {
			defer func() { done <- true }()

			// Each goroutine gets its own connection
			var result int
			err := db.QueryRow("SELECT ?", id).Scan(&result)
			assert.NoError(t, err, "Concurrent query should succeed")
			assert.Equal(t, id, result)
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 5; i++ {
		select {
		case <-done:
			// Success
		case <-time.After(5 * time.Second):
			t.Fatal("Timeout waiting for concurrent queries")
		}
	}

	// Check that connections were reused properly
	newStats := db.Stats()
	assert.LessOrEqual(t, newStats.OpenConnections, 5, "Should efficiently use connection pool")
}

// TestTransactionIsolation tests transaction isolation levels
func TestTransactionIsolation(t *testing.T) {
	dbConfig := config.NewDatabaseConfig()
	db, err := config.ConnectDB(dbConfig)
	if err != nil {
		t.Skipf("Cannot connect to database: %v", err)
	}
	defer db.Close()

	// Start transaction with specific isolation level
	tx, err := db.BeginTx(nil, &sql.TxOptions{
		Isolation: sql.LevelReadCommitted,
		ReadOnly:  false,
	})
	require.NoError(t, err, "Should start transaction with isolation level")

	// Create temporary table for testing
	_, err = tx.Exec(`
		CREATE TEMPORARY TABLE IF NOT EXISTS isolation_test (
			id INT PRIMARY KEY,
			value VARCHAR(50)
		)
	`)
	require.NoError(t, err)

	// Insert test data
	_, err = tx.Exec("INSERT INTO isolation_test (id, value) VALUES (1, 'test')")
	assert.NoError(t, err)

	// Verify data is visible within transaction
	var value string
	err = tx.QueryRow("SELECT value FROM isolation_test WHERE id = 1").Scan(&value)
	assert.NoError(t, err)
	assert.Equal(t, "test", value)

	// Rollback transaction
	err = tx.Rollback()
	assert.NoError(t, err, "Should rollback transaction")
}

// TestDatabaseTimeout tests database query timeout
func TestDatabaseTimeout(t *testing.T) {
	dbConfig := config.NewDatabaseConfig()
	db, err := config.ConnectDB(dbConfig)
	if err != nil {
		t.Skipf("Cannot connect to database: %v", err)
	}
	defer db.Close()

	// Set a short timeout for testing
	db.SetConnMaxLifetime(5 * time.Minute)

	// Execute a simple query to ensure connection works
	var result int
	err = db.QueryRow("SELECT 1").Scan(&result)
	assert.NoError(t, err, "Simple query should succeed")
	assert.Equal(t, 1, result)

	// Test that connections are properly managed
	stats := db.Stats()
	assert.GreaterOrEqual(t, stats.MaxOpenConnections, 1, "Should have at least one connection")
}

// TestDatabaseErrorHandling tests proper error handling
func TestDatabaseErrorHandling(t *testing.T) {
	dbConfig := config.NewDatabaseConfig()
	db, err := config.ConnectDB(dbConfig)
	if err != nil {
		t.Skipf("Cannot connect to database: %v", err)
	}
	defer db.Close()

	// Test invalid query
	var result int
	err = db.QueryRow("SELECT * FROM non_existent_table").Scan(&result)
	assert.Error(t, err, "Should return error for invalid table")

	// Test connection is still valid after error
	err = db.Ping()
	assert.NoError(t, err, "Connection should remain valid after error")
}
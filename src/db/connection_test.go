package db

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	assert.NotNil(t, config)
	assert.NotEmpty(t, config.DatabaseURL)
	assert.Equal(t, 10, config.MaxIdleConns)
	assert.Equal(t, 100, config.MaxOpenConns)
	assert.Equal(t, time.Hour, config.ConnMaxLifetime)
	assert.Equal(t, time.Minute*10, config.ConnMaxIdleTime)
	assert.Equal(t, 3, config.MaxRetries)
	assert.Equal(t, time.Second*2, config.RetryDelay)
	assert.True(t, config.AutoMigrate)
}

func TestSQLiteConnection(t *testing.T) {
	// Create a temporary directory for test database
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	config := &ConnectionConfig{
		DatabaseURL:     "sqlite://" + dbPath,
		MaxIdleConns:    5,
		MaxOpenConns:    10,
		ConnMaxLifetime: time.Minute * 30,
		ConnMaxIdleTime: time.Minute * 5,
		MaxRetries:      2,
		RetryDelay:      time.Second,
		AutoMigrate:     false, // Disable auto-migration for this test
	}

	conn := GetConnectionWithConfig(config)

	// Test connection
	err := conn.Connect()
	require.NoError(t, err)

	// Test that database file was created
	assert.FileExists(t, dbPath)

	// Test health check
	err = conn.HealthCheck()
	assert.NoError(t, err)

	// Test that we can get the DB instance
	db := conn.GetDB()
	assert.NotNil(t, db)

	// Test connection status
	assert.True(t, conn.IsConnected())

	// Test stats
	stats, err := conn.GetStats()
	assert.NoError(t, err)
	assert.NotNil(t, stats)

	// Test close
	err = conn.Close()
	assert.NoError(t, err)
	assert.False(t, conn.IsConnected())
}

func TestParseDatabaseURL(t *testing.T) {
	conn := &DatabaseConnection{config: DefaultConfig()}

	tests := []struct {
		name     string
		url      string
		wantType string
		wantDSN  string
		wantErr  bool
	}{
		{
			name:     "MySQL URL format",
			url:      "mysql://user:pass@localhost:3306/dbname",
			wantType: "mysql",
			wantDSN:  "user:pass@localhost:3306/dbname",
			wantErr:  false,
		},
		{
			name:     "MySQL DSN format",
			url:      "user:pass@tcp(localhost:3306)/dbname",
			wantType: "mysql",
			wantDSN:  "user:pass@tcp(localhost:3306)/dbname",
			wantErr:  false,
		},
		{
			name:     "SQLite URL format",
			url:      "sqlite://./test.db",
			wantType: "sqlite",
			wantDSN:  "./test.db",
			wantErr:  false,
		},
		{
			name:     "SQLite file path",
			url:      "test.db",
			wantType: "sqlite",
			wantDSN:  "test.db",
			wantErr:  false,
		},
		{
			name:     "Invalid URL",
			url:      "invalid://url",
			wantType: "",
			wantDSN:  "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotType, gotDSN, err := conn.parseDatabaseURL(tt.url)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantType, gotType)
				assert.Equal(t, tt.wantDSN, gotDSN)
			}
		})
	}
}

func TestSQLiteBackup(t *testing.T) {
	// Create temporary directories
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")
	backupPath := filepath.Join(tmpDir, "backup", "test_backup.db")

	config := &ConnectionConfig{
		DatabaseURL: "sqlite://" + dbPath,
		AutoMigrate: false,
	}

	conn := GetConnectionWithConfig(config)
	err := conn.Connect()
	require.NoError(t, err)

	// Create a simple table and insert some data
	db := conn.GetDB()
	err = db.Exec("CREATE TABLE test (id INTEGER PRIMARY KEY, name TEXT)").Error
	require.NoError(t, err)

	err = db.Exec("INSERT INTO test (name) VALUES ('test1'), ('test2')").Error
	require.NoError(t, err)

	// Test backup
	err = conn.Backup(backupPath)
	assert.NoError(t, err)

	// Verify backup file exists
	assert.FileExists(t, backupPath)

	// Verify backup is valid by connecting to it
	backupConfig := &ConnectionConfig{
		DatabaseURL: "sqlite://" + backupPath,
		AutoMigrate: false,
	}
	backupConn := GetConnectionWithConfig(backupConfig)
	err = backupConn.Connect()
	require.NoError(t, err)

	// Verify data in backup
	var count int64
	err = backupConn.GetDB().Raw("SELECT COUNT(*) FROM test").Scan(&count).Error
	assert.NoError(t, err)
	assert.Equal(t, int64(2), count)

	// Clean up
	err = conn.Close()
	assert.NoError(t, err)
	err = backupConn.Close()
	assert.NoError(t, err)
}

func TestEnvironmentVariables(t *testing.T) {
	// Test with environment variable set
	originalURL := os.Getenv("DATABASE_URL")
	defer func() {
		if originalURL != "" {
			os.Setenv("DATABASE_URL", originalURL)
		} else {
			os.Unsetenv("DATABASE_URL")
		}
	}()

	testURL := "sqlite://./env_test.db"
	os.Setenv("DATABASE_URL", testURL)

	config := DefaultConfig()
	assert.Equal(t, testURL, config.DatabaseURL)
}

func TestConnectionRetry(t *testing.T) {
	// Test with invalid database URL to trigger retry logic
	config := &ConnectionConfig{
		DatabaseURL: "mysql://invalid:invalid@invalid:9999/invalid",
		MaxRetries:  2,
		RetryDelay:  time.Millisecond * 100, // Short delay for testing
		AutoMigrate: false,
	}

	conn := GetConnectionWithConfig(config)

	start := time.Now()
	err := conn.Connect()
	duration := time.Since(start)

	// Should fail after retries
	assert.Error(t, err)

	// Should have taken at least the retry delay time
	assert.True(t, duration >= time.Millisecond*100)
}
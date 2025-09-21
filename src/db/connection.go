package db

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// ConnectionConfig holds database connection configuration
type ConnectionConfig struct {
	// Database connection URL
	DatabaseURL string
	// Connection pool settings
	MaxIdleConns    int           // Maximum number of idle connections
	MaxOpenConns    int           // Maximum number of open connections
	ConnMaxLifetime time.Duration // Maximum lifetime of a connection
	ConnMaxIdleTime time.Duration // Maximum idle time of a connection
	// Retry settings
	MaxRetries   int           // Maximum number of connection retries
	RetryDelay   time.Duration // Delay between retries
	// Logging
	LogLevel logger.LogLevel // GORM log level
	// Auto-migration
	AutoMigrate bool // Whether to run auto-migration on connect
}

// DefaultConfig returns a default database configuration
func DefaultConfig() *ConnectionConfig {
	return &ConnectionConfig{
		DatabaseURL:     getEnvOrDefault("DATABASE_URL", "sqlite://./data/etc_meisai.db"),
		MaxIdleConns:    10,
		MaxOpenConns:    100,
		ConnMaxLifetime: time.Hour,
		ConnMaxIdleTime: time.Minute * 10,
		MaxRetries:      3,
		RetryDelay:      time.Second * 2,
		LogLevel:        logger.Warn,
		AutoMigrate:     true,
	}
}

// DatabaseConnection manages the database connection
type DatabaseConnection struct {
	db     *gorm.DB
	config *ConnectionConfig
	mutex  sync.RWMutex
	once   sync.Once
}

var (
	// Global connection instance (singleton)
	instance *DatabaseConnection
	initOnce sync.Once
)

// GetConnection returns the singleton database connection
func GetConnection() *DatabaseConnection {
	initOnce.Do(func() {
		instance = &DatabaseConnection{
			config: DefaultConfig(),
		}
	})
	return instance
}

// GetConnectionWithConfig returns a database connection with custom configuration
func GetConnectionWithConfig(config *ConnectionConfig) *DatabaseConnection {
	return &DatabaseConnection{
		config: config,
	}
}

// Connect establishes a database connection with retry logic
func (dc *DatabaseConnection) Connect() error {
	dc.mutex.Lock()
	defer dc.mutex.Unlock()

	if dc.db != nil {
		return nil // Already connected
	}

	var err error
	for attempt := 1; attempt <= dc.config.MaxRetries; attempt++ {
		log.Printf("Attempting to connect to database (attempt %d/%d)", attempt, dc.config.MaxRetries)

		dc.db, err = dc.createConnection()
		if err == nil {
			if err = dc.configureConnection(); err == nil {
				log.Printf("Successfully connected to database on attempt %d", attempt)
				return nil
			}
		}

		log.Printf("Failed to connect to database (attempt %d): %v", attempt, err)

		if attempt < dc.config.MaxRetries {
			log.Printf("Retrying connection in %v...", dc.config.RetryDelay)
			time.Sleep(dc.config.RetryDelay)
		}
	}

	return fmt.Errorf("failed to connect to database after %d attempts: %w", dc.config.MaxRetries, err)
}

// createConnection creates a new GORM database connection based on the URL
func (dc *DatabaseConnection) createConnection() (*gorm.DB, error) {
	if dc.config.DatabaseURL == "" {
		return nil, fmt.Errorf("database URL is required")
	}

	// Parse the database URL
	dbType, dsn, err := dc.parseDatabaseURL(dc.config.DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse database URL: %w", err)
	}

	// Configure GORM
	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(dc.config.LogLevel),
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
	}

	// Create connection based on database type
	switch dbType {
	case "mysql":
		return gorm.Open(mysql.Open(dsn), gormConfig)
	case "sqlite":
		// Ensure directory exists for SQLite
		if err := dc.ensureSQLiteDir(dsn); err != nil {
			return nil, fmt.Errorf("failed to create SQLite directory: %w", err)
		}
		return gorm.Open(sqlite.Open(dsn), gormConfig)
	default:
		return nil, fmt.Errorf("unsupported database type: %s", dbType)
	}
}

// parseDatabaseURL parses a database URL and returns the type and DSN
func (dc *DatabaseConnection) parseDatabaseURL(url string) (string, string, error) {
	if strings.HasPrefix(url, "mysql://") {
		// Convert mysql:// URL to DSN format
		dsn := strings.TrimPrefix(url, "mysql://")
		return "mysql", dsn, nil
	} else if strings.HasPrefix(url, "sqlite://") {
		// Convert sqlite:// URL to file path
		dsn := strings.TrimPrefix(url, "sqlite://")
		return "sqlite", dsn, nil
	} else if strings.Contains(url, "@tcp(") {
		// Already in MySQL DSN format
		return "mysql", url, nil
	} else if strings.HasSuffix(url, ".db") || strings.HasSuffix(url, ".sqlite") {
		// SQLite file path
		return "sqlite", url, nil
	} else {
		return "", "", fmt.Errorf("unable to determine database type from URL: %s", url)
	}
}

// ensureSQLiteDir creates the directory for SQLite database file if it doesn't exist
func (dc *DatabaseConnection) ensureSQLiteDir(dsn string) error {
	dir := filepath.Dir(dsn)
	if dir == "." {
		return nil // Current directory, no need to create
	}

	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
		log.Printf("Created directory for SQLite database: %s", dir)
	}

	return nil
}

// configureConnection configures the database connection pool and settings
func (dc *DatabaseConnection) configureConnection() error {
	sqlDB, err := dc.db.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	// Configure connection pool
	sqlDB.SetMaxIdleConns(dc.config.MaxIdleConns)
	sqlDB.SetMaxOpenConns(dc.config.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(dc.config.ConnMaxLifetime)
	sqlDB.SetConnMaxIdleTime(dc.config.ConnMaxIdleTime)

	// Test the connection
	if err := sqlDB.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	// Apply SQLite-specific optimizations
	if err := dc.applySQLiteOptimizations(); err != nil {
		log.Printf("Warning: failed to apply SQLite optimizations: %v", err)
	}

	// Run auto-migration if enabled
	if dc.config.AutoMigrate {
		if err := dc.runAutoMigration(); err != nil {
			return fmt.Errorf("failed to run auto-migration: %w", err)
		}
	}

	log.Printf("Database connection configured successfully")
	return nil
}

// applySQLiteOptimizations applies SQLite-specific performance optimizations
func (dc *DatabaseConnection) applySQLiteOptimizations() error {
	// Check if we're using SQLite
	if dc.db.Dialector.Name() != "sqlite" {
		return nil // Not SQLite, skip optimizations
	}

	optimizations := []string{
		"PRAGMA journal_mode = WAL",        // Write-Ahead Logging mode
		"PRAGMA synchronous = NORMAL",      // Balance between safety and performance
		"PRAGMA cache_size = -32000",       // 32MB cache
		"PRAGMA foreign_keys = ON",         // Enable foreign key constraints
		"PRAGMA temp_store = MEMORY",       // Store temporary tables in memory
		"PRAGMA mmap_size = 268435456",     // 256MB memory-mapped I/O
	}

	for _, pragma := range optimizations {
		if err := dc.db.Exec(pragma).Error; err != nil {
			log.Printf("Warning: failed to execute pragma '%s': %v", pragma, err)
		}
	}

	log.Printf("Applied SQLite optimizations")
	return nil
}

// runAutoMigration runs automatic database migration
func (dc *DatabaseConnection) runAutoMigration() error {
	// Import models for auto-migration
	// Note: In a real implementation, you might want to import these from a models package
	log.Printf("Running auto-migration...")

	// This is a placeholder - actual models should be imported and migrated here
	// Example:
	// if err := dc.db.AutoMigrate(&models.ETCMeisaiRecord{}, &models.ETCMapping{}, &models.ImportSession{}); err != nil {
	//     return fmt.Errorf("auto-migration failed: %w", err)
	// }

	log.Printf("Auto-migration completed successfully")
	return nil
}

// GetDB returns the GORM database instance
func (dc *DatabaseConnection) GetDB() *gorm.DB {
	dc.mutex.RLock()
	defer dc.mutex.RUnlock()
	return dc.db
}

// IsConnected checks if the database connection is active
func (dc *DatabaseConnection) IsConnected() bool {
	dc.mutex.RLock()
	defer dc.mutex.RUnlock()

	if dc.db == nil {
		return false
	}

	sqlDB, err := dc.db.DB()
	if err != nil {
		return false
	}

	return sqlDB.Ping() == nil
}

// HealthCheck performs a comprehensive health check of the database connection
func (dc *DatabaseConnection) HealthCheck() error {
	dc.mutex.RLock()
	defer dc.mutex.RUnlock()

	if dc.db == nil {
		return fmt.Errorf("database connection is nil")
	}

	sqlDB, err := dc.db.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	// Test connection
	if err := sqlDB.Ping(); err != nil {
		return fmt.Errorf("database ping failed: %w", err)
	}

	// Check connection stats
	stats := sqlDB.Stats()
	log.Printf("Database connection stats - Open: %d, InUse: %d, Idle: %d",
		stats.OpenConnections, stats.InUse, stats.Idle)

	// Test a simple query
	var result int
	if err := dc.db.Raw("SELECT 1").Scan(&result).Error; err != nil {
		return fmt.Errorf("test query failed: %w", err)
	}

	if result != 1 {
		return fmt.Errorf("test query returned unexpected result: %d", result)
	}

	return nil
}

// Close closes the database connection
func (dc *DatabaseConnection) Close() error {
	dc.mutex.Lock()
	defer dc.mutex.Unlock()

	if dc.db == nil {
		return nil // Already closed
	}

	sqlDB, err := dc.db.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	if err := sqlDB.Close(); err != nil {
		return fmt.Errorf("failed to close database connection: %w", err)
	}

	dc.db = nil
	log.Printf("Database connection closed")
	return nil
}

// Reconnect closes and reopens the database connection
func (dc *DatabaseConnection) Reconnect() error {
	log.Printf("Reconnecting to database...")

	if err := dc.Close(); err != nil {
		log.Printf("Warning: error closing existing connection: %v", err)
	}

	return dc.Connect()
}

// WithTransaction executes a function within a database transaction
func (dc *DatabaseConnection) WithTransaction(fn func(*gorm.DB) error) error {
	return dc.db.Transaction(fn)
}

// Backup creates a backup of the database (SQLite only)
func (dc *DatabaseConnection) Backup(backupPath string) error {
	if dc.db.Dialector.Name() != "sqlite" {
		return fmt.Errorf("backup is only supported for SQLite databases")
	}

	// Ensure backup directory exists
	backupDir := filepath.Dir(backupPath)
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return fmt.Errorf("failed to create backup directory: %w", err)
	}

	// Use SQLite backup API
	backupSQL := fmt.Sprintf("VACUUM INTO '%s'", backupPath)
	if err := dc.db.Exec(backupSQL).Error; err != nil {
		return fmt.Errorf("backup failed: %w", err)
	}

	log.Printf("Database backup created: %s", backupPath)
	return nil
}

// GetStats returns database connection statistics
func (dc *DatabaseConnection) GetStats() (map[string]interface{}, error) {
	dc.mutex.RLock()
	defer dc.mutex.RUnlock()

	if dc.db == nil {
		return nil, fmt.Errorf("database connection is nil")
	}

	sqlDB, err := dc.db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	stats := sqlDB.Stats()
	return map[string]interface{}{
		"max_open_connections": stats.MaxOpenConnections,
		"open_connections":     stats.OpenConnections,
		"in_use":              stats.InUse,
		"idle":                stats.Idle,
		"wait_count":          stats.WaitCount,
		"wait_duration":       stats.WaitDuration.String(),
		"max_idle_closed":     stats.MaxIdleClosed,
		"max_idle_time_closed": stats.MaxIdleTimeClosed,
		"max_lifetime_closed":  stats.MaxLifetimeClosed,
	}, nil
}

// Helper function to get environment variable with default
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// Helper functions for easy access

// MustConnect connects to the database and panics on failure
func MustConnect() *gorm.DB {
	conn := GetConnection()
	if err := conn.Connect(); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	return conn.GetDB()
}

// MustConnectWithConfig connects with custom config and panics on failure
func MustConnectWithConfig(config *ConnectionConfig) *gorm.DB {
	conn := GetConnectionWithConfig(config)
	if err := conn.Connect(); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	return conn.GetDB()
}
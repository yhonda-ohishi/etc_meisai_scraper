package config

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// Settings holds all application configuration
type Settings struct {
	Database DatabaseSettings `json:"database"`
	Server   ServerSettings   `json:"server"`
	GRPC     GRPCSettings     `json:"grpc"`
	Scraping ScrapingSettings `json:"scraping"`
	Import   ImportSettings   `json:"import"`
	Logging  LoggingSettings  `json:"logging"`

	// Additional settings for backward compatibility with tests
	RequestTimeout         time.Duration `json:"request_timeout"`
	MaxRetries             int           `json:"max_retries"`
	RetryDelay             time.Duration `json:"retry_delay"`
	BatchSize              int           `json:"batch_size"`
	MaxConcurrentDownloads int           `json:"max_concurrent_downloads"`
	EnableProgressTracking bool          `json:"enable_progress_tracking"`
	EnableDebugLogging     bool          `json:"enable_debug_logging"`
	SessionTimeout         time.Duration `json:"session_timeout"`
	MaxFileSize            int           `json:"max_file_size"`
}

// DatabaseSettings holds database configuration
type DatabaseSettings struct {
	Driver      string `json:"driver"`
	Host        string `json:"host"`
	Port        int    `json:"port"`
	Database    string `json:"database"`
	Username    string `json:"username"`
	Password    string `json:"password"`
	MaxConn     int    `json:"max_connections"`
	MaxIdleConn int    `json:"max_idle_connections"`
	ConnTimeout int    `json:"connection_timeout"`
	Path        string `json:"path"` // For SQLite databases
}

// GRPCSettings holds gRPC client configuration
type GRPCSettings struct {
	DBServiceAddress string        `json:"db_service_address"`
	Timeout          time.Duration `json:"timeout"`
	MaxRetries       int           `json:"max_retries"`
	RetryDelay       time.Duration `json:"retry_delay"`
	EnableTLS        bool          `json:"enable_tls"`
	CertFile         string        `json:"cert_file"`
	KeyFile          string        `json:"key_file"`
	CAFile           string        `json:"ca_file"`
}

// ServerSettings holds server configuration
type ServerSettings struct {
	Host         string `json:"host"`
	Port         int    `json:"port"`
	ReadTimeout  int    `json:"read_timeout"`
	WriteTimeout int    `json:"write_timeout"`
	MaxBodySize  int64  `json:"max_body_size"`
}

// ScrapingSettings holds scraping configuration
type ScrapingSettings struct {
	MaxWorkers      int           `json:"max_workers"`
	RetryCount      int           `json:"retry_count"`
	RetryDelay      time.Duration `json:"retry_delay"`
	RequestTimeout  time.Duration `json:"request_timeout"`
	HeadlessBrowser bool          `json:"headless_browser"`
}

// ImportSettings holds import configuration
type ImportSettings struct {
	BatchSize        int    `json:"batch_size"`
	MaxFileSize      int64  `json:"max_file_size"`
	TempDir          string `json:"temp_directory"`
	AllowedFormats   []string `json:"allowed_formats"`
	DuplicateCheck   bool   `json:"duplicate_check"`
}

// LoggingSettings holds logging configuration
type LoggingSettings struct {
	Level          string `json:"level"`
	OutputPath     string `json:"output_path"`
	MaxSize        int    `json:"max_size_mb"`
	MaxBackups     int    `json:"max_backups"`
	MaxAge         int    `json:"max_age_days"`
	EnableConsole  bool   `json:"enable_console"`
	EnableJSON     bool   `json:"enable_json"`
}

// NewSettings creates settings with defaults
func NewSettings() *Settings {
	return &Settings{
		Database: DatabaseSettings{
			Driver:      "gorm",
			Host:        "localhost",
			Port:        3306,
			MaxConn:     25,
			MaxIdleConn: 5,
			ConnTimeout: 30,
			Path:        "./data/etc_meisai.db", // Default SQLite path
		},
		Server: ServerSettings{
			Host:         "0.0.0.0",
			Port:         8080,
			ReadTimeout:  30,
			WriteTimeout: 30,
			MaxBodySize:  32 << 20, // 32MB
		},
		GRPC: GRPCSettings{
			DBServiceAddress: "localhost:50051",
			Timeout:          30 * time.Second,
			MaxRetries:       3,
			RetryDelay:       1 * time.Second,
			EnableTLS:        false,
		},
		Scraping: ScrapingSettings{
			MaxWorkers:      5,
			RetryCount:      3,
			RetryDelay:      5 * time.Second,
			RequestTimeout:  30 * time.Second,
			HeadlessBrowser: true,
		},
		Import: ImportSettings{
			BatchSize:      1000,
			MaxFileSize:    100 << 20, // 100MB
			TempDir:        "./temp",
			AllowedFormats: []string{".csv", ".xlsx"},
			DuplicateCheck: true,
		},
		Logging: LoggingSettings{
			Level:         "info",
			OutputPath:    "./logs",
			MaxSize:       100,
			MaxBackups:    7,
			MaxAge:        30,
			EnableConsole: true,
			EnableJSON:    true,
		},

		// Additional settings defaults
		RequestTimeout:         5 * time.Second,
		MaxRetries:             3,
		RetryDelay:             time.Second,
		BatchSize:              1000,
		MaxConcurrentDownloads: 5,
		EnableProgressTracking: true,
		EnableDebugLogging:     false,
		SessionTimeout:         30 * time.Second,
		MaxFileSize:            100 * 1024 * 1024, // 100MB
	}
}

// LoadFromEnv loads settings from environment variables
func LoadFromEnv() *Settings {
	settings := NewSettings()

	// Database settings
	if v := os.Getenv("DB_DRIVER"); v != "" {
		settings.Database.Driver = v
	}
	if v := os.Getenv("DB_HOST"); v != "" {
		settings.Database.Host = v
	}
	if v := os.Getenv("DB_PORT"); v != "" {
		if port, err := strconv.Atoi(v); err == nil {
			settings.Database.Port = port
		}
	}
	if v := os.Getenv("DB_NAME"); v != "" {
		settings.Database.Database = v
	}
	if v := os.Getenv("DB_USER"); v != "" {
		settings.Database.Username = v
	}
	if v := os.Getenv("DB_PASSWORD"); v != "" {
		settings.Database.Password = v
	}
	if v := os.Getenv("DB_PATH"); v != "" {
		settings.Database.Path = v
	}

	// gRPC settings
	if v := os.Getenv("GRPC_DB_SERVICE_ADDRESS"); v != "" {
		settings.GRPC.DBServiceAddress = v
	}
	if v := os.Getenv("GRPC_TIMEOUT"); v != "" {
		if timeout, err := time.ParseDuration(v); err == nil {
			settings.GRPC.Timeout = timeout
		}
	}
	if v := os.Getenv("GRPC_MAX_RETRIES"); v != "" {
		if retries, err := strconv.Atoi(v); err == nil {
			settings.GRPC.MaxRetries = retries
		}
	}
	if v := os.Getenv("GRPC_ENABLE_TLS"); v != "" {
		settings.GRPC.EnableTLS = strings.ToLower(v) == "true"
	}
	if v := os.Getenv("GRPC_CERT_FILE"); v != "" {
		settings.GRPC.CertFile = v
	}
	if v := os.Getenv("GRPC_KEY_FILE"); v != "" {
		settings.GRPC.KeyFile = v
	}
	if v := os.Getenv("GRPC_CA_FILE"); v != "" {
		settings.GRPC.CAFile = v
	}

	// Server settings
	if v := os.Getenv("SERVER_HOST"); v != "" {
		settings.Server.Host = v
	}
	if v := os.Getenv("SERVER_PORT"); v != "" {
		if port, err := strconv.Atoi(v); err == nil {
			settings.Server.Port = port
		}
	}

	// Scraping settings
	if v := os.Getenv("SCRAPING_MAX_WORKERS"); v != "" {
		if workers, err := strconv.Atoi(v); err == nil {
			settings.Scraping.MaxWorkers = workers
		}
	}
	if v := os.Getenv("SCRAPING_HEADLESS"); v != "" {
		settings.Scraping.HeadlessBrowser = strings.ToLower(v) == "true"
	}

	// Import settings
	if v := os.Getenv("IMPORT_BATCH_SIZE"); v != "" {
		if size, err := strconv.Atoi(v); err == nil {
			settings.Import.BatchSize = size
		}
	}
	if v := os.Getenv("IMPORT_TEMP_DIR"); v != "" {
		settings.Import.TempDir = v
	}

	// Logging settings
	if v := os.Getenv("LOG_LEVEL"); v != "" {
		settings.Logging.Level = v
	}
	if v := os.Getenv("LOG_PATH"); v != "" {
		settings.Logging.OutputPath = v
	}
	if v := os.Getenv("LOG_JSON"); v != "" {
		settings.Logging.EnableJSON = strings.ToLower(v) == "true"
	}

	return settings
}

// LoadFromFile loads settings from a JSON file
func LoadFromFile(path string) (*Settings, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open config file: %w", err)
	}
	defer file.Close()

	settings := NewSettings()
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(settings); err != nil {
		return nil, fmt.Errorf("failed to decode config file: %w", err)
	}

	return settings, nil
}

// Validate validates all settings
func (s *Settings) Validate() error {
	// Validate test-specific fields first for early error detection
	if err := s.validateTestFields(); err != nil {
		return err
	}

	// Database validation
	if s.Database.Driver == "" {
		return fmt.Errorf("database driver is required")
	}
	if s.Database.Host == "" {
		return fmt.Errorf("database host is required")
	}
	if s.Database.Port <= 0 || s.Database.Port > 65535 {
		return fmt.Errorf("invalid database port: %d", s.Database.Port)
	}

	// Server validation
	if s.Server.Port <= 0 || s.Server.Port > 65535 {
		return fmt.Errorf("invalid server port: %d", s.Server.Port)
	}
	if s.Server.ReadTimeout <= 0 {
		s.Server.ReadTimeout = 30
	}
	if s.Server.WriteTimeout <= 0 {
		s.Server.WriteTimeout = 30
	}

	// gRPC validation
	if s.GRPC.DBServiceAddress == "" {
		return fmt.Errorf("gRPC db_service address is required")
	}
	if s.GRPC.Timeout <= 0 {
		s.GRPC.Timeout = 30 * time.Second
	}
	if s.GRPC.MaxRetries < 0 {
		s.GRPC.MaxRetries = 0
	}
	if s.GRPC.RetryDelay <= 0 {
		s.GRPC.RetryDelay = 1 * time.Second
	}

	// TLS validation
	if s.GRPC.EnableTLS {
		if s.GRPC.CertFile == "" || s.GRPC.KeyFile == "" {
			return fmt.Errorf("TLS cert_file and key_file are required when TLS is enabled")
		}
	}

	// Scraping validation
	if s.Scraping.MaxWorkers <= 0 {
		s.Scraping.MaxWorkers = 1
	}
	if s.Scraping.RetryCount < 0 {
		s.Scraping.RetryCount = 0
	}

	// Import validation
	if s.Import.BatchSize <= 0 {
		s.Import.BatchSize = 100
	}
	if s.Import.MaxFileSize <= 0 {
		s.Import.MaxFileSize = 10 << 20 // 10MB default
	}
	if s.Import.TempDir == "" {
		s.Import.TempDir = "./temp"
	}

	// Logging validation
	validLevels := []string{"debug", "info", "warn", "error", "fatal"}
	levelValid := false
	for _, level := range validLevels {
		if strings.ToLower(s.Logging.Level) == level {
			levelValid = true
			break
		}
	}
	if !levelValid {
		return fmt.Errorf("invalid log level: %s", s.Logging.Level)
	}

	return nil
}

// GetDSN returns database connection string
func (s *DatabaseSettings) GetDSN() string {
	switch s.Driver {
	case "mysql":
		return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=true&loc=Local",
			s.Username, s.Password, s.Host, s.Port, s.Database)
	case "postgres":
		return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
			s.Host, s.Port, s.Username, s.Password, s.Database)
	default:
		return ""
	}
}

// GetServerAddress returns the server address
func (s *ServerSettings) GetServerAddress() string {
	return fmt.Sprintf("%s:%d", s.Host, s.Port)
}

// GetDBServiceAddress returns the gRPC db_service address
func (g *GRPCSettings) GetDBServiceAddress() string {
	return g.DBServiceAddress
}

// IsSecure returns true if TLS is enabled
func (g *GRPCSettings) IsSecure() bool {
	return g.EnableTLS
}

// GetConnectionTimeout returns the connection timeout
func (g *GRPCSettings) GetConnectionTimeout() time.Duration {
	return g.Timeout
}

// GlobalSettings holds the global application settings
var GlobalSettings *Settings

// InitSettings initializes global settings
func InitSettings() error {
	// Try to load from file first
	if _, err := os.Stat("config.json"); err == nil {
		settings, err := LoadFromFile("config.json")
		if err != nil {
			return fmt.Errorf("failed to load config file: %w", err)
		}
		GlobalSettings = settings
	} else {
		// Load from environment variables
		GlobalSettings = LoadFromEnv()
	}

	// Validate settings
	if err := GlobalSettings.Validate(); err != nil {
		return fmt.Errorf("invalid settings: %w", err)
	}

	return nil
}

// Additional validation for test-specific fields
func (s *Settings) validateTestFields() error {
	if s.RequestTimeout <= 0 {
		return fmt.Errorf("RequestTimeout must be positive")
	}
	if s.MaxRetries < 0 {
		return fmt.Errorf("MaxRetries must be non-negative")
	}
	if s.RetryDelay <= 0 {
		return fmt.Errorf("RetryDelay must be positive")
	}
	if s.BatchSize <= 0 {
		return fmt.Errorf("BatchSize must be positive")
	}
	if s.MaxConcurrentDownloads <= 0 {
		return fmt.Errorf("MaxConcurrentDownloads must be positive")
	}
	if s.SessionTimeout <= 0 {
		return fmt.Errorf("SessionTimeout must be positive")
	}
	if s.MaxFileSize <= 0 {
		return fmt.Errorf("MaxFileSize must be positive")
	}

	// Additional validation for reasonable limits
	if s.MaxConcurrentDownloads > 10 {
		return fmt.Errorf("MaxConcurrentDownloads should not exceed 10")
	}
	if s.BatchSize > 10000 {
		return fmt.Errorf("BatchSize should not exceed 10000")
	}

	return nil
}

// Setter methods for test compatibility
func (s *Settings) SetRequestTimeout(timeout time.Duration) {
	s.RequestTimeout = timeout
}

func (s *Settings) SetMaxRetries(retries int) {
	s.MaxRetries = retries
}

func (s *Settings) SetRetryDelay(delay time.Duration) {
	s.RetryDelay = delay
}

func (s *Settings) SetBatchSize(size int) {
	s.BatchSize = size
}

func (s *Settings) SetMaxConcurrentDownloads(max int) {
	s.MaxConcurrentDownloads = max
}

func (s *Settings) SetSessionTimeout(timeout time.Duration) {
	s.SessionTimeout = timeout
}

func (s *Settings) SetMaxFileSize(size int) {
	s.MaxFileSize = size
}

func (s *Settings) SetDebugLogging(enabled bool) {
	s.EnableDebugLogging = enabled
}

func (s *Settings) SetProgressTracking(enabled bool) {
	s.EnableProgressTracking = enabled
}

// String returns a string representation of the settings
func (s *Settings) String() string {
	return fmt.Sprintf("Settings{RequestTimeout: %v, MaxRetries: %d, BatchSize: %d, EnableDebugLogging: %t}",
		s.RequestTimeout, s.MaxRetries, s.BatchSize, s.EnableDebugLogging)
}

// Clone creates a copy of the settings
func (s *Settings) Clone() *Settings {
	clone := *s
	return &clone
}

// ToMap converts settings to a map
func (s *Settings) ToMap() map[string]string {
	return map[string]string{
		"request_timeout":           s.RequestTimeout.String(),
		"max_retries":               fmt.Sprintf("%d", s.MaxRetries),
		"retry_delay":               s.RetryDelay.String(),
		"batch_size":                fmt.Sprintf("%d", s.BatchSize),
		"max_concurrent_downloads":  fmt.Sprintf("%d", s.MaxConcurrentDownloads),
		"enable_debug_logging":      fmt.Sprintf("%t", s.EnableDebugLogging),
		"enable_progress_tracking":  fmt.Sprintf("%t", s.EnableProgressTracking),
		"session_timeout":           s.SessionTimeout.String(),
		"max_file_size":             fmt.Sprintf("%d", s.MaxFileSize),
	}
}

// FromMap loads settings from a map
func (s *Settings) FromMap(settingsMap map[string]string) error {
	for key, value := range settingsMap {
		switch key {
		case "request_timeout":
			if timeout, err := time.ParseDuration(value); err != nil {
				return fmt.Errorf("invalid request_timeout: %w", err)
			} else {
				s.RequestTimeout = timeout
			}
		case "max_retries":
			if retries, err := strconv.Atoi(value); err != nil {
				return fmt.Errorf("invalid max_retries: %w", err)
			} else {
				s.MaxRetries = retries
			}
		case "retry_delay":
			if delay, err := time.ParseDuration(value); err != nil {
				return fmt.Errorf("invalid retry_delay: %w", err)
			} else {
				s.RetryDelay = delay
			}
		case "batch_size":
			if size, err := strconv.Atoi(value); err != nil {
				return fmt.Errorf("invalid batch_size: %w", err)
			} else {
				s.BatchSize = size
			}
		case "max_concurrent_downloads":
			if max, err := strconv.Atoi(value); err != nil {
				return fmt.Errorf("invalid max_concurrent_downloads: %w", err)
			} else {
				s.MaxConcurrentDownloads = max
			}
		case "enable_debug_logging":
			if enabled, err := strconv.ParseBool(value); err != nil {
				return fmt.Errorf("invalid enable_debug_logging: %w", err)
			} else {
				s.EnableDebugLogging = enabled
			}
		case "enable_progress_tracking":
			if enabled, err := strconv.ParseBool(value); err != nil {
				return fmt.Errorf("invalid enable_progress_tracking: %w", err)
			} else {
				s.EnableProgressTracking = enabled
			}
		case "session_timeout":
			if timeout, err := time.ParseDuration(value); err != nil {
				return fmt.Errorf("invalid session_timeout: %w", err)
			} else {
				s.SessionTimeout = timeout
			}
		case "max_file_size":
			if size, err := strconv.Atoi(value); err != nil {
				return fmt.Errorf("invalid max_file_size: %w", err)
			} else {
				s.MaxFileSize = size
			}
		}
	}
	return nil
}

// GetRetryBackoff returns exponential backoff duration for retry attempts
func (s *Settings) GetRetryBackoff(attempt int) time.Duration {
	if attempt == 0 {
		return s.RetryDelay
	}
	// Exponential backoff: base * 2^attempt
	multiplier := 1 << attempt // 2^attempt
	return s.RetryDelay * time.Duration(multiplier)
}

// IsMaxRetriesReached returns true if the attempt count has reached max retries
func (s *Settings) IsMaxRetriesReached(attempt int) bool {
	return attempt >= s.MaxRetries
}

// GetTimeoutContext returns a context with request timeout
func (s *Settings) GetTimeoutContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), s.RequestTimeout)
}
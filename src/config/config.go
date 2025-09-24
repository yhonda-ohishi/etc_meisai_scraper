package config

import (
	"fmt"
	"os"
	"strconv"
)

// Config holds the main application configuration
type Config struct {
	HTTPServerPort    string   `json:"http_server_port"`
	GRPCServerPort    string   `json:"grpc_server_port"`
	DatabaseURL       string   `json:"database_url"`
	DownloadDir       string   `json:"download_dir"`
	CorporateAccounts []string `json:"corporate_accounts"`
	PersonalAccounts  []string `json:"personal_accounts"`
}

// Load loads configuration from environment variables with defaults
func Load() *Config {
	return &Config{
		HTTPServerPort:    getEnvOrDefault("HTTP_SERVER_PORT", "8080"),
		GRPCServerPort:    getEnvOrDefault("GRPC_SERVER_PORT", "9090"),
		DatabaseURL:       getEnvOrDefault("DATABASE_URL", "mysql://localhost:3306/etc_meisai"),
		DownloadDir:       getEnvOrDefault("DOWNLOAD_DIR", "./downloads"),
		CorporateAccounts: ParseAccounts(os.Getenv("ETC_CORPORATE_ACCOUNTS")),
		PersonalAccounts:  ParseAccounts(os.Getenv("ETC_PERSONAL_ACCOUNTS")),
	}
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.HTTPServerPort == "" {
		return fmt.Errorf("HTTPServerPort is required")
	}
	if c.GRPCServerPort == "" {
		return fmt.Errorf("GRPCServerPort is required")
	}
	if c.DatabaseURL == "" {
		return fmt.Errorf("DatabaseURL is required")
	}
	if c.DownloadDir == "" {
		return fmt.Errorf("DownloadDir is required")
	}

	// Validate port numbers
	if err := validatePort(c.HTTPServerPort, "HTTPServerPort"); err != nil {
		return err
	}
	if err := validatePort(c.GRPCServerPort, "GRPCServerPort"); err != nil {
		return err
	}

	// Check ports are not the same
	if c.HTTPServerPort == c.GRPCServerPort {
		return fmt.Errorf("HTTPServerPort and GRPCServerPort cannot be the same")
	}

	return nil
}

// GetHTTPAddress returns the HTTP server address
func (c *Config) GetHTTPAddress() string {
	return ":" + c.HTTPServerPort
}

// GetGRPCAddress returns the gRPC server address
func (c *Config) GetGRPCAddress() string {
	return ":" + c.GRPCServerPort
}

// HasCorporateAccounts returns true if corporate accounts are configured
func (c *Config) HasCorporateAccounts() bool {
	return len(c.CorporateAccounts) > 0
}

// HasPersonalAccounts returns true if personal accounts are configured
func (c *Config) HasPersonalAccounts() bool {
	return len(c.PersonalAccounts) > 0
}

// GetTotalAccounts returns the total number of configured accounts
func (c *Config) GetTotalAccounts() int {
	return len(c.CorporateAccounts) + len(c.PersonalAccounts)
}

// String returns a string representation of the config
func (c *Config) String() string {
	return fmt.Sprintf("Config{HTTP: %s, gRPC: %s, DownloadDir: %s, Corporate: %d, Personal: %d}",
		c.HTTPServerPort, c.GRPCServerPort, c.DownloadDir, len(c.CorporateAccounts), len(c.PersonalAccounts))
}

// Helper functions

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func validatePort(portStr, name string) error {
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return fmt.Errorf("%s must be a valid port number", name)
	}
	if port < 1 || port > 65535 {
		return fmt.Errorf("%s must be between 1 and 65535", name)
	}
	return nil
}
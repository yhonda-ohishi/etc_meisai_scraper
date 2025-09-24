package config_test

import (
	"os"
	"testing"

	"github.com/yhonda-ohishi/etc_meisai/src/config"
	"github.com/yhonda-ohishi/etc_meisai/tests/helpers"
)

func TestLoad_DefaultValues(t *testing.T) {
	// Clear environment variables to test defaults
	originalVars := map[string]string{
		"ETC_CORPORATE_ACCOUNTS": os.Getenv("ETC_CORPORATE_ACCOUNTS"),
		"ETC_PERSONAL_ACCOUNTS":  os.Getenv("ETC_PERSONAL_ACCOUNTS"),
		"DATABASE_URL":           os.Getenv("DATABASE_URL"),
		"GRPC_SERVER_PORT":       os.Getenv("GRPC_SERVER_PORT"),
		"HTTP_SERVER_PORT":       os.Getenv("HTTP_SERVER_PORT"),
		"DOWNLOAD_DIR":           os.Getenv("DOWNLOAD_DIR"),
	}

	// Clear all environment variables
	for key := range originalVars {
		os.Unsetenv(key)
	}

	defer func() {
		// Restore original environment variables
		for key, value := range originalVars {
			if value != "" {
				os.Setenv(key, value)
			}
		}
	}()

	cfg := config.Load()

	// Test default values
	helpers.AssertEqual(t, "8080", cfg.HTTPServerPort)
	helpers.AssertEqual(t, "9090", cfg.GRPCServerPort)
	helpers.AssertEqual(t, "./downloads", cfg.DownloadDir)
	helpers.AssertEqual(t, "mysql://localhost:3306/etc_meisai", cfg.DatabaseURL)
	helpers.AssertLen(t, cfg.CorporateAccounts, 0)
	helpers.AssertLen(t, cfg.PersonalAccounts, 0)
}

func TestLoad_EnvironmentVariables(t *testing.T) {
	// Set test environment variables
	testVars := map[string]string{
		"ETC_CORPORATE_ACCOUNTS": "corp1,corp2,corp3",
		"ETC_PERSONAL_ACCOUNTS":  "personal1,personal2",
		"DATABASE_URL":           "mysql://testdb:3306/test",
		"GRPC_SERVER_PORT":       "9999",
		"HTTP_SERVER_PORT":       "8888",
		"DOWNLOAD_DIR":           "/tmp/test-downloads",
	}

	// Set environment variables
	for key, value := range testVars {
		os.Setenv(key, value)
	}

	defer func() {
		// Clean up environment variables
		for key := range testVars {
			os.Unsetenv(key)
		}
	}()

	cfg := config.Load()

	// Test environment variable values
	helpers.AssertEqual(t, "8888", cfg.HTTPServerPort)
	helpers.AssertEqual(t, "9999", cfg.GRPCServerPort)
	helpers.AssertEqual(t, "/tmp/test-downloads", cfg.DownloadDir)
	helpers.AssertEqual(t, "mysql://testdb:3306/test", cfg.DatabaseURL)

	helpers.AssertLen(t, cfg.CorporateAccounts, 3)
	helpers.AssertEqual(t, "corp1", cfg.CorporateAccounts[0])
	helpers.AssertEqual(t, "corp2", cfg.CorporateAccounts[1])
	helpers.AssertEqual(t, "corp3", cfg.CorporateAccounts[2])

	helpers.AssertLen(t, cfg.PersonalAccounts, 2)
	helpers.AssertEqual(t, "personal1", cfg.PersonalAccounts[0])
	helpers.AssertEqual(t, "personal2", cfg.PersonalAccounts[1])
}

func TestLoad_EmptyAccountLists(t *testing.T) {
	// Set empty account lists
	os.Setenv("ETC_CORPORATE_ACCOUNTS", "")
	os.Setenv("ETC_PERSONAL_ACCOUNTS", "")

	defer func() {
		os.Unsetenv("ETC_CORPORATE_ACCOUNTS")
		os.Unsetenv("ETC_PERSONAL_ACCOUNTS")
	}()

	cfg := config.Load()

	helpers.AssertLen(t, cfg.CorporateAccounts, 0)
	helpers.AssertLen(t, cfg.PersonalAccounts, 0)
}

func TestLoad_WhitespaceInAccounts(t *testing.T) {
	// Set accounts with whitespace
	os.Setenv("ETC_CORPORATE_ACCOUNTS", " corp1 , corp2 , corp3 ")
	os.Setenv("ETC_PERSONAL_ACCOUNTS", " personal1 , personal2 ")

	defer func() {
		os.Unsetenv("ETC_CORPORATE_ACCOUNTS")
		os.Unsetenv("ETC_PERSONAL_ACCOUNTS")
	}()

	cfg := config.Load()

	// Should trim whitespace
	helpers.AssertLen(t, cfg.CorporateAccounts, 3)
	helpers.AssertEqual(t, "corp1", cfg.CorporateAccounts[0])
	helpers.AssertEqual(t, "corp2", cfg.CorporateAccounts[1])
	helpers.AssertEqual(t, "corp3", cfg.CorporateAccounts[2])

	helpers.AssertLen(t, cfg.PersonalAccounts, 2)
	helpers.AssertEqual(t, "personal1", cfg.PersonalAccounts[0])
	helpers.AssertEqual(t, "personal2", cfg.PersonalAccounts[1])
}

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  *config.Config
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid config",
			config: &config.Config{
				HTTPServerPort: "8080",
				GRPCServerPort: "9090",
				DatabaseURL:    "mysql://localhost:3306/etc_meisai",
				DownloadDir:    "./downloads",
			},
			wantErr: false,
		},
		{
			name: "empty HTTP port",
			config: &config.Config{
				HTTPServerPort: "",
				GRPCServerPort: "9090",
				DatabaseURL:    "mysql://localhost:3306/etc_meisai",
				DownloadDir:    "./downloads",
			},
			wantErr: true,
			errMsg:  "HTTPServerPort is required",
		},
		{
			name: "empty gRPC port",
			config: &config.Config{
				HTTPServerPort: "8080",
				GRPCServerPort: "",
				DatabaseURL:    "mysql://localhost:3306/etc_meisai",
				DownloadDir:    "./downloads",
			},
			wantErr: true,
			errMsg:  "GRPCServerPort is required",
		},
		{
			name: "empty database URL",
			config: &config.Config{
				HTTPServerPort: "8080",
				GRPCServerPort: "9090",
				DatabaseURL:    "",
				DownloadDir:    "./downloads",
			},
			wantErr: true,
			errMsg:  "DatabaseURL is required",
		},
		{
			name: "empty download directory",
			config: &config.Config{
				HTTPServerPort: "8080",
				GRPCServerPort: "9090",
				DatabaseURL:    "mysql://localhost:3306/etc_meisai",
				DownloadDir:    "",
			},
			wantErr: true,
			errMsg:  "DownloadDir is required",
		},
		{
			name: "invalid HTTP port",
			config: &config.Config{
				HTTPServerPort: "invalid",
				GRPCServerPort: "9090",
				DatabaseURL:    "mysql://localhost:3306/etc_meisai",
				DownloadDir:    "./downloads",
			},
			wantErr: true,
			errMsg:  "HTTPServerPort must be a valid port number",
		},
		{
			name: "invalid gRPC port",
			config: &config.Config{
				HTTPServerPort: "8080",
				GRPCServerPort: "invalid",
				DatabaseURL:    "mysql://localhost:3306/etc_meisai",
				DownloadDir:    "./downloads",
			},
			wantErr: true,
			errMsg:  "GRPCServerPort must be a valid port number",
		},
		{
			name: "port out of range",
			config: &config.Config{
				HTTPServerPort: "99999",
				GRPCServerPort: "9090",
				DatabaseURL:    "mysql://localhost:3306/etc_meisai",
				DownloadDir:    "./downloads",
			},
			wantErr: true,
			errMsg:  "HTTPServerPort must be between 1 and 65535",
		},
		{
			name: "same HTTP and gRPC ports",
			config: &config.Config{
				HTTPServerPort: "8080",
				GRPCServerPort: "8080",
				DatabaseURL:    "mysql://localhost:3306/etc_meisai",
				DownloadDir:    "./downloads",
			},
			wantErr: true,
			errMsg:  "HTTPServerPort and GRPCServerPort cannot be the same",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()

			if tt.wantErr {
				helpers.AssertError(t, err)
				if tt.errMsg != "" {
					helpers.AssertContains(t, err.Error(), tt.errMsg)
				}
			} else {
				helpers.AssertNoError(t, err)
			}
		})
	}
}

func TestConfig_GetHTTPAddress(t *testing.T) {
	cfg := &config.Config{
		HTTPServerPort: "8080",
	}

	address := cfg.GetHTTPAddress()
	helpers.AssertEqual(t, ":8080", address)
}

func TestConfig_GetGRPCAddress(t *testing.T) {
	cfg := &config.Config{
		GRPCServerPort: "9090",
	}

	address := cfg.GetGRPCAddress()
	helpers.AssertEqual(t, ":9090", address)
}

func TestConfig_HasCorporateAccounts(t *testing.T) {
	tests := []struct {
		name     string
		accounts []string
		expected bool
	}{
		{
			name:     "has corporate accounts",
			accounts: []string{"corp1", "corp2"},
			expected: true,
		},
		{
			name:     "no corporate accounts",
			accounts: []string{},
			expected: false,
		},
		{
			name:     "nil corporate accounts",
			accounts: nil,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{
				CorporateAccounts: tt.accounts,
			}
			result := cfg.HasCorporateAccounts()
			helpers.AssertEqual(t, tt.expected, result)
		})
	}
}

func TestConfig_HasPersonalAccounts(t *testing.T) {
	tests := []struct {
		name     string
		accounts []string
		expected bool
	}{
		{
			name:     "has personal accounts",
			accounts: []string{"personal1", "personal2"},
			expected: true,
		},
		{
			name:     "no personal accounts",
			accounts: []string{},
			expected: false,
		},
		{
			name:     "nil personal accounts",
			accounts: nil,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{
				PersonalAccounts: tt.accounts,
			}
			result := cfg.HasPersonalAccounts()
			helpers.AssertEqual(t, tt.expected, result)
		})
	}
}

func TestConfig_GetTotalAccounts(t *testing.T) {
	cfg := &config.Config{
		CorporateAccounts: []string{"corp1", "corp2", "corp3"},
		PersonalAccounts:  []string{"personal1", "personal2"},
	}

	total := cfg.GetTotalAccounts()
	helpers.AssertEqual(t, 5, total)
}

func TestConfig_String(t *testing.T) {
	cfg := &config.Config{
		HTTPServerPort:    "8080",
		GRPCServerPort:    "9090",
		DatabaseURL:       "mysql://localhost:3306/etc_meisai",
		DownloadDir:       "./downloads",
		CorporateAccounts: []string{"corp1", "corp2"},
		PersonalAccounts:  []string{"personal1"},
	}

	str := cfg.String()

	// Should contain key configuration information
	helpers.AssertContains(t, str, "8080")       // HTTP Port
	helpers.AssertContains(t, str, "9090")       // gRPC Port
	helpers.AssertContains(t, str, "downloads")  // Download Dir
	helpers.AssertContains(t, str, "2")          // Corporate account count
	helpers.AssertContains(t, str, "1")          // Personal account count
}
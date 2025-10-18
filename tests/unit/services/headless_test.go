package services_test

import (
	"os"
	"testing"

	"github.com/yhonda-ohishi/etc_meisai_scraper/src/services"
)

func TestGetHeadlessMode(t *testing.T) {
	tests := []struct {
		name     string
		envValue string
		setEnv   bool
		expected bool
	}{
		{
			name:     "default (env not set)",
			setEnv:   false,
			expected: true,
		},
		{
			name:     "explicitly true",
			envValue: "true",
			setEnv:   true,
			expected: true,
		},
		{
			name:     "explicitly false",
			envValue: "false",
			setEnv:   true,
			expected: false,
		},
		{
			name:     "value '1' (true)",
			envValue: "1",
			setEnv:   true,
			expected: true,
		},
		{
			name:     "value '0' (false)",
			envValue: "0",
			setEnv:   true,
			expected: false,
		},
		{
			name:     "invalid value defaults to true",
			envValue: "invalid",
			setEnv:   true,
			expected: true,
		},
		{
			name:     "empty string defaults to true",
			envValue: "",
			setEnv:   true,
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			oldValue := os.Getenv("ETC_HEADLESS")
			defer func() {
				if oldValue != "" {
					os.Setenv("ETC_HEADLESS", oldValue)
				} else {
					os.Unsetenv("ETC_HEADLESS")
				}
			}()

			if tt.setEnv {
				os.Setenv("ETC_HEADLESS", tt.envValue)
			} else {
				os.Unsetenv("ETC_HEADLESS")
			}

			// Execute
			result := services.GetHeadlessMode()

			// Verify
			if result != tt.expected {
				t.Errorf("GetHeadlessMode() = %v, want %v", result, tt.expected)
			}
		})
	}
}

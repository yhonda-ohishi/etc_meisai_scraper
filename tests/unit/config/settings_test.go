package config_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/yhonda-ohishi/etc_meisai/src/config"
	"github.com/yhonda-ohishi/etc_meisai/tests/helpers"
)

func TestSettings_Default(t *testing.T) {
	settings := config.NewSettings()

	// Test default values
	helpers.AssertEqual(t, 5*time.Second, settings.RequestTimeout)
	helpers.AssertEqual(t, 3, settings.MaxRetries)
	helpers.AssertEqual(t, time.Second, settings.RetryDelay)
	helpers.AssertEqual(t, 1000, settings.BatchSize)
	helpers.AssertEqual(t, 5, settings.MaxConcurrentDownloads)
	helpers.AssertTrue(t, settings.EnableProgressTracking)
	helpers.AssertFalse(t, settings.EnableDebugLogging)
	helpers.AssertEqual(t, 30*time.Second, settings.SessionTimeout)
	helpers.AssertEqual(t, 100*1024*1024, settings.MaxFileSize) // 100MB
}

func TestSettings_Validate(t *testing.T) {
	tests := []struct {
		name     string
		settings *config.Settings
		wantErr  bool
		errMsg   string
	}{
		{
			name:     "valid settings",
			settings: config.NewSettings(),
			wantErr:  false,
		},
		{
			name: "zero request timeout",
			settings: &config.Settings{
				RequestTimeout:         0,
				MaxRetries:             3,
				RetryDelay:             time.Second,
				BatchSize:              1000,
				MaxConcurrentDownloads: 5,
				SessionTimeout:         30 * time.Second,
				MaxFileSize:            100 * 1024 * 1024,
			},
			wantErr: true,
			errMsg:  "RequestTimeout must be positive",
		},
		{
			name: "negative max retries",
			settings: &config.Settings{
				RequestTimeout:         5 * time.Second,
				MaxRetries:             -1,
				RetryDelay:             time.Second,
				BatchSize:              1000,
				MaxConcurrentDownloads: 5,
				SessionTimeout:         30 * time.Second,
				MaxFileSize:            100 * 1024 * 1024,
			},
			wantErr: true,
			errMsg:  "MaxRetries must be non-negative",
		},
		{
			name: "zero retry delay",
			settings: &config.Settings{
				RequestTimeout:         5 * time.Second,
				MaxRetries:             3,
				RetryDelay:             0,
				BatchSize:              1000,
				MaxConcurrentDownloads: 5,
				SessionTimeout:         30 * time.Second,
				MaxFileSize:            100 * 1024 * 1024,
			},
			wantErr: true,
			errMsg:  "RetryDelay must be positive",
		},
		{
			name: "zero batch size",
			settings: &config.Settings{
				RequestTimeout:         5 * time.Second,
				MaxRetries:             3,
				RetryDelay:             time.Second,
				BatchSize:              0,
				MaxConcurrentDownloads: 5,
				SessionTimeout:         30 * time.Second,
				MaxFileSize:            100 * 1024 * 1024,
			},
			wantErr: true,
			errMsg:  "BatchSize must be positive",
		},
		{
			name: "zero max concurrent downloads",
			settings: &config.Settings{
				RequestTimeout:         5 * time.Second,
				MaxRetries:             3,
				RetryDelay:             time.Second,
				BatchSize:              1000,
				MaxConcurrentDownloads: 0,
				SessionTimeout:         30 * time.Second,
				MaxFileSize:            100 * 1024 * 1024,
			},
			wantErr: true,
			errMsg:  "MaxConcurrentDownloads must be positive",
		},
		{
			name: "zero session timeout",
			settings: &config.Settings{
				RequestTimeout:         5 * time.Second,
				MaxRetries:             3,
				RetryDelay:             time.Second,
				BatchSize:              1000,
				MaxConcurrentDownloads: 5,
				SessionTimeout:         0,
				MaxFileSize:            100 * 1024 * 1024,
			},
			wantErr: true,
			errMsg:  "SessionTimeout must be positive",
		},
		{
			name: "zero max file size",
			settings: &config.Settings{
				RequestTimeout:         5 * time.Second,
				MaxRetries:             3,
				RetryDelay:             time.Second,
				BatchSize:              1000,
				MaxConcurrentDownloads: 5,
				SessionTimeout:         30 * time.Second,
				MaxFileSize:            0,
			},
			wantErr: true,
			errMsg:  "MaxFileSize must be positive",
		},
		{
			name: "excessive max concurrent downloads",
			settings: &config.Settings{
				RequestTimeout:         5 * time.Second,
				MaxRetries:             3,
				RetryDelay:             time.Second,
				BatchSize:              1000,
				MaxConcurrentDownloads: 100,
				SessionTimeout:         30 * time.Second,
				MaxFileSize:            100 * 1024 * 1024,
			},
			wantErr: true,
			errMsg:  "MaxConcurrentDownloads should not exceed 10",
		},
		{
			name: "excessive batch size",
			settings: &config.Settings{
				RequestTimeout:         5 * time.Second,
				MaxRetries:             3,
				RetryDelay:             time.Second,
				BatchSize:              100000,
				MaxConcurrentDownloads: 5,
				SessionTimeout:         30 * time.Second,
				MaxFileSize:            100 * 1024 * 1024,
			},
			wantErr: true,
			errMsg:  "BatchSize should not exceed 10000",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.settings.Validate()

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

func TestSettings_SetRequestTimeout(t *testing.T) {
	settings := config.NewSettings()

	settings.SetRequestTimeout(10 * time.Second)
	helpers.AssertEqual(t, 10*time.Second, settings.RequestTimeout)
}

func TestSettings_SetMaxRetries(t *testing.T) {
	settings := config.NewSettings()

	settings.SetMaxRetries(5)
	helpers.AssertEqual(t, 5, settings.MaxRetries)
}

func TestSettings_SetRetryDelay(t *testing.T) {
	settings := config.NewSettings()

	settings.SetRetryDelay(2 * time.Second)
	helpers.AssertEqual(t, 2*time.Second, settings.RetryDelay)
}

func TestSettings_SetBatchSize(t *testing.T) {
	settings := config.NewSettings()

	settings.SetBatchSize(2000)
	helpers.AssertEqual(t, 2000, settings.BatchSize)
}

func TestSettings_SetMaxConcurrentDownloads(t *testing.T) {
	settings := config.NewSettings()

	settings.SetMaxConcurrentDownloads(8)
	helpers.AssertEqual(t, 8, settings.MaxConcurrentDownloads)
}

func TestSettings_SetSessionTimeout(t *testing.T) {
	settings := config.NewSettings()

	settings.SetSessionTimeout(60 * time.Second)
	helpers.AssertEqual(t, 60*time.Second, settings.SessionTimeout)
}

func TestSettings_SetMaxFileSize(t *testing.T) {
	settings := config.NewSettings()

	settings.SetMaxFileSize(200 * 1024 * 1024) // 200MB
	helpers.AssertEqual(t, 200*1024*1024, settings.MaxFileSize)
}

func TestSettings_EnableDebugLogging(t *testing.T) {
	settings := config.NewSettings()

	helpers.AssertFalse(t, settings.EnableDebugLogging)

	settings.SetDebugLogging(true)
	helpers.AssertTrue(t, settings.EnableDebugLogging)

	settings.SetDebugLogging(false)
	helpers.AssertFalse(t, settings.EnableDebugLogging)
}

func TestSettings_EnableProgressTracking(t *testing.T) {
	settings := config.NewSettings()

	helpers.AssertTrue(t, settings.EnableProgressTracking)

	settings.SetProgressTracking(false)
	helpers.AssertFalse(t, settings.EnableProgressTracking)

	settings.SetProgressTracking(true)
	helpers.AssertTrue(t, settings.EnableProgressTracking)
}

func TestSettings_String(t *testing.T) {
	settings := config.NewSettings()
	settings.SetDebugLogging(true)
	settings.SetMaxRetries(5)

	str := settings.String()

	// Should contain key settings information
	helpers.AssertContains(t, str, "5s")    // RequestTimeout
	helpers.AssertContains(t, str, "5")     // MaxRetries
	helpers.AssertContains(t, str, "1000")  // BatchSize
	helpers.AssertContains(t, str, "true")  // EnableDebugLogging
}

func TestSettings_Clone(t *testing.T) {
	original := config.NewSettings()
	original.SetRequestTimeout(10 * time.Second)
	original.SetMaxRetries(5)
	original.SetDebugLogging(true)

	cloned := original.Clone()

	// Should have same values
	helpers.AssertEqual(t, original.RequestTimeout, cloned.RequestTimeout)
	helpers.AssertEqual(t, original.MaxRetries, cloned.MaxRetries)
	helpers.AssertEqual(t, original.EnableDebugLogging, cloned.EnableDebugLogging)

	// Should be independent objects
	cloned.SetRequestTimeout(20 * time.Second)
	helpers.AssertNotEqual(t, original.RequestTimeout, cloned.RequestTimeout)
}

func TestSettings_ToMap(t *testing.T) {
	settings := config.NewSettings()
	settings.SetRequestTimeout(10 * time.Second)
	settings.SetMaxRetries(5)
	settings.SetDebugLogging(true)

	settingsMap := settings.ToMap()

	helpers.AssertEqual(t, "10s", settingsMap["request_timeout"])
	helpers.AssertEqual(t, "5", settingsMap["max_retries"])
	helpers.AssertEqual(t, "true", settingsMap["enable_debug_logging"])
	helpers.AssertEqual(t, "true", settingsMap["enable_progress_tracking"])
}

func TestSettings_FromMap(t *testing.T) {
	settingsMap := map[string]string{
		"request_timeout":           "15s",
		"max_retries":               "7",
		"retry_delay":               "2s",
		"batch_size":                "2000",
		"max_concurrent_downloads":  "8",
		"enable_debug_logging":      "true",
		"enable_progress_tracking":  "false",
		"session_timeout":           "60s",
		"max_file_size":             "209715200", // 200MB
	}

	settings := config.NewSettings()
	err := settings.FromMap(settingsMap)
	helpers.AssertNoError(t, err)

	helpers.AssertEqual(t, 15*time.Second, settings.RequestTimeout)
	helpers.AssertEqual(t, 7, settings.MaxRetries)
	helpers.AssertEqual(t, 2*time.Second, settings.RetryDelay)
	helpers.AssertEqual(t, 2000, settings.BatchSize)
	helpers.AssertEqual(t, 8, settings.MaxConcurrentDownloads)
	helpers.AssertTrue(t, settings.EnableDebugLogging)
	helpers.AssertFalse(t, settings.EnableProgressTracking)
	helpers.AssertEqual(t, 60*time.Second, settings.SessionTimeout)
	helpers.AssertEqual(t, 200*1024*1024, settings.MaxFileSize)
}

func TestSettings_FromMap_InvalidValues(t *testing.T) {
	tests := []struct {
		name        string
		settingsMap map[string]string
		wantErr     bool
		errMsg      string
	}{
		{
			name: "invalid request timeout",
			settingsMap: map[string]string{
				"request_timeout": "invalid",
			},
			wantErr: true,
			errMsg:  "invalid request_timeout",
		},
		{
			name: "invalid max retries",
			settingsMap: map[string]string{
				"max_retries": "invalid",
			},
			wantErr: true,
			errMsg:  "invalid max_retries",
		},
		{
			name: "invalid batch size",
			settingsMap: map[string]string{
				"batch_size": "invalid",
			},
			wantErr: true,
			errMsg:  "invalid batch_size",
		},
		{
			name: "invalid boolean",
			settingsMap: map[string]string{
				"enable_debug_logging": "invalid",
			},
			wantErr: true,
			errMsg:  "invalid enable_debug_logging",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			settings := config.NewSettings()
			err := settings.FromMap(tt.settingsMap)

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

func TestSettings_GetRetryBackoff(t *testing.T) {
	settings := config.NewSettings()
	settings.SetRetryDelay(time.Second)

	// Test exponential backoff
	tests := []struct {
		attempt  int
		expected time.Duration
	}{
		{0, time.Second},
		{1, 2 * time.Second},
		{2, 4 * time.Second},
		{3, 8 * time.Second},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("attempt_%d", tt.attempt), func(t *testing.T) {
			backoff := settings.GetRetryBackoff(tt.attempt)
			helpers.AssertEqual(t, tt.expected, backoff)
		})
	}
}

func TestSettings_IsMaxRetriesReached(t *testing.T) {
	settings := config.NewSettings()
	settings.SetMaxRetries(3)

	helpers.AssertFalse(t, settings.IsMaxRetriesReached(0))
	helpers.AssertFalse(t, settings.IsMaxRetriesReached(1))
	helpers.AssertFalse(t, settings.IsMaxRetriesReached(2))
	helpers.AssertTrue(t, settings.IsMaxRetriesReached(3))
	helpers.AssertTrue(t, settings.IsMaxRetriesReached(4))
}

func TestSettings_GetTimeoutContext(t *testing.T) {
	settings := config.NewSettings()
	settings.SetRequestTimeout(100 * time.Millisecond)

	ctx, cancel := settings.GetTimeoutContext()
	defer cancel()

	select {
	case <-ctx.Done():
		t.Error("Context should not be done immediately")
	default:
		// Expected
	}

	// Wait for timeout
	time.Sleep(150 * time.Millisecond)

	select {
	case <-ctx.Done():
		// Expected
	default:
		t.Error("Context should be done after timeout")
	}
}
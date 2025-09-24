package services_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/yhonda-ohishi/etc_meisai/src/services"
	"github.com/yhonda-ohishi/etc_meisai/tests/helpers"
	"github.com/yhonda-ohishi/etc_meisai/tests/mocks"
)

func TestBaseService_NewBaseService(t *testing.T) {
	mockRegistry := mocks.NewMockRegistry()
	baseService := services.NewBaseService(mockRegistry)

	helpers.AssertNotNil(t, baseService)
}

func TestBaseService_GetContext(t *testing.T) {
	mockRegistry := mocks.NewMockRegistry()
	baseService := services.NewBaseService(mockRegistry)

	ctx := baseService.GetContext()
	helpers.AssertNotNil(t, ctx)
	helpers.AssertEqual(t, context.Background(), ctx)
}

func TestBaseService_GetContextWithTimeout(t *testing.T) {
	mockRegistry := mocks.NewMockRegistry()
	baseService := services.NewBaseService(mockRegistry)

	timeout := 5 * time.Second
	ctx, cancel := baseService.GetContextWithTimeout(timeout)
	defer cancel()

	helpers.AssertNotNil(t, ctx)
	helpers.AssertNotNil(t, cancel)

	// Check that context has a deadline
	deadline, ok := ctx.Deadline()
	helpers.AssertTrue(t, ok)
	helpers.AssertTrue(t, deadline.After(time.Now()))
}

func TestBaseService_GetContextWithCancel(t *testing.T) {
	mockRegistry := mocks.NewMockRegistry()
	baseService := services.NewBaseService(mockRegistry)

	ctx, cancel := baseService.GetContextWithCancel()
	defer cancel()

	helpers.AssertNotNil(t, ctx)
	helpers.AssertNotNil(t, cancel)

	// Context should not be cancelled initially
	select {
	case <-ctx.Done():
		t.Error("Context should not be cancelled initially")
	default:
		// Expected
	}

	// Cancel and verify
	cancel()

	select {
	case <-ctx.Done():
		// Expected
	case <-time.After(100 * time.Millisecond):
		t.Error("Context should be cancelled after calling cancel()")
	}
}

func TestBaseService_ValidateInput(t *testing.T) {
	mockRegistry := mocks.NewMockRegistry()
	baseService := services.NewBaseService(mockRegistry)

	tests := []struct {
		name    string
		input   interface{}
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid string input",
			input:   "test string",
			wantErr: false,
		},
		{
			name:    "valid struct input",
			input:   struct{ Name string }{Name: "test"},
			wantErr: false,
		},
		{
			name:    "nil input",
			input:   nil,
			wantErr: true,
			errMsg:  "input cannot be nil",
		},
		{
			name:    "empty string",
			input:   "",
			wantErr: true,
			errMsg:  "input cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := baseService.ValidateInput(tt.input)

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

func TestBaseService_HandleError(t *testing.T) {
	mockRegistry := mocks.NewMockRegistry()
	baseService := services.NewBaseService(mockRegistry)

	tests := []struct {
		name       string
		err        error
		operation  string
		shouldWrap bool
	}{
		{
			name:       "nil error",
			err:        nil,
			operation:  "test operation",
			shouldWrap: false,
		},
		{
			name:       "existing error",
			err:        assert.AnError,
			operation:  "test operation",
			shouldWrap: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := baseService.HandleError(tt.err, tt.operation)

			if tt.shouldWrap {
				helpers.AssertError(t, result)
				helpers.AssertContains(t, result.Error(), tt.operation)
			} else {
				helpers.AssertNoError(t, result)
			}
		})
	}
}

func TestBaseService_LogOperation(t *testing.T) {
	mockRegistry := mocks.NewMockRegistry()
	baseService := services.NewBaseService(mockRegistry)

	// Test that LogOperation doesn't panic
	baseService.LogOperation("test operation", "test details")
	baseService.LogOperation("test operation", nil)
	baseService.LogOperation("", "test details")
}

func TestBaseService_GetLogger(t *testing.T) {
	mockRegistry := mocks.NewMockRegistry()
	baseService := services.NewBaseService(mockRegistry)

	logger := baseService.GetLogger()
	helpers.AssertNotNil(t, logger)
}

func TestBaseService_GetMetrics(t *testing.T) {
	mockRegistry := mocks.NewMockRegistry()
	baseService := services.NewBaseService(mockRegistry)

	metrics := baseService.GetMetrics()
	helpers.AssertNotNil(t, metrics)
}

func TestBaseService_RecordMetric(t *testing.T) {
	mockRegistry := mocks.NewMockRegistry()
	baseService := services.NewBaseService(mockRegistry)

	// Test that RecordMetric doesn't panic
	baseService.RecordMetric("test.metric", 1.0)
	baseService.RecordMetric("test.counter", 5)
	baseService.RecordMetric("", 0)
}

func TestBaseService_StartTransaction(t *testing.T) {
	mockRegistry := mocks.NewMockRegistry()
	baseService := services.NewBaseService(mockRegistry)

	ctx := context.Background()
	txCtx, commit, rollback := baseService.StartTransaction(ctx)

	helpers.AssertNotNil(t, txCtx)
	helpers.AssertNotNil(t, commit)
	helpers.AssertNotNil(t, rollback)

	// Test commit
	err := commit()
	helpers.AssertNoError(t, err)

	// Test rollback (should be safe to call after commit)
	err = rollback()
	helpers.AssertNoError(t, err)
}

func TestBaseService_WithRetry(t *testing.T) {
	mockRegistry := mocks.NewMockRegistry()
	baseService := services.NewBaseService(mockRegistry)

	tests := []struct {
		name        string
		operation   func() error
		maxRetries  int
		expectError bool
	}{
		{
			name: "successful operation",
			operation: func() error {
				return nil
			},
			maxRetries:  3,
			expectError: false,
		},
		{
			name: "operation fails all retries",
			operation: func() error {
				return assert.AnError
			},
			maxRetries:  3,
			expectError: true,
		},
		{
			name: "operation succeeds on second try",
			operation: func() func() error {
				attempts := 0
				return func() error {
					attempts++
					if attempts == 1 {
						return assert.AnError
					}
					return nil
				}
			}(),
			maxRetries:  3,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := baseService.WithRetry(tt.operation, tt.maxRetries)

			if tt.expectError {
				helpers.AssertError(t, err)
			} else {
				helpers.AssertNoError(t, err)
			}
		})
	}
}

func TestBaseService_GetConfig(t *testing.T) {
	mockRegistry := mocks.NewMockRegistry()
	baseService := services.NewBaseService(mockRegistry)

	config := baseService.GetConfig()
	helpers.AssertNotNil(t, config)
}

func TestBaseService_IsHealthy(t *testing.T) {
	mockRegistry := mocks.NewMockRegistry()
	baseService := services.NewBaseService(mockRegistry)

	healthy := baseService.IsHealthy()
	helpers.AssertTrue(t, healthy) // Should be healthy by default
}

func TestBaseService_GetStatus(t *testing.T) {
	mockRegistry := mocks.NewMockRegistry()
	baseService := services.NewBaseService(mockRegistry)

	status := baseService.GetStatus()
	helpers.AssertNotNil(t, status)
	helpers.AssertEqual(t, "running", status.State)
}

func TestBaseService_Shutdown(t *testing.T) {
	mockRegistry := mocks.NewMockRegistry()
	baseService := services.NewBaseService(mockRegistry)

	ctx := context.Background()
	err := baseService.Shutdown(ctx)
	helpers.AssertNoError(t, err)

	// After shutdown, service should not be healthy
	healthy := baseService.IsHealthy()
	helpers.AssertFalse(t, healthy)
}

func TestBaseService_ConcurrentAccess(t *testing.T) {
	mockRegistry := mocks.NewMockRegistry()
	baseService := services.NewBaseService(mockRegistry)

	// Test concurrent access to service methods
	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func() {
			defer func() { done <- true }()

			// Test multiple concurrent operations
			baseService.GetContext()
			baseService.IsHealthy()
			baseService.GetStatus()
			baseService.LogOperation("concurrent test", "test data")
			baseService.RecordMetric("concurrent.test", 1.0)
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	// Service should still be healthy after concurrent access
	helpers.AssertTrue(t, baseService.IsHealthy())
}
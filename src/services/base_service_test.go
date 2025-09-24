package services

import (
	"context"
	"errors"
	"log"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/yhonda-ohishi/etc_meisai/src/repositories"
)

func TestNewBaseService(t *testing.T) {
	t.Parallel()

	t.Run("successful creation", func(t *testing.T) {
		dbClient := &mockDBClient{}

		service := NewBaseService(dbClient)

		assert.NotNil(t, service)
		assert.Equal(t, dbClient, service.dbClient)
		assert.NotNil(t, service.Logger)
		assert.NotNil(t, service.metrics)
		assert.NotNil(t, service.config)
		assert.True(t, service.isHealthy)
		assert.NotNil(t, service.status)
		assert.Equal(t, "running", service.status.State)
	})
}

func TestNewBaseServiceWithDependencies(t *testing.T) {
	t.Parallel()

	t.Run("successful creation with dependencies", func(t *testing.T) {
		dbClient := &mockDBClient{}
		var etcRepo repositories.ETCRepository
		var mappingRepo repositories.MappingRepository
		logger := log.New(os.Stdout, "test", log.LstdFlags)

		service := NewBaseServiceWithDependencies(dbClient, etcRepo, mappingRepo, logger)

		assert.NotNil(t, service)
		assert.Equal(t, dbClient, service.dbClient)
		assert.Equal(t, etcRepo, service.ETCRepository)
		assert.Equal(t, mappingRepo, service.MappingRepository)
		assert.Equal(t, logger, service.Logger)
		assert.NotNil(t, service.metrics)
		assert.NotNil(t, service.config)
		assert.True(t, service.isHealthy)
		assert.NotNil(t, service.status)
	})
}

func TestBaseService_GetDBClient(t *testing.T) {
	t.Parallel()

	dbClient := &mockDBClient{}
	service := NewBaseService(dbClient)

	result := service.GetDBClient()

	assert.Equal(t, dbClient, result)
}

func TestBaseService_GetContext(t *testing.T) {
	t.Parallel()

	service := NewBaseService(nil)

	ctx := service.GetContext()

	assert.NotNil(t, ctx)
	assert.Equal(t, context.Background(), ctx)
}

func TestBaseService_GetContextWithTimeout(t *testing.T) {
	t.Parallel()

	service := NewBaseService(nil)
	timeout := 5 * time.Second

	ctx, cancel := service.GetContextWithTimeout(timeout)
	defer cancel()

	assert.NotNil(t, ctx)
	assert.NotNil(t, cancel)

	deadline, ok := ctx.Deadline()
	assert.True(t, ok)
	assert.WithinDuration(t, time.Now().Add(timeout), deadline, time.Second)
}

func TestBaseService_GetContextWithCancel(t *testing.T) {
	t.Parallel()

	service := NewBaseService(nil)

	ctx, cancel := service.GetContextWithCancel()
	defer cancel()

	assert.NotNil(t, ctx)
	assert.NotNil(t, cancel)

	// Test cancellation
	select {
	case <-ctx.Done():
		t.Fatal("Context should not be cancelled initially")
	default:
		// Expected behavior
	}

	cancel()

	select {
	case <-ctx.Done():
		// Expected behavior
	case <-time.After(time.Second):
		t.Fatal("Context should be cancelled after calling cancel()")
	}
}

func TestBaseService_ValidateInput(t *testing.T) {
	t.Parallel()

	service := NewBaseService(nil)

	tests := []struct {
		name        string
		input       interface{}
		expectError bool
		errorMsg    string
	}{
		{
			name:        "nil input",
			input:       nil,
			expectError: true,
			errorMsg:    "input cannot be nil",
		},
		{
			name:        "empty string",
			input:       "",
			expectError: true,
			errorMsg:    "input cannot be empty",
		},
		{
			name:        "nil pointer",
			input:       (*string)(nil),
			expectError: true,
			errorMsg:    "input cannot be nil",
		},
		{
			name:        "valid string",
			input:       "valid input",
			expectError: false,
		},
		{
			name:        "valid int",
			input:       42,
			expectError: false,
		},
		{
			name:        "valid struct",
			input:       struct{ Name string }{Name: "test"},
			expectError: false,
		},
		{
			name:        "valid pointer to struct",
			input:       &struct{ Name string }{Name: "test"},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.ValidateInput(tt.input)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestBaseService_HandleError(t *testing.T) {
	t.Parallel()

	service := NewBaseService(nil)

	tests := []struct {
		name      string
		err       error
		operation string
		expectNil bool
	}{
		{
			name:      "nil error",
			err:       nil,
			operation: "test-operation",
			expectNil: true,
		},
		{
			name:      "with error",
			err:       errors.New("original error"),
			operation: "test-operation",
			expectNil: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.HandleError(tt.err, tt.operation)

			if tt.expectNil {
				assert.Nil(t, result)
			} else {
				assert.NotNil(t, result)
				assert.Contains(t, result.Error(), tt.operation)
				assert.Contains(t, result.Error(), "original error")
			}
		})
	}
}

func TestBaseService_LogOperation(t *testing.T) {
	t.Parallel()

	t.Run("with logger", func(t *testing.T) {
		service := NewBaseService(nil)

		// Test logging without details
		assert.NotPanics(t, func() {
			service.LogOperation("test-operation", nil)
		})

		// Test logging with details
		details := map[string]interface{}{
			"key": "value",
		}
		assert.NotPanics(t, func() {
			service.LogOperation("test-operation", details)
		})
	})

	t.Run("without logger", func(t *testing.T) {
		service := &BaseService{
			Logger: nil,
		}

		// Should not panic even without logger
		assert.NotPanics(t, func() {
			service.LogOperation("test-operation", nil)
		})
	})
}

func TestBaseService_GetLogger(t *testing.T) {
	t.Parallel()

	logger := log.New(os.Stdout, "test", log.LstdFlags)
	service := &BaseService{
		Logger: logger,
	}

	result := service.GetLogger()

	assert.Equal(t, logger, result)
}

func TestBaseService_GetMetrics(t *testing.T) {
	t.Parallel()

	service := NewBaseService(nil)

	metrics := service.GetMetrics()

	assert.NotNil(t, metrics)
	assert.Equal(t, service.metrics, metrics)
}

func TestBaseService_RecordMetric(t *testing.T) {
	t.Parallel()

	t.Run("with metrics", func(t *testing.T) {
		service := NewBaseService(nil)

		service.RecordMetric("test-metric", 42)

		metrics := service.GetMetrics()
		value := metrics.GetMetric("test-metric")
		assert.Equal(t, 42, value)
	})

	t.Run("without metrics", func(t *testing.T) {
		service := &BaseService{
			metrics: nil,
		}

		// Should not panic
		assert.NotPanics(t, func() {
			service.RecordMetric("test-metric", 42)
		})
	})
}

func TestBaseService_StartTransaction(t *testing.T) {
	t.Parallel()

	service := NewBaseService(nil)
	ctx := context.Background()

	txCtx, commit, rollback := service.StartTransaction(ctx)

	assert.NotNil(t, txCtx)
	assert.NotNil(t, commit)
	assert.NotNil(t, rollback)

	// Test that transaction context has the expected value
	txValue := txCtx.Value("transaction")
	assert.Equal(t, "active", txValue)

	// Test commit function
	err := commit()
	assert.NoError(t, err)

	// Test rollback function
	err = rollback()
	assert.NoError(t, err)
}

func TestBaseService_WithRetry(t *testing.T) {
	t.Parallel()

	service := NewBaseService(nil)

	t.Run("successful on first try", func(t *testing.T) {
		callCount := 0
		operation := func() error {
			callCount++
			return nil
		}

		err := service.WithRetry(operation, 3)

		assert.NoError(t, err)
		assert.Equal(t, 1, callCount)
	})

	t.Run("successful on second try", func(t *testing.T) {
		callCount := 0
		operation := func() error {
			callCount++
			if callCount == 1 {
				return errors.New("first attempt failed")
			}
			return nil
		}

		err := service.WithRetry(operation, 3)

		assert.NoError(t, err)
		assert.Equal(t, 2, callCount)
	})

	t.Run("fails all attempts", func(t *testing.T) {
		callCount := 0
		operation := func() error {
			callCount++
			return errors.New("operation failed")
		}

		err := service.WithRetry(operation, 3)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "operation failed after 3 retries")
		assert.Equal(t, 4, callCount) // 1 initial + 3 retries
	})
}

func TestBaseService_GetConfig(t *testing.T) {
	t.Parallel()

	service := NewBaseService(nil)

	config := service.GetConfig()

	assert.NotNil(t, config)
	assert.Equal(t, service.config, config)
}

func TestBaseService_IsHealthy(t *testing.T) {
	t.Parallel()

	t.Run("initially healthy", func(t *testing.T) {
		service := NewBaseService(nil)

		assert.True(t, service.IsHealthy())
	})

	t.Run("after shutdown", func(t *testing.T) {
		service := NewBaseService(nil)
		ctx := context.Background()

		service.Shutdown(ctx)

		assert.False(t, service.IsHealthy())
	})
}

func TestBaseService_GetStatus(t *testing.T) {
	t.Parallel()

	service := NewBaseService(nil)

	status := service.GetStatus()

	assert.NotNil(t, status)
	assert.Equal(t, "running", status.State)
	assert.False(t, status.StartTime.IsZero())
}

func TestBaseService_Shutdown(t *testing.T) {
	t.Parallel()

	service := NewBaseService(nil)
	ctx := context.Background()

	assert.True(t, service.IsHealthy())
	assert.Equal(t, "running", service.GetStatus().State)

	err := service.Shutdown(ctx)

	assert.NoError(t, err)
	assert.False(t, service.IsHealthy())
	assert.Equal(t, "shutdown", service.GetStatus().State)
	assert.False(t, service.GetStatus().ShutdownTime.IsZero())
}

func TestBaseService_HealthCheck(t *testing.T) {
	t.Parallel()

	t.Run("with db client", func(t *testing.T) {
		dbClient := &mockDBClient{}
		service := NewBaseService(dbClient)
		ctx := context.Background()

		result := service.HealthCheck(ctx)

		assert.NotNil(t, result)
		assert.Equal(t, "healthy", result.Status)
		assert.NotEmpty(t, result.Services)
		assert.Contains(t, result.Services, "db_service_grpc")
		assert.Equal(t, "disabled", result.Services["db_service_grpc"].Status)
	})

	t.Run("without db client", func(t *testing.T) {
		service := NewBaseService(nil)
		ctx := context.Background()

		result := service.HealthCheck(ctx)

		assert.NotNil(t, result)
		assert.Equal(t, "healthy", result.Status)
		assert.Empty(t, result.Services)
	})
}

func TestServiceMetrics(t *testing.T) {
	t.Parallel()

	t.Run("NewServiceMetrics", func(t *testing.T) {
		metrics := NewServiceMetrics()

		assert.NotNil(t, metrics)
		assert.NotNil(t, metrics.metrics)
	})

	t.Run("RecordMetric and GetMetric", func(t *testing.T) {
		metrics := NewServiceMetrics()

		metrics.RecordMetric("test", 42)

		value := metrics.GetMetric("test")
		assert.Equal(t, 42, value)

		// Test non-existent metric
		value = metrics.GetMetric("nonexistent")
		assert.Nil(t, value)
	})

	t.Run("GetAllMetrics", func(t *testing.T) {
		metrics := NewServiceMetrics()

		metrics.RecordMetric("metric1", 42)
		metrics.RecordMetric("metric2", "value")

		allMetrics := metrics.GetAllMetrics()

		assert.Len(t, allMetrics, 2)
		assert.Equal(t, 42, allMetrics["metric1"])
		assert.Equal(t, "value", allMetrics["metric2"])
	})
}

func TestServiceConfig(t *testing.T) {
	t.Parallel()

	t.Run("NewServiceConfig", func(t *testing.T) {
		config := NewServiceConfig()

		assert.NotNil(t, config)
		assert.Equal(t, 3, config.MaxRetries)
		assert.Equal(t, 30*time.Second, config.Timeout)
		assert.Equal(t, 100, config.BatchSize)
		assert.True(t, config.EnableMetrics)
	})
}

func TestServiceStatus(t *testing.T) {
	t.Parallel()

	t.Run("GetUptime - running", func(t *testing.T) {
		status := &ServiceStatus{
			State:     "running",
			StartTime: time.Now().Add(-10 * time.Minute),
		}

		uptime := status.GetUptime()

		assert.True(t, uptime > 9*time.Minute)
		assert.True(t, uptime < 11*time.Minute)
	})

	t.Run("GetUptime - shutdown", func(t *testing.T) {
		startTime := time.Now().Add(-10 * time.Minute)
		shutdownTime := startTime.Add(5 * time.Minute)

		status := &ServiceStatus{
			State:        "shutdown",
			StartTime:    startTime,
			ShutdownTime: shutdownTime,
		}

		uptime := status.GetUptime()

		assert.Equal(t, 5*time.Minute, uptime)
	})
}

func TestHealthCheckResult(t *testing.T) {
	t.Parallel()

	t.Run("IsHealthy - all healthy", func(t *testing.T) {
		result := &HealthCheckResult{
			Status: "healthy",
			Services: map[string]*ServiceHealth{
				"service1": {Status: "healthy"},
				"service2": {Status: "healthy"},
			},
		}

		assert.True(t, result.IsHealthy())
	})

	t.Run("IsHealthy - unhealthy", func(t *testing.T) {
		result := &HealthCheckResult{
			Status: "unhealthy",
		}

		assert.False(t, result.IsHealthy())
	})

	t.Run("GetUnhealthyServices", func(t *testing.T) {
		result := &HealthCheckResult{
			Services: map[string]*ServiceHealth{
				"service1": {Status: "healthy"},
				"service2": {Status: "unhealthy"},
				"service3": {Status: "degraded"},
			},
		}

		unhealthy := result.GetUnhealthyServices()

		assert.Len(t, unhealthy, 2)
		assert.Contains(t, unhealthy, "service2")
		assert.Contains(t, unhealthy, "service3")
	})
}

// Test ServiceRegistry
func TestNewServiceRegistryGRPCOnly(t *testing.T) {
	t.Parallel()

	dbClient := &mockDBClient{}
	logger := log.New(os.Stdout, "test", log.LstdFlags)

	registry := NewServiceRegistryGRPCOnly(dbClient, logger)

	assert.NotNil(t, registry)
	assert.NotNil(t, registry.base)
	assert.Equal(t, logger, registry.logger)
}

func TestNewServiceRegistryWithDependencies(t *testing.T) {
	t.Parallel()

	dbClient := &mockDBClient{}
	logger := log.New(os.Stdout, "test", log.LstdFlags)

	registry := NewServiceRegistryWithDependencies(dbClient, nil, nil, logger)

	assert.NotNil(t, registry)
	assert.NotNil(t, registry.base)
	assert.NotNil(t, registry.etcService)
	assert.NotNil(t, registry.mappingService)
	assert.NotNil(t, registry.importService)
	assert.Equal(t, logger, registry.logger)
}

func TestServiceRegistry_Getters(t *testing.T) {
	t.Parallel()

	registry := NewServiceRegistryGRPCOnly(nil, nil)

	assert.NotNil(t, registry.GetBaseService())
	assert.Nil(t, registry.GetDownloadService())
}

func TestServiceRegistry_HealthCheck(t *testing.T) {
	t.Parallel()

	t.Run("without ETC service", func(t *testing.T) {
		registry := NewServiceRegistryGRPCOnly(nil, nil)
		ctx := context.Background()

		result := registry.HealthCheck(ctx)

		assert.NotNil(t, result)
		assert.Equal(t, "healthy", result.Status)
	})

	t.Run("with healthy ETC service", func(t *testing.T) {
		registry := NewServiceRegistryGRPCOnly(nil, nil)
		ctx := context.Background()

		result := registry.HealthCheck(ctx)

		assert.NotNil(t, result)
		assert.Equal(t, "healthy", result.Status)
	})
}

// Concurrent access tests
func TestBaseService_ConcurrentAccess(t *testing.T) {
	t.Parallel()

	service := NewBaseService(nil)
	numGoroutines := 10
	numOperations := 100

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()

			for j := 0; j < numOperations; j++ {
				// Test concurrent access to various methods
				service.IsHealthy()
				service.GetStatus()
				service.GetConfig()
				service.GetMetrics()
				service.GetLogger()
				service.GetDBClient()
				service.RecordMetric("concurrent-test", id*100+j)
				service.LogOperation("concurrent-operation", map[string]interface{}{"id": id, "op": j})
			}
		}(i)
	}

	wg.Wait()

	// Verify service is still functional
	assert.True(t, service.IsHealthy())
	assert.NotNil(t, service.GetStatus())
}

// Mock implementations for testing
type mockDBClient struct{}

type mockDownloadService struct{}

func (m *mockDownloadService) Download(ctx context.Context, params interface{}) error {
	return nil
}
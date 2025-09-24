package services

import (
	"context"
	"log"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/yhonda-ohishi/etc_meisai/src/mocks"
)

// Simple test to get basic coverage for all service constructors
func TestServiceConstructors(t *testing.T) {
	t.Run("ETCMeisaiService constructor", func(t *testing.T) {
		mockRepo := &mocks.MockETCMeisaiRecordRepository{}
		service := NewETCMeisaiService(mockRepo, nil)
		assert.NotNil(t, service)
	})

	t.Run("ETCMappingService constructor", func(t *testing.T) {
		mockMappingRepo := &mocks.MockETCMappingRepository{}
		mockRecordRepo := &mocks.MockETCMeisaiRecordRepository{}
		service := NewETCMappingService(mockMappingRepo, mockRecordRepo, nil)
		assert.NotNil(t, service)
	})

	t.Run("StatisticsService constructor", func(t *testing.T) {
		mockRepo := &mocks.MockStatisticsRepository{}
		service := NewStatisticsService(mockRepo, nil)
		assert.NotNil(t, service)
	})

	t.Run("BaseService constructor", func(t *testing.T) {
		service := NewBaseService(nil)
		assert.NotNil(t, service)
	})
}

func TestServiceMethods(t *testing.T) {
	t.Run("BaseService methods", func(t *testing.T) {
		service := NewBaseService(nil)

		// Test basic methods
		ctx := service.GetContext()
		assert.Equal(t, context.Background(), ctx)

		ctxTimeout, cancel := service.GetContextWithTimeout(5 * time.Second)
		assert.NotNil(t, ctxTimeout)
		cancel()

		ctxCancel, cancel2 := service.GetContextWithCancel()
		assert.NotNil(t, ctxCancel)
		cancel2()

		// Test validation
		err := service.ValidateInput("valid input")
		assert.NoError(t, err)

		err = service.ValidateInput(nil)
		assert.Error(t, err)

		// Test error handling
		wrappedErr := service.HandleError(nil, "test")
		assert.Nil(t, wrappedErr)

		// Test logging
		service.LogOperation("test", nil)

		// Test metrics
		service.RecordMetric("test", 42)
		metrics := service.GetMetrics()
		assert.NotNil(t, metrics)

		// Test transaction
		txCtx, commit, rollback := service.StartTransaction(context.Background())
		assert.NotNil(t, txCtx)
		assert.NoError(t, commit())
		assert.NoError(t, rollback())

		// Test retry
		callCount := 0
		err = service.WithRetry(func() error {
			callCount++
			return nil
		}, 3)
		assert.NoError(t, err)
		assert.Equal(t, 1, callCount)

		// Test health and status
		assert.True(t, service.IsHealthy())
		status := service.GetStatus()
		assert.Equal(t, "running", status.State)

		// Test health check
		result := service.HealthCheck(context.Background())
		assert.NotNil(t, result)

		// Test shutdown
		err = service.Shutdown(context.Background())
		assert.NoError(t, err)
		assert.False(t, service.IsHealthy())
	})

	t.Run("ServiceMetrics methods", func(t *testing.T) {
		metrics := NewServiceMetrics()

		metrics.RecordMetric("test1", 42)
		metrics.RecordMetric("test2", "value")

		value := metrics.GetMetric("test1")
		assert.Equal(t, 42, value)

		all := metrics.GetAllMetrics()
		assert.Len(t, all, 2)
		assert.Equal(t, 42, all["test1"])
		assert.Equal(t, "value", all["test2"])
	})

	t.Run("ServiceConfig", func(t *testing.T) {
		config := NewServiceConfig()
		assert.Equal(t, 3, config.MaxRetries)
		assert.Equal(t, 30*time.Second, config.Timeout)
		assert.Equal(t, 100, config.BatchSize)
		assert.True(t, config.EnableMetrics)
	})

	t.Run("ServiceStatus", func(t *testing.T) {
		status := &ServiceStatus{
			State:     "running",
			StartTime: time.Now().Add(-10 * time.Minute),
		}

		uptime := status.GetUptime()
		assert.True(t, uptime > 9*time.Minute)
	})

	t.Run("ServiceRegistry", func(t *testing.T) {
		logger := log.New(log.Writer(), "test", log.LstdFlags)
		registry := NewServiceRegistryGRPCOnly(nil, logger)

		assert.NotNil(t, registry.GetBaseService())

		result := registry.HealthCheck(context.Background())
		assert.NotNil(t, result)
	})
}
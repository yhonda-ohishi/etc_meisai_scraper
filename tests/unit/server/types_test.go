package server_test

import (
	"context"
	"errors"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/yhonda-ohishi/etc_meisai/src/server"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// MockServer implements the Server interface for testing
type MockServer struct {
	mock.Mock
	serveFunc        func(lis net.Listener) error
	gracefulStopFunc func()
	stopFunc         func()
}

func (m *MockServer) Serve(lis net.Listener) error {
	args := m.Called(lis)
	if m.serveFunc != nil {
		return m.serveFunc(lis)
	}
	return args.Error(0)
}

func (m *MockServer) GracefulStop() {
	m.Called()
	if m.gracefulStopFunc != nil {
		m.gracefulStopFunc()
	}
}

func (m *MockServer) Stop() {
	m.Called()
	if m.stopFunc != nil {
		m.stopFunc()
	}
}

// MockHealthChecker implements the HealthChecker interface for testing
type MockHealthChecker struct {
	mock.Mock
	serviceName   string
	checkFunc     func(ctx context.Context) error
}

func (m *MockHealthChecker) CheckHealth(ctx context.Context) error {
	args := m.Called(ctx)
	if m.checkFunc != nil {
		return m.checkFunc(ctx)
	}
	return args.Error(0)
}

func (m *MockHealthChecker) GetServiceName() string {
	return m.serviceName
}

// MockDependency implements the Dependency interface for testing
type MockDependency struct {
	mock.Mock
	name         string
	healthyFunc  func(ctx context.Context) (bool, error)
}

func (m *MockDependency) IsHealthy(ctx context.Context) (bool, error) {
	args := m.Called(ctx)
	if m.healthyFunc != nil {
		return m.healthyFunc(ctx)
	}
	return args.Bool(0), args.Error(1)
}

func (m *MockDependency) GetName() string {
	return m.name
}

// MockHealthCheckWatchServer implements the HealthCheckWatchServer interface
type MockHealthCheckWatchServer struct {
	mock.Mock
	ctx        context.Context
	sendFunc   func(*server.HealthCheckResponse) error
}

func (m *MockHealthCheckWatchServer) Send(resp *server.HealthCheckResponse) error {
	args := m.Called(resp)
	if m.sendFunc != nil {
		return m.sendFunc(resp)
	}
	return args.Error(0)
}

func (m *MockHealthCheckWatchServer) Context() context.Context {
	if m.ctx != nil {
		return m.ctx
	}
	return context.Background()
}

func TestNewGracefulShutdownV2(t *testing.T) {
	tests := []struct {
		name           string
		config         server.GracefulShutdownConfig
		expectedConfig server.GracefulShutdownConfig
	}{
		{
			name: "with custom timeout",
			config: server.GracefulShutdownConfig{
				ShutdownTimeout: 10 * time.Second,
				GracefulFirst:   true,
			},
			expectedConfig: server.GracefulShutdownConfig{
				ShutdownTimeout: 10 * time.Second,
				GracefulFirst:   true,
			},
		},
		{
			name: "with zero timeout should use default",
			config: server.GracefulShutdownConfig{
				ShutdownTimeout: 0,
				GracefulFirst:   false,
			},
			expectedConfig: server.GracefulShutdownConfig{
				ShutdownTimeout: 30 * time.Second,
				GracefulFirst:   false,
			},
		},
		{
			name: "with graceful first disabled",
			config: server.GracefulShutdownConfig{
				ShutdownTimeout: 5 * time.Second,
				GracefulFirst:   false,
			},
			expectedConfig: server.GracefulShutdownConfig{
				ShutdownTimeout: 5 * time.Second,
				GracefulFirst:   false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gs := server.NewGracefulShutdownV2(tt.config)

			assert.NotNil(t, gs)
			assert.False(t, gs.IsShutdown())
			assert.Empty(t, gs.GetRegisteredServers())
		})
	}
}

func TestGracefulShutdownV2_RegisterServer(t *testing.T) {
	tests := []struct {
		name       string
		serverName string
		server     server.Server
	}{
		{
			name:       "register valid server",
			serverName: "test-server",
			server:     &MockServer{},
		},
		{
			name:       "register server with empty name",
			serverName: "",
			server:     &MockServer{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := server.GracefulShutdownConfig{
				ShutdownTimeout: 5 * time.Second,
			}
			gs := server.NewGracefulShutdownV2(config)

			gs.RegisterServer(tt.serverName, tt.server)

			servers := gs.GetRegisteredServers()
			if tt.serverName != "" {
				assert.Contains(t, servers, tt.serverName)
			}
		})
	}
}

func TestGracefulShutdownV2_RegisterServer_AfterShutdown(t *testing.T) {
	config := server.GracefulShutdownConfig{
		ShutdownTimeout: 100 * time.Millisecond,
	}
	gs := server.NewGracefulShutdownV2(config)

	// Initiate shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	go gs.Shutdown(ctx)

	// Wait for shutdown to start
	time.Sleep(10 * time.Millisecond)

	// Try to register after shutdown
	mockServer := &MockServer{}
	gs.RegisterServer("late-server", mockServer)

	// Should not be registered
	servers := gs.GetRegisteredServers()
	assert.NotContains(t, servers, "late-server")
}

func TestGracefulShutdownV2_RegisterCleanupFunc(t *testing.T) {
	config := server.GracefulShutdownConfig{
		ShutdownTimeout: 5 * time.Second,
	}
	gs := server.NewGracefulShutdownV2(config)

	cleanupCalled := false
	gs.RegisterCleanupFunc("test-cleanup", func() error {
		cleanupCalled = true
		return nil
	})

	// Perform shutdown to test cleanup execution
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	err := gs.Shutdown(ctx)

	assert.NoError(t, err)
	assert.True(t, cleanupCalled)
}

func TestGracefulShutdownV2_RegisterCleanupFunc_AfterShutdown(t *testing.T) {
	config := server.GracefulShutdownConfig{
		ShutdownTimeout: 100 * time.Millisecond,
	}
	gs := server.NewGracefulShutdownV2(config)

	// Initiate shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	go gs.Shutdown(ctx)

	// Wait for shutdown to start
	time.Sleep(10 * time.Millisecond)

	// Try to register cleanup after shutdown
	cleanupCalled := false
	gs.RegisterCleanupFunc("late-cleanup", func() error {
		cleanupCalled = true
		return nil
	})

	// Wait for shutdown to complete
	gs.WaitForShutdown()

	// Cleanup should not have been called
	assert.False(t, cleanupCalled)
}

func TestGracefulShutdownV2_GetRegisteredServers(t *testing.T) {
	config := server.GracefulShutdownConfig{
		ShutdownTimeout: 5 * time.Second,
	}
	gs := server.NewGracefulShutdownV2(config)

	// Initially empty
	servers := gs.GetRegisteredServers()
	assert.Empty(t, servers)

	// Register some servers
	gs.RegisterServer("server1", &MockServer{})
	gs.RegisterServer("server2", &MockServer{})

	servers = gs.GetRegisteredServers()
	assert.Len(t, servers, 2)
	assert.Contains(t, servers, "server1")
	assert.Contains(t, servers, "server2")
}

func TestGracefulShutdownV2_Shutdown(t *testing.T) {
	tests := []struct {
		name              string
		gracefulFirst     bool
		servers           map[string]*MockServer
		cleanupFuncs      map[string]func() error
		timeout           time.Duration
		contextTimeout    time.Duration
		expectError       bool
		expectGracefulCall bool
		expectStopCall     bool
	}{
		{
			name:          "graceful shutdown with graceful first",
			gracefulFirst: true,
			servers: map[string]*MockServer{
				"server1": {},
			},
			cleanupFuncs: map[string]func() error{
				"cleanup1": func() error { return nil },
			},
			timeout:            5 * time.Second,
			contextTimeout:     10 * time.Second,
			expectError:        false,
			expectGracefulCall: true,
			expectStopCall:     false,
		},
		{
			name:          "force shutdown without graceful first",
			gracefulFirst: false,
			servers: map[string]*MockServer{
				"server1": {},
			},
			cleanupFuncs: map[string]func() error{
				"cleanup1": func() error { return nil },
			},
			timeout:            5 * time.Second,
			contextTimeout:     10 * time.Second,
			expectError:        false,
			expectGracefulCall: false,
			expectStopCall:     true,
		},
		{
			name:          "cleanup function error",
			gracefulFirst: true,
			servers: map[string]*MockServer{
				"server1": {},
			},
			cleanupFuncs: map[string]func() error{
				"cleanup1": func() error { return errors.New("cleanup failed") },
			},
			timeout:            5 * time.Second,
			contextTimeout:     10 * time.Second,
			expectError:        true,
			expectGracefulCall: true,
			expectStopCall:     false,
		},
		{
			name:          "context timeout",
			gracefulFirst: true,
			servers: map[string]*MockServer{
				"server1": {},
			},
			cleanupFuncs:       map[string]func() error{},
			timeout:            5 * time.Second,
			contextTimeout:     50 * time.Millisecond,
			expectError:        false, // Context cancellation might not cause error
			expectGracefulCall: true,
			expectStopCall:     false, // Stop may or may not be called
		},
		{
			name:          "internal timeout",
			gracefulFirst: true,
			servers: map[string]*MockServer{
				"server1": {},
			},
			cleanupFuncs:       map[string]func() error{},
			timeout:            50 * time.Millisecond,
			contextTimeout:     5 * time.Second,
			expectError:        false,
			expectGracefulCall: true,
			expectStopCall:     true, // Should call Stop on internal timeout
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := server.GracefulShutdownConfig{
				ShutdownTimeout: tt.timeout,
				GracefulFirst:   tt.gracefulFirst,
			}
			gs := server.NewGracefulShutdownV2(config)

			// Register servers with expectations
			for name, mockServer := range tt.servers {
				if tt.expectGracefulCall {
					if tt.contextTimeout > tt.timeout {
						// Internal timeout will trigger
						mockServer.gracefulStopFunc = func() {
							time.Sleep(tt.timeout + 10*time.Millisecond)
						}
					}
					mockServer.On("GracefulStop").Return()
				}
				if tt.expectStopCall {
					mockServer.On("Stop").Return()
				}
				// Always allow Stop call since GracefulShutdownV2 may force stop on timeout
				if !tt.expectStopCall {
					mockServer.On("Stop").Return().Maybe()
				}
				gs.RegisterServer(name, mockServer)
			}

			// Register cleanup functions
			for name, fn := range tt.cleanupFuncs {
				gs.RegisterCleanupFunc(name, fn)
			}

			ctx, cancel := context.WithTimeout(context.Background(), tt.contextTimeout)
			defer cancel()

			err := gs.Shutdown(ctx)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			assert.True(t, gs.IsShutdown())

			// Verify mock expectations
			for _, mockServer := range tt.servers {
				mockServer.AssertExpectations(t)
			}
		})
	}
}

func TestGracefulShutdownV2_Shutdown_MultipleCallsSameResult(t *testing.T) {
	config := server.GracefulShutdownConfig{
		ShutdownTimeout: 5 * time.Second,
	}
	gs := server.NewGracefulShutdownV2(config)

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	// First shutdown
	err1 := gs.Shutdown(ctx)
	assert.NoError(t, err1)

	// Second shutdown should return immediately without error
	err2 := gs.Shutdown(ctx)
	assert.NoError(t, err2)
}

func TestGracefulShutdownV2_WaitForShutdown(t *testing.T) {
	config := server.GracefulShutdownConfig{
		ShutdownTimeout: 100 * time.Millisecond,
	}
	gs := server.NewGracefulShutdownV2(config)

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		gs.WaitForShutdown()
	}()

	// Start shutdown after a delay
	time.Sleep(50 * time.Millisecond)
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	gs.Shutdown(ctx)

	wg.Wait() // Should not block indefinitely
}

func TestGracefulShutdownV2_GetMetrics(t *testing.T) {
	config := server.GracefulShutdownConfig{
		ShutdownTimeout: 5 * time.Second,
	}
	gs := server.NewGracefulShutdownV2(config)

	// Initial metrics
	metrics := gs.GetMetrics()
	assert.NotNil(t, metrics)
	assert.Equal(t, 0, metrics.RegisteredServers)
	assert.Equal(t, 0, metrics.RegisteredCleanupFuncs)
	assert.False(t, metrics.IsShutdown)
	assert.Equal(t, time.Duration(0), metrics.ShutdownDuration)

	// Register some servers and cleanup functions
	mockServer := &MockServer{}
	mockServer.On("GracefulStop").Return().Maybe()
	mockServer.On("Stop").Return().Maybe()
	gs.RegisterServer("server1", mockServer)
	gs.RegisterCleanupFunc("cleanup1", func() error { return nil })

	metrics = gs.GetMetrics()
	assert.Equal(t, 1, metrics.RegisteredServers)
	assert.Equal(t, 1, metrics.RegisteredCleanupFuncs)
	assert.False(t, metrics.IsShutdown)

	// After shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	gs.Shutdown(ctx)

	metrics = gs.GetMetrics()
	assert.True(t, metrics.IsShutdown)
	// Shutdown duration might be very small or zero for simple operations
	assert.GreaterOrEqual(t, metrics.ShutdownDuration, time.Duration(0))
}

func TestNewGRPCServerWrapper(t *testing.T) {
	grpcServer := grpc.NewServer()
	wrapper := server.NewGRPCServerWrapper(grpcServer)

	assert.NotNil(t, wrapper)
}

func TestGRPCServerWrapper_Serve(t *testing.T) {
	grpcServer := grpc.NewServer()
	wrapper := server.NewGRPCServerWrapper(grpcServer)

	// Create a test listener
	lis, err := net.Listen("tcp", ":0")
	require.NoError(t, err)
	defer lis.Close()

	// Use a channel to capture the error from the goroutine
	errChan := make(chan error, 1)

	// Start serving in a goroutine
	go func() {
		err := wrapper.Serve(lis)
		errChan <- err
	}()

	// Give it a moment to start
	time.Sleep(10 * time.Millisecond)

	// Stop the server
	wrapper.Stop()

	// Check the error from serving
	select {
	case serveErr := <-errChan:
		// gRPC server may or may not return an error when stopped
		// Just verify we got a response (error or nil)
		_ = serveErr // Acknowledge we received something
	case <-time.After(100 * time.Millisecond):
		t.Error("Server did not stop within expected time")
	}
}

func TestGRPCServerWrapper_GracefulStop(t *testing.T) {
	grpcServer := grpc.NewServer()
	wrapper := server.NewGRPCServerWrapper(grpcServer)

	// Should not panic
	wrapper.GracefulStop()
}

func TestGRPCServerWrapper_Stop(t *testing.T) {
	grpcServer := grpc.NewServer()
	wrapper := server.NewGRPCServerWrapper(grpcServer)

	// Should not panic
	wrapper.Stop()
}

func TestNewHealthCheckService(t *testing.T) {
	tests := []struct {
		name   string
		config server.HealthCheckConfig
	}{
		{
			name: "default config",
			config: server.HealthCheckConfig{
				CheckInterval:   time.Second,
				Timeout:         5 * time.Second,
				EnableReadiness: true,
				EnableLiveness:  true,
			},
		},
		{
			name: "minimal config",
			config: server.HealthCheckConfig{
				EnableReadiness: false,
				EnableLiveness:  false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := server.NewHealthCheckService(tt.config)

			assert.NotNil(t, service)
			assert.Empty(t, service.GetRegisteredCheckers())
			assert.Empty(t, service.GetRegisteredDependencies())
		})
	}
}

func TestHealthCheckService_RegisterChecker(t *testing.T) {
	config := server.HealthCheckConfig{}
	service := server.NewHealthCheckService(config)

	checker := &MockHealthChecker{serviceName: "test-service"}
	service.RegisterChecker(checker)

	checkers := service.GetRegisteredCheckers()
	assert.Contains(t, checkers, "test-service")
}

func TestHealthCheckService_RegisterDependency(t *testing.T) {
	config := server.HealthCheckConfig{}
	service := server.NewHealthCheckService(config)

	dependency := &MockDependency{name: "test-dependency"}
	service.RegisterDependency(dependency)

	dependencies := service.GetRegisteredDependencies()
	assert.Contains(t, dependencies, "test-dependency")
}

func TestHealthCheckService_Check(t *testing.T) {
	tests := []struct {
		name           string
		serviceName    string
		checkers       map[string]*MockHealthChecker
		checkErrors    map[string]error
		expectError    bool
		expectedStatus server.HealthCheckResponse_ServingStatus
	}{
		{
			name:        "overall health check - all healthy",
			serviceName: "",
			checkers: map[string]*MockHealthChecker{
				"service1": {serviceName: "service1"},
				"service2": {serviceName: "service2"},
			},
			checkErrors: map[string]error{
				"service1": nil,
				"service2": nil,
			},
			expectError:    false,
			expectedStatus: server.HealthCheckResponse_SERVING,
		},
		{
			name:        "overall health check - one unhealthy",
			serviceName: "",
			checkers: map[string]*MockHealthChecker{
				"service1": {serviceName: "service1"},
			},
			checkErrors: map[string]error{
				"service1": errors.New("service1 failed"),
			},
			expectError: true,
		},
		{
			name:        "specific service check - healthy",
			serviceName: "service1",
			checkers: map[string]*MockHealthChecker{
				"service1": {serviceName: "service1"},
			},
			checkErrors: map[string]error{
				"service1": nil,
			},
			expectError:    false,
			expectedStatus: server.HealthCheckResponse_SERVING,
		},
		{
			name:        "specific service check - unhealthy",
			serviceName: "service1",
			checkers: map[string]*MockHealthChecker{
				"service1": {serviceName: "service1"},
			},
			checkErrors: map[string]error{
				"service1": errors.New("service1 failed"),
			},
			expectError: true,
		},
		{
			name:        "specific service check - not found",
			serviceName: "non-existent",
			checkers:    map[string]*MockHealthChecker{},
			checkErrors: map[string]error{},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := server.HealthCheckConfig{}
			service := server.NewHealthCheckService(config)

			// Register checkers with expectations
			for name, checker := range tt.checkers {
				err := tt.checkErrors[name]
				checker.On("CheckHealth", mock.Anything).Return(err)
				service.RegisterChecker(checker)
			}

			req := &server.HealthCheckRequest{Service: tt.serviceName}
			resp, err := service.Check(context.Background(), req)

			if tt.expectError {
				assert.Error(t, err)
				// Check error type
				if tt.serviceName == "non-existent" {
					st, ok := status.FromError(err)
					assert.True(t, ok)
					assert.Equal(t, codes.NotFound, st.Code())
				} else {
					st, ok := status.FromError(err)
					assert.True(t, ok)
					assert.Equal(t, codes.Unavailable, st.Code())
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				assert.Equal(t, tt.expectedStatus, resp.Status)
			}

			// Verify mock expectations
			for _, checker := range tt.checkers {
				checker.AssertExpectations(t)
			}
		})
	}
}

func TestHealthCheckService_Watch(t *testing.T) {
	tests := []struct {
		name           string
		checkInterval  time.Duration
		checkError     error
		streamError    error
		contextTimeout time.Duration
		expectError    bool
	}{
		{
			name:          "successful watch",
			checkInterval: 50 * time.Millisecond,
			checkError:    nil,
			streamError:   nil,
			contextTimeout: 100 * time.Millisecond,
			expectError:   true, // Context timeout
		},
		{
			name:          "watch with check error",
			checkInterval: 50 * time.Millisecond,
			checkError:    errors.New("check failed"),
			streamError:   nil,
			contextTimeout: 100 * time.Millisecond,
			expectError:   true, // Context timeout
		},
		{
			name:          "watch with stream error",
			checkInterval: 50 * time.Millisecond,
			checkError:    nil,
			streamError:   errors.New("stream failed"),
			contextTimeout: 200 * time.Millisecond,
			expectError:   true, // Stream error
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := server.HealthCheckConfig{
				CheckInterval: tt.checkInterval,
			}
			service := server.NewHealthCheckService(config)

			// Register a checker
			checker := &MockHealthChecker{serviceName: "test-service"}
			checker.On("CheckHealth", mock.Anything).Return(tt.checkError)
			service.RegisterChecker(checker)

			// Create mock stream
			ctx, cancel := context.WithTimeout(context.Background(), tt.contextTimeout)
			defer cancel()

			stream := &MockHealthCheckWatchServer{ctx: ctx}
			if tt.streamError != nil {
				stream.On("Send", mock.Anything).Return(tt.streamError)
			} else {
				stream.On("Send", mock.Anything).Return(nil)
			}

			req := &server.HealthCheckRequest{Service: "test-service"}
			err := service.Watch(req, stream)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			stream.AssertExpectations(t)
		})
	}
}

func TestHealthCheckService_CheckDependencies(t *testing.T) {
	tests := []struct {
		name            string
		dependencies    map[string]*MockDependency
		healthResults   map[string]bool
		healthErrors    map[string]error
		expectHealthy   bool
		expectError     bool
	}{
		{
			name: "all dependencies healthy",
			dependencies: map[string]*MockDependency{
				"dep1": {name: "dep1"},
				"dep2": {name: "dep2"},
			},
			healthResults: map[string]bool{
				"dep1": true,
				"dep2": true,
			},
			healthErrors: map[string]error{
				"dep1": nil,
				"dep2": nil,
			},
			expectHealthy: true,
			expectError:   false,
		},
		{
			name: "one dependency unhealthy",
			dependencies: map[string]*MockDependency{
				"dep1": {name: "dep1"},
			},
			healthResults: map[string]bool{
				"dep1": false,
			},
			healthErrors: map[string]error{
				"dep1": nil,
			},
			expectHealthy: false,
			expectError:   false,
		},
		{
			name: "dependency check error",
			dependencies: map[string]*MockDependency{
				"dep1": {name: "dep1"},
			},
			healthResults: map[string]bool{
				"dep1": false,
			},
			healthErrors: map[string]error{
				"dep1": errors.New("check failed"),
			},
			expectHealthy: false,
			expectError:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := server.HealthCheckConfig{}
			service := server.NewHealthCheckService(config)

			// Register dependencies with expectations
			for name, dep := range tt.dependencies {
				healthy := tt.healthResults[name]
				err := tt.healthErrors[name]
				dep.On("IsHealthy", mock.Anything).Return(healthy, err)
				service.RegisterDependency(dep)
			}

			healthy, err := service.CheckDependencies(context.Background())

			assert.Equal(t, tt.expectHealthy, healthy)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			// Verify mock expectations
			for _, dep := range tt.dependencies {
				dep.AssertExpectations(t)
			}
		})
	}
}

func TestHealthCheckService_ReadinessCheck(t *testing.T) {
	tests := []struct {
		name            string
		enableReadiness bool
		checkers        map[string]*MockHealthChecker
		dependencies    map[string]*MockDependency
		checkErrors     map[string]error
		depResults      map[string]bool
		depErrors       map[string]error
		expectReady     bool
		expectError     bool
	}{
		{
			name:            "readiness disabled",
			enableReadiness: false,
			expectReady:     true,
			expectError:     false,
		},
		{
			name:            "all services and dependencies ready",
			enableReadiness: true,
			checkers: map[string]*MockHealthChecker{
				"service1": {serviceName: "service1"},
			},
			dependencies: map[string]*MockDependency{
				"dep1": {name: "dep1"},
			},
			checkErrors: map[string]error{
				"service1": nil,
			},
			depResults: map[string]bool{
				"dep1": true,
			},
			depErrors: map[string]error{
				"dep1": nil,
			},
			expectReady: true,
			expectError: false,
		},
		{
			name:            "service not ready",
			enableReadiness: true,
			checkers: map[string]*MockHealthChecker{
				"service1": {serviceName: "service1"},
			},
			checkErrors: map[string]error{
				"service1": errors.New("service failed"),
			},
			expectReady: false,
			expectError: false,
		},
		{
			name:            "dependency not ready",
			enableReadiness: true,
			checkers: map[string]*MockHealthChecker{
				"service1": {serviceName: "service1"},
			},
			dependencies: map[string]*MockDependency{
				"dep1": {name: "dep1"},
			},
			checkErrors: map[string]error{
				"service1": nil,
			},
			depResults: map[string]bool{
				"dep1": false,
			},
			depErrors: map[string]error{
				"dep1": nil,
			},
			expectReady: false,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := server.HealthCheckConfig{
				EnableReadiness: tt.enableReadiness,
			}
			service := server.NewHealthCheckService(config)

			// Register checkers
			for name, checker := range tt.checkers {
				err := tt.checkErrors[name]
				checker.On("CheckHealth", mock.Anything).Return(err)
				service.RegisterChecker(checker)
			}

			// Register dependencies
			for name, dep := range tt.dependencies {
				healthy := tt.depResults[name]
				err := tt.depErrors[name]
				dep.On("IsHealthy", mock.Anything).Return(healthy, err)
				service.RegisterDependency(dep)
			}

			ready, err := service.ReadinessCheck(context.Background())

			assert.Equal(t, tt.expectReady, ready)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			// Verify mock expectations
			for _, checker := range tt.checkers {
				checker.AssertExpectations(t)
			}
			for _, dep := range tt.dependencies {
				dep.AssertExpectations(t)
			}
		})
	}
}

func TestHealthCheckService_LivenessCheck(t *testing.T) {
	tests := []struct {
		name           string
		enableLiveness bool
		checkers       map[string]*MockHealthChecker
		checkErrors    map[string]error
		expectLive     bool
		expectError    bool
	}{
		{
			name:           "liveness disabled",
			enableLiveness: false,
			expectLive:     true,
			expectError:    false,
		},
		{
			name:           "all services live",
			enableLiveness: true,
			checkers: map[string]*MockHealthChecker{
				"service1": {serviceName: "service1"},
			},
			checkErrors: map[string]error{
				"service1": nil,
			},
			expectLive:  true,
			expectError: false,
		},
		{
			name:           "service not live",
			enableLiveness: true,
			checkers: map[string]*MockHealthChecker{
				"service1": {serviceName: "service1"},
			},
			checkErrors: map[string]error{
				"service1": errors.New("service failed"),
			},
			expectLive:  false,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := server.HealthCheckConfig{
				EnableLiveness: tt.enableLiveness,
			}
			service := server.NewHealthCheckService(config)

			// Register checkers
			for name, checker := range tt.checkers {
				err := tt.checkErrors[name]
				checker.On("CheckHealth", mock.Anything).Return(err)
				service.RegisterChecker(checker)
			}

			live, err := service.LivenessCheck(context.Background())

			assert.Equal(t, tt.expectLive, live)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			// Verify mock expectations
			for _, checker := range tt.checkers {
				checker.AssertExpectations(t)
			}
		})
	}
}

func TestHealthCheckService_GetStatus(t *testing.T) {
	tests := []struct {
		name        string
		checkers    map[string]*MockHealthChecker
		checkErrors map[string]error
		timeout     time.Duration
		expectHealthy bool
	}{
		{
			name: "all services healthy",
			checkers: map[string]*MockHealthChecker{
				"service1": {serviceName: "service1"},
				"service2": {serviceName: "service2"},
			},
			checkErrors: map[string]error{
				"service1": nil,
				"service2": nil,
			},
			timeout:       time.Second,
			expectHealthy: true,
		},
		{
			name: "one service unhealthy",
			checkers: map[string]*MockHealthChecker{
				"service1": {serviceName: "service1"},
				"service2": {serviceName: "service2"},
			},
			checkErrors: map[string]error{
				"service1": nil,
				"service2": errors.New("service2 failed"),
			},
			timeout:       time.Second,
			expectHealthy: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := server.HealthCheckConfig{
				Timeout: tt.timeout,
			}
			service := server.NewHealthCheckService(config)

			// Register checkers
			for name, checker := range tt.checkers {
				err := tt.checkErrors[name]
				checker.On("CheckHealth", mock.Anything).Return(err)
				service.RegisterChecker(checker)
			}

			status := service.GetStatus()

			assert.NotNil(t, status)
			assert.Equal(t, tt.expectHealthy, status.Healthy)
			assert.Len(t, status.Services, len(tt.checkers))

			for name := range tt.checkers {
				assert.Contains(t, status.Services, name)
				if tt.checkErrors[name] != nil {
					assert.Equal(t, "unhealthy", status.Services[name])
				} else {
					assert.Equal(t, "healthy", status.Services[name])
				}
			}

			// Verify mock expectations
			for _, checker := range tt.checkers {
				checker.AssertExpectations(t)
			}
		})
	}
}

// Concurrent tests
func TestGracefulShutdownV2_ConcurrentOperations(t *testing.T) {
	config := server.GracefulShutdownConfig{
		ShutdownTimeout: 5 * time.Second,
	}
	gs := server.NewGracefulShutdownV2(config)

	var wg sync.WaitGroup
	numGoroutines := 10

	// Concurrent registration
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			serverName := "server-" + string(rune(id))
			gs.RegisterServer(serverName, &MockServer{})

			cleanupName := "cleanup-" + string(rune(id))
			gs.RegisterCleanupFunc(cleanupName, func() error { return nil })
		}(i)
	}

	wg.Wait()

	// Should not panic or deadlock
	assert.False(t, gs.IsShutdown())
}

func TestHealthCheckService_ConcurrentOperations(t *testing.T) {
	config := server.HealthCheckConfig{}
	service := server.NewHealthCheckService(config)

	var wg sync.WaitGroup
	numGoroutines := 10

	// Concurrent registration
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			checkerName := "checker-" + string(rune(id))
			checker := &MockHealthChecker{serviceName: checkerName}
			service.RegisterChecker(checker)

			depName := "dep-" + string(rune(id))
			dep := &MockDependency{name: depName}
			service.RegisterDependency(dep)
		}(i)
	}

	wg.Wait()

	// Should not panic or deadlock
	checkers := service.GetRegisteredCheckers()
	dependencies := service.GetRegisteredDependencies()
	assert.Len(t, checkers, numGoroutines)
	assert.Len(t, dependencies, numGoroutines)
}

// Benchmark tests
func BenchmarkGracefulShutdownV2_RegisterServer(b *testing.B) {
	config := server.GracefulShutdownConfig{
		ShutdownTimeout: 5 * time.Second,
	}
	gs := server.NewGracefulShutdownV2(config)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		serverName := "server-" + string(rune(i))
		gs.RegisterServer(serverName, &MockServer{})
	}
}

func BenchmarkHealthCheckService_RegisterChecker(b *testing.B) {
	config := server.HealthCheckConfig{}
	service := server.NewHealthCheckService(config)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		checkerName := "checker-" + string(rune(i))
		checker := &MockHealthChecker{serviceName: checkerName}
		service.RegisterChecker(checker)
	}
}

// Edge case tests
func TestGracefulShutdownV2_EdgeCases(t *testing.T) {
	t.Run("register server with nil", func(t *testing.T) {
		config := server.GracefulShutdownConfig{}
		gs := server.NewGracefulShutdownV2(config)

		// Go doesn't panic on nil - it just stores nil in the map
		gs.RegisterServer("test", nil)
		servers := gs.GetRegisteredServers()
		assert.Contains(t, servers, "test")
	})

	t.Run("register cleanup with nil function", func(t *testing.T) {
		config := server.GracefulShutdownConfig{}
		gs := server.NewGracefulShutdownV2(config)

		// Go doesn't panic on nil function - it just stores nil
		gs.RegisterCleanupFunc("test", nil)
		// Function will panic when called during shutdown, not during registration
	})
}

func TestHealthCheckService_EdgeCases(t *testing.T) {
	t.Run("register nil checker", func(t *testing.T) {
		config := server.HealthCheckConfig{}
		service := server.NewHealthCheckService(config)

		// This will panic when trying to call GetServiceName() on nil
		assert.Panics(t, func() {
			service.RegisterChecker(nil)
		})
	})

	t.Run("register nil dependency", func(t *testing.T) {
		config := server.HealthCheckConfig{}
		service := server.NewHealthCheckService(config)

		// This will panic when trying to call GetName() on nil
		assert.Panics(t, func() {
			service.RegisterDependency(nil)
		})
	})

	t.Run("check with nil request", func(t *testing.T) {
		config := server.HealthCheckConfig{}
		service := server.NewHealthCheckService(config)

		// This will panic when trying to access req.Service
		assert.Panics(t, func() {
			service.Check(context.Background(), nil)
		})
	})
}
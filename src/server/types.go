package server

import (
	"context"
	"fmt"
	"net"
	"os"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Server interface defines a generic server that can be gracefully shut down
type Server interface {
	Serve(lis net.Listener) error
	GracefulStop()
	Stop()
}

// GracefulShutdownConfig holds configuration for graceful shutdown
type GracefulShutdownConfig struct {
	ShutdownTimeout time.Duration
	GracefulFirst   bool
	Signals         []os.Signal
}

// GracefulShutdownV2 is the enhanced version of graceful shutdown manager
type GracefulShutdownV2 struct {
	config        GracefulShutdownConfig
	servers       map[string]Server
	cleanupFuncs  map[string]func() error
	isShutdown    bool
	shutdownDur   time.Duration
	shutdownCh    chan struct{}
	mu            sync.RWMutex
}

// NewGracefulShutdown creates a new enhanced graceful shutdown manager
func NewGracefulShutdownV2(config GracefulShutdownConfig) *GracefulShutdownV2 {
	if config.ShutdownTimeout == 0 {
		config.ShutdownTimeout = 30 * time.Second
	}

	return &GracefulShutdownV2{
		config:       config,
		servers:      make(map[string]Server),
		cleanupFuncs: make(map[string]func() error),
		shutdownCh:   make(chan struct{}),
	}
}

// RegisterServer registers a server for graceful shutdown
func (gs *GracefulShutdownV2) RegisterServer(name string, server Server) {
	gs.mu.Lock()
	defer gs.mu.Unlock()

	if gs.isShutdown {
		return
	}

	gs.servers[name] = server
}

// RegisterCleanupFunc registers a cleanup function to be called during shutdown
func (gs *GracefulShutdownV2) RegisterCleanupFunc(name string, fn func() error) {
	gs.mu.Lock()
	defer gs.mu.Unlock()

	if gs.isShutdown {
		return
	}

	gs.cleanupFuncs[name] = fn
}

// GetRegisteredServers returns the list of registered server names
func (gs *GracefulShutdownV2) GetRegisteredServers() []string {
	gs.mu.RLock()
	defer gs.mu.RUnlock()

	names := make([]string, 0, len(gs.servers))
	for name := range gs.servers {
		names = append(names, name)
	}
	return names
}

// Shutdown performs graceful shutdown of all registered servers and cleanup functions
func (gs *GracefulShutdownV2) Shutdown(ctx context.Context) error {
	gs.mu.Lock()
	if gs.isShutdown {
		gs.mu.Unlock()
		return nil
	}
	gs.isShutdown = true

	// Copy maps to avoid concurrent access issues
	servers := make(map[string]Server)
	for k, v := range gs.servers {
		servers[k] = v
	}

	cleanupFuncs := make(map[string]func() error)
	for k, v := range gs.cleanupFuncs {
		cleanupFuncs[k] = v
	}
	gs.mu.Unlock()

	start := time.Now()
	defer func() {
		gs.shutdownDur = time.Since(start)
		close(gs.shutdownCh)
	}()

	// Shutdown servers
	var wg sync.WaitGroup
	for name, server := range servers {
		wg.Add(1)
		go func(n string, s Server) {
			defer wg.Done()

			if gs.config.GracefulFirst {
				s.GracefulStop()
			} else {
				s.Stop()
			}
		}(name, server)
	}

	// Wait for servers to shutdown with timeout
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// All servers shut down successfully
	case <-ctx.Done():
		// Timeout - force stop remaining servers
		for _, server := range servers {
			server.Stop()
		}
		return ctx.Err()
	case <-time.After(gs.config.ShutdownTimeout):
		// Internal timeout - force stop
		for _, server := range servers {
			server.Stop()
		}
	}

	// Run cleanup functions
	var cleanupErr error
	for _, fn := range cleanupFuncs {
		if err := fn(); err != nil && cleanupErr == nil {
			cleanupErr = err
		}
	}

	return cleanupErr
}

// IsShutdown returns true if shutdown has been initiated
func (gs *GracefulShutdownV2) IsShutdown() bool {
	gs.mu.RLock()
	defer gs.mu.RUnlock()
	return gs.isShutdown
}

// WaitForShutdown blocks until shutdown is complete
func (gs *GracefulShutdownV2) WaitForShutdown() {
	<-gs.shutdownCh
}

// GetMetrics returns metrics about the shutdown manager
func (gs *GracefulShutdownV2) GetMetrics() *ShutdownMetrics {
	gs.mu.RLock()
	defer gs.mu.RUnlock()

	return &ShutdownMetrics{
		RegisteredServers:      len(gs.servers),
		RegisteredCleanupFuncs: len(gs.cleanupFuncs),
		IsShutdown:             gs.isShutdown,
		ShutdownDuration:       gs.shutdownDur,
	}
}

// ShutdownMetrics contains metrics about the shutdown process
type ShutdownMetrics struct {
	RegisteredServers      int
	RegisteredCleanupFuncs int
	IsShutdown             bool
	ShutdownDuration       time.Duration
}

// GRPCServerWrapper wraps a gRPC server to implement the Server interface
type GRPCServerWrapper struct {
	server *grpc.Server
}

// NewGRPCServerWrapper creates a new gRPC server wrapper
func NewGRPCServerWrapper(server *grpc.Server) *GRPCServerWrapper {
	return &GRPCServerWrapper{server: server}
}

// Serve implements the Server interface
func (w *GRPCServerWrapper) Serve(lis net.Listener) error {
	return w.server.Serve(lis)
}

// GracefulStop implements the Server interface
func (w *GRPCServerWrapper) GracefulStop() {
	w.server.GracefulStop()
}

// Stop implements the Server interface
func (w *GRPCServerWrapper) Stop() {
	w.server.Stop()
}

// HealthCheckConfig holds configuration for health check service
type HealthCheckConfig struct {
	CheckInterval    time.Duration
	Timeout          time.Duration
	EnableReadiness  bool
	EnableLiveness   bool
}

// HealthChecker interface for services that can be health checked
type HealthChecker interface {
	CheckHealth(ctx context.Context) error
	GetServiceName() string
}

// Dependency interface for external dependencies
type Dependency interface {
	IsHealthy(ctx context.Context) (bool, error)
	GetName() string
}

// HealthCheckRequest represents a health check request
type HealthCheckRequest struct {
	Service string
}

// HealthCheckResponse_ServingStatus represents the serving status
type HealthCheckResponse_ServingStatus int32

const (
	HealthCheckResponse_UNKNOWN     HealthCheckResponse_ServingStatus = 0
	HealthCheckResponse_SERVING     HealthCheckResponse_ServingStatus = 1
	HealthCheckResponse_NOT_SERVING HealthCheckResponse_ServingStatus = 2
)

// HealthCheckResponse represents a health check response
type HealthCheckResponse struct {
	Status HealthCheckResponse_ServingStatus
}

// HealthCheckService implements health checking functionality
type HealthCheckService struct {
	config       HealthCheckConfig
	checkers     map[string]HealthChecker
	dependencies map[string]Dependency
	mu           sync.RWMutex
}

// NewHealthCheckService creates a new health check service
func NewHealthCheckService(config HealthCheckConfig) *HealthCheckService {
	return &HealthCheckService{
		config:       config,
		checkers:     make(map[string]HealthChecker),
		dependencies: make(map[string]Dependency),
	}
}

// RegisterChecker registers a health checker
func (s *HealthCheckService) RegisterChecker(checker HealthChecker) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.checkers[checker.GetServiceName()] = checker
}

// RegisterDependency registers a dependency
func (s *HealthCheckService) RegisterDependency(dep Dependency) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.dependencies[dep.GetName()] = dep
}

// GetRegisteredCheckers returns registered checker names
func (s *HealthCheckService) GetRegisteredCheckers() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	names := make([]string, 0, len(s.checkers))
	for name := range s.checkers {
		names = append(names, name)
	}
	return names
}

// GetRegisteredDependencies returns registered dependency names
func (s *HealthCheckService) GetRegisteredDependencies() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	names := make([]string, 0, len(s.dependencies))
	for name := range s.dependencies {
		names = append(names, name)
	}
	return names
}

// Check performs a health check
func (s *HealthCheckService) Check(ctx context.Context, req *HealthCheckRequest) (*HealthCheckResponse, error) {
	if req.Service == "" {
		// Overall health check
		return s.checkOverallHealth(ctx)
	}

	// Specific service health check
	return s.checkServiceHealth(ctx, req.Service)
}

// checkOverallHealth checks the health of all registered services
func (s *HealthCheckService) checkOverallHealth(ctx context.Context) (*HealthCheckResponse, error) {
	s.mu.RLock()
	checkers := make(map[string]HealthChecker)
	for k, v := range s.checkers {
		checkers[k] = v
	}
	s.mu.RUnlock()

	// Check all services
	for _, checker := range checkers {
		if err := checker.CheckHealth(ctx); err != nil {
			return nil, status.Errorf(codes.Unavailable, "service %s is unhealthy: %v", checker.GetServiceName(), err)
		}
	}

	return &HealthCheckResponse{Status: HealthCheckResponse_SERVING}, nil
}

// checkServiceHealth checks the health of a specific service
func (s *HealthCheckService) checkServiceHealth(ctx context.Context, serviceName string) (*HealthCheckResponse, error) {
	s.mu.RLock()
	checker, exists := s.checkers[serviceName]
	s.mu.RUnlock()

	if !exists {
		return nil, status.Error(codes.NotFound, "service not found")
	}

	if err := checker.CheckHealth(ctx); err != nil {
		return nil, status.Errorf(codes.Unavailable, "service unavailable: %v", err)
	}

	return &HealthCheckResponse{Status: HealthCheckResponse_SERVING}, nil
}

// Watch streams health check responses (simplified implementation)
func (s *HealthCheckService) Watch(req *HealthCheckRequest, stream HealthCheckWatchServer) error {
	ticker := time.NewTicker(s.config.CheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-stream.Context().Done():
			return stream.Context().Err()
		case <-ticker.C:
			resp, err := s.Check(stream.Context(), req)
			if err != nil {
				// Send NOT_SERVING status on error
				resp = &HealthCheckResponse{Status: HealthCheckResponse_NOT_SERVING}
			}

			if err := stream.Send(resp); err != nil {
				return err
			}
		}
	}
}

// CheckDependencies checks all registered dependencies
func (s *HealthCheckService) CheckDependencies(ctx context.Context) (bool, error) {
	s.mu.RLock()
	deps := make(map[string]Dependency)
	for k, v := range s.dependencies {
		deps[k] = v
	}
	s.mu.RUnlock()

	for _, dep := range deps {
		healthy, err := dep.IsHealthy(ctx)
		if err != nil {
			return false, fmt.Errorf("dependency %s check failed: %w", dep.GetName(), err)
		}
		if !healthy {
			return false, nil
		}
	}

	return true, nil
}

// ReadinessCheck performs a readiness check
func (s *HealthCheckService) ReadinessCheck(ctx context.Context) (bool, error) {
	if !s.config.EnableReadiness {
		return true, nil
	}

	// Check services
	_, err := s.checkOverallHealth(ctx)
	if err != nil {
		return false, nil
	}

	// Check dependencies
	return s.CheckDependencies(ctx)
}

// LivenessCheck performs a liveness check
func (s *HealthCheckService) LivenessCheck(ctx context.Context) (bool, error) {
	if !s.config.EnableLiveness {
		return true, nil
	}

	// Simple liveness check - just verify services respond
	_, err := s.checkOverallHealth(ctx)
	return err == nil, nil
}

// GetStatus returns current health status
func (s *HealthCheckService) GetStatus() *HealthStatus {
	s.mu.RLock()
	defer s.mu.RUnlock()

	services := make(map[string]string)
	healthy := true

	for name, checker := range s.checkers {
		ctx, cancel := context.WithTimeout(context.Background(), s.config.Timeout)
		err := checker.CheckHealth(ctx)
		cancel()

		if err != nil {
			services[name] = "unhealthy"
			healthy = false
		} else {
			services[name] = "healthy"
		}
	}

	return &HealthStatus{
		Healthy:  healthy,
		Services: services,
	}
}

// HealthStatus represents the current health status
type HealthStatus struct {
	Healthy  bool              `json:"healthy"`
	Services map[string]string `json:"services"`
}

// HealthCheckWatchServer interface for streaming health checks
type HealthCheckWatchServer interface {
	Send(*HealthCheckResponse) error
	Context() context.Context
}
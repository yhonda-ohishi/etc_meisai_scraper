package services

import (
	"context"
	"errors"
	"fmt"
	"log"
	"reflect"
	"sync"
	"time"

	// "github.com/yhonda-ohishi/etc_meisai/src/clients"
	"github.com/yhonda-ohishi/etc_meisai/src/repositories"
)

// BaseService provides common functionality for all services
type BaseService struct {
	// dbClient *clients.DBServiceClient // TODO: Fix after clients package is reimplemented
	dbClient          interface{} // temporary placeholder
	ETCRepository     repositories.ETCRepository
	MappingRepository repositories.MappingRepository
	Logger            *log.Logger
	mu                sync.RWMutex // protects state fields
	logMu             sync.RWMutex // separate mutex for logging operations
	metrics           *ServiceMetrics
	config            *ServiceConfig
	isHealthy         bool
	status            *ServiceStatus
}

// NewBaseService creates a new base service
func NewBaseService(dbClient interface{}) *BaseService {
	return &BaseService{
		dbClient:  dbClient,
		Logger:    log.New(log.Writer(), "[BaseService] ", log.LstdFlags),
		metrics:   NewServiceMetrics(),
		config:    NewServiceConfig(),
		isHealthy: true,
		status:    &ServiceStatus{State: "running", StartTime: time.Now()},
	}
}

// NewBaseServiceWithDependencies creates a new base service with full dependencies
func NewBaseServiceWithDependencies(dbClient interface{}, etcRepo repositories.ETCRepository, mappingRepo repositories.MappingRepository, logger *log.Logger) *BaseService {
	return &BaseService{
		dbClient:          dbClient,
		ETCRepository:     etcRepo,
		MappingRepository: mappingRepo,
		Logger:            logger,
		metrics:           NewServiceMetrics(),
		config:            NewServiceConfig(),
		isHealthy:         true,
		status:            &ServiceStatus{State: "running", StartTime: time.Now()},
	}
}

// GetDBClient returns the db_service gRPC client
// TODO: Restore when clients package is available
// func (s *BaseService) GetDBClient() *clients.DBServiceClient {
//	s.mu.RLock()
//	defer s.mu.RUnlock()
//	return s.dbClient
// }
func (s *BaseService) GetDBClient() interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.dbClient
}

// GetContext returns a base context
func (s *BaseService) GetContext() context.Context {
	return context.Background()
}

// GetContextWithTimeout returns a context with the specified timeout
func (s *BaseService) GetContextWithTimeout(timeout time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), timeout)
}

// GetContextWithCancel returns a context with cancel functionality
func (s *BaseService) GetContextWithCancel() (context.Context, context.CancelFunc) {
	return context.WithCancel(context.Background())
}

// ValidateInput validates the provided input
func (s *BaseService) ValidateInput(input interface{}) error {
	if input == nil {
		return errors.New("input cannot be nil")
	}

	// Check for empty string
	if str, ok := input.(string); ok && str == "" {
		return errors.New("input cannot be empty")
	}

	// Additional validation can be added here based on type
	val := reflect.ValueOf(input)
	if val.Kind() == reflect.Ptr && val.IsNil() {
		return errors.New("input cannot be nil")
	}

	return nil
}

// HandleError wraps an error with additional context information
func (s *BaseService) HandleError(err error, operation string) error {
	if err == nil {
		return nil
	}

	return fmt.Errorf("operation '%s' failed: %w", operation, err)
}

// LogOperation logs an operation with optional details
func (s *BaseService) LogOperation(operation string, details interface{}) {
	s.logMu.RLock()
	logger := s.Logger
	s.logMu.RUnlock()

	if logger == nil {
		return
	}

	if details != nil {
		logger.Printf("Operation: %s, Details: %+v", operation, details)
	} else {
		logger.Printf("Operation: %s", operation)
	}
}

// GetLogger returns the service logger
func (s *BaseService) GetLogger() *log.Logger {
	s.logMu.RLock()
	defer s.logMu.RUnlock()
	return s.Logger
}

// GetMetrics returns the service metrics
func (s *BaseService) GetMetrics() *ServiceMetrics {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.metrics
}

// RecordMetric records a metric value
func (s *BaseService) RecordMetric(name string, value interface{}) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.metrics == nil {
		return
	}

	s.metrics.RecordMetric(name, value)
}

// StartTransaction starts a new transaction context
func (s *BaseService) StartTransaction(ctx context.Context) (context.Context, func() error, func() error) {
	// For now, we'll create a simple mock transaction
	// TODO: Implement proper transaction management when database client is available

	txCtx := context.WithValue(ctx, "transaction", "active")

	commit := func() error {
		s.LogOperation("transaction_commit", "committed successfully")
		return nil
	}

	rollback := func() error {
		s.LogOperation("transaction_rollback", "rolled back successfully")
		return nil
	}

	return txCtx, commit, rollback
}

// WithRetry executes an operation with retry logic
func (s *BaseService) WithRetry(operation func() error, maxRetries int) error {
	return s.WithRetryContext(context.Background(), operation, maxRetries)
}

// WithRetryContext executes an operation with retry logic and context support
func (s *BaseService) WithRetryContext(ctx context.Context, operation func() error, maxRetries int) error {
	var lastErr error

	for attempt := 0; attempt <= maxRetries; attempt++ {
		// Check context before each attempt
		select {
		case <-ctx.Done():
			return fmt.Errorf("operation cancelled: %w", ctx.Err())
		default:
		}

		if err := operation(); err != nil {
			lastErr = err
			if attempt < maxRetries {
				// Simple exponential backoff with context-aware sleep
				backoff := time.Duration(attempt+1) * 100 * time.Millisecond
				select {
				case <-time.After(backoff):
					continue
				case <-ctx.Done():
					return fmt.Errorf("operation cancelled during retry: %w", ctx.Err())
				}
			}
		} else {
			return nil
		}
	}

	return fmt.Errorf("operation failed after %d retries: %w", maxRetries, lastErr)
}

// GetConfig returns the service configuration
func (s *BaseService) GetConfig() *ServiceConfig {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.config
}

// IsHealthy returns the current health status
func (s *BaseService) IsHealthy() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.isHealthy
}

// GetStatus returns the current service status
func (s *BaseService) GetStatus() *ServiceStatus {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.status
}

// Shutdown gracefully shuts down the service
func (s *BaseService) Shutdown(ctx context.Context) error {
	// Update state first
	s.mu.Lock()
	s.isHealthy = false
	s.status.State = "shutdown"
	s.status.ShutdownTime = time.Now()
	s.mu.Unlock() // Release lock before logging

	// Log operation without holding the state mutex
	s.LogOperation("service_shutdown", "service shutting down gracefully")

	return nil
}


// HealthCheck performs a comprehensive health check
func (s *BaseService) HealthCheck(ctx context.Context) *HealthCheckResult {
	result := &HealthCheckResult{
		Timestamp: time.Now(),
		Services:  make(map[string]*ServiceHealth),
	}


	// Check db_service connectivity
	// TODO: Restore when clients package is available
	// if s.dbClient != nil {
	//	grpcHealth := &ServiceHealth{Name: "db_service_grpc"}
	//	if err := s.dbClient.HealthCheck(ctx); err != nil {
	//		grpcHealth.Status = "unhealthy"
	//		grpcHealth.Error = err.Error()
	//	} else {
	//		grpcHealth.Status = "healthy"
	//	}
	//	result.Services["db_service_grpc"] = grpcHealth
	// }
	if s.dbClient != nil {
		grpcHealth := &ServiceHealth{Name: "db_service_grpc", Status: "disabled", Error: "clients package not available"}
		result.Services["db_service_grpc"] = grpcHealth
	}

	// Determine overall status
	result.Status = "healthy"
	for _, service := range result.Services {
		if service.Status != "healthy" {
			result.Status = "unhealthy"
			break
		}
	}

	return result
}

// ServiceRegistry manages service instances
type ServiceRegistry struct {
	base            *BaseService
	etcService      *ETCService
	mappingService  *MappingService
	importService   *ImportServiceLegacy
	downloadService DownloadServiceInterface
	etcRepo         repositories.ETCRepository
	mappingRepo     repositories.MappingRepository
	logger          *log.Logger
	mu              sync.RWMutex
}

// ServiceFactory defines the interface for creating services
type ServiceFactory interface {
	CreateETCService() *ETCService
	CreateMappingService() *MappingService
	CreateBaseService() *BaseService
	CreateImportService() *ImportServiceLegacy
	CreateDownloadService() DownloadServiceInterface
}

// NewServiceRegistryGRPCOnly creates a service registry using only gRPC (no local database)
// TODO: Restore when clients package is available
// func NewServiceRegistryGRPCOnly(dbClient *clients.DBServiceClient, logger *log.Logger) *ServiceRegistry {
func NewServiceRegistryGRPCOnly(dbClient interface{}, logger *log.Logger) *ServiceRegistry {
	// Create gRPC-only repositories
	// TODO: Restore when clients package is available
	// etcRepo := repositories.NewGRPCRepository(dbClient)
	// mappingRepo := repositories.NewMappingGRPCRepository(dbClient)
	// Temporary placeholders
	var etcRepo repositories.ETCRepository
	var mappingRepo repositories.MappingRepository

	// Create services
	// TODO: Restore when clients package is available
	// etcService := NewETCService(etcRepo, dbClient)
	// mappingService := NewMappingService(mappingRepo, etcRepo)
	// importService := NewImportService(dbClient, etcRepo, mappingRepo)
	// Temporary placeholders
	var etcService *ETCService
	var mappingService *MappingService
	var importService *ImportServiceLegacy

	// Create base service without local DB
	base := NewBaseService(dbClient)

	return &ServiceRegistry{
		base:            base,
		etcService:      etcService,
		mappingService:  mappingService,
		importService:   importService,
		downloadService: nil, // TODO: Initialize when download service is implemented
		etcRepo:         etcRepo,
		mappingRepo:     mappingRepo,
		logger:          logger,
	}
}

// NewServiceRegistryWithFactory creates a service registry using a factory for dependency injection
func NewServiceRegistryWithFactory(factory ServiceFactory, etcRepo repositories.ETCRepository, mappingRepo repositories.MappingRepository, logger *log.Logger) *ServiceRegistry {
	return &ServiceRegistry{
		base:            factory.CreateBaseService(),
		etcService:      factory.CreateETCService(),
		mappingService:  factory.CreateMappingService(),
		importService:   factory.CreateImportService(),
		downloadService: factory.CreateDownloadService(),
		etcRepo:         etcRepo,
		mappingRepo:     mappingRepo,
		logger:          logger,
	}
}

// NewServiceRegistryWithDependencies creates a service registry with explicit dependencies (for testing)
func NewServiceRegistryWithDependencies(
	dbClient interface{},
	etcRepo repositories.ETCRepository,
	mappingRepo repositories.MappingRepository,
	logger *log.Logger,
) *ServiceRegistry {
	// Create base service with dependencies
	base := NewBaseServiceWithDependencies(dbClient, etcRepo, mappingRepo, logger)

	// Create services with repositories
	etcService := &ETCService{
		repo:     etcRepo,
		dbClient: dbClient,
	}

	mappingService := &MappingService{
		mappingRepo: mappingRepo,
		etcRepo:     etcRepo,
	}

	// ImportServiceLegacy needs proper initialization with parser
	// For now, create with nil parser (will be set later if needed)
	importService := &ImportServiceLegacy{
		dbClient:    dbClient,
		etcRepo:     etcRepo,
		mappingRepo: mappingRepo,
		// parser will be nil - initialized when needed
	}

	return &ServiceRegistry{
		base:            base,
		etcService:      etcService,
		mappingService:  mappingService,
		importService:   importService,
		downloadService: nil, // Will be set separately if needed
		etcRepo:         etcRepo,
		mappingRepo:     mappingRepo,
		logger:          logger,
	}
}

// GetETCService returns the ETC service instance
func (r *ServiceRegistry) GetETCService() *ETCService {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.etcService
}

// GetETCRepository returns the ETC repository instance
func (r *ServiceRegistry) GetETCRepository() repositories.ETCRepository {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.etcRepo
}

// GetMappingService returns the mapping service instance
func (r *ServiceRegistry) GetMappingService() *MappingService {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.mappingService
}

// GetMappingRepository returns the mapping repository instance
func (r *ServiceRegistry) GetMappingRepository() repositories.MappingRepository {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.mappingRepo
}

// GetImportService returns the import service instance
func (r *ServiceRegistry) GetImportService() *ImportServiceLegacy {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.importService
}

// GetBaseService returns the base service instance
func (r *ServiceRegistry) GetBaseService() *BaseService {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.base
}

// GetDownloadService returns the download service instance
func (r *ServiceRegistry) GetDownloadService() DownloadServiceInterface {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.downloadService
}

// HealthCheck performs health check on all services
func (r *ServiceRegistry) HealthCheck(ctx context.Context) *HealthCheckResult {
	result := r.base.HealthCheck(ctx)

	// Add ETC service health check
	if r.etcService != nil {
		etcHealth := &ServiceHealth{Name: "etc_service"}
		if err := r.etcService.HealthCheck(ctx); err != nil {
			etcHealth.Status = "unhealthy"
			etcHealth.Error = err.Error()
			result.Status = "unhealthy"
		} else {
			etcHealth.Status = "healthy"
		}
		result.Services["etc_service"] = etcHealth
	}

	return result
}

// SetDownloadService sets the download service (for testing)
func (r *ServiceRegistry) SetDownloadService(service DownloadServiceInterface) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.downloadService = service
}

// HealthCheckResult represents the result of a health check
type HealthCheckResult struct {
	Status    string                    `json:"status"`
	Timestamp time.Time                 `json:"timestamp"`
	Services  map[string]*ServiceHealth `json:"services"`
}

// ServiceHealth represents the health status of a single service
type ServiceHealth struct {
	Name   string `json:"name"`
	Status string `json:"status"` // healthy, unhealthy, degraded
	Error  string `json:"error,omitempty"`
}

// IsHealthy returns true if all services are healthy
func (r *HealthCheckResult) IsHealthy() bool {
	return r.Status == "healthy"
}

// GetUnhealthyServices returns a list of unhealthy services
func (r *HealthCheckResult) GetUnhealthyServices() []string {
	var unhealthy []string
	for name, service := range r.Services {
		if service.Status != "healthy" {
			unhealthy = append(unhealthy, name)
		}
	}
	return unhealthy
}

// ServiceMetrics represents metrics for the service
type ServiceMetrics struct {
	metrics map[string]interface{}
	mu      sync.RWMutex
}

// NewServiceMetrics creates a new service metrics instance
func NewServiceMetrics() *ServiceMetrics {
	return &ServiceMetrics{
		metrics: make(map[string]interface{}),
	}
}

// RecordMetric records a metric value
func (m *ServiceMetrics) RecordMetric(name string, value interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.metrics[name] = value
}

// GetMetric retrieves a metric value
func (m *ServiceMetrics) GetMetric(name string) interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.metrics[name]
}

// GetAllMetrics returns all metrics
func (m *ServiceMetrics) GetAllMetrics() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make(map[string]interface{})
	for k, v := range m.metrics {
		result[k] = v
	}
	return result
}

// ServiceConfig represents configuration for the service
type ServiceConfig struct {
	MaxRetries    int           `json:"max_retries"`
	Timeout       time.Duration `json:"timeout"`
	BatchSize     int           `json:"batch_size"`
	EnableMetrics bool          `json:"enable_metrics"`
}

// NewServiceConfig creates a new service configuration with defaults
func NewServiceConfig() *ServiceConfig {
	return &ServiceConfig{
		MaxRetries:    3,
		Timeout:       30 * time.Second,
		BatchSize:     100,
		EnableMetrics: true,
	}
}

// ServiceStatus represents the current status of the service
type ServiceStatus struct {
	State        string    `json:"state"`
	StartTime    time.Time `json:"start_time"`
	ShutdownTime time.Time `json:"shutdown_time,omitempty"`
	Uptime       string    `json:"uptime"`
}

// GetUptime calculates and returns the service uptime
func (s *ServiceStatus) GetUptime() time.Duration {
	if s.State == "shutdown" && !s.ShutdownTime.IsZero() {
		return s.ShutdownTime.Sub(s.StartTime)
	}
	return time.Since(s.StartTime)
}
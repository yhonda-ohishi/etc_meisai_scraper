package services

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/yhonda-ohishi/etc_meisai/src/clients"
	"github.com/yhonda-ohishi/etc_meisai/src/repositories"
)

// BaseService provides common functionality for all services
type BaseService struct {
	dbClient *clients.DBServiceClient
	mu       sync.RWMutex
}

// NewBaseService creates a new base service
func NewBaseService(dbClient *clients.DBServiceClient) *BaseService {
	return &BaseService{
		dbClient: dbClient,
	}
}

// GetDBClient returns the db_service gRPC client
func (s *BaseService) GetDBClient() *clients.DBServiceClient {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.dbClient
}


// HealthCheck performs a comprehensive health check
func (s *BaseService) HealthCheck(ctx context.Context) *HealthCheckResult {
	result := &HealthCheckResult{
		Timestamp: time.Now(),
		Services:  make(map[string]*ServiceHealth),
	}


	// Check db_service connectivity
	if s.dbClient != nil {
		grpcHealth := &ServiceHealth{Name: "db_service_grpc"}
		if err := s.dbClient.HealthCheck(ctx); err != nil {
			grpcHealth.Status = "unhealthy"
			grpcHealth.Error = err.Error()
		} else {
			grpcHealth.Status = "healthy"
		}
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
	base           *BaseService
	etcService     *ETCService
	mappingService *MappingService
	importService  *ImportService
	etcRepo        repositories.ETCRepository
	mappingRepo    repositories.MappingRepository
	mu             sync.RWMutex
}

// NewServiceRegistryGRPCOnly creates a service registry using only gRPC (no local database)
func NewServiceRegistryGRPCOnly(dbClient *clients.DBServiceClient, logger *log.Logger) *ServiceRegistry {
	// Create gRPC-only repositories
	etcRepo := repositories.NewGRPCRepository(dbClient)
	mappingRepo := repositories.NewMappingGRPCRepository(dbClient)

	// Create services
	etcService := NewETCService(etcRepo, dbClient)
	mappingService := NewMappingService(mappingRepo, etcRepo)
	importService := NewImportService(dbClient, etcRepo, mappingRepo)

	// Create base service without local DB
	base := NewBaseService(dbClient)

	return &ServiceRegistry{
		base:           base,
		etcService:     etcService,
		mappingService: mappingService,
		importService:  importService,
		etcRepo:        etcRepo,
		mappingRepo:    mappingRepo,
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
func (r *ServiceRegistry) GetImportService() *ImportService {
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

// HealthCheck performs health check on all services
func (r *ServiceRegistry) HealthCheck(ctx context.Context) *HealthCheckResult {
	result := r.base.HealthCheck(ctx)

	// Add ETC service health check
	etcHealth := &ServiceHealth{Name: "etc_service"}
	if err := r.etcService.HealthCheck(ctx); err != nil {
		etcHealth.Status = "unhealthy"
		etcHealth.Error = err.Error()
		result.Status = "unhealthy"
	} else {
		etcHealth.Status = "healthy"
	}
	result.Services["etc_service"] = etcHealth

	return result
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
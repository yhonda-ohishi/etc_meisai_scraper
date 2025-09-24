package services

import "context"

// ServiceRegistryInterface defines the contract for service registry
type ServiceRegistryInterface interface {
	// Service getters - using concrete types to match existing implementation
	GetETCService() *ETCService
	GetMappingService() *MappingService
	GetBaseService() *BaseService
	GetImportService() *ImportServiceLegacy
	GetDownloadService() DownloadServiceInterface // This doesn't exist yet, keep as interface

	// Health check
	HealthCheck(ctx context.Context) *HealthCheckResult
}
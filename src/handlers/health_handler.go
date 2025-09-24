package handlers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"runtime"
	"time"

	"github.com/yhonda-ohishi/etc_meisai/src/services"
)

// HealthHandler handles health check endpoints
type HealthHandler struct {
	*BaseHandler
}

// HealthStatus represents the health status of the service
type HealthStatus struct {
	Status    string                 `json:"status"`
	Timestamp time.Time              `json:"timestamp"`
	Version   string                 `json:"version"`
	Uptime    string                 `json:"uptime"`
	Services  map[string]ServiceInfo `json:"services"`
	System    SystemInfo             `json:"system"`
}

// ServiceInfo represents individual service health
type ServiceInfo struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
	Latency string `json:"latency,omitempty"`
}

// SystemInfo represents system resource information
type SystemInfo struct {
	GoVersion   string `json:"go_version"`
	NumCPU      int    `json:"num_cpu"`
	NumGoroutine int    `json:"num_goroutine"`
	Memory      MemoryInfo `json:"memory"`
}

// MemoryInfo represents memory usage
type MemoryInfo struct {
	Alloc      uint64 `json:"alloc"`
	TotalAlloc uint64 `json:"total_alloc"`
	Sys        uint64 `json:"sys"`
	NumGC      uint32 `json:"num_gc"`
}

var startTime = time.Now()

// NewHealthHandler creates a new health handler
func NewHealthHandler(serviceRegistry *services.ServiceRegistry, logger *log.Logger) *HealthHandler {
	return &HealthHandler{
		BaseHandler: NewBaseHandler(serviceRegistry, logger),
	}
}

// HealthCheck handles GET /api/health - Main health check endpoint
func (h *HealthHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	// Check if service registry is available
	if h.ServiceRegistry == nil {
		h.RespondError(w, http.StatusServiceUnavailable, "service_unavailable",
			"Service registry not initialized", nil)
		return
	}

	// Get base service for health check
	baseService := h.ServiceRegistry.GetBaseService()
	if baseService == nil {
		h.RespondError(w, http.StatusServiceUnavailable, "service_unavailable",
			"Base service not available", nil)
		return
	}

	// Perform health check
	healthResult := baseService.HealthCheck(ctx)

	// Build response
	status := HealthStatus{
		Status:    "healthy",
		Timestamp: time.Now(),
		Version:   "1.0.0",
		Uptime:    time.Since(startTime).String(),
		Services:  make(map[string]ServiceInfo),
		System:    getSystemInfo(),
	}

	// Map service health results
	for name, service := range healthResult.Services {
		info := ServiceInfo{
			Status: service.Status,
		}
		if service.Error != "" {
			info.Message = service.Error
			status.Status = "degraded"
		}
		status.Services[name] = info
	}

	// Overall status
	if !healthResult.IsHealthy() {
		status.Status = "unhealthy"
	}

	// Set appropriate status code
	statusCode := http.StatusOK
	if status.Status == "unhealthy" {
		statusCode = http.StatusServiceUnavailable
	} else if status.Status == "degraded" {
		statusCode = http.StatusOK // Still return 200 for degraded
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(status)
}

// Liveness handles GET /api/health/live - Kubernetes liveness probe
func (h *HealthHandler) Liveness(w http.ResponseWriter, r *http.Request) {
	// Simple liveness check - just return OK if server is running
	response := map[string]interface{}{
		"status":    "alive",
		"timestamp": time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// Readiness handles GET /api/health/ready - Kubernetes readiness probe
func (h *HealthHandler) Readiness(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	// Check if service registry is available
	if h.ServiceRegistry == nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "not_ready",
			"message": "Service registry not initialized",
		})
		return
	}

	// Check if services are ready
	baseService := h.ServiceRegistry.GetBaseService()
	if baseService == nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "not_ready",
			"message": "Services not initialized",
		})
		return
	}

	healthResult := baseService.HealthCheck(ctx)

	// Check if all critical services are healthy
	isReady := healthResult.IsHealthy()

	statusCode := http.StatusOK
	status := "ready"
	if !isReady {
		statusCode = http.StatusServiceUnavailable
		status = "not_ready"
	}

	response := map[string]interface{}{
		"status":    status,
		"timestamp": time.Now(),
		"services":  healthResult.GetUnhealthyServices(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(response)
}

// DeepHealthCheck handles GET /api/health/deep - Detailed health check
func (h *HealthHandler) DeepHealthCheck(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	results := make(map[string]interface{})

	// Check ETC Service
	etcService := h.ServiceRegistry.GetETCService()
	if etcService != nil {
		start := time.Now()
		err := etcService.HealthCheck(ctx)
		latency := time.Since(start)

		results["etc_service"] = map[string]interface{}{
			"status":  getStatusFromError(err),
			"latency": latency.String(),
			"error":   errorString(err),
		}
	}

	// Check Mapping Service
	mappingService := h.ServiceRegistry.GetMappingService()
	if mappingService != nil {
		start := time.Now()
		err := mappingService.HealthCheck(ctx)
		latency := time.Since(start)

		results["mapping_service"] = map[string]interface{}{
			"status":  getStatusFromError(err),
			"latency": latency.String(),
			"error":   errorString(err),
		}
	}

	// Check Import Service
	importService := h.ServiceRegistry.GetImportService()
	if importService != nil {
		start := time.Now()
		err := importService.HealthCheck(ctx)
		latency := time.Since(start)

		results["import_service"] = map[string]interface{}{
			"status":  getStatusFromError(err),
			"latency": latency.String(),
			"error":   errorString(err),
		}
	}

	// Add system metrics
	results["system"] = getSystemInfo()
	results["timestamp"] = time.Now()
	results["uptime"] = time.Since(startTime).String()

	h.RespondSuccess(w, results, "Deep health check completed")
}

// getSystemInfo returns system resource information
func getSystemInfo() SystemInfo {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	return SystemInfo{
		GoVersion:    runtime.Version(),
		NumCPU:       runtime.NumCPU(),
		NumGoroutine: runtime.NumGoroutine(),
		Memory: MemoryInfo{
			Alloc:      m.Alloc,
			TotalAlloc: m.TotalAlloc,
			Sys:        m.Sys,
			NumGC:      m.NumGC,
		},
	}
}

// getStatusFromError returns status string based on error
func getStatusFromError(err error) string {
	if err == nil {
		return "healthy"
	}
	return "unhealthy"
}

// errorString returns error string or empty if nil
func errorString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}
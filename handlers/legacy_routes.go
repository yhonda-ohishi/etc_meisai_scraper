package handlers

import (
	"log"
	"net/http"
	"os"
	"time"

	"etc_meisai/src/adapters"
	"etc_meisai/src/middleware"
)

// LegacyRouteConfig holds configuration for legacy routes
type LegacyRouteConfig struct {
	AdapterConfig   *adapters.AdapterConfig
	EnableAPIKey    bool
	APIKeys         []string
	EnableBasicAuth bool
	Username        string
	Password        string
	RateLimit       int
	RateDuration    time.Duration
}

// DefaultLegacyRouteConfig returns default configuration
func DefaultLegacyRouteConfig() *LegacyRouteConfig {
	return &LegacyRouteConfig{
		AdapterConfig:   adapters.DefaultAdapterConfig(),
		EnableAPIKey:    false,
		APIKeys:         []string{},
		EnableBasicAuth: false,
		Username:        "",
		Password:        "",
		RateLimit:       60,  // 60 requests per minute
		RateDuration:    time.Minute,
	}
}

// LoadLegacyRouteConfigFromEnv loads configuration from environment variables
func LoadLegacyRouteConfigFromEnv() *LegacyRouteConfig {
	config := DefaultLegacyRouteConfig()

	// Load adapter config from environment
	if grpcAddr := os.Getenv("GRPC_ADDRESS"); grpcAddr != "" {
		config.AdapterConfig.GRPCAddress = grpcAddr
	}

	if timeout := os.Getenv("GRPC_TIMEOUT"); timeout != "" {
		if duration, err := time.ParseDuration(timeout); err == nil {
			config.AdapterConfig.Timeout = duration
		}
	}

	// Load authentication config
	if os.Getenv("LEGACY_API_KEY_AUTH") == "true" {
		config.EnableAPIKey = true
		if keys := os.Getenv("LEGACY_API_KEYS"); keys != "" {
			config.APIKeys = parseCommaSeparated(keys)
		}
	}

	if os.Getenv("LEGACY_BASIC_AUTH") == "true" {
		config.EnableBasicAuth = true
		config.Username = os.Getenv("LEGACY_AUTH_USERNAME")
		config.Password = os.Getenv("LEGACY_AUTH_PASSWORD")
	}

	return config
}

// RegisterLegacyRoutes registers all legacy HTTP routes with backward compatibility
func RegisterLegacyRoutes(mux *http.ServeMux, config *LegacyRouteConfig) error {
	if config == nil {
		config = LoadLegacyRouteConfigFromEnv()
	}

	// Create adapter
	adapter, err := adapters.NewChiToGRPCAdapter(config.AdapterConfig)
	if err != nil {
		return err
	}

	// Create rate limiter for legacy endpoints
	rateLimiter := middleware.NewRateLimiter(config.RateLimit, config.RateDuration)

	// Apply middleware chain
	middlewareChain := []func(http.Handler) http.Handler{
		middleware.SecurityMiddleware,
		rateLimiter.RateLimitMiddleware,
		addDeprecationWarnings,
	}

	// Add authentication middleware if enabled
	if config.EnableAPIKey && len(config.APIKeys) > 0 {
		middlewareChain = append(middlewareChain, middleware.APIKeyMiddleware(config.APIKeys))
	}

	if config.EnableBasicAuth && config.Username != "" && config.Password != "" {
		middlewareChain = append(middlewareChain, middleware.BasicAuthMiddleware(config.Username, config.Password))
	}

	// Register health check endpoint
	mux.Handle("/api/health", applyMiddleware(adapter.Health(), middlewareChain...))
	mux.Handle("/health", applyMiddleware(adapter.Health(), middlewareChain...))

	// Register account management endpoints
	mux.Handle("/api/accounts", applyMiddleware(
		handleMethod(map[string]http.HandlerFunc{
			"GET":  adapter.ListAccounts(),
			"POST": adapter.CreateAccount(),
		}), middlewareChain...))

	mux.Handle("/api/accounts/", applyMiddleware(
		handleMethod(map[string]http.HandlerFunc{
			"GET": adapter.GetAccount(),
		}), middlewareChain...))

	// Register transaction endpoints
	mux.Handle("/api/transactions", applyMiddleware(adapter.ListTransactions(), middlewareChain...))

	// Register download endpoints
	mux.Handle("/api/download", applyMiddleware(
		handleMethod(map[string]http.HandlerFunc{
			"POST": adapter.DownloadStatements(),
		}), middlewareChain...))

	mux.Handle("/api/download/statements", applyMiddleware(
		handleMethod(map[string]http.HandlerFunc{
			"POST": adapter.DownloadStatements(),
		}), middlewareChain...))

	// Register statistics endpoints
	mux.Handle("/api/stats", applyMiddleware(adapter.GetStats(), middlewareChain...))
	mux.Handle("/api/statistics", applyMiddleware(adapter.GetStats(), middlewareChain...))

	// Legacy v1 API endpoints (for backward compatibility)
	mux.Handle("/v1/accounts", applyMiddleware(adapter.ListAccounts(), middlewareChain...))
	mux.Handle("/v1/accounts/", applyMiddleware(adapter.GetAccount(), middlewareChain...))
	mux.Handle("/v1/transactions", applyMiddleware(adapter.ListTransactions(), middlewareChain...))
	mux.Handle("/v1/download", applyMiddleware(adapter.DownloadStatements(), middlewareChain...))
	mux.Handle("/v1/stats", applyMiddleware(adapter.GetStats(), middlewareChain...))

	// Legacy endpoints without /api prefix
	mux.Handle("/accounts", applyMiddleware(
		handleMethod(map[string]http.HandlerFunc{
			"GET":  adapter.ListAccounts(),
			"POST": adapter.CreateAccount(),
		}), middlewareChain...))

	mux.Handle("/accounts/", applyMiddleware(adapter.GetAccount(), middlewareChain...))
	mux.Handle("/transactions", applyMiddleware(adapter.ListTransactions(), middlewareChain...))
	mux.Handle("/download", applyMiddleware(adapter.DownloadStatements(), middlewareChain...))
	mux.Handle("/stats", applyMiddleware(adapter.GetStats(), middlewareChain...))

	log.Println("Legacy routes registered successfully")
	log.Printf("gRPC Adapter connected to: %s", config.AdapterConfig.GRPCAddress)

	if config.EnableAPIKey {
		log.Println("API Key authentication enabled for legacy routes")
	}
	if config.EnableBasicAuth {
		log.Println("Basic authentication enabled for legacy routes")
	}

	return nil
}

// Helper functions

// applyMiddleware applies a chain of middleware to a handler
func applyMiddleware(handler http.Handler, middlewares ...func(http.Handler) http.Handler) http.Handler {
	for i := len(middlewares) - 1; i >= 0; i-- {
		handler = middlewares[i](handler)
	}
	return handler
}

// handleMethod creates a handler that routes to different handlers based on HTTP method
func handleMethod(methodHandlers map[string]http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if handler, ok := methodHandlers[r.Method]; ok {
			handler(w, r)
			return
		}

		// Method not allowed
		allowedMethods := make([]string, 0, len(methodHandlers))
		for method := range methodHandlers {
			allowedMethods = append(allowedMethods, method)
		}

		w.Header().Set("Allow", joinStrings(allowedMethods, ", "))
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte(`{"error":"method not allowed"}`))
	}
}

// addDeprecationWarnings adds deprecation headers to legacy endpoints
func addDeprecationWarnings(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Add deprecation headers
		w.Header().Set("X-API-Deprecated", "true")
		w.Header().Set("X-API-Deprecation-Date", "2024-12-31")
		w.Header().Set("X-API-Deprecation-Info", "This API is deprecated. Please migrate to the gRPC API or use grpc-gateway endpoints.")
		w.Header().Set("X-API-Migration-Guide", "/docs#migration-guide")

		// Add sunset header (6 months from now)
		sunsetDate := time.Now().AddDate(0, 6, 0)
		w.Header().Set("Sunset", sunsetDate.Format(time.RFC1123))

		next.ServeHTTP(w, r)
	})
}

// parseCommaSeparated splits a comma-separated string into a slice
func parseCommaSeparated(value string) []string {
	if value == "" {
		return nil
	}

	var result []string
	for _, item := range splitString(value, ",") {
		trimmed := trimSpace(item)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

// splitString splits a string by separator
func splitString(s, sep string) []string {
	if s == "" {
		return nil
	}

	var result []string
	var current string

	for _, char := range s {
		if string(char) == sep {
			result = append(result, current)
			current = ""
		} else {
			current += string(char)
		}
	}

	if current != "" {
		result = append(result, current)
	}

	return result
}

// trimSpace removes leading and trailing whitespace
func trimSpace(s string) string {
	start := 0
	end := len(s)

	// Trim leading spaces
	for start < end && isSpace(s[start]) {
		start++
	}

	// Trim trailing spaces
	for end > start && isSpace(s[end-1]) {
		end--
	}

	return s[start:end]
}

// isSpace checks if a character is whitespace
func isSpace(c byte) bool {
	return c == ' ' || c == '\t' || c == '\n' || c == '\r'
}

// joinStrings joins a slice of strings with a separator
func joinStrings(strs []string, sep string) string {
	if len(strs) == 0 {
		return ""
	}

	if len(strs) == 1 {
		return strs[0]
	}

	// Calculate total length
	totalLen := len(strs) - 1 // separators
	for _, s := range strs {
		totalLen += len(s)
	}

	// Build result
	result := make([]byte, 0, totalLen)
	for i, s := range strs {
		if i > 0 {
			result = append(result, sep...)
		}
		result = append(result, s...)
	}

	return string(result)
}

// Compatibility handlers for specific legacy endpoints

// CreateCompatibilityHandler creates handlers for specific legacy endpoints that need special handling
func CreateCompatibilityHandler(adapter *adapters.ChiToGRPCAdapter) map[string]http.HandlerFunc {
	return map[string]http.HandlerFunc{
		// Legacy CSV upload endpoint
		"/upload/csv": func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusGone)
			w.Write([]byte(`{
				"error": "endpoint removed",
				"message": "CSV upload functionality has been replaced by the download/import workflow",
				"migration_guide": "/docs#csv-migration",
				"alternative_endpoints": ["/api/download", "/api/import"]
			}`))
		},

		// Legacy export endpoint
		"/export": func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusGone)
			w.Write([]byte(`{
				"error": "endpoint removed",
				"message": "Export functionality has been moved to the reports service",
				"migration_guide": "/docs#export-migration",
				"alternative_endpoints": ["/api/reports/export"]
			}`))
		},

		// Legacy settings endpoint
		"/settings": func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusGone)
			w.Write([]byte(`{
				"error": "endpoint removed",
				"message": "Settings are now managed through environment variables and configuration files",
				"migration_guide": "/docs#settings-migration"
			}`))
		},
	}
}

// RegisterCompatibilityHandlers registers handlers for removed/changed endpoints
func RegisterCompatibilityHandlers(mux *http.ServeMux, adapter *adapters.ChiToGRPCAdapter) {
	handlers := CreateCompatibilityHandler(adapter)

	for path, handler := range handlers {
		mux.HandleFunc(path, handler)
	}
}
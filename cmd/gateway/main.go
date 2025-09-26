package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "github.com/yhonda-ohishi/etc_meisai/src/pb"
	"github.com/yhonda-ohishi/etc_meisai/src/middleware"
)

type Config struct {
	GatewayPort string
	GRPCAddress string
	CORSOrigins []string
	Environment string
}

func loadConfig() *Config {
	config := &Config{
		GatewayPort: getEnv("GATEWAY_PORT", "8080"),
		GRPCAddress: getEnv("GRPC_ADDRESS", "localhost:9090"),
		Environment: getEnv("ENVIRONMENT", "development"),
	}

	// CORS origins configuration
	origins := getEnv("CORS_ORIGINS", "http://localhost:3000,http://localhost:8080")
	if origins != "" {
		config.CORSOrigins = parseCommaSeparated(origins)
	} else {
		config.CORSOrigins = []string{"*"} // Default to allow all in development
	}

	return config
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func parseCommaSeparated(value string) []string {
	if value == "" {
		return nil
	}
	var result []string
	for _, item := range splitAndTrim(value, ",") {
		if item != "" {
			result = append(result, item)
		}
	}
	return result
}

func splitAndTrim(s, sep string) []string {
	var result []string
	for _, item := range splitString(s, sep) {
		trimmed := trimSpace(item)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

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

func trimSpace(s string) string {
	start := 0
	end := len(s)

	// Trim leading spaces
	for start < end && (s[start] == ' ' || s[start] == '\t' || s[start] == '\n' || s[start] == '\r') {
		start++
	}

	// Trim trailing spaces
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t' || s[end-1] == '\n' || s[end-1] == '\r') {
		end--
	}

	return s[start:end]
}

func main() {
	config := loadConfig()

	log.Printf("Starting ETC Meisai Gateway Server...")
	log.Printf("Gateway Port: %s", config.GatewayPort)
	log.Printf("gRPC Address: %s", config.GRPCAddress)
	log.Printf("Environment: %s", config.Environment)
	log.Printf("CORS Origins: %v", config.CORSOrigins)

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Create gRPC connection
	conn, err := grpc.NewClient(
		config.GRPCAddress,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatalf("Failed to connect to gRPC server: %v", err)
	}
	defer conn.Close()

	// Create gRPC-Gateway mux
	mux := runtime.NewServeMux()

	// Register gRPC-Gateway handlers for repository services
	if err := pb.RegisterETCMappingRepositoryHandler(ctx, mux, conn); err != nil {
		log.Fatalf("Failed to register ETCMappingRepository handler: %v", err)
	}
	if err := pb.RegisterETCMeisaiRecordRepositoryHandler(ctx, mux, conn); err != nil {
		log.Fatalf("Failed to register ETCMeisaiRecordRepository handler: %v", err)
	}
	if err := pb.RegisterImportRepositoryHandler(ctx, mux, conn); err != nil {
		log.Fatalf("Failed to register ImportRepository handler: %v", err)
	}
	if err := pb.RegisterStatisticsRepositoryHandler(ctx, mux, conn); err != nil {
		log.Fatalf("Failed to register StatisticsRepository handler: %v", err)
	}

	// Register gRPC-Gateway handlers for business services
	if err := pb.RegisterMappingBusinessServiceHandler(ctx, mux, conn); err != nil {
		log.Fatalf("Failed to register MappingBusinessService handler: %v", err)
	}
	if err := pb.RegisterMeisaiBusinessServiceHandler(ctx, mux, conn); err != nil {
		log.Fatalf("Failed to register MeisaiBusinessService handler: %v", err)
	}

	// Create HTTP mux and apply middleware
	httpMux := http.NewServeMux()

	// Health check endpoint
	httpMux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"healthy","service":"etc-meisai-gateway"}`))
	})

	// Add Swagger UI endpoints (if available)
	// setupSwaggerRoutes(httpMux)

	// gRPC-Gateway routes
	httpMux.Handle("/", mux)

	// Apply middleware stack
	handler := middleware.Chain(
		httpMux,
		middleware.CORS(config.CORSOrigins),
		middleware.Security(),
		middleware.RateLimit(),
		middleware.RequestSize(10*1024*1024), // 10MB limit
		middleware.Logging(),
	)

	// Create HTTP server
	server := &http.Server{
		Addr:         ":" + config.GatewayPort,
		Handler:      handler,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Start server in goroutine
	go func() {
		log.Printf("HTTP Gateway server listening on :%s", config.GatewayPort)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server error: %v", err)
		}
	}()

	// Wait for interrupt signal for graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down HTTP Gateway server...")

	// Graceful shutdown with timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("HTTP server forced to shutdown: %v", err)
	}

	log.Println("HTTP Gateway server stopped")
}
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"github.com/yhonda-ohishi/etc_meisai/src/clients"
	"github.com/yhonda-ohishi/etc_meisai/src/config"
	"github.com/yhonda-ohishi/etc_meisai/src/handlers"
	customMiddleware "github.com/yhonda-ohishi/etc_meisai/src/middleware"
	"github.com/yhonda-ohishi/etc_meisai/src/services"
)

func main() {
	// Initialize configuration
	if err := config.InitSettings(); err != nil {
		log.Fatalf("Failed to initialize settings: %v", err)
	}

	cfg := config.GlobalSettings

	// Create logger
	logger := log.New(os.Stdout, "[SERVER] ", log.LstdFlags)
	logger.Println("Starting ETC Meisai Server (db_repo mode)...")

	// Initialize gRPC client (required for db_repo)
	dbClient, err := initGRPCClient(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize gRPC client for db_service: %v", err)
	}

	// Initialize services (using db_repo through gRPC)
	serviceRegistry := initServices(dbClient, logger)

	// Initialize handlers
	router := initRouter(serviceRegistry, logger)

	// Create HTTP server
	server := &http.Server{
		Addr:         cfg.Server.GetServerAddress(),
		Handler:      router,
		ReadTimeout:  time.Duration(cfg.Server.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.Server.WriteTimeout) * time.Second,
	}

	// Start server in goroutine
	go func() {
		logger.Printf("Server starting on %s", server.Addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Println("Shutting down server...")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	// Close gRPC client
	if dbClient != nil {
		if err := dbClient.Close(); err != nil {
			log.Printf("Error closing gRPC client: %v", err)
		}
	}

	logger.Println("Server stopped")
}


// initGRPCClient initializes the gRPC client
func initGRPCClient(cfg *config.Settings) (*clients.DBServiceClient, error) {
	address := cfg.GRPC.GetDBServiceAddress()
	timeout := cfg.GRPC.GetConnectionTimeout()

	client, err := clients.NewDBServiceClient(address, timeout)
	if err != nil {
		return nil, fmt.Errorf("failed to create gRPC client: %w", err)
	}

	return client, nil
}

// initServices initializes the service registry (db_repo mode - gRPC only)
func initServices(dbClient *clients.DBServiceClient, logger *log.Logger) *services.ServiceRegistry {
	// In db_repo mode, we don't use local database - everything goes through gRPC
	return services.NewServiceRegistryGRPCOnly(dbClient, logger)
}

// initRouter initializes the HTTP router
func initRouter(serviceRegistry *services.ServiceRegistry, logger *log.Logger) chi.Router {
	r := chi.NewRouter()

	// Chi middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Heartbeat("/ping"))

	// Custom security and monitoring middleware
	r.Use(customMiddleware.SecurityMiddleware)
	r.Use(customMiddleware.MetricsMiddleware)
	r.Use(customMiddleware.SanitizeMiddleware)

	// Rate limiting (100 requests per minute)
	rateLimiter := customMiddleware.NewRateLimiter(100, time.Minute)
	r.Use(rateLimiter.RateLimitMiddleware)

	// CORS configuration
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"}, // Configure appropriately for production
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token", "X-Request-ID"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Initialize handlers
	baseHandler := handlers.NewBaseHandler(serviceRegistry, logger)
	etcHandler := handlers.NewETCHandler(serviceRegistry, logger)
	parseHandler := handlers.NewParseHandler(serviceRegistry, logger)

	// System endpoints
	r.Get("/health", baseHandler.HealthCheck)
	r.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("pong"))
	})
	r.Get("/metrics", customMiddleware.MetricsHandler())

	// API routes
	r.Route("/api", func(r chi.Router) {
		// ETC endpoints
		r.Route("/etc", func(r chi.Router) {
			r.Post("/import", etcHandler.ImportData)
			r.Get("/meisai", etcHandler.GetMeisai)
			r.Get("/meisai/{id}", etcHandler.GetMeisaiByID)
			r.Post("/meisai", etcHandler.CreateMeisai)
			r.Get("/summary", etcHandler.GetSummary)
			r.Post("/bulk-import", etcHandler.BulkImport)
		})

		// Parse endpoints
		r.Route("/parse", func(r chi.Router) {
			r.Post("/csv", parseHandler.ParseCSV)
			r.Post("/import", parseHandler.ParseAndImport)
		})
	})

	// Static file serving (for development)
	workDir, _ := os.Getwd()
	filesDir := http.Dir(fmt.Sprintf("%s/web/static", workDir))
	r.Handle("/static/*", http.StripPrefix("/static", http.FileServer(filesDir)))

	return r
}
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/yhonda-ohishi/etc_meisai/src/config"
	"github.com/yhonda-ohishi/etc_meisai/src/handlers"
	// custommw "github.com/yhonda-ohishi/etc_meisai/src/middleware"
	"github.com/yhonda-ohishi/etc_meisai/src/server"
	"github.com/yhonda-ohishi/etc_meisai/src/services"
)

func main() {
	// Parse command-line flags
	var (
		configPath = flag.String("config", "", "Path to configuration file")
		port       = flag.String("port", "8080", "Server port")
		grpcAddr   = flag.String("grpc", "localhost:50051", "db_service gRPC address")
		logFile    = flag.String("log", "", "Log file path (empty for stdout)")
	)
	flag.Parse()

	// Setup logging
	logger := setupLogger(*logFile)
	logger.Println("[SERVER] Starting ETC Meisai Server with graceful shutdown...")

	// Load configuration
	cfg, err := loadConfig(*configPath)
	if err != nil {
		logger.Printf("[ERROR] Failed to load configuration: %v", err)
		cfg = &config.Settings{
			GRPC: config.GRPCSettings{
				DBServiceAddress: *grpcAddr,
				// Connection timeout is now handled internally
			},
		}
	}
	_ = cfg // TODO: Use cfg when clients package is restored

	// Initialize gRPC client
	// TODO: Fix after clients package is reimplemented
	// dbClient, err := initGRPCClient(cfg)
	// if err != nil {
	// 	logger.Fatalf("[ERROR] Failed to initialize gRPC client: %v", err)
	// }
	var dbClient interface{} = nil // temporary placeholder

	// Initialize shutdown manager
	shutdownManager := server.NewShutdownManager(logger)

	// Register db_service client for cleanup
	// TODO: Uncomment when dbClient has proper type with Close() method
	// shutdownManager.Register(server.NewDBServiceComponent(dbClient))

	// Initialize service registry
	serviceRegistry := services.NewServiceRegistryGRPCOnly(dbClient, logger)

	// Initialize metrics
	// Metrics middleware removed for now
	// metrics := custommw.NewMetrics()

	// Initialize router
	router := initRouterWithMiddleware(serviceRegistry, logger)

	// Create HTTP server
	httpServer := &http.Server{
		Addr:         ":" + *port,
		Handler:      router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Setup graceful shutdown
	gracefulShutdown := server.NewGracefulShutdown(httpServer, logger, 30*time.Second)

	// Register cleanup functions
	gracefulShutdown.RegisterCleanup(func() error {
		logger.Println("Closing database connections...")
		return shutdownManager.Shutdown(context.Background())
	})

	gracefulShutdown.RegisterCleanup(func() error {
		logger.Println("Flushing logs...")
		// Flush any buffered logs
		return nil
	})

	gracefulShutdown.RegisterCleanup(func() error {
		logger.Println("Saving metrics...")
		// Save final metrics
		// Final metrics logging
		logger.Printf("Saving final state...")
		return nil
	})

	// Start listening for shutdown signals
	gracefulShutdown.Start()

	// Start server
	logger.Printf("[SERVER] Starting HTTP server on %s", httpServer.Addr)
	if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Fatalf("[ERROR] Server failed: %v", err)
	}

	logger.Println("[SERVER] Server stopped gracefully")
}

func setupLogger(logFile string) *log.Logger {
	var logger *log.Logger
	if logFile != "" {
		file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			log.Printf("Failed to open log file %s: %v. Using stdout.", logFile, err)
			logger = log.New(os.Stdout, "[ETC_MEISAI] ", log.LstdFlags|log.Lshortfile)
		} else {
			logger = log.New(file, "[ETC_MEISAI] ", log.LstdFlags|log.Lshortfile)
		}
	} else {
		logger = log.New(os.Stdout, "[ETC_MEISAI] ", log.LstdFlags)
	}
	return logger
}

func loadConfig(configPath string) (*config.Settings, error) {
	if configPath == "" {
		// Try default locations
		for _, path := range []string{"config.yaml", "config.json", ".env"} {
			if _, err := os.Stat(path); err == nil {
				configPath = path
				break
			}
		}
	}

	if configPath == "" {
		return nil, fmt.Errorf("no configuration file found")
	}

	// LoadSettings not available, return default config
	return &config.Settings{
		GRPC: config.GRPCSettings{
			DBServiceAddress: "localhost:50051",
		},
	}, nil
}

// TODO: Fix after clients package is reimplemented
// func initGRPCClient(cfg *config.Settings) (*clients.DBServiceClient, error) {
// 	address := cfg.GRPC.GetDBServiceAddress()
// 	timeout := cfg.GRPC.GetConnectionTimeout()
//
// 	client, err := clients.NewDBServiceClient(address, timeout)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to create gRPC client: %w", err)
// 	}
//
// 	// Test connection
// 	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
// 	defer cancel()
//
// 	if err := client.HealthCheck(ctx); err != nil {
// 		log.Printf("[WARNING] db_service health check failed: %v", err)
// 		// Continue anyway - the service might become available later
// 	}
//
// 	return client, nil
// }

func initRouterWithMiddleware(serviceRegistry *services.ServiceRegistry, logger *log.Logger) chi.Router {
	r := chi.NewRouter()

	// Chi middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Heartbeat("/ping"))

	// Custom middleware - simplified for now
	// r.Use(custommw.RequestLogger(logger))
	// r.Use(custommw.MonitoringMiddleware(metrics, logger))
	// r.Use(custommw.ErrorHandler(logger))
	// r.Use(custommw.GRPCErrorHandler(logger))

	// CORS (if needed)
	r.Use(middleware.AllowContentType("application/json", "multipart/form-data"))

	// Rate limiting (basic)
	r.Use(middleware.Throttle(100))

	// Routes
	r.Route("/api", func(r chi.Router) {
		// Health and metrics
		r.Get("/health", handlers.NewHealthHandler(serviceRegistry, logger).HealthCheck)
		r.Get("/metrics", func(w http.ResponseWriter, req *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			metricsData := map[string]interface{}{
				"status": "ok",
				"timestamp": time.Now(),
			}
			if err := json.NewEncoder(w).Encode(metricsData); err != nil {
				logger.Printf("Failed to encode metrics: %v", err)
				http.Error(w, "Failed to get metrics", http.StatusInternalServerError)
			}
		})

		// ETC endpoints
		r.Route("/etc", func(r chi.Router) {
			etcHandler := handlers.NewETCHandler(serviceRegistry, logger)
			r.Get("/", etcHandler.ListETCMeisai)
			r.Post("/", etcHandler.CreateETCMeisai)
			r.Get("/{id}", etcHandler.GetETCMeisai)
			r.Put("/{id}", etcHandler.UpdateETCMeisai)
			r.Delete("/{id}", etcHandler.DeleteETCMeisai)
			r.Post("/bulk", etcHandler.BulkCreateETCMeisai)
			r.Get("/summary", etcHandler.GetETCSummary)
		})

		// Mapping endpoints
		r.Route("/mapping", func(r chi.Router) {
			mappingHandler := handlers.NewMappingHandler(serviceRegistry, logger)
			r.Get("/", mappingHandler.GetMappings)
			r.Post("/", mappingHandler.CreateMapping)
			r.Put("/{id}", mappingHandler.UpdateMapping)
			r.Delete("/{id}", mappingHandler.DeleteMapping)
			r.Post("/auto-match", mappingHandler.AutoMatch)
		})

		// Parse/Import endpoints
		r.Route("/import", func(r chi.Router) {
			parseHandler := handlers.NewParseHandler(serviceRegistry, logger)
			r.Post("/csv", parseHandler.ParseCSV)
			r.Post("/csv/import", parseHandler.ParseAndImport)
		})

		// Download endpoints - commented out for now
		// r.Route("/download", func(r chi.Router) {
		// 	downloadHandler := handlers.NewDownloadHandler(serviceRegistry, logger)
		// 	r.Post("/", downloadHandler.StartDownload)
		// 	r.Get("/{id}/status", downloadHandler.GetDownloadStatus)
		// })
	})

	// Static files (if needed)
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		info := map[string]interface{}{
			"service": "ETC Meisai API",
			"version": "1.0.0",
			"status":  "running",
			"endpoints": []string{
				"/api/health",
				"/api/metrics",
				"/api/etc",
				"/api/mapping",
				"/api/import",
				"/api/download",
			},
		}
		json.NewEncoder(w).Encode(info)
	})

	return r
}
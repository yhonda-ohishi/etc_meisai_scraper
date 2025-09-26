package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	grpcServer "github.com/yhonda-ohishi/etc_meisai/src/grpc"
)

func main() {
	// Parse command-line flags
	var (
		port    = flag.String("port", "50051", "gRPC server port")
		logFile = flag.String("log", "", "Log file path (empty for stdout)")
	)
	flag.Parse()

	// Setup logging
	logger := setupLogger(*logFile)
	logger.Println("[GRPC] Starting ETC Meisai gRPC Server...")

	// Create and start gRPC server
	server := grpcServer.NewServer(*port, logger)

	// Setup signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Start server in goroutine
	errChan := make(chan error, 1)
	go func() {
		if err := server.Start(); err != nil {
			errChan <- err
		}
	}()

	// Wait for shutdown signal or error
	select {
	case <-sigChan:
		logger.Println("[GRPC] Received shutdown signal")
		server.Stop()
	case err := <-errChan:
		logger.Fatalf("[GRPC] Server failed: %v", err)
	}

	logger.Println("[GRPC] Server stopped gracefully")
}

func setupLogger(logFile string) *log.Logger {
	var logger *log.Logger
	if logFile != "" {
		file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			log.Printf("Failed to open log file %s: %v. Using stdout.", logFile, err)
			logger = log.New(os.Stdout, "[ETC_MEISAI_GRPC] ", log.LstdFlags|log.Lshortfile)
		} else {
			logger = log.New(file, "[ETC_MEISAI_GRPC] ", log.LstdFlags|log.Lshortfile)
		}
	} else {
		logger = log.New(os.Stdout, "[ETC_MEISAI_GRPC] ", log.LstdFlags)
	}
	return logger
}
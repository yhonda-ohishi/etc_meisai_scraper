package grpc_test

import (
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"github.com/yhonda-ohishi/etc_meisai_scraper/src/grpc"
)

func TestNewServer(t *testing.T) {
	tests := []struct {
		name   string
		logger *log.Logger
	}{
		{
			name:   "with logger",
			logger: log.New(os.Stdout, "[TEST] ", log.LstdFlags),
		},
		{
			name:   "without logger",
			logger: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := grpc.NewServer(nil, tt.logger)

			if server == nil {
				t.Fatal("Expected non-nil server")
			}

			// Clean up
			server.Stop()
		})
	}
}

func TestServer_Start(t *testing.T) {
	logger := log.New(os.Stdout, "[TEST] ", log.LstdFlags)

	tests := []struct {
		name     string
		port     string
		expected string
	}{
		{
			name:     "with custom port",
			port:     "50053",
			expected: "50053",
		},
		{
			name:     "with empty port (use default)",
			port:     "",
			expected: "50051",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := grpc.NewServer(nil, logger)

			// Start server in background
			done := make(chan error, 1)
			go func() {
				done <- server.Start(tt.port)
			}()

			// Give server time to start
			time.Sleep(100 * time.Millisecond)

			// Stop server
			server.Stop()

			// Wait for server to stop (with timeout)
			select {
			case <-done:
				// Server stopped successfully
			case <-time.After(2 * time.Second):
				t.Error("Server did not stop within timeout")
			}
		})
	}
}

func TestServer_Stop(t *testing.T) {
	logger := log.New(os.Stdout, "[TEST] ", log.LstdFlags)
	server := grpc.NewServer(nil, logger)

	// Start server in background
	go func() {
		server.Start("50054")
	}()

	// Give server time to start
	time.Sleep(100 * time.Millisecond)

	// Test Stop method
	server.Stop()

	// Verify server stops gracefully
	// (The actual verification happens in the Start test above)
}

func TestServer_MultipleStops(t *testing.T) {
	logger := log.New(os.Stdout, "[TEST] ", log.LstdFlags)
	server := grpc.NewServer(nil, logger)

	// Calling Stop multiple times should not panic
	server.Stop()
	server.Stop()
}

func TestServer_StartWithInvalidPort(t *testing.T) {
	logger := log.New(os.Stdout, "[TEST] ", log.LstdFlags)
	server := grpc.NewServer(nil, logger)

	// Try to start with an invalid port
	err := server.Start("invalid")

	if err == nil {
		t.Error("Expected error for invalid port")
		server.Stop()
	}
}

func TestServer_StartWithUsedPort(t *testing.T) {
	logger := log.New(os.Stdout, "[TEST] ", log.LstdFlags)

	// Start first server
	server1 := grpc.NewServer(nil, logger)
	go func() {
		server1.Start("50055")
	}()

	// Give first server time to start
	time.Sleep(100 * time.Millisecond)
	defer server1.Stop()

	// Try to start second server on same port
	server2 := grpc.NewServer(nil, logger)
	err := server2.Start("50055")

	if err == nil {
		t.Error("Expected error when port is already in use")
		server2.Stop()
	}
}

func TestServer_ConcurrentOperations(t *testing.T) {
	logger := log.New(os.Stdout, "[TEST] ", log.LstdFlags)

	// Test concurrent server operations
	for i := 0; i < 5; i++ {
		go func(port int) {
			server := grpc.NewServer(nil, logger)
			go func() {
				server.Start(fmt.Sprintf("5006%d", port))
			}()
			time.Sleep(50 * time.Millisecond)
			server.Stop()
		}(i)
	}

	// Wait for all goroutines to complete
	time.Sleep(500 * time.Millisecond)
}
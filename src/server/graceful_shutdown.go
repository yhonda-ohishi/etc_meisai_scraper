package server

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

// GracefulShutdown manages graceful shutdown of the server
type GracefulShutdown struct {
	server         *http.Server
	logger         *log.Logger
	shutdownChan   chan os.Signal
	cleanupFuncs   []func() error
	timeout        time.Duration
	mu             sync.Mutex
}

// NewGracefulShutdown creates a new graceful shutdown manager
func NewGracefulShutdown(server *http.Server, logger *log.Logger, timeout time.Duration) *GracefulShutdown {
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	gs := &GracefulShutdown{
		server:       server,
		logger:       logger,
		shutdownChan: make(chan os.Signal, 1),
		cleanupFuncs: make([]func() error, 0),
		timeout:      timeout,
	}

	// Register signal handlers
	signal.Notify(gs.shutdownChan,
		os.Interrupt,    // Ctrl+C
		syscall.SIGTERM, // Termination signal
		syscall.SIGQUIT, // Quit signal
	)

	return gs
}

// RegisterCleanup registers a cleanup function to be called on shutdown
func (gs *GracefulShutdown) RegisterCleanup(fn func() error) {
	gs.mu.Lock()
	defer gs.mu.Unlock()
	gs.cleanupFuncs = append(gs.cleanupFuncs, fn)
}

// Start begins listening for shutdown signals
func (gs *GracefulShutdown) Start() {
	go func() {
		sig := <-gs.shutdownChan
		gs.logger.Printf("Received shutdown signal: %v", sig)
		gs.Shutdown()
	}()
}

// Shutdown performs graceful shutdown
func (gs *GracefulShutdown) Shutdown() {
	gs.logger.Println("Starting graceful shutdown...")

	// Create shutdown context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), gs.timeout)
	defer cancel()

	// Shutdown HTTP server
	if gs.server != nil {
		gs.logger.Println("Shutting down HTTP server...")
		if err := gs.server.Shutdown(ctx); err != nil {
			gs.logger.Printf("HTTP server shutdown error: %v", err)
			// Force shutdown if graceful fails
			if err := gs.server.Close(); err != nil {
				gs.logger.Printf("HTTP server force close error: %v", err)
			}
		} else {
			gs.logger.Println("HTTP server shut down successfully")
		}
	}

	// Run cleanup functions
	gs.mu.Lock()
	cleanupFuncs := gs.cleanupFuncs
	gs.mu.Unlock()

	var wg sync.WaitGroup
	for i, cleanup := range cleanupFuncs {
		wg.Add(1)
		go func(idx int, fn func() error) {
			defer wg.Done()
			gs.logger.Printf("Running cleanup function %d...", idx+1)
			if err := fn(); err != nil {
				gs.logger.Printf("Cleanup function %d error: %v", idx+1, err)
			} else {
				gs.logger.Printf("Cleanup function %d completed", idx+1)
			}
		}(i, cleanup)
	}

	// Wait for cleanup with timeout
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		gs.logger.Println("All cleanup functions completed")
	case <-ctx.Done():
		gs.logger.Println("Cleanup timeout exceeded")
	}

	gs.logger.Println("Graceful shutdown completed")
}

// ShutdownManager handles application-wide shutdown coordination
type ShutdownManager struct {
	logger       *log.Logger
	components   map[string]ShutdownComponent
	mu           sync.RWMutex
	shutdownOnce sync.Once
	isShutdown   bool
}

// ShutdownComponent represents a component that can be shut down
type ShutdownComponent interface {
	Name() string
	Shutdown(ctx context.Context) error
}

// NewShutdownManager creates a new shutdown manager
func NewShutdownManager(logger *log.Logger) *ShutdownManager {
	return &ShutdownManager{
		logger:     logger,
		components: make(map[string]ShutdownComponent),
	}
}

// Register registers a component for shutdown
func (sm *ShutdownManager) Register(component ShutdownComponent) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if sm.isShutdown {
		sm.logger.Printf("Cannot register component %s: shutdown in progress", component.Name())
		return
	}

	sm.components[component.Name()] = component
	sm.logger.Printf("Registered component for shutdown: %s", component.Name())
}

// Unregister removes a component from shutdown list
func (sm *ShutdownManager) Unregister(name string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	delete(sm.components, name)
	sm.logger.Printf("Unregistered component: %s", name)
}

// Shutdown performs coordinated shutdown of all components
func (sm *ShutdownManager) Shutdown(ctx context.Context) error {
	var finalErr error

	sm.shutdownOnce.Do(func() {
		sm.mu.Lock()
		sm.isShutdown = true
		components := make(map[string]ShutdownComponent)
		for k, v := range sm.components {
			components[k] = v
		}
		sm.mu.Unlock()

		sm.logger.Printf("Starting shutdown of %d components", len(components))

		// Shutdown components in parallel
		var wg sync.WaitGroup
		errChan := make(chan error, len(components))

		for name, component := range components {
			wg.Add(1)
			go func(n string, c ShutdownComponent) {
				defer wg.Done()

				sm.logger.Printf("Shutting down %s...", n)

				// Create component-specific timeout
				compCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
				defer cancel()

				if err := c.Shutdown(compCtx); err != nil {
					sm.logger.Printf("Error shutting down %s: %v", n, err)
					errChan <- err
				} else {
					sm.logger.Printf("Successfully shut down %s", n)
				}
			}(name, component)
		}

		// Wait for all components or timeout
		done := make(chan struct{})
		go func() {
			wg.Wait()
			close(done)
			close(errChan)
		}()

		select {
		case <-done:
			sm.logger.Println("All components shut down")
		case <-ctx.Done():
			sm.logger.Println("Component shutdown timeout exceeded")
			finalErr = ctx.Err()
		}

		// Collect errors
		for err := range errChan {
			if finalErr == nil {
				finalErr = err
			}
		}
	})

	return finalErr
}

// Example component implementations

// DBServiceComponent manages db_service client shutdown
type DBServiceComponent struct {
	client interface{ Close() error }
	name   string
}

func NewDBServiceComponent(client interface{ Close() error }) *DBServiceComponent {
	return &DBServiceComponent{
		client: client,
		name:   "db_service_client",
	}
}

func (c *DBServiceComponent) Name() string {
	return c.name
}

func (c *DBServiceComponent) Shutdown(ctx context.Context) error {
	return c.client.Close()
}

// WorkerPoolComponent manages worker pool shutdown
type WorkerPoolComponent struct {
	name    string
	workers int
	stopCh  chan struct{}
	wg      *sync.WaitGroup
}

func NewWorkerPoolComponent(name string, workers int) *WorkerPoolComponent {
	return &WorkerPoolComponent{
		name:    name,
		workers: workers,
		stopCh:  make(chan struct{}),
		wg:      &sync.WaitGroup{},
	}
}

func (c *WorkerPoolComponent) Name() string {
	return c.name
}

func (c *WorkerPoolComponent) Start(work func()) {
	for i := 0; i < c.workers; i++ {
		c.wg.Add(1)
		go func() {
			defer c.wg.Done()
			for {
				select {
				case <-c.stopCh:
					return
				default:
					work()
				}
			}
		}()
	}
}

func (c *WorkerPoolComponent) Shutdown(ctx context.Context) error {
	close(c.stopCh)

	done := make(chan struct{})
	go func() {
		c.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
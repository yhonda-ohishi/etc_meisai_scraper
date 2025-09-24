package integration

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// T011-D: Network resilience testing with connection failures and retries
func TestNetworkResilience_ConnectionFailures(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping network resilience test in short mode")
	}

	tests := []struct {
		name     string
		testFunc func(*testing.T)
	}{
		{"Connection Timeout", testConnectionTimeout},
		{"Connection Refused", testConnectionRefused},
		{"DNS Resolution Failure", testDNSResolutionFailure},
		{"Partial Response", testPartialResponse},
		{"Network Partition", testNetworkPartition},
		{"Retry With Backoff", testRetryWithBackoff},
		{"Circuit Breaker", testCircuitBreakerResilience},
		{"Connection Pool Exhaustion", testConnectionPoolExhaustion},
		{"Slow Network", testSlowNetwork},
		{"Intermittent Failures", testIntermittentFailures},
		{"Graceful Degradation", testGracefulDegradation},
		{"Failover", testFailover},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.testFunc(t)
		})
	}
}

func testConnectionTimeout(t *testing.T) {
	// Create server that delays response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(5 * time.Second)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := &ResilientHTTPClient{
		client: &http.Client{
			Timeout: 1 * time.Second,
		},
		maxRetries: 3,
	}

	ctx := context.Background()
	_, err := client.Get(ctx, server.URL)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "timeout")
}

func testConnectionRefused(t *testing.T) {
	// Use a port that's likely not in use
	client := &ResilientHTTPClient{
		client: &http.Client{
			Timeout: 1 * time.Second,
		},
		maxRetries: 3,
	}

	ctx := context.Background()
	_, err := client.Get(ctx, "http://localhost:59999")
	assert.Error(t, err)
}

func testDNSResolutionFailure(t *testing.T) {
	client := &ResilientHTTPClient{
		client: &http.Client{
			Timeout: 5 * time.Second,
		},
		maxRetries: 2,
	}

	ctx := context.Background()
	_, err := client.Get(ctx, "http://non-existent-domain-12345.com")
	assert.Error(t, err)
}

func testPartialResponse(t *testing.T) {
	// Server that sends partial response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", "100")
		w.WriteHeader(http.StatusOK)
		// Write less than promised
		w.Write([]byte("partial"))
		// Force close connection
		if hijacker, ok := w.(http.Hijacker); ok {
			conn, _, _ := hijacker.Hijack()
			conn.Close()
		}
	}))
	defer server.Close()

	client := &ResilientHTTPClient{
		client:     http.DefaultClient,
		maxRetries: 3,
	}

	ctx := context.Background()
	resp, err := client.Get(ctx, server.URL)
	if err == nil {
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		assert.Less(t, len(body), 100)
	}
}

func testNetworkPartition(t *testing.T) {
	// Simulate network partition by closing listener
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	addr := listener.Addr().String()
	listener.Close() // Immediately close to simulate partition

	client := &ResilientHTTPClient{
		client: &http.Client{
			Timeout: 1 * time.Second,
		},
		maxRetries: 2,
	}

	ctx := context.Background()
	_, err = client.Get(ctx, "http://"+addr)
	assert.Error(t, err)
}

func testRetryWithBackoff(t *testing.T) {
	attemptCount := atomic.Int32{}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		count := attemptCount.Add(1)
		if count < 3 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	}))
	defer server.Close()

	client := &ResilientHTTPClient{
		client:     http.DefaultClient,
		maxRetries: 5,
		backoff:    ExponentialBackoff{},
	}

	ctx := context.Background()
	start := time.Now()
	resp, err := client.Get(ctx, server.URL)
	duration := time.Since(start)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, int32(3), attemptCount.Load())
	// Should have some backoff delay
	assert.Greater(t, duration, 100*time.Millisecond)
}

func testCircuitBreakerResilience(t *testing.T) {
	failureCount := atomic.Int32{}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		failureCount.Add(1)
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	breaker := NewCircuitBreakerClient(5, 1*time.Second)

	// Trigger circuit breaker to open
	for i := 0; i < 6; i++ {
		_, _ = breaker.Get(context.Background(), server.URL)
	}

	// Circuit should be open now
	_, err := breaker.Get(context.Background(), server.URL)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "circuit breaker open")

	// Wait for half-open state
	time.Sleep(1100 * time.Millisecond)

	// Should try once more
	_, err = breaker.Get(context.Background(), server.URL)
	assert.Error(t, err) // Still fails, but attempted
}

func testConnectionPoolExhaustion(t *testing.T) {
	// Server that holds connections
	var connections sync.WaitGroup
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		connections.Add(1)
		defer connections.Done()

		// Hold connection for a bit
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Client with limited connection pool
	transport := &http.Transport{
		MaxIdleConns:        2,
		MaxIdleConnsPerHost: 2,
		MaxConnsPerHost:     2,
	}

	client := &ResilientHTTPClient{
		client: &http.Client{
			Transport: transport,
			Timeout:   5 * time.Second,
		},
		maxRetries: 1,
	}

	// Try to make more concurrent requests than pool allows
	var wg sync.WaitGroup
	errors := make(chan error, 10)

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			ctx := context.Background()
			_, err := client.Get(ctx, server.URL)
			if err != nil {
				errors <- err
			}
		}()
	}

	wg.Wait()
	close(errors)

	// Some requests should succeed despite pool limitations
	errCount := 0
	for range errors {
		errCount++
	}
	assert.Less(t, errCount, 10, "Some requests should succeed")
}

func testSlowNetwork(t *testing.T) {
	// Server that sends response slowly
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)

		// Send data in small chunks
		for i := 0; i < 10; i++ {
			w.Write([]byte(fmt.Sprintf("chunk_%d\n", i)))
			if f, ok := w.(http.Flusher); ok {
				f.Flush()
			}
			time.Sleep(50 * time.Millisecond)
		}
	}))
	defer server.Close()

	client := &ResilientHTTPClient{
		client: &http.Client{
			Timeout: 2 * time.Second,
		},
		maxRetries: 1,
	}

	ctx := context.Background()
	start := time.Now()
	resp, err := client.Get(ctx, server.URL)
	duration := time.Since(start)

	assert.NoError(t, err)
	assert.NotNil(t, resp)

	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()

	assert.Contains(t, string(body), "chunk_9")
	assert.Greater(t, duration, 400*time.Millisecond)
}

func testIntermittentFailures(t *testing.T) {
	requestCount := atomic.Int32{}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		count := requestCount.Add(1)

		// Fail every other request
		if count%2 == 1 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	}))
	defer server.Close()

	client := &ResilientHTTPClient{
		client:     http.DefaultClient,
		maxRetries: 3,
	}

	successCount := 0
	for i := 0; i < 10; i++ {
		ctx := context.Background()
		resp, err := client.Get(ctx, server.URL)
		if err == nil && resp.StatusCode == http.StatusOK {
			successCount++
			resp.Body.Close()
		}
	}

	assert.Greater(t, successCount, 5, "Should have reasonable success rate with retries")
}

func testGracefulDegradation(t *testing.T) {
	// Primary server (fails)
	primaryDown := atomic.Bool{}
	primaryDown.Store(true)

	primary := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if primaryDown.Load() {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("primary"))
	}))
	defer primary.Close()

	// Fallback server (works)
	fallback := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("fallback"))
	}))
	defer fallback.Close()

	client := &DegradableClient{
		primary:  primary.URL,
		fallback: fallback.URL,
		client:   http.DefaultClient,
	}

	// Should use fallback when primary is down
	resp, err := client.Get(context.Background())
	assert.NoError(t, err)

	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	assert.Equal(t, "fallback", string(body))

	// Restore primary
	primaryDown.Store(false)

	// Should use primary when available
	resp, err = client.Get(context.Background())
	assert.NoError(t, err)

	body, _ = io.ReadAll(resp.Body)
	resp.Body.Close()
	assert.Equal(t, "primary", string(body))
}

func testFailover(t *testing.T) {
	// Multiple servers with different availability
	servers := make([]*httptest.Server, 3)
	availability := []atomic.Bool{{}, {}, {}}

	for i := range servers {
		idx := i
		servers[i] = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !availability[idx].Load() {
				w.WriteHeader(http.StatusServiceUnavailable)
				return
			}
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(fmt.Sprintf("server_%d", idx)))
		}))
		defer servers[i].Close()
	}

	// Only server 2 is available initially
	availability[2].Store(true)

	client := &FailoverClient{
		servers: []string{
			servers[0].URL,
			servers[1].URL,
			servers[2].URL,
		},
		client: http.DefaultClient,
	}

	// Should failover to server 2
	resp, err := client.Get(context.Background())
	assert.NoError(t, err)

	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	assert.Equal(t, "server_2", string(body))

	// Make server 0 available
	availability[0].Store(true)
	availability[2].Store(false)

	// Should use server 0 now
	resp, err = client.Get(context.Background())
	assert.NoError(t, err)

	body, _ = io.ReadAll(resp.Body)
	resp.Body.Close()
	assert.Equal(t, "server_0", string(body))
}

// Test health check monitoring
func TestHealthCheckMonitoring(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping health check monitoring test in short mode")
	}

	healthy := atomic.Bool{}
	healthy.Store(true)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/health" {
			if healthy.Load() {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("healthy"))
			} else {
				w.WriteHeader(http.StatusServiceUnavailable)
				w.Write([]byte("unhealthy"))
			}
			return
		}

		// Regular endpoint
		if healthy.Load() {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("response"))
		} else {
			w.WriteHeader(http.StatusServiceUnavailable)
		}
	}))
	defer server.Close()

	monitor := NewHealthMonitor(server.URL, 100*time.Millisecond)
	monitor.Start()
	defer monitor.Stop()

	// Wait for initial health check
	time.Sleep(150 * time.Millisecond)
	assert.True(t, monitor.IsHealthy())

	// Make unhealthy
	healthy.Store(false)
	time.Sleep(150 * time.Millisecond)
	assert.False(t, monitor.IsHealthy())

	// Make healthy again
	healthy.Store(true)
	time.Sleep(150 * time.Millisecond)
	assert.True(t, monitor.IsHealthy())
}

// Test adaptive timeout
func TestAdaptiveTimeout(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping adaptive timeout test in short mode")
	}

	responseTime := atomic.Int64{}
	responseTime.Store(100) // Start with 100ms response time

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		delay := time.Duration(responseTime.Load()) * time.Millisecond
		time.Sleep(delay)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewAdaptiveTimeoutClient(500*time.Millisecond, 2*time.Second)

	// Fast responses should reduce timeout
	for i := 0; i < 5; i++ {
		ctx := context.Background()
		_, err := client.Get(ctx, server.URL)
		assert.NoError(t, err)
	}
	assert.Less(t, client.GetTimeout(), 1*time.Second)

	// Slow responses should increase timeout
	responseTime.Store(1000)
	for i := 0; i < 3; i++ {
		ctx := context.Background()
		_, _ = client.Get(ctx, server.URL)
	}
	assert.Greater(t, client.GetTimeout(), 1*time.Second)
}

// Helper types

type ResilientHTTPClient struct {
	client     *http.Client
	maxRetries int
	backoff    BackoffStrategy
	mu         sync.Mutex
}

func (c *ResilientHTTPClient) Get(ctx context.Context, url string) (*http.Response, error) {
	var lastErr error

	for attempt := 0; attempt <= c.maxRetries; attempt++ {
		if attempt > 0 && c.backoff != nil {
			delay := c.backoff.NextDelay(attempt)
			time.Sleep(delay)
		}

		req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
		if err != nil {
			return nil, err
		}

		resp, err := c.client.Do(req)
		if err != nil {
			lastErr = err
			continue
		}

		// Retry on server errors
		if resp.StatusCode >= 500 {
			resp.Body.Close()
			lastErr = fmt.Errorf("server error: %d", resp.StatusCode)
			continue
		}

		return resp, nil
	}

	return nil, fmt.Errorf("failed after %d retries: %w", c.maxRetries, lastErr)
}

type BackoffStrategy interface {
	NextDelay(attempt int) time.Duration
}

type ExponentialBackoff struct{}

func (e ExponentialBackoff) NextDelay(attempt int) time.Duration {
	delay := time.Duration(1<<uint(attempt-1)) * 100 * time.Millisecond
	if delay > 10*time.Second {
		delay = 10 * time.Second
	}
	return delay
}

type CircuitBreakerClient struct {
	client         *http.Client
	maxFailures    int
	resetTimeout   time.Duration
	failures       atomic.Int32
	lastFailTime   atomic.Int64
	state          atomic.Value // "closed", "open", "half-open"
}

func NewCircuitBreakerClient(maxFailures int, resetTimeout time.Duration) *CircuitBreakerClient {
	c := &CircuitBreakerClient{
		client:       http.DefaultClient,
		maxFailures:  maxFailures,
		resetTimeout: resetTimeout,
	}
	c.state.Store("closed")
	return c
}

func (c *CircuitBreakerClient) Get(ctx context.Context, url string) (*http.Response, error) {
	state := c.state.Load().(string)

	// Check if circuit should be reset
	if state == "open" {
		lastFail := time.Unix(0, c.lastFailTime.Load())
		if time.Since(lastFail) > c.resetTimeout {
			c.state.Store("half-open")
			c.failures.Store(0)
		} else {
			return nil, errors.New("circuit breaker open")
		}
	}

	// Make request
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.client.Do(req)

	if err != nil || resp.StatusCode >= 500 {
		failures := c.failures.Add(1)
		c.lastFailTime.Store(time.Now().UnixNano())

		if failures >= int32(c.maxFailures) {
			c.state.Store("open")
		}

		if err != nil {
			return nil, err
		}
		return resp, nil
	}

	// Success
	if c.state.Load().(string) == "half-open" {
		c.state.Store("closed")
		c.failures.Store(0)
	}

	return resp, nil
}

type DegradableClient struct {
	primary  string
	fallback string
	client   *http.Client
}

func (c *DegradableClient) Get(ctx context.Context) (*http.Response, error) {
	// Try primary first
	req, err := http.NewRequestWithContext(ctx, "GET", c.primary, nil)
	if err == nil {
		resp, err := c.client.Do(req)
		if err == nil && resp.StatusCode == http.StatusOK {
			return resp, nil
		}
		if resp != nil {
			resp.Body.Close()
		}
	}

	// Fallback
	req, err = http.NewRequestWithContext(ctx, "GET", c.fallback, nil)
	if err != nil {
		return nil, err
	}

	return c.client.Do(req)
}

type FailoverClient struct {
	servers []string
	client  *http.Client
	current atomic.Int32
}

func (c *FailoverClient) Get(ctx context.Context) (*http.Response, error) {
	var lastErr error

	for i := 0; i < len(c.servers); i++ {
		serverIdx := (int(c.current.Load()) + i) % len(c.servers)
		url := c.servers[serverIdx]

		req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
		if err != nil {
			lastErr = err
			continue
		}

		resp, err := c.client.Do(req)
		if err != nil {
			lastErr = err
			continue
		}

		if resp.StatusCode == http.StatusOK {
			c.current.Store(int32(serverIdx))
			return resp, nil
		}

		resp.Body.Close()
		lastErr = fmt.Errorf("server %d returned %d", serverIdx, resp.StatusCode)
	}

	return nil, fmt.Errorf("all servers failed: %w", lastErr)
}

type HealthMonitor struct {
	url      string
	interval time.Duration
	healthy  atomic.Bool
	stop     chan bool
	client   *http.Client
}

func NewHealthMonitor(url string, interval time.Duration) *HealthMonitor {
	return &HealthMonitor{
		url:      url,
		interval: interval,
		stop:     make(chan bool),
		client:   &http.Client{Timeout: 5 * time.Second},
	}
}

func (m *HealthMonitor) Start() {
	go func() {
		ticker := time.NewTicker(m.interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				m.checkHealth()
			case <-m.stop:
				return
			}
		}
	}()

	// Initial check
	m.checkHealth()
}

func (m *HealthMonitor) Stop() {
	close(m.stop)
}

func (m *HealthMonitor) IsHealthy() bool {
	return m.healthy.Load()
}

func (m *HealthMonitor) checkHealth() {
	resp, err := m.client.Get(m.url + "/health")
	if err != nil {
		m.healthy.Store(false)
		return
	}
	defer resp.Body.Close()

	m.healthy.Store(resp.StatusCode == http.StatusOK)
}

type AdaptiveTimeoutClient struct {
	baseTimeout time.Duration
	maxTimeout  time.Duration
	timeout     atomic.Int64
	samples     []time.Duration
	mu          sync.Mutex
}

func NewAdaptiveTimeoutClient(baseTimeout, maxTimeout time.Duration) *AdaptiveTimeoutClient {
	c := &AdaptiveTimeoutClient{
		baseTimeout: baseTimeout,
		maxTimeout:  maxTimeout,
		samples:     make([]time.Duration, 0, 10),
	}
	c.timeout.Store(int64(baseTimeout))
	return c
}

func (c *AdaptiveTimeoutClient) Get(ctx context.Context, url string) (*http.Response, error) {
	client := &http.Client{
		Timeout: time.Duration(c.timeout.Load()),
	}

	start := time.Now()
	resp, err := client.Get(url)
	duration := time.Since(start)

	// Update timeout based on response time
	c.updateTimeout(duration)

	return resp, err
}

func (c *AdaptiveTimeoutClient) GetTimeout() time.Duration {
	return time.Duration(c.timeout.Load())
}

func (c *AdaptiveTimeoutClient) updateTimeout(duration time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.samples = append(c.samples, duration)
	if len(c.samples) > 10 {
		c.samples = c.samples[1:]
	}

	// Calculate average
	var total time.Duration
	for _, d := range c.samples {
		total += d
	}
	avg := total / time.Duration(len(c.samples))

	// New timeout is 2x average, bounded
	newTimeout := avg * 2
	if newTimeout < c.baseTimeout {
		newTimeout = c.baseTimeout
	}
	if newTimeout > c.maxTimeout {
		newTimeout = c.maxTimeout
	}

	c.timeout.Store(int64(newTimeout))
}
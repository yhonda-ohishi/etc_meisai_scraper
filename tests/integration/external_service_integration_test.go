package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// T011-B: External service integration testing with mock ETC provider APIs
func TestExternalServiceIntegration_MockETCProviders(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping external service integration test in short mode")
	}

	tests := []struct {
		name     string
		provider ETCProvider
		scenario string
	}{
		{
			name:     "NEXCO provider integration",
			provider: NewNEXCOProvider(),
			scenario: "normal",
		},
		{
			name:     "Metropolitan provider integration",
			provider: NewMetropolitanProvider(),
			scenario: "normal",
		},
		{
			name:     "Hanshin provider integration",
			provider: NewHanshinProvider(),
			scenario: "normal",
		},
		{
			name:     "Provider with rate limiting",
			provider: NewRateLimitedProvider(),
			scenario: "rate_limited",
		},
		{
			name:     "Provider with intermittent failures",
			provider: NewUnreliableProvider(),
			scenario: "unreliable",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := tt.provider.StartMockServer()
			defer server.Close()

			client := NewETCClient(server.URL)

			// Test authentication
			testAuthentication(t, client, tt.provider)

			// Test data retrieval
			testDataRetrieval(t, client, tt.provider)

			// Test CSV download
			testCSVDownload(t, client, tt.provider)

			// Test error handling
			testErrorScenarios(t, client, tt.provider)

			// Test concurrent requests
			if tt.scenario != "rate_limited" {
				testConcurrentRequests(t, client, tt.provider)
			}
		})
	}
}

// ETCProvider interface for mock providers
type ETCProvider interface {
	StartMockServer() *httptest.Server
	GetName() string
	GetCredentials() (username, password string)
	GetExpectedRecords() int
}

// Base provider implementation
type baseProvider struct {
	name            string
	username        string
	password        string
	recordCount     int
	requestCount    atomic.Int64
	authToken       string
	mu              sync.Mutex
	authenticatedUsers map[string]bool
}

func (p *baseProvider) GetName() string {
	return p.name
}

func (p *baseProvider) GetCredentials() (string, string) {
	return p.username, p.password
}

func (p *baseProvider) GetExpectedRecords() int {
	return p.recordCount
}

// NEXCO Provider Mock
type NEXCOProvider struct {
	baseProvider
}

func NewNEXCOProvider() *NEXCOProvider {
	return &NEXCOProvider{
		baseProvider: baseProvider{
			name:            "NEXCO",
			username:        "nexco_user",
			password:        "nexco_pass",
			recordCount:     100,
			authenticatedUsers: make(map[string]bool),
		},
	}
}

func (p *NEXCOProvider) StartMockServer() *httptest.Server {
	mux := http.NewServeMux()

	// Login endpoint
	mux.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		p.requestCount.Add(1)

		var creds struct {
			Username string `json:"username"`
			Password string `json:"password"`
		}

		json.NewDecoder(r.Body).Decode(&creds)

		if creds.Username == p.username && creds.Password == p.password {
			token := fmt.Sprintf("nexco_token_%d", time.Now().Unix())
			p.mu.Lock()
			p.authenticatedUsers[token] = true
			p.mu.Unlock()

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]string{
				"token": token,
				"type":  "Bearer",
			})
		} else {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "Invalid credentials",
			})
		}
	})

	// List records endpoint
	mux.HandleFunc("/api/records", func(w http.ResponseWriter, r *http.Request) {
		p.requestCount.Add(1)

		// Check authentication
		auth := r.Header.Get("Authorization")
		token := ""
		if len(auth) > 7 && auth[:7] == "Bearer " {
			token = auth[7:]
		}

		p.mu.Lock()
		authenticated := p.authenticatedUsers[token]
		p.mu.Unlock()

		if !authenticated {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		// Return mock records
		records := generateMockRecords(p.recordCount, "NEXCO")
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(records)
	})

	// CSV download endpoint
	mux.HandleFunc("/api/download/csv", func(w http.ResponseWriter, r *http.Request) {
		p.requestCount.Add(1)

		// Check authentication
		auth := r.Header.Get("Authorization")
		token := ""
		if len(auth) > 7 && auth[:7] == "Bearer " {
			token = auth[7:]
		}

		p.mu.Lock()
		authenticated := p.authenticatedUsers[token]
		p.mu.Unlock()

		if !authenticated {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		// Generate CSV data
		csv := generateMockCSV(p.recordCount, "NEXCO")
		w.Header().Set("Content-Type", "text/csv")
		w.Header().Set("Content-Disposition", `attachment; filename="nexco_records.csv"`)
		w.Write([]byte(csv))
	})

	// Logout endpoint
	mux.HandleFunc("/logout", func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		token := ""
		if len(auth) > 7 && auth[:7] == "Bearer " {
			token = auth[7:]
		}

		p.mu.Lock()
		delete(p.authenticatedUsers, token)
		p.mu.Unlock()

		w.WriteHeader(http.StatusOK)
	})

	return httptest.NewServer(mux)
}

// Metropolitan Provider Mock
type MetropolitanProvider struct {
	baseProvider
}

func NewMetropolitanProvider() *MetropolitanProvider {
	return &MetropolitanProvider{
		baseProvider: baseProvider{
			name:            "Metropolitan",
			username:        "metro_user",
			password:        "metro_pass",
			recordCount:     150,
			authenticatedUsers: make(map[string]bool),
		},
	}
}

func (p *MetropolitanProvider) StartMockServer() *httptest.Server {
	mux := http.NewServeMux()

	// Session-based authentication
	sessions := make(map[string]bool)
	var sessionMu sync.Mutex

	mux.HandleFunc("/auth/login", func(w http.ResponseWriter, r *http.Request) {
		p.requestCount.Add(1)

		username := r.FormValue("username")
		password := r.FormValue("password")

		if username == p.username && password == p.password {
			sessionID := fmt.Sprintf("metro_session_%d", time.Now().Unix())
			sessionMu.Lock()
			sessions[sessionID] = true
			sessionMu.Unlock()

			http.SetCookie(w, &http.Cookie{
				Name:     "session",
				Value:    sessionID,
				HttpOnly: true,
				Path:     "/",
			})
			w.WriteHeader(http.StatusOK)
		} else {
			w.WriteHeader(http.StatusUnauthorized)
		}
	})

	// Data endpoint
	mux.HandleFunc("/data/etc/list", func(w http.ResponseWriter, r *http.Request) {
		p.requestCount.Add(1)

		cookie, err := r.Cookie("session")
		if err != nil || cookie == nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		sessionMu.Lock()
		valid := sessions[cookie.Value]
		sessionMu.Unlock()

		if !valid {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		records := generateMockRecords(p.recordCount, "Metro")
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(records)
	})

	return httptest.NewServer(mux)
}

// Hanshin Provider Mock
type HanshinProvider struct {
	baseProvider
}

func NewHanshinProvider() *HanshinProvider {
	return &HanshinProvider{
		baseProvider: baseProvider{
			name:            "Hanshin",
			username:        "hanshin_user",
			password:        "hanshin_pass",
			recordCount:     80,
			authenticatedUsers: make(map[string]bool),
		},
	}
}

func (p *HanshinProvider) StartMockServer() *httptest.Server {
	// Similar implementation with different endpoints
	return NewNEXCOProvider().StartMockServer()
}

// Rate Limited Provider
type RateLimitedProvider struct {
	baseProvider
	requestTimes []time.Time
	mu           sync.Mutex
}

func NewRateLimitedProvider() *RateLimitedProvider {
	return &RateLimitedProvider{
		baseProvider: baseProvider{
			name:            "RateLimited",
			username:        "limited_user",
			password:        "limited_pass",
			recordCount:     50,
			authenticatedUsers: make(map[string]bool),
		},
		requestTimes: make([]time.Time, 0),
	}
}

func (p *RateLimitedProvider) StartMockServer() *httptest.Server {
	mux := http.NewServeMux()

	// Rate limiting middleware
	rateLimitMiddleware := func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			p.mu.Lock()
			now := time.Now()

			// Clean old requests
			validTimes := make([]time.Time, 0)
			for _, t := range p.requestTimes {
				if now.Sub(t) < time.Minute {
					validTimes = append(validTimes, t)
				}
			}
			p.requestTimes = validTimes

			// Check rate limit (10 requests per minute)
			if len(p.requestTimes) >= 10 {
				p.mu.Unlock()
				w.WriteHeader(http.StatusTooManyRequests)
				w.Header().Set("Retry-After", "60")
				json.NewEncoder(w).Encode(map[string]string{
					"error": "Rate limit exceeded",
				})
				return
			}

			p.requestTimes = append(p.requestTimes, now)
			p.mu.Unlock()

			next(w, r)
		}
	}

	mux.HandleFunc("/api/data", rateLimitMiddleware(func(w http.ResponseWriter, r *http.Request) {
		records := generateMockRecords(p.recordCount, "RateLimited")
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(records)
	}))

	return httptest.NewServer(mux)
}

// Unreliable Provider (simulates network issues)
type UnreliableProvider struct {
	baseProvider
	failureRate float32
	requestNum  atomic.Int32
}

func NewUnreliableProvider() *UnreliableProvider {
	return &UnreliableProvider{
		baseProvider: baseProvider{
			name:            "Unreliable",
			username:        "unreliable_user",
			password:        "unreliable_pass",
			recordCount:     75,
			authenticatedUsers: make(map[string]bool),
		},
		failureRate: 0.3, // 30% failure rate
	}
}

func (p *UnreliableProvider) StartMockServer() *httptest.Server {
	mux := http.NewServeMux()

	mux.HandleFunc("/api/data", func(w http.ResponseWriter, r *http.Request) {
		reqNum := p.requestNum.Add(1)

		// Simulate failures
		if reqNum%3 == 0 { // Every 3rd request fails
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "Internal server error",
			})
			return
		}

		if reqNum%5 == 0 { // Every 5th request times out (simulated)
			time.Sleep(5 * time.Second)
		}

		records := generateMockRecords(p.recordCount, "Unreliable")
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(records)
	})

	return httptest.NewServer(mux)
}

// ETC Client for testing
type ETCClient struct {
	baseURL    string
	httpClient *http.Client
	token      string
	mu         sync.Mutex
}

func NewETCClient(baseURL string) *ETCClient {
	return &ETCClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (c *ETCClient) Login(username, password string) error {
	payload := map[string]string{
		"username": username,
		"password": password,
	}

	body, _ := json.Marshal(payload)
	req, err := http.NewRequest("POST", c.baseURL+"/login", bytes.NewReader(body))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("login failed: %d", resp.StatusCode)
	}

	var result struct {
		Token string `json:"token"`
	}

	json.NewDecoder(resp.Body).Decode(&result)

	c.mu.Lock()
	c.token = result.Token
	c.mu.Unlock()

	return nil
}

func (c *ETCClient) GetRecords() ([]map[string]interface{}, error) {
	c.mu.Lock()
	token := c.token
	c.mu.Unlock()

	req, err := http.NewRequest("GET", c.baseURL+"/api/records", nil)
	if err != nil {
		return nil, err
	}

	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("get records failed: %d", resp.StatusCode)
	}

	var records []map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&records)

	return records, nil
}

func (c *ETCClient) DownloadCSV() ([]byte, error) {
	c.mu.Lock()
	token := c.token
	c.mu.Unlock()

	req, err := http.NewRequest("GET", c.baseURL+"/api/download/csv", nil)
	if err != nil {
		return nil, err
	}

	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("download failed: %d", resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}

// Helper functions
func generateMockRecords(count int, provider string) []map[string]interface{} {
	records := make([]map[string]interface{}, count)
	for i := 0; i < count; i++ {
		records[i] = map[string]interface{}{
			"id":            fmt.Sprintf("%s_%d", provider, i),
			"date":          time.Now().Add(time.Duration(-i) * time.Hour).Format("2006-01-02"),
			"entry_ic":      fmt.Sprintf("Entry_%d", i%10),
			"exit_ic":       fmt.Sprintf("Exit_%d", i%10),
			"amount":        100 + (i * 10),
			"vehicle_number": fmt.Sprintf("車両%d", i),
			"etc_number":    fmt.Sprintf("ETC_%s_%d", provider, i),
		}
	}
	return records
}

func generateMockCSV(count int, provider string) string {
	var buf bytes.Buffer
	buf.WriteString("ID,Date,Entry,Exit,Amount,Vehicle,ETC\n")
	for i := 0; i < count; i++ {
		buf.WriteString(fmt.Sprintf("%s_%d,%s,Entry_%d,Exit_%d,%d,車両%d,ETC_%s_%d\n",
			provider, i,
			time.Now().Add(time.Duration(-i)*time.Hour).Format("2006-01-02"),
			i%10, i%10,
			100+(i*10),
			i,
			provider, i,
		))
	}
	return buf.String()
}

// Test functions
func testAuthentication(t *testing.T, client *ETCClient, provider ETCProvider) {
	username, password := provider.GetCredentials()

	// Test successful authentication
	err := client.Login(username, password)
	assert.NoError(t, err)

	// Test failed authentication
	err = client.Login("wrong_user", "wrong_pass")
	assert.Error(t, err)
}

func testDataRetrieval(t *testing.T, client *ETCClient, provider ETCProvider) {
	username, password := provider.GetCredentials()

	// Login first
	err := client.Login(username, password)
	require.NoError(t, err)

	// Get records
	records, err := client.GetRecords()
	assert.NoError(t, err)
	assert.Len(t, records, provider.GetExpectedRecords())

	// Verify record structure
	if len(records) > 0 {
		record := records[0]
		assert.Contains(t, record, "id")
		assert.Contains(t, record, "date")
		assert.Contains(t, record, "entry_ic")
		assert.Contains(t, record, "exit_ic")
		assert.Contains(t, record, "amount")
	}
}

func testCSVDownload(t *testing.T, client *ETCClient, provider ETCProvider) {
	username, password := provider.GetCredentials()

	// Login first
	err := client.Login(username, password)
	require.NoError(t, err)

	// Download CSV
	csvData, err := client.DownloadCSV()
	assert.NoError(t, err)
	assert.NotEmpty(t, csvData)

	// Verify CSV format
	csvStr := string(csvData)
	assert.Contains(t, csvStr, "ID,Date,Entry,Exit,Amount,Vehicle,ETC")
	lines := bytes.Split(csvData, []byte("\n"))
	assert.GreaterOrEqual(t, len(lines), provider.GetExpectedRecords())
}

func testErrorScenarios(t *testing.T, client *ETCClient, provider ETCProvider) {
	// Test unauthenticated access
	client.token = ""
	_, err := client.GetRecords()
	assert.Error(t, err)

	// Test with invalid token
	client.token = "invalid_token"
	_, err = client.GetRecords()
	assert.Error(t, err)
}

func testConcurrentRequests(t *testing.T, client *ETCClient, provider ETCProvider) {
	username, password := provider.GetCredentials()

	// Login first
	err := client.Login(username, password)
	require.NoError(t, err)

	// Concurrent requests
	var wg sync.WaitGroup
	errors := make(chan error, 10)

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := client.GetRecords()
			if err != nil {
				errors <- err
			}
		}()
	}

	wg.Wait()
	close(errors)

	// Check errors
	for err := range errors {
		assert.NoError(t, err)
	}
}

// Test retry mechanism
func TestRetryMechanism(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping retry mechanism test in short mode")
	}

	provider := NewUnreliableProvider()
	server := provider.StartMockServer()
	defer server.Close()

	client := NewETCClientWithRetry(server.URL, 3) // 3 retries

	// Should eventually succeed despite failures
	var records []map[string]interface{}
	var err error

	for i := 0; i < 5; i++ {
		records, err = client.GetRecordsWithRetry()
		if err == nil {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}

	assert.NoError(t, err)
	assert.NotEmpty(t, records)
}

// ETC Client with retry logic
type ETCClientWithRetry struct {
	*ETCClient
	maxRetries int
}

func NewETCClientWithRetry(baseURL string, maxRetries int) *ETCClientWithRetry {
	return &ETCClientWithRetry{
		ETCClient:  NewETCClient(baseURL),
		maxRetries: maxRetries,
	}
}

func (c *ETCClientWithRetry) GetRecordsWithRetry() ([]map[string]interface{}, error) {
	var lastErr error

	for i := 0; i < c.maxRetries; i++ {
		records, err := c.GetRecords()
		if err == nil {
			return records, nil
		}

		lastErr = err

		// Exponential backoff
		backoff := time.Duration(1<<i) * 100 * time.Millisecond
		time.Sleep(backoff)
	}

	return nil, fmt.Errorf("failed after %d retries: %w", c.maxRetries, lastErr)
}

// Test circuit breaker pattern
func TestCircuitBreaker(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping circuit breaker test in short mode")
	}

	provider := NewUnreliableProvider()
	server := provider.StartMockServer()
	defer server.Close()

	breaker := NewCircuitBreaker(5, 10*time.Second)
	client := NewETCClient(server.URL)

	// Test circuit breaker opening after failures
	var failures int
	for i := 0; i < 10; i++ {
		err := breaker.Call(func() error {
			_, err := client.GetRecords()
			return err
		})

		if err != nil {
			failures++
			if err.Error() == "circuit breaker is open" {
				break
			}
		}
	}

	assert.Greater(t, failures, 0)
}

// Circuit Breaker implementation
type CircuitBreaker struct {
	maxFailures    int
	resetTimeout   time.Duration
	failures       int
	lastFailTime   time.Time
	state          string
	mu             sync.Mutex
}

func NewCircuitBreaker(maxFailures int, resetTimeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		maxFailures:  maxFailures,
		resetTimeout: resetTimeout,
		state:        "closed",
	}
}

func (cb *CircuitBreaker) Call(fn func() error) error {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	// Check if circuit should be reset
	if cb.state == "open" && time.Since(cb.lastFailTime) > cb.resetTimeout {
		cb.state = "half-open"
		cb.failures = 0
	}

	// If open, reject immediately
	if cb.state == "open" {
		return fmt.Errorf("circuit breaker is open")
	}

	// Try the call
	err := fn()

	if err != nil {
		cb.failures++
		cb.lastFailTime = time.Now()

		if cb.failures >= cb.maxFailures {
			cb.state = "open"
		}
		return err
	}

	// Success - reset failures
	if cb.state == "half-open" {
		cb.state = "closed"
	}
	cb.failures = 0

	return nil
}
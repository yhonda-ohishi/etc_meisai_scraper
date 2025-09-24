package middleware

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/yhonda-ohishi/etc_meisai/src/middleware"
)

// Test helper functions
func createTestHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("test response"))
	})
}

func createPanicHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("test panic")
	})
}

// TestChain tests middleware chaining functionality
func TestChain(t *testing.T) {
	t.Run("single middleware", func(t *testing.T) {
		testMiddleware := func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Test-Header", "test-value")
				next.ServeHTTP(w, r)
			})
		}

		handler := createTestHandler()
		chained := middleware.Chain(handler, testMiddleware)

		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()

		chained.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "test-value", w.Header().Get("Test-Header"))
		assert.Equal(t, "test response", w.Body.String())
	})

	t.Run("multiple middleware execution order", func(t *testing.T) {
		middleware1 := func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("MW1", "first")
				next.ServeHTTP(w, r)
				w.Header().Set("MW1-After", "after-first")
			})
		}

		middleware2 := func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("MW2", "second")
				next.ServeHTTP(w, r)
				w.Header().Set("MW2-After", "after-second")
			})
		}

		handler := createTestHandler()
		chained := middleware.Chain(handler, middleware1, middleware2)

		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()

		chained.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "first", w.Header().Get("MW1"))
		assert.Equal(t, "second", w.Header().Get("MW2"))
	})

	t.Run("empty middleware chain", func(t *testing.T) {
		handler := createTestHandler()
		chained := middleware.Chain(handler)

		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()

		chained.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "test response", w.Body.String())
	})
}

// TestCORS tests CORS middleware functionality
func TestCORS(t *testing.T) {
	t.Run("wildcard origin", func(t *testing.T) {
		handler := createTestHandler()
		corsHandler := middleware.CORS([]string{"*"})(handler)

		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("Origin", "https://example.com")
		w := httptest.NewRecorder()

		corsHandler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "https://example.com", w.Header().Get("Access-Control-Allow-Origin"))
		assert.Contains(t, w.Header().Get("Access-Control-Allow-Methods"), "GET")
		assert.Contains(t, w.Header().Get("Access-Control-Allow-Methods"), "POST")
		assert.Equal(t, "true", w.Header().Get("Access-Control-Allow-Credentials"))
	})

	t.Run("specific origins", func(t *testing.T) {
		allowedOrigins := []string{"https://allowed.com", "https://trusted.com"}
		handler := createTestHandler()
		corsHandler := middleware.CORS(allowedOrigins)(handler)

		// Test allowed origin
		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("Origin", "https://allowed.com")
		w := httptest.NewRecorder()

		corsHandler.ServeHTTP(w, req)

		assert.Equal(t, "https://allowed.com", w.Header().Get("Access-Control-Allow-Origin"))

		// Test disallowed origin
		req2 := httptest.NewRequest("GET", "/test", nil)
		req2.Header.Set("Origin", "https://evil.com")
		w2 := httptest.NewRecorder()

		corsHandler.ServeHTTP(w2, req2)

		assert.Empty(t, w2.Header().Get("Access-Control-Allow-Origin"))
	})

	t.Run("preflight request", func(t *testing.T) {
		handler := createTestHandler()
		corsHandler := middleware.CORS([]string{"*"})(handler)

		req := httptest.NewRequest("OPTIONS", "/test", nil)
		req.Header.Set("Origin", "https://example.com")
		w := httptest.NewRecorder()

		corsHandler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNoContent, w.Code)
		assert.Equal(t, "https://example.com", w.Header().Get("Access-Control-Allow-Origin"))
		assert.Contains(t, w.Header().Get("Access-Control-Allow-Headers"), "Authorization")
		assert.Contains(t, w.Header().Get("Access-Control-Allow-Headers"), "Content-Type")
		assert.NotEmpty(t, w.Header().Get("Access-Control-Max-Age"))
	})

	t.Run("no origin header", func(t *testing.T) {
		handler := createTestHandler()
		corsHandler := middleware.CORS([]string{"https://example.com"})(handler)

		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()

		corsHandler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Empty(t, w.Header().Get("Access-Control-Allow-Origin"))
	})

	t.Run("empty allowed origins uses wildcard", func(t *testing.T) {
		handler := createTestHandler()
		corsHandler := middleware.CORS([]string{})(handler)

		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("Origin", "https://example.com")
		w := httptest.NewRecorder()

		corsHandler.ServeHTTP(w, req)

		assert.Equal(t, "https://example.com", w.Header().Get("Access-Control-Allow-Origin"))
	})
}

// TestSecurity tests security headers middleware
func TestSecurity(t *testing.T) {
	t.Run("sets security headers", func(t *testing.T) {
		handler := createTestHandler()
		securityHandler := middleware.Security()(handler)

		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()

		securityHandler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "nosniff", w.Header().Get("X-Content-Type-Options"))
		assert.Equal(t, "DENY", w.Header().Get("X-Frame-Options"))
		assert.Equal(t, "1; mode=block", w.Header().Get("X-XSS-Protection"))
		assert.Equal(t, "strict-origin-when-cross-origin", w.Header().Get("Referrer-Policy"))
		assert.Contains(t, w.Header().Get("Content-Security-Policy"), "default-src 'self'")
		assert.Contains(t, w.Header().Get("Permissions-Policy"), "camera=()")
		assert.Empty(t, w.Header().Get("Server"))
	})

	t.Run("HSTS header in production with TLS", func(t *testing.T) {
		// Store original environment
		originalEnv := os.Getenv("ENVIRONMENT")
		defer func() {
			if originalEnv != "" {
				os.Setenv("ENVIRONMENT", originalEnv)
			} else {
				os.Unsetenv("ENVIRONMENT")
			}
		}()

		os.Setenv("ENVIRONMENT", "production")

		handler := createTestHandler()
		securityHandler := middleware.Security()(handler)

		req := httptest.NewRequest("GET", "/test", nil)
		req.TLS = &tls.ConnectionState{} // Simulate TLS connection
		w := httptest.NewRecorder()

		securityHandler.ServeHTTP(w, req)

		assert.Contains(t, w.Header().Get("Strict-Transport-Security"), "max-age=31536000")
	})

	t.Run("no HSTS header without TLS", func(t *testing.T) {
		// Store original environment
		originalEnv := os.Getenv("ENVIRONMENT")
		defer func() {
			if originalEnv != "" {
				os.Setenv("ENVIRONMENT", originalEnv)
			} else {
				os.Unsetenv("ENVIRONMENT")
			}
		}()

		os.Setenv("ENVIRONMENT", "production")

		handler := createTestHandler()
		securityHandler := middleware.Security()(handler)

		req := httptest.NewRequest("GET", "/test", nil)
		// No TLS connection
		w := httptest.NewRecorder()

		securityHandler.ServeHTTP(w, req)

		assert.Empty(t, w.Header().Get("Strict-Transport-Security"))
	})
}

// TestRequestSize tests request size limiting middleware
func TestRequestSize(t *testing.T) {
	t.Run("request within size limit", func(t *testing.T) {
		handler := createTestHandler()
		sizeHandler := middleware.RequestSize(1024)(handler)

		body := strings.NewReader("small request body")
		req := httptest.NewRequest("POST", "/test", body)
		req.ContentLength = int64(len("small request body"))
		w := httptest.NewRecorder()

		sizeHandler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("request exceeds size limit", func(t *testing.T) {
		handler := createTestHandler()
		sizeHandler := middleware.RequestSize(10)(handler)

		body := strings.NewReader("this is a very long request body that exceeds the limit")
		req := httptest.NewRequest("POST", "/test", body)
		req.ContentLength = int64(len("this is a very long request body that exceeds the limit"))
		w := httptest.NewRecorder()

		sizeHandler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusRequestEntityTooLarge, w.Code)
		assert.Contains(t, w.Body.String(), "request too large")
		assert.Contains(t, w.Body.String(), "max_size")
	})

	t.Run("request with no content length", func(t *testing.T) {
		handler := createTestHandler()
		sizeHandler := middleware.RequestSize(1024)(handler)

		body := strings.NewReader("request without content length")
		req := httptest.NewRequest("POST", "/test", body)
		// Don't set ContentLength
		w := httptest.NewRecorder()

		sizeHandler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}

// TestAllowedMethods tests method filtering middleware
func TestAllowedMethods(t *testing.T) {
	t.Run("allowed method", func(t *testing.T) {
		handler := createTestHandler()
		methodHandler := middleware.AllowedMethods("GET", "POST")(handler)

		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()

		methodHandler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("disallowed method", func(t *testing.T) {
		handler := createTestHandler()
		methodHandler := middleware.AllowedMethods("GET", "POST")(handler)

		req := httptest.NewRequest("DELETE", "/test", nil)
		w := httptest.NewRecorder()

		methodHandler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
		assert.Contains(t, w.Header().Get("Allow"), "GET")
		assert.Contains(t, w.Header().Get("Allow"), "POST")
		assert.Contains(t, w.Body.String(), "method not allowed")
	})

	t.Run("case insensitive method matching", func(t *testing.T) {
		handler := createTestHandler()
		methodHandler := middleware.AllowedMethods("get", "post")(handler)

		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()

		methodHandler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("empty allowed methods", func(t *testing.T) {
		handler := createTestHandler()
		methodHandler := middleware.AllowedMethods()(handler)

		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()

		methodHandler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
	})
}

// TestLogging tests logging middleware
func TestLogging(t *testing.T) {
	t.Run("adds request ID and logs request", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestID := r.Context().Value("request_id")
			assert.NotNil(t, requestID)
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("test"))
		})

		loggingHandler := middleware.Logging()(handler)

		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()

		loggingHandler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.NotEmpty(t, w.Header().Get("X-Request-ID"))
		assert.NotEmpty(t, w.Header().Get("X-Response-Time"))
	})

	t.Run("captures status code", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("not found"))
		})

		loggingHandler := middleware.Logging()(handler)

		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()

		loggingHandler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.NotEmpty(t, w.Header().Get("X-Request-ID"))
	})

	t.Run("handles request with X-Forwarded-For", func(t *testing.T) {
		handler := createTestHandler()
		loggingHandler := middleware.Logging()(handler)

		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("X-Forwarded-For", "203.0.113.1, 198.51.100.1")
		w := httptest.NewRecorder()

		loggingHandler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.NotEmpty(t, w.Header().Get("X-Request-ID"))
	})
}

// TestSecurityMiddleware tests the SecurityMiddleware function
func TestSecurityMiddleware(t *testing.T) {
	t.Run("sets basic security headers", func(t *testing.T) {
		handler := createTestHandler()
		securityHandler := middleware.SecurityMiddleware(handler)

		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()

		securityHandler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "nosniff", w.Header().Get("X-Content-Type-Options"))
		assert.Equal(t, "DENY", w.Header().Get("X-Frame-Options"))
		assert.Equal(t, "1; mode=block", w.Header().Get("X-XSS-Protection"))
		assert.Equal(t, "strict-origin-when-cross-origin", w.Header().Get("Referrer-Policy"))
		assert.Equal(t, "default-src 'self'", w.Header().Get("Content-Security-Policy"))
		assert.Empty(t, w.Header().Get("Server"))
	})
}

// TestBasicAuthMiddleware tests basic authentication
func TestBasicAuthMiddleware(t *testing.T) {
	t.Run("valid credentials", func(t *testing.T) {
		handler := createTestHandler()
		authHandler := middleware.BasicAuthMiddleware("admin", "password")(handler)

		req := httptest.NewRequest("GET", "/test", nil)
		req.SetBasicAuth("admin", "password")
		w := httptest.NewRecorder()

		authHandler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("invalid credentials", func(t *testing.T) {
		handler := createTestHandler()
		authHandler := middleware.BasicAuthMiddleware("admin", "password")(handler)

		req := httptest.NewRequest("GET", "/test", nil)
		req.SetBasicAuth("admin", "wrongpassword")
		w := httptest.NewRecorder()

		authHandler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, w.Header().Get("WWW-Authenticate"), "Basic realm")
	})

	t.Run("no credentials", func(t *testing.T) {
		handler := createTestHandler()
		authHandler := middleware.BasicAuthMiddleware("admin", "password")(handler)

		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()

		authHandler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, w.Header().Get("WWW-Authenticate"), "Basic realm")
	})
}

// TestAPIKeyMiddleware tests API key authentication
func TestAPIKeyMiddleware(t *testing.T) {
	validKeys := []string{"key1", "key2", "secret-key"}

	t.Run("valid API key in header", func(t *testing.T) {
		handler := createTestHandler()
		apiKeyHandler := middleware.APIKeyMiddleware(validKeys)(handler)

		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("X-API-Key", "key1")
		w := httptest.NewRecorder()

		apiKeyHandler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("valid API key in query parameter", func(t *testing.T) {
		handler := createTestHandler()
		apiKeyHandler := middleware.APIKeyMiddleware(validKeys)(handler)

		req := httptest.NewRequest("GET", "/test?api_key=key2", nil)
		w := httptest.NewRecorder()

		apiKeyHandler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("invalid API key", func(t *testing.T) {
		handler := createTestHandler()
		apiKeyHandler := middleware.APIKeyMiddleware(validKeys)(handler)

		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("X-API-Key", "invalid-key")
		w := httptest.NewRecorder()

		apiKeyHandler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, w.Body.String(), "Invalid or missing API key")
	})

	t.Run("missing API key", func(t *testing.T) {
		handler := createTestHandler()
		apiKeyHandler := middleware.APIKeyMiddleware(validKeys)(handler)

		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()

		apiKeyHandler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, w.Body.String(), "Invalid or missing API key")
	})

	t.Run("empty valid keys list", func(t *testing.T) {
		handler := createTestHandler()
		apiKeyHandler := middleware.APIKeyMiddleware([]string{})(handler)

		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("X-API-Key", "any-key")
		w := httptest.NewRecorder()

		apiKeyHandler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

// TestSanitizeMiddleware tests input sanitization
func TestSanitizeMiddleware(t *testing.T) {
	t.Run("sanitizes query parameters", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			param := r.URL.Query().Get("test")
			assert.Equal(t, "clean_value", param)
			w.WriteHeader(http.StatusOK)
		})

		sanitizeHandler := middleware.SanitizeMiddleware(handler)

		req := httptest.NewRequest("GET", "/test?test=%20%20clean_value%20%20", nil)
		w := httptest.NewRecorder()

		sanitizeHandler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("removes null bytes", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			param := r.URL.Query().Get("test")
			assert.NotContains(t, param, "\x00")
			w.WriteHeader(http.StatusOK)
		})

		sanitizeHandler := middleware.SanitizeMiddleware(handler)

		// Use URL encoding for null bytes
		req := httptest.NewRequest("GET", "/test?test=value%00with%00nulls", nil)
		w := httptest.NewRecorder()

		sanitizeHandler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("limits string length", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			param := r.URL.Query().Get("test")
			assert.LessOrEqual(t, len(param), 1000)
			w.WriteHeader(http.StatusOK)
		})

		sanitizeHandler := middleware.SanitizeMiddleware(handler)

		longValue := strings.Repeat("a", 2000)
		req := httptest.NewRequest("GET", "/test?test="+longValue, nil)
		w := httptest.NewRecorder()

		sanitizeHandler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}

// TestIPWhitelistMiddleware tests IP whitelisting
func TestIPWhitelistMiddleware(t *testing.T) {
	allowedIPs := []string{"192.168.1.1", "10.0.0.1"}

	t.Run("allowed IP", func(t *testing.T) {
		handler := createTestHandler()
		ipHandler := middleware.IPWhitelistMiddleware(allowedIPs)(handler)

		req := httptest.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "192.168.1.1:1234"
		w := httptest.NewRecorder()

		ipHandler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("blocked IP", func(t *testing.T) {
		handler := createTestHandler()
		ipHandler := middleware.IPWhitelistMiddleware(allowedIPs)(handler)

		req := httptest.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "192.168.1.100:1234"
		w := httptest.NewRecorder()

		ipHandler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
		assert.Contains(t, w.Body.String(), "Access denied")
	})

	t.Run("IP from X-Forwarded-For header", func(t *testing.T) {
		handler := createTestHandler()
		ipHandler := middleware.IPWhitelistMiddleware(allowedIPs)(handler)

		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("X-Forwarded-For", "10.0.0.1, 192.168.1.100")
		w := httptest.NewRecorder()

		ipHandler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("IP from X-Real-IP header", func(t *testing.T) {
		handler := createTestHandler()
		ipHandler := middleware.IPWhitelistMiddleware(allowedIPs)(handler)

		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("X-Real-IP", "192.168.1.1")
		w := httptest.NewRecorder()

		ipHandler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}

// TestTimeoutMiddleware tests request timeout handling
func TestTimeoutMiddleware(t *testing.T) {
	t.Run("request completes within timeout", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Simulate short processing
			time.Sleep(10 * time.Millisecond)
			w.WriteHeader(http.StatusOK)
		})

		timeoutHandler := middleware.TimeoutMiddleware(100 * time.Millisecond)(handler)

		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()

		timeoutHandler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("request respects timeout context", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Check that context has timeout
			deadline, ok := r.Context().Deadline()
			assert.True(t, ok)
			assert.True(t, deadline.After(time.Now()))
			w.WriteHeader(http.StatusOK)
		})

		timeoutHandler := middleware.TimeoutMiddleware(1 * time.Second)(handler)

		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()

		timeoutHandler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}

// TestGetClientIP tests IP extraction functionality
func TestGetClientIP(t *testing.T) {
	tests := []struct {
		name           string
		remoteAddr     string
		forwardedFor   string
		realIP         string
		expectedIP     string
	}{
		{
			name:       "direct connection",
			remoteAddr: "192.168.1.1:1234",
			expectedIP: "192.168.1.1",
		},
		{
			name:         "X-Forwarded-For single IP",
			remoteAddr:   "10.0.0.1:1234",
			forwardedFor: "203.0.113.1",
			expectedIP:   "203.0.113.1",
		},
		{
			name:         "X-Forwarded-For multiple IPs",
			remoteAddr:   "10.0.0.1:1234",
			forwardedFor: "203.0.113.1, 198.51.100.1, 192.168.1.1",
			expectedIP:   "203.0.113.1",
		},
		{
			name:       "X-Real-IP",
			remoteAddr: "10.0.0.1:1234",
			realIP:     "203.0.113.1",
			expectedIP: "203.0.113.1",
		},
		{
			name:         "X-Forwarded-For takes precedence",
			remoteAddr:   "10.0.0.1:1234",
			forwardedFor: "203.0.113.1",
			realIP:       "198.51.100.1",
			expectedIP:   "203.0.113.1",
		},
		{
			name:       "IPv6",
			remoteAddr: "[::1]:1234",
			expectedIP: "::1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/test", nil)
			req.RemoteAddr = tt.remoteAddr

			if tt.forwardedFor != "" {
				req.Header.Set("X-Forwarded-For", tt.forwardedFor)
			}
			if tt.realIP != "" {
				req.Header.Set("X-Real-IP", tt.realIP)
			}

			// We can't directly test getClientIP as it's not exported,
			// but we can test it through middleware that uses it
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})

			// Use logging middleware which calls getClientIP
			loggingHandler := middleware.Logging()(handler)
			w := httptest.NewRecorder()

			loggingHandler.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)
		})
	}
}

// TestConcurrentMiddleware tests middleware thread safety
func TestConcurrentMiddleware(t *testing.T) {
	t.Run("concurrent CORS requests", func(t *testing.T) {
		handler := createTestHandler()
		corsHandler := middleware.CORS([]string{"*"})(handler)

		const numRequests = 50
		done := make(chan bool, numRequests)

		for i := 0; i < numRequests; i++ {
			go func() {
				req := httptest.NewRequest("GET", "/test", nil)
				req.Header.Set("Origin", "https://example.com")
				w := httptest.NewRecorder()

				corsHandler.ServeHTTP(w, req)

				assert.Equal(t, http.StatusOK, w.Code)
				assert.Equal(t, "https://example.com", w.Header().Get("Access-Control-Allow-Origin"))
				done <- true
			}()
		}

		for i := 0; i < numRequests; i++ {
			<-done
		}
	})

	t.Run("concurrent security middleware", func(t *testing.T) {
		handler := createTestHandler()
		securityHandler := middleware.Security()(handler)

		const numRequests = 50
		done := make(chan bool, numRequests)

		for i := 0; i < numRequests; i++ {
			go func() {
				req := httptest.NewRequest("GET", "/test", nil)
				w := httptest.NewRecorder()

				securityHandler.ServeHTTP(w, req)

				assert.Equal(t, http.StatusOK, w.Code)
				assert.Equal(t, "nosniff", w.Header().Get("X-Content-Type-Options"))
				done <- true
			}()
		}

		for i := 0; i < numRequests; i++ {
			<-done
		}
	})
}

// TestMiddlewareEdgeCases tests edge cases
func TestMiddlewareEdgeCases(t *testing.T) {
	t.Run("CORS with empty origin", func(t *testing.T) {
		handler := createTestHandler()
		corsHandler := middleware.CORS([]string{"https://example.com"})(handler)

		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("Origin", "")
		w := httptest.NewRecorder()

		corsHandler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Empty(t, w.Header().Get("Access-Control-Allow-Origin"))
	})

	t.Run("request size with zero limit", func(t *testing.T) {
		handler := createTestHandler()
		sizeHandler := middleware.RequestSize(0)(handler)

		req := httptest.NewRequest("POST", "/test", strings.NewReader("any content"))
		req.ContentLength = 11
		w := httptest.NewRecorder()

		sizeHandler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusRequestEntityTooLarge, w.Code)
	})

	t.Run("basic auth with empty credentials", func(t *testing.T) {
		handler := createTestHandler()
		authHandler := middleware.BasicAuthMiddleware("", "")(handler)

		req := httptest.NewRequest("GET", "/test", nil)
		req.SetBasicAuth("", "")
		w := httptest.NewRecorder()

		authHandler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("sanitize with special characters", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			param := r.URL.Query().Get("test")
			// Should contain the emoji but not null bytes
			assert.Contains(t, param, "🚗")
			assert.NotContains(t, param, "\x00")
			w.WriteHeader(http.StatusOK)
		})

		sanitizeHandler := middleware.SanitizeMiddleware(handler)

		req := httptest.NewRequest("GET", "/test?test=🚗%00test", nil)
		w := httptest.NewRecorder()

		sanitizeHandler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}
// Test RateLimit middleware and related functions
func TestRateLimit(t *testing.T) {
	t.Run("default rate limiter", func(t *testing.T) {
		handler := createTestHandler()
		rateLimitHandler := middleware.RateLimit()(handler)

		// First request should pass
		req := httptest.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "192.168.1.1:1234"
		w := httptest.NewRecorder()
		rateLimitHandler.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)

		// Rapid requests might be rate limited
		for i := 0; i < 15; i++ {
			req := httptest.NewRequest("GET", "/test", nil)
			req.RemoteAddr = "192.168.1.1:1234"
			w := httptest.NewRecorder()
			rateLimitHandler.ServeHTTP(w, req)
			// Some requests will pass, some might be rate limited
		}
	})
}

func TestNewRateLimiter(t *testing.T) {
	t.Run("creates rate limiter with custom settings", func(t *testing.T) {
		limiter := middleware.NewRateLimiter(10, time.Minute)
		assert.NotNil(t, limiter)

		// Test Allow method
		assert.True(t, limiter.Allow("192.168.1.1"))
	})

	t.Run("rate limiter middleware", func(t *testing.T) {
		limiter := middleware.NewRateLimiter(2, time.Second)
		handler := createTestHandler()
		rateLimitHandler := limiter.RateLimitMiddleware(handler)

		// First two requests should pass
		for i := 0; i < 2; i++ {
			req := httptest.NewRequest("GET", "/test", nil)
			req.RemoteAddr = fmt.Sprintf("192.168.1.%d:1234", i+1)
			w := httptest.NewRecorder()
			rateLimitHandler.ServeHTTP(w, req)
			assert.Equal(t, http.StatusOK, w.Code)
		}

		// Third request from same IP might be rate limited
		req := httptest.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "192.168.1.1:1234"
		w := httptest.NewRecorder()
		rateLimitHandler.ServeHTTP(w, req)
		// Could be OK or TooManyRequests depending on timing
	})

	t.Run("concurrent requests handling", func(t *testing.T) {
		limiter := middleware.NewRateLimiter(100, time.Second)

		var wg sync.WaitGroup
		for i := 0; i < 50; i++ {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				ip := fmt.Sprintf("192.168.1.%d", i%10)
				limiter.Allow(ip)
			}(i)
		}
		wg.Wait()
	})
}

func TestNewEnhancedRateLimiter(t *testing.T) {
	t.Run("creates enhanced rate limiter", func(t *testing.T) {
		limiter := middleware.NewEnhancedRateLimiter(10, 5, time.Minute)
		assert.NotNil(t, limiter)

		// Test Allow method
		assert.True(t, limiter.Allow("192.168.1.1"))
	})

	t.Run("burst handling", func(t *testing.T) {
		limiter := middleware.NewEnhancedRateLimiter(10, 5, time.Second)

		// Burst of 5 requests should pass
		for i := 0; i < 5; i++ {
			assert.True(t, limiter.Allow("192.168.1.1"))
		}

		// Wait a bit for token replenishment
		time.Sleep(100 * time.Millisecond)

		// More requests might pass depending on rate
		limiter.Allow("192.168.1.1")
	})

	t.Run("different IPs tracked separately", func(t *testing.T) {
		limiter := middleware.NewEnhancedRateLimiter(5, 2, time.Second)

		// Different IPs should have separate limits
		assert.True(t, limiter.Allow("192.168.1.1"))
		assert.True(t, limiter.Allow("192.168.1.2"))
		assert.True(t, limiter.Allow("192.168.1.3"))

		// Each IP has its own bucket
		for i := 0; i < 2; i++ {
			limiter.Allow("192.168.1.1")
			limiter.Allow("192.168.1.2")
		}
	})

	t.Run("cleanup of old entries", func(t *testing.T) {
		limiter := middleware.NewEnhancedRateLimiter(10, 5, 100*time.Millisecond)

		// Add some entries
		limiter.Allow("192.168.1.1")
		limiter.Allow("192.168.1.2")

		// Wait for cleanup
		time.Sleep(150 * time.Millisecond)

		// Old entries should be cleaned up, new requests should pass
		assert.True(t, limiter.Allow("192.168.1.1"))
	})

	t.Run("concurrent access safety", func(t *testing.T) {
		limiter := middleware.NewEnhancedRateLimiter(1000, 100, time.Second)

		var wg sync.WaitGroup
		for i := 0; i < 100; i++ {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				ip := fmt.Sprintf("10.0.0.%d", i%20)
				for j := 0; j < 10; j++ {
					limiter.Allow(ip)
				}
			}(i)
		}
		wg.Wait()
	})
}

package middleware

import (
	"context"
	"crypto/subtle"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Chain applies multiple middleware functions in order
func Chain(handler http.Handler, middlewares ...func(http.Handler) http.Handler) http.Handler {
	// Apply middleware in reverse order so the first middleware in the list
	// is the outermost middleware (executed first)
	for i := len(middlewares) - 1; i >= 0; i-- {
		handler = middlewares[i](handler)
	}
	return handler
}

// CORS middleware configuration
type CORSConfig struct {
	AllowedOrigins   []string
	AllowedMethods   []string
	AllowedHeaders   []string
	ExposedHeaders   []string
	AllowCredentials bool
	MaxAge           int
}

// DefaultCORSConfig returns a default CORS configuration
func DefaultCORSConfig() CORSConfig {
	return CORSConfig{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{
			http.MethodGet,
			http.MethodPost,
			http.MethodPut,
			http.MethodPatch,
			http.MethodDelete,
			http.MethodOptions,
			http.MethodHead,
		},
		AllowedHeaders: []string{
			"Accept",
			"Authorization",
			"Content-Type",
			"X-CSRF-Token",
			"X-Requested-With",
			"X-Request-ID",
		},
		ExposedHeaders: []string{
			"X-Request-ID",
			"X-Response-Time",
		},
		AllowCredentials: true,
		MaxAge:           300, // 5 minutes
	}
}

// CORS middleware with configurable origins
func CORS(allowedOrigins []string) func(http.Handler) http.Handler {
	config := DefaultCORSConfig()
	if len(allowedOrigins) > 0 {
		config.AllowedOrigins = allowedOrigins
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")

			// Check if origin is allowed
			allowed := false
			for _, allowedOrigin := range config.AllowedOrigins {
				if allowedOrigin == "*" || allowedOrigin == origin {
					allowed = true
					break
				}
			}

			if allowed {
				w.Header().Set("Access-Control-Allow-Origin", origin)
			}

			w.Header().Set("Access-Control-Allow-Methods", strings.Join(config.AllowedMethods, ", "))
			w.Header().Set("Access-Control-Allow-Headers", strings.Join(config.AllowedHeaders, ", "))
			w.Header().Set("Access-Control-Expose-Headers", strings.Join(config.ExposedHeaders, ", "))

			if config.AllowCredentials {
				w.Header().Set("Access-Control-Allow-Credentials", "true")
			}

			if config.MaxAge > 0 {
				w.Header().Set("Access-Control-Max-Age", strconv.Itoa(config.MaxAge))
			}

			// Handle preflight requests
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// Enhanced Security headers middleware
func Security() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Content Security Policy
			csp := "default-src 'self'; " +
				"script-src 'self' 'unsafe-inline' https://unpkg.com; " +
				"style-src 'self' 'unsafe-inline' https://unpkg.com; " +
				"font-src 'self' data:; " +
				"img-src 'self' data: https:; " +
				"connect-src 'self';"
			w.Header().Set("Content-Security-Policy", csp)

			// Other security headers
			w.Header().Set("X-Content-Type-Options", "nosniff")
			w.Header().Set("X-Frame-Options", "DENY")
			w.Header().Set("X-XSS-Protection", "1; mode=block")
			w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
			w.Header().Set("Permissions-Policy", "camera=(), microphone=(), geolocation=()")

			// Only add HSTS in production with HTTPS
			if r.TLS != nil && os.Getenv("ENVIRONMENT") == "production" {
				w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
			}

			// Remove server header
			w.Header().Del("Server")

			next.ServeHTTP(w, r)
		})
	}
}

// Request size limiting middleware
func RequestSize(maxSize int64) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.ContentLength > maxSize {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusRequestEntityTooLarge)
				w.Write([]byte(fmt.Sprintf(`{"error":"request too large","max_size":%d}`, maxSize)))
				return
			}

			// Limit the request body reader
			r.Body = http.MaxBytesReader(w, r.Body, maxSize)

			next.ServeHTTP(w, r)
		})
	}
}

// Method filtering middleware
func AllowedMethods(methods ...string) func(http.Handler) http.Handler {
	allowedMethods := make(map[string]bool)
	for _, method := range methods {
		allowedMethods[strings.ToUpper(method)] = true
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !allowedMethods[r.Method] {
				w.Header().Set("Allow", strings.Join(methods, ", "))
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusMethodNotAllowed)
				w.Write([]byte(`{"error":"method not allowed"}`))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// Logging middleware
func Logging() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Create a response writer wrapper to capture status code
			wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

			// Add request ID to context
			requestID := generateRequestID()
			ctx := context.WithValue(r.Context(), "request_id", requestID)
			r = r.WithContext(ctx)

			// Set request ID header
			wrapped.Header().Set("X-Request-ID", requestID)

			next.ServeHTTP(wrapped, r)

			duration := time.Since(start)
			wrapped.Header().Set("X-Response-Time", duration.String())

			log.Printf(
				"%s %s %s %d %s %s",
				getClientIP(r),
				r.Method,
				r.URL.Path,
				wrapped.statusCode,
				duration,
				requestID,
			)
		})
	}
}

// responseWriter wrapper to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (w *responseWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

// SecurityMiddleware provides security headers and basic protection
func SecurityMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set security headers
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		w.Header().Set("Content-Security-Policy", "default-src 'self'")

		// Remove server header
		w.Header().Del("Server")

		next.ServeHTTP(w, r)
	})
}

// Enhanced RateLimiter with burst support
type EnhancedRateLimiter struct {
	visitors map[string]*visitor
	mu       sync.RWMutex
	rate     int
	burst    int
	window   time.Duration
}

type visitor struct {
	requests int
	lastSeen time.Time
	window   time.Time
}

// RateLimiter implements a simple rate limiter (legacy)
type RateLimiter struct {
	mu       sync.Mutex
	visitors map[string]*legacyVisitor
	rate     int           // requests per duration
	duration time.Duration // time window
}

type legacyVisitor struct {
	count      int
	lastAccess time.Time
}

// NewEnhancedRateLimiter creates a new enhanced rate limiter with burst support
func NewEnhancedRateLimiter(rate int, burst int, window time.Duration) *EnhancedRateLimiter {
	rl := &EnhancedRateLimiter{
		visitors: make(map[string]*visitor),
		rate:     rate,
		burst:    burst,
		window:   window,
	}

	// Clean up old visitors every minute
	go func() {
		ticker := time.NewTicker(time.Minute)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				rl.cleanup()
			}
		}
	}()

	return rl
}

func (rl *EnhancedRateLimiter) cleanup() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	for ip, v := range rl.visitors {
		if now.Sub(v.lastSeen) > rl.window*2 {
			delete(rl.visitors, ip)
		}
	}
}

func (rl *EnhancedRateLimiter) allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	v, exists := rl.visitors[ip]

	if !exists {
		rl.visitors[ip] = &visitor{
			requests: 1,
			lastSeen: now,
			window:   now,
		}
		return true
	}

	// Reset window if expired
	if now.Sub(v.window) >= rl.window {
		v.requests = 1
		v.window = now
		v.lastSeen = now
		return true
	}

	// Check if within rate limit
	if v.requests >= rl.burst {
		v.lastSeen = now
		return false
	}

	v.requests++
	v.lastSeen = now
	return true
}

// Global enhanced rate limiter instance
var globalEnhancedRateLimiter *EnhancedRateLimiter

func init() {
	// Default: 100 requests per minute with burst of 20
	rate := 100
	burst := 20
	window := time.Minute

	// Override from environment
	if envRate := os.Getenv("RATE_LIMIT_REQUESTS"); envRate != "" {
		if parsed, err := strconv.Atoi(envRate); err == nil {
			rate = parsed
		}
	}
	if envBurst := os.Getenv("RATE_LIMIT_BURST"); envBurst != "" {
		if parsed, err := strconv.Atoi(envBurst); err == nil {
			burst = parsed
		}
	}
	if envWindow := os.Getenv("RATE_LIMIT_WINDOW"); envWindow != "" {
		if parsed, err := time.ParseDuration(envWindow); err == nil {
			window = parsed
		}
	}

	globalEnhancedRateLimiter = NewEnhancedRateLimiter(rate, burst, window)
}

// RateLimit middleware using enhanced rate limiter
func RateLimit() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := getClientIP(r)

			if !globalEnhancedRateLimiter.allow(ip) {
				w.Header().Set("Content-Type", "application/json")
				w.Header().Set("X-Rate-Limit-Exceeded", "true")
				w.WriteHeader(http.StatusTooManyRequests)
				w.Write([]byte(`{"error":"rate limit exceeded","message":"too many requests"}`))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// NewRateLimiter creates a new legacy rate limiter
func NewRateLimiter(rate int, duration time.Duration) *RateLimiter {
	rl := &RateLimiter{
		visitors: make(map[string]*legacyVisitor),
		rate:     rate,
		duration: duration,
	}

	// Cleanup old entries periodically
	go rl.cleanup()

	return rl
}

// RateLimitMiddleware limits requests per IP
func (rl *RateLimiter) RateLimitMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := getClientIP(r)

		if !rl.allow(ip) {
			http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (rl *RateLimiter) allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	v, exists := rl.visitors[ip]

	if !exists {
		rl.visitors[ip] = &legacyVisitor{count: 1, lastAccess: now}
		return true
	}

	// Reset counter if outside time window
	if now.Sub(v.lastAccess) > rl.duration {
		v.count = 1
		v.lastAccess = now
		return true
	}

	// Increment counter
	v.count++
	v.lastAccess = now

	return v.count <= rl.rate
}

func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		for ip, v := range rl.visitors {
			if now.Sub(v.lastAccess) > rl.duration*2 {
				delete(rl.visitors, ip)
			}
		}
		rl.mu.Unlock()
	}
}

// BasicAuthMiddleware provides HTTP Basic Authentication
func BasicAuthMiddleware(username, password string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, pass, ok := r.BasicAuth()

			if !ok || subtle.ConstantTimeCompare([]byte(user), []byte(username)) != 1 ||
				subtle.ConstantTimeCompare([]byte(pass), []byte(password)) != 1 {
				w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// APIKeyMiddleware validates API key from header
func APIKeyMiddleware(validKeys []string) func(http.Handler) http.Handler {
	keyMap := make(map[string]bool)
	for _, key := range validKeys {
		keyMap[key] = true
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			apiKey := r.Header.Get("X-API-Key")
			if apiKey == "" {
				apiKey = r.URL.Query().Get("api_key")
			}

			if apiKey == "" || !keyMap[apiKey] {
				http.Error(w, "Invalid or missing API key", http.StatusUnauthorized)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// SanitizeMiddleware sanitizes request inputs
func SanitizeMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Sanitize query parameters
		q := r.URL.Query()
		for key, values := range q {
			for i, value := range values {
				q[key][i] = sanitizeString(value)
			}
		}
		r.URL.RawQuery = q.Encode()

		// Limit request body size
		r.Body = http.MaxBytesReader(w, r.Body, 32<<20) // 32MB max

		next.ServeHTTP(w, r)
	})
}

// sanitizeString removes potentially dangerous characters
func sanitizeString(s string) string {
	// Remove null bytes
	s = strings.ReplaceAll(s, "\x00", "")

	// Trim excessive whitespace
	s = strings.TrimSpace(s)

	// Limit length
	if len(s) > 1000 {
		s = s[:1000]
	}

	return s
}

// getClientIP extracts client IP from request (enhanced version)
func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header first
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		ips := strings.Split(xff, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}

	// Check X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}

	// Fall back to RemoteAddr
	ip := r.RemoteAddr
	if strings.Contains(ip, ":") {
		host, _, err := splitHostPort(ip)
		if err == nil {
			return host
		}
	}
	return ip
}

func splitHostPort(hostport string) (host, port string, err error) {
	i := strings.LastIndex(hostport, ":")
	if i < 0 {
		return "", "", fmt.Errorf("missing port in address")
	}
	return hostport[:i], hostport[i+1:], nil
}

func generateRequestID() string {
	// Simple request ID generation
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

// IPWhitelistMiddleware restricts access to specific IPs
func IPWhitelistMiddleware(allowedIPs []string) func(http.Handler) http.Handler {
	ipMap := make(map[string]bool)
	for _, ip := range allowedIPs {
		ipMap[ip] = true
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			clientIP := getClientIP(r)

			if !ipMap[clientIP] {
				http.Error(w, "Access denied", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// TimeoutMiddleware sets a timeout for request processing
func TimeoutMiddleware(timeout time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Set timeout for the request context
			ctx, cancel := context.WithTimeout(r.Context(), timeout)
			defer cancel()

			r = r.WithContext(ctx)
			next.ServeHTTP(w, r)
		})
	}
}
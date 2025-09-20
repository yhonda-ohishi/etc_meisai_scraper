package middleware

import (
	"context"
	"crypto/subtle"
	"net/http"
	"strings"
	"sync"
	"time"
)

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

// RateLimiter implements a simple rate limiter
type RateLimiter struct {
	mu       sync.Mutex
	visitors map[string]*visitor
	rate     int           // requests per duration
	duration time.Duration // time window
}

type visitor struct {
	count      int
	lastAccess time.Time
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(rate int, duration time.Duration) *RateLimiter {
	rl := &RateLimiter{
		visitors: make(map[string]*visitor),
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
		rl.visitors[ip] = &visitor{count: 1, lastAccess: now}
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

// getClientIP extracts client IP from request
func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header
	xff := r.Header.Get("X-Forwarded-For")
	if xff != "" {
		parts := strings.Split(xff, ",")
		if len(parts) > 0 {
			return strings.TrimSpace(parts[0])
		}
	}

	// Check X-Real-IP header
	xri := r.Header.Get("X-Real-IP")
	if xri != "" {
		return xri
	}

	// Fall back to RemoteAddr
	ip := r.RemoteAddr
	if idx := strings.LastIndex(ip, ":"); idx != -1 {
		ip = ip[:idx]
	}

	return ip
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
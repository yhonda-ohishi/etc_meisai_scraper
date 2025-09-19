package middleware

import (
	"net/http"
	"strconv"
	"sync"
	"time"
)

// Metrics holds application metrics
type Metrics struct {
	mu sync.RWMutex

	// Request metrics
	TotalRequests      int64
	TotalErrors        int64
	ActiveConnections  int64

	// Performance metrics
	TotalResponseTime  time.Duration
	MaxResponseTime    time.Duration
	MinResponseTime    time.Duration

	// Business metrics
	TotalRecordsImported int64
	TotalCSVProcessed    int64
	TotalDuplicatesFound int64

	// Status code counters
	StatusCodes map[int]int64

	// Endpoint metrics
	EndpointMetrics map[string]*EndpointStat
}

// EndpointStat holds metrics for a specific endpoint
type EndpointStat struct {
	Count         int64
	TotalTime     time.Duration
	AverageTime   time.Duration
	MaxTime       time.Duration
	MinTime       time.Duration
	LastAccessed  time.Time
}

// GlobalMetrics is the global metrics instance
var GlobalMetrics = &Metrics{
	StatusCodes:     make(map[int]int64),
	EndpointMetrics: make(map[string]*EndpointStat),
	MinResponseTime: time.Hour, // Initialize to a large value
}

// MetricsMiddleware records metrics for each request
func MetricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Increment active connections
		GlobalMetrics.incrementActiveConnections()
		defer GlobalMetrics.decrementActiveConnections()

		// Wrap response writer to capture status code
		wrapped := &responseWriter{
			ResponseWriter: w,
			statusCode:    http.StatusOK,
		}

		// Process request
		next.ServeHTTP(wrapped, r)

		// Record metrics
		duration := time.Since(start)
		GlobalMetrics.recordRequest(r.Method, r.URL.Path, wrapped.statusCode, duration)
	})
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
	written    bool
}

func (rw *responseWriter) WriteHeader(code int) {
	if !rw.written {
		rw.statusCode = code
		rw.ResponseWriter.WriteHeader(code)
		rw.written = true
	}
}

func (rw *responseWriter) Write(data []byte) (int, error) {
	if !rw.written {
		rw.WriteHeader(http.StatusOK)
	}
	return rw.ResponseWriter.Write(data)
}

// Metrics methods

func (m *Metrics) incrementActiveConnections() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.ActiveConnections++
}

func (m *Metrics) decrementActiveConnections() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.ActiveConnections--
}

func (m *Metrics) recordRequest(method, path string, statusCode int, duration time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Update total requests
	m.TotalRequests++

	// Update error count
	if statusCode >= 400 {
		m.TotalErrors++
	}

	// Update status code counter
	m.StatusCodes[statusCode]++

	// Update response time metrics
	m.TotalResponseTime += duration
	if duration > m.MaxResponseTime {
		m.MaxResponseTime = duration
	}
	if duration < m.MinResponseTime {
		m.MinResponseTime = duration
	}

	// Update endpoint metrics
	endpoint := method + " " + path
	stat, exists := m.EndpointMetrics[endpoint]
	if !exists {
		stat = &EndpointStat{
			MinTime: duration,
		}
		m.EndpointMetrics[endpoint] = stat
	}

	stat.Count++
	stat.TotalTime += duration
	stat.AverageTime = stat.TotalTime / time.Duration(stat.Count)
	stat.LastAccessed = time.Now()

	if duration > stat.MaxTime {
		stat.MaxTime = duration
	}
	if duration < stat.MinTime {
		stat.MinTime = duration
	}
}

// IncrementImportedRecords increments the imported records counter
func (m *Metrics) IncrementImportedRecords(count int64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.TotalRecordsImported += count
}

// IncrementCSVProcessed increments the CSV processed counter
func (m *Metrics) IncrementCSVProcessed() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.TotalCSVProcessed++
}

// IncrementDuplicates increments the duplicates counter
func (m *Metrics) IncrementDuplicates(count int64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.TotalDuplicatesFound += count
}

// GetMetrics returns a copy of current metrics
func (m *Metrics) GetMetrics() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	avgResponseTime := time.Duration(0)
	if m.TotalRequests > 0 {
		avgResponseTime = m.TotalResponseTime / time.Duration(m.TotalRequests)
	}

	errorRate := float64(0)
	if m.TotalRequests > 0 {
		errorRate = float64(m.TotalErrors) / float64(m.TotalRequests) * 100
	}

	return map[string]interface{}{
		"total_requests":       m.TotalRequests,
		"total_errors":        m.TotalErrors,
		"error_rate":          errorRate,
		"active_connections":  m.ActiveConnections,
		"avg_response_time":   avgResponseTime.String(),
		"max_response_time":   m.MaxResponseTime.String(),
		"min_response_time":   m.MinResponseTime.String(),
		"records_imported":    m.TotalRecordsImported,
		"csv_processed":       m.TotalCSVProcessed,
		"duplicates_found":    m.TotalDuplicatesFound,
		"status_codes":        m.StatusCodes,
		"endpoint_metrics":    m.formatEndpointMetrics(),
	}
}

func (m *Metrics) formatEndpointMetrics() map[string]interface{} {
	result := make(map[string]interface{})
	for endpoint, stat := range m.EndpointMetrics {
		result[endpoint] = map[string]interface{}{
			"count":         stat.Count,
			"average_time":  stat.AverageTime.String(),
			"max_time":      stat.MaxTime.String(),
			"min_time":      stat.MinTime.String(),
			"last_accessed": stat.LastAccessed.Format(time.RFC3339),
		}
	}
	return result
}

// MetricsHandler returns an HTTP handler for metrics endpoint
func MetricsHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		metrics := GlobalMetrics.GetMetrics()

		// Add system metrics
		metrics["timestamp"] = time.Now().Format(time.RFC3339)
		metrics["uptime"] = time.Since(startTime).String()

		// Simple JSON encoding
		w.Write([]byte("{\n"))
		i := 0
		for key, value := range metrics {
			if i > 0 {
				w.Write([]byte(",\n"))
			}
			w.Write([]byte(`  "` + key + `": `))

			switch v := value.(type) {
			case string:
				w.Write([]byte(`"` + v + `"`))
			case int64:
				w.Write([]byte(strconv.FormatInt(v, 10)))
			case float64:
				w.Write([]byte(strconv.FormatFloat(v, 'f', 2, 64)))
			default:
				// For complex types, just convert to string
				w.Write([]byte(`"` + toString(v) + `"`))
			}
			i++
		}
		w.Write([]byte("\n}\n"))
	}
}

var startTime = time.Now()

func toString(v interface{}) string {
	switch v.(type) {
	case map[string]interface{}:
		return "complex_object"
	case map[int]int64:
		return "status_codes_map"
	default:
		return "unknown"
	}
}
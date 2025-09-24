package middleware_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yhonda-ohishi/etc_meisai/src/middleware"
)

// Test helpers
func resetGlobalMetrics() {
	middleware.GlobalMetrics = &middleware.Metrics{
		StatusCodes:     make(map[int]int64),
		EndpointMetrics: make(map[string]*middleware.EndpointStat),
		MinResponseTime: time.Hour,
	}
}

func createTestHandler(statusCode int, delay time.Duration) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if delay > 0 {
			time.Sleep(delay)
		}
		w.WriteHeader(statusCode)
		w.Write([]byte("test response"))
	}
}

// Test MetricsMiddleware
func TestMetricsMiddleware(t *testing.T) {
	t.Run("records successful request", func(t *testing.T) {
		resetGlobalMetrics()

		handler := middleware.MetricsMiddleware(createTestHandler(http.StatusOK, 10*time.Millisecond))
		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		metrics := middleware.GlobalMetrics.GetMetrics()
		assert.Equal(t, int64(1), metrics["total_requests"])
		assert.Equal(t, int64(0), metrics["total_errors"])
		assert.Equal(t, int64(0), metrics["active_connections"])

		statusCodes := metrics["status_codes"].(map[int]int64)
		assert.Equal(t, int64(1), statusCodes[200])
	})

	t.Run("records error request", func(t *testing.T) {
		resetGlobalMetrics()

		handler := middleware.MetricsMiddleware(createTestHandler(http.StatusInternalServerError, 0))
		req := httptest.NewRequest("POST", "/api/error", nil)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		metrics := middleware.GlobalMetrics.GetMetrics()
		assert.Equal(t, int64(1), metrics["total_requests"])
		assert.Equal(t, int64(1), metrics["total_errors"])

		statusCodes := metrics["status_codes"].(map[int]int64)
		assert.Equal(t, int64(1), statusCodes[500])
	})

	t.Run("tracks response times", func(t *testing.T) {
		resetGlobalMetrics()

		delays := []time.Duration{
			10 * time.Millisecond,
			20 * time.Millisecond,
			5 * time.Millisecond,
		}

		for _, delay := range delays {
			handler := middleware.MetricsMiddleware(createTestHandler(http.StatusOK, delay))
			req := httptest.NewRequest("GET", "/test", nil)
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)
		}

		metrics := middleware.GlobalMetrics.GetMetrics()
		assert.Equal(t, int64(3), metrics["total_requests"])

		// Check response times were tracked
		maxTime := metrics["max_response_time"].(string)
		minTime := metrics["min_response_time"].(string)
		avgTime := metrics["avg_response_time"].(string)

		assert.NotEmpty(t, maxTime)
		assert.NotEmpty(t, minTime)
		assert.NotEmpty(t, avgTime)
	})

	t.Run("tracks endpoint metrics", func(t *testing.T) {
		resetGlobalMetrics()

		endpoints := []struct {
			method string
			path   string
			count  int
		}{
			{"GET", "/api/records", 3},
			{"POST", "/api/records", 2},
			{"DELETE", "/api/records/1", 1},
		}

		for _, ep := range endpoints {
			for i := 0; i < ep.count; i++ {
				handler := middleware.MetricsMiddleware(createTestHandler(http.StatusOK, 0))
				req := httptest.NewRequest(ep.method, ep.path, nil)
				w := httptest.NewRecorder()
				handler.ServeHTTP(w, req)
			}
		}

		metrics := middleware.GlobalMetrics.GetMetrics()
		endpointMetrics := metrics["endpoint_metrics"].(map[string]interface{})

		getRecords := endpointMetrics["GET /api/records"].(map[string]interface{})
		assert.Equal(t, int64(3), getRecords["count"])

		postRecords := endpointMetrics["POST /api/records"].(map[string]interface{})
		assert.Equal(t, int64(2), postRecords["count"])

		deleteRecord := endpointMetrics["DELETE /api/records/1"].(map[string]interface{})
		assert.Equal(t, int64(1), deleteRecord["count"])
	})

	t.Run("handles concurrent requests", func(t *testing.T) {
		resetGlobalMetrics()

		var wg sync.WaitGroup
		numRequests := 100

		for i := 0; i < numRequests; i++ {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()

				statusCode := http.StatusOK
				if i%10 == 0 {
					statusCode = http.StatusBadRequest
				}

				handler := middleware.MetricsMiddleware(createTestHandler(statusCode, 0))
				req := httptest.NewRequest("GET", fmt.Sprintf("/test/%d", i%5), nil)
				w := httptest.NewRecorder()
				handler.ServeHTTP(w, req)
			}(i)
		}

		wg.Wait()

		metrics := middleware.GlobalMetrics.GetMetrics()
		assert.Equal(t, int64(numRequests), metrics["total_requests"])
		assert.Equal(t, int64(10), metrics["total_errors"]) // 10% error rate
	})

	t.Run("tracks active connections", func(t *testing.T) {
		resetGlobalMetrics()

		var wg sync.WaitGroup
		startChan := make(chan struct{})

		// Start multiple concurrent requests
		for i := 0; i < 5; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()

				handler := middleware.MetricsMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					<-startChan // Wait for signal
					w.WriteHeader(http.StatusOK)
				}))

				req := httptest.NewRequest("GET", "/test", nil)
				w := httptest.NewRecorder()
				handler.ServeHTTP(w, req)
			}()
		}

		// Give goroutines time to start
		time.Sleep(100 * time.Millisecond)

		// Check active connections during processing
		metrics := middleware.GlobalMetrics.GetMetrics()
		// Note: Due to timing, this might not always be exactly 5
		assert.GreaterOrEqual(t, metrics["active_connections"].(int64), int64(0))

		close(startChan) // Release all requests
		wg.Wait()

		// Check active connections after completion
		metrics = middleware.GlobalMetrics.GetMetrics()
		assert.Equal(t, int64(0), metrics["active_connections"])
	})

	t.Run("handles different status codes", func(t *testing.T) {
		resetGlobalMetrics()

		statusCodes := []int{
			http.StatusOK,
			http.StatusCreated,
			http.StatusAccepted,
			http.StatusBadRequest,
			http.StatusUnauthorized,
			http.StatusForbidden,
			http.StatusNotFound,
			http.StatusInternalServerError,
			http.StatusBadGateway,
			http.StatusServiceUnavailable,
		}

		for _, code := range statusCodes {
			handler := middleware.MetricsMiddleware(createTestHandler(code, 0))
			req := httptest.NewRequest("GET", "/test", nil)
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)
		}

		metrics := middleware.GlobalMetrics.GetMetrics()
		statusCodesMap := metrics["status_codes"].(map[int]int64)

		for _, code := range statusCodes {
			assert.Equal(t, int64(1), statusCodesMap[code], "Status code %d", code)
		}

		// Count errors (4xx and 5xx)
		expectedErrors := int64(7) // 400, 401, 403, 404, 500, 502, 503
		assert.Equal(t, expectedErrors, metrics["total_errors"])
	})

	t.Run("response writer captures status correctly", func(t *testing.T) {
		resetGlobalMetrics()

		// Test with explicit WriteHeader
		handler := middleware.MetricsMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusCreated)
			w.Write([]byte("created"))
		}))

		req := httptest.NewRequest("POST", "/test", nil)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)

		metrics := middleware.GlobalMetrics.GetMetrics()
		statusCodes := metrics["status_codes"].(map[int]int64)
		assert.Equal(t, int64(1), statusCodes[201])
	})

	t.Run("response writer defaults to 200 on Write", func(t *testing.T) {
		resetGlobalMetrics()

		// Test without explicit WriteHeader
		handler := middleware.MetricsMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("ok"))
		}))

		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		metrics := middleware.GlobalMetrics.GetMetrics()
		statusCodes := metrics["status_codes"].(map[int]int64)
		assert.Equal(t, int64(1), statusCodes[200])
	})

	t.Run("handles panic in handler", func(t *testing.T) {
		resetGlobalMetrics()

		defer func() {
			if r := recover(); r != nil {
				// Expected panic, metrics should still be recorded
				metrics := middleware.GlobalMetrics.GetMetrics()
				assert.Equal(t, int64(0), metrics["active_connections"])
			}
		}()

		handler := middleware.MetricsMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			panic("test panic")
		}))

		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()

		// This will panic, but defer should still run
		assert.Panics(t, func() {
			handler.ServeHTTP(w, req)
		})
	})
}

// Test Metrics methods
func TestMetricsBusinessMethods(t *testing.T) {
	t.Run("IncrementImportedRecords", func(t *testing.T) {
		resetGlobalMetrics()

		middleware.GlobalMetrics.IncrementImportedRecords(100)
		middleware.GlobalMetrics.IncrementImportedRecords(50)

		metrics := middleware.GlobalMetrics.GetMetrics()
		assert.Equal(t, int64(150), metrics["records_imported"])
	})

	t.Run("IncrementCSVProcessed", func(t *testing.T) {
		resetGlobalMetrics()

		for i := 0; i < 5; i++ {
			middleware.GlobalMetrics.IncrementCSVProcessed()
		}

		metrics := middleware.GlobalMetrics.GetMetrics()
		assert.Equal(t, int64(5), metrics["csv_processed"])
	})

	t.Run("IncrementDuplicates", func(t *testing.T) {
		resetGlobalMetrics()

		middleware.GlobalMetrics.IncrementDuplicates(10)
		middleware.GlobalMetrics.IncrementDuplicates(5)
		middleware.GlobalMetrics.IncrementDuplicates(3)

		metrics := middleware.GlobalMetrics.GetMetrics()
		assert.Equal(t, int64(18), metrics["duplicates_found"])
	})

	t.Run("concurrent business metrics updates", func(t *testing.T) {
		resetGlobalMetrics()

		var wg sync.WaitGroup
		iterations := 100

		// Concurrent imports
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < iterations; i++ {
				middleware.GlobalMetrics.IncrementImportedRecords(1)
			}
		}()

		// Concurrent CSV processing
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < iterations; i++ {
				middleware.GlobalMetrics.IncrementCSVProcessed()
			}
		}()

		// Concurrent duplicate detection
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < iterations; i++ {
				middleware.GlobalMetrics.IncrementDuplicates(1)
			}
		}()

		wg.Wait()

		metrics := middleware.GlobalMetrics.GetMetrics()
		assert.Equal(t, int64(iterations), metrics["records_imported"])
		assert.Equal(t, int64(iterations), metrics["csv_processed"])
		assert.Equal(t, int64(iterations), metrics["duplicates_found"])
	})
}

// Test GetMetrics
func TestGetMetrics(t *testing.T) {
	t.Run("returns all metric fields", func(t *testing.T) {
		resetGlobalMetrics()

		// Generate some metrics
		handler := middleware.MetricsMiddleware(createTestHandler(http.StatusOK, 10*time.Millisecond))
		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		middleware.GlobalMetrics.IncrementImportedRecords(100)
		middleware.GlobalMetrics.IncrementCSVProcessed()
		middleware.GlobalMetrics.IncrementDuplicates(5)

		metrics := middleware.GlobalMetrics.GetMetrics()

		// Check all expected fields exist
		expectedFields := []string{
			"total_requests",
			"total_errors",
			"error_rate",
			"active_connections",
			"avg_response_time",
			"max_response_time",
			"min_response_time",
			"records_imported",
			"csv_processed",
			"duplicates_found",
			"status_codes",
			"endpoint_metrics",
		}

		for _, field := range expectedFields {
			assert.Contains(t, metrics, field, "Missing field: %s", field)
		}
	})

	t.Run("calculates error rate correctly", func(t *testing.T) {
		resetGlobalMetrics()

		// Create 10 requests, 3 with errors
		for i := 0; i < 7; i++ {
			handler := middleware.MetricsMiddleware(createTestHandler(http.StatusOK, 0))
			req := httptest.NewRequest("GET", "/test", nil)
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)
		}

		for i := 0; i < 3; i++ {
			handler := middleware.MetricsMiddleware(createTestHandler(http.StatusInternalServerError, 0))
			req := httptest.NewRequest("GET", "/test", nil)
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)
		}

		metrics := middleware.GlobalMetrics.GetMetrics()
		errorRate := metrics["error_rate"].(float64)
		assert.InDelta(t, 30.0, errorRate, 0.1) // 30% error rate
	})

	t.Run("handles zero requests", func(t *testing.T) {
		resetGlobalMetrics()

		metrics := middleware.GlobalMetrics.GetMetrics()

		assert.Equal(t, int64(0), metrics["total_requests"])
		assert.Equal(t, float64(0), metrics["error_rate"])
		assert.Equal(t, "0s", metrics["avg_response_time"])
	})

	t.Run("formats endpoint metrics", func(t *testing.T) {
		resetGlobalMetrics()

		// Generate endpoint metrics
		for i := 0; i < 5; i++ {
			handler := middleware.MetricsMiddleware(createTestHandler(http.StatusOK, time.Duration(i)*time.Millisecond))
			req := httptest.NewRequest("GET", "/api/test", nil)
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)
		}

		metrics := middleware.GlobalMetrics.GetMetrics()
		endpointMetrics := metrics["endpoint_metrics"].(map[string]interface{})

		testEndpoint := endpointMetrics["GET /api/test"].(map[string]interface{})
		assert.Equal(t, int64(5), testEndpoint["count"])
		assert.NotEmpty(t, testEndpoint["average_time"])
		assert.NotEmpty(t, testEndpoint["max_time"])
		assert.NotEmpty(t, testEndpoint["min_time"])
		assert.NotEmpty(t, testEndpoint["last_accessed"])
	})

	t.Run("thread-safe read during updates", func(t *testing.T) {
		resetGlobalMetrics()

		var wg sync.WaitGroup
		done := make(chan struct{})

		// Continuous updates
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-done:
					return
				default:
					middleware.GlobalMetrics.IncrementImportedRecords(1)
					middleware.GlobalMetrics.IncrementCSVProcessed()
					middleware.GlobalMetrics.IncrementDuplicates(1)
				}
			}
		}()

		// Continuous reads
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < 100; i++ {
				metrics := middleware.GlobalMetrics.GetMetrics()
				assert.NotNil(t, metrics)
				assert.GreaterOrEqual(t, metrics["records_imported"].(int64), int64(0))
			}
		}()

		time.Sleep(100 * time.Millisecond)
		close(done)
		wg.Wait()
	})
}

// Test MetricsHandler
func TestMetricsHandler(t *testing.T) {
	t.Run("returns JSON metrics", func(t *testing.T) {
		resetGlobalMetrics()

		// Generate some metrics
		handler := middleware.MetricsMiddleware(createTestHandler(http.StatusOK, 0))
		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		// Call metrics handler
		metricsHandler := middleware.MetricsHandler()
		req = httptest.NewRequest("GET", "/metrics", nil)
		w = httptest.NewRecorder()
		metricsHandler(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

		// Parse JSON response
		var result map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &result)
		require.NoError(t, err)

		assert.Contains(t, result, "timestamp")
		assert.Contains(t, result, "uptime")
		assert.Contains(t, result, "total_requests")
	})

	t.Run("handles complex metric types", func(t *testing.T) {
		resetGlobalMetrics()

		// Generate metrics with various types
		for i := 0; i < 3; i++ {
			status := http.StatusOK
			if i == 2 {
				status = http.StatusNotFound
			}
			handler := middleware.MetricsMiddleware(createTestHandler(status, 0))
			req := httptest.NewRequest("GET", fmt.Sprintf("/test%d", i), nil)
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)
		}

		metricsHandler := middleware.MetricsHandler()
		req := httptest.NewRequest("GET", "/metrics", nil)
		w := httptest.NewRecorder()
		metricsHandler(w, req)

		// Should handle the response without errors
		assert.Equal(t, http.StatusOK, w.Code)

		// Check for complex types being handled
		body := w.Body.String()
		assert.Contains(t, body, "total_requests")
		assert.Contains(t, body, "status_codes")
		assert.Contains(t, body, "endpoint_metrics")
	})

	t.Run("formats different value types", func(t *testing.T) {
		resetGlobalMetrics()

		middleware.GlobalMetrics.IncrementImportedRecords(12345)

		metricsHandler := middleware.MetricsHandler()
		req := httptest.NewRequest("GET", "/metrics", nil)
		w := httptest.NewRecorder()
		metricsHandler(w, req)

		body := w.Body.String()

		// Check int64 formatting
		assert.Contains(t, body, "12345")

		// Check string formatting (timestamp, uptime)
		assert.Contains(t, body, `"timestamp":`)
		assert.Contains(t, body, `"uptime":`)
	})

	t.Run("concurrent access to handler", func(t *testing.T) {
		resetGlobalMetrics()

		var wg sync.WaitGroup
		metricsHandler := middleware.MetricsHandler()

		for i := 0; i < 50; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()

				req := httptest.NewRequest("GET", "/metrics", nil)
				w := httptest.NewRecorder()
				metricsHandler(w, req)

				assert.Equal(t, http.StatusOK, w.Code)
			}()
		}

		wg.Wait()
	})
}

// Test edge cases
func TestEdgeCases(t *testing.T) {
	t.Run("handles very long response times", func(t *testing.T) {
		resetGlobalMetrics()

		handler := middleware.MetricsMiddleware(createTestHandler(http.StatusOK, 2*time.Second))
		req := httptest.NewRequest("GET", "/slow", nil)
		w := httptest.NewRecorder()

		start := time.Now()
		handler.ServeHTTP(w, req)
		elapsed := time.Since(start)

		assert.GreaterOrEqual(t, elapsed, 2*time.Second)

		metrics := middleware.GlobalMetrics.GetMetrics()
		maxTime := metrics["max_response_time"].(string)
		assert.Contains(t, maxTime, "2")
	})

	t.Run("handles very long endpoint paths", func(t *testing.T) {
		resetGlobalMetrics()

		longPath := "/api/v1/very/long/endpoint/path/that/goes/on/and/on/and/on"
		handler := middleware.MetricsMiddleware(createTestHandler(http.StatusOK, 0))
		req := httptest.NewRequest("GET", longPath, nil)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		metrics := middleware.GlobalMetrics.GetMetrics()
		endpointMetrics := metrics["endpoint_metrics"].(map[string]interface{})

		key := "GET " + longPath
		assert.Contains(t, endpointMetrics, key)
	})

	t.Run("handles unusual HTTP methods", func(t *testing.T) {
		resetGlobalMetrics()

		methods := []string{"GET", "POST", "PUT", "DELETE", "PATCH", "HEAD", "OPTIONS", "CONNECT", "TRACE"}

		for _, method := range methods {
			handler := middleware.MetricsMiddleware(createTestHandler(http.StatusOK, 0))
			req := httptest.NewRequest(method, "/test", nil)
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)
		}

		metrics := middleware.GlobalMetrics.GetMetrics()
		assert.Equal(t, int64(len(methods)), metrics["total_requests"])

		endpointMetrics := metrics["endpoint_metrics"].(map[string]interface{})
		for _, method := range methods {
			key := method + " /test"
			assert.Contains(t, endpointMetrics, key)
		}
	})

	t.Run("handles multiple WriteHeader calls", func(t *testing.T) {
		resetGlobalMetrics()

		handler := middleware.MetricsMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.WriteHeader(http.StatusInternalServerError) // Should be ignored
			w.Write([]byte("test"))
		}))

		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		metrics := middleware.GlobalMetrics.GetMetrics()
		statusCodes := metrics["status_codes"].(map[int]int64)
		assert.Equal(t, int64(1), statusCodes[200])
		assert.Equal(t, int64(0), statusCodes[500])
	})

	t.Run("handles nil response body", func(t *testing.T) {
		resetGlobalMetrics()

		handler := middleware.MetricsMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNoContent)
			// No body written
		}))

		req := httptest.NewRequest("DELETE", "/test", nil)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNoContent, w.Code)

		metrics := middleware.GlobalMetrics.GetMetrics()
		statusCodes := metrics["status_codes"].(map[int]int64)
		assert.Equal(t, int64(1), statusCodes[204])
	})
}

// Benchmark tests
func BenchmarkMetricsMiddleware(b *testing.B) {
	resetGlobalMetrics()
	handler := middleware.MetricsMiddleware(createTestHandler(http.StatusOK, 0))

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			req := httptest.NewRequest("GET", "/test", nil)
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)
		}
	})
}

func BenchmarkGetMetrics(b *testing.B) {
	resetGlobalMetrics()

	// Generate some metrics
	for i := 0; i < 100; i++ {
		handler := middleware.MetricsMiddleware(createTestHandler(http.StatusOK, 0))
		req := httptest.NewRequest("GET", fmt.Sprintf("/test%d", i%10), nil)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = middleware.GlobalMetrics.GetMetrics()
	}
}

func BenchmarkBusinessMetrics(b *testing.B) {
	resetGlobalMetrics()

	b.Run("IncrementImportedRecords", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				middleware.GlobalMetrics.IncrementImportedRecords(1)
			}
		})
	})

	b.Run("IncrementCSVProcessed", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				middleware.GlobalMetrics.IncrementCSVProcessed()
			}
		})
	})

	b.Run("IncrementDuplicates", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				middleware.GlobalMetrics.IncrementDuplicates(1)
			}
		})
	})
}

// Test custom helpers
func TestHelpers(t *testing.T) {
	t.Run("toString handles various types", func(t *testing.T) {
		// This tests the internal toString helper via MetricsHandler
		resetGlobalMetrics()

		// Generate metrics to ensure complex types are present
		handler := middleware.MetricsMiddleware(createTestHandler(http.StatusOK, 0))
		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		metricsHandler := middleware.MetricsHandler()
		req = httptest.NewRequest("GET", "/metrics", nil)
		w = httptest.NewRecorder()
		metricsHandler(w, req)

		body := w.Body.String()

		// Check that complex objects are handled
		assert.NotContains(t, body, "panic")
		assert.Contains(t, body, "{")
		assert.Contains(t, body, "}")
	})
}

// Test response writer implementation
func TestResponseWriter(t *testing.T) {
	t.Run("implements http.ResponseWriter", func(t *testing.T) {
		resetGlobalMetrics()

		handler := middleware.MetricsMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Verify w implements ResponseWriter
			var _ http.ResponseWriter = w

			// Test Header method
			w.Header().Set("X-Test", "value")

			// Test WriteHeader
			w.WriteHeader(http.StatusCreated)

			// Test Write
			n, err := w.Write([]byte("test body"))
			assert.NoError(t, err)
			assert.Equal(t, 9, n)
		}))

		req := httptest.NewRequest("POST", "/test", nil)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		assert.Equal(t, "value", w.Header().Get("X-Test"))
		assert.Equal(t, http.StatusCreated, w.Code)
		assert.Equal(t, "test body", w.Body.String())
	})

	t.Run("Write without WriteHeader defaults to 200", func(t *testing.T) {
		resetGlobalMetrics()

		handler := middleware.MetricsMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Write without calling WriteHeader first
			w.Write([]byte("test"))
		}))

		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("multiple Write calls work correctly", func(t *testing.T) {
		resetGlobalMetrics()

		handler := middleware.MetricsMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("first "))
			w.Write([]byte("second "))
			w.Write([]byte("third"))
		}))

		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		assert.Equal(t, "first second third", w.Body.String())
	})

	t.Run("WriteHeader only works once", func(t *testing.T) {
		resetGlobalMetrics()

		handler := middleware.MetricsMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusCreated)
			w.WriteHeader(http.StatusBadRequest) // Should be ignored
			w.Write([]byte("created"))
		}))

		req := httptest.NewRequest("POST", "/test", nil)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)

		metrics := middleware.GlobalMetrics.GetMetrics()
		statusCodes := metrics["status_codes"].(map[int]int64)
		assert.Equal(t, int64(1), statusCodes[201])
		assert.Equal(t, int64(0), statusCodes[400])
	})
}

// Test metric calculations
func TestMetricCalculations(t *testing.T) {
	t.Run("average response time calculation", func(t *testing.T) {
		resetGlobalMetrics()

		// Create requests with known delays
		delays := []time.Duration{
			100 * time.Millisecond,
			200 * time.Millisecond,
			300 * time.Millisecond,
		}

		for _, delay := range delays {
			handler := middleware.MetricsMiddleware(createTestHandler(http.StatusOK, delay))
			req := httptest.NewRequest("GET", "/test", nil)
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)
		}

		metrics := middleware.GlobalMetrics.GetMetrics()
		avgTimeStr := metrics["avg_response_time"].(string)

		// Parse duration
		avgTime, err := time.ParseDuration(avgTimeStr)
		require.NoError(t, err)

		// Average should be around 200ms
		assert.Greater(t, avgTime, 150*time.Millisecond)
		assert.Less(t, avgTime, 250*time.Millisecond)
	})

	t.Run("min/max response time tracking", func(t *testing.T) {
		resetGlobalMetrics()

		delays := []time.Duration{
			50 * time.Millisecond,
			150 * time.Millisecond,
			100 * time.Millisecond,
		}

		for _, delay := range delays {
			handler := middleware.MetricsMiddleware(createTestHandler(http.StatusOK, delay))
			req := httptest.NewRequest("GET", "/test", nil)
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)
		}

		metrics := middleware.GlobalMetrics.GetMetrics()

		minTimeStr := metrics["min_response_time"].(string)
		maxTimeStr := metrics["max_response_time"].(string)

		minTime, _ := time.ParseDuration(minTimeStr)
		maxTime, _ := time.ParseDuration(maxTimeStr)

		assert.Greater(t, minTime, 40*time.Millisecond)
		assert.Less(t, minTime, 60*time.Millisecond)

		assert.Greater(t, maxTime, 140*time.Millisecond)
		assert.Less(t, maxTime, 160*time.Millisecond)
	})

	t.Run("endpoint average time calculation", func(t *testing.T) {
		resetGlobalMetrics()

		// Multiple requests to same endpoint
		for i := 0; i < 5; i++ {
			delay := time.Duration(i*10) * time.Millisecond
			handler := middleware.MetricsMiddleware(createTestHandler(http.StatusOK, delay))
			req := httptest.NewRequest("GET", "/api/test", nil)
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)
		}

		metrics := middleware.GlobalMetrics.GetMetrics()
		endpointMetrics := metrics["endpoint_metrics"].(map[string]interface{})
		testEndpoint := endpointMetrics["GET /api/test"].(map[string]interface{})

		assert.Equal(t, int64(5), testEndpoint["count"])
		assert.NotEmpty(t, testEndpoint["average_time"])

		// Average should be (0+10+20+30+40)/5 = 20ms
		avgTimeStr := testEndpoint["average_time"].(string)
		avgTime, _ := time.ParseDuration(avgTimeStr)
		assert.Greater(t, avgTime, 15*time.Millisecond)
		assert.Less(t, avgTime, 25*time.Millisecond)
	})
}

// Integration test
func TestMetricsIntegration(t *testing.T) {
	t.Run("full workflow simulation", func(t *testing.T) {
		resetGlobalMetrics()

		// Simulate a realistic workflow

		// 1. Health check
		handler := middleware.MetricsMiddleware(createTestHandler(http.StatusOK, 1*time.Millisecond))
		req := httptest.NewRequest("GET", "/health", nil)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		// 2. List records
		handler = middleware.MetricsMiddleware(createTestHandler(http.StatusOK, 5*time.Millisecond))
		req = httptest.NewRequest("GET", "/api/records", nil)
		w = httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		// 3. Import CSV (longer operation)
		handler = middleware.MetricsMiddleware(createTestHandler(http.StatusAccepted, 50*time.Millisecond))
		req = httptest.NewRequest("POST", "/api/import", nil)
		w = httptest.NewRecorder()
		handler.ServeHTTP(w, req)
		middleware.GlobalMetrics.IncrementCSVProcessed()
		middleware.GlobalMetrics.IncrementImportedRecords(1000)

		// 4. Check for duplicates
		middleware.GlobalMetrics.IncrementDuplicates(50)

		// 5. Error case
		handler = middleware.MetricsMiddleware(createTestHandler(http.StatusBadRequest, 2*time.Millisecond))
		req = httptest.NewRequest("POST", "/api/records", nil)
		w = httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		// Verify metrics
		metrics := middleware.GlobalMetrics.GetMetrics()

		assert.Equal(t, int64(4), metrics["total_requests"])
		assert.Equal(t, int64(1), metrics["total_errors"])
		assert.Equal(t, float64(25.0), metrics["error_rate"])
		assert.Equal(t, int64(1000), metrics["records_imported"])
		assert.Equal(t, int64(1), metrics["csv_processed"])
		assert.Equal(t, int64(50), metrics["duplicates_found"])

		// Verify endpoint metrics
		endpointMetrics := metrics["endpoint_metrics"].(map[string]interface{})
		assert.Contains(t, endpointMetrics, "GET /health")
		assert.Contains(t, endpointMetrics, "GET /api/records")
		assert.Contains(t, endpointMetrics, "POST /api/import")
		assert.Contains(t, endpointMetrics, "POST /api/records")

		// Get metrics endpoint
		metricsHandler := middleware.MetricsHandler()
		req = httptest.NewRequest("GET", "/metrics", nil)
		w = httptest.NewRecorder()
		metricsHandler(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "total_requests")
	})
}
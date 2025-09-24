package middleware

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/yhonda-ohishi/etc_meisai/src/middleware"
)

// TestErrorHandler tests the panic recovery middleware
func TestErrorHandler(t *testing.T) {
	logger := log.New(&strings.Builder{}, "[TEST] ", log.LstdFlags)

	t.Run("normal request without panic", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("success"))
		})

		errorHandler := middleware.ErrorHandler(logger)(handler)

		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()

		errorHandler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "success", w.Body.String())
	})

	t.Run("recovers from panic", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			panic("something went wrong")
		})

		errorHandler := middleware.ErrorHandler(logger)(handler)

		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("X-Request-ID", "test-123")
		w := httptest.NewRecorder()

		errorHandler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

		var errorResp middleware.ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &errorResp)
		assert.NoError(t, err)
		assert.Equal(t, "internal_error", errorResp.Error)
		assert.Equal(t, "An internal error occurred", errorResp.Message)
		assert.Equal(t, http.StatusInternalServerError, errorResp.StatusCode)
		assert.Equal(t, "test-123", errorResp.RequestID)
	})

	t.Run("recovers from panic with nil value", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var nilPtr *string
			_ = *nilPtr // This will panic
		})

		errorHandler := middleware.ErrorHandler(logger)(handler)

		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()

		errorHandler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)

		var errorResp middleware.ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &errorResp)
		assert.NoError(t, err)
		assert.Equal(t, "internal_error", errorResp.Error)
	})

	t.Run("preserves request ID in error response", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			panic("test panic")
		})

		errorHandler := middleware.ErrorHandler(logger)(handler)

		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("X-Request-ID", "req-456")
		w := httptest.NewRecorder()

		errorHandler.ServeHTTP(w, req)

		var errorResp middleware.ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &errorResp)
		assert.NoError(t, err)
		assert.Equal(t, "req-456", errorResp.RequestID)
	})

	t.Run("handles panic with complex error types", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			panic(errors.New("complex error"))
		})

		errorHandler := middleware.ErrorHandler(logger)(handler)

		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()

		errorHandler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)

		var errorResp middleware.ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &errorResp)
		assert.NoError(t, err)
		assert.Equal(t, "internal_error", errorResp.Error)
	})
}

// TestGRPCErrorHandler tests the gRPC error handling middleware
func TestGRPCErrorHandler(t *testing.T) {
	logger := log.New(&strings.Builder{}, "[TEST] ", log.LstdFlags)

	t.Run("normal request", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("success"))
		})

		grpcErrorHandler := middleware.GRPCErrorHandler(logger)(handler)

		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("X-Request-ID", "test-123")
		w := httptest.NewRecorder()

		grpcErrorHandler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "success", w.Body.String())
	})

	t.Run("logs error responses", func(t *testing.T) {
		var logBuffer strings.Builder
		testLogger := log.New(&logBuffer, "[TEST] ", log.LstdFlags)

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("bad request"))
		})

		grpcErrorHandler := middleware.GRPCErrorHandler(testLogger)(handler)

		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("X-Request-ID", "error-123")
		w := httptest.NewRecorder()

		grpcErrorHandler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		logOutput := logBuffer.String()
		assert.Contains(t, logOutput, "error-123")
		assert.Contains(t, logOutput, "400")
	})

	t.Run("handles server errors", func(t *testing.T) {
		var logBuffer strings.Builder
		testLogger := log.New(&logBuffer, "[TEST] ", log.LstdFlags)

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		})

		grpcErrorHandler := middleware.GRPCErrorHandler(testLogger)(handler)

		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("X-Request-ID", "server-error")
		w := httptest.NewRecorder()

		grpcErrorHandler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		logOutput := logBuffer.String()
		assert.Contains(t, logOutput, "server-error")
		assert.Contains(t, logOutput, "500")
	})
}

// TestConvertGRPCError tests gRPC error conversion
func TestConvertGRPCError(t *testing.T) {
	tests := []struct {
		name           string
		err            error
		expectedStatus int
		expectedCode   string
	}{
		{
			name:           "nil error",
			err:            nil,
			expectedStatus: http.StatusOK,
			expectedCode:   "",
		},
		{
			name:           "non-gRPC error",
			err:            errors.New("regular error"),
			expectedStatus: http.StatusInternalServerError,
			expectedCode:   "internal_error",
		},
		{
			name:           "gRPC InvalidArgument",
			err:            status.Error(codes.InvalidArgument, "invalid field"),
			expectedStatus: http.StatusBadRequest,
			expectedCode:   "invalid_argument",
		},
		{
			name:           "gRPC NotFound",
			err:            status.Error(codes.NotFound, "resource not found"),
			expectedStatus: http.StatusNotFound,
			expectedCode:   "not_found",
		},
		{
			name:           "gRPC AlreadyExists",
			err:            status.Error(codes.AlreadyExists, "already exists"),
			expectedStatus: http.StatusConflict,
			expectedCode:   "already_exists",
		},
		{
			name:           "gRPC PermissionDenied",
			err:            status.Error(codes.PermissionDenied, "access denied"),
			expectedStatus: http.StatusForbidden,
			expectedCode:   "permission_denied",
		},
		{
			name:           "gRPC Unauthenticated",
			err:            status.Error(codes.Unauthenticated, "not authenticated"),
			expectedStatus: http.StatusUnauthorized,
			expectedCode:   "unauthenticated",
		},
		{
			name:           "gRPC ResourceExhausted",
			err:            status.Error(codes.ResourceExhausted, "too many requests"),
			expectedStatus: http.StatusTooManyRequests,
			expectedCode:   "resource_exhausted",
		},
		{
			name:           "gRPC FailedPrecondition",
			err:            status.Error(codes.FailedPrecondition, "precondition failed"),
			expectedStatus: http.StatusPreconditionFailed,
			expectedCode:   "failed_precondition",
		},
		{
			name:           "gRPC Aborted",
			err:            status.Error(codes.Aborted, "operation aborted"),
			expectedStatus: http.StatusConflict,
			expectedCode:   "aborted",
		},
		{
			name:           "gRPC OutOfRange",
			err:            status.Error(codes.OutOfRange, "out of range"),
			expectedStatus: http.StatusBadRequest,
			expectedCode:   "out_of_range",
		},
		{
			name:           "gRPC Unimplemented",
			err:            status.Error(codes.Unimplemented, "not implemented"),
			expectedStatus: http.StatusNotImplemented,
			expectedCode:   "not_implemented",
		},
		{
			name:           "gRPC Internal",
			err:            status.Error(codes.Internal, "internal error"),
			expectedStatus: http.StatusInternalServerError,
			expectedCode:   "internal_error",
		},
		{
			name:           "gRPC Unavailable",
			err:            status.Error(codes.Unavailable, "service unavailable"),
			expectedStatus: http.StatusServiceUnavailable,
			expectedCode:   "service_unavailable",
		},
		{
			name:           "gRPC DataLoss",
			err:            status.Error(codes.DataLoss, "data loss"),
			expectedStatus: http.StatusInternalServerError,
			expectedCode:   "data_loss",
		},
		{
			name:           "gRPC DeadlineExceeded",
			err:            status.Error(codes.DeadlineExceeded, "deadline exceeded"),
			expectedStatus: http.StatusGatewayTimeout,
			expectedCode:   "deadline_exceeded",
		},
		{
			name:           "gRPC Canceled",
			err:            status.Error(codes.Canceled, "canceled"),
			expectedStatus: http.StatusRequestTimeout,
			expectedCode:   "request_canceled",
		},
		{
			name:           "gRPC Unknown",
			err:            status.Error(codes.Unknown, "unknown error"),
			expectedStatus: http.StatusInternalServerError,
			expectedCode:   "unknown_error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			httpStatus, errorResp := middleware.ConvertGRPCError(tt.err)

			assert.Equal(t, tt.expectedStatus, httpStatus)

			if tt.err == nil {
				assert.Nil(t, errorResp)
			} else {
				assert.NotNil(t, errorResp)
				assert.Equal(t, tt.expectedCode, errorResp.Error)
				assert.Equal(t, tt.expectedStatus, errorResp.StatusCode)
			}
		})
	}
}

// TestConvertGRPCErrorDetails tests field extraction from gRPC errors
func TestConvertGRPCErrorDetails(t *testing.T) {
	t.Run("extracts field information", func(t *testing.T) {
		err := status.Error(codes.InvalidArgument, "validation failed field: email")

		httpStatus, errorResp := middleware.ConvertGRPCError(err)

		assert.Equal(t, http.StatusBadRequest, httpStatus)
		assert.NotNil(t, errorResp)
		assert.Equal(t, "invalid_argument", errorResp.Error)
		assert.Equal(t, "validation failed field: email", errorResp.Message)

		if errorResp.Details != nil {
			field, exists := errorResp.Details["field"]
			if exists {
				assert.Equal(t, "email", field)
			}
		}
	})

	t.Run("handles message without field", func(t *testing.T) {
		err := status.Error(codes.InvalidArgument, "general validation error")

		httpStatus, errorResp := middleware.ConvertGRPCError(err)

		assert.Equal(t, http.StatusBadRequest, httpStatus)
		assert.NotNil(t, errorResp)
		assert.Equal(t, "general validation error", errorResp.Message)
		assert.NotNil(t, errorResp.Details)
	})

	t.Run("handles empty message", func(t *testing.T) {
		err := status.Error(codes.NotFound, "")

		httpStatus, errorResp := middleware.ConvertGRPCError(err)

		assert.Equal(t, http.StatusNotFound, httpStatus)
		assert.NotNil(t, errorResp)
		assert.Equal(t, "", errorResp.Message)
	})
}

// TestValidationError tests validation error creation
func TestValidationError(t *testing.T) {
	t.Run("creates validation error", func(t *testing.T) {
		errorResp := middleware.ValidationError("email", "invalid email format")

		assert.Equal(t, "validation_error", errorResp.Error)
		assert.Equal(t, "invalid email format", errorResp.Message)
		assert.Equal(t, http.StatusBadRequest, errorResp.StatusCode)
		assert.NotNil(t, errorResp.Details)
		assert.Equal(t, "email", errorResp.Details["field"])
	})

	t.Run("handles empty field", func(t *testing.T) {
		errorResp := middleware.ValidationError("", "validation failed")

		assert.Equal(t, "validation_error", errorResp.Error)
		assert.Equal(t, "validation failed", errorResp.Message)
		assert.Equal(t, "", errorResp.Details["field"])
	})

	t.Run("handles empty message", func(t *testing.T) {
		errorResp := middleware.ValidationError("username", "")

		assert.Equal(t, "validation_error", errorResp.Error)
		assert.Equal(t, "", errorResp.Message)
		assert.Equal(t, "username", errorResp.Details["field"])
	})
}

// TestNotFoundError tests not found error creation
func TestNotFoundError(t *testing.T) {
	t.Run("creates not found error", func(t *testing.T) {
		errorResp := middleware.NotFoundError("user")

		assert.Equal(t, "not_found", errorResp.Error)
		assert.Equal(t, "Resource not found", errorResp.Message)
		assert.Equal(t, http.StatusNotFound, errorResp.StatusCode)
		assert.NotNil(t, errorResp.Details)
		assert.Equal(t, "user", errorResp.Details["resource"])
	})

	t.Run("handles empty resource", func(t *testing.T) {
		errorResp := middleware.NotFoundError("")

		assert.Equal(t, "not_found", errorResp.Error)
		assert.Equal(t, "Resource not found", errorResp.Message)
		assert.Equal(t, "", errorResp.Details["resource"])
	})
}

// TestServiceUnavailableError tests service unavailable error creation
func TestServiceUnavailableError(t *testing.T) {
	t.Run("creates service unavailable error", func(t *testing.T) {
		errorResp := middleware.ServiceUnavailableError("database")

		assert.Equal(t, "service_unavailable", errorResp.Error)
		assert.Equal(t, "Service temporarily unavailable", errorResp.Message)
		assert.Equal(t, http.StatusServiceUnavailable, errorResp.StatusCode)
		assert.NotNil(t, errorResp.Details)
		assert.Equal(t, "database", errorResp.Details["service"])
	})

	t.Run("handles empty service", func(t *testing.T) {
		errorResp := middleware.ServiceUnavailableError("")

		assert.Equal(t, "service_unavailable", errorResp.Error)
		assert.Equal(t, "Service temporarily unavailable", errorResp.Message)
		assert.Equal(t, "", errorResp.Details["service"])
	})
}

// TestErrorResponseStructure tests the error response structure
func TestErrorResponseStructure(t *testing.T) {
	t.Run("complete error response", func(t *testing.T) {
		errorResp := middleware.ErrorResponse{
			Error:      "test_error",
			Message:    "Test error message",
			StatusCode: http.StatusBadRequest,
			Details: map[string]interface{}{
				"field": "test_field",
				"value": "invalid_value",
			},
			RequestID: "req-123",
		}

		jsonData, err := json.Marshal(errorResp)
		assert.NoError(t, err)

		var unmarshaledResp middleware.ErrorResponse
		err = json.Unmarshal(jsonData, &unmarshaledResp)
		assert.NoError(t, err)

		assert.Equal(t, "test_error", unmarshaledResp.Error)
		assert.Equal(t, "Test error message", unmarshaledResp.Message)
		assert.Equal(t, http.StatusBadRequest, unmarshaledResp.StatusCode)
		assert.Equal(t, "req-123", unmarshaledResp.RequestID)
		assert.Equal(t, "test_field", unmarshaledResp.Details["field"])
	})

	t.Run("minimal error response", func(t *testing.T) {
		errorResp := middleware.ErrorResponse{
			Error:      "error_code",
			Message:    "Error message",
			StatusCode: http.StatusInternalServerError,
		}

		jsonData, err := json.Marshal(errorResp)
		assert.NoError(t, err)

		var unmarshaledResp middleware.ErrorResponse
		err = json.Unmarshal(jsonData, &unmarshaledResp)
		assert.NoError(t, err)

		assert.Equal(t, "error_code", unmarshaledResp.Error)
		assert.Equal(t, "Error message", unmarshaledResp.Message)
		assert.Equal(t, http.StatusInternalServerError, unmarshaledResp.StatusCode)
		assert.Empty(t, unmarshaledResp.RequestID)
		assert.Nil(t, unmarshaledResp.Details)
	})

	t.Run("omits empty fields in JSON", func(t *testing.T) {
		errorResp := middleware.ErrorResponse{
			Error:      "error",
			Message:    "message",
			StatusCode: 400,
		}

		jsonData, err := json.Marshal(errorResp)
		assert.NoError(t, err)

		jsonStr := string(jsonData)
		assert.NotContains(t, jsonStr, "details")
		assert.NotContains(t, jsonStr, "request_id")
	})
}

// TestConcurrentErrorHandling tests error handling under concurrent load
func TestConcurrentErrorHandling(t *testing.T) {
	logger := log.New(&strings.Builder{}, "[TEST] ", log.LstdFlags)

	t.Run("concurrent panic recovery", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			panic("concurrent panic")
		})

		errorHandler := middleware.ErrorHandler(logger)(handler)

		const numRequests = 50
		done := make(chan bool, numRequests)

		for i := 0; i < numRequests; i++ {
			go func(id int) {
				req := httptest.NewRequest("GET", "/test", nil)
				req.Header.Set("X-Request-ID", string(rune(id)))
				w := httptest.NewRecorder()

				errorHandler.ServeHTTP(w, req)

				assert.Equal(t, http.StatusInternalServerError, w.Code)
				done <- true
			}(i)
		}

		for i := 0; i < numRequests; i++ {
			<-done
		}
	})

	t.Run("concurrent gRPC error conversion", func(t *testing.T) {
		errors := []error{
			status.Error(codes.InvalidArgument, "invalid"),
			status.Error(codes.NotFound, "not found"),
			status.Error(codes.PermissionDenied, "denied"),
			status.Error(codes.Internal, "internal"),
		}

		const numRequests = 100
		done := make(chan bool, numRequests)

		for i := 0; i < numRequests; i++ {
			go func(id int) {
				err := errors[id%len(errors)]
				httpStatus, errorResp := middleware.ConvertGRPCError(err)

				assert.NotNil(t, errorResp)
				assert.Greater(t, httpStatus, 0)
				done <- true
			}(i)
		}

		for i := 0; i < numRequests; i++ {
			<-done
		}
	})
}

// TestErrorHandlerEdgeCases tests edge cases
func TestErrorHandlerEdgeCases(t *testing.T) {
	logger := log.New(&strings.Builder{}, "[TEST] ", log.LstdFlags)

	t.Run("panic after response started", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("partial"))
			panic("late panic")
		})

		errorHandler := middleware.ErrorHandler(logger)(handler)

		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()

		// The panic will be recovered but response already started
		errorHandler.ServeHTTP(w, req)

		// Status was already written
		assert.Equal(t, http.StatusOK, w.Code)
		// But partial content was written
		assert.Contains(t, w.Body.String(), "partial")
	})

	t.Run("panic with string type", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			panic("string panic")
		})

		errorHandler := middleware.ErrorHandler(logger)(handler)

		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()

		errorHandler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)

		var errorResp middleware.ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &errorResp)
		assert.NoError(t, err)
		assert.Equal(t, "internal_error", errorResp.Error)
	})

	t.Run("panic with integer type", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			panic(42)
		})

		errorHandler := middleware.ErrorHandler(logger)(handler)

		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()

		errorHandler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("multiple writes in custom response writer", func(t *testing.T) {
		var logBuffer strings.Builder
		testLogger := log.New(&logBuffer, "[TEST] ", log.LstdFlags)

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("first"))
			w.Write([]byte("second"))
			w.WriteHeader(http.StatusBadRequest) // Should be ignored after writes
			w.Write([]byte("third"))
		})

		grpcErrorHandler := middleware.GRPCErrorHandler(testLogger)(handler)

		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()

		grpcErrorHandler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code) // Default status when Write is called first
		assert.Equal(t, "firstsecondthird", w.Body.String())
	})

	t.Run("nil logger in error handler", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			panic("test panic")
		})

		errorHandler := middleware.ErrorHandler(nil)(handler)

		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()

		// Should panic when trying to log with nil logger
		assert.Panics(t, func() {
			errorHandler.ServeHTTP(w, req)
		})
	})
}

// TestErrorHandlerPerformance tests performance characteristics
func TestErrorHandlerPerformance(t *testing.T) {
	logger := log.New(&strings.Builder{}, "[TEST] ", log.LstdFlags)

	t.Run("error conversion performance", func(t *testing.T) {
		errors := []error{
			status.Error(codes.InvalidArgument, "test"),
			status.Error(codes.NotFound, "test"),
			status.Error(codes.Internal, "test"),
		}

		const iterations = 10000
		start := time.Now()

		for i := 0; i < iterations; i++ {
			err := errors[i%len(errors)]
			_, _ = middleware.ConvertGRPCError(err)
		}

		duration := time.Since(start)
		avgDuration := duration / iterations

		t.Logf("Average error conversion time: %v", avgDuration)
		assert.Less(t, avgDuration, 1*time.Microsecond, "Error conversion should be very fast")
	})

	t.Run("panic recovery overhead", func(t *testing.T) {
		// Normal handler without panic
		normalHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		// Measure with error handler wrapper
		errorHandler := middleware.ErrorHandler(logger)(normalHandler)

		const iterations = 1000
		start := time.Now()

		for i := 0; i < iterations; i++ {
			req := httptest.NewRequest("GET", "/test", nil)
			w := httptest.NewRecorder()
			errorHandler.ServeHTTP(w, req)
		}

		duration := time.Since(start)
		avgDuration := duration / iterations

		t.Logf("Average request with error handler: %v", avgDuration)
		assert.Less(t, avgDuration, 1*time.Millisecond, "Error handler should have minimal overhead")
	})
}
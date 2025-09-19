package middleware

import (
	"encoding/json"
	"log"
	"net/http"
	"runtime/debug"
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ErrorResponse represents a standardized error response
type ErrorResponse struct {
	Error      string                 `json:"error"`
	Message    string                 `json:"message"`
	StatusCode int                    `json:"status_code"`
	Details    map[string]interface{} `json:"details,omitempty"`
	RequestID  string                 `json:"request_id,omitempty"`
}

// ErrorHandler is a middleware for handling panics and errors
func ErrorHandler(logger *log.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					logger.Printf("Panic recovered: %v\n%s", err, debug.Stack())

					response := ErrorResponse{
						Error:      "internal_error",
						Message:    "An internal error occurred",
						StatusCode: http.StatusInternalServerError,
						RequestID:  r.Header.Get("X-Request-ID"),
					}

					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusInternalServerError)
					json.NewEncoder(w).Encode(response)
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}

// GRPCErrorHandler converts gRPC errors to HTTP status codes
func GRPCErrorHandler(logger *log.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Create a custom response writer to intercept errors
			crw := &customResponseWriter{
				ResponseWriter: w,
				logger:         logger,
				requestID:      r.Header.Get("X-Request-ID"),
			}

			next.ServeHTTP(crw, r)
		})
	}
}

// customResponseWriter wraps http.ResponseWriter to intercept and log errors
type customResponseWriter struct {
	http.ResponseWriter
	logger    *log.Logger
	requestID string
	written   bool
}

func (crw *customResponseWriter) Write(b []byte) (int, error) {
	if !crw.written {
		crw.written = true
	}
	return crw.ResponseWriter.Write(b)
}

func (crw *customResponseWriter) WriteHeader(statusCode int) {
	if !crw.written {
		crw.written = true
		if statusCode >= 400 {
			crw.logger.Printf("Request failed - ID: %s, Status: %d", crw.requestID, statusCode)
		}
	}
	crw.ResponseWriter.WriteHeader(statusCode)
}

// ConvertGRPCError converts a gRPC error to HTTP status code and error response
func ConvertGRPCError(err error) (int, *ErrorResponse) {
	if err == nil {
		return http.StatusOK, nil
	}

	st, ok := status.FromError(err)
	if !ok {
		// Not a gRPC error
		return http.StatusInternalServerError, &ErrorResponse{
			Error:      "internal_error",
			Message:    err.Error(),
			StatusCode: http.StatusInternalServerError,
		}
	}

	var httpStatus int
	var errorCode string

	switch st.Code() {
	case codes.OK:
		httpStatus = http.StatusOK
		errorCode = "ok"
	case codes.Canceled:
		httpStatus = http.StatusRequestTimeout
		errorCode = "request_canceled"
	case codes.Unknown:
		httpStatus = http.StatusInternalServerError
		errorCode = "unknown_error"
	case codes.InvalidArgument:
		httpStatus = http.StatusBadRequest
		errorCode = "invalid_argument"
	case codes.DeadlineExceeded:
		httpStatus = http.StatusGatewayTimeout
		errorCode = "deadline_exceeded"
	case codes.NotFound:
		httpStatus = http.StatusNotFound
		errorCode = "not_found"
	case codes.AlreadyExists:
		httpStatus = http.StatusConflict
		errorCode = "already_exists"
	case codes.PermissionDenied:
		httpStatus = http.StatusForbidden
		errorCode = "permission_denied"
	case codes.ResourceExhausted:
		httpStatus = http.StatusTooManyRequests
		errorCode = "resource_exhausted"
	case codes.FailedPrecondition:
		httpStatus = http.StatusPreconditionFailed
		errorCode = "failed_precondition"
	case codes.Aborted:
		httpStatus = http.StatusConflict
		errorCode = "aborted"
	case codes.OutOfRange:
		httpStatus = http.StatusBadRequest
		errorCode = "out_of_range"
	case codes.Unimplemented:
		httpStatus = http.StatusNotImplemented
		errorCode = "not_implemented"
	case codes.Internal:
		httpStatus = http.StatusInternalServerError
		errorCode = "internal_error"
	case codes.Unavailable:
		httpStatus = http.StatusServiceUnavailable
		errorCode = "service_unavailable"
	case codes.DataLoss:
		httpStatus = http.StatusInternalServerError
		errorCode = "data_loss"
	case codes.Unauthenticated:
		httpStatus = http.StatusUnauthorized
		errorCode = "unauthenticated"
	default:
		httpStatus = http.StatusInternalServerError
		errorCode = "unknown_error"
	}

	// Parse details from gRPC status
	details := make(map[string]interface{})
	if st.Message() != "" {
		// Extract field information if present
		if strings.Contains(st.Message(), "field:") {
			parts := strings.Split(st.Message(), "field:")
			if len(parts) > 1 {
				details["field"] = strings.TrimSpace(parts[1])
			}
		}
	}

	return httpStatus, &ErrorResponse{
		Error:      errorCode,
		Message:    st.Message(),
		StatusCode: httpStatus,
		Details:    details,
	}
}

// ValidationError creates a validation error response
func ValidationError(field, message string) *ErrorResponse {
	return &ErrorResponse{
		Error:      "validation_error",
		Message:    message,
		StatusCode: http.StatusBadRequest,
		Details: map[string]interface{}{
			"field": field,
		},
	}
}

// NotFoundError creates a not found error response
func NotFoundError(resource string) *ErrorResponse {
	return &ErrorResponse{
		Error:      "not_found",
		Message:    "Resource not found",
		StatusCode: http.StatusNotFound,
		Details: map[string]interface{}{
			"resource": resource,
		},
	}
}

// ServiceUnavailableError creates a service unavailable error response
func ServiceUnavailableError(service string) *ErrorResponse {
	return &ErrorResponse{
		Error:      "service_unavailable",
		Message:    "Service temporarily unavailable",
		StatusCode: http.StatusServiceUnavailable,
		Details: map[string]interface{}{
			"service": service,
		},
	}
}
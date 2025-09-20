package handlers

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// GRPCErrorHandler provides centralized gRPC error handling
type GRPCErrorHandler struct{}

// NewGRPCErrorHandler creates a new gRPC error handler
func NewGRPCErrorHandler() *GRPCErrorHandler {
	return &GRPCErrorHandler{}
}

// HandleGRPCError converts gRPC errors to appropriate HTTP responses
func (e *GRPCErrorHandler) HandleGRPCError(err error) (int, string, string) {
	if err == nil {
		return http.StatusOK, "success", "Operation completed successfully"
	}

	// Check if it's a gRPC status error
	if st, ok := status.FromError(err); ok {
		return e.mapGRPCStatusToHTTP(st)
	}

	// Check for context errors
	if errors.Is(err, context.DeadlineExceeded) {
		return http.StatusRequestTimeout, "timeout", "Request timeout exceeded"
	}

	if errors.Is(err, context.Canceled) {
		return http.StatusRequestTimeout, "canceled", "Request was canceled"
	}

	// Default to internal server error
	return http.StatusInternalServerError, "internal_error", err.Error()
}

// mapGRPCStatusToHTTP maps gRPC status codes to HTTP status codes
func (e *GRPCErrorHandler) mapGRPCStatusToHTTP(st *status.Status) (int, string, string) {
	code := st.Code()
	message := st.Message()

	switch code {
	case codes.OK:
		return http.StatusOK, "success", "Operation completed successfully"
	case codes.InvalidArgument:
		return http.StatusBadRequest, "invalid_argument", message
	case codes.NotFound:
		return http.StatusNotFound, "not_found", message
	case codes.AlreadyExists:
		return http.StatusConflict, "already_exists", message
	case codes.PermissionDenied:
		return http.StatusForbidden, "permission_denied", message
	case codes.Unauthenticated:
		return http.StatusUnauthorized, "unauthenticated", message
	case codes.ResourceExhausted:
		return http.StatusTooManyRequests, "resource_exhausted", message
	case codes.FailedPrecondition:
		return http.StatusPreconditionFailed, "failed_precondition", message
	case codes.Aborted:
		return http.StatusConflict, "aborted", message
	case codes.OutOfRange:
		return http.StatusBadRequest, "out_of_range", message
	case codes.Unimplemented:
		return http.StatusNotImplemented, "unimplemented", message
	case codes.Internal:
		return http.StatusInternalServerError, "internal_error", message
	case codes.Unavailable:
		return http.StatusServiceUnavailable, "unavailable", message
	case codes.DataLoss:
		return http.StatusInternalServerError, "data_loss", message
	case codes.DeadlineExceeded:
		return http.StatusRequestTimeout, "deadline_exceeded", message
	case codes.Canceled:
		return http.StatusRequestTimeout, "canceled", message
	default:
		return http.StatusInternalServerError, "unknown_error",
			fmt.Sprintf("Unknown gRPC error: %s", message)
	}
}

// ErrorDetail provides additional error context
type ErrorDetail struct {
	Code       string      `json:"code"`
	Message    string      `json:"message"`
	Details    interface{} `json:"details,omitempty"`
	GRPCCode   string      `json:"grpc_code,omitempty"`
	Timestamp  string      `json:"timestamp"`
	RequestID  string      `json:"request_id,omitempty"`
}

// CreateErrorDetail creates a detailed error response
func (e *GRPCErrorHandler) CreateErrorDetail(err error, requestID string) *ErrorDetail {
	httpStatus, errorCode, message := e.HandleGRPCError(err)

	detail := &ErrorDetail{
		Code:      errorCode,
		Message:   message,
		Timestamp: fmt.Sprintf("%d", httpStatus), // Temporary use HTTP status
		RequestID: requestID,
	}

	// Add gRPC specific details if available
	if st, ok := status.FromError(err); ok {
		detail.GRPCCode = st.Code().String()
		if len(st.Details()) > 0 {
			detail.Details = st.Details()
		}
	}

	return detail
}

// ValidationError represents validation errors
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Value   string `json:"value,omitempty"`
}

// BusinessLogicError represents business logic errors
type BusinessLogicError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Context string `json:"context,omitempty"`
}

// HandleValidationErrors handles validation-specific errors
func (e *GRPCErrorHandler) HandleValidationErrors(errors []ValidationError) (int, string, interface{}) {
	return http.StatusBadRequest, "validation_failed", map[string]interface{}{
		"validation_errors": errors,
		"message":          "Request validation failed",
	}
}

// HandleBusinessLogicError handles business logic errors
func (e *GRPCErrorHandler) HandleBusinessLogicError(err BusinessLogicError) (int, string, interface{}) {
	return http.StatusUnprocessableEntity, err.Code, map[string]interface{}{
		"business_error": err,
		"message":       err.Message,
	}
}
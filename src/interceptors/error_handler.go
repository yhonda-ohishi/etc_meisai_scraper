package interceptors

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"regexp"
	"runtime/debug"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ErrorConfig holds configuration for the error handling interceptor
type ErrorConfig struct {
	Logger                *slog.Logger
	IncludeStackTrace     bool
	SanitizeErrors        bool
	EnableMetrics         bool
	ProductionMode        bool
	AlertOnCriticalErrors bool
	ErrorCodeMapping      map[string]codes.Code
}

// ServiceError represents a custom service error with additional context
type ServiceError struct {
	Code    string
	Message string
	Details map[string]interface{}
	Cause   error
}

func (e *ServiceError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %s (caused by: %v)", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// NewErrorConfig creates a new error handling configuration
func NewErrorConfig() *ErrorConfig {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelError,
	}))

	productionMode := os.Getenv("ENVIRONMENT") == "production"

	return &ErrorConfig{
		Logger:                logger,
		IncludeStackTrace:     !productionMode,
		SanitizeErrors:        productionMode,
		EnableMetrics:         true,
		ProductionMode:        productionMode,
		AlertOnCriticalErrors: productionMode,
		ErrorCodeMapping:      getDefaultErrorCodeMapping(),
	}
}

// getDefaultErrorCodeMapping returns default error code to gRPC status code mapping
func getDefaultErrorCodeMapping() map[string]codes.Code {
	return map[string]codes.Code{
		"VALIDATION_ERROR":      codes.InvalidArgument,
		"NOT_FOUND":            codes.NotFound,
		"ALREADY_EXISTS":       codes.AlreadyExists,
		"PERMISSION_DENIED":    codes.PermissionDenied,
		"UNAUTHENTICATED":      codes.Unauthenticated,
		"RESOURCE_EXHAUSTED":   codes.ResourceExhausted,
		"FAILED_PRECONDITION": codes.FailedPrecondition,
		"ABORTED":             codes.Aborted,
		"OUT_OF_RANGE":        codes.OutOfRange,
		"UNIMPLEMENTED":       codes.Unimplemented,
		"INTERNAL":            codes.Internal,
		"UNAVAILABLE":         codes.Unavailable,
		"DATA_LOSS":           codes.DataLoss,
		"TIMEOUT":             codes.DeadlineExceeded,
		"CANCELLED":           codes.Canceled,
		"UNKNOWN":             codes.Unknown,
		"DATABASE_ERROR":      codes.Internal,
		"NETWORK_ERROR":       codes.Unavailable,
		"PARSING_ERROR":       codes.InvalidArgument,
		"EXTERNAL_API_ERROR":  codes.Unavailable,
		"RATE_LIMITED":        codes.ResourceExhausted,
	}
}

// UnaryErrorHandlerInterceptor creates a unary server interceptor for error handling
func UnaryErrorHandlerInterceptor(config *ErrorConfig) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (resp interface{}, err error) {
		// Recover from panics
		defer func() {
			if r := recover(); r != nil {
				err = handlePanic(r, config, ctx, info.FullMethod)
			}
		}()

		// Call the handler
		resp, err = handler(ctx, req)

		// Process any error
		if err != nil {
			err = processError(err, config, ctx, info.FullMethod)
		}

		return resp, err
	}
}

// StreamErrorHandlerInterceptor creates a stream server interceptor for error handling
func StreamErrorHandlerInterceptor(config *ErrorConfig) grpc.StreamServerInterceptor {
	return func(
		srv interface{},
		stream grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) (err error) {
		// Recover from panics
		defer func() {
			if r := recover(); r != nil {
				err = handlePanic(r, config, stream.Context(), info.FullMethod)
			}
		}()

		// Call the handler
		err = handler(srv, stream)

		// Process any error
		if err != nil {
			err = processError(err, config, stream.Context(), info.FullMethod)
		}

		return err
	}
}

// processError processes and converts errors to appropriate gRPC status
func processError(err error, config *ErrorConfig, ctx context.Context, method string) error {
	if err == nil {
		return nil
	}

	// Log the error
	logError(config, ctx, method, err)

	// Convert to gRPC status
	grpcErr := convertToGRPCError(err, config)

	// Record metrics if enabled
	if config.EnableMetrics {
		recordErrorMetrics(err, method)
	}

	// Send alerts for critical errors if enabled
	if config.AlertOnCriticalErrors && isCriticalError(err) {
		sendErrorAlert(ctx, method, err)
	}

	return grpcErr
}

// convertToGRPCError converts various error types to gRPC status errors
func convertToGRPCError(err error, config *ErrorConfig) error {
	// If already a gRPC status error, return as-is or sanitize
	if st, ok := status.FromError(err); ok {
		if config.SanitizeErrors {
			return sanitizeGRPCError(st)
		}
		return err
	}

	// Handle custom service errors
	if serviceErr, ok := err.(*ServiceError); ok {
		grpcCode := codes.Internal // default
		if code, exists := config.ErrorCodeMapping[serviceErr.Code]; exists {
			grpcCode = code
		}

		message := serviceErr.Message
		if config.SanitizeErrors {
			message = sanitizeErrorMessage(message)
		}

		return status.Error(grpcCode, message)
	}

	// Handle common Go errors
	grpcCode, message := classifyError(err)

	if config.SanitizeErrors {
		message = sanitizeErrorMessage(message)
	}

	return status.Error(grpcCode, message)
}

// classifyError classifies common Go errors to gRPC codes
func classifyError(err error) (codes.Code, string) {
	errStr := strings.ToLower(err.Error())

	switch {
	case strings.Contains(errStr, "not found"):
		return codes.NotFound, err.Error()
	case strings.Contains(errStr, "already exists"):
		return codes.AlreadyExists, err.Error()
	case strings.Contains(errStr, "permission denied"):
		return codes.PermissionDenied, err.Error()
	case strings.Contains(errStr, "unauthorized") || strings.Contains(errStr, "unauthenticated"):
		return codes.Unauthenticated, err.Error()
	case strings.Contains(errStr, "timeout") || strings.Contains(errStr, "deadline"):
		return codes.DeadlineExceeded, err.Error()
	case strings.Contains(errStr, "cancelled"):
		return codes.Canceled, err.Error()
	case strings.Contains(errStr, "invalid") || strings.Contains(errStr, "bad request"):
		return codes.InvalidArgument, err.Error()
	case strings.Contains(errStr, "unavailable") || strings.Contains(errStr, "connection"):
		return codes.Unavailable, err.Error()
	case strings.Contains(errStr, "rate limit"):
		return codes.ResourceExhausted, err.Error()
	case strings.Contains(errStr, "unimplemented"):
		return codes.Unimplemented, err.Error()
	default:
		return codes.Internal, err.Error()
	}
}

// handlePanic handles recovered panics and converts them to errors
func handlePanic(r interface{}, config *ErrorConfig, ctx context.Context, method string) error {
	stack := debug.Stack()

	// Log panic with stack trace
	config.Logger.ErrorContext(ctx, "panic recovered",
		"method", method,
		"panic", r,
		"stack_trace", string(stack),
	)

	// Create error message
	var errMsg string
	if config.ProductionMode {
		errMsg = "internal server error"
	} else {
		errMsg = fmt.Sprintf("panic recovered: %v", r)
	}

	// Send critical alert
	if config.AlertOnCriticalErrors {
		sendPanicAlert(ctx, method, r, stack)
	}

	return status.Error(codes.Internal, errMsg)
}

// logError logs the error with appropriate context
func logError(config *ErrorConfig, ctx context.Context, method string, err error) {
	fields := []interface{}{
		"method", method,
		"error", err.Error(),
		"error_type", fmt.Sprintf("%T", err),
	}

	// Add request ID if available
	if requestID := GetRequestIDFromContext(ctx); requestID != "" {
		fields = append(fields, "request_id", requestID)
	}

	// Add user ID if available
	if userID, ok := GetUserIDFromContext(ctx); ok {
		fields = append(fields, "user_id", userID)
	}

	// Add stack trace for internal errors if configured
	if config.IncludeStackTrace && isInternalError(err) {
		fields = append(fields, "stack_trace", string(debug.Stack()))
	}

	// Add service error details if available
	if serviceErr, ok := err.(*ServiceError); ok {
		fields = append(fields, "error_code", serviceErr.Code)
		if serviceErr.Details != nil {
			fields = append(fields, "error_details", serviceErr.Details)
		}
		if serviceErr.Cause != nil {
			fields = append(fields, "cause", serviceErr.Cause.Error())
		}
	}

	config.Logger.ErrorContext(ctx, "gRPC error", fields...)
}

// sanitizeGRPCError sanitizes gRPC status errors for production
func sanitizeGRPCError(st *status.Status) error {
	switch st.Code() {
	case codes.Internal, codes.Unknown:
		return status.Error(st.Code(), "internal server error")
	case codes.Unavailable:
		return status.Error(st.Code(), "service temporarily unavailable")
	default:
		return status.Error(st.Code(), st.Message())
	}
}

// sanitizeErrorMessage sanitizes error messages for production
func sanitizeErrorMessage(message string) string {
	// Remove sensitive information patterns using regex
	// Order matters: more specific patterns first
	sensitivePatterns := []string{
		`api_key=[\w\-]+`,
		`password=[\w\-]+`,
		`token=[\w\-]+`,
		`secret=[\w\-]+`,
		`key=[\w\-]+`,
	}

	sanitized := message
	for _, pattern := range sensitivePatterns {
		re := regexp.MustCompile(pattern)
		sanitized = re.ReplaceAllString(sanitized, "[REDACTED]")
	}

	// Generic sanitization for internal errors
	if strings.Contains(strings.ToLower(sanitized), "internal") {
		return "internal server error"
	}

	return sanitized
}

// isInternalError checks if an error is an internal error
func isInternalError(err error) bool {
	if st, ok := status.FromError(err); ok {
		return st.Code() == codes.Internal
	}
	return strings.Contains(strings.ToLower(err.Error()), "internal")
}

// isCriticalError checks if an error is critical and requires alerting
func isCriticalError(err error) bool {
	if st, ok := status.FromError(err); ok {
		switch st.Code() {
		case codes.Internal, codes.DataLoss, codes.Unavailable:
			return true
		}
	}

	// Check for specific error patterns
	errStr := strings.ToLower(err.Error())
	criticalPatterns := []string{
		"database connection failed",
		"out of memory",
		"disk full",
		"service unavailable",
		"panic",
	}

	for _, pattern := range criticalPatterns {
		if strings.Contains(errStr, pattern) {
			return true
		}
	}

	return false
}

// recordErrorMetrics records error metrics (placeholder implementation)
func recordErrorMetrics(err error, method string) {
	// This would integrate with your metrics system (Prometheus, etc.)
	// For now, just log the metric
	code := codes.Internal
	if st, ok := status.FromError(err); ok {
		code = st.Code()
	}

	slog.Debug("error metric recorded",
		"method", method,
		"error_code", code.String(),
		"metric_type", "grpc_error_total",
	)
}

// sendErrorAlert sends an alert for critical errors (placeholder implementation)
func sendErrorAlert(ctx context.Context, method string, err error) {
	// This would integrate with your alerting system (PagerDuty, Slack, etc.)
	slog.Error("critical error alert",
		"method", method,
		"error", err.Error(),
		"alert_type", "critical_error",
	)
}

// sendPanicAlert sends an alert for panics (placeholder implementation)
func sendPanicAlert(ctx context.Context, method string, panicValue interface{}, stack []byte) {
	// This would integrate with your alerting system
	slog.Error("panic alert",
		"method", method,
		"panic", panicValue,
		"stack_trace", string(stack),
		"alert_type", "panic",
	)
}

// Helper functions for creating service errors

// NewServiceError creates a new service error
func NewServiceError(code, message string) *ServiceError {
	return &ServiceError{
		Code:    code,
		Message: message,
		Details: make(map[string]interface{}),
	}
}

// NewServiceErrorWithCause creates a new service error with a cause
func NewServiceErrorWithCause(code, message string, cause error) *ServiceError {
	return &ServiceError{
		Code:    code,
		Message: message,
		Details: make(map[string]interface{}),
		Cause:   cause,
	}
}

// WithDetail adds a detail to the service error
func (e *ServiceError) WithDetail(key string, value interface{}) *ServiceError {
	e.Details[key] = value
	return e
}

// Common service error constructors

// NewValidationError creates a validation error
func NewValidationError(message string) *ServiceError {
	return NewServiceError("VALIDATION_ERROR", message)
}

// NewNotFoundError creates a not found error
func NewNotFoundError(resource string) *ServiceError {
	return NewServiceError("NOT_FOUND", fmt.Sprintf("%s not found", resource))
}

// NewAlreadyExistsError creates an already exists error
func NewAlreadyExistsError(resource string) *ServiceError {
	return NewServiceError("ALREADY_EXISTS", fmt.Sprintf("%s already exists", resource))
}

// NewInternalError creates an internal error
func NewInternalError(message string) *ServiceError {
	return NewServiceError("INTERNAL", message)
}

// NewDatabaseError creates a database error
func NewDatabaseError(operation string, cause error) *ServiceError {
	return NewServiceErrorWithCause("DATABASE_ERROR",
		fmt.Sprintf("database %s failed", operation), cause)
}
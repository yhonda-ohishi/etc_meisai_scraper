package handlers

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/yhonda-ohishi/etc_meisai/src/handlers"
)

// TestNewGRPCErrorHandler tests error handler creation
func TestNewGRPCErrorHandler(t *testing.T) {
	handler := handlers.NewGRPCErrorHandler()
	assert.NotNil(t, handler)
}

// TestHandleGRPCError tests gRPC error handling
func TestHandleGRPCError(t *testing.T) {
	handler := handlers.NewGRPCErrorHandler()

	tests := []struct {
		name           string
		err            error
		expectedStatus int
		expectedCode   string
		expectedMsg    string
	}{
		{
			name:           "nil error",
			err:            nil,
			expectedStatus: 200,
			expectedCode:   "success",
			expectedMsg:    "Operation completed successfully",
		},
		{
			name:           "context deadline exceeded",
			err:            context.DeadlineExceeded,
			expectedStatus: 408,
			expectedCode:   "timeout",
			expectedMsg:    "Request timeout exceeded",
		},
		{
			name:           "context canceled",
			err:            context.Canceled,
			expectedStatus: 408,
			expectedCode:   "canceled",
			expectedMsg:    "Request was canceled",
		},
		{
			name:           "generic error",
			err:            errors.New("some generic error"),
			expectedStatus: 500,
			expectedCode:   "internal_error",
			expectedMsg:    "some generic error",
		},
		{
			name:           "grpc invalid argument",
			err:            status.Error(codes.InvalidArgument, "invalid input"),
			expectedStatus: 400,
			expectedCode:   "invalid_argument",
			expectedMsg:    "invalid input",
		},
		{
			name:           "grpc not found",
			err:            status.Error(codes.NotFound, "resource not found"),
			expectedStatus: 404,
			expectedCode:   "not_found",
			expectedMsg:    "resource not found",
		},
		{
			name:           "grpc already exists",
			err:            status.Error(codes.AlreadyExists, "resource already exists"),
			expectedStatus: 409,
			expectedCode:   "already_exists",
			expectedMsg:    "resource already exists",
		},
		{
			name:           "grpc permission denied",
			err:            status.Error(codes.PermissionDenied, "access denied"),
			expectedStatus: 403,
			expectedCode:   "permission_denied",
			expectedMsg:    "access denied",
		},
		{
			name:           "grpc unauthenticated",
			err:            status.Error(codes.Unauthenticated, "authentication required"),
			expectedStatus: 401,
			expectedCode:   "unauthenticated",
			expectedMsg:    "authentication required",
		},
		{
			name:           "grpc resource exhausted",
			err:            status.Error(codes.ResourceExhausted, "quota exceeded"),
			expectedStatus: 429,
			expectedCode:   "resource_exhausted",
			expectedMsg:    "quota exceeded",
		},
		{
			name:           "grpc failed precondition",
			err:            status.Error(codes.FailedPrecondition, "precondition failed"),
			expectedStatus: 412,
			expectedCode:   "failed_precondition",
			expectedMsg:    "precondition failed",
		},
		{
			name:           "grpc aborted",
			err:            status.Error(codes.Aborted, "operation aborted"),
			expectedStatus: 409,
			expectedCode:   "aborted",
			expectedMsg:    "operation aborted",
		},
		{
			name:           "grpc out of range",
			err:            status.Error(codes.OutOfRange, "parameter out of range"),
			expectedStatus: 400,
			expectedCode:   "out_of_range",
			expectedMsg:    "parameter out of range",
		},
		{
			name:           "grpc unimplemented",
			err:            status.Error(codes.Unimplemented, "not implemented"),
			expectedStatus: 501,
			expectedCode:   "unimplemented",
			expectedMsg:    "not implemented",
		},
		{
			name:           "grpc internal",
			err:            status.Error(codes.Internal, "internal server error"),
			expectedStatus: 500,
			expectedCode:   "internal_error",
			expectedMsg:    "internal server error",
		},
		{
			name:           "grpc unavailable",
			err:            status.Error(codes.Unavailable, "service unavailable"),
			expectedStatus: 503,
			expectedCode:   "unavailable",
			expectedMsg:    "service unavailable",
		},
		{
			name:           "grpc data loss",
			err:            status.Error(codes.DataLoss, "data corruption detected"),
			expectedStatus: 500,
			expectedCode:   "data_loss",
			expectedMsg:    "data corruption detected",
		},
		{
			name:           "grpc deadline exceeded",
			err:            status.Error(codes.DeadlineExceeded, "deadline exceeded"),
			expectedStatus: 408,
			expectedCode:   "deadline_exceeded",
			expectedMsg:    "deadline exceeded",
		},
		{
			name:           "grpc canceled",
			err:            status.Error(codes.Canceled, "request canceled"),
			expectedStatus: 408,
			expectedCode:   "canceled",
			expectedMsg:    "request canceled",
		},
		{
			name:           "grpc ok",
			err:            status.Error(codes.OK, "success"),
			expectedStatus: 200,
			expectedCode:   "success",
			expectedMsg:    "Operation completed successfully",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			httpStatus, code, message := handler.HandleGRPCError(tt.err)

			assert.Equal(t, tt.expectedStatus, httpStatus)
			assert.Equal(t, tt.expectedCode, code)
			assert.Equal(t, tt.expectedMsg, message)
		})
	}
}

// TestCreateErrorDetail tests error detail creation
func TestCreateErrorDetail(t *testing.T) {
	handler := handlers.NewGRPCErrorHandler()

	tests := []struct {
		name      string
		err       error
		requestID string
	}{
		{
			name:      "generic error",
			err:       errors.New("test error"),
			requestID: "req-123",
		},
		{
			name:      "grpc error with details",
			err:       status.Error(codes.InvalidArgument, "invalid input"),
			requestID: "req-456",
		},
		{
			name:      "nil error",
			err:       nil,
			requestID: "req-789",
		},
		{
			name:      "context error",
			err:       context.DeadlineExceeded,
			requestID: "req-abc",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			detail := handler.CreateErrorDetail(tt.err, tt.requestID)

			assert.NotNil(t, detail)
			assert.Equal(t, tt.requestID, detail.RequestID)
			assert.NotEmpty(t, detail.Code)
			assert.NotEmpty(t, detail.Message)
			assert.NotEmpty(t, detail.Timestamp)

			// Check gRPC specific fields for gRPC errors
			if st, ok := status.FromError(tt.err); ok {
				assert.Equal(t, st.Code().String(), detail.GRPCCode)
			}
		})
	}
}

// TestHandleValidationErrors tests validation error handling
func TestHandleValidationErrors(t *testing.T) {
	handler := handlers.NewGRPCErrorHandler()

	validationErrors := []handlers.ValidationError{
		{Field: "email", Message: "invalid email format", Value: "invalid-email"},
		{Field: "age", Message: "must be greater than 0", Value: "-5"},
	}

	httpStatus, code, details := handler.HandleValidationErrors(validationErrors)

	assert.Equal(t, 400, httpStatus)
	assert.Equal(t, "validation_failed", code)
	assert.NotNil(t, details)

	detailsMap, ok := details.(map[string]interface{})
	assert.True(t, ok)
	assert.Contains(t, detailsMap, "validation_errors")
	assert.Contains(t, detailsMap, "message")

	errors, ok := detailsMap["validation_errors"].([]handlers.ValidationError)
	assert.True(t, ok)
	assert.Len(t, errors, 2)
}

// TestHandleBusinessLogicError tests business logic error handling
func TestHandleBusinessLogicError(t *testing.T) {
	handler := handlers.NewGRPCErrorHandler()

	businessError := handlers.BusinessLogicError{
		Code:    "INSUFFICIENT_BALANCE",
		Message: "Account balance is insufficient for this transaction",
		Context: "account_id: 12345, requested_amount: 1000",
	}

	httpStatus, code, details := handler.HandleBusinessLogicError(businessError)

	assert.Equal(t, 422, httpStatus)
	assert.Equal(t, "INSUFFICIENT_BALANCE", code)
	assert.NotNil(t, details)

	detailsMap, ok := details.(map[string]interface{})
	assert.True(t, ok)
	assert.Contains(t, detailsMap, "business_error")
	assert.Contains(t, detailsMap, "message")
	assert.Equal(t, businessError.Message, detailsMap["message"])
}

// TestErrorStructures tests error structure definitions
func TestErrorStructures(t *testing.T) {
	t.Run("ValidationError structure", func(t *testing.T) {
		validationError := handlers.ValidationError{
			Field:   "username",
			Message: "Username must be at least 3 characters",
			Value:   "ab",
		}

		assert.Equal(t, "username", validationError.Field)
		assert.Equal(t, "Username must be at least 3 characters", validationError.Message)
		assert.Equal(t, "ab", validationError.Value)
	})

	t.Run("BusinessLogicError structure", func(t *testing.T) {
		businessError := handlers.BusinessLogicError{
			Code:    "DUPLICATE_ENTRY",
			Message: "An entry with this identifier already exists",
			Context: "table: users, field: email, value: test@example.com",
		}

		assert.Equal(t, "DUPLICATE_ENTRY", businessError.Code)
		assert.Equal(t, "An entry with this identifier already exists", businessError.Message)
		assert.Equal(t, "table: users, field: email, value: test@example.com", businessError.Context)
	})

	t.Run("ErrorDetail structure", func(t *testing.T) {
		errorDetail := &handlers.ErrorDetail{
			Code:      "TEST_ERROR",
			Message:   "This is a test error",
			Details:   map[string]string{"key": "value"},
			GRPCCode:  "INVALID_ARGUMENT",
			Timestamp: "1234567890",
			RequestID: "req-test-123",
		}

		assert.Equal(t, "TEST_ERROR", errorDetail.Code)
		assert.Equal(t, "This is a test error", errorDetail.Message)
		assert.NotNil(t, errorDetail.Details)
		assert.Equal(t, "INVALID_ARGUMENT", errorDetail.GRPCCode)
		assert.Equal(t, "1234567890", errorDetail.Timestamp)
		assert.Equal(t, "req-test-123", errorDetail.RequestID)
	})
}

// TestErrorHandlerEdgeCases tests edge cases and boundary conditions
func TestErrorHandlerEdgeCases(t *testing.T) {
	handler := handlers.NewGRPCErrorHandler()

	t.Run("unknown grpc error code", func(t *testing.T) {
		// Create a status with an unknown code (use a high number)
		unknownStatus := status.New(codes.Code(999), "unknown error")
		err := unknownStatus.Err()

		httpStatus, code, message := handler.HandleGRPCError(err)

		assert.Equal(t, 500, httpStatus)
		assert.Equal(t, "unknown_error", code)
		assert.Contains(t, message, "Unknown gRPC error")
	})

	t.Run("empty validation errors", func(t *testing.T) {
		var emptyErrors []handlers.ValidationError

		httpStatus, code, details := handler.HandleValidationErrors(emptyErrors)

		assert.Equal(t, 400, httpStatus)
		assert.Equal(t, "validation_failed", code)
		assert.NotNil(t, details)

		detailsMap, ok := details.(map[string]interface{})
		assert.True(t, ok)
		errors, ok := detailsMap["validation_errors"].([]handlers.ValidationError)
		assert.True(t, ok)
		assert.Len(t, errors, 0)
	})

	t.Run("business error with empty fields", func(t *testing.T) {
		businessError := handlers.BusinessLogicError{}

		httpStatus, code, details := handler.HandleBusinessLogicError(businessError)

		assert.Equal(t, 422, httpStatus)
		assert.Equal(t, "", code)
		assert.NotNil(t, details)
	})

	t.Run("error detail with empty request id", func(t *testing.T) {
		err := errors.New("test error")
		detail := handler.CreateErrorDetail(err, "")

		assert.NotNil(t, detail)
		assert.Equal(t, "", detail.RequestID)
		assert.NotEmpty(t, detail.Code)
		assert.NotEmpty(t, detail.Message)
	})
}
package interceptors_test

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/yhonda-ohishi/etc_meisai/src/interceptors"
)

// Mock logger for testing
type mockLogger struct {
	mock.Mock
}

func (m *mockLogger) Error(msg string, args ...interface{}) {
	m.Called(append([]interface{}{msg}, args...)...)
}

func (m *mockLogger) ErrorContext(ctx context.Context, msg string, args ...interface{}) {
	m.Called(append([]interface{}{ctx, msg}, args...)...)
}

func (m *mockLogger) Log(ctx context.Context, level slog.Level, msg string, args ...interface{}) {
	m.Called(append([]interface{}{ctx, level, msg}, args...)...)
}

// Mock handlers for testing panics
type panicUnaryHandler struct {
	panicValue interface{}
}

func (h *panicUnaryHandler) Handle(ctx context.Context, req interface{}) (interface{}, error) {
	if h.panicValue != nil {
		panic(h.panicValue)
	}
	return "response", nil
}

type panicStreamHandler struct {
	panicValue interface{}
}

func (h *panicStreamHandler) Handle(srv interface{}, stream grpc.ServerStream) error {
	if h.panicValue != nil {
		panic(h.panicValue)
	}
	return nil
}

func TestNewErrorConfig(t *testing.T) {
	tests := []struct {
		name        string
		environment string
		want        *interceptors.ErrorConfig
	}{
		{
			name:        "production environment",
			environment: "production",
			want: &interceptors.ErrorConfig{
				IncludeStackTrace:     false,
				SanitizeErrors:        true,
				EnableMetrics:         true,
				ProductionMode:        true,
				AlertOnCriticalErrors: true,
			},
		},
		{
			name:        "development environment",
			environment: "development",
			want: &interceptors.ErrorConfig{
				IncludeStackTrace:     true,
				SanitizeErrors:        false,
				EnableMetrics:         true,
				ProductionMode:        false,
				AlertOnCriticalErrors: false,
			},
		},
		{
			name:        "no environment set",
			environment: "",
			want: &interceptors.ErrorConfig{
				IncludeStackTrace:     true,
				SanitizeErrors:        false,
				EnableMetrics:         true,
				ProductionMode:        false,
				AlertOnCriticalErrors: false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			originalEnv := os.Getenv("ENVIRONMENT")
			defer os.Setenv("ENVIRONMENT", originalEnv)

			if tt.environment == "" {
				os.Unsetenv("ENVIRONMENT")
			} else {
				os.Setenv("ENVIRONMENT", tt.environment)
			}

			// Execute
			got := interceptors.NewErrorConfig()

			// Assert
			assert.NotNil(t, got.Logger)
			assert.Equal(t, tt.want.IncludeStackTrace, got.IncludeStackTrace)
			assert.Equal(t, tt.want.SanitizeErrors, got.SanitizeErrors)
			assert.Equal(t, tt.want.EnableMetrics, got.EnableMetrics)
			assert.Equal(t, tt.want.ProductionMode, got.ProductionMode)
			assert.Equal(t, tt.want.AlertOnCriticalErrors, got.AlertOnCriticalErrors)
			assert.NotNil(t, got.ErrorCodeMapping)
			assert.Contains(t, got.ErrorCodeMapping, "VALIDATION_ERROR")
			assert.Equal(t, codes.InvalidArgument, got.ErrorCodeMapping["VALIDATION_ERROR"])
		})
	}
}

func TestServiceError(t *testing.T) {
	t.Run("Error method with cause", func(t *testing.T) {
		cause := errors.New("root cause")
		err := &interceptors.ServiceError{
			Code:    "TEST_ERROR",
			Message: "test message",
			Cause:   cause,
		}

		expected := "TEST_ERROR: test message (caused by: root cause)"
		assert.Equal(t, expected, err.Error())
	})

	t.Run("Error method without cause", func(t *testing.T) {
		err := &interceptors.ServiceError{
			Code:    "TEST_ERROR",
			Message: "test message",
		}

		expected := "TEST_ERROR: test message"
		assert.Equal(t, expected, err.Error())
	})

	t.Run("WithDetail method", func(t *testing.T) {
		err := &interceptors.ServiceError{
			Code:    "TEST_ERROR",
			Message: "test message",
			Details: make(map[string]interface{}),
		}

		result := err.WithDetail("key1", "value1").WithDetail("key2", 42)

		assert.Equal(t, err, result) // Should return same instance
		assert.Equal(t, "value1", err.Details["key1"])
		assert.Equal(t, 42, err.Details["key2"])
	})
}

func TestUnaryErrorHandlerInterceptor_Success(t *testing.T) {
	config := &interceptors.ErrorConfig{
		Logger:        slog.Default(),
		EnableMetrics: false,
	}

	handler := &mockUnaryHandler{}
	handler.On("Handle", mock.Anything, mock.Anything).Return("response", nil)

	interceptor := interceptors.UnaryErrorHandlerInterceptor(config)

	info := &grpc.UnaryServerInfo{
		FullMethod: "/test.Service/Method",
	}

	ctx := context.Background()
	req := "request"

	resp, err := interceptor(ctx, req, info, handler.Handle)

	assert.NoError(t, err)
	assert.Equal(t, "response", resp)
	handler.AssertExpectations(t)
}

func TestUnaryErrorHandlerInterceptor_Panic(t *testing.T) {
	tests := []struct {
		name           string
		panicValue     interface{}
		productionMode bool
		expectedMsg    string
	}{
		{
			name:           "string panic in development",
			panicValue:     "panic message",
			productionMode: false,
			expectedMsg:    "panic recovered: panic message",
		},
		{
			name:           "string panic in production",
			panicValue:     "panic message",
			productionMode: true,
			expectedMsg:    "internal server error",
		},
		{
			name:           "error panic in development",
			panicValue:     errors.New("panic error"),
			productionMode: false,
			expectedMsg:    "panic recovered: panic error",
		},
		{
			name:           "nil panic in development",
			panicValue:     nil,
			productionMode: false,
			expectedMsg:    "panic recovered: <nil>",
		},
		{
			name:           "integer panic in development",
			panicValue:     42,
			productionMode: false,
			expectedMsg:    "panic recovered: 42",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &interceptors.ErrorConfig{
				Logger:                slog.Default(),
				ProductionMode:        tt.productionMode,
				AlertOnCriticalErrors: false,
				EnableMetrics:         false,
			}

			handler := &panicUnaryHandler{panicValue: tt.panicValue}

			interceptor := interceptors.UnaryErrorHandlerInterceptor(config)

			info := &grpc.UnaryServerInfo{
				FullMethod: "/test.Service/Method",
			}

			ctx := context.Background()
			req := "request"

			resp, err := interceptor(ctx, req, info, handler.Handle)

			assert.Error(t, err)
			assert.Nil(t, resp)
			st, ok := status.FromError(err)
			assert.True(t, ok)
			assert.Equal(t, codes.Internal, st.Code())
			assert.Equal(t, tt.expectedMsg, st.Message())
		})
	}
}

func TestUnaryErrorHandlerInterceptor_ErrorHandling(t *testing.T) {
	tests := []struct {
		name         string
		inputError   error
		expectedCode codes.Code
		expectedMsg  string
		sanitize     bool
	}{
		{
			name:         "service error with mapping",
			inputError:   interceptors.NewValidationError("invalid input"),
			expectedCode: codes.InvalidArgument,
			expectedMsg:  "invalid input",
		},
		{
			name:         "service error without mapping",
			inputError:   interceptors.NewServiceError("CUSTOM_ERROR", "custom message"),
			expectedCode: codes.Internal,
			expectedMsg:  "custom message",
		},
		{
			name:         "grpc status error",
			inputError:   status.Error(codes.NotFound, "resource not found"),
			expectedCode: codes.NotFound,
			expectedMsg:  "resource not found",
		},
		{
			name:         "grpc internal error with sanitization",
			inputError:   status.Error(codes.Internal, "database connection failed"),
			expectedCode: codes.Internal,
			expectedMsg:  "internal server error",
			sanitize:     true,
		},
		{
			name:         "generic error classified as not found",
			inputError:   errors.New("user not found"),
			expectedCode: codes.NotFound,
			expectedMsg:  "user not found",
		},
		{
			name:         "generic error classified as already exists",
			inputError:   errors.New("record already exists"),
			expectedCode: codes.AlreadyExists,
			expectedMsg:  "record already exists",
		},
		{
			name:         "generic error classified as permission denied",
			inputError:   errors.New("permission denied"),
			expectedCode: codes.PermissionDenied,
			expectedMsg:  "permission denied",
		},
		{
			name:         "generic error classified as unauthenticated",
			inputError:   errors.New("unauthorized access"),
			expectedCode: codes.Unauthenticated,
			expectedMsg:  "unauthorized access",
		},
		{
			name:         "generic error classified as timeout",
			inputError:   errors.New("operation timeout"),
			expectedCode: codes.DeadlineExceeded,
			expectedMsg:  "operation timeout",
		},
		{
			name:         "generic error classified as cancelled",
			inputError:   errors.New("operation cancelled"),
			expectedCode: codes.Canceled,
			expectedMsg:  "operation cancelled",
		},
		{
			name:         "generic error classified as invalid argument",
			inputError:   errors.New("invalid parameter"),
			expectedCode: codes.InvalidArgument,
			expectedMsg:  "invalid parameter",
		},
		{
			name:         "generic error classified as unavailable",
			inputError:   errors.New("service unavailable"),
			expectedCode: codes.Unavailable,
			expectedMsg:  "service unavailable",
		},
		{
			name:         "generic error classified as rate limited",
			inputError:   errors.New("rate limit exceeded"),
			expectedCode: codes.ResourceExhausted,
			expectedMsg:  "rate limit exceeded",
		},
		{
			name:         "generic error classified as unimplemented",
			inputError:   errors.New("feature unimplemented"),
			expectedCode: codes.Unimplemented,
			expectedMsg:  "feature unimplemented",
		},
		{
			name:         "generic internal error",
			inputError:   errors.New("unexpected error"),
			expectedCode: codes.Internal,
			expectedMsg:  "unexpected error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &interceptors.ErrorConfig{
				Logger:         slog.Default(),
				SanitizeErrors: tt.sanitize,
				EnableMetrics:  false,
				ErrorCodeMapping: map[string]codes.Code{
					"VALIDATION_ERROR": codes.InvalidArgument,
				},
			}

			handler := &mockUnaryHandler{}
			handler.On("Handle", mock.Anything, mock.Anything).Return(nil, tt.inputError)

			interceptor := interceptors.UnaryErrorHandlerInterceptor(config)

			info := &grpc.UnaryServerInfo{
				FullMethod: "/test.Service/Method",
			}

			ctx := context.Background()
			req := "request"

			resp, err := interceptor(ctx, req, info, handler.Handle)

			assert.Error(t, err)
			assert.Nil(t, resp)
			st, ok := status.FromError(err)
			assert.True(t, ok)
			assert.Equal(t, tt.expectedCode, st.Code())
			assert.Equal(t, tt.expectedMsg, st.Message())
			handler.AssertExpectations(t)
		})
	}
}

func TestStreamErrorHandlerInterceptor_Success(t *testing.T) {
	config := &interceptors.ErrorConfig{
		Logger:        slog.Default(),
		EnableMetrics: false,
	}

	handler := &mockStreamHandler{}
	handler.On("Handle", mock.Anything, mock.Anything).Return(nil)

	stream := &mockServerStream{}
	stream.On("Context").Return(context.Background())

	interceptor := interceptors.StreamErrorHandlerInterceptor(config)

	info := &grpc.StreamServerInfo{
		FullMethod: "/test.Service/StreamMethod",
	}

	err := interceptor(nil, stream, info, handler.Handle)

	assert.NoError(t, err)
	handler.AssertExpectations(t)
}

func TestStreamErrorHandlerInterceptor_Panic(t *testing.T) {
	config := &interceptors.ErrorConfig{
		Logger:                slog.Default(),
		ProductionMode:        false,
		AlertOnCriticalErrors: false,
		EnableMetrics:         false,
	}

	handler := &panicStreamHandler{panicValue: "stream panic"}

	stream := &mockServerStream{}
	stream.On("Context").Return(context.Background())

	interceptor := interceptors.StreamErrorHandlerInterceptor(config)

	info := &grpc.StreamServerInfo{
		FullMethod: "/test.Service/StreamMethod",
	}

	err := interceptor(nil, stream, info, handler.Handle)

	assert.Error(t, err)
	st, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, codes.Internal, st.Code())
	assert.Equal(t, "panic recovered: stream panic", st.Message())
}

func TestStreamErrorHandlerInterceptor_Error(t *testing.T) {
	config := &interceptors.ErrorConfig{
		Logger:        slog.Default(),
		EnableMetrics: false,
	}

	testError := errors.New("stream error")
	handler := &mockStreamHandler{}
	handler.On("Handle", mock.Anything, mock.Anything).Return(testError)

	stream := &mockServerStream{}
	stream.On("Context").Return(context.Background())

	interceptor := interceptors.StreamErrorHandlerInterceptor(config)

	info := &grpc.StreamServerInfo{
		FullMethod: "/test.Service/StreamMethod",
	}

	err := interceptor(nil, stream, info, handler.Handle)

	assert.Error(t, err)
	st, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, codes.Internal, st.Code())
	handler.AssertExpectations(t)
}

func TestServiceErrorConstructors(t *testing.T) {
	t.Run("NewServiceError", func(t *testing.T) {
		err := interceptors.NewServiceError("TEST_CODE", "test message")
		assert.Equal(t, "TEST_CODE", err.Code)
		assert.Equal(t, "test message", err.Message)
		assert.NotNil(t, err.Details)
		assert.Nil(t, err.Cause)
	})

	t.Run("NewServiceErrorWithCause", func(t *testing.T) {
		cause := errors.New("root cause")
		err := interceptors.NewServiceErrorWithCause("TEST_CODE", "test message", cause)
		assert.Equal(t, "TEST_CODE", err.Code)
		assert.Equal(t, "test message", err.Message)
		assert.NotNil(t, err.Details)
		assert.Equal(t, cause, err.Cause)
	})

	t.Run("NewValidationError", func(t *testing.T) {
		err := interceptors.NewValidationError("validation failed")
		assert.Equal(t, "VALIDATION_ERROR", err.Code)
		assert.Equal(t, "validation failed", err.Message)
	})

	t.Run("NewNotFoundError", func(t *testing.T) {
		err := interceptors.NewNotFoundError("user")
		assert.Equal(t, "NOT_FOUND", err.Code)
		assert.Equal(t, "user not found", err.Message)
	})

	t.Run("NewAlreadyExistsError", func(t *testing.T) {
		err := interceptors.NewAlreadyExistsError("record")
		assert.Equal(t, "ALREADY_EXISTS", err.Code)
		assert.Equal(t, "record already exists", err.Message)
	})

	t.Run("NewInternalError", func(t *testing.T) {
		err := interceptors.NewInternalError("internal failure")
		assert.Equal(t, "INTERNAL", err.Code)
		assert.Equal(t, "internal failure", err.Message)
	})

	t.Run("NewDatabaseError", func(t *testing.T) {
		cause := errors.New("connection failed")
		err := interceptors.NewDatabaseError("insert", cause)
		assert.Equal(t, "DATABASE_ERROR", err.Code)
		assert.Equal(t, "database insert failed", err.Message)
		assert.Equal(t, cause, err.Cause)
	})
}

func TestErrorMessageSanitization(t *testing.T) {
	tests := []struct {
		name           string
		inputMessage   string
		expectedResult string
	}{
		{
			name:           "sanitize api_key",
			inputMessage:   "error with api_key=secret123",
			expectedResult: "error with [REDACTED]",
		},
		{
			name:           "sanitize password",
			inputMessage:   "failed: password=mypassword",
			expectedResult: "failed: [REDACTED]",
		},
		{
			name:           "sanitize token",
			inputMessage:   "auth failed token=abc123",
			expectedResult: "auth failed [REDACTED]",
		},
		{
			name:           "sanitize multiple secrets",
			inputMessage:   "error api_key=key123 password=pass456",
			expectedResult: "error [REDACTED] [REDACTED]",
		},
		{
			name:           "sanitize internal error",
			inputMessage:   "internal database error occurred",
			expectedResult: "internal server error",
		},
		{
			name:           "no sanitization needed",
			inputMessage:   "simple error message",
			expectedResult: "simple error message",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &interceptors.ErrorConfig{
				Logger:         slog.Default(),
				SanitizeErrors: true,
				EnableMetrics:  false,
			}

			// Create a custom service error to test sanitization
			serviceErr := &interceptors.ServiceError{
				Code:    "TEST_ERROR",
				Message: tt.inputMessage,
			}

			handler := &mockUnaryHandler{}
			handler.On("Handle", mock.Anything, mock.Anything).Return(nil, serviceErr)

			interceptor := interceptors.UnaryErrorHandlerInterceptor(config)

			info := &grpc.UnaryServerInfo{
				FullMethod: "/test.Service/Method",
			}

			ctx := context.Background()
			req := "request"

			_, err := interceptor(ctx, req, info, handler.Handle)

			require.Error(t, err)
			st, ok := status.FromError(err)
			require.True(t, ok)
			assert.Equal(t, tt.expectedResult, st.Message())
		})
	}
}

func TestCriticalErrorDetection(t *testing.T) {
	tests := []struct {
		name        string
		inputError  error
		expectAlert bool
	}{
		{
			name:        "internal grpc error",
			inputError:  status.Error(codes.Internal, "internal error"),
			expectAlert: true,
		},
		{
			name:        "data loss grpc error",
			inputError:  status.Error(codes.DataLoss, "data corrupted"),
			expectAlert: true,
		},
		{
			name:        "unavailable grpc error",
			inputError:  status.Error(codes.Unavailable, "service down"),
			expectAlert: true,
		},
		{
			name:        "database connection error",
			inputError:  errors.New("database connection failed"),
			expectAlert: true,
		},
		{
			name:        "out of memory error",
			inputError:  errors.New("out of memory"),
			expectAlert: true,
		},
		{
			name:        "disk full error",
			inputError:  errors.New("disk full"),
			expectAlert: true,
		},
		{
			name:        "panic error",
			inputError:  errors.New("panic occurred"),
			expectAlert: true,
		},
		{
			name:        "validation error",
			inputError:  status.Error(codes.InvalidArgument, "validation failed"),
			expectAlert: false,
		},
		{
			name:        "not found error",
			inputError:  status.Error(codes.NotFound, "resource not found"),
			expectAlert: false,
		},
		{
			name:        "generic error",
			inputError:  errors.New("simple error"),
			expectAlert: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &interceptors.ErrorConfig{
				Logger:                slog.Default(),
				AlertOnCriticalErrors: true,
				EnableMetrics:         false,
			}

			handler := &mockUnaryHandler{}
			handler.On("Handle", mock.Anything, mock.Anything).Return(nil, tt.inputError)

			interceptor := interceptors.UnaryErrorHandlerInterceptor(config)

			info := &grpc.UnaryServerInfo{
				FullMethod: "/test.Service/Method",
			}

			ctx := context.Background()
			req := "request"

			_, err := interceptor(ctx, req, info, handler.Handle)

			assert.Error(t, err)
			// For this test, we're mainly checking that the interceptor processes
			// critical errors without panicking. The actual alerting logic is
			// a placeholder implementation.
		})
	}
}

func TestContextInjection(t *testing.T) {
	config := &interceptors.ErrorConfig{
		Logger:        slog.Default(),
		EnableMetrics: false,
	}

	handler := &mockUnaryHandler{}
	handler.On("Handle", mock.Anything, mock.Anything).Return(nil, errors.New("test error")).Run(func(args mock.Arguments) {
		ctx := args.Get(0).(context.Context)

		// Test various context values
		if requestID := ctx.Value("request_id"); requestID != nil {
			assert.IsType(t, "", requestID)
		}

		if userID := ctx.Value("user_id"); userID != nil {
			assert.IsType(t, "", userID)
		}
	})

	interceptor := interceptors.UnaryErrorHandlerInterceptor(config)

	info := &grpc.UnaryServerInfo{
		FullMethod: "/test.Service/Method",
	}

	// Create context with user information
	ctx := context.WithValue(context.Background(), "request_id", "req-123")
	ctx = context.WithValue(ctx, "user_id", "user-456")
	req := "request"

	_, err := interceptor(ctx, req, info, handler.Handle)

	assert.Error(t, err)
	handler.AssertExpectations(t)
}

// Benchmark tests
func BenchmarkUnaryErrorHandlerInterceptor_Success(b *testing.B) {
	config := &interceptors.ErrorConfig{
		Logger:        slog.Default(),
		EnableMetrics: false,
	}

	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return "response", nil
	}

	interceptor := interceptors.UnaryErrorHandlerInterceptor(config)

	info := &grpc.UnaryServerInfo{
		FullMethod: "/test.Service/Method",
	}

	ctx := context.Background()
	req := "request"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = interceptor(ctx, req, info, handler)
	}
}

func BenchmarkUnaryErrorHandlerInterceptor_Error(b *testing.B) {
	config := &interceptors.ErrorConfig{
		Logger:         slog.Default(),
		EnableMetrics:  false,
		SanitizeErrors: false,
	}

	testError := errors.New("benchmark error")
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return nil, testError
	}

	interceptor := interceptors.UnaryErrorHandlerInterceptor(config)

	info := &grpc.UnaryServerInfo{
		FullMethod: "/test.Service/Method",
	}

	ctx := context.Background()
	req := "request"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = interceptor(ctx, req, info, handler)
	}
}

func BenchmarkErrorClassification(b *testing.B) {
	testErrors := []error{
		errors.New("user not found"),
		errors.New("record already exists"),
		errors.New("permission denied"),
		errors.New("operation timeout"),
		errors.New("invalid parameter"),
		errors.New("service unavailable"),
		errors.New("unexpected error"),
	}

	config := &interceptors.ErrorConfig{
		Logger:        slog.Default(),
		EnableMetrics: false,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := testErrors[i%len(testErrors)]
		handler := func(ctx context.Context, req interface{}) (interface{}, error) {
			return nil, err
		}

		interceptor := interceptors.UnaryErrorHandlerInterceptor(config)

		info := &grpc.UnaryServerInfo{
			FullMethod: "/test.Service/Method",
		}

		_, _ = interceptor(context.Background(), "request", info, handler)
	}
}

// Concurrent execution tests
func TestUnaryErrorHandlerInterceptor_Concurrent(t *testing.T) {
	config := &interceptors.ErrorConfig{
		Logger:        slog.Default(),
		EnableMetrics: false,
	}

	interceptor := interceptors.UnaryErrorHandlerInterceptor(config)

	info := &grpc.UnaryServerInfo{
		FullMethod: "/test.Service/Method",
	}

	// Test concurrent execution with different error types
	tests := []struct {
		name    string
		handler func(ctx context.Context, req interface{}) (interface{}, error)
	}{
		{
			name: "success",
			handler: func(ctx context.Context, req interface{}) (interface{}, error) {
				return "response", nil
			},
		},
		{
			name: "error",
			handler: func(ctx context.Context, req interface{}) (interface{}, error) {
				return nil, errors.New("test error")
			},
		},
		{
			name: "panic",
			handler: func(ctx context.Context, req interface{}) (interface{}, error) {
				panic("test panic")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			for i := 0; i < 100; i++ {
				go func() {
					_, _ = interceptor(context.Background(), "request", info, tt.handler)
				}()
			}

			// Allow goroutines to complete
			time.Sleep(100 * time.Millisecond)
		})
	}
}
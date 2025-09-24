package interceptors_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/yhonda-ohishi/etc_meisai/src/interceptors"
)

// logCapture captures log calls for testing
type logCapture struct {
	calls []logCall
	mutex sync.Mutex
}

type logCall struct {
	ctx     context.Context
	level   slog.Level
	msg     string
	args    []interface{}
}

func (l *logCapture) getCalls() []logCall {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	return append([]logCall{}, l.calls...)
}

func (l *logCapture) reset() {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	l.calls = nil
}

// createCaptureLogger creates a logger that captures calls
func createCaptureLogger(capture *logCapture) *slog.Logger {
	return slog.New(&captureHandler{capture: capture})
}

// captureHandler implements slog.Handler for testing
type captureHandler struct {
	capture *logCapture
}

func (h *captureHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return true
}

func (h *captureHandler) Handle(ctx context.Context, record slog.Record) error {
	h.capture.mutex.Lock()
	defer h.capture.mutex.Unlock()

	args := make([]interface{}, 0, record.NumAttrs()*2)
	record.Attrs(func(attr slog.Attr) bool {
		args = append(args, attr.Key, attr.Value.Any())
		return true
	})

	h.capture.calls = append(h.capture.calls, logCall{
		ctx:   ctx,
		level: record.Level,
		msg:   record.Message,
		args:  args,
	})
	return nil
}

func (h *captureHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return h
}

func (h *captureHandler) WithGroup(name string) slog.Handler {
	return h
}

// Test request/response structures
type testRequest struct {
	Name        string `json:"name"`
	CardNumber  string `json:"etc_card_number"`
	Password    string `json:"password"`
	Token       string `json:"token"`
}

type testResponse struct {
	ID      string `json:"id"`
	Message string `json:"message"`
}

// Mock stream for testing
type mockLoggingServerStream struct {
	mock.Mock
	grpc.ServerStream
	ctx       context.Context
	sendMsgs  []interface{}
	recvMsgs  []interface{}
	sendError error
	recvError error
}

func (m *mockLoggingServerStream) Context() context.Context {
	if m.ctx != nil {
		return m.ctx
	}
	args := m.Called()
	return args.Get(0).(context.Context)
}

func (m *mockLoggingServerStream) SendMsg(msg interface{}) error {
	m.sendMsgs = append(m.sendMsgs, msg)
	if m.sendError != nil {
		return m.sendError
	}
	args := m.Called(msg)
	return args.Error(0)
}

func (m *mockLoggingServerStream) RecvMsg(msg interface{}) error {
	m.recvMsgs = append(m.recvMsgs, msg)
	if m.recvError != nil {
		return m.recvError
	}
	args := m.Called(msg)
	return args.Error(0)
}

func TestNewLoggingConfig(t *testing.T) {
	config := interceptors.NewLoggingConfig()

	assert.NotNil(t, config.Logger)
	assert.True(t, config.LogPayloads)
	assert.Equal(t, 1024, config.MaxPayloadSize)
	assert.True(t, config.MaskSensitiveData)
	assert.Equal(t, slog.LevelInfo, config.LogLevel)
	assert.Contains(t, config.SensitiveFields, "etc_card_number")
	assert.Contains(t, config.SensitiveFields, "password")
	assert.Contains(t, config.SensitiveFields, "token")
	assert.Contains(t, config.ExcludedMethods, "/grpc.health.v1.Health/Check")
}

func TestUnaryLoggingInterceptor_ExcludedMethods(t *testing.T) {
	capture := &logCapture{}
	config := &interceptors.LoggingConfig{
		Logger:          createCaptureLogger(capture),
		LogPayloads:     true,
		ExcludedMethods: []string{"/grpc.health.v1.Health/Check"},
	}

	handler := &mockUnaryHandler{}
	handler.On("Handle", mock.Anything, mock.Anything).Return("response", nil)

	interceptor := interceptors.UnaryLoggingInterceptor(config)

	info := &grpc.UnaryServerInfo{
		FullMethod: "/grpc.health.v1.Health/Check",
	}

	ctx := context.Background()
	req := "request"

	resp, err := interceptor(ctx, req, info, handler.Handle)

	assert.NoError(t, err)
	assert.Equal(t, "response", resp)
	assert.Empty(t, capture.getCalls()) // No logging should occur for excluded methods
	handler.AssertExpectations(t)
}

func TestUnaryLoggingInterceptor_Success(t *testing.T) {
	capture := &logCapture{}
	config := &interceptors.LoggingConfig{
		Logger:      createCaptureLogger(capture),
		LogPayloads: true,
	}

	handler := &mockUnaryHandler{}
	handler.On("Handle", mock.Anything, mock.Anything).Return("response", nil).Run(func(args mock.Arguments) {
		ctx := args.Get(0).(context.Context)
		// Verify request ID was added to context
		requestID := interceptors.GetRequestIDFromContext(ctx)
		assert.NotEmpty(t, requestID)
	})

	interceptor := interceptors.UnaryLoggingInterceptor(config)

	info := &grpc.UnaryServerInfo{
		FullMethod: "/test.Service/Method",
	}

	ctx := context.Background()
	req := testRequest{
		Name:       "test",
		CardNumber: "1234-5678-9012-3456",
		Password:   "secret123",
	}

	resp, err := interceptor(ctx, req, info, handler.Handle)

	assert.NoError(t, err)
	assert.Equal(t, "response", resp)

	calls := capture.getCalls()
	assert.Len(t, calls, 2) // Request and response logs

	// Verify request log
	requestLog := calls[0]
	assert.Equal(t, slog.LevelInfo, requestLog.level)
	assert.Equal(t, "gRPC request", requestLog.msg)
	assert.Contains(t, requestLog.args, "method")
	assert.Contains(t, requestLog.args, "/test.Service/Method")
	assert.Contains(t, requestLog.args, "request_id")
	assert.Contains(t, requestLog.args, "type")
	assert.Contains(t, requestLog.args, "request")

	// Verify response log
	responseLog := calls[1]
	assert.Equal(t, slog.LevelInfo, responseLog.level)
	assert.Equal(t, "gRPC response", responseLog.msg)
	assert.Contains(t, responseLog.args, "status")
	assert.Contains(t, responseLog.args, "OK")
	assert.Contains(t, responseLog.args, "duration_ms")

	handler.AssertExpectations(t)
}

func TestUnaryLoggingInterceptor_Error(t *testing.T) {
	capture := &logCapture{}
	config := &interceptors.LoggingConfig{
		Logger:      createCaptureLogger(capture),
		LogPayloads: true,
	}

	testError := status.Error(codes.NotFound, "resource not found")
	handler := &mockUnaryHandler{}
	handler.On("Handle", mock.Anything, mock.Anything).Return(nil, testError)

	interceptor := interceptors.UnaryLoggingInterceptor(config)

	info := &grpc.UnaryServerInfo{
		FullMethod: "/test.Service/Method",
	}

	ctx := context.Background()
	req := "request"

	resp, err := interceptor(ctx, req, info, handler.Handle)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, testError, err)

	calls := capture.getCalls()
	assert.Len(t, calls, 2)

	// Verify error response log
	responseLog := calls[1]
	assert.Equal(t, slog.LevelError, responseLog.level)
	assert.Contains(t, responseLog.args, "error")
	assert.Contains(t, responseLog.args, "resource not found")
	assert.Contains(t, responseLog.args, "error_code")
	assert.Contains(t, responseLog.args, "NOT_FOUND")
	assert.Contains(t, responseLog.args, "status")
	assert.Contains(t, responseLog.args, "NotFound")

	handler.AssertExpectations(t)
}

func TestUnaryLoggingInterceptor_WithUserContext(t *testing.T) {
	capture := &logCapture{}
	config := &interceptors.LoggingConfig{
		Logger:      createCaptureLogger(capture),
		LogPayloads: false, // Disable payload logging for this test
	}

	handler := &mockUnaryHandler{}
	handler.On("Handle", mock.Anything, mock.Anything).Return("response", nil)

	interceptor := interceptors.UnaryLoggingInterceptor(config)

	info := &grpc.UnaryServerInfo{
		FullMethod: "/test.Service/Method",
	}

	// Add user context
	ctx := context.WithValue(context.Background(), "user_id", "user123")
	req := "request"

	resp, err := interceptor(ctx, req, info, handler.Handle)

	assert.NoError(t, err)
	assert.Equal(t, "response", resp)

	calls := capture.getCalls()
	assert.Len(t, calls, 2)

	// Both request and response logs should contain user_id
	for _, call := range calls {
		assert.Contains(t, call.args, "user_id")
		assert.Contains(t, call.args, "user123")
	}

	handler.AssertExpectations(t)
}

func TestUnaryLoggingInterceptor_RequestIDFromMetadata(t *testing.T) {
	capture := &logCapture{}
	config := &interceptors.LoggingConfig{
		Logger:      createCaptureLogger(capture),
		LogPayloads: false,
	}

	handler := &mockUnaryHandler{}
	handler.On("Handle", mock.Anything, mock.Anything).Return("response", nil)

	interceptor := interceptors.UnaryLoggingInterceptor(config)

	info := &grpc.UnaryServerInfo{
		FullMethod: "/test.Service/Method",
	}

	// Add request ID to metadata
	expectedRequestID := "custom-request-id-123"
	md := metadata.Pairs("x-request-id", expectedRequestID)
	ctx := metadata.NewIncomingContext(context.Background(), md)
	req := "request"

	resp, err := interceptor(ctx, req, info, handler.Handle)

	assert.NoError(t, err)
	assert.Equal(t, "response", resp)

	calls := capture.getCalls()
	assert.Len(t, calls, 2)

	// Both logs should use the request ID from metadata
	for _, call := range calls {
		assert.Contains(t, call.args, "request_id")
		assert.Contains(t, call.args, expectedRequestID)
	}

	handler.AssertExpectations(t)
}

func TestStreamLoggingInterceptor_Success(t *testing.T) {
	capture := &logCapture{}
	config := &interceptors.LoggingConfig{
		Logger:      createCaptureLogger(capture),
		LogPayloads: true,
	}

	handler := &mockStreamHandler{}
	handler.On("Handle", mock.Anything, mock.Anything).Return(nil)

	stream := &mockLoggingServerStream{}
	stream.On("Context").Return(context.Background())

	interceptor := interceptors.StreamLoggingInterceptor(config)

	info := &grpc.StreamServerInfo{
		FullMethod:      "/test.Service/StreamMethod",
		IsClientStream:  true,
		IsServerStream:  true,
	}

	err := interceptor(nil, stream, info, handler.Handle)

	assert.NoError(t, err)

	calls := capture.getCalls()
	assert.Len(t, calls, 2) // Start and completed logs

	// Verify stream start log
	startLog := calls[0]
	assert.Equal(t, slog.LevelInfo, startLog.level)
	assert.Equal(t, "gRPC stream started", startLog.msg)
	assert.Contains(t, startLog.args, "stream_type")
	assert.Contains(t, startLog.args, "bidirectional")

	// Verify stream completed log
	completedLog := calls[1]
	assert.Equal(t, slog.LevelInfo, completedLog.level)
	assert.Equal(t, "gRPC stream completed", completedLog.msg)
	assert.Contains(t, completedLog.args, "status")
	assert.Contains(t, completedLog.args, "OK")

	handler.AssertExpectations(t)
}

func TestStreamLoggingInterceptor_StreamTypes(t *testing.T) {
	tests := []struct {
		name             string
		isClientStream   bool
		isServerStream   bool
		expectedType     string
	}{
		{
			name:             "bidirectional stream",
			isClientStream:   true,
			isServerStream:   true,
			expectedType:     "bidirectional",
		},
		{
			name:             "client stream",
			isClientStream:   true,
			isServerStream:   false,
			expectedType:     "client_stream",
		},
		{
			name:             "server stream",
			isClientStream:   false,
			isServerStream:   true,
			expectedType:     "server_stream",
		},
		{
			name:             "unknown stream",
			isClientStream:   false,
			isServerStream:   false,
			expectedType:     "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			capture := &logCapture{}
			config := &interceptors.LoggingConfig{
				Logger:      createCaptureLogger(capture),
				LogPayloads: false,
			}

			handler := &mockStreamHandler{}
			handler.On("Handle", mock.Anything, mock.Anything).Return(nil)

			stream := &mockLoggingServerStream{}
			stream.On("Context").Return(context.Background())

			interceptor := interceptors.StreamLoggingInterceptor(config)

			info := &grpc.StreamServerInfo{
				FullMethod:     "/test.Service/StreamMethod",
				IsClientStream: tt.isClientStream,
				IsServerStream: tt.isServerStream,
			}

			err := interceptor(nil, stream, info, handler.Handle)

			assert.NoError(t, err)

			calls := capture.getCalls()
			assert.Len(t, calls, 2)

			startLog := calls[0]
			assert.Contains(t, startLog.args, "stream_type")
			assert.Contains(t, startLog.args, tt.expectedType)

			handler.AssertExpectations(t)
		})
	}
}

func TestStreamLoggingInterceptor_Error(t *testing.T) {
	capture := &logCapture{}
	config := &interceptors.LoggingConfig{
		Logger:      createCaptureLogger(capture),
		LogPayloads: true,
	}

	testError := errors.New("stream error")
	handler := &mockStreamHandler{}
	handler.On("Handle", mock.Anything, mock.Anything).Return(testError)

	stream := &mockLoggingServerStream{}
	stream.On("Context").Return(context.Background())

	interceptor := interceptors.StreamLoggingInterceptor(config)

	info := &grpc.StreamServerInfo{
		FullMethod: "/test.Service/StreamMethod",
	}

	err := interceptor(nil, stream, info, handler.Handle)

	assert.Error(t, err)
	assert.Equal(t, testError, err)

	calls := capture.getCalls()
	assert.Len(t, calls, 2)

	// Verify error log
	errorLog := calls[1]
	assert.Equal(t, slog.LevelError, errorLog.level)
	assert.Contains(t, errorLog.args, "error")
	assert.Contains(t, errorLog.args, "stream error")

	handler.AssertExpectations(t)
}

func TestWrappedServerStream_SendMsg(t *testing.T) {
	capture := &logCapture{}
	config := &interceptors.LoggingConfig{
		Logger:      createCaptureLogger(capture),
		LogPayloads: true,
	}

	handler := &mockStreamHandler{}
	handler.On("Handle", mock.Anything, mock.Anything).Return(nil).Run(func(args mock.Arguments) {
		wrappedStream := args.Get(1).(grpc.ServerStream)

		// Test SendMsg
		err := wrappedStream.SendMsg(testResponse{ID: "123", Message: "test"})
		assert.NoError(t, err)
	})

	stream := &mockLoggingServerStream{}
	stream.On("Context").Return(context.Background())
	stream.On("SendMsg", mock.Anything).Return(nil)

	interceptor := interceptors.StreamLoggingInterceptor(config)

	info := &grpc.StreamServerInfo{
		FullMethod: "/test.Service/StreamMethod",
	}

	err := interceptor(nil, stream, info, handler.Handle)

	assert.NoError(t, err)

	// Verify SendMsg was called on underlying stream
	assert.Len(t, stream.sendMsgs, 1)

	handler.AssertExpectations(t)
	stream.AssertExpectations(t)
}

func TestWrappedServerStream_RecvMsg(t *testing.T) {
	capture := &logCapture{}
	config := &interceptors.LoggingConfig{
		Logger:      createCaptureLogger(capture),
		LogPayloads: true,
	}

	handler := &mockStreamHandler{}
	handler.On("Handle", mock.Anything, mock.Anything).Return(nil).Run(func(args mock.Arguments) {
		wrappedStream := args.Get(1).(grpc.ServerStream)

		// Test RecvMsg
		var msg testRequest
		err := wrappedStream.RecvMsg(&msg)
		assert.NoError(t, err)
	})

	stream := &mockLoggingServerStream{}
	stream.On("Context").Return(context.Background())
	stream.On("RecvMsg", mock.Anything).Return(nil)

	interceptor := interceptors.StreamLoggingInterceptor(config)

	info := &grpc.StreamServerInfo{
		FullMethod: "/test.Service/StreamMethod",
	}

	err := interceptor(nil, stream, info, handler.Handle)

	assert.NoError(t, err)

	// Verify RecvMsg was called on underlying stream
	assert.Len(t, stream.recvMsgs, 1)

	handler.AssertExpectations(t)
	stream.AssertExpectations(t)
}

func TestWrappedServerStream_ErrorHandling(t *testing.T) {
	capture := &logCapture{}
	config := &interceptors.LoggingConfig{
		Logger:      createCaptureLogger(capture),
		LogPayloads: true,
	}

	sendError := errors.New("send error")
	recvError := errors.New("recv error")

	handler := &mockStreamHandler{}
	handler.On("Handle", mock.Anything, mock.Anything).Return(nil).Run(func(args mock.Arguments) {
		wrappedStream := args.Get(1).(grpc.ServerStream)

		// Test SendMsg with error
		err := wrappedStream.SendMsg("test")
		assert.Equal(t, sendError, err)

		// Test RecvMsg with error
		var msg string
		err = wrappedStream.RecvMsg(&msg)
		assert.Equal(t, recvError, err)
	})

	stream := &mockLoggingServerStream{
		sendError: sendError,
		recvError: recvError,
	}
	stream.On("Context").Return(context.Background())

	interceptor := interceptors.StreamLoggingInterceptor(config)

	info := &grpc.StreamServerInfo{
		FullMethod: "/test.Service/StreamMethod",
	}

	err := interceptor(nil, stream, info, handler.Handle)

	assert.NoError(t, err)

	handler.AssertExpectations(t)
}

func TestPayloadFormatting(t *testing.T) {
	tests := []struct {
		name             string
		payload          interface{}
		maxSize          int
		maskSensitive    bool
		sensitiveFields  []string
		expectedContains []string
		expectedNotContains []string
	}{
		{
			name:    "nil payload",
			payload: nil,
			expectedContains: []string{""},
		},
		{
			name: "simple struct",
			payload: testRequest{
				Name: "test",
			},
			expectedContains: []string{"test", "name"},
		},
		{
			name: "protobuf message",
			payload: &emptypb.Empty{},
			expectedContains: []string{"{}"},
		},
		{
			name: "sensitive data masking",
			payload: testRequest{
				Name:       "test",
				CardNumber: "1234-5678-9012-3456",
				Password:   "secret123",
				Token:      "token123",
			},
			maskSensitive:   true,
			sensitiveFields: []string{"etc_card_number", "password", "token"},
			expectedContains: []string{"test", "***MASKED***"},
			expectedNotContains: []string{"secret123", "token123"},
		},
		{
			name: "ETC card number pattern masking",
			payload: map[string]interface{}{
				"card": "1234-5678-9012-3456",
				"name": "test",
			},
			maskSensitive: true,
			expectedContains: []string{"****-****-****-****", "test"},
			expectedNotContains: []string{"1234-5678-9012-3456"},
		},
		{
			name: "payload size truncation",
			payload: map[string]interface{}{
				"large_field": strings.Repeat("a", 2000),
			},
			maxSize: 100,
			expectedContains: []string{"<truncated>"},
		},
		{
			name: "marshal error handling",
			payload: make(chan int), // Channels can't be marshaled to JSON
			expectedContains: []string{"<marshal_error:"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			capture := &logCapture{}
			config := &interceptors.LoggingConfig{
				Logger:            createCaptureLogger(capture),
				MaxPayloadSize:    tt.maxSize,
				MaskSensitiveData: tt.maskSensitive,
				SensitiveFields:   tt.sensitiveFields,
				LogPayloads:       true,
			}
			if config.MaxPayloadSize == 0 {
				config.MaxPayloadSize = 1024
			}

			handler := &mockUnaryHandler{}
			handler.On("Handle", mock.Anything, mock.Anything).Return("response", nil)

			interceptor := interceptors.UnaryLoggingInterceptor(config)

			info := &grpc.UnaryServerInfo{
				FullMethod: "/test.Service/Method",
			}

			ctx := context.Background()

			_, err := interceptor(ctx, tt.payload, info, handler.Handle)

			assert.NoError(t, err)

			calls := capture.getCalls()
			require.Len(t, calls, 2)

			// Find payload in request log
			requestLog := calls[0]
			var payloadStr string
			for i, arg := range requestLog.args {
				if i > 0 && requestLog.args[i-1] == "payload" {
					payloadStr = arg.(string)
					break
				}
			}

			for _, expected := range tt.expectedContains {
				assert.Contains(t, payloadStr, expected)
			}

			for _, notExpected := range tt.expectedNotContains {
				assert.NotContains(t, payloadStr, notExpected)
			}

			handler.AssertExpectations(t)
		})
	}
}

func TestGetRequestIDFromContext(t *testing.T) {
	t.Run("with request ID", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), "request_id", "test-id-123")
		requestID := interceptors.GetRequestIDFromContext(ctx)
		assert.Equal(t, "test-id-123", requestID)
	})

	t.Run("without request ID", func(t *testing.T) {
		ctx := context.Background()
		requestID := interceptors.GetRequestIDFromContext(ctx)
		assert.Empty(t, requestID)
	})

	t.Run("with wrong type", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), "request_id", 123)
		requestID := interceptors.GetRequestIDFromContext(ctx)
		assert.Empty(t, requestID)
	})
}

func TestLogWithRequestID(t *testing.T) {
	capture := &logCapture{}
	logger := createCaptureLogger(capture)

	t.Run("with request ID", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), "request_id", "test-id-123")

		interceptors.LogWithRequestID(ctx, logger, slog.LevelInfo, "test message", "key", "value")

		calls := capture.getCalls()
		assert.Len(t, calls, 1)
		call := calls[0]
		assert.Contains(t, call.args, "request_id")
		assert.Contains(t, call.args, "test-id-123")
		assert.Contains(t, call.args, "key")
		assert.Contains(t, call.args, "value")

		capture.reset()
	})

	t.Run("without request ID", func(t *testing.T) {
		ctx := context.Background()

		interceptors.LogWithRequestID(ctx, logger, slog.LevelInfo, "test message", "key", "value")

		calls := capture.getCalls()
		assert.Len(t, calls, 1)
		call := calls[0]
		assert.NotContains(t, call.args, "request_id")
		assert.Contains(t, call.args, "key")
		assert.Contains(t, call.args, "value")
	})
}

func TestRequestIDGeneration(t *testing.T) {
	t.Run("generates UUID when no metadata", func(t *testing.T) {
		capture := &logCapture{}
		config := &interceptors.LoggingConfig{
			Logger:      createCaptureLogger(capture),
			LogPayloads: false,
		}

		handler := &mockUnaryHandler{}
		handler.On("Handle", mock.Anything, mock.Anything).Return("response", nil)

		interceptor := interceptors.UnaryLoggingInterceptor(config)

		info := &grpc.UnaryServerInfo{
			FullMethod: "/test.Service/Method",
		}

		ctx := context.Background()
		req := "request"

		_, err := interceptor(ctx, req, info, handler.Handle)

		assert.NoError(t, err)

		calls := capture.getCalls()
		assert.Len(t, calls, 2)

		// Extract request ID from first call
		requestLog := calls[0]
		var requestID string
		for i, arg := range requestLog.args {
			if i > 0 && requestLog.args[i-1] == "request_id" {
				requestID = arg.(string)
				break
			}
		}

		// Should be a valid UUID
		_, err = uuid.Parse(requestID)
		assert.NoError(t, err)

		handler.AssertExpectations(t)
	})
}

// Benchmark tests
func BenchmarkUnaryLoggingInterceptor_WithPayloads(b *testing.B) {
	capture := &logCapture{}
	config := &interceptors.LoggingConfig{
		Logger:      createCaptureLogger(capture),
		LogPayloads: true,
	}

	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return "response", nil
	}

	interceptor := interceptors.UnaryLoggingInterceptor(config)

	info := &grpc.UnaryServerInfo{
		FullMethod: "/test.Service/Method",
	}

	ctx := context.Background()
	req := testRequest{
		Name:       "benchmark",
		CardNumber: "1234-5678-9012-3456",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = interceptor(ctx, req, info, handler)
	}
}

func BenchmarkUnaryLoggingInterceptor_WithoutPayloads(b *testing.B) {
	capture := &logCapture{}
	config := &interceptors.LoggingConfig{
		Logger:      createCaptureLogger(capture),
		LogPayloads: false,
	}

	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return "response", nil
	}

	interceptor := interceptors.UnaryLoggingInterceptor(config)

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

func BenchmarkPayloadFormatting(b *testing.B) {
	payload := testRequest{
		Name:       "benchmark",
		CardNumber: "1234-5678-9012-3456",
		Password:   "secret123",
		Token:      "token123",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// We can't directly call formatPayload as it's not exported,
		// but we can measure the effect through JSON marshaling
		json.Marshal(payload)
	}
}

func BenchmarkSensitiveDataMasking(b *testing.B) {
	input := `{"password": "secret123", "token": "abc123", "name": "test", "id": 12345}`
	sensitiveFields := []string{"password", "token", "secret"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// This simulates the masking operation
		result := input
		for _, field := range sensitiveFields {
			pattern := fmt.Sprintf(`"%s"\s*:\s*"[^"]*"`, field)
			result = strings.ReplaceAll(result, pattern, fmt.Sprintf(`"%s": "***MASKED***"`, field))
		}
		_ = result
	}
}

// Concurrent execution tests
func TestLoggingInterceptor_Concurrent(t *testing.T) {
	capture := &logCapture{}
	config := &interceptors.LoggingConfig{
		Logger:      createCaptureLogger(capture),
		LogPayloads: true,
	}

	interceptor := interceptors.UnaryLoggingInterceptor(config)

	info := &grpc.UnaryServerInfo{
		FullMethod: "/test.Service/Method",
	}

	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return "response", nil
	}

	// Test concurrent execution
	for i := 0; i < 10; i++ {
		go func(id int) {
			ctx := context.Background()
			req := fmt.Sprintf("request-%d", id)
			_, _ = interceptor(ctx, req, info, handler)
		}(i)
	}

	// Allow goroutines to complete
	time.Sleep(100 * time.Millisecond)

	// Verify we got logs from all concurrent executions
	calls := capture.getCalls()
	assert.GreaterOrEqual(t, len(calls), 10) // At least 10 calls (some might be response logs)
}
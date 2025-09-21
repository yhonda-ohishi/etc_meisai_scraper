package interceptors

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

// LoggingConfig holds configuration for the logging interceptor
type LoggingConfig struct {
	Logger               *slog.Logger
	LogPayloads          bool
	MaxPayloadSize       int
	SensitiveFields      []string
	MaskSensitiveData    bool
	LogLevel             slog.Level
	ExcludedMethods      []string
}

// NewLoggingConfig creates a new logging configuration
func NewLoggingConfig() *LoggingConfig {
	// Create structured JSON logger
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	return &LoggingConfig{
		Logger:            logger,
		LogPayloads:       true,
		MaxPayloadSize:    1024, // 1KB limit for payload logging
		SensitiveFields:   []string{"etc_card_number", "card_number", "password", "secret", "token"},
		MaskSensitiveData: true,
		LogLevel:          slog.LevelInfo,
		ExcludedMethods: []string{
			"/grpc.health.v1.Health/Check",
			"/grpc.reflection.v1alpha.ServerReflection/ServerReflectionInfo",
		},
	}
}

// UnaryLoggingInterceptor creates a unary server interceptor for logging
func UnaryLoggingInterceptor(config *LoggingConfig) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		// Skip excluded methods
		if isExcludedMethod(info.FullMethod, config.ExcludedMethods) {
			return handler(ctx, req)
		}

		start := time.Now()
		requestID := generateRequestID(ctx)

		// Add request ID to context
		ctx = context.WithValue(ctx, "request_id", requestID)

		// Log request
		logRequest(config, ctx, info.FullMethod, req, requestID)

		// Call handler
		resp, err := handler(ctx, req)

		duration := time.Since(start)
		statusCode := codes.OK
		if err != nil {
			statusCode = status.Code(err)
		}

		// Log response
		logResponse(config, ctx, info.FullMethod, resp, err, duration, statusCode, requestID)

		return resp, err
	}
}

// StreamLoggingInterceptor creates a stream server interceptor for logging
func StreamLoggingInterceptor(config *LoggingConfig) grpc.StreamServerInterceptor {
	return func(
		srv interface{},
		stream grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		// Skip excluded methods
		if isExcludedMethod(info.FullMethod, config.ExcludedMethods) {
			return handler(srv, stream)
		}

		start := time.Now()
		requestID := generateRequestID(stream.Context())

		// Create wrapped stream with request ID
		wrappedStream := &loggingServerStream{
			ServerStream: stream,
			ctx:          context.WithValue(stream.Context(), "request_id", requestID),
			config:       config,
			method:       info.FullMethod,
			requestID:    requestID,
		}

		// Log stream start
		config.Logger.InfoContext(stream.Context(), "gRPC stream started",
			"method", info.FullMethod,
			"request_id", requestID,
			"stream_type", getStreamType(info),
		)

		// Call handler
		err := handler(srv, wrappedStream)

		duration := time.Since(start)
		statusCode := codes.OK
		if err != nil {
			statusCode = status.Code(err)
		}

		// Log stream end
		logLevel := slog.LevelInfo
		if err != nil {
			logLevel = slog.LevelError
		}

		config.Logger.Log(stream.Context(), logLevel, "gRPC stream completed",
			"method", info.FullMethod,
			"request_id", requestID,
			"duration_ms", duration.Milliseconds(),
			"status", statusCode.String(),
			"error", errorToString(err),
		)

		return err
	}
}

// loggingServerStream wraps grpc.ServerStream for logging
type loggingServerStream struct {
	grpc.ServerStream
	ctx       context.Context
	config    *LoggingConfig
	method    string
	requestID string
}

func (s *loggingServerStream) Context() context.Context {
	return s.ctx
}

func (s *loggingServerStream) SendMsg(m interface{}) error {
	err := s.ServerStream.SendMsg(m)

	if s.config.LogPayloads {
		s.config.Logger.DebugContext(s.ctx, "gRPC stream send",
			"method", s.method,
			"request_id", s.requestID,
			"payload", s.formatPayload(m),
			"error", errorToString(err),
		)
	}

	return err
}

func (s *loggingServerStream) RecvMsg(m interface{}) error {
	err := s.ServerStream.RecvMsg(m)

	if s.config.LogPayloads && err == nil {
		s.config.Logger.DebugContext(s.ctx, "gRPC stream receive",
			"method", s.method,
			"request_id", s.requestID,
			"payload", s.formatPayload(m),
		)
	}

	return err
}

func (s *loggingServerStream) formatPayload(payload interface{}) string {
	return formatPayload(s.config, payload)
}

// generateRequestID generates a unique request ID
func generateRequestID(ctx context.Context) string {
	// Try to get request ID from metadata first
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		if reqIDs := md.Get("x-request-id"); len(reqIDs) > 0 {
			return reqIDs[0]
		}
	}

	// Generate new UUID if not found
	return uuid.New().String()
}

// logRequest logs the incoming request
func logRequest(config *LoggingConfig, ctx context.Context, method string, req interface{}, requestID string) {
	fields := []interface{}{
		"method", method,
		"request_id", requestID,
		"type", "request",
	}

	// Add user info if available
	if userID, ok := GetUserIDFromContext(ctx); ok {
		fields = append(fields, "user_id", userID)
	}

	// Add payload if configured
	if config.LogPayloads {
		fields = append(fields, "payload", formatPayload(config, req))
	}

	config.Logger.InfoContext(ctx, "gRPC request", fields...)
}

// logResponse logs the response
func logResponse(config *LoggingConfig, ctx context.Context, method string, resp interface{}, err error, duration time.Duration, statusCode codes.Code, requestID string) {
	logLevel := slog.LevelInfo
	if err != nil {
		logLevel = slog.LevelError
	}

	fields := []interface{}{
		"method", method,
		"request_id", requestID,
		"type", "response",
		"duration_ms", duration.Milliseconds(),
		"status", statusCode.String(),
	}

	// Add user info if available
	if userID, ok := GetUserIDFromContext(ctx); ok {
		fields = append(fields, "user_id", userID)
	}

	// Add error info
	if err != nil {
		fields = append(fields, "error", errorToString(err))
		if st, ok := status.FromError(err); ok {
			fields = append(fields, "error_code", st.Code().String())
			fields = append(fields, "error_message", st.Message())
		}
	}

	// Add response payload if configured and no error
	if config.LogPayloads && err == nil && resp != nil {
		fields = append(fields, "payload", formatPayload(config, resp))
	}

	config.Logger.Log(ctx, logLevel, "gRPC response", fields...)
}

// formatPayload formats the payload for logging
func formatPayload(config *LoggingConfig, payload interface{}) string {
	if payload == nil {
		return ""
	}

	var payloadStr string

	// Handle protobuf messages
	if protoMsg, ok := payload.(proto.Message); ok {
		jsonData, err := protojson.Marshal(protoMsg)
		if err != nil {
			payloadStr = fmt.Sprintf("<marshal_error: %v>", err)
		} else {
			payloadStr = string(jsonData)
		}
	} else {
		// Handle other types
		jsonData, err := json.Marshal(payload)
		if err != nil {
			payloadStr = fmt.Sprintf("<marshal_error: %v>", err)
		} else {
			payloadStr = string(jsonData)
		}
	}

	// Truncate if too large
	if len(payloadStr) > config.MaxPayloadSize {
		payloadStr = payloadStr[:config.MaxPayloadSize] + "...<truncated>"
	}

	// Mask sensitive data
	if config.MaskSensitiveData {
		payloadStr = maskSensitiveData(payloadStr, config.SensitiveFields)
	}

	return payloadStr
}

// maskSensitiveData masks sensitive fields in the payload string
func maskSensitiveData(payload string, sensitiveFields []string) string {
	for _, field := range sensitiveFields {
		// Create regex patterns for different JSON formats
		patterns := []string{
			fmt.Sprintf(`"%s"\s*:\s*"[^"]*"`, field),                    // "field": "value"
			fmt.Sprintf(`"%s"\s*:\s*\d+`, field),                       // "field": 123
			fmt.Sprintf(`%s=[\w\-]+`, field),                           // field=value (query params)
			fmt.Sprintf(`(?i)(%s[\s]*[=:])[\s]*([^\s,}]+)`, field),     // case insensitive with various separators
		}

		for _, pattern := range patterns {
			re := regexp.MustCompile(pattern)
			payload = re.ReplaceAllStringFunc(payload, func(match string) string {
				if strings.Contains(match, ":") {
					parts := strings.Split(match, ":")
					if len(parts) >= 2 {
						return parts[0] + ": \"***MASKED***\""
					}
				} else if strings.Contains(match, "=") {
					parts := strings.Split(match, "=")
					if len(parts) >= 2 {
						return parts[0] + "=***MASKED***"
					}
				}
				return match
			})
		}
	}

	// Special handling for ETC card numbers (pattern matching)
	etcCardPattern := regexp.MustCompile(`\b\d{4}-\d{4}-\d{4}-\d{4}\b`)
	payload = etcCardPattern.ReplaceAllString(payload, "****-****-****-****")

	return payload
}

// isExcludedMethod checks if a method should be excluded from logging
func isExcludedMethod(method string, excludedMethods []string) bool {
	for _, excludedMethod := range excludedMethods {
		if method == excludedMethod {
			return true
		}
	}
	return false
}

// errorToString converts an error to a string, handling nil
func errorToString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

// getStreamType determines the type of stream
func getStreamType(info *grpc.StreamServerInfo) string {
	if info.IsClientStream && info.IsServerStream {
		return "bidirectional"
	} else if info.IsClientStream {
		return "client_stream"
	} else if info.IsServerStream {
		return "server_stream"
	}
	return "unknown"
}

// GetRequestIDFromContext extracts request ID from context
func GetRequestIDFromContext(ctx context.Context) string {
	if requestID, ok := ctx.Value("request_id").(string); ok {
		return requestID
	}
	return ""
}

// LogWithRequestID logs a message with request ID from context
func LogWithRequestID(ctx context.Context, logger *slog.Logger, level slog.Level, msg string, args ...interface{}) {
	if requestID := GetRequestIDFromContext(ctx); requestID != "" {
		args = append([]interface{}{"request_id", requestID}, args...)
	}
	logger.Log(ctx, level, msg, args...)
}
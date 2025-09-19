package services

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// LogLevel represents the severity of a log entry
type LogLevel string

const (
	LogLevelDebug LogLevel = "DEBUG"
	LogLevelInfo  LogLevel = "INFO"
	LogLevelWarn  LogLevel = "WARN"
	LogLevelError LogLevel = "ERROR"
	LogLevelFatal LogLevel = "FATAL"
)

// LogEntry represents a structured log entry
type LogEntry struct {
	Timestamp   time.Time              `json:"timestamp"`
	Level       LogLevel               `json:"level"`
	Message     string                 `json:"message"`
	Service     string                 `json:"service"`
	Method      string                 `json:"method,omitempty"`
	RequestID   string                 `json:"request_id,omitempty"`
	UserID      string                 `json:"user_id,omitempty"`
	Duration    *time.Duration         `json:"duration_ms,omitempty"`
	Error       string                 `json:"error,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// LoggingService provides structured logging
type LoggingService struct {
	logger      *log.Logger
	level       LogLevel
	serviceName string
	logFile     *os.File
	mu          sync.Mutex
	metrics     *LogMetrics
}

// LogMetrics tracks logging metrics
type LogMetrics struct {
	TotalLogs    int64
	ErrorCount   int64
	WarnCount    int64
	InfoCount    int64
	DebugCount   int64
	LastLogTime  time.Time
	mu           sync.RWMutex
}

// NewLoggingService creates a new logging service
func NewLoggingService(serviceName string, logLevel LogLevel) (*LoggingService, error) {
	// Create logs directory
	logDir := "./logs"
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %w", err)
	}

	// Create log file with date
	logFileName := fmt.Sprintf("%s_%s.log", serviceName, time.Now().Format("2006-01-02"))
	logPath := filepath.Join(logDir, logFileName)

	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}

	return &LoggingService{
		logger:      log.New(logFile, "", 0),
		level:       logLevel,
		serviceName: serviceName,
		logFile:     logFile,
		metrics:     &LogMetrics{},
	}, nil
}

// Debug logs a debug message
func (s *LoggingService) Debug(message string, metadata map[string]interface{}) {
	if s.shouldLog(LogLevelDebug) {
		s.log(LogLevelDebug, message, nil, metadata)
		s.updateMetrics(LogLevelDebug)
	}
}

// Info logs an info message
func (s *LoggingService) Info(message string, metadata map[string]interface{}) {
	if s.shouldLog(LogLevelInfo) {
		s.log(LogLevelInfo, message, nil, metadata)
		s.updateMetrics(LogLevelInfo)
	}
}

// Warn logs a warning message
func (s *LoggingService) Warn(message string, metadata map[string]interface{}) {
	if s.shouldLog(LogLevelWarn) {
		s.log(LogLevelWarn, message, nil, metadata)
		s.updateMetrics(LogLevelWarn)
	}
}

// Error logs an error message
func (s *LoggingService) Error(message string, err error, metadata map[string]interface{}) {
	if s.shouldLog(LogLevelError) {
		s.log(LogLevelError, message, err, metadata)
		s.updateMetrics(LogLevelError)
	}
}

// Fatal logs a fatal error and exits
func (s *LoggingService) Fatal(message string, err error, metadata map[string]interface{}) {
	s.log(LogLevelFatal, message, err, metadata)
	s.updateMetrics(LogLevelFatal)
	s.Close()
	os.Exit(1)
}

// LogOperation logs an operation with duration
func (s *LoggingService) LogOperation(operation string, startTime time.Time, err error, metadata map[string]interface{}) {
	duration := time.Since(startTime)

	entry := LogEntry{
		Timestamp:   time.Now(),
		Service:     s.serviceName,
		Method:      operation,
		Duration:    &duration,
		Metadata:    metadata,
	}

	if err != nil {
		entry.Level = LogLevelError
		entry.Message = fmt.Sprintf("Operation failed: %s", operation)
		entry.Error = err.Error()
		s.updateMetrics(LogLevelError)
	} else {
		entry.Level = LogLevelInfo
		entry.Message = fmt.Sprintf("Operation completed: %s", operation)
		s.updateMetrics(LogLevelInfo)
	}

	s.writeEntry(entry)
}

// LogHTTPRequest logs an HTTP request
func (s *LoggingService) LogHTTPRequest(method, path string, statusCode int, duration time.Duration, metadata map[string]interface{}) {
	level := LogLevelInfo
	if statusCode >= 500 {
		level = LogLevelError
	} else if statusCode >= 400 {
		level = LogLevelWarn
	}

	if metadata == nil {
		metadata = make(map[string]interface{})
	}
	metadata["method"] = method
	metadata["path"] = path
	metadata["status_code"] = statusCode

	entry := LogEntry{
		Timestamp: time.Now(),
		Level:     level,
		Service:   s.serviceName,
		Message:   fmt.Sprintf("HTTP %s %s - %d", method, path, statusCode),
		Duration:  &duration,
		Metadata:  metadata,
	}

	s.writeEntry(entry)
	s.updateMetrics(level)
}

// log creates and writes a log entry
func (s *LoggingService) log(level LogLevel, message string, err error, metadata map[string]interface{}) {
	entry := LogEntry{
		Timestamp: time.Now(),
		Level:     level,
		Service:   s.serviceName,
		Message:   message,
		Metadata:  metadata,
	}

	if err != nil {
		entry.Error = err.Error()
	}

	s.writeEntry(entry)
}

// writeEntry writes a log entry to file
func (s *LoggingService) writeEntry(entry LogEntry) {
	s.mu.Lock()
	defer s.mu.Unlock()

	jsonData, err := json.Marshal(entry)
	if err != nil {
		// Fallback to plain text logging
		s.logger.Printf("[%s] %s: %s\n", entry.Level, entry.Service, entry.Message)
		return
	}

	s.logger.Println(string(jsonData))
}

// shouldLog checks if a message should be logged based on level
func (s *LoggingService) shouldLog(level LogLevel) bool {
	levelMap := map[LogLevel]int{
		LogLevelDebug: 0,
		LogLevelInfo:  1,
		LogLevelWarn:  2,
		LogLevelError: 3,
		LogLevelFatal: 4,
	}

	return levelMap[level] >= levelMap[s.level]
}

// updateMetrics updates logging metrics
func (s *LoggingService) updateMetrics(level LogLevel) {
	s.metrics.mu.Lock()
	defer s.metrics.mu.Unlock()

	s.metrics.TotalLogs++
	s.metrics.LastLogTime = time.Now()

	switch level {
	case LogLevelDebug:
		s.metrics.DebugCount++
	case LogLevelInfo:
		s.metrics.InfoCount++
	case LogLevelWarn:
		s.metrics.WarnCount++
	case LogLevelError, LogLevelFatal:
		s.metrics.ErrorCount++
	}
}

// GetMetrics returns logging metrics
func (s *LoggingService) GetMetrics() LogMetrics {
	s.metrics.mu.RLock()
	defer s.metrics.mu.RUnlock()

	return *s.metrics
}

// SetLevel changes the logging level
func (s *LoggingService) SetLevel(level LogLevel) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.level = level
}

// RotateLog rotates the log file
func (s *LoggingService) RotateLog() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Close current file
	if s.logFile != nil {
		s.logFile.Close()
	}

	// Create new log file
	logDir := "./logs"
	logFileName := fmt.Sprintf("%s_%s.log", s.serviceName, time.Now().Format("2006-01-02_15-04-05"))
	logPath := filepath.Join(logDir, logFileName)

	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("failed to create new log file: %w", err)
	}

	s.logFile = logFile
	s.logger = log.New(logFile, "", 0)

	return nil
}

// Close closes the logging service
func (s *LoggingService) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.logFile != nil {
		return s.logFile.Close()
	}
	return nil
}

// CreateRequestLogger creates a logger with request context
func (s *LoggingService) CreateRequestLogger(requestID, userID string) *RequestLogger {
	return &RequestLogger{
		service:   s,
		requestID: requestID,
		userID:    userID,
		startTime: time.Now(),
	}
}

// RequestLogger provides request-scoped logging
type RequestLogger struct {
	service   *LoggingService
	requestID string
	userID    string
	startTime time.Time
}

// Info logs info with request context
func (r *RequestLogger) Info(message string, metadata map[string]interface{}) {
	if metadata == nil {
		metadata = make(map[string]interface{})
	}
	metadata["request_id"] = r.requestID
	metadata["user_id"] = r.userID
	r.service.Info(message, metadata)
}

// Error logs error with request context
func (r *RequestLogger) Error(message string, err error, metadata map[string]interface{}) {
	if metadata == nil {
		metadata = make(map[string]interface{})
	}
	metadata["request_id"] = r.requestID
	metadata["user_id"] = r.userID
	r.service.Error(message, err, metadata)
}

// Complete logs request completion
func (r *RequestLogger) Complete() {
	duration := time.Since(r.startTime)
	metadata := map[string]interface{}{
		"request_id": r.requestID,
		"user_id":    r.userID,
	}
	r.service.LogOperation("request", r.startTime, nil, metadata)
}
package logger

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/johnpr01/home-automation/internal/errors"
	"github.com/johnpr01/home-automation/pkg/kafka"
)

// LogLevel represents the severity of a log message
type LogLevel string

const (
	LogLevelDebug LogLevel = "DEBUG"
	LogLevelInfo  LogLevel = "INFO"
	LogLevelWarn  LogLevel = "WARN"
	LogLevelError LogLevel = "ERROR"
	LogLevelFatal LogLevel = "FATAL"
)

// Logger provides structured logging with error handling integration
type Logger struct {
	serviceName string
	kafkaClient *kafka.Client
	stdLogger   *log.Logger
}

// NewLogger creates a new logger instance
func NewLogger(serviceName string, kafkaClient *kafka.Client) *Logger {
	return &Logger{
		serviceName: serviceName,
		kafkaClient: kafkaClient,
		stdLogger:   log.New(os.Stdout, fmt.Sprintf("[%s] ", serviceName), log.LstdFlags|log.Lshortfile),
	}
}

// LogEntry represents a structured log entry
type LogEntry struct {
	Timestamp time.Time              `json:"timestamp"`
	Level     LogLevel               `json:"level"`
	Service   string                 `json:"service"`
	Message   string                 `json:"message"`
	Error     string                 `json:"error,omitempty"`
	ErrorType string                 `json:"error_type,omitempty"`
	Severity  string                 `json:"severity,omitempty"`
	DeviceID  string                 `json:"device_id,omitempty"`
	RoomID    string                 `json:"room_id,omitempty"`
	File      string                 `json:"file,omitempty"`
	Line      int                    `json:"line,omitempty"`
	Function  string                 `json:"function,omitempty"`
	Context   map[string]interface{} `json:"context,omitempty"`
	TraceID   string                 `json:"trace_id,omitempty"`
}

// Debug logs a debug message
func (l *Logger) Debug(message string, context ...map[string]interface{}) {
	l.log(LogLevelDebug, message, nil, context...)
}

// Info logs an info message
func (l *Logger) Info(message string, context ...map[string]interface{}) {
	l.log(LogLevelInfo, message, nil, context...)
}

// Warn logs a warning message
func (l *Logger) Warn(message string, context ...map[string]interface{}) {
	l.log(LogLevelWarn, message, nil, context...)
}

// Error logs an error message
func (l *Logger) Error(message string, err error, context ...map[string]interface{}) {
	l.log(LogLevelError, message, err, context...)
}

// Fatal logs a fatal error and exits
func (l *Logger) Fatal(message string, err error, context ...map[string]interface{}) {
	l.log(LogLevelFatal, message, err, context...)
	os.Exit(1)
}

// ErrorWithContext logs an error with additional context
func (l *Logger) ErrorWithContext(ctx context.Context, message string, err error, context ...map[string]interface{}) {
	mergedContext := l.extractContextFromCtx(ctx)
	if len(context) > 0 {
		for k, v := range context[0] {
			mergedContext[k] = v
		}
	}
	l.log(LogLevelError, message, err, mergedContext)
}

// LogHomeAutomationError logs a HomeAutomationError with full context
func (l *Logger) LogHomeAutomationError(err *errors.HomeAutomationError) {
	if err == nil {
		return
	}

	entry := &LogEntry{
		Timestamp: time.Now(),
		Level:     LogLevelError,
		Service:   l.serviceName,
		Message:   err.Message,
		Error:     err.Error(),
		ErrorType: string(err.Type),
		Severity:  string(err.Severity),
		DeviceID:  err.DeviceID,
		RoomID:    err.RoomID,
		File:      err.File,
		Line:      err.Line,
		Function:  err.Function,
		Context:   err.Context,
	}

	l.writeLog(entry)

	// Send to Kafka if available and error is significant
	if l.kafkaClient != nil && (err.Severity == errors.SeverityHigh || err.Severity == errors.SeverityCritical) {
		l.sendToKafka(entry)
	}
}

// log is the main logging function
func (l *Logger) log(level LogLevel, message string, err error, context ...map[string]interface{}) {
	entry := &LogEntry{
		Timestamp: time.Now(),
		Level:     level,
		Service:   l.serviceName,
		Message:   message,
		Context:   make(map[string]interface{}),
	}

	// Add error information if present
	if err != nil {
		entry.Error = err.Error()

		// Extract additional info if it's a HomeAutomationError
		if homeErr, ok := err.(*errors.HomeAutomationError); ok {
			entry.ErrorType = string(homeErr.Type)
			entry.Severity = string(homeErr.Severity)
			entry.DeviceID = homeErr.DeviceID
			entry.RoomID = homeErr.RoomID
			entry.File = homeErr.File
			entry.Line = homeErr.Line
			entry.Function = homeErr.Function

			// Merge error context
			for k, v := range homeErr.Context {
				entry.Context[k] = v
			}
		}
	}

	// Merge additional context
	if len(context) > 0 {
		for k, v := range context[0] {
			entry.Context[k] = v
		}
	}

	l.writeLog(entry)

	// Send critical errors to Kafka
	if l.kafkaClient != nil && (level == LogLevelError || level == LogLevelFatal) {
		l.sendToKafka(entry)
	}
}

// writeLog writes the log entry to stdout
func (l *Logger) writeLog(entry *LogEntry) {
	// Write structured JSON for automated processing
	if jsonData, err := json.Marshal(entry); err == nil {
		l.stdLogger.Println(string(jsonData))
	} else {
		// Fallback to simple text logging
		l.stdLogger.Printf("[%s] %s: %s", entry.Level, entry.Message, entry.Error)
	}
}

// sendToKafka sends the log entry to Kafka
func (l *Logger) sendToKafka(entry *LogEntry) {
	if l.kafkaClient == nil {
		return
	}

	kafkaMsg := &kafka.LogMessage{
		Timestamp: entry.Timestamp.Format(time.RFC3339),
		Level:     string(entry.Level),
		Service:   entry.Service,
		Message:   entry.Message,
		DeviceID:  entry.DeviceID,
		Action:    "log",
		Metadata: map[string]interface{}{
			"error_type": entry.ErrorType,
			"severity":   entry.Severity,
			"room_id":    entry.RoomID,
			"file":       entry.File,
			"line":       entry.Line,
			"function":   entry.Function,
			"context":    entry.Context,
		},
	}

	// Send asynchronously to avoid blocking
	go func() {
		if err := l.kafkaClient.PublishLogMessage(kafkaMsg); err != nil {
			// Log to stderr to avoid infinite recursion
			fmt.Fprintf(os.Stderr, "Failed to send log to Kafka: %v\n", err)
		}
	}()
}

// extractContextFromCtx extracts tracing information from context
func (l *Logger) extractContextFromCtx(ctx context.Context) map[string]interface{} {
	context := make(map[string]interface{})

	// Extract trace ID if present
	if traceID := ctx.Value("trace_id"); traceID != nil {
		context["trace_id"] = traceID
	}

	// Extract user ID if present
	if userID := ctx.Value("user_id"); userID != nil {
		context["user_id"] = userID
	}

	// Extract request ID if present
	if requestID := ctx.Value("request_id"); requestID != nil {
		context["request_id"] = requestID
	}

	return context
}

// WithContext creates a new logger with additional context
func (l *Logger) WithContext(context map[string]interface{}) *ContextLogger {
	return &ContextLogger{
		logger:  l,
		context: context,
	}
}

// ContextLogger is a logger with pre-set context
type ContextLogger struct {
	logger  *Logger
	context map[string]interface{}
}

// Debug logs a debug message with context
func (cl *ContextLogger) Debug(message string) {
	cl.logger.Debug(message, cl.context)
}

// Info logs an info message with context
func (cl *ContextLogger) Info(message string) {
	cl.logger.Info(message, cl.context)
}

// Warn logs a warning message with context
func (cl *ContextLogger) Warn(message string) {
	cl.logger.Warn(message, cl.context)
}

// Error logs an error message with context
func (cl *ContextLogger) Error(message string, err error) {
	cl.logger.Error(message, err, cl.context)
}

// Fatal logs a fatal error with context and exits
func (cl *ContextLogger) Fatal(message string, err error) {
	cl.logger.Fatal(message, err, cl.context)
}

// GetGlobalLogger returns a global logger instance
var globalLogger *Logger

// InitGlobalLogger initializes the global logger
func InitGlobalLogger(serviceName string, kafkaClient *kafka.Client) {
	globalLogger = NewLogger(serviceName, kafkaClient)
}

// Debug logs a debug message using the global logger
func Debug(message string, context ...map[string]interface{}) {
	if globalLogger != nil {
		globalLogger.Debug(message, context...)
	}
}

// Info logs an info message using the global logger
func Info(message string, context ...map[string]interface{}) {
	if globalLogger != nil {
		globalLogger.Info(message, context...)
	}
}

// Warn logs a warning message using the global logger
func Warn(message string, context ...map[string]interface{}) {
	if globalLogger != nil {
		globalLogger.Warn(message, context...)
	}
}

// Error logs an error message using the global logger
func Error(message string, err error, context ...map[string]interface{}) {
	if globalLogger != nil {
		globalLogger.Error(message, err, context...)
	}
}

// Fatal logs a fatal error using the global logger and exits
func Fatal(message string, err error, context ...map[string]interface{}) {
	if globalLogger != nil {
		globalLogger.Fatal(message, err, context...)
	}
}

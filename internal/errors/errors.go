package errors

import (
	"fmt"
	"runtime"
	"time"
)

// ErrorType represents the category of error
type ErrorType string

const (
	// Connection errors
	ErrorTypeConnection ErrorType = "CONNECTION"
	ErrorTypeMQTT       ErrorType = "MQTT"
	ErrorTypeKafka      ErrorType = "KAFKA"

	// Service errors
	ErrorTypeService    ErrorType = "SERVICE"
	ErrorTypeValidation ErrorType = "VALIDATION"
	ErrorTypeBusiness   ErrorType = "BUSINESS"

	// Device errors
	ErrorTypeDevice   ErrorType = "DEVICE"
	ErrorTypeSensor   ErrorType = "SENSOR"
	ErrorTypeActuator ErrorType = "ACTUATOR"

	// System errors
	ErrorTypeSystem  ErrorType = "SYSTEM"
	ErrorTypeConfig  ErrorType = "CONFIG"
	ErrorTypeTimeout ErrorType = "TIMEOUT"
)

// Severity represents error severity levels
type Severity string

const (
	SeverityLow      Severity = "LOW"
	SeverityMedium   Severity = "MEDIUM"
	SeverityHigh     Severity = "HIGH"
	SeverityCritical Severity = "CRITICAL"
)

// HomeAutomationError is a custom error type with additional context
type HomeAutomationError struct {
	Type        ErrorType              `json:"type"`
	Severity    Severity               `json:"severity"`
	Message     string                 `json:"message"`
	Cause       error                  `json:"cause,omitempty"`
	Timestamp   time.Time              `json:"timestamp"`
	ServiceName string                 `json:"service_name,omitempty"`
	DeviceID    string                 `json:"device_id,omitempty"`
	RoomID      string                 `json:"room_id,omitempty"`
	File        string                 `json:"file,omitempty"`
	Line        int                    `json:"line,omitempty"`
	Function    string                 `json:"function,omitempty"`
	Context     map[string]interface{} `json:"context,omitempty"`
}

// Error implements the error interface
func (e *HomeAutomationError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("[%s:%s] %s: %v", e.Type, e.Severity, e.Message, e.Cause)
	}
	return fmt.Sprintf("[%s:%s] %s", e.Type, e.Severity, e.Message)
}

// Unwrap returns the underlying cause error for error wrapping
func (e *HomeAutomationError) Unwrap() error {
	return e.Cause
}

// IsCritical returns true if the error severity is critical
func (e *HomeAutomationError) IsCritical() bool {
	return e.Severity == SeverityCritical
}

// IsRetryable returns true if the error might be resolved by retrying
func (e *HomeAutomationError) IsRetryable() bool {
	switch e.Type {
	case ErrorTypeConnection, ErrorTypeMQTT, ErrorTypeKafka, ErrorTypeTimeout:
		return true
	case ErrorTypeValidation, ErrorTypeBusiness, ErrorTypeConfig:
		return false
	default:
		return e.Severity != SeverityCritical
	}
}

// NewError creates a new HomeAutomationError with stack trace information
func NewError(errorType ErrorType, severity Severity, message string, cause error) *HomeAutomationError {
	// Get caller information
	pc, file, line, ok := runtime.Caller(1)
	var function string
	if ok {
		fn := runtime.FuncForPC(pc)
		if fn != nil {
			function = fn.Name()
		}
	}

	return &HomeAutomationError{
		Type:      errorType,
		Severity:  severity,
		Message:   message,
		Cause:     cause,
		Timestamp: time.Now(),
		File:      file,
		Line:      line,
		Function:  function,
		Context:   make(map[string]interface{}),
	}
}

// WithContext adds context information to the error
func (e *HomeAutomationError) WithContext(key string, value interface{}) *HomeAutomationError {
	if e.Context == nil {
		e.Context = make(map[string]interface{})
	}
	e.Context[key] = value
	return e
}

// WithService adds service name to the error
func (e *HomeAutomationError) WithService(serviceName string) *HomeAutomationError {
	e.ServiceName = serviceName
	return e
}

// WithDevice adds device ID to the error
func (e *HomeAutomationError) WithDevice(deviceID string) *HomeAutomationError {
	e.DeviceID = deviceID
	return e
}

// WithRoom adds room ID to the error
func (e *HomeAutomationError) WithRoom(roomID string) *HomeAutomationError {
	e.RoomID = roomID
	return e
}

// Convenience constructors for common error types

// NewConnectionError creates a connection-related error
func NewConnectionError(message string, cause error) *HomeAutomationError {
	return NewError(ErrorTypeConnection, SeverityHigh, message, cause)
}

// NewMQTTError creates an MQTT-related error
func NewMQTTError(message string, cause error) *HomeAutomationError {
	return NewError(ErrorTypeMQTT, SeverityMedium, message, cause)
}

// NewKafkaError creates a Kafka-related error
func NewKafkaError(message string, cause error) *HomeAutomationError {
	return NewError(ErrorTypeKafka, SeverityLow, message, cause)
}

// NewServiceError creates a service-related error
func NewServiceError(message string, cause error) *HomeAutomationError {
	return NewError(ErrorTypeService, SeverityMedium, message, cause)
}

// NewValidationError creates a validation error
func NewValidationError(message string, cause error) *HomeAutomationError {
	return NewError(ErrorTypeValidation, SeverityLow, message, cause)
}

// NewBusinessError creates a business logic error
func NewBusinessError(message string, cause error) *HomeAutomationError {
	return NewError(ErrorTypeBusiness, SeverityMedium, message, cause)
}

// NewDeviceError creates a device-related error
func NewDeviceError(message string, cause error) *HomeAutomationError {
	return NewError(ErrorTypeDevice, SeverityHigh, message, cause)
}

// NewSensorError creates a sensor-related error
func NewSensorError(message string, cause error) *HomeAutomationError {
	return NewError(ErrorTypeSensor, SeverityMedium, message, cause)
}

// NewSystemError creates a system-related error
func NewSystemError(message string, cause error) *HomeAutomationError {
	return NewError(ErrorTypeSystem, SeverityCritical, message, cause)
}

// NewConfigError creates a configuration error
func NewConfigError(message string, cause error) *HomeAutomationError {
	return NewError(ErrorTypeConfig, SeverityCritical, message, cause)
}

// NewTimeoutError creates a timeout error
func NewTimeoutError(message string, cause error) *HomeAutomationError {
	return NewError(ErrorTypeTimeout, SeverityMedium, message, cause)
}

// ErrorHandler provides utilities for error handling
type ErrorHandler struct {
	serviceName string
}

// NewErrorHandler creates a new ErrorHandler for a specific service
func NewErrorHandler(serviceName string) *ErrorHandler {
	return &ErrorHandler{
		serviceName: serviceName,
	}
}

// Handle processes an error and returns an appropriate HomeAutomationError
func (eh *ErrorHandler) Handle(err error) *HomeAutomationError {
	if err == nil {
		return nil
	}

	// If it's already a HomeAutomationError, add service context
	if homeErr, ok := err.(*HomeAutomationError); ok {
		if homeErr.ServiceName == "" {
			homeErr.ServiceName = eh.serviceName
		}
		return homeErr
	}

	// Convert standard error to HomeAutomationError
	return NewServiceError("Service error occurred", err).WithService(eh.serviceName)
}

// WrapError wraps an error with additional context
func (eh *ErrorHandler) WrapError(err error, message string) *HomeAutomationError {
	if err == nil {
		return nil
	}

	return NewServiceError(message, err).WithService(eh.serviceName)
}

// HandleWithContext processes an error and adds context
func (eh *ErrorHandler) HandleWithContext(err error, context map[string]interface{}) *HomeAutomationError {
	if err == nil {
		return nil
	}

	homeErr := eh.Handle(err)
	for k, v := range context {
		homeErr.WithContext(k, v)
	}
	return homeErr
}

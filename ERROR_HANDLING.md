# Comprehensive Error Handling Implementation

## Overview

This document outlines the comprehensive error handling system implemented throughout the home automation repository. The system provides structured error handling, retry mechanisms, circuit breakers, health monitoring, and robust logging.

## 🏗️ **Error Handling Architecture**

### 1. Custom Error Types (`internal/errors/errors.go`)

**Core Features:**
- **Structured Error Types**: Connection, Service, Device, System, Validation errors
- **Severity Levels**: Low, Medium, High, Critical
- **Rich Context**: Stack traces, service names, device IDs, room IDs, timestamps
- **Error Wrapping**: Support for error chains and cause tracking
- **Retry Logic**: Built-in retryability assessment

**Example Usage:**
```go
// Create a device-specific error with context
err := errors.NewDeviceError("sensor offline", originalErr).
    WithDevice("temp-sensor-001").
    WithRoom("living-room").
    WithContext("last_reading", time.Now())

// Check if error is critical
if err.IsCritical() {
    // Handle critical error
}

// Check if error is retryable
if err.IsRetryable() {
    // Attempt retry
}
```

### 2. Retry Mechanisms (`internal/utils/retry.go`)

**Core Features:**
- **Exponential Backoff**: Configurable backoff strategy with jitter
- **Context-Aware**: Respects cancellation and timeouts
- **Type-Based Retries**: Only retry appropriate error types
- **Circuit Breaker**: Prevents cascade failures
- **Health Checks**: Monitor service health

**Example Usage:**
```go
// Retry with default configuration
err := utils.Retry(ctx, nil, func() error {
    return riskyOperation()
})

// Custom retry configuration
config := &utils.RetryConfig{
    MaxAttempts:   5,
    InitialDelay:  200 * time.Millisecond,
    MaxDelay:      30 * time.Second,
    BackoffFactor: 2.0,
    Jitter:        true,
}

err := utils.Retry(ctx, config, operation)
```

### 3. Structured Logging (`internal/logger/logger.go`)

**Core Features:**
- **Structured JSON**: Machine-readable log format
- **Error Integration**: Automatic error context extraction
- **Kafka Integration**: High-severity logs sent to Kafka
- **Context Propagation**: Trace IDs, request IDs, user context
- **Multiple Levels**: Debug, Info, Warn, Error, Fatal

**Example Usage:**
```go
logger := logger.NewLogger("service-name", kafkaClient)

// Simple logging
logger.Info("Operation completed successfully")

// Contextual logging
logger.Error("Database operation failed", err, map[string]interface{}{
    "table": "users",
    "operation": "insert",
    "user_id": 12345,
})

// HomeAutomationError logging (automatic context extraction)
logger.LogHomeAutomationError(homeErr)
```

## 🔧 **Enhanced Client Libraries**

### 1. MQTT Client (`pkg/mqtt/client.go`)

**Error Handling Features:**
- **Connection State Management**: Track connection status
- **Automatic Reconnection**: Circuit breaker with exponential backoff
- **Message Validation**: Input validation with structured errors
- **Health Monitoring**: Built-in health checks
- **Graceful Shutdown**: Proper cleanup and error reporting

**Example Usage:**
```go
options := &mqtt.ClientOptions{
    RetryConfig:    customRetryConfig,
    CircuitBreaker: circuitBreaker,
    Logger:         serviceLogger,
}

client := mqtt.NewClient(config, options)

// Connect with automatic retry
if err := client.Connect(); err != nil {
    // Handle connection failure
}

// Health check
healthStatus := client.GetHealthStatus(ctx)
for check, err := range healthStatus {
    if err != nil {
        logger.Error("Health check failed", err, map[string]interface{}{
            "check": check,
        })
    }
}
```

### 2. Kafka Client (`pkg/kafka/client.go`)

**Error Handling Features:**
- **Async Message Queue**: Non-blocking message publishing
- **Connection Management**: Automatic reconnection and health monitoring
- **Circuit Breaker**: Prevent cascade failures
- **Message Validation**: Structured validation with helpful errors
- **Graceful Degradation**: Continue operating without Kafka if needed

**Example Usage:**
```go
client := kafka.NewClient(brokers, topic, &kafka.ClientOptions{
    QueueSize: 1000,
    RetryConfig: retryConfig,
})

if err := client.Connect(); err != nil {
    // Kafka is optional - log but continue
    logger.Warn("Kafka unavailable", map[string]interface{}{
        "error": err.Error(),
    })
}

// Async message publishing
err := client.PublishLogMessage(&kafka.LogMessage{
    Level:   "ERROR",
    Service: "thermostat",
    Message: "Sensor offline",
})
```

## 🏥 **Health Monitoring**

### Health Checker Implementation

**Features:**
- **Service Health Checks**: Monitor all critical components
- **Timeout Protection**: Prevent hanging health checks
- **Comprehensive Reporting**: Detailed health status
- **Background Monitoring**: Continuous health assessment

**Example Usage:**
```go
healthChecker := utils.NewHealthChecker()

// Register health checks
healthChecker.RegisterCheck("mqtt_connection", func() error {
    return mqttClient.GetHealthStatus(ctx)["mqtt_connection"]
})

healthChecker.RegisterCheck("database", func() error {
    return database.Ping()
})

// Perform health checks
results := healthChecker.CheckHealth(ctx)
for name, err := range results {
    if err != nil {
        logger.Error("Health check failed", err, map[string]interface{}{
            "check": name,
        })
    }
}
```

## 🚦 **Circuit Breaker Pattern**

### Implementation Details

**Features:**
- **State Management**: Closed, Open, Half-Open states
- **Configurable Thresholds**: Failure count and timeout settings
- **Automatic Recovery**: Test and restore functionality
- **Fail-Fast**: Immediate rejection when open

**Example Usage:**
```go
circuitBreaker := utils.NewCircuitBreaker(5, 60*time.Second)

err := circuitBreaker.Execute(func() error {
    return externalServiceCall()
})

if err != nil {
    // Handle circuit breaker or operation error
    switch circuitBreaker.GetState() {
    case utils.CircuitBreakerOpen:
        logger.Warn("Circuit breaker is open")
    case utils.CircuitBreakerHalfOpen:
        logger.Info("Circuit breaker testing recovery")
    }
}
```

## 📊 **Error Metrics and Monitoring**

### Key Metrics to Monitor

1. **Error Rate by Type**
   - Connection errors
   - Validation errors
   - Device errors
   - System errors

2. **Retry Statistics**
   - Retry attempts per operation
   - Success rate after retries
   - Average retry delay

3. **Circuit Breaker Metrics**
   - Open/close transitions
   - Failure thresholds
   - Recovery success rate

4. **Health Check Status**
   - Service availability
   - Response times
   - Failure patterns

## 🎯 **Best Practices**

### 1. Error Creation
```go
// ✅ Good: Specific error with context
return errors.NewDeviceError("temperature sensor timeout", err).
    WithDevice(deviceID).
    WithRoom(roomID).
    WithContext("timeout_duration", "30s")

// ❌ Bad: Generic error without context
return fmt.Errorf("operation failed")
```

### 2. Error Handling
```go
// ✅ Good: Proper error handling with logging
if err := operation(); err != nil {
    logger.LogHomeAutomationError(errorHandler.Handle(err))
    return err
}

// ❌ Bad: Silent error ignoring
operation()
```

### 3. Retry Logic
```go
// ✅ Good: Context-aware retry with appropriate config
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

err := utils.Retry(ctx, retryConfig, operation)

// ❌ Bad: Infinite retry without context
for {
    if err := operation(); err == nil {
        break
    }
    time.Sleep(1 * time.Second)
}
```

### 4. Health Monitoring
```go
// ✅ Good: Comprehensive health checks
healthChecker.RegisterCheck("external_api", func() error {
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    return apiClient.Ping(ctx)
})

// ❌ Bad: No health monitoring
// (missing health checks entirely)
```

## 🔄 **Integration with Existing Services**

### Service Updates Required

1. **Update Constructor Signatures**
   ```go
   // Old
   NewThermostatService(mqttClient *mqtt.Client, logger *log.Logger)
   
   // New
   NewThermostatService(mqttClient *mqtt.Client, logger *logger.Logger)
   ```

2. **Replace Logging Calls**
   ```go
   // Old
   logger.Printf("Operation completed: %s", result)
   
   // New
   logger.Info("Operation completed", map[string]interface{}{
       "result": result,
   })
   ```

3. **Add Error Handling**
   ```go
   // Old
   if err != nil {
       log.Printf("Error: %v", err)
       return err
   }
   
   // New
   if err != nil {
       homeErr := errorHandler.Handle(err)
       logger.LogHomeAutomationError(homeErr)
       return homeErr
   }
   ```

## 🧪 **Testing Error Handling**

### Unit Test Examples

```go
func TestRetryMechanism(t *testing.T) {
    attempts := 0
    operation := func() error {
        attempts++
        if attempts < 3 {
            return errors.NewConnectionError("temporary failure", nil)
        }
        return nil
    }

    err := utils.Retry(context.Background(), nil, operation)
    assert.NoError(t, err)
    assert.Equal(t, 3, attempts)
}

func TestCircuitBreaker(t *testing.T) {
    cb := utils.NewCircuitBreaker(2, time.Second)
    
    // Cause failures to open circuit
    for i := 0; i < 3; i++ {
        cb.Execute(func() error {
            return errors.NewConnectionError("failure", nil)
        })
    }
    
    // Verify circuit is open
    assert.Equal(t, utils.CircuitBreakerOpen, cb.GetState())
}
```

## 📋 **Migration Checklist**

- [x] ✅ Created comprehensive error types and utilities
- [x] ✅ Enhanced MQTT client with error handling
- [x] ✅ Enhanced Kafka client with error handling  
- [x] ✅ Created structured logging system
- [x] ✅ Implemented retry mechanisms
- [x] ✅ Added circuit breaker pattern
- [x] ✅ Created health monitoring system
- [ ] ⏳ Update all service constructors (in progress)
- [ ] ⏳ Replace all logging calls (in progress)
- [ ] ⏳ Add error handling to all operations
- [ ] ⏳ Update unit tests
- [ ] ⏳ Add integration tests for error scenarios
- [ ] ⏳ Update documentation

## 🚀 **Immediate Next Steps**

1. **Service Updates**: Update remaining services to use new logger and error handling
2. **Test Updates**: Modify existing tests to work with new error types
3. **Documentation**: Update README and service documentation
4. **Monitoring**: Set up error metrics collection
5. **Integration Testing**: Add tests for error scenarios and recovery

## 💡 **Benefits Achieved**

1. **🔍 Improved Observability**: Structured errors with full context
2. **🛡️ Enhanced Reliability**: Retry mechanisms and circuit breakers
3. **📊 Better Monitoring**: Health checks and structured logging
4. **🚀 Faster Recovery**: Automatic reconnection and healing
5. **🎯 Easier Debugging**: Rich error context and stack traces
6. **📈 Production Ready**: Comprehensive error handling for production systems

This implementation provides a robust foundation for error handling across the entire home automation system, ensuring reliability, observability, and maintainability in production environments.

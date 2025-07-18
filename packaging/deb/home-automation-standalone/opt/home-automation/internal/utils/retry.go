package utils

import (
	"context"
	"math/rand"
	"sync"
	"time"

	"github.com/johnpr01/home-automation/internal/errors"
)

// RetryConfig configures retry behavior
type RetryConfig struct {
	MaxAttempts    int                `json:"max_attempts"`
	InitialDelay   time.Duration      `json:"initial_delay"`
	MaxDelay       time.Duration      `json:"max_delay"`
	BackoffFactor  float64            `json:"backoff_factor"`
	Jitter         bool               `json:"jitter"`
	RetryableTypes []errors.ErrorType `json:"retryable_types"`
}

// DefaultRetryConfig returns a sensible default retry configuration
func DefaultRetryConfig() *RetryConfig {
	return &RetryConfig{
		MaxAttempts:   3,
		InitialDelay:  100 * time.Millisecond,
		MaxDelay:      30 * time.Second,
		BackoffFactor: 2.0,
		Jitter:        true,
		RetryableTypes: []errors.ErrorType{
			errors.ErrorTypeConnection,
			errors.ErrorTypeMQTT,
			errors.ErrorTypeKafka,
			errors.ErrorTypeTimeout,
		},
	}
}

// RetryableOperation is a function that may need to be retried
type RetryableOperation func() error

// Retry executes an operation with retry logic
func Retry(ctx context.Context, config *RetryConfig, operation RetryableOperation) error {
	if config == nil {
		config = DefaultRetryConfig()
	}

	var lastErr error
	delay := config.InitialDelay

	for attempt := 1; attempt <= config.MaxAttempts; attempt++ {
		// Execute the operation
		err := operation()
		if err == nil {
			return nil // Success
		}

		lastErr = err

		// Check if this is the last attempt
		if attempt == config.MaxAttempts {
			break
		}

		// Check if the error is retryable
		if !isRetryable(err, config.RetryableTypes) {
			break
		}

		// Check context cancellation
		select {
		case <-ctx.Done():
			return errors.NewTimeoutError("operation cancelled during retry", ctx.Err())
		default:
		}

		// Wait before retry
		if delay > 0 {
			select {
			case <-ctx.Done():
				return errors.NewTimeoutError("operation cancelled during retry delay", ctx.Err())
			case <-time.After(delay):
			}
		}

		// Calculate next delay with exponential backoff
		delay = time.Duration(float64(delay) * config.BackoffFactor)
		if delay > config.MaxDelay {
			delay = config.MaxDelay
		}

		// Add jitter if enabled
		if config.Jitter && delay > 0 {
			jitter := time.Duration(float64(delay) * 0.1) // 10% jitter
			delay += time.Duration(float64(jitter) * (2.0*rand.Float64() - 1.0))
		}
	}

	// All retries exhausted, return the last error
	if homeErr, ok := lastErr.(*errors.HomeAutomationError); ok {
		return homeErr.WithContext("retry_attempts", config.MaxAttempts)
	}

	return errors.NewServiceError("operation failed after retries", lastErr).
		WithContext("retry_attempts", config.MaxAttempts)
}

// isRetryable checks if an error should be retried
func isRetryable(err error, retryableTypes []errors.ErrorType) bool {
	homeErr, ok := err.(*errors.HomeAutomationError)
	if !ok {
		// Assume standard errors are retryable for network operations
		return true
	}

	// Check if error is explicitly marked as retryable
	if homeErr.IsRetryable() {
		return true
	}

	// Check if error type is in the retryable types list
	for _, retryableType := range retryableTypes {
		if homeErr.Type == retryableType {
			return true
		}
	}

	return false
}

// CircuitBreakerState represents the state of a circuit breaker
type CircuitBreakerState int

const (
	CircuitBreakerClosed CircuitBreakerState = iota
	CircuitBreakerOpen
	CircuitBreakerHalfOpen
)

// CircuitBreaker implements the circuit breaker pattern for fault tolerance
type CircuitBreaker struct {
	maxFailures     int
	resetTimeout    time.Duration
	state           CircuitBreakerState
	failureCount    int
	lastFailureTime time.Time
	mutex           sync.RWMutex
}

// NewCircuitBreaker creates a new circuit breaker
func NewCircuitBreaker(maxFailures int, resetTimeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		maxFailures:  maxFailures,
		resetTimeout: resetTimeout,
		state:        CircuitBreakerClosed,
	}
}

// Execute runs an operation through the circuit breaker
func (cb *CircuitBreaker) Execute(operation RetryableOperation) error {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	// Check if we should transition from Open to HalfOpen
	if cb.state == CircuitBreakerOpen {
		if time.Since(cb.lastFailureTime) > cb.resetTimeout {
			cb.state = CircuitBreakerHalfOpen
			cb.failureCount = 0
		} else {
			return errors.NewServiceError("circuit breaker is open", nil).
				WithContext("state", "open").
				WithContext("time_until_retry", cb.resetTimeout-time.Since(cb.lastFailureTime))
		}
	}

	// Execute the operation
	err := operation()

	if err != nil {
		cb.onFailure()
		return err
	}

	cb.onSuccess()
	return nil
}

// onFailure handles a failed operation
func (cb *CircuitBreaker) onFailure() {
	cb.failureCount++
	cb.lastFailureTime = time.Now()

	if cb.failureCount >= cb.maxFailures {
		cb.state = CircuitBreakerOpen
	}
}

// onSuccess handles a successful operation
func (cb *CircuitBreaker) onSuccess() {
	cb.failureCount = 0
	cb.state = CircuitBreakerClosed
}

// GetState returns the current state of the circuit breaker
func (cb *CircuitBreaker) GetState() CircuitBreakerState {
	cb.mutex.RLock()
	defer cb.mutex.RUnlock()
	return cb.state
}

// HealthCheck performs a health check operation
type HealthCheck func() error

// HealthChecker manages health checks for services
type HealthChecker struct {
	checks map[string]HealthCheck
	mutex  sync.RWMutex
}

// NewHealthChecker creates a new health checker
func NewHealthChecker() *HealthChecker {
	return &HealthChecker{
		checks: make(map[string]HealthCheck),
	}
}

// RegisterCheck adds a health check
func (hc *HealthChecker) RegisterCheck(name string, check HealthCheck) {
	hc.mutex.Lock()
	defer hc.mutex.Unlock()
	hc.checks[name] = check
}

// CheckHealth performs all registered health checks
func (hc *HealthChecker) CheckHealth(ctx context.Context) map[string]error {
	hc.mutex.RLock()
	defer hc.mutex.RUnlock()

	results := make(map[string]error)

	for name, check := range hc.checks {
		// Run each check with timeout
		checkCtx, cancel := context.WithTimeout(ctx, 5*time.Second)

		checkErr := make(chan error, 1)
		go func() {
			checkErr <- check()
		}()

		select {
		case err := <-checkErr:
			results[name] = err
		case <-checkCtx.Done():
			results[name] = errors.NewTimeoutError("health check timed out", checkCtx.Err())
		}

		cancel()
	}

	return results
}

package kafka

import (
	"context"
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/johnpr01/home-automation/internal/errors"
	"github.com/johnpr01/home-automation/internal/utils"
)

// LogMessage represents a structured log message for Kafka
type LogMessage struct {
	Timestamp string                 `json:"timestamp"`
	Level     string                 `json:"level"`
	Service   string                 `json:"service"`
	Message   string                 `json:"message"`
	DeviceID  string                 `json:"device_id,omitempty"`
	Action    string                 `json:"action,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// ConnectionState represents the Kafka connection state
type ConnectionState int

const (
	StateDisconnected ConnectionState = iota
	StateConnecting
	StateConnected
	StateReconnecting
)

// Client represents a Kafka client for publishing log messages
type Client struct {
	brokers        []string
	topic          string
	state          ConnectionState
	stateMutex     sync.RWMutex
	errorHandler   *errors.ErrorHandler
	retryConfig    *utils.RetryConfig
	circuitBreaker *utils.CircuitBreaker
	healthChecker  *utils.HealthChecker
	ctx            context.Context
	cancel         context.CancelFunc
	messageQueue   chan *LogMessage
	// queueMutex removed as it was unused
}

// ClientOptions provides configuration options for the Kafka client
type ClientOptions struct {
	RetryConfig    *utils.RetryConfig
	CircuitBreaker *utils.CircuitBreaker
	QueueSize      int
}

// NewClient creates a new Kafka client
func NewClient(brokers []string, topic string, options *ClientOptions) *Client {
	if len(brokers) == 0 {
		return nil
	}

	ctx, cancel := context.WithCancel(context.Background())

	// Set defaults if not provided
	var retryConfig *utils.RetryConfig
	var circuitBreaker *utils.CircuitBreaker
	queueSize := 1000

	if options != nil {
		retryConfig = options.RetryConfig
		circuitBreaker = options.CircuitBreaker
		if options.QueueSize > 0 {
			queueSize = options.QueueSize
		}
	}

	if retryConfig == nil {
		retryConfig = utils.DefaultRetryConfig()
		retryConfig.MaxAttempts = 3
		retryConfig.MaxDelay = 10 * time.Second
	}

	if circuitBreaker == nil {
		circuitBreaker = utils.NewCircuitBreaker(5, 60*time.Second)
	}

	client := &Client{
		brokers:        brokers,
		topic:          topic,
		state:          StateDisconnected,
		errorHandler:   errors.NewErrorHandler("kafka-client"),
		retryConfig:    retryConfig,
		circuitBreaker: circuitBreaker,
		healthChecker:  utils.NewHealthChecker(),
		ctx:            ctx,
		cancel:         cancel,
		messageQueue:   make(chan *LogMessage, queueSize),
	}

	// Register health check
	client.healthChecker.RegisterCheck("kafka_connection", client.healthCheck)

	return client
}

// Connect establishes connection to Kafka brokers
func (c *Client) Connect() error {
	if c == nil {
		return errors.NewKafkaError("kafka client is nil", nil)
	}

	log.Printf("Kafka client: Attempting to connect to brokers %v, topic: %s", c.brokers, c.topic)

	operation := func() error {
		c.setState(StateConnecting)

		// Validate configuration
		if len(c.brokers) == 0 {
			return errors.NewKafkaError("no brokers specified", nil)
		}

		if c.topic == "" {
			return errors.NewKafkaError("no topic specified", nil)
		}

		// TODO: Implement actual Kafka connection logic
		// For now, we'll simulate a successful connection
		c.setState(StateConnected)

		log.Printf("Kafka client: Successfully connected to brokers")
		return nil
	}

	err := utils.Retry(c.ctx, c.retryConfig, operation)
	if err != nil {
		c.setState(StateDisconnected)
		return c.errorHandler.WrapError(err, "failed to connect to Kafka brokers")
	}

	// Start background message processor
	go c.processMessageQueue()

	return nil
}

// Disconnect closes the connection to Kafka brokers
func (c *Client) Disconnect() error {
	if c == nil {
		return nil
	}

	log.Printf("Kafka client: Disconnecting from brokers")

	// Cancel background operations
	c.cancel()

	c.setState(StateDisconnected)

	// Process remaining messages in queue
	c.drainMessageQueue()

	// TODO: Implement actual Kafka disconnection logic

	log.Printf("Kafka client: Successfully disconnected from brokers")
	return nil
}

// PublishLogMessage sends a log message to Kafka
func (c *Client) PublishLogMessage(logMsg *LogMessage) error {
	if c == nil {
		return errors.NewKafkaError("kafka client is nil", nil)
	}

	if logMsg == nil {
		return errors.NewValidationError("log message cannot be nil", nil)
	}

	// Validate message
	if logMsg.Service == "" {
		return errors.NewValidationError("log message service cannot be empty", nil)
	}

	if logMsg.Message == "" {
		return errors.NewValidationError("log message cannot be empty", nil)
	}

	// Add to queue for async processing
	select {
	case c.messageQueue <- logMsg:
		return nil
	default:
		// Queue is full, try to publish synchronously
		return c.publishMessageSync(logMsg)
	}
}

// publishMessageSync publishes a message synchronously
func (c *Client) publishMessageSync(logMsg *LogMessage) error {
	if !c.isConnected() {
		return errors.NewKafkaError("client is not connected", nil)
	}

	operation := func() error {
		// Convert log message to JSON
		jsonData, err := json.Marshal(logMsg)
		if err != nil {
			return errors.NewKafkaError("failed to marshal log message", err)
		}

		// TODO: Implement actual Kafka publish logic
		log.Printf("Kafka client: Publishing message to topic %s (service: %s, level: %s, size: %d bytes)",
			c.topic, logMsg.Service, logMsg.Level, len(jsonData))

		return nil
	}

	err := c.circuitBreaker.Execute(operation)
	if err != nil {
		return c.errorHandler.WrapError(err, "failed to publish message to Kafka").
			WithContext("topic", c.topic).
			WithContext("service", logMsg.Service)
	}

	return nil
}

// setState safely updates the connection state
func (c *Client) setState(state ConnectionState) {
	c.stateMutex.Lock()
	defer c.stateMutex.Unlock()
	c.state = state
}

// GetState returns the current connection state
func (c *Client) GetState() ConnectionState {
	c.stateMutex.RLock()
	defer c.stateMutex.RUnlock()
	return c.state
}

// isConnected checks if the client is connected
func (c *Client) isConnected() bool {
	return c.GetState() == StateConnected
}

// healthCheck performs a health check on the Kafka connection
func (c *Client) healthCheck() error {
	if !c.isConnected() {
		return errors.NewKafkaError("Kafka client is not connected", nil)
	}

	// TODO: Implement actual health check (e.g., metadata request)
	return nil
}

// processMessageQueue processes messages from the queue asynchronously
func (c *Client) processMessageQueue() {
	for {
		select {
		case <-c.ctx.Done():
			return
		case msg := <-c.messageQueue:
			if err := c.publishMessageSync(msg); err != nil {
				log.Printf("Kafka client: Failed to publish queued message from service %s: %v", msg.Service, err)
			}
		}
	}
}

// drainMessageQueue processes remaining messages in the queue
func (c *Client) drainMessageQueue() {
	close(c.messageQueue)

	for msg := range c.messageQueue {
		if err := c.publishMessageSync(msg); err != nil {
			log.Printf("Kafka client: Failed to publish message during shutdown from service %s: %v", msg.Service, err)
		}
	}
}

// GetHealthStatus returns the health status of the Kafka client
func (c *Client) GetHealthStatus(ctx context.Context) map[string]error {
	if c == nil {
		return map[string]error{
			"kafka_connection": errors.NewKafkaError("kafka client is nil", nil),
		}
	}
	return c.healthChecker.CheckHealth(ctx)
}

// GetQueueSize returns the current size of the message queue
func (c *Client) GetQueueSize() int {
	if c == nil {
		return 0
	}
	return len(c.messageQueue)
}

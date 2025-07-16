package mqtt

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/johnpr01/home-automation/internal/config"
	"github.com/johnpr01/home-automation/internal/errors"
	"github.com/johnpr01/home-automation/internal/logger"
	"github.com/johnpr01/home-automation/internal/utils"
)

// ConnectionState represents the MQTT connection state
type ConnectionState int

const (
	StateDisconnected ConnectionState = iota
	StateConnecting
	StateConnected
	StateReconnecting
)

type Client struct {
	config         *config.MQTTConfig
	handlers       map[string]MessageHandler
	state          ConnectionState
	stateMutex     sync.RWMutex
	logger         *logger.Logger
	errorHandler   *errors.ErrorHandler
	retryConfig    *utils.RetryConfig
	circuitBreaker *utils.CircuitBreaker
	healthChecker  *utils.HealthChecker
	ctx            context.Context
	cancel         context.CancelFunc
	reconnectChan  chan struct{}
}

type MessageHandler func(topic string, payload []byte) error

type Message struct {
	Topic   string
	Payload []byte
	QoS     byte
	Retain  bool
}

// ClientOptions provides configuration options for the MQTT client
type ClientOptions struct {
	RetryConfig    *utils.RetryConfig
	CircuitBreaker *utils.CircuitBreaker
	Logger         *logger.Logger
}

func NewClient(cfg *config.MQTTConfig, options *ClientOptions) *Client {
	ctx, cancel := context.WithCancel(context.Background())

	// Set defaults if not provided
	var retryConfig *utils.RetryConfig
	var circuitBreaker *utils.CircuitBreaker
	var clientLogger *logger.Logger

	if options != nil {
		retryConfig = options.RetryConfig
		circuitBreaker = options.CircuitBreaker
		clientLogger = options.Logger
	}

	if retryConfig == nil {
		retryConfig = utils.DefaultRetryConfig()
		retryConfig.MaxAttempts = 5
		retryConfig.MaxDelay = 30 * time.Second
	}

	if circuitBreaker == nil {
		circuitBreaker = utils.NewCircuitBreaker(3, 30*time.Second)
	}

	if clientLogger == nil {
		clientLogger = logger.NewLogger("mqtt-client", nil)
	}

	client := &Client{
		config:         cfg,
		handlers:       make(map[string]MessageHandler),
		state:          StateDisconnected,
		logger:         clientLogger,
		errorHandler:   errors.NewErrorHandler("mqtt-client"),
		retryConfig:    retryConfig,
		circuitBreaker: circuitBreaker,
		healthChecker:  utils.NewHealthChecker(),
		ctx:            ctx,
		cancel:         cancel,
		reconnectChan:  make(chan struct{}, 1),
	}

	// Register health check
	client.healthChecker.RegisterCheck("mqtt_connection", client.healthCheck)

	return client
}

func (c *Client) Connect() error {
	c.logger.Info("Attempting to connect to MQTT broker", map[string]interface{}{
		"broker": c.config.Broker,
		"port":   c.config.Port,
	})

	operation := func() error {
		c.setState(StateConnecting)

		// Simulate connection logic - replace with actual MQTT client
		if c.config.Broker == "" {
			return errors.NewMQTTError("broker address is empty", nil)
		}

		if c.config.Port == "" {
			return errors.NewMQTTError("broker port is empty", nil)
		}

		// TODO: Implement actual MQTT connection logic here
		// For now, we'll simulate a successful connection
		c.setState(StateConnected)

		c.logger.Info("Successfully connected to MQTT broker")
		return nil
	}

	err := utils.Retry(c.ctx, c.retryConfig, operation)
	if err != nil {
		c.setState(StateDisconnected)
		return c.errorHandler.WrapError(err, "failed to connect to MQTT broker")
	}

	// Start background reconnection handler
	go c.handleReconnection()

	return nil
}

func (c *Client) Disconnect() error {
	c.logger.Info("Disconnecting from MQTT broker")

	// Cancel background operations
	c.cancel()

	c.setState(StateDisconnected)

	// TODO: Implement actual MQTT disconnection logic

	c.logger.Info("Successfully disconnected from MQTT broker")
	return nil
}

func (c *Client) Subscribe(topic string, handler MessageHandler) error {
	if topic == "" {
		return errors.NewValidationError("topic cannot be empty", nil)
	}

	if handler == nil {
		return errors.NewValidationError("handler cannot be nil", nil)
	}

	if !c.isConnected() {
		return errors.NewMQTTError("client is not connected", nil)
	}

	operation := func() error {
		// TODO: Implement actual MQTT subscription logic
		c.handlers[topic] = handler

		c.logger.Info("Subscribed to MQTT topic", map[string]interface{}{
			"topic": topic,
		})
		return nil
	}

	return c.circuitBreaker.Execute(operation)
}

func (c *Client) Publish(msg *Message) error {
	if msg == nil {
		return errors.NewValidationError("message cannot be nil", nil)
	}

	if msg.Topic == "" {
		return errors.NewValidationError("message topic cannot be empty", nil)
	}

	if !c.isConnected() {
		return errors.NewMQTTError("client is not connected", nil)
	}

	operation := func() error {
		// TODO: Implement actual MQTT publish logic
		c.logger.Debug("Publishing MQTT message", map[string]interface{}{
			"topic":   msg.Topic,
			"qos":     msg.QoS,
			"retain":  msg.Retain,
			"payload": string(msg.Payload),
		})
		return nil
	}

	err := c.circuitBreaker.Execute(operation)
	if err != nil {
		return c.errorHandler.WrapError(err, "failed to publish MQTT message").
			WithContext("topic", msg.Topic).
			WithContext("qos", msg.QoS)
	}

	return nil
}

func (c *Client) PublishDeviceState(deviceID string, state map[string]interface{}) error {
	if deviceID == "" {
		return errors.NewValidationError("deviceID cannot be empty", nil)
	}

	if state == nil {
		return errors.NewValidationError("state cannot be nil", nil)
	}

	topic := fmt.Sprintf("homeautomation/devices/%s/state", deviceID)

	payload, err := json.Marshal(state)
	if err != nil {
		return c.errorHandler.WrapError(err, "failed to marshal device state").
			WithDevice(deviceID)
	}

	msg := &Message{
		Topic:   topic,
		Payload: payload,
		QoS:     1,
		Retain:  true,
	}

	err = c.Publish(msg)
	if err != nil {
		return c.errorHandler.WrapError(err, "failed to publish device state").
			WithDevice(deviceID)
	}

	return nil
}

func (c *Client) PublishSensorReading(sensorID string, reading map[string]interface{}) error {
	if sensorID == "" {
		return errors.NewValidationError("sensorID cannot be empty", nil)
	}

	if reading == nil {
		return errors.NewValidationError("reading cannot be nil", nil)
	}

	topic := fmt.Sprintf("homeautomation/sensors/%s/reading", sensorID)

	payload, err := json.Marshal(reading)
	if err != nil {
		return c.errorHandler.WrapError(err, "failed to marshal sensor reading").
			WithContext("sensor_id", sensorID)
	}

	msg := &Message{
		Topic:   topic,
		Payload: payload,
		QoS:     1,
		Retain:  false,
	}

	err = c.Publish(msg)
	if err != nil {
		return c.errorHandler.WrapError(err, "failed to publish sensor reading").
			WithContext("sensor_id", sensorID)
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

// healthCheck performs a health check on the MQTT connection
func (c *Client) healthCheck() error {
	if !c.isConnected() {
		return errors.NewMQTTError("MQTT client is not connected", nil)
	}

	// TODO: Implement actual health check (e.g., ping the broker)
	return nil
}

// handleReconnection manages automatic reconnection
func (c *Client) handleReconnection() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-c.ctx.Done():
			return
		case <-ticker.C:
			if !c.isConnected() {
				c.logger.Warn("MQTT client disconnected, attempting to reconnect")
				c.reconnect()
			}
		case <-c.reconnectChan:
			if !c.isConnected() {
				c.reconnect()
			}
		}
	}
}

// reconnect attempts to reconnect to the MQTT broker
func (c *Client) reconnect() {
	c.setState(StateReconnecting)

	operation := func() error {
		return c.Connect()
	}

	err := utils.Retry(c.ctx, c.retryConfig, operation)
	if err != nil {
		c.logger.Error("Failed to reconnect to MQTT broker", err)
		c.setState(StateDisconnected)
	}
}

// TriggerReconnect manually triggers a reconnection attempt
func (c *Client) TriggerReconnect() {
	select {
	case c.reconnectChan <- struct{}{}:
	default:
		// Channel full, reconnection already pending
	}
}

// GetHealthStatus returns the health status of the MQTT client
func (c *Client) GetHealthStatus(ctx context.Context) map[string]error {
	return c.healthChecker.CheckHealth(ctx)
}

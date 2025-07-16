package influxdb

import (
	"context"
	"fmt"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
	"github.com/johnpr01/home-automation/internal/errors"
	"github.com/johnpr01/home-automation/internal/utils"
)

// Client represents an InfluxDB client for storing time series data
type Client struct {
	client      influxdb2.Client
	writeAPI    api.WriteAPI
	queryAPI    api.QueryAPI
	org         string
	bucket      string
	state       ConnectionState
	retryConfig *utils.RetryConfig
}

// ConnectionState represents the InfluxDB connection state
type ConnectionState int

const (
	StateDisconnected ConnectionState = iota
	StateConnecting
	StateConnected
)

// NewClient creates a new InfluxDB client
func NewClient(url, token, org, bucket string) *Client {
	if url == "" || token == "" || org == "" || bucket == "" {
		return nil
	}

	retryConfig := utils.DefaultRetryConfig()
	retryConfig.MaxAttempts = 3
	retryConfig.MaxDelay = 5 * time.Second

	influxClient := influxdb2.NewClient(url, token)

	client := &Client{
		client:      influxClient,
		org:         org,
		bucket:      bucket,
		state:       StateDisconnected,
		retryConfig: retryConfig,
	}

	client.writeAPI = client.client.WriteAPI(org, bucket)
	client.queryAPI = client.client.QueryAPI(org)

	return client
}

// Connect establishes connection to InfluxDB
func (c *Client) Connect() error {
	c.state = StateConnecting

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	health, err := c.client.Health(ctx)
	if err != nil {
		c.state = StateDisconnected
		return errors.NewConnectionError("Failed to connect to InfluxDB", err)
	}

	if health.Status != "pass" {
		c.state = StateDisconnected
		return errors.NewConnectionError(fmt.Sprintf("InfluxDB health check failed: %s", health.Status), nil)
	}

	c.state = StateConnected
	return nil
}

// Disconnect closes the connection to InfluxDB
func (c *Client) Disconnect() {
	if c.client != nil {
		c.writeAPI.Flush()
		c.client.Close()
	}
	c.state = StateDisconnected
}

// IsConnected returns true if connected to InfluxDB
func (c *Client) IsConnected() bool {
	return c.state == StateConnected
}

// EnergyReading represents smart plug energy consumption data
type EnergyReading struct {
	DeviceID       string    `json:"device_id"`
	DeviceName     string    `json:"device_name"`
	RoomID         string    `json:"room_id"`
	PowerW         float64   `json:"power_w"`                   // Current power in watts
	EnergyWh       float64   `json:"energy_wh"`                 // Total energy consumed in watt-hours
	VoltageV       float64   `json:"voltage_v"`                 // Voltage in volts
	CurrentA       float64   `json:"current_a"`                 // Current in amperes
	IsOn           bool      `json:"is_on"`                     // Device on/off state
	Temperature    float64   `json:"temperature,omitempty"`     // Device temperature if available
	SignalStrength int       `json:"signal_strength,omitempty"` // WiFi signal strength
	Timestamp      time.Time `json:"timestamp"`
}

// WriteEnergyReading writes smart plug energy data to InfluxDB
func (c *Client) WriteEnergyReading(reading *EnergyReading) error {
	if c.state != StateConnected {
		return errors.NewConnectionError("InfluxDB client not connected", nil)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	return utils.Retry(ctx, c.retryConfig, func() error {
		// Create point for energy data
		point := influxdb2.NewPointWithMeasurement("energy").
			AddTag("device_id", reading.DeviceID).
			AddTag("device_name", reading.DeviceName).
			AddTag("room_id", reading.RoomID).
			AddField("power_w", reading.PowerW).
			AddField("energy_wh", reading.EnergyWh).
			AddField("voltage_v", reading.VoltageV).
			AddField("current_a", reading.CurrentA).
			AddField("is_on", reading.IsOn).
			SetTime(reading.Timestamp)

		// Add optional fields
		if reading.Temperature > 0 {
			point.AddField("temperature", reading.Temperature)
		}
		if reading.SignalStrength != 0 {
			point.AddField("signal_strength", reading.SignalStrength)
		}

		// Write point
		c.writeAPI.WritePoint(point)

		// Check for write errors (non-blocking)
		select {
		case err := <-c.writeAPI.Errors():
			return errors.NewServiceError("Failed to write energy reading to InfluxDB", err)
		default:
			return nil
		}
	})
}

// QueryEnergyData queries energy consumption data from InfluxDB
func (c *Client) QueryEnergyData(deviceID, roomID string, timeRange time.Duration) ([]map[string]interface{}, error) {
	if c.state != StateConnected {
		return nil, errors.NewConnectionError("InfluxDB client not connected", nil)
	}

	query := fmt.Sprintf(`
		from(bucket: "%s")
		|> range(start: -%s)
		|> filter(fn: (r) => r._measurement == "energy")`,
		c.bucket, timeRange.String())

	if deviceID != "" {
		query += fmt.Sprintf(`|> filter(fn: (r) => r.device_id == "%s")`, deviceID)
	}
	if roomID != "" {
		query += fmt.Sprintf(`|> filter(fn: (r) => r.room_id == "%s")`, roomID)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	result, err := c.queryAPI.Query(ctx, query)
	if err != nil {
		return nil, errors.NewServiceError("Failed to query energy data from InfluxDB", err)
	}

	var data []map[string]interface{}
	for result.Next() {
		record := result.Record()
		data = append(data, map[string]interface{}{
			"time":        record.Time(),
			"measurement": record.Measurement(),
			"field":       record.Field(),
			"value":       record.Value(),
			"device_id":   record.ValueByKey("device_id"),
			"device_name": record.ValueByKey("device_name"),
			"room_id":     record.ValueByKey("room_id"),
		})
	}

	if result.Err() != nil {
		return nil, errors.NewServiceError("Error processing energy query result", result.Err())
	}

	return data, nil
}

// Flush forces all pending writes to be sent to InfluxDB
func (c *Client) Flush() {
	if c.writeAPI != nil {
		c.writeAPI.Flush()
	}
}

package prometheus

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/api"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Client represents a Prometheus client
type Client struct {
	client api.Client
	api    v1.API
	url    string

	// Metrics for Tapo energy monitoring
	energyMetrics *EnergyMetrics
}

// EnergyMetrics holds all Prometheus metrics for energy monitoring
type EnergyMetrics struct {
	PowerConsumption *prometheus.GaugeVec
	EnergyTotal      *prometheus.CounterVec
	Voltage          *prometheus.GaugeVec
	Current          *prometheus.GaugeVec
	DeviceStatus     *prometheus.GaugeVec
	SignalStrength   *prometheus.GaugeVec
	Temperature      *prometheus.GaugeVec
}

// EnergyReading represents energy data from Tapo devices
type EnergyReading struct {
	DeviceID       string
	DeviceName     string
	RoomID         string
	PowerW         float64
	EnergyWh       float64
	VoltageV       float64
	CurrentA       float64
	IsOn           bool
	SignalStrength float64
	Temperature    float64
	Timestamp      time.Time
}

// NewClient creates a new Prometheus client
func NewClient(url string) *Client {
	config := api.Config{
		Address: url,
	}

	client, err := api.NewClient(config)
	if err != nil {
		log.Printf("Error creating Prometheus client: %v", err)
		return nil
	}

	v1api := v1.NewAPI(client)

	// Initialize energy metrics
	energyMetrics := &EnergyMetrics{
		PowerConsumption: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "tapo_power_consumption_watts",
				Help: "Current power consumption in watts",
			},
			[]string{"device_id", "device_name", "room_id"},
		),
		EnergyTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "tapo_energy_total_wh",
				Help: "Total energy consumption in watt-hours",
			},
			[]string{"device_id", "device_name", "room_id"},
		),
		Voltage: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "tapo_voltage_volts",
				Help: "Supply voltage in volts",
			},
			[]string{"device_id", "device_name", "room_id"},
		),
		Current: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "tapo_current_amperes",
				Help: "Current draw in amperes",
			},
			[]string{"device_id", "device_name", "room_id"},
		),
		DeviceStatus: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "tapo_device_status",
				Help: "Device on/off status (1 = on, 0 = off)",
			},
			[]string{"device_id", "device_name", "room_id"},
		),
		SignalStrength: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "tapo_signal_strength_dbm",
				Help: "WiFi signal strength in dBm",
			},
			[]string{"device_id", "device_name", "room_id"},
		),
		Temperature: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "tapo_temperature_celsius",
				Help: "Device temperature in Celsius",
			},
			[]string{"device_id", "device_name", "room_id"},
		),
	}

	return &Client{
		client:        client,
		api:           v1api,
		url:           url,
		energyMetrics: energyMetrics,
	}
}

// Connect establishes connection to Prometheus
func (c *Client) Connect() error {
	if c == nil {
		return fmt.Errorf("prometheus client is nil")
	}

	// Test connection by making a simple query
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := c.api.Config(ctx)
	if err != nil {
		return fmt.Errorf("failed to connect to Prometheus: %w", err)
	}

	return nil
}

// Disconnect cleans up the client connection
func (c *Client) Disconnect() {
	// Prometheus client doesn't require explicit disconnection
	log.Println("Prometheus client disconnected")
}

// WriteEnergyReading records energy metrics to Prometheus
func (c *Client) WriteEnergyReading(ctx context.Context, deviceID, roomID string, powerW, energyWh, voltageV, currentA float64, isOn bool, timestamp time.Time) error {
	if c == nil || c.energyMetrics == nil {
		return fmt.Errorf("prometheus client or metrics not initialized")
	}

	deviceName := deviceID // Use device ID as name if not provided separately
	labels := prometheus.Labels{
		"device_id":   deviceID,
		"device_name": deviceName,
		"room_id":     roomID,
	}

	// Set metrics
	c.energyMetrics.PowerConsumption.With(labels).Set(powerW)
	c.energyMetrics.EnergyTotal.With(labels).Add(energyWh) // Counter should only increase
	c.energyMetrics.Voltage.With(labels).Set(voltageV)
	c.energyMetrics.Current.With(labels).Set(currentA)

	if isOn {
		c.energyMetrics.DeviceStatus.With(labels).Set(1)
	} else {
		c.energyMetrics.DeviceStatus.With(labels).Set(0)
	}

	return nil
}

// WriteEnergyReadingComplete records all energy metrics including signal and temperature
func (c *Client) WriteEnergyReadingComplete(ctx context.Context, reading *EnergyReading) error {
	if c == nil || c.energyMetrics == nil {
		return fmt.Errorf("prometheus client or metrics not initialized")
	}

	labels := prometheus.Labels{
		"device_id":   reading.DeviceID,
		"device_name": reading.DeviceName,
		"room_id":     reading.RoomID,
	}

	// Set all metrics
	c.energyMetrics.PowerConsumption.With(labels).Set(reading.PowerW)
	c.energyMetrics.EnergyTotal.With(labels).Add(reading.EnergyWh)
	c.energyMetrics.Voltage.With(labels).Set(reading.VoltageV)
	c.energyMetrics.Current.With(labels).Set(reading.CurrentA)
	c.energyMetrics.SignalStrength.With(labels).Set(reading.SignalStrength)
	c.energyMetrics.Temperature.With(labels).Set(reading.Temperature)

	if reading.IsOn {
		c.energyMetrics.DeviceStatus.With(labels).Set(1)
	} else {
		c.energyMetrics.DeviceStatus.With(labels).Set(0)
	}

	return nil
}

// WriteTemperatureReading records temperature metrics to Prometheus
func (c *Client) WriteTemperatureReading(ctx context.Context, deviceID, roomID string, tempF, humidity float64, timestamp time.Time) error {
	if c == nil {
		return fmt.Errorf("prometheus client not initialized")
	}

	// Convert Fahrenheit to Celsius for Prometheus
	tempC := (tempF - 32) * 5 / 9

	labels := prometheus.Labels{
		"device_id": deviceID,
		"room_id":   roomID,
	}

	c.energyMetrics.Temperature.With(labels).Set(tempC)

	return nil
}

// QueryEnergyConsumption queries energy consumption for a device
func (c *Client) QueryEnergyConsumption(ctx context.Context, deviceID string, duration time.Duration) (float64, error) {
	if c == nil {
		return 0, fmt.Errorf("prometheus client not initialized")
	}

	query := fmt.Sprintf(`increase(tapo_energy_total_wh{device_id="%s"}[%s])`, deviceID, duration.String())

	_, warnings, err := c.api.Query(ctx, query, time.Now())
	if err != nil {
		return 0, fmt.Errorf("error querying Prometheus: %w", err)
	}

	if len(warnings) > 0 {
		log.Printf("Prometheus query warnings: %v", warnings)
	}

	// Parse result and return energy consumption
	// This is a simplified implementation
	return 0, nil
}

// GetLatestEnergyReadings queries the latest energy readings for a device
func (c *Client) GetLatestEnergyReadings(ctx context.Context, deviceID string, limit int) ([]*EnergyReading, error) {
	if c == nil {
		return nil, fmt.Errorf("prometheus client not initialized")
	}

	// Query current power consumption
	query := fmt.Sprintf(`tapo_power_consumption_watts{device_id="%s"}`, deviceID)

	_, warnings, err := c.api.Query(ctx, query, time.Now())
	if err != nil {
		return nil, fmt.Errorf("error querying Prometheus: %w", err)
	}

	if len(warnings) > 0 {
		log.Printf("Prometheus query warnings: %v", warnings)
	}

	// This is a simplified implementation - in practice you'd parse the result
	// and create EnergyReading structs
	readings := make([]*EnergyReading, 0)

	return readings, nil
}

// GetMetricsHandler returns HTTP handler for Prometheus metrics
func (c *Client) GetMetricsHandler() http.Handler {
	return promhttp.Handler()
}

// Health checks if Prometheus is healthy
func (c *Client) Health(ctx context.Context) error {
	if c == nil {
		return fmt.Errorf("prometheus client not initialized")
	}

	_, err := c.api.Config(ctx)
	return err
}

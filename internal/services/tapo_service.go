package services

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/johnpr01/home-automation/internal/errors"
	"github.com/johnpr01/home-automation/internal/logger"
	"github.com/johnpr01/home-automation/pkg/influxdb"
	"github.com/johnpr01/home-automation/pkg/mqtt"
	"github.com/johnpr01/home-automation/pkg/tapo"
)

// TapoService manages TP-Link Tapo smart plugs and energy monitoring
type TapoService struct {
	devices      map[string]*TapoDeviceManager
	mqttClient   *mqtt.Client
	influxClient *influxdb.Client
	logger       *logger.Logger
	mu           sync.RWMutex
	running      bool
	stopChan     chan struct{}
}

// TapoDeviceManager manages a single Tapo device
type TapoDeviceManager struct {
	DeviceID     string
	DeviceName   string
	RoomID       string
	IPAddress    string
	Username     string
	Password     string
	Client       *tapo.TapoClient
	PollInterval time.Duration
	LastReading  time.Time
	IsConnected  bool
}

// TapoConfig represents configuration for Tapo devices
type TapoConfig struct {
	DeviceID     string        `json:"device_id"`
	DeviceName   string        `json:"device_name"`
	RoomID       string        `json:"room_id"`
	IPAddress    string        `json:"ip_address"`
	Username     string        `json:"username"`
	Password     string        `json:"password"`
	PollInterval time.Duration `json:"poll_interval"`
}

// NewTapoService creates a new Tapo service
func NewTapoService(mqttClient *mqtt.Client, influxClient *influxdb.Client, serviceLogger *logger.Logger) *TapoService {
	return &TapoService{
		devices:      make(map[string]*TapoDeviceManager),
		mqttClient:   mqttClient,
		influxClient: influxClient,
		logger:       serviceLogger,
		stopChan:     make(chan struct{}),
	}
}

// AddDevice adds a new Tapo device to monitor
func (ts *TapoService) AddDevice(config *TapoConfig) error {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	if config.PollInterval == 0 {
		config.PollInterval = 30 * time.Second // Default 30 seconds
	}

	// Create Tapo client
	client := tapo.NewTapoClient(config.IPAddress, config.Username, config.Password, ts.logger)

	// Test connection
	if err := client.Connect(); err != nil {
		return errors.NewDeviceError(fmt.Sprintf("Failed to connect to Tapo device %s", config.DeviceID), err)
	}

	manager := &TapoDeviceManager{
		DeviceID:     config.DeviceID,
		DeviceName:   config.DeviceName,
		RoomID:       config.RoomID,
		IPAddress:    config.IPAddress,
		Username:     config.Username,
		Password:     config.Password,
		Client:       client,
		PollInterval: config.PollInterval,
		IsConnected:  true,
	}

	ts.devices[config.DeviceID] = manager

	ts.logger.Info("Added Tapo device", map[string]interface{}{
		"device_id":   config.DeviceID,
		"device_name": config.DeviceName,
		"room_id":     config.RoomID,
		"ip_address":  config.IPAddress,
	})

	return nil
}

// RemoveDevice removes a Tapo device from monitoring
func (ts *TapoService) RemoveDevice(deviceID string) error {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	if _, exists := ts.devices[deviceID]; !exists {
		return errors.NewValidationError(fmt.Sprintf("Device %s not found", deviceID), nil)
	}

	delete(ts.devices, deviceID)

	ts.logger.Info("Removed Tapo device", map[string]interface{}{
		"device_id": deviceID,
	})

	return nil
}

// Start begins monitoring all configured devices
func (ts *TapoService) Start() error {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	if ts.running {
		return errors.NewServiceError("Tapo service is already running", nil)
	}

	ts.running = true

	// Start monitoring goroutines for each device
	for deviceID, manager := range ts.devices {
		go ts.monitorDevice(deviceID, manager)
	}

	ts.logger.Info("Started Tapo monitoring service", map[string]interface{}{
		"device_count": len(ts.devices),
	})

	return nil
}

// Stop stops monitoring all devices
func (ts *TapoService) Stop() error {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	if !ts.running {
		return nil
	}

	ts.running = false
	close(ts.stopChan)

	ts.logger.Info("Stopped Tapo monitoring service")
	return nil
}

// monitorDevice continuously monitors a single Tapo device
func (ts *TapoService) monitorDevice(deviceID string, manager *TapoDeviceManager) {
	ticker := time.NewTicker(manager.PollInterval)
	defer ticker.Stop()

	ts.logger.Info("Started monitoring Tapo device", map[string]interface{}{
		"device_id":     deviceID,
		"poll_interval": manager.PollInterval.String(),
	})

	for {
		select {
		case <-ts.stopChan:
			return
		case <-ticker.C:
			ts.pollDevice(manager)
		}
	}
}

// pollDevice polls a single device for energy data
func (ts *TapoService) pollDevice(manager *TapoDeviceManager) {
	// Reconnect if needed
	if !manager.IsConnected {
		if err := manager.Client.Connect(); err != nil {
			ts.logger.Error("Failed to reconnect to Tapo device", err, map[string]interface{}{
				"device_id": manager.DeviceID,
			})
			return
		}
		manager.IsConnected = true
	}

	// Get device info
	deviceInfo, err := manager.Client.GetDeviceInfo()
	if err != nil {
		ts.logger.Error("Failed to get device info", err, map[string]interface{}{
			"device_id": manager.DeviceID,
		})
		manager.IsConnected = false
		return
	}

	// Get energy usage
	energyUsage, err := manager.Client.GetEnergyUsage()
	if err != nil {
		ts.logger.Error("Failed to get energy usage", err, map[string]interface{}{
			"device_id": manager.DeviceID,
		})
		return
	}

	// Convert to InfluxDB reading
	reading := &influxdb.EnergyReading{
		DeviceID:       manager.DeviceID,
		DeviceName:     manager.DeviceName,
		RoomID:         manager.RoomID,
		PowerW:         float64(energyUsage.CurrentPowerMw) / 1000.0, // Convert mW to W
		EnergyWh:       float64(energyUsage.TodayEnergyWh),
		IsOn:           deviceInfo.IsOn,
		SignalStrength: deviceInfo.SignalLevel,
		Timestamp:      time.Now(),
	}

	// Store in InfluxDB
	if ts.influxClient != nil && ts.influxClient.IsConnected() {
		if err := ts.influxClient.WriteEnergyReading(reading); err != nil {
			ts.logger.Error("Failed to write energy reading to InfluxDB", err, map[string]interface{}{
				"device_id": manager.DeviceID,
			})
		}
	}

	// Publish to MQTT
	if ts.mqttClient != nil {
		topic := fmt.Sprintf("tapo/%s/energy", manager.DeviceID)

		payload := map[string]interface{}{
			"device_id":       reading.DeviceID,
			"device_name":     reading.DeviceName,
			"room_id":         reading.RoomID,
			"power_w":         reading.PowerW,
			"energy_wh":       reading.EnergyWh,
			"is_on":           reading.IsOn,
			"signal_strength": reading.SignalStrength,
			"timestamp":       reading.Timestamp.Unix(),
		}

		// Convert to JSON
		payloadBytes, err := json.Marshal(payload)
		if err != nil {
			ts.logger.Error("Failed to marshal MQTT payload", err, map[string]interface{}{
				"device_id": manager.DeviceID,
			})
			return
		}

		message := &mqtt.Message{
			Topic:   topic,
			Payload: payloadBytes,
			QoS:     1,
			Retain:  false,
		}

		if err := ts.mqttClient.Publish(message); err != nil {
			ts.logger.Error("Failed to publish energy data to MQTT", err, map[string]interface{}{
				"device_id": manager.DeviceID,
				"topic":     topic,
			})
		}
	}

	manager.LastReading = time.Now()

	ts.logger.Debug("Polled Tapo device", map[string]interface{}{
		"device_id": manager.DeviceID,
		"power_w":   reading.PowerW,
		"energy_wh": reading.EnergyWh,
		"is_on":     reading.IsOn,
	})
}

// SetDeviceState turns a device on or off
func (ts *TapoService) SetDeviceState(deviceID string, on bool) error {
	ts.mu.RLock()
	manager, exists := ts.devices[deviceID]
	ts.mu.RUnlock()

	if !exists {
		return errors.NewValidationError(fmt.Sprintf("Device %s not found", deviceID), nil)
	}

	if !manager.IsConnected {
		if err := manager.Client.Connect(); err != nil {
			return errors.NewDeviceError("Failed to connect to device", err)
		}
		manager.IsConnected = true
	}

	if err := manager.Client.SetDeviceOn(on); err != nil {
		manager.IsConnected = false
		return errors.NewDeviceError("Failed to set device state", err)
	}

	ts.logger.Info("Changed device state", map[string]interface{}{
		"device_id": deviceID,
		"state":     on,
	})

	return nil
}

// GetDeviceStatus returns the current status of all devices
func (ts *TapoService) GetDeviceStatus() map[string]interface{} {
	ts.mu.RLock()
	defer ts.mu.RUnlock()

	status := make(map[string]interface{})
	for deviceID, manager := range ts.devices {
		status[deviceID] = map[string]interface{}{
			"device_name":   manager.DeviceName,
			"room_id":       manager.RoomID,
			"ip_address":    manager.IPAddress,
			"is_connected":  manager.IsConnected,
			"last_reading":  manager.LastReading,
			"poll_interval": manager.PollInterval.String(),
		}
	}

	return map[string]interface{}{
		"running":      ts.running,
		"device_count": len(ts.devices),
		"devices":      status,
	}
}

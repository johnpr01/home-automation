package services

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/johnpr01/home-automation/internal/errors"
	"github.com/johnpr01/home-automation/internal/logger"
	"github.com/johnpr01/home-automation/pkg/mqtt"
	"github.com/johnpr01/home-automation/pkg/tapo"
)

// TapoService manages TP-Link Tapo smart plugs and energy monitoring
type TapoService struct {
	devices    map[string]*TapoDeviceManager
	mqttClient *mqtt.Client
	tsClient   TimeSeriesClient
	logger     *logger.Logger
	mu         sync.RWMutex
	running    bool
	stopChan   chan struct{}
}

// TapoDeviceManager manages a single Tapo device
type TapoDeviceManager struct {
	DeviceID     string
	DeviceName   string
	RoomID       string
	IPAddress    string
	Username     string
	Password     string
	Client       interface{} // Can be *tapo.TapoClient or *tapo.KlapClient
	KlapClient   *tapo.KlapClient
	PollInterval time.Duration
	LastReading  time.Time
	IsConnected  bool
	UseKlap      bool
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
	UseKlap      bool          `json:"use_klap"`
}

// NewTapoService creates a new Tapo service
func NewTapoService(mqttClient *mqtt.Client, tsClient TimeSeriesClient, serviceLogger *logger.Logger) *TapoService {
	return &TapoService{
		devices:    make(map[string]*TapoDeviceManager),
		mqttClient: mqttClient,
		tsClient:   tsClient,
		logger:     serviceLogger,
		stopChan:   make(chan struct{}),
	}
}

// AddDevice adds a new Tapo device to monitor
func (ts *TapoService) AddDevice(config *TapoConfig) error {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	if config.PollInterval == 0 {
		config.PollInterval = 30 * time.Second // Default 30 seconds
	}

	manager := &TapoDeviceManager{
		DeviceID:     config.DeviceID,
		DeviceName:   config.DeviceName,
		RoomID:       config.RoomID,
		IPAddress:    config.IPAddress,
		Username:     config.Username,
		Password:     config.Password,
		PollInterval: config.PollInterval,
		UseKlap:      config.UseKlap,
	}

	// Create appropriate client based on configuration
	if config.UseKlap {
		// Create KLAP client for newer firmware
		klapClient := tapo.NewKlapClient(config.IPAddress, config.Username, config.Password, 30*time.Second, *ts.logger)
		manager.KlapClient = klapClient

		// Test connection
		ctx := context.Background()
		if err := klapClient.Connect(ctx); err != nil {
			return errors.NewDeviceError(fmt.Sprintf("Failed to connect to Tapo device %s using KLAP", config.DeviceID), err)
		}
	} else {
		// Create legacy client for older firmware
		client := tapo.NewTapoClient(config.IPAddress, config.Username, config.Password, ts.logger)
		manager.Client = client

		// Test connection
		if err := client.Connect(); err != nil {
			return errors.NewDeviceError(fmt.Sprintf("Failed to connect to Tapo device %s", config.DeviceID), err)
		}
	}

	manager.IsConnected = true
	ts.devices[config.DeviceID] = manager

	ts.logger.Info("Added Tapo device", map[string]interface{}{
		"device_id":   config.DeviceID,
		"device_name": config.DeviceName,
		"room_id":     config.RoomID,
		"ip_address":  config.IPAddress,
		"use_klap":    config.UseKlap,
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
		if manager.UseKlap && manager.KlapClient != nil {
			ctx := context.Background()
			if err := manager.KlapClient.Connect(ctx); err != nil {
				ts.logger.Error("Failed to reconnect to Tapo device using KLAP", err, map[string]interface{}{
					"device_id": manager.DeviceID,
				})
				return
			}
		} else if client, ok := manager.Client.(*tapo.TapoClient); ok {
			if err := client.Connect(); err != nil {
				ts.logger.Error("Failed to reconnect to Tapo device", err, map[string]interface{}{
					"device_id": manager.DeviceID,
				})
				return
			}
		} else {
			ts.logger.Error("Invalid client type for device", nil, map[string]interface{}{
				"device_id": manager.DeviceID,
			})
			return
		}
		manager.IsConnected = true
	}

	var deviceInfo interface{}
	var energyUsage interface{}

	// Get device info and energy usage based on client type
	if manager.UseKlap && manager.KlapClient != nil {
		ctx := context.Background()
		klapDeviceInfo, err := manager.KlapClient.GetDeviceInfo(ctx)
		if err != nil {
			ts.logger.Error("Failed to get device info via KLAP", err, map[string]interface{}{
				"device_id": manager.DeviceID,
			})
			manager.IsConnected = false
			return
		}
		deviceInfo = klapDeviceInfo

		klapEnergyUsage, err := manager.KlapClient.GetEnergyUsage(ctx)
		if err != nil {
			ts.logger.Error("Failed to get energy usage via KLAP", err, map[string]interface{}{
				"device_id": manager.DeviceID,
			})
			return
		}
		energyUsage = klapEnergyUsage
	} else if client, ok := manager.Client.(*tapo.TapoClient); ok {
		legacyDeviceInfo, err := client.GetDeviceInfo()
		if err != nil {
			ts.logger.Error("Failed to get device info", err, map[string]interface{}{
				"device_id": manager.DeviceID,
			})
			manager.IsConnected = false
			return
		}
		deviceInfo = legacyDeviceInfo

		legacyEnergyUsage, err := client.GetEnergyUsage()
		if err != nil {
			ts.logger.Error("Failed to get energy usage", err, map[string]interface{}{
				"device_id": manager.DeviceID,
			})
			return
		}
		energyUsage = legacyEnergyUsage
	} else {
		ts.logger.Error("Invalid client type for device", nil, map[string]interface{}{
			"device_id": manager.DeviceID,
		})
		return
	}

	// Convert to energy reading (handle both device info types)
	var reading *EnergyReading
	if manager.UseKlap {
		klapDeviceInfo := deviceInfo.(*tapo.KlapDeviceInfo)
		klapEnergyUsage := energyUsage.(*tapo.KlapEnergyUsage)

		reading = &EnergyReading{
			DeviceID:       manager.DeviceID,
			DeviceName:     manager.DeviceName,
			RoomID:         manager.RoomID,
			PowerW:         float64(klapEnergyUsage.CurrentPower) / 1000.0, // Convert mW to W
			EnergyWh:       float64(klapEnergyUsage.TodayEnergy),
			IsOn:           klapDeviceInfo.DeviceOn,
			SignalStrength: float64(klapDeviceInfo.SignalLevel),
			Timestamp:      time.Now(),
		}
	} else {
		legacyDeviceInfo := deviceInfo.(*tapo.TapoDevice)
		legacyEnergyUsage := energyUsage.(*tapo.EnergyUsage)

		reading = &EnergyReading{
			DeviceID:       manager.DeviceID,
			DeviceName:     manager.DeviceName,
			RoomID:         manager.RoomID,
			PowerW:         float64(legacyEnergyUsage.CurrentPowerMw) / 1000.0, // Convert mW to W
			EnergyWh:       float64(legacyEnergyUsage.TodayEnergyWh),
			IsOn:           legacyDeviceInfo.IsOn,
			SignalStrength: float64(legacyDeviceInfo.SignalLevel),
			Timestamp:      time.Now(),
		}
	}

	// Store in time series database
	if ts.tsClient != nil {
		if err := ts.tsClient.WriteEnergyReading(context.Background(), reading.DeviceID, reading.RoomID,
			reading.PowerW, reading.EnergyWh, 0, 0, reading.IsOn, reading.Timestamp); err != nil {
			ts.logger.Error("Failed to write energy reading to time series database", err, map[string]interface{}{
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
		if manager.UseKlap && manager.KlapClient != nil {
			ctx := context.Background()
			if err := manager.KlapClient.Connect(ctx); err != nil {
				return errors.NewDeviceError("Failed to connect to device via KLAP", err)
			}
		} else if client, ok := manager.Client.(*tapo.TapoClient); ok {
			if err := client.Connect(); err != nil {
				return errors.NewDeviceError("Failed to connect to device", err)
			}
		} else {
			return errors.NewDeviceError("Invalid client type for device", nil)
		}
		manager.IsConnected = true
	}

	// Set device state based on client type
	if manager.UseKlap && manager.KlapClient != nil {
		// KLAP protocol doesn't have SetDeviceOn implemented yet
		// For now, we'll return an error indicating this feature is not available
		return errors.NewBusinessError("SetDeviceOn not implemented for KLAP protocol", nil)
	} else if client, ok := manager.Client.(*tapo.TapoClient); ok {
		if err := client.SetDeviceOn(on); err != nil {
			manager.IsConnected = false
			return errors.NewDeviceError("Failed to set device state", err)
		}
	} else {
		return errors.NewDeviceError("Invalid client type for device", nil)
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

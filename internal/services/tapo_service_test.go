package services

import (
	"testing"
	"time"

	"github.com/johnpr01/home-automation/internal/logger"
)

func TestNewTapoService(t *testing.T) {
	// Create a test logger
	serviceLogger := logger.NewLogger("test-tapo-service", nil)

	// Create a new Tapo service
	service := NewTapoService(nil, nil, serviceLogger)

	if service == nil {
		t.Fatal("NewTapoService returned nil")
	}

	if service.devices == nil {
		t.Error("Service devices map is nil")
	}

	if service.logger == nil {
		t.Error("Service logger is nil")
	}

	if service.stopChan == nil {
		t.Error("Service stop channel is nil")
	}
}

func TestTapoConfig(t *testing.T) {
	// Test creating a Tapo config for KLAP protocol
	config := &TapoConfig{
		DeviceID:     "test_device",
		DeviceName:   "Test Device",
		RoomID:       "test_room",
		IPAddress:    "192.168.1.100",
		Username:     "test_user",
		Password:     "test_pass",
		PollInterval: 30 * time.Second,
		UseKlap:      true,
	}

	if config.DeviceID != "test_device" {
		t.Errorf("Expected device ID to be 'test_device', got '%s'", config.DeviceID)
	}

	if !config.UseKlap {
		t.Error("Expected UseKlap to be true")
	}

	if config.PollInterval != 30*time.Second {
		t.Errorf("Expected poll interval to be 30s, got %v", config.PollInterval)
	}

	// Test creating a config for legacy protocol
	legacyConfig := &TapoConfig{
		DeviceID:     "legacy_device",
		DeviceName:   "Legacy Device",
		RoomID:       "test_room",
		IPAddress:    "192.168.1.101",
		Username:     "test_user",
		Password:     "test_pass",
		PollInterval: 60 * time.Second,
		UseKlap:      false,
	}

	if legacyConfig.UseKlap {
		t.Error("Expected UseKlap to be false for legacy config")
	}
}

func TestTapoDeviceManager(t *testing.T) {
	// Test creating a device manager
	manager := &TapoDeviceManager{
		DeviceID:     "test_device",
		DeviceName:   "Test Device",
		RoomID:       "test_room",
		IPAddress:    "192.168.1.100",
		Username:     "test_user",
		Password:     "test_pass",
		PollInterval: 30 * time.Second,
		UseKlap:      true,
		IsConnected:  false,
	}

	if manager.DeviceID != "test_device" {
		t.Errorf("Expected device ID to be 'test_device', got '%s'", manager.DeviceID)
	}

	if !manager.UseKlap {
		t.Error("Expected UseKlap to be true")
	}

	if manager.IsConnected {
		t.Error("Expected device to not be connected initially")
	}
}

func TestEnergyReading(t *testing.T) {
	// Test creating an energy reading
	reading := &EnergyReading{
		DeviceID:       "test_device",
		DeviceName:     "Test Device",
		RoomID:         "test_room",
		PowerW:         2.5,  // 2.5 watts
		EnergyWh:       1000, // 1 kWh
		IsOn:           true,
		SignalStrength: 75.0,
		Timestamp:      time.Now(),
	}

	if reading.DeviceID != "test_device" {
		t.Errorf("Expected device ID to be 'test_device', got '%s'", reading.DeviceID)
	}

	if reading.PowerW != 2.5 {
		t.Errorf("Expected power to be 2.5W, got %f", reading.PowerW)
	}

	if !reading.IsOn {
		t.Error("Expected device to be on")
	}

	if reading.EnergyWh != 1000 {
		t.Errorf("Expected energy to be 1000Wh, got %f", reading.EnergyWh)
	}
}

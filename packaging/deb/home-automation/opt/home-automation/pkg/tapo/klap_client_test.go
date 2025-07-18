package tapo

import (
	"testing"
	"time"

	"github.com/johnpr01/home-automation/internal/logger"
)

func TestNewKlapClient(t *testing.T) {
	// Test creating a new KLAP client
	logger := logger.NewLogger("test", nil)

	client := NewKlapClient("192.168.1.100", "test_user", "test_pass", 30*time.Second, *logger)

	if client == nil {
		t.Fatal("NewKlapClient returned nil")
	}

	if client.baseURL != "http://192.168.1.100" {
		t.Errorf("Expected baseURL to be 'http://192.168.1.100', got '%s'", client.baseURL)
	}

	if client.username != "test_user" {
		t.Errorf("Expected username to be 'test_user', got '%s'", client.username)
	}

	if client.password != "test_pass" {
		t.Errorf("Expected password to be 'test_pass', got '%s'", client.password)
	}

	if client.timeout != 30*time.Second {
		t.Errorf("Expected timeout to be 30s, got %v", client.timeout)
	}
}

func TestCalcAuthHash(t *testing.T) {
	logger := logger.NewLogger("test", nil)
	client := NewKlapClient("192.168.1.100", "test_user", "test_pass", 30*time.Second, *logger)

	// Test auth hash calculation
	authHash := client.calcAuthHash("test_user", "test_pass")

	if len(authHash) != 32 { // SHA256 produces 32 bytes
		t.Errorf("Expected auth hash to be 32 bytes, got %d", len(authHash))
	}

	// Test that the same inputs produce the same hash
	authHash2 := client.calcAuthHash("test_user", "test_pass")

	if string(authHash) != string(authHash2) {
		t.Error("Auth hash calculation is not deterministic")
	}
}

func TestCryptoHelpers(t *testing.T) {
	// Test SHA256 helper
	input := []byte("test data")
	hash := sha256Hash(input)

	if len(hash) != 32 {
		t.Errorf("Expected SHA256 hash to be 32 bytes, got %d", len(hash))
	}

	// Test SHA1 helper
	hash1 := sha1Hash(input)

	if len(hash1) != 20 {
		t.Errorf("Expected SHA1 hash to be 20 bytes, got %d", len(hash1))
	}

	// Test concat helper
	arr1 := []byte("hello")
	arr2 := []byte("world")
	result := concat(arr1, arr2)

	expected := "helloworld"
	if string(result) != expected {
		t.Errorf("Expected concat result to be '%s', got '%s'", expected, string(result))
	}
}

func TestDeviceInfoStruct(t *testing.T) {
	// Test that our device info struct can be created and used
	deviceInfo := &KlapDeviceInfo{
		DeviceID:    "test_device",
		Model:       "P110",
		FwVersion:   "1.1.0",
		DeviceOn:    true,
		RSSI:        -45,
		SignalLevel: 3,
	}

	if deviceInfo.DeviceID != "test_device" {
		t.Errorf("Expected device ID to be 'test_device', got '%s'", deviceInfo.DeviceID)
	}

	if !deviceInfo.DeviceOn {
		t.Error("Expected device to be on")
	}

	if deviceInfo.RSSI != -45 {
		t.Errorf("Expected RSSI to be -45, got %d", deviceInfo.RSSI)
	}
}

func TestEnergyUsageStruct(t *testing.T) {
	// Test that our energy usage struct can be created and used
	energyUsage := &KlapEnergyUsage{
		CurrentPower: 2500,  // 2.5W in mW
		TodayEnergy:  1000,  // 1kWh in Wh
		MonthEnergy:  30000, // 30kWh in Wh
		TodayRuntime: 480,   // 8 hours in minutes
	}

	if energyUsage.CurrentPower != 2500 {
		t.Errorf("Expected current power to be 2500, got %d", energyUsage.CurrentPower)
	}

	if energyUsage.TodayEnergy != 1000 {
		t.Errorf("Expected today energy to be 1000, got %d", energyUsage.TodayEnergy)
	}
}

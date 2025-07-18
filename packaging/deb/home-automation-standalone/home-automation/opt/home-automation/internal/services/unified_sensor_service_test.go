package services

import (
	"encoding/json"
	"log"
	"os"
	"testing"
	"time"

	"github.com/johnpr01/home-automation/internal/config"
	"github.com/johnpr01/home-automation/pkg/mqtt"
)

func TestUnifiedSensorService(t *testing.T) {
	// Create logger
	logger := log.New(os.Stdout, "[TEST] ", log.LstdFlags)

	// Create MQTT client
	mqttConfig := &config.MQTTConfig{
		Broker: "localhost",
		Port:   "1883",
	}
	mqttClient := mqtt.NewClient(mqttConfig, nil)

	// Create unified sensor service
	service := NewUnifiedSensorService(mqttClient, logger)

	// Test temperature callback
	tempCallbackCalled := false
	service.AddTemperatureCallback(func(roomID string, temperature float64) {
		tempCallbackCalled = true
		if roomID != "living-room" {
			t.Errorf("Expected roomID 'living-room', got '%s'", roomID)
		}
		if temperature != 72.5 {
			t.Errorf("Expected temperature 72.5, got %f", temperature)
		}
	})

	// Test motion callback
	motionCallbackCalled := false
	service.AddMotionCallback(func(roomID string, occupied bool) {
		motionCallbackCalled = true
		if roomID != "living-room" {
			t.Errorf("Expected roomID 'living-room', got '%s'", roomID)
		}
		if !occupied {
			t.Errorf("Expected occupied true, got %t", occupied)
		}
	})

	// Test light callback
	lightCallbackCalled := false
	service.AddLightCallback(func(roomID string, lightState string, lightLevel float64) {
		lightCallbackCalled = true
		if roomID != "living-room" {
			t.Errorf("Expected roomID 'living-room', got '%s'", roomID)
		}
		if lightState != "bright" {
			t.Errorf("Expected lightState 'bright', got '%s'", lightState)
		}
		if lightLevel != 85.5 {
			t.Errorf("Expected lightLevel 85.5, got %f", lightLevel)
		}
	})

	// Simulate temperature message
	tempMsg := UnifiedSensorMessage{
		Temperature: 72.5,
		TempUnit:    "F",
		Room:        "living-room",
		Sensor:      "SHT30",
		Timestamp:   time.Now().Unix(),
		DeviceID:    "pico-living-room",
	}
	tempPayload, _ := json.Marshal(tempMsg)
	service.handleTemperatureMessage("room-temp/living-room", tempPayload)

	// Simulate motion message
	motionTrue := true
	motionMsg := UnifiedSensorMessage{
		Motion:    &motionTrue,
		Room:      "living-room",
		Sensor:    "PIR",
		Timestamp: time.Now().Unix(),
		DeviceID:  "pico-living-room",
	}
	motionPayload, _ := json.Marshal(motionMsg)
	service.handleMotionMessage("room-motion/living-room", motionPayload)

	// Simulate light message
	lightLevel := 85.5
	lightMsg := UnifiedSensorMessage{
		LightLevel: &lightLevel,
		LightState: "bright",
		Room:       "living-room",
		Sensor:     "PhotoTransistor",
		Timestamp:  time.Now().Unix(),
		DeviceID:   "pico-living-room",
	}
	lightPayload, _ := json.Marshal(lightMsg)
	service.handleLightMessage("room-light/living-room", lightPayload)

	// Give callbacks time to execute
	time.Sleep(100 * time.Millisecond)

	// Check that callbacks were called
	if !tempCallbackCalled {
		t.Error("Temperature callback was not called")
	}
	if !motionCallbackCalled {
		t.Error("Motion callback was not called")
	}
	if !lightCallbackCalled {
		t.Error("Light callback was not called")
	}

	// Test getting room sensor data
	roomData, exists := service.GetRoomSensorData("living-room")
	if !exists {
		t.Error("Room data should exist for living-room")
	}
	if roomData.Temperature != 72.5 {
		t.Errorf("Expected temperature 72.5, got %f", roomData.Temperature)
	}
	if !roomData.IsOccupied {
		t.Error("Expected room to be occupied")
	}
	if roomData.LightLevel != 85.5 {
		t.Errorf("Expected light level 85.5, got %f", roomData.LightLevel)
	}

	// Test sensor summary
	summary := service.GetSensorSummary()
	totalRooms := summary["total_rooms"].(int)
	if totalRooms != 1 {
		t.Errorf("Expected 1 room, got %d", totalRooms)
	}

	onlineDevices := summary["online_devices"].(int)
	if onlineDevices != 1 {
		t.Errorf("Expected 1 online device, got %d", onlineDevices)
	}
}

func TestExtractRoomID(t *testing.T) {
	service := &UnifiedSensorService{}

	tests := []struct {
		topic    string
		expected string
		hasError bool
	}{
		{"room-temp/living-room", "living-room", false},
		{"room-motion/bedroom", "bedroom", false},
		{"room-light/kitchen", "kitchen", false},
		{"invalid-topic", "", true},
		{"room-temp/bedroom/extra", "", true},
	}

	for _, test := range tests {
		roomID, err := service.extractRoomID(test.topic)
		if test.hasError {
			if err == nil {
				t.Errorf("Expected error for topic '%s', but got none", test.topic)
			}
		} else {
			if err != nil {
				t.Errorf("Unexpected error for topic '%s': %v", test.topic, err)
			}
			if roomID != test.expected {
				t.Errorf("Expected roomID '%s' for topic '%s', got '%s'", test.expected, test.topic, roomID)
			}
		}
	}
}

func TestDayNightCycle(t *testing.T) {
	service := &UnifiedSensorService{}

	tests := []struct {
		lightLevel float64
		expected   string
	}{
		// Note: Day/night cycle depends on both light level AND current time
		// During daytime testing, low light levels might be "transitional"
		// instead of "night" due to time-of-day logic
		{95.0, "day"},          // High light during day
		{50.0, "transitional"}, // Mid-level light
		{35.0, "transitional"}, // Mid-level light
		{25.0, "transitional"}, // Mid-level light
	}

	for _, test := range tests {
		result := service.determineDayNightCycle(test.lightLevel)
		if result != test.expected {
			t.Errorf("Expected '%s' for light level %.1f, got '%s'", test.expected, test.lightLevel, result)
		}
	}
}

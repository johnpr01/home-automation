package services

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"github.com/johnpr01/home-automation/internal/config"
	"github.com/johnpr01/home-automation/pkg/mqtt"
)

func TestNewLightService(t *testing.T) {
	logger := log.New(os.Stdout, "[TEST] ", log.LstdFlags)
	mqttConfig := &config.MQTTConfig{Broker: "localhost", Port: "1883"}
	mqttClient := mqtt.NewClient(mqttConfig, nil)

	service := NewLightService(mqttClient, logger)

	if service == nil {
		t.Fatal("Expected service to be created")
	}

	if service.roomLightLevels == nil {
		t.Error("Expected roomLightLevels map to be initialized")
	}

	if service.logger != logger {
		t.Error("Expected logger to be set")
	}
}

func TestAddLightCallback(t *testing.T) {
	logger := log.New(os.Stdout, "[TEST] ", log.LstdFlags)
	mqttConfig := &config.MQTTConfig{Broker: "localhost", Port: "1883"}
	mqttClient := mqtt.NewClient(mqttConfig, nil)

	service := NewLightService(mqttClient, logger)

	callbackCalled := false
	var callbackRoomID string
	var callbackLightState string
	var callbackLightLevel float64

	callback := func(roomID string, lightState string, lightLevel float64) {
		callbackCalled = true
		callbackRoomID = roomID
		callbackLightState = lightState
		callbackLightLevel = lightLevel
	}

	service.AddLightCallback(callback)

	// Simulate light sensor data
	lightMsg := LightSensorMessage{
		LightLevel: 85.5,
		LightState: "bright",
		Room:       "living-room",
		Sensor:     "PhotoTransistor",
		Timestamp:  time.Now().Unix(),
		DeviceID:   "pico-living-room",
	}

	payload, err := json.Marshal(lightMsg)
	if err != nil {
		t.Fatalf("Failed to marshal JSON: %v", err)
	}
	err := service.handleLightMessage("room-light/living-room", payload)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Give callback time to execute
	time.Sleep(100 * time.Millisecond)

	if !callbackCalled {
		t.Error("Expected callback to be called")
	}

	if callbackRoomID != "living-room" {
		t.Errorf("Expected callback roomID 'living-room', got '%s'", callbackRoomID)
	}

	if callbackLightState != "bright" {
		t.Errorf("Expected callback lightState 'bright', got '%s'", callbackLightState)
	}

	if callbackLightLevel != 85.5 {
		t.Errorf("Expected callback lightLevel 85.5, got %.1f", callbackLightLevel)
	}
}

func TestGetRoomLightLevel(t *testing.T) {
	logger := log.New(os.Stdout, "[TEST] ", log.LstdFlags)
	mqttConfig := &config.MQTTConfig{Broker: "localhost", Port: "1883"}
	mqttClient := mqtt.NewClient(mqttConfig, nil)

	service := NewLightService(mqttClient, logger)

	// Test non-existent room
	lightData, exists := service.GetRoomLightLevel("non-existent")
	if exists {
		t.Error("Expected non-existent room to not exist")
	}
	if lightData != nil {
		t.Error("Expected nil light data for non-existent room")
	}

	// Add light sensor data
	lightMsg := LightSensorMessage{
		LightLevel: 45.2,
		LightState: "moderate",
		Room:       "bedroom",
		Sensor:     "PhotoTransistor",
		Timestamp:  time.Now().Unix(),
		DeviceID:   "pico-bedroom",
	}

	payload, err := json.Marshal(lightMsg)
	if err != nil {
		t.Fatalf("Failed to marshal JSON: %v", err)
	}
	service.handleLightMessage("room-light/bedroom", payload)

	// Test existing room
	lightData, exists = service.GetRoomLightLevel("bedroom")
	if !exists {
		t.Error("Expected bedroom to exist after light sensor data")
	}

	if lightData == nil {
		t.Fatal("Expected non-nil light data")
	}

	if lightData.LightLevel != 45.2 {
		t.Errorf("Expected light level 45.2, got %.1f", lightData.LightLevel)
	}

	if lightData.LightState != "moderate" {
		t.Errorf("Expected light state 'moderate', got '%s'", lightData.LightState)
	}

	if lightData.RoomID != "bedroom" {
		t.Errorf("Expected room ID 'bedroom', got '%s'", lightData.RoomID)
	}
}

func TestGetAllLightLevels(t *testing.T) {
	logger := log.New(os.Stdout, "[TEST] ", log.LstdFlags)
	mqttConfig := &config.MQTTConfig{Broker: "localhost", Port: "1883"}
	mqttClient := mqtt.NewClient(mqttConfig, nil)

	service := NewLightService(mqttClient, logger)

	// Initially should be empty
	lightDataMap := service.GetAllLightLevels()
	if len(lightDataMap) != 0 {
		t.Errorf("Expected 0 light data entries, got %d", len(lightDataMap))
	}

	// Add light data for multiple rooms
	rooms := []struct {
		name       string
		lightLevel float64
		lightState string
	}{
		{"living-room", 85.0, "bright"},
		{"bedroom", 25.5, "dim"},
		{"kitchen", 60.2, "moderate"},
	}

	for _, room := range rooms {
		lightMsg := LightSensorMessage{
			LightLevel: room.lightLevel,
			LightState: room.lightState,
			Room:       room.name,
			Sensor:     "PhotoTransistor",
			Timestamp:  time.Now().Unix(),
			DeviceID:   "pico-" + room.name,
		}

		payload, err := json.Marshal(lightMsg)
		if err != nil {
			t.Fatalf("Failed to marshal JSON: %v", err)
		}
		service.handleLightMessage("room-light/"+room.name, payload)
	}

	lightDataMap = service.GetAllLightLevels()
	if len(lightDataMap) != 3 {
		t.Errorf("Expected 3 light data entries, got %d", len(lightDataMap))
	}

	// Check if all rooms are present
	for _, room := range rooms {
		if lightData, exists := lightDataMap[room.name]; !exists {
			t.Errorf("Expected to find room '%s' in light data", room.name)
		} else {
			if lightData.LightLevel != room.lightLevel {
				t.Errorf("Expected light level %.1f for room %s, got %.1f", room.lightLevel, room.name, lightData.LightLevel)
			}
			if lightData.LightState != room.lightState {
				t.Errorf("Expected light state '%s' for room %s, got '%s'", room.lightState, room.name, lightData.LightState)
			}
		}
	}
}

func TestHandleLightMessage(t *testing.T) {
	logger := log.New(os.Stdout, "[TEST] ", log.LstdFlags)
	mqttConfig := &config.MQTTConfig{Broker: "localhost", Port: "1883"}
	mqttClient := mqtt.NewClient(mqttConfig, nil)

	service := NewLightService(mqttClient, logger)

	// Test light sensor data
	lightMsg := LightSensorMessage{
		LightLevel: 92.8,
		LightState: "bright",
		Room:       "office",
		Sensor:     "PhotoTransistor",
		Timestamp:  time.Now().Unix(),
		DeviceID:   "pico-office",
	}

	payload, err := json.Marshal(lightMsg)
	if err != nil {
		t.Fatalf("Failed to marshal JSON: %v", err)
	}
	err := service.handleLightMessage("room-light/office", payload)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Check light data
	lightData, exists := service.GetRoomLightLevel("office")
	if !exists {
		t.Fatal("Expected light data to be created")
	}

	if lightData.LightLevel != 92.8 {
		t.Errorf("Expected light level 92.8, got %.1f", lightData.LightLevel)
	}

	if lightData.LightState != "bright" {
		t.Errorf("Expected light state 'bright', got '%s'", lightData.LightState)
	}

	if lightData.DeviceID != "pico-office" {
		t.Errorf("Expected device ID 'pico-office', got '%s'", lightData.DeviceID)
	}

	// Test updating with new light level
	lightMsg.LightLevel = 15.3
	lightMsg.LightState = "dim"
	payload, _ = json.Marshal(lightMsg)

	err = service.handleLightMessage("room-light/office", payload)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Check updated light data
	lightData, _ = service.GetRoomLightLevel("office")
	if lightData.LightLevel != 15.3 {
		t.Errorf("Expected updated light level 15.3, got %.1f", lightData.LightLevel)
	}

	if lightData.LightState != "dim" {
		t.Errorf("Expected updated light state 'dim', got '%s'", lightData.LightState)
	}
}

func TestLightServiceSummary(t *testing.T) {
	logger := log.New(os.Stdout, "[TEST] ", log.LstdFlags)
	mqttConfig := &config.MQTTConfig{Broker: "localhost", Port: "1883"}
	mqttClient := mqtt.NewClient(mqttConfig, nil)

	service := NewLightService(mqttClient, logger)

	// Add light data for multiple rooms
	rooms := []struct {
		name       string
		lightLevel float64
		lightState string
	}{
		{"living-room", 95.0, "bright"},
		{"bedroom", 5.2, "dark"},
		{"kitchen", 65.8, "moderate"},
		{"office", 30.1, "dim"},
	}

	for _, room := range rooms {
		lightMsg := LightSensorMessage{
			LightLevel: room.lightLevel,
			LightState: room.lightState,
			Room:       room.name,
			Sensor:     "PhotoTransistor",
			Timestamp:  time.Now().Unix(),
			DeviceID:   "pico-" + room.name,
		}

		payload, err := json.Marshal(lightMsg)
		if err != nil {
			t.Fatalf("Failed to marshal JSON: %v", err)
		}
		service.handleLightMessage("room-light/"+room.name, payload)
	}

	summary := service.GetLightSummary()

	totalRooms := summary["total_rooms"].(int)
	onlineDevices := summary["online_sensors"].(int)
	avgLightLevel := summary["average_light_level"].(float64)

	if totalRooms != 4 {
		t.Errorf("Expected 4 total rooms, got %d", totalRooms)
	}

	if onlineDevices != 4 {
		t.Errorf("Expected 4 online devices, got %d", onlineDevices)
	}

	// Calculate expected average: (95.0 + 5.2 + 65.8 + 30.1) / 4 = 49.025
	expectedAvg := 49.025
	if avgLightLevel < expectedAvg-0.1 || avgLightLevel > expectedAvg+0.1 {
		t.Errorf("Expected average light level around %.3f, got %.3f", expectedAvg, avgLightLevel)
	}

	roomInfo := summary["rooms"].([]map[string]interface{})
	if len(roomInfo) != 4 {
		t.Errorf("Expected 4 rooms in summary, got %d", len(roomInfo))
	}
}

func TestLightStates(t *testing.T) {
	logger := log.New(os.Stdout, "[TEST] ", log.LstdFlags)
	mqttConfig := &config.MQTTConfig{Broker: "localhost", Port: "1883"}
	mqttClient := mqtt.NewClient(mqttConfig, nil)

	service := NewLightService(mqttClient, logger)

	// Test different light states
	testCases := []struct {
		lightLevel    float64
		expectedState string
	}{
		{0.0, "dark"},
		{5.0, "dark"},
		{15.0, "dim"},
		{25.0, "dim"},
		{45.0, "moderate"},
		{65.0, "moderate"},
		{85.0, "bright"},
		{100.0, "bright"},
	}

	for i, testCase := range testCases {
		roomID := fmt.Sprintf("test-room-%d", i)
		lightMsg := LightSensorMessage{
			LightLevel: testCase.lightLevel,
			LightState: testCase.expectedState,
			Room:       roomID,
			Sensor:     "PhotoTransistor",
			Timestamp:  time.Now().Unix(),
			DeviceID:   "pico-" + roomID,
		}

		payload, err := json.Marshal(lightMsg)
		if err != nil {
			t.Fatalf("Failed to marshal JSON: %v", err)
		}
		service.handleLightMessage("room-light/"+roomID, payload)

		lightData, exists := service.GetRoomLightLevel(roomID)
		if !exists {
			t.Errorf("Expected light data to exist for room %s", roomID)
			continue
		}

		if lightData.LightState != testCase.expectedState {
			t.Errorf("For light level %.1f, expected state '%s', got '%s'",
				testCase.lightLevel, testCase.expectedState, lightData.LightState)
		}
	}
}

func TestInvalidLightMessage(t *testing.T) {
	logger := log.New(os.Stdout, "[TEST] ", log.LstdFlags)
	mqttConfig := &config.MQTTConfig{Broker: "localhost", Port: "1883"}
	mqttClient := mqtt.NewClient(mqttConfig, nil)

	service := NewLightService(mqttClient, logger)

	// Test invalid JSON
	invalidPayload := []byte(`{invalid json}`)
	err := service.handleLightMessage("room-light/test", invalidPayload)
	if err == nil {
		t.Error("Expected error for invalid JSON")
	}

	// Test invalid topic
	validPayload := []byte(`{"light_level": 50.0, "light_state": "moderate", "room": "test"}`)
	err = service.handleLightMessage("invalid-topic", validPayload)
	if err == nil {
		t.Error("Expected error for invalid topic")
	}
}

func TestConcurrentLightUpdates(t *testing.T) {
	logger := log.New(os.Stdout, "[TEST] ", log.LstdFlags)
	mqttConfig := &config.MQTTConfig{Broker: "localhost", Port: "1883"}
	mqttClient := mqtt.NewClient(mqttConfig, nil)

	service := NewLightService(mqttClient, logger)

	// Test concurrent light updates
	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func(i int) {
			defer func() { done <- true }()

			roomID := fmt.Sprintf("concurrent-room-%d", i)
			for j := 0; j < 100; j++ {
				lightLevel := float64(j % 101) // 0-100% light level
				var lightState string
				if lightLevel < 10 {
					lightState = "dark"
				} else if lightLevel < 30 {
					lightState = "dim"
				} else if lightLevel < 70 {
					lightState = "moderate"
				} else {
					lightState = "bright"
				}

				lightMsg := LightSensorMessage{
					LightLevel: lightLevel,
					LightState: lightState,
					Room:       roomID,
					Sensor:     "PhotoTransistor",
					Timestamp:  time.Now().Unix(),
					DeviceID:   "pico-" + roomID,
				}

				payload, err := json.Marshal(lightMsg)
				if err != nil {
					t.Fatalf("Failed to marshal JSON: %v", err)
				}
				service.handleLightMessage("room-light/"+roomID, payload)
			}
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	// Check if all rooms were tracked
	lightDataMap := service.GetAllLightLevels()
	if len(lightDataMap) != 10 {
		t.Errorf("Expected 10 rooms, got %d", len(lightDataMap))
	}
}

func TestDeviceOnlineStatusLight(t *testing.T) {
	logger := log.New(os.Stdout, "[TEST] ", log.LstdFlags)
	mqttConfig := &config.MQTTConfig{Broker: "localhost", Port: "1883"}
	mqttClient := mqtt.NewClient(mqttConfig, nil)

	service := NewLightService(mqttClient, logger)

	// Add light sensor data
	lightMsg := LightSensorMessage{
		LightLevel: 75.0,
		LightState: "bright",
		Room:       "online-test",
		Sensor:     "PhotoTransistor",
		Timestamp:  time.Now().Unix(),
		DeviceID:   "pico-online-test",
	}

	payload, err := json.Marshal(lightMsg)
	if err != nil {
		t.Fatalf("Failed to marshal JSON: %v", err)
	}
	service.handleLightMessage("room-light/online-test", payload)

	// Check device is marked online
	lightData, exists := service.GetRoomLightLevel("online-test")
	if !exists {
		t.Fatal("Expected light data to exist")
	}

	if !lightData.IsOnline {
		t.Error("Expected device to be marked online after recent message")
	}
}

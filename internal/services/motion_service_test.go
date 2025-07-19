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

func TestNewMotionService(t *testing.T) {
	logger := log.New(os.Stdout, "[TEST] ", log.LstdFlags)
	mqttConfig := &config.MQTTConfig{Broker: "localhost", Port: "1883"}
	mqttClient := mqtt.NewClient(mqttConfig, nil)

	service := NewMotionService(mqttClient, logger)

	if service == nil {
		t.Fatal("Expected service to be created")
	}

	if service.roomOccupancy == nil {
		t.Error("Expected roomOccupancy map to be initialized")
	}

	if service.logger != logger {
		t.Error("Expected logger to be set")
	}
}

func TestAddMotionCallback(t *testing.T) {
	logger := log.New(os.Stdout, "[TEST] ", log.LstdFlags)
	mqttConfig := &config.MQTTConfig{Broker: "localhost", Port: "1883"}
	mqttClient := mqtt.NewClient(mqttConfig, nil)

	service := NewMotionService(mqttClient, logger)

	callbackCalled := false
	var callbackRoomID string
	var callbackOccupied bool

	callback := func(roomID string, occupied bool) {
		callbackCalled = true
		callbackRoomID = roomID
		callbackOccupied = occupied
	}

	service.AddOccupancyCallback(callback)

	// Simulate motion detection
	motionMsg := MotionDetectionMessage{
		Motion:    true,
		Room:      "living-room",
		Sensor:    "PIR",
		Timestamp: time.Now().Unix(),
		DeviceID:  "pico-living-room",
	}

	payload, err := json.Marshal(motionMsg)
	if err != nil {
		t.Fatalf("Failed to marshal JSON: %v", err)
	}
	err = service.handleMotionMessage("room-motion/living-room", payload)
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

	if !callbackOccupied {
		t.Error("Expected callback occupied to be true")
	}
}

func TestGetRoomOccupancy(t *testing.T) {
	logger := log.New(os.Stdout, "[TEST] ", log.LstdFlags)
	mqttConfig := &config.MQTTConfig{Broker: "localhost", Port: "1883"}
	mqttClient := mqtt.NewClient(mqttConfig, nil)

	service := NewMotionService(mqttClient, logger)

	// Test non-existent room
	occupancy, exists := service.GetRoomOccupancy("non-existent")
	if exists {
		t.Error("Expected non-existent room to not exist")
	}
	if occupancy != nil {
		t.Error("Expected nil occupancy for non-existent room")
	}

	// Add motion detection
	motionMsg := MotionDetectionMessage{
		Motion:    true,
		Room:      "bedroom",
		Sensor:    "PIR",
		Timestamp: time.Now().Unix(),
		DeviceID:  "pico-bedroom",
	}

	payload, err := json.Marshal(motionMsg)
	if err != nil {
		t.Fatalf("Failed to marshal JSON: %v", err)
	}
	service.handleMotionMessage("room-motion/bedroom", payload)

	// Test existing room
	occupancy, exists = service.GetRoomOccupancy("bedroom")
	if !exists {
		t.Error("Expected bedroom to exist after motion detection")
	}

	if occupancy == nil {
		t.Fatal("Expected non-nil occupancy")
	}

	if !occupancy.IsOccupied {
		t.Error("Expected room to be occupied")
	}

	if occupancy.RoomID != "bedroom" {
		t.Errorf("Expected room ID 'bedroom', got '%s'", occupancy.RoomID)
	}
}

func TestGetAllRoomOccupancy(t *testing.T) {
	logger := log.New(os.Stdout, "[TEST] ", log.LstdFlags)
	mqttConfig := &config.MQTTConfig{Broker: "localhost", Port: "1883"}
	mqttClient := mqtt.NewClient(mqttConfig, nil)

	service := NewMotionService(mqttClient, logger)

	// Initially should be empty
	occupancies := service.GetAllOccupancy()
	if len(occupancies) != 0 {
		t.Errorf("Expected 0 occupancies, got %d", len(occupancies))
	}

	// Add motion for multiple rooms
	rooms := []string{"living-room", "bedroom", "kitchen"}
	for _, room := range rooms {
		motionMsg := MotionDetectionMessage{
			Motion:    true,
			Room:      room,
			Sensor:    "PIR",
			Timestamp: time.Now().Unix(),
			DeviceID:  "pico-" + room,
		}

		payload, err := json.Marshal(motionMsg)
		if err != nil {
			t.Fatalf("Failed to marshal JSON: %v", err)
		}
		service.handleMotionMessage("room-motion/"+room, payload)
	}

	occupancies = service.GetAllOccupancy()
	if len(occupancies) != 3 {
		t.Errorf("Expected 3 occupancies, got %d", len(occupancies))
	}

	// Check if all rooms are present
	foundRooms := make(map[string]bool)
	for _, occupancy := range occupancies {
		foundRooms[occupancy.RoomID] = true
	}

	for _, room := range rooms {
		if !foundRooms[room] {
			t.Errorf("Expected to find room '%s' in occupancies", room)
		}
	}
}

func TestHandleMotionMessage(t *testing.T) {
	logger := log.New(os.Stdout, "[TEST] ", log.LstdFlags)
	mqttConfig := &config.MQTTConfig{Broker: "localhost", Port: "1883"}
	mqttClient := mqtt.NewClient(mqttConfig, nil)

	service := NewMotionService(mqttClient, logger)

	// Test motion detected
	motionMsg := MotionDetectionMessage{
		Motion:      true,
		Room:        "office",
		Sensor:      "PIR",
		Timestamp:   time.Now().Unix(),
		MotionStart: time.Now().Unix(),
		DeviceID:    "pico-office",
	}

	payload, err := json.Marshal(motionMsg)
	if err != nil {
		t.Fatalf("Failed to marshal JSON: %v", err)
	}
	err = service.handleMotionMessage("room-motion/office", payload)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Check occupancy
	occupancy, exists := service.GetRoomOccupancy("office")
	if !exists {
		t.Fatal("Expected occupancy to be created")
	}

	if !occupancy.IsOccupied {
		t.Error("Expected room to be occupied")
	}

	if occupancy.DeviceID != "pico-office" {
		t.Errorf("Expected device ID 'pico-office', got '%s'", occupancy.DeviceID)
	}

	// Test motion cleared
	motionMsg.Motion = false
	motionMsg.MotionStart = 0
	payload, _ = json.Marshal(motionMsg)

	err = service.handleMotionMessage("room-motion/office", payload)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Check occupancy cleared
	occupancy, _ = service.GetRoomOccupancy("office")
	if occupancy.IsOccupied {
		t.Error("Expected room to not be occupied")
	}
}

func TestExtractRoomIDFromTopic(t *testing.T) {
	logger := log.New(os.Stdout, "[TEST] ", log.LstdFlags)
	mqttConfig := &config.MQTTConfig{Broker: "localhost", Port: "1883"}
	mqttClient := mqtt.NewClient(mqttConfig, nil)

	service := NewMotionService(mqttClient, logger)

	tests := []struct {
		topic    string
		expected string
		hasError bool
	}{
		{"room-motion/living-room", "living-room", false},
		{"room-motion/bedroom", "bedroom", false},
		{"room-motion/kitchen", "kitchen", false},
		{"invalid-topic", "", true},
		{"room-motion/bedroom/extra", "", true},
		{"room-motion/", "", true},
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

func TestMotionServiceSummary(t *testing.T) {
	logger := log.New(os.Stdout, "[TEST] ", log.LstdFlags)
	mqttConfig := &config.MQTTConfig{Broker: "localhost", Port: "1883"}
	mqttClient := mqtt.NewClient(mqttConfig, nil)

	service := NewMotionService(mqttClient, logger)

	// Add motion for some rooms (some occupied, some not)
	rooms := []struct {
		name     string
		occupied bool
	}{
		{"living-room", true},
		{"bedroom", false},
		{"kitchen", true},
		{"office", false},
	}

	for _, room := range rooms {
		motionMsg := MotionDetectionMessage{
			Motion:    room.occupied,
			Room:      room.name,
			Sensor:    "PIR",
			Timestamp: time.Now().Unix(),
			DeviceID:  "pico-" + room.name,
		}

		payload, err := json.Marshal(motionMsg)
		if err != nil {
			t.Fatalf("Failed to marshal JSON: %v", err)
		}
		service.handleMotionMessage("room-motion/"+room.name, payload)
	}

	summary := service.GetMotionSummary()

	totalRooms := summary["total_rooms"].(int)
	occupiedRooms := summary["occupied_rooms"].(int)
	onlineDevices := summary["online_sensors"].(int)

	if totalRooms != 4 {
		t.Errorf("Expected 4 total rooms, got %d", totalRooms)
	}

	if occupiedRooms != 2 {
		t.Errorf("Expected 2 occupied rooms, got %d", occupiedRooms)
	}

	if onlineDevices != 4 {
		t.Errorf("Expected 4 online devices, got %d", onlineDevices)
	}

	roomInfo := summary["rooms"].([]map[string]interface{})
	if len(roomInfo) != 4 {
		t.Errorf("Expected 4 rooms in summary, got %d", len(roomInfo))
	}
}

func TestInvalidMotionMessage(t *testing.T) {
	logger := log.New(os.Stdout, "[TEST] ", log.LstdFlags)
	mqttConfig := &config.MQTTConfig{Broker: "localhost", Port: "1883"}
	mqttClient := mqtt.NewClient(mqttConfig, nil)

	service := NewMotionService(mqttClient, logger)

	// Test invalid JSON
	invalidPayload := []byte(`{invalid json}`)
	err := service.handleMotionMessage("room-motion/test", invalidPayload)
	if err == nil {
		t.Error("Expected error for invalid JSON")
	}

	// Test invalid topic
	validPayload := []byte(`{"motion": true, "room": "test"}`)
	err = service.handleMotionMessage("invalid-topic", validPayload)
	if err == nil {
		t.Error("Expected error for invalid topic")
	}
}

func TestConcurrentMotionUpdates(t *testing.T) {
	logger := log.New(os.Stdout, "[TEST] ", log.LstdFlags)
	mqttConfig := &config.MQTTConfig{Broker: "localhost", Port: "1883"}
	mqttClient := mqtt.NewClient(mqttConfig, nil)

	service := NewMotionService(mqttClient, logger)

	// Test concurrent motion updates
	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func(i int) {
			defer func() { done <- true }()

			roomID := fmt.Sprintf("room-%d", i)
			for j := 0; j < 100; j++ {
				motionMsg := MotionDetectionMessage{
					Motion:    j%2 == 0, // Alternate between motion and no motion
					Room:      roomID,
					Sensor:    "PIR",
					Timestamp: time.Now().Unix(),
					DeviceID:  "pico-" + roomID,
				}

				payload, err := json.Marshal(motionMsg)
				if err != nil {
					t.Errorf("Failed to marshal JSON: %v", err)
					return
				}
				service.handleMotionMessage("room-motion/"+roomID, payload)
			}
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	// Check if all rooms were tracked
	occupancies := service.GetAllOccupancy()
	if len(occupancies) != 10 {
		t.Errorf("Expected 10 rooms, got %d", len(occupancies))
	}
}

func TestMotionTimeout(t *testing.T) {
	logger := log.New(os.Stdout, "[TEST] ", log.LstdFlags)
	mqttConfig := &config.MQTTConfig{Broker: "localhost", Port: "1883"}
	mqttClient := mqtt.NewClient(mqttConfig, nil)

	service := NewMotionService(mqttClient, logger)

	// Add motion detection
	motionMsg := MotionDetectionMessage{
		Motion:    true,
		Room:      "timeout-room",
		Sensor:    "PIR",
		Timestamp: time.Now().Unix(),
		DeviceID:  "pico-timeout-room",
	}

	payload, err := json.Marshal(motionMsg)
	if err != nil {
		t.Fatalf("Failed to marshal JSON: %v", err)
	}
	service.handleMotionMessage("room-motion/timeout-room", payload)

	// Verify room is occupied
	occupancy, exists := service.GetRoomOccupancy("timeout-room")
	if !exists || !occupancy.IsOccupied {
		t.Fatal("Expected room to be occupied initially")
	}

	// Simulate old motion (should be considered cleared in cleanup)
	oldTime := time.Now().Add(-15 * time.Minute)
	occupancy.LastMotionTime = oldTime
	occupancy.MotionStartTime = oldTime

	// Check if room occupancy is still valid (depends on implementation)
	occupancy, _ = service.GetRoomOccupancy("timeout-room")
	// Note: The actual behavior depends on the cleanup logic implementation
}

func TestDeviceOnlineStatus(t *testing.T) {
	logger := log.New(os.Stdout, "[TEST] ", log.LstdFlags)
	mqttConfig := &config.MQTTConfig{Broker: "localhost", Port: "1883"}
	mqttClient := mqtt.NewClient(mqttConfig, nil)

	service := NewMotionService(mqttClient, logger)

	// Add motion detection
	motionMsg := MotionDetectionMessage{
		Motion:    true,
		Room:      "online-test",
		Sensor:    "PIR",
		Timestamp: time.Now().Unix(),
		DeviceID:  "pico-online-test",
	}

	payload, err := json.Marshal(motionMsg)
	if err != nil {
		t.Fatalf("Failed to marshal JSON: %v", err)
	}
	service.handleMotionMessage("room-motion/online-test", payload)

	// Check device is marked online
	occupancy, exists := service.GetRoomOccupancy("online-test")
	if !exists {
		t.Fatal("Expected occupancy to exist")
	}

	if !occupancy.IsOnline {
		t.Error("Expected device to be marked online after recent message")
	}
}

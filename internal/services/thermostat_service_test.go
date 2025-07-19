package services

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/johnpr01/home-automation/internal/config"
	"github.com/johnpr01/home-automation/internal/logger"
	"github.com/johnpr01/home-automation/internal/models"
	"github.com/johnpr01/home-automation/pkg/mqtt"
)

// MockMQTTClient for testing
type MockMQTTClient struct {
	published map[string][]byte
	handlers  map[string]mqtt.MessageHandler
}

func NewMockMQTTClient() *MockMQTTClient {
	return &MockMQTTClient{
		published: make(map[string][]byte),
		handlers:  make(map[string]mqtt.MessageHandler),
	}
}

func (m *MockMQTTClient) Subscribe(topic string, handler mqtt.MessageHandler) error {
	m.handlers[topic] = handler
	return nil
}

func (m *MockMQTTClient) Publish(msg *mqtt.Message) error {
	m.published[msg.Topic] = msg.Payload
	return nil
}

func (m *MockMQTTClient) Connect() error    { return nil }
func (m *MockMQTTClient) Disconnect() error { return nil }

// SimulateMessage simulates receiving an MQTT message
func (m *MockMQTTClient) SimulateMessage(topic string, payload []byte) error {
	if handler, exists := m.handlers[topic]; exists {
		return handler(topic, payload)
	}
	return nil
}

func TestNewThermostatService(t *testing.T) {
	testLogger := logger.NewLogger("thermostat-test", nil)

	// Create MQTT client wrapper
	mqttConfig := &config.MQTTConfig{Broker: "localhost", Port: "1883"}
	mqttClient := mqtt.NewClient(mqttConfig, nil)

	service := NewThermostatService(mqttClient, testLogger)

	if service == nil {
		t.Fatal("Expected service to be created")
	}

	if service.thermostats == nil {
		t.Error("Expected thermostats map to be initialized")
	}

	if service.logger != testLogger {
		t.Error("Expected logger to be set")
	}
}

func TestHandleTemperatureUpdate(t *testing.T) {
	testLogger := logger.NewLogger("thermostat-test", nil)
	mqttConfig := &config.MQTTConfig{Broker: "localhost", Port: "1883"}
	mqttClient := mqtt.NewClient(mqttConfig, nil)

	service := NewThermostatService(mqttClient, testLogger)

	// Test creating new thermostat for room
	roomID := "living-room"
	temperature := 72.5

	service.HandleTemperatureUpdate(roomID, temperature)

	// Check if thermostat was created
	thermostat, err := service.GetThermostat(roomID)
	if err != nil {
		t.Fatal("Expected thermostat to be created for room")
	}

	if thermostat.CurrentTemp != temperature {
		t.Errorf("Expected current temp %.1f, got %.1f", temperature, thermostat.CurrentTemp)
	}

	if thermostat.TargetTemp != 72.0 {
		t.Errorf("Expected default target temp 72.0, got %.1f", thermostat.TargetTemp)
	}

	if thermostat.Mode != models.ModeAuto {
		t.Errorf("Expected mode Auto, got %s", thermostat.Mode)
	}

	// Test updating existing thermostat
	newTemp := 74.2
	service.HandleTemperatureUpdate(roomID, newTemp)

	thermostat, _ = service.GetThermostat(roomID)
	if thermostat.CurrentTemp != newTemp {
		t.Errorf("Expected updated temp %.1f, got %.1f", newTemp, thermostat.CurrentTemp)
	}
}

func TestRegisterThermostat(t *testing.T) {
	testLogger := logger.NewLogger("thermostat-test", nil)
	mqttConfig := &config.MQTTConfig{Broker: "localhost", Port: "1883"}
	mqttClient := mqtt.NewClient(mqttConfig, nil)

	service := NewThermostatService(mqttClient, testLogger)

	thermostat := &models.Thermostat{
		ID:         "test-thermostat",
		Name:       "Test Thermostat",
		RoomID:     "bedroom",
		TargetTemp: 68.0,
		Mode:       models.ModeHeat,
	}

	service.RegisterThermostat(thermostat)

	// Check if thermostat was registered
	retrieved, err := service.GetThermostat("test-thermostat")
	if err != nil {
		t.Fatal("Expected thermostat to be registered")
	}

	if retrieved.ID != thermostat.ID {
		t.Errorf("Expected ID %s, got %s", thermostat.ID, retrieved.ID)
	}

	// Check default values were set
	if retrieved.Hysteresis == 0 {
		t.Error("Expected hysteresis to be set to default value")
	}

	if retrieved.MinTemp == 0 {
		t.Error("Expected MinTemp to be set to default value")
	}

	if retrieved.MaxTemp == 0 {
		t.Error("Expected MaxTemp to be set to default value")
	}
}

func TestSetTargetTemperature(t *testing.T) {
	testLogger := logger.NewLogger("thermostat-test", nil)
	mqttConfig := &config.MQTTConfig{Broker: "localhost", Port: "1883"}
	mqttClient := mqtt.NewClient(mqttConfig, nil)

	service := NewThermostatService(mqttClient, testLogger)

	// Register a thermostat first
	thermostat := &models.Thermostat{
		ID:         "test-thermostat",
		TargetTemp: 70.0,
		MinTemp:    60.0,
		MaxTemp:    85.0,
	}
	service.RegisterThermostat(thermostat)

	// Test valid temperature
	err := service.SetTargetTemperature("test-thermostat", 75.0)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	retrieved, _ := service.GetThermostat("test-thermostat")
	if retrieved.TargetTemp != 75.0 {
		t.Errorf("Expected target temp 75.0, got %.1f", retrieved.TargetTemp)
	}

	// Test temperature too low
	err = service.SetTargetTemperature("test-thermostat", 50.0)
	if err == nil {
		t.Error("Expected error for temperature below minimum")
	}

	// Test temperature too high
	err = service.SetTargetTemperature("test-thermostat", 95.0)
	if err == nil {
		t.Error("Expected error for temperature above maximum")
	}

	// Test non-existent thermostat
	err = service.SetTargetTemperature("non-existent", 72.0)
	if err == nil {
		t.Error("Expected error for non-existent thermostat")
	}
}

func TestSetMode(t *testing.T) {
	testLogger := logger.NewLogger("thermostat-test", nil)
	mqttConfig := &config.MQTTConfig{Broker: "localhost", Port: "1883"}
	mqttClient := mqtt.NewClient(mqttConfig, nil)

	service := NewThermostatService(mqttClient, testLogger)

	// Register a thermostat first
	thermostat := &models.Thermostat{
		ID:   "test-thermostat",
		Mode: models.ModeAuto,
	}
	service.RegisterThermostat(thermostat)

	// Test valid mode changes
	validModes := []models.ThermostatMode{
		models.ModeOff,
		models.ModeHeat,
		models.ModeCool,
		models.ModeAuto,
		models.ModeFan,
	}

	for _, mode := range validModes {
		err := service.SetMode("test-thermostat", mode)
		if err != nil {
			t.Errorf("Unexpected error for mode %s: %v", mode, err)
		}

		retrieved, _ := service.GetThermostat("test-thermostat")
		if retrieved.Mode != mode {
			t.Errorf("Expected mode %s, got %s", mode, retrieved.Mode)
		}
	}

	// Test non-existent thermostat
	err := service.SetMode("non-existent", models.ModeHeat)
	if err == nil {
		t.Error("Expected error for non-existent thermostat")
	}
}

func TestGetAllThermostats(t *testing.T) {
	testLogger := logger.NewLogger("thermostat-test", nil)
	mqttConfig := &config.MQTTConfig{Broker: "localhost", Port: "1883"}
	mqttClient := mqtt.NewClient(mqttConfig, nil)

	service := NewThermostatService(mqttClient, testLogger)

	// Initially should be empty
	thermostats := service.GetAllThermostats()
	if len(thermostats) != 0 {
		t.Errorf("Expected 0 thermostats, got %d", len(thermostats))
	}

	// Register some thermostats
	thermostat1 := &models.Thermostat{ID: "thermostat-1"}
	thermostat2 := &models.Thermostat{ID: "thermostat-2"}

	service.RegisterThermostat(thermostat1)
	service.RegisterThermostat(thermostat2)

	thermostats = service.GetAllThermostats()
	if len(thermostats) != 2 {
		t.Errorf("Expected 2 thermostats, got %d", len(thermostats))
	}

	// Check if both thermostats are returned
	foundIDs := make(map[string]bool)
	for _, t := range thermostats {
		foundIDs[t.ID] = true
	}

	if !foundIDs["thermostat-1"] || !foundIDs["thermostat-2"] {
		t.Error("Expected to find both registered thermostats")
	}
}

func TestHandleTemperatureMessage(t *testing.T) {
	testLogger := logger.NewLogger("thermostat-test", nil)
	mqttConfig := &config.MQTTConfig{Broker: "localhost", Port: "1883"}
	mqttClient := mqtt.NewClient(mqttConfig, nil)

	service := NewThermostatService(mqttClient, testLogger)

	// Create a temperature message
	tempData := map[string]interface{}{
		"temperature": 73.5,
		"room":        "kitchen",
		"sensor":      "SHT30",
		"timestamp":   time.Now().Unix(),
		"unit":        "F",
	}

	payload, err := json.Marshal(tempData)
	if err != nil {
		t.Fatalf("Failed to marshal temperature message: %v", err)
	}

	// Test handling the message
	err = service.handleTemperatureMessage("room-temp/kitchen", payload)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Check if thermostat was created/updated
	thermostat, err := service.GetThermostat("kitchen")
	if err != nil {
		t.Fatal("Expected thermostat to be created for kitchen")
	}

	if thermostat.CurrentTemp != 73.5 {
		t.Errorf("Expected temperature 73.5, got %.1f", thermostat.CurrentTemp)
	}
}

func TestProcessThermostat(t *testing.T) {
	testLogger := logger.NewLogger("thermostat-test", nil)
	mqttConfig := &config.MQTTConfig{Broker: "localhost", Port: "1883"}
	mqttClient := mqtt.NewClient(mqttConfig, nil)

	service := NewThermostatService(mqttClient, testLogger)

	// Test heating scenario
	thermostat := &models.Thermostat{
		ID:             "test-thermostat",
		CurrentTemp:    68.0, // Below target
		TargetTemp:     72.0,
		Hysteresis:     1.0,
		Mode:           models.ModeAuto,
		Status:         models.StatusIdle,
		HeatingEnabled: true,
		CoolingEnabled: true,
	}

	service.RegisterThermostat(thermostat)

	// Process the thermostat
	service.processThermostat(thermostat)

	// Check if status changed to heating
	if thermostat.Status != models.StatusHeating {
		t.Errorf("Expected status heating, got %s", thermostat.Status)
	}

	// Test cooling scenario
	thermostat.CurrentTemp = 74.0 // Above target + hysteresis
	thermostat.Status = models.StatusIdle

	service.processThermostat(thermostat)

	// Check if status changed to cooling
	if thermostat.Status != models.StatusCooling {
		t.Errorf("Expected status cooling, got %s", thermostat.Status)
	}

	// Test within range (should stay idle)
	thermostat.CurrentTemp = 72.0 // At target
	thermostat.Status = models.StatusIdle

	service.processThermostat(thermostat)

	// Check if status stayed idle
	if thermostat.Status != models.StatusIdle {
		t.Errorf("Expected status idle, got %s", thermostat.Status)
	}
}

func TestGetRoomTemperature(t *testing.T) {
	testLogger := logger.NewLogger("thermostat-test", nil)
	mqttConfig := &config.MQTTConfig{Broker: "localhost", Port: "1883"}
	mqttClient := mqtt.NewClient(mqttConfig, nil)

	service := NewThermostatService(mqttClient, testLogger)

	// Test room with no thermostat
	_, err := service.GetRoomTemperature("non-existent")
	if err == nil {
		t.Error("Expected error for non-existent room")
	}

	// Register a thermostat
	thermostat := &models.Thermostat{
		ID:          "test-thermostat",
		RoomID:      "living-room",
		CurrentTemp: 71.5,
	}
	service.RegisterThermostat(thermostat)

	// Test getting temperature
	temp, err := service.GetRoomTemperature("living-room")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if temp != 71.5 {
		t.Errorf("Expected temperature 71.5, got %.1f", temp)
	}
}

func TestConcurrentAccess(t *testing.T) {
	testLogger := logger.NewLogger("thermostat-test", nil)
	mqttConfig := &config.MQTTConfig{Broker: "localhost", Port: "1883"}
	mqttClient := mqtt.NewClient(mqttConfig, nil)

	service := NewThermostatService(mqttClient, testLogger)

	// Test concurrent temperature updates
	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func(i int) {
			defer func() { done <- true }()
			roomID := fmt.Sprintf("room-%d", i)
			for j := 0; j < 100; j++ {
				service.HandleTemperatureUpdate(roomID, float64(70+j%10))
			}
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	// Check if all thermostats were created
	thermostats := service.GetAllThermostats()
	if len(thermostats) != 10 {
		t.Errorf("Expected 10 thermostats, got %d", len(thermostats))
	}
}

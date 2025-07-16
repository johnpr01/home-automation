package services

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/johnpr01/home-automation/internal/models"
	"github.com/johnpr01/home-automation/pkg/mqtt"
	"github.com/johnpr01/home-automation/pkg/utils"
)

// ThermostatService manages smart thermostats and processes sensor data
type ThermostatService struct {
	thermostats map[string]*models.Thermostat
	mqttClient  *mqtt.Client
	mu          sync.RWMutex
	logger      *log.Logger
}

// NewThermostatService creates a new thermostat service
func NewThermostatService(mqttClient *mqtt.Client, logger *log.Logger) *ThermostatService {
	service := &ThermostatService{
		thermostats: make(map[string]*models.Thermostat),
		mqttClient:  mqttClient,
		logger:      logger,
	}

	// Subscribe to sensor topics
	service.subscribeSensorTopics()

	// Start control loop
	go service.controlLoop()

	return service
}

// RegisterThermostat registers a new thermostat
func (ts *ThermostatService) RegisterThermostat(thermostat *models.Thermostat) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	// Set default values in Fahrenheit
	if thermostat.Hysteresis == 0 {
		thermostat.Hysteresis = utils.DefaultHysteresis // 1°F default hysteresis
	}
	if thermostat.MinTemp == 0 {
		thermostat.MinTemp = utils.DefaultMinTemp // 50°F minimum
	}
	if thermostat.MaxTemp == 0 {
		thermostat.MaxTemp = utils.DefaultMaxTemp // 95°F maximum
	}
	if thermostat.TargetTemp == 0 {
		thermostat.TargetTemp = utils.DefaultTargetTemp // 70°F default target
	}

	ts.thermostats[thermostat.ID] = thermostat
	ts.logger.Printf("Registered thermostat: %s in room %s", thermostat.ID, thermostat.RoomID)
}

// GetThermostat retrieves a thermostat by ID
func (ts *ThermostatService) GetThermostat(id string) (*models.Thermostat, error) {
	ts.mu.RLock()
	defer ts.mu.RUnlock()

	thermostat, exists := ts.thermostats[id]
	if !exists {
		return nil, fmt.Errorf("thermostat not found: %s", id)
	}

	return thermostat, nil
}

// GetAllThermostats returns all registered thermostats
func (ts *ThermostatService) GetAllThermostats() []*models.Thermostat {
	ts.mu.RLock()
	defer ts.mu.RUnlock()

	thermostats := make([]*models.Thermostat, 0, len(ts.thermostats))
	for _, t := range ts.thermostats {
		thermostats = append(thermostats, t)
	}

	return thermostats
}

// SetTargetTemperature sets the target temperature for a thermostat
func (ts *ThermostatService) SetTargetTemperature(id string, temp float64) error {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	thermostat, exists := ts.thermostats[id]
	if !exists {
		return fmt.Errorf("thermostat not found: %s", id)
	}

	if !thermostat.IsValidTargetTemp(temp) {
		return fmt.Errorf("invalid target temperature: %.1f (range: %.1f-%.1f)",
			temp, thermostat.MinTemp, thermostat.MaxTemp)
	}

	thermostat.TargetTemp = temp
	thermostat.UpdatedAt = time.Now()

	ts.logger.Printf("Set target temperature for %s to %.1f°C", id, temp)

	// Publish command to MQTT
	ts.publishThermostatCommand(id, models.CmdSetTargetTemp, temp)

	return nil
}

// SetMode sets the operating mode for a thermostat
func (ts *ThermostatService) SetMode(id string, mode models.ThermostatMode) error {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	thermostat, exists := ts.thermostats[id]
	if !exists {
		return fmt.Errorf("thermostat not found: %s", id)
	}

	if !thermostat.IsValidMode(mode) {
		return fmt.Errorf("invalid mode: %s", mode)
	}

	thermostat.Mode = mode
	thermostat.UpdatedAt = time.Now()

	ts.logger.Printf("Set mode for %s to %s", id, mode)

	// Publish command to MQTT
	ts.publishThermostatCommand(id, models.CmdSetMode, string(mode))

	return nil
}

// subscribeSensorTopics subscribes to MQTT topics for sensor data
func (ts *ThermostatService) subscribeSensorTopics() {
	// Subscribe to temperature topics from Pi Pico sensors
	ts.mqttClient.Subscribe("room-temp/+", ts.handleTemperatureMessage)
	ts.mqttClient.Subscribe("room-hum/+", ts.handleHumidityMessage)

	ts.logger.Println("Subscribed to sensor MQTT topics")
}

// handleTemperatureMessage processes temperature messages from Pi Pico sensors
func (ts *ThermostatService) handleTemperatureMessage(topic string, payload []byte) error {
	// Extract room number from topic (room-temp/1)
	parts := strings.Split(topic, "/")
	if len(parts) != 2 {
		ts.logger.Printf("Invalid temperature topic format: %s", topic)
		return fmt.Errorf("invalid topic format: %s", topic)
	}

	roomID := parts[1]

	// Parse JSON payload
	var sensorData map[string]interface{}
	if err := json.Unmarshal(payload, &sensorData); err != nil {
		ts.logger.Printf("Failed to parse temperature message: %v", err)
		return err
	}

	// Convert to SensorReading
	reading := models.SensorReading{
		SensorID:  fmt.Sprintf("pico-%s", roomID),
		Value:     sensorData["temperature"],
		Timestamp: time.Now(),
	}

	// Find thermostat for this room
	ts.mu.Lock()
	defer ts.mu.Unlock()

	for _, thermostat := range ts.thermostats {
		if thermostat.RoomID == roomID {
			oldTemp := thermostat.CurrentTemp
			// Extract temperature value (now in Fahrenheit from Pi Pico)
			if tempFahrenheit, ok := reading.Value.(float64); ok {
				thermostat.CurrentTemp = tempFahrenheit + thermostat.TemperatureOffset
				thermostat.LastSensorUpdate = time.Now()
				thermostat.IsOnline = true
				thermostat.UpdatedAt = time.Now()
			}

			ts.logger.Printf("Updated thermostat %s: %.1f°F -> %.1f°F",
				thermostat.ID, oldTemp, thermostat.CurrentTemp)
			break
		}
	}

	return nil
}

// handleHumidityMessage processes humidity messages from Pi Pico sensors
func (ts *ThermostatService) handleHumidityMessage(topic string, payload []byte) error {
	// Extract room number from topic (room-hum/1)
	parts := strings.Split(topic, "/")
	if len(parts) != 2 {
		ts.logger.Printf("Invalid humidity topic format: %s", topic)
		return fmt.Errorf("invalid topic format: %s", topic)
	}

	roomID := parts[1]

	// Parse JSON payload
	var sensorData map[string]interface{}
	if err := json.Unmarshal(payload, &sensorData); err != nil {
		ts.logger.Printf("Failed to parse humidity message: %v", err)
		return err
	}

	// Find thermostat for this room and update humidity
	ts.mu.Lock()
	defer ts.mu.Unlock()

	for _, thermostat := range ts.thermostats {
		if thermostat.RoomID == roomID {
			// Extract humidity value
			if humidity, ok := sensorData["humidity"].(float64); ok {
				oldHumidity := thermostat.CurrentHumidity
				thermostat.CurrentHumidity = humidity
				thermostat.LastSensorUpdate = time.Now()
				thermostat.IsOnline = true
				thermostat.UpdatedAt = time.Now()

				ts.logger.Printf("Updated thermostat %s humidity: %.1f%% -> %.1f%%",
					thermostat.ID, oldHumidity, thermostat.CurrentHumidity)
			}
			break
		}
	}

	return nil
}

// controlLoop runs the main control logic for all thermostats
func (ts *ThermostatService) controlLoop() {
	ticker := time.NewTicker(30 * time.Second) // Check every 30 seconds
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			ts.processAllThermostats()
		}
	}
}

// processAllThermostats processes control logic for all thermostats
func (ts *ThermostatService) processAllThermostats() {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	for _, thermostat := range ts.thermostats {
		ts.processThermostat(thermostat)
	}
}

// processThermostat processes control logic for a single thermostat
func (ts *ThermostatService) processThermostat(thermostat *models.Thermostat) {
	// Check if sensor data is stale
	if time.Since(thermostat.LastSensorUpdate) > 5*time.Minute {
		thermostat.IsOnline = false
		ts.logger.Printf("Thermostat %s is offline - no sensor data", thermostat.ID)
		return
	}

	// Determine next action
	nextStatus := thermostat.GetNextAction()

	// Only act if status changed
	if nextStatus != thermostat.Status {
		oldStatus := thermostat.Status
		thermostat.Status = nextStatus
		thermostat.UpdatedAt = time.Now()

		ts.logger.Printf("Thermostat %s status: %s -> %s (current: %.1f°C, target: %.1f°C)",
			thermostat.ID, oldStatus, nextStatus, thermostat.CurrentTemp, thermostat.TargetTemp)

		// Send control command
		ts.sendControlCommand(thermostat, nextStatus)
	}
}

// sendControlCommand sends a control command to the HVAC system
func (ts *ThermostatService) sendControlCommand(thermostat *models.Thermostat, status models.ThermostatStatus) {
	topic := fmt.Sprintf("thermostat/%s/control", thermostat.ID)

	command := map[string]interface{}{
		"action":    string(status),
		"target":    thermostat.TargetTemp,
		"current":   thermostat.CurrentTemp,
		"fan_speed": thermostat.FanSpeed,
		"timestamp": time.Now().Unix(),
	}

	payload, err := json.Marshal(command)
	if err != nil {
		ts.logger.Printf("Failed to marshal control command: %v", err)
		return
	}

	msg := &mqtt.Message{
		Topic:   topic,
		Payload: payload,
		QoS:     1,
		Retain:  false,
	}

	if err := ts.mqttClient.Publish(msg); err != nil {
		ts.logger.Printf("Failed to publish control command: %v", err)
	} else {
		ts.logger.Printf("Sent control command to %s: %s", thermostat.ID, string(status))
	}
}

// publishThermostatCommand publishes a command to the thermostat
func (ts *ThermostatService) publishThermostatCommand(id string, cmdType string, value interface{}) {
	topic := fmt.Sprintf("thermostat/%s/command", id)

	command := models.ThermostatCommand{
		Type:      cmdType,
		Value:     value,
		Timestamp: time.Now(),
	}

	payload, err := json.Marshal(command)
	if err != nil {
		ts.logger.Printf("Failed to marshal thermostat command: %v", err)
		return
	}

	msg := &mqtt.Message{
		Topic:   topic,
		Payload: payload,
		QoS:     1,
		Retain:  false,
	}

	if err := ts.mqttClient.Publish(msg); err != nil {
		ts.logger.Printf("Failed to publish thermostat command: %v", err)
	}
}

// GetRoomTemperature gets the current temperature for a room
func (ts *ThermostatService) GetRoomTemperature(roomID string) (float64, error) {
	ts.mu.RLock()
	defer ts.mu.RUnlock()

	for _, thermostat := range ts.thermostats {
		if thermostat.RoomID == roomID {
			return thermostat.CurrentTemp, nil
		}
	}

	return 0, fmt.Errorf("no thermostat found for room: %s", roomID)
}

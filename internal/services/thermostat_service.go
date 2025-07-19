package services

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/johnpr01/home-automation/internal/errors"
	"github.com/johnpr01/home-automation/internal/logger"
	"github.com/johnpr01/home-automation/internal/models"
	"github.com/johnpr01/home-automation/pkg/mqtt"
	"github.com/johnpr01/home-automation/pkg/utils"
)

// ThermostatService manages smart thermostats and processes sensor data
type ThermostatService struct {
	thermostats  map[string]*models.Thermostat
	mqttClient   *mqtt.Client
	mu           sync.RWMutex
	logger       *logger.Logger
	errorHandler *errors.ErrorHandler
}

// NewThermostatService creates a new thermostat service
func NewThermostatService(mqttClient *mqtt.Client, serviceLogger *logger.Logger) *ThermostatService {
	service := &ThermostatService{
		thermostats:  make(map[string]*models.Thermostat),
		mqttClient:   mqttClient,
		logger:       serviceLogger,
		errorHandler: errors.NewErrorHandler("thermostat-service"),
	}

	// Subscribe to sensor topics
	service.subscribeSensorTopics()

	// Start control loop
	go service.controlLoop()

	return service
}

// HandleTemperatureUpdate handles temperature updates from unified sensor service
func (ts *ThermostatService) HandleTemperatureUpdate(roomID string, temperature float64) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	// Get or create thermostat for this room
	thermostat, exists := ts.thermostats[roomID]
	if !exists {
		// Create default thermostat for this room
		thermostat = &models.Thermostat{
			ID:               roomID,
			Name:             "Thermostat-" + roomID,
			RoomID:           roomID,
			CurrentTemp:      temperature,
			TargetTemp:       72.0, // Default 72°F
			Mode:             models.ModeAuto,
			Status:           models.StatusIdle,
			Hysteresis:       1.0,
			MinTemp:          60.0,
			MaxTemp:          85.0,
			HeatingEnabled:   true,
			CoolingEnabled:   true,
			LastSensorUpdate: time.Now(),
			CreatedAt:        time.Now(),
			UpdatedAt:        time.Now(),
			IsOnline:         true,
		}
		ts.thermostats[roomID] = thermostat
		ts.logger.Info("Created new thermostat for room", map[string]interface{}{
			"room_id": roomID,
		})
	}

	// Update current temperature
	oldTemp := thermostat.CurrentTemp
	thermostat.CurrentTemp = temperature
	thermostat.LastSensorUpdate = time.Now()
	
	ts.logger.Info("Temperature update received", map[string]interface{}{
		"room_id":     roomID,
		"old_temp":    oldTemp,
		"new_temp":    temperature,
		"thermostat":  thermostat.ID,
		"updated_at":  thermostat.LastSensorUpdate,
	})
	thermostat.UpdatedAt = time.Now()
	thermostat.IsOnline = true

	ts.logger.Info(fmt.Sprintf("Thermostat %s temperature update: %.1f°F -> %.1f°F", roomID, oldTemp, temperature), map[string]interface{}{
		"room_id":   roomID,
		"old_temp":  oldTemp,
		"new_temp":  temperature,
		"device_id": thermostat.ID,
	})

	// Trigger control logic evaluation
	go ts.processThermostat(thermostat)
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
	ts.logger.Info("Registered new thermostat", map[string]interface{}{
		"thermostat_id": thermostat.ID,
		"room_id":       thermostat.RoomID,
		"target_temp":   thermostat.TargetTemp,
		"mode":          thermostat.Mode,
		"created_at":    thermostat.CreatedAt,
	})
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
		ts.logger.Error("Thermostat not found when setting target temperature", fmt.Errorf("error"), map[string]interface{}{
			"thermostat_id": id,
			"target_temp":   temp,
		})
		return fmt.Errorf("thermostat not found: %s", id)
	}

	if !thermostat.IsValidTargetTemp(temp) {
		ts.logger.Error("Invalid target temperature", fmt.Errorf("invalid target temperature"), map[string]interface{}{
			"thermostat_id": id,
			"target_temp":   temp,
			"min_temp":      thermostat.MinTemp,
			"max_temp":      thermostat.MaxTemp,
		})
		return fmt.Errorf("invalid target temperature: %.1f (range: %.1f-%.1f)",
			temp, thermostat.MinTemp, thermostat.MaxTemp)
	}

	thermostat.TargetTemp = temp
	thermostat.UpdatedAt = time.Now()

	ts.logger.Info("Set target temperature", map[string]interface{}{
		"thermostat_id": id,
		"target_temp":   temp,
		"previous_temp": thermostat.TargetTemp,
		"mode":          thermostat.Mode,
		"updated_at":    thermostat.UpdatedAt,
	})

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
		ts.logger.Error("Thermostat not found when setting mode", fmt.Errorf("error"), map[string]interface{}{
			"thermostat_id": id,
			"mode":         mode,
		})
		return fmt.Errorf("thermostat not found: %s", id)
	}

	if !thermostat.IsValidMode(mode) {
		ts.logger.Error("Invalid thermostat mode", fmt.Errorf("error"), map[string]interface{}{
			"thermostat_id": id,
			"mode":         mode,
			"current_mode": thermostat.Mode,
		})
		return fmt.Errorf("invalid mode: %s", mode)
	}

	oldMode := thermostat.Mode
	thermostat.Mode = mode
	thermostat.UpdatedAt = time.Now()
	
	ts.logger.Info("Set thermostat mode", map[string]interface{}{
		"thermostat_id": id,
		"old_mode":     oldMode,
		"new_mode":     mode,
		"updated_at":   thermostat.UpdatedAt,
	})

	ts.logger.Info(fmt.Sprintf("Set mode for %s to %s", id, mode))

	// Publish command to MQTT
	ts.publishThermostatCommand(id, models.CmdSetMode, string(mode))

	return nil
}

// subscribeSensorTopics subscribes to MQTT topics for sensor data
func (ts *ThermostatService) subscribeSensorTopics() {
	// Subscribe to temperature topics from Pi Pico sensors
	ts.mqttClient.Subscribe("room-temp/+", ts.handleTemperatureMessage)
	ts.mqttClient.Subscribe("room-hum/+", ts.handleHumidityMessage)

	ts.logger.Info("Subscribed to sensor MQTT topics: temp, humidity")
}

// handleTemperatureMessage processes temperature messages from Pi Pico sensors
func (ts *ThermostatService) handleTemperatureMessage(topic string, payload []byte) error {
	// Extract room number from topic (room-temp/1)
	parts := strings.Split(topic, "/")
	if len(parts) != 2 {
		ts.logger.Error("Invalid temperature topic format", fmt.Errorf("error"), map[string]interface{}{
			"topic": topic,
			"parts": len(parts),
		})
		return fmt.Errorf("invalid topic format: %s", topic)
	}

	roomID := parts[1]

	// Parse JSON payload
	var sensorData map[string]interface{}
	if err := json.Unmarshal(payload, &sensorData); err != nil {
		ts.logger.Error("Failed to parse temperature message", fmt.Errorf("error"), map[string]interface{}{
			"error":   err.Error(),
			"topic":   topic,
			"payload": string(payload),
		})
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

			ts.logger.Info("Updated thermostat temperature", map[string]interface{}{
				"thermostat_id": thermostat.ID,
				"room_id":       roomID,
				"old_temp":      oldTemp,
				"new_temp":      thermostat.CurrentTemp,
				"offset":        thermostat.TemperatureOffset,
				"is_online":     thermostat.IsOnline,
				"updated_at":    thermostat.UpdatedAt,
			})
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
		ts.logger.Info(fmt.Sprintf("Invalid humidity topic format: %s", topic))
		return fmt.Errorf("invalid topic format: %s", topic)
	}

	roomID := parts[1]

	// Parse JSON payload
	var sensorData map[string]interface{}
	if err := json.Unmarshal(payload, &sensorData); err != nil {
		ts.logger.Info(fmt.Sprintf("Failed to parse humidity message: %v", err))
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

				ts.logger.Info(fmt.Sprintf("Updated thermostat %s humidity: %.1f%% -> %.1f%%", thermostat.ID, oldHumidity, thermostat.CurrentHumidity))
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

	for range ticker.C {
		ts.processAllThermostats()
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
		ts.logger.Warn("Thermostat offline - no recent sensor data", map[string]interface{}{
			"thermostat_id":      thermostat.ID,
			"room_id":           thermostat.RoomID,
			"last_update":       thermostat.LastSensorUpdate,
			"minutes_since_update": time.Since(thermostat.LastSensorUpdate).Minutes(),
		})
		return
	}

	// Determine next action
	nextStatus := thermostat.GetNextAction()

	// Only act if status changed
	if nextStatus != thermostat.Status {
		oldStatus := thermostat.Status
		thermostat.Status = nextStatus
		thermostat.UpdatedAt = time.Now()

		ts.logger.Info("Thermostat status changed", map[string]interface{}{
			"thermostat_id": thermostat.ID,
			"room_id":      thermostat.RoomID,
			"old_status":   oldStatus,
			"new_status":   nextStatus,
			"current_temp": thermostat.CurrentTemp,
			"target_temp":  thermostat.TargetTemp,
			"mode":         thermostat.Mode,
			"updated_at":   thermostat.UpdatedAt,
		})

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
		ts.logger.Error("Failed to marshal control command", fmt.Errorf("error"), map[string]interface{}{
			"error":         err.Error(),
			"thermostat_id": thermostat.ID,
			"status":        status,
			"target_temp":   thermostat.TargetTemp,
		})
		return
	}

	msg := &mqtt.Message{
		Topic:   topic,
		Payload: payload,
		QoS:     1,
		Retain:  false,
	}

	if err := ts.mqttClient.Publish(msg); err != nil {
		ts.logger.Error("Failed to publish control command", fmt.Errorf("error"), map[string]interface{}{
			"error":         err.Error(),
			"thermostat_id": thermostat.ID,
			"topic":         topic,
			"status":        status,
			"target_temp":   thermostat.TargetTemp,
		})
	} else {
		ts.logger.Info("Published control command", map[string]interface{}{
			"thermostat_id": thermostat.ID,
			"status":        status,
			"topic":         topic,
			"target_temp":   thermostat.TargetTemp,
			"current_temp":  thermostat.CurrentTemp,
			"fan_speed":     thermostat.FanSpeed,
		})
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
		ts.logger.Error("Failed to marshal thermostat command", fmt.Errorf("error"), map[string]interface{}{
			"error":         err.Error(),
			"thermostat_id": id,
			"command_type":  cmdType,
			"value":         value,
		})
		return
	}

	msg := &mqtt.Message{
		Topic:   topic,
		Payload: payload,
		QoS:     1,
		Retain:  false,
	}

	if err := ts.mqttClient.Publish(msg); err != nil {
		ts.logger.Error("Failed to publish thermostat command", fmt.Errorf("error"), map[string]interface{}{
			"error":         err.Error(),
			"thermostat_id": id,
			"topic":         topic,
			"command_type":  cmdType,
			"value":         value,
		})
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

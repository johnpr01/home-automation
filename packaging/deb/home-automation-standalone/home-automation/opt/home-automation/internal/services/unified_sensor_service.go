package services

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/johnpr01/home-automation/pkg/mqtt"
)

// UnifiedSensorMessage represents all sensor data from a single Pi Pico device
type UnifiedSensorMessage struct {
	// Temperature data
	Temperature float64 `json:"temperature,omitempty"`
	TempUnit    string  `json:"temp_unit,omitempty"`

	// Humidity data
	Humidity     float64 `json:"humidity,omitempty"`
	HumidityUnit string  `json:"humidity_unit,omitempty"`

	// Motion data
	Motion      *bool  `json:"motion,omitempty"`
	MotionStart *int64 `json:"motion_start,omitempty"`

	// Light data
	LightLevel   *float64 `json:"light_level,omitempty"`
	LightPercent *float64 `json:"light_percent,omitempty"`
	LightState   string   `json:"light_state,omitempty"`

	// Common metadata
	Room      string `json:"room"`
	Sensor    string `json:"sensor"`
	Timestamp int64  `json:"timestamp"`
	DeviceID  string `json:"device_id"`
}

// RoomSensorData aggregates all sensor data for a room
type RoomSensorData struct {
	RoomID   string `json:"room_id"`
	DeviceID string `json:"device_id"`

	// Temperature/Humidity
	Temperature    float64   `json:"temperature"`
	Humidity       float64   `json:"humidity"`
	TempLastUpdate time.Time `json:"temp_last_update"`

	// Motion
	IsOccupied      bool      `json:"is_occupied"`
	MotionLastTime  time.Time `json:"motion_last_time"`
	MotionClearTime time.Time `json:"motion_clear_time"`

	// Light
	LightLevel      float64   `json:"light_level"`
	LightState      string    `json:"light_state"`
	DayNightCycle   string    `json:"day_night_cycle"`
	LightLastUpdate time.Time `json:"light_last_update"`

	// Device status
	IsOnline bool      `json:"is_online"`
	LastSeen time.Time `json:"last_seen"`
}

// UnifiedSensorService manages all sensor data from Pi Pico devices
type UnifiedSensorService struct {
	roomSensors map[string]*RoomSensorData
	mqttClient  *mqtt.Client
	mu          sync.RWMutex
	logger      *log.Logger

	// Callbacks for other services
	tempCallbacks   []func(roomID string, temperature float64)
	motionCallbacks []func(roomID string, occupied bool)
	lightCallbacks  []func(roomID string, lightState string, lightLevel float64)
}

// NewUnifiedSensorService creates a new unified sensor service
func NewUnifiedSensorService(mqttClient *mqtt.Client, logger *log.Logger) *UnifiedSensorService {
	service := &UnifiedSensorService{
		roomSensors:     make(map[string]*RoomSensorData),
		mqttClient:      mqttClient,
		logger:          logger,
		tempCallbacks:   make([]func(string, float64), 0),
		motionCallbacks: make([]func(string, bool), 0),
		lightCallbacks:  make([]func(string, string, float64), 0),
	}

	// Subscribe to all sensor topics from Pi Pico devices
	service.subscribeSensorTopics()

	// Start cleanup routine
	go service.cleanupRoutine()

	return service
}

// AddTemperatureCallback registers a callback for temperature updates
func (uss *UnifiedSensorService) AddTemperatureCallback(callback func(roomID string, temperature float64)) {
	uss.mu.Lock()
	defer uss.mu.Unlock()
	uss.tempCallbacks = append(uss.tempCallbacks, callback)
}

// AddMotionCallback registers a callback for motion updates
func (uss *UnifiedSensorService) AddMotionCallback(callback func(roomID string, occupied bool)) {
	uss.mu.Lock()
	defer uss.mu.Unlock()
	uss.motionCallbacks = append(uss.motionCallbacks, callback)
}

// AddLightCallback registers a callback for light updates
func (uss *UnifiedSensorService) AddLightCallback(callback func(roomID string, lightState string, lightLevel float64)) {
	uss.mu.Lock()
	defer uss.mu.Unlock()
	uss.lightCallbacks = append(uss.lightCallbacks, callback)
}

// GetRoomSensorData returns all sensor data for a room
func (uss *UnifiedSensorService) GetRoomSensorData(roomID string) (*RoomSensorData, bool) {
	uss.mu.RLock()
	defer uss.mu.RUnlock()

	data, exists := uss.roomSensors[roomID]
	if !exists {
		return nil, false
	}

	// Return a copy to avoid race conditions
	dataCopy := *data
	return &dataCopy, true
}

// GetAllRoomSensors returns sensor data for all rooms
func (uss *UnifiedSensorService) GetAllRoomSensors() map[string]*RoomSensorData {
	uss.mu.RLock()
	defer uss.mu.RUnlock()

	result := make(map[string]*RoomSensorData)
	for roomID, data := range uss.roomSensors {
		dataCopy := *data
		result[roomID] = &dataCopy
	}
	return result
}

// subscribeSensorTopics sets up MQTT subscriptions for all sensor data
func (uss *UnifiedSensorService) subscribeSensorTopics() {
	// Subscribe to all sensor topics from Pi Pico devices
	uss.mqttClient.Subscribe("room-temp/+", uss.handleTemperatureMessage)
	uss.mqttClient.Subscribe("room-hum/+", uss.handleHumidityMessage)
	uss.mqttClient.Subscribe("room-motion/+", uss.handleMotionMessage)
	uss.mqttClient.Subscribe("room-light/+", uss.handleLightMessage)

	uss.logger.Println("UnifiedSensorService: Subscribed to all Pi Pico sensor topics")
}

// handleTemperatureMessage processes temperature messages from Pi Pico
func (uss *UnifiedSensorService) handleTemperatureMessage(topic string, payload []byte) error {
	roomID, err := uss.extractRoomID(topic)
	if err != nil {
		return err
	}

	var tempMsg UnifiedSensorMessage
	if err := json.Unmarshal(payload, &tempMsg); err != nil {
		uss.logger.Printf("Failed to parse temperature message for room %s: %v", roomID, err)
		return err
	}

	uss.mu.Lock()
	defer uss.mu.Unlock()

	// Get or create room sensor data
	roomData := uss.getOrCreateRoomData(roomID, tempMsg.DeviceID)

	// Update temperature data
	oldTemp := roomData.Temperature
	roomData.Temperature = tempMsg.Temperature
	roomData.TempLastUpdate = time.Now()
	roomData.LastSeen = time.Now()
	roomData.IsOnline = true

	uss.logger.Printf("UnifiedSensor: Room %s temperature: %.1f°F -> %.1f°F (device: %s)",
		roomID, oldTemp, roomData.Temperature, roomData.DeviceID)

	// Notify temperature callbacks
	for _, callback := range uss.tempCallbacks {
		go callback(roomID, roomData.Temperature)
	}

	return nil
}

// handleHumidityMessage processes humidity messages from Pi Pico
func (uss *UnifiedSensorService) handleHumidityMessage(topic string, payload []byte) error {
	roomID, err := uss.extractRoomID(topic)
	if err != nil {
		return err
	}

	var humMsg UnifiedSensorMessage
	if err := json.Unmarshal(payload, &humMsg); err != nil {
		uss.logger.Printf("Failed to parse humidity message for room %s: %v", roomID, err)
		return err
	}

	uss.mu.Lock()
	defer uss.mu.Unlock()

	// Get or create room sensor data
	roomData := uss.getOrCreateRoomData(roomID, humMsg.DeviceID)

	// Update humidity data
	oldHumidity := roomData.Humidity
	roomData.Humidity = humMsg.Humidity
	roomData.LastSeen = time.Now()
	roomData.IsOnline = true

	uss.logger.Printf("UnifiedSensor: Room %s humidity: %.1f%% -> %.1f%% (device: %s)",
		roomID, oldHumidity, roomData.Humidity, roomData.DeviceID)

	return nil
}

// handleMotionMessage processes motion messages from Pi Pico
func (uss *UnifiedSensorService) handleMotionMessage(topic string, payload []byte) error {
	roomID, err := uss.extractRoomID(topic)
	if err != nil {
		return err
	}

	var motionMsg UnifiedSensorMessage
	if err := json.Unmarshal(payload, &motionMsg); err != nil {
		uss.logger.Printf("Failed to parse motion message for room %s: %v", roomID, err)
		return err
	}

	uss.mu.Lock()
	defer uss.mu.Unlock()

	// Get or create room sensor data
	roomData := uss.getOrCreateRoomData(roomID, motionMsg.DeviceID)

	// Update motion data
	previouslyOccupied := roomData.IsOccupied
	currentTime := time.Now()

	if motionMsg.Motion != nil {
		roomData.IsOccupied = *motionMsg.Motion

		if *motionMsg.Motion {
			roomData.MotionLastTime = currentTime
		} else {
			roomData.MotionClearTime = currentTime
		}

		roomData.LastSeen = currentTime
		roomData.IsOnline = true

		// Log state changes
		if previouslyOccupied != roomData.IsOccupied {
			status := "OCCUPIED"
			if !roomData.IsOccupied {
				status = "UNOCCUPIED"
			}
			uss.logger.Printf("UnifiedSensor: Room %s is now %s (device: %s)",
				roomID, status, roomData.DeviceID)

			// Notify motion callbacks
			for _, callback := range uss.motionCallbacks {
				go callback(roomID, roomData.IsOccupied)
			}
		}
	}

	return nil
}

// handleLightMessage processes light sensor messages from Pi Pico
func (uss *UnifiedSensorService) handleLightMessage(topic string, payload []byte) error {
	roomID, err := uss.extractRoomID(topic)
	if err != nil {
		return err
	}

	var lightMsg UnifiedSensorMessage
	if err := json.Unmarshal(payload, &lightMsg); err != nil {
		uss.logger.Printf("Failed to parse light message for room %s: %v", roomID, err)
		return err
	}

	uss.mu.Lock()
	defer uss.mu.Unlock()

	// Get or create room sensor data
	roomData := uss.getOrCreateRoomData(roomID, lightMsg.DeviceID)

	// Update light data
	previousState := roomData.LightState
	currentTime := time.Now()

	if lightMsg.LightLevel != nil {
		roomData.LightLevel = *lightMsg.LightLevel
		roomData.LightState = lightMsg.LightState
		roomData.DayNightCycle = uss.determineDayNightCycle(*lightMsg.LightLevel)
		roomData.LightLastUpdate = currentTime
		roomData.LastSeen = currentTime
		roomData.IsOnline = true

		// Log state changes
		if previousState != roomData.LightState {
			uss.logger.Printf("UnifiedSensor: Room %s light: %s -> %s (%.1f%%) (device: %s)",
				roomID, previousState, roomData.LightState, roomData.LightLevel, roomData.DeviceID)

			// Notify light callbacks
			for _, callback := range uss.lightCallbacks {
				go callback(roomID, roomData.LightState, roomData.LightLevel)
			}
		}
	}

	return nil
}

// extractRoomID extracts room ID from MQTT topic
func (uss *UnifiedSensorService) extractRoomID(topic string) (string, error) {
	parts := strings.Split(topic, "/")
	if len(parts) != 2 {
		return "", fmt.Errorf("invalid topic format: %s", topic)
	}
	return parts[1], nil
}

// getOrCreateRoomData gets existing room data or creates new entry
func (uss *UnifiedSensorService) getOrCreateRoomData(roomID, deviceID string) *RoomSensorData {
	roomData, exists := uss.roomSensors[roomID]
	if !exists {
		roomData = &RoomSensorData{
			RoomID:        roomID,
			DeviceID:      deviceID,
			LightState:    "unknown",
			DayNightCycle: "unknown",
			IsOnline:      false,
		}
		uss.roomSensors[roomID] = roomData
	}

	// Update device ID if it changed
	if roomData.DeviceID != deviceID {
		roomData.DeviceID = deviceID
	}

	return roomData
}

// determineDayNightCycle determines day/night cycle based on light level
func (uss *UnifiedSensorService) determineDayNightCycle(lightLevel float64) string {
	currentHour := time.Now().Hour()

	if lightLevel < 5.0 && (currentHour < 6 || currentHour > 22) {
		return "night"
	} else if lightLevel > 70.0 && currentHour >= 10 && currentHour <= 16 {
		return "day"
	} else if lightLevel > 30.0 && (currentHour >= 6 && currentHour < 10) {
		return "dawn"
	} else if lightLevel > 20.0 && (currentHour >= 17 && currentHour <= 22) {
		return "dusk"
	} else {
		return "transitional"
	}
}

// cleanupRoutine marks sensors as offline if no recent updates
func (uss *UnifiedSensorService) cleanupRoutine() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		uss.mu.Lock()
		currentTime := time.Now()

		for roomID, roomData := range uss.roomSensors {
			// Mark as offline if no updates for 10 minutes
			if currentTime.Sub(roomData.LastSeen) > 10*time.Minute {
				if roomData.IsOnline {
					roomData.IsOnline = false
					uss.logger.Printf("UnifiedSensor: Room %s sensors marked offline (device: %s)",
						roomID, roomData.DeviceID)
				}
			}
		}
		uss.mu.Unlock()
	}
}

// GetSensorSummary returns a summary of all sensors
func (uss *UnifiedSensorService) GetSensorSummary() map[string]interface{} {
	uss.mu.RLock()
	defer uss.mu.RUnlock()

	summary := make(map[string]interface{})
	summary["total_rooms"] = len(uss.roomSensors)

	onlineCount := 0
	occupiedCount := 0
	avgTemp := 0.0
	avgHumidity := 0.0
	avgLight := 0.0

	rooms := make([]map[string]interface{}, 0, len(uss.roomSensors))

	for _, roomData := range uss.roomSensors {
		if roomData.IsOnline {
			onlineCount++
			avgTemp += roomData.Temperature
			avgHumidity += roomData.Humidity
			avgLight += roomData.LightLevel
		}
		if roomData.IsOccupied {
			occupiedCount++
		}

		roomInfo := map[string]interface{}{
			"room_id":         roomData.RoomID,
			"device_id":       roomData.DeviceID,
			"temperature":     roomData.Temperature,
			"humidity":        roomData.Humidity,
			"is_occupied":     roomData.IsOccupied,
			"light_level":     roomData.LightLevel,
			"light_state":     roomData.LightState,
			"day_night_cycle": roomData.DayNightCycle,
			"is_online":       roomData.IsOnline,
			"last_seen":       roomData.LastSeen.Format(time.RFC3339),
		}
		rooms = append(rooms, roomInfo)
	}

	if onlineCount > 0 {
		avgTemp = avgTemp / float64(onlineCount)
		avgHumidity = avgHumidity / float64(onlineCount)
		avgLight = avgLight / float64(onlineCount)
	}

	summary["online_devices"] = onlineCount
	summary["occupied_rooms"] = occupiedCount
	summary["average_temperature"] = avgTemp
	summary["average_humidity"] = avgHumidity
	summary["average_light_level"] = avgLight
	summary["rooms"] = rooms

	return summary
}

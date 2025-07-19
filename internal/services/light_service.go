package services

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/johnpr01/home-automation/internal/logger"
	"github.com/johnpr01/home-automation/pkg/mqtt"
)

// LightSensorMessage represents light sensor data from Pi Pico
type LightSensorMessage struct {
	LightLevel   float64 `json:"light_level"`
	LightPercent float64 `json:"light_percent"`
	LightState   string  `json:"light_state"` // "dark", "normal", "bright"
	Unit         string  `json:"unit"`
	Room         string  `json:"room"`
	Sensor       string  `json:"sensor"`
	Timestamp    int64   `json:"timestamp"`
	DeviceID     string  `json:"device_id"`
}

// RoomLightLevel tracks light levels for each room
type RoomLightLevel struct {
	RoomID         string    `json:"room_id"`
	LightLevel     float64   `json:"light_level"` // Light level percentage (0-100%)
	LightState     string    `json:"light_state"` // "dark", "normal", "bright"
	LastUpdateTime time.Time `json:"last_update_time"`
	DeviceID       string    `json:"device_id"`
	SensorType     string    `json:"sensor_type"`
	IsOnline       bool      `json:"is_online"`
	DayNightCycle  string    `json:"day_night_cycle"` // "day", "night", "dawn", "dusk"
}

// LightService manages photo transistor light sensors and ambient light tracking
type LightService struct {
	roomLightLevels map[string]*RoomLightLevel
	mqttClient      *mqtt.Client
	mu              sync.RWMutex
	logger          *logger.Logger
	callbacks       []func(roomID string, lightState string, lightLevel float64)

	// Configuration thresholds
	darkThreshold   float64 // Below this is considered "dark"
	brightThreshold float64 // Above this is considered "bright"
}

// NewLightService creates a new light sensor service
func NewLightService(mqttClient *mqtt.Client, logger *logger.Logger) *LightService {
	service := &LightService{
		roomLightLevels: make(map[string]*RoomLightLevel),
		mqttClient:      mqttClient,
		logger:          logger,
		callbacks:       make([]func(string, string, float64), 0),
		darkThreshold:   10.0, // Default: <10% is dark
		brightThreshold: 80.0, // Default: >80% is bright
	}

	// Subscribe to light sensor topics
	service.subscribeLightTopics()

	// Start cleanup routine for stale data
	go service.cleanupRoutine()

	// Start day/night cycle detection
	go service.dayNightDetection()

	return service
}

// SetThresholds allows customization of light level thresholds
func (ls *LightService) SetThresholds(darkThreshold, brightThreshold float64) {
	ls.mu.Lock()
	defer ls.mu.Unlock()
	ls.darkThreshold = darkThreshold
	ls.brightThreshold = brightThreshold
}

// AddLightCallback registers a callback for light level changes
func (ls *LightService) AddLightCallback(callback func(roomID string, lightState string, lightLevel float64)) {
	ls.mu.Lock()
	defer ls.mu.Unlock()
	ls.callbacks = append(ls.callbacks, callback)
}

// GetRoomLightLevel returns the current light level for a room
func (ls *LightService) GetRoomLightLevel(roomID string) (*RoomLightLevel, bool) {
	ls.mu.RLock()
	defer ls.mu.RUnlock()

	lightLevel, exists := ls.roomLightLevels[roomID]
	if !exists {
		return nil, false
	}

	// Return a copy to avoid race conditions
	lightLevelCopy := *lightLevel
	return &lightLevelCopy, true
}

// GetAllLightLevels returns light levels for all rooms
func (ls *LightService) GetAllLightLevels() map[string]*RoomLightLevel {
	ls.mu.RLock()
	defer ls.mu.RUnlock()

	result := make(map[string]*RoomLightLevel)
	for roomID, lightLevel := range ls.roomLightLevels {
		lightLevelCopy := *lightLevel
		result[roomID] = &lightLevelCopy
	}
	return result
}

// subscribeLightTopics sets up MQTT subscriptions for light sensor data
func (ls *LightService) subscribeLightTopics() {
	// Subscribe to light sensor messages
	ls.mqttClient.Subscribe("room-light/+", ls.handleLightMessage)
	ls.logger.Info("Subscribed to room-light/+ topics")
}

// handleLightMessage processes light sensor messages from Pi Pico sensors
func (ls *LightService) handleLightMessage(topic string, payload []byte) error {
	// Extract room number from topic (room-light/1)
	parts := strings.Split(topic, "/")
	if len(parts) != 2 {
		return fmt.Errorf("invalid light topic format: %s", topic)
	}
	roomID := parts[1]

	// Parse light message
	var lightMsg LightSensorMessage
	if err := json.Unmarshal(payload, &lightMsg); err != nil {
		ls.logger.Error(fmt.Sprintf("Failed to parse light message for room %s", roomID), err)
		return err
	}

	ls.mu.Lock()
	defer ls.mu.Unlock()

	// Get or create room light level record
	lightLevel, exists := ls.roomLightLevels[roomID]
	if !exists {
		lightLevel = &RoomLightLevel{
			RoomID:     roomID,
			LightLevel: 0,
			LightState: "unknown",
			IsOnline:   false,
		}
		ls.roomLightLevels[roomID] = lightLevel
	}

	// Update sensor connectivity and data
	lightLevel.IsOnline = true
	lightLevel.DeviceID = lightMsg.DeviceID
	lightLevel.SensorType = lightMsg.Sensor
	lightLevel.LastUpdateTime = time.Now()

	// Track light level changes
	previousLevel := lightLevel.LightLevel
	previousState := lightLevel.LightState

	lightLevel.LightLevel = lightMsg.LightLevel
	lightLevel.LightState = lightMsg.LightState

	// Determine day/night cycle based on light level patterns
	lightLevel.DayNightCycle = ls.determineDayNightCycle(lightMsg.LightLevel)

	// Log significant changes
	if previousState != lightLevel.LightState {
		ls.logger.Info(fmt.Sprintf("Room %s light state changed: %s -> %s (%.1f%%)",
			roomID, previousState, lightLevel.LightState, lightLevel.LightLevel))

		// Notify callbacks of light state change
		for _, callback := range ls.callbacks {
			go callback(roomID, lightLevel.LightState, lightLevel.LightLevel)
		}
	} else if abs(previousLevel-lightLevel.LightLevel) > 10.0 {
		// Log significant level changes (>10%)
		ls.logger.Info(fmt.Sprintf("Room %s light level: %.1f%% -> %.1f%% (%s)",
			roomID, previousLevel, lightLevel.LightLevel, lightLevel.LightState))
	}

	return nil
}

// determineDayNightCycle determines the day/night cycle based on light patterns
func (ls *LightService) determineDayNightCycle(lightLevel float64) string {
	currentHour := time.Now().Hour()

	// Basic day/night cycle detection combining time and light level
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

// cleanupRoutine periodically marks sensors as offline if no recent updates
func (ls *LightService) cleanupRoutine() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		ls.mu.Lock()
		currentTime := time.Now()

		for roomID, lightLevel := range ls.roomLightLevels {
			// Mark as offline if no updates for 10 minutes
			if currentTime.Sub(lightLevel.LastUpdateTime) > 10*time.Minute {
				if lightLevel.IsOnline {
					lightLevel.IsOnline = false
					ls.logger.Warn(fmt.Sprintf("Room %s light sensor marked offline (device: %s)",
						roomID, lightLevel.DeviceID))
				}
			}
		}
		ls.mu.Unlock()
	}
}

// dayNightDetection tracks overall day/night patterns
func (ls *LightService) dayNightDetection() {
	ticker := time.NewTicker(30 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		ls.mu.RLock()

		// Analyze light patterns across all rooms
		totalRooms := len(ls.roomLightLevels)
		if totalRooms == 0 {
			ls.mu.RUnlock()
			continue
		}

		dayCount := 0
		nightCount := 0

		for _, lightLevel := range ls.roomLightLevels {
			if !lightLevel.IsOnline {
				continue
			}

			switch lightLevel.DayNightCycle {
			case "day":
				dayCount++
			case "night":
				nightCount++
			}
		}

		ls.mu.RUnlock()

		// Log overall day/night status
		if dayCount > nightCount {
			ls.logger.Info(fmt.Sprintf("Overall lighting suggests DAY time (%d day, %d night out of %d rooms)",
				dayCount, nightCount, totalRooms))
		} else if nightCount > dayCount {
			ls.logger.Info(fmt.Sprintf("Overall lighting suggests NIGHT time (%d day, %d night out of %d rooms)",
				dayCount, nightCount, totalRooms))
		}
	}
}

// GetLightSummary returns a summary of all light sensors and their status
func (ls *LightService) GetLightSummary() map[string]interface{} {
	ls.mu.RLock()
	defer ls.mu.RUnlock()

	summary := make(map[string]interface{})
	summary["total_rooms"] = len(ls.roomLightLevels)

	darkCount := 0
	brightCount := 0
	onlineCount := 0
	averageLightLevel := 0.0

	rooms := make([]map[string]interface{}, 0, len(ls.roomLightLevels))

	for _, lightLevel := range ls.roomLightLevels {
		if lightLevel.LightState == "dark" {
			darkCount++
		} else if lightLevel.LightState == "bright" {
			brightCount++
		}
		if lightLevel.IsOnline {
			onlineCount++
			averageLightLevel += lightLevel.LightLevel
		}

		roomInfo := map[string]interface{}{
			"room_id":          lightLevel.RoomID,
			"light_level":      lightLevel.LightLevel,
			"light_state":      lightLevel.LightState,
			"day_night_cycle":  lightLevel.DayNightCycle,
			"is_online":        lightLevel.IsOnline,
			"device_id":        lightLevel.DeviceID,
			"sensor_type":      lightLevel.SensorType,
			"last_update_time": lightLevel.LastUpdateTime.Format(time.RFC3339),
		}
		rooms = append(rooms, roomInfo)
	}

	if onlineCount > 0 {
		averageLightLevel = averageLightLevel / float64(onlineCount)
	}

	summary["dark_rooms"] = darkCount
	summary["bright_rooms"] = brightCount
	summary["online_sensors"] = onlineCount
	summary["average_light_level"] = averageLightLevel
	summary["rooms"] = rooms

	return summary
}

// abs returns the absolute value of a float64
func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

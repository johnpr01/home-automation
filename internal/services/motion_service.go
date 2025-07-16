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

// MotionDetectionMessage represents motion sensor data
type MotionDetectionMessage struct {
	Motion      bool   `json:"motion"`
	Room        string `json:"room"`
	Sensor      string `json:"sensor"`
	Timestamp   int64  `json:"timestamp"`
	MotionStart int64  `json:"motion_start,omitempty"`
	DeviceID    string `json:"device_id"`
}

// RoomOccupancy tracks motion state for each room
type RoomOccupancy struct {
	RoomID          string    `json:"room_id"`
	IsOccupied      bool      `json:"is_occupied"`
	LastMotionTime  time.Time `json:"last_motion_time"`
	LastClearedTime time.Time `json:"last_cleared_time"`
	MotionStartTime time.Time `json:"motion_start_time"`
	DeviceID        string    `json:"device_id"`
	SensorType      string    `json:"sensor_type"`
	IsOnline        bool      `json:"is_online"`
}

// MotionService manages PIR motion detection and room occupancy tracking
type MotionService struct {
	roomOccupancy map[string]*RoomOccupancy
	mqttClient    *mqtt.Client
	mu            sync.RWMutex
	logger        *log.Logger
	callbacks     []func(roomID string, occupied bool)
}

// NewMotionService creates a new motion detection service
func NewMotionService(mqttClient *mqtt.Client, logger *log.Logger) *MotionService {
	service := &MotionService{
		roomOccupancy: make(map[string]*RoomOccupancy),
		mqttClient:    mqttClient,
		logger:        logger,
		callbacks:     make([]func(string, bool), 0),
	}

	// Subscribe to motion topics
	service.subscribeMotionTopics()

	// Start cleanup routine for stale data
	go service.cleanupRoutine()

	return service
}

// AddOccupancyCallback registers a callback for occupancy changes
func (ms *MotionService) AddOccupancyCallback(callback func(roomID string, occupied bool)) {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	ms.callbacks = append(ms.callbacks, callback)
}

// GetRoomOccupancy returns the current occupancy status for a room
func (ms *MotionService) GetRoomOccupancy(roomID string) (*RoomOccupancy, bool) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	occupancy, exists := ms.roomOccupancy[roomID]
	if !exists {
		return nil, false
	}

	// Return a copy to avoid race conditions
	occupancyCopy := *occupancy
	return &occupancyCopy, true
}

// GetAllOccupancy returns occupancy status for all rooms
func (ms *MotionService) GetAllOccupancy() map[string]*RoomOccupancy {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	result := make(map[string]*RoomOccupancy)
	for roomID, occupancy := range ms.roomOccupancy {
		occupancyCopy := *occupancy
		result[roomID] = &occupancyCopy
	}
	return result
}

// subscribeMotionTopics sets up MQTT subscriptions for motion detection
func (ms *MotionService) subscribeMotionTopics() {
	// Subscribe to motion detection messages
	ms.mqttClient.Subscribe("room-motion/+", ms.handleMotionMessage)
	ms.logger.Println("MotionService: Subscribed to room-motion/+ topics")
}

// extractRoomID extracts the room ID from an MQTT topic
func (ms *MotionService) extractRoomID(topic string) (string, error) {
	// Extract room number from topic (room-motion/1)
	parts := strings.Split(topic, "/")
	if len(parts) != 2 {
		return "", fmt.Errorf("invalid motion topic format: %s", topic)
	}
	roomID := parts[1]
	if roomID == "" {
		return "", fmt.Errorf("empty room ID in topic: %s", topic)
	}
	return roomID, nil
}

// handleMotionMessage processes motion detection messages from Pi Pico sensors
func (ms *MotionService) handleMotionMessage(topic string, payload []byte) error {
	roomID, err := ms.extractRoomID(topic)
	if err != nil {
		return err
	}

	// Parse motion message
	var motionMsg MotionDetectionMessage
	if err := json.Unmarshal(payload, &motionMsg); err != nil {
		ms.logger.Printf("MotionService: Failed to parse motion message for room %s: %v", roomID, err)
		return err
	}

	ms.mu.Lock()
	defer ms.mu.Unlock()

	// Get or create room occupancy record
	occupancy, exists := ms.roomOccupancy[roomID]
	if !exists {
		occupancy = &RoomOccupancy{
			RoomID:     roomID,
			IsOccupied: false,
			IsOnline:   false,
		}
		ms.roomOccupancy[roomID] = occupancy
	}

	// Update sensor connectivity
	occupancy.IsOnline = true
	occupancy.DeviceID = motionMsg.DeviceID
	occupancy.SensorType = motionMsg.Sensor

	// Track occupancy state changes
	previouslyOccupied := occupancy.IsOccupied
	currentTime := time.Now()

	if motionMsg.Motion {
		// Motion detected
		occupancy.IsOccupied = true
		occupancy.LastMotionTime = currentTime

		if motionMsg.MotionStart > 0 {
			occupancy.MotionStartTime = time.Unix(motionMsg.MotionStart, 0)
		} else {
			occupancy.MotionStartTime = currentTime
		}

		if !previouslyOccupied {
			ms.logger.Printf("MotionService: Motion DETECTED in room %s (device: %s)",
				roomID, motionMsg.DeviceID)

			// Notify callbacks of occupancy change
			for _, callback := range ms.callbacks {
				go callback(roomID, true)
			}
		}
	} else {
		// Motion cleared
		occupancy.IsOccupied = false
		occupancy.LastClearedTime = currentTime

		if previouslyOccupied {
			ms.logger.Printf("MotionService: Motion CLEARED in room %s (device: %s)",
				roomID, motionMsg.DeviceID)

			// Notify callbacks of occupancy change
			for _, callback := range ms.callbacks {
				go callback(roomID, false)
			}
		}
	}

	return nil
}

// cleanupRoutine periodically marks sensors as offline if no recent updates
func (ms *MotionService) cleanupRoutine() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		ms.mu.Lock()
		currentTime := time.Now()

		for roomID, occupancy := range ms.roomOccupancy {
			// Mark as offline if no updates for 10 minutes
			if currentTime.Sub(occupancy.LastMotionTime) > 10*time.Minute &&
				currentTime.Sub(occupancy.LastClearedTime) > 10*time.Minute {
				if occupancy.IsOnline {
					occupancy.IsOnline = false
					ms.logger.Printf("MotionService: Room %s sensor marked offline (device: %s)",
						roomID, occupancy.DeviceID)
				}
			}
		}
		ms.mu.Unlock()
	}
}

// GetMotionSummary returns a summary of all motion sensors and their status
func (ms *MotionService) GetMotionSummary() map[string]interface{} {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	summary := make(map[string]interface{})
	summary["total_rooms"] = len(ms.roomOccupancy)

	occupiedCount := 0
	onlineCount := 0

	rooms := make([]map[string]interface{}, 0, len(ms.roomOccupancy))

	for _, occupancy := range ms.roomOccupancy {
		if occupancy.IsOccupied {
			occupiedCount++
		}
		if occupancy.IsOnline {
			onlineCount++
		}

		roomInfo := map[string]interface{}{
			"room_id":           occupancy.RoomID,
			"is_occupied":       occupancy.IsOccupied,
			"is_online":         occupancy.IsOnline,
			"device_id":         occupancy.DeviceID,
			"sensor_type":       occupancy.SensorType,
			"last_motion_time":  occupancy.LastMotionTime.Format(time.RFC3339),
			"last_cleared_time": occupancy.LastClearedTime.Format(time.RFC3339),
		}
		rooms = append(rooms, roomInfo)
	}

	summary["occupied_rooms"] = occupiedCount
	summary["online_sensors"] = onlineCount
	summary["rooms"] = rooms

	return summary
}

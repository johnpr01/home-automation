package services

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/johnpr01/home-automation/internal/models"
	"github.com/johnpr01/home-automation/pkg/mqtt"
)

// AutomationRule represents a rule for home automation
type AutomationRule struct {
	ID            string                 `json:"id"`
	Name          string                 `json:"name"`
	RoomID        string                 `json:"room_id"`
	DeviceID      string                 `json:"device_id"`
	Conditions    map[string]interface{} `json:"conditions"`
	Actions       []models.DeviceCommand `json:"actions"`
	Enabled       bool                   `json:"enabled"`
	Priority      int                    `json:"priority"`
	Cooldown      time.Duration          `json:"cooldown"`
	LastTriggered time.Time              `json:"last_triggered"`
}

// AutomationService coordinates between sensors and devices for automated control
type AutomationService struct {
	motionService *MotionService
	lightService  *LightService
	deviceService *DeviceService
	mqttClient    *mqtt.Client
	logger        *log.Logger

	// Automation rules and state
	rules      map[string]*AutomationRule
	rulesMutex sync.RWMutex

	// Configuration
	motionLightCooldown time.Duration
	darkThreshold       float64
}

// NewAutomationService creates a new automation service
func NewAutomationService(motionService *MotionService, lightService *LightService, deviceService *DeviceService, mqttClient *mqtt.Client, logger *log.Logger) *AutomationService {
	service := &AutomationService{
		motionService:       motionService,
		lightService:        lightService,
		deviceService:       deviceService,
		mqttClient:          mqttClient,
		logger:              logger,
		rules:               make(map[string]*AutomationRule),
		motionLightCooldown: 5 * time.Minute, // Prevent rapid on/off cycles
		darkThreshold:       20.0,            // Below 20% light level is considered dark
	}

	// Register callbacks with sensor services
	service.setupSensorCallbacks()

	// Create default motion-activated lighting rules
	service.createDefaultRules()

	return service
}

// setupSensorCallbacks registers callbacks with motion and light sensors
func (as *AutomationService) setupSensorCallbacks() {
	// Motion sensor callback
	as.motionService.AddOccupancyCallback(func(roomID string, occupied bool) {
		as.handleMotionUpdate(roomID, occupied)
	})

	// Light sensor callback
	as.lightService.AddLightCallback(func(roomID string, lightState string, lightLevel float64) {
		as.handleLightUpdate(roomID, lightState, lightLevel)
	})
}

// createDefaultRules creates standard automation rules for motion-activated lighting
func (as *AutomationService) createDefaultRules() {
	// Create motion-activated lighting rule for each room with lights
	rooms := []string{"living-room", "kitchen", "bedroom", "bathroom", "office", "hallway"}

	for _, roomID := range rooms {
		lightDeviceID := fmt.Sprintf("light-%s", roomID)

		rule := &AutomationRule{
			ID:       fmt.Sprintf("motion-light-%s", roomID),
			Name:     fmt.Sprintf("Motion-Activated Lighting - %s", roomID),
			RoomID:   roomID,
			DeviceID: lightDeviceID,
			Conditions: map[string]interface{}{
				"motion_detected": true,
				"light_level":     fmt.Sprintf("< %.1f", as.darkThreshold),
			},
			Actions: []models.DeviceCommand{
				{
					DeviceID: lightDeviceID,
					Action:   "turn_on",
					Value:    nil,
					Options: map[string]interface{}{
						"automation": "motion-activated",
						"reason":     "motion detected in dark room",
					},
				},
			},
			Enabled:  true,
			Priority: 1,
			Cooldown: as.motionLightCooldown,
		}

		as.addRule(rule)
		as.logger.Printf("AutomationService: Created motion-light rule for room %s", roomID)
	}
}

// handleMotionUpdate processes motion sensor updates for automation
func (as *AutomationService) handleMotionUpdate(roomID string, occupied bool) {
	as.logger.Printf("AutomationService: Motion update - Room %s occupied: %v", roomID, occupied)

	if !occupied {
		// Room is now unoccupied - could turn off lights after delay
		as.handleRoomUnoccupied(roomID)
		return
	}

	// Room is occupied - check if we should turn on lights
	lightLevel, lightState := as.getCurrentLightLevel(roomID)

	as.logger.Printf("AutomationService: Room %s occupied, light level: %.1f%% (%s)",
		roomID, lightLevel, lightState)

	// If room is dark and motion detected, turn on lights
	if lightLevel < as.darkThreshold || lightState == "dark" {
		as.triggerMotionLighting(roomID)
	} else {
		as.logger.Printf("AutomationService: Room %s has sufficient light (%.1f%%), not turning on lights",
			roomID, lightLevel)
	}
}

// handleLightUpdate processes light sensor updates
func (as *AutomationService) handleLightUpdate(roomID string, lightState string, lightLevel float64) {
	as.logger.Printf("AutomationService: Light update - Room %s: %s (%.1f%%)",
		roomID, lightState, lightLevel)

	// Check if room is occupied and now dark - turn on lights
	if occupancy, exists := as.motionService.GetRoomOccupancy(roomID); exists && occupancy.IsOccupied {
		if lightLevel < as.darkThreshold || lightState == "dark" {
			as.logger.Printf("AutomationService: Room %s became dark while occupied, turning on lights", roomID)
			as.triggerMotionLighting(roomID)
		}
	}
}

// triggerMotionLighting turns on lights when motion is detected in dark conditions
func (as *AutomationService) triggerMotionLighting(roomID string) {
	ruleID := fmt.Sprintf("motion-light-%s", roomID)

	as.rulesMutex.RLock()
	rule, exists := as.rules[ruleID]
	as.rulesMutex.RUnlock()

	if !exists || !rule.Enabled {
		as.logger.Printf("AutomationService: No enabled motion-light rule for room %s", roomID)
		return
	}

	// Check cooldown to prevent rapid triggering
	if time.Since(rule.LastTriggered) < rule.Cooldown {
		remaining := rule.Cooldown - time.Since(rule.LastTriggered)
		as.logger.Printf("AutomationService: Rule %s on cooldown, %.0f seconds remaining",
			ruleID, remaining.Seconds())
		return
	}

	// Execute the light control action
	for _, action := range rule.Actions {
		as.logger.Printf("AutomationService: Executing action: Turn on %s (motion detected in dark room %s)",
			action.DeviceID, roomID)

		err := as.deviceService.ExecuteCommand(&action)
		if err != nil {
			as.logger.Printf("AutomationService: Failed to execute light command for room %s: %v",
				roomID, err)
		} else {
			// Send MQTT message to notify about automation
			as.publishAutomationEvent(roomID, "lights_on", "motion_detected_dark")

			// Update rule trigger time
			as.rulesMutex.Lock()
			rule.LastTriggered = time.Now()
			as.rulesMutex.Unlock()

			as.logger.Printf("AutomationService: Successfully turned on lights in room %s due to motion in dark conditions", roomID)
		}
	}
}

// handleRoomUnoccupied handles when a room becomes unoccupied
func (as *AutomationService) handleRoomUnoccupied(roomID string) {
	as.logger.Printf("AutomationService: Room %s is now unoccupied", roomID)

	// Could implement auto-off after delay, but for now just log
	// This prevents accidentally turning off lights when someone briefly leaves
	go func() {
		time.Sleep(10 * time.Minute) // Wait 10 minutes before auto-off

		// Check if room is still unoccupied
		if occupancy, exists := as.motionService.GetRoomOccupancy(roomID); exists && !occupancy.IsOccupied {
			as.logger.Printf("AutomationService: Room %s unoccupied for 10 minutes, could auto-turn off lights", roomID)
			// Could implement auto-off here
		}
	}()
}

// getCurrentLightLevel gets the current light level for a room
func (as *AutomationService) getCurrentLightLevel(roomID string) (float64, string) {
	if lightData, exists := as.lightService.GetRoomLightLevel(roomID); exists {
		return lightData.LightLevel, lightData.LightState
	}

	// Default to bright if no sensor data (assume adequate lighting)
	return 100.0, "unknown"
}

// publishAutomationEvent publishes automation events to MQTT
func (as *AutomationService) publishAutomationEvent(roomID, action, reason string) {
	event := map[string]interface{}{
		"room_id":   roomID,
		"action":    action,
		"reason":    reason,
		"timestamp": time.Now().Unix(),
		"service":   "automation",
	}

	payload, err := json.Marshal(event)
	if err != nil {
		as.logger.Printf("AutomationService: Failed to marshal automation event: %v", err)
		return
	}

	topic := fmt.Sprintf("automation/%s", roomID)
	msg := &mqtt.Message{
		Topic:   topic,
		Payload: payload,
		QoS:     1,
		Retain:  false,
	}
	err = as.mqttClient.Publish(msg)
	if err != nil {
		as.logger.Printf("AutomationService: Failed to publish automation event: %v", err)
	}
}

// addRule adds a new automation rule
func (as *AutomationService) addRule(rule *AutomationRule) {
	as.rulesMutex.Lock()
	defer as.rulesMutex.Unlock()
	as.rules[rule.ID] = rule
}

// GetRule returns a specific automation rule
func (as *AutomationService) GetRule(id string) (*AutomationRule, bool) {
	as.rulesMutex.RLock()
	defer as.rulesMutex.RUnlock()
	rule, exists := as.rules[id]
	return rule, exists
}

// GetAllRules returns all automation rules
func (as *AutomationService) GetAllRules() map[string]*AutomationRule {
	as.rulesMutex.RLock()
	defer as.rulesMutex.RUnlock()

	rules := make(map[string]*AutomationRule)
	for id, rule := range as.rules {
		rules[id] = rule
	}
	return rules
}

// EnableRule enables or disables a specific rule
func (as *AutomationService) EnableRule(id string, enabled bool) error {
	as.rulesMutex.Lock()
	defer as.rulesMutex.Unlock()

	rule, exists := as.rules[id]
	if !exists {
		return fmt.Errorf("rule %s not found", id)
	}

	rule.Enabled = enabled
	status := "disabled"
	if enabled {
		status = "enabled"
	}

	as.logger.Printf("AutomationService: Rule %s %s", id, status)
	return nil
}

// SetDarkThreshold sets the light level threshold for considering a room "dark"
func (as *AutomationService) SetDarkThreshold(threshold float64) {
	as.darkThreshold = threshold
	as.logger.Printf("AutomationService: Dark threshold set to %.1f%%", threshold)
}

// GetStatus returns the current status of the automation service
func (as *AutomationService) GetStatus() map[string]interface{} {
	as.rulesMutex.RLock()
	defer as.rulesMutex.RUnlock()

	enabledRules := 0
	totalRules := len(as.rules)

	for _, rule := range as.rules {
		if rule.Enabled {
			enabledRules++
		}
	}

	return map[string]interface{}{
		"service":         "automation",
		"status":          "active",
		"total_rules":     totalRules,
		"enabled_rules":   enabledRules,
		"dark_threshold":  as.darkThreshold,
		"motion_cooldown": as.motionLightCooldown.String(),
	}
}

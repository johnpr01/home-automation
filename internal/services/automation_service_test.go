package services

import (
	"encoding/json"
	"log"
	"os"
	"testing"
	"time"

	"github.com/johnpr01/home-automation/internal/config"
	"github.com/johnpr01/home-automation/internal/models"
	"github.com/johnpr01/home-automation/pkg/kafka"
	"github.com/johnpr01/home-automation/pkg/mqtt"
)

func TestAutomationService_MotionActivatedLighting(t *testing.T) {
	// Create test logger
	logger := log.New(os.Stdout, "[TEST] ", log.LstdFlags)

	// Create mock MQTT client
	mqttConfig := &config.MQTTConfig{
		Broker: "localhost",
		Port:   "1883",
	}
	mqttClient := mqtt.NewClient(mqttConfig, nil)

	// Create mock Kafka client
	kafkaClient := kafka.NewClient([]string{"localhost:9092"}, "test-logs", nil)

	// Create services
	motionService := NewMotionService(mqttClient, logger)
	lightService := NewLightService(mqttClient, logger)
	deviceService := NewDeviceService(mqttClient, kafkaClient)
	automationService := NewAutomationService(motionService, lightService, deviceService, mqttClient, logger)

	// Add a test light device (using one of the default room names)
	lightDevice := &models.Device{
		ID:     "light-living-room",
		Name:   "Living Room Light",
		Type:   models.DeviceTypeLight,
		Status: "off",
		Properties: map[string]interface{}{
			"power":      false,
			"brightness": 100,
			"room_id":    "living-room",
		},
		LastUpdated: time.Now(),
	}
	deviceService.AddDevice(lightDevice)

	// Test case 1: Motion detected in dark room should trigger lights
	t.Run("Motion in dark room triggers lighting", func(t *testing.T) {
		// Set room to dark
		darkLightMsg := map[string]interface{}{
			"light_level":   5.0,
			"light_percent": 5.0,
			"light_state":   "dark",
			"room":          "living-room",
			"timestamp":     time.Now().Unix(),
			"device_id":     "pico-test",
		}
		darkPayload, err := json.Marshal(darkLightMsg)
		lightService.handleLightMessage("room-light/living-room", darkPayload)

		// Give light service time to process
		time.Sleep(100 * time.Millisecond)

		// Trigger motion
		motionMsg := map[string]interface{}{
			"motion":    true,
			"room":      "living-room",
			"timestamp": time.Now().Unix(),
			"device_id": "pico-test",
		}
		motionPayload, err := json.Marshal(motionMsg)
		motionService.handleMotionMessage("room-motion/living-room", motionPayload)

		// Give automation service time to process
		time.Sleep(100 * time.Millisecond)

		// Check if light was turned on
		device, err := deviceService.GetDevice("light-living-room")
		if err != nil {
			t.Fatalf("Failed to get light device: %v", err)
		}

		if device.Status != "on" {
			t.Errorf("Expected light to be on, got %s", device.Status)
		}

		if power, ok := device.Properties["power"].(bool); !ok || !power {
			t.Errorf("Expected light power to be true, got %v", device.Properties["power"])
		}
	})

	// Test case 2: Motion detected in bright room should NOT trigger lights
	t.Run("Motion in bright room does not trigger lighting", func(t *testing.T) {
		// Reset light device
		lightDevice.Status = "off"
		lightDevice.Properties["power"] = false

		// Set room to bright
		brightLightMsg := map[string]interface{}{
			"light_level":   85.0,
			"light_percent": 85.0,
			"light_state":   "bright",
			"room":          "living-room",
			"timestamp":     time.Now().Unix(),
			"device_id":     "pico-test",
		}
		brightPayload, err := json.Marshal(brightLightMsg)
		lightService.handleLightMessage("room-light/living-room", brightPayload)

		// Give light service time to process
		time.Sleep(100 * time.Millisecond)

		// Trigger motion
		motionMsg := map[string]interface{}{
			"motion":    true,
			"room":      "living-room",
			"timestamp": time.Now().Unix(),
			"device_id": "pico-test",
		}
		motionPayload, err := json.Marshal(motionMsg)
		motionService.handleMotionMessage("room-motion/living-room", motionPayload)

		// Give automation service time to process
		time.Sleep(100 * time.Millisecond)

		// Check if light remained off
		device, err := deviceService.GetDevice("light-living-room")
		if err != nil {
			t.Fatalf("Failed to get light device: %v", err)
		}

		if device.Status != "off" {
			t.Errorf("Expected light to remain off in bright room, got %s", device.Status)
		}

		if power, ok := device.Properties["power"].(bool); ok && power {
			t.Errorf("Expected light power to remain false in bright room, got %v", device.Properties["power"])
		}
	})

	// Test case 3: Automation service status
	t.Run("Automation service returns correct status", func(t *testing.T) {
		status := automationService.GetStatus()

		if status["service"] != "automation" {
			t.Errorf("Expected service to be 'automation', got %v", status["service"])
		}

		if status["status"] != "active" {
			t.Errorf("Expected status to be 'active', got %v", status["status"])
		}

		totalRules, ok := status["total_rules"].(int)
		if !ok || totalRules <= 0 {
			t.Errorf("Expected total_rules to be positive integer, got %v", status["total_rules"])
		}

		enabledRules, ok := status["enabled_rules"].(int)
		if !ok || enabledRules <= 0 {
			t.Errorf("Expected enabled_rules to be positive integer, got %v", status["enabled_rules"])
		}

		darkThreshold, ok := status["dark_threshold"].(float64)
		if !ok || darkThreshold <= 0 {
			t.Errorf("Expected dark_threshold to be positive float, got %v", status["dark_threshold"])
		}
	})

	// Test case 4: Rule management
	t.Run("Rule management functions work correctly", func(t *testing.T) {
		// Get all rules
		rules := automationService.GetAllRules()
		if len(rules) == 0 {
			t.Error("Expected at least one rule to exist")
		}

		// Find a rule to test with
		var testRuleID string
		for id := range rules {
			testRuleID = id
			break
		}

		// Test getting specific rule
		rule, exists := automationService.GetRule(testRuleID)
		if !exists {
			t.Errorf("Expected rule %s to exist", testRuleID)
		}

		if rule.ID != testRuleID {
			t.Errorf("Expected rule ID to be %s, got %s", testRuleID, rule.ID)
		}

		// Test disabling rule
		err := automationService.EnableRule(testRuleID, false)
		if err != nil {
			t.Errorf("Failed to disable rule: %v", err)
		}

		rule, _ = automationService.GetRule(testRuleID)
		if rule.Enabled {
			t.Error("Expected rule to be disabled")
		}

		// Test re-enabling rule
		err = automationService.EnableRule(testRuleID, true)
		if err != nil {
			t.Errorf("Failed to enable rule: %v", err)
		}

		rule, _ = automationService.GetRule(testRuleID)
		if !rule.Enabled {
			t.Error("Expected rule to be enabled")
		}
	})

	// Test case 5: Dark threshold configuration
	t.Run("Dark threshold can be configured", func(t *testing.T) {
		// Set new threshold
		newThreshold := 15.0
		automationService.SetDarkThreshold(newThreshold)

		// Check status reflects new threshold
		status := automationService.GetStatus()
		if threshold, ok := status["dark_threshold"].(float64); !ok || threshold != newThreshold {
			t.Errorf("Expected dark_threshold to be %.1f, got %v", newThreshold, status["dark_threshold"])
		}
	})
}

func TestAutomationService_CooldownLogic(t *testing.T) {
	// Test cooldown logic to ensure lights don't rapidly cycle
	logger := log.New(os.Stdout, "[TEST-COOLDOWN] ", log.LstdFlags)

	mqttConfig := &config.MQTTConfig{
		Broker: "localhost",
		Port:   "1883",
	}
	mqttClient := mqtt.NewClient(mqttConfig, nil)
	kafkaClient := kafka.NewClient([]string{"localhost:9092"}, "test-logs", nil)

	motionService := NewMotionService(mqttClient, logger)
	lightService := NewLightService(mqttClient, logger)
	deviceService := NewDeviceService(mqttClient, kafkaClient)
	automationService := NewAutomationService(motionService, lightService, deviceService, mqttClient, logger)

	// Add test light device
	lightDevice := &models.Device{
		ID:     "light-kitchen",
		Name:   "Kitchen Test Light",
		Type:   models.DeviceTypeLight,
		Status: "off",
		Properties: map[string]interface{}{
			"power":   false,
			"room_id": "kitchen",
		},
		LastUpdated: time.Now(),
	}
	deviceService.AddDevice(lightDevice)

	t.Run("Cooldown prevents rapid triggering", func(t *testing.T) {
		// Set room to dark
		darkLightMsg := map[string]interface{}{
			"light_level":   5.0,
			"light_percent": 5.0,
			"light_state":   "dark",
			"room":          "kitchen",
			"timestamp":     time.Now().Unix(),
			"device_id":     "pico-cooldown",
		}
		darkPayload, err := json.Marshal(darkLightMsg)
		if err != nil {
			t.Logf("Failed to marshal dark light message: %v", err)
		}
		lightService.handleLightMessage("room-light/kitchen", darkPayload)
		time.Sleep(50 * time.Millisecond)

		// First motion trigger should work
		motionMsg := map[string]interface{}{
			"motion":    true,
			"room":      "kitchen",
			"timestamp": time.Now().Unix(),
			"device_id": "pico-cooldown",
		}
		motionPayload, err := json.Marshal(motionMsg)
		motionService.handleMotionMessage("room-motion/kitchen", motionPayload)
		time.Sleep(50 * time.Millisecond)

		// Verify light turned on
		device, _ := deviceService.GetDevice("light-kitchen")
		if device.Status != "on" {
			t.Error("Expected light to turn on from first motion trigger")
		}

		// Reset light for second test
		lightDevice.Status = "off"
		lightDevice.Properties["power"] = false

		// Second motion trigger immediately should be ignored due to cooldown
		motionService.handleMotionMessage("room-motion/kitchen", motionPayload)
		time.Sleep(50 * time.Millisecond)

		// Verify light stayed off due to cooldown
		device, _ = deviceService.GetDevice("light-kitchen")
		if device.Status != "off" {
			t.Error("Expected light to stay off due to cooldown logic")
		}

		// Verify cooldown is working by checking automation service status
		status := automationService.GetStatus()
		if status["service"] != "automation" {
			t.Error("Automation service should be active during cooldown test")
		}
	})
}

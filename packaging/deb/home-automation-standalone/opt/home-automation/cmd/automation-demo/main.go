package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/johnpr01/home-automation/internal/config"
	"github.com/johnpr01/home-automation/internal/models"
	"github.com/johnpr01/home-automation/internal/services"
	"github.com/johnpr01/home-automation/pkg/kafka"
	"github.com/johnpr01/home-automation/pkg/mqtt"
)

func main() {
	logger := log.New(os.Stdout, "[AUTOMATION-DEMO] ", log.LstdFlags|log.Lshortfile)
	logger.Println("🚀 Starting Motion-Activated Lighting Demo...")

	// Create MQTT client
	mqttConfig := &config.MQTTConfig{
		Broker:   "localhost",
		Port:     "1883",
		Username: "",
		Password: "",
	}
	mqttClient := mqtt.NewClient(mqttConfig, nil)
	if err := mqttClient.Connect(); err != nil {
		logger.Fatalf("Failed to connect to MQTT broker: %v", err)
	}
	defer mqttClient.Disconnect()

	// Create Kafka client
	kafkaClient := kafka.NewClient([]string{"localhost:9092"}, "home-automation-logs", nil)

	// Create services
	motionService := services.NewMotionService(mqttClient, logger)
	lightService := services.NewLightService(mqttClient, logger)
	deviceService := services.NewDeviceService(mqttClient, kafkaClient)
	automationService := services.NewAutomationService(motionService, lightService, deviceService, mqttClient, logger)

	// Add light devices for the demo
	rooms := []string{"living-room", "kitchen", "bedroom"}
	for _, roomID := range rooms {
		lightDevice := &models.Device{
			ID:     fmt.Sprintf("light-%s", roomID),
			Name:   fmt.Sprintf("%s Light", roomID),
			Type:   models.DeviceTypeLight,
			Status: "off",
			Properties: map[string]interface{}{
				"power":      false,
				"brightness": 100,
				"room_id":    roomID,
			},
			LastUpdated: time.Now(),
		}
		deviceService.AddDevice(lightDevice)
		logger.Printf("✅ Added light device: %s", lightDevice.Name)
	}

	// Log automation status
	status := automationService.GetStatus()
	logger.Printf("🤖 Automation Service: %d rules enabled, dark threshold: %.1f%%",
		status["enabled_rules"], status["dark_threshold"])

	logger.Println("\n🏠 Motion-Activated Lighting Demo Ready!")
	logger.Println("📋 Test Scenarios:")
	logger.Println("   1. Send motion detected + dark light level → Lights turn on")
	logger.Println("   2. Send motion detected + bright light level → Lights stay off")
	logger.Println("   3. Send no motion → Lights could auto-turn off after delay")
	logger.Println("\n🧪 Test Commands:")
	logger.Printf("   Motion: mosquitto_pub -h localhost -t 'room-motion/living-room' -m '{\"motion\":true,\"room\":\"living-room\",\"timestamp\":%d,\"device_id\":\"pico-living\"}'", time.Now().Unix())
	logger.Printf("   Dark:   mosquitto_pub -h localhost -t 'room-light/living-room' -m '{\"light_level\":5.0,\"light_percent\":5.0,\"light_state\":\"dark\",\"room\":\"living-room\",\"timestamp\":%d,\"device_id\":\"pico-living\"}'", time.Now().Unix())
	logger.Printf("   Bright: mosquitto_pub -h localhost -t 'room-light/living-room' -m '{\"light_level\":80.0,\"light_percent\":80.0,\"light_state\":\"bright\",\"room\":\"living-room\",\"timestamp\":%d,\"device_id\":\"pico-living\"}'", time.Now().Unix())

	// Simulate some sensor data for testing
	go func() {
		time.Sleep(3 * time.Second)
		logger.Println("\n🧪 Simulating sensor data for demo...")

		// Simulate dark room first
		lightMsg := `{"light_level":8.0,"light_percent":8.0,"light_state":"dark","room":"living-room","timestamp":` + fmt.Sprintf("%d", time.Now().Unix()) + `,"device_id":"pico-living-demo"}`
		mqttClient.Publish(&mqtt.Message{
			Topic:   "room-light/living-room",
			Payload: []byte(lightMsg),
			QoS:     1,
			Retain:  false,
		})
		logger.Println("📡 Simulated: Living room is DARK (8% light)")

		time.Sleep(2 * time.Second)

		// Simulate motion detection
		motionMsg := `{"motion":true,"room":"living-room","timestamp":` + fmt.Sprintf("%d", time.Now().Unix()) + `,"device_id":"pico-living-demo"}`
		mqttClient.Publish(&mqtt.Message{
			Topic:   "room-motion/living-room",
			Payload: []byte(motionMsg),
			QoS:     1,
			Retain:  false,
		})
		logger.Println("📡 Simulated: MOTION DETECTED in living room")
		logger.Println("💡 Expected: Lights should turn ON (motion + dark = automation trigger)")

		time.Sleep(5 * time.Second)

		// Test with bright light condition
		logger.Println("\n🧪 Testing bright room scenario...")
		brightMsg := `{"light_level":85.0,"light_percent":85.0,"light_state":"bright","room":"kitchen","timestamp":` + fmt.Sprintf("%d", time.Now().Unix()) + `,"device_id":"pico-kitchen-demo"}`
		mqttClient.Publish(&mqtt.Message{
			Topic:   "room-light/kitchen",
			Payload: []byte(brightMsg),
			QoS:     1,
			Retain:  false,
		})
		logger.Println("📡 Simulated: Kitchen is BRIGHT (85% light)")

		time.Sleep(2 * time.Second)

		motionMsg2 := `{"motion":true,"room":"kitchen","timestamp":` + fmt.Sprintf("%d", time.Now().Unix()) + `,"device_id":"pico-kitchen-demo"}`
		mqttClient.Publish(&mqtt.Message{
			Topic:   "room-motion/kitchen",
			Payload: []byte(motionMsg2),
			QoS:     1,
			Retain:  false,
		})
		logger.Println("📡 Simulated: MOTION DETECTED in kitchen")
		logger.Println("💡 Expected: Lights should stay OFF (motion + bright = no automation)")
	}()

	// Keep running to observe automation
	logger.Println("\n⏳ Demo running... Press Ctrl+C to stop")
	select {}
}

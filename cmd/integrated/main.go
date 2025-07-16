package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/johnpr01/home-automation/internal/config"
	"github.com/johnpr01/home-automation/internal/models"
	"github.com/johnpr01/home-automation/internal/services"
	"github.com/johnpr01/home-automation/pkg/kafka"
	"github.com/johnpr01/home-automation/pkg/mqtt"
	"github.com/johnpr01/home-automation/pkg/utils"
)

func main() {
	logger := log.New(os.Stdout, "[INTEGRATED] ", log.LstdFlags|log.Lshortfile)
	logger.Println("Starting Integrated Home Automation Service...")

	// Load MQTT configuration
	mqttConfig := &config.MQTTConfig{
		Broker:   "localhost",
		Port:     "1883",
		Username: "",
		Password: "",
	}

	// Create MQTT client
	mqttClient := mqtt.NewClient(mqttConfig)
	if err := mqttClient.Connect(); err != nil {
		logger.Fatalf("Failed to connect to MQTT broker: %v", err)
	}
	defer mqttClient.Disconnect()

	logger.Println("Connected to MQTT broker")

	// Create Kafka client for device service logging
	kafkaClient := kafka.NewClient([]string{"localhost:9092"}, "home-automation-logs")

	// Create independent services
	motionService := services.NewMotionService(mqttClient, logger)
	lightService := services.NewLightService(mqttClient, logger)
	thermostatService := services.NewThermostatService(mqttClient, logger)
	deviceService := services.NewDeviceService(mqttClient, kafkaClient)

	// Create automation service that coordinates between sensors and devices
	automationService := services.NewAutomationService(motionService, lightService, deviceService, mqttClient, logger)

	logger.Println("🏠 Automation Service: Motion-activated lighting enabled!")
	logger.Println("📋 Rules: When motion detected + dark conditions → Turn on lights")

	// Log automation service status
	status := automationService.GetStatus()
	logger.Printf("Automation Service Status: %d rules enabled, dark threshold: %.1f%%",
		status["enabled_rules"], status["dark_threshold"])

	// Register sample thermostat
	thermostat := &models.Thermostat{
		ID:             "thermostat-001",
		Name:           "Living Room Thermostat",
		RoomID:         "1",
		TargetTemp:     72.0, // 72°F
		Mode:           models.ModeAuto,
		Hysteresis:     utils.DefaultHysteresis, // 1°F
		MinTemp:        utils.DefaultMinTemp,    // 45°F
		MaxTemp:        utils.DefaultMaxTemp,    // 90°F
		IsOnline:       true,
		Status:         models.StatusIdle,
		HeatingEnabled: true,
		CoolingEnabled: true,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	thermostatService.RegisterThermostat(thermostat)

	// Register light devices for automation
	rooms := []struct {
		id   string
		name string
	}{
		{"living-room", "Living Room"},
		{"kitchen", "Kitchen"},
		{"bedroom", "Bedroom"},
		{"bathroom", "Bathroom"},
		{"office", "Office"},
		{"hallway", "Hallway"},
	}

	for _, room := range rooms {
		lightDevice := &models.Device{
			ID:     fmt.Sprintf("light-%s", room.id),
			Name:   fmt.Sprintf("%s Light", room.name),
			Type:   models.DeviceTypeLight,
			Status: "off",
			Properties: map[string]interface{}{
				"power":      false,
				"brightness": 100,
				"room_id":    room.id,
			},
			LastUpdated: time.Now(),
		}

		err := deviceService.AddDevice(lightDevice)
		if err != nil {
			logger.Printf("Failed to add light device for %s: %v", room.name, err)
		} else {
			logger.Printf("Added light device: %s", lightDevice.Name)
		}
	}

	// Optional: Create integration between services for additional monitoring
	// The AutomationService already handles motion + light → light control
	// These callbacks provide additional logging for monitoring
	motionService.AddOccupancyCallback(func(roomID string, occupied bool) {
		status := "UNOCCUPIED"
		if occupied {
			status = "OCCUPIED"
		}
		logger.Printf("Integration Monitor: Room %s is now %s", roomID, status)
	})

	// Optional: Additional light level monitoring
	lightService.AddLightCallback(func(roomID string, lightState string, lightLevel float64) {
		logger.Printf("Integration Monitor: Room %s light level: %s (%.1f%%)", roomID, lightState, lightLevel)
	})

	// Start periodic status reporting
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()

		for range ticker.C {
			// Motion service status
			motionSummary := motionService.GetMotionSummary()
			logger.Printf("Motion Summary: %d rooms total, %d occupied, %d sensors online",
				motionSummary["total_rooms"], motionSummary["occupied_rooms"], motionSummary["online_sensors"])

			// Light service status
			lightSummary := lightService.GetLightSummary()
			logger.Printf("Light Summary: %d rooms total, %d dark, %d bright, avg %.1f%%",
				lightSummary["total_rooms"], lightSummary["dark_rooms"],
				lightSummary["bright_rooms"], lightSummary["average_light_level"])

			// Thermostat service status
			thermostats := thermostatService.GetAllThermostats()
			logger.Printf("Thermostat Summary: %d thermostats registered", len(thermostats))
		}
	}()

	logger.Println("Integrated home automation service started successfully")
	logger.Println("Running independent Motion Detection, Light Sensor, and Thermostat services")

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	logger.Println("Shutting down integrated home automation service...")
}

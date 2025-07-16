package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/johnpr01/home-automation/internal/config"
	"github.com/johnpr01/home-automation/internal/models"
	"github.com/johnpr01/home-automation/internal/services"
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

	// Create independent services
	motionService := services.NewMotionService(mqttClient, logger)
	lightService := services.NewLightService(mqttClient, logger)
	thermostatService := services.NewThermostatService(mqttClient, logger)

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

	// Optional: Create integration between services
	// This shows how the services can communicate if needed, but they remain independent
	motionService.AddOccupancyCallback(func(roomID string, occupied bool) {
		status := "UNOCCUPIED"
		if occupied {
			status = "OCCUPIED"
		}
		logger.Printf("Integration: Room %s is now %s", roomID, status)

		// Example: You could adjust thermostat behavior based on occupancy
		// For now, just log the occupancy change
		if occupancy, exists := motionService.GetRoomOccupancy(roomID); exists {
			logger.Printf("Integration: Room %s occupancy details - Device: %s, Online: %v",
				roomID, occupancy.DeviceID, occupancy.IsOnline)
		}
	})

	// Optional: Light level integration for automation
	lightService.AddLightCallback(func(roomID string, lightState string, lightLevel float64) {
		logger.Printf("Integration: Room %s light level changed to %s (%.1f%%)", roomID, lightState, lightLevel)

		// Example automation based on light and occupancy
		if occupancy, exists := motionService.GetRoomOccupancy(roomID); exists && occupancy.IsOccupied {
			switch lightState {
			case "dark":
				logger.Printf("Integration: Room %s is occupied and dark - could turn on lights", roomID)
			case "bright":
				logger.Printf("Integration: Room %s is occupied and bright - could turn off lights", roomID)
			}
		}
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

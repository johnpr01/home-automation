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
)

func main() {
	logger := log.New(os.Stdout, "[THERMOSTAT] ", log.LstdFlags|log.Lshortfile)
	logger.Println("Starting Home Automation Thermostat Service...")

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

	// Create thermostat service
	thermostatService := services.NewThermostatService(mqttClient, logger)

	// Register a sample thermostat for room 1 (using Fahrenheit)
	sampleThermostat := &models.Thermostat{
		ID:                "thermostat-001",
		Name:              "Living Room Thermostat",
		RoomID:            "1",
		CurrentTemp:       68.0, // 68°F (20°C)
		CurrentHumidity:   50.0,
		TargetTemp:        72.0, // 72°F (22°C)
		Mode:              models.ModeAuto,
		Status:            models.StatusIdle,
		FanSpeed:          50,
		HeatingEnabled:    true,
		CoolingEnabled:    true,
		TemperatureOffset: 0.0,
		Hysteresis:        2.0,  // 2°F hysteresis (1.1°C)
		MinTemp:           50.0, // 50°F (10°C)
		MaxTemp:           86.0, // 86°F (30°C)
		LastSensorUpdate:  time.Now(),
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
		IsOnline:          true,
	}

	thermostatService.RegisterThermostat(sampleThermostat)
	logger.Printf("Registered thermostat: %s for room %s", sampleThermostat.ID, sampleThermostat.RoomID)

	// Set up graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	logger.Println("Thermostat service is running...")
	logger.Println("Listening for sensor data on topics:")
	logger.Println("  - room-temp/+ (temperature readings)")
	logger.Println("  - room-hum/+ (humidity readings)")
	logger.Println("Press Ctrl+C to stop")

	// Wait for shutdown signal
	<-sigChan
	logger.Println("Shutting down gracefully...")
}

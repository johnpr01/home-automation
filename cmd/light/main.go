package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/johnpr01/home-automation/internal/config"
	"github.com/johnpr01/home-automation/internal/services"
	"github.com/johnpr01/home-automation/pkg/mqtt"
)

func main() {
	// Create logger
	logger := log.New(os.Stdout, "[LightService] ", log.LstdFlags|log.Lshortfile)

	// Load MQTT configuration
	mqttConfig := &config.MQTTConfig{
		Broker:   "localhost",
		Port:     "1883",
		Username: "",
		Password: "",
	}

	// Initialize MQTT client
	mqttClient := mqtt.NewClient(mqttConfig, nil)

	// Connect to MQTT broker
	if err := mqttClient.Connect(); err != nil {
		logger.Fatalf("Failed to connect to MQTT broker: %v", err)
	}
	defer mqttClient.Disconnect()

	logger.Println("Connected to MQTT broker")

	// Create light sensor service
	lightService := services.NewLightService(mqttClient, logger)

	// Set custom thresholds if needed (optional)
	lightService.SetThresholds(15.0, 75.0) // dark < 15%, bright > 75%

	// Add example callback for light level changes
	lightService.AddLightCallback(func(roomID string, lightState string, lightLevel float64) {
		logger.Printf("Room %s light changed to %s (%.1f%%)", roomID, lightState, lightLevel)

		// Example automation based on light levels
		switch lightState {
		case "dark":
			logger.Printf("Room %s is now DARK - could trigger evening automation", roomID)
		case "bright":
			logger.Printf("Room %s is now BRIGHT - could trigger morning automation", roomID)
		case "normal":
			logger.Printf("Room %s has NORMAL lighting", roomID)
		}
	})

	// Start periodic status reporting
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()

		for range ticker.C {
			summary := lightService.GetLightSummary()
			logger.Printf("Light Summary: %d rooms total, %d dark, %d bright, %d sensors online, avg light: %.1f%%",
				summary["total_rooms"], summary["dark_rooms"], summary["bright_rooms"],
				summary["online_sensors"], summary["average_light_level"])
		}
	}()

	logger.Println("Light sensor service started successfully")
	logger.Println("Monitoring MQTT topics: room-light/+")
	logger.Println("Light thresholds: dark < 15%, bright > 75%")

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	logger.Println("Shutting down light sensor service...")
}

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
	logger := log.New(os.Stdout, "[MotionService] ", log.LstdFlags|log.Lshortfile)

	// Load MQTT configuration
	mqttConfig := &config.MQTTConfig{
		Broker:   "localhost",
		Port:     "1883",
		Username: "",
		Password: "",
	}

	// Initialize MQTT client
	mqttClient := mqtt.NewClient(mqttConfig)

	// Connect to MQTT broker
	if err := mqttClient.Connect(); err != nil {
		logger.Fatalf("Failed to connect to MQTT broker: %v", err)
	}
	defer mqttClient.Disconnect()

	logger.Println("Connected to MQTT broker")

	// Create motion detection service
	motionService := services.NewMotionService(mqttClient, logger)

	// Add example callback for occupancy changes
	motionService.AddOccupancyCallback(func(roomID string, occupied bool) {
		status := "UNOCCUPIED"
		if occupied {
			status = "OCCUPIED"
		}
		logger.Printf("Room %s is now %s", roomID, status)
	})

	// Start periodic status reporting
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()

		for range ticker.C {
			summary := motionService.GetMotionSummary()
			logger.Printf("Motion Summary: %d rooms total, %d occupied, %d sensors online",
				summary["total_rooms"], summary["occupied_rooms"], summary["online_sensors"])
		}
	}()

	logger.Println("Motion detection service started successfully")
	logger.Println("Monitoring MQTT topics: room-motion/+")

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	logger.Println("Shutting down motion detection service...")
}

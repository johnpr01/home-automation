package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/johnpr01/home-automation/internal/config"
	"github.com/johnpr01/home-automation/internal/logger"
	"github.com/johnpr01/home-automation/internal/services"
	"github.com/johnpr01/home-automation/pkg/influxdb"
	"github.com/johnpr01/home-automation/pkg/mqtt"
)

func main() {
	// Initialize context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize InfluxDB client
	influxClient := influxdb.NewClient(
		"http://localhost:8086",
		"home-automation-token",
		"home-automation",
		"sensor-data",
	)

	if influxClient != nil {
		if err := influxClient.Connect(); err != nil {
			// Log error but continue - InfluxDB is optional
			influxClient = nil
		}
		defer func() {
			if influxClient != nil {
				influxClient.Disconnect()
			}
		}()
	}

	// Initialize logger
	serviceLogger := logger.NewLogger("tapo-service", nil)
	serviceLogger.Info("Starting Tapo Smart Plug Monitoring Service")

	// Initialize MQTT client
	mqttConfig := &config.MQTTConfig{
		Broker:   "localhost",
		Port:     "1883",
		Username: "",
		Password: "",
	}

	mqttClient := mqtt.NewClient(mqttConfig, nil)
	if err := mqttClient.Connect(); err != nil {
		serviceLogger.Error("Failed to connect to MQTT broker", err)
		return
	}
	defer mqttClient.Disconnect()

	serviceLogger.Info("Connected to MQTT broker")

	// Create Tapo service
	tapoService := services.NewTapoService(mqttClient, influxClient, serviceLogger)

	// Add example Tapo devices (replace with your actual device details)
	exampleDevices := []*services.TapoConfig{
		{
			DeviceID:     "tapo_living_room_1",
			DeviceName:   "Living Room Lamp",
			RoomID:       "living_room",
			IPAddress:    "192.168.1.100",      // Replace with your device IP
			Username:     "your_tapo_username", // Replace with your Tapo account username
			Password:     "your_tapo_password", // Replace with your Tapo account password
			PollInterval: 30 * time.Second,
		},
		{
			DeviceID:     "tapo_kitchen_1",
			DeviceName:   "Kitchen Coffee Maker",
			RoomID:       "kitchen",
			IPAddress:    "192.168.1.101",      // Replace with your device IP
			Username:     "your_tapo_username", // Replace with your Tapo account username
			Password:     "your_tapo_password", // Replace with your Tapo account password
			PollInterval: 30 * time.Second,
		},
		{
			DeviceID:     "tapo_office_1",
			DeviceName:   "Office Monitor",
			RoomID:       "office",
			IPAddress:    "192.168.1.102",      // Replace with your device IP
			Username:     "your_tapo_username", // Replace with your Tapo account username
			Password:     "your_tapo_password", // Replace with your Tapo account password
			PollInterval: 60 * time.Second,
		},
	}

	// Add devices to service
	for _, deviceConfig := range exampleDevices {
		if err := tapoService.AddDevice(deviceConfig); err != nil {
			serviceLogger.Error("Failed to add Tapo device", err, map[string]interface{}{
				"device_id": deviceConfig.DeviceID,
			})
		}
	}

	// Start monitoring
	if err := tapoService.Start(); err != nil {
		serviceLogger.Error("Failed to start Tapo service", err)
		return
	}

	// Setup graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	serviceLogger.Info("Tapo monitoring service started successfully")
	serviceLogger.Info("Monitoring energy consumption for smart plugs")
	serviceLogger.Info("Data is being stored in InfluxDB and published to MQTT")
	serviceLogger.Info("Press Ctrl+C to stop...")

	// Main monitoring loop
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-sigChan:
			serviceLogger.Info("Received shutdown signal, stopping service...")
			tapoService.Stop()
			return

		case <-ticker.C:
			// Log service status
			status := tapoService.GetDeviceStatus()
			serviceLogger.Info("Tapo service status", status)

		case <-ctx.Done():
			serviceLogger.Info("Context cancelled, stopping service...")
			tapoService.Stop()
			return
		}
	}
}

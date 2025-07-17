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
	"github.com/johnpr01/home-automation/pkg/mqtt"
	"github.com/johnpr01/home-automation/pkg/prometheus"
)

func main() {
	// Initialize context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize Prometheus client
	prometheusClient := prometheus.NewClient("http://localhost:9090")

	if prometheusClient != nil {
		if err := prometheusClient.Connect(); err != nil {
			// Log error but continue - Prometheus is optional
			prometheusClient = nil
		}
		defer func() {
			if prometheusClient != nil {
				prometheusClient.Disconnect()
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
	tapoService := services.NewTapoService(mqttClient, prometheusClient, serviceLogger)

	// Get password from environment variable (GitHub Actions secret)
	tplinkPassword := os.Getenv("TPLINK_PASSWORD")
	if tplinkPassword == "" {
		serviceLogger.Error("TPLINK_PASSWORD environment variable not set", nil)
		return
	}

	// Add example Tapo devices (replace with your actual device details)
	exampleDevices := []*services.TapoConfig{
		{
			DeviceID:     "dryer",
			DeviceName:   "dryer",
			RoomID:       "laundry_room",
			IPAddress:    "192.168.68.54",      // Replace with your device IP
			Username:     "johnpr01@gmail.com", // Replace with your Tapo account username
			Password:     tplinkPassword,       // Using environment variable from GitHub Actions secret
			PollInterval: 30 * time.Second,
			UseKlap:      true, // Enable KLAP protocol for newer firmware (1.1.0+)
		},
		{
			DeviceID:     "wasjhing_machine",
			DeviceName:   "Washing Machine",
			RoomID:       "laundry_room",
			IPAddress:    "192.168.68.53",      // Replace with your device IP
			Username:     "johnpr01@gmail.com", // Replace with your Tapo account username
			Password:     tplinkPassword,       // Using environment variable from GitHub Actions secret
			PollInterval: 30 * time.Second,
			UseKlap:      true, // Enable KLAP protocol for newer firmware
		},
		{
			DeviceID:     "Hi-Fi Power",
			DeviceName:   "Hi-Fi Power",
			RoomID:       "den",
			IPAddress:    "192.168.68.60",      // Replace with your device IP
			Username:     "johnpr01@gmail.com", // Replace with your Tapo account username
			Password:     tplinkPassword,       // Using environment variable from GitHub Actions secret
			PollInterval: 60 * time.Second,
			UseKlap:      false, // Use legacy protocol for testing compatibility
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

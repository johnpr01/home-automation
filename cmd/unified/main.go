package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/johnpr01/home-automation/internal/config"
	"github.com/johnpr01/home-automation/internal/services"
	"github.com/johnpr01/home-automation/pkg/mqtt"
)

// HomeAutomationSystem coordinates all home automation services
type HomeAutomationSystem struct {
	unifiedSensorService *services.UnifiedSensorService
	thermostatService    *services.ThermostatService
	mqttClient           *mqtt.Client
	logger               *log.Logger
	ctx                  context.Context
	cancel               context.CancelFunc
}

func main() {
	// Create logger
	logger := log.New(os.Stdout, "[HOME-AUTO] ", log.LstdFlags|log.Lshortfile)
	logger.Println("Starting Home Automation System...")

	// Initialize MQTT client
	mqttConfig := &config.MQTTConfig{
		Broker:   "localhost",
		Port:     "1883",
		Username: "",
		Password: "",
	}
	mqttClient := mqtt.NewClient(mqttConfig, nil)

	// Connect to MQTT broker
	if err := mqttClient.Connect(); err != nil {
		logger.Fatalf("Failed to connect to MQTT broker: %v", err)
	}
	defer mqttClient.Disconnect()

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize home automation system
	homeSystem := &HomeAutomationSystem{
		mqttClient: mqttClient,
		logger:     logger,
		ctx:        ctx,
		cancel:     cancel,
	}

	// Start all services
	if err := homeSystem.initializeServices(); err != nil {
		logger.Fatalf("Failed to initialize services: %v", err)
	}

	// Start system monitoring
	go homeSystem.startSystemMonitoring()

	// Start sensor data analysis
	go homeSystem.startSensorAnalysis()

	// Setup graceful shutdown
	homeSystem.setupGracefulShutdown()

	logger.Println("Home Automation System started successfully")
	logger.Println("Press Ctrl+C to shutdown...")

	// Wait for shutdown signal
	<-homeSystem.ctx.Done()
	logger.Println("Home Automation System shutting down...")
}

// initializeServices sets up all home automation services
func (has *HomeAutomationSystem) initializeServices() error {
	// Initialize unified sensor service
	has.unifiedSensorService = services.NewUnifiedSensorService(has.mqttClient, has.logger)

	// Initialize thermostat service
	has.thermostatService = services.NewThermostatService(has.mqttClient, has.logger)

	// Connect sensor service to thermostat service
	has.unifiedSensorService.AddTemperatureCallback(has.thermostatService.HandleTemperatureUpdate)
	has.unifiedSensorService.AddMotionCallback(has.handleMotionUpdate)
	has.unifiedSensorService.AddLightCallback(has.handleLightUpdate)

	has.logger.Println("All services initialized successfully")
	return nil
}

// handleMotionUpdate processes motion sensor updates for automation
func (has *HomeAutomationSystem) handleMotionUpdate(roomID string, occupied bool) {
	has.logger.Printf("Motion automation: Room %s is %s", roomID, map[bool]string{true: "occupied", false: "unoccupied"}[occupied])

	// Example automation logic:
	// - Adjust thermostat setpoints based on occupancy
	// - Trigger lighting automation
	// - Log occupancy patterns

	if occupied {
		// Room became occupied - could increase thermostat target
		has.logger.Printf("Room %s occupied - considering thermostat adjustment", roomID)
	} else {
		// Room became unoccupied - could decrease thermostat target after delay
		has.logger.Printf("Room %s unoccupied - considering energy saving mode", roomID)
	}
}

// handleLightUpdate processes light sensor updates for automation
func (has *HomeAutomationSystem) handleLightUpdate(roomID string, lightState string, lightLevel float64) {
	has.logger.Printf("Light automation: Room %s light is %s (%.1f%%)", roomID, lightState, lightLevel)

	// Example automation logic:
	// - Adjust thermostat behavior based on natural light
	// - Control automated blinds/curtains
	// - Adapt HVAC scheduling to day/night cycles

	if lightState == "dark" && lightLevel < 10.0 {
		has.logger.Printf("Room %s is dark - considering night mode adjustments", roomID)
	} else if lightState == "bright" && lightLevel > 80.0 {
		has.logger.Printf("Room %s is bright - considering day mode adjustments", roomID)
	}
}

// startSystemMonitoring runs periodic system health checks
func (has *HomeAutomationSystem) startSystemMonitoring() {
	ticker := time.NewTicker(2 * time.Minute)
	defer ticker.Stop()

	has.logger.Println("System monitoring started")

	for {
		select {
		case <-has.ctx.Done():
			has.logger.Println("System monitoring stopping...")
			return
		case <-ticker.C:
			has.performSystemHealthCheck()
		}
	}
}

// performSystemHealthCheck checks system health and logs status
func (has *HomeAutomationSystem) performSystemHealthCheck() {
	summary := has.unifiedSensorService.GetSensorSummary()

	totalRooms := summary["total_rooms"].(int)
	onlineDevices := summary["online_devices"].(int)
	occupiedRooms := summary["occupied_rooms"].(int)

	has.logger.Printf("System Health: %d/%d devices online, %d/%d rooms occupied",
		onlineDevices, totalRooms, occupiedRooms, totalRooms)

	if avgTemp, ok := summary["average_temperature"].(float64); ok && avgTemp > 0 {
		has.logger.Printf("Average conditions: %.1f°F, %.1f%% humidity, %.1f%% light",
			avgTemp,
			summary["average_humidity"].(float64),
			summary["average_light_level"].(float64))
	}

	// Check for offline devices
	if onlineDevices < totalRooms {
		has.logger.Printf("WARNING: %d devices are offline", totalRooms-onlineDevices)
	}
}

// startSensorAnalysis runs periodic analysis of sensor patterns
func (has *HomeAutomationSystem) startSensorAnalysis() {
	ticker := time.NewTicker(15 * time.Minute)
	defer ticker.Stop()

	has.logger.Println("Sensor pattern analysis started")

	for {
		select {
		case <-has.ctx.Done():
			has.logger.Println("Sensor analysis stopping...")
			return
		case <-ticker.C:
			has.analyzeSensorPatterns()
		}
	}
}

// analyzeSensorPatterns analyzes sensor data for automation insights
func (has *HomeAutomationSystem) analyzeSensorPatterns() {
	allRoomData := has.unifiedSensorService.GetAllRoomSensors()

	has.logger.Printf("Analyzing patterns across %d rooms...", len(allRoomData))

	currentTime := time.Now()
	occupancyByHour := make(map[int]int)
	tempVariations := make(map[string]float64)
	lightPatterns := make(map[string]string)

	for roomID, roomData := range allRoomData {
		if !roomData.IsOnline {
			continue
		}

		// Analyze occupancy patterns
		if roomData.IsOccupied {
			hour := currentTime.Hour()
			occupancyByHour[hour]++
		}

		// Track temperature variations
		tempVariations[roomID] = roomData.Temperature

		// Track light patterns
		lightPatterns[roomID] = roomData.DayNightCycle
	}

	// Log insights
	if len(occupancyByHour) > 0 {
		has.logger.Printf("Current hour (%d) occupancy: %d rooms", currentTime.Hour(), occupancyByHour[currentTime.Hour()])
	}

	// Check for temperature anomalies
	for roomID, temp := range tempVariations {
		if temp > 80.0 {
			has.logger.Printf("High temperature alert: Room %s is %.1f°F", roomID, temp)
		} else if temp < 60.0 {
			has.logger.Printf("Low temperature alert: Room %s is %.1f°F", roomID, temp)
		}
	}

	// Analyze light patterns
	dayCount := 0
	nightCount := 0
	for _, cycle := range lightPatterns {
		if cycle == "day" {
			dayCount++
		} else if cycle == "night" {
			nightCount++
		}
	}

	if dayCount > 0 || nightCount > 0 {
		has.logger.Printf("Light patterns: %d rooms in day cycle, %d in night cycle", dayCount, nightCount)
	}
}

// setupGracefulShutdown handles shutdown signals
func (has *HomeAutomationSystem) setupGracefulShutdown() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigChan
		has.logger.Printf("Received signal: %v", sig)
		has.logger.Println("Initiating graceful shutdown...")

		// Cancel context to stop all services
		has.cancel()
	}()
}

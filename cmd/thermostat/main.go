package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/johnpr01/home-automation/internal/config"
	"github.com/johnpr01/home-automation/internal/logger"
	"github.com/johnpr01/home-automation/internal/models"
	"github.com/johnpr01/home-automation/internal/services"
	"github.com/johnpr01/home-automation/internal/utils"
	"github.com/johnpr01/home-automation/pkg/kafka"
	"github.com/johnpr01/home-automation/pkg/mqtt"
)

func main() {
	// Initialize error handling context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize Kafka client for logging
	kafkaClient := kafka.NewClient([]string{"localhost:9092"}, "home-automation-logs", nil)
	if kafkaClient != nil {
		if err := kafkaClient.Connect(); err != nil {
			// Log error but don't fail - Kafka is optional for this service
			kafkaClient = nil
		}
		defer func() {
			if kafkaClient != nil {
				kafkaClient.Disconnect()
			}
		}()
	}

	// Initialize logger
	serviceLogger := logger.NewLogger("thermostat-service", kafkaClient)
	serviceLogger.Info("Starting Home Automation Thermostat Service")

	// Load MQTT configuration
	mqttConfig := &config.MQTTConfig{
		Broker:   "localhost",
		Port:     "1883",
		Username: "",
		Password: "",
	}

	// Create MQTT client with enhanced error handling
	retryConfig := utils.DefaultRetryConfig()
	retryConfig.MaxAttempts = 5
	circuitBreaker := utils.NewCircuitBreaker(3, 30*time.Second)

	mqttOptions := &mqtt.ClientOptions{
		RetryConfig:    retryConfig,
		CircuitBreaker: circuitBreaker,
		Logger:         serviceLogger,
	}

	mqttClient := mqtt.NewClient(mqttConfig, mqttOptions)

	// Connect with retry logic
	connectOperation := func() error {
		return mqttClient.Connect()
	}

	if err := utils.Retry(ctx, retryConfig, connectOperation); err != nil {
		serviceLogger.Fatal("Failed to connect to MQTT broker after retries", err)
	}

	defer func() {
		if err := mqttClient.Disconnect(); err != nil {
			serviceLogger.Error("Error disconnecting from MQTT broker", err)
		}
	}()

	// Create thermostat service with enhanced error handling
	thermostatService := services.NewThermostatService(mqttClient, serviceLogger)

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
	serviceLogger.Info("Registered thermostat", map[string]interface{}{
		"thermostat_id": sampleThermostat.ID,
		"room_id":       sampleThermostat.RoomID,
		"name":          sampleThermostat.Name,
	})

	// Initialize health checker
	healthChecker := utils.NewHealthChecker()
	healthChecker.RegisterCheck("mqtt_connection", func() error {
		return mqttClient.GetHealthStatus(ctx)["mqtt_connection"]
	})
	healthChecker.RegisterCheck("thermostat_service", func() error {
		// Check if thermostat is responsive
		_, err := thermostatService.GetThermostat("thermostat-001")
		return err
	})

	// Set up graceful shutdown with enhanced error handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	serviceLogger.Info("Thermostat service is running", map[string]interface{}{
		"topics":      []string{"room-temp/+", "room-hum/+"},
		"thermostats": 1,
	})

	// Health check routine
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				healthResults := healthChecker.CheckHealth(ctx)
				hasErrors := false
				for checkName, err := range healthResults {
					if err != nil {
						hasErrors = true
						serviceLogger.Error("Health check failed", err, map[string]interface{}{
							"check": checkName,
						})
					}
				}
				if !hasErrors {
					serviceLogger.Debug("All health checks passed")
				}
			}
		}
	}()

	// Wait for shutdown signal
	<-sigChan
	serviceLogger.Info("Received shutdown signal, shutting down gracefully")

	// Cancel context to stop all background operations
	cancel()

	// Give services time to clean up
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	// Perform health check before shutdown
	finalHealthResults := healthChecker.CheckHealth(shutdownCtx)
	for checkName, err := range finalHealthResults {
		if err != nil {
			serviceLogger.Warn("Service unhealthy during shutdown", map[string]interface{}{
				"check": checkName,
				"error": err.Error(),
			})
		}
	}

	serviceLogger.Info("Thermostat service shutdown complete")
}

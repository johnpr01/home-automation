package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/johnpr01/home-automation/internal/logger"
	"github.com/johnpr01/home-automation/internal/services"
	"github.com/johnpr01/home-automation/pkg/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	log.Println("🔌 Starting Tapo Metrics Scraper")

	// Get configuration from environment variables
	metricsPort := getEnvWithDefault("METRICS_PORT", "2112")
	tplinkUsername := os.Getenv("TPLINK_USERNAME")
	tplinkPassword := os.Getenv("TPLINK_PASSWORD")
	logLevel := getEnvWithDefault("LOG_LEVEL", "info")
	pollIntervalStr := getEnvWithDefault("POLL_INTERVAL", "30s")

	// Parse poll interval
	pollInterval, err := time.ParseDuration(pollIntervalStr)
	if err != nil {
		log.Fatalf("Invalid poll interval: %v", err)
	}

	// Validate required environment variables
	if tplinkPassword == "" {
		log.Fatal("TPLINK_PASSWORD environment variable is required")
	}

	// Initialize logger
	serviceLogger := logger.NewLogger("tapo-metrics", nil)
	serviceLogger.Info("Tapo Metrics Scraper starting", map[string]interface{}{
		"metrics_port":  metricsPort,
		"poll_interval": pollInterval,
		"log_level":     logLevel,
		"has_username":  tplinkUsername != "",
		"has_password":  tplinkPassword != "",
	})

	// Create Prometheus client
	prometheusClient := prometheus.NewClient("http://prometheus:9090")

	// Create Tapo service
	tapoService := services.NewTapoService(nil, prometheusClient, serviceLogger)

	// Configure devices from environment or config file
	err = configureDevices(tapoService, tplinkUsername, tplinkPassword, pollInterval, serviceLogger)
	if err != nil {
		serviceLogger.Error("Failed to configure devices", err)
		log.Fatalf("Device configuration failed: %v", err)
	}

	// Start Tapo service
	if err := tapoService.Start(); err != nil {
		serviceLogger.Error("Failed to start Tapo service", err)
		log.Fatalf("Service start failed: %v", err)
	}

	// Setup metrics HTTP server
	http.Handle("/metrics", promhttp.Handler())
	http.HandleFunc("/health", healthHandler)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprintf(w, `
<html>
<head><title>Tapo Metrics Scraper</title></head>
<body>
<h1>Tapo Smart Plug Metrics</h1>
<p><a href="/metrics">Metrics</a> | <a href="/health">Health</a></p>
<p>Scraping %d devices every %v</p>
</body>
</html>`, len(getConfiguredDevices()), pollInterval)
	})

	// Start HTTP server in a goroutine
	server := &http.Server{
		Addr:    ":" + metricsPort,
		Handler: http.DefaultServeMux,
	}

	go func() {
		serviceLogger.Info("Starting metrics server", map[string]interface{}{
			"port": metricsPort,
		})
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			serviceLogger.Error("Metrics server failed", err)
			log.Fatalf("Server failed: %v", err)
		}
	}()

	// Setup graceful shutdown

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	serviceLogger.Info("Tapo metrics scraper started successfully", map[string]interface{}{
		"metrics_endpoint": fmt.Sprintf("http://localhost:%s/metrics", metricsPort),
		"health_endpoint":  fmt.Sprintf("http://localhost:%s/health", metricsPort),
	})

	// Wait for shutdown signal
	<-sigChan
	serviceLogger.Info("Shutdown signal received, stopping gracefully...")

	// Stop Tapo service
	if err := tapoService.Stop(); err != nil {
		serviceLogger.Error("Error stopping Tapo service", err)
	}

	// Shutdown HTTP server
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		serviceLogger.Error("Error shutting down HTTP server", err)
	}

	serviceLogger.Info("Tapo metrics scraper stopped")
}

func configureDevices(tapoService *services.TapoService, username, password string, pollInterval time.Duration, logger *logger.Logger) error {
	// Default configuration - can be overridden by config file or environment variables
	defaultDevices := []*services.TapoConfig{
		{
			DeviceID:     "tapo_device_1",
			DeviceName:   "washer",
			RoomID:       "laundry_room",
			IPAddress:    getEnvWithDefault("TAPO_DEVICE_1_IP", "192.168.1.53"),
			Username:     getEnvWithDefault("TPLINK_USERNAME", username),
			Password:     password,
			PollInterval: pollInterval,
			UseKlap:      getBoolEnvWithDefault("TAPO_DEVICE_1_USE_KLAP", false),
		},
		{
			DeviceID:     "tapo_device_2",
			DeviceName:   "dryer",
			RoomID:       "laundry_room",
			IPAddress:    getEnvWithDefault("TAPO_DEVICE_2_IP", "192.168.68.54"),
			Username:     getEnvWithDefault("TPLINK_USERNAME", username),
			Password:     password,
			PollInterval: pollInterval,
			UseKlap:      getBoolEnvWithDefault("TAPO_DEVICE_2_USE_KLAP", false),
		},
		{
			DeviceID:     "tapo_device_3",
			DeviceName:   "Hi-Fi",
			RoomID:       "living_room",
			IPAddress:    getEnvWithDefault("TAPO_DEVICE_3_IP", "192.168.68.60"),
			Username:     getEnvWithDefault("TPLINK_USERNAME", username),
			Password:     password,
			PollInterval: pollInterval,
			UseKlap:      getBoolEnvWithDefault("TAPO_DEVICE_3_USE_KLAP", false),
		},
		{
			DeviceID:     "tapo_device_4",
			DeviceName:   "boiler",
			RoomID:       "utility_room",
			IPAddress:    getEnvWithDefault("TAPO_DEVICE_4_IP", "192.168.68.63"),
			Username:     getEnvWithDefault("TPLINK_USERNAME", username),
			Password:     password,
			PollInterval: pollInterval,
			UseKlap:      getBoolEnvWithDefault("TAPO_DEVICE_4_USE_KLAP", false),
		},
	}

	// Add devices to service
	for _, deviceConfig := range defaultDevices {
		// Skip devices without valid IP addresses (using default values)
		if deviceConfig.IPAddress == "192.168.68.53" ||
			deviceConfig.IPAddress == "192.168.68.54" ||
			deviceConfig.IPAddress == "192.168.68.60" ||
			deviceConfig.IPAddress == "192.168.68.63" {
			logger.Info("Skipping default device IP - configure TAPO_DEVICE_X_IP environment variables", map[string]interface{}{
				"device_id": deviceConfig.DeviceID,
				"ip":        deviceConfig.IPAddress,
			})
			continue
		}

		if err := tapoService.AddDevice(deviceConfig); err != nil {
			logger.Error("Failed to add Tapo device", err, map[string]interface{}{
				"device_id": deviceConfig.DeviceID,
			})
			// Continue with other devices instead of failing completely
			continue
		}

		logger.Info("Added Tapo device", map[string]interface{}{
			"device_id":   deviceConfig.DeviceID,
			"device_name": deviceConfig.DeviceName,
			"ip_address":  deviceConfig.IPAddress,
			"use_klap":    deviceConfig.UseKlap,
		})
	}

	return nil
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"status":"healthy","timestamp":"%s","service":"tapo-metrics"}`, time.Now().Format(time.RFC3339))
}

func getConfiguredDevices() []string {
	// This would return actual configured devices in a real implementation
	devices := []string{}
	if os.Getenv("TAPO_DEVICE_1_IP") != "" && os.Getenv("TAPO_DEVICE_1_IP") != "192.168.1.100" {
		devices = append(devices, "device_1")
	}
	if os.Getenv("TAPO_DEVICE_2_IP") != "" && os.Getenv("TAPO_DEVICE_2_IP") != "192.168.1.101" {
		devices = append(devices, "device_2")
	}
	return devices
}

func getEnvWithDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getBoolEnvWithDefault(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.ParseBool(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}

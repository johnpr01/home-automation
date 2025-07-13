package config

import (
	"os"
)

type Config struct {
	Port     string
	Database string
	MQTT     MQTTConfig
}

type MQTTConfig struct {
	Broker   string
	Port     string
	Username string
	Password string
}

func Load() *Config {
	return &Config{
		Port:     getEnv("PORT", "8080"),
		Database: getEnv("DATABASE_URL", ""),
		MQTT: MQTTConfig{
			Broker:   getEnv("MQTT_BROKER", "localhost"),
			Port:     getEnv("MQTT_PORT", "1883"),
			Username: getEnv("MQTT_USERNAME", ""),
			Password: getEnv("MQTT_PASSWORD", ""),
		},
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

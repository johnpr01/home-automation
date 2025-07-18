package config

import (
	"os"
)

type Config struct {
	Port     string
	Database string
	MQTT     MQTTConfig
	Kafka    KafkaConfig
}

type MQTTConfig struct {
	Broker   string
	Port     string
	Username string
	Password string
}

type KafkaConfig struct {
	Brokers   []string
	LogTopic  string
	ClientID  string
	BatchSize int
	Timeout   string
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
		Kafka: KafkaConfig{
			Brokers:   []string{getEnv("KAFKA_BROKERS", "localhost:9092")},
			LogTopic:  getEnv("KAFKA_LOG_TOPIC", "home-automation-logs"),
			ClientID:  getEnv("KAFKA_CLIENT_ID", "home-automation-logger"),
			BatchSize: 100,
			Timeout:   "5s",
		},
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

package models

import "time"

type Sensor struct {
	ID          string      `json:"id"`
	Name        string      `json:"name"`
	Type        SensorType  `json:"type"`
	Value       interface{} `json:"value"`
	Unit        string      `json:"unit"`
	LastUpdated time.Time   `json:"last_updated"`
}

type SensorType string

const (
	SensorTypeTemperature SensorType = "temperature"
	SensorTypeHumidity    SensorType = "humidity"
	SensorTypeMotion      SensorType = "motion"
	SensorTypeLight       SensorType = "light"
	SensorTypeDoor        SensorType = "door"
	SensorTypeWindow      SensorType = "window"
	SensorTypeSmoke       SensorType = "smoke"
	SensorTypePressure    SensorType = "pressure"
)

type SensorReading struct {
	SensorID  string      `json:"sensor_id"`
	Value     interface{} `json:"value"`
	Timestamp time.Time   `json:"timestamp"`
}

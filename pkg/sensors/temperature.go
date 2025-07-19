package sensors

import (
	"math/rand"
	"time"

	"github.com/johnpr01/home-automation/internal/models"
)

type TemperatureSensor struct {
	*models.Sensor
	MinTemp float64 `json:"min_temp"`
	MaxTemp float64 `json:"max_temp"`
}

func NewTemperatureSensor(id, name string) *TemperatureSensor {
	return &TemperatureSensor{
		Sensor: &models.Sensor{
			ID:          id,
			Name:        name,
			Type:        models.SensorTypeTemperature,
			Value:       20.0,
			Unit:        "°C",
			LastUpdated: time.Now(),
		},
		MinTemp: -10.0,
		MaxTemp: 40.0,
	}
}

func (ts *TemperatureSensor) ReadValue() (float64, error) {
	// Simulate reading from actual sensor
	// In real implementation, this would read from actual hardware
	temperature := ts.simulateReading()

	ts.Sensor.Value = temperature
	ts.Sensor.LastUpdated = time.Now()

	return temperature, nil
}

func (ts *TemperatureSensor) simulateReading() float64 {
	// Simulate realistic temperature variations
	baseTemp := 22.0                          // Base room temperature
	variation := (rand.Float64() - 0.5) * 4.0 // ±2°C variation

	temperature := baseTemp + variation

	// Ensure within sensor limits
	if temperature < ts.MinTemp {
		temperature = ts.MinTemp
	} else if temperature > ts.MaxTemp {
		temperature = ts.MaxTemp
	}

	// Round to 1 decimal place
	return float64(int(temperature*10)) / 10.0
}

func (ts *TemperatureSensor) SetCalibration(offset float64) {
	// Apply calibration offset
	if currentValue, ok := ts.Value.(float64); ok {
		ts.Value = currentValue + offset
		ts.LastUpdated = time.Now()
	}
}

func (ts *TemperatureSensor) IsWithinRange(min, max float64) bool {
	if currentValue, ok := ts.Value.(float64); ok {
		return currentValue >= min && currentValue <= max
	}
	return false
}

package services

import (
	"context"
	"time"
)

// TimeSeriesClient defines the interface for time series databases
type TimeSeriesClient interface {
	Connect() error
	Disconnect()
	WriteEnergyReading(ctx context.Context, deviceID, roomID string, powerW, energyWh, voltageV, currentA float64, isOn bool, timestamp time.Time) error
	WriteTemperatureReading(ctx context.Context, deviceID, roomID string, tempF, humidity float64, timestamp time.Time) error
}

// EnergyReading represents energy data from smart plugs
type EnergyReading struct {
	DeviceID       string    `json:"device_id"`
	DeviceName     string    `json:"device_name"`
	RoomID         string    `json:"room_id"`
	PowerW         float64   `json:"power_w"`
	EnergyWh       float64   `json:"energy_wh"`
	VoltageV       float64   `json:"voltage_v"`
	CurrentA       float64   `json:"current_a"`
	IsOn           bool      `json:"is_on"`
	SignalStrength float64   `json:"signal_strength"`
	Temperature    float64   `json:"temperature"`
	Timestamp      time.Time `json:"timestamp"`
}

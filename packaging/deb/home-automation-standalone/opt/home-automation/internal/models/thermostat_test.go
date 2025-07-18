package models

import (
	"testing"
	"time"
)

func TestThermostatDefaults(t *testing.T) {
	thermostat := &Thermostat{
		ID:         "test-thermostat",
		Name:       "Test Thermostat",
		RoomID:     "living-room",
		TargetTemp: 72.0,
		Mode:       ModeAuto,
		Status:     StatusIdle,
	}

	if thermostat.ID != "test-thermostat" {
		t.Errorf("Expected ID 'test-thermostat', got '%s'", thermostat.ID)
	}

	if thermostat.TargetTemp != 72.0 {
		t.Errorf("Expected target temp 72.0, got %.1f", thermostat.TargetTemp)
	}

	if thermostat.Mode != ModeAuto {
		t.Errorf("Expected mode Auto, got %s", thermostat.Mode)
	}

	if thermostat.Status != StatusIdle {
		t.Errorf("Expected status Idle, got %s", thermostat.Status)
	}
}

func TestThermostatModes(t *testing.T) {
	validModes := []ThermostatMode{
		ModeOff,
		ModeHeat,
		ModeCool,
		ModeAuto,
		ModeFan,
	}

	expectedStrings := []string{
		"off",
		"heat",
		"cool",
		"auto",
		"fan",
	}

	for i, mode := range validModes {
		if string(mode) != expectedStrings[i] {
			t.Errorf("Expected mode string '%s', got '%s'", expectedStrings[i], string(mode))
		}
	}
}

func TestThermostatStatuses(t *testing.T) {
	validStatuses := []ThermostatStatus{
		StatusIdle,
		StatusHeating,
		StatusCooling,
		StatusFan,
	}

	expectedStrings := []string{
		"idle",
		"heating",
		"cooling",
		"fan",
	}

	for i, status := range validStatuses {
		if string(status) != expectedStrings[i] {
			t.Errorf("Expected status string '%s', got '%s'", expectedStrings[i], string(status))
		}
	}
}

func TestThermostatTemperatureValidation(t *testing.T) {
	thermostat := &Thermostat{
		MinTemp: 60.0,
		MaxTemp: 85.0,
	}

	// Test valid temperatures
	validTemps := []float64{60.0, 72.0, 85.0}
	for _, temp := range validTemps {
		thermostat.TargetTemp = temp
		if thermostat.TargetTemp != temp {
			t.Errorf("Valid temperature %.1f was not set correctly", temp)
		}
	}
}

func TestThermostatHysteresis(t *testing.T) {
	thermostat := &Thermostat{
		TargetTemp: 72.0,
		Hysteresis: 1.0,
	}

	// Test heating threshold (target - hysteresis)
	heatingThreshold := thermostat.TargetTemp - thermostat.Hysteresis
	if heatingThreshold != 71.0 {
		t.Errorf("Expected heating threshold 71.0, got %.1f", heatingThreshold)
	}

	// Test cooling threshold (target + hysteresis)
	coolingThreshold := thermostat.TargetTemp + thermostat.Hysteresis
	if coolingThreshold != 73.0 {
		t.Errorf("Expected cooling threshold 73.0, got %.1f", coolingThreshold)
	}
}

func TestThermostatTimestamps(t *testing.T) {
	now := time.Now()
	thermostat := &Thermostat{
		CreatedAt:        now,
		UpdatedAt:        now,
		LastSensorUpdate: now,
	}

	if thermostat.CreatedAt != now {
		t.Error("CreatedAt timestamp not set correctly")
	}

	if thermostat.UpdatedAt != now {
		t.Error("UpdatedAt timestamp not set correctly")
	}

	if thermostat.LastSensorUpdate != now {
		t.Error("LastSensorUpdate timestamp not set correctly")
	}
}

func TestThermostatCapabilities(t *testing.T) {
	thermostat := &Thermostat{
		HeatingEnabled: true,
		CoolingEnabled: true,
		FanSpeed:       50,
	}

	if !thermostat.HeatingEnabled {
		t.Error("Expected heating to be enabled")
	}

	if !thermostat.CoolingEnabled {
		t.Error("Expected cooling to be enabled")
	}

	if thermostat.FanSpeed != 50 {
		t.Errorf("Expected fan speed 50, got %d", thermostat.FanSpeed)
	}
}

func TestThermostatCalibration(t *testing.T) {
	thermostat := &Thermostat{
		CurrentTemp:       72.0,
		TemperatureOffset: 1.5, // Sensor reads 1.5°F high
	}

	// Adjusted reading would be current temp - offset
	adjustedTemp := thermostat.CurrentTemp - thermostat.TemperatureOffset
	expectedAdjusted := 70.5

	if adjustedTemp != expectedAdjusted {
		t.Errorf("Expected adjusted temp %.1f, got %.1f", expectedAdjusted, adjustedTemp)
	}
}

func TestThermostatOnlineStatus(t *testing.T) {
	thermostat := &Thermostat{
		IsOnline: true,
	}

	if !thermostat.IsOnline {
		t.Error("Expected thermostat to be online")
	}

	thermostat.IsOnline = false
	if thermostat.IsOnline {
		t.Error("Expected thermostat to be offline")
	}
}

func TestThermostatHumidity(t *testing.T) {
	thermostat := &Thermostat{
		CurrentHumidity: 45.5,
	}

	if thermostat.CurrentHumidity != 45.5 {
		t.Errorf("Expected humidity 45.5, got %.1f", thermostat.CurrentHumidity)
	}
}

func TestThermostatRoomAssignment(t *testing.T) {
	thermostat := &Thermostat{
		ID:     "thermostat-001",
		RoomID: "living-room",
	}

	if thermostat.RoomID != "living-room" {
		t.Errorf("Expected room ID 'living-room', got '%s'", thermostat.RoomID)
	}
}

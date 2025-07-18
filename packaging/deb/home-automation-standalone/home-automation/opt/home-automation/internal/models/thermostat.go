package models

import (
	"time"
)

// ThermostatMode represents the operating mode of the thermostat
type ThermostatMode string

const (
	ModeOff  ThermostatMode = "off"
	ModeHeat ThermostatMode = "heat"
	ModeCool ThermostatMode = "cool"
	ModeAuto ThermostatMode = "auto"
	ModeFan  ThermostatMode = "fan"
)

// ThermostatStatus represents the current status of the thermostat
type ThermostatStatus string

const (
	StatusIdle    ThermostatStatus = "idle"
	StatusHeating ThermostatStatus = "heating"
	StatusCooling ThermostatStatus = "cooling"
	StatusFan     ThermostatStatus = "fan"
)

// Thermostat represents a smart thermostat device
// All temperature values are stored and processed in Fahrenheit
type Thermostat struct {
	ID                string           `json:"id" db:"id"`
	Name              string           `json:"name" db:"name"`
	RoomID            string           `json:"room_id" db:"room_id"`
	CurrentTemp       float64          `json:"current_temp" db:"current_temp"`         // Temperature in Fahrenheit
	CurrentHumidity   float64          `json:"current_humidity" db:"current_humidity"` // Humidity percentage
	TargetTemp        float64          `json:"target_temp" db:"target_temp"`           // Target temperature in Fahrenheit
	Mode              ThermostatMode   `json:"mode" db:"mode"`
	Status            ThermostatStatus `json:"status" db:"status"`
	FanSpeed          int              `json:"fan_speed" db:"fan_speed"` // 0-100
	HeatingEnabled    bool             `json:"heating_enabled" db:"heating_enabled"`
	CoolingEnabled    bool             `json:"cooling_enabled" db:"cooling_enabled"`
	TemperatureOffset float64          `json:"temperature_offset" db:"temperature_offset"` // Calibration offset in Fahrenheit
	Hysteresis        float64          `json:"hysteresis" db:"hysteresis"`                 // Temperature dead band in Fahrenheit
	MinTemp           float64          `json:"min_temp" db:"min_temp"`                     // Minimum temperature in Fahrenheit
	MaxTemp           float64          `json:"max_temp" db:"max_temp"`                     // Maximum temperature in Fahrenheit
	LastSensorUpdate  time.Time        `json:"last_sensor_update" db:"last_sensor_update"`
	CreatedAt         time.Time        `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time        `json:"updated_at" db:"updated_at"`
	IsOnline          bool             `json:"is_online" db:"is_online"`
}

// ThermostatSchedule represents a scheduled temperature setting
type ThermostatSchedule struct {
	ID           string         `json:"id" db:"id"`
	ThermostatID string         `json:"thermostat_id" db:"thermostat_id"`
	Name         string         `json:"name" db:"name"`
	DayOfWeek    int            `json:"day_of_week" db:"day_of_week"` // 0=Sunday, 1=Monday, etc.
	StartTime    string         `json:"start_time" db:"start_time"`   // HH:MM format
	TargetTemp   float64        `json:"target_temp" db:"target_temp"`
	Mode         ThermostatMode `json:"mode" db:"mode"`
	Enabled      bool           `json:"enabled" db:"enabled"`
	CreatedAt    time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at" db:"updated_at"`
}

// ThermostatCommand represents a command to send to the thermostat
type ThermostatCommand struct {
	Type      string      `json:"type"`
	Value     interface{} `json:"value"`
	Timestamp time.Time   `json:"timestamp"`
}

// ThermostatCommands
const (
	CmdSetTargetTemp = "set_target_temp"
	CmdSetMode       = "set_mode"
	CmdSetFanSpeed   = "set_fan_speed"
	CmdTurnOn        = "turn_on"
	CmdTurnOff       = "turn_off"
	CmdGetStatus     = "get_status"
)

// IsValidMode checks if the thermostat mode is valid
func (t *Thermostat) IsValidMode(mode ThermostatMode) bool {
	switch mode {
	case ModeOff, ModeHeat, ModeCool, ModeAuto, ModeFan:
		return true
	default:
		return false
	}
}

// IsValidTargetTemp checks if the target temperature is within acceptable range
func (t *Thermostat) IsValidTargetTemp(temp float64) bool {
	return temp >= t.MinTemp && temp <= t.MaxTemp
}

// ShouldHeat determines if heating should be activated
func (t *Thermostat) ShouldHeat() bool {
	if !t.HeatingEnabled || t.Mode == ModeOff || t.Mode == ModeCool {
		return false
	}

	// Use hysteresis to prevent frequent on/off cycling
	return t.CurrentTemp < (t.TargetTemp - t.Hysteresis/2)
}

// ShouldCool determines if cooling should be activated
func (t *Thermostat) ShouldCool() bool {
	if !t.CoolingEnabled || t.Mode == ModeOff || t.Mode == ModeHeat {
		return false
	}

	// Use hysteresis to prevent frequent on/off cycling
	return t.CurrentTemp > (t.TargetTemp + t.Hysteresis/2)
}

// GetNextAction determines what action the thermostat should take
func (t *Thermostat) GetNextAction() ThermostatStatus {
	if t.Mode == ModeOff {
		return StatusIdle
	}

	switch t.Mode {
	case ModeHeat:
		if t.ShouldHeat() {
			return StatusHeating
		}
		return StatusIdle

	case ModeCool:
		if t.ShouldCool() {
			return StatusCooling
		}
		return StatusIdle

	case ModeAuto:
		if t.ShouldHeat() {
			return StatusHeating
		} else if t.ShouldCool() {
			return StatusCooling
		}
		return StatusIdle

	case ModeFan:
		return StatusFan

	default:
		return StatusIdle
	}
}

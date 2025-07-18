package utils

// Temperature conversion utilities

// CelsiusToFahrenheit converts Celsius to Fahrenheit
func CelsiusToFahrenheit(celsius float64) float64 {
	return (celsius * 9.0 / 5.0) + 32.0
}

// FahrenheitToCelsius converts Fahrenheit to Celsius
func FahrenheitToCelsius(fahrenheit float64) float64 {
	return (fahrenheit - 32.0) * 5.0 / 9.0
}

// Default temperature constants in Fahrenheit
const (
	DefaultHysteresis   = 1.0  // 1°F default hysteresis
	DefaultMinTemp      = 50.0 // 50°F minimum (10°C)
	DefaultMaxTemp      = 95.0 // 95°F maximum (35°C)
	DefaultTargetTemp   = 70.0 // 70°F default target (21°C)
	ComfortableRoomTemp = 72.0 // 72°F comfortable room temperature
	HeatingThreshold    = 68.0 // 68°F typical heating threshold
	CoolingThreshold    = 76.0 // 76°F typical cooling threshold
)

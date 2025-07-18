package utils

import (
	"math"
	"testing"
)

func TestTemperatureConversion(t *testing.T) {
	tests := []struct {
		celsius    float64
		fahrenheit float64
	}{
		{0, 32},    // Freezing point of water
		{100, 212}, // Boiling point of water
		{20, 68},   // Room temperature
		{21, 69.8}, // Slightly warm room temperature
		{22, 71.6}, // Comfortable room temperature
		{37, 98.6}, // Body temperature
		{-40, -40}, // Same in both scales
	}

	for _, test := range tests {
		// Test Celsius to Fahrenheit
		result := CelsiusToFahrenheit(test.celsius)
		if math.Abs(result-test.fahrenheit) > 0.1 {
			t.Errorf("CelsiusToFahrenheit(%.1f) = %.1f, want %.1f",
				test.celsius, result, test.fahrenheit)
		}

		// Test Fahrenheit to Celsius
		result = FahrenheitToCelsius(test.fahrenheit)
		if math.Abs(result-test.celsius) > 0.1 {
			t.Errorf("FahrenheitToCelsius(%.1f) = %.1f, want %.1f",
				test.fahrenheit, result, test.celsius)
		}
	}
}

func TestDefaultTemperatureConstants(t *testing.T) {
	// Verify that default constants are reasonable for US thermostats
	if DefaultMinTemp < 40 || DefaultMinTemp > 60 {
		t.Errorf("DefaultMinTemp = %.1f°F, should be between 40-60°F", DefaultMinTemp)
	}

	if DefaultMaxTemp < 85 || DefaultMaxTemp > 100 {
		t.Errorf("DefaultMaxTemp = %.1f°F, should be between 85-100°F", DefaultMaxTemp)
	}

	if DefaultTargetTemp < 65 || DefaultTargetTemp > 75 {
		t.Errorf("DefaultTargetTemp = %.1f°F, should be between 65-75°F", DefaultTargetTemp)
	}

	if DefaultHysteresis < 0.5 || DefaultHysteresis > 3 {
		t.Errorf("DefaultHysteresis = %.1f°F, should be between 0.5-3°F", DefaultHysteresis)
	}
}

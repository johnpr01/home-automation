package main

import (
	"testing"
	"time"

	"github.com/johnpr01/home-automation/internal/models"
	"github.com/johnpr01/home-automation/pkg/devices"
)

func TestLightCreation(t *testing.T) {
	light := devices.NewLight("test-light-1", "Test Light")

	if light.ID != "test-light-1" {
		t.Errorf("Expected ID 'test-light-1', got '%s'", light.ID)
	}

	if light.Name != "Test Light" {
		t.Errorf("Expected Name 'Test Light', got '%s'", light.Name)
	}

	if light.Type != models.DeviceTypeLight {
		t.Errorf("Expected Type 'light', got '%s'", light.Type)
	}

	if light.IsOn {
		t.Error("Expected light to be off initially")
	}
}

func TestLightOperations(t *testing.T) {
	light := devices.NewLight("test-light-2", "Test Light 2")

	// Test turning on
	err := light.TurnOn()
	if err != nil {
		t.Errorf("Error turning on light: %v", err)
	}

	if !light.IsOn {
		t.Error("Expected light to be on")
	}

	if light.Status != "on" {
		t.Errorf("Expected status 'on', got '%s'", light.Status)
	}

	// Test turning off
	err = light.TurnOff()
	if err != nil {
		t.Errorf("Error turning off light: %v", err)
	}

	if light.IsOn {
		t.Error("Expected light to be off")
	}

	if light.Status != "off" {
		t.Errorf("Expected status 'off', got '%s'", light.Status)
	}
}

func TestLightBrightness(t *testing.T) {
	light := devices.NewLight("test-light-3", "Test Light 3")

	// Test valid brightness
	err := light.SetBrightness(50)
	if err != nil {
		t.Errorf("Error setting brightness: %v", err)
	}

	if light.Brightness != 50 {
		t.Errorf("Expected brightness 50, got %d", light.Brightness)
	}

	// Test invalid brightness
	err = light.SetBrightness(-10)
	if err == nil {
		t.Error("Expected error for negative brightness")
	}

	err = light.SetBrightness(150)
	if err == nil {
		t.Error("Expected error for brightness > 100")
	}
}

func TestLightColor(t *testing.T) {
	light := devices.NewLight("test-light-4", "Test Light 4")

	// Test valid color
	err := light.SetColor(255, 128, 64)
	if err != nil {
		t.Errorf("Error setting color: %v", err)
	}

	if light.Color.Red != 255 || light.Color.Green != 128 || light.Color.Blue != 64 {
		t.Errorf("Expected color RGB(255,128,64), got RGB(%d,%d,%d)",
			light.Color.Red, light.Color.Green, light.Color.Blue)
	}

	// Test invalid color
	err = light.SetColor(-1, 128, 64)
	if err == nil {
		t.Error("Expected error for negative red value")
	}

	err = light.SetColor(255, 300, 64)
	if err == nil {
		t.Error("Expected error for green value > 255")
	}
}

func TestLightLastUpdated(t *testing.T) {
	light := devices.NewLight("test-light-5", "Test Light 5")
	initialTime := light.LastUpdated

	// Wait a bit to ensure time difference
	time.Sleep(10 * time.Millisecond)

	err := light.TurnOn()
	if err != nil {
		t.Errorf("Error turning on light: %v", err)
	}

	if !light.LastUpdated.After(initialTime) {
		t.Error("Expected LastUpdated to be updated after turning on")
	}
}

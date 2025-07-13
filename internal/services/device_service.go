package services

import (
	"fmt"
	"sync"

	"github.com/johnpr01/home-automation/internal/models"
)

type DeviceService struct {
	devices map[string]*models.Device
	mutex   sync.RWMutex
}

func NewDeviceService() *DeviceService {
	return &DeviceService{
		devices: make(map[string]*models.Device),
	}
}

func (s *DeviceService) GetDevice(id string) (*models.Device, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	
	device, exists := s.devices[id]
	if !exists {
		return nil, fmt.Errorf("device with id %s not found", id)
	}
	
	return device, nil
}

func (s *DeviceService) GetAllDevices() []*models.Device {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	
	devices := make([]*models.Device, 0, len(s.devices))
	for _, device := range s.devices {
		devices = append(devices, device)
	}
	
	return devices
}

func (s *DeviceService) AddDevice(device *models.Device) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	
	s.devices[device.ID] = device
	return nil
}

func (s *DeviceService) UpdateDevice(id string, updates map[string]interface{}) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	
	device, exists := s.devices[id]
	if !exists {
		return fmt.Errorf("device with id %s not found", id)
	}
	
	// Update device properties
	for key, value := range updates {
		device.Properties[key] = value
	}
	
	return nil
}

func (s *DeviceService) ExecuteCommand(cmd *models.DeviceCommand) error {
	device, err := s.GetDevice(cmd.DeviceID)
	if err != nil {
		return err
	}
	
	// Execute command based on device type and action
	switch device.Type {
	case models.DeviceTypeLight:
		return s.executeLightCommand(device, cmd)
	case models.DeviceTypeSwitch:
		return s.executeSwitchCommand(device, cmd)
	case models.DeviceTypeClimate:
		return s.executeClimateCommand(device, cmd)
	default:
		return fmt.Errorf("unsupported device type: %s", device.Type)
	}
}

func (s *DeviceService) executeLight// Internal command execution methods
func (s *DeviceService) executeLightCommand(device *models.Device, cmd *models.DeviceCommand) error {
	// Implement light-specific commands (on, off, dim, etc.)
	return nil
}

func (s *DeviceService) executeSwitchCommand(device *models.Device, cmd *models.DeviceCommand) error {
	// Implement switch-specific commands (on, off)
	return nil
}

func (s *DeviceService) executeClimateCommand(device *models.Device, cmd *models.DeviceCommand) error {
	// Implement climate-specific commands (temperature, mode)
	return nil
}

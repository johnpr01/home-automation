package services

import (
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/johnpr01/home-automation/internal/models"
	"github.com/johnpr01/home-automation/pkg/kafka"
	"github.com/johnpr01/home-automation/pkg/mqtt"
)

type DeviceService struct {
	devices     map[string]*models.Device
	mutex       sync.RWMutex
	mqttClient  *mqtt.Client
	kafkaClient *kafka.Client
	logger      *log.Logger
}

func NewDeviceService(mqttClient *mqtt.Client, kafkaClient *kafka.Client) *DeviceService {
	// Create or open log file
	logFile, err := os.OpenFile("logs/device_service.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		// Fallback to stdout if log file can't be created
		logFile = os.Stdout
	}

	logger := log.New(logFile, "[DeviceService] ", log.LstdFlags|log.Lshortfile)

	return &DeviceService{
		devices:     make(map[string]*models.Device),
		mqttClient:  mqttClient,
		kafkaClient: kafkaClient,
		logger:      logger,
	}
}

// logWithKafka logs to both file and Kafka
func (s *DeviceService) logWithKafka(level, message string, deviceID, action string, metadata map[string]interface{}) {
	// Log to file
	s.logger.Printf(message)

	// Send to Kafka if client is available
	if s.kafkaClient != nil {
		err := s.kafkaClient.PublishLog(level, "DeviceService", message, deviceID, action, metadata)
		if err != nil {
			// Log Kafka publishing errors to file only (avoid infinite recursion)
			s.logger.Printf("Failed to publish log to Kafka: %v", err)
		}
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

	message := fmt.Sprintf("Device added: %s (%s)", device.Name, device.ID)
	metadata := map[string]interface{}{
		"device_name": device.Name,
		"device_type": string(device.Type),
	}
	s.logWithKafka("INFO", message, device.ID, "add_device", metadata)

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
		message := fmt.Sprintf("Failed to execute command: device %s not found", cmd.DeviceID)
		s.logWithKafka("ERROR", message, cmd.DeviceID, cmd.Action, nil)
		return err
	}

	message := fmt.Sprintf("Executing command '%s' on device %s", cmd.Action, cmd.DeviceID)
	metadata := map[string]interface{}{
		"device_type":   string(device.Type),
		"command_value": cmd.Value,
	}
	s.logWithKafka("INFO", message, cmd.DeviceID, cmd.Action, metadata)

	// Execute command based on device type and action
	switch device.Type {
	case models.DeviceTypeLight:
		return s.executeLightCommand(device, cmd)
	case models.DeviceTypeSwitch:
		return s.executeSwitchCommand(device, cmd)
	case models.DeviceTypeClimate:
		return s.executeClimateCommand(device, cmd)
	default:
		message := fmt.Sprintf("Unsupported device type: %s for device %s", device.Type, device.ID)
		s.logWithKafka("ERROR", message, device.ID, cmd.Action, metadata)
		return fmt.Errorf("unsupported device type: %s", device.Type)
	}
}

// Internal command execution methods
func (s *DeviceService) executeLightCommand(device *models.Device, cmd *models.DeviceCommand) error {
	// Implement light-specific commands (on, off, dim, etc.)
	switch cmd.Action {
	case "turn_on":
		device.Status = "on"
		device.Properties["power"] = true
		message := fmt.Sprintf("Light %s turned on", device.ID)
		metadata := map[string]interface{}{"status": "on", "power": true}
		s.logWithKafka("INFO", message, device.ID, cmd.Action, metadata)
	case "turn_off":
		device.Status = "off"
		device.Properties["power"] = false
		message := fmt.Sprintf("Light %s turned off", device.ID)
		metadata := map[string]interface{}{"status": "off", "power": false}
		s.logWithKafka("INFO", message, device.ID, cmd.Action, metadata)
	case "set_brightness":
		if value, ok := cmd.Value.(float64); ok {
			device.Properties["brightness"] = value
			message := fmt.Sprintf("Light %s brightness set to %.0f", device.ID, value)
			metadata := map[string]interface{}{"brightness": value}
			s.logWithKafka("INFO", message, device.ID, cmd.Action, metadata)
		}
	default:
		message := fmt.Sprintf("Unknown light command: %s for device %s", cmd.Action, device.ID)
		s.logWithKafka("WARN", message, device.ID, cmd.Action, nil)
	}
	return nil
}

func (s *DeviceService) executeSwitchCommand(device *models.Device, cmd *models.DeviceCommand) error {
	// Implement switch-specific commands (on, off)
	switch cmd.Action {
	case "turn_on":
		device.Status = "on"
		device.Properties["power"] = true
		message := fmt.Sprintf("Switch %s turned on", device.ID)
		metadata := map[string]interface{}{"status": "on", "power": true}
		s.logWithKafka("INFO", message, device.ID, cmd.Action, metadata)
	case "turn_off":
		device.Status = "off"
		device.Properties["power"] = false
		message := fmt.Sprintf("Switch %s turned off", device.ID)
		metadata := map[string]interface{}{"status": "off", "power": false}
		s.logWithKafka("INFO", message, device.ID, cmd.Action, metadata)
	default:
		message := fmt.Sprintf("Unknown switch command: %s for device %s", cmd.Action, device.ID)
		s.logWithKafka("WARN", message, device.ID, cmd.Action, nil)
	}
	return nil
}

func (s *DeviceService) executeClimateCommand(device *models.Device, cmd *models.DeviceCommand) error {
	// Implement climate-specific commands (temperature, mode)
	if cmd.Action == "set_temperature" {
		if value, ok := cmd.Value.(float64); ok {
			// Set temperature logic here
			device.Properties["temperature"] = value
			message := fmt.Sprintf("Temperature set to %.2f for climate device %s", value, device.ID)
			metadata := map[string]interface{}{"temperature": value}
			s.logWithKafka("INFO", message, device.ID, cmd.Action, metadata)
		}
	} else if cmd.Action == "get_temperature" {
		// Get temperature logic here
		if temp, ok := device.Properties["temperature"].(float64); ok {
			message := fmt.Sprintf("Current temperature for device %s: %.2f", device.ID, temp)
			metadata := map[string]interface{}{"temperature": temp}
			s.logWithKafka("INFO", message, device.ID, cmd.Action, metadata)

			// Send temperature via MQTT
			if s.mqttClient != nil {
				tempStr := fmt.Sprintf("%.2f", temp)
				mqttMessage := &mqtt.Message{
					Topic:   "temp",
					Payload: []byte(tempStr),
					QoS:     1,
					Retain:  false,
				}
				err := s.mqttClient.Publish(mqttMessage)
				if err != nil {
					errorMsg := fmt.Sprintf("Failed to publish temperature to MQTT for device %s: %v", device.ID, err)
					errorMetadata := map[string]interface{}{"temperature": temp, "mqtt_error": err.Error()}
					s.logWithKafka("ERROR", errorMsg, device.ID, "mqtt_publish", errorMetadata)
				} else {
					successMsg := fmt.Sprintf("Successfully published temperature %.2f to MQTT topic 'temp' for device %s", temp, device.ID)
					successMetadata := map[string]interface{}{"temperature": temp, "mqtt_topic": "temp"}
					s.logWithKafka("INFO", successMsg, device.ID, "mqtt_publish", successMetadata)
				}
			}
		} else {
			return fmt.Errorf("temperature not set for device %s", device.ID)
		}
	}
	return nil
}

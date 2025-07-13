package models

import "time"

type Device struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Type        DeviceType             `json:"type"`
	Status      string                 `json:"status"`
	Properties  map[string]interface{} `json:"properties"`
	LastUpdated time.Time              `json:"last_updated"`
}

type DeviceType string

const (
	DeviceTypeLight   DeviceType = "light"
	DeviceTypeSwitch  DeviceType = "switch"
	DeviceTypeClimate DeviceType = "climate"
	DeviceTypeSensor  DeviceType = "sensor"
	DeviceTypeCamera  DeviceType = "camera"
	DeviceTypeLock    DeviceType = "lock"
)

type DeviceCommand struct {
	DeviceID string                 `json:"device_id"`
	Action   string                 `json:"action"`
	Value    interface{}            `json:"value"`
	Options  map[string]interface{} `json:"options,omitempty"`
}

package models

import "time"

// Room represents a physical room in the home automation system
type Room struct {
	ID                string                 `json:"id" db:"id"`
	Name              string                 `json:"name" db:"name"`
	Floor             string                 `json:"floor" db:"floor"`
	Type              RoomType               `json:"type" db:"type"`
	Area              float64                `json:"area" db:"area"`
	Height            float64                `json:"height" db:"height"`
	Properties        map[string]interface{} `json:"properties" db:"properties"`
	IsActive          bool                   `json:"is_active" db:"is_active"`
	CreatedAt         time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time              `json:"updated_at" db:"updated_at"`
	TargetTemperature float64                `json:"target_temperature" db:"target_temperature"`
	TemperatureRange  struct {
		Min float64 `json:"min" db:"temp_min"`
		Max float64 `json:"max" db:"temp_max"`
	} `json:"temperature_range" db:"temperature_range"`
	DeviceIDs []string `json:"device_ids" db:"device_ids"`
}

// RoomType represents different categories of rooms
type RoomType string

const (
	RoomTypeLivingRoom RoomType = "living_room"
	RoomTypeKitchen    RoomType = "kitchen"
	RoomTypeBedroom    RoomType = "bedroom"
	RoomTypeBathroom   RoomType = "bathroom"
	RoomTypeOffice     RoomType = "office"
	RoomTypeHallway    RoomType = "hallway"
	RoomTypeLaundry    RoomType = "laundry"
	RoomTypeGarage     RoomType = "garage"
	RoomTypeBasement   RoomType = "basement"
	RoomTypeAttic      RoomType = "attic"
	RoomTypeDining     RoomType = "dining_room"
	RoomTypeUtility    RoomType = "utility"
	RoomTypeOther      RoomType = "other"
)

package devices

import (
	"fmt"
	"time"

	"github.com/johnpr01/home-automation/internal/models"
)

type Light struct {
	*models.Device
	Brightness int  `json:"brightness"`
	Color      RGB  `json:"color"`
	IsOn       bool `json:"is_on"`
}

type RGB struct {
	Red   int `json:"red"`
	Green int `json:"green"`
	Blue  int `json:"blue"`
}

func NewLight(id, name string) *Light {
	return &Light{
		Device: &models.Device{
			ID:          id,
			Name:        name,
			Type:        models.DeviceTypeLight,
			Status:      "off",
			Properties:  make(map[string]interface{}),
			LastUpdated: time.Now(),
		},
		Brightness: 0,
		Color:      RGB{255, 255, 255},
		IsOn:       false,
	}
}

func (l *Light) TurnOn() error {
	l.IsOn = true
	l.Device.Status = "on"
	l.Device.LastUpdated = time.Now()
	return nil
}

func (l *Light) TurnOff() error {
	l.IsOn = false
	l.Device.Status = "off"
	l.Device.LastUpdated = time.Now()
	return nil
}

func (l *Light) SetBrightness(brightness int) error {
	if brightness < 0 || brightness > 100 {
		return fmt.Errorf("brightness must be between 0 and 100")
	}
	l.Brightness = brightness
	l.LastUpdated = time.Now()
	return nil
}

func (l *Light) SetColor(r, g, b int) error {
	if r < 0 || r > 255 || g < 0 || g > 255 || b < 0 || b > 255 {
		return fmt.Errorf("RGB values must be between 0 and 255")
	}
	l.Color = RGB{Red: r, Green: g, Blue: b}
	l.LastUpdated = time.Now()
	return nil
}

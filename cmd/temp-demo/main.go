package main

import (
	"fmt"
	"log"

	"github.com/johnpr01/home-automation/pkg/utils"
)

func main() {
	fmt.Println("🌡️  Temperature Conversion Demo")
	fmt.Println("================================")

	// Common temperatures
	temps := []struct {
		celsius float64
		name    string
	}{
		{10, "Cool day"},
		{20, "Comfortable room"},
		{21, "Default target (old)"},
		{22, "Warm room"},
		{25, "Warm day"},
		{30, "Hot day"},
		{35, "Very hot"},
	}

	fmt.Println("Celsius → Fahrenheit Conversion:")
	for _, temp := range temps {
		fahrenheit := utils.CelsiusToFahrenheit(temp.celsius)
		fmt.Printf("  %5.1f°C → %5.1f°F  (%s)\n",
			temp.celsius, fahrenheit, temp.name)
	}

	fmt.Println("\n🏠 Thermostat Default Values (Fahrenheit):")
	fmt.Printf("  Default Target Temperature: %.1f°F\n", utils.DefaultTargetTemp)
	fmt.Printf("  Default Hysteresis:         %.1f°F\n", utils.DefaultHysteresis)
	fmt.Printf("  Minimum Temperature:        %.1f°F\n", utils.DefaultMinTemp)
	fmt.Printf("  Maximum Temperature:        %.1f°F\n", utils.DefaultMaxTemp)

	fmt.Println("\n💡 Example Thermostat Operation:")
	target := utils.DefaultTargetTemp
	hysteresis := utils.DefaultHysteresis

	heatOn := target - hysteresis/2
	heatOff := target
	coolOn := target + hysteresis/2
	coolOff := target

	fmt.Printf("  Target Temperature: %.1f°F\n", target)
	fmt.Printf("  Heating turns ON:   %.1f°F\n", heatOn)
	fmt.Printf("  Heating turns OFF:  %.1f°F\n", heatOff)
	fmt.Printf("  Cooling turns ON:   %.1f°F\n", coolOn)
	fmt.Printf("  Cooling turns OFF:  %.1f°F\n", coolOff)

	log.Println("✅ Your thermostat system now uses Fahrenheit!")
}

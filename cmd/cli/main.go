package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	var (
		command = flag.String("cmd", "", "Command to execute (status, devices, sensors)")
		//device  = flag.String("device", "", "Device ID")
		//action  = flag.String("action", "", "Action to perform")
	)
	flag.Parse()

	switch *command {
	case "status":
		fmt.Println("Home automation system status: OK")
	case "devices":
		fmt.Println("Listing devices...")
	case "sensors":
		fmt.Println("Listing sensors...")
	default:
		fmt.Println("Usage: home-automation-cli -cmd [status|devices|sensors]")
		os.Exit(1)
	}
}

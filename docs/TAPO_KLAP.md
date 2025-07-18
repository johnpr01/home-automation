# TP-Link Tapo KLAP Protocol Implementation

This implementation provides support for TP-Link Tapo smart plugs using both the legacy protocol and the newer KLAP (Key-based Local Access Protocol) introduced in firmware version 1.1.0.

## Overview

As of firmware version 1.1.0 Build 230721 Rel.224802, TP-Link introduced the KLAP protocol for Tapo devices, making the previous secure pass-through protocol non-operational on newer firmware versions.

## Features

- **Dual Protocol Support**: Supports both legacy and KLAP protocols
- **Device Information**: Retrieve device details, status, and configuration
- **Energy Monitoring**: Get real-time power consumption and energy usage data
- **Prometheus Integration**: Export metrics to Prometheus for monitoring and alerting
- **Error Handling**: Robust error handling with retry mechanisms

## Protocol Support

### KLAP Protocol (Firmware 1.1.0+)
- Uses AES-GCM encryption for secure communication
- Implements handshake-based session establishment
- Supports all energy monitoring functions
- **Note**: Device control (on/off) not yet implemented for KLAP

### Legacy Protocol (Firmware < 1.1.0)
- Uses RSA + AES encryption
- Supports both monitoring and control functions
- Maintained for compatibility with older devices

## Usage

### Basic KLAP Client Usage

```go
package main

import (
    "context"
    "fmt"
    "time"

    "github.com/johnpr01/home-automation/internal/logger"
    "github.com/johnpr01/home-automation/pkg/tapo"
)

func main() {
    logger := logger.NewLogger("tapo-test", nil)
    
    // Create KLAP client for newer firmware
    client := tapo.NewKlapClient(
        "192.168.1.100",     // Device IP
        "your_username",     // TP-Link account username
        "your_password",     // TP-Link account password
        30*time.Second,      // Timeout
        *logger,
    )

    ctx := context.Background()
    
    // Connect to device
    if err := client.Connect(ctx); err != nil {
        panic(err)
    }

    // Get device information
    deviceInfo, err := client.GetDeviceInfo(ctx)
    if err != nil {
        panic(err)
    }
    
    fmt.Printf("Device: %s (Model: %s)\n", deviceInfo.DeviceID, deviceInfo.Model)
    fmt.Printf("Status: %t\n", deviceInfo.DeviceOn)

    // Get energy usage
    energyUsage, err := client.GetEnergyUsage(ctx)
    if err != nil {
        panic(err)
    }
    
    fmt.Printf("Current Power: %d mW\n", energyUsage.CurrentPower)
    fmt.Printf("Today Energy: %d Wh\n", energyUsage.TodayEnergy)
}
```

### Tapo Service with Mixed Protocol Support

```go
// Configure devices with appropriate protocol
configs := []*services.TapoConfig{
    {
        DeviceID:     "new_device",
        IPAddress:    "192.168.1.100",
        Username:     "username",
        Password:     "password",
        UseKlap:      true,  // Use KLAP for firmware 1.1.0+
        PollInterval: 30 * time.Second,
    },
    {
        DeviceID:     "old_device", 
        IPAddress:    "192.168.1.101",
        Username:     "username",
        Password:     "password",
        UseKlap:      false, // Use legacy protocol
        PollInterval: 30 * time.Second,
    },
}

// Add devices to service
for _, config := range configs {
    if err := tapoService.AddDevice(config); err != nil {
        log.Printf("Failed to add device %s: %v", config.DeviceID, err)
    }
}
```

## Determining Protocol Version

To determine which protocol to use:

1. **Check Firmware Version**: If firmware is 1.1.0 or later, use KLAP
2. **Try KLAP First**: Attempt KLAP connection, fallback to legacy if it fails
3. **Manual Configuration**: Explicitly set `UseKlap` based on known device firmware

## Metrics

The following Prometheus metrics are exported:

- `tapo_device_power_watts`: Current power consumption in watts
- `tapo_device_energy_wh`: Total energy consumption in watt-hours
- `tapo_device_status`: Device on/off status (1 = on, 0 = off)
- `tapo_device_signal_strength`: WiFi signal strength

Each metric includes labels for:
- `device_id`: Unique device identifier
- `device_name`: Human-readable device name
- `room_id`: Room/location identifier

## Error Handling

The implementation includes comprehensive error handling:

- **Connection Errors**: Automatic reconnection attempts
- **Authentication Errors**: Clear error messages for invalid credentials
- **Protocol Errors**: Graceful handling of unsupported operations
- **Network Errors**: Timeout and retry mechanisms

## Limitations

### KLAP Protocol
- Device control (on/off switching) not yet implemented
- Some advanced configuration options may not be available

### Legacy Protocol
- Not supported on firmware 1.1.0+ devices
- May have compatibility issues with newer device models

## Testing

Run the test applications:

```bash
# Test KLAP client directly
go run ./cmd/test-klap

# Test full service with both protocols
go run ./cmd/tapo-demo
```

## Configuration

Environment variables for demo applications:
- `TPLINK_PASSWORD`: TP-Link account password (for security)

## References

- [TP-Link Tapo Device Documentation](https://www.tp-link.com/support/download/)
- [KLAP Protocol Analysis](https://github.com/mirorucka/tapo-smartplug) (Java implementation)
- [Prometheus Metrics](https://prometheus.io/docs/concepts/metric_types/)

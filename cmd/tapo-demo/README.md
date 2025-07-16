# Tapo Demo Application

This application demonstrates how to monitor TP-Link Tapo smart plugs for energy consumption and store the data in InfluxDB.

## Prerequisites

- TP-Link Tapo smart plugs on your network
- InfluxDB running (optional, but recommended)
- MQTT broker running (Mosquitto)
- Go 1.21 or later

## Configuration

### Environment Variables

The application requires the following environment variable:

- `TPLINK_PASSWORD`: Your TP-Link account password

### Setting Environment Variables

**Linux/macOS:**
```bash
export TPLINK_PASSWORD="your_tapo_account_password"
```

**Windows (PowerShell):**
```powershell
$env:TPLINK_PASSWORD="your_tapo_account_password"
```

**Windows (Command Prompt):**
```cmd
set TPLINK_PASSWORD=your_tapo_account_password
```

## Running the Application

### Build and Run
```bash
go build
TPLINK_PASSWORD="your_password" ./tapo-demo
```

### Run Directly
```bash
TPLINK_PASSWORD="your_password" go run main.go
```

## Configuration

Before running, update the device configuration in `main.go`:

1. Replace IP addresses with your actual Tapo device IPs
2. Replace usernames with your actual TP-Link account username
3. The password will be automatically loaded from the `TPLINK_PASSWORD` environment variable

```go
exampleDevices := []*services.TapoConfig{
    {
        DeviceID:     "dryer",
        DeviceName:   "dryer",
        RoomID:       "laundry_room",
        IPAddress:    "192.168.68.54",      // Your actual device IP
        Username:     "johnpr01@gmail.com", // Your actual username
        Password:     tplinkPassword,        // Loaded from environment
        PollInterval: 30 * time.Second,
    },
    // Add more devices...
}
```

## Features

- **Real-time Energy Monitoring**: Polls devices every 30-60 seconds
- **InfluxDB Storage**: Stores time-series energy data
- **MQTT Publishing**: Publishes energy data to MQTT topics
- **Graceful Shutdown**: Handles SIGINT/SIGTERM signals
- **Error Handling**: Robust error handling with retries
- **Status Monitoring**: Logs service status every 5 minutes

## MQTT Topics

Energy data is published to:
- `tapo/{device_id}/energy` - Energy metrics (power, voltage, current, etc.)

## InfluxDB Data

Data is stored in the `sensor-data` bucket with measurements:
- `tapo_energy` - Power consumption and energy metrics

## Troubleshooting

### "TPLINK_PASSWORD environment variable not set"
Make sure you've set the environment variable before running the application.

### Device Connection Errors
- Verify device IP addresses are correct
- Check that devices are on the same network
- Ensure your TP-Link account credentials are correct

### MQTT Connection Issues
- Verify MQTT broker is running on localhost:1883
- Check firewall settings
- Ensure MQTT broker allows anonymous connections

### InfluxDB Issues
- InfluxDB is optional - the application will continue without it
- Verify InfluxDB is running on localhost:8086
- Check bucket and organization names

## Security Notes

- Never commit passwords to Git
- Use environment variables for sensitive data
- Consider using GitHub Actions secrets for CI/CD
- Use strong, unique passwords for your TP-Link account

## GitHub Actions Integration

This application supports GitHub Actions with secrets. Set up the `TPLINK_PASSWORD` secret in your repository settings, and the CI/CD pipeline will automatically use it for testing.

For more information, see [docs/GITHUB_SECRETS.md](../../docs/GITHUB_SECRETS.md).
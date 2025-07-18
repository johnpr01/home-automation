# test-klap - Tapo KLAP Protocol Test Utility

A command-line utility to test KLAP protocol connectivity with TP-Link Tapo smart plugs.

## Usage

```bash
./test-klap -host <IP_ADDRESS> -username <EMAIL> -password <PASSWORD>
```

## Options

- `-host`: IP address of the Tapo device (required)
- `-username`: TP-Link cloud account username/email (required)  
- `-password`: TP-Link cloud account password (required)
- `-timeout`: Connection timeout duration (default: 30s)
- `-help`: Show help message

## Examples

### Basic Test
```bash
./test-klap -host 192.168.1.100 -username your@email.com -password yourpassword
```

### With Custom Timeout
```bash
./test-klap -host 192.168.1.100 -username your@email.com -password yourpassword -timeout 60s
```

### Show Help
```bash
./test-klap -help
```

## Building

```bash
go build -o test-klap ./cmd/test-klap
```

## Output

The utility will:
1. Test KLAP protocol connection
2. Display device information (ID, model, firmware, status)
3. Show current energy usage (power, energy consumption, runtime)
4. Confirm successful connection

## Protocol Support

This utility specifically tests the **KLAP protocol**, which is used by newer Tapo devices (firmware 1.1.0+). For older devices, use the legacy protocol test utility instead.

## Troubleshooting

- **Connection failed**: Verify IP address is correct and device is reachable
- **Authentication failed**: Check username/password credentials
- **Timeout**: Try increasing timeout value or check network connectivity
- **Protocol errors**: Ensure device supports KLAP protocol (newer firmware)

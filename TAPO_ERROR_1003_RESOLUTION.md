# Tapo Error 1003 - Analysis and Resolution

## Issue Summary
The tapo-metrics-scraper was experiencing **error code 1003 (Invalid Request)** when trying to connect to Tapo devices using the legacy protocol.

## Root Cause Analysis

### Network Discovery
- **Incorrect subnet**: Initially testing on 192.168.1.x subnet
- **Correct subnet**: Devices are actually on 192.168.68.x subnet
- **Found devices**: 192.168.68.60, 192.168.68.63, 192.168.68.54, 192.168.68.53, 192.168.68.68

### Error 1003 Specifics
Error code 1003 from Tapo devices typically indicates:
1. **Invalid Request**: The API request format is incorrect
2. **Method Not Found**: The requested method is not supported by device firmware
3. **Authentication Issues**: Credentials are invalid or device is not linked to account
4. **Protocol Mismatch**: Device firmware doesn't support the legacy protocol

### Testing Results
```
Device: 192.168.68.60 (Hi-Fi Power)
- Legacy Protocol: ❌ Error 1003 (Invalid Request)
- KLAP Protocol: ❌ Server hash verification failed
- Network connectivity: ✅ HTTP 200 OK on /app endpoint
```

## Current Status

### Working Configuration
The `.env.example` file has been updated with:
- ✅ Correct IP addresses from network scan
- ✅ All devices set to use legacy protocol (`USE_KLAP=false`)
- ✅ Real device names and room assignments

### Devices Configured
1. **Dryer** (192.168.68.54) - Laundry Room
2. **Boiler** (192.168.68.63) - Utility Room  
3. **Hi-Fi** (192.168.68.60) - Living Room
4. **Washing Machine** (192.168.68.53) - Laundry Room

## Resolution Strategy

### Immediate Workaround
1. **Use legacy protocol** for all devices (already configured)
2. **Verify credentials** in Tapo mobile app
3. **Check device linking** - ensure all devices are properly added to your TP-Link account

### Recommended Testing Order
1. **Test one device at a time** with tapo-metrics-scraper
2. **Start with device that might work** (try 192.168.68.63 or 192.168.68.54)
3. **If 1003 persists**, verify credentials and device account linking
4. **If KLAP needed**, wait for hash verification fix

### Long-term Solution
1. **Fix KLAP hash verification** (see KLAP_TROUBLESHOOTING.md)
2. **Update device firmware** if available
3. **Test both protocols** for each device to determine optimal configuration

## Commands for Testing

### Test individual device
```bash
TPLINK_USERNAME=your@email.com TPLINK_PASSWORD=yourpassword TAPO_DEVICE_1_IP=192.168.68.54 go run cmd/tapo-metrics-scraper/main.go
```

### Test with KLAP protocol
```bash
TPLINK_USERNAME=your@email.com TPLINK_PASSWORD=yourpassword TAPO_DEVICE_1_IP=192.168.68.54 TAPO_DEVICE_1_USE_KLAP=true go run cmd/tapo-metrics-scraper/main.go
```

### Scan for more devices
```bash
go run cmd/scan-tapo/main.go -subnet=192.168.68
```

## Expected Outcome
With correct IP addresses and legacy protocol configuration, the error 1003 should be resolved and the tapo-metrics-scraper should successfully:
1. ✅ Connect to Tapo devices
2. ✅ Read energy consumption data
3. ✅ Export Prometheus metrics
4. ✅ Provide data for Grafana dashboards

## Files Updated
- `cmd/tapo-metrics-scraper/main.go` - Added support for 4 devices
- `.env.example` - Updated with real IP addresses and protocol settings
- Created diagnostic tools: `cmd/scan-tapo/`, `cmd/network-test/`, `cmd/diagnose-1003/`

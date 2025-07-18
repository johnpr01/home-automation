# Mosquitto Configuration Files

This directory contains configuration files for the Mosquitto MQTT broker.

## Files

### mosquitto.conf
Main configuration file with settings optimized for Raspberry Pi 5:
- Standard MQTT listener on port 1883
- WebSocket listener on port 9001  
- Logging configuration
- Performance tuning for Pi 5
- Security settings (currently allows anonymous connections)

### passwd.example
Example password file template. To use authentication:

1. Copy to `passwd`: `cp passwd.example passwd`
2. Generate user passwords:
   ```bash
   # Install mosquitto-clients if not available
   sudo apt install mosquitto-clients
   
   # Create password file with admin user
   mosquitto_passwd -c passwd admin
   
   # Add more users
   mosquitto_passwd passwd home-automation-server
   mosquitto_passwd passwd webclient
   ```
3. Enable in mosquitto.conf:
   ```
   allow_anonymous false
   password_file /mosquitto/config/passwd
   ```

### acl.example  
Example Access Control List for topic-level permissions. To use:

1. Copy to `acl`: `cp acl.example acl`
2. Customize user permissions as needed
3. Enable in mosquitto.conf:
   ```
   acl_file /mosquitto/config/acl
   ```

## Security Levels

### Development (Current)
- Anonymous connections allowed
- No topic restrictions
- Full access for all clients

### Production (Recommended)
- Password authentication required
- ACL-based topic restrictions
- User-specific permissions
- SSL/TLS encryption (optional)

## Docker Volume Mounts

The docker-compose.yml maps these directories:
- `./mosquitto/mosquitto.conf` → `/mosquitto/config/mosquitto.conf`
- `mosquitto_data` → `/mosquitto/data` (named volume)
- `mosquitto_logs` → `/mosquitto/log` (named volume)

To add authentication files, update docker-compose.yml:
```yaml
volumes:
  - ./mosquitto/mosquitto.conf:/mosquitto/config/mosquitto.conf
  - ./mosquitto/passwd:/mosquitto/config/passwd
  - ./mosquitto/acl:/mosquitto/config/acl
  - mosquitto_data:/mosquitto/data
  - mosquitto_logs:/mosquitto/log
```

## Testing MQTT Connection

```bash
# Test basic connection
mosquitto_pub -h localhost -p 1883 -t test -m "hello world"

# Subscribe to sensor topics
mosquitto_sub -h localhost -p 1883 -t "room-temp/+"

# Test WebSocket (if using web client)
# Connect to ws://your-pi-ip:9001
```

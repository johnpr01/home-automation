# 🔧 KLAP Protocol Hash Verification Troubleshooting Guide

## 🚨 Problem: "server hash verification failed"

This error occurs during the KLAP handshake when the device's computed hash doesn't match our expected hash.

## 🔍 Root Causes & Solutions

### 1. **Firmware Compatibility** (Most Common)
**Problem**: Device firmware < 1.1.0 doesn't support KLAP
**Solution**: Check firmware version first

```bash
# Test basic device connectivity
curl -v http://192.168.68.60/app/handshake1

# If device responds but KLAP fails, try legacy protocol
```

### 2. **Credential Format Issues**
**Problem**: KLAP expects exact credential format
**Solutions to try**:

#### A. Use exact TP-Link cloud credentials
- Must be the **exact** email and password used in TP-Link Kasa app
- Case-sensitive
- No extra spaces

#### B. Try device-specific credentials
Some devices require local device credentials instead of cloud credentials:
- Username: `admin` 
- Password: Device-specific password (check device label)

#### C. Try URL-encoded credentials
Some special characters might need encoding.

### 3. **Account Linking Issues**
**Problem**: Device not properly linked to TP-Link account
**Solution**:
1. Open TP-Link Kasa app
2. Verify device shows up and is controllable
3. Re-link device if necessary

### 4. **Network/Timing Issues**
**Problem**: Network latency causing hash mismatch
**Solutions**:
- Increase timeout: `-timeout 60s`
- Ensure stable network connection
- Try from same network segment

## 🛠️ Debugging Steps

### Step 1: Check Device Firmware
```bash
# Try to get device info via web interface
curl -s http://192.168.68.60/app | grep -i version
```

### Step 2: Test with Debug Utility
```bash
./build/debug-klap -host 192.168.68.60 -username johnpr01@gmail.com -password 8VJZ4S8UfyLyyh -debug
```

### Step 3: Try Alternative Credentials
```bash
# Try with 'admin' username (device-local)
./build/debug-klap -host 192.168.68.60 -username admin -password 8VJZ4S8UfyLyyh -debug

# Try different password formats
```

### Step 4: Check Device Logs
If available, check device logs for authentication attempts.

## 🔄 Alternative Approaches

### Option 1: Use Legacy Protocol
If KLAP continues to fail, use legacy protocol:
- Set `TAPO_DEVICE_X_USE_KLAP=false` in environment
- The system will automatically fall back to legacy protocol

### Option 2: Update Device Firmware
1. Open TP-Link Kasa app
2. Go to device settings
3. Check for firmware updates
4. Update to latest version (should support KLAP)

### Option 3: Factory Reset Device
As a last resort:
1. Factory reset the device
2. Re-setup with TP-Link Kasa app
3. Ensure firmware is latest
4. Try KLAP again

## 📊 Expected Hash Values
For `johnpr01@gmail.com` + `8VJZ4S8UfyLyyh`:
- Username SHA1: `9b37b105c3c5ce77c461fc23551f4bb2fc01347e`
- Password SHA1: `32c7c60f54b9c64940d5010e9935fe617ca4f0a3`
- Auth Hash: `819db1dd33301a8d90dcbacb06c9411b3a4637bf8fa20fdda4fc22692a0e837f`

If these don't match in debug output, there's a credential issue.

## ✅ Quick Fix Checklist

- [ ] Device responds to HTTP requests
- [ ] Credentials work in TP-Link Kasa app  
- [ ] Device firmware ≥ 1.1.0
- [ ] Using exact email/password from app
- [ ] No special characters causing encoding issues
- [ ] Stable network connection
- [ ] Device properly linked to account

## 🎯 Immediate Recommendation

**For your current setup, try:**

1. **Set devices to legacy protocol** in environment:
   ```bash
   TAPO_DEVICE_1_USE_KLAP=false
   TAPO_DEVICE_2_USE_KLAP=false
   TAPO_DEVICE_3_USE_KLAP=false
   TAPO_DEVICE_4_USE_KLAP=false
   ```

2. **Test one device with legacy first** to confirm credentials work

3. **Check firmware versions** and update if needed

4. **Re-enable KLAP** after firmware update

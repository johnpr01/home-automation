"""
SHT-30 Temperature and Humidity Sensor Driver for MicroPython
Compatible with Raspberry Pi Pico WH
"""

import time
from machine import I2C

class SHT30:
    """
    Driver for SHT-30 temperature and humidity sensor using I2C
    """
    
    # SHT-30 I2C address
    ADDRESS = 0x44
    
    # SHT-30 commands
    CMD_MEASURE_HIGH = b'\x24\x00'    # High repeatability measurement
    CMD_MEASURE_MED = b'\x24\x0B'     # Medium repeatability measurement
    CMD_MEASURE_LOW = b'\x24\x16'     # Low repeatability measurement
    CMD_SOFT_RESET = b'\x30\xA2'      # Soft reset
    CMD_STATUS = b'\xF3\x2D'          # Read status register
    
    def __init__(self, i2c, address=ADDRESS):
        """
        Initialize SHT-30 sensor
        
        Args:
            i2c: I2C interface object
            address: I2C address of the sensor (default: 0x44)
        """
        self.i2c = i2c
        self.address = address
        self._check_sensor()
    
    def _check_sensor(self):
        """
        Check if sensor is connected and responding
        """
        devices = self.i2c.scan()
        if self.address not in devices:
            raise RuntimeError(f"SHT-30 sensor not found at address 0x{self.address:02X}")
    
    def _crc8(self, data):
        """
        Calculate CRC8 checksum for data validation
        
        Args:
            data: bytes to calculate checksum for
            
        Returns:
            int: CRC8 checksum
        """
        crc = 0xFF
        for byte in data:
            crc ^= byte
            for _ in range(8):
                if crc & 0x80:
                    crc = (crc << 1) ^ 0x31
                else:
                    crc = crc << 1
        return crc & 0xFF
    
    def soft_reset(self):
        """
        Perform a soft reset of the sensor
        """
        try:
            self.i2c.writeto(self.address, self.CMD_SOFT_RESET)
            time.sleep_ms(100)  # Wait for reset to complete
        except Exception as e:
            raise RuntimeError(f"Failed to reset SHT-30: {e}")
    
    def read_status(self):
        """
        Read the status register
        
        Returns:
            int: Status register value
        """
        try:
            self.i2c.writeto(self.address, self.CMD_STATUS)
            time.sleep_ms(10)
            data = self.i2c.readfrom(self.address, 3)
            
            # Verify CRC
            if self._crc8(data[:2]) != data[2]:
                raise RuntimeError("CRC check failed for status register")
                
            return (data[0] << 8) | data[1]
        except Exception as e:
            raise RuntimeError(f"Failed to read status: {e}")
    
    def measure(self, repeatability='high'):
        """
        Perform a measurement
        
        Args:
            repeatability: Measurement repeatability ('high', 'medium', 'low')
            
        Returns:
            tuple: (temperature_celsius, humidity_percent)
        """
        # Select command based on repeatability
        if repeatability == 'high':
            cmd = self.CMD_MEASURE_HIGH
            delay_ms = 15
        elif repeatability == 'medium':
            cmd = self.CMD_MEASURE_MED
            delay_ms = 6
        elif repeatability == 'low':
            cmd = self.CMD_MEASURE_LOW
            delay_ms = 4
        else:
            raise ValueError("Repeatability must be 'high', 'medium', or 'low'")
        
        try:
            # Send measurement command
            self.i2c.writeto(self.address, cmd)
            
            # Wait for measurement to complete
            time.sleep_ms(delay_ms)
            
            # Read measurement data (6 bytes: temp_msb, temp_lsb, temp_crc, hum_msb, hum_lsb, hum_crc)
            data = self.i2c.readfrom(self.address, 6)
            
            # Verify CRC for temperature
            if self._crc8(data[:2]) != data[2]:
                raise RuntimeError("CRC check failed for temperature data")
            
            # Verify CRC for humidity
            if self._crc8(data[3:5]) != data[5]:
                raise RuntimeError("CRC check failed for humidity data")
            
            # Convert raw data to temperature and humidity
            temp_raw = (data[0] << 8) | data[1]
            hum_raw = (data[3] << 8) | data[4]
            
            # Convert to physical values
            temperature = -45 + (175 * temp_raw / 65535)
            humidity = 100 * hum_raw / 65535
            
            # Clamp humidity to valid range
            humidity = max(0, min(100, humidity))
            
            return temperature, humidity
            
        except Exception as e:
            raise RuntimeError(f"Failed to read measurement: {e}")
    
    def read_temperature_humidity(self):
        """
        Convenience method to read temperature and humidity with high repeatability
        
        Returns:
            tuple: (temperature_celsius, humidity_percent)
        """
        return self.measure('high')
    
    def read_temperature(self):
        """
        Read only temperature
        
        Returns:
            float: Temperature in Celsius
        """
        temp, _ = self.measure('high')
        return temp
    
    def read_humidity(self):
        """
        Read only humidity
        
        Returns:
            float: Humidity percentage
        """
        _, humidity = self.measure('high')
        return humidity

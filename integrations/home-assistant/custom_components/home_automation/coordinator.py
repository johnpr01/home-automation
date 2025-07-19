"""Data Update Coordinator for Home Automation."""
import asyncio
import logging
from datetime import timedelta
from typing import Any, Dict, List, Optional

import aiohttp
import async_timeout
from homeassistant.config_entries import ConfigEntry
from homeassistant.core import HomeAssistant
from homeassistant.helpers.update_coordinator import DataUpdateCoordinator, UpdateFailed
from homeassistant.helpers import aiohttp_client

from .const import DOMAIN, API_ENDPOINTS, DEFAULT_SCAN_INTERVAL

_LOGGER = logging.getLogger(__name__)

class HomeAutomationCoordinator(DataUpdateCoordinator):
    """Data update coordinator for Home Automation."""

    def __init__(self, hass: HomeAssistant, entry: ConfigEntry):
        """Initialize the coordinator."""
        self.hass = hass
        self.entry = entry
        self.host = entry.data["host"]
        self.port = entry.data["port"]
        self.base_url = f"http://{self.host}:{self.port}"
        
        self.session = aiohttp_client.async_get_clientsession(hass)
        
        super().__init__(
            hass,
            _LOGGER,
            name=DOMAIN,
            update_interval=timedelta(seconds=DEFAULT_SCAN_INTERVAL),
        )

    async def _async_update_data(self) -> Dict[str, Any]:
        """Update data via API."""
        try:
            async with async_timeout.timeout(10):
                data = {}
                
                # Get system status
                data["status"] = await self._api_request("status")
                
                # Get devices
                data["devices"] = await self._api_request("devices")
                
                # Get sensors
                data["sensors"] = await self._api_request("sensors")
                
                # Get rooms
                try:
                    data["rooms"] = await self._api_request("rooms")
                except:
                    data["rooms"] = []  # Handle case where rooms API might not be available
                
                return data
                
        except asyncio.TimeoutError as err:
            raise UpdateFailed(f"Timeout communicating with API") from err
        except (aiohttp.ClientError, aiohttp.ClientResponseError) as err:
            raise UpdateFailed(f"Error communicating with API: {err}") from err
        except Exception as err:
            raise UpdateFailed(f"Unknown error: {err}") from err

    async def _api_request(self, endpoint: str) -> Any:
        """Make an API request."""
        url = f"{self.base_url}{API_ENDPOINTS[endpoint]}"
        
        try:
            async with self.session.get(url) as response:
                response.raise_for_status()
                return await response.json()
        except aiohttp.ClientResponseError as err:
            _LOGGER.error(f"API request failed: {err}")
            raise
        except Exception as err:
            _LOGGER.error(f"Unexpected error during API request: {err}")
            raise

    async def async_get_device_data(self, device_id: str) -> Optional[Dict[str, Any]]:
        """Get specific device data."""
        if not self.data or "devices" not in self.data:
            return None
            
        for device in self.data["devices"]:
            if device.get("id") == device_id:
                return device
        return None

    async def async_get_sensor_data(self, sensor_id: str) -> Optional[Dict[str, Any]]:
        """Get specific sensor data."""
        if not self.data or "sensors" not in self.data:
            return None
            
        for sensor in self.data["sensors"]:
            if sensor.get("id") == sensor_id:
                return sensor
        return None

    async def async_get_room_data(self, room_id: str) -> Optional[Dict[str, Any]]:
        """Get specific room data."""
        if not self.data or "rooms" not in self.data:
            return None
            
        for room in self.data["rooms"]:
            if room.get("id") == room_id:
                return room
        return None

    async def async_send_device_command(self, device_id: str, command: Dict[str, Any]) -> bool:
        """Send a command to a device."""
        url = f"{self.base_url}/api/devices/{device_id}/command"
        
        try:
            async with async_timeout.timeout(10):
                async with self.session.post(url, json=command) as response:
                    if response.status == 200:
                        await self.async_request_refresh()
                        return True
                    else:
                        _LOGGER.error(f"Device command failed: {response.status}")
                        return False
        except Exception as err:
            _LOGGER.error(f"Error sending device command: {err}")
            return False

    async def async_set_switch_state(self, switch_id: str, state: bool) -> bool:
        """Set switch state via API."""
        try:
            url = f"{self.base_url}/api/switches/{switch_id}"
            data = {"state": state}
            
            async with aiohttp.ClientSession() as session:
                async with session.put(url, json=data, headers=self.headers) as response:
                    return response.status == 200
        except Exception as err:
            _LOGGER.error("Error setting switch state: %s", err)
            return False

    async def async_set_light_state(self, light_id: str, state: bool, kwargs: dict) -> bool:
        """Set light state via API."""
        try:
            url = f"{self.base_url}/api/lights/{light_id}"
            data = {"state": state}
            
            # Add optional light parameters
            if "brightness" in kwargs:
                # Convert from 0-255 to 0-100
                data["brightness"] = int(kwargs["brightness"] * 100 / 255)
            if "rgb_color" in kwargs:
                data["rgb_color"] = kwargs["rgb_color"]
            if "color_temp" in kwargs:
                data["color_temp"] = kwargs["color_temp"]
            
            async with aiohttp.ClientSession() as session:
                async with session.put(url, json=data, headers=self.headers) as response:
                    return response.status == 200
        except Exception as err:
            _LOGGER.error("Error setting light state: %s", err)
            return False

    async def async_set_climate_temperature(self, climate_id: str, temperature: float) -> bool:
        """Set climate target temperature via API."""
        try:
            url = f"{self.base_url}/api/climate/{climate_id}/temperature"
            data = {"target_temperature": temperature}
            
            async with aiohttp.ClientSession() as session:
                async with session.put(url, json=data, headers=self.headers) as response:
                    return response.status == 200
        except Exception as err:
            _LOGGER.error("Error setting climate temperature: %s", err)
            return False

    async def async_set_climate_mode(self, climate_id: str, mode: str) -> bool:
        """Set climate HVAC mode via API."""
        try:
            url = f"{self.base_url}/api/climate/{climate_id}/mode"
            data = {"hvac_mode": mode}
            
            async with aiohttp.ClientSession() as session:
                async with session.put(url, json=data, headers=self.headers) as response:
                    return response.status == 200
        except Exception as err:
            _LOGGER.error("Error setting climate mode: %s", err)
            return False

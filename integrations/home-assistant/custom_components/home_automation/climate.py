"""Climate platform for Home Automation integration."""
import logging
from typing import Any, Dict, List, Optional

from homeassistant.components.climate import (
    ClimateEntity,
    ClimateEntityDescription,
    ClimateEntityFeature,
    HVACMode,
)
from homeassistant.config_entries import ConfigEntry
from homeassistant.const import UnitOfTemperature, ATTR_TEMPERATURE
from homeassistant.core import HomeAssistant
from homeassistant.helpers.entity_platform import AddEntitiesCallback
from homeassistant.helpers.update_coordinator import CoordinatorEntity

from .const import DOMAIN
from .coordinator import HomeAutomationCoordinator

_LOGGER = logging.getLogger(__name__)

async def async_setup_entry(
    hass: HomeAssistant,
    entry: ConfigEntry,
    async_add_entities: AddEntitiesCallback,
) -> None:
    """Set up Home Automation climate platform."""
    coordinator = hass.data[DOMAIN][entry.entry_id]
    entities = []

    if coordinator.data:
        # Add climate devices from API
        if "climate_devices" in coordinator.data:
            for climate in coordinator.data["climate_devices"]:
                entities.append(HomeAutomationClimate(coordinator, climate))

        # Add room-based climate control
        if "rooms" in coordinator.data:
            for room in coordinator.data["rooms"]:
                room_id = room.get("id")
                room_name = room.get("name", room_id)
                
                # Add climate control for the room
                entities.append(
                    HomeAutomationRoomClimate(
                        coordinator,
                        room,
                        f"{room_name} Climate"
                    )
                )

    async_add_entities(entities)

class HomeAutomationClimate(CoordinatorEntity, ClimateEntity):
    """Representation of a Home Automation climate device."""

    def __init__(
        self,
        coordinator: HomeAutomationCoordinator,
        climate_data: Dict[str, Any],
    ) -> None:
        """Initialize the climate device."""
        super().__init__(coordinator)
        self._climate_data = climate_data
        self._attr_name = climate_data.get("name", f"Climate {climate_data['id']}")
        self._attr_unique_id = f"{DOMAIN}_{climate_data['id']}"
        
        # Set supported features based on device capabilities
        capabilities = climate_data.get("capabilities", {})
        features = ClimateEntityFeature(0)
        
        if capabilities.get("target_temperature", True):
            features |= ClimateEntityFeature.TARGET_TEMPERATURE
        if capabilities.get("fan_mode", False):
            features |= ClimateEntityFeature.FAN_MODE
        if capabilities.get("swing_mode", False):
            features |= ClimateEntityFeature.SWING_MODE
        if capabilities.get("preset_mode", False):
            features |= ClimateEntityFeature.PRESET_MODE
            
        self._attr_supported_features = features
        
        # Set HVAC modes
        self._attr_hvac_modes = [
            HVACMode.OFF,
            HVACMode.HEAT,
            HVACMode.COOL,
            HVACMode.AUTO,
        ]
        
        # Temperature unit
        self._attr_temperature_unit = UnitOfTemperature.FAHRENHEIT
        
        # Temperature range
        self._attr_min_temp = climate_data.get("min_temp", 50)
        self._attr_max_temp = climate_data.get("max_temp", 90)

    @property
    def current_temperature(self) -> Optional[float]:
        """Return the current temperature."""
        if self.coordinator.data and "climate_devices" in self.coordinator.data:
            for climate in self.coordinator.data["climate_devices"]:
                if climate.get("id") == self._climate_data["id"]:
                    return climate.get("current_temperature")
        return None

    @property
    def target_temperature(self) -> Optional[float]:
        """Return the target temperature."""
        if self.coordinator.data and "climate_devices" in self.coordinator.data:
            for climate in self.coordinator.data["climate_devices"]:
                if climate.get("id") == self._climate_data["id"]:
                    return climate.get("target_temperature")
        return None

    @property
    def hvac_mode(self) -> Optional[HVACMode]:
        """Return current operation mode."""
        if self.coordinator.data and "climate_devices" in self.coordinator.data:
            for climate in self.coordinator.data["climate_devices"]:
                if climate.get("id") == self._climate_data["id"]:
                    mode = climate.get("hvac_mode", "off").lower()
                    if mode in ["heat", "cool", "auto", "off"]:
                        return HVACMode(mode)
        return HVACMode.OFF

    @property
    def available(self) -> bool:
        """Return if entity is available."""
        return self.coordinator.last_update_success

    async def async_set_temperature(self, **kwargs: Any) -> None:
        """Set new target temperature."""
        climate_id = self._climate_data["id"]
        temperature = kwargs.get(ATTR_TEMPERATURE)
        
        if temperature is None:
            return
            
        try:
            success = await self.coordinator.async_set_climate_temperature(
                climate_id, temperature
            )
            if success:
                await self.coordinator.async_request_refresh()
        except Exception as err:
            _LOGGER.error("Error setting temperature for climate %s: %s", climate_id, err)

    async def async_set_hvac_mode(self, hvac_mode: HVACMode) -> None:
        """Set new target hvac mode."""
        climate_id = self._climate_data["id"]
        try:
            success = await self.coordinator.async_set_climate_mode(
                climate_id, hvac_mode
            )
            if success:
                await self.coordinator.async_request_refresh()
        except Exception as err:
            _LOGGER.error("Error setting HVAC mode for climate %s: %s", climate_id, err)

class HomeAutomationRoomClimate(CoordinatorEntity, ClimateEntity):
    """Representation of a room-based climate control."""

    def __init__(
        self,
        coordinator: HomeAutomationCoordinator,
        room_data: Dict[str, Any],
        name: str,
    ) -> None:
        """Initialize the room climate control."""
        super().__init__(coordinator)
        self._room_data = room_data
        self._attr_name = name
        self._attr_unique_id = f"{DOMAIN}_{room_data['id']}_climate"
        
        # Basic features
        self._attr_supported_features = ClimateEntityFeature.TARGET_TEMPERATURE
        
        # HVAC modes
        self._attr_hvac_modes = [
            HVACMode.OFF,
            HVACMode.HEAT,
            HVACMode.COOL,
            HVACMode.AUTO,
        ]
        
        # Temperature unit and range
        self._attr_temperature_unit = UnitOfTemperature.FAHRENHEIT
        self._attr_min_temp = 50
        self._attr_max_temp = 90

    @property
    def current_temperature(self) -> Optional[float]:
        """Return the current temperature from room sensors."""
        # This would come from MQTT temperature sensor data
        # For now, return None until MQTT integration is set up
        return None

    @property
    def target_temperature(self) -> Optional[float]:
        """Return the target temperature."""
        # This would be stored in room state/MQTT
        # For now, return a default
        return 72

    @property
    def hvac_mode(self) -> Optional[HVACMode]:
        """Return current operation mode."""
        # This would come from MQTT state
        # For now, return OFF
        return HVACMode.OFF

    @property
    def available(self) -> bool:
        """Return if entity is available."""
        return self.coordinator.last_update_success

    async def async_set_temperature(self, **kwargs: Any) -> None:
        """Set new target temperature."""
        room_id = self._room_data["id"]
        temperature = kwargs.get(ATTR_TEMPERATURE)
        
        if temperature is None:
            return
            
        try:
            # This would send MQTT command to set room temperature
            # For now, just log the action
            _LOGGER.info("Setting temperature to %s for room %s", temperature, room_id)
            # await self.coordinator.async_publish_mqtt(f"room-climate/{room_id}/target_temp/set", str(temperature))
        except Exception as err:
            _LOGGER.error("Error setting temperature for room %s: %s", room_id, err)

    async def async_set_hvac_mode(self, hvac_mode: HVACMode) -> None:
        """Set new target hvac mode."""
        room_id = self._room_data["id"]
        try:
            # This would send MQTT command to set HVAC mode
            # For now, just log the action
            _LOGGER.info("Setting HVAC mode to %s for room %s", hvac_mode, room_id)
            # await self.coordinator.async_publish_mqtt(f"room-climate/{room_id}/mode/set", hvac_mode)
        except Exception as err:
            _LOGGER.error("Error setting HVAC mode for room %s: %s", room_id, err)

    @property
    def device_info(self) -> Dict[str, Any]:
        """Return device info for grouping entities."""
        room_id = self._room_data.get("id")
        room_name = self._room_data.get("name", room_id)
        
        return {
            "identifiers": {(DOMAIN, f"room_{room_id}")},
            "name": f"{room_name}",
            "manufacturer": "Home Automation System",
            "model": "Smart Room",
            "sw_version": "1.0",
        }

"""Light platform for Home Automation integration."""
import logging
from typing import Any, Dict, Optional

from homeassistant.components.light import (
    LightEntity,
    ColorMode,
    ATTR_BRIGHTNESS,
    ATTR_RGB_COLOR,
    ATTR_COLOR_TEMP,
)
from homeassistant.config_entries import ConfigEntry
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
    """Set up Home Automation light platform."""
    coordinator = hass.data[DOMAIN][entry.entry_id]
    entities = []

    if coordinator.data:
        # Add lights from API
        if "lights" in coordinator.data:
            for light in coordinator.data["lights"]:
                entities.append(HomeAutomationLight(coordinator, light))

        # Add room-based lights
        if "rooms" in coordinator.data:
            for room in coordinator.data["rooms"]:
                room_id = room.get("id")
                room_name = room.get("name", room_id)
                
                # Add main light for the room
                entities.append(
                    HomeAutomationRoomLight(
                        coordinator,
                        room,
                        "main",
                        f"{room_name} Light"
                    )
                )

    async_add_entities(entities)

class HomeAutomationLight(CoordinatorEntity, LightEntity):
    """Representation of a Home Automation light."""

    def __init__(
        self,
        coordinator: HomeAutomationCoordinator,
        light_data: Dict[str, Any],
    ) -> None:
        """Initialize the light."""
        super().__init__(coordinator)
        self._light_data = light_data
        self._attr_name = light_data.get("name", f"Light {light_data['id']}")
        self._attr_unique_id = f"{DOMAIN}_{light_data['id']}"
        
        # Determine supported color modes based on light capabilities
        capabilities = light_data.get("capabilities", {})
        color_modes = set()
        
        if capabilities.get("brightness", False):
            color_modes.add(ColorMode.BRIGHTNESS)
        if capabilities.get("rgb", False):
            color_modes.add(ColorMode.RGB)
        if capabilities.get("color_temp", False):
            color_modes.add(ColorMode.COLOR_TEMP)
        if not color_modes:
            color_modes.add(ColorMode.ONOFF)
            
        self._attr_supported_color_modes = color_modes
        self._attr_color_mode = list(color_modes)[0]

    @property
    def is_on(self) -> Optional[bool]:
        """Return the state of the light."""
        # Get updated light data
        if self.coordinator.data and "lights" in self.coordinator.data:
            for light in self.coordinator.data["lights"]:
                if light.get("id") == self._light_data["id"]:
                    state = light.get("state", light.get("value"))
                    if isinstance(state, bool):
                        return state
                    if isinstance(state, str):
                        return state.lower() in ("true", "on", "1", "yes")
                    if isinstance(state, (int, float)):
                        return state > 0
        return None

    @property
    def brightness(self) -> Optional[int]:
        """Return the brightness of the light."""
        if ColorMode.BRIGHTNESS in self.supported_color_modes:
            if self.coordinator.data and "lights" in self.coordinator.data:
                for light in self.coordinator.data["lights"]:
                    if light.get("id") == self._light_data["id"]:
                        brightness = light.get("brightness")
                        if brightness is not None:
                            # Convert from 0-100 to 0-255
                            return int(brightness * 255 / 100)
        return None

    @property
    def rgb_color(self) -> Optional[tuple]:
        """Return the RGB color of the light."""
        if ColorMode.RGB in self.supported_color_modes:
            if self.coordinator.data and "lights" in self.coordinator.data:
                for light in self.coordinator.data["lights"]:
                    if light.get("id") == self._light_data["id"]:
                        rgb = light.get("rgb_color")
                        if rgb and len(rgb) == 3:
                            return tuple(rgb)
        return None

    @property
    def color_temp(self) -> Optional[int]:
        """Return the color temperature of the light."""
        if ColorMode.COLOR_TEMP in self.supported_color_modes:
            if self.coordinator.data and "lights" in self.coordinator.data:
                for light in self.coordinator.data["lights"]:
                    if light.get("id") == self._light_data["id"]:
                        return light.get("color_temp")
        return None

    @property
    def available(self) -> bool:
        """Return if entity is available."""
        return self.coordinator.last_update_success

    async def async_turn_on(self, **kwargs: Any) -> None:
        """Turn the light on."""
        light_id = self._light_data["id"]
        try:
            success = await self.coordinator.async_set_light_state(
                light_id, True, kwargs
            )
            if success:
                await self.coordinator.async_request_refresh()
        except Exception as err:
            _LOGGER.error("Error turning on light %s: %s", light_id, err)

    async def async_turn_off(self, **kwargs: Any) -> None:
        """Turn the light off."""
        light_id = self._light_data["id"]
        try:
            success = await self.coordinator.async_set_light_state(
                light_id, False, kwargs
            )
            if success:
                await self.coordinator.async_request_refresh()
        except Exception as err:
            _LOGGER.error("Error turning off light %s: %s", light_id, err)

class HomeAutomationRoomLight(CoordinatorEntity, LightEntity):
    """Representation of a room-based light."""

    def __init__(
        self,
        coordinator: HomeAutomationCoordinator,
        room_data: Dict[str, Any],
        light_type: str,
        name: str,
    ) -> None:
        """Initialize the room light."""
        super().__init__(coordinator)
        self._room_data = room_data
        self._light_type = light_type
        self._attr_name = name
        self._attr_unique_id = f"{DOMAIN}_{room_data['id']}_{light_type}_light"
        
        # Default to basic on/off functionality
        self._attr_supported_color_modes = {ColorMode.ONOFF}
        self._attr_color_mode = ColorMode.ONOFF

    @property
    def is_on(self) -> Optional[bool]:
        """Return the light state from MQTT/room data."""
        # This would be populated by MQTT light data
        # For now, return False until MQTT integration is set up
        return False

    @property
    def available(self) -> bool:
        """Return if entity is available."""
        return self.coordinator.last_update_success

    async def async_turn_on(self, **kwargs: Any) -> None:
        """Turn the room light on."""
        room_id = self._room_data["id"]
        try:
            # This would send MQTT command to turn on room light
            # For now, just log the action
            _LOGGER.info("Turning on %s light for room %s", self._light_type, room_id)
            # await self.coordinator.async_publish_mqtt(f"room-light/{room_id}/set", "true")
        except Exception as err:
            _LOGGER.error("Error turning on room light %s/%s: %s", room_id, self._light_type, err)

    async def async_turn_off(self, **kwargs: Any) -> None:
        """Turn the room light off."""
        room_id = self._room_data["id"]
        try:
            # This would send MQTT command to turn off room light
            # For now, just log the action
            _LOGGER.info("Turning off %s light for room %s", self._light_type, room_id)
            # await self.coordinator.async_publish_mqtt(f"room-light/{room_id}/set", "false")
        except Exception as err:
            _LOGGER.error("Error turning off room light %s/%s: %s", room_id, self._light_type, err)

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

"""Switch platform for Home Automation integration."""
import logging
from typing import Any, Dict, Optional

from homeassistant.components.switch import SwitchEntity
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
    """Set up Home Automation switch platform."""
    coordinator = hass.data[DOMAIN][entry.entry_id]
    entities = []

    if coordinator.data:
        # Add switches from API
        if "switches" in coordinator.data:
            for switch in coordinator.data["switches"]:
                entities.append(HomeAutomationSwitch(coordinator, switch))

        # Add room-based switches
        if "rooms" in coordinator.data:
            for room in coordinator.data["rooms"]:
                room_id = room.get("id")
                room_name = room.get("name", room_id)
                
                # Add a general switch for the room
                entities.append(
                    HomeAutomationRoomSwitch(
                        coordinator,
                        room,
                        "power",
                        f"{room_name} Power"
                    )
                )

    async_add_entities(entities)

class HomeAutomationSwitch(CoordinatorEntity, SwitchEntity):
    """Representation of a Home Automation switch."""

    def __init__(
        self,
        coordinator: HomeAutomationCoordinator,
        switch_data: Dict[str, Any],
    ) -> None:
        """Initialize the switch."""
        super().__init__(coordinator)
        self._switch_data = switch_data
        self._attr_name = switch_data.get("name", f"Switch {switch_data['id']}")
        self._attr_unique_id = f"{DOMAIN}_{switch_data['id']}"

    @property
    def is_on(self) -> Optional[bool]:
        """Return the state of the switch."""
        # Get updated switch data
        if self.coordinator.data and "switches" in self.coordinator.data:
            for switch in self.coordinator.data["switches"]:
                if switch.get("id") == self._switch_data["id"]:
                    state = switch.get("state", switch.get("value"))
                    if isinstance(state, bool):
                        return state
                    if isinstance(state, str):
                        return state.lower() in ("true", "on", "1", "yes")
                    if isinstance(state, (int, float)):
                        return state > 0
        return None

    @property
    def available(self) -> bool:
        """Return if entity is available."""
        return self.coordinator.last_update_success

    async def async_turn_on(self, **kwargs: Any) -> None:
        """Turn the switch on."""
        switch_id = self._switch_data["id"]
        try:
            success = await self.coordinator.async_set_switch_state(switch_id, True)
            if success:
                await self.coordinator.async_request_refresh()
        except Exception as err:
            _LOGGER.error("Error turning on switch %s: %s", switch_id, err)

    async def async_turn_off(self, **kwargs: Any) -> None:
        """Turn the switch off."""
        switch_id = self._switch_data["id"]
        try:
            success = await self.coordinator.async_set_switch_state(switch_id, False)
            if success:
                await self.coordinator.async_request_refresh()
        except Exception as err:
            _LOGGER.error("Error turning off switch %s: %s", switch_id, err)

class HomeAutomationRoomSwitch(CoordinatorEntity, SwitchEntity):
    """Representation of a room-based switch."""

    def __init__(
        self,
        coordinator: HomeAutomationCoordinator,
        room_data: Dict[str, Any],
        switch_type: str,
        name: str,
    ) -> None:
        """Initialize the room switch."""
        super().__init__(coordinator)
        self._room_data = room_data
        self._switch_type = switch_type
        self._attr_name = name
        self._attr_unique_id = f"{DOMAIN}_{room_data['id']}_{switch_type}"

    @property
    def is_on(self) -> Optional[bool]:
        """Return the switch state from MQTT/room data."""
        # This would be populated by MQTT switch data
        # For now, return False until MQTT integration is set up
        return False

    @property
    def available(self) -> bool:
        """Return if entity is available."""
        return self.coordinator.last_update_success

    async def async_turn_on(self, **kwargs: Any) -> None:
        """Turn the room switch on."""
        room_id = self._room_data["id"]
        try:
            # This would send MQTT command to turn on room switch
            # For now, just log the action
            _LOGGER.info("Turning on %s for room %s", self._switch_type, room_id)
            # await self.coordinator.async_publish_mqtt(f"room-{self._switch_type}/{room_id}/set", "true")
        except Exception as err:
            _LOGGER.error("Error turning on room switch %s/%s: %s", room_id, self._switch_type, err)

    async def async_turn_off(self, **kwargs: Any) -> None:
        """Turn the room switch off."""
        room_id = self._room_data["id"]
        try:
            # This would send MQTT command to turn off room switch
            # For now, just log the action
            _LOGGER.info("Turning off %s for room %s", self._switch_type, room_id)
            # await self.coordinator.async_publish_mqtt(f"room-{self._switch_type}/{room_id}/set", "false")
        except Exception as err:
            _LOGGER.error("Error turning off room switch %s/%s: %s", room_id, self._switch_type, err)

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

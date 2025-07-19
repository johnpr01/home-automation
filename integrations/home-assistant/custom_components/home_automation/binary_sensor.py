"""Binary sensor platform for Home Automation integration."""
import logging
from typing import Any, Dict, Optional

from homeassistant.components.binary_sensor import (
    BinarySensorEntity,
    BinarySensorEntityDescription,
    BinarySensorDeviceClass,
)
from homeassistant.config_entries import ConfigEntry
from homeassistant.core import HomeAssistant
from homeassistant.helpers.entity_platform import AddEntitiesCallback
from homeassistant.helpers.update_coordinator import CoordinatorEntity

from .const import DOMAIN
from .coordinator import HomeAutomationCoordinator

_LOGGER = logging.getLogger(__name__)

BINARY_SENSOR_DESCRIPTIONS = {
    "motion": BinarySensorEntityDescription(
        key="motion",
        name="Motion",
        device_class=BinarySensorDeviceClass.MOTION,
    ),
    "door": BinarySensorEntityDescription(
        key="door",
        name="Door",
        device_class=BinarySensorDeviceClass.DOOR,
    ),
    "window": BinarySensorEntityDescription(
        key="window", 
        name="Window",
        device_class=BinarySensorDeviceClass.WINDOW,
    ),
    "occupancy": BinarySensorEntityDescription(
        key="occupancy",
        name="Occupancy",
        device_class=BinarySensorDeviceClass.OCCUPANCY,
    ),
}

async def async_setup_entry(
    hass: HomeAssistant,
    entry: ConfigEntry,
    async_add_entities: AddEntitiesCallback,
) -> None:
    """Set up Home Automation binary sensor platform."""
    coordinator = hass.data[DOMAIN][entry.entry_id]
    entities = []

    if coordinator.data:
        # Add binary sensors from API
        if "binary_sensors" in coordinator.data:
            for sensor in coordinator.data["binary_sensors"]:
                sensor_type = sensor.get("type", "").lower()
                if sensor_type in BINARY_SENSOR_DESCRIPTIONS:
                    entities.append(
                        HomeAutomationBinarySensor(
                            coordinator,
                            sensor,
                            BINARY_SENSOR_DESCRIPTIONS[sensor_type]
                        )
                    )

        # Add room-based binary sensors
        if "rooms" in coordinator.data:
            for room in coordinator.data["rooms"]:
                room_id = room.get("id")
                room_name = room.get("name", room_id)
                
                # Motion sensor for room
                entities.append(
                    HomeAutomationRoomBinarySensor(
                        coordinator,
                        room,
                        "motion",
                        BINARY_SENSOR_DESCRIPTIONS["motion"],
                        f"{room_name} Motion"
                    )
                )
                
                # Occupancy sensor for room
                entities.append(
                    HomeAutomationRoomBinarySensor(
                        coordinator,
                        room,
                        "occupancy",
                        BINARY_SENSOR_DESCRIPTIONS["occupancy"],
                        f"{room_name} Occupancy"
                    )
                )

    async_add_entities(entities)

class HomeAutomationBinarySensor(CoordinatorEntity, BinarySensorEntity):
    """Representation of a Home Automation binary sensor."""

    def __init__(
        self,
        coordinator: HomeAutomationCoordinator,
        sensor_data: Dict[str, Any],
        description: BinarySensorEntityDescription,
    ) -> None:
        """Initialize the binary sensor."""
        super().__init__(coordinator)
        self.entity_description = description
        self._sensor_data = sensor_data
        self._attr_name = sensor_data.get("name", description.name)
        self._attr_unique_id = f"{DOMAIN}_{sensor_data['id']}_{description.key}"

    @property
    def is_on(self) -> Optional[bool]:
        """Return the state of the binary sensor."""
        # Get updated sensor data
        if self.coordinator.data and "binary_sensors" in self.coordinator.data:
            for sensor in self.coordinator.data["binary_sensors"]:
                if sensor.get("id") == self._sensor_data["id"]:
                    value = sensor.get("value")
                    # Convert various true/false representations
                    if isinstance(value, bool):
                        return value
                    if isinstance(value, str):
                        return value.lower() in ("true", "on", "1", "yes")
                    if isinstance(value, (int, float)):
                        return value > 0
        return None

    @property
    def available(self) -> bool:
        """Return if entity is available."""
        return self.coordinator.last_update_success

class HomeAutomationRoomBinarySensor(CoordinatorEntity, BinarySensorEntity):
    """Representation of a room-based binary sensor from MQTT data."""

    def __init__(
        self,
        coordinator: HomeAutomationCoordinator,
        room_data: Dict[str, Any],
        sensor_type: str,
        description: BinarySensorEntityDescription,
        name: str,
    ) -> None:
        """Initialize the room binary sensor."""
        super().__init__(coordinator)
        self.entity_description = description
        self._room_data = room_data
        self._sensor_type = sensor_type
        self._attr_name = name
        self._attr_unique_id = f"{DOMAIN}_{room_data['id']}_{sensor_type}"

    @property
    def is_on(self) -> Optional[bool]:
        """Return the binary sensor state from MQTT/room data."""
        # This would be populated by MQTT sensor data
        # For now, return None until MQTT integration is set up
        return None

    @property
    def available(self) -> bool:
        """Return if entity is available."""
        return self.coordinator.last_update_success

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

"""Sensor platform for Home Automation integration."""
import logging
from typing import Any, Dict, List, Optional

from homeassistant.components.sensor import (
    SensorEntity,
    SensorEntityDescription,
    SensorDeviceClass,
    SensorStateClass,
)
from homeassistant.config_entries import ConfigEntry
from homeassistant.const import (
    UnitOfTemperature,
    PERCENTAGE,
    LIGHT_LUX,
)
from homeassistant.core import HomeAssistant
from homeassistant.helpers.entity_platform import AddEntitiesCallback
from homeassistant.helpers.update_coordinator import CoordinatorEntity

from .const import DOMAIN
from .coordinator import HomeAutomationCoordinator

_LOGGER = logging.getLogger(__name__)

SENSOR_DESCRIPTIONS = {
    "temperature": SensorEntityDescription(
        key="temperature",
        name="Temperature",
        device_class=SensorDeviceClass.TEMPERATURE,
        state_class=SensorStateClass.MEASUREMENT,
        native_unit_of_measurement=UnitOfTemperature.FAHRENHEIT,
    ),
    "humidity": SensorEntityDescription(
        key="humidity", 
        name="Humidity",
        device_class=SensorDeviceClass.HUMIDITY,
        state_class=SensorStateClass.MEASUREMENT,
        native_unit_of_measurement=PERCENTAGE,
    ),
    "light_level": SensorEntityDescription(
        key="light_level",
        name="Light Level",
        device_class=SensorDeviceClass.ILLUMINANCE,
        state_class=SensorStateClass.MEASUREMENT,
        native_unit_of_measurement=PERCENTAGE,
    ),
}

async def async_setup_entry(
    hass: HomeAssistant,
    entry: ConfigEntry,
    async_add_entities: AddEntitiesCallback,
) -> None:
    """Set up Home Automation sensor platform."""
    coordinator = hass.data[DOMAIN][entry.entry_id]
    entities = []

    if coordinator.data:
        # Add sensors from API
        if "sensors" in coordinator.data:
            for sensor in coordinator.data["sensors"]:
                sensor_type = sensor.get("type", "").lower()
                if sensor_type in SENSOR_DESCRIPTIONS:
                    entities.append(
                        HomeAutomationSensor(
                            coordinator,
                            sensor,
                            SENSOR_DESCRIPTIONS[sensor_type]
                        )
                    )

        # Add room-based sensors
        if "rooms" in coordinator.data:
            for room in coordinator.data["rooms"]:
                room_id = room.get("id")
                room_name = room.get("name", room_id)
                
                # Temperature sensor for room
                entities.append(
                    HomeAutomationRoomSensor(
                        coordinator,
                        room,
                        "temperature", 
                        SENSOR_DESCRIPTIONS["temperature"],
                        f"{room_name} Temperature"
                    )
                )
                
                # Humidity sensor for room  
                entities.append(
                    HomeAutomationRoomSensor(
                        coordinator,
                        room,
                        "humidity",
                        SENSOR_DESCRIPTIONS["humidity"], 
                        f"{room_name} Humidity"
                    )
                )
                
                # Light level sensor for room
                entities.append(
                    HomeAutomationRoomSensor(
                        coordinator,
                        room,
                        "light_level",
                        SENSOR_DESCRIPTIONS["light_level"],
                        f"{room_name} Light Level"
                    )
                )

    async_add_entities(entities)

class HomeAutomationSensor(CoordinatorEntity, SensorEntity):
    """Representation of a Home Automation sensor."""

    def __init__(
        self,
        coordinator: HomeAutomationCoordinator,
        sensor_data: Dict[str, Any],
        description: SensorEntityDescription,
    ) -> None:
        """Initialize the sensor."""
        super().__init__(coordinator)
        self.entity_description = description
        self._sensor_data = sensor_data
        self._attr_name = sensor_data.get("name", description.name)
        self._attr_unique_id = f"{DOMAIN}_{sensor_data['id']}_{description.key}"

    @property
    def native_value(self) -> Optional[str]:
        """Return the state of the sensor."""
        # Get updated sensor data
        if self.coordinator.data and "sensors" in self.coordinator.data:
            for sensor in self.coordinator.data["sensors"]:
                if sensor.get("id") == self._sensor_data["id"]:
                    return sensor.get("value")
        return None

    @property
    def available(self) -> bool:
        """Return if entity is available."""
        return self.coordinator.last_update_success

class HomeAutomationRoomSensor(CoordinatorEntity, SensorEntity):
    """Representation of a room-based sensor from MQTT data."""

    def __init__(
        self,
        coordinator: HomeAutomationCoordinator,
        room_data: Dict[str, Any],
        sensor_type: str,
        description: SensorEntityDescription,
        name: str,
    ) -> None:
        """Initialize the room sensor.""" 
        super().__init__(coordinator)
        self.entity_description = description
        self._room_data = room_data
        self._sensor_type = sensor_type
        self._attr_name = name
        self._attr_unique_id = f"{DOMAIN}_{room_data['id']}_{sensor_type}"
        
    @property
    def native_value(self) -> Optional[str]:
        """Return the sensor value from MQTT/room data."""
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

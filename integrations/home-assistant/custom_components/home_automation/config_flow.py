"""Config flow for Home Automation integration."""
import logging
from typing import Any, Dict, Optional
import voluptuous as vol
import asyncio

import aiohttp
import async_timeout
from homeassistant import config_entries
from homeassistant.const import CONF_HOST, CONF_PORT
from homeassistant.core import HomeAssistant
from homeassistant.data_entry_flow import FlowResult
from homeassistant.helpers import aiohttp_client

from .const import DOMAIN, DEFAULT_HOST, DEFAULT_PORT, ERROR_CANNOT_CONNECT, ERROR_TIMEOUT

_LOGGER = logging.getLogger(__name__)

STEP_USER_DATA_SCHEMA = vol.Schema(
    {
        vol.Required(CONF_HOST, default=DEFAULT_HOST): str,
        vol.Required(CONF_PORT, default=DEFAULT_PORT): int,
        vol.Optional("mqtt_host", default=DEFAULT_HOST): str,
        vol.Optional("mqtt_port", default=1883): int,
    }
)

async def validate_input(hass: HomeAssistant, data: Dict[str, Any]) -> Dict[str, Any]:
    """Validate the user input allows us to connect."""
    session = aiohttp_client.async_get_clientsession(hass)
    
    host = data[CONF_HOST]
    port = data[CONF_PORT]
    url = f"http://{host}:{port}/api/status"
    
    try:
        async with async_timeout.timeout(10):
            async with session.get(url) as response:
                if response.status != 200:
                    raise Exception("Invalid response from server")
                
                result = await response.json()
                if not isinstance(result, dict):
                    raise Exception("Invalid JSON response")
                    
                return {"title": f"Home Automation ({host}:{port})"}
                
    except asyncio.TimeoutError:
        raise Exception(ERROR_TIMEOUT)
    except (aiohttp.ClientError, aiohttp.ClientResponseError):
        raise Exception(ERROR_CANNOT_CONNECT)
    except Exception as err:
        _LOGGER.error(f"Unexpected error: {err}")
        raise Exception(ERROR_CANNOT_CONNECT)

class ConfigFlow(config_entries.ConfigFlow, domain=DOMAIN):
    """Handle a config flow for Home Automation."""

    VERSION = 1

    async def async_step_user(
        self, user_input: Optional[Dict[str, Any]] = None
    ) -> FlowResult:
        """Handle the initial step."""
        if user_input is None:
            return self.async_show_form(
                step_id="user", data_schema=STEP_USER_DATA_SCHEMA
            )

        errors = {}

        try:
            info = await validate_input(self.hass, user_input)
        except Exception as err:
            _LOGGER.exception("Unexpected exception")
            if str(err) == ERROR_CANNOT_CONNECT:
                errors["base"] = "cannot_connect"
            elif str(err) == ERROR_TIMEOUT:
                errors["base"] = "timeout_connect"
            else:
                errors["base"] = "unknown"
        else:
            # Check if already configured
            await self.async_set_unique_id(f"{user_input[CONF_HOST]}:{user_input[CONF_PORT]}")
            self._abort_if_unique_id_configured()
            
            return self.async_create_entry(title=info["title"], data=user_input)

        return self.async_show_form(
            step_id="user", data_schema=STEP_USER_DATA_SCHEMA, errors=errors
        )

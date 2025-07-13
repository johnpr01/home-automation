// Home Automation Dashboard JavaScript

class HomeAutomationApp {
    constructor() {
        this.apiBaseUrl = '/api';
        this.devices = [];
        this.sensors = [];
        this.init();
    }

    async init() {
        await this.loadSystemStatus();
        await this.loadDevices();
        await this.loadSensors();
        this.setupEventListeners();
        this.startPolling();
    }

    async loadSystemStatus() {
        try {
            const response = await fetch(`${this.apiBaseUrl}/status`);
            const status = await response.json();
            this.updateSystemStatus(status);
        } catch (error) {
            console.error('Failed to load system status:', error);
            this.updateSystemStatus({ status: 'error', devices_count: 0, sensors_count: 0 });
        }
    }

    updateSystemStatus(status) {
        const statusElement = document.getElementById('system-status');
        const deviceCountElement = document.getElementById('device-count');
        const sensorCountElement = document.getElementById('sensor-count');

        if (statusElement) {
            statusElement.textContent = status.status === 'ok' ? 'Online' : 'Offline';
            statusElement.className = `status-indicator ${status.status === 'ok' ? 'online' : 'offline'}`;
        }

        if (deviceCountElement) {
            deviceCountElement.textContent = status.devices_count || 0;
        }

        if (sensorCountElement) {
            sensorCountElement.textContent = status.sensors_count || 0;
        }
    }

    async loadDevices() {
        try {
            const response = await fetch(`${this.apiBaseUrl}/devices`);
            this.devices = await response.json();
            this.renderDevices();
        } catch (error) {
            console.error('Failed to load devices:', error);
            this.devices = [];
            this.renderDevices();
        }
    }

    renderDevices() {
        const container = document.getElementById('devices-container');
        if (!container) return;

        if (this.devices.length === 0) {
            container.innerHTML = '<p>No devices found.</p>';
            return;
        }

        container.innerHTML = this.devices.map(device => `
            <div class="device-card" data-device-id="${device.id}">
                <div class="device-header">
                    <div class="device-name">${device.name}</div>
                    <div class="device-type">${device.type}</div>
                </div>
                <div class="device-status">
                    <div class="status-dot ${device.status}"></div>
                    <span>${device.status.charAt(0).toUpperCase() + device.status.slice(1)}</span>
                </div>
                <div class="device-controls">
                    ${this.renderDeviceControls(device)}
                </div>
            </div>
        `).join('');
    }

    renderDeviceControls(device) {
        switch (device.type) {
            case 'light':
                return `
                    <button class="btn btn-primary" onclick="app.toggleDevice('${device.id}')">
                        ${device.status === 'on' ? 'Turn Off' : 'Turn On'}
                    </button>
                    ${device.status === 'on' ? `
                        <button class="btn btn-secondary" onclick="app.dimDevice('${device.id}')">
                            Dim
                        </button>
                    ` : ''}
                `;
            case 'switch':
                return `
                    <button class="btn btn-primary" onclick="app.toggleDevice('${device.id}')">
                        ${device.status === 'on' ? 'Turn Off' : 'Turn On'}
                    </button>
                `;
            case 'climate':
                return `
                    <button class="btn btn-secondary" onclick="app.adjustTemperature('${device.id}', -1)">
                        -
                    </button>
                    <span>22°C</span>
                    <button class="btn btn-secondary" onclick="app.adjustTemperature('${device.id}', 1)">
                        +
                    </button>
                `;
            default:
                return '<span class="text-muted">No controls available</span>';
        }
    }

    async loadSensors() {
        try {
            const response = await fetch(`${this.apiBaseUrl}/sensors`);
            this.sensors = await response.json();
            this.renderSensors();
        } catch (error) {
            console.error('Failed to load sensors:', error);
            this.sensors = [];
            this.renderSensors();
        }
    }

    renderSensors() {
        const container = document.getElementById('sensors-container');
        if (!container) return;

        if (this.sensors.length === 0) {
            container.innerHTML = '<p>No sensors found.</p>';
            return;
        }

        container.innerHTML = this.sensors.map(sensor => `
            <div class="sensor-card" data-sensor-id="${sensor.id}">
                <div class="sensor-header">
                    <div class="sensor-name">${sensor.name}</div>
                    <div class="sensor-type">${sensor.type}</div>
                </div>
                <div class="sensor-value">
                    ${this.formatSensorValue(sensor.value, sensor.type)}
                    <span class="sensor-unit">${this.getSensorUnit(sensor.type)}</span>
                </div>
                <div class="sensor-timestamp">
                    Last updated: ${this.formatTimestamp(sensor.last_updated)}
                </div>
            </div>
        `).join('');
    }

    formatSensorValue(value, type) {
        if (typeof value === 'number') {
            return type === 'temperature' ? value.toFixed(1) : value.toString();
        }
        return value.toString();
    }

    getSensorUnit(type) {
        const units = {
            temperature: '°C',
            humidity: '%',
            pressure: 'hPa',
            light: 'lux',
            motion: '',
            door: '',
            window: ''
        };
        return units[type] || '';
    }

    formatTimestamp(timestamp) {
        if (!timestamp) return 'Unknown';
        const date = new Date(timestamp);
        return date.toLocaleString();
    }

    async toggleDevice(deviceId) {
        const device = this.devices.find(d => d.id === deviceId);
        if (!device) return;

        const action = device.status === 'on' ? 'turn_off' : 'turn_on';
        
        try {
            const response = await fetch(`${this.apiBaseUrl}/devices/${deviceId}/command`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({ action })
            });

            if (response.ok) {
                // Update local state and re-render
                device.status = device.status === 'on' ? 'off' : 'on';
                this.renderDevices();
            } else {
                console.error('Failed to toggle device');
            }
        } catch (error) {
            console.error('Error toggling device:', error);
        }
    }

    async dimDevice(deviceId) {
        try {
            const response = await fetch(`${this.apiBaseUrl}/devices/${deviceId}/command`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({ 
                    action: 'set_brightness',
                    value: 50 
                })
            });

            if (response.ok) {
                console.log('Device dimmed successfully');
            } else {
                console.error('Failed to dim device');
            }
        } catch (error) {
            console.error('Error dimming device:', error);
        }
    }

    async adjustTemperature(deviceId, change) {
        try {
            const response = await fetch(`${this.apiBaseUrl}/devices/${deviceId}/command`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({ 
                    action: 'set_temperature',
                    value: change 
                })
            });

            if (response.ok) {
                console.log('Temperature adjusted successfully');
            } else {
                console.error('Failed to adjust temperature');
            }
        } catch (error) {
            console.error('Error adjusting temperature:', error);
        }
    }

    setupEventListeners() {
        // Navigation
        document.querySelectorAll('.nav-link').forEach(link => {
            link.addEventListener('click', (e) => {
                e.preventDefault();
                const targetId = link.getAttribute('href').substring(1);
                const targetElement = document.getElementById(targetId);
                if (targetElement) {
                    targetElement.scrollIntoView({ behavior: 'smooth' });
                }
            });
        });
    }

    startPolling() {
        // Poll for updates every 30 seconds
        setInterval(() => {
            this.loadSystemStatus();
            this.loadSensors(); // Sensors update more frequently
        }, 30000);

        // Poll devices less frequently (every 60 seconds)
        setInterval(() => {
            this.loadDevices();
        }, 60000);
    }
}

// Initialize the app when DOM is loaded
let app;
document.addEventListener('DOMContentLoaded', () => {
    app = new HomeAutomationApp();
});

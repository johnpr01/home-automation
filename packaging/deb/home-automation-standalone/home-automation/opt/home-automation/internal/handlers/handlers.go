package handlers

import (
	"encoding/json"
	"net/http"
)

func RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/", homeHandler)
	mux.HandleFunc("/api/devices", devicesHandler)
	mux.HandleFunc("/api/sensors", sensorsHandler)
	mux.HandleFunc("/api/status", statusHandler)
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("<h1>Home Automation System</h1><p>Welcome to your smart home!</p>"))
}

func devicesHandler(w http.ResponseWriter, r *http.Request) {
	devices := []map[string]interface{}{
		{"id": "1", "name": "Living Room Light", "type": "light", "status": "on"},
		{"id": "2", "name": "Thermostat", "type": "climate", "status": "auto"},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(devices)
}

func sensorsHandler(w http.ResponseWriter, r *http.Request) {
	sensors := []map[string]interface{}{
		{"id": "1", "name": "Temperature Sensor", "type": "temperature", "value": 22.5},
		{"id": "2", "name": "Motion Sensor", "type": "motion", "value": false},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(sensors)
}

func statusHandler(w http.ResponseWriter, r *http.Request) {
	status := map[string]interface{}{
		"status":        "ok",
		"uptime":        "24h",
		"devices_count": 2,
		"sensors_count": 2,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

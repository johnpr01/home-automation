package tapo

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/johnpr01/home-automation/internal/errors"
	"github.com/johnpr01/home-automation/internal/logger"
)

// TapoClient represents a client for TP-Link Tapo smart plugs
type TapoClient struct {
	baseURL    string
	username   string
	password   string
	httpClient *http.Client
	logger     *logger.Logger
	// sessionID removed as it was unused
	token      string
}

// TapoDevice represents a Tapo smart plug device
type TapoDevice struct {
	DeviceID        string `json:"device_id"`
	Nickname        string `json:"nickname"`
	Model           string `json:"model"`
	IPAddress       string `json:"ip_address"`
	MACAddress      string `json:"mac"`
	FirmwareVer     string `json:"fw_ver"`
	HardwareVer     string `json:"hw_ver"`
	Type            string `json:"type"`
	IsOn            bool   `json:"device_on"`
	SignalLevel     int    `json:"signal_level"`
	RSSI            int    `json:"rssi"`
	OverHeated      bool   `json:"overheated"`
	PowerProtection bool   `json:"power_protection"`
}

// EnergyUsage represents current energy consumption data
type EnergyUsage struct {
	CurrentPowerMw  int64  `json:"current_power"` // Power in milliwatts
	TodayRuntimeMin int    `json:"today_runtime"` // Runtime in minutes
	MonthRuntimeMin int    `json:"month_runtime"` // Runtime in minutes
	TodayEnergyWh   int    `json:"today_energy"`  // Energy in watt-hours
	MonthEnergyWh   int    `json:"month_energy"`  // Energy in watt-hours
	LocalTimeStr    string `json:"local_time"`
}

// EnergyStats represents detailed energy statistics
type EnergyStats struct {
	TodayEnergyWh int     `json:"today_energy"`
	MonthEnergyWh int     `json:"month_energy"`
	VoltageV      float64 `json:"voltage_mv"` // Voltage in millivolts (convert to V)
	CurrentA      float64 `json:"current_ma"` // Current in milliamps (convert to A)
	PowerW        float64 `json:"power_mw"`   // Power in milliwatts (convert to W)
}

// TapoResponse represents the standard Tapo API response
type TapoResponse struct {
	ErrorCode int                    `json:"error_code"`
	Result    map[string]interface{} `json:"result,omitempty"`
	Message   string                 `json:"msg,omitempty"`
}

// LoginRequest represents the login request payload
type LoginRequest struct {
	Method string                 `json:"method"`
	Params map[string]interface{} `json:"params"`
}

// NewTapoClient creates a new Tapo client
func NewTapoClient(deviceIP, username, password string, serviceLogger *logger.Logger) *TapoClient {
	return &TapoClient{
		baseURL:  fmt.Sprintf("http://%s", deviceIP),
		username: username,
		password: password,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		logger: serviceLogger,
	}
}

// Connect establishes connection and authenticates with the Tapo device
func (c *TapoClient) Connect() error {
	// Step 1: Handshake to get session
	handshakeURL := fmt.Sprintf("%s/app", c.baseURL)

	handshakeReq := LoginRequest{
		Method: "handshake",
		Params: map[string]interface{}{
			"key": generatePublicKey(),
		},
	}

	handshakeResp, err := c.makeRequest(handshakeURL, handshakeReq)
	if err != nil {
		return errors.NewConnectionError("Failed to handshake with Tapo device", err)
	}

	if handshakeResp.ErrorCode != 0 {
		return errors.NewConnectionError(fmt.Sprintf("Handshake failed with error code: %d", handshakeResp.ErrorCode), nil)
	}

	// Step 2: Login with credentials
	loginURL := fmt.Sprintf("%s/app", c.baseURL)

	loginReq := LoginRequest{
		Method: "login_device",
		Params: map[string]interface{}{
			"username": base64.StdEncoding.EncodeToString([]byte(c.username)),
			"password": base64.StdEncoding.EncodeToString([]byte(c.password)),
		},
	}

	loginResp, err := c.makeRequest(loginURL, loginReq)
	if err != nil {
		return errors.NewConnectionError("Failed to login to Tapo device", err)
	}

	if loginResp.ErrorCode != 0 {
		return errors.NewConnectionError(fmt.Sprintf("Login failed with error code: %d, message: %s", loginResp.ErrorCode, loginResp.Message), nil)
	}

	// Extract token from response
	if token, ok := loginResp.Result["token"].(string); ok {
		c.token = token
	}

	c.logger.Info("Successfully connected to Tapo device", map[string]interface{}{
		"device_ip": c.baseURL,
	})

	return nil
}

// GetDeviceInfo retrieves device information
func (c *TapoClient) GetDeviceInfo() (*TapoDevice, error) {
	if c.token == "" {
		return nil, errors.NewConnectionError("Not authenticated with Tapo device", nil)
	}

	req := LoginRequest{
		Method: "get_device_info",
		Params: map[string]interface{}{},
	}

	resp, err := c.makeAuthenticatedRequest(req)
	if err != nil {
		return nil, errors.NewDeviceError("Failed to get device info", err)
	}

	if resp.ErrorCode != 0 {
		return nil, errors.NewDeviceError(fmt.Sprintf("Get device info failed with error code: %d", resp.ErrorCode), nil)
	}

	// Parse device info
	device := &TapoDevice{}
	if err := mapToStruct(resp.Result, device); err != nil {
		return nil, errors.NewDeviceError("Failed to parse device info", err)
	}

	return device, nil
}

// GetEnergyUsage retrieves current energy usage
func (c *TapoClient) GetEnergyUsage() (*EnergyUsage, error) {
	if c.token == "" {
		return nil, errors.NewConnectionError("Not authenticated with Tapo device", nil)
	}

	req := LoginRequest{
		Method: "get_energy_usage",
		Params: map[string]interface{}{},
	}

	resp, err := c.makeAuthenticatedRequest(req)
	if err != nil {
		return nil, errors.NewDeviceError("Failed to get energy usage", err)
	}

	if resp.ErrorCode != 0 {
		return nil, errors.NewDeviceError(fmt.Sprintf("Get energy usage failed with error code: %d", resp.ErrorCode), nil)
	}

	// Parse energy usage
	usage := &EnergyUsage{}
	if err := mapToStruct(resp.Result, usage); err != nil {
		return nil, errors.NewDeviceError("Failed to parse energy usage", err)
	}

	return usage, nil
}

// SetDeviceOn turns the device on or off
func (c *TapoClient) SetDeviceOn(on bool) error {
	if c.token == "" {
		return errors.NewConnectionError("Not authenticated with Tapo device", nil)
	}

	req := LoginRequest{
		Method: "set_device_info",
		Params: map[string]interface{}{
			"device_on": on,
		},
	}

	resp, err := c.makeAuthenticatedRequest(req)
	if err != nil {
		return errors.NewDeviceError("Failed to set device state", err)
	}

	if resp.ErrorCode != 0 {
		return errors.NewDeviceError(fmt.Sprintf("Set device state failed with error code: %d", resp.ErrorCode), nil)
	}

	c.logger.Info("Device state changed", map[string]interface{}{
		"device_ip": c.baseURL,
		"state":     on,
	})

	return nil
}

// makeRequest makes an HTTP request to the Tapo device
func (c *TapoClient) makeRequest(url string, payload interface{}) (*TapoResponse, error) {
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var tapoResp TapoResponse
	if err := json.Unmarshal(body, &tapoResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &tapoResp, nil
}

// makeAuthenticatedRequest makes an authenticated request with token
func (c *TapoClient) makeAuthenticatedRequest(payload LoginRequest) (*TapoResponse, error) {
	url := fmt.Sprintf("%s/app?token=%s", c.baseURL, c.token)
	return c.makeRequest(url, payload)
}

// generatePublicKey generates a public key for handshake (simplified implementation)
func generatePublicKey() string {
	// Generate RSA key pair
	privateKey, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		return ""
	}

	// Get public key
	publicKey := &privateKey.PublicKey

	// Convert to PEM format
	pubKeyBytes, err := x509.MarshalPKIXPublicKey(publicKey)
	if err != nil {
		return ""
	}

	pubKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: pubKeyBytes,
	})

	return base64.StdEncoding.EncodeToString(pubKeyPEM)
}

// mapToStruct converts a map to a struct using JSON marshaling/unmarshaling
func mapToStruct(m map[string]interface{}, s interface{}) error {
	jsonData, err := json.Marshal(m)
	if err != nil {
		return err
	}
	return json.Unmarshal(jsonData, s)
}

package tapo

import (
	"bytes"
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/johnpr01/home-automation/internal/errors"
	"github.com/johnpr01/home-automation/internal/logger"
)

// KlapClient implements the KLAP protocol for Tapo smart plugs
type KlapClient struct {
	baseURL    string
	username   string
	password   string
	timeout    time.Duration
	httpClient *http.Client
	logger     logger.Logger

	// KLAP session state
	localSeed  []byte
	remoteSeed []byte
	authHash   []byte
	sessionKey []byte
	iv         []byte
	seq        int32
	cookies    []*http.Cookie
}

// NewKlapClient creates a new KLAP protocol client for Tapo devices
func NewKlapClient(host, username, password string, timeout time.Duration, logger logger.Logger) *KlapClient {
	return &KlapClient{
		baseURL:  fmt.Sprintf("http://%s", host),
		username: username,
		password: password,
		timeout:  timeout,
		httpClient: &http.Client{
			Timeout: timeout,
		},
		logger: logger,
	}
}

// DeviceInfo represents device information response
type KlapDeviceInfo struct {
	DeviceID           string `json:"device_id"`
	FwVersion          string `json:"fw_ver"`
	HwVersion          string `json:"hw_ver"`
	Type               string `json:"type"`
	Model              string `json:"model"`
	MAC                string `json:"mac"`
	HWID               string `json:"hw_id"`
	FWID               string `json:"fw_id"`
	OEMID              string `json:"oem_id"`
	Specs              string `json:"specs"`
	DeviceOn           bool   `json:"device_on"`
	OnTime             int64  `json:"on_time"`
	Overheated         bool   `json:"overheated"`
	Nickname           string `json:"nickname"`
	Location           string `json:"location"`
	Avatar             string `json:"avatar"`
	Longitude          int64  `json:"longitude"`
	Latitude           int64  `json:"latitude"`
	HasSetLocationInfo bool   `json:"has_set_location_info"`
	IP                 string `json:"ip"`
	SSID               string `json:"ssid"`
	RSSI               int    `json:"rssi"`
	SignalLevel        int    `json:"signal_level"`
	AutoOffStatus      string `json:"auto_off_status"`
	AutoOffRemainTime  int    `json:"auto_off_remain_time"`
}

// EnergyUsage represents energy usage data
type KlapEnergyUsage struct {
	TodayRuntime      int    `json:"today_runtime"`
	MonthRuntime      int    `json:"month_runtime"`
	TodayEnergy       int    `json:"today_energy"`
	MonthEnergy       int    `json:"month_energy"`
	LocalTime         string `json:"local_time"`
	ElectricityCharge []int  `json:"electricity_charge"`
	CurrentPower      int    `json:"current_power"`
}

// TapoRequest represents a request to the device
type TapoRequest struct {
	Method string      `json:"method"`
	Params interface{} `json:"params,omitempty"`
}

// TapoResponse represents a response from the device
type KlapTapoResponse struct {
	ErrorCode int         `json:"error_code"`
	Result    interface{} `json:"result,omitempty"`
}

// Connect establishes a KLAP session with the device
func (c *KlapClient) Connect(ctx context.Context) error {
	// Generate local seed
	c.localSeed = make([]byte, 16)
	if _, err := rand.Read(c.localSeed); err != nil {
		return errors.NewDeviceError("failed to generate local seed", err)
	}

	// Calculate auth hash
	c.authHash = c.calcAuthHash(c.username, c.password)

	// Perform handshake 1
	remoteSeed, cookies, err := c.handshake1(ctx)
	if err != nil {
		return errors.NewDeviceError("handshake 1 failed", err)
	}
	c.remoteSeed = remoteSeed
	c.cookies = cookies

	// Perform handshake 2
	if err := c.handshake2(ctx); err != nil {
		return errors.NewDeviceError("handshake 2 failed", err)
	}

	// Generate session keys
	c.generateSessionKeys()

	c.logger.Info("KLAP session established successfully")
	return nil
}

// GetDeviceInfo retrieves device information
func (c *KlapClient) GetDeviceInfo(ctx context.Context) (*KlapDeviceInfo, error) {
	request := TapoRequest{
		Method: "get_device_info",
	}

	var response struct {
		ErrorCode int            `json:"error_code"`
		Result    KlapDeviceInfo `json:"result"`
	}

	if err := c.secureRequest(ctx, request, &response); err != nil {
		return nil, err
	}

	if response.ErrorCode != 0 {
		return nil, errors.NewDeviceError(fmt.Sprintf("device returned error code: %d", response.ErrorCode), nil)
	}

	return &response.Result, nil
}

// GetEnergyUsage retrieves energy usage data
func (c *KlapClient) GetEnergyUsage(ctx context.Context) (*KlapEnergyUsage, error) {
	request := TapoRequest{
		Method: "get_energy_usage",
	}

	var response struct {
		ErrorCode int             `json:"error_code"`
		Result    KlapEnergyUsage `json:"result"`
	}

	if err := c.secureRequest(ctx, request, &response); err != nil {
		return nil, err
	}

	if response.ErrorCode != 0 {
		return nil, errors.NewDeviceError(fmt.Sprintf("device returned error code: %d", response.ErrorCode), nil)
	}

	return &response.Result, nil
}

// handshake1 performs the first KLAP handshake
func (c *KlapClient) handshake1(ctx context.Context) ([]byte, []*http.Cookie, error) {
	url := c.baseURL + "/app/handshake1"

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(c.localSeed))
	if err != nil {
		return nil, nil, err
	}

	req.Header.Set("Content-Type", "application/octet-stream")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, nil, fmt.Errorf("handshake1 failed with status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, err
	}

	if len(body) < 32 {
		return nil, nil, fmt.Errorf("invalid handshake1 response length: %d", len(body))
	}

	remoteSeed := body[:16]
	serverHash := body[16:32]

	// Verify server hash
	localHash := sha256Hash(concat(c.localSeed, remoteSeed, c.authHash))
	if !bytes.Equal(localHash, serverHash) {
		return nil, nil, fmt.Errorf("server hash verification failed")
	}

	return remoteSeed, resp.Cookies(), nil
}

// handshake2 performs the second KLAP handshake
func (c *KlapClient) handshake2(ctx context.Context) error {
	url := c.baseURL + "/app/handshake2"

	payload := sha256Hash(concat(c.remoteSeed, c.localSeed, c.authHash))

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(payload))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Add cookies from handshake1
	for _, cookie := range c.cookies {
		req.AddCookie(cookie)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("handshake2 failed with status %d", resp.StatusCode)
	}

	return nil
}

// secureRequest sends an encrypted request using the KLAP protocol
func (c *KlapClient) secureRequest(ctx context.Context, request TapoRequest, response interface{}) error {
	// Serialize request
	payload, err := json.Marshal(request)
	if err != nil {
		return err
	}

	// Encrypt payload
	encryptedPayload, err := c.encrypt(payload)
	if err != nil {
		return err
	}

	// Prepare request
	url := fmt.Sprintf("%s/app/request?seq=%d", c.baseURL, c.seq)

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(encryptedPayload))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/octet-stream")

	// Add cookies
	for _, cookie := range c.cookies {
		req.AddCookie(cookie)
	}

	// Send request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("secure request failed with status %d", resp.StatusCode)
	}

	// Read and decrypt response
	encryptedResponse, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	decryptedResponse, err := c.decrypt(encryptedResponse)
	if err != nil {
		return err
	}

	// Increment sequence number
	c.seq++

	// Parse response
	return json.Unmarshal(decryptedResponse, response)
}

// generateSessionKeys generates the session encryption keys
func (c *KlapClient) generateSessionKeys() {
	// Generate session key
	c.sessionKey = sha256Hash(concat([]byte("lsk"), c.localSeed, c.remoteSeed, c.authHash))[:16]

	// Generate IV
	c.iv = sha256Hash(concat([]byte("iv"), c.localSeed, c.remoteSeed, c.authHash))[:12]

	// Initialize sequence number
	c.seq = 1
}

// encrypt encrypts data using AES-GCM
func (c *KlapClient) encrypt(data []byte) ([]byte, error) {
	block, err := aes.NewCipher(c.sessionKey)
	if err != nil {
		return nil, err
	}

	// Create IV with sequence number
	iv := make([]byte, 12)
	copy(iv, c.iv)
	binary.BigEndian.PutUint32(iv[8:], uint32(c.seq))

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	return aesGCM.Seal(nil, iv, data, nil), nil
}

// decrypt decrypts data using AES-GCM
func (c *KlapClient) decrypt(data []byte) ([]byte, error) {
	block, err := aes.NewCipher(c.sessionKey)
	if err != nil {
		return nil, err
	}

	// Create IV with sequence number
	iv := make([]byte, 12)
	copy(iv, c.iv)
	binary.BigEndian.PutUint32(iv[8:], uint32(c.seq))

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	return aesGCM.Open(nil, iv, data, nil)
}

// calcAuthHash calculates the authentication hash
func (c *KlapClient) calcAuthHash(username, password string) []byte {
	usernameSha1 := sha1Hash([]byte(username))
	passwordSha1 := sha1Hash([]byte(password))
	return sha256Hash(concat(usernameSha1, passwordSha1))
}

// Helper functions
func sha256Hash(data []byte) []byte {
	hash := sha256.Sum256(data)
	return hash[:]
}

func sha1Hash(data []byte) []byte {
	hash := sha1.Sum(data)
	return hash[:]
}

func concat(arrays ...[]byte) []byte {
	var result []byte
	for _, array := range arrays {
		result = append(result, array...)
	}
	return result
}

package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
)

// Config represents CLI configuration
type Config struct {
	APIBaseURL   string `json:"api_base_url"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	UserID       string `json:"user_id"`
	Email        string `json:"email"`
	Username     string `json:"username"`
}

// Client is the API client for Aureo VPN
type Client struct {
	baseURL    string
	httpClient *http.Client
	config     *Config
	configPath string
}

// User represents a user from the API
type User struct {
	ID               uuid.UUID `json:"id"`
	Email            string    `json:"email"`
	Username         string    `json:"username"`
	FullName         string    `json:"full_name"`
	IsActive         bool      `json:"is_active"`
	SubscriptionTier string    `json:"subscription_tier"`
	DataTransferredGB float64  `json:"data_transferred_gb"`
	ConnectionCount  int       `json:"connection_count"`
	CreatedAt        time.Time `json:"created_at"`
}

// AuthResponse is returned from login/register
type AuthResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	User         User   `json:"user"`
}

// VPNNode represents a VPN node
type VPNNode struct {
	ID                 uuid.UUID `json:"id"`
	Name               string    `json:"name"`
	Hostname           string    `json:"hostname"`
	Country            string    `json:"country"`
	CountryCode        string    `json:"country_code"`
	City               string    `json:"city"`
	PublicIP           string    `json:"public_ip"`
	Status             string    `json:"status"`
	CurrentConnections int       `json:"current_connections"`
	MaxConnections     int       `json:"max_connections"`
	LoadScore          float64   `json:"load_score"`
	Latency            int       `json:"latency_ms"`
	SupportsWireGuard  bool      `json:"supports_wireguard"`
	SupportsOpenVPN    bool      `json:"supports_openvpn"`
	WireGuardPort      int       `json:"wireguard_port"`
	OpenVPNPort        int       `json:"openvpn_port"`
	UptimePercentage   float64   `json:"uptime_percentage"`
	PublicKey          string    `json:"public_key"`
}

// WireGuardConfigResponse is returned from /config/generate
type WireGuardConfigResponse struct {
	SessionID       string `json:"session_id"`
	ServerPublicKey string `json:"server_public_key"`
	ServerEndpoint  string `json:"server_endpoint"`
	ClientIP        string `json:"client_ip"`
	DNS             string `json:"dns"`
	AllowedIPs      string `json:"allowed_ips"`
	Keepalive       int    `json:"keepalive"`
}

// ConnectionState represents the current VPN connection state
type ConnectionState struct {
	NodeID      string    `json:"node_id"`
	NodeName    string    `json:"node_name"`
	NodeIP      string    `json:"node_ip"`
	ClientIP    string    `json:"client_ip"`
	ConnectedAt time.Time `json:"connected_at"`
	SessionID   string    `json:"session_id"`
}

// NodesResponse is the response from listing nodes
type NodesResponse struct {
	Nodes  []VPNNode `json:"nodes"`
	Count  int       `json:"count"`
	Source string    `json:"source"`
}

// Session represents a VPN session
type Session struct {
	ID            uuid.UUID  `json:"id"`
	UserID        uuid.UUID  `json:"user_id"`
	NodeID        uuid.UUID  `json:"node_id"`
	Protocol      string     `json:"protocol"`
	ClientIP      string     `json:"client_ip"`
	TunnelIP      string     `json:"tunnel_ip"`
	Status        string     `json:"status"`
	ConnectedAt   time.Time  `json:"connected_at"`
	DisconnectedAt *time.Time `json:"disconnected_at,omitempty"`
	BytesSent     int64      `json:"bytes_sent"`
	BytesReceived int64      `json:"bytes_received"`
	DataUsedGB    float64    `json:"data_used_gb"`
	Latency       int        `json:"latency_ms"`
	PacketLoss    float64    `json:"packet_loss"`
	Node          *VPNNode   `json:"node,omitempty"`
}

// SessionsResponse is the response from listing sessions
type SessionsResponse struct {
	Sessions []Session `json:"sessions"`
	Count    int       `json:"count"`
}

// UserStats represents user statistics
type UserStats struct {
	TotalSessions      int     `json:"total_sessions"`
	ActiveSessions     int     `json:"active_sessions"`
	TotalDataGB        float64 `json:"total_data_gb"`
	TotalConnectionTime int64  `json:"total_connection_time_seconds"`
	AverageLatency     int     `json:"average_latency_ms"`
	FavoriteCountry    string  `json:"favorite_country"`
	LastConnected      *time.Time `json:"last_connected,omitempty"`
}

// VPNConfig represents a WireGuard/OpenVPN config
type VPNConfig struct {
	ID            uuid.UUID `json:"id"`
	UserID        uuid.UUID `json:"user_id"`
	NodeID        uuid.UUID `json:"node_id"`
	Protocol      string    `json:"protocol"`
	ConfigName    string    `json:"config_name"`
	ConfigContent string    `json:"config_content"`
	PublicKey     string    `json:"public_key"`
	IsActive      bool      `json:"is_active"`
	CreatedAt     time.Time `json:"created_at"`
	ExpiresAt     *time.Time `json:"expires_at,omitempty"`
	Node          *VPNNode  `json:"node,omitempty"`
}

// ConfigsResponse is the response from listing configs
type ConfigsResponse struct {
	Configs []VPNConfig `json:"configs"`
	Count   int         `json:"count"`
}

// APIError represents an API error response
type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// ErrorResponse is the standard error response
type ErrorResponse struct {
	Error     APIError `json:"error"`
	RequestID string   `json:"request_id"`
}

// NewClient creates a new API client
func NewClient(baseURL string) (*Client, error) {
	configPath, err := getConfigPath()
	if err != nil {
		return nil, err
	}

	client := &Client{
		baseURL:    baseURL,
		httpClient: &http.Client{Timeout: 30 * time.Second},
		configPath: configPath,
	}

	// Load existing config if available
	client.loadConfig()

	return client, nil
}

// getConfigPath returns the path to the config file
func getConfigPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	configDir := filepath.Join(homeDir, ".aureo")
	if err := os.MkdirAll(configDir, 0700); err != nil {
		return "", err
	}
	return filepath.Join(configDir, "config.json"), nil
}

// loadConfig loads the configuration from disk
func (c *Client) loadConfig() {
	data, err := os.ReadFile(c.configPath)
	if err != nil {
		c.config = &Config{APIBaseURL: c.baseURL}
		return
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		c.config = &Config{APIBaseURL: c.baseURL}
		return
	}

	c.config = &config
	if c.config.APIBaseURL == "" {
		c.config.APIBaseURL = c.baseURL
	}
}

// saveConfig saves the configuration to disk
func (c *Client) saveConfig() error {
	data, err := json.MarshalIndent(c.config, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(c.configPath, data, 0600)
}

// IsAuthenticated checks if the user is authenticated
func (c *Client) IsAuthenticated() bool {
	return c.config != nil && c.config.AccessToken != ""
}

// GetCurrentUser returns the current logged-in user info
func (c *Client) GetCurrentUser() (string, string) {
	if c.config == nil {
		return "", ""
	}
	return c.config.Email, c.config.Username
}

// doRequest performs an HTTP request with authentication
func (c *Client) doRequest(method, path string, body interface{}, result interface{}) error {
	var reqBody io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("failed to marshal request: %w", err)
		}
		reqBody = bytes.NewBuffer(jsonBody)
	}

	req, err := http.NewRequest(method, c.baseURL+path, reqBody)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	if c.config != nil && c.config.AccessToken != "" {
		req.Header.Set("Authorization", "Bearer "+c.config.AccessToken)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		var errResp ErrorResponse
		if err := json.Unmarshal(respBody, &errResp); err == nil && errResp.Error.Message != "" {
			return fmt.Errorf("%s", errResp.Error.Message)
		}
		return fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	if result != nil {
		if err := json.Unmarshal(respBody, result); err != nil {
			return fmt.Errorf("failed to parse response: %w", err)
		}
	}

	return nil
}

// Login authenticates the user
func (c *Client) Login(email, password string) (*AuthResponse, error) {
	req := map[string]string{
		"email":    email,
		"password": password,
	}

	var resp AuthResponse
	if err := c.doRequest("POST", "/api/v1/auth/login", req, &resp); err != nil {
		return nil, err
	}

	c.config.AccessToken = resp.AccessToken
	c.config.RefreshToken = resp.RefreshToken
	c.config.UserID = resp.User.ID.String()
	c.config.Email = resp.User.Email
	c.config.Username = resp.User.Username

	if err := c.saveConfig(); err != nil {
		return nil, fmt.Errorf("failed to save config: %w", err)
	}

	return &resp, nil
}

// Register creates a new user account
func (c *Client) Register(email, password, username, fullName string) (*AuthResponse, error) {
	req := map[string]string{
		"email":     email,
		"password":  password,
		"username":  username,
		"full_name": fullName,
	}

	var resp AuthResponse
	if err := c.doRequest("POST", "/api/v1/auth/register", req, &resp); err != nil {
		return nil, err
	}

	c.config.AccessToken = resp.AccessToken
	c.config.RefreshToken = resp.RefreshToken
	c.config.UserID = resp.User.ID.String()
	c.config.Email = resp.User.Email
	c.config.Username = resp.User.Username

	if err := c.saveConfig(); err != nil {
		return nil, fmt.Errorf("failed to save config: %w", err)
	}

	return &resp, nil
}

// Logout clears the authentication tokens
func (c *Client) Logout() error {
	c.config.AccessToken = ""
	c.config.RefreshToken = ""
	c.config.UserID = ""
	c.config.Email = ""
	c.config.Username = ""
	return c.saveConfig()
}

// RefreshToken refreshes the access token
func (c *Client) RefreshToken() error {
	if c.config.RefreshToken == "" {
		return fmt.Errorf("no refresh token available")
	}

	req := map[string]string{
		"refresh_token": c.config.RefreshToken,
	}

	var resp struct {
		AccessToken string `json:"access_token"`
	}

	if err := c.doRequest("POST", "/api/v1/auth/refresh", req, &resp); err != nil {
		return err
	}

	c.config.AccessToken = resp.AccessToken
	return c.saveConfig()
}

// GetProfile returns the current user's profile
func (c *Client) GetProfile() (*User, error) {
	var user User
	if err := c.doRequest("GET", "/api/v1/user/profile", nil, &user); err != nil {
		return nil, err
	}
	return &user, nil
}

// GetUserStats returns the current user's statistics
func (c *Client) GetUserStats() (*UserStats, error) {
	var stats UserStats
	if err := c.doRequest("GET", "/api/v1/user/stats", nil, &stats); err != nil {
		return nil, err
	}
	return &stats, nil
}

// GetNodes returns a list of VPN nodes
func (c *Client) GetNodes(country, protocol string) (*NodesResponse, error) {
	path := "/api/v1/nodes"
	params := []string{}
	if country != "" {
		params = append(params, "country="+country)
	}
	if protocol != "" {
		params = append(params, "protocol="+protocol)
	}
	if len(params) > 0 {
		path += "?"
		for i, p := range params {
			if i > 0 {
				path += "&"
			}
			path += p
		}
	}

	var resp NodesResponse
	if err := c.doRequest("GET", path, nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetBestNode returns the optimal VPN node
func (c *Client) GetBestNode(country, protocol string) (*VPNNode, error) {
	path := "/api/v1/nodes/best"
	params := []string{}
	if country != "" {
		params = append(params, "country="+country)
	}
	if protocol != "" {
		params = append(params, "protocol="+protocol)
	}
	if len(params) > 0 {
		path += "?"
		for i, p := range params {
			if i > 0 {
				path += "&"
			}
			path += p
		}
	}

	var node VPNNode
	if err := c.doRequest("GET", path, nil, &node); err != nil {
		return nil, err
	}
	return &node, nil
}

// GetNode returns a specific node by ID
func (c *Client) GetNode(nodeID string) (*VPNNode, error) {
	var node VPNNode
	if err := c.doRequest("GET", "/api/v1/nodes/"+nodeID, nil, &node); err != nil {
		return nil, err
	}
	return &node, nil
}

// CreateSession creates a new VPN session
func (c *Client) CreateSession(nodeID, protocol string) (*Session, error) {
	req := map[string]string{
		"node_id":  nodeID,
		"protocol": protocol,
	}

	var session Session
	if err := c.doRequest("POST", "/api/v1/sessions", req, &session); err != nil {
		return nil, err
	}
	return &session, nil
}

// DisconnectSession terminates a VPN session
func (c *Client) DisconnectSession(sessionID string) error {
	return c.doRequest("DELETE", "/api/v1/sessions/"+sessionID, nil, nil)
}

// GetSession returns a specific session by ID
func (c *Client) GetSession(sessionID string) (*Session, error) {
	var session Session
	if err := c.doRequest("GET", "/api/v1/sessions/"+sessionID, nil, &session); err != nil {
		return nil, err
	}
	return &session, nil
}

// GetActiveSessions returns active sessions for the current user
func (c *Client) GetActiveSessions() (*SessionsResponse, error) {
	var resp SessionsResponse
	if err := c.doRequest("GET", "/api/v1/user/sessions", nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GenerateConfig generates a VPN configuration
func (c *Client) GenerateConfig(nodeID, protocol, configName string) (*VPNConfig, error) {
	req := map[string]string{
		"node_id":     nodeID,
		"protocol":    protocol,
		"config_name": configName,
	}

	var config VPNConfig
	if err := c.doRequest("POST", "/api/v1/config/generate", req, &config); err != nil {
		return nil, err
	}
	return &config, nil
}

// GetConfigs returns all VPN configurations for the user
func (c *Client) GetConfigs() (*ConfigsResponse, error) {
	var resp ConfigsResponse
	if err := c.doRequest("GET", "/api/v1/config", nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetConfig returns a specific configuration by ID
func (c *Client) GetConfig(configID string) (*VPNConfig, error) {
	var config VPNConfig
	if err := c.doRequest("GET", "/api/v1/config/"+configID, nil, &config); err != nil {
		return nil, err
	}
	return &config, nil
}

// GenerateWireGuardConfig registers with the VPN server and gets WireGuard config
func (c *Client) GenerateWireGuardConfig(nodeID, publicKey string) (*WireGuardConfigResponse, error) {
	req := map[string]string{
		"node_id":    nodeID,
		"public_key": publicKey,
	}

	var resp WireGuardConfigResponse
	if err := c.doRequest("POST", "/api/v1/config/generate", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetConfigDir returns the configuration directory path
func (c *Client) GetConfigDir() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "/tmp/.aureo"
	}
	return filepath.Join(homeDir, ".aureo")
}

// GetConnectionFilePath returns the path to the connection state file
func (c *Client) GetConnectionFilePath() string {
	return filepath.Join(c.GetConfigDir(), "connection.json")
}

// GetWireGuardConfigPath returns the path to the WireGuard config file
func (c *Client) GetWireGuardConfigPath() string {
	return filepath.Join(c.GetConfigDir(), "wg0.conf")
}

// SaveConnectionState saves the current connection state
func (c *Client) SaveConnectionState(state *ConnectionState) error {
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(c.GetConnectionFilePath(), data, 0600)
}

// LoadConnectionState loads the current connection state
func (c *Client) LoadConnectionState() (*ConnectionState, error) {
	data, err := os.ReadFile(c.GetConnectionFilePath())
	if err != nil {
		return nil, err
	}

	var state ConnectionState
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, err
	}
	return &state, nil
}

// ClearConnectionState removes the connection state file
func (c *Client) ClearConnectionState() error {
	connFile := c.GetConnectionFilePath()
	if _, err := os.Stat(connFile); err == nil {
		return os.Remove(connFile)
	}
	return nil
}

// IsConnected checks if there's an active VPN connection
func (c *Client) IsConnected() bool {
	_, err := c.LoadConnectionState()
	return err == nil
}

// GetAPIBaseURL returns the API base URL
func (c *Client) GetAPIBaseURL() string {
	if c.config != nil && c.config.APIBaseURL != "" {
		return c.config.APIBaseURL
	}
	return c.baseURL
}

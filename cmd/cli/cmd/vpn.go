package cmd

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/nikola43/aureo-vpn/cmd/cli/client"
	"github.com/spf13/cobra"
)

var (
	connectNodeID   string
	connectProtocol string
	connectCountry  string
	connectBest     bool
)

// connectCmd connects to a VPN node
var connectCmd = &cobra.Command{
	Use:   "connect",
	Short: "Connect to a VPN node",
	Long: `Connect to a VPN node using WireGuard.

Examples:
  aureo connect                              # Interactive node selection
  aureo connect --node <node-id>             # Connect to a specific node
  aureo connect --country US                 # Connect to a node in the US
  aureo connect --best                       # Connect to the best available node`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if !APIClient.IsAuthenticated() {
			return fmt.Errorf("not logged in. Use 'aureo login' first")
		}

		// Check if already connected
		if APIClient.IsConnected() {
			state, _ := APIClient.LoadConnectionState()
			fmt.Printf("Already connected to %s\n", state.NodeName)
			fmt.Println("Use 'aureo disconnect' first")
			return nil
		}

		// Check for WireGuard tools
		if !checkWireGuardInstalled() {
			return fmt.Errorf("WireGuard not installed. Install with: brew install wireguard-tools")
		}

		var selectedNode *client.VPNNode
		var err error

		if connectNodeID != "" {
			// Specific node requested
			selectedNode, err = APIClient.GetNode(connectNodeID)
			if err != nil {
				return fmt.Errorf("failed to get node: %w", err)
			}
		} else if connectBest || connectCountry != "" {
			// Best node requested
			fmt.Println("Finding best available node...")
			selectedNode, err = APIClient.GetBestNode(connectCountry, "wireguard")
			if err != nil {
				return fmt.Errorf("failed to find best node: %w", err)
			}
			fmt.Printf("Selected: %s (%s, %s)\n", selectedNode.Name, selectedNode.City, selectedNode.Country)
		} else {
			// Interactive selection
			selectedNode, err = selectNodeInteractively()
			if err != nil {
				return err
			}
		}

		return connectToNode(selectedNode)
	},
}

// selectNodeInteractively shows a list of nodes for user to select
func selectNodeInteractively() (*client.VPNNode, error) {
	fmt.Println("Fetching available nodes...")
	fmt.Println()

	resp, err := APIClient.GetNodes("", "")
	if err != nil {
		return nil, fmt.Errorf("failed to fetch nodes: %w", err)
	}

	if len(resp.Nodes) == 0 {
		return nil, fmt.Errorf("no nodes available")
	}

	fmt.Println("Available Nodes:")
	fmt.Println()

	for i, node := range resp.Nodes {
		statusColor := "[OK]"
		if node.Status != "online" {
			statusColor = "[!]"
		}

		location := node.City
		if location != "" && node.Country != "" {
			location += ", " + node.Country
		} else if node.Country != "" {
			location = node.Country
		}

		fmt.Printf("  [%d] %s %s\n", i+1, statusColor, node.Name)
		fmt.Printf("      Location: %s\n", location)
		fmt.Printf("      IP: %s\n", node.PublicIP)
		fmt.Println()
	}

	// Get user selection
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("Select node [1-%d]: ", len(resp.Nodes))
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)

	selection, err := strconv.Atoi(input)
	if err != nil || selection < 1 || selection > len(resp.Nodes) {
		return nil, fmt.Errorf("invalid selection")
	}

	return &resp.Nodes[selection-1], nil
}

// connectToNode handles the actual connection process
func connectToNode(node *client.VPNNode) error {
	fmt.Println()
	fmt.Printf("Connecting to %s...\n", node.Name)
	fmt.Println()

	// Generate WireGuard keys
	fmt.Println("Generating WireGuard keys...")
	privateKey, publicKey, err := generateWireGuardKeys()
	if err != nil {
		return fmt.Errorf("failed to generate keys: %w", err)
	}

	// Register with VPN server
	fmt.Println("Registering with VPN server...")
	config, err := APIClient.GenerateWireGuardConfig(node.ID.String(), publicKey)
	if err != nil {
		return fmt.Errorf("failed to register: %w", err)
	}

	if config.ServerPublicKey == "" {
		return fmt.Errorf("server did not return public key")
	}

	fmt.Printf("  Your VPN IP: %s\n", config.ClientIP)
	fmt.Printf("  Server: %s\n", config.ServerEndpoint)
	fmt.Println()

	// Create WireGuard config file
	fmt.Println("Creating WireGuard configuration...")
	wgConfigPath := APIClient.GetWireGuardConfigPath()

	dns := config.DNS
	if dns == "" {
		dns = "1.1.1.1,8.8.8.8"
	}

	allowedIPs := config.AllowedIPs
	if allowedIPs == "" {
		allowedIPs = "0.0.0.0/0"
	}

	keepalive := config.Keepalive
	if keepalive == 0 {
		keepalive = 25
	}

	wgConfig := fmt.Sprintf(`[Interface]
PrivateKey = %s
Address = %s/32
DNS = %s

[Peer]
PublicKey = %s
Endpoint = %s
AllowedIPs = %s
PersistentKeepalive = %d
`, privateKey, config.ClientIP, dns, config.ServerPublicKey, config.ServerEndpoint, allowedIPs, keepalive)

	if err := os.WriteFile(wgConfigPath, []byte(wgConfig), 0600); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	// Connect via WireGuard
	fmt.Println("Starting VPN tunnel...")
	if err := runWgQuickUp(wgConfigPath); err != nil {
		// Clean up config file on failure
		os.Remove(wgConfigPath)
		return fmt.Errorf("failed to connect: %w", err)
	}

	// Save connection state
	state := &client.ConnectionState{
		NodeID:      node.ID.String(),
		NodeName:    node.Name,
		NodeIP:      node.PublicIP,
		ClientIP:    config.ClientIP,
		ConnectedAt: time.Now(),
		SessionID:   config.SessionID,
	}
	if err := APIClient.SaveConnectionState(state); err != nil {
		fmt.Printf("Warning: failed to save connection state: %v\n", err)
	}

	fmt.Println()
	fmt.Println("========================================")
	fmt.Printf("  Connected to %s\n", node.Name)
	fmt.Println("========================================")
	fmt.Println()
	fmt.Printf("  Your IP: %s\n", node.PublicIP)
	if node.City != "" {
		fmt.Printf("  Location: %s, %s\n", node.City, node.Country)
	} else {
		fmt.Printf("  Location: %s\n", node.Country)
	}
	fmt.Println()
	fmt.Println("VPN is active. To disconnect, run: aureo disconnect")
	fmt.Println()

	return nil
}

var disconnectSessionID string

// disconnectCmd disconnects from a VPN node
var disconnectCmd = &cobra.Command{
	Use:   "disconnect",
	Short: "Disconnect from VPN",
	Long:  `Disconnect from the current VPN connection.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if !APIClient.IsConnected() {
			fmt.Println("Not connected")
			return nil
		}

		state, err := APIClient.LoadConnectionState()
		if err != nil {
			fmt.Println("No active connection found")
			return nil
		}

		fmt.Printf("Disconnecting from %s...\n", state.NodeName)

		// Stop WireGuard
		wgConfigPath := APIClient.GetWireGuardConfigPath()
		if err := runWgQuickDown(wgConfigPath); err != nil {
			fmt.Printf("Warning: %v\n", err)
		}

		// Clean up files
		os.Remove(wgConfigPath)
		APIClient.ClearConnectionState()

		fmt.Println("Disconnected")
		return nil
	},
}

// statusCmd shows VPN connection status
var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show VPN connection status",
	Long:  `Show the current VPN connection status.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		email, username := APIClient.GetCurrentUser()

		fmt.Println("Aureo VPN Status")
		fmt.Println("================")
		fmt.Println()

		// Auth status
		if APIClient.IsAuthenticated() {
			fmt.Printf("User: %s (%s)\n", username, email)
		} else {
			fmt.Println("User: Not logged in")
			fmt.Println()
			fmt.Println("Use 'aureo login' to authenticate")
			return nil
		}

		fmt.Println()

		// Connection status
		if !APIClient.IsConnected() {
			fmt.Println("Status: Disconnected")
			fmt.Println()
			fmt.Println("Use 'aureo connect' to establish a connection")
			return nil
		}

		state, err := APIClient.LoadConnectionState()
		if err != nil {
			fmt.Println("Status: Unknown")
			return nil
		}

		duration := time.Since(state.ConnectedAt)

		fmt.Println("Status: Connected")
		fmt.Println()
		fmt.Printf("  Node: %s\n", state.NodeName)
		fmt.Printf("  Server IP: %s\n", state.NodeIP)
		fmt.Printf("  Your VPN IP: %s\n", state.ClientIP)
		fmt.Printf("  Connected: %s ago\n", formatDuration(duration))

		// Try to get WireGuard stats
		if stats := getWireGuardStats(); stats != "" {
			fmt.Println()
			fmt.Println("Traffic:")
			fmt.Println(stats)
		}

		return nil
	},
}

// sessionsCmd lists all sessions (including history)
var sessionsCmd = &cobra.Command{
	Use:   "sessions",
	Short: "List VPN sessions",
	Long:  `List VPN session information.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if !APIClient.IsAuthenticated() {
			return fmt.Errorf("not logged in. Use 'aureo login' first")
		}

		// Check current connection
		if APIClient.IsConnected() {
			state, _ := APIClient.LoadConnectionState()
			duration := time.Since(state.ConnectedAt)

			fmt.Println("Current Session")
			fmt.Println("===============")
			fmt.Println()

			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "NODE\tSERVER IP\tVPN IP\tDURATION")
			fmt.Fprintln(w, "----\t---------\t------\t--------")
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
				state.NodeName,
				state.NodeIP,
				state.ClientIP,
				formatDuration(duration),
			)
			w.Flush()
		} else {
			fmt.Println("No active session")
			fmt.Println()
			fmt.Println("Use 'aureo connect' to establish a connection")
		}

		return nil
	},
}

func init() {
	// Connect flags
	connectCmd.Flags().StringVarP(&connectNodeID, "node", "n", "", "Node ID to connect to")
	connectCmd.Flags().StringVarP(&connectProtocol, "protocol", "p", "", "Protocol to use (wireguard)")
	connectCmd.Flags().StringVarP(&connectCountry, "country", "c", "", "Country to connect to")
	connectCmd.Flags().BoolVarP(&connectBest, "best", "b", false, "Connect to the best available node")

	// Disconnect flags
	disconnectCmd.Flags().StringVarP(&disconnectSessionID, "session", "s", "", "Session ID to disconnect")
}

// checkWireGuardInstalled checks if WireGuard tools are installed
func checkWireGuardInstalled() bool {
	_, err := exec.LookPath("wg")
	if err != nil {
		return false
	}
	_, err = exec.LookPath("wg-quick")
	return err == nil
}

// generateWireGuardKeys generates a WireGuard key pair
func generateWireGuardKeys() (privateKey, publicKey string, err error) {
	// Generate private key
	privCmd := exec.Command("wg", "genkey")
	privOut, err := privCmd.Output()
	if err != nil {
		return "", "", fmt.Errorf("failed to generate private key: %w", err)
	}
	privateKey = strings.TrimSpace(string(privOut))

	// Generate public key from private key
	pubCmd := exec.Command("wg", "pubkey")
	pubCmd.Stdin = strings.NewReader(privateKey)
	pubOut, err := pubCmd.Output()
	if err != nil {
		return "", "", fmt.Errorf("failed to generate public key: %w", err)
	}
	publicKey = strings.TrimSpace(string(pubOut))

	return privateKey, publicKey, nil
}

// runWgQuickUp starts the WireGuard tunnel
func runWgQuickUp(configPath string) error {
	cmd := exec.Command("sudo", "wg-quick", "up", configPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// runWgQuickDown stops the WireGuard tunnel
func runWgQuickDown(configPath string) error {
	cmd := exec.Command("sudo", "wg-quick", "down", configPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// getWireGuardStats gets transfer statistics from WireGuard
func getWireGuardStats() string {
	// Try to find the interface
	cmd := exec.Command("sudo", "wg", "show", "all", "transfer")
	out, err := cmd.Output()
	if err != nil {
		return ""
	}

	output := strings.TrimSpace(string(out))
	if output == "" {
		return ""
	}

	// Parse the output (format: interface\trx\ttx)
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		parts := strings.Fields(line)
		if len(parts) >= 3 {
			rx, _ := strconv.ParseInt(parts[1], 10, 64)
			tx, _ := strconv.ParseInt(parts[2], 10, 64)
			return fmt.Sprintf("  Received: %s\n  Sent: %s", formatBytes(rx), formatBytes(tx))
		}
	}

	return ""
}

// formatDuration formats a duration for display
func formatDuration(d time.Duration) string {
	d = d.Round(time.Second)
	h := d / time.Hour
	d -= h * time.Hour
	m := d / time.Minute
	d -= m * time.Minute
	s := d / time.Second

	if h > 0 {
		return fmt.Sprintf("%dh %dm", h, m)
	}
	if m > 0 {
		return fmt.Sprintf("%dm %ds", m, s)
	}
	return fmt.Sprintf("%ds", s)
}

// formatBytes formats bytes for display
func formatBytes(bytes int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)

	switch {
	case bytes >= GB:
		return fmt.Sprintf("%.2f GB", float64(bytes)/GB)
	case bytes >= MB:
		return fmt.Sprintf("%.2f MB", float64(bytes)/MB)
	case bytes >= KB:
		return fmt.Sprintf("%.2f KB", float64(bytes)/KB)
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}

// pluralize adds 's' for plural
func pluralize(n int) string {
	if n == 1 {
		return ""
	}
	return "s"
}

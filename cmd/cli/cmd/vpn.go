package cmd

import (
	"fmt"
	"os"
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
	Long: `Connect to a VPN node using WireGuard or OpenVPN.

Examples:
  aureo connect                              # Connect to the best available node
  aureo connect --node <node-id>             # Connect to a specific node
  aureo connect --country US                 # Connect to a node in the US
  aureo connect --protocol wireguard         # Use WireGuard protocol
  aureo connect --best --country DE          # Connect to the best node in Germany`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if !APIClient.IsAuthenticated() {
			return fmt.Errorf("not logged in. Use 'aureo login' first")
		}

		// Determine which node to connect to
		nodeID := connectNodeID

		if nodeID == "" {
			// Find the best node
			fmt.Println("Finding best available node...")
			node, err := APIClient.GetBestNode(connectCountry, connectProtocol)
			if err != nil {
				return fmt.Errorf("failed to find node: %w", err)
			}
			nodeID = node.ID.String()
			fmt.Printf("Selected: %s (%s, %s)\n", node.Name, node.City, node.CountryCode)
		}

		// Default to WireGuard if not specified
		protocol := connectProtocol
		if protocol == "" {
			protocol = "wireguard"
		}

		fmt.Printf("Connecting via %s...\n", protocol)

		session, err := APIClient.CreateSession(nodeID, protocol)
		if err != nil {
			return fmt.Errorf("failed to connect: %w", err)
		}

		fmt.Println()
		fmt.Println("Connected successfully!")
		fmt.Println("=======================")
		fmt.Printf("Session ID:  %s\n", session.ID)
		fmt.Printf("Protocol:    %s\n", session.Protocol)
		fmt.Printf("Tunnel IP:   %s\n", session.TunnelIP)
		fmt.Printf("Connected:   %s\n", session.ConnectedAt.Format(time.RFC3339))

		if session.Node != nil {
			fmt.Printf("Node:        %s (%s)\n", session.Node.Name, session.Node.CountryCode)
		}

		fmt.Println()
		fmt.Println("Your internet traffic is now encrypted and routed through the VPN.")
		fmt.Println("Use 'aureo status' to check connection status.")
		fmt.Println("Use 'aureo disconnect' to disconnect.")

		return nil
	},
}

var disconnectSessionID string

// disconnectCmd disconnects from a VPN node
var disconnectCmd = &cobra.Command{
	Use:   "disconnect",
	Short: "Disconnect from VPN",
	Long: `Disconnect from the current VPN session or a specific session.

Examples:
  aureo disconnect                    # Disconnect all active sessions
  aureo disconnect --session <id>     # Disconnect a specific session`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if !APIClient.IsAuthenticated() {
			return fmt.Errorf("not logged in")
		}

		if disconnectSessionID != "" {
			// Disconnect specific session
			if err := APIClient.DisconnectSession(disconnectSessionID); err != nil {
				return fmt.Errorf("failed to disconnect: %w", err)
			}
			fmt.Printf("Disconnected session %s\n", truncateID(disconnectSessionID))
			return nil
		}

		// Disconnect all active sessions
		resp, err := APIClient.GetActiveSessions()
		if err != nil {
			return fmt.Errorf("failed to get sessions: %w", err)
		}

		activeSessions := 0
		for _, session := range resp.Sessions {
			if session.Status == "active" {
				activeSessions++
			}
		}

		if activeSessions == 0 {
			fmt.Println("No active VPN sessions")
			return nil
		}

		fmt.Printf("Disconnecting %d active session(s)...\n", activeSessions)

		disconnected := 0
		for _, session := range resp.Sessions {
			if session.Status == "active" {
				if err := APIClient.DisconnectSession(session.ID.String()); err != nil {
					fmt.Printf("Failed to disconnect session %s: %v\n", truncateID(session.ID.String()), err)
				} else {
					disconnected++
				}
			}
		}

		fmt.Printf("\nDisconnected %d session(s)\n", disconnected)
		fmt.Println("Your internet traffic is no longer routed through the VPN.")

		return nil
	},
}

// statusCmd shows VPN connection status
var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show VPN connection status",
	Long: `Show the current VPN connection status and active sessions.

Examples:
  aureo status                        # Show all active sessions
  aureo status --session <id>         # Show details of a specific session`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if !APIClient.IsAuthenticated() {
			fmt.Println("Status: Not logged in")
			fmt.Println("Use 'aureo login' to authenticate")
			return nil
		}

		resp, err := APIClient.GetActiveSessions()
		if err != nil {
			return fmt.Errorf("failed to get sessions: %w", err)
		}

		// Count active sessions
		activeSessions := []client.Session{}
		for _, session := range resp.Sessions {
			if session.Status == "active" {
				activeSessions = append(activeSessions, session)
			}
		}

		email, username := APIClient.GetCurrentUser()
		fmt.Printf("User: %s (%s)\n", username, email)
		fmt.Println()

		if len(activeSessions) == 0 {
			fmt.Println("Status: Disconnected")
			fmt.Println()
			fmt.Println("No active VPN sessions")
			fmt.Println("Use 'aureo connect' to establish a connection")
			return nil
		}

		fmt.Printf("Status: Connected (%d active session%s)\n", len(activeSessions), pluralize(len(activeSessions)))
		fmt.Println()

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "SESSION\tNODE\tPROTOCOL\tTUNNEL IP\tDATA\tDURATION\tLATENCY")
		fmt.Fprintln(w, "-------\t----\t--------\t---------\t----\t--------\t-------")

		for _, session := range activeSessions {
			nodeName := "Unknown"
			if session.Node != nil {
				nodeName = session.Node.Name
			}

			duration := time.Since(session.ConnectedAt)
			durationStr := formatDuration(duration)

			dataStr := formatBytes(session.BytesSent + session.BytesReceived)

			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\t%dms\n",
				truncateID(session.ID.String()),
				nodeName,
				session.Protocol,
				session.TunnelIP,
				dataStr,
				durationStr,
				session.Latency,
			)
		}

		w.Flush()

		return nil
	},
}

// sessionsCmd lists all sessions (including history)
var sessionsCmd = &cobra.Command{
	Use:   "sessions",
	Short: "List VPN sessions",
	Long: `List all VPN sessions including connection history.

Examples:
  aureo sessions                      # List all sessions
  aureo sessions --active             # List only active sessions`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if !APIClient.IsAuthenticated() {
			return fmt.Errorf("not logged in. Use 'aureo login' first")
		}

		resp, err := APIClient.GetActiveSessions()
		if err != nil {
			return fmt.Errorf("failed to get sessions: %w", err)
		}

		if len(resp.Sessions) == 0 {
			fmt.Println("No sessions found")
			return nil
		}

		fmt.Printf("VPN Sessions (%d total)\n\n", resp.Count)

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "SESSION\tNODE\tSTATUS\tPROTOCOL\tDATA\tCONNECTED\tDURATION")
		fmt.Fprintln(w, "-------\t----\t------\t--------\t----\t---------\t--------")

		for _, session := range resp.Sessions {
			nodeName := "Unknown"
			if session.Node != nil {
				nodeName = session.Node.Name
			}

			statusIcon := "[OK]"
			if session.Status != "active" {
				statusIcon = "[X]"
			}

			dataStr := formatBytes(session.BytesSent + session.BytesReceived)
			connectedAt := session.ConnectedAt.Format("Jan 02 15:04")

			var durationStr string
			if session.DisconnectedAt != nil {
				durationStr = formatDuration(session.DisconnectedAt.Sub(session.ConnectedAt))
			} else {
				durationStr = formatDuration(time.Since(session.ConnectedAt))
			}

			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
				truncateID(session.ID.String()),
				nodeName,
				statusIcon,
				session.Protocol,
				dataStr,
				connectedAt,
				durationStr,
			)
		}

		w.Flush()

		return nil
	},
}

func init() {
	// Connect flags
	connectCmd.Flags().StringVarP(&connectNodeID, "node", "n", "", "Node ID to connect to")
	connectCmd.Flags().StringVarP(&connectProtocol, "protocol", "p", "", "Protocol to use (wireguard, openvpn)")
	connectCmd.Flags().StringVarP(&connectCountry, "country", "c", "", "Country to connect to")
	connectCmd.Flags().BoolVarP(&connectBest, "best", "b", false, "Connect to the best available node")

	// Disconnect flags
	disconnectCmd.Flags().StringVarP(&disconnectSessionID, "session", "s", "", "Session ID to disconnect")
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

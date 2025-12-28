package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/nikola43/aureo-vpn/cmd/cli/client"
	"github.com/spf13/cobra"
)

var (
	nodesCountry  string
	nodesProtocol string
	nodesLimit    int
)

// nodesCmd lists VPN nodes
var nodesCmd = &cobra.Command{
	Use:   "nodes",
	Short: "List available VPN nodes",
	Long: `List available VPN nodes with filtering options.

Examples:
  aureo nodes                         # List all nodes
  aureo nodes --country US            # List nodes in United States
  aureo nodes --protocol wireguard    # List WireGuard-capable nodes
  aureo nodes best                    # Get the optimal node
  aureo nodes info <node-id>          # Get details about a specific node`,
	RunE: func(cmd *cobra.Command, args []string) error {
		resp, err := APIClient.GetNodes(nodesCountry, nodesProtocol)
		if err != nil {
			return fmt.Errorf("failed to fetch nodes: %w", err)
		}

		if len(resp.Nodes) == 0 {
			fmt.Println("No nodes available matching your criteria")
			return nil
		}

		// Limit output if specified
		nodes := resp.Nodes
		if nodesLimit > 0 && nodesLimit < len(nodes) {
			nodes = nodes[:nodesLimit]
		}

		fmt.Printf("Available VPN Nodes (%d total)\n\n", resp.Count)

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "ID\tNAME\tLOCATION\tSTATUS\tLOAD\tLATENCY\tPROTOCOLS")
		fmt.Fprintln(w, "--\t----\t--------\t------\t----\t-------\t---------")

		for _, node := range nodes {
			location := node.City
			if location != "" && node.CountryCode != "" {
				location += ", " + node.CountryCode
			} else if node.CountryCode != "" {
				location = node.CountryCode
			} else {
				location = node.Country
			}

			protocols := ""
			if node.SupportsWireGuard {
				protocols = "WG"
			}
			if node.SupportsOpenVPN {
				if protocols != "" {
					protocols += ","
				}
				protocols += "OVPN"
			}

			loadStr := fmt.Sprintf("%.0f%%", node.LoadScore*100)
			statusIcon := getStatusIcon(node.Status)

			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%dms\t%s\n",
				truncateID(node.ID.String()),
				node.Name,
				location,
				statusIcon,
				loadStr,
				node.Latency,
				protocols,
			)
		}

		w.Flush()

		fmt.Printf("\nSource: %s\n", resp.Source)
		if nodesLimit > 0 && nodesLimit < resp.Count {
			fmt.Printf("Showing %d of %d nodes\n", nodesLimit, resp.Count)
		}
		return nil
	},
}

// nodesBestCmd gets the optimal node
var nodesBestCmd = &cobra.Command{
	Use:   "best",
	Short: "Get the optimal VPN node",
	Long: `Get the optimal VPN node based on load, latency, and availability.

Examples:
  aureo nodes best                    # Get the best node overall
  aureo nodes best --country DE       # Get the best node in Germany
  aureo nodes best --protocol wg      # Get the best WireGuard node`,
	RunE: func(cmd *cobra.Command, args []string) error {
		node, err := APIClient.GetBestNode(nodesCountry, nodesProtocol)
		if err != nil {
			return fmt.Errorf("failed to fetch best node: %w", err)
		}

		printNodeDetails(node)

		fmt.Println("\nTo connect to this node:")
		fmt.Printf("  aureo connect --node %s\n", node.ID.String())
		return nil
	},
}

// nodesInfoCmd gets details about a specific node
var nodesInfoCmd = &cobra.Command{
	Use:   "info <node-id>",
	Short: "Get details about a specific node",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		nodeID := args[0]

		node, err := APIClient.GetNode(nodeID)
		if err != nil {
			return fmt.Errorf("failed to fetch node: %w", err)
		}

		printNodeDetails(node)
		return nil
	},
}

func init() {
	// Nodes list flags
	nodesCmd.Flags().StringVarP(&nodesCountry, "country", "c", "", "Filter by country (e.g., US, DE, JP)")
	nodesCmd.Flags().StringVarP(&nodesProtocol, "protocol", "p", "", "Filter by protocol (wireguard, openvpn)")
	nodesCmd.Flags().IntVarP(&nodesLimit, "limit", "l", 0, "Limit number of results")

	// Best node flags (inherit from parent)
	nodesBestCmd.Flags().StringVarP(&nodesCountry, "country", "c", "", "Filter by country")
	nodesBestCmd.Flags().StringVarP(&nodesProtocol, "protocol", "p", "", "Filter by protocol")

	// Add subcommands
	nodesCmd.AddCommand(nodesBestCmd)
	nodesCmd.AddCommand(nodesInfoCmd)
}

// printNodeDetails prints detailed information about a node
func printNodeDetails(node *client.VPNNode) {
	fmt.Println("VPN Node Details")
	fmt.Println("================")
	fmt.Printf("ID:          %s\n", node.ID)
	fmt.Printf("Name:        %s\n", node.Name)
	fmt.Printf("Hostname:    %s\n", node.Hostname)
	fmt.Printf("Status:      %s %s\n", getStatusIcon(node.Status), node.Status)
	fmt.Println()

	fmt.Println("Location")
	fmt.Println("--------")
	fmt.Printf("Country:     %s (%s)\n", node.Country, node.CountryCode)
	if node.City != "" {
		fmt.Printf("City:        %s\n", node.City)
	}
	fmt.Printf("Public IP:   %s\n", node.PublicIP)
	fmt.Println()

	fmt.Println("Performance")
	fmt.Println("-----------")
	fmt.Printf("Load:        %.1f%%\n", node.LoadScore*100)
	fmt.Printf("Latency:     %d ms\n", node.Latency)
	fmt.Printf("Connections: %d / %d\n", node.CurrentConnections, node.MaxConnections)
	fmt.Printf("Uptime:      %.1f%%\n", node.UptimePercentage)
	fmt.Println()

	fmt.Println("Protocols")
	fmt.Println("---------")
	if node.SupportsWireGuard {
		fmt.Printf("WireGuard:   Port %d\n", node.WireGuardPort)
	} else {
		fmt.Println("WireGuard:   Not available")
	}
	if node.SupportsOpenVPN {
		fmt.Printf("OpenVPN:     Port %d\n", node.OpenVPNPort)
	} else {
		fmt.Println("OpenVPN:     Not available")
	}
}

// getStatusIcon returns an icon for the node status
func getStatusIcon(status string) string {
	switch status {
	case "online":
		return "[OK]"
	case "offline":
		return "[X]"
	case "maintenance":
		return "[!]"
	default:
		return "[?]"
	}
}

// truncateID shortens a UUID for display
func truncateID(id string) string {
	if len(id) > 8 {
		return id[:8]
	}
	return id
}

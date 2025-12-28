package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"
)

var (
	configNodeID   string
	configProtocol string
	configName     string
	configOutput   string
)

// configCmd is the parent command for config operations
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage VPN configurations",
	Long: `Generate and manage VPN configurations for WireGuard and OpenVPN.

Examples:
  aureo config generate --node <id>           # Generate a config for a node
  aureo config list                           # List all your configs
  aureo config show <id>                      # Show config content
  aureo config export <id> --output file.conf # Export config to file`,
}

// configGenerateCmd generates a new VPN configuration
var configGenerateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate a new VPN configuration",
	Long: `Generate a new VPN configuration for a specific node.

The configuration can be used with WireGuard or OpenVPN clients.

Examples:
  aureo config generate --node <node-id>
  aureo config generate --node <node-id> --protocol wireguard --name "Home Server"
  aureo config generate --node <node-id> --output ~/vpn.conf`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if !APIClient.IsAuthenticated() {
			return fmt.Errorf("not logged in. Use 'aureo login' first")
		}

		if configNodeID == "" {
			// Get best node if not specified
			fmt.Println("No node specified, finding best node...")
			node, err := APIClient.GetBestNode("", configProtocol)
			if err != nil {
				return fmt.Errorf("failed to find node: %w", err)
			}
			configNodeID = node.ID.String()
			fmt.Printf("Using node: %s (%s, %s)\n", node.Name, node.City, node.CountryCode)
		}

		// Default protocol
		protocol := configProtocol
		if protocol == "" {
			protocol = "wireguard"
		}

		// Default name
		name := configName
		if name == "" {
			name = fmt.Sprintf("aureo-%s-%s", protocol, time.Now().Format("20060102"))
		}

		fmt.Printf("Generating %s configuration...\n", protocol)

		config, err := APIClient.GenerateConfig(configNodeID, protocol, name)
		if err != nil {
			return fmt.Errorf("failed to generate config: %w", err)
		}

		fmt.Println()
		fmt.Println("Configuration Generated")
		fmt.Println("=======================")
		fmt.Printf("ID:         %s\n", config.ID)
		fmt.Printf("Name:       %s\n", config.ConfigName)
		fmt.Printf("Protocol:   %s\n", config.Protocol)
		fmt.Printf("Created:    %s\n", config.CreatedAt.Format(time.RFC3339))
		if config.ExpiresAt != nil {
			fmt.Printf("Expires:    %s\n", config.ExpiresAt.Format(time.RFC3339))
		}

		// Output to file if specified
		if configOutput != "" {
			if err := writeConfigToFile(configOutput, config.ConfigContent); err != nil {
				return fmt.Errorf("failed to write config: %w", err)
			}
			fmt.Printf("\nConfig saved to: %s\n", configOutput)
		} else {
			fmt.Println("\nConfiguration Content:")
			fmt.Println("----------------------")
			fmt.Println(config.ConfigContent)
			fmt.Println("\nTo save this config, use: aureo config export " + config.ID.String() + " --output file.conf")
		}

		return nil
	},
}

// configListCmd lists all configurations
var configListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all your VPN configurations",
	RunE: func(cmd *cobra.Command, args []string) error {
		if !APIClient.IsAuthenticated() {
			return fmt.Errorf("not logged in. Use 'aureo login' first")
		}

		resp, err := APIClient.GetConfigs()
		if err != nil {
			return fmt.Errorf("failed to fetch configs: %w", err)
		}

		if len(resp.Configs) == 0 {
			fmt.Println("No configurations found")
			fmt.Println("Use 'aureo config generate' to create a new configuration")
			return nil
		}

		fmt.Printf("VPN Configurations (%d total)\n\n", resp.Count)

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "ID\tNAME\tPROTOCOL\tNODE\tSTATUS\tCREATED")
		fmt.Fprintln(w, "--\t----\t--------\t----\t------\t-------")

		for _, config := range resp.Configs {
			nodeName := "Unknown"
			if config.Node != nil {
				nodeName = config.Node.Name
			}

			status := "Active"
			if !config.IsActive {
				status = "Inactive"
			}
			if config.ExpiresAt != nil && config.ExpiresAt.Before(time.Now()) {
				status = "Expired"
			}

			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n",
				truncateID(config.ID.String()),
				config.ConfigName,
				config.Protocol,
				nodeName,
				status,
				config.CreatedAt.Format("Jan 02, 2006"),
			)
		}

		w.Flush()
		return nil
	},
}

// configShowCmd shows a specific configuration
var configShowCmd = &cobra.Command{
	Use:   "show <config-id>",
	Short: "Show configuration details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if !APIClient.IsAuthenticated() {
			return fmt.Errorf("not logged in. Use 'aureo login' first")
		}

		configID := args[0]

		config, err := APIClient.GetConfig(configID)
		if err != nil {
			return fmt.Errorf("failed to fetch config: %w", err)
		}

		fmt.Println("Configuration Details")
		fmt.Println("=====================")
		fmt.Printf("ID:         %s\n", config.ID)
		fmt.Printf("Name:       %s\n", config.ConfigName)
		fmt.Printf("Protocol:   %s\n", config.Protocol)
		fmt.Printf("Public Key: %s\n", config.PublicKey)
		fmt.Printf("Active:     %v\n", config.IsActive)
		fmt.Printf("Created:    %s\n", config.CreatedAt.Format(time.RFC3339))
		if config.ExpiresAt != nil {
			fmt.Printf("Expires:    %s\n", config.ExpiresAt.Format(time.RFC3339))
		}

		if config.Node != nil {
			fmt.Println()
			fmt.Println("Node Information")
			fmt.Println("----------------")
			fmt.Printf("Node:       %s\n", config.Node.Name)
			fmt.Printf("Location:   %s, %s\n", config.Node.City, config.Node.CountryCode)
			fmt.Printf("IP:         %s\n", config.Node.PublicIP)
		}

		fmt.Println()
		fmt.Println("Configuration Content")
		fmt.Println("---------------------")
		fmt.Println(config.ConfigContent)

		return nil
	},
}

// configExportCmd exports a configuration to a file
var configExportCmd = &cobra.Command{
	Use:   "export <config-id>",
	Short: "Export configuration to a file",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if !APIClient.IsAuthenticated() {
			return fmt.Errorf("not logged in. Use 'aureo login' first")
		}

		configID := args[0]

		config, err := APIClient.GetConfig(configID)
		if err != nil {
			return fmt.Errorf("failed to fetch config: %w", err)
		}

		// Determine output filename
		output := configOutput
		if output == "" {
			ext := ".conf"
			if config.Protocol == "openvpn" {
				ext = ".ovpn"
			}
			output = config.ConfigName + ext
		}

		if err := writeConfigToFile(output, config.ConfigContent); err != nil {
			return fmt.Errorf("failed to write config: %w", err)
		}

		fmt.Printf("Configuration exported to: %s\n", output)
		fmt.Println()
		fmt.Println("Usage:")
		if config.Protocol == "wireguard" {
			fmt.Printf("  wg-quick up %s\n", output)
		} else {
			fmt.Printf("  openvpn --config %s\n", output)
		}

		return nil
	},
}

func init() {
	// Generate flags
	configGenerateCmd.Flags().StringVarP(&configNodeID, "node", "n", "", "Node ID to generate config for")
	configGenerateCmd.Flags().StringVarP(&configProtocol, "protocol", "p", "", "Protocol (wireguard, openvpn)")
	configGenerateCmd.Flags().StringVar(&configName, "name", "", "Configuration name")
	configGenerateCmd.Flags().StringVarP(&configOutput, "output", "o", "", "Output file path")

	// Export flags
	configExportCmd.Flags().StringVarP(&configOutput, "output", "o", "", "Output file path")

	// Add subcommands
	configCmd.AddCommand(configGenerateCmd)
	configCmd.AddCommand(configListCmd)
	configCmd.AddCommand(configShowCmd)
	configCmd.AddCommand(configExportCmd)
}

// writeConfigToFile writes configuration content to a file
func writeConfigToFile(path, content string) error {
	// Expand home directory
	if len(path) >= 2 && path[:2] == "~/" {
		home, err := os.UserHomeDir()
		if err != nil {
			return err
		}
		path = filepath.Join(home, path[2:])
	}

	// Create directory if needed
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	// Write file with restricted permissions
	return os.WriteFile(path, []byte(content), 0600)
}

package cmd

import (
	"fmt"
	"os"

	"github.com/nikola43/aureo-vpn/cmd/cli/client"
	"github.com/spf13/cobra"
)

var (
	// APIClient is the global API client
	APIClient *client.Client

	// Default API URL
	apiURL string

	// Version information
	Version   = "1.0.0"
	BuildTime = "unknown"
)

// rootCmd represents the base command
var rootCmd = &cobra.Command{
	Use:   "aureo",
	Short: "Aureo VPN - Secure, decentralized VPN service",
	Long: `Aureo VPN CLI - A command-line interface for the Aureo VPN service.

Aureo is a decentralized VPN service that provides secure, private internet access
with operator-owned nodes and transparent earnings.

Get started:
  aureo login              Login to your account
  aureo register           Create a new account
  aureo nodes              List available VPN nodes
  aureo connect            Connect to a VPN node
  aureo status             Check connection status

For more information, visit: https://aureo.vpn`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Skip client initialization for version command
		if cmd.Name() == "version" {
			return nil
		}

		var err error
		APIClient, err = client.NewClient(apiURL)
		if err != nil {
			return fmt.Errorf("failed to initialize client: %w", err)
		}
		return nil
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&apiURL, "api-url", "http://localhost:8080", "API server URL")

	// Add subcommands
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(loginCmd)
	rootCmd.AddCommand(logoutCmd)
	rootCmd.AddCommand(registerCmd)
	rootCmd.AddCommand(nodesCmd)
	rootCmd.AddCommand(connectCmd)
	rootCmd.AddCommand(disconnectCmd)
	rootCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(profileCmd)
	rootCmd.AddCommand(statsCmd)
	rootCmd.AddCommand(configCmd)
	rootCmd.AddCommand(trafficCmd)
	rootCmd.AddCommand(sessionsCmd)
	rootCmd.AddCommand(whoamiCmd)
}

// versionCmd shows version information
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Aureo VPN CLI\n")
		fmt.Printf("Version:    %s\n", Version)
		fmt.Printf("Build Time: %s\n", BuildTime)
	},
}

// whoamiCmd shows current logged in user
var whoamiCmd = &cobra.Command{
	Use:   "whoami",
	Short: "Show current logged in user",
	Run: func(cmd *cobra.Command, args []string) {
		if !APIClient.IsAuthenticated() {
			fmt.Println("Not logged in")
			fmt.Println("Use 'aureo login' to authenticate")
			return
		}

		email, username := APIClient.GetCurrentUser()
		fmt.Printf("Logged in as: %s (%s)\n", username, email)
	},
}

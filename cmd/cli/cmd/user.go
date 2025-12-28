package cmd

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
)

// profileCmd shows user profile
var profileCmd = &cobra.Command{
	Use:   "profile",
	Short: "Show your profile information",
	Long: `Display your Aureo VPN account profile information.

This includes your account details, subscription status, and usage summary.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if !APIClient.IsAuthenticated() {
			return fmt.Errorf("not logged in. Use 'aureo login' first")
		}

		profile, err := APIClient.GetProfile()
		if err != nil {
			return fmt.Errorf("failed to fetch profile: %w", err)
		}

		fmt.Println("Account Profile")
		fmt.Println("===============")
		fmt.Printf("User ID:      %s\n", profile.ID)
		fmt.Printf("Email:        %s\n", profile.Email)
		fmt.Printf("Username:     %s\n", profile.Username)
		if profile.FullName != "" {
			fmt.Printf("Full Name:    %s\n", profile.FullName)
		}
		fmt.Println()

		fmt.Println("Subscription")
		fmt.Println("------------")
		fmt.Printf("Tier:         %s\n", formatTier(profile.SubscriptionTier))
		fmt.Printf("Status:       %s\n", formatActiveStatus(profile.IsActive))
		fmt.Println()

		fmt.Println("Usage Summary")
		fmt.Println("-------------")
		fmt.Printf("Data Used:    %.2f GB\n", profile.DataTransferredGB)
		fmt.Printf("Connections:  %d\n", profile.ConnectionCount)
		fmt.Printf("Member Since: %s\n", profile.CreatedAt.Format("January 2006"))

		return nil
	},
}

// statsCmd shows user statistics
var statsCmd = &cobra.Command{
	Use:   "stats",
	Short: "Show your usage statistics",
	Long: `Display detailed usage statistics for your Aureo VPN account.

This includes session history, data transfer metrics, and connection patterns.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if !APIClient.IsAuthenticated() {
			return fmt.Errorf("not logged in. Use 'aureo login' first")
		}

		stats, err := APIClient.GetUserStats()
		if err != nil {
			return fmt.Errorf("failed to fetch stats: %w", err)
		}

		email, username := APIClient.GetCurrentUser()

		fmt.Printf("Usage Statistics for %s\n", username)
		fmt.Printf("Email: %s\n", email)
		fmt.Println()

		fmt.Println("Session Statistics")
		fmt.Println("==================")
		fmt.Printf("Total Sessions:     %d\n", stats.TotalSessions)
		fmt.Printf("Active Sessions:    %d\n", stats.ActiveSessions)
		if stats.LastConnected != nil {
			fmt.Printf("Last Connected:     %s\n", stats.LastConnected.Format(time.RFC3339))
		}
		fmt.Println()

		fmt.Println("Data Transfer")
		fmt.Println("=============")
		fmt.Printf("Total Data:         %.2f GB\n", stats.TotalDataGB)
		fmt.Println()

		fmt.Println("Connection Quality")
		fmt.Println("==================")
		fmt.Printf("Total Time Online:  %s\n", formatConnectionTime(stats.TotalConnectionTime))
		fmt.Printf("Average Latency:    %d ms\n", stats.AverageLatency)
		if stats.FavoriteCountry != "" {
			fmt.Printf("Favorite Country:   %s\n", stats.FavoriteCountry)
		}

		// Visual bar for data usage
		fmt.Println()
		fmt.Println("Data Usage Breakdown")
		fmt.Println("====================")
		printDataUsageBar(stats.TotalDataGB)

		return nil
	},
}

func init() {
	// No additional flags needed for profile and stats
}

// formatTier formats subscription tier for display
func formatTier(tier string) string {
	switch tier {
	case "free":
		return "Free"
	case "basic":
		return "Basic"
	case "premium":
		return "Premium"
	default:
		return tier
	}
}

// formatActiveStatus formats active status
func formatActiveStatus(active bool) string {
	if active {
		return "Active"
	}
	return "Inactive"
}

// formatConnectionTime formats connection time in seconds to human readable
func formatConnectionTime(seconds int64) string {
	duration := time.Duration(seconds) * time.Second

	days := duration / (24 * time.Hour)
	duration -= days * 24 * time.Hour
	hours := duration / time.Hour
	duration -= hours * time.Hour
	minutes := duration / time.Minute

	if days > 0 {
		return fmt.Sprintf("%dd %dh %dm", days, hours, minutes)
	}
	if hours > 0 {
		return fmt.Sprintf("%dh %dm", hours, minutes)
	}
	return fmt.Sprintf("%dm", minutes)
}

// printDataUsageBar prints a visual bar for data usage
func printDataUsageBar(dataGB float64) {
	// Assume 100GB as maximum for visualization
	maxGB := 100.0
	percentage := dataGB / maxGB
	if percentage > 1.0 {
		percentage = 1.0
	}

	barWidth := 40
	filledWidth := int(percentage * float64(barWidth))

	bar := "["
	for i := 0; i < barWidth; i++ {
		if i < filledWidth {
			bar += "="
		} else if i == filledWidth {
			bar += ">"
		} else {
			bar += " "
		}
	}
	bar += "]"

	fmt.Printf("%s %.2f GB\n", bar, dataGB)
}

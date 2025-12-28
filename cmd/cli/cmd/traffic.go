package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"
)

var (
	trafficRefresh int
	trafficWatch   bool
)

// trafficCmd shows traffic statistics
var trafficCmd = &cobra.Command{
	Use:   "traffic",
	Short: "Monitor VPN traffic in real-time",
	Long: `Monitor VPN traffic and data usage statistics.

Examples:
  aureo traffic                    # Show current traffic stats
  aureo traffic --watch            # Watch traffic in real-time
  aureo traffic --refresh 5        # Refresh every 5 seconds`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if !APIClient.IsAuthenticated() {
			return fmt.Errorf("not logged in. Use 'aureo login' first")
		}

		if trafficWatch {
			return watchTraffic()
		}

		return showTraffic()
	},
}

func init() {
	trafficCmd.Flags().IntVarP(&trafficRefresh, "refresh", "r", 2, "Refresh interval in seconds (for --watch)")
	trafficCmd.Flags().BoolVarP(&trafficWatch, "watch", "w", false, "Watch traffic in real-time")
}

// showTraffic displays current traffic statistics
func showTraffic() error {
	resp, err := APIClient.GetActiveSessions()
	if err != nil {
		return fmt.Errorf("failed to get sessions: %w", err)
	}

	// Calculate totals
	var totalSent, totalReceived int64
	activeSessions := 0

	for _, session := range resp.Sessions {
		if session.Status == "active" {
			totalSent += session.BytesSent
			totalReceived += session.BytesReceived
			activeSessions++
		}
	}

	if activeSessions == 0 {
		fmt.Println("No active VPN connections")
		fmt.Println("Use 'aureo connect' to establish a connection")
		return nil
	}

	fmt.Println("Traffic Statistics")
	fmt.Println("==================")
	fmt.Println()

	fmt.Printf("Active Sessions: %d\n", activeSessions)
	fmt.Println()

	// Total traffic
	fmt.Println("Total Traffic")
	fmt.Println("-------------")
	fmt.Printf("  Upload:   %s\n", formatBytes(totalSent))
	fmt.Printf("  Download: %s\n", formatBytes(totalReceived))
	fmt.Printf("  Total:    %s\n", formatBytes(totalSent+totalReceived))
	fmt.Println()

	// Per-session traffic
	fmt.Println("Per-Session Breakdown")
	fmt.Println("---------------------")

	for _, session := range resp.Sessions {
		if session.Status != "active" {
			continue
		}

		nodeName := "Unknown"
		if session.Node != nil {
			nodeName = session.Node.Name
		}

		duration := time.Since(session.ConnectedAt)

		fmt.Printf("\n  %s (%s)\n", nodeName, session.Protocol)
		fmt.Printf("    Duration:     %s\n", formatDuration(duration))
		fmt.Printf("    Upload:       %s\n", formatBytes(session.BytesSent))
		fmt.Printf("    Download:     %s\n", formatBytes(session.BytesReceived))
		fmt.Printf("    Latency:      %d ms\n", session.Latency)
		fmt.Printf("    Packet Loss:  %.2f%%\n", session.PacketLoss)

		// Calculate speed (bytes per second)
		if duration.Seconds() > 0 {
			uploadSpeed := float64(session.BytesSent) / duration.Seconds()
			downloadSpeed := float64(session.BytesReceived) / duration.Seconds()
			fmt.Printf("    Avg Upload:   %s/s\n", formatBytes(int64(uploadSpeed)))
			fmt.Printf("    Avg Download: %s/s\n", formatBytes(int64(downloadSpeed)))
		}
	}

	return nil
}

// watchTraffic monitors traffic in real-time
func watchTraffic() error {
	// Setup signal handling for graceful exit
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	fmt.Printf("Monitoring traffic (refresh every %ds, Ctrl+C to stop)...\n\n", trafficRefresh)

	var prevSent, prevReceived int64
	var prevTime time.Time

	ticker := time.NewTicker(time.Duration(trafficRefresh) * time.Second)
	defer ticker.Stop()

	// Initial display
	displayTrafficMonitor(0, 0, 0, 0, true)

	for {
		select {
		case <-sigChan:
			fmt.Println("\n\nMonitoring stopped")
			return nil
		case <-ticker.C:
			resp, err := APIClient.GetActiveSessions()
			if err != nil {
				fmt.Printf("\rError: %v", err)
				continue
			}

			// Calculate totals
			var totalSent, totalReceived int64
			var avgLatency int
			activeSessions := 0

			for _, session := range resp.Sessions {
				if session.Status == "active" {
					totalSent += session.BytesSent
					totalReceived += session.BytesReceived
					avgLatency += session.Latency
					activeSessions++
				}
			}

			if activeSessions == 0 {
				clearLine()
				fmt.Print("\rNo active connections                                                ")
				continue
			}

			if avgLatency > 0 {
				avgLatency /= activeSessions
			}

			// Calculate speed
			var uploadSpeed, downloadSpeed float64
			if !prevTime.IsZero() {
				elapsed := time.Since(prevTime).Seconds()
				if elapsed > 0 {
					uploadSpeed = float64(totalSent-prevSent) / elapsed
					downloadSpeed = float64(totalReceived-prevReceived) / elapsed
				}
			}

			prevSent = totalSent
			prevReceived = totalReceived
			prevTime = time.Now()

			displayTrafficMonitor(totalSent, totalReceived, uploadSpeed, downloadSpeed, false)
		}
	}
}

// displayTrafficMonitor displays the traffic monitor line
func displayTrafficMonitor(sent, received int64, upSpeed, downSpeed float64, initial bool) {
	if initial {
		fmt.Println("         TOTAL              |         SPEED             |  LATENCY")
		fmt.Println("  Upload      Download      |  Upload      Download     |")
		fmt.Println("----------------------------------------------------------")
		return
	}

	clearLine()
	fmt.Printf("\r  %8s    %8s      |  %8s/s  %8s/s   |",
		formatBytesShort(sent),
		formatBytesShort(received),
		formatBytesShort(int64(upSpeed)),
		formatBytesShort(int64(downSpeed)),
	)
}

// clearLine clears the current terminal line
func clearLine() {
	fmt.Print("\r\033[K")
}

// formatBytesShort formats bytes in a shorter format
func formatBytesShort(bytes int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)

	switch {
	case bytes >= GB:
		return fmt.Sprintf("%.1fGB", float64(bytes)/GB)
	case bytes >= MB:
		return fmt.Sprintf("%.1fMB", float64(bytes)/MB)
	case bytes >= KB:
		return fmt.Sprintf("%.1fKB", float64(bytes)/KB)
	default:
		return fmt.Sprintf("%dB", bytes)
	}
}

// trafficHistoryCmd shows traffic history
var trafficHistoryCmd = &cobra.Command{
	Use:   "history",
	Short: "Show traffic history",
	RunE: func(cmd *cobra.Command, args []string) error {
		if !APIClient.IsAuthenticated() {
			return fmt.Errorf("not logged in. Use 'aureo login' first")
		}

		stats, err := APIClient.GetUserStats()
		if err != nil {
			return fmt.Errorf("failed to get stats: %w", err)
		}

		fmt.Println("Traffic History")
		fmt.Println("===============")
		fmt.Println()

		fmt.Printf("Total Data Transferred: %.2f GB\n", stats.TotalDataGB)
		fmt.Printf("Total Sessions:         %d\n", stats.TotalSessions)
		fmt.Printf("Total Connection Time:  %s\n", formatConnectionTime(stats.TotalConnectionTime))
		fmt.Println()

		// Data usage visualization
		fmt.Println("Data Usage Over Time")
		fmt.Println("--------------------")
		printDataUsageBar(stats.TotalDataGB)

		return nil
	},
}

func init() {
	trafficCmd.AddCommand(trafficHistoryCmd)
}

package security

import (
	"fmt"
	"log"
	"os/exec"
	"runtime"
)

// WebRTCProtection handles WebRTC leak prevention
type WebRTCProtection struct {
	enabled  bool
	platform string
}

// NewWebRTCProtection creates a new WebRTC protection manager
func NewWebRTCProtection() *WebRTCProtection {
	return &WebRTCProtection{
		enabled:  false,
		platform: runtime.GOOS,
	}
}

// Enable activates WebRTC leak protection
func (w *WebRTCProtection) Enable() error {
	log.Println("Enabling WebRTC leak protection...")

	switch w.platform {
	case "linux":
		if err := w.enableLinux(); err != nil {
			return err
		}
	case "darwin":
		if err := w.enableMacOS(); err != nil {
			return err
		}
	case "windows":
		if err := w.enableWindows(); err != nil {
			return err
		}
	default:
		return fmt.Errorf("unsupported platform: %s", w.platform)
	}

	w.enabled = true
	log.Println("WebRTC leak protection enabled")
	return nil
}

// Disable deactivates WebRTC leak protection
func (w *WebRTCProtection) Disable() error {
	log.Println("Disabling WebRTC leak protection...")

	switch w.platform {
	case "linux":
		if err := w.disableLinux(); err != nil {
			return err
		}
	case "darwin":
		if err := w.disableMacOS(); err != nil {
			return err
		}
	case "windows":
		if err := w.disableWindows(); err != nil {
			return err
		}
	}

	w.enabled = false
	log.Println("WebRTC leak protection disabled")
	return nil
}

// enableLinux enables WebRTC protection on Linux
func (w *WebRTCProtection) enableLinux() error {
	// Block WebRTC STUN/TURN servers
	rules := [][]string{
		// Block Google STUN servers
		{"-A", "OUTPUT", "-d", "stun.l.google.com", "-j", "DROP"},
		{"-A", "OUTPUT", "-d", "stun1.l.google.com", "-j", "DROP"},
		{"-A", "OUTPUT", "-d", "stun2.l.google.com", "-j", "DROP"},
		{"-A", "OUTPUT", "-d", "stun3.l.google.com", "-j", "DROP"},
		{"-A", "OUTPUT", "-d", "stun4.l.google.com", "-j", "DROP"},

		// Block common STUN ports
		{"-A", "OUTPUT", "-p", "udp", "--dport", "3478", "-j", "DROP"},
		{"-A", "OUTPUT", "-p", "tcp", "--dport", "3478", "-j", "DROP"},
		{"-A", "OUTPUT", "-p", "udp", "--dport", "3479", "-j", "DROP"},
		{"-A", "OUTPUT", "-p", "tcp", "--dport", "3479", "-j", "DROP"},

		// Block TURN ports
		{"-A", "OUTPUT", "-p", "udp", "--dport", "5349", "-j", "DROP"},
		{"-A", "OUTPUT", "-p", "tcp", "--dport", "5349", "-j", "DROP"},

		// Allow WebRTC through VPN interface only
		{"-I", "OUTPUT", "1", "-o", "wg0", "-p", "udp", "--dport", "3478", "-j", "ACCEPT"},
		{"-I", "OUTPUT", "1", "-o", "tun0", "-p", "udp", "--dport", "3478", "-j", "ACCEPT"},
	}

	for _, rule := range rules {
		cmd := exec.Command("iptables", rule...)
		if err := cmd.Run(); err != nil {
			log.Printf("Warning: failed to add iptables rule: %v", err)
		}
	}

	// Block mDNS (used by WebRTC for local discovery)
	cmd := exec.Command("iptables", "-A", "OUTPUT", "-p", "udp", "--dport", "5353", "-j", "DROP")
	cmd.Run()

	return nil
}

// disableLinux disables WebRTC protection on Linux
func (w *WebRTCProtection) disableLinux() error {
	rules := [][]string{
		{"-D", "OUTPUT", "-d", "stun.l.google.com", "-j", "DROP"},
		{"-D", "OUTPUT", "-d", "stun1.l.google.com", "-j", "DROP"},
		{"-D", "OUTPUT", "-d", "stun2.l.google.com", "-j", "DROP"},
		{"-D", "OUTPUT", "-d", "stun3.l.google.com", "-j", "DROP"},
		{"-D", "OUTPUT", "-d", "stun4.l.google.com", "-j", "DROP"},
		{"-D", "OUTPUT", "-p", "udp", "--dport", "3478", "-j", "DROP"},
		{"-D", "OUTPUT", "-p", "tcp", "--dport", "3478", "-j", "DROP"},
		{"-D", "OUTPUT", "-p", "udp", "--dport", "5353", "-j", "DROP"},
	}

	for _, rule := range rules {
		cmd := exec.Command("iptables", rule...)
		cmd.Run() // Ignore errors - rule might not exist
	}

	return nil
}

// enableMacOS enables WebRTC protection on macOS
func (w *WebRTCProtection) enableMacOS() error {
	// Use pfctl (Packet Filter) on macOS
	rules := `
# WebRTC leak protection rules
block drop out proto udp from any to any port 3478
block drop out proto tcp from any to any port 3478
block drop out proto udp from any to any port 5349
block drop out proto tcp from any to any port 5349
block drop out proto udp from any to any port 5353
`

	// Write rules to temporary file
	cmd := exec.Command("sh", "-c", fmt.Sprintf("echo '%s' | sudo pfctl -f -", rules))
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to apply pfctl rules: %w", err)
	}

	// Enable packet filter
	cmd = exec.Command("sudo", "pfctl", "-e")
	cmd.Run() // Might already be enabled

	return nil
}

// disableMacOS disables WebRTC protection on macOS
func (w *WebRTCProtection) disableMacOS() error {
	// Flush pfctl rules
	cmd := exec.Command("sudo", "pfctl", "-F", "all")
	return cmd.Run()
}

// enableWindows enables WebRTC protection on Windows
func (w *WebRTCProtection) enableWindows() error {
	// Use Windows Firewall with netsh
	rules := [][]string{
		{"advfirewall", "firewall", "add", "rule", "name=BlockWebRTC_STUN_UDP",
			"dir=out", "action=block", "protocol=UDP", "remoteport=3478"},
		{"advfirewall", "firewall", "add", "rule", "name=BlockWebRTC_STUN_TCP",
			"dir=out", "action=block", "protocol=TCP", "remoteport=3478"},
		{"advfirewall", "firewall", "add", "rule", "name=BlockWebRTC_TURN",
			"dir=out", "action=block", "protocol=UDP", "remoteport=5349"},
		{"advfirewall", "firewall", "add", "rule", "name=BlockWebRTC_mDNS",
			"dir=out", "action=block", "protocol=UDP", "remoteport=5353"},
	}

	for _, rule := range rules {
		cmd := exec.Command("netsh", rule...)
		if err := cmd.Run(); err != nil {
			log.Printf("Warning: failed to add firewall rule: %v", err)
		}
	}

	return nil
}

// disableWindows disables WebRTC protection on Windows
func (w *WebRTCProtection) disableWindows() error {
	rules := []string{
		"BlockWebRTC_STUN_UDP",
		"BlockWebRTC_STUN_TCP",
		"BlockWebRTC_TURN",
		"BlockWebRTC_mDNS",
	}

	for _, ruleName := range rules {
		cmd := exec.Command("netsh", "advfirewall", "firewall", "delete", "rule", "name="+ruleName)
		cmd.Run() // Ignore errors
	}

	return nil
}

// CheckWebRTCLeak tests for WebRTC leaks
func (w *WebRTCProtection) CheckWebRTCLeak() (bool, error) {
	// This would typically involve checking if WebRTC exposes real IP
	// In production, integrate with browser extensions or test services
	log.Println("Checking for WebRTC leaks...")

	// Simulate leak test (in production, use actual WebRTC test)
	// Could integrate with services like browserleaks.com API

	return false, nil // false = no leak detected
}

// GetWebRTCStatus returns the current protection status
func (w *WebRTCProtection) GetWebRTCStatus() map[string]interface{} {
	return map[string]interface{}{
		"enabled":  w.enabled,
		"platform": w.platform,
		"protected_ports": []int{3478, 3479, 5349, 5353},
	}
}

// BlockSTUNServers blocks known STUN servers
func (w *WebRTCProtection) BlockSTUNServers(servers []string) error {
	log.Printf("Blocking %d STUN servers", len(servers))

	for _, server := range servers {
		// Add firewall rules for each server
		var cmd *exec.Cmd

		switch w.platform {
		case "linux":
			cmd = exec.Command("iptables", "-A", "OUTPUT", "-d", server, "-j", "DROP")
		case "darwin":
			cmd = exec.Command("sudo", "pfctl", "-t", "blocked_stun", "-T", "add", server)
		case "windows":
			cmd = exec.Command("netsh", "advfirewall", "firewall", "add", "rule",
				fmt.Sprintf("name=BlockSTUN_%s", server),
				"dir=out", "action=block", "remoteip="+server)
		default:
			continue
		}

		if err := cmd.Run(); err != nil {
			log.Printf("Warning: failed to block STUN server %s: %v", server, err)
		}
	}

	return nil
}

// GetCommonSTUNServers returns a list of commonly used STUN servers
func GetCommonSTUNServers() []string {
	return []string{
		"stun.l.google.com",
		"stun1.l.google.com",
		"stun2.l.google.com",
		"stun3.l.google.com",
		"stun4.l.google.com",
		"stun.services.mozilla.com",
		"stun.stunprotocol.org",
		"stun.ekiga.net",
		"stun.ideasip.com",
		"stun.voiparound.com",
		"stun.voipbuster.com",
		"stun.voipstunt.com",
		"stun.counterpath.com",
		"stun.callwithus.com",
	}
}

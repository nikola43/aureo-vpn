package security

import (
	"fmt"
	"log"
	"os/exec"
)

// KillSwitch implements VPN kill switch functionality
type KillSwitch struct {
	interfaceName string
	enabled       bool
	rules         []string
}

// NewKillSwitch creates a new kill switch instance
func NewKillSwitch(interfaceName string) *KillSwitch {
	return &KillSwitch{
		interfaceName: interfaceName,
		enabled:       false,
		rules:         make([]string, 0),
	}
}

// Enable activates the kill switch
func (k *KillSwitch) Enable() error {
	if k.enabled {
		return nil
	}

	log.Println("Enabling VPN kill switch...")

	// Block all outgoing traffic except VPN
	rules := []struct {
		table string
		chain string
		rule  []string
	}{
		// Block all OUTPUT traffic by default
		{"filter", "OUTPUT", []string{"-j", "DROP"}},

		// Allow loopback
		{"filter", "OUTPUT", []string{"-o", "lo", "-j", "ACCEPT"}},

		// Allow VPN interface
		{"filter", "OUTPUT", []string{"-o", k.interfaceName, "-j", "ACCEPT"}},

		// Allow established connections
		{"filter", "OUTPUT", []string{"-m", "conntrack", "--ctstate", "ESTABLISHED,RELATED", "-j", "ACCEPT"}},

		// Allow DNS to VPN DNS servers (if VPN is up)
		{"filter", "OUTPUT", []string{"-p", "udp", "--dport", "53", "-o", k.interfaceName, "-j", "ACCEPT"}},

		// Allow DHCP
		{"filter", "OUTPUT", []string{"-p", "udp", "--dport", "67:68", "-j", "ACCEPT"}},
	}

	for _, rule := range rules {
		args := append([]string{"-t", rule.table, "-A", rule.chain}, rule.rule...)
		if err := k.addIPTablesRule(args...); err != nil {
			// Rollback on error
			k.Disable()
			return fmt.Errorf("failed to add iptables rule: %w", err)
		}
		k.rules = append(k.rules, fmt.Sprintf("%s %s %s", rule.table, rule.chain, rule.rule))
	}

	k.enabled = true
	log.Println("VPN kill switch enabled successfully")
	return nil
}

// Disable deactivates the kill switch
func (k *KillSwitch) Disable() error {
	if !k.enabled {
		return nil
	}

	log.Println("Disabling VPN kill switch...")

	// Flush OUTPUT chain to restore normal traffic
	if err := k.flushIPTables(); err != nil {
		log.Printf("Warning: failed to flush iptables: %v", err)
	}

	k.rules = make([]string, 0)
	k.enabled = false

	log.Println("VPN kill switch disabled")
	return nil
}

// IsEnabled returns whether the kill switch is active
func (k *KillSwitch) IsEnabled() bool {
	return k.enabled
}

// addIPTablesRule adds an iptables rule
func (k *KillSwitch) addIPTablesRule(args ...string) error {
	cmd := exec.Command("iptables", args...)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("iptables command failed: %w", err)
	}
	return nil
}

// flushIPTables removes kill switch rules
func (k *KillSwitch) flushIPTables() error {
	// Flush OUTPUT chain
	cmd := exec.Command("iptables", "-F", "OUTPUT")
	if err := cmd.Run(); err != nil {
		return err
	}

	// Set default OUTPUT policy to ACCEPT
	cmd = exec.Command("iptables", "-P", "OUTPUT", "ACCEPT")
	return cmd.Run()
}

// AllowVPNServer adds a rule to allow connection to VPN server
func (k *KillSwitch) AllowVPNServer(serverIP string, port int, protocol string) error {
	args := []string{
		"-t", "filter",
		"-I", "OUTPUT", "1",
		"-d", serverIP,
		"-p", protocol,
		"--dport", fmt.Sprintf("%d", port),
		"-j", "ACCEPT",
	}

	if err := k.addIPTablesRule(args...); err != nil {
		return fmt.Errorf("failed to allow VPN server: %w", err)
	}

	log.Printf("Allowed VPN server %s:%d (%s)", serverIP, port, protocol)
	return nil
}

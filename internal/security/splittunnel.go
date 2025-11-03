package security

import (
	"fmt"
	"log"
	"net"
	"os/exec"
)

// SplitTunnel implements split tunneling functionality
type SplitTunnel struct {
	vpnInterface  string
	includedRules []Rule
	excludedRules []Rule
	routingTable  int
}

// Rule represents a routing rule
type Rule struct {
	Type        string // "ip", "domain", "app"
	Value       string
	Description string
}

// NewSplitTunnel creates a new split tunnel manager
func NewSplitTunnel(vpnInterface string) *SplitTunnel {
	return &SplitTunnel{
		vpnInterface:  vpnInterface,
		includedRules: make([]Rule, 0),
		excludedRules: make([]Rule, 0),
		routingTable:  100, // Custom routing table number
	}
}

// AddIncludeRule adds a rule to route specific traffic through VPN
func (s *SplitTunnel) AddIncludeRule(rule Rule) error {
	log.Printf("Adding include rule: %s (%s)", rule.Value, rule.Type)

	switch rule.Type {
	case "ip":
		if err := s.routeIPThroughVPN(rule.Value); err != nil {
			return err
		}
	case "domain":
		if err := s.routeDomainThroughVPN(rule.Value); err != nil {
			return err
		}
	case "subnet":
		if err := s.routeSubnetThroughVPN(rule.Value); err != nil {
			return err
		}
	default:
		return fmt.Errorf("unknown rule type: %s", rule.Type)
	}

	s.includedRules = append(s.includedRules, rule)
	return nil
}

// AddExcludeRule adds a rule to route specific traffic outside VPN
func (s *SplitTunnel) AddExcludeRule(rule Rule) error {
	log.Printf("Adding exclude rule: %s (%s)", rule.Value, rule.Type)

	switch rule.Type {
	case "ip":
		if err := s.routeIPOutsideVPN(rule.Value); err != nil {
			return err
		}
	case "domain":
		if err := s.routeDomainOutsideVPN(rule.Value); err != nil {
			return err
		}
	case "subnet":
		if err := s.routeSubnetOutsideVPN(rule.Value); err != nil {
			return err
		}
	default:
		return fmt.Errorf("unknown rule type: %s", rule.Type)
	}

	s.excludedRules = append(s.excludedRules, rule)
	return nil
}

// Enable activates split tunneling
func (s *SplitTunnel) Enable() error {
	log.Println("Enabling split tunneling...")

	// Create custom routing table
	if err := s.setupCustomRoutingTable(); err != nil {
		return fmt.Errorf("failed to setup routing table: %w", err)
	}

	// Apply all include rules
	for _, rule := range s.includedRules {
		if err := s.AddIncludeRule(rule); err != nil {
			log.Printf("Warning: failed to apply include rule: %v", err)
		}
	}

	// Apply all exclude rules
	for _, rule := range s.excludedRules {
		if err := s.AddExcludeRule(rule); err != nil {
			log.Printf("Warning: failed to apply exclude rule: %v", err)
		}
	}

	log.Println("Split tunneling enabled")
	return nil
}

// Disable deactivates split tunneling
func (s *SplitTunnel) Disable() error {
	log.Println("Disabling split tunneling...")

	// Flush custom routing table
	if err := s.flushRoutingTable(); err != nil {
		log.Printf("Warning: failed to flush routing table: %v", err)
	}

	// Clear rules
	s.includedRules = make([]Rule, 0)
	s.excludedRules = make([]Rule, 0)

	log.Println("Split tunneling disabled")
	return nil
}

// setupCustomRoutingTable creates a custom routing table for split tunneling
func (s *SplitTunnel) setupCustomRoutingTable() error {
	// Add custom routing table to rt_tables if not exists
	cmd := exec.Command("bash", "-c",
		fmt.Sprintf("grep -q '%d vpn' /etc/iproute2/rt_tables || echo '%d vpn' >> /etc/iproute2/rt_tables",
			s.routingTable, s.routingTable))
	if err := cmd.Run(); err != nil {
		return err
	}

	// Add default route via VPN to custom table
	cmd = exec.Command("ip", "route", "add", "default", "dev", s.vpnInterface, "table", fmt.Sprintf("%d", s.routingTable))
	if err := cmd.Run(); err != nil {
		log.Printf("Warning: failed to add default route to custom table: %v", err)
	}

	return nil
}

// routeIPThroughVPN routes a specific IP through VPN
func (s *SplitTunnel) routeIPThroughVPN(ip string) error {
	// Add IP rule to use VPN routing table
	cmd := exec.Command("ip", "rule", "add", "to", ip, "table", fmt.Sprintf("%d", s.routingTable))
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to add IP rule: %w", err)
	}

	// Add route
	cmd = exec.Command("ip", "route", "add", ip, "dev", s.vpnInterface)
	if err := cmd.Run(); err != nil {
		log.Printf("Warning: failed to add route: %v", err)
	}

	return nil
}

// routeIPOutsideVPN routes a specific IP outside VPN (direct connection)
func (s *SplitTunnel) routeIPOutsideVPN(ip string) error {
	// Get default gateway
	gateway, err := s.getDefaultGateway()
	if err != nil {
		return err
	}

	// Add route via default gateway (outside VPN)
	cmd := exec.Command("ip", "route", "add", ip, "via", gateway)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to add route: %w", err)
	}

	return nil
}

// routeDomainThroughVPN resolves domain and routes through VPN
func (s *SplitTunnel) routeDomainThroughVPN(domain string) error {
	// Resolve domain to IPs
	ips, err := net.LookupIP(domain)
	if err != nil {
		return fmt.Errorf("failed to resolve domain: %w", err)
	}

	// Route all resolved IPs through VPN
	for _, ip := range ips {
		if err := s.routeIPThroughVPN(ip.String()); err != nil {
			log.Printf("Warning: failed to route IP %s: %v", ip.String(), err)
		}
	}

	return nil
}

// routeDomainOutsideVPN resolves domain and routes outside VPN
func (s *SplitTunnel) routeDomainOutsideVPN(domain string) error {
	ips, err := net.LookupIP(domain)
	if err != nil {
		return fmt.Errorf("failed to resolve domain: %w", err)
	}

	for _, ip := range ips {
		if err := s.routeIPOutsideVPN(ip.String()); err != nil {
			log.Printf("Warning: failed to route IP %s: %v", ip.String(), err)
		}
	}

	return nil
}

// routeSubnetThroughVPN routes an entire subnet through VPN
func (s *SplitTunnel) routeSubnetThroughVPN(subnet string) error {
	// Validate CIDR
	_, _, err := net.ParseCIDR(subnet)
	if err != nil {
		return fmt.Errorf("invalid subnet: %w", err)
	}

	cmd := exec.Command("ip", "route", "add", subnet, "dev", s.vpnInterface)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to add subnet route: %w", err)
	}

	return nil
}

// routeSubnetOutsideVPN routes an entire subnet outside VPN
func (s *SplitTunnel) routeSubnetOutsideVPN(subnet string) error {
	_, _, err := net.ParseCIDR(subnet)
	if err != nil {
		return fmt.Errorf("invalid subnet: %w", err)
	}

	gateway, err := s.getDefaultGateway()
	if err != nil {
		return err
	}

	cmd := exec.Command("ip", "route", "add", subnet, "via", gateway)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to add subnet route: %w", err)
	}

	return nil
}

// getDefaultGateway gets the default network gateway
func (s *SplitTunnel) getDefaultGateway() (string, error) {
	cmd := exec.Command("ip", "route", "show", "default")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	// Parse output to get gateway
	// Format: "default via 192.168.1.1 dev eth0"
	var gateway string
	_, err = fmt.Sscanf(string(output), "default via %s", &gateway)
	if err != nil {
		return "", fmt.Errorf("failed to parse default gateway: %w", err)
	}

	return gateway, nil
}

// flushRoutingTable removes all routes from custom table
func (s *SplitTunnel) flushRoutingTable() error {
	cmd := exec.Command("ip", "route", "flush", "table", fmt.Sprintf("%d", s.routingTable))
	return cmd.Run()
}

// GetActiveRules returns all active split tunnel rules
func (s *SplitTunnel) GetActiveRules() (included, excluded []Rule) {
	return s.includedRules, s.excludedRules
}

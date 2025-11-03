package main

import (
	"net"
	"strings"
)

// InterfaceInfo represents basic network interface information
type InterfaceInfo struct {
	Name         string
	HardwareAddr string
	Flags        string
	MTU          int
	Status       string // "up" or "down"
	IPv4Addrs    []string
	IPv6Addrs    []string
	Type         string // "wired", "wireless", "vpn", "loopback", "other"
}

// RouteInfo represents routing table information
type RouteInfo struct {
	Destination string
	Gateway     string
	Netmask     string
	Interface   string
	Flags       string
}

// DNSInfo represents DNS configuration
type DNSInfo struct {
	Servers      []string
	SearchDomains []string
}

// WiredInterfaceInfo represents wired interface information
type WiredInterfaceInfo struct {
	InterfaceName string
	AdapterName   string
	Description   string
	MACAddress    string
	IPAddress     string
	IPv6Address   string
	SubnetMask    string
	Status        string
	Speed         string
	Duplex        string
	MTU           int
}

// GetInterfaces retrieves all network interfaces
func GetInterfaces() ([]*InterfaceInfo, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}

	var result []*InterfaceInfo
	for _, iface := range interfaces {
		info := &InterfaceInfo{
			Name:         iface.Name,
			HardwareAddr: iface.HardwareAddr.String(),
			Flags:        iface.Flags.String(),
			MTU:          iface.MTU,
		}

		// Determine status
		if iface.Flags&net.FlagUp != 0 {
			info.Status = "up"
		} else {
			info.Status = "down"
		}

		// Get addresses
		addrs, err := iface.Addrs()
		if err == nil {
			for _, addr := range addrs {
				if ipNet, ok := addr.(*net.IPNet); ok {
					if ipNet.IP.To4() != nil {
						info.IPv4Addrs = append(info.IPv4Addrs, ipNet.IP.String())
					} else {
						info.IPv6Addrs = append(info.IPv6Addrs, ipNet.IP.String())
					}
				}
			}
		}

		// Determine interface type
		info.Type = GetInterfaceType(info)

		result = append(result, info)
	}

	return result, nil
}

// IsActiveInterface determines if an interface is actively used for real network communication
// An active interface has:
// - Interface is up
// - Has a non-loopback, non-link-local IPv4 address (or meaningful IPv6)
// - Has a hardware MAC address (not purely virtual)
// - Not a loopback interface
func IsActiveInterface(iface *InterfaceInfo) bool {
	// Must be up
	if iface.Status != "up" {
		return false
	}
	
	// Skip loopback interfaces
	if iface.Name == "lo" || iface.Name == "lo0" || strings.Contains(iface.Flags, "Loopback") {
		return false
	}
	
	// Must have a hardware address (real network interface)
	if iface.HardwareAddr == "" || iface.HardwareAddr == "00:00:00:00:00:00" {
		return false
	}
	
	// Must have a real IPv4 address (not loopback, not link-local)
	hasRealIPv4 := false
	for _, ip := range iface.IPv4Addrs {
		ipAddr := net.ParseIP(ip)
		if ipAddr == nil {
			continue
		}
		
		// Skip loopback addresses
		if ipAddr.IsLoopback() {
			continue
		}
		
		// Skip link-local addresses (169.254.x.x)
		if ipAddr.IsLinkLocalUnicast() {
			continue
		}
		
		// This is a real, routable IPv4 address
		hasRealIPv4 = true
		break
	}
	
	// If no real IPv4, check for meaningful IPv6 (not link-local)
	if !hasRealIPv4 {
		for _, ip := range iface.IPv6Addrs {
			ipAddr := net.ParseIP(ip)
			if ipAddr == nil {
				continue
			}
			
			// Skip loopback
			if ipAddr.IsLoopback() {
				continue
			}
			
			// Skip link-local (fe80::)
			if ipAddr.IsLinkLocalUnicast() {
				continue
			}
			
			// Has a real IPv6 address
			hasRealIPv4 = true
			break
		}
	}
	
	return hasRealIPv4
}

// FilterInterfaces filters interfaces based on criteria
func FilterInterfaces(interfaces []*InterfaceInfo, nameFilter string, upOnly, downOnly, activeOnly bool) []*InterfaceInfo {
	var result []*InterfaceInfo
	for _, iface := range interfaces {
		// Filter by name
		if nameFilter != "" && iface.Name != nameFilter {
			continue
		}

		// Filter by status
		if upOnly && iface.Status != "up" {
			continue
		}
		if downOnly && iface.Status != "down" {
			continue
		}
		
		// Filter for active interfaces (primary interfaces with real IP configs)
		if activeOnly && !IsActiveInterface(iface) {
			continue
		}

		result = append(result, iface)
	}
	return result
}

// IsVPNInterface determines if an interface is likely a VPN interface
// This is a heuristic based on interface name patterns and flags
func IsVPNInterface(ifaceName string, flags string) bool {
	ifaceLower := strings.ToLower(ifaceName)
	
	// Common VPN interface patterns
	vpnPatterns := []string{"tun", "tap", "ppp", "wg", "utun", "ipsec", "vpn", "pptp", "l2tp", "openvpn", "wireguard"}
	
	// Check name patterns
	for _, pattern := range vpnPatterns {
		if strings.Contains(ifaceLower, pattern) {
			return true
		}
	}
	
	// Check for Point-to-Point flag (common in VPNs)
	if strings.Contains(flags, "PointToPoint") || strings.Contains(flags, "point-to-point") {
		return true
	}
	
	return false
}

// IsWirelessInterface determines if an interface is likely a wireless interface
func IsWirelessInterface(ifaceName string) bool {
	ifaceLower := strings.ToLower(ifaceName)
	
	// Common wireless interface patterns
	wifiPatterns := []string{"wlan", "wifi", "wlp", "wl", "airport"}
	
	// Check name patterns
	for _, pattern := range wifiPatterns {
		if strings.Contains(ifaceLower, pattern) {
			return true
		}
	}
	
	// For macOS, en0 is usually WiFi on MacBooks
	// We'll check this more carefully - en0 could be either, but we'll use other heuristics
	if ifaceLower == "en0" {
		// Could be WiFi or wired - we'll need to check elsewhere
		return false // Let other logic determine
	}
	
	return false
}

// IsWiredInterface determines if an interface is likely a wired interface
// This is a heuristic based on interface name patterns
func IsWiredInterface(ifaceName string) bool {
	ifaceLower := strings.ToLower(ifaceName)
	
	// Common wired interface patterns (excluding WiFi)
	wiredPatterns := []string{"eth", "enp", "ens", "em", "usb", "ethernet", "local area connection"}
	
	// Check for wired patterns
	for _, pattern := range wiredPatterns {
		if strings.Contains(ifaceLower, pattern) {
			return true
		}
	}
	
	// For macOS, en1, en2, en3, en4, en5+ are usually wired (Thunderbolt, USB Ethernet, etc.)
	// en0 is ambiguous (could be WiFi or wired depending on Mac model)
	if strings.HasPrefix(ifaceLower, "en") && len(ifaceLower) > 2 {
		// Check if it's en1, en2, en3, etc. (usually wired)
		if ifaceLower == "en1" || ifaceLower == "en2" || ifaceLower == "en3" || ifaceLower == "en4" || ifaceLower == "en5" || ifaceLower == "en6" || ifaceLower == "en7" || ifaceLower == "en8" {
			return true
		}
	}
	
	return false
}

// GetInterfaceType determines the type of interface (wired, wireless, vpn, loopback, other)
func GetInterfaceType(iface *InterfaceInfo) string {
	name := iface.Name
	flags := iface.Flags
	
	// Check for loopback first
	if name == "lo" || name == "lo0" || strings.Contains(flags, "Loopback") {
		return "loopback"
	}
	
	// Check for VPN
	if IsVPNInterface(name, flags) {
		return "vpn"
	}
	
	// Check if this interface is in the WiFi adapters list (most accurate)
	// This helps identify WiFi interfaces that might not match name patterns
	wifiAdapters, err := GetAllWiFiAdapters()
	if err == nil {
		for _, wifiAdapter := range wifiAdapters {
			if wifiAdapter.InterfaceName == name {
				return "wireless"
			}
		}
	}
	
	// Check for wireless (WiFi) by name patterns
	if IsWirelessInterface(name) {
		return "wireless"
	}
	
	// Special case for macOS en0 - check WiFi adapters first, then fallback
	if name == "en0" {
		// If we found it in WiFi adapters, it's wireless
		// Otherwise, check if it matches wired patterns
		foundInWiFi := false
		for _, wifiAdapter := range wifiAdapters {
			if wifiAdapter.InterfaceName == name {
				foundInWiFi = true
				break
			}
		}
		if !foundInWiFi && IsWiredInterface(name) {
			return "wired"
		}
		// Default to wireless for en0 if not found elsewhere
		return "wireless"
	}
	
	// Check for wired
	if IsWiredInterface(name) {
		return "wired"
	}
	
	return "other"
}

// GetWiredInterfaces retrieves wired interfaces from all interfaces
func GetWiredInterfaces() ([]*WiredInterfaceInfo, error) {
	interfaces, err := GetInterfaces()
	if err != nil {
		return nil, err
	}
	
	var wired []*WiredInterfaceInfo
	for _, iface := range interfaces {
		if IsWiredInterface(iface.Name) && iface.HardwareAddr != "" {
			wiredInfo := &WiredInterfaceInfo{
				InterfaceName: iface.Name,
				AdapterName:   iface.Name,
				MACAddress:    iface.HardwareAddr,
				Status:        iface.Status,
				MTU:           iface.MTU,
			}
			
			if len(iface.IPv4Addrs) > 0 {
				wiredInfo.IPAddress = iface.IPv4Addrs[0]
			}
			if len(iface.IPv6Addrs) > 0 {
				wiredInfo.IPv6Address = iface.IPv6Addrs[0]
			}
			
			wired = append(wired, wiredInfo)
		}
	}
	
	return wired, nil
}


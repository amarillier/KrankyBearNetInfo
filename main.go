package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"

	updatechecker "github.com/amarillier/go-update-checker"
)

var appName = "KrankyBear NetInfo"
var appVersion = "0.1.2"
var appCopyright = "Copyright (c) Allan Marillier, 2025-" + strconv.Itoa(time.Now().Year())

func showMainHelp() {
	fmt.Printf("%s, version %s\n", appName, appVersion)
	fmt.Printf("%s\n\n", appCopyright)
	fmt.Println("Usage: netinfo <command> [options]")
	fmt.Println("\nCommands:")
	fmt.Println("  arp               Show ARP table")
	fmt.Println("  connections       Show active TCP/UDP connections")
	fmt.Println("  dns               Show DNS servers and resolver configuration")
	fmt.Println("  interfaces        Show all network interfaces with basic info")
	fmt.Println("  ping              Ping a host (network diagnostics)")
	fmt.Println("  routes            Show routing table and default gateway")
	fmt.Println("  stats             Show network interface statistics")
	fmt.Println("  trace             Traceroute to a host (network diagnostics)")
	fmt.Println("  wifi              Show WiFi adapter information")
	fmt.Println("  wired             Show wired (Ethernet) interface information")
	fmt.Println("\nGlobal Options:")
	fmt.Println("  -checkupdate, -cu  Check for available updates")
	fmt.Println("\nFor detailed help on any command, including all options and examples:")
	fmt.Println("  netinfo <command> -help")
}

func showWiFiHelp() {
	fmt.Printf("%s - WiFi Command\n", appName)
	fmt.Printf("%s\n\n", appCopyright)
	fmt.Println("Usage: netinfo wifi [options]")
	fmt.Println("\nOptions:")
	fmt.Println("  -help, -?          Show this help message")
	fmt.Println("  -interface <name>  Filter by interface name")
	fmt.Println("  -list              List all WiFi adapters (summary)")
	fmt.Println("\nExamples:")
	fmt.Println("  netinfo wifi                    # Show detailed info for all adapters")
	fmt.Println("  netinfo wifi -list              # List all adapters")
	fmt.Println("  netinfo wifi -interface wlan0   # Show info for specific interface")
}

func showWiredHelp() {
	fmt.Printf("%s - Wired Command\n", appName)
	fmt.Printf("%s\n\n", appCopyright)
	fmt.Println("Usage: netinfo wired [options]")
	fmt.Println("\nOptions:")
	fmt.Println("  -active            Show only active interfaces")
	fmt.Println("  -help, -?          Show this help message")
	fmt.Println("  -inactive          Show only inactive interfaces")
	fmt.Println("  -interface <name>  Filter by interface name")
	fmt.Println("  -list              List all wired interfaces (summary)")
	fmt.Println("\nExamples:")
	fmt.Println("  netinfo wired                   # Show detailed info for all wired interfaces")
	fmt.Println("  netinfo wired -list             # List all wired interfaces")
	fmt.Println("  netinfo wired -interface eth0   # Show info for specific interface")
	fmt.Println("  netinfo wired -active           # Show only active wired interfaces")
}

func showInterfacesHelp() {
	fmt.Printf("%s - Interfaces Command\n", appName)
	fmt.Printf("%s\n\n", appCopyright)
	fmt.Println("Usage: netinfo interfaces [options]")
	fmt.Println("\nOptions:")
	fmt.Println("  -active             Show only primary interfaces with active IP configs")
	fmt.Println("  -down               Show only interfaces that are down")
	fmt.Println("  -help, -?            Show this help message")
	fmt.Println("  -interface <name>   Filter by interface name")
	fmt.Println("  -list               List all interfaces (summary)")
	fmt.Println("  -up                 Show only interfaces that are up")
	fmt.Println("\nExamples:")
	fmt.Println("  netinfo interfaces              # Show all network interfaces")
	fmt.Println("  netinfo interfaces -list        # List all interfaces")
	fmt.Println("  netinfo interfaces -active      # Show only primary active interfaces")
	fmt.Println("  netinfo interfaces -interface eth0  # Show specific interface")
	fmt.Println("  netinfo interfaces -up          # Show only interfaces that are up")
}

func showRoutesHelp() {
	fmt.Printf("%s - Routes Command\n", appName)
	fmt.Printf("%s\n\n", appCopyright)
	fmt.Println("Usage: netinfo routes [options]")
	fmt.Println("\nOptions:")
	fmt.Println("  -default            Show only default route")
	fmt.Println("  -help, -?           Show this help message")
	fmt.Println("  -interface <name>   Filter by interface")
	fmt.Println("  -ipv4               Show only IPv4 routes")
	fmt.Println("  -ipv6               Show only IPv6 routes")
	fmt.Println("\nExamples:")
	fmt.Println("  netinfo routes                  # Show all routes")
	fmt.Println("  netinfo routes -default          # Show default gateway")
	fmt.Println("  netinfo routes -interface eth0    # Show routes for specific interface")
	fmt.Println("  netinfo routes -ipv4             # Show only IPv4 routes")
}

func showDNSHelp() {
	fmt.Printf("%s - DNS Command\n", appName)
	fmt.Printf("%s\n\n", appCopyright)
	fmt.Println("Usage: netinfo dns [options]")
	fmt.Println("\nOptions:")
	fmt.Println("  -help, -?            Show this help message")
	fmt.Println("  -search              Show only search domains")
	fmt.Println("  -servers             Show only DNS servers")
	fmt.Println("\nExamples:")
	fmt.Println("  netinfo dns                     # Show all DNS configuration")
	fmt.Println("  netinfo dns -servers             # Show only DNS servers")
	fmt.Println("  netinfo dns -search              # Show only search domains")
}

func showStatsHelp() {
	fmt.Printf("%s - Statistics Command\n", appName)
	fmt.Printf("%s\n\n", appCopyright)
	fmt.Println("Usage: netinfo stats [options]")
	fmt.Println("\nOptions:")
	fmt.Println("  -bytes               Show only byte counts")
	fmt.Println("  -errors              Show only error counts")
	fmt.Println("  -help, -?            Show this help message")
	fmt.Println("  -interface <name>    Filter by interface name")
	fmt.Println("  -packets             Show only packet counts")
	fmt.Println("\nExamples:")
	fmt.Println("  netinfo stats                    # Show statistics for all interfaces")
	fmt.Println("  netinfo stats -interface eth0    # Show stats for specific interface")
	fmt.Println("  netinfo stats -bytes             # Show only byte statistics")
	fmt.Println("  netinfo stats -errors            # Show only error statistics")
}

func showConnectionsHelp() {
	fmt.Printf("%s - Connections Command\n", appName)
	fmt.Printf("%s\n\n", appCopyright)
	fmt.Println("Usage: netinfo connections [options]")
	fmt.Println("\nOptions:")
	fmt.Println("  -established         Show only established connections")
	fmt.Println("  -help, -?            Show this help message")
	fmt.Println("  -interface <name>    Filter by interface")
	fmt.Println("  -listening           Show only listening ports")
	fmt.Println("  -tcp                 Show only TCP connections")
	fmt.Println("  -udp                 Show only UDP connections")
	fmt.Println("\nExamples:")
	fmt.Println("  netinfo connections              # Show all connections")
	fmt.Println("  netinfo connections -tcp          # Show only TCP connections")
	fmt.Println("  netinfo connections -listening    # Show only listening ports")
	fmt.Println("  netinfo connections -established  # Show only established connections")
}

func showARPHelp() {
	fmt.Printf("%s - ARP Command\n", appName)
	fmt.Printf("%s\n\n", appCopyright)
	fmt.Println("Usage: netinfo arp [options]")
	fmt.Println("\nOptions:")
	fmt.Println("  -help, -?            Show this help message")
	fmt.Println("  -interface <name>    Filter by interface name")
	fmt.Println("\nExamples:")
	fmt.Println("  netinfo arp                     # Show ARP table")
	fmt.Println("  netinfo arp -interface eth0     # Show ARP entries for specific interface")
}

func showPingHelp() {
	fmt.Printf("%s - Ping Command\n", appName)
	fmt.Printf("%s\n\n", appCopyright)
	fmt.Println("Usage: netinfo ping [options] <host>")
	fmt.Println("\nOptions:")
	fmt.Println("  -count <n>           Number of packets to send (default: 4)")
	fmt.Println("  -help, -?             Show this help message")
	fmt.Println("  -interval <sec>      Interval between packets in seconds (default: 1)")
	fmt.Println("  -ipv4                Use IPv4 only")
	fmt.Println("  -ipv6                Use IPv6 only")
	fmt.Println("  -timeout <sec>       Timeout in seconds (default: 1)")
	fmt.Println("\nExamples:")
	fmt.Println("  netinfo ping google.com         # Ping google.com")
	fmt.Println("  netinfo ping -count 10 8.8.8.8  # Send 10 pings to 8.8.8.8")
	fmt.Println("  netinfo ping -ipv4 google.com  # Ping using IPv4 only")
}

func showTraceHelp() {
	fmt.Printf("%s - Traceroute Command\n", appName)
	fmt.Printf("%s\n\n", appCopyright)
	fmt.Println("Usage: netinfo trace [options] <host>")
	fmt.Println("\nOptions:")
	fmt.Println("  -help, -?            Show this help message")
	fmt.Println("  -ipv4                Use IPv4 only")
	fmt.Println("  -ipv6                Use IPv6 only")
	fmt.Println("  -max-hops <n>        Maximum number of hops (default: 30)")
	fmt.Println("\nExamples:")
	fmt.Println("  netinfo trace google.com        # Traceroute to google.com")
	fmt.Println("  netinfo trace -max-hops 20 example.com  # Max 20 hops")
	fmt.Println("  netinfo trace -ipv4 google.com  # Traceroute using IPv4 only")
}

// DisplayWiFiList shows a summary list of WiFi adapters
func DisplayWiFiList(adapters []*WiFiInformation) {
	for i, adapter := range adapters {
		fmt.Printf("Adapter %d:\n", i+1)
		fmt.Printf("  Interface: %s\n", adapter.InterfaceName)
		if adapter.AdapterName != "" && adapter.AdapterName != adapter.InterfaceName {
			fmt.Printf("  Name: %s\n", adapter.AdapterName)
		}
		if adapter.Description != "" {
			fmt.Printf("  Description: %s\n", adapter.Description)
		}
		if adapter.MACAddress != "" {
			fmt.Printf("  MAC Address: %s\n", adapter.MACAddress)
		}
		if adapter.IPAddress != "" {
			fmt.Printf("  IP Address: %s\n", adapter.IPAddress)
		}
		if adapter.SSID != "" {
			fmt.Printf("  SSID: %s\n", adapter.SSID)
		}
		fmt.Println()
	}
}

// DisplayWiFiDetailed shows detailed information for WiFi adapters
func DisplayWiFiDetailed(adapters []*WiFiInformation) {
	for i, info := range adapters {
		if len(adapters) > 1 {
			fmt.Printf("=== Adapter %d ===\n", i+1)
		}

		if info.AdapterName != "" && info.AdapterName != info.InterfaceName {
			fmt.Printf("WiFi Adapter Name: %s\n", info.AdapterName)
		}
		if info.Description != "" {
			fmt.Printf("Description: %s\n", info.Description)
		}
		fmt.Printf("Interface: %s\n", info.InterfaceName)
		if info.MACAddress != "" {
			fmt.Printf("MAC Address: %s\n", info.MACAddress)
		}
		if info.IPAddress != "" {
			fmt.Printf("IP Address: %s\n", info.IPAddress)
		}

		if info.Channel > 0 {
			fmt.Printf("Channel: %d\n", info.Channel)
		}
		if info.ChannelWidth != "" && info.ChannelWidth != "Unknown" {
			fmt.Printf("Channel Width: %s\n", info.ChannelWidth)
		}
		if info.RadioType != "" && info.RadioType != "Unknown" {
			fmt.Printf("Network Type: %s\n", info.RadioType)
		}
		if info.RadioBand != "" && info.RadioBand != "Unknown" {
			fmt.Printf("Radio Band: %s\n", info.RadioBand)
		}
		if info.SSID != "" {
			fmt.Printf("SSID: %s\n", info.SSID)
		}
		if info.BSSID != "" {
			fmt.Printf("BSSID: %s\n", info.BSSID)
		}
		if info.RSSI != 0 {
			fmt.Printf("Signal Strength (RSSI): %d dBm\n", info.RSSI)
		}
		if info.SignalPercent > 0 {
			fmt.Printf("Signal Strength: %d%%\n", info.SignalPercent)
		}
		if info.TransmitRate > 0 {
			fmt.Printf("Transmit Rate: %d Mbps\n", info.TransmitRate)
		}
		if info.ReceiveRate > 0 {
			fmt.Printf("Receive Rate: %d Mbps\n", info.ReceiveRate)
		}
		if info.LinkSpeed > 0 && info.TransmitRate == 0 && info.ReceiveRate == 0 {
			fmt.Printf("Link Speed: %d Mbps\n", info.LinkSpeed)
		}
		if info.Security != "" && info.Security != "Unknown" {
			fmt.Printf("Security: %s\n", info.Security)
		}
		if info.NoiseLevel != 0 {
			fmt.Printf("Noise Level: %d dBm\n", info.NoiseLevel)
			if info.RSSI != 0 {
				snr := info.RSSI - info.NoiseLevel
				fmt.Printf("SNR: %d dB\n", snr)
			}
		}
		if info.TransmitPower >= 0 {
			fmt.Printf("Transmit Power: %d dBm\n", info.TransmitPower)
		}

		if len(adapters) > 1 && i < len(adapters)-1 {
			fmt.Println()
		}
	}
}

// handleWiFiCommand processes the wifi subcommand
func handleWiFiCommand(args []string) {
	fs := flag.NewFlagSet("wifi", flag.ExitOnError)
	listAll := fs.Bool("list", false, "List all WiFi adapters")
	interfaceFilter := fs.String("interface", "", "Filter by interface name")
	showHelpFlag := fs.Bool("help", false, "Show help message")
	showHelpAlt := fs.Bool("?", false, "Show help message")

	fs.Usage = showWiFiHelp
	fs.Parse(args)

	if *showHelpFlag || *showHelpAlt {
		showWiFiHelp()
		return
	}

	adapters, err := GetAllWiFiAdapters()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting WiFi adapters: %v\n", err)
		os.Exit(1)
	}

	if len(adapters) == 0 {
		fmt.Fprintf(os.Stderr, "No WiFi adapters found\n")
		os.Exit(1)
	}

	// Filter if requested
	var filteredAdapters []*WiFiInformation
	if *interfaceFilter != "" {
		for _, adapter := range adapters {
			if adapter.InterfaceName == *interfaceFilter {
				filteredAdapters = append(filteredAdapters, adapter)
				break
			}
		}
		if len(filteredAdapters) == 0 {
			fmt.Fprintf(os.Stderr, "No adapter found with interface name: %s\n", *interfaceFilter)
			os.Exit(1)
		}
	} else {
		filteredAdapters = adapters
	}

	// List mode
	if *listAll {
		DisplayWiFiList(filteredAdapters)
		return
	}

	// Display detailed info
	DisplayWiFiDetailed(filteredAdapters)
}

// DisplayWiredList shows a summary list of wired interfaces
func DisplayWiredList(interfaces []*WiredInterfaceInfo) {
	for i, iface := range interfaces {
		fmt.Printf("Interface %d:\n", i+1)
		fmt.Printf("  Interface: %s\n", iface.InterfaceName)
		if iface.AdapterName != "" && iface.AdapterName != iface.InterfaceName {
			fmt.Printf("  Name: %s\n", iface.AdapterName)
		}
		if iface.Description != "" {
			fmt.Printf("  Description: %s\n", iface.Description)
		}
		if iface.MACAddress != "" {
			fmt.Printf("  MAC Address: %s\n", iface.MACAddress)
		}
		if iface.IPAddress != "" {
			fmt.Printf("  IP Address: %s\n", iface.IPAddress)
		}
		if iface.Status != "" {
			fmt.Printf("  Status: %s\n", iface.Status)
		}
		fmt.Println()
	}
}

// DisplayWiredDetailed shows detailed information for wired interfaces
func DisplayWiredDetailed(interfaces []*WiredInterfaceInfo) {
	for i, iface := range interfaces {
		if len(interfaces) > 1 {
			fmt.Printf("=== Interface %d ===\n", i+1)
		}
		fmt.Printf("Interface: %s\n", iface.InterfaceName)
		if iface.AdapterName != "" && iface.AdapterName != iface.InterfaceName {
			fmt.Printf("Adapter Name: %s\n", iface.AdapterName)
		}
		if iface.Description != "" {
			fmt.Printf("Description: %s\n", iface.Description)
		}
		if iface.MACAddress != "" {
			fmt.Printf("MAC Address: %s\n", iface.MACAddress)
		}
		if iface.IPAddress != "" {
			fmt.Printf("IP Address: %s\n", iface.IPAddress)
		}
		if iface.IPv6Address != "" {
			fmt.Printf("IPv6 Address: %s\n", iface.IPv6Address)
		}
		if iface.Status != "" {
			fmt.Printf("Status: %s\n", iface.Status)
		}
		if iface.MTU > 0 {
			fmt.Printf("MTU: %d\n", iface.MTU)
		}
		if iface.Speed != "" {
			fmt.Printf("Speed: %s\n", iface.Speed)
		}
		if iface.Duplex != "" {
			fmt.Printf("Duplex: %s\n", iface.Duplex)
		}

		if len(interfaces) > 1 && i < len(interfaces)-1 {
			fmt.Println()
		}
	}
}

// handleWiredCommand processes the wired subcommand
func handleWiredCommand(args []string) {
	fs := flag.NewFlagSet("wired", flag.ExitOnError)
	listAll := fs.Bool("list", false, "List all wired interfaces")
	interfaceFilter := fs.String("interface", "", "Filter by interface name")
	activeOnly := fs.Bool("active", false, "Show only active interfaces")
	inactiveOnly := fs.Bool("inactive", false, "Show only inactive interfaces")
	showHelpFlag := fs.Bool("help", false, "Show help message")
	showHelpAlt := fs.Bool("?", false, "Show help message")

	fs.Usage = showWiredHelp
	fs.Parse(args)

	if *showHelpFlag || *showHelpAlt {
		showWiredHelp()
		return
	}

	interfaces, err := GetWiredInterfaces()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting wired interfaces: %v\n", err)
		os.Exit(1)
	}

	if len(interfaces) == 0 {
		fmt.Fprintf(os.Stderr, "No wired interfaces found\n")
		os.Exit(1)
	}

	// Filter interfaces
	var filteredInterfaces []*WiredInterfaceInfo
	for _, iface := range interfaces {
		if *interfaceFilter != "" && iface.InterfaceName != *interfaceFilter {
			continue
		}
		if *activeOnly && iface.Status != "up" {
			continue
		}
		if *inactiveOnly && iface.Status != "down" {
			continue
		}
		filteredInterfaces = append(filteredInterfaces, iface)
	}

	if len(filteredInterfaces) == 0 {
		fmt.Fprintf(os.Stderr, "No wired interfaces match the filter criteria\n")
		os.Exit(1)
	}

	// List mode
	if *listAll {
		DisplayWiredList(filteredInterfaces)
		return
	}

	// Display detailed info
	DisplayWiredDetailed(filteredInterfaces)
}

// DisplayInterfacesList shows a summary list of interfaces
func DisplayInterfacesList(interfaces []*InterfaceInfo) {
	for i, iface := range interfaces {
		fmt.Printf("Interface %d:\n", i+1)
		fmt.Printf("  Name: %s\n", iface.Name)
		fmt.Printf("  Type: %s\n", getInterfaceTypeDisplay(iface.Type))
		fmt.Printf("  Status: %s\n", iface.Status)
		if iface.HardwareAddr != "" {
			fmt.Printf("  MAC Address: %s\n", iface.HardwareAddr)
		}
		if iface.MTU > 0 {
			fmt.Printf("  MTU: %d\n", iface.MTU)
		}
		for _, ip := range iface.IPv4Addrs {
			fmt.Printf("  IPv4: %s\n", ip)
		}
		for _, ip := range iface.IPv6Addrs {
			fmt.Printf("  IPv6: %s\n", ip)
		}
		fmt.Println()
	}
}

// getInterfaceTypeDisplay returns a formatted display string for interface type
func getInterfaceTypeDisplay(ifaceType string) string {
	switch ifaceType {
	case "wired":
		return "Wired"
	case "wireless":
		return "Wireless"
	case "vpn":
		return "VPN"
	case "loopback":
		return "Loopback"
	default:
		return "Other"
	}
}

// DisplayInterfacesDetailed shows detailed information for interfaces
func DisplayInterfacesDetailed(interfaces []*InterfaceInfo) {
	for i, iface := range interfaces {
		if len(interfaces) > 1 {
			fmt.Printf("=== Interface %d ===\n", i+1)
		}
		fmt.Printf("Name: %s\n", iface.Name)
		fmt.Printf("Type: %s\n", getInterfaceTypeDisplay(iface.Type))
		fmt.Printf("Status: %s\n", iface.Status)
		if iface.HardwareAddr != "" {
			fmt.Printf("MAC Address: %s\n", iface.HardwareAddr)
		}
		if iface.MTU > 0 {
			fmt.Printf("MTU: %d\n", iface.MTU)
		}
		if iface.Flags != "" {
			fmt.Printf("Flags: %s\n", iface.Flags)
		}
		if len(iface.IPv4Addrs) > 0 {
			fmt.Printf("IPv4 Addresses:\n")
			for _, ip := range iface.IPv4Addrs {
				fmt.Printf("  %s\n", ip)
			}
		}
		if len(iface.IPv6Addrs) > 0 {
			fmt.Printf("IPv6 Addresses:\n")
			for _, ip := range iface.IPv6Addrs {
				fmt.Printf("  %s\n", ip)
			}
		}

		if len(interfaces) > 1 && i < len(interfaces)-1 {
			fmt.Println()
		}
	}
}

// handleInterfacesCommand processes the interfaces subcommand
func handleInterfacesCommand(args []string) {
	fs := flag.NewFlagSet("interfaces", flag.ExitOnError)
	listAll := fs.Bool("list", false, "List all interfaces")
	interfaceFilter := fs.String("interface", "", "Filter by interface name")
	upOnly := fs.Bool("up", false, "Show only interfaces that are up")
	downOnly := fs.Bool("down", false, "Show only interfaces that are down")
	activeOnly := fs.Bool("active", false, "Show only primary interfaces with active IP configs")
	showHelpFlag := fs.Bool("help", false, "Show help message")
	showHelpAlt := fs.Bool("?", false, "Show help message")

	fs.Usage = showInterfacesHelp
	fs.Parse(args)

	if *showHelpFlag || *showHelpAlt {
		showInterfacesHelp()
		return
	}

	interfaces, err := GetInterfaces()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting interfaces: %v\n", err)
		os.Exit(1)
	}

	if len(interfaces) == 0 {
		fmt.Fprintf(os.Stderr, "No interfaces found\n")
		os.Exit(1)
	}

	// Filter interfaces
	filteredInterfaces := FilterInterfaces(interfaces, *interfaceFilter, *upOnly, *downOnly, *activeOnly)

	if len(filteredInterfaces) == 0 {
		fmt.Fprintf(os.Stderr, "No interfaces match the filter criteria\n")
		os.Exit(1)
	}

	// List mode
	if *listAll {
		DisplayInterfacesList(filteredInterfaces)
		return
	}

	// Display detailed info
	DisplayInterfacesDetailed(filteredInterfaces)
}

// DisplayRoutes shows routing table information
func DisplayRoutes(routes []*RouteInfo) {
	for _, route := range routes {
		fmt.Printf("Destination: %s\n", route.Destination)
		if route.Gateway != "" && route.Gateway != "*" && route.Gateway != "0.0.0.0" {
			fmt.Printf("  Gateway: %s\n", route.Gateway)
		}
		if route.Interface != "" {
			fmt.Printf("  Interface: %s\n", route.Interface)
		}
		if route.Netmask != "" {
			fmt.Printf("  Netmask: %s\n", route.Netmask)
		}
		if route.Flags != "" {
			fmt.Printf("  Flags: %s\n", route.Flags)
		}
		fmt.Println()
	}
}

// handleRoutesCommand processes the routes subcommand
func handleRoutesCommand(args []string) {
	fs := flag.NewFlagSet("routes", flag.ExitOnError)
	defaultOnly := fs.Bool("default", false, "Show only default route")
	interfaceFilter := fs.String("interface", "", "Filter by interface")
	ipv4Only := fs.Bool("ipv4", false, "Show only IPv4 routes")
	ipv6Only := fs.Bool("ipv6", false, "Show only IPv6 routes")
	showHelpFlag := fs.Bool("help", false, "Show help message")
	showHelpAlt := fs.Bool("?", false, "Show help message")

	fs.Usage = showRoutesHelp
	fs.Parse(args)

	if *showHelpFlag || *showHelpAlt {
		showRoutesHelp()
		return
	}

	routes, err := GetRoutes()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting routes: %v\n", err)
		os.Exit(1)
	}

	if len(routes) == 0 {
		fmt.Fprintf(os.Stderr, "No routes found\n")
		os.Exit(1)
	}

	// Filter routes
	var filteredRoutes []*RouteInfo
	for _, route := range routes {
		if *defaultOnly && route.Destination != "default" && route.Destination != "0.0.0.0" && !strings.HasPrefix(route.Destination, "0.0.0.0/") {
			continue
		}
		if *interfaceFilter != "" && route.Interface != *interfaceFilter {
			continue
		}
		if *ipv4Only && strings.Contains(route.Destination, ":") {
			continue
		}
		if *ipv6Only && !strings.Contains(route.Destination, ":") {
			continue
		}
		filteredRoutes = append(filteredRoutes, route)
	}

	if len(filteredRoutes) == 0 {
		fmt.Fprintf(os.Stderr, "No routes match the filter criteria\n")
		os.Exit(1)
	}

	DisplayRoutes(filteredRoutes)
}

// DisplayDNS shows DNS configuration information
func DisplayDNS(dns *DNSInfo, serversOnly, searchOnly bool) {
	if serversOnly {
		fmt.Println("DNS Servers:")
		for _, server := range dns.Servers {
			fmt.Printf("  %s\n", server)
		}
	} else if searchOnly {
		fmt.Println("Search Domains:")
		for _, domain := range dns.SearchDomains {
			fmt.Printf("  %s\n", domain)
		}
	} else {
		if len(dns.Servers) > 0 {
			fmt.Println("DNS Servers:")
			for _, server := range dns.Servers {
				fmt.Printf("  %s\n", server)
			}
		}
		if len(dns.SearchDomains) > 0 {
			fmt.Println("\nSearch Domains:")
			for _, domain := range dns.SearchDomains {
				fmt.Printf("  %s\n", domain)
			}
		}
	}
}

// handleDNSCommand processes the dns subcommand
func handleDNSCommand(args []string) {
	fs := flag.NewFlagSet("dns", flag.ExitOnError)
	serversOnly := fs.Bool("servers", false, "Show only DNS servers")
	searchOnly := fs.Bool("search", false, "Show only search domains")
	showHelpFlag := fs.Bool("help", false, "Show help message")
	showHelpAlt := fs.Bool("?", false, "Show help message")

	fs.Usage = showDNSHelp
	fs.Parse(args)

	if *showHelpFlag || *showHelpAlt {
		showDNSHelp()
		return
	}

	dns, err := GetDNS()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting DNS configuration: %v\n", err)
		os.Exit(1)
	}

	DisplayDNS(dns, *serversOnly, *searchOnly)
}

// DisplayStats shows network interface statistics
func DisplayStats(stats []*InterfaceStats, bytesOnly, packetsOnly, errorsOnly bool, ifaceFilter string) {
	for _, stat := range stats {
		if ifaceFilter != "" && stat.Interface != ifaceFilter {
			continue
		}

		fmt.Printf("Interface: %s\n", stat.Interface)

		if bytesOnly || (!packetsOnly && !errorsOnly) {
			fmt.Printf("  Bytes Received: %d\n", stat.BytesReceived)
			fmt.Printf("  Bytes Sent: %d\n", stat.BytesSent)
		}

		if packetsOnly || (!bytesOnly && !errorsOnly) {
			fmt.Printf("  Packets Received: %d\n", stat.PacketsReceived)
			fmt.Printf("  Packets Sent: %d\n", stat.PacketsSent)
		}

		if errorsOnly || (!bytesOnly && !packetsOnly) {
			fmt.Printf("  Errors Received: %d\n", stat.ErrorsReceived)
			fmt.Printf("  Errors Sent: %d\n", stat.ErrorsSent)
			fmt.Printf("  Drops Received: %d\n", stat.DropsReceived)
			fmt.Printf("  Drops Sent: %d\n", stat.DropsSent)
		}

		fmt.Println()
	}
}

// handleStatsCommand processes the stats subcommand
func handleStatsCommand(args []string) {
	fs := flag.NewFlagSet("stats", flag.ExitOnError)
	interfaceFilter := fs.String("interface", "", "Filter by interface name")
	bytesOnly := fs.Bool("bytes", false, "Show only byte counts")
	packetsOnly := fs.Bool("packets", false, "Show only packet counts")
	errorsOnly := fs.Bool("errors", false, "Show only error counts")
	showHelpFlag := fs.Bool("help", false, "Show help message")
	showHelpAlt := fs.Bool("?", false, "Show help message")

	fs.Usage = showStatsHelp
	fs.Parse(args)

	if *showHelpFlag || *showHelpAlt {
		showStatsHelp()
		return
	}

	stats, err := GetInterfaceStats()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting statistics: %v\n", err)
		os.Exit(1)
	}

	if len(stats) == 0 {
		fmt.Fprintf(os.Stderr, "No statistics available\n")
		os.Exit(1)
	}

	DisplayStats(stats, *bytesOnly, *packetsOnly, *errorsOnly, *interfaceFilter)
}

// DisplayConnections shows network connections
func DisplayConnections(connections []*Connection, tcpOnly, udpOnly, listeningOnly, establishedOnly bool, ifaceFilter string) {
	for _, conn := range connections {
		if tcpOnly && conn.Protocol != "tcp" {
			continue
		}
		if udpOnly && conn.Protocol != "udp" {
			continue
		}
		if listeningOnly && conn.State != "LISTEN" && conn.State != "LISTENING" {
			continue
		}
		if establishedOnly && conn.State != "ESTABLISHED" {
			continue
		}
		if ifaceFilter != "" && conn.Interface != ifaceFilter {
			continue
		}

		fmt.Printf("%s", strings.ToUpper(conn.Protocol))
		if conn.LocalAddr != "" {
			fmt.Printf(" %s:%s", conn.LocalAddr, conn.LocalPort)
		}
		if conn.RemoteAddr != "" && conn.RemoteAddr != "*" {
			fmt.Printf(" -> %s:%s", conn.RemoteAddr, conn.RemotePort)
		}
		if conn.State != "" {
			fmt.Printf(" [%s]", conn.State)
		}
		fmt.Println()
	}
}

// handleConnectionsCommand processes the connections subcommand
func handleConnectionsCommand(args []string) {
	fs := flag.NewFlagSet("connections", flag.ExitOnError)
	tcpOnly := fs.Bool("tcp", false, "Show only TCP connections")
	udpOnly := fs.Bool("udp", false, "Show only UDP connections")
	listeningOnly := fs.Bool("listening", false, "Show only listening ports")
	establishedOnly := fs.Bool("established", false, "Show only established connections")
	interfaceFilter := fs.String("interface", "", "Filter by interface")
	showHelpFlag := fs.Bool("help", false, "Show help message")
	showHelpAlt := fs.Bool("?", false, "Show help message")

	fs.Usage = showConnectionsHelp
	fs.Parse(args)

	if *showHelpFlag || *showHelpAlt {
		showConnectionsHelp()
		return
	}

	connections, err := GetConnections()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting connections: %v\n", err)
		os.Exit(1)
	}

	if len(connections) == 0 {
		fmt.Fprintf(os.Stderr, "No connections found\n")
		os.Exit(1)
	}

	DisplayConnections(connections, *tcpOnly, *udpOnly, *listeningOnly, *establishedOnly, *interfaceFilter)
}

// DisplayARP shows ARP table entries
func DisplayARP(entries []*ARPEntry, ifaceFilter string) {
	for _, entry := range entries {
		if ifaceFilter != "" && entry.Interface != ifaceFilter {
			continue
		}

		fmt.Printf("IP Address: %s\n", entry.IPAddress)
		if entry.MACAddress != "" {
			fmt.Printf("  MAC Address: %s\n", entry.MACAddress)
		}
		if entry.Interface != "" {
			fmt.Printf("  Interface: %s\n", entry.Interface)
		}
		if entry.Type != "" {
			fmt.Printf("  Type: %s\n", entry.Type)
		}
		fmt.Println()
	}
}

// handleARPCommand processes the arp subcommand
func handleARPCommand(args []string) {
	fs := flag.NewFlagSet("arp", flag.ExitOnError)
	interfaceFilter := fs.String("interface", "", "Filter by interface name")
	showHelpFlag := fs.Bool("help", false, "Show help message")
	showHelpAlt := fs.Bool("?", false, "Show help message")

	fs.Usage = showARPHelp
	fs.Parse(args)

	if *showHelpFlag || *showHelpAlt {
		showARPHelp()
		return
	}

	entries, err := GetARP()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting ARP table: %v\n", err)
		os.Exit(1)
	}

	if len(entries) == 0 {
		fmt.Fprintf(os.Stderr, "No ARP entries found\n")
		os.Exit(1)
	}

	DisplayARP(entries, *interfaceFilter)
}

// handlePingCommand processes the ping subcommand
func handlePingCommand(args []string) {
	fs := flag.NewFlagSet("ping", flag.ExitOnError)
	count := fs.Int("count", 4, "Number of packets to send")
	interval := fs.Float64("interval", 1.0, "Interval between packets in seconds")
	timeout := fs.Float64("timeout", 1.0, "Timeout in seconds")
	_ = fs.Bool("ipv4", false, "Use IPv4 only") // TODO: Implement IPv4/IPv6 filtering
	_ = fs.Bool("ipv6", false, "Use IPv6 only")
	showHelpFlag := fs.Bool("help", false, "Show help message")
	showHelpAlt := fs.Bool("?", false, "Show help message")

	fs.Usage = showPingHelp
	fs.Parse(args)

	if *showHelpFlag || *showHelpAlt {
		showPingHelp()
		return
	}

	if fs.NArg() == 0 {
		fmt.Fprintf(os.Stderr, "Error: host is required\n")
		showPingHelp()
		os.Exit(1)
	}

	host := fs.Arg(0)

	// Use system ping command for now
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		// Windows ping syntax
		cmd = exec.Command("ping", "-n", strconv.Itoa(*count), "-w", strconv.Itoa(int(*timeout*1000)), host)
	} else {
		// Unix-like ping syntax
		cmd = exec.Command("ping", "-c", strconv.Itoa(*count), "-i", strconv.FormatFloat(*interval, 'f', 1, 64), "-W", strconv.FormatFloat(*timeout, 'f', 1, 64), host)
	}

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		os.Exit(1)
	}
}

// handleTraceCommand processes the trace subcommand
func handleTraceCommand(args []string) {
	fs := flag.NewFlagSet("trace", flag.ExitOnError)
	maxHops := fs.Int("max-hops", 30, "Maximum number of hops")
	_ = fs.Bool("ipv4", false, "Use IPv4 only") // TODO: Implement IPv4/IPv6 filtering
	_ = fs.Bool("ipv6", false, "Use IPv6 only")
	showHelpFlag := fs.Bool("help", false, "Show help message")
	showHelpAlt := fs.Bool("?", false, "Show help message")

	fs.Usage = showTraceHelp
	fs.Parse(args)

	if *showHelpFlag || *showHelpAlt {
		showTraceHelp()
		return
	}

	if fs.NArg() == 0 {
		fmt.Fprintf(os.Stderr, "Error: host is required\n")
		showTraceHelp()
		os.Exit(1)
	}

	host := fs.Arg(0)

	// Use system traceroute command
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		// Windows uses tracert
		cmd = exec.Command("tracert", "-h", strconv.Itoa(*maxHops), host)
	} else {
		// Unix-like uses traceroute
		cmd = exec.Command("traceroute", "-m", strconv.Itoa(*maxHops), host)
	}

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		os.Exit(1)
	}
}

// getLatestReleaseVersion fetches the latest release version from GitHub
func getLatestReleaseVersion(owner, repo string) string {
	client := http.Client{
		Timeout: 5 * time.Second,
	}

	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", owner, repo)
	resp, err := client.Get(url)
	if err != nil {
		return "unknown"
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "unknown"
	}

	var apiResponse struct {
		TagName string `json:"tag_name"`
	}
	if err := json.Unmarshal(body, &apiResponse); err != nil {
		return "unknown"
	}

	return apiResponse.TagName
}

// checkUpdate checks for available updates using the update checker module
func checkUpdate() {
	// TODO: Update these values with your actual GitHub repository information
	owner := "amarillier"       // GitHub username/organization
	repo := "KrankyBearNetInfo" // Repository name
	software := appName
	downloadLink := ""   // Empty string will default to GitHub releases page
	minDaysInterval := 0 // Check every time (0 = always check)
	verbose := false     // Set to true for debug output

	// Get latest release version for display
	latestVersion := getLatestReleaseVersion(owner, repo)

	checker := updatechecker.New(owner, repo, software, downloadLink, minDaysInterval, verbose)
	checker.CheckForUpdate(appVersion)

	fmt.Printf("Current version: %s\n", appVersion)
	if latestVersion != "unknown" {
		fmt.Printf("Latest release version: %s\n\n", latestVersion)
	} else {
		fmt.Println()
	}
	checker.PrintMessage()
}

func main() {
	// Check for update check flags first (before subcommand parsing)
	if len(os.Args) >= 2 {
		if os.Args[1] == "-checkupdate" || os.Args[1] == "-cu" {
			checkUpdate()
			return
		}
	}

	// Check if we have a subcommand
	if len(os.Args) < 2 {
		// Default to showing main help if no command provided
		showMainHelp()
		return
	}

	// Parse subcommand
	command := os.Args[1]
	args := os.Args[2:]

	// Handle help flags before subcommand
	if command == "help" || command == "-help" || command == "-h" || command == "-?" {
		showMainHelp()
		return
	}

	switch command {
	case "wifi":
		handleWiFiCommand(args)
	case "wired":
		handleWiredCommand(args)
	case "interfaces":
		handleInterfacesCommand(args)
	case "routes":
		handleRoutesCommand(args)
	case "dns":
		handleDNSCommand(args)
	case "stats":
		handleStatsCommand(args)
	case "connections":
		handleConnectionsCommand(args)
	case "arp":
		handleARPCommand(args)
	case "ping":
		handlePingCommand(args)
	case "trace":
		handleTraceCommand(args)
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n\n", command)
		showMainHelp()
		os.Exit(1)
	}
}

// "Now this is not the end. It is not even the beginning of the end. But it is, perhaps, the end of the beginning." Winston Churchill, November 10, 1942

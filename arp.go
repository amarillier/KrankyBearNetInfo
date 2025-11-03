package main

import (
	"os/exec"
	"runtime"
	"strings"
)

// ARPEntry represents an ARP table entry
type ARPEntry struct {
	IPAddress  string
	MACAddress string
	Interface  string
	Type       string // "static", "dynamic", etc.
}

// GetARP retrieves ARP table information
func GetARP() ([]*ARPEntry, error) {
	if runtime.GOOS == "windows" {
		return getARPWindows()
	} else if runtime.GOOS == "darwin" {
		return getARPDarwin()
	} else if runtime.GOOS == "linux" {
		return getARPLinux()
	}
	return nil, nil
}

// getARPLinux gets ARP table on Linux using `ip neigh` or `arp -a`
func getARPLinux() ([]*ARPEntry, error) {
	// Try `ip neigh` first (modern Linux)
	cmd := exec.Command("ip", "neigh")
	output, err := cmd.Output()
	if err == nil {
		return parseARPLinux(string(output))
	}

	// Fallback to `arp -a`
	cmd = exec.Command("arp", "-a")
	output, err = cmd.Output()
	if err != nil {
		return nil, err
	}

	return parseARPLinuxLegacy(string(output))
}

// parseARPLinux parses output from `ip neigh`
func parseARPLinux(output string) ([]*ARPEntry, error) {
	var entries []*ARPEntry
	lines := strings.Split(output, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 4 {
			continue
		}

		entry := &ARPEntry{
			IPAddress:  fields[0],
			MACAddress: fields[4],
			Interface:  fields[2],
		}

		if len(fields) > 5 && fields[5] == "PERMANENT" {
			entry.Type = "static"
		} else {
			entry.Type = "dynamic"
		}

		entries = append(entries, entry)
	}

	return entries, nil
}

// parseARPLinuxLegacy parses output from `arp -a`
func parseARPLinuxLegacy(output string) ([]*ARPEntry, error) {
	var entries []*ARPEntry
	lines := strings.Split(output, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Format: hostname (192.168.1.1) at aa:bb:cc:dd:ee:ff [ether] on eth0
		parts := strings.Split(line, " at ")
		if len(parts) < 2 {
			continue
		}

		// Extract IP
		ipPart := parts[0]
		ip := ""
		if idx := strings.Index(ipPart, "("); idx >= 0 {
			if idx2 := strings.Index(ipPart, ")"); idx2 > idx {
				ip = ipPart[idx+1 : idx2]
			}
		}

		// Extract MAC and interface
		rest := parts[1]
		fields := strings.Fields(rest)
		if len(fields) < 2 {
			continue
		}

		mac := fields[0]
		iface := ""
		if len(fields) >= 3 && fields[len(fields)-2] == "on" {
			iface = fields[len(fields)-1]
		}

		entry := &ARPEntry{
			IPAddress:  ip,
			MACAddress: mac,
			Interface:  iface,
			Type:       "dynamic",
		}

		if strings.Contains(rest, "PERMANENT") {
			entry.Type = "static"
		}

		entries = append(entries, entry)
	}

	return entries, nil
}

// getARPDarwin gets ARP table on macOS using `arp -a`
func getARPDarwin() ([]*ARPEntry, error) {
	cmd := exec.Command("arp", "-a")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	return parseARPDarwin(string(output))
}

// parseARPDarwin parses output from `arp -a` on macOS
func parseARPDarwin(output string) ([]*ARPEntry, error) {
	var entries []*ARPEntry
	lines := strings.Split(output, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Format: hostname (192.168.1.1) at aa:bb:cc:dd:ee:ff on en0 ifscope [ethernet]
		parts := strings.Split(line, " at ")
		if len(parts) < 2 {
			continue
		}

		// Extract IP
		ipPart := parts[0]
		ip := ""
		if idx := strings.Index(ipPart, "("); idx >= 0 {
			if idx2 := strings.Index(ipPart, ")"); idx2 > idx {
				ip = ipPart[idx+1 : idx2]
			}
		}

		// Extract MAC and interface
		rest := parts[1]
		fields := strings.Fields(rest)
		if len(fields) < 3 {
			continue
		}

		mac := fields[0]
		iface := ""
		if len(fields) >= 2 && fields[1] == "on" && len(fields) >= 3 {
			iface = fields[2]
		}

		entry := &ARPEntry{
			IPAddress:  ip,
			MACAddress: mac,
			Interface:  iface,
			Type:       "dynamic",
		}

		entries = append(entries, entry)
	}

	return entries, nil
}

// getARPWindows gets ARP table on Windows using `arp -a`
func getARPWindows() ([]*ARPEntry, error) {
	cmd := exec.Command("arp", "-a")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	return parseARPWindows(string(output))
}

// parseARPWindows parses output from `arp -a` on Windows
func parseARPWindows(output string) ([]*ARPEntry, error) {
	var entries []*ARPEntry
	lines := strings.Split(output, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "Interface:") {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 3 {
			continue
		}

		ip := fields[0]
		mac := fields[1]
		arpType := fields[2]

		entry := &ARPEntry{
			IPAddress:  ip,
			MACAddress: mac,
			Type:       arpType,
		}

		entries = append(entries, entry)
	}

	return entries, nil
}

// "Now this is not the end. It is not even the beginning of the end. But it is, perhaps, the end of the beginning." Winston Churchill, November 10, 1942

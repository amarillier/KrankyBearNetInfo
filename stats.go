package main

import (
	"bufio"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
)

// InterfaceStats represents network interface statistics
type InterfaceStats struct {
	Interface       string
	BytesReceived   uint64
	BytesSent       uint64
	PacketsReceived uint64
	PacketsSent     uint64
	ErrorsReceived  uint64
	ErrorsSent      uint64
	DropsReceived   uint64
	DropsSent       uint64
}

// GetInterfaceStats retrieves statistics for network interfaces
func GetInterfaceStats() ([]*InterfaceStats, error) {
	if runtime.GOOS == "windows" {
		return getStatsWindows()
	} else if runtime.GOOS == "darwin" {
		return getStatsDarwin()
	} else if runtime.GOOS == "linux" {
		return getStatsLinux()
	}
	return nil, nil
}

// getStatsLinux gets statistics from /proc/net/dev
func getStatsLinux() ([]*InterfaceStats, error) {
	file, err := os.Open("/proc/net/dev")
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var stats []*InterfaceStats
	scanner := bufio.NewScanner(file)

	// Skip header lines
	scanner.Scan()
	scanner.Scan()

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		// Format: interface: bytes_recv packets_recv errs_recv drop_recv ... bytes_sent packets_sent errs_sent drop_sent ...
		parts := strings.Split(line, ":")
		if len(parts) < 2 {
			continue
		}

		iface := strings.TrimSpace(parts[0])
		data := strings.Fields(parts[1])

		if len(data) < 16 {
			continue
		}

		stat := &InterfaceStats{Interface: iface}

		// Parse received stats
		stat.BytesReceived, _ = strconv.ParseUint(data[0], 10, 64)
		stat.PacketsReceived, _ = strconv.ParseUint(data[1], 10, 64)
		stat.ErrorsReceived, _ = strconv.ParseUint(data[2], 10, 64)
		stat.DropsReceived, _ = strconv.ParseUint(data[3], 10, 64)

		// Parse sent stats
		stat.BytesSent, _ = strconv.ParseUint(data[8], 10, 64)
		stat.PacketsSent, _ = strconv.ParseUint(data[9], 10, 64)
		stat.ErrorsSent, _ = strconv.ParseUint(data[10], 10, 64)
		stat.DropsSent, _ = strconv.ParseUint(data[11], 10, 64)

		stats = append(stats, stat)
	}

	return stats, scanner.Err()
}

// getStatsDarwin gets statistics using netstat on macOS
func getStatsDarwin() ([]*InterfaceStats, error) {
	cmd := exec.Command("netstat", "-ibn")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	return parseStatsDarwin(string(output))
}

// parseStatsDarwin parses netstat -ibn output
func parseStatsDarwin(output string) ([]*InterfaceStats, error) {
	var stats []*InterfaceStats
	statsMap := make(map[string]*InterfaceStats)

	lines := strings.Split(output, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "Name") {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 7 {
			continue
		}

		iface := fields[0]
		if statsMap[iface] == nil {
			statsMap[iface] = &InterfaceStats{Interface: iface}
		}

		stat := statsMap[iface]

		// Parse based on field position
		if strings.Contains(line, "Link") || strings.Contains(line, "Ibytes") {
			// Interface statistics line
			if len(fields) >= 6 {
				stat.BytesReceived, _ = strconv.ParseUint(fields[4], 10, 64)
				stat.BytesSent, _ = strconv.ParseUint(fields[5], 10, 64)
			}
		} else {
			// Try to parse as numeric stats
			for i, field := range fields {
				if i >= 1 {
					if val, err := strconv.ParseUint(field, 10, 64); err == nil {
						switch i {
						case 4:
							stat.BytesReceived = val
						case 5:
							stat.BytesSent = val
						}
					}
				}
			}
		}
	}

	for _, stat := range statsMap {
		stats = append(stats, stat)
	}

	return stats, nil
}

// getStatsWindows gets statistics using netstat on Windows
func getStatsWindows() ([]*InterfaceStats, error) {
	cmd := exec.Command("netstat", "-e")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	return parseStatsWindows(string(output))
}

// parseStatsWindows parses netstat -e output
func parseStatsWindows(output string) ([]*InterfaceStats, error) {
	var stats []*InterfaceStats
	lines := strings.Split(output, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.Contains(line, "Bytes") || strings.Contains(line, "Interface") {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) >= 3 {
			stat := &InterfaceStats{
				Interface:     "default",
				BytesReceived: 0,
				BytesSent:     0,
			}

			if len(fields) >= 2 {
				stat.BytesReceived, _ = strconv.ParseUint(fields[0], 10, 64)
				stat.BytesSent, _ = strconv.ParseUint(fields[1], 10, 64)
			}

			stats = append(stats, stat)
		}
	}

	return stats, nil
}

// "Now this is not the end. It is not even the beginning of the end. But it is, perhaps, the end of the beginning." Winston Churchill, November 10, 1942

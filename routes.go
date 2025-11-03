package main

import (
	"os/exec"
	"runtime"
	"strings"
)

// GetRoutes retrieves routing table information
// This is a cross-platform stub - platform-specific implementations should override
func GetRoutes() ([]*RouteInfo, error) {
	if runtime.GOOS == "windows" {
		return getRoutesWindows()
	} else if runtime.GOOS == "darwin" {
		return getRoutesDarwin()
	} else if runtime.GOOS == "linux" {
		return getRoutesLinux()
	}
	return nil, nil
}

// getRoutesLinux gets routes on Linux using `ip route` or `route -n`
func getRoutesLinux() ([]*RouteInfo, error) {
	var cmd *exec.Cmd
	
	// Try `ip route` first (modern Linux)
	cmd = exec.Command("ip", "route")
	output, err := cmd.Output()
	if err != nil {
		// Fallback to `route -n`
		cmd = exec.Command("route", "-n")
		output, err = cmd.Output()
		if err != nil {
			return nil, err
		}
		return parseRoutesLinuxLegacy(string(output))
	}
	
	return parseRoutesLinux(string(output))
}

// parseRoutesLinux parses output from `ip route`
func parseRoutesLinux(output string) ([]*RouteInfo, error) {
	var routes []*RouteInfo
	lines := strings.Split(output, "\n")
	
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		
		route := &RouteInfo{}
		fields := strings.Fields(line)
		
		if len(fields) < 2 {
			continue
		}
		
		// Parse destination
		if fields[0] == "default" {
			route.Destination = "0.0.0.0/0"
		} else {
			route.Destination = fields[0]
		}
		
		// Find gateway and interface
		for i, field := range fields {
			if field == "via" && i+1 < len(fields) {
				route.Gateway = fields[i+1]
			}
			if field == "dev" && i+1 < len(fields) {
				route.Interface = fields[i+1]
			}
		}
		
		routes = append(routes, route)
	}
	
	return routes, nil
}

// parseRoutesLinuxLegacy parses output from `route -n`
func parseRoutesLinuxLegacy(output string) ([]*RouteInfo, error) {
	var routes []*RouteInfo
	lines := strings.Split(output, "\n")
	
	// Skip header line
	if len(lines) < 2 {
		return routes, nil
	}
	
	for i := 1; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])
		if line == "" {
			continue
		}
		
		fields := strings.Fields(line)
		if len(fields) < 8 {
			continue
		}
		
		route := &RouteInfo{
			Destination: fields[0],
			Gateway:     fields[1],
			Netmask:     fields[2],
			Flags:       fields[3],
			Interface:   fields[7],
		}
		
		routes = append(routes, route)
	}
	
	return routes, nil
}

// getRoutesDarwin gets routes on macOS using `netstat -rn`
func getRoutesDarwin() ([]*RouteInfo, error) {
	cmd := exec.Command("netstat", "-rn")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	
	return parseRoutesDarwin(string(output))
}

// parseRoutesDarwin parses output from `netstat -rn`
func parseRoutesDarwin(output string) ([]*RouteInfo, error) {
	var routes []*RouteInfo
	lines := strings.Split(output, "\n")
	
	// Skip header lines
	skipLines := 4
	for i, line := range lines {
		if i < skipLines {
			continue
		}
		
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		
		fields := strings.Fields(line)
		if len(fields) < 4 {
			continue
		}
		
		route := &RouteInfo{
			Destination: fields[0],
			Gateway:     fields[1],
			Flags:       fields[2],
			Interface:   fields[3],
		}
		
		if len(fields) > 4 {
			route.Netmask = fields[4]
		}
		
		routes = append(routes, route)
	}
	
	return routes, nil
}

// getRoutesWindows gets routes on Windows using `route print`
func getRoutesWindows() ([]*RouteInfo, error) {
	cmd := exec.Command("route", "print")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	
	return parseRoutesWindows(string(output))
}

// parseRoutesWindows parses output from `route print`
func parseRoutesWindows(output string) ([]*RouteInfo, error) {
	var routes []*RouteInfo
	lines := strings.Split(output, "\n")
	
	inTable := false
	for _, line := range lines {
		line = strings.TrimSpace(line)
		
		// Look for IPv4 Route Table header
		if strings.Contains(line, "IPv4 Route Table") {
			inTable = true
			continue
		}
		
		if !inTable {
			continue
		}
		
		// Skip separator lines
		if strings.HasPrefix(line, "===") || line == "" {
			continue
		}
		
		fields := strings.Fields(line)
		if len(fields) < 5 {
			continue
		}
		
		route := &RouteInfo{
			Destination: fields[0],
			Netmask:     fields[1],
			Gateway:     fields[2],
			Interface:   fields[3],
			Flags:       fields[4],
		}
		
		routes = append(routes, route)
	}
	
	return routes, nil
}


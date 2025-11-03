package main

import (
	"os/exec"
	"runtime"
	"strings"
)

// Connection represents a network connection
type Connection struct {
	Protocol   string // "tcp", "udp", etc.
	LocalAddr  string
	LocalPort  string
	RemoteAddr string
	RemotePort string
	State      string // "ESTABLISHED", "LISTEN", etc.
	Interface  string
}

// GetConnections retrieves active network connections
func GetConnections() ([]*Connection, error) {
	if runtime.GOOS == "windows" {
		return getConnectionsWindows()
	} else if runtime.GOOS == "darwin" {
		return getConnectionsDarwin()
	} else if runtime.GOOS == "linux" {
		return getConnectionsLinux()
	}
	return nil, nil
}

// getConnectionsLinux gets connections using `ss` or `netstat`
func getConnectionsLinux() ([]*Connection, error) {
	var cmd *exec.Cmd

	// Try `ss` first (modern Linux)
	cmd = exec.Command("ss", "-tunp")
	output, err := cmd.Output()
	if err == nil {
		return parseConnectionsSS(string(output))
	}

	// Fallback to `netstat`
	cmd = exec.Command("netstat", "-tunp")
	output, err = cmd.Output()
	if err != nil {
		return nil, err
	}

	return parseConnectionsNetstat(string(output))
}

// parseConnectionsSS parses `ss` output
func parseConnectionsSS(output string) ([]*Connection, error) {
	var connections []*Connection
	lines := strings.Split(output, "\n")

	// Skip header
	for i := 1; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])
		if line == "" {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 5 {
			continue
		}

		conn := &Connection{}

		// Parse protocol
		proto := strings.ToLower(fields[0])
		if strings.Contains(proto, "tcp") {
			conn.Protocol = "tcp"
		} else if strings.Contains(proto, "udp") {
			conn.Protocol = "udp"
		}

		// Parse local address:port
		local := fields[4]
		parts := strings.Split(local, ":")
		if len(parts) >= 2 {
			conn.LocalAddr = parts[0]
			conn.LocalPort = parts[1]
		}

		// Parse remote address:port (if exists)
		if len(fields) > 5 {
			remote := fields[5]
			parts := strings.Split(remote, ":")
			if len(parts) >= 2 {
				conn.RemoteAddr = parts[0]
				conn.RemotePort = parts[1]
			}
		}

		// Parse state
		if len(fields) > 6 {
			conn.State = fields[6]
		}

		connections = append(connections, conn)
	}

	return connections, nil
}

// parseConnectionsNetstat parses `netstat` output
func parseConnectionsNetstat(output string) ([]*Connection, error) {
	var connections []*Connection
	lines := strings.Split(output, "\n")

	// Skip header
	for i := 2; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])
		if line == "" {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 6 {
			continue
		}

		conn := &Connection{
			Protocol: strings.ToLower(fields[0]),
		}

		// Parse local address:port
		local := fields[3]
		parts := strings.Split(local, ":")
		if len(parts) >= 2 {
			conn.LocalAddr = parts[0]
			conn.LocalPort = parts[1]
		}

		// Parse remote address:port
		remote := fields[4]
		parts = strings.Split(remote, ":")
		if len(parts) >= 2 {
			conn.RemoteAddr = parts[0]
			conn.RemotePort = parts[1]
		}

		// Parse state
		if len(fields) > 5 {
			conn.State = fields[5]
		}

		connections = append(connections, conn)
	}

	return connections, nil
}

// getConnectionsDarwin gets connections using `netstat` on macOS
func getConnectionsDarwin() ([]*Connection, error) {
	cmd := exec.Command("netstat", "-an")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	return parseConnectionsDarwin(string(output))
}

// parseConnectionsDarwin parses netstat output on macOS
func parseConnectionsDarwin(output string) ([]*Connection, error) {
	var connections []*Connection
	lines := strings.Split(output, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "Active") || strings.HasPrefix(line, "Proto") {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 4 {
			continue
		}

		conn := &Connection{
			Protocol: strings.ToLower(fields[0]),
		}

		// Parse local address:port
		local := fields[3]
		parts := strings.Split(local, ".")
		if len(parts) >= 2 {
			conn.LocalAddr = strings.Join(parts[:len(parts)-1], ".")
			conn.LocalPort = parts[len(parts)-1]
		}

		// Parse remote address:port
		if len(fields) > 4 {
			remote := fields[4]
			parts := strings.Split(remote, ".")
			if len(parts) >= 2 {
				conn.RemoteAddr = strings.Join(parts[:len(parts)-1], ".")
				conn.RemotePort = parts[len(parts)-1]
			}
		}

		// Parse state
		if len(fields) > 5 {
			conn.State = fields[5]
		}

		connections = append(connections, conn)
	}

	return connections, nil
}

// getConnectionsWindows gets connections using `netstat` on Windows
func getConnectionsWindows() ([]*Connection, error) {
	cmd := exec.Command("netstat", "-an")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	return parseConnectionsWindows(string(output))
}

// parseConnectionsWindows parses netstat output on Windows
func parseConnectionsWindows(output string) ([]*Connection, error) {
	var connections []*Connection
	lines := strings.Split(output, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "Active") || strings.HasPrefix(line, "Proto") {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 4 {
			continue
		}

		conn := &Connection{
			Protocol: strings.ToLower(fields[0]),
		}

		// Parse local address:port
		local := fields[1]
		parts := strings.Split(local, ":")
		if len(parts) >= 2 {
			conn.LocalAddr = parts[0]
			conn.LocalPort = parts[1]
		}

		// Parse remote address:port
		remote := fields[2]
		parts = strings.Split(remote, ":")
		if len(parts) >= 2 {
			conn.RemoteAddr = parts[0]
			conn.RemotePort = parts[1]
		}

		// Parse state
		if len(fields) > 3 {
			conn.State = fields[3]
		}

		connections = append(connections, conn)
	}

	return connections, nil
}

// "Now this is not the end. It is not even the beginning of the end. But it is, perhaps, the end of the beginning." Winston Churchill, November 10, 1942

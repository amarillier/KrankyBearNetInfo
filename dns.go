package main

import (
	"bufio"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

// GetDNS retrieves DNS configuration information
func GetDNS() (*DNSInfo, error) {
	if runtime.GOOS == "windows" {
		return getDNSWindows()
	} else if runtime.GOOS == "darwin" {
		return getDNSDarwin()
	} else if runtime.GOOS == "linux" {
		return getDNSLinux()
	}
	return &DNSInfo{}, nil
}

// getDNSLinux gets DNS servers from /etc/resolv.conf
func getDNSLinux() (*DNSInfo, error) {
	dns := &DNSInfo{}
	
	file, err := os.Open("/etc/resolv.conf")
	if err != nil {
		return nil, err
	}
	defer file.Close()
	
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "nameserver") {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				dns.Servers = append(dns.Servers, fields[1])
			}
		} else if strings.HasPrefix(line, "search") {
			fields := strings.Fields(line)
			if len(fields) > 1 {
				dns.SearchDomains = append(dns.SearchDomains, fields[1:]...)
			}
		}
	}
	
	return dns, scanner.Err()
}

// getDNSDarwin gets DNS servers on macOS using scutil
func getDNSDarwin() (*DNSInfo, error) {
	dns := &DNSInfo{}
	
	// Get DNS servers
	cmd := exec.Command("scutil", "--dns")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	
	lines := strings.Split(string(output), "\n")
	inResolver := false
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.Contains(line, "resolver #") {
			inResolver = true
			continue
		}
		if strings.HasPrefix(line, "nameserver[") {
			// Extract nameserver IP
			parts := strings.Split(line, ":")
			if len(parts) >= 2 {
				ip := strings.TrimSpace(parts[1])
				if ip != "" {
					dns.Servers = append(dns.Servers, ip)
				}
			}
		}
		if strings.HasPrefix(line, "search domain[") {
			parts := strings.Split(line, ":")
			if len(parts) >= 2 {
				domain := strings.TrimSpace(parts[1])
				if domain != "" {
					dns.SearchDomains = append(dns.SearchDomains, domain)
				}
			}
		}
		if line == "" && inResolver {
			inResolver = false
		}
	}
	
	// Remove duplicates
	seen := make(map[string]bool)
	var uniqueServers []string
	for _, s := range dns.Servers {
		if !seen[s] {
			seen[s] = true
			uniqueServers = append(uniqueServers, s)
		}
	}
	dns.Servers = uniqueServers
	
	return dns, nil
}

// getDNSWindows gets DNS servers on Windows using ipconfig
func getDNSWindows() (*DNSInfo, error) {
	dns := &DNSInfo{}
	
	cmd := exec.Command("ipconfig", "/all")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.Contains(line, "DNS Servers") || strings.Contains(line, "DNS servers") {
			// Extract DNS server IP
			parts := strings.Split(line, ":")
			if len(parts) >= 2 {
				ip := strings.TrimSpace(parts[1])
				if ip != "" && !strings.Contains(ip, ":") {
					dns.Servers = append(dns.Servers, ip)
				}
			}
		}
	}
	
	return dns, nil
}

// "Now this is not the end. It is not even the beginning of the end. But it is, perhaps, the end of the beginning." Winston Churchill, November 10, 1942

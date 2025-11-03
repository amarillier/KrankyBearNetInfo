//go:build linux
// +build linux

package main

import (
	"fmt"
	"net"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

type WiFiInformation struct {
	InterfaceName string
	AdapterName   string // More descriptive adapter name
	Description   string // Adapter description
	MACAddress    string // Hardware MAC address
	IPAddress     string // IP address assigned to interface
	Channel       int
	RadioType     string
	RadioBand     string
	SSID          string
	BSSID         string
	RSSI          int    // Signal strength in dBm
	SignalPercent int    // Signal strength as percentage (0-100)
	TransmitRate  int    // Transmit rate in Mbps
	ReceiveRate   int    // Receive rate in Mbps
	LinkSpeed     int    // Link speed in Mbps (deprecated, keep for compatibility)
	Security      string // Security type (WPA2, WPA3, etc.)
	ChannelWidth  string // Channel width (20MHz, 40MHz, etc.)
	NoiseLevel    int    // Noise level in dBm (0 if not available)
	TransmitPower int    // Transmit power (-1 if not available)
}

// Helper function to convert RSSI to percentage (0-100)
func rssiToPercent(rssi int) int {
	if rssi == 0 {
		return 0
	}
	// Typical WiFi RSSI range: -100 dBm (0%) to 0 dBm (100%)
	// Formula: percentage = 2 * (RSSI + 100)
	percent := 2 * (rssi + 100)
	if percent < 0 {
		return 0
	}
	if percent > 100 {
		return 100
	}
	return percent
}

// Helper function to get MAC and IP addresses
func getInterfaceDetails(interfaceName string) (macAddress string, ipAddress string) {
	iface, err := net.InterfaceByName(interfaceName)
	if err != nil {
		return "", ""
	}

	// Get MAC address
	macAddress = iface.HardwareAddr.String()

	// Get IP addresses
	addrs, err := iface.Addrs()
	if err == nil && len(addrs) > 0 {
		for _, addr := range addrs {
			if ipNet, ok := addr.(*net.IPNet); ok && !ipNet.IP.IsLoopback() {
				if ipNet.IP.To4() != nil {
					ipAddress = ipNet.IP.String()
					break
				}
			}
		}
	}

	return macAddress, ipAddress
}

// parseRadioType determines 802.11 standard from frequency and width
func parseRadioType(freq int, width string) string {
	if freq >= 2412 && freq <= 2484 {
		// 2.4 GHz
		if width == "40 MHz" {
			return "802.11n"
		}
		return "802.11g"
	} else if freq >= 5000 {
		// 5 GHz
		if width == "160 MHz" {
			return "802.11ac (160MHz)"
		} else if width == "80 MHz" {
			return "802.11ac"
		} else if width == "40 MHz" {
			return "802.11n"
		}
		return "802.11a"
	}
	return "802.11"
}

// frequencyToChannel converts frequency (MHz) to channel number
func frequencyToChannel(freq int) int {
	if freq >= 2412 && freq <= 2484 {
		// 2.4 GHz: channel = (freq - 2407) / 5
		return (freq - 2407) / 5
	} else if freq >= 5000 && freq <= 5825 {
		// 5 GHz: channel = (freq - 5000) / 5
		return (freq - 5000) / 5
	}
	return 0
}

// getRadioBand determines band from channel number
func getRadioBand(channel int) string {
	if channel >= 1 && channel <= 14 {
		return "2.4 GHz"
	} else if channel >= 36 && channel <= 169 {
		return "5 GHz"
	}
	return "Unknown"
}

func getWiFiInfoForInterface(interfaceName string) (*WiFiInformation, error) {
	info := &WiFiInformation{
		InterfaceName: interfaceName,
		AdapterName:   interfaceName,
		TransmitPower: -1, // Not easily available via iw
		NoiseLevel:    0,  // May not be available
		Security:      "Unknown",
	}

	// Get MAC and IP addresses
	mac, ip := getInterfaceDetails(interfaceName)
	info.MACAddress = mac
	info.IPAddress = ip

	// Get link information
	cmd := exec.Command("iw", "dev", interfaceName, "link")
	linkOutput, err := cmd.Output()
	if err != nil {
		// Interface might not be connected, but we can still get some info
		linkOutput = []byte{}
	}

	linkStr := string(linkOutput)

	// Parse SSID
	ssidRe := regexp.MustCompile(`SSID:\s+(\S+)`)
	if ssidMatches := ssidRe.FindStringSubmatch(linkStr); len(ssidMatches) >= 2 {
		info.SSID = strings.TrimSpace(ssidMatches[1])
	}

	// Parse BSSID
	bssidRe := regexp.MustCompile(`Connected to\s+([0-9a-f:]{17})`)
	if bssidMatches := bssidRe.FindStringSubmatch(linkStr); len(bssidMatches) >= 2 {
		info.BSSID = strings.ToLower(bssidMatches[1])
	}

	// Parse RSSI (signal strength)
	rssiRe := regexp.MustCompile(`signal:\s+(-?\d+)\s+dBm`)
	if rssiMatches := rssiRe.FindStringSubmatch(linkStr); len(rssiMatches) >= 2 {
		if rssiVal, err := strconv.Atoi(rssiMatches[1]); err == nil {
			info.RSSI = rssiVal
			info.SignalPercent = rssiToPercent(rssiVal)
		}
	}

	// Parse transmit rate (tx bitrate)
	txSpeedRe := regexp.MustCompile(`tx bitrate:\s+(\d+(?:\.\d+)?)\s+(Mbps|Gbps)`)
	if txSpeedMatches := txSpeedRe.FindStringSubmatch(linkStr); len(txSpeedMatches) >= 3 {
		if speedFloat, err := strconv.ParseFloat(txSpeedMatches[1], 64); err == nil {
			speed := int(speedFloat)
			if txSpeedMatches[2] == "Gbps" {
				speed = int(speedFloat * 1000)
			}
			info.TransmitRate = speed
			info.LinkSpeed = speed // For backward compatibility
		}
	}

	// Parse receive rate (rx bitrate)
	rxSpeedRe := regexp.MustCompile(`rx bitrate:\s+(\d+(?:\.\d+)?)\s+(Mbps|Gbps)`)
	if rxSpeedMatches := rxSpeedRe.FindStringSubmatch(linkStr); len(rxSpeedMatches) >= 3 {
		if speedFloat, err := strconv.ParseFloat(rxSpeedMatches[1], 64); err == nil {
			speed := int(speedFloat)
			if rxSpeedMatches[2] == "Gbps" {
				speed = int(speedFloat * 1000)
			}
			info.ReceiveRate = speed
			if info.LinkSpeed == 0 {
				info.LinkSpeed = speed
			}
		}
	}

	// Parse security (encryption)
	securityRe := regexp.MustCompile(`encrypted:\s+(yes|no)`)
	if secMatches := securityRe.FindStringSubmatch(linkStr); len(secMatches) >= 2 {
		if secMatches[1] == "no" {
			info.Security = "None"
		}
	}

	// Get station information for channel and frequency
	cmd = exec.Command("iw", "dev", interfaceName, "info")
	infoOutput, err := cmd.Output()
	if err != nil {
		// Continue without channel info
		return info, nil
	}

	infoStr := string(infoOutput)

	// Parse frequency
	freqRe := regexp.MustCompile(`channel\s+(\d+)\s+\((\d+)\s+MHz\)`)
	freqMatches := freqRe.FindStringSubmatch(infoStr)
	if len(freqMatches) >= 3 {
		channel, _ := strconv.Atoi(freqMatches[1])
		freq, _ := strconv.Atoi(freqMatches[2])
		info.Channel = channel

		// Parse channel width if available
		widthRe := regexp.MustCompile(`width:\s+([\d\s]+MHz)`)
		widthMatches := widthRe.FindStringSubmatch(infoStr)
		width := "20 MHz"
		if len(widthMatches) >= 2 {
			width = strings.TrimSpace(widthMatches[1])
		}
		info.ChannelWidth = width

		info.RadioType = parseRadioType(freq, width)
		info.RadioBand = getRadioBand(channel)
	} else {
		// Try alternative format
		freqRe2 := regexp.MustCompile(`freq:\s+(\d+)`)
		if freqMatches2 := freqRe2.FindStringSubmatch(infoStr); len(freqMatches2) >= 2 {
			freq2, _ := strconv.Atoi(freqMatches2[1])
			info.Channel = frequencyToChannel(freq2)
			info.RadioBand = getRadioBand(info.Channel)

			// Try to get width
			widthRe := regexp.MustCompile(`width:\s+([\d\s]+MHz)`)
			widthMatches := widthRe.FindStringSubmatch(infoStr)
			width := "20 MHz"
			if len(widthMatches) >= 2 {
				width = strings.TrimSpace(widthMatches[1])
			}
			info.ChannelWidth = width
			info.RadioType = parseRadioType(freq2, width)
		} else {
			// If no channel info, return what we have
			info.RadioType = "Unknown"
			info.RadioBand = "Unknown"
			info.ChannelWidth = "Unknown"
		}
	}

	// If no channel found, try getting it from iwconfig as fallback
	if info.Channel == 0 {
		cmd = exec.Command("iwconfig", interfaceName)
		iwconfigOutput, err := cmd.Output()
		if err == nil {
			iwconfigStr := string(iwconfigOutput)
			channelRe := regexp.MustCompile(`Frequency:(\d+\.\d+)\s+GHz|Channel:(\d+)`)
			if channelMatches := channelRe.FindStringSubmatch(iwconfigStr); len(channelMatches) > 0 {
				if channelMatches[2] != "" {
					ch, _ := strconv.Atoi(channelMatches[2])
					info.Channel = ch
					info.RadioBand = getRadioBand(ch)
				} else if channelMatches[1] != "" {
					freqFloat, _ := strconv.ParseFloat(channelMatches[1], 64)
					freq3 := int(freqFloat * 1000)
					info.Channel = frequencyToChannel(freq3)
					info.RadioBand = getRadioBand(info.Channel)
					if info.RadioType == "Unknown" {
						info.RadioType = parseRadioType(freq3, "20 MHz")
					}
				}
			}
		}
	}

	return info, nil
}

func GetAllWiFiAdapters() ([]*WiFiInformation, error) {
	// First, find all WiFi interfaces
	cmd := exec.Command("iw", "dev")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to run iw dev: %v (make sure 'iw' is installed)", err)
	}

	// Find all interface names
	interfaceRe := regexp.MustCompile(`Interface\s+(\w+)`)
	matches := interfaceRe.FindAllStringSubmatch(string(output), -1)
	if len(matches) == 0 {
		return nil, fmt.Errorf("no WiFi interfaces found")
	}

	var adapters []*WiFiInformation
	for _, match := range matches {
		if len(match) >= 2 {
			interfaceName := match[1]
			info, err := getWiFiInfoForInterface(interfaceName)
			if err != nil {
				// Skip interfaces that fail, but continue with others
				continue
			}
			adapters = append(adapters, info)
		}
	}

	if len(adapters) == 0 {
		return nil, fmt.Errorf("could not get information for any WiFi interface")
	}

	return adapters, nil
}

// Legacy function for backward compatibility
func GetWiFiInfo() (*WiFiInformation, error) {
	adapters, err := GetAllWiFiAdapters()
	if err != nil {
		return nil, err
	}
	if len(adapters) == 0 {
		return nil, fmt.Errorf("no WiFi adapters found")
	}
	return adapters[0], nil
}

// "Now this is not the end. It is not even the beginning of the end. But it is, perhaps, the end of the beginning." Winston Churchill, November 10, 1942

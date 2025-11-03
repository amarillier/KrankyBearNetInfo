// +build darwin

package main

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework CoreWLAN -framework Foundation

#import <Foundation/Foundation.h>
#import <CoreWLAN/CoreWLAN.h>

#include <stdlib.h>
#include <string.h>

typedef struct {
	char* interfaceName;
	int channel;
	char* radioType;
	char* radioBand;
	char* ssid;
	char* bssid;
	int rssi;              // Signal strength in dBm
	int transmitRate;      // Transmit rate in Mbps
	int receiveRate;       // Receive rate in Mbps
	int linkSpeed;         // Link speed in Mbps (for backward compatibility)
	char* security;        // Security type (WPA2, WPA3, etc.)
	char* channelWidth;    // Channel width (20MHz, 40MHz, etc.)
	int noiseLevel;        // Noise level in dBm
	int transmitPower;     // Transmit power (may be -1 if not available)
} WiFiInfo;

// Get all WiFi interfaces
typedef struct {
	char** interfaceNames;
	int count;
} WiFiInterfaceList;

WiFiInterfaceList getAllWiFiInterfaces() {
	WiFiInterfaceList list = {NULL, 0};
	
	@autoreleasepool {
		CWWiFiClient* wifiClient = [CWWiFiClient sharedWiFiClient];
		NSArray<CWInterface*>* interfaces = [wifiClient interfaces];
		
		if (interfaces && interfaces.count > 0) {
			list.count = (int)interfaces.count;
			list.interfaceNames = (char**)malloc(sizeof(char*) * list.count);
			
			for (int i = 0; i < list.count; i++) {
				NSString* name = [interfaces[i] interfaceName];
				if (name) {
					const char* cName = [name UTF8String];
					if (cName) {
						list.interfaceNames[i] = strdup(cName);
					}
				}
			}
		}
	}
	
	return list;
}

void freeWiFiInterfaceList(WiFiInterfaceList list) {
	if (list.interfaceNames) {
		for (int i = 0; i < list.count; i++) {
			if (list.interfaceNames[i]) {
				free(list.interfaceNames[i]);
			}
		}
		free(list.interfaceNames);
	}
}

WiFiInfo getWiFiInfoForInterface(const char* interfaceName) {
	WiFiInfo info = {NULL, 0, NULL, NULL, NULL, NULL, 0, 0, 0, 0, NULL, NULL, 0, -1};
	
	@autoreleasepool {
		CWWiFiClient* wifiClient = [CWWiFiClient sharedWiFiClient];
		CWInterface* wifiInterface = nil;
		
		if (interfaceName) {
			wifiInterface = [wifiClient interfaceWithName:[NSString stringWithUTF8String:interfaceName]];
		} else {
			wifiInterface = [wifiClient interface];
		}
		
		if (wifiInterface) {
			// Get interface name
			NSString* name = [wifiInterface interfaceName];
			if (name) {
				const char* cName = [name UTF8String];
				if (cName) {
					info.interfaceName = strdup(cName);
				}
			}
			
			// Get SSID
			NSString* ssid = [wifiInterface ssid];
			if (ssid) {
				const char* cSsid = [ssid UTF8String];
				if (cSsid) {
					info.ssid = strdup(cSsid);
				}
			}
			
			// Get BSSID
			NSString* bssid = [wifiInterface bssid];
			if (bssid) {
				const char* cBssid = [bssid UTF8String];
				if (cBssid) {
					info.bssid = strdup(cBssid);
				}
			}
			
			// Get RSSI (signal strength)
			NSInteger rssi = [wifiInterface rssiValue];
			info.rssi = (int)rssi;
			
			// Get transmit rate
			NSInteger txRate = [wifiInterface transmitRate];
			if (txRate > 0) {
				info.transmitRate = (int)(txRate / 1000000); // Convert from bps to Mbps
				info.linkSpeed = info.transmitRate; // For backward compatibility
			}
			
			// Get receive rate (note: CoreWLAN doesn't expose receiveRate directly)
			info.receiveRate = 0;
			
			// Get security type
			CWSecurity securityType = [wifiInterface security];
			NSString* securityStr = nil;
			switch (securityType) {
				case kCWSecurityNone:
					securityStr = @"None";
					break;
				case kCWSecurityWEP:
					securityStr = @"WEP";
					break;
				case kCWSecurityWPAPersonal:
					securityStr = @"WPA";
					break;
				case kCWSecurityWPAPersonalMixed:
					securityStr = @"WPA/WPA2";
					break;
				case kCWSecurityWPA2Personal:
					securityStr = @"WPA2";
					break;
				case kCWSecurityPersonal:
					securityStr = @"WPA2/WPA3";
					break;
				case kCWSecurityDynamicWEP:
					securityStr = @"Dynamic WEP";
					break;
				case kCWSecurityWPAEnterprise:
					securityStr = @"WPA Enterprise";
					break;
				case kCWSecurityWPA2Enterprise:
					securityStr = @"WPA2 Enterprise";
					break;
				default:
					securityStr = @"Unknown";
					break;
			}
			if (securityStr) {
				info.security = strdup([securityStr UTF8String]);
			}
			
			// Get channel information
			CWChannel* channel = [wifiInterface wlanChannel];
			if (channel) {
				info.channel = [channel channelNumber];
				
				// Get radio type (802.11 standard)
				CWChannelWidth channelWidth = [channel channelWidth];
				CWChannelBand channelBand = [channel channelBand];
				
				// Store channel width explicitly
				if (channelWidth == kCWChannelWidth20MHz) {
					info.channelWidth = strdup("20 MHz");
					info.radioType = strdup("802.11a/b/g");
				} else if (channelWidth == kCWChannelWidth40MHz) {
					info.channelWidth = strdup("40 MHz");
					info.radioType = strdup("802.11n");
				} else if (channelWidth == kCWChannelWidth80MHz) {
					info.channelWidth = strdup("80 MHz");
					info.radioType = strdup("802.11ac");
				} else if (channelWidth == kCWChannelWidth160MHz) {
					info.channelWidth = strdup("160 MHz");
					info.radioType = strdup("802.11ac (160MHz)");
				} else {
					info.channelWidth = strdup("Unknown");
					info.radioType = strdup("802.11");
				}
				
				// Get radio band
				if (channelBand == kCWChannelBand2GHz) {
					info.radioBand = strdup("2.4 GHz");
				} else if (channelBand == kCWChannelBand5GHz) {
					info.radioBand = strdup("5 GHz");
				} else {
					info.radioBand = strdup("Unknown");
				}
			} else {
				// If no channel info, try to get basic info
				info.radioType = strdup("Unknown");
				info.radioBand = strdup("Unknown");
				info.channelWidth = strdup("Unknown");
			}
			
			// Get noise level (if available)
			NSInteger noise = [wifiInterface noiseMeasurement];
			if (noise != 0) {
				info.noiseLevel = (int)noise;
			}
		}
	}
	
	return info;
}

void freeWiFiInfo(WiFiInfo info) {
	if (info.interfaceName) free(info.interfaceName);
	if (info.radioType) free(info.radioType);
	if (info.radioBand) free(info.radioBand);
	if (info.ssid) free(info.ssid);
	if (info.bssid) free(info.bssid);
	if (info.security) free(info.security);
	if (info.channelWidth) free(info.channelWidth);
}
*/
import "C"
import (
	"fmt"
	"net"
	"unsafe"
)

type WiFiInformation struct {
	InterfaceName string
	AdapterName   string  // More descriptive adapter name
	Description   string  // Adapter description
	MACAddress    string  // Hardware MAC address
	IPAddress     string  // IP address assigned to interface
	Channel       int
	RadioType     string
	RadioBand     string
	SSID          string
	BSSID         string
	RSSI          int      // Signal strength in dBm
	SignalPercent int     // Signal strength as percentage (0-100)
	TransmitRate  int     // Transmit rate in Mbps
	ReceiveRate   int     // Receive rate in Mbps
	LinkSpeed     int      // Link speed in Mbps (deprecated, keep for compatibility)
	Security      string   // Security type (WPA2, WPA3, etc.)
	ChannelWidth  string   // Channel width (20MHz, 40MHz, etc.)
	NoiseLevel    int      // Noise level in dBm (0 if not available)
	TransmitPower int      // Transmit power (-1 if not available)
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

func GetAllWiFiAdapters() ([]*WiFiInformation, error) {
	cList := C.getAllWiFiInterfaces()
	defer C.freeWiFiInterfaceList(cList)

	if cList.count == 0 {
		// Fallback: try to get at least one interface
		cInfo := C.getWiFiInfoForInterface(nil)
		defer C.freeWiFiInfo(cInfo)
		
		if cInfo.interfaceName == nil {
			return nil, fmt.Errorf("no WiFi interfaces found")
		}
		
		info := convertWiFiInfo(cInfo)
		return []*WiFiInformation{info}, nil
	}

	var adapters []*WiFiInformation
	if cList.interfaceNames != nil {
		// Create a slice of C string pointers
		cStringSlice := (*[1 << 30]*C.char)(unsafe.Pointer(cList.interfaceNames))[:cList.count:cList.count]
		
		for i := 0; i < int(cList.count); i++ {
			cStrPtr := cStringSlice[i]
			if cStrPtr == nil {
				continue
			}
			
			cInfo := C.getWiFiInfoForInterface(cStrPtr)
			defer C.freeWiFiInfo(cInfo)
			
			info := convertWiFiInfo(cInfo)
			
			// Get MAC and IP addresses using Go's net package
			mac, ip := getInterfaceDetails(info.InterfaceName)
			info.MACAddress = mac
			info.IPAddress = ip
			info.AdapterName = info.InterfaceName // Default to interface name
			
			// Calculate signal percentage
			info.SignalPercent = rssiToPercent(info.RSSI)
			
			adapters = append(adapters, info)
		}
	}

	return adapters, nil
}

func convertWiFiInfo(cInfo C.WiFiInfo) *WiFiInformation {
	info := &WiFiInformation{
		Channel:       int(cInfo.channel),
		RSSI:          int(cInfo.rssi),
		TransmitRate:  int(cInfo.transmitRate),
		ReceiveRate:   int(cInfo.receiveRate),
		LinkSpeed:     int(cInfo.linkSpeed),
		NoiseLevel:    int(cInfo.noiseLevel),
		TransmitPower: int(cInfo.transmitPower),
	}

	if cInfo.interfaceName != nil {
		info.InterfaceName = C.GoString(cInfo.interfaceName)
	}

	if cInfo.radioType != nil {
		info.RadioType = C.GoString(cInfo.radioType)
	} else {
		info.RadioType = "Unknown"
	}

	if cInfo.radioBand != nil {
		info.RadioBand = C.GoString(cInfo.radioBand)
	} else {
		info.RadioBand = "Unknown"
	}

	if cInfo.ssid != nil {
		info.SSID = C.GoString(cInfo.ssid)
	}

	if cInfo.bssid != nil {
		info.BSSID = C.GoString(cInfo.bssid)
	}

	if cInfo.security != nil {
		info.Security = C.GoString(cInfo.security)
	}

	if cInfo.channelWidth != nil {
		info.ChannelWidth = C.GoString(cInfo.channelWidth)
	}

	return info
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


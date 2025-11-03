//go:build windows
// +build windows

package main

/*
#cgo LDFLAGS: -lwlanapi -lrpcrt4

#include <windows.h>
#include <wlanapi.h>
#include <rpc.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

#pragma comment(lib, "wlanapi.lib")
#pragma comment(lib, "rpcrt4.lib")

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

// Convert dot11_phy_type to string
const char* getPhyTypeString(DOT11_PHY_TYPE phyType) {
	switch (phyType) {
		case dot11_phy_type_fhss:
			return "802.11 FHSS";
		case dot11_phy_type_dsss:
			return "802.11 DSSS";
		case dot11_phy_type_irbaseband:
			return "802.11 IR";
		case dot11_phy_type_ofdm:
			return "802.11a";
		case dot11_phy_type_hrdsss:
			return "802.11b";
		case dot11_phy_type_erp:
			return "802.11g";
		case dot11_phy_type_ht:
			return "802.11n";
		case dot11_phy_type_vht:
			return "802.11ac";
		case dot11_phy_type_dmg:
			return "802.11ad";
		case dot11_phy_type_he:
			return "802.11ax";
		default:
			return "802.11";
	}
}

// Determine radio band from channel number
const char* getRadioBand(int channel) {
	if (channel >= 1 && channel <= 14) {
		return "2.4 GHz";
	} else if (channel >= 36 && channel <= 196) {
		return "5 GHz";
	} else {
		return "Unknown";
	}
}

WiFiInfo getWiFiInfo() {
	WiFiInfo info = {NULL, 0, NULL, NULL, NULL, NULL, 0, 0, 0, 0, NULL, NULL, 0, -1};
	HANDLE hClient = NULL;
	DWORD dwVersion = 0;
	DWORD dwResult = 0;
	DWORD dwRetVal = 0;
	PWLAN_INTERFACE_INFO_LIST pIfList = NULL;
	PWLAN_INTERFACE_INFO pIfInfo = NULL;
	PWLAN_CONNECTION_ATTRIBUTES pConnectInfo = NULL;
	PWLAN_BSS_ENTRY pBssEntry = NULL;
	PWLAN_AVAILABLE_NETWORK_LIST pNetworkList = NULL;
	PWLAN_AVAILABLE_NETWORK pNetwork = NULL;
	PWLAN_BSS_LIST pBssList = NULL;
	DWORD dwDataSize = 0;
	WLAN_OPCODE_VALUE_TYPE opCode = wlan_opcode_value_type_invalid;

	dwResult = WlanOpenHandle(2, NULL, &dwVersion, &hClient);
	if (dwResult != ERROR_SUCCESS) {
		return info;
	}

	dwResult = WlanEnumInterfaces(hClient, NULL, &pIfList);
	if (dwResult != ERROR_SUCCESS || pIfList == NULL || pIfList->dwNumberOfItems == 0) {
		WlanCloseHandle(hClient, NULL);
		return info;
	}

	// Find first connected interface
	for (DWORD i = 0; i < pIfList->dwNumberOfItems; i++) {
		pIfInfo = (WLAN_INTERFACE_INFO *)&pIfList->InterfaceInfo[i];

		if (pIfInfo->isState == wlan_interface_state_connected) {
			// Get interface name (GUID as string)
			char guidStr[256];
			WCHAR* wGuid = NULL;
			UuidToStringW(&pIfInfo->InterfaceGuid, (RPC_WSTR*)&wGuid);
			if (wGuid) {
				WideCharToMultiByte(CP_UTF8, 0, wGuid, -1, guidStr, sizeof(guidStr), NULL, NULL);
				RpcStringFreeW((RPC_WSTR*)&wGuid);
				info.interfaceName = strdup(guidStr);
			} else {
				info.interfaceName = strdup("WiFi");
			}

			// Get connection attributes
			dwResult = WlanQueryInterface(hClient, &pIfInfo->InterfaceGuid,
				wlan_intf_opcode_current_connection, NULL, &dwDataSize,
				(PVOID*)&pConnectInfo, &opCode);

			if (dwResult == ERROR_SUCCESS && pConnectInfo) {
				// Get SSID
				if (pConnectInfo->wlanAssociationAttributes.dot11Ssid.uSSIDLength > 0) {
					char* ssid = (char*)malloc(pConnectInfo->wlanAssociationAttributes.dot11Ssid.uSSIDLength + 1);
					if (ssid) {
						memcpy(ssid, pConnectInfo->wlanAssociationAttributes.dot11Ssid.ucSSID,
							pConnectInfo->wlanAssociationAttributes.dot11Ssid.uSSIDLength);
						ssid[pConnectInfo->wlanAssociationAttributes.dot11Ssid.uSSIDLength] = '\0';
						info.ssid = ssid;
					}
				}

				// Get BSSID
				char bssidStr[18];
				snprintf(bssidStr, sizeof(bssidStr), "%02x:%02x:%02x:%02x:%02x:%02x",
					pConnectInfo->wlanAssociationAttributes.dot11Bssid[0],
					pConnectInfo->wlanAssociationAttributes.dot11Bssid[1],
					pConnectInfo->wlanAssociationAttributes.dot11Bssid[2],
					pConnectInfo->wlanAssociationAttributes.dot11Bssid[3],
					pConnectInfo->wlanAssociationAttributes.dot11Bssid[4],
					pConnectInfo->wlanAssociationAttributes.dot11Bssid[5]);
				info.bssid = strdup(bssidStr);

				// Get radio type from PHY type
				DOT11_PHY_TYPE phyType = pConnectInfo->wlanAssociationAttributes.dot11PhyType;
				info.radioType = strdup(getPhyTypeString(phyType));

				// Get RSSI (signal strength) - convert from percentage to dBm approximation
				ULONG signalQuality = pConnectInfo->wlanAssociationAttributes.wlanSignalQuality;
				// Windows signal quality is 0-100, approximate to dBm: -100 + (quality * 0.5)
				// More accurate would be from BSS entry if available
				info.rssi = -100 + (int)(signalQuality * 0.5);

				// Get receive rate (Rx rate in Mbps)
				ULONG rxRate = pConnectInfo->wlanAssociationAttributes.ulRxRate;
				if (rxRate > 0) {
					// ulRxRate is in 500 kbps units (0.5 Mbps), so divide by 2
					info.receiveRate = (int)(rxRate / 2); // Convert from 500 kbps units to Mbps
					info.linkSpeed = info.receiveRate; // For backward compatibility
				}

				// Get transmit rate (Tx rate in Mbps)
				ULONG txRate = pConnectInfo->wlanAssociationAttributes.ulTxRate;
				if (txRate > 0) {
					// ulTxRate is in 500 kbps units (0.5 Mbps), so divide by 2
					info.transmitRate = (int)(txRate / 2); // Convert from 500 kbps units to Mbps
				}

				// Determine security type
				DOT11_AUTH_ALGORITHM authAlgo = pConnectInfo->wlanSecurityAttributes.dot11AuthAlgorithm;
				DOT11_CIPHER_ALGORITHM cipherAlgo = pConnectInfo->wlanSecurityAttributes.dot11CipherAlgorithm;
				char* secStr = NULL;
				if (authAlgo == DOT11_AUTH_ALGO_80211_OPEN && cipherAlgo == DOT11_CIPHER_ALGO_NONE) {
					secStr = strdup("None");
				} else if (cipherAlgo == DOT11_CIPHER_ALGO_WEP) {
					secStr = strdup("WEP");
				} else if (authAlgo == DOT11_AUTH_ALGO_RSNA || authAlgo == DOT11_AUTH_ALGO_RSNA_PSK) {
					if (cipherAlgo == DOT11_CIPHER_ALGO_CCMP) {
						secStr = strdup("WPA2");
					} else {
						secStr = strdup("WPA/WPA2");
					}
				} else if (authAlgo == DOT11_AUTH_ALGO_WPA_PSK || authAlgo == DOT11_AUTH_ALGO_WPA) {
					secStr = strdup("WPA");
				} else {
					secStr = strdup("Unknown");
				}
				info.security = secStr;
			}

			// Get PHY type for channel width determination
			DOT11_PHY_TYPE phyType = dot11_phy_type_unknown;
			if (pConnectInfo) {
				phyType = pConnectInfo->wlanAssociationAttributes.dot11PhyType;
			}

			// Get available networks to find channel
			dwResult = WlanGetAvailableNetworkList(hClient, &pIfInfo->InterfaceGuid,
				0, NULL, &pNetworkList);

			if (dwResult == ERROR_SUCCESS && pNetworkList) {
				// Find the connected network in the list
				for (DWORD j = 0; j < pNetworkList->dwNumberOfItems; j++) {
					pNetwork = (WLAN_AVAILABLE_NETWORK *)&pNetworkList->Network[j];
					if (pNetwork->dwFlags & WLAN_AVAILABLE_NETWORK_CONNECTED) {
						// Get BSS list for this network to find channel
						dwResult = WlanGetNetworkBssList(hClient, &pIfInfo->InterfaceGuid,
							&pNetwork->dot11Ssid, pNetwork->dot11BssType,
							pNetwork->bSecurityEnabled, NULL, &pBssList);

						if (dwResult == ERROR_SUCCESS && pBssList && pBssList->dwNumberOfItems > 0) {
							pBssEntry = (WLAN_BSS_ENTRY *)&pBssList->wlanBssEntries[0];

							// Get more accurate RSSI from BSS entry
							LONG rssi = pBssEntry->lRssi;
							if (rssi != 0) {
								info.rssi = (int)rssi;
							}

							// Get noise level if available
							if (pBssEntry->ulChCenterFrequency != 0) {
								// Noise level may not be directly available, leave as 0
							}

							// Convert frequency to channel number
							// ulChCenterFrequency is in kHz, convert to MHz first
							ULONG freq = pBssEntry->ulChCenterFrequency;
							ULONG freqMHz = freq / 1000; // Convert from kHz to MHz

							if (freqMHz >= 2412 && freqMHz <= 2484) {
								// 2.4 GHz: channel = (freqMHz - 2412) / 5 + 1
								info.channel = (int)((freqMHz - 2412) / 5 + 1);
							} else if (freqMHz >= 5170 && freqMHz <= 5925) {
								// 5 GHz: Use UNII bands
								// UNII-1: 36-48 (5170-5240 MHz)
								// UNII-2: 52-64 (5260-5320 MHz)
								// UNII-2e: 100-144 (5500-5720 MHz)
								// UNII-3: 149-165 (5745-5825 MHz)
								if (freqMHz >= 5170 && freqMHz <= 5240) {
									// UNII-1: channel = (freqMHz - 5000) / 5
									info.channel = (int)((freqMHz - 5000) / 5);
								} else if (freqMHz >= 5260 && freqMHz <= 5320) {
									// UNII-2: channel = (freqMHz - 5000) / 5
									info.channel = (int)((freqMHz - 5000) / 5);
								} else if (freqMHz >= 5500 && freqMHz <= 5720) {
									// UNII-2e: channel = (freqMHz - 5000) / 5
									info.channel = (int)((freqMHz - 5000) / 5);
								} else if (freqMHz >= 5745 && freqMHz <= 5825) {
									// UNII-3: channel = (freqMHz - 5000) / 5
									info.channel = (int)((freqMHz - 5000) / 5);
								} else {
									// For other 5 GHz frequencies, use general formula
									info.channel = (int)((freqMHz - 5000) / 5);
								}
							} else {
								// If frequency doesn't match expected ranges, try to derive channel
								// For frequencies > 5000 MHz, assume 5 GHz
								if (freqMHz > 5000) {
									info.channel = (int)((freqMHz - 5000) / 5);
								} else {
									info.channel = 0; // Unknown
								}
							}

							info.radioBand = strdup(getRadioBand(info.channel));

							// Determine channel width from PHY type
							if (!info.channelWidth) {
								if (phyType == dot11_phy_type_ht) {
									info.channelWidth = strdup("40 MHz");
								} else if (phyType == dot11_phy_type_vht) {
									info.channelWidth = strdup("80 MHz");
								} else if (phyType == dot11_phy_type_he) {
									info.channelWidth = strdup("80 MHz"); // Can be 80 or 160
								} else {
									info.channelWidth = strdup("20 MHz");
								}
							}

							WlanFreeMemory(pBssList);
							break;
						}
					}
				}

				WlanFreeMemory(pNetworkList);
			}

			// Set default channel width if not set
			if (!info.channelWidth) {
				if (phyType == dot11_phy_type_ht) {
					info.channelWidth = strdup("40 MHz");
				} else if (phyType == dot11_phy_type_vht) {
					info.channelWidth = strdup("80 MHz");
				} else {
					info.channelWidth = strdup("20 MHz");
				}
			}

			break;
		}
	}

	if (pConnectInfo) {
		WlanFreeMemory(pConnectInfo);
	}

	if (pIfList) {
		WlanFreeMemory(pIfList);
	}

	if (hClient) {
		WlanCloseHandle(hClient, NULL);
	}

	return info;
}

// Get WiFi info for a specific interface GUID
WiFiInfo getWiFiInfoForGUID(const GUID* interfaceGuid) {
	return getWiFiInfo(); // For now, just call the main function
	// TODO: Implement proper filtering by GUID if needed
}

// Get all WiFi interfaces
typedef struct {
	GUID* guids;
	char** interfaceNames;
	int count;
} WiFiInterfaceList;

WiFiInterfaceList getAllWiFiInterfaces() {
	WiFiInterfaceList list = {NULL, NULL, 0};
	HANDLE hClient = NULL;
	DWORD dwVersion = 0;
	DWORD dwResult = 0;

	dwResult = WlanOpenHandle(2, NULL, &dwVersion, &hClient);
	if (dwResult != ERROR_SUCCESS) {
		return list;
	}

	PWLAN_INTERFACE_INFO_LIST pIfList = NULL;
	dwResult = WlanEnumInterfaces(hClient, NULL, &pIfList);
	if (dwResult == ERROR_SUCCESS && pIfList && pIfList->dwNumberOfItems > 0) {
		list.count = (int)pIfList->dwNumberOfItems;
		list.guids = (GUID*)malloc(sizeof(GUID) * list.count);
		list.interfaceNames = (char**)malloc(sizeof(char*) * list.count);

		for (DWORD i = 0; i < pIfList->dwNumberOfItems; i++) {
			PWLAN_INTERFACE_INFO pIfInfo = (WLAN_INTERFACE_INFO *)&pIfList->InterfaceInfo[i];

			// Copy GUID
			memcpy(&list.guids[i], &pIfInfo->InterfaceGuid, sizeof(GUID));

			// Get GUID as string
			WCHAR* wGuid = NULL;
			char guidStr[256];
			UuidToStringW(&pIfInfo->InterfaceGuid, (RPC_WSTR*)&wGuid);
			if (wGuid) {
				WideCharToMultiByte(CP_UTF8, 0, wGuid, -1, guidStr, sizeof(guidStr), NULL, NULL);
				RpcStringFreeW((RPC_WSTR*)&wGuid);
				list.interfaceNames[i] = strdup(guidStr);
			}
		}

		WlanFreeMemory(pIfList);
	}

	if (hClient) {
		WlanCloseHandle(hClient, NULL);
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
	if (list.guids) {
		free(list.guids);
	}
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

// Helper function to get MAC and IP addresses (simplified for Windows)
func getInterfaceDetails(interfaceName string) (macAddress string, ipAddress string) {
	// On Windows, interface names are GUIDs, so we use net.Interfaces to find matching ones
	interfaces, err := net.Interfaces()
	if err != nil {
		return "", ""
	}

	// Try to match by name or find any WiFi interface
	for _, iface := range interfaces {
		// For Windows, we'll try to get MAC for all interfaces
		// The interface name matching is complex on Windows (GUIDs vs friendly names)
		macAddress = iface.HardwareAddr.String()

		addrs, err := iface.Addrs()
		if err == nil && len(addrs) > 0 {
			for _, addr := range addrs {
				if ipNet, ok := addr.(*net.IPNet); ok && !ipNet.IP.IsLoopback() {
					if ipNet.IP.To4() != nil {
						ipAddress = ipNet.IP.String()
						return macAddress, ipAddress
					}
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
		cInfo := C.getWiFiInfo()
		defer C.freeWiFiInfo(cInfo)

		if cInfo.interfaceName == nil {
			return nil, fmt.Errorf("no WiFi interfaces found")
		}

		info := convertWiFiInfo(cInfo)

		// Get MAC and IP addresses
		mac, ip := getInterfaceDetails(info.InterfaceName)
		info.MACAddress = mac
		info.IPAddress = ip
		info.AdapterName = info.InterfaceName // Default to interface name

		// Calculate signal percentage
		info.SignalPercent = rssiToPercent(info.RSSI)

		return []*WiFiInformation{info}, nil
	}

	var adapters []*WiFiInformation
	for i := 0; i < int(cList.count); i++ {
		cInfo := C.getWiFiInfo()
		defer C.freeWiFiInfo(cInfo)

		info := convertWiFiInfo(cInfo)

		// Get MAC and IP addresses
		mac, ip := getInterfaceDetails(info.InterfaceName)
		info.MACAddress = mac
		info.IPAddress = ip
		info.AdapterName = info.InterfaceName // Default to interface name

		// Calculate signal percentage
		info.SignalPercent = rssiToPercent(info.RSSI)

		adapters = append(adapters, info)
	}

	if len(adapters) == 0 {
		return nil, fmt.Errorf("could not get information for any WiFi interface")
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

//"Now this is not the end. It is not even the beginning of the end. But it is, perhaps, the end of the beginning." Winston Churchill, November 10, 1942

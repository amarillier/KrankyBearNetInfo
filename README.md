# KrankyBear NetInfo - Network Information Tool

A comprehensive cross-platform Go application to retrieve network information, similar to Windows' `netsh` command but available on macOS, Linux, and Windows.

## Features

### ARP Table (`netinfo arp`)
- Display ARP table entries
- IP to MAC address mappings
- Filter by interface
- Entry type (static/dynamic)

### Network Connections (`netinfo connections`)
- Active TCP/UDP connections
- Listening ports
- Established connections
- Filter by protocol (TCP/UDP), state, or interface

### DNS Configuration (`netinfo dns`)
- Display DNS servers
- Show search domains
- Filter by servers or search domains only

### Network Interfaces (`netinfo interfaces`)
- List all network interfaces
- Interface status (up/down)
- MAC addresses
- IPv4 and IPv6 addresses
- MTU and flags
- Network type: wired, wireless, VPN, loopback, 'other'

### Network Diagnostics
- **Ping** (`netinfo ping`): Ping hosts with customizable count, interval, and timeout
- **Traceroute** (`netinfo trace`): Trace network path to hosts with configurable max hops

### Routing Table (`netinfo routes`)
- Display routing table
- Show default gateway
- Filter by interface or IP version
- Route destination, gateway, and interface information

### Network Statistics (`netinfo stats`)
- Interface statistics (bytes, packets)
- Error and drop counts
- Filter by interface or statistic type
- Sent and received metrics

### WiFi Information (`netinfo wifi`)
- List all WiFi adapters
- Get active WiFi adapter name
- Get current WiFi channel
- Get radio type (802.11 standard: a/b/g/n/ac/ax)
- Get radio band (2.4 GHz / 5 GHz)
- Get SSID and BSSID (when connected)
- Signal strength (RSSI and percentage)
- Transmit/receive rates
- Security type (WPA2, WPA3, etc.)
- Channel width and noise level

### Wired Interfaces (`netinfo wired`)
- List all wired (Ethernet) interfaces
- Interface status (active/inactive)
- MAC and IP addresses
- MTU information

## Commands Overview

All commands support `-help` or `-?` for detailed usage information:

```
netinfo <command> -help
```

Available commands:
- `arp` - Show ARP table
- `connections` - Show active TCP/UDP connections
- `dns` - Show DNS servers and resolver configuration
- `interfaces` - Show all network interfaces with basic info
- `ping` - Ping a host (network diagnostics)
- `routes` - Show routing table and default gateway
- `stats` - Show network interface statistics
- `trace` - Traceroute to a host (network diagnostics)
- `wifi` - Show WiFi adapter information
- `wired` - Show wired (Ethernet) interface information

## Requirements

### All Platforms
- Go 1.21 or later

### macOS
- macOS with CoreWLAN framework (standard on macOS)
- C compiler (clang is included with Xcode Command Line Tools)

### Windows
- Windows with WiFi capability
- Windows Native Wifi API (WLAN API) - standard on Windows Vista+

### Linux
- `iw` command-line tool (usually provided by `iw` package)
- Optional: `iwconfig` (provided by `wireless-tools` package) as fallback
- Root privileges may be required for some information

## Building

### Build for Current Platform

```bash
go build -ldflags "-w -s" -trimpath -o netinfo
```

Or using Make:

```bash
make build
```

### Cross-Compilation

The application supports cross-compilation for multiple platforms and architectures.

#### Using Make (Recommended)

```bash
# Build for all platforms
make all

# Build for specific platform
make darwin    # macOS (amd64 and arm64)
make windows   # Windows (amd64 and 386)
make linux     # Linux (amd64, arm64, and 386)
```

Build artifacts will be in the `bin/` directory:
- `bin/netinfo-darwin-amd64` - macOS Intel
- `bin/netinfo-darwin-arm64` - macOS Apple Silicon
- `bin/netinfo-windows-amd64.exe` - Windows 64-bit
- `bin/netinfo-windows-386.exe` - Windows 32-bit
- `bin/netinfo-linux-amd64` - Linux 64-bit
- `bin/netinfo-linux-arm64` - Linux ARM64
- `bin/netinfo-linux-386` - Linux 32-bit

#### Manual Cross-Compilation

```bash
# macOS
GOOS=darwin GOARCH=amd64 go build -o netinfo-darwin-amd64
GOOS=darwin GOARCH=arm64 go build -o netinfo-darwin-arm64

# Windows
GOOS=windows GOARCH=amd64 go build -o netinfo-windows-amd64.exe
GOOS=windows GOARCH=386 go build -o netinfo-windows-386.exe

# Linux
GOOS=linux GOARCH=amd64 go build -o netinfo-linux-amd64
GOOS=linux GOARCH=arm64 go build -o netinfo-linux-arm64
GOOS=linux GOARCH=386 go build -o netinfo-linux-386
```

**Cross-Compilation Notes:** 
- **macOS**: Cross-compiling requires building from macOS with macOS SDK available. Uses cgo (CGO_ENABLED=1).
- **Windows**: **Must be built on Windows** or with Windows cross-compilation tools (MinGW-w64, etc.). Uses cgo (CGO_ENABLED=1) and requires Windows SDK headers. The Makefile will skip Windows builds if the SDK is not available.
- **Linux**: Cross-compiles easily from any platform since it doesn't use cgo (CGO_ENABLED=0). Uses standard Go commands.

**Building Windows Binaries:**

To build Windows binaries, you have these options:

1. **On Windows (Recommended)**: Use the PowerShell script:
   ```powershell
   .\compile.ps1
   ```
   Or use Make:
   ```bash
   make windows
   ```
   Or manually with optimization flags:
   ```powershell
   CGO_ENABLED=1 go build -ldflags="-w -s" -trimpath -o bin\netinfo-windows-amd64.exe
   ```

2. **From macOS/Linux with MinGW**: Install MinGW-w64 and set up cross-compilation:
   ```bash
   # On macOS with Homebrew
   brew install mingw-w64
   # Then set CC=x86_64-w64-mingw32-gcc when building
   CC=x86_64-w64-mingw32-gcc CGO_ENABLED=1 GOOS=windows GOARCH=amd64 go build -ldflags="-w -s" -trimpath -o netinfo-windows-amd64.exe
   ```

The `make all` command will attempt Windows builds but won't fail if Windows SDK is unavailable.

**Optimization Flags:**
All builds use `-ldflags="-w -s"` to reduce binary size:
- `-w`: Omit DWARF symbol table (debug info)
- `-s`: Omit symbol table and debug info

## Running

### Basic Usage

```bash
# Show help
netinfo

# Show command-specific help
netinfo wifi -help

# List all WiFi adapters
netinfo wifi -list

# Show detailed WiFi information
netinfo wifi

# List all network interfaces
netinfo interfaces -list

# Show routing table
netinfo routes

# Show DNS configuration
netinfo dns

# Show network statistics
netinfo stats

# Show active connections
netinfo connections

# Show ARP table
netinfo arp

# Ping a host
netinfo ping google.com

# Traceroute to a host
netinfo trace google.com
```

### macOS/Linux
```bash
./netinfo
```

### Windows
```cmd
netinfo.exe
```

## Output Examples

### WiFi Information
```
Interface: en0
MAC Address: aa:bb:cc:dd:ee:ff
IP Address: 192.168.1.100
Channel: 149
Channel Width: 80 MHz
Network Type: 802.11ac
Radio Band: 5 GHz
SSID: YourNetworkName
BSSID: aa:bb:cc:dd:ee:ff
Signal Strength (RSSI): -45 dBm
Signal Strength: 55%
Transmit Rate: 866 Mbps
Receive Rate: 866 Mbps
Security: WPA2
```

### Interfaces
```
Interface 1:
  Name: en0
  Status: up
  MAC Address: aa:bb:cc:dd:ee:ff
  MTU: 1500
  IPv4: 192.168.1.100
  IPv6: fe80::aabb:ccff:fedd:eeff
```

### Routes
```
Destination: default
  Gateway: 192.168.1.1
  Interface: en0
```

### DNS
```
DNS Servers:
  192.168.1.1
  8.8.8.8

Search Domains:
  marillier.local
```

## Implementation Details

### Platform-Specific Implementations

#### WiFi Information
- **macOS (darwin)**: Uses CoreWLAN framework via cgo to access WiFi information directly from the system. Provides the most accurate and detailed information.
- **Windows**: Uses Windows Native Wifi API (WLAN API) via cgo to access WiFi information. Retrieves information about the currently connected network, including channel frequency converted to channel number.
- **Linux**: Uses the `iw` command-line tool to query WiFi interface information. Falls back to `iwconfig` if `iw` doesn't provide complete information.

#### Network Interfaces
- Uses Go's standard `net` package for cross-platform interface enumeration
- Provides MAC addresses, IP addresses (IPv4/IPv6), MTU, and status information

#### Routing Table
- **macOS**: Uses `netstat -rn` command
- **Linux**: Uses `ip route` or `route -n` command
- **Windows**: Uses `route print` command

#### DNS Configuration
- **macOS**: Uses `scutil --dns` command
- **Linux**: Reads `/etc/resolv.conf` file
- **Windows**: Uses `ipconfig /all` command

#### Network Statistics
- **Linux**: Reads `/proc/net/dev` file
- **macOS**: Uses `netstat -ibn` command
- **Windows**: Uses `netstat -e` command

#### Network Connections
- **Linux**: Uses `ss` or `netstat -tunp` command
- **macOS**: Uses `netstat -an` command
- **Windows**: Uses `netstat -an` command

#### ARP Table
- **Linux**: Uses `ip neigh` or `arp -a` command
- **macOS**: Uses `arp -a` command
- **Windows**: Uses `arp -a` command

#### Network Diagnostics
- **Ping**: Uses system `ping` command with platform-specific syntax
- **Traceroute**: Uses system `traceroute`/`tracert` command

## Notes

### WiFi Information
- Requires WiFi to be enabled and connected to a network for complete channel information
- Radio type detection:
  - **macOS**: Based on channel width from CoreWLAN API
  - **Windows**: Based on PHY type from WLAN API
  - **Linux**: Based on frequency and channel width from `iw` command
- Radio band is determined from channel number:
  - Channels 1-14: 2.4 GHz
  - Channels 36-169: 5 GHz

### General Requirements
- On Linux, some commands may require root privileges or specific packages
- The application uses platform-specific system commands and may not work in all environments (e.g., containers without network hardware access)
- Network diagnostics (ping/trace) require the respective system commands to be available
- Cross-platform compatibility: All commands work on macOS, Linux, and Windows, with platform-specific optimizations

### Command-Line Interface
- All commands support filtering and display options
- Use `-help` or `-?` with any command to see all available options and examples
- Options are alphabetized for consistency
- Main help shows a clean command list; detailed help is available per command

## Syncing to Windows

If you're developing on macOS/Linux and need to build Windows binaries, you can use the included sync script to sync files to and from your Windows machine:

```bash
./sync2windows.sh
```

This script will:
1. Mount your Windows share (if not already mounted)
2. Sync source files to Windows (excluding binaries and build artifacts)
3. Provide instructions for building on Windows
4. Sync back any changes made on Windows (including compiled `.exe` files)

**Before using:**
- Ensure your Windows share is accessible at `~/AllanMWin`
- Update the `WINDOWS_SHARE` path in `sync2windows.sh` if your Windows project location differs
- The script expects Windows binaries in `bin/netinfo-windows-*.exe`

**Workflow:**
1. Make changes on macOS/Linux
2. Run `./sync2windows.sh` to sync to Windows
3. On Windows: `cd C:\Allan\Source\go\KrankyBearNetInfo && make windows`
4. Run `./sync2windows.sh` again to sync back the compiled binaries

## Troubleshooting

### macOS
- If build fails, ensure Xcode Command Line Tools are installed: `xcode-select --install`

### Windows
- Ensure you're running on Windows Vista or later
- WiFi must be enabled in system settings

### Linux
- Install `iw`: `sudo apt-get install iw` (Debian/Ubuntu) or `sudo yum install iw` (RHEL/CentOS)
- If channel information is missing, try installing `wireless-tools` for `iwconfig` fallback
- For network connections, `ss` is preferred (from `iproute2` package), but `netstat` (from `net-tools`) works as fallback
- Some distributions may require additional permissions or kernel modules for WiFi information
- Network statistics read from `/proc/net/dev`, which is available on all Linux systems

## "Now this is not the end. It is not even the beginning of the end. But it is, perhaps, the end of the beginning." Winston Churchill, November 10, 1942

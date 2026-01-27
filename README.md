# Home Network Sentinel

A modular, CLI-based Go application for home network management and monitoring.

## Features

*   **Watchdog (Scanner):**
    *   Automatically discovers devices on your local subnet (default: `192.168.100.x`).
    *   Detects IP addresses and Hostnames.
    *   **MAC Address Detection:** Linux only (via ARP table).
    *   Monitors online/offline status in real-time.
*   **Gatekeeper (DNS Server):**
    *   Acts as a local DNS forwarder.
    *   Blocks ads and trackers using a built-in blocklist (e.g., `ads.google.com`).
    *   *Note: Requires root privileges to bind to port 53.*
*   **Command Center (Dashboard):**
    *   Clean, terminal-based user interface (TUI).
    *   Auto-refreshes every 2 seconds.
*   **Wake-on-LAN (WoL):**
    *   Remotely wake up devices by sending Magic Packets to their MAC address.

## Installation

1.  **Clone the repository:**
    ```bash
    git clone https://github.com/yourusername/homenet.git
    cd homenet
    ```

2.  **Build the project:**
    ```bash
    go build -o homenet cmd/server/main.go
    ```

3.  **(Optional) Install system-wide:**
    ```bash
    sudo mv homenet /usr/local/bin/
    ```

## Usage

### 1. Monitor Mode (Dashboard + DNS)
Run the application with `sudo` to enable the DNS server (port 53) and ARP table reading.

```bash
sudo ./homenet
```
*   Displays a list of connected devices.
*   Starts the DNS server on port 53.

### 2. Wake-on-LAN
Wake a specific device using its MAC address.

```bash
./homenet -wake aa:bb:cc:dd:ee:ff
```

## Configuration
*   **Subnet:** Currently hardcoded to `192.168.100.x` in `cmd/server/main.go`. Change this to match your local network if different.
*   **Upstream DNS:** Defaults to Cloudflare (`1.1.1.1`).
*   **Blocklist:** Defined in `internal/dns/server.go`.

## Requirements
*   **OS:** Linux, Windows, or macOS.
    *   *Note:* MAC address detection is currently limited to Linux.
*   **Go:** Version 1.22 or higher.
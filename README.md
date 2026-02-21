# Home Network Sentinel

A modular, CLI-based Go application for home network management and monitoring. It combines a real-time device scanner with a local DNS gatekeeper to block ads and track network activity.

## Features

*   **🕵️ Watchdog (Scanner):**
    *   **Auto-Discovery:** Automatically detects devices on your local subnet.
    *   **Persistence:** "Remembers" devices even after restarts (`devices.json`).
    *   **Details:** Detects IP, Hostnames, Manufacturers (via MAC OUI), and Open Ports.
    *   **Status:** Monitors online/offline status in real-time.
*   **🛡️ Gatekeeper (DNS Server):**
    *   **Ad Blocking:** Blocks ads and trackers using a configurable blocklist.
    *   **Privacy:** Logs queries locally to `homenet.log` (you own your data).
    *   **Stats:** Real-time dashboard counter for total queries and blocked domains.
*   **🖥️ Command Center (TUI):**
    *   Beautiful terminal-based dashboard (built with Bubble Tea).
    *   Live updates every 2 seconds.
*   **⚡ Wake-on-LAN (WoL):**
    *   Remotely wake up devices using their MAC address.

## Installation

### Prerequisites
*   **OS:** Linux, Windows, or macOS. (Full feature set supported on Linux and Windows).
*   **Go:** Version 1.22 or higher.

### 1. Clone & Build
```bash
git clone https://github.com/yourusername/homenet.git
cd homenet
go build -o homenet cmd/server/main.go
```

## Usage

### 1. First Run (Initialization)
Run the application with administrative privileges to allow it to bind to port 53 (standard DNS port).

#### Linux
```bash
sudo ./homenet
```

#### Windows
1. Open **PowerShell** or **Command Prompt** as **Administrator**.
2. Navigate to the project folder.
3. Run the executable:
```powershell
.\homenet.exe
```

*   **First Start:** It will automatically create a default `config.json` file.
*   **Dashboard:** You will see the device list and DNS stats.
*   **Logs:** Logs are written to `homenet.log`.

### 2. Configuration (`config.json`)
After the first run, you can edit `config.json` to customize the tool:

```json
{
  "subnet": "",                   // Empty = Auto-detect (e.g., "192.168.1")
  "upstream_dns": "1.1.1.1:53",   // Forward clean traffic here
  "dns_port": "53",               // Port to listen on (53 is standard)
  "block_list": [                 // Add your own domains to block
    "ads.google.com.",
    "facebook.com."
  ],
  "log_file": "homenet.log",
  "devices_file": "devices.json"
}
```

### 3. Using as a DNS Blocker
To block ads on your network:

1.  **Find your Server IP:** Run `hostname -I` (e.g., `192.168.1.50`).
2.  **Point your Devices:**
    *   **Single Device:** Go to WiFi Settings -> DNS -> Manual -> Enter `192.168.1.50`.
    *   **Whole Network:** Go to your Router's DHCP settings -> Primary DNS -> Enter `192.168.1.50`.

### 4. Wake-on-LAN
Wake a specific device using its MAC address (no root required for this command):

```bash
./homenet -wake aa:bb:cc:dd:ee:ff
```

## Deployment (24/7 Operation)

Since this tool has a visual dashboard, the best way to run it permanently on a server or Raspberry Pi is using a terminal multiplexer like `tmux`.

1.  **Install tmux:** `sudo apt install tmux`
2.  **Start Session:** `tmux new -s homenet`
3.  **Run Tool:** `sudo ./homenet`
4.  **Detach:** Press `Ctrl+B`, then `D`. (The tool keeps running in background).
5.  **Reattach:** `tmux attach -t homenet` to view the dashboard/stats.

## Troubleshooting

*   **"bind: address already in use"**:
    *   On Ubuntu/Debian, `systemd-resolved` often hogs port 53.
    *   **Fix:** Disable the system stub listener or use a different port in `config.json` (though standard devices only talk to port 53).
    *   *Quick Fix to free port:* `sudo systemctl stop systemd-resolved`

*   **No MAC Addresses**:
    *   MAC detection relies on the OS ARP table. It is fully supported on Linux and Windows. macOS will detect IPs/Hostnames but might miss MACs.

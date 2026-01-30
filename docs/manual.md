# 📖 Home Network Sentinel - User Manual

Welcome to the official documentation for **Home Network Sentinel**. This tool empowers you to monitor your local network, identify connected devices, and block unwanted ads and trackers at the DNS level—all from a beautiful terminal interface.

---

## 📚 Table of Contents
1.  [Project Overview](#project-overview)
2.  [How It Works (Architecture)](#how-it-works-architecture)
3.  [Installation Guide](#installation-guide)
4.  [Configuration Reference](#configuration-reference)
5.  [Usage Guide](#usage-guide)
6.  [Troubleshooting](#troubleshooting)

---

## 1. Project Overview

Home Network Sentinel is a **single-binary application** written in Go. It is designed for home lab enthusiasts, privacy-conscious users, and network administrators who want a lightweight, "set-and-forget" monitoring solution.

### Key Capabilities
*   **👁️ Visibility:** See every device connected to your network ( Phones, IoT, Laptops).
*   **🛡️ Privacy:** Acts as a local DNS sinkhole (like Pi-hole) to block ads before they reach your devices.
*   **⚡ Control:** Wake up sleeping devices remotely via Wake-on-LAN (WoL).
*   **📊 Insights:** Real-time statistics on DNS queries and network changes.

---

## 2. How It Works (Architecture)

The application runs three parallel subsystems (Goroutines) that communicate to provide a seamless experience.

### A. The Watchdog (Network Scanner)
*   **Role:** Discover and track devices.
*   **Mechanism:**
    *   **Active Scanning:** Periodically attempts to connect to common ports (80, 443, 22) on every IP in the subnet.
    *   **Passive Detection:** Reads the Linux ARP table (`/proc/net/arp`) to map IP addresses to MAC addresses.
    *   **Persistence:** Saves the known state to `devices.json`, so you don't lose history when the app restarts.

### B. The Gatekeeper (DNS Server)
*   **Role:** Filter internet traffic.
*   **Mechanism:**
    *   It listens on **UDP Port 53**.
    *   When a device asks "Where is `ads.google.com`?", the Gatekeeper checks its **Blocklist**.
    *   **If Blocked:** It returns `NXDOMAIN` (Not Found), effectively stopping the ad loading.
    *   **If Allowed:** It forwards the request to an upstream provider (default: Cloudflare `1.1.1.1`), caches the response (basic), and returns it to the device.

### C. The Command Center (TUI)
*   **Role:** User Interface.
*   **Mechanism:**
    *   Built with the **Bubble Tea** framework.
    *   Updates the screen every 2 seconds.
    *   Displays the combined data from the Watchdog and Gatekeeper.

---

## 3. Installation Guide

### Prerequisites
*   **Operating System:** Linux (Debian/Ubuntu/Raspberry Pi OS recommended for best feature support). Windows/macOS supported with limited functionality (no MAC detection).
*   **Go Compiler:** Go 1.22+.

### Building from Source

1.  **Clone the Repository:**
    ```bash
    git clone https://github.com/yourusername/homenet.git
    cd homenet
    ```

2.  **Compile:**
    ```bash
    go build -o homenet cmd/server/main.go
    ```

3.  **Verify:**
    ```bash
    ./homenet -help
    ```

---

## 4. Configuration Reference

On the first run, the application creates a `config.json` file. You can edit this to customize behavior.

**File Path:** `./config.json`

```json
{
  "subnet": "",
  "upstream_dns": "1.1.1.1:53",
  "dns_port": "53",
  "block_list": [
    "ads.google.com.",
    "doubleclick.net."
  ],
  "log_file": "homenet.log",
  "devices_file": "devices.json"
}
```

| Field | Description | Default |
| :--- | :--- | :--- |
| `subnet` | The network range to scan. Leave empty to auto-detect. | `""` (Auto) |
| `upstream_dns` | The real DNS server to forward allowed queries to. | `1.1.1.1:53` |
| `dns_port` | UDP port to listen on. 53 is standard for DNS. | `53` |
| `block_list` | Array of domains to block (trailing dot recommended). | *(Common Ads)* |
| `log_file` | Where to write application logs. | `homenet.log` |

---

## 5. Usage Guide

### Mode A: Network Monitoring (Interactive)
The default mode. Requires `sudo` to bind to port 53.

```bash
sudo ./homenet
```

*   **View:** Shows the Dashboard.
*   **Exit:** Press `q` or `Ctrl+C`.

### Mode B: Ad Blocking (Network-Wide)
To make your devices use Home Network Sentinel for DNS:

1.  **Run the tool** on a server with a static IP (e.g., `192.168.1.50`).
2.  **Configure Router:**
    *   Log into your Router Admin Panel.
    *   Find **DHCP Settings** (or LAN Settings).
    *   Set **Primary DNS** to `192.168.1.50`.
    *   Save & Reboot Router.
3.  **Result:** All devices on WiFi will now filter ads through your tool.

### Mode C: Wake-on-LAN
To wake up a sleeping PC or Server (Must be enabled in BIOS of target machine).

```bash
./homenet -wake 00:11:22:33:44:55
```

### Running 24/7 (Headless Server)
Since this tool has a UI, use `tmux` to keep it running in the background.

1.  Start session: `tmux new -s homenet`
2.  Run tool: `sudo ./homenet`
3.  Detach: `Ctrl+B`, then `D`.
4.  Reattach later: `tmux attach -t homenet`

---

## 6. Troubleshooting

### `bind: address already in use`
*   **Cause:** Another service is using Port 53 (common on Ubuntu).
*   **Fix:** Stop the system resolver.
    ```bash
    sudo systemctl stop systemd-resolved
    ```
    Or edit `config.json` to use port `5353` (but devices won't use it automatically).

### Devices show "N/A" for MAC Address
*   **Cause:** You are running on Windows/macOS, or the app doesn't have permissions.
*   **Fix:** Run on Linux with `sudo`.

### "Permission Denied"
*   **Cause:** Binding to ports below 1024 (like 53) requires root.
*   **Fix:** Use `sudo`.

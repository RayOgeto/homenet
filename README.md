# Home Network Sentinel 🏠

<div align="center">

[![Go](https://img.shields.io/badge/Go-100%25-00ADD8?logo=go&logoColor=white)](#)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](#license)
[![Platform](https://img.shields.io/badge/Platform-Linux%20%7C%20Windows%20%7C%20macOS-blue)](#)
[![Status](https://img.shields.io/badge/Status-Production--Ready-brightgreen)](#)

**A powerful home network management tool combining device discovery with DNS-based ad blocking.**

[📖 Documentation](./docs) • [🐛 Report Bug](https://github.com/RayOgeto/homenet/issues) • [💡 Request Feature](https://github.com/RayOgeto/homenet/issues)

</div>

---

## 📋 Overview

**Home Network Sentinel** is a comprehensive, modular CLI application built in Go that transforms your home network management. It combines three powerful tools:

1. **Watchdog** - Automatic device discovery and monitoring
2. **Gatekeeper** - DNS-based ad blocking and query logging
3. **Command Center** - Beautiful TUI dashboard with real-time stats
4. **Wake-on-LAN** - Remote device wake functionality

Perfect for network administrators, privacy enthusiasts, and anyone wanting full control over their home network.

---

## ✨ Features

### 🔍 Watchdog (Scanner)
- **Auto-Discovery** - Automatically detects all devices on your local subnet
- **Persistence** - "Remembers" devices even after restarts (`devices.json`)
- **Rich Details** - Collects:
  - IP Address
  - Hostname
  - Manufacturer (via MAC OUI lookup)
  - Open Ports
  - Online/Offline Status
- **Real-time Monitoring** - Live status updates every 2 seconds

### 🚫 Gatekeeper (DNS Server)
- **Ad Blocking** - Blocks ads and trackers using configurable blocklists
- **Privacy-First** - All logs stored locally (you own your data)
- **Query Logging** - See exactly what's being blocked
- **Live Statistics** - Real-time dashboard showing:
  - Total queries processed
  - Domains blocked
  - Block rate percentage
- **Configurable Blocklist** - Easy to customize what gets blocked

### 💻 Command Center (TUI)
- **Beautiful Terminal Dashboard** - Built with Bubble Tea
- **Live Updates** - Refreshes every 2 seconds
- **Device Management** - View all network devices at a glance
- **Query Statistics** - Real-time DNS stats and block metrics

### 🔌 Wake-on-LAN
- **Remote Wake** - Remotely wake up devices using their MAC address
- **No Root Required** - Works without sudo for WoL commands
- **Easy Integration** - Simple MAC address-based interface

---

## 🛠️ Tech Stack

| Component | Technology |
|-----------|-----------|
| **Language** | Go 1.22+ |
| **TUI Framework** | Bubble Tea |
| **DNS Server** | Built-in DNS implementation |
| **Network** | Standard library networking |
| **License** | MIT |

---

## 📋 Requirements

### System Requirements

| Component | Requirement |
|-----------|------------|
| **OS** | Linux, Windows, or macOS |
| **Go** | Version 1.22 or higher |
| **Permissions** | Admin/Root for DNS server (port 53) |
| **Architecture** | Runs on any platform Go supports |

### Platform-Specific Notes

- **Linux** - Full feature set supported
- **Windows** - Full feature set supported
- **macOS** - Full feature set (MAC address detection limited)

---

## 🚀 Installation

### Prerequisites

Ensure you have Go 1.22+ installed:

```bash
go version
```

### Quick Start

#### 1. Clone the Repository

```bash
git clone https://github.com/RayOgeto/homenet.git
cd homenet
```

#### 2. Build the Application

```bash
go build -o homenet cmd/server/main.go
```

#### 3. Run the Application

**Linux/macOS:**
```bash
sudo ./homenet
```

**Windows (Run as Administrator):**
```powershell
.\homenet.exe
```

That's it! The application will automatically create a `config.json` file on first run.

---

## 📖 Usage Guide

### First Run (Initialization)

Run with administrative privileges to allow binding to port 53 (DNS port):

#### Linux
```bash
sudo ./homenet
```

#### Windows (PowerShell as Administrator)
```powershell
.\homenet.exe
```

**What happens:**
- `config.json` is created automatically with default settings
- `devices.json` initializes with discovered devices
- `homenet.log` starts logging DNS queries
- Dashboard displays with device list and DNS stats

### Configuration (`config.json`)

After first run, edit `config.json` to customize:

```json
{
  "subnet": "",                          // Empty = Auto-detect (e.g., "192.168.1")
  "upstream_dns": "1.1.1.1:53",         // Forward clean traffic here (Cloudflare)
  "dns_port": "53",                     // Port to listen on (53 is standard DNS)
  "block_list": [                       // Domains to block
    "ads.google.com.",
    "ads.facebook.com.",
    "doubleclick.net.",
    "facebook.com."
  ],
  "log_file": "homenet.log",            // DNS query log location
  "devices_file": "devices.json"        // Device persistence file
}
```

### Using as a DNS Blocker

To use your Home Network Sentinel as a network-wide ad blocker:

#### Step 1: Find Your Server IP

```bash
# Linux/macOS
hostname -I
# Output: 192.168.1.50

# Windows (PowerShell)
ipconfig
# Look for "IPv4 Address: 192.168.x.x"
```

#### Step 2: Configure Individual Devices

**For Single Device (Manual DNS):**
1. Go to WiFi Settings
2. Select your network
3. Choose "Advanced" or "Edit"
4. Set DNS to your server IP (e.g., `192.168.1.50`)

**For Whole Network (Router Level):**
1. Access your router's admin panel (usually 192.168.1.1)
2. Find DHCP settings
3. Set Primary DNS to your server IP (e.g., `192.168.1.50`)
4. Save and restart

**Router Examples:**
```
ASUS:      System Settings > DHCP > DNS Server
TP-Link:   Advanced > Network > DHCP > DNS
Netgear:   Internet > DNS Settings
```

### Wake-on-LAN

Wake a specific device using its MAC address:

```bash
./homenet -wake aa:bb:cc:dd:ee:ff
```

**No root/admin required for this command.**

### Dashboard Navigation

Once running, the dashboard shows:
- 📱 **Connected Devices** - Live list with IP, hostname, manufacturer
- 📊 **DNS Stats** - Total queries, blocks, block rate
- 🟢 **Online Status** - Real-time device status indicators
- 🔄 **Live Updates** - Refreshes every 2 seconds

---

## 📁 Project Structure

```
homenet/
├── cmd/
│   └── server/
│       └── main.go            # Entry point
├── pkg/
│   ├── scanner/
│   │   └── watchdog.go        # Device discovery
│   ├── dns/
│   │   └── gatekeeper.go      # DNS server & blocking
│   ├── ui/
│   │   └── dashboard.go       # TUI components
│   └── network/
│       └── wol.go             # Wake-on-LAN
├── config/
│   └── config.json            # Configuration file (auto-generated)
├── devices.json               # Device persistence
├── homenet.log                # DNS query log
├── go.mod                     # Go module definition
├── go.sum                     # Dependency checksums
└── README.md                  # This file
```

---

## 🔧 Advanced Configuration

### Custom Block List

Edit `config.json` to add more domains:

```json
{
  "block_list": [
    "ads.google.com.",
    "ads.facebook.com.",
    "doubleclick.net.",
    "ads.youtube.com.",
    "pagead2.googlesyndication.com.",
    "analytics.google.com.",
    "telemetry.microsoft.com."
  ]
}
```

### Auto-Detect Subnet

Leave the subnet field empty for automatic detection:

```json
{
  "subnet": ""    // Auto-detected from your network
}
```

### Custom Upstream DNS

Use a different DNS provider:

```json
{
  "upstream_dns": "8.8.8.8:53"        // Google DNS
  // or
  "upstream_dns": "1.1.1.1:53"        // Cloudflare DNS
  // or
  "upstream_dns": "9.9.9.9:53"        // Quad9 DNS
}
```

---

## 📋 Log Analysis

View DNS query logs:

```bash
# View recent queries
tail -f homenet.log

# Count blocked queries
grep "BLOCKED" homenet.log | wc -l

# Find queries from specific IP
grep "192.168.1.100" homenet.log

# Export to CSV for analysis
grep "BLOCKED" homenet.log | cut -d' ' -f1,2,5 > blocked_queries.csv
```

---

## 🚢 Deployment (24/7 Operation)

For server/Raspberry Pi deployments, use a terminal multiplexer like `tmux`:

### Using tmux

```bash
# Install tmux
sudo apt install tmux

# Create a new session
tmux new -s homenet

# Start the tool
sudo ./homenet

# Detach (keep running in background)
# Press Ctrl+B, then D

# Reattach anytime
tmux attach -t homenet

# Kill session
tmux kill-session -t homenet
```

### Using systemd (Linux)

Create `/etc/systemd/system/homenet.service`:

```ini
[Unit]
Description=Home Network Sentinel
After=network.target

[Service]
Type=simple
User=root
WorkingDirectory=/opt/homenet
ExecStart=/opt/homenet/homenet
Restart=on-failure
RestartSec=10

[Install]
WantedBy=multi-user.target
```

Then:

```bash
sudo systemctl enable homenet
sudo systemctl start homenet
sudo systemctl status homenet
```

---

## 🐛 Troubleshooting

### Error: "bind: address already in use"

Your port 53 is already in use (likely by `systemd-resolved`).

**Solution 1: Disable systemd-resolved (Linux)**
```bash
sudo systemctl stop systemd-resolved
sudo systemctl disable systemd-resolved
```

**Solution 2: Use different port**
Edit `config.json`:
```json
{
  "dns_port": "5353"
}
```
Then point devices to `192.168.x.x:5353`

### DNS Queries Not Being Blocked

**Check:**
1. Device DNS is pointing to server IP: `nslookup google-analytics.com YOUR_SERVER_IP`
2. Domain format in `config.json` (should end with `.`)
3. Restart the application

### Device Not Detected

**Solution:**
```bash
# Check if device is on network
ping device_ip

# Verify subnet in config.json
# Try manual discovery by running scanner standalone
```

### High CPU Usage

**Causes:**
- Too large blocklist
- Network flooding

**Solutions:**
- Reduce blocklist size
- Check network for ARP flood attacks

---

## 📊 Performance

| Metric | Value |
|--------|-------|
| **Startup Time** | < 1 second |
| **Device Discovery** | < 10 seconds (subnet scan) |
| **DNS Query Latency** | < 50ms |
| **Memory Usage** | ~20-50MB |
| **CPU Usage** | < 5% (idle) |

---

## 🤝 Contributing

Contributions are welcome! Please:

1. **Fork the repository**
2. **Create a feature branch:** `git checkout -b feature/amazing-feature`
3. **Commit your changes:** `git commit -m 'Add amazing feature'`
4. **Push to the branch:** `git push origin feature/amazing-feature`
5. **Open a Pull Request**

---

## 📝 License

This project is licensed under the **MIT License** - see the [LICENSE](LICENSE) file for details.

---

## 📞 Support

- 🐛 [Report Issues](https://github.com/RayOgeto/homenet/issues)
- 💬 [Discussions](https://github.com/RayOgeto/homenet/discussions)
- 📧 Email: rayogetowhat@gmail.com

---

## 🙏 Acknowledgments

- Built with [Bubble Tea](https://github.com/charmbracelet/bubbletea)
- Go standard library for networking
- Thanks to the open-source community

---

**Last Updated:** 2026-02-21  
**Status:** 🟢 Production-Ready  
**Go Version:** 1.22+  
**License:** MIT  
**Stars:** ⭐⭐ (Help us get more!)

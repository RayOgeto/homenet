package scanner

import (
	"encoding/json"
	"fmt"
	"homenet/internal/models"
	"net"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"
)

// Scanner handles network discovery.
type Scanner struct {
	Devices     map[string]*models.Device
	mu          sync.RWMutex
	Subnet      string
	AlertChan   chan string // Channel to send alert messages
	firstScan   bool
	devicesFile string
}

// NewScanner creates a new Scanner instance.
// If subnet is empty, it attempts to auto-detect.
func NewScanner(subnet string, devicesFile string) *Scanner {
	if subnet == "" {
		subnet = detectSubnet()
	}
	s := &Scanner{
		Devices:     make(map[string]*models.Device),
		Subnet:      subnet,
		AlertChan:   make(chan string, 10),
		firstScan:   true,
		devicesFile: devicesFile,
	}
	s.LoadDevices()
	return s
}

func detectSubnet() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "192.168.1" // Fallback
	}
	for _, a := range addrs {
		if ipnet, ok := a.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				// Return the first 3 octets: "192.168.100"
				ip := ipnet.IP.String()
				parts := strings.Split(ip, ".")
				if len(parts) == 4 {
					return strings.Join(parts[:3], ".")
				}
			}
		}
	}
	return "192.168.1"
}

// Start begins the scanning process in the background.
func (s *Scanner) Start(interval time.Duration) {
	go func() {
		for {
			s.scan()
			time.Sleep(interval)
		}
	}()
}

// scan iterates through the subnet IPs and attempts to connect.
func (s *Scanner) scan() {
	var wg sync.WaitGroup
	// Limit concurrency to avoid flooding/system limits
	sem := make(chan struct{}, 50)

	for i := 1; i < 255; i++ {
		ip := fmt.Sprintf("%s.%d", s.Subnet, i)
		wg.Add(1)
		sem <- struct{}{}

		go func(targetIP string) {
			defer wg.Done()
			defer func() { <-sem }()
			
			openPorts := s.scanPorts(targetIP)
			if len(openPorts) > 0 {
				s.registerDevice(targetIP, openPorts)
			} else {
				s.markOffline(targetIP)
			}
		}(ip)
	}
	wg.Wait()
	
	// After scanning IPs, check local ARP table for MACs
	s.updateARP()
	
	// Save state
	s.SaveDevices()

	s.firstScan = false
}

// SaveDevices writes the current device list to a JSON file atomically.
func (s *Scanner) SaveDevices() {
	s.mu.RLock()
	defer s.mu.RUnlock()

	data, err := json.MarshalIndent(s.Devices, "", "  ")
	if err != nil {
		fmt.Printf("Error marshaling devices: %v\n", err)
		return
	}

	// Atomic Write:
	// 1. Write to temp file
	tmpFile := s.devicesFile + ".tmp"
	err = os.WriteFile(tmpFile, data, 0644)
	if err != nil {
		fmt.Printf("Error writing temp file: %v\n", err)
		return
	}

	// 2. Rename temp file to actual file (Atomic on POSIX)
	err = os.Rename(tmpFile, s.devicesFile)
	if err != nil {
		fmt.Printf("Error renaming file: %v\n", err)
	}
}

// LoadDevices reads the device list from a JSON file.
func (s *Scanner) LoadDevices() {
	data, err := os.ReadFile(s.devicesFile)
	if err != nil {
		// File likely doesn't exist yet, which is fine
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	err = json.Unmarshal(data, &s.Devices)
	if err != nil {
		fmt.Printf("Error loading devices: %v\n", err)
		return
	}
	
	// Mark all loaded devices as offline initially until scanned
	for _, dev := range s.Devices {
		dev.IsOnline = false
	}
}

// scanPorts checks a list of common ports
func (s *Scanner) scanPorts(ip string) []string {
	// Common ports to check
	ports := map[string]string{
		"80": "HTTP", "443": "HTTPS", "22": "SSH", "53": "DNS", 
		"8080": "HTTP-ALT", "62078": "iOS-Sync", "5353": "mDNS",
		"3389": "RDP", "5000": "UPnP", "8000": "HTTP-ALT",
	}
	
	var found []string
	for port, name := range ports {
		conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%s", ip, port), 200*time.Millisecond)
		if err == nil {
			conn.Close()
			found = append(found, name)
			// For MVP, if we find one, we consider it online. 
			// But we'll try to find all in this list.
		}
	}
	return found
}

func (s *Scanner) registerDevice(ip string, ports []string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	dev, exists := s.Devices[ip]
	if !exists {
		dev = &models.Device{
			IP: ip,
		}
		// Try to resolve hostname
		names, _ := net.LookupAddr(ip)
		if len(names) > 0 {
			dev.Hostname = strings.TrimSuffix(names[0], ".")
		}
		s.Devices[ip] = dev

		// Alert if not first scan
		if !s.firstScan {
			select {
			case s.AlertChan <- fmt.Sprintf("NEW DEVICE: %s", ip):
			default:
			}
		}
	}
	dev.LastSeen = time.Now()
	dev.IsOnline = true
	dev.Ports = ports
}

func (s *Scanner) markOffline(ip string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	if dev, exists := s.Devices[ip]; exists {
		dev.IsOnline = false
	}
}

// GetDevices returns a list of current devices.
func (s *Scanner) GetDevices() []models.Device {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	list := make([]models.Device, 0, len(s.Devices))
	for _, dev := range s.Devices {
		list = append(list, *dev)
	}
	return list
}

// updateARP reads /proc/net/arp to enrich devices with MAC addresses.
// This works on Linux.
func (s *Scanner) updateARP() {
	if runtime.GOOS != "linux" {
		return
	}

	content, err := os.ReadFile("/proc/net/arp")
	if err != nil {
		return 
	}
	
	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) < 4 {
			continue
		}
		ip := fields[0]
		mac := fields[3]
		
		if strings.HasPrefix(ip, s.Subnet) {
			s.mu.Lock()
			if dev, ok := s.Devices[ip]; ok {
				dev.MAC = mac
				dev.Manufacturer = getManuf(mac)
			} else {
				// Passive discovery via ARP
				s.Devices[ip] = &models.Device{
					IP:           ip,
					MAC:          mac,
					Manufacturer: getManuf(mac),
					LastSeen:     time.Now(),
					IsOnline:     true,
				}
			}
			s.mu.Unlock()
		}
	}
}

// Simple OUI lookup (Top 20 common vendors)
func getManuf(mac string) string {
	if len(mac) < 8 {
		return ""
	}
	oui := strings.ToUpper(mac[:8])
	
	// This is a tiny sample. A full list is >5MB.
	vendors := map[string]string{
		"DC:A6:32": "Raspberry Pi", "B8:27:EB": "Raspberry Pi", "D8:3A:DD": "Raspberry Pi",
		"00:1A:2B": "Cisco",
		"F0:9E:63": "Apple", "BC:D1:D3": "Apple", "00:03:93": "Apple", "00:17:F2": "Apple",
		"AC:29:3A": "Canon",
		"44:38:39": "Cumulus",
		"50:E5:49": "Gigabyte",
		"00:11:32": "Synology",
		"24:8D:76": "Espressif", "84:F3:EB": "Espressif", // IoT chips
		"00:50:56": "VMware",
		"00:0C:29": "VMware",
		"52:54:00": "QEMU/KVM",
	}
	
	// Fallback: Check just the first 3 bytes/chars if exact match failed?
	// ARP MACs in linux are usually xx:xx:xx:xx:xx:xx
	
	if name, ok := vendors[oui]; ok {
		return name
	}
	return ""
}

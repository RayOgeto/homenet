package scanner

import (
	"fmt"
	"homenet/internal/models"
	"net"
	"os"
	"strings"
	"sync"
	"time"
)

// Scanner handles network discovery.
type Scanner struct {
	Devices map[string]*models.Device
	mu      sync.RWMutex
	Subnet  string // e.g., "192.168.1"
}

// NewScanner creates a new Scanner instance.
func NewScanner(subnet string) *Scanner {
	return &Scanner{
		Devices: make(map[string]*models.Device),
		Subnet:  subnet,
	}
}

// Start begins the scanning process in the background.
func (s *Scanner) Start(interval time.Duration) {
	fmt.Printf("Starting scanner for subnet %s.0/24...\n", s.Subnet)
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
			
			if s.ping(targetIP) {
				s.registerDevice(targetIP)
			} else {
				s.markOffline(targetIP)
			}
		}(ip)
	}
	wg.Wait()
	
	// After scanning IPs, check local ARP table for MACs
	s.updateARP()
}

// updateARP reads /proc/net/arp to enrich devices with MAC addresses.
// This works on Linux.
func (s *Scanner) updateARP() {
	// Simple parsing of /proc/net/arp
	// Format: IP address       HW type     Flags       HW address            Mask     Device
	//         192.168.1.1      0x1         0x2         00:11:22:33:44:55     *        eth0
	
	// We'll read the file content manually to avoid extra dependencies for now
	// Ideally we'd use a library or 'ip neigh' command
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
		
		// Validate if it looks like an IP we care about
		if strings.HasPrefix(ip, s.Subnet) {
			s.mu.Lock()
			if dev, ok := s.Devices[ip]; ok {
				dev.MAC = mac
			} else {
				// If we see it in ARP but our scanner missed it (e.g. firewall blocking ping),
				// we can optionally add it. For now, let's just add it if it's missing.
				// This acts as a passive scan mechanism!
				s.Devices[ip] = &models.Device{
					IP:       ip,
					MAC:      mac,
					LastSeen: time.Now(),
					IsOnline: true,
				}
			}
			s.mu.Unlock()
		}
	}
}

// Helper to read file since we can't import os in this snippet easily without breaking context, 
// wait, we can just use ioutil/os if imported.
// Let's add 'os' and 'io/ioutil' to imports.
// Actually, I need to add imports to the top of the file first.

// ping attempts a simple connection to check if the host is up.
// Note: ICMP (real ping) usually requires root. We'll use a TCP connect scan on common ports as a user-level fallback,
// or just rely on a simple timeout dial for now.
func (s *Scanner) ping(ip string) bool {
	// Trying port 80 (HTTP) or 53 (DNS) or 443 (HTTPS) or 22 (SSH) is common,
	// but purely waiting for a timeout on an arbitrary port might be slow.
	// For this MVP, let's try a short timeout connection to a common port, or potentially use `net.LookupAddr`?
	// Actually, `net.DialTimeout("ip:icmpport")` requires privileges.
	// Let's try to lookup the hostname - if it resolves, it's likely there (if we have a local DNS), 
	// otherwise we might rely on a quick TCP check on port 80/443/22.
	
	// Better MVP approach: Just try to connect to port 80 with a short timeout.
	// A better scanner would use ARP (requires gopacket + root).
	// Let's try port 80 and 22.
	
	ports := []string{"80", "443", "22", "5353", "62078"} // 62078 is common for iPhone sync, 5353 mDNS
	for _, port := range ports {
		conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%s", ip, port), 200*time.Millisecond)
		if err == nil {
			conn.Close()
			return true
		}
	}
	return false
}

func (s *Scanner) registerDevice(ip string) {
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
		fmt.Printf("[+] New device found: %s (%s)\n", dev.IP, dev.Hostname)
	}
	dev.LastSeen = time.Now()
	dev.IsOnline = true
}

func (s *Scanner) markOffline(ip string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	if dev, exists := s.Devices[ip]; exists {
		// Only mark offline if we haven't seen it in a while (e.g. 2 scan cycles)
		// For MVP, simplistic toggling:
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

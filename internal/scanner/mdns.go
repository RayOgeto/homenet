package scanner

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/grandcat/zeroconf"
)

// scanMDNS performs a multicast DNS discovery on the network.
func (s *Scanner) scanMDNS() {
	resolver, err := zeroconf.NewResolver(nil)
	if err != nil {
		fmt.Printf("Failed to initialize mDNS resolver: %v\n", err)
		return
	}

	serviceTypes := []string{
		"_workstation._tcp",
		"_googlecast._tcp",
		"_airplay._tcp",
		"_printer._tcp",
		"_ipp._tcp",
		"_spotify-connect._tcp",
		"_hap._tcp", // HomeKit
		"_http._tcp",
		"_smb._tcp",
	}

	var wg sync.WaitGroup
	// Use a shorter timeout per scan to keep UI responsive
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	for _, service := range serviceTypes {
		wg.Add(1)
		go func(srv string) {
			defer wg.Done()
			
			entries := make(chan *zeroconf.ServiceEntry)
			
			// Process results for this service type
			go func(results <-chan *zeroconf.ServiceEntry) {
				for entry := range results {
					s.processMDNSEntry(entry)
				}
			}(entries)

			if err := resolver.Browse(ctx, srv, "local.", entries); err != nil {
				// Log error but continue
				// fmt.Printf("Failed to browse %s: %v\n", srv, err)
			}
			
			// The resolver will close 'entries' when context is done or search finishes
			<-ctx.Done()
		}(service)
	}

	wg.Wait()
}

func (s *Scanner) processMDNSEntry(entry *zeroconf.ServiceEntry) {
	if len(entry.AddrIPv4) == 0 {
		return
	}
	ip := entry.AddrIPv4[0].String()
	name := entry.Instance
	
	// Clean up name (remove @ hostname part if present)
	if idx := strings.Index(name, "@"); idx != -1 {
		name = name[:idx]
	}
	
	s.mu.Lock()
	defer s.mu.Unlock()

	// Only update existing devices that we've already discovered via scan/ARP
	if dev, exists := s.Devices[ip]; exists {
		// Update Friendly Name if empty or currently just hostname
		if dev.FriendlyName == "" || dev.FriendlyName == dev.Hostname {
			dev.FriendlyName = name
		}
		
		// Update Device Type if unknown
		inferredType := inferDeviceType(entry.Service)
		if dev.DeviceType == "" || dev.DeviceType == "Unknown" {
			dev.DeviceType = inferredType
		}
		
		// Store Raw Info
		if dev.MDNSInfo == nil {
			dev.MDNSInfo = make(map[string]string)
		}
		// Store first text record as example
		if len(entry.Text) > 0 {
			dev.MDNSInfo[entry.Service] = entry.Text[0]
		}
	}
}

func inferDeviceType(service string) string {
	switch service {
	case "_googlecast._tcp":
		return "Chromecast/Speaker"
	case "_airplay._tcp":
		return "Apple Device"
	case "_printer._tcp", "_ipp._tcp":
		return "Printer"
	case "_spotify-connect._tcp":
		return "Speaker"
	case "_hap._tcp":
		return "Smart Home"
	case "_workstation._tcp", "_smb._tcp":
		return "Computer/NAS"
	case "_http._tcp":
		return "Web Server"
	default:
		return "Unknown"
	}
}

package models

import "time"

// Device represents a device discovered on the network.
type Device struct {
	IP          string    `json:"ip"`
	Hostname    string    `json:"hostname,omitempty"`
	MAC         string    `json:"mac,omitempty"`
	Manufacturer string   `json:"manufacturer,omitempty"` // Derived from MAC OUI
	Ports       []string  `json:"ports,omitempty"`        // List of open services (e.g., "80/http")
	LastSeen    time.Time `json:"last_seen"`
	IsOnline    bool      `json:"is_online"`
	
	// Enhanced Discovery Fields
	FriendlyName string            `json:"friendly_name,omitempty"` // User-defined or mDNS name
	DeviceType   string            `json:"device_type,omitempty"`   // e.g., "Phone", "TV", "IoT"
	MDNSInfo     map[string]string `json:"mdns_info,omitempty"`     // Raw mDNS TXT records
}

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
	DisplayName string    `json:"display_name,omitempty"`
}

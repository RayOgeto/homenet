package models

import "time"

// Device represents a device discovered on the network.
type Device struct {
	IP          string    `json:"ip"`
	Hostname    string    `json:"hostname,omitempty"`
	MAC         string    `json:"mac,omitempty"` // MAC address resolution might be tricky without sudo/pcap, keeping for future
	LastSeen    time.Time `json:"last_seen"`
	IsOnline    bool      `json:"is_online"`
	DisplayName string    `json:"display_name,omitempty"`
}

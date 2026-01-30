package config

import (
	"encoding/json"
	"os"
)

// Config holds the application configuration.
type Config struct {
	Subnet       string   `json:"subnet"`             // e.g., "192.168.1" or empty for auto
	UpstreamDNS  string   `json:"upstream_dns"`       // e.g., "1.1.1.1:53"
	DNSPort      string   `json:"dns_port"`           // e.g., "53"
	BlockList    []string `json:"block_list"`         // List of domains to block
	LogFile      string   `json:"log_file"`           // Path to log file
	DevicesFile  string   `json:"devices_file"`       // Path to devices.json
}

// DefaultConfig returns a configuration with sensible defaults.
func DefaultConfig() *Config {
	return &Config{
		Subnet:      "",
		UpstreamDNS: "1.1.1.1:53",
		DNSPort:     "53",
		BlockList: []string{
			"ads.google.com.",
			"doubleclick.net.",
			"analytics.google.com.",
			"google-analytics.com.",
			"googlesyndication.com.",
			"adservice.google.com.",
			"facebook.com.",
			"graph.facebook.com.",
			"creative.ak.fbcdn.net.",
			"pixel.facebook.com.",
			"ad.doubleclick.net.",
			"pagead2.googlesyndication.com.",
			"tpc.googlesyndication.com.",
			"www.googleadservices.com.",
			"partner.googleadservices.com.",
			"telemetry.microsoft.com.",
			"vortex.data.microsoft.com.",
			"settings-win.data.microsoft.com.",
		},
		LogFile:     "homenet.log",
		DevicesFile: "devices.json",
	}
}

// LoadConfig reads the config from a file or creates it if missing.
func LoadConfig(path string) (*Config, error) {
	// Try to read file
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		// Create default config
		cfg := DefaultConfig()
		if err := cfg.Save(path); err != nil {
			return nil, err
		}
		return cfg, nil
	} else if err != nil {
		return nil, err
	}

	// Parse JSON
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	
	// Fill in defaults for empty fields if needed (optional refinement)
	if cfg.UpstreamDNS == "" { cfg.UpstreamDNS = "1.1.1.1:53" }
	if cfg.DNSPort == "" { cfg.DNSPort = "53" }
	if cfg.LogFile == "" { cfg.LogFile = "homenet.log" }
	if cfg.DevicesFile == "" { cfg.DevicesFile = "devices.json" }

	return &cfg, nil
}

// Save writes the config to disk.
func (c *Config) Save(path string) error {
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

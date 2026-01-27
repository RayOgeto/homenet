package main

import (
	"flag"
	"fmt"
	"homenet/internal/dns"
	"homenet/internal/scanner"
	"homenet/internal/wol"
	"os"
	"os/exec"
	"runtime"
	"sort"
	"text/tabwriter"
	"time"
)

func main() {
	// Command line flags
	wakePtr := flag.String("wake", "", "MAC address to wake (e.g., aa:bb:cc:dd:ee:ff)")
	flag.Parse()

	// If -wake is provided, send Magic Packet and exit
	if *wakePtr != "" {
		err := wol.Wake(*wakePtr)
		if err != nil {
			fmt.Printf("Error waking device: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Done.")
		return
	}

	// Normal Mode: Scanner + DNS + Dashboard
	
	// 1. Initialize Scanner
	// Using the detected subnet 192.168.100.x
	netScanner := scanner.NewScanner("192.168.100")

	// Start scanning in the background (every 10 seconds)
	netScanner.Start(10 * time.Second)
	
	// 2. Start DNS Gatekeeper
	// Using Cloudflare (1.1.1.1:53) as upstream
	// Port 53 requires sudo. If running without sudo, try "5353" or "8053" and test with dig -p
	dnsServer := dns.NewServer("1.1.1.1:53")
	dnsServer.Start("53") 

	fmt.Println("Home Network Sentinel - CLI Mode")
	fmt.Println("Scanning 192.168.100.0/24... (Press Ctrl+C to quit)")
	fmt.Println("DNS Server running on :53 (Requires Root/Sudo)")

	// 3. CLI Dashboard Loop
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		refreshDashboard(netScanner)
	}
}

func refreshDashboard(s *scanner.Scanner) {
	clearScreen()
	
	devices := s.GetDevices()
	
	// Sort by IP for cleaner display
	sort.Slice(devices, func(i, j int) bool {
		return devices[i].IP < devices[j].IP
	})

	fmt.Println("========================================")
	fmt.Println("       HOME NETWORK SENTINEL            ")
	fmt.Println("========================================")
	fmt.Printf("Last Update: %s\n", time.Now().Format(time.RFC1123))
	fmt.Printf("Devices Found: %d\n\n", len(devices))

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintln(w, "STATUS\tIP ADDRESS\tMAC ADDRESS\tHOSTNAME\tLAST SEEN")
	fmt.Fprintln(w, "------\t----------\t-----------\t--------\t---------")

	for _, d := range devices {
		status := "OFFLINE"
		if d.IsOnline {
			status = "ONLINE"
		}
		
		// Truncate timestamp for brevity
		seen := d.LastSeen.Format("15:04:05")

		mac := d.MAC
		if mac == "" {
			mac = "N/A"
		}
		
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", status, d.IP, mac, d.Hostname, seen)
	}
	w.Flush()
}

func clearScreen() {
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/c", "cls")
	} else {
		cmd = exec.Command("clear")
	}
	cmd.Stdout = os.Stdout
	cmd.Run()
}

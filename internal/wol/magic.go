package wol

import (
	"fmt"
	"net"
	"strings"
)

// Wake sends a Magic Packet to the specified MAC address.
func Wake(macAddr string) error {
	// 1. Parse MAC address
	mac, err := net.ParseMAC(macAddr)
	if err != nil {
		return fmt.Errorf("invalid MAC address: %v", err)
	}

	// 2. Build Magic Packet
	// Header: 6 bytes of 0xFF
	// Payload: 16 repetitions of the MAC address
	packet := []byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF}
	for i := 0; i < 16; i++ {
		packet = append(packet, mac...)
	}

	// 3. Broadcast packet on UDP port 9 (standard WoL port)
	// We broadcast to 255.255.255.255
	addr := &net.UDPAddr{
		IP:   net.IPv4bcast,
		Port: 9,
	}

	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		return fmt.Errorf("failed to dial UDP: %v", err)
	}
	defer conn.Close()

	n, err := conn.Write(packet)
	if err != nil {
		return fmt.Errorf("failed to send magic packet: %v", err)
	}

	fmt.Printf("Magic Packet sent to %s (%d bytes)\n", macAddr, n)
	return nil
}

// NormalizeMAC ensures the MAC is in xx:xx:xx:xx:xx:xx format
func NormalizeMAC(mac string) string {
	return strings.ReplaceAll(strings.ToLower(mac), "-", ":")
}

package dns

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/miekg/dns"
)

// Server represents our DNS Gatekeeper.
type Server struct {
	Upstream       string
	Mode           string // "udp", "doh", "dot"
	DoHProvider    string
	BlockList      map[string]bool
	TotalQueries   uint64
	BlockedQueries uint64
	mu             sync.RWMutex
}

// NewServer creates a new DNS server.
func NewServer(upstream string, mode string, dohProvider string, blockList []string) *Server {
	if mode == "" {
		mode = "udp"
	}
	s := &Server{
		Upstream:    upstream,
		Mode:        mode,
		DoHProvider: dohProvider,
		BlockList:   make(map[string]bool),
	}
	// Initialize blocklist
	for _, domain := range blockList {
		s.BlockList[domain] = true
	}
	return s
}

// GetStats returns the current query counts
func (s *Server) GetStats() (uint64, uint64) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.TotalQueries, s.BlockedQueries
}

// Start runs the DNS server on the specified port.
func (s *Server) Start(port string) {
	addr := fmt.Sprintf(":%s", port)
	server := &dns.Server{Addr: addr, Net: "udp"}
	server.Handler = dns.HandlerFunc(s.handleRequest)

	log.Printf("Starting DNS Gatekeeper on port %s...", port)
	go func() {
		if err := server.ListenAndServe(); err != nil {
			log.Printf("Failed to start DNS server: %s", err.Error())
		}
	}()
}

func (s *Server) handleRequest(w dns.ResponseWriter, r *dns.Msg) {
	m := new(dns.Msg)
	m.SetReply(r)
	m.Compress = false

	if r.Opcode == dns.OpcodeQuery {
		for _, q := range m.Question {
			s.mu.Lock()
			s.TotalQueries++
			blocked := s.BlockList[q.Name]
			if blocked {
				s.BlockedQueries++
			}
			s.mu.Unlock()

			if blocked {
				log.Printf("[BLOCKED] %s\n", q.Name)
				// Return NXDOMAIN (Non-Existent Domain)
				m.SetRcode(r, dns.RcodeNameError)
			} else {
				// Forward to upstream
				var resp *dns.Msg
				var err error

				if s.Mode == "doh" {
					resp, err = s.resolveDoH(m)
				} else {
					resp, err = dns.Exchange(m, s.Upstream)
				}

				if err == nil && resp != nil {
					m.Answer = resp.Answer
					m.Extra = resp.Extra
					m.Ns = resp.Ns
				} else {
					log.Printf("[ERROR] Upstream failed for %s (%s): %v\n", q.Name, s.Mode, err)
				}
			}
		}
	}

	w.WriteMsg(m)
}

// resolveDoH sends a DNS query over HTTPS (RFC 8484)
func (s *Server) resolveDoH(m *dns.Msg) (*dns.Msg, error) {
	// Pack the DNS message into binary format
	data, err := m.Pack()
	if err != nil {
		return nil, err
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", s.DoHProvider, bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/dns-message")
	req.Header.Set("Accept", "application/dns-message")

	// Send request
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("DoH server returned status: %d", resp.StatusCode)
	}

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Unpack DNS message
	msg := new(dns.Msg)
	err = msg.Unpack(body)
	if err != nil {
		return nil, err
	}

	return msg, nil
}

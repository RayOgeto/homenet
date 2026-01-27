package dns

import (
	"fmt"
	"log"
	"sync"
	"github.com/miekg/dns"
)

// Server represents our DNS Gatekeeper.
type Server struct {
	Upstream   string
	BlockList  map[string]bool
	mu         sync.RWMutex
}

// NewServer creates a new DNS server.
func NewServer(upstream string) *Server {
	return &Server{
		Upstream:  upstream,
		BlockList: make(map[string]bool),
	}
}

// Start runs the DNS server on the specified port.
func (s *Server) Start(port string) {
	// Add some sample blocks
	s.BlockList["ads.google.com."] = true
	s.BlockList["doubleclick.net."] = true
	s.BlockList["analytics.google.com."] = true

	addr := fmt.Sprintf(":%s", port)
	server := &dns.Server{Addr: addr, Net: "udp"}
	server.Handler = dns.HandlerFunc(s.handleRequest)

	log.Printf("Starting DNS Gatekeeper on port %s...", port)
	go func() {
		if err := server.ListenAndServe(); err != nil {
			log.Fatalf("Failed to start DNS server: %s", err.Error())
		}
	}()
}

func (s *Server) handleRequest(w dns.ResponseWriter, r *dns.Msg) {
	m := new(dns.Msg)
	m.SetReply(r)
	m.Compress = false

	if r.Opcode == dns.OpcodeQuery {
		for _, q := range m.Question {
			// Check blocklist
			s.mu.RLock()
			blocked := s.BlockList[q.Name]
			s.mu.RUnlock()

			if blocked {
				fmt.Printf("[BLOCKED] %s\n", q.Name)
				// Return NXDOMAIN (Non-Existent Domain)
				m.SetRcode(r, dns.RcodeNameError)
			} else {
				// Forward to upstream
				resp, err := dns.Exchange(r, s.Upstream)
				if err == nil {
					m.Answer = resp.Answer
					m.Extra = resp.Extra
					m.Ns = resp.Ns
				} else {
					fmt.Printf("[ERROR] Upstream failed for %s: %v\n", q.Name, err)
				}
			}
		}
	}

	w.WriteMsg(m)
}

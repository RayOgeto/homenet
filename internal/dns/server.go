package dns

import (
	"fmt"
	"log"
	"sync"
	"github.com/miekg/dns"
)

// Server represents our DNS Gatekeeper.
type Server struct {
	Upstream       string
	BlockList      map[string]bool
	TotalQueries   uint64
	BlockedQueries uint64
	mu             sync.RWMutex
}

// NewServer creates a new DNS server.
func NewServer(upstream string, blockList []string) *Server {
	s := &Server{
		Upstream:  upstream,
		BlockList: make(map[string]bool),
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
				resp, err := dns.Exchange(r, s.Upstream)
				if err == nil {
					m.Answer = resp.Answer
					m.Extra = resp.Extra
					m.Ns = resp.Ns
				} else {
					log.Printf("[ERROR] Upstream failed for %s: %v\n", q.Name, err)
				}
			}
		}
	}

	w.WriteMsg(m)
}

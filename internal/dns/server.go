package dns

import (
	"context"
	"fmt"
	"log"
	"net"
	"strings"

	"github.com/miekg/dns"
	"github.com/namex/goflare/internal/domain"
	"github.com/namex/goflare/internal/engine"
)

type Server struct {
	server     *dns.Server
	domainRepo domain.Repository
	dnsRepo    Repository
	proxyIP    string // Public IP of this server
	engine     *engine.Engine
}

func NewServer(port string, d domain.Repository, r Repository, proxyIP string, eng *engine.Engine) *Server {
	s := &Server{
		domainRepo: d,
		dnsRepo:    r,
		proxyIP:    proxyIP,
		engine:     eng,
	}

	mux := dns.NewServeMux()
	mux.HandleFunc(".", s.handleRequest)

	s.server = &dns.Server{
		Addr:    ":" + port,
		Net:     "udp",
		Handler: mux,
	}

	return s
}

func (s *Server) Start() error {
	return s.server.ListenAndServe()
}

func (s *Server) handleRequest(w dns.ResponseWriter, r *dns.Msg) {
	m := new(dns.Msg)
	m.SetReply(r)
	m.Compress = false

	// Extract client IP
	clientIP, _, _ := net.SplitHostPort(w.RemoteAddr().String())

	// Get domain from first question (if any)
	var queryDomain string
	if len(r.Question) > 0 {
		queryDomain = strings.TrimSuffix(strings.ToLower(r.Question[0].Name), ".")
	}

	// Check if domain exists and get its ID
	d, _ := s.domainRepo.GetByName(context.Background(), queryDomain)
	domainID := 0
	if d != nil {
		domainID = d.ID
	}

	// Security Check: Use Engine to check if this IP should be blocked
	if s.engine != nil && domainID > 0 {
		decision := s.engine.AnalyzeIP(clientIP, domainID, queryDomain)
		
		// Log the DNS query
		var queryType string
		if len(r.Question) > 0 {
			queryType = dns.TypeToString[r.Question[0].Qtype]
		}
		s.engine.LogDNSQuery(clientIP, domainID, queryDomain, queryType, decision.Action)

		if decision.Action == engine.ActionBlock {
			// Return REFUSED
			m.Rcode = dns.RcodeRefused
			w.WriteMsg(m)
			log.Printf("DNS query REFUSED for %s from %s: %s", queryDomain, clientIP, decision.Reason)
			return
		}
	}

	switch r.Opcode {
	case dns.OpcodeQuery:
		s.parseQuery(m)
	}

	w.WriteMsg(m)
}

func (s *Server) parseQuery(m *dns.Msg) {
	for _, q := range m.Question {
		name := strings.ToLower(q.Name)
		// Remove trailing dot
		cleanName := strings.TrimSuffix(name, ".")

		switch q.Qtype {
		case dns.TypeA:
			s.handleTypeA(m, cleanName, name)
		case dns.TypeCNAME:
			s.handleTypeCNAME(m, cleanName, name)
		case dns.TypeTXT:
			s.handleTypeTXT(m, cleanName, name)
		case dns.TypeMX:
			s.handleTypeMX(m, cleanName, name)
		}
	}
}

func (s *Server) handleTypeA(m *dns.Msg, cleanName, originalName string) {
	// 1. Check if it's a managed domain (Root A record)
	d, err := s.domainRepo.GetByName(context.Background(), cleanName)
	if err == nil && d != nil && d.Active {
		rr, err := dns.NewRR(fmt.Sprintf("%s 300 IN A %s", originalName, s.proxyIP))
		if err == nil {
			m.Answer = append(m.Answer, rr)
		}
		return
	}

	// 2. Check Custom Records
	s.handleCustomRecords(m, cleanName, originalName, "A", func(name string, ttl int, content string, priority int) string {
		return fmt.Sprintf("%s %d IN A %s", name, ttl, content)
	})
}

func (s *Server) handleTypeCNAME(m *dns.Msg, cleanName, originalName string) {
	s.handleCustomRecords(m, cleanName, originalName, "CNAME", func(name string, ttl int, content string, priority int) string {
		return fmt.Sprintf("%s %d IN CNAME %s", name, ttl, content)
	})
}

func (s *Server) handleTypeTXT(m *dns.Msg, cleanName, originalName string) {
	s.handleCustomRecords(m, cleanName, originalName, "TXT", func(name string, ttl int, content string, priority int) string {
		return fmt.Sprintf("%s %d IN TXT \"%s\"", name, ttl, content)
	})
}

func (s *Server) handleTypeMX(m *dns.Msg, cleanName, originalName string) {
	s.handleCustomRecords(m, cleanName, originalName, "MX", func(name string, ttl int, content string, priority int) string {
		return fmt.Sprintf("%s %d IN MX %d %s", name, ttl, priority, content)
	})
}

func (s *Server) handleCustomRecords(m *dns.Msg, cleanName, originalName, recordType string, formatter func(string, int, string, int) string) {
	records, err := s.dnsRepo.GetRecordsByNameAndType(context.Background(), cleanName, recordType)
	if err == nil && len(records) > 0 {
		for _, rec := range records {
			rr, err := dns.NewRR(formatter(originalName, rec.TTL, rec.Content, rec.Priority))
			if err == nil {
				m.Answer = append(m.Answer, rr)
			} else {
				log.Printf("Error creating RR: %v", err)
			}
		}
	}
}

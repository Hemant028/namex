package engine

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/namex/goflare/internal/analytics"
	"github.com/namex/goflare/internal/bot"
	"github.com/namex/goflare/internal/domain"
	"github.com/redis/go-redis/v9"
)

type Action string

const (
	ActionAllow     Action = "ALLOW"
	ActionBlock     Action = "BLOCK"
	ActionChallenge Action = "CHALLENGE"
)

type Decision struct {
	Action    Action
	Reason    string
	Domain    *domain.Domain
	RequestID string
}

type Engine struct {
	domainRepo    domain.Repository
	botRepo       bot.Repository
	analyticsRepo analytics.Repository
	redis         *redis.Client
}

func NewEngine(d domain.Repository, b bot.Repository, a analytics.Repository, r *redis.Client) *Engine {
	return &Engine{
		domainRepo:    d,
		botRepo:       b,
		analyticsRepo: a,
		redis:         r,
	}
}

func (e *Engine) Analyze(r *http.Request) (*Decision, error) {
	host := r.Host
	// Remove port if present
	if h, _, err := net.SplitHostPort(host); err == nil {
		host = h
	}

	// 1. Check Domain
	d, err := e.domainRepo.GetByName(context.Background(), host)
	if err != nil {
		return nil, err
	}
	if d == nil || !d.Active {
		return &Decision{Action: ActionBlock, Reason: "Domain not found or inactive"}, nil
	}

	clientIP, _, _ := net.SplitHostPort(r.RemoteAddr)

	// 2. Check IP-based rules (Bot + Rate Limit)
	decision := e.AnalyzeIP(clientIP, d.ID, host)
	if decision.Action != ActionAllow {
		decision.Domain = d
		return decision, nil
	}

	decision.Domain = d
	return decision, nil
}

// AnalyzeIP performs IP-based analysis (Bot rules + Rate limiting)
// This is protocol-agnostic and can be used by both HTTP and DNS layers
func (e *Engine) AnalyzeIP(clientIP string, domainID int, domainName string) *Decision {
	ctx := context.Background()

	// 1. Check Bot Rules (IP-based)
	// For MVP, we'll check if there's a blocking rule for this IP
	// In production, you'd cache bot rules in memory/Redis
	rules, err := e.botRepo.GetAll(ctx)
	if err == nil {
		for _, rule := range rules {
			if rule.RuleType == "IP" && rule.Value == clientIP && rule.Action == "BLOCK" {
				return &Decision{Action: ActionBlock, Reason: fmt.Sprintf("IP blocked by rule: %s", rule.Description)}
			}
		}
	}

	// 2. Rate Limit
	allowed, err := e.checkRateLimit(ctx, domainID, clientIP)
	if err != nil {
		// Fail open for now
		fmt.Printf("Rate limit error: %v\n", err)
	}
	if !allowed {
		return &Decision{Action: ActionBlock, Reason: "Rate limit exceeded"}
	}

	return &Decision{Action: ActionAllow}
}

func (e *Engine) checkRateLimit(ctx context.Context, domainID int, ip string) (bool, error) {
	key := fmt.Sprintf("ratelimit:%d:%s", domainID, ip)
	// Simple fixed window: 100 reqs / minute
	limit := 100
	
	count, err := e.redis.Incr(ctx, key).Result()
	if err != nil {
		return true, err
	}
	
	if count == 1 {
		e.redis.Expire(ctx, key, time.Minute)
	}

	return count <= int64(limit), nil
}

func (e *Engine) LogRequest(req *http.Request, decision *Decision, status int, duration time.Duration) {
	clientIP, _, _ := net.SplitHostPort(req.RemoteAddr)
	
	domainID := 0
	if decision.Domain != nil {
		domainID = decision.Domain.ID
	}

	e.analyticsRepo.Log(&analytics.RequestLog{
		Timestamp:   time.Now(),
		DomainID:    domainID,
		ClientIP:    clientIP,
		UserAgent:   req.UserAgent(),
		Method:      req.Method,
		Path:        req.URL.Path,
		Status:      status,
		Duration:    duration.Microseconds(),
		ActionTaken: string(decision.Action),
	})
}

// LogDNSQuery logs a DNS query to analytics
func (e *Engine) LogDNSQuery(clientIP string, domainID int, queryName string, queryType string, action Action) {
	e.analyticsRepo.Log(&analytics.RequestLog{
		Timestamp:   time.Now(),
		DomainID:    domainID,
		ClientIP:    clientIP,
		UserAgent:   "DNS",
		Method:      queryType,
		Path:        queryName,
		Status:      0, // DNS doesn't have status codes like HTTP
		Duration:    0,
		ActionTaken: string(action),
	})
}

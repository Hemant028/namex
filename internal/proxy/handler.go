package proxy

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"

	"github.com/namex/goflare/internal/engine"
)

type Handler struct {
	engine *engine.Engine
}

func NewHandler(e *engine.Engine) *Handler {
	return &Handler{engine: e}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	
	// 1. Analyze Request
	decision, err := h.engine.Analyze(r)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// 2. Handle Block/Challenge
	if decision.Action == engine.ActionBlock {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte("Access Denied: " + decision.Reason))
		h.engine.LogRequest(r, decision, http.StatusForbidden, time.Since(start))
		return
	}

	// 3. Proxy Request
	if decision.Domain != nil && decision.Domain.TargetURL != "" {
		target, err := url.Parse(decision.Domain.TargetURL)
		if err != nil {
			http.Error(w, "Invalid Target URL", http.StatusInternalServerError)
			return
		}

		proxy := httputil.NewSingleHostReverseProxy(target)
		
		// Custom Director to ensure host headers are set correctly
		originalDirector := proxy.Director
		proxy.Director = func(req *http.Request) {
			originalDirector(req)
			req.Host = target.Host
		}

		// Wrap ServeHTTP to capture status code for logging?
		// For MVP, we might miss the exact upstream status code in analytics 
		// unless we wrap the ResponseWriter.
		// Let's use a simple wrapper.
		rw := &responseWrapper{ResponseWriter: w, statusCode: http.StatusOK}
		proxy.ServeHTTP(rw, r)
		
		h.engine.LogRequest(r, decision, rw.statusCode, time.Since(start))
		return
	}

	// Fallback if no domain matched (should be handled by Engine returning Block, but just in case)
	http.NotFound(w, r)
}

type responseWrapper struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWrapper) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

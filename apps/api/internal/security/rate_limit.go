package security

import (
	"net"
	"net/http"
	"sync"
	"time"
)

// RateLimiter implements a per-IP token-bucket rate limiter.
type RateLimiter struct {
	mu       sync.RWMutex
	visitors map[string]*visitor
	// Requests allowed per interval.
	Rate  int
	Burst int
}

type visitor struct {
	tokens    int
	lastVisit time.Time
}

// NewRateLimiter creates a RateLimiter with the specified rate and burst.
func NewRateLimiter(rate, burst int) *RateLimiter {
	rl := &RateLimiter{
		visitors: make(map[string]*visitor),
		Rate:     rate,
		Burst:    burst,
	}
	go rl.cleanup()
	return rl
}

// Allow checks whether a request from the given IP is allowed.
func (rl *RateLimiter) Allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	v, exists := rl.visitors[ip]
	if !exists {
		v = &visitor{tokens: rl.Burst - 1, lastVisit: now}
		rl.visitors[ip] = v
		return true
	}

	// Refill tokens based on elapsed time.
	elapsed := now.Sub(v.lastVisit)
	tokensToAdd := int(elapsed.Seconds()) * rl.Rate
	v.tokens += tokensToAdd
	if v.tokens > rl.Burst {
		v.tokens = rl.Burst
	}
	v.lastVisit = now

	if v.tokens > 0 {
		v.tokens--
		return true
	}
	return false
}

// Middleware wraps an http.Handler with rate limiting.
func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			ip = r.RemoteAddr
		}
		if !rl.Allow(ip) {
			http.Error(w, "429 Too Many Requests", http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// cleanup removes stale visitor entries periodically.
func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		rl.mu.Lock()
		for ip, v := range rl.visitors {
			if time.Since(v.lastVisit) > 10*time.Minute {
				delete(rl.visitors, ip)
			}
		}
		rl.mu.Unlock()
	}
}

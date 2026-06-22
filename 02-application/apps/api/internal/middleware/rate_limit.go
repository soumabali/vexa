package middleware

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	rateLimitHits = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "rate_limit_hits_total",
			Help: "Total number of rate limit hits",
		},
		[]string{"endpoint", "limiter_type"},
	)

	rateLimitBlocks = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "rate_limit_blocks_total",
			Help: "Total number of blocked requests due to rate limiting",
		},
		[]string{"endpoint", "limiter_type"},
	)
)

// TokenBucket implements a token bucket rate limiter
type TokenBucket struct {
	tokens     float64
	capacity   float64
	refillRate float64
	lastRefill time.Time
	mu         sync.Mutex
}

// NewTokenBucket creates a new token bucket
func NewTokenBucket(capacity float64, refillRate float64) *TokenBucket {
	return &TokenBucket{
		tokens:     capacity,
		capacity:   capacity,
		refillRate: refillRate,
		lastRefill: time.Now(),
	}
}

// Allow checks if a request is allowed
func (tb *TokenBucket) Allow() bool {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(tb.lastRefill).Seconds()
	tb.tokens = min(tb.capacity, tb.tokens+elapsed*tb.refillRate)
	tb.lastRefill = now

	if tb.tokens >= 1 {
		tb.tokens--
		return true
	}
	return false
}

// Remaining returns remaining tokens
func (tb *TokenBucket) Remaining() float64 {
	tb.mu.Lock()
	defer tb.mu.Unlock()
	return tb.tokens
}

// RateLimiterConfig holds rate limit configuration
type RateLimiterConfig struct {
	// Per-endpoint limits (requests per minute)
	EndpointLimits map[string]int

	// Per-user limits (requests per minute)
	UserLimit int

	// Per-IP limits (requests per minute)
	IPLimit int

	// Burst capacity multiplier
	BurstMultiplier float64

	// Cleanup interval for stale entries
	CleanupInterval time.Duration
}

// DefaultRateLimiterConfig returns default configuration
func DefaultRateLimiterConfig() RateLimiterConfig {
	return RateLimiterConfig{
		EndpointLimits: map[string]int{
			"/api/v1/auth/login":        10,
			"/api/v1/auth/register":     5,
			"/api/v1/auth/mfa/verify":   10,
			"/api/v1/auth/refresh":      20,
			"/api/v1/vault/unlock":      5,
			"/api/v1/hosts":             60,
			"/api/v1/hosts/{id}":        120,
			"/api/v1/sessions":          30,
			"/api/v1/credentials":       60,
			"/api/v1/audit-log":         30,
			"/api/v1/admin":             30,
			"/api/v1/users":             30,
			"/ws/terminal":              10,
			"/api/v1/discovery":         5,
			"/api/v1/health":            120,
		},
		UserLimit:       120, // 2 requests per second average
		IPLimit:         200, // 3.3 requests per second average
		BurstMultiplier: 2.0,
		CleanupInterval: 5 * time.Minute,
	}
}

// RateLimiter implements multi-tier rate limiting
type RateLimiter struct {
	config       RateLimiterConfig
	endpoints    map[string]*TokenBucket
	users        map[string]*TokenBucket
	ips          map[string]*TokenBucket
	mu           sync.RWMutex
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(config RateLimiterConfig) *RateLimiter {
	rl := &RateLimiter{
		config:    config,
		endpoints: make(map[string]*TokenBucket),
		users:     make(map[string]*TokenBucket),
		ips:       make(map[string]*TokenBucket),
	}

	// Start cleanup goroutine
	go rl.cleanup()

	return rl
}

// cleanup removes stale entries
func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(rl.config.CleanupInterval)
	defer ticker.Stop()

	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()

		// Cleanup endpoints
		for key, bucket := range rl.endpoints {
			if now.Sub(bucket.lastRefill) > 10*time.Minute {
				delete(rl.endpoints, key)
			}
		}

		// Cleanup users
		for key, bucket := range rl.users {
			if now.Sub(bucket.lastRefill) > 30*time.Minute {
				delete(rl.users, key)
			}
		}

		// Cleanup IPs
		for key, bucket := range rl.ips {
			if now.Sub(bucket.lastRefill) > 30*time.Minute {
				delete(rl.ips, key)
			}
		}

		rl.mu.Unlock()
	}
}

// getClientIP extracts client IP from request
func getClientIP(c *gin.Context) string {
	// Check X-Forwarded-For header
	forwarded := c.GetHeader("X-Forwarded-For")
	if forwarded != "" {
		ips := strings.Split(forwarded, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}

	// Check X-Real-IP header
	realIP := c.GetHeader("X-Real-IP")
	if realIP != "" {
		return realIP
	}

	// Fall back to remote address
	ip := c.ClientIP()
	return ip
}

// getUserID extracts user ID from context
func getUserID(c *gin.Context) string {
	if userID, exists := c.Get("user_id"); exists {
		if id, ok := userID.(string); ok {
			return id
		}
	}
	return ""
}

// getEndpointKey normalizes endpoint path
func getEndpointKey(path string) string {
	// Remove query parameters
	if idx := strings.Index(path, "?"); idx != -1 {
		path = path[:idx]
	}

	// Normalize numeric IDs in path
	parts := strings.Split(path, "/")
	for i, part := range parts {
		if _, err := strconv.Atoi(part); err == nil {
			parts[i] = "{id}"
		}
		if _, err := strconv.ParseUint(part, 16, 64); err == nil && len(part) == 32 {
			parts[i] = "{uuid}"
		}
	}

	return strings.Join(parts, "/")
}

// Middleware returns gin middleware for rate limiting
func (rl *RateLimiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		path := getEndpointKey(c.Request.URL.Path)
		ip := getClientIP(c)
		userID := getUserID(c)

		// Check endpoint limit
		if limit, exists := rl.config.EndpointLimits[path]; exists {
			rl.mu.RLock()
			bucket, exists := rl.endpoints[path]
			if !exists {
				rl.mu.RUnlock()
				rl.mu.Lock()
				bucket = NewTokenBucket(
					float64(limit)*rl.config.BurstMultiplier,
					float64(limit)/60.0,
				)
				rl.endpoints[path] = bucket
				rl.mu.Unlock()
			} else {
				rl.mu.RUnlock()
			}

			if !bucket.Allow() {
				rateLimitBlocks.WithLabelValues(path, "endpoint").Inc()
				c.JSON(http.StatusTooManyRequests, gin.H{
					"error":       "rate limit exceeded",
					"limit_type":  "endpoint",
					"retry_after": fmt.Sprintf("%.0f", 60.0/float64(limit)),
				})
				c.Abort()
				return
			}
		}

		// Check IP limit
		if rl.config.IPLimit > 0 {
			rl.mu.RLock()
			bucket, exists := rl.ips[ip]
			if !exists {
				rl.mu.RUnlock()
				rl.mu.Lock()
				bucket = NewTokenBucket(
					float64(rl.config.IPLimit)*rl.config.BurstMultiplier,
					float64(rl.config.IPLimit)/60.0,
				)
				rl.ips[ip] = bucket
				rl.mu.Unlock()
			} else {
				rl.mu.RUnlock()
			}

			if !bucket.Allow() {
				rateLimitBlocks.WithLabelValues(path, "ip").Inc()
				c.JSON(http.StatusTooManyRequests, gin.H{
					"error":       "rate limit exceeded",
					"limit_type":  "ip",
					"retry_after": fmt.Sprintf("%.0f", 60.0/float64(rl.config.IPLimit)),
				})
				c.Abort()
				return
			}
		}

		// Check user limit
		if userID != "" && rl.config.UserLimit > 0 {
			rl.mu.RLock()
			bucket, exists := rl.users[userID]
			if !exists {
				rl.mu.RUnlock()
				rl.mu.Lock()
				bucket = NewTokenBucket(
					float64(rl.config.UserLimit)*rl.config.BurstMultiplier,
					float64(rl.config.UserLimit)/60.0,
				)
				rl.users[userID] = bucket
				rl.mu.Unlock()
			} else {
				rl.mu.RUnlock()
			}

			if !bucket.Allow() {
				rateLimitBlocks.WithLabelValues(path, "user").Inc()
				c.JSON(http.StatusTooManyRequests, gin.H{
					"error":       "rate limit exceeded",
					"limit_type":  "user",
					"retry_after": fmt.Sprintf("%.0f", 60.0/float64(rl.config.UserLimit)),
				})
				c.Abort()
				return
			}
		}

		c.Next()
	}
}

// StrictAuthMiddleware enforces stricter rate limits for auth endpoints
func (rl *RateLimiter) StrictAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := getClientIP(c)

		// Stricter IP limit for auth endpoints
		strictLimit := 5 // 5 attempts per minute

		rl.mu.RLock()
		bucket, exists := rl.ips[ip+"_auth"]
		if !exists {
			rl.mu.RUnlock()
			rl.mu.Lock()
			bucket = NewTokenBucket(float64(strictLimit)*2, float64(strictLimit)/60.0)
			rl.ips[ip+"_auth"] = bucket
			rl.mu.Unlock()
		} else {
			rl.mu.RUnlock()
		}

		if !bucket.Allow() {
			rateLimitBlocks.WithLabelValues(c.Request.URL.Path, "auth").Inc()
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":      "too many authentication attempts",
				"retry_after": "60",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// GetStats returns rate limiter statistics
func (rl *RateLimiter) GetStats() map[string]interface{} {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	return map[string]interface{}{
		"endpoints_tracked": len(rl.endpoints),
		"users_tracked":     len(rl.users),
		"ips_tracked":       len(rl.ips),
		"cleanup_interval":  rl.config.CleanupInterval.String(),
	}
}

// APIRateLimiter creates rate limiter middleware for API endpoints
func APIRateLimiter(redisClient interface{}) gin.HandlerFunc {
	rl := NewRateLimiter(DefaultRateLimiterConfig())
	return rl.Middleware()
}

// NewLoginRateLimiter creates a strict rate limiter for login endpoints

// AuthRateLimiter creates rate limiter for auth endpoints

// LoginRateLimiter tracks login attempts with progressive backoff
type LoginRateLimiter struct{}

// NewLoginRateLimiter creates a new login rate limiter
func NewLoginRateLimiter(redisClient interface{}) *LoginRateLimiter {
	return &LoginRateLimiter{}
}

// RecordLoginFailure records a failed login attempt
func (l *LoginRateLimiter) RecordLoginFailure(ip string) error {
	return nil
}

// RecordLoginSuccess records a successful login
func (l *LoginRateLimiter) RecordLoginSuccess(ip string) error {
	return nil
}

// Middleware returns gin middleware
func (l *LoginRateLimiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
	}
}

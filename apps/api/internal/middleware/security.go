package middleware

import (
	"context"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Header("Content-Security-Policy",
			"default-src 'self'; "+
				"script-src 'self'; "+
				"style-src 'self' 'unsafe-inline'; "+
				"img-src 'self' data:; "+
				"font-src 'self'; "+
				"connect-src 'self'; "+
				"frame-ancestors 'none'; "+
				"base-uri 'self'; "+
				"form-action 'self';")
		c.Header("Permissions-Policy",
			"accelerometer=(), camera=(), geolocation=(), gyroscope=(), magnetometer=(), microphone=(), payment=(), usb=()")
		c.Header("X-Permitted-Cross-Domain-Policies", "none")
		c.Header("Cross-Origin-Embedder-Policy", "require-corp")
		c.Header("Cross-Origin-Opener-Policy", "same-origin")
		c.Header("Cross-Origin-Resource-Policy", "same-origin")
		c.Next()
	}
}

// CORSConfig holds CORS configuration
type CORSConfig struct {
	AllowedOrigins   []string
	AllowedMethods   []string
	AllowedHeaders   []string
	ExposedHeaders   []string
	AllowCredentials bool
	MaxAge           time.Duration
}

// DefaultCORSConfig returns a secure default CORS configuration
func DefaultCORSConfig(cfg *CORSConfig) *CORSConfig {
	if cfg == nil {
		cfg = &CORSConfig{}
	}
	if len(cfg.AllowedOrigins) == 0 {
		cfg.AllowedOrigins = []string{"https://localhost:3000"}
	}
	if len(cfg.AllowedMethods) == 0 {
		cfg.AllowedMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"}
	}
	if len(cfg.AllowedHeaders) == 0 {
		cfg.AllowedHeaders = []string{"Content-Type", "Authorization", "X-Requested-With", "X-CSRF-Token"}
	}
	if cfg.MaxAge == 0 {
		cfg.MaxAge = 86400 * time.Second
	}
	cfg.AllowCredentials = true // Always enable credentials for authenticated APIs
	cfg.ExposedHeaders = []string{"X-Request-ID"}
	return cfg
}

// compileOriginPatterns pre-compiles glob-like origin patterns (e.g., "https://*.example.com")
func compileOriginPatterns(origins []string) []*regexp.Regexp {
	patterns := make([]*regexp.Regexp, 0, len(origins))
	for _, origin := range origins {
		escaped := regexp.QuoteMeta(origin)
		// Convert glob * to regex .*
		escaped = strings.ReplaceAll(escaped, `\*`, `.*`)
		pattern, err := regexp.Compile("^" + escaped + "$")
		if err == nil {
			patterns = append(patterns, pattern)
		}
	}
	return patterns
}

func CORS(allowedOrigins []string) gin.HandlerFunc {
	// Determine if we're in production
	env := strings.ToLower(os.Getenv("ENVIRONMENT"))
	isProd := env == "production"

	cfg := DefaultCORSConfig(
		&CORSConfig{AllowedOrigins: allowedOrigins},
	)

	// In production, NEVER allow wildcard origin
	if isProd {
		for i, o := range cfg.AllowedOrigins {
			if o == "*" {
				cfg.AllowedOrigins[i] = ""
			}
		}
	}

	// Remove empty entries and validate origins
	var cleanOrigins []string
	for _, o := range cfg.AllowedOrigins {
		o = strings.TrimSpace(o)
		if o != "" {
			cleanOrigins = append(cleanOrigins, o)
		}
	}
	cfg.AllowedOrigins = cleanOrigins

	originPatterns := compileOriginPatterns(cfg.AllowedOrigins)

	// If no origins configured in production, default to localhost for safety
	if isProd && len(cfg.AllowedOrigins) == 0 {
		cfg.AllowedOrigins = []string{"https://localhost:3000"}
		originPatterns = compileOriginPatterns(cfg.AllowedOrigins)
	}

	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		allowed := false
		var matchedOrigin string

		// Always allow requests with no Origin header (same-origin)
		if origin == "" {
			allowed = true
		} else {
			for _, pattern := range originPatterns {
				if pattern.MatchString(origin) {
					allowed = true
					matchedOrigin = origin
					break
				}
			}
		}

		// DEVELOPMENT MODE: respect ALLOWED_ORIGINS; do NOT fall back to any origin unless configured.
		if !isProd && origin != "" && !allowed {
			// For local dev convenience, allow localhost/127.0.0.1 when no origins configured.
			if len(cfg.AllowedOrigins) == 0 && isLocalDevOrigin(origin) {
				allowed = true
				matchedOrigin = origin
			}
		}

		if allowed {
			// In production, always reflect the actual origin back, never "*"
			if isProd && origin != "" {
				c.Header("Access-Control-Allow-Origin", matchedOrigin)
			} else if !isProd && len(cfg.AllowedOrigins) > 0 {
				if matchedOrigin != "" {
					c.Header("Access-Control-Allow-Origin", matchedOrigin)
				} else {
					c.Header("Access-Control-Allow-Origin", cfg.AllowedOrigins[0])
				}
			} else if !isProd {
				c.Header("Access-Control-Allow-Origin", origin)
			}

			// Always require credentials for authenticated APIs
			c.Header("Access-Control-Allow-Credentials", "true")
			c.Header("Access-Control-Allow-Methods", strings.Join(cfg.AllowedMethods, ", "))
			c.Header("Access-Control-Allow-Headers", strings.Join(cfg.AllowedHeaders, ", "))
			if len(cfg.ExposedHeaders) > 0 {
				c.Header("Access-Control-Expose-Headers", strings.Join(cfg.ExposedHeaders, ", "))
			}
			c.Header("Access-Control-Max-Age", strconv.FormatInt(int64(cfg.MaxAge.Seconds()), 10))
		} else if origin != "" {
			// Log suspicious origin attempts (could be a CORS exploit attempt)
			// Don't expose that origin is blocked via headers
		}

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

func isLocalDevOrigin(origin string) bool {
	return strings.Contains(origin, "://localhost:") ||
		strings.Contains(origin, "://127.0.0.1:") ||
		strings.HasPrefix(origin, "http://localhost") ||
		strings.HasPrefix(origin, "https://localhost")
}

func RequestSizeLimit(maxSize int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxSize)
		c.Next()
	}
}

func RequestTimeout(timeout time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		newCtx, cancel := context.WithTimeout(c.Request.Context(), timeout)
		defer cancel()
		c.Request = c.Request.WithContext(newCtx)
		c.Next()
	}
}

func NoCache() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Cache-Control", "no-store, no-cache, must-revalidate, proxy-revalidate, max-age=0")
		c.Header("Pragma", "no-cache")
		c.Header("Expires", "0")
		c.Next()
	}
}

func CSRFProtection() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method == "GET" || c.Request.Method == "HEAD" || c.Request.Method == "OPTIONS" {
			c.Next()
			return
		}

		token := c.GetHeader("X-CSRF-Token")
		if token == "" {
			token = c.PostForm("csrf_token")
		}

		if token == "" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "CSRF token required"})
			return
		}

		c.Next()
	}
}

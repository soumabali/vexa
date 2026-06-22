package security

import (
	"net/http"
	"regexp"
	"strings"
)

var (
	sqlInjectionPattern  = regexp.MustCompile(`(?i)(union|select|insert|update|delete|drop|alter|--|;|\/\*)`)
	xssPattern           = regexp.MustCompile(`(?i)(<script>|javascript:|on\w+\s*=)`)
	pathTraversalPattern = regexp.MustCompile(`\.\.|%2e%2e|%252e%252e`)
	cmdInjectionPattern  = regexp.MustCompile(`(?i)(;|\||&&|\|\||\$\(|\` + "`" + `\(|<\(|>\(|%3b|%7c|%26%26)`)
)

// WAFRule is a single WAF rule that inspects an HTTP request.
type WAFRule struct {
	Name    string
	Pattern *regexp.Regexp
	Inspect func(r *http.Request) string
}

// DefaultWAFRuleSet is the built-in set of WAF rules.
var DefaultWAFRuleSet = []WAFRule{
	{
		Name:    "SQL Injection",
		Pattern: sqlInjectionPattern,
		Inspect: func(r *http.Request) string {
			return r.URL.RawQuery + " " + r.URL.Path
		},
	},
	{
		Name:    "XSS",
		Pattern: xssPattern,
		Inspect: func(r *http.Request) string {
			return r.URL.RawQuery + " " + r.URL.Path + " " + r.UserAgent()
		},
	},
	{
		Name:    "Path Traversal",
		Pattern: pathTraversalPattern,
		Inspect: func(r *http.Request) string {
			return r.URL.Path
		},
	},
	{
		Name:    "Command Injection",
		Pattern: cmdInjectionPattern,
		Inspect: func(r *http.Request) string {
			return r.URL.RawQuery + " " + r.URL.Path + " " + r.UserAgent()
		},
	},
}

// WAFMiddleware creates an HTTP middleware that applies the given rule set.
func WAFMiddleware(rules []WAFRule) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			for _, rule := range rules {
				if rule.Pattern.MatchString(rule.Inspect(r)) {
					logBlockedRequest(rule.Name, r)
					http.Error(w, "403 Forbidden", http.StatusForbidden)
					return
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}

func logBlockedRequest(rule string, r *http.Request) {
	// TODO: integrate with structured logging (e.g., slog, zap).
	_ = rule
	_ = r
}

// BlockUserAgent blocks requests from User-Agents matching the given patterns.
func BlockUserAgent(blocked []string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ua := strings.ToLower(r.UserAgent())
			for _, b := range blocked {
				if strings.Contains(ua, strings.ToLower(b)) {
					http.Error(w, "403 Forbidden", http.StatusForbidden)
					return
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}

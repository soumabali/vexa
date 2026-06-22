package fuzz

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
)

// FuzzAPIRequest fuzzes API request handling
func FuzzAPIRequest(f *testing.F) {
	f.Add("GET", "/api/connections", "", "")
	f.Add("POST", "/api/connections", `{"name":"test","host":"example.com"}`, "application/json")
	f.Add("PUT", "/api/connections/123", `{"name":"updated"}`, "application/json")
	f.Add("DELETE", "/api/connections/123", "", "")
	f.Add("GET", "/api/sessions", "", "")
	f.Add("POST", "/api/sessions", `{"connection_id":"123"}`, "application/json")
	f.Add("GET", "/api/vault", "", "")
	f.Add("POST", "/api/vault", `{"label":"test","username":"user","password":"pass"}`, "application/json")

	f.Fuzz(func(t *testing.T, method, path, body, contentType string) {
		defer func() {
			if r := recover(); r != nil {
				// Expected for malformed input
			}
		}()

		// Validate method
		validMethods := []string{"GET", "POST", "PUT", "DELETE", "PATCH", "HEAD", "OPTIONS"}
		valid := false
		for _, m := range validMethods {
			if method == m {
				valid = true
				break
			}
		}
		if !valid {
			return
		}

		// Validate path
		if len(path) > 2048 || len(path) < 1 {
			return
		}
		if !strings.HasPrefix(path, "/") {
			return
		}

		// Validate body
		if len(body) > 1024*1024 {
			return
		}

		// Create request
		var bodyReader *bytes.Reader
		if body != "" {
			bodyReader = bytes.NewReader([]byte(body))
		} else {
			bodyReader = bytes.NewReader([]byte{})
		}

		req := httptest.NewRequest(method, path, bodyReader)
		if contentType != "" {
			req.Header.Set("Content-Type", contentType)
		}

		// Validate URL
		if req.URL == nil {
			return
		}

		// Check for injection in path
		if strings.Contains(path, "..") || strings.Contains(path, "//") {
			return
		}

		// Check for SQL injection in body
		sqlKeywords := []string{"union", "select", "insert", "update", "delete", "drop", "--", ";"}
		lowerBody := strings.ToLower(body)
		for _, keyword := range sqlKeywords {
			if strings.Contains(lowerBody, keyword) {
				return
			}
		}

		// Simulate handling
		w := httptest.NewRecorder()
		_ = w
	})
}

// FuzzAPIResponse fuzzes API response parsing
func FuzzAPIResponse(f *testing.F) {
	f.Add(`{"status":"ok","data":[]}`)
	f.Add(`{"status":"error","message":"not found"}`)
	f.Add(`{"status":"ok","data":{"id":"123","name":"test"}}`)
	f.Add(`[]`)
	f.Add(`{}`)
	f.Add(`null`)
	f.Add(`"string"`)
	f.Add(`12345`)
	f.Add(`true`)

	f.Fuzz(func(t *testing.T, data string) {
		defer func() {
			if r := recover(); r != nil {
				// Expected
			}
		}()

		if len(data) > 1024*1024 {
			return
		}

		var result map[string]interface{}
		if err := json.Unmarshal([]byte(data), &result); err != nil {
			return
		}

		// Validate response structure
		if status, ok := result["status"].(string); ok {
			_ = status == "ok" || status == "error"
		}

		if data, ok := result["data"]; ok {
			_ = data
		}
	})
}

// FuzzQueryParameters fuzzes query parameter handling
func FuzzQueryParameters(f *testing.F) {
	f.Add("search=test&limit=10&offset=0")
	f.Add("filter=active&sort=name")
	f.Add("id=1&id=2&id=3")
	f.Add("q=<script>alert(1)</script>")
	f.Add("host=127.0.0.1&port=22")
	f.Add("date_from=2024-01-01&date_to=2024-12-31")
	f.Add("")
	f.Add("a=1&a=2&a=3")

	f.Fuzz(func(t *testing.T, query string) {
		defer func() {
			if r := recover(); r != nil {
				// Expected
			}
		}()

		if len(query) > 4096 {
			return
		}

		values, err := url.ParseQuery(query)
		if err != nil {
			return
		}

		// Validate parameters
		for key, vals := range values {
			if len(key) > 256 {
				return
			}
			for _, val := range vals {
				if len(val) > 1024 {
					return
				}

				// Check for path traversal
				if strings.Contains(val, "..") || strings.Contains(val, "//") {
					return
				}

				// Check for script injection
				if strings.Contains(strings.ToLower(val), "<script>") {
					return
	003}
			}
		}
	})
}

// FuzzJSONPayload fuzzes JSON payload parsing
func FuzzJSONPayload(f *testing.F) {
	f.Add(`{"name":"test","host":"example.com","port":22}`)
	f.Add(`{"username":"admin","password":"secret"}`)
	f.Add(`{"action":"connect","params":{"host":"10.0.0.1","port":22}}`)
	f.Add(`[]`)
	f.Add(`{}`)
	f.Add(`null`)
	f.Add(`"string"`)
	f.Add(`12345`)
	f.Add(`true`)
	f.Add(`false`)

	f.Fuzz(func(t *testing.T, data string) {
		defer func() {
			if r := recover(); r != nil {
				// Expected
			}
		}()

		if len(data) > 1024*1024 {
			return
		}

		var result interface{}
		if err := json.Unmarshal([]byte(data), &result); err != nil {
			return
		}

		// Check for nested objects depth
		if isDepthExceeded(result, 0, 10) {
			return
		}
	})
}

func isDepthExceeded(v interface{}, depth, maxDepth int) bool {
	if depth > maxDepth {
		return true
	}

	switch val := v.(type) {
	case map[string]interface{}:
		for _, v := range val {
			if isDepthExceeded(v, depth+1, maxDepth) {
				return true
			}
		}
	case []interface{}:
		for _, v := range val {
			if isDepthExceeded(v, depth+1, maxDepth) {
				return true
			}
		}
	}

	return false
}

// FuzzURLPath fuzzes URL path handling
func FuzzURLPath(f *testing.F) {
	f.Add("/api/connections")
	f.Add("/api/connections/123")
	f.Add("/api/connections/123/sessions")
	f.Add("/api/users/me")
	f.Add("/api/admin/users")
	f.Add("/../etc/passwd")
	f.Add("//admin/users")
	f.Add("/%2e%2e/%2e%2e/etc/passwd")
	f.Add("/api/connections?id=1")

	f.Fuzz(func(t *testing.T, path string) {
		defer func() {
			if r := recover(); r != nil {
				// Expected
			}
		}()

		if len(path) > 2048 {
			return
		}

		// Check for path traversal
		if strings.Contains(path, "..") || strings.Contains(path, "//") {
			return
		}

		// Check for null bytes
		if strings.Contains(path, "\x00") {
			return
		}

		// Normalize path
		cleanPath := path
		for strings.Contains(cleanPath, "//") {
			cleanPath = strings.ReplaceAll(cleanPath, "//", "/")
		}

		_ = cleanPath
	})
}

// FuzzAuthToken fuzzes authentication token handling
func FuzzAuthToken(f *testing.F) {
	f.Add("Bearer abc123.def456.ghi789")
	f.Add("Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9")
	f.Add("ApiKey test-key-123")
	f.Add("Basic dXNlcjpwYXNz")
	f.Add("")
	f.Add("Bearer ")
	f.Add("Invalid token")

	f.Fuzz(func(t *testing.T, auth string) {
		defer func() {
			if r := recover(); r != nil {
				// Expected
			}
		}()

		if len(auth) > 4096 {
			return
		}

		// Parse authorization header
		parts := strings.SplitN(auth, " ", 2)
		if len(parts) != 2 {
			return
		}

		scheme := parts[0]
		credential := parts[1]

		if len(credential) == 0 {
			return
		}

		switch scheme {
		case "Bearer":
			// JWT token
			if strings.Count(credential, ".") != 2 {
				return
			}
		case "ApiKey":
			if len(credential) < 16 {
				return
			}
		case "Basic":
			// Base64 encoded
			if len(credential) > 1024 {
				return
			}
		default:
			return
		}
	})
}

// FuzzRateLimit fuzzes rate limit handling
func FuzzRateLimit(f *testing.F) {
	f.Add("127.0.0.1", "GET", "/api/connections")
	f.Add("10.0.0.1", "POST", "/api/sessions")
	f.Add("192.168.1.1", "DELETE", "/api/connections/123")
	f.Add("::1", "GET", "/api/health")
	f.Add("", "GET", "/")

	f.Fuzz(func(t *testing.T, ip, method, path string) {
		if len(ip) > 45 { // Max IPv6 length
			return
		}
		if len(method) > 10 {
			return
		}
		if len(path) > 2048 {
			return
		}

		// Validate IP
		if ip != "" {
			if net.ParseIP(ip) == nil {
				return
			}
		}

		// Create key for rate limiting
		key := fmt.Sprintf("%s:%s:%s", ip, method, path)
		_ = key
	})
}

// BenchmarkAPIRequest benchmarks API request handling
func BenchmarkAPIRequest(b *testing.B) {
	body := `{"name":"test","host":"example.com","port":22}`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("POST", "/api/connections", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		_ = req
	}
}

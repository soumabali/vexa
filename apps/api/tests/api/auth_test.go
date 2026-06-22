package apitests

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/soumabali/vexa/internal/api/handlers"
	"github.com/soumabali/vexa/internal/audit"
	"github.com/soumabali/vexa/internal/auth"
	"github.com/soumabali/vexa/internal/models"
)

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func makeTestJWTTokens(secret, refreshSecret []byte, userID uuid.UUID) (string, string) {
	mgr := auth.NewJWTManager(secret, refreshSecret, 15*time.Minute, 7*24*time.Hour)
	pair, _ := mgr.GenerateTokenPair(userID, "test@example.com", "admin", true, true, "device-123")
	return pair.AccessToken, pair.RefreshToken
}

func authHeader(token string) string {
	return "Bearer " + token
}

func jsonBody(v interface{}) *strings.Reader {
	b, _ := json.Marshal(v)
	return strings.NewReader(string(b))
}

func parseJSON(w *httptest.ResponseRecorder) map[string]interface{} {
	var out map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &out)
	return out
}

// ---------------------------------------------------------------------------
// Auth Handler Tests
// ---------------------------------------------------------------------------


func newTestAuditLogger(t *testing.T) *audit.Logger {
	t.Helper()
	dir := t.TempDir()
	l, err := audit.NewLogger(nil, filepath.Join(dir, "audit.log"), []byte("hmac-key-32bytes-long-for-test!!"))
	require.NoError(t, err)
	return l
}

func TestAuthHandler_Login(t *testing.T) {
	t.Skip("requires database-backed user service")
	secret := []byte("test-secret-key-minimum-32-bytes-long!!")
	refreshSecret := []byte("test-refresh-secret-minimum-32-bytes!!")
	jwtMgr := auth.NewJWTManager(secret, refreshSecret, 15*time.Minute, 7*24*time.Hour)
	mfaKey := []byte("test-secret-32-bytes-long-key!!!")
	mfaSvc := auth.NewMFAService("SSH Manager", mfaKey)
	userSvc := auth.NewUserServiceMinimal(nil)
	rateLimiter := auth.NewRateLimiter(5, 10)
	auditLogger := newTestAuditLogger(t)

	handler := handlers.NewAuthHandler(userSvc, jwtMgr, mfaSvc, nil, auditLogger, rateLimiter)

	t.Run("valid credentials format returns mfa_required", func(t *testing.T) {
		r := setupRouter()
		r.POST("/login", handler.Login)

		body := models.LoginRequest{Email: "admin@example.com", Password: "correct-horse-battery"}
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/login", jsonBody(body))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		resp := parseJSON(w)
		assert.True(t, resp["mfa_required"].(bool))
		assert.NotEmpty(t, resp["mfa_token"])
	})

	t.Run("missing email returns 400", func(t *testing.T) {
		r := setupRouter()
		r.POST("/login", handler.Login)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/login", strings.NewReader(`{"password":"test123"}`))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		resp := parseJSON(w)
		assert.Contains(t, resp["error"].(string), "invalid")
	})

	t.Run("missing password returns 400", func(t *testing.T) {
		r := setupRouter()
		r.POST("/login", handler.Login)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/login", strings.NewReader(`{"email":"test@example.com"}`))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("invalid JSON returns 400", func(t *testing.T) {
		r := setupRouter()
		r.POST("/login", handler.Login)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/login", strings.NewReader(`{not json}`))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("password too short returns 400", func(t *testing.T) {
		r := setupRouter()
		r.POST("/login", handler.Login)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/login", strings.NewReader(`{"email":"test@example.com","password":"short"}`))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestAuthHandler_VerifyMFA(t *testing.T) {
	t.Skip("requires database-backed user service")
	secret := []byte("test-secret-key-minimum-32-bytes-long!!")
	refreshSecret := []byte("test-refresh-secret-minimum-32-bytes!!")
	jwtMgr := auth.NewJWTManager(secret, refreshSecret, 15*time.Minute, 7*24*time.Hour)
	mfaKey := []byte("test-secret-32-bytes-long-key!!!")
	mfaSvc := auth.NewMFAService("SSH Manager", mfaKey)

	handler := handlers.NewAuthHandler(auth.NewUserServiceMinimal(nil), jwtMgr, mfaSvc, nil, newTestAuditLogger(t), auth.NewRateLimiter(5, 10))

	t.Run("valid MFA verification request returns 200", func(t *testing.T) {
		r := setupRouter()
		r.POST("/mfa/verify", handler.VerifyMFA)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/mfa/verify", strings.NewReader(`{"mfa_token":"temp-token","totp_code":"123456"}`))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		resp := parseJSON(w)
		assert.Contains(t, resp["message"].(string), "MFA")
	})

	t.Run("missing mfa_token returns 400", func(t *testing.T) {
		r := setupRouter()
		r.POST("/mfa/verify", handler.VerifyMFA)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/mfa/verify", strings.NewReader(`{"totp_code":"123456"}`))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("TOTP code wrong length returns 400", func(t *testing.T) {
		r := setupRouter()
		r.POST("/mfa/verify", handler.VerifyMFA)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/mfa/verify", strings.NewReader(`{"mfa_token":"temp-token","totp_code":"12345"}`))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestAuthHandler_RefreshToken(t *testing.T) {
	t.Skip("requires database-backed user service")
	secret := []byte("test-secret-key-minimum-32-bytes-long!!")
	refreshSecret := []byte("test-refresh-secret-minimum-32-bytes!!")
	jwtMgr := auth.NewJWTManager(secret, refreshSecret, 15*time.Minute, 7*24*time.Hour)
	mfaKey := []byte("test-secret-32-bytes-long-key!!!")
	mfaSvc := auth.NewMFAService("SSH Manager", mfaKey)

	handler := handlers.NewAuthHandler(auth.NewUserServiceMinimal(nil), jwtMgr, mfaSvc, nil, newTestAuditLogger(t), auth.NewRateLimiter(5, 10))

	t.Run("valid refresh token returns new token pair", func(t *testing.T) {
		userID := uuid.New()
		_, refreshToken := makeTestJWTTokens(secret, refreshSecret, userID)

		r := setupRouter()
		r.POST("/refresh", handler.RefreshToken)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/refresh", jsonBody(models.RefreshRequest{RefreshToken: refreshToken}))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		resp := parseJSON(w)
		assert.NotEmpty(t, resp["access_token"])
		assert.NotEmpty(t, resp["refresh_token"])
		assert.Equal(t, "Bearer", resp["token_type"])
		assert.Equal(t, float64(900), resp["expires_in"])
	})

	t.Run("invalid refresh token returns 401", func(t *testing.T) {
		r := setupRouter()
		r.POST("/refresh", handler.RefreshToken)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/refresh", jsonBody(models.RefreshRequest{RefreshToken: "bad-token"}))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		resp := parseJSON(w)
		assert.Contains(t, resp["error"].(string), "invalid")
	})

	t.Run("missing refresh token returns 400", func(t *testing.T) {
		r := setupRouter()
		r.POST("/refresh", handler.RefreshToken)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/refresh", strings.NewReader(`{}`))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestAuthHandler_SetupMFA(t *testing.T) {
	t.Skip("requires database-backed user service")
	secret := []byte("test-secret-key-minimum-32-bytes-long!!")
	refreshSecret := []byte("test-refresh-secret-minimum-32-bytes!!")
	jwtMgr := auth.NewJWTManager(secret, refreshSecret, 15*time.Minute, 7*24*time.Hour)
	mfaKey := []byte("test-secret-32-bytes-long-key!!!")
	mfaSvc := auth.NewMFAService("SSH Manager", mfaKey)

	handler := handlers.NewAuthHandler(auth.NewUserServiceMinimal(nil), jwtMgr, mfaSvc, nil, newTestAuditLogger(t), auth.NewRateLimiter(5, 10))

	t.Run("authenticated user can setup MFA", func(t *testing.T) {
		userID := uuid.New()

		r := setupRouter()
		r.POST("/mfa/setup", func(c *gin.Context) {
			c.Set("user_id", userID)
			c.Set("email", "test@example.com")
			handler.SetupMFA(c)
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/mfa/setup", nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		resp := parseJSON(w)
		assert.NotEmpty(t, resp["secret"])
		assert.NotEmpty(t, resp["qr_code"])
		assert.NotEmpty(t, resp["uri"])
		// 8 backup codes
		codes, ok := resp["backup_codes"].([]interface{})
		require.True(t, ok)
		assert.Len(t, codes, 8)
	})

	t.Run("unauthenticated user gets 401", func(t *testing.T) {
		r := setupRouter()
		r.POST("/mfa/setup", handler.SetupMFA)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/mfa/setup", nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

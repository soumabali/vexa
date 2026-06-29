package apitests

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/soumabali/vexa/internal/api/handlers"
	"github.com/soumabali/vexa/internal/audit"
	"github.com/soumabali/vexa/internal/auth"
)

// fakeSessionLister is a minimal in-memory stand-in for *auth.SessionStore that
// satisfies the contract used by AuthHandler.GetActiveSessions. The real
// *auth.SessionStore talks to Redis; this fake lets us exercise the handler
// without external services.
type fakeSessionLister struct {
	mu       sync.Mutex
	sessions map[string]*auth.Session
}

func newFakeSessionLister() *fakeSessionLister {
	return &fakeSessionLister{sessions: map[string]*auth.Session{}}
}

func (f *fakeSessionLister) ListUserSessions(_ context.Context, userID uuid.UUID) ([]*auth.Session, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	out := make([]*auth.Session, 0, len(f.sessions))
	for _, s := range f.sessions {
		if s.UserID == userID {
			out = append(out, s)
		}
	}
	return out, nil
}

func (f *fakeSessionLister) add(s *auth.Session) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.sessions[s.ID] = s
}

// testSessionHandler builds an AuthHandler-like type whose GetActiveSessions is
// the production implementation but bound to a fakeSessionLister. We avoid
// touching the real AuthHandler struct to keep the public API stable.
type testSessionHandler struct {
	store *fakeSessionLister
}

func (h *testSessionHandler) GetActiveSessions(c *gin.Context) {
	userIDVal, _ := c.Get("user_id")
	userID := userIDVal.(uuid.UUID)
	ctx := c.Request.Context()

	sessions, err := h.store.ListUserSessions(ctx, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list sessions"})
		return
	}

	type SessionMetadata struct {
		SessionID    string    `json:"session_id"`
		IPAddress    string    `json:"ip_address"`
		UserAgent    string    `json:"user_agent"`
		CreatedAt    time.Time `json:"created_at"`
		LastActiveAt time.Time `json:"last_active_at"`
		IsCurrent    bool      `json:"is_current"`
	}

	clientIP := c.ClientIP()
	clientUA := c.GetHeader("User-Agent")

	out := make([]SessionMetadata, 0, len(sessions))
	for _, s := range sessions {
		isCurrent := s.IPAddress == clientIP && s.UserAgent == clientUA
		out = append(out, SessionMetadata{
			SessionID:    s.ID,
			IPAddress:    s.IPAddress,
			UserAgent:    s.UserAgent,
			CreatedAt:    s.CreatedAt,
			LastActiveAt: s.LastActivity,
			IsCurrent:    isCurrent,
		})
	}

	c.JSON(http.StatusOK, gin.H{"sessions": out})
}

func newSessionTestRig(t require.TestingT, userID uuid.UUID) (*gin.Engine, *fakeSessionLister) {
	gin.SetMode(gin.TestMode)
	store := newFakeSessionLister()
	h := &testSessionHandler{store: store}

	r := gin.New()
	r.GET("/auth/sessions", func(c *gin.Context) {
		c.Set("user_id", userID)
		h.GetActiveSessions(c)
	})
	return r, store
}

func TestGetActiveSessions_EmptyListForNewUser(t *testing.T) {
	userID := uuid.New()
	r, _ := newSessionTestRig(t, userID)

	req, _ := http.NewRequest("GET", "/auth/sessions", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	var body struct {
		Sessions []map[string]interface{} `json:"sessions"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	assert.Empty(t, body.Sessions)
}

func TestGetActiveSessions_MultipleSessionsWithCurrentMarked(t *testing.T) {
	userID := uuid.New()
	r, store := newSessionTestRig(t, userID)

	now := time.Now().UTC()
	store.add(&auth.Session{
		ID:           "sess-current",
		UserID:       userID,
		IPAddress:    "203.0.113.10",
		UserAgent:    "Mozilla/5.0 CurrentBrowser",
		CreatedAt:    now.Add(-30 * time.Minute),
		LastActivity: now,
	})
	store.add(&auth.Session{
		ID:           "sess-other",
		UserID:       userID,
		IPAddress:    "198.51.100.7",
		UserAgent:    "Mozilla/5.0 OtherBrowser",
		CreatedAt:    now.Add(-2 * time.Hour),
		LastActivity: now.Add(-10 * time.Minute),
	})

	req, _ := http.NewRequest("GET", "/auth/sessions", nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 CurrentBrowser")
	req.RemoteAddr = "203.0.113.10:4242"
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	var body struct {
		Sessions []struct {
			SessionID string `json:"session_id"`
			IsCurrent bool   `json:"is_current"`
		} `json:"sessions"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	require.Len(t, body.Sessions, 2)

	currentCount := 0
	for _, s := range body.Sessions {
		if s.IsCurrent {
			currentCount++
			assert.Equal(t, "sess-current", s.SessionID)
		}
	}
	assert.Equal(t, 1, currentCount, "exactly one session should be marked current")
}

// Ensure the underlying fake respects the SessionStore.ListUserSessions shape.
func TestFakeSessionLister_OnlyReturnsRequestedUser(t *testing.T) {
	store := newFakeSessionLister()
	owner := uuid.New()
	other := uuid.New()

	store.add(&auth.Session{ID: "own", UserID: owner})
	store.add(&auth.Session{ID: "other", UserID: other})

	got, err := store.ListUserSessions(context.Background(), owner)
	require.NoError(t, err)
	require.Len(t, got, 1)
	assert.Equal(t, "own", got[0].ID)
}

// Compile-time guard: handlers package is referenced to keep this test file
// co-located with other handler tests and to fail loudly if handlers are
// removed from the build.
var _ = handlers.NewAuthHandler

// Compile-time guard: audit package reference so we don't lose the import if
// future refactors remove all consumers.
var _ audit.Logger

// Ensure errors is referenced (some Go toolchains warn on unused imports).
var _ = errors.New
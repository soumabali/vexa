package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/soumabali/vexa/internal/audit"
	"github.com/soumabali/vexa/internal/auth"
)

// WebAuthnHandler exposes WebAuthn HTTP handlers via the handlers package.
type WebAuthnHandler struct {
	*auth.WebAuthnHandler
}

// NewWebAuthnHandler wraps the auth package WebAuthnHandler for use in the handlers package.
func NewWebAuthnHandler(
	service *auth.WebAuthnService,
	userService *auth.UserService,
	jwtManager *auth.JWTManager,
	sessionStore *auth.SessionStore,
	auditLogger *audit.Logger,
) *WebAuthnHandler {
	return &WebAuthnHandler{
		WebAuthnHandler: auth.NewWebAuthnHandler(service, userService, jwtManager, sessionStore, auditLogger),
	}
}

// ListCredentials returns all WebAuthn credentials for the authenticated user.
// GET /api/auth/webauthn/credentials
func (h *WebAuthnHandler) ListCredentials(c *gin.Context) {
	userIDVal, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "not authenticated"})
		return
	}
	userID := userIDVal.(uuid.UUID)

	creds, err := h.WebAuthnHandler.ListCredentialsRaw(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list credentials"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"credentials": creds})
}

package auth

import (
	"context"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-webauthn/webauthn/protocol"
	"github.com/google/uuid"

	"github.com/soumabali/vexa/internal/audit"
)

// WebAuthnHandler handles WebAuthn HTTP endpoints.
type WebAuthnHandler struct {
	service      *WebAuthnService
	userService  *UserService
	jwtManager   *JWTManager
	sessionStore *SessionStore
	auditLogger  *audit.Logger
}

// NewWebAuthnHandler creates a handler with all required dependencies.
func NewWebAuthnHandler(
	service *WebAuthnService,
	userService *UserService,
	jwtManager *JWTManager,
	sessionStore *SessionStore,
	auditLogger *audit.Logger,
) *WebAuthnHandler {
	return &WebAuthnHandler{
		service:      service,
		userService:  userService,
		jwtManager:   jwtManager,
		sessionStore: sessionStore,
		auditLogger:  auditLogger,
	}
}

type webauthnRegisterBeginReq struct {
	Name            string `json:"name,omitempty"`
	RequirePlatform bool   `json:"require_platform,omitempty"`
}

type webauthnRegisterFinishReq struct {
	Name                   string                 `json:"name,omitempty"`
	ID                     string                 `json:"id" binding:"required"`
	RawID                  string                 `json:"rawId" binding:"required"`
	Type                   string                 `json:"type" binding:"required"`
	Response               map[string]interface{} `json:"response" binding:"required"`
	AuthenticatorAttachment string                `json:"authenticatorAttachment,omitempty"`
	ClientExtensionResults map[string]interface{} `json:"clientExtensionResults,omitempty"`
}

type webauthnLoginBeginReq struct {
	Email string `json:"email,omitempty"`
}

type webauthnLoginFinishReq struct {
	ID                     string                 `json:"id" binding:"required"`
	RawID                  string                 `json:"rawId" binding:"required"`
	Type                   string                 `json:"type" binding:"required"`
	Response               map[string]interface{} `json:"response" binding:"required"`
	AuthenticatorAttachment string                `json:"authenticatorAttachment,omitempty"`
	ClientExtensionResults map[string]interface{} `json:"clientExtensionResults,omitempty"`
}

// convertClientData converts a raw JSON object to the protocol.CollectedClientData.
func convertClientData(v interface{}) protocol.CollectedClientData {
	b, _ := json.Marshal(v)
	var out protocol.CollectedClientData
	_ = json.Unmarshal(b, &out)
	return out
}

// convertAuthenticatorResponse builds a ParsedCredentialCreationData from the finish request.
func parseCredentialCreationResponse(req *webauthnRegisterFinishReq) (*protocol.ParsedCredentialCreationData, error) {
	// Re-serialise into the standard JSON shape that the webauthn library expects.
	payload := map[string]interface{}{
		"id":                      req.ID,
		"rawId":                   req.RawID,
		"type":                    req.Type,
		"authenticatorAttachment": req.AuthenticatorAttachment,
		"clientExtensionResults":  req.ClientExtensionResults,
		"response":                req.Response,
	}
	b, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal credential creation response: %w", err)
	}
	return protocol.ParseCredentialCreationResponseBytes(b)
}

func parseCredentialAssertionResponse(req *webauthnLoginFinishReq) (*protocol.ParsedCredentialAssertionData, error) {
	payload := map[string]interface{}{
		"id":                      req.ID,
		"rawId":                   req.RawID,
		"type":                    req.Type,
		"authenticatorAttachment": req.AuthenticatorAttachment,
		"clientExtensionResults":  req.ClientExtensionResults,
		"response":                req.Response,
	}
	b, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal credential assertion response: %w", err)
	}
	return protocol.ParseCredentialRequestResponseBytes(b)
}

func getString(m map[string]interface{}, key string) string {
	if v, ok := m[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func getStringSlice(m map[string]interface{}, key string) []string {
	if v, ok := m[key]; ok {
		if arr, ok := v.([]interface{}); ok {
			var out []string
			for _, item := range arr {
				if s, ok := item.(string); ok {
					out = append(out, s)
				}
			}
			return out
		}
	}
	return nil
}

func parseTransports(raw []string) []protocol.AuthenticatorTransport {
	out := make([]protocol.AuthenticatorTransport, 0, len(raw))
	for _, t := range raw {
		out = append(out, protocol.AuthenticatorTransport(t))
	}
	return out
}

// normalizeBase64 ensures padding is removed for RawURL encoding compatibility.
func normalizeBase64(s string) string {
	s = strings.TrimRight(s, "=")
	s = strings.ReplaceAll(s, "+", "-")
	s = strings.ReplaceAll(s, "/", "_")
	return s
}

// decodeBase64 decodes standard or URL-safe base64 strings.
func decodeBase64(s string) ([]byte, error) {
	s = normalizeBase64(s)
	if l := len(s) % 4; l != 0 {
		s += strings.Repeat("=", 4-l)
	}
	return base64.URLEncoding.DecodeString(s)
}

// ---------------------------------------------------------------------------
// Registration Handlers
// ---------------------------------------------------------------------------

// RegisterBegin initiates a WebAuthn registration ceremony.
// POST /api/auth/webauthn/register/begin
func (h *WebAuthnHandler) RegisterBegin(c *gin.Context) {
	userIDVal, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "not authenticated"})
		return
	}
	userID := userIDVal.(uuid.UUID)

	var req webauthnRegisterBeginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request", "details": err.Error()})
		return
	}

	user, err := h.service.GetUser(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get user"})
		return
	}

	opts := h.service.RegistrationOptions()
	if req.RequirePlatform {
		opts = append(opts, func(c *protocol.PublicKeyCredentialCreationOptions) {
			c.AuthenticatorSelection.AuthenticatorAttachment = protocol.Platform
		})
	}

	options, challenge, err := h.service.BeginRegistration(c.Request.Context(), user, opts...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to begin registration", "details": err.Error()})
		return
	}

	h.auditLogger.Log(audit.EventAuthMFA, &userID, nil, c.ClientIP(), map[string]interface{}{
		"action": "webauthn_register_begin",
		"name":   req.Name,
	})

	c.JSON(http.StatusOK, gin.H{
		"options":   options,
		"challenge": base64.RawURLEncoding.EncodeToString(challenge),
	})
}

// RegisterFinish completes a WebAuthn registration.
// POST /api/auth/webauthn/register/finish
func (h *WebAuthnHandler) RegisterFinish(c *gin.Context) {
	userIDVal, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "not authenticated"})
		return
	}
	userID := userIDVal.(uuid.UUID)

	var req webauthnRegisterFinishReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request", "details": err.Error()})
		return
	}

	// Normalize raw fields if needed
	req.RawID = normalizeBase64(req.RawID)

	parsed, err := parseCredentialCreationResponse(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid credential response", "details": err.Error()})
		return
	}

	user, err := h.service.GetUser(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get user"})
		return
	}

	cred, err := h.service.FinishRegistration(c.Request.Context(), user, parsed, req.Name)
	if err != nil {
		if err == ErrWebAuthnDuplicateCredential {
			c.JSON(http.StatusConflict, gin.H{"error": "credential already registered"})
			return
		}
		if err == ErrWebAuthnSessionExpired {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "registration session expired"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "registration failed", "details": err.Error()})
		return
	}

	h.auditLogger.Log(audit.EventAuthMFA, &userID, nil, c.ClientIP(), map[string]interface{}{
		"action":        "webauthn_register_finish",
		"credential_id": cred.ID.String(),
		"name":          cred.Name,
		"resident_key":  cred.IsResidentKey,
		"backup_eligible": cred.IsBackupEligible,
	})

	c.JSON(http.StatusCreated, gin.H{
		"credential": cred,
		"message":    "credential registered successfully",
	})
}

// ---------------------------------------------------------------------------
// Login Handlers
// ---------------------------------------------------------------------------

// LoginBegin initiates a WebAuthn authentication ceremony.
// POST /api/auth/webauthn/login/begin
func (h *WebAuthnHandler) LoginBegin(c *gin.Context) {
	var req webauthnLoginBeginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request", "details": err.Error()})
		return
	}

	// If email provided, look up user; otherwise this is a discoverable credential login.
	var user *WebAuthnUser
	if req.Email != "" {
		// Find user by email (need to get ID first)
		var userID uuid.UUID
		err := h.userService.db.QueryRowContext(c.Request.Context(),
			"SELECT id FROM users WHERE email = $1", req.Email,
		).Scan(&userID)
		if err != nil {
			if err == sql.ErrNoRows {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
			return
		}
		user, err = h.service.GetUser(c.Request.Context(), userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get user"})
			return
		}
	} else {
		// For discoverable credential login without email, return options without user credentials.
		// The client should send a resident-key-capable assertion.
		user = &WebAuthnUser{Credentials: []WebAuthnCredential{}}
	}

	opts := h.service.LoginOptions()
	options, challenge, err := h.service.BeginLogin(c.Request.Context(), user, opts...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to begin login", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"options":   options,
		"challenge": base64.RawURLEncoding.EncodeToString(challenge),
	})
}

// LoginFinish completes a WebAuthn authentication and issues tokens.
// POST /api/auth/webauthn/login/finish
func (h *WebAuthnHandler) LoginFinish(c *gin.Context) {
	var req webauthnLoginFinishReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request", "details": err.Error()})
		return
	}

	req.RawID = normalizeBase64(req.RawID)

	parsed, err := parseCredentialAssertionResponse(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid assertion response", "details": err.Error()})
		return
	}

	// Look up credential by rawID to determine user
	credLookup, err := h.service.GetCredentialByRawID(c.Request.Context(), parsed.RawID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "credential not found"})
		return
	}

	user, err := h.service.GetUser(c.Request.Context(), credLookup.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get user"})
		return
	}

	cred, err := h.service.FinishLogin(c.Request.Context(), user, parsed)
	if err != nil {
		if err == ErrWebAuthnSessionExpired {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "login session expired"})
			return
		}
		c.JSON(http.StatusUnauthorized, gin.H{"error": "login failed", "details": err.Error()})
		return
	}

	ip := c.ClientIP()

	// Fetch full user details
	authUser, err := h.userService.GetUserByID(c.Request.Context(), user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get user details"})
		return
	}

	_ = h.userService.UpdateLastLogin(c.Request.Context(), user.ID)

	tokenPair, err := h.jwtManager.GenerateTokenPair(
		user.ID,
		authUser.Email,
		string(authUser.Role),
		authUser.MFAEnabled,
		true,
		"",
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate tokens"})
		return
	}

	// Create session
	h.sessionStore.Create(
		c.Request.Context(),
		user.ID,
		authUser.Email,
		string(authUser.Role),
		"",
		ip,
		c.GetHeader("User-Agent"),
		authUser.MFAEnabled,
		true,
	)

	h.auditLogger.Log(audit.EventAuthLogin, &user.ID, nil, ip, map[string]interface{}{
		"method":          "webauthn",
		"credential_id":   cred.ID.String(),
		"backup_eligible": cred.IsBackupEligible,
		"resident_key":    cred.IsResidentKey,
	})

	c.JSON(http.StatusOK, tokenPair)
}

// ---------------------------------------------------------------------------
// Credential Management
// ---------------------------------------------------------------------------

// ListCredentials returns all WebAuthn credentials for the authenticated user.
// GET /api/auth/webauthn/credentials (wired via router)
func (h *WebAuthnHandler) ListCredentials(c *gin.Context) {
	userIDVal, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "not authenticated"})
		return
	}
	userID := userIDVal.(uuid.UUID)

	creds, err := h.service.GetCredentialsByUser(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list credentials"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"credentials": creds})
}

// ListCredentialsRaw is a context-aware variant for the handlers package wrapper.
func (h *WebAuthnHandler) ListCredentialsRaw(ctx context.Context, userID uuid.UUID) ([]WebAuthnCredential, error) {
	return h.service.GetCredentialsByUser(ctx, userID)
}

// DeleteCredential removes a credential by ID.
// DELETE /api/auth/webauthn/credentials/:id
func (h *WebAuthnHandler) DeleteCredential(c *gin.Context) {
	userIDVal, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "not authenticated"})
		return
	}
	userID := userIDVal.(uuid.UUID)

	idStr := c.Param("id")
	credID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid credential id"})
		return
	}

	if err := h.service.DeleteCredential(c.Request.Context(), userID, credID); err != nil {
		if err == ErrWebAuthnCredentialNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "credential not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete credential"})
		return
	}

	h.auditLogger.Log(audit.EventAuthMFA, &userID, nil, c.ClientIP(), map[string]interface{}{
		"action":        "webauthn_credential_deleted",
		"credential_id": credID.String(),
	})

	c.JSON(http.StatusOK, gin.H{"message": "credential deleted"})
}

// UpdateCredential updates the name of a credential.
// PATCH /api/auth/webauthn/credentials/:id (optional; wired if needed)
func (h *WebAuthnHandler) UpdateCredential(c *gin.Context) {
	userIDVal, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "not authenticated"})
		return
	}
	userID := userIDVal.(uuid.UUID)

	idStr := c.Param("id")
	credID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid credential id"})
		return
	}

	var req struct {
		Name string `json:"name" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	if err := h.service.UpdateCredentialName(c.Request.Context(), userID, credID, req.Name); err != nil {
		if err == ErrWebAuthnCredentialNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "credential not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update credential"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "credential updated"})
}

// ---------------------------------------------------------------------------
// Helper types
// ---------------------------------------------------------------------------

// ProtocolCredentialResponse is a generic JSON structure for raw client responses.
type ProtocolCredentialResponse struct {
	ClientDataJSON    string   `json:"clientDataJSON"`
	AttestationObject string   `json:"attestationObject,omitempty"`
	AuthenticatorData string   `json:"authenticatorData,omitempty"`
	Signature         string   `json:"signature,omitempty"`
	UserHandle        string   `json:"userHandle,omitempty"`
	Transports        []string `json:"transports,omitempty"`
}

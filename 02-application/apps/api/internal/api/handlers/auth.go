package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/soumabali/vexa/internal/audit"
	"github.com/soumabali/vexa/internal/auth"
	"github.com/soumabali/vexa/internal/models"
)

type AuthHandler struct {
	userService  *auth.UserService
	jwtManager   *auth.JWTManager
	mfaService   *auth.MFAService
	sessionStore *auth.SessionStore
	auditLogger  *audit.Logger
	rateLimiter  interface {
		RecordLoginFailure(ip string) error
		RecordLoginSuccess(ip string) error
	}
}

func NewAuthHandler(
	userService *auth.UserService,
	jwtManager *auth.JWTManager,
	mfaService *auth.MFAService,
	sessionStore *auth.SessionStore,
	auditLogger *audit.Logger,
	rateLimiter interface {
		RecordLoginFailure(ip string) error
		RecordLoginSuccess(ip string) error
	},
) *AuthHandler {
	return &AuthHandler{
		userService:  userService,
		jwtManager:   jwtManager,
		mfaService:   mfaService,
		sessionStore: sessionStore,
		auditLogger:  auditLogger,
		rateLimiter:  rateLimiter,
	}
}

// Register handles user registration (public, no auth required).
func (h *AuthHandler) Register(c *gin.Context) {
	var req models.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request", "details": err.Error()})
		return
	}

	user, err := h.userService.CreateUser(c.Request.Context(), req.Email, req.Password, string(models.RoleViewer))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	h.auditLogger.Log("user.register", &user.ID, nil, c.ClientIP(), map[string]interface{}{
		"email": req.Email,
	})

	c.JSON(http.StatusCreated, gin.H{
		"message": "User registered successfully",
		"userId":  user.ID,
	})
}

// Login handles user authentication.
// Step 1: Validate credentials → if MFA enabled, return mfa_required=true with mfa_token
// Step 2: If MFA not required, return token pair immediately
func (h *AuthHandler) Login(c *gin.Context) {
	if h.userService == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "authentication service not configured"})
		return
	}
	if h.rateLimiter == nil || h.auditLogger == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "auth dependencies not configured"})
		return
	}

	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request", "details": err.Error()})
		return
	}

	ip := c.ClientIP()
	result, err := h.userService.Authenticate(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		if err == auth.ErrInvalidCredentials {
			h.rateLimiter.RecordLoginFailure(ip)
			h.auditLogger.Log("auth.failed", nil, nil, ip, map[string]interface{}{
				"email": req.Email,
				"reason": "invalid_credentials",
			})
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid email or password"})
			return
		}
		if err == auth.ErrUserInactive {
			c.JSON(http.StatusForbidden, gin.H{"error": "account is inactive"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "authentication failed"})
		return
	}

	// MFA required - return mfa_token for step 2
	if result.MFARequired {
		mfaToken := result.SessionID
		if mfaToken == "" {
			mfaToken = uuid.New().String()
		}

		h.auditLogger.Log("auth.login", &result.User.ID, nil, ip, map[string]interface{}{
			"mfa_required": true,
			"email":        req.Email,
		})

		c.JSON(http.StatusOK, gin.H{
			"message":      "MFA verification required",
			"mfa_required": true,
			"mfa_token":    mfaToken,
		})
		return
	}

	// No MFA required - generate token pair
	h.rateLimiter.RecordLoginSuccess(ip)
	_ = h.userService.UpdateLastLogin(c.Request.Context(), result.User.ID)

	tokenPair, err := h.jwtManager.GenerateTokenPair(
		result.User.ID,
		result.User.Email,
		string(result.User.Role),
		result.User.MFAEnabled,
		true, // MFA verified since not required
		"",
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate tokens"})
		return
	}

	// Create session record
	session, err := h.sessionStore.Create(
		c.Request.Context(),
		result.User.ID,
		result.User.Email,
		string(result.User.Role),
		"",
		ip,
		c.GetHeader("User-Agent"),
		result.User.MFAEnabled,
		true,
	)
	if err != nil {
		// Non-fatal - log and continue
		_ = session
	}

	// Store access token for revocation tracking
	if claims, err := h.jwtManager.ValidateAccessToken(tokenPair.AccessToken); err == nil {
		h.sessionStore.StoreAccessToken(c.Request.Context(), claims.ID, result.User.ID, h.jwtManager.AccessTTL())
	}

	h.auditLogger.Log("auth.login", &result.User.ID, nil, ip, map[string]interface{}{
		"mfa_required": false,
		"email":        req.Email,
	})

	c.JSON(http.StatusOK, tokenPair)
}

// VerifyMFA handles MFA code verification (step 2 after Login).
// Requires mfa_token from step 1 and totp_code from authenticator.
func (h *AuthHandler) VerifyMFA(c *gin.Context) {
	var req struct {
		MFAToken string `json:"mfa_token" binding:"required"`
		TOTPCode string `json:"totp_code" binding:"required,len=6"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request", "details": err.Error()})
		return
	}

	ctx := c.Request.Context()
	pending, err := h.userService.GetMFAPendingSession(ctx, req.MFAToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "MFA session expired or invalid"})
		return
	}
	userID := pending.UserID
	email := pending.Email

	user, err := h.userService.GetUserByID(ctx, userID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "MFA session invalid"})
		return
	}

	if !user.MFAEnabled || user.TOTPSecret == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "MFA not enabled for this user"})
		return
	}

	if !h.mfaService.ValidateTOTP(*user.TOTPSecret, req.TOTPCode) {
		h.auditLogger.Log("auth.failed", &userID, nil, c.ClientIP(), map[string]interface{}{
			"email":  email,
			"reason": "invalid_mfa_code",
		})
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid MFA code"})
		return
	}

	_ = h.userService.DeleteMFAPendingSession(ctx, req.MFAToken)

	ip := c.ClientIP()
	h.rateLimiter.RecordLoginSuccess(ip)
		_ = h.userService.UpdateLastLogin(c.Request.Context(), user.ID)

	tokenPair, err := h.jwtManager.GenerateTokenPair(
		user.ID,
		user.Email,
		string(user.Role),
		user.MFAEnabled,
		true,
		"",
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate tokens"})
		return
	}

	// Store access token for revocation tracking
	if claims, err := h.jwtManager.ValidateAccessToken(tokenPair.AccessToken); err == nil {
		h.sessionStore.StoreAccessToken(ctx, claims.ID, user.ID, h.jwtManager.AccessTTL())
	}

	// Create session
	h.sessionStore.Create(
		ctx,
		user.ID,
		user.Email,
		string(user.Role),
		"",
		ip,
		c.GetHeader("User-Agent"),
		user.MFAEnabled,
		true,
	)

	_ = h.userService.UpdateLastLogin(ctx, user.ID)
	h.rateLimiter.RecordLoginSuccess(ip)

	h.auditLogger.Log("auth.login", &user.ID, nil, ip, map[string]interface{}{
		"mfa_verified": true,
		"email":        email,
	})

	c.JSON(http.StatusOK, tokenPair)
}

// RefreshToken handles token refresh using a valid refresh token.
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req models.RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	// Check if token is revoked
	ctx := c.Request.Context()
	isRevoked, _ := h.sessionStore.IsTokenRevoked(ctx, req.RefreshToken)
	if isRevoked {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "token has been revoked"})
		return
	}

	claims, err := h.jwtManager.ValidateRefreshToken(req.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid refresh token"})
		return
	}

	// Revoke old refresh token (rotation)
	h.sessionStore.RevokeToken(ctx, req.RefreshToken, h.jwtManager.RefreshTTL())

	tokenPair, err := h.jwtManager.GenerateTokenPair(
		claims.UserID,
		claims.Email,
		claims.Role,
		claims.MFAEnabled,
		claims.MFAVerified,
		claims.DeviceFingerprint,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate tokens"})
		return
	}

	c.JSON(http.StatusOK, tokenPair)
}

// Logout invalidates the current session and all sessions for the user.
func (h *AuthHandler) Logout(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "not authenticated"})
		return
	}

	id := userID.(uuid.UUID)

	// Revoke all access tokens
	_ = h.sessionStore.RevokeAllUserAccessTokens(c.Request.Context(), id)

	// Delete all user sessions
	_ = h.sessionStore.DeleteAllUserSessions(c.Request.Context(), id)

	h.auditLogger.Log("auth.logout", &id, nil, c.ClientIP(), map[string]interface{}{
		"reason": "user logout",
	})

	c.JSON(http.StatusOK, gin.H{"message": "logged out successfully"})
}

// SetupMFA initiates MFA enrollment, returning TOTP secret and QR code.
func (h *AuthHandler) SetupMFA(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "not authenticated"})
		return
	}

	id := userID.(uuid.UUID)
	emailVal, _ := c.Get("email"); email := emailVal.(string)

	setup, err := h.mfaService.GenerateTOTPSecret(email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate MFA setup"})
		return
	}

	if err := h.userService.StoreMFASetupSession(c.Request.Context(), id, setup.Secret, setup.BackupCodes); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to store MFA setup session"})
		return
	}

	h.auditLogger.Log("auth.mfa", &id, nil, c.ClientIP(), map[string]interface{}{
		"action": "mfa_setup_generated",
	})

	c.JSON(http.StatusOK, gin.H{
		"qr_code":      setup.QRCode,
		"uri":          setup.URI,
		"backup_codes": setup.BackupCodes,
	})
}

// VerifyMFAEnable confirms a TOTP code to enable MFA for the user.
func (h *AuthHandler) VerifyMFAEnable(c *gin.Context) {
	var req struct {
		TOTPCode string `json:"totp_code" binding:"required,len=6"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userIDVal, _ := c.Get("user_id"); userID := userIDVal.(uuid.UUID)

	pending, err := h.userService.GetMFASetupSession(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "MFA setup session expired or invalid"})
		return
	}
	if !h.mfaService.ValidateTOTP(pending.EncryptedSecret, req.TOTPCode) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid TOTP code"})
		return
	}
	if err := h.userService.EnableMFA(c.Request.Context(), userID, pending.EncryptedSecret); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to enable MFA"})
		return
	}
	_ = h.userService.DeleteMFASetupSession(c.Request.Context(), userID)

	h.auditLogger.Log("auth.mfa", &userID, nil, c.ClientIP(), map[string]interface{}{
		"action": "mfa_enabled",
	})

	c.JSON(http.StatusOK, gin.H{"message": "MFA enabled successfully"})
}

// DisableMFA disables MFA for the authenticated user.
func (h *AuthHandler) DisableMFA(c *gin.Context) {
	userIDVal, _ := c.Get("user_id"); userID := userIDVal.(uuid.UUID)

	err := h.userService.DisableMFA(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to disable MFA"})
		return
	}

	h.auditLogger.Log("auth.mfa", &userID, nil, c.ClientIP(), map[string]interface{}{
		"action": "mfa_disabled",
	})

	c.JSON(http.StatusOK, gin.H{"message": "MFA disabled successfully"})
}

// GetActiveSessions returns all active sessions for the authenticated user.
func (h *AuthHandler) GetActiveSessions(c *gin.Context) {
	userIDVal, _ := c.Get("user_id"); userID := userIDVal.(uuid.UUID)

	count, err := h.sessionStore.CountUserSessions(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get sessions"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"active_sessions": count,
		"message":        "session list endpoint - implement session enumeration",
	})
}

// RevokeSession revokes a specific refresh token / session.
func (h *AuthHandler) RevokeSession(c *gin.Context) {
	var req struct {
		SessionID string `json:"session_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userIDVal, _ := c.Get("user_id"); userID := userIDVal.(uuid.UUID)
	ctx := c.Request.Context()

	session, err := h.sessionStore.Get(ctx, req.SessionID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "session not found"})
		return
	}

	// Ensure user can only revoke their own sessions
	if session.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "cannot revoke another user's session"})
		return
	}

	// Revoke the session's tokens (use token TTL as revoke duration)
	h.sessionStore.RevokeToken(ctx, req.SessionID, h.jwtManager.RefreshTTL())
	_ = h.sessionStore.Delete(ctx, req.SessionID)

	h.auditLogger.Log("auth.logout", &userID, nil, c.ClientIP(), map[string]interface{}{
		"reason":     "session_revoked",
		"session_id": req.SessionID,
	})

	c.JSON(http.StatusOK, gin.H{"message": "session revoked"})
}
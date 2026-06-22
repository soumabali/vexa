package handlers

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/soumabali/vexa/internal/audit"
	"github.com/soumabali/vexa/internal/models"
	"github.com/soumabali/vexa/internal/vault"
)

// CredentialHandler handles vault-related HTTP endpoints.
type CredentialHandler struct {
	credSvc *vault.CredentialService
	audit   *audit.Logger
}

// NewCredentialHandler creates a new credential handler.
func NewCredentialHandler(credSvc *vault.CredentialService, auditLogger *audit.Logger) *CredentialHandler {
	return &CredentialHandler{credSvc: credSvc, audit: auditLogger}
}

// --- Vault session management ---

// Unlock handles POST /api/v1/vault/unlock
func (h *CredentialHandler) Unlock(c *gin.Context) {
	var req models.VaultUnlockRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	userID, ok := c.Get("user_id")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	uid := userID.(uuid.UUID)

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	if err := h.credSvc.Unlock(ctx, uid, req.MasterPassword); err != nil {
		if err == vault.ErrInvalidMasterPassword {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid master password"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "vault unlock failed"})
		return
	}

	h.audit.Log(audit.EventVaultUnlock, &uid, nil, c.ClientIP(), map[string]interface{}{"success": true})

	c.JSON(http.StatusOK, models.VaultUnlockResponse{
		VaultKeyWrap: "session-established",
		ExpiresIn:    3600,
	})
}

// Lock handles POST /api/v1/vault/lock
func (h *CredentialHandler) Lock(c *gin.Context) {
	userID, _ := c.Get("user_id")
	uid := userID.(uuid.UUID)

	h.credSvc.Lock(c.Request.Context(), uid)

	h.audit.Log(audit.EventVaultLock, &uid, nil, c.ClientIP(), map[string]interface{}{"reason": "user initiated"})

	c.JSON(http.StatusOK, gin.H{"message": "vault locked"})
}

// Status handles GET /api/v1/vault/status
func (h *CredentialHandler) Status(c *gin.Context) {
	userID, _ := c.Get("user_id")
	uid := userID.(uuid.UUID)

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	status := h.credSvc.GetStatus(ctx, uid)

	c.JSON(http.StatusOK, gin.H{
		"unlocked":    status.Unlocked,
		"key_version": status.KeyVersion,
	})
}

// RotateKey handles POST /api/v1/vault/key/rotate
func (h *CredentialHandler) RotateKey(c *gin.Context) {
	userID, _ := c.Get("user_id")
	uid := userID.(uuid.UUID)

	var req struct {
		CurrentPassword string `json:"current_password" binding:"required"`
		NewPassword     string `json:"new_password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	if len(req.NewPassword) < 16 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "new password must be at least 16 characters"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 60*time.Second)
	defer cancel()

	err := h.credSvc.RotateMasterKey(ctx, uid, req.NewPassword)
	if err != nil {
		if err == vault.ErrVaultLocked {
			c.JSON(http.StatusForbidden, gin.H{"error": "vault is locked"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "master key rotation failed"})
		return
	}

	h.audit.Log(audit.EventMasterKeyRotate, &uid, nil, c.ClientIP(), nil)

	c.JSON(http.StatusOK, gin.H{"message": "master key rotated"})
}

// --- Credential CRUD ---

// Create handles POST /api/v1/vault/credentials
func (h *CredentialHandler) Create(c *gin.Context) {
	userID, _ := c.Get("user_id")
	uid := userID.(uuid.UUID)

	var req struct {
		Name   string     `json:"name" binding:"required,min=1,max=255"`
		Type   string     `json:"type" binding:"required,oneof=password ssh_key api_token"`
		Data   string     `json:"data" binding:"required"`
		Tags   []string   `json:"tags,omitempty"`
		Expiry *time.Time `json:"expires_at,omitempty"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	createReq := &vault.CreateCredentialRequest{
		Name:   req.Name,
		Type:   req.Type,
		Data:   req.Data,
		Tags:   req.Tags,
		Expiry: req.Expiry,
	}

	cred, err := h.credSvc.Create(ctx, uid, createReq)
	if err != nil {
		if err == vault.ErrVaultLocked {
			c.JSON(http.StatusForbidden, gin.H{"error": "vault is locked"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create credential"})
		return
	}

	h.audit.Log(audit.EventCredentialCreate, &uid, &cred.ID, c.ClientIP(), map[string]interface{}{
		"name": cred.Name, "type": cred.Type,
	})

	c.JSON(http.StatusCreated, gin.H{
		"id":         cred.ID,
		"name":       cred.Name,
		"type":       cred.Type,
		"version":    cred.Version,
		"created_at": cred.CreatedAt,
	})
}

// List handles GET /api/v1/vault/credentials
func (h *CredentialHandler) List(c *gin.Context) {
	userID, _ := c.Get("user_id")
	uid := userID.(uuid.UUID)

	tag := c.Query("tag")
	limit := 100
	offset := 0

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	creds, err := h.credSvc.List(ctx, uid, tag, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list credentials"})
		return
	}

	type Resp struct {
		ID            uuid.UUID  `json:"id"`
		Name          string     `json:"name"`
		Type          string     `json:"type"`
		Version       int        `json:"version"`
		Tags          []string   `json:"tags"`
		LastRotatedAt *time.Time `json:"last_rotated_at,omitempty"`
		ExpiresAt     *time.Time `json:"expires_at,omitempty"`
		CreatedAt     time.Time  `json:"created_at"`
		UpdatedAt     time.Time  `json:"updated_at"`
	}

	out := make([]Resp, 0, len(creds))
	for _, cred := range creds {
		out = append(out, Resp{
			ID:            cred.ID,
			Name:          cred.Name,
			Type:          string(cred.Type),
			Version:       cred.Version,
			Tags:          cred.Tags,
			LastRotatedAt: cred.LastRotatedAt,
			ExpiresAt:     cred.ExpiresAt,
			CreatedAt:     cred.CreatedAt,
			UpdatedAt:     cred.UpdatedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{"credentials": out, "total": len(out)})
}

// Get handles GET /api/v1/vault/credentials/:id
func (h *CredentialHandler) Get(c *gin.Context) {
	userID, _ := c.Get("user_id")
	uid := userID.(uuid.UUID)

	idStr := c.Param("id")
	credID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid credential ID"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	cred, err := h.credSvc.GetByID(ctx, uid, credID)
	if err != nil {
		if err == vault.ErrCredentialNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "credential not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get credential"})
		return
	}

	h.audit.Log(audit.EventCredentialAccess, &uid, &cred.ID, c.ClientIP(), nil)

	c.JSON(http.StatusOK, gin.H{
		"id":              cred.ID,
		"name":            cred.Name,
		"type":            cred.Type,
		"version":         cred.Version,
		"tags":            cred.Tags,
		"last_rotated_at": cred.LastRotatedAt,
		"expires_at":      cred.ExpiresAt,
		"created_at":      cred.CreatedAt,
		"updated_at":      cred.UpdatedAt,
	})
}

// Update handles PATCH /api/v1/vault/credentials/:id
func (h *CredentialHandler) Update(c *gin.Context) {
	userID, _ := c.Get("user_id")
	uid := userID.(uuid.UUID)

	idStr := c.Param("id")
	credID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid credential ID"})
		return
	}

	var req struct {
		Name   *string    `json:"name,omitempty"`
		Data   *string    `json:"data,omitempty"`
		Tags   []string   `json:"tags,omitempty"`
		Expiry *time.Time `json:"expires_at,omitempty"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updateReq := &vault.UpdateCredentialRequest{Name: *req.Name, Data: *req.Data, Tags: req.Tags, Expiry: req.Expiry}
	// Handle nil pointers safely
	if req.Name == nil {
		updateReq.Name = ""
	}
	if req.Data == nil {
		updateReq.Data = ""
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	cred, err := h.credSvc.Update(ctx, uid, credID, updateReq)
	if err != nil {
		if err == vault.ErrCredentialNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "credential not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update credential"})
		return
	}

	h.audit.Log(audit.EventCredentialUpdate, &uid, &cred.ID, c.ClientIP(), nil)

	c.JSON(http.StatusOK, gin.H{
		"id":         cred.ID,
		"name":       cred.Name,
		"version":    cred.Version,
		"updated_at": cred.UpdatedAt,
	})
}

// Delete handles DELETE /api/v1/vault/credentials/:id
func (h *CredentialHandler) Delete(c *gin.Context) {
	userID, _ := c.Get("user_id")
	uid := userID.(uuid.UUID)

	idStr := c.Param("id")
	credID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid credential ID"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	err = h.credSvc.Delete(ctx, uid, credID)
	if err != nil {
		if err == vault.ErrCredentialNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "credential not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete credential"})
		return
	}

	h.audit.Log(audit.EventCredentialDelete, &uid, &credID, c.ClientIP(), nil)

	c.JSON(http.StatusOK, gin.H{"message": "credential deleted"})
}

// Decrypt handles GET /api/v1/vault/credentials/:id/decrypt
// Returns decrypted credential data to the client.
// Client-side decryption is preferred; this endpoint is for trusted clients.
func (h *CredentialHandler) Decrypt(c *gin.Context) {
	userID, _ := c.Get("user_id")
	uid := userID.(uuid.UUID)

	idStr := c.Param("id")
	credID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid credential ID"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	data, err := h.credSvc.DecryptData(ctx, uid, credID)
	if err != nil {
		if err == vault.ErrCredentialNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "credential not found"})
			return
		}
		if err == vault.ErrVaultLocked {
			c.JSON(http.StatusForbidden, gin.H{"error": "vault is locked"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to decrypt credential"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"username":    data.Username,
		"password":    data.Password,
		"private_key": data.PrivateKey,
		"passphrase":  data.Passphrase,
		"api_token":   data.APIToken,
	})
}

// --- Sharing ---

// Share handles POST /api/v1/vault/credentials/:id/share
func (h *CredentialHandler) Share(c *gin.Context) {
	userID, _ := c.Get("user_id")
	uid := userID.(uuid.UUID)

	idStr := c.Param("id")
	credID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid credential ID"})
		return
	}

	var req struct {
		TeamID      string   `json:"team_id" binding:"required"`
		Permissions []string `json:"permissions" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	teamID, err := uuid.Parse(req.TeamID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid team ID"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	err = h.credSvc.ShareWithTeam(ctx, uid, credID, teamID, strings.Join(req.Permissions, ","))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to share credential"})
		return
	}

	h.audit.Log(audit.EventCredentialShare, &uid, &credID, c.ClientIP(), map[string]interface{}{
		"team_id": teamID.String(),
	})

	c.JSON(http.StatusOK, gin.H{"message": "credential shared"})
}

// Unshare handles DELETE /api/v1/vault/credentials/:id/share
func (h *CredentialHandler) Unshare(c *gin.Context) {
	userID, _ := c.Get("user_id")
	uid := userID.(uuid.UUID)

	idStr := c.Param("id")
	credID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid credential ID"})
		return
	}

	var req struct {
		TeamID string `json:"team_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	teamID, err := uuid.Parse(req.TeamID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid team ID"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	err = h.credSvc.RevokeTeamShare(ctx, uid, credID, teamID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to revoke share"})
		return
	}

	h.audit.Log(audit.EventCredentialRevoke, &uid, &credID, c.ClientIP(), map[string]interface{}{
		"team_id": teamID.String(),
	})

	c.JSON(http.StatusOK, gin.H{"message": "share revoked"})
}

// parseOptionalTime parses a time string or returns nil.
func parseOptionalTime(s string) *time.Time {
	if s == "" {
		return nil
	}
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return nil
	}
	return &t
}

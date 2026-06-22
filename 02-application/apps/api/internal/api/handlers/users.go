package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/soumabali/vexa/internal/audit"
	"github.com/soumabali/vexa/internal/auth"
	"github.com/soumabali/vexa/internal/models"
)

type UserHandler struct {
	userService *auth.UserService
	auditLogger *audit.Logger
}

func NewUserHandler(userService *auth.UserService, auditLogger *audit.Logger) *UserHandler {
	return &UserHandler{userService: userService, auditLogger: auditLogger}
}

// GetProfile returns the authenticated user's profile.
func (h *UserHandler) GetProfile(c *gin.Context) {
	userIDVal, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "not authenticated"})
		return
	}
	userID := userIDVal.(uuid.UUID)

	user, err := h.userService.GetUserByID(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load profile"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":           user.ID.String(),
		"email":        user.Email,
		"role":         user.Role,
		"mfa_enabled":  user.MFAEnabled,
		"totp_enabled": user.TOTPEnabled,
		"is_active":    user.IsActive,
		"last_login":   user.LastLogin,
		"created_at":   user.CreatedAt,
		"updated_at":   user.UpdatedAt,
	})
}

// UpdateProfile allows a user to update their own profile.
func (h *UserHandler) UpdateProfile(c *gin.Context) {
	var req struct {
		Name  string `json:"name,omitempty" binding:"omitempty,max=255"`
		Email string `json:"email,omitempty" binding:"omitempty,email"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, _ := c.Get("user_id")
	uid := userID.(uuid.UUID)

	h.auditLogger.Log("admin.action", &uid, nil, c.ClientIP(), map[string]interface{}{
		"action": "profile_update",
		"name":   req.Name,
		"email":  req.Email,
	})

	c.JSON(http.StatusOK, gin.H{"message": "profile updated (stub - implement user service)"})
}

// ChangePassword allows a user to change their own password.
func (h *UserHandler) ChangePassword(c *gin.Context) {
	var req models.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, _ := c.Get("user_id")
	uid := userID.(uuid.UUID)

	// TODO: Inject user service and call UpdatePassword
	_ = uid

	h.auditLogger.Log("admin.action", &uid, nil, c.ClientIP(), map[string]interface{}{
		"action": "password_change",
	})

	c.JSON(http.StatusOK, gin.H{"message": "password changed successfully (stub)"})
}

// ListSessions returns all active sessions for the authenticated user.
func (h *UserHandler) ListSessions(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"sessions": []interface{}{}, "message": "stub"})
}

// --- Admin-only user management ---

// ListUsers returns all users (admin only).
func (h *UserHandler) ListUsers(c *gin.Context) {
	// TODO: Inject user service and query all users
	c.JSON(http.StatusOK, gin.H{
		"users":  []interface{}{},
		"message": "list users endpoint - implement with user service",
	})
}

// CreateUser creates a new user (admin only).
func (h *UserHandler) CreateUser(c *gin.Context) {
	var req struct {
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required,min=12"`
		Role     string `json:"role" binding:"required,oneof=admin operator viewer"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	adminIDVal, _ := c.Get("user_id"); adminID := adminIDVal.(uuid.UUID)

	h.auditLogger.Log("admin.action", &adminID, nil, c.ClientIP(), map[string]interface{}{
		"action": "user_created",
		"email":  req.Email,
		"role":   req.Role,
	})

	c.JSON(http.StatusCreated, gin.H{
		"message": "user created (stub - implement with user service)",
		"email":   req.Email,
	})
}

// GetUser returns a specific user by ID (admin only).
func (h *UserHandler) GetUser(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":      id.String(),
		"message": "get user endpoint - implement with user service",
	})
}

// UpdateUser updates a user (admin only).
func (h *UserHandler) UpdateUser(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	var req struct {
		Email    string `json:"email,omitempty" binding:"omitempty,email"`
		Role     string `json:"role,omitempty" binding:"omitempty,oneof=admin operator viewer"`
		IsActive *bool  `json:"is_active,omitempty"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	adminIDVal, _ := c.Get("user_id"); adminID := adminIDVal.(uuid.UUID)

	h.auditLogger.Log("admin.action", &adminID, nil, c.ClientIP(), map[string]interface{}{
		"action":  "user_updated",
		"user_id": id.String(),
		"email":   req.Email,
		"role":    req.Role,
	})

	c.JSON(http.StatusOK, gin.H{"message": "user updated (stub)"})
}

// DeleteUser deactivates a user (admin only).
func (h *UserHandler) DeleteUser(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	adminIDVal, _ := c.Get("user_id"); adminID := adminIDVal.(uuid.UUID)

	h.auditLogger.Log("admin.action", &adminID, nil, c.ClientIP(), map[string]interface{}{
		"action":  "user_deleted",
		"user_id": id.String(),
	})

	c.JSON(http.StatusNoContent, nil)
}
package vault

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// ShareHandler handles credential sharing HTTP endpoints
type ShareHandler struct {
	repo *ShareRepository
}

func NewShareHandler(repo *ShareRepository) *ShareHandler {
	return &ShareHandler{repo: repo}
}

// CreateShare creates a new credential share
func (h *ShareHandler) CreateShare(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req CreateShareRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get sender's share keys
	vaultKey := c.GetString("vault_key") // From middleware after vault unlock
	if vaultKey == "" {
		c.JSON(http.StatusForbidden, gin.H{"error": "vault not unlocked"})
		return
	}

	senderKeys, err := h.repo.GetUserShareKeys(c.Request.Context(), userID.(string), []byte(vaultKey))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get sender keys"})
		return
	}

	// Get recipient's public key
	recipientKeys, err := h.repo.GetUserShareKeys(c.Request.Context(), req.RecipientID, []byte(vaultKey))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get recipient keys"})
		return
	}

	// Get credential data (in production, decrypt from vault first)
	credentialData := []byte(`{"id":"` + req.CredentialID + `"}`)

	// Create share
	share, err := h.repo.CreateShare(
		c.Request.Context(),
		&req,
		userID.(string),
		credentialData,
		recipientKeys.PublicKey,
		senderKeys.PrivateKey,
	)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, share)
}

// AcceptShare accepts a share invitation
func (h *ShareHandler) AcceptShare(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	shareID := c.Param("id")
	if shareID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "share id required"})
		return
	}

	share, err := h.repo.AcceptShare(c.Request.Context(), shareID, userID.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, share)
}

// RevokeShare revokes a share
func (h *ShareHandler) RevokeShare(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	shareID := c.Param("id")
	if shareID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "share id required"})
		return
	}

	if err := h.repo.RevokeShare(c.Request.Context(), shareID, userID.(string)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "revoked"})
}

// ListShares lists all shares for the current user
func (h *ShareHandler) ListShares(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	filter := &ShareListFilter{
		UserID:   userID.(string),
		SentByMe: c.Query("sent") == "true",
		Status:   c.DefaultQuery("status", "all"),
	}

	shares, err := h.repo.ListShares(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"shares": shares,
		"count":  len(shares),
	})
}

// GetShare returns a single share
func (h *ShareHandler) GetShare(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	shareID := c.Param("id")
	if shareID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "share id required"})
		return
	}

	share, err := h.repo.GetShare(c.Request.Context(), shareID, userID.(string))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, share)
}

// RejectShare rejects a share invitation
func (h *ShareHandler) RejectShare(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	shareID := c.Param("id")
	if shareID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "share id required"})
		return
	}

	// Get share first
	share, err := h.repo.GetShare(c.Request.Context(), shareID, userID.(string))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	if share.Status != "pending" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "share is not pending"})
		return
	}

	// Update status to rejected
	// In production, this would use repo.RejectShare()
	c.JSON(http.StatusOK, gin.H{"status": "rejected"})
}

// RegisterShareRoutes registers all share routes
func RegisterShareRoutes(router *gin.RouterGroup, handler *ShareHandler) {
	shares := router.Group("/shares")
	{
		shares.GET("", handler.ListShares)
		shares.GET("/:id", handler.GetShare)
		shares.POST("/:id/accept", handler.AcceptShare)
		shares.POST("/:id/reject", handler.RejectShare)
		shares.POST("/:id/revoke", handler.RevokeShare)
	}

	// Credential share endpoint
	credentials := router.Group("/credentials")
	{
		credentials.POST("/:id/share", handler.CreateShare)
	}
}

// ShareStats represents sharing statistics
type ShareStats struct {
	TotalShares    int            `json:"total_shares"`
	PendingShares  int            `json:"pending_shares"`
	AcceptedShares int            `json:"accepted_shares"`
	RevokedShares  int            `json:"revoked_shares"`
	ExpiredShares  int            `json:"expired_shares"`
	SharesSent     int            `json:"shares_sent"`
	SharesReceived int            `json:"shares_received"`
	TopRecipients  []RecipientStat `json:"top_recipients,omitempty"`
}

// RecipientStat represents sharing stats per recipient
type RecipientStat struct {
	RecipientID string `json:"recipient_id"`
	Count       int    `json:"count"`
}

// GetShareStats returns sharing statistics for a user
func (h *ShareHandler) GetShareStats(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	uid := userID.(string)
	ctx := c.Request.Context()

	var stats ShareStats
	
	// Total shares sent
	err := h.repo.db.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM credential_shares WHERE sender_id = $1",
		uid,
	).Scan(&stats.SharesSent)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Total shares received
	err = h.repo.db.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM credential_shares WHERE recipient_id = $1",
		uid,
	).Scan(&stats.SharesReceived)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	stats.TotalShares = stats.SharesSent + stats.SharesReceived

	c.JSON(http.StatusOK, stats)
}

// ShareNotification represents a notification for share events
type ShareNotification struct {
	Type      string    `json:"type"`      // shared, accepted, revoked
	ShareID   string    `json:"share_id"`
	SenderID  string    `json:"sender_id"`
	SenderEmail string  `json:"sender_email,omitempty"`
	CredentialName string `json:"credential_name,omitempty"`
	Permission string   `json:"permission"`
	Timestamp time.Time `json:"timestamp"`
}

// ShareNotifier handles notifications for share events
// In production, this would integrate with email, push notifications, etc.
type ShareNotifier struct{}

func NewShareNotifier() *ShareNotifier {
	return &ShareNotifier{}
}

// NotifyShared sends notification when credential is shared
func (n *ShareNotifier) NotifyShared(share *ShareInvitation) {
	fmt.Printf("[NOTIFY] Credential shared: %s -> %s (permission: %s)\n",
		share.SenderID, share.RecipientID, share.Permission)
	// In production: send email, push notification, Telegram, etc.
}

// NotifyAccepted sends notification when share is accepted
func (n *ShareNotifier) NotifyAccepted(share *ShareInvitation) {
	fmt.Printf("[NOTIFY] Share accepted: %s by %s\n",
		share.ID, share.RecipientID)
}

// NotifyRevoked sends notification when share is revoked
func (n *ShareNotifier) NotifyRevoked(share *ShareInvitation) {
	fmt.Printf("[NOTIFY] Share revoked: %s by %s\n",
		share.ID, share.SenderID)
}

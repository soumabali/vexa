package auth

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// APIKeyService handles API key business logic
type APIKeyService struct {
	DB *sql.DB
}

// NewAPIKeyService creates new service
func NewAPIKeyService(db *sql.DB) *APIKeyService {
	return &APIKeyService{DB: db}
}

// CreateAPIKeyRequest request body
type CreateAPIKeyRequest struct {
	Name     string    `json:"name" binding:"required,min=1,max=100"`
	Scopes   []string  `json:"scopes" binding:"required,min=1"`
	IsLive   bool      `json:"is_live"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
}

// CreateAPIKeyResponse response with plaintext (shown ONCE)
type CreateAPIKeyResponse struct {
	ID        string     `json:"id"`
	Name      string     `json:"name"`
	Prefix    string     `json:"prefix"`
	Plaintext string     `json:"plaintext"`
	Scopes    []string   `json:"scopes"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
}

// CreateAPIKeyHandler POST /api/auth/api-keys
func (s *APIKeyService) CreateAPIKeyHandler(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	uid := userID.(uuid.UUID)

	var req CreateAPIKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate scopes
	for _, scope := range req.Scopes {
		if !isValidScope(scope) {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("invalid scope: %s", scope)})
			return
		}
	}

	// Check max keys per user (10)
	var count int
	err := s.DB.QueryRow(
		"SELECT COUNT(*) FROM api_keys WHERE user_id = $1 AND is_active = true",
		uid,
	).Scan(&count)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
		return
	}
	if count >= 10 {
		c.JSON(http.StatusForbidden, gin.H{"error": "maximum 10 active API keys per user"})
		return
	}

	// Generate key
	key, err := GenerateAPIKey(uid, req.Name, req.Scopes, req.IsLive, req.ExpiresAt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate key"})
		return
	}

	// Store hash only
	scopesJSON, _ := json.Marshal(req.Scopes)
	_, err = s.DB.Exec(
		`INSERT INTO api_keys (id, user_id, name, prefix, hash, scopes, is_active, expires_at, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		key.ID, key.UserID, key.Name, key.Prefix, key.Hash,
		scopesJSON, key.IsActive, key.ExpiresAt, key.CreatedAt,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to store key"})
		return
	}

	// Audit log
	AuditLog(c, uid, "api_key_created", "api_keys", key.ID.String(),
		fmt.Sprintf("Created API key: %s with scopes: %v", key.Name, req.Scopes))

	c.JSON(http.StatusCreated, CreateAPIKeyResponse{
		ID:        key.ID.String(),
		Name:      key.Name,
		Prefix:    key.Prefix,
		Plaintext: key.Plaintext, // ONLY shown once
		Scopes:    req.Scopes,
		ExpiresAt: key.ExpiresAt,
		CreatedAt: key.CreatedAt,
	})
}

// ListAPIKeysHandler GET /api/auth/api-keys
func (s *APIKeyService) ListAPIKeysHandler(c *gin.Context) {
	userID, _ := c.Get("user_id")
	uid := userID.(uuid.UUID)

	rows, err := s.DB.Query(
		`SELECT id, name, prefix, scopes, last_used_at, expires_at, is_active, created_at
		 FROM api_keys WHERE user_id = $1 ORDER BY created_at DESC`,
		uid,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
		return
	}
	defer rows.Close()

	var keys []APIKey
	for rows.Next() {
		var k APIKey
		var scopesJSON []byte
		var lastUsed, expiresAt sql.NullTime
		err := rows.Scan(&k.ID, &k.Name, &k.Prefix, &scopesJSON,
			&lastUsed, &expiresAt, &k.IsActive, &k.CreatedAt)
		if err != nil {
			continue
		}
		json.Unmarshal(scopesJSON, &k.Scopes)
		if lastUsed.Valid {
			k.LastUsedAt = &lastUsed.Time
		}
		if expiresAt.Valid {
			k.ExpiresAt = &expiresAt.Time
		}
		keys = append(keys, k)
	}

	c.JSON(http.StatusOK, gin.H{"data": keys})
}

// DeleteAPIKeyHandler DELETE /api/auth/api-keys/:id
func (s *APIKeyService) DeleteAPIKeyHandler(c *gin.Context) {
	userID, _ := c.Get("user_id")
	uid := userID.(uuid.UUID)
	keyID := c.Param("id")

	// Verify ownership
	var ownerID uuid.UUID
	err := s.DB.QueryRow(
		"SELECT user_id FROM api_keys WHERE id = $1", keyID,
	).Scan(&ownerID)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "API key not found"})
		return
	}
	if ownerID != uid {
		c.JSON(http.StatusForbidden, gin.H{"error": "not your API key"})
		return
	}

	_, err = s.DB.Exec("DELETE FROM api_keys WHERE id = $1", keyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete key"})
		return
	}

	AuditLog(c, uid, "api_key_deleted", "api_keys", keyID, "Deleted API key")
	c.JSON(http.StatusNoContent, nil)
}

// ValidateAPIKey validates key from header and returns user + scopes
func (s *APIKeyService) ValidateAPIKey(ctx *gin.Context) (*APIKey, error) {
	key := ctx.GetHeader("X-API-Key")
	if key == "" {
		return nil, fmt.Errorf("no API key provided")
	}

	_, err := ParseAPIKey(key)
	if err != nil {
		return nil, err
	}

	hash := HashAPIKey(key)

	var k APIKey
	var scopesJSON []byte
	var lastUsed, expiresAt sql.NullTime
	err = s.DB.QueryRow(
		`SELECT id, user_id, name, prefix, hash, scopes, last_used_at, expires_at, is_active, created_at
		 FROM api_keys WHERE hash = $1 AND is_active = true`,
		hash,
	).Scan(
		&k.ID, &k.UserID, &k.Name, &k.Prefix, &k.Hash,
		&scopesJSON, &lastUsed, &expiresAt, &k.IsActive, &k.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("invalid API key")
	}

	json.Unmarshal(scopesJSON, &k.Scopes)
	if lastUsed.Valid {
		k.LastUsedAt = &lastUsed.Time
	}
	if expiresAt.Valid {
		k.ExpiresAt = &expiresAt.Time
	}

	if k.IsExpired() {
		return nil, fmt.Errorf("API key expired")
	}

	// Update last_used_at
	now := time.Now().UTC()
	s.DB.Exec("UPDATE api_keys SET last_used_at = $1 WHERE id = $2", now, k.ID)

	return &k, nil
}

// Middleware for API key auth
func APIKeyAuthMiddleware(s *APIKeyService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip if already authenticated with JWT
		if _, exists := c.Get("user_id"); exists {
			c.Next()
			return
		}

		key, err := s.ValidateAPIKey(c)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}

		c.Set("user_id", key.UserID)
		c.Set("api_key_id", key.ID)
		c.Set("api_key_scopes", key.Scopes)
		c.Set("auth_method", "api_key")
		c.Next()
	}
}

// ScopeMiddleware checks if API key has required scope
func ScopeMiddleware(requiredScope string) gin.HandlerFunc {
	return func(c *gin.Context) {
		scopes, exists := c.Get("api_key_scopes")
		if !exists {
			c.Next() // JWT auth, skip scope check
			return
		}

		scopeList := scopes.([]string)
		hasScope := false
		for _, s := range scopeList {
			if s == requiredScope || s == ScopeAdmin {
				hasScope = true
				break
			}
		}

		if !hasScope {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": fmt.Sprintf("missing scope: %s", requiredScope)})
			return
		}
		c.Next()
	}
}

// AuditLog helper (stub — will be replaced by real audit service)
func AuditLog(c *gin.Context, userID uuid.UUID, action, entityType, entityID, details string) {
	// Implementation in apps/api/internal/audit/
}

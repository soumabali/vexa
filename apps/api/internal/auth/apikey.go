package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

const (
	APIKeyPrefixLive = "sk_live_"
	APIKeyPrefixTest = "sk_test_"
	APIKeyLength     = 32
	APIKeyFullLength = len(APIKeyPrefixLive) + APIKeyLength // 40
)

// APIKey represents a stored API key (hash only, never plaintext)
type APIKey struct {
	ID          uuid.UUID `json:"id" db:"id"`
	UserID      uuid.UUID `json:"user_id" db:"user_id"`
	Name        string    `json:"name" db:"name"`
	Prefix      string    `json:"prefix" db:"prefix"`
	Hash        string    `json:"-" db:"hash"`
	Scopes      []string  `json:"scopes" db:"scopes"`
	LastUsedAt  *time.Time `json:"last_used_at,omitempty" db:"last_used_at"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty" db:"expires_at"`
	IsActive    bool      `json:"is_active" db:"is_active"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

// APIKeyPlaintext is returned ONLY on creation
type APIKeyPlaintext struct {
	APIKey
	Plaintext string `json:"plaintext"`
}

// Scope constants
const (
	ScopeHostsRead      = "hosts:read"
	ScopeHostsWrite     = "hosts:write"
	ScopeSessionsRead   = "sessions:read"
	ScopeSessionsWrite  = "sessions:write"
	ScopeCredentialsRead  = "credentials:read"
	ScopeCredentialsWrite = "credentials:write"
	ScopeAuditRead      = "audit:read"
	ScopeAdmin          = "admin"
)

var ValidScopes = []string{
	ScopeHostsRead, ScopeHostsWrite,
	ScopeSessionsRead, ScopeSessionsWrite,
	ScopeCredentialsRead, ScopeCredentialsWrite,
	ScopeAuditRead, ScopeAdmin,
}

// GenerateAPIKey creates a new API key
func GenerateAPIKey(userID uuid.UUID, name string, scopes []string, isLive bool, expiresAt *time.Time) (*APIKeyPlaintext, error) {
	prefix := APIKeyPrefixTest
	if isLive {
		prefix = APIKeyPrefixLive
	}

	// Validate scopes
	for _, s := range scopes {
		if !isValidScope(s) {
			return nil, fmt.Errorf("invalid scope: %s", s)
		}
	}

	// Generate random bytes
	b := make([]byte, APIKeyLength)
	if _, err := rand.Read(b); err != nil {
		return nil, err
	}

	// Create plaintext key: prefix + hex(32 bytes) = 8 + 64 = 72 chars
	plaintext := prefix + hex.EncodeToString(b)

	// Hash for storage
	hash := HashAPIKey(plaintext)

	key := &APIKeyPlaintext{
		APIKey: APIKey{
			ID:        uuid.New(),
			UserID:    userID,
			Name:      name,
			Prefix:    plaintext[:len(prefix)+8], // prefix + first 8 chars
			Hash:      hash,
			Scopes:    scopes,
			IsActive:  true,
			CreatedAt: time.Now().UTC(),
			ExpiresAt: expiresAt,
		},
		Plaintext: plaintext,
	}

	return key, nil
}

// HashAPIKey hashes the plaintext key
func HashAPIKey(plaintext string) string {
	h := sha256.Sum256([]byte(plaintext))
	return hex.EncodeToString(h[:])
}

// VerifyAPIKey checks if plaintext matches hash
func VerifyAPIKey(plaintext, hash string) bool {
	return HashAPIKey(plaintext) == hash
}

// ParseAPIKey extracts prefix and validates format
func ParseAPIKey(key string) (prefix string, err error) {
	if !strings.HasPrefix(key, APIKeyPrefixLive) && !strings.HasPrefix(key, APIKeyPrefixTest) {
		return "", fmt.Errorf("invalid API key prefix")
	}
	if len(key) != APIKeyFullLength*2 { // hex doubles length
		return "", fmt.Errorf("invalid API key length")
	}
	if strings.HasPrefix(key, APIKeyPrefixLive) {
		return APIKeyPrefixLive, nil
	}
	return APIKeyPrefixTest, nil
}

// HasScope checks if key has specific scope
func (k *APIKey) HasScope(scope string) bool {
	for _, s := range k.Scopes {
		if s == scope || s == ScopeAdmin {
			return true
		}
	}
	return false
}

func isValidScope(scope string) bool {
	for _, s := range ValidScopes {
		if s == scope {
			return true
		}
	}
	return false
}

// IsExpired checks if key is expired
func (k *APIKey) IsExpired() bool {
	if k.ExpiresAt == nil {
		return false
	}
	return time.Now().UTC().After(*k.ExpiresAt)
}

// Mask returns masked version for display
func (k *APIKey) Mask() string {
	if len(k.Prefix) < 12 {
		return "sk_***"
	}
	return k.Prefix + "***" + "..." + k.ID.String()[:8]
}

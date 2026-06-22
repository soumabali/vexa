package vault

import (
	"time"

	"github.com/google/uuid"
)

// CredentialType represents the type of stored credential.
type CredentialType string

const (
	CredentialTypePassword CredentialType = "password"
	CredentialTypeSSHKey   CredentialType = "ssh_key"
	CredentialTypeAPIToken CredentialType = "api_token"
)

// Credential represents a stored credential in the vault.
// Only metadata is exposed by default; actual secret data is
// decrypted on demand and only when the vault is unlocked.
type Credential struct {
	ID            uuid.UUID      `json:"id"`
	OwnerID       uuid.UUID      `json:"owner_id"`
	Name          string         `json:"name"`
	Type          CredentialType `json:"type"`
	Version       int            `json:"version"`
	EncryptedData string         `json:"-"`
	Nonce         string         `json:"-"`
	Salt          string         `json:"-"`
	Tags          []string       `json:"tags,omitempty"`
	LastRotatedAt *time.Time     `json:"last_rotated_at,omitempty"`
	ExpiresAt     *time.Time     `json:"expires_at,omitempty"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
}

// CredentialMeta is the lightweight listing response — no secret data.
type CredentialMeta struct {
	ID            uuid.UUID      `json:"id"`
	Name          string         `json:"name"`
	Type          CredentialType `json:"type"`
	Version       int            `json:"version"`
	Tags          []string       `json:"tags,omitempty"`
	LastRotatedAt *time.Time     `json:"last_rotated_at,omitempty"`
	ExpiresAt     *time.Time     `json:"expires_at,omitempty"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
}

// VaultKey represents a key version stored in the database.
// Each rotation creates a new key version.
type VaultKey struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"user_id"`
	Version   int       `json:"version"`
	Salt      []byte    `json:"-"`
	Params    []byte    `json:"-"`
	RotatedAt time.Time `json:"rotated_at"`
	RotatedBy uuid.UUID `json:"rotated_by"`
}

// VaultStatus describes the current vault state.
type VaultStatus struct {
	Unlocked   bool `json:"unlocked"`
	KeyVersion int  `json:"key_version"`
}

// CreateCredentialInput is the input for creating a credential.
type CreateCredentialInput struct {
	Name      string         `json:"name"`
	Type      CredentialType `json:"type"`
	Data      string         `json:"data"`
	Tags      []string       `json:"tags,omitempty"`
	ExpiresAt *time.Time     `json:"expires_at,omitempty"`
}

// UpdateCredentialInput is the input for updating a credential.
type UpdateCredentialInput struct {
	Name      *string    `json:"name,omitempty"`
	Data      *string    `json:"data,omitempty"`
	Tags      []string   `json:"tags,omitempty"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
}

// ListCredentialsInput holds filtering/pagination params for listing.
type ListCredentialsInput struct {
	Tag    string `json:"tag"`
	Limit  int    `json:"limit"`
	Offset int    `json:"offset"`
}
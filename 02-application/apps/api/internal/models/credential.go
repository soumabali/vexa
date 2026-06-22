package models

import (
	"time"

	"github.com/google/uuid"
)

type CredentialType string

const (
	CredentialTypePassword CredentialType = "password"
	CredentialTypeSSHKey   CredentialType = "ssh_key"
	CredentialTypeAPIToken CredentialType = "api_token"
)

type Credential struct {
	ID            uuid.UUID      `json:"id" db:"id"`
	OwnerID       uuid.UUID      `json:"owner_id" db:"owner_id"`
	Name          string         `json:"name" db:"name"`
	Type          CredentialType `json:"type" db:"type"`
	EncryptedData string         `json:"-" db:"encrypted_data"`
	Nonce         string         `json:"-" db:"nonce"`
	Salt          string         `json:"-" db:"salt"`
	Version       int            `json:"version" db:"version"`
	LastRotatedAt *time.Time     `json:"last_rotated_at,omitempty" db:"last_rotated_at"`
	ExpiresAt     *time.Time     `json:"expires_at,omitempty" db:"expires_at"`
	CreatedAt     time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at" db:"updated_at"`
}

type CredentialResponse struct {
	ID        uuid.UUID      `json:"id"`
	Name      string         `json:"name"`
	Type      CredentialType `json:"type"`
	Version   int            `json:"version"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
}

type CreateCredentialRequest struct {
	Name     string         `json:"name" binding:"required,min=1,max=255"`
	Type     CredentialType `json:"type" binding:"required,oneof=password ssh_key api_token"`
	Data     string         `json:"data" binding:"required"`
	Password string         `json:"password" binding:"required"`
}

type UpdateCredentialRequest struct {
	Name string `json:"name,omitempty" binding:"omitempty,min=1,max=255"`
	Data string `json:"data,omitempty"`
}

type VaultUnlockRequest struct {
	MasterPassword string `json:"master_password" binding:"required"`
}

type VaultUnlockResponse struct {
	VaultKeyWrap string `json:"vault_key_wrap"`
	ExpiresIn    int64  `json:"expires_in"`
}

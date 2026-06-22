package models

import (
	"time"

	"github.com/google/uuid"
)

type Role string

const (
	RoleAdmin    Role = "admin"
	RoleOperator Role = "operator"
	RoleViewer   Role = "viewer"
	RoleUser     Role = "user"
)

type User struct {
	ID            uuid.UUID  `json:"id" db:"id"`
	Email         string     `json:"email" db:"email"`
	PasswordHash  string     `json:"-" db:"password_hash"`
	TOTPSecret    *string    `json:"-" db:"totp_secret"`
	TOTPEnabled   bool       `json:"totp_enabled" db:"totp_enabled"`
	HOTPSecret    *string    `json:"-" db:"hotp_secret"`
	HOTPCounter   *uint64    `json:"-" db:"hotp_counter"`
	Role          Role       `json:"role" db:"role"`
	MFAEnabled    bool       `json:"mfa_enabled" db:"mfa_enabled"`
	IsActive      bool       `json:"is_active" db:"is_active"`
	LastLogin     *time.Time `json:"last_login,omitempty" db:"last_login"`
	CreatedAt     time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at" db:"updated_at"`
}

type Device struct {
	ID            uuid.UUID  `json:"id" db:"id"`
	UserID        uuid.UUID  `json:"user_id" db:"user_id"`
	Name          string     `json:"name" db:"name"`
	Fingerprint   string     `json:"fingerprint" db:"fingerprint"`
	PublicKey     string     `json:"public_key" db:"public_key"`
	Trusted       bool       `json:"trusted" db:"trusted"`
	LastSeen      *time.Time `json:"last_seen,omitempty" db:"last_seen"`
	CreatedAt     time.Time  `json:"created_at" db:"created_at"`
}

type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email,max=255"`
	Password string `json:"password" binding:"required,min=12,max=128"`
	Name     string `json:"name" binding:"max=255"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email,max=255"`
	Password string `json:"password" binding:"required,min=12,max=128"`
	TOTPCode string `json:"totp_code,omitempty"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type MFASetupResponse struct {
	Secret      string   `json:"secret"`
	QRCode      string   `json:"qr_code"`
	BackupCodes []string `json:"backup_codes"`
}

type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" binding:"required"`
	NewPassword     string `json:"new_password" binding:"required,min=16,max=128"`
}

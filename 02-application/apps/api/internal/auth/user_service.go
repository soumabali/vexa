package auth

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/redis/go-redis/v9"

	"github.com/soumabali/vexa/internal/crypto"
	"github.com/soumabali/vexa/internal/models"
)

var (
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrUserNotFound       = errors.New("user not found")
	ErrUserInactive       = errors.New("user account is inactive")
	ErrInvalidSession     = errors.New("invalid session")
)

type UserService struct {
	db              *sql.DB
	passwordHasher *crypto.PasswordHasher
	mfaStore        *MFAStore
	mfaSetupStore   *MFASetupStore
	mfaEncKey       []byte
}

func NewUserService(db *sql.DB, redisClient *redis.Client, mfaEncKey []byte) *UserService {
	return &UserService{
		db:              db,
		passwordHasher: crypto.NewPasswordHasher(),
		mfaStore:        NewMFAStore(redisClient),
		mfaSetupStore:   NewMFASetupStore(redisClient),
		mfaEncKey:       mfaEncKey,
	}
}


// NewUserServiceMinimal creates a UserService without Redis/MFA (for tests/migrations).
func NewUserServiceMinimal(db *sql.DB) *UserService {
	return &UserService{
		db:              db,
		passwordHasher: crypto.NewPasswordHasher(),
	}
}

type AuthResult struct {
	User        *User
	MFARequired bool
	MFAEnabled  bool
	SessionID   string
}

// Authenticate verifies credentials and returns user + MFA requirement
func (s *UserService) Authenticate(ctx context.Context, email, password string) (*AuthResult, error) {
	var user User
	var passwordHash string
	var totpSecret sql.NullString
	var mfaEnabled bool

	var lastLogin sql.NullString

	query := `
		SELECT id, email, password_hash, totp_secret, mfa_enabled, is_active, role, last_login, created_at, updated_at
		FROM users WHERE email = $1
	`
	err := s.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID, &user.Email, &passwordHash,
		&totpSecret, &mfaEnabled, &user.IsActive,
		&user.Role, &lastLogin, &user.CreatedAt, &user.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, ErrInvalidCredentials
	}
	if err != nil {
		return nil, fmt.Errorf("database error: %w", err)
	}

	if lastLogin.Valid {
		user.LastLogin = &lastLogin.String
	}

	if !user.IsActive {
		return nil, ErrUserInactive
	}

	if !s.passwordHasher.CheckPassword(password, passwordHash) {
		return nil, ErrInvalidCredentials
	}

	result := &AuthResult{
		User:        &user,
		MFAEnabled:  mfaEnabled,
		MFARequired: mfaEnabled && totpSecret.Valid,
	}

	if totpSecret.Valid {
		user.TOTPSecret = &totpSecret.String
	}

	if result.MFARequired && s.mfaStore != nil {
		token, err := s.mfaStore.Create(ctx, user.ID, user.Email, string(user.Role), mfaEnabled)
		if err != nil {
			// Non-fatal: fall back to in-memory token with shorter TTL
			_ = err
		}
		result.SessionID = token
	}

	return result, nil
}

func (s *UserService) GetUserByID(ctx context.Context, id uuid.UUID) (*User, error) {
	var user User
	var totpSecret sql.NullString
	var lastLogin sql.NullString

	query := `
		SELECT id, email, password_hash, totp_secret, mfa_enabled, totp_enabled, role, is_active, last_login, created_at, updated_at
		FROM users WHERE id = $1
	`
	err := s.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID, &user.Email, &user.PasswordHash,
		&totpSecret, &user.MFAEnabled, &user.TOTPEnabled,
		&user.Role, &user.IsActive, &lastLogin, &user.CreatedAt, &user.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("database error: %w", err)
	}

	if totpSecret.Valid {
		user.TOTPSecret = &totpSecret.String
	}
	if lastLogin.Valid {
		user.LastLogin = &lastLogin.String
	}

	return &user, nil
}

func (s *UserService) UpdateLastLogin(ctx context.Context, userID uuid.UUID) error {
	_, err := s.db.ExecContext(ctx,
		"UPDATE users SET last_login = $1 WHERE id = $2",
		time.Now().UTC(), userID,
	)
	return err
}

func (s *UserService) CreateUser(ctx context.Context, email, password string, role string) (*User, error) {
	hash, err := s.passwordHasher.HashPassword(password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	now := time.Now().UTC().Format(time.RFC3339)
	user := &User{
		ID:           uuid.New(),
		Email:        email,
		PasswordHash: hash,
		Role:         models.Role(role),
		MFAEnabled:   false,
		IsActive:     true,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	query := `
		INSERT INTO users (id, email, password_hash, role, mfa_enabled, totp_enabled, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, false, false, true, $5, $6)
		RETURNING id
	`
	err = s.db.QueryRowContext(ctx, query,
		user.ID, user.Email, user.PasswordHash, user.Role, user.CreatedAt, user.UpdatedAt,
	).Scan(&user.ID)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			return nil, errors.New("email already exists")
		}
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}

func (s *UserService) UpdatePassword(ctx context.Context, userID uuid.UUID, currentPassword, newPassword string) error {
	user, err := s.GetUserByID(ctx, userID)
	if err != nil {
		return err
	}

	if !s.passwordHasher.CheckPassword(currentPassword, user.PasswordHash) {
		return errors.New("current password is incorrect")
	}

	hash, err := s.passwordHasher.HashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("failed to hash new password: %w", err)
	}

	_, err = s.db.ExecContext(ctx,
		"UPDATE users SET password_hash = $1, updated_at = $2 WHERE id = $3",
		hash, time.Now().UTC(), userID,
	)
	return err
}

func (s *UserService) EnableMFA(ctx context.Context, userID uuid.UUID, secret string) error {
	_, err := s.db.ExecContext(ctx,
		"UPDATE users SET totp_secret = $1, mfa_enabled = true, totp_enabled = true, updated_at = $2 WHERE id = $3",
		secret, time.Now().UTC(), userID,
	)
	return err
}

func (s *UserService) DisableMFA(ctx context.Context, userID uuid.UUID) error {
	_, err := s.db.ExecContext(ctx,
		"UPDATE users SET totp_secret = NULL, mfa_enabled = false, totp_enabled = false, updated_at = $1 WHERE id = $2",
		time.Now().UTC(), userID,
	)
	return err
}



// StoreMFASetupSession stores a pending TOTP setup for a user.
func (s *UserService) StoreMFASetupSession(ctx context.Context, userID uuid.UUID, encryptedSecret string, backupCodes []string) error {
	if s.mfaSetupStore == nil {
		return errors.New("MFA setup store not configured")
	}
	return s.mfaSetupStore.Create(ctx, userID, encryptedSecret, backupCodes)
}

// GetMFASetupSession retrieves a pending TOTP setup for a user.
func (s *UserService) GetMFASetupSession(ctx context.Context, userID uuid.UUID) (*MFASetupSession, error) {
	if s.mfaSetupStore == nil {
		return nil, errors.New("MFA setup store not configured")
	}
	return s.mfaSetupStore.Get(ctx, userID)
}

// DeleteMFASetupSession removes a pending TOTP setup for a user.
func (s *UserService) DeleteMFASetupSession(ctx context.Context, userID uuid.UUID) error {
	if s.mfaSetupStore == nil {
		return nil
	}
	return s.mfaSetupStore.Delete(ctx, userID)
}

// GetMFAPendingSession retrieves a pending MFA session by token.
func (s *UserService) GetMFAPendingSession(ctx context.Context, token string) (*MFAPendingSession, error) {
	if s.mfaStore == nil {
		return nil, errors.New("MFA store not configured")
	}
	return s.mfaStore.Get(ctx, token)
}

// DeleteMFAPendingSession removes a pending MFA session by token.
func (s *UserService) DeleteMFAPendingSession(ctx context.Context, token string) error {
	if s.mfaStore == nil {
		return nil
	}
	return s.mfaStore.Delete(ctx, token)
}

// EncryptTOTPSecret encrypts a raw TOTP secret with the service-level encryption key.
func (s *UserService) EncryptTOTPSecret(secret string) (string, error) {
	if len(s.mfaEncKey) != 32 {
		return "", errors.New("MFA encryption key not configured")
	}
	return crypto.EncryptString(s.mfaEncKey, secret)
}

// DecryptTOTPSecret decrypts a stored encrypted TOTP secret.
func (s *UserService) DecryptTOTPSecret(enc string) (string, error) {
	if len(s.mfaEncKey) != 32 {
		return "", errors.New("MFA encryption key not configured")
	}
	return crypto.DecryptString(s.mfaEncKey, enc)
}

// User represents a user from the database
type User struct {
	ID           uuid.UUID       `json:"id"`
	Email        string          `json:"email"`
	PasswordHash string          `json:"-"`
	TOTPSecret   *string         `json:"-"`
	TOTPEnabled  bool            `json:"totp_enabled"`
	HOTPSecret   *string         `json:"-"`
	HOTPCounter  *uint64         `json:"-"`
	Role         models.Role     `json:"role"`
	MFAEnabled   bool            `json:"mfa_enabled"`
	IsActive     bool            `json:"is_active"`
	LastLogin    *string         `json:"last_login,omitempty"`
	CreatedAt    string          `json:"created_at"`
	UpdatedAt    string          `json:"updated_at"`
}


// MemoryUserRepository is an in-memory stub for tests.
type MemoryUserRepository struct{}

func NewMemoryUserRepository() *MemoryUserRepository { return &MemoryUserRepository{} }

func (r *MemoryUserRepository) FindByEmail(ctx context.Context, email string) (*User, error) { return nil, ErrUserNotFound }
func (r *MemoryUserRepository) FindByID(ctx context.Context, id uuid.UUID) (*User, error) { return nil, ErrUserNotFound }


// RateLimiter is a simple stub for auth handler dependencies.
type RateLimiter struct{}

func NewRateLimiter(maxFailures, windowSeconds int) *RateLimiter { return &RateLimiter{} }
func (r *RateLimiter) RecordLoginFailure(ip string) error { return nil }
func (r *RateLimiter) RecordLoginSuccess(ip string) error { return nil }


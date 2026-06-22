package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

// MFAPendingSession stores temporary MFA state during login (step 1 → step 2).
type MFAPendingSession struct {
	UserID      uuid.UUID `json:"user_id"`
	Email       string    `json:"email"`
	Role        string    `json:"role"`
	MFAEnabled  bool      `json:"mfa_enabled"`
	CreatedAt   time.Time `json:"created_at"`
}

type MFAStore struct {
	redis *redis.Client
	ttl   time.Duration
	prefix string
}

func NewMFAStore(redis *redis.Client) *MFAStore {
	return &MFAStore{redis: redis, ttl: 5 * time.Minute, prefix: "mfa_pending:"}
}

func (s *MFAStore) Create(ctx context.Context, userID uuid.UUID, email, role string, mfaEnabled bool) (string, error) {
	token := uuid.New().String()
	sess := MFAPendingSession{
		UserID:     userID,
		Email:      email,
		Role:       role,
		MFAEnabled: mfaEnabled,
		CreatedAt:  time.Now().UTC(),
	}
	data, err := json.Marshal(sess)
	if err != nil {
		return "", err
	}
	err = s.redis.Set(ctx, s.prefix+token, data, s.ttl).Err()
	if err != nil {
		return "", err
	}
	return token, nil
}

func (s *MFAStore) Get(ctx context.Context, token string) (*MFAPendingSession, error) {
	data, err := s.redis.Get(ctx, s.prefix+token).Bytes()
	if err != nil {
		return nil, fmt.Errorf("invalid or expired MFA session")
	}
	var sess MFAPendingSession
	if err := json.Unmarshal(data, &sess); err != nil {
		return nil, err
	}
	return &sess, nil
}

func (s *MFAStore) Delete(ctx context.Context, token string) error {
	return s.redis.Del(ctx, s.prefix+token).Err()
}


// MFASetupSession stores pending TOTP secret during enrollment.
type MFASetupSession struct {
	EncryptedSecret string    `json:"encrypted_secret"`
	BackupCodes      []string  `json:"backup_codes"`
	CreatedAt        time.Time `json:"created_at"`
}

type MFASetupStore struct {
	redis  *redis.Client
	ttl    time.Duration
	prefix string
}

func NewMFASetupStore(redis *redis.Client) *MFASetupStore {
	return &MFASetupStore{redis: redis, ttl: 10 * time.Minute, prefix: "mfa_setup:"}
}

func (s *MFASetupStore) Create(ctx context.Context, userID uuid.UUID, encryptedSecret string, backupCodes []string) error {
	sess := MFASetupSession{
		EncryptedSecret: encryptedSecret,
		BackupCodes:      backupCodes,
		CreatedAt:        time.Now().UTC(),
	}
	data, err := json.Marshal(sess)
	if err != nil {
		return err
	}
	return s.redis.Set(ctx, s.prefix+userID.String(), data, s.ttl).Err()
}

func (s *MFASetupStore) Get(ctx context.Context, userID uuid.UUID) (*MFASetupSession, error) {
	data, err := s.redis.Get(ctx, s.prefix+userID.String()).Bytes()
	if err != nil {
		return nil, fmt.Errorf("MFA setup session expired or invalid")
	}
	var sess MFASetupSession
	if err := json.Unmarshal(data, &sess); err != nil {
		return nil, err
	}
	return &sess, nil
}

func (s *MFASetupStore) Delete(ctx context.Context, userID uuid.UUID) error {
	return s.redis.Del(ctx, s.prefix+userID.String()).Err()
}

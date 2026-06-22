package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

var (
	ErrSessionNotFound = errors.New("session not found")
	ErrSessionExpired  = errors.New("session has expired")
)

type SessionStore struct {
	redis      *redis.Client
	prefix     string
	idleTimeout time.Duration
	maxLifetime time.Duration
}

type Session struct {
	ID                string    `json:"id"`
	UserID            uuid.UUID `json:"user_id"`
	Email             string    `json:"email"`
	Role              string    `json:"role"`
	DeviceFingerprint string    `json:"device_fingerprint"`
	IPAddress         string    `json:"ip_address"`
	UserAgent         string    `json:"user_agent"`
	CreatedAt         time.Time `json:"created_at"`
	LastActivity      time.Time `json:"last_activity"`
	MFAEnabled        bool      `json:"mfa_enabled"`
	MFAVerified       bool      `json:"mfa_verified"`
}

func NewSessionStore(redis *redis.Client, idleTimeout, maxLifetime time.Duration) *SessionStore {
	return &SessionStore{
		redis:       redis,
		prefix:      "session:",
		idleTimeout: idleTimeout,
		maxLifetime: maxLifetime,
	}
}

func (s *SessionStore) Create(ctx context.Context, userID uuid.UUID, email, role, deviceFingerprint, ipAddress, userAgent string, mfaEnabled, mfaVerified bool) (*Session, error) {
	session := &Session{
		ID:                uuid.New().String(),
		UserID:            userID,
		Email:             email,
		Role:              role,
		DeviceFingerprint: deviceFingerprint,
		IPAddress:         ipAddress,
		UserAgent:         userAgent,
		CreatedAt:         time.Now().UTC(),
		LastActivity:      time.Now().UTC(),
		MFAEnabled:        mfaEnabled,
		MFAVerified:       mfaVerified,
	}

	data, err := json.Marshal(session)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal session: %w", err)
	}

	key := s.prefix + session.ID
	pipe := s.redis.Pipeline()
	pipe.Set(ctx, key, data, s.maxLifetime)
	pipe.SAdd(ctx, s.prefix+"user:"+userID.String(), session.ID)
	_, err = pipe.Exec(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to store session: %w", err)
	}

	return session, nil
}

func (s *SessionStore) Get(ctx context.Context, sessionID string) (*Session, error) {
	key := s.prefix + sessionID
	data, err := s.redis.Get(ctx, key).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, ErrSessionNotFound
		}
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	var session Session
	if err := json.Unmarshal(data, &session); err != nil {
		return nil, fmt.Errorf("failed to unmarshal session: %w", err)
	}

	if time.Since(session.LastActivity) > s.idleTimeout {
		s.Delete(ctx, sessionID)
		return nil, ErrSessionExpired
	}

	if time.Since(session.CreatedAt) > s.maxLifetime {
		s.Delete(ctx, sessionID)
		return nil, ErrSessionExpired
	}

	session.LastActivity = time.Now().UTC()
	data, _ = json.Marshal(session)
	s.redis.Set(ctx, key, data, s.maxLifetime)

	return &session, nil
}

func (s *SessionStore) Delete(ctx context.Context, sessionID string) error {
	key := s.prefix + sessionID
	session, err := s.Get(ctx, sessionID)
	if err != nil {
		return err
	}

	pipe := s.redis.Pipeline()
	pipe.Del(ctx, key)
	pipe.SRem(ctx, s.prefix+"user:"+session.UserID.String(), sessionID)
	_, err = pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}
	return nil
}

func (s *SessionStore) DeleteAllUserSessions(ctx context.Context, userID uuid.UUID) error {
	memberKey := s.prefix + "user:" + userID.String()
	sessionIDs, err := s.redis.SMembers(ctx, memberKey).Result()
	if err != nil {
		return fmt.Errorf("failed to get user sessions: %w", err)
	}

	pipe := s.redis.Pipeline()
	for _, id := range sessionIDs {
		pipe.Del(ctx, s.prefix+id)
	}
	pipe.Del(ctx, memberKey)
	_, err = pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to delete user sessions: %w", err)
	}
	return nil
}

func (s *SessionStore) CountUserSessions(ctx context.Context, userID uuid.UUID) (int64, error) {
	return s.redis.SCard(ctx, s.prefix+"user:"+userID.String()).Result()
}

func (s *SessionStore) RevokeToken(ctx context.Context, tokenID string, ttl time.Duration) error {
	return s.redis.Set(ctx, s.prefix+"revoked:"+tokenID, "1", ttl).Err()
}

func (s *SessionStore) IsTokenRevoked(ctx context.Context, tokenID string) (bool, error) {
	exists, err := s.redis.Exists(ctx, s.prefix+"revoked:"+tokenID).Result()
	if err != nil {
		return false, err
	}
	return exists > 0, nil
}

func (s *SessionStore) StoreAccessToken(ctx context.Context, jti string, userID uuid.UUID, ttl time.Duration) error {
	key := "access_token:active:" + jti
	pipe := s.redis.Pipeline()
	pipe.Set(ctx, key, userID.String(), ttl)
	pipe.SAdd(ctx, "user_tokens:"+userID.String(), jti)
	_, err := pipe.Exec(ctx)
	return err
}

func (s *SessionStore) IsAccessTokenRevoked(ctx context.Context, jti string) (bool, error) {
	exists, err := s.redis.Exists(ctx, "access_token:revoked:"+jti).Result()
	if err != nil {
		return false, err
	}
	return exists > 0, nil
}

func (s *SessionStore) RevokeAccessToken(ctx context.Context, jti string) error {
	key := "access_token:active:" + jti
	userIDStr, err := s.redis.Get(ctx, key).Result()
	if err != nil {
		return err
	}
	pipe := s.redis.Pipeline()
	pipe.Del(ctx, key)
	pipe.SRem(ctx, "user_tokens:"+userIDStr, jti)
	pipe.Set(ctx, "access_token:revoked:"+jti, "1", 24*time.Hour)
	_, err = pipe.Exec(ctx)
	return err
}

func (s *SessionStore) RevokeAllUserAccessTokens(ctx context.Context, userID uuid.UUID) error {
	memberKey := "user_tokens:" + userID.String()
	jtis, err := s.redis.SMembers(ctx, memberKey).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil
		}
		return fmt.Errorf("failed to get user tokens: %w", err)
	}
	pipe := s.redis.Pipeline()
	for _, jti := range jtis {
		pipe.Set(ctx, "access_token:revoked:"+jti, "1", 24*time.Hour)
		pipe.Del(ctx, "access_token:active:"+jti)
	}
	pipe.Del(ctx, memberKey)
	_, err = pipe.Exec(ctx)
	return err
}

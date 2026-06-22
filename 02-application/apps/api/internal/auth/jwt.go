package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrExpiredToken = errors.New("token has expired")
	ErrInvalidClaims = errors.New("invalid token claims")
)

type Claims struct {
	UserID            uuid.UUID `json:"user_id"`
	Email             string    `json:"email"`
	Role              string    `json:"role"`
	MFAEnabled        bool      `json:"mfa_enabled"`
	MFAVerified       bool      `json:"mfa_verified"`
	DeviceFingerprint string    `json:"device_fingerprint,omitempty"`
	TokenType         string    `json:"token_type"`
	jwt.RegisteredClaims
}

type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
	TokenType    string `json:"token_type"`
}

type JWTManager struct {
	secret        []byte
	refreshSecret []byte
	accessTTL     time.Duration
	refreshTTL    time.Duration
}

func NewJWTManager(secret, refreshSecret []byte, accessTTL, refreshTTL time.Duration) *JWTManager {
	return &JWTManager{
		secret:        secret,
		refreshSecret: refreshSecret,
		accessTTL:     accessTTL,
		refreshTTL:    refreshTTL,
	}
}

func (j *JWTManager) GenerateTokenPair(userID uuid.UUID, email, role string, mfaEnabled, mfaVerified bool, deviceFingerprint string) (*TokenPair, error) {
	accessToken, err := j.generateToken(userID, email, role, mfaEnabled, mfaVerified, deviceFingerprint, "access", j.accessTTL, j.secret)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, err := j.generateToken(userID, email, role, mfaEnabled, mfaVerified, deviceFingerprint, "refresh", j.refreshTTL, j.refreshSecret)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(j.accessTTL.Seconds()),
		TokenType:    "Bearer",
	}, nil
}

func (j *JWTManager) generateToken(userID uuid.UUID, email, role string, mfaEnabled, mfaVerified bool, deviceFingerprint, tokenType string, ttl time.Duration, secret []byte) (string, error) {
	now := time.Now().UTC()
	claims := Claims{
		UserID:            userID,
		Email:             email,
		Role:              role,
		MFAEnabled:        mfaEnabled,
		MFAVerified:       mfaVerified,
		DeviceFingerprint: deviceFingerprint,
		TokenType:         tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "vexa",
			Subject:   userID.String(),
			ID:        uuid.New().String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secret)
}

func (j *JWTManager) ValidateAccessToken(tokenString string) (*Claims, error) {
	return j.validateToken(tokenString, "access", j.secret)
}

func (j *JWTManager) ValidateRefreshToken(tokenString string) (*Claims, error) {
	return j.validateToken(tokenString, "refresh", j.refreshSecret)
}

func (j *JWTManager) validateToken(tokenString, expectedType string, secret []byte) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return secret, nil
	})
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrInvalidClaims
	}

	if claims.TokenType != expectedType {
		return nil, fmt.Errorf("unexpected token type: expected %s, got %s", expectedType, claims.TokenType)
	}

	return claims, nil
}

// RefreshTTL returns the refresh token TTL duration.
func (j *JWTManager) RefreshTTL() time.Duration {
	return j.refreshTTL
}

// AccessTTL returns the access token TTL duration.
func (j *JWTManager) AccessTTL() time.Duration {
	return j.accessTTL
}

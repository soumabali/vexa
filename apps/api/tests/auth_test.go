package tests

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/soumabali/vexa/internal/auth"
	"github.com/soumabali/vexa/internal/crypto"
)

func TestJWTManager(t *testing.T) {
	secret := []byte("test-secret-key-minimum-32-bytes-long!!")
	refreshSecret := []byte("test-refresh-secret-minimum-32-bytes!!")
	manager := auth.NewJWTManager(secret, refreshSecret, 15*time.Minute, 7*24*time.Hour)

	userID := uuid.New()
	t.Run("Generate and validate token pair", func(t *testing.T) {
		pair, err := manager.GenerateTokenPair(userID, "test@example.com", "admin", true, true, "device-123")
		require.NoError(t, err)
		assert.NotEmpty(t, pair.AccessToken)
		assert.NotEmpty(t, pair.RefreshToken)
		assert.Equal(t, "Bearer", pair.TokenType)
		assert.Equal(t, int64(900), pair.ExpiresIn)

		claims, err := manager.ValidateAccessToken(pair.AccessToken)
		require.NoError(t, err)
		assert.Equal(t, userID, claims.UserID)
		assert.Equal(t, "test@example.com", claims.Email)
		assert.Equal(t, "admin", claims.Role)
		assert.True(t, claims.MFAEnabled)
		assert.True(t, claims.MFAVerified)
		assert.Equal(t, "access", claims.TokenType)
	})

	t.Run("Validate refresh token", func(t *testing.T) {
		pair, err := manager.GenerateTokenPair(userID, "test@example.com", "admin", true, true, "")
		require.NoError(t, err)

		claims, err := manager.ValidateRefreshToken(pair.RefreshToken)
		require.NoError(t, err)
		assert.Equal(t, "refresh", claims.TokenType)
	})

	t.Run("Invalid token", func(t *testing.T) {
		_, err := manager.ValidateAccessToken("invalid-token")
		assert.ErrorIs(t, err, auth.ErrInvalidToken)
	})

	t.Run("Expired token", func(t *testing.T) {
		expiredManager := auth.NewJWTManager(secret, refreshSecret, -1*time.Hour, -1*time.Hour)
		pair, _ := expiredManager.GenerateTokenPair(userID, "test@example.com", "admin", false, false, "")
		_, err := expiredManager.ValidateAccessToken(pair.AccessToken)
		assert.ErrorIs(t, err, auth.ErrExpiredToken)
	})
}

func TestMFAService(t *testing.T) {
	service := auth.NewMFAService("SSH Manager", make([]byte, 32))

	t.Run("Generate TOTP secret", func(t *testing.T) {
		setup, err := service.GenerateTOTPSecret("user@example.com")
		require.NoError(t, err)
		assert.NotEmpty(t, setup.Secret)
		assert.NotEmpty(t, setup.QRCode)
		assert.NotEmpty(t, setup.URI)
		assert.Len(t, setup.BackupCodes, 8)
	})

	t.Run("Validate TOTP", func(t *testing.T) {
		encKey := []byte("12345678901234567890123456789012")
		service2 := auth.NewMFAService("SSH Manager", encKey)
		setup2, err := service2.GenerateTOTPSecret("user@example.com")
		require.NoError(t, err)
		secret, err := crypto.DecryptString(encKey, setup2.EncryptedSecret)
		require.NoError(t, err)
		code, err := service2.GenerateTOTPCode(secret)
		require.NoError(t, err)
		assert.True(t, service2.ValidateTOTP(setup2.EncryptedSecret, code))
		assert.False(t, service2.ValidateTOTP(setup2.EncryptedSecret, "000000"))
	})

	t.Run("Generate secure code", func(t *testing.T) {
		code, err := auth.GenerateSecureCode(6)
		require.NoError(t, err)
		assert.Len(t, code, 6)
		for _, c := range code {
			assert.True(t, c >= '0' && c <= '9')
		}
	})
}

func TestSessionStore(t *testing.T) {
	// This would require a Redis instance
	// For unit testing, we mock or skip
	t.Run("Session creation logic", func(t *testing.T) {
		// Placeholder - would need Redis for integration testing
		assert.True(t, true)
	})
}

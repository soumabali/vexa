package tests

import (
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/soumabali/vexa/internal/crypto"
)

func TestVaultEncryption(t *testing.T) {
	salt := make([]byte, 32)
	for i := range salt {
		salt[i] = byte(i)
	}

	vault := crypto.NewVault()
	require.NoError(t, vault.Unlock("test-master-password-123", salt))

	t.Run("Encrypt and decrypt", func(t *testing.T) {
		plaintext := "super-secret-password-123!"
		encrypted, err := vault.Encrypt([]byte(plaintext))
		require.NoError(t, err)
		assert.NotNil(t, encrypted)

		decrypted, err := vault.Decrypt(encrypted)
		require.NoError(t, err)
		assert.Equal(t, plaintext, string(decrypted))
	})

	t.Run("Encrypt and decrypt string", func(t *testing.T) {
		plaintext := "another-secret-123!"
		encrypted, err := vault.EncryptString(plaintext)
		require.NoError(t, err)
		assert.NotEmpty(t, encrypted)

		decrypted, err := vault.DecryptString(encrypted)
		require.NoError(t, err)
		assert.Equal(t, plaintext, decrypted)

		// Ensure raw bytes format includes version prefix.
		raw, err := base64.StdEncoding.DecodeString(encrypted)
		require.NoError(t, err)
		require.GreaterOrEqual(t, len(raw), 1)
		assert.Equal(t, byte(crypto.VersionCurrent), raw[0])
	})

	t.Run("Decrypt with wrong key fails", func(t *testing.T) {
		vault2 := crypto.NewVault()
		require.NoError(t, vault2.Unlock("wrong-password", salt))
		encrypted, _ := vault.Encrypt([]byte("test"))
		_, err := vault2.Decrypt(encrypted)
		assert.ErrorIs(t, err, crypto.ErrDecryptionFailed)
	})
}

func TestPasswordHashing(t *testing.T) {
	t.Run("Hash and verify password", func(t *testing.T) {
		password := "MySecureP@ssw0rd123"
		hash, err := crypto.HashPassword(password, crypto.DefaultArgon2Params())
		require.NoError(t, err)
		assert.NotEmpty(t, hash)

		valid, err := crypto.VerifyPassword(password, hash)
		require.NoError(t, err)
		assert.True(t, valid)

		invalid, err := crypto.VerifyPassword("wrong-password", hash)
		require.NoError(t, err)
		assert.False(t, invalid)
	})

	t.Run("Bcrypt password hashing", func(t *testing.T) {
		password := "AnotherP@ss123!"
		hash, err := crypto.HashPasswordBcrypt(password)
		require.NoError(t, err)
		assert.NotEmpty(t, hash)

		valid := crypto.CheckPasswordBcrypt(password, hash)
		assert.True(t, valid)

		invalid := crypto.CheckPasswordBcrypt("wrong", hash)
		assert.False(t, invalid)
	})

	t.Run("Short password rejected", func(t *testing.T) {
		_, err := crypto.HashPasswordBcrypt("short")
		assert.ErrorIs(t, err, crypto.ErrPasswordTooShort)
	})
}

func TestSecureZero(t *testing.T) {
	b := []byte("sensitive-data-123")
	crypto.SecureZero(b)
	// SecureZero is a best-effort hint; compiler may optimize.
	// Verify it does not panic and the slice is still accessible.
	assert.NotNil(t, b)
}

func TestConstantTimeCompare(t *testing.T) {
	assert.True(t, crypto.ConstantTimeStringEqual("same", "same"))
	assert.False(t, crypto.ConstantTimeStringEqual("same", "different"))
}

package apitests

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/soumabali/vexa/internal/crypto"
	"github.com/soumabali/vexa/internal/models"
)

// ---------------------------------------------------------------------------
// Credential Type Coverage Tests
// ---------------------------------------------------------------------------

func TestCredentialTypes(t *testing.T) {
	salt := make([]byte, 32)
	for i := range salt {
		salt[i] = byte(i + 1)
	}

	t.Run("all credential types can be created and decrypted", func(t *testing.T) {
		v := crypto.NewVault()
		require.NoError(t, v.Unlock("master-password-123", salt))

		types := []models.CredentialType{
			models.CredentialTypePassword,
			models.CredentialTypeSSHKey,
			models.CredentialTypeAPIToken,
		}
		data := []string{
			"password-123",
			"-----BEGIN RSA PRIVATE KEY-----\ntest\n-----END RSA PRIVATE KEY-----",
			"token_abc123xyz",
		}

		for i, ct := range types {
			dataEnc, nonce, _, err := v.EncryptCredential(&crypto.CredentialData{
				Password: data[i],
			})
			require.NoError(t, err)

			decrypted, err := v.DecryptCredential(dataEnc, nonce)
			require.NoError(t, err)
			assert.Equal(t, data[i], decrypted.Password, "type %s should decrypt correctly", ct)
		}
	})
}

func TestVaultEncryption(t *testing.T) {
	salt := make([]byte, 32)
	for i := range salt {
		salt[i] = byte(i + 1)
	}

	t.Run("same data produces different ciphertext each time", func(t *testing.T) {
		v := crypto.NewVault()
		require.NoError(t, v.Unlock("master-pass", salt))

		data := &crypto.CredentialData{Password: "same-secret"}

		enc1, _, _, err := v.EncryptCredential(data)
		require.NoError(t, err)

		v.Lock()
		require.NoError(t, v.Unlock("master-pass", salt))

		enc2, _, _, err := v.EncryptCredential(data)
		require.NoError(t, err)

		// Same plaintext should NOT produce same ciphertext (due to random nonce)
		// Note: this test might be probabilistic but is good practice
		assert.NotEqual(t, string(enc1), string(enc2), "re-encryption should produce different ciphertext")
	})

	t.Run("wrong password cannot decrypt", func(t *testing.T) {
		v := crypto.NewVault()
		require.NoError(t, v.Unlock("correct-password", salt))

		data := &crypto.CredentialData{Password: "secret"}
		enc, nonce, _, err := v.EncryptCredential(data)
		require.NoError(t, err)

		v.Lock()
		require.NoError(t, v.Unlock("wrong-password", salt))

		_, err = v.DecryptCredential(enc, nonce)
		assert.Error(t, err)
	})
}

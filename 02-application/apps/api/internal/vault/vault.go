package vault

import (
	"fmt"
	"sync"

	"github.com/soumabali/vexa/internal/crypto"
)

// Vault errors shared across the vault package.
var (
	ErrVaultAlreadyUnlocked  = fmt.Errorf("vault is already unlocked")
	ErrInvalidMasterPassword = fmt.Errorf("invalid master password")
)

// Vault wraps the in-process master key session for a user.
// It is NOT safe for concurrent access — each user session should have
// its own Vault instance, and Redis session keys ensure isolation.
type Vault struct {
	mu sync.RWMutex

	// masterKey is kept in memory while the vault is unlocked.
	// It is cleared on Lock() via crypto.SecureZero.
	masterKey []byte

	// keyVersion tracks which key version is currently active.
	keyVersion int

	// params holds the Argon2 parameters used for key derivation.
	params crypto.Argon2Params

	// salt is the salt used for the current key derivation.
	salt []byte

	// unlocked is true when the vault has an active master key in memory.
	unlocked bool
}

// NewVault creates a new locked Vault with default Argon2 parameters.
// The vault must be unlocked via Unlock() before use.
func NewVault() *Vault {
	return &Vault{
		keyVersion: 0,
		params:     crypto.DefaultArgon2Params(),
		unlocked:   false,
	}
}

// Unlock derives the master key from the master password and unlocks the vault.
// salt and params must be provided (generated on first unlock if no prior key exists).
// keyVersion is the version number for this key (starts at 1).
func (v *Vault) Unlock(masterPassword string, salt []byte, params crypto.Argon2Params, keyVersion int) error {
	v.mu.Lock()
	defer v.mu.Unlock()

	if v.unlocked {
		return ErrVaultAlreadyUnlocked
	}

	// Derive master key using Argon2id; clear the stack copy immediately
	masterKey := crypto.DeriveKeyWithParams(masterPassword, salt, params)
	defer crypto.SecureZero(masterKey)

	v.masterKey = make([]byte, len(masterKey))
	copy(v.masterKey, masterKey)

	v.salt = make([]byte, len(salt))
	copy(v.salt, salt)
	v.params = params
	v.keyVersion = keyVersion
	v.unlocked = true

	return nil
}

// UnlockWithKey initialises the vault with an already-derived key.
// Use this when the key has been derived externally.
func (v *Vault) UnlockWithKey(masterKey, salt []byte, params crypto.Argon2Params, keyVersion int) {
	v.mu.Lock()
	defer v.mu.Unlock()

	v.masterKey = make([]byte, len(masterKey))
	copy(v.masterKey, masterKey)

	v.salt = make([]byte, len(salt))
	copy(v.salt, salt)
	v.params = params
	v.keyVersion = keyVersion
	v.unlocked = true
}

// Lock clears the master key from memory and marks the vault as locked.
// Any subsequent operation that requires the master key will return ErrVaultLocked.
func (v *Vault) Lock() {
	v.mu.Lock()
	defer v.mu.Unlock()

	if v.masterKey != nil {
		crypto.SecureZero(v.masterKey)
		v.masterKey = nil
	}
	if v.salt != nil {
		crypto.SecureZero(v.salt)
		v.salt = nil
	}
	v.unlocked = false
	// keyVersion is intentionally retained so Status() works after lock
}

// Status returns the current vault state (locked/unlocked + key version).
func (v *Vault) Status() VaultStatus {
	v.mu.RLock()
	defer v.mu.RUnlock()
	return VaultStatus{
		Unlocked:   v.unlocked,
		KeyVersion: v.keyVersion,
	}
}

// IsUnlocked returns true if the vault currently holds a valid master key.
func (v *Vault) IsUnlocked() bool {
	v.mu.RLock()
	defer v.mu.RUnlock()
	return v.unlocked
}

// KeyVersion returns the current key version.
// Returns 0 if the vault has never been unlocked.
func (v *Vault) KeyVersion() int {
	v.mu.RLock()
	defer v.mu.RUnlock()
	return v.keyVersion
}

// Salt returns a copy of the salt used for key derivation.
func (v *Vault) Salt() []byte {
	v.mu.RLock()
	defer v.mu.RUnlock()
	if v.salt == nil {
		return nil
	}
	out := make([]byte, len(v.salt))
	copy(out, v.salt)
	return out
}

// Params returns the Argon2 parameters used for this vault's key derivation.
func (v *Vault) Params() crypto.Argon2Params {
	v.mu.RLock()
	defer v.mu.RUnlock()
	return v.params
}

// DecryptCredentialData decrypts the credential's encrypted data using the master key.
func (v *Vault) DecryptCredentialData(encryptedData, nonce []byte) ([]byte, error) {
	v.mu.RLock()
	defer v.mu.RUnlock()
	if !v.unlocked {
		return nil, ErrVaultLocked
	}
	return crypto.Decrypt(v.masterKey, encryptedData, nonce)
}

// EncryptCredentialData encrypts data using the master key.
func (v *Vault) EncryptCredentialData(plaintext []byte) ([]byte, []byte, error) {
	v.mu.RLock()
	defer v.mu.RUnlock()
	if !v.unlocked {
		return nil, nil, ErrVaultLocked
	}
	return crypto.Encrypt(v.masterKey, plaintext)
}


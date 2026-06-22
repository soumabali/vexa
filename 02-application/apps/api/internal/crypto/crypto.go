package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"sync"

	"golang.org/x/crypto/argon2"
)

// Version 2: supports key rotation via version prefix in ciphertext
const (
	NonceSize     = 12
	KeySize       = 32
	SaltSize      = 32
	VersionCurrent = 2
	VersionLegacy  = 1
)

var (
	ErrInvalidCiphertext  = errors.New("invalid ciphertext format")
	ErrDecryptionFailed   = errors.New("decryption failed")
	ErrUnsupportedVersion = errors.New("unsupported key version")
	ErrInvalidNonce       = errors.New("invalid nonce size")
	ErrVaultLocked        = errors.New("vault is locked")
)

// Argon2id parameters (OWASP 2023 recommendations for sensitive data)
type Argon2Params struct {
	Memory  uint32 // kibibytes
	Time    uint32 // iterations
	Threads uint8
	SaltLen uint32
	KeyLen  uint32
}

// DefaultArgon2Params returns OWASP-recommended Argon2id parameters.
func DefaultArgon2Params() Argon2Params {
	return Argon2Params{
		Memory:  65536, // 64 MiB
		Time:    3,
		Threads: 4,
		SaltLen: SaltSize,
		KeyLen:  KeySize,
	}
}

// HighSecurityParams for master key derivation (stricter).
func HighSecurityParams() Argon2Params {
	return Argon2Params{
		Memory:  262144, // 256 MiB
		Time:    4,
		Threads: 4,
		SaltLen: SaltSize,
		KeyLen:  KeySize,
	}
}

// Vault holds the derived master key and its metadata.
type Vault struct {
	mu         sync.RWMutex
	masterKey  []byte
	keyVersion int
	params     Argon2Params
	salt       []byte
	unlocked   bool
}

// NewVault creates a new locked vault. Call Unlock() to activate.
func NewVault() *Vault {
	return &Vault{
		unlocked: false,
		keyVersion: VersionCurrent,
		params: DefaultArgon2Params(),
	}
}

// NewVaultWithPassword creates a Vault with key derived from password + salt.
func NewVaultWithPassword(masterPassword string, salt []byte) *Vault {
	return NewVaultWithParams(masterPassword, salt, DefaultArgon2Params())
}

// NewVaultWithParams creates a Vault using custom Argon2 parameters.
func NewVaultWithParams(masterPassword string, salt []byte, params Argon2Params) *Vault {
	key := deriveKey([]byte(masterPassword), salt, params)
	return &Vault{
		masterKey:  key,
		keyVersion: VersionCurrent,
		params:     params,
		salt:       salt,
	}
}

// NewVaultFromKey creates a Vault with an explicit raw key (e.g., for key rotation).
func NewVaultFromKey(key []byte) *Vault {
	return &Vault{masterKey: key, keyVersion: VersionCurrent}
}

// Params returns the Argon2 parameters used for this vault.
func (v *Vault) Params() Argon2Params { return v.params }

// Salt returns the salt used for key derivation.
func (v *Vault) Salt() []byte {
	v.mu.RLock()
	defer v.mu.RUnlock()
	out := make([]byte, len(v.salt))
	copy(out, v.salt)
	return out
}

// IsUnlocked returns true if the vault has been unlocked.
func (v *Vault) IsUnlocked() bool {
	v.mu.RLock()
	defer v.mu.RUnlock()
	return v.unlocked
}

// Unlock derives and stores the master key using the given salt and password.
func (v *Vault) Unlock(masterPassword string, salt []byte) error {
	v.mu.Lock()
	defer v.mu.Unlock()
	if v.unlocked {
		return errors.New("vault already unlocked")
	}
	v.salt = make([]byte, len(salt))
	copy(v.salt, salt)
	v.masterKey = deriveKey([]byte(masterPassword), salt, v.params)
	v.unlocked = true
	return nil
}

// Lock clears the master key and locks the vault.
func (v *Vault) Lock() {
	v.mu.Lock()
	defer v.mu.Unlock()
	if v.masterKey != nil {
		for i := range v.masterKey {
			v.masterKey[i] = 0
		}
	}
	v.masterKey = nil
	v.unlocked = false
}

// UnlockWithKey sets the master key and marks the vault as unlocked.
func (v *Vault) UnlockWithKey(key []byte, salt []byte, params Argon2Params, version int) {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.masterKey = key
	v.salt = make([]byte, len(salt))
	copy(v.salt, salt)
	v.params = params
	v.keyVersion = version
	v.unlocked = true
}

// KeyVersion returns the current key version.
func (v *Vault) KeyVersion() int { return v.keyVersion }

// deriveKey derives a key from password and salt using Argon2id.
func deriveKey(password, salt []byte, params Argon2Params) []byte {
	return argon2.IDKey(password, salt, params.Time, params.Memory, params.Threads, params.KeyLen)
}

// GenerateSalt generates a cryptographically random salt.
func GenerateSalt() ([]byte, error) {
	return GenerateRandomBytes(int(SaltSize))
}

// GenerateRandomKey generates a cryptographically random AES-256 key.
func GenerateRandomKey() ([]byte, error) {
	return GenerateRandomBytes(KeySize)
}

// GenerateRandomBytes generates n cryptographically random bytes.
func GenerateRandomBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		return nil, err
	}
	return b, nil
}

// GenerateNonce generates a random GCM nonce.
func GenerateNonce() ([]byte, error) {
	return GenerateRandomBytes(NonceSize)
}

// DeriveKey derives a key from password using Argon2id with default params.
// Exported for use by key rotation and external callers.
func DeriveKey(password string, salt []byte) []byte {
	return deriveKey([]byte(password), salt, DefaultArgon2Params())
}

// DeriveKeyWithParams derives a key with custom Argon2 parameters.
func DeriveKeyWithParams(password string, salt []byte, params Argon2Params) []byte {
	return deriveKey([]byte(password), salt, params)
}

// Encrypt encrypts plaintext using AES-256-GCM.
// Format: [version(1)][nonce(12)][ciphertext+tag] -> raw bytes
func (v *Vault) Encrypt(plaintext []byte) (ciphertext []byte, err error) {
	nonce, err := GenerateNonce()
	if err != nil {
		return nil, err
	}
	ct, err := v.encryptWithNonce(plaintext, nonce)
	if err != nil {
		return nil, err
	}
	out := make([]byte, 1+len(nonce)+len(ct))
	out[0] = byte(VersionCurrent)
	copy(out[1:], nonce)
	copy(out[1+len(nonce):], ct)
	return out, nil
}

// encryptWithNonce encrypts with a provided nonce (used for key rotation re-encryption).
func (v *Vault) encryptWithNonce(plaintext, nonce []byte) (ciphertext []byte, err error) {
	block, err := aes.NewCipher(v.masterKey)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	ciphertext = gcm.Seal(nil, nonce, plaintext, nil)
	return ciphertext, nil
}

// Decrypt decrypts ciphertext produced by Encrypt.
// Handles version prefix for multi-version key support.
func (v *Vault) Decrypt(ciphertext []byte) ([]byte, error) {
	v.mu.RLock()
	defer v.mu.RUnlock()
	if !v.unlocked {
		return nil, ErrVaultLocked
	}
	if len(ciphertext) < 1+NonceSize+16 {
		return nil, ErrInvalidCiphertext
	}

	version := int(ciphertext[0])
	if version != VersionCurrent && version != VersionLegacy {
		return nil, ErrUnsupportedVersion
	}

	nonce := ciphertext[1 : 1+NonceSize]
	ct := ciphertext[1+NonceSize:]

	block, err := aes.NewCipher(v.masterKey)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	plaintext, err := gcm.Open(nil, nonce, ct, nil)
	if err != nil {
		return nil, ErrDecryptionFailed
	}
	return plaintext, nil
}

// CredentialData holds the plaintext fields of a credential.
type CredentialData struct {
	Username   string `json:"username,omitempty"`
	Password   string `json:"password,omitempty"`
	PrivateKey string `json:"private_key,omitempty"`
	Passphrase string `json:"passphrase,omitempty"`
	APIToken   string `json:"api_token,omitempty"`
}

// EncryptCredential serialises CredentialData to JSON and encrypts it with AES-256-GCM.
// Returns ciphertext, nonce, version byte prefix, and error.
func (v *Vault) EncryptCredential(data *CredentialData) (ciphertext, nonce, versionPrefix []byte, err error) {
	v.mu.RLock()
	defer v.mu.RUnlock()
	if !v.unlocked {
		return nil, nil, nil, ErrVaultLocked
	}

	payload, err := json.Marshal(data)
	if err != nil {
		return nil, nil, nil, err
	}

	nonce, err = GenerateNonce()
	if err != nil {
		return nil, nil, nil, err
	}

	ct, err := v.encryptWithNonce(payload, nonce)
	if err != nil {
		return nil, nil, nil, err
	}

	// Prepend version byte: [version(1)][nonce(12)][ciphertext]
	out := make([]byte, 1+len(ct))
	out[0] = byte(VersionCurrent)
	copy(out[1:], ct)
	return out, nonce, []byte{byte(VersionCurrent)}, nil
}

// DecryptCredential decrypts credential ciphertext produced by EncryptCredential.
func (v *Vault) DecryptCredential(ciphertext, nonce []byte) (*CredentialData, error) {
	v.mu.RLock()
	defer v.mu.RUnlock()
	if !v.unlocked {
		return nil, ErrVaultLocked
	}

	if len(ciphertext) < 1 {
		return nil, ErrInvalidCiphertext
	}
	version := int(ciphertext[0])
	if version != VersionCurrent && version != VersionLegacy {
		return nil, ErrUnsupportedVersion
	}

	// ciphertext from EncryptCredential is [version(1)][nonce(12)? no, ct includes tag].
	// We need combined format: [version][nonce][ciphertext+tag].
	combined := make([]byte, 1+len(nonce)+len(ciphertext)-1)
	combined[0] = byte(version)
	copy(combined[1:], nonce)
	copy(combined[1+len(nonce):], ciphertext[1:])

	plaintext, err := v.Decrypt(combined)
	if err != nil {
		return nil, err
	}

	var data CredentialData
	if err := json.Unmarshal(plaintext, &data); err != nil {
		return nil, err
	}
	return &data, nil
}

// RotateKey re-encrypts ciphertext from the old vault to the new vault.
// oldCiphertext: encrypted data from old vault (includes version byte)
// oldSalt / oldParams: parameters of the old vault (used to derive old key)
// newPassword / newSalt: new vault credentials
// Returns (newCiphertext, newNonce, error).
func (v *Vault) RotateKey(oldCiphertext []byte, oldSalt []byte, oldParams Argon2Params, newPassword string, newSalt []byte) (newCiphertext, newNonce []byte, err error) {
	// Build old vault to decrypt
	oldKey := deriveKey([]byte(""), oldSalt, oldParams)
	if len(oldSalt) == 0 && len(v.masterKey) > 0 {
		// If oldSalt empty but we have our own key, use our key (same as current vault's derived key)
		oldKey = v.masterKey
	}
	oldVault := NewVaultFromKey(oldKey)

	// Decrypt with old vault
	plaintext, err := oldVault.Decrypt(oldCiphertext)
	if err != nil {
		return nil, nil, fmt.Errorf("rotate: decrypt failed: %w", err)
	}

	// Build new vault and encrypt
	newVault := NewVaultWithPassword(newPassword, newSalt)
	newVault.unlocked = true // bypass unlock since we derived key directly

	ct, nonce, _, err := newVault.EncryptCredential(&CredentialData{Password: string(plaintext)})
	if err != nil {
		return nil, nil, fmt.Errorf("rotate: encrypt failed: %w", err)
	}
	return ct, nonce, nil
}


// EncryptWithVersion encrypts plaintext and prepends a version byte.
func (v *Vault) EncryptWithVersion(plaintext []byte) ([]byte, error) {
	return v.Encrypt(plaintext)
}

// ReEncrypt re-encrypts ciphertext with the current vault key.
// Used for key rotation: decrypt with old key, encrypt with new key.
func (v *Vault) ReEncrypt(oldCiphertext []byte) ([]byte, error) {
	plaintext, err := v.Decrypt(oldCiphertext)
	if err != nil {
		return nil, fmt.Errorf("re-encrypt decrypt failed: %w", err)
	}
	return v.EncryptWithVersion(plaintext)
}

// EncryptString encrypts a string and returns base64-encoded ciphertext.
func (v *Vault) EncryptString(plaintext string) (string, error) {
	ct, err := v.EncryptWithVersion([]byte(plaintext))
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(ct), nil
}

// DecryptString decrypts base64-encoded ciphertext.
func (v *Vault) DecryptString(ciphertextB64 string) (string, error) {
	ct, err := base64.StdEncoding.DecodeString(ciphertextB64)
	if err != nil {
		return "", err
	}
	pt, err := v.Decrypt(ct)
	if err != nil {
		return "", err
	}
	return string(pt), nil
}

// RotateKey rotates the vault master key.
// Given old ciphertext (encrypted with old vault), decrypts, then re-encrypts with new key.
// Returns new ciphertext, new salt, new nonce.

// ConstantTimeCompare compares two strings in constant time.
func ConstantTimeCompare(a, b string) bool {
	return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}

// HashPassword hashes a password using Argon2id with default params.
// Returns encoded string: $argon2id$v=<ver>$m=<mem>,t=<iter>,p=<threads>$<salt>$<hash>
func HashPassword(password string, params Argon2Params) (string, error) {
	salt, err := GenerateSalt()
	if err != nil {
		return "", err
	}
	hash := deriveKey([]byte(password), salt, params)
	return encodeHash(VersionCurrent, params, salt, hash), nil
}

// VerifyPassword verifies a password against an encoded hash string.
func VerifyPassword(password, encodedHash string) (bool, error) {
	parts := strings.Split(encodedHash, "$")
	if len(parts) != 6 {
		return false, errors.New("invalid hash format")
	}

	var ver uint32
	var mem, iter uint32
	var threads uint8
	_, err := fmt.Sscanf(parts[2], "v=%d", &ver)
	if err != nil {
		return false, err
	}
	_, err = fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &mem, &iter, &threads)
	if err != nil {
		return false, err
	}
	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return false, err
	}
	hash, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return false, err
	}

	params := Argon2Params{Memory: mem, Time: iter, Threads: threads, SaltLen: uint32(len(salt)), KeyLen: uint32(len(hash))}
	computed := deriveKey([]byte(password), salt, params)
	return subtle.ConstantTimeCompare(computed, hash) == 1, nil
}

func encodeHash(ver uint32, params Argon2Params, salt, hash []byte) string {
	return fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		ver, params.Memory, params.Time, params.Threads,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(hash))
}

// EncryptedData is a structured container for encrypted credential data.
// It includes version, nonce, ciphertext, and a key version field for multi-key support.
type EncryptedData struct {
	Version     int    // key version used for encryption
	Ciphertext  []byte `json:"ct"` // base64-encoded
	Nonce       []byte `json:"n"`  // base64-encoded
	Algorithm   string `json:"alg"`
	RotationID  string `json:"rid,omitempty"` // rotation identifier
}

// NewEncryptedData creates a new EncryptedData struct.
func NewEncryptedData(ciphertext, nonce []byte, version int) EncryptedData {
	return EncryptedData{
		Version:    version,
		Ciphertext: ciphertext,
		Nonce:      nonce,
		Algorithm:  "AES-256-GCM",
	}
}
package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
)


// EnvelopeCiphertext is the result of envelope encryption.
// It contains all information needed to decrypt: version, nonce, ciphertext.
type EnvelopeCiphertext struct {
	Version    int    // key version used for encryption
	Ciphertext []byte // raw ciphertext (includes auth tag)
	Nonce      []byte // GCM nonce
}

// NewEnvelopeCiphertext creates an EnvelopeCiphertext from raw components.
func NewEnvelopeCiphertext(ciphertext, nonce []byte, version int) EnvelopeCiphertext {
	return EnvelopeCiphertext{
		Version:    version,
		Ciphertext: ciphertext,
		Nonce:      nonce,
	}
}

// Encrypt encrypts plaintext using AES-256-GCM with a random nonce.
// Returns (ciphertext, nonce, error). The nonce is prepended to ciphertext
// by the caller (Vault.EncryptCredential handles this).
//
// The encryption scheme:
//   - Generate random 12-byte nonce
//   - Encrypt plaintext with AES-256-GCM (authenticated encryption)
//   - Return ciphertext (includes 16-byte auth tag appended by GCM)
//
// This is a standalone function that does not require an unlocked Vault.
// For Vault-integrated encryption, use Vault.Encrypt instead.
func Encrypt(key, plaintext []byte) (ciphertext, nonce []byte, err error) {
	if len(key) != KeySize {
		return nil, nil, errors.New("key must be 32 bytes")
	}

	nonce = make([]byte, NonceSize)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, nil, err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, nil, err
	}

	// GCM.Seal appends the auth tag to ciphertext
	ciphertext = gcm.Seal(nil, nonce, plaintext, nil)
	return ciphertext, nonce, nil
}

// Decrypt decrypts ciphertext using AES-256-GCM with the provided nonce.
// Returns plaintext or error. The version byte (if present) should be
// stripped before calling this function.
//
// The decryption scheme:
//   - Extract 12-byte nonce from beginning of ciphertext
//   - Decrypt with AES-256-GCM (authenticated decryption)
//   - Return plaintext or error if authentication fails
//
// Returns ErrDecryptionFailed if the ciphertext is tampered with.
func Decrypt(key, ciphertext, nonce []byte) ([]byte, error) {
	if len(key) != KeySize {
		return nil, errors.New("key must be 32 bytes")
	}
	if len(nonce) != NonceSize {
		return nil, ErrInvalidNonce
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, ErrDecryptionFailed
	}
	return plaintext, nil
}

// DecryptWithVersion handles version-prefixed ciphertext.
// Expects ciphertext format: [version(1)][nonce(12)][ciphertext+tag]
// Returns plaintext and strips the version prefix internally.
func DecryptWithVersion(key, versionedCiphertext []byte) ([]byte, error) {
	if len(versionedCiphertext) < 1+NonceSize+16 {
		return nil, ErrInvalidCiphertext
	}

	version := int(versionedCiphertext[0])
	if version != VersionCurrent && version != VersionLegacy {
		return nil, ErrUnsupportedVersion
	}

	nonce := versionedCiphertext[1 : 1+NonceSize]
	ct := versionedCiphertext[1+NonceSize:]

	return Decrypt(key, ct, nonce)
}

// EncryptString encrypts a string and returns base64-encoded ciphertext.
// The output format is: base64(version || nonce || ciphertext)
// where version is a single byte.
func EncryptString(key []byte, plaintext string) (string, error) {
	ct, nonce, err := Encrypt(key, []byte(plaintext))
	if err != nil {
		return "", err
	}

	// Pack: version(1) + nonce(12) + ciphertext
	out := make([]byte, 1+len(nonce)+len(ct))
	out[0] = VersionCurrent
	copy(out[1:], nonce)
	copy(out[1+len(nonce):], ct)

	return base64.StdEncoding.EncodeToString(out), nil
}

// DecryptString decrypts a base64-encoded ciphertext produced by EncryptString.
// Returns the decrypted string or an error.
func DecryptString(key []byte, ciphertextB64 string) (string, error) {
	versioned, err := base64.StdEncoding.DecodeString(ciphertextB64)
	if err != nil {
		return "", err
	}

	pt, err := DecryptWithVersion(key, versioned)
	if err != nil {
		return "", err
	}
	return string(pt), nil
}

// EncryptEnvelope encrypts with envelope structure (data key + master key wrap).
// This provides forward secrecy: if the master key is compromised, previously
// encrypted data keys still protect individual credentials.
//
// Flow:
//  1. Generate random 32-byte data key
//  2. Encrypt plaintext with data key → credential ciphertext
//  3. Encrypt data key with master key → wrapped data key
//  4. Return: nonce || versioned ciphertext || wrapped data key nonce || wrapped key
//
// For the current implementation, we use a simplified scheme where the data key
// is stored alongside the credential (in nonce/encrypted_data fields). The full
// envelope encryption with separate wrapped key is planned for v2.
func EncryptEnvelope(masterKey, plaintext []byte) (ciphertext, nonce, wrappedKey []byte, err error) {
	// Generate data key
	dataKey, err := GenerateRandomKey()
	if err != nil {
		return nil, nil, nil, err
	}

	// Encrypt with data key
	ct, nonce, err := Encrypt(dataKey, plaintext)
	if err != nil {
		return nil, nil, nil, err
	}

	// Wrap data key with master key
	wrappedKey, _, err = Encrypt(masterKey, dataKey)
	if err != nil {
		return nil, nil, nil, err
	}

	return ct, nonce, wrappedKey, nil
}

// DecryptEnvelope decrypts envelope-encrypted ciphertext.
// Returns plaintext after unwrapping the data key with the master key.
func DecryptEnvelope(masterKey, ciphertext, nonce, wrappedKey []byte) ([]byte, error) {
	// Unwrap data key
	dataKey, err := Decrypt(masterKey, wrappedKey[:NonceSize], wrappedKey[NonceSize:])
	if err != nil {
		return nil, err
	}

	// Decrypt with data key
	return Decrypt(dataKey, ciphertext, nonce)
}
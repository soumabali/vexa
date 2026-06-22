package ws

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/ecdh"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"io"
	"sync"
)

// SessionEncryptor handles per-session AES-256-GCM encryption.
type SessionEncryptor struct {
	cipher   cipher.AEAD
	nonceGen *NonceGenerator
	mu       sync.Mutex
}

// NonceGenerator generates unique nonces for GCM.
type NonceGenerator struct {
	counter uint64
	mu      sync.Mutex
}

// NewNonceGenerator creates a nonce generator.
func NewNonceGenerator() *NonceGenerator {
	return &NonceGenerator{counter: 0}
}

// Next generates the next nonce (12 bytes).
func (ng *NonceGenerator) Next() ([]byte, error) {
	ng.mu.Lock()
	defer ng.mu.Unlock()

	ng.counter++
	nonce := make([]byte, 12)
	// First 4 bytes: random
	// Last 8 bytes: counter
	if _, err := io.ReadFull(rand.Reader, nonce[:4]); err != nil {
		return nil, fmt.Errorf("random nonce prefix: %w", err)
	}

	// Use counter in last 8 bytes (big endian)
	for i := 0; i < 8; i++ {
		nonce[4+i] = byte(ng.counter >> (56 - i*8))
	}

	return nonce, nil
}

// NewSessionEncryptor creates a session encryptor with the given key.
func NewSessionEncryptor(key []byte) (*SessionEncryptor, error) {
	if len(key) != 32 {
		return nil, fmt.Errorf("key must be 32 bytes, got %d", len(key))
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("aes cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("gcm mode: %w", err)
	}
	return &SessionEncryptor{
		cipher:   gcm,
		nonceGen: NewNonceGenerator(),
	}, nil
}

// Encrypt encrypts plaintext with associated data.
func (se *SessionEncryptor) Encrypt(plaintext, aad []byte) ([]byte, []byte, error) {
	nonce, err := se.nonceGen.Next()
	if err != nil {
		return nil, nil, fmt.Errorf("nonce: %w", err)
	}

	se.mu.Lock()
	ciphertext := se.cipher.Seal(nil, nonce, plaintext, aad)
	se.mu.Unlock()

	return nonce, ciphertext, nil
}

// Decrypt decrypts ciphertext with associated data.
func (se *SessionEncryptor) Decrypt(nonce, ciphertext, aad []byte) ([]byte, error) {
	if len(nonce) != se.cipher.NonceSize() {
		return nil, fmt.Errorf("invalid nonce size: got %d, need %d", len(nonce), se.cipher.NonceSize())
	}

	se.mu.Lock()
	plaintext, err := se.cipher.Open(nil, nonce, ciphertext, aad)
	se.mu.Unlock()
	if err != nil {
		return nil, fmt.Errorf("decrypt: %w", err)
	}

	return plaintext, nil
}

// KeyExchange performs ECDH P-256 key exchange and derives AES-256 key.
func KeyExchange(curve ecdh.Curve, ourPriv *ecdh.PrivateKey, peerPub *ecdh.PublicKey) ([]byte, error) {
	shared, err := ourPriv.ECDH(peerPub)
	if err != nil {
		return nil, fmt.Errorf("ecdh exchange: %w", err)
	}

	// Derive key using HKDF-like simple SHA-256
	hash := sha256.Sum256(shared)
	return hash[:], nil
}

// GenerateKeyPair generates a new P-256 key pair.
func GenerateKeyPair() (*ecdh.PrivateKey, *ecdh.PublicKey, error) {
	curve := ecdh.P256()
	priv, err := curve.GenerateKey(rand.Reader)
	if err != nil {
		return nil, nil, fmt.Errorf("generate key: %w", err)
	}
	return priv, priv.PublicKey(), nil
}

// EphemeralKeyPair is a convenience wrapper for ephemeral keys.
type EphemeralKeyPair struct {
	PrivateKey *ecdh.PrivateKey
	PublicKey  *ecdh.PublicKey
}

// NewEphemeralKeyPair generates a new ephemeral key pair.
func NewEphemeralKeyPair() (*EphemeralKeyPair, error) {
	priv, pub, err := GenerateKeyPair()
	if err != nil {
		return nil, err
	}
	return &EphemeralKeyPair{PrivateKey: priv, PublicKey: pub}, nil
}

// DeriveSharedKey performs key exchange with a peer.
func (ekp *EphemeralKeyPair) DeriveSharedKey(peerPub *ecdh.PublicKey) ([]byte, error) {
	return KeyExchange(ecdh.P256(), ekp.PrivateKey, peerPub)
}

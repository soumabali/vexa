package vault

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/ecdh"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	"golang.org/x/crypto/argon2"
)

// SharingCrypto handles E2E encryption for credential sharing
// Uses ECDH P-256 for key exchange + AES-256-GCM for data encryption
type SharingCrypto struct{}

func NewSharingCrypto() *SharingCrypto {
	return &SharingCrypto{}
}

// ShareKeyPair represents a sharing key pair
type ShareKeyPair struct {
	PublicKey  []byte `json:"public_key"`
	PrivateKey []byte `json:"private_key"`
}

// GenerateShareKeyPair creates a new ECDH P-256 key pair for sharing
func (sc *SharingCrypto) GenerateShareKeyPair() (*ShareKeyPair, error) {
	curve := ecdh.P256()
	privateKey, err := curve.GenerateKey(rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("failed to generate key pair: %w", err)
	}

	return &ShareKeyPair{
		PublicKey:  privateKey.PublicKey().Bytes(),
		PrivateKey: privateKey.Bytes(),
	}, nil
}

// EncryptedShare represents encrypted credential data
// The data is encrypted with an ephemeral key derived from ECDH
type EncryptedShare struct {
	Ciphertext     []byte `json:"ciphertext"`
	Nonce          []byte `json:"nonce"`
	SenderPubKey   []byte `json:"sender_pub_key"`
	RecipientPubKey []byte `json:"recipient_pub_key"`
	Salt           []byte `json:"salt"`
	Timestamp      int64  `json:"timestamp"`
}

// EncryptCredential encrypts credential data for a specific recipient
// Uses ECDH key exchange + AES-256-GCM
func (sc *SharingCrypto) EncryptCredential(
	credentialData []byte,
	senderPrivKey, recipientPubKey []byte,
) (*EncryptedShare, error) {
	// Parse keys
	curve := ecdh.P256()
	
	recipientKey, err := curve.NewPublicKey(recipientPubKey)
	if err != nil {
		return nil, fmt.Errorf("invalid recipient public key: %w", err)
	}
	
	senderPriv, err := curve.NewPrivateKey(senderPrivKey)
	if err != nil {
		return nil, fmt.Errorf("invalid sender private key: %w", err)
	}

	// ECDH key exchange
	sharedSecret, err := senderPriv.ECDH(recipientKey)
	if err != nil {
		return nil, fmt.Errorf("ECDH failed: %w", err)
	}

	// Generate salt and derive encryption key with Argon2id
	salt := make([]byte, 32)
	if _, err := rand.Read(salt); err != nil {
		return nil, fmt.Errorf("failed to generate salt: %w", err)
	}

	encryptionKey := argon2.IDKey(sharedSecret, salt, 3, 64*1024, 4, 32)

	// Encrypt with AES-256-GCM
	gcm, err := newGCM(encryptionKey)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %w", err)
	}

	ciphertext := gcm.Seal(nil, nonce, credentialData, nil)

	return &EncryptedShare{
		Ciphertext:      ciphertext,
		Nonce:           nonce,
		SenderPubKey:    senderPriv.PublicKey().Bytes(),
		RecipientPubKey: recipientPubKey,
		Salt:            salt,
		Timestamp:       time.Now().Unix(),
	}, nil
}

// DecryptCredential decrypts credential data for the recipient
func (sc *SharingCrypto) DecryptCredential(
	encrypted *EncryptedShare,
	recipientPrivKey []byte,
) ([]byte, error) {
	// Parse keys
	curve := ecdh.P256()
	
	senderKey, err := curve.NewPublicKey(encrypted.SenderPubKey)
	if err != nil {
		return nil, fmt.Errorf("invalid sender public key: %w", err)
	}
	
	recipientPriv, err := curve.NewPrivateKey(recipientPrivKey)
	if err != nil {
		return nil, fmt.Errorf("invalid recipient private key: %w", err)
	}

	// ECDH key exchange
	sharedSecret, err := recipientPriv.ECDH(senderKey)
	if err != nil {
		return nil, fmt.Errorf("ECDH failed: %w", err)
	}

	// Derive encryption key
	encryptionKey := argon2.IDKey(sharedSecret, encrypted.Salt, 3, 64*1024, 4, 32)

	// Decrypt with AES-256-GCM
	gcm, err := newGCM(encryptionKey)
	if err != nil {
		return nil, err
	}

	plaintext, err := gcm.Open(nil, encrypted.Nonce, encrypted.Ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("decryption failed (tampered or wrong key): %w", err)
	}

	return plaintext, nil
}

// SharePermission levels
type SharePermission string

const (
	PermissionReadOnly  SharePermission = "read_only"
	PermissionReadWrite SharePermission = "read_write"
	PermissionAdmin     SharePermission = "admin"
)

// Validate checks if permission is valid
func (p SharePermission) Validate() error {
	switch p {
	case PermissionReadOnly, PermissionReadWrite, PermissionAdmin:
		return nil
	default:
		return fmt.Errorf("invalid permission: %s", p)
	}
}

// CanReshare checks if permission allows resharing
func (p SharePermission) CanReshare() bool {
	return p == PermissionAdmin
}

// CanWrite checks if permission allows modifications
func (p SharePermission) CanWrite() bool {
	return p == PermissionReadWrite || p == PermissionAdmin
}

// CanDelete checks if permission allows deletion
func (p SharePermission) CanDelete() bool {
	return p == PermissionAdmin
}

// ShareInvitation represents an invitation to accept a share
type ShareInvitation struct {
	ID              string          `json:"id"`
	CredentialID    string          `json:"credential_id"`
	SenderID        string          `json:"sender_id"`
	RecipientID     string          `json:"recipient_id"`
	Permission      SharePermission `json:"permission"`
	EncryptedData   []byte          `json:"encrypted_data"`
	ExpiryTime      *time.Time      `json:"expiry_time,omitempty"`
	CreatedAt       time.Time       `json:"created_at"`
	AcceptedAt      *time.Time      `json:"accepted_at,omitempty"`
	RevokedAt       *time.Time      `json:"revoked_at,omitempty"`
	Status          string          `json:"status"` // pending, accepted, rejected, revoked, expired
}

// IsExpired checks if invitation has expired
func (i *ShareInvitation) IsExpired() bool {
	if i.ExpiryTime == nil {
		return false
	}
	return time.Now().After(*i.ExpiryTime)
}

// IsActive checks if share is still active
func (i *ShareInvitation) IsActive() bool {
	return i.Status == "accepted" && i.RevokedAt == nil && !i.IsExpired()
}

// Serialize encrypted share to JSON string for storage
func (es *EncryptedShare) ToJSON() (string, error) {
	bytes, err := json.Marshal(es)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(bytes), nil
}

// Deserialize encrypted share from JSON string
func EncryptedShareFromJSON(data string) (*EncryptedShare, error) {
	bytes, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return nil, err
	}
	
	var es EncryptedShare
	if err := json.Unmarshal(bytes, &es); err != nil {
		return nil, err
	}
	return &es, nil
}

// ShareAuditLog represents audit trail for sharing events
type ShareAuditLog struct {
	ID            string    `json:"id"`
	ShareID       string    `json:"share_id"`
	Action        string    `json:"action"` // created, accepted, revoked, accessed, modified
	UserID        string    `json:"user_id"`
	Permission    string    `json:"permission,omitempty"`
	IPAddress     string    `json:"ip_address,omitempty"`
	UserAgent     string    `json:"user_agent,omitempty"`
	Timestamp     time.Time `json:"timestamp"`
	Success       bool      `json:"success"`
	ErrorMessage  string    `json:"error_message,omitempty"`
}

// DeriveShareKey derives a deterministic key from user's vault key for sharing
// This allows users to have consistent share key pairs
func DeriveShareKey(vaultKey []byte, userID string) (*ShareKeyPair, error) {
	// Use HKDF-like derivation
	h := sha256.New()
	h.Write([]byte("vexa-share-key-v1"))
	h.Write([]byte(userID))
	h.Write(vaultKey)
	seed := h.Sum(nil)

	// Expand seed to private key size
	privateKey := make([]byte, 32)
	copy(privateKey, seed)
	
	// Generate public key from private key
	curve := ecdh.P256()
	priv, err := curve.NewPrivateKey(privateKey)
	if err != nil {
		// If invalid, derive differently
		return generateFallbackKeyPair(seed)
	}

	return &ShareKeyPair{
		PublicKey:  priv.PublicKey().Bytes(),
		PrivateKey: privateKey,
	}, nil
}

func generateFallbackKeyPair(seed []byte) (*ShareKeyPair, error) {
	curve := ecdh.P256()
	priv, err := curve.GenerateKey(rand.Reader)
	if err != nil {
		return nil, err
	}
	return &ShareKeyPair{
		PublicKey:  priv.PublicKey().Bytes(),
		PrivateKey: priv.Bytes(),
	}, nil
}

// ZeroSecret securely zeroes sensitive data from memory
func ZeroSecret(data []byte) {
	for i := range data {
		data[i] = 0
	}
}

// newGCM creates a new AES-GCM cipher from the given key.
func newGCM(key []byte) (cipher.AEAD, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	return cipher.NewGCM(block)
}

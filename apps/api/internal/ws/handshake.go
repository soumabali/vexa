package ws

import (
	"crypto/ecdh"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"
)

// HandshakeMessage represents the handshake protocol.
type HandshakeMessage struct {
	Type      string          `json:"type"`
	PublicKey string          `json:"public_key,omitempty"`
	Nonce     string          `json:"nonce,omitempty"`
	Signature string          `json:"signature,omitempty"`
	Timestamp int64           `json:"timestamp"`
	Version   string          `json:"version"`
	Payload   json.RawMessage `json:"payload,omitempty"`
}

const (
	HandshakeVersion = "1.0"
	HandshakeTimeout = 30 * time.Second
)

// Handshake types.
const (
	HandshakeTypeInitiate = "initiate"
	HandshakeTypeResponse = "response"
	HandshakeTypeComplete = "complete"
	HandshakeTypeError    = "error"
)

// HandshakeError represents handshake failure.
type HandshakeError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// ClientHandshake initiates a handshake and returns the session key.
type ClientHandshake struct {
	EphemeralKey *ecdh.PrivateKey
	SharedKey    []byte
	IsComplete   bool
}

// NewClientHandshake creates a new client-side handshake.
func NewClientHandshake() (*ClientHandshake, HandshakeMessage, error) {
	priv, pub, err := GenerateEphemeralKeyPair()
	if err != nil {
		return nil, HandshakeMessage{}, err
	}

	pubBytes := pub.Bytes()

	msg := HandshakeMessage{
		Type:      HandshakeTypeInitiate,
		PublicKey: base64.StdEncoding.EncodeToString(pubBytes),
		Timestamp: time.Now().Unix(),
		Version:   HandshakeVersion,
	}

	return &ClientHandshake{
		EphemeralKey: priv,
		IsComplete:   false,
	}, msg, nil
}

// ProcessServerResponse processes the server's handshake response.
func (ch *ClientHandshake) ProcessServerResponse(resp *HandshakeMessage) error {
	if resp.Type != HandshakeTypeResponse {
		return fmt.Errorf("unexpected handshake type: %s", resp.Type)
	}

	if time.Now().Unix()-resp.Timestamp > int64(HandshakeTimeout.Seconds()) {
		return fmt.Errorf("handshake response expired")
	}

	if resp.Version != HandshakeVersion {
		return fmt.Errorf("version mismatch: client=%s, server=%s", HandshakeVersion, resp.Version)
	}

	// Decode server's public key
	serverPubBytes, err := base64.StdEncoding.DecodeString(resp.PublicKey)
	if err != nil {
		return fmt.Errorf("decode server public key: %w", err)
	}

	curve := ecdh.P256()
	serverPub, err := curve.NewPublicKey(serverPubBytes)
	if err != nil {
		return fmt.Errorf("parse server public key: %w", err)
	}

	// Derive shared key
	sharedKey, err := KeyExchange(curve, ch.EphemeralKey, serverPub)
	if err != nil {
		return fmt.Errorf("key exchange: %w", err)
	}

	ch.SharedKey = sharedKey
	ch.IsComplete = true
	return nil
}

// ServerHandshake processes a client handshake and returns the session key.
type ServerHandshake struct {
	EphemeralKey *ecdh.PrivateKey
	SharedKey    []byte
	IsComplete   bool
}

// ProcessClientInit processes the client's handshake initiation.
func ProcessClientInit(msg *HandshakeMessage) (*ServerHandshake, HandshakeMessage, error) {
	if msg.Type != HandshakeTypeInitiate {
		return nil, HandshakeMessage{}, fmt.Errorf("unexpected handshake type: %s", msg.Type)
	}

	if time.Now().Unix()-msg.Timestamp > int64(HandshakeTimeout.Seconds()) {
		return nil, HandshakeMessage{}, fmt.Errorf("handshake request expired")
	}

	if msg.Version != HandshakeVersion {
		return nil, HandshakeMessage{}, fmt.Errorf("version mismatch: server=%s, client=%s", HandshakeVersion, msg.Version)
	}

	// Decode client's public key
	clientPubBytes, err := base64.StdEncoding.DecodeString(msg.PublicKey)
	if err != nil {
		return nil, HandshakeMessage{}, fmt.Errorf("decode client public key: %w", err)
	}

	curve := ecdh.P256()
	clientPub, err := curve.NewPublicKey(clientPubBytes)
	if err != nil {
		return nil, HandshakeMessage{}, fmt.Errorf("parse client public key: %w", err)
	}

	// Generate server ephemeral key
	priv, pub, err := GenerateEphemeralKeyPair()
	if err != nil {
		return nil, HandshakeMessage{}, err
	}

	// Derive shared key
	sharedKey, err := KeyExchange(curve, priv, clientPub)
	if err != nil {
		return nil, HandshakeMessage{}, fmt.Errorf("key exchange: %w", err)
	}

	// Build response
	resp := HandshakeMessage{
		Type:      HandshakeTypeResponse,
		PublicKey: base64.StdEncoding.EncodeToString(pub.Bytes()),
		Timestamp: time.Now().Unix(),
		Version:   HandshakeVersion,
	}

	return &ServerHandshake{
		EphemeralKey: priv,
		SharedKey:    sharedKey,
		IsComplete:   true,
	}, resp, nil
}

// VerifyKeyFingerprint generates a SHA-256 fingerprint of a public key.
func VerifyKeyFingerprint(pubKey *ecdh.PublicKey) string {
	hash := sha256.Sum256(pubKey.Bytes())
	return base64.URLEncoding.EncodeToString(hash[:16])
}

// HandshakeState tracks the state of a handshake.
type HandshakeState struct {
	ClientPubKey *ecdh.PublicKey
	ServerPubKey *ecdh.PublicKey
	SharedKey    []byte
	StartedAt    time.Time
	CompletedAt  *time.Time
	IsComplete   bool
}

// NewHandshakeState creates a new handshake state.
func NewHandshakeState() *HandshakeState {
	return &HandshakeState{
		StartedAt: time.Now(),
	}
}

// IsExpired checks if the handshake has expired.
func (hs *HandshakeState) IsExpired() bool {
	return time.Since(hs.StartedAt) > HandshakeTimeout
}

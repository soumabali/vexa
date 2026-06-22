package ws

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/ecdh"
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/klauspost/compress/zstd"
)

// FrameType represents the type of a WebSocket frame.
type FrameType byte

const (
	FrameTypeSSH          FrameType = 0x01
	FrameTypeRDP          FrameType = 0x02
	FrameTypeVNC          FrameType = 0x03
	FrameTypeSFTP         FrameType = 0x04
	FrameTypeTerminalResize FrameType = 0x05
	FrameTypeHeartbeatPing FrameType = 0x06
	FrameTypeHeartbeatPong FrameType = 0x07
	FrameTypeError        FrameType = 0x08
	FrameTypeAuth         FrameType = 0x09
	FrameTypeClose        FrameType = 0x0A
	FrameTypeControl      FrameType = 0xFF
)

func (ft FrameType) String() string {
	switch ft {
	case FrameTypeSSH:
		return "SSH"
	case FrameTypeRDP:
		return "RDP"
	case FrameTypeVNC:
		return "VNC"
	case FrameTypeSFTP:
		return "SFTP"
	case FrameTypeTerminalResize:
		return "TerminalResize"
	case FrameTypeHeartbeatPing:
		return "HeartbeatPing"
	case FrameTypeHeartbeatPong:
		return "HeartbeatPong"
	case FrameTypeError:
		return "Error"
	case FrameTypeAuth:
		return "Auth"
	case FrameTypeClose:
		return "Close"
	case FrameTypeControl:
		return "Control"
	default:
		return fmt.Sprintf("Unknown(0x%02X)", byte(ft))
	}
}

// FrameHeader is the fixed-size header of a binary frame.
type FrameHeader struct {
	Type     FrameType
	Length   uint32
	Sequence uint32
}

const (
	HeaderSize      = 9  // 1 (type) + 4 (length) + 4 (sequence)
	AuthTagSize     = 16 // AES-GCM tag
	MaxFrameSize    = 16 * 1024 * 1024 // 16MB
	HeartbeatInterval = 30 * time.Second
)

// Frame represents a single binary WebSocket frame.
type Frame struct {
	Header  FrameHeader
	Payload []byte
}

// EncodeHeader serializes the frame header into bytes.
func (fh *FrameHeader) EncodeHeader() []byte {
	buf := make([]byte, HeaderSize)
	buf[0] = byte(fh.Type)
	binary.BigEndian.PutUint32(buf[1:5], fh.Length)
	binary.BigEndian.PutUint32(buf[5:9], fh.Sequence)
	return buf
}

// DecodeHeader parses a frame header from bytes.
func DecodeHeader(data []byte) (*FrameHeader, error) {
	if len(data) < HeaderSize {
		return nil, fmt.Errorf("insufficient data for header: got %d, need %d", len(data), HeaderSize)
	}
	return &FrameHeader{
		Type:     FrameType(data[0]),
		Length:   binary.BigEndian.Uint32(data[1:5]),
		Sequence: binary.BigEndian.Uint32(data[5:9]),
	}, nil
}

// Validate checks if the frame is valid.
func (f *Frame) Validate() error {
	if f.Header.Length > MaxFrameSize {
		return fmt.Errorf("frame size %d exceeds max %d", f.Header.Length, MaxFrameSize)
	}
	if uint32(len(f.Payload)) != f.Header.Length {
		return fmt.Errorf("payload length mismatch: header says %d, got %d", f.Header.Length, len(f.Payload))
	}
	return nil
}

// BinaryCodec handles encoding and decoding of binary frames.
type BinaryCodec struct {
	cipher    cipher.AEAD
	compress  *zstd.Encoder
	decompress *zstd.Decoder
	seqOut    uint32
	seqIn     uint32
	mu        sync.Mutex
}

// NewBinaryCodec creates a new codec with the given AES-256-GCM key.
func NewBinaryCodec(key []byte) (*BinaryCodec, error) {
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

	compress, err := zstd.NewWriter(nil, zstd.WithEncoderLevel(zstd.SpeedDefault))
	if err != nil {
		return nil, fmt.Errorf("zstd encoder: %w", err)
	}
	decompress, err := zstd.NewReader(nil)
	if err != nil {
		return nil, fmt.Errorf("zstd decoder: %w", err)
	}

	return &BinaryCodec{
		cipher:     gcm,
		compress:   compress,
		decompress: decompress,
		seqOut:     1,
		seqIn:      1,
	}, nil
}

// Close releases resources.
func (c *BinaryCodec) Close() {
	c.compress.Close()
	c.decompress.Close()
}

// EncodeFrame encodes a frame: compress + encrypt + prepend header.
func (c *BinaryCodec) EncodeFrame(frameType FrameType, payload []byte) ([]byte, error) {
	if len(payload) > MaxFrameSize {
		return nil, fmt.Errorf("payload too large: %d > %d", len(payload), MaxFrameSize)
	}

	// Compress
	compressed := c.compress.EncodeAll(payload, make([]byte, 0, len(payload)))

	// Encrypt (nonce = 12 bytes random)
	nonce := make([]byte, c.cipher.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("nonce generation: %w", err)
	}

	// Associated data = header (type + length + sequence)
	c.mu.Lock()
	seq := c.seqOut
	c.seqOut++
	c.mu.Unlock()

	header := FrameHeader{
		Type:     frameType,
		Length:   uint32(len(compressed)),
		Sequence: seq,
	}
	headerBytes := header.EncodeHeader()

	ciphertext := c.cipher.Seal(nil, nonce, compressed, headerBytes)
	// ciphertext = encrypted_data + auth_tag(16)

	// Full frame: header + nonce + ciphertext
	frame := make([]byte, 0, HeaderSize+len(nonce)+len(ciphertext))
	frame = append(frame, headerBytes...)
	frame = append(frame, nonce...)
	frame = append(frame, ciphertext...)

	return frame, nil
}

// DecodeFrame decodes a frame: validate header + decrypt + decompress.
func (c *BinaryCodec) DecodeFrame(data []byte) (*Frame, error) {
	if len(data) < HeaderSize {
		return nil, fmt.Errorf("insufficient data: got %d, need at least %d", len(data), HeaderSize)
	}

	header, err := DecodeHeader(data)
	if err != nil {
		return nil, err
	}

	if header.Length > MaxFrameSize {
		return nil, fmt.Errorf("frame size %d exceeds max %d", header.Length, MaxFrameSize)
	}

	// Check sequence for reorder detection
	c.mu.Lock()
	if header.Sequence < c.seqIn {
		c.mu.Unlock()
		return nil, fmt.Errorf("frame sequence %d is out of order (expected >= %d)", header.Sequence, c.seqIn)
	}
	c.seqIn = header.Sequence + 1
	c.mu.Unlock()

	nonceSize := c.cipher.NonceSize()
	if len(data) < HeaderSize+nonceSize {
		return nil, fmt.Errorf("insufficient data for nonce: got %d", len(data))
	}

	nonce := data[HeaderSize : HeaderSize+nonceSize]
	ciphertext := data[HeaderSize+nonceSize:]

	headerBytes := header.EncodeHeader()
	compressed, err := c.cipher.Open(nil, nonce, ciphertext, headerBytes)
	if err != nil {
		return nil, fmt.Errorf("decrypt failed: %w", err)
	}

	if uint32(len(compressed)) != header.Length {
		return nil, fmt.Errorf("decompressed length mismatch: header says %d, got %d", header.Length, len(compressed))
	}

	// Decompress
	payload, err := c.decompress.DecodeAll(compressed, nil)
	if err != nil {
		return nil, fmt.Errorf("decompress failed: %w", err)
	}

	return &Frame{
		Header:  *header,
		Payload: payload,
	}, nil
}

// HandshakeKeyExchange performs ECDH P-256 key exchange.
func HandshakeKeyExchange(curve ecdh.Curve, privateKey *ecdh.PrivateKey, peerPublicKey *ecdh.PublicKey) ([]byte, error) {
	shared, err := privateKey.ECDH(peerPublicKey)
	if err != nil {
		return nil, fmt.Errorf("ecdh: %w", err)
	}
	// Derive AES-256 key using SHA-256
	hash := sha256.Sum256(shared)
	return hash[:], nil
}

// GenerateEphemeralKeyPair generates a new P-256 key pair for handshake.
func GenerateEphemeralKeyPair() (*ecdh.PrivateKey, *ecdh.PublicKey, error) {
	curve := ecdh.P256()
	priv, err := curve.GenerateKey(rand.Reader)
	if err != nil {
		return nil, nil, fmt.Errorf("generate key: %w", err)
	}
	return priv, priv.PublicKey(), nil
}

// IsControlFrame returns true if the frame type is a control frame.
func IsControlFrame(ft FrameType) bool {
	return ft == FrameTypeHeartbeatPing || ft == FrameTypeHeartbeatPong ||
		ft == FrameTypeError || ft == FrameTypeAuth || ft == FrameTypeClose ||
		ft == FrameTypeControl
}

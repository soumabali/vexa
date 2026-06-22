package crypto

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrPasswordTooShort = errors.New("password must be at least 12 characters")
	ErrPasswordMismatch = errors.New("password does not match")
	ErrMalformedHash    = errors.New("malformed password hash")
)

const (
	argon2Version      = "argon2id$v=19"
	argon2SaltLen      = 16
	argon2KeyLen       = 32
	argon2TimeCostMin  = 3
	argon2MemoryCostKB = 65536
	argon2Threads      = 4
)

// PasswordHasher wraps Argon2id (new default) with backward-compatible bcrypt verification.
type PasswordHasher struct {
	time    uint32
	memory  uint32
	threads uint8
	keyLen  uint32
}

func NewPasswordHasher() *PasswordHasher {
	return &PasswordHasher{
		time:    argon2TimeCostMin,
		memory:  argon2MemoryCostKB,
		threads: argon2Threads,
		keyLen:  argon2KeyLen,
	}
}

// HashPassword hashes a password using Argon2id.
func (h *PasswordHasher) HashPassword(password string) (string, error) {
	return HashPasswordArgon2id(password)
}

// CheckPassword verifies a password against either an Argon2id or legacy bcrypt hash.
func (h *PasswordHasher) CheckPassword(password, hash string) bool {
	if strings.HasPrefix(hash, "$2") {
		return CheckPasswordBcrypt(password, hash)
	}
	return CheckPasswordArgon2id(password, hash)
}

func HashPasswordArgon2id(password string) (string, error) {
	if len(password) < 12 {
		return "", ErrPasswordTooShort
	}
	salt := make([]byte, argon2SaltLen)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}
	hash := argon2.IDKey([]byte(password), salt, argon2TimeCostMin, argon2MemoryCostKB, argon2Threads, argon2KeyLen)
	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)
	return fmt.Sprintf("%s$m=%d,t=%d,p=%d$%s$%s", argon2Version, argon2MemoryCostKB, argon2TimeCostMin, argon2Threads, b64Salt, b64Hash), nil
}

func CheckPasswordArgon2id(password, encodedHash string) bool {
	parts := strings.Split(encodedHash, "$")
	if len(parts) != 5 {
		return false
	}
	if parts[0] != "argon2id" {
		return false
	}
	if parts[1] != "v=19" {
		return false
	}
	var memory uint32
	var time uint32
	var threads uint8
	if _, err := fmt.Sscanf(parts[2], "m=%d,t=%d,p=%d", &memory, &time, &threads); err != nil {
		return false
	}
	if memory < 65536 || time < 3 || threads < 1 {
		return false
	}
	salt, err := base64.RawStdEncoding.DecodeString(parts[3])
	if err != nil {
		return false
	}
	expectedHash, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return false
	}
	if len(salt) < 8 || len(expectedHash) < 16 {
		return false
	}
	candidate := argon2.IDKey([]byte(password), salt, time, memory, threads, uint32(len(expectedHash)))
	return subtle.ConstantTimeCompare(candidate, expectedHash) == 1
}

func HashPasswordBcrypt(password string) (string, error) {
	if len(password) < 12 {
		return "", ErrPasswordTooShort
	}
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func CheckPasswordBcrypt(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func ConstantTimeStringEqual(a, b string) bool {
	return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}


// NeedsRehash returns true if the hash uses legacy/weak parameters and should be upgraded.
func NeedsRehash(hash string) bool {
	if strings.HasPrefix(hash, "$2") {
		return true // migrate bcrypt to Argon2id
	}
	if !strings.HasPrefix(hash, "argon2id$") {
		return true
	}
	parts := strings.Split(hash, "$")
	if len(parts) != 5 {
		return true
	}
	if parts[1] != "v=19" {
		return true
	}
	var memory, time uint32
	var threads uint8
	if _, err := fmt.Sscanf(parts[2], "m=%d,t=%d,p=%d", &memory, &time, &threads); err != nil {
		return true
	}
	return memory < argon2MemoryCostKB || time < argon2TimeCostMin || threads < 1
}

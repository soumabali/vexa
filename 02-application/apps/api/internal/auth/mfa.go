package auth

import (
	"bytes"
	"crypto/rand"
	"encoding/base32"
	"encoding/base64"
	"errors"
	"fmt"
	"image/png"
	"time"

	"github.com/pquerna/otp/totp"
	"github.com/pquerna/otp/hotp"
	"github.com/pquerna/otp"

	"github.com/soumabali/vexa/internal/crypto"
)

var (
	ErrInvalidTOTP      = errors.New("invalid TOTP code")
	ErrInvalidHOTP      = errors.New("invalid HOTP code")
	ErrMFAAlreadyEnabled = errors.New("MFA already enabled")
	ErrMFANotEnabled     = errors.New("MFA not enabled")
)

type MFAService struct {
	issuer string
	encKey []byte
}

func NewMFAService(issuer string, encKey []byte) *MFAService {
	if len(encKey) != 32 {
		panic("MFA encryption key must be 32 bytes")
	}
	return &MFAService{issuer: issuer, encKey: encKey}
}

type TOTPSetup struct {
	Secret          string   `json:"secret"`
	EncryptedSecret string   `json:"encrypted_secret"`
	QRCode          string   `json:"qr_code"`
	URI             string   `json:"uri"`
	BackupCodes     []string `json:"backup_codes"`
}

func (s *MFAService) GenerateTOTPSecret(accountName string) (*TOTPSetup, error) {
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      s.issuer,
		AccountName: accountName,
		SecretSize:  32,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to generate TOTP key: %w", err)
	}

	var buf bytes.Buffer
	img, err := key.Image(200, 200)
	if err != nil {
		return nil, fmt.Errorf("failed to generate QR code: %w", err)
	}
	if err := png.Encode(&buf, img); err != nil {
		return nil, fmt.Errorf("failed to encode QR code: %w", err)
	}

	backupCodes, err := generateBackupCodes(8)
	if err != nil {
		return nil, fmt.Errorf("failed to generate backup codes: %w", err)
	}

	encSecret, err := crypto.EncryptString(s.encKey, key.Secret())
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt TOTP secret: %w", err)
	}
	return &TOTPSetup{
		Secret:          key.Secret(),
		EncryptedSecret: encSecret,
		QRCode:          base64.StdEncoding.EncodeToString(buf.Bytes()),
		URI:             key.URL(),
		BackupCodes:     backupCodes,
	}, nil
}

func (s *MFAService) ValidateTOTP(encryptedSecret, code string) bool {
	secret, err := crypto.DecryptString(s.encKey, encryptedSecret)
	if err != nil {
		return false
	}
	return totp.Validate(code, secret)
}

func (s *MFAService) GenerateHOTPSecret(accountName string, counter uint64) (*otp.Key, error) {
	key, err := hotp.Generate(hotp.GenerateOpts{
		Issuer:      s.issuer,
		AccountName: accountName,
		SecretSize:  32,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to generate HOTP key: %w", err)
	}
	return key, nil
}

func (s *MFAService) ValidateHOTP(secret, code string, counter *uint64) (bool, error) {
	valid := hotp.Validate(code, *counter, secret)
	if valid {
		*counter++
	}
	return valid, nil
}

func generateBackupCodes(count int) ([]string, error) {
	codes := make([]string, count)
	for i := 0; i < count; i++ {
		b := make([]byte, 10)
		if _, err := rand.Read(b); err != nil {
			return nil, err
		}
		codes[i] = base32.StdEncoding.EncodeToString(b)[:10]
	}
	return codes, nil
}

func (s *MFAService) GenerateRecoveryCodes(count int) ([]string, error) {
	return generateBackupCodes(count)
}

func GenerateSecureCode(length int) (string, error) {
	const digits = "0123456789"
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	for i := range b {
		b[i] = digits[int(b[i])%len(digits)]
	}
	return string(b), nil
}

func (s *MFAService) GenerateTOTPCode(secret string) (string, error) {
	code, err := totp.GenerateCode(secret, time.Now())
	if err != nil {
		return "", err
	}
	return code, nil
}

package config

import (
	"crypto/rand"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/soumabali/vexa/internal/security"
)

var secretsDir = resolveSecretsDir()

func init() {
	// Best-effort migration from old /tmp location
	oldSecretsDir := "/tmp/vexa-data/secrets"
	if secretsDir != oldSecretsDir {
		_ = migrateSecrets(oldSecretsDir, secretsDir)
	}
}

func resolveSecretsDir() string {
	if v := os.Getenv("SECRETS_DIR"); v != "" {
		return v
	}
	if getEnv("ENVIRONMENT", "development") == "development" {
		return "/home/ubuntu/projects/vexa/05-config/secrets"
	}
	return "/var/lib/vexa/secrets"
}

func migrateSecrets(oldDir, newDir string) error {
	if _, err := os.Stat(oldDir); os.IsNotExist(err) {
		return nil
	}
	if err := os.MkdirAll(newDir, 0700); err != nil {
		return err
	}
	entries, err := os.ReadDir(oldDir)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		oldPath := filepath.Join(oldDir, entry.Name())
		newPath := filepath.Join(newDir, entry.Name())
		if _, err := os.Stat(newPath); err == nil {
			continue
		}
		data, err := os.ReadFile(oldPath)
		if err != nil {
			continue
		}
		_ = os.WriteFile(newPath, data, 0600)
	}
	return nil
}

// SecretsDir returns the configured persistent secrets directory.
func SecretsDir() string { return secretsDir }

func buildDatabaseURL() string {
	if url := os.Getenv("DATABASE_URL"); url != "" {
		return url
	}
	user := getEnv("DB_USER", "vexa")
	password := getEnv("DB_PASSWORD", "changeme")
	host := getEnv("DB_HOST", "localhost")
	port := getEnv("DB_PORT", "5432")
	dbname := getEnv("DB_NAME", "vexa")
	sslmode := getEnv("DB_SSL_MODE", "disable")
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s", user, password, host, port, dbname, sslmode)
}

type Config struct {
	DatabaseURL       string
	RedisAddr         string
	JWTSecret         []byte
	JWTRefreshSecret  []byte
	AccessTokenTTL    time.Duration
	RefreshTokenTTL   time.Duration
	ServerPort        string
	Environment       string
	MaxRequestSize    int64
	RateLimitRPS      int
	RateLimitBurst    int
	AllowedOrigins    []string
	TLSMinVersion     string
	TLSCertPath       string
	TLSKeyPath        string
	Argon2Memory      uint32
	Argon2Iterations  uint32
	Argon2Parallelism uint8
	Argon2KeyLen      uint32
	EncryptionKey     []byte
	WebAuthnRPID          string
	WebAuthnRPOrigin      string
	WebAuthnRPDisplayName string
}

func Load() *Config {
	// Ensure secrets directory exists
	if err := os.MkdirAll(secretsDir, 0700); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create secrets directory: %v\n", err)
	}

	jwtSecret := loadOrGenerateSecret("JWT_SECRET", filepath.Join(secretsDir, ".jwt_secret"), 64)
	refreshSecret := loadOrGenerateSecret("JWT_REFRESH_SECRET", filepath.Join(secretsDir, ".jwt_refresh_secret"), 64)
		encryptionKey := loadOrGenerateSecret("ENCRYPTION_KEY", filepath.Join(secretsDir, ".encryption_key"), 32)

		allowedOrigins := parseOrigins(getEnv("ALLOWED_ORIGINS", ""))
		if err := validateOrigins(allowedOrigins); err != nil {
			fmt.Fprintf(os.Stderr, "CORS security error: %v\n", err)
			if getEnv("ENVIRONMENT", "development") == "production" {
				panic(err)
			}
		}

		return &Config{
		DatabaseURL:       buildDatabaseURL(),
		RedisAddr:         getEnv("REDIS_ADDR", "localhost:6379"),
		JWTSecret:         jwtSecret,
		JWTRefreshSecret:  refreshSecret,
		AccessTokenTTL:    getDuration("ACCESS_TOKEN_TTL", 15*time.Minute),
		RefreshTokenTTL:   getDuration("REFRESH_TOKEN_TTL", 7*24*time.Hour),
		ServerPort:        getEnv("SERVER_PORT", "8443"),
		Environment:       getEnv("ENVIRONMENT", "development"),
		MaxRequestSize:    getInt64("MAX_REQUEST_SIZE", 1*1024*1024), // 1MB
		RateLimitRPS:      getInt("RATE_LIMIT_RPS", 10),
		RateLimitBurst:    getInt("RATE_LIMIT_BURST", 20),
		AllowedOrigins:    allowedOrigins,
		TLSMinVersion:     getEnv("TLS_MIN_VERSION", "1.3"),
		TLSCertPath:       getEnv("TLS_CERT_PATH", ""),
		TLSKeyPath:        getEnv("TLS_KEY_PATH", ""),
		Argon2Memory:      uint32(getInt("ARGON2_MEMORY", 65536)),
		Argon2Iterations:  uint32(getInt("ARGON2_ITERATIONS", 3)),
		Argon2Parallelism: uint8(getInt("ARGON2_PARALLELISM", 4)),
		Argon2KeyLen:      uint32(getInt("ARGON2_KEYLEN", 32)),
		EncryptionKey:     encryptionKey,
		WebAuthnRPID:          getEnv("WEBAUTHN_RP_ID", "localhost"),
		WebAuthnRPOrigin:      getEnv("WEBAUTHN_RP_ORIGIN", "https://localhost:3000"),
		WebAuthnRPDisplayName: getEnv("WEBAUTHN_RP_DISPLAY_NAME", "vexa"),
	}
}

// loadOrGenerateSecret loads a secret from env var or file, or generates a new one if neither exists.
// This eliminates hardcoded secrets while supporting secret rotation.
func loadOrGenerateSecret(envKey, filePath string, length int) []byte {
	// Priority 1: Environment variable
	if v := os.Getenv(envKey); v != "" {
		return []byte(v)
	}

	// Priority 2: Secret file (persisted across restarts)
	if data, err := os.ReadFile(filePath); err == nil && len(data) > 0 {
		decoded := make([]byte, base64.StdEncoding.DecodedLen(len(data)))
		n, err := base64.StdEncoding.Decode(decoded, data)
		if err == nil {
			return decoded[:n]
		}
		// If not base64, use raw bytes
		return data
	}

	// Priority 3: Generate new random secret
	secret := make([]byte, length)
	if _, err := rand.Read(secret); err != nil {
		panic(fmt.Sprintf("Failed to generate random secret: %v", err))
	}

	// Persist to file with restricted permissions (owner read/write only)
	encoded := base64.StdEncoding.EncodeToString(secret)
	if err := os.WriteFile(filePath, []byte(encoded), 0600); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Failed to persist secret to %s: %v\n", filePath, err)
	}

	return secret
}

// RotateSecrets generates new random secrets and overwrites the persisted files.
// Call this on-demand for secret rotation (e.g., via admin API).
func RotateSecrets(cfg *Config) {
	cfg.JWTSecret = generateRandomSecret(64)
	cfg.JWTRefreshSecret = generateRandomSecret(64)
	cfg.EncryptionKey = generateRandomSecret(32)

	_ = persistSecret(filepath.Join(secretsDir, ".jwt_secret"), cfg.JWTSecret)
	_ = persistSecret(filepath.Join(secretsDir, ".jwt_refresh_secret"), cfg.JWTRefreshSecret)
	_ = persistSecret(filepath.Join(secretsDir, ".encryption_key"), cfg.EncryptionKey)
}

func generateRandomSecret(length int) []byte {
	secret := make([]byte, length)
	if _, err := rand.Read(secret); err != nil {
		panic(fmt.Sprintf("Failed to generate random secret: %v", err))
	}
	return secret
}

func persistSecret(filePath string, secret []byte) error {
	encoded := base64.StdEncoding.EncodeToString(secret)
	return os.WriteFile(filePath, []byte(encoded), 0600)
}

func parseOrigins(originsStr string) []string {
	if originsStr == "" {
		return nil
	}
	parts := strings.Split(originsStr, ",")
	var result []string
	for _, o := range parts {
		trimmed := strings.TrimSpace(o)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	if len(result) == 0 {
		return nil
	}
	return result
}

func validateOrigins(origins []string) error {
	for _, o := range origins {
		if o == "*" {
			return fmt.Errorf("ALLOWED_ORIGINS wildcard '*' is not permitted in production or secure deployments")
		}
	}
	return nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return fallback
}

func getInt64(key string, fallback int64) int64 {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.ParseInt(v, 10, 64); err == nil {
			return i
		}
	}
	return fallback
}

func getDuration(key string, fallback time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return fallback
}

func (c *Config) IsProduction() bool {
	return c.Environment == "production"
}

func (c *Config) GetTLSConfig() (*tls.Config, error) {
	if c.TLSCertPath == "" || c.TLSKeyPath == "" {
		return nil, nil
	}
	certPEM, err := os.ReadFile(c.TLSCertPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read TLS cert: %w", err)
	}
	keyPEM, err := os.ReadFile(c.TLSKeyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read TLS key: %w", err)
	}
	return security.SecureServerConfig(certPEM, keyPEM)
}

package audit

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/google/uuid"
)

const (
	EventAuthLogin        = "auth.login"
	EventAuthLogout       = "auth.logout"
	EventAuthMFA          = "auth.mfa"
	EventAuthFailed       = "auth.failed"
	EventVaultUnlock      = "vault.unlock"
	EventVaultLock        = "vault.lock"
	EventCredentialAccess = "credential.access"
	EventCredentialCreate = "credential.create"
	EventCredentialUpdate = "credential.update"
	EventCredentialDelete = "credential.delete"
	EventCredentialShare  = "credential.share"
	EventCredentialRevoke = "credential.revoke"
	EventCredentialRotate = "credential.rotate"
	EventMasterKeyRotate  = "master_key.rotate"
	EventHostCreate       = "host.create"
	EventHostUpdate       = "host.update"
	EventHostDelete       = "host.delete"
	EventHostConnect      = "host.connect"
	EventHostDisconnect   = "host.disconnect"
	EventSessionStart     = "session.start"
	EventSessionEnd       = "session.end"
	EventSessionKill      = "session.kill"
	EventAdminAction      = "admin.action"
	EventConfigChange     = "config.change"
	EventSystemStartup    = "system.startup"
	EventSystemShutdown   = "system.shutdown"
	EventSecurityAlert    = "security.alert"

	// WireGuard tunnel events
	EventTunnelCreate  = "tunnel.create"
	EventTunnelDelete  = "tunnel.delete"
	EventTunnelEnable  = "tunnel.enable"
	EventTunnelDisable = "tunnel.disable"
	EventTunnelRotate  = "tunnel.rotate"
	EventTunnelUpdate  = "tunnel.update"
)

type Logger struct {
	mu           sync.RWMutex
	db           *sql.DB
	file         *os.File
	filePath     string
	hmacKey      []byte
	previousHash string
	redactFields []string
}

type LogEntry struct {
	ID            int64                  `json:"id"`
	Timestamp     time.Time              `json:"timestamp"`
	EventType     string                 `json:"event_type"`
	UserID        *uuid.UUID             `json:"user_id,omitempty"`
	DeviceID      *uuid.UUID             `json:"device_id,omitempty"`
	IPAddress     string                 `json:"ip_address,omitempty"`
	Details       map[string]interface{} `json:"details"`
	IntegrityHash string                 `json:"integrity_hash"`
}

// NewTestLogger creates a minimal audit logger for unit tests.
func NewTestLogger(t interface{ TempDir() string }) (*Logger, error) {
	filePath := filepath.Join(t.TempDir(), "audit.log")
	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return nil, err
	}
	return &Logger{
		file:         file,
		filePath:     filePath,
		hmacKey:      make([]byte, 32),
		redactFields: []string{"password", "secret", "token", "key", "credential", "private_key", "passphrase"},
	}, nil
}

// NewNoOpLogger creates a file-only audit logger that writes to a temporary file.
func NewNoOpLogger() (*Logger, error) {
	dir, err := os.MkdirTemp("", "audit-noop-*")
	if err != nil {
		return nil, err
	}
	filePath := filepath.Join(dir, "audit.log")
	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return nil, err
	}
	return &Logger{
		file:         file,
		filePath:     filePath,
		hmacKey:      make([]byte, 32),
		redactFields: []string{"password", "secret", "token", "key", "credential", "private_key", "passphrase"},
	}, nil
}

func NewLogger(db *sql.DB, filePath string, hmacKey []byte) (*Logger, error) {
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0750); err != nil {
		return nil, fmt.Errorf("failed to create audit log directory: %w", err)
	}

	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return nil, fmt.Errorf("failed to open audit log file: %w", err)
	}

	return &Logger{
		db:           db,
		file:         file,
		filePath:     filePath,
		hmacKey:      hmacKey,
		redactFields: []string{"password", "secret", "token", "key", "credential", "private_key", "passphrase"},
	}, nil
}

func (l *Logger) Log(eventType string, userID, deviceID *uuid.UUID, ipAddress string, details map[string]interface{}) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	redactedDetails := l.redactSensitiveData(details)

	entry := &LogEntry{
		Timestamp: time.Now().UTC(),
		EventType: eventType,
		UserID:    userID,
		DeviceID:  deviceID,
		IPAddress: sanitizeIP(ipAddress),
		Details:   redactedDetails,
	}

	entry.IntegrityHash = l.computeHash(entry)

	data, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("failed to marshal audit log entry: %w", err)
	}

	if _, err := l.file.WriteString(string(data) + "\n"); err != nil {
		return fmt.Errorf("failed to write to audit log file: %w", err)
	}

	if l.db != nil {
		query := `
			INSERT INTO audit_log (timestamp, event_type, user_id, device_id, ip_address, details, integrity_hash)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
		`
		_, err := l.db.Exec(query, entry.Timestamp, entry.EventType, entry.UserID, entry.DeviceID,
			net.ParseIP(entry.IPAddress), entry.Details, entry.IntegrityHash)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to write audit log to database: %v\n", err)
		}
	}

	l.previousHash = entry.IntegrityHash
	return nil
}

func (l *Logger) computeHash(entry *LogEntry) string {
	h := hmac.New(sha256.New, l.hmacKey)
	data, _ := json.Marshal(struct {
		Timestamp    time.Time              `json:"timestamp"`
		EventType    string                 `json:"event_type"`
		UserID       *uuid.UUID             `json:"user_id,omitempty"`
		DeviceID     *uuid.UUID             `json:"device_id,omitempty"`
		IPAddress    string                 `json:"ip_address,omitempty"`
		Details      map[string]interface{} `json:"details"`
		PreviousHash string                 `json:"previous_hash"`
	}{
		Timestamp:    entry.Timestamp,
		EventType:    entry.EventType,
		UserID:       entry.UserID,
		DeviceID:     entry.DeviceID,
		IPAddress:    entry.IPAddress,
		Details:      entry.Details,
		PreviousHash: l.previousHash,
	})
	h.Write(data)
	return hex.EncodeToString(h.Sum(nil))
}

func (l *Logger) redactSensitiveData(data map[string]interface{}) map[string]interface{} {
	redacted := make(map[string]interface{})
	for key, value := range data {
		if l.isSensitiveField(key) {
			redacted[key] = "[REDACTED]"
		} else {
			switch v := value.(type) {
			case map[string]interface{}:
				redacted[key] = l.redactSensitiveData(v)
			default:
				redacted[key] = value
			}
		}
	}
	return redacted
}

func (l *Logger) isSensitiveField(field string) bool {
	for _, redact := range l.redactFields {
		if field == redact {
			return true
		}
	}
	return false
}

func sanitizeIP(ip string) string {
	if ip == "" {
		return ""
	}
	parsed := net.ParseIP(ip)
	if parsed == nil {
		return ""
	}
	return parsed.String()
}

func (l *Logger) RotateLog() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.file != nil {
		l.file.Close()
	}

	timestamp := time.Now().UTC().Format("2006-01-02_15-04-05")
	newPath := fmt.Sprintf("%s.%s", l.filePath, timestamp)
	if err := os.Rename(l.filePath, newPath); err != nil {
		return fmt.Errorf("failed to rotate audit log: %w", err)
	}

	file, err := os.OpenFile(l.filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return fmt.Errorf("failed to open new audit log file: %w", err)
	}
	l.file = file
	return nil
}

func (l *Logger) Export(start, end time.Time, format string) ([]byte, error) {
	if format != "json" && format != "csv" {
		return nil, fmt.Errorf("unsupported export format: %s", format)
	}

	entries, err := l.Query(start, end, 10000, 0)
	if err != nil {
		return nil, err
	}

	if format == "json" {
		return json.MarshalIndent(entries, "", "  ")
	}

	var buf bytes.Buffer
	buf.WriteString("id,timestamp,event_type,user_id,ip_address,details\n")
	for _, e := range entries {
		details, _ := json.Marshal(e.Details)
		uid := ""
		if e.UserID != nil {
			uid = e.UserID.String()
		}
		buf.WriteString(fmt.Sprintf("%d,%s,%s,%s,%s,%s\n", e.ID, e.Timestamp.Format(time.RFC3339), e.EventType, uid, e.IPAddress, string(details)))
	}
	return buf.Bytes(), nil
}

func (l *Logger) Query(start, end time.Time, limit, offset int) ([]*LogEntry, error) {
	if l.db == nil {
		// File-only mode: return empty results; file can be read separately if needed
		return []*LogEntry{}, nil
	}

	query := `
		SELECT id, timestamp, event_type, user_id, device_id, ip_address, details, integrity_hash
		FROM audit_log
		WHERE timestamp BETWEEN $1 AND $2
		ORDER BY timestamp DESC
		LIMIT $3 OFFSET $4
	`
	rows, err := l.db.Query(query, start, end, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query audit log: %w", err)
	}
	defer rows.Close()

	var entries []*LogEntry
	for rows.Next() {
		var entry LogEntry
		var userID, deviceID sql.NullString
		var detailsJSON []byte
		err := rows.Scan(
			&entry.ID, &entry.Timestamp, &entry.EventType,
			&userID, &deviceID, &entry.IPAddress,
			&detailsJSON, &entry.IntegrityHash,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan audit log entry: %w", err)
		}
		if userID.Valid {
			uid, err := uuid.Parse(userID.String)
			if err != nil {
				return nil, fmt.Errorf("invalid user_id in audit log: %w", err)
			}
			entry.UserID = &uid
		}
		if deviceID.Valid {
			did, err := uuid.Parse(deviceID.String)
			if err != nil {
				return nil, fmt.Errorf("invalid device_id in audit log: %w", err)
			}
			entry.DeviceID = &did
		}
		if err := json.Unmarshal(detailsJSON, &entry.Details); err != nil {
			return nil, fmt.Errorf("failed to unmarshal audit log details: %w", err)
		}
		entries = append(entries, &entry)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate audit log rows: %w", err)
	}
	return entries, nil
}

// Close closes the underlying audit log file.
func (l *Logger) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.file != nil {
		return l.file.Close()
	}
	return nil
}

// VerifyIntegrity verifies the integrity hash chain of a set of entries.
func (l *Logger) VerifyIntegrity(entries []*LogEntry) bool {
	previousHash := ""
	for _, entry := range entries {
		data, _ := json.Marshal(struct {
			Timestamp    time.Time              `json:"timestamp"`
			EventType    string                 `json:"event_type"`
			UserID       *uuid.UUID             `json:"user_id,omitempty"`
			DeviceID     *uuid.UUID             `json:"device_id,omitempty"`
			IPAddress    string                 `json:"ip_address,omitempty"`
			Details      map[string]interface{} `json:"details"`
			PreviousHash string                 `json:"previous_hash"`
		}{
			Timestamp:    entry.Timestamp,
			EventType:    entry.EventType,
			UserID:       entry.UserID,
			DeviceID:     entry.DeviceID,
			IPAddress:    entry.IPAddress,
			Details:      entry.Details,
			PreviousHash: previousHash,
		})
		h := hmac.New(sha256.New, l.hmacKey)
		h.Write(data)
		expectedHash := hex.EncodeToString(h.Sum(nil))
		if expectedHash != entry.IntegrityHash {
			return false
		}
		previousHash = entry.IntegrityHash
	}
	return true
}

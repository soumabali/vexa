package vault

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

// Sentinel errors for vault repository.
var (
	ErrCredentialNotFound = errors.New("credential not found")
	ErrVaultLocked        = errors.New("vault is locked")
	ErrInvalidKeyVersion  = errors.New("invalid key version")
)

// CredentialRepository defines the interface for credential persistence.
// It abstracts the database operations so the service layer remains
// independent of storage details.
type CredentialRepository interface {
	// Create inserts a new credential and returns the created record.
	Create(ctx context.Context, cred *Credential, encryptedData, nonce, salt []byte) error

	// GetByID retrieves a credential by ID for a specific owner.
	// Returns ErrCredentialNotFound if not found.
	GetByID(ctx context.Context, ownerID, credentialID uuid.UUID) (*Credential, error)

	// List returns all credentials for a user, optionally filtered by tag.
	// Returns credentials with encrypted data still encrypted.
	List(ctx context.Context, ownerID uuid.UUID, tag string, limit, offset int) ([]*Credential, error)

	// Update updates a credential's metadata and/or re-encrypts its data.
	Update(ctx context.Context, credentialID uuid.UUID, ownerID uuid.UUID,
		updates *UpdateCredentialInput,
		encryptedData, nonce []byte) error

	// SoftDelete marks a credential as deleted (sets deleted_at).
	Delete(ctx context.Context, ownerID, credentialID uuid.UUID) error

	// Count returns the total number of credentials for a user.
	Count(ctx context.Context, ownerID uuid.UUID) (int, error)
}

// PostgresCredentialRepository implements CredentialRepository using PostgreSQL.
type PostgresCredentialRepository struct {
	db *sql.DB
}

// NewPostgresCredentialRepository creates a new Postgres-backed repository.
func NewPostgresCredentialRepository(db *sql.DB) *PostgresCredentialRepository {
	return &PostgresCredentialRepository{db: db}
}

// Create inserts a new credential.
func (r *PostgresCredentialRepository) Create(ctx context.Context, cred *Credential, encryptedData, nonce, salt []byte) error {
	query := `
		INSERT INTO credentials
			(id, owner_id, name, type, encrypted_data, nonce, salt, version, tags, expires_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`
	if len(cred.Tags) > 0 {
		_, _ = json.Marshal(cred.Tags)
	}

	_, err := r.db.ExecContext(ctx, query,
		cred.ID, cred.OwnerID, cred.Name, string(cred.Type),
		encryptedData, nonce, salt, cred.Version,
		pq.Array(cred.Tags), cred.ExpiresAt,
		cred.CreatedAt, cred.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("inserting credential: %w", err)
	}
	return nil
}

// GetByID retrieves a credential by ID and owner.
func (r *PostgresCredentialRepository) GetByID(ctx context.Context, ownerID, credentialID uuid.UUID) (*Credential, error) {
	query := `
		SELECT id, owner_id, name, type, encrypted_data, nonce, salt, version,
		       tags, last_rotated_at, expires_at, created_at, updated_at
		FROM credentials
		WHERE id = $1 AND owner_id = $2 AND deleted_at IS NULL
	`
	var cred Credential
	var tagsBytes []byte
	var expiresAt, lastRotated sql.NullTime

	err := r.db.QueryRowContext(ctx, query, credentialID, ownerID).Scan(
		&cred.ID, &cred.OwnerID, &cred.Name, &cred.Type,
		&cred.EncryptedData, &cred.Nonce, &cred.Salt, &cred.Version,
		&tagsBytes, &lastRotated, &expiresAt, &cred.CreatedAt, &cred.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, ErrCredentialNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("querying credential: %w", err)
	}

	cred.Tags = parseTags(tagsBytes)
	if lastRotated.Valid {
		cred.LastRotatedAt = &lastRotated.Time
	}
	if expiresAt.Valid {
		cred.ExpiresAt = &expiresAt.Time
	}

	return &cred, nil
}

// List returns credentials for a user, optionally filtered.
func (r *PostgresCredentialRepository) List(ctx context.Context, ownerID uuid.UUID, tag string, limit, offset int) ([]*Credential, error) {
	if limit <= 0 {
		limit = 100
	}

	query := `
		SELECT id, owner_id, name, type, encrypted_data, nonce, salt, version,
		       tags, last_rotated_at, expires_at, created_at, updated_at
		FROM credentials
		WHERE owner_id = $1 AND deleted_at IS NULL
	`
	args := []interface{}{ownerID}
	argIdx := 2

	if tag != "" {
		query += fmt.Sprintf(" AND $%d = ANY(tags)", argIdx)
		args = append(args, tag)
		argIdx++
	}

	query += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d", argIdx)
	args = append(args, limit)
	argIdx++

	if offset > 0 {
		query += fmt.Sprintf(" OFFSET $%d", argIdx)
		args = append(args, offset)
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("listing credentials: %w", err)
	}
	defer rows.Close()

	var creds []*Credential
	for rows.Next() {
		var cred Credential
		var tagsBytes []byte
		var expiresAt, lastRotated sql.NullTime

		err := rows.Scan(
			&cred.ID, &cred.OwnerID, &cred.Name, &cred.Type,
			&cred.EncryptedData, &cred.Nonce, &cred.Salt, &cred.Version,
			&tagsBytes, &lastRotated, &expiresAt, &cred.CreatedAt, &cred.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		cred.Tags = parseTags(tagsBytes)
		if lastRotated.Valid {
			cred.LastRotatedAt = &lastRotated.Time
		}
		if expiresAt.Valid {
			cred.ExpiresAt = &expiresAt.Time
		}
		creds = append(creds, &cred)
	}
	return creds, rows.Err()
}

// Update updates a credential.
func (r *PostgresCredentialRepository) Update(ctx context.Context, credentialID, ownerID uuid.UUID,
	updates *UpdateCredentialInput, encryptedData, nonce []byte) error {

	query := `
		UPDATE credentials SET
			name = COALESCE($1, name),
			encrypted_data = COALESCE($2, encrypted_data),
			nonce = COALESCE($3, nonce),
			version = version + 1,
			tags = COALESCE($4, tags),
			last_rotated_at = CASE WHEN $2 IS NOT NULL THEN NOW() ELSE last_rotated_at END,
			expires_at = COALESCE($5, expires_at),
			updated_at = NOW()
		WHERE id = $6 AND owner_id = $7 AND deleted_at IS NULL
	`
	if updates != nil && len(updates.Tags) > 0 {
		_, _ = json.Marshal(updates.Tags)
	}

	result, err := r.db.ExecContext(ctx, query,
		nullString(updates.Name),
		nullBytes(encryptedData),
		nullBytes(nonce),
		nullBytes(nil),
		nullTime(updates.ExpiresAt),
		credentialID, ownerID,
	)
	if err != nil {
		return fmt.Errorf("updating credential: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrCredentialNotFound
	}
	return nil
}

// Delete performs a soft delete on a credential.
func (r *PostgresCredentialRepository) Delete(ctx context.Context, ownerID, credentialID uuid.UUID) error {
	query := `
		UPDATE credentials SET deleted_at = NOW(), updated_at = NOW()
		WHERE id = $1 AND owner_id = $2 AND deleted_at IS NULL
	`
	result, err := r.db.ExecContext(ctx, query, credentialID, ownerID)
	if err != nil {
		return fmt.Errorf("deleting credential: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrCredentialNotFound
	}
	return nil
}

// Count returns the total number of non-deleted credentials for a user.
func (r *PostgresCredentialRepository) Count(ctx context.Context, ownerID uuid.UUID) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM credentials WHERE owner_id = $1 AND deleted_at IS NULL",
		ownerID,
	).Scan(&count)
	return count, err
}

// --- Vault Key Repository ---

// VaultKeyRepository manages master key versions per user.
type VaultKeyRepository interface {
	// GetLatest returns the most recent key version for a user.
	// Returns nil if no key exists yet.
	GetLatest(ctx context.Context, userID uuid.UUID) (*VaultKey, error)

	// Store creates a new key version.
	Store(ctx context.Context, key *VaultKey) error

	// GetByVersion returns a specific key version.
	GetByVersion(ctx context.Context, userID uuid.UUID, version int) (*VaultKey, error)
}

// PostgresVaultKeyRepository implements VaultKeyRepository.
type PostgresVaultKeyRepository struct {
	db *sql.DB
}

// NewPostgresVaultKeyRepository creates a new Postgres key repository.
func NewPostgresVaultKeyRepository(db *sql.DB) *PostgresVaultKeyRepository {
	return &PostgresVaultKeyRepository{db: db}
}

// GetLatest returns the most recent key version.
func (r *PostgresVaultKeyRepository) GetLatest(ctx context.Context, userID uuid.UUID) (*VaultKey, error) {
	query := `
		SELECT id, user_id, version, salt, params, rotated_at, rotated_by
		FROM master_key_versions
		WHERE user_id = $1
		ORDER BY version DESC
		LIMIT 1
	`
	var vk VaultKey
	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&vk.ID, &vk.UserID, &vk.Version, &vk.Salt, &vk.Params, &vk.RotatedAt, &vk.RotatedBy,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("getting latest key version: %w", err)
	}
	return &vk, nil
}

// Store creates a new key version.
func (r *PostgresVaultKeyRepository) Store(ctx context.Context, key *VaultKey) error {
	query := `
		INSERT INTO master_key_versions (id, user_id, version, salt, params, rotated_at, rotated_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (user_id, version) DO UPDATE SET
			salt = EXCLUDED.salt,
			params = EXCLUDED.params,
			rotated_at = NOW()
	`
	_, err := r.db.ExecContext(ctx, query,
		key.ID, key.UserID, key.Version, key.Salt, key.Params, key.RotatedAt, key.RotatedBy,
	)
	return err
}

// GetByVersion returns a specific key version.
func (r *PostgresVaultKeyRepository) GetByVersion(ctx context.Context, userID uuid.UUID, version int) (*VaultKey, error) {
	query := `
		SELECT id, user_id, version, salt, params, rotated_at, rotated_by
		FROM master_key_versions
		WHERE user_id = $1 AND version = $2
	`
	var vk VaultKey
	err := r.db.QueryRowContext(ctx, query, userID, version).Scan(
		&vk.ID, &vk.UserID, &vk.Version, &vk.Salt, &vk.Params, &vk.RotatedAt, &vk.RotatedBy,
	)
	if err == sql.ErrNoRows {
		return nil, ErrInvalidKeyVersion
	}
	if err != nil {
		return nil, fmt.Errorf("getting key version: %w", err)
	}
	return &vk, nil
}

// --- Helpers ---

// parseTags parses PostgreSQL TEXT[] format: {a,b,c} or {a,"b with space",c}
func parseTags(data []byte) []string {
	if len(data) < 2 {
		return nil
	}
	s := string(data[1 : len(data)-1])
	if s == "" {
		return nil
	}
	var result []string
	inQuote := false
	var current strings.Builder
	for i := 0; i < len(s); i++ {
		switch s[i] {
		case '"':
			inQuote = !inQuote
		case ',':
			if !inQuote {
				result = append(result, current.String())
				current.Reset()
				continue
			}
			current.WriteByte(s[i])
		default:
			current.WriteByte(s[i])
		}
	}
	if current.Len() > 0 {
		result = append(result, current.String())
	}
	return result
}

func nullString(s *string) interface{} {
	if s == nil {
		return nil
	}
	return *s
}

func nullBytes(b []byte) interface{} {
	if len(b) == 0 {
		return nil
	}
	return b
}

func nullTime(t *time.Time) interface{} {
	if t == nil {
		return nil
	}
	return *t
}
package vault

import (
	"context"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"

	"github.com/soumabali/vexa/internal/crypto"
)

// CredentialService provides credential CRUD operations with encryption.
// All operations require the vault to be unlocked for the user.
//
// The service uses Redis to track per-user vault session state (unlocked/locked),
// and the database for credential persistence.
//
// This is the service layer that the HTTP handlers use. The handlers expect
// these methods in the exact signature below.
type CredentialService struct {
	db    *sql.DB
	redis *redis.Client
	vault *crypto.Vault
}

// NewCredentialService creates a new CredentialService backed by PostgreSQL and Redis.
func NewCredentialService(db *sql.DB, redisClient *redis.Client) *CredentialService {
	return &CredentialService{
		db:    db,
		redis: redisClient,
		vault: crypto.NewVault(),
	}
}

// vaultState is stored in Redis for each user.
type vaultState struct {
	Unlocked   bool   `json:"unlocked"`
	KeyVersion int    `json:"key_version"`
	Salt       string `json:"salt,omitempty"`
}

func (s *CredentialService) vaultRedisKey(userID uuid.UUID) string {
	return "vault:session:" + userID.String()
}

// --- Vault lifecycle ---

// Unlock derives the master key from the password and unlocks the vault.
// If no key version exists for this user yet, a new salt and key version are generated.
// The vault must be unlocked before any credential operations.
func (s *CredentialService) Unlock(ctx context.Context, userID uuid.UUID, masterPassword string) error {
	// Check if already unlocked
	if raw, err := s.redis.Get(ctx, s.vaultRedisKey(userID)).Result(); err == nil && raw != "" {
		var st vaultState
		if json.Unmarshal([]byte(raw), &st) == nil && st.Unlocked {
			return ErrVaultAlreadyUnlocked
		}
	}

	// Get latest key version (if any)
	mkv, err := s.getLatestKeyVersion(ctx, userID)
	if err != nil && err != sql.ErrNoRows {
		return fmt.Errorf("unlock: getting key version: %w", err)
	}

	var salt []byte
	var params crypto.Argon2Params

	if mkv == nil {
		// First unlock: generate new salt and params
		salt, err = crypto.GenerateSalt()
		if err != nil {
			return fmt.Errorf("unlock: salt generation: %w", err)
		}
		params = crypto.DefaultArgon2Params()
	} else {
		salt = mkv.Salt
		params = decodeParams(mkv.Params)
	}

	// Derive the master key from password
	masterKey := crypto.DeriveKeyWithParams(masterPassword, salt, params)
	// Note: masterKey is passed to vault; the local copy will be cleared when
	// the vault locks or the variable goes out of scope.

	s.vault.UnlockWithKey(masterKey, salt, params, mkv.Version+1)

	// Store session state in Redis
	state := vaultState{
		Unlocked:   true,
		KeyVersion: mkv.Version + 1,
		Salt:       base64.StdEncoding.EncodeToString(salt),
	}
	stateJSON, _ := json.Marshal(state)
	if err := s.redis.Set(ctx, s.vaultRedisKey(userID), string(stateJSON), 1*time.Hour).Err(); err != nil {
		s.vault.Lock()
		return fmt.Errorf("unlock: storing session state: %w", err)
	}

	// If this is the first key version, store it
	if mkv == nil {
		mkv = &keyVersionRow{
			ID:        uuid.New(),
			Version:   1,
			Salt:      salt,
			Params:    encodeParams(params),
			RotatedAt: time.Now().UTC(),
			RotatedBy: userID,
		}
		if err := s.storeKeyVersion(ctx, userID, mkv); err != nil {
			s.vault.Lock()
			return fmt.Errorf("unlock: storing key version: %w", err)
		}
	}

	return nil
}

// Lock clears the in-process master key and removes the Redis session.
func (s *CredentialService) Lock(ctx context.Context, userID uuid.UUID) {
	s.vault.Lock()
	_ = s.redis.Del(ctx, s.vaultRedisKey(userID)).Err()
}

// GetStatus returns the current vault unlock status for a user.
func (s *CredentialService) GetStatus(ctx context.Context, userID uuid.UUID) *VaultStatus {
	mkv, _ := s.getLatestKeyVersion(ctx, userID)
	unlocked := s.isUnlocked(ctx, userID)
	version := 0
	if mkv != nil {
		version = mkv.Version
	}
	return &VaultStatus{Unlocked: unlocked, KeyVersion: version}
}

// RotateMasterKey re-encrypts all credentials with a new master key.
// The vault must be unlocked. The new password derives the new key.
// All credentials are decrypted with the old key and re-encrypted with the new key.
func (s *CredentialService) RotateMasterKey(ctx context.Context, userID uuid.UUID, newPassword string) error {
	if !s.isUnlocked(ctx, userID) {
		return ErrVaultLocked
	}

	mkv, err := s.getLatestKeyVersion(ctx, userID)
	if err != nil {
		return err
	}

	oldSalt := mkv.Salt
	oldParams := decodeParams(mkv.Params)
	// Note: old key is the same as current vault key (derived during Unlock)
	_ = oldSalt // silence unused (vault uses old key via mkv.Salt/params but we re-derive below)
	_ = oldParams

	// Re-derive old key for decryption
	// Since the vault was already unlocked with this key, we get it from vault state
	// But for rotation we need the old key separately, so we re-derive from stored salt
	_ = crypto.DeriveKeyWithParams(newPassword, mkv.Salt, decodeParams(mkv.Params)) // oldKey — currently unused

	// Generate new salt and derive new key
	newSalt, err := crypto.GenerateSalt()
	if err != nil {
		return fmt.Errorf("rotate: generating salt: %w", err)
	}
	newParams := crypto.DefaultArgon2Params()
	newKey := crypto.DeriveKeyWithParams(newPassword, newSalt, newParams)
	newVersion := mkv.Version + 1

	// Re-encrypt each credential
	creds, err := s.listCredentials(ctx, userID, "", 0, 0)
	if err != nil {
		return fmt.Errorf("rotate: listing credentials: %w", err)
	}

	for _, cred := range creds {
		encrypted, err := base64.StdEncoding.DecodeString(cred.EncryptedData)
		if err != nil {
			continue
		}
		nonce, err := base64.StdEncoding.DecodeString(cred.Nonce)
		if err != nil {
			continue
		}

		decrypted, err := s.vault.DecryptCredential(encrypted, nonce)
		if err != nil {
			continue
		}

		// Build a temporary new vault just for re-encryption
		tmpVault := crypto.NewVaultFromKey(newKey)
		tmpVault.UnlockWithKey(newKey, newSalt, newParams, newVersion)
		reEncrypted, newNonce, _, err := tmpVault.EncryptCredential(decrypted)
		if err != nil {
			return fmt.Errorf("re-encrypting credential %s: %w", cred.ID, err)
		}

		now := time.Now().UTC()
		_, err = s.db.ExecContext(ctx,
			`UPDATE credentials SET encrypted_data=$1, nonce=$2, version=version+1,
			 last_rotated_at=$3, updated_at=$3 WHERE id=$4`,
			base64.StdEncoding.EncodeToString(reEncrypted),
			base64.StdEncoding.EncodeToString(newNonce),
			now, cred.ID,
		)
		if err != nil {
			return fmt.Errorf("storing rotated credential: %w", err)
		}
	}

	// Store new key version
	newMKV := &keyVersionRow{
		ID:        uuid.New(),
		Version:   newVersion,
		Salt:      newSalt,
		Params:    encodeParams(newParams),
		RotatedAt: time.Now().UTC(),
		RotatedBy: userID,
	}
	if err := s.storeKeyVersion(ctx, userID, newMKV); err != nil {
		return fmt.Errorf("storing new key version: %w", err)
	}

	// Update in-process vault with new key
	s.vault.UnlockWithKey(newKey, newSalt, newParams, newVersion)

	// Update Redis session
	state := vaultState{
		Unlocked:   true,
		KeyVersion: newVersion,
		Salt:       base64.StdEncoding.EncodeToString(newSalt),
	}
	stateJSON, _ := json.Marshal(state)
	_ = s.redis.Set(ctx, s.vaultRedisKey(userID), string(stateJSON), 1*time.Hour).Err()

	return nil
}

// --- Credential CRUD ---

// Create creates a new credential after encrypting its data.
// The vault must be unlocked.
// req.Type must be one of: "password", "ssh_key", "api_token".
func (s *CredentialService) Create(ctx context.Context, ownerID uuid.UUID, req *CreateCredentialRequest) (*Credential, error) {
	if !s.isUnlocked(ctx, ownerID) {
		return nil, ErrVaultLocked
	}

	credTypeEnum := CredentialType(req.Type)
	credData := &crypto.CredentialData{}
	switch credTypeEnum {
	case CredentialTypePassword:
		credData.Password = req.Data
	case CredentialTypeSSHKey:
		credData.PrivateKey = req.Data
	case CredentialTypeAPIToken:
		credData.APIToken = req.Data
	}

	encrypted, nonce, _, err := s.vault.EncryptCredential(credData)
	if err != nil {
		return nil, fmt.Errorf("create: encrypting: %w", err)
	}

	now := time.Now().UTC()
	cred := &Credential{
		ID:            uuid.New(),
		OwnerID:       ownerID,
		Name:          req.Name,
		Type:          credTypeEnum,
		EncryptedData: base64.StdEncoding.EncodeToString(encrypted),
		Nonce:         base64.StdEncoding.EncodeToString(nonce),
		Version:       1,
		Tags:          req.Tags,
		ExpiresAt:     req.Expiry,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	query := `
		INSERT INTO credentials (id, owner_id, name, type, encrypted_data, nonce, version, tags, expires_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`
	_, err = s.db.ExecContext(ctx, query,
		cred.ID, cred.OwnerID, cred.Name, cred.Type,
		cred.EncryptedData, cred.Nonce, cred.Version,
		pqArray(cred.Tags), cred.ExpiresAt,
		cred.CreatedAt, cred.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("create: storing: %w", err)
	}
	return cred, nil
}

// List returns all credentials for a user (metadata only, no decrypted data).
// The vault does not need to be unlocked for listing.
func (s *CredentialService) List(ctx context.Context, ownerID uuid.UUID, tag string, limit, offset int) ([]*Credential, error) {
	return s.listCredentials(ctx, ownerID, tag, limit, offset)
}

// GetByID retrieves a credential by ID (metadata only, no decrypted data).
// The vault does not need to be unlocked for getting metadata.
func (s *CredentialService) GetByID(ctx context.Context, ownerID, credentialID uuid.UUID) (*Credential, error) {
	return s.getCredential(ctx, ownerID, credentialID)
}

// Update updates a credential's metadata and/or re-encrypts its data.
// The vault must be unlocked.
func (s *CredentialService) Update(ctx context.Context, ownerID, credentialID uuid.UUID, req *UpdateCredentialRequest) (*Credential, error) {
	if !s.isUnlocked(ctx, ownerID) {
		return nil, ErrVaultLocked
	}

	cred, err := s.getCredential(ctx, ownerID, credentialID)
	if err != nil {
		return nil, err
	}

	// Update metadata if provided
	if req.Name != "" {
		cred.Name = req.Name
	}
	if req.Tags != nil {
		cred.Tags = req.Tags
	}
	if req.Expiry != nil {
		cred.ExpiresAt = req.Expiry
	}

	// Re-encrypt data if provided
	if req.Data != "" {
		credData := &crypto.CredentialData{}
		switch cred.Type {
		case CredentialTypePassword:
			credData.Password = req.Data
		case CredentialTypeSSHKey:
			credData.PrivateKey = req.Data
		case CredentialTypeAPIToken:
			credData.APIToken = req.Data
		}
		encrypted, nonce, _, err := s.vault.EncryptCredential(credData)
		if err != nil {
			return nil, fmt.Errorf("update: encrypting: %w", err)
		}
		cred.EncryptedData = base64.StdEncoding.EncodeToString(encrypted)
		cred.Nonce = base64.StdEncoding.EncodeToString(nonce)
		cred.Version++
		now := time.Now().UTC()
		cred.LastRotatedAt = &now
	}

	cred.UpdatedAt = time.Now().UTC()

	query := `
		UPDATE credentials SET name=$1, encrypted_data=$2, nonce=$3, version=$4,
			tags=$5, last_rotated_at=$6, expires_at=$7, updated_at=$8
		WHERE id=$9 AND owner_id=$10
	`
	_, err = s.db.ExecContext(ctx, query,
		cred.Name, cred.EncryptedData, cred.Nonce, cred.Version,
		pqArray(cred.Tags), cred.LastRotatedAt, cred.ExpiresAt, cred.UpdatedAt,
		credentialID, ownerID,
	)
	if err != nil {
		return nil, fmt.Errorf("update: updating: %w", err)
	}
	return cred, nil
}

// Delete soft-deletes a credential (hard delete for now since soft delete uses deleted_at).
// The vault must be unlocked.
func (s *CredentialService) Delete(ctx context.Context, ownerID, credentialID uuid.UUID) error {
	if !s.isUnlocked(ctx, ownerID) {
		return ErrVaultLocked
	}
	query := `DELETE FROM credentials WHERE id=$1 AND owner_id=$2`
	result, err := s.db.ExecContext(ctx, query, credentialID, ownerID)
	if err != nil {
		return fmt.Errorf("delete: deleting: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrCredentialNotFound
	}
	return nil
}

// DecryptData decrypts and returns the full credential data for a credential.
// The vault must be unlocked.
// Returns a crypto.CredentialData with username/password/private_key/etc fields populated.
func (s *CredentialService) DecryptData(ctx context.Context, ownerID, credentialID uuid.UUID) (*crypto.CredentialData, error) {
	if !s.isUnlocked(ctx, ownerID) {
		return nil, ErrVaultLocked
	}

	cred, err := s.getCredential(ctx, ownerID, credentialID)
	if err != nil {
		return nil, err
	}

	encrypted, err := base64.StdEncoding.DecodeString(cred.EncryptedData)
	if err != nil {
		return nil, fmt.Errorf("decrypt: invalid encrypted data: %w", err)
	}
	nonce, err := base64.StdEncoding.DecodeString(cred.Nonce)
	if err != nil {
		return nil, fmt.Errorf("decrypt: invalid nonce: %w", err)
	}

	return s.vault.DecryptCredential(encrypted, nonce)
}

// --- DB helpers ---

type keyVersionRow struct {
	ID        uuid.UUID
	Version   int
	Salt      []byte
	Params    []byte
	RotatedAt time.Time
	RotatedBy uuid.UUID
}

func (s *CredentialService) getLatestKeyVersion(ctx context.Context, userID uuid.UUID) (*keyVersionRow, error) {
	query := `
		SELECT id, version, salt, params, rotated_at, rotated_by
		FROM master_key_versions
		WHERE user_id = $1 ORDER BY version DESC LIMIT 1
	`
	var row keyVersionRow
	err := s.db.QueryRowContext(ctx, query, userID).Scan(
		&row.ID, &row.Version, &row.Salt, &row.Params, &row.RotatedAt, &row.RotatedBy,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &row, err
}

func (s *CredentialService) storeKeyVersion(ctx context.Context, userID uuid.UUID, kv *keyVersionRow) error {
	query := `
		INSERT INTO master_key_versions (id, user_id, version, salt, params, rotated_at, rotated_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (user_id, version) DO UPDATE SET
			salt = EXCLUDED.salt,
			params = EXCLUDED.params,
			rotated_at = NOW()
	`
	_, err := s.db.ExecContext(ctx, query, kv.ID, userID, kv.Version, kv.Salt, kv.Params, kv.RotatedAt, kv.RotatedBy)
	return err
}

func (s *CredentialService) getCredential(ctx context.Context, ownerID, credentialID uuid.UUID) (*Credential, error) {
	query := `
		SELECT id, owner_id, name, type, encrypted_data, nonce, version, tags,
		       last_rotated_at, expires_at, created_at, updated_at
		FROM credentials WHERE id = $1 AND owner_id = $2
	`
	var cred Credential
	var tags []byte
	var expiresAt, lastRotated sql.NullTime

	err := s.db.QueryRowContext(ctx, query, credentialID, ownerID).Scan(
		&cred.ID, &cred.OwnerID, &cred.Name, &cred.Type,
		&cred.EncryptedData, &cred.Nonce, &cred.Version,
		&tags, &lastRotated, &expiresAt, &cred.CreatedAt, &cred.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, ErrCredentialNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get: querying: %w", err)
	}

	cred.Tags = parseTextArray(tags)
	if lastRotated.Valid {
		cred.LastRotatedAt = &lastRotated.Time
	}
	if expiresAt.Valid {
		cred.ExpiresAt = &expiresAt.Time
	}
	return &cred, nil
}

func (s *CredentialService) listCredentials(ctx context.Context, ownerID uuid.UUID, tag string, limit, offset int) ([]*Credential, error) {
	if limit <= 0 {
		limit = 100
	}
	query := `
		SELECT id, owner_id, name, type, encrypted_data, nonce, version, tags,
		       last_rotated_at, expires_at, created_at, updated_at
		FROM credentials WHERE owner_id = $1
	`
	args := []interface{}{ownerID}
	idx := 2

	if tag != "" {
		query += fmt.Sprintf(" AND $%d = ANY(tags)", idx)
		args = append(args, tag)
		idx++
	}
	query += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d", idx)
	args = append(args, limit)
	idx++
	if offset > 0 {
		query += fmt.Sprintf(" OFFSET $%d", idx)
		args = append(args, offset)
	}

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list: querying: %w", err)
	}
	defer rows.Close()

	var creds []*Credential
	for rows.Next() {
		var cred Credential
		var tags []byte
		var expiresAt, lastRotated sql.NullTime

		err := rows.Scan(
			&cred.ID, &cred.OwnerID, &cred.Name, &cred.Type,
			&cred.EncryptedData, &cred.Nonce, &cred.Version,
			&tags, &lastRotated, &expiresAt, &cred.CreatedAt, &cred.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		cred.Tags = parseTextArray(tags)
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

func (s *CredentialService) isUnlocked(ctx context.Context, userID uuid.UUID) bool {
	raw, err := s.redis.Get(ctx, s.vaultRedisKey(userID)).Result()
	if err != nil {
		return false
	}
	var st vaultState
	if json.Unmarshal([]byte(raw), &st) != nil {
		return false
	}
	return st.Unlocked
}

// parseTextArray parses a PostgreSQL TEXT[] (e.g., {a,b,c}) into []string.
func parseTextArray(data []byte) []string {
	if data == nil || len(data) < 2 {
		return nil
	}
	s := string(data[1 : len(data)-1])
	if s == "" {
		return nil
	}
	var result []string
	inQuote := false
	current := ""
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c == '"' {
			inQuote = !inQuote
		} else if c == ',' && !inQuote {
			result = append(result, current)
			current = ""
		} else {
			current += string(c)
		}
	}
	if current != "" {
		result = append(result, current)
	}
	return result
}

// pqArray is a helper to use lib/pq array format for tags.
func pqArray(tags []string) interface{} {
	if len(tags) == 0 {
		return nil
	}
	// Format as PostgreSQL array literal: {tag1,tag2,tag3}
	out := "{"
	for i, t := range tags {
		if i > 0 {
			out += ","
		}
		// Escape double quotes in tags
		escaped := ""
		for _, c := range t {
			if c == '"' {
				escaped += `\"`
			} else {
				escaped += string(c)
			}
		}
		out += `"` + escaped + `"`
	}
	out += "}"
	return out
}

func encodeParams(p crypto.Argon2Params) []byte {
	return []byte(fmt.Sprintf(`{"memory":%d,"time":%d,"threads":%d,"saltLen":%d,"keyLen":%d}`,
		p.Memory, p.Time, p.Threads, p.SaltLen, p.KeyLen))
}

func decodeParams(data []byte) crypto.Argon2Params {
	var p crypto.Argon2Params
	fmt.Sscanf(string(data), `{"memory":%d,"time":%d,"threads":%d,"saltLen":%d,"keyLen":%d}`,
		&p.Memory, &p.Time, &p.Threads, &p.SaltLen, &p.KeyLen)
	if p.Memory == 0 {
		p = crypto.DefaultArgon2Params()
	}
	return p
}

// ShareWithTeam is a placeholder for team credential sharing.
func (s *CredentialService) ShareWithTeam(ctx context.Context, userID uuid.UUID, credID uuid.UUID, teamID uuid.UUID, permissions string) error {
	return fmt.Errorf("ShareWithTeam not yet implemented")
}

// RevokeTeamShare is a placeholder for revoking team credential share.
func (s *CredentialService) RevokeTeamShare(ctx context.Context, userID uuid.UUID, credID uuid.UUID, teamID uuid.UUID) error {
	return fmt.Errorf("RevokeTeamShare not yet implemented")
}

// CreateCredentialRequest is the request body for creating a credential.
type CreateCredentialRequest struct {
	Name   string    `json:"name" binding:"required,min=1,max=255"`
	Type   string    `json:"type" binding:"required,oneof=password ssh_key api_token"`
	Data   string    `json:"data" binding:"required"`
	Tags   []string  `json:"tags,omitempty"`
	Expiry *time.Time `json:"expires_at,omitempty"`
}

// UpdateCredentialRequest is the request body for updating a credential.
type UpdateCredentialRequest struct {
	Name   string    `json:"name,omitempty"`
	Data   string    `json:"data,omitempty"`
	Tags   []string  `json:"tags,omitempty"`
	Expiry *time.Time `json:"expires_at,omitempty"`
}
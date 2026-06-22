package vault

import (
	"context"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// ShareRepository handles database operations for credential sharing
type ShareRepository struct {
	db     *sql.DB
	crypto *SharingCrypto
}

func NewShareRepository(db *sql.DB) *ShareRepository {
	return &ShareRepository{
		db:     db,
		crypto: NewSharingCrypto(),
	}
}

// CreateShareRequest represents a request to share a credential
type CreateShareRequest struct {
	CredentialID string          `json:"credential_id"`
	RecipientID  string          `json:"recipient_id"`
	Permission   SharePermission `json:"permission"`
	ExpiryDays   *int            `json:"expiry_days,omitempty"` // nil = never
}

// Validate validates the share request
func (r *CreateShareRequest) Validate() error {
	if r.CredentialID == "" {
		return fmt.Errorf("credential_id is required")
	}
	if r.RecipientID == "" {
		return fmt.Errorf("recipient_id is required")
	}
	if err := r.Permission.Validate(); err != nil {
		return err
	}
	return nil
}

// CreateShare creates a new credential share
func (sr *ShareRepository) CreateShare(
	ctx context.Context,
	req *CreateShareRequest,
	senderID string,
	credentialData []byte,
	recipientPubKey []byte,
	senderPrivKey []byte,
) (*ShareInvitation, error) {
	// Validate request
	if err := req.Validate(); err != nil {
		return nil, err
	}

	// Check if sender owns the credential
	var ownerID string
	err := sr.db.QueryRowContext(ctx,
		"SELECT user_id FROM credentials WHERE id = $1 AND deleted_at IS NULL",
		req.CredentialID,
	).Scan(&ownerID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("credential not found")
		}
		return nil, fmt.Errorf("database error: %w", err)
	}
	if ownerID != senderID {
		return nil, fmt.Errorf("only credential owner can share")
	}

	// Check if share already exists
	var existingID string
	err = sr.db.QueryRowContext(ctx,
		"SELECT id FROM credential_shares WHERE credential_id = $1 AND recipient_id = $2 AND status = 'accepted' AND revoked_at IS NULL",
		req.CredentialID, req.RecipientID,
	).Scan(&existingID)
	if err == nil {
		return nil, fmt.Errorf("credential already shared with this user")
	}

	// Encrypt credential data for recipient
	encryptedShare, err := sr.crypto.EncryptCredential(
		credentialData,
		senderPrivKey,
		recipientPubKey,
	)
	if err != nil {
		return nil, fmt.Errorf("encryption failed: %w", err)
	}

	// Serialize encrypted data
	encryptedData, err := encryptedShare.ToJSON()
	if err != nil {
		return nil, fmt.Errorf("failed to serialize: %w", err)
	}

	// Calculate expiry
	var expiryTime *time.Time
	if req.ExpiryDays != nil && *req.ExpiryDays > 0 {
		t := time.Now().AddDate(0, 0, *req.ExpiryDays)
		expiryTime = &t
	}

	// Create share record
	shareID := uuid.New().String()
	now := time.Now()

	_, err = sr.db.ExecContext(ctx,
		`INSERT INTO credential_shares (
			id, credential_id, sender_id, recipient_id, 
			permission, encrypted_data, expiry_time, 
			status, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, 'pending', $8, $8)`,
		shareID, req.CredentialID, senderID, req.RecipientID,
		string(req.Permission), encryptedData, expiryTime, now,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create share: %w", err)
	}

	// Log audit event
	if err := sr.logAudit(ctx, shareID, "created", senderID, string(req.Permission), true, ""); err != nil {
		// Log but don't fail
		fmt.Printf("audit log warning: %v\n", err)
	}

	return &ShareInvitation{
		ID:            shareID,
		CredentialID:  req.CredentialID,
		SenderID:      senderID,
		RecipientID:   req.RecipientID,
		Permission:    req.Permission,
		EncryptedData: []byte(encryptedData),
		ExpiryTime:    expiryTime,
		CreatedAt:     now,
		Status:        "pending",
	}, nil
}

// AcceptShare accepts a share invitation
func (sr *ShareRepository) AcceptShare(ctx context.Context, shareID, recipientID string) (*ShareInvitation, error) {
	// Get share
	var share ShareInvitation
	var encryptedDataStr string
	err := sr.db.QueryRowContext(ctx,
		`SELECT id, credential_id, sender_id, recipient_id, 
			permission, encrypted_data, expiry_time, 
			status, created_at 
		FROM credential_shares 
		WHERE id = $1 AND recipient_id = $2`,
		shareID, recipientID,
	).Scan(
		&share.ID, &share.CredentialID, &share.SenderID, &share.RecipientID,
		&share.Permission, &encryptedDataStr, &share.ExpiryTime,
		&share.Status, &share.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("share not found or unauthorized")
		}
		return nil, fmt.Errorf("database error: %w", err)
	}

	// Validate status
	if share.Status != "pending" {
		return nil, fmt.Errorf("share is not pending (status: %s)", share.Status)
	}

	// Check expiry
	if share.IsExpired() {
		// Update status to expired
		_, _ = sr.db.ExecContext(ctx,
			"UPDATE credential_shares SET status = 'expired', updated_at = $1 WHERE id = $2",
			time.Now(), shareID,
		)
		return nil, fmt.Errorf("share has expired")
	}

	// Accept share
	now := time.Now()
	_, err = sr.db.ExecContext(ctx,
		`UPDATE credential_shares 
		SET status = 'accepted', accepted_at = $1, updated_at = $1 
		WHERE id = $2`,
		now, shareID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to accept share: %w", err)
	}

	share.Status = "accepted"
	share.AcceptedAt = &now

	// Log audit event
	if err := sr.logAudit(ctx, shareID, "accepted", recipientID, string(share.Permission), true, ""); err != nil {
		fmt.Printf("audit log warning: %v\n", err)
	}

	return &share, nil
}

// RevokeShare revokes access to a shared credential
func (sr *ShareRepository) RevokeShare(ctx context.Context, shareID, userID string) error {
	// Get share to check ownership
	var senderID, recipientID, credentialID string
	var status string
	err := sr.db.QueryRowContext(ctx,
		"SELECT sender_id, recipient_id, credential_id, status FROM credential_shares WHERE id = $1",
		shareID,
	).Scan(&senderID, &recipientID, &credentialID, &status)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("share not found")
		}
		return fmt.Errorf("database error: %w", err)
	}

	// Only sender or admin can revoke
	if senderID != userID {
		// Check if user is admin
		var userRole string
		err := sr.db.QueryRowContext(ctx,
			"SELECT role FROM users WHERE id = $1",
			userID,
		).Scan(&userRole)
		if err != nil || userRole != "admin" {
			return fmt.Errorf("only sender or admin can revoke")
		}
	}

	if status == "revoked" {
		return fmt.Errorf("share already revoked")
	}

	now := time.Now()
	_, err = sr.db.ExecContext(ctx,
		`UPDATE credential_shares 
		SET status = 'revoked', revoked_at = $1, updated_at = $1 
		WHERE id = $2`,
		now, shareID,
	)
	if err != nil {
		return fmt.Errorf("failed to revoke share: %w", err)
	}

	// Log audit event
	if err := sr.logAudit(ctx, shareID, "revoked", userID, "", true, ""); err != nil {
		fmt.Printf("audit log warning: %v\n", err)
	}

	// Re-encrypt remaining shares (async in production)
	go sr.reencryptRemainingShares(credentialID)

	return nil
}

// ListShares lists all shares for a user (either sent or received)
type ShareListFilter struct {
	UserID    string
	SentByMe  bool
	Status    string // all, pending, accepted, revoked, expired
}

func (sr *ShareRepository) ListShares(ctx context.Context, filter *ShareListFilter) ([]*ShareInvitation, error) {
	query := `SELECT id, credential_id, sender_id, recipient_id, 
		permission, encrypted_data, expiry_time, 
		status, created_at, accepted_at, revoked_at 
	FROM credential_shares `

	var args []interface{}
	var conditions []string

	if filter.SentByMe {
		conditions = append(conditions, "sender_id = $1")
		args = append(args, filter.UserID)
	} else {
		conditions = append(conditions, "recipient_id = $1")
		args = append(args, filter.UserID)
	}

	if filter.Status != "all" && filter.Status != "" {
		conditions = append(conditions, fmt.Sprintf("status = $%d", len(args)+1))
		args = append(args, filter.Status)
	}

	query += " WHERE " + conditions[0]
	for i := 1; i < len(conditions); i++ {
		query += " AND " + conditions[i]
	}
	query += " ORDER BY created_at DESC"

	rows, err := sr.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("database error: %w", err)
	}
	defer rows.Close()

	var shares []*ShareInvitation
	for rows.Next() {
		var s ShareInvitation
		var encryptedDataStr string
		err := rows.Scan(
			&s.ID, &s.CredentialID, &s.SenderID, &s.RecipientID,
			&s.Permission, &encryptedDataStr, &s.ExpiryTime,
			&s.Status, &s.CreatedAt, &s.AcceptedAt, &s.RevokedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan error: %w", err)
		}
		s.EncryptedData = []byte(encryptedDataStr)
		shares = append(shares, &s)
	}

	return shares, nil
}

// GetShare returns a single share by ID
func (sr *ShareRepository) GetShare(ctx context.Context, shareID, userID string) (*ShareInvitation, error) {
	var s ShareInvitation
	var encryptedDataStr string
	err := sr.db.QueryRowContext(ctx,
		`SELECT id, credential_id, sender_id, recipient_id, 
			permission, encrypted_data, expiry_time, 
			status, created_at, accepted_at, revoked_at 
		FROM credential_shares 
		WHERE id = $1 AND (sender_id = $2 OR recipient_id = $2)`,
		shareID, userID,
	).Scan(
		&s.ID, &s.CredentialID, &s.SenderID, &s.RecipientID,
		&s.Permission, &encryptedDataStr, &s.ExpiryTime,
		&s.Status, &s.CreatedAt, &s.AcceptedAt, &s.RevokedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("share not found or unauthorized")
		}
		return nil, fmt.Errorf("database error: %w", err)
	}
	s.EncryptedData = []byte(encryptedDataStr)

	// Check expiry
	if s.IsExpired() && s.Status != "expired" {
		_, _ = sr.db.ExecContext(ctx,
			"UPDATE credential_shares SET status = 'expired', updated_at = $1 WHERE id = $2",
			time.Now(), shareID,
		)
		s.Status = "expired"
	}

	return &s, nil
}

// DecryptSharedCredential decrypts credential data for the recipient
func (sr *ShareRepository) DecryptSharedCredential(
	encryptedData []byte,
	recipientPrivKey []byte,
) (*Credential, error) {
	// Parse encrypted share
	es, err := EncryptedShareFromJSON(string(encryptedData))
	if err != nil {
		return nil, fmt.Errorf("invalid encrypted data: %w", err)
	}

	// Decrypt
	plaintext, err := sr.crypto.DecryptCredential(es, recipientPrivKey)
	if err != nil {
		return nil, fmt.Errorf("decryption failed: %w", err)
	}

	// Deserialize credential
	var cred Credential
	if err := json.Unmarshal(plaintext, &cred); err != nil {
		return nil, fmt.Errorf("failed to deserialize credential: %w", err)
	}

	return &cred, nil
}

// logAudit creates an audit log entry for sharing events
func (sr *ShareRepository) logAudit(
	ctx context.Context,
	shareID, action, userID, permission string,
	success bool, errorMsg string,
) error {
	logID := uuid.New().String()
	now := time.Now()

	_, err := sr.db.ExecContext(ctx,
		`INSERT INTO share_audit_log (
			id, share_id, action, user_id, permission, 
			ip_address, user_agent, timestamp, success, error_message
		) VALUES ($1, $2, $3, $4, $5, NULL, NULL, $6, $7, $8)`,
		logID, shareID, action, userID, permission,
		now, success, errorMsg,
	)
	return err
}

// reencryptRemainingShares re-encrypts credential for remaining active shares
// This is run async after a revoke
func (sr *ShareRepository) reencryptRemainingShares(credentialID string) {
	// In production, this would:
	// 1. Fetch all active shares for this credential
	// 2. Decrypt credential with owner's key
	// 3. Re-encrypt for each remaining recipient
	// 4. Update share records
	// For now, just log
	fmt.Printf("re-encrypting shares for credential %s\n", credentialID)
}

// GetUserShareKeys gets or generates user's share key pair
func (sr *ShareRepository) GetUserShareKeys(ctx context.Context, userID string, vaultKey []byte) (*ShareKeyPair, error) {
	// Try to get existing keys
	var pubKeyStr, privKeyStr string
	err := sr.db.QueryRowContext(ctx,
		"SELECT public_key, private_key FROM user_share_keys WHERE user_id = $1",
		userID,
	).Scan(&pubKeyStr, &privKeyStr)
	
	if err == sql.ErrNoRows {
		// Generate new keys
		keyPair, err := DeriveShareKey(vaultKey, userID)
		if err != nil {
			// Fallback to random key pair
			keyPair, err = sr.crypto.GenerateShareKeyPair()
			if err != nil {
				return nil, fmt.Errorf("failed to generate keys: %w", err)
			}
		}

		// Store keys (in production, private key should be encrypted)
		_, err = sr.db.ExecContext(ctx,
			`INSERT INTO user_share_keys (user_id, public_key, private_key, created_at) 
			VALUES ($1, $2, $3, $4)`,
			userID, base64.StdEncoding.EncodeToString(keyPair.PublicKey),
			base64.StdEncoding.EncodeToString(keyPair.PrivateKey), time.Now(),
		)
		if err != nil {
			return nil, fmt.Errorf("failed to store keys: %w", err)
		}

		return keyPair, nil
	}
	if err != nil {
		return nil, fmt.Errorf("database error: %w", err)
	}

	// Decode existing keys
	pubKey, err := base64.StdEncoding.DecodeString(pubKeyStr)
	if err != nil {
		return nil, fmt.Errorf("invalid public key: %w", err)
	}
	privKey, err := base64.StdEncoding.DecodeString(privKeyStr)
	if err != nil {
		return nil, fmt.Errorf("invalid private key: %w", err)
	}

	return &ShareKeyPair{
		PublicKey:  pubKey,
		PrivateKey: privKey,
	}, nil
}

package auth

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/google/uuid"
)

var (
	ErrWebAuthnCredentialNotFound = errors.New("webauthn credential not found")
	ErrWebAuthnSessionExpired     = errors.New("webauthn session expired")
	ErrWebAuthnInvalidChallenge   = errors.New("invalid challenge")
	ErrWebAuthnDuplicateCredential = errors.New("credential already registered")
)

// WebAuthnUser implements webauthn.User for library integration.
type WebAuthnUser struct {
	ID          uuid.UUID
	Email       string
	DisplayName string
	Credentials []WebAuthnCredential
}

func (u *WebAuthnUser) WebAuthnID() []byte {
	return []byte(u.ID.String())
}

func (u *WebAuthnUser) WebAuthnName() string {
	return u.Email
}

func (u *WebAuthnUser) WebAuthnDisplayName() string {
	if u.DisplayName != "" {
		return u.DisplayName
	}
	return u.Email
}

func (u *WebAuthnUser) WebAuthnIcon() string {
	return ""
}

func (u *WebAuthnUser) WebAuthnCredentials() []webauthn.Credential {
	creds := make([]webauthn.Credential, 0, len(u.Credentials))
	for _, c := range u.Credentials {
		creds = append(creds, *c.ToWebAuthn())
	}
	return creds
}

// WebAuthnCredential represents a stored FIDO2 credential.
type WebAuthnCredential struct {
	ID                      uuid.UUID
	UserID                  uuid.UUID
	CredentialID            []byte
	PublicKey               []byte
	AttestationType         string
	Transport               []string
	AuthenticatorAAGUID     []byte
	AuthenticatorSignCount  uint32
	AuthenticatorCloneWarning bool
	AuthenticatorAttachment string
	IsResidentKey           bool
	IsBackupEligible        bool
	IsBackedUp              bool
	UserVerified            bool
	Name                    string
	LastUsedAt              *time.Time
	CreatedAt               time.Time
	UpdatedAt               time.Time
}

func (c *WebAuthnCredential) ToWebAuthn() *webauthn.Credential {
	transports := make([]protocol.AuthenticatorTransport, 0, len(c.Transport))
	for _, t := range c.Transport {
		transports = append(transports, protocol.AuthenticatorTransport(t))
	}

	return &webauthn.Credential{
		ID:        c.CredentialID,
		PublicKey: c.PublicKey,
		Flags: webauthn.CredentialFlags{
			UserPresent:    true,
			UserVerified:   c.UserVerified,
			BackupEligible: c.IsBackupEligible,
			BackupState:    c.IsBackedUp,
		},
		Authenticator: webauthn.Authenticator{
			AAGUID:       c.AuthenticatorAAGUID,
			SignCount:    c.AuthenticatorSignCount,
			CloneWarning: c.AuthenticatorCloneWarning,
			Attachment:   protocol.AuthenticatorAttachment(c.AuthenticatorAttachment),
		},
		AttestationType: c.AttestationType,
		Transport:       transports,
	}
}

// WebAuthnService wraps the go-webauthn library and database storage.
type WebAuthnService struct {
	db       *sql.DB
	web      *webauthn.WebAuthn
}

// NewWebAuthnService creates a new WebAuthnService.
func NewWebAuthnService(db *sql.DB, rpID, rpOrigin, rpDisplayName string) (*WebAuthnService, error) {
	wconfig := &webauthn.Config{
		RPDisplayName:         rpDisplayName,
		RPID:                  rpID,
		RPOrigins:             []string{rpOrigin},
		AttestationPreference: protocol.PreferDirectAttestation,
		AuthenticatorSelection: protocol.AuthenticatorSelection{
			AuthenticatorAttachment: protocol.CrossPlatform,
			ResidentKey:             protocol.ResidentKeyRequirementPreferred,
			UserVerification:        protocol.VerificationRequired,
		},
		Debug: false,
	}
	w, err := webauthn.New(wconfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create webauthn instance: %w", err)
	}

	return &WebAuthnService{
		db:  db,
		web: w,
	}, nil
}

// WebAuthn returns the underlying webauthn instance (for handlers that need raw access).
func (s *WebAuthnService) WebAuthn() *webauthn.WebAuthn {
	return s.web
}

// GetUser returns a WebAuthnUser populated with stored credentials.
func (s *WebAuthnService) GetUser(ctx context.Context, userID uuid.UUID) (*WebAuthnUser, error) {
	var email, displayName string
	err := s.db.QueryRowContext(ctx, "SELECT email, email FROM users WHERE id = $1", userID).Scan(&email, &displayName)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	creds, err := s.GetCredentialsByUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	return &WebAuthnUser{
		ID:          userID,
		Email:       email,
		DisplayName: displayName,
		Credentials: creds,
	}, nil
}

// GetCredentialsByUser returns all WebAuthn credentials for a user.
func (s *WebAuthnService) GetCredentialsByUser(ctx context.Context, userID uuid.UUID) ([]WebAuthnCredential, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, user_id, credential_id, public_key, attestation_type, transport,
		       authenticator_aaguid, authenticator_sign_count, authenticator_clone_warning,
		       authenticator_attachment, is_resident_key, is_backup_eligible, is_backed_up,
		       user_verified, name, last_used_at, created_at, updated_at
		FROM webauthn_credentials WHERE user_id = $1 ORDER BY created_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var creds []WebAuthnCredential
	for rows.Next() {
		var c WebAuthnCredential
		var transportRaw []byte
		err := rows.Scan(
			&c.ID, &c.UserID, &c.CredentialID, &c.PublicKey, &c.AttestationType,
			&transportRaw, &c.AuthenticatorAAGUID, &c.AuthenticatorSignCount,
			&c.AuthenticatorCloneWarning, &c.AuthenticatorAttachment,
			&c.IsResidentKey, &c.IsBackupEligible, &c.IsBackedUp,
			&c.UserVerified, &c.Name, &c.LastUsedAt, &c.CreatedAt, &c.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		if len(transportRaw) > 0 {
			_ = json.Unmarshal(transportRaw, &c.Transport)
		}
		creds = append(creds, c)
	}
	return creds, rows.Err()
}

// BeginRegistration creates a registration challenge and stores the session.
func (s *WebAuthnService) BeginRegistration(ctx context.Context, user *WebAuthnUser, opts ...webauthn.RegistrationOption) (*protocol.CredentialCreation, []byte, error) {
	options, session, err := s.web.BeginRegistration(user, opts...)
	if err != nil {
		return nil, nil, fmt.Errorf("begin registration failed: %w", err)
	}

	sessionData, err := json.Marshal(session)
	if err != nil {
		return nil, nil, err
	}

	challenge := []byte(session.Challenge)

	_, err = s.db.ExecContext(ctx, `
		INSERT INTO webauthn_registration_sessions (user_id, challenge, session_data, expires_at)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (user_id) DO UPDATE SET
			challenge = EXCLUDED.challenge,
			session_data = EXCLUDED.session_data,
			expires_at = EXCLUDED.expires_at,
			created_at = NOW()
	`, user.ID, challenge, sessionData, time.Now().UTC().Add(5*time.Minute))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to store registration session: %w", err)
	}

	return options, challenge, nil
}

// FinishRegistration validates the registration response and stores the credential.
func (s *WebAuthnService) FinishRegistration(ctx context.Context, user *WebAuthnUser, req *protocol.ParsedCredentialCreationData, credentialName string) (*WebAuthnCredential, error) {
	var sessionData []byte
	var challenge []byte
	err := s.db.QueryRowContext(ctx, `
		SELECT session_data, challenge FROM webauthn_registration_sessions
		WHERE user_id = $1 AND expires_at > NOW()
	`, user.ID).Scan(&sessionData, &challenge)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrWebAuthnSessionExpired
		}
		return nil, err
	}

	var session webauthn.SessionData
	if err := json.Unmarshal(sessionData, &session); err != nil {
		return nil, err
	}

	credential, err := s.web.CreateCredential(user, session, req)
	if err != nil {
		return nil, fmt.Errorf("credential creation failed: %w", err)
	}

	// Check for duplicate credential ID
	var existing uuid.UUID
	dupeErr := s.db.QueryRowContext(ctx,
		"SELECT id FROM webauthn_credentials WHERE credential_id = $1",
		credential.ID,
	).Scan(&existing)
	if dupeErr != sql.ErrNoRows {
		if dupeErr == nil {
			return nil, ErrWebAuthnDuplicateCredential
		}
		return nil, dupeErr
	}

	transportJSON, _ := json.Marshal(credential.Transport)
	if credentialName == "" {
		credentialName = "Unnamed Credential"
	}

	cred := &WebAuthnCredential{
		ID:                      uuid.New(),
		UserID:                  user.ID,
		CredentialID:            credential.ID,
		PublicKey:               credential.PublicKey,
		AttestationType:         credential.AttestationType,
		Transport:               []string{},
		AuthenticatorAAGUID:     credential.Authenticator.AAGUID,
		AuthenticatorSignCount:  credential.Authenticator.SignCount,
		AuthenticatorCloneWarning: credential.Authenticator.CloneWarning,
		AuthenticatorAttachment: string(credential.Authenticator.Attachment),
		IsResidentKey:           false, // resident key info not available in this API version
		IsBackupEligible:        credential.Flags.BackupEligible,
		IsBackedUp:              credential.Flags.BackupState,
		UserVerified:            credential.Flags.UserVerified,
		Name:                    credentialName,
		CreatedAt:               time.Now().UTC(),
		UpdatedAt:               time.Now().UTC(),
	}

	if len(transportJSON) > 0 {
		_ = json.Unmarshal(transportJSON, &cred.Transport)
	}

	_, err = s.db.ExecContext(ctx, `
		INSERT INTO webauthn_credentials (
			id, user_id, credential_id, public_key, attestation_type, transport,
			authenticator_aaguid, authenticator_sign_count, authenticator_clone_warning,
			authenticator_attachment, is_resident_key, is_backup_eligible, is_backed_up,
			user_verified, name, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)
	`,
		cred.ID, cred.UserID, cred.CredentialID, cred.PublicKey, cred.AttestationType,
		transportJSON, cred.AuthenticatorAAGUID, cred.AuthenticatorSignCount,
		cred.AuthenticatorCloneWarning, cred.AuthenticatorAttachment,
		cred.IsResidentKey, cred.IsBackupEligible, cred.IsBackedUp,
		cred.UserVerified, cred.Name, cred.CreatedAt, cred.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to store credential: %w", err)
	}

	// Clean up session
	_, _ = s.db.ExecContext(ctx,
		"DELETE FROM webauthn_registration_sessions WHERE user_id = $1", user.ID,
	)

	return cred, nil
}

// BeginLogin starts an authentication ceremony.
func (s *WebAuthnService) BeginLogin(ctx context.Context, user *WebAuthnUser, opts ...webauthn.LoginOption) (*protocol.CredentialAssertion, []byte, error) {
	options, session, err := s.web.BeginLogin(user, opts...)
	if err != nil {
		return nil, nil, fmt.Errorf("begin login failed: %w", err)
	}

	sessionData, err := json.Marshal(session)
	if err != nil {
		return nil, nil, err
	}

	challenge := []byte(session.Challenge)
	var allowedIDs [][]byte
	for _, c := range user.Credentials {
		allowedIDs = append(allowedIDs, c.CredentialID)
	}

	_, err = s.db.ExecContext(ctx, `
		INSERT INTO webauthn_login_sessions (challenge, session_data, allowed_credential_ids, expires_at)
		VALUES ($1, $2, $3, $4)
	`, challenge, sessionData, pqArrayBytes(allowedIDs), time.Now().UTC().Add(5*time.Minute))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to store login session: %w", err)
	}

	return options, challenge, nil
}

// FinishLogin validates an authentication response and returns the matched credential.
func (s *WebAuthnService) FinishLogin(ctx context.Context, user *WebAuthnUser, req *protocol.ParsedCredentialAssertionData) (*WebAuthnCredential, error) {
	var sessionData []byte
	err := s.db.QueryRowContext(ctx, `
		SELECT session_data FROM webauthn_login_sessions
		WHERE challenge = $1 AND expires_at > NOW()
	`, req.Raw.AssertionResponse.ClientDataJSON).Scan(&sessionData)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrWebAuthnSessionExpired
		}
		return nil, err
	}

	var session webauthn.SessionData
	if err := json.Unmarshal(sessionData, &session); err != nil {
		return nil, err
	}

	credential, err := s.web.ValidateLogin(user, session, req)
	if err != nil {
		return nil, fmt.Errorf("login validation failed: %w", err)
	}

	// Update sign count, last used, backup state
	_, err = s.db.ExecContext(ctx, `
		UPDATE webauthn_credentials
		SET authenticator_sign_count = $1,
		    is_backed_up = $2,
		    last_used_at = NOW(),
		    updated_at = NOW()
		WHERE credential_id = $3
	`, credential.Authenticator.SignCount, credential.Flags.BackupState, credential.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to update credential: %w", err)
	}

	// Fetch updated credential
	var cred WebAuthnCredential
	var transportRaw []byte
	err = s.db.QueryRowContext(ctx, `
		SELECT id, user_id, credential_id, public_key, attestation_type, transport,
		       authenticator_aaguid, authenticator_sign_count, authenticator_clone_warning,
		       authenticator_attachment, is_resident_key, is_backup_eligible, is_backed_up,
		       user_verified, name, last_used_at, created_at, updated_at
		FROM webauthn_credentials WHERE credential_id = $1
	`, credential.ID).Scan(
		&cred.ID, &cred.UserID, &cred.CredentialID, &cred.PublicKey, &cred.AttestationType,
		&transportRaw, &cred.AuthenticatorAAGUID, &cred.AuthenticatorSignCount,
		&cred.AuthenticatorCloneWarning, &cred.AuthenticatorAttachment,
		&cred.IsResidentKey, &cred.IsBackupEligible, &cred.IsBackedUp,
		&cred.UserVerified, &cred.Name, &cred.LastUsedAt, &cred.CreatedAt, &cred.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	if len(transportRaw) > 0 {
		_ = json.Unmarshal(transportRaw, &cred.Transport)
	}

	// Clean up session
	_, _ = s.db.ExecContext(ctx,
		"DELETE FROM webauthn_login_sessions WHERE challenge = $1", session.Challenge,
	)

	return &cred, nil
}

// DeleteCredential removes a credential by ID, ensuring it belongs to the user.
func (s *WebAuthnService) DeleteCredential(ctx context.Context, userID, credentialID uuid.UUID) error {
	res, err := s.db.ExecContext(ctx,
		"DELETE FROM webauthn_credentials WHERE id = $1 AND user_id = $2",
		credentialID, userID,
	)
	if err != nil {
		return err
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return ErrWebAuthnCredentialNotFound
	}
	return nil
}

// GetCredentialByID returns a single credential.
func (s *WebAuthnService) GetCredentialByID(ctx context.Context, credentialID uuid.UUID) (*WebAuthnCredential, error) {
	var c WebAuthnCredential
	var transportRaw []byte
	err := s.db.QueryRowContext(ctx, `
		SELECT id, user_id, credential_id, public_key, attestation_type, transport,
		       authenticator_aaguid, authenticator_sign_count, authenticator_clone_warning,
		       authenticator_attachment, is_resident_key, is_backup_eligible, is_backed_up,
		       user_verified, name, last_used_at, created_at, updated_at
		FROM webauthn_credentials WHERE id = $1
	`, credentialID).Scan(
		&c.ID, &c.UserID, &c.CredentialID, &c.PublicKey, &c.AttestationType,
		&transportRaw, &c.AuthenticatorAAGUID, &c.AuthenticatorSignCount,
		&c.AuthenticatorCloneWarning, &c.AuthenticatorAttachment,
		&c.IsResidentKey, &c.IsBackupEligible, &c.IsBackedUp,
		&c.UserVerified, &c.Name, &c.LastUsedAt, &c.CreatedAt, &c.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrWebAuthnCredentialNotFound
		}
		return nil, err
	}
	if len(transportRaw) > 0 {
		_ = json.Unmarshal(transportRaw, &c.Transport)
	}
	return &c, nil
}

// GetCredentialByRawID looks up a credential by its raw credential ID bytes.
func (s *WebAuthnService) GetCredentialByRawID(ctx context.Context, rawID []byte) (*WebAuthnCredential, error) {
	var c WebAuthnCredential
	var transportRaw []byte
	err := s.db.QueryRowContext(ctx, `
		SELECT id, user_id, credential_id, public_key, attestation_type, transport,
		       authenticator_aaguid, authenticator_sign_count, authenticator_clone_warning,
		       authenticator_attachment, is_resident_key, is_backup_eligible, is_backed_up,
		       user_verified, name, last_used_at, created_at, updated_at
		FROM webauthn_credentials WHERE credential_id = $1
	`, rawID).Scan(
		&c.ID, &c.UserID, &c.CredentialID, &c.PublicKey, &c.AttestationType,
		&transportRaw, &c.AuthenticatorAAGUID, &c.AuthenticatorSignCount,
		&c.AuthenticatorCloneWarning, &c.AuthenticatorAttachment,
		&c.IsResidentKey, &c.IsBackupEligible, &c.IsBackedUp,
		&c.UserVerified, &c.Name, &c.LastUsedAt, &c.CreatedAt, &c.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrWebAuthnCredentialNotFound
		}
		return nil, err
	}
	if len(transportRaw) > 0 {
		_ = json.Unmarshal(transportRaw, &c.Transport)
	}
	return &c, nil
}

// UpdateCredentialName renames a credential.
func (s *WebAuthnService) UpdateCredentialName(ctx context.Context, userID, credentialID uuid.UUID, name string) error {
	res, err := s.db.ExecContext(ctx,
		"UPDATE webauthn_credentials SET name = $1, updated_at = NOW() WHERE id = $2 AND user_id = $3",
		name, credentialID, userID,
	)
	if err != nil {
		return err
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return ErrWebAuthnCredentialNotFound
	}
	return nil
}

// CleanupExpiredSessions removes stale challenge sessions.
func (s *WebAuthnService) CleanupExpiredSessions(ctx context.Context) error {
	_, err := s.db.ExecContext(ctx, `
		DELETE FROM webauthn_registration_sessions WHERE expires_at <= NOW();
		DELETE FROM webauthn_login_sessions WHERE expires_at <= NOW();
	`)
	return err
}

// GenerateChallenge creates a random challenge for client-side use.
func GenerateChallenge() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// pqArrayBytes converts [][]byte to a driver.Value for PostgreSQL BYTEA arrays.
func pqArrayBytes(data [][]byte) interface{} {
	if len(data) == 0 {
		return nil
	}
	return data
}

// RegistrationOptions returns options configured for user-verified registration
// supporting both platform and roaming authenticators with resident key preference.
func (s *WebAuthnService) RegistrationOptions() []webauthn.RegistrationOption {
	return []webauthn.RegistrationOption{
		func(c *protocol.PublicKeyCredentialCreationOptions) {
			c.AuthenticatorSelection = protocol.AuthenticatorSelection{
				AuthenticatorAttachment: "", // no preference — empty string per spec
				ResidentKey:             protocol.ResidentKeyRequirementPreferred,
				UserVerification:        protocol.VerificationRequired,
			}
			c.Attestation = protocol.PreferDirectAttestation
		},
	}
}

// LoginOptions returns options configured for user-verified login (preferred).
func (s *WebAuthnService) LoginOptions() []webauthn.LoginOption {
	return []webauthn.LoginOption{
		func(c *protocol.PublicKeyCredentialRequestOptions) {
			c.UserVerification = protocol.VerificationPreferred
		},
	}
}

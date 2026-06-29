package tests

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/soumabali/vexa/internal/vault"
)

func TestVaultSharing(t *testing.T) {
	// Exercise the CreateShareRequest contract and SharePermission constants.
	days := 30
	req := &vault.CreateShareRequest{
		CredentialID: uuid.New().String(),
		RecipientID:  uuid.New().String(),
		Permission:   vault.PermissionReadOnly,
		ExpiryDays:   &days,
	}
	require.NotNil(t, req)
	require.NoError(t, req.Validate())

	// Permission constants must be stable strings (consumed by clients).
	assert.Equal(t, "read_only", string(vault.PermissionReadOnly))
	assert.Equal(t, "read_write", string(vault.PermissionReadWrite))
	assert.Equal(t, "admin", string(vault.PermissionAdmin))

	// Filter shape: list-shares call sites rely on UserID + SentByMe.
	filter := &vault.ShareListFilter{
		UserID:   uuid.New().String(),
		SentByMe: true,
		Status:   "pending",
	}
	assert.Equal(t, "pending", filter.Status)
	assert.True(t, filter.SentByMe)
}

func TestCreateShareRequest_Validation(t *testing.T) {
	// Empty credential_id must fail validation.
	req := &vault.CreateShareRequest{}
	assert.Error(t, req.Validate())

	// Missing recipient_id must fail.
	req = &vault.CreateShareRequest{
		CredentialID: uuid.New().String(),
		Permission:   vault.PermissionReadWrite,
	}
	assert.Error(t, req.Validate())

	// Valid request succeeds regardless of expiry days.
	days := 30
	req = &vault.CreateShareRequest{
		CredentialID: uuid.New().String(),
		RecipientID:  uuid.New().String(),
		Permission:   vault.PermissionReadWrite,
		ExpiryDays:   &days,
	}
	assert.NoError(t, req.Validate())
}

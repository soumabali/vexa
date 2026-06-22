package wireguard

import (
	"context"
	"database/sql"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/soumabali/vexa/internal/models"
)

type tunnelRow struct {
	id            uuid.UUID
	userID        uuid.UUID
	hostID        uuid.UUID
	ifaceName     string
	privKey       string
	pubKey        string
	clientPubKey  string
	psk           interface{}
	serverIP      string
	clientIP      string
	port          int
	allowedIPs    interface{}
	dnsServers    interface{}
	mtu           int
	keepalive     int
	status        string
	enabled       bool
	lastHandshake interface{}
	bytesSent     int64
	bytesReceived int64
	lastRotated   interface{}
	createdAt     time.Time
	updatedAt     time.Time
}

func newTunnelRow() tunnelRow {
	return tunnelRow{
		id:            uuid.New(),
		userID:        uuid.New(),
		hostID:        uuid.New(),
		ifaceName:     "wg0",
		privKey:       "priv",
		pubKey:        "pub",
		clientPubKey:  "cpub",
		psk:           nil,
		serverIP:      "10.200.200.1/24",
		clientIP:      "10.200.200.2/24",
		port:          51820,
		allowedIPs:    []byte("{0.0.0.0/0}"),
		dnsServers:    []byte("{}"),
		mtu:           1420,
		keepalive:     25,
		status:        "connected",
		enabled:       true,
		lastHandshake: nil,
		bytesSent:     100,
		bytesReceived: 200,
		lastRotated:   nil,
		createdAt:     time.Now(),
		updatedAt:     time.Now(),
	}
}

func (r tunnelRow) addRow(rows *sqlmock.Rows) *sqlmock.Rows {
	return rows.AddRow(
		r.id, r.userID, r.hostID, r.ifaceName,
		r.privKey, r.pubKey, r.clientPubKey, r.psk,
		r.serverIP, r.clientIP, r.port,
		r.allowedIPs, r.dnsServers,
		r.mtu, r.keepalive,
		r.status, r.enabled,
		r.lastHandshake, r.bytesSent, r.bytesReceived,
		r.lastRotated, r.createdAt, r.updatedAt,
	)
}

func tunnelColumns() []string {
	return []string{
		"id", "user_id", "host_id", "interface_name",
		"server_private_key", "server_public_key", "client_public_key",
		"preshared_key", "server_ip", "client_ip", "listen_port",
		"allowed_ips", "dns_servers", "mtu", "persistent_keepalive",
		"status", "is_enabled", "last_handshake_at", "bytes_sent",
		"bytes_received", "last_rotated_at", "created_at", "updated_at",
	}
}

func TestNewRepository(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewRepository(db)
	assert.NotNil(t, repo)
}

func TestRepository_CountByUser(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewRepository(db)
	userID := uuid.New()

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM wireguard_tunnels WHERE user_id = \$1`).
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(3))

	count, err := repo.CountByUser(context.Background(), userID)
	require.NoError(t, err)
	assert.Equal(t, 3, count)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRepository_CountTotal(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewRepository(db)

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM wireguard_tunnels`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(5))

	count, err := repo.CountTotal(context.Background())
	require.NoError(t, err)
	assert.Equal(t, 5, count)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRepository_Create(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewRepository(db)
	tunnel := &models.WireGuardTunnel{
		UserID:              uuid.New(),
		HostID:              uuid.New(),
		InterfaceName:       "wg0",
		ServerPrivateKey:    "privKey",
		ServerPublicKey:     "pubKey",
		ClientPublicKey:     "clientPubKey",
		ServerIP:            "10.200.200.1/24",
		ClientIP:            "10.200.200.2/24",
		ListenPort:          51820,
		AllowedIPs:          []string{"0.0.0.0/0"},
		Status:              models.TunnelStatusCreating,
		IsEnabled:           false,
		MTU:                 1420,
		PersistentKeepalive: 25,
	}

	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO wireguard_tunnels`)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at"}).
			AddRow(uuid.New(), tunnel.CreatedAt, tunnel.UpdatedAt))

	err = repo.Create(context.Background(), tunnel)
	require.NoError(t, err)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRepository_GetByID_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewRepository(db)
	id := uuid.New()

	mock.ExpectQuery(`SELECT .+ FROM wireguard_tunnels WHERE id = \$1`).
		WithArgs(id).
		WillReturnError(sql.ErrNoRows)

	_, err = repo.GetByID(context.Background(), id)
	assert.ErrorIs(t, err, ErrTunnelNotFound)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRepository_Delete(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewRepository(db)
	id := uuid.New()

	mock.ExpectExec(`DELETE FROM wireguard_tunnels WHERE id = \$1`).
		WithArgs(id).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err = repo.Delete(context.Background(), id)
	require.NoError(t, err)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRepository_Delete_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewRepository(db)
	id := uuid.New()

	mock.ExpectExec(`DELETE FROM wireguard_tunnels WHERE id = \$1`).
		WithArgs(id).
		WillReturnResult(sqlmock.NewResult(0, 0))

	err = repo.Delete(context.Background(), id)
	assert.ErrorIs(t, err, ErrTunnelNotFound)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRepository_Update(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewRepository(db)
	id := uuid.New()

	mock.ExpectExec(`UPDATE wireguard_tunnels SET`).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err = repo.Update(context.Background(), id, map[string]interface{}{
		"status": models.TunnelStatusConnected,
	})
	require.NoError(t, err)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRepository_Update_NoUpdates(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewRepository(db)
	err = repo.Update(context.Background(), uuid.New(), map[string]interface{}{})
	assert.Error(t, err)
}

func TestRepository_GetHostAddress(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewRepository(db)
	hostID := uuid.New()

	mock.ExpectQuery(`SELECT address FROM hosts WHERE id = \$1`).
		WithArgs(hostID).
		WillReturnRows(sqlmock.NewRows([]string{"address"}).AddRow("192.168.1.1"))

	addr, err := repo.GetHostAddress(context.Background(), hostID)
	require.NoError(t, err)
	assert.Equal(t, "192.168.1.1", addr)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRepository_ListByUser(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewRepository(db)
	userID := uuid.New()

	rows := sqlmock.NewRows([]string{
		"id", "user_id", "host_id", "interface_name",
		"server_private_key", "server_public_key", "client_public_key",
		"preshared_key", "server_ip", "client_ip", "listen_port",
		"allowed_ips", "dns_servers", "mtu", "persistent_keepalive",
		"status", "is_enabled", "last_handshake_at", "bytes_sent",
		"bytes_received", "last_rotated_at", "created_at", "updated_at",
	}).
		AddRow(uuid.New(), userID, uuid.New(), "wg0",
			"priv", "pub", "cpub", nil, "10.200.200.1/24", "10.200.200.2/24",
			51820, []byte("{0.0.0.0/0}"), []byte("{}"), 1420, 25,
			"connected", true, nil, int64(100), int64(200),
			nil, time.Now(), time.Now())

	mock.ExpectQuery(`SELECT .+ FROM wireguard_tunnels WHERE user_id = \$1 ORDER BY created_at DESC`).
		WithArgs(userID).
		WillReturnRows(rows)

	tunnels, err := repo.ListByUser(context.Background(), userID, "", nil)
	require.NoError(t, err)
	assert.Len(t, tunnels, 1)
	assert.Equal(t, "wg0", tunnels[0].InterfaceName)
	assert.Equal(t, models.TunnelStatus("connected"), tunnels[0].Status)
	assert.Equal(t, int64(100), tunnels[0].BytesSent)
	assert.Equal(t, int64(200), tunnels[0].BytesReceived)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRepository_GetTunnelsDueForRotation(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewRepository(db)
	tr := newTunnelRow()

	rows := tr.addRow(sqlmock.NewRows(tunnelColumns()))

	mock.ExpectQuery(`SELECT .+ FROM wireguard_tunnels WHERE is_enabled = true`).
		WithArgs("90").
		WillReturnRows(rows)

	tunnels, err := repo.GetTunnelsDueForRotation(context.Background(), 90)
	require.NoError(t, err)
	assert.Len(t, tunnels, 1)
	assert.Equal(t, tr.id, tunnels[0].ID)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRepository_GetTunnelsDueForRotation_Empty(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewRepository(db)

	rows := sqlmock.NewRows(tunnelColumns())

	mock.ExpectQuery(`SELECT .+ FROM wireguard_tunnels WHERE is_enabled = true`).
		WithArgs("90").
		WillReturnRows(rows)

	tunnels, err := repo.GetTunnelsDueForRotation(context.Background(), 90)
	require.NoError(t, err)
	assert.Len(t, tunnels, 0)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRepository_GetHostAddress_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewRepository(db)
	hostID := uuid.New()

	mock.ExpectQuery(`SELECT address FROM hosts WHERE id = \$1`).
		WithArgs(hostID).
		WillReturnError(sql.ErrNoRows)

	_, err = repo.GetHostAddress(context.Background(), hostID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "host not found")

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRepository_Update_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewRepository(db)
	id := uuid.New()

	mock.ExpectExec(`UPDATE wireguard_tunnels SET`).
		WillReturnResult(sqlmock.NewResult(0, 0))

	err = repo.Update(context.Background(), id, map[string]interface{}{
		"status": models.TunnelStatusConnected,
	})
	assert.ErrorIs(t, err, ErrTunnelNotFound)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRepository_ListByUser_WithHostID(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewRepository(db)
	userID := uuid.New()
	tr := newTunnelRow()
	tr.userID = userID

	rows := tr.addRow(sqlmock.NewRows(tunnelColumns()))

	mock.ExpectQuery(`SELECT .+ FROM wireguard_tunnels WHERE user_id = \$1 AND host_id = \$2 ORDER BY created_at DESC`).
		WithArgs(userID, tr.hostID.String()).
		WillReturnRows(rows)

	tunnels, err := repo.ListByUser(context.Background(), userID, tr.hostID.String(), nil)
	require.NoError(t, err)
	assert.Len(t, tunnels, 1)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRepository_ListByUser_WithEnabled(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewRepository(db)
	userID := uuid.New()
	tr := newTunnelRow()
	tr.userID = userID
	enabled := true

	rows := tr.addRow(sqlmock.NewRows(tunnelColumns()))

	mock.ExpectQuery(`SELECT .+ FROM wireguard_tunnels WHERE user_id = \$1 AND is_enabled = \$2 ORDER BY created_at DESC`).
		WithArgs(userID, enabled).
		WillReturnRows(rows)

	tunnels, err := repo.ListByUser(context.Background(), userID, "", &enabled)
	require.NoError(t, err)
	assert.Len(t, tunnels, 1)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRepository_ListByUser_WithBoth(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewRepository(db)
	userID := uuid.New()
	tr := newTunnelRow()
	tr.userID = userID
	enabled := false

	rows := tr.addRow(sqlmock.NewRows(tunnelColumns()))

	mock.ExpectQuery(`SELECT .+ FROM wireguard_tunnels WHERE user_id = \$1 AND host_id = \$2 AND is_enabled = \$3 ORDER BY created_at DESC`).
		WithArgs(userID, tr.hostID.String(), enabled).
		WillReturnRows(rows)

	tunnels, err := repo.ListByUser(context.Background(), userID, tr.hostID.String(), &enabled)
	require.NoError(t, err)
	assert.Len(t, tunnels, 1)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRepository_ScanTunnel_WithPSK(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewRepository(db)
	tr := newTunnelRow()
	psk := "test-psk-value"
	tr.psk = &psk

	now := time.Now()
	tr.lastRotated = time.Now().UTC().Format(time.RFC3339)

	rows := tr.addRow(sqlmock.NewRows(tunnelColumns()))

	mock.ExpectQuery(`SELECT .+ FROM wireguard_tunnels WHERE id = \$1`).
		WithArgs(tr.id).
		WillReturnRows(rows)

	tunnel, err := repo.GetByID(context.Background(), tr.id)
	require.NoError(t, err)
	require.NotNil(t, tunnel.PresharedKey)
	assert.Equal(t, psk, *tunnel.PresharedKey)
	assert.NotNil(t, tunnel.LastRotatedAt)
	_ = now

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRepository_ListAll(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewRepository(db)

	rows := newTunnelRow().addRow(sqlmock.NewRows(tunnelColumns()))

	mock.ExpectQuery(`SELECT .+ FROM wireguard_tunnels ORDER BY created_at DESC LIMIT \$1 OFFSET \$2`).
		WithArgs(10, 0).
		WillReturnRows(rows)

	tunnels, err := repo.ListAll(context.Background(), 10, 0)
	require.NoError(t, err)
	assert.Len(t, tunnels, 1)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRepository_IncrementStats(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewRepository(db)
	id := uuid.New()

	now := time.Now().UTC()

	mock.ExpectExec(`UPDATE wireguard_tunnels SET`).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err = repo.IncrementStats(context.Background(), id, 100, 200, &now)
	require.NoError(t, err)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRepository_GetByID(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewRepository(db)
	tunnelID := uuid.New()
	userID := uuid.New()

	rows := sqlmock.NewRows([]string{
		"id", "user_id", "host_id", "interface_name",
		"server_private_key", "server_public_key", "client_public_key",
		"preshared_key", "server_ip", "client_ip", "listen_port",
		"allowed_ips", "dns_servers", "mtu", "persistent_keepalive",
		"status", "is_enabled", "last_handshake_at", "bytes_sent",
		"bytes_received", "last_rotated_at", "created_at", "updated_at",
	}).
		AddRow(tunnelID, userID, uuid.New(), "wg0",
			"priv", "pub", "cpub", nil, "10.200.200.1/24", "10.200.200.2/24",
			51820, []byte("{0.0.0.0/0}"), []byte("{}"), 1420, 25,
			"disabled", false, nil, int64(0), int64(0),
			nil, time.Now(), time.Now())

	mock.ExpectQuery(`SELECT .+ FROM wireguard_tunnels WHERE id = \$1 AND user_id = \$2`).
		WithArgs(tunnelID, userID).
		WillReturnRows(rows)

	tunnel, err := repo.GetByIDAndUser(context.Background(), tunnelID, userID)
	require.NoError(t, err)
	assert.Equal(t, tunnelID, tunnel.ID)
	assert.Equal(t, models.TunnelStatusDisabled, tunnel.Status)
	assert.False(t, tunnel.IsEnabled)

	assert.NoError(t, mock.ExpectationsWereMet())
}

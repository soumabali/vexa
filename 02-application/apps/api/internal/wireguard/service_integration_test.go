package wireguard

import (
	"context"
	"database/sql"
	"os"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/soumabali/vexa/internal/audit"
	"github.com/soumabali/vexa/internal/models"
)

func serviceConfig() ServerConfig {
	return ServerConfig{
		Subnet:            "10.200.200.0/24",
		WGPortRange:       [2]int{51820, 51830},
		RotationDays:      90,
		MaxTunnelsPerUser: 10,
		MaxTotalTunnels:   1000,
		Enabled:           true,
		BinPath:           "/nonexistent/wg",
	}
}

func buildTunnelRow(id, userID, hostID uuid.UUID, enabled bool, now time.Time) *sqlmock.Rows {
	return sqlmock.NewRows(tunnelColumns()).
		AddRow(id, userID, hostID, "wg0",
			"priv", "pub", "cpub", nil, "10.200.200.1/24", "10.200.200.2/24",
			51820, []byte("{0.0.0.0/0}"), []byte("{}"), 1420, 25,
			"connected", enabled, nil, int64(100), int64(200),
			nil, now, now)
}

func buildTunnelRowWithPSK(id, userID, hostID uuid.UUID, enabled bool, psk string, now time.Time) *sqlmock.Rows {
	return sqlmock.NewRows(tunnelColumns()).
		AddRow(id, userID, hostID, "wg0",
			"priv", "pub", "cpub", &psk, "10.200.200.1/24", "10.200.200.2/24",
			51820, []byte("{0.0.0.0/0}"), []byte("{}"), 1420, 25,
			"connected", enabled, nil, int64(100), int64(200),
			nil, now, now)
}

func TestWireGuardService_CreateTunnel_Disabled(t *testing.T) {
	svc := NewWireGuardService(nil, nil, ServerConfig{
		Enabled: false,
	})

	_, err := svc.CreateTunnel(context.Background(), uuid.New(), uuid.New(), &models.CreateTunnelRequest{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "disabled")
}

func TestWireGuardService_CreateTunnel_SubnetError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	svc := NewWireGuardService(db, nil, ServerConfig{
		Subnet:            "invalid-subnet",
		WGPortRange:       [2]int{51820, 51830},
		MaxTunnelsPerUser: 10,
		MaxTotalTunnels:   1000,
		Enabled:           true,
		BinPath:           "/nonexistent/wg",
	})

	userID := uuid.New()
	hostID := uuid.New()

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM wireguard_tunnels WHERE user_id = \$1`).
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM wireguard_tunnels`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	mock.ExpectQuery(`SELECT .+ FROM wireguard_tunnels WHERE user_id = \$1 AND host_id = \$2 ORDER BY created_at DESC`).
		WithArgs(userID, hostID.String()).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "user_id", "host_id", "interface_name",
			"server_private_key", "server_public_key", "client_public_key",
			"preshared_key", "server_ip", "client_ip", "listen_port",
			"allowed_ips", "dns_servers", "mtu", "persistent_keepalive",
			"status", "is_enabled", "last_handshake_at", "bytes_sent",
			"bytes_received", "last_rotated_at", "created_at", "updated_at",
		}))

	_, err = svc.CreateTunnel(context.Background(), userID, hostID, &models.CreateTunnelRequest{
		AllowedIPs: []string{"0.0.0.0/0"},
	})
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrInvalidSubnet)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestWireGuardService_CountLimits(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	svc := NewWireGuardService(db, nil, ServerConfig{
		Subnet:            "10.200.200.0/24",
		WGPortRange:       [2]int{51820, 51830},
		MaxTunnelsPerUser: 1,
		MaxTotalTunnels:   1000,
		Enabled:           true,
		BinPath:           "/nonexistent/wg",
	})

	userID := uuid.New()
	hostID := uuid.New()

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM wireguard_tunnels WHERE user_id = \$1`).
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	_, err = svc.CreateTunnel(context.Background(), userID, hostID, &models.CreateTunnelRequest{})
	assert.ErrorIs(t, err, ErrMaxTunnelsReached)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestWireGuardService_CreateTunnel_FullFlow(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	svc := NewWireGuardService(db, nil, ServerConfig{
		Subnet:            "10.200.200.0/24",
		WGPortRange:       [2]int{51820, 51830},
		MaxTunnelsPerUser: 10,
		MaxTotalTunnels:   1000,
		Enabled:           true,
		BinPath:           "/nonexistent/wg",
	})

	userID := uuid.New()
	hostID := uuid.New()

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM wireguard_tunnels WHERE user_id = \$1`).
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM wireguard_tunnels`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	mock.ExpectQuery(`SELECT .+ FROM wireguard_tunnels WHERE user_id = \$1 AND host_id = \$2 ORDER BY created_at DESC`).
		WithArgs(userID, hostID.String()).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "user_id", "host_id", "interface_name",
			"server_private_key", "server_public_key", "client_public_key",
			"preshared_key", "server_ip", "client_ip", "listen_port",
			"allowed_ips", "dns_servers", "mtu", "persistent_keepalive",
			"status", "is_enabled", "last_handshake_at", "bytes_sent",
			"bytes_received", "last_rotated_at", "created_at", "updated_at",
		}))

	mock.ExpectQuery(`INSERT INTO wireguard_tunnels`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at"}).
			AddRow(uuid.New(), time.Now(), time.Now()))

	mock.ExpectExec(`UPDATE wireguard_tunnels SET`).
		WillReturnResult(sqlmock.NewResult(0, 1))

	tunnel, err := svc.CreateTunnel(context.Background(), userID, hostID, &models.CreateTunnelRequest{
		AllowedIPs: []string{"10.0.0.0/8"},
		Port:       51820,
		UsePSK:     true,
	})
	require.NoError(t, err)
	assert.NotNil(t, tunnel)
	assert.Equal(t, models.TunnelStatusConnected, tunnel.Status)
	assert.Contains(t, tunnel.ServerIP, "10.200.200.1")
	assert.Contains(t, tunnel.ClientIP, "10.200.200.2")
	assert.True(t, tunnel.PresharedKey != nil)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestEnableTunnel(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	svc := NewWireGuardService(db, nil, serviceConfig())
	tunnelID := uuid.New()
	userID := uuid.New()
	hostID := uuid.New()
	now := time.Now()

	mock.ExpectQuery(`SELECT .+ FROM wireguard_tunnels WHERE id = \$1 AND user_id = \$2`).
		WithArgs(tunnelID, userID).
		WillReturnRows(buildTunnelRow(tunnelID, userID, hostID, false, now))

	mock.ExpectExec(`UPDATE wireguard_tunnels SET`).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err = svc.EnableTunnel(context.Background(), tunnelID, userID)
	require.NoError(t, err)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestEnableTunnel_AlreadyEnabled(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	svc := NewWireGuardService(db, nil, serviceConfig())
	tunnelID := uuid.New()
	userID := uuid.New()
	hostID := uuid.New()
	now := time.Now()

	mock.ExpectQuery(`SELECT .+ FROM wireguard_tunnels WHERE id = \$1 AND user_id = \$2`).
		WithArgs(tunnelID, userID).
		WillReturnRows(buildTunnelRow(tunnelID, userID, hostID, true, now))

	err = svc.EnableTunnel(context.Background(), tunnelID, userID)
	require.NoError(t, err)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestEnableTunnel_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	svc := NewWireGuardService(db, nil, serviceConfig())
	tunnelID := uuid.New()
	userID := uuid.New()

	mock.ExpectQuery(`SELECT .+ FROM wireguard_tunnels WHERE id = \$1 AND user_id = \$2`).
		WithArgs(tunnelID, userID).
		WillReturnError(sql.ErrNoRows)

	err = svc.EnableTunnel(context.Background(), tunnelID, userID)
	assert.ErrorIs(t, err, ErrTunnelNotFound)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDisableTunnel(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	svc := NewWireGuardService(db, nil, serviceConfig())
	tunnelID := uuid.New()
	userID := uuid.New()
	hostID := uuid.New()
	now := time.Now()

	mock.ExpectQuery(`SELECT .+ FROM wireguard_tunnels WHERE id = \$1 AND user_id = \$2`).
		WithArgs(tunnelID, userID).
		WillReturnRows(buildTunnelRow(tunnelID, userID, hostID, true, now))

	mock.ExpectExec(`UPDATE wireguard_tunnels SET`).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err = svc.DisableTunnel(context.Background(), tunnelID, userID)
	require.NoError(t, err)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDisableTunnel_AlreadyDisabled(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	svc := NewWireGuardService(db, nil, serviceConfig())
	tunnelID := uuid.New()
	userID := uuid.New()
	hostID := uuid.New()
	now := time.Now()

	mock.ExpectQuery(`SELECT .+ FROM wireguard_tunnels WHERE id = \$1 AND user_id = \$2`).
		WithArgs(tunnelID, userID).
		WillReturnRows(buildTunnelRow(tunnelID, userID, hostID, false, now))

	err = svc.DisableTunnel(context.Background(), tunnelID, userID)
	require.NoError(t, err)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDeleteTunnel(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	svc := NewWireGuardService(db, nil, serviceConfig())
	tunnelID := uuid.New()
	userID := uuid.New()
	hostID := uuid.New()
	now := time.Now()

	mock.ExpectQuery(`SELECT .+ FROM wireguard_tunnels WHERE id = \$1 AND user_id = \$2`).
		WithArgs(tunnelID, userID).
		WillReturnRows(buildTunnelRow(tunnelID, userID, hostID, true, now))

	mock.ExpectExec(`DELETE FROM wireguard_tunnels WHERE id = \$1`).
		WithArgs(tunnelID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err = svc.DeleteTunnel(context.Background(), tunnelID, userID)
	require.NoError(t, err)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDeleteTunnel_NotEnabled(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	svc := NewWireGuardService(db, nil, serviceConfig())
	tunnelID := uuid.New()
	userID := uuid.New()
	hostID := uuid.New()
	now := time.Now()

	mock.ExpectQuery(`SELECT .+ FROM wireguard_tunnels WHERE id = \$1 AND user_id = \$2`).
		WithArgs(tunnelID, userID).
		WillReturnRows(buildTunnelRow(tunnelID, userID, hostID, false, now))

	mock.ExpectExec(`DELETE FROM wireguard_tunnels WHERE id = \$1`).
		WithArgs(tunnelID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err = svc.DeleteTunnel(context.Background(), tunnelID, userID)
	require.NoError(t, err)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRotateKeys(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	svc := NewWireGuardService(db, nil, serviceConfig())
	tunnelID := uuid.New()
	userID := uuid.New()
	hostID := uuid.New()
	now := time.Now()

	mock.ExpectQuery(`SELECT .+ FROM wireguard_tunnels WHERE id = \$1 AND user_id = \$2`).
		WithArgs(tunnelID, userID).
		WillReturnRows(buildTunnelRow(tunnelID, userID, hostID, true, now))

	mock.ExpectExec(`UPDATE wireguard_tunnels SET`).
		WillReturnResult(sqlmock.NewResult(0, 1))

	tunnel, err := svc.RotateKeys(context.Background(), tunnelID, userID)
	require.NoError(t, err)
	assert.NotNil(t, tunnel)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRotateKeys_NotEnabled(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	svc := NewWireGuardService(db, nil, serviceConfig())
	tunnelID := uuid.New()
	userID := uuid.New()
	hostID := uuid.New()
	now := time.Now()

	mock.ExpectQuery(`SELECT .+ FROM wireguard_tunnels WHERE id = \$1 AND user_id = \$2`).
		WithArgs(tunnelID, userID).
		WillReturnRows(buildTunnelRow(tunnelID, userID, hostID, false, now))

	mock.ExpectExec(`UPDATE wireguard_tunnels SET`).
		WillReturnResult(sqlmock.NewResult(0, 1))

	tunnel, err := svc.RotateKeys(context.Background(), tunnelID, userID)
	require.NoError(t, err)
	assert.NotNil(t, tunnel)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetConfig(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	svc := NewWireGuardService(db, nil, serviceConfig())
	tunnelID := uuid.New()
	userID := uuid.New()
	hostID := uuid.New()
	now := time.Now()

	mock.ExpectQuery(`SELECT .+ FROM wireguard_tunnels WHERE id = \$1 AND user_id = \$2`).
		WithArgs(tunnelID, userID).
		WillReturnRows(buildTunnelRow(tunnelID, userID, hostID, true, now))

	mock.ExpectQuery(`SELECT address FROM hosts WHERE id = \$1`).
		WithArgs(hostID).
		WillReturnRows(sqlmock.NewRows([]string{"address"}).AddRow("203.0.113.1"))

	export, err := svc.GetConfig(context.Background(), tunnelID, userID)
	require.NoError(t, err)
	assert.Equal(t, "203.0.113.1", export.Endpoint)
	assert.Equal(t, "10.200.200.2/24", export.Address)
	assert.Contains(t, export.AllowedIPs, "0.0.0.0/0")

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetConfig_WithPSK(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	svc := NewWireGuardService(db, nil, serviceConfig())
	tunnelID := uuid.New()
	userID := uuid.New()
	hostID := uuid.New()
	now := time.Now()

	psk := "my-preshared-key"
	mock.ExpectQuery(`SELECT .+ FROM wireguard_tunnels WHERE id = \$1 AND user_id = \$2`).
		WithArgs(tunnelID, userID).
		WillReturnRows(buildTunnelRowWithPSK(tunnelID, userID, hostID, true, psk, now))

	mock.ExpectQuery(`SELECT address FROM hosts WHERE id = \$1`).
		WithArgs(hostID).
		WillReturnRows(sqlmock.NewRows([]string{"address"}).AddRow("203.0.113.1"))

	export, err := svc.GetConfig(context.Background(), tunnelID, userID)
	require.NoError(t, err)
	assert.Equal(t, psk, export.PresharedKey)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetStats(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	svc := NewWireGuardService(db, nil, serviceConfig())
	tunnelID := uuid.New()
	userID := uuid.New()
	hostID := uuid.New()
	now := time.Now()

	mock.ExpectQuery(`SELECT .+ FROM wireguard_tunnels WHERE id = \$1 AND user_id = \$2`).
		WithArgs(tunnelID, userID).
		WillReturnRows(buildTunnelRow(tunnelID, userID, hostID, true, now))

	mock.ExpectExec(`UPDATE wireguard_tunnels SET bytes_sent`).
		WillReturnResult(sqlmock.NewResult(0, 1))

	stats, err := svc.GetStats(context.Background(), tunnelID, userID)
	require.NoError(t, err)
	assert.Equal(t, tunnelID, stats.ID)
	assert.Equal(t, int64(0), stats.BytesSent)
	assert.Equal(t, int64(0), stats.BytesReceived)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetTunnel(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	svc := NewWireGuardService(db, nil, serviceConfig())
	tunnelID := uuid.New()
	userID := uuid.New()
	hostID := uuid.New()
	now := time.Now()

	mock.ExpectQuery(`SELECT .+ FROM wireguard_tunnels WHERE id = \$1 AND user_id = \$2`).
		WithArgs(tunnelID, userID).
		WillReturnRows(buildTunnelRow(tunnelID, userID, hostID, true, now))

	tunnel, err := svc.GetTunnel(context.Background(), tunnelID, userID)
	require.NoError(t, err)
	assert.Equal(t, tunnelID, tunnel.ID)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetTunnel_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	svc := NewWireGuardService(db, nil, serviceConfig())
	tunnelID := uuid.New()
	userID := uuid.New()

	mock.ExpectQuery(`SELECT .+ FROM wireguard_tunnels WHERE id = \$1 AND user_id = \$2`).
		WithArgs(tunnelID, userID).
		WillReturnError(sql.ErrNoRows)

	_, err = svc.GetTunnel(context.Background(), tunnelID, userID)
	assert.ErrorIs(t, err, ErrTunnelNotFound)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestListTunnels(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	svc := NewWireGuardService(db, nil, serviceConfig())
	userID := uuid.New()
	hostID := uuid.New()
	now := time.Now()
	tunnelID := uuid.New()

	mock.ExpectQuery(`SELECT .+ FROM wireguard_tunnels WHERE user_id = \$1 AND host_id = \$2 ORDER BY created_at DESC`).
		WithArgs(userID, hostID.String()).
		WillReturnRows(buildTunnelRow(tunnelID, userID, hostID, true, now))

	tunnels, err := svc.ListTunnels(context.Background(), userID, hostID.String(), nil)
	require.NoError(t, err)
	assert.Len(t, tunnels, 1)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUpdateTunnel(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	svc := NewWireGuardService(db, nil, serviceConfig())
	tunnelID := uuid.New()
	userID := uuid.New()
	hostID := uuid.New()
	now := time.Now()

	mock.ExpectQuery(`SELECT .+ FROM wireguard_tunnels WHERE id = \$1 AND user_id = \$2`).
		WithArgs(tunnelID, userID).
		WillReturnRows(buildTunnelRow(tunnelID, userID, hostID, true, now))

	port := 51821
	mtu := 1400
	mock.ExpectExec(`UPDATE wireguard_tunnels SET`).
		WillReturnResult(sqlmock.NewResult(0, 1))

	tunnel, err := svc.UpdateTunnel(context.Background(), tunnelID, userID, &models.UpdateTunnelRequest{
		Port:       &port,
		MTU:        &mtu,
		AllowedIPs: []string{"10.0.0.0/8"},
		DNSServers: []string{"1.1.1.1"},
	})
	require.NoError(t, err)
	assert.Equal(t, port, tunnel.ListenPort)
	assert.Equal(t, mtu, tunnel.MTU)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUpdateTunnel_NoChanges(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	svc := NewWireGuardService(db, nil, serviceConfig())
	tunnelID := uuid.New()
	userID := uuid.New()
	hostID := uuid.New()
	now := time.Now()

	mock.ExpectQuery(`SELECT .+ FROM wireguard_tunnels WHERE id = \$1 AND user_id = \$2`).
		WithArgs(tunnelID, userID).
		WillReturnRows(buildTunnelRow(tunnelID, userID, hostID, true, now))

	tunnel, err := svc.UpdateTunnel(context.Background(), tunnelID, userID, &models.UpdateTunnelRequest{})
	require.NoError(t, err)
	assert.NotNil(t, tunnel)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestService_Stop(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	svc := NewWireGuardService(db, nil, serviceConfig())
	svc.Stop()

	assert.False(t, svc.StatsCollector().IsRunning())
	assert.False(t, svc.RotationScheduler().IsRunning())
}

func TestWireGuardService_CreateTunnel_Duplicate(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	svc := NewWireGuardService(db, nil, ServerConfig{
		Subnet:            "10.200.200.0/24",
		WGPortRange:       [2]int{51820, 51830},
		MaxTunnelsPerUser: 10,
		MaxTotalTunnels:   1000,
		Enabled:           true,
		BinPath:           "/nonexistent/wg",
	})

	userID := uuid.New()
	hostID := uuid.New()
	now := time.Now()
	existingID := uuid.New()

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM wireguard_tunnels WHERE user_id = \$1`).
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM wireguard_tunnels`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	mock.ExpectQuery(`SELECT .+ FROM wireguard_tunnels WHERE user_id = \$1 AND host_id = \$2 ORDER BY created_at DESC`).
		WithArgs(userID, hostID.String()).
		WillReturnRows(buildTunnelRow(existingID, userID, hostID, true, now))

	_, err = svc.CreateTunnel(context.Background(), userID, hostID, &models.CreateTunnelRequest{})
	assert.ErrorIs(t, err, ErrTunnelExists)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestWireGuardService_CreateTunnel_MaxTotal(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	svc := NewWireGuardService(db, nil, ServerConfig{
		Subnet:            "10.200.200.0/24",
		WGPortRange:       [2]int{51820, 51830},
		MaxTunnelsPerUser: 10,
		MaxTotalTunnels:   1,
		Enabled:           true,
		BinPath:           "/nonexistent/wg",
	})

	userID := uuid.New()
	hostID := uuid.New()

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM wireguard_tunnels WHERE user_id = \$1`).
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM wireguard_tunnels`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	_, err = svc.CreateTunnel(context.Background(), userID, hostID, &models.CreateTunnelRequest{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "maximum total tunnels reached")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestWireGuardService_CreateTunnel_WithAuditLog(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "audit-*.log")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	auditLogger, err := audit.NewLogger(nil, tmpFile.Name(), []byte("test-key-32-bytes-long-for-hmac!"))
	require.NoError(t, err)
	defer auditLogger.Close()

	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	svc := NewWireGuardService(db, auditLogger, ServerConfig{
		Subnet:            "10.200.200.0/24",
		WGPortRange:       [2]int{51820, 51830},
		MaxTunnelsPerUser: 10,
		MaxTotalTunnels:   1000,
		Enabled:           true,
		BinPath:           "/nonexistent/wg",
	})

	userID := uuid.New()
	hostID := uuid.New()

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM wireguard_tunnels WHERE user_id = \$1`).
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM wireguard_tunnels`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	mock.ExpectQuery(`SELECT .+ FROM wireguard_tunnels WHERE user_id = \$1 AND host_id = \$2 ORDER BY created_at DESC`).
		WithArgs(userID, hostID.String()).
		WillReturnRows(sqlmock.NewRows(tunnelColumns()))

	mock.ExpectQuery(`INSERT INTO wireguard_tunnels`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at"}).
			AddRow(uuid.New(), time.Now(), time.Now()))

	mock.ExpectExec(`UPDATE wireguard_tunnels SET`).
		WillReturnResult(sqlmock.NewResult(0, 1))

	tunnel, err := svc.CreateTunnel(context.Background(), userID, hostID, &models.CreateTunnelRequest{})
	require.NoError(t, err)
	assert.NotNil(t, tunnel)
	assert.NoError(t, mock.ExpectationsWereMet())

	content, err := os.ReadFile(tmpFile.Name())
	require.NoError(t, err)
	assert.Contains(t, string(content), "tunnel.create")
	assert.Contains(t, string(content), tunnel.ID.String())
}

func TestDisableTunnel_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	svc := NewWireGuardService(db, nil, serviceConfig())
	tunnelID := uuid.New()
	userID := uuid.New()

	mock.ExpectQuery(`SELECT .+ FROM wireguard_tunnels WHERE id = \$1 AND user_id = \$2`).
		WithArgs(tunnelID, userID).
		WillReturnError(sql.ErrNoRows)

	err = svc.DisableTunnel(context.Background(), tunnelID, userID)
	assert.ErrorIs(t, err, ErrTunnelNotFound)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDeleteTunnel_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	svc := NewWireGuardService(db, nil, serviceConfig())
	tunnelID := uuid.New()
	userID := uuid.New()

	mock.ExpectQuery(`SELECT .+ FROM wireguard_tunnels WHERE id = \$1 AND user_id = \$2`).
		WithArgs(tunnelID, userID).
		WillReturnError(sql.ErrNoRows)

	err = svc.DeleteTunnel(context.Background(), tunnelID, userID)
	assert.ErrorIs(t, err, ErrTunnelNotFound)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRotateKeys_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	svc := NewWireGuardService(db, nil, serviceConfig())
	tunnelID := uuid.New()
	userID := uuid.New()

	mock.ExpectQuery(`SELECT .+ FROM wireguard_tunnels WHERE id = \$1 AND user_id = \$2`).
		WithArgs(tunnelID, userID).
		WillReturnError(sql.ErrNoRows)

	_, err = svc.RotateKeys(context.Background(), tunnelID, userID)
	assert.ErrorIs(t, err, ErrTunnelNotFound)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestListTunnels_Empty(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	svc := NewWireGuardService(db, nil, serviceConfig())
	userID := uuid.New()

	mock.ExpectQuery(`SELECT .+ FROM wireguard_tunnels WHERE user_id = \$1 ORDER BY created_at DESC`).
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows(tunnelColumns()))

	tunnels, err := svc.ListTunnels(context.Background(), userID, "", nil)
	require.NoError(t, err)
	assert.Len(t, tunnels, 0)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestEnableTunnel_WithAuditLog(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "audit-*.log")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	auditLogger, err := audit.NewLogger(nil, tmpFile.Name(), []byte("test-key-32-bytes-long-for-hmac!"))
	require.NoError(t, err)
	defer auditLogger.Close()

	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	svc := NewWireGuardService(db, auditLogger, serviceConfig())
	tunnelID := uuid.New()
	userID := uuid.New()
	hostID := uuid.New()
	now := time.Now()

	mock.ExpectQuery(`SELECT .+ FROM wireguard_tunnels WHERE id = \$1 AND user_id = \$2`).
		WithArgs(tunnelID, userID).
		WillReturnRows(buildTunnelRow(tunnelID, userID, hostID, false, now))

	mock.ExpectExec(`UPDATE wireguard_tunnels SET`).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err = svc.EnableTunnel(context.Background(), tunnelID, userID)
	require.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())

	content, err := os.ReadFile(tmpFile.Name())
	require.NoError(t, err)
	assert.Contains(t, string(content), "tunnel.enable")
}

func TestCreateTunnel_DefaultAllowedIPs(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	svc := NewWireGuardService(db, nil, ServerConfig{
		Subnet:            "10.200.200.0/24",
		WGPortRange:       [2]int{51820, 51830},
		MaxTunnelsPerUser: 10,
		MaxTotalTunnels:   1000,
		Enabled:           true,
		BinPath:           "/nonexistent/wg",
	})

	userID := uuid.New()
	hostID := uuid.New()

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM wireguard_tunnels WHERE user_id = \$1`).
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM wireguard_tunnels`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	mock.ExpectQuery(`SELECT .+ FROM wireguard_tunnels WHERE user_id = \$1 AND host_id = \$2 ORDER BY created_at DESC`).
		WithArgs(userID, hostID.String()).
		WillReturnRows(sqlmock.NewRows(tunnelColumns()))

	mock.ExpectQuery(`INSERT INTO wireguard_tunnels`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at"}).
			AddRow(uuid.New(), time.Now(), time.Now()))

	mock.ExpectExec(`UPDATE wireguard_tunnels SET`).
		WillReturnResult(sqlmock.NewResult(0, 1))

	tunnel, err := svc.CreateTunnel(context.Background(), userID, hostID, &models.CreateTunnelRequest{})
	require.NoError(t, err)
	assert.Equal(t, []string{"0.0.0.0/0", "::/0"}, tunnel.AllowedIPs)
	assert.Equal(t, 1420, tunnel.MTU)
	assert.Equal(t, 51820, tunnel.ListenPort)
	assert.NoError(t, mock.ExpectationsWereMet())
}

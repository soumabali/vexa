package integration

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/soumabali/vexa/internal/models"
	"github.com/soumabali/vexa/internal/wireguard"
)

func tunnelCols() []string {
	return []string{
		"id", "user_id", "host_id", "interface_name",
		"server_private_key", "server_public_key", "client_public_key",
		"preshared_key", "server_ip", "client_ip", "listen_port",
		"allowed_ips", "dns_servers", "mtu", "persistent_keepalive",
		"status", "is_enabled", "last_handshake_at", "bytes_sent",
		"bytes_received", "last_rotated_at", "created_at", "updated_at",
	}
}

func TestTunnelE2E_FullLifecycle(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	svc := wireguard.NewWireGuardService(db, nil, wireguard.ServerConfig{
		Subnet:            "10.200.200.0/24",
		WGPortRange:       [2]int{51820, 51830},
		RotationDays:      90,
		MaxTunnelsPerUser: 10,
		MaxTotalTunnels:   1000,
		Enabled:           true,
		BinPath:           "/nonexistent/wg",
	})

	userID := uuid.New()
	hostID := uuid.New()
	now := time.Now()

	// Step 1: Create tunnel
	t.Log("Step 1: Create tunnel")
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM wireguard_tunnels WHERE user_id = \$1`).
		WithArgs(userID).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM wireguard_tunnels`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	mock.ExpectQuery(`SELECT .+ FROM wireguard_tunnels WHERE user_id = \$1 AND host_id = \$2 ORDER BY created_at DESC`).
		WithArgs(userID, hostID.String()).
		WillReturnRows(sqlmock.NewRows(tunnelCols()))
	mock.ExpectQuery(`INSERT INTO wireguard_tunnels`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at"}).
			AddRow(uuid.New(), now, now))
	mock.ExpectExec(`UPDATE wireguard_tunnels SET`).
		WillReturnResult(sqlmock.NewResult(0, 1))

	tunnel, err := svc.CreateTunnel(context.Background(), userID, hostID, &models.CreateTunnelRequest{
		Port:       51820,
		AllowedIPs: []string{"10.0.0.0/8"},
		MTU:        1400,
	})
	require.NoError(t, err)
	require.NotNil(t, tunnel)
	assert.Equal(t, models.TunnelStatusConnected, tunnel.Status)
	assert.Contains(t, tunnel.ServerIP, "10.200.200.1")
	assert.Contains(t, tunnel.ClientIP, "10.200.200.2")
	t.Logf("Created tunnel: %s", tunnel.ID)

	// Step 2: Enable tunnel
	t.Log("Step 2: Enable tunnel")
	mock.ExpectQuery(`SELECT .+ FROM wireguard_tunnels WHERE id = \$1 AND user_id = \$2`).
		WithArgs(tunnel.ID, userID).
		WillReturnRows(sqlmock.NewRows(tunnelCols()).
			AddRow(tunnel.ID, userID, hostID, "wg0",
				"priv", "pub", "cpub", nil, "10.200.200.1/24", "10.200.200.2/24",
				51820, []byte("{10.0.0.0/8}"), []byte("{}"), 1400, 25,
				"connected", false, nil, int64(0), int64(0),
				nil, now, now))
	mock.ExpectExec(`UPDATE wireguard_tunnels SET`).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err = svc.EnableTunnel(context.Background(), tunnel.ID, userID)
	require.NoError(t, err)

	// Step 3: Verify tunnel stats
	t.Log("Step 3: Verify tunnel stats")
	mock.ExpectQuery(`SELECT .+ FROM wireguard_tunnels WHERE id = \$1 AND user_id = \$2`).
		WithArgs(tunnel.ID, userID).
		WillReturnRows(sqlmock.NewRows(tunnelCols()).
			AddRow(tunnel.ID, userID, hostID, "wg0",
				"priv", "pub", "cpub", nil, "10.200.200.1/24", "10.200.200.2/24",
				51820, []byte("{10.0.0.0/8}"), []byte("{}"), 1400, 25,
				"connected", true, nil, int64(500), int64(1000),
				nil, now, now))
	mock.ExpectExec(`UPDATE wireguard_tunnels SET bytes_sent`).
		WillReturnResult(sqlmock.NewResult(0, 1))

	stats, err := svc.GetStats(context.Background(), tunnel.ID, userID)
	require.NoError(t, err)
	assert.True(t, stats.IsEnabled)
	t.Logf("Stats: sent=%d recv=%d", stats.BytesSent, stats.BytesReceived)

	// Step 4: Get config
	t.Log("Step 4: Get config")
	mock.ExpectQuery(`SELECT .+ FROM wireguard_tunnels WHERE id = \$1 AND user_id = \$2`).
		WithArgs(tunnel.ID, userID).
		WillReturnRows(sqlmock.NewRows(tunnelCols()).
			AddRow(tunnel.ID, userID, hostID, "wg0",
				"priv", "pub", "cpub", nil, "10.200.200.1/24", "10.200.200.2/24",
				51820, []byte("{10.0.0.0/8}"), []byte("{}"), 1400, 25,
				"connected", true, nil, int64(500), int64(1000),
				nil, now, now))
	mock.ExpectQuery(`SELECT address FROM hosts WHERE id = \$1`).
		WithArgs(hostID).
		WillReturnRows(sqlmock.NewRows([]string{"address"}).AddRow("203.0.113.1"))

	export, err := svc.GetConfig(context.Background(), tunnel.ID, userID)
	require.NoError(t, err)
	assert.Equal(t, "203.0.113.1", export.Endpoint)
	assert.Equal(t, "10.200.200.2/24", export.Address)
	t.Logf("Config endpoint: %s:%d", export.Endpoint, export.Port)

	// Step 5: Rotate keys
	t.Log("Step 5: Rotate keys")
	mock.ExpectQuery(`SELECT .+ FROM wireguard_tunnels WHERE id = \$1 AND user_id = \$2`).
		WithArgs(tunnel.ID, userID).
		WillReturnRows(sqlmock.NewRows(tunnelCols()).
			AddRow(tunnel.ID, userID, hostID, "wg0",
				"priv", "pub", "cpub", nil, "10.200.200.1/24", "10.200.200.2/24",
				51820, []byte("{10.0.0.0/8}"), []byte("{}"), 1400, 25,
				"connected", true, nil, int64(500), int64(1000),
				nil, now, now))
	mock.ExpectExec(`UPDATE wireguard_tunnels SET`).
		WillReturnResult(sqlmock.NewResult(0, 1))

	rotated, err := svc.RotateKeys(context.Background(), tunnel.ID, userID)
	require.NoError(t, err)
	assert.NotEqual(t, tunnel.ServerPublicKey, rotated.ServerPublicKey, "public key should change after rotation")
	require.NotNil(t, rotated.LastRotatedAt)
	t.Logf("Rotated at: %s", rotated.LastRotatedAt.Format(time.RFC3339))

	// Step 6: Disable tunnel
	t.Log("Step 6: Disable tunnel")
	mock.ExpectQuery(`SELECT .+ FROM wireguard_tunnels WHERE id = \$1 AND user_id = \$2`).
		WithArgs(tunnel.ID, userID).
		WillReturnRows(sqlmock.NewRows(tunnelCols()).
			AddRow(tunnel.ID, userID, hostID, "wg0",
				"priv", "pub", "cpub", nil, "10.200.200.1/24", "10.200.200.2/24",
				51820, []byte("{10.0.0.0/8}"), []byte("{}"), 1400, 25,
				"connected", true, nil, int64(500), int64(1000),
				nil, now, now))
	mock.ExpectExec(`UPDATE wireguard_tunnels SET`).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err = svc.DisableTunnel(context.Background(), tunnel.ID, userID)
	require.NoError(t, err)

	// Step 7: Delete tunnel
	t.Log("Step 7: Delete tunnel")
	mock.ExpectQuery(`SELECT .+ FROM wireguard_tunnels WHERE id = \$1 AND user_id = \$2`).
		WithArgs(tunnel.ID, userID).
		WillReturnRows(sqlmock.NewRows(tunnelCols()).
			AddRow(tunnel.ID, userID, hostID, "wg0",
				"priv", "pub", "cpub", nil, "10.200.200.1/24", "10.200.200.2/24",
				51820, []byte("{10.0.0.0/8}"), []byte("{}"), 1400, 25,
				"disconnected", false, nil, int64(500), int64(1000),
				nil, now, now))
	mock.ExpectExec(`DELETE FROM wireguard_tunnels WHERE id = \$1`).
		WithArgs(tunnel.ID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err = svc.DeleteTunnel(context.Background(), tunnel.ID, userID)
	require.NoError(t, err)

	assert.NoError(t, mock.ExpectationsWereMet())
	t.Log("E2E tunnel lifecycle complete: create → enable → stats → config → rotate → disable → delete")
}

func TestTunnelE2E_RotateTwice(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	svc := wireguard.NewWireGuardService(db, nil, wireguard.ServerConfig{
		Subnet:            "10.200.200.0/24",
		WGPortRange:       [2]int{51820, 51830},
		RotationDays:      90,
		MaxTunnelsPerUser: 10,
		MaxTotalTunnels:   1000,
		Enabled:           true,
		BinPath:           "/nonexistent/wg",
	})

	userID := uuid.New()
	hostID := uuid.New()
	tunnelID := uuid.New()
	now := time.Now()

	for i := 0; i < 2; i++ {
		mock.ExpectQuery(`SELECT .+ FROM wireguard_tunnels WHERE id = \$1 AND user_id = \$2`).
			WithArgs(tunnelID, userID).
			WillReturnRows(sqlmock.NewRows(tunnelCols()).
				AddRow(tunnelID, userID, hostID, "wg0",
					"priv", "pub", "cpub", nil, "10.200.200.1/24", "10.200.200.2/24",
					51820, []byte("{0.0.0.0/0}"), []byte("{}"), 1420, 25,
					"connected", true, nil, int64(0), int64(0),
					nil, now, now))
		mock.ExpectExec(`UPDATE wireguard_tunnels SET`).
			WillReturnResult(sqlmock.NewResult(0, 1))

		tunnel, err := svc.RotateKeys(context.Background(), tunnelID, userID)
		require.NoError(t, err)
		assert.NotNil(t, tunnel)
		assert.NotNil(t, tunnel.LastRotatedAt)
		t.Logf("Rotation %d complete at %s", i+1, tunnel.LastRotatedAt.Format(time.RFC3339))
	}

	assert.NoError(t, mock.ExpectationsWereMet())
}

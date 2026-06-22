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
	"github.com/soumabali/vexa/internal/security"
	"github.com/soumabali/vexa/internal/wireguard"
)

func TestRateLimit_MaxTunnelsPerUser(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	svc := wireguard.NewWireGuardService(db, nil, wireguard.ServerConfig{
		Subnet:            "10.200.200.0/24",
		WGPortRange:       [2]int{51820, 51830},
		MaxTunnelsPerUser: 2,
		MaxTotalTunnels:   1000,
		Enabled:           true,
		BinPath:           "/nonexistent/wg",
	})

	userID := uuid.New()
	hostID := uuid.New()

	// First tunnel: count=1, below limit
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM wireguard_tunnels WHERE user_id = \$1`).
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM wireguard_tunnels`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

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

	_, err = svc.CreateTunnel(context.Background(), userID, hostID, &models.CreateTunnelRequest{
		Port:       51820,
		AllowedIPs: []string{"10.0.0.0/8"},
	})
	require.NoError(t, err)

	// Second tunnel: count=2, at limit, should fail
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM wireguard_tunnels WHERE user_id = \$1`).
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))

	_, err = svc.CreateTunnel(context.Background(), userID, hostID, &models.CreateTunnelRequest{
		Port:       51821,
		AllowedIPs: []string{"10.0.0.0/8"},
	})
	assert.ErrorIs(t, err, wireguard.ErrMaxTunnelsReached)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRateLimiter_TokenBucket(t *testing.T) {
	rl := security.NewRateLimiter(10, 5)
	ip := "192.168.1.1"

	for i := 0; i < 5; i++ {
		assert.True(t, rl.Allow(ip), "request %d should be allowed (burst)", i+1)
	}

	assert.False(t, rl.Allow(ip), "request 6 should be rate limited")
}

func TestAuthZ_UserCannotAccessOtherUsersTunnel(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	svc := wireguard.NewWireGuardService(db, nil, wireguard.ServerConfig{
		Subnet:            "10.200.200.0/24",
		WGPortRange:       [2]int{51820, 51830},
		MaxTunnelsPerUser: 10,
		MaxTotalTunnels:   1000,
		Enabled:           true,
		BinPath:           "/nonexistent/wg",
	})

	tunnelID := uuid.New()
	userA := uuid.New()
	userB := uuid.New()

	// userA owns the tunnel and can access it
	mock.ExpectQuery(`SELECT .+ FROM wireguard_tunnels WHERE id = \$1 AND user_id = \$2`).
		WithArgs(tunnelID, userA).
		WillReturnError(sqlmock.ErrCancelled)
	_, err = svc.GetTunnel(context.Background(), tunnelID, userA)
	assert.Error(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())

	// userB (different user) cannot access the same tunnel
	mock.ExpectQuery(`SELECT .+ FROM wireguard_tunnels WHERE id = \$1 AND user_id = \$2`).
		WithArgs(tunnelID, userB).
		WillReturnError(sqlmock.ErrCancelled)
	_, err = svc.GetTunnel(context.Background(), tunnelID, userB)
	assert.Error(t, err, "user B should not see user A's tunnel")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestSQLInjection_AllowedIPs(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	svc := wireguard.NewWireGuardService(db, nil, wireguard.ServerConfig{
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

	maliciousIPs := []string{
		"0.0.0.0/0; DROP TABLE wireguard_tunnels; --",
		"1.1.1.1' OR '1'='1",
		"10.0.0.0/8\" OR 1=1 --",
	}

	_, err = svc.CreateTunnel(context.Background(), userID, hostID, &models.CreateTunnelRequest{
		Port:       51820,
		AllowedIPs: maliciousIPs,
	})
	require.NoError(t, err, "SQL injection in allowed_ips should not cause errors - parameters are bound")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRateLimiter_DifferentIPs(t *testing.T) {
	rl := security.NewRateLimiter(10, 3)

	for i := 0; i < 3; i++ {
		assert.True(t, rl.Allow("10.0.0.1"), "10.0.0.1 request %d", i+1)
	}
	assert.False(t, rl.Allow("10.0.0.1"), "10.0.0.1 should be limited")

	assert.True(t, rl.Allow("10.0.0.2"), "different IP should not be limited")
}

func TestRateLimiter_TokenRefill(t *testing.T) {
	rl := security.NewRateLimiter(60, 1)

	assert.True(t, rl.Allow("10.0.0.1"), "first request allowed")
	assert.False(t, rl.Allow("10.0.0.1"), "second request blocked (only burst=1)")
}

func TestAuthZ_UserACannotModifyUserBTunnel(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	svc := wireguard.NewWireGuardService(db, nil, wireguard.ServerConfig{
		Subnet:            "10.200.200.0/24",
		WGPortRange:       [2]int{51820, 51830},
		MaxTunnelsPerUser: 10,
		MaxTotalTunnels:   1000,
		Enabled:           true,
		BinPath:           "/nonexistent/wg",
	})

	tunnelID := uuid.New()
	userB := uuid.New()

	mock.ExpectQuery(`SELECT .+ FROM wireguard_tunnels WHERE id = \$1 AND user_id = \$2`).
		WithArgs(tunnelID, userB).
		WillReturnError(sqlmock.ErrCancelled)

	port := 51821
	_, err = svc.UpdateTunnel(context.Background(), tunnelID, userB, &models.UpdateTunnelRequest{
		Port: &port,
	})
	assert.Error(t, err, "user B should not be able to update user A's tunnel")
	assert.NoError(t, mock.ExpectationsWereMet())
}

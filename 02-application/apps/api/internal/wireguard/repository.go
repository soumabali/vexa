package wireguard

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"

	"github.com/soumabali/vexa/internal/models"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(ctx context.Context, tunnel *models.WireGuardTunnel) error {
	query := `
		INSERT INTO wireguard_tunnels
			(user_id, host_id, interface_name, server_private_key, server_public_key,
			 client_public_key, preshared_key, server_ip, client_ip, listen_port,
			 allowed_ips, dns_servers, mtu, persistent_keepalive, status, is_enabled)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
		RETURNING id, created_at, updated_at
	`
	err := r.db.QueryRowContext(ctx, query,
		tunnel.UserID, tunnel.HostID, tunnel.InterfaceName,
		tunnel.ServerPrivateKey, tunnel.ServerPublicKey,
		tunnel.ClientPublicKey, tunnel.PresharedKey,
		tunnel.ServerIP, tunnel.ClientIP, tunnel.ListenPort,
		pq.Array(tunnel.AllowedIPs), pq.Array(tunnel.DNSServers),
		tunnel.MTU, tunnel.PersistentKeepalive,
		tunnel.Status, tunnel.IsEnabled,
	).Scan(&tunnel.ID, &tunnel.CreatedAt, &tunnel.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to create wireguard tunnel: %w", err)
	}
	return nil
}

func (r *Repository) GetByID(ctx context.Context, id uuid.UUID) (*models.WireGuardTunnel, error) {
	query := `
		SELECT id, user_id, host_id, interface_name, server_private_key, server_public_key,
		       client_public_key, preshared_key, server_ip, client_ip, listen_port,
		       allowed_ips, dns_servers, mtu, persistent_keepalive, status, is_enabled,
		       last_handshake_at, bytes_sent, bytes_received, last_rotated_at, created_at, updated_at
		FROM wireguard_tunnels WHERE id = $1
	`
	return r.scanTunnel(r.db.QueryRowContext(ctx, query, id))
}

func (r *Repository) GetByIDAndUser(ctx context.Context, id, userID uuid.UUID) (*models.WireGuardTunnel, error) {
	query := `
		SELECT id, user_id, host_id, interface_name, server_private_key, server_public_key,
		       client_public_key, preshared_key, server_ip, client_ip, listen_port,
		       allowed_ips, dns_servers, mtu, persistent_keepalive, status, is_enabled,
		       last_handshake_at, bytes_sent, bytes_received, last_rotated_at, created_at, updated_at
		FROM wireguard_tunnels WHERE id = $1 AND user_id = $2
	`
	return r.scanTunnel(r.db.QueryRowContext(ctx, query, id, userID))
}

func (r *Repository) ListByUser(ctx context.Context, userID uuid.UUID, hostID string, enabled *bool) ([]*models.WireGuardTunnel, error) {
	args := []interface{}{userID}
	query := `
		SELECT id, user_id, host_id, interface_name, server_private_key, server_public_key,
		       client_public_key, preshared_key, server_ip, client_ip, listen_port,
		       allowed_ips, dns_servers, mtu, persistent_keepalive, status, is_enabled,
		       last_handshake_at, bytes_sent, bytes_received, last_rotated_at, created_at, updated_at
		FROM wireguard_tunnels WHERE user_id = $1
	`

	argIdx := 1
	if hostID != "" {
		argIdx++
		query += fmt.Sprintf(" AND host_id = $%d", argIdx)
		args = append(args, hostID)
	}
	if enabled != nil {
		argIdx++
		query += fmt.Sprintf(" AND is_enabled = $%d", argIdx)
		args = append(args, *enabled)
	}

	query += " ORDER BY created_at DESC"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list tunnels: %w", err)
	}
	defer rows.Close()

	var tunnels []*models.WireGuardTunnel
	for rows.Next() {
		t, err := r.scanTunnelRow(rows)
		if err != nil {
			return nil, err
		}
		tunnels = append(tunnels, t)
	}
	return tunnels, rows.Err()
}

func (r *Repository) ListAll(ctx context.Context, limit, offset int) ([]*models.WireGuardTunnel, error) {
	query := `
		SELECT id, user_id, host_id, interface_name, server_private_key, server_public_key,
		       client_public_key, preshared_key, server_ip, client_ip, listen_port,
		       allowed_ips, dns_servers, mtu, persistent_keepalive, status, is_enabled,
		       last_handshake_at, bytes_sent, bytes_received, last_rotated_at, created_at, updated_at
		FROM wireguard_tunnels
		ORDER BY created_at DESC LIMIT $1 OFFSET $2
	`
	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list all tunnels: %w", err)
	}
	defer rows.Close()

	var tunnels []*models.WireGuardTunnel
	for rows.Next() {
		t, err := r.scanTunnelRow(rows)
		if err != nil {
			return nil, err
		}
		tunnels = append(tunnels, t)
	}
	return tunnels, rows.Err()
}

func (r *Repository) Update(ctx context.Context, id uuid.UUID, updates map[string]interface{}) error {
	if len(updates) == 0 {
		return fmt.Errorf("no updates provided")
	}

	var setClauses []string
	var args []interface{}
	argCount := 0

	for field, value := range updates {
		argCount++
		setClauses = append(setClauses, fmt.Sprintf("%s = $%d", field, argCount))
		args = append(args, value)
	}

	argCount++
	query := fmt.Sprintf("UPDATE wireguard_tunnels SET %s, updated_at = NOW() WHERE id = $%d",
		joinStrings(setClauses, ", "), argCount)
	args = append(args, id)

	result, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to update tunnel: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrTunnelNotFound
	}
	return nil
}

func (r *Repository) Delete(ctx context.Context, id uuid.UUID) error {
	result, err := r.db.ExecContext(ctx, "DELETE FROM wireguard_tunnels WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("failed to delete tunnel: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrTunnelNotFound
	}
	return nil
}

func (r *Repository) CountByUser(ctx context.Context, userID uuid.UUID) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM wireguard_tunnels WHERE user_id = $1", userID).Scan(&count)
	return count, err
}

func (r *Repository) CountTotal(ctx context.Context) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM wireguard_tunnels").Scan(&count)
	return count, err
}

func (r *Repository) GetHostAddress(ctx context.Context, hostID uuid.UUID) (string, error) {
	var address string
	err := r.db.QueryRowContext(ctx, "SELECT address FROM hosts WHERE id = $1", hostID).Scan(&address)
	if err != nil {
		return "", fmt.Errorf("host not found: %w", err)
	}
	return address, nil
}

func (r *Repository) IncrementStats(ctx context.Context, id uuid.UUID, bytesSent, bytesRecv int64, lastHandshake *time.Time) error {
	query := `UPDATE wireguard_tunnels SET bytes_sent = bytes_sent + $2, bytes_received = bytes_received + $3`
	args := []interface{}{id, bytesSent, bytesRecv}
	argIdx := 3

	if lastHandshake != nil {
		argIdx++
		query += fmt.Sprintf(", last_handshake_at = $%d", argIdx)
		args = append(args, *lastHandshake)
	}

	query += fmt.Sprintf(", updated_at = NOW() WHERE id = $1")
	_, err := r.db.ExecContext(ctx, query, args...)
	return err
}

func (r *Repository) GetTunnelsDueForRotation(ctx context.Context, maxAgeDays int) ([]*models.WireGuardTunnel, error) {
	query := `
		SELECT id, user_id, host_id, interface_name, server_private_key, server_public_key,
		       client_public_key, preshared_key, server_ip, client_ip, listen_port,
		       allowed_ips, dns_servers, mtu, persistent_keepalive, status, is_enabled,
		       last_handshake_at, bytes_sent, bytes_received, last_rotated_at, created_at, updated_at
		FROM wireguard_tunnels
		WHERE is_enabled = true
		  AND (last_rotated_at IS NULL OR last_rotated_at < NOW() - ($1 || ' days')::INTERVAL)
	`
	rows, err := r.db.QueryContext(ctx, query, fmt.Sprintf("%d", maxAgeDays))
	if err != nil {
		return nil, fmt.Errorf("failed to query tunnels due for rotation: %w", err)
	}
	defer rows.Close()

	var tunnels []*models.WireGuardTunnel
	for rows.Next() {
		t, err := r.scanTunnelRow(rows)
		if err != nil {
			return nil, err
		}
		tunnels = append(tunnels, t)
	}
	return tunnels, rows.Err()
}

func (r *Repository) scanTunnel(row interface {
	Scan(dest ...interface{}) error
}) (*models.WireGuardTunnel, error) {
	return scanTunnelFromRow(row)
}

func (r *Repository) scanTunnelRow(row interface {
	Scan(dest ...interface{}) error
}) (*models.WireGuardTunnel, error) {
	return scanTunnelFromRow(row)
}

func scanTunnelFromRow(row interface {
	Scan(dest ...interface{}) error
}) (*models.WireGuardTunnel, error) {
	var t models.WireGuardTunnel
	var psk, lastRotatedAt sql.NullString
	var lastHandshakeAt sql.NullTime
	var allowedIPs, dnsSvrs pq.StringArray
	var status string

	err := row.Scan(
		&t.ID, &t.UserID, &t.HostID, &t.InterfaceName,
		&t.ServerPrivateKey, &t.ServerPublicKey,
		&t.ClientPublicKey, &psk,
		&t.ServerIP, &t.ClientIP, &t.ListenPort,
		&allowedIPs, &dnsSvrs,
		&t.MTU, &t.PersistentKeepalive,
		&status, &t.IsEnabled,
		&lastHandshakeAt, &t.BytesSent, &t.BytesReceived,
		&lastRotatedAt, &t.CreatedAt, &t.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrTunnelNotFound
		}
		return nil, fmt.Errorf("failed to scan tunnel: %w", err)
	}

	t.Status = models.TunnelStatus(status)
	t.AllowedIPs = []string(allowedIPs)
	t.DNSServers = []string(dnsSvrs)

	if psk.Valid {
		t.PresharedKey = &psk.String
	}
	if lastHandshakeAt.Valid {
		t.LastHandshakeAt = &lastHandshakeAt.Time
	}
	if lastRotatedAt.Valid {
		parsed, _ := time.Parse(time.RFC3339, lastRotatedAt.String)
		t.LastRotatedAt = &parsed
	}

	return &t, nil
}

func joinStrings(elems []string, sep string) string {
	if len(elems) == 0 {
		return ""
	}
	result := elems[0]
	for _, e := range elems[1:] {
		result += sep + e
	}
	return result
}

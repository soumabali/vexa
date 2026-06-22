package wireguard

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"golang.org/x/crypto/curve25519"

	"github.com/soumabali/vexa/internal/audit"
	"github.com/soumabali/vexa/internal/models"
)

type WireGuardService struct {
	repo         *Repository
	auditLogger  *audit.Logger
	ifaceCtrl    *InterfaceController
	statsCol     *StatsCollector
	rotSched     *RotationScheduler
	cfg          ServerConfig
	mu           sync.RWMutex
	nextIfaceIdx int
}

func NewWireGuardService(db *sql.DB, auditLogger *audit.Logger, cfg ServerConfig) *WireGuardService {
	repo := NewRepository(db)

	ifaceCtrl := NewInterfaceController(cfg.BinPath, cfg.Subnet)
	statsCol := NewStatsCollector(repo)
	rotSched := NewRotationScheduler(repo, ifaceCtrl, cfg.RotationDays)

	s := &WireGuardService{
		repo:        repo,
		auditLogger: auditLogger,
		ifaceCtrl:   ifaceCtrl,
		statsCol:    statsCol,
		rotSched:    rotSched,
		cfg:         cfg,
	}

	if cfg.Enabled {
		s.startBackground()
	}

	return s
}

func (s *WireGuardService) Repository() *Repository {
	return s.repo
}

func (s *WireGuardService) CreateTunnel(ctx context.Context, userID, hostID uuid.UUID, req *models.CreateTunnelRequest) (*models.WireGuardTunnel, error) {
	if !s.cfg.Enabled {
		return nil, fmt.Errorf("wireguard service is disabled")
	}

	count, err := s.repo.CountByUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to check tunnel count: %w", err)
	}
	if count >= s.cfg.MaxTunnelsPerUser {
		return nil, ErrMaxTunnelsReached
	}

	total, err := s.repo.CountTotal(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to check total tunnel count: %w", err)
	}
	if total >= s.cfg.MaxTotalTunnels {
		return nil, fmt.Errorf("maximum total tunnels reached")
	}

	existing, _ := s.repo.ListByUser(ctx, userID, hostID.String(), nil)
	for _, t := range existing {
		if t.HostID == hostID && t.Status != models.TunnelStatusDisabled {
			return nil, ErrTunnelExists
		}
	}

	port := req.Port
	if port == 0 {
		port = s.cfg.WGPortRange[0]
	}

	allowedIPs := req.AllowedIPs
	if len(allowedIPs) == 0 {
		allowedIPs = []string{"0.0.0.0/0", "::/0"}
	}

	mtu := req.MTU
	if mtu == 0 {
		mtu = 1420
	}

	serverPriv, serverPub, err := generateKeyPair()
	if err != nil {
		return nil, fmt.Errorf("failed to generate server keypair: %w", err)
	}

	clientPub, err := generateClientPublicKey()
	if err != nil {
		return nil, fmt.Errorf("failed to generate client key: %w", err)
	}

	var psk *string
	if req.UsePSK {
		p := generatePSK()
		psk = &p
	}

	serverIP, clientIP, err := s.allocateIPs()
	if err != nil {
		return nil, fmt.Errorf("failed to allocate IPs: %w", err)
	}

	ifaceName := s.nextInterfaceName()

	tunnel := &models.WireGuardTunnel{
		UserID:              userID,
		HostID:              hostID,
		InterfaceName:       ifaceName,
		ServerPrivateKey:    serverPriv,
		ServerPublicKey:     serverPub,
		ClientPublicKey:     clientPub,
		PresharedKey:        psk,
		ServerIP:            serverIP,
		ClientIP:            clientIP,
		ListenPort:          port,
		AllowedIPs:          allowedIPs,
		DNSServers:          req.DNSServers,
		MTU:                 mtu,
		PersistentKeepalive: 25,
		Status:              models.TunnelStatusCreating,
		IsEnabled:           false,
	}

	if err := s.repo.Create(ctx, tunnel); err != nil {
		return nil, fmt.Errorf("failed to save tunnel: %w", err)
	}

	if err := s.ifaceCtrl.CreateInterface(ifaceName, tunnel.ServerIP, port, tunnel.MTU); err != nil {
		log.Printf("warning: failed to create wg interface %s: %v", ifaceName, err)
		tunnel.Status = models.TunnelStatusError
		s.repo.Update(ctx, tunnel.ID, map[string]interface{}{"status": models.TunnelStatusError})
		return tunnel, nil
	}

	tunnel.Status = models.TunnelStatusConnected
	if err := s.repo.Update(ctx, tunnel.ID, map[string]interface{}{
		"status": models.TunnelStatusConnected,
	}); err != nil {
		log.Printf("warning: failed to update tunnel status: %v", err)
	}

	s.auditLog(ctx, audit.EventTunnelCreate, &userID, &tunnel.ID, map[string]interface{}{
		"host_id": hostID.String(),
		"iface":   ifaceName,
		"port":    port,
	})

	return tunnel, nil
}

func (s *WireGuardService) EnableTunnel(ctx context.Context, id, userID uuid.UUID) error {
	tunnel, err := s.repo.GetByIDAndUser(ctx, id, userID)
	if err != nil {
		return err
	}

	if tunnel.IsEnabled {
		return nil
	}

	if err := s.ifaceCtrl.SetInterfaceUp(tunnel.InterfaceName); err != nil {
		return fmt.Errorf("failed to enable interface: %w", err)
	}

	if err := s.configurePeer(tunnel); err != nil {
		return fmt.Errorf("failed to configure peer: %w", err)
	}

	if err := s.repo.Update(ctx, id, map[string]interface{}{
		"is_enabled": true,
		"status":     models.TunnelStatusConnected,
	}); err != nil {
		return err
	}

	s.auditLog(ctx, audit.EventTunnelEnable, &userID, &id, nil)
	return nil
}

func (s *WireGuardService) DisableTunnel(ctx context.Context, id, userID uuid.UUID) error {
	tunnel, err := s.repo.GetByIDAndUser(ctx, id, userID)
	if err != nil {
		return err
	}

	if !tunnel.IsEnabled {
		return nil
	}

	if err := s.ifaceCtrl.SetInterfaceDown(tunnel.InterfaceName); err != nil {
		return fmt.Errorf("failed to disable interface: %w", err)
	}

	if err := s.repo.Update(ctx, id, map[string]interface{}{
		"is_enabled": false,
		"status":     models.TunnelStatusDisabled,
	}); err != nil {
		return err
	}

	s.auditLog(ctx, audit.EventTunnelDisable, &userID, &id, nil)
	return nil
}

func (s *WireGuardService) DeleteTunnel(ctx context.Context, id, userID uuid.UUID) error {
	tunnel, err := s.repo.GetByIDAndUser(ctx, id, userID)
	if err != nil {
		return err
	}

	if tunnel.IsEnabled {
		if err := s.ifaceCtrl.SetInterfaceDown(tunnel.InterfaceName); err != nil {
			log.Printf("warning: failed to bring down interface %s: %v", tunnel.InterfaceName, err)
		}
	}

	if err := s.ifaceCtrl.DeleteInterface(tunnel.InterfaceName); err != nil {
		log.Printf("warning: failed to delete interface %s: %v", tunnel.InterfaceName, err)
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		return err
	}

	s.auditLog(ctx, audit.EventTunnelDelete, &userID, &id, map[string]interface{}{
		"host_id": tunnel.HostID.String(),
		"iface":   tunnel.InterfaceName,
	})
	return nil
}

func (s *WireGuardService) RotateKeys(ctx context.Context, id, userID uuid.UUID) (*models.WireGuardTunnel, error) {
	tunnel, err := s.repo.GetByIDAndUser(ctx, id, userID)
	if err != nil {
		return nil, err
	}

	newPriv, newPub, err := generateKeyPair()
	if err != nil {
		return nil, fmt.Errorf("failed to generate new keypair: %w", err)
	}

	oldPub := tunnel.ServerPublicKey
	now := time.Now().UTC()

	if err := s.repo.Update(ctx, id, map[string]interface{}{
		"server_private_key": newPriv,
		"server_public_key":  newPub,
		"last_rotated_at":    now,
	}); err != nil {
		return nil, fmt.Errorf("failed to save rotated keys: %w", err)
	}

	if tunnel.IsEnabled {
		if err := s.ifaceCtrl.UpdatePeerKey(tunnel.InterfaceName, tunnel.ClientPublicKey, newPriv); err != nil {
			log.Printf("warning: failed to update peer key on interface %s: %v", tunnel.InterfaceName, err)
		}
	}

	tunnel.ServerPrivateKey = newPriv
	tunnel.ServerPublicKey = newPub
	tunnel.LastRotatedAt = &now

	s.auditLog(ctx, audit.EventTunnelRotate, &userID, &id, map[string]interface{}{
		"old_public_key": oldPub,
	})
	return tunnel, nil
}

func (s *WireGuardService) GetConfig(ctx context.Context, id, userID uuid.UUID) (*models.TunnelConfigExport, error) {
	tunnel, err := s.repo.GetByIDAndUser(ctx, id, userID)
	if err != nil {
		return nil, err
	}

	host, err := s.getHostAddress(ctx, tunnel.HostID)
	if err != nil {
		return nil, fmt.Errorf("failed to get host address: %w", err)
	}

	export := &models.TunnelConfigExport{
		ClientPrivateKey:    "<client_private_key_generated_client_side>",
		Address:             tunnel.ClientIP,
		DNS:                 tunnel.DNSServers,
		ServerPublicKey:     tunnel.ServerPublicKey,
		AllowedIPs:          tunnel.AllowedIPs,
		Endpoint:            host,
		Port:                tunnel.ListenPort,
		PersistentKeepalive: tunnel.PersistentKeepalive,
		MTU:                 tunnel.MTU,
	}

	if tunnel.PresharedKey != nil {
		export.PresharedKey = *tunnel.PresharedKey
	}

	return export, nil
}

func (s *WireGuardService) GetStats(ctx context.Context, id, userID uuid.UUID) (*models.TunnelStats, error) {
	tunnel, err := s.repo.GetByIDAndUser(ctx, id, userID)
	if err != nil {
		return nil, err
	}

	wgStats := s.ifaceCtrl.GetTransferStats(tunnel.InterfaceName)

	stats := &models.TunnelStats{
		ID:              tunnel.ID,
		InterfaceName:   tunnel.InterfaceName,
		Status:          tunnel.Status,
		IsEnabled:       tunnel.IsEnabled,
		BytesSent:       wgStats.BytesSent,
		BytesReceived:   wgStats.BytesReceived,
		LastHandshakeAt: tunnel.LastHandshakeAt,
		ClientIP:        tunnel.ClientIP,
		ServerIP:        tunnel.ServerIP,
		ListenPort:      tunnel.ListenPort,
	}

	now := time.Now().UTC()
	if err := s.repo.IncrementStats(ctx, id, 0, 0, nil); err != nil {
		log.Printf("warning: failed to update stats timestamp: %v", err)
	}
	_ = now

	return stats, nil
}

func (s *WireGuardService) GetTunnel(ctx context.Context, id, userID uuid.UUID) (*models.WireGuardTunnel, error) {
	return s.repo.GetByIDAndUser(ctx, id, userID)
}

func (s *WireGuardService) ListTunnels(ctx context.Context, userID uuid.UUID, hostID string, enabled *bool) ([]*models.WireGuardTunnel, error) {
	return s.repo.ListByUser(ctx, userID, hostID, enabled)
}

func (s *WireGuardService) UpdateTunnel(ctx context.Context, id, userID uuid.UUID, req *models.UpdateTunnelRequest) (*models.WireGuardTunnel, error) {
	tunnel, err := s.repo.GetByIDAndUser(ctx, id, userID)
	if err != nil {
		return nil, err
	}

	updates := make(map[string]interface{})
	if req.AllowedIPs != nil {
		updates["allowed_ips"] = pq.Array(req.AllowedIPs)
		tunnel.AllowedIPs = req.AllowedIPs
	}
	if req.Port != nil {
		updates["listen_port"] = *req.Port
		tunnel.ListenPort = *req.Port
	}
	if req.DNSServers != nil {
		updates["dns_servers"] = pq.Array(req.DNSServers)
		tunnel.DNSServers = req.DNSServers
	}
	if req.MTU != nil {
		updates["mtu"] = *req.MTU
		tunnel.MTU = *req.MTU
	}

	if len(updates) > 0 {
		if err := s.repo.Update(ctx, id, updates); err != nil {
			return nil, err
		}
	}

	s.auditLog(ctx, audit.EventTunnelUpdate, &userID, &id, map[string]interface{}{
		"host_id": tunnel.HostID.String(),
	})
	return tunnel, nil
}

func (s *WireGuardService) StatsCollector() *StatsCollector {
	return s.statsCol
}

func (s *WireGuardService) RotationScheduler() *RotationScheduler {
	return s.rotSched
}

func (s *WireGuardService) startBackground() {
	s.rotSched.Start()
	s.statsCol.Start()
}

func (s *WireGuardService) Stop() {
	s.statsCol.Stop()
	s.rotSched.Stop()
}

func (s *WireGuardService) allocateIPs() (serverIP, clientIP string, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, subnet, err := net.ParseCIDR(s.cfg.Subnet)
	if err != nil {
		return "", "", ErrInvalidSubnet
	}

	ones, bits := subnet.Mask.Size()
	if bits == 0 {
		return "", "", ErrInvalidSubnet
	}

	networkIP := subnet.IP.To4()
	if networkIP == nil {
		networkIP = subnet.IP.To16()
	}

	serverIP = fmt.Sprintf("%s/%d", incrementIP(networkIP, 1).String(), ones)
	clientIP = fmt.Sprintf("%s/%d", incrementIP(networkIP, 2).String(), ones)

	return serverIP, clientIP, nil
}

func (s *WireGuardService) nextInterfaceName() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.nextIfaceIdx++
	return fmt.Sprintf("wg%d", s.nextIfaceIdx-1)
}

func (s *WireGuardService) configurePeer(tunnel *models.WireGuardTunnel) error {
	return s.ifaceCtrl.SetPeer(tunnel.InterfaceName, PeerConfig{
		PublicKey:           tunnel.ClientPublicKey,
		PresharedKey:        tunnel.PresharedKey,
		AllowedIPs:          tunnel.AllowedIPs,
		PersistentKeepalive: tunnel.PersistentKeepalive,
	})
}

func (s *WireGuardService) getHostAddress(ctx context.Context, hostID uuid.UUID) (string, error) {
	return s.repo.GetHostAddress(ctx, hostID)
}

func (s *WireGuardService) auditLog(ctx context.Context, eventType string, userID *uuid.UUID, tunnelID *uuid.UUID, details map[string]interface{}) {
	if s.auditLogger != nil {
		if details == nil {
			details = make(map[string]interface{})
		}
		if tunnelID != nil {
			details["tunnel_id"] = tunnelID.String()
		}
		s.auditLogger.Log(eventType, userID, nil, "", details)
	}
}

func generateKeyPair() (privateKey, publicKey string, err error) {
	var priv [32]byte
	if _, err := rand.Read(priv[:]); err != nil {
		return "", "", fmt.Errorf("failed to generate random bytes: %w", err)
	}

	priv[0] &= 248
	priv[31] &= 127
	priv[31] |= 64

	pub, err := curve25519.X25519(priv[:], curve25519.Basepoint)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate public key: %w", err)
	}

	return base64.StdEncoding.EncodeToString(priv[:]),
		base64.StdEncoding.EncodeToString(pub),
		nil
}

func generateClientPublicKey() (string, error) {
	_, pub, err := generateKeyPair()
	return pub, err
}

func generatePSK() string {
	key := make([]byte, 32)
	rand.Read(key)
	return base64.StdEncoding.EncodeToString(key)
}

func incrementIP(ip net.IP, increment int) net.IP {
	result := make(net.IP, len(ip))
	copy(result, ip)
	for i := len(result) - 1; i >= 0 && increment > 0; i-- {
		sum := int(result[i]) + increment
		result[i] = byte(sum & 0xFF)
		increment = sum >> 8
	}
	return result
}

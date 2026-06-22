package tunnel

import (
	"context"
	"errors"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/ssh"

	sshutil "github.com/soumabali/vexa/internal/ssh"
)

var (
	ErrTunnelNotFound   = errors.New("tunnel not found")
	ErrTunnelExists     = errors.New("tunnel already exists")
	ErrInvalidTunnelType = errors.New("invalid tunnel type")
	ErrTunnelFailed     = errors.New("tunnel creation failed")
	ErrPortInUse        = errors.New("local port already in use")
)

// TunnelType defines the type of port forwarding tunnel
type TunnelType string

const (
	TunnelTypeLocal   TunnelType = "local"
	TunnelTypeRemote  TunnelType = "remote"
	TunnelTypeDynamic TunnelType = "dynamic"
)

// TunnelStatus represents the current state of a tunnel
type TunnelStatus string

const (
	TunnelStatusPending   TunnelStatus = "pending"
	TunnelStatusActive    TunnelStatus = "active"
	TunnelStatusFailed    TunnelStatus = "failed"
	TunnelStatusClosed    TunnelStatus = "closed"
	TunnelStatusError     TunnelStatus = "error"
)

// Tunnel represents an SSH port forwarding tunnel
type Tunnel struct {
	ID          string
	Type        TunnelType
	LocalAddr   string
	RemoteAddr  string
	SSHConn     *ssh.Client
	Listener    net.Listener // For local/dynamic tunnels
	Status      TunnelStatus
	UserID      string
	HostID      string
	CreatedAt   time.Time
	LastUsed    time.Time
	BytesIn     int64
	BytesOut    int64
	ErrorMsg    string
	mu          sync.RWMutex
	active      bool
	ctx         context.Context
	cancel      context.CancelFunc
}

// TunnelConfig holds configuration for creating a tunnel
type TunnelConfig struct {
	Type       TunnelType
	LocalHost  string
	LocalPort  int
	RemoteHost string
	RemotePort int
	SSHHost    string
	SSHPort    int
	SSHUser    string
	SSHPassword string
	SSHKey     string
}

// TunnelManager manages active SSH tunnels
type TunnelManager struct {
	tunnels map[string]*Tunnel
	mu      sync.RWMutex

	// Port tracking
	localPorts map[int]bool
	portMu     sync.Mutex

	// Stats
	totalTunnels   int64
	activeTunnels  int64
	failedTunnels  int64
}

// TunnelStats provides statistics about tunnel usage
type TunnelStats struct {
	ID         string
	Type       TunnelType
	Status     TunnelStatus
	LocalAddr  string
	RemoteAddr string
	BytesIn    int64
	BytesOut   int64
	Uptime     time.Duration
	CreatedAt  time.Time
}

// NewTunnelManager creates a new tunnel manager
func NewTunnelManager() *TunnelManager {
	return &TunnelManager{
		tunnels:    make(map[string]*Tunnel),
		localPorts: make(map[int]bool),
	}
}

// CreateLocalTunnel creates a local port forwarding tunnel (SSH -L equivalent)
// LocalPort -> SSH -> RemoteHost:RemotePort
func (tm *TunnelManager) CreateLocalTunnel(config *TunnelConfig) (*Tunnel, error) {
	if config.Type != TunnelTypeLocal {
		return nil, ErrInvalidTunnelType
	}

	// Check if local port is available
	tm.portMu.Lock()
	if tm.localPorts[config.LocalPort] {
		tm.portMu.Unlock()
		return nil, ErrPortInUse
	}
	tm.localPorts[config.LocalPort] = true
	tm.portMu.Unlock()

	// Establish SSH connection
	sshClient, err := tm.dialSSH(config.SSHHost, config.SSHPort, config.SSHUser, config.SSHPassword)
	if err != nil {
		tm.releasePort(config.LocalPort)
		return nil, fmt.Errorf("%w: %v", ErrTunnelFailed, err)
	}

	localAddr := fmt.Sprintf("%s:%d", config.LocalHost, config.LocalPort)
	remoteAddr := fmt.Sprintf("%s:%d", config.RemoteHost, config.RemotePort)

	listener, err := net.Listen("tcp", localAddr)
	if err != nil {
		sshClient.Close()
		tm.releasePort(config.LocalPort)
		return nil, fmt.Errorf("%w: failed to bind to local port: %v", ErrTunnelFailed, err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	tunnel := &Tunnel{
		ID:         uuid.New().String(),
		Type:       TunnelTypeLocal,
		LocalAddr:  localAddr,
		RemoteAddr: remoteAddr,
		SSHConn:    sshClient,
		Listener:   listener,
		Status:     TunnelStatusActive,
		CreatedAt:  time.Now().UTC(),
		LastUsed:   time.Now().UTC(),
		active:     true,
		ctx:        ctx,
		cancel:     cancel,
	}

	tm.mu.Lock()
	tm.tunnels[tunnel.ID] = tunnel
	tm.mu.Unlock()

	// Start accepting connections
	go tm.handleLocalTunnel(tunnel, config.RemoteHost, config.RemotePort)

	return tunnel, nil
}

// CreateRemoteTunnel creates a remote port forwarding tunnel (SSH -R equivalent)
// RemotePort -> SSH -> LocalHost:LocalPort
func (tm *TunnelManager) CreateRemoteTunnel(config *TunnelConfig) (*Tunnel, error) {
	if config.Type != TunnelTypeRemote {
		return nil, ErrInvalidTunnelType
	}

	// Establish SSH connection
	sshClient, err := tm.dialSSH(config.SSHHost, config.SSHPort, config.SSHUser, config.SSHPassword)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrTunnelFailed, err)
	}

	localAddr := fmt.Sprintf("%s:%d", config.LocalHost, config.LocalPort)
	remoteAddr := fmt.Sprintf("%s:%d", config.SSHHost, config.RemotePort)

	ctx, cancel := context.WithCancel(context.Background())

	tunnel := &Tunnel{
		ID:         uuid.New().String(),
		Type:       TunnelTypeRemote,
		LocalAddr:  localAddr,
		RemoteAddr: remoteAddr,
		SSHConn:    sshClient,
		Status:     TunnelStatusActive,
		CreatedAt:  time.Now().UTC(),
		LastUsed:   time.Now().UTC(),
		active:     true,
		ctx:        ctx,
		cancel:     cancel,
	}

	tm.mu.Lock()
	tm.tunnels[tunnel.ID] = tunnel
	tm.mu.Unlock()

	// Start remote port forwarding
	go tm.handleRemoteTunnel(tunnel, config.LocalHost, config.LocalPort)

	return tunnel, nil
}

// CreateDynamicTunnel creates a dynamic SOCKS proxy tunnel (SSH -D equivalent)
func (tm *TunnelManager) CreateDynamicTunnel(config *TunnelConfig) (*Tunnel, error) {
	if config.Type != TunnelTypeDynamic {
		return nil, ErrInvalidTunnelType
	}

	// Check if local port is available
	tm.portMu.Lock()
	if tm.localPorts[config.LocalPort] {
		tm.portMu.Unlock()
		return nil, ErrPortInUse
	}
	tm.localPorts[config.LocalPort] = true
	tm.portMu.Unlock()

	// Establish SSH connection
	sshClient, err := tm.dialSSH(config.SSHHost, config.SSHPort, config.SSHUser, config.SSHPassword)
	if err != nil {
		tm.releasePort(config.LocalPort)
		return nil, fmt.Errorf("%w: %v", ErrTunnelFailed, err)
	}

	localAddr := fmt.Sprintf("%s:%d", config.LocalHost, config.LocalPort)

	listener, err := net.Listen("tcp", localAddr)
	if err != nil {
		sshClient.Close()
		tm.releasePort(config.LocalPort)
		return nil, fmt.Errorf("%w: failed to bind to local port: %v", ErrTunnelFailed, err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	tunnel := &Tunnel{
		ID:         uuid.New().String(),
		Type:       TunnelTypeDynamic,
		LocalAddr:  localAddr,
		RemoteAddr: "dynamic",
		SSHConn:    sshClient,
		Listener:   listener,
		Status:     TunnelStatusActive,
		CreatedAt:  time.Now().UTC(),
		LastUsed:   time.Now().UTC(),
		active:     true,
		ctx:        ctx,
		cancel:     cancel,
	}

	tm.mu.Lock()
	tm.tunnels[tunnel.ID] = tunnel
	tm.mu.Unlock()

	// Start SOCKS proxy handler
	go tm.handleDynamicTunnel(tunnel)

	return tunnel, nil
}

// GetTunnel returns a tunnel by ID
func (tm *TunnelManager) GetTunnel(id string) (*Tunnel, error) {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	tunnel, exists := tm.tunnels[id]
	if !exists {
		return nil, ErrTunnelNotFound
	}
	return tunnel, nil
}

// CloseTunnel closes a specific tunnel
func (tm *TunnelManager) CloseTunnel(id string) error {
	tm.mu.Lock()
	tunnel, exists := tm.tunnels[id]
	if !exists {
		tm.mu.Unlock()
		return ErrTunnelNotFound
	}
	delete(tm.tunnels, id)
	tm.mu.Unlock()

	tunnel.mu.Lock()
	tunnel.active = false
	tunnel.Status = TunnelStatusClosed
	tunnel.mu.Unlock()

	if tunnel.cancel != nil {
		tunnel.cancel()
	}

	if tunnel.Listener != nil {
		tunnel.Listener.Close()
	}

	if tunnel.SSHConn != nil {
		tunnel.SSHConn.Close()
	}

	// Release port if local/dynamic tunnel
	if tunnel.Type == TunnelTypeLocal || tunnel.Type == TunnelTypeDynamic {
		if port := tm.extractPort(tunnel.LocalAddr); port > 0 {
			tm.releasePort(port)
		}
	}

	return nil
}

// ListTunnels returns all active tunnels
func (tm *TunnelManager) ListTunnels() []*TunnelStats {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	stats := make([]*TunnelStats, 0, len(tm.tunnels))
	for _, t := range tm.tunnels {
		t.mu.RLock()
		stats = append(stats, &TunnelStats{
			ID:         t.ID,
			Type:       t.Type,
			Status:     t.Status,
			LocalAddr:  t.LocalAddr,
			RemoteAddr: t.RemoteAddr,
			BytesIn:    t.BytesIn,
			BytesOut:   t.BytesOut,
			Uptime:     time.Since(t.CreatedAt),
			CreatedAt:  t.CreatedAt,
		})
		t.mu.RUnlock()
	}
	return stats
}

// GetStats returns overall tunnel manager statistics
func (tm *TunnelManager) GetStats() map[string]interface{} {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	active := 0
	pending := 0
	failed := 0

	for _, t := range tm.tunnels {
		t.mu.RLock()
		switch t.Status {
		case TunnelStatusActive:
			active++
		case TunnelStatusPending:
			pending++
		case TunnelStatusFailed, TunnelStatusError:
			failed++
		}
		t.mu.RUnlock()
	}

	return map[string]interface{}{
		"total":   len(tm.tunnels),
		"active":  active,
		"pending": pending,
		"failed":  failed,
	}
}

// Cleanup closes all tunnels and cleans up resources
func (tm *TunnelManager) Cleanup() {
	tm.mu.Lock()
	tunnels := make([]*Tunnel, 0, len(tm.tunnels))
	for _, t := range tm.tunnels {
		tunnels = append(tunnels, t)
	}
	tm.tunnels = make(map[string]*Tunnel)
	tm.mu.Unlock()

	for _, tunnel := range tunnels {
		tunnel.mu.Lock()
		tunnel.active = false
		tunnel.Status = TunnelStatusClosed
		tunnel.mu.Unlock()

		if tunnel.cancel != nil {
			tunnel.cancel()
		}
		if tunnel.Listener != nil {
			tunnel.Listener.Close()
		}
		if tunnel.SSHConn != nil {
			tunnel.SSHConn.Close()
		}
	}

	tm.portMu.Lock()
	tm.localPorts = make(map[int]bool)
	tm.portMu.Unlock()
}

// --- Internal handlers (stubs for Sprint 3) ---

func (tm *TunnelManager) handleLocalTunnel(tunnel *Tunnel, remoteHost string, remotePort int) {
	defer func() {
		tunnel.mu.Lock()
		tunnel.active = false
		tunnel.Status = TunnelStatusClosed
		tunnel.mu.Unlock()
	}()

	for {
		select {
		case <-tunnel.ctx.Done():
			return
		default:
		}

		conn, err := tunnel.Listener.Accept()
		if err != nil {
			select {
			case <-tunnel.ctx.Done():
				return
			default:
				tunnel.mu.Lock()
				tunnel.Status = TunnelStatusError
				tunnel.ErrorMsg = err.Error()
				tunnel.mu.Unlock()
				continue
			}
		}

		tunnel.mu.Lock()
		tunnel.LastUsed = time.Now().UTC()
		tunnel.mu.Unlock()

		go tm.proxyLocalConnection(tunnel, conn, remoteHost, remotePort)
	}
}

func (tm *TunnelManager) handleRemoteTunnel(tunnel *Tunnel, localHost string, localPort int) {
	// Sprint 3: Implement remote port forwarding
	// This requires using ssh.Request "tcpip-forward"
	defer func() {
		tunnel.mu.Lock()
		tunnel.active = false
		tunnel.Status = TunnelStatusClosed
		tunnel.mu.Unlock()
	}()

	// Placeholder: mark as pending until implemented
	tunnel.mu.Lock()
	tunnel.Status = TunnelStatusPending
	tunnel.mu.Unlock()

	<-tunnel.ctx.Done()
}

func (tm *TunnelManager) handleDynamicTunnel(tunnel *Tunnel) {
	// Sprint 3: Implement SOCKS proxy
	defer func() {
		tunnel.mu.Lock()
		tunnel.active = false
		tunnel.Status = TunnelStatusClosed
		tunnel.mu.Unlock()
	}()

	// Placeholder: mark as pending until implemented
	tunnel.mu.Lock()
	tunnel.Status = TunnelStatusPending
	tunnel.mu.Unlock()

	<-tunnel.ctx.Done()
}

func (tm *TunnelManager) proxyLocalConnection(tunnel *Tunnel, localConn net.Conn, remoteHost string, remotePort int) {
	defer localConn.Close()

	// Sprint 3: Implement actual SSH channel forwarding
	// For now, just update stats
	defer func() {
		tunnel.mu.Lock()
		tunnel.LastUsed = time.Now().UTC()
		tunnel.mu.Unlock()
	}()

	// Placeholder connection to remote via SSH
	remoteAddr := fmt.Sprintf("%s:%d", remoteHost, remotePort)
	_ = remoteAddr

	// This would dial through SSH: tunnel.SSHConn.Dial("tcp", remoteAddr)
}

// --- Helpers ---

func (tm *TunnelManager) dialSSH(host string, port int, user, password string) (*ssh.Client, error) {
	config := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.Password(password),
		},
		HostKeyCallback: sshutil.HostKeyCallback(),
		Timeout:         10 * time.Second,
	}

	client, err := ssh.Dial("tcp", fmt.Sprintf("%s:%d", host, port), config)
	if err != nil {
		return nil, fmt.Errorf("SSH dial failed: %w", err)
	}

	return client, nil
}

func (tm *TunnelManager) releasePort(port int) {
	tm.portMu.Lock()
	delete(tm.localPorts, port)
	tm.portMu.Unlock()
}

func (tm *TunnelManager) extractPort(addr string) int {
	_, portStr, err := net.SplitHostPort(addr)
	if err != nil {
		return 0
	}
	var port int
	fmt.Sscanf(portStr, "%d", &port)
	return port
}
